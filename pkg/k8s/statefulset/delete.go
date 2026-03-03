package statefulset

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

// DeleteStatefulSet 删除 StatefulSet（前台级联删除，会删除关联的 Pod）
// 注意：PVC 默认不会自动删除，需要调用 DeleteStatefulSetWithPVC 来手动清理
// 删除 PVC 不会删除 PV 和底层存储数据（前提是 PV 的 ReclaimPolicy 为 Retain）
func DeleteStatefulSet(ctx context.Context, Kube kubernetes.Interface, name, ns string, timeout time.Duration) error {
	fg := metav1.DeletePropagationForeground
	if err := Kube.AppsV1().StatefulSets(ns).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &fg}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		if apierrors.IsForbidden(err) {
			return fmt.Errorf("没有权限删除 StatefulSet %s/%s", ns, name)
		}
		return err
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		_, err := Kube.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil && (apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)) {
			return false, nil
		}
		return false, err
	})
}

// DeleteStatefulSetWithPVC 删除 StatefulSet、Pod 和 PVC（保留 PV 和底层存储数据）
// 流程：
//   1) 获取 StatefulSet 信息（用于查找关联的 PVC）
//   2) 删除 StatefulSet（会级联删除 Pod）
//   3) 删除关联的 PVC（不会删除 PV，前提是 PV 的 ReclaimPolicy 设置为 Retain）
// 
// 重要提示：
//   - PVC 删除后，对应的 PV 会变为 Released 状态（如果 ReclaimPolicy=Retain）
//   - PV 和底层存储数据会被保留，可以手动回收或重新绑定
//   - 如果 PV 的 ReclaimPolicy 是 Delete，则会删除底层存储，请谨慎！
func DeleteStatefulSetWithPVC(ctx context.Context, Kube kubernetes.Interface, name, ns string, timeout time.Duration) error {
	// 1. 获取 StatefulSet 信息（用于后续查找 PVC）
	sts, err := Kube.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// StatefulSet 不存在，但可能有遗留的 PVC，尝试清理
			return cleanupOrphanedPVCs(ctx, Kube, name, ns)
		}
		return fmt.Errorf("获取 StatefulSet 失败: %w", err)
	}

	// 保存 selector 和 volumeClaimTemplates 信息
	selector := sts.Spec.Selector
	vctNames := make([]string, 0, len(sts.Spec.VolumeClaimTemplates))
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		vctNames = append(vctNames, vct.Name)
	}

	// 2. 删除 StatefulSet
	if err := DeleteStatefulSet(ctx, Kube, name, ns, timeout); err != nil {
		return err
	}

	// 3. 删除关联的 PVC（保留 PV）
	// 删除 PVC 只会解除 PVC 与 PV 的绑定关系，不会删除 PV 本身
	// PV 的删除取决于其 ReclaimPolicy：
	//   - Retain: PV 保留，状态变为 Released，数据完整保留
	//   - Delete: PV 和底层存储会被删除（需谨慎）
	return deletePVCsForStatefulSet(ctx, Kube, name, ns, selector, vctNames)
}

// deletePVCsForStatefulSet 删除 StatefulSet 创建的 PVC（但保留 PV 和数据）
// PVC 命名规则: <volumeClaimTemplate.name>-<statefulset.name>-<ordinal>
// 
// 删除逻辑说明：
//   1. 删除 PVC 对象会解除 PVC 与 PV 的绑定
//   2. PV 不会被删除，会根据 ReclaimPolicy 进入不同状态：
//      - Retain: PV 保留并变为 Released 状态，数据完整
//      - Delete: PV 和底层存储会被自动删除
//   3. 底层存储数据是否保留取决于 StorageClass 或 PV 的 persistentVolumeReclaimPolicy
func deletePVCsForStatefulSet(ctx context.Context, Kube kubernetes.Interface, stsName, ns string, selector *metav1.LabelSelector, vctNames []string) error {
	if selector == nil || len(vctNames) == 0 {
		return nil
	}

	// 构建 label selector
	labelSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return fmt.Errorf("解析 label selector 失败: %w", err)
	}

	// 列出所有匹配的 PVC
	pvcList, err := Kube.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("列出 PVC 失败: %w", err)
	}

	// 删除匹配 StatefulSet 命名规则的 PVC
	var deleteErr error
	for _, pvc := range pvcList.Items {
		// PVC 名称格式: <vct-name>-<sts-name>-<ordinal>
		for _, vctName := range vctNames {
			prefix := fmt.Sprintf("%s-%s-", vctName, stsName)
			if len(pvc.Name) > len(prefix) && pvc.Name[:len(prefix)] == prefix {
				if err := Kube.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, pvc.Name, metav1.DeleteOptions{}); err != nil {
					if !apierrors.IsNotFound(err) {
						deleteErr = fmt.Errorf("删除 PVC %s 失败: %w", pvc.Name, err)
					}
				}
				break
			}
		}
	}

	return deleteErr
}

// cleanupOrphanedPVCs 清理孤儿 PVC（StatefulSet 已删除但 PVC 残留）
func cleanupOrphanedPVCs(ctx context.Context, Kube kubernetes.Interface, stsName, ns string) error {
	// 列出命名空间下所有 PVC
	pvcList, err := Kube.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil // 忽略错误
	}

	// 查找包含 StatefulSet 名称的 PVC
	for _, pvc := range pvcList.Items {
		// 检查 PVC 名称是否包含 -<stsName>- 模式
		if containsStatefulSetName(pvc.Name, stsName) {
			_ = Kube.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
		}
	}

	return nil
}

// containsStatefulSetName 检查 PVC 名称是否属于指定的 StatefulSet
func containsStatefulSetName(pvcName, stsName string) bool {
	// PVC 命名规则: <vct-name>-<sts-name>-<ordinal>
	pattern := "-" + stsName + "-"
	return len(pvcName) > len(pattern) && contains(pvcName, pattern)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetStatefulSetPVCs 获取 StatefulSet 关联的所有 PVC
func GetStatefulSetPVCs(ctx context.Context, Kube kubernetes.Interface, sts *appv1.StatefulSet) ([]string, error) {
	if sts.Spec.Selector == nil {
		return nil, nil
	}

	labelSelector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, err
	}

	pvcList, err := Kube.CoreV1().PersistentVolumeClaims(sts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}

	vctNames := make(map[string]bool)
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		vctNames[vct.Name] = true
	}

	var result []string
	for _, pvc := range pvcList.Items {
		for vctName := range vctNames {
			prefix := fmt.Sprintf("%s-%s-", vctName, sts.Name)
			if len(pvc.Name) > len(prefix) && pvc.Name[:len(prefix)] == prefix {
				result = append(result, pvc.Name)
				break
			}
		}
	}

	return result, nil
}

// 确保 labels 包已使用
var _ = labels.Nothing

package node

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// PatchTaints 修改节点污点（添加/删除）
func PatchTaints(ctx context.Context, kube kubernetes.Interface, nodeName string, addTaints []corev1.Taint, removeKeys []string) error {
	// 获取当前节点
	nodeObj, err := kube.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get node failed: %w", err)
	}

	// 构建新的污点列表
	newTaints := make([]corev1.Taint, 0)

	// 保留不在删除列表中的现有污点
	removeSet := make(map[string]bool)
	for _, key := range removeKeys {
		removeSet[key] = true
	}

	for _, t := range nodeObj.Spec.Taints {
		if !removeSet[t.Key] {
			newTaints = append(newTaints, t)
		}
	}

	// 添加/更新污点
	for _, t := range addTaints {
		// 检查是否已存在相同 key+effect 的污点
		found := false
		for i, existing := range newTaints {
			if existing.Key == t.Key && existing.Effect == t.Effect {
				newTaints[i] = t // 更新
				found = true
				break
			}
		}
		if !found {
			newTaints = append(newTaints, t)
		}
	}

	// 构建 patch
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"taints": newTaints,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch failed: %w", err)
	}

	_, err = kube.CoreV1().Nodes().Patch(ctx, nodeName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("patch node taints failed: %w", err)
	}

	return nil
}

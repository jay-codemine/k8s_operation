package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/pvc"
	"sigs.k8s.io/yaml"
)

// KubeCreatePVC 创建 PersistentVolumeClaim
func (s *Services) KubeCreatePVC(ctx context.Context, cli *K8sClients, req *requests.KubePVCCreateRequest) (*corev1.PersistentVolumeClaim, error) {
	// 2) 调用资源层进行构建 + 创建
	created, err := pvc.CreatePersistentVolumeClaim(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s already exists in namespace %s", req.Name, req.Namespace)
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Errorf("create PersistentVolumeClaim failed: %v", err)
		return nil, err
	}

	// 3) 成功日志
	global.Logger.Infof("PersistentVolumeClaim %q created successfully in namespace %q", created.Name, req.Namespace)
	return created, nil
}

// KubePVCList 获取 PVC 列表（支持分页与名称模糊）
func (s *Services) KubePVCList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubePVCListRequest,
) ([]corev1.PersistentVolumeClaim, int64, error) {

	items, total, err := pvc.GetPVCList(
		ctx,
		cli.Kube,
		param.Namespace,
		param.Name,
		param.Page,
		param.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// KubePVCDetail 获取 PVC 详情
func (s *Services) KubePVCDetail(ctx context.Context, cli *K8sClients, param *requests.KubePVCDetailRequest) (*corev1.PersistentVolumeClaim, error) {
	pvcDetail, err := pvc.GetPVCDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil, fmt.Errorf("PersistentVolumeClaim %q not found in namespace %q", param.Name, param.Namespace)
		}
		global.Logger.Error("get PersistentVolumeClaim detail failed", zap.Error(err))
		return nil, err
	}

	return pvcDetail, nil
}

func (s *Services) KubePVCDelete(ctx context.Context, cli *K8sClients, param *requests.KubePVCDeleteRequest) error {
	if err := pvc.DeletePersistentVolumeClaim(ctx, cli.Kube, param.Namespace, param.Name); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("PersistentVolumeClaim %s/%s not found", param.Namespace, param.Name)
			return nil // 幂等
		}
		global.Logger.Errorf("delete PersistentVolumeClaim %s/%s failed: %v", param.Namespace, param.Name, err)
		return err
	}

	global.Logger.Infof("PersistentVolumeClaim %s/%s deleted successfully", param.Namespace, param.Name)
	return nil
}

// 扩容 PVC：仅允许修改 spec.resources.requests.storage
func (s *Services) KubePVCResize(ctx context.Context, cli *K8sClients, req *requests.KubePVCResizeRequest,
) (*corev1.PersistentVolumeClaim, error) {
	return pvc.ResizePVC(ctx, cli.Kube, req)
}

// KubePVCCreateFromYaml 从 YAML 创建 PVC
func (s *Services) KubePVCCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	// 1. 解析 YAML 到 Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlContent), &unstructuredObj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证是否为 PVC
	if unstructuredObj.GetKind() != "PersistentVolumeClaim" {
		return nil, fmt.Errorf("YAML kind must be PersistentVolumeClaim, got %q", unstructuredObj.GetKind())
	}

	// 3. 转换为 PVC 对象
	pvcObj := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, pvcObj); err != nil {
		return nil, fmt.Errorf("failed to convert to PersistentVolumeClaim: %w", err)
	}

	// 4. 确保 namespace 不为空
	if pvcObj.Namespace == "" {
		pvcObj.Namespace = "default"
	}

	// 5. 调用 K8s API 创建
	created, err := cli.Kube.CoreV1().PersistentVolumeClaims(pvcObj.Namespace).Create(ctx, pvcObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", pvcObj.Name, pvcObj.Namespace)
		}
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	global.Logger.Infof("PVC created from YAML: %s/%s", created.Namespace, created.Name)
	return created, nil
}

// KubePVCApplyYaml 应用 PVC YAML 更改（更新已存在的 PVC）
func (s *Services) KubePVCApplyYaml(ctx context.Context, cli *K8sClients, namespace, name, yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	// 1. 解析 YAML
	unstructuredObj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(yamlContent), &unstructuredObj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 转换为 PVC 对象
	pvcObj := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, pvcObj); err != nil {
		return nil, fmt.Errorf("failed to convert to PersistentVolumeClaim: %w", err)
	}

	// 3. 确保 namespace/name 匹配
	if pvcObj.Namespace == "" {
		pvcObj.Namespace = namespace
	}
	if pvcObj.Name != name || pvcObj.Namespace != namespace {
		return nil, fmt.Errorf("YAML name/namespace mismatch: expected %s/%s, got %s/%s", namespace, name, pvcObj.Namespace, pvcObj.Name)
	}

	// 4. 获取现有 PVC（保留 ResourceVersion）
	existing, err := cli.Kube.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing PVC: %w", err)
	}

	// 5. 保留必要的元数据
	pvcObj.ResourceVersion = existing.ResourceVersion
	pvcObj.UID = existing.UID

	// 6. 更新 PVC
	updated, err := cli.Kube.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvcObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update PVC: %w", err)
	}

	global.Logger.Infof("PVC updated from YAML: %s/%s", updated.Namespace, updated.Name)
	return updated, nil
}

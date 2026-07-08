package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/namespace"
)

// KubeCreateNamespace 封装 Namespace 创建逻辑（调用资源层）
func (s *Services) KubeCreateNamespace(ctx context.Context, cli *K8sClients, req *requests.KubeNamespaceCreateRequest) (*corev1.Namespace, error) {
	return namespace.CreateNamespace(ctx, cli.Kube, req)
}

func (s *Services) KubeNamespaceList(ctx context.Context, cli *K8sClients, param *requests.KubeNamespaceListRequest) ([]corev1.Namespace, int, error) {
	items, total, err := namespace.GetNamespaceList(ctx, cli.Kube, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list Namespace failed: %v", err)
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Services) KubeNamespaceDetail(ctx context.Context, cli *K8sClients, param *requests.KubeNamespaceDetailRequest) (*corev1.Namespace, error) {
	nsDetail, err := namespace.GetNamespaceDetail(ctx, cli.Kube, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Namespace %s not found", param.Name)
			return nil, fmt.Errorf("Namespace %q not found", param.Name)
		}

		global.Logger.Error("get Namespace detail failed", zap.Error(err))
		return nil, err
	}

	return nsDetail, nil
}

func (s *Services) KubeNamespaceDelete(ctx context.Context, cli *K8sClients, param *requests.KubeNamespaceDeleteRequest) error {
	// 调用内部逻辑删除 Namespace
	if err := namespace.DeleteNamespace(ctx, cli.Kube, param.Name); err != nil {
		// 不存在视为删除成功（幂等）
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Namespace %s not found", param.Name)
			return nil
		}

		global.Logger.Errorf("delete Namespace %s failed: %v", param.Name, err)
		return err
	}

	global.Logger.Infof("Namespace %s deleted successfully", param.Name)
	return nil
}

func (s *Services) KubeNamespaceUpdate(ctx context.Context, cli *K8sClients, param *requests.KubeNamespaceUpdateRequest) (*corev1.Namespace, error) {
	updated, err := namespace.PatchNamespace(ctx, cli.Kube, param.Name, param.Content)
	if err != nil {
		global.Logger.Errorf("update Namespace %s failed: %v", param.Name, err)
		return nil, err
	}

	global.Logger.Infof("Namespace %s updated successfully", param.Name)
	return updated, nil
}

// KubeNamespacePatchLabels 修改 Namespace 标签
func (s *Services) KubeNamespacePatchLabels(ctx context.Context, cli *K8sClients, param *requests.KubeNamespaceLabelPatchRequest) error {
	err := namespace.PatchLabels(ctx, cli.Kube, param.Name, param.Add, param.Remove)
	if err != nil {
		global.Logger.Errorf("patch Namespace labels failed: name=%s err=%v", param.Name, err)
		return err
	}
	global.Logger.Infof("patch Namespace labels success: name=%s", param.Name)
	return nil
}

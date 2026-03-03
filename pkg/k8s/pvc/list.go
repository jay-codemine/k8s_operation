package pvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetPVCList 列出指定命名空间（或全局）的 PVC，支持按名称模糊、分页
func GetPVCList(
	ctx context.Context,
	kube kubernetes.Interface,
	namespace, name string,
	page, limit int,
) ([]corev1.PersistentVolumeClaim, int64, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 支持全局：namespace 为空时列出所有命名空间
	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll // 等价于 ""
	}

	list, err := kube.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		// 只有“显式传入 namespace 且不存在”才给友好错误
		if apierrors.IsNotFound(err) && namespace != "" {
			return nil, 0, fmt.Errorf("namespace %q not found", namespace)
		}
		return nil, 0, err
	}

	selector := dataselect.NewPersistentVolumeClaimSelector(list.Items, name, page, limit)
	selector.Filter().Sort()

	total := int64(selector.TotalCount())

	data := selector.Paginate()

	return dataselect.FromPersistentVolumeClaimCells(data.GenericDataList), total, nil
}

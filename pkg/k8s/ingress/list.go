package ingress

import (
	"context"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetIngressList 拉取 Ingress 列表，支持分页、名称过滤
func GetIngressList(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace string,
	page, limit int,
) ([]networkingv1.Ingress, int64, error) {

	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := kube.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// 保留你原来的“namespace 不存在”的友好错误
		if apierrors.IsNotFound(err) {
			return nil, 0, fmt.Errorf("namespace %s not found", namespace)
		}
		return nil, 0, err
	}

	selector := dataselect.NewIngressSelector(list.Items, name, page, limit)

	selector.Filter().Sort()

	total := int64(selector.TotalCount())

	data := selector.Paginate()

	return dataselect.FromIngressCells(data.GenericDataList), total, nil
}

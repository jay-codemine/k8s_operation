package svc

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetServiceList 与 GetDeploymentList 逻辑一致，返回 Service 列表 + 总数（分页前）
func GetServiceList(
	ctx context.Context,
	kube kubernetes.Interface,
	name, namespace string,
	page, limit int,
) ([]corev1.Service, int64, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := kube.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	selector := dataselect.NewServiceSelector(list.Items, name, page, limit)

	selector.Filter().Sort()

	total := int64(selector.TotalCount())

	data := selector.Paginate()

	return dataselect.FromServiceCells(data.GenericDataList), total, nil
}

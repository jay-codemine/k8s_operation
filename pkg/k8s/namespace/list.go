package namespace

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetNamespaceList 列出所有 Namespace，支持名称模糊过滤 + 分页
func GetNamespaceList(ctx context.Context, Kube kubernetes.Interface, name string, page, limit int) ([]corev1.Namespace, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := Kube.CoreV1().
		Namespaces().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, 0, err
	}

	// 基于你项目里的 DataSelect 体系
	selector := dataselect.NewNamespaceSelector(list.Items, name, page, limit)
	selector.Filter().Sort()

	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromNamespaceCells(data.GenericDataList), total, nil
}

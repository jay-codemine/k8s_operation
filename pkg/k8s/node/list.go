package node

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/pkg/k8s/dataselect"
)

// GetNodeList 列出集群所有 Node，支持名称模糊 + 分页
func GetNodeList(ctx context.Context, kube kubernetes.Interface, name string, page, limit int) ([]corev1.Node, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 集群级资源：不需要 namespace
	list, err := kube.CoreV1().
		Nodes().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		// Node 为集群级，一般不会是 namespace NotFound，这里直接返回底层错误
		return nil, 0, err
	}

	selector := dataselect.NewNodeSelector(list.Items, name, page, limit)
	selector.Filter().Sort()
	total := selector.TotalCount()
	data := selector.Paginate()

	return dataselect.FromNodeCells(data.GenericDataList), total, nil
}

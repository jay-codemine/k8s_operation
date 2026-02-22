package event

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/utils"
)

func listEventsCoreV1(ctx context.Context, Kube kubernetes.Interface, namespace string, q *requests.KubeEventListRequest) ([]models.EventItem, string, error) {
	// 使用BuildFieldSelectorCoreV1函数构建字段选择器，传入查询参数的Kind、Name、Type和Reason
	fs := BuildFieldSelectorCoreV1(q.Kind, q.Name, q.Type, q.Reason)
	// 创建ListOptions结构体，设置字段选择器、限制数量和继续令牌
	opts := metav1.ListOptions{
		FieldSelector: fs,                        // 设置字段选择器
		Limit:         utils.ClampLimit(q.Limit), // 使用ClampLimit函数限制查询数量
		Continue:      q.ContinueToken,           // 设置继续令牌用于分页
	}
	// 使用Kubernetes客户端获取指定namespace中的事件列表
	lst, err := Kube.CoreV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, "", err // 如果发生错误，返回空值和错误信息
	}

	// 创建EventItem切片，容量为lst.Items的长度
	items := make([]models.EventItem, 0, len(lst.Items))
	// 遍历lst.Items，将每个事件转换为EventItem并添加到items切片中
	for _, ev := range lst.Items {
		items = append(items, BuildEventItemFromCoreV1(&ev))
	}
	// 应用时间过滤和排序，并返回处理后的items、继续令牌和nil错误
	items = utils.ApplySinceAndSort(items, q.SinceSeconds)

	// 返回查询结果，包含三个元素：
	// 1. items: 查询到的数据项列表
	// 2. lst.Continue: 用于分页的继续令牌，表示是否有更多数据可以获取
	// 3. nil: 表示操作成功，没有错误发生
	return items, lst.Continue, nil
}

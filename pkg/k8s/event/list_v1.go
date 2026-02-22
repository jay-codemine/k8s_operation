package event

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/utils"
)

func ListEventsV1(ctx context.Context, kube kubernetes.Interface, namespace string, q *requests.KubeEventListRequest) ([]models.EventItem, string, error) {
	// 构建FieldSelector用于EventsV1的事件选择
	fs := BuildFieldSelectorEventsV1(q.Kind, q.Name, q.Type, q.Reason)
	// 设置ListOptions，包含字段选择器、限制数量和继续令牌
	opts := metav1.ListOptions{
		FieldSelector: fs,                        // 字段选择器，用于过滤事件
		Limit:         utils.ClampLimit(q.Limit), // 限制返回的事件数量，使用工具函数进行限制
		Continue:      q.ContinueToken,           // 分页继续令牌，用于分页查询
	}
	// 使用Kubernetes客户端获取事件列表
	lst, err := kube.EventsV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, "", err // 如果发生错误，返回空结果和错误信息
	}

	// 创建EventItem切片，容量为lst.Items的长度
	items := make([]models.EventItem, 0, len(lst.Items))

	// 遍历事件列表，将每个事件转换为EventItem
	for _, ev := range lst.Items {
		items = append(items, BuildEventItemFromEventV1(ev))
	}

	// 应用时间过滤和排序
	items = utils.ApplySinceAndSort(items, q.SinceSeconds)

	// 返回三个值：
	// 1. items - 查询到的项目列表
	// 2. lst.Continue - 用于分页的继续令牌，如果还有更多数据则不为空
	// 3. nil - 表示操作成功，没有错误发生
	return items, lst.Continue, nil
}

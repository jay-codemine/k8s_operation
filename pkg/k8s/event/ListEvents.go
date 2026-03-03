package event

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/utils"
)

func ListEvents(ctx context.Context, Kube kubernetes.Interface, q *requests.KubeEventListRequest) (items []models.EventItem, next string, err error) {
	// 使用 utils.NormalizeNamespace 对命名空间进行规范化处理
	ns := utils.NormalizeNamespace(q.Namespace)

	// 优先新版，失败回退旧版
	if utils.TryEventsV1First() {
		if items, next, err = ListEventsV1(ctx, Kube, ns, q); err == nil {
			return items, next, nil
		}
	}
	return listEventsCoreV1(ctx, Kube, ns, q)
}

package daemonset

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RevisionItem 版本历史项
type RevisionItem struct {
	Name         string    `json:"name"`          // ControllerRevision 名称
	Revision     int64     `json:"revision"`      // 版本号
	Image        string    `json:"image"`         // 首个容器镜像
	Images       []string  `json:"images"`        // 全部容器镜像
	CreationTime time.Time `json:"creation_time"` // 创建时间
}

// GetDaemonSetHistory 获取 DaemonSet 的历史版本（ControllerRevision 列表）
func GetDaemonSetHistory(ctx context.Context, kube kubernetes.Interface, namespace, name string) ([]RevisionItem, error) {
	// 1. 获取 DaemonSet
	ds, err := kube.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet 失败: %w", err)
	}

	// 2. 获取 selector
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析 selector 失败: %w", err)
	}

	// 3. 列出 ControllerRevision
	revList, err := kube.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取 ControllerRevision 列表失败: %w", err)
	}

	// 4. 过滤：只保留属于该 DaemonSet 的 ControllerRevision
	var result []RevisionItem
	for _, rev := range revList.Items {
		for _, ref := range rev.OwnerReferences {
			if ref.Kind == "DaemonSet" && ref.Name == name {
				item := RevisionItem{
					Name:         rev.Name,
					Revision:     rev.Revision,
					CreationTime: rev.CreationTimestamp.Time,
				}
				// 从 Data.Raw({"spec":{"template":{...}}}) 解析历史模板镜像，供前端回滚弹窗展示
				var dsTemplate appv1.DaemonSet
				if err := json.Unmarshal(rev.Data.Raw, &dsTemplate); err == nil {
					for _, c := range dsTemplate.Spec.Template.Spec.Containers {
						item.Images = append(item.Images, c.Image)
					}
					if len(item.Images) > 0 {
						item.Image = item.Images[0]
					}
				}
				result = append(result, item)
				break
			}
		}
	}

	// 5. 按版本号降序排序（最新在前）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Revision > result[j].Revision
	})

	return result, nil
}

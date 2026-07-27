package statefulset

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

// ControllerRevisionItem 版本历史项（供前端回滚弹窗展示）
type ControllerRevisionItem struct {
	Name         string    `json:"name"`          // ControllerRevision 名称
	Revision     int64     `json:"revision"`      // 版本号
	Image        string    `json:"image"`         // 首个容器镜像
	Images       []string  `json:"images"`        // 全部容器镜像
	CreationTime time.Time `json:"creation_time"` // 创建时间
}

// GetStatefulSetHistory 获取 StatefulSet 的历史版本（ControllerRevision 列表）
// 从 ControllerRevision.Data 中解析出历史 Pod 模板的镜像，便于前端展示回滚目标
func GetStatefulSetHistory(ctx context.Context, kube kubernetes.Interface, namespace, name string) ([]ControllerRevisionItem, error) {
	revs, err := GetStatefulSetControllerRevisions(ctx, kube, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("获取 ControllerRevision 列表失败: %w", err)
	}

	result := make([]ControllerRevisionItem, 0, len(revs))
	for i := range revs {
		item := ControllerRevisionItem{
			Name:         revs[i].Name,
			Revision:     revs[i].Revision,
			CreationTime: revs[i].CreationTimestamp.Time,
		}
		// Data.Raw 结构为 {"spec":{"template":{...}}}，与回滚时的解析方式一致
		var stsTemplate appv1.StatefulSet
		if err := json.Unmarshal(revs[i].Data.Raw, &stsTemplate); err == nil {
			for _, c := range stsTemplate.Spec.Template.Spec.Containers {
				item.Images = append(item.Images, c.Image)
			}
			if len(item.Images) > 0 {
				item.Image = item.Images[0]
			}
		}
		result = append(result, item)
	}

	// 按版本号降序排序（最新在前）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Revision > result[j].Revision
	})

	return result, nil
}

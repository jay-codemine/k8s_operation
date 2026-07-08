package hpa

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// HPAItem 列表/响应数据结构（前端友好）
type HPAItem struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	TargetKind        string            `json:"target_kind"`
	TargetName        string            `json:"target_name"`
	MinReplicas       int32             `json:"min_replicas"`
	MaxReplicas       int32             `json:"max_replicas"`
	CurrentReplicas   int32             `json:"current_replicas"`
	DesiredReplicas   int32             `json:"desired_replicas"`
	CPUTargetUtil     *int32            `json:"cpu_target_util,omitempty"`     // 目标 CPU 利用率
	MemTargetUtil     *int32            `json:"mem_target_util,omitempty"`     // 目标内存利用率
	CurrentCPU        *int32            `json:"current_cpu_util,omitempty"`    // 当前 CPU 利用率
	CurrentMemory     *int32            `json:"current_memory_util,omitempty"` // 当前内存利用率
	Status            string            `json:"status"`                         // Active/Pending/Failed
	Conditions        []HPACondition    `json:"conditions,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTimestamp string            `json:"creation_timestamp"`
}

// HPACondition 简化的 HPA 条件
type HPACondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// BuildHPAItem 把 K8s HPA 对象转换成前端友好结构
func BuildHPAItem(h *autoscalingv2.HorizontalPodAutoscaler) HPAItem {
	item := HPAItem{
		Name:              h.Name,
		Namespace:         h.Namespace,
		TargetKind:        h.Spec.ScaleTargetRef.Kind,
		TargetName:        h.Spec.ScaleTargetRef.Name,
		MaxReplicas:       h.Spec.MaxReplicas,
		CurrentReplicas:   h.Status.CurrentReplicas,
		DesiredReplicas:   h.Status.DesiredReplicas,
		Labels:            h.Labels,
		CreationTimestamp: h.CreationTimestamp.Time.Local().Format("2006-01-02 15:04:05"),
	}
	if h.Spec.MinReplicas != nil {
		item.MinReplicas = *h.Spec.MinReplicas
	}

	// 提取目标利用率
	for _, m := range h.Spec.Metrics {
		if m.Type == autoscalingv2.ResourceMetricSourceType && m.Resource != nil {
			if m.Resource.Target.AverageUtilization != nil {
				v := *m.Resource.Target.AverageUtilization
				switch m.Resource.Name {
				case "cpu":
					item.CPUTargetUtil = &v
				case "memory":
					item.MemTargetUtil = &v
				}
			}
		}
	}

	// 提取当前利用率
	for _, cm := range h.Status.CurrentMetrics {
		if cm.Type == autoscalingv2.ResourceMetricSourceType && cm.Resource != nil {
			if cm.Resource.Current.AverageUtilization != nil {
				v := *cm.Resource.Current.AverageUtilization
				switch cm.Resource.Name {
				case "cpu":
					item.CurrentCPU = &v
				case "memory":
					item.CurrentMemory = &v
				}
			}
		}
	}

	// 状态判定
	item.Status = "Pending"
	for _, c := range h.Status.Conditions {
		item.Conditions = append(item.Conditions, HPACondition{
			Type:    string(c.Type),
			Status:  string(c.Status),
			Reason:  c.Reason,
			Message: c.Message,
		})
		if c.Type == autoscalingv2.ScalingActive && c.Status == "True" {
			item.Status = "Active"
		}
		if c.Type == autoscalingv2.AbleToScale && c.Status == "False" {
			item.Status = "Failed"
		}
	}
	return item
}

// BuildHPAItemList 批量转换
func BuildHPAItemList(list []autoscalingv2.HorizontalPodAutoscaler) []HPAItem {
	items := make([]HPAItem, 0, len(list))
	for i := range list {
		items = append(items, BuildHPAItem(&list[i]))
	}
	return items
}

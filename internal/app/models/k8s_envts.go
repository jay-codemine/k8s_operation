package models

import "time"

// 统一事件输出结构（对外返回）
type EventItem struct {
	// 定义事件相关字段的结构体
	// 这些字段通常用于表示Kubernetes或其他系统中的事件信息
	Namespace       string    `json:"namespace"`                  // 命名空间，资源所在的命名空间
	Kind            string    `json:"kind"`                       // 资源类型，如Pod、Service等
	Name            string    `json:"name"`                       // 资源名称
	Type            string    `json:"type"`                       // 事件类型，如Normal、Warning等
	Reason          string    `json:"reason"`                     // 事件发生的原因
	Message         string    `json:"message"`                    // 事件的详细描述信息
	Count           int32     `json:"count"`                      // 事件发生的次数
	EventTime       time.Time `json:"event_time"`                 // 事件发生的时间
	SourceComponent string    `json:"source_component,omitempty"` // 事件来源组件，omitempty表示字段为空时不在JSON中显示
	SourceInstance  string    `json:"source_instance,omitempty"`  // 事件来源实例，omitempty表示字段为空时不在JSON中显示
}

func NewEventItem() *EventItem {
	return &EventItem{
		Type: "Normal", // 默认事件类型
	}
}

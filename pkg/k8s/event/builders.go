package event

import (
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8soperation/internal/app/models"
)

// 组装 fieldSelector：新版/旧版分别构造
func BuildFieldSelectorEventsV1(kind, name, typ, reason string) string {
	sels := []fields.Selector{}
	if kind != "" {
		sels = append(sels, fields.OneTermEqualSelector("regarding.kind", kind))
	}
	if name != "" {
		sels = append(sels, fields.OneTermEqualSelector("regarding.name", name))
	}
	if typ != "" {
		sels = append(sels, fields.OneTermEqualSelector("type", typ))
	}
	if reason != "" {
		sels = append(sels, fields.OneTermEqualSelector("reason", reason))
	}
	if len(sels) == 0 {
		return ""
	}
	return fields.AndSelectors(sels...).String()
}

func BuildFieldSelectorCoreV1(kind, name, typ, reason string) string {
	sels := []fields.Selector{}
	if kind != "" {
		sels = append(sels, fields.OneTermEqualSelector("involvedObject.kind", kind))
	}
	if name != "" {
		sels = append(sels, fields.OneTermEqualSelector("involvedObject.name", name))
	}
	if typ != "" {
		sels = append(sels, fields.OneTermEqualSelector("type", typ))
	}
	if reason != "" {
		sels = append(sels, fields.OneTermEqualSelector("reason", reason))
	}
	if len(sels) == 0 {
		return ""
	}
	return fields.AndSelectors(sels...).String()
}

func BuildEventItemFromEventV1(ev eventsv1.Event) models.EventItem {
	// 获取事件时间
	t := ev.EventTime.Time
	// 如果事件时间为零且事件系列不为空，则使用事件系列的最后观察时间
	if t.IsZero() && ev.Series != nil {
		t = ev.Series.LastObservedTime.Time
	}
	// 设置事件计数为1
	cnt := int32(1)
	// 返回一个格式化后的事件项结构体
	return models.EventItem{
		Namespace:       ev.Namespace,           // 命名空间
		Kind:            ev.Regarding.Kind,      // 资源类型
		Name:            ev.Regarding.Name,      // 资源名称
		Type:            ev.Type,                // 事件类型
		Reason:          ev.Reason,              // 事件原因
		Message:         ev.Note,                // 事件消息
		Count:           cnt,                    // 事件计数
		EventTime:       t,                      // 事件时间
		SourceComponent: ev.ReportingController, // 报告组件
		SourceInstance:  ev.ReportingInstance,   // 报告实例
	}
}

func BuildEventItemFromCoreV1(ev *corev1.Event) models.EventItem {
	// 从事件中获取时间t，优先使用LastTimestamp，其次使用FirstTimestamp，默认使用EventTime
	t := ev.EventTime.Time
	if !ev.LastTimestamp.IsZero() {
		t = ev.LastTimestamp.Time // 如果LastTimestamp不为空，则使用LastTimestamp的时间
	} else if !ev.FirstTimestamp.IsZero() {
		t = ev.FirstTimestamp.Time // 如果FirstTimestamp不为空，则使用FirstTimestamp的时间
	}

	// 返回一个包含事件各项信息的EventItem结构体
	return models.EventItem{
		Namespace:       ev.Namespace,           // 事件所属的命名空间
		Kind:            ev.InvolvedObject.Kind, // 被涉及对象的类型
		Name:            ev.InvolvedObject.Name, // 被涉及对象的名称
		Type:            ev.Type,                // 事件的类型
		Reason:          ev.Reason,              // 事件的原因
		Message:         ev.Message,             // 事件的消息内容
		Count:           ev.Count,               // 事件发生的次数
		EventTime:       t,                      // 事件发生的时间
		SourceComponent: ev.Source.Component,    // 事件的来源组件
		SourceInstance:  ev.Source.Host,         // 事件的来源实例
	}
}

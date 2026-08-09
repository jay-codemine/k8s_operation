package models

import dm "k8soperation/internal/domain/k8s"

// EventItem K8s 事件（领域定义在 domain/k8s）
type EventItem = dm.EventItem

// NewEventItem 创建默认事件
var NewEventItem = dm.NewEventItem

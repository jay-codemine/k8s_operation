package k8s

import "k8soperation/internal/domain/events"

// ——— 集群领域事件 ———

// ClusterCreated 集群创建事件
type ClusterCreated struct {
	events.BaseEvent
	ClusterID   uint32
	ClusterName string
}

func NewClusterCreated(id uint32, name string) ClusterCreated {
	return ClusterCreated{BaseEvent: events.NewBaseEvent("k8s.cluster.created"), ClusterID: id, ClusterName: name}
}

// ClusterDeleted 集群删除事件
type ClusterDeleted struct {
	events.BaseEvent
	ClusterID   uint32
	ClusterName string
}

func NewClusterDeleted(id uint32, name string) ClusterDeleted {
	return ClusterDeleted{BaseEvent: events.NewBaseEvent("k8s.cluster.deleted"), ClusterID: id, ClusterName: name}
}

// ClusterHealthChanged 集群健康状态变化事件
type ClusterHealthChanged struct {
	events.BaseEvent
	ClusterID uint32
	OldStatus uint8
	NewStatus uint8
}

func NewClusterHealthChanged(id uint32, old, new uint8) ClusterHealthChanged {
	return ClusterHealthChanged{
		BaseEvent: events.NewBaseEvent("k8s.cluster.health_changed"),
		ClusterID: id, OldStatus: old, NewStatus: new,
	}
}

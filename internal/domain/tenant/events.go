package tenant

import "k8soperation/internal/domain/events"

// TenantCreated 租户创建事件
type TenantCreated struct {
	events.BaseEvent
	TenantID uint32
	Name     string
}

func NewTenantCreated(id uint32, name string) TenantCreated {
	return TenantCreated{BaseEvent: events.NewBaseEvent("tenant.created"), TenantID: id, Name: name}
}

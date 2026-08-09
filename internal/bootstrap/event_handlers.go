package bootstrap

import (
	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/domain/events"
)

// registerEventHandlers 注册全局领域事件处理器
func registerEventHandlers() {
	bus := services.EventBus()

	bus.Subscribe("k8s.cluster.created", func(e events.DomainEvent) {
		global.Logger.Infof("[DomainEvent] 集群创建事件已触发: %s", e.EventName())
	})
	bus.Subscribe("k8s.cluster.deleted", func(e events.DomainEvent) {
		global.Logger.Infof("[DomainEvent] 集群删除事件已触发: %s", e.EventName())
	})
	bus.Subscribe("cicd.pipeline.triggered", func(e events.DomainEvent) {
		global.Logger.Infof("[DomainEvent] 流水线触发事件: %s", e.EventName())
	})
	bus.Subscribe("user.registered", func(e events.DomainEvent) {
		global.Logger.Infof("[DomainEvent] 用户注册事件: %s", e.EventName())
	})
	bus.Subscribe("tenant.created", func(e events.DomainEvent) {
		global.Logger.Infof("[DomainEvent] 租户创建事件: %s", e.EventName())
	})
}

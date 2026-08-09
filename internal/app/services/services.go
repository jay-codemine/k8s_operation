package services

import (
	"gorm.io/gorm"

	"k8soperation/global"
	"k8soperation/internal/app/infra"
	"k8soperation/internal/domain/events"
	"k8soperation/pkg/logger"
	"k8soperation/pkg/tenant"
)

var sharedEventBus = events.NewEventBus(global.Logger)

type Services struct {
	db        *gorm.DB
	stream    *infra.RedisStream
	logger    *logger.Logger
	eventBus  *events.EventBus
	tenantID  uint32
}

// NewServices 创建全局 Services 实例（启动期/后台 Worker 使用 global.DB）
func NewServices() *Services {
	return &Services{
		db:       global.DB,
		stream:   infra.NewRedisStream(global.RedisCli),
		logger:   global.Logger,
		eventBus: sharedEventBus,
	}
}

// NewServicesWithDB 创建租户隔离的 Services 实例（HTTP 请求使用）
func NewServicesWithDB(db *gorm.DB) *Services {
	tid, _ := tenant.GetTenantID(db)
	return &Services{
		db:       db,
		stream:   infra.NewRedisStream(global.RedisCli),
		logger:   global.Logger,
		eventBus: sharedEventBus,
		tenantID: tid,
	}
}

// NewBackgroundServices 启动期/后台任务使用 global.DB（跨租户）
func NewBackgroundServices() *Services {
	return &Services{
		db:       global.DB,
		stream:   infra.NewRedisStream(global.RedisCli),
		logger:   global.Logger,
		eventBus: sharedEventBus,
	}
}

// EventBus 返回共享事件总线（用于注册全局事件处理器）
func EventBus() *events.EventBus { return sharedEventBus }


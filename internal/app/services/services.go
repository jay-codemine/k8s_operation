package services

import (
	"gorm.io/gorm"

	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/infra"
)

type Services struct {
	dao    *dao.Dao
	stream *infra.RedisStream
}

// NewServices 创建全局 Services 实例（启动期/后台 Worker 使用 global.DB）
func NewServices() *Services {
	return &Services{
		dao:    dao.NewDao(global.DB),
		stream: infra.NewRedisStream(global.RedisCli),
	}
}

// NewServicesWithDB 创建租户隔离的 Services 实例（HTTP 请求使用）
// db 应为 middlewares.GetTenantDB(c) 返回的租户隔离 DB
func NewServicesWithDB(db *gorm.DB) *Services {
	return &Services{
		dao:    dao.NewDao(db),
		stream: infra.NewRedisStream(global.RedisCli),
	}
}

// NewBackgroundServices 启动期/后台任务使用 global.DB（跨租户）
func NewBackgroundServices() *Services {
	return &Services{
		dao: dao.NewDao(global.DB),
	}
}

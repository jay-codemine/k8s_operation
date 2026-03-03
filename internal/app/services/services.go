package services

import (
	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/infra"
)

type Services struct {
	dao    *dao.Dao
	stream *infra.RedisStream
}

func NewServices() *Services {
	return &Services{
		dao:    dao.NewDao(global.DB),
		stream: infra.NewRedisStream(global.RedisCli),
	}
}

// 启动期/后台任务使用（新增）
func NewBackgroundServices() *Services {
	return &Services{
		dao: dao.NewDao(global.DB),
	}
}

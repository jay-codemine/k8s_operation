package bootstrap

import (
	"context"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/initialize"
	"k8soperation/internal/app/services"
	"k8soperation/internal/app/worker"
	"k8soperation/pkg/k8s/crd"
)

var (
	// cicdWorker CICD 部署任务消费者
	cicdWorker *worker.CicdWorker
)

func InitAll() error {
	// 初始化配置
	if err := initialize.SetupSetting(); err != nil {
		return err
	}
	// 初始校验规则
	if err := initialize.SetupValidator(); err != nil {
		return err
	}

	// 初始化日志
	if err := initialize.SetupLogger(); err != nil {
		return err
	}

	// 初始化数据库
	if err := initialize.SetupDB(); err != nil {
		global.Logger.Error("init db failed", zap.Error(err))
	}

	// 初始化Redis-session
	if err := initialize.SetupSession(); err != nil {
		return err
	}

	// 初始化Redis客户端
	if err := initialize.SetupRedis(); err != nil {
		panic(err)
	}

	// 初始化K8s
	if err := initialize.SetupK8sBootstrap(); err != nil {
		return err
	}

	// 初始化 AppConfig CRD 客户端
	if err := crd.SetupAppConfigClient(); err != nil {
		return err
	}

	// 加载 swagger 接口文档
	initialize.LogDocsReady()

	// 初始化并启动 CICD Worker
	if err := StartCicdWorker(); err != nil {
		global.Logger.Warn("start cicd worker failed", zap.Error(err))
		// 不返回错误，Worker 启动失败不影响主服务
	}

	return nil
}

// Sync() 会做两件事：
// 调用底层 WriteSyncer 的 Sync()（例如 os.File.Sync()）；
// 把缓冲日志强制写到文件。
func FlushLoggers() {
	// 系统日志落盘
	_ = global.Logger.Sync()
	if global.BizLogger != nil {
		// 业务日志落盘
		_ = global.BizLogger.Sync()
	}
}

// StartCicdWorker 启动 CICD Worker
func StartCicdWorker() error {
	if global.RedisCli == nil {
		global.Logger.Warn("redis client not initialized, cicd worker will not start")
		return nil
	}

	// 创建集群客户端工厂
	svc := services.NewBackgroundServices()
	factory := services.NewClusterClientFactory(svc)

	// 创建并启动 Worker
	cicdWorker = worker.NewCicdWorker(global.RedisCli, factory)
	if err := cicdWorker.Start(context.Background()); err != nil {
		return err
	}

	global.Logger.Info("cicd worker started successfully")
	return nil
}

// StopCicdWorker 停止 CICD Worker
func StopCicdWorker() {
	if cicdWorker != nil {
		cicdWorker.Stop()
	}
}

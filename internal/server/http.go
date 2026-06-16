package server

import (
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/initialize"
	"k8soperation/internal/app/services"
	"net/http"
	"time"
)

func NewHTTPServer() *http.Server {
	// 1) 创建 Services（包含 DB/DAO/配置等）
	svc := services.NewServices()

	// 2) 创建多集群 Client 工厂（TTL/TTLJitter 走配置）
	factory := services.NewClusterClientFactory(svc)

	// 3) 初始化引擎（注入 factory，路由里才能 Use ClusterMiddleware(factory)）
	engine := initialize.NewEngine(factory)

	// 兜底超时
	shutdownTimeout := time.Duration(global.ServerSetting.ShutdownTimeout) * time.Second
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	srv := &http.Server{
		Addr:              ":" + global.ServerSetting.Port,
		Handler:           engine.Engine, //注意：用 *gin.Engine
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       global.ServerSetting.ReadTimeout * time.Second,
		WriteTimeout:      global.ServerSetting.WriteTimeout * time.Second,
		IdleTimeout:       global.ServerSetting.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20,
		ErrorLog:          global.Logger.StdLogger(),
	}

	srv.RegisterOnShutdown(func() {
		global.Logger.Info("http k8soperation shutdown")
		if global.SQLDB != nil {
			_ = global.SQLDB.Close()
		}
	})

	_ = shutdownTimeout // 如果你后面还有优雅关闭逻辑用到它，就保留；否则可以删

	return srv
}

// 记录服务器启动日志
func logServerStart(srv *http.Server) {
	global.Logger.Info("http k8soperation starting",
		zap.String("addr", srv.Addr),
		zap.String("mode", global.ServerSetting.RunMode))
}

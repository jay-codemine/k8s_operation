// test/main.go

package main

import (
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/initialize"
	"k8soperation/internal/server"
	"log"
	"time"
)

func main() {
	if err := initialize.SetupSetting(); err != nil {
		log.Fatalf("SetupSetting err: %v", err)
	}
	if err := initialize.SetupLogger(); err != nil {
		log.Fatalf("SetupLogger err: %v", err)
	}

	defer func() { _ = global.Logger.Sync() }()

	if err := initialize.SetupDB(); err != nil {
		global.Logger.Error("SetupDB err", zap.Error(err))
	}

	// 测试日志轮转等逻辑
	// logtest.FloodLogs()

	// 启动 HTTP 服务并支持优雅退出
	srv := server.NewHTTPServer()
	server.ListenAndServeAsync(srv)
	server.GracefulShutdown(srv, 5*time.Second)

	// 关闭数据库连接
	if global.SQLDB != nil {
		_ = global.SQLDB.Close()
	}
}

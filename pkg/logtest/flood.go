package logtest

import (
	"go.uber.org/zap"
	"k8soperation/global"
	"strings"
	"time"
)

// FloodLogs 会快速写满日志文件，触发 lumberjack 轮转
func FloodLogs() {
	big := strings.Repeat("X", 300*1024) // 300KB
	for i := 0; i < 10; i++ {
		global.Logger.Info("rotate-check", zap.Int("i", i), zap.String("payload", big))
		time.Sleep(20 * time.Millisecond)
	}
}

// package logtest
func FloodBizLogs() {
	big := strings.Repeat("X", 300*1024) // 每条 ~300KB
	for i := 0; i < 10; i++ {            // 共 ~3MB
		global.BizLogger.Info("biz-rotate-check",
			zap.Int("i", i),
			zap.String("payload", big),
		)
		time.Sleep(20 * time.Millisecond)
	}
	_ = global.BizLogger.Sync() // 刷盘，确保切割及时落盘
}

// FloodAllLogs 会同时写系统日志和业务日志，触发 lumberjack 轮转
func FloodAllLogs() {
	// 系统日志压测
	FloodLogs()

	// 业务日志压测
	FloodBizLogs()
}

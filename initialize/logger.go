// initialize/logger.go
package initialize

import (
	"fmt"
	"k8soperation/global"
	logger2 "k8soperation/pkg/logger"
	"os"
	"path/filepath"
)

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create log dir %q: %w", dir, err)
			}
			return nil
		}
		return fmt.Errorf("stat log dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("log path %q exists but is not a directory", dir)
	}
	return nil
}

// SetupLogger 初始化系统日志与业务日志
func SetupLogger() error {
	if global.AppSetting == nil {
		return fmt.Errorf("AppSetting is nil")
	}

	// —— 系统日志目录
	if err := ensureDir(global.AppSetting.LogFileName); err != nil {
		return err
	}
	// —— 业务日志目录（给默认值）
	if global.AppSetting.BusinessLogFileName == "" {
		global.AppSetting.BusinessLogFileName = "storage/logs/biz.log"
	}
	if err := ensureDir(global.AppSetting.BusinessLogFileName); err != nil {
		return err
	}

	// —— 日志级别与选项（统一只算/声明一次）
	lvl := logger2.WithLevel(global.AppSetting.LogLevel)
	sysOpts := []logger2.Option{
		logger2.AddCaller(),
		logger2.AddCallerSkip(1),
		logger2.AddStacktrace(logger2.ErrorLevel),
	}
	bizOpts := []logger2.Option{
		logger2.AddCaller(),
		logger2.AddCallerSkip(1),
	}

	// —— 初始化系统日志（stdout + 文件轮转）
	global.Logger = logger2.NewLogger(
		lvl,
		logger2.RotateOptions{
			FileName:   global.AppSetting.LogFileName,
			MaxSize:    global.AppSetting.LogMaxSize,
			MaxBackups: global.AppSetting.LogMaxBackup,
			MaxAge:     global.AppSetting.LogMaxAge,
			Compress:   global.AppSetting.LogCompress,
		},
		sysOpts...,
	)

	// —— 初始化业务日志（仅文件，Info 级别）
	global.BizLogger = logger2.NewBusinessLogger(
		logger2.RotateOptions{
			FileName:   global.AppSetting.BusinessLogFileName,
			MaxSize:    global.AppSetting.LogMaxSize,
			MaxBackups: global.AppSetting.LogMaxBackup,
			MaxAge:     global.AppSetting.LogMaxAge,
			Compress:   global.AppSetting.LogCompress,
		},
		bizOpts...,
	)

	// —— 镜像开关
	global.MirrorBizToSys = global.AppSetting.MirrorBusinessToSystem

	return nil
}

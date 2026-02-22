package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

// 只收 Info 级别的业务日志 Logger（zap.Logger）
// NewBusinessLogger 创建一个只记录 Info 级别的业务日志 Logger
// 参数：
//   - ropt    : 日志轮转配置（文件路径、大小、保留数量、压缩等）
//   - options : zap 提供的可选配置（如 AddCaller、AddStacktrace）
//
// 功能：
//  1. 配置日志编码器（JSON 格式，时间字段为 "time"，格式化为 YYYY-MM-DD HH:mm:ss）。
//  2. 输出目标为日志文件（通过 lumberjack 自动切割/归档），不输出到 stdout。
//  3. 日志级别过滤：只接收 Info 级别的日志（忽略 Debug/Warn/Error）。
//  4. 返回一个封装后的 *Logger（内部含 zap.Logger）。
//
// 使用场景：
//   - 用于记录“业务操作日志 / 审计日志”，如用户注册、登录、修改配置等。
//   - 业务日志通常只需要 Info 级别，因为它主要是行为追踪，而不是错误调试。
func NewBusinessLogger(ropt RotateOptions, options ...Option) *Logger {
	// 1. 配置日志编码器（时间格式改为人类可读）
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time" // 字段名由 "ts" 改为 "time"
	encCfg.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.Format("2006-01-02 15:04:05"))
	}

	// 2. 配置日志文件输出（带轮转）
	//    - 超过 MaxSize MB 自动切割
	//    - 最多保留 MaxBackups 个历史文件
	//    - 超过 MaxAge 天的日志会被删除
	//    - Compress = true 时会压缩归档旧日志
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   ropt.FileName,
		MaxSize:    ropt.MaxSize,
		MaxBackups: ropt.MaxBackups,
		MaxAge:     ropt.MaxAge,
		Compress:   ropt.Compress,
	})

	// 3. 输出目标：这里只写文件，不写控制台
	ws := zapcore.NewMultiWriteSyncer(fileWriter)

	// 4. 创建 zapcore.Core：只接收 Info 级别的日志
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg), // JSON 编码
		ws,                             // 输出目标
		zap.LevelEnablerFunc(func(l zapcore.Level) bool { // 过滤器
			return l == zapcore.InfoLevel // 只允许 Info
		}),
	)

	// 5. 返回自定义 Logger 封装（内部持有 zap.Logger）
	return &Logger{
		logger: zap.New(core, options...), // 真正的 zap.Logger
		level:  InfoLevel,                 // 保存当前日志等级（仅 Info）
	}
}

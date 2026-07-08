package utils

import (
	"k8soperation/global"
	"sync"

	"go.uber.org/zap"
)

// SafeGo 安全启动 goroutine，自动捕获 panic 防止服务崩溃
// 用法: utils.SafeGo(func() { ... })
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("goroutine panic recovered",
					zap.Any("panic", r),
					zap.Stack("stacktrace"),
				)
			}
		}()
		fn()
	}()
}

// SafeGoWithCallback 带回调的安全 goroutine
// 注意：panic 时 fn 内部的 defer 会先执行，然后才调用 onPanic
// 因此不要在 fn 和 onPanic 中重复调用相同操作（如 wg.Done）
func SafeGoWithCallback(fn func(), onPanic func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("goroutine panic recovered",
					zap.Any("panic", r),
					zap.Stack("stacktrace"),
				)
				if onPanic != nil {
					onPanic(r)
				}
			}
		}()
		fn()
	}()
}

// SafeGoWithWaitGroup 专门用于 WaitGroup 场景的安全 goroutine
// 自动在完成或 panic 时调用 wg.Done()，避免重复调用
func SafeGoWithWaitGroup(wg *sync.WaitGroup, fn func(), onPanic func(r interface{})) {
	go func() {
		defer func() {
			wg.Done() // 统一在这里调用，无论成功还是 panic
			if r := recover(); r != nil {
				global.Logger.Error("goroutine panic recovered",
					zap.Any("panic", r),
					zap.Stack("stacktrace"),
				)
				if onPanic != nil {
					onPanic(r)
				}
			}
		}()
		fn()
	}()
}

// SafeGoWithName 带名称标识的安全 goroutine，便于日志追踪
func SafeGoWithName(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("goroutine panic recovered",
					zap.String("goroutine", name),
					zap.Any("panic", r),
					zap.Stack("stacktrace"),
				)
			}
		}()
		fn()
	}()
}

// MustRecover 用于在已存在的 goroutine 中添加 panic 保护
// 用法: defer utils.MustRecover("task-name")
func MustRecover(name string) {
	if r := recover(); r != nil {
		global.Logger.Error("panic recovered",
			zap.String("location", name),
			zap.Any("panic", r),
			zap.Stack("stacktrace"),
		)
	}
}

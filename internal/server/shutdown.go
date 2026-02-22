package server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"k8soperation/global"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ListenAndServeAsync(srv *http.Server) {
	go func() {
		logServerStart(srv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.Logger.Error("http k8soperation err", zap.Error(err))
		}
	}()
}

func GracefulShutdown(srv *http.Server, timeout time.Duration) {
	// 使用 signal.NotifyContext 创建一个上下文，用于捕获系统中断信号
	// 包括中断信号(SIGINT)、终止信号(SIGTERM)和退出信号(SIGQUIT)
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 确保在函数退出时停止信号监听
	defer stop()
	// 阻塞直到接收到中断信号
	<-stopCtx.Done()

	// 记录服务器关闭开始日志
	global.Logger.Info("shutting down k8soperation...")
	// 创建一个带超时的上下文，用于控制服务器关闭的等待时间
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// 确保在函数退出时取消上下文
	defer cancel()

	// 尝试正常关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		// 如果关闭过程中出现错误，记录错误日志
		global.Logger.Error("k8soperation shutdown err", zap.Error(err))
	} else {
		// 如果服务器成功关闭，记录成功日志和超时时间
		global.Logger.Info("k8soperation exiting", zap.Duration("timeout", timeout))
	}

	// 如果业务日志器存在，记录服务停止日志
	if global.BizLogger != nil {
		global.BizLogger.Info("service.stop")
	}
}

package main

import (
	"k8soperation/global"
	"k8soperation/internal/bootstrap"
	"k8soperation/internal/server"
	"log"
	"time"
)

// 构建信息，由 docker/backend/Dockerfile 的 -ldflags -X 注入。
//
// 必须在 main 包里真实声明：-X 作用于不存在的符号时链接器会静默忽略，
// 于是镜像里查不出自己是哪个 commit 编出来的。v1.1.2 就是这么变成一笔
// 糊涂账的——二进制经 -trimpath 剥离，又因 Docker 内 git 的 dubious
// ownership 没能留下 vcs.revision，只能靠构建时间去反推基线。
var (
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// @title K8s管理平台
// @version 1.0
// @description 基于Gin+Vue开发的K8s管理平台
// @termsOfService https://gitee.com/jay-kim/k8s_operation
// 省略 import…

func main() {
	log.Printf("[Build] commit=%s build_time=%s", GitCommit, BuildTime)

	// 初始化所有组件，如果初始化失败则panic
	if err := bootstrap.InitAll(); err != nil {
		panic(err)
	}

	// 确保程序退出时刷新所有日志记录器和停止Worker
	defer func() {
		bootstrap.StopCicdWorker()
		bootstrap.FlushLoggers()
	}()

	// 创建新的HTTP服务器实例
	srv := server.NewHTTPServer()

	// 异步启动HTTP服务器
	server.ListenAndServeAsync(srv)

	// 从全局配置中获取服务器关闭超时时间，并转换为time.Duration类型
	timeout := global.ServerSetting.ShutdownTimeout * time.Second

	// 如果配置的超时时间小于等于0，则默认设置为5秒
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// 优雅地关闭服务器，等待指定超时时间
	server.GracefulShutdown(srv, timeout)
}

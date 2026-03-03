package shutdown

import (
	"context"
	"fmt"
	"k8soperation/global"
	"net/http"
	"os"
	"time"
)

// 优雅退出
// Shutdown 优雅关闭服务器的函数
// 参数:
//
//	k8soperation: HTTP服务器实例
//	quit: 接收系统信号的通道
//	done: 用于通知关闭完成的通道
/*
Shutdown 函数负责执行服务器的优雅关闭流程
它接收一个HTTP服务器实例、一个退出信号通道和一个完成通知通道作为参数
在接收到系统退出信号后，它会安全地关闭服务器并执行一些清理工作
*/
func Shutdown(server *http.Server, quit <-chan os.Signal, done chan<- struct{}) {
	//等待接收到退出信号：
	// 阻塞在此，直到从quit通道接收到系统发送的退出信号
	<-quit
	// 记录服务器开始关闭的信息
	global.Logger.Info("Server is shutting down...")

	// 创建一个30秒超时的上下文
	// 用于控制服务器关闭操作的最大等待时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // 确保上下文资源被释放

	// 禁用keep-alive连接
	// 这有助于更快地关闭现有连接
	server.SetKeepAlivesEnabled(false)

	// 执行服务器的优雅关闭
	// 这会处理所有活跃的请求，但不再接受新请求
	err := server.Shutdown(ctx)
	if err != nil {
		// 如果关闭过程中发生错误，记录致命错误
		global.Logger.Fatal(fmt.Sprintf("Could not gracefully shutdown the k8soperation: %v \n", err))
	}

	//do Something ：
	fmt.Println("do something start ..... ", time.Now()) // 记录开始执行额外任务的时间
	time.Sleep(5 * time.Second)                          // 模拟执行额外任务需要5秒
	fmt.Println("do something end ..... ", time.Now())   // 记录完成额外任务的时间

	close(done) // 通知调用方服务器已完全关闭
}

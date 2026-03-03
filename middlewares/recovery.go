package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

// Recovery 使用 zap.Error() 来记录 Panic 和 call stack
// Recovery 是一个 Gin 框架的中间件函数，用于捕获和处理 panic，防止服务器崩溃
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			// 使用 recover() 捕获可能的 panic
			// 这是一个错误处理机制，用于防止程序因未处理的异常而崩溃
			if err := recover(); err != nil {

				// 获取用户的请求信息，用于日志记录
				httpRequest, _ := httputil.DumpRequest(c.Request, true)

				// 链接中断，客户端中断连接为正常行为，不需要记录堆栈信息
				// 检查是否是管道错误或连接被对端重置
				var brokenPipe bool // 定义一个布尔变量用于标记是否是管道错误
				// 尝试将错误断言为网络操作错误
				if ne, ok := err.(*net.OpError); ok {
					// 尝试将网络操作错误中的错误字段断言为系统调用错误
					if se, ok := ne.Err.(*os.SyscallError); ok {
						// 将错误信息转换为小写形式，以便进行不区分大小写的比较
						// 检查错误信息中是否包含"broken pipe"或"connection reset by peer"
						errStr := strings.ToLower(se.Error())
						// 检查错误信息中是否包含"broken pipe"或"connection reset by peer"字符串  // 如果是管道错误或连接重置，设置标记为true
						if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer") {
							// 如果是上述两种错误之一，将brokenPipe标记为true
							brokenPipe = true
						}
					}
				}
				// 链接中断的情况
				// 检查是否为断开连接的情况
				if brokenPipe {
					// 记录错误日志，包含请求路径、时间戳、错误信息和请求数据
					global.Logger.Error(c.Request.URL.Path,
						zap.Time("time", time.Now()),                // 添加当前时间戳
						zap.Any("error", err),                       // 添加错误信息
						zap.String("requests", string(httpRequest)), // 添加请求数据
					)
					c.Error(err.(error)) // 返回错误信息
					c.Abort()            // 终止请求处理
					// 链接已断开，无法写状态码
					return
				}

				// 如果不是链接中断，就开始记录堆栈信息
				// 使用全局日志记录器记录panic恢复信息
				// 使用zap日志库记录结构化日志，包含多个字段信息
				global.Logger.Error("recovery from panic", // 主错误消息，表明发生了panic并已恢复
					zap.Time("time", time.Now()),                // 记录时间字段，使用当前时间
					zap.Any("error", err),                       // 记录错误信息字段，使用Any类型可以记录任意类型错误
					zap.String("requests", string(httpRequest)), // 请求信息
					zap.Stack("stacktrace"),                     // 调用堆栈信息
				)

				// 返回 500 状态码
				// 使用AbortWithStatusJSON方法返回一个500内部服务器错误状态码
				// 并附带JSON格式的错误信息
				// gin.H是Go语言中的一种便捷方式，用于创建map结构
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "服务器内部错误，请稍后再试", // 返回给客户端的错误提示信息
				})
			}
		}()
		// 调用Next()方法继续处理请求链中的下一个处理器
		// 这是Gin框架中的关键方法，用于将控制权传递给下一个中间件或路由处理器
		c.Next()
	}
}

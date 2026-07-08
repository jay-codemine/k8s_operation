package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Checks: 依赖注入（需要检查哪些下游就加到这里）
type Checks struct {
	DB *sql.DB
	// Redis *redis.Client
	// MQ    *amqp.Connection
	// 需要再加的依赖按需扩展
}

// Register: 注册健康检查路由
func Register(r *gin.Engine, c Checks) {
	// 使用路由组功能，创建一个以 "/healthz" 为前缀的路由组
	api := r.Group("/healthz")

	// 存活探针：只要进程活着能响应即可（不要检查DB等外部依赖）
	api.GET("/live", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	// 就绪探针：只有核心依赖就绪才返回200
	api.GET("/ready", func(ctx *gin.Context) {
		// 兜底：如果上游没带超时，这里统一加个短超时，避免探针阻塞
		// 从请求上下文中获取上下文信息
		reqCtx := ctx.Request.Context()
		// 检查上下文是否设置了截止时间
		if _, has := reqCtx.Deadline(); !has {
			// 如果没有设置截止时间，则创建一个带有300毫秒超时的上下文
			var cancel context.CancelFunc
			reqCtx, cancel = context.WithTimeout(reqCtx, 300*time.Millisecond)
			// 确保在函数返回时取消上下文，避免资源泄漏
			defer cancel()
		}

		// 1) DB（按需增加更多依赖检查）
		// 检查数据库连接是否初始化
		if c.DB == nil {
			// 如果数据库未初始化，返回服务不可用状态码和错误信息
			ctx.String(http.StatusServiceUnavailable, "db not initialized")
			return
		}
		// 尝试ping数据库以检查连接是否可用
		if err := c.DB.PingContext(reqCtx); err != nil {
			// 如果数据库连接不可用，返回服务不可用状态码和详细的错误信息
			ctx.String(http.StatusServiceUnavailable, "db not ready: %v", err)
			return
		}

		// 2) 其他依赖（示例）
		// if c.Redis == nil || c.Redis.Ping(reqCtx).Err() != nil { ... }
		// if c.MQ == nil || c.MQ.IsClosed() { ... }

		ctx.String(http.StatusOK, "ok")
	})
}

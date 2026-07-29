package health

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Checks 依赖注入（需要检查哪些下游就加到这里）
type Checks struct {
	DB *sql.DB
	// ReadyTimeout 就绪探针 DB Ping 的超时（<=0 时用默认 2s），从配置 Server.ReadinessTimeout 注入
	ReadyTimeout time.Duration
}

// defaultReadyTimeout 就绪探针默认超时（未配置时的兜底值）
const defaultReadyTimeout = 2 * time.Second

// dbProbeCache 缓存 DB ping 结果，避免每次探针请求都真实 ping 远程数据库
// K8s readinessProbe 默认每 10s 一次，远程 MySQL 网络波动频繁 → 间歇性 503
type dbProbeCache struct {
	mu        sync.RWMutex
	lastOk    bool
	lastCheck time.Time
	ttl       time.Duration // 缓存有效期，默认与 ping 超时一致或略长
}

var probeCache = &dbProbeCache{ttl: 10 * time.Second}

// Register 注册健康检查路由
func Register(r *gin.Engine, c Checks) {
	api := r.Group("/healthz")

	// 存活探针：只要进程活着能响应即可（不检查外部依赖）
	api.GET("/live", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	// 就绪探针：检查核心依赖，带缓存避免网络抖动误判
	api.GET("/ready", func(ctx *gin.Context) {
		timeout := c.ReadyTimeout
		if timeout <= 0 {
			timeout = defaultReadyTimeout
		}

		if c.DB == nil {
			ctx.String(http.StatusServiceUnavailable, "db not initialized")
			return
		}

		// 检查缓存：如果在 TTL 内且上次 OK，直接返回 200，跳过真实 ping
		probeCache.mu.RLock()
		cached := probeCache.lastOk && time.Since(probeCache.lastCheck) < probeCache.ttl
		probeCache.mu.RUnlock()
		if cached {
			ctx.String(http.StatusOK, "ok")
			return
		}

		// 真实 ping（带超时）
		reqCtx := ctx.Request.Context()
		if _, has := reqCtx.Deadline(); !has {
			var cancel context.CancelFunc
			reqCtx, cancel = context.WithTimeout(reqCtx, timeout)
			defer cancel()
		}

		err := c.DB.PingContext(reqCtx)
		ok := err == nil

		// 更新缓存
		probeCache.mu.Lock()
		probeCache.lastOk = ok
		probeCache.lastCheck = time.Now()
		probeCache.mu.Unlock()

		if !ok {
			ctx.String(http.StatusServiceUnavailable, "db not ready: %v", err)
			return
		}

		ctx.String(http.StatusOK, "ok")
	})
}

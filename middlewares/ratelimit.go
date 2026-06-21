package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// =====================================================================
// 基于令牌桶的 IP 限流中间件
// 使用 golang.org/x/time/rate（标准库级别，零依赖）
// =====================================================================

// ipEntry 存储每个 IP 对应的限流器 + 最近访问时间
type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ipRateLimiter 管理所有 IP 的限流器
type ipRateLimiter struct {
	mu    sync.Mutex
	ips   map[string]*ipEntry
	r     rate.Limit // 每秒产生的令牌数
	burst int        // 桶容量（最大并发突发量）
}

// newIPRateLimiter 创建并启动定期清理的限流器管理器
func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	rl := &ipRateLimiter{
		ips:   make(map[string]*ipEntry),
		r:     r,
		burst: burst,
	}
	go rl.cleanupLoop()
	return rl
}

// getLimiter 获取（或创建）该 IP 的限流器
func (rl *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	e, ok := rl.ips[ip]
	if !ok {
		e = &ipEntry{
			limiter:  rate.NewLimiter(rl.r, rl.burst),
			lastSeen: time.Now(),
		}
		rl.ips[ip] = e
	} else {
		e.lastSeen = time.Now()
	}
	return e.limiter
}

// cleanupLoop 每 5 分钟清理 10 分钟未访问的 IP，防止内存无限增长
func (rl *ipRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, e := range rl.ips {
			if now.Sub(e.lastSeen) > 10*time.Minute {
				delete(rl.ips, ip)
			}
		}
		rl.mu.Unlock()
	}
}

var (
	// apiLimiter：通用 API 限流
	// 每 IP 100 个令牌/秒，突发上限 200
	// 对于 DevOps 平台的前端操作已经足够宽松
	apiLimiter = newIPRateLimiter(100, 200)

	// loginLimiter：登录接口限流（防暴力破解）
	// 每 IP 每 6 秒 1 个令牌 = 10次/分钟，突发上限 5
	loginLimiter = newIPRateLimiter(rate.Every(6*time.Second), 5)
)

// RateLimit 通用 API 限流中间件（每 IP 100 req/s，burst 200）
// 建议注册在全局中间件链中（PrometheusMiddleware 之后）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !apiLimiter.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 42900,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// LoginRateLimit 登录接口专用限流中间件（每 IP 10次/分钟，防暴力破解）
// 建议只注册在 /auth/login 等认证路由上
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !loginLimiter.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 42901,
				"msg":  "登录尝试过于频繁，请 1 分钟后再试",
			})
			return
		}
		c.Next()
	}
}

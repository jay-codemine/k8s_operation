package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"k8soperation/middlewares"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestEngine 创建带有限流中间件的测试引擎
func newTestEngine(mw gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// ----------------------------------------------------------------
// 测试1：通用 API 限流 —— 正常速率（100 req/s burst 200），不应触发
// ----------------------------------------------------------------
func TestRateLimit_NormalRate_NoThrottle(t *testing.T) {
	r := newTestEngine(middlewares.RateLimit())

	var success int32
	for i := 0; i < 50; i++ { // 50次，远低于 burst 200
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:9000"
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			atomic.AddInt32(&success, 1)
		}
	}
	if int(success) != 50 {
		t.Errorf("正常速率下期望 50 次成功，实际 %d 次成功", success)
	}
	t.Logf("✅ 正常速率：%d/50 成功（无限流触发）", success)
}

// ----------------------------------------------------------------
// 测试2：通用 API 限流 —— 爆发超过 burst，应触发 429
// ----------------------------------------------------------------
func TestRateLimit_Burst_Triggers429(t *testing.T) {
	r := newTestEngine(middlewares.RateLimit())

	total := 300 // 超过 burst=200
	var success, throttled int32

	for i := 0; i < total; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234" // 固定 IP
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			atomic.AddInt32(&success, 1)
		} else if w.Code == http.StatusTooManyRequests {
			atomic.AddInt32(&throttled, 1)
		}
	}

	t.Logf("✅ 爆发测试：总请求=%d  成功=%d  被限流(429)=%d", total, success, throttled)
	if throttled == 0 {
		t.Errorf("❌ 期望触发 429 限流，但未触发（burst=200，总请求=%d）", total)
	}
	if int(success) > 200 {
		t.Errorf("❌ 成功次数 %d 超过 burst 200", success)
	}
}

// ----------------------------------------------------------------
// 测试3：不同 IP 互不影响（各自独立令牌桶）
// ----------------------------------------------------------------
func TestRateLimit_DifferentIPs_Independent(t *testing.T) {
	r := newTestEngine(middlewares.RateLimit())

	var wg sync.WaitGroup
	results := make([]int32, 3) // 3 个 IP

	for ipIdx, ip := range []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"} {
		wg.Add(1)
		go func(idx int, ipAddr string) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = ipAddr + ":8080"
				r.ServeHTTP(w, req)
				if w.Code == http.StatusOK {
					atomic.AddInt32(&results[idx], 1)
				}
			}
		}(ipIdx, ip)
	}
	wg.Wait()

	for i, cnt := range results {
		t.Logf("✅ IP-%d：%d/50 成功", i+1, cnt)
		if int(cnt) < 50 {
			t.Errorf("❌ IP-%d 期望 50 次成功，实际 %d", i+1, cnt)
		}
	}
}

// ----------------------------------------------------------------
// 测试4：登录限流 —— 第6次起应被限流（burst=5）
// ----------------------------------------------------------------
func TestLoginRateLimit_BurstExceeded(t *testing.T) {
	r := gin.New()
	r.Use(middlewares.LoginRateLimit())
	r.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	var success, throttled int32
	ip := "172.16.0.100:5000"

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = ip
		r.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			atomic.AddInt32(&success, 1)
		} else if w.Code == http.StatusTooManyRequests {
			atomic.AddInt32(&throttled, 1)
			t.Logf("  第 %d 次请求：429 限流 ✅", i+1)
		}
	}

	t.Logf("✅ 登录限流测试：总10次  成功=%d  被限流=%d", success, throttled)
	if throttled == 0 {
		t.Errorf("❌ 登录限流未触发：burst=5，请求10次，期望至少触发5次限流")
	}
	if int(success) > 5 {
		t.Errorf("❌ 成功次数 %d 超过 burst 5", success)
	}
}

// ----------------------------------------------------------------
// 测试5：限流后等待令牌恢复，再次请求应成功
// ----------------------------------------------------------------
func TestRateLimit_Recovery_AfterWait(t *testing.T) {
	r := gin.New()
	r.Use(middlewares.LoginRateLimit())
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	ip := "10.10.10.10:9090"

	// 先打满 burst
	var firstThrottled int32
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = ip
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			atomic.AddInt32(&firstThrottled, 1)
		}
	}
	t.Logf("  阶段1：打满后被限流 %d 次", firstThrottled)

	// 等待 6s，令牌桶补充 1 个（每 6s 产生 1 个令牌）
	t.Log("  等待 7s 令牌恢复...")
	time.Sleep(7 * time.Second)

	// 再发 1 次，应该成功
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api", nil)
	req.RemoteAddr = ip
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Log("✅ 等待恢复后请求成功（令牌桶已补充）")
	} else {
		t.Errorf("❌ 等待 7s 后请求仍被限流（状态码 %d）", w.Code)
	}
}

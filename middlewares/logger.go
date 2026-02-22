package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
)

// responseBodyWriter 用于包装 http.ResponseWriter，并同时记录响应体内容
type responseBodyWriter struct {
	body               *bytes.Buffer // 缓存响应体数据
	gin.ResponseWriter               // 原始的 ResponseWriter
}

// 注意：方法名必须大写 Write 才会覆盖接口方法
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // 先写缓存，便于后续日志或调试使用
	return w.ResponseWriter.Write(b)
}

// 覆盖 WriteString，防止有些地方直接写字符串
func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 包装 Writer，拦截响应体
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// 备份并复位请求体
		var requestBody []byte
		if c.Request.Body != nil {
			// c.Request.Body 是一个只读一次的流，这里先读出来
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 读完要复位，否则后续 handler 读不到
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 计时
		start := time.Now()

		// 放行，执行业务
		// 调用处理函数链中的下一个处理函数
		c.Next()

		// 计算请求处理耗时
		cost := time.Since(start)
		// 获取响应状态码
		status := c.Writer.Status()

		// 构造日志字段数组
		logFields := []zap.Field{
			zap.String("method", c.Request.Method),          // HTTP请求方法
			zap.String("path", c.Request.URL.Path),          // 请求路径
			zap.String("ip", c.ClientIP()),                  // 客户端IP地址
			zap.String("user-agent", c.Request.UserAgent()), // 用户代理
			zap.Int("status", status),                       // 响应状态码
			zap.Int64("latency_ms", cost.Milliseconds()),    // 请求处理耗时（毫秒）
		}

		// 仅在变更类方法时记录请求/响应体（保持你的原有逻辑）
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete {

			logFields = append(logFields,
				zap.String("requests-body", string(requestBody)),
				zap.String("response-body", w.body.String()),
			)
		}

		// 分级输出日志（使用 strconv.Itoa）
		switch {
		case status >= http.StatusInternalServerError:
			global.Logger.Error("HTTP Error "+strconv.Itoa(status), logFields...)
		case status >= http.StatusBadRequest:
			global.Logger.Warn("HTTP Warning "+strconv.Itoa(status), logFields...)
		default:
			global.Logger.Debug("HTTP Access Log", logFields...)
		}
	}
}

package middlewares

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8soperation/pkg/metrics"
)

// =====================================================================
// Prometheus 指标采集中间件
// 自动记录每个 HTTP 请求的计数、延迟、body 大小等指标
// =====================================================================

// pathNormalizers 用于将动态路径段归一化，避免指标基数爆炸
var pathNormalizers = []struct {
	pattern *regexp.Regexp
	replace string
}{
	// /api/v1/k8s/pod/xxx → /api/v1/k8s/pod/:name
	{regexp.MustCompile(`/c/\d+`), "/c/:id"},
	{regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), "/:uuid"},
}

// normalizePath 将 URL 路径归一化，消除高基数动态参数
func normalizePath(c *gin.Context) string {
	// 优先使用 Gin 的模板路径（如 /api/v1/k8s/pod/list）
	if p := c.FullPath(); p != "" {
		return p
	}

	// 对于未注册的路径，做正则归一化
	path := c.Request.URL.Path
	for _, n := range pathNormalizers {
		path = n.pattern.ReplaceAllString(path, n.replace)
	}

	// 限制路径深度，避免意外的高基数
	parts := strings.Split(path, "/")
	if len(parts) > 7 {
		path = strings.Join(parts[:7], "/")
	}
	return path
}

// PrometheusMiddleware Gin Prometheus 指标采集中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过 /metrics 和 /healthz 本身，避免自我循环
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/metrics") || strings.HasPrefix(path, "/healthz") {
			c.Next()
			return
		}

		// 当前并发请求 +1
		metrics.HTTPInflightRequests.Inc()
		defer metrics.HTTPInflightRequests.Dec()

		start := time.Now()

		// 执行请求
		c.Next()

		// 计算指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		normalPath := normalizePath(c)
		reqSize := float64(c.Request.ContentLength)
		if reqSize < 0 {
			reqSize = 0
		}
		respSize := float64(c.Writer.Size())
		if respSize < 0 {
			respSize = 0
		}

		// 写入 Prometheus 指标
		metrics.HTTPRequestsTotal.WithLabelValues(method, normalPath, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, normalPath).Observe(duration)
		metrics.HTTPRequestSizeBytes.WithLabelValues(method, normalPath).Observe(reqSize)
		metrics.HTTPResponseSizeBytes.WithLabelValues(method, normalPath).Observe(respSize)

		// 4xx/5xx → 错误计数
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			errType := "client_error"
			if statusCode >= 500 {
				errType = "server_error"
			}
			module := inferModule(normalPath)
			metrics.ErrorsTotal.WithLabelValues(module, errType).Inc()
		}
	}
}

// inferModule 根据路径推断所属模块
func inferModule(path string) string {
	switch {
	case strings.Contains(path, "/k8s/deployment") || strings.Contains(path, "/k8s/pod") ||
		strings.Contains(path, "/k8s/statefulset") || strings.Contains(path, "/k8s/daemonset") ||
		strings.Contains(path, "/k8s/job") || strings.Contains(path, "/k8s/cronjob"):
		return "workload"
	case strings.Contains(path, "/k8s/service") || strings.Contains(path, "/k8s/ingress"):
		return "network"
	case strings.Contains(path, "/k8s/configmap") || strings.Contains(path, "/k8s/secret"):
		return "config"
	case strings.Contains(path, "/k8s/pv") || strings.Contains(path, "/k8s/pvc") || strings.Contains(path, "/k8s/storageclass"):
		return "storage"
	case strings.Contains(path, "/k8s/crd") || strings.Contains(path, "/k8s/cr/"):
		return "crd"
	case strings.Contains(path, "/cicd"):
		return "cicd"
	case strings.Contains(path, "/auth"):
		return "auth"
	case strings.Contains(path, "/rbac") || strings.Contains(path, "/user"):
		return "rbac"
	case strings.Contains(path, "/monitoring"):
		return "monitoring"
	case strings.Contains(path, "/ai"):
		return "ai"
	case strings.Contains(path, "/image"):
		return "image"
	case strings.Contains(path, "/appstore"):
		return "appstore"
	default:
		return "platform"
	}
}

// =====================================================================
// K8s 操作埋点辅助函数（供 Controller/Service 层调用）
// =====================================================================

// RecordK8sAPICall 记录一次 K8s API 调用指标
//
//	clusterID: 集群 ID
//	resource:  资源类型（deployment, pod, service ...）
//	action:    操作类型（list, create, update, delete, get, yaml）
//	err:       如果有错误则传入
//	duration:  调用耗时
func RecordK8sAPICall(clusterID uint32, resource, action string, err error, duration time.Duration) {
	cid := fmt.Sprintf("%d", clusterID)
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.K8sAPICallsTotal.WithLabelValues(cid, resource, action, status).Inc()
	metrics.K8sAPICallDuration.WithLabelValues(cid, resource, action).Observe(duration.Seconds())
}

// RecordCRDOperation 记录 CRD/CR 操作
func RecordCRDOperation(resourceType, action string, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.CRDOperationsTotal.WithLabelValues(resourceType, action, status).Inc()
}

// RecordAuthLogin 记录登录事件
func RecordAuthLogin(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	metrics.AuthLoginTotal.WithLabelValues(status).Inc()
}

// RecordTokenValidation 记录 Token 校验
func RecordTokenValidation(status string) {
	metrics.AuthTokenValidationTotal.WithLabelValues(status).Inc()
}

// RecordPipelineRun 记录流水线执行
func RecordPipelineRun(pipelineName, status string, duration time.Duration) {
	metrics.CICDPipelineRunsTotal.WithLabelValues(pipelineName, status).Inc()
	if duration > 0 {
		metrics.CICDPipelineDuration.WithLabelValues(pipelineName).Observe(duration.Seconds())
	}
}

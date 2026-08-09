package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
)

// ========== 审计中间件 ==========
// 自动记录所有写操作（POST/PUT/DELETE）到审计日志表
// GET 请求默认不记录，可通过配置开启

// AuditConfig 审计中间件配置
type AuditConfig struct {
	// 是否记录 GET 请求（默认不记录，减少噪音）
	RecordGET bool
	// 需要排除的路径前缀（如健康检查、静态资源等）
	ExcludePaths []string
	// 请求体最大记录长度（超出截断，防止大文件上传撑爆审计表）
	MaxBodySize int
}

// DefaultAuditConfig 默认审计配置
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		RecordGET: false,
		ExcludePaths: []string{
			"/api/v1/platform/health",
			"/api/v1/monitoring/",
			"/swagger/",
			"/api/v1/ai/chat",
		},
		MaxBodySize: 4096, // 4KB
	}
}

// Audit 审计日志中间件
func Audit(cfg *AuditConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultAuditConfig()
	}

	return func(c *gin.Context) {
		// 跳过不需要审计的请求
		if shouldSkipAudit(c, cfg) {
			c.Next()
			return
		}

		// 记录请求体（仅写操作）
		var requestBody string
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "multipart/form-data") {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(bodyBytes) > cfg.MaxBodySize {
					requestBody = string(bodyBytes[:cfg.MaxBodySize]) + "...[truncated]"
				} else {
					requestBody = string(bodyBytes)
				}
				// 脱敏处理（移除密码等敏感字段）
				requestBody = sanitizeBody(requestBody)
			} else {
				requestBody = "[multipart/form-data upload]"
			}
		}

		start := time.Now()

		// 执行后续 handler
		c.Next()

		// ⭐ 关键：在当前 goroutine 内提取所有需要的数据（gin.Context 不可跨 goroutine 访问）
		userID, _ := c.Get("user_id")
		username, _ := c.Get("current_user_name")
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		path := c.Request.URL.Path
		method := c.Request.Method
		responseCode := c.Writer.Status()
		clusterIDHeader := c.GetHeader("X-Cluster-ID")
		clusterIDQuery := c.Query("cluster_id")
		namespace := c.Query("namespace")
		if namespace == "" {
			namespace = c.Param("namespace")
		}
		duration := time.Since(start)

		// 异步写入数据库（不阻塞响应）
		go writeAuditLog(userID, username, clientIP, userAgent, path, method,
			responseCode, clusterIDHeader, clusterIDQuery, namespace,
			requestBody, duration)
	}
}

// shouldSkipAudit 判断是否跳过审计
func shouldSkipAudit(c *gin.Context, cfg *AuditConfig) bool {
	path := c.Request.URL.Path

	// 跳过 GET 请求（如果未配置记录）
	if !cfg.RecordGET && c.Request.Method == http.MethodGet {
		return true
	}

	// 跳过 OPTIONS 预检请求
	if c.Request.Method == http.MethodOptions {
		return true
	}

	// 跳过排除路径
	for _, p := range cfg.ExcludePaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// writeAuditLog 异步写入审计日志（参数均为值类型，不依赖 gin.Context）
func writeAuditLog(userIDVal, usernameVal interface{}, clientIP, userAgent, path, method string,
	responseCode int, clusterIDHeader, clusterIDQuery, namespace, requestBody string, duration time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			global.Logger.Error("审计日志记录 panic", zap.Any("recover", r))
		}
	}()

	if global.DB == nil {
		return
	}

	// 提取用户信息
	uid := int64(0)
	if id, ok := userIDVal.(int64); ok {
		uid = id
	}
	uname := ""
	if name, ok := usernameVal.(string); ok {
		uname = name
	}

	// 解析操作信息
	module, action, targetType, actionDisplay := parseRouteInfo(path, method)

	// 提取集群信息
	var clusterID *int64
	if clusterIDHeader != "" {
		if cid, err := strconv.ParseInt(clusterIDHeader, 10, 64); err == nil {
			clusterID = &cid
		}
	}
	if clusterIDQuery != "" && clusterID == nil {
		if cid, err := strconv.ParseInt(clusterIDQuery, 10, 64); err == nil {
			clusterID = &cid
		}
	}

	// 响应状态
	status := models.AuditStatusSuccess
	var errMsg string
	if responseCode >= 400 {
		status = models.AuditStatusFailed
		errMsg = "HTTP " + strconv.Itoa(responseCode)
	}

	// 构建审计日志
	auditLog := &models.AuditLog{
		UserID:        uid,
		Username:      uname,
		UserIP:        clientIP,
		UserAgent:     truncateString(userAgent, 500),
		Action:        action,
		ActionDisplay: actionDisplay,
		Module:        module,
		TargetType:    targetType,
		TargetName:    extractTargetName(path),
		RequestURI:    truncateString(path, 500),
		RequestMethod: method,
		RequestBody:   requestBody,
		ResponseCode:  responseCode,
		Status:        status,
		ErrorMessage:  errMsg,
		ClusterID:     clusterID,
		Namespace:     namespace,
		DurationMs:    int(duration.Milliseconds()),
		CreatedAt:     time.Now().Unix(),
	}

	// 写入数据库（审计日志为跨租户后台写入，使用 Services 层）
	if err := services.NewBackgroundServices().AuditLogRecord(context.Background(), auditLog); err != nil {
		global.Logger.Error("写入审计日志失败", zap.Error(err))
	}
}

// parseRouteInfo 根据路由路径解析模块、操作、目标类型
func parseRouteInfo(path, method string) (module, action, targetType, actionDisplay string) {
	// 默认值
	module = "platform"
	action = methodToAction(method)
	targetType = ""
	actionDisplay = ""

	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) == 0 {
		return
	}

	// 路由映射
	switch {
	case strings.Contains(path, "/auth/login"):
		module, action, targetType, actionDisplay = "auth", "login", "session", "用户登录"
	case strings.Contains(path, "/auth/logout"):
		module, action, targetType, actionDisplay = "auth", "logout", "session", "用户登出"
	case strings.Contains(path, "/k8s/deployment"):
		module, targetType = "workload", "deployment"
		actionDisplay = action + " Deployment"
	case strings.Contains(path, "/k8s/statefulset"):
		module, targetType = "workload", "statefulset"
		actionDisplay = action + " StatefulSet"
	case strings.Contains(path, "/k8s/daemonset"):
		module, targetType = "workload", "daemonset"
		actionDisplay = action + " DaemonSet"
	case strings.Contains(path, "/k8s/pod"):
		module, targetType = "workload", "pod"
		actionDisplay = action + " Pod"
	case strings.Contains(path, "/k8s/job"):
		module, targetType = "workload", "job"
		actionDisplay = action + " Job"
	case strings.Contains(path, "/k8s/cronjob"):
		module, targetType = "workload", "cronjob"
		actionDisplay = action + " CronJob"
	case strings.Contains(path, "/k8s/service"):
		module, targetType = "network", "service"
		actionDisplay = action + " Service"
	case strings.Contains(path, "/k8s/ingress"):
		module, targetType = "network", "ingress"
		actionDisplay = action + " Ingress"
	case strings.Contains(path, "/k8s/configmap"):
		module, targetType = "config", "configmap"
		actionDisplay = action + " ConfigMap"
	case strings.Contains(path, "/k8s/secret"):
		module, targetType = "config", "secret"
		actionDisplay = action + " Secret"
	case strings.Contains(path, "/k8s/pv"):
		module, targetType = "storage", "pv"
		actionDisplay = action + " PV"
	case strings.Contains(path, "/k8s/pvc"):
		module, targetType = "storage", "pvc"
		actionDisplay = action + " PVC"
	case strings.Contains(path, "/k8s/storageclass"):
		module, targetType = "storage", "storageclass"
		actionDisplay = action + " StorageClass"
	case strings.Contains(path, "/k8s/node"):
		module, targetType = "cluster", "node"
		actionDisplay = action + " Node"
	case strings.Contains(path, "/k8s/namespace"):
		module, targetType = "cluster", "namespace"
		actionDisplay = action + " Namespace"
	case strings.Contains(path, "/k8s/cluster"):
		module, targetType = "cluster", "cluster"
		actionDisplay = action + " 集群"
	case strings.Contains(path, "/k8s/cicd"):
		module, targetType = "cicd", "pipeline"
		actionDisplay = action + " 流水线"
	case strings.Contains(path, "/rbac"):
		module, targetType = "rbac", "role"
		actionDisplay = action + " 权限"
	case strings.Contains(path, "/user"):
		module, targetType = "rbac", "user"
		actionDisplay = action + " 用户"
	case strings.Contains(path, "/platform/settings"):
		module, targetType = "platform", "settings"
		actionDisplay = "更新平台设置"
	case strings.Contains(path, "/platform/appstore"):
		module, targetType = "platform", "appstore"
		actionDisplay = action + " 应用商城"
	case strings.Contains(path, "/ai/"):
		module, targetType = "ai", "assistant"
		actionDisplay = "AI助手操作"
	case strings.Contains(path, "/image/"):
		module, targetType = "image", "registry"
		actionDisplay = action + " 镜像"
	case strings.Contains(path, "/monitoring/"):
		module, targetType = "monitoring", "alert"
		actionDisplay = action + " 监控"
	}

	return
}

// methodToAction HTTP 方法映射为操作类型
func methodToAction(method string) string {
	switch method {
	case http.MethodPost:
		return models.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return models.AuditActionUpdate
	case http.MethodDelete:
		return models.AuditActionDelete
	default:
		return models.AuditActionView
	}
}

// extractTargetName 从路径中提取目标名称
func extractTargetName(path string) string {
	parts := strings.Split(path, "/")
	// 取最后一个非空路径段作为目标名
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !isCommonPathSegment(parts[i]) {
			return parts[i]
		}
	}
	return ""
}

// isCommonPathSegment 判断是否为常见路径段（非目标名）
func isCommonPathSegment(s string) bool {
	commons := []string{"api", "v1", "k8s", "list", "create", "update", "delete", "detail"}
	for _, c := range commons {
		if s == c {
			return true
		}
	}
	return false
}

// sanitizeBody 脱敏请求体中的敏感信息
func sanitizeBody(body string) string {
	if body == "" {
		return body
	}

	// 尝试解析 JSON 并脱敏
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	sensitiveKeys := []string{"password", "secret", "token", "kube_config", "kubeconfig",
		"access_key_secret", "access_key_id", "webhook"}
	for _, key := range sensitiveKeys {
		if _, exists := data[key]; exists {
			data[key] = "***"
		}
	}

	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return string(sanitized)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

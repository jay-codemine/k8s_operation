package middlewares

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// K8sError K8s 错误处理中间件，解析错误并返回友好的错误信息
func K8sError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		ge := c.Errors.Last()
		if ge == nil || ge.Err == nil {
			return
		}

		err := ge.Err

		// 如果已经返回过响应，就不要再覆盖
		if c.Writer.Written() {
			return
		}

		// 解析 K8s API 错误，返回友好的错误信息
		httpCode, errCode, friendlyMsg, detail := parseK8sError(err)

		c.AbortWithStatusJSON(httpCode, gin.H{
			"code":    errCode,
			"msg":     friendlyMsg,
			"details": detail,
		})
	}
}

// parseK8sError 解析 K8s API 错误，返回 HTTP 状态码、错误码、友好消息和详细信息
func parseK8sError(err error) (httpCode int, errCode int, friendlyMsg string, detail string) {
	// 默认值
	httpCode = http.StatusInternalServerError
	errCode = 500
	friendlyMsg = "操作失败"
	detail = ""

	// 尝试解析为 K8s StatusError
	if statusErr, ok := err.(*apierrors.StatusError); ok {
		status := statusErr.ErrStatus

		// 根据 K8s 错误类型返回友好信息
		switch {
		case apierrors.IsNotFound(err):
			httpCode = http.StatusNotFound
			errCode = 404
			friendlyMsg = "资源不存在"
			detail = extractResourceInfo(status.Message)

		case apierrors.IsAlreadyExists(err):
			httpCode = http.StatusConflict
			errCode = 409
			friendlyMsg = "资源已存在"
			detail = extractResourceInfo(status.Message)

		case apierrors.IsInvalid(err):
			httpCode = http.StatusBadRequest
			errCode = 400
			friendlyMsg = "参数无效"
			detail = extractValidationError(status.Message)

		case apierrors.IsForbidden(err):
			httpCode = http.StatusForbidden
			errCode = 403
			friendlyMsg = "权限不足"
			detail = extractResourceInfo(status.Message)

		case apierrors.IsUnauthorized(err):
			httpCode = http.StatusUnauthorized
			errCode = 401
			friendlyMsg = "未授权访问"
			detail = "请检查集群认证配置"

		case apierrors.IsConflict(err):
			httpCode = http.StatusConflict
			errCode = 409
			friendlyMsg = "资源冲突"
			detail = "资源已被修改，请刷新后重试"

		case apierrors.IsTimeout(err):
			httpCode = http.StatusGatewayTimeout
			errCode = 504
			friendlyMsg = "请求超时"
			detail = "集群响应超时，请稍后重试"

		case apierrors.IsServerTimeout(err):
			httpCode = http.StatusGatewayTimeout
			errCode = 504
			friendlyMsg = "服务器超时"
			detail = "集群服务器响应超时"

		case apierrors.IsTooManyRequests(err):
			httpCode = http.StatusTooManyRequests
			errCode = 429
			friendlyMsg = "请求过于频繁"
			detail = "请稍后重试"

		case apierrors.IsServiceUnavailable(err):
			httpCode = http.StatusServiceUnavailable
			errCode = 503
			friendlyMsg = "集群服务不可用"
			detail = "请检查集群状态"

		default:
			// 其他错误，尝试提取核心信息
			friendlyMsg = "操作失败"
			detail = extractCoreMessage(status.Message)
		}

		return
	}

	// 非 K8s API 错误，尝试从错误消息中提取有用信息
	errMsg := err.Error()
	detail = extractCoreMessage(errMsg)

	// 常见错误模式匹配
	switch {
	case strings.Contains(errMsg, "not found"):
		httpCode = http.StatusNotFound
		errCode = 404
		friendlyMsg = "资源不存在"

	case strings.Contains(errMsg, "already exists"):
		httpCode = http.StatusConflict
		errCode = 409
		friendlyMsg = "资源已存在"

	case strings.Contains(errMsg, "is invalid"):
		httpCode = http.StatusBadRequest
		errCode = 400
		friendlyMsg = "参数无效"
		detail = extractValidationError(errMsg)

	case strings.Contains(errMsg, "forbidden"):
		httpCode = http.StatusForbidden
		errCode = 403
		friendlyMsg = "权限不足"

	case strings.Contains(errMsg, "port is already allocated"):
		httpCode = http.StatusConflict
		errCode = 409
		friendlyMsg = "端口已被占用"
		detail = extractPortConflict(errMsg)

	case strings.Contains(errMsg, "connection refused"):
		httpCode = http.StatusServiceUnavailable
		errCode = 503
		friendlyMsg = "无法连接到集群"
		detail = "请检查集群网络连接"

	case strings.Contains(errMsg, "timeout"):
		httpCode = http.StatusGatewayTimeout
		errCode = 504
		friendlyMsg = "请求超时"
		detail = "操作超时，请稍后重试"
	}

	return
}

// extractResourceInfo 提取资源信息（如 "deployments.apps 'nginx' not found"）
func extractResourceInfo(msg string) string {
	// 匹配模式: "resource 'name' not found" 或 "resource/name"
	re := regexp.MustCompile(`["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(msg, -1)
	if len(matches) > 0 {
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			if len(m) > 1 {
				names = append(names, m[1])
			}
		}
		if len(names) > 0 {
			return strings.Join(names, ", ")
		}
	}
	return extractCoreMessage(msg)
}

// extractValidationError 提取验证错误信息，返回具体的错误原因
func extractValidationError(msg string) string {
	var result strings.Builder

	// 匹配所有 "field: Invalid value: value: reason" 格式的错误
	re := regexp.MustCompile(`([a-zA-Z0-9_.\[\]]+):\s*Invalid value:\s*([^:]+):\s*([^,;]+)`)
	matches := re.FindAllStringSubmatch(msg, -1)

	if len(matches) > 0 {
		for i, m := range matches {
			if len(m) >= 4 {
				field := m[1]    // 字段名
				value := m[2]    // 无效值
				reason := m[3]   // 原因

				// 翻译常见的错误原因
				reason = translateK8sReason(reason)

				if i > 0 {
					result.WriteString("; ")
				}
				result.WriteString(translateFieldName(field))
				result.WriteString(" ")
				result.WriteString(value)
				result.WriteString(": ")
				result.WriteString(reason)
			}
		}
		return result.String()
	}

	// 匹配 "field: reason" 格式
	re2 := regexp.MustCompile(`([a-zA-Z0-9_.\[\]]+):\s*(.+)`)
	if matches := re2.FindStringSubmatch(msg); len(matches) >= 3 {
		field := translateFieldName(matches[1])
		reason := translateK8sReason(matches[2])
		return field + ": " + reason
	}

	return extractCoreMessage(msg)
}

// translateFieldName 翻译字段名为友好的中文
func translateFieldName(field string) string {
	fieldMap := map[string]string{
		"spec.replicas":                    "副本数",
		"spec.ports[0].nodePort":           "NodePort 端口",
		"spec.ports[0].port":               "服务端口",
		"spec.ports[0].targetPort":         "目标端口",
		"spec.selector":                    "选择器",
		"spec.template.spec.containers":    "容器配置",
		"spec.template.spec.containers[0]": "容器",
		"metadata.name":                    "名称",
		"metadata.namespace":               "命名空间",
		"metadata.labels":                  "标签",
		"spec.containers[0].image":         "镜像",
		"spec.containers[0].name":          "容器名",
		"spec.containers[0].ports":         "容器端口",
		"spec.volumes":                     "卷配置",
		"spec.type":                        "服务类型",
		"spec.schedule":                    "调度时间",
		"spec.jobTemplate":                 "Job 模板",
	}

	if translated, ok := fieldMap[field]; ok {
		return translated
	}

	// 处理数组索引，如 spec.ports[0] -> 端口配置[0]
	if strings.HasPrefix(field, "spec.ports[") {
		return "端口配置" + field[len("spec.ports"):]
	}
	if strings.HasPrefix(field, "spec.containers[") {
		return "容器" + field[len("spec.containers"):]
	}
	if strings.HasPrefix(field, "spec.template.spec.") {
		return "模板." + field[len("spec.template.spec."):]
	}

	return field
}

// translateK8sReason 翻译常见的 K8s 错误原因
func translateK8sReason(reason string) string {
	reason = strings.TrimSpace(reason)

	// 常见错误翻译映射
	reasonMap := map[string]string{
		"provided port is already allocated":                      "端口已被占用",
		"must be greater than or equal to 0":                       "必须大于或等于 0",
		"must be no more than 65535":                               "不能超过 65535",
		"must be between 1 and 65535, inclusive":                   "必须在 1-65535 之间",
		"must be between 30000-32767":                              "必须在 30000-32767 之间",
		"Required value":                                           "必填字段",
		"required field":                                           "必填字段",
		"may not be empty":                                         "不能为空",
		"must be non-empty":                                        "不能为空",
		"a DNS-1123 label must consist of lower case alphanumeric": "必须是小写字母数字和连字符",
		"must start with a letter or number":                       "必须以字母或数字开头",
		"must end with an alphanumeric character":                  "必须以字母或数字结尾",
		"must be no more than 63 characters":                       "不能超过 63 个字符",
		"name is required":                                         "名称不能为空",
		"namespace is required":                                    "命名空间不能为空",
		"image is required":                                        "镜像不能为空",
		"containers is required":                                   "容器配置不能为空",
		"Duplicate value":                                          "值重复",
		"cannot be changed":                                        "不可修改",
		"field is immutable":                                       "字段不可修改",
		"selector does not match":                                  "选择器不匹配",
	}

	for pattern, translation := range reasonMap {
		if strings.Contains(strings.ToLower(reason), strings.ToLower(pattern)) {
			return translation
		}
	}

	// 限制长度
	if len(reason) > 100 {
		reason = reason[:100] + "..."
	}

	return reason
}

// extractPortConflict 提取端口冲突信息
func extractPortConflict(msg string) string {
	// 匹配模式: "Invalid value: 30080: provided port is already allocated"
	re := regexp.MustCompile(`Invalid value: (\d+): (.+)`)
	if matches := re.FindStringSubmatch(msg); len(matches) > 2 {
		return "端口 " + matches[1] + " " + matches[2]
	}
	return extractCoreMessage(msg)
}

// extractCoreMessage 提取核心错误信息，移除堆栈信息和冗余内容
func extractCoreMessage(msg string) string {
	// 移除堆栈信息（从 github.com 开始）
	if idx := strings.Index(msg, "github.com"); idx > 0 {
		msg = strings.TrimSpace(msg[:idx])
	}

	// 移除 JSON 格式部分
	if idx := strings.Index(msg, `","response-body":`); idx > 0 {
		msg = strings.TrimSpace(msg[:idx])
	}

	// 提取 "failed to xxx: reason" 格式中的 reason
	if strings.Contains(msg, "failed to") {
		parts := strings.SplitN(msg, ": ", 3)
		if len(parts) >= 3 {
			// 返回最后一部分作为核心错误
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}

	// 限制长度，避免过长
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}

	return msg
}

// Package metrics 平台可观测性埋点
// 提供 Prometheus 指标采集，覆盖 HTTP 请求、K8s 操作、CICD、认证、系统运行时等维度
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =====================================================================
// 1. HTTP 请求指标（全局，由 Gin 中间件自动采集）
// =====================================================================

// HTTPRequestsTotal HTTP 请求总量计数器
// 标签: method, path, status
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "HTTP 请求总量",
	},
	[]string{"method", "path", "status"},
)

// HTTPRequestDuration HTTP 请求延迟直方图（秒）
// 标签: method, path
var HTTPRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP 请求延迟分布（秒）",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"method", "path"},
)

// HTTPRequestSizeBytes 请求体大小直方图
var HTTPRequestSizeBytes = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "http",
		Name:      "request_size_bytes",
		Help:      "HTTP 请求体大小分布（字节）",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B ~ 100MB
	},
	[]string{"method", "path"},
)

// HTTPResponseSizeBytes 响应体大小直方图
var HTTPResponseSizeBytes = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "http",
		Name:      "response_size_bytes",
		Help:      "HTTP 响应体大小分布（字节）",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
	},
	[]string{"method", "path"},
)

// HTTPInflightRequests 当前正在处理的请求数
var HTTPInflightRequests = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "http",
		Name:      "inflight_requests",
		Help:      "当前正在处理的 HTTP 请求数",
	},
)

// =====================================================================
// 2. K8s 集群操作指标
// =====================================================================

// K8sAPICallsTotal K8s API 调用总量
// 标签: cluster_id, resource, action, status
var K8sAPICallsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "k8s",
		Name:      "api_calls_total",
		Help:      "K8s API 调用总量",
	},
	[]string{"cluster_id", "resource", "action", "status"},
)

// K8sAPICallDuration K8s API 调用延迟（秒）
var K8sAPICallDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "k8s",
		Name:      "api_call_duration_seconds",
		Help:      "K8s API 调用延迟分布（秒）",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
	},
	[]string{"cluster_id", "resource", "action"},
)

// K8sClusterStatus 集群连接状态 (1=正常, 0=异常)
var K8sClusterStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "k8s",
		Name:      "cluster_status",
		Help:      "K8s 集群连接状态 (1=正常, 0=异常)",
	},
	[]string{"cluster_id", "cluster_name"},
)

// =====================================================================
// 3. CICD 流水线指标
// =====================================================================

// CICDPipelineRunsTotal 流水线运行总量
var CICDPipelineRunsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "cicd",
		Name:      "pipeline_runs_total",
		Help:      "流水线运行总量",
	},
	[]string{"pipeline_name", "status"}, // status: success, failed, aborted
)

// CICDPipelineDuration 流水线执行耗时（秒）
var CICDPipelineDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "cicd",
		Name:      "pipeline_duration_seconds",
		Help:      "流水线执行耗时分布（秒）",
		Buckets:   []float64{10, 30, 60, 120, 300, 600, 1200, 1800, 3600},
	},
	[]string{"pipeline_name"},
)

// CICDBuildQueueSize 构建队列大小
var CICDBuildQueueSize = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "cicd",
		Name:      "build_queue_size",
		Help:      "当前构建队列中等待的任务数",
	},
)

// =====================================================================
// 4. CRD/CR 动态资源指标
// =====================================================================

// CRDOperationsTotal CRD/CR 操作计数
var CRDOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "crd",
		Name:      "operations_total",
		Help:      "CRD/CR 动态资源操作总量",
	},
	[]string{"resource_type", "action", "status"}, // resource_type: crd/cr, action: list/create/update/delete/dryrun
)

// =====================================================================
// 5. 认证与权限指标
// =====================================================================

// AuthLoginTotal 登录计数
var AuthLoginTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "auth",
		Name:      "login_total",
		Help:      "用户登录总量",
	},
	[]string{"status"}, // success, failed
)

// AuthTokenValidationTotal Token 校验计数
var AuthTokenValidationTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "auth",
		Name:      "token_validation_total",
		Help:      "JWT Token 校验总量",
	},
	[]string{"status"}, // success, expired, invalid
)

// RBACDeniedTotal 权限拒绝次数
var RBACDeniedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "rbac",
		Name:      "denied_total",
		Help:      "权限拒绝总量",
	},
	[]string{"user", "resource", "action"},
)

// =====================================================================
// 6. 数据库指标
// =====================================================================

// DBQueryTotal 数据库查询计数
var DBQueryTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "db",
		Name:      "query_total",
		Help:      "数据库操作总量",
	},
	[]string{"table", "operation"}, // operation: select/insert/update/delete
)

// DBQueryDuration 数据库查询耗时
var DBQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "k8sop",
		Subsystem: "db",
		Name:      "query_duration_seconds",
		Help:      "数据库操作延迟分布（秒）",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5},
	},
	[]string{"table", "operation"},
)

// DBConnectionPoolActive 数据库连接池活跃连接数
var DBConnectionPoolActive = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "db",
		Name:      "connection_pool_active",
		Help:      "数据库连接池活跃连接数",
	},
)

// =====================================================================
// 7. Panic / 错误指标
// =====================================================================

// PanicsTotal Panic 恢复计数
var PanicsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "runtime",
		Name:      "panics_total",
		Help:      "panic 恢复总次数",
	},
)

// ErrorsTotal 业务错误计数（按模块分类）
var ErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "k8sop",
		Subsystem: "runtime",
		Name:      "errors_total",
		Help:      "业务错误总量",
	},
	[]string{"module", "error_type"},
)

// =====================================================================
// 8. 应用信息
// =====================================================================

// AppInfo 应用版本信息 Gauge（常量标签 = 1）
var AppInfo = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "app",
		Name:      "info",
		Help:      "应用版本信息（常量值 = 1）",
	},
	[]string{"version", "go_version", "build_time"},
)

// WebSocketConnections 当前 WebSocket 连接数
var WebSocketConnections = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "k8sop",
		Subsystem: "ws",
		Name:      "connections",
		Help:      "当前 WebSocket 连接数",
	},
)

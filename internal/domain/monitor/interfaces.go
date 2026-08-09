package monitor

import (
	"context"
	"time"
)

// =========================================================================
// 跨域接口 — Monitor 域对外暴露的契约
// 其他域依赖这些接口，而非具体实现
// =========================================================================

// LogQuerier Loki 日志查询能力（AIOps 域依赖此接口）
type LogQuerier interface {
	IsEnabled() bool
	QueryLogs(ctx context.Context, query string, start, end time.Time, limit int, direction string) (*LogQueryResult, error)
}

// PrometheusURLResolver 解析 Prometheus 数据源地址
// PlatformHealth / Canary 等域依赖此接口
type PrometheusURLResolver interface {
	ResolvePrometheusURL(ctx context.Context) string
}

// RoutePolicyResolver 根据告警规则匹配通知路由策略
// AlertEvalWorker 依赖此接口
type RoutePolicyResolver interface {
	ResolveRoutePolicyChannels(ctx context.Context, rule *AlertRule) string
}

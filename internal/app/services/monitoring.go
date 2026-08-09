package services

import (
	dm "k8soperation/internal/domain/monitor"
	"k8soperation/global"
)

// PrometheusSvc 返回租户隔离的 Prometheus 监控服务（HTTP 请求使用）
func (s *Services) PrometheusSvc(prometheusURL string) *dm.PrometheusService {
	return dm.NewPrometheusService(s.db, prometheusURL, monitoringTimeout())
}

func monitoringTimeout() int {
	if global.MonitoringSetting != nil {
		return global.MonitoringSetting.QueryTimeout
	}
	return 0
}

// WithDatasourceID 把前端切换的 datasource_id 注入 ctx
var WithDatasourceID = dm.WithDatasourceID


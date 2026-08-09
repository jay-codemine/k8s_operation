package services

import dm "k8soperation/internal/domain/monitor"

// LokiSvc 返回租户隔离的 Loki 日志服务（HTTP 请求使用）
func (s *Services) LokiSvc(lokiURL string) *dm.LokiService {
	return dm.NewLokiService(s.db, lokiURL, monitoringTimeout())
}


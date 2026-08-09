package services

import (
	dm "k8soperation/internal/domain/monitor"
	"k8soperation/internal/infra/persistence"
)

func (s *Services) monitorCRUDSvc() *dm.MonitorCRUDService {
	return dm.NewMonitorCRUDService(s.db, persistence.NewMonitorRepository(s.db))
}

// MonitorCRUDSvc 返回租户隔离的监控 CRUD 服务（HTTP/Worker/Bootstrap 使用）
func (s *Services) MonitorCRUDSvc() *dm.MonitorCRUDService {
	return s.monitorCRUDSvc()
}

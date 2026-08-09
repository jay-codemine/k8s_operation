package services

import (
	"context"

	"k8soperation/internal/app/models"
	"k8soperation/internal/domain/audit"
	"k8soperation/internal/infra/persistence"
)

// ========== 审计日志 Service（委托到 domain/audit.AuditService）==========

func (s *Services) auditSvc() *audit.AuditService {
	return audit.NewAuditService(persistence.NewAuditLogRepository(s.db), s.db)
}

// AuditLogList 查询审计日志列表
func (s *Services) AuditLogList(ctx context.Context, query *models.AuditLogQuery) (*models.AuditLogListResponse, error) {
	return s.auditSvc().List(ctx, query)
}

// AuditLogGetByID 获取审计日志详情
func (s *Services) AuditLogGetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	return s.auditSvc().GetByID(ctx, id)
}

// AuditLogGetStatistics 获取审计统计信息
func (s *Services) AuditLogGetStatistics(ctx context.Context) (*models.AuditStatistics, error) {
	return s.auditSvc().GetStatistics(ctx)
}

// AuditLogGetRetention 获取审计日志保留策略
func (s *Services) AuditLogGetRetention(ctx context.Context) (*models.AuditRetentionPolicy, error) {
	return s.auditSvc().GetRetention(ctx)
}

// AuditLogUpdateRetention 更新审计日志保留策略
func (s *Services) AuditLogUpdateRetention(ctx context.Context, req *models.AuditRetentionUpdateReq) error {
	return s.auditSvc().UpdateRetention(ctx, req)
}

// AuditLogCleanup 手动触发清理过期审计日志
func (s *Services) AuditLogCleanup(ctx context.Context) (int64, error) {
	return s.auditSvc().CleanupByPolicy(ctx)
}

// AuditLogExport 导出审计日志
func (s *Services) AuditLogExport(ctx context.Context, query *models.AuditLogQuery) ([]*models.AuditLog, error) {
	return s.auditSvc().Export(ctx, query)
}

// AuditLogRecord 记录单条审计日志
func (s *Services) AuditLogRecord(ctx context.Context, log *models.AuditLog) error {
	return s.auditSvc().Record(ctx, log)
}

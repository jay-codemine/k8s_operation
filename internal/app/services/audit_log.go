package services

import (
	"context"
	"k8soperation/internal/app/models"
)

// ========== 审计日志 Service ==========

// AuditLogList 查询审计日志列表
func (s *Services) AuditLogList(ctx context.Context, query *models.AuditLogQuery) (*models.AuditLogListResponse, error) {
	return s.dao.AuditLogList(ctx, query)
}

// AuditLogGetByID 获取审计日志详情
func (s *Services) AuditLogGetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	return s.dao.AuditLogGetByID(ctx, id)
}

// AuditLogGetStatistics 获取审计统计信息
func (s *Services) AuditLogGetStatistics(ctx context.Context) (*models.AuditStatistics, error) {
	return s.dao.AuditLogGetStatistics(ctx)
}

// AuditLogGetRetention 获取审计日志保留策略
func (s *Services) AuditLogGetRetention(ctx context.Context) (*models.AuditRetentionPolicy, error) {
	return s.dao.AuditLogGetRetentionPolicy(ctx)
}

// AuditLogUpdateRetention 更新审计日志保留策略
func (s *Services) AuditLogUpdateRetention(ctx context.Context, req *models.AuditRetentionUpdateReq) error {
	return s.dao.AuditLogUpdateRetentionPolicy(ctx, req)
}

// AuditLogCleanup 手动触发清理过期审计日志
func (s *Services) AuditLogCleanup(ctx context.Context) (int64, error) {
	policy, err := s.dao.AuditLogGetRetentionPolicy(ctx)
	if err != nil {
		return 0, err
	}
	if policy.IsPermanent || policy.RetentionDays == 0 {
		return 0, nil // 永久保留不清理
	}
	return s.dao.AuditLogCleanup(ctx, policy.RetentionDays)
}

// AuditLogExport 导出审计日志
func (s *Services) AuditLogExport(ctx context.Context, query *models.AuditLogQuery) ([]*models.AuditLog, error) {
	return s.dao.AuditLogExport(ctx, query)
}

// AuditLogRecord 记录单条审计日志（供中间件或业务层直接调用）
func (s *Services) AuditLogRecord(ctx context.Context, log *models.AuditLog) error {
	return s.dao.AuditLogCreate(ctx, log)
}

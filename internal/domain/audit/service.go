package audit

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/settings"
)

// AuditService 审计日志领域服务
type AuditService struct {
	repo AuditLogRepository
	db   *gorm.DB // 仅用于跨域 settings 查询（GetRetention / UpdateRetention），待 settings 域改造后移除
}

// NewAuditService 创建审计服务
func NewAuditService(repo AuditLogRepository, db *gorm.DB) *AuditService {
	return &AuditService{repo: repo, db: db}
}

// ========== 业务方法 ==========

// List 查询审计日志列表
func (s *AuditService) List(ctx context.Context, query *AuditLogQuery) (*AuditLogListResponse, error) {
	return s.repo.Query(ctx, query)
}

// GetByID 获取审计日志详情
func (s *AuditService) GetByID(ctx context.Context, id int64) (*AuditLog, error) {
	return s.repo.FindByID(ctx, id)
}

// GetStatistics 获取审计统计数据
func (s *AuditService) GetStatistics(ctx context.Context) (*AuditStatistics, error) {
	return s.repo.QueryStatistics(ctx)
}

// GetRetention 获取当前保留策略（跨域查询 settings）
func (s *AuditService) GetRetention(ctx context.Context) (*AuditRetentionPolicy, error) {
	var ps settings.PlatformSettings
	err := s.db.WithContext(ctx).Where("category = ? AND `key` = ?", "security", "audit_retention").First(&ps).Error
	if err != nil {
		return &AuditRetentionPolicy{RetentionDays: 30, IsPermanent: false}, nil
	}

	days := 30
	fmt.Sscanf(ps.Value, "%d", &days)
	return &AuditRetentionPolicy{
		RetentionDays: days,
		IsPermanent:   days == 0,
	}, nil
}

// UpdateRetention 更新保留策略（跨域写入 settings）
func (s *AuditService) UpdateRetention(ctx context.Context, policy *AuditRetentionUpdateReq) error {
	days := policy.RetentionDays
	if policy.IsPermanent {
		days = 0
	}
	value := fmt.Sprintf("%d", days)
	now := uint32(time.Now().Unix())

	var existing settings.PlatformSettings
	err := s.db.WithContext(ctx).Where("category = ? AND `key` = ?", "security", "audit_retention").First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.WithContext(ctx).Create(&settings.PlatformSettings{
			Category: "security", Key: "audit_retention", Value: value,
			ValueType: "int", Label: "审计日志保留", Desc: "审计日志保留天数，0表示永久保留",
			ModifiedAt: now,
		}).Error
	}
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"value": value, "modified_at": now,
	}).Error
}

// Export 导出审计日志（复用 Query 不分页）
func (s *AuditService) Export(ctx context.Context, query *AuditLogQuery) ([]*AuditLog, error) {
	query.Page = 1
	query.PageSize = 10000
	query.SortField = "created_at"
	query.SortOrder = "desc"
	resp, err := s.repo.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return resp.List, nil
}

// Cleanup 清理过期审计日志
func (s *AuditService) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	return s.repo.DeleteBefore(ctx, cutoff)
}

// CleanupByPolicy 根据当前保留策略执行清理
func (s *AuditService) CleanupByPolicy(ctx context.Context) (int64, error) {
	policy, err := s.GetRetention(ctx)
	if err != nil {
		return 0, err
	}
	if policy.IsPermanent || policy.RetentionDays == 0 {
		return 0, nil
	}
	return s.Cleanup(ctx, policy.RetentionDays)
}

// Record 记录单条审计日志
func (s *AuditService) Record(ctx context.Context, log *AuditLog) error {
	return s.repo.Save(ctx, log)
}

// BatchRecord 批量记录审计日志
func (s *AuditService) BatchRecord(ctx context.Context, logs []*AuditLog) error {
	return s.repo.BatchSave(ctx, logs)
}

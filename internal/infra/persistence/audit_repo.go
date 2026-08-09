package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/audit"
)

// auditLogRepo 审计日志仓储 GORM 实现
type auditLogRepo struct {
	db *gorm.DB
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *gorm.DB) audit.AuditLogRepository {
	return &auditLogRepo{db: db}
}

func (r *auditLogRepo) Save(ctx context.Context, log *audit.AuditLog) error {
	if log.CreatedAt == 0 {
		log.CreatedAt = time.Now().Unix()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepo) BatchSave(ctx context.Context, logs []*audit.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, l := range logs {
		if l.CreatedAt == 0 {
			l.CreatedAt = now
		}
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

func (r *auditLogRepo) FindByID(ctx context.Context, id int64) (*audit.AuditLog, error) {
	var log audit.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	return &log, err
}

func (r *auditLogRepo) DeleteBefore(ctx context.Context, beforeTimestamp int64) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", beforeTimestamp).Delete(&audit.AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *auditLogRepo) Query(ctx context.Context, query *audit.AuditLogQuery) (*audit.AuditLogListResponse, error) {
	db := r.db.WithContext(ctx).Model(&audit.AuditLog{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.TargetType != "" {
		db = db.Where("target_type = ?", query.TargetType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.ClusterID > 0 {
		db = db.Where("cluster_id = ?", query.ClusterID)
	}
	if query.StartTime > 0 {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime > 0 {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("(target_name LIKE ? OR action_display LIKE ? OR request_uri LIKE ? OR error_message LIKE ?)",
			keyword, keyword, keyword, keyword)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	sortField := "created_at"
	sortOrder := "DESC"
	if query.SortField != "" {
		allowedFields := map[string]bool{
			"created_at": true, "duration_ms": true, "user_id": true,
			"action": true, "module": true, "status": true,
		}
		if allowedFields[query.SortField] {
			sortField = query.SortField
		}
	}
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	db = db.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	page := query.Page
	pageSize := query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var list []*audit.AuditLog
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}

	return &audit.AuditLogListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *auditLogRepo) QueryStatistics(ctx context.Context) (*audit.AuditStatistics, error) {
	stats := &audit.AuditStatistics{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	weekStart := todayStart - 7*86400

	db := r.db.WithContext(ctx).Model(&audit.AuditLog{})

	db.Where("created_at >= ?", todayStart).Count(&stats.TotalToday)
	db.Where("created_at >= ?", weekStart).Count(&stats.TotalWeek)
	db.Count(&stats.TotalAll)

	if stats.TotalAll > 0 {
		var successCount int64
		db.Where("status = ?", "success").Count(&successCount)
		stats.SuccessRate = float64(successCount) / float64(stats.TotalAll) * 100
	}

	db.Select("username, COUNT(*) as count").
		Where("created_at >= ?", weekStart).
		Group("username").Order("count DESC").Limit(5).
		Scan(&stats.TopUsers)

	db.Select("module, COUNT(*) as count").
		Where("created_at >= ?", weekStart).
		Group("module").Order("count DESC").Limit(5).
		Scan(&stats.TopModules)

	db.Select("action, COUNT(*) as count").
		Where("created_at >= ?", todayStart).
		Group("action").Order("count DESC").
		Scan(&stats.ActionSummary)

	var hourlyCounts []audit.HourlyCount
	db.Select("HOUR(FROM_UNIXTIME(created_at)) as hour, COUNT(*) as count").
		Where("created_at >= ?", todayStart).
		Group("hour").Order("hour ASC").
		Scan(&hourlyCounts)
	stats.HourlyCounts = hourlyCounts

	return stats, nil
}

// Export 导出审计日志（非 Repository 接口方法，直接在 repo 层实现）
func (r *auditLogRepo) Export(ctx context.Context, query *audit.AuditLogQuery) ([]*audit.AuditLog, error) {
	db := r.db.WithContext(ctx).Model(&audit.AuditLog{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Username != "" {
		db = db.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime > 0 {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime > 0 {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var list []*audit.AuditLog
	err := db.Order("created_at DESC").Limit(10000).Find(&list).Error
	return list, err
}

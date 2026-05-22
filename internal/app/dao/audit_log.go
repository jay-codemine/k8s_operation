package dao

import (
	"context"
	"fmt"
	"k8soperation/internal/app/models"
	"time"
)

// ========== 审计日志 DAO ==========

// AuditLogCreate 创建单条审计日志
func (d *Dao) AuditLogCreate(ctx context.Context, log *models.AuditLog) error {
	if log.CreatedAt == 0 {
		log.CreatedAt = time.Now().Unix()
	}
	return d.db.WithContext(ctx).Create(log).Error
}

// AuditLogBatchCreate 批量创建审计日志
func (d *Dao) AuditLogBatchCreate(ctx context.Context, logs []*models.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, l := range logs {
		if l.CreatedAt == 0 {
			l.CreatedAt = now
		}
	}
	return d.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// AuditLogList 查询审计日志列表（支持复杂筛选 + 分页）
func (d *Dao) AuditLogList(ctx context.Context, query *models.AuditLogQuery) (*models.AuditLogListResponse, error) {
	db := d.db.WithContext(ctx).Model(&models.AuditLog{})

	// 筛选条件
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

	// 统计总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	sortField := "created_at"
	sortOrder := "DESC"
	if query.SortField != "" {
		// 安全校验排序字段白名单
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

	// 分页
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
	offset := (page - 1) * pageSize

	// 查询数据
	var list []*models.AuditLog
	if err := db.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}

	return &models.AuditLogListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// AuditLogGetByID 根据ID获取审计日志详情
func (d *Dao) AuditLogGetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	var log models.AuditLog
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	return &log, err
}

// AuditLogGetStatistics 获取审计统计数据
func (d *Dao) AuditLogGetStatistics(ctx context.Context) (*models.AuditStatistics, error) {
	stats := &models.AuditStatistics{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	weekStart := todayStart - 7*86400

	// 今日总量
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at >= ?", todayStart).Count(&stats.TotalToday)

	// 本周总量
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at >= ?", weekStart).Count(&stats.TotalWeek)

	// 总量
	d.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&stats.TotalAll)

	// 成功率
	if stats.TotalAll > 0 {
		var successCount int64
		d.db.WithContext(ctx).Model(&models.AuditLog{}).
			Where("status = ?", "success").Count(&successCount)
		stats.SuccessRate = float64(successCount) / float64(stats.TotalAll) * 100
	}

	// Top 5 用户
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("username, COUNT(*) as count").
		Where("created_at >= ?", weekStart).
		Group("username").
		Order("count DESC").
		Limit(5).
		Scan(&stats.TopUsers)

	// Top 5 模块
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("module, COUNT(*) as count").
		Where("created_at >= ?", weekStart).
		Group("module").
		Order("count DESC").
		Limit(5).
		Scan(&stats.TopModules)

	// 操作类型汇总
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ?", todayStart).
		Group("action").
		Order("count DESC").
		Scan(&stats.ActionSummary)

	// 24小时每小时操作量
	var hourlyCounts []models.HourlyCount
	d.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("HOUR(FROM_UNIXTIME(created_at)) as hour, COUNT(*) as count").
		Where("created_at >= ?", todayStart).
		Group("hour").
		Order("hour ASC").
		Scan(&hourlyCounts)
	stats.HourlyCounts = hourlyCounts

	return stats, nil
}

// AuditLogCleanup 清理过期审计日志
func (d *Dao) AuditLogCleanup(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		// 永久保留，不清理
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	result := d.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&models.AuditLog{})
	return result.RowsAffected, result.Error
}

// AuditLogGetRetentionPolicy 获取当前保留策略
func (d *Dao) AuditLogGetRetentionPolicy(ctx context.Context) (*models.AuditRetentionPolicy, error) {
	var settings models.PlatformSettings
	err := d.db.WithContext(ctx).Where("category = ? AND `key` = ?", "security", "audit_retention").First(&settings).Error
	if err != nil {
		// 默认 30 天
		return &models.AuditRetentionPolicy{RetentionDays: 30, IsPermanent: false}, nil
	}

	days := 30
	fmt.Sscanf(settings.Value, "%d", &days)
	return &models.AuditRetentionPolicy{
		RetentionDays: days,
		IsPermanent:   days == 0,
	}, nil
}

// AuditLogUpdateRetentionPolicy 更新保留策略
func (d *Dao) AuditLogUpdateRetentionPolicy(ctx context.Context, policy *models.AuditRetentionUpdateReq) error {
	days := policy.RetentionDays
	if policy.IsPermanent {
		days = 0
	}
	return d.PlatformSettingsUpsert(ctx, "security", "audit_retention",
		fmt.Sprintf("%d", days), "int", "审计日志保留", "审计日志保留天数，0表示永久保留")
}

// AuditLogExport 导出审计日志（获取全部匹配记录，最多10000条）
func (d *Dao) AuditLogExport(ctx context.Context, query *models.AuditLogQuery) ([]*models.AuditLog, error) {
	db := d.db.WithContext(ctx).Model(&models.AuditLog{})

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

	var list []*models.AuditLog
	err := db.Order("created_at DESC").Limit(10000).Find(&list).Error
	return list, err
}

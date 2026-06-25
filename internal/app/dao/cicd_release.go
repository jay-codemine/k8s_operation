package dao

import (
	"context"
	"k8soperation/internal/app/models"
	"time"
)

// CicdReleaseCreate 创建发布单（cicd_release）
func (d *Dao) CicdReleaseCreate(ctx context.Context, rel *models.CicdRelease) error {
	return d.db.WithContext(ctx).
		Create(rel).
		Error
}

// CicdReleaseUpdateStatusCAS 更新发布单状态（cicd_release）
func (d *Dao) CicdReleaseUpdateStatusCAS(
	ctx context.Context,
	releaseID int64,
	from []string,
	to string,
	message string,
) (bool, error) {

	res := d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Where("status IN ?", from).
		Updates(map[string]any{
			"status":      to,
			"message":     message,
			"modified_at": time.Now().Unix(),
		})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// CicdReleaseGetByRequestID 根据 RequestID 获取发布单（幂等校验用）
func (d *Dao) CicdReleaseGetByRequestID(ctx context.Context, requestID string) (*models.CicdRelease, error) {
	var rel models.CicdRelease
	err := d.db.WithContext(ctx).
		Where("request_id = ? AND is_del = 0", requestID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// CicdReleaseGetByID 根据 ID 获取发布单
func (d *Dao) CicdReleaseGetByID(ctx context.Context, releaseID int64) (*models.CicdRelease, error) {
	var rel models.CicdRelease
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", releaseID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// CicdReleaseList 发布单列表查询
func (d *Dao) CicdReleaseList(ctx context.Context, keyword, appName, status string, page, pageSize int) ([]*models.CicdRelease, int64, error) {
	var list []*models.CicdRelease
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdRelease{}).Where("is_del = 0")

	// keyword 模糊搜索：应用名、工作负载名、镜像等
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where(
			"app_name LIKE ? OR workload_name LIKE ? OR image_repo LIKE ? OR image_tag LIKE ?",
			likePattern, likePattern, likePattern, likePattern,
		)
	}
	// 精确匹配 app_name（兼容旧参数）
	if appName != "" {
		query = query.Where("app_name LIKE ?", "%"+appName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 先查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 再查询分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CicdReleaseStats 按状态统计发布单数量
func (d *Dao) CicdReleaseStats(ctx context.Context) (map[string]int64, error) {
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Select("status, COUNT(*) AS cnt").
		Where("is_del = 0").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	stats := map[string]int64{
		"Pending": 0, "Queued": 0, "Running": 0,
		"Succeeded": 0, "Failed": 0, "Canceled": 0, "Rollback": 0,
		"total": 0,
	}
	for _, r := range rows {
		stats[r.Status] = r.Cnt
		stats["total"] += r.Cnt
	}
	return stats, nil
}

// CicdReleaseCancel 取消发布单
func (d *Dao) CicdReleaseCancel(ctx context.Context, releaseID int64) (bool, error) {
	return d.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{models.CicdReleaseStatusPending, models.CicdReleaseStatusAwaitingApproval, models.CicdReleaseStatusQueued, models.CicdReleaseStatusRunning},
		models.CicdReleaseStatusCanceled,
		"user canceled",
	)
}

// CicdReleaseGetByBuildID 根据 Jenkins 构建 ID 获取发布单
func (d *Dao) CicdReleaseGetByBuildID(ctx context.Context, buildID int64) (*models.CicdRelease, error) {
	var rel models.CicdRelease
	err := d.db.WithContext(ctx).
		Where("build_id = ? AND is_del = 0", buildID).
		First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// CicdReleaseUpdate 编辑发布单（仅 Pending/AwaitingApproval/Failed/Canceled 状态可编辑）
func (d *Dao) CicdReleaseUpdate(ctx context.Context, releaseID int64, updates map[string]any) error {
	updates["modified_at"] = time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("id = ? AND is_del = 0 AND status IN ?", releaseID, []string{models.CicdReleaseStatusPending, models.CicdReleaseStatusAwaitingApproval, models.CicdReleaseStatusFailed, models.CicdReleaseStatusCanceled}).
		Updates(updates).Error
}

// CicdReleaseDelete 软删除发布单
func (d *Dao) CicdReleaseDelete(ctx context.Context, releaseID int64) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Updates(map[string]any{
			"is_del":     1,
			"deleted_at":  now,
			"modified_at": now,
		}).Error
}

// CicdReleaseUpdateImage 更新发布单的镜像信息
func (d *Dao) CicdReleaseUpdateImage(ctx context.Context, releaseID int64, imageRepo, imageTag, imageDigest string) error {
	updates := map[string]any{
		"image_repo":  imageRepo,
		"image_tag":   imageTag,
		"modified_at": time.Now().Unix(),
	}
	if imageDigest != "" {
		updates["image_digest"] = imageDigest
	}
	return d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("id = ? AND is_del = 0", releaseID).
		Updates(updates).Error
}

// ==================== 发布历史增强 ====================

// CicdReleaseHistory 应用发布历史查询（增强版）支持时间范围、环境筛选
func (d *Dao) CicdReleaseHistory(ctx context.Context, appName, namespace, status string, startTime, endTime int64, page, pageSize int) ([]*models.CicdRelease, int64, error) {
	var list []*models.CicdRelease
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdRelease{}).Where("is_del = 0")

	if appName != "" {
		query = query.Where("app_name LIKE ?", "%"+appName+"%")
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime > 0 {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ReleaseStatsEnhanced 发布统计增强（包含平均发布时间、今日发布数等）
type ReleaseStatsEnhanced struct {
	Total       int64              `json:"total"`
	ByStatus    map[string]int64   `json:"by_status"`
	TodayCount  int64              `json:"today_count"`
	WeekCount   int64              `json:"week_count"`
	SuccessRate float64            `json:"success_rate"`
}

// CicdReleaseStatsEnhanced 获取增强版发布统计
func (d *Dao) CicdReleaseStatsEnhanced(ctx context.Context) (*ReleaseStatsEnhanced, error) {
	stats := &ReleaseStatsEnhanced{
		ByStatus: make(map[string]int64),
	}

	// 按状态统计
	type statusCount struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []statusCount
	err := d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Select("status, COUNT(*) AS cnt").
		Where("is_del = 0").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	var successCount int64
	var terminalCount int64
	for _, r := range rows {
		stats.Total += r.Cnt
		stats.ByStatus[r.Status] = r.Cnt
		if r.Status == models.CicdReleaseStatusSucceeded {
			successCount = r.Cnt
		}
		if r.Status == models.CicdReleaseStatusSucceeded || r.Status == models.CicdReleaseStatusFailed {
			terminalCount += r.Cnt
		}
	}
	if terminalCount > 0 {
		stats.SuccessRate = float64(successCount) / float64(terminalCount) * 100
	}

	// 今日发布数
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("is_del = 0 AND created_at >= ?", todayStart.Unix()).
		Count(&stats.TodayCount)

	// 本周发布数
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	d.db.WithContext(ctx).
		Model(&models.CicdRelease{}).
		Where("is_del = 0 AND created_at >= ?", weekStart.Unix()).
		Count(&stats.WeekCount)

	return stats, nil
}

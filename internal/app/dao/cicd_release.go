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
func (d *Dao) CicdReleaseList(ctx context.Context, appName, status string, page, pageSize int) ([]*models.CicdRelease, int64, error) {
	var list []*models.CicdRelease
	var total int64

	query := d.db.WithContext(ctx).Model(&models.CicdRelease{}).Where("is_del = 0")

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

// CicdReleaseCancel 取消发布单
func (d *Dao) CicdReleaseCancel(ctx context.Context, releaseID int64) (bool, error) {
	return d.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{models.CicdReleaseStatusPending, models.CicdReleaseStatusQueued, models.CicdReleaseStatusRunning},
		models.CicdReleaseStatusCanceled,
		"user canceled",
	)
}

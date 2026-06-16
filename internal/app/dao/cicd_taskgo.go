package dao

import (
	"context"
	"k8soperation/internal/app/models"
	"time"
)

// CicdTasksCreate 批量创建发布任务（cicd_release_task）
// 注意：逐条创建以确保每条记录的 ID 正确回填
func (d *Dao) CicdTasksCreate(ctx context.Context, tasks []*models.CicdReleaseTask) error {
	if len(tasks) == 0 {
		return nil
	}
	// 逐条创建，确保 GORM 正确回填每条记录的自增 ID
	for _, task := range tasks {
		if err := d.db.WithContext(ctx).Create(task).Error; err != nil {
			return err
		}
	}
	return nil
}

// CicdTaskUpdatePrevImage 更新任务的原镜像
func (d *Dao) CicdTaskUpdatePrevImage(ctx context.Context, taskID int64, prev string) error {
	now := uint64(time.Now().Unix())
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"prev_image":  prev,
			"modified_at": now,
		}).Error
}

// CicdTaskGetByID 根据 ID 获取任务
func (d *Dao) CicdTaskGetByID(ctx context.Context, taskID int64) (*models.CicdReleaseTask, error) {
	var task models.CicdReleaseTask
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_del = 0", taskID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// CicdTaskUpdateStatus 更新任务状态
func (d *Dao) CicdTaskUpdateStatus(ctx context.Context, taskID int64, status, message string) error {
	now := uint64(time.Now().Unix())
	updates := map[string]any{
		"status":      status,
		"message":     message,
		"modified_at": now,
	}
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(updates).Error
}

// CicdTaskUpdateStatusCAS CAS 更新任务状态（乐观锁）
func (d *Dao) CicdTaskUpdateStatusCAS(ctx context.Context, taskID int64, from []string, to, message string) (bool, error) {
	now := uint64(time.Now().Unix())
	res := d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Where("status IN ?", from).
		Updates(map[string]any{
			"status":      to,
			"message":     message,
			"modified_at": now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// CicdTaskMarkStarted 标记任务开始执行
func (d *Dao) CicdTaskMarkStarted(ctx context.Context, taskID int64) error {
	now := uint64(time.Now().Unix())
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"status":      models.CicdTaskStatusRunning,
			"started_at":  now,
			"modified_at": now,
		}).Error
}

// CicdTaskMarkFinished 标记任务完成
func (d *Dao) CicdTaskMarkFinished(ctx context.Context, taskID int64, status, message string) error {
	now := uint64(time.Now().Unix())
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"status":      status,
			"message":     message,
			"finished_at": now,
			"modified_at": now,
		}).Error
}

// CicdTasksByReleaseID 获取发布单下的所有任务
func (d *Dao) CicdTasksByReleaseID(ctx context.Context, releaseID int64) ([]*models.CicdReleaseTask, error) {
	var tasks []*models.CicdReleaseTask
	err := d.db.WithContext(ctx).
		Where("release_id = ? AND is_del = 0", releaseID).
		Order("id ASC").
		Find(&tasks).Error
	return tasks, err
}

// CicdTaskUpdateTargetImage 更新任务的目标镜像
func (d *Dao) CicdTaskUpdateTargetImage(ctx context.Context, taskID int64, targetImage string) error {
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("id = ? AND is_del = 0", taskID).
		Updates(map[string]any{
			"target_image": targetImage,
			"modified_at":  time.Now().Unix(),
		}).Error
}

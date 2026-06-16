package dao

import (
	"context"
	"k8soperation/internal/app/models"
	"time"
)

// 把某个 release 下仍未执行的任务（Pending/Queued）批量置为 Failed，避免悬挂
func (d *Dao) CicdTasksFailByRelease(ctx context.Context, releaseID int64, msg string) error {
	return d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Where("release_id = ? AND is_del = 0", releaseID).
		Where("status IN ?", []string{"Pending", "Queued"}).
		Updates(map[string]any{
			"status":      "Failed",
			"message":     msg,
			"modified_at": time.Now().Unix(),
		}).Error
}

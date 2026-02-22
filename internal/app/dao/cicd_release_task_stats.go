package dao

import (
	"context"

	"k8soperation/internal/app/models"
)

type CicdTaskStats struct {
	Pending   int64
	Queued    int64
	Running   int64
	Succeeded int64
	Failed    int64
	Canceled  int64
}

func (d *Dao) CicdTaskStatsByRelease(ctx context.Context, releaseID int64) (*CicdTaskStats, error) {
	var rows []struct {
		Status string
		Cnt    int64
	}

	if err := d.db.WithContext(ctx).
		Model(&models.CicdReleaseTask{}).
		Select("status AS status, COUNT(1) AS cnt").
		Where("release_id = ? AND is_del = 0", releaseID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	st := &CicdTaskStats{}
	for _, r := range rows {
		switch r.Status {
		case "Pending":
			st.Pending = r.Cnt
		case "Queued":
			st.Queued = r.Cnt
		case "Running":
			st.Running = r.Cnt
		case "Succeeded":
			st.Succeeded = r.Cnt
		case "Failed":
			st.Failed = r.Cnt
		case "Canceled":
			st.Canceled = r.Cnt
		}
	}
	return st, nil
}

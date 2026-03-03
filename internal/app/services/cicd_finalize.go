package services

import (
	"context"
)

func (s *Services) tryFinalizeRelease(ctx context.Context, releaseID int64) {
	st, err := s.dao.CicdTaskStatsByRelease(ctx, releaseID)
	if err != nil {
		return
	}

	// 1) 任意失败 => release Failed（CAS 防并发覆盖）
	if st.Failed > 0 {
		_, _ = s.dao.CicdReleaseUpdateStatusCAS(ctx, releaseID,
			[]string{"Pending", "Queued", "Running"},
			"Failed",
			"one or more tasks failed",
		)
		return
	}

	// 2) 还有没结束的 => 不动
	if st.Pending > 0 || st.Queued > 0 || st.Running > 0 {
		return
	}

	// 3) 全部结束且无失败 => Succeeded
	_, _ = s.dao.CicdReleaseUpdateStatusCAS(ctx, releaseID,
		[]string{"Pending", "Queued", "Running"},
		"Succeeded",
		"all tasks succeeded",
	)
}

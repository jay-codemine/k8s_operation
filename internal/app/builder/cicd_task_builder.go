package builder

import "k8soperation/internal/app/models"

func BuildCicdReleaseTasks(
	releaseID int64,
	clusterIDs []int64,
	target string,
	now uint64,
) []*models.CicdReleaseTask {

	tasks := make([]*models.CicdReleaseTask, 0, len(clusterIDs))

	for _, cid := range clusterIDs {
		tasks = append(tasks, &models.CicdReleaseTask{
			ReleaseID: releaseID,
			ClusterID: cid,

			Status:  models.CicdReleaseStatusPending,
			Message: "",

			PrevImage:   "",
			TargetImage: target,

			StartedAt:  0,
			FinishedAt: 0,

			CreatedAt:  now,
			ModifiedAt: now,
		})
	}

	return tasks
}

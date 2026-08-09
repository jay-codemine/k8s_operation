package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type CicdReleaseTask = dm.CicdReleaseTask

// ===== 常量别名 =====
const (
	CicdTaskStatusPending   = dm.CicdTaskStatusPending
	CicdTaskStatusRunning   = dm.CicdTaskStatusRunning
	CicdTaskStatusSucceeded = dm.CicdTaskStatusSucceeded
	CicdTaskStatusFailed    = dm.CicdTaskStatusFailed
	CicdTaskStatusCanceled  = dm.CicdTaskStatusCanceled
)

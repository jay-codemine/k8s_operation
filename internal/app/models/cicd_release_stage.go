package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	CicdReleaseStageStatus = dm.CicdReleaseStageStatus
	CicdReleaseStage       = dm.CicdReleaseStage
)

// ===== 常量别名 =====
const (
	CicdReleaseStageStatusPending = dm.CicdReleaseStageStatusPending
	CicdReleaseStageStatusRunning = dm.CicdReleaseStageStatusRunning
	CicdReleaseStageStatusSuccess = dm.CicdReleaseStageStatusSuccess
	CicdReleaseStageStatusFailed  = dm.CicdReleaseStageStatusFailed
	CicdReleaseStageStatusSkipped = dm.CicdReleaseStageStatusSkipped
	CicdReleaseStageStatusWaiting = dm.CicdReleaseStageStatusWaiting
	CicdReleaseStageStatusAborted = dm.CicdReleaseStageStatusAborted
)

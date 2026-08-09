package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	CicdRelease               = dm.CicdRelease
	CicdReleaseWithDeployMode = dm.CicdReleaseWithDeployMode
)

// ===== 常量别名 =====
const (
	CicdReleaseStatusPending          = dm.CicdReleaseStatusPending
	CicdReleaseStatusAwaitingApproval = dm.CicdReleaseStatusAwaitingApproval
	CicdReleaseStatusQueued           = dm.CicdReleaseStatusQueued
	CicdReleaseStatusRunning          = dm.CicdReleaseStatusRunning
	CicdReleaseStatusSucceeded        = dm.CicdReleaseStatusSucceeded
	CicdReleaseStatusFailed           = dm.CicdReleaseStatusFailed
	CicdReleaseStatusCanceled         = dm.CicdReleaseStatusCanceled
	CicdReleaseStatusRollback         = dm.CicdReleaseStatusRollback
)

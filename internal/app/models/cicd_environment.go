package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	ApprovalLevel       = dm.ApprovalLevel
	ApprovalLevels      = dm.ApprovalLevels
	CicdEnvironment     = dm.CicdEnvironment
	CicdApproval        = dm.CicdApproval
	ApprovalListItem    = dm.ApprovalListItem
	EnvironmentListItem = dm.EnvironmentListItem
)

// ===== 常量别名 =====
const (
	ApprovalStatusPending  = dm.ApprovalStatusPending
	ApprovalStatusApproved = dm.ApprovalStatusApproved
	ApprovalStatusRejected = dm.ApprovalStatusRejected
	ApprovalStatusExpired  = dm.ApprovalStatusExpired
)

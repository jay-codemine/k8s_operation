package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	CicdResourceTemplate      = dm.CicdResourceTemplate
	CicdEnvResourceRule       = dm.CicdEnvResourceRule
	RiskLevel                 = dm.RiskLevel
	CicdDeployApproval        = dm.CicdDeployApproval
	ChangeType                = dm.ChangeType
	CicdResourceChangeLog     = dm.CicdResourceChangeLog
	ResourceConfig            = dm.ResourceConfig
	ResourceRequests          = dm.ResourceRequests
	ResourceValues            = dm.ResourceValues
	HPAConfig                 = dm.HPAConfig
	ResourceValidationResult  = dm.ResourceValidationResult
)

// ===== 常量别名 =====
const (
	ApprovalStatusCancelled = dm.ApprovalStatusCancelled

	RiskLevelLow    = dm.RiskLevelLow
	RiskLevelMedium = dm.RiskLevelMedium
	RiskLevelHigh   = dm.RiskLevelHigh

	ChangeTypeCreate   = dm.ChangeTypeCreate
	ChangeTypeUpdate   = dm.ChangeTypeUpdate
	ChangeTypeScale    = dm.ChangeTypeScale
	ChangeTypeRollback = dm.ChangeTypeRollback
)

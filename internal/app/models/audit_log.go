package models

import dm "k8soperation/internal/domain/audit"

// ===== 类型别名 =====

type (
	AuditLog               = dm.AuditLog
	AuditRetentionPolicy   = dm.AuditRetentionPolicy
	AuditLogQuery          = dm.AuditLogQuery
	AuditLogListResponse   = dm.AuditLogListResponse
	AuditStatistics        = dm.AuditStatistics
	TopUserStat            = dm.TopUserStat
	TopModuleStat          = dm.TopModuleStat
	ActionSummary          = dm.ActionSummary
	HourlyCount            = dm.HourlyCount
	AuditRetentionUpdateReq = dm.AuditRetentionUpdateReq
)

// ===== 审计模块/操作常量 =====

const (
	AuditModuleAuth     = dm.AuditModuleAuth
	AuditModuleCluster  = dm.AuditModuleCluster
	AuditModuleWorkload = dm.AuditModuleWorkload
	AuditModuleNetwork  = dm.AuditModuleNetwork
	AuditModuleConfig   = dm.AuditModuleConfig
	AuditModuleStorage  = dm.AuditModuleStorage
	AuditModuleCICD     = dm.AuditModuleCICD
	AuditModuleRBAC     = dm.AuditModuleRBAC
	AuditModulePlatform = dm.AuditModulePlatform
	AuditModuleAI       = dm.AuditModuleAI
	AuditModuleMonitor  = dm.AuditModuleMonitor
	AuditModuleImage    = dm.AuditModuleImage

	AuditActionCreate  = dm.AuditActionCreate
	AuditActionUpdate  = dm.AuditActionUpdate
	AuditActionDelete  = dm.AuditActionDelete
	AuditActionLogin   = dm.AuditActionLogin
	AuditActionLogout  = dm.AuditActionLogout
	AuditActionExec    = dm.AuditActionExec
	AuditActionApprove = dm.AuditActionApprove
	AuditActionReject  = dm.AuditActionReject
	AuditActionDeploy  = dm.AuditActionDeploy
	AuditActionScale   = dm.AuditActionScale
	AuditActionView    = dm.AuditActionView

	AuditStatusSuccess = dm.AuditStatusSuccess
	AuditStatusFailed  = dm.AuditStatusFailed
)


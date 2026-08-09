package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type GitOpsConfig = dm.GitOpsConfig

// ===== 常量别名 =====
const (
	DeployModeJenkins = dm.DeployModeJenkins
	DeployModeGitOps  = dm.DeployModeGitOps

	SyncStatusSynced    = dm.SyncStatusSynced
	SyncStatusOutOfSync = dm.SyncStatusOutOfSync
	SyncStatusUnknown   = dm.SyncStatusUnknown
	SyncStatusSyncing   = dm.SyncStatusSyncing

	WorkflowStatusPending   = dm.WorkflowStatusPending
	WorkflowStatusRunning   = dm.WorkflowStatusRunning
	WorkflowStatusSucceeded = dm.WorkflowStatusSucceeded
	WorkflowStatusFailed    = dm.WorkflowStatusFailed
	WorkflowStatusError     = dm.WorkflowStatusError
)

// ===== 变量别名 =====
var (
	ValidDeployModes               = dm.ValidDeployModes
	DefaultArgoWorkflowTemplateMap = dm.DefaultArgoWorkflowTemplateMap
	GitOpsConfigFromJSONMap        = dm.GitOpsConfigFromJSONMap
)

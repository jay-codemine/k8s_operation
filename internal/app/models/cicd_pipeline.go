package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	EnvVar              = dm.EnvVar
	EnvVars             = dm.EnvVars
	DeployConfig        = dm.DeployConfig
	JSONMap             = dm.JSONMap
	JSONArray           = dm.JSONArray
	CicdPipeline        = dm.CicdPipeline
	CicdPipelineRun     = dm.CicdPipelineRun
	CicdPipelineStage   = dm.CicdPipelineStage
	StageDisplayInfo    = dm.StageDisplayInfo
	StageApprovalInfo   = dm.StageApprovalInfo
	StageDeployInfo     = dm.StageDeployInfo
	StageSonarInfo      = dm.StageSonarInfo
	PipelineListItem    = dm.PipelineListItem
)

// ===== 常量别名 =====
const (
	PipelineStatusIdle     = dm.PipelineStatusIdle
	PipelineStatusRunning  = dm.PipelineStatusRunning
	PipelineStatusDisabled = dm.PipelineStatusDisabled

	PipelineRunStatusPending = dm.PipelineRunStatusPending
	PipelineRunStatusRunning = dm.PipelineRunStatusRunning
	PipelineRunStatusSuccess = dm.PipelineRunStatusSuccess
	PipelineRunStatusFailed  = dm.PipelineRunStatusFailed
	PipelineRunStatusAborted = dm.PipelineRunStatusAborted

	TriggerTypeManual    = dm.TriggerTypeManual
	TriggerTypeWebhook   = dm.TriggerTypeWebhook
	TriggerTypeScheduled = dm.TriggerTypeScheduled

	LanguageTypeGo       = dm.LanguageTypeGo
	LanguageTypeJava     = dm.LanguageTypeJava
	LanguageTypeFrontend = dm.LanguageTypeFrontend
	LanguageTypePython   = dm.LanguageTypePython

	StageTypeClean           = dm.StageTypeClean
	StageTypeSCM             = dm.StageTypeSCM
	StageTypeCheckout        = dm.StageTypeCheckout
	StageTypeDependencies    = dm.StageTypeDependencies
	StageTypeCompile         = dm.StageTypeCompile
	StageTypeTest            = dm.StageTypeTest
	StageTypeLint            = dm.StageTypeLint
	StageTypeSonar           = dm.StageTypeSonar
	StageTypeQualityGate     = dm.StageTypeQualityGate
	StageTypeBuildBinary     = dm.StageTypeBuildBinary
	StageTypeUploadArtifact  = dm.StageTypeUploadArtifact
	StageTypeBuild           = dm.StageTypeBuild
	StageTypePush            = dm.StageTypePush
	StageTypePrepareAgents   = dm.StageTypePrepareAgents
	StageTypeApproval        = dm.StageTypeApproval
	StageTypeDeploy          = dm.StageTypeDeploy

	QualityGateOK    = dm.QualityGateOK
	QualityGateWarn  = dm.QualityGateWarn
	QualityGateError = dm.QualityGateError
	QualityGateNone  = dm.QualityGateNone

	StageStatusPending = dm.StageStatusPending
	StageStatusRunning = dm.StageStatusRunning
	StageStatusSuccess = dm.StageStatusSuccess
	StageStatusFailed  = dm.StageStatusFailed
	StageStatusSkipped = dm.StageStatusSkipped
	StageStatusWaiting = dm.StageStatusWaiting
	StageStatusAborted = dm.StageStatusAborted

	DeployEnvDev     = dm.DeployEnvDev
	DeployEnvTest    = dm.DeployEnvTest
	DeployEnvStaging = dm.DeployEnvStaging
	DeployEnvProd    = dm.DeployEnvProd
)

// ===== 变量别名 =====
var (
	DefaultJenkinsJobMap = dm.DefaultJenkinsJobMap
	DefaultScriptPathMap = dm.DefaultScriptPathMap
	ValidLanguageTypes   = dm.ValidLanguageTypes
)

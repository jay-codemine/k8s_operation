package cicd

// Pipeline 状态常量
const (
	PipelineStatusIdle     = "idle"
	PipelineStatusRunning  = "running"
	PipelineStatusDisabled = "disabled"
)

// Pipeline 运行状态常量
const (
	PipelineRunStatusPending = "pending"
	PipelineRunStatusRunning = "running"
	PipelineRunStatusSuccess = "success"
	PipelineRunStatusFailed  = "failed"
	PipelineRunStatusAborted = "aborted"
)

// 触发类型常量
const (
	TriggerTypeManual    = "manual"
	TriggerTypeWebhook   = "webhook"
	TriggerTypeScheduled = "scheduled"
)

// 语言类型常量
const (
	LanguageTypeGo       = "go"
	LanguageTypeJava     = "java"
	LanguageTypeFrontend = "frontend"
	LanguageTypePython   = "python"
)

// DefaultJenkinsJobMap 语言类型 -> Jenkins 通用 Builder Job 名称
var DefaultJenkinsJobMap = map[string]string{
	LanguageTypeGo:       "go-pipeline",
	LanguageTypeJava:     "java-spring-pipeline",
	LanguageTypeFrontend: "frontend-pipeline",
	LanguageTypePython:   "python-pipeline",
}

// DefaultScriptPathMap 语言类型 -> Jenkins Pipeline Script Path
var DefaultScriptPathMap = map[string]string{
	LanguageTypeGo:       "configs/jenkins-templates/go-pipeline.groovy",
	LanguageTypeJava:     "configs/jenkins-templates/java-spring-pipeline.groovy",
	LanguageTypeFrontend: "configs/jenkins-templates/frontend-pipeline.groovy",
	LanguageTypePython:   "configs/jenkins-templates/python-pipeline.groovy",
}

// ValidLanguageTypes 合法的语言类型列表
var ValidLanguageTypes = []string{
	LanguageTypeGo, LanguageTypeJava, LanguageTypeFrontend, LanguageTypePython,
}

// 部署环境常量
const (
	DeployEnvDev     = "dev"
	DeployEnvTest    = "test"
	DeployEnvStaging = "staging"
	DeployEnvProd    = "prod"
)

// 阶段类型常量
const (
	StageTypeClean           = "clean"
	StageTypeSCM             = "scm"
	StageTypeCheckout        = "checkout"
	StageTypeDependencies    = "dependencies"
	StageTypeCompile         = "compile"
	StageTypeTest            = "test"
	StageTypeLint            = "lint"
	StageTypeSonar           = "sonar"
	StageTypeQualityGate     = "quality_gate"
	StageTypeBuildBinary     = "build_binary"
	StageTypeUploadArtifact  = "upload_artifact"
	StageTypeBuild           = "build"
	StageTypePush            = "push"
	StageTypePrepareAgents   = "prepare_agents"
	StageTypeApproval        = "approval"
	StageTypeDeploy          = "deploy"
)

// 质量门禁状态
const (
	QualityGateOK    = "OK"
	QualityGateWarn  = "WARN"
	QualityGateError = "ERROR"
	QualityGateNone  = "NONE"
)

// 阶段状态常量
const (
	StageStatusPending = "pending"
	StageStatusRunning = "running"
	StageStatusSuccess = "success"
	StageStatusFailed  = "failed"
	StageStatusSkipped = "skipped"
	StageStatusWaiting = "waiting"
	StageStatusAborted = "aborted"
)

// CicdPipeline 对应表：cicd_pipeline
type CicdPipeline struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;uniqueIndex:idx_pipeline_name_del" json:"name"`
	Description string `gorm:"column:description" json:"description"`

	GitRepo   string `gorm:"column:git_repo" json:"git_repo"`
	GitBranch string `gorm:"column:git_branch" json:"git_branch"`

	DeployMode   string  `gorm:"column:deploy_mode;size:20;default:'jenkins'" json:"deploy_mode"`
	GitOpsConfig JSONMap `gorm:"column:gitops_config;type:json" json:"gitops_config,omitempty"`

	JenkinsURL          string `gorm:"column:jenkins_url" json:"jenkins_url"`
	JenkinsJob          string `gorm:"column:jenkins_job" json:"jenkins_job"`
	JenkinsCredentialID string `gorm:"column:jenkins_credential_id" json:"jenkins_credential_id"`
	LanguageType        string `gorm:"column:language_type;size:20;default:'custom'" json:"language_type"`

	AutoDeploy           bool   `gorm:"column:auto_deploy" json:"auto_deploy"`
	EnvironmentID        int64  `gorm:"column:environment_id;index" json:"environment_id"`
	TargetClusterID      int64  `gorm:"column:target_cluster_id" json:"target_cluster_id"`
	TargetNamespace      string `gorm:"column:target_namespace" json:"target_namespace"`
	TargetWorkloadKind   string `gorm:"column:target_workload_kind" json:"target_workload_kind"`
	TargetWorkloadName   string `gorm:"column:target_workload_name" json:"target_workload_name"`
	TargetContainer      string `gorm:"column:target_container" json:"target_container"`
	DeployEnv            string `gorm:"column:deploy_env" json:"deploy_env"`
	RequireApproval      bool   `gorm:"column:require_approval" json:"require_approval"`

	EnableCanary          bool   `gorm:"column:enable_canary" json:"enable_canary"`
	CanaryReplicas        int32  `gorm:"column:canary_replicas;default:1" json:"canary_replicas"`
	CanaryTrafficRatio    int32  `gorm:"column:canary_traffic_ratio;default:10" json:"canary_traffic_ratio"`
	CanaryDurationSec     int32  `gorm:"column:canary_duration_sec;default:300" json:"canary_duration_sec"`
	CanaryAutoPromote     bool   `gorm:"column:canary_auto_promote" json:"canary_auto_promote"`
	CanaryAnalysisRules   string `gorm:"column:canary_analysis_rules;type:text" json:"canary_analysis_rules"`
	EnableSonar           bool   `gorm:"column:enable_sonar" json:"enable_sonar"`
	EnableArtifactUpload  bool   `gorm:"column:enable_artifact_upload" json:"enable_artifact_upload"`
	EnableBuildCache      bool   `gorm:"column:enable_build_cache" json:"enable_build_cache"`

	EnableDeploySilence  bool   `gorm:"column:enable_deploy_silence" json:"enable_deploy_silence"`
	SilenceBufferMinutes int    `gorm:"column:silence_buffer_minutes;default:10" json:"silence_buffer_minutes"`
	SilenceSeverities    string `gorm:"column:silence_severities;size:100;default:'warning,info'" json:"silence_severities"`

	LastDeployImage   string `gorm:"column:last_deploy_image" json:"last_deploy_image"`
	LastDeployDigest  string `gorm:"column:last_deploy_digest" json:"last_deploy_digest"`
	LastDeployTime    uint64 `gorm:"column:last_deploy_time" json:"last_deploy_time"`
	LastDeployStatus  string `gorm:"column:last_deploy_status" json:"last_deploy_status"`
	LastDeployVersion string `gorm:"column:last_deploy_version" json:"last_deploy_version"`

	Status          string `gorm:"column:status" json:"status"`
	LastRunStatus   string `gorm:"column:last_run_status" json:"last_run_status"`
	LastRunTime     uint64 `gorm:"column:last_run_time" json:"last_run_time"`
	LastBuildNumber int    `gorm:"column:last_build_number" json:"last_build_number"`
	LastBuildURL    string `gorm:"column:last_build_url" json:"last_build_url"`

	EnvVars      EnvVars `gorm:"column:env_vars;type:json" json:"env_vars"`
	DeployConfig JSONMap `gorm:"column:deploy_config;type:json" json:"deploy_config"`
	Stages       JSONMap `gorm:"column:stages;type:json" json:"stages"`

	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at;uniqueIndex:idx_pipeline_name_del" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del" json:"is_del"`
}

func (CicdPipeline) TableName() string { return "cicd_pipeline" }

// CicdPipelineRun 对应表：cicd_pipeline_run
type CicdPipelineRun struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PipelineID  int64  `gorm:"column:pipeline_id" json:"pipeline_id"`
	BuildNumber int    `gorm:"column:build_number" json:"build_number"`
	Status      string `gorm:"column:status" json:"status"`

	TriggerType   string `gorm:"column:trigger_type" json:"trigger_type"`
	TriggerUserID int64  `gorm:"column:trigger_user_id" json:"trigger_user_id"`

	GitCommit        string `gorm:"column:git_commit" json:"git_commit"`
	GitBranch        string `gorm:"column:git_branch" json:"git_branch"`
	GitCommitMessage string `gorm:"column:git_commit_message" json:"git_commit_message"`
	JenkinsBuildURL  string `gorm:"column:jenkins_build_url" json:"jenkins_build_url"`

	WorkflowName string `gorm:"column:workflow_name" json:"workflow_name,omitempty"`
	ArgoAppName  string `gorm:"column:argo_app_name" json:"argo_app_name,omitempty"`
	SyncRevision string `gorm:"column:sync_revision" json:"sync_revision,omitempty"`
	SyncStatus   string `gorm:"column:sync_status" json:"sync_status,omitempty"`

	ImageURL    string `gorm:"column:image_url" json:"image_url,omitempty"`
	ImageDigest string `gorm:"column:image_digest" json:"image_digest,omitempty"`

	CallbackReceived uint8 `gorm:"column:callback_received" json:"callback_received"`

	DurationSec  int     `gorm:"column:duration_sec" json:"duration_sec"`
	ConsoleLog   string  `gorm:"column:console_log" json:"console_log,omitempty"`
	StagesResult JSONMap `gorm:"column:stages_result;type:json" json:"stages_result"`
	ErrorMessage string  `gorm:"column:error_message" json:"error_message,omitempty"`

	StartedAt  uint64 `gorm:"column:started_at" json:"started_at"`
	FinishedAt uint64 `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdPipelineRun) TableName() string { return "cicd_pipeline_run" }

// CicdPipelineStage 流水线阶段执行记录表
type CicdPipelineStage struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RunID      int64  `gorm:"column:run_id;index" json:"run_id"`
	PipelineID int64  `gorm:"column:pipeline_id;index" json:"pipeline_id"`
	StageOrder int    `gorm:"column:stage_order" json:"stage_order"`
	StageType  string `gorm:"column:stage_type" json:"stage_type"`
	StageName  string `gorm:"column:stage_name" json:"stage_name"`
	Status     string `gorm:"column:status" json:"status"`

	StartedAt   uint64 `gorm:"column:started_at" json:"started_at"`
	FinishedAt  uint64 `gorm:"column:finished_at" json:"finished_at"`
	DurationSec int    `gorm:"column:duration_sec" json:"duration_sec"`
	Logs        string `gorm:"column:logs;type:longtext" json:"logs"`

	JenkinsStageID string `gorm:"column:jenkins_stage_id" json:"jenkins_stage_id,omitempty"`

	ApprovalUserID   int64  `gorm:"column:approval_user_id" json:"approval_user_id,omitempty"`
	ApprovalComment  string `gorm:"column:approval_comment" json:"approval_comment,omitempty"`
	ApprovalDecision string `gorm:"column:approval_decision" json:"approval_decision,omitempty"`

	DeployClusterID    int64  `gorm:"column:deploy_cluster_id" json:"deploy_cluster_id,omitempty"`
	DeployNamespace    string `gorm:"column:deploy_namespace" json:"deploy_namespace,omitempty"`
	DeployWorkloadKind string `gorm:"column:deploy_workload_kind" json:"deploy_workload_kind,omitempty"`
	DeployWorkloadName string `gorm:"column:deploy_workload_name" json:"deploy_workload_name,omitempty"`
	DeployContainer    string `gorm:"column:deploy_container" json:"deploy_container,omitempty"`
	DeployImage        string `gorm:"column:deploy_image" json:"deploy_image,omitempty"`
	DeployOldImage     string `gorm:"column:deploy_old_image" json:"deploy_old_image,omitempty"`
	DeployReplicas     int    `gorm:"column:deploy_replicas" json:"deploy_replicas,omitempty"`

	ErrorMessage string `gorm:"column:error_message" json:"error_message,omitempty"`

	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdPipelineStage) TableName() string { return "cicd_pipeline_stage" }

// ========== DTOs ==========

// StageDisplayInfo 阶段展示信息
type StageDisplayInfo struct {
	ID            int64  `json:"id"`
	Order         int    `json:"order"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Duration      string `json:"duration"`
	StartedAt     uint64 `json:"started_at"`
	FinishedAt    uint64 `json:"finished_at"`
	CanOperate    bool   `json:"can_operate"`
	HasLogs       bool   `json:"has_logs"`
	Logs          string `json:"logs,omitempty"`
	ErrorMsg      string `json:"error_msg,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	ConfigWarning string `json:"config_warning,omitempty"`

	ApprovalInfo *StageApprovalInfo `json:"approval_info,omitempty"`
	DeployInfo   *StageDeployInfo   `json:"deploy_info,omitempty"`
	SonarInfo    *StageSonarInfo    `json:"sonar_info,omitempty"`
}

// StageApprovalInfo 审批信息
type StageApprovalInfo struct {
	ApproverID   int64  `json:"approver_id,omitempty"`
	ApproverName string `json:"approver_name,omitempty"`
	Decision     string `json:"decision,omitempty"`
	Comment      string `json:"comment,omitempty"`
	ApprovedAt   uint64 `json:"approved_at,omitempty"`
}

// StageDeployInfo 部署信息
type StageDeployInfo struct {
	ClusterID    int64  `json:"cluster_id"`
	ClusterName  string `json:"cluster_name,omitempty"`
	Namespace    string `json:"namespace"`
	WorkloadKind string `json:"workload_kind"`
	WorkloadName string `json:"workload_name"`
	Container    string `json:"container"`
	Image        string `json:"image"`
	OldImage     string `json:"old_image,omitempty"`
	Replicas     int    `json:"replicas,omitempty"`
	DeployedAt   uint64 `json:"deployed_at,omitempty"`
}

// StageSonarInfo SonarQube 代码质量信息
type StageSonarInfo struct {
	ProjectKey        string  `json:"project_key"`
	ProjectName       string  `json:"project_name,omitempty"`
	QualityGate       string  `json:"quality_gate"`
	DashboardURL      string  `json:"dashboard_url,omitempty"`
	Bugs              int     `json:"bugs"`
	Vulnerabilities   int     `json:"vulnerabilities"`
	CodeSmells        int     `json:"code_smells"`
	Coverage          float64 `json:"coverage"`
	Duplications      float64 `json:"duplications"`
	LinesOfCode       int     `json:"lines_of_code"`
	SecurityHotspots  int     `json:"security_hotspots"`
	ReliabilityRating string  `json:"reliability_rating"`
	SecurityRating    string  `json:"security_rating"`
	Maintainability   string  `json:"maintainability_rating"`
	NewBugs           int     `json:"new_bugs,omitempty"`
	NewVulnerabilities int    `json:"new_vulnerabilities,omitempty"`
	NewCodeSmells     int     `json:"new_code_smells,omitempty"`
	NewCoverage       float64 `json:"new_coverage,omitempty"`
	ScanTime          uint64  `json:"scan_time,omitempty"`
}

// PipelineListItem 列表查询返回结构
type PipelineListItem struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	GitRepo             string `json:"git_repo"`
	GitBranch           string `json:"git_branch"`
	DeployMode          string `json:"deploy_mode"`
	JenkinsJob          string `json:"jenkins_job"`
	LanguageType        string `json:"language_type"`
	Status              string `json:"status"`
	LastRunStatus       string `json:"last_run_status"`
	LastRunTime         uint64 `json:"last_run_time"`
	LastBuildNumber     int    `json:"last_build_number"`
	AutoDeploy          bool   `json:"auto_deploy"`
	RequireApproval     bool   `json:"require_approval"`
	DeployEnv           string `json:"deploy_env"`
	EnvironmentID       int64  `json:"environment_id"`
	TargetClusterID     int64  `json:"target_cluster_id"`
	TargetNamespace     string `json:"target_namespace"`
	TargetWorkloadKind  string `json:"target_workload_kind"`
	TargetWorkloadName  string `json:"target_workload_name"`
	EnableCanary        bool   `gorm:"column:enable_canary" json:"enable_canary"`
	CanaryReplicas      int32  `gorm:"column:canary_replicas;default:1" json:"canary_replicas"`
	CanaryTrafficRatio  int32  `gorm:"column:canary_traffic_ratio;default:10" json:"canary_traffic_ratio"`
	CanaryDurationSec   int32  `gorm:"column:canary_duration_sec;default:300" json:"canary_duration_sec"`
	CanaryAutoPromote   bool   `gorm:"column:canary_auto_promote" json:"canary_auto_promote"`
	CanaryAnalysisRules string `gorm:"column:canary_analysis_rules;type:text" json:"canary_analysis_rules"`
	LastDeployStatus    string `json:"last_deploy_status"`
	LastRunImage        string `json:"last_run_image"`
	LastRunTag          string `json:"last_run_tag"`
	LastCommit          string `json:"last_commit"`
	LastCommitMsg       string `json:"last_commit_msg"`
	LastDuration        int    `json:"last_duration"`
	LastTriggerUser     string `json:"last_trigger_user"`
	CreatedAt           uint64 `json:"created_at"`
}

// ToPipelineListItem 转换为列表项
func (p *CicdPipeline) ToPipelineListItem() *PipelineListItem {
	return &PipelineListItem{
		ID:                  p.ID,
		Name:                p.Name,
		Description:         p.Description,
		GitRepo:             p.GitRepo,
		GitBranch:           p.GitBranch,
		DeployMode:          p.DeployMode,
		JenkinsJob:          p.JenkinsJob,
		LanguageType:        p.LanguageType,
		Status:              p.Status,
		LastRunStatus:       p.LastRunStatus,
		LastRunTime:         p.LastRunTime,
		LastBuildNumber:     p.LastBuildNumber,
		AutoDeploy:          p.AutoDeploy,
		RequireApproval:     p.RequireApproval,
		DeployEnv:           p.DeployEnv,
		EnvironmentID:       p.EnvironmentID,
		TargetClusterID:     p.TargetClusterID,
		TargetNamespace:     p.TargetNamespace,
		TargetWorkloadKind:  p.TargetWorkloadKind,
		TargetWorkloadName:  p.TargetWorkloadName,
		CanaryReplicas:        p.CanaryReplicas,
		CanaryTrafficRatio:    p.CanaryTrafficRatio,
		CanaryDurationSec:     p.CanaryDurationSec,
		CanaryAutoPromote:     p.CanaryAutoPromote,
		CanaryAnalysisRules:   p.CanaryAnalysisRules,
		EnableCanary:          p.EnableCanary,
		LastDeployStatus: p.LastDeployStatus,
		CreatedAt:       p.CreatedAt,
	}
}

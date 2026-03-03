package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Pipeline 状态常量
const (
	PipelineStatusIdle     = "idle"     // 空闲
	PipelineStatusRunning  = "running"  // 运行中
	PipelineStatusDisabled = "disabled" // 已禁用
)

// Pipeline 运行状态常量
const (
	PipelineRunStatusPending = "pending" // 等待中
	PipelineRunStatusRunning = "running" // 运行中
	PipelineRunStatusSuccess = "success" // 成功
	PipelineRunStatusFailed  = "failed"  // 失败
	PipelineRunStatusAborted = "aborted" // 已中止
)

// 触发类型常量
const (
	TriggerTypeManual    = "manual"    // 手动触发
	TriggerTypeWebhook   = "webhook"   // Webhook触发
	TriggerTypeScheduled = "scheduled" // 定时触发
)

// EnvVar 环境变量结构
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnvVars JSON数组类型
type EnvVars []EnvVar

func (e EnvVars) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

func (e *EnvVars) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, e)
}

// DeployConfig 部署配置结构
type DeployConfig struct {
	Namespace      string `json:"namespace"`
	DeploymentName string `json:"deployment_name"`
	Image          string `json:"image"`
	Replicas       int    `json:"replicas"`
	Strategy       string `json:"strategy"`
}

// JSONMap 通用JSON Map类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// 部署环境常量
const (
	DeployEnvDev     = "dev"     // 开发环境
	DeployEnvTest    = "test"    // 测试环境
	DeployEnvStaging = "staging" // 预发环境
	DeployEnvProd    = "prod"    // 生产环境
)

// CicdPipeline 对应表：cicd_pipeline
type CicdPipeline struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`

	// Git配置
	GitRepo   string `gorm:"column:git_repo" json:"git_repo"`
	GitBranch string `gorm:"column:git_branch" json:"git_branch"`

	// Jenkins配置
	JenkinsURL          string `gorm:"column:jenkins_url" json:"jenkins_url"`
	JenkinsJob          string `gorm:"column:jenkins_job" json:"jenkins_job"`
	JenkinsCredentialID string `gorm:"column:jenkins_credential_id" json:"jenkins_credential_id"`

	// 部署配置（构建成功后自动部署）
	AutoDeploy         bool   `gorm:"column:auto_deploy" json:"auto_deploy"`                    // 是否自动部署
	TargetClusterID    int64  `gorm:"column:target_cluster_id" json:"target_cluster_id"`        // 目标集群ID
	TargetNamespace    string `gorm:"column:target_namespace" json:"target_namespace"`          // 目标命名空间
	TargetWorkloadKind string `gorm:"column:target_workload_kind" json:"target_workload_kind"` // 工作负载类型(Deployment/StatefulSet/DaemonSet)
	TargetWorkloadName string `gorm:"column:target_workload_name" json:"target_workload_name"` // 工作负载名称
	TargetContainer    string `gorm:"column:target_container" json:"target_container"`         // 容器名称
	DeployEnv          string `gorm:"column:deploy_env" json:"deploy_env"`                      // 部署环境(dev/staging/prod)
	RequireApproval    bool   `gorm:"column:require_approval" json:"require_approval"`          // 是否需要审批

	// 最新部署信息
	LastDeployImage   string `gorm:"column:last_deploy_image" json:"last_deploy_image"`     // 最新部署镜像
	LastDeployDigest  string `gorm:"column:last_deploy_digest" json:"last_deploy_digest"`   // 最新部署镜像摘要
	LastDeployTime    uint64 `gorm:"column:last_deploy_time" json:"last_deploy_time"`       // 最新部署时间
	LastDeployStatus  string `gorm:"column:last_deploy_status" json:"last_deploy_status"`   // 最新部署状态
	LastDeployVersion string `gorm:"column:last_deploy_version" json:"last_deploy_version"` // 最新部署版本

	// 状态
	Status          string `gorm:"column:status" json:"status"`
	LastRunStatus   string `gorm:"column:last_run_status" json:"last_run_status"`
	LastRunTime     uint64 `gorm:"column:last_run_time" json:"last_run_time"`
	LastBuildNumber int    `gorm:"column:last_build_number" json:"last_build_number"`
	LastBuildURL    string `gorm:"column:last_build_url" json:"last_build_url"`

	// JSON配置
	EnvVars      EnvVars `gorm:"column:env_vars;type:json" json:"env_vars"`
	DeployConfig JSONMap `gorm:"column:deploy_config;type:json" json:"deploy_config"`
	Stages       JSONMap `gorm:"column:stages;type:json" json:"stages"`

	// 元数据
	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
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

	GitCommit      string `gorm:"column:git_commit" json:"git_commit"`
	GitBranch      string `gorm:"column:git_branch" json:"git_branch"`
	GitCommitMessage string `gorm:"column:git_commit_message" json:"git_commit_message"` // Git提交消息
	JenkinsBuildURL string `gorm:"column:jenkins_build_url" json:"jenkins_build_url"` // Jenkins构建URL

	// 构建产物
	ImageURL    string `gorm:"column:image_url" json:"image_url,omitempty"`       // 构建产出的镜像地址
	ImageDigest string `gorm:"column:image_digest" json:"image_digest,omitempty"` // 镜像 digest

	// 回调状态
	CallbackReceived uint8 `gorm:"column:callback_received" json:"callback_received"` // 是否已收到回调

	DurationSec  int     `gorm:"column:duration_sec" json:"duration_sec"`
	ConsoleLog   string  `gorm:"column:console_log" json:"console_log,omitempty"`
	StagesResult JSONMap `gorm:"column:stages_result;type:json" json:"stages_result"`
	ErrorMessage string  `gorm:"column:error_message" json:"error_message,omitempty"` // 错误信息

	StartedAt  uint64 `gorm:"column:started_at" json:"started_at"`
	FinishedAt uint64 `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdPipelineRun) TableName() string { return "cicd_pipeline_run" }

// ==================== 流水线阶段执行记录 ====================

// 阶段类型常量
const (
	StageTypeCheckout = "checkout" // 代码检出
	StageTypeBuild    = "build"    // 构建
	StageTypeTest     = "test"     // 测试
	StageTypePush     = "push"     // 推送镜像
	StageTypeApproval = "approval" // 人工审批
	StageTypeDeploy   = "deploy"   // 部署
)

// 阶段状态常量
const (
	StageStatusPending   = "pending"   // 等待中
	StageStatusRunning   = "running"   // 执行中
	StageStatusSuccess   = "success"   // 成功
	StageStatusFailed    = "failed"    // 失败
	StageStatusSkipped   = "skipped"   // 跳过
	StageStatusWaiting   = "waiting"   // 等待审批
	StageStatusAborted   = "aborted"   // 已中止
)

// CicdPipelineStage 流水线阶段执行记录表
type CicdPipelineStage struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RunID      int64  `gorm:"column:run_id;index" json:"run_id"`       // 关联流水线运行记录
	PipelineID int64  `gorm:"column:pipeline_id;index" json:"pipeline_id"` // 关联流水线
	StageOrder int    `gorm:"column:stage_order" json:"stage_order"` // 阶段顺序 (1,2,3...)
	StageType  string `gorm:"column:stage_type" json:"stage_type"`   // 阶段类型
	StageName  string `gorm:"column:stage_name" json:"stage_name"`   // 阶段名称
	Status     string `gorm:"column:status" json:"status"`           // 执行状态

	// 执行信息
	StartedAt   uint64 `gorm:"column:started_at" json:"started_at"`     // 开始时间
	FinishedAt  uint64 `gorm:"column:finished_at" json:"finished_at"`   // 结束时间
	DurationSec int    `gorm:"column:duration_sec" json:"duration_sec"` // 执行时长(秒)
	Logs        string `gorm:"column:logs;type:longtext" json:"logs"`   // 阶段日志

	// Jenkins 构建信息（适用于 build/test/push 类型）
	JenkinsStageID string `gorm:"column:jenkins_stage_id" json:"jenkins_stage_id,omitempty"` // Jenkins 阶段ID

	// 审批信息（适用于 approval 类型）
	ApprovalUserID   int64  `gorm:"column:approval_user_id" json:"approval_user_id,omitempty"`     // 审批人
	ApprovalComment  string `gorm:"column:approval_comment" json:"approval_comment,omitempty"`   // 审批评论
	ApprovalDecision string `gorm:"column:approval_decision" json:"approval_decision,omitempty"` // 审批决定: approved/rejected

	// 部署信息（适用于 deploy 类型）
	DeployClusterID    int64  `gorm:"column:deploy_cluster_id" json:"deploy_cluster_id,omitempty"`       // 目标集群
	DeployNamespace    string `gorm:"column:deploy_namespace" json:"deploy_namespace,omitempty"`       // 目标命名空间
	DeployWorkloadKind string `gorm:"column:deploy_workload_kind" json:"deploy_workload_kind,omitempty"` // 工作负载类型
	DeployWorkloadName string `gorm:"column:deploy_workload_name" json:"deploy_workload_name,omitempty"` // 工作负载名称
	DeployContainer    string `gorm:"column:deploy_container" json:"deploy_container,omitempty"`       // 容器名称
	DeployImage        string `gorm:"column:deploy_image" json:"deploy_image,omitempty"`               // 部署的新镜像
	DeployOldImage     string `gorm:"column:deploy_old_image" json:"deploy_old_image,omitempty"`       // 部署前的旧镜像
	DeployReplicas     int    `gorm:"column:deploy_replicas" json:"deploy_replicas,omitempty"`         // 副本数

	// 错误信息
	ErrorMessage string `gorm:"column:error_message" json:"error_message,omitempty"`

	// 元数据
	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdPipelineStage) TableName() string { return "cicd_pipeline_stage" }

// StageDisplayInfo 阶段展示信息（前端使用）
type StageDisplayInfo struct {
	ID           int64  `json:"id"`
	Order        int    `json:"order"`
	Type         string `json:"type"`          // checkout/build/test/push/approval/deploy
	Name         string `json:"name"`
	Status       string `json:"status"`
	Duration     string `json:"duration"`      // 格式化后的时长
	StartedAt    uint64 `json:"started_at"`
	FinishedAt   uint64 `json:"finished_at"`
	CanOperate   bool   `json:"can_operate"`   // 是否可操作(审批/部署)
	HasLogs      bool   `json:"has_logs"`      // 是否有日志
	Logs         string `json:"logs,omitempty"`         // 部署日志（仅部署阶段返回）
	ErrorMsg     string `json:"error_msg,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"` // 错误信息（部署失败时）
	ConfigWarning string `json:"config_warning,omitempty"` // 配置警告（如部署参数不完整）
	
	// 审批信息
	ApprovalInfo *StageApprovalInfo `json:"approval_info,omitempty"`
	// 部署信息
	DeployInfo   *StageDeployInfo   `json:"deploy_info,omitempty"`
}

// StageApprovalInfo 审批信息
type StageApprovalInfo struct {
	ApproverID   int64  `json:"approver_id,omitempty"`
	ApproverName string `json:"approver_name,omitempty"`
	Decision     string `json:"decision,omitempty"` // approved/rejected
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
	Image        string `json:"image"`                       // 部署的新镜像
	OldImage     string `json:"old_image,omitempty"`          // 部署前的旧镜像
	Replicas     int    `json:"replicas,omitempty"`
	DeployedAt   uint64 `json:"deployed_at,omitempty"`        // 部署完成时间
}

// PipelineListItem 列表查询返回结构（去除敏感字段）
type PipelineListItem struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	GitRepo         string `json:"git_repo"`
	GitBranch       string `json:"git_branch"`
	JenkinsJob      string `json:"jenkins_job"`
	Status          string `json:"status"`
	LastRunStatus   string `json:"last_run_status"`
	LastRunTime     uint64 `json:"last_run_time"`
	LastBuildNumber int    `json:"last_build_number"`
	CreatedAt       uint64 `json:"created_at"`
}

// ToPipelineListItem 转换为列表项
func (p *CicdPipeline) ToPipelineListItem() *PipelineListItem {
	return &PipelineListItem{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		GitRepo:         p.GitRepo,
		GitBranch:       p.GitBranch,
		JenkinsJob:      p.JenkinsJob,
		Status:          p.Status,
		LastRunStatus:   p.LastRunStatus,
		LastRunTime:     p.LastRunTime,
		LastBuildNumber: p.LastBuildNumber,
		CreatedAt:       p.CreatedAt,
	}
}

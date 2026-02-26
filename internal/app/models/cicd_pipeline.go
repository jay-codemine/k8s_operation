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

	GitCommit string `gorm:"column:git_commit" json:"git_commit"`
	GitBranch string `gorm:"column:git_branch" json:"git_branch"`

	DurationSec  int     `gorm:"column:duration_sec" json:"duration_sec"`
	ConsoleLog   string  `gorm:"column:console_log" json:"console_log,omitempty"`
	StagesResult JSONMap `gorm:"column:stages_result;type:json" json:"stages_result"`

	StartedAt  uint64 `gorm:"column:started_at" json:"started_at"`
	FinishedAt uint64 `gorm:"column:finished_at" json:"finished_at"`
	CreatedAt  uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt uint64 `gorm:"column:modified_at" json:"modified_at"`
}

func (CicdPipelineRun) TableName() string { return "cicd_pipeline_run" }

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

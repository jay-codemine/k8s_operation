package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/valid"
)

// ==================== 创建流水线 ====================

type PipelineCreateRequest struct {
	Name        string           `json:"name" valid:"name"`
	Description string           `json:"description" valid:"description"`
	GitRepo     string           `json:"git_repo" valid:"git_repo"`
	GitBranch   string           `json:"git_branch" valid:"git_branch"`
	JenkinsURL  string           `json:"jenkins_url" valid:"jenkins_url"`
	JenkinsJob  string           `json:"jenkins_job" valid:"jenkins_job"`
	EnvVars     []models.EnvVar  `json:"env_vars"`
	DeployConfig map[string]any  `json:"deploy_config"`
	
	// 部署配置
	AutoDeploy         bool   `json:"auto_deploy"`          // 是否自动部署
	TargetClusterID    int64  `json:"target_cluster_id"`    // 目标集群ID
	TargetNamespace    string `json:"target_namespace"`     // 目标命名空间
	TargetWorkloadKind string `json:"target_workload_kind"` // 工作负载类型
	TargetWorkloadName string `json:"target_workload_name"` // 工作负载名称
	TargetContainer    string `json:"target_container"`     // 容器名称
	DeployEnv          string `json:"deploy_env"`           // 部署环境
	RequireApproval    bool   `json:"require_approval"`     // 是否需要审批
}

func NewPipelineCreateRequest() *PipelineCreateRequest {
	return &PipelineCreateRequest{
		GitBranch: "main",
	}
}

func ValidPipelineCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":        []string{"required", "between:1,100"},
		"git_repo":    []string{"required", "url"},
		"jenkins_job": []string{"required", "between:1,100"},
	}
	messages := govalidator.MapData{
		"name":        []string{"required:流水线名称不能为空", "between:流水线名称长度应在1-100之间"},
		"git_repo":    []string{"required:Git仓库地址不能为空", "url:Git仓库地址格式无效"},
		"jenkins_job": []string{"required:Jenkins Job名称不能为空", "between:Jenkins Job名称长度应在1-100之间"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 更新流水线 ====================

type PipelineUpdateRequest struct {
	ID          int64            `json:"id" valid:"id"`
	Name        string           `json:"name" valid:"name"`
	Description string           `json:"description" valid:"description"`
	GitRepo     string           `json:"git_repo" valid:"git_repo"`
	GitBranch   string           `json:"git_branch" valid:"git_branch"`
	JenkinsURL  string           `json:"jenkins_url" valid:"jenkins_url"`
	JenkinsJob  string           `json:"jenkins_job" valid:"jenkins_job"`
	Status      string           `json:"status" valid:"status"`
	EnvVars     []models.EnvVar  `json:"env_vars"`
	DeployConfig map[string]any  `json:"deploy_config"`
	
	// 部署配置
	AutoDeploy         *bool   `json:"auto_deploy"`          // 是否自动部署
	TargetClusterID    *int64  `json:"target_cluster_id"`    // 目标集群ID
	TargetNamespace    *string `json:"target_namespace"`     // 目标命名空间
	TargetWorkloadKind *string `json:"target_workload_kind"` // 工作负载类型
	TargetWorkloadName *string `json:"target_workload_name"` // 工作负载名称
	TargetContainer    *string `json:"target_container"`     // 容器名称
	DeployEnv          *string `json:"deploy_env"`           // 部署环境
	RequireApproval    *bool   `json:"require_approval"`     // 是否需要审批
}

func NewPipelineUpdateRequest() *PipelineUpdateRequest {
	return &PipelineUpdateRequest{}
}

func ValidPipelineUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":          []string{"required"},
		"name":        []string{"between:1,100"},
		"git_repo":    []string{"url"},
		"jenkins_job": []string{"between:1,100"},
		"status":      []string{"in:idle,running,disabled"},
	}
	messages := govalidator.MapData{
		"id":          []string{"required:流水线ID不能为空"},
		"name":        []string{"between:流水线名称长度应在1-100之间"},
		"git_repo":    []string{"url:Git仓库地址格式无效"},
		"jenkins_job": []string{"between:Jenkins Job名称长度应在1-100之间"},
		"status":      []string{"in:状态值无效，可选值: idle, running, disabled"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 流水线ID请求 ====================

type PipelineIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidPipelineIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:流水线ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 流水线列表请求 ====================

type PipelineListRequest struct {
	Page     int    `form:"page" valid:"page"`
	PageSize int    `form:"page_size" valid:"page_size"`
	Keyword  string `form:"keyword" valid:"keyword"`
	Status   string `form:"status" valid:"status"`
}

func NewPipelineListRequest() *PipelineListRequest {
	return &PipelineListRequest{
		Page:     1,
		PageSize: 10,
	}
}

func ValidPipelineListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return nil // 全部可选
}

// ==================== 运行流水线请求 ====================

type PipelineRunRequest struct {
	ID        int64             `json:"id" valid:"id"`
	Branch    string            `json:"branch"`     // 可选：覆盖默认分支
	EnvVars   map[string]string `json:"env_vars"`   // 可选：覆盖环境变量
	Force     bool              `json:"force"`      // 强制运行：自动清理旧的失败/运行中构建
	
	// 运行时部署配置（可覆盖流水线默认配置）
	AutoDeploy         *bool   `json:"auto_deploy"`          // 是否自动部署
	TargetClusterID    *int64  `json:"target_cluster_id"`    // 目标集群ID
	TargetNamespace    *string `json:"target_namespace"`     // 目标命名空间
	TargetWorkloadKind *string `json:"target_workload_kind"` // 工作负载类型
	TargetWorkloadName *string `json:"target_workload_name"` // 工作负载名称
	TargetContainer    *string `json:"target_container"`     // 容器名称
}

func ValidPipelineRunRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:流水线ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 停止流水线请求 ====================

type PipelineStopRequest struct {
	ID          int64 `json:"id" valid:"id"`
	BuildNumber int   `json:"build_number"` // 可选：指定构建号，不传则停止最新的
}

func ValidPipelineStopRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:流水线ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 获取流水线日志请求 ====================

type PipelineLogsRequest struct {
	ID          int64 `form:"id" valid:"id"`
	BuildNumber int   `form:"build_number"` // 可选：指定构建号，不传则获取最新的
	StartLine   int   `form:"start_line"`   // 可选：从第几行开始（用于增量获取）
}

func ValidPipelineLogsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:流水线ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 获取运行历史请求 ====================

type PipelineHistoryRequest struct {
	ID       int64 `form:"id" valid:"id"`
	Page     int   `form:"page"`
	PageSize int   `form:"page_size"`
}

func NewPipelineHistoryRequest() *PipelineHistoryRequest {
	return &PipelineHistoryRequest{
		Page:     1,
		PageSize: 10,
	}
}

func ValidPipelineHistoryRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:流水线ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== Jenkins构建状态回调（生产级） ====================

// PipelineCallbackRequest Jenkins 构建回调请求
// 幂等键: job_name + build_number 或 pipeline_id + build_number
type PipelineCallbackRequest struct {
	// 必须字段
	JobName     string `json:"job_name" valid:"job_name"`         // Jenkins Job 名称
	BuildNumber int    `json:"build_number" valid:"build_number"` // 构建号
	Status      string `json:"status" valid:"status"`             // SUCCESS / FAILURE / ABORTED

	// 平台关联字段
	PipelineID int64  `json:"pipeline_id" valid:"pipeline_id"` // 流水线ID（用于快速匹配）
	RequestID  string `json:"request_id" valid:"request_id"`   // 请求ID（用于日志追踪）

	// 构建产物 - 支持 image 或 image_url
	Image       string `json:"image" valid:"image"`               // 构建产出的镜像地址
	ImageURL    string `json:"image_url" valid:"image_url"`       // 构建产出的镜像地址（兼容旧字段）
	ImageDigest string `json:"image_digest" valid:"image_digest"` // 镜像 digest (e.g., sha256:xxx)

	// 额外信息
	Duration int    `json:"duration" valid:"duration"` // 构建耗时(秒)
	Message  string `json:"message" valid:"message"`   // 错误或补充信息

	// HMAC 签名（防伪造，放在 Header: X-Signature）
	// 签名算法: HMAC-SHA256(secret, job_name+build_number+status)
}

func ValidPipelineCallbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"job_name":     []string{"required"},
		"build_number": []string{"required"},
		"status":       []string{"required", "in:SUCCESS,FAILURE,ABORTED"},
	}
	messages := govalidator.MapData{
		"job_name":     []string{"required:Job名称不能为空"},
		"build_number": []string{"required:构建号不能为空"},
		"status":       []string{"required:状态不能为空", "in:状态值无效，可选值: SUCCESS, FAILURE, ABORTED"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== Git 仓库操作 ====================

// GitBranchesRequest 获取分支列表请求
type GitBranchesRequest struct {
	RepoURL      string `json:"repo_url" valid:"repo_url"`
	CredentialID string `json:"credential_id"` // 可选：Jenkins凭证ID
}

func ValidGitBranchesRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"repo_url": []string{"required"},
	}
	messages := govalidator.MapData{
		"repo_url": []string{"required:Git仓库地址不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// GitValidateRequest 验证仓库连接请求
type GitValidateRequest struct {
	RepoURL      string `json:"repo_url" valid:"repo_url"`
	CredentialID string `json:"credential_id"` // 可选：Jenkins凭证ID
}

func ValidGitValidateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"repo_url": []string{"required"},
	}
	messages := govalidator.MapData{
		"repo_url": []string{"required:Git仓库地址不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// GitBranch 分支信息
type GitBranch struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// ==================== Jenkins 阶段回调（实时更新UI） ====================

// StageCallbackRequest Jenkins 阶段回调请求
type StageCallbackRequest struct {
	JobName     string `json:"job_name" valid:"job_name"`         // Jenkins Job 名称
	BuildNumber int    `json:"build_number" valid:"build_number"` // 构建号
	PipelineID  int64  `json:"pipeline_id" valid:"pipeline_id"`   // 流水线ID
	StageType   string `json:"stage_type" valid:"stage_type"`     // 阶段类型: checkout/build/test/push
	Status      string `json:"status" valid:"status"`             // 阶段状态: running/success/failed
}

func ValidStageCallbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"job_name":     []string{"required"},
		"build_number": []string{"required"},
		"stage_type":   []string{"required", "in:checkout,build,test,push"},
		"status":       []string{"required", "in:running,success,failed"},
	}
	messages := govalidator.MapData{
		"job_name":     []string{"required:Job名称不能为空"},
		"build_number": []string{"required:构建号不能为空"},
		"stage_type":   []string{"required:阶段类型不能为空", "in:阶段类型无效"},
		"status":       []string{"required:状态不能为空", "in:状态值无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

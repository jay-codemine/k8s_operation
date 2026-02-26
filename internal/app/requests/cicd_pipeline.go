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

// ==================== Jenkins构建状态回调 ====================

type JenkinsBuildStatusCallback struct {
	JobName     string `json:"job_name"`
	BuildNumber int    `json:"build_number"`
	Status      string `json:"status"` // SUCCESS / FAILURE / ABORTED
	Duration    int    `json:"duration"`
	Message     string `json:"message"`
}

func ValidJenkinsBuildStatusCallback(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"job_name":     []string{"required"},
		"build_number": []string{"required"},
		"status":       []string{"required", "in:SUCCESS,FAILURE,ABORTED"},
	}
	messages := govalidator.MapData{
		"job_name":     []string{"required:Job名称不能为空"},
		"build_number": []string{"required:构建号不能为空"},
		"status":       []string{"required:状态不能为空", "in:状态值无效"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

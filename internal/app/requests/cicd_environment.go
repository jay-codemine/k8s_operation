package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ==================== 环境管理 ====================

// EnvironmentCreateRequest 创建环境请求
type EnvironmentCreateRequest struct {
	Name            string   `json:"name" valid:"name"`                 // 环境名称(dev/staging/prod)
	DisplayName     string   `json:"display_name" valid:"display_name"` // 显示名称
	Description     string   `json:"description"`                       // 描述
	ClusterID       int64    `json:"cluster_id" valid:"cluster_id"`     // 关联集群ID
	Namespace       string   `json:"namespace"`                         // 默认命名空间
	Color           string   `json:"color"`                             // 环境颜色标识
	SortOrder       int      `json:"sort_order"`                        // 排序
	RequireApproval bool     `json:"require_approval"`                  // 是否需要审批
	ApprovalUserIDs []int64  `json:"approval_user_ids"`                 // 审批人员ID列表
}

func ValidEnvironmentCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":         []string{"required", "between:1,50"},
		"display_name": []string{"required", "between:1,100"},
		"cluster_id":   []string{"required"},
	}
	messages := govalidator.MapData{
		"name":         []string{"required:环境名称不能为空", "between:环境名称长度应在1-50之间"},
		"display_name": []string{"required:显示名称不能为空", "between:显示名称长度应在1-100之间"},
		"cluster_id":   []string{"required:关联集群不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// EnvironmentUpdateRequest 更新环境请求
type EnvironmentUpdateRequest struct {
	ID              int64    `json:"id" valid:"id"`
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	ClusterID       *int64   `json:"cluster_id"`
	Namespace       string   `json:"namespace"`
	Color           string   `json:"color"`
	SortOrder       *int     `json:"sort_order"`
	RequireApproval *bool    `json:"require_approval"`
	ApprovalUserIDs []int64  `json:"approval_user_ids"`
}

func ValidEnvironmentUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:环境ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// EnvironmentIDRequest 环境ID请求
type EnvironmentIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidEnvironmentIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:环境ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// EnvironmentListRequest 环境列表请求
type EnvironmentListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
}

func NewEnvironmentListRequest() *EnvironmentListRequest {
	return &EnvironmentListRequest{
		Page:     1,
		PageSize: 20,
	}
}

func ValidEnvironmentListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return nil // 全部可选
}

// ==================== 审批流程 ====================

// ApprovalCreateRequest 创建审批请求
type ApprovalCreateRequest struct {
	PipelineID    int64  `json:"pipeline_id" valid:"pipeline_id"`
	PipelineRunID int64  `json:"pipeline_run_id"`
	EnvName       string `json:"env_name" valid:"env_name"`
	Image         string `json:"image" valid:"image"`
	ImageDigest   string `json:"image_digest"`
	RequestReason string `json:"request_reason"`
}

func ValidApprovalCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"pipeline_id": []string{"required"},
		"env_name":    []string{"required"},
		"image":       []string{"required"},
	}
	messages := govalidator.MapData{
		"pipeline_id": []string{"required:流水线ID不能为空"},
		"env_name":    []string{"required:目标环境不能为空"},
		"image":       []string{"required:镜像地址不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ApprovalActionRequest 审批操作请求
type ApprovalActionRequest struct {
	ID     int64  `json:"id" valid:"id"`
	Action string `json:"action" valid:"action"` // approve/reject
	Reason string `json:"reason"`                // 审批意见
}

func ValidApprovalActionRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":     []string{"required"},
		"action": []string{"required", "in:approve,reject"},
	}
	messages := govalidator.MapData{
		"id":     []string{"required:审批ID不能为空"},
		"action": []string{"required:操作类型不能为空", "in:操作类型无效，可选值: approve, reject"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ApprovalIDRequest 审批ID请求
type ApprovalIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidApprovalIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:审批ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ApprovalListRequest 审批列表请求
type ApprovalListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	PipelineID int64  `form:"pipeline_id"`
	Status     string `form:"status"` // pending/approved/rejected/expired
}

func NewApprovalListRequest() *ApprovalListRequest {
	return &ApprovalListRequest{
		Page:     1,
		PageSize: 10,
	}
}

func ValidApprovalListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return nil // 全部可选
}

// ==================== 流水线阶段 ====================

// StageListRequest 获取运行记录的阶段列表请求
type StageListRequest struct {
	RunID int64 `form:"run_id" valid:"run_id"`
}

func ValidStageListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"run_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"run_id": []string{"required:运行记录ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// StageIDRequest 阶段ID请求
type StageIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidStageIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:阶段ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// StageLogsRequest 获取阶段日志请求
type StageLogsRequest struct {
	ID int64 `form:"id" valid:"id"`
}

func ValidStageLogsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:阶段ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// StageApproveRequest 阶段审批请求
type StageApproveRequest struct {
	StageID int64  `json:"stage_id" valid:"stage_id"`
	Action  string `json:"action" valid:"action"` // approve/reject
	Comment string `json:"comment"`               // 审批意见
}

func ValidStageApproveRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"stage_id": []string{"required"},
		"action":   []string{"required", "in:approve,reject"},
	}
	messages := govalidator.MapData{
		"stage_id": []string{"required:阶段ID不能为空"},
		"action":   []string{"required:操作类型不能为空", "in:操作类型无效，可选值: approve, reject"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// StageDeployRequest 阶段部署请求
type StageDeployRequest struct {
	StageID      int64  `json:"stage_id" valid:"stage_id"`
	ClusterID    int64  `json:"cluster_id"`              // 可覆盖默认集群
	Namespace    string `json:"namespace"`               // 可覆盖默认命名空间
	WorkloadKind string `json:"workload_kind"`           // 可覆盖默认工作负载类型
	WorkloadName string `json:"workload_name"`           // 可覆盖默认工作负载名称
	Container    string `json:"container"`               // 可覆盖默认容器
	Image        string `json:"image"`                   // 可覆盖默认镜像
}

func ValidStageDeployRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"stage_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"stage_id": []string{"required:阶段ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

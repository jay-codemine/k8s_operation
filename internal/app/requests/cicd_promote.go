package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ========== 环境目标配置（cicd_pipeline_target）==========

// PipelineTargetItem 单个环境部署目标
type PipelineTargetItem struct {
	Env             string `json:"env"`              // 环境标识(dev/test/staging/prod)
	ClusterID       int64  `json:"cluster_id"`       // 目标集群ID
	Namespace       string `json:"namespace"`        // 目标命名空间
	WorkloadKind    string `json:"workload_kind"`    // 工作负载类型
	WorkloadName    string `json:"workload_name"`    // 工作负载名称
	Container       string `json:"container"`        // 容器名称
	AutoDeploy      bool   `json:"auto_deploy"`      // 是否自动部署
	RequireApproval bool   `json:"require_approval"` // 是否需要审批
	PromoteFrom     string `json:"promote_from"`     // 上游来源环境
	SortOrder       int    `json:"sort_order"`       // 顺序
}

// PipelineTargetSaveRequest 批量保存某条流水线的环境目标（全量覆盖）
type PipelineTargetSaveRequest struct {
	PipelineID int64                `json:"pipeline_id" valid:"pipeline_id"`
	Targets    []PipelineTargetItem `json:"targets" valid:"targets"`
}

func ValidPipelineTargetSaveRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"pipeline_id": []string{"required"},
	}
	messages := govalidator.MapData{
		"pipeline_id": []string{"required: pipeline_id不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// PipelineTargetDeleteRequest 删除单个环境目标
type PipelineTargetDeleteRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidPipelineTargetDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required: id不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ========== 镜像晋级 ==========

// PipelinePromoteRequest 镜像晋级请求
//
// 优雅方式：用户只需选「目标环境」，集群/命名空间/工作负载由 cicd_pipeline_target 预先绑定；
// 待晋级的不可变镜像由 source_run_id（流水线运行记录）或 source_release_id（已有发布单）解析，
// 也支持直接显式传 image_repo/tag/digest。
type PipelinePromoteRequest struct {
	PipelineID int64  `json:"pipeline_id" valid:"pipeline_id"` // 流水线ID
	TargetEnv  string `json:"target_env" valid:"target_env"`   // 目标环境(dev/test/staging/prod)

	// 镜像来源（三选一，优先级：显式镜像 > source_release_id > source_run_id）
	SourceRunID     int64  `json:"source_run_id"`     // 构建产出镜像的流水线运行记录ID
	SourceReleaseID int64  `json:"source_release_id"` // 已有发布单ID（从其它环境晋级）
	ImageRepo       string `json:"image_repo"`        // 显式镜像仓库
	ImageTag        string `json:"image_tag"`         // 显式镜像tag
	ImageDigest     string `json:"image_digest"`      // 显式镜像digest

	Reason string `json:"reason"` // 晋级说明
}

func ValidPipelinePromoteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"pipeline_id": []string{"required"},
		"target_env":  []string{"required"},
	}
	messages := govalidator.MapData{
		"pipeline_id": []string{"required: pipeline_id不能为空"},
		"target_env":  []string{"required: target_env不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

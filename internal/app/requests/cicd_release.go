package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ========== 创建发布单 ==========

func NewCicdReleaseCreateRequest() *CicdReleaseCreateRequest {
	return &CicdReleaseCreateRequest{
		Namespace:    "default",
		WorkloadKind: "Deployment",
		Strategy:     "rolling",
		TimeoutSec:   300,
		Concurrency:  3,
	}
}

type CicdReleaseCreateRequest struct {
	AppName       string `json:"app_name" valid:"app_name"`
	Namespace     string `json:"namespace" valid:"namespace"`
	WorkloadKind  string `json:"workload_kind" valid:"workload_kind"`
	WorkloadName  string `json:"workload_name" valid:"workload_name"`
	ContainerName string `json:"container_name" valid:"container_name"`

	Strategy    string `json:"strategy" valid:"strategy"`
	TimeoutSec  uint32 `json:"timeout_sec" valid:"timeout_sec"`
	Concurrency uint32 `json:"concurrency" valid:"concurrency"`

	ImageRepo   string `json:"image_repo" valid:"image_repo"`
	ImageTag    string `json:"image_tag" valid:"image_tag"`
	ImageDigest string `json:"image_digest" valid:"image_digest"`

	ClusterIDs []int64 `json:"cluster_ids" valid:"cluster_ids"`

	RequestID string `json:"request_id"` // 可选，幂等校验用
}

func ValidCicdReleaseCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"app_name":       []string{"required"},
		"namespace":      []string{"required"},
		"workload_name":  []string{"required"},
		"container_name": []string{"required"},
		"image_repo":     []string{"required"},
		// image_tag / image_digest 至少一个
		"cluster_ids": []string{"required"},
	}
	messages := govalidator.MapData{
		"app_name":       []string{"required: app_name不能为空"},
		"namespace":      []string{"required: namespace不能为空"},
		"workload_name":  []string{"required: workload_name不能为空"},
		"container_name": []string{"required: container_name不能为空"},
		"image_repo":     []string{"required: image_repo不能为空"},
		"cluster_ids":    []string{"required: cluster_ids不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ========== 发布单列表查询 ==========

func NewCicdReleaseListRequest() *CicdReleaseListRequest {
	return &CicdReleaseListRequest{
		Page:     1,
		PageSize: 20,
	}
}

type CicdReleaseListRequest struct {
	Page     int    `form:"page" valid:"page"`
	PageSize int    `form:"page_size" valid:"page_size"`
	AppName  string `form:"app_name" valid:"app_name"`
	Status   string `form:"status" valid:"status"`
}

func ValidCicdReleaseListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return nil // 全部可选
}

// ========== 发布单ID请求（取消/重试） ==========

type CicdReleaseIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidCicdReleaseIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required: id不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ========== Jenkins构建回调 ==========

type CicdBuildCallbackRequest struct {
	BuildID     int64  `json:"build_id"`     // Jenkins 构建ID
	Status      string `json:"status"`       // SUCCESS / FAILURE / ABORTED
	ImageRepo   string `json:"image_repo"`   // 构建出的镜像Repo
	ImageTag    string `json:"image_tag"`    // 构建出的镜像Tag
	ImageDigest string `json:"image_digest"` // 镜像摘要（可选）
	Message     string `json:"message"`      // 构建消息
}

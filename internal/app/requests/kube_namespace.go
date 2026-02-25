package requests

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

//
// ================= Namespace 创建 =================
//

func NewKubeNamespaceCreateRequest() *KubeNamespaceCreateRequest {
	return &KubeNamespaceCreateRequest{}
}

type KubeNamespaceCreateRequest struct {
	// Namespace 名称（必填）
	Name string `json:"name" form:"name" valid:"name"`

	// 描述信息（可选） -> 写入 annotation
	Description string `json:"description" form:"description" valid:"-"`

	// Labels（可选）
	Labels      map[string]string `json:"labels" form:"labels" swaggertype:"object"`
	QuotaCPU    string            `json:"quota_cpu" form:"quota_cpu" valid:"-"`
	QuotaMemory string            `json:"quota_memory" form:"quota_memory" valid:"-"`
	QuotaPods   string            `json:"quota_pods" form:"quota_pods" valid:"-"`
	Annotations map[string]string `json:"annotations" form:"annotations" swaggertype:"object"`
}

func ValidKubeNamespaceCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(
		data,
		govalidator.MapData{
			"name": {"required"},
		},
		govalidator.MapData{
			"name": {"required: namespace 名称不能为空"},
		},
	)
}

//
// ================= Namespace 列表 =================
//

func NewKubeNamespaceListRequest() *KubeNamespaceListRequest {
	return &KubeNamespaceListRequest{}
}

type KubeNamespaceListRequest struct {
	// 名称模糊查询
	Name string `json:"name" form:"name" valid:"name"`

	Page  int `json:"page"  form:"page"  valid:"page"`
	Limit int `json:"limit" form:"limit" valid:"limit"`
}

func ValidKubeNamespaceListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  {"numeric", "min:1"},
		"limit": {"numeric", "min:1", "max:1000"},
	}

	msgs := govalidator.MapData{
		"page": {
			"numeric: 页码必须为数字",
			"min: 页码不能小于 1",
		},
		"limit": {
			"numeric: 每页数量必须为数字",
			"min: 每页数量不能小于 1",
			"max: 每页数量不能超过 1000",
		},
	}

	return valid.ValidateOptions(data, rules, msgs)
}

//
// ================= Namespace 详情 =================
//

func NewKubeNamespaceDetailRequest() *KubeNamespaceDetailRequest {
	return &KubeNamespaceDetailRequest{}
}

type KubeNamespaceDetailRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeNamespaceDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(
		data,
		govalidator.MapData{
			"name": {"required"},
		},
		govalidator.MapData{
			"name": {"required: name 不能为空"},
		},
	)
}

//
// ================= Namespace 删除 =================
//

func NewKubeNamespaceDeleteRequest() *KubeNamespaceDeleteRequest {
	return &KubeNamespaceDeleteRequest{}
}

type KubeNamespaceDeleteRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeNamespaceDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(
		data,
		govalidator.MapData{
			"name": {"required"},
		},
		govalidator.MapData{
			"name": {"required: name 不能为空"},
		},
	)
}

//
// ================= Namespace Patch（Labels / Annotations） =================
//

func NewKubeNamespaceUpdateRequest() *KubeNamespaceUpdateRequest {
	return &KubeNamespaceUpdateRequest{}
}

// KubeNamespaceUpdateRequest 适配 Namespace 更新（完整 JSON/YAML Patch）
type KubeNamespaceUpdateRequest struct {
	// Namespace 名称（必填）
	Name string `json:"name" valid:"name"`

	// 更新内容（一般是 Namespace 的 metadata JSON）
	Content json.RawMessage `json:"content" valid:"content"`
}

func ValidKubeNamespaceUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":    []string{"required"},
		"content": []string{"required"},
	}

	messages := govalidator.MapData{
		"name":    []string{"required: name 不能为空"},
		"content": []string{"required: content 不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// ================= Namespace 标签变更 =================

func NewKubeNamespaceLabelPatchRequest() *KubeNamespaceLabelPatchRequest {
	return &KubeNamespaceLabelPatchRequest{}
}

type KubeNamespaceLabelPatchRequest struct {
	Name   string            `json:"name" valid:"name"`
	Add    map[string]string `json:"add"`    // 添加/更新的标签
	Remove []string          `json:"remove"` // 要删除的标签 key
}

func ValidKubeNamespaceLabelPatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

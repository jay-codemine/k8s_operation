package requests

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

type KubePodCreateRequest struct {
	Namespace     string            `json:"namespace" form:"namespace" valid:"namespace"`
	Name          string            `json:"name"      form:"name"      valid:"name"`
	ContainerName string            `json:"containerName,omitempty" form:"containerName"`
	Image         string            `json:"image"     form:"image"     valid:"image"`
	Labels        map[string]string `json:"labels,omitempty" form:"labels"`
}

func NewKubePodCreateRequest() *KubePodCreateRequest {
	return &KubePodCreateRequest{}
}
func ValidKubePodCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"image":     {"required"},
	}
	messages := govalidator.MapData{
		"namespace": {"required: namespace不能为空"},
		"name":      {"required: name不能为空"},
		"image":     {"required: image不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 列表请求 DTO
type KubePodListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" form:"page" valid:"page"`    // 页码
	Limit int `json:"limit" form:"limit" valid:"limit"` // 每页条数
}

func NewKubePodListRequest() *KubePodListRequest {
	return &KubePodListRequest{}
}

func ValidKubePodListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"max:63"}, // 可选
		"page":      {"min:1"},  // 可选
		"limit":     {"min:1", "max:1000"},
	}
	return valid.ValidateOptions(data, rules, nil)
}

// KubePodDetailRequest 详情请求 DTO
type KubePodDetailRequest struct {
	Name      string `json:"name" form:"name" valid:"name"`
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
}

func NewKubePodDetailRequest() *KubePodDetailRequest {
	return &KubePodDetailRequest{}
}

func ValidKubePodDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
	}
	messages := govalidator.MapData{
		"namespace": {"required: namespace不能为空"},
		"name":      {"required: name不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 更新请求 DTO
type KubePodUpdateRequest struct {
	KubeCommonRequest
	Content json.RawMessage `json:"content" form:"content" valid:"content" swaggertype:"object"` // 原始 Pod JSON
}

func NewKubePodUpdateRequest() *KubePodUpdateRequest {
	return &KubePodUpdateRequest{}
}

type PatchPodImageRequest struct {
	Name      string `json:"name" form:"name" valid:"name"`
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Container string `json:"container" form:"container" valid:"container"`
	NewImage  string `json:"new_image" form:"new_image" valid:"new_image"`
}

func NewPatchPodImageRequest() *PatchPodImageRequest {
	return &PatchPodImageRequest{}
}

func ValidKubePodPatchContainerImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"new_image": {"required"},
	}
	message := govalidator.MapData{
		"namespace": {"required: namespace不能为空"},
		"name":      {"required: name不能为空"},
		"new_image": {"required: new_image不能为空"},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, message)
}

func ValidKubePodUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"required"},
		"content":   {"required"},
	}
	messages := govalidator.MapData{
		"namespace": {"required: namespace不能为空"},
		"content":   {"required: content不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

type KubePodDeleteRequest struct {
	Namespace    string `form:"namespace" json:"namespace" valid:"namespace"`
	Name         string `form:"name"      json:"name"      valid:"name"`
	GraceSeconds *int64 `form:"grace_seconds" json:"grace_seconds"`
	Force        bool   `form:"force" json:"force"`
}

func NewKubePodDeleteRequest() *KubePodDeleteRequest {
	return &KubePodDeleteRequest{}
}

func ValidKubePodDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
	}
	messages := govalidator.MapData{
		"namespace": {"required: namespace不能为空"},
		"name":      {"required: name不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 日志请求 DTO
// KubePodLogRequest 定义获取 Pod 日志的请求参数
type KubePodLogRequest struct {
	Name      string `json:"name"      form:"name"      valid:"name"`
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Container string `json:"container" form:"container" valid:"container"`
	Tail      *int64 `json:"tail"      form:"tail"      valid:"tail"`
	Follow    bool   `json:"follow"    form:"follow"    valid:"follow"`
}

func NewKubePodLogRequest() *KubePodLogRequest {
	return &KubePodLogRequest{}
}

func ValidKubePodLogRequest(obj interface{}, c *gin.Context) map[string][]string {
	req, ok := obj.(*KubePodLogRequest)
	if !ok {
		return map[string][]string{
			"request": {"对象类型断言失败"},
		}
	}

	rules := govalidator.MapData{
		"name":      []string{"required", "between:1,100"},
		"namespace": []string{"required", "between:1,100"},
		"container": []string{"between:0,100"},           // 可选，限制长度
		"tail":      []string{"numeric_between:0,10000"}, // 可选，数值范围
	}

	messages := govalidator.MapData{
		"name": []string{
			"required:Pod 名不能为空",
			"between:Pod 名长度必须在 1~100",
		},
		"namespace": []string{
			"required:命名空间不能为空",
			"between:命名空间长度必须在 1~100",
		},
		"container": []string{
			"between:容器名长度不能超过 100",
		},
		"tail": []string{
			"numeric_between:tail 范围必须在 0~10000",
		},
	}

	return valid.ValidateOptions(req, rules, messages)
}

// ================= Pod 标签变更 =================

func NewKubePodLabelPatchRequest() *KubePodLabelPatchRequest { return &KubePodLabelPatchRequest{} }

type KubePodLabelPatchRequest struct {
	Namespace string            `json:"namespace" valid:"namespace"`
	Name      string            `json:"name" valid:"name"`
	Add       map[string]string `json:"add"`    // 添加/更新的标签
	Remove    []string          `json:"remove"` // 要删除的标签 key
}

func ValidKubePodLabelPatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	})
}

// ================= Pod YAML 应用 =================

func NewKubePodApplyYamlRequest() *KubePodApplyYamlRequest { return &KubePodApplyYamlRequest{} }

type KubePodApplyYamlRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`
	Yaml      string `json:"yaml" valid:"yaml"`
}

func ValidKubePodApplyYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"yaml":      {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"yaml":      {"required: yaml 不能为空"},
	})
}

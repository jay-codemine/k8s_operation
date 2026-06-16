package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

//
// ================== DTO Struct 定义 ==================
//

// EnvVarKV 用于表示请求里的环境变量
type EnvVarKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewEnvVarKV() *EnvVarKV {
	return &EnvVarKV{}
}

// ================== Create 使用 ==================

func NewKubeAppConfigCreateRequest() *KubeAppConfigCreateRequest {
	return &KubeAppConfigCreateRequest{}
}

// 创建 AppConfig
type KubeAppConfigCreateRequest struct {
	ClusterID     uint32     `json:"cluster_id"      valid:"cluster_id"`
	Namespace     string     `json:"namespace"       valid:"namespace"`
	AppName       string     `json:"app_name"        valid:"app_name"`
	Image         string     `json:"image"           valid:"image"`
	Replicas      *int32     `json:"replicas"`                    // 可选
	Env           []EnvVarKV `json:"env"             valid:"env"` // 只校验 env 存在，不校验内部字段
	EnableMetrics bool       `json:"enable_metrics"`
	Strategy      string     `json:"strategy"`
}

// Create 校验
func ValidKubeAppConfigCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
		"image":      []string{"required"},
		"strategy":   []string{"omitempty", "in:RollingUpdate,Recreate"},
		"env":        []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
		"image":     {"required:image不能为空"},
		"strategy":  {"in:strategy必须是RollingUpdate或Recreate"},
		"env":       {"required:env不能为空（至少传 []）"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================== Update 使用 ==================

func NewKubeAppConfigUpdateRequest() *KubeAppConfigUpdateRequest {
	return &KubeAppConfigUpdateRequest{}
}

type KubeAppConfigUpdateRequest struct {
	ClusterID     uint32     `json:"cluster_id"      valid:"cluster_id"`
	Namespace     string     `json:"namespace"       valid:"namespace"`
	AppName       string     `json:"app_name"        valid:"app_name"`
	Image         string     `json:"image"`
	Replicas      *int32     `json:"replicas"`
	Env           []EnvVarKV `json:"env"` // 可选
	EnableMetrics bool       `json:"enable_metrics"`
	Strategy      string     `json:"strategy"`
}

// Update 校验
func ValidKubeAppConfigUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
		"strategy":   []string{"omitempty", "in:RollingUpdate,Recreate"},
		"env":        []string{"omitempty"}, // env 可选
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
		"strategy":  {"in:strategy必须是RollingUpdate或Recreate"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================== Detail 使用 ==================

func NewKubeAppConfigDetailRequest() *KubeAppConfigDetailRequest {
	return &KubeAppConfigDetailRequest{}
}

type KubeAppConfigDetailRequest struct {
	ClusterID uint32 `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Namespace string `json:"namespace"  form:"namespace"  valid:"namespace"`
	AppName   string `json:"app_name"   form:"app_name"   valid:"app_name"`
}

// Detail 校验
func ValidKubeAppConfigDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================== List 使用 ==================

func NewKubeAppConfigListRequest() *KubeAppConfigListRequest {
	return &KubeAppConfigListRequest{}
}

type KubeAppConfigListRequest struct {
	ClusterID uint32 `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Namespace string `json:"namespace"  form:"namespace"`
	Page      int    `json:"page"       form:"page"       valid:"page"`
	Limit     int    `json:"limit"      form:"limit"      valid:"limit"`
}

// List 校验
func ValidKubeAppConfigListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required", "numeric", "min:1"},
		"page":       []string{"numeric", "min:1"},
		"limit":      []string{"numeric", "min:1"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"page": {
			"numeric:page必须是数字",
			"min:page必须大于0",
		},
		"limit": {
			"numeric:limit必须是数字",
			"min:limit必须大于0",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================== Delete 使用 ==================

func NewKubeAppConfigDeleteRequest() *KubeAppConfigDeleteRequest {
	return &KubeAppConfigDeleteRequest{}
}

type KubeAppConfigDeleteRequest struct {
	ClusterID uint32 `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Namespace string `json:"namespace"  form:"namespace"  valid:"namespace"`
	AppName   string `json:"app_name"   form:"app_name"   valid:"app_name"`
}

// Delete 校验
func ValidKubeAppConfigDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_id": []string{"required"},
		"namespace":  []string{"required"},
		"app_name":   []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_id": {
			"required:cluster_id不能为空",
			"numeric:cluster_id必须是数字",
			"min:cluster_id必须大于0",
		},
		"namespace": {"required:namespace不能为空"},
		"app_name":  {"required:app_name不能为空"},
	}

	return valid.ValidateOptions(data, rules, messages)
}

package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

/* ====================================================================
 * HorizontalPodAutoscaler (HPA) 请求 DTO
 * ==================================================================== */

// ---------------------- HPA 列表 ----------------------

type KubeHPAListRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name" form:"name" valid:"name"`
	Page      int    `json:"page" form:"page" valid:"page"`
	Limit     int    `json:"limit" form:"limit" valid:"limit"`
}

func NewKubeHPAListRequest() *KubeHPAListRequest {
	return &KubeHPAListRequest{Page: 1, Limit: 20}
}

func ValidKubeHPAListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"max:64"},
		"name":      []string{"max:64"},
		"page":      []string{"min:1"},
		"limit":     []string{"min:1", "max:1000"},
	}
	msg := govalidator.MapData{
		"page":  []string{"min: page 必须 >= 1"},
		"limit": []string{"min: limit 必须 >= 1", "max: limit 不能超过 1000"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

// ---------------------- HPA 详情/删除 ----------------------

type KubeHPADetailRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name" form:"name" valid:"name"`
}

func NewKubeHPADetailRequest() *KubeHPADetailRequest {
	return &KubeHPADetailRequest{}
}

func ValidKubeHPADetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return ValidNameNamespace(data, nil)
}

// ---------------------- HPA 创建/更新 ----------------------

// KubeHPACreateRequest 创建/更新 HPA
type KubeHPACreateRequest struct {
	Namespace     string `json:"namespace" valid:"namespace"`
	Name          string `json:"name" valid:"name"`
	TargetKind    string `json:"target_kind" valid:"target_kind"`     // Deployment/StatefulSet
	TargetName    string `json:"target_name" valid:"target_name"`     // 目标负载名称
	TargetAPIVer  string `json:"target_api_version,omitempty"`        // 默认 apps/v1
	MinReplicas   int32  `json:"min_replicas" valid:"min_replicas"`   // 最小副本数 >=0
	MaxReplicas   int32  `json:"max_replicas" valid:"max_replicas"`   // 最大副本数 > min
	CPUTargetUtil *int32 `json:"cpu_target_util,omitempty"`           // 目标 CPU 利用率 % (1-100)
	MemTargetUtil *int32 `json:"mem_target_util,omitempty"`           // 目标内存利用率 %
	ScaleUpStab   *int32 `json:"scale_up_stab_seconds,omitempty"`     // 扩容稳定窗口（秒）
	ScaleDownStab *int32 `json:"scale_down_stab_seconds,omitempty"`   // 缩容稳定窗口（秒）

	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func NewKubeHPACreateRequest() *KubeHPACreateRequest {
	return &KubeHPACreateRequest{}
}

func ValidKubeHPACreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":    []string{"required"},
		"name":         []string{"required"},
		"target_kind":  []string{"required", "in:Deployment,StatefulSet,ReplicaSet,ReplicationController"},
		"target_name":  []string{"required"},
		"min_replicas": []string{"required", "numeric", "min:0"},
		"max_replicas": []string{"required", "numeric", "min:1"},
	}
	msg := govalidator.MapData{
		"namespace":    []string{"required: namespace 不能为空"},
		"name":         []string{"required: name 不能为空"},
		"target_kind":  []string{"required: target_kind 不能为空", "in: target_kind 必须是 Deployment/StatefulSet/ReplicaSet/ReplicationController"},
		"target_name":  []string{"required: target_name 不能为空"},
		"min_replicas": []string{"required: min_replicas 不能为空", "numeric: min_replicas 必须是数字", "min: min_replicas 必须 >= 0"},
		"max_replicas": []string{"required: max_replicas 不能为空", "numeric: max_replicas 必须是数字", "min: max_replicas 必须 >= 1"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

// KubeHPAUpdateRequest 更新 HPA（结构同 Create）
type KubeHPAUpdateRequest = KubeHPACreateRequest

func NewKubeHPAUpdateRequest() *KubeHPAUpdateRequest {
	return &KubeHPAUpdateRequest{}
}

func ValidKubeHPAUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return ValidKubeHPACreateRequest(data, ctx)
}

// ---------------------- HPA 单独修改副本数 ----------------------

// KubeHPAScaleRequest 单独修改副本数（min/max）
type KubeHPAScaleRequest struct {
	Namespace   string `json:"namespace" valid:"namespace"`
	Name        string `json:"name" valid:"name"`
	MinReplicas int32  `json:"min_replicas" valid:"min_replicas"` // 最小副本数
	MaxReplicas int32  `json:"max_replicas" valid:"max_replicas"` // 最大副本数
}

func NewKubeHPAScaleRequest() *KubeHPAScaleRequest {
	return &KubeHPAScaleRequest{}
}

func ValidKubeHPAScaleRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":    []string{"required"},
		"name":         []string{"required"},
		"min_replicas": []string{"required", "numeric", "min:0"},
		"max_replicas": []string{"required", "numeric", "min:1"},
	}
	msg := govalidator.MapData{
		"namespace":    []string{"required: namespace 不能为空"},
		"name":         []string{"required: name 不能为空"},
		"min_replicas": []string{"required: min_replicas 不能为空", "numeric: 必须是数字", "min: 必须 >= 0"},
		"max_replicas": []string{"required: max_replicas 不能为空", "numeric: 必须是数字", "min: 必须 >= 1"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

// ---------------------- HPA 批量扩容/缩容（618 促销场景） ----------------------

// KubeHPABatchItem 批量项
type KubeHPABatchItem struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	MinReplicas int32  `json:"min_replicas"`
	MaxReplicas int32  `json:"max_replicas"`
}

// KubeHPABatchScaleRequest 批量扩缩容请求
type KubeHPABatchScaleRequest struct {
	Items []KubeHPABatchItem `json:"items" valid:"items"`
}

func NewKubeHPABatchScaleRequest() *KubeHPABatchScaleRequest {
	return &KubeHPABatchScaleRequest{}
}

func ValidKubeHPABatchScaleRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"items": []string{"required"},
	}
	msg := govalidator.MapData{
		"items": []string{"required: items 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

/* ====================================================================
 * VerticalPodAutoscaler (VPA) 请求 DTO
 * ==================================================================== */

// ---------------------- VPA 列表 ----------------------

type KubeVPAListRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name" form:"name" valid:"name"`
	Page      int    `json:"page" form:"page" valid:"page"`
	Limit     int    `json:"limit" form:"limit" valid:"limit"`
}

func NewKubeVPAListRequest() *KubeVPAListRequest {
	return &KubeVPAListRequest{Page: 1, Limit: 20}
}

func ValidKubeVPAListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"max:64"},
		"name":      []string{"max:64"},
		"page":      []string{"min:1"},
		"limit":     []string{"min:1", "max:1000"},
	}
	msg := govalidator.MapData{
		"page":  []string{"min: page 必须 >= 1"},
		"limit": []string{"min: limit 必须 >= 1"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

// ---------------------- VPA 详情/删除 ----------------------

type KubeVPADetailRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name" form:"name" valid:"name"`
}

func NewKubeVPADetailRequest() *KubeVPADetailRequest {
	return &KubeVPADetailRequest{}
}

func ValidKubeVPADetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return ValidNameNamespace(data, nil)
}

// ---------------------- VPA 创建/更新 ----------------------

type KubeVPACreateRequest struct {
	Namespace     string   `json:"namespace" valid:"namespace"`
	Name          string   `json:"name" valid:"name"`
	TargetKind    string   `json:"target_kind" valid:"target_kind"`         // Deployment/StatefulSet
	TargetName    string   `json:"target_name" valid:"target_name"`
	TargetAPIVer  string   `json:"target_api_version,omitempty"`            // 默认 apps/v1
	UpdateMode    string   `json:"update_mode" valid:"update_mode"`         // Off/Initial/Recreate/Auto
	ContainerName string   `json:"container_name,omitempty"`                // 默认 *
	ControlledRes []string `json:"controlled_resources,omitempty"`          // 默认 ["cpu","memory"]
	MinAllowedCPU string   `json:"min_allowed_cpu,omitempty"`               // 例如 "100m"
	MinAllowedMem string   `json:"min_allowed_mem,omitempty"`               // 例如 "128Mi"
	MaxAllowedCPU string   `json:"max_allowed_cpu,omitempty"`
	MaxAllowedMem string   `json:"max_allowed_mem,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`
}

func NewKubeVPACreateRequest() *KubeVPACreateRequest {
	return &KubeVPACreateRequest{}
}

func ValidKubeVPACreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":   []string{"required"},
		"name":        []string{"required"},
		"target_kind": []string{"required", "in:Deployment,StatefulSet,DaemonSet,ReplicaSet"},
		"target_name": []string{"required"},
		"update_mode": []string{"in:Off,Initial,Recreate,Auto"},
	}
	msg := govalidator.MapData{
		"namespace":   []string{"required: namespace 不能为空"},
		"name":        []string{"required: name 不能为空"},
		"target_kind": []string{"required: target_kind 不能为空", "in: target_kind 不合法"},
		"target_name": []string{"required: target_name 不能为空"},
		"update_mode": []string{"in: update_mode 必须为 Off/Initial/Recreate/Auto"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

type KubeVPAUpdateRequest = KubeVPACreateRequest

func NewKubeVPAUpdateRequest() *KubeVPAUpdateRequest {
	return &KubeVPAUpdateRequest{}
}

func ValidKubeVPAUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return ValidKubeVPACreateRequest(data, ctx)
}

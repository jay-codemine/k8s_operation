package requests

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"

	corev1 "k8s.io/api/core/v1"

	"k8soperation/pkg/valid"
)

/* ---------------------- DaemonSet 创建 ---------------------- */

func NewKubeDaemonSetCreateRequest() *KubeDaemonSetCreateRequest {
	return &KubeDaemonSetCreateRequest{}
}

// KubeDaemonSetCreateRequest 定义创建 DaemonSet 的请求结构
// 注意：DaemonSet 没有 replicas。Pod 数量=符合条件的节点数（由节点选择/污点容忍等决定）
type KubeDaemonSetCreateRequest struct {
	// 基本
	Name        string                `json:"name" valid:"name"`               // DaemonSet 名称
	Namespace   string                `json:"namespace" valid:"namespace"`     // 命名空间
	Description *string               `json:"description" valid:"description"` // 描述（可选）
	Labels      []Label               `json:"labels" valid:"labels"`           // 资源标签（会与系统关键标签合并）
	Variables   []EnvironmentVariable `json:"variables" valid:"variables"`     // 环境变量

	// 容器
	ContainerImage       string                        `json:"container_image" valid:"container_image"`                 // 容器镜像
	ImagePullSecret      *string                       `json:"image_pull_secret" valid:"image_pull_secret"`             // 单个镜像拉取密钥名（可选，兼容你原有风格）
	ImagePullSecrets     []corev1.LocalObjectReference `json:"image_pull_secrets,omitempty" valid:"image_pull_secrets"` // 多个镜像密钥
	ContainerCommand     *string                       `json:"container_command" valid:"container_command"`             // 启动命令（可选）
	ContainerCommandArgs *string                       `json:"container_command_args" valid:"container_command_args"`   // 启动参数（可选）
	PortMappings         []PortMapping                 `json:"port_mappings" valid:"port_mappings"`                     // 端口映射
	RunAsPrivileged      bool                          `json:"run_as_privileged" valid:"run_as_privileged"`             // 是否特权
	MemoryRequirement    *string                       `json:"memory_requirement" valid:"memory_requirement"`           // 请求内存（可选）
	CpuRequirement       *string                       `json:"cpu_requirement" valid:"cpu_requirement"`                 // 请求 CPU（可选）

	// 探针
	IsReadinessEnable bool              `json:"is_readiness_enable" valid:"is_readiness_enable"`
	ReadinessProbe    HealthCheckDetail `json:"readiness_probe" valid:"readiness_probe"`
	IsLivenessEnable  bool              `json:"is_liveness_enable" valid:"is_liveness_enable"`
	LivenessProbe     HealthCheckDetail `json:"liveness_probe" valid:"liveness_probe"`

	// DaemonSet 常用：节点/污点/亲和/网络/SA/卷
	NodeSelector   map[string]string   `json:"node_selector,omitempty" valid:"node_selector"`
	Tolerations    []corev1.Toleration `json:"tolerations,omitempty" valid:"tolerations"`
	Affinity       *corev1.Affinity    `json:"affinity,omitempty" valid:"affinity"`
	ServiceAccount string              `json:"service_account,omitempty" valid:"service_account"`
	HostNetwork    bool                `json:"host_network,omitempty" valid:"host_network"`
	Volumes        []corev1.Volume     `json:"volumes,omitempty" valid:"volumes"`

	// 可选：是否同时创建 Service（如需要暴露端口）
	IsCreateService      bool    `json:"is_create_service" valid:"is_create_service"`
	ServiceType          string  `json:"service_type" valid:"service_type"`
	ServiceName          string  `json:"service_name" valid:"service_name"`
	MaxUnavailable       *string `json:"max_unavailable" valid:"max_unavailable"`
	RevisionHistoryLimit *int32  `json:"revision_history_limit" valid:"revision_history_limit"`
	MinReadySeconds      int32   `json:"min_ready_seconds" valid:"min_ready_seconds"`
}

func ValidKubeDaemonSetCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":            []string{"required"},
		"namespace":       []string{"required"},
		"container_image": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name 不能为空"},
		"namespace": []string{"required: namespace 不能为空"},
		"container_image": []string{
			"required: image 不能为空",
		},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 更新（原样 JSON/YAML） ---------------------- */

func NewKubeDaemonSetUpdateRequest() *KubeDaemonSetUpdateRequest {
	return &KubeDaemonSetUpdateRequest{}
}

type KubeDaemonSetUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"` // 命名空间
	Content   json.RawMessage `json:"content" valid:"content"`     // 更新内容（一般是 YAML/JSON）
	Name      string          `json:"name" valid:"name"`           // DaemonSet 名称
}

func ValidKubeDaemonSetUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"content":   []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"content":   []string{"required: content 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 列表 ---------------------- */

func NewKubeDaemonSetListRequest() *KubeDaemonSetListRequest {
	return &KubeDaemonSetListRequest{}
}

type KubeDaemonSetListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`   // 页码
	Limit int `json:"limit" valid:"limit"` // 每页条数
}

func ValidKubeDaemonSetListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		// name：非必填；如果传了，长度 1~64
		"name": []string{"min:1", "max:64"},

		// namespace：非必填；如果传了，长度 1~64
		"namespace": []string{"min:1", "max:64"},

		// page / limit：非必填；如果传了必须合法
		"page":  []string{"min:1"},
		"limit": []string{"min:1", "max:1000"},
	}

	messages := govalidator.MapData{
		"name": {
			"min: name长度不能小于1",
			"max: name长度不能超过64",
		},
		"namespace": {
			"min: namespace长度不能小于1",
			"max: namespace长度不能超过64",
		},
		"page": {
			"min: page必须>=1",
		},
		"limit": {
			"min: limit必须>=1",
			"max: limit不能超过200",
		},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 重启（rollout restart） ---------------------- */

func NewKubeDaemonSetRestartRequest() *KubeDaemonSetRestartRequest {
	return &KubeDaemonSetRestartRequest{}
}

type KubeDaemonSetRestartRequest struct {
	KubeCommonRequest
}

func ValidKubeDaemonSetRestartRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 详情 ---------------------- */

func NewKubeDaemonSetDetailRequest() *KubeDaemonSetDetailRequest {
	return &KubeDaemonSetDetailRequest{}
}

type KubeDaemonSetDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeDaemonSetDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 删除 ---------------------- */

func NewKubeDaemonSetDeleteRequest() *KubeDaemonSetDeleteRequest {
	return &KubeDaemonSetDeleteRequest{}
}

type KubeDaemonSetDeleteRequest struct {
	KubeCommonRequest
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"` // 优雅终止时间（秒）
	Force              bool   `json:"force,omitempty" valid:"force"`                               // 是否强制删除（一般不建议对 DS 使用 --force --grace-period=0）
	Wait               bool   `json:"wait,omitempty" valid:"wait"`                                 // 是否等待删除完成
	TimeoutSeconds     *int64 `json:"timeout_seconds,omitempty" valid:"timeout_seconds"`           // 等待超时（秒）
}

func ValidKubeDaemonSetDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name 不能为空"},
		"namespace": []string{"required: namespace 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet 镜像更新 ---------------------- */

func NewKubeDaemonSetUpdateImageRequest() *KubeDaemonSetUpdateImageRequest {
	return &KubeDaemonSetUpdateImageRequest{}
}

type KubeDaemonSetUpdateImageRequest struct {
	KubeCommonRequest
	Container string `json:"container" valid:"container"` // 目标容器名称
	Image     string `json:"image" valid:"image"`         // 新镜像地址，例如 nginx:1.27
}

func ValidKubeDaemonSetUpdateImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"container": []string{"required"},
		"image":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"container": []string{"required: container 不能为空"},
		"image":     []string{"required: image 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet Svc（如需要） ---------------------- */

func NewKubeDaemonSetCreateSvcRequest() *KubeDaemonSetCreateSvcRequest {
	return &KubeDaemonSetCreateSvcRequest{}
}

type KubeDaemonSetCreateSvcRequest struct {
	KubeCommonRequest
}

func ValidKubeDaemonSetCreateSvcRequest(data interface{}, ctx context.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

type KubeDaemonSetRollbackRequest struct {
	KubeCommonRequest
	// 指定回滚到的历史版本（ControllerRevision）
	// 建议用 revision_name（更直观），如果你喜欢数字 revision 也可以再加一个 int64 字段做二选一
	RevisionName string `json:"revision_name" valid:"revision_name"`
	// 可选：若你想支持 “不传则回滚到上一个版本”
	// 若启用该策略，可将 valid 规则从 required 改为 optional，并在服务层做默认选取
}

// DaemonSet 回滚请求
func NewKubeDaemonSetRollbackRequest() *KubeDaemonSetRollbackRequest {
	return &KubeDaemonSetRollbackRequest{}
}

func ValidKubeDaemonSetRollbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":     []string{"required"},
		"name":          []string{"required"},
		"revision_name": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace":     []string{"required: namespace 不能为空"},
		"name":          []string{"required: name 不能为空"},
		"revision_name": []string{"required: revision_name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet History（版本历史） ---------------------- */

func NewKubeDaemonSetHistoryRequest() *KubeDaemonSetHistoryRequest {
	return &KubeDaemonSetHistoryRequest{}
}

type KubeDaemonSetHistoryRequest struct {
	KubeCommonRequest
}

func ValidKubeDaemonSetHistoryRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

/* ---------------------- DaemonSet Pods（关联 Pod 列表） ---------------------- */

func NewKubeDaemonSetPodsRequest() *KubeDaemonSetPodsRequest {
	return &KubeDaemonSetPodsRequest{}
}

type KubeDaemonSetPodsRequest struct {
	KubeCommonRequest
}

func ValidKubeDaemonSetPodsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

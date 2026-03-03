package requests

import (
	"encoding/json"
	"k8soperation/pkg/valid"

	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"

	corev1 "k8s.io/api/core/v1"
)

// --------- Job 创建 ---------

func NewKubeJobCreateRequest() *KubeJobCreateRequest { return &KubeJobCreateRequest{} }

// 说明：Job 没有 replicas；控制并发/完成数通过 Parallelism/Completions
type KubeJobCreateRequest struct {
	// 基本
	Name        string  `json:"name" valid:"name"`               // Job 名称
	Namespace   string  `json:"namespace" valid:"namespace"`     // 命名空间
	Description *string `json:"description" valid:"description"` // 描述（可选）
	Labels      []Label `json:"labels" valid:"labels"`           // 资源标签（会与系统关键标签合并）

	// 容器
	ContainerName        string                `json:"container_name"`                          // 可空；为空则用 name
	ContainerImage       string                `json:"container_image" valid:"container_image"` // 镜像（必填）
	ContainerCommand     *string               `json:"container_command" valid:"container_command"`
	ContainerCommandArgs *string               `json:"container_command_args" valid:"container_command_args"`
	Variables            []EnvironmentVariable `json:"variables" valid:"variables"`
	PortMappings         []PortMapping         `json:"port_mappings" valid:"port_mappings"`
	RunAsPrivileged      *bool                 `json:"run_as_privileged" valid:"run_as_privileged"`
	MemoryRequirement    *string               `json:"memory_requirement" valid:"memory_requirement"`
	CpuRequirement       *string               `json:"cpu_requirement" valid:"cpu_requirement"`

	// PodSpec 相关
	RestartPolicy    string                        `json:"restart_policy,omitempty" valid:"restart_policy"` // "OnFailure"/"Never"（默认 OnFailure）
	ImagePullSecrets []corev1.LocalObjectReference `json:"image_pull_secrets,omitempty" valid:"image_pull_secrets"`
	ServiceAccount   string                        `json:"service_account,omitempty" valid:"service_account"`
	NodeSelector     map[string]string             `json:"node_selector,omitempty" valid:"node_selector"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty" valid:"tolerations"`
	Affinity         *corev1.Affinity              `json:"affinity,omitempty" valid:"affinity"`
	Volumes          []corev1.Volume               `json:"volumes,omitempty" valid:"volumes"`
	InitContainers   []corev1.Container            `json:"init_containers,omitempty" valid:"init_containers"`

	// 探针（Job 也可配置，长跑/有健康需求时有用）
	IsReadinessEnable bool              `json:"is_readiness_enable" valid:"is_readiness_enable"`
	ReadinessProbe    HealthCheckDetail `json:"readiness_probe" valid:"readiness_probe"`
	IsLivenessEnable  bool              `json:"is_liveness_enable" valid:"is_liveness_enable"`
	LivenessProbe     HealthCheckDetail `json:"liveness_probe" valid:"liveness_probe"`

	// JobSpec
	Parallelism             *int32 `json:"parallelism,omitempty" valid:"parallelism"`                         // 并发度
	Completions             *int32 `json:"completions,omitempty" valid:"completions"`                         // 需完成数
	BackoffLimit            *int32 `json:"backoff_limit,omitempty" valid:"backoff_limit"`                     // 失败重试次数
	ActiveDeadlineSeconds   *int64 `json:"active_deadline_seconds,omitempty" valid:"active_deadline_seconds"` // 最长期限
	TTLSecondsAfterFinished *int32 `json:"ttl_seconds_after_finished,omitempty" valid:"ttl_seconds_after_finished"`
	Suspend                 *bool  `json:"suspend,omitempty" valid:"suspend"` // 创建即挂起（可选）

	// 若你想显式设置 selector（一般不建议，留空即可）
	SetExplicitSelector bool        `json:"set_explicit_selector,omitempty" valid:"set_explicit_selector"`
	PodLabels           interface{} `json:"pod_labels,omitempty" valid:"pod_labels"`
	Image               string      `json:"image,omitempty" valid:"image"`
}

func ValidKubeJobCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":            []string{"required"},
		"namespace":       []string{"required"},
		"container_image": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":            []string{"required: name 不能为空"},
		"namespace":       []string{"required: namespace 不能为空"},
		"container_image": []string{"required: image 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- Job 更新（原样 JSON/YAML） ---------

func NewKubeJobUpdateRequest() *KubeJobUpdateRequest { return &KubeJobUpdateRequest{} }

type KubeJobUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"`
	Name      string          `json:"name" valid:"name"`
	Content   json.RawMessage `json:"content" valid:"content"`
}

func ValidKubeJobUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"content":   []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"content":   []string{"required: content 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- Job 列表 ---------

func NewKubeJobListRequest() *KubeJobListRequest { return &KubeJobListRequest{} }

type KubeJobListRequest struct {
	KubeCommonRequest
	Page          int    `json:"page" valid:"page"`
	Limit         int    `json:"limit" valid:"limit"`
	CronJob       string `json:"cronjob" valid:"cronjob"`               // 按 CronJob 名称筛选
	LabelSelector string `json:"label_selector" valid:"label_selector"` // 标签选择器 (k8s label selector 格式)
}

func ValidKubeJobListRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// --------- Job 详情 ---------

func NewKubeJobDetailRequest() *KubeJobDetailRequest { return &KubeJobDetailRequest{} }

type KubeJobDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeJobDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- Job 删除 ---------

func NewKubeJobDeleteRequest() *KubeJobDeleteRequest { return &KubeJobDeleteRequest{} }

type KubeJobDeleteRequest struct {
	KubeCommonRequest
	// 删除策略：通常 Job 用 Background/Foreground；是否保留已完成的 Pod/Logs 由 PropagationPolicy 决定
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"`
	Force              bool   `json:"force,omitempty" valid:"force"`
	Wait               bool   `json:"wait,omitempty" valid:"wait"`
	TimeoutSeconds     *int64 `json:"timeout_seconds,omitempty" valid:"timeout_seconds"`
	// 可选：是否连同 Pods 一起删（如果想要“清空后重跑”，可以先删 Pods 再重建）
	Cascade *bool `json:"cascade,omitempty" valid:"cascade"`
}

func ValidKubeJobDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- Job 挂起/恢复（spec.suspend=true/false） ---------

func NewKubeJobSuspendRequest() *KubeJobSuspendRequest { return &KubeJobSuspendRequest{} }

type KubeJobSuspendRequest struct {
	KubeCommonRequest
	Suspend bool `json:"suspend" valid:"suspend"` // true=挂起，false=恢复
}

func ValidKubeJobSuspendRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- Job 更新镜像（可选：修改 PodTemplate 中容器镜像） ---------
// 注意：Job 不是 Deployment，不存在“滚动升级”；修改模板后仅影响**之后新创建**的 Pod。

func NewKubeJobUpdateImageRequest() *KubeJobUpdateImageRequest { return &KubeJobUpdateImageRequest{} }

type KubeJobUpdateImageRequest struct {
	KubeCommonRequest
	Container string `json:"container" valid:"container"`
	Image     string `json:"image" valid:"image"`
	// 可选：是否打一个 annotation 强制模板变化（方便触发 controller 识别到变更）
	Touch bool `json:"touch,omitempty" valid:"touch"`
}

func ValidKubeJobUpdateImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
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
	return ValidateOptions(data, rules, messages)
}

// NewKubeJobRestartRequest 创建 Job 重启请求结构体实例
func NewKubeJobRestartRequest() *KubeJobRestartRequest {
	return &KubeJobRestartRequest{}
}

// KubeJobRestartRequest 定义重启 Job 的请求结构
// 用于 /api/v1/k8s/job/restart 接口
type KubeJobRestartRequest struct {
	KubeCommonRequest
}

// ValidKubeJobRestartRequest 参数校验规则
func ValidKubeJobRestartRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return ValidateOptions(data, rules, messages)
}

// --------- 公共校验入口别名（与现有保持一致） ---------

// 你已有的 valid.ValidateOptions 别名，避免到处 import 包名不同
func ValidateOptions(data interface{}, rules, messages map[string][]string) map[string][]string {
	return valid.ValidateOptions(data, rules, messages)
}

/* 说明与对齐点
1) 字段命名、校验函数命名、New*Request 工厂函数，均与 DaemonSet/Deployment 版保持一致。
2) Job 不支持“回滚/历史版本”（没有 ControllerRevision 的复用语义），因此不提供 Rollback DTO。
3) Job 的运维常见动作是：创建、暂停/恢复、删除、（必要时）修改模板或重跑。
4) 若你希望支持“重跑 Job”：通常是删除 Job（或删除其 Pods）后重新创建；也可把 .spec.suspend=true → false 触发调度。
*/

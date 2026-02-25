package requests

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/valid"
)

/* =========================
   CronJob 创建
   ========================= */

type KubeCronJobCreateRequest struct {
	// === 基本信息 ===
	Name        string  `json:"name" valid:"name"`                         // CronJob 名称
	Namespace   string  `json:"namespace" valid:"namespace"`               // 命名空间
	Description *string `json:"description,omitempty" valid:"description"` // 描述，可选
	Labels      []Label `json:"labels,omitempty" valid:"labels"`           // 标签，可选

	// === 调度配置 ===
	Schedule                   string `json:"schedule" valid:"schedule"`                               // Cron 表达式（分 时 日 月 周）
	TimeZone                   string `json:"time_zone,omitempty" valid:"time_zone"`                   // 时区
	Suspend                    *bool  `json:"suspend,omitempty" valid:"suspend"`                       // 是否暂停任务
	ConcurrencyPolicy          string `json:"concurrency_policy,omitempty" valid:"concurrency_policy"` // Allow|Forbid|Replace
	StartingDeadlineSeconds    *int64 `json:"starting_deadline_seconds,omitempty" valid:"starting_deadline_seconds"`
	SuccessfulJobsHistoryLimit *int32 `json:"successful_jobs_history_limit,omitempty" valid:"successful_jobs_history_limit"`
	FailedJobsHistoryLimit     *int32 `json:"failed_jobs_history_limit,omitempty" valid:"failed_jobs_history_limit"`

	// === JobSpec 级字段 ===
	Parallelism             *int32 `json:"parallelism,omitempty" valid:"parallelism"`
	Completions             *int32 `json:"completions,omitempty" valid:"completions"`
	BackoffLimit            *int32 `json:"backoff_limit,omitempty" valid:"backoff_limit"`
	ActiveDeadlineSeconds   *int64 `json:"active_deadline_seconds,omitempty" valid:"active_deadline_seconds"`
	TTLSecondsAfterFinished *int32 `json:"ttl_seconds_after_finished,omitempty" valid:"ttl_seconds_after_finished"`

	// === Pod 模板 ===
	ContainerName        string                `json:"container_name,omitempty" valid:"container_name"`
	ContainerImage       string                `json:"container_image" valid:"container_image"`
	ContainerCommand     []string              `json:"container_command,omitempty" valid:"container_command"`
	ContainerCommandArgs []string              `json:"container_command_args,omitempty" valid:"container_command_args"`
	Variables            []EnvironmentVariable `json:"variables,omitempty" valid:"variables"`
	PortMappings         []PortMapping         `json:"port_mappings,omitempty" valid:"port_mappings"`
	RunAsPrivileged      *bool                 `json:"run_as_privileged,omitempty" valid:"run_as_privileged"`
	MemoryRequirement    *string               `json:"memory_requirement,omitempty" valid:"memory_requirement"`
	CpuRequirement       *string               `json:"cpu_requirement,omitempty" valid:"cpu_requirement"`

	// === PodSpec 配置 ===
	RestartPolicy    string                        `json:"restart_policy,omitempty" valid:"restart_policy"` // OnFailure / Never
	ImagePullSecrets []corev1.LocalObjectReference `json:"image_pull_secrets,omitempty" valid:"image_pull_secrets"`
	ServiceAccount   string                        `json:"service_account,omitempty" valid:"service_account"`
	NodeSelector     map[string]string             `json:"node_selector,omitempty" valid:"node_selector"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty" valid:"tolerations"`
	Affinity         *corev1.Affinity              `json:"affinity,omitempty" valid:"affinity"`
	Volumes          []corev1.Volume               `json:"volumes,omitempty" valid:"volumes"`
	InitContainers   []corev1.Container            `json:"init_containers,omitempty" valid:"init_containers"`
	Containers       []corev1.Container            `json:"containers,omitempty" valid:"containers"`

	// === 健康探针 ===
	IsReadinessEnable bool              `json:"is_readiness_enable,omitempty" valid:"is_readiness_enable"`
	ReadinessProbe    HealthCheckDetail `json:"readiness_probe,omitempty" valid:"readiness_probe"`
	IsLivenessEnable  bool              `json:"is_liveness_enable,omitempty" valid:"is_liveness_enable"`
	LivenessProbe     HealthCheckDetail `json:"liveness_probe,omitempty" valid:"liveness_probe"`
}

func NewKubeCronJobCreateRequest() *KubeCronJobCreateRequest {
	return &KubeCronJobCreateRequest{}
}

func ValidKubeCronJobCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		// 必填字段
		"name":            {"required"},
		"namespace":       {"required"},
		"schedule":        {"required"},
		"container_image": {"required"},

		// 合法值检查
		"restart_policy":     {"in:Never,OnFailure"},
		"concurrency_policy": {"in:Allow,Forbid,Replace"},

		// 数值范围
		"starting_deadline_seconds":     {"numeric", "min:0"},
		"successful_jobs_history_limit": {"numeric", "min:0"},
		"failed_jobs_history_limit":     {"numeric", "min:0"},

		// 时区格式
		"time_zone": {"ascii"},
	}

	messages := govalidator.MapData{
		"name":                          {"required: name 不能为空"},
		"namespace":                     {"required: namespace 不能为空"},
		"schedule":                      {"required: schedule 不能为空"},
		"container_image":               {"required: container_image 不能为空"},
		"restart_policy":                {"in: restart_policy 仅支持 Never/OnFailure"},
		"concurrency_policy":            {"in: concurrency_policy 仅支持 Allow/Forbid/Replace"},
		"starting_deadline_seconds":     {"numeric: 必须为数字", "min: 不能小于 0"},
		"successful_jobs_history_limit": {"numeric: 必须为数字", "min: 不能小于 0"},
		"failed_jobs_history_limit":     {"numeric: 必须为数字", "min: 不能小于 0"},
		"time_zone":                     {"ascii: time_zone 必须为 ASCII（示例 Asia/Shanghai）"},
	}

	return ValidateOptions(data, rules, messages)
}

/* =========================
   CronJob 更新（原样 JSON/YAML）
   ========================= */

func NewKubeCronJobUpdateRequest() *KubeCronJobUpdateRequest { return &KubeCronJobUpdateRequest{} }

type KubeCronJobUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"`
	Name      string          `json:"name" valid:"name"`
	Content   json.RawMessage `json:"content" valid:"content"`
}

func ValidKubeCronJobUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* =========================
   CronJob 列表
   ========================= */

func NewKubeCronJobListRequest() *KubeCronJobListRequest { return &KubeCronJobListRequest{} }

type KubeCronJobListRequest struct {
	KubeCommonRequest
	Page          int    `json:"page" valid:"page"`
	Limit         int    `json:"limit" valid:"limit"`
	LabelSelector string `json:"label_selector" valid:"label_selector"` // 标签选择器 (k8s label selector 格式)
}

func ValidKubeCronJobListRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* =========================
   CronJob 详情
   ========================= */

func NewKubeCronJobDetailRequest() *KubeCronJobDetailRequest { return &KubeCronJobDetailRequest{} }

type KubeCronJobDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeCronJobDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* =========================
   CronJob 删除
   ========================= */

func NewKubeCronJobDeleteRequest() *KubeCronJobDeleteRequest { return &KubeCronJobDeleteRequest{} }

type KubeCronJobDeleteRequest struct {
	KubeCommonRequest
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"`
	Force              bool   `json:"force,omitempty" valid:"force"`
	Wait               bool   `json:"wait,omitempty" valid:"wait"`
	TimeoutSeconds     *int64 `json:"timeout_seconds,omitempty" valid:"timeout_seconds"`
	// 是否级联清理：通常删除 CronJob 时是否一并清理其历史 Job
	Cascade *bool `json:"cascade,omitempty" valid:"cascade"`
}

func ValidKubeCronJobDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* =========================
   CronJob 挂起/恢复（spec.suspend=true/false）
   ========================= */

func NewKubeCronJobSuspendRequest() *KubeCronJobSuspendRequest { return &KubeCronJobSuspendRequest{} }

type KubeCronJobSuspendRequest struct {
	KubeCommonRequest
	Suspend bool `json:"suspend" valid:"suspend"` // true=暂停，false=恢复
}

func ValidKubeCronJobSuspendRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* =========================
   CronJob 手动触发（立即创建一个 Job）
   ========================= */

func NewKubeCronJobTriggerRequest() *KubeCronJobTriggerRequest { return &KubeCronJobTriggerRequest{} }

type KubeCronJobTriggerRequest struct {
	KubeCommonRequest
}

func ValidKubeCronJobTriggerRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

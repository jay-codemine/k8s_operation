package requests

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/valid"
)

// ---------------------- Deployment 创建 ----------------------

func NewKubeDeploymentCreateRequest() *KubeDeploymentCreateRequest {
	return &KubeDeploymentCreateRequest{}
}

// SchedulingPolicy 调度策略类型
// default: 默认规则（不设置亲和性）
// spread: 分散调度（Pod 反亲和性，尽量分散到不同节点）
// pack: 集中调度（Pod 亲和性，尽量调度到同一节点）
// custom: 自定义规则（使用 NodeSelector/Affinity/Tolerations）
type SchedulingPolicy string

const (
	SchedulingPolicyDefault SchedulingPolicy = "default"
	SchedulingPolicySpread  SchedulingPolicy = "spread"
	SchedulingPolicyPack    SchedulingPolicy = "pack"
	SchedulingPolicyCustom  SchedulingPolicy = "custom"
)

// NodeAffinityRule 简化的节点亲和性规则（前端友好）
type NodeAffinityRule struct {
	Key      string `json:"key"`                // 节点标签键
	Operator string `json:"operator"`           // 操作符: In, NotIn, Exists, DoesNotExist, Gt, Lt
	Values   string `json:"values,omitempty"`   // 值（逗号分隔）
	Required bool   `json:"required,omitempty"` // true=硬性要求, false=软性偏好
	Weight   int32  `json:"weight,omitempty"`   // 软性偏好权重 (1-100)
}

// TopologySpread 简化的拓扑分布约束（前端友好）
type TopologySpread struct {
	TopologyKey       string `json:"topology_key"`                 // 拓扑域键，如 kubernetes.io/hostname, topology.kubernetes.io/zone
	MaxSkew           int32  `json:"max_skew"`                     // 最大偏差值
	WhenUnsatisfiable string `json:"when_unsatisfiable,omitempty"` // DoNotSchedule 或 ScheduleAnyway
	LabelSelector     string `json:"label_selector,omitempty"`     // 标签选择器 (key=value)
}

// KubeDeploymentCreateRequest 定义创建 Deployment 的请求结构
type KubeDeploymentCreateRequest struct {
	Name                 string                `json:"name" valid:"name"`                                     // Deployment 名称
	ContainerImage       string                `json:"container_image" valid:"container_image"`               // 容器镜像
	ImagePullSecret      *string               `json:"image_pull_secret" valid:"image_pull_secret"`           // 镜像拉取密钥（可选）
	ContainerCommand     *string               `json:"container_command" valid:"container_command"`           // 容器启动命令（可选）
	ContainerCommandArgs *string               `json:"container_command_args" valid:"container_command_args"` // 容器启动参数（可选）
	Replicas             int32                 `json:"replicas" valid:"replicas"`                             // 副本数
	PortMappings         []PortMapping         `json:"port_mappings" valid:"port_mappings"`                   // 容器端口映射
	Variables            []EnvironmentVariable `json:"variables" valid:"variables"`                           // 环境变量
	IsCreateService      bool                  `json:"is_create_service" valid:"is_create_service"`           // 是否同时创建 Service
	Description          *string               `json:"description" valid:"description"`                       // 描述信息（可选）
	Namespace            string                `json:"namespace" valid:"namespace"`                           // 命名空间
	MemoryRequirement    *string               `json:"memory_requirement" valid:"memory_requirement"`         // 内存需求（可选）
	CpuRequirement       *string               `json:"cpu_requirement" valid:"cpu_requirement"`               // CPU 需求（可选）
	Labels               []Label               `json:"labels" valid:"labels"`                                 // 标签
	RunAsPrivileged      bool                  `json:"run_as_privileged" valid:"run_as_privileged"`           // 是否以特权模式运行
	IsReadinessEnable    bool                  `json:"is_readiness_enable" valid:"is_readiness_enable"`       // 是否启用 Readiness 探针
	ReadinessProbe       HealthCheckDetail     `json:"readiness_probe" valid:"readiness_probe"`               // Readiness 探针配置
	IsLivenessEnable     bool                  `json:"is_liveness_enable" valid:"is_liveness_enable"`         // 是否启用 Liveness 探针
	LivenessProbe        HealthCheckDetail     `json:"liveness_probe" valid:"liveness_probe"`                 // Liveness 探针配置
	ServiceType          string                `json:"service_type" valid:"service_type"`                     // Service 类型
	ServiceName          string                `json:"service_name" valid:"service_name"`                     // Service 名称
	RevisionHistoryLimit *int32                `json:"revision_history_limit" valid:"revision_history_limit"` // 保留的历史版本数量（可选，默认10）

	// 调度规则配置
	SchedulingPolicy SchedulingPolicy    `json:"scheduling_policy,omitempty" valid:"scheduling_policy"` // 调度策略: default/spread/pack/custom
	NodeSelector     map[string]string   `json:"node_selector,omitempty" valid:"node_selector"`         // 节点选择器（custom 模式使用）
	Affinity         *corev1.Affinity    `json:"affinity,omitempty" valid:"affinity"`                   // 亲和性配置（原生K8s格式，高级用户）
	Tolerations      []corev1.Toleration `json:"tolerations,omitempty" valid:"tolerations"`             // 容忍配置

	// 高级调度规则（前端友好格式）
	NodeAffinityRules     []NodeAffinityRule `json:"node_affinity_rules,omitempty"`     // 节点亲和性规则（简化格式）
	TopologySpreadConfigs []TopologySpread   `json:"topology_spread_configs,omitempty"` // 拓扑分布约束（简化格式）

	// 资源配置和探针配置（Rancher/Kuboard 风格）
	Resources *ResourceRequirements `json:"resources,omitempty" valid:"resources"` // 资源需求（requests/limits）
	Probes    *ProbeConfig          `json:"probes,omitempty"    valid:"probes"`    // 探针配置（liveness/readiness/startup）
}

// EnvironmentVariable 定义环境变量
type EnvironmentVariable struct {
	Name  string `json:"name" valid:"name"`   // 环境变量名
	Value string `json:"value" valid:"value"` // 环境变量值
}

// ---------------------- Deployment 更新 ----------------------

func NewKubeDeploymentUpdateRequest() *KubeDeploymentUpdateRequest {
	return &KubeDeploymentUpdateRequest{}
}

type KubeDeploymentUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"` // 命名空间
	Content   json.RawMessage `json:"content" valid:"content"`     // 更新内容（一般是 YAML/JSON）
	Name      string          `json:"name" valid:"name"`           // Deployment 名称
}

func ValidKubeDeploymentUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Deployment 列表 ----------------------

func NewKubeDeploymentListRequest() *KubeDeploymentListRequest {
	return &KubeDeploymentListRequest{}
}

type KubeDeploymentListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`   // 页码
	Limit int `json:"limit" valid:"limit"` // 每页条数
}

func ValidKubeDeploymentListRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Deployment 扩缩容 ----------------------
func NewKubeDeploymentScaleRequest() *KubeDeploymentScaleRequest {
	return &KubeDeploymentScaleRequest{}
}

type KubeDeploymentScaleRequest struct {
	KubeCommonRequest
	ScaleNum int32 `json:"scale_num" valid:"scale_num"` // 副本数量
}

func ValidKubeDeploymentScaleRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
		"scale_num": []string{"numeric"}, // 改为 numeric，允许 0
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name不能为空"},
		"namespace": []string{"required: namespace不能为空"},
		"scale_num": []string{"numeric: scale_num必须是数字"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 回滚 ----------------------

func NewKubeDeploymentRollbackRequest() *KubeDeploymentRollbackRequest {
	return &KubeDeploymentRollbackRequest{}
}

type KubeDeploymentRollbackRequest struct {
	KubeCommonRequest
	ReplicaSet string `json:"replica_set" valid:"replica_set"` // 指定回滚的 ReplicaSet
}

func ValidKubeDeploymentRollbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":   []string{"required"},
		"name":        []string{"required"},
		"replica_set": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace":   []string{"required: namespace 不能为空"},
		"name":        []string{"required: name 不能为空"},
		"replica_set": []string{"required: replica_set 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 重启 ----------------------

func NewKubeDeploymentRestartRequest() *KubeDeploymentRestartRequest {
	return &KubeDeploymentRestartRequest{}
}

type KubeDeploymentRestartRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentRestartRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

/* ---------------- 获取详情 ---------------- */

func NewKubeDeploymentDetailRequest() *KubeDeploymentDetailRequest {
	return &KubeDeploymentDetailRequest{}
}

type KubeDeploymentDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Deployment 删除 ----------------------

func NewKubeDeploymentDeleteRequest() *KubeDeploymentDeleteRequest {
	return &KubeDeploymentDeleteRequest{}
}

type KubeDeploymentDeleteRequest struct {
	KubeCommonRequest
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"` // 优雅终止时间（秒）
	Force              bool   `json:"force,omitempty" valid:"force"`                               // 是否强制删除
}

func ValidKubeDeploymentDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"name":      []string{"required: name不能为空"},
		"namespace": []string{"required: namespace不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Deployment 镜像更新 ----------------------
func NewKubeDeploymentUpdateImageRequest() *KubeDeploymentUpdateImageRequest {
	return &KubeDeploymentUpdateImageRequest{}
}

type KubeDeploymentUpdateImageRequest struct {
	KubeCommonRequest
	Container string `json:"container" valid:"container"` // 目标容器名称
	Image     string `json:"image" valid:"image"`         // 新镜像地址，例如 nginx:1.27
}

func ValidKubeDeploymentUpdateImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Deployment 创建和server服务 ----------------------

func NewKubeDeploymentCreateSvcRequest() *KubeDeploymentCreateSvcRequest {
	return &KubeDeploymentCreateSvcRequest{}
}

type KubeDeploymentCreateSvcRequest struct {
	KubeCommonRequest
}

func ValidKubeDeploymentCreateSvcRequest(data interface{}, ctx context.Context) map[string][]string {
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

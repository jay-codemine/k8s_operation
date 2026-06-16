package requests

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"

	"k8soperation/pkg/valid"
)

/*
StatefulSet 与 Deployment 的主要差异（本文件已覆盖）：
1) 创建时必须指定 ServiceName（通常搭配 Headless Service）。
2) 常见场景需要每个 Pod 绑定独立 PVC（.spec.volumeClaimTemplates）。
3) 不支持“回滚到某个历史版本”的能力（没有 ReplicaSet 历史）。
4) 其它操作（列表、详情、扩缩容、重启、更新镜像、删除、事件查询）与 Deployment 基本一致。
*/

// ---------------------- StatefulSet 创建 ----------------------

func NewKubeStatefulSetCreateRequest() *KubeStatefulSetCreateRequest {
	return &KubeStatefulSetCreateRequest{}
}

// KubeStatefulSetCreateRequest 定义创建 StatefulSet 的请求结构
type KubeStatefulSetCreateRequest struct {
	// 基本信息
	Name           string  `json:"name" valid:"name"`                       // StatefulSet 名称
	Namespace      string  `json:"namespace" valid:"namespace"`             // 命名空间
	Description    *string `json:"description" valid:"description"`         // 描述信息（可选）
	Replicas       int32   `json:"replicas" valid:"replicas"`               // 副本数
	ContainerImage string  `json:"container_image" valid:"container_image"` // 主容器镜像

	// 镜像/启动相关（可选）
	ImagePullSecret      *string `json:"image_pull_secret" valid:"image_pull_secret"`           // 镜像拉取密钥
	ContainerCommand     *string `json:"container_command" valid:"container_command"`           // 容器启动命令
	ContainerCommandArgs *string `json:"container_command_args" valid:"container_command_args"` // 容器启动参数

	// 资源/标签/环境变量/端口
	MemoryRequirement *string               `json:"memory_requirement" valid:"memory_requirement"` // 内存需求（可选）
	CpuRequirement    *string               `json:"cpu_requirement" valid:"cpu_requirement"`       // CPU 需求（可选）
	Labels            []Label               `json:"labels" valid:"labels"`                         // 标签
	Variables         []EnvironmentVariable `json:"variables" valid:"variables"`                   // 环境变量
	PortMappings      []PortMapping         `json:"port_mappings" valid:"port_mappings"`           // 端口映射

	// 探针/特权
	RunAsPrivileged   bool              `json:"run_as_privileged" valid:"run_as_privileged"`     // 是否以特权模式运行
	IsReadinessEnable bool              `json:"is_readiness_enable" valid:"is_readiness_enable"` // 是否启用 Readiness
	ReadinessProbe    HealthCheckDetail `json:"readiness_probe" valid:"readiness_probe"`         // Readiness 配置
	IsLivenessEnable  bool              `json:"is_liveness_enable" valid:"is_liveness_enable"`   // 是否启用 Liveness
	LivenessProbe     HealthCheckDetail `json:"liveness_probe" valid:"liveness_probe"`           // Liveness 配置

	// Service 相关（StatefulSet 必须绑定一个服务名，通常为 Headless Service）
	IsCreateService bool   `json:"is_create_service" valid:"is_create_service"` // 是否同时创建 Service（建议 true）
	ServiceName     string `json:"service_name" valid:"service_name"`           // 必填：绑定的 Service 名（Headless 推荐）
	ServiceType     string `json:"service_type" valid:"service_type"`           // Service 类型（ClusterIP/NodePort/...；Headless 时忽略）

	// 滚动更新策略（可选）
	UpdatePartition *int32 `json:"update_partition" valid:"update_partition"` // RollingUpdate partition 参数（金丝雀发布），默认 0 表示全量更新

	// 每个 Pod 的独立 PVC（StatefulSet 常用）
	VolumeClaimTemplates []VolumeClaimTemplate `json:"volume_claim_templates" valid:"volume_claim_templates"`
}

// VolumeClaimTemplate 定义 StatefulSet 的 PVC 模板
type VolumeClaimTemplate struct {
	Name         string `json:"name" valid:"name"`                   // PVC 名称（与 volumeMounts 对应）
	StorageClass string `json:"storage_class" valid:"storage_class"` // 存储类
	AccessMode   string `json:"access_mode" valid:"access_mode"`     // 例如 ReadWriteOnce
	StorageSize  string `json:"storage_size" valid:"storage_size"`   // 例如 "5Gi"
	MountPath    string `json:"mount_path" valid:"mount_path"`       // 容器内挂载路径
}

// 校验：StatefulSet 创建
func ValidKubeStatefulSetCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":            []string{"required"},
		"namespace":       []string{"required"},
		"container_image": []string{"required"},
		"replicas":        []string{"required"},
		"service_name":    []string{"required"}, // StatefulSet 必须绑定 ServiceName
		// volume_claim_templates 可选，但若传入需做基本校验（此处交给服务层做更细校验）
	}
	messages := govalidator.MapData{
		"name":            []string{"required: name不能为空"},
		"namespace":       []string{"required: namespace不能为空"},
		"container_image": []string{"required: image不能为空"},
		"replicas":        []string{"required: replicas不能为空"},
		"service_name":    []string{"required: service_name 不能为空（StatefulSet 需绑定 Service）"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- StatefulSet 更新 ----------------------

func NewKubeStatefulSetUpdateRequest() *KubeStatefulSetUpdateRequest {
	return &KubeStatefulSetUpdateRequest{}
}

type KubeStatefulSetUpdateRequest struct {
	Namespace string          `json:"namespace" valid:"namespace"` // 命名空间
	Content   json.RawMessage `json:"content" valid:"content"`     // 更新内容（YAML/JSON）
	Name      string          `json:"name" valid:"name"`           // StatefulSet 名称
}

func ValidKubeStatefulSetUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 列表 ----------------------

func NewKubeStatefulSetListRequest() *KubeStatefulSetListRequest {
	return &KubeStatefulSetListRequest{}
}

type KubeStatefulSetListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`   // 页码
	Limit int `json:"limit" valid:"limit"` // 每页条数
}

func ValidKubeStatefulSetListRequest(data interface{}, ctx *gin.Context) map[string][]string {
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
			"max: limit不能超过1000",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- StatefulSet 扩缩容 ----------------------

func NewKubeStatefulSetScaleRequest() *KubeStatefulSetScaleRequest {
	return &KubeStatefulSetScaleRequest{}
}

type KubeStatefulSetScaleRequest struct {
	KubeCommonRequest
	ScaleNum int32 `json:"scale_num" valid:"scale_num"` // 副本数量
}

func ValidKubeStatefulSetScaleRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 重启（滚动） ----------------------
func NewKubeStatefulSetRestartRequest() *KubeStatefulSetRestartRequest {
	return &KubeStatefulSetRestartRequest{}
}

type KubeStatefulSetRestartRequest struct {
	KubeCommonRequest
}

func ValidKubeStatefulSetRestartRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 详情 ----------------------

func NewKubeStatefulSetDetailRequest() *KubeStatefulSetDetailRequest {
	return &KubeStatefulSetDetailRequest{}
}

type KubeStatefulSetDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeStatefulSetDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 删除 ----------------------

func NewKubeStatefulSetDeleteRequest() *KubeStatefulSetDeleteRequest {
	return &KubeStatefulSetDeleteRequest{}
}

type KubeStatefulSetDeleteRequest struct {
	KubeCommonRequest
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"` // 优雅终止时间（秒）
	Force              bool   `json:"force,omitempty" valid:"force"`                               // 是否强制删除
	// 注意：删除 StatefulSet 时通常不会自动删除 PVC，需要额外逻辑（控制层处理）
}

func ValidKubeStatefulSetDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 更新镜像 ----------------------

func NewKubeStatefulSetUpdateImageRequest() *KubeStatefulSetUpdateImageRequest {
	return &KubeStatefulSetUpdateImageRequest{}
}

type KubeStatefulSetUpdateImageRequest struct {
	KubeCommonRequest
	Container string `json:"container" valid:"container"` // 目标容器名称
	Image     string `json:"image" valid:"image"`         // 新镜像地址，例如 mysql:8.0
}

func ValidKubeStatefulSetUpdateImageRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- StatefulSet 创建 Service（可选，通常为 Headless） ----------------------

func NewKubeStatefulSetCreateSvcRequest() *KubeStatefulSetCreateSvcRequest {
	return &KubeStatefulSetCreateSvcRequest{}
}

type KubeStatefulSetCreateSvcRequest struct {
	KubeCommonRequest
	// 是否创建 Headless Service（创建时应设置 clusterIP=None）
	Headless bool `json:"headless" valid:"headless"`
}

func ValidKubeStatefulSetCreateSvcRequest(data interface{}, ctx context.Context) map[string][]string {
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

// ---------------------- StatefulSet 回滚 ----------------------

type KubeStatefulSetRollbackRequest struct {
	KubeCommonRequest
	// 指定回滚到的历史版本（ControllerRevision）
	RevisionName string `json:"revision_name" valid:"revision_name"`
}

// StatefulSet 回滚请求
func NewKubeStatefulSetRollbackRequest() *KubeStatefulSetRollbackRequest {
	return &KubeStatefulSetRollbackRequest{}
}

func ValidKubeStatefulSetRollbackRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

type CommonIdRequest struct {
	ID uint32 `json:"id" form:"id" valid:"id"`
}

func NewCommonIdRequest() *CommonIdRequest {
	return &CommonIdRequest{}
}

func ValidCommonIdRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required", "numeric", "min:1"},
	}
	messages := govalidator.MapData{
		"id": []string{
			"required:id不能为空",
			"numeric:id必须是数字",
			"min:id必须大于0",
		},
	}

	// 校验入参
	errs := valid.ValidateOptions(data, rules, messages)

	return errs
}

// KubeCommonRequest 公共请求
type KubeCommonRequest struct {
	Name      string `json:"name" form:"name" valid:"name"`
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
}

func NewKubeCommonRequest() *KubeCommonRequest {
	return &KubeCommonRequest{}
}

func VaildKubeCommonRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":      []string{"required"},
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"name": []string{
			"required: name不能为空",
		},
		"namespace": []string{
			"required: namespace不能为空",
		},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, messages)
}

// 通用标签
type Label struct {
	Key   string `json:"key"   valid:"key"`
	Value string `json:"value" valid:"value"`
}

// ValidKubeDeploymentCreateRequest 校验创建 Deployment 请求
func ValidKubeDeploymentCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":            []string{"required"},
		"namespace":       []string{"required"},
		"container_image": []string{"required"},
		"replicas":        []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace不能为空"},
		"name":      []string{"required: name不能为空"},
		"container_image": []string{
			"required: image不能为空",
		},
		"replicas": []string{"required: replicas不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 容器端口映射
type PortMapping struct {
	Port       int32  `json:"port"        valid:"port"`
	TargetPort int32  `json:"target_port" valid:"target_port"`
	Protocol   string `json:"protocol"    valid:"protocol"` // TCP/UDP/SCTP
}

// HTTP 头（用于探针）
type HttpHeader struct {
	Name  string `json:"name"  valid:"name"`
	Value string `json:"value" valid:"value"`
}

// 健康检查探针（readiness/liveness 共用）
type HealthCheckDetail struct {
	Type                string       `json:"type"                  valid:"type"` // HTTP|TCP|Command
	Protocol            string       `json:"protocol"              valid:"protocol"`
	Path                string       `json:"path"                  valid:"path"`
	Port                int32        `json:"port"                  valid:"port"`
	HttpHeader          []HttpHeader `json:"http_header"           valid:"http_header"`
	InitialDelaySeconds int32        `json:"initial_delay_seconds" valid:"initial_delay_seconds"`
	PeriodSeconds       int32        `json:"period_seconds"        valid:"period_seconds"`
	TimeoutSeconds      int32        `json:"timeout_seconds"       valid:"timeout_seconds"`
	SuccessThreshold    int32        `json:"success_threshold"     valid:"success_threshold"`
	FailureThreshold    int32        `json:"failure_threshold"     valid:"failure_threshold"`
	Command             string       `json:"command"               valid:"command"` // Type=Command 时使用
}

// ResourceRequirements 资源需求配置（Rancher/Kuboard 风格）
type ResourceRequirements struct {
	CPURequest    *string `json:"cpu_request,omitempty"    valid:"cpu_request"`    // CPU 请求，如 100m, 0.5, 1
	CPULimit      *string `json:"cpu_limit,omitempty"      valid:"cpu_limit"`      // CPU 限制
	MemoryRequest *string `json:"memory_request,omitempty" valid:"memory_request"` // 内存请求，如 64Mi, 128Mi, 1Gi
	MemoryLimit   *string `json:"memory_limit,omitempty"   valid:"memory_limit"`   // 内存限制
}

// ProbeConfig 探针配置（支持三种探针）
type ProbeConfig struct {
	// Liveness Probe - 存活探针
	EnableLiveness bool              `json:"enable_liveness,omitempty" valid:"enable_liveness"`
	LivenessProbe  HealthCheckDetail `json:"liveness_probe,omitempty"  valid:"liveness_probe"`
	
	// Readiness Probe - 就绪探针
	EnableReadiness bool              `json:"enable_readiness,omitempty" valid:"enable_readiness"`
	ReadinessProbe  HealthCheckDetail `json:"readiness_probe,omitempty"  valid:"readiness_probe"`
	
	// Startup Probe - 启动探针
	EnableStartup bool              `json:"enable_startup,omitempty" valid:"enable_startup"`
	StartupProbe  HealthCheckDetail `json:"startup_probe,omitempty"  valid:"startup_probe"`
}

/* ========== 通用分页 DTO ========== */

type PageLimit struct {
	Page  int `json:"page"  form:"page"  valid:"page"`  // 页码，从 1 开始
	Limit int `json:"limit" form:"limit" valid:"limit"` // 每页条数
}

/* ========== 通用删除选项 DTO ========== */

type DeleteOptions struct {
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty" valid:"grace_period_seconds"`
	Force              bool   `json:"force,omitempty"                 valid:"force"`
	Wait               bool   `json:"wait,omitempty"                  valid:"wait"`
	TimeoutSeconds     *int64 `json:"timeout_seconds,omitempty"       valid:"timeout_seconds"`
}

/* ========== 通用事件查询 DTO ========== */

type KubeEventListRequest struct {
	Namespace     string `json:"namespace,omitempty" valid:"namespace"` // 为空=全局
	Kind          string `json:"kind,omitempty"      valid:"kind"`      // Deployment/StatefulSet/DaemonSet/Pod/Node...
	Name          string `json:"name,omitempty"      valid:"name"`
	Type          string `json:"type,omitempty"      valid:"type"` // Normal | Warning
	Reason        string `json:"reason,omitempty"    valid:"reason"`
	Limit         int64  `json:"limit,omitempty"     valid:"limit"` // 默认 50
	ContinueToken string `json:"continue,omitempty"  valid:"continue"`
	SinceSeconds  int64  `json:"since_seconds,omitempty" valid:"since_seconds"`
}

func NewKubeEventListRequest() *KubeEventListRequest {
	return &KubeEventListRequest{
		Limit:        50,
		SinceSeconds: 1000,
	}
}

/* ========== 通用校验辅助（可选） ========== */

// 详情/删除/重启等：Name & Namespace 必填
func ValidNameNamespace(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
	}
	msg := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

// 分页规则：Page>=1, 1<=Limit<=1000
func ValidPageLimit(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required", "min:1"},
		"limit": []string{"required", "min:1", "max:1000"},
	}
	msg := govalidator.MapData{
		"page":  []string{"required: page 不能为空", "min: page 必须 >= 1"},
		"limit": []string{"required: limit 不能为空", "min: limit 必须 >= 1", "max: limit 不能超过 1000"},
	}
	return valid.ValidateOptions(data, rules, msg)
}

func ValidKubeEventListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"limit":         []string{"required"}, // 与后端 clamp/文档一致
		"since_seconds": []string{"required"},
		// 如需限制：打开下一行；不想限制就删掉
		// "type":          []string{"in:Normal,Warning"},
	}
	msgs := govalidator.MapData{
		"limit":         []string{"numeric_between: limit 取值范围 1~500"},
		"since_seconds": []string{"numeric_min: since_seconds 不能为负数"},
		// "type":          []string{"in: type 仅支持 Normal 或 Warning"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

/* ========== 通用 YAML 应用 DTO ========== */

// KubeApplyYamlRequest 通用 YAML 应用请求（用于命名空间级资源）
type KubeApplyYamlRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name" form:"name" valid:"name"`
	Yaml      string `json:"yaml" valid:"yaml"`
}

func NewKubeApplyYamlRequest() *KubeApplyYamlRequest {
	return &KubeApplyYamlRequest{}
}

func ValidKubeApplyYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
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

// KubeApplyYamlClusterRequest 通用 YAML 应用请求（用于集群级资源，如 Node、Namespace）
type KubeApplyYamlClusterRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
	Yaml string `json:"yaml" valid:"yaml"`
}

func NewKubeApplyYamlClusterRequest() *KubeApplyYamlClusterRequest {
	return &KubeApplyYamlClusterRequest{}
}

func ValidKubeApplyYamlClusterRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
		"yaml": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
		"yaml": {"required: yaml 不能为空"},
	})
}

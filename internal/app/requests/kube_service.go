package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ---------------------- Service 公共类型 ----------------------

type ServicePortRule struct {
	// 端口名，可选
	Name string `json:"name" valid:"name"`
	// Service 暴露端口（1~65535）
	Port int32 `json:"port" valid:"port"`
	// 指向 Pod 的容器端口；为兼容命名端口，这里用字符串（Swagger 友好）
	TargetPort string `json:"target_port" valid:"target_port"`
	// 协议：TCP/UDP/SCTP
	Protocol string `json:"protocol" valid:"protocol"`
	// NodePort（仅在 NodePort/LoadBalancer 类型下生效），可选
	NodePort *int32 `json:"node_port" valid:"node_port"`
	// App 协议（可选）
	AppProtocol *string `json:"app_protocol" valid:"app_protocol"`
}

// ---------------------- Service 创建 ----------------------

func NewKubeServiceCreateRequest() *KubeServiceCreateRequest { return &KubeServiceCreateRequest{} }

type KubeServiceCreateRequest struct {
	// 命名空间 & 名称
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`

	// Service 类型：ClusterIP / NodePort / LoadBalancer / Headless
	// Headless 时 cluster_ip 建议传 "None"
	Type string `json:"type" valid:"type"`

	// 选择器：建议用 label 列表；也支持传 map（Swagger 里以字符串展示）
	SelectorLabels []Label           `json:"selector_labels" valid:"labels"`
	Selector       map[string]string `json:"selector" swaggertype:"string" valid:"-"`

	// 端口规则
	Ports []ServicePortRule `json:"ports" valid:"ports"`

	// 可选项
	ClusterIP          *string           `json:"cluster_ip" valid:"cluster_ip"`                           // 设为 "None" 代表 Headless
	SessionAffinity    *string           `json:"session_affinity" valid:"session_affinity"`               // None / ClientIP
	ExternalTrafficPol *string           `json:"external_traffic_policy" valid:"external_traffic_policy"` // Local / Cluster（LB/NodePort 常用）
	Annotations        map[string]string `json:"annotations" swaggertype:"string" valid:"-"`
}

func ValidKubeServiceCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"type":      []string{"required"},
		"ports":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"type":      []string{"required: type 不能为空(ClusterIP/NodePort/LoadBalancer/Headless)"},
		"ports":     []string{"required: ports 至少配置一个"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Service 更新（通用 JSON/字符串） ----------------------

func NewKubeServiceUpdateRequest() *KubeServiceUpdateRequest { return &KubeServiceUpdateRequest{} }

type KubeServiceUpdateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`
	// 为避免 swagger 解析 RawMessage 报错，这里用字符串承载 JSON/YAML
	// 后端自行 json.Unmarshal/结构化校验
	Content string `json:"content" valid:"content"` // JSON/YAML 字符串
}

func ValidKubeServiceUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"content":   []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"content":   []string{"required: content 不能为空（JSON/YAML 字符串）"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Service 列表 ----------------------

func NewKubeServiceListRequest() *KubeServiceListRequest { return &KubeServiceListRequest{} }

type KubeServiceListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`
	Limit int `json:"limit" valid:"limit"`
}

func ValidKubeServiceListRequest(data interface{}, ctx *gin.Context) map[string][]string {
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
			"max: limit不能超过000",
		},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Service 详情 ----------------------

func NewKubeServiceDetailRequest() *KubeServiceDetailRequest { return &KubeServiceDetailRequest{} }

type KubeServiceDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeServiceDetailRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Service 删除 ----------------------

func NewKubeServiceDeleteRequest() *KubeServiceDeleteRequest { return &KubeServiceDeleteRequest{} }

type KubeServiceDeleteRequest struct {
	KubeCommonRequest
}

func ValidKubeServiceDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
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

// ---------------------- Service 仅更新端口（便于前端单独改端口） ----------------------

func NewKubeServiceUpdatePortsRequest() *KubeServiceUpdatePortsRequest {
	return &KubeServiceUpdatePortsRequest{}
}

type KubeServiceUpdatePortsRequest struct {
	KubeCommonRequest
	Ports []ServicePortRule `json:"ports" valid:"ports"`
}

func ValidKubeServiceUpdatePortsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"ports":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"ports":     []string{"required: ports 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ---------------------- Service 仅更新类型（ClusterIP/NodePort/LB/Headless） ----------------------

func NewKubeServiceUpdateTypeRequest() *KubeServiceUpdateTypeRequest {
	return &KubeServiceUpdateTypeRequest{}
}

type KubeServiceUpdateTypeRequest struct {
	KubeCommonRequest
	Type string `json:"type" valid:"service_type"`
	// Headless 时可传 "None"
	ClusterIP *string `json:"cluster_ip" valid:"cluster_ip"`
}

func ValidKubeServiceUpdateTypeRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"type":      []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"type":      []string{"required: type 不能为空(ClusterIP/NodePort/LoadBalancer/Headless)"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// KubeServiceEndpointsRequest 定义获取 Service Endpoints 的请求参数
type KubeServiceEndpointsRequest struct {
	Namespace string `form:"namespace" json:"namespace" binding:"required" valid:"namespace"`
	Name      string `form:"name" json:"name" binding:"required" valid:"name"`
}

// NewKubeServiceEndpointsRequest 工厂函数，方便控制器层初始化
func NewKubeServiceEndpointsRequest() *KubeServiceEndpointsRequest {
	return &KubeServiceEndpointsRequest{}
}

// ValidKubeServiceEndpointsRequest 用于参数校验（供 valid.Validate 调用）
func ValidKubeServiceEndpointsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	req := data.(*KubeServiceEndpointsRequest)
	rules := govalidator.MapData{
		"namespace": []string{"required", "min:1"},
		"name":      []string{"required", "min:1"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required:命名空间不能为空"},
		"name":      []string{"required:Service 名称不能为空"},
	}
	return ValidateOptions(req, rules, messages)
}

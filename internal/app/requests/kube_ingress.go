package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// IngressPathRule 单条路径→后端 Service 映射
type IngressPathRule struct {
	// 访问路径（如 /、/api、/api/(.*)）
	Path string `json:"path" valid:"required"`

	// pathType: Exact / Prefix / ImplementationSpecific
	PathType string `json:"path_type" valid:"ingress_path_type"`

	// 后端 Service 名称
	ServiceName string `json:"service_name" valid:"required"`

	// 后端 Service 端口（支持命名端口，故用字符串；可传 "80" 或 "http"）
	ServicePort string `json:"service_port" valid:"required"`
}

// IngressRule 单个 Host 下多条 Path
type IngressRule struct {
	Host  string            `json:"host" valid:"required"`
	Paths []IngressPathRule `json:"paths" valid:"required"`
}

// IngressTLS TLS 绑定（hosts + 证书 Secret）
type IngressTLS struct {
	Hosts      []string `json:"hosts" valid:"required"`
	SecretName string   `json:"secret_name" valid:"required"`
}

//
// ============ Ingress 创建 ============
//

func NewKubeIngressCreateRequest() *KubeIngressCreateRequest { return &KubeIngressCreateRequest{} }

type KubeIngressCreateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`

	// 例如："nginx"、"traefik"；可为空（走默认 class）
	IngressClassName string `json:"ingress_class_name" valid:"-"`

	// 规则：支持多个 host
	Rules []IngressRule `json:"rules" valid:"rules"`

	// 可选 TLS 列表
	TLS []IngressTLS `json:"tls" valid:"-"`

	// 可选注解（如 rewrite、限流等）
	Annotations map[string]string `json:"annotations" swaggertype:"string" valid:"-"`

	// 可选标签
	Labels map[string]string `json:"labels" swaggertype:"string" valid:"-"`
}

func ValidKubeIngressCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"rules":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"rules":     []string{"required: 至少配置一条规则（host + paths）"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

//
// ============ Ingress 列表 ============
//

func NewKubeIngressListRequest() *KubeIngressListRequest { return &KubeIngressListRequest{} }

type KubeIngressListRequest struct {
	KubeCommonRequest
	Page  int `json:"page" valid:"page"`
	Limit int `json:"limit" valid:"limit"`
}

func ValidKubeIngressListRequest(data interface{}, _ *gin.Context) map[string][]string {
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

//
// ============ Ingress 详情/删除 ============
//

func NewKubeIngressDetailRequest() *KubeIngressDetailRequest { return &KubeIngressDetailRequest{} }

type KubeIngressDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeIngressDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
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

func NewKubeIngressDeleteRequest() *KubeIngressDeleteRequest { return &KubeIngressDeleteRequest{} }

type KubeIngressDeleteRequest struct {
	KubeCommonRequest
}

func ValidKubeIngressDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
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

//
// ============ Ingress 通用更新（content 字符串，后端自行 Unmarshal） ============
//

func NewKubeIngressUpdateRequest() *KubeIngressUpdateRequest { return &KubeIngressUpdateRequest{} }

// 支持 StrategicMergePatch / MergePatch 的原文字符串（JSON/YAML）
// 方便你和 Service 的 UpdateRequest 一致
type KubeIngressUpdateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`
	Content   string `json:"content" valid:"content"` // JSON/YAML 字符串
}

func ValidKubeIngressUpdateRequest(data interface{}, _ *gin.Context) map[string][]string {
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

//
// ============ Ingress 局部更新（便于前端做表单化改动） ============
//

// 仅更新注解（常用：rewrite、proxy-* 等）
func NewKubeIngressUpdateAnnotationsRequest() *KubeIngressUpdateAnnotationsRequest {
	return &KubeIngressUpdateAnnotationsRequest{}
}

type KubeIngressUpdateAnnotationsRequest struct {
	KubeCommonRequest
	Annotations map[string]string `json:"annotations" swaggertype:"string" valid:"required"`
}

func ValidKubeIngressUpdateAnnotationsRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace":   []string{"required"},
		"name":        []string{"required"},
		"annotations": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace":   []string{"required: namespace 不能为空"},
		"name":        []string{"required: name 不能为空"},
		"annotations": []string{"required: annotations 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 仅更新规则（hosts/paths），方便前端“路径编辑器”
func NewKubeIngressUpdateRulesRequest() *KubeIngressUpdateRulesRequest {
	return &KubeIngressUpdateRulesRequest{}
}

type KubeIngressUpdateRulesRequest struct {
	KubeCommonRequest
	Rules []IngressRule `json:"rules" valid:"required"`
}

func ValidKubeIngressUpdateRulesRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"rules":     []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"rules":     []string{"required: 规则不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 仅更新 TLS（绑定/更换证书）
func NewKubeIngressUpdateTLSRequest() *KubeIngressUpdateTLSRequest {
	return &KubeIngressUpdateTLSRequest{}
}

type KubeIngressUpdateTLSRequest struct {
	KubeCommonRequest
	TLS []IngressTLS `json:"tls" valid:"required"`
}

func ValidKubeIngressUpdateTLSRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"name":      []string{"required"},
		"tls":       []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"name":      []string{"required: name 不能为空"},
		"tls":       []string{"required: tls 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

//
// ============ Ingress 辅助查询 ============
//

// 查询 IngressClass 列表（用于前端下拉）
func NewKubeIngressControllersRequest() *KubeIngressControllersRequest {
	return &KubeIngressControllersRequest{}
}

type KubeIngressControllersRequest struct {
	// 可扩展查询条件，这里预留
}

// 查询命名空间可用 TLS Secret（kubernetes.io/tls）
func NewKubeIngressTlsSecretsRequest() *KubeIngressTlsSecretsRequest {
	return &KubeIngressTlsSecretsRequest{}
}

type KubeIngressTlsSecretsRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
}

func ValidKubeIngressTlsSecretsRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// 查询已占用的 host（用于前端避免重复域名）
func NewKubeIngressHostsRequest() *KubeIngressHostsRequest { return &KubeIngressHostsRequest{} }

type KubeIngressHostsRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	// 可选模糊过滤
	Q string `json:"q" form:"q" valid:"-"`
}

func ValidKubeIngressHostsRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
	}
	messages := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

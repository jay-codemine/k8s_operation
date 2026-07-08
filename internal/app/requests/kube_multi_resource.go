package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ================= 多资源 YAML 创建 =================

func NewKubeMultiResourceApplyYamlRequest() *KubeMultiResourceApplyYamlRequest {
	return &KubeMultiResourceApplyYamlRequest{}
}

// KubeMultiResourceApplyYamlRequest 多资源 YAML 应用请求
type KubeMultiResourceApplyYamlRequest struct {
	Yaml string `json:"yaml" valid:"yaml"` // 多个 YAML 文档，用 --- 分隔
}

// ValidKubeMultiResourceApplyYamlRequest 校验多资源 YAML 请求
func ValidKubeMultiResourceApplyYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{"yaml": {"required"}}
	messages := govalidator.MapData{
		"yaml": {"required: yaml 不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// MultiResourceParsedResult 多资源解析结果
type MultiResourceParsedResult struct {
	Resources []ParsedResource `json:"resources"`
	Total     int              `json:"total"`
	Errors    []string         `json:"errors,omitempty"`
}

// ParsedResource 解析后的单个资源
type ParsedResource struct {
	Index      int               `json:"index"`      // 在原始 YAML 中的索引
	Kind       string            `json:"kind"`       // 资源类型 (Deployment, Service, ConfigMap 等)
	APIVersion string            `json:"apiVersion"` // API 版本
	Name       string            `json:"name"`       // 资源名称
	Namespace  string            `json:"namespace"`  // 命名空间
	Content    string            `json:"content"`    // 原始 YAML 内容
	Order      int               `json:"order"`      // 创建顺序 (数值越小优先级越高)
	DependsOn  []ResourceRef     `json:"dependsOn,omitempty"` // 依赖的资源
	Warnings   []string          `json:"warnings,omitempty"`  // 警告信息
}

// ResourceRef 资源引用
type ResourceRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	FieldPath string `json:"fieldPath"` // 引用的字段路径
}

// MultiResourceCreateResult 多资源创建结果
type MultiResourceCreateResult struct {
	Created []CreatedResource `json:"created"` // 成功创建的资源
	Failed  []FailedResource  `json:"failed"`  // 创建失败的资源
	Total   int               `json:"total"`   // 总资源数
}

// CreatedResource 成功创建的资源
type CreatedResource struct {
	Index     int    `json:"index"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Message   string `json:"message"`
}

// FailedResource 创建失败的资源
type FailedResource struct {
	Index   int    `json:"index"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Error   string `json:"error"`
	Message string `json:"message"`
}
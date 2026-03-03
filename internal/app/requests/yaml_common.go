package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// YamlCreateRequest 通用 YAML 创建请求
type YamlCreateRequest struct {
	Yaml string `json:"yaml" valid:"yaml" binding:"required"` // YAML 内容
}

// NewYamlCreateRequest 创建 YAML 请求对象
func NewYamlCreateRequest() *YamlCreateRequest {
	return &YamlCreateRequest{}
}

// ValidYamlCreateRequest 校验规则
func ValidYamlCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"yaml": []string{"required"},
	}
	messages := govalidator.MapData{
		"yaml": []string{
			"required: YAML 内容为必填字段",
		},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, messages)
}

// YamlApplyRequest YAML 应用请求（用于更新已有资源）
type YamlApplyRequest struct {
	Yaml string `json:"yaml" valid:"yaml" binding:"required"` // YAML 内容
}

// NewYamlApplyRequest 创建 YAML Apply 请求对象
func NewYamlApplyRequest() *YamlApplyRequest {
	return &YamlApplyRequest{}
}

// ValidYamlApplyRequest 校验规则
func ValidYamlApplyRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"yaml": []string{"required"},
	}
	messages := govalidator.MapData{
		"yaml": []string{
			"required: YAML 内容为必填字段",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

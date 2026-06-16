package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// ==================== 创建流水线模板 ====================

type TemplateCreateRequest struct {
	Name            string                 `json:"name" valid:"name"`
	Description     string                 `json:"description" valid:"description"`
	Type            string                 `json:"type" valid:"type"`
	Stages          []map[string]any       `json:"stages" valid:"stages"`
	DefaultEnvVars  []map[string]any       `json:"default_env_vars" valid:"default_env_vars"`
	DeployConfig    map[string]any         `json:"deploy_config" valid:"deploy_config"`
	JenkinsTemplate string                 `json:"jenkins_template" valid:"jenkins_template"`
}

func NewTemplateCreateRequest() *TemplateCreateRequest {
	return &TemplateCreateRequest{}
}

func ValidTemplateCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name": []string{"required", "between:1,100"},
		"type": []string{"required", "in:frontend,backend,microservice,database,custom"},
	}
	messages := govalidator.MapData{
		"name": []string{"required:模板名称不能为空", "between:模板名称长度应在1-100之间"},
		"type": []string{"required:模板类型不能为空", "in:模板类型无效，可选值: frontend, backend, microservice, database, custom"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 更新流水线模板 ====================

type TemplateUpdateRequest struct {
	ID              int64                  `json:"id" valid:"id"`
	Name            string                 `json:"name" valid:"name"`
	Description     string                 `json:"description" valid:"description"`
	Type            string                 `json:"type" valid:"type"`
	Stages          []map[string]any       `json:"stages" valid:"stages"`
	DefaultEnvVars  []map[string]any       `json:"default_env_vars" valid:"default_env_vars"`
	DeployConfig    map[string]any         `json:"deploy_config" valid:"deploy_config"`
	JenkinsTemplate string                 `json:"jenkins_template" valid:"jenkins_template"`
}

func NewTemplateUpdateRequest() *TemplateUpdateRequest {
	return &TemplateUpdateRequest{}
}

func ValidTemplateUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":   []string{"required"},
		"name": []string{"between:1,100"},
		"type": []string{"in:frontend,backend,microservice,database,custom"},
	}
	messages := govalidator.MapData{
		"id":   []string{"required:模板ID不能为空"},
		"name": []string{"between:模板名称长度应在1-100之间"},
		"type": []string{"in:模板类型无效，可选值: frontend, backend, microservice, database, custom"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 模板ID请求 ====================

type TemplateIDRequest struct {
	ID int64 `json:"id" valid:"id"`
}

func ValidTemplateIDRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{"required:模板ID不能为空"},
	}
	return valid.ValidateOptions(data, rules, messages)
}

// ==================== 模板列表请求 ====================

type TemplateListRequest struct {
	Page     int    `form:"page" valid:"page"`
	PageSize int    `form:"page_size" valid:"page_size"`
	Keyword  string `form:"keyword" valid:"keyword"`
	Type     string `form:"type" valid:"type"`
}

func NewTemplateListRequest() *TemplateListRequest {
	return &TemplateListRequest{
		Page:     1,
		PageSize: 10,
	}
}

func ValidTemplateListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	return nil // 全部可选
}

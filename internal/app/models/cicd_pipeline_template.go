package models

// 模板类型常量
const (
	TemplateTypeFrontend     = "frontend"     // 前端应用
	TemplateTypeBackend      = "backend"      // 后端服务
	TemplateTypeMicroservice = "microservice" // 微服务
	TemplateTypeDatabase     = "database"     // 数据库
	TemplateTypeCustom       = "custom"       // 自定义
)

// TemplateStage 模板阶段定义
type TemplateStage struct {
	Name        string `json:"name"`        // 阶段名称
	Description string `json:"description"` // 阶段描述
	Order       int    `json:"order"`       // 执行顺序
	Required    bool   `json:"required"`    // 是否必须
}

// TemplateStages 模板阶段数组类型
type TemplateStages []TemplateStage

// TemplateDeployConfig 模板默认部署配置
type TemplateDeployConfig struct {
	Replicas  int                    `json:"replicas"`           // 副本数
	Strategy  string                 `json:"strategy"`           // 部署策略
	Resources map[string]interface{} `json:"resources"`          // 资源配置
}

// CicdPipelineTemplate 流水线模板表
type CicdPipelineTemplate struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;size:100;not null" json:"name"`
	Description string `gorm:"column:description;size:500" json:"description"`
	Type        string `gorm:"column:type;size:50;not null" json:"type"` // frontend/backend/microservice/database/custom

	// 模板配置（JSON 存储）
	Stages          JSONArray `gorm:"column:stages;type:json" json:"stages"`                       // 阶段配置
	DefaultEnvVars  JSONArray `gorm:"column:default_env_vars;type:json" json:"default_env_vars"`   // 默认环境变量
	DeployConfig    JSONMap   `gorm:"column:deploy_config;type:json" json:"deploy_config"`         // 默认部署配置
	JenkinsTemplate string  `gorm:"column:jenkins_template;type:text" json:"jenkins_template"`   // Jenkinsfile 模板

	// 使用统计
	UsageCount int64 `gorm:"column:usage_count;default:0" json:"usage_count"` // 使用次数

	// 元数据
	CreatedUserID int64  `gorm:"column:created_user_id" json:"created_user_id"`
	CreatedAt     uint64 `gorm:"column:created_at" json:"created_at"`
	ModifiedAt    uint64 `gorm:"column:modified_at" json:"modified_at"`
	DeletedAt     uint64 `gorm:"column:deleted_at" json:"deleted_at"`
	IsDel         uint8  `gorm:"column:is_del;default:0" json:"is_del"`
}

func (CicdPipelineTemplate) TableName() string { return "cicd_pipeline_template" }

// TemplateListItem 列表查询返回结构
type TemplateListItem struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Stages      JSONArray `json:"stages"`
	UsageCount  int64     `json:"usage_count"`
	CreatedAt   uint64    `json:"created_at"`
	ModifiedAt  uint64    `json:"modified_at"`
}

// ToTemplateListItem 转换为列表项
func (t *CicdPipelineTemplate) ToTemplateListItem() *TemplateListItem {
	return &TemplateListItem{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Type:        t.Type,
		Stages:      t.Stages,
		UsageCount:  t.UsageCount,
		CreatedAt:   t.CreatedAt,
		ModifiedAt:  t.ModifiedAt,
	}
}

// TemplateDetailResponse 模板详情响应
type TemplateDetailResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Type            string    `json:"type"`
	Stages          JSONArray `json:"stages"`
	DefaultEnvVars  JSONArray `json:"default_env_vars"`
	DeployConfig    JSONMap   `json:"deploy_config"`
	JenkinsTemplate string  `json:"jenkins_template,omitempty"`
	UsageCount      int64   `json:"usage_count"`
	CreatedAt       uint64  `json:"created_at"`
	ModifiedAt      uint64  `json:"modified_at"`
}

// ToDetailResponse 转换为详情响应
func (t *CicdPipelineTemplate) ToDetailResponse() *TemplateDetailResponse {
	return &TemplateDetailResponse{
		ID:              t.ID,
		Name:            t.Name,
		Description:     t.Description,
		Type:            t.Type,
		Stages:          t.Stages,
		DefaultEnvVars:  t.DefaultEnvVars,
		DeployConfig:    t.DeployConfig,
		JenkinsTemplate: t.JenkinsTemplate,
		UsageCount:      t.UsageCount,
		CreatedAt:       t.CreatedAt,
		ModifiedAt:      t.ModifiedAt,
	}
}

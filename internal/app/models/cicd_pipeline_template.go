package models

import dm "k8soperation/internal/domain/cicd"

// ===== 类型别名 =====
type (
	TemplateStage         = dm.TemplateStage
	TemplateStages        = dm.TemplateStages
	TemplateDeployConfig  = dm.TemplateDeployConfig
	CicdPipelineTemplate  = dm.CicdPipelineTemplate
	TemplateListItem      = dm.TemplateListItem
	TemplateDetailResponse = dm.TemplateDetailResponse
)

// ===== 常量别名 =====
const (
	TemplateTypeFrontend     = dm.TemplateTypeFrontend
	TemplateTypeBackend      = dm.TemplateTypeBackend
	TemplateTypeMicroservice = dm.TemplateTypeMicroservice
	TemplateTypeDatabase     = dm.TemplateTypeDatabase
	TemplateTypeCustom       = dm.TemplateTypeCustom
)

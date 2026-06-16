package dao

import (
	"context"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// TemplateCreate 创建流水线模板
func (d *Dao) TemplateCreate(ctx context.Context, template *models.CicdPipelineTemplate) error {
	return global.DB.WithContext(ctx).Create(template).Error
}

// TemplateGetByID 根据ID获取模板
func (d *Dao) TemplateGetByID(ctx context.Context, id int64) (*models.CicdPipelineTemplate, error) {
	var template models.CicdPipelineTemplate
	err := global.DB.WithContext(ctx).
		Where("id = ? AND is_del = 0", id).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// TemplateGetByName 根据名称获取模板
func (d *Dao) TemplateGetByName(ctx context.Context, name string) (*models.CicdPipelineTemplate, error) {
	var template models.CicdPipelineTemplate
	err := global.DB.WithContext(ctx).
		Where("name = ? AND is_del = 0", name).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// TemplateUpdate 更新模板
func (d *Dao) TemplateUpdate(ctx context.Context, template *models.CicdPipelineTemplate) error {
	return global.DB.WithContext(ctx).
		Model(template).
		Select("name", "description", "type", "stages", "default_env_vars", "deploy_config", "jenkins_template", "modified_at").
		Updates(template).Error
}

// TemplateDelete 软删除模板
func (d *Dao) TemplateDelete(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).
		Model(&models.CicdPipelineTemplate{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_del":     1,
			"deleted_at": global.DB.NowFunc().Unix(),
		}).Error
}

// TemplateList 获取模板列表
func (d *Dao) TemplateList(ctx context.Context, keyword, templateType string, page, pageSize int) ([]*models.CicdPipelineTemplate, int64, error) {
	var templates []*models.CicdPipelineTemplate
	var total int64

	db := global.DB.WithContext(ctx).Model(&models.CicdPipelineTemplate{}).Where("is_del = 0")

	// 关键字搜索
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型筛选
	if templateType != "" {
		db = db.Where("type = ?", templateType)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// TemplateIncrUsageCount 增加使用次数
func (d *Dao) TemplateIncrUsageCount(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).
		Model(&models.CicdPipelineTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", global.DB.Raw("usage_count + 1")).Error
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
)

// ==================== 流水线模板 CRUD ====================

// TemplateCreate 创建流水线模板
func (s *Services) TemplateCreate(ctx context.Context, req *requests.TemplateCreateRequest, userID int64) (int64, error) {
	// 检查名称是否已存在
	_, err := s.dao.TemplateGetByName(ctx, req.Name)
	if err == nil {
		return 0, errors.New("模板名称已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("检查名称失败: %w", err)
	}

	now := uint64(time.Now().Unix())

	// 转换 stages
	var stages models.JSONArray
	if req.Stages != nil {
		stagesBytes, _ := json.Marshal(req.Stages)
		json.Unmarshal(stagesBytes, &stages)
	}

	// 转换 defaultEnvVars
	var defaultEnvVars models.JSONArray
	if req.DefaultEnvVars != nil {
		envBytes, _ := json.Marshal(req.DefaultEnvVars)
		json.Unmarshal(envBytes, &defaultEnvVars)
	}

	template := &models.CicdPipelineTemplate{
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Stages:          stages,
		DefaultEnvVars:  defaultEnvVars,
		DeployConfig:    models.JSONMap(req.DeployConfig),
		JenkinsTemplate: req.JenkinsTemplate,
		CreatedUserID:   userID,
		CreatedAt:       now,
		ModifiedAt:      now,
	}

	if err := s.dao.TemplateCreate(ctx, template); err != nil {
		return 0, fmt.Errorf("创建模板失败: %w", err)
	}

	return template.ID, nil
}

// TemplateDetail 获取模板详情
func (s *Services) TemplateDetail(ctx context.Context, id int64) (*models.TemplateDetailResponse, error) {
	template, err := s.dao.TemplateGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模板不存在")
		}
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	return template.ToDetailResponse(), nil
}

// TemplateList 获取模板列表
func (s *Services) TemplateList(ctx context.Context, req *requests.TemplateListRequest) ([]*models.TemplateListItem, int64, error) {
	list, total, err := s.dao.TemplateList(ctx, req.Keyword, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询模板列表失败: %w", err)
	}

	// 转换为列表项
	items := make([]*models.TemplateListItem, 0, len(list))
	for _, t := range list {
		items = append(items, t.ToTemplateListItem())
	}

	return items, total, nil
}

// TemplateUpdate 更新模板
func (s *Services) TemplateUpdate(ctx context.Context, req *requests.TemplateUpdateRequest) error {
	// 检查模板是否存在
	template, err := s.dao.TemplateGetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模板不存在")
		}
		return fmt.Errorf("查询模板失败: %w", err)
	}

	// 如果修改了名称，检查新名称是否已存在
	if req.Name != "" && req.Name != template.Name {
		_, err := s.dao.TemplateGetByName(ctx, req.Name)
		if err == nil {
			return errors.New("模板名称已存在")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("检查名称失败: %w", err)
		}
		template.Name = req.Name
	}

	// 更新字段
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Stages != nil {
		stagesBytes, _ := json.Marshal(req.Stages)
		json.Unmarshal(stagesBytes, &template.Stages)
	}
	if req.DefaultEnvVars != nil {
		envBytes, _ := json.Marshal(req.DefaultEnvVars)
		json.Unmarshal(envBytes, &template.DefaultEnvVars)
	}
	if req.DeployConfig != nil {
		template.DeployConfig = models.JSONMap(req.DeployConfig)
	}
	if req.JenkinsTemplate != "" {
		template.JenkinsTemplate = req.JenkinsTemplate
	}

	template.ModifiedAt = uint64(time.Now().Unix())

	return s.dao.TemplateUpdate(ctx, template)
}

// TemplateDelete 删除模板
func (s *Services) TemplateDelete(ctx context.Context, id int64) error {
	// 检查是否存在
	_, err := s.dao.TemplateGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模板不存在")
		}
		return fmt.Errorf("查询模板失败: %w", err)
	}

	return s.dao.TemplateDelete(ctx, id)
}

// TemplateUse 使用模板（增加使用次数）
func (s *Services) TemplateUse(ctx context.Context, id int64) error {
	return s.dao.TemplateIncrUsageCount(ctx, id)
}

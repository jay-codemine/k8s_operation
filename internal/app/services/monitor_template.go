package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8soperation/internal/app/models"
)

// ============================================================
// 通知模板 CRUD
// ============================================================

// NotifyTemplateListReq 模板列表请求
type NotifyTemplateListReq struct {
	Page  int    `form:"page" json:"page"`
	Size  int    `form:"size" json:"size"`
	Type  string `form:"type" json:"type"`
	Scene string `form:"scene" json:"scene"`
}

// NotifyTemplateListResp 模板列表响应
type NotifyTemplateListResp struct {
	Total int64                          `json:"total"`
	Items []models.MonitorNotifyTemplate `json:"items"`
}

// ListNotifyTemplates 列表
func (s *MonitorCRUDService) ListNotifyTemplates(ctx context.Context, req NotifyTemplateListReq) (*NotifyTemplateListResp, error) {
	db := s.db.WithContext(ctx).Where("is_del = 0")

	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Scene != "" {
		db = db.Where("scene = ?", req.Scene)
	}

	var total int64
	db.Model(&models.MonitorNotifyTemplate{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorNotifyTemplate
	err := db.Order("is_default DESC, id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &NotifyTemplateListResp{Total: total, Items: items}, nil
}

// GetNotifyTemplate 详情
func (s *MonitorCRUDService) GetNotifyTemplate(ctx context.Context, id int64) (*models.MonitorNotifyTemplate, error) {
	var tpl models.MonitorNotifyTemplate
	err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// CreateNotifyTemplate 创建
func (s *MonitorCRUDService) CreateNotifyTemplate(ctx context.Context, tpl *models.MonitorNotifyTemplate) error {
	return s.db.WithContext(ctx).Create(tpl).Error
}

// UpdateNotifyTemplate 更新
func (s *MonitorCRUDService) UpdateNotifyTemplate(ctx context.Context, tpl *models.MonitorNotifyTemplate) error {
	return s.db.WithContext(ctx).Model(tpl).
		Where("id = ? AND is_del = 0", tpl.ID).
		Updates(map[string]interface{}{
			"name":        tpl.Name,
			"type":        tpl.Type,
			"scene":       tpl.Scene,
			"title":       tpl.Title,
			"content":     tpl.Content,
			"description": tpl.Description,
			"is_default":  tpl.IsDefault,
			"enabled":     tpl.Enabled,
		}).Error
}

// DeleteNotifyTemplate 删除（软删除）
func (s *MonitorCRUDService) DeleteNotifyTemplate(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&models.MonitorNotifyTemplate{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// PreviewNotifyTemplate 预览模板渲染效果
func (s *MonitorCRUDService) PreviewNotifyTemplate(ctx context.Context, tpl *models.MonitorNotifyTemplate) string {
	// 用示例数据替换变量
	content := tpl.Content
	now := time.Now().Format("2006-01-02 15:04:05")

	replacements := map[string]string{
		"{{.RuleName}}":    "CPU 使用率过高告警",
		"{{.Severity}}":    "warning",
		"{{.SeverityText}}": "🟡 P1-Warning",
		"{{.Status}}":      "firing",
		"{{.StatusText}}":  "🔥 告警触发",
		"{{.Summary}}":     "节点 CPU 使用率超过 85%",
		"{{.Description}}": "生产环境 node-01 CPU 持续高负载，请及时处理",
		"{{.Value}}":       "85.6%",
		"{{.FiredAt}}":     now,
		"{{.ResolvedAt}}":  "",
		"{{.Duration}}":    "5m32s",
		"{{.Labels}}":      "env=production, instance=node-01:9100, cluster=k8s-prod",
		"{{.Labels.env}}":  "production",
		"{{.Labels.instance}}": "node-01:9100",
		"{{.Labels.cluster}}":  "k8s-prod",
		"{{.Platform}}":    "K8s Operation 监控平台",
	}

	for k, v := range replacements {
		content = strings.ReplaceAll(content, k, v)
	}

	return content
}

// SetDefaultTemplate 设置默认模板（同类型+场景只能有一个默认）
func (s *MonitorCRUDService) SetDefaultTemplate(ctx context.Context, id int64) error {
	var tpl models.MonitorNotifyTemplate
	if err := s.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&tpl).Error; err != nil {
		return fmt.Errorf("模板不存在")
	}

	// 取消同类型+场景的其他默认
	s.db.WithContext(ctx).Model(&models.MonitorNotifyTemplate{}).
		Where("type = ? AND scene = ? AND is_del = 0 AND id != ?", tpl.Type, tpl.Scene, id).
		Update("is_default", false)

	// 设为默认
	return s.db.WithContext(ctx).Model(&models.MonitorNotifyTemplate{}).
		Where("id = ?", id).Update("is_default", true).Error
}

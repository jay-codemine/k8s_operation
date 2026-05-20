package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ============================================================
// 静默规则 CRUD
// ============================================================

// SilenceRuleListReq 静默规则列表请求
type SilenceRuleListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Type    string `form:"type" json:"type"`
	Keyword string `form:"keyword" json:"keyword"`
	Status  string `form:"status" json:"status"` // active/expired/all
}

// SilenceRuleListResp 静默规则列表响应
type SilenceRuleListResp struct {
	Total int64                        `json:"total"`
	Items []models.MonitorSilenceRule  `json:"items"`
}

// ListSilenceRules 列表
func (s *MonitorCRUDService) ListSilenceRules(ctx context.Context, req SilenceRuleListReq) (*SilenceRuleListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR comment LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	now := time.Now().Unix()
	switch req.Status {
	case "active":
		db = db.Where("enabled = 1 AND (ends_at = 0 OR ends_at > ?)", now)
	case "expired":
		db = db.Where("ends_at > 0 AND ends_at <= ?", now)
	}

	var total int64
	db.Model(&models.MonitorSilenceRule{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorSilenceRule
	err := db.Order("id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &SilenceRuleListResp{Total: total, Items: items}, nil
}

// GetSilenceRule 详情
func (s *MonitorCRUDService) GetSilenceRule(ctx context.Context, id int64) (*models.MonitorSilenceRule, error) {
	var rule models.MonitorSilenceRule
	err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateSilenceRule 创建
func (s *MonitorCRUDService) CreateSilenceRule(ctx context.Context, rule *models.MonitorSilenceRule) error {
	return global.DB.WithContext(ctx).Create(rule).Error
}

// UpdateSilenceRule 更新
func (s *MonitorCRUDService) UpdateSilenceRule(ctx context.Context, rule *models.MonitorSilenceRule) error {
	return global.DB.WithContext(ctx).Model(rule).
		Where("id = ? AND is_del = 0", rule.ID).
		Updates(map[string]interface{}{
			"name":        rule.Name,
			"type":        rule.Type,
			"matchers":    rule.Matchers,
			"starts_at":   rule.StartsAt,
			"ends_at":     rule.EndsAt,
			"duration":    rule.Duration,
			"repeat_type": rule.RepeatType,
			"repeat_cron": rule.RepeatCron,
			"comment":     rule.Comment,
			"enabled":     rule.Enabled,
		}).Error
}

// DeleteSilenceRule 删除（软删除）
func (s *MonitorCRUDService) DeleteSilenceRule(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorSilenceRule{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// ============================================================
// 抑制规则 CRUD
// ============================================================

// InhibitRuleListReq 抑制规则列表请求
type InhibitRuleListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Keyword string `form:"keyword" json:"keyword"`
}

// InhibitRuleListResp 抑制规则列表响应
type InhibitRuleListResp struct {
	Total int64                         `json:"total"`
	Items []models.MonitorInhibitRule   `json:"items"`
}

// ListInhibitRules 列表
func (s *MonitorCRUDService) ListInhibitRules(ctx context.Context, req InhibitRuleListReq) (*InhibitRuleListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Model(&models.MonitorInhibitRule{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorInhibitRule
	err := db.Order("id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &InhibitRuleListResp{Total: total, Items: items}, nil
}

// CreateInhibitRule 创建
func (s *MonitorCRUDService) CreateInhibitRule(ctx context.Context, rule *models.MonitorInhibitRule) error {
	return global.DB.WithContext(ctx).Create(rule).Error
}

// UpdateInhibitRule 更新
func (s *MonitorCRUDService) UpdateInhibitRule(ctx context.Context, rule *models.MonitorInhibitRule) error {
	return global.DB.WithContext(ctx).Model(rule).
		Where("id = ? AND is_del = 0", rule.ID).
		Updates(map[string]interface{}{
			"name":            rule.Name,
			"source_matchers": rule.SourceMatchers,
			"target_matchers": rule.TargetMatchers,
			"equal_labels":    rule.EqualLabels,
			"description":     rule.Description,
			"enabled":         rule.Enabled,
		}).Error
}

// DeleteInhibitRule 删除
func (s *MonitorCRUDService) DeleteInhibitRule(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorInhibitRule{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// ============================================================
// 聚合规则 CRUD
// ============================================================

// AggregateRuleListReq 聚合规则列表请求
type AggregateRuleListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Keyword string `form:"keyword" json:"keyword"`
}

// AggregateRuleListResp 聚合规则列表响应
type AggregateRuleListResp struct {
	Total int64                           `json:"total"`
	Items []models.MonitorAggregateRule   `json:"items"`
}

// ListAggregateRules 列表
func (s *MonitorCRUDService) ListAggregateRules(ctx context.Context, req AggregateRuleListReq) (*AggregateRuleListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	var total int64
	db.Model(&models.MonitorAggregateRule{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorAggregateRule
	err := db.Order("id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &AggregateRuleListResp{Total: total, Items: items}, nil
}

// CreateAggregateRule 创建
func (s *MonitorCRUDService) CreateAggregateRule(ctx context.Context, rule *models.MonitorAggregateRule) error {
	return global.DB.WithContext(ctx).Create(rule).Error
}

// UpdateAggregateRule 更新
func (s *MonitorCRUDService) UpdateAggregateRule(ctx context.Context, rule *models.MonitorAggregateRule) error {
	return global.DB.WithContext(ctx).Model(rule).
		Where("id = ? AND is_del = 0", rule.ID).
		Updates(map[string]interface{}{
			"name":            rule.Name,
			"group_by":        rule.GroupBy,
			"group_wait":      rule.GroupWait,
			"group_interval":  rule.GroupInterval,
			"repeat_interval": rule.RepeatInterval,
			"matchers":        rule.Matchers,
			"channel_ids":     rule.ChannelIDs,
			"enabled":         rule.Enabled,
		}).Error
}

// DeleteAggregateRule 删除
func (s *MonitorCRUDService) DeleteAggregateRule(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorAggregateRule{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// ============================================================
// 告警匹配与静默判断
// ============================================================

// LabelMatcher 标签匹配器
type LabelMatcher struct {
	Label string `json:"label"`
	Op    string `json:"op"`    // = / != / =~ / !~
	Value string `json:"value"`
}

// IsAlertSilenced 判断告警是否被静默
func (s *MonitorCRUDService) IsAlertSilenced(ctx context.Context, labels map[string]string) (bool, string) {
	now := time.Now().Unix()

	var rules []models.MonitorSilenceRule
	global.DB.WithContext(ctx).
		Where("is_del = 0 AND enabled = 1 AND type = 'silence'").
		Where("(starts_at = 0 OR starts_at <= ?) AND (ends_at = 0 OR ends_at > ?)", now, now).
		Find(&rules)

	for _, rule := range rules {
		if matchLabels(rule.Matchers, labels) {
			return true, rule.Name
		}
	}
	return false, ""
}

// IsAlertInhibited 判断告警是否被抑制
func (s *MonitorCRUDService) IsAlertInhibited(ctx context.Context, labels map[string]string, severity string) (bool, string) {
	var rules []models.MonitorInhibitRule
	global.DB.WithContext(ctx).
		Where("is_del = 0 AND enabled = 1").
		Find(&rules)

	for _, rule := range rules {
		// 检查目标匹配
		if !matchLabels(rule.TargetMatchers, labels) {
			continue
		}

		// 检查是否存在对应的源告警正在 firing
		// 简化实现：检查更高级别的同标签告警是否活跃
		equalLabels := strings.Split(rule.EqualLabels, ",")
		if s.hasActiveFiringSource(ctx, rule.SourceMatchers, labels, equalLabels) {
			return true, rule.Name
		}
	}
	return false, ""
}

// hasActiveFiringSource 检查是否有活跃的源告警
func (s *MonitorCRUDService) hasActiveFiringSource(ctx context.Context, sourceMatchersJSON string, targetLabels map[string]string, equalLabels []string) bool {
	// 查找当前 firing 的告警事件，检查是否有匹配源条件的
	var events []models.MonitorAlertEvent
	global.DB.WithContext(ctx).
		Where("status = 'firing'").
		Limit(100).
		Find(&events)

	for _, event := range events {
		eventLabels := parseLabelsJSON(event.Labels)
		if !matchLabels(sourceMatchersJSON, eventLabels) {
			continue
		}
		// 检查 equalLabels 是否一致
		allEqual := true
		for _, el := range equalLabels {
			el = strings.TrimSpace(el)
			if el == "" {
				continue
			}
			if eventLabels[el] != targetLabels[el] {
				allEqual = false
				break
			}
		}
		if allEqual {
			return true
		}
	}
	return false
}

// matchLabels 检查标签是否匹配规则
func matchLabels(matchersJSON string, labels map[string]string) bool {
	if matchersJSON == "" || matchersJSON == "[]" {
		return true // 空匹配器匹配所有
	}

	var matchers []LabelMatcher
	if err := json.Unmarshal([]byte(matchersJSON), &matchers); err != nil {
		return false
	}

	for _, m := range matchers {
		val, exists := labels[m.Label]
		switch m.Op {
		case "=":
			if val != m.Value {
				return false
			}
		case "!=":
			if val == m.Value {
				return false
			}
		case "=~":
			if !strings.Contains(val, m.Value) {
				return false
			}
		case "!~":
			if strings.Contains(val, m.Value) {
				return false
			}
		case "exists":
			if !exists {
				return false
			}
		}
	}
	return true
}

// parseLabelsJSON 解析标签 JSON
func parseLabelsJSON(labelsStr string) map[string]string {
	labels := make(map[string]string)
	if labelsStr == "" {
		return labels
	}
	_ = json.Unmarshal([]byte(labelsStr), &labels)
	return labels
}

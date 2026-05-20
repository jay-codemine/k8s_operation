package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// MonitorCRUDService 监控 CRUD 服务
type MonitorCRUDService struct{}

func NewMonitorCRUDService() *MonitorCRUDService {
	return &MonitorCRUDService{}
}

// ============================================================
// 数据源 CRUD
// ============================================================

// DatasourceListReq 数据源列表查询
type DatasourceListReq struct {
	Page    int    `form:"page" json:"page"`
	Size    int    `form:"size" json:"size"`
	Type    string `form:"type" json:"type"`
	Keyword string `form:"keyword" json:"keyword"`
	Enabled *bool  `form:"enabled" json:"enabled"`
}

// DatasourceListResp 数据源列表响应
type DatasourceListResp struct {
	Total int64                     `json:"total"`
	Items []models.MonitorDatasource `json:"items"`
}

// ListDatasources 列表
func (s *MonitorCRUDService) ListDatasources(ctx context.Context, req DatasourceListReq) (*DatasourceListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Enabled != nil {
		db = db.Where("enabled = ?", *req.Enabled)
	}

	var total int64
	db.Model(&models.MonitorDatasource{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorDatasource
	err := db.Order("is_default DESC, id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &DatasourceListResp{Total: total, Items: items}, nil
}

// GetDatasource 详情
func (s *MonitorCRUDService) GetDatasource(ctx context.Context, id int64) (*models.MonitorDatasource, error) {
	var ds models.MonitorDatasource
	err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&ds).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

// CreateDatasource 创建
func (s *MonitorCRUDService) CreateDatasource(ctx context.Context, ds *models.MonitorDatasource) error {
	ds.Status = "unknown"
	return global.DB.WithContext(ctx).Create(ds).Error
}

// UpdateDatasource 更新
func (s *MonitorCRUDService) UpdateDatasource(ctx context.Context, ds *models.MonitorDatasource) error {
	return global.DB.WithContext(ctx).Model(ds).
		Where("id = ? AND is_del = 0", ds.ID).
		Updates(map[string]interface{}{
			"name":            ds.Name,
			"type":            ds.Type,
			"url":             ds.URL,
			"description":     ds.Description,
			"access_mode":     ds.AccessMode,
			"auth_type":       ds.AuthType,
			"auth_user":       ds.AuthUser,
			"auth_pass":       ds.AuthPass,
			"is_default":      ds.IsDefault,
			"enabled":         ds.Enabled,
			"timeout":         ds.Timeout,
			"scrape_interval": ds.ScrapeInterval,
		}).Error
}

// DeleteDatasource 删除（软删除）
func (s *MonitorCRUDService) DeleteDatasource(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorDatasource{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// TestDatasourceConnection 测试数据源连通性
func (s *MonitorCRUDService) TestDatasourceConnection(ctx context.Context, ds *models.MonitorDatasource) (bool, string) {
	client := &http.Client{Timeout: time.Duration(ds.Timeout) * time.Second}

	var testURL string
	switch ds.Type {
	case "prometheus":
		testURL = ds.URL + "/-/healthy"
	case "loki":
		testURL = ds.URL + "/ready"
	case "alertmanager":
		testURL = ds.URL + "/-/healthy"
	case "victoriametrics":
		testURL = ds.URL + "/health"
	case "n9e", "nightingale":
		testURL = ds.URL + "/api/n9e/heartbeat"
	case "thanos":
		testURL = ds.URL + "/-/healthy"
	default:
		testURL = ds.URL + "/api/health"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	if ds.AuthType == "bearer" && ds.AuthPass != "" {
		req.Header.Set("Authorization", "Bearer "+ds.AuthPass)
	} else if ds.AuthType == "basic" && ds.AuthUser != "" {
		req.SetBasicAuth(ds.AuthUser, ds.AuthPass)
	}

	resp, err := client.Do(req)
	if err != nil {
		// 更新状态
		s.updateDatasourceStatus(ctx, ds.ID, "disconnected")
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		s.updateDatasourceStatus(ctx, ds.ID, "connected")
		return true, "连接成功"
	}

	s.updateDatasourceStatus(ctx, ds.ID, "disconnected")
	return false, fmt.Sprintf("返回状态码: %d", resp.StatusCode)
}

func (s *MonitorCRUDService) updateDatasourceStatus(ctx context.Context, id int64, status string) {
	if id > 0 {
		global.DB.WithContext(ctx).Model(&models.MonitorDatasource{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":        status,
				"last_check_at": time.Now().Unix(),
			})
	}
}

// ============================================================
// 告警规则 CRUD
// ============================================================

// AlertRuleListReq 告警规则列表查询
type AlertRuleListReq struct {
	Page         int    `form:"page" json:"page"`
	Size         int    `form:"size" json:"size"`
	DatasourceID int64  `form:"datasource_id" json:"datasource_id"`
	Severity     string `form:"severity" json:"severity"`
	Group        string `form:"group" json:"group"`
	Keyword      string `form:"keyword" json:"keyword"`
	Enabled      *bool  `form:"enabled" json:"enabled"`
}

// AlertRuleListResp 告警规则列表响应
type AlertRuleListResp struct {
	Total int64                      `json:"total"`
	Items []models.MonitorAlertRule   `json:"items"`
}

// ListAlertRules 列表
func (s *MonitorCRUDService) ListAlertRules(ctx context.Context, req AlertRuleListReq) (*AlertRuleListResp, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0")

	if req.DatasourceID > 0 {
		db = db.Where("datasource_id = ?", req.DatasourceID)
	}
	if req.Severity != "" {
		db = db.Where("severity = ?", req.Severity)
	}
	if req.Group != "" {
		db = db.Where("`group` = ?", req.Group)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Enabled != nil {
		db = db.Where("enabled = ?", *req.Enabled)
	}

	var total int64
	db.Model(&models.MonitorAlertRule{}).Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorAlertRule
	err := db.Order("severity DESC, id DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &AlertRuleListResp{Total: total, Items: items}, nil
}

// GetAlertRule 详情
func (s *MonitorCRUDService) GetAlertRule(ctx context.Context, id int64) (*models.MonitorAlertRule, error) {
	var rule models.MonitorAlertRule
	err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateAlertRule 创建
func (s *MonitorCRUDService) CreateAlertRule(ctx context.Context, rule *models.MonitorAlertRule) error {
	return global.DB.WithContext(ctx).Create(rule).Error
}

// UpdateAlertRule 更新
func (s *MonitorCRUDService) UpdateAlertRule(ctx context.Context, rule *models.MonitorAlertRule) error {
	return global.DB.WithContext(ctx).Model(rule).
		Where("id = ? AND is_del = 0", rule.ID).
		Updates(map[string]interface{}{
			"datasource_id":   rule.DatasourceID,
			"name":            rule.Name,
			"group":           rule.Group,
			"severity":        rule.Severity,
			"expr":            rule.Expr,
			"duration":        rule.Duration,
			"summary":         rule.Summary,
			"description":     rule.Description,
			"labels":          rule.Labels,
			"annotations":     rule.Annotations,
			"enabled":         rule.Enabled,
			"notify_channels": rule.NotifyChannels,
			"notify_url":      rule.NotifyURL,
			"eval_interval":   rule.EvalInterval,
		}).Error
}

// DeleteAlertRule 删除
func (s *MonitorCRUDService) DeleteAlertRule(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
		Where("id = ? AND is_del = 0", id).
		Update("is_del", 1).Error
}

// ToggleAlertRule 启用/禁用告警规则
func (s *MonitorCRUDService) ToggleAlertRule(ctx context.Context, id int64, enabled bool) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
		Where("id = ? AND is_del = 0", id).
		Update("enabled", enabled).Error
}

// ============================================================
// 告警事件查询
// ============================================================

// AlertEventListReq 告警事件列表查询
type AlertEventListReq struct {
	Page         int    `form:"page" json:"page"`
	Size         int    `form:"size" json:"size"`
	RuleID       int64  `form:"rule_id" json:"rule_id"`
	DatasourceID int64  `form:"datasource_id" json:"datasource_id"`
	Severity     string `form:"severity" json:"severity"`
	Status       string `form:"status" json:"status"`
	StartTime    int64  `form:"start_time" json:"start_time"`
	EndTime      int64  `form:"end_time" json:"end_time"`
	Keyword      string `form:"keyword" json:"keyword"`
}

// AlertEventListResp 告警事件列表响应
type AlertEventListResp struct {
	Total int64                       `json:"total"`
	Items []models.MonitorAlertEvent   `json:"items"`
}

// ListAlertEvents 列表
func (s *MonitorCRUDService) ListAlertEvents(ctx context.Context, req AlertEventListReq) (*AlertEventListResp, error) {
	db := global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{})

	if req.RuleID > 0 {
		db = db.Where("rule_id = ?", req.RuleID)
	}
	if req.DatasourceID > 0 {
		db = db.Where("datasource_id = ?", req.DatasourceID)
	}
	if req.Severity != "" {
		db = db.Where("severity = ?", req.Severity)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.StartTime > 0 {
		db = db.Where("fired_at >= ?", req.StartTime)
	}
	if req.EndTime > 0 {
		db = db.Where("fired_at <= ?", req.EndTime)
	}
	if req.Keyword != "" {
		db = db.Where("rule_name LIKE ? OR summary LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	var items []models.MonitorAlertEvent
	err := db.Order("fired_at DESC").
		Offset((req.Page - 1) * req.Size).Limit(req.Size).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &AlertEventListResp{Total: total, Items: items}, nil
}

// GetAlertEvent 详情
func (s *MonitorCRUDService) GetAlertEvent(ctx context.Context, id int64) (*models.MonitorAlertEvent, error) {
	var event models.MonitorAlertEvent
	err := global.DB.WithContext(ctx).Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// AckAlertEvent 确认告警
func (s *MonitorCRUDService) AckAlertEvent(ctx context.Context, id int64, userID int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{}).
		Where("id = ? AND acked_by = 0", id).
		Updates(map[string]interface{}{
			"acked_by": userID,
			"acked_at": time.Now().Unix(),
		}).Error
}

// ResolveAlertEvent 手动解决告警
func (s *MonitorCRUDService) ResolveAlertEvent(ctx context.Context, id int64) error {
	return global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{}).
		Where("id = ? AND status = 'firing'", id).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": time.Now().Unix(),
		}).Error
}

// GetAlertStats 告警统计
type AlertStats struct {
	TotalFiring   int64 `json:"total_firing"`
	TotalResolved int64 `json:"total_resolved"`
	Critical      int64 `json:"critical"`
	Warning       int64 `json:"warning"`
	Info          int64 `json:"info"`
}

func (s *MonitorCRUDService) GetAlertStats(ctx context.Context) (*AlertStats, error) {
	stats := &AlertStats{}
	db := global.DB.WithContext(ctx).Model(&models.MonitorAlertEvent{})

	db.Where("status = 'firing'").Count(&stats.TotalFiring)
	db.Where("status = 'resolved'").Count(&stats.TotalResolved)
	db.Where("status = 'firing' AND severity = 'critical'").Count(&stats.Critical)
	db.Where("status = 'firing' AND severity = 'warning'").Count(&stats.Warning)
	db.Where("status = 'firing' AND severity = 'info'").Count(&stats.Info)

	return stats, nil
}

// GetAlertRuleGroups 获取告警规则分组列表
func (s *MonitorCRUDService) GetAlertRuleGroups(ctx context.Context) ([]string, error) {
	var groups []string
	err := global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
		Where("is_del = 0").
		Distinct("`group`").Pluck("`group`", &groups).Error
	return groups, err
}

// ignore unused import
var _ = gorm.ErrRecordNotFound

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// 用于检测并渲染告警摘要中未渲染的 Prometheus 模板变量
var (
	crudLabelsVarRegex = regexp.MustCompile(`\{\{\s*\$labels\.(\w+)\s*\}\}`)
	crudValueVarRegex  = regexp.MustCompile(`\{\{\s*\$value\s*\}\}`)
)

// ============================================================
// YAML 批量导入/导出 告警规则
// ============================================================

// AlertRuleYAMLSpec YAML 格式的告警规则（兼容 PrometheusRule 风格）
type AlertRuleYAMLSpec struct {
	Groups []AlertRuleGroup `yaml:"groups" json:"groups"`
}

// AlertRuleGroup 规则组（对应 PrometheusRule.spec.groups）
type AlertRuleGroup struct {
	Name  string          `yaml:"name" json:"name"`
	Rules []AlertRuleItem `yaml:"rules" json:"rules"`
}

// AlertRuleItem 单条规则项
type AlertRuleItem struct {
	Alert       string            `yaml:"alert" json:"alert"`
	Expr        string            `yaml:"expr" json:"expr"`
	For         string            `yaml:"for,omitempty" json:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// AlertRuleImportReq 批量导入请求
type AlertRuleImportReq struct {
	YAML                  string            `json:"yaml" binding:"required"`           // YAML 内容
	DatasourceID          int64             `json:"datasource_id" binding:"required"`  // 绑定数据源
	Overwrite             bool              `json:"overwrite"`                         // 同名规则是否覆盖
	DefaultNotifyChannels string            `json:"default_notify_channels"`           // 默认通知渠道ID(逗号分隔),导入时自动绑定
	GroupChannels         map[string]string `json:"group_channels"`                    // 按组指定通知渠道 {"infra":"1,2", "app":"3"}
	AutoRoute             bool              `json:"auto_route"`                        // 是否启用路由策略自动匹配
	// 渠道优先级: 规则annotations.notify_channels > group_channels[组名] > default_notify_channels > auto_route
}

// AlertRuleImportResult 导入结果
type AlertRuleImportResult struct {
	Total    int      `json:"total"`     // 总规则数
	Created  int      `json:"created"`   // 新建数
	Updated  int      `json:"updated"`   // 更新数（覆盖模式）
	Skipped  int      `json:"skipped"`   // 跳过数（已存在且不覆盖）
	Failed   int      `json:"failed"`    // 失败数
	Errors   []string `json:"errors"`    // 失败详情
}

// ImportAlertRulesFromYAML 从 YAML 批量导入告警规则
func (s *MonitorCRUDService) ImportAlertRulesFromYAML(ctx context.Context, req AlertRuleImportReq) (*AlertRuleImportResult, error) {
	var spec AlertRuleYAMLSpec
	if err := yaml.Unmarshal([]byte(req.YAML), &spec); err != nil {
		return nil, fmt.Errorf("YAML 解析失败: %w", err)
	}

	// 预加载路由策略（auto_route 模式下用于自动匹配渠道）
	var routePolicies []models.MonitorNotifyRoutePolicy
	if req.AutoRoute {
		global.DB.WithContext(ctx).
			Where("enabled = 1 AND is_del = 0").
			Order("priority ASC, id ASC").
			Find(&routePolicies)
	}

	result := &AlertRuleImportResult{}

	for _, group := range spec.Groups {
		for _, item := range group.Rules {
			result.Total++

			if item.Alert == "" || item.Expr == "" {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("[%s] alert 或 expr 为空", item.Alert))
				continue
			}

			// 提取 severity
			severity := "warning"
			if s, ok := item.Labels["severity"]; ok {
				severity = s
			}

			// 构建标签 JSON（排除 severity）
			labelsMap := make(map[string]string)
			for k, v := range item.Labels {
				if k != "severity" {
					labelsMap[k] = v
				}
			}
			labelsJSON, _ := json.Marshal(labelsMap)

			// 构建 annotations JSON
			annotationsJSON, _ := json.Marshal(item.Annotations)

			// 提取 summary / description
			summary := item.Annotations["summary"]
			description := item.Annotations["description"]

			// 决定通知渠道（优先级：规则 annotation > 组级映射 > 全局默认 > 路由策略）
			notifyChannels := ""
			
			// 1. 最高优先级：规则自身 annotations 中指定 notify_channels
			if ch, ok := item.Annotations["notify_channels"]; ok && ch != "" {
				notifyChannels = ch
			}
			// 2. 按组指定渠道
			if notifyChannels == "" && len(req.GroupChannels) > 0 {
				if ch, ok := req.GroupChannels[group.Name]; ok && ch != "" {
					notifyChannels = ch
				}
			}
			// 3. 全局默认渠道
			if notifyChannels == "" && req.DefaultNotifyChannels != "" {
				notifyChannels = req.DefaultNotifyChannels
			}
			// 4. 路由策略自动匹配
			if notifyChannels == "" && req.AutoRoute && len(routePolicies) > 0 {
				notifyChannels = s.matchRoutePolicy(routePolicies, severity, group.Name, labelsMap)
			}

			// 检查是否已存在同名规则
			var existing models.MonitorAlertRule
			err := global.DB.WithContext(ctx).
				Where("name = ? AND is_del = 0", item.Alert).
				First(&existing).Error

			if err == nil {
				// 已存在
				if !req.Overwrite {
					result.Skipped++
					continue
				}
				// 覆盖更新
				existing.Group = group.Name
				existing.Severity = severity
				existing.Expr = item.Expr
				existing.Duration = item.For
				existing.Summary = summary
				existing.Description = description
				existing.Labels = string(labelsJSON)
				existing.Annotations = string(annotationsJSON)
				existing.DatasourceID = req.DatasourceID
				// 覆盖模式下也更新通知渠道（如果指定了）
				if notifyChannels != "" {
					existing.NotifyChannels = notifyChannels
				}
				if err := s.UpdateAlertRule(ctx, &existing); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("[%s] 更新失败: %v", item.Alert, err))
				} else {
					result.Updated++
				}
			} else {
				// 新建
				duration := item.For
				if duration == "" {
					duration = "5m"
				}
				rule := &models.MonitorAlertRule{
					DatasourceID:   req.DatasourceID,
					Name:           item.Alert,
					Group:          group.Name,
					Severity:       severity,
					Expr:           item.Expr,
					Duration:       duration,
					Summary:        summary,
					Description:    description,
					Labels:         string(labelsJSON),
					Annotations:    string(annotationsJSON),
					Enabled:        true,
					EvalInterval:   60,
					NotifyChannels: notifyChannels,
				}
				if err := s.CreateAlertRule(ctx, rule); err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("[%s] 创建失败: %v", item.Alert, err))
				} else {
					result.Created++
				}
			}
		}
	}

	return result, nil
}

// matchRoutePolicy 根据路由策略匹配通知渠道（优先级从高到低）
// 优先匹配有明确条件的策略；若均不匹配，回退到 IsDefault=true 的兜底策略
func (s *MonitorCRUDService) matchRoutePolicy(policies []models.MonitorNotifyRoutePolicy, severity, group string, labels map[string]string) string {
	for _, policy := range policies {
		if policy.IsDefault {
			continue // 兜底策略最后匹配
		}
		if s.policyMatches(&policy, severity, group, labels) {
			return policy.ChannelIDs
		}
	}
	// 回退到兜底默认策略
	for _, policy := range policies {
		if policy.IsDefault {
			return policy.ChannelIDs
		}
	}
	return ""
}

// policyMatches 检查告警是否匹配路由策略
func (s *MonitorCRUDService) policyMatches(policy *models.MonitorNotifyRoutePolicy, severity, group string, labels map[string]string) bool {
	matchAll := policy.MatchMode == "all"
	hasCondition := false
	matchCount := 0

	// 检查 severity 匹配
	if policy.Severities != "" {
		hasCondition = true
		severities := strings.Split(policy.Severities, ",")
		for _, s := range severities {
			if strings.TrimSpace(s) == severity {
				matchCount++
				break
			}
		}
		if matchAll && matchCount == 0 {
			return false
		}
	}

	// 检查 group 匹配
	if policy.Groups != "" {
		hasCondition = true
		groups := strings.Split(policy.Groups, ",")
		matched := false
		for _, g := range groups {
			if strings.TrimSpace(g) == group {
				matched = true
				break
			}
		}
		if matched {
			matchCount++
		}
		if matchAll && !matched {
			return false
		}
	}

	// 检查标签匹配
	if policy.LabelMatch != "" {
		hasCondition = true
		var matchers []struct {
			Key   string `json:"key"`
			Op    string `json:"op"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(policy.LabelMatch), &matchers); err == nil && len(matchers) > 0 {
			labelMatched := true
			for _, m := range matchers {
				v, exists := labels[m.Key]
				switch m.Op {
				case "=", "==":
					if v != m.Value {
						labelMatched = false
					}
				case "!=":
					if v == m.Value {
						labelMatched = false
					}
				case "=~":
					if !strings.Contains(v, m.Value) {
						labelMatched = false
					}
				case "exists":
					if !exists {
						labelMatched = false
					}
				}
				if !labelMatched {
					break
				}
			}
			if labelMatched {
				matchCount++
			}
			if matchAll && !labelMatched {
				return false
			}
		}
	}

	if !hasCondition {
		return false
	}

	if matchAll {
		return true
	}
	// any 模式：至少一个条件匹配
	return matchCount > 0
}

// ExportAlertRulesToYAML 导出告警规则为 PrometheusRule 兼容 YAML
func (s *MonitorCRUDService) ExportAlertRulesToYAML(ctx context.Context, group string, ids []int64) (string, error) {
	db := global.DB.WithContext(ctx).Where("is_del = 0 AND enabled = 1")

	if group != "" {
		db = db.Where("`group` = ?", group)
	}
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}

	var rules []models.MonitorAlertRule
	if err := db.Order("`group` ASC, id ASC").Find(&rules).Error; err != nil {
		return "", err
	}

	if len(rules) == 0 {
		return "", fmt.Errorf("没有可导出的告警规则")
	}

	// 按 group 分组
	groupMap := make(map[string][]AlertRuleItem)
	groupOrder := make([]string, 0)
	for _, rule := range rules {
		grp := rule.Group
		if grp == "" {
			grp = "default"
		}
		if _, exists := groupMap[grp]; !exists {
			groupOrder = append(groupOrder, grp)
		}

		item := AlertRuleItem{
			Alert: rule.Name,
			Expr:  rule.Expr,
			For:   rule.Duration,
		}

		// 构建 labels
		item.Labels = make(map[string]string)
		item.Labels["severity"] = rule.Severity
		if rule.Labels != "" {
			var extraLabels map[string]string
			if json.Unmarshal([]byte(rule.Labels), &extraLabels) == nil {
				for k, v := range extraLabels {
					item.Labels[k] = v
				}
			}
		}

		// 构建 annotations
		item.Annotations = make(map[string]string)
		if rule.Summary != "" {
			item.Annotations["summary"] = rule.Summary
		}
		if rule.Description != "" {
			item.Annotations["description"] = rule.Description
		}
		if rule.Annotations != "" {
			var extraAnnotations map[string]string
			if json.Unmarshal([]byte(rule.Annotations), &extraAnnotations) == nil {
				for k, v := range extraAnnotations {
					if k != "summary" && k != "description" {
						item.Annotations[k] = v
					}
				}
			}
		}

		groupMap[grp] = append(groupMap[grp], item)
	}

	// 构建 YAML 结构
	spec := AlertRuleYAMLSpec{}
	for _, grp := range groupOrder {
		spec.Groups = append(spec.Groups, AlertRuleGroup{
			Name:  grp,
			Rules: groupMap[grp],
		})
	}

	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("YAML 序列化失败: %w", err)
	}

	// 添加 PrometheusRule 头部信息
	var sb strings.Builder
	sb.WriteString("# =========================================================\n")
	sb.WriteString("# K8s Operation 监控平台 - 告警规则导出\n")
	sb.WriteString(fmt.Sprintf("# 导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("# 规则总数: %d\n", len(rules)))
	sb.WriteString("# =========================================================\n")
	sb.WriteString("# 可直接作为 PrometheusRule CR 的 spec 部分使用\n")
	sb.WriteString("# 也可通过平台「批量导入」功能重新导入\n")
	sb.WriteString("# =========================================================\n\n")
	sb.Write(yamlData)

	return sb.String(), nil
}

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
	Page      int    `form:"page" json:"page"`
	Size      int    `form:"size" json:"size"`
	Type      string `form:"type" json:"type"`
	Keyword   string `form:"keyword" json:"keyword"`
	Enabled   *bool  `form:"enabled" json:"enabled"`
	ClusterID *int64 `form:"cluster_id" json:"cluster_id"` // 可选：按集群过滤（0=全局数据源，不传=不过滤）
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
	if req.ClusterID != nil {
		db = db.Where("cluster_id = ?", *req.ClusterID)
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
	db := global.DB.WithContext(ctx)
	// 若新增的是默认，先把同 type 的其他默认取消，保证同 type 只有一条 is_default=1
	if ds.IsDefault {
		if err := db.Model(&models.MonitorDatasource{}).
			Where("type = ? AND is_del = 0", ds.Type).
			Update("is_default", false).Error; err != nil {
			return err
		}
	}
	return db.Create(ds).Error
}

// UpdateDatasource 更新
func (s *MonitorCRUDService) UpdateDatasource(ctx context.Context, ds *models.MonitorDatasource) error {
	db := global.DB.WithContext(ctx)
	// 若本次更新把它设为默认，先把同 type 内其他记录的 is_default 全部置 false
	if ds.IsDefault {
		if err := db.Model(&models.MonitorDatasource{}).
			Where("type = ? AND id <> ? AND is_del = 0", ds.Type, ds.ID).
			Update("is_default", false).Error; err != nil {
			return err
		}
	}
	return db.Model(ds).
		Where("id = ? AND is_del = 0", ds.ID).
		Updates(map[string]interface{}{
			"name":            ds.Name,
			"type":            ds.Type,
			"url":             ds.URL,
			"description":     ds.Description,
			"cluster_id":      ds.ClusterID,
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
// 批量操作
// ============================================================

// BatchBindChannelsReq 批量绑定通知渠道请求
// 支持两种方式：
//  1. 指定 rule_ids → 精确绑定
//  2. 使用筛选条件(group/severity/keyword/datasource_id) → 按条件匹配后绑定
type BatchBindChannelsReq struct {
	RuleIDs        []int64 `json:"rule_ids"`                             // 精确指定规则 ID 列表（优先使用）
	NotifyChannels string  `json:"notify_channels" binding:"required"`   // 通知渠道 ID(逗号分隔)
	Mode           string  `json:"mode"`                                 // 绑定模式: replace(替换)/append(追加)/remove(移除)

	// ========= 条件匹配（当 rule_ids 为空时生效） =========
	Group        string `json:"group"`         // 按分组匹配（精确匹配）
	Severity     string `json:"severity"`      // 按级别匹配: critical/warning/info
	Keyword      string `json:"keyword"`       // 按规则名称模糊匹配
	DatasourceID int64  `json:"datasource_id"` // 按数据源匹配
	MatchAll     bool   `json:"match_all"`     // 是否匹配所有启用的规则（慎用）
}

// BatchBindChannelsResult 批量绑定结果
type BatchBindChannelsResult struct {
	Total   int    `json:"total"`   // 处理总数
	Success int    `json:"success"` // 成功数
	Failed  int    `json:"failed"`  // 失败数
	Matched int    `json:"matched"` // 条件匹配到的规则数（仅条件模式返回）
	Filter  string `json:"filter"`  // 使用的筛选条件描述
}

// BatchBindChannels 批量绑定/追加/移除通知渠道
// 支持按 rule_ids 精确绑定或按 group/severity/keyword 条件匹配后绑定
func (s *MonitorCRUDService) BatchBindChannels(ctx context.Context, req BatchBindChannelsReq) (*BatchBindChannelsResult, error) {
	// 如果没有指定 rule_ids，通过条件匹配查找规则
	if len(req.RuleIDs) == 0 {
		matchedIDs, filterDesc, err := s.resolveRuleIDsByFilter(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(matchedIDs) == 0 {
			return &BatchBindChannelsResult{Filter: filterDesc}, fmt.Errorf("未匹配到任何告警规则，请检查筛选条件")
		}
		req.RuleIDs = matchedIDs

		// 执行绑定并返回带匹配信息的结果
		result, err := s.executeBatchBind(ctx, req)
		if err != nil {
			return nil, err
		}
		result.Matched = len(matchedIDs)
		result.Filter = filterDesc
		return result, nil
	}

	if len(req.RuleIDs) > 500 {
		return nil, fmt.Errorf("单次最多操作 500 条规则")
	}

	return s.executeBatchBind(ctx, req)
}

// resolveRuleIDsByFilter 按条件匹配告警规则 ID
func (s *MonitorCRUDService) resolveRuleIDsByFilter(ctx context.Context, req BatchBindChannelsReq) ([]int64, string, error) {
	if !req.MatchAll && req.Group == "" && req.Severity == "" && req.Keyword == "" && req.DatasourceID == 0 {
		return nil, "", fmt.Errorf("请指定 rule_ids 或至少一个筛选条件(group/severity/keyword/datasource_id/match_all)")
	}

	db := global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).Where("is_del = 0 AND enabled = 1")

	var conditions []string

	if req.Group != "" {
		db = db.Where("`group` = ?", req.Group)
		conditions = append(conditions, fmt.Sprintf("group=%s", req.Group))
	}
	if req.Severity != "" {
		db = db.Where("severity = ?", req.Severity)
		conditions = append(conditions, fmt.Sprintf("severity=%s", req.Severity))
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
		conditions = append(conditions, fmt.Sprintf("keyword=%s", req.Keyword))
	}
	if req.DatasourceID > 0 {
		db = db.Where("datasource_id = ?", req.DatasourceID)
		conditions = append(conditions, fmt.Sprintf("datasource_id=%d", req.DatasourceID))
	}
	if req.MatchAll {
		conditions = append(conditions, "match_all=true")
	}

	var ruleIDs []int64
	if err := db.Pluck("id", &ruleIDs).Error; err != nil {
		return nil, "", err
	}

	// 安全限制：匹配数量上限
	if len(ruleIDs) > 500 {
		return nil, "", fmt.Errorf("匹配到 %d 条规则，超过单次上限 500，请缩小筛选范围", len(ruleIDs))
	}

	filterDesc := strings.Join(conditions, " AND ")
	return ruleIDs, filterDesc, nil
}

// executeBatchBind 执行实际的批量绑定操作
func (s *MonitorCRUDService) executeBatchBind(ctx context.Context, req BatchBindChannelsReq) (*BatchBindChannelsResult, error) {
	mode := req.Mode
	if mode == "" {
		mode = "replace"
	}

	result := &BatchBindChannelsResult{Total: len(req.RuleIDs)}

	switch mode {
	case "replace":
		// 直接批量更新
		tx := global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
			Where("id IN ? AND is_del = 0", req.RuleIDs).
			Update("notify_channels", req.NotifyChannels)
		if tx.Error != nil {
			return nil, tx.Error
		}
		result.Success = int(tx.RowsAffected)
		result.Failed = result.Total - result.Success

	case "append":
		// 逐条追加（去重）
		for _, ruleID := range req.RuleIDs {
			var rule models.MonitorAlertRule
			if err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", ruleID).First(&rule).Error; err != nil {
				result.Failed++
				continue
			}
			merged := mergeChannelIDs(rule.NotifyChannels, req.NotifyChannels)
			if err := global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
				Where("id = ?", ruleID).Update("notify_channels", merged).Error; err != nil {
				result.Failed++
			} else {
				result.Success++
			}
		}

	case "remove":
		// 逐条移除
		for _, ruleID := range req.RuleIDs {
			var rule models.MonitorAlertRule
			if err := global.DB.WithContext(ctx).Where("id = ? AND is_del = 0", ruleID).First(&rule).Error; err != nil {
				result.Failed++
				continue
			}
			cleaned := removeChannelIDs(rule.NotifyChannels, req.NotifyChannels)
			if err := global.DB.WithContext(ctx).Model(&models.MonitorAlertRule{}).
				Where("id = ?", ruleID).Update("notify_channels", cleaned).Error; err != nil {
				result.Failed++
			} else {
				result.Success++
			}
		}

	default:
		return nil, fmt.Errorf("不支持的绑定模式: %s (可选: replace/append/remove)", mode)
	}

	return result, nil
}

// mergeChannelIDs 合并通知渠道 ID（去重）
func mergeChannelIDs(existing, incoming string) string {
	idSet := make(map[string]bool)
	var result []string

	for _, id := range strings.Split(existing, ",") {
		id = strings.TrimSpace(id)
		if id != "" && !idSet[id] {
			idSet[id] = true
			result = append(result, id)
		}
	}
	for _, id := range strings.Split(incoming, ",") {
		id = strings.TrimSpace(id)
		if id != "" && !idSet[id] {
			idSet[id] = true
			result = append(result, id)
		}
	}
	return strings.Join(result, ",")
}

// removeChannelIDs 从现有渠道中移除指定 ID
func removeChannelIDs(existing, toRemove string) string {
	removeSet := make(map[string]bool)
	for _, id := range strings.Split(toRemove, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			removeSet[id] = true
		}
	}

	var result []string
	for _, id := range strings.Split(existing, ",") {
		id = strings.TrimSpace(id)
		if id != "" && !removeSet[id] {
			result = append(result, id)
		}
	}
	return strings.Join(result, ",")
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

	// 后处理：对含未渲染模板变量的旧事件补充渲染
	for i := range items {
		items[i].Summary = renderEventTemplate(items[i].Summary, items[i].Labels, items[i].Value)
		items[i].Description = renderEventTemplate(items[i].Description, items[i].Labels, items[i].Value)
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
	// 补渲染未解析的模板变量
	event.Summary = renderEventTemplate(event.Summary, event.Labels, event.Value)
	event.Description = renderEventTemplate(event.Description, event.Labels, event.Value)
	return &event, nil
}

// renderEventTemplate 对告警事件的 summary/description 做模板变量后处理渲染
// 用于历史旧事件中含有 {{ $labels.xxx }} / {{ $value }} 未被渲染的情况
func renderEventTemplate(tpl string, labelsJSON string, value string) string {
	if tpl == "" {
		return tpl
	}
	// 快速检查是否包含未渲染的模板变量
	if !strings.Contains(tpl, "{{") {
		return tpl
	}

	// 解析事件存储的 labels JSON
	labels := make(map[string]string)
	if labelsJSON != "" {
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
	}

	// 替换 {{ $value }}
	result := crudValueVarRegex.ReplaceAllString(tpl, value)

	// 替换 {{ $labels.xxx }}
	result = crudLabelsVarRegex.ReplaceAllStringFunc(result, func(match string) string {
		submatch := crudLabelsVarRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		labelName := submatch[1]
		if v, ok := labels[labelName]; ok {
			return v
		}
		return "<" + labelName + ":未知>"
	})

	return result
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

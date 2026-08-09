package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"k8soperation/internal/domain/monitor"
)

type monitorRepo struct{ db *gorm.DB }

func NewMonitorRepository(db *gorm.DB) monitor.MonitorRepository { return &monitorRepo{db: db} }

// ——— Datasource ———

func (r *monitorRepo) DatasourceSave(ctx context.Context, ds *monitor.Datasource) error {
	return r.db.WithContext(ctx).Create(ds).Error
}
func (r *monitorRepo) DatasourceFindByID(ctx context.Context, id int64) (*monitor.Datasource, error) {
	var ds monitor.Datasource
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}
func (r *monitorRepo) DatasourceUpdate(ctx context.Context, ds *monitor.Datasource) error {
	return r.db.WithContext(ctx).Model(ds).Where("id = ?", ds.ID).Updates(map[string]interface{}{
		"name": ds.Name, "url": ds.URL, "type": ds.Type, "description": ds.Description,
		"access_mode": ds.AccessMode, "auth_type": ds.AuthType, "is_default": ds.IsDefault,
		"enabled": ds.Enabled, "timeout": ds.Timeout, "scrape_interval": ds.ScrapeInterval,
		"updated_at": time.Now(),
	}).Error
}
func (r *monitorRepo) DatasourceDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&monitor.Datasource{}).Where("id = ?", id).Update("is_del", 1).Error
}
func (r *monitorRepo) DatasourceQuery(ctx context.Context, req monitor.DatasourceListReq) (*monitor.DatasourceListResp, error) {
	db := r.db.WithContext(ctx).Model(&monitor.Datasource{}).Where("is_del = 0")
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR url LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []monitor.Datasource
	page, size := req.Page, req.Size
	if page <= 0 { page = 1 }
	if size <= 0 { size = 20 }
	if err := db.Order("id DESC").Offset((page-1)*size).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}
	return &monitor.DatasourceListResp{Total: total, Items: items}, nil
}
func (r *monitorRepo) DatasourceFindDefault(ctx context.Context, types []string) (*monitor.Datasource, error) {
	var ds monitor.Datasource
	if err := r.db.WithContext(ctx).Where("type IN ? AND is_default = 1 AND enabled = 1 AND is_del = 0", types).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

// ——— AlertRule ———

func (r *monitorRepo) AlertRuleSave(ctx context.Context, rule *monitor.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}
func (r *monitorRepo) AlertRuleFindByID(ctx context.Context, id int64) (*monitor.AlertRule, error) {
	var rule monitor.AlertRule
	if err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
func (r *monitorRepo) AlertRuleUpdate(ctx context.Context, rule *monitor.AlertRule) error {
	return r.db.WithContext(ctx).Model(rule).Where("id = ?", rule.ID).Select("*").Updates(rule).Error
}
func (r *monitorRepo) AlertRuleDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&monitor.AlertRule{}).Where("id = ?", id).Update("is_del", 1).Error
}
func (r *monitorRepo) AlertRuleToggle(ctx context.Context, id int64, enabled bool) error {
	return r.db.WithContext(ctx).Model(&monitor.AlertRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}
func (r *monitorRepo) AlertRuleQuery(ctx context.Context, req monitor.AlertRuleListReq) (*monitor.AlertRuleListResp, error) {
	db := r.db.WithContext(ctx).Model(&monitor.AlertRule{}).Where("is_del = 0")
	if req.DatasourceID > 0 { db = db.Where("datasource_id = ?", req.DatasourceID) }
	if req.Severity != "" { db = db.Where("severity = ?", req.Severity) }
	if req.Keyword != "" { db = db.Where("name LIKE ?", "%"+req.Keyword+"%") }
	if req.Enabled != nil { db = db.Where("enabled = ?", *req.Enabled) }
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, err }
	var items []monitor.AlertRule
	page, size := req.Page, req.Size
	if page <= 0 { page = 1 }
	if size <= 0 { size = 20 }
	if err := db.Order("id DESC").Offset((page-1)*size).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}
	return &monitor.AlertRuleListResp{Total: total, Items: items}, nil
}
func (r *monitorRepo) AlertRuleGroups(ctx context.Context) ([]string, error) {
	var groups []string
	err := r.db.WithContext(ctx).Model(&monitor.AlertRule{}).Where("is_del = 0").Distinct("`group`").Pluck("`group`", &groups).Error
	return groups, err
}

// ——— AlertEvent ———

func (r *monitorRepo) AlertEventFindByID(ctx context.Context, id int64) (*monitor.AlertEvent, error) {
	var e monitor.AlertEvent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, fmt.Errorf("告警事件不存在: %w", err)
	}
	return &e, nil
}
func (r *monitorRepo) AlertEventQuery(ctx context.Context, req monitor.AlertEventListReq) (*monitor.AlertEventListResp, error) {
	db := r.db.WithContext(ctx).Model(&monitor.AlertEvent{})
	if req.Status != "" { db = db.Where("status = ?", req.Status) }
	if req.Severity != "" { db = db.Where("severity = ?", req.Severity) }
	if req.RuleID > 0 { db = db.Where("rule_id = ?", req.RuleID) }
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, err }
	var items []monitor.AlertEvent
	page, size := req.Page, req.Size
	if page <= 0 { page = 1 }
	if size <= 0 { size = 20 }
	if err := db.Order("fired_at DESC").Offset((page-1)*size).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}
	return &monitor.AlertEventListResp{Total: total, Items: items}, nil
}
func (r *monitorRepo) AlertEventAck(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": "acknowledged", "ack_by": userID, "ack_at": time.Now(),
	}).Error
}
func (r *monitorRepo) AlertEventResolve(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": "resolved", "resolved_at": time.Now(),
	}).Error
}
func (r *monitorRepo) AlertEventStats(ctx context.Context) (*monitor.AlertStats, error) {
	stats := &monitor.AlertStats{}
	r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("status = 'firing'").Count(&stats.TotalFiring)
	r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("status = 'resolved'").Count(&stats.TotalResolved)
	r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("status = 'firing' AND severity = 'critical'").Count(&stats.Critical)
	r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("status = 'firing' AND severity = 'warning'").Count(&stats.Warning)
	r.db.WithContext(ctx).Model(&monitor.AlertEvent{}).Where("status = 'firing' AND severity = 'info'").Count(&stats.Info)
	return stats, nil
}

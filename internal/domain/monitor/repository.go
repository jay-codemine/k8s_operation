package monitor

import "context"

// MonitorRepository 监控域仓储接口
type MonitorRepository interface {
	// 数据源
	DatasourceSave(ctx context.Context, ds *Datasource) error
	DatasourceFindByID(ctx context.Context, id int64) (*Datasource, error)
	DatasourceUpdate(ctx context.Context, ds *Datasource) error
	DatasourceDelete(ctx context.Context, id int64) error
	DatasourceQuery(ctx context.Context, req DatasourceListReq) (*DatasourceListResp, error)
	DatasourceFindDefault(ctx context.Context, types []string) (*Datasource, error)

	// 告警规则
	AlertRuleSave(ctx context.Context, rule *AlertRule) error
	AlertRuleFindByID(ctx context.Context, id int64) (*AlertRule, error)
	AlertRuleUpdate(ctx context.Context, rule *AlertRule) error
	AlertRuleDelete(ctx context.Context, id int64) error
	AlertRuleToggle(ctx context.Context, id int64, enabled bool) error
	AlertRuleQuery(ctx context.Context, req AlertRuleListReq) (*AlertRuleListResp, error)
	AlertRuleGroups(ctx context.Context) ([]string, error)

	// 告警事件
	AlertEventFindByID(ctx context.Context, id int64) (*AlertEvent, error)
	AlertEventQuery(ctx context.Context, req AlertEventListReq) (*AlertEventListResp, error)
	AlertEventAck(ctx context.Context, id int64, userID int64) error
	AlertEventResolve(ctx context.Context, id int64) error
	AlertEventStats(ctx context.Context) (*AlertStats, error)
}

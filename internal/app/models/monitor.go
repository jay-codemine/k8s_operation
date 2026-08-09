package models

import dm "k8soperation/internal/domain/monitor"

// Monitor domain type aliases — 保持向后兼容，实际定义在 internal/domain/monitor/
// 新代码应直接导入 domain/monitor 包

type MonitorDatasource = dm.Datasource
type MonitorAlertRule = dm.AlertRule
type MonitorAlertEvent = dm.AlertEvent
type MonitorNotifyChannel = dm.NotifyChannel
type MonitorSilenceRule = dm.SilenceRule
type MonitorInhibitRule = dm.InhibitRule
type MonitorAggregateRule = dm.AggregateRule
type MonitorNotifyTemplate = dm.NotifyTemplate
type MonitorNotifyRoutePolicy = dm.NotifyRoutePolicy

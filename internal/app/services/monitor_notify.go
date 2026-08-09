package services

import dm "k8soperation/internal/domain/monitor"

// ===== 类型和函数别名（领域实现位于 domain/monitor）=====

// AlertNotification 告警通知数据
type AlertNotification = dm.AlertNotification

// SendNotification 根据渠道类型发送通知
var SendNotification = dm.SendNotification

// NotifyChannelListReq 通知渠道列表请求
type NotifyChannelListReq = dm.NotifyChannelListReq


// RoutePolicyListReq 路由策略列表请求
type RoutePolicyListReq = dm.RoutePolicyListReq


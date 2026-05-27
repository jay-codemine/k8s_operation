// src/api/monitoring.js
import http from './http'

// 监控模块统一使用静默模式，由页面自行展示连接状态，避免 Network Error toast 堆叠
const silentConfig = { _silent: true }

// ==================== 监控指标查询 ====================
// 所有指标查询函数支持可选的 extra 参数对象（如 { datasource_id: 3 }），
// 用于把前端选择的数据源 ID 透传到后端，让 service 层按指定 ID 解析数据源

// 获取集群监控总览
export const getClusterOverview = (extra = {}) =>
  http.get('/api/v1/monitoring/overview', { params: extra, ...silentConfig })

// 获取节点指标列表
export const getNodeMetrics = (extra = {}) =>
  http.get('/api/v1/monitoring/nodes', { params: extra, ...silentConfig })

// 获取资源趋势数据
export const getResourceTrend = (resource, duration = '1h', extra = {}) =>
  http.get(`/api/v1/monitoring/trend/${resource}`, { params: { duration, ...extra }, ...silentConfig })

// 获取 Top N Pods
export const getTopPods = (metric = 'cpu', extra = {}) =>
  http.get('/api/v1/monitoring/top-pods', { params: { metric, ...extra }, ...silentConfig })

// Prometheus 健康检查
export const checkHealth = (extra = {}) =>
  http.get('/api/v1/monitoring/health', { params: extra, ...silentConfig })

// 集群健康评分
export const getHealthScore = (extra = {}) =>
  http.get('/api/v1/monitoring/score', { params: extra, ...silentConfig })

// 节点热力图（metric=cpu|memory, duration=1h|3h|6h）
export const getNodeHeatmap = (metric = 'cpu', duration = '1h', extra = {}) =>
  http.get('/api/v1/monitoring/heatmap', { params: { metric, duration, ...extra }, ...silentConfig })

// Pod 状态分布
export const getPodStatusDistribution = (extra = {}) =>
  http.get('/api/v1/monitoring/pod-status', { params: extra, ...silentConfig })

// 异常 Pod（重启超过阈值）
export const getAbnormalPods = (extra = {}) =>
  http.get('/api/v1/monitoring/abnormal-pods', { params: extra, ...silentConfig })

// Namespace 维度聚合指标
export const getNamespaceMetrics = (extra = {}) =>
  http.get('/api/v1/monitoring/namespaces', { params: extra, ...silentConfig })

// 单节点详情聚合（当前指标 + 5 维趋势 + Top Pod + 元信息）
export const getNodeDetail = (instance, duration = '1h', extra = {}) =>
  http.get('/api/v1/monitoring/node-detail', { params: { instance, duration, ...extra }, ...silentConfig })

// ==================== 数据源管理 ====================

export const listDatasources = (params = {}) =>
  http.get('/api/v1/monitoring/datasource', { params })

export const getDatasource = (id) =>
  http.get(`/api/v1/monitoring/datasource/${id}`)

export const createDatasource = (data) =>
  http.post('/api/v1/monitoring/datasource', data)

export const updateDatasource = (id, data) =>
  http.put(`/api/v1/monitoring/datasource/${id}`, data)

export const deleteDatasource = (id) =>
  http.delete(`/api/v1/monitoring/datasource/${id}`)

export const testDatasourceConnection = (data) =>
  http.post('/api/v1/monitoring/datasource/test', data)

export const testDatasourceById = (id) =>
  http.post(`/api/v1/monitoring/datasource/${id}/test`)

// ==================== 告警规则管理 ====================

export const listAlertRules = (params = {}) =>
  http.get('/api/v1/monitoring/alert-rule', { params })

export const getAlertRuleGroups = () =>
  http.get('/api/v1/monitoring/alert-rule/groups')

export const getAlertRule = (id) =>
  http.get(`/api/v1/monitoring/alert-rule/${id}`)

export const createAlertRule = (data) =>
  http.post('/api/v1/monitoring/alert-rule', data)

export const updateAlertRule = (id, data) =>
  http.put(`/api/v1/monitoring/alert-rule/${id}`, data)

export const deleteAlertRule = (id) =>
  http.delete(`/api/v1/monitoring/alert-rule/${id}`)

export const toggleAlertRule = (id, enabled) =>
  http.put(`/api/v1/monitoring/alert-rule/${id}/toggle`, { enabled })

// YAML 批量导入/导出
export const importAlertRulesYAML = (data) =>
  http.post('/api/v1/monitoring/alert-rule/import-yaml', data)

export const exportAlertRulesYAML = (params = {}) =>
  http.get('/api/v1/monitoring/alert-rule/export-yaml', { params })

// ==================== 告警事件 ====================

export const listAlertEvents = (params = {}) =>
  http.get('/api/v1/monitoring/alert-event', { params })

export const getAlertStats = () =>
  http.get('/api/v1/monitoring/alert-event/stats')

export const getAlertEvent = (id) =>
  http.get(`/api/v1/monitoring/alert-event/${id}`)

export const ackAlertEvent = (id) =>
  http.put(`/api/v1/monitoring/alert-event/${id}/ack`)

export const resolveAlertEvent = (id) =>
  http.put(`/api/v1/monitoring/alert-event/${id}/resolve`)

// ==================== 通知渠道管理 ====================

export const listNotifyChannels = (params = {}) =>
  http.get('/api/v1/monitoring/notify-channel', { params })

export const getNotifyChannel = (id) =>
  http.get(`/api/v1/monitoring/notify-channel/${id}`)

export const createNotifyChannel = (data) =>
  http.post('/api/v1/monitoring/notify-channel', data)

export const updateNotifyChannel = (id, data) =>
  http.put(`/api/v1/monitoring/notify-channel/${id}`, data)

export const deleteNotifyChannel = (id) =>
  http.delete(`/api/v1/monitoring/notify-channel/${id}`)

export const testNotifyChannel = (id) =>
  http.post(`/api/v1/monitoring/notify-channel/${id}/test`)

// ==================== 静默规则管理 ====================

export const listSilenceRules = (params = {}) =>
  http.get('/api/v1/monitoring/silence-rule', { params })

export const getSilenceRule = (id) =>
  http.get(`/api/v1/monitoring/silence-rule/${id}`)

export const createSilenceRule = (data) =>
  http.post('/api/v1/monitoring/silence-rule', data)

export const updateSilenceRule = (id, data) =>
  http.put(`/api/v1/monitoring/silence-rule/${id}`, data)

export const deleteSilenceRule = (id) =>
  http.delete(`/api/v1/monitoring/silence-rule/${id}`)

// ==================== 抑制规则管理 ====================

export const listInhibitRules = (params = {}) =>
  http.get('/api/v1/monitoring/inhibit-rule', { params })

export const createInhibitRule = (data) =>
  http.post('/api/v1/monitoring/inhibit-rule', data)

export const updateInhibitRule = (id, data) =>
  http.put(`/api/v1/monitoring/inhibit-rule/${id}`, data)

export const deleteInhibitRule = (id) =>
  http.delete(`/api/v1/monitoring/inhibit-rule/${id}`)

// ==================== 聚合规则管理 ====================

export const listAggregateRules = (params = {}) =>
  http.get('/api/v1/monitoring/aggregate-rule', { params })

export const createAggregateRule = (data) =>
  http.post('/api/v1/monitoring/aggregate-rule', data)

export const updateAggregateRule = (id, data) =>
  http.put(`/api/v1/monitoring/aggregate-rule/${id}`, data)

export const deleteAggregateRule = (id) =>
  http.delete(`/api/v1/monitoring/aggregate-rule/${id}`)

// ==================== 通知模板管理 ====================

export const listNotifyTemplates = (params = {}) =>
  http.get('/api/v1/monitoring/notify-template', { params })

export const getNotifyTemplate = (id) =>
  http.get(`/api/v1/monitoring/notify-template/${id}`)

export const createNotifyTemplate = (data) =>
  http.post('/api/v1/monitoring/notify-template', data)

export const updateNotifyTemplate = (id, data) =>
  http.put(`/api/v1/monitoring/notify-template/${id}`, data)

export const deleteNotifyTemplate = (id) =>
  http.delete(`/api/v1/monitoring/notify-template/${id}`)

export const previewNotifyTemplate = (data) =>
  http.post('/api/v1/monitoring/notify-template/preview', data)

export const setDefaultNotifyTemplate = (id) =>
  http.put(`/api/v1/monitoring/notify-template/${id}/default`)

// ==================== Loki 日志查询 ====================

// Loki 健康检查
export const checkLokiHealth = () => http.get('/api/v1/monitoring/loki/health', silentConfig)

// 查询日志
export const queryLokiLogs = (params = {}) =>
  http.get('/api/v1/monitoring/loki/query', { params, ...silentConfig })

// 获取标签列表
export const getLokiLabels = (params = {}) =>
  http.get('/api/v1/monitoring/loki/labels', { params, ...silentConfig })

// 获取标签值
export const getLokiLabelValues = (name, params = {}) =>
  http.get(`/api/v1/monitoring/loki/label/${name}/values`, { params, ...silentConfig })

// 获取日志流列表
export const getLokiStreams = (params = {}) =>
  http.get('/api/v1/monitoring/loki/streams', { params, ...silentConfig })

// 获取日志量趋势
export const getLokiVolume = (params = {}) =>
  http.get('/api/v1/monitoring/loki/volume', { params, ...silentConfig })

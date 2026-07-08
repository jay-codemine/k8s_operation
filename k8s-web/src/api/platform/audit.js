import http from '@/api/http'

// ========== 审计日志 API ==========

/**
 * 获取审计日志列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页条数
 * @param {string} params.username - 用户名筛选
 * @param {string} params.action - 操作类型筛选
 * @param {string} params.module - 模块筛选
 * @param {string} params.status - 状态筛选
 * @param {number} params.start_time - 开始时间(unix)
 * @param {number} params.end_time - 结束时间(unix)
 * @param {string} params.keyword - 关键词搜索
 */
export function getAuditLogs(params) {
  return http.get('/api/v1/platform/audit/logs', { params })
}

/**
 * 获取审计日志详情
 * @param {number} id - 日志ID
 */
export function getAuditLogDetail(id) {
  return http.get(`/api/v1/platform/audit/logs/${id}`)
}

/**
 * 获取审计统计数据
 */
export function getAuditStatistics() {
  return http.get('/api/v1/platform/audit/statistics')
}

/**
 * 获取审计日志保留策略
 */
export function getAuditRetention() {
  return http.get('/api/v1/platform/audit/retention')
}

/**
 * 更新审计日志保留策略
 * @param {Object} data - 保留策略
 * @param {number} data.retention_days - 保留天数(0=永久)
 * @param {boolean} data.is_permanent - 是否永久保留
 */
export function updateAuditRetention(data) {
  return http.put('/api/v1/platform/audit/retention', data)
}

/**
 * 手动清理过期审计日志
 */
export function cleanupAuditLogs() {
  return http.post('/api/v1/platform/audit/cleanup')
}

/**
 * 导出审计日志
 * @param {Object} params - 筛选参数
 */
export function exportAuditLogs(params) {
  return http.get('/api/v1/platform/audit/export', { params })
}

import http from '@/api/http'

const AIOPS_BASE = '/api/v1/ai/ops'

export default {
  // 仪表盘统计
  getDashboard() {
    return http.get(`${AIOPS_BASE}/dashboard`)
  },

  // AI 告警分析
  analyzeAlert(data) {
    return http.post(`${AIOPS_BASE}/alert/analyze`, data)
  },

  // AI 日志诊断
  diagnoseLogs(data) {
    return http.post(`${AIOPS_BASE}/log/diagnose`, data)
  },

  // 手动触发巡检
  runInspection() {
    return http.post(`${AIOPS_BASE}/inspection/run`)
  },

  // 巡检报告列表
  getInspectionList(params) {
    return http.get(`${AIOPS_BASE}/inspection/list`, { params })
  },

  // 巡检报告详情
  getInspectionDetail(id) {
    return http.get(`${AIOPS_BASE}/inspection/${id}`)
  },

  // 导出巡检报告 (Markdown)
  exportReport(id) {
    return http.get(`${AIOPS_BASE}/inspection/${id}/export`)
  },

  // 下载巡检报告文件
  async downloadReport(id) {
    const res = await http.get(`${AIOPS_BASE}/inspection/${id}/export`)
    if (res.code === 0 && res.data?.content) {
      const blob = new Blob([res.data.content], { type: 'text/markdown;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `inspection_report_${id}.md`
      a.click()
      URL.revokeObjectURL(url)
    }
  },

  // 发送巡检报告到通知渠道
  notifyReport(id, channelIds) {
    return http.post(`${AIOPS_BASE}/inspection/${id}/notify`, { channel_ids: channelIds })
  },

  // 获取可用通知渠道
  getNotifyChannels() {
    return http.get(`${AIOPS_BASE}/channels`)
  },

  // 分析记录列表
  getRecords(params) {
    return http.get(`${AIOPS_BASE}/records`, { params })
  }
}

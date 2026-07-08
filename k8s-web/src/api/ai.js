// src/api/ai.js - AI 助手 API
import http from './http'

const BASE = '/api/v1/ai'

// AI 请求超时时间（毫秒）—— AI 调用通常较慢（Function Calling 可能需要多轮），单独设置更长超时
const AI_TIMEOUT = 200000

// AI 状态检查
export const getAIStatus = () => http.get(`${BASE}/status`)

// 获取可用 AI 提供商和模型列表
export const getAIModels = () => http.get(`${BASE}/models`)

// 普通对话
export const aiChat = (data) => http.post(`${BASE}/chat`, data, { timeout: AI_TIMEOUT })

// 快捷问答
export const aiQuickAsk = (data) => http.post(`${BASE}/quick-ask`, data, { timeout: AI_TIMEOUT })

// 会话列表
export const getConversations = () => http.get(`${BASE}/conversations`)

// 会话消息历史
export const getConversationMessages = (id) => http.get(`${BASE}/conversations/${id}/messages`)

// 删除会话
export const deleteConversation = (id) => http.delete(`${BASE}/conversations/${id}`)

// 审批列表
export const getApprovals = (params) => http.get(`${BASE}/approvals`, { params })

// 我的审批
export const getMyApprovals = (params) => http.get(`${BASE}/approvals/mine`, { params })

// 待审批数量（静默模式，不弹错误提示）
export const getPendingApprovalCount = () =>
  http.get(`${BASE}/approvals/pending-count`, { _silent: true })

// 审批详情
export const getApprovalDetail = (id) => http.get(`${BASE}/approvals/${id}`)

// 通过审批
export const approveApproval = (id, data) => http.post(`${BASE}/approvals/${id}/approve`, data)

// 拒绝审批
export const rejectApproval = (id, data) => http.post(`${BASE}/approvals/${id}/reject`, data)

// 取消审批
export const cancelApproval = (id) => http.post(`${BASE}/approvals/${id}/cancel`)

// 删除审批记录
export const deleteApproval = (id) => http.delete(`${BASE}/approvals/${id}`)

// 更新审批备注
export const updateApproval = (id, data) => http.put(`${BASE}/approvals/${id}`, data)

// 审批统计数据
export const getApprovalStats = () => http.get(`${BASE}/approvals/stats`)

// AI 日志查询（排查问题）
export const getAILogs = (params) => http.get(`${BASE}/logs`, { params })

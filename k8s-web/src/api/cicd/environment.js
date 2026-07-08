// src/api/cicd/environment.js
import http from '../http'
import { API_BASE } from '../paths'

// ==================== 环境管理 ====================

// 获取环境列表
export function getEnvironmentList(params) {
  return http.get(`${API_BASE}/k8s/cicd/environment/list`, { params })
}

// 获取环境详情
export function getEnvironmentDetail(id) {
  return http.get(`${API_BASE}/k8s/cicd/environment/detail`, { params: { id } })
}

// 创建环境
export function createEnvironment(data) {
  return http.post(`${API_BASE}/k8s/cicd/environment/create`, data)
}

// 更新环境
export function updateEnvironment(data) {
  return http.post(`${API_BASE}/k8s/cicd/environment/update`, data)
}

// 删除环境
export function deleteEnvironment(id) {
  return http.post(`${API_BASE}/k8s/cicd/environment/delete`, { id })
}

// ==================== 审批策略管理 ====================

// 获取环境审批策略（通过环境详情接口获取 approval_levels）
export function getApprovalPolicy(envId) {
  return http.get(`${API_BASE}/k8s/cicd/environment/detail`, { params: { id: envId } })
}

// 更新环境审批策略（更新环境的 require_approval + approval_levels）
export function updateApprovalPolicy(envId, requireApproval, approvalLevels) {
  return http.post(`${API_BASE}/k8s/cicd/environment/update`, {
    id: envId,
    require_approval: requireApproval,
    approval_levels: approvalLevels
  })
}

// ==================== 审批流程 ====================

// 获取审批列表
export function getApprovalList(params) {
  return http.get(`${API_BASE}/k8s/cicd/approval/list`, { params })
}

// 获取审批统计
export function getApprovalStats() {
  return http.get(`${API_BASE}/k8s/cicd/approval/stats`)
}

// 获取审批详情
export function getApprovalDetail(id) {
  return http.get(`${API_BASE}/k8s/cicd/approval/detail`, { params: { id } })
}

// 获取待审批列表
export function getPendingApprovals() {
  return http.get(`${API_BASE}/k8s/cicd/approval/pending`)
}

// 创建审批申请
export function createApproval(data) {
  return http.post(`${API_BASE}/k8s/cicd/approval/create`, data)
}

// 审批操作（通过/拒绝）
export function approvalAction(data) {
  return http.post(`${API_BASE}/k8s/cicd/approval/action`, data)
}

// 批量审批操作（通过/拒绝）
export function approvalBatchAction(data) {
  return http.post(`${API_BASE}/k8s/cicd/approval/batch-action`, data)
}

// 更新审批记录
export function updateApproval(data) {
  return http.post(`${API_BASE}/k8s/cicd/approval/update`, data)
}

// 删除审批记录
export function deleteApproval(id) {
  return http.post(`${API_BASE}/k8s/cicd/approval/delete`, { id })
}

export default {
  getEnvironmentList,
  getEnvironmentDetail,
  createEnvironment,
  updateEnvironment,
  deleteEnvironment,
  getApprovalPolicy,
  updateApprovalPolicy,
  getApprovalList,
  getApprovalStats,
  getApprovalDetail,
  getPendingApprovals,
  createApproval,
  approvalAction,
  approvalBatchAction,
  updateApproval,
  deleteApproval
}

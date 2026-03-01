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

// ==================== 审批流程 ====================

// 获取审批列表
export function getApprovalList(params) {
  return http.get(`${API_BASE}/k8s/cicd/approval/list`, { params })
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

export default {
  getEnvironmentList,
  getEnvironmentDetail,
  createEnvironment,
  updateEnvironment,
  deleteEnvironment,
  getApprovalList,
  getApprovalDetail,
  getPendingApprovals,
  createApproval,
  approvalAction
}

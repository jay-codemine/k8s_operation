// src/api/rbac.js
import http from './http'
import { API_BASE } from './paths'

// ==================== 角色管理 ====================

// 获取角色列表（分页）
export function getRoleList(params) {
  return http.get(`${API_BASE}/rbac/role/list`, { params })
}

// 获取所有角色（下拉选择用）
export function getAllRoles() {
  return http.get(`${API_BASE}/rbac/role/all`)
}

// 获取角色详情
export function getRoleDetail(id) {
  return http.get(`${API_BASE}/rbac/role/detail`, { params: { id } })
}

// 创建角色
export function createRole(data) {
  return http.post(`${API_BASE}/rbac/role/create`, data)
}

// 更新角色
export function updateRole(data) {
  return http.post(`${API_BASE}/rbac/role/update`, data)
}

// 删除角色
export function deleteRole(id) {
  return http.post(`${API_BASE}/rbac/role/delete?id=${id}`)
}

// ==================== 权限管理 ====================

// 获取权限列表
export function getPermissionList() {
  return http.get(`${API_BASE}/rbac/permission/list`)
}

// 获取角色权限列表
export function getRolePermissions(roleId) {
  return http.get(`${API_BASE}/rbac/role/permissions`, { params: { role_id: roleId } })
}

// 更新角色权限
export function updateRolePermissions(data) {
  return http.post(`${API_BASE}/rbac/role/permissions/update`, data)
}

// 获取角色绑定的用户列表
export function getRoleUsers(roleId) {
  return http.get(`${API_BASE}/rbac/role/users`, { params: { role_id: roleId } })
}

// ==================== 用户角色管理 ====================

// 分配用户角色
export function assignUserRole(data) {
  return http.post(`${API_BASE}/rbac/user-role/assign`, data)
}

// 获取用户角色
export function getUserRoles(userId) {
  return http.get(`${API_BASE}/rbac/user-role/list`, { params: { user_id: userId } })
}

// ==================== 集群权限管理 ====================

// 创建集群权限
export function createClusterPermission(data) {
  return http.post(`${API_BASE}/rbac/cluster-permission/create`, data)
}

// 更新集群权限
export function updateClusterPermission(data) {
  return http.post(`${API_BASE}/rbac/cluster-permission/update`, data)
}

// 删除集群权限
export function deleteClusterPermission(id) {
  return http.post(`${API_BASE}/rbac/cluster-permission/delete?id=${id}`)
}

// 获取集群权限列表
export function getClusterPermissionList(params) {
  return http.get(`${API_BASE}/rbac/cluster-permission/list`, { params })
}

// 批量分配集群权限
export function batchAssignClusterPermission(data) {
  return http.post(`${API_BASE}/rbac/cluster-permission/batch`, data)
}

// ==================== 用户RBAC信息 ====================

// 获取用户RBAC信息（角色+集群权限）
export function getUserRBACInfo(userId) {
  return http.get(`${API_BASE}/rbac/user/info`, { params: { user_id: userId } })
}

// 获取当前登录用户的完整权限信息（用于权限隔离）
export function getUserPermissions() {
  return http.get(`${API_BASE}/rbac/user/permissions`)
}

// 获取当前用户可访问的集群
export function getAccessibleClusters() {
  return http.get(`${API_BASE}/rbac/user/clusters`)
}

// 获取当前用户在指定集群可访问的命名空间
export function getAccessibleNamespaces(clusterId) {
  return http.get(`${API_BASE}/rbac/user/namespaces`, { params: { cluster_id: clusterId } })
}

// 检查权限
export function checkPermission(clusterId, action) {
  return http.get(`${API_BASE}/rbac/check`, { params: { cluster_id: clusterId, action } })
}

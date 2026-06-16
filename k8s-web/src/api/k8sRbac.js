import http from './http'

/**
 * K8s RBAC API - ServiceAccount / Role / RoleBinding 管理
 */

// ==================== ServiceAccount ====================

// 获取 ServiceAccount 列表
export function listServiceAccounts(clusterId, namespace = '') {
  return http({
    url: '/api/v1/k8s/rbac/serviceaccounts',
    method: 'get',
    headers: { 'X-Cluster-ID': clusterId },
    params: { namespace }
  })
}

// 获取 ServiceAccount 详情
export function getServiceAccount(clusterId, namespace, name) {
  return http({
    url: '/api/v1/k8s/rbac/serviceaccount',
    method: 'get',
    headers: { 'X-Cluster-ID': clusterId },
    params: { namespace, name }
  })
}

// 创建 ServiceAccount
export function createServiceAccount(clusterId, data) {
  return http({
    url: '/api/v1/k8s/rbac/serviceaccount',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data
  })
}

// 删除 ServiceAccount
export function deleteServiceAccount(clusterId, namespace, name) {
  return http({
    url: '/api/v1/k8s/rbac/serviceaccount',
    method: 'delete',
    headers: { 'X-Cluster-ID': clusterId },
    params: { namespace, name }
  })
}

// ==================== Role / ClusterRole ====================

// 获取 Role 列表（包含 ClusterRole）
export function listRoles(clusterId, namespace = '', type = '') {
  return http({
    url: '/api/v1/k8s/rbac/roles',
    method: 'get',
    headers: { 'X-Cluster-ID': clusterId },
    params: { namespace, type }
  })
}

// 创建 Role 或 ClusterRole
export function createRole(clusterId, data) {
  return http({
    url: '/api/v1/k8s/rbac/role',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data
  })
}

// 删除 Role 或 ClusterRole
export function deleteRole(clusterId, type, namespace, name) {
  return http({
    url: '/api/v1/k8s/rbac/role',
    method: 'delete',
    headers: { 'X-Cluster-ID': clusterId },
    params: { type, namespace, name }
  })
}

// ==================== RoleBinding / ClusterRoleBinding ====================

// 获取 RoleBinding 列表（包含 ClusterRoleBinding）
export function listRoleBindings(clusterId, namespace = '') {
  return http({
    url: '/api/v1/k8s/rbac/rolebindings',
    method: 'get',
    headers: { 'X-Cluster-ID': clusterId },
    params: { namespace }
  })
}

// 创建 RoleBinding 或 ClusterRoleBinding
export function createRoleBinding(clusterId, data) {
  return http({
    url: '/api/v1/k8s/rbac/rolebinding',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data
  })
}

// 删除 RoleBinding 或 ClusterRoleBinding
export function deleteRoleBinding(clusterId, type, namespace, name) {
  return http({
    url: '/api/v1/k8s/rbac/rolebinding',
    method: 'delete',
    headers: { 'X-Cluster-ID': clusterId },
    params: { type, namespace, name }
  })
}

// ==================== SubjectAccessReview ====================

// 检查主体权限
export function checkSubjectAccess(clusterId, data) {
  return http({
    url: '/api/v1/k8s/rbac/check',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data
  })
}

// 批量检查主体权限
export function batchCheckSubjectAccess(clusterId, checks) {
  return http({
    url: '/api/v1/k8s/rbac/check/batch',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data: { checks }
  })
}

// ==================== K8s Events (Audit) ====================

// 获取 K8s 事件列表（用于审计日志）
export function getK8sEvents(clusterId, params = {}) {
  return http({
    url: '/api/v1/k8s/deployment/events',
    method: 'post',
    headers: { 'X-Cluster-ID': clusterId },
    data: {
      namespace: params.namespace || '',
      kind: params.kind || '',
      name: params.name || '',
      type: params.type || '',
      limit: params.limit || 100,
      since_seconds: params.since_seconds || 86400 // 默认最近24小时
    }
  })
}

// src/api/namespace.js
// 命名空间 API 包装（兼容旧导入方式）
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'
import namespacesApi from '@/api/cluster/namespaces'

/**
 * 获取命名空间列表（支持显式传递集群ID）
 * @param {number|string} clusterId - 集群ID
 * @param {Object} params - 查询参数
 */
export function getNamespaces(clusterId, params = {}) {
  // 显式传递集群ID到请求头，不依赖全局 store
  if (clusterId) {
    return http.get(`${K8S_BASE}/namespace/list`, {
      params,
      headers: { 'X-Cluster-ID': String(clusterId) }
    })
  }
  // 兜底：使用原始 API（依赖全局 store）
  return namespacesApi.list(params)
}

/**
 * 创建命名空间
 * @param {Object} data
 */
export function createNamespace(data) {
  return namespacesApi.create(data)
}

/**
 * 删除命名空间
 * @param {Object} params
 */
export function deleteNamespace(params) {
  return namespacesApi.delete(params)
}

/**
 * 获取命名空间详情
 * @param {Object} params
 */
export function getNamespaceDetail(params) {
  return namespacesApi.detail(params)
}

export default namespacesApi

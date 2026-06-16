/**
 * 镜像管理 API
 * 参考: Rancher, KubeSphere, Kuboard
 */
import request from './http'

// ========== 镜像仓库管理 ==========

/**
 * 获取镜像仓库列表（分页）
 */
export function getRegistryList(params) {
  return request({
    url: '/api/v1/image/registry/list',
    method: 'get',
    params
  })
}

/**
 * 获取所有镜像仓库（不分页，用于下拉选择）
 */
export function getAllRegistries() {
  return request({
    url: '/api/v1/image/registry/all',
    method: 'get'
  })
}

/**
 * 获取镜像仓库详情
 */
export function getRegistryDetail(id) {
  return request({
    url: '/api/v1/image/registry/detail',
    method: 'get',
    params: { id }
  })
}

/**
 * 获取仓库统计信息
 */
export function getRegistryStats() {
  return request({
    url: '/api/v1/image/registry/stats',
    method: 'get'
  })
}

/**
 * 创建镜像仓库
 */
export function createRegistry(data) {
  return request({
    url: '/api/v1/image/registry/create',
    method: 'post',
    data
  })
}

/**
 * 更新镜像仓库
 */
export function updateRegistry(data) {
  return request({
    url: '/api/v1/image/registry/update',
    method: 'post',
    data
  })
}

/**
 * 删除镜像仓库
 */
export function deleteRegistry(id) {
  return request({
    url: '/api/v1/image/registry/delete',
    method: 'post',
    params: { id }
  })
}

/**
 * 检测仓库连接
 */
export function checkRegistryConnection(id) {
  return request({
    url: '/api/v1/image/registry/check',
    method: 'post',
    params: { id }
  })
}

/**
 * 设置默认仓库
 */
export function setDefaultRegistry(id) {
  return request({
    url: '/api/v1/image/registry/default',
    method: 'post',
    params: { id }
  })
}

// ========== 镜像浏览 ==========

/**
 * 获取镜像项目列表
 */
export function listRepositories(registryId) {
  return request({
    url: '/api/v1/image/browse/repositories',
    method: 'get',
    params: { registry_id: registryId }
  })
}

/**
 * 获取镜像标签列表
 */
export function listTags(registryId, repository) {
  return request({
    url: '/api/v1/image/browse/tags',
    method: 'get',
    params: { registry_id: registryId, repository }
  })
}

/**
 * 获取镜像详情
 */
export function getImageDetail(registryId, repository, tag) {
  return request({
    url: '/api/v1/image/browse/detail',
    method: 'get',
    params: { registry_id: registryId, repository, tag }
  })
}

/**
 * 删除镜像标签
 */
export function deleteImageTag(registryId, repository, tag) {
  return request({
    url: '/api/v1/image/browse/delete',
    method: 'post',
    params: { registry_id: registryId, repository, tag }
  })
}

// ========== 清理策略 ==========

/**
 * 获取清理策略列表
 */
export function getCleanupPolicies(params) {
  return request({
    url: '/api/v1/image/cleanup/list',
    method: 'get',
    params
  })
}

/**
 * 创建清理策略
 */
export function createCleanupPolicy(data) {
  return request({
    url: '/api/v1/image/cleanup/create',
    method: 'post',
    data
  })
}

/**
 * 更新清理策略
 */
export function updateCleanupPolicy(data) {
  return request({
    url: '/api/v1/image/cleanup/update',
    method: 'post',
    data
  })
}

/**
 * 删除清理策略
 */
export function deleteCleanupPolicy(id) {
  return request({
    url: '/api/v1/image/cleanup/delete',
    method: 'post',
    params: { id }
  })
}

/**
 * 启用/禁用清理策略
 */
export function toggleCleanupPolicy(id, enabled) {
  return request({
    url: '/api/v1/image/cleanup/toggle',
    method: 'post',
    params: { id, enabled }
  })
}

/**
 * 手动执行清理策略
 */
export function runCleanupPolicy(id) {
  return request({
    url: '/api/v1/image/cleanup/run',
    method: 'post',
    params: { id }
  })
}

/**
 * 获取清理日志
 */
export function getCleanupLogs(params) {
  return request({
    url: '/api/v1/image/cleanup/logs',
    method: 'get',
    params
  })
}

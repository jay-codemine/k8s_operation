// src/api/cluster/workloads/deployments.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// Deployment 模块 API（全量）
// 对应后端路由: /api/v1/k8s/deployment/*
// =========================
const deploymentsApi = {
  // =========================
  // Deployment 基础 CRUD
  // =========================

  /**
   * 创建 Deployment（可选创建 Service）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {string} data.container_image - 容器镜像
   * @param {number} [data.replicas=1] - 副本数
   * @param {Object} [data.labels] - 标签
   * @param {boolean} [data.is_create_service] - 是否同时创建 Service
   */
  create(data) {
    return http.post(`${K8S_BASE}/deployment/create`, data)
  },

  /**
   * 从 YAML 创建 Deployment
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/deployment/create-from-yaml`, data)
  },

  /**
   * Deployment 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - Deployment 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/deployment/list`, {params})
  },

  /**
   * Deployment 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/deployment/detail`, {params})
  },

  /**
   * 删除 Deployment
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/deployment/delete`, {params})
  },

  /**
   * 删除 Deployment 对应的 Service
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称（用于查找关联的 Service）
   */
  deleteService(params) {
    return http.delete(`${K8S_BASE}/deployment/delete_service`, {params})
  },

  // =========================
  // 扩缩容（后端使用 POST）
  // =========================

  /**
   * 扩缩容（修改副本数）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {number} data.scale_num - 目标副本数（注意：后端字段名是 scale_num）
   */
  scale(data) {
    return http.post(`${K8S_BASE}/deployment/scale`, data)
  },

  // =========================
  // 镜像更新（后端使用 POST）
  // =========================

  /**
   * 更新镜像（触发滚动升级）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {string} data.container - 容器名称
   * @param {string} data.image - 新镜像地址
   */
  updateImage(data) {
    return http.post(`${K8S_BASE}/deployment/update-image`, data)
  },

  // =========================
  // Patch 模板（后端路由为 /patch_template）
  // =========================

  /**
   * Patch Deployment 模板
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {Object|Array} data.patch - patch 内容
   */
  patch(data) {
    return http.post(`${K8S_BASE}/deployment/patch_template`, data)
  },

  // =========================
  // 重启与回滚
  // =========================

  /**
   * 滚动重启 Deployment
   * 通过更新 annotation 触发滚动更新
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   */
  restart(data) {
    return http.post(`${K8S_BASE}/deployment/restart`, data)
  },

  /**
   * 回滚到指定 ReplicaSet
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {number} [data.revision] - 目标版本号（不传则回滚到上一个版本）
   */
  rollback(data) {
    return http.post(`${K8S_BASE}/deployment/rollback`, data)
  },

  // =========================
  // 关联资源
  // =========================

  /**
   * 获取 Deployment 对应的 Pod 列表
   * 后端路由为 /deploy_pods
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   */
  pods(params) {
    return http.get(`${K8S_BASE}/deployment/deploy_pods`, {params})
  },

  /**
   * 获取 Deployment 相关事件
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   * @param {string} [params.type] - 事件类型（Normal | Warning）
   * @param {number} [params.limit=50] - 返回条数限制
   * @param {number} [params.since_seconds=3600] - 最近N秒的事件
   */
  events(params) {
    return http.post(`${K8S_BASE}/deployment/events`, {
      namespace: params.namespace,
      kind: 'Deployment',
      name: params.name,
      type: params.type || '',
      limit: params.limit || 50,
      since_seconds: params.since_seconds || 3600,
    })
  },

  // =========================
  // 历史版本
  // =========================

  /**
   * 获取 Deployment 的历史版本（ReplicaSet 列表）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   */
  history(params) {
    return http.get(`${K8S_BASE}/deployment/history`, {params})
  },

  /**
   * 获取 Deployment YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Deployment 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/deployment/yaml`, {params})
  },

  /**
   * 应用 Deployment YAML
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Deployment 名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/deployment/apply_yaml`, data)
  },
}

export default deploymentsApi

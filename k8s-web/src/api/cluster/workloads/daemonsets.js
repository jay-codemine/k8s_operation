// src/api/cluster/workloads/daemonsets.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// DaemonSet 模块 API（全量）
// 对应后端路由: /api/v1/k8s/daemonset/*
// =========================
const daemonsetsApi = {
  // =========================
  // DaemonSet 基础 CRUD
  // =========================

  /**
   * 创建 DaemonSet（可选创建 Service）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - DaemonSet 名称
   * @param {string} data.container_image - 容器镜像
   * @param {Object} [data.labels] - 标签
   * @param {boolean} [data.is_create_service] - 是否同时创建 Service
   */
  create(data) {
    return http.post(`${K8S_BASE}/daemonset/create`, data)
  },

  /**
   * DaemonSet 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - DaemonSet 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/daemonset/list`, {params})
  },

  /**
   * DaemonSet 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/daemonset/detail`, {params})
  },

  /**
   * 删除 DaemonSet
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/daemonset/delete`, {params})
  },

  /**
   * 删除 DaemonSet 对应的 Service
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称（用于查找关联的 Service）
   */
  deleteService(params) {
    return http.delete(`${K8S_BASE}/daemonset/delete_service`, {params})
  },

  // =========================
  // 镜像更新（后端使用 POST）
  // =========================

  /**
   * 更新镜像（触发滚动升级）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - DaemonSet 名称
   * @param {string} data.container - 容器名称
   * @param {string} data.image - 新镜像地址
   */
  updateImage(data) {
    return http.put(`${K8S_BASE}/daemonset/update_image`, data)
  },

  // =========================
  // 重启与回滚
  // =========================

  /**
   * 滚动重启 DaemonSet
   * 通过更新 annotation 触发滚动更新
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - DaemonSet 名称
   */
  restart(data) {
    return http.post(`${K8S_BASE}/daemonset/restart`, data)
  },

  /**
   * 回滚到指定 ControllerRevision
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - DaemonSet 名称
   * @param {string} data.revision_name - ControllerRevision 名称
   */
  rollback(data) {
    return http.post(`${K8S_BASE}/daemonset/rollback`, data)
  },

  // =========================
  // 关联资源
  // =========================

  /**
   * 获取 DaemonSet 对应的 Pod 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   */
  pods(params) {
    return http.get(`${K8S_BASE}/daemonset/ds_pods`, {params})
  },

  /**
   * 获取 DaemonSet 相关事件
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   * @param {string} [params.type] - 事件类型（Normal | Warning）
   * @param {number} [params.limit=50] - 返回条数限制
   * @param {number} [params.since_seconds=3600] - 最近N秒的事件
   */
  events(params) {
    return http.post(`${K8S_BASE}/daemonset/events`, {
      namespace: params.namespace,
      kind: 'DaemonSet',
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
   * 获取 DaemonSet 的历史版本（ControllerRevision 列表）
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   */
  history(params) {
    return http.get(`${K8S_BASE}/daemonset/history`, {params})
  },

  /**
   * 获取 DaemonSet YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - DaemonSet 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/daemonset/yaml`, {params})
  },

  /**
   * 应用 DaemonSet YAML
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - DaemonSet 名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/daemonset/apply_yaml`, data)
  },

  /**
   * 从 YAML 创建 DaemonSet
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/daemonset/create-from-yaml`, data)
  },
}

export default daemonsetsApi

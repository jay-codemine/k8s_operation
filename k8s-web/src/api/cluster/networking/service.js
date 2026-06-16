// src/api/cluster/networking/service.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

// =========================
// Service 模块 API
// 对应后端路由: /api/v1/k8s/service/*
// =========================
const serviceApi = {
  // =========================
  // Service 基础 CRUD
  // =========================

  /**
   * 创建 Service
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Service 名称
   * @param {string} data.type - Service 类型（ClusterIP/NodePort/LoadBalancer）
   * @param {Array} data.ports - 端口配置
   * @param {Object} data.selector - 选择器
   */
  create(data) {
    return http.post(`${K8S_BASE}/service/create`, data)
  },

  /**
   * 从 YAML 创建 Service
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/service/create-from-yaml`, data)
  },

  /**
   * Service 列表
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - Service 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/service/list`, { params })
  },

  /**
   * Service 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Service 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/service/detail`, { params })
  },

  /**
   * 删除 Service
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Service 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/service/delete`, { params })
  },

  // =========================
  // Patch 更新
  // =========================

  /**
   * Patch Service（Strategic Merge Patch）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Service 名称
   * @param {string} data.content - Patch 内容（JSON字符串）
   */
  patch(data) {
    return http.patch(`${K8S_BASE}/service/patch`, data, {
      params: { namespace: data.namespace, name: data.name }
    })
  },

  /**
   * Patch Service（JSON Merge Patch）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Service 名称
   * @param {string} data.content - Patch 内容（JSON字符串）
   */
  patchJson(data) {
    return http.post(`${K8S_BASE}/service/patch_json`, data)
  },

  // =========================
  // Endpoints
  // =========================

  /**
   * 获取 Service Endpoints
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Service 名称
   */
  endpoints(params) {
    return http.get(`${K8S_BASE}/service/endpoints`, { params })
  },

  // =========================
  // YAML 操作
  // =========================

  /**
   * 获取 Service YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Service 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/service/yaml`, { params })
  },

  /**
   * 应用 Service YAML（创建或更新）
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/service/apply-yaml`, data)
  },
}

export default serviceApi

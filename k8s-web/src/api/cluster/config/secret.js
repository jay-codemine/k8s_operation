// src/api/cluster/config/secret.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// Secret 模块 API
// 对应后端路由: /api/v1/k8s/secret/*
// =========================
const secretApi = {
  // =========================
  // Secret 基础 CRUD
  // =========================

  /**
   * 创建 Secret
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Secret 名称
   * @param {string} data.type - Secret 类型 (Opaque, kubernetes.io/tls, etc.)
   * @param {Object} [data.data] - Secret 数据（Base64 编码）
   * @param {Object} [data.string_data] - Secret 数据（明文，后端自动编码）
   */
  create(data) {
    return http.post(`${K8S_BASE}/secret/create`, data)
  },

  /**
   * 从 YAML 创建 Secret
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/secret/create-from-yaml`, data)
  },

  /**
   * Secret 列表
   * @param {Object} params
   * @param {string} [params.namespace] - 命名空间（不传则查全部）
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - Secret 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/secret/list`, {params})
  },

  /**
   * Secret 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Secret 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/secret/detail`, {params})
  },

  /**
   * 删除 Secret
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Secret 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/secret/delete`, {params})
  },

  // =========================
  // Secret 特殊操作
  // =========================

  /**
   * Base64 解码 Secret 数据
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Secret 名称
   */
  decode(data) {
    return http.post(`${K8S_BASE}/secret/decode`, data)
  },

  /**
   * Patch Secret（StrategicMergePatch）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Secret 名称
   * @param {Object} data.patch - patch 内容
   */
  patch(data) {
    return http.patch(`${K8S_BASE}/secret/patch`, data)
  },

  /**
   * Patch Secret（JSON Merge Patch - 覆盖式）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Secret 名称
   * @param {Object} data.patch - patch 内容
   */
  patchJson(data) {
    return http.post(`${K8S_BASE}/secret/patch-json`, data)
  },

  /**
   * 获取 Secret YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Secret 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/secret/yaml`, {params})
  },

  /**
   * 应用 Secret YAML（更新）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Secret 名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/secret/apply-yaml`, data)
  },
}

export default secretApi

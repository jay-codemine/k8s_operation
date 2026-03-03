// src/api/cluster/namespaces.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// Namespace 模块 API
// 对应后端路由: /api/v1/k8s/namespace/*
// =========================
const namespacesApi = {
  // =========================
  // Namespace 基础 CRUD
  // =========================

  /**
   * 创建 Namespace
   * @param {Object} data
   * @param {string} data.name - 命名空间名称
   */
  create(data) {
    return http.post(`${K8S_BASE}/namespace/create`, data)
  },

  /**
   * Namespace 列表
   * @param {Object} params
   * @param {number} [params.page] - 页码
   * @param {number} [params.limit] - 每页数量
   * @param {string} [params.name] - 命名空间名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/namespace/list`, {params})
  },

  /**
   * Namespace 详情
   * 后端参数名：name（不是 namespace）
   * @param {Object} params
   * @param {string} params.name - 命名空间名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/namespace/detail`, {
      params: { name: params.name || params.namespace }
    })
  },

  /**
   * 删除 Namespace
   * 后端参数名：name（不是 namespace）
   * @param {Object} params
   * @param {string} params.name - 命名空间名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/namespace/delete`, {
      params: { name: params.name || params.namespace }
    })
  },

  /**
   * Patch Namespace (修改 labels/annotations)
   * 后端参数名：name + content
   * @param {Object} data
   * @param {string} data.name - 命名空间名称
   * @param {Object} data.content - Patch 内容（JSON 格式）
   */
  patch(data) {
    return http.patch(`${K8S_BASE}/namespace/patch`, {
      name: data.name || data.namespace,
      content: data.content || data.patch,
    })
  },

  /**
   * 修改 Namespace 标签
   * @param {Object} data
   * @param {string} data.name - 命名空间名称
   * @param {Object} [data.add] - 要添加/更新的标签 { key: value }
   * @param {Array} [data.remove] - 要删除的标签 key 数组
   */
  patchLabels(data) {
    return http.patch(`${K8S_BASE}/namespace/labels`, data)
  },

  /**
   * 获取 Namespace YAML
   * @param {Object} params
   * @param {string} params.name - 命名空间名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/namespace/yaml`, { params })
  },

  /**
   * 应用 Namespace YAML
   * @param {Object} data
   * @param {string} data.name - 命名空间名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/namespace/apply_yaml`, data)
  },
}

export default namespacesApi

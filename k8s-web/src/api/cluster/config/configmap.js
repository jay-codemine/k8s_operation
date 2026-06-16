// src/api/cluster/config/configmap.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

/**
 * ConfigMap API
 * 对应后端路由: /api/v1/k8s/configmap/*
 */
const configmapApi = {
  /**
   * 获取 ConfigMap 列表
   * @param {Object} params
   * @param {string} [params.namespace] - 命名空间（不传则查全部）
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - ConfigMap 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/configmap/list`, { params })
  },

  /**
   * 获取 ConfigMap 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - ConfigMap 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/configmap/detail`, { params })
  },

  /**
   * 创建 ConfigMap
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - ConfigMap 名称
   * @param {Object} data.data - 数据 key-value 对象
   */
  create(data) {
    return http.post(`${K8S_BASE}/configmap/create`, data)
  },

  /**
   * 删除 ConfigMap
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - ConfigMap 名称
   */
  deleteConfigMap(params) {
    return http.delete(`${K8S_BASE}/configmap/delete`, { params })
  },

  /**
   * 更新 ConfigMap data 字段
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - ConfigMap 名称
   * @param {Object} data.data - 数据 key-value 对象
   */
  updateData(data) {
    return http.put(`${K8S_BASE}/configmap/update-data`, data)
  },

  /**
   * Strategic Merge Patch 更新 ConfigMap
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - ConfigMap 名称
   * @param {string} data.content - JSON 字符串
   */
  patch(data) {
    return http.patch(`${K8S_BASE}/configmap/patch`, data)
  },

  /**
   * 获取 YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - ConfigMap 名称
   */
  getYaml(params) {
    return http.get(`${K8S_BASE}/configmap/yaml`, { params })
  },

  /**
   * 应用 YAML 更新 ConfigMap
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    console.log('[ConfigMap API] applyYaml 调用，参数:', data)
    console.log('[ConfigMap API] 请求体内容:', JSON.stringify(data))
    // 确保数据通过请求体发送，而不是查询参数
    return http({
      method: 'POST',
      url: `${K8S_BASE}/configmap/apply-yaml`,
      data: data,  // 注意：使用 data 而不是 params
      headers: {
        'Content-Type': 'application/json'
      }
    })
  },

  /**
   * 解析多资源 YAML
   * @param {Object} data
   * @param {string} data.yaml - 多资源 YAML 内容
   */
  parseMultiYaml(data) {
    return http({
      method: 'POST',
      url: `${K8S_BASE}/multi-resource/parse-yaml`,
      data: data,
      headers: {
        'Content-Type': 'application/json'
      }
    })
  },

  /**
   * 应用多资源 YAML
   * @param {Object} data
   * @param {string} data.yaml - 多资源 YAML 内容
   */
  applyMultiYaml(data) {
    return http({
      method: 'POST',
      url: `${K8S_BASE}/multi-resource/apply-yaml`,
      data: data,
      headers: {
        'Content-Type': 'application/json'
      }
    })
  }
}

export default configmapApi

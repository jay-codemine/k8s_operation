// src/api/cluster/extensions/crd.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

// =========================
// CRD/CR 动态资源管理 API
// 对应后端路由: /api/v1/k8s/crd/* 和 /api/v1/k8s/cr/*
// =========================
const crdApi = {
  // =========================
  // CRD 管理
  // =========================

  /**
   * 列出所有 CRD
   * @param {Object} params
   * @param {string} [params.keyword] - 搜索关键词（名称/Kind 模糊匹配）
   * @param {string} [params.group] - 按 API Group 过滤
   */
  listCRDs(params) {
    return http.get(`${K8S_BASE}/crd/list`, { params })
  },

  /**
   * 获取 CRD 详情
   * @param {Object} params
   * @param {string} params.name - CRD 名称（如 prometheusrules.monitoring.coreos.com）
   */
  getCRD(params) {
    return http.get(`${K8S_BASE}/crd/detail`, { params })
  },

  /**
   * 删除 CRD
   * @param {Object} params
   * @param {string} params.name - CRD 名称
   */
  deleteCRD(params) {
    return http.delete(`${K8S_BASE}/crd/delete`, { params })
  },

  // =========================
  // CR 实例管理
  // =========================

  /**
   * 列出 CR 实例
   * @param {Object} params
   * @param {string} [params.group] - API Group
   * @param {string} params.version - API Version
   * @param {string} params.resource - 资源复数名（如 prometheusrules）
   * @param {string} [params.namespace] - 命名空间（为空则查所有）
   * @param {string} [params.label_selector] - 标签选择器
   */
  listCRs(params) {
    return http.get(`${K8S_BASE}/cr/list`, { params })
  },

  /**
   * 获取单个 CR 详情
   * @param {Object} params
   * @param {string} [params.group] - API Group
   * @param {string} params.version - API Version
   * @param {string} params.resource - 资源复数名
   * @param {string} [params.namespace] - 命名空间
   * @param {string} params.name - CR 名称
   */
  getCR(params) {
    return http.get(`${K8S_BASE}/cr/detail`, { params })
  },

  /**
   * 创建 CR 实例（支持 DryRun）
   * @param {Object} data
   * @param {string} [data.group] - API Group
   * @param {string} data.version - API Version
   * @param {string} data.resource - 资源复数名
   * @param {string} [data.namespace] - 命名空间
   * @param {string} data.yaml - YAML 内容
   * @param {boolean} [data.dry_run] - 是否仅做 DryRun 校验
   */
  createCR(data) {
    return http.post(`${K8S_BASE}/cr/create`, data)
  },

  /**
   * 更新 CR 实例（支持 DryRun）
   * @param {Object} data
   * @param {string} [data.group] - API Group
   * @param {string} data.version - API Version
   * @param {string} data.resource - 资源复数名
   * @param {string} [data.namespace] - 命名空间
   * @param {string} data.name - CR 名称
   * @param {string} data.yaml - YAML 内容
   * @param {boolean} [data.dry_run] - 是否仅做 DryRun 校验
   */
  updateCR(data) {
    return http.put(`${K8S_BASE}/cr/update`, data)
  },

  /**
   * 删除 CR 实例
   * @param {Object} params
   * @param {string} [params.group] - API Group
   * @param {string} params.version - API Version
   * @param {string} params.resource - 资源复数名
   * @param {string} [params.namespace] - 命名空间
   * @param {string} params.name - CR 名称
   */
  deleteCR(params) {
    return http.delete(`${K8S_BASE}/cr/delete`, { params })
  },

  /**
   * 获取 CR 的 YAML 表示
   * @param {Object} params
   * @param {string} [params.group] - API Group
   * @param {string} params.version - API Version
   * @param {string} params.resource - 资源复数名
   * @param {string} [params.namespace] - 命名空间
   * @param {string} params.name - CR 名称
   */
  getCRYaml(params) {
    return http.get(`${K8S_BASE}/cr/yaml`, { params })
  },

  /**
   * DryRun 校验 CR
   * @param {Object} data
   * @param {string} [data.group] - API Group
   * @param {string} data.version - API Version
   * @param {string} data.resource - 资源复数名
   * @param {string} [data.namespace] - 命名空间
   * @param {string} [data.name] - CR 名称（更新时必传）
   * @param {string} data.yaml - YAML 内容
   * @param {boolean} data.is_update - 是否为更新校验
   */
  dryRun(data) {
    return http.post(`${K8S_BASE}/cr/dry-run`, data)
  },
}

export default crdApi

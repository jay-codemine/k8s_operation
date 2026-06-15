// src/api/cluster/workloads/autoscaler.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

/**
 * HPA（HorizontalPodAutoscaler）API
 * 对应后端: /api/v1/k8s/hpa/*
 */
export const hpaApi = {
  /** 列表（支持 namespace/name 过滤、分页） */
  list(params) {
    return http.get(`${K8S_BASE}/hpa/list`, { params })
  },

  /** 详情 */
  detail(params) {
    return http.get(`${K8S_BASE}/hpa/detail`, { params })
  },

  /** 创建 HPA */
  create(data) {
    return http.post(`${K8S_BASE}/hpa/create`, data)
  },

  /** 通过 YAML 创建 HPA */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/hpa/create-from-yaml`, data)
  },

  /** 更新 HPA */
  update(data) {
    return http.post(`${K8S_BASE}/hpa/update`, data)
  },

  /** 删除 HPA */
  delete(params) {
    return http.delete(`${K8S_BASE}/hpa/delete`, { params })
  },

  /** 单独修改副本数（min/max） */
  scale(data) {
    return http.post(`${K8S_BASE}/hpa/scale`, data)
  },

  /**
   * 批量扩容/缩容副本数（618 促销场景）
   * data: { items: [{namespace, name, min_replicas, max_replicas}, ...] }
   */
  batchScale(data) {
    return http.post(`${K8S_BASE}/hpa/batch-scale`, data)
  },

  /**
   * 批量查询当前 HPA 状态（统一数据，验证扩缩容是否生效）
   * data: { items: [{namespace, name}, ...] }
   */
  batchStatus(data) {
    return http.post(`${K8S_BASE}/hpa/batch-status`, data)
  },
}

/**
 * VPA（VerticalPodAutoscaler）API
 * 对应后端: /api/v1/k8s/vpa/*
 */
export const vpaApi = {
  /** 检测 VPA Operator 是否安装 */
  available() {
    return http.get(`${K8S_BASE}/vpa/available`)
  },

  /** 列表 */
  list(params) {
    return http.get(`${K8S_BASE}/vpa/list`, { params })
  },

  /** 详情 */
  detail(params) {
    return http.get(`${K8S_BASE}/vpa/detail`, { params })
  },

  /** 创建 VPA */
  create(data) {
    return http.post(`${K8S_BASE}/vpa/create`, data)
  },

  /** 通过 YAML 创建 VPA */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/vpa/create-from-yaml`, data)
  },

  /** 更新 VPA */
  update(data) {
    return http.post(`${K8S_BASE}/vpa/update`, data)
  },

  /** 删除 VPA */
  delete(params) {
    return http.delete(`${K8S_BASE}/vpa/delete`, { params })
  },
}

export default { hpaApi, vpaApi }

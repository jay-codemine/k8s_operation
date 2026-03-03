// src/api/cluster/workloads/jobs.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// Job 模块 API
// 对应后端路由: /api/v1/k8s/job/*
// =========================
const jobsApi = {
  // =========================
  // Job 基础 CRUD
  // =========================

  /**
   * 创建 Job
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Job 名称
   * @param {string} data.container_image - 容器镜像
   * @param {number} [data.completions=1] - 完成数
   * @param {number} [data.parallelism=1] - 并行度
   * @param {number} [data.backoff_limit=6] - 退避限制
   * @param {number} [data.active_deadline_seconds] - 活动截止时间（秒）
   * @param {Object} [data.labels] - 标签
   */
  create(data) {
    return http.post(`${K8S_BASE}/job/create`, data)
  },

  /**
   * 从 YAML 创建 Job
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/job/create-from-yaml`, data)
  },

  /**
   * Job 列表
   * @param {Object} params
   * @param {string} [params.namespace] - 命名空间（不传则查全部）
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - Job 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/job/list`, {params})
  },

  /**
   * Job 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Job 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/job/detail`, {params})
  },

  /**
   * 删除 Job
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Job 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/job/delete`, {params})
  },

  // =========================
  // 运行管理
  // =========================

  /**
   * 暂停/恢复 Job
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Job 名称
   * @param {boolean} data.suspend - true=暂停, false=恢复
   */
  suspend(data) {
    return http.put(`${K8S_BASE}/job/suspend`, data)
  },

  /**
   * 重启 Job（基于原Job模板重新创建）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Job 名称
   */
  restart(data) {
    return http.post(`${K8S_BASE}/job/restart`, data)
  },

  // =========================
  // 镜像更新
  // =========================

  /**
   * 更新 Job 镜像
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Job 名称
   * @param {string} data.container - 容器名称
   * @param {string} data.image - 新镜像地址
   */
  updateImage(data) {
    return http.put(`${K8S_BASE}/job/update-image`, data)
  },

  // =========================
  // YAML 操作
  // =========================

  /**
   * 获取 Job YAML
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - Job 名称
   */
  getYaml(params) {
    return http.get(`${K8S_BASE}/job/yaml`, { params })
  },

  /**
   * 应用 YAML 修改
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - Job 名称
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.put(`${K8S_BASE}/job/apply-yaml`, data)
  },
}

export default jobsApi

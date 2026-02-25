// src/api/cluster/workloads/cronjobs.js
import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

// =========================
// CronJob 模块 API
// 对应后端路由: /api/v1/k8s/cronjob/*
// =========================
const cronjobsApi = {
  // =========================
  // CronJob 基础 CRUD
  // =========================

  /**
   * 创建 CronJob
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - CronJob 名称
   * @param {string} data.schedule - Cron 表达式 (如 "0 * * * *")
   * @param {string} data.container_image - 容器镜像
   * @param {boolean} [data.suspend=false] - 是否暂停
   * @param {number} [data.successful_jobs_history_limit=3] - 成功任务历史限制
   * @param {number} [data.failed_jobs_history_limit=1] - 失败任务历史限制
   * @param {string} [data.concurrency_policy='Allow'] - 并发策略
   */
  create(data) {
    return http.post(`${K8S_BASE}/cronjob/create`, data)
  },

  /**
   * 从 YAML 创建 CronJob
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/cronjob/create-from-yaml`, data)
  },

  /**
   * 从 YAML 更新 CronJob
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  updateFromYaml(data) {
    return http.put(`${K8S_BASE}/cronjob/update-from-yaml`, data)
  },

  /**
   * CronJob 列表
   * @param {Object} params
   * @param {string} [params.namespace] - 命名空间（不传则查全部）
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - CronJob 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/cronjob/list`, {params})
  },

  /**
   * CronJob 详情
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - CronJob 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/cronjob/detail`, {params})
  },

  /**
   * 删除 CronJob
   * @param {Object} params
   * @param {string} params.namespace - 命名空间
   * @param {string} params.name - CronJob 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/cronjob/delete`, {params})
  },

  // =========================
  // 运行管理
  // =========================

  /**
   * 暂停/恢复 CronJob
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - CronJob 名称
   * @param {boolean} data.suspend - true=暂停, false=恢复
   */
  suspend(data) {
    return http.put(`${K8S_BASE}/cronjob/suspend`, data)
  },

  /**
   * 手动触发 CronJob（立即创建一个 Job）
   * @param {Object} data
   * @param {string} data.namespace - 命名空间
   * @param {string} data.name - CronJob 名称
   */
  trigger(data) {
    return http.post(`${K8S_BASE}/cronjob/trigger`, data)
  },
}

export default cronjobsApi

// src/api/cluster/storage/storageclasses.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

// =========================
// StorageClass 模块 API
// 对应后端路由: /api/v1/k8s/storageclass/*
// =========================
const storageclassesApi = {
  // =========================
  // StorageClass 基础 CRUD
  // =========================

  /**
   * 创建 StorageClass
   * @param {Object} data
   * @param {string} data.name - StorageClass 名称
   * @param {string} data.provisioner - 存储配置器
   * @param {Object} [data.parameters] - 参数
   * @param {string} [data.reclaimPolicy] - 回收策略 (Delete/Retain)
   * @param {string} [data.volumeBindingMode] - 绑定模式 (Immediate/WaitForFirstConsumer)
   * @param {boolean} [data.allowVolumeExpansion] - 是否允许扩容
   * @param {Array} [data.mountOptions] - 挂载选项
   */
  create(data) {
    return http.post(`${K8S_BASE}/storageclass/create`, data)
  },

  /**
   * 从 YAML 创建 StorageClass
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/storageclass/create-from-yaml`, data)
  },

  /**
   * StorageClass 列表
   * @param {Object} params
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - StorageClass 名称（模糊匹配）
   */
  list(params) {
    return http.get(`${K8S_BASE}/storageclass/list`, { params })
  },

  /**
   * StorageClass 详情
   * @param {Object} params
   * @param {string} params.name - StorageClass 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/storageclass/detail`, { params })
  },

  /**
   * 删除 StorageClass
   * @param {Object} params
   * @param {string} params.name - StorageClass 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/storageclass/delete`, { params })
  },

  // =========================
  // YAML 操作
  // =========================

  /**
   * 获取 StorageClass YAML
   * @param {Object} params
   * @param {string} params.name - StorageClass 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/storageclass/yaml`, { params })
  },

  /**
   * 应用 StorageClass YAML（创建/更新）
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.post(`${K8S_BASE}/storageclass/apply-yaml`, data)
  },

  // =========================
  // 辅助方法
  // =========================

  /**
   * 批量删除 StorageClass
   * @param {Array<string>} names - StorageClass 名称列表
   */
  batchDelete(names) {
    const promises = names.map(name =>
      this.delete({ name })
    )
    return Promise.all(promises)
  },

  /**
   * 下载 StorageClass YAML
   * @param {string} name - StorageClass 名称
   */
  async downloadYaml(name) {
    const res = await this.yaml({ name })
    if (res.code === 0 && res.data.yaml) {
      const blob = new Blob([res.data.yaml], { type: 'text/yaml' })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `storageclass-${name}.yaml`
      link.click()
      URL.revokeObjectURL(link.href)
    }
  },
}

export default storageclassesApi

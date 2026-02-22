// src/api/cluster/storage/pv.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

// =========================
// PersistentVolume 模块 API
// 对应后端路由: /api/v1/k8s/pv/*
// =========================
const pvApi = {
  // =========================
  // PV 基础 CRUD
  // =========================

  /**
   * 创建 PersistentVolume
   * @param {Object} data
   * @param {string} data.name - PV 名称
   * @param {string} data.capacity - 容量 (如 10Gi)
   * @param {Array<string>} data.accessModes - 访问模式 (ReadWriteOnce, ReadOnlyMany, ReadWriteMany)
   * @param {string} data.reclaimPolicy - 回收策略 (Delete/Retain/Recycle)
   * @param {string} [data.storageClassName] - StorageClass 名称
   * @param {Object} data.volumeSource - 卷来源配置 (hostPath, nfs, etc.)
   */
  create(data) {
    return http.post(`${K8S_BASE}/pv/create`, data)
  },

  /**
   * 从 YAML 创建 PersistentVolume
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/pv/create-from-yaml`, data)
  },

  /**
   * PersistentVolume 列表
   * @param {Object} params
   * @param {number} params.page - 页码
   * @param {number} params.limit - 每页数量
   * @param {string} [params.name] - PV 名称（模糊匹配）
   * @param {string} [params.status] - 状态过滤 (Available, Bound, Released, Failed)
   * @param {string} [params.storageClassName] - StorageClass 过滤
   */
  list(params) {
    return http.get(`${K8S_BASE}/pv/list`, { params })
  },

  /**
   * PersistentVolume 详情
   * @param {Object} params
   * @param {string} params.name - PV 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/pv/detail`, { params })
  },

  /**
   * 删除 PersistentVolume
   * @param {Object} params
   * @param {string} params.name - PV 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/pv/delete`, { params })
  },

  /**
   * 修改 PersistentVolume 回收策略
   * @param {Object} data
   * @param {string} data.name - PV 名称
   * @param {string} data.reclaimPolicy - 新的回收策略 (Delete/Retain)
   */
  reclaim(data) {
    return http.patch(`${K8S_BASE}/pv/reclaim`, data)
  },

  /**
   * PV 扩容
   * @param {Object} data
   * @param {string} data.name - PV 名称
   * @param {string} data.newCapacity - 新容量 (如 20Gi，只能扩大不能缩小)
   */
  expand(data) {
    return http.post(`${K8S_BASE}/pv/expand`, data)
  },

  // =========================
  // YAML 操作
  // =========================

  /**
   * 获取 PersistentVolume YAML
   * @param {Object} params
   * @param {string} params.name - PV 名称
   */
  yaml(params) {
    return http.get(`${K8S_BASE}/pv/yaml`, { params })
  },

  /**
   * 应用 PersistentVolume YAML（创建/更新）
   * @param {Object} data
   * @param {string} data.yaml - YAML 内容
   */
  applyYaml(data) {
    return http.post(`${K8S_BASE}/pv/apply-yaml`, data)
  },

  // =========================
  // 辅助方法
  // =========================

  /**
   * 批量删除 PersistentVolume
   * @param {Array<string>} names - PV 名称列表
   */
  batchDelete(names) {
    const promises = names.map(name =>
      this.delete({ name })
    )
    return Promise.all(promises)
  },

  /**
   * 下载 PersistentVolume YAML
   * @param {string} name - PV 名称
   */
  async downloadYaml(name) {
    const res = await this.yaml({ name })
    if (res.code === 0 && res.data.yaml) {
      const blob = new Blob([res.data.yaml], { type: 'text/yaml' })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `pv-${name}.yaml`
      link.click()
      URL.revokeObjectURL(link.href)
    }
  },
}

export default pvApi

// src/api/cluster/storage/storageclass.js
import http from '@/api/http'
import { K8S_BASE } from '@/api/paths'

/**
 * StorageClass API
 * 存储类管理接口
 */
const storageclassApi = {
  /**
   * 获取 StorageClass 列表
   * @param {Object} params
   * @param {string} [params.name] - 名称（模糊匹配）
   * @param {number} [params.page=1] - 页码
   * @param {number} [params.limit=100] - 每页数量
   */
  list(params = {}) {
    return http.get(`${K8S_BASE}/storageclass/list`, { params })
  },

  /**
   * 获取 StorageClass 详情
   * @param {Object} params
   * @param {string} params.name - StorageClass 名称
   */
  detail(params) {
    return http.get(`${K8S_BASE}/storageclass/detail`, { params })
  },

  /**
   * 创建 StorageClass
   * @param {Object} data
   * @param {string} data.name - 名称
   * @param {string} data.provisioner - 存储供应商
   * @param {string} [data.reclaim_policy] - 回收策略 (Delete/Retain)
   * @param {string} [data.volume_binding_mode] - 卷绑定模式
   * @param {boolean} [data.allow_volume_expansion] - 是否允许扩容
   */
  create(data) {
    return http.post(`${K8S_BASE}/storageclass/create`, data)
  },

  /**
   * 删除 StorageClass
   * @param {Object} params
   * @param {string} params.name - StorageClass 名称
   */
  delete(params) {
    return http.delete(`${K8S_BASE}/storageclass/delete`, { params })
  }
}

export default storageclassApi

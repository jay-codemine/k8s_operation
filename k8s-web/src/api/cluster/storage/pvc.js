import request from '@/api/http'
import { K8S_BASE } from '@/api/paths'

/**
 * PVC API
 */
export default {
  /**
   * 获取 PVC 列表（支持分页、模糊匹配）
   */
  list(params) {
    return request({
      url: `${K8S_BASE}/pvc/list`,
      method: 'get',
      params
    })
  },

  /**
   * 获取 PVC 详情
   */
  detail(params) {
    return request({
      url: `${K8S_BASE}/pvc/detail`,
      method: 'get',
      params
    })
  },

  /**
   * 创建 PVC（表单模式）
   */
  create(data) {
    return request({
      url: `${K8S_BASE}/pvc/create`,
      method: 'post',
      data
    })
  },

  /**
   * 从 YAML 创建 PVC
   */
  createFromYaml(data) {
    return request({
      url: `${K8S_BASE}/pvc/create-from-yaml`,
      method: 'post',
      data
    })
  },

  /**
   * 应用 YAML 更改
   */
  applyYaml(data) {
    return request({
      url: `${K8S_BASE}/pvc/apply-yaml`,
      method: 'put',
      data
    })
  },

  /**
   * 删除 PVC
   */
  delete(params) {
    return request({
      url: `${K8S_BASE}/pvc/delete`,
      method: 'delete',
      params
    })
  },

  /**
   * 扩容 PVC
   */
  resize(data) {
    return request({
      url: `${K8S_BASE}/pvc/resize`,
      method: 'patch',
      data
    })
  }
}

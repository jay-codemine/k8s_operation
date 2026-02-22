import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

const namespaceApi = {
  // Namespace 列表
  list(params) {
    return http.get(`${K8S_BASE}/namespace/list`, {params})
  },

  // Namespace 详情
  detail(params) {
    // params: { name }
    return http.get(`${K8S_BASE}/namespace/detail`, {params})
  },

  // 创建 Namespace
  create(data) {
    // data: { name, labels?, annotations? }
    return http.post(`${K8S_BASE}/namespace/create`, data)
  },

  // 删除 Namespace
  delete(params) {
    // params: { name }
    return http.delete(`${K8S_BASE}/namespace/delete`, {params})
  },

  // Patch Namespace（labels / annotations）
  patch(data) {
    return http.patch(`${K8S_BASE}/namespace/patch`, data)
  },
}

export default namespaceApi

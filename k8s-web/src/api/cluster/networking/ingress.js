import http from '@/api/http'
import {K8S_BASE} from '@/api/paths'

const BASE = `${K8S_BASE}/ingress`

const ingressApi = {
  // 创建 Ingress
  create(data) {
    return http.post(`${BASE}/create`, data)
  },

  // 从 YAML 创建 Ingress
  createFromYaml(data) {
    return http.post(`${BASE}/create-from-yaml`, data)
  },

  // 获取 Ingress 列表
  list(params) {
    return http.get(`${BASE}/list`, { params })
  },

  // 获取 Ingress 详情
  detail(params) {
    return http.get(`${BASE}/detail`, { params })
  },

  // 删除 Ingress
  delete(params) {
    return http.delete(`${BASE}/delete`, { params })
  },

  // 获取 Ingress YAML
  yaml(params) {
    return http.get(`${BASE}/yaml`, { params })
  },

  // 应用 YAML 更新
  applyYaml(data) {
    return http.put(`${BASE}/apply-yaml`, data)
  },

  // Strategic Merge Patch 更新
  patch(data) {
    return http.patch(`${BASE}/patch`, data)
  },

  // JSON Merge Patch 更新
  patchJson(data) {
    return http.post(`${BASE}/patch_json`, data)
  }
}

export default ingressApi

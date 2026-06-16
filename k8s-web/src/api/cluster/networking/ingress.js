import http from '@/api/http'

const K8S_BASE = '/api/v1/k8s/ingress'

const ingressApi = {
  // 创建 Ingress
  create(data) {
    return http.post(`${K8S_BASE}/create`, data)
  },

  // 从 YAML 创建 Ingress
  createFromYaml(data) {
    return http.post(`${K8S_BASE}/create-from-yaml`, data)
  },

  // 获取 Ingress 列表
  list(params) {
    return http.get(`${K8S_BASE}/list`, { params })
  },

  // 获取 Ingress 详情
  detail(params) {
    return http.get(`${K8S_BASE}/detail`, { params })
  },

  // 删除 Ingress
  delete(params) {
    return http.delete(`${K8S_BASE}/delete`, { params })
  },

  // 获取 Ingress YAML
  yaml(params) {
    return http.get(`${K8S_BASE}/yaml`, { params })
  },

  // 应用 YAML 更新
  applyYaml(data) {
    return http.put(`${K8S_BASE}/apply-yaml`, data)
  },

  // Strategic Merge Patch 更新
  patch(data) {
    return http.patch(`${K8S_BASE}/patch`, data)
  },

  // JSON Merge Patch 更新
  patchJson(data) {
    return http.post(`${K8S_BASE}/patch_json`, data)
  }
}

export default ingressApi

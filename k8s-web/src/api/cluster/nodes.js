import http from '../http'

const K8S_BASE = '/api/v1/k8s'

const nodesApi = {
  // Node 列表
  list(params) {
    return http.get(`${K8S_BASE}/node/list`, { params })
  },

  // Node 详情
  detail(params) {
    return http.get(`${K8S_BASE}/node/detail`, { params })
  },

  // Node 上的 Pods
  listPods(params) {
    return http.get(`${K8S_BASE}/node/pods`, { params })
  },

  // Node 指标（CPU/内存使用率）
  metrics(params) {
    return http.get(`${K8S_BASE}/node/metrics`, { params })
  },

  // Node 事件
  events(params) {
    return http.get(`${K8S_BASE}/node/events`, { params })
  },

  // Cordon/Uncordon
  cordon(data) {
    return http.post(`${K8S_BASE}/node/cordon`, data)
  },

  // Drain（驱逐 Pod）
  drain(data) {
    return http.post(`${K8S_BASE}/node/drain`, data)
  },

  // 修改节点标签
  patchLabels(data) {
    return http.patch(`${K8S_BASE}/node/labels`, data)
  },

  // 修改节点污点
  patchTaints(data) {
    return http.patch(`${K8S_BASE}/node/taints`, data)
  },

  // 获取节点 YAML
  yaml(params) {
    return http.get(`${K8S_BASE}/node/yaml`, { params })
  },

  // 应用节点 YAML
  applyYaml(data) {
    return http.put(`${K8S_BASE}/node/apply_yaml`, data)
  },
}

export default nodesApi

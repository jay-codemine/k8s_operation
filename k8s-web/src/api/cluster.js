// src/api/cluster.js
import http from './http'
import { API_BASE } from './paths'

// 列出 K8s 集群
export function getClusterList(params) {
  return http.get(`${API_BASE}/k8s/cluster/list`, {
    params,
  })
}

// 创建 K8s 集群
export function createCluster(data) {
  return http.post(`${API_BASE}/k8s/cluster/create`, data)
}

// 修改 K8s 集群
export function updateCluster(data) {
  return http.post(`${API_BASE}/k8s/cluster/update`, data)
}

// 删除 K8s 集群
export function deleteCluster(data) {
  return http.post(`${API_BASE}/k8s/cluster/delete`, data)
}

// 初始化 K8s 集群
export function initCluster(data) {
  return http.post(`${API_BASE}/k8s/cluster/init`, data)
}

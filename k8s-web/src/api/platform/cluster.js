// src/api/cicd.js

import http from "@/api/http.js";
import {API_BASE} from "@/api/paths.js";
// =======================
// K8s集群管理（真实后端接口）
// 对应：/api/v1/k8s/cluster/*
// =======================

// 创建 K8s 集群
// payload: { cluster_name, kube_config, cluster_version}
export const createK8sCluster = (payload) => {
  return http.post(`${API_BASE}/k8s/cluster/create`, payload)
}

// 更新 K8s 集群（✅ 统一为对象入参，避免视图/接口签名不一致）
// payload: { id, cluster_name?, kube_config?, cluster_version?, set_as_default? }
export const updateK8sCluster = (payload) => {
  if (!payload || payload.id === undefined || payload.id === null) {
    throw new Error('updateK8sCluster: payload.id is required')
  }
  return http.post(`${API_BASE}/k8s/cluster/update`, payload)
}

// 删除 K8s 集群（✅ 统一为对象入参）
// payload: { id }
export const deleteK8sCluster = (payload) => {
  const id = payload?.id ?? payload
  if (id === undefined || id === null) {
    throw new Error('deleteK8sCluster: id is required')
  }
  return http.post(`${API_BASE}/k8s/cluster/delete`, {id})
}

// 集群列表（✅ 统一命名，避免视图 import 找不到）
// params: { cluster_name?, page?, limit? }
export const getK8sClusterList = (params) => {
  return http.get(`${API_BASE}/k8s/cluster/list`, {params})
}

// 初始化集群（可选）
export const initK8sCluster = (data) => {
  return http.post(`${API_BASE}/k8s/cluster/init`, data)
}

// /**
//  * ===== 兼容导出（可选，但强烈建议保留一段时间避免别的页面炸）=====
//  * 如果你项目其它地方还在用旧名字，就先保留：
//  */
// export const getK8sClusters = getK8sClusterList

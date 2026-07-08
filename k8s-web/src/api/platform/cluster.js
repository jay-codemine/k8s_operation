// src/api/platform/cluster.js
// 集群管理 API — 统一适配层，委托给 @/api/cluster.js（唯一权威实现）
import {
  getClusterList,
  createCluster,
  updateCluster,
  deleteCluster,
  batchDeleteCluster,
  initCluster,
} from '@/api/cluster'

// 别名导出（保持向后兼容，避免大量文件 import 报错）
export const getK8sClusterList = getClusterList

export const createK8sCluster = (payload) => createCluster(payload)

export const updateK8sCluster = (payload) => {
  if (!payload || payload.id === undefined || payload.id === null) {
    throw new Error('updateK8sCluster: payload.id is required')
  }
  return updateCluster(payload, { _silent: true })
}

export const deleteK8sCluster = (payload) => {
  const id = payload?.id ?? payload
  if (id === undefined || id === null) {
    throw new Error('deleteK8sCluster: id is required')
  }
  return deleteCluster({ id })
}

export const initK8sCluster = (data) => initCluster(data)

export const batchDeleteK8sCluster = (data) => batchDeleteCluster(data)

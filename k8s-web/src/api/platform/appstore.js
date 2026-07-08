// src/api/platform/appstore.js
// 应用商城 API

import http from '@/api/http.js'
import { API_BASE } from '@/api/paths.js'

const BASE = `${API_BASE}/platform/appstore`

/** 获取应用列表（分页 + 筛选） */
export const getAppStoreList = (params) => {
  return http.get(`${BASE}/list`, { params })
}

/** 获取应用详情 */
export const getAppStoreDetail = (id) => {
  return http.get(`${BASE}/detail/${id}`)
}

/** 获取分类列表 */
export const getAppStoreCategories = () => {
  return http.get(`${BASE}/categories`)
}

/** 创建应用 */
export const createAppStoreApp = (payload) => {
  return http.post(`${BASE}/create`, payload)
}

/** 更新应用 */
export const updateAppStoreApp = (payload) => {
  return http.put(`${BASE}/update`, payload)
}

/** 删除应用 */
export const deleteAppStoreApp = (id) => {
  return http.delete(`${BASE}/delete/${id}`)
}

// ====== 安装管理 ======

/** 安装应用到集群 */
export const installApp = (payload) => {
  return http.post(`${BASE}/install`, payload)
}

/** 卸载应用 */
export const uninstallApp = (id) => {
  return http.post(`${BASE}/uninstall/${id}`)
}

/** 获取安装记录列表 */
export const getInstallList = (params) => {
  return http.get(`${BASE}/installs`, { params })
}

/** 获取安装记录详情 */
export const getInstallDetail = (id) => {
  return http.get(`${BASE}/installs/${id}`)
}

/** 获取安装实时状态（K8s 集群 Pod/Deployment/Service 状态） */
export const getInstallStatus = (id) => {
  return http.get(`${BASE}/installs/${id}/status`)
}

/** 编辑安装（更新 Deployment 参数：副本数/镜像/资源限制） */
export const updateInstall = (id, payload) => {
  return http.put(`${BASE}/installs/${id}/update`, payload)
}

// ====== 组件管理 ======

/** 获取应用的组件列表 */
export const getComponentList = (appId) => {
  return http.get(`${BASE}/components/${appId}`)
}

/** 创建组件 */
export const createComponent = (payload) => {
  return http.post(`${BASE}/components/create`, payload)
}

/** 更新组件 */
export const updateComponent = (payload) => {
  return http.put(`${BASE}/components/update`, payload)
}

/** 删除组件 */
export const deleteComponent = (compId) => {
  return http.delete(`${BASE}/components/delete/${compId}`)
}

/** 批量删除组件 */
export const batchDeleteComponents = (ids) => {
  return http.post(`${BASE}/components/batch-delete`, { ids })
}

/** 更新组件排序 */
export const sortComponents = (items) => {
  return http.put(`${BASE}/components/sort`, { items })
}

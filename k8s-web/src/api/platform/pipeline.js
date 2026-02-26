// k8s-web/src/api/platform/pipeline.js
// 流水线管理 API - 调用真实后端接口

import http from '../http'

const BASE_URL = '/api/v1/k8s/cicd/pipeline'

/**
 * 获取流水线列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码，默认1
 * @param {number} params.page_size - 每页数量，默认10
 * @param {string} params.keyword - 搜索关键字（名称/描述/Git仓库）
 * @param {string} params.status - 状态筛选（idle/running/disabled）
 */
export const getPipelines = (params = {}) => {
  return http.get(`${BASE_URL}/list`, { params })
}

/**
 * 获取流水线详情
 * @param {number} id - 流水线ID
 */
export const getPipelineDetail = (id) => {
  return http.get(`${BASE_URL}/detail`, { params: { id } })
}

/**
 * 创建流水线
 * @param {Object} data - 创建参数
 * @param {string} data.name - 流水线名称
 * @param {string} data.description - 描述
 * @param {string} data.git_repo - Git仓库URL
 * @param {string} data.git_branch - Git分支，默认main
 * @param {string} data.jenkins_url - Jenkins服务器地址
 * @param {string} data.jenkins_job - Jenkins Job名称
 * @param {Array} data.env_vars - 环境变量 [{name, value}]
 * @param {Object} data.deploy_config - 部署配置
 */
export const createPipeline = (data) => {
  return http.post(`${BASE_URL}/create`, data)
}

/**
 * 更新流水线
 * @param {Object} data - 更新参数
 * @param {number} data.id - 流水线ID
 * @param {string} data.name - 流水线名称
 * @param {string} data.description - 描述
 * @param {string} data.git_repo - Git仓库URL
 * @param {string} data.git_branch - Git分支
 * @param {string} data.jenkins_url - Jenkins服务器地址
 * @param {string} data.jenkins_job - Jenkins Job名称
 * @param {string} data.status - 状态
 * @param {Array} data.env_vars - 环境变量
 * @param {Object} data.deploy_config - 部署配置
 */
export const updatePipeline = (data) => {
  return http.post(`${BASE_URL}/update`, data)
}

/**
 * 删除流水线
 * @param {number} id - 流水线ID
 */
export const deletePipeline = (id) => {
  return http.post(`${BASE_URL}/delete`, { id })
}

/**
 * 运行流水线（触发Jenkins构建）
 * @param {number} id - 流水线ID
 * @param {Object} options - 可选参数
 * @param {string} options.branch - 覆盖默认分支
 * @param {Object} options.env_vars - 覆盖环境变量 {KEY: VALUE}
 */
export const runPipeline = (id, options = {}) => {
  return http.post(`${BASE_URL}/run`, { id, ...options })
}

/**
 * 停止流水线
 * @param {number} id - 流水线ID
 * @param {number} buildNumber - 可选：指定构建号
 */
export const stopPipeline = (id, buildNumber = null) => {
  const data = { id }
  if (buildNumber) {
    data.build_number = buildNumber
  }
  return http.post(`${BASE_URL}/stop`, data)
}

/**
 * 获取流水线日志
 * @param {number} id - 流水线ID
 * @param {number} buildNumber - 可选：指定构建号
 * @param {number} startLine - 可选：起始行号（增量获取）
 */
export const getPipelineLogs = (id, buildNumber = null, startLine = 0) => {
  const params = { id }
  if (buildNumber) {
    params.build_number = buildNumber
  }
  if (startLine > 0) {
    params.start_line = startLine
  }
  return http.get(`${BASE_URL}/logs`, { params })
}

/**
 * 获取流水线实时状态
 * @param {number} id - 流水线ID
 */
export const getPipelineStatus = (id) => {
  return http.get(`${BASE_URL}/status`, { params: { id } })
}

/**
 * 获取流水线运行历史
 * @param {number} id - 流水线ID
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 */
export const getPipelineHistory = (id, page = 1, pageSize = 10) => {
  return http.get(`${BASE_URL}/history`, {
    params: { id, page, page_size: pageSize }
  })
}

// ==================== 兼容旧API名称（方便迁移） ====================

// 旧API名称映射
export const getPipelineById = getPipelineDetail
export const triggerPipeline = runPipeline
export const cancelPipeline = stopPipeline

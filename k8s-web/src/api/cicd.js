// src/api/cicd.js
import http from './http'
import { API_BASE } from './paths'

// =======================
// K8s集群管理（真实后端接口）
// 对应 Swagger：/api/v1/k8s/cluster/*
// =======================

// 创建 K8s 集群
export const createK8sCluster = (clusterData) => {
  return http.post(`${API_BASE}/k8s/cluster/create`, clusterData)
}

// 更新 K8s 集群
export const updateK8sCluster = (id, clusterData) => {
  return http.post(`${API_BASE}/k8s/cluster/update`, { id, ...clusterData })
}

// 删除 K8s 集群
export const deleteK8sCluster = (id) => {
  return http.post(`${API_BASE}/k8s/cluster/delete`, { id })
}

// 集群列表
export const getK8sClusters = (params) => {
  return http.get(`${API_BASE}/k8s/cluster/list`, { params })
}

// 初始化集群
export const initK8sCluster = (data) => {
  return http.post(`${API_BASE}/k8s/cluster/init`, data)
}

// =======================
// CI/CD 流水线管理（统一使用 platform/pipeline.js）
// 对应后端路由: /api/v1/k8s/cicd/pipeline/*
// =======================

// 统一从 pipeline.js 导出，避免重复定义
export {
  getPipelines,
  getPipelineDetail,
  createPipeline,
  updatePipeline,
  deletePipeline,
  runPipeline,
  stopPipeline,
  getPipelineLogs,
  getPipelineStatus,
  getPipelineHistory
} from './platform/pipeline'

/**
 * 获取流水线模板列表（TODO: 后端需实现）
 */
export const getPipelineTemplates = () => {
  return Promise.resolve({ code: 0, msg: 'success', data: [] })
}

// 部署到K8s（兼容旧接口）
export { runPipeline as deployToK8s } from './platform/pipeline'

// 获取部署历史（兼容旧接口）
export { getPipelineHistory as getDeploymentHistory } from './platform/pipeline'

// =======================
// CI/CD 发布单管理（CICD Release）
// 对应后端路由: /api/v1/k8s/cicd/release/*
// =======================

const RELEASE_BASE = `${API_BASE}/k8s/cicd/release`

/**
 * 获取发布单列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.keyword - 搜索关键字
 * @param {string} params.status - 状态筛选
 */
export const getReleases = (params = {}) => {
  return http.get(`${RELEASE_BASE}/list`, { params })
}

/**
 * 获取发布单详情
 * @param {number} id - 发布单ID
 */
export const getReleaseDetail = (id) => {
  return http.get(`${RELEASE_BASE}/detail`, { params: { id } })
}

/**
 * 创建发布单
 * @param {Object} data - 创建参数
 * @param {number} data.pipeline_id - 流水线ID
 * @param {string} data.name - 发布单名称
 * @param {string} data.description - 描述
 * @param {string} data.version - 版本号
 * @param {string} data.image - 镜像地址
 * @param {string} data.namespace - 目标命名空间
 * @param {Object} data.deploy_config - 部署配置
 */
export const createRelease = (data) => {
  return http.post(`${RELEASE_BASE}/create`, data)
}

/**
 * 取消发布单
 * @param {number} id - 发布单ID
 */
export const cancelRelease = (id) => {
  return http.post(`${RELEASE_BASE}/cancel`, { id })
}

/**
 * 重试发布单
 * @param {number} id - 发布单ID
 */
export const retryRelease = (id) => {
  return http.post(`${RELEASE_BASE}/retry`, { id })
}

/**
 * 回滚发布单
 * @param {number} id - 发布单ID
 * @param {number} target_version - 目标版本（可选）
 */
export const rollbackRelease = (id, targetVersion = null) => {
  const data = { id }
  if (targetVersion) {
    data.target_version = targetVersion
  }
  return http.post(`${RELEASE_BASE}/rollback`, data)
}

/**
 * 获取发布单下的任务列表
 * @param {number} id - 发布单ID
 */
export const getReleaseTasks = (id) => {
  return http.get(`${RELEASE_BASE}/tasks`, { params: { id } })
}

// =======================
// CI/CD 回调接口（CICD Callback）
// 对应后端路由: /api/v1/k8s/cicd/callback/*
// =======================

const CALLBACK_BASE = `${API_BASE}/k8s/cicd/callback`

/**
 * Jenkins 构建回调（通常由 Jenkins 调用，前端一般不使用）
 * @param {Object} data - 回调数据
 * @param {string} data.job_name - Jenkins Job名称
 * @param {number} data.build_number - 构建号
 * @param {string} data.status - 构建状态 (SUCCESS/FAILURE/ABORTED)
 * @param {number} data.duration - 构建时长(毫秒)
 * @param {string} data.message - 构建信息
 */
export const jenkinsBuildCallback = (data) => {
  return http.post(`${CALLBACK_BASE}/build`, data)
}

// Pipeline 回调（兼容旧路径）
export const pipelineCallback = (data) => {
  return http.post(`${API_BASE}/k8s/cicd/pipeline/callback`, data)
}

// =======================
// Git 仓库操作
// =======================

/**
 * 获取 Git 仓库的远程分支列表
 * @param {string} repoUrl - Git 仓库地址
 * @param {string} credentialId - 凭证ID（可选）
 */
export const getGitBranches = (repoUrl, credentialId = '') => {
  return http.post(`${API_BASE}/k8s/cicd/git/branches`, {
    repo_url: repoUrl,
    credential_id: credentialId
  })
}

/**
 * 验证 Git 仓库连接
 * @param {string} repoUrl - Git 仓库地址
 * @param {string} credentialId - 凭证ID（可选）
 */
export const validateGitRepo = (repoUrl, credentialId = '') => {
  return http.post(`${API_BASE}/k8s/cicd/git/validate`, {
    repo_url: repoUrl,
    credential_id: credentialId
  })
}

// =======================
// K8s环境管理（静态Mock数据）
// =======================

const mockK8sEnvironments = [
  { id: 1, name: '开发环境', description: '开发测试集群', clusterName: 'dev-cluster', apiUrl: 'https://dev-k8s.example.com', namespace: 'dev', type: 'development', status: 'connected' },
  { id: 2, name: '测试环境', description: '集成测试集群', clusterName: 'test-cluster', apiUrl: 'https://test-k8s.example.com', namespace: 'test', type: 'testing', status: 'connected' },
  { id: 3, name: '生产环境', description: '生产集群', clusterName: 'prod-cluster', apiUrl: 'https://prod-k8s.example.com', namespace: 'production', type: 'production', status: 'connected' }
]

export const getK8sEnvironments = () => {
  return Promise.resolve({ code: 0, msg: 'success', data: mockK8sEnvironments })
}

export const createK8sEnvironment = (data) => {
  return Promise.resolve({ code: 0, msg: '创建成功', data: { id: Date.now(), ...data } })
}

export const updateK8sEnvironment = (id, data) => {
  return Promise.resolve({ code: 0, msg: '更新成功', data: { id, ...data } })
}

export const deleteK8sEnvironment = (id) => {
  return Promise.resolve({ code: 0, msg: '删除成功' })
}

export const getK8sEnvironmentDetail = (id) => {
  const env = mockK8sEnvironments.find(e => e.id === parseInt(id)) || mockK8sEnvironments[0]
  return Promise.resolve({ code: 0, msg: 'success', data: env })
}

// =======================
// 镜像仓库管理（静态Mock数据）
// =======================

const mockImageRepositories = [
  { id: 1, name: 'Docker Hub', type: 'docker', url: 'https://registry.hub.docker.com', status: 'connected' },
  { id: 2, name: 'Harbor私有仓库', type: 'harbor', url: 'https://harbor.example.com', status: 'connected' },
  { id: 3, name: 'Aliyun ACR', type: 'acr', url: 'https://registry.cn-hangzhou.aliyuncs.com', status: 'disconnected' }
]

const mockImages = [
  { id: 1, name: 'nginx', tags: ['latest', '1.21', '1.20', '1.19'], size: '133MB', lastUpdated: '2024-01-10' },
  { id: 2, name: 'redis', tags: ['latest', '7.0', '6.2'], size: '117MB', lastUpdated: '2024-01-08' },
  { id: 3, name: 'mysql', tags: ['latest', '8.0', '5.7'], size: '446MB', lastUpdated: '2024-01-05' }
]

export const getImageRepositories = () => {
  return Promise.resolve({ code: 0, msg: 'success', data: mockImageRepositories })
}

export const createImageRepository = (data) => {
  return Promise.resolve({ code: 0, msg: '创建成功', data: { id: Date.now(), ...data } })
}

export const updateImageRepository = (id, data) => {
  return Promise.resolve({ code: 0, msg: '更新成功', data: { id, ...data } })
}

export const deleteImageRepository = (id) => {
  return Promise.resolve({ code: 0, msg: '删除成功' })
}

export const getImages = (repoId) => {
  return Promise.resolve({ code: 0, msg: 'success', data: mockImages })
}

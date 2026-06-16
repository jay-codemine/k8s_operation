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
 * 获取流水线模板列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.keyword - 搜索关键字
 * @param {string} params.type - 类型筛选
 */
export const getPipelineTemplates = (params = {}) => {
  return http.get(`${API_BASE}/k8s/cicd/template/list`, { params })
}

/**
 * 获取流水线模板详情
 * @param {number} id - 模板ID
 */
export const getPipelineTemplateDetail = (id) => {
  return http.get(`${API_BASE}/k8s/cicd/template/detail`, { params: { id } })
}

/**
 * 创建流水线模板
 * @param {Object} data - 创建参数
 */
export const createPipelineTemplate = (data) => {
  return http.post(`${API_BASE}/k8s/cicd/template/create`, data)
}

/**
 * 更新流水线模板
 * @param {Object} data - 更新参数
 */
export const updatePipelineTemplate = (data) => {
  return http.post(`${API_BASE}/k8s/cicd/template/update`, data)
}

/**
 * 删除流水线模板
 * @param {number} id - 模板ID
 */
export const deletePipelineTemplate = (id) => {
  return http.post(`${API_BASE}/k8s/cicd/template/delete`, { id })
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
 * 获取发布单统计
 */
export const getReleaseStats = () => {
  return http.get(`${RELEASE_BASE}/stats`)
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
 * 编辑发布单
 * @param {Object} data - 更新参数
 * @param {number} data.id - 发布单ID
 */
export const updateRelease = (data) => {
  return http.post(`${RELEASE_BASE}/update`, data)
}

/**
 * 删除发布单
 * @param {number} id - 发布单ID
 */
export const deleteRelease = (id) => {
  return http.post(`${RELEASE_BASE}/delete`, { id })
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
 * 批量重新发布
 * @param {number[]} ids - 发布单ID列表
 */
export const batchRetryRelease = (ids) => {
  return http.post(`${RELEASE_BASE}/batch-retry`, { ids })
}

/**
 * 批量回滚发布单
 * @param {number[]} ids - 发布单ID列表
 */
export const batchRollbackRelease = (ids) => {
  return http.post(`${RELEASE_BASE}/batch-rollback`, { ids })
}

/**
 * 同步流水线运行记录到发布管理
 */
export const syncReleasesFromPipeline = () => {
  return http.post(`${RELEASE_BASE}/sync-from-pipeline`)
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
// K8s环境管理（真实后端接口）
// 对应后端路由: /api/v1/k8s/cicd/environment/*
// =======================

const ENVIRONMENT_BASE = `${API_BASE}/k8s/cicd/environment`

/**
 * 获取 K8s 部署环境列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.keyword - 搜索关键字
 */
export const getK8sEnvironments = (params = {}) => {
  return http.get(`${ENVIRONMENT_BASE}/list`, { params })
}

/**
 * 创建 K8s 部署环境
 * @param {Object} data - 环境数据
 * @param {string} data.name - 环境名称
 * @param {string} data.description - 描述
 * @param {number} data.cluster_id - 关联集群ID
 * @param {string} data.namespace - 默认命名空间
 * @param {string} data.type - 环境类型 (development/testing/staging/production)
 */
export const createK8sEnvironment = (data) => {
  return http.post(`${ENVIRONMENT_BASE}/create`, data)
}

/**
 * 更新 K8s 部署环境
 * @param {number} id - 环境ID
 * @param {Object} data - 更新数据
 */
export const updateK8sEnvironment = (id, data) => {
  return http.post(`${ENVIRONMENT_BASE}/update`, { id, ...data })
}

/**
 * 删除 K8s 部署环境
 * @param {number} id - 环境ID
 */
export const deleteK8sEnvironment = (id) => {
  return http.post(`${ENVIRONMENT_BASE}/delete`, { id })
}

/**
 * 获取 K8s 部署环境详情
 * @param {number} id - 环境ID
 */
export const getK8sEnvironmentDetail = (id) => {
  return http.get(`${ENVIRONMENT_BASE}/detail`, { params: { id } })
}

// =======================
// 镜像仓库管理（真实后端接口）
// 对应后端路由: /api/v1/image/registry/*
// =======================

const IMAGE_REGISTRY_BASE = `${API_BASE}/image/registry`
const IMAGE_BROWSE_BASE = `${API_BASE}/image/browse`

/**
 * 获取镜像仓库列表（分页）
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.keyword - 搜索关键字
 */
export const getImageRepositories = (params = {}) => {
  return http.get(`${IMAGE_REGISTRY_BASE}/list`, { params })
}

/**
 * 获取所有镜像仓库（下拉选择用）
 */
export const getAllImageRepositories = () => {
  return http.get(`${IMAGE_REGISTRY_BASE}/all`)
}

/**
 * 创建镜像仓库
 * @param {Object} data - 仓库数据
 * @param {string} data.name - 仓库名称
 * @param {string} data.type - 类型 (docker/harbor/acr)
 * @param {string} data.url - 仓库URL
 * @param {string} data.username - 用户名
 * @param {string} data.password - 密码
 */
export const createImageRepository = (data) => {
  return http.post(`${IMAGE_REGISTRY_BASE}/create`, data)
}

/**
 * 更新镜像仓库
 * @param {number} id - 仓库ID
 * @param {Object} data - 更新数据
 */
export const updateImageRepository = (id, data) => {
  return http.post(`${IMAGE_REGISTRY_BASE}/update`, { id, ...data })
}

/**
 * 删除镜像仓库
 * @param {number} id - 仓库ID
 */
export const deleteImageRepository = (id) => {
  return http.post(`${IMAGE_REGISTRY_BASE}/delete`, { id })
}

/**
 * 获取镜像仓库详情
 * @param {number} id - 仓库ID
 */
export const getImageRepositoryDetail = (id) => {
  return http.get(`${IMAGE_REGISTRY_BASE}/detail`, { params: { id } })
}

/**
 * 检查镜像仓库连接
 * @param {number} id - 仓库ID
 */
export const checkImageRepositoryConnection = (id) => {
  return http.post(`${IMAGE_REGISTRY_BASE}/check`, { id })
}

/**
 * 获取镜像仓库统计信息
 */
export const getImageRepositoryStats = () => {
  return http.get(`${IMAGE_REGISTRY_BASE}/stats`)
}

/**
 * 设置默认镜像仓库
 * @param {number} id - 仓库ID
 */
export const setDefaultImageRepository = (id) => {
  return http.post(`${IMAGE_REGISTRY_BASE}/default`, { id })
}

// =======================
// 镜像浏览管理（真实后端接口）
// 对应后端路由: /api/v1/image/browse/*
// =======================

/**
 * 获取镜像列表（从指定仓库）
 * @param {number} registryId - 仓库ID
 * @param {Object} params - 查询参数
 */
export const getImages = (registryId, params = {}) => {
  return http.get(`${IMAGE_BROWSE_BASE}/repositories`, { 
    params: { registry_id: registryId, ...params } 
  })
}

/**
 * 获取镜像标签列表
 * @param {number} registryId - 仓库ID
 * @param {string} repository - 镜像名称
 */
export const getImageTags = (registryId, repository) => {
  return http.get(`${IMAGE_BROWSE_BASE}/tags`, { 
    params: { registry_id: registryId, repository } 
  })
}

/**
 * 获取镜像详情
 * @param {number} registryId - 仓库ID
 * @param {string} repository - 镜像名称
 * @param {string} tag - 标签
 */
export const getImageDetail = (registryId, repository, tag) => {
  return http.get(`${IMAGE_BROWSE_BASE}/detail`, { 
    params: { registry_id: registryId, repository, tag } 
  })
}

/**
 * 删除镜像
 * @param {number} registryId - 仓库ID
 * @param {string} repository - 镜像名称
 * @param {string} tag - 标签（可选，不传则删除整个仓库）
 */
export const deleteImage = (registryId, repository, tag = '') => {
  return http.post(`${IMAGE_BROWSE_BASE}/delete`, { 
    registry_id: registryId, 
    repository,
    tag 
  })
}

// =======================
// CICD 资源配置管理
// =======================

const RESOURCE_BASE = `${API_BASE}/k8s/cicd/resource`

/**
 * 获取资源模板列表
 * @param {Object} params - 查询参数
 * @param {string} params.env - 环境
 * @param {string} params.service_type - 服务类型
 */
export const getResourceTemplates = (params = {}) => {
  return http.get(`${RESOURCE_BASE}/templates`, { params })
}

/**
 * 获取默认资源模板
 * @param {string} env - 环境
 * @param {string} serviceType - 服务类型
 */
export const getDefaultResourceTemplate = (env, serviceType) => {
  return http.get(`${RESOURCE_BASE}/template/default`, { params: { env, service_type: serviceType } })
}

/**
 * 校验资源配置
 * @param {Object} data - 校验数据
 * @param {string} data.env - 环境
 * @param {string} data.service_type - 服务类型
 * @param {Object} data.config - 资源配置
 */
export const validateResourceConfig = (data) => {
  return http.post(`${RESOURCE_BASE}/validate`, data)
}

/**
 * 获取环境资源规则
 * @param {string} env - 环境
 */
export const getResourceRules = (env) => {
  return http.get(`${RESOURCE_BASE}/rules`, { params: { env } })
}

/**
 * 获取审批列表
 * @param {Object} params - 查询参数
 */
export const getResourceApprovals = (params = {}) => {
  return http.get(`${RESOURCE_BASE}/approvals`, { params })
}

/**
 * 通过审批
 * @param {number} id - 审批ID
 * @param {string} comment - 审批意见
 */
export const approveResourceConfig = (id, comment = '') => {
  return http.put(`${RESOURCE_BASE}/approval/${id}/approve`, { comment })
}

/**
 * 拒绝审批
 * @param {number} id - 审批ID
 * @param {string} comment - 拒绝原因
 */
export const rejectResourceConfig = (id, comment = '') => {
  return http.put(`${RESOURCE_BASE}/approval/${id}/reject`, { comment })
}

// =======================
// CI/CD 制品库管理
// 对应后端路由: /api/v1/k8s/cicd/artifact/*
// =======================

const ARTIFACT_BASE = `${API_BASE}/k8s/cicd/artifact`

/**
 * 获取制品列表
 * @param {Object} params
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {number} params.pipeline_id - 流水线ID
 * @param {string} params.artifact_type - 制品类型(jar/binary/dist/wheel/image)
 * @param {string} params.language_type - 语言类型(go/java/frontend/python)
 * @param {string} params.status - 状态(ready/expired)
 */
export const getArtifacts = (params = {}) => {
  return http.get(`${ARTIFACT_BASE}/list`, { params })
}

/**
 * 获取制品详情
 * @param {number} id - 制品ID
 */
export const getArtifactDetail = (id) => {
  return http.get(`${ARTIFACT_BASE}/detail`, { params: { id } })
}

/**
 * 获取制品统计
 * @param {number} pipelineId - 流水线ID（可选）
 */
export const getArtifactStats = (pipelineId = null) => {
  const params = {}
  if (pipelineId) params.pipeline_id = pipelineId
  return http.get(`${ARTIFACT_BASE}/stats`, { params })
}

/**
 * 创建制品记录（镜像类型/无需上传文件）
 * @param {Object} data - 制品信息
 * @param {string} data.name - 制品名称（必填）
 * @param {number} data.pipeline_id - 流水线ID
 * @param {number} data.run_id - 运行记录ID
 * @param {string} data.artifact_type - 制品类型
 * @param {string} data.version - 版本号
 * @param {string} data.image_repo - 镜像仓库
 * @param {string} data.image_tag - 镜像标签
 */
export const createArtifactRecord = (data) => {
  return http.post(`${ARTIFACT_BASE}/create`, data)
}

/**
 * 更新制品信息
 * @param {Object} data - 更新参数
 * @param {number} data.id - 制品ID（必填）
 * @param {string} data.name - 制品名称
 * @param {string} data.version - 版本号
 * @param {string} data.artifact_type - 制品类型
 * @param {string} data.status - 状态
 * @param {string} data.image_repo - 镜像仓库
 * @param {string} data.image_tag - 镜像标签
 * @param {string} data.image_digest - 镜像摘要
 */
export const updateArtifact = (data) => {
  return http.post(`${ARTIFACT_BASE}/update`, data)
}

/**
 * 删除制品
 * @param {number} id - 制品ID
 */
export const deleteArtifact = (id) => {
  return http.post(`${ARTIFACT_BASE}/delete`, { id })
}

/**
 * 批量删除制品
 * @param {number[]} ids - 制品ID数组（最多100条）
 */
export const batchDeleteArtifacts = (ids) => {
  return http.post(`${ARTIFACT_BASE}/batch-delete`, { ids })
}

/**
 * 获取某次运行产出的制品列表
 * @param {number} runId - 运行记录ID
 */
export const getArtifactsByRunID = (runId) => {
  return http.get(`${ARTIFACT_BASE}/by-run`, { params: { run_id: runId } })
}

/**
 * 下载制品文件（通过 window.open 流式下载）
 * @param {number} id - 制品ID
 * @param {string} token - 认证Token
 */
export const getArtifactDownloadUrl = (id, token) => {
  return `${ARTIFACT_BASE}/download?id=${id}&token=${encodeURIComponent(token)}`
}

/**
 * 下载制品文件（通过 blob 下载，备用方案）
 * @param {number} id - 制品ID
 */
export const downloadArtifact = (id) => {
  return http.get(`${ARTIFACT_BASE}/download`, {
    params: { id },
    responseType: 'blob'
  })
}

/**
 * 为已有制品补传/替换文件
 * @param {number} id - 制品ID
 * @param {File} file - 文件对象
 */
/**
 * 为已有制品补传/替换文件
 * @param {number} id - 制品ID
 * @param {File} file - 文件对象
 * @param {Function} onProgress - 上传进度回调 (percent: 0-100)
 */
export const attachArtifactFile = (id, file, onProgress) => {
  const formData = new FormData()
  formData.append('id', id)
  formData.append('file', file)
  return http.post(`${ARTIFACT_BASE}/attach`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 10 * 60 * 1000, // 10分钟，大文件上传需要更长超时
    onUploadProgress: (e) => {
      if (onProgress && e.total) onProgress(Math.round((e.loaded / e.total) * 100))
    }
  })
}

/**
 * 上传制品（带文件 + 元数据，新建记录）
 * @param {File} file - 文件对象
 * @param {Object} meta - 元数据
 * @param {Function} onProgress - 上传进度回调 (percent: 0-100)
 */
export const uploadArtifact = (file, meta = {}, onProgress) => {
  const formData = new FormData()
  formData.append('file', file)
  Object.entries(meta).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') formData.append(k, v)
  })
  return http.post(`${ARTIFACT_BASE}/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 10 * 60 * 1000, // 10分钟，大文件上传需要更长超时
    onUploadProgress: (e) => {
      if (onProgress && e.total) onProgress(Math.round((e.loaded / e.total) * 100))
    }
  })
}

// =======================
// SonarQube 代码质量管理
// =======================

const PIPELINE_BASE = `${API_BASE}/k8s/cicd/pipeline`

/**
 * 获取 SonarQube 代码质量报告
 * @param {number} pipelineId - 流水线ID
 * @param {number} runId - 运行记录ID（可选，空则获取最新一次）
 */
export const getSonarReport = (pipelineId, runId = null) => {
  const params = { pipeline_id: pipelineId }
  if (runId) params.run_id = runId
  return http.get(`${PIPELINE_BASE}/sonar-report`, { params })
}

/**
 * 模板化发布验证
 */
export const getTemplateVerify = () => {
  return http.get(`${PIPELINE_BASE}/template-verify`)
}

/**
 * 模拟模板化发布流程
 * @param {string} languageType - 语言类型
 * @param {string} gitRepo - Git 仓库地址
 * @param {string} imageRepo - 镜像仓库地址
 * @param {string} gitBranch - Git 分支
 */
export const getTemplateSimulate = (languageType, gitRepo, imageRepo, gitBranch = 'main') => {
  return http.get(`${PIPELINE_BASE}/template-simulate`, {
    params: { language_type: languageType, git_repo: gitRepo, image_repo: imageRepo, git_branch: gitBranch }
  })
}

// =======================
// 构建探针管理（Build Agent）
// 对应后端路由: /api/v1/k8s/cicd/agent/*
// =======================

const AGENT_BASE = `${API_BASE}/k8s/cicd/agent`

/**
 * 获取探针列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.category - 分类：observability/diagnostics/security/custom
 * @param {string} params.scope - 适用语言：java/go/python/all
 * @param {string} params.status - 状态：active/inactive
 * @param {string} params.keyword - 搜索关键字
 */
export const getBuildAgents = (params = {}) => {
  return http.get(`${AGENT_BASE}/list`, { params })
}

/**
 * 获取探针详情
 * @param {number} id - 探针ID
 */
export const getBuildAgentDetail = (id) => {
  return http.get(`${AGENT_BASE}/detail`, { params: { id } })
}

/**
 * 上传探针文件
 * @param {FormData} formData - 包含 file, name, display_name, category, scope, version 等
 * @param {Function} onUploadProgress - 上传进度回调
 */
export const uploadBuildAgent = (formData, onUploadProgress) => {
  return http.post(`${AGENT_BASE}/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 300000, // 5 分钟，大文件上传需要更长超时
    onUploadProgress
  })
}

/**
 * 更新探针信息
 * @param {Object} data - 更新数据
 */
export const updateBuildAgent = (data) => {
  return http.post(`${AGENT_BASE}/update`, data)
}

/**
 * 切换探针启用/停用状态
 * @param {number} id - 探针ID
 */
export const toggleBuildAgent = (id) => {
  return http.post(`${AGENT_BASE}/toggle`, { id })
}

/**
 * 删除探针
 * @param {number} id - 探针ID
 */
export const deleteBuildAgent = (id) => {
  return http.post(`${AGENT_BASE}/delete`, { id })
}

/**
 * 下载探针文件
 * @param {number} id - 探针ID
 */
export const downloadBuildAgent = (id) => {
  return `${AGENT_BASE}/download?id=${id}`
}

/**
 * 按名称下载探针（流水线使用）
 * @param {string} name - 探针名称
 */
export const downloadBuildAgentByName = (name) => {
  return `${AGENT_BASE}/download?name=${name}`
}

/**
 * 获取指定语言的已启用探针列表
 * @param {string} scope - 语言类型：java/go/python
 */
export const getBuildAgentsByScope = (scope) => {
  return http.get(`${AGENT_BASE}/by-scope`, { params: { scope } })
}

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
// CI/CD 流水线管理（静态Mock数据）
// =======================

// Mock 流水线数据
const mockPipelines = [
  {
    id: 1,
    name: 'frontend-deploy',
    description: '前端项目部署流水线',
    status: 'success',
    lastRunTime: '2024-01-15 10:30:00',
    lastRunStatus: 'success',
    gitRepo: 'https://github.com/example/frontend.git',
    branch: 'main',
    stages: [
      { name: '代码拉取', description: '从git仓库拉取代码', status: 'success', startTime: '2024-01-15 10:30:00', endTime: '2024-01-15 10:30:30' },
      { name: '构建', description: '编译构建项目', status: 'success', startTime: '2024-01-15 10:30:30', endTime: '2024-01-15 10:32:00' },
      { name: '测试', description: '运行单元测试', status: 'success', startTime: '2024-01-15 10:32:00', endTime: '2024-01-15 10:33:00' },
      { name: '部署', description: '部署到K8s集群', status: 'success', startTime: '2024-01-15 10:33:00', endTime: '2024-01-15 10:35:00' }
    ],
    envVars: [{ name: 'NODE_ENV', value: 'production' }],
    deploymentConfig: { namespace: 'default', deploymentName: 'frontend', image: 'frontend:latest', replicas: 3, strategy: 'rollingUpdate' }
  },
  {
    id: 2,
    name: 'backend-api',
    description: '后端API服务部署',
    status: 'running',
    lastRunTime: '2024-01-15 11:00:00',
    lastRunStatus: 'running',
    gitRepo: 'https://github.com/example/backend.git',
    branch: 'develop',
    stages: [
      { name: '代码拉取', description: '从git仓库拉取代码', status: 'success', startTime: '2024-01-15 11:00:00', endTime: '2024-01-15 11:00:20' },
      { name: '构建', description: '编译Go项目', status: 'running', startTime: '2024-01-15 11:00:20', endTime: null },
      { name: '测试', description: '运行单元测试', status: 'pending', startTime: null, endTime: null },
      { name: '部署', description: '部署到K8s集群', status: 'pending', startTime: null, endTime: null }
    ],
    envVars: [{ name: 'GO_ENV', value: 'production' }],
    deploymentConfig: { namespace: 'default', deploymentName: 'backend-api', image: 'backend:latest', replicas: 2, strategy: 'rollingUpdate' }
  }
]

const mockTemplates = [
  { id: 1, name: 'Node.js标准流水线', description: '适用于Node.js项目的CI/CD模板' },
  { id: 2, name: 'Go微服务流水线', description: '适用于Go微服务的CI/CD模板' },
  { id: 3, name: 'Java Spring Boot', description: '适用于Spring Boot项目的CI/CD模板' }
]

// 创建流水线
export const createPipeline = (data) => {
  return Promise.resolve({ code: 0, msg: 'success', data: { id: Date.now(), ...data } })
}

// 更新流水线
export const updatePipeline = (id, data) => {
  return Promise.resolve({ code: 0, msg: 'success', data: { id, ...data } })
}

// 获取流水线详情
export const getPipelineDetail = (id) => {
  const pipeline = mockPipelines.find(p => p.id === parseInt(id)) || mockPipelines[0]
  return Promise.resolve({ code: 0, msg: 'success', data: pipeline })
}

// 获取流水线模板列表
export const getPipelineTemplates = () => {
  return Promise.resolve({ code: 0, msg: 'success', data: mockTemplates })
}

// 运行流水线
export const runPipeline = (id) => {
  return Promise.resolve({ code: 0, msg: '流水线已启动' })
}

// 停止流水线
export const stopPipeline = (id) => {
  return Promise.resolve({ code: 0, msg: '流水线已停止' })
}

// 获取流水线日志
export const getPipelineLogs = (id, timestamp) => {
  return Promise.resolve({
    code: 0,
    msg: 'success',
    data: {
      logs: [
        '[2024-01-15 10:30:00] 开始执行流水线...',
        '[2024-01-15 10:30:05] 拉取代码中...',
        '[2024-01-15 10:30:20] 代码拉取完成',
        '[2024-01-15 10:30:25] 开始构建...',
        '[2024-01-15 10:31:00] 构建完成',
        '[2024-01-15 10:31:05] 运行测试...',
        '[2024-01-15 10:32:00] 测试通过'
      ]
    }
  })
}

// 部署到K8s
export const deployToK8s = (data) => {
  return Promise.resolve({ code: 0, msg: '部署成功', data: { deploymentId: Date.now() } })
}

// 获取部署历史
export const getDeploymentHistory = (pipelineId) => {
  return Promise.resolve({
    code: 0,
    msg: 'success',
    data: [
      { id: 1, version: 'v1.0.0', status: 'success', deployTime: '2024-01-15 10:35:00', operator: 'admin' },
      { id: 2, version: 'v0.9.0', status: 'success', deployTime: '2024-01-14 15:20:00', operator: 'admin' }
    ]
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

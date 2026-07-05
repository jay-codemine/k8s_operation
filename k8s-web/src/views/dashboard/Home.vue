<template>
  <div class="dashboard-container">
    <!-- 顶部欢迎区域 -->
    <div class="dashboard-header">
      <div class="header-content">
        <h1>👋 欢迎回来，{{ username }}</h1>
        <p>当前时间：{{ currentTime }}</p>
      </div>
      <div class="header-actions">
        <!-- 集群连接状态指示器 -->
        <div class="connection-status" :class="clusterConnectionStatus">
          <span v-if="clusterConnectionStatus === 'checking'" class="status-indicator">
            <span class="spinner"></span> 连接中...
          </span>
          <span v-else-if="clusterConnectionStatus === 'connected'" class="status-indicator connected">
            ✅ 已连接
          </span>
          <span v-else-if="clusterConnectionStatus === 'timeout'" class="status-indicator timeout">
            ⚠️ 部分集群超时
          </span>
          <span v-else class="status-indicator error">
            ❌ 连接失败
          </span>
        </div>
        <button class="btn btn-primary" @click="refreshData">
          <span v-if="!loading">🔄 刷新数据</span>
          <span v-else>加载中...</span>
        </button>
      </div>
    </div>

    <!-- 集群概览卡片 - 仅有集群权限时显示 -->
    <div v-if="hasClusterAccess" class="cluster-overview">
      <div class="section-title">
        <h2>🖥️ 集群概览</h2>
      </div>
      <div class="stats-grid">
        <div v-if="canView('clusters')" class="stat-card cluster-card">
          <div class="stat-icon">🏛️</div>
          <div class="stat-content">
            <div class="stat-label">集群数量</div>
            <div class="stat-value">{{ clusterStats.total }}</div>
            <div class="stat-detail">
              <span class="status-active">活跃: {{ clusterStats.active }}</span>
              <span class="status-inactive">离线: {{ clusterStats.inactive }}</span>
            </div>
          </div>
        </div>

        <div v-if="canView('nodes')" class="stat-card node-card">
          <div class="stat-icon">💻</div>
          <div class="stat-content">
            <div class="stat-label">节点总数</div>
            <div class="stat-value">{{ nodeStats.total }}</div>
            <div class="stat-detail">
              <span class="status-ready">就绪: {{ nodeStats.ready }}</span>
              <span class="status-notready">异常: {{ nodeStats.notReady }}</span>
            </div>
          </div>
        </div>

        <div v-if="canView('pods')" class="stat-card pod-card">
          <div class="stat-icon">📦</div>
          <div class="stat-content">
            <div class="stat-label">Pod 总数</div>
            <div class="stat-value">{{ podStats.total }}</div>
            <div class="stat-detail">
              <span class="status-running">运行: {{ podStats.running }}</span>
              <span class="status-pending">等待: {{ podStats.pending }}</span>
              <span class="status-failed">失败: {{ podStats.failed }}</span>
            </div>
          </div>
        </div>

        <div v-if="canView('namespaces')" class="stat-card namespace-card">
          <div class="stat-icon">📁</div>
          <div class="stat-content">
            <div class="stat-label">命名空间</div>
            <div class="stat-value">{{ namespaceStats.total }}</div>
            <div class="stat-detail">
              <span class="status-system">系统: {{ namespaceStats.system }}</span>
              <span class="status-user">用户: {{ namespaceStats.user }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 工作负载统计 - 仅有集群权限时显示 -->
    <div v-if="hasClusterAccess && hasAnyWorkloadPermission" class="workload-section">
      <div class="section-title">
        <h2>🚀 工作负载统计</h2>
      </div>
      <div class="workload-grid">
        <div 
          v-if="canView('deployments')"
          class="workload-card" 
          @click="handleClusterNavigate('workloads/deployments')"
        >
          <div class="workload-icon deployment">🚀</div>
          <div class="workload-info">
            <div class="workload-name">Deployments</div>
            <div class="workload-count">{{ workloadStats.deployments }}</div>
          </div>
        </div>

        <div 
          v-if="canView('statefulsets')"
          class="workload-card" 
          @click="handleClusterNavigate('workloads/statefulsets')"
        >
          <div class="workload-icon statefulset">📊</div>
          <div class="workload-info">
            <div class="workload-name">StatefulSets</div>
            <div class="workload-count">{{ workloadStats.statefulsets }}</div>
          </div>
        </div>

        <div 
          v-if="canView('daemonsets')"
          class="workload-card" 
          @click="handleClusterNavigate('workloads/daemonsets')"
        >
          <div class="workload-icon daemonset">🔄</div>
          <div class="workload-info">
            <div class="workload-name">DaemonSets</div>
            <div class="workload-count">{{ workloadStats.daemonsets }}</div>
          </div>
        </div>

        <div 
          v-if="canView('jobs')"
          class="workload-card" 
          @click="handleClusterNavigate('workloads/jobs')"
        >
          <div class="workload-icon job">⚙️</div>
          <div class="workload-info">
            <div class="workload-name">Jobs</div>
            <div class="workload-count">{{ workloadStats.jobs }}</div>
          </div>
        </div>

        <div 
          v-if="canView('cronjobs')"
          class="workload-card" 
          @click="handleClusterNavigate('workloads/cronjobs')"
        >
          <div class="workload-icon cronjob">⏰</div>
          <div class="workload-info">
            <div class="workload-name">CronJobs</div>
            <div class="workload-count">{{ workloadStats.cronjobs }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 网络与存储 - 仅有集群权限时显示 -->
    <div v-if="hasClusterAccess && hasAnyNetworkOrStoragePermission" class="resource-section">
      <div class="section-row">
        <div class="resource-group">
          <div class="section-title">
            <h2>🌐 网络资源</h2>
          </div>
          <div class="resource-list">
            <div 
              v-if="canView('services')"
              class="resource-item" 
              @click="handleClusterNavigate('networking/services')"
            >
              <div class="resource-icon">🔌</div>
              <div class="resource-info">
                <div class="resource-name">Services</div>
                <div class="resource-count">{{ networkStats.services }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div 
              v-if="canView('ingress')"
              class="resource-item" 
              @click="handleClusterNavigate('networking/ingresses')"
            >
              <div class="resource-icon">🌍</div>
              <div class="resource-info">
                <div class="resource-name">Ingresses</div>
                <div class="resource-count">{{ networkStats.ingresses }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>
          </div>
        </div>

        <div class="resource-group">
          <div class="section-title">
            <h2>💾 存储资源</h2>
          </div>
          <div class="resource-list">
            <div 
              v-if="canView('pv')"
              class="resource-item" 
              @click="handleClusterNavigate('storage/persistentvolumes')"
            >
              <div class="resource-icon">💿</div>
              <div class="resource-info">
                <div class="resource-name">PersistentVolumes</div>
                <div class="resource-count">{{ storageStats.pvs }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div 
              v-if="canView('pvc')"
              class="resource-item" 
              @click="handleClusterNavigate('storage/persistentvolumeclaims')"
            >
              <div class="resource-icon">📝</div>
              <div class="resource-info">
                <div class="resource-name">PVCs</div>
                <div class="resource-count">{{ storageStats.pvcs }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div 
              v-if="canView('storageclasses')"
              class="resource-item" 
              @click="handleClusterNavigate('storage/storageclasses')"
            >
              <div class="resource-icon">📦</div>
              <div class="resource-info">
                <div class="resource-name">StorageClasses</div>
                <div class="resource-count">{{ storageStats.storageClasses }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 配置资源 - 仅有集群权限时显示 -->
    <div v-if="hasClusterAccess && hasAnyConfigPermission" class="config-section">
      <div class="section-title">
        <h2>⚙️ 配置资源</h2>
      </div>
      <div class="config-grid">
        <div 
          v-if="canView('configmaps')"
          class="config-card" 
          @click="handleClusterNavigate('config/configmaps')"
        >
          <div class="config-icon">🗂️</div>
          <div class="config-info">
            <div class="config-name">ConfigMaps</div>
            <div class="config-count">{{ configStats.configmaps }}</div>
          </div>
        </div>

        <div 
          v-if="canView('secrets')"
          class="config-card" 
          @click="handleClusterNavigate('config/secrets')"
        >
          <div class="config-icon">🔐</div>
          <div class="config-info">
            <div class="config-name">Secrets</div>
            <div class="config-count">{{ configStats.secrets }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速链接 -->
    <div class="quick-links-section">
      <div class="section-title">
        <h2>🔗 快速链接</h2>
      </div>
      <div class="quick-links-grid">
        <a 
          v-if="canView('clusters')"
          class="quick-link" 
          @click="handleNavigate('/clusters')"
        >
          <div class="link-icon">🏛️</div>
          <div class="link-text">集群管理</div>
        </a>
        <a 
          v-if="canView('nodes')"
          class="quick-link" 
          @click="handleClusterNavigate('nodes')"
        >
          <div class="link-icon">💻</div>
          <div class="link-text">节点管理</div>
        </a>
        <a 
          v-if="canView('namespaces')"
          class="quick-link" 
          @click="handleClusterNavigate('namespaces')"
        >
          <div class="link-icon">📁</div>
          <div class="link-text">命名空间</div>
        </a>
        <a 
          v-if="canView('pipelines')"
          class="quick-link" 
          @click="handleNavigate('/cicd/pipelines')"
        >
          <div class="link-icon">🚀</div>
          <div class="link-text">CI/CD 流水线</div>
        </a>
        <a 
          v-if="canView('repositories')"
          class="quick-link" 
          @click="handleNavigate('/images/repositories')"
        >
          <div class="link-icon">📷</div>
          <div class="link-text">镜像仓库</div>
        </a>
        <a 
          v-if="canView('users')"
          class="quick-link" 
          @click="handleNavigate('/users')"
        >
          <div class="link-icon">👥</div>
          <div class="link-text">用户管理</div>
        </a>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { useClusterStore } from '@/stores/cluster'
import permissionStore from '@/stores/permission'

// 导入 API
import { getClusterList } from '@/api/cluster'
import nodesApi from '@/api/cluster/nodes'
import podsApi from '@/api/cluster/workloads/pods'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'
import jobsApi from '@/api/cluster/workloads/jobs'
import cronjobsApi from '@/api/cluster/workloads/cronjobs'
import serviceApi from '@/api/cluster/networking/service'
import ingressApi from '@/api/cluster/networking/ingress'
import pvApi from '@/api/cluster/storage/pv'
import pvcApi from '@/api/cluster/storage/pvc'
import storageclassApi from '@/api/cluster/storage/storageclasses'
import configmapApi from '@/api/cluster/config/configmap'
import secretApi from '@/api/cluster/config/secret'
import namespaceApi from '@/api/cluster/namespaces'

const router = useRouter()
const clusterStore = useClusterStore()

// =============================================================================
// 超时控制工具（大厂设计模式）
// 
// 设计原则：
// 1. 快速失败：单个 API 超时不阻塞其他请求
// 2. 优雅降级：超时后显示默认数据，不影响用户体验
// 3. 可观测性：记录超时日志，便于问题定位
// =============================================================================

// API 超时配置（单位：毫秒）
const API_TIMEOUT = 8000 // 8 秒，给后端 5 秒超时 + 网络延迟

/**
 * 包装 Promise，添加超时控制
 * @param {Promise} promise 原始 Promise
 * @param {number} timeoutMs 超时时间（毫秒）
 * @param {string} name 请求名称（用于日志）
 * @returns {Promise} 带超时控制的 Promise
 */
const withTimeout = (promise, timeoutMs = API_TIMEOUT, name = 'API') => {
  return Promise.race([
    promise,
    new Promise((_, reject) => 
      setTimeout(() => reject(new Error(`${name} 请求超时(${timeoutMs}ms)`)), timeoutMs)
    )
  ])
}

// 集群连接状态
const clusterConnectionStatus = ref('checking') // checking | connected | timeout | error

/**
 * 基于实际授权的资源权限配置
 * 参考 Rancher/Kuboard/KubeSphere 权限模型
 * 
 * 角色权限层级：
 * - cluster_admin: 集群完整权限，可查看所有资源
 * - developer: 开发者权限，可查看常用工作负载、服务、配置
 * - viewer: 只读权限，仅查看基础资源
 * - cicd_admin: CI/CD管理员，可查看部署相关资源
 */
const ROLE_RESOURCE_MAP = {
  // 集群管理员 - 所有资源
  cluster_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ],
  // 开发者 - 常用资源（劅除节点、存储类等集群级资源）
  developer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets',
    'namespaces', 'clusters', 'pipelines'
  ],
  // 只读 - 可查看所有资源（与 cluster_admin 相同，但无操作权限）
  viewer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ],
  // CI/CD管理员 - 部署相关
  cicd_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets',
    'namespaces', 'clusters', 'pipelines'
  ]
}

// 平台级管理员可查看的额外资源
const PLATFORM_ADMIN_RESOURCES = ['repositories', 'users']

/**
 * 判断是否有任何集群访问权限
 */
const hasClusterAccess = computed(() => {
  if (permissionStore.state.isSuperAdmin) return true
  // 检查是否有任何集群权限
  const clusterPerms = permissionStore.state.clusterPermissions
  return Object.keys(clusterPerms).length > 0
})

/**
 * 获取用户的最高角色类型
 */
const userHighestRole = computed(() => {
  if (permissionStore.state.isSuperAdmin) return 'super_admin'
  
  const roleTypes = permissionStore.roleTypes.value
  
  // 角色优先级：super_admin > platform_admin > cluster_admin > cicd_admin > developer > viewer
  if (roleTypes.includes('super_admin')) return 'super_admin'
  if (roleTypes.includes('platform_admin')) return 'platform_admin'
  if (roleTypes.includes('cluster_admin')) return 'cluster_admin'
  if (roleTypes.includes('cicd_admin')) return 'cicd_admin'
  if (roleTypes.includes('developer')) return 'developer'
  if (roleTypes.includes('viewer')) return 'viewer'
  
  return null
})

/**
 * 获取用户可查看的资源列表
 */
const userAccessibleResources = computed(() => {
  if (permissionStore.state.isSuperAdmin) {
    // 超级管理员可查看所有资源
    return [
      ...ROLE_RESOURCE_MAP.cluster_admin,
      ...PLATFORM_ADMIN_RESOURCES,
      'pipelines'
    ]
  }
  
  const role = userHighestRole.value
  if (!role) return []
  
  // platform_admin 等同于 cluster_admin + 平台资源
  if (role === 'platform_admin') {
    return [
      ...ROLE_RESOURCE_MAP.cluster_admin,
      ...PLATFORM_ADMIN_RESOURCES,
      'pipelines'
    ]
  }
  
  return ROLE_RESOURCE_MAP[role] || []
})

/**
 * 检查是否有权限查看资源（基于实际授权）
 */
const canView = (resource) => {
  // 超级管理员全部可见
  if (permissionStore.state.isSuperAdmin) return true
  
  // 集群相关资源需要检查是否有集群权限
  const clusterResources = [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ]
  
  if (clusterResources.includes(resource)) {
    // 没有任何集群权限则不显示
    if (!hasClusterAccess.value) return false
  }
  
  // 检查资源是否在用户可访问列表中
  return userAccessibleResources.value.includes(resource)
}

/**
 * 检查是否有任何工作负载权限
 */
const hasAnyWorkloadPermission = computed(() => {
  const workloads = ['deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs']
  return workloads.some(r => canView(r))
})

/**
 * 检查是否有任何网络或存储权限
 */
const hasAnyNetworkOrStoragePermission = computed(() => {
  const resources = ['services', 'ingress', 'pv', 'pvc', 'storageclasses']
  return resources.some(r => canView(r))
})

/**
 * 检查是否有任何配置资源权限
 */
const hasAnyConfigPermission = computed(() => {
  const configs = ['configmaps', 'secrets']
  return configs.some(r => canView(r))
})

/**
 * 导航处理
 */
const handleNavigate = (path) => {
  router.push(path)
}

// 用户名
const username = computed(() => {
  const userStr = localStorage.getItem('user') || sessionStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user.username || 'Admin'
    } catch (e) {
      return 'Admin'
    }
  }
  return 'Admin'
})

// 当前时间
const currentTime = ref('')
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

// 加载状态
const loading = ref(false)

// 统计数据
const clusterStats = ref({
  total: 0,
  active: 0,
  inactive: 0
})

const nodeStats = ref({
  total: 0,
  ready: 0,
  notReady: 0
})

const podStats = ref({
  total: 0,
  running: 0,
  pending: 0,
  failed: 0
})

const namespaceStats = ref({
  total: 0,
  system: 0,
  user: 0
})

const workloadStats = ref({
  deployments: 0,
  statefulsets: 0,
  daemonsets: 0,
  jobs: 0,
  cronjobs: 0
})

const networkStats = ref({
  services: 0,
  ingresses: 0
})

const storageStats = ref({
  pvs: 0,
  pvcs: 0,
  storageClasses: 0
})

const configStats = ref({
  configmaps: 0,
  secrets: 0
})

// 确保有默认集群
const ensureDefaultCluster = async () => {
  try {
    // 如果 store 中已有当前集群，直接使用
    if (clusterStore.current?.id) {
      return true
    }
    
    // 获取集群列表
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0 && res.data) {
      const allClusters = res.data.list || []
      
      // 权限过滤：只显示用户有权限访问的集群
      const clusters = allClusters.filter(c => 
        permissionStore.state.isSuperAdmin ||
        permissionStore.state.accessibleClusterIds.includes(c.id)
      )
      
      if (clusters.length === 0) {
        return false
      }
      
      // 优先使用默认集群，否则使用第一个
      const defaultCluster = clusters.find(c => c.is_default) || clusters[0]
      clusterStore.setCurrent(defaultCluster)
      return true
    }
    return false
  } catch (error) {
    console.error('获取默认集群失败:', error)
    return false
  }
}

// 加载数据（带超时控制）
const loadData = async () => {
  loading.value = true
  clusterConnectionStatus.value = 'checking'
  
  try {
    // 先确保有默认集群（快速超时）
    const hasCluster = await withTimeout(
      ensureDefaultCluster(), 
      5000, 
      '获取集群列表'
    ).catch(() => false)
    
    if (!hasCluster) {
      clusterConnectionStatus.value = 'timeout'
      useMockData()
      Message.warning({ content: '无可用集群或连接超时，暂无数据', duration: 3000 })
      return
    }
    
    clusterConnectionStatus.value = 'connected'
    
    // 并发请求所有数据（每个请求独立超时，互不影响）
    const results = await Promise.allSettled([
      withTimeout(loadClusterStats(), API_TIMEOUT, '集群统计'),
      withTimeout(loadNodeStats(), API_TIMEOUT, '节点统计'),
      withTimeout(loadPodStats(), API_TIMEOUT, 'Pod统计'),
      withTimeout(loadNamespaceStats(), API_TIMEOUT, '命名空间统计'),
      withTimeout(loadWorkloadStats(), API_TIMEOUT, '工作负载统计'),
      withTimeout(loadNetworkStats(), API_TIMEOUT, '网络统计'),
      withTimeout(loadStorageStats(), API_TIMEOUT, '存储统计'),
      withTimeout(loadConfigStats(), API_TIMEOUT, '配置统计')
    ])
    
    // 统计成功/失败数量
    const failed = results.filter(r => r.status === 'rejected')
    if (failed.length > 0) {
      console.warn(`Dashboard: ${failed.length}/${results.length} 个请求超时或失败`)
      // 部分超时时只显示警告，不影响已成功的数据
      if (failed.length >= results.length / 2) {
        clusterConnectionStatus.value = 'timeout'
      }
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    clusterConnectionStatus.value = 'error'
    useMockData()
  } finally {
    loading.value = false
  }
}

// 无数据时显示真实零值（不使用虚假mock数据）
const useMockData = () => {
  clusterStats.value = { total: 0, active: 0, inactive: 0 }
  nodeStats.value = { total: 0, ready: 0, notReady: 0 }
  podStats.value = { total: 0, running: 0, pending: 0, failed: 0 }
  namespaceStats.value = { total: 0, system: 0, user: 0 }
  workloadStats.value = { deployments: 0, statefulsets: 0, daemonsets: 0, jobs: 0, cronjobs: 0 }
  networkStats.value = { services: 0, ingresses: 0 }
  storageStats.value = { pvs: 0, pvcs: 0, storageClasses: 0 }
  configStats.value = { configmaps: 0, secrets: 0 }
}

// 加载集群统计
const loadClusterStats = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const allClusters = res.data.list || []
      // 权限过滤：只统计用户有权限访问的集群
      const clusters = allClusters.filter(c => 
        permissionStore.state.isSuperAdmin ||
        permissionStore.state.accessibleClusterIds.includes(c.id)
      )
      clusterStats.value.total = clusters.length
      // status: 0=正常, 1=异常, 2=待检测
      clusterStats.value.active = clusters.filter(c => Number(c.status) === 0).length
      clusterStats.value.inactive = clusters.filter(c => Number(c.status) === 1 || Number(c.status) === 2).length
    }
  } catch (error) {
    console.error('加载集群数据失败:', error)
    clusterStats.value = { total: 0, active: 0, inactive: 0 }
  }
}

// 加载节点统计
const loadNodeStats = async () => {
  try {
    const res = await nodesApi.list({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const nodes = res.data.list || res.data || []
      nodeStats.value.total = nodes.length
      nodeStats.value.ready = nodes.filter(n => 
        n.status === 'Ready' || n.status === 'ready'
      ).length
      nodeStats.value.notReady = nodeStats.value.total - nodeStats.value.ready
    }
  } catch (error) {
    console.error('加载节点数据失败:', error)
    nodeStats.value = { total: 0, ready: 0, notReady: 0 }
  }
}

// 加载 Pod 统计
const loadPodStats = async () => {
  try {
    const res = await podsApi.list({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const pods = res.data.list || res.data || []
      podStats.value.total = pods.length
      podStats.value.running = pods.filter(p => p.status === 'Running').length
      podStats.value.pending = pods.filter(p => p.status === 'Pending').length
      podStats.value.failed = pods.filter(p => 
        p.status === 'Failed' || p.status === 'Error' || p.status === 'CrashLoopBackOff'
      ).length
    }
  } catch (error) {
    console.error('加载 Pod 数据失败:', error)
    podStats.value = { total: 0, running: 0, pending: 0, failed: 0 }
  }
}

// 加载命名空间统计
const loadNamespaceStats = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      let namespaces = res.data.list || res.data || []
      
      // 权限过滤：只统计用户有权限访问的命名空间
      if (!permissionStore.state.isSuperAdmin) {
        const clusterId = clusterStore.current?.id
        if (clusterId) {
          const accessibleNs = permissionStore.getAccessibleNamespaces(clusterId)
          if (accessibleNs.length > 0 && !accessibleNs.includes('*') && !accessibleNs.includes('__none__')) {
            namespaces = namespaces.filter(ns => accessibleNs.includes(ns.name || ns.metadata?.name))
          } else if (accessibleNs.includes('__none__')) {
            namespaces = []
          }
        }
      }
      
      namespaceStats.value.total = namespaces.length
      const systemNs = ['kube-system', 'kube-public', 'kube-node-lease', 'default']
      namespaceStats.value.system = namespaces.filter(ns => 
        systemNs.includes(ns.name)
      ).length
      namespaceStats.value.user = namespaceStats.value.total - namespaceStats.value.system
    }
  } catch (error) {
    console.error('加载命名空间数据失败:', error)
    namespaceStats.value = { total: 0, system: 0, user: 0 }
  }
}

// 加载工作负载统计
const loadWorkloadStats = async () => {
  try {
    const [deployRes, stsRes, dsRes, jobRes, cronRes] = await Promise.allSettled([
      deploymentsApi.list({ page: 1, limit: 1000 }),
      statefulsetsApi.list({ page: 1, limit: 1000 }),
      daemonsetsApi.list({ page: 1, limit: 1000 }),
      jobsApi.list({ page: 1, limit: 1000 }),
      cronjobsApi.list({ page: 1, limit: 1000 })
    ])
    
    if (deployRes.status === 'fulfilled' && deployRes.value.code === 0) {
      workloadStats.value.deployments = (deployRes.value.data?.list || deployRes.value.data || []).length
    } else {
      workloadStats.value.deployments = 0
    }
    
    if (stsRes.status === 'fulfilled' && stsRes.value.code === 0) {
      workloadStats.value.statefulsets = (stsRes.value.data?.list || stsRes.value.data || []).length
    } else {
      workloadStats.value.statefulsets = 0
    }
    
    if (dsRes.status === 'fulfilled' && dsRes.value.code === 0) {
      workloadStats.value.daemonsets = (dsRes.value.data?.list || dsRes.value.data || []).length
    } else {
      workloadStats.value.daemonsets = 0
    }
    
    if (jobRes.status === 'fulfilled' && jobRes.value.code === 0) {
      workloadStats.value.jobs = (jobRes.value.data?.list || jobRes.value.data || []).length
    } else {
      workloadStats.value.jobs = 0
    }
    
    if (cronRes.status === 'fulfilled' && cronRes.value.code === 0) {
      workloadStats.value.cronjobs = (cronRes.value.data?.list || cronRes.value.data || []).length
    } else {
      workloadStats.value.cronjobs = 0
    }
  } catch (error) {
    console.error('加载工作负载数据失败:', error)
    workloadStats.value = { deployments: 0, statefulsets: 0, daemonsets: 0, jobs: 0, cronjobs: 0 }
  }
}

// 加载网络统计
const loadNetworkStats = async () => {
  try {
    const [svcRes, ingRes] = await Promise.allSettled([
      serviceApi.list({ page: 1, limit: 1000 }),
      ingressApi.list({ page: 1, limit: 1000 })
    ])
    
    if (svcRes.status === 'fulfilled' && svcRes.value.code === 0) {
      networkStats.value.services = (svcRes.value.data?.list || svcRes.value.data || []).length
    } else {
      networkStats.value.services = 0
    }
    
    if (ingRes.status === 'fulfilled' && ingRes.value.code === 0) {
      networkStats.value.ingresses = (ingRes.value.data?.list || ingRes.value.data || []).length
    } else {
      networkStats.value.ingresses = 0
    }
  } catch (error) {
    console.error('加载网络数据失败:', error)
    networkStats.value = { services: 0, ingresses: 0 }
  }
}

// 加载存储统计
const loadStorageStats = async () => {
  try {
    const [pvRes, pvcRes, scRes] = await Promise.allSettled([
      pvApi.list({ page: 1, limit: 1000 }),
      pvcApi.list({ page: 1, limit: 1000 }),
      storageclassApi.list({ page: 1, limit: 1000 })
    ])
    
    if (pvRes.status === 'fulfilled' && pvRes.value.code === 0) {
      storageStats.value.pvs = (pvRes.value.data?.list || pvRes.value.data || []).length
    } else {
      storageStats.value.pvs = 0
    }
    
    if (pvcRes.status === 'fulfilled' && pvcRes.value.code === 0) {
      storageStats.value.pvcs = (pvcRes.value.data?.list || pvcRes.value.data || []).length
    } else {
      storageStats.value.pvcs = 0
    }
    
    if (scRes.status === 'fulfilled' && scRes.value.code === 0) {
      storageStats.value.storageClasses = (scRes.value.data?.list || scRes.value.data || []).length
    } else {
      storageStats.value.storageClasses = 0
    }
  } catch (error) {
    console.error('加载存储数据失败:', error)
    storageStats.value = { pvs: 0, pvcs: 0, storageClasses: 0 }
  }
}

// 加载配置统计
const loadConfigStats = async () => {
  try {
    const [cmRes, secretRes] = await Promise.allSettled([
      configmapApi.list({ page: 1, limit: 1000 }),
      secretApi.list({ page: 1, limit: 1000 })
    ])
    
    if (cmRes.status === 'fulfilled' && cmRes.value.code === 0) {
      configStats.value.configmaps = (cmRes.value.data?.list || cmRes.value.data || []).length
    } else {
      configStats.value.configmaps = 0
    }
    
    if (secretRes.status === 'fulfilled' && secretRes.value.code === 0) {
      configStats.value.secrets = (secretRes.value.data?.list || secretRes.value.data || []).length
    } else {
      configStats.value.secrets = 0
    }
  } catch (error) {
    console.error('加载配置数据失败:', error)
    configStats.value = { configmaps: 0, secrets: 0 }
  }
}

// 刷新数据
const refreshData = async () => {
  Message.info({ content: '正在刷新数据...' })
  await loadData()
  Message.success({ content: '数据刷新成功' })
}

// 导航
const navigateTo = (path) => {
  router.push(path)
}

// 时间定时器
let timeInterval = null

onMounted(() => {
  updateTime()
  timeInterval = setInterval(updateTime, 1000)
  loadData()
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})
</script>

<style scoped>
.dashboard-container {
  padding: 24px;
  max-width: 1600px;
  margin: 0 auto;
  background: #f8fafc;
  min-height: calc(100vh - 60px);
}

/* 顶部欢迎区域 */
.dashboard-header {
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 24px;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.35);
}

.header-content h1 {
  font-size: 28px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.header-content p {
  font-size: 14px;
  margin: 0;
  opacity: 0.9;
}

.header-actions .btn {
  padding: 10px 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  backdrop-filter: blur(10px);
}

.header-actions .btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
}

.header-actions .btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 连接状态指示器（大厂设计风格） */
.connection-status {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  margin-right: 12px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.3s ease;
}

.connection-status .status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.9);
}

.connection-status .status-indicator.connected {
  color: #86efac;
}

.connection-status .status-indicator.timeout {
  color: #fcd34d;
}

.connection-status .status-indicator.error {
  color: #fca5a5;
}

/* 加载动画 */
.connection-status .spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 区域标题 */
.section-title {
  margin-bottom: 16px;
}

.section-title h2 {
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
  margin: 0;
}

/* 集群概览 */
.cluster-overview {
  margin-bottom: 32px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  gap: 16px;
  align-items: center;
  transition: all 0.3s ease;
  border-left: 4px solid;
}

.stat-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.stat-card.cluster-card { border-left-color: #6366f1; }
.stat-card.node-card { border-left-color: #10b981; }
.stat-card.pod-card { border-left-color: #f59e0b; }
.stat-card.namespace-card { border-left-color: #3b82f6; }

.stat-icon {
  font-size: 40px;
  line-height: 1;
}

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  color: #718096;
  margin-bottom: 4px;
  font-weight: 500;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #1a202c;
  margin-bottom: 8px;
}

.stat-detail {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
}

.stat-detail span {
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status-active { background: #c6f6d5; color: #22543d; }
.status-inactive { background: #fed7d7; color: #742a2a; }
.status-ready { background: #c6f6d5; color: #22543d; }
.status-notready { background: #fed7d7; color: #742a2a; }
.status-running { background: #bee3f8; color: #2c5282; }
.status-pending { background: #feebc8; color: #7c2d12; }
.status-failed { background: #fed7d7; color: #742a2a; }
.status-system { background: #e9d8fd; color: #44337a; }
.status-user { background: #bee3f8; color: #2c5282; }

/* 工作负载统计 */
.workload-section {
  margin-bottom: 32px;
}

.workload-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.workload-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 16px;
}

.workload-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-4px);
}

.workload-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.workload-icon.deployment { background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%); }
.workload-icon.statefulset { background: linear-gradient(135deg, #ec4899 0%, #f472b6 100%); }
.workload-icon.daemonset { background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%); }
.workload-icon.job { background: linear-gradient(135deg, #10b981 0%, #06b6d4 100%); }
.workload-icon.cronjob { background: linear-gradient(135deg, #f59e0b 0%, #fb923c 100%); }

.workload-info {
  flex: 1;
}

.workload-name {
  font-size: 14px;
  color: #718096;
  margin-bottom: 4px;
}

.workload-count {
  font-size: 24px;
  font-weight: 700;
  color: #1a202c;
}

/* 资源区域 */
.resource-section {
  margin-bottom: 32px;
}

.section-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 24px;
}

.resource-group {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.resource-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.resource-item:hover {
  background: #f7fafc;
}

.resource-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.resource-info {
  flex: 1;
}

.resource-name {
  font-size: 14px;
  font-weight: 500;
  color: #2d3748;
  margin-bottom: 2px;
}

.resource-count {
  font-size: 18px;
  font-weight: 700;
  color: #6366f1;
}

.resource-arrow {
  font-size: 20px;
  color: #cbd5e0;
}

/* 配置资源 */
.config-section {
  margin-bottom: 32px;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.config-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 16px;
}

.config-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-4px);
}

.config-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: linear-gradient(135deg, #ec4899 0%, #f472b6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.config-info {
  flex: 1;
}

.config-name {
  font-size: 14px;
  color: #718096;
  margin-bottom: 4px;
}

.config-count {
  font-size: 24px;
  font-weight: 700;
  color: #1a202c;
}

/* 快速链接 */
.quick-links-section {
  margin-bottom: 32px;
}

.quick-links-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.quick-link {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  text-decoration: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.quick-link:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-4px);
}

.link-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}

.link-text {
  font-size: 14px;
  font-weight: 500;
  color: #2d3748;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .section-row {
    grid-template-columns: 1fr;
  }
  
  .workload-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  }
}
</style>

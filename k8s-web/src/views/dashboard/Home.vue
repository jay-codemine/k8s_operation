<template>
  <div class="dashboard-container">
    <!-- 顶部欢迎区域 -->
    <div class="dashboard-header">
      <div class="header-content">
        <h1>👋 欢迎回来，{{ username }}</h1>
        <p>当前时间：{{ currentTime }}</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="refreshData">
          <span v-if="!loading">🔄 刷新数据</span>
          <span v-else>加载中...</span>
        </button>
      </div>
    </div>

    <!-- 集群概览卡片 -->
    <div class="cluster-overview">
      <div class="section-title">
        <h2>🖥️ 集群概览</h2>
      </div>
      <div class="stats-grid">
        <div class="stat-card cluster-card">
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

        <div class="stat-card node-card">
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

        <div class="stat-card pod-card">
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

        <div class="stat-card namespace-card">
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

    <!-- 工作负载统计 -->
    <div class="workload-section">
      <div class="section-title">
        <h2>🚀 工作负载统计</h2>
      </div>
      <div class="workload-grid">
        <div class="workload-card" @click="navigateTo('/c/default/workloads/deployments')">
          <div class="workload-icon deployment">🚀</div>
          <div class="workload-info">
            <div class="workload-name">Deployments</div>
            <div class="workload-count">{{ workloadStats.deployments }}</div>
          </div>
        </div>

        <div class="workload-card" @click="navigateTo('/c/default/workloads/statefulsets')">
          <div class="workload-icon statefulset">📊</div>
          <div class="workload-info">
            <div class="workload-name">StatefulSets</div>
            <div class="workload-count">{{ workloadStats.statefulsets }}</div>
          </div>
        </div>

        <div class="workload-card" @click="navigateTo('/c/default/workloads/daemonsets')">
          <div class="workload-icon daemonset">🔄</div>
          <div class="workload-info">
            <div class="workload-name">DaemonSets</div>
            <div class="workload-count">{{ workloadStats.daemonsets }}</div>
          </div>
        </div>

        <div class="workload-card" @click="navigateTo('/c/default/workloads/jobs')">
          <div class="workload-icon job">⚙️</div>
          <div class="workload-info">
            <div class="workload-name">Jobs</div>
            <div class="workload-count">{{ workloadStats.jobs }}</div>
          </div>
        </div>

        <div class="workload-card" @click="navigateTo('/c/default/workloads/cronjobs')">
          <div class="workload-icon cronjob">⏰</div>
          <div class="workload-info">
            <div class="workload-name">CronJobs</div>
            <div class="workload-count">{{ workloadStats.cronjobs }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 网络与存储 -->
    <div class="resource-section">
      <div class="section-row">
        <div class="resource-group">
          <div class="section-title">
            <h2>🌐 网络资源</h2>
          </div>
          <div class="resource-list">
            <div class="resource-item" @click="navigateTo('/c/default/networking/services')">
              <div class="resource-icon">🔌</div>
              <div class="resource-info">
                <div class="resource-name">Services</div>
                <div class="resource-count">{{ networkStats.services }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div class="resource-item" @click="navigateTo('/c/default/networking/ingresses')">
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
            <div class="resource-item" @click="navigateTo('/c/default/storage/persistentvolumes')">
              <div class="resource-icon">💿</div>
              <div class="resource-info">
                <div class="resource-name">PersistentVolumes</div>
                <div class="resource-count">{{ storageStats.pvs }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div class="resource-item" @click="navigateTo('/c/default/storage/persistentvolumeclaims')">
              <div class="resource-icon">📝</div>
              <div class="resource-info">
                <div class="resource-name">PVCs</div>
                <div class="resource-count">{{ storageStats.pvcs }}</div>
              </div>
              <div class="resource-arrow">›</div>
            </div>

            <div class="resource-item" @click="navigateTo('/c/default/storage/storageclasses')">
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

    <!-- 配置资源 -->
    <div class="config-section">
      <div class="section-title">
        <h2>⚙️ 配置资源</h2>
      </div>
      <div class="config-grid">
        <div class="config-card" @click="navigateTo('/c/default/config/configmaps')">
          <div class="config-icon">🗂️</div>
          <div class="config-info">
            <div class="config-name">ConfigMaps</div>
            <div class="config-count">{{ configStats.configmaps }}</div>
          </div>
        </div>

        <div class="config-card" @click="navigateTo('/c/default/config/secrets')">
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
        <a class="quick-link" @click="navigateTo('/clusters')">
          <div class="link-icon">🏛️</div>
          <div class="link-text">集群管理</div>
        </a>
        <a class="quick-link" @click="navigateTo('/c/default/nodes')">
          <div class="link-icon">💻</div>
          <div class="link-text">节点管理</div>
        </a>
        <a class="quick-link" @click="navigateTo('/c/default/namespaces')">
          <div class="link-icon">📁</div>
          <div class="link-text">命名空间</div>
        </a>
        <a class="quick-link" @click="navigateTo('/cicd/pipelines')">
          <div class="link-icon">🚀</div>
          <div class="link-text">CI/CD 流水线</div>
        </a>
        <a class="quick-link" @click="navigateTo('/images/repositories')">
          <div class="link-icon">📷</div>
          <div class="link-text">镜像仓库</div>
        </a>
        <a class="quick-link" @click="navigateTo('/users')">
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
import storageclassApi from '@/api/cluster/storage/storageclass'
import configmapApi from '@/api/cluster/config/configmap'
import secretApi from '@/api/cluster/config/secret'
import namespaceApi from '@/api/cluster/config/namespace'

const router = useRouter()
const clusterStore = useClusterStore()

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
      const clusters = res.data.list || []
      
      if (clusters.length === 0) {
        Message.warning({ content: '暂无可用集群，请先添加集群' })
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

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    // 先确保有默认集群
    const hasCluster = await ensureDefaultCluster()
    if (!hasCluster) {
      // 使用模拟数据
      useMockData()
      return
    }
    
    // 并发请求所有数据
    await Promise.allSettled([
      loadClusterStats(),
      loadNodeStats(),
      loadPodStats(),
      loadNamespaceStats(),
      loadWorkloadStats(),
      loadNetworkStats(),
      loadStorageStats(),
      loadConfigStats()
    ])
  } catch (error) {
    console.error('加载数据失败:', error)
    useMockData()
  } finally {
    loading.value = false
  }
}

// 使用模拟数据
const useMockData = () => {
  clusterStats.value = { total: 3, active: 2, inactive: 1 }
  nodeStats.value = { total: 5, ready: 4, notReady: 1 }
  podStats.value = { total: 48, running: 42, pending: 3, failed: 3 }
  namespaceStats.value = { total: 12, system: 4, user: 8 }
  workloadStats.value = { deployments: 15, statefulsets: 6, daemonsets: 8, jobs: 12, cronjobs: 5 }
  networkStats.value = { services: 24, ingresses: 8 }
  storageStats.value = { pvs: 10, pvcs: 18, storageClasses: 3 }
  configStats.value = { configmaps: 28, secrets: 15 }
}

// 加载集群统计
const loadClusterStats = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const clusters = res.data.list || []
      clusterStats.value.total = clusters.length
      // status: 0=正常, 1=异常, 2=待检测
      clusterStats.value.active = clusters.filter(c => Number(c.status) === 0).length
      clusterStats.value.inactive = clusters.filter(c => Number(c.status) === 1 || Number(c.status) === 2).length
    }
  } catch (error) {
    console.error('加载集群数据失败:', error)
    clusterStats.value = { total: 3, active: 2, inactive: 1 }
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
    nodeStats.value = { total: 5, ready: 4, notReady: 1 }
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
    podStats.value = { total: 48, running: 42, pending: 3, failed: 3 }
  }
}

// 加载命名空间统计
const loadNamespaceStats = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const namespaces = res.data.list || res.data || []
      namespaceStats.value.total = namespaces.length
      const systemNs = ['kube-system', 'kube-public', 'kube-node-lease', 'default']
      namespaceStats.value.system = namespaces.filter(ns => 
        systemNs.includes(ns.name)
      ).length
      namespaceStats.value.user = namespaceStats.value.total - namespaceStats.value.system
    }
  } catch (error) {
    console.error('加载命名空间数据失败:', error)
    namespaceStats.value = { total: 12, system: 4, user: 8 }
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 24px;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.3);
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

.stat-card.cluster-card { border-left-color: #667eea; }
.stat-card.node-card { border-left-color: #48bb78; }
.stat-card.pod-card { border-left-color: #ed8936; }
.stat-card.namespace-card { border-left-color: #4299e1; }

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

.workload-icon.deployment { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.workload-icon.statefulset { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.workload-icon.daemonset { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.workload-icon.job { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
.workload-icon.cronjob { background: linear-gradient(135deg, #fa709a 0%, #fee140 100%); }

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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  color: #667eea;
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
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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

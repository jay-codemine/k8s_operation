<template>
  <div class="health-dashboard">
    <!-- 头部 -->
    <div class="header">
      <div class="title-section">
        <h2 class="title"><span class="title-icon">🏥</span> 平台健康中心</h2>
        <span class="subtitle">实时监控平台运行状态 · 支持多集群汇总视图</span>
      </div>
      <div class="actions">
        <!-- 集群选择器 -->
        <div class="cluster-selector">
          <select v-model="selectedCluster" class="cluster-select">
            <option value="all">📊 全部集群汇总</option>
            <option v-for="c in clusterDetails" :key="c.id" :value="c.id">
              {{ getClusterIcon(c) }} {{ c.name }}
            </option>
          </select>
        </div>
        <span class="last-update" v-if="lastUpdate">
          <span class="dot" :class="statusClass"></span>
          最后更新: {{ lastUpdateText }}
        </span>
        <div class="auto-refresh">
          <label class="switch">
            <input type="checkbox" v-model="autoRefresh" @change="toggleAutoRefresh">
            <span class="slider"></span>
          </label>
          <span class="switch-label">自动刷新</span>
        </div>
        <button class="refresh-btn" @click="loadHealth" :disabled="loading">
          <span class="refresh-icon" :class="{ spinning: loading }">⟳</span>
          {{ loading ? '刷新中...' : '刷新' }}
        </button>
      </div>
    </div>

    <div v-if="loading && !health" class="loading-container">
      <div class="spinner"></div>
      <p>正在获取健康状态...</p>
    </div>

    <template v-else-if="health">
      <!-- 第一行：平台状态 + 集群概览 + 节点概览 -->
      <div class="overview-row">
        <div class="overview-card platform-card" :class="platformStatusClass">
          <div class="card-header">
            <span class="card-icon">🖥️</span>
            <span class="card-title">平台状态</span>
            <span class="status-indicator" :class="platformStatusClass">{{ platformStatusText }}</span>
          </div>
          <div class="platform-metrics">
            <div class="metric-item">
              <span class="metric-value">{{ health.platform?.uptime || '-' }}</span>
              <span class="metric-label">运行时长</span>
            </div>
            <div class="metric-item">
              <span class="metric-value">{{ health.platform?.num_goroutine || 0 }}</span>
              <span class="metric-label">协程数</span>
            </div>
            <div class="metric-item">
              <span class="metric-value">{{ health.platform?.num_cpu || 0 }}</span>
              <span class="metric-label">CPU核心</span>
            </div>
          </div>
          <div class="card-footer">{{ health.platform?.version || 'v1.0.0' }} · {{ health.platform?.go_version || 'Go' }}</div>
        </div>

        <div class="overview-card">
          <div class="card-header">
            <span class="card-icon">☸️</span>
            <span class="card-title">集群概览</span>
          </div>
          <div class="cluster-stats">
            <div class="big-stat">
              <span class="big-num success">{{ health.clusters?.online || 0 }}</span>
              <span class="stat-divider">/</span>
              <span class="big-num total">{{ health.clusters?.total || 0 }}</span>
            </div>
            <span class="stat-label">在线 / 总数</span>
          </div>
          <div class="mini-stats">
            <div class="mini-stat"><span class="mini-value warning">{{ health.clusters?.offline || 0 }}</span><span class="mini-label">离线</span></div>
          </div>
        </div>

        <div class="overview-card">
          <div class="card-header">
            <span class="card-icon">🖲️</span>
            <span class="card-title">节点概览{{ selectedCluster !== 'all' ? ' (当前集群)' : ' (汇总)' }}</span>
          </div>
          <div class="node-stats">
            <div class="node-count">
              <span class="big-num success">{{ currentNodes.ready || 0 }}</span>
              <span class="stat-divider">/</span>
              <span class="big-num total">{{ currentNodes.total || 0 }}</span>
            </div>
            <span class="stat-label">就绪 / 总数</span>
          </div>
          <div class="node-roles">
            <span class="role-badge master">Master {{ currentNodes.master || 0 }}</span>
            <span class="role-badge worker">Worker {{ currentNodes.worker || 0 }}</span>
          </div>
        </div>
      </div>

      <!-- 集群列表 (可展开详情) -->
      <div class="cluster-list-section" v-if="clusterDetails.length > 0">
        <div class="section-header">
          <h3 class="section-title"><span class="section-icon">☸️</span> 集群列表</h3>
          <span class="cluster-count">{{ clusterDetails.length }} 个集群</span>
        </div>
        <div class="cluster-list">
          <div v-for="cluster in clusterDetails" :key="cluster.id" class="cluster-item" :class="getClusterStatusClass(cluster)">
            <div class="cluster-header" @click="toggleCluster(cluster.id)">
              <div class="cluster-info">
                <span class="cluster-status-dot" :class="cluster.status"></span>
                <span class="cluster-name">{{ cluster.name }}</span>
                <span class="cluster-id">#{{ cluster.id }}</span>
                <span class="cluster-badge" :class="cluster.status">{{ getStatusLabel(cluster.status) }}</span>
              </div>
              <div class="cluster-summary">
                <span class="summary-item" v-if="cluster.connectable">
                  <span class="item-icon">🖲️</span> {{ cluster.nodes?.total || 0 }} 节点
                </span>
                <span class="summary-item" v-if="cluster.connectable">
                  <span class="item-icon">📦</span> {{ cluster.workloads?.pods?.total || 0 }} Pods
                </span>
                <span class="summary-item" v-if="cluster.connectable">
                  <span class="item-icon">🚀</span> {{ cluster.workloads?.deployments?.total || 0 }} Deployments
                </span>
                <span class="latency-badge" :class="{ 'latency-timeout': cluster.latency === 'timeout' }" v-if="cluster.latency && cluster.latency !== '-'">{{ cluster.latency === 'timeout' ? '⬛ 超时' : cluster.latency }}</span>
                <!-- 单独检测按钮 -->
                <button class="check-btn" @click.stop="checkSingleCluster(cluster.id)" :disabled="checkingCluster === cluster.id" :title="'检测集群 ' + cluster.name + ' 连通性'">
                  <span :class="{ 'spinning': checkingCluster === cluster.id }">{{ checkingCluster === cluster.id ? '↻' : '📶' }}</span>
                </button>
              </div>
              <span class="expand-icon">{{ expandedClusters.includes(cluster.id) ? '▼' : '▶' }}</span>
            </div>
            
            <!-- 展开详情 -->
            <div class="cluster-details" v-show="expandedClusters.includes(cluster.id)">
              <div v-if="!cluster.connectable" class="connect-error">
                <span class="error-icon">⚠️</span> 
                {{ cluster.latency === 'timeout' ? '集群连接超时（>30s），请检查集群状态' : '无法连接到此集群' }}
                <button class="retry-connectivity-btn" @click.stop="checkSingleCluster(cluster.id)" :disabled="checkingCluster === cluster.id">
                  {{ checkingCluster === cluster.id ? '检测中...' : '重新检测' }}
                </button>
              </div>
              <template v-else>
                <div class="detail-grid">
                  <!-- 节点 -->
                  <div class="detail-card">
                    <div class="detail-title">节点状态</div>
                    <div class="detail-stats">
                      <div class="stat-row"><span class="stat-name">就绪</span><span class="stat-val success">{{ cluster.nodes?.ready || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">未就绪</span><span class="stat-val danger">{{ cluster.nodes?.not_ready || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Master</span><span class="stat-val">{{ cluster.nodes?.master || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Worker</span><span class="stat-val">{{ cluster.nodes?.worker || 0 }}</span></div>
                    </div>
                  </div>
                  <!-- 工作负载 -->
                  <div class="detail-card">
                    <div class="detail-title">工作负载</div>
                    <div class="detail-stats">
                      <div class="stat-row"><span class="stat-name">Deployments</span><span class="stat-val">{{ cluster.workloads?.deployments?.running || 0 }}/{{ cluster.workloads?.deployments?.total || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">StatefulSets</span><span class="stat-val">{{ cluster.workloads?.statefulsets?.running || 0 }}/{{ cluster.workloads?.statefulsets?.total || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">DaemonSets</span><span class="stat-val">{{ cluster.workloads?.daemonsets?.running || 0 }}/{{ cluster.workloads?.daemonsets?.total || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Jobs/CronJobs</span><span class="stat-val">{{ cluster.workloads?.jobs?.total || 0 }}/{{ cluster.workloads?.cronjobs?.total || 0 }}</span></div>
                    </div>
                  </div>
                  <!-- Pods -->
                  <div class="detail-card">
                    <div class="detail-title">Pod 状态</div>
                    <div class="detail-stats">
                      <div class="stat-row"><span class="stat-name">Running</span><span class="stat-val success">{{ cluster.workloads?.pods?.running || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Pending</span><span class="stat-val warning">{{ cluster.workloads?.pods?.pending || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Succeeded</span><span class="stat-val">{{ cluster.workloads?.pods?.succeeded || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Failed</span><span class="stat-val danger">{{ cluster.workloads?.pods?.failed || 0 }}</span></div>
                    </div>
                  </div>
                  <!-- 服务 -->
                  <div class="detail-card">
                    <div class="detail-title">服务 & 事件</div>
                    <div class="detail-stats">
                      <div class="stat-row"><span class="stat-name">Services</span><span class="stat-val">{{ cluster.services?.total || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Ingresses</span><span class="stat-val">{{ cluster.services?.ingresses || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">事件(今日)</span><span class="stat-val">{{ cluster.events?.today || 0 }}</span></div>
                      <div class="stat-row"><span class="stat-name">Warning</span><span class="stat-val warning">{{ cluster.events?.warning || 0 }}</span></div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- 资源使用率 -->
      <div class="usage-section">
        <div class="section-header">
          <h3 class="section-title"><span class="section-icon">⚡</span> 资源使用率{{ selectedCluster !== 'all' ? ' (当前集群)' : ' (全部集群汇总)' }}</h3>
          <span class="usage-cluster-info" v-if="selectedCluster === 'all'">共 {{ health.clusters?.total || 0 }} 个集群</span>
          <span class="usage-cluster-info" v-else>{{ currentClusterData?.name || '' }}</span>
        </div>
        <div class="usage-row">
        <div class="usage-card">
          <div class="usage-header"><span class="usage-icon">⚡</span><span class="usage-title">CPU 使用率</span></div>
          <div class="usage-gauge"><div class="gauge-circle" :style="{ '--usage': currentNodes.cpu_usage || 0 }"><span class="gauge-value">{{ (currentNodes.cpu_usage || 0).toFixed(1) }}%</span></div></div>
          <div class="usage-bar"><div class="bar-fill" :class="getUsageClass(currentNodes.cpu_usage)" :style="{ width: (currentNodes.cpu_usage || 0) + '%' }"></div></div>
        </div>
        <div class="usage-card">
          <div class="usage-header"><span class="usage-icon">💾</span><span class="usage-title">内存使用率</span></div>
          <div class="usage-gauge"><div class="gauge-circle" :style="{ '--usage': currentNodes.memory_usage || 0 }"><span class="gauge-value">{{ (currentNodes.memory_usage || 0).toFixed(1) }}%</span></div></div>
          <div class="usage-bar"><div class="bar-fill" :class="getUsageClass(currentNodes.memory_usage)" :style="{ width: (currentNodes.memory_usage || 0) + '%' }"></div></div>
        </div>
        <div class="usage-card">
          <div class="usage-header"><span class="usage-icon">📦</span><span class="usage-title">Pod 使用率</span></div>
          <div class="usage-gauge"><div class="gauge-circle" :style="{ '--usage': currentNodes.pod_usage || 0 }"><span class="gauge-value">{{ (currentNodes.pod_usage || 0).toFixed(1) }}%</span></div></div>
          <div class="usage-bar"><div class="bar-fill" :class="getUsageClass(currentNodes.pod_usage)" :style="{ width: (currentNodes.pod_usage || 0) + '%' }"></div></div>
        </div>
        </div>
      </div>

      <!-- 工作负载统计 -->
      <div class="workload-section">
        <div class="section-header">
          <h3 class="section-title"><span class="section-icon">📊</span> 工作负载统计{{ selectedCluster !== 'all' ? ' (当前集群)' : ' (汇总)' }}</h3>
        </div>
        <div class="workload-grid">
          <div class="workload-card">
            <div class="workload-icon">🚀</div>
            <div class="workload-info">
              <span class="workload-name">Deployments</span>
              <div class="workload-stats"><span class="stat-running">{{ currentWorkloads.deployments?.running || 0 }} 运行</span><span class="stat-total">/ {{ currentWorkloads.deployments?.total || 0 }} 总数</span></div>
            </div>
            <div class="workload-badge" :class="getWorkloadBadgeClass(currentWorkloads.deployments)">{{ getWorkloadStatus(currentWorkloads.deployments) }}</div>
          </div>
          <div class="workload-card">
            <div class="workload-icon">🔄</div>
            <div class="workload-info">
              <span class="workload-name">StatefulSets</span>
              <div class="workload-stats"><span class="stat-running">{{ currentWorkloads.statefulsets?.running || 0 }} 运行</span><span class="stat-total">/ {{ currentWorkloads.statefulsets?.total || 0 }} 总数</span></div>
            </div>
            <div class="workload-badge" :class="getWorkloadBadgeClass(currentWorkloads.statefulsets)">{{ getWorkloadStatus(currentWorkloads.statefulsets) }}</div>
          </div>
          <div class="workload-card">
            <div class="workload-icon">👹</div>
            <div class="workload-info">
              <span class="workload-name">DaemonSets</span>
              <div class="workload-stats"><span class="stat-running">{{ currentWorkloads.daemonsets?.running || 0 }} 运行</span><span class="stat-total">/ {{ currentWorkloads.daemonsets?.total || 0 }} 总数</span></div>
            </div>
            <div class="workload-badge" :class="getWorkloadBadgeClass(currentWorkloads.daemonsets)">{{ getWorkloadStatus(currentWorkloads.daemonsets) }}</div>
          </div>
          <div class="workload-card">
            <div class="workload-icon">⏱️</div>
            <div class="workload-info">
              <span class="workload-name">Jobs</span>
              <div class="workload-stats"><span class="stat-running">{{ currentWorkloads.jobs?.running || 0 }} 成功</span><span class="stat-total">/ {{ currentWorkloads.jobs?.total || 0 }} 总数</span></div>
            </div>
            <div class="workload-badge" :class="getWorkloadBadgeClass(currentWorkloads.jobs)">{{ getWorkloadStatus(currentWorkloads.jobs) }}</div>
          </div>
          <div class="workload-card">
            <div class="workload-icon">🕐</div>
            <div class="workload-info">
              <span class="workload-name">CronJobs</span>
              <div class="workload-stats"><span class="stat-running">{{ currentWorkloads.cronjobs?.running || 0 }} 活跃</span><span class="stat-total">/ {{ currentWorkloads.cronjobs?.total || 0 }} 总数</span></div>
            </div>
            <div class="workload-badge badge-ok">正常</div>
          </div>
          <div class="workload-card pods-card">
            <div class="workload-icon">📦</div>
            <div class="workload-info">
              <span class="workload-name">Pods</span>
              <div class="pod-breakdown">
                <span class="pod-stat running">{{ currentWorkloads.pods?.running || 0 }} Running</span>
                <span class="pod-stat pending">{{ currentWorkloads.pods?.pending || 0 }} Pending</span>
                <span class="pod-stat succeeded">{{ currentWorkloads.pods?.succeeded || 0 }} Succeeded</span>
                <span class="pod-stat failed">{{ currentWorkloads.pods?.failed || 0 }} Failed</span>
              </div>
            </div>
            <div class="workload-badge" :class="currentWorkloads.pods?.failed > 0 ? 'badge-warning' : 'badge-ok'">{{ currentWorkloads.pods?.total || 0 }} 总数</div>
          </div>
        </div>
      </div>

      <!-- 服务 & 事件 & 任务队列 -->
      <div class="services-events-row">
        <div class="services-card">
          <div class="card-header"><span class="card-icon">🌐</span><span class="card-title">服务统计</span></div>
          <div class="services-grid">
            <div class="service-item"><span class="service-count">{{ currentServices.total || 0 }}</span><span class="service-label">总服务数</span></div>
            <div class="service-item"><span class="service-count">{{ currentServices.cluster_ip || 0 }}</span><span class="service-label">ClusterIP</span></div>
            <div class="service-item"><span class="service-count">{{ currentServices.node_port || 0 }}</span><span class="service-label">NodePort</span></div>
            <div class="service-item"><span class="service-count">{{ currentServices.load_balancer || 0 }}</span><span class="service-label">LoadBalancer</span></div>
            <div class="service-item highlight"><span class="service-count">{{ currentServices.ingresses || 0 }}</span><span class="service-label">Ingresses</span></div>
          </div>
        </div>
        <div class="events-card">
          <div class="card-header"><span class="card-icon">📋</span><span class="card-title">K8s 事件</span></div>
          <div class="events-stats">
            <div class="event-item"><span class="event-count warning">{{ currentEvents.warning || 0 }}</span><span class="event-label">Warning</span></div>
            <div class="event-item"><span class="event-count normal">{{ currentEvents.normal || 0 }}</span><span class="event-label">Normal</span></div>
            <div class="event-item"><span class="event-count">{{ currentEvents.today || 0 }}</span><span class="event-label">今日</span></div>
            <div class="event-item"><span class="event-count">{{ currentEvents.last_hour || 0 }}</span><span class="event-label">最近1小时</span></div>
          </div>
        </div>
        <div class="queue-card">
          <div class="card-header"><span class="card-icon">🔧</span><span class="card-title">CI/CD 任务队列</span></div>
          <div class="queue-stats">
            <div class="queue-item"><span class="queue-count pending">{{ health.task_queue?.pending || 0 }}</span><span class="queue-label">待处理</span></div>
            <div class="queue-item"><span class="queue-count running">{{ health.task_queue?.running || 0 }}</span><span class="queue-label">运行中</span></div>
            <div class="queue-item"><span class="queue-count success">{{ health.task_queue?.completed || 0 }}</span><span class="queue-label">已完成</span></div>
            <div class="queue-item"><span class="queue-count failed">{{ health.task_queue?.failed || 0 }}</span><span class="queue-label">失败</span></div>
          </div>
        </div>
      </div>

      <!-- 组件状态 -->
      <div class="components-section">
        <div class="section-header">
          <h3 class="section-title"><span class="section-icon">🧩</span> 系统组件状态</h3>
          <span class="component-summary">{{ healthyComponents }}/{{ health.components?.length || 0 }} 组件正常</span>
        </div>
        <div class="components-grid">
          <div v-for="comp in health.components" :key="comp.name" class="component-card" :class="getComponentClass(comp)">
            <div class="component-header">
              <span class="component-icon">{{ getComponentIcon(comp.name) }}</span>
              <span class="component-name">{{ comp.name }}</span>
              <span class="status-badge" :class="'status-' + comp.status">{{ getStatusText(comp.status) }}</span>
            </div>
            <div class="component-body">
              <div class="component-metric" v-if="comp.latency"><span class="metric-label">响应延迟</span><span class="metric-value">{{ comp.latency }}</span></div>
              <div class="component-desc">{{ comp.description || '-' }}</div>
            </div>
            <div class="component-footer">
              <span class="check-time">{{ comp.checked_at }}</span>
              <span v-if="comp.memory" class="memory-info">{{ comp.memory }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="error" class="error-container">
      <div class="error-icon">⚠️</div>
      <h3>获取健康状态失败</h3>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadHealth">重试</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import http from '@/api/http'

const health = ref(null)
const loading = ref(false)
const error = ref(null)
const lastUpdate = ref(null)
const autoRefresh = ref(true)
const selectedCluster = ref('all')
const expandedClusters = ref([])
const checkingCluster = ref(null) // 当前正在检测的集群ID
let refreshTimer = null

// 计算属性
const clusterDetails = computed(() => health.value?.cluster_details || [])

const currentClusterData = computed(() => {
  if (selectedCluster.value === 'all') return null
  return clusterDetails.value.find(c => c.id === selectedCluster.value)
})

const currentNodes = computed(() => currentClusterData.value?.nodes || health.value?.nodes || {})
const currentWorkloads = computed(() => currentClusterData.value?.workloads || health.value?.workloads || {})
const currentServices = computed(() => currentClusterData.value?.services || health.value?.services || {})
const currentEvents = computed(() => currentClusterData.value?.events || health.value?.events || {})

const platformStatusClass = computed(() => ({
  'status-healthy': health.value?.platform?.status === 'healthy',
  'status-degraded': health.value?.platform?.status === 'degraded',
  'status-unhealthy': health.value?.platform?.status === 'unhealthy'
}))

const platformStatusText = computed(() => {
  const map = { healthy: '运行正常', degraded: '性能降级', unhealthy: '服务异常' }
  return map[health.value?.platform?.status] || '检测中...'
})

const statusClass = computed(() => {
  const s = health.value?.platform?.status
  if (s === 'healthy') return 'dot-ok'
  if (s === 'degraded') return 'dot-warning'
  return 'dot-error'
})

const lastUpdateText = computed(() => lastUpdate.value?.toLocaleString('zh-CN') || '-')
const healthyComponents = computed(() => health.value?.components?.filter(c => c.status === 'ok').length || 0)

// 方法
const loadHealth = async () => {
  loading.value = true
  error.value = null
  
  // 创建30秒超时控制
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 30000)
  
  try {
    const res = await http.get('/api/v1/platform/health', { signal: controller.signal })
    clearTimeout(timeoutId)
    if (res.code === 0 && res.data) {
      health.value = res.data
      lastUpdate.value = new Date()
    } else {
      throw new Error(res.msg || '获取数据失败')
    }
  } catch (err) {
    clearTimeout(timeoutId)
    if (err.name === 'AbortError' || err.message?.includes('timeout')) {
      error.value = '请求超时，请检查集群连接状态'
      Message.error({ content: '健康检查超时（>30s），集群可能异常' })
    } else {
      error.value = err.message || '网络请求失败'
      Message.error({ content: '获取健康状态失败: ' + err.message })
    }
  } finally {
    loading.value = false
  }
}

const toggleAutoRefresh = () => {
  autoRefresh.value ? startAutoRefresh() : stopAutoRefresh()
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshTimer = setInterval(loadHealth, 30000)
}

const stopAutoRefresh = () => {
  if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null }
}

const toggleCluster = (id) => {
  const idx = expandedClusters.value.indexOf(id)
  if (idx >= 0) expandedClusters.value.splice(idx, 1)
  else expandedClusters.value.push(id)
}

// 单独检测某个集群的连通性
const checkSingleCluster = async (clusterId) => {
  if (checkingCluster.value) return
  checkingCluster.value = clusterId
  
  try {
    const res = await http.get(`/api/v1/platform/health/cluster/${clusterId}/connectivity`)
    if (res.code === 0 && res.data) {
      const result = res.data
      // 更新对应集群的状态
      const clusterIdx = health.value?.cluster_details?.findIndex(c => c.id === clusterId)
      if (clusterIdx >= 0 && health.value?.cluster_details) {
        const cluster = health.value.cluster_details[clusterIdx]
        cluster.connectable = result.connected
        cluster.latency = result.latency || '-'
        cluster.status = result.connected ? 'online' : 'error'
        
        // 重新计算集群汇总
        const details = health.value.cluster_details
        health.value.clusters = {
          total: details.length,
          online: details.filter(d => d.connectable && d.status === 'online').length,
          offline: details.filter(d => !d.connectable && d.status !== 'error').length,
          abnormal: details.filter(d => d.status === 'error' || d.latency === 'timeout').length
        }
      }
      Message.success({ content: `集群 ${result.cluster_name || clusterId} ${result.connected ? '连接正常' : '连接失败'}，延迟: ${result.latency}` })
    } else {
      Message.error({ content: res.msg || '检测失败' })
    }
  } catch (err) {
    Message.error({ content: '检测集群连通性失败: ' + (err.message || '网络错误') })
  } finally {
    checkingCluster.value = null
  }
}

const getClusterIcon = (c) => c.status === 'online' ? '🟢' : c.status === 'error' ? '🔴' : '🟡'
const getStatusLabel = (s) => ({ online: '在线', offline: '离线', error: '连接失败', timeout: '超时' }[s] || s)
const getClusterStatusClass = (c) => ({ 'cluster-online': c.status === 'online', 'cluster-offline': c.status !== 'online' })
const getUsageClass = (u) => (u >= 90 ? 'usage-critical' : u >= 70 ? 'usage-warning' : 'usage-normal')
const getWorkloadBadgeClass = (w) => !w ? 'badge-ok' : w.failed > 0 ? 'badge-danger' : w.running < w.total ? 'badge-warning' : 'badge-ok'
const getWorkloadStatus = (w) => !w || w.total === 0 ? '无' : w.failed > 0 ? `${w.failed} 异常` : w.running < w.total ? '部分就绪' : '正常'
const getComponentIcon = (n) => ({
  'API Server': '🚀',
  'MySQL': '🗄️',
  'PostgreSQL': '🐘',
  'Redis': '📦',
  'Kubernetes': '☸️',
  'ETCD': '💾',
  'Controller Manager': '🎛️',
  'Scheduler': '📅',
  'CoreDNS': '🌐'
}[n] || '🔧')
const getStatusText = (s) => ({ ok: 'OK', warning: 'WARNING', error: 'ERROR' }[s] || s?.toUpperCase())
const getComponentClass = (c) => ({ 'component-ok': c.status === 'ok', 'component-warning': c.status === 'warning', 'component-error': c.status === 'error' })

onMounted(() => { loadHealth(); if (autoRefresh.value) startAutoRefresh() })
onUnmounted(() => stopAutoRefresh())
</script>

<style scoped>
.health-dashboard { padding: 20px; min-height: 100vh; background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%); }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; flex-wrap: wrap; gap: 16px; }
.title-section { display: flex; flex-direction: column; }
.title { margin: 0; font-size: 24px; font-weight: 700; color: #1e293b; display: flex; align-items: center; gap: 8px; }
.title-icon { font-size: 28px; }
.subtitle { color: #64748b; font-size: 13px; margin-top: 4px; }
.actions { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.cluster-selector { position: relative; }
.cluster-select { padding: 8px 32px 8px 12px; font-size: 14px; border: 2px solid #e2e8f0; border-radius: 8px; background: white; cursor: pointer; color: #1e293b; font-weight: 500; }
.cluster-select:focus { border-color: #3b82f6; outline: none; }
.last-update { display: flex; align-items: center; gap: 6px; color: #64748b; font-size: 13px; }
.dot { width: 8px; height: 8px; border-radius: 50%; animation: pulse 2s infinite; }
.dot-ok { background: #10b981; }
.dot-warning { background: #f59e0b; }
.dot-error { background: #ef4444; }
@keyframes pulse { 0%, 100% { opacity: 1; transform: scale(1); } 50% { opacity: 0.6; transform: scale(1.2); } }
.auto-refresh { display: flex; align-items: center; gap: 8px; }
.switch { position: relative; width: 44px; height: 24px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #cbd5e1; transition: 0.3s; border-radius: 24px; }
.slider:before { position: absolute; content: ""; height: 18px; width: 18px; left: 3px; bottom: 3px; background-color: white; transition: 0.3s; border-radius: 50%; box-shadow: 0 2px 4px rgba(0,0,0,0.15); }
.switch input:checked + .slider { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.switch input:checked + .slider:before { transform: translateX(20px); }
.switch-label { font-size: 13px; color: #64748b; }
.refresh-btn { display: flex; align-items: center; gap: 6px; padding: 10px 20px; background: linear-gradient(135deg, #3b82f6, #2563eb); color: white; border: none; border-radius: 10px; font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.3s; box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
.refresh-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4); }
.refresh-btn:disabled { opacity: 0.7; cursor: not-allowed; }
.refresh-icon { font-size: 16px; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.loading-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80px 20px; }
.spinner { width: 48px; height: 48px; border: 4px solid #e2e8f0; border-top-color: #3b82f6; border-radius: 50%; animation: spin 1s linear infinite; }
.overview-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 20px; }
.overview-card { background: white; border-radius: 16px; padding: 20px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); transition: transform 0.3s, box-shadow 0.3s; }
.overview-card:hover { transform: translateY(-4px); box-shadow: 0 8px 30px rgba(0, 0, 0, 0.1); }
.platform-card.status-healthy { border-left: 4px solid #10b981; }
.platform-card.status-degraded { border-left: 4px solid #f59e0b; }
.platform-card.status-unhealthy { border-left: 4px solid #ef4444; }
.card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; }
.card-icon { font-size: 20px; }
.card-title { font-size: 15px; font-weight: 600; color: #1e293b; flex: 1; }
.status-indicator { padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; }
.status-indicator.status-healthy { background: #dcfce7; color: #16a34a; }
.status-indicator.status-degraded { background: #fef3c7; color: #d97706; }
.status-indicator.status-unhealthy { background: #fee2e2; color: #dc2626; }
.platform-metrics { display: flex; justify-content: space-between; margin-bottom: 16px; }
.metric-item { display: flex; flex-direction: column; align-items: center; }
.metric-value { font-size: 20px; font-weight: 700; color: #1e293b; }
.metric-label { font-size: 12px; color: #64748b; margin-top: 4px; }
.card-footer { font-size: 11px; color: #94a3b8; padding-top: 12px; border-top: 1px solid #f1f5f9; text-align: center; }
.cluster-stats, .node-stats { text-align: center; margin-bottom: 16px; }
.big-stat { display: flex; justify-content: center; align-items: baseline; gap: 4px; }
.big-num { font-size: 32px; font-weight: 800; }
.big-num.success { color: #10b981; }
.big-num.total { color: #64748b; font-size: 24px; }
.stat-divider { color: #cbd5e1; font-size: 24px; margin: 0 4px; }
.stat-label { font-size: 13px; color: #64748b; }
.mini-stats { display: flex; justify-content: center; gap: 24px; }
.mini-stat { display: flex; flex-direction: column; align-items: center; }
.mini-value { font-size: 18px; font-weight: 700; }
.mini-value.warning { color: #f59e0b; }
.mini-label { font-size: 12px; color: #64748b; }
.node-roles { display: flex; justify-content: center; gap: 12px; }
.role-badge { padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 500; }
.role-badge.master { background: #ede9fe; color: #7c3aed; }
.role-badge.worker { background: #e0f2fe; color: #0284c7; }
/* 集群列表 */
.cluster-list-section { background: white; border-radius: 16px; padding: 24px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); margin-bottom: 20px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.section-title { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; display: flex; align-items: center; gap: 8px; }
.section-icon { font-size: 18px; }
.cluster-count { color: #64748b; font-size: 14px; }
.cluster-list { display: flex; flex-direction: column; gap: 12px; }
.cluster-item { border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; transition: all 0.3s; }
.cluster-item.cluster-online { border-left: 4px solid #10b981; }
.cluster-item.cluster-offline { border-left: 4px solid #f59e0b; }
.cluster-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; cursor: pointer; background: #fafafa; transition: background 0.2s; }
.cluster-header:hover { background: #f1f5f9; }
.cluster-info { display: flex; align-items: center; gap: 12px; }
.cluster-status-dot { width: 10px; height: 10px; border-radius: 50%; }
.cluster-status-dot.online { background: #10b981; }
.cluster-status-dot.offline { background: #f59e0b; }
.cluster-status-dot.error { background: #ef4444; }
.cluster-name { font-weight: 600; color: #1e293b; font-size: 15px; }
.cluster-id { color: #94a3b8; font-size: 12px; }
.cluster-badge { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.cluster-badge.online { background: #dcfce7; color: #16a34a; }
.cluster-badge.offline { background: #fef3c7; color: #d97706; }
.cluster-badge.error { background: #fee2e2; color: #dc2626; }
.cluster-summary { display: flex; align-items: center; gap: 16px; }
.summary-item { display: flex; align-items: center; gap: 4px; color: #64748b; font-size: 13px; }
.item-icon { font-size: 14px; }
.latency-badge { padding: 2px 8px; background: #e0f2fe; color: #0284c7; border-radius: 8px; font-size: 11px; font-weight: 500; }
.latency-badge.latency-timeout { background: #fee2e2; color: #dc2626; }
.expand-icon { color: #94a3b8; font-size: 12px; }
.cluster-details { padding: 20px; background: white; border-top: 1px solid #e2e8f0; }
.connect-error { display: flex; align-items: center; gap: 8px; color: #ef4444; padding: 16px; background: #fef2f2; border-radius: 8px; flex-wrap: wrap; }
.error-icon { font-size: 20px; }
/* 集群连通性检测按钮 */
.check-btn { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; padding: 0; background: #e0f2fe; border: 1px solid #7dd3fc; border-radius: 6px; cursor: pointer; font-size: 14px; transition: all 0.2s; margin-left: 8px; }
.check-btn:hover:not(:disabled) { background: #bae6fd; border-color: #38bdf8; transform: scale(1.1); }
.check-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.retry-connectivity-btn { margin-left: auto; padding: 6px 14px; background: #3b82f6; color: white; border: none; border-radius: 6px; font-size: 12px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.retry-connectivity-btn:hover:not(:disabled) { background: #2563eb; }
.retry-connectivity-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.detail-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.detail-card { background: #f8fafc; border-radius: 10px; padding: 16px; }
.detail-title { font-size: 13px; font-weight: 600; color: #64748b; margin-bottom: 12px; border-bottom: 1px solid #e2e8f0; padding-bottom: 8px; }
.detail-stats { display: flex; flex-direction: column; gap: 8px; }
.stat-row { display: flex; justify-content: space-between; font-size: 13px; }
.stat-name { color: #64748b; }
.stat-val { font-weight: 600; color: #1e293b; }
.stat-val.success { color: #10b981; }
.stat-val.warning { color: #f59e0b; }
.stat-val.danger { color: #ef4444; }
/* 使用率 */
.usage-section { background: white; border-radius: 16px; padding: 24px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); margin-bottom: 20px; }
.usage-cluster-info { color: #64748b; font-size: 14px; background: #f1f5f9; padding: 4px 12px; border-radius: 20px; }
.usage-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.usage-card { background: #f8fafc; border-radius: 12px; padding: 20px; text-align: center; transition: all 0.3s; }
.usage-card:hover { background: #f1f5f9; transform: translateY(-2px); }
.usage-header { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 16px; }
.usage-icon { font-size: 18px; }
.usage-title { font-size: 14px; font-weight: 600; color: #1e293b; }
.usage-gauge { margin-bottom: 16px; }
.gauge-circle { width: 80px; height: 80px; border-radius: 50%; background: conic-gradient(#3b82f6 calc(var(--usage) * 3.6deg), #e2e8f0 calc(var(--usage) * 3.6deg)); display: flex; align-items: center; justify-content: center; margin: 0 auto; position: relative; }
.gauge-circle::before { content: ''; position: absolute; width: 60px; height: 60px; background: #f8fafc; border-radius: 50%; }
.gauge-value { position: relative; z-index: 1; font-size: 16px; font-weight: 700; color: #1e293b; }
.usage-bar { height: 8px; background: #e2e8f0; border-radius: 4px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 4px; transition: width 0.5s ease; }
.bar-fill.usage-normal { background: linear-gradient(90deg, #10b981, #059669); }
.bar-fill.usage-warning { background: linear-gradient(90deg, #f59e0b, #d97706); }
.bar-fill.usage-critical { background: linear-gradient(90deg, #ef4444, #dc2626); }
/* 工作负载 */
.workload-section, .components-section { background: white; border-radius: 16px; padding: 24px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); margin-bottom: 20px; }
.workload-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.workload-card { display: flex; align-items: center; gap: 16px; padding: 16px; background: #f8fafc; border-radius: 12px; transition: all 0.3s; }
.workload-card:hover { background: #f1f5f9; transform: translateX(4px); }
.workload-icon { font-size: 28px; }
.workload-info { flex: 1; }
.workload-name { font-size: 14px; font-weight: 600; color: #1e293b; display: block; margin-bottom: 4px; }
.workload-stats { font-size: 13px; }
.stat-running { color: #10b981; font-weight: 600; }
.stat-total { color: #64748b; }
.workload-badge { padding: 4px 12px; border-radius: 20px; font-size: 11px; font-weight: 600; }
.badge-ok { background: #dcfce7; color: #16a34a; }
.badge-warning { background: #fef3c7; color: #d97706; }
.badge-danger { background: #fee2e2; color: #dc2626; }
.pods-card { grid-column: span 3; }
.pod-breakdown { display: flex; gap: 16px; flex-wrap: wrap; }
.pod-stat { font-size: 13px; font-weight: 500; }
.pod-stat.running { color: #10b981; }
.pod-stat.pending { color: #f59e0b; }
.pod-stat.succeeded { color: #3b82f6; }
.pod-stat.failed { color: #ef4444; }
/* 服务 & 事件 */
.services-events-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 20px; }
.services-card, .events-card, .queue-card { background: white; border-radius: 16px; padding: 20px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); }
.services-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-top: 16px; }
.service-item { text-align: center; padding: 12px 8px; background: #f8fafc; border-radius: 10px; }
.service-item.highlight { background: #ede9fe; }
.service-count { font-size: 20px; font-weight: 700; color: #1e293b; display: block; }
.service-label { font-size: 11px; color: #64748b; }
.events-stats, .queue-stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-top: 16px; }
.event-item, .queue-item { text-align: center; }
.event-count, .queue-count { font-size: 24px; font-weight: 700; display: block; }
.event-count.warning, .queue-count.pending { color: #f59e0b; }
.event-count.normal { color: #10b981; }
.queue-count.running { color: #3b82f6; }
.queue-count.success { color: #10b981; }
.queue-count.failed { color: #ef4444; }
.event-label, .queue-label { font-size: 12px; color: #64748b; }
/* 组件 */
.component-summary { color: #64748b; font-size: 14px; }
.components-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.component-card { background: #f8fafc; border-radius: 12px; padding: 16px; border-left: 4px solid #e2e8f0; transition: all 0.3s; }
.component-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); }
.component-card.component-ok { border-left-color: #10b981; }
.component-card.component-warning { border-left-color: #f59e0b; }
.component-card.component-error { border-left-color: #ef4444; }
.component-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.component-icon { font-size: 20px; }
.component-name { font-weight: 600; color: #1e293b; flex: 1; }
.status-badge { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.status-badge.status-ok { background: #dcfce7; color: #16a34a; }
.status-badge.status-warning { background: #fef3c7; color: #d97706; }
.status-badge.status-error { background: #fee2e2; color: #dc2626; }
.component-body { margin-bottom: 12px; }
.component-metric { display: flex; justify-content: space-between; margin-bottom: 8px; padding: 8px; background: white; border-radius: 6px; }
.component-desc { font-size: 12px; color: #64748b; }
.component-footer { display: flex; justify-content: space-between; font-size: 11px; color: #94a3b8; padding-top: 8px; border-top: 1px solid #e2e8f0; }
.memory-info { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; }
.error-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80px 20px; text-align: center; }
.error-container .error-icon { font-size: 48px; margin-bottom: 16px; }
.error-container h3 { color: #1e293b; margin-bottom: 8px; }
.error-container p { color: #64748b; margin-bottom: 24px; }
.retry-btn { padding: 10px 24px; background: #3b82f6; color: white; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; }
.retry-btn:hover { background: #2563eb; }
@media (max-width: 1200px) {
  .overview-row, .usage-row, .services-events-row { grid-template-columns: repeat(2, 1fr); }
  .workload-grid, .components-grid { grid-template-columns: repeat(2, 1fr); }
  .detail-grid { grid-template-columns: repeat(2, 1fr); }
  .pods-card { grid-column: span 2; }
}
@media (max-width: 768px) {
  .overview-row, .usage-row, .services-events-row, .workload-grid, .components-grid, .detail-grid { grid-template-columns: 1fr; }
  .pods-card { grid-column: span 1; }
  .header { flex-direction: column; align-items: flex-start; }
  .actions { width: 100%; justify-content: flex-start; }
}
</style>

<template>
  <div class="home-page">
    <!-- 顶部工具条 -->
    <div class="page-toolbar">
      <div class="toolbar-title">
        <h1>平台总览</h1>
        <span class="conn-badge" :class="clusterConnectionStatus">
          <span v-if="clusterConnectionStatus === 'checking'"><span class="spinner"></span> 连接中</span>
          <span v-else-if="clusterConnectionStatus === 'connected'">● 已连接</span>
          <span v-else-if="clusterConnectionStatus === 'timeout'">● 部分超时</span>
          <span v-else>● 连接失败</span>
        </span>
      </div>
      <div class="toolbar-actions">
        <select v-model.number="autoRefresh" class="auto-refresh-select" @change="setupAutoRefresh">
          <option :value="0">自动刷新：关闭</option>
          <option :value="30">自动刷新 30s</option>
          <option :value="60">自动刷新 60s</option>
        </select>
        <button class="btn-refresh" :disabled="loading" @click="refreshData">
          {{ loading ? '刷新中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 顶部统计卡片 -->
    <div v-if="hasClusterAccess" class="stats-row">
      <div v-if="canView('clusters')" class="stat-card clickable" @click="handleNavigate('/clusters')">
        <div class="stat-top">
          <span class="stat-label">集群总数</span>
          <span class="stat-emoji bg-indigo">🏛️</span>
        </div>
        <div class="stat-value">{{ clusterStats.total }}</div>
        <div class="stat-badges">
          <span class="badge badge-green">运行中 {{ clusterStats.active }}</span>
          <span class="badge badge-red">异常 {{ clusterStats.inactive }}</span>
        </div>
      </div>

      <div v-if="canView('nodes')" class="stat-card clickable" @click="handleClusterNavigate('nodes')">
        <div class="stat-top">
          <span class="stat-label">节点总数</span>
          <span class="stat-emoji bg-blue">🖥️</span>
        </div>
        <div class="stat-value">{{ nodeStats.total }}</div>
        <div class="stat-badges">
          <span class="badge badge-green">正常 {{ nodeStats.ready }}</span>
          <span class="badge badge-red">异常 {{ nodeStats.notReady }}</span>
        </div>
      </div>

      <div v-if="canView('pods')" class="stat-card clickable" @click="handleClusterNavigate('workloads/pods')">
        <div class="stat-top">
          <span class="stat-label">Pod 总数</span>
          <span class="stat-emoji bg-green">📦</span>
        </div>
        <div class="stat-value">{{ podStats.total }}</div>
        <div class="stat-badges">
          <span class="badge badge-green">运行中 {{ podStats.running }}</span>
          <span class="badge badge-amber">待调度 {{ podStats.pending }}</span>
        </div>
      </div>

      <div class="stat-card ring-card">
        <div class="stat-top">
          <span class="stat-label">CPU 使用率</span>
        </div>
        <div class="ring-body">
          <svg viewBox="0 0 72 72" class="ring">
            <circle cx="36" cy="36" r="30" class="ring-bg" />
            <circle cx="36" cy="36" r="30" class="ring-val" stroke="#3b82f6"
              :stroke-dasharray="ringDash(overview.cpu_usage)" transform="rotate(-90 36 36)" />
            <text x="36" y="41" text-anchor="middle" class="ring-text">{{ overview.cpu_usage.toFixed(0) }}%</text>
          </svg>
          <div class="ring-sub">所有节点平均<br />告警阈值 80%</div>
        </div>
      </div>

      <div class="stat-card ring-card">
        <div class="stat-top">
          <span class="stat-label">内存使用率</span>
        </div>
        <div class="ring-body">
          <svg viewBox="0 0 72 72" class="ring">
            <circle cx="36" cy="36" r="30" class="ring-bg" />
            <circle cx="36" cy="36" r="30" class="ring-val" stroke="#10b981"
              :stroke-dasharray="ringDash(overview.memory_usage)" transform="rotate(-90 36 36)" />
            <text x="36" y="41" text-anchor="middle" class="ring-text">{{ overview.memory_usage.toFixed(0) }}%</text>
          </svg>
          <div class="ring-sub">所有节点平均<br />告警阈值 85%</div>
        </div>
      </div>

      <div class="stat-card clickable" @click="handleNavigate('/monitoring/alert-events')">
        <div class="stat-top">
          <span class="stat-label">告警数量</span>
          <span class="stat-emoji bg-red">🔔</span>
        </div>
        <div class="stat-value">{{ alertTotal }}</div>
        <div class="stat-badges">
          <span class="badge badge-red">严重 {{ alertStats.critical }}</span>
          <span class="badge badge-amber">警告 {{ alertStats.warning }}</span>
        </div>
      </div>
    </div>

    <!-- 集群健康状态 + 资源使用趋势 -->
    <div v-if="hasClusterAccess" class="grid-row row-2">
      <div class="panel">
        <div class="panel-header">
          <h3>集群健康状态</h3>
        </div>
        <table class="mini-table">
          <thead>
            <tr><th>集群名称</th><th>状态</th><th>版本</th><th>说明</th></tr>
          </thead>
          <tbody>
            <tr v-for="c in clustersList" :key="c.id" class="clickable" @click="handleNavigate('/clusters')">
              <td class="td-name">{{ c.cluster_name }}</td>
              <td><span class="badge" :class="clusterStatusClass(c.status)">{{ clusterStatusText(c.status) }}</span></td>
              <td class="td-muted">{{ c.cluster_version || '—' }}</td>
              <td class="td-muted td-ellipsis">{{ c.last_error || '—' }}</td>
            </tr>
            <tr v-if="!clustersList.length && !loading"><td colspan="4" class="td-empty">暂无集群数据</td></tr>
          </tbody>
        </table>
        <div class="panel-footer">
          <a class="footer-link" @click="handleNavigate('/clusters')">查看全部集群 →</a>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3>资源使用趋势</h3>
          <div class="panel-tools">
            <div class="tab-group">
              <button v-for="t in trendTabs" :key="t.key" class="tab-btn"
                :class="{ active: trendTab === t.key }" @click="switchTrendTab(t.key)">{{ t.label }}</button>
            </div>
            <select v-model="trendDuration" class="duration-select" @change="loadTrend">
              <option value="1h">最近 1 小时</option>
              <option value="6h">最近 6 小时</option>
              <option value="24h">最近 24 小时</option>
            </select>
          </div>
        </div>
        <div ref="trendChartRef" class="trend-chart"></div>
      </div>
    </div>

    <!-- CI/CD 流水线状态 + 最近发布记录 -->
    <div v-if="canView('pipelines')" class="grid-row row-3">
      <div class="panel">
        <div class="panel-header">
          <h3>CI/CD 流水线状态</h3>
        </div>
        <div class="cicd-chips">
          <div class="chip">
            <span class="chip-icon chip-blue">▶</span>
            <div class="chip-info"><span class="chip-label">运行中</span><span class="chip-value">{{ cicdRunning }}</span></div>
          </div>
          <div class="chip">
            <span class="chip-icon chip-green">✔</span>
            <div class="chip-info"><span class="chip-label">今日构建</span><span class="chip-value">{{ buildStats.today_builds ?? '—' }}</span></div>
          </div>
          <div class="chip">
            <span class="chip-icon chip-amber">％</span>
            <div class="chip-info"><span class="chip-label">成功率 7d</span><span class="chip-value">{{ cicdSuccessRate }}</span></div>
          </div>
          <div class="chip">
            <span class="chip-icon chip-gray">Σ</span>
            <div class="chip-info"><span class="chip-label">流水线总数</span><span class="chip-value">{{ pipelineTotal }}</span></div>
          </div>
        </div>
        <table class="mini-table">
          <thead>
            <tr><th>流水线名称</th><th>状态</th><th>分支</th><th>更新时间</th></tr>
          </thead>
          <tbody>
            <tr v-for="p in pipelines" :key="p.id" class="clickable" @click="handleNavigate(`/cicd/pipelines/${p.id}`)">
              <td class="td-name">{{ p.name }}</td>
              <td><span class="badge" :class="pipeStatusClass(p)">{{ pipeStatusText(p) }}</span></td>
              <td class="td-muted">{{ p.git_branch || p.branch || 'main' }}</td>
              <td class="td-muted">{{ formatTime(p.last_run_time || p.created_at) }}</td>
            </tr>
            <tr v-if="!pipelines.length && !loading"><td colspan="4" class="td-empty">暂无流水线</td></tr>
          </tbody>
        </table>
        <div class="panel-footer">
          <a class="footer-link" @click="handleNavigate('/cicd/pipelines')">查看全部流水线 →</a>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3>最近发布记录</h3>
        </div>
        <div class="release-list">
          <div v-for="rel in releases" :key="rel.id" class="release-item clickable"
            @click="handleNavigate('/cicd/releases')">
            <span class="rel-avatar" :class="'rel-' + releaseStatusKey(rel.status)">
              {{ (rel.app_name || rel.name || '?').charAt(0).toUpperCase() }}
            </span>
            <div class="rel-info">
              <div class="rel-line1">
                <span class="rel-name">{{ rel.app_name || rel.name || '—' }}</span>
                <code v-if="rel.image_tag" class="rel-tag">{{ rel.image_tag }}</code>
              </div>
              <div class="rel-line2">
                <span>{{ rel.creator || rel.created_by || '—' }}</span>
                <span v-if="rel.env || rel.namespace" class="rel-env">{{ rel.env || rel.namespace }}</span>
              </div>
            </div>
            <div class="rel-right">
              <span class="badge" :class="'badge-' + releaseStatusKey(rel.status)">{{ releaseStatusText(rel.status) }}</span>
              <span class="rel-time">{{ formatTime(rel.created_at) }}</span>
            </div>
          </div>
          <div v-if="!releases.length && !loading" class="td-empty">暂无发布记录</div>
        </div>
        <div class="panel-footer">
          <a class="footer-link" @click="handleNavigate('/cicd/releases')">查看全部发布记录 →</a>
        </div>
      </div>
    </div>

    <!-- 关键指标监控 + Pod 状态分布 + 集群事件 -->
    <div v-if="hasClusterAccess" class="grid-row row-4">
      <div class="panel">
        <div class="panel-header">
          <h3>关键指标监控</h3>
          <a class="footer-link" @click="handleNavigate('/monitoring')">查看大盘</a>
        </div>
        <div class="kpi-grid">
          <div v-for="k in kpiCards" :key="k.label" class="kpi-card">
            <div class="kpi-label">{{ k.label }}</div>
            <div class="kpi-value" :style="{ color: k.color }">{{ k.value }}</div>
            <svg class="kpi-spark" viewBox="0 0 120 36" preserveAspectRatio="none">
              <path :d="k.spark" fill="none" :stroke="k.color" stroke-width="1.5" />
            </svg>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3>Pod 状态分布</h3>
          <a class="footer-link" @click="handleNavigate('/monitoring')">更多</a>
        </div>
        <div class="donut-body">
          <svg viewBox="0 0 140 140" class="donut">
            <circle cx="70" cy="70" r="52" class="donut-bg" />
            <circle v-for="(seg, i) in donutSegments" :key="i" cx="70" cy="70" r="52"
              fill="none" :stroke="seg.color" stroke-width="18"
              :stroke-dasharray="seg.dash" :stroke-dashoffset="seg.offset"
              transform="rotate(-90 70 70)" />
            <text x="70" y="66" text-anchor="middle" class="donut-num">{{ podDistTotal }}</text>
            <text x="70" y="84" text-anchor="middle" class="donut-label">总数</text>
          </svg>
          <div class="donut-legend">
            <div v-for="it in podDistItems" :key="it.name" class="legend-item">
              <span class="legend-dot" :style="{ background: it.color }"></span>
              <span class="legend-name">{{ it.label }}</span>
              <span class="legend-count">{{ it.value }} ({{ podDistTotal ? ((it.value / podDistTotal) * 100).toFixed(1) : 0 }}%)</span>
            </div>
            <div v-if="!podDistItems.length && !loading" class="td-empty">暂无数据</div>
          </div>
        </div>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3>集群事件</h3>
          <a class="footer-link" @click="handleNavigate('/monitoring/alert-events')">查看全部</a>
        </div>
        <div class="event-list">
          <div v-for="ev in events" :key="ev.id" class="event-item">
            <span class="event-dot" :class="'sev-' + (ev.severity || 'info')"></span>
            <div class="event-info">
              <div class="event-summary">{{ ev.summary || ev.rule_name || '—' }}</div>
              <div class="event-meta">{{ ev.rule_name }}</div>
            </div>
            <span class="event-time">{{ formatTime(ev.fired_at) }}</span>
          </div>
          <div v-if="!events.length && !loading" class="td-empty">暂无告警事件</div>
        </div>
      </div>
    </div>

    <!-- 系统状态栏 -->
    <div class="system-bar">
      <span class="sys-title">系统状态</span>
      <div class="sys-items">
        <div v-for="comp in healthComponents" :key="comp.name" class="sys-item">
          <span class="sys-name">{{ comp.name }}</span>
          <span class="badge" :class="comp.status === 'ok' ? 'badge-green' : 'badge-red'">
            {{ comp.status === 'ok' ? '正常' : '异常' }}
          </span>
        </div>
        <div v-if="!healthComponents.length" class="td-empty">健康数据加载中...</div>
      </div>
      <div v-if="platformUptime" class="sys-uptime">系统运行时间 <b>{{ platformUptime }}</b></div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import * as echarts from 'echarts'
import { useClusterStore } from '@/stores/cluster'
import permissionStore from '@/stores/permission'

import http from '@/api/http'
import { getClusterList } from '@/api/cluster'
import nodesApi from '@/api/cluster/nodes'
import podsApi from '@/api/cluster/workloads/pods'
import {
  getClusterOverview,
  getResourceTrend,
  getPodStatusDistribution,
  listAlertEvents,
  getAlertStats,
} from '@/api/monitoring'
import { getPipelines, getBuildStats } from '@/api/platform/pipeline'
import { getReleases } from '@/api/cicd'

const router = useRouter()
const clusterStore = useClusterStore()

// =============================================================================
// 超时控制工具
// 设计原则：快速失败、优雅降级、可观测（超时不阻塞其他面板）
// =============================================================================
const API_TIMEOUT = 8000
// 平台健康是聚合接口：需逐个集群做连通性探测，存在离线集群时后端约需 15s 才降级返回，
// 因此单独放宽超时，避免健康面板长期空白（该请求与其他面板并发，不阻塞首屏）
const HEALTH_TIMEOUT = 20000

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

// =============================================================================
// 权限模型（与原版保持一致）
// =============================================================================
const ROLE_RESOURCE_MAP = {
  cluster_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ],
  developer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets',
    'namespaces', 'clusters', 'pipelines'
  ],
  viewer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ],
  cicd_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets',
    'namespaces', 'clusters', 'pipelines'
  ]
}

const PLATFORM_ADMIN_RESOURCES = ['repositories', 'users']

const hasClusterAccess = computed(() => {
  if (permissionStore.state.isSuperAdmin) return true
  const clusterPerms = permissionStore.state.clusterPermissions
  return Object.keys(clusterPerms).length > 0
})

const userHighestRole = computed(() => {
  if (permissionStore.state.isSuperAdmin) return 'super_admin'
  const roleTypes = permissionStore.roleTypes.value
  if (roleTypes.includes('super_admin')) return 'super_admin'
  if (roleTypes.includes('platform_admin')) return 'platform_admin'
  if (roleTypes.includes('cluster_admin')) return 'cluster_admin'
  if (roleTypes.includes('cicd_admin')) return 'cicd_admin'
  if (roleTypes.includes('developer')) return 'developer'
  if (roleTypes.includes('viewer')) return 'viewer'
  return null
})

const userAccessibleResources = computed(() => {
  if (permissionStore.state.isSuperAdmin) {
    return [...ROLE_RESOURCE_MAP.cluster_admin, ...PLATFORM_ADMIN_RESOURCES, 'pipelines']
  }
  const role = userHighestRole.value
  if (!role) return []
  if (role === 'platform_admin') {
    return [...ROLE_RESOURCE_MAP.cluster_admin, ...PLATFORM_ADMIN_RESOURCES, 'pipelines']
  }
  return ROLE_RESOURCE_MAP[role] || []
})

const canView = (resource) => {
  if (permissionStore.state.isSuperAdmin) return true
  const clusterResources = [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'clusters'
  ]
  if (clusterResources.includes(resource)) {
    if (!hasClusterAccess.value) return false
  }
  return userAccessibleResources.value.includes(resource)
}

// =============================================================================
// 导航
// =============================================================================
const handleNavigate = (path) => {
  router.push(path)
}

const handleClusterNavigate = (subPath) => {
  const cid = clusterStore.current?.id
  if (!cid) {
    router.push('/clusters')
    return
  }
  router.push(`/c/${cid}/${subPath}`)
}

// =============================================================================
// 状态数据
// =============================================================================
const loading = ref(true)

const clusterStats = ref({ total: 0, active: 0, inactive: 0 })
const clustersList = ref([])
const nodeStats = ref({ total: 0, ready: 0, notReady: 0 })
const podStats = ref({ total: 0, running: 0, pending: 0, failed: 0 })

// 监控总览（Prometheus 数据源）
const overview = reactive({
  cpu_usage: 0,
  memory_usage: 0,
  node_count: 0,
  pod_count: 0,
  alert_count: 0,
  network_in: 0,
  network_out: 0,
})

const alertStats = ref({ critical: 0, warning: 0, info: 0, total_firing: 0 })
// 告警总数优先取后端 total_firing，降级为各级别求和
const alertTotal = computed(() =>
  alertStats.value.total_firing ||
  (alertStats.value.critical || 0) + (alertStats.value.warning || 0) + (alertStats.value.info || 0)
)

// CI/CD
const pipelines = ref([])
const pipelineTotal = ref(0)
const buildStats = ref({})
const releases = ref([])

// 告警事件（集群事件面板）
const events = ref([])

// 平台健康
const health = ref({})
const healthComponents = computed(() => health.value?.components || [])
const platformUptime = computed(() => health.value?.platform?.uptime || '')

// =============================================================================
// 顶部环形卡片
// =============================================================================
const RING_C = 188.5 // 2 * PI * 30
const ringDash = (pct) => {
  const p = Math.max(0, Math.min(100, Number(pct) || 0))
  return `${(p / 100) * RING_C} ${RING_C}`
}

// =============================================================================
// 集群健康状态表
// =============================================================================
const clusterStatusText = (status) => {
  const map = { 0: '健康', 1: '异常', 2: '待检测' }
  return map[Number(status)] || '未知'
}
const clusterStatusClass = (status) => {
  const map = { 0: 'badge-green', 1: 'badge-red', 2: 'badge-amber' }
  return map[Number(status)] || 'badge-gray'
}

// =============================================================================
// 资源使用趋势（ECharts）
// =============================================================================
const trendTabs = [
  { key: 'cpu', label: 'CPU 使用率' },
  { key: 'memory', label: '内存使用率' },
  { key: 'network', label: '网络流量' },
]
const trendTab = ref('cpu')
const trendDuration = ref('6h')
const trendChartRef = ref(null)
let trendChart = null

const formatBytes = (v) => {
  const n = Number(v) || 0
  if (n >= 1073741824) return (n / 1073741824).toFixed(1) + ' GB'
  if (n >= 1048576) return (n / 1048576).toFixed(1) + ' MB'
  if (n >= 1024) return (n / 1024).toFixed(1) + ' KB'
  return n.toFixed(0) + ' B'
}

const TREND_COLORS = ['#3b82f6', '#f59e0b', '#10b981', '#8b5cf6', '#ef4444']

const buildTrendOption = (seriesList, isPercent) => ({
  color: TREND_COLORS,
  tooltip: {
    trigger: 'axis',
    valueFormatter: (v) => isPercent ? Number(v).toFixed(1) + '%' : formatBytes(v) + '/s'
  },
  grid: { left: 48, right: 16, top: 16, bottom: 44 },
  legend: { bottom: 0, icon: 'circle', itemWidth: 8, itemHeight: 8, textStyle: { fontSize: 11, color: '#64748b' } },
  xAxis: {
    type: 'time',
    axisLabel: { fontSize: 11, color: '#94a3b8' },
    axisLine: { lineStyle: { color: '#e2e8f0' } }
  },
  yAxis: {
    type: 'value',
    max: isPercent ? 100 : null,
    axisLabel: {
      fontSize: 11, color: '#94a3b8',
      formatter: isPercent ? '{value}%' : (v) => formatBytes(v)
    },
    splitLine: { lineStyle: { color: '#f1f5f9' } }
  },
  series: seriesList.map((s, i) => ({
    name: s.label,
    type: 'line',
    smooth: true,
    showSymbol: false,
    data: s.data,
    lineStyle: { width: 2 },
    areaStyle: i === 0 ? { opacity: 0.1 } : undefined
  }))
})

const switchTrendTab = (key) => {
  if (trendTab.value === key) return
  trendTab.value = key
  loadTrend()
}

const loadTrend = async () => {
  try {
    const res = await withTimeout(
      getResourceTrend(trendTab.value, trendDuration.value),
      API_TIMEOUT, '资源趋势'
    )
    if (res.code === 0 && res.data) {
      const seriesList = res.data.map(td => ({
        label: td.label,
        // 后端 timestamp 为秒级 Unix 时间戳，ECharts time 轴需毫秒
        data: (td.points || []).map(p => [p.timestamp * 1000, p.value])
      }))
      await nextTick()
      if (!trendChart && trendChartRef.value) {
        trendChart = echarts.init(trendChartRef.value)
      }
      if (trendChart) {
        trendChart.setOption(buildTrendOption(seriesList, trendTab.value !== 'network'), true)
      }
    }
  } catch (e) {
    console.warn('加载资源趋势失败:', e.message)
  }
}

const onResize = () => {
  trendChart?.resize()
}

// =============================================================================
// 关键指标监控（sparkline 迷你图）
// =============================================================================
const sparkData = ref({ cpu: [], memory: [], netIn: [], netOut: [] })

const sparkPath = (points, w = 120, h = 36) => {
  if (!points || points.length < 2) return ''
  const vals = points.map(p => Number(p.value) || 0)
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const span = max - min || 1
  return points.map((p, i) => {
    const x = (i / (points.length - 1)) * w
    const y = h - 4 - (((Number(p.value) || 0) - min) / span) * (h - 8)
    return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`
  }).join(' ')
}

const kpiCards = computed(() => [
  {
    label: 'CPU 使用率',
    value: overview.cpu_usage.toFixed(1) + '%',
    color: '#3b82f6',
    spark: sparkPath(sparkData.value.cpu)
  },
  {
    label: '内存使用率',
    value: overview.memory_usage.toFixed(1) + '%',
    color: '#10b981',
    spark: sparkPath(sparkData.value.memory)
  },
  {
    label: '入站流量',
    value: formatBytes(overview.network_in) + '/s',
    color: '#8b5cf6',
    spark: sparkPath(sparkData.value.netIn)
  },
  {
    label: '出站流量',
    value: formatBytes(overview.network_out) + '/s',
    color: '#f59e0b',
    spark: sparkPath(sparkData.value.netOut)
  },
])

const loadSparklines = async () => {
  try {
    const [cpuRes, memRes, netRes] = await Promise.allSettled([
      withTimeout(getResourceTrend('cpu', '1h'), API_TIMEOUT, 'CPU迷你图'),
      withTimeout(getResourceTrend('memory', '1h'), API_TIMEOUT, '内存迷你图'),
      withTimeout(getResourceTrend('network', '1h'), API_TIMEOUT, '网络迷你图'),
    ])
    if (cpuRes.status === 'fulfilled' && cpuRes.value.code === 0) {
      sparkData.value.cpu = cpuRes.value.data?.[0]?.points || []
    }
    if (memRes.status === 'fulfilled' && memRes.value.code === 0) {
      sparkData.value.memory = memRes.value.data?.[0]?.points || []
    }
    if (netRes.status === 'fulfilled' && netRes.value.code === 0) {
      const netSeries = netRes.value.data || []
      // 后端 network 系列由 Go map 生成，顺序不固定，按 label 匹配入/出站
      const inSeries = netSeries.find(s => (s.label || '').includes('入'))
      const outSeries = netSeries.find(s => (s.label || '').includes('出'))
      sparkData.value.netIn = inSeries?.points || netSeries[0]?.points || []
      sparkData.value.netOut = outSeries?.points || netSeries[1]?.points || []
    }
  } catch (e) {
    console.warn('加载迷你图数据失败:', e.message)
  }
}

// =============================================================================
// Pod 状态分布（SVG 环形图）
// =============================================================================
const POD_STATUS_META = {
  Running: { label: '运行中', color: '#22c55e' },
  Succeeded: { label: '已完成', color: '#3b82f6' },
  Pending: { label: '待调度', color: '#f59e0b' },
  Failed: { label: '异常', color: '#ef4444' },
  Unknown: { label: '未知', color: '#94a3b8' },
}

const podDistRaw = ref([])

const podDistItems = computed(() => {
  return podDistRaw.value
    .map(it => {
      // 后端 PodStatusCount 字段为 { phase, count }
      const name = it.phase || it.status || it.name || 'Unknown'
      const meta = POD_STATUS_META[name] || { label: name, color: '#a78bfa' }
      return { name, label: meta.label, color: meta.color, value: Number(it.count ?? it.value ?? 0) }
    })
    .filter(it => it.value > 0)
    .sort((a, b) => b.value - a.value)
})

const podDistTotal = computed(() => podDistItems.value.reduce((s, it) => s + it.value, 0))

const DONUT_C = 326.73 // 2 * PI * 52
const donutSegments = computed(() => {
  const total = podDistTotal.value
  if (!total) return []
  let acc = 0
  return podDistItems.value.map(it => {
    const frac = it.value / total
    const seg = {
      color: it.color,
      dash: `${frac * DONUT_C} ${DONUT_C}`,
      offset: -acc * DONUT_C
    }
    acc += frac
    return seg
  })
})

// =============================================================================
// CI/CD 面板
// =============================================================================
const cicdRunning = computed(() => {
  if (buildStats.value.running_builds != null) return buildStats.value.running_builds
  return pipelines.value.filter(p => p.status === 'running').length
})

const cicdSuccessRate = computed(() => {
  const r = buildStats.value.success_rate
  if (r == null || r === '') return '—'
  return typeof r === 'number' ? r.toFixed(1) + '%' : String(r)
})

// 流水线状态（后端 status: idle/running/disabled；last_run_status: pending/running/success/failed/aborted）
const pipeStatusText = (p) => {
  if (p.status === 'running') return '运行中'
  if (p.status === 'disabled') return '已禁用'
  const last = p.last_run_status || p.lastRunStatus || p.last_status
  const map = { success: '成功', failed: '失败', pending: '等待', running: '运行中', aborted: '已中止' }
  return map[last] || '空闲'
}

const pipeStatusClass = (p) => {
  if (p.status === 'running') return 'badge-blue'
  if (p.status === 'disabled') return 'badge-gray'
  const last = p.last_run_status || p.lastRunStatus || p.last_status
  const map = { success: 'badge-green', failed: 'badge-red', pending: 'badge-amber', running: 'badge-blue', aborted: 'badge-gray' }
  return map[last] || 'badge-gray'
}

// 发布状态归一化（后端枚举：Pending/AwaitingApproval/Queued/Running/Succeeded/Failed/Canceled/Rollback）
const releaseStatusKey = (status) => {
  const s = String(status || '').toLowerCase()
  if (s === 'success' || s === 'succeeded') return 'green'
  if (s === 'failed' || s === 'error') return 'red'
  if (s === 'deploying' || s === 'running' || s === 'pending' || s === 'queued' || s === 'awaitingapproval') return 'blue'
  if (s === 'canceled' || s === 'cancelled' || s === 'rollback') return 'gray'
  return 'gray'
}

const releaseStatusText = (status) => {
  const s = String(status || '').toLowerCase()
  const map = {
    success: '成功', succeeded: '成功', failed: '失败', error: '失败',
    deploying: '发布中', running: '发布中', pending: '等待中', queued: '排队中',
    awaitingapproval: '待审批', rollback: '已回滚',
    canceled: '已取消', cancelled: '已取消'
  }
  return map[s] || (status || '—')
}

// =============================================================================
// 时间格式化
// =============================================================================
// 兼容后端秒级 Unix 时间戳（pipeline/release/alert 均为秒级）与 ISO 字符串
const formatTime = (t) => {
  if (!t) return '—'
  let v = t
  const n = Number(t)
  if (!isNaN(n) && n > 0 && n < 1e12) v = n * 1000 // 秒级数字 → 毫秒
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const pad = (n2) => String(n2).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// =============================================================================
// 数据加载
// =============================================================================
const ensureDefaultCluster = async () => {
  try {
    if (clusterStore.current?.id) return true
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0 && res.data) {
      const allClusters = res.data.list || []
      const clusters = allClusters.filter(c =>
        permissionStore.state.isSuperAdmin ||
        permissionStore.state.accessibleClusterIds.includes(c.id)
      )
      if (clusters.length === 0) return false
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

const loadClusterStats = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const allClusters = res.data.list || []
      const clusters = allClusters.filter(c =>
        permissionStore.state.isSuperAdmin ||
        permissionStore.state.accessibleClusterIds.includes(c.id)
      )
      clusterStats.value.total = clusters.length
      clusterStats.value.active = clusters.filter(c => Number(c.status) === 0).length
      clusterStats.value.inactive = clusters.filter(c => Number(c.status) === 1 || Number(c.status) === 2).length
      clustersList.value = clusters.slice(0, 5)
    }
  } catch (error) {
    console.error('加载集群数据失败:', error)
    clusterStats.value = { total: 0, active: 0, inactive: 0 }
    clustersList.value = []
  }
}

const loadNodeStats = async () => {
  try {
    const res = await nodesApi.list({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const nodes = res.data.list || res.data || []
      nodeStats.value.total = nodes.length
      nodeStats.value.ready = nodes.filter(n => n.status === 'Ready' || n.status === 'ready').length
      nodeStats.value.notReady = nodeStats.value.total - nodeStats.value.ready
    }
  } catch (error) {
    console.error('加载节点数据失败:', error)
    nodeStats.value = { total: 0, ready: 0, notReady: 0 }
  }
}

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

const loadOverview = async () => {
  try {
    const res = await getClusterOverview()
    if (res.code === 0 && res.data) {
      Object.assign(overview, res.data)
    }
  } catch (e) {
    console.warn('加载监控总览失败:', e.message)
  }
}

const loadAlertStats = async () => {
  try {
    const res = await getAlertStats()
    if (res.code === 0 && res.data) {
      alertStats.value = {
        critical: res.data.critical || 0,
        warning: res.data.warning || 0,
        info: res.data.info || 0,
        total_firing: res.data.total_firing || 0,
      }
    }
  } catch (e) {
    console.warn('加载告警统计失败:', e.message)
  }
}

const loadPodDistribution = async () => {
  try {
    const res = await getPodStatusDistribution()
    if (res.code === 0) {
      podDistRaw.value = res.data || []
    }
  } catch (e) {
    console.warn('加载 Pod 状态分布失败:', e.message)
  }
}

const loadPipelines = async () => {
  try {
    const [listRes, statsRes] = await Promise.allSettled([
      getPipelines({ page: 1, page_size: 5 }),
      getBuildStats(7),
    ])
    if (listRes.status === 'fulfilled' && listRes.value.code === 0 && listRes.value.data) {
      pipelines.value = listRes.value.data.list || []
      pipelineTotal.value = listRes.value.data.total ?? pipelines.value.length
    }
    if (statsRes.status === 'fulfilled' && statsRes.value.code === 0) {
      buildStats.value = statsRes.value.data?.stats || statsRes.value.data || {}
    }
  } catch (e) {
    console.warn('加载流水线数据失败:', e.message)
  }
}

const loadReleases = async () => {
  try {
    const res = await getReleases({ page: 1, page_size: 5 })
    if (res.code === 0 && res.data) {
      releases.value = res.data.list || []
    }
  } catch (e) {
    console.warn('加载发布记录失败:', e.message)
  }
}

const loadEvents = async () => {
  try {
    const res = await listAlertEvents({ page: 1, size: 5 })
    if (res.code === 0 && res.data) {
      events.value = res.data.items || res.data.list || []
    }
  } catch (e) {
    console.warn('加载告警事件失败:', e.message)
  }
}

const loadHealth = async () => {
  try {
    const res = await http.get('/api/v1/platform/health')
    if (res.code === 0 && res.data) {
      health.value = res.data
    }
  } catch (e) {
    console.warn('加载平台健康状态失败:', e.message)
  }
}

const loadData = async (silent = false) => {
  loading.value = true
  if (!silent) clusterConnectionStatus.value = 'checking'

  try {
    const hasCluster = await withTimeout(ensureDefaultCluster(), 5000, '获取集群列表').catch(() => false)

    if (!hasCluster) {
      clusterConnectionStatus.value = 'timeout'
      if (!silent) Message.warning({ content: '无可用集群或连接超时，暂无数据', duration: 3000 })
      // 平台级数据不依赖集群，仍然加载
      await Promise.allSettled([
        canView('pipelines') ? withTimeout(loadPipelines(), API_TIMEOUT, '流水线') : Promise.resolve(),
        canView('pipelines') ? withTimeout(loadReleases(), API_TIMEOUT, '发布记录') : Promise.resolve(),
        withTimeout(loadHealth(), HEALTH_TIMEOUT, '平台健康'),
      ])
      return
    }

    clusterConnectionStatus.value = 'connected'

    // 并发请求所有面板数据（每个请求独立超时，互不影响）
    const results = await Promise.allSettled([
      withTimeout(loadClusterStats(), API_TIMEOUT, '集群统计'),
      withTimeout(loadNodeStats(), API_TIMEOUT, '节点统计'),
      withTimeout(loadPodStats(), API_TIMEOUT, 'Pod统计'),
      withTimeout(loadOverview(), API_TIMEOUT, '监控总览'),
      withTimeout(loadAlertStats(), API_TIMEOUT, '告警统计'),
      withTimeout(loadPodDistribution(), API_TIMEOUT, 'Pod分布'),
      withTimeout(loadEvents(), API_TIMEOUT, '告警事件'),
      withTimeout(loadHealth(), HEALTH_TIMEOUT, '平台健康'),
      withTimeout(loadSparklines(), API_TIMEOUT * 2, '迷你图'),
      loadTrend(),
      canView('pipelines') ? withTimeout(loadPipelines(), API_TIMEOUT, '流水线') : Promise.resolve(),
      canView('pipelines') ? withTimeout(loadReleases(), API_TIMEOUT, '发布记录') : Promise.resolve(),
    ])

    const failed = results.filter(r => r.status === 'rejected')
    if (failed.length > 0) {
      console.warn(`Dashboard: ${failed.length}/${results.length} 个请求超时或失败`)
      if (failed.length >= results.length / 2) {
        clusterConnectionStatus.value = 'timeout'
      }
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    clusterConnectionStatus.value = 'error'
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await loadData()
  Message.success({ content: '数据刷新成功' })
}

// =============================================================================
// 自动刷新
// =============================================================================
const autoRefresh = ref(30)
let refreshTimer = null

const setupAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  if (autoRefresh.value > 0) {
    refreshTimer = setInterval(() => loadData(true), autoRefresh.value * 1000)
  }
}

onMounted(() => {
  loadData()
  setupAutoRefresh()
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  window.removeEventListener('resize', onResize)
  if (trendChart) {
    trendChart.dispose()
    trendChart = null
  }
})
</script>

<style scoped>
.home-page {
  padding: 20px 24px;
  max-width: 1600px;
  margin: 0 auto;
  background: #f8fafc;
  min-height: calc(100vh - 60px);
}

/* ===== 顶部工具条 ===== */
.page-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-title h1 {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.conn-badge {
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.conn-badge.connected { background: #dcfce7; color: #16a34a; }
.conn-badge.checking { background: #e0e7ff; color: #4f46e5; }
.conn-badge.timeout { background: #fef3c7; color: #d97706; }
.conn-badge.error { background: #fee2e2; color: #dc2626; }

.spinner {
  display: inline-block;
  width: 10px;
  height: 10px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  vertical-align: -1px;
}

@keyframes spin { to { transform: rotate(360deg); } }

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.auto-refresh-select,
.duration-select {
  height: 32px;
  padding: 0 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
  color: #475569;
  font-size: 12px;
  cursor: pointer;
  outline: none;
}

.btn-refresh {
  height: 32px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  background: #4f46e5;
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-refresh:hover { background: #4338ca; }
.btn-refresh:disabled { opacity: 0.6; cursor: not-allowed; }

/* ===== 顶部统计卡片 ===== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 14px;
  margin-bottom: 16px;
}

.stat-card {
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
  transition: transform 0.15s, box-shadow 0.15s;
}

.stat-card.clickable { cursor: pointer; }
.stat-card.clickable:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.08);
}

.stat-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.stat-emoji {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
}

.bg-indigo { background: #eef2ff; }
.bg-blue { background: #eff6ff; }
.bg-green { background: #ecfdf5; }
.bg-red { background: #fef2f2; }

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.2;
  margin-bottom: 8px;
}

.stat-badges {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.badge {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
  white-space: nowrap;
}

.badge-green { background: #dcfce7; color: #16a34a; }
.badge-red { background: #fee2e2; color: #dc2626; }
.badge-amber { background: #fef3c7; color: #d97706; }
.badge-blue { background: #dbeafe; color: #2563eb; }
.badge-gray { background: #f1f5f9; color: #64748b; }

/* 环形卡片 */
.ring-body {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ring {
  width: 72px;
  height: 72px;
  flex-shrink: 0;
}

.ring-bg {
  fill: none;
  stroke: #f1f5f9;
  stroke-width: 7;
}

.ring-val {
  fill: none;
  stroke-width: 7;
  stroke-linecap: round;
  transition: stroke-dasharray 0.6s ease;
}

.ring-text {
  font-size: 15px;
  font-weight: 700;
  fill: #0f172a;
}

.ring-sub {
  font-size: 11px;
  color: #94a3b8;
  line-height: 1.6;
}

/* ===== 面板通用 ===== */
.grid-row {
  display: grid;
  gap: 14px;
  margin-bottom: 16px;
}

.row-2 { grid-template-columns: 5fr 7fr; }
.row-3 { grid-template-columns: 7fr 5fr; }
.row-4 { grid-template-columns: 1fr 1fr 1fr; }

.panel {
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 18px;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
  flex-wrap: wrap;
}

.panel-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.panel-tools {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.tab-group {
  display: flex;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 2px;
}

.tab-btn {
  border: none;
  background: transparent;
  padding: 5px 12px;
  font-size: 12px;
  color: #64748b;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}

.tab-btn.active {
  background: #4f46e5;
  color: #fff;
  font-weight: 500;
}

.panel-footer {
  margin-top: auto;
  padding-top: 12px;
  text-align: center;
}

.footer-link {
  font-size: 12px;
  color: #4f46e5;
  cursor: pointer;
  font-weight: 500;
}

.footer-link:hover { text-decoration: underline; }

/* ===== 迷你表格 ===== */
.mini-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.mini-table th {
  text-align: left;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  padding: 8px 10px;
  border-bottom: 1px solid #f1f5f9;
}

.mini-table td {
  padding: 10px;
  border-bottom: 1px solid #f8fafc;
  color: #334155;
}

.mini-table tr.clickable { cursor: pointer; }
.mini-table tr.clickable:hover td { background: #f8fafc; }

.td-name { font-weight: 600; color: #0f172a; }
.td-muted { color: #64748b; font-size: 12px; }
.td-ellipsis {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-empty {
  text-align: center;
  color: #94a3b8;
  font-size: 12px;
  padding: 20px 0;
}

/* ===== 趋势图 ===== */
.trend-chart {
  width: 100%;
  height: 280px;
  min-height: 280px;
}

/* ===== CI/CD 面板 ===== */
.cicd-chips {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 14px;
}

.chip {
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid #eef2f7;
  border-radius: 10px;
  padding: 10px 12px;
  background: #fafbfc;
}

.chip-icon {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #fff;
  flex-shrink: 0;
}

.chip-blue { background: #3b82f6; }
.chip-green { background: #22c55e; }
.chip-amber { background: #f59e0b; }
.chip-gray { background: #94a3b8; }

.chip-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.chip-label { font-size: 11px; color: #94a3b8; }
.chip-value { font-size: 17px; font-weight: 700; color: #0f172a; }

/* ===== 发布记录 ===== */
.release-list {
  display: flex;
  flex-direction: column;
}

.release-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 6px;
  border-bottom: 1px solid #f8fafc;
  border-radius: 8px;
}

.release-item.clickable { cursor: pointer; }
.release-item.clickable:hover { background: #f8fafc; }

.rel-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.rel-green { background: #22c55e; }
.rel-red { background: #ef4444; }
.rel-blue { background: #3b82f6; }
.rel-gray { background: #94a3b8; }

.rel-info { flex: 1; min-width: 0; }

.rel-line1 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;
}

.rel-name {
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rel-tag {
  font-size: 11px;
  background: #f1f5f9;
  color: #475569;
  padding: 1px 6px;
  border-radius: 5px;
  flex-shrink: 0;
}

.rel-line2 {
  font-size: 11px;
  color: #94a3b8;
  display: flex;
  gap: 8px;
}

.rel-env {
  background: #eef2ff;
  color: #4f46e5;
  padding: 0 6px;
  border-radius: 5px;
}

.rel-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  flex-shrink: 0;
}

.rel-time { font-size: 11px; color: #94a3b8; }

/* ===== 关键指标监控 ===== */
.kpi-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.kpi-card {
  border: 1px solid #eef2f7;
  border-radius: 10px;
  padding: 12px;
  background: #fafbfc;
}

.kpi-label { font-size: 12px; color: #64748b; margin-bottom: 4px; }

.kpi-value {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 6px;
}

.kpi-spark {
  width: 100%;
  height: 36px;
  display: block;
}

/* ===== Pod 状态分布 ===== */
.donut-body {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.donut {
  width: 140px;
  height: 140px;
  flex-shrink: 0;
}

.donut-bg {
  fill: none;
  stroke: #f1f5f9;
  stroke-width: 18;
}

.donut-num {
  font-size: 24px;
  font-weight: 700;
  fill: #0f172a;
}

.donut-label {
  font-size: 11px;
  fill: #94a3b8;
}

.donut-legend {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-name { color: #475569; }
.legend-count { margin-left: auto; color: #94a3b8; }

/* ===== 集群事件 ===== */
.event-list {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.event-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 9px 4px;
  border-bottom: 1px solid #f8fafc;
}

.event-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-top: 5px;
  flex-shrink: 0;
}

.sev-critical { background: #ef4444; }
.sev-warning { background: #f59e0b; }
.sev-info { background: #3b82f6; }

.event-info { flex: 1; min-width: 0; }

.event-summary {
  font-size: 12px;
  color: #334155;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.event-meta {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.event-time {
  font-size: 11px;
  color: #94a3b8;
  flex-shrink: 0;
}

/* ===== 系统状态栏 ===== */
.system-bar {
  background: #fff;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  padding: 14px 18px;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.04);
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.sys-title {
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  flex-shrink: 0;
}

.sys-items {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
  flex: 1;
}

.sys-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.sys-name { font-size: 12px; color: #64748b; }

.sys-uptime {
  font-size: 12px;
  color: #64748b;
  flex-shrink: 0;
}

.sys-uptime b { color: #0f172a; }

/* ===== 响应式 ===== */
@media (max-width: 1400px) {
  .stats-row { grid-template-columns: repeat(3, 1fr); }
}

@media (max-width: 1100px) {
  .row-2, .row-3, .row-4 { grid-template-columns: 1fr; }
  .cicd-chips { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .home-page { padding: 12px; }
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .stats-row, .grid-row { gap: 10px; }
  .donut-body { flex-direction: column; }
}
</style>

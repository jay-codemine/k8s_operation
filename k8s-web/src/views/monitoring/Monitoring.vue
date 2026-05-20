<template>
  <div class="monitoring-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">{{ viewTitle }}</h2>
        <span class="page-subtitle">{{ viewSubtitle }}</span>
      </div>
      <div class="header-right">
        <!-- 数据源切换器 -->
        <div class="ds-switcher" :class="{ open: dsSwitcherOpen }">
          <button class="ds-switcher-btn" @click="dsSwitcherOpen = !dsSwitcherOpen">
            <span class="ds-switcher-dot" :class="currentDs.status || 'unknown'"></span>
            <span class="ds-switcher-icon">{{ getDsIcon(currentDs.type) }}</span>
            <span class="ds-switcher-name">{{ currentDs.name || '未选择数据源' }}</span>
            <span class="ds-switcher-arrow">▾</span>
          </button>
          <div class="ds-switcher-dropdown" v-if="dsSwitcherOpen">
            <div class="ds-dropdown-header">
              <span>切换数据源</span>
              <router-link to="/monitoring/datasources" class="ds-manage-link" @click="dsSwitcherOpen = false">管理 →</router-link>
            </div>
            <!-- 数据源类型分组 -->
            <div class="ds-dropdown-list">
              <div class="ds-type-group" v-for="group in dsGroups" :key="group.type">
                <div class="ds-group-title">
                  <span class="ds-group-icon">{{ getDsIcon(group.type) }}</span>
                  <span>{{ getDsTypeName(group.type) }}</span>
                  <span class="ds-group-count">{{ group.items.length }}</span>
                </div>
                <div
                  class="ds-dropdown-item"
                  v-for="ds in group.items"
                  :key="ds.id"
                  :class="{ active: ds.id === currentDs.id, disabled: !ds.enabled }"
                  @click="switchDatasource(ds)"
                >
                  <div class="ds-item-info">
                    <span class="ds-item-name">{{ ds.name }}</span>
                    <span class="ds-item-url">{{ ds.url }}</span>
                  </div>
                  <span class="ds-item-status" :class="ds.status || 'unknown'">●</span>
                  <span class="ds-item-badge" v-if="ds.is_default">默认</span>
                </div>
              </div>
              <div class="ds-dropdown-empty" v-if="!dsList.length">
                暂无数据源，<router-link to="/monitoring/datasources">去添加</router-link>
              </div>
            </div>
          </div>
        </div>
        <!-- 视图模式标签 -->
        <div class="view-mode-tag" :class="currentViewMode">
          <span class="view-mode-icon">{{ currentViewMode === 'metrics' ? '📈' : '📜' }}</span>
          <span>{{ currentViewMode === 'metrics' ? 'Metrics' : 'Logs' }}</span>
        </div>
        <a class="targets-link" :href="prometheusTargetsURL" target="_blank" v-if="prometheusTargetsURL && healthy">
          {{ targetsLinkText }}
        </a>
        <button class="refresh-btn" @click="refreshAll" :disabled="loading">
          <span :class="{ spinning: loading }">↻</span> 刷新
        </button>
      </div>
    </div>

    <!-- 未连接提示 -->
    <div class="connect-hint" v-if="!healthy && !loading && currentViewMode === 'metrics'">
      <div class="hint-content">
        <span class="hint-icon">⚠️</span>
        <div class="hint-text">
          <b>{{ currentDs.type === 'loki' ? 'Loki' : 'Prometheus' }} 未连接</b>
          <p>请前往「数据源管理」配置数据源地址，或检查网络连通性</p>
        </div>
        <router-link to="/monitoring/datasources" class="hint-btn">前往配置 →</router-link>
      </div>
    </div>

    <!-- 已连接但无数据提示 -->
    <div class="connect-hint info-hint" v-if="healthy && !loading && isAllZero && currentViewMode === 'metrics'">
      <div class="hint-content">
        <span class="hint-icon">💡</span>
        <div class="hint-text">
          <b>Prometheus 已连接，但未采集到监控指标</b>
          <p>
            当前数据源: <code>{{ datasourceInfo }}</code><br/>
            可能原因: 集群中未部署 <b>node_exporter</b>（节点指标）或 <b>kube-state-metrics</b>（K8s对象指标）。
            请确认 Prometheus 的 Targets 页面是否有对应采集任务。
          </p>
        </div>
        <a :href="prometheusTargetsURL" target="_blank" class="hint-btn" v-if="prometheusTargetsURL">
          {{ targetsLinkText }}
        </a>
      </div>
    </div>

    <!-- ========== Loki 日志视图 ========== -->
    <LokiView v-if="currentViewMode === 'logs'" />

    <!-- ========== Prometheus 指标视图 ========== -->
    <template v-if="currentViewMode === 'metrics'">
    <!-- 指标卡片 -->
    <div class="metrics-grid">
      <div class="metric-card" v-for="m in metricCards" :key="m.label">
        <div class="metric-icon" :style="{ background: m.bg }">{{ m.icon }}</div>
        <div class="metric-info">
          <span class="metric-value">{{ m.value }}</span>
          <span class="metric-label">{{ m.label }}</span>
        </div>
      </div>
    </div>

    <!-- 趋势图表区域 -->
    <div class="charts-grid">
      <div class="chart-card">
        <div class="chart-header">
          <h3>CPU 使用率趋势</h3>
          <select v-model="trendDuration" @change="loadTrends" class="duration-select">
            <option value="1h">近 1 小时</option>
            <option value="3h">近 3 小时</option>
            <option value="6h">近 6 小时</option>
            <option value="24h">近 24 小时</option>
          </select>
        </div>
        <div ref="cpuChartRef" class="chart-body"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>内存使用率趋势</h3>
        </div>
        <div ref="memChartRef" class="chart-body"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>磁盘使用率趋势</h3>
        </div>
        <div ref="diskChartRef" class="chart-body"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>网络流量趋势</h3>
        </div>
        <div ref="networkChartRef" class="chart-body"></div>
      </div>
    </div>

    <!-- 节点列表 + Top Pods -->
    <div class="bottom-grid">
      <!-- 节点指标 -->
      <div class="table-card">
        <div class="table-header">
          <h3>节点资源使用</h3>
        </div>
        <table class="data-table">
          <thead>
            <tr>
              <th>节点</th>
              <th>CPU</th>
              <th>内存</th>
              <th>磁盘</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="node in nodes" :key="node.name">
              <td class="node-name">{{ node.name }}</td>
              <td>
                <div class="usage-bar-wrap">
                  <div class="usage-bar" :style="{ width: node.cpu_usage + '%', background: getUsageColor(node.cpu_usage) }"></div>
                  <span>{{ node.cpu_usage.toFixed(1) }}%</span>
                </div>
              </td>
              <td>
                <div class="usage-bar-wrap">
                  <div class="usage-bar" :style="{ width: node.memory_usage + '%', background: getUsageColor(node.memory_usage) }"></div>
                  <span>{{ node.memory_usage.toFixed(1) }}%</span>
                </div>
              </td>
              <td>
                <div class="usage-bar-wrap">
                  <div class="usage-bar" :style="{ width: node.disk_usage + '%', background: getUsageColor(node.disk_usage) }"></div>
                  <span>{{ node.disk_usage.toFixed(1) }}%</span>
                </div>
              </td>
              <td>
                <span class="status-tag" :class="node.status === 'Ready' ? 'ready' : 'not-ready'">
                  {{ node.status }}
                </span>
              </td>
            </tr>
            <tr v-if="nodes.length === 0">
              <td colspan="5" class="empty-row">暂无节点数据</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Top Pods -->
      <div class="table-card">
        <div class="table-header">
          <h3>资源占用 Top 10</h3>
          <div class="tab-switch">
            <button :class="{ active: topMetric === 'cpu' }" @click="switchTopMetric('cpu')">CPU</button>
            <button :class="{ active: topMetric === 'memory' }" @click="switchTopMetric('memory')">内存</button>
          </div>
        </div>
        <table class="data-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Pod</th>
              <th>命名空间</th>
              <th>{{ topMetric === 'cpu' ? 'CPU (cores)' : '内存 (MB)' }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(pod, i) in topPods" :key="pod.name">
              <td class="rank-cell">
                <span class="rank-badge" :class="'rank-' + (i + 1)">{{ i + 1 }}</span>
              </td>
              <td class="pod-name">{{ pod.name }}</td>
              <td><span class="ns-tag">{{ pod.namespace }}</span></td>
              <td class="usage-value">
                {{ topMetric === 'cpu' ? pod.cpu_usage.toFixed(3) : (pod.memory_usage / 1048576).toFixed(1) }}
              </td>
            </tr>
            <tr v-if="topPods.length === 0">
              <td colspan="4" class="empty-row">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick, computed } from 'vue'
import * as echarts from 'echarts'
import LokiView from './LokiView.vue'
import {
  getClusterOverview,
  getNodeMetrics,
  getResourceTrend,
  getTopPods,
  checkHealth,
  listDatasources,
} from '@/api/monitoring'

// ===== 状态 =====
const loading = ref(false)
const healthy = ref(false)
const datasourceInfo = ref('')

// 数据源切换
const dsSwitcherOpen = ref(false)
const dsList = ref([])
const currentDs = ref({})

// 根据数据源类型决定视图模式
const currentViewMode = computed(() => {
  const type = currentDs.value?.type || 'prometheus'
  if (type === 'loki') return 'logs'
  return 'metrics'
})

// 页面标题根据视图模式动态变化
const viewTitle = computed(() => {
  if (currentViewMode.value === 'logs') return '日志探索'
  return '集群资源总览'
})

const viewSubtitle = computed(() => {
  if (currentViewMode.value === 'logs') return '实时查询、分析和探索集群日志数据'
  return '实时监控集群 CPU、内存、磁盘、网络等核心指标'
})

// 数据源按类型分组
const dsGroups = computed(() => {
  const groups = {}
  for (const ds of dsList.value) {
    if (!groups[ds.type]) {
      groups[ds.type] = { type: ds.type, items: [] }
    }
    groups[ds.type].items.push(ds)
  }
  // 按类型排序: prometheus > loki > 其他
  const order = ['prometheus', 'victoriametrics', 'thanos', 'loki', 'alertmanager', 'grafana', 'n9e']
  return Object.values(groups).sort((a, b) => {
    return (order.indexOf(a.type) === -1 ? 99 : order.indexOf(a.type)) -
           (order.indexOf(b.type) === -1 ? 99 : order.indexOf(b.type))
  })
})

function getDsIcon(type) {
  const map = { prometheus: '🔥', loki: '📜', alertmanager: '🚨', victoriametrics: '📈', grafana: '📊', n9e: '🦉', thanos: '♾️' }
  return map[type] || '📡'
}

function getDsTypeName(type) {
  const map = { prometheus: 'Prometheus', loki: 'Loki', alertmanager: 'Alertmanager', victoriametrics: 'VictoriaMetrics', grafana: 'Grafana', n9e: 'Nightingale', thanos: 'Thanos' }
  return map[type] || type
}

async function loadDatasources() {
  try {
    const res = await listDatasources({ page: 1, size: 100 })
    if (res?.code === 0) {
      dsList.value = res.data?.items || []
      // 默认选中：优先用当前已选，否则用默认数据源，否则用第一个已连接的
      if (!currentDs.value.id && dsList.value.length) {
        const def = dsList.value.find(d => d.is_default && d.enabled)
        const first = dsList.value.find(d => d.enabled)
        currentDs.value = def || first || dsList.value[0]
      }
    }
  } catch {}
}

function switchDatasource(ds) {
  if (!ds.enabled) return
  currentDs.value = ds
  dsSwitcherOpen.value = false
  // 更新连接信息并刷新
  datasourceInfo.value = ds.url
  healthy.value = ds.status === 'connected'
  refreshAll()
}

// 点击外部关闭下拉
function handleClickOutside(e) {
  if (!e.target.closest('.ds-switcher')) dsSwitcherOpen.value = false
}
const overview = reactive({
  cpu_usage: 0,
  memory_usage: 0,
  disk_usage: 0,
  node_count: 0,
  pod_count: 0,
  alert_count: 0,
  network_in: 0,
  network_out: 0,
})
const nodes = ref([])
const topPods = ref([])
const topMetric = ref('cpu')
const trendDuration = ref('1h')

// 计算是否所有指标都为0（已连接但无数据）
const isAllZero = computed(() => {
  return overview.cpu_usage === 0 && overview.memory_usage === 0 &&
    overview.disk_usage === 0 && overview.node_count === 0 &&
    overview.pod_count === 0 && overview.network_in === 0 && overview.network_out === 0
})

// 数据源外部页面地址（根据类型生成不同的链接）
const prometheusTargetsURL = computed(() => {
  const url = currentDs.value?.url || datasourceInfo.value
  if (!url) return ''
  const base = url.replace(/\/$/, '')
  const type = currentDs.value?.type || 'prometheus'
  switch (type) {
    case 'prometheus': return base + '/targets?search='
    case 'loki': return base + '/ready'
    case 'alertmanager': return base + '/#/alerts'
    case 'victoriametrics': return base + '/targets'
    case 'thanos': return base + '/targets'
    case 'grafana': return base + '/connections/datasources'
    default: return base + '/targets?search='
  }
})

// 外部链接按钮文案
const targetsLinkText = computed(() => {
  const type = currentDs.value?.type || 'prometheus'
  switch (type) {
    case 'prometheus': return '查看 Targets →'
    case 'loki': return '查看 Loki 状态 →'
    case 'alertmanager': return '查看 Alerts →'
    case 'grafana': return '打开 Grafana →'
    default: return '查看 Targets →'
  }
})

// 图表引用
const cpuChartRef = ref(null)
const memChartRef = ref(null)
const diskChartRef = ref(null)
const networkChartRef = ref(null)
let cpuChart, memChart, diskChart, networkChart

// 指标卡片
const metricCards = computed(() => [
  { icon: '💻', label: 'CPU 使用率', value: overview.cpu_usage.toFixed(1) + '%', bg: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
  { icon: '🧠', label: '内存使用率', value: overview.memory_usage.toFixed(1) + '%', bg: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' },
  { icon: '💾', label: '磁盘使用率', value: overview.disk_usage.toFixed(1) + '%', bg: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' },
  { icon: '🖥️', label: '节点数量', value: String(overview.node_count), bg: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' },
  { icon: '📦', label: 'Pod 数量', value: String(overview.pod_count), bg: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)' },
  { icon: '🔔', label: '活跃告警', value: String(overview.alert_count), bg: 'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)' },
  { icon: '📥', label: '入站流量', value: formatBytes(overview.network_in) + '/s', bg: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)' },
  { icon: '📤', label: '出站流量', value: formatBytes(overview.network_out) + '/s', bg: 'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)' },
])

// ===== 工具函数 =====
function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function getUsageColor(val) {
  if (val >= 90) return '#f5222d'
  if (val >= 70) return '#fa8c16'
  if (val >= 50) return '#fadb14'
  return '#52c41a'
}

function formatTime(ts) {
  const d = new Date(ts * 1000)
  return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0')
}

// ===== 图表配置 =====
function buildLineOption(title, series, isPercent = true) {
  const colors = ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de']
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.96)',
      borderColor: '#e5e7eb',
      textStyle: { color: '#374151', fontSize: 12 },
      formatter: (params) => {
        let html = `<div style="font-weight:600;margin-bottom:4px">${params[0]?.axisValueLabel}</div>`
        params.forEach(p => {
          const val = isPercent ? p.value.toFixed(1) + '%' : formatBytes(p.value) + '/s'
          html += `<div style="display:flex;align-items:center;gap:6px">
            <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${p.color}"></span>
            ${p.seriesName}: <b>${val}</b>
          </div>`
        })
        return html
      },
    },
    grid: { top: 12, right: 16, bottom: 28, left: 50 },
    xAxis: {
      type: 'category',
      data: series[0]?.data?.map(p => formatTime(p[0])) || [],
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#9ca3af', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: '#9ca3af',
        fontSize: 11,
        formatter: isPercent ? '{value}%' : (v) => formatBytes(v),
      },
      splitLine: { lineStyle: { color: '#f3f4f6' } },
    },
    series: series.map((s, i) => ({
      name: s.label,
      type: 'line',
      smooth: true,
      symbol: 'none',
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.08 },
      itemStyle: { color: colors[i % colors.length] },
      data: s.data?.map(p => p[1]) || [],
    })),
  }
}

// ===== 数据加载 =====
async function loadOverview() {
  try {
    const res = await getClusterOverview()
    if (res.code === 0 && res.data) {
      Object.assign(overview, res.data)
    }
  } catch (e) {
    console.warn('加载集群总览失败', e)
  }
}

async function loadNodes() {
  try {
    const res = await getNodeMetrics()
    if (res.code === 0) nodes.value = res.data || []
  } catch (e) {
    console.warn('加载节点指标失败', e)
  }
}

async function loadTopPodsData() {
  try {
    const res = await getTopPods(topMetric.value)
    if (res.code === 0) topPods.value = res.data || []
  } catch (e) {
    console.warn('加载 Top Pods 失败', e)
  }
}

async function loadTrends() {
  const resources = ['cpu', 'memory', 'disk', 'network']
  const charts = [cpuChart, memChart, diskChart, networkChart]
  const isPercent = [true, true, true, false]

  for (let i = 0; i < resources.length; i++) {
    try {
      const res = await getResourceTrend(resources[i], trendDuration.value)
      if (res.code === 0 && res.data && charts[i]) {
        const series = res.data.map(td => ({
          label: td.label,
          data: td.points.map(p => [p.timestamp, p.value]),
        }))
        charts[i].setOption(buildLineOption(resources[i], series, isPercent[i]))
      }
    } catch (e) {
      console.warn(`加载 ${resources[i]} 趋势失败`, e)
    }
  }
}

async function loadHealth() {
  try {
    const res = await checkHealth()
    if (res.code === 0) {
      healthy.value = res.data?.prometheus || false
      datasourceInfo.value = res.data?.url || ''
    }
  } catch (e) {
    healthy.value = false
  }
}

async function refreshAll() {
  loading.value = true
  try {
    // 先检查 Prometheus 连接状态
    await loadHealth()
    // 只有连接正常时才拉取监控数据，避免无意义的失败请求
    if (healthy.value) {
      await Promise.all([loadOverview(), loadNodes(), loadTopPodsData(), loadTrends()])
    }
  } finally {
    loading.value = false
  }
}

function switchTopMetric(metric) {
  topMetric.value = metric
  loadTopPodsData()
}

// ===== 初始化图表 =====
function initCharts() {
  if (cpuChartRef.value) cpuChart = echarts.init(cpuChartRef.value)
  if (memChartRef.value) memChart = echarts.init(memChartRef.value)
  if (diskChartRef.value) diskChart = echarts.init(diskChartRef.value)
  if (networkChartRef.value) networkChart = echarts.init(networkChartRef.value)
}

function handleResize() {
  cpuChart?.resize()
  memChart?.resize()
  diskChart?.resize()
  networkChart?.resize()
}

let autoRefreshTimer = null

onMounted(async () => {
  await nextTick()
  initCharts()
  await loadDatasources()
  await refreshAll()
  // 自动刷新（30秒）
  autoRefreshTimer = setInterval(refreshAll, 30000)
  window.addEventListener('resize', handleResize)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  if (autoRefreshTimer) clearInterval(autoRefreshTimer)
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('click', handleClickOutside)
  cpuChart?.dispose()
  memChart?.dispose()
  diskChart?.dispose()
  networkChart?.dispose()
})
</script>

<style scoped>
.monitoring-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

/* 页面标题 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: #9ca3af;
  margin-left: 12px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
/* 数据源切换器 */
.ds-switcher { position: relative; }
.ds-switcher-btn { display: flex; align-items: center; gap: 8px; padding: 7px 14px; background: #fff; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer; font-size: 13px; color: #334155; transition: all 0.2s; }
.ds-switcher-btn:hover { border-color: #4f46e5; box-shadow: 0 2px 8px rgba(79,70,229,0.08); }
.ds-switcher.open .ds-switcher-btn { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.ds-switcher-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.ds-switcher-dot.connected { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.4); }
.ds-switcher-dot.disconnected { background: #ef4444; }
.ds-switcher-dot.unknown { background: #9ca3af; }
.ds-switcher-icon { font-size: 16px; }
.ds-switcher-name { font-weight: 500; max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ds-switcher-arrow { font-size: 12px; color: #94a3b8; transition: transform 0.2s; }
.ds-switcher.open .ds-switcher-arrow { transform: rotate(180deg); }

.ds-switcher-dropdown { position: absolute; top: calc(100% + 8px); right: 0; width: 360px; background: #fff; border: 1px solid #e8ecf0; border-radius: 12px; box-shadow: 0 12px 36px rgba(0,0,0,0.12); z-index: 100; overflow: hidden; animation: fadeDown 0.15s ease-out; }
@keyframes fadeDown { from { opacity: 0; transform: translateY(-6px); } to { opacity: 1; transform: translateY(0); } }
.ds-dropdown-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; border-bottom: 1px solid #f1f5f9; font-size: 13px; font-weight: 600; color: #475569; }
.ds-manage-link { font-size: 12px; color: #4f46e5; text-decoration: none; font-weight: 500; }
.ds-manage-link:hover { text-decoration: underline; }
.ds-dropdown-list { max-height: 320px; overflow-y: auto; padding: 6px; }
.ds-dropdown-item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: 8px; cursor: pointer; transition: all 0.15s; }
.ds-dropdown-item:hover { background: #f8fafc; }
.ds-dropdown-item.active { background: #f0f0ff; border: 1px solid #e0e7ff; }
.ds-dropdown-item.disabled { opacity: 0.45; cursor: not-allowed; }
.ds-item-icon { font-size: 18px; flex-shrink: 0; }
.ds-item-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.ds-item-name { font-size: 13px; font-weight: 500; color: #1e293b; }
.ds-item-url { font-size: 11px; color: #94a3b8; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ds-item-status { font-size: 10px; }
.ds-item-status.connected { color: #10b981; }
.ds-item-status.disconnected { color: #ef4444; }
.ds-item-status.unknown { color: #9ca3af; }
.ds-item-badge { font-size: 10px; background: #4f46e5; color: #fff; padding: 1px 6px; border-radius: 4px; }
.ds-dropdown-empty { padding: 20px; text-align: center; font-size: 13px; color: #94a3b8; }
.ds-dropdown-empty a { color: #4f46e5; }

/* 数据源类型分组 */
.ds-type-group { margin-bottom: 4px; }
.ds-group-title {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px 4px;
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.ds-group-icon { font-size: 13px; }
.ds-group-count {
  font-size: 10px;
  background: #f1f5f9;
  color: #64748b;
  padding: 1px 5px;
  border-radius: 4px;
  margin-left: auto;
}

/* 视图模式标签 */
.view-mode-tag {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s;
}
.view-mode-tag.metrics {
  background: linear-gradient(135deg, #ecfdf5, #d1fae5);
  color: #059669;
  border: 1px solid #a7f3d0;
}
.view-mode-tag.logs {
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
  color: #2563eb;
  border: 1px solid #93c5fd;
}
.view-mode-icon { font-size: 14px; }

.targets-link { font-size: 13px; color: #4f46e5; text-decoration: none; padding: 6px 12px; border: 1px solid #e0e7ff; border-radius: 8px; transition: all 0.2s; font-weight: 500; white-space: nowrap; }
.targets-link:hover { background: #eef2ff; border-color: #4f46e5; }
.refresh-btn {
  padding: 6px 16px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: #fff;
  font-size: 13px;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
}
.refresh-btn:hover { border-color: #3b82f6; color: #3b82f6; }
.refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.spinning { display: inline-block; animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* 未连接提示 */
.connect-hint { margin-bottom: 20px; }
.connect-hint.info-hint .hint-content { background: #eff6ff; border-color: #bfdbfe; }
.connect-hint.info-hint .hint-text b { color: #1e40af; }
.connect-hint.info-hint .hint-text p { color: #3b82f6; }
.connect-hint.info-hint .hint-btn { background: #3b82f6; }
.connect-hint.info-hint .hint-btn:hover { background: #2563eb; }
.connect-hint.info-hint code { background: #e0e7ff; padding: 1px 6px; border-radius: 4px; font-size: 12px; color: #4338ca; }
.hint-content { display: flex; align-items: center; gap: 12px; padding: 14px 20px; background: #fffbeb; border: 1px solid #fde68a; border-radius: 10px; }
.hint-icon { font-size: 20px; }
.hint-text { flex: 1; }
.hint-text b { font-size: 14px; color: #92400e; }
.hint-text p { margin: 2px 0 0; font-size: 13px; color: #a16207; }
.hint-btn { padding: 6px 14px; background: #f59e0b; color: #fff; border-radius: 6px; font-size: 13px; text-decoration: none; font-weight: 500; white-space: nowrap; }
.hint-btn:hover { background: #d97706; }

/* 指标卡片 */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.metric-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  transition: transform 0.2s, box-shadow 0.2s;
}
.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.metric-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}
.metric-info {
  display: flex;
  flex-direction: column;
}
.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.2;
}
.metric-label {
  font-size: 13px;
  color: #9ca3af;
  margin-top: 2px;
}

/* 图表卡片 */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.chart-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.chart-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  margin: 0;
}
.duration-select {
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 12px;
  color: #6b7280;
  background: #fff;
  cursor: pointer;
}
.chart-body {
  width: 100%;
  height: 240px;
}

/* 底部表格区域 */
.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.table-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.table-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  margin: 0;
}
.tab-switch {
  display: flex;
  gap: 0;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  overflow: hidden;
}
.tab-switch button {
  padding: 4px 12px;
  border: none;
  background: #fff;
  font-size: 12px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}
.tab-switch button.active {
  background: #3b82f6;
  color: #fff;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th {
  text-align: left;
  padding: 10px 12px;
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  border-bottom: 1px solid #f3f4f6;
}
.data-table td {
  padding: 10px 12px;
  font-size: 13px;
  color: #374151;
  border-bottom: 1px solid #f9fafb;
}
.data-table tr:hover td {
  background: #f9fafb;
}
.node-name {
  font-weight: 500;
  color: #1f2937;
}
.pod-name {
  font-weight: 500;
  color: #1f2937;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.usage-bar-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.usage-bar-wrap .usage-bar {
  width: 0;
  height: 6px;
  border-radius: 3px;
  transition: width 0.6s ease;
  flex-shrink: 0;
  max-width: 80px;
}
.usage-bar-wrap span {
  font-size: 12px;
  color: #6b7280;
  white-space: nowrap;
}
.status-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}
.status-tag.ready {
  background: #f0fdf4;
  color: #16a34a;
}
.status-tag.not-ready {
  background: #fef2f2;
  color: #dc2626;
}
.ns-tag {
  display: inline-block;
  padding: 2px 8px;
  background: #eff6ff;
  color: #3b82f6;
  border-radius: 4px;
  font-size: 12px;
}
.rank-cell {
  width: 32px;
}
.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  background: #f3f4f6;
  color: #6b7280;
}
.rank-1 { background: #fef3c7; color: #d97706; }
.rank-2 { background: #e5e7eb; color: #4b5563; }
.rank-3 { background: #fce7f3; color: #db2777; }
.usage-value {
  font-family: 'SF Mono', 'Menlo', monospace;
  font-weight: 500;
}
.empty-row {
  text-align: center;
  color: #9ca3af;
  padding: 24px 0 !important;
}

/* 响应式 */
@media (max-width: 1200px) {
  .metrics-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-grid { grid-template-columns: 1fr; }
  .bottom-grid { grid-template-columns: 1fr; }
}
</style>

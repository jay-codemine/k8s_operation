<template>
  <div class="monitoring-page">
    <!-- 【大厂风】监控视野工作台栏 -->
    <div class="vista-bar">
      <div class="vista-bar-bg"></div>
      <div class="vista-content">
        <!-- 左：选择器组（集群下拉 + 数据源下拉） -->
        <div class="vista-section vista-selectors">
          <!-- 集群下拉 -->
          <div class="vsel" :class="{ open: clusterSelOpen }" @click.stop="toggleClusterSel">
            <div class="vsel-label"><span class="vsl-icon">☸️</span>监控集群</div>
            <button class="vsel-btn">
              <span class="vsel-icon">{{ vistaClusterId === 0 ? 'language' : '☸️' }}</span>
              <span class="vsel-name">{{ currentVistaClusterName }}</span>
              <span class="vsel-meta">{{ vistaClusterId === 0 ? globalDsCount : clusterDsCount(vistaClusterId) }} 个数据源</span>
              <span class="vsel-arrow">▾</span>
            </button>
            <div class="vsel-dropdown" v-if="clusterSelOpen" @click.stop>
              <div class="vsel-dd-header">
                <span>选择监控集群</span>
                <router-link to="/platform/cluster" class="vsel-link" @click="clusterSelOpen = false">管理 →</router-link>
              </div>
              <div
                class="vsel-item"
                :class="{ active: vistaClusterId === 0 }"
                @click="switchVistaCluster(0)"
              >
                <span class="vsel-it-icon">🌐</span>
                <div class="vsel-it-info">
                  <span class="vsel-it-name">全部集群</span>
                  <span class="vsel-it-sub">查看全部数据源（含全局共享）</span>
                </div>
                <span class="vsel-it-badge">{{ globalDsCount }}</span>
              </div>
              <div
                v-for="c in vistaClusters"
                :key="c.id"
                class="vsel-item"
                :class="{ active: vistaClusterId === Number(c.id) }"
                @click="switchVistaCluster(Number(c.id))"
              >
                <span class="vsel-status-dot" :class="clusterDsHealth(Number(c.id))"></span>
                <span class="vsel-it-icon">☸️</span>
                <div class="vsel-it-info">
                  <span class="vsel-it-name">{{ c.cluster_name }}</span>
                  <span class="vsel-it-sub">{{ c.api_server || '集群 #' + c.id }}</span>
                </div>
                <span class="vsel-it-badge">{{ clusterDsCount(Number(c.id)) }}</span>
              </div>
              <div v-if="!vistaClusters.length" class="vsel-empty">未配置 K8s 集群，<router-link to="/platform/cluster">去添加→</router-link></div>
            </div>
          </div>

          <!-- 数据源下拉 -->
          <div class="vsel" :class="{ open: dsSelOpen }" @click.stop="toggleDsSel">
            <div class="vsel-label"><span class="vsl-icon">📡</span>数据源</div>
            <button class="vsel-btn" :class="{ 'is-empty': !scopedDsList.length }">
              <span class="vsel-status-dot" :class="currentDs.status || 'unknown'"></span>
              <span class="vsel-icon">{{ getDsIcon(currentDs.type) }}</span>
              <span class="vsel-name">{{ currentDs.name || (scopedDsList.length ? '请选择数据源' : '该集群暂无数据源') }}</span>
              <span class="vsel-meta" v-if="currentDs.url">{{ currentDs.url }}</span>
              <span class="vsel-arrow">▾</span>
            </button>
            <div class="vsel-dropdown" v-if="dsSelOpen" @click.stop>
              <div class="vsel-dd-header">
                <span>切换数据源（当前集群范围）</span>
                <router-link to="/monitoring/datasources" class="vsel-link" @click="dsSelOpen = false">管理 →</router-link>
              </div>
              <template v-if="dsGroups.length">
                <div class="vsel-group" v-for="group in dsGroups" :key="group.type">
                  <div class="vsel-group-title">
                    <span class="vsel-it-icon">{{ getDsIcon(group.type) }}</span>
                    <span>{{ getDsTypeName(group.type) }}</span>
                    <span class="vsel-group-count">{{ group.items.length }}</span>
                  </div>
                  <div
                    class="vsel-item"
                    :class="{ active: ds.id === currentDs.id, disabled: !ds.enabled }"
                    v-for="ds in group.items"
                    :key="ds.id"
                    @click="switchDatasource(ds)"
                  >
                    <span class="vsel-status-dot" :class="ds.status || 'unknown'"></span>
                    <span class="vsel-it-icon">{{ getDsIcon(ds.type) }}</span>
                    <div class="vsel-it-info">
                      <span class="vsel-it-name">
                        {{ ds.name }}
                        <span class="vsel-default-tag" v-if="ds.is_default">默认</span>
                        <span class="vsel-shared-tag" v-if="Number(ds.cluster_id || 0) === 0">全局共享</span>
                      </span>
                      <span class="vsel-it-sub">{{ ds.url }}</span>
                    </div>
                  </div>
                </div>
              </template>
              <div v-else class="vsel-empty">
                当前集群暂无可用数据源，<router-link to="/monitoring/datasources">去添加</router-link>
              </div>
            </div>
          </div>
        </div>

        <!-- 右：状态概述 + 状态指示灯 -->
        <div class="vista-stats">
          <div class="vstat">
            <span class="vstat-icon">📡</span>
            <div class="vstat-meta">
              <span class="vstat-num">{{ scopedDsList.length }}</span>
              <span class="vstat-label">可用数据源</span>
            </div>
          </div>
          <div class="vstat">
            <span class="vstat-dot connected"></span>
            <div class="vstat-meta">
              <span class="vstat-num">{{ scopedConnectedCount }}</span>
              <span class="vstat-label">已连接</span>
            </div>
          </div>
          <div class="vstat">
            <span class="vstat-dot disconnected"></span>
            <div class="vstat-meta">
              <span class="vstat-num">{{ scopedDisconnectedCount }}</span>
              <span class="vstat-label">异常</span>
            </div>
          </div>
          <router-link to="/monitoring/datasources" class="vista-config-btn">
            <span>⚙️</span>配置
          </router-link>
        </div>
      </div>
    </div>

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
          <span class="view-mode-icon">{{ currentViewMode === 'metrics' ? '📊' : '📜' }}</span>
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
          <b v-if="noDedicatedPrometheus">「{{ currentVistaClusterName }}」暂无专属 Prometheus 数据源</b>
          <b v-else>{{ currentDs.type === 'loki' ? 'Loki' : 'Prometheus' }} 未连接</b>
          <p v-if="noDedicatedPrometheus">
            请前往「数据源管理」为该集群添加并绑定 Prometheus 数据源，确保集群监控数据隔离。
          </p>
          <p v-else>请前往「数据源管理」配置数据源地址，或检查网络连通性</p>
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
    <!-- 健康评分 + KPI 卡片 -->
    <div class="hero-grid">
      <div class="hero-score-card" :class="'score-' + (healthScore.level || 'good')">
        <div class="score-ring">
          <svg viewBox="0 0 120 120" class="ring-svg">
            <circle cx="60" cy="60" r="52" class="ring-bg"></circle>
            <circle cx="60" cy="60" r="52" class="ring-fg"
              :style="{ strokeDasharray: 326.7, strokeDashoffset: 326.7 * (1 - (healthScore.score || 0) / 100) }"></circle>
          </svg>
          <div class="ring-text">
            <div class="ring-score">{{ healthScore.score ?? '--' }}</div>
            <div class="ring-label">集群健康分</div>
          </div>
        </div>
        <div class="score-info">
          <div class="score-level">
            <span class="level-badge" :class="'level-' + (healthScore.level || 'good')">{{ scoreLevelText }}</span>
          </div>
          <ul class="score-suggestions">
            <li v-for="(tip, i) in (healthScore.suggestions || []).slice(0, 3)" :key="i">{{ tip }}</li>
          </ul>
        </div>
      </div>
      <div class="kpi-grid">
        <div class="kpi-card" v-for="m in metricCards" :key="m.label">
          <div class="kpi-head">
            <span class="kpi-icon" :style="{ color: m.color }">{{ m.icon }}</span>
            <span class="kpi-label">{{ m.label }}</span>
          </div>
          <div class="kpi-value" :style="{ color: m.color }">{{ m.value }}</div>
          <div class="kpi-sub">{{ m.sub }}</div>
        </div>
      </div>
    </div>

    <!-- 趋势图表区域 -->
    <div class="charts-grid">
      <div class="chart-card" :class="{ 'is-loading': chartsLoading }">
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
        <div class="chart-loading-mask" v-if="chartsLoading"><span class="chart-spinner"></span></div>
      </div>
      <div class="chart-card" :class="{ 'is-loading': chartsLoading }">
        <div class="chart-header">
          <h3>内存使用率趋势</h3>
        </div>
        <div ref="memChartRef" class="chart-body"></div>
        <div class="chart-loading-mask" v-if="chartsLoading"><span class="chart-spinner"></span></div>
      </div>
      <div class="chart-card" :class="{ 'is-loading': chartsLoading }">
        <div class="chart-header">
          <h3>磁盘使用率趋势</h3>
        </div>
        <div ref="diskChartRef" class="chart-body"></div>
        <div class="chart-loading-mask" v-if="chartsLoading"><span class="chart-spinner"></span></div>
      </div>
      <div class="chart-card" :class="{ 'is-loading': chartsLoading }">
        <div class="chart-header">
          <h3>网络流量趋势</h3>
        </div>
        <div ref="networkChartRef" class="chart-body"></div>
        <div class="chart-loading-mask" v-if="chartsLoading"><span class="chart-spinner"></span></div>
      </div>
    </div>

    <!-- 热力图 + Pod 状态饱和度 -->
    <div class="viz-grid">
      <div class="chart-card heatmap-card">
        <div class="chart-header">
          <h3>节点负载热力图</h3>
          <div class="tab-switch">
            <button :class="{ active: heatmapMetric === 'cpu' }" @click="switchHeatmap('cpu')">CPU</button>
            <button :class="{ active: heatmapMetric === 'memory' }" @click="switchHeatmap('memory')">内存</button>
          </div>
        </div>
        <div ref="heatmapRef" class="chart-body" style="height:220px"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>Pod 状态分布</h3>
          <span class="chart-sub">实时 Phase 聚合</span>
        </div>
        <div ref="podDonutRef" class="chart-body" style="height:220px"></div>
      </div>
    </div>

    <!-- 三视角 Tab：主机 / Pod / Namespace -->
    <div class="scope-card">
      <div class="scope-tabs">
        <button v-for="t in scopeTabs" :key="t.key"
                :class="{ active: scopeView === t.key }" @click="switchScope(t.key)">
          <span class="scope-icon">{{ t.icon }}</span>{{ t.label }}
        </button>
        <div class="scope-toolbar">
          <select v-if="scopeView === 'pods'" v-model="topMetric" @change="loadTopPodsData" class="duration-select">
            <option value="cpu">按 CPU</option>
            <option value="memory">按 内存</option>
          </select>
        </div>
      </div>

      <!-- 主机视角 -->
      <table v-if="scopeView === 'nodes'" class="data-table">
        <thead>
          <tr>
            <th>节点</th>
            <th>CPU</th>
            <th>内存</th>
            <th>磁盘</th>
            <th>Load1</th>
            <th>Pod 数</th>
            <th>入/出网络</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="node in nodes" :key="node.name">
            <td class="node-name">{{ node.name }}</td>
            <td><div class="usage-bar-wrap"><div class="usage-bar" :style="{ width: node.cpu_usage + '%', background: getUsageColor(node.cpu_usage) }"></div><span>{{ node.cpu_usage.toFixed(1) }}%</span></div></td>
            <td><div class="usage-bar-wrap"><div class="usage-bar" :style="{ width: node.memory_usage + '%', background: getUsageColor(node.memory_usage) }"></div><span>{{ node.memory_usage.toFixed(1) }}%</span></div></td>
            <td><div class="usage-bar-wrap"><div class="usage-bar" :style="{ width: node.disk_usage + '%', background: getUsageColor(node.disk_usage) }"></div><span>{{ node.disk_usage.toFixed(1) }}%</span></div></td>
            <td class="mono">{{ (node.load1 || 0).toFixed(2) }}</td>
            <td class="mono">{{ node.pod_count || '-' }}</td>
            <td class="mono">{{ formatBytes(node.network_in || 0) }}/s ↑ {{ formatBytes(node.network_out || 0) }}/s</td>
            <td><span class="status-tag" :class="node.status === 'Ready' ? 'ready' : 'not-ready'">{{ node.status }}</span></td>
            <td><button class="detail-btn" @click="openNodeDetail(node)" title="查看节点详情">📊 详情</button></td>
          </tr>
          <tr v-if="nodes.length === 0"><td colspan="9" class="empty-row">暂无节点数据</td></tr>
        </tbody>
      </table>

      <!-- Pod 视角 -->
      <div v-else-if="scopeView === 'pods'" class="pod-view">
        <div class="pod-section">
          <h4 class="sub-title">资源占用 Top 10 ({{ topMetric === 'cpu' ? 'CPU' : '内存' }})</h4>
          <table class="data-table">
            <thead>
              <tr><th>#</th><th>Pod</th><th>命名空间</th><th>{{ topMetric === 'cpu' ? 'CPU (cores)' : '内存 (MB)' }}</th></tr>
            </thead>
            <tbody>
              <tr v-for="(pod, i) in topPods" :key="pod.name">
                <td class="rank-cell"><span class="rank-badge" :class="'rank-' + (i + 1)">{{ i + 1 }}</span></td>
                <td class="pod-name">{{ pod.name }}</td>
                <td><span class="ns-tag">{{ pod.namespace }}</span></td>
                <td class="usage-value">{{ topMetric === 'cpu' ? pod.cpu_usage.toFixed(3) : (pod.memory_usage / 1048576).toFixed(1) }}</td>
              </tr>
              <tr v-if="topPods.length === 0"><td colspan="4" class="empty-row">暂无数据</td></tr>
            </tbody>
          </table>
        </div>
        <div class="pod-section">
          <h4 class="sub-title">重启异常 Pod (restart > 0)</h4>
          <table class="data-table">
            <thead>
              <tr><th>Pod</th><th>命名空间</th><th>容器</th><th>重启次数</th></tr>
            </thead>
            <tbody>
              <tr v-for="p in abnormalPods" :key="p.namespace + '/' + p.name + '/' + p.container">
                <td class="pod-name">{{ p.name }}</td>
                <td><span class="ns-tag">{{ p.namespace }}</span></td>
                <td class="mono">{{ p.container }}</td>
                <td><span class="restart-badge" :class="{ hot: p.restarts >= 5 }">{{ p.restarts }}</span></td>
              </tr>
              <tr v-if="abnormalPods.length === 0"><td colspan="4" class="empty-row">所有 Pod 运行正常，无重启异常</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Namespace 视角 -->
      <table v-else class="data-table">
        <thead>
          <tr><th>命名空间</th><th>CPU (cores)</th><th>内存 (MB)</th><th>Pod 数</th></tr>
        </thead>
        <tbody>
          <tr v-for="ns in namespaceMetrics" :key="ns.namespace">
            <td><span class="ns-tag">{{ ns.namespace }}</span></td>
            <td class="usage-value">{{ (ns.cpu_usage || 0).toFixed(3) }}</td>
            <td class="usage-value">{{ ((ns.memory_usage || 0) / 1048576).toFixed(1) }}</td>
            <td class="mono">{{ ns.pod_count || 0 }}</td>
          </tr>
          <tr v-if="namespaceMetrics.length === 0"><td colspan="4" class="empty-row">暂无数据</td></tr>
        </tbody>
      </table>
    </div>
    </template>

    <!-- 节点详情抽屉 -->
    <NodeDetailDrawer v-model:visible="detailVisible" :instance="detailInstance" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, onActivated, nextTick, computed } from 'vue'
import * as echarts from 'echarts'
import LokiView from './LokiView.vue'
import NodeDetailDrawer from './NodeDetailDrawer.vue'

// 组件名，配合 MonitorLayout 的 keep-alive include
defineOptions({ name: 'MonitoringOverview' })
import {
  getClusterOverview,
  getNodeMetrics,
  getResourceTrend,
  getTopPods,
  checkHealth,
  listDatasources,
  getHealthScore,
  getNodeHeatmap,
  getPodStatusDistribution,
  getAbnormalPods,
  getNamespaceMetrics,
} from '@/api/monitoring'
import { getK8sClusterList } from '@/api/platform/cluster.js'

// 节点详情抽屉
const detailVisible = ref(false)
const detailInstance = ref('')
function openNodeDetail(node) {
  // 节点表中 node.name 即 instance（如 192.168.124.10:9100）
  detailInstance.value = node.name
  detailVisible.value = true
}

// ===== 状态 =====
const loading = ref(true)
const chartsLoading = ref(false)
const healthy = ref(false)
const datasourceInfo = ref('')

// 数据源切换
const dsSwitcherOpen = ref(false)
const dsList = ref([])
const currentDs = ref({})

// 【大厂风下拉】集群 + 数据源选择器状态
const clusterSelOpen = ref(false)
const dsSelOpen = ref(false)
function toggleClusterSel() {
  clusterSelOpen.value = !clusterSelOpen.value
  if (clusterSelOpen.value) dsSelOpen.value = false
}
function toggleDsSel() {
  // 即使当前集群暂无可用数据源也允许展开，让用户看到「暂无可用数据源」提示和管理入口
  dsSelOpen.value = !dsSelOpen.value
  if (dsSelOpen.value) clusterSelOpen.value = false
}

// 当前请求需携带的 datasource_id 参数
const dsQuery = computed(() => currentDs.value?.id ? { datasource_id: currentDs.value.id } : {})

// 【监控视野】集群选择
const VISTA_LS_KEY = 'monitoring.vista.cluster_id'
const vistaClusters = ref([])
const vistaClusterId = ref(Number(localStorage.getItem(VISTA_LS_KEY) || 0))

async function loadVistaClusters() {
  try {
    const res = await getK8sClusterList({ page: 1, limit: 200 })
    vistaClusters.value = res?.data?.list || []
  } catch { vistaClusters.value = [] }
}

const currentVistaClusterName = computed(() => {
  if (vistaClusterId.value === 0) return '全部集群'
  const c = vistaClusters.value.find(x => Number(x.id) === vistaClusterId.value)
  return c?.cluster_name || ('集群 #' + vistaClusterId.value)
})

// 根据当前视野过滤出的数据源列表
// 规则：选中集群后优先展示该集群专属数据源；仅当集群无专属时才 fallback 全局共享(cluster_id=0)
const scopedDsList = computed(() => {
  if (vistaClusterId.value === 0) return dsList.value
  const cid = vistaClusterId.value
  const dedicated = dsList.value.filter(d => Number(d.cluster_id || 0) === cid)
  // 有专属数据源时只展示专属的，避免跨集群混用
  if (dedicated.length > 0) return dedicated
  // 完全没有专属时，fallback 全局共享
  return dsList.value.filter(d => Number(d.cluster_id || 0) === 0)
})
const scopedConnectedCount = computed(() => scopedDsList.value.filter(d => d.status === 'connected').length)
const scopedDisconnectedCount = computed(() => scopedDsList.value.filter(d => d.status === 'disconnected').length)
const globalDsCount = computed(() => dsList.value.length)

function clusterDsCount(cid) {
  // 与 scopedDsList 保持一致：有专属显示专属数量，无专属显示全局共享数量
  const dedicated = dsList.value.filter(d => Number(d.cluster_id || 0) === cid)
  if (dedicated.length > 0) return dedicated.length
  return dsList.value.filter(d => Number(d.cluster_id || 0) === 0).length
}
function clusterDsHealth(cid) {
  const items = dsList.value.filter(d => Number(d.cluster_id || 0) === cid && d.enabled)
  if (!items.length) return 'unknown'
  if (items.every(d => d.status === 'connected')) return 'connected'
  if (items.some(d => d.status === 'disconnected')) return 'disconnected'
  return 'unknown'
}

function switchVistaCluster(cid) {
  vistaClusterId.value = cid
  localStorage.setItem(VISTA_LS_KEY, String(cid))
  clusterSelOpen.value = false
  // 自动按优先级选中数据源：集群专属 > 全局共享(cluster_id=0)
  const candidates = scopedDsList.value
  const isDedicated = (d) => cid > 0 && Number(d.cluster_id || 0) === cid
  const pickList = [
    // P0: 集群专属 + prometheus + connected
    candidates.find(d => isDedicated(d) && d.type === 'prometheus' && d.status === 'connected' && d.enabled),
    // P1: 集群专属 + prometheus
    candidates.find(d => isDedicated(d) && d.type === 'prometheus' && d.enabled),
    // P2: 集群专属 + 任何 connected
    candidates.find(d => isDedicated(d) && d.status === 'connected' && d.enabled),
    // P3: fallback 全局共享 prometheus + connected
    candidates.find(d => d.type === 'prometheus' && d.status === 'connected' && d.enabled),
    // P4: fallback 全局共享 prometheus
    candidates.find(d => d.type === 'prometheus' && d.enabled),
    // P5: 任何 connected
    candidates.find(d => d.status === 'connected' && d.enabled),
    // P6: 任何 enabled
    candidates.find(d => d.enabled),
    candidates[0],
  ]
  const next = pickList.find(Boolean)
  if (next) {
    currentDs.value = next
    datasourceInfo.value = next.url
    healthy.value = next.status === 'connected'
  } else {
    currentDs.value = {}
    datasourceInfo.value = ''
    healthy.value = false
  }
  // 无论有无数据源都重新刷新，让 chart 、卡片提示、健康分需要准确响应当前选择
  refreshAll()
}

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

// 数据源按类型分组（仅在当前视野范围内）
const dsGroups = computed(() => {
  const groups = {}
  for (const ds of scopedDsList.value) {
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
  const map = { prometheus: '🔥', loki: '📜', alertmanager: '🚨', victoriametrics: '📊', grafana: '📈', n9e: '🦉', thanos: '♾️' }
  return map[type] || '🔌'
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
      // 默认选中：在当前视野范围内 优先默认数据源 > 首个已启用
      const scope = scopedDsList.value
      if (!currentDs.value.id && scope.length) {
        const def = scope.find(d => d.is_default && d.enabled)
        const first = scope.find(d => d.enabled)
        currentDs.value = def || first || scope[0]
      } else if (currentDs.value.id) {
        // 如果当前选中的数据源不在当前视野范围内，重新挑一个
        const stillIn = scope.find(d => d.id === currentDs.value.id)
        if (!stillIn && scope.length) {
          const def = scope.find(d => d.is_default && d.enabled)
          const first = scope.find(d => d.enabled)
          currentDs.value = def || first || scope[0]
        }
      }
    }
  } catch {}
}

function switchDatasource(ds) {
  if (!ds.enabled) return
  currentDs.value = ds
  dsSwitcherOpen.value = false
  dsSelOpen.value = false
  // 更新连接信息并刷新
  datasourceInfo.value = ds.url
  healthy.value = ds.status === 'connected'
  refreshAll()
}

// 点击外部关闭下拉
function handleClickOutside(e) {
  if (!e.target.closest('.ds-switcher')) dsSwitcherOpen.value = false
  if (!e.target.closest('.vsel')) {
    clusterSelOpen.value = false
    dsSelOpen.value = false
  }
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

// 新增：健康评分 + 热力图 + Pod 状态 + 异常 Pod + Namespace、三视角 Tab
const healthScore = ref({ score: 0, level: 'good', factors: {}, suggestions: [] })
const heatmapMetric = ref('cpu')
const heatmapData = ref([])
const podStatusData = ref([])
const abnormalPods = ref([])
const namespaceMetrics = ref([])
const scopeView = ref('nodes')
const scopeTabs = [
  { key: 'nodes', label: '主机视角', icon: '🖥️' },
  { key: 'pods', label: 'Pod 视角', icon: '📦' },
  { key: 'namespaces', label: 'Namespace 视角', icon: '🗂️' },
]
const scoreLevelText = computed(() => {
  const map = { excellent: '优秀', good: '良好', warning: '需关注', critical: '严重' }
  return map[healthScore.value.level] || '良好'
})

// 计算是否所有指标都为0（已连接但无数据）
const isAllZero = computed(() => {
  return overview.cpu_usage === 0 && overview.memory_usage === 0 &&
    overview.disk_usage === 0 && overview.node_count === 0 &&
    overview.pod_count === 0 && overview.network_in === 0 && overview.network_out === 0
})

// 当前集群是否缺少专属 Prometheus 数据源（用于精确提示）
const noDedicatedPrometheus = computed(() => {
  if (vistaClusterId.value === 0) return false
  return dsList.value.filter(
    d => Number(d.cluster_id || 0) === vistaClusterId.value &&
         ['prometheus', 'victoriametrics', 'thanos'].includes(d.type) &&
         d.enabled
  ).length === 0
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
const heatmapRef = ref(null)
const podDonutRef = ref(null)
let cpuChart, memChart, diskChart, networkChart, heatmapChart, podDonutChart

// 指标卡片（大厂风格的 KPI 网格，使用色彩 + 副文本）
const metricCards = computed(() => [
  { icon: '💻', label: 'CPU 使用率',  value: overview.cpu_usage.toFixed(1) + '%',    sub: '所有节点平均',         color: '#4f46e5' },
  { icon: '🧠', label: '内存使用率', value: overview.memory_usage.toFixed(1) + '%', sub: '所有节点平均',         color: '#ec4899' },
  { icon: '💾', label: '磁盘使用率', value: overview.disk_usage.toFixed(1) + '%',   sub: '根分区最大值',         color: '#0ea5e9' },
  { icon: '🖥️', label: '节点数量',   value: String(overview.node_count),            sub: 'Ready 状态',          color: '#10b981' },
  { icon: '📦', label: 'Pod 数量',    value: String(overview.pod_count),             sub: 'Running 总数',        color: '#f59e0b' },
  { icon: '🔔', label: '活跃告警',   value: String(overview.alert_count),           sub: 'Firing/Pending',  color: '#ef4444' },
  { icon: '📥', label: '入站流量',   value: formatBytes(overview.network_in) + '/s',  sub: '集群聚合',           color: '#8b5cf6' },
  { icon: '📤', label: '出站流量',   value: formatBytes(overview.network_out) + '/s', sub: '集群聚合',           color: '#f97316' },
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
    const res = await getClusterOverview(dsQuery.value)
    if (res.code === 0 && res.data) {
      Object.assign(overview, res.data)
    }
  } catch (e) {
    console.warn('加载集群总览失败', e)
  }
}

async function loadNodes() {
  try {
    const res = await getNodeMetrics(dsQuery.value)
    if (res.code === 0) nodes.value = res.data || []
  } catch (e) {
    console.warn('加载节点指标失败', e)
  }
}

async function loadTopPodsData() {
  try {
    const res = await getTopPods(topMetric.value, dsQuery.value)
    if (res.code === 0) topPods.value = res.data || []
  } catch (e) {
    console.warn('加载 Top Pods 失败', e)
  }
}

async function loadTrends() {
  const resources = ['cpu', 'memory', 'disk', 'network']
  const isPercent = [true, true, true, false]

  // 确保 chart 实例存在（首次进入时容器可能尚未有尺寸，延迟重试 init）
  if (!cpuChart && cpuChartRef.value) cpuChart = echarts.init(cpuChartRef.value)
  if (!memChart && memChartRef.value) memChart = echarts.init(memChartRef.value)
  if (!diskChart && diskChartRef.value) diskChart = echarts.init(diskChartRef.value)
  if (!networkChart && networkChartRef.value) networkChart = echarts.init(networkChartRef.value)

  const charts = [cpuChart, memChart, diskChart, networkChart]

  chartsLoading.value = true
  try {
    // 并行加载所有趋势数据，提升加载速度
    const results = await Promise.allSettled(
      resources.map(r => getResourceTrend(r, trendDuration.value, dsQuery.value))
    )

    results.forEach((result, i) => {
      if (result.status === 'fulfilled' && result.value?.code === 0 && result.value.data && charts[i]) {
        const series = result.value.data.map(td => ({
          label: td.label,
          data: td.points.map(p => [p.timestamp, p.value]),
        }))
        charts[i].setOption(buildLineOption(resources[i], series, isPercent[i]))
      } else if (result.status === 'rejected') {
        console.warn(`加载 ${resources[i]} 趋势失败`, result.reason)
      }
    })
  } finally {
    chartsLoading.value = false
  }
}

async function loadHealth() {
  try {
    const res = await checkHealth(dsQuery.value)
    if (res.code === 0) {
      healthy.value = res.data?.prometheus || false
      datasourceInfo.value = res.data?.url || ''
    }
  } catch (e) {
    healthy.value = false
  }
}

// 集群健康评分
async function loadHealthScore() {
  try {
    const res = await getHealthScore(dsQuery.value)
    if (res.code === 0 && res.data) healthScore.value = res.data
  } catch (e) { console.warn('加载健康评分失败', e) }
}

// 节点热力图
async function loadHeatmap() {
  try {
    const res = await getNodeHeatmap(heatmapMetric.value, '1h', dsQuery.value)
    if (res.code === 0) {
      heatmapData.value = res.data || []
      renderHeatmap()
    }
  } catch (e) { console.warn('加载热力图失败', e) }
}

// Pod 状态分布
async function loadPodStatus() {
  try {
    const res = await getPodStatusDistribution(dsQuery.value)
    if (res.code === 0) {
      podStatusData.value = res.data || []
      renderPodDonut()
    }
  } catch (e) { console.warn('加载 Pod 状态分布失败', e) }
}

// 重启异常 Pod
async function loadAbnormalPods() {
  try {
    const res = await getAbnormalPods(dsQuery.value)
    if (res.code === 0) abnormalPods.value = res.data || []
  } catch (e) { console.warn('加载异常 Pod 失败', e) }
}

// Namespace 分组指标
async function loadNamespaceMetrics() {
  try {
    const res = await getNamespaceMetrics(dsQuery.value)
    if (res.code === 0) namespaceMetrics.value = res.data || []
  } catch (e) { console.warn('加载 Namespace 指标失败', e) }
}

async function refreshAll() {
  loading.value = true
  try {
    // 先检查 Prometheus 连接状态
    await loadHealth()
    // 只有连接正常时才拉取监控数据，避免无意义的失败请求
    if (healthy.value) {
      await Promise.all([
        loadOverview(),
        loadNodes(),
        loadTopPodsData(),
        loadTrends(),
        loadHealthScore(),
        loadHeatmap(),
        loadPodStatus(),
        loadAbnormalPods(),
        loadNamespaceMetrics(),
      ])
      // 数据加载完成后强制 resize，修复首次进入时容器尺寸为 0 的问题
      await nextTick()
      handleResize()
    }
  } finally {
    loading.value = false
  }
}

function switchTopMetric(metric) {
  topMetric.value = metric
  loadTopPodsData()
}

function switchHeatmap(metric) {
  heatmapMetric.value = metric
  loadHeatmap()
}

function switchScope(view) {
  scopeView.value = view
}

// ===== 初始化图表 =====
function initCharts() {
  if (cpuChartRef.value) cpuChart = echarts.init(cpuChartRef.value)
  if (memChartRef.value) memChart = echarts.init(memChartRef.value)
  if (diskChartRef.value) diskChart = echarts.init(diskChartRef.value)
  if (networkChartRef.value) networkChart = echarts.init(networkChartRef.value)
  if (heatmapRef.value) heatmapChart = echarts.init(heatmapRef.value)
  if (podDonutRef.value) podDonutChart = echarts.init(podDonutRef.value)
}

// 节点热力图渲染
function renderHeatmap() {
  if (!heatmapChart && heatmapRef.value) heatmapChart = echarts.init(heatmapRef.value)
  if (!heatmapChart) return
  const cells = heatmapData.value || []
  if (!cells.length) { heatmapChart.clear(); return }
  const nodeSet = []
  const tsSet = []
  cells.forEach(c => {
    if (!nodeSet.includes(c.node)) nodeSet.push(c.node)
    if (!tsSet.includes(c.timestamp)) tsSet.push(c.timestamp)
  })
  tsSet.sort((a, b) => a - b)
  const data = cells.map(c => [tsSet.indexOf(c.timestamp), nodeSet.indexOf(c.node), Number(c.value.toFixed(2))])
  heatmapChart.setOption({
    tooltip: { position: 'top', formatter: (p) => `${nodeSet[p.value[1]]}<br/>${formatTime(tsSet[p.value[0]])}<br/><b>${p.value[2]}%</b>` },
    grid: { top: 10, left: 110, right: 20, bottom: 30 },
    xAxis: { type: 'category', data: tsSet.map(formatTime), axisLabel: { color: '#9ca3af', fontSize: 10 }, splitArea: { show: true } },
    yAxis: { type: 'category', data: nodeSet, axisLabel: { color: '#374151', fontSize: 11 }, splitArea: { show: true } },
    visualMap: { min: 0, max: 100, calculable: false, orient: 'horizontal', left: 'center', bottom: 0, itemHeight: 60, itemWidth: 12, textStyle: { color: '#6b7280', fontSize: 10 }, inRange: { color: ['#dbeafe', '#60a5fa', '#1e40af', '#dc2626'] } },
    series: [{ name: heatmapMetric.value, type: 'heatmap', data, label: { show: false }, emphasis: { itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0,0,0,0.3)' } } }],
  })
}

// Pod 状态饼图渲染
function renderPodDonut() {
  if (!podDonutChart && podDonutRef.value) podDonutChart = echarts.init(podDonutRef.value)
  if (!podDonutChart) return
  const items = podStatusData.value || []
  const colorMap = { Running: '#10b981', Pending: '#f59e0b', Succeeded: '#0ea5e9', Failed: '#ef4444', Unknown: '#9ca3af' }
  const data = items.map(i => ({ name: i.phase, value: i.count, itemStyle: { color: colorMap[i.phase] || '#a78bfa' } }))
  podDonutChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: <b>{c}</b> ({d}%)' },
    legend: { bottom: 0, textStyle: { color: '#6b7280', fontSize: 12 } },
    series: [{
      name: 'Pod 状态', type: 'pie', radius: ['45%', '70%'], center: ['50%', '45%'],
      avoidLabelOverlap: false,
      label: { show: true, position: 'outer', formatter: '{b}\n{c}', color: '#374151', fontSize: 11 },
      labelLine: { show: true, length: 8, length2: 8 },
      data,
    }],
  })
}

function handleResize() {
  cpuChart?.resize()
  memChart?.resize()
  diskChart?.resize()
  networkChart?.resize()
  heatmapChart?.resize()
  podDonutChart?.resize()
}

// 防抖版 resize，避免窗口拖动时高频触发
let resizeTimer = null
function debouncedResize() {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(handleResize, 150)
}

let autoRefreshTimer = null

onMounted(async () => {
  await nextTick()
  initCharts()
  // 并行加载集群列表和数据源（二者互不依赖）
  await Promise.all([loadVistaClusters(), loadDatasources()])
  await refreshAll()
  // 延迟二次 resize，确保父布局完全渲染后图表尺寸正确
  setTimeout(handleResize, 300)
  // 自动刷新（30秒）
  autoRefreshTimer = setInterval(refreshAll, 30000)
  window.addEventListener('resize', debouncedResize)
  document.addEventListener('click', handleClickOutside)
})

// keep-alive 激活时重新 resize 图表（从其他 Tab 返回时）
onActivated(() => {
  nextTick(() => handleResize())
})

onUnmounted(() => {
  if (autoRefreshTimer) clearInterval(autoRefreshTimer)
  window.removeEventListener('resize', debouncedResize)
  document.removeEventListener('click', handleClickOutside)
  if (resizeTimer) clearTimeout(resizeTimer)
  cpuChart?.dispose()
  memChart?.dispose()
  diskChart?.dispose()
  networkChart?.dispose()
  heatmapChart?.dispose()
  podDonutChart?.dispose()
})
</script>

<style scoped>
.monitoring-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

/* ============ 【大厂风】监控视野工作台栏 ============ */
.vista-bar {
  position: relative;
  margin: -8px -8px 20px;
  padding: 18px 22px;
  border-radius: 16px;
  /* 注意：这里不能用 overflow: hidden，否则会把 .vsel-dropdown 下拉裁掉 */
  overflow: visible;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #1e1b4b 100%);
  box-shadow: 0 8px 32px rgba(15, 23, 42, 0.18), inset 0 1px 0 rgba(255,255,255,0.04);
  border: 1px solid rgba(99, 102, 241, 0.18);
}
.vista-bar-bg {
  position: absolute; inset: 0;
  /* 与父容器同样的圆角 + overflow:hidden，保证径向渐变不溢出圆角 */
  border-radius: 16px;
  overflow: hidden;
  background:
    radial-gradient(800px 200px at 10% 0%, rgba(79, 70, 229, 0.28), transparent 50%),
    radial-gradient(600px 200px at 90% 100%, rgba(14, 165, 233, 0.22), transparent 50%),
    radial-gradient(400px 150px at 50% 50%, rgba(168, 85, 247, 0.15), transparent 50%);
  pointer-events: none;
}
.vista-content {
  position: relative; z-index: 1;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 24px;
  align-items: center;
}
@media (max-width: 1280px) {
  .vista-content { grid-template-columns: 1fr; gap: 14px; }
}

.vista-section { display: flex; align-items: center; gap: 14px; min-width: 0; }
.vista-section-label {
  display: inline-flex; align-items: center; gap: 6px;
  font-size: 12px; font-weight: 600; color: #c7d2fe;
  text-transform: uppercase; letter-spacing: 0.5px;
  padding: 4px 10px; background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.25); border-radius: 6px;
  white-space: nowrap; flex-shrink: 0;
}
.vsl-icon { font-size: 14px; }

.vista-chips {
  display: flex; gap: 8px; flex-wrap: wrap; align-items: center;
  min-width: 0;
}
.vista-chip {
  display: inline-flex; align-items: center; gap: 7px;
  padding: 7px 14px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 999px; cursor: pointer;
  font-size: 13px; color: #e2e8f0;
  transition: all 0.2s;
  position: relative; overflow: hidden;
  white-space: nowrap;
}
.vista-chip::before {
  content: ''; position: absolute; inset: 0;
  background: linear-gradient(135deg, transparent, rgba(255,255,255,0.05));
  opacity: 0; transition: opacity 0.2s;
}
.vista-chip:hover { background: rgba(255, 255, 255, 0.10); border-color: rgba(99, 102, 241, 0.5); transform: translateY(-1px); }
.vista-chip:hover::before { opacity: 1; }
.vista-chip.active {
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  border-color: rgba(167, 139, 250, 0.6);
  color: #fff;
  box-shadow: 0 4px 16px rgba(79, 70, 229, 0.4), 0 0 0 3px rgba(99, 102, 241, 0.2);
}
.chip-icon { font-size: 14px; }
.chip-text { font-weight: 500; }
.chip-badge {
  font-size: 10px; padding: 1px 6px;
  background: rgba(255,255,255,0.15); color: rgba(255,255,255,0.85);
  border-radius: 10px; min-width: 16px; text-align: center;
  font-weight: 600;
}
.vista-chip.active .chip-badge { background: rgba(255,255,255,0.25); color: #fff; }
.chip-status-dot {
  width: 8px; height: 8px; border-radius: 50%;
  flex-shrink: 0;
}
.chip-status-dot.connected { background: #10b981; box-shadow: 0 0 8px rgba(16,185,129,0.6); animation: pulse-dot 2s infinite; }
.chip-status-dot.disconnected { background: #ef4444; box-shadow: 0 0 8px rgba(239,68,68,0.6); }
.chip-status-dot.unknown { background: #6b7280; }
@keyframes pulse-dot { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
.vista-empty-tip { font-size: 12px; color: #94a3b8; padding: 6px 10px; }
.vista-empty-tip a { color: #a78bfa; text-decoration: none; }
.vista-empty-tip a:hover { color: #c4b5fd; text-decoration: underline; }

/* 面包屑（保留兼容，现不使用） */
.vista-breadcrumb {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 10px;
  font-size: 13px; color: #cbd5e1;
  white-space: nowrap; overflow: hidden;
}
.crumb { display: inline-flex; align-items: center; gap: 5px; }
.crumb-icon { font-size: 13px; }
.crumb.active { color: #fff; font-weight: 500; }
.crumb.leaf { color: #a78bfa; font-weight: 600; }
.crumb-sep { color: #475569; font-size: 11px; }

/* 【大厂风下拉】集群 + 数据源选择器 */
.vista-selectors {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) minmax(320px, 1.6fr);
  gap: 14px;
  align-items: stretch;
  width: 100%;
}
@media (max-width: 1100px) {
  .vista-selectors { grid-template-columns: 1fr; }
}
.vsel {
  position: relative;
  display: flex; flex-direction: column; gap: 4px;
  min-width: 0;
}
.vsel-label {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 11px; font-weight: 600; color: #c7d2fe;
  text-transform: uppercase; letter-spacing: 0.5px;
  padding-left: 4px;
}
.vsel-btn {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(148, 163, 184, 0.25);
  border-radius: 12px; cursor: pointer;
  font-size: 13px; color: #f1f5f9;
  transition: all 0.2s;
  width: 100%; min-width: 0;
  outline: none;
}
.vsel-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.10);
  border-color: rgba(99, 102, 241, 0.6);
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.25);
}
.vsel-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.vsel.open .vsel-btn {
  background: linear-gradient(135deg, rgba(99,102,241,0.20), rgba(168,85,247,0.20));
  border-color: #a78bfa;
  box-shadow: 0 0 0 3px rgba(167, 139, 250, 0.18);
}
.vsel-icon { font-size: 16px; flex-shrink: 0; }
.vsel-name {
  font-weight: 600; font-size: 14px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  max-width: 200px;
}
.vsel-meta {
  margin-left: auto;
  font-size: 11px; color: #94a3b8;
  font-family: 'SFMono-Regular', Consolas, monospace;
  background: rgba(15, 23, 42, 0.4);
  border: 1px solid rgba(148, 163, 184, 0.15);
  padding: 2px 8px; border-radius: 6px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  max-width: 50%;
}
.vsel-arrow { font-size: 12px; color: #94a3b8; transition: transform 0.2s; flex-shrink: 0; }
.vsel.open .vsel-arrow { transform: rotate(180deg); color: #c4b5fd; }
.vsel-status-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  box-shadow: 0 0 0 2px rgba(15, 23, 42, 0.6);
}
.vsel-status-dot.connected { background: #10b981; box-shadow: 0 0 8px rgba(16,185,129,0.6); }
.vsel-status-dot.disconnected { background: #ef4444; box-shadow: 0 0 8px rgba(239,68,68,0.6); }
.vsel-status-dot.unknown { background: #6b7280; }

.vsel-dropdown {
  position: absolute;
  top: calc(100% + 6px); left: 0;
  width: max(360px, 100%);
  background: #0f172a;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0,0,0,0.45), 0 0 0 1px rgba(99, 102, 241, 0.15);
  z-index: 200; overflow: hidden;
  animation: fadeDown 0.15s ease-out;
}
.vsel-dd-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(99, 102, 241, 0.18);
  font-size: 12px; font-weight: 600; color: #c7d2fe;
  letter-spacing: 0.3px;
}
.vsel-link { font-size: 12px; color: #a78bfa; text-decoration: none; font-weight: 500; }
.vsel-link:hover { color: #c4b5fd; text-decoration: underline; }
.vsel-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background 0.15s;
  border-left: 2px solid transparent;
}
.vsel-item:hover { background: rgba(99, 102, 241, 0.10); }
.vsel-item.active {
  background: linear-gradient(90deg, rgba(99,102,241,0.18), rgba(168,85,247,0.10));
  border-left-color: #a78bfa;
}
.vsel-item.disabled { opacity: 0.45; cursor: not-allowed; }
.vsel-it-icon { font-size: 16px; flex-shrink: 0; }
.vsel-it-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.vsel-it-name {
  font-size: 13px; font-weight: 500; color: #e2e8f0;
  display: inline-flex; align-items: center; gap: 6px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.vsel-it-sub {
  font-size: 11px; color: #94a3b8;
  font-family: 'SFMono-Regular', Consolas, monospace;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.vsel-it-badge {
  font-size: 11px; font-weight: 700;
  background: rgba(99, 102, 241, 0.20);
  color: #c7d2fe;
  padding: 2px 8px; border-radius: 10px;
  border: 1px solid rgba(99, 102, 241, 0.3);
  flex-shrink: 0;
}
.vsel-default-tag, .vsel-shared-tag {
  font-size: 10px; font-weight: 600;
  padding: 1px 6px; border-radius: 4px;
  background: rgba(99,102,241,0.25); color: #c7d2fe;
  border: 1px solid rgba(99,102,241,0.4);
}
.vsel-shared-tag { background: rgba(14, 165, 233, 0.20); color: #7dd3fc; border-color: rgba(14, 165, 233, 0.35); }
.vsel-group { padding: 4px 0 6px; border-bottom: 1px dashed rgba(99,102,241,0.1); }
.vsel-group:last-child { border-bottom: none; }
.vsel-group-title {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 14px 4px;
  font-size: 10px; font-weight: 700;
  color: #94a3b8; text-transform: uppercase; letter-spacing: 0.6px;
}
.vsel-group-count {
  font-size: 10px;
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
  padding: 1px 6px; border-radius: 8px;
  margin-left: auto;
}
.vsel-empty { padding: 20px; text-align: center; font-size: 13px; color: #94a3b8; }
.vsel-empty a { color: #a78bfa; text-decoration: none; }
.vsel-empty a:hover { text-decoration: underline; }

/* 右侧统计 */
.vista-stats { display: flex; align-items: center; gap: 16px; }
.vstat { display: flex; align-items: center; gap: 8px; padding: 6px 12px; background: rgba(255,255,255,0.04); border: 1px solid rgba(148, 163, 184, 0.15); border-radius: 10px; }
.vstat-icon { font-size: 18px; }
.vstat-dot { width: 10px; height: 10px; border-radius: 50%; }
.vstat-dot.connected { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.5); }
.vstat-dot.disconnected { background: #ef4444; box-shadow: 0 0 6px rgba(239,68,68,0.5); }
.vstat-meta { display: flex; flex-direction: column; line-height: 1.1; }
.vstat-num { font-size: 16px; font-weight: 700; color: #f1f5f9; font-variant-numeric: tabular-nums; }
.vstat-label { font-size: 10px; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; }
.vista-config-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 8px 14px;
  background: linear-gradient(135deg, rgba(99,102,241,0.15), rgba(168,85,247,0.15));
  border: 1px solid rgba(167, 139, 250, 0.4);
  border-radius: 10px; cursor: pointer;
  font-size: 13px; font-weight: 500; color: #e0e7ff;
  text-decoration: none; transition: all 0.2s;
  white-space: nowrap;
}
.vista-config-btn:hover { background: linear-gradient(135deg, rgba(99,102,241,0.3), rgba(168,85,247,0.3)); border-color: #a78bfa; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(99,102,241,0.25); }

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
  position: relative;
  overflow: hidden;
  transition: opacity 0.3s;
}
.chart-card.is-loading {
  opacity: 0.7;
}
.chart-loading-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(1px);
  z-index: 2;
  pointer-events: none;
}
.chart-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid #e5e7eb;
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: chart-spin 0.8s linear infinite;
}
@keyframes chart-spin {
  to { transform: rotate(360deg); }
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
/* ============ 大厂风格新增样式 ============ */
/* Hero 区：健康评分圆环 + KPI 网格 */
.hero-grid {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 16px;
  margin-bottom: 20px;
}
.hero-score-card {
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border-radius: 14px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  border: 1px solid #eef2f7;
  position: relative;
  overflow: hidden;
}
.hero-score-card::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 3px;
  background: linear-gradient(90deg, #10b981, #0ea5e9);
}
.hero-score-card.score-critical::before { background: linear-gradient(90deg, #ef4444, #f97316); }
.hero-score-card.score-warning::before  { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.hero-score-card.score-good::before     { background: linear-gradient(90deg, #10b981, #34d399); }
.hero-score-card.score-excellent::before{ background: linear-gradient(90deg, #6366f1, #06b6d4); }

.score-ring { position: relative; width: 130px; height: 130px; flex-shrink: 0; }
.ring-svg { width: 100%; height: 100%; transform: rotate(-90deg); }
.ring-bg { fill: none; stroke: #f1f5f9; stroke-width: 10; }
.ring-fg { fill: none; stroke: #10b981; stroke-width: 10; stroke-linecap: round; transition: stroke-dashoffset 0.8s ease; }
.score-critical .ring-fg { stroke: #ef4444; }
.score-warning  .ring-fg { stroke: #f59e0b; }
.score-good     .ring-fg { stroke: #10b981; }
.score-excellent .ring-fg { stroke: #6366f1; }
.ring-text { position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; }
.ring-score { font-size: 32px; font-weight: 800; color: #1f2937; line-height: 1; }
.ring-label { font-size: 11px; color: #6b7280; margin-top: 6px; }

.score-info { flex: 1; min-width: 0; }
.score-level { margin-bottom: 8px; }
.level-badge { display: inline-block; padding: 3px 10px; border-radius: 6px; font-size: 12px; font-weight: 600; }
.level-badge.level-critical  { background: #fee2e2; color: #b91c1c; }
.level-badge.level-warning   { background: #fef3c7; color: #b45309; }
.level-badge.level-good      { background: #d1fae5; color: #047857; }
.level-badge.level-excellent { background: #e0e7ff; color: #4338ca; }
.score-suggestions { margin: 0; padding-left: 18px; font-size: 12px; color: #6b7280; line-height: 1.7; }
.score-suggestions li { margin: 0; }

/* KPI 网格 */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}
.kpi-card {
  background: #fff;
  border-radius: 12px;
  padding: 14px 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  border: 1px solid #eef2f7;
  transition: all 0.2s;
}
.kpi-card:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.08); border-color: #c7d2fe; }
.kpi-head { display: flex; align-items: center; gap: 6px; margin-bottom: 6px; }
.kpi-icon { font-size: 16px; }
.kpi-label { font-size: 12px; color: #6b7280; font-weight: 500; }
.kpi-value { font-size: 22px; font-weight: 700; line-height: 1.2; }
.kpi-sub { font-size: 11px; color: #9ca3af; margin-top: 2px; }

/* 多维可视化区 */
.viz-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}
.heatmap-card { min-width: 0; }
.chart-sub { font-size: 11px; color: #9ca3af; margin-top: 2px; }

/* 三视角 Tab 卡片 */
.scope-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  border: 1px solid #eef2f7;
  margin-bottom: 24px;
}
.scope-tabs {
  display: flex;
  gap: 4px;
  border-bottom: 2px solid #f1f5f9;
  margin-bottom: 16px;
}
.scope-tabs button {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 10px 18px;
  border: none; background: transparent; cursor: pointer;
  font-size: 13px; font-weight: 500; color: #6b7280;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
}
.scope-tabs button:hover { color: #4f46e5; }
.scope-tabs button.active { color: #4f46e5; border-bottom-color: #4f46e5; font-weight: 600; }
.scope-icon { font-size: 14px; }

.sub-title {
  display: flex; align-items: center;
  font-size: 13px; font-weight: 600; color: #374151;
  margin: 6px 0 10px;
}

/* Pod 视角 双列 */
.pod-view { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.pod-section { min-width: 0; }
.restart-badge {
  display: inline-block; padding: 1px 8px; border-radius: 10px;
  background: #fef3c7; color: #b45309;
  font-family: 'SF Mono','Menlo',monospace; font-size: 12px; font-weight: 600;
}
.restart-badge.hot { background: #fee2e2; color: #b91c1c; }
.mono { font-family: 'SF Mono','Menlo',monospace; color: #4b5563; font-size: 12px; }

/* 节点表详情按钮 */
.detail-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; border: 1px solid #c7d2fe; background: #eef2ff;
  border-radius: 6px; cursor: pointer; font-size: 12px; color: #4338ca;
  transition: all 0.2s; font-weight: 500;
}
.detail-btn:hover { background: #4f46e5; color: #fff; border-color: #4f46e5; box-shadow: 0 2px 6px rgba(79,70,229,0.25); }

/* 响应式 */
@media (max-width: 1280px) {
  .hero-grid { grid-template-columns: 1fr; }
  .kpi-grid { grid-template-columns: repeat(4, 1fr); }
  .viz-grid { grid-template-columns: 1fr; }
  .pod-view { grid-template-columns: 1fr; }
}
@media (max-width: 900px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
}

</style>

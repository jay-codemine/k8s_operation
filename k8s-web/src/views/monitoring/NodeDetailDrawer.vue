<template>
  <transition name="drawer">
    <div v-if="visible" class="drawer-mask" @click.self="close">
      <div class="drawer-panel">
        <!-- 头部 -->
        <div class="drawer-header">
          <div class="drawer-title-wrap">
            <span class="drawer-icon">🖥️</span>
            <div class="drawer-title-info">
              <div class="drawer-title">
                {{ detail?.node_name || '加载中...' }}
                <span v-if="detail?.info?.role" class="role-badge" :class="roleBadgeClass">{{ detail.info.role }}</span>
              </div>
              <div class="drawer-subtitle">
                <span class="instance-tag">{{ instance }}</span>
                <span class="status-dot" :class="detail?.current?.status === 'Ready' ? 'ready' : 'not-ready'"></span>
                <span class="status-text">{{ detail?.current?.status || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="drawer-actions">
            <select v-model="durationVal" class="duration-select" @change="reload">
              <option value="30m">最近 30 分钟</option>
              <option value="1h">最近 1 小时</option>
              <option value="3h">最近 3 小时</option>
              <option value="6h">最近 6 小时</option>
              <option value="12h">最近 12 小时</option>
              <option value="24h">最近 24 小时</option>
            </select>
            <button class="icon-btn" :disabled="loading" @click="reload" title="刷新">
              <span :class="{ spinning: loading }">🔄</span>
            </button>
            <button class="close-btn" @click="close">✕</button>
          </div>
        </div>

        <div v-if="loading && !detail" class="drawer-loading">
          <div class="loading-spin"></div>
          <div>正在拉取节点详情…</div>
        </div>

        <div v-else-if="detail" class="drawer-body">
          <!-- KPI 卡片 -->
          <div class="kpi-row">
            <div class="kpi-card" :class="usageClass(detail.current.cpu_usage)">
              <div class="kpi-head"><span class="kpi-icon">⚡</span><span class="kpi-label">CPU 使用率</span></div>
              <div class="kpi-value">{{ (detail.current.cpu_usage || 0).toFixed(1) }}<span class="unit">%</span></div>
              <div class="kpi-bar"><div class="kpi-fill" :style="{ width: Math.min(detail.current.cpu_usage, 100) + '%' }"></div></div>
              <div class="kpi-sub">共 {{ detail.info?.cpu_total || '-' }}</div>
            </div>
            <div class="kpi-card" :class="usageClass(detail.current.memory_usage)">
              <div class="kpi-head"><span class="kpi-icon">💾</span><span class="kpi-label">内存使用率</span></div>
              <div class="kpi-value">{{ (detail.current.memory_usage || 0).toFixed(1) }}<span class="unit">%</span></div>
              <div class="kpi-bar"><div class="kpi-fill" :style="{ width: Math.min(detail.current.memory_usage, 100) + '%' }"></div></div>
              <div class="kpi-sub">共 {{ detail.info?.memory_total || '-' }}</div>
            </div>
            <div class="kpi-card" :class="usageClass(detail.current.disk_usage)">
              <div class="kpi-head"><span class="kpi-icon">💿</span><span class="kpi-label">磁盘使用率</span></div>
              <div class="kpi-value">{{ (detail.current.disk_usage || 0).toFixed(1) }}<span class="unit">%</span></div>
              <div class="kpi-bar"><div class="kpi-fill" :style="{ width: Math.min(detail.current.disk_usage, 100) + '%' }"></div></div>
              <div class="kpi-sub">共 {{ detail.info?.disk_total || '-' }}</div>
            </div>
            <div class="kpi-card neutral">
              <div class="kpi-head"><span class="kpi-icon">📦</span><span class="kpi-label">Pod 数量</span></div>
              <div class="kpi-value">{{ detail.current.pod_count || 0 }}</div>
              <div class="kpi-sub">Load1 {{ (detail.current.load1 || 0).toFixed(2) }} / Load5 {{ (detail.current.load5 || 0).toFixed(2) }}</div>
              <div class="kpi-sub">网络 ↓{{ formatBytes(detail.current.network_in) }}/s ↑{{ formatBytes(detail.current.network_out) }}/s</div>
            </div>
          </div>

          <!-- 元信息 -->
          <div class="section">
            <div class="section-title"><span>📋</span> 节点元信息</div>
            <div class="meta-grid">
              <div class="meta-item"><span class="meta-key">操作系统</span><span class="meta-val">{{ detail.info?.os || '-' }} ({{ detail.info?.arch || '-' }})</span></div>
              <div class="meta-item"><span class="meta-key">内核版本</span><span class="meta-val mono" :title="detail.info?.kernel">{{ detail.info?.kernel || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">Kubelet</span><span class="meta-val mono">{{ detail.info?.kubelet_version || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">运行时长</span><span class="meta-val">{{ detail.info?.uptime || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">CPU 总核</span><span class="meta-val">{{ detail.info?.cpu_total || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">内存总量</span><span class="meta-val">{{ detail.info?.memory_total || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">磁盘总量</span><span class="meta-val">{{ detail.info?.disk_total || '-' }}</span></div>
              <div class="meta-item"><span class="meta-key">节点角色</span><span class="meta-val">{{ detail.info?.role || '-' }}</span></div>
            </div>
          </div>

          <!-- 趋势图 -->
          <div class="section">
            <div class="section-title"><span>📈</span> 资源趋势（{{ durationLabel }}）</div>
            <div class="trend-grid">
              <div class="trend-item">
                <div class="trend-title">CPU 使用率 (%)</div>
                <div ref="cpuTrendRef" class="trend-chart"></div>
              </div>
              <div class="trend-item">
                <div class="trend-title">内存使用率 (%)</div>
                <div ref="memTrendRef" class="trend-chart"></div>
              </div>
              <div class="trend-item">
                <div class="trend-title">磁盘使用率 (%)</div>
                <div ref="diskTrendRef" class="trend-chart"></div>
              </div>
              <div class="trend-item">
                <div class="trend-title">网络流量 (Bytes/s)</div>
                <div ref="netTrendRef" class="trend-chart"></div>
              </div>
            </div>
          </div>

          <!-- Top Pod -->
          <div class="section">
            <div class="section-title"><span>🚀</span> 该节点 Top 10 Pod（按 CPU）</div>
            <table class="pod-table">
              <thead>
                <tr><th>#</th><th>Pod</th><th>命名空间</th><th>CPU (cores)</th><th>内存 (MB)</th></tr>
              </thead>
              <tbody>
                <tr v-for="(p, i) in detail.top_pods || []" :key="p.namespace + '/' + p.name">
                  <td><span class="rank-badge" :class="'rank-' + (i + 1)">{{ i + 1 }}</span></td>
                  <td class="pod-name" :title="p.name">{{ p.name }}</td>
                  <td><span class="ns-tag">{{ p.namespace }}</span></td>
                  <td class="mono">{{ (p.cpu_usage || 0).toFixed(3) }}</td>
                  <td class="mono">{{ ((p.memory_usage || 0) / 1048576).toFixed(1) }}</td>
                </tr>
                <tr v-if="!detail.top_pods || detail.top_pods.length === 0">
                  <td colspan="5" class="empty-row">该节点暂无 Pod 数据</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else class="drawer-loading">
          <div>暂无数据</div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { getNodeDetail } from '@/api/monitoring'

const props = defineProps({
  visible: { type: Boolean, default: false },
  instance: { type: String, default: '' },
})
const emit = defineEmits(['update:visible'])

const loading = ref(true)
const detail = ref(null)
const durationVal = ref('1h')

const cpuTrendRef = ref(null)
const memTrendRef = ref(null)
const diskTrendRef = ref(null)
const netTrendRef = ref(null)
let cpuChart, memChart, diskChart, netChart

const durationLabel = computed(() => {
  const map = { '30m': '最近 30 分钟', '1h': '最近 1 小时', '3h': '最近 3 小时', '6h': '最近 6 小时', '12h': '最近 12 小时', '24h': '最近 24 小时' }
  return map[durationVal.value] || durationVal.value
})

const roleBadgeClass = computed(() => {
  const r = (detail.value?.info?.role || '').toLowerCase()
  if (r.includes('master') || r.includes('control')) return 'role-master'
  return 'role-worker'
})

function usageClass(v) {
  const n = Number(v) || 0
  if (n >= 85) return 'level-critical'
  if (n >= 70) return 'level-warning'
  return 'level-good'
}

function formatBytes(v) {
  const n = Number(v) || 0
  if (n < 1024) return n.toFixed(0) + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB'
  if (n < 1024 * 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + ' MB'
  return (n / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

function formatTime(ts) {
  const d = new Date(ts * 1000)
  return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0')
}

async function reload() {
  if (!props.instance) return
  loading.value = true
  try {
    const res = await getNodeDetail(props.instance, durationVal.value)
    if (res.code === 0) {
      detail.value = res.data
      await nextTick()
      renderCharts()
    }
  } catch (e) {
    console.warn('加载节点详情失败', e)
  } finally {
    loading.value = false
  }
}

function buildLineOption(points, color, unit) {
  const xs = (points || []).map(p => formatTime(p.timestamp))
  const ys = (points || []).map(p => Number((p.value || 0).toFixed(2)))
  return {
    grid: { top: 8, left: 40, right: 12, bottom: 24 },
    tooltip: { trigger: 'axis', formatter: (ps) => `${ps[0].axisValue}<br/><b>${ps[0].data}${unit || ''}</b>` },
    xAxis: { type: 'category', data: xs, axisLabel: { color: '#9ca3af', fontSize: 10, interval: Math.max(0, Math.floor(xs.length / 6) - 1) }, axisLine: { lineStyle: { color: '#e5e7eb' } } },
    yAxis: { type: 'value', axisLabel: { color: '#9ca3af', fontSize: 10 }, splitLine: { lineStyle: { color: '#f3f4f6' } } },
    series: [{
      type: 'line', data: ys, smooth: true, symbol: 'none',
      lineStyle: { color, width: 2 },
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: color + '55' }, { offset: 1, color: color + '00' }] } },
    }],
  }
}

function buildNetOption(inPts, outPts) {
  const xs = (inPts || []).map(p => formatTime(p.timestamp))
  const yIn = (inPts || []).map(p => Number((p.value || 0).toFixed(0)))
  const yOut = (outPts || []).map(p => Number((p.value || 0).toFixed(0)))
  return {
    grid: { top: 28, left: 50, right: 12, bottom: 24 },
    legend: { top: 0, right: 0, textStyle: { color: '#6b7280', fontSize: 10 }, itemWidth: 10, itemHeight: 8 },
    tooltip: { trigger: 'axis', formatter: (ps) => ps.map(p => `${p.marker}${p.seriesName}: <b>${formatBytes(p.data)}/s</b>`).join('<br/>') },
    xAxis: { type: 'category', data: xs, axisLabel: { color: '#9ca3af', fontSize: 10, interval: Math.max(0, Math.floor(xs.length / 6) - 1) }, axisLine: { lineStyle: { color: '#e5e7eb' } } },
    yAxis: { type: 'value', axisLabel: { color: '#9ca3af', fontSize: 10, formatter: (v) => formatBytes(v) }, splitLine: { lineStyle: { color: '#f3f4f6' } } },
    series: [
      { name: '入站', type: 'line', data: yIn, smooth: true, symbol: 'none', lineStyle: { color: '#0ea5e9', width: 2 }, areaStyle: { color: '#0ea5e933' } },
      { name: '出站', type: 'line', data: yOut, smooth: true, symbol: 'none', lineStyle: { color: '#f97316', width: 2 }, areaStyle: { color: '#f9731633' } },
    ],
  }
}

function renderCharts() {
  const t = detail.value?.trends || {}
  if (cpuTrendRef.value) {
    cpuChart = cpuChart || echarts.init(cpuTrendRef.value)
    cpuChart.setOption(buildLineOption(t.cpu, '#3b82f6', '%'), true)
  }
  if (memTrendRef.value) {
    memChart = memChart || echarts.init(memTrendRef.value)
    memChart.setOption(buildLineOption(t.memory, '#8b5cf6', '%'), true)
  }
  if (diskTrendRef.value) {
    diskChart = diskChart || echarts.init(diskTrendRef.value)
    diskChart.setOption(buildLineOption(t.disk, '#10b981', '%'), true)
  }
  if (netTrendRef.value) {
    netChart = netChart || echarts.init(netTrendRef.value)
    netChart.setOption(buildNetOption(t.net_in, t.net_out), true)
  }
}

function disposeCharts() {
  cpuChart?.dispose(); cpuChart = null
  memChart?.dispose(); memChart = null
  diskChart?.dispose(); diskChart = null
  netChart?.dispose(); netChart = null
}

function close() {
  emit('update:visible', false)
}

function onResize() {
  cpuChart?.resize(); memChart?.resize(); diskChart?.resize(); netChart?.resize()
}

watch(() => props.visible, (v) => {
  if (v) {
    detail.value = null
    durationVal.value = '1h'
    reload()
    window.addEventListener('resize', onResize)
  } else {
    disposeCharts()
    window.removeEventListener('resize', onResize)
  }
})

watch(() => props.instance, () => {
  if (props.visible) reload()
})

onUnmounted(() => {
  disposeCharts()
  window.removeEventListener('resize', onResize)
})
</script>

<style scoped>
.drawer-mask {
  position: fixed; inset: 0; background: rgba(15, 23, 42, 0.45);
  z-index: 2000; display: flex; justify-content: flex-end;
}
.drawer-panel {
  width: 760px; max-width: 95vw; height: 100vh; background: #f7f9fc;
  display: flex; flex-direction: column; overflow: hidden;
  box-shadow: -8px 0 32px rgba(0,0,0,0.18);
}
.drawer-enter-active, .drawer-leave-active { transition: opacity 0.2s; }
.drawer-enter-active .drawer-panel, .drawer-leave-active .drawer-panel { transition: transform 0.25s ease; }
.drawer-enter-from, .drawer-leave-to { opacity: 0; }
.drawer-enter-from .drawer-panel, .drawer-leave-to .drawer-panel { transform: translateX(100%); }

.drawer-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 20px; background: #fff; border-bottom: 1px solid #eef2f7;
  flex-shrink: 0;
}
.drawer-title-wrap { display: flex; align-items: center; gap: 12px; min-width: 0; flex: 1; }
.drawer-icon { font-size: 28px; }
.drawer-title-info { min-width: 0; }
.drawer-title { font-size: 17px; font-weight: 700; color: #1f2937; display: flex; align-items: center; gap: 8px; }
.role-badge { font-size: 11px; padding: 2px 8px; border-radius: 6px; font-weight: 600; }
.role-master { background: #ede9fe; color: #6d28d9; }
.role-worker { background: #dcfce7; color: #15803d; }
.drawer-subtitle { display: flex; align-items: center; gap: 8px; margin-top: 4px; font-size: 12px; color: #6b7280; }
.instance-tag { font-family: 'SF Mono','Menlo',monospace; background: #f1f5f9; padding: 2px 8px; border-radius: 4px; color: #475569; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-dot.ready { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.5); }
.status-dot.not-ready { background: #ef4444; }
.status-text { font-weight: 500; }

.drawer-actions { display: flex; align-items: center; gap: 8px; }
.duration-select { padding: 5px 10px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 12px; color: #475569; background: #fff; cursor: pointer; }
.icon-btn { padding: 4px 10px; border: 1px solid #e2e8f0; background: #fff; border-radius: 6px; cursor: pointer; font-size: 14px; }
.icon-btn:hover { border-color: #4f46e5; }
.icon-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.spinning { display: inline-block; animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.close-btn { padding: 4px 10px; border: none; background: transparent; cursor: pointer; font-size: 16px; color: #94a3b8; }
.close-btn:hover { color: #ef4444; }

.drawer-body { flex: 1; overflow-y: auto; padding: 16px 20px 32px; }
.drawer-loading { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: #94a3b8; font-size: 13px; }
.loading-spin { width: 32px; height: 32px; border: 3px solid #e5e7eb; border-top-color: #4f46e5; border-radius: 50%; animation: spin 0.8s linear infinite; }

/* KPI 卡片行 */
.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 18px; }
.kpi-card {
  background: #fff; border-radius: 10px; padding: 14px;
  border: 1px solid #eef2f7; transition: all 0.2s;
  border-left: 3px solid #cbd5e1;
}
.kpi-card.level-good { border-left-color: #10b981; }
.kpi-card.level-warning { border-left-color: #f59e0b; }
.kpi-card.level-critical { border-left-color: #ef4444; }
.kpi-card.neutral { border-left-color: #6366f1; }
.kpi-head { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; }
.kpi-icon { font-size: 14px; }
.kpi-label { font-size: 12px; color: #6b7280; font-weight: 500; }
.kpi-value { font-size: 22px; font-weight: 700; color: #1f2937; line-height: 1; }
.kpi-value .unit { font-size: 13px; color: #9ca3af; font-weight: 500; margin-left: 2px; }
.kpi-bar { width: 100%; height: 4px; background: #f1f5f9; border-radius: 2px; margin-top: 8px; overflow: hidden; }
.kpi-fill { height: 100%; background: linear-gradient(90deg, #10b981, #0ea5e9); transition: width 0.6s ease; border-radius: 2px; }
.level-warning .kpi-fill { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.level-critical .kpi-fill { background: linear-gradient(90deg, #ef4444, #f97316); }
.kpi-sub { font-size: 11px; color: #94a3b8; margin-top: 6px; }

/* 通用 section */
.section { background: #fff; border-radius: 10px; padding: 16px 18px; margin-bottom: 16px; border: 1px solid #eef2f7; }
.section-title { font-size: 14px; font-weight: 600; color: #1f2937; margin-bottom: 14px; display: flex; align-items: center; gap: 6px; }

/* 元信息 grid */
.meta-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px 24px; }
.meta-item { display: flex; justify-content: space-between; align-items: center; padding: 6px 0; border-bottom: 1px dashed #f1f5f9; gap: 12px; }
.meta-key { font-size: 12px; color: #94a3b8; flex-shrink: 0; }
.meta-val { font-size: 13px; color: #1f2937; font-weight: 500; text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.meta-val.mono { font-family: 'SF Mono','Menlo',monospace; font-size: 12px; }

/* 趋势 */
.trend-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.trend-item { background: #fafbff; border: 1px solid #eef2f7; border-radius: 8px; padding: 10px 12px 6px; }
.trend-title { font-size: 12px; color: #6b7280; font-weight: 500; margin-bottom: 4px; }
.trend-chart { width: 100%; height: 150px; }

/* Pod 表 */
.pod-table { width: 100%; border-collapse: collapse; }
.pod-table th { text-align: left; padding: 8px 10px; font-size: 11px; font-weight: 600; color: #94a3b8; text-transform: uppercase; border-bottom: 1px solid #f1f5f9; }
.pod-table td { padding: 8px 10px; font-size: 13px; color: #374151; border-bottom: 1px solid #f9fafb; }
.pod-table tr:hover td { background: #fafbff; }
.rank-badge { display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; border-radius: 6px; font-size: 11px; font-weight: 700; background: #f3f4f6; color: #6b7280; }
.rank-1 { background: #fef3c7; color: #d97706; }
.rank-2 { background: #e5e7eb; color: #4b5563; }
.rank-3 { background: #fce7f3; color: #db2777; }
.pod-name { font-weight: 500; color: #1f2937; max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ns-tag { display: inline-block; padding: 2px 8px; background: #eff6ff; color: #3b82f6; border-radius: 4px; font-size: 12px; }
.mono { font-family: 'SF Mono','Menlo',monospace; }
.empty-row { text-align: center; color: #9ca3af; padding: 24px 0 !important; }

/* 响应式 */
@media (max-width: 900px) {
  .drawer-panel { width: 100vw; }
  .kpi-row { grid-template-columns: repeat(2, 1fr); }
  .trend-grid { grid-template-columns: 1fr; }
  .meta-grid { grid-template-columns: 1fr; }
}
</style>

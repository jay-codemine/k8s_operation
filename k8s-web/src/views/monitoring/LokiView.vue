<template>
  <div class="loki-view">
    <!-- 顶部状态条 -->
    <div class="status-bar">
      <div class="status-left">
        <span class="status-dot" :class="healthStatus"></span>
        <span class="status-text">
          Loki
          <em v-if="healthStatus === 'healthy'">已连接</em>
          <em v-else-if="healthStatus === 'unhealthy'" class="err">未连接</em>
          <em v-else>检测中…</em>
        </span>
        <span class="status-url" v-if="lokiUrl" :title="lokiUrl">{{ lokiUrl }}</span>
      </div>
      <div class="status-right">
        <button class="refresh-btn" :disabled="querying" @click="executeQuery(); loadVolume();" title="刷新">
          <span :class="{ rotating: querying }">⟳</span>
          刷新
        </button>
      </div>
    </div>

    <!-- 查询区域 -->
    <div class="query-panel">
      <div class="query-bar">
        <div class="query-input-wrap">
          <span class="query-prefix">LogQL</span>
          <input
            v-model="queryExpr"
            class="query-input"
            placeholder='{job="varlogs"} |= "error"'
            @keydown.enter="executeQuery"
          />
          <button v-if="queryExpr" class="query-clear" @click="queryExpr = ''" title="清空">×</button>
        </div>
        <div class="query-controls">
          <select v-model="queryDuration" class="ctrl-select">
            <option value="5m">近 5 分钟</option>
            <option value="15m">近 15 分钟</option>
            <option value="30m">近 30 分钟</option>
            <option value="1h">近 1 小时</option>
            <option value="3h">近 3 小时</option>
            <option value="6h">近 6 小时</option>
            <option value="12h">近 12 小时</option>
            <option value="24h">近 24 小时</option>
          </select>
          <select v-model="queryLimit" class="ctrl-select ctrl-select-sm">
            <option :value="50">50 条</option>
            <option :value="100">100 条</option>
            <option :value="200">200 条</option>
            <option :value="500">500 条</option>
          </select>
          <button class="query-btn" @click="executeQuery" :disabled="querying">
            <span v-if="querying" class="btn-spinner"></span>
            <span v-else>▶</span>
            查询
          </button>
        </div>
      </div>
      <!-- 标签快捷选择 -->
      <div class="label-bar" v-if="labels.length">
        <span class="label-bar-title">标签筛选</span>
        <div class="label-chips">
          <button
            v-for="label in labels.slice(0, 12)"
            :key="label"
            class="label-chip"
            :class="{ active: selectedLabel === label }"
            @click="selectLabel(label)"
          >
            {{ label }}
          </button>
          <button v-if="labels.length > 12" class="label-chip label-more" @click="showAllLabels = !showAllLabels">
            +{{ labels.length - 12 }} 更多
          </button>
        </div>
        <!-- 标签值下拉 -->
        <div class="label-values-panel" v-if="selectedLabel && labelValues.length">
          <span class="lv-title">{{ selectedLabel }} =</span>
          <div class="lv-list">
            <button
              v-for="val in labelValues.slice(0, 20)"
              :key="val"
              class="lv-item"
              @click="applyLabelFilter(selectedLabel, val)"
            >
              {{ val }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 统计摘要 -->
    <div class="stats-row" v-if="queryResult">
      <div class="stat-item">
        <span class="stat-icon">📊</span>
        <span class="stat-label">日志流</span>
        <span class="stat-value">{{ queryResult.streams }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-icon">📝</span>
        <span class="stat-label">总行数</span>
        <span class="stat-value">{{ queryResult.total_lines }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-icon">⏱️</span>
        <span class="stat-label">耗时</span>
        <span class="stat-value">{{ queryTime }}ms</span>
      </div>
    </div>

    <!-- 日志量趋势图 -->
    <div class="volume-card" v-if="showVolume">
      <div class="volume-header">
        <h3>日志量趋势</h3>
        <button class="volume-toggle" @click="showVolume = false">收起 ▴</button>
      </div>
      <div ref="volumeChartRef" class="volume-chart"></div>
    </div>
    <div class="volume-expand" v-else>
      <button @click="loadVolume(); showVolume = true">展开日志量趋势 ▾</button>
    </div>

    <!-- 日志列表 -->
    <div class="logs-panel">
      <div class="logs-toolbar">
        <div class="toolbar-left">
          <h3>日志结果</h3>
          <span class="log-count" v-if="logEntries.length">{{ logEntries.length }} 条记录</span>
        </div>
        <div class="toolbar-right">
          <button class="tool-btn" :class="{ active: wrapLines }" @click="wrapLines = !wrapLines" title="自动换行">
            ↩ 换行
          </button>
          <button class="tool-btn" :class="{ active: showLabels }" @click="showLabels = !showLabels" title="显示标签">
            🏷️ 标签
          </button>
          <select v-model="logDirection" class="ctrl-select ctrl-select-sm" @change="executeQuery">
            <option value="backward">最新在前</option>
            <option value="forward">最旧在前</option>
          </select>
        </div>
      </div>

      <!-- 日志条目 -->
      <div class="logs-list" :class="{ 'wrap-mode': wrapLines }">
        <!-- 错误状态 -->
        <div class="log-empty error" v-if="queryError && !querying && !logEntries.length">
          <div class="empty-icon">⚠️</div>
          <p class="empty-title">查询出错</p>
          <p class="empty-desc">{{ queryError }}</p>
          <button class="empty-action" @click="executeQuery">重试</button>
        </div>
        <!-- 空数据 / 引导 -->
        <div class="log-empty" v-else-if="!logEntries.length && !querying">
          <div class="empty-icon">📋</div>
          <p class="empty-title">{{ queryResult ? '该时间范围内未匹配到日志' : '输入 LogQL 查询表达式开始探索日志' }}</p>
          <p class="empty-desc" v-if="queryResult">尝试调整时间范围或修改查询条件</p>
          <div class="empty-examples">
            <span>示例:</span>
            <code @click="setExample(exampleQueries[0])">{{ exampleQueries[0] }}</code>
            <code @click="setExample(exampleQueries[1])">{{ exampleQueries[1] }}</code>
            <code @click="setExample(exampleQueries[2])">{{ exampleQueries[2] }}</code>
          </div>
        </div>

        <div class="log-entry" v-for="(entry, idx) in logEntries" :key="idx">
          <span class="log-ts">{{ formatTimestamp(entry.timestamp) }}</span>
          <span class="log-labels" v-if="showLabels">
            <span class="log-label-tag" v-for="(v, k) in entry.labels" :key="k">{{ k }}={{ v }}</span>
          </span>
          <span class="log-line" :class="getLogLevel(entry.line)">{{ entry.line }}</span>
        </div>

        <div class="log-loading" v-if="querying">
          <div class="loading-spinner"></div>
          <span>查询中...</span>
        </div>
      </div>
    </div>

    <!-- 右侧快捷面板 - 日志流 -->
    <div class="streams-panel" v-if="streams.length">
      <div class="streams-header">
        <h4>活跃日志流</h4>
        <span class="stream-count">{{ streams.length }}</span>
      </div>
      <div class="streams-list">
        <div
          class="stream-item"
          v-for="(stream, i) in streams.slice(0, 20)"
          :key="i"
          @click="applyStreamQuery(stream)"
        >
          <div class="stream-labels">
            <span class="stream-label" v-for="(v, k) in stream.labels" :key="k">
              <em>{{ k }}</em>={{ v }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { Message } from '@arco-design/web-vue'
import {
  checkLokiHealth,
  queryLokiLogs,
  getLokiLabels,
  getLokiLabelValues,
  getLokiStreams,
  getLokiVolume,
} from '@/api/monitoring'

// 数据源相关
const querying = ref(false)
const queryError = ref('')
const healthStatus = ref('checking') // checking | healthy | unhealthy
const lokiUrl = ref('')

// 示例查询
const exampleQueries = [
  '{job="varlogs"}',
  '{namespace="default"} |= "error"',
  '{app="nginx"} | json | status >= 400',
]
const queryExpr = ref('')
const queryDuration = ref('1h')
const queryLimit = ref(100)
const logDirection = ref('backward')
const queryTime = ref(0)
const queryResult = ref(null)
const logEntries = ref([])
const wrapLines = ref(true)
const showLabels = ref(false)
const showVolume = ref(true)

// 标签相关
const labels = ref([])
const selectedLabel = ref('')
const labelValues = ref([])
const showAllLabels = ref(false)

// 日志流
const streams = ref([])

// 图表
const volumeChartRef = ref(null)
let volumeChart = null

// ===== 方法 =====

async function executeQuery() {
  const expr = queryExpr.value.trim()
  if (!expr) {
    Message.warning({ content: '请输入 LogQL 查询表达式（例如 {job="varlogs"}）' })
    return
  }
  querying.value = true
  queryError.value = ''
  const startTime = Date.now()

  try {
    const res = await queryLokiLogs({
      query: expr,
      duration: queryDuration.value,
      limit: queryLimit.value,
      direction: logDirection.value,
    })
    if (res?.code === 0 && res.data) {
      queryResult.value = res.data
      logEntries.value = res.data.entries || []
      if (logEntries.value.length === 0) {
        Message.info({ content: '查询成功，但所选时间范围内无匹配日志' })
      }
    } else {
      queryError.value = res?.msg || '查询失败'
      Message.error({ content: queryError.value })
    }
  } catch (e) {
    queryError.value = e?.msg || e?.message || 'Loki 查询失败，请检查数据源连通性'
    Message.error({ content: queryError.value })
    console.warn('Loki 查询失败', e)
  } finally {
    queryTime.value = Date.now() - startTime
    querying.value = false
  }
}

async function checkHealth() {
  try {
    const res = await checkLokiHealth()
    if (res?.code === 0 && res.data) {
      healthStatus.value = res.data.healthy ? 'healthy' : 'unhealthy'
      lokiUrl.value = res.data.url || ''
    } else {
      healthStatus.value = 'unhealthy'
    }
  } catch {
    healthStatus.value = 'unhealthy'
  }
}

async function loadLabels() {
  try {
    const res = await getLokiLabels({ duration: queryDuration.value })
    if (res?.code === 0) {
      labels.value = res.data || []
    }
  } catch {}
}

async function selectLabel(label) {
  if (selectedLabel.value === label) {
    selectedLabel.value = ''
    labelValues.value = []
    return
  }
  selectedLabel.value = label
  try {
    const res = await getLokiLabelValues(label, { duration: queryDuration.value })
    if (res?.code === 0) {
      labelValues.value = res.data || []
    }
  } catch {}
}

function applyLabelFilter(label, value) {
  queryExpr.value = `{${label}="${value}"}`
  selectedLabel.value = ''
  labelValues.value = []
  executeQuery()
}

function applyStreamQuery(stream) {
  const parts = Object.entries(stream.labels).map(([k, v]) => `${k}="${v}"`)
  queryExpr.value = `{${parts.join(', ')}}`
  executeQuery()
}

function setExample(expr) {
  queryExpr.value = expr
  // 点击示例后自动执行，提升体验
  executeQuery()
}

async function loadStreams() {
  try {
    const res = await getLokiStreams({ duration: queryDuration.value })
    if (res?.code === 0) {
      streams.value = res.data || []
    }
  } catch {}
}

async function loadVolume() {
  try {
    const res = await getLokiVolume({
      query: queryExpr.value || '',
      duration: queryDuration.value,
    })
    if (res?.code === 0 && res.data) {
      renderVolumeChart(res.data)
    }
  } catch {}
}

function renderVolumeChart(data) {
  if (!volumeChartRef.value) return
  if (!volumeChart) {
    volumeChart = echarts.init(volumeChartRef.value)
  }

  const allPoints = data.flatMap(s => s.points || [])
  const timestamps = [...new Set(allPoints.map(p => p.timestamp))].sort()

  const series = data.map((s, i) => {
    const colors = ['#6366f1', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b']
    const name = s.labels?.job || `stream-${i + 1}`
    const pointMap = {}
    ;(s.points || []).forEach(p => { pointMap[p.timestamp] = p.count })

    return {
      name,
      type: 'bar',
      stack: 'volume',
      barWidth: '80%',
      itemStyle: { color: colors[i % colors.length], borderRadius: [2, 2, 0, 0] },
      data: timestamps.map(ts => pointMap[ts] || 0),
    }
  })

  const option = {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { top: 10, right: 16, bottom: 24, left: 50 },
    xAxis: {
      type: 'category',
      data: timestamps.map(ts => {
        const d = new Date(ts * 1000)
        return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0')
      }),
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#9ca3af', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#9ca3af', fontSize: 11 },
      splitLine: { lineStyle: { color: '#f3f4f6' } },
    },
    series,
  }
  volumeChart.setOption(option, true)
}

function formatTimestamp(ms) {
  const d = new Date(ms)
  return d.toLocaleTimeString('zh-CN', { hour12: false }) + '.' + String(d.getMilliseconds()).padStart(3, '0')
}

function getLogLevel(line) {
  const lower = line.toLowerCase()
  if (lower.includes('error') || lower.includes('fatal') || lower.includes('panic')) return 'level-error'
  if (lower.includes('warn')) return 'level-warn'
  if (lower.includes('info')) return 'level-info'
  if (lower.includes('debug') || lower.includes('trace')) return 'level-debug'
  return ''
}

function handleResize() {
  volumeChart?.resize()
}

onMounted(async () => {
  await nextTick()
  // 1. 先做 Loki 健康检查
  await checkHealth()
  if (healthStatus.value === 'unhealthy') {
    Message.warning({
      content: '未检测到可用的 Loki 数据源，请前往【数据源管理】添加并设为默认',
      duration: 4000,
    })
    return
  }
  // 2. 并行拉取标签 / 流
  await Promise.all([loadLabels(), loadStreams()])
  // 3. 自动选择默认查询表达式：优先使用第一个活跃流的 selector
  if (!queryExpr.value && streams.value.length) {
    const first = streams.value[0]
    if (first.label_str) {
      queryExpr.value = first.label_str
    } else if (first.labels && Object.keys(first.labels).length) {
      const parts = Object.entries(first.labels).map(([k, v]) => `${k}="${v}"`)
      queryExpr.value = `{${parts.join(', ')}}`
    }
  }
  // 4. 自动执行首次查询，避免页面进来一片空白
  if (queryExpr.value) {
    await executeQuery()
  }
  // 5. 加载日志量趋势
  if (showVolume.value) {
    await loadVolume()
  }
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  volumeChart?.dispose()
})
</script>

<style scoped>
.loki-view {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100vh;
  background: #f8fafc;
}

/* 顶部状态条 */
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.status-left { display: flex; align-items: center; gap: 10px; }
.status-dot {
  width: 8px; height: 8px; border-radius: 50%; background: #cbd5e1;
  box-shadow: 0 0 0 3px rgba(203,213,225,0.25);
}
.status-dot.healthy { background: #10b981; box-shadow: 0 0 0 3px rgba(16,185,129,0.18); animation: pulse 2s infinite; }
.status-dot.unhealthy { background: #ef4444; box-shadow: 0 0 0 3px rgba(239,68,68,0.18); }
.status-dot.checking { background: #f59e0b; box-shadow: 0 0 0 3px rgba(245,158,11,0.18); animation: pulse 1.2s infinite; }
@keyframes pulse {
  0%,100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.15); opacity: 0.7; }
}
.status-text { font-size: 13px; color: #1e293b; font-weight: 600; }
.status-text em { font-style: normal; color: #10b981; font-weight: 500; margin-left: 4px; }
.status-text em.err { color: #ef4444; }
.status-url {
  font-size: 11px; color: #94a3b8; font-family: 'SF Mono', monospace;
  max-width: 360px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  background: #f1f5f9; padding: 2px 8px; border-radius: 4px;
}
.refresh-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 6px; font-size: 12px; color: #475569;
  cursor: pointer; transition: all 0.15s;
}
.refresh-btn:hover { border-color: #6366f1; color: #6366f1; }
.refresh-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.refresh-btn .rotating { display: inline-block; animation: spin 0.8s linear infinite; }

/* 查询输入框清空按钮 */
.query-clear {
  margin-right: 8px; padding: 0; width: 22px; height: 22px;
  border: none; background: #f1f5f9; color: #64748b;
  border-radius: 50%; font-size: 14px; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
}
.query-clear:hover { background: #e2e8f0; color: #1e293b; }

/* 查询面板 */
.query-panel {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.query-bar {
  display: flex;
  gap: 12px;
  align-items: center;
}
.query-input-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.query-input-wrap:focus-within {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}
.query-prefix {
  padding: 10px 14px;
  background: #f1f5f9;
  color: #6366f1;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
  white-space: nowrap;
  border-right: 1px solid #e2e8f0;
}
.query-input {
  flex: 1;
  padding: 10px 14px;
  border: none;
  outline: none;
  font-size: 13px;
  font-family: 'JetBrains Mono', 'SF Mono', 'Menlo', monospace;
  color: #1e293b;
  background: transparent;
}
.query-input::placeholder { color: #94a3b8; }
.query-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}
.ctrl-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 12px;
  color: #475569;
  background: #fff;
  cursor: pointer;
}
.ctrl-select-sm { padding: 8px 8px; }
.query-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 20px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.query-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(99,102,241,0.3); }
.query-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.btn-spinner {
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* 标签筛选 */
.label-bar {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #f1f5f9;
}
.label-bar-title {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  margin-right: 12px;
}
.label-chips {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}
.label-chip {
  padding: 4px 10px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 12px;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
}
.label-chip:hover { background: #e0e7ff; border-color: #c7d2fe; color: #4338ca; }
.label-chip.active { background: #6366f1; color: #fff; border-color: #6366f1; }
.label-more { font-style: italic; color: #94a3b8; }

.label-values-panel {
  margin-top: 10px;
  padding: 10px 14px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}
.lv-title { font-size: 12px; font-weight: 600; color: #6366f1; }
.lv-list { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.lv-item {
  padding: 3px 8px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 11px;
  color: #334155;
  font-family: monospace;
  cursor: pointer;
  transition: all 0.15s;
}
.lv-item:hover { background: #6366f1; color: #fff; border-color: #6366f1; }

/* 统计摘要 */
.stats-row {
  display: flex;
  gap: 16px;
}
.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.stat-icon { font-size: 18px; }
.stat-label { font-size: 12px; color: #94a3b8; }
.stat-value { font-size: 16px; font-weight: 700; color: #1e293b; font-family: 'SF Mono', monospace; }

/* 日志量趋势 */
.volume-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.volume-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.volume-header h3 { font-size: 14px; font-weight: 600; color: #374151; margin: 0; }
.volume-toggle {
  font-size: 12px; color: #94a3b8; cursor: pointer; background: none; border: none;
}
.volume-chart { width: 100%; height: 140px; }
.volume-expand {
  text-align: center;
}
.volume-expand button {
  font-size: 12px; color: #6366f1; cursor: pointer; background: #fff;
  border: 1px solid #e0e7ff; border-radius: 6px; padding: 6px 16px;
}

/* 日志面板 */
.logs-panel {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  overflow: hidden;
}
.logs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px;
  border-bottom: 1px solid #f1f5f9;
}
.toolbar-left { display: flex; align-items: center; gap: 12px; }
.toolbar-left h3 { font-size: 14px; font-weight: 600; color: #374151; margin: 0; }
.log-count { font-size: 12px; color: #94a3b8; background: #f1f5f9; padding: 2px 8px; border-radius: 4px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.tool-btn {
  padding: 5px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 12px;
  color: #64748b;
  cursor: pointer;
  background: #fff;
  transition: all 0.15s;
}
.tool-btn:hover { border-color: #6366f1; color: #6366f1; }
.tool-btn.active { background: #6366f1; color: #fff; border-color: #6366f1; }

/* 日志列表 */
.logs-list {
  max-height: 600px;
  overflow-y: auto;
  font-family: 'JetBrains Mono', 'SF Mono', 'Menlo', monospace;
  font-size: 12px;
}
.logs-list.wrap-mode .log-line { white-space: pre-wrap; word-break: break-all; }
.log-entry {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 6px 20px;
  border-bottom: 1px solid #fafbfc;
  transition: background 0.1s;
}
.log-entry:hover { background: #f8fafc; }
.log-ts {
  flex-shrink: 0;
  color: #94a3b8;
  font-size: 11px;
  padding-top: 1px;
  min-width: 90px;
}
.log-labels { display: flex; flex-wrap: wrap; gap: 4px; flex-shrink: 0; max-width: 300px; }
.log-label-tag {
  font-size: 10px;
  background: #f1f5f9;
  color: #64748b;
  padding: 1px 5px;
  border-radius: 3px;
}
.log-line {
  flex: 1;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.5;
}
.log-line.level-error { color: #dc2626; }
.log-line.level-warn { color: #d97706; }
.log-line.level-info { color: #1e293b; }
.log-line.level-debug { color: #6b7280; }

.log-empty {
  text-align: center;
  padding: 60px 20px;
  color: #94a3b8;
}
.log-empty.error .empty-icon { filter: hue-rotate(-30deg); }
.log-empty.error .empty-title { color: #ef4444; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.log-empty p { font-size: 14px; margin: 0 0 8px; }
.empty-title { font-weight: 600; color: #475569; font-size: 15px !important; }
.empty-desc { font-size: 12px !important; color: #94a3b8; margin-bottom: 16px !important; max-width: 480px; margin-left: auto; margin-right: auto; word-break: break-all; }
.empty-action {
  margin-top: 8px; padding: 7px 22px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff; border: none; border-radius: 6px;
  font-size: 13px; cursor: pointer;
  transition: all 0.2s;
}
.empty-action:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(99,102,241,0.3); }
.empty-examples { display: flex; align-items: center; justify-content: center; gap: 8px; flex-wrap: wrap; }
.empty-examples span { font-size: 12px; color: #64748b; }
.empty-examples code {
  padding: 4px 10px;
  background: #f1f5f9;
  border-radius: 6px;
  font-size: 11px;
  color: #6366f1;
  cursor: pointer;
  transition: all 0.15s;
}
.empty-examples code:hover { background: #e0e7ff; }

.log-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 30px;
  color: #94a3b8;
  font-size: 13px;
}
.loading-spinner {
  width: 18px; height: 18px;
  border: 2px solid #e2e8f0;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* 日志流面板 */
.streams-panel {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.streams-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.streams-header h4 { font-size: 14px; font-weight: 600; color: #374151; margin: 0; }
.stream-count {
  font-size: 11px;
  background: #f1f5f9;
  color: #6366f1;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.streams-list { display: flex; flex-direction: column; gap: 6px; max-height: 300px; overflow-y: auto; }
.stream-item {
  padding: 8px 12px;
  border: 1px solid #f1f5f9;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.stream-item:hover { background: #f8fafc; border-color: #e0e7ff; }
.stream-labels { display: flex; flex-wrap: wrap; gap: 4px; }
.stream-label {
  font-size: 11px;
  color: #475569;
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 4px;
}
.stream-label em { font-style: normal; color: #6366f1; font-weight: 500; }
</style>

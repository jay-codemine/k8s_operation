<template>
  <div class="aiops-page">
    <!-- 顶部 Hero 区域 -->
    <div class="hero-banner">
      <div class="hero-content">
        <div class="hero-text">
          <h1 class="hero-title">智能运维中心</h1>
          <p class="hero-desc">AI 驱动的全栈可观测性平台 · 告警分析 · 日志诊断 · 智能巡检</p>
        </div>
        <div class="hero-stats">
          <div class="hero-stat-item">
            <div class="hero-stat-value">{{ stats.last_health_score || '--' }}</div>
            <div class="hero-stat-label">健康评分</div>
            <div class="hero-stat-badge" :class="stats.last_health_level || 'healthy'">
              {{ levelText(stats.last_health_level) }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 数据概览卡片 -->
    <div class="metrics-row">
      <div class="metric-card" v-for="m in metricsCards" :key="m.key">
        <div class="metric-icon-wrap" :class="m.color">
          <span class="metric-icon">{{ m.icon }}</span>
        </div>
        <div class="metric-body">
          <div class="metric-value">{{ stats[m.key] || 0 }}</div>
          <div class="metric-label">{{ m.label }}</div>
        </div>
      </div>
    </div>

    <!-- 三大功能入口 -->
    <div class="feature-section">
      <div class="section-header">
        <h2 class="section-title"><span class="title-dot"></span>智能运维操作</h2>
      </div>
      <div class="feature-grid">
        <div class="feature-card card-alert" @click="showAlertAnalysis = true">
          <div class="feature-glow"></div>
          <div class="feature-icon-large">⚡</div>
          <div class="feature-info">
            <h3>AI 告警分析</h3>
            <p>智能分析告警事件根因，给出处置建议与 kubectl 命令</p>
          </div>
          <div class="feature-arrow">→</div>
        </div>
        <div class="feature-card card-log" @click="showLogDiagnosis = true">
          <div class="feature-glow"></div>
          <div class="feature-icon-large">📋</div>
          <div class="feature-info">
            <h3>AI 日志诊断</h3>
            <p>基于 Loki 日志的智能异常检测与错误模式识别</p>
          </div>
          <div class="feature-arrow">→</div>
        </div>
        <div class="feature-card card-inspect" @click="triggerInspection">
          <div class="feature-glow"></div>
          <div class="feature-icon-large">🔍</div>
          <div class="feature-info">
            <h3>智能巡检</h3>
            <p>四维度全面检查：集群 · 节点 · 工作负载 · 告警</p>
          </div>
          <div class="feature-arrow">→</div>
        </div>
      </div>
    </div>

    <!-- 巡检报告列表 -->
    <div class="report-section">
      <div class="section-header">
        <h2 class="section-title"><span class="title-dot"></span>巡检报告</h2>
        <div class="header-actions">
          <button class="btn-ghost" @click="loadReports" :disabled="loadingReports">
            <span class="btn-icon">↻</span> 刷新
          </button>
        </div>
      </div>

      <div v-if="loadingReports" class="loading-skeleton">
        <div class="skeleton-item" v-for="i in 3" :key="i"></div>
      </div>
      <div v-else-if="reports.length === 0" class="empty-state-pro">
        <div class="empty-icon">📊</div>
        <h3>暂无巡检报告</h3>
        <p>点击上方"智能巡检"立即生成第一份 AI 巡检报告</p>
      </div>
      <div v-else class="report-grid">
        <div v-for="r in reports" :key="r.id" class="report-card" @click="viewReport(r)">
          <div class="report-card-header">
            <div class="report-score-ring" :class="r.level || 'healthy'">
              <svg viewBox="0 0 36 36" class="score-svg">
                <path class="score-bg" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                <path class="score-fg" :stroke-dasharray="`${r.health_score || 0}, 100`" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
              </svg>
              <div class="score-text">{{ r.health_score || 0 }}</div>
            </div>
            <div class="report-card-meta">
              <div class="report-card-title">巡检报告 #{{ r.id }}</div>
              <div class="report-card-tags">
                <span class="tag" :class="r.type">{{ r.type === 'scheduled' ? '⏰ 定时' : '👤 手动' }}</span>
                <span class="tag status" :class="r.status">{{ statusText(r.status) }}</span>
              </div>
              <div class="report-card-time">{{ formatTime(r.created_at) }}</div>
            </div>
          </div>
          <div class="report-card-footer" v-if="r.status === 'completed'">
            <div class="report-stat-mini">
              <span class="stat-dot red"></span>
              <span>{{ r.findings || 0 }} 问题</span>
            </div>
            <div class="report-stat-mini">
              <span class="stat-dot blue"></span>
              <span>{{ r.suggestions_count || 0 }} 建议</span>
            </div>
            <div class="report-stat-mini">
              <span class="stat-dot green"></span>
              <span>{{ r.duration ? (r.duration / 1000).toFixed(1) + 's' : '--' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI 分析记录 -->
    <div class="record-section">
      <div class="section-header">
        <h2 class="section-title"><span class="title-dot"></span>AI 分析记录</h2>
        <div class="header-actions">
          <div class="filter-tabs">
            <button :class="['tab-btn', { active: recordFilter === '' }]" @click="recordFilter=''; loadRecords()">全部</button>
            <button :class="['tab-btn', { active: recordFilter === 'alert_analysis' }]" @click="recordFilter='alert_analysis'; loadRecords()">告警</button>
            <button :class="['tab-btn', { active: recordFilter === 'log_diagnosis' }]" @click="recordFilter='log_diagnosis'; loadRecords()">日志</button>
            <button :class="['tab-btn', { active: recordFilter === 'inspection' }]" @click="recordFilter='inspection'; loadRecords()">巡检</button>
          </div>
        </div>
      </div>
      <div v-if="records.length === 0" class="empty-state-pro small">
        <p>暂无分析记录</p>
      </div>
      <div v-else class="record-table">
        <div v-for="rec in records" :key="rec.id" class="record-row" @click="viewAnalysis(rec)">
          <div class="record-type-icon" :class="rec.type">{{ typeIcon(rec.type) }}</div>
          <div class="record-main">
            <div class="record-title">{{ rec.title }}</div>
            <div class="record-meta-row">
              <span>{{ formatTime(rec.created_at) }}</span>
              <span v-if="rec.latency_ms" class="meta-latency">⏱ {{ rec.latency_ms }}ms</span>
            </div>
          </div>
          <div class="record-status-badge" :class="rec.status">{{ rec.status === 'success' ? '成功' : '失败' }}</div>
        </div>
      </div>
    </div>

    <!-- ==================== 弹窗：告警分析 ==================== -->
    <div v-if="showAlertAnalysis" class="modal-overlay" @click.self="showAlertAnalysis = false">
      <div class="modal-panel">
        <div class="modal-panel-header">
          <h3>⚡ AI 告警分析</h3>
          <button class="modal-close-btn" @click="showAlertAnalysis = false">✕</button>
        </div>
        <div class="modal-panel-body">
          <div class="form-group">
            <label>告警事件 ID</label>
            <input v-model.number="alertForm.event_id" type="number" placeholder="输入告警事件 ID" class="input-modern" />
          </div>
          <button class="btn-primary-lg" @click="submitAlertAnalysis" :disabled="analyzing">
            {{ analyzing ? '🔄 分析中...' : '🧠 开始 AI 分析' }}
          </button>
          <div v-if="alertResult" class="analysis-result-pro">
            <div class="result-badge-row">
              <span class="severity-pill" :class="alertResult.severity">{{ alertResult.severity }}</span>
              <span class="rule-name">{{ alertResult.rule_name }}</span>
            </div>
            <div class="result-markdown" v-html="renderMarkdown(alertResult.analysis)"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 弹窗：日志诊断 ==================== -->
    <div v-if="showLogDiagnosis" class="modal-overlay" @click.self="showLogDiagnosis = false">
      <div class="modal-panel">
        <div class="modal-panel-header">
          <h3>📋 AI 日志诊断</h3>
          <button class="modal-close-btn" @click="showLogDiagnosis = false">✕</button>
        </div>
        <div class="modal-panel-body">
          <div class="form-row-2">
            <div class="form-group"><label>命名空间</label><input v-model="logForm.namespace" placeholder="如 default" class="input-modern" /></div>
            <div class="form-group"><label>Pod 名称</label><input v-model="logForm.pod" placeholder="如 nginx-xxx" class="input-modern" /></div>
          </div>
          <div class="form-row-2">
            <div class="form-group"><label>时间范围</label>
              <select v-model="logForm.time_range" class="input-modern">
                <option value="5m">最近 5 分钟</option><option value="15m">最近 15 分钟</option>
                <option value="1h">最近 1 小时</option><option value="6h">最近 6 小时</option>
              </select>
            </div>
            <div class="form-group"><label>容器名（可选）</label><input v-model="logForm.container" placeholder="容器名" class="input-modern" /></div>
          </div>
          <div class="form-group">
            <label>LogQL 查询（高级）</label>
            <input v-model="logForm.query" placeholder='{namespace="xxx", pod=~"yyy.*"}' class="input-modern" />
          </div>
          <button class="btn-primary-lg" @click="submitLogDiagnosis" :disabled="diagnosing">
            {{ diagnosing ? '🔄 诊断中...' : '🧠 开始 AI 诊断' }}
          </button>
          <div v-if="logResult" class="analysis-result-pro">
            <div class="result-badge-row">
              <span class="severity-pill" :class="logResult.severity">{{ logResult.severity }}</span>
              <span>日志行数: {{ logResult.log_lines }} | 错误: {{ logResult.error_count }}</span>
            </div>
            <div class="result-markdown" v-html="renderMarkdown(logResult.analysis)"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 弹窗：巡检报告详情 ==================== -->
    <div v-if="showReportDetail" class="modal-overlay" @click.self="showReportDetail = false">
      <div class="modal-panel modal-wide">
        <div class="modal-panel-header">
          <h3>🔍 巡检报告详情 #{{ currentReport?.id }}</h3>
          <div class="modal-header-actions">
            <button class="btn-action" @click="handleExport" title="导出 Markdown">📥 导出</button>
            <button class="btn-action" @click="handleDownload" title="下载报告文件">⬇️ 下载</button>
            <button class="btn-action accent" @click="showNotifyPanel = true" title="发送到通知渠道">📤 发送通知</button>
            <button class="modal-close-btn" @click="showReportDetail = false">✕</button>
          </div>
        </div>
        <div class="modal-panel-body" v-if="currentReport">
          <!-- 报告评分头部 -->
          <div class="report-detail-hero">
            <div class="detail-score-circle" :class="currentReport.level">
              <div class="score-number">{{ currentReport.health_score }}</div>
              <div class="score-unit">/100</div>
            </div>
            <div class="detail-hero-info">
              <div class="detail-level-badge" :class="currentReport.level">{{ levelText(currentReport.level) }}</div>
              <div class="detail-summary">{{ currentReport.summary }}</div>
              <div class="detail-meta-row">
                <span>{{ currentReport.type === 'scheduled' ? '⏰ 定时巡检' : '👤 手动巡检' }}</span>
                <span>⏱ {{ currentReport.duration ? (currentReport.duration/1000).toFixed(1)+'s' : '--' }}</span>
                <span>📅 {{ formatTime(currentReport.created_at) }}</span>
              </div>
            </div>
          </div>

          <!-- 四维度数据 -->
          <div class="dimension-grid" v-if="reportDetails">
            <div class="dim-card" v-for="dim in dimensions" :key="dim.key">
              <div class="dim-header">
                <span class="dim-icon">{{ dim.icon }}</span>
                <span class="dim-title">{{ dim.title }}</span>
              </div>
              <div class="dim-bar-wrap">
                <div class="dim-bar" :style="{ width: dimPercent(dim) + '%' }" :class="dimStatus(dim)"></div>
              </div>
              <div class="dim-numbers">
                <span class="dim-healthy">{{ dim.healthy(reportDetails) }}</span> / <span class="dim-total">{{ dim.total(reportDetails) }}</span>
              </div>
            </div>
          </div>

          <!-- AI 分析内容 -->
          <div v-if="currentReport.ai_analysis" class="ai-analysis-block">
            <h4>🤖 AI 智能分析</h4>
            <div class="result-markdown" v-html="renderMarkdown(currentReport.ai_analysis)"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 弹窗：发送通知 ==================== -->
    <div v-if="showNotifyPanel" class="modal-overlay" @click.self="showNotifyPanel = false">
      <div class="modal-panel modal-sm">
        <div class="modal-panel-header">
          <h3>📤 发送巡检报告</h3>
          <button class="modal-close-btn" @click="showNotifyPanel = false">✕</button>
        </div>
        <div class="modal-panel-body">
          <p class="notify-desc">选择通知渠道，将巡检报告发送到钉钉、飞书、企业微信等平台</p>
          <div v-if="channelsLoading" class="loading-text">加载渠道中...</div>
          <div v-else-if="channels.length === 0" class="empty-channels">
            <p>暂无可用通知渠道，请先在<strong>监控 → 通知渠道</strong>中配置</p>
          </div>
          <div v-else class="channel-list">
            <label v-for="ch in channels" :key="ch.id" class="channel-item" :class="{ selected: selectedChannels.includes(ch.id) }">
              <input type="checkbox" :value="ch.id" v-model="selectedChannels" />
              <span class="channel-type-icon">{{ channelIcon(ch.type) }}</span>
              <span class="channel-name">{{ ch.name }}</span>
              <span class="channel-type-label">{{ ch.type }}</span>
            </label>
          </div>
          <div class="notify-actions">
            <button class="btn-primary-lg" @click="doNotify" :disabled="notifying || selectedChannels.length === 0">
              {{ notifying ? '发送中...' : `发送到 ${selectedChannels.length} 个渠道` }}
            </button>
          </div>
          <div v-if="notifyResults.length > 0" class="notify-results">
            <div v-for="r in notifyResults" :key="r.channel_id" class="notify-result-item" :class="{ success: r.success, fail: !r.success }">
              <span>{{ channelIcon(r.channel_type) }} {{ r.channel_name }}</span>
              <span>{{ r.success ? '✅ 发送成功' : '❌ ' + r.error }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== 弹窗：分析详情 ==================== -->
    <div v-if="showAnalysisDetail" class="modal-overlay" @click.self="showAnalysisDetail = false">
      <div class="modal-panel">
        <div class="modal-panel-header">
          <h3>{{ currentAnalysis?.title || 'AI 分析详情' }}</h3>
          <button class="modal-close-btn" @click="showAnalysisDetail = false">✕</button>
        </div>
        <div class="modal-panel-body">
          <div v-if="currentAnalysis" class="result-markdown" v-html="renderMarkdown(currentAnalysis.result)"></div>
        </div>
      </div>
    </div>

    <!-- 导出预览弹窗 -->
    <div v-if="showExportPreview" class="modal-overlay" @click.self="showExportPreview = false">
      <div class="modal-panel modal-wide">
        <div class="modal-panel-header">
          <h3>📥 报告导出预览</h3>
          <div class="modal-header-actions">
            <button class="btn-action" @click="copyExportContent">📋 复制</button>
            <button class="btn-action accent" @click="handleDownload">⬇️ 下载文件</button>
            <button class="modal-close-btn" @click="showExportPreview = false">✕</button>
          </div>
        </div>
        <div class="modal-panel-body">
          <pre class="export-preview-content">{{ exportContent }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import aiopsApi from '@/api/platform/aiops'

// ===== 状态 =====
const stats = ref({})
const reports = ref([])
const records = ref([])
const loadingReports = ref(false)
const recordFilter = ref('')

// 弹窗
const showAlertAnalysis = ref(false)
const showLogDiagnosis = ref(false)
const showReportDetail = ref(false)
const showAnalysisDetail = ref(false)
const showNotifyPanel = ref(false)
const showExportPreview = ref(false)

// 告警分析
const alertForm = reactive({ event_id: null })
const analyzing = ref(false)
const alertResult = ref(null)

// 日志诊断
const logForm = reactive({ namespace: '', pod: '', container: '', time_range: '15m', query: '' })
const diagnosing = ref(false)
const logResult = ref(null)

// 详情
const currentReport = ref(null)
const currentAnalysis = ref(null)
const exportContent = ref('')

// 通知
const channels = ref([])
const channelsLoading = ref(false)
const selectedChannels = ref([])
const notifying = ref(false)
const notifyResults = ref([])

// 概览卡片
const metricsCards = [
  { key: 'today_analysis', label: '今日分析', icon: '🧠', color: 'purple' },
  { key: 'firing_alerts', label: '活跃告警', icon: '🔥', color: 'red' },
  { key: 'week_analysis', label: '本周分析', icon: 'dashboard', color: 'blue' },
  { key: 'total_analysis', label: '累计分析', icon: 'dashboard', color: 'green' },
]

// 四维度配置
const dimensions = [
  { key: 'cluster', icon: '☁️', title: '集群', total: d => d.clusters_total, healthy: d => d.clusters_healthy },
  { key: 'node', icon: 'desktop', title: '节点', total: d => d.nodes_total, healthy: d => d.nodes_ready },
  { key: 'workload', icon: 'settings', title: '工作负载', total: d => d.workloads_total, healthy: d => d.workloads_healthy },
  { key: 'alert', icon: '🚨', title: '告警', total: d => d.alerts_firing, healthy: d => 0 },
]

const reportDetails = computed(() => {
  if (!currentReport.value?.details) return null
  try { return JSON.parse(currentReport.value.details) } catch { return null }
})

// ===== 生命周期 =====
onMounted(() => {
  loadDashboard()
  loadReports()
  loadRecords()
})

// ===== API 调用 =====
async function loadDashboard() {
  try {
    const res = await aiopsApi.getDashboard()
    if (res.code === 0) stats.value = res.data || {}
  } catch (e) { console.error(e) }
}

async function loadReports() {
  loadingReports.value = true
  try {
    const res = await aiopsApi.getInspectionList({ page: 1, page_size: 12 })
    if (res.code === 0) reports.value = res.data?.list || []
  } catch (e) { console.error(e) }
  loadingReports.value = false
}

async function loadRecords() {
  try {
    const params = { page: 1, page_size: 20 }
    if (recordFilter.value) params.type = recordFilter.value
    const res = await aiopsApi.getRecords(params)
    if (res.code === 0) records.value = res.data?.list || []
  } catch (e) { console.error(e) }
}

async function triggerInspection() {
  if (!confirm('确定要立即执行全量智能巡检？（覆盖集群/节点/工作负载/告警四维度）')) return
  try {
    const res = await aiopsApi.runInspection()
    if (res.code === 0) {
      alert('✅ 巡检已提交执行，AI 分析中... 请稍后刷新查看报告')
      setTimeout(loadReports, 5000)
    } else {
      alert(res.msg || '触发失败')
    }
  } catch (e) { alert('触发巡检失败: ' + e.message) }
}

async function submitAlertAnalysis() {
  if (!alertForm.event_id) return alert('请输入告警事件 ID')
  analyzing.value = true; alertResult.value = null
  try {
    const res = await aiopsApi.analyzeAlert({ event_id: alertForm.event_id })
    if (res.code === 0) alertResult.value = res.data
    else alert(res.msg || '分析失败')
  } catch (e) { alert('分析失败: ' + e.message) }
  analyzing.value = false
}

async function submitLogDiagnosis() {
  if (!logForm.namespace && !logForm.pod && !logForm.query) return alert('请至少填写命名空间或 Pod 名称')
  diagnosing.value = true; logResult.value = null
  try {
    const res = await aiopsApi.diagnoseLogs(logForm)
    if (res.code === 0) logResult.value = res.data
    else alert(res.msg || '诊断失败')
  } catch (e) { alert('诊断失败: ' + e.message) }
  diagnosing.value = false
}

async function viewReport(r) {
  try {
    const res = await aiopsApi.getInspectionDetail(r.id)
    if (res.code === 0) { currentReport.value = res.data; showReportDetail.value = true }
  } catch (e) { console.error(e) }
}

function viewAnalysis(rec) { currentAnalysis.value = rec; showAnalysisDetail.value = true }

// ===== 导出/下载/通知 =====
async function handleExport() {
  if (!currentReport.value) return
  try {
    const res = await aiopsApi.exportReport(currentReport.value.id)
    if (res.code === 0) { exportContent.value = res.data.content; showExportPreview.value = true }
  } catch (e) { alert('导出失败: ' + e.message) }
}

function handleDownload() {
  if (!currentReport.value) return
  aiopsApi.downloadReport(currentReport.value.id)
}

function copyExportContent() {
  navigator.clipboard.writeText(exportContent.value)
  alert('已复制到剪贴板')
}

async function loadChannels() {
  channelsLoading.value = true
  try {
    const res = await aiopsApi.getNotifyChannels()
    if (res.code === 0) channels.value = res.data || []
  } catch (e) { console.error(e) }
  channelsLoading.value = false
}

// watch showNotifyPanel
import { watch } from 'vue'
watch(showNotifyPanel, (v) => {
  if (v) { loadChannels(); selectedChannels.value = []; notifyResults.value = [] }
})

async function doNotify() {
  if (!currentReport.value || selectedChannels.value.length === 0) return
  notifying.value = true; notifyResults.value = []
  try {
    const res = await aiopsApi.notifyReport(currentReport.value.id, selectedChannels.value)
    if (res.code === 0) notifyResults.value = res.data || []
    else alert(res.msg || '发送失败')
  } catch (e) { alert('发送失败: ' + e.message) }
  notifying.value = false
}

// ===== 工具方法 =====
function formatTime(ts) {
  if (!ts) return '--'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}
function statusText(s) { return { running: '运行中', completed: '已完成', failed: '失败' }[s] || s }
function levelText(l) { return { healthy: '健康', warning: '警告', critical: '严重' }[l] || '未知' }
function typeIcon(t) { return { alert_analysis: 'thunderbolt', log_diagnosis: 'file', inspection: 'search' }[t] || '🧠' }
function channelIcon(t) { return { dingtalk: '🔷', feishu: '🟦', wechat: '🟩', email: '📧', webhook: '🔗' }[t] || 'wifi' }

function dimPercent(dim) {
  if (!reportDetails.value) return 0
  const t = dim.total(reportDetails.value)
  if (t === 0) return 100
  if (dim.key === 'alert') return Math.min(t * 10, 100) // 告警越多越红
  return (dim.healthy(reportDetails.value) / t) * 100
}
function dimStatus(dim) {
  const p = dimPercent(dim)
  if (dim.key === 'alert') return p > 50 ? 'critical' : p > 20 ? 'warning' : 'healthy'
  return p >= 80 ? 'healthy' : p >= 60 ? 'warning' : 'critical'
}

function renderMarkdown(text) {
  if (!text) return ''
  return text
    .replace(/^### (.*$)/gm, '<h4>$1</h4>')
    .replace(/^## (.*$)/gm, '<h3>$1</h3>')
    .replace(/^# (.*$)/gm, '<h2>$1</h2>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/^- (.*$)/gm, '<li>$1</li>')
    .replace(/(<li>[\s\S]*?<\/li>)/g, '<ul>$1</ul>')
    .replace(/\n/g, '<br>')
}
</script>

<style scoped>
/* ==================== 全局 ==================== */
.aiops-page { padding: 0; max-width: 1440px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }

/* ==================== Hero Banner ==================== */
.hero-banner { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border-radius: 0 0 24px 24px; padding: 32px 36px; margin-bottom: 24px; position: relative; overflow: hidden; }
.hero-banner::before { content: ''; position: absolute; inset: 0; background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.05'%3E%3Ccircle cx='30' cy='30' r='2'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E"); }
.hero-content { display: flex; justify-content: space-between; align-items: center; position: relative; z-index: 1; }
.hero-title { font-size: 26px; font-weight: 700; color: #fff; margin: 0 0 8px; }
.hero-desc { font-size: 14px; color: rgba(255,255,255,.8); margin: 0; }
.hero-stats { display: flex; align-items: center; gap: 16px; }
.hero-stat-value { font-size: 48px; font-weight: 800; color: #fff; line-height: 1; }
.hero-stat-label { font-size: 13px; color: rgba(255,255,255,.7); margin-top: 4px; }
.hero-stat-badge { display: inline-block; margin-top: 6px; padding: 3px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.hero-stat-badge.healthy { background: rgba(16,185,129,.3); color: #6ee7b7; }
.hero-stat-badge.warning { background: rgba(245,158,11,.3); color: #fcd34d; }
.hero-stat-badge.critical { background: rgba(239,68,68,.3); color: #fca5a5; }

/* ==================== Metrics Row ==================== */
.metrics-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; padding: 0 24px; margin-bottom: 28px; }
.metric-card { display: flex; align-items: center; gap: 14px; padding: 18px 20px; background: #fff; border-radius: 14px; box-shadow: 0 2px 12px rgba(0,0,0,.05); transition: all .2s; }
.metric-card:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,.08); }
.metric-icon-wrap { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 22px; }
.metric-icon-wrap.purple { background: linear-gradient(135deg, #eef2ff, #e0e7ff); }
.metric-icon-wrap.red { background: linear-gradient(135deg, #fef2f2, #fee2e2); }
.metric-icon-wrap.blue { background: linear-gradient(135deg, #eff6ff, #dbeafe); }
.metric-icon-wrap.green { background: linear-gradient(135deg, #ecfdf5, #d1fae5); }
.metric-value { font-size: 24px; font-weight: 700; color: #1e293b; }
.metric-label { font-size: 12px; color: #94a3b8; margin-top: 2px; }

/* ==================== Section Header ==================== */
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; padding: 0 24px; }
.section-title { font-size: 17px; font-weight: 700; color: #1e293b; display: flex; align-items: center; gap: 8px; margin: 0; }
.title-dot { width: 4px; height: 18px; border-radius: 2px; background: linear-gradient(180deg, #667eea, #764ba2); }

/* ==================== Feature Grid ==================== */
.feature-section { margin-bottom: 32px; }
.feature-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; padding: 0 24px; }
.feature-card { position: relative; padding: 24px; border-radius: 16px; background: #fff; border: 1px solid #f1f5f9; cursor: pointer; transition: all .3s; display: flex; align-items: center; gap: 16px; overflow: hidden; }
.feature-card:hover { transform: translateY(-3px); box-shadow: 0 12px 32px rgba(0,0,0,.1); }
.feature-card:hover .feature-glow { opacity: 1; }
.feature-glow { position: absolute; inset: 0; opacity: 0; transition: opacity .3s; }
.card-alert .feature-glow { background: linear-gradient(135deg, rgba(245,158,11,.05), rgba(239,68,68,.05)); }
.card-log .feature-glow { background: linear-gradient(135deg, rgba(59,130,246,.05), rgba(99,102,241,.05)); }
.card-inspect .feature-glow { background: linear-gradient(135deg, rgba(16,185,129,.05), rgba(6,182,212,.05)); }
.feature-icon-large { font-size: 36px; flex-shrink: 0; }
.feature-info h3 { font-size: 15px; font-weight: 600; color: #1e293b; margin: 0 0 4px; }
.feature-info p { font-size: 13px; color: #64748b; margin: 0; line-height: 1.5; }
.feature-arrow { font-size: 20px; color: #cbd5e1; margin-left: auto; transition: transform .2s; }
.feature-card:hover .feature-arrow { transform: translateX(4px); color: #4f46e5; }

/* ==================== Report Section ==================== */
.report-section { margin-bottom: 32px; }
.report-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px; padding: 0 24px; }
.report-card { background: #fff; border-radius: 14px; padding: 18px; border: 1px solid #f1f5f9; cursor: pointer; transition: all .2s; }
.report-card:hover { border-color: #e0e7ff; box-shadow: 0 8px 24px rgba(99,102,241,.08); }
.report-card-header { display: flex; align-items: center; gap: 14px; }
.report-score-ring { position: relative; width: 52px; height: 52px; flex-shrink: 0; }
.score-svg { width: 100%; height: 100%; transform: rotate(-90deg); }
.score-bg { fill: none; stroke: #f1f5f9; stroke-width: 3; }
.score-fg { fill: none; stroke-width: 3; stroke-linecap: round; transition: stroke-dasharray .6s; }
.report-score-ring.healthy .score-fg { stroke: #10b981; }
.report-score-ring.warning .score-fg { stroke: #f59e0b; }
.report-score-ring.critical .score-fg { stroke: #ef4444; }
.score-text { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 700; color: #1e293b; }
.report-card-meta { flex: 1; }
.report-card-title { font-size: 14px; font-weight: 600; color: #1e293b; }
.report-card-tags { display: flex; gap: 6px; margin-top: 4px; }
.tag { padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 500; }
.tag.scheduled { background: #eff6ff; color: #3b82f6; }
.tag.manual { background: #f5f3ff; color: #7c3aed; }
.tag.status.completed { background: #ecfdf5; color: #059669; }
.tag.status.running { background: #eff6ff; color: #2563eb; }
.tag.status.failed { background: #fef2f2; color: #dc2626; }
.report-card-time { font-size: 11px; color: #94a3b8; margin-top: 4px; }
.report-card-footer { display: flex; gap: 16px; margin-top: 12px; padding-top: 12px; border-top: 1px solid #f8fafc; }
.report-stat-mini { display: flex; align-items: center; gap: 4px; font-size: 12px; color: #64748b; }
.stat-dot { width: 6px; height: 6px; border-radius: 50%; }
.stat-dot.red { background: #ef4444; }
.stat-dot.blue { background: #3b82f6; }
.stat-dot.green { background: #10b981; }

/* ==================== Records ==================== */
.record-section { margin-bottom: 32px; }
.header-actions { display: flex; align-items: center; gap: 12px; }
.filter-tabs { display: flex; gap: 4px; background: #f8fafc; border-radius: 8px; padding: 3px; }
.tab-btn { padding: 5px 12px; border: none; background: transparent; border-radius: 6px; font-size: 12px; color: #64748b; cursor: pointer; transition: all .15s; }
.tab-btn.active { background: #fff; color: #4f46e5; font-weight: 600; box-shadow: 0 1px 3px rgba(0,0,0,.08); }
.record-table { padding: 0 24px; display: flex; flex-direction: column; gap: 6px; }
.record-row { display: flex; align-items: center; gap: 12px; padding: 12px 16px; background: #fff; border-radius: 10px; border: 1px solid #f8fafc; cursor: pointer; transition: all .15s; }
.record-row:hover { background: #fafbff; border-color: #e0e7ff; }
.record-type-icon { width: 34px; height: 34px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 16px; }
.record-type-icon.alert_analysis { background: #fef3c7; }
.record-type-icon.log_diagnosis { background: #dbeafe; }
.record-type-icon.inspection { background: #dcfce7; }
.record-main { flex: 1; }
.record-title { font-size: 13px; font-weight: 500; color: #1e293b; }
.record-meta-row { display: flex; gap: 12px; font-size: 11px; color: #94a3b8; margin-top: 2px; }
.meta-latency { color: #6366f1; }
.record-status-badge { padding: 3px 10px; border-radius: 10px; font-size: 11px; font-weight: 500; }
.record-status-badge.success { background: #ecfdf5; color: #059669; }
.record-status-badge.failed { background: #fef2f2; color: #dc2626; }

/* ==================== Buttons ==================== */
.btn-ghost { display: flex; align-items: center; gap: 4px; padding: 6px 14px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; font-size: 12px; color: #64748b; cursor: pointer; transition: all .15s; }
.btn-ghost:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-icon { font-size: 14px; }
.btn-primary-lg { width: 100%; padding: 12px; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; border: none; border-radius: 10px; font-size: 14px; font-weight: 600; cursor: pointer; transition: opacity .2s; margin-top: 12px; }
.btn-primary-lg:hover { opacity: .9; }
.btn-primary-lg:disabled { opacity: .5; cursor: not-allowed; }
.btn-action { padding: 6px 12px; border: 1px solid #e2e8f0; border-radius: 6px; background: #fff; font-size: 12px; cursor: pointer; transition: all .15s; }
.btn-action:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-action.accent { background: #4f46e5; color: #fff; border-color: #4f46e5; }
.btn-action.accent:hover { background: #4338ca; }

/* ==================== Modals ==================== */
.modal-overlay { position: fixed; inset: 0; background: rgba(15,23,42,.6); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-panel { background: #fff; border-radius: 20px; width: 92%; max-width: 640px; max-height: 88vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,.2); }
.modal-panel.modal-wide { max-width: 860px; }
.modal-panel.modal-sm { max-width: 480px; }
.modal-panel-header { display: flex; justify-content: space-between; align-items: center; padding: 18px 24px; border-bottom: 1px solid #f1f5f9; position: sticky; top: 0; background: #fff; z-index: 1; border-radius: 20px 20px 0 0; }
.modal-panel-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.modal-header-actions { display: flex; gap: 8px; align-items: center; }
.modal-close-btn { background: none; border: none; font-size: 18px; cursor: pointer; color: #94a3b8; padding: 4px 8px; border-radius: 6px; }
.modal-close-btn:hover { background: #f1f5f9; color: #1e293b; }
.modal-panel-body { padding: 24px; }

/* ==================== Forms ==================== */
.form-group { margin-bottom: 14px; }
.form-group label { display: block; font-size: 12px; font-weight: 600; color: #374151; margin-bottom: 6px; text-transform: uppercase; letter-spacing: .3px; }
.input-modern { width: 100%; padding: 10px 14px; border: 1.5px solid #e2e8f0; border-radius: 10px; font-size: 14px; box-sizing: border-box; transition: all .15s; background: #fafbff; }
.input-modern:focus { outline: none; border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,.12); background: #fff; }
.form-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

/* ==================== Analysis Results ==================== */
.analysis-result-pro { margin-top: 20px; padding: 18px; border-radius: 12px; background: #fafbff; border: 1px solid #e0e7ff; }
.result-badge-row { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
.severity-pill { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; color: #fff; }
.severity-pill.critical { background: #ef4444; }
.severity-pill.warning { background: #f59e0b; }
.severity-pill.info { background: #3b82f6; }
.rule-name { font-size: 13px; font-weight: 500; color: #374151; }
.result-markdown { font-size: 13px; line-height: 1.8; color: #374151; }
.result-markdown :deep(h3) { font-size: 15px; margin: 16px 0 8px; color: #1e293b; font-weight: 600; }
.result-markdown :deep(h4) { font-size: 14px; margin: 12px 0 6px; }
.result-markdown :deep(code) { background: #e0e7ff; padding: 2px 6px; border-radius: 4px; font-size: 12px; color: #4338ca; }
.result-markdown :deep(li) { margin-left: 16px; margin-bottom: 4px; }
.result-markdown :deep(ul) { list-style: disc; padding-left: 12px; }
.result-markdown :deep(strong) { color: #1e293b; }

/* ==================== Report Detail ==================== */
.report-detail-hero { display: flex; align-items: center; gap: 24px; margin-bottom: 24px; padding: 20px; background: linear-gradient(135deg, #fafbff, #f0f4ff); border-radius: 16px; }
.detail-score-circle { width: 88px; height: 88px; border-radius: 50%; display: flex; flex-direction: column; align-items: center; justify-content: center; flex-shrink: 0; }
.detail-score-circle.healthy { background: linear-gradient(135deg, #10b981, #059669); }
.detail-score-circle.warning { background: linear-gradient(135deg, #f59e0b, #d97706); }
.detail-score-circle.critical { background: linear-gradient(135deg, #ef4444, #dc2626); }
.score-number { font-size: 28px; font-weight: 800; color: #fff; line-height: 1; }
.score-unit { font-size: 11px; color: rgba(255,255,255,.8); }
.detail-hero-info { flex: 1; }
.detail-level-badge { display: inline-block; padding: 3px 10px; border-radius: 10px; font-size: 12px; font-weight: 600; margin-bottom: 6px; }
.detail-level-badge.healthy { background: #ecfdf5; color: #059669; }
.detail-level-badge.warning { background: #fffbeb; color: #d97706; }
.detail-level-badge.critical { background: #fef2f2; color: #dc2626; }
.detail-summary { font-size: 14px; color: #374151; margin-bottom: 8px; font-weight: 500; }
.detail-meta-row { display: flex; gap: 16px; font-size: 12px; color: #94a3b8; }

/* Dimension Grid */
.dimension-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 24px; }
.dim-card { padding: 14px; background: #fff; border-radius: 12px; border: 1px solid #f1f5f9; }
.dim-header { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; }
.dim-icon { font-size: 16px; }
.dim-title { font-size: 12px; font-weight: 600; color: #64748b; }
.dim-bar-wrap { height: 6px; border-radius: 3px; background: #f1f5f9; overflow: hidden; margin-bottom: 6px; }
.dim-bar { height: 100%; border-radius: 3px; transition: width .6s ease; }
.dim-bar.healthy { background: linear-gradient(90deg, #10b981, #34d399); }
.dim-bar.warning { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.dim-bar.critical { background: linear-gradient(90deg, #ef4444, #f87171); }
.dim-numbers { font-size: 12px; color: #64748b; }
.dim-healthy { font-weight: 700; color: #10b981; }
.dim-total { font-weight: 600; }

/* AI Analysis Block */
.ai-analysis-block { padding: 20px; background: #fff; border-radius: 14px; border: 1px solid #e0e7ff; }
.ai-analysis-block h4 { font-size: 15px; font-weight: 600; margin: 0 0 12px; color: #1e293b; }

/* ==================== Notify Panel ==================== */
.notify-desc { font-size: 13px; color: #64748b; margin: 0 0 16px; }
.channel-list { display: flex; flex-direction: column; gap: 8px; max-height: 280px; overflow-y: auto; }
.channel-item { display: flex; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 10px; border: 1.5px solid #f1f5f9; cursor: pointer; transition: all .15s; }
.channel-item:hover { border-color: #e0e7ff; background: #fafbff; }
.channel-item.selected { border-color: #667eea; background: #eef2ff; }
.channel-item input[type="checkbox"] { accent-color: #667eea; }
.channel-type-icon { font-size: 18px; }
.channel-name { font-size: 13px; font-weight: 500; color: #1e293b; flex: 1; }
.channel-type-label { font-size: 11px; color: #94a3b8; padding: 2px 6px; background: #f8fafc; border-radius: 4px; }
.notify-actions { margin-top: 16px; }
.notify-results { margin-top: 14px; display: flex; flex-direction: column; gap: 6px; }
.notify-result-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-radius: 8px; font-size: 12px; }
.notify-result-item.success { background: #ecfdf5; color: #059669; }
.notify-result-item.fail { background: #fef2f2; color: #dc2626; }

/* ==================== States ==================== */
.loading-skeleton { padding: 0 24px; display: flex; flex-direction: column; gap: 12px; }
.skeleton-item { height: 80px; border-radius: 14px; background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%); background-size: 200% 100%; animation: shimmer 1.5s infinite; }
@keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }
.empty-state-pro { text-align: center; padding: 40px 20px; }
.empty-state-pro .empty-icon { font-size: 40px; margin-bottom: 12px; }
.empty-state-pro h3 { font-size: 15px; color: #64748b; margin: 0 0 6px; }
.empty-state-pro p { font-size: 13px; color: #94a3b8; margin: 0; }
.empty-state-pro.small { padding: 20px; }
.loading-text { text-align: center; padding: 20px; color: #94a3b8; font-size: 13px; }
.empty-channels { text-align: center; padding: 20px; color: #94a3b8; font-size: 13px; }

/* Export Preview */
.export-preview-content { white-space: pre-wrap; font-size: 12px; line-height: 1.7; color: #374151; background: #fafbff; padding: 16px; border-radius: 10px; border: 1px solid #e0e7ff; max-height: 60vh; overflow-y: auto; font-family: 'JetBrains Mono', monospace; }

/* ==================== Responsive ==================== */
@media (max-width: 1024px) {
  .metrics-row { grid-template-columns: repeat(2, 1fr); }
  .feature-grid { grid-template-columns: 1fr; }
  .dimension-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .hero-content { flex-direction: column; text-align: center; gap: 16px; }
  .metrics-row { grid-template-columns: 1fr 1fr; }
  .report-grid { grid-template-columns: 1fr; }
  .form-row-2 { grid-template-columns: 1fr; }
}
</style>

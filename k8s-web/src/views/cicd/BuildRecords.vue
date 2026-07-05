<template>
  <div class="build-records-page">
    <!-- 顶部 Banner -->
    <div class="page-banner">
      <div class="banner-inner">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
          </div>
          <div>
            <h1 class="banner-title">流水线构建中心</h1>
            <p class="banner-desc">全局构建运行记录，实时追踪构建状态、性能分析与趋势洞察</p>
          </div>
        </div>
        <div class="banner-actions">
          <button class="btn-action" @click="exportRecords" :disabled="exporting">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span>{{ exporting ? '导出中...' : '导出 CSV' }}</span>
          </button>
          <button class="btn-action refresh" @click="loadAll" :disabled="loading">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            <span>刷新</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 数据指标面板 -->
    <div class="metrics-panel">
      <div class="metric-card primary">
        <div class="metric-header">
          <span class="metric-label">总构建数</span>
          <div class="metric-icon-sm"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/></svg></div>
        </div>
        <div class="metric-value">{{ statsData.total_builds || 0 }}</div>
        <div class="metric-sub">今日 <strong>{{ statsData.today_builds || 0 }}</strong> · 本周 <strong>{{ statsData.week_builds || 0 }}</strong></div>
      </div>
      <div class="metric-card success">
        <div class="metric-header">
          <span class="metric-label">成功率</span>
          <div class="metric-icon-sm"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></div>
        </div>
        <div class="metric-value">{{ (statsData.success_rate || 0).toFixed(1) }}<span class="metric-unit">%</span></div>
        <div class="metric-sub">成功 <strong>{{ statsData.success_builds || 0 }}</strong> · 失败 <strong>{{ statsData.failed_builds || 0 }}</strong></div>
      </div>
      <div class="metric-card warning">
        <div class="metric-header">
          <span class="metric-label">平均耗时</span>
          <div class="metric-icon-sm"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg></div>
        </div>
        <div class="metric-value">{{ formatDuration(Math.round(statsData.avg_duration || 0)) }}</div>
        <div class="metric-sub">运行中 <strong>{{ statsData.running_builds || 0 }}</strong></div>
      </div>
      <div class="metric-card info">
        <div class="metric-header">
          <span class="metric-label">构建趋势</span>
          <div class="metric-icon-sm"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg></div>
        </div>
        <div class="trend-mini">
          <div v-for="(t, i) in trendData.slice(-7)" :key="i" class="trend-bar-wrap">
            <div class="trend-bar" :style="{ height: trendBarHeight(t.total) + '%' }">
              <div class="trend-bar-success" :style="{ height: trendBarSuccessRatio(t) + '%' }"></div>
            </div>
          </div>
        </div>
        <div class="metric-sub">近7天趋势</div>
      </div>
    </div>

    <!-- 筛选工具栏 -->
    <div class="filter-toolbar">
      <div class="filter-left">
        <div class="status-tabs">
          <button :class="['tab-btn', { active: statusFilter === '' }]" @click="setFilter('')">
            全部 <span class="tab-count">{{ total }}</span>
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'running' }]" @click="setFilter('running')">
            <span class="dot running"></span> 运行中
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'success' }]" @click="setFilter('success')">
            <span class="dot success"></span> 成功
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'failed' }]" @click="setFilter('failed')">
            <span class="dot failed"></span> 失败
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'aborted' }]" @click="setFilter('aborted')">
            <span class="dot aborted"></span> 已中止
          </button>
        </div>
      </div>
      <div class="filter-right">
        <select v-model="pipelineFilter" class="filter-select" @change="handleFilterChange">
          <option value="">全部流水线</option>
          <option v-for="p in pipelines" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <div class="search-input-wrap">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="keyword" placeholder="搜索镜像、错误信息..." @input="debouncedLoad" />
          <button v-if="keyword" class="clear-btn" @click="keyword = ''; loadRecords()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrapper">
      <div v-if="loading" class="loading-state">
        <div class="loader"><div class="dot-bounce"></div><div class="dot-bounce"></div><div class="dot-bounce"></div></div>
        <span>加载构建记录中...</span>
      </div>
      <table v-else-if="records.length > 0" class="data-table">
        <thead>
          <tr>
            <th class="col-id">构建号</th>
            <th class="col-pipeline">流水线</th>
            <th class="col-status">状态</th>
            <th class="col-trigger">触发方式</th>
            <th class="col-branch">分支</th>
            <th class="col-image">产出镜像</th>
            <th class="col-duration">耗时</th>
            <th class="col-time">构建时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in records" :key="r.id" @click="goToDetail(r)" class="row-hover">
            <td class="col-id">
              <span class="build-id">#{{ r.build_number || '-' }}</span>
            </td>
            <td class="col-pipeline">
              <div class="pipeline-cell">
                <span class="pipeline-name">{{ r.pipeline_name || `Pipeline#${r.pipeline_id}` }}</span>
              </div>
            </td>
            <td class="col-status">
              <span :class="['status-tag', r.status]">
                <span class="status-dot"></span>
                {{ statusText(r.status) }}
              </span>
            </td>
            <td class="col-trigger">
              <span class="trigger-badge">{{ triggerText(r.trigger_type) }}</span>
            </td>
            <td class="col-branch">
              <span class="branch-tag" v-if="r.git_branch">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="branch-icon"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>
                {{ r.git_branch }}
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="col-image">
              <span class="image-text" :title="r.image_url">{{ shortImage(r.image_url) }}</span>
            </td>
            <td class="col-duration">
              <span class="duration-text">{{ formatDuration(r.duration_sec) }}</span>
            </td>
            <td class="col-time">
              <span class="time-text">{{ formatTime(r.created_at) }}</span>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        </div>
        <h3>暂无构建记录</h3>
        <p>触发流水线构建后，构建记录将在此展示</p>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination-bar">
      <div class="page-info">
        共 <strong>{{ total }}</strong> 条记录
      </div>
      <div class="page-controls">
        <select v-model="pageSize" class="page-size-select" @change="page = 1; loadRecords()">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
        <div class="page-nav">
          <button class="page-btn" :disabled="page <= 1" @click="page--; loadRecords()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <template v-for="p in visiblePages" :key="p">
            <button v-if="p === '...'" class="page-btn ellipsis" disabled>...</button>
            <button v-else :class="['page-btn', { active: p === page }]" @click="page = p; loadRecords()">{{ p }}</button>
          </template>
          <button class="page-btn" :disabled="page >= totalPages" @click="page++; loadRecords()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getPipelines } from '@/api/platform/pipeline'
import { getBuildRecords, getBuildStats } from '@/api/cicd'

const router = useRouter()

const records = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const exporting = ref(false)
const keyword = ref('')
const statusFilter = ref('')
const pipelineFilter = ref('')
const pipelines = ref([])
const statsData = ref({})
const trendData = ref([])

const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

const visiblePages = computed(() => {
  const tp = totalPages.value
  const cp = page.value
  if (tp <= 7) return Array.from({ length: tp }, (_, i) => i + 1)
  const pages = []
  pages.push(1)
  if (cp > 3) pages.push('...')
  for (let i = Math.max(2, cp - 1); i <= Math.min(tp - 1, cp + 1); i++) {
    pages.push(i)
  }
  if (cp < tp - 2) pages.push('...')
  pages.push(tp)
  return pages
})

const loadPipelines = async () => {
  try {
    const res = await getPipelines({ page: 1, page_size: 200 })
    pipelines.value = res.data?.data?.list || res.data?.list || []
  } catch (e) { /* ignore */ }
}

const loadStats = async () => {
  try {
    const res = await getBuildStats({ days: 7 })
    const data = res?.data || res || {}
    statsData.value = data.stats || {}
    trendData.value = data.trend || []
  } catch (e) {
    console.error('加载统计失败', e)
  }
}

const loadRecords = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    if (keyword.value) params.keyword = keyword.value
    if (pipelineFilter.value) params.pipeline_id = pipelineFilter.value
    const res = await getBuildRecords(params)
    const data = res?.data || res || {}
    records.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    console.error('加载构建记录失败', e)
  } finally {
    loading.value = false
  }
}

const loadAll = () => {
  loadStats()
  loadRecords()
}

const exportRecords = () => {
  exporting.value = true
  const params = new URLSearchParams()
  if (statusFilter.value) params.set('status', statusFilter.value)
  if (keyword.value) params.set('keyword', keyword.value)
  if (pipelineFilter.value) params.set('pipeline_id', pipelineFilter.value)
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  const url = `/api/v1/k8s/cicd/pipeline/build-records/export?${params.toString()}`
  fetch(url, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(blob => {
      const blobUrl = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = blobUrl
      link.setAttribute('download', 'build_records.csv')
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(blobUrl)
    })
    .finally(() => { exporting.value = false })
}

const setFilter = (s) => {
  statusFilter.value = s
  page.value = 1
  loadRecords()
}

const handleFilterChange = () => {
  page.value = 1
  loadRecords()
}

let debounceTimer = null
const debouncedLoad = () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { page.value = 1; loadRecords() }, 300)
}

const goToDetail = (r) => {
  router.push(`/cicd/pipelines/${r.pipeline_id}`)
}

const statusText = (s) => {
  const m = { pending: '等待中', running: '构建中', success: '成功', failed: '失败', aborted: '已中止' }
  return m[s] || s
}

const triggerText = (t) => {
  const m = { manual: '手动', webhook: 'Webhook', scheduled: '定时', callback: '回调' }
  return m[t] || t || '手动'
}

const shortImage = (url) => {
  if (!url) return '-'
  const parts = url.split('/')
  return parts.length > 1 ? parts.slice(-2).join('/') : url
}

const formatDuration = (sec) => {
  if (!sec) return '-'
  if (sec < 60) return `${sec}s`
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return s > 0 ? `${m}m${s}s` : `${m}m`
}

const formatTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const trendBarHeight = (total) => {
  const max = Math.max(...trendData.value.map(t => t.total || 0), 1)
  return Math.max((total / max) * 100, 8)
}

const trendBarSuccessRatio = (t) => {
  if (!t.total) return 0
  return (t.success / t.total) * 100
}

onMounted(() => {
  loadPipelines()
  loadAll()
})
</script>

<style scoped>
.build-records-page { padding: 24px; max-width: 1440px; margin: 0 auto; }

/* Banner */
.page-banner {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #334155 100%);
  border-radius: 16px; padding: 28px 32px; margin-bottom: 24px; color: #fff;
  box-shadow: 0 4px 24px rgba(15, 23, 42, 0.3);
}
.banner-inner { display: flex; justify-content: space-between; align-items: center; }
.banner-left { display: flex; align-items: center; gap: 16px; }
.banner-icon { width: 48px; height: 48px; border-radius: 12px; background: rgba(99, 102, 241, 0.2); display: flex; align-items: center; justify-content: center; }
.banner-icon svg { width: 28px; height: 28px; color: #818cf8; }
.banner-title { font-size: 22px; font-weight: 700; margin: 0; letter-spacing: -0.02em; }
.banner-desc { font-size: 14px; color: #94a3b8; margin: 4px 0 0; }
.banner-actions { display: flex; gap: 10px; }
.btn-action {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px;
  background: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);
  color: #e2e8f0; font-size: 13px; cursor: pointer; transition: all 0.2s;
}
.btn-action:hover { background: rgba(255,255,255,0.15); border-color: rgba(255,255,255,0.3); }
.btn-action:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-action svg { width: 16px; height: 16px; }

/* Metrics Panel */
.metrics-panel { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.metric-card {
  background: #fff; border-radius: 12px; padding: 20px;
  border: 1px solid #f1f5f9; box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  transition: all 0.2s;
}
.metric-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); transform: translateY(-1px); }
.metric-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.metric-label { font-size: 13px; color: #64748b; font-weight: 500; }
.metric-icon-sm { width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.metric-icon-sm svg { width: 16px; height: 16px; }
.metric-card.primary .metric-icon-sm { background: #eff6ff; color: #3b82f6; }
.metric-card.success .metric-icon-sm { background: #f0fdf4; color: #22c55e; }
.metric-card.warning .metric-icon-sm { background: #fffbeb; color: #f59e0b; }
.metric-card.info .metric-icon-sm { background: #f5f3ff; color: #8b5cf6; }
.metric-value { font-size: 28px; font-weight: 700; color: #0f172a; line-height: 1.2; }
.metric-unit { font-size: 16px; font-weight: 500; color: #64748b; margin-left: 2px; }
.metric-sub { font-size: 12px; color: #94a3b8; margin-top: 8px; }
.metric-sub strong { color: #475569; }

/* Mini Trend Chart */
.trend-mini { display: flex; gap: 3px; align-items: flex-end; height: 48px; padding: 4px 0; }
.trend-bar-wrap { flex: 1; height: 100%; display: flex; align-items: flex-end; }
.trend-bar { width: 100%; border-radius: 3px; background: #e2e8f0; position: relative; overflow: hidden; min-height: 4px; transition: height 0.3s; }
.trend-bar-success { position: absolute; bottom: 0; left: 0; right: 0; background: #22c55e; border-radius: 3px; transition: height 0.3s; }

/* Filter Toolbar */
.filter-toolbar {
  display: flex; justify-content: space-between; align-items: center;
  background: #fff; border-radius: 12px; padding: 12px 20px;
  border: 1px solid #f1f5f9; margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.status-tabs { display: flex; gap: 4px; }
.tab-btn {
  padding: 6px 14px; border-radius: 6px; border: none;
  background: transparent; font-size: 13px; color: #64748b;
  cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;
}
.tab-btn:hover { background: #f8fafc; color: #334155; }
.tab-btn.active { background: #0f172a; color: #fff; }
.tab-count { font-size: 11px; background: rgba(0,0,0,0.06); padding: 1px 6px; border-radius: 10px; }
.tab-btn.active .tab-count { background: rgba(255,255,255,0.2); }
.dot { width: 7px; height: 7px; border-radius: 50%; display: inline-block; }
.dot.running { background: #f59e0b; }
.dot.success { background: #22c55e; }
.dot.failed { background: #ef4444; }
.dot.aborted { background: #94a3b8; }

.filter-right { display: flex; gap: 10px; align-items: center; }
.filter-select {
  padding: 7px 12px; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 13px; background: #fff; color: #334155; cursor: pointer; min-width: 140px;
}
.search-input-wrap {
  display: flex; align-items: center; gap: 8px;
  border: 1px solid #e2e8f0; border-radius: 8px; padding: 0 12px;
  background: #fff; transition: border-color 0.2s; min-width: 220px;
}
.search-input-wrap:focus-within { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }
.search-input-wrap svg { width: 15px; height: 15px; color: #94a3b8; flex-shrink: 0; }
.search-input-wrap input { border: none; outline: none; flex: 1; padding: 7px 0; font-size: 13px; }
.clear-btn { border: none; background: none; cursor: pointer; padding: 2px; color: #94a3b8; }
.clear-btn:hover { color: #ef4444; }
.clear-btn svg { width: 14px; height: 14px; }

/* Table */
.table-wrapper {
  background: #fff; border-radius: 12px; overflow: hidden;
  border: 1px solid #f1f5f9; box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.data-table { width: 100%; border-collapse: collapse; }
.data-table th {
  padding: 12px 16px; text-align: left; font-size: 12px; font-weight: 600;
  color: #64748b; background: #f8fafc; border-bottom: 1px solid #e2e8f0;
  text-transform: uppercase; letter-spacing: 0.05em;
}
.data-table td {
  padding: 14px 16px; font-size: 13px; color: #334155;
  border-bottom: 1px solid #f1f5f9;
}
.row-hover { cursor: pointer; transition: background 0.15s; }
.row-hover:hover { background: #f8fafc; }
.build-id { font-family: 'SF Mono', 'JetBrains Mono', monospace; font-weight: 600; color: #6366f1; font-size: 13px; }
.pipeline-name { font-weight: 500; color: #0f172a; }
.status-tag {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: 500;
}
.status-dot { width: 6px; height: 6px; border-radius: 50%; }
.status-tag.success { background: #f0fdf4; color: #16a34a; }
.status-tag.success .status-dot { background: #22c55e; }
.status-tag.failed { background: #fef2f2; color: #dc2626; }
.status-tag.failed .status-dot { background: #ef4444; }
.status-tag.running { background: #fffbeb; color: #d97706; }
.status-tag.running .status-dot { background: #f59e0b; animation: pulse 1.5s infinite; }
.status-tag.pending { background: #f0f9ff; color: #0369a1; }
.status-tag.pending .status-dot { background: #38bdf8; }
.status-tag.aborted { background: #f8fafc; color: #64748b; }
.status-tag.aborted .status-dot { background: #94a3b8; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }

.trigger-badge { font-size: 12px; color: #64748b; padding: 2px 8px; background: #f1f5f9; border-radius: 4px; }
.branch-tag { display: inline-flex; align-items: center; gap: 4px; font-size: 12px; color: #7c3aed; background: #f5f3ff; padding: 3px 8px; border-radius: 4px; font-family: monospace; }
.branch-icon { width: 12px; height: 12px; }
.image-text { font-size: 12px; color: #64748b; font-family: monospace; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }
.duration-text { font-size: 13px; color: #475569; font-family: monospace; }
.time-text { font-size: 12px; color: #94a3b8; }
.text-muted { color: #cbd5e1; }

/* Loading */
.loading-state { display: flex; flex-direction: column; align-items: center; padding: 60px; color: #94a3b8; gap: 12px; }
.loader { display: flex; gap: 6px; }
.dot-bounce { width: 8px; height: 8px; background: #6366f1; border-radius: 50%; animation: bounce-anim 0.6s infinite alternate; }
.dot-bounce:nth-child(2) { animation-delay: 0.2s; }
.dot-bounce:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce-anim { to { transform: translateY(-8px); opacity: 0.3; } }

/* Empty */
.empty-state { display: flex; flex-direction: column; align-items: center; padding: 80px 20px; }
.empty-icon { width: 64px; height: 64px; background: #f1f5f9; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.empty-icon svg { width: 32px; height: 32px; color: #94a3b8; }
.empty-state h3 { font-size: 16px; color: #334155; margin: 0 0 4px; }
.empty-state p { font-size: 13px; color: #94a3b8; margin: 0; }

/* Pagination */
.pagination-bar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 0; margin-top: 16px;
}
.page-info { font-size: 13px; color: #64748b; }
.page-info strong { color: #334155; }
.page-controls { display: flex; align-items: center; gap: 12px; }
.page-size-select {
  padding: 6px 10px; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 12px; background: #fff; color: #475569; cursor: pointer;
}
.page-nav { display: flex; gap: 4px; }
.page-btn {
  width: 32px; height: 32px; border-radius: 6px; border: 1px solid #e2e8f0;
  background: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 13px; color: #475569; cursor: pointer; transition: all 0.15s;
}
.page-btn:hover:not(:disabled):not(.active) { border-color: #6366f1; color: #6366f1; }
.page-btn.active { background: #6366f1; color: #fff; border-color: #6366f1; }
.page-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.page-btn.ellipsis { border: none; background: none; }
.page-btn svg { width: 14px; height: 14px; }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Responsive */
@media (max-width: 1200px) {
  .metrics-panel { grid-template-columns: repeat(2, 1fr); }
  .filter-toolbar { flex-direction: column; gap: 12px; align-items: stretch; }
}
</style>

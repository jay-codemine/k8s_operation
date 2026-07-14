<template>
  <div class="release-history-page">
    <!-- 顶部 Banner -->
    <div class="page-banner">
      <div class="banner-inner">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
          </div>
          <div>
            <h1 class="banner-title">应用发布历史</h1>
            <p class="banner-desc">全量发布记录时间线，追踪应用版本变更、部署状态与回滚历史</p>
          </div>
        </div>
        <div class="banner-actions">
          <button class="btn-action" @click="$router.push('/cicd/releases')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/></svg>
            <span>发布管理</span>
          </button>
          <button class="btn-action refresh" @click="loadAll" :disabled="loading">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            <span>刷新</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-panel">
      <div class="stat-card">
        <div class="stat-icon total"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg></div>
        <div class="stat-content">
          <div class="stat-value">{{ enhancedStats.total || 0 }}</div>
          <div class="stat-label">总发布数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></div>
        <div class="stat-content">
          <div class="stat-value">{{ (enhancedStats.success_rate || 0).toFixed(1) }}%</div>
          <div class="stat-label">发布成功率</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon today"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg></div>
        <div class="stat-content">
          <div class="stat-value">{{ enhancedStats.today_count || 0 }}</div>
          <div class="stat-label">今日发布</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon week"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg></div>
        <div class="stat-content">
          <div class="stat-value">{{ enhancedStats.week_count || 0 }}</div>
          <div class="stat-label">本周发布</div>
        </div>
      </div>
    </div>

    <!-- 筛选工具栏 -->
    <div class="filter-toolbar">
      <div class="filter-group">
        <div class="status-tabs">
          <button :class="['tab-btn', { active: statusFilter === '' }]" @click="setFilter('')">全部</button>
          <button :class="['tab-btn', { active: statusFilter === 'Succeeded' }]" @click="setFilter('Succeeded')">
            <span class="dot succeeded"></span>成功
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'Failed' }]" @click="setFilter('Failed')">
            <span class="dot failed"></span>失败
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'Running' }]" @click="setFilter('Running')">
            <span class="dot running"></span>部署中
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'AwaitingApproval' }]" @click="setFilter('AwaitingApproval')">
            <span class="dot pending"></span>待审批
          </button>
          <button :class="['tab-btn', { active: statusFilter === 'Rollback' }]" @click="setFilter('Rollback')">
            <span class="dot rollback"></span>已回滚
          </button>
        </div>
      </div>
      <div class="filter-group right">
        <input v-model="appNameFilter" class="filter-input" placeholder="应用名称" @input="debouncedLoad" />
        <input v-model="namespaceFilter" class="filter-input" placeholder="命名空间" @input="debouncedLoad" />
        <div class="search-input-wrap">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="appNameFilter" placeholder="搜索应用..." @input="debouncedLoad" />
        </div>
      </div>
    </div>

    <!-- 时间线列表 -->
    <div class="timeline-wrapper">
      <div v-if="loading" class="loading-state">
        <div class="loader"><div class="dot-bounce"></div><div class="dot-bounce"></div><div class="dot-bounce"></div></div>
        <span>加载发布历史...</span>
      </div>

      <div v-else-if="releases.length > 0" class="timeline-list">
        <div v-for="r in releases" :key="r.id" class="timeline-item">
          <div class="timeline-marker">
            <div :class="['marker-dot', statusClass(r.status)]"></div>
            <div class="marker-line"></div>
          </div>
          <div class="timeline-card">
            <div class="card-header">
              <div class="card-title-row">
                <span class="app-name">{{ r.app_name }}</span>
                <span :class="['release-status', statusClass(r.status)]">{{ statusLabel(r.status) }}</span>
              </div>
              <div class="card-meta">
                <span class="meta-item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/></svg>
                  {{ r.namespace }}
                </span>
                <span class="meta-item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
                  {{ r.workload_kind }}/{{ r.workload_name }}
                </span>
                <span class="meta-item time">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  {{ formatTime(r.created_at) }}
                </span>
              </div>
            </div>
            <div class="card-body">
              <div class="image-info" v-if="r.image_repo">
                <span class="image-label">镜像</span>
                <span class="image-value">{{ r.image_repo }}{{ r.image_tag ? ':' + r.image_tag : '' }}</span>
              </div>
              <div class="message-info" v-if="r.message">
                <span class="message-label">信息</span>
                <span class="message-value">{{ r.message }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
        </div>
        <h3>暂无发布记录</h3>
        <p>创建发布单或通过流水线自动发布后，记录将在此展示</p>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination-bar">
      <div class="page-info">
        共 <strong>{{ total }}</strong> 条记录
      </div>
      <div class="page-controls">
        <select v-model="pageSize" class="page-size-select" @change="page = 1; loadReleases()">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
        <div class="page-nav">
          <button class="page-btn" :disabled="page <= 1" @click="page--; loadReleases()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <template v-for="p in visiblePages" :key="p">
            <button v-if="p === '...'" class="page-btn ellipsis" disabled>...</button>
            <button v-else :class="['page-btn', { active: p === page }]" @click="page = p; loadReleases()">{{ p }}</button>
          </template>
          <button class="page-btn" :disabled="page >= totalPages" @click="page++; loadReleases()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { getReleaseHistory, getReleaseStatsEnhanced } from '@/api/cicd'

const releases = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const statusFilter = ref('')
const appNameFilter = ref('')
const namespaceFilter = ref('')
const enhancedStats = ref({})

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

const loadStats = async () => {
  try {
    const res = await getReleaseStatsEnhanced()
    enhancedStats.value = res?.data || res || {}
  } catch (e) {
    console.error('加载发布统计失败', e)
  }
}

const loadReleases = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    if (appNameFilter.value) params.app_name = appNameFilter.value
    if (namespaceFilter.value) params.namespace = namespaceFilter.value
    const res = await getReleaseHistory(params)
    const data = res?.data || res || {}
    releases.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    console.error('加载发布历史失败', e)
  } finally {
    loading.value = false
  }
}

const loadAll = () => {
  loadStats()
  loadReleases()
}

const setFilter = (s) => {
  statusFilter.value = s
  page.value = 1
  loadReleases()
}

let debounceTimer = null
const debouncedLoad = () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { page.value = 1; loadReleases() }, 400)
}

const statusClass = (s) => {
  const map = {
    Succeeded: 'succeeded', Failed: 'failed', Running: 'running',
    Pending: 'pending', Queued: 'pending', AwaitingApproval: 'pending',
    Canceled: 'canceled', Rollback: 'rollback'
  }
  return map[s] || 'pending'
}

const statusLabel = (s) => {
  const map = {
    Pending: '待处理', Queued: '排队中', Running: '部署中',
    Succeeded: '发布成功', Failed: '发布失败',
    Canceled: '已取消', Rollback: '已回滚', AwaitingApproval: '待审批'
  }
  return map[s] || s
}

const formatTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.release-history-page { padding: 24px; max-width: 1440px; margin: 0 auto; }

/* Banner */
.page-banner {
  background: linear-gradient(135deg, #0c4a6e 0%, #164e63 50%, #1e3a5f 100%);
  border-radius: 16px; padding: 28px 32px; margin-bottom: 24px; color: #fff;
  box-shadow: 0 4px 24px rgba(12, 74, 110, 0.3);
}
.banner-inner { display: flex; justify-content: space-between; align-items: center; }
.banner-left { display: flex; align-items: center; gap: 16px; }
.banner-icon { width: 48px; height: 48px; border-radius: 12px; background: rgba(34, 211, 238, 0.15); display: flex; align-items: center; justify-content: center; }
.banner-icon svg { width: 28px; height: 28px; color: #67e8f9; }
.banner-title { font-size: 22px; font-weight: 700; margin: 0; letter-spacing: -0.02em; }
.banner-desc { font-size: 14px; color: #a5f3fc; margin: 4px 0 0; opacity: 0.85; }
.banner-actions { display: flex; gap: 10px; }
.btn-action {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px;
  background: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);
  color: #e0f2fe; font-size: 13px; cursor: pointer; transition: all 0.2s;
}
.btn-action:hover { background: rgba(255,255,255,0.15); }
.btn-action:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-action svg { width: 16px; height: 16px; }

/* Stats */
.stats-panel { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.stat-card {
  background: #fff; border-radius: 12px; padding: 20px;
  display: flex; align-items: center; gap: 14px;
  border: 1px solid #f1f5f9; box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  transition: all 0.2s;
}
.stat-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); transform: translateY(-1px); }
.stat-icon { width: 44px; height: 44px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.stat-icon svg { width: 20px; height: 20px; }
.stat-icon.total { background: #eff6ff; color: #3b82f6; }
.stat-icon.success { background: #f0fdf4; color: #22c55e; }
.stat-icon.today { background: #fefce8; color: #eab308; }
.stat-icon.week { background: #f5f3ff; color: #8b5cf6; }
.stat-value { font-size: 24px; font-weight: 700; color: #0f172a; }
.stat-label { font-size: 12px; color: #94a3b8; margin-top: 2px; }

/* Filter */
.filter-toolbar {
  display: flex; justify-content: space-between; align-items: center;
  background: #fff; border-radius: 12px; padding: 12px 20px;
  border: 1px solid #f1f5f9; margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04); flex-wrap: wrap; gap: 12px;
}
.filter-group { display: flex; align-items: center; gap: 8px; }
.filter-group.right { display: flex; gap: 10px; }
.status-tabs { display: flex; gap: 4px; }
.tab-btn {
  padding: 6px 12px; border-radius: 6px; border: none;
  background: transparent; font-size: 13px; color: #64748b;
  cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 5px;
}
.tab-btn:hover { background: #f1f5f9; }
.tab-btn.active { background: #0c4a6e; color: #fff; }
.dot { width: 7px; height: 7px; border-radius: 50%; }
.dot.succeeded { background: #22c55e; }
.dot.failed { background: #ef4444; }
.dot.running { background: #f59e0b; }
.dot.pending { background: #38bdf8; }
.dot.rollback { background: #a855f7; }
.filter-input {
  padding: 7px 12px; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 13px; width: 120px; outline: none; transition: border-color 0.2s;
}
.filter-input:focus { border-color: #0ea5e9; }
.search-input-wrap {
  display: flex; align-items: center; gap: 8px;
  border: 1px solid #e2e8f0; border-radius: 8px; padding: 0 12px;
  min-width: 180px; transition: border-color 0.2s;
}
.search-input-wrap:focus-within { border-color: #0ea5e9; box-shadow: 0 0 0 3px rgba(14,165,233,0.1); }
.search-input-wrap svg { width: 14px; height: 14px; color: #94a3b8; flex-shrink: 0; }
.search-input-wrap input { border: none; outline: none; flex: 1; padding: 7px 0; font-size: 13px; }

/* Timeline */
.timeline-wrapper { background: #fff; border-radius: 12px; border: 1px solid #f1f5f9; padding: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); }
.timeline-list { display: flex; flex-direction: column; }
.timeline-item { display: flex; gap: 16px; position: relative; }
.timeline-item:last-child .marker-line { display: none; }
.timeline-marker { display: flex; flex-direction: column; align-items: center; padding-top: 4px; }
.marker-dot { width: 12px; height: 12px; border-radius: 50%; border: 3px solid; flex-shrink: 0; }
.marker-dot.succeeded { border-color: #22c55e; background: #dcfce7; }
.marker-dot.failed { border-color: #ef4444; background: #fef2f2; }
.marker-dot.running { border-color: #f59e0b; background: #fef9c3; }
.marker-dot.pending { border-color: #38bdf8; background: #e0f2fe; }
.marker-dot.canceled { border-color: #94a3b8; background: #f1f5f9; }
.marker-dot.rollback { border-color: #a855f7; background: #f3e8ff; }
.marker-line { width: 2px; flex: 1; background: #e2e8f0; margin: 4px 0; min-height: 20px; }
.timeline-card {
  flex: 1; border: 1px solid #f1f5f9; border-radius: 10px; padding: 16px 20px;
  margin-bottom: 16px; transition: all 0.2s; cursor: default;
}
.timeline-card:hover { border-color: #e2e8f0; box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.card-header { margin-bottom: 10px; }
.card-title-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.app-name { font-size: 15px; font-weight: 600; color: #0f172a; }
.release-status {
  font-size: 12px; font-weight: 500; padding: 3px 10px; border-radius: 12px;
}
.release-status.succeeded { background: #f0fdf4; color: #16a34a; }
.release-status.failed { background: #fef2f2; color: #dc2626; }
.release-status.running { background: #fffbeb; color: #d97706; }
.release-status.pending { background: #f0f9ff; color: #0369a1; }
.release-status.canceled { background: #f8fafc; color: #64748b; }
.release-status.rollback { background: #faf5ff; color: #7c3aed; }
.card-meta { display: flex; gap: 16px; flex-wrap: wrap; }
.meta-item { display: flex; align-items: center; gap: 4px; font-size: 12px; color: #64748b; }
.meta-item svg { width: 13px; height: 13px; }
.meta-item.time { color: #94a3b8; }
.card-body { display: flex; gap: 24px; flex-wrap: wrap; }
.image-info, .message-info { display: flex; align-items: center; gap: 8px; }
.image-label, .message-label { font-size: 11px; color: #94a3b8; text-transform: uppercase; font-weight: 600; }
.image-value { font-size: 12px; color: #475569; font-family: monospace; background: #f8fafc; padding: 2px 8px; border-radius: 4px; }
.message-value { font-size: 12px; color: #64748b; }

/* Loading / Empty */
.loading-state { display: flex; flex-direction: column; align-items: center; padding: 60px; color: #94a3b8; gap: 12px; }
.loader { display: flex; gap: 6px; }
.dot-bounce { width: 8px; height: 8px; background: #0ea5e9; border-radius: 50%; animation: bounce-anim 0.6s infinite alternate; }
.dot-bounce:nth-child(2) { animation-delay: 0.2s; }
.dot-bounce:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce-anim { to { transform: translateY(-8px); opacity: 0.3; } }
.empty-state { display: flex; flex-direction: column; align-items: center; padding: 80px 20px; }
.empty-icon { width: 64px; height: 64px; background: #f0f9ff; border-radius: 16px; display: flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.empty-icon svg { width: 32px; height: 32px; color: #38bdf8; }
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
.page-btn:hover:not(:disabled):not(.active) { border-color: #0ea5e9; color: #0ea5e9; }
.page-btn.active { background: #0ea5e9; color: #fff; border-color: #0ea5e9; }
.page-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.page-btn.ellipsis { border: none; background: none; }
.page-btn svg { width: 14px; height: 14px; }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 1200px) {
  .stats-panel { grid-template-columns: repeat(2, 1fr); }
}
</style>

<template>
  <div class="build-records-page">
    <!-- 顶部 Banner -->
    <div class="page-banner">
      <div class="banner-inner">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
          </div>
          <div>
            <h1 class="banner-title">构建记录</h1>
            <p class="banner-desc">全局构建运行记录，跨流水线查看构建状态、耗时与产物</p>
          </div>
        </div>
        <div class="banner-actions">
          <button class="btn-action" @click="exportRecords" :disabled="exporting">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span>{{ exporting ? '导出中...' : '导出 CSV' }}</span>
          </button>
          <button class="btn-action refresh" @click="loadRecords" :disabled="loading">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            <span>刷新</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" :class="{ active: statusFilter === '' }" @click="setFilter('')">
        <div class="stat-icon total"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/></svg></div>
        <div class="stat-body"><span class="stat-num">{{ total }}</span><span class="stat-label">全部</span></div>
      </div>
      <div class="stat-card" :class="{ active: statusFilter === 'running' }" @click="setFilter('running')">
        <div class="stat-icon running"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg></div>
        <div class="stat-body"><span class="stat-num">{{ countByStatus('running') + countByStatus('pending') }}</span><span class="stat-label">运行中</span></div>
      </div>
      <div class="stat-card" :class="{ active: statusFilter === 'success' }" @click="setFilter('success')">
        <div class="stat-icon success"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></div>
        <div class="stat-body"><span class="stat-num">{{ countByStatus('success') }}</span><span class="stat-label">成功</span></div>
      </div>
      <div class="stat-card" :class="{ active: statusFilter === 'failed' }" @click="setFilter('failed')">
        <div class="stat-icon failed"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg></div>
        <div class="stat-body"><span class="stat-num">{{ countByStatus('failed') }}</span><span class="stat-label">失败</span></div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input v-model="keyword" placeholder="搜索镜像、错误信息..." @input="debouncedLoad" />
      </div>
      <select v-model="pipelineFilter" class="filter-select" @change="loadRecords">
        <option value="">全部流水线</option>
        <option v-for="p in pipelines" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
    </div>

    <!-- 表格 -->
    <div class="table-container">
      <div v-if="loading" class="loading-state">
        <div class="loader"><div class="dot"></div><div class="dot"></div><div class="dot"></div></div>
        <span>加载中...</span>
      </div>
      <table v-else-if="records.length > 0" class="data-table">
        <thead>
          <tr>
            <th>构建号</th>
            <th>流水线</th>
            <th>状态</th>
            <th>触发方式</th>
            <th>分支</th>
            <th>镜像</th>
            <th>耗时</th>
            <th>时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in records" :key="r.id" @click="goToDetail(r)">
            <td class="mono">#{{ r.build_number || '-' }}</td>
            <td>
              <span class="pipeline-name">{{ r.pipeline_name || `Pipeline#${r.pipeline_id}` }}</span>
            </td>
            <td>
              <span :class="['status-badge', r.status]">{{ statusText(r.status) }}</span>
            </td>
            <td>{{ triggerText(r.trigger_type) }}</td>
            <td class="mono">{{ r.git_branch || '-' }}</td>
            <td class="image-cell" :title="r.image_url">{{ shortImage(r.image_url) }}</td>
            <td>{{ formatDuration(r.duration_sec) }}</td>
            <td>{{ formatTime(r.created_at) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
        <p>暂无构建记录</p>
        <span>触发流水线构建后，记录将在此展示</span>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="pagination">
      <button :disabled="page <= 1" @click="page--; loadRecords()">上一页</button>
      <span class="page-info">第 {{ page }} / {{ Math.ceil(total / pageSize) }} 页 · 共 {{ total }} 条</span>
      <button :disabled="page >= Math.ceil(total / pageSize)" @click="page++; loadRecords()">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/api/http'
import { API_BASE } from '@/api/paths'
import { getPipelines } from '@/api/platform/pipeline'

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

// 加载流水线列表（用于筛选下拉）
const loadPipelines = async () => {
  try {
    const res = await getPipelines({ page: 1, page_size: 200 })
    pipelines.value = res.data?.data?.list || res.data?.list || []
  } catch (e) { /* ignore */ }
}

// 加载构建记录
const loadRecords = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    if (keyword.value) params.keyword = keyword.value
    if (pipelineFilter.value) params.pipeline_id = pipelineFilter.value
    const res = await http.get(`${API_BASE}/k8s/cicd/pipeline/build-records`, { params })
    const data = res.data?.data || res.data || {}
    records.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    console.error('加载构建记录失败', e)
  } finally {
    loading.value = false
  }
}

// 导出 CSV
const exportRecords = () => {
  exporting.value = true
  const params = new URLSearchParams()
  if (statusFilter.value) params.set('status', statusFilter.value)
  if (keyword.value) params.set('keyword', keyword.value)
  if (pipelineFilter.value) params.set('pipeline_id', pipelineFilter.value)
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  const url = `${API_BASE}/k8s/cicd/pipeline/build-records/export?${params.toString()}`
  // 通过创建隐藏 iframe 触发下载
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', 'build_records.csv')
  // 使用 fetch 带 token
  fetch(url, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(blob => {
      const blobUrl = URL.createObjectURL(blob)
      link.href = blobUrl
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

const countByStatus = (status) => {
  return records.value.filter(r => r.status === status).length
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
  const m = { pending: '等待中', running: '运行中', success: '成功', failed: '失败', aborted: '已中止' }
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

onMounted(() => {
  loadPipelines()
  loadRecords()
})
</script>

<style scoped>
.build-records-page { padding: 1.5rem; max-width: 1400px; }

.page-banner {
  background: linear-gradient(135deg, #1d2129 0%, #2d3748 100%);
  border-radius: 0.75rem; padding: 1.5rem 2rem; margin-bottom: 1.5rem; color: #fff;
}
.banner-inner { display: flex; justify-content: space-between; align-items: center; }
.banner-left { display: flex; align-items: center; gap: 1rem; }
.banner-icon svg { width: 2.5rem; height: 2.5rem; }
.banner-title { font-size: 1.5rem; font-weight: 700; margin: 0; }
.banner-desc { font-size: 0.875rem; color: #a0aec0; margin: 0.25rem 0 0; }
.banner-actions { display: flex; gap: 0.75rem; }
.btn-action {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.5rem 1rem; border-radius: 0.5rem;
  background: rgba(255,255,255,0.1); border: 1px solid rgba(255,255,255,0.2);
  color: #fff; font-size: 0.8125rem; cursor: pointer; transition: all 0.2s;
}
.btn-action:hover { background: rgba(255,255,255,0.2); }
.btn-action:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-action svg { width: 1rem; height: 1rem; }

.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
.stat-card {
  background: #fff; border-radius: 0.75rem; padding: 1rem 1.25rem;
  display: flex; align-items: center; gap: 0.75rem; cursor: pointer;
  border: 2px solid transparent; transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.stat-card:hover { border-color: #e2e8f0; }
.stat-card.active { border-color: #165DFF; background: #f0f5ff; }
.stat-icon { width: 2.5rem; height: 2.5rem; border-radius: 0.5rem; display: flex; align-items: center; justify-content: center; }
.stat-icon svg { width: 1.25rem; height: 1.25rem; }
.stat-icon.total { background: #e8f4fd; color: #1890ff; }
.stat-icon.running { background: #fff7e6; color: #fa8c16; }
.stat-icon.success { background: #f6ffed; color: #52c41a; }
.stat-icon.failed { background: #fff2f0; color: #ff4d4f; }
.stat-body { display: flex; flex-direction: column; }
.stat-num { font-size: 1.5rem; font-weight: 700; color: #1d2129; line-height: 1.2; }
.stat-label { font-size: 0.75rem; color: #86909c; }

.filter-bar { display: flex; gap: 1rem; margin-bottom: 1rem; }
.search-box {
  flex: 1; display: flex; align-items: center; gap: 0.5rem;
  background: #fff; border: 1px solid #e5e6eb; border-radius: 0.5rem; padding: 0 1rem;
}
.search-box svg { width: 1rem; height: 1rem; color: #86909c; flex-shrink: 0; }
.search-box input { border: none; outline: none; flex: 1; padding: 0.625rem 0; font-size: 0.875rem; }
.filter-select {
  padding: 0.625rem 1rem; border: 1px solid #e5e6eb; border-radius: 0.5rem;
  font-size: 0.875rem; background: #fff; min-width: 160px; cursor: pointer;
}

.table-container { background: #fff; border-radius: 0.75rem; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th {
  padding: 0.75rem 1rem; text-align: left; font-size: 0.75rem; font-weight: 600;
  color: #86909c; background: #fafbfc; border-bottom: 1px solid #e5e6eb;
  text-transform: uppercase; letter-spacing: 0.03em;
}
.data-table td {
  padding: 0.75rem 1rem; font-size: 0.8125rem; color: #1d2129;
  border-bottom: 1px solid #f2f3f5;
}
.data-table tr { cursor: pointer; transition: background 0.15s; }
.data-table tbody tr:hover { background: #f7f8fa; }
.mono { font-family: 'SF Mono', Monaco, monospace; font-size: 0.8rem; }
.pipeline-name { font-weight: 500; color: #165DFF; }
.image-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.75rem; color: #4e5969; }

.status-badge {
  display: inline-flex; padding: 0.2rem 0.6rem; border-radius: 0.25rem;
  font-size: 0.75rem; font-weight: 500;
}
.status-badge.success { background: #f6ffed; color: #52c41a; }
.status-badge.failed { background: #fff2f0; color: #ff4d4f; }
.status-badge.running { background: #fff7e6; color: #fa8c16; }
.status-badge.pending { background: #f0f5ff; color: #165DFF; }
.status-badge.aborted { background: #f2f3f5; color: #86909c; }

.loading-state { display: flex; flex-direction: column; align-items: center; padding: 3rem; color: #86909c; gap: 1rem; }
.loader { display: flex; gap: 0.25rem; }
.dot { width: 0.5rem; height: 0.5rem; background: #165DFF; border-radius: 50%; animation: bounce 0.6s infinite alternate; }
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce { to { transform: translateY(-0.5rem); opacity: 0.4; } }

.empty-state { display: flex; flex-direction: column; align-items: center; padding: 4rem; color: #86909c; }
.empty-state svg { width: 3rem; height: 3rem; margin-bottom: 1rem; opacity: 0.5; }
.empty-state p { font-size: 1rem; font-weight: 500; margin: 0; }
.empty-state span { font-size: 0.8125rem; margin-top: 0.25rem; }

.pagination {
  display: flex; align-items: center; justify-content: center; gap: 1rem;
  padding: 1rem 0; margin-top: 1rem;
}
.pagination button {
  padding: 0.5rem 1rem; border: 1px solid #e5e6eb; border-radius: 0.375rem;
  background: #fff; font-size: 0.8125rem; cursor: pointer; transition: all 0.2s;
}
.pagination button:hover:not(:disabled) { border-color: #165DFF; color: #165DFF; }
.pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { font-size: 0.8125rem; color: #4e5969; }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>

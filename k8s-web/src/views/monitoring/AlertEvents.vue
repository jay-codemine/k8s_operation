<template>
  <div class="alert-events-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h3>告警事件</h3>
        <span class="header-desc">记录所有触发的告警事件，支持确认、解决与批量操作</span>
      </div>
      <div class="header-actions">
        <button class="btn-action btn-batch-delete" :disabled="selectedIds.length === 0" @click="batchDeleteVisible = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
          批量删除<span v-if="selectedIds.length">({{ selectedIds.length }})</span>
        </button>
        <button class="btn-action btn-batch-resolve" :disabled="selectedFiringIds.length === 0" @click="handleBatchResolve">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
          批量解决<span v-if="selectedFiringIds.length">({{ selectedFiringIds.length }})</span>
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-bar">
      <div class="stat-card critical">
        <span class="stat-value">{{ stats.critical }}</span>
        <span class="stat-label">🔴 Critical</span>
      </div>
      <div class="stat-card warning">
        <span class="stat-value">{{ stats.warning }}</span>
        <span class="stat-label">🟡 Warning</span>
      </div>
      <div class="stat-card info">
        <span class="stat-value">{{ stats.info }}</span>
        <span class="stat-label">🔵 Info</span>
      </div>
      <div class="stat-card resolved">
        <span class="stat-value">{{ stats.total_resolved }}</span>
        <span class="stat-label">✅ 已恢复</span>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <input v-model="filters.keyword" placeholder="搜索规则名/摘要..." class="search-input" @input="debouncedLoad" />
      <select v-model="filters.severity" @change="page = 1; loadList()" class="filter-select">
        <option value="">全部级别</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </select>
      <select v-model="filters.status" @change="page = 1; loadList()" class="filter-select">
        <option value="">全部状态</option>
        <option value="firing">告警中</option>
        <option value="resolved">已恢复</option>
        <option value="silenced">已静默</option>
      </select>
      <select v-model="timeRange" @change="applyTimeRange" class="filter-select">
        <option value="1h">最近 1 小时</option>
        <option value="6h">最近 6 小时</option>
        <option value="24h">最近 24 小时</option>
        <option value="7d">最近 7 天</option>
        <option value="">全部时间</option>
      </select>
    </div>

    <!-- 事件列表 -->
    <div class="events-table-wrapper" v-if="list.length">
      <table class="data-table">
        <thead>
          <tr>
            <th class="th-checkbox">
              <label class="table-checkbox">
                <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
                <span class="checkmark"></span>
              </label>
            </th>
            <th>级别</th>
            <th>规则</th>
            <th>摘要</th>
            <th>触发时间</th>
            <th>恢复时间</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ev in list" :key="ev.id" :class="['clickable-row', { selected: selectedIds.includes(ev.id) }]" @click="openDetail(ev)">
            <td class="td-checkbox" @click.stop>
              <label class="table-checkbox">
                <input type="checkbox" :checked="selectedIds.includes(ev.id)" @change="toggleSelect(ev.id)" />
                <span class="checkmark"></span>
              </label>
            </td>
            <td>
              <span class="severity-badge" :class="ev.severity">{{ ev.severity }}</span>
            </td>
            <td class="rule-cell">{{ ev.rule_name }}</td>
            <td class="summary-cell">{{ ev.summary || '-' }}</td>
            <td class="time-cell">{{ formatTime(ev.fired_at) }}</td>
            <td class="time-cell">{{ ev.resolved_at ? formatTime(ev.resolved_at) : '-' }}</td>
            <td>
              <span class="status-badge" :class="ev.status">{{ statusMap[ev.status] }}</span>
            </td>
            <td class="action-cell" @click.stop>
              <button v-if="ev.status === 'firing' && !ev.acked_by" class="btn-sm" @click="doAck(ev)">确认</button>
              <button v-if="ev.status === 'firing'" class="btn-sm btn-resolve" @click="doResolve(ev)">解决</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar" v-if="total > 0">
      <div class="pagination-info">
        共 <b>{{ total }}</b> 条，第 <b>{{ page }}</b> / {{ totalPages }} 页
      </div>
      <div class="pagination-controls">
        <button class="page-btn" :disabled="page <= 1" @click="goPage(1)">«</button>
        <button class="page-btn" :disabled="page <= 1" @click="goPage(page - 1)">‹</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" class="page-btn page-ellipsis" disabled>...</button>
          <button v-else class="page-btn" :class="{ active: p === page }" @click="goPage(p)">{{ p }}</button>
        </template>
        <button class="page-btn" :disabled="page >= totalPages" @click="goPage(page + 1)">›</button>
        <button class="page-btn" :disabled="page >= totalPages" @click="goPage(totalPages)">»</button>
        <select class="page-size-select" v-model="pageSize" @change="page = 1; loadList()">
          <option :value="10">10条/页</option>
          <option :value="20">20条/页</option>
          <option :value="50">50条/页</option>
        </select>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-if="!list.length">
      <div class="empty-icon">🎉</div>
      <h3>暂无告警事件</h3>
      <p>当告警规则检测到异常时，事件将自动记录在此</p>
    </div>

    <!-- 批量删除确认弹窗 -->
    <div class="modal-overlay" v-if="batchDeleteVisible" @click.self="batchDeleteVisible = false">
      <div class="modal-dialog modal-sm">
        <div class="modal-header">
          <h3>⚠️ 批量删除确认</h3>
          <button class="modal-close" @click="batchDeleteVisible = false">×</button>
        </div>
        <div class="modal-body">
          <p>确定删除选中的 <b style="color:#dc2626">{{ selectedIds.length }}</b> 条告警事件？此操作不可恢复。</p>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="batchDeleteVisible = false">取消</button>
          <button class="btn-danger" @click="handleBatchDelete" :disabled="batchDeleting">
            {{ batchDeleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情侧抽屉 -->
    <div class="drawer-overlay" v-if="detailEvent" @click.self="detailEvent = null">
      <div class="drawer-panel">
        <div class="drawer-header">
          <h3>告警详情</h3>
          <button class="modal-close" @click="detailEvent = null">×</button>
        </div>
        <div class="drawer-body">
          <div class="detail-section">
            <div class="detail-row">
              <span class="detail-label">规则名称</span>
              <span class="detail-value">{{ detailEvent.rule_name }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">级别</span>
              <span class="severity-badge" :class="detailEvent.severity">{{ detailEvent.severity }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">状态</span>
              <span class="status-badge" :class="detailEvent.status">{{ statusMap[detailEvent.status] }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">触发值</span>
              <span class="detail-value mono">{{ detailEvent.value || '-' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">触发时间</span>
              <span class="detail-value">{{ formatTime(detailEvent.fired_at) }}</span>
            </div>
            <div class="detail-row" v-if="detailEvent.resolved_at">
              <span class="detail-label">恢复时间</span>
              <span class="detail-value">{{ formatTime(detailEvent.resolved_at) }}</span>
            </div>
            <div class="detail-row" v-if="detailEvent.acked_at">
              <span class="detail-label">确认时间</span>
              <span class="detail-value">{{ formatTime(detailEvent.acked_at) }}</span>
            </div>
          </div>
          <div class="detail-section" v-if="detailEvent.summary">
            <h4>摘要</h4>
            <p class="detail-text">{{ detailEvent.summary }}</p>
          </div>
          <div class="detail-section" v-if="detailEvent.description">
            <h4>描述</h4>
            <p class="detail-text">{{ detailEvent.description }}</p>
          </div>
          <div class="detail-section" v-if="detailEvent.labels">
            <h4>标签</h4>
            <pre class="detail-code">{{ formatJSON(detailEvent.labels) }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { listAlertEvents, getAlertStats, ackAlertEvent, resolveAlertEvent, batchDeleteAlertEvents } from '@/api/monitoring'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)
const stats = reactive({ total_firing: 0, total_resolved: 0, critical: 0, warning: 0, info: 0 })
const detailEvent = ref(null)
const timeRange = ref('24h')
const filters = reactive({ keyword: '', severity: '', status: '', start_time: 0, end_time: 0 })
const statusMap = { firing: '告警中', resolved: '已恢复', silenced: '已静默' }

// 选择
const selectedIds = ref([])
const isAllSelected = computed(() => list.value.length > 0 && list.value.every(r => selectedIds.value.includes(r.id)))
const selectedFiringIds = computed(() => selectedIds.value.filter(id => {
  const ev = list.value.find(e => e.id === id)
  return ev && ev.status === 'firing'
}))

// 批量操作
const batchDeleteVisible = ref(false)
const batchDeleting = ref(false)

function toggleSelect(id) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function toggleSelectAll() {
  if (isAllSelected.value) selectedIds.value = []
  else selectedIds.value = list.value.map(r => r.id)
}

let debounceTimer = null
const debouncedLoad = () => { clearTimeout(debounceTimer); debounceTimer = setTimeout(() => { page.value = 1; loadList() }, 300) }

function applyTimeRange() {
  if (!timeRange.value) {
    filters.start_time = 0
  } else {
    const hours = { '1h': 1, '6h': 6, '24h': 24, '7d': 168 }
    const h = hours[timeRange.value] || 24
    filters.start_time = Math.floor(Date.now() / 1000) - h * 3600
  }
  filters.end_time = 0
  page.value = 1
  loadList()
}

function formatTime(ts) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN', { hour12: false })
}

function formatJSON(str) {
  try { return JSON.stringify(JSON.parse(str), null, 2) } catch { return str }
}

// 分页逻辑
const visiblePages = computed(() => {
  const curr = page.value
  const tp = totalPages.value
  if (tp <= 7) return Array.from({ length: tp }, (_, i) => i + 1)
  const pages = []
  pages.push(1)
  if (curr > 3) pages.push('...')
  for (let i = Math.max(2, curr - 1); i <= Math.min(tp - 1, curr + 1); i++) pages.push(i)
  if (curr < tp - 2) pages.push('...')
  pages.push(tp)
  return pages
})

function goPage(p) {
  if (p < 1 || p > totalPages.value || p === page.value) return
  page.value = p
  loadList()
}

async function loadList() {
  try {
    const res = await listAlertEvents({ page: page.value, size: pageSize.value, ...filters })
    if (res?.code === 0) {
      list.value = res.data?.items || []
      total.value = res.data?.total || 0
    }
  } catch {}
}

async function loadStats() {
  try {
    const res = await getAlertStats()
    if (res?.code === 0 && res.data) Object.assign(stats, res.data)
  } catch {}
}

function openDetail(ev) { detailEvent.value = ev }

async function doAck(ev) {
  try { await ackAlertEvent(ev.id); loadList(); loadStats() } catch {}
}

async function doResolve(ev) {
  try { await resolveAlertEvent(ev.id); loadList(); loadStats() } catch {}
}

async function handleBatchDelete() {
  if (!selectedIds.value.length) return
  batchDeleting.value = true
  try {
    const res = await batchDeleteAlertEvents({ ids: selectedIds.value })
    if (res?.code === 0) {
      Message.success(`成功删除 ${res.data?.success || selectedIds.value.length} 条事件`)
      selectedIds.value = []
      batchDeleteVisible.value = false
      loadList()
      loadStats()
    } else {
      Message.error(res?.msg || '批量删除失败')
    }
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally {
    batchDeleting.value = false
  }
}

async function handleBatchResolve() {
  if (!selectedFiringIds.value.length) return
  try {
    for (const id of selectedFiringIds.value) {
      await resolveAlertEvent(id)
    }
    Message.success(`已解决 ${selectedFiringIds.value.length} 条告警`)
    selectedIds.value = []
    loadList()
    loadStats()
  } catch {
    Message.error('批量解决失败')
  }
}

onMounted(() => { applyTimeRange(); loadStats() })
</script>

<style scoped>
.alert-events-page { padding: 24px; }

/* Header */
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-left { display: flex; align-items: baseline; gap: 12px; }
.page-header h3 { font-size: 18px; font-weight: 700; color: #1f2937; margin: 0; }
.header-desc { font-size: 13px; color: #9ca3af; }
.header-actions { display: flex; gap: 10px; }
.btn-action { display: inline-flex; align-items: center; gap: 6px; padding: 8px 14px; border-radius: 8px; font-size: 13px; font-weight: 500; border: 1.5px solid #e2e8f0; background: #fff; cursor: pointer; transition: all 0.2s; }
.btn-action:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-batch-delete { border-color: #dc2626; color: #dc2626; }
.btn-batch-delete:hover:not(:disabled) { background: #fef2f2; }
.btn-batch-resolve { border-color: #059669; color: #059669; }
.btn-batch-resolve:hover:not(:disabled) { background: #ecfdf5; }

.stats-bar { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }
.stat-card { background: #fff; border-radius: 12px; padding: 16px 20px; border: 1px solid #e8ecf0; display: flex; flex-direction: column; align-items: center; gap: 4px; transition: all 0.2s; }
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.06); }
.stat-card.critical { border-left: 3px solid #dc2626; }
.stat-card.warning { border-left: 3px solid #d97706; }
.stat-card.info { border-left: 3px solid #2563eb; }
.stat-card.resolved { border-left: 3px solid #059669; }
.stat-value { font-size: 28px; font-weight: 700; color: #1a202c; }
.stat-label { font-size: 13px; color: #718096; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.search-input { flex: 1; min-width: 200px; max-width: 300px; padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.08); }
.filter-select { padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: #fff; }

.events-table-wrapper { background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: #f7f8fa; padding: 12px 16px; text-align: left; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e8ecf0; }
.data-table td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; }
.th-checkbox, .td-checkbox { width: 40px; text-align: center; }
.clickable-row { cursor: pointer; transition: background 0.15s; }
.clickable-row:hover { background: #f7fafc; }
.clickable-row.selected { background: #f5f3ff; }

/* Checkbox */
.table-checkbox { position: relative; display: inline-flex; cursor: pointer; }
.table-checkbox input { position: absolute; opacity: 0; width: 0; height: 0; }
.checkmark { width: 16px; height: 16px; border: 1.5px solid #d1d5db; border-radius: 4px; background: #fff; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.table-checkbox input:checked + .checkmark { background: linear-gradient(135deg, #4f46e5, #7c3aed); border-color: transparent; }
.table-checkbox input:checked + .checkmark::after { content: '✓'; color: #fff; font-size: 11px; font-weight: 700; }

.severity-badge { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; }
.severity-badge.critical { background: #fef2f2; color: #dc2626; }
.severity-badge.warning { background: #fffbeb; color: #d97706; }
.severity-badge.info { background: #eff6ff; color: #2563eb; }
.status-badge { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 500; }
.status-badge.firing { background: #fef2f2; color: #dc2626; }
.status-badge.resolved { background: #ecfdf5; color: #059669; }
.status-badge.silenced { background: #f3f4f6; color: #6b7280; }
.rule-cell { font-weight: 500; color: #1a202c; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.summary-cell { max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #4a5568; }
.time-cell { font-size: 12px; color: #718096; white-space: nowrap; }
.action-cell { white-space: nowrap; }
.btn-sm { padding: 4px 10px; border-radius: 6px; font-size: 12px; cursor: pointer; border: 1px solid #e2e8f0; background: #fff; margin-right: 4px; }
.btn-sm:hover { background: #f7fafc; border-color: #4f46e5; color: #4f46e5; }
.btn-resolve { border-color: #059669; color: #059669; }
.btn-resolve:hover { background: #059669; color: #fff; }

/* Pagination */
.pagination-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; margin-top: 16px; }
.pagination-info { font-size: 13px; color: #6b7280; }
.pagination-info b { color: #1e293b; }
.pagination-controls { display: flex; align-items: center; gap: 4px; }
.page-btn { min-width: 32px; height: 32px; padding: 0 8px; border: 1px solid #e5e7eb; border-radius: 6px; background: #fff; font-size: 13px; color: #374151; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.page-btn:hover:not(:disabled):not(.active) { border-color: #4f46e5; color: #4f46e5; }
.page-btn.active { background: linear-gradient(135deg, #4f46e5, #6d28d9); color: #fff; border-color: transparent; font-weight: 600; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-size-select { margin-left: 12px; padding: 6px 10px; border: 1px solid #e5e7eb; border-radius: 6px; font-size: 13px; background: #fff; }

.empty-state { text-align: center; padding: 80px 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; }
.empty-icon { font-size: 56px; margin-bottom: 12px; }
.empty-state h3 { color: #1a202c; margin: 0 0 8px; }
.empty-state p { color: #718096; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 16px; width: 90%; max-height: 85vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.2); }
.modal-sm { max-width: 420px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #f3f4f6; }
.modal-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.modal-body { padding: 24px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 24px; border-top: 1px solid #f3f4f6; }
.btn-outline { padding: 8px 18px; border: 1px solid #d1d5db; border-radius: 8px; background: #fff; font-size: 13px; color: #4b5563; cursor: pointer; }
.btn-danger { padding: 8px 18px; border: none; border-radius: 8px; background: linear-gradient(135deg, #dc2626, #b91c1c); color: #fff; font-size: 13px; cursor: pointer; font-weight: 500; }
.btn-danger:hover { box-shadow: 0 4px 12px rgba(220,38,38,0.3); }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

/* Drawer */
.drawer-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 1000; display: flex; justify-content: flex-end; }
.drawer-panel { width: 480px; background: #fff; height: 100%; overflow-y: auto; box-shadow: -4px 0 20px rgba(0,0,0,0.1); }
.drawer-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #e8ecf0; }
.drawer-header h3 { margin: 0; font-size: 18px; }
.modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #a0aec0; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 8px; }
.modal-close:hover { background: #f3f4f6; }
.drawer-body { padding: 24px; }
.detail-section { margin-bottom: 24px; }
.detail-section h4 { margin: 0 0 8px; font-size: 14px; color: #4a5568; }
.detail-row { display: flex; align-items: center; padding: 8px 0; border-bottom: 1px solid #f1f5f9; }
.detail-label { width: 100px; font-size: 13px; color: #718096; flex-shrink: 0; }
.detail-value { font-size: 14px; color: #1a202c; }
.detail-value.mono { font-family: 'Fira Code', monospace; }
.detail-text { font-size: 14px; color: #4a5568; line-height: 1.6; }
.detail-code { background: #f7f8fa; padding: 12px; border-radius: 8px; font-size: 12px; font-family: monospace; overflow-x: auto; white-space: pre-wrap; }
</style>

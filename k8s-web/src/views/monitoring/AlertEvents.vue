<template>
  <div class="alert-events-page">
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
      <select v-model="filters.severity" @change="loadList" class="filter-select">
        <option value="">全部级别</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </select>
      <select v-model="filters.status" @change="loadList" class="filter-select">
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
          <tr v-for="ev in list" :key="ev.id" @click="openDetail(ev)" class="clickable-row">
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

      <!-- 分页 -->
      <div class="pagination" v-if="total > pageSize">
        <button class="page-btn" :disabled="page <= 1" @click="page--; loadList()">‹ 上一页</button>
        <span class="page-info">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button class="page-btn" :disabled="page >= Math.ceil(total / pageSize)" @click="page++; loadList()">下一页 ›</button>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-else>
      <div class="empty-icon">🎉</div>
      <h3>暂无告警事件</h3>
      <p>当告警规则检测到异常时，事件将自动记录在此</p>
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
import { ref, reactive, onMounted } from 'vue'
import { listAlertEvents, getAlertStats, ackAlertEvent, resolveAlertEvent } from '@/api/monitoring'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const stats = reactive({ total_firing: 0, total_resolved: 0, critical: 0, warning: 0, info: 0 })
const detailEvent = ref(null)
const timeRange = ref('24h')
const filters = reactive({ keyword: '', severity: '', status: '', start_time: 0, end_time: 0 })
const statusMap = { firing: '告警中', resolved: '已恢复', silenced: '已静默' }

let debounceTimer = null
const debouncedLoad = () => { clearTimeout(debounceTimer); debounceTimer = setTimeout(loadList, 300) }

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

async function loadList() {
  try {
    const res = await listAlertEvents({ page: page.value, size: pageSize, ...filters })
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

onMounted(() => { applyTimeRange(); loadStats() })
</script>

<style scoped>
.alert-events-page { padding: 24px; }

.stats-bar { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }
.stat-card { background: #fff; border-radius: 12px; padding: 16px 20px; border: 1px solid #e8ecf0; display: flex; flex-direction: column; align-items: center; gap: 4px; }
.stat-card.critical { border-left: 3px solid #dc2626; }
.stat-card.warning { border-left: 3px solid #d97706; }
.stat-card.info { border-left: 3px solid #2563eb; }
.stat-card.resolved { border-left: 3px solid #059669; }
.stat-value { font-size: 28px; font-weight: 700; color: #1a202c; }
.stat-label { font-size: 13px; color: #718096; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.search-input { flex: 1; min-width: 200px; max-width: 300px; padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #4f46e5; }
.filter-select { padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: #fff; }

.events-table-wrapper { background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: #f7f8fa; padding: 12px 16px; text-align: left; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e8ecf0; }
.data-table td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; }
.clickable-row { cursor: pointer; }
.clickable-row:hover { background: #f7fafc; }
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

.pagination { display: flex; align-items: center; justify-content: center; gap: 16px; padding: 16px; }
.page-btn { padding: 6px 14px; border: 1px solid #e2e8f0; border-radius: 6px; background: #fff; cursor: pointer; font-size: 13px; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-btn:not(:disabled):hover { border-color: #4f46e5; color: #4f46e5; }
.page-info { font-size: 13px; color: #718096; }

.empty-state { text-align: center; padding: 80px 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; }
.empty-icon { font-size: 56px; margin-bottom: 12px; }
.empty-state h3 { color: #1a202c; margin: 0 0 8px; }
.empty-state p { color: #718096; }

/* Drawer */
.drawer-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 1000; display: flex; justify-content: flex-end; }
.drawer-panel { width: 480px; background: #fff; height: 100%; overflow-y: auto; box-shadow: -4px 0 20px rgba(0,0,0,0.1); }
.drawer-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #e8ecf0; }
.drawer-header h3 { margin: 0; font-size: 18px; }
.modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #a0aec0; }
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

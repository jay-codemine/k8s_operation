<template>
  <div class="ai-approvals-page">
    <!-- 页面头部 Banner -->
    <div class="page-banner">
      <div class="banner-content">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 2a4 4 0 0 1 4 4v1a3 3 0 0 1 3 3v1a2 2 0 0 1-2 2h-1l-1 7H9l-1-7H7a2 2 0 0 1-2-2v-1a3 3 0 0 1 3-3V6a4 4 0 0 1 4-4z"/>
              <circle cx="9" cy="9" r="1" fill="currentColor"/>
              <circle cx="15" cy="9" r="1" fill="currentColor"/>
              <path d="M9 13h6"/>
            </svg>
          </div>
          <div>
            <h1 class="banner-title">AI 智能审批</h1>
            <p class="banner-desc">管理 AI 助手高危操作审批请求，保障集群安全运维合规</p>
          </div>
        </div>
        <div class="banner-actions">
          <div class="role-badge" :class="isAdmin ? 'admin' : 'user'">
            <span class="role-dot"></span>
            {{ isAdmin ? '审批管理员' : '普通用户' }}
          </div>
          <button class="btn-refresh" @click="loadData" :disabled="loading">
            <svg class="refresh-icon" :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 统计指标卡片 -->
    <div class="metrics-row">
      <div class="metric-card" v-for="m in metrics" :key="m.key" :class="[m.key, { active: statusFilter === m.filterVal }]" @click="setFilter(m.filterVal)">
        <div class="metric-icon-wrap" :class="m.key">
          <component :is="m.icon" />
        </div>
        <div class="metric-body">
          <span class="metric-value">{{ m.count }}</span>
          <span class="metric-label">{{ m.label }}</span>
        </div>
        <div class="metric-tag" :class="m.key">{{ m.tag }}</div>
      </div>
    </div>

    <!-- 内容面板 -->
    <div class="content-panel">
      <!-- 工具栏 -->
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h3 class="panel-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:18px;height:18px;margin-right:6px;vertical-align:-3px;">
              <rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/>
            </svg>
            审批记录
          </h3>
          <span class="record-count">{{ filteredList.length }} 条</span>
        </div>
        <div class="toolbar-right">
          <div class="search-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="searchQuery" placeholder="搜索操作、资源、用户..." />
          </div>
        </div>
      </div>

      <!-- 加载态 -->
      <div v-if="loading" class="loading-state">
        <div class="loader"><div class="loader-ring"></div><div class="loader-ring"></div><div class="loader-ring"></div></div>
        <span>正在加载审批数据...</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="displayedList.length === 0" class="empty-state">
        <svg viewBox="0 0 200 160" fill="none" style="width:160px;height:128px;">
          <rect x="40" y="20" width="120" height="100" rx="8" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
          <rect x="56" y="44" width="88" height="8" rx="4" fill="#d6e4ff"/>
          <rect x="56" y="60" width="60" height="8" rx="4" fill="#d6e4ff"/>
          <circle cx="100" cy="130" r="20" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
          <path d="M92 130l5 5 11-11" stroke="#6366f1" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <h3>暂无审批记录</h3>
        <p>当 AI 助手检测到高危操作时，审批请求会自动出现在这里</p>
      </div>

      <!-- 表格 -->
      <div v-else class="table-wrapper">
        <table class="approval-table">
          <thead>
            <tr>
              <th class="col-id">ID</th>
              <th class="col-status">状态</th>
              <th class="col-intent">操作意图</th>
              <th class="col-resource">目标资源</th>
              <th class="col-risk">风险</th>
              <th class="col-user">申请人</th>
              <th class="col-time">申请时间</th>
              <th class="col-approver">审批人</th>
              <th class="col-comment">备注</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in displayedList" :key="item.id" :class="'row-' + statusKey(item.status)" @click="openDetail(item)">
              <td class="col-id"><span class="id-badge">#{{ item.id }}</span></td>
              <td class="col-status">
                <span class="status-pill" :class="statusKey(item.status)">
                  <span class="status-dot"></span>
                  {{ statusLabel(item.status) }}
                </span>
              </td>
              <td class="col-intent">
                <div class="intent-cell">
                  <span class="intent-tag" :class="item.intent">{{ item.intent }}</span>
                  <span class="tool-name" v-if="item.tool_name">{{ item.tool_name }}</span>
                </div>
              </td>
              <td class="col-resource">
                <div class="resource-cell">
                  <code class="resource-name">{{ item.namespace ? item.namespace + '/' : '' }}{{ item.resource_name || '-' }}</code>
                  <span class="resource-type" v-if="item.resource">{{ item.resource }}</span>
                </div>
              </td>
              <td class="col-risk">
                <span class="risk-badge" :class="item.risk_level">{{ riskLabel(item.risk_level) }}</span>
              </td>
              <td class="col-user">
                <div class="user-cell" v-if="item.request_user_name">
                  <div class="avatar">{{ item.request_user_name.charAt(0).toUpperCase() }}</div>
                  <span>{{ item.request_user_name }}</span>
                </div>
                <span v-else class="text-muted">UID:{{ item.request_user_id }}</span>
              </td>
              <td class="col-time"><span class="time-text">{{ formatTime(item.created_at) }}</span></td>
              <td class="col-approver">
                <div class="user-cell" v-if="item.approver_user_name">
                  <div class="avatar approver">{{ item.approver_user_name.charAt(0).toUpperCase() }}</div>
                  <span>{{ item.approver_user_name }}</span>
                </div>
                <span v-else class="text-muted">-</span>
              </td>
              <td class="col-comment">
                <span class="comment-text" :title="item.approve_comment">{{ item.approve_comment || '-' }}</span>
              </td>
              <td class="col-actions" @click.stop>
                <!-- 管理员：待审批/已过期状态显示通过/拒绝按钮 -->
                <template v-if="(item.status === 1 || item.status === 4) && isAdmin">
                  <button class="act-btn approve" @click="openAction(item, 'approve')" title="通过">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                  </button>
                  <button class="act-btn reject" @click="openAction(item, 'reject')" title="拒绝">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </template>
                <!-- 待审批/已过期状态：编辑备注 -->
                <button v-if="item.status === 1 || item.status === 4" class="act-btn edit" @click="openEdit(item)" title="编辑备注">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                </button>
                <!-- 非已通过状态：删除 -->
                <button class="act-btn delete" @click="handleDelete(item)" title="删除" v-if="item.status !== 2">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
                <!-- 非管理员：待审批状态显示取消按钮 -->
                <button v-if="item.status === 1 && !isAdmin" class="act-btn cancel" @click="handleCancel(item)" title="取消">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M8 12h8"/></svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 审批/拒绝弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showActionModal" class="modal-overlay" @click="closeAction">
          <div class="modal-dialog" @click.stop>
            <div class="modal-head" :class="actionType">
              <div class="modal-head-icon">
                <svg v-if="actionType==='approve'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
              </div>
              <h3>{{ actionType === 'approve' ? '确认通过审批' : '确认拒绝审批' }}</h3>
              <button class="modal-close" @click="closeAction">✕</button>
            </div>
            <div class="modal-main">
              <div class="summary-card" v-if="currentItem">
                <div class="summary-row"><span class="s-label">操作意图</span><span class="s-value intent-tag" :class="currentItem.intent">{{ currentItem.intent }}</span></div>
                <div class="summary-row"><span class="s-label">目标资源</span><code class="s-value">{{ currentItem.namespace }}/{{ currentItem.resource_name }}</code></div>
                <div class="summary-row"><span class="s-label">风险等级</span><span class="risk-badge" :class="currentItem.risk_level">{{ riskLabel(currentItem.risk_level) }}</span></div>
                <div class="summary-row"><span class="s-label">摘要</span><span class="s-value">{{ currentItem.summary || '-' }}</span></div>
              </div>
              <div class="form-field">
                <label>审批意见 <span class="optional">(选填)</span></label>
                <textarea v-model="actionComment" :placeholder="actionType==='approve' ? '输入通过意见...' : '请说明拒绝原因...'" rows="3"></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="closeAction">取消</button>
              <button class="btn-confirm" :class="actionType" @click="submitAction" :disabled="actionLoading">
                {{ actionLoading ? '处理中...' : (actionType === 'approve' ? '确认通过' : '确认拒绝') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 编辑备注弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showEditModal" class="modal-overlay" @click="closeEdit">
          <div class="modal-dialog small" @click.stop>
            <div class="modal-head edit">
              <h3>编辑审批备注</h3>
              <button class="modal-close" @click="closeEdit">✕</button>
            </div>
            <div class="modal-main">
              <div class="form-field">
                <label>备注内容</label>
                <textarea v-model="editComment" placeholder="输入备注信息..." rows="3"></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="closeEdit">取消</button>
              <button class="btn-confirm edit" @click="submitEdit" :disabled="editLoading">
                {{ editLoading ? '保存中...' : '保存' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 详情弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDetailModal" class="modal-overlay" @click="closeDetail">
          <div class="modal-dialog wide" @click.stop>
            <div class="modal-head detail">
              <h3>审批详情 #{{ detailItem?.id }}</h3>
              <button class="modal-close" @click="closeDetail">✕</button>
            </div>
            <div class="modal-main" v-if="detailItem">
              <div class="detail-grid">
                <div class="detail-item"><span class="d-label">状态</span><span class="status-pill" :class="statusKey(detailItem.status)"><span class="status-dot"></span>{{ statusLabel(detailItem.status) }}</span></div>
                <div class="detail-item"><span class="d-label">操作意图</span><span class="intent-tag" :class="detailItem.intent">{{ detailItem.intent }}</span></div>
                <div class="detail-item"><span class="d-label">工具名称</span><code>{{ detailItem.tool_name || '-' }}</code></div>
                <div class="detail-item"><span class="d-label">风险等级</span><span class="risk-badge" :class="detailItem.risk_level">{{ riskLabel(detailItem.risk_level) }}</span></div>
                <div class="detail-item"><span class="d-label">命名空间</span><code>{{ detailItem.namespace || '-' }}</code></div>
                <div class="detail-item"><span class="d-label">资源名称</span><code>{{ detailItem.resource_name || '-' }}</code></div>
                <div class="detail-item"><span class="d-label">集群ID</span><span>{{ detailItem.cluster_id || '-' }}</span></div>
                <div class="detail-item"><span class="d-label">申请人</span><span>{{ detailItem.request_user_name || 'UID:' + detailItem.request_user_id }}</span></div>
                <div class="detail-item"><span class="d-label">审批人</span><span>{{ detailItem.approver_user_name || '-' }}</span></div>
                <div class="detail-item"><span class="d-label">申请时间</span><span>{{ formatTime(detailItem.created_at) }}</span></div>
                <div class="detail-item"><span class="d-label">过期时间</span><span>{{ detailItem.expire_at ? formatTime(detailItem.expire_at) : '-' }}</span></div>
                <div class="detail-item"><span class="d-label">是否已执行</span><span>{{ detailItem.executed ? '✅ 已执行' : '❌ 未执行' }}</span></div>
              </div>
              <div class="detail-section" v-if="detailItem.summary">
                <h4>操作摘要</h4>
                <p class="detail-text">{{ detailItem.summary }}</p>
              </div>
              <div class="detail-section" v-if="detailItem.tool_args_json">
                <h4>工具参数</h4>
                <pre class="detail-code">{{ formatJSON(detailItem.tool_args_json) }}</pre>
              </div>
              <div class="detail-section" v-if="detailItem.execute_result">
                <h4>执行结果</h4>
                <pre class="detail-code result">{{ detailItem.execute_result }}</pre>
              </div>
              <div class="detail-section" v-if="detailLogs.length > 0">
                <h4>操作日志</h4>
                <div class="log-timeline">
                  <div class="log-item" v-for="log in detailLogs" :key="log.id">
                    <span class="log-dot" :class="log.action"></span>
                    <span class="log-action">{{ log.action }}</span>
                    <span class="log-comment">{{ log.comment }}</span>
                    <span class="log-time">{{ formatTime(log.created_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import {
  getApprovals, getApprovalDetail, approveApproval, rejectApproval,
  cancelApproval, deleteApproval, updateApproval, getApprovalStats
} from '@/api/ai'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// ====== State ======
const loading = ref(true)
const actionLoading = ref(false)
const editLoading = ref(false)
const list = ref([])
const statusFilter = ref(0)
const searchQuery = ref('')
const stats = ref({ pending: 0, approved: 0, rejected: 0, expired: 0, canceled: 0 })

// 权限判断：使用前端 permissionStore（不依赖后端 is_admin 字段）
const isAdmin = computed(() => {
  return permissionStore.state.isSuperAdmin ||
    permissionStore.isAdmin.value ||
    permissionStore.isClusterAdmin.value
})

// action modal
const showActionModal = ref(false)
const currentItem = ref(null)
const actionType = ref('approve')
const actionComment = ref('')

// edit modal
const showEditModal = ref(false)
const editItem = ref(null)
const editComment = ref('')

// detail modal
const showDetailModal = ref(false)
const detailItem = ref(null)
const detailLogs = ref([])

// ====== Metrics ======
const totalCount = computed(() => list.value.length)
const metrics = computed(() => [
  { key: 'total', label: '全部', count: totalCount.value, tag: 'ALL', filterVal: 0, icon: iconGrid },
  { key: 'pending', label: '待审批', count: stats.value.pending || 0, tag: 'WAIT', filterVal: 1, icon: iconClock },
  { key: 'approved', label: '已通过', count: stats.value.approved || 0, tag: 'PASS', filterVal: 2, icon: iconCheck },
  { key: 'rejected', label: '已拒绝', count: stats.value.rejected || 0, tag: 'DENY', filterVal: 3, icon: iconX },
  { key: 'canceled', label: '已取消', count: stats.value.canceled || 0, tag: 'OFF', filterVal: 5, icon: iconMinus },
])

// icon components
const iconGrid = { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [h('rect', { x: 3, y: 3, width: 18, height: 18, rx: 2 }), h('path', { d: 'M3 9h18' }), h('path', { d: 'M9 21V9' })]) }
const iconClock = { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [h('circle', { cx: 12, cy: 12, r: 10 }), h('polyline', { points: '12 6 12 12 16 14' })]) }
const iconCheck = { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [h('path', { d: 'M22 11.08V12a10 10 0 1 1-5.93-9.14' }), h('polyline', { points: '22 4 12 14.01 9 11.01' })]) }
const iconX = { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [h('circle', { cx: 12, cy: 12, r: 10 }), h('line', { x1: 15, y1: 9, x2: 9, y2: 15 }), h('line', { x1: 9, y1: 9, x2: 15, y2: 15 })]) }
const iconMinus = { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [h('circle', { cx: 12, cy: 12, r: 10 }), h('path', { d: 'M8 12h8' })]) }

// ====== Computed ======
const filteredList = computed(() => {
  let l = list.value
  if (statusFilter.value > 0) l = l.filter(i => i.status === statusFilter.value)
  return l
})
const displayedList = computed(() => {
  if (!searchQuery.value.trim()) return filteredList.value
  const q = searchQuery.value.toLowerCase()
  return filteredList.value.filter(i =>
    (i.intent || '').toLowerCase().includes(q) ||
    (i.resource_name || '').toLowerCase().includes(q) ||
    (i.namespace || '').toLowerCase().includes(q) ||
    (i.tool_name || '').toLowerCase().includes(q) ||
    (i.request_user_name || '').toLowerCase().includes(q) ||
    (i.summary || '').toLowerCase().includes(q) ||
    String(i.id).includes(q)
  )
})

// ====== Helpers ======
const statusKey = (s) => ({ 1: 'pending', 2: 'approved', 3: 'rejected', 4: 'expired', 5: 'canceled' }[s] || 'unknown')
const statusLabel = (s) => ({ 1: '待审批', 2: '已通过', 3: '已拒绝', 4: '已过期', 5: '已取消' }[s] || '未知')
const riskLabel = (r) => ({ low: '低', medium: '中', high: '高', critical: '严重' }[r] || r)
const formatTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
const formatJSON = (s) => { try { return JSON.stringify(JSON.parse(s), null, 2) } catch { return s } }
const setFilter = (v) => { statusFilter.value = statusFilter.value === v ? 0 : v }

// ====== Data Loading ======
const loadData = async () => {
  loading.value = true
  try {
    const [listRes, statsRes] = await Promise.all([
      getApprovals({ page: 1, page_size: 200, view: 'all' }),
      getApprovalStats()
    ])
    if (listRes.code === 0) {
      list.value = listRes.data?.list || []
    }
    if (statsRes.code === 0) {
      stats.value = statsRes.data?.stats || {}
    }
  } catch (err) {
    console.error('加载审批数据失败:', err)
  } finally {
    loading.value = false
  }
}

// ====== Actions ======
const openAction = (item, type) => { currentItem.value = item; actionType.value = type; actionComment.value = ''; showActionModal.value = true }
const closeAction = () => { showActionModal.value = false }
const submitAction = async () => {
  if (!currentItem.value) return
  actionLoading.value = true
  try {
    const fn = actionType.value === 'approve' ? approveApproval : rejectApproval
    const res = await fn(currentItem.value.id, { comment: actionComment.value, admin_override: true })
    if (res.code === 0) { closeAction(); loadData() } else {
      alert(res.details?.[0] || res.msg || '操作失败')
    }
  } catch (e) { alert(e.message || '操作失败') } finally { actionLoading.value = false }
}

const openEdit = (item) => { editItem.value = item; editComment.value = item.approve_comment || ''; showEditModal.value = true }
const closeEdit = () => { showEditModal.value = false }
const submitEdit = async () => {
  if (!editItem.value) return
  editLoading.value = true
  try {
    const res = await updateApproval(editItem.value.id, { comment: editComment.value })
    if (res.code === 0) { closeEdit(); loadData() } else { alert(res.msg || '更新失败') }
  } catch (e) { alert(e.message || '更新失败') } finally { editLoading.value = false }
}

const handleDelete = async (item) => {
  const ok = await showConfirm({
    title: '确认删除审批',
    content: '此操作不可恢复。',
    type: 'danger',
    details: [
      { label: '审批 ID', value: `#${item.id}`, mono: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  try {
    const res = await deleteApproval(item.id)
    if (res.code === 0) loadData(); else alert(res.msg || '删除失败')
  } catch (e) { alert(e.message || '删除失败') }
}

const handleCancel = async (item) => {
  const ok = await showConfirm({
    title: '确认取消审批',
    type: 'warning',
    details: [
      { label: '审批 ID', value: `#${item.id}`, mono: true },
    ],
    confirmText: '确认取消',
    cancelText: '返回',
  })
  if (!ok) return
  try {
    const res = await cancelApproval(item.id)
    if (res.code === 0) loadData(); else alert(res.msg || '取消失败')
  } catch (e) { alert(e.message || '取消失败') }
}

const openDetail = async (item) => {
  detailItem.value = item
  detailLogs.value = []
  showDetailModal.value = true
  try {
    const res = await getApprovalDetail(item.id)
    if (res.code === 0) {
      detailItem.value = { ...item, ...res.data?.approval }
      detailLogs.value = res.data?.logs || []
    }
  } catch (e) { console.error('加载详情失败', e) }
}
const closeDetail = () => { showDetailModal.value = false }

onMounted(loadData)
</script>

<style scoped>
/* ===== Page Layout ===== */
.ai-approvals-page { min-height: 100%; }

/* ===== Banner ===== */
.page-banner {
  background: linear-gradient(135deg, #1e1b4b 0%, #312e81 40%, #4338ca 100%);
  border-radius: 12px;
  padding: 1.5rem 2rem;
  margin-bottom: 1.25rem;
  position: relative;
  overflow: hidden;
}
.page-banner::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(139,92,246,0.3) 0%, transparent 70%);
  pointer-events: none;
}
.banner-content { display: flex; align-items: center; justify-content: space-between; position: relative; z-index: 1; }
.banner-left { display: flex; align-items: center; gap: 1rem; }
.banner-icon {
  width: 3rem; height: 3rem;
  background: rgba(255,255,255,0.15);
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  backdrop-filter: blur(8px);
}
.banner-icon svg { width: 1.75rem; height: 1.75rem; color: #c4b5fd; }
.banner-title { font-size: 1.375rem; font-weight: 700; color: #fff; margin: 0; }
.banner-desc { font-size: 0.8125rem; color: #a5b4fc; margin-top: 4px; }
.banner-actions { display: flex; align-items: center; gap: 0.75rem; }
.role-badge {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}
.role-badge.admin { background: rgba(16,185,129,0.2); color: #6ee7b7; }
.role-badge.user { background: rgba(139,92,246,0.2); color: #c4b5fd; }
.role-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  animation: pulse-dot 2s infinite;
}
.role-badge.admin .role-dot { background: #10b981; }
.role-badge.user .role-dot { background: #8b5cf6; }
@keyframes pulse-dot { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
.btn-refresh {
  width: 36px; height: 36px;
  border: 1px solid rgba(255,255,255,0.2);
  background: rgba(255,255,255,0.1);
  border-radius: 8px;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}
.btn-refresh:hover { background: rgba(255,255,255,0.2); }
.btn-refresh svg { width: 18px; height: 18px; color: #e0e7ff; }
.refresh-icon.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* ===== Metrics ===== */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.metric-card {
  background: #fff;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.875rem;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.25s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.metric-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.metric-card.active { border-color: #6366f1; background: #f5f3ff; }
.metric-icon-wrap {
  width: 2.5rem; height: 2.5rem;
  border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.metric-icon-wrap svg { width: 1.25rem; height: 1.25rem; }
.metric-icon-wrap.total { background: #ede9fe; color: #7c3aed; }
.metric-icon-wrap.pending { background: #fef3c7; color: #d97706; }
.metric-icon-wrap.approved { background: #d1fae5; color: #059669; }
.metric-icon-wrap.rejected { background: #fee2e2; color: #dc2626; }
.metric-icon-wrap.canceled { background: #e5e7eb; color: #6b7280; }
.metric-body { flex: 1; }
.metric-value { display: block; font-size: 1.5rem; font-weight: 700; color: #1e293b; line-height: 1.2; }
.metric-label { font-size: 0.75rem; color: #64748b; }
.metric-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.metric-tag.total { background: #ede9fe; color: #7c3aed; }
.metric-tag.pending { background: #fef3c7; color: #d97706; }
.metric-tag.approved { background: #d1fae5; color: #059669; }
.metric-tag.rejected { background: #fee2e2; color: #dc2626; }
.metric-tag.canceled { background: #e5e7eb; color: #6b7280; }

/* ===== Content Panel ===== */
.content-panel {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  overflow: hidden;
}
.panel-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
}
.toolbar-left { display: flex; align-items: center; gap: 0.75rem; }
.panel-title { font-size: 0.9375rem; font-weight: 600; color: #1e293b; margin: 0; display: flex; align-items: center; }
.record-count { font-size: 0.75rem; color: #94a3b8; background: #f1f5f9; padding: 2px 10px; border-radius: 10px; }
.search-box {
  display: flex; align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0 12px;
  transition: border-color 0.2s;
}
.search-box:focus-within { border-color: #6366f1; }
.search-box svg { width: 16px; height: 16px; color: #94a3b8; flex-shrink: 0; }
.search-box input {
  border: none; background: none; outline: none;
  padding: 8px; font-size: 0.8125rem; color: #334155; width: 220px;
}

/* ===== Loading & Empty ===== */
.loading-state { display: flex; flex-direction: column; align-items: center; padding: 3rem; gap: 1rem; color: #94a3b8; }
.loader { display: flex; gap: 6px; }
.loader-ring { width: 10px; height: 10px; border-radius: 50%; background: #6366f1; animation: bounce 1.4s infinite ease-in-out both; }
.loader-ring:nth-child(1) { animation-delay: -0.32s; }
.loader-ring:nth-child(2) { animation-delay: -0.16s; }
@keyframes bounce { 0%,80%,100% { transform: scale(0); } 40% { transform: scale(1); } }
.empty-state { display: flex; flex-direction: column; align-items: center; padding: 4rem 2rem; color: #94a3b8; }
.empty-state h3 { color: #475569; margin: 1rem 0 0.5rem; font-size: 1rem; }
.empty-state p { font-size: 0.8125rem; }

/* ===== Table ===== */
.table-wrapper { overflow-x: auto; }
.approval-table { width: 100%; border-collapse: collapse; font-size: 0.8125rem; }
.approval-table thead { background: #f8fafc; }
.approval-table th {
  padding: 0.75rem 1rem; text-align: left;
  font-weight: 600; color: #64748b; font-size: 0.75rem;
  text-transform: uppercase; letter-spacing: 0.5px;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}
.approval-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  color: #334155;
  vertical-align: middle;
}
.approval-table tbody tr { cursor: pointer; transition: background 0.15s; }
.approval-table tbody tr:hover { background: #f8fafc; }
.approval-table tbody tr.row-pending { border-left: 3px solid #f59e0b; }
.approval-table tbody tr.row-approved { border-left: 3px solid #10b981; }
.approval-table tbody tr.row-rejected { border-left: 3px solid #ef4444; }
.approval-table tbody tr.row-canceled { border-left: 3px solid #9ca3af; }
.approval-table tbody tr.row-expired { border-left: 3px solid #6b7280; }

.id-badge { font-weight: 600; color: #6366f1; }
.status-pill {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 3px 10px; border-radius: 12px; font-size: 0.6875rem; font-weight: 600;
}
.status-dot { width: 6px; height: 6px; border-radius: 50%; }
.status-pill.pending { background: #fef3c7; color: #92400e; }
.status-pill.pending .status-dot { background: #f59e0b; }
.status-pill.approved { background: #d1fae5; color: #065f46; }
.status-pill.approved .status-dot { background: #10b981; }
.status-pill.rejected { background: #fee2e2; color: #991b1b; }
.status-pill.rejected .status-dot { background: #ef4444; }
.status-pill.canceled { background: #f3f4f6; color: #4b5563; }
.status-pill.canceled .status-dot { background: #9ca3af; }
.status-pill.expired { background: #f3f4f6; color: #6b7280; }
.status-pill.expired .status-dot { background: #6b7280; }

.intent-cell { display: flex; flex-direction: column; gap: 2px; }
.intent-tag {
  display: inline-block; padding: 2px 8px; border-radius: 4px;
  font-size: 0.6875rem; font-weight: 600; font-family: 'SF Mono', monospace;
}
.intent-tag.delete_pod, .intent-tag.delete_deployment, .intent-tag.drain_node { background: #fee2e2; color: #dc2626; }
.intent-tag.scale_deployment, .intent-tag.restart_deployment { background: #fef3c7; color: #d97706; }
.intent-tag { background: #ede9fe; color: #7c3aed; }
.tool-name { font-size: 0.6875rem; color: #94a3b8; }

.resource-cell { display: flex; flex-direction: column; gap: 2px; }
.resource-name { font-size: 0.8125rem; color: #334155; background: #f8fafc; padding: 1px 6px; border-radius: 3px; }
.resource-type { font-size: 0.6875rem; color: #94a3b8; }

.risk-badge { padding: 2px 8px; border-radius: 4px; font-size: 0.6875rem; font-weight: 600; }
.risk-badge.low { background: #d1fae5; color: #065f46; }
.risk-badge.medium { background: #fef3c7; color: #92400e; }
.risk-badge.high { background: #fee2e2; color: #991b1b; }
.risk-badge.critical { background: #fce7f3; color: #9d174d; }

.user-cell { display: flex; align-items: center; gap: 6px; }
.avatar {
  width: 24px; height: 24px; border-radius: 6px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff; font-size: 0.6875rem; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.avatar.approver { background: linear-gradient(135deg, #10b981, #34d399); }
.text-muted { color: #cbd5e1; }
.time-text { font-size: 0.75rem; color: #64748b; white-space: nowrap; }
.comment-text { font-size: 0.75rem; color: #64748b; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; }

/* Action buttons */
.act-btn {
  width: 28px; height: 28px; border: none; border-radius: 6px; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center;
  transition: all 0.2s; margin-right: 4px;
}
.act-btn svg { width: 14px; height: 14px; }
.act-btn.approve { background: #d1fae5; color: #059669; }
.act-btn.approve:hover { background: #10b981; color: #fff; }
.act-btn.reject { background: #fee2e2; color: #dc2626; }
.act-btn.reject:hover { background: #ef4444; color: #fff; }
.act-btn.edit { background: #e0e7ff; color: #4f46e5; }
.act-btn.edit:hover { background: #6366f1; color: #fff; }
.act-btn.delete { background: #f3f4f6; color: #9ca3af; }
.act-btn.delete:hover { background: #ef4444; color: #fff; }
.act-btn.cancel { background: #fef3c7; color: #d97706; }
.act-btn.cancel:hover { background: #f59e0b; color: #fff; }

/* ===== Modals ===== */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 9999; backdrop-filter: blur(4px);
}
.modal-dialog {
  background: #fff; border-radius: 16px; width: 480px; max-height: 80vh;
  box-shadow: 0 25px 50px rgba(0,0,0,0.25); overflow: hidden;
  display: flex; flex-direction: column;
}
.modal-dialog.small { width: 400px; }
.modal-dialog.wide { width: 680px; }
.modal-head {
  padding: 1.25rem 1.5rem;
  display: flex; align-items: center; gap: 0.75rem;
}
.modal-head.approve { background: linear-gradient(135deg, #d1fae5, #a7f3d0); }
.modal-head.reject { background: linear-gradient(135deg, #fee2e2, #fecaca); }
.modal-head.edit { background: linear-gradient(135deg, #e0e7ff, #c7d2fe); }
.modal-head.detail { background: linear-gradient(135deg, #ede9fe, #ddd6fe); }
.modal-head h3 { flex: 1; font-size: 1rem; font-weight: 600; color: #1e293b; margin: 0; }
.modal-head-icon { width: 2rem; height: 2rem; }
.modal-head-icon svg { width: 100%; height: 100%; }
.modal-head.approve .modal-head-icon { color: #059669; }
.modal-head.reject .modal-head-icon { color: #dc2626; }
.modal-close { background: none; border: none; cursor: pointer; font-size: 1.25rem; color: #64748b; padding: 4px; }
.modal-close:hover { color: #1e293b; }
.modal-main { padding: 1.25rem 1.5rem; overflow-y: auto; flex: 1; }
.modal-foot { padding: 1rem 1.5rem; display: flex; justify-content: flex-end; gap: 0.75rem; border-top: 1px solid #f1f5f9; }

.summary-card { background: #f8fafc; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
.summary-row { display: flex; align-items: center; justify-content: space-between; padding: 0.375rem 0; }
.s-label { font-size: 0.75rem; color: #64748b; }
.s-value { font-size: 0.8125rem; color: #334155; }
.form-field { margin-bottom: 0.75rem; }
.form-field label { display: block; font-size: 0.8125rem; font-weight: 500; color: #475569; margin-bottom: 0.375rem; }
.optional { color: #94a3b8; font-weight: 400; }
.form-field textarea {
  width: 100%; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.625rem;
  font-size: 0.8125rem; color: #334155; resize: vertical; outline: none;
  transition: border-color 0.2s; box-sizing: border-box;
}
.form-field textarea:focus { border-color: #6366f1; }
.btn-cancel {
  padding: 0.5rem 1.25rem; border: 1px solid #e2e8f0; background: #fff;
  border-radius: 8px; font-size: 0.8125rem; color: #64748b; cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover { background: #f8fafc; }
.btn-confirm {
  padding: 0.5rem 1.5rem; border: none; border-radius: 8px;
  font-size: 0.8125rem; font-weight: 600; color: #fff; cursor: pointer;
  transition: all 0.2s;
}
.btn-confirm.approve { background: linear-gradient(135deg, #10b981, #059669); }
.btn-confirm.approve:hover { background: linear-gradient(135deg, #059669, #047857); }
.btn-confirm.reject { background: linear-gradient(135deg, #ef4444, #dc2626); }
.btn-confirm.reject:hover { background: linear-gradient(135deg, #dc2626, #b91c1c); }
.btn-confirm.edit { background: linear-gradient(135deg, #6366f1, #4f46e5); }
.btn-confirm.edit:hover { background: linear-gradient(135deg, #4f46e5, #4338ca); }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

/* ===== Detail Modal ===== */
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; margin-bottom: 1rem; }
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.d-label { font-size: 0.6875rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; }
.detail-item code { font-size: 0.8125rem; background: #f8fafc; padding: 2px 6px; border-radius: 4px; }
.detail-section { margin-top: 1rem; }
.detail-section h4 { font-size: 0.8125rem; font-weight: 600; color: #475569; margin: 0 0 0.5rem; }
.detail-text { font-size: 0.8125rem; color: #334155; line-height: 1.6; }
.detail-code {
  background: #1e293b; color: #e2e8f0; padding: 1rem; border-radius: 8px;
  font-size: 0.75rem; line-height: 1.5; overflow-x: auto; white-space: pre-wrap;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.detail-code.result { background: #064e3b; color: #a7f3d0; }

.log-timeline { display: flex; flex-direction: column; gap: 0.5rem; }
.log-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.75rem; padding: 0.375rem 0; }
.log-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.log-dot.create { background: #6366f1; }
.log-dot.approve { background: #10b981; }
.log-dot.reject { background: #ef4444; }
.log-dot.cancel { background: #f59e0b; }
.log-dot.execute { background: #3b82f6; }
.log-dot.delete { background: #9ca3af; }
.log-dot.update { background: #8b5cf6; }
.log-action { font-weight: 600; color: #475569; min-width: 48px; }
.log-comment { flex: 1; color: #64748b; }
.log-time { color: #94a3b8; white-space: nowrap; }

/* ===== Transitions ===== */
.modal-enter-active { animation: modal-in 0.3s ease; }
.modal-leave-active { animation: modal-in 0.2s ease reverse; }
@keyframes modal-in { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }

/* ===== Responsive ===== */
@media (max-width: 1200px) {
  .metrics-row { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .metrics-row { grid-template-columns: repeat(2, 1fr); }
  .page-banner { padding: 1rem 1.25rem; }
  .banner-title { font-size: 1.125rem; }
  .modal-dialog, .modal-dialog.wide { width: 95vw; }
}
</style>

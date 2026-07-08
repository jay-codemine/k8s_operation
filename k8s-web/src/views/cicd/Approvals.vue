<template>
  <div class="approvals-page">
    <!-- 页面头部 - Rancher 风格深色渐变 banner -->
    <div class="page-banner">
      <div class="banner-content">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 11l3 3L22 4"/>
              <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
            </svg>
          </div>
          <div>
            <h1 class="banner-title">部署审批管理</h1>
            <p class="banner-desc">管理生产环境部署审批申请，保障发布安全合规</p>
          </div>
        </div>
        <button class="btn-refresh-banner" @click="loadApprovals" :disabled="loading">
          <svg class="refresh-icon" :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="23 4 23 10 17 10"/>
            <polyline points="1 20 1 14 7 14"/>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
          <span>{{ loading ? '刷新中' : '刷新' }}</span>
        </button>
      </div>
    </div>

    <!-- 统计面板 - Kuboard 风格指标卡片 -->
    <div class="metrics-panel">
      <div class="metric-card" :class="{ active: statusFilter === '' }" @click="setFilter('')">
        <div class="metric-icon-wrapper total">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-value">{{ totalCount }}</span>
          <span class="metric-label">全部</span>
        </div>
        <div class="metric-trend total">ALL</div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'pending' }" @click="setFilter('pending')">
        <div class="metric-icon-wrapper pending">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-value">{{ pendingCount }}</span>
          <span class="metric-label">待审批</span>
        </div>
        <div class="metric-trend pending" v-if="pendingCount > 0">PENDING</div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'approved' }" @click="setFilter('approved')">
        <div class="metric-icon-wrapper approved">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-value">{{ approvedCount }}</span>
          <span class="metric-label">已通过</span>
        </div>
        <div class="metric-trend approved" v-if="approvedCount > 0">PASS</div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'rejected' }" @click="setFilter('rejected')">
        <div class="metric-icon-wrapper rejected">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-value">{{ rejectedCount }}</span>
          <span class="metric-label">已拒绝</span>
        </div>
        <div class="metric-trend rejected" v-if="rejectedCount > 0">DENY</div>
      </div>
    </div>

    <!-- 内容区 - 表格 -->
    <div class="content-panel">
      <!-- 工具栏 -->
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h3 class="panel-title">审批记录</h3>
          <span class="record-count">共 {{ filteredApprovals.length }} 条</span>
          <!-- 批量操作按钮 -->
          <template v-if="canApprove && selectedIds.length > 0">
            <button class="btn-batch approve" @click="handleBatchAction('approve')" :disabled="batchLoading">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
              <span>批量通过 ({{ selectedIds.length }})</span>
            </button>
            <button class="btn-batch reject" @click="handleBatchAction('reject')" :disabled="batchLoading">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              <span>批量拒绝 ({{ selectedIds.length }})</span>
            </button>
          </template>
        </div>
        <div class="toolbar-right">
          <button class="btn-create" @click="openCreateModal" v-if="canApprove">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            <span>新增审批</span>
          </button>
          <div class="search-box">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="searchQuery" placeholder="搜索流水线、镜像..." class="search-input" />
          </div>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <div class="loader">
          <div class="loader-ring"></div>
          <div class="loader-ring"></div>
          <div class="loader-ring"></div>
        </div>
        <span class="loading-text">正在加载审批数据...</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="displayedApprovals.length === 0" class="empty-state">
        <div class="empty-illustration">
          <svg viewBox="0 0 200 160" fill="none">
            <rect x="40" y="20" width="120" height="100" rx="8" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
            <rect x="56" y="44" width="88" height="8" rx="4" fill="#d6e4ff"/>
            <rect x="56" y="60" width="60" height="8" rx="4" fill="#d6e4ff"/>
            <rect x="56" y="76" width="74" height="8" rx="4" fill="#d6e4ff"/>
            <circle cx="100" cy="130" r="20" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
            <path d="M92 130l5 5 11-11" stroke="#4e7cf6" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h3 class="empty-title">暂无审批记录</h3>
        <p class="empty-desc">当流水线触发部署审批时，审批记录会显示在这里</p>
      </div>

      <!-- 表格列表 - Rancher 风格 -->
      <div v-else class="approval-table-wrapper">
        <table class="approval-table">
          <thead>
            <tr>
              <th class="col-check" v-if="canApprove">
                <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" title="全选/取消全选待审批" />
              </th>
              <th class="col-id">ID</th>
              <th class="col-status">状态</th>
              <th class="col-pipeline">流水线</th>
              <th class="col-env">环境</th>
              <th class="col-image">镜像</th>
              <th class="col-user">申请人</th>
              <th class="col-time">申请时间</th>
              <th class="col-approver">审批人</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="approval in paginatedApprovals" :key="approval.id" :class="[`row-${approval.status}`]">
              <td class="col-check" v-if="canApprove">
                <input
                  v-if="approval.status === 'pending'"
                  type="checkbox"
                  :checked="selectedIds.includes(approval.id)"
                  @change="toggleSelect(approval.id)"
                />
              </td>
              <td class="col-id">
                <span class="id-badge">#{{ approval.id }}</span>
              </td>
              <td class="col-status">
                <span class="status-badge" :class="approval.status">
                  <span class="status-dot"></span>
                  {{ statusLabel(approval.status) }}
                </span>
              </td>
              <td class="col-pipeline">
                <div class="pipeline-info">
                  <svg class="pipeline-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
                  <a class="pipeline-link" @click="goToPipeline(approval.pipeline_id)" :title="`查看流水线 #${approval.pipeline_id} 详情`">
                    {{ approval.pipeline_name || `Pipeline #${approval.pipeline_id}` }}
                  </a>
                  <span v-if="approval.pipeline_run_id" class="run-id-tag" :title="`运行记录 #${approval.pipeline_run_id}`">
                    Run #{{ approval.pipeline_run_id }}
                  </span>
                </div>
              </td>
              <td class="col-env">
                <span class="env-badge" :class="approval.env_name">
                  {{ envLabel(approval.env_name) }}
                </span>
              </td>
              <td class="col-image">
                <code class="image-tag" :title="approval.image">{{ truncateImage(approval.image) }}</code>
              </td>
              <td class="col-user">
                <div class="user-cell" v-if="approval.request_username">
                  <div class="avatar">{{ approval.request_username?.charAt(0)?.toUpperCase() }}</div>
                  <span>{{ approval.request_username }}</span>
                </div>
                <span v-else class="text-muted">-</span>
              </td>
              <td class="col-time">
                <span class="time-text">{{ formatTime(approval.created_at) }}</span>
              </td>
              <td class="col-approver">
                <div class="user-cell" v-if="approval.approve_username">
                  <div class="avatar approver">{{ approval.approve_username?.charAt(0)?.toUpperCase() }}</div>
                  <span>{{ approval.approve_username }}</span>
                </div>
                <span v-else class="text-muted">-</span>
              </td>
              <td class="col-actions">
                <template v-if="approval.status === 'pending' && canApprove">
                  <button class="action-btn approve" @click="handleApprove(approval)" :disabled="actionLoading" title="通过">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                  </button>
                  <button class="action-btn reject" @click="handleReject(approval)" :disabled="actionLoading" title="拒绝">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                  <button class="action-btn edit" @click="openEditModal(approval)" :disabled="actionLoading" title="编辑">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                </template>
                <span v-if="approval.status === 'pending' && !canApprove" class="text-muted">待审批</span>
                <button class="action-btn delete" @click="handleDelete(approval)" :disabled="actionLoading" title="删除" v-if="approval.status !== 'approved' && canApprove">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                </button>
                <span v-if="approval.status === 'approved' || (approval.status !== 'pending' && approval.status !== 'rejected' && approval.status !== 'expired')" class="reason-tip" :title="approval.approve_reason">
                  <svg v-if="approval.approve_reason" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页（现代化三段式布局） -->
      <div v-if="displayedApprovals.length > 0" class="pagination-wrapper">
        <div class="pagination-left">
          <span class="pagination-summary">共 <strong>{{ displayedApprovals.length }}</strong> 条</span>
        </div>
        <div class="pagination-center">
          <button class="pagination-btn" @click="goToPage(1)" :disabled="currentPage === 1" title="首页">«</button>
          <button class="pagination-btn" @click="goToPage(currentPage - 1)" :disabled="currentPage === 1" title="上一页">‹</button>
          <template v-for="page in visiblePages" :key="page">
            <button v-if="typeof page === 'number'" class="pagination-btn page-number" :class="{ active: currentPage === page }" @click="goToPage(page)">{{ page }}</button>
            <span v-else class="pagination-ellipsis">...</span>
          </template>
          <button class="pagination-btn" @click="goToPage(currentPage + 1)" :disabled="currentPage === totalPages" title="下一页">›</button>
          <button class="pagination-btn" @click="goToPage(totalPages)" :disabled="currentPage === totalPages" title="尾页">»</button>
        </div>
        <div class="pagination-right">
          <select v-model.number="pageSize" @change="onPageSizeChange" class="page-size-select">
            <option :value="10">10 条/页</option>
            <option :value="20">20 条/页</option>
            <option :value="50">50 条/页</option>
            <option :value="100">100 条/页</option>
          </select>
          <span class="pagination-goto">前往</span>
          <input v-model.number="jumpPage" type="number" min="1" :max="totalPages" class="page-jump-input" @keyup.enter="jumpToPage" />
        </div>
      </div>
    </div>

    <!-- 审批弹窗 - 现代风格 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showActionModal" class="modal-overlay" @click="closeModal">
          <div class="modal-dialog" @click.stop>
            <div class="modal-head" :class="actionType">
              <div class="modal-head-icon">
                <svg v-if="actionType === 'approve'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
              </div>
              <h3>{{ actionType === 'approve' ? '确认通过审批' : '确认拒绝审批' }}</h3>
              <button class="modal-close" @click="closeModal">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-main">
              <div class="summary-card">
                <div class="summary-row">
                  <span class="summary-label">流水线</span>
                  <span class="summary-value">{{ currentApproval?.pipeline_name || `Pipeline #${currentApproval?.pipeline_id}` }}</span>
                </div>
                <div class="summary-row">
                  <span class="summary-label">目标环境</span>
                  <span class="env-badge" :class="currentApproval?.env_name">{{ envLabel(currentApproval?.env_name) }}</span>
                </div>
                <div class="summary-row">
                  <span class="summary-label">部署镜像</span>
                  <code class="summary-image">{{ currentApproval?.image || '-' }}</code>
                </div>
              </div>
              <div class="form-field">
                <label class="field-label">审批意见 <span class="optional">(选填)</span></label>
                <textarea 
                  v-model="approvalReason" 
                  :placeholder="actionType === 'approve' ? '输入审批通过意见...' : '请说明拒绝原因...'"
                  rows="3"
                  class="field-textarea"
                ></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="closeModal">取消</button>
              <button 
                class="btn-confirm"
                :class="actionType"
                @click="submitAction"
                :disabled="actionLoading"
              >
                <svg v-if="actionLoading" class="btn-spinner" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" fill="none" stroke-dasharray="31.4 31.4" stroke-linecap="round"/></svg>
                {{ actionLoading ? '处理中...' : (actionType === 'approve' ? '确认通过' : '确认拒绝') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 新增/编辑弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showFormModal" class="modal-overlay" @click="closeFormModal">
          <div class="modal-dialog" @click.stop>
            <div class="modal-head create">
              <div class="modal-head-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              </div>
              <h3>{{ isEditing ? '编辑审批记录' : '新增审批申请' }}</h3>
              <button class="modal-close" @click="closeFormModal">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-main">
              <div class="form-field" v-if="!isEditing">
                <label class="field-label">流水线ID <span class="required">*</span></label>
                <input v-model.number="formData.pipeline_id" type="number" class="field-input" placeholder="输入流水线ID" />
              </div>
              <div class="form-field">
                <label class="field-label">目标环境</label>
                <select v-model="formData.env_name" class="field-input">
                  <option value="">请选择</option>
                  <option value="dev">开发环境</option>
                  <option value="staging">预发环境</option>
                  <option value="prod">生产环境</option>
                </select>
              </div>
              <div class="form-field">
                <label class="field-label">镜像地址 <span class="required" v-if="!isEditing">*</span></label>
                <input v-model="formData.image" type="text" class="field-input" placeholder="例如: registry.cn-hangzhou.aliyuncs.com/xxx:latest" />
              </div>
              <div class="form-field">
                <label class="field-label">申请原因</label>
                <textarea v-model="formData.request_reason" rows="3" class="field-textarea" placeholder="请输入申请原因..."></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="closeFormModal">取消</button>
              <button class="btn-confirm create" @click="submitForm" :disabled="formLoading">
                <svg v-if="formLoading" class="btn-spinner" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" fill="none" stroke-dasharray="31.4 31.4" stroke-linecap="round"/></svg>
                {{ formLoading ? '提交中...' : (isEditing ? '保存修改' : '提交申请') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getApprovalList, approvalAction, approvalBatchAction, getApprovalStats, createApproval, updateApproval, deleteApproval } from '@/api/cicd/environment.js'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

const router = useRouter()

const route = useRoute()
const approvals = ref([])
const loading = ref(false)
const actionLoading = ref(false)
const statusFilter = ref('')
const pipelineIdFilter = ref(null)
const searchQuery = ref('')
const canApprove = ref(false) // 是否有审批权限（由后端返回）

// 批量操作
const selectedIds = ref([])
const batchLoading = ref(false)

// 分页
const currentPage = ref(1)
const pageSize = ref(10)
const jumpPage = ref(1)

// 弹窗状态
const showActionModal = ref(false)
const currentApproval = ref(null)
const actionType = ref('approve')
const approvalReason = ref('')

// 新增/编辑弹窗状态
const showFormModal = ref(false)
const isEditing = ref(false)
const formLoading = ref(false)
const formData = ref({
  id: null,
  pipeline_id: null,
  env_name: '',
  image: '',
  request_reason: ''
})

// 统计数据（通过后端接口加载）
const pendingCount = ref(0)
const approvedCount = ref(0)
const rejectedCount = ref(0)
const totalCount = computed(() => pendingCount.value + approvedCount.value + rejectedCount.value)

// 过滤后的列表
const filteredApprovals = computed(() => {
  let list = approvals.value
  if (statusFilter.value) {
    list = list.filter(a => a.status === statusFilter.value)
  }
  if (pipelineIdFilter.value) {
    list = list.filter(a => a.pipeline_id === pipelineIdFilter.value)
  }
  return list
})

// 搜索过滤
const displayedApprovals = computed(() => {
  if (!searchQuery.value.trim()) return filteredApprovals.value
  const q = searchQuery.value.toLowerCase()
  return filteredApprovals.value.filter(a => 
    (a.pipeline_name || '').toLowerCase().includes(q) ||
    (a.image || '').toLowerCase().includes(q) ||
    (a.request_username || '').toLowerCase().includes(q) ||
    String(a.id).includes(q)
  )
})

// 分页计算属性
const totalPages = computed(() => Math.ceil(displayedApprovals.value.length / pageSize.value) || 1)

const paginatedApprovals = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return displayedApprovals.value.slice(start, start + pageSize.value)
})

const visiblePages = computed(() => {
  const tp = totalPages.value
  const current = currentPage.value
  const pages = []
  if (tp <= 7) {
    for (let i = 1; i <= tp; i++) pages.push(i)
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push('...')
      pages.push(tp)
    } else if (current >= tp - 3) {
      pages.push(1)
      pages.push('...')
      for (let i = tp - 4; i <= tp; i++) pages.push(i)
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) pages.push(i)
      pages.push('...')
      pages.push(tp)
    }
  }
  return pages
})

const goToPage = (page) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  jumpPage.value = page
}

const jumpToPage = () => {
  const page = parseInt(jumpPage.value)
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

const onPageSizeChange = () => {
  currentPage.value = 1
}

const setFilter = (status) => {
  statusFilter.value = statusFilter.value === status ? '' : status
  currentPage.value = 1
}

// 加载审批列表
const loadApprovals = async () => {
  loading.value = true
  try {
    const [listRes, statsRes] = await Promise.all([
      getApprovalList({ page: 1, page_size: 100, status: statusFilter.value || undefined, pipeline_id: pipelineIdFilter.value || undefined }),
      getApprovalStats()
    ])
    if (listRes.code === 0) {
      approvals.value = listRes.data?.list || []
      canApprove.value = listRes.data?.can_approve || false
    } else {
      console.error('加载审批列表失败:', listRes.msg)
    }
    if (statsRes.code === 0) {
      const stats = statsRes.data?.stats || {}
      pendingCount.value = stats.pending || 0
      approvedCount.value = stats.approved || 0
      rejectedCount.value = stats.rejected || 0
    }
  } catch (err) {
    console.error('加载审批列表异常:', err)
  } finally {
    loading.value = false
  }
}

const statusLabel = (status) => {
  const map = { pending: '待审批', approved: '已通过', rejected: '已拒绝', expired: '已过期' }
  return map[status] || status
}

const envLabel = (env) => {
  const map = { dev: '开发环境', staging: '预发环境', prod: '生产环境' }
  return map[env] || env || '-'
}

const truncateImage = (image) => {
  if (!image) return '-'
  if (image.length > 45) return image.slice(0, 22) + '...' + image.slice(-18)
  return image
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const d = new Date(timestamp * 1000)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const goToPipeline = (pipelineId) => {
  if (pipelineId) {
    router.push(`/cicd/pipelines/${pipelineId}`)
  }
}

const handleApprove = (approval) => {
  currentApproval.value = approval
  actionType.value = 'approve'
  approvalReason.value = ''
  showActionModal.value = true
}

const handleReject = (approval) => {
  currentApproval.value = approval
  actionType.value = 'reject'
  approvalReason.value = ''
  showActionModal.value = true
}

const closeModal = () => {
  showActionModal.value = false
  currentApproval.value = null
  approvalReason.value = ''
}

const submitAction = async () => {
  if (!currentApproval.value) return
  actionLoading.value = true
  try {
    const res = await approvalAction({
      id: currentApproval.value.id,
      action: actionType.value,
      reason: approvalReason.value
    })
    if (res.code === 0) {
      closeModal()
      await loadApprovals()
    } else {
      alert(res.msg || '操作失败')
    }
  } catch (err) {
    console.error('审批操作异常:', err)
    alert('操作失败，请重试')
  } finally {
    actionLoading.value = false
  }
}

// 新增审批
const openCreateModal = () => {
  isEditing.value = false
  formData.value = { id: null, pipeline_id: null, env_name: '', image: '', request_reason: '' }
  showFormModal.value = true
}

// 编辑审批
const openEditModal = (approval) => {
  isEditing.value = true
  formData.value = {
    id: approval.id,
    pipeline_id: approval.pipeline_id,
    env_name: approval.env_name || '',
    image: approval.image || '',
    request_reason: approval.request_reason || ''
  }
  showFormModal.value = true
}

const closeFormModal = () => {
  showFormModal.value = false
  formData.value = { id: null, pipeline_id: null, env_name: '', image: '', request_reason: '' }
}

const submitForm = async () => {
  if (!isEditing.value && !formData.value.pipeline_id) {
    alert('请输入流水线 ID')
    return
  }
  if (!formData.value.env_name) {
    alert('请选择目标环境')
    return
  }
  formLoading.value = true
  try {
    let res
    if (isEditing.value) {
      res = await updateApproval(formData.value)
    } else {
      res = await createApproval(formData.value)
    }
    if (res.code === 0) {
      closeFormModal()
      await loadApprovals()
    } else {
      alert(res.msg || '操作失败')
    }
  } catch (err) {
    console.error('提交审批异常:', err)
    alert('操作失败，请重试')
  } finally {
    formLoading.value = false
  }
}

// 删除审批
const handleDelete = async (approval) => {
  const ok = await showConfirm({
    title: '确认删除审批记录',
    content: '删除后不可恢复。',
    type: 'danger',
    details: [
      { label: '审批 ID', value: `#${approval.id}`, mono: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  actionLoading.value = true
  try {
    const res = await deleteApproval(approval.id)
    if (res.code === 0) {
      await loadApprovals()
    } else {
      alert(res.msg || '删除失败')
    }
  } catch (err) {
    console.error('删除审批异常:', err)
    alert('删除失败，请重试')
  } finally {
    actionLoading.value = false
  }
}

// ==================== 批量操作 ====================

// 可批量操作的待审批记录（当前页中状态为pending且不是自己提交的）
const selectablePending = computed(() => {
  return paginatedApprovals.value.filter(a => a.status === 'pending')
})

const isAllSelected = computed(() => {
  if (selectablePending.value.length === 0) return false
  return selectablePending.value.every(a => selectedIds.value.includes(a.id))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = selectablePending.value.map(a => a.id)
  }
}

const toggleSelect = (id) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx > -1) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

const handleBatchAction = async (action) => {
  if (selectedIds.value.length === 0) {
    alert('请先勾选要操作的审批记录')
    return
  }
  const actionLabel = action === 'approve' ? '批量通过' : '批量拒绝'
  const ok = await showConfirm({
    title: `确认${actionLabel}`,
    content: `将对选中的 ${selectedIds.value.length} 条审批记录执行${actionLabel}操作`,
    type: action === 'approve' ? 'success' : 'danger',
    confirmText: actionLabel,
    cancelText: '取消',
  })
  if (!ok) return

  batchLoading.value = true
  try {
    const res = await approvalBatchAction({
      ids: selectedIds.value,
      action: action,
      reason: `${actionLabel}操作`
    })
    if (res.code === 0) {
      const data = res.data || {}
      const msg = `${actionLabel}完成：成功 ${data.success || 0} 条` +
        (data.failures && data.failures.length > 0 ? `，失败 ${data.failures.length} 条` : '')
      alert(msg)
      selectedIds.value = []
      await loadApprovals()
    } else {
      alert(res.msg || '批量操作失败')
    }
  } catch (err) {
    console.error('批量审批异常:', err)
    alert('批量操作失败，请重试')
  } finally {
    batchLoading.value = false
  }
}

onMounted(() => {
  if (route.query.status) statusFilter.value = route.query.status
  if (route.query.pipeline_id) pipelineIdFilter.value = parseInt(route.query.pipeline_id)
  loadApprovals()
})

watch(() => route.query, (q) => {
  if (q.status) statusFilter.value = q.status
  if (q.pipeline_id) pipelineIdFilter.value = parseInt(q.pipeline_id)
}, { immediate: false })
</script>

<style scoped>
/* ============================================
   Rancher / Kuboard 企业级风格 - 审批管理
   ============================================ */

.approvals-page {
  min-height: 100vh;
  background: #f4f6f9;
}

/* ---- 顶部 Banner ---- */
.page-banner {
  background: linear-gradient(135deg, #1a2332 0%, #2d3e50 50%, #34495e 100%);
  padding: 28px 32px;
  position: relative;
  overflow: hidden;
}
.page-banner::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(78,124,246,0.12) 0%, transparent 70%);
  pointer-events: none;
}
.page-banner::after {
  content: '';
  position: absolute;
  bottom: -30%;
  left: 20%;
  width: 300px;
  height: 300px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(45,200,120,0.08) 0%, transparent 70%);
  pointer-events: none;
}
.banner-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
  z-index: 1;
  max-width: 100%;
  margin: 0 auto;
}
.banner-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.banner-icon {
  width: 48px;
  height: 48px;
  background: rgba(255,255,255,0.1);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255,255,255,0.08);
}
.banner-icon svg {
  width: 26px;
  height: 26px;
  color: #67d5b5;
}
.banner-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.5px;
}
.banner-desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: rgba(255,255,255,0.55);
}
.btn-refresh-banner {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 18px;
  background: rgba(255,255,255,0.1);
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 8px;
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.25s;
  backdrop-filter: blur(10px);
}
.btn-refresh-banner:hover {
  background: rgba(255,255,255,0.18);
  border-color: rgba(255,255,255,0.25);
}
.btn-refresh-banner:disabled { opacity: 0.6; cursor: not-allowed; }
.refresh-icon { width: 16px; height: 16px; }
.refresh-icon.spinning { animation: spin 1s linear infinite; }

/* ---- 指标卡片面板 ---- */
.metrics-panel {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 20px 32px 0;
  max-width: 100%;
  margin: -20px auto 0;
  position: relative;
  z-index: 2;
}
.metric-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04);
  cursor: pointer;
  transition: all 0.25s;
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
}
.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}
.metric-card.active {
  border-color: #4e7cf6;
  box-shadow: 0 2px 12px rgba(78,124,246,0.15);
}
.metric-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.metric-icon-wrapper svg { width: 22px; height: 22px; }
.metric-icon-wrapper.total   { background: #eef2ff; color: #4e7cf6; }
.metric-icon-wrapper.pending  { background: #fff8e1; color: #f59e0b; }
.metric-icon-wrapper.approved { background: #ecfdf5; color: #10b981; }
.metric-icon-wrapper.rejected { background: #fef2f2; color: #ef4444; }
.metric-body { display: flex; flex-direction: column; flex: 1; }
.metric-value {
  font-size: 26px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}
.metric-label {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.metric-trend {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1px;
  padding: 3px 8px;
  border-radius: 4px;
  align-self: flex-start;
}
.metric-trend.total    { background: #eef2ff; color: #4e7cf6; }
.metric-trend.pending  { background: #fff8e1; color: #f59e0b; }
.metric-trend.approved { background: #ecfdf5; color: #10b981; }
.metric-trend.rejected { background: #fef2f2; color: #ef4444; }

/* ---- 内容面板 ---- */
.content-panel {
  margin: 20px 32px 32px;
  max-width: 100%;
  margin-left: auto;
  margin-right: auto;
  padding: 0 32px;
}
.panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}
.record-count {
  font-size: 12px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 3px 10px;
  border-radius: 10px;
  font-weight: 500;
}
.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  transition: all 0.2s;
}
.search-box:focus-within {
  border-color: #4e7cf6;
  box-shadow: 0 0 0 3px rgba(78,124,246,0.1);
}
.search-icon { width: 16px; height: 16px; color: #94a3b8; flex-shrink: 0; }
.search-input {
  border: none;
  outline: none;
  font-size: 13px;
  color: #334155;
  width: 220px;
  background: transparent;
}
.search-input::placeholder { color: #cbd5e1; }

/* ---- 批量操作按钮 ---- */
.btn-batch {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid;
  transition: all 0.2s;
}
.btn-batch svg { width: 14px; height: 14px; }
.btn-batch.approve {
  background: #ecfdf5;
  color: #059669;
  border-color: #a7f3d0;
}
.btn-batch.approve:hover {
  background: #d1fae5;
  border-color: #6ee7b7;
}
.btn-batch.reject {
  background: #fef2f2;
  color: #dc2626;
  border-color: #fecaca;
}
.btn-batch.reject:hover {
  background: #fee2e2;
  border-color: #fca5a5;
}
.btn-batch:disabled { opacity: 0.6; cursor: not-allowed; }

/* ---- 复选框列 ---- */
.col-check {
  width: 40px;
  text-align: center;
}
.col-check input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #4e7cf6;
}

/* ---- 表格 ---- */
.approval-table-wrapper {
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04);
  overflow: hidden;
}
.approval-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.approval-table thead {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}
.approval-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #64748b;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  white-space: nowrap;
}
.approval-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #334155;
  vertical-align: middle;
}
.approval-table tbody tr {
  transition: background 0.15s;
}
.approval-table tbody tr:hover {
  background: #f8fafc;
}
.approval-table tbody tr:last-child td {
  border-bottom: none;
}
.approval-table tbody tr.row-pending {
  border-left: 3px solid #f59e0b;
}
.approval-table tbody tr.row-approved {
  border-left: 3px solid #10b981;
}
.approval-table tbody tr.row-rejected {
  border-left: 3px solid #ef4444;
}
.approval-table tbody tr.row-expired {
  border-left: 3px solid #cbd5e1;
  opacity: 0.65;
}

/* 表格单元格 */
.id-badge {
  color: #64748b;
  font-weight: 600;
  font-size: 12px;
}
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-badge.pending  { background: #fffbeb; color: #d97706; }
.status-badge.pending .status-dot  { background: #f59e0b; box-shadow: 0 0 6px rgba(245,158,11,0.4); }
.status-badge.approved { background: #ecfdf5; color: #059669; }
.status-badge.approved .status-dot { background: #10b981; }
.status-badge.rejected { background: #fef2f2; color: #dc2626; }
.status-badge.rejected .status-dot { background: #ef4444; }
.status-badge.expired  { background: #f8fafc; color: #94a3b8; }
.status-badge.expired .status-dot  { background: #cbd5e1; }

.pipeline-info {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pipeline-icon { width: 16px; height: 16px; color: #94a3b8; flex-shrink: 0; }
.pipeline-link {
  font-weight: 500;
  color: #4e7cf6;
  cursor: pointer;
  transition: color 0.2s;
}
.pipeline-link:hover {
  color: #3b65d9;
  text-decoration: underline;
}
.run-id-tag {
  font-size: 10px;
  background: #f1f5f9;
  color: #64748b;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 600;
  font-family: 'SF Mono', 'Fira Code', monospace;
  white-space: nowrap;
}

.env-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 5px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.3px;
}
.env-badge.dev     { background: #eff6ff; color: #2563eb; }
.env-badge.staging { background: #fffbeb; color: #d97706; }
.env-badge.prod    { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }

.image-tag {
  font-size: 11px;
  background: #f1f5f9;
  padding: 3px 8px;
  border-radius: 4px;
  color: #475569;
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
  word-break: break-all;
}
.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.avatar {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea, #4e7cf6);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.avatar.approver {
  background: linear-gradient(135deg, #10b981, #059669);
}
.time-text {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}
.text-muted { color: #cbd5e1; }
.reason-tip {
  display: inline-flex;
  cursor: help;
}
.reason-tip svg { width: 18px; height: 18px; color: #94a3b8; }

/* 操作按钮 */
.action-btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  margin-right: 6px;
}
.action-btn svg { width: 16px; height: 16px; }
.action-btn.approve {
  background: #ecfdf5;
  color: #10b981;
}
.action-btn.approve:hover {
  background: #10b981;
  color: #fff;
  box-shadow: 0 2px 8px rgba(16,185,129,0.3);
}
.action-btn.reject {
  background: #fef2f2;
  color: #ef4444;
}
.action-btn.reject:hover {
  background: #ef4444;
  color: #fff;
  box-shadow: 0 2px 8px rgba(239,68,68,0.3);
}
.action-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ---- 加载 & 空状态 ---- */
.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}
.loader {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}
.loader-ring {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #4e7cf6;
  animation: bounce 1.4s ease-in-out infinite both;
}
.loader-ring:nth-child(1) { animation-delay: -0.32s; }
.loader-ring:nth-child(2) { animation-delay: -0.16s; }
@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}
.loading-text { color: #94a3b8; font-size: 14px; }

.empty-illustration svg { width: 160px; height: 130px; }
.empty-title {
  margin: 16px 0 6px;
  font-size: 16px;
  font-weight: 600;
  color: #475569;
}
.empty-desc {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}

/* ---- Modal ---- */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15,23,42,0.55);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.modal-dialog {
  background: #fff;
  border-radius: 14px;
  width: 520px;
  max-width: 92%;
  box-shadow: 0 25px 60px rgba(0,0,0,0.2);
  overflow: hidden;
}
.modal-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px;
  position: relative;
}
.modal-head.approve { background: linear-gradient(135deg, #ecfdf5, #d1fae5); }
.modal-head.reject  { background: linear-gradient(135deg, #fef2f2, #fecaca); }
.modal-head-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal-head.approve .modal-head-icon { background: #10b981; color: #fff; }
.modal-head.reject .modal-head-icon  { background: #ef4444; color: #fff; }
.modal-head-icon svg { width: 22px; height: 22px; }
.modal-head h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
}
.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  color: #94a3b8;
  transition: all 0.2s;
}
.modal-close:hover { background: rgba(0,0,0,0.06); color: #475569; }
.modal-close svg { width: 20px; height: 20px; }

.modal-main { padding: 20px 24px; }
.summary-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 18px;
}
.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}
.summary-row + .summary-row { border-top: 1px solid #f1f5f9; }
.summary-label { font-size: 13px; color: #64748b; font-weight: 500; }
.summary-value { font-size: 13px; color: #1e293b; font-weight: 600; }
.summary-image {
  font-size: 11px;
  background: #e2e8f0;
  padding: 2px 8px;
  border-radius: 4px;
  color: #475569;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  font-family: 'SF Mono', monospace;
}
.field-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
  margin-bottom: 8px;
}
.optional { color: #94a3b8; font-weight: 400; }
.field-textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  color: #334155;
  resize: vertical;
  transition: all 0.2s;
  box-sizing: border-box;
  font-family: inherit;
}
.field-textarea:focus {
  outline: none;
  border-color: #4e7cf6;
  box-shadow: 0 0 0 3px rgba(78,124,246,0.1);
}

.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  background: #f8fafc;
  border-top: 1px solid #f1f5f9;
}
.btn-cancel {
  padding: 9px 20px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover { background: #f1f5f9; color: #334155; }
.btn-confirm {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 24px;
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-confirm.approve { background: #10b981; }
.btn-confirm.approve:hover { background: #059669; box-shadow: 0 2px 10px rgba(16,185,129,0.3); }
.btn-confirm.reject  { background: #ef4444; }
.btn-confirm.reject:hover  { background: #dc2626; box-shadow: 0 2px 10px rgba(239,68,68,0.3); }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

/* ---- Transition ---- */
.modal-enter-active { transition: all 0.3s ease; }
.modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal-dialog { transform: scale(0.95) translateY(10px); }
.modal-leave-to .modal-dialog { transform: scale(0.97); }

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* ---- 新增按钮 & 操作按钮扩展 ---- */
.btn-create {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #4e7cf6;
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-create:hover {
  background: #3b65d9;
  box-shadow: 0 2px 10px rgba(78,124,246,0.3);
}
.btn-create svg { width: 16px; height: 16px; }
.action-btn.edit {
  background: #eff6ff;
  color: #3b82f6;
}
.action-btn.edit:hover {
  background: #3b82f6;
  color: #fff;
  box-shadow: 0 2px 8px rgba(59,130,246,0.3);
}
.action-btn.delete {
  background: #fef2f2;
  color: #ef4444;
}
.action-btn.delete:hover {
  background: #ef4444;
  color: #fff;
  box-shadow: 0 2px 8px rgba(239,68,68,0.3);
}

/* ---- 新增/编辑弹窗样式 ---- */
.modal-head.create { background: linear-gradient(135deg, #eef2ff, #dbeafe); }
.modal-head.create .modal-head-icon { background: #4e7cf6; color: #fff; }
.field-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  color: #334155;
  transition: all 0.2s;
  box-sizing: border-box;
  font-family: inherit;
  background: #fff;
}
.field-input:focus {
  outline: none;
  border-color: #4e7cf6;
  box-shadow: 0 0 0 3px rgba(78,124,246,0.1);
}
.form-field { margin-bottom: 16px; }
.required { color: #ef4444; font-weight: 400; }
.btn-confirm.create { background: #4e7cf6; }
.btn-confirm.create:hover { background: #3b65d9; box-shadow: 0 2px 10px rgba(78,124,246,0.3); }

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .metrics-panel { grid-template-columns: repeat(2, 1fr); }
  .page-banner { padding: 20px 20px; }
  .content-panel { padding: 0 20px; margin: 16px 20px 24px; }
}
@media (max-width: 640px) {
  .metrics-panel { grid-template-columns: 1fr; }
  .panel-toolbar { flex-direction: column; gap: 12px; align-items: stretch; }
  .approval-table-wrapper { overflow-x: auto; }
}

/* ---- Pagination (Modern) ---- */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 14px 20px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  flex-wrap: wrap;
  gap: 14px;
}
.pagination-left { display: flex; align-items: center; }
.pagination-summary { font-size: 13px; color: #64748b; }
.pagination-summary strong { color: #1e293b; font-weight: 600; }
.pagination-center { display: flex; align-items: center; gap: 4px; }
.pagination-btn {
  min-width: 34px; height: 34px; border: 1px solid #e2e8f0; border-radius: 6px;
  background: #fff; color: #475569; font-size: 14px; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; transition: all 0.2s;
}
.pagination-btn:hover:not(:disabled) { border-color: #4e7cf6; color: #4e7cf6; background: #f0f5ff; }
.pagination-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pagination-btn.page-number.active { background: #4e7cf6; color: #fff; border-color: #4e7cf6; font-weight: 600; }
.pagination-ellipsis { color: #94a3b8; font-size: 14px; padding: 0 4px; }
.pagination-right { display: flex; align-items: center; gap: 8px; }
.page-size-select {
  padding: 6px 10px; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 12px; color: #475569; background: #fff; cursor: pointer;
}
.page-size-select:focus { outline: none; border-color: #4e7cf6; }
.pagination-goto { font-size: 12px; color: #64748b; }
.page-jump-input {
  width: 50px; padding: 5px 8px; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 12px; text-align: center; color: #475569;
}
.page-jump-input:focus { outline: none; border-color: #4e7cf6; }
</style>

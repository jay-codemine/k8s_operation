<template>
  <div class="approvals-page">
    <div class="page-header">
      <h2>部署审批管理</h2>
      <p class="page-desc">管理生产环境部署审批申请</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards">
      <div class="stat-card pending">
        <div class="stat-icon">⏳</div>
        <div class="stat-info">
          <span class="stat-value">{{ pendingCount }}</span>
          <span class="stat-label">待审批</span>
        </div>
      </div>
      <div class="stat-card approved">
        <div class="stat-icon">✓</div>
        <div class="stat-info">
          <span class="stat-value">{{ approvedCount }}</span>
          <span class="stat-label">已通过</span>
        </div>
      </div>
      <div class="stat-card rejected">
        <div class="stat-icon">✗</div>
        <div class="stat-info">
          <span class="stat-value">{{ rejectedCount }}</span>
          <span class="stat-label">已拒绝</span>
        </div>
      </div>
    </div>

    <!-- 筛选工具栏 -->
    <div class="toolbar">
      <div class="filter-group">
        <select v-model="statusFilter" class="filter-select">
          <option value="">全部状态</option>
          <option value="pending">待审批</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
          <option value="expired">已过期</option>
        </select>
      </div>
      <button class="btn btn-refresh" @click="loadApprovals">
        🔄 刷新
      </button>
    </div>

    <!-- 审批列表 -->
    <div class="approval-list">
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="filteredApprovals.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p>暂无审批记录</p>
      </div>

      <div v-else class="approval-cards">
        <div 
          v-for="approval in filteredApprovals" 
          :key="approval.id" 
          class="approval-card"
          :class="[`status-${approval.status}`]"
        >
          <div class="card-header">
            <div class="approval-id">#{{ approval.id }}</div>
            <div class="approval-status" :class="approval.status">
              {{ statusLabel(approval.status) }}
            </div>
          </div>

          <div class="card-body">
            <div class="info-row">
              <span class="label">流水线:</span>
              <span class="value">{{ approval.pipeline_name || `Pipeline #${approval.pipeline_id}` }}</span>
            </div>
            <div class="info-row">
              <span class="label">目标环境:</span>
              <span class="value env-tag" :class="approval.env_name">
                {{ envLabel(approval.env_name) }}
              </span>
            </div>
            <div class="info-row">
              <span class="label">部署镜像:</span>
              <span class="value image-value" :title="approval.image">
                {{ truncateImage(approval.image) }}
              </span>
            </div>
            <div class="info-row" v-if="approval.request_reason">
              <span class="label">申请原因:</span>
              <span class="value">{{ approval.request_reason }}</span>
            </div>
            <div class="info-row">
              <span class="label">申请时间:</span>
              <span class="value">{{ formatTime(approval.created_at) }}</span>
            </div>
            <div class="info-row" v-if="approval.approve_time">
              <span class="label">审批时间:</span>
              <span class="value">{{ formatTime(approval.approve_time) }}</span>
            </div>
            <div class="info-row" v-if="approval.approve_reason">
              <span class="label">审批意见:</span>
              <span class="value">{{ approval.approve_reason }}</span>
            </div>
          </div>

          <div class="card-footer" v-if="approval.status === 'pending'">
            <button 
              class="btn btn-approve" 
              @click="handleApprove(approval)"
              :disabled="actionLoading"
            >
              ✓ 通过
            </button>
            <button 
              class="btn btn-reject" 
              @click="handleReject(approval)"
              :disabled="actionLoading"
            >
              ✗ 拒绝
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 审批弹窗 -->
    <div v-if="showActionModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ actionType === 'approve' ? '确认通过' : '确认拒绝' }}</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="approval-summary">
            <p><strong>流水线:</strong> {{ currentApproval?.pipeline_name || `Pipeline #${currentApproval?.pipeline_id}` }}</p>
            <p><strong>目标环境:</strong> {{ envLabel(currentApproval?.env_name) }}</p>
            <p><strong>部署镜像:</strong> {{ currentApproval?.image }}</p>
          </div>
          <div class="form-group">
            <label>审批意见（可选）:</label>
            <textarea 
              v-model="approvalReason" 
              placeholder="请输入审批意见..."
              rows="3"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button 
            class="btn" 
            :class="actionType === 'approve' ? 'btn-success' : 'btn-danger'"
            @click="submitAction"
            :disabled="actionLoading"
          >
            {{ actionLoading ? '处理中...' : (actionType === 'approve' ? '确认通过' : '确认拒绝') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getApprovalList, approvalAction } from '@/api/cicd/environment.js'

const approvals = ref([])
const loading = ref(false)
const actionLoading = ref(false)
const statusFilter = ref('')

// 弹窗状态
const showActionModal = ref(false)
const currentApproval = ref(null)
const actionType = ref('approve')
const approvalReason = ref('')

// 统计数据
const pendingCount = computed(() => approvals.value.filter(a => a.status === 'pending').length)
const approvedCount = computed(() => approvals.value.filter(a => a.status === 'approved').length)
const rejectedCount = computed(() => approvals.value.filter(a => a.status === 'rejected').length)

// 过滤后的列表
const filteredApprovals = computed(() => {
  if (!statusFilter.value) return approvals.value
  return approvals.value.filter(a => a.status === statusFilter.value)
})

// 加载审批列表
const loadApprovals = async () => {
  loading.value = true
  try {
    const res = await getApprovalList({ page: 1, page_size: 100 })
    if (res.code === 0) {
      approvals.value = res.data?.list || []
    } else {
      console.error('加载审批列表失败:', res.msg)
    }
  } catch (err) {
    console.error('加载审批列表异常:', err)
  } finally {
    loading.value = false
  }
}

// 状态标签
const statusLabel = (status) => {
  const map = {
    pending: '待审批',
    approved: '已通过',
    rejected: '已拒绝',
    expired: '已过期'
  }
  return map[status] || status
}

// 环境标签
const envLabel = (env) => {
  const map = {
    dev: '开发环境',
    staging: '预发环境',
    prod: '生产环境'
  }
  return map[env] || env
}

// 截断镜像地址
const truncateImage = (image) => {
  if (!image) return '-'
  if (image.length > 50) {
    return image.slice(0, 25) + '...' + image.slice(-20)
  }
  return image
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

// 处理通过
const handleApprove = (approval) => {
  currentApproval.value = approval
  actionType.value = 'approve'
  approvalReason.value = ''
  showActionModal.value = true
}

// 处理拒绝
const handleReject = (approval) => {
  currentApproval.value = approval
  actionType.value = 'reject'
  approvalReason.value = ''
  showActionModal.value = true
}

// 关闭弹窗
const closeModal = () => {
  showActionModal.value = false
  currentApproval.value = null
  approvalReason.value = ''
}

// 提交审批操作
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

onMounted(() => {
  loadApprovals()
})
</script>

<style scoped>
.approvals-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #1a1a1a;
}

.page-desc {
  margin: 0;
  color: #666;
  font-size: 14px;
}

/* 统计卡片 */
.stats-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.stat-card.pending { border-left: 4px solid #faad14; }
.stat-card.approved { border-left: 4px solid #52c41a; }
.stat-card.rejected { border-left: 4px solid #f5222d; }

.stat-icon {
  font-size: 32px;
  margin-right: 16px;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  min-width: 150px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-refresh {
  background: #f5f5f5;
  color: #333;
}

.btn-refresh:hover {
  background: #e8e8e8;
}

/* 审批卡片列表 */
.approval-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
}

.approval-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.approval-card.status-pending {
  border-top: 3px solid #faad14;
}

.approval-card.status-approved {
  border-top: 3px solid #52c41a;
}

.approval-card.status-rejected {
  border-top: 3px solid #f5222d;
}

.approval-card.status-expired {
  border-top: 3px solid #d9d9d9;
  opacity: 0.7;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}

.approval-id {
  font-weight: 600;
  color: #666;
}

.approval-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.approval-status.pending {
  background: #fff7e6;
  color: #d46b08;
}

.approval-status.approved {
  background: #f6ffed;
  color: #389e0d;
}

.approval-status.rejected {
  background: #fff1f0;
  color: #cf1322;
}

.approval-status.expired {
  background: #f5f5f5;
  color: #8c8c8c;
}

.card-body {
  padding: 16px;
}

.info-row {
  display: flex;
  margin-bottom: 10px;
  font-size: 14px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-row .label {
  flex-shrink: 0;
  width: 80px;
  color: #8c8c8c;
}

.info-row .value {
  color: #1a1a1a;
  word-break: break-all;
}

.env-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.env-tag.dev {
  background: #e6f7ff;
  color: #1890ff;
}

.env-tag.staging {
  background: #fff7e6;
  color: #fa8c16;
}

.env-tag.prod {
  background: #fff1f0;
  color: #f5222d;
}

.image-value {
  font-family: monospace;
  font-size: 13px;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
}

.card-footer {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
}

.btn-approve {
  flex: 1;
  background: #52c41a;
  color: #fff;
}

.btn-approve:hover {
  background: #389e0d;
}

.btn-reject {
  flex: 1;
  background: #ff4d4f;
  color: #fff;
}

.btn-reject:hover {
  background: #cf1322;
}

/* 加载和空状态 */
.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  background: #fff;
  border-radius: 8px;
  color: #8c8c8c;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #1890ff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

/* 弹窗 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #fff;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #8c8c8c;
  cursor: pointer;
}

.modal-body {
  padding: 20px;
}

.approval-summary {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.approval-summary p {
  margin: 0 0 8px 0;
  font-size: 14px;
}

.approval-summary p:last-child {
  margin-bottom: 0;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: #333;
}

.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  resize: vertical;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #f0f0f0;
}

.btn-secondary {
  background: #f5f5f5;
  color: #333;
}

.btn-success {
  background: #52c41a;
  color: #fff;
}

.btn-danger {
  background: #ff4d4f;
  color: #fff;
}
</style>

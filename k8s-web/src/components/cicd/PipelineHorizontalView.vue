<template>
  <!-- 水平流水线阶段视图（阿里云/腾讯云风格） -->
  <div class="pipeline-horizontal-view">
    <!-- 统计栏 -->
    <div class="stats-tabs">
      <div 
        v-for="tab in statsTabs" 
        :key="tab.key"
        :class="['tab-item', { active: activeTab === tab.key }]"
        @click="activeTab = tab.key"
      >
        <span :class="['tab-dot', tab.key]"></span>
        {{ tab.label }}
        <span class="tab-count">{{ tab.count }}</span>
      </div>
    </div>

    <!-- 流水线轨道 -->
    <div class="pipeline-track-wrapper">
      <div class="pipeline-track">
        <div 
          v-for="(stage, index) in filteredStages" 
          :key="stage.id || index"
          class="stage-node-group"
        >
          <!-- 连接线（左） -->
          <div 
            v-if="index > 0" 
            :class="['connector-line', `status-${getLineStatus(index - 1)}`]"
          >
            <div class="line-fill" :style="{ width: getLineFillWidth(index - 1) }"></div>
          </div>

          <!-- 阶段卡片 -->
          <div 
            :class="['stage-node', `status-${stage.status}`, { selected: selectedStage?.name === stage.name }]"
            @click="selectStage(stage)"
          >
            <!-- 状态图标 -->
            <div :class="['status-circle', `status-${stage.status}`]">
              <svg v-if="stage.status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              <svg v-else-if="stage.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
              <div v-else-if="stage.status === 'running' || stage.status === 'deploying'" class="spinner"></div>
              <svg v-else-if="stage.status === 'waiting'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
              <div v-else class="pending-inner"></div>
            </div>

            <!-- 阶段名称 -->
            <div class="stage-title">{{ stage.name }}</div>

            <!-- 耗时/状态 -->
            <div :class="['stage-meta', `status-${stage.status}`]">
              <template v-if="stage.status === 'running' || stage.status === 'deploying'">
                <span class="running-indicator"></span>
                {{ formatElapsed(stage.started_at) || '执行中' }}
              </template>
              <template v-else-if="stage.duration && stage.duration !== '-'">
                {{ stage.duration }}
              </template>
              <template v-else>
                -
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 当前运行阶段提示 -->
    <div v-if="runningStage" class="running-hint">
      <div class="hint-icon">
        <div class="spinner-small"></div>
      </div>
      <span>正在执行: <strong>{{ runningStage.name }}</strong></span>
      <span class="hint-time">{{ formatElapsed(runningStage.started_at) }}</span>
    </div>

    <!-- 选中阶段详情 -->
    <transition name="slide-up">
      <div v-if="selectedStage" class="selected-detail">
        <div class="detail-header">
          <div class="detail-title">
            <span :class="['status-badge', `status-${selectedStage.status}`]">
              {{ getStatusText(selectedStage.status) }}
            </span>
            <span class="name">{{ selectedStage.name }}</span>
            <span v-if="selectedStage.type" class="type-tag">{{ getTypeText(selectedStage.type) }}</span>
          </div>
          <button class="close-btn" @click="selectedStage = null">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div class="detail-body">
          <!-- 步骤列表 -->
          <div v-if="selectedStage.steps?.length" class="steps-section">
            <div class="section-title">执行步骤</div>
            <div class="steps-grid">
              <div 
                v-for="step in selectedStage.steps" 
                :key="step.id" 
                :class="['step-row', `status-${step.status}`]"
              >
                <span :class="['step-dot', `status-${step.status}`]"></span>
                <span class="step-name">{{ step.name }}</span>
                <span class="step-duration">{{ step.duration || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- 错误信息 -->
          <div v-if="selectedStage.error_message" class="error-section">
            <div class="section-title error">错误信息</div>
            <div class="error-content">{{ selectedStage.error_message }}</div>
          </div>

          <!-- 操作按钮 -->
          <div class="actions-section">
            <button 
              v-if="selectedStage.type === 'approval' && selectedStage.status === 'waiting'"
              class="action-btn approve"
              @click="$emit('approve', selectedStage.id, 'approve')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              通过审批
            </button>
            <button 
              v-if="selectedStage.type === 'approval' && selectedStage.status === 'waiting'"
              class="action-btn reject"
              @click="$emit('approve', selectedStage.id, 'reject')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
              拒绝
            </button>
            <button 
              v-if="selectedStage.type === 'deploy' && selectedStage.status === 'pending'"
              class="action-btn deploy"
              @click="$emit('deploy', selectedStage.id)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
              </svg>
              执行部署
            </button>
            <button 
              v-if="selectedStage.type === 'deploy' && selectedStage.status === 'failed'"
              class="action-btn retry"
              @click="$emit('retry-deploy', selectedStage.id)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              重新部署
            </button>
            <button class="action-btn logs" @click="$emit('view-logs', selectedStage)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              查看日志
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  stages: { type: Array, default: () => [] }
})

defineEmits(['approve', 'deploy', 'retry-deploy', 'view-logs'])

const activeTab = ref('all')
const selectedStage = ref(null)

// 统计 tabs
const statsTabs = computed(() => {
  const counts = { all: 0, success: 0, failed: 0, running: 0, pending: 0 }
  props.stages.forEach(s => {
    counts.all++
    if (s.status === 'success') counts.success++
    else if (s.status === 'failed') counts.failed++
    else if (s.status === 'running' || s.status === 'deploying') counts.running++
    else counts.pending++
  })
  return [
    { key: 'all', label: '全部', count: counts.all },
    { key: 'success', label: '成功', count: counts.success },
    { key: 'failed', label: '失败', count: counts.failed },
    { key: 'running', label: '运行中', count: counts.running },
    { key: 'pending', label: '待执行', count: counts.pending }
  ]
})

// 过滤阶段
const filteredStages = computed(() => {
  if (activeTab.value === 'all') return props.stages
  if (activeTab.value === 'running') return props.stages.filter(s => s.status === 'running' || s.status === 'deploying')
  if (activeTab.value === 'pending') return props.stages.filter(s => s.status === 'pending' || s.status === 'waiting')
  return props.stages.filter(s => s.status === activeTab.value)
})

// 运行中的阶段
const runningStage = computed(() => props.stages.find(s => s.status === 'running' || s.status === 'deploying'))

// 方法
const selectStage = (stage) => {
  selectedStage.value = selectedStage.value?.name === stage.name ? null : stage
}

const getLineStatus = (index) => {
  const stage = props.stages[index]
  if (stage.status === 'success') return 'success'
  if (stage.status === 'failed') return 'failed'
  if (stage.status === 'running') return 'running'
  return 'pending'
}

const getLineFillWidth = (index) => {
  const stage = props.stages[index]
  if (stage.status === 'success') return '100%'
  if (stage.status === 'failed') return '100%'
  if (stage.status === 'running') return '50%'
  return '0%'
}

const getStatusText = (status) => {
  const map = { success: '成功', failed: '失败', running: '运行中', deploying: '部署中', waiting: '等待', pending: '待执行' }
  return map[status] || status
}

const getTypeText = (type) => {
  const map = { checkout: '检出', build: '构建', test: '测试', push: '推送', approval: '审批', deploy: '部署' }
  return map[type] || type
}

const formatElapsed = (startedAt) => {
  if (!startedAt) return ''
  const start = typeof startedAt === 'number' ? startedAt * 1000 : new Date(startedAt).getTime()
  const secs = Math.floor((Date.now() - start) / 1000)
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.floor(secs / 60)}m${secs % 60}s`
  return `${Math.floor(secs / 3600)}h${Math.floor((secs % 3600) / 60)}m`
}

// 自动选中运行中阶段
watch(runningStage, (stage) => {
  if (stage && !selectedStage.value) selectedStage.value = stage
}, { immediate: true })
</script>

<style scoped>
.pipeline-horizontal-view {
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e5e7eb;
  overflow: hidden;
}

/* 统计 tabs */
.stats-tabs {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
  background: linear-gradient(180deg, #f9fafb 0%, #fff 100%);
  border-bottom: 1px solid #f3f4f6;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.tab-item:hover {
  background: #f3f4f6;
}

.tab-item.active {
  background: #fff;
  border-color: #e5e7eb;
  color: #111827;
  font-weight: 500;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.tab-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.tab-dot.all { background: #6b7280; }
.tab-dot.success { background: #10b981; }
.tab-dot.failed { background: #ef4444; }
.tab-dot.running { background: #3b82f6; }
.tab-dot.pending { background: #9ca3af; }

.tab-count {
  font-weight: 600;
  margin-left: 2px;
}

/* 流水线轨道 */
.pipeline-track-wrapper {
  padding: 32px 24px;
  overflow-x: auto;
}

.pipeline-track {
  display: flex;
  align-items: center;
  min-width: min-content;
}

.stage-node-group {
  display: flex;
  align-items: center;
}

/* 连接线 */
.connector-line {
  width: 60px;
  height: 4px;
  background: #e5e7eb;
  position: relative;
  overflow: hidden;
}

.connector-line .line-fill {
  height: 100%;
  background: #10b981;
  transition: width 0.5s ease;
}

.connector-line.status-failed .line-fill {
  background: #ef4444;
}

.connector-line.status-running .line-fill {
  background: linear-gradient(90deg, #10b981, #3b82f6);
  animation: flow 1.5s linear infinite;
}

@keyframes flow {
  0% { background-position: -100% 0; }
  100% { background-position: 100% 0; }
}

/* 阶段节点 */
.stage-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 24px;
  border-radius: 12px;
  border: 2px solid #e5e7eb;
  background: #fff;
  min-width: 120px;
  cursor: pointer;
  transition: all 0.3s;
}

.stage-node:hover {
  border-color: #d1d5db;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.stage-node.selected {
  border-color: #3b82f6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
}

.stage-node.status-success {
  background: linear-gradient(180deg, #f0fdf4 0%, #fff 100%);
  border-color: #86efac;
}

.stage-node.status-failed {
  background: linear-gradient(180deg, #fef2f2 0%, #fff 100%);
  border-color: #fca5a5;
}

.stage-node.status-running, .stage-node.status-deploying {
  background: linear-gradient(180deg, #eff6ff 0%, #fff 100%);
  border-color: #93c5fd;
}

/* 状态圆圈 */
.status-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  background: #f3f4f6;
  border: 2px solid #e5e7eb;
}

.status-circle.status-success {
  background: #10b981;
  border-color: #10b981;
  color: #fff;
}

.status-circle.status-failed {
  background: #ef4444;
  border-color: #ef4444;
  color: #fff;
}

.status-circle.status-running, .status-circle.status-deploying {
  background: #3b82f6;
  border-color: #3b82f6;
}

.status-circle.status-waiting {
  background: #fbbf24;
  border-color: #fbbf24;
  color: #fff;
}

.status-circle svg {
  width: 24px;
  height: 24px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 3px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pending-inner {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #9ca3af;
}

/* 阶段标题 */
.stage-title {
  font-weight: 600;
  color: #111827;
  margin-bottom: 6px;
  text-align: center;
}

/* 阶段 meta */
.stage-meta {
  font-size: 13px;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 6px;
}

.stage-meta.status-running, .stage-meta.status-deploying {
  color: #3b82f6;
}

.running-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3b82f6;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

/* 运行提示 */
.running-hint {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: #eff6ff;
  border-top: 1px solid #dbeafe;
}

.hint-icon {
  display: flex;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid #93c5fd;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.running-hint span {
  font-size: 13px;
  color: #1e40af;
}

.hint-time {
  margin-left: auto;
  font-weight: 500;
}

/* 选中详情 */
.selected-detail {
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #fff;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.status-success { background: #d1fae5; color: #059669; }
.status-badge.status-failed { background: #fee2e2; color: #dc2626; }
.status-badge.status-running, .status-badge.status-deploying { background: #dbeafe; color: #2563eb; }
.status-badge.status-waiting { background: #fef3c7; color: #d97706; }
.status-badge.status-pending { background: #f3f4f6; color: #6b7280; }

.detail-title .name {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.type-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  background: #e5e7eb;
  color: #6b7280;
}

.close-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #6b7280;
}

.close-btn svg {
  width: 18px;
  height: 18px;
}

.detail-body {
  padding: 20px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 12px;
}

.section-title.error {
  color: #dc2626;
}

/* 步骤列表 */
.steps-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 20px;
}

.step-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.step-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #9ca3af;
}

.step-dot.status-success { background: #10b981; }
.step-dot.status-failed { background: #ef4444; }
.step-dot.status-running { background: #3b82f6; }

.step-name {
  flex: 1;
  font-size: 13px;
  color: #374151;
}

.step-duration {
  font-size: 12px;
  color: #9ca3af;
}

/* 错误区域 */
.error-section {
  margin-bottom: 20px;
}

.error-content {
  padding: 14px;
  background: #fef2f2;
  border-radius: 8px;
  font-size: 13px;
  color: #991b1b;
  line-height: 1.6;
  border: 1px solid #fecaca;
}

/* 操作按钮 */
.actions-section {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.action-btn svg {
  width: 16px;
  height: 16px;
}

.action-btn.approve {
  background: #10b981;
  color: #fff;
}

.action-btn.approve:hover {
  background: #059669;
}

.action-btn.reject {
  background: #fee2e2;
  color: #dc2626;
}

.action-btn.reject:hover {
  background: #fecaca;
}

.action-btn.deploy {
  background: #3b82f6;
  color: #fff;
}

.action-btn.deploy:hover {
  background: #2563eb;
}

.action-btn.retry {
  background: #f59e0b;
  color: #fff;
}

.action-btn.retry:hover {
  background: #d97706;
}

.action-btn.logs {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #e5e7eb;
}

.action-btn.logs:hover {
  background: #e5e7eb;
}

/* 动画 */
.slide-up-enter-active, .slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from, .slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>

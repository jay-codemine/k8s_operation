<template>
  <!-- 流水线阶段视图（参考 GitHub Actions + 阿里云 DevOps 风格） -->
  <div class="pipeline-stages-view">
    <!-- 状态统计栏 -->
    <div class="stages-stats-bar">
      <div 
        v-for="stat in stageStats" 
        :key="stat.key"
        :class="['stat-item', stat.key, { active: activeFilter === stat.key }]"
        @click="setFilter(stat.key)"
      >
        <span :class="['stat-dot', stat.key]"></span>
        <span class="stat-label">{{ stat.label }}</span>
        <span class="stat-count">{{ stat.count }}</span>
      </div>
    </div>

    <!-- 流水线主体 -->
    <div class="stages-main">
      <!-- 时间轴（左侧竖线） -->
      <div class="timeline-line"></div>

      <!-- 阶段列表 -->
      <div class="stages-list">
        <div 
          v-for="(stage, index) in displayStages" 
          :key="stage.id || stage.name"
          :class="['stage-item', `status-${stage.status}`, { expanded: expandedStage === stage.name }]"
        >
          <!-- 时间轴节点 -->
          <div :class="['timeline-node', `status-${stage.status}`]">
            <!-- 成功 -->
            <svg v-if="stage.status === 'success'" class="status-icon" viewBox="0 0 16 16">
              <path fill="currentColor" d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z"/>
            </svg>
            <!-- 失败 -->
            <svg v-else-if="stage.status === 'failed'" class="status-icon" viewBox="0 0 16 16">
              <path fill="currentColor" d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z"/>
            </svg>
            <!-- 运行中 -->
            <div v-else-if="stage.status === 'running'" class="running-spinner"></div>
            <!-- 等待中 -->
            <svg v-else-if="stage.status === 'waiting'" class="status-icon waiting" viewBox="0 0 16 16">
              <path fill="currentColor" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l2.817 2.817a.75.75 0 11-1.06 1.06l-2.817-2.817A6 6 0 012 8z"/>
            </svg>
            <!-- 待执行 -->
            <div v-else class="pending-dot"></div>
          </div>

          <!-- 阶段卡片 -->
          <div class="stage-card" @click="toggleExpand(stage.name)">
            <!-- 头部 -->
            <div class="stage-header">
              <div class="stage-info">
                <span class="stage-name">{{ stage.name }}</span>
                <span v-if="stage.type" :class="['stage-type-tag', stage.type]">
                  {{ getStageTypeLabel(stage.type) }}
                </span>
              </div>
              <div class="stage-meta">
                <span v-if="stage.status === 'running'" class="duration running">
                  <span class="pulse-dot"></span>
                  {{ formatRunningTime(stage.started_at) }}
                </span>
                <span v-else-if="stage.duration && stage.duration !== '-'" class="duration">
                  {{ stage.duration }}
                </span>
                <span :class="['status-badge', `status-${stage.status}`]">
                  {{ getStatusLabel(stage.status) }}
                </span>
              </div>
            </div>

            <!-- 进度条（运行中时显示） -->
            <div v-if="stage.status === 'running'" class="stage-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: getProgressWidth(stage) }"></div>
              </div>
            </div>

            <!-- 展开内容 -->
            <transition name="expand">
              <div v-if="expandedStage === stage.name" class="stage-detail">
                <!-- 步骤列表 -->
                <div v-if="stage.steps && stage.steps.length" class="steps-list">
                  <div 
                    v-for="step in stage.steps" 
                    :key="step.id || step.name"
                    :class="['step-item', `status-${step.status}`]"
                  >
                    <span :class="['step-status-dot', `status-${step.status}`]"></span>
                    <span class="step-name">{{ step.name }}</span>
                    <span class="step-duration">{{ step.duration || '-' }}</span>
                  </div>
                </div>

                <!-- 错误信息 -->
                <div v-if="stage.error_message" class="error-message">
                  <svg viewBox="0 0 16 16" class="error-icon">
                    <path fill="currentColor" d="M8 1.5a6.5 6.5 0 100 13 6.5 6.5 0 000-13zM0 8a8 8 0 1116 0A8 8 0 010 8zm6.5-.25A.75.75 0 017.25 7h1.5a.75.75 0 01.75.75v2.75h.25a.75.75 0 010 1.5h-2a.75.75 0 010-1.5h.25v-2h-.25a.75.75 0 01-.75-.75zM8 6a1 1 0 100-2 1 1 0 000 2z"/>
                  </svg>
                  <span>{{ stage.error_message }}</span>
                </div>

                <!-- 审批操作 -->
                <div v-if="stage.type === 'approval' && stage.status === 'waiting'" class="approval-actions">
                  <button class="btn btn-approve" @click.stop="$emit('approve', stage, 'approve')">
                    <svg viewBox="0 0 16 16"><path fill="currentColor" d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z"/></svg>
                    通过
                  </button>
                  <button class="btn btn-reject" @click.stop="$emit('approve', stage, 'reject')">
                    <svg viewBox="0 0 16 16"><path fill="currentColor" d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z"/></svg>
                    拒绝
                  </button>
                </div>

                <!-- 部署操作 -->
                <div v-if="stage.type === 'deploy' && stage.status === 'pending'" class="deploy-actions">
                  <button class="btn btn-deploy" @click.stop="$emit('deploy', stage)">
                    <svg viewBox="0 0 16 16"><path fill="currentColor" d="M8.75 1.75a.75.75 0 00-1.5 0V5H4.56L8 1.56l3.44 3.44H8.75V1.75zM1.5 8.75a.75.75 0 001.5 0v-3.5h10v3.5a.75.75 0 001.5 0v-4.5a1 1 0 00-1-1H2.5a1 1 0 00-1 1v4.5z"/></svg>
                    执行部署
                  </button>
                </div>

                <!-- 部署失败重试 -->
                <div v-if="stage.type === 'deploy' && stage.status === 'failed'" class="deploy-actions">
                  <button class="btn btn-retry-deploy" @click.stop="$emit('retry-deploy', stage)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                    重新部署
                  </button>
                </div>

                <!-- 查看日志按钮 -->
                <button class="btn btn-logs" @click.stop="$emit('view-logs', stage)">
                  <svg viewBox="0 0 16 16"><path fill="currentColor" d="M2 1.75C2 .784 2.784 0 3.75 0h6.586c.464 0 .909.184 1.237.513l2.914 2.914c.329.328.513.773.513 1.237v9.586A1.75 1.75 0 0113.25 16h-9.5A1.75 1.75 0 012 14.25V1.75zm1.75-.25a.25.25 0 00-.25.25v12.5c0 .138.112.25.25.25h9.5a.25.25 0 00.25-.25V4.664a.25.25 0 00-.073-.177l-2.914-2.914a.25.25 0 00-.177-.073H3.75zM8 5.5a.75.75 0 01.75.75v3.59l1.72-1.72a.75.75 0 111.06 1.06l-3 3a.75.75 0 01-1.06 0l-3-3a.75.75 0 111.06-1.06l1.72 1.72V6.25A.75.75 0 018 5.5z"/></svg>
                  查看日志
                </button>
              </div>
            </transition>
          </div>

          <!-- 连接线 -->
          <div v-if="index < displayStages.length - 1" :class="['connector', `status-${getConnectorStatus(index)}`]"></div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!stages || stages.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25"/>
      </svg>
      <p>暂无阶段数据</p>
      <span>运行流水线后将显示执行阶段</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  stages: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['approve', 'deploy', 'retry-deploy', 'view-logs', 'select-stage'])

const expandedStage = ref(null)
const activeFilter = ref('all')

// 阶段统计
const stageStats = computed(() => {
  const stats = {
    all: { key: 'all', label: '全部', count: props.stages.length },
    success: { key: 'success', label: '成功', count: 0 },
    failed: { key: 'failed', label: '失败', count: 0 },
    running: { key: 'running', label: '运行中', count: 0 },
    pending: { key: 'pending', label: '待执行', count: 0 }
  }
  
  props.stages.forEach(stage => {
    if (stage.status === 'success') stats.success.count++
    else if (stage.status === 'failed') stats.failed.count++
    else if (stage.status === 'running' || stage.status === 'deploying') stats.running.count++
    else stats.pending.count++
  })
  
  return Object.values(stats)
})

// 过滤后的阶段
const displayStages = computed(() => {
  if (activeFilter.value === 'all') return props.stages
  if (activeFilter.value === 'running') {
    return props.stages.filter(s => s.status === 'running' || s.status === 'deploying')
  }
  if (activeFilter.value === 'pending') {
    return props.stages.filter(s => s.status === 'pending' || s.status === 'waiting')
  }
  return props.stages.filter(s => s.status === activeFilter.value)
})

// 方法
const setFilter = (filter) => {
  activeFilter.value = filter
}

const toggleExpand = (stageName) => {
  expandedStage.value = expandedStage.value === stageName ? null : stageName
  const stage = props.stages.find(s => s.name === stageName)
  if (stage) emit('select-stage', stage)
}

const getStageTypeLabel = (type) => {
  const labels = {
    checkout: '检出',
    build: '构建',
    test: '测试',
    push: '推送',
    approval: '审批',
    deploy: '部署'
  }
  return labels[type] || type
}

const getStatusLabel = (status) => {
  const labels = {
    success: '成功',
    failed: '失败',
    running: '运行中',
    deploying: '部署中',
    waiting: '等待中',
    pending: '待执行',
    skipped: '已跳过',
    aborted: '已中止'
  }
  return labels[status] || status
}

const getConnectorStatus = (index) => {
  const currentStage = displayStages.value[index]
  const nextStage = displayStages.value[index + 1]
  
  if (currentStage.status === 'success') {
    if (nextStage.status === 'success' || nextStage.status === 'running') return 'success'
    return 'pending'
  }
  if (currentStage.status === 'failed') return 'failed'
  return 'pending'
}

const formatRunningTime = (startedAt) => {
  if (!startedAt) return '运行中...'
  const start = typeof startedAt === 'number' ? startedAt * 1000 : new Date(startedAt).getTime()
  const now = Date.now()
  const seconds = Math.floor((now - start) / 1000)
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

const getProgressWidth = (stage) => {
  // 模拟进度，实际可根据阶段平均耗时计算
  if (!stage.started_at) return '10%'
  const start = typeof stage.started_at === 'number' ? stage.started_at * 1000 : new Date(stage.started_at).getTime()
  const elapsed = (Date.now() - start) / 1000
  const estimatedTime = 60 // 估计60秒
  const progress = Math.min((elapsed / estimatedTime) * 100, 95)
  return `${progress}%`
}

// 自动展开运行中的阶段
watch(() => props.stages, (newStages) => {
  const runningStage = newStages.find(s => s.status === 'running' || s.status === 'deploying')
  if (runningStage && !expandedStage.value) {
    expandedStage.value = runningStage.name
  }
}, { deep: true, immediate: true })
</script>

<style scoped>
.pipeline-stages-view {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  overflow: hidden;
}

/* 状态统计栏 */
.stages-stats-bar {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.stat-item:hover, .stat-item.active {
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.stat-item.active {
  color: #111827;
  font-weight: 500;
}

.stat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #9ca3af;
}

.stat-dot.success { background: #10b981; }
.stat-dot.failed { background: #ef4444; }
.stat-dot.running { background: #3b82f6; }
.stat-dot.pending { background: #9ca3af; }

.stat-count {
  font-weight: 600;
  color: #374151;
}

/* 主体区域 */
.stages-main {
  position: relative;
  padding: 24px;
}

.timeline-line {
  position: absolute;
  left: 39px;
  top: 24px;
  bottom: 24px;
  width: 2px;
  background: #e5e7eb;
}

.stages-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* 阶段项 */
.stage-item {
  display: flex;
  align-items: flex-start;
  position: relative;
}

.timeline-node {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border: 2px solid #e5e7eb;
  z-index: 1;
  flex-shrink: 0;
}

.timeline-node.status-success {
  background: #10b981;
  border-color: #10b981;
  color: #fff;
}

.timeline-node.status-failed {
  background: #ef4444;
  border-color: #ef4444;
  color: #fff;
}

.timeline-node.status-running {
  background: #3b82f6;
  border-color: #3b82f6;
}

.timeline-node.status-waiting {
  background: #fbbf24;
  border-color: #fbbf24;
  color: #fff;
}

.status-icon {
  width: 16px;
  height: 16px;
}

.running-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pending-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #9ca3af;
}

/* 阶段卡片 */
.stage-card {
  flex: 1;
  margin-left: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
}

.stage-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.stage-item.expanded .stage-card {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
}

.stage-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stage-name {
  font-weight: 500;
  color: #111827;
}

.stage-type-tag {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #e5e7eb;
  color: #6b7280;
}

.stage-type-tag.approval {
  background: #fef3c7;
  color: #d97706;
}

.stage-type-tag.deploy {
  background: #dbeafe;
  color: #2563eb;
}

.stage-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.duration {
  font-size: 13px;
  color: #6b7280;
}

.duration.running {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #3b82f6;
}

.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3b82f6;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.status-badge {
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status-badge.status-success {
  background: #d1fae5;
  color: #059669;
}

.status-badge.status-failed {
  background: #fee2e2;
  color: #dc2626;
}

.status-badge.status-running, .status-badge.status-deploying {
  background: #dbeafe;
  color: #2563eb;
}

.status-badge.status-waiting {
  background: #fef3c7;
  color: #d97706;
}

.status-badge.status-pending {
  background: #f3f4f6;
  color: #6b7280;
}

/* 进度条 */
.stage-progress {
  padding: 0 16px 12px;
}

.progress-bar {
  height: 4px;
  background: #e5e7eb;
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  border-radius: 2px;
  transition: width 0.3s ease;
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

/* 展开详情 */
.stage-detail {
  padding: 12px 16px;
  border-top: 1px solid #e5e7eb;
  background: #fff;
}

.expand-enter-active, .expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from, .expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

/* 步骤列表 */
.steps-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f9fafb;
  border-radius: 6px;
  font-size: 13px;
}

.step-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #9ca3af;
}

.step-status-dot.status-success { background: #10b981; }
.step-status-dot.status-failed { background: #ef4444; }
.step-status-dot.status-running { background: #3b82f6; }

.step-name {
  flex: 1;
  color: #374151;
}

.step-duration {
  color: #9ca3af;
  font-size: 12px;
}

/* 错误信息 */
.error-message {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fef2f2;
  border-radius: 6px;
  margin-bottom: 12px;
}

.error-icon {
  width: 16px;
  height: 16px;
  color: #ef4444;
  flex-shrink: 0;
  margin-top: 2px;
}

.error-message span {
  font-size: 13px;
  color: #991b1b;
  line-height: 1.5;
}

/* 操作按钮 */
.approval-actions, .deploy-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn svg {
  width: 14px;
  height: 14px;
}

.btn-approve {
  background: #10b981;
  color: #fff;
}

.btn-approve:hover {
  background: #059669;
}

.btn-reject {
  background: #fee2e2;
  color: #dc2626;
}

.btn-reject:hover {
  background: #fecaca;
}

.btn-deploy {
  background: #3b82f6;
  color: #fff;
}

.btn-deploy:hover {
  background: #2563eb;
}

.btn-retry-deploy {
  background: #f59e0b;
  color: #fff;
}

.btn-retry-deploy:hover {
  background: #d97706;
}

.btn-retry-deploy svg {
  width: 14px;
  height: 14px;
}

.btn-logs {
  background: #f3f4f6;
  color: #374151;
}

.btn-logs:hover {
  background: #e5e7eb;
}

/* 连接线 */
.connector {
  width: 2px;
  height: 16px;
  background: #e5e7eb;
  margin-left: 15px;
}

.connector.status-success {
  background: #10b981;
}

.connector.status-failed {
  background: #ef4444;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #9ca3af;
}

.empty-state svg {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 15px;
  color: #6b7280;
  margin: 0 0 4px;
}

.empty-state span {
  font-size: 13px;
}
</style>

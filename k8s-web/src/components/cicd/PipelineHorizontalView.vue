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

          <!-- 操作按鈕 -->
          <div class="actions-section">
            <!-- 审批等待提示（非管理员看到的禁用状态） -->
            <div
              v-if="selectedStage.type === 'approval' && selectedStage.status === 'waiting' && !props.canApprove"
              class="approval-no-perm-tip"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;color:#f59e0b">
                <circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>
              </svg>
              等待管理员审批
            </div>
            <button 
              v-if="selectedStage.type === 'approval' && selectedStage.status === 'waiting' && props.canApprove"
              class="action-btn approve"
              @click="$emit('approve', selectedStage.id, 'approve')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              通过审批
            </button>
            <button 
              v-if="selectedStage.type === 'approval' && selectedStage.status === 'waiting' && props.canApprove"
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
            <button
              v-if="selectedStage.type === 'deploy' && selectedStage.status === 'success'"
              class="action-btn rollback"
              @click="$emit('rollback', selectedStage)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="1 4 1 10 7 10"/>
                <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
              </svg>
              回滚
            </button>
            <!-- 查看Pod按钮：部署成功后显示 -->
            <button
              v-if="selectedStage.type === 'deploy' && selectedStage.status === 'success' && deployTarget.namespace && deployTarget.name"
              :class="['action-btn', 'view-pods', { loading: podLoading }]"
              @click="togglePodPanel"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="2" width="20" height="20" rx="3"/>
                <path d="M2 8h20"/>
                <path d="M8 2v6"/>
                <circle cx="7" cy="13" r="1.5"/>
                <circle cx="12" cy="13" r="1.5"/>
                <circle cx="7" cy="18" r="1.5"/>
                <circle cx="12" cy="18" r="1.5"/>
                <line x1="18" y1="12" x2="18" y2="19"/>
                <polyline points="15 16 18 19 21 16"/>
              </svg>
              {{ showPodPanel ? '收起Pod' : `查看Pod (${podCount})` }}
            </button>
            <button class="action-btn logs" @click="$emit('view-logs', selectedStage)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              查看日志
            </button>
          </div>

          <!-- Pod 列表面板（部署成功后展示） -->
          <transition name="slide-up">
            <div v-if="showPodPanel && deployTarget.namespace && deployTarget.name" class="pod-panel">
              <!-- 加载状态 -->
              <div v-if="podLoading" class="pod-loading">
                <div class="spinner-small"></div>
                <span>正在获取最新 Pod...</span>
              </div>
              <!-- 错误状态 -->
              <div v-else-if="podError" class="pod-error">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                <span>{{ podError }}</span>
                <button class="retry-link" @click="fetchDeployPods">重试</button>
              </div>
              <!-- Pod 列表 -->
              <template v-else-if="pods.length > 0">
                <div class="pod-panel-header">
                  <div class="pod-panel-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="pod-icon">
                      <rect x="2" y="2" width="20" height="20" rx="3"/><path d="M2 8h20"/><path d="M8 2v6"/>
                    </svg>
                    <span>关联 Pod</span>
                    <span v-if="deployTarget.namespace" class="pod-ns-tag">{{ deployTarget.namespace }}</span>
                    <span v-if="deployTarget.kind" class="pod-kind-tag">{{ deployTarget.kind }}</span>
                    <span class="pod-count-badge">{{ pods.length }}</span>
                  </div>
                  <router-link
                    v-if="props.clusterId && deployTarget.namespace"
                    :to="`/c/${props.clusterId}/workloads/pods?namespace=${encodeURIComponent(deployTarget.namespace)}`"
                    class="view-all-link"
                  >
                    Pod列表查看全部 →
                  </router-link>
                </div>
                <div class="pod-list">
                  <div
                    v-for="(pod, idx) in pods"
                    :key="pod.name"
                    :class="['pod-item', { 'is-latest': idx === 0, [`status-${pod.status.toLowerCase()}`]: true }]"
                  >
                    <!-- 最新标记 -->
                    <span v-if="idx === 0" class="latest-tag">最新</span>
                    <!-- Pod 状态指示 -->
                    <div :class="['pod-status-indicator', pod.status.toLowerCase()]"></div>
                    <div class="pod-info">
                      <div class="pod-name-row">
                        <span class="pod-name" :title="pod.name">{{ pod.name }}</span>
                        <span v-if="pod.nodeName !== '-'" class="pod-node">{{ pod.nodeName }}</span>
                      </div>
                      <div class="pod-meta-row">
                        <span :class="['pod-status-text', pod.status.toLowerCase()]">{{ pod.status }}</span>
                        <span class="pod-separator">|</span>
                        <span>{{ readyText(pod) }}</span>
                        <span class="pod-separator">|</span>
                        <span>重启 {{ pod.restartCount }} 次</span>
                        <span class="pod-separator">|</span>
                        <span>{{ formatAge(pod.createdAt) }}</span>
                        <span v-if="pod.podIP !== '-'" class="pod-separator">|</span>
                        <span v-if="pod.podIP !== '-'" class="pod-ip">{{ pod.podIP }}</span>
                      </div>
                    </div>
                    <div class="pod-actions">
                      <button
                        v-if="props.clusterId"
                        class="pod-mini-btn"
                        title="查看详情"
                        @click="viewPodDetail(pod)"
                      >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                          <circle cx="12" cy="12" r="3"/>
                        </svg>
                      </button>
                      <button class="pod-mini-btn" title="查看日志" @click="viewPodLogs(pod)">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                          <polyline points="14 2 14 8 20 8"/>
                        </svg>
                      </button>
                      <button class="pod-mini-btn" title="打开终端" @click="viewPodTerminal(pod)">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <polyline points="4 17 10 11 4 5"/>
                          <line x1="12" y1="19" x2="20" y2="19"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              </template>
              <!-- 空状态 -->
              <div v-else class="pod-empty">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="2" width="20" height="20" rx="3"/><path d="M2 8h20"/><path d="M8 2v6"/>
                </svg>
                <span>该部署暂无运行中的 Pod</span>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useDeploymentPods } from '@/composables/useDeploymentPods'

const router = useRouter()
const { pods, loading: podLoading, error: podError, latestPod, fetchPods, formatAge } = useDeploymentPods()

const props = defineProps({
  stages: { type: Array, default: () => [] },
  canApprove: { type: Boolean, default: false },
  pipeline: { type: Object, default: null },
  clusterId: { type: [String, Number], default: '' }
})

const emit = defineEmits(['approve', 'deploy', 'retry-deploy', 'view-logs', 'rollback', 'view-pods'])

const activeTab = ref('all')
const selectedStage = ref(null)
const showPodPanel = ref(false)
const podsFetched = ref(false)

// 从 stage.deploy_info 或 pipeline 配置提取部署目标
const deployTarget = computed(() => {
  const stage = selectedStage.value
  const p = props.pipeline
  // 优先级：stage.deploy_info（实际部署信息）> pipeline 配置（静态目标）
  const stageInfo = stage?.deploy_info || {}
  return {
    namespace: stageInfo.namespace || p?.target_namespace || '',
    name: stageInfo.workload_name || stageInfo.WorkloadName || p?.target_workload_name || '',
    kind: stageInfo.workload_kind || stageInfo.WorkloadKind || p?.target_workload_kind || 'Deployment',
  }
})

const podCount = computed(() => pods.value.length)

// Pod 面板切换 + 触发获取
async function togglePodPanel() {
  showPodPanel.value = !showPodPanel.value
  if (showPodPanel.value && !podsFetched.value) {
    await fetchDeployPods()
  }
}

// 获取关联 Pod
async function fetchDeployPods() {
  const { namespace, name, kind } = deployTarget.value
  if (!namespace || !name) return
  await fetchPods(namespace, name, kind)
  podsFetched.value = true
}

// Pod 列表中 ready 容器数
function readyText(pod) {
  const total = pod.containers.length
  const ready = pod.containers.filter(c => c.ready).length
  return `${ready}/${total} Ready`
}

// 跳转 Pod 详情
function viewPodDetail(pod) {
  if (!props.clusterId) return
  router.push({
    path: `/c/${props.clusterId}/workloads/pods`,
    query: { namespace: pod.namespace, name: pod.name }
  })
}

// 跳转 Pod 日志（通过 emit）
function viewPodLogs(pod) {
  emit('view-pods', { type: 'logs', pod })
}

// 跳转 Pod 终端（通过 emit）
function viewPodTerminal(pod) {
  emit('view-pods', { type: 'terminal', pod })
}

// 选中阶段时自动重置 Pod 面板
watch(selectedStage, async (stage) => {
  showPodPanel.value = false
  podsFetched.value = false
  // 如果选中部署成功阶段，自动拉取 Pod
  if (stage?.type === 'deploy' && stage?.status === 'success') {
    await nextTick()
    if (deployTarget.value.namespace && deployTarget.value.name) {
      showPodPanel.value = true
      await fetchDeployPods()
    }
  }
})

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
  align-items: center;
}

/* 审批无权限提示 */
.approval-no-perm-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
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

.action-btn.rollback {
  background: #8b5cf6;
  color: #fff;
}

.action-btn.rollback:hover {
  background: #7c3aed;
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

/* ---- Pod 列表面板 ---- */
.pod-panel {
  margin: 0 20px 20px;
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  overflow: hidden;
}

.pod-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
}

.pod-panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
}

.pod-icon {
  width: 18px;
  height: 18px;
  color: #3b82f6;
}

.pod-count-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  background: #eff6ff;
  color: #2563eb;
  font-weight: 600;
}

.pod-ns-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  background: #f0fdf4;
  color: #059669;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-weight: 500;
}

.pod-kind-tag {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  background: #fef3c7;
  color: #92400e;
  font-weight: 500;
}

.view-all-link {
  font-size: 13px;
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.view-all-link:hover {
  color: #2563eb;
}

.pod-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 24px;
  color: #6b7280;
  font-size: 13px;
}

.pod-loading .spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.pod-error {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 18px;
  background: #fef2f2;
  border-radius: 8px;
  margin: 16px;
  font-size: 13px;
  color: #991b1b;
}

.pod-error svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.retry-link {
  margin-left: auto;
  background: none;
  border: none;
  color: #2563eb;
  font-size: 13px;
  cursor: pointer;
  font-weight: 500;
}

.retry-link:hover {
  text-decoration: underline;
}

.pod-list {
  display: flex;
  flex-direction: column;
}

.pod-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  border-bottom: 1px solid #f3f4f6;
  transition: background 0.15s;
}

.pod-item:last-child {
  border-bottom: none;
}

.pod-item:hover {
  background: #f9fafb;
}

.pod-item.is-latest {
  background: linear-gradient(90deg, #f0fdf4 0%, #fff 40%);
}

.pod-item.status-terminating {
  opacity: 0.6;
  background: #fefce8;
}

.latest-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  background: #10b981;
  color: #fff;
  flex-shrink: 0;
}

.pod-status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.pod-status-indicator.running { background: #10b981; }
.pod-status-indicator.pending { background: #f59e0b; }
.pod-status-indicator.failed { background: #ef4444; }
.pod-status-indicator.succeeded { background: #6b7280; }
.pod-status-indicator.unknown { background: #9ca3af; }
.pod-status-indicator.terminating { background: #f59e0b; animation: pulse 1.5s infinite; }

.pod-info {
  flex: 1;
  min-width: 0;
}

.pod-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.pod-name {
  font-size: 13px;
  font-weight: 600;
  color: #111827;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pod-node {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #f3f4f6;
  color: #6b7280;
  flex-shrink: 0;
}

.pod-meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #9ca3af;
}

.pod-separator {
  color: #e5e7eb;
}

.pod-status-text { font-weight: 500; }
.pod-status-text.running { color: #059669; }
.pod-status-text.pending { color: #d97706; }
.pod-status-text.failed { color: #dc2626; }
.pod-status-text.succeeded { color: #6b7280; }
.pod-status-text.terminating { color: #d97706; }

.pod-ip {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
}

.pod-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.pod-mini-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  background: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  transition: all 0.2s;
}

.pod-mini-btn:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
  color: #374151;
}

.pod-mini-btn svg {
  width: 15px;
  height: 15px;
}

.pod-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 32px;
  color: #9ca3af;
  font-size: 13px;
}

.pod-empty svg {
  width: 36px;
  height: 36px;
  color: #d1d5db;
}

/* 查看 Pod 按钮 */
.action-btn.view-pods {
  background: #059669;
  color: #fff;
}

.action-btn.view-pods:hover {
  background: #047857;
}

.action-btn.view-pods.loading {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>

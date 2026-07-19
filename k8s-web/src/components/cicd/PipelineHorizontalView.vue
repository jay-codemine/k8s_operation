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
            <!-- 获取详情按钮：部署已触发后显示，查看最新 Pod 状态 -->
            <button
              v-if="canFetchDeployStatus"
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
              {{ showPodPanel ? '收起详情' : '获取详情' }}
            </button>
            <button class="action-btn logs" @click="$emit('view-logs', selectedStage)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              查看日志
            </button>
          </div>

          <!-- 部署详情面板（实时状态 + Pod 列表） -->
          <transition name="slide-up">
            <div v-if="showPodPanel" class="pod-panel">
              <!-- 加载状态 -->
              <div v-if="podLoading" class="pod-loading">
                <div class="spinner-small"></div>
                <span>正在获取部署详情...</span>
              </div>
              <!-- 错误状态 -->
              <div v-else-if="podError" class="pod-error">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                <span>{{ podError }}</span>
                <button class="retry-link" @click="fetchDeployStatus">重试</button>
              </div>
              <!-- 详情内容 -->
              <template v-else-if="deployStatus">
                <!-- 状态横幅（大厂风） -->
                <div :class="['deploy-status-banner', realStatusClass]">
                  <div class="banner-left">
                    <div class="banner-icon">
                      <svg v-if="realStatus === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
                      </svg>
                      <svg v-else-if="realStatus === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
                      </svg>
                      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                      </svg>
                    </div>
                    <div class="banner-text">
                      <div class="banner-title">{{ realStatusText }}</div>
                      <div class="banner-desc">{{ deployStatus.message }}</div>
                    </div>
                  </div>
                  <div v-if="workload" class="banner-metrics">
                    <div class="metric">
                      <span class="metric-val">{{ workload.ready }}/{{ workload.desired }}</span>
                      <span class="metric-label">就绪副本</span>
                    </div>
                    <div class="metric">
                      <span class="metric-val">{{ workload.updated }}</span>
                      <span class="metric-label">已更新</span>
                    </div>
                    <div class="metric">
                      <span class="metric-val">{{ workload.available }}</span>
                      <span class="metric-label">可用</span>
                    </div>
                  </div>
                </div>

                <!-- Pod 列表头部 -->
                <div class="pod-panel-header">
                  <div class="pod-panel-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="pod-icon">
                      <rect x="2" y="2" width="20" height="20" rx="3"/><path d="M2 8h20"/><path d="M8 2v6"/>
                    </svg>
                    <span>关联 Pod</span>
                    <span v-if="workload?.namespace" class="pod-ns-tag">{{ workload.namespace }}</span>
                    <span v-if="workload?.kind" class="pod-kind-tag">{{ workload.kind }}</span>
                    <span class="pod-count-badge">{{ pods.length }}</span>
                  </div>
                  <div class="pod-header-right">
                    <button class="refresh-link" @click="fetchDeployStatus">刷新</button>
                    <router-link
                      v-if="props.clusterId && workload?.namespace"
                      :to="`/c/${props.clusterId}/workloads/pods?namespace=${encodeURIComponent(workload.namespace)}`"
                      class="view-all-link"
                    >
                      查看全部 →
                    </router-link>
                  </div>
                </div>

                <!-- Pod 列表 -->
                <div v-if="pods.length > 0" class="pod-list">
                  <div
                    v-for="(pod, idx) in pods"
                    :key="pod.name"
                    :class="['pod-item', `status-${(pod.status || '').toLowerCase()}`, { 'is-latest': idx === 0 }]"
                  >
                    <div :class="['pod-status-indicator', (pod.status || '').toLowerCase()]"></div>
                    <div class="pod-info">
                      <div class="pod-name-row">
                        <span class="pod-name" :title="pod.name">{{ pod.name }}</span>
                        <span v-if="idx === 0" class="latest-tag">最新</span>
                        <span v-if="pod.node_name" class="pod-node">{{ pod.node_name }}</span>
                      </div>
                      <div class="pod-meta-row">
                        <span :class="['pod-status-text', (pod.status || '').toLowerCase()]">{{ pod.status }}</span>
                        <span class="pod-separator">|</span>
                        <span :class="{ 'text-danger': !pod.ready }">{{ readyText(pod) }}</span>
                        <span class="pod-separator">|</span>
                        <span :class="{ 'text-warn': pod.restart_count > 0 }">重启 {{ pod.restart_count }} 次</span>
                        <span class="pod-separator">|</span>
                        <span>{{ formatAge(pod.created_at) }}</span>
                        <span v-if="pod.pod_ip" class="pod-separator">|</span>
                        <span v-if="pod.pod_ip" class="pod-ip">{{ pod.pod_ip }}</span>
                      </div>
                      <div v-if="pod.reason" class="pod-reason-row">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                          <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
                        </svg>
                        <span>{{ pod.reason }}<template v-if="pod.message">：{{ pod.message }}</template></span>
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
                <!-- 空状态 -->
                <div v-else class="pod-empty">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="2" y="2" width="20" height="20" rx="3"/><path d="M2 8h20"/><path d="M8 2v6"/>
                  </svg>
                  <span>该部署暂无运行中的 Pod</span>
                </div>
              </template>
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
import { getDeployStatus } from '@/api/platform/pipeline'

const router = useRouter()

const props = defineProps({
  stages: { type: Array, default: () => [] },
  canApprove: { type: Boolean, default: false },
  pipeline: { type: Object, default: null },
  clusterId: { type: [String, Number], default: '' }
})

const emit = defineEmits(['approve', 'deploy', 'retry-deploy', 'view-logs', 'rollback', 'view-pods', 'refresh'])

const activeTab = ref('all')
const selectedStage = ref(null)
const showPodPanel = ref(false)
const podsFetched = ref(false)

// 部署实时状态（来自后端 deploy-status 端点，服务端解析集群，无需 X-Cluster-ID）
const podLoading = ref(false)
const podError = ref('')
const deployStatus = ref(null)

const pods = computed(() => deployStatus.value?.pods || [])
const workload = computed(() => deployStatus.value?.workload || null)
const realStatus = computed(() => deployStatus.value?.real_status || '')
const podCount = computed(() => pods.value.length)

// 真实部署状态展示
const realStatusText = computed(() => {
  const map = { success: '部署成功', failed: '部署失败', deploying: '部署进行中' }
  return map[realStatus.value] || '状态未知'
})
const realStatusClass = computed(() => `banner-${realStatus.value || 'unknown'}`)

// 从 stage.deploy_info 或 pipeline 配置提取部署目标（用于展示与跳转）
const deployTarget = computed(() => {
  const stage = selectedStage.value
  const p = props.pipeline
  const stageInfo = stage?.deploy_info || {}
  return {
    namespace: stageInfo.namespace || p?.target_namespace || '',
    name: stageInfo.workload_name || stageInfo.WorkloadName || p?.target_workload_name || '',
    kind: stageInfo.workload_kind || stageInfo.WorkloadKind || p?.target_workload_kind || 'Deployment',
  }
})

// 是否可获取部署详情（部署已触发，即非 pending 的部署阶段）
const canFetchDeployStatus = computed(() => {
  const stage = selectedStage.value
  return stage?.type === 'deploy' && stage?.status !== 'pending' && !!stage?.id
})

// 详情面板切换 + 触发获取
async function togglePodPanel() {
  showPodPanel.value = !showPodPanel.value
  if (showPodPanel.value && !podsFetched.value) {
    await fetchDeployStatus()
  }
}

// 获取部署实时状态与 Pod 列表
async function fetchDeployStatus() {
  const stage = selectedStage.value
  if (!stage?.id) return
  podLoading.value = true
  podError.value = ''
  try {
    const res = await getDeployStatus(stage.id)
    if (res.code === 0) {
      deployStatus.value = res.data
      podsFetched.value = true
      // 后端修正了卡住的阶段状态 → 通知父组件刷新阶段/流水线，消除“发布中”
      if (res.data?.reconciled) {
        emit('refresh')
      }
    } else {
      podError.value = res.msg || '获取部署详情失败'
    }
  } catch (e) {
    podError.value = e?.message || '获取部署详情失败'
  } finally {
    podLoading.value = false
  }
}

// 就绪容器数展示
function readyText(pod) {
  return `${pod.ready_containers ?? 0}/${pod.total_containers ?? 0} Ready`
}

// 相对时间格式化
function formatAge(ts) {
  if (!ts) return '-'
  const secs = Math.floor(Date.now() / 1000 - Number(ts))
  if (secs < 0) return '-'
  if (secs < 60) return `${secs}秒`
  if (secs < 3600) return `${Math.floor(secs / 60)}分钟`
  if (secs < 86400) return `${Math.floor(secs / 3600)}小时`
  return `${Math.floor(secs / 86400)}天`
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

// 选中阶段时自动重置详情面板
watch(selectedStage, async (stage) => {
  showPodPanel.value = false
  podsFetched.value = false
  deployStatus.value = null
  podError.value = ''
  // 选中已触发的部署阶段时，自动拉取实时状态
  if (stage?.type === 'deploy' && stage?.status !== 'pending' && stage?.id) {
    await nextTick()
    showPodPanel.value = true
    await fetchDeployStatus()
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
.pod-status-indicator.imagepullbackoff,
.pod-status-indicator.errimagepull,
.pod-status-indicator.crashloopbackoff,
.pod-status-indicator.createcontainererror,
.pod-status-indicator.createcontainerconfigerror,
.pod-status-indicator.invalidimagename { background: #ef4444; animation: pulse 1.5s infinite; }

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
.pod-status-text.imagepullbackoff,
.pod-status-text.errimagepull,
.pod-status-text.crashloopbackoff,
.pod-status-text.createcontainererror,
.pod-status-text.createcontainerconfigerror,
.pod-status-text.invalidimagename { color: #dc2626; font-weight: 600; }

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

/* 部署状态横幅（大厂风） */
.deploy-status-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  margin: 12px 14px 4px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.deploy-status-banner.banner-success {
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdf4 100%);
  border-color: #a7f3d0;
}

.deploy-status-banner.banner-failed {
  background: linear-gradient(135deg, #fef2f2 0%, #fff1f2 100%);
  border-color: #fecaca;
}

.deploy-status-banner.banner-deploying,
.deploy-status-banner.banner-unknown {
  background: linear-gradient(135deg, #eff6ff 0%, #f0f9ff 100%);
  border-color: #bfdbfe;
}

.banner-left {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.banner-icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.banner-icon svg {
  width: 24px;
  height: 24px;
}

.banner-success .banner-icon { background: #d1fae5; color: #059669; }
.banner-failed .banner-icon { background: #fee2e2; color: #dc2626; }
.banner-deploying .banner-icon,
.banner-unknown .banner-icon { background: #dbeafe; color: #2563eb; }
.banner-deploying .banner-icon svg { animation: spin 1.4s linear infinite; }

.banner-text {
  min-width: 0;
}

.banner-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin-bottom: 2px;
}

.banner-success .banner-title { color: #047857; }
.banner-failed .banner-title { color: #b91c1c; }
.banner-deploying .banner-title,
.banner-unknown .banner-title { color: #1d4ed8; }

.banner-desc {
  font-size: 12px;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 360px;
}

.banner-metrics {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 56px;
  padding: 6px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.metric-val {
  font-size: 17px;
  font-weight: 700;
  color: #111827;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  line-height: 1.2;
}

.metric-label {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

/* Pod 列表头部右侧操作 */
.pod-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh-link {
  border: none;
  background: transparent;
  color: #2563eb;
  font-size: 12px;
  cursor: pointer;
  padding: 2px 4px;
}

.refresh-link:hover {
  text-decoration: underline;
}

/* Pod 异常原因行 */
.pod-reason-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 6px;
  padding: 6px 8px;
  border-radius: 6px;
  background: #fef2f2;
  font-size: 12px;
  color: #b91c1c;
  line-height: 1.5;
}

.pod-reason-row svg {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  margin-top: 2px;
  color: #ef4444;
}

.pod-meta-row .text-danger { color: #dc2626; font-weight: 600; }
.pod-meta-row .text-warn { color: #d97706; font-weight: 600; }
</style>

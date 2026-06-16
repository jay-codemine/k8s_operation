/**
 * useResourceWatcher.js - K8s 资源状态实时监听 composable
 * 
 * 功能：
 * - 镜像更新后自动开启快速轮询，追踪 Rollout 状态变化
 * - 实时拉取资源事件（Normal/Warning），展示在事件面板
 * - 状态到达 Running/Complete 后自动停止监听
 * - 支持 Deployment / StatefulSet / DaemonSet / Job 等工作负载
 * - 多阶段进度追踪（更新 → 就绪 → 可用）
 * 
 * 设计参考：KubeSphere / Argo Rollouts 资源状态追踪机制
 */
import { ref, computed, onBeforeUnmount } from 'vue'

/**
 * @param {Function} fetchStatusFn - 获取资源最新状态的函数，返回 { status, readyReplicas, ... }
 * @param {Function} fetchEventsFn - 获取资源事件的函数，返回 events 数组
 */
export function useResourceWatcher() {
  // ========== 状态 ==========
  const watching = ref(false)          // 是否在监听中
  const watchTarget = ref(null)        // 当前监听的资源 { namespace, name, kind }
  const watchPhase = ref('')           // Updating / Progressing / Running / Failed / Complete
  const watchProgress = ref(0)         // 进度 0~100
  const watchEvents = ref([])          // 事件列表
  const watchStartTime = ref(null)     // 监听开始时间
  const watchElapsed = ref(0)          // 已用时间（秒）
  const watchDetail = ref(null)        // 详细副本信息 { desired, updated, ready, available }
  const watchSteps = ref([])           // 多阶段步骤状态

  let pollTimer = null
  let eventTimer = null
  let elapsedTimer = null
  let animFrame = null
  let targetProgress = 0

  // ========== 工具：兼容 snake_case 和 camelCase ==========
  const getField = (obj, ...keys) => {
    for (const k of keys) {
      if (obj[k] !== undefined && obj[k] !== null) return obj[k]
    }
    return 0
  }

  // ========== 进度平滑动画 ==========
  const animateProgress = () => {
    const current = watchProgress.value
    if (Math.abs(current - targetProgress) < 1) {
      watchProgress.value = targetProgress
      return
    }
    watchProgress.value = current + (targetProgress - current) * 0.15
    animFrame = requestAnimationFrame(animateProgress)
  }

  // ========== 核心：开始监听 ==========
  const startWatching = (target, options = {}) => {
    stopWatching() // 先停掉旧的

    const {
      getStatus,
      getEvents,
      onComplete,
      pollInterval = 2000,
      eventInterval = 4000,
      timeout = 300000, // 5 分钟
    } = options

    watchTarget.value = target
    watching.value = true
    watchPhase.value = 'Updating'
    watchProgress.value = 0
    targetProgress = 0
    watchEvents.value = []
    watchStartTime.value = Date.now()
    watchElapsed.value = 0
    watchDetail.value = null
    watchSteps.value = [
      { key: 'update', label: '镜像更新', status: 'active' },
      { key: 'ready', label: 'Pod 就绪', status: 'pending' },
      { key: 'available', label: '服务可用', status: 'pending' },
    ]

    // 计时器
    elapsedTimer = setInterval(() => {
      watchElapsed.value = Math.floor((Date.now() - watchStartTime.value) / 1000)
      // 超时自动停止
      if (Date.now() - watchStartTime.value > timeout) {
        watchPhase.value = 'Timeout'
        watchSteps.value = watchSteps.value.map(s => 
          s.status === 'active' ? { ...s, status: 'error' } : s
        )
        stopWatching()
      }
    }, 1000)

    // 状态轮询
    const pollStatus = async () => {
      if (!watching.value) return
      try {
        const status = await getStatus()
        if (!status) return

        // 兼容 snake_case / camelCase 字段名
        const desired = getField(status, 'desired_replicas', 'desiredReplicas', 'replicas') || 1
        const updated = getField(status, 'updated_replicas', 'updatedReplicas')
        const ready = getField(status, 'ready_replicas', 'readyReplicas')
        const available = getField(status, 'available_replicas', 'availableReplicas')
        const unavailable = getField(status, 'unavailable_replicas', 'unavailableReplicas')

        // 保存详细信息
        watchDetail.value = { desired, updated, ready, available, unavailable }

        // 多阶段加权进度计算
        // Phase 1: 更新副本 (0-40%)
        // Phase 2: Pod 就绪 (40-80%)
        // Phase 3: 服务可用 (80-100%)
        const updateProgress = Math.min(1, updated / Math.max(desired, 1))
        const readyProgress = Math.min(1, ready / Math.max(desired, 1))
        const availableProgress = Math.min(1, available / Math.max(desired, 1))

        const weightedProgress = Math.round(
          updateProgress * 40 + readyProgress * 40 + availableProgress * 20
        )
        targetProgress = Math.min(100, weightedProgress)

        // 启动平滑动画
        if (animFrame) cancelAnimationFrame(animFrame)
        animateProgress()

        // 更新步骤状态
        const newSteps = [...watchSteps.value]
        // Step 1: 镜像更新
        if (updated >= desired) {
          newSteps[0] = { ...newSteps[0], status: 'done' }
          newSteps[1] = { ...newSteps[1], status: ready >= desired ? 'done' : 'active' }
        } else {
          newSteps[0] = { ...newSteps[0], status: 'active' }
        }
        // Step 2: Pod 就绪
        if (ready >= desired) {
          newSteps[1] = { ...newSteps[1], status: 'done' }
          newSteps[2] = { ...newSteps[2], status: available >= desired ? 'done' : 'active' }
        }
        // Step 3: 服务可用
        if (available >= desired) {
          newSteps[2] = { ...newSteps[2], status: 'done' }
        }
        watchSteps.value = newSteps

        // 阶段判定
        const s = (status.status || '').toLowerCase()
        if (s === 'running' || s === 'available' || s === 'complete') {
          if (ready >= desired && updated >= desired && available >= desired) {
            watchPhase.value = 'Complete'
            targetProgress = 100
            watchProgress.value = 100
            watchSteps.value = watchSteps.value.map(step => ({ ...step, status: 'done' }))
            onComplete?.({ success: true, elapsed: watchElapsed.value })
            setTimeout(() => stopWatching(), 3000)
            return
          }
        }

        if (s === 'failed' || s === 'crashloopbackoff' || s === 'error') {
          watchPhase.value = 'Failed'
          watchSteps.value = watchSteps.value.map(step => 
            step.status === 'active' ? { ...step, status: 'error' } : step
          )
          onComplete?.({ success: false, elapsed: watchElapsed.value })
          setTimeout(() => stopWatching(), 5000)
          return
        }

        // 进行中阶段
        if (updated < desired) {
          watchPhase.value = 'Updating'
        } else if (ready < desired) {
          watchPhase.value = 'Progressing'
        } else {
          watchPhase.value = 'Progressing'
        }
      } catch (e) {
        console.warn('[ResourceWatcher] poll status error:', e)
      }
    }

    // 事件轮询
    const pollEvents = async () => {
      if (!watching.value) return
      try {
        const events = await getEvents()
        if (Array.isArray(events) && events.length > 0) {
          // 合并去重（基于 time + message）
          const seen = new Set(watchEvents.value.map(e => `${e.time}|${e.message}`))
          const newEvents = events.filter(e => !seen.has(`${e.time}|${e.message}`))
          if (newEvents.length > 0) {
            watchEvents.value = [...newEvents, ...watchEvents.value].slice(0, 50)
          }
        }
      } catch (e) {
        console.warn('[ResourceWatcher] poll events error:', e)
      }
    }

    // 立即执行一次
    pollStatus()
    pollEvents()

    // 定时轮询
    pollTimer = setInterval(pollStatus, pollInterval)
    eventTimer = setInterval(pollEvents, eventInterval)
  }

  // ========== 停止监听 ==========
  const stopWatching = () => {
    watching.value = false
    clearInterval(pollTimer)
    clearInterval(eventTimer)
    clearInterval(elapsedTimer)
    if (animFrame) cancelAnimationFrame(animFrame)
    pollTimer = null
    eventTimer = null
    elapsedTimer = null
    animFrame = null
  }

  // ========== 格式化 ==========
  const formatElapsed = (seconds) => {
    if (seconds < 60) return `${seconds}s`
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}m${s}s`
  }

  const phaseColor = (phase) => {
    const map = {
      Updating: '#f0a020',
      Progressing: '#4e8ff7',
      Running: '#52c41a',
      Complete: '#52c41a',
      Failed: '#ff4d4f',
      Timeout: '#fa8c16',
    }
    return map[phase] || '#8c8c8c'
  }

  const phaseIcon = (phase) => {
    const map = {
      Updating: '🔄',
      Progressing: '⏳',
      Running: '✅',
      Complete: '✅',
      Failed: '❌',
      Timeout: '⏰',
    }
    return map[phase] || '📌'
  }

  const phaseLabel = (phase) => {
    const map = {
      Updating: '正在更新',
      Progressing: '部署中',
      Running: '运行中',
      Complete: '部署完成',
      Failed: '部署失败',
      Timeout: '超时',
    }
    return map[phase] || phase
  }

  // ========== Cleanup ==========
  onBeforeUnmount(() => {
    stopWatching()
  })

  return {
    // state
    watching,
    watchTarget,
    watchPhase,
    watchProgress,
    watchEvents,
    watchElapsed,
    watchDetail,
    watchSteps,
    // methods
    startWatching,
    stopWatching,
    formatElapsed,
    phaseColor,
    phaseIcon,
    phaseLabel,
  }
}

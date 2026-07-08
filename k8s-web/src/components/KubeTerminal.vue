<!--
  KubeTerminal.vue - 大厂级容器终端组件
  支持：WebSocket 实时交互 / 自动 shell 检测 / 窗口自适应 / 心跳保活
  风格参考：KubeSphere / Rancher / Lens 容器终端
-->
<template>
  <teleport to="body">
    <transition name="terminal-slide">
      <div v-if="visible" class="kube-terminal-overlay" @click.self="$emit('close')">
        <div 
          class="kube-terminal-panel"
          :class="{ fullscreen: isFullscreen }"
          :style="panelStyle"
        >
          <!-- 标题栏 -->
          <div class="terminal-header" @mousedown="startDrag">
            <div class="header-left">
              <span class="terminal-icon">&#x2588;&#x2588;</span>
              <span class="terminal-title">
                {{ namespace }}/{{ podName }}
                <span class="container-tag">{{ containerName }}</span>
              </span>
              <span class="shell-badge" v-if="detectedShell">{{ detectedShell }}</span>
              <span class="status-dot" :class="connectionStatus"></span>
              <span class="status-text">{{ statusText }}</span>
            </div>
            <div class="header-actions">
              <button class="term-btn" @click="reconnect" title="重新连接" :disabled="connectionStatus === 'connected'">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/>
                </svg>
              </button>
              <button class="term-btn" @click="toggleFullscreen" :title="isFullscreen ? '退出全屏' : '全屏'">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <template v-if="!isFullscreen">
                    <polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/>
                    <line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/>
                  </template>
                  <template v-else>
                    <polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/>
                    <line x1="14" y1="10" x2="21" y2="3"/><line x1="3" y1="21" x2="10" y2="14"/>
                  </template>
                </svg>
              </button>
              <button class="term-btn close-btn" @click="$emit('close')" title="关闭终端">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 终端体 -->
          <div class="terminal-body" ref="terminalRef"></div>

          <!-- 底部状态栏 -->
          <div class="terminal-footer">
            <span class="footer-info">
              <span class="key-hint">Ctrl+C</span> 中断
              <span class="key-hint">Ctrl+D</span> 退出
              <span class="key-hint">exit</span> 断开
            </span>
            <span class="footer-meta">
              {{ termSize }}
            </span>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { useClusterStore } from '@/stores/cluster'

const props = defineProps({
  visible: { type: Boolean, default: false },
  namespace: { type: String, required: true },
  podName: { type: String, required: true },
  containerName: { type: String, default: '' },
  shell: { type: String, default: '' },
})

const emit = defineEmits(['close'])

// ========== State ==========
const terminalRef = ref(null)
const connectionStatus = ref('disconnected') // disconnected / connecting / connected / error
const detectedShell = ref('')
const isFullscreen = ref(false)
const termSize = ref('80×24')
let term = null
let fitAddon = null
let ws = null
let pingTimer = null
let connectTimer = null

// ========== Computed ==========
const statusText = computed(() => {
  const map = {
    disconnected: '已断开',
    connecting: '连接中...',
    connected: '已连接',
    error: '连接异常',
  }
  return map[connectionStatus.value] || ''
})

const panelStyle = computed(() => {
  if (isFullscreen.value) return {}
  return {
    width: '80vw',
    height: '60vh',
    maxWidth: '1200px',
    maxHeight: '800px',
  }
})

// ========== WebSocket URL ==========
const buildWsUrl = () => {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  // 判断是否是远程穿透访问
  const isRemoteAccess = !['localhost', '127.0.0.1'].includes(window.location.hostname)
  const host = isRemoteAccess ? 'james521.gnway.cc:80' : window.location.host
  const base = `${proto}//${host}`

  const token = localStorage.getItem('token') || sessionStorage.getItem('token') || ''
  const clusterStore = useClusterStore()
  const clusterId = clusterStore.current?.id || ''

  const params = new URLSearchParams({
    namespace: props.namespace,
    name: props.podName,
    container: props.containerName,
    shell: props.shell || '',
    token: token,
    cluster_id: String(clusterId),
  })

  return `${base}/api/v1/k8s/pod/terminal?${params.toString()}`
}

// ========== Terminal Setup ==========
const initTerminal = () => {
  if (term) return

  term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'bar',
    fontSize: 13,
    fontFamily: '"JetBrains Mono", "Cascadia Code", "Fira Code", Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1a1b26',    // Tokyo Night 风格
      foreground: '#a9b1d6',
      cursor: '#c0caf5',
      cursorAccent: '#1a1b26',
      selectionBackground: '#33467c',
      selectionForeground: '#c0caf5',
      black: '#32344a',
      red: '#f7768e',
      green: '#9ece6a',
      yellow: '#e0af68',
      blue: '#7aa2f7',
      magenta: '#ad8ee6',
      cyan: '#449dab',
      white: '#787c99',
      brightBlack: '#444b6a',
      brightRed: '#ff7a93',
      brightGreen: '#b9f27c',
      brightYellow: '#ff9e64',
      brightBlue: '#7da6ff',
      brightMagenta: '#bb9af7',
      brightCyan: '#0db9d7',
      brightWhite: '#acb0d0',
    },
    scrollback: 10000,
    tabStopWidth: 4,
    allowProposedApi: true,
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalRef.value)

  // 自适应尺寸
  requestAnimationFrame(() => {
    fitAddon.fit()
    updateTermSize()
  })

  // 用户输入 → WebSocket
  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ op: 'stdin', data }))
    }
  })

  // 窗口 resize
  const ro = new ResizeObserver(() => {
    if (fitAddon && term) {
      fitAddon.fit()
      updateTermSize()
      // 通知后端 resize
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          op: 'resize',
          cols: term.cols,
          rows: term.rows,
        }))
      }
    }
  })
  ro.observe(terminalRef.value)

  term._resizeObserver = ro
}

const updateTermSize = () => {
  if (term) {
    termSize.value = `${term.cols}×${term.rows}`
  }
}

// ========== WebSocket Connect ==========

// 安全关闭旧 WebSocket，先移除事件处理器防止异步回调干扰新连接
const closeOldWs = () => {
  if (ws) {
    // 移除所有回调，防止旧 WS 的 onclose/onerror 异步触发时污染新连接状态
    ws.onopen = null
    ws.onmessage = null
    ws.onerror = null
    ws.onclose = null
    try { ws.close() } catch { /* ignore */ }
    ws = null
  }
  clearTimeout(connectTimer)
  clearInterval(pingTimer)
}

const connect = () => {
  closeOldWs()

  connectionStatus.value = 'connecting'
  const url = buildWsUrl()
  console.log('[KubeTerminal] WebSocket URL:', url)

  try {
    ws = new WebSocket(url)
  } catch (e) {
    console.error('[KubeTerminal] WebSocket 创建失败:', e)
    connectionStatus.value = 'error'
    term?.write(`\r\n\x1b[1;31m✗ WebSocket 创建失败: ${e.message}\x1b[0m\r\n`)
    return
  }

  // 保存当前 ws 引用，用于在回调中判断是否仍是活跃连接
  const currentWs = ws

  // 连接超时检测：15秒内未 onopen 则报错
  connectTimer = setTimeout(() => {
    if (ws === currentWs && connectionStatus.value === 'connecting') {
      console.warn('[KubeTerminal] 连接超时 (15s)')
      connectionStatus.value = 'error'
      term?.write('\r\n\x1b[1;31m✗ 连接超时，请检查后端服务是否启动，或网络代理是否支持 WebSocket\x1b[0m\r\n')
      closeOldWs()
    }
  }, 15000)

  ws.onopen = () => {
    // 防止旧连接的 onopen 延迟触发
    if (ws !== currentWs) return
    clearTimeout(connectTimer)
    connectionStatus.value = 'connected'
    console.log('[KubeTerminal] WebSocket 已连接')
    term?.focus()

    // 初始发送 resize
    if (term) {
      currentWs.send(JSON.stringify({
        op: 'resize',
        cols: term.cols,
        rows: term.rows,
      }))
    }

    // 心跳
    pingTimer = setInterval(() => {
      if (currentWs.readyState === WebSocket.OPEN) {
        currentWs.send(JSON.stringify({ op: 'ping' }))
      }
    }, 25000)
  }

  ws.onmessage = (event) => {
    if (ws !== currentWs) return
    try {
      const msg = JSON.parse(event.data)
      switch (msg.op) {
        case 'stdout':
        case 'stderr':
          term?.write(msg.data)
          if (msg.data) {
            if (!detectedShell.value) {
              if (msg.data.includes('检测到 /bin/bash') || msg.data.includes('bash')) detectedShell.value = 'bash'
              else if (msg.data.includes('检测到 /bin/sh') || msg.data.includes('Connected')) detectedShell.value = 'sh'
              else if (msg.data.includes('检测到 /bin/zsh')) detectedShell.value = 'zsh'
              else if (msg.data.includes('检测到 /bin/ash')) detectedShell.value = 'ash'
            }
            if (msg.data.includes('所有 Shell 均不可用') || msg.data.includes('distroless')) {
              connectionStatus.value = 'error'
            }
          }
          break
        case 'pong':
          break
      }
    } catch {
      term?.write(event.data)
    }
  }

  ws.onerror = (event) => {
    if (ws !== currentWs) return
    clearTimeout(connectTimer)
    connectionStatus.value = 'error'
    console.error('[KubeTerminal] WebSocket 错误:', event)
    term?.write('\r\n\x1b[1;31m✗ WebSocket 连接异常\x1b[0m\r\n')
  }

  ws.onclose = (event) => {
    // 关键：如果已经被 reconnect/connect 替换为新 WS，忽略旧 WS 的 close 事件
    if (ws !== currentWs) return
    clearTimeout(connectTimer)
    clearInterval(pingTimer)
    connectionStatus.value = 'disconnected'
    console.log(`[KubeTerminal] WebSocket 关闭: code=${event.code}, reason=${event.reason}`)
    if (term) {
      term.write('\r\n\x1b[1;33m⚠ 终端连接已断开')
      if (event.code !== 1000) {
        term.write(` (code: ${event.code})`)
      }
      term.write('\x1b[0m\r\n')
      term.write('\x1b[90m按上方 ↻ 按钮或关闭后重新打开以重新连接\x1b[0m\r\n')
    }
  }
}

const reconnect = () => {
  closeOldWs()
  connectionStatus.value = 'disconnected'
  detectedShell.value = ''
  term?.clear()
  term?.write('\x1b[1;36m↻ 正在重新连接...\x1b[0m\r\n')
  setTimeout(connect, 300)
}

const disconnect = () => {
  closeOldWs()
  connectionStatus.value = 'disconnected'
}

// 彻底销毁 xterm 实例（关闭面板时调用，因为 v-if 会销毁 DOM）
const destroyTerminal = () => {
  if (term) {
    term._resizeObserver?.disconnect()
    term.dispose()
    term = null
  }
  if (fitAddon) {
    fitAddon = null
  }
  detectedShell.value = ''
}

// ========== Fullscreen ==========
const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    fitAddon?.fit()
    updateTermSize()
  })
}

// ========== Drag ==========
let dragState = null
const startDrag = (e) => {
  if (isFullscreen.value) return
  // only left button
  if (e.button !== 0) return
  // ignore clicks on buttons
  if (e.target.closest('.header-actions') || e.target.closest('.term-btn')) return
  
  dragState = { startX: e.clientX, startY: e.clientY }
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

const onDrag = () => {
  // Simplified - actual drag logic could be added
}
const stopDrag = () => {
  dragState = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// ========== Lifecycle ==========
watch(() => props.visible, async (val) => {
  if (val) {
    await nextTick()
    initTerminal()
    connect()
  } else {
    disconnect()
    // v-if="visible" 会销毁 DOM，必须同步销毁 xterm 实例
    // 否则再次打开时 initTerminal() 会因 if(term) return 跳过，
    // 导致新 DOM 没有 xterm 挂载，终端白屏
    destroyTerminal()
  }
})

onMounted(() => {
  if (props.visible) {
    nextTick(() => {
      initTerminal()
      connect()
    })
  }
})

onBeforeUnmount(() => {
  disconnect()
  destroyTerminal()
})
</script>

<style scoped>
/* ========== 弹层 ========== */
.kube-terminal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ========== 面板 ========== */
.kube-terminal-panel {
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  overflow: hidden;
  box-shadow:
    0 25px 60px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.06);
  background: #1a1b26;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.kube-terminal-panel.fullscreen {
  width: 100vw !important;
  height: 100vh !important;
  max-width: none !important;
  max-height: none !important;
  border-radius: 0;
}

/* ========== 标题栏 ========== */
.terminal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: linear-gradient(180deg, #24283b 0%, #1f2335 100%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  cursor: move;
  user-select: none;
  min-height: 40px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #a9b1d6;
  overflow: hidden;
}

.terminal-icon {
  font-size: 10px;
  color: #7aa2f7;
  letter-spacing: -2px;
}

.terminal-title {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.container-tag {
  display: inline-block;
  padding: 1px 8px;
  margin-left: 6px;
  background: rgba(122, 162, 247, 0.15);
  color: #7aa2f7;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.shell-badge {
  padding: 1px 6px;
  background: rgba(158, 206, 106, 0.15);
  color: #9ece6a;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: all 0.3s;
}

.status-dot.connected { background: #9ece6a; box-shadow: 0 0 6px rgba(158, 206, 106, 0.6); }
.status-dot.connecting { background: #e0af68; animation: pulse 1.5s infinite; }
.status-dot.disconnected { background: #565f89; }
.status-dot.error { background: #f7768e; box-shadow: 0 0 6px rgba(247, 118, 142, 0.6); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.status-text {
  font-size: 11px;
  color: #565f89;
}

/* ========== 标题栏按钮 ========== */
.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.term-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #565f89;
  cursor: pointer;
  transition: all 0.2s;
}

.term-btn:hover { background: rgba(255, 255, 255, 0.08); color: #a9b1d6; }
.term-btn:active { transform: scale(0.92); }
.term-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.term-btn.close-btn:hover { background: rgba(247, 118, 142, 0.2); color: #f7768e; }

/* ========== 终端体 ========== */
.terminal-body {
  flex: 1;
  overflow: hidden;
  padding: 4px 8px;
}

.terminal-body :deep(.xterm) {
  height: 100%;
}

.terminal-body :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

.terminal-body :deep(.xterm-viewport::-webkit-scrollbar) {
  width: 6px;
}

.terminal-body :deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}

.terminal-body :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.terminal-body :deep(.xterm-viewport::-webkit-scrollbar-thumb:hover) {
  background: rgba(255, 255, 255, 0.2);
}

/* ========== 底部状态栏 ========== */
.terminal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 16px;
  background: #1f2335;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  font-size: 11px;
  color: #565f89;
}

.footer-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.key-hint {
  display: inline-block;
  padding: 0 4px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  font-family: "JetBrains Mono", monospace;
  font-size: 10px;
  margin-right: 2px;
}

.footer-meta {
  font-family: "JetBrains Mono", monospace;
}

/* ========== 动画 ========== */
.terminal-slide-enter-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.terminal-slide-leave-active { transition: all 0.2s cubic-bezier(0.4, 0, 1, 1); }

.terminal-slide-enter-from,
.terminal-slide-leave-to {
  opacity: 0;
}

.terminal-slide-enter-from .kube-terminal-panel {
  transform: translateY(30px) scale(0.95);
  opacity: 0;
}

.terminal-slide-leave-to .kube-terminal-panel {
  transform: translateY(10px) scale(0.98);
  opacity: 0;
}
</style>

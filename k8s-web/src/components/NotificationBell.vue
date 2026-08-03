<template>
  <div class="nb-wrap" ref="wrapRef">
    <button class="nav-action-btn" title="通知" @click.stop="toggle">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
        <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
      </svg>
      <span v-if="unreadCount > 0" class="notification-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>

    <transition name="nb-drop">
      <div v-if="open" class="nb-panel" @click.stop>
        <div class="nb-header">
          <span class="nb-title">通知</span>
          <button v-if="unreadCount > 0" class="nb-read-all" @click="markAllRead">全部已读</button>
        </div>

        <div class="nb-list">
          <div v-if="loading && !events.length" class="nb-empty">加载中...</div>
          <div v-else-if="!events.length" class="nb-empty">
            <div class="nb-empty-icon">🔔</div>
            <div>暂无通知</div>
          </div>
          <div
            v-for="ev in events"
            :key="ev.id"
            class="nb-item"
            :class="{ 'nb-unread': isUnread(ev) }"
            @click="goDetail(ev)"
          >
            <span class="nb-dot" :class="ev.severity"></span>
            <div class="nb-item-main">
              <div class="nb-item-title">
                <span class="nb-rule">{{ ev.rule_name }}</span>
                <span class="nb-status" :class="ev.status">{{ statusText(ev.status) }}</span>
              </div>
              <div class="nb-summary">{{ ev.summary || ev.description || '—' }}</div>
              <div class="nb-time">{{ formatTime(ev.fired_at) }}</div>
            </div>
          </div>
        </div>

        <div class="nb-footer" @click="goAll">
          查看全部告警事件 →
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { listAlertEvents } from '@/api/monitoring'

const router = useRouter()

const open = ref(false)
const loading = ref(false)
const events = ref([])
const wrapRef = ref(null)

const READ_KEY = 'notif_last_read_at'
const lastReadAt = ref(Number(localStorage.getItem(READ_KEY) || 0))

// 未读 = 触发时间晚于上次已读时间
const isUnread = (ev) => (ev.fired_at || 0) * 1000 > lastReadAt.value
const unreadCount = computed(() => events.value.filter(isUnread).length)

const statusText = (s) => ({ firing: '告警中', resolved: '已恢复', silenced: '已静默' }[s] || s)

const formatTime = (ts) => {
  if (!ts) return ''
  const diff = Date.now() - ts * 1000
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  if (day < 30) return `${day} 天前`
  return new Date(ts * 1000).toLocaleDateString()
}

const load = async () => {
  loading.value = true
  try {
    const res = await listAlertEvents({ page: 1, size: 20 })
    if (res?.code === 0) {
      events.value = res.data?.items || []
    }
  } catch {
    // 监控服务未配置/不可用时静默降级为空列表
  } finally {
    loading.value = false
  }
}

const toggle = () => {
  open.value = !open.value
  if (open.value) load()
}

const markAllRead = () => {
  lastReadAt.value = Date.now()
  localStorage.setItem(READ_KEY, String(lastReadAt.value))
}

const goDetail = (ev) => {
  // 单项点击即视为已读该时间之前的全部
  if (isUnread(ev)) markAllRead()
  open.value = false
  router.push('/monitoring/alert-events')
}

const goAll = () => {
  markAllRead()
  open.value = false
  router.push('/monitoring/alert-events')
}

const onClickOutside = (e) => {
  if (wrapRef.value && !wrapRef.value.contains(e.target)) open.value = false
}

let timer = null
onMounted(() => {
  document.addEventListener('click', onClickOutside)
  load() // 首屏拉一次用于角标
  timer = setInterval(() => {
    if (!document.hidden) load()
  }, 60000)
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  clearInterval(timer)
})
</script>

<style scoped>
.nb-wrap { position: relative; }

/* 与 Layout 顶栏按钮一致（Layout 为 scoped 样式，无法直接复用） */
.nav-action-btn {
  position: relative;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0.5rem;
  background: transparent;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s ease;
}
.nav-action-btn:hover { background: #f1f5f9; color: #334155; }
.nav-action-btn svg { width: 100%; height: 100%; }

.notification-badge {
  position: absolute; top: 0.125rem; right: 0.125rem;
  min-width: 1rem; height: 1rem; padding: 0 0.25rem;
  background: linear-gradient(135deg, #ef4444 0%, #ec4899 100%);
  color: #fff; font-size: 0.625rem; font-weight: 600;
  border-radius: 0.5rem;
  display: flex; align-items: center; justify-content: center;
}

.nb-panel {
  position: absolute; top: calc(100% + 10px); right: 0; z-index: 900;
  width: 360px; max-height: 480px;
  display: flex; flex-direction: column;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0,0,0,.12);
  overflow: hidden;
}

.nb-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; border-bottom: 1px solid #f1f5f9;
}
.nb-title { font-size: 14px; font-weight: 600; color: #1a202c; }
.nb-read-all {
  border: none; background: none; cursor: pointer;
  font-size: 12px; color: #4f46e5;
}
.nb-read-all:hover { text-decoration: underline; }

.nb-list { flex: 1; overflow-y: auto; }

.nb-empty {
  padding: 48px 20px; text-align: center;
  color: #a0aec0; font-size: 13px;
}
.nb-empty-icon { font-size: 32px; margin-bottom: 8px; }

.nb-item {
  display: flex; gap: 10px;
  padding: 12px 16px; cursor: pointer;
  border-bottom: 1px solid #f7fafc;
  transition: background .15s;
}
.nb-item:hover { background: #f7fafc; }
.nb-unread { background: #f5f7ff; }
.nb-unread:hover { background: #eef1ff; }

.nb-dot {
  width: 8px; height: 8px; border-radius: 50%;
  margin-top: 6px; flex-shrink: 0; background: #cbd5e1;
}
.nb-dot.critical { background: #dc2626; }
.nb-dot.warning { background: #d97706; }
.nb-dot.info { background: #2563eb; }

.nb-item-main { flex: 1; min-width: 0; }
.nb-item-title { display: flex; align-items: center; gap: 8px; }
.nb-rule {
  font-size: 13px; font-weight: 600; color: #2d3748;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.nb-status {
  flex-shrink: 0; padding: 1px 8px; border-radius: 8px;
  font-size: 11px; font-weight: 500;
}
.nb-status.firing { background: #fef2f2; color: #dc2626; }
.nb-status.resolved { background: #ecfdf5; color: #059669; }
.nb-status.silenced { background: #f3f4f6; color: #6b7280; }

.nb-summary {
  margin-top: 3px; font-size: 12px; color: #718096;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.nb-time { margin-top: 3px; font-size: 11px; color: #a0aec0; }

.nb-footer {
  padding: 10px 16px; text-align: center;
  font-size: 12px; color: #4f46e5; cursor: pointer;
  border-top: 1px solid #f1f5f9;
  transition: background .15s;
}
.nb-footer:hover { background: #f7fafc; }

.nb-drop-enter-active { transition: all .18s ease; }
.nb-drop-leave-active { transition: all .12s ease; }
.nb-drop-enter-from, .nb-drop-leave-to { opacity: 0; transform: translateY(-6px); }
</style>

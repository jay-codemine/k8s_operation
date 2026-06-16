<template>
  <div class="monitor-layout">
    <!-- 顶部 Tab 导航 -->
    <div class="monitor-tabs">
      <router-link
        v-for="tab in tabs"
        :key="tab.path"
        :to="tab.path"
        class="tab-item"
        :class="{ active: isActive(tab.path) }"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        <span class="tab-label">{{ tab.label }}</span>
        <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
      </router-link>
    </div>
    <!-- 子页面内容（keep-alive 缓存监控总览，避免 Tab 切换时重建图表） -->
    <div class="monitor-content">
      <router-view v-slot="{ Component }">
        <keep-alive :include="['MonitoringOverview']">
          <component :is="Component" />
        </keep-alive>
      </router-view>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getAlertStats } from '@/api/monitoring'

const route = useRoute()
const alertCount = ref(0)

const tabs = ref([
  { path: '/monitoring', label: '监控总览', icon: '📊', badge: null },
  { path: '/monitoring/datasources', label: '数据源管理', icon: '🔌', badge: null },
  { path: '/monitoring/alert-rules', label: '告警规则', icon: '📋', badge: null },
  { path: '/monitoring/alert-events', label: '告警事件', icon: '🔔', badge: null },
  { path: '/monitoring/notify-channels', label: '通知渠道', icon: '📡', badge: null },
  { path: '/monitoring/silence-rules', label: '告警降噪', icon: '🔇', badge: null },
])

const isActive = (path) => {
  if (path === '/monitoring') {
    return route.path === '/monitoring'
  }
  return route.path.startsWith(path)
}

onMounted(async () => {
  try {
    const res = await getAlertStats()
    if (res?.code === 0 && res.data) {
      alertCount.value = res.data.total_firing || 0
      tabs.value[3].badge = alertCount.value > 0 ? alertCount.value : null
    }
  } catch {}
})
</script>

<style scoped>
.monitor-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.monitor-tabs {
  display: flex;
  gap: 4px;
  padding: 16px 24px 0;
  background: #fff;
  border-bottom: 1px solid #e8ecf0;
  position: sticky;
  top: 0;
  z-index: 10;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
  color: #5f6b7a;
  text-decoration: none;
  border-radius: 8px 8px 0 0;
  border: 1px solid transparent;
  border-bottom: none;
  transition: all 0.2s;
  position: relative;
  font-weight: 500;
}

.tab-item:hover {
  color: #4f46e5;
  background: #f5f3ff;
}

.tab-item.active {
  color: #4f46e5;
  background: #fff;
  border-color: #e8ecf0;
  margin-bottom: -1px;
  padding-bottom: 11px;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: #4f46e5;
  border-radius: 2px 2px 0 0;
}

.tab-icon {
  font-size: 16px;
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  font-size: 11px;
  font-weight: 600;
  color: #fff;
  background: #ef4444;
  border-radius: 9px;
  line-height: 1;
}

.monitor-content {
  flex: 1;
  overflow-y: auto;
  background: #f7f8fa;
}
</style>

<template>
  <div class="cluster-selector" :class="{ 'expanded': showDropdown }">
    <!-- 选择器触发按钮 -->
    <div class="selector-trigger" @click="toggleDropdown">
      <div class="selected-cluster">
        <div class="cluster-icon" :class="currentClusterStatus">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
          </svg>
        </div>
        <div class="cluster-info">
          <span class="cluster-label">{{ label }}</span>
          <span class="cluster-name">{{ displayName }}</span>
        </div>
      </div>
      <div class="selector-arrow">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" :class="{ 'rotated': showDropdown }">
          <path d="M7 10l5 5 5-5z"/>
        </svg>
      </div>
    </div>

    <!-- 下拉菜单 -->
    <transition name="dropdown">
      <div v-if="showDropdown" class="selector-dropdown">
        <!-- 搜索框 -->
        <div class="dropdown-search" v-if="clusters.length > 5">
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="搜索集群..."
            @click.stop
          />
        </div>

        <!-- 集群列表 -->
        <div class="dropdown-list">
          <!-- 全部集群选项 -->
          <div 
            v-if="showAllOption"
            class="cluster-option all-clusters"
            :class="{ 'active': modelValue === '' || modelValue === 'all' }"
            @click="selectCluster('')"
          >
            <div class="option-icon all">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M4 6h16v2H4zm0 5h16v2H4zm0 5h16v2H4z"/>
              </svg>
            </div>
            <div class="option-info">
              <span class="option-name">全部集群</span>
              <span class="option-count">{{ onlineCount }} 在线 / {{ clusters.length }} 总计</span>
            </div>
          </div>

          <!-- 分割线 -->
          <div v-if="showAllOption" class="dropdown-divider"></div>

          <!-- 单个集群选项 -->
          <div 
            v-for="cluster in filteredClusters" 
            :key="cluster.id"
            class="cluster-option"
            :class="{ 
              'active': String(modelValue) === String(cluster.id),
              'offline': cluster.status !== 0 && cluster.status !== 'online'
            }"
            @click="selectCluster(cluster.id)"
          >
            <div class="option-icon" :class="getStatusClass(cluster)">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
              </svg>
            </div>
            <div class="option-info">
              <span class="option-name">{{ cluster.name || cluster.cluster_name }}</span>
              <span class="option-meta">
                <span class="status-dot" :class="getStatusClass(cluster)"></span>
                {{ getStatusText(cluster) }}
                <span v-if="cluster.version" class="version">v{{ cluster.version }}</span>
              </span>
            </div>
            <div class="option-check" v-if="String(modelValue) === String(cluster.id)">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="filteredClusters.length === 0" class="empty-state">
            <span>{{ searchQuery ? '未找到匹配的集群' : '暂无可用集群' }}</span>
          </div>
        </div>

        <!-- 底部操作 -->
        <div class="dropdown-footer" v-if="showManageLink">
          <a href="/cluster/list" class="manage-link">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19.14 12.94c.04-.31.06-.63.06-.94 0-.31-.02-.63-.06-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
            </svg>
            管理集群
          </a>
        </div>
      </div>
    </transition>

    <!-- 点击外部关闭 -->
    <div v-if="showDropdown" class="dropdown-overlay" @click="showDropdown = false"></div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  clusters: {
    type: Array,
    default: () => []
  },
  label: {
    type: String,
    default: '集群'
  },
  showAllOption: {
    type: Boolean,
    default: true
  },
  showManageLink: {
    type: Boolean,
    default: false
  },
  placeholder: {
    type: String,
    default: '请选择集群'
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const showDropdown = ref(false)
const searchQuery = ref('')

// 计算在线集群数
const onlineCount = computed(() => {
  return props.clusters.filter(c => c.status === 0 || c.status === 'online').length
})

// 当前选中的集群
const currentCluster = computed(() => {
  if (!props.modelValue || props.modelValue === 'all') return null
  return props.clusters.find(c => String(c.id) === String(props.modelValue))
})

// 显示名称
const displayName = computed(() => {
  if (!props.modelValue || props.modelValue === 'all') {
    return props.showAllOption ? '全部集群' : props.placeholder
  }
  return currentCluster.value?.name || currentCluster.value?.cluster_name || props.placeholder
})

// 当前集群状态
const currentClusterStatus = computed(() => {
  if (!props.modelValue || props.modelValue === 'all') return 'all'
  return getStatusClass(currentCluster.value)
})

// 过滤后的集群列表
const filteredClusters = computed(() => {
  if (!searchQuery.value) return props.clusters
  const query = searchQuery.value.toLowerCase()
  return props.clusters.filter(c => {
    const name = (c.name || c.cluster_name || '').toLowerCase()
    return name.includes(query)
  })
})

// 获取状态样式类
function getStatusClass(cluster) {
  if (!cluster) return 'unknown'
  const status = cluster.status
  if (status === 0 || status === 'online') return 'online'
  if (status === 1 || status === 'offline') return 'offline'
  return 'error'
}

// 获取状态文本
function getStatusText(cluster) {
  if (!cluster) return '未知'
  const status = cluster.status
  if (status === 0 || status === 'online') return '在线'
  if (status === 1 || status === 'offline') return '离线'
  return '异常'
}

// 切换下拉框
function toggleDropdown() {
  showDropdown.value = !showDropdown.value
  if (!showDropdown.value) {
    searchQuery.value = ''
  }
}

// 选择集群
function selectCluster(id) {
  emit('update:modelValue', id)
  emit('change', id)
  showDropdown.value = false
  searchQuery.value = ''
}

// ESC 键关闭
function handleKeydown(e) {
  if (e.key === 'Escape') {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.cluster-selector {
  position: relative;
  min-width: 220px;
}

/* 选择器触发按钮 */
.selector-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: linear-gradient(135deg, #1a1f36 0%, #252b43 100%);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.selector-trigger:hover {
  border-color: rgba(99, 102, 241, 0.6);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.cluster-selector.expanded .selector-trigger {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

.selected-cluster {
  display: flex;
  align-items: center;
  gap: 10px;
}

.cluster-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.1);
}

.cluster-icon svg {
  width: 18px;
  height: 18px;
}

.cluster-icon.online {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.cluster-icon.offline {
  background: rgba(107, 114, 128, 0.15);
  color: #6b7280;
}

.cluster-icon.error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.cluster-icon.all {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.cluster-info {
  display: flex;
  flex-direction: column;
}

.cluster-label {
  font-size: 11px;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.cluster-name {
  font-size: 14px;
  font-weight: 500;
  color: #f3f4f6;
}

.selector-arrow {
  color: #9ca3af;
  transition: transform 0.2s ease;
}

.selector-arrow svg {
  width: 20px;
  height: 20px;
}

.selector-arrow svg.rotated {
  transform: rotate(180deg);
}

/* 下拉菜单 */
.selector-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  min-width: 280px;
  background: #1a1f36;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  overflow: hidden;
}

.dropdown-search {
  padding: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.dropdown-search input {
  width: 100%;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #f3f4f6;
  font-size: 13px;
  outline: none;
  transition: all 0.2s;
}

.dropdown-search input:focus {
  border-color: rgba(99, 102, 241, 0.5);
  background: rgba(255, 255, 255, 0.08);
}

.dropdown-search input::placeholder {
  color: #6b7280;
}

.dropdown-list {
  max-height: 320px;
  overflow-y: auto;
  padding: 6px;
}

.dropdown-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
  margin: 6px 0;
}

/* 集群选项 */
.cluster-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.cluster-option:hover {
  background: rgba(99, 102, 241, 0.1);
}

.cluster-option.active {
  background: rgba(99, 102, 241, 0.2);
}

.cluster-option.offline {
  opacity: 0.6;
}

.option-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
  flex-shrink: 0;
}

.option-icon svg {
  width: 16px;
  height: 16px;
}

.option-icon.online {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.option-icon.offline {
  background: rgba(107, 114, 128, 0.15);
  color: #6b7280;
}

.option-icon.error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.option-icon.all {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.option-info {
  flex: 1;
  min-width: 0;
}

.option-name {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #f3f4f6;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

.option-count {
  display: block;
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-dot.online {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.5);
}

.status-dot.offline {
  background: #6b7280;
}

.status-dot.error {
  background: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.5);
}

.version {
  padding: 1px 5px;
  background: rgba(99, 102, 241, 0.15);
  border-radius: 3px;
  font-size: 10px;
  color: #a5b4fc;
}

.option-check {
  color: #6366f1;
  flex-shrink: 0;
}

.option-check svg {
  width: 18px;
  height: 18px;
}

/* 空状态 */
.empty-state {
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

/* 底部链接 */
.dropdown-footer {
  padding: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.manage-link {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  color: #9ca3af;
  font-size: 12px;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.manage-link:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #f3f4f6;
}

.manage-link svg {
  width: 14px;
  height: 14px;
}

/* 遮罩层 */
.dropdown-overlay {
  position: fixed;
  inset: 0;
  z-index: 999;
}

/* 下拉动画 */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 滚动条 */
.dropdown-list::-webkit-scrollbar {
  width: 6px;
}

.dropdown-list::-webkit-scrollbar-track {
  background: transparent;
}

.dropdown-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.dropdown-list::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>

<template>
  <div class="namespace-selector" :class="{ 'expanded': showDropdown }">
    <!-- 选择器触发按钮 -->
    <div class="selector-trigger" @click="toggleDropdown">
      <div class="selected-namespace">
        <div class="namespace-icon">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M3 3h18v2H3V3zm0 4h12v2H3V7zm0 4h18v2H3v-2zm0 4h12v2H3v-2zm0 4h18v2H3v-2z"/>
          </svg>
        </div>
        <div class="namespace-info">
          <span class="namespace-label">{{ label }}</span>
          <span class="namespace-name">{{ displayName }}</span>
        </div>
      </div>
      <div class="selector-arrow">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" :class="{ 'rotated': showDropdown }">
          <path d="M7 10l5 5 5-5z"/>
        </svg>
      </div>
    </div>

    <!-- 权限限制提示 -->
    <div v-if="isRestricted && !loading" class="restriction-hint">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <span>仅显示授权范围内的命名空间</span>
    </div>

    <!-- 下拉菜单 -->
    <transition name="dropdown">
      <div v-if="showDropdown" class="selector-dropdown">
        <!-- 搜索框 -->
        <div class="dropdown-search" v-if="filteredNamespaces.length > 5">
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="搜索命名空间..."
            @click.stop
          />
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <span>加载中...</span>
        </div>

        <!-- 命名空间列表 -->
        <div v-else class="dropdown-list">
          <!-- 全部命名空间选项 -->
          <div 
            v-if="showAllOption && !isRestricted"
            class="namespace-option all-namespaces"
            :class="{ 'active': modelValue === '' || modelValue === 'all' }"
            @click="selectNamespace('')"
          >
            <div class="option-icon all">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M4 6h16v2H4zm0 5h16v2H4zm0 5h16v2H4z"/>
              </svg>
            </div>
            <div class="option-info">
              <span class="option-name">所有命名空间</span>
              <span class="option-count">共 {{ internalNamespaces.length }} 个</span>
            </div>
          </div>

          <!-- 分割线 -->
          <div v-if="showAllOption && !isRestricted" class="dropdown-divider"></div>

          <!-- 单个命名空间选项 -->
          <div 
            v-for="ns in filteredNamespaces" 
            :key="getNamespaceName(ns)"
            class="namespace-option"
            :class="{ 
              'active': modelValue === getNamespaceName(ns),
              'system': isSystemNamespace(getNamespaceName(ns))
            }"
            @click="selectNamespace(getNamespaceName(ns))"
          >
            <div class="option-icon" :class="{ 'system': isSystemNamespace(getNamespaceName(ns)) }">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M3 3h18v2H3V3zm0 4h12v2H3V7zm0 4h18v2H3v-2zm0 4h12v2H3v-2zm0 4h18v2H3v-2z"/>
              </svg>
            </div>
            <div class="option-info">
              <span class="option-name">{{ getNamespaceName(ns) }}</span>
              <span v-if="isSystemNamespace(getNamespaceName(ns))" class="system-badge">系统</span>
            </div>
            <div class="option-check" v-if="modelValue === getNamespaceName(ns)">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
              </svg>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="filteredNamespaces.length === 0" class="empty-state">
            <span>{{ searchQuery ? '未找到匹配的命名空间' : (isRestricted ? '无可访问的命名空间' : '暂无命名空间') }}</span>
          </div>
        </div>
      </div>
    </transition>

    <!-- 点击外部关闭 -->
    <div v-if="showDropdown" class="dropdown-overlay" @click="showDropdown = false"></div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { getAccessibleNamespaces } from '@/api/rbac'
import permissionStore from '@/stores/permission'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  // 集群ID（用于获取该集群下可访问的命名空间）
  clusterId: {
    type: [String, Number],
    required: true
  },
  // 外部传入的完整命名空间列表（可选，用于过滤）
  namespaces: {
    type: Array,
    default: () => []
  },
  // 获取命名空间列表的函数（可选，如果不提供 namespaces）
  fetchNamespaces: {
    type: Function,
    default: null
  },
  label: {
    type: String,
    default: '命名空间'
  },
  showAllOption: {
    type: Boolean,
    default: true
  },
  placeholder: {
    type: String,
    default: '请选择命名空间'
  },
  // 是否自动选择第一个可用的命名空间
  autoSelectFirst: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change', 'loaded'])

const showDropdown = ref(false)
const searchQuery = ref('')
const loading = ref(false)
const internalNamespaces = ref([])

// 系统命名空间列表
const systemNamespaces = ['kube-system', 'kube-public', 'kube-node-lease', 'default']

// 是否受权限限制
const isRestricted = computed(() => {
  if (permissionStore.state.isSuperAdmin) return false
  const accessible = permissionStore.getAccessibleNamespaces(props.clusterId)
  return accessible.length > 0 && !accessible.includes('*')
})

// 获取命名空间名称（支持字符串和对象格式）
function getNamespaceName(ns) {
  if (typeof ns === 'string') return ns
  return ns?.metadata?.name || ns?.name || String(ns)
}

// 检查是否是系统命名空间
function isSystemNamespace(name) {
  return systemNamespaces.includes(name) || name?.startsWith('kube-')
}

// 显示名称
const displayName = computed(() => {
  if (!props.modelValue || props.modelValue === 'all') {
    return props.showAllOption ? '所有命名空间' : props.placeholder
  }
  return props.modelValue
})

// 过滤后的命名空间列表
const filteredNamespaces = computed(() => {
  let list = internalNamespaces.value
  
  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    list = list.filter(ns => {
      const name = getNamespaceName(ns).toLowerCase()
      return name.includes(query)
    })
  }
  
  return list
})

// 加载命名空间（考虑权限过滤）
async function loadNamespaces() {
  if (!props.clusterId) {
    internalNamespaces.value = []
    return
  }
  
  loading.value = true
  
  try {
    let allNamespaces = []
    
    // 1. 获取完整的命名空间列表
    if (props.namespaces && props.namespaces.length > 0) {
      // 使用外部传入的列表
      allNamespaces = [...props.namespaces]
    } else if (props.fetchNamespaces) {
      // 使用自定义获取函数
      allNamespaces = await props.fetchNamespaces(props.clusterId)
    }
    
    // 2. 应用权限过滤
    if (permissionStore.state.isSuperAdmin) {
      // 超级管理员可以看到所有
      internalNamespaces.value = allNamespaces
    } else {
      // 获取用户在该集群可访问的命名空间
      const accessible = permissionStore.getAccessibleNamespaces(props.clusterId)
      
      if (accessible.length === 0 || accessible.includes('*')) {
        // 没有限制，显示所有
        internalNamespaces.value = allNamespaces
      } else if (accessible.includes('__none__')) {
        // 无任何权限
        internalNamespaces.value = []
      } else {
        // 过滤出有权限的
        internalNamespaces.value = allNamespaces.filter(ns => {
          const name = getNamespaceName(ns)
          return accessible.includes(name)
        })
      }
    }
    
    // 3. 自动选择第一个
    if (props.autoSelectFirst && internalNamespaces.value.length > 0 && !props.modelValue) {
      const firstName = getNamespaceName(internalNamespaces.value[0])
      emit('update:modelValue', firstName)
      emit('change', firstName)
    }
    
    emit('loaded', internalNamespaces.value)
  } catch (e) {
    console.error('[NamespaceSelector] 加载命名空间失败:', e)
    internalNamespaces.value = []
  } finally {
    loading.value = false
  }
}

// 切换下拉框
function toggleDropdown() {
  showDropdown.value = !showDropdown.value
  if (!showDropdown.value) {
    searchQuery.value = ''
  }
}

// 选择命名空间
function selectNamespace(name) {
  emit('update:modelValue', name)
  emit('change', name)
  showDropdown.value = false
  searchQuery.value = ''
}

// ESC 键关闭
function handleKeydown(e) {
  if (e.key === 'Escape') {
    showDropdown.value = false
  }
}

// 监听 clusterId 变化
watch(() => props.clusterId, (newVal) => {
  if (newVal) {
    loadNamespaces()
  } else {
    internalNamespaces.value = []
  }
}, { immediate: true })

// 监听外部 namespaces 变化
watch(() => props.namespaces, () => {
  if (props.namespaces && props.namespaces.length > 0) {
    loadNamespaces()
  }
}, { deep: true })

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

// 暴露方法供外部调用
defineExpose({
  refresh: loadNamespaces,
  getFilteredNamespaces: () => internalNamespaces.value
})
</script>

<style scoped>
.namespace-selector {
  position: relative;
  min-width: 200px;
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

.namespace-selector.expanded .selector-trigger {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

.selected-namespace {
  display: flex;
  align-items: center;
  gap: 10px;
}

.namespace-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.namespace-icon svg {
  width: 16px;
  height: 16px;
}

.namespace-info {
  display: flex;
  flex-direction: column;
}

.namespace-label {
  font-size: 10px;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.namespace-name {
  font-size: 13px;
  font-weight: 500;
  color: #f3f4f6;
}

.selector-arrow {
  color: #9ca3af;
  transition: transform 0.2s ease;
}

.selector-arrow svg {
  width: 18px;
  height: 18px;
}

.selector-arrow svg.rotated {
  transform: rotate(180deg);
}

/* 权限限制提示 */
.restriction-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  margin-top: 6px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: 6px;
  font-size: 11px;
  color: #f59e0b;
}

.restriction-hint svg {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

/* 下拉菜单 */
.selector-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  min-width: 260px;
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
  max-height: 300px;
  overflow-y: auto;
  padding: 6px;
}

.dropdown-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
  margin: 6px 0;
}

/* 加载状态 */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px;
  color: #9ca3af;
  font-size: 13px;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(99, 102, 241, 0.2);
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 命名空间选项 */
.namespace-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.namespace-option:hover {
  background: rgba(99, 102, 241, 0.1);
}

.namespace-option.active {
  background: rgba(99, 102, 241, 0.2);
}

.namespace-option.system {
  opacity: 0.7;
}

.option-icon {
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
  flex-shrink: 0;
}

.option-icon svg {
  width: 14px;
  height: 14px;
}

.option-icon.system {
  background: rgba(107, 114, 128, 0.15);
  color: #6b7280;
}

.option-icon.all {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.option-info {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.option-name {
  font-size: 13px;
  font-weight: 500;
  color: #f3f4f6;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-count {
  font-size: 11px;
  color: #9ca3af;
}

.system-badge {
  padding: 2px 6px;
  background: rgba(107, 114, 128, 0.15);
  border-radius: 3px;
  font-size: 10px;
  color: #9ca3af;
}

.option-check {
  color: #6366f1;
  flex-shrink: 0;
}

.option-check svg {
  width: 16px;
  height: 16px;
}

/* 空状态 */
.empty-state {
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
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

/**
 * 命名空间权限过滤组合函数
 * 用于工作负载等页面自动过滤用户有权限访问的命名空间
 */
import { ref, computed, onMounted, watch } from 'vue'
import { useClusterStore } from '@/stores/cluster'
import permissionStore from '@/stores/permission'
import namespaceApi from '@/api/cluster/config/namespace'

/**
 * 命名空间权限过滤 Hook
 * @param {Object} options 配置选项
 * @param {boolean} options.autoLoad - 是否自动加载命名空间（默认 true）
 * @param {string} options.defaultNamespace - 默认命名空间（默认 ''）
 * @param {boolean} options.showAllOption - 是否显示"所有命名空间"选项（默认 true）
 * @returns {Object} 命名空间相关状态和方法
 */
export function useFilteredNamespaces(options = {}) {
  const {
    autoLoad = true,
    defaultNamespace = '',
    showAllOption = true
  } = options

  // 状态
  const loading = ref(false)
  const error = ref('')
  const allNamespaces = ref([]) // 所有命名空间（从 K8s 获取）
  const selectedNamespace = ref(defaultNamespace)
  
  // 获取集群 store
  const clusterStore = useClusterStore()
  
  /**
   * 当前集群 ID
   */
  const clusterId = computed(() => {
    return clusterStore.current?.id || 0
  })

  /**
   * 用户可访问的命名空间列表（根据权限过滤）
   */
  const filteredNamespaces = computed(() => {
    if (!clusterId.value) return []
    
    // 超级管理员可以访问所有
    if (permissionStore.state.isSuperAdmin) {
      return allNamespaces.value
    }
    
    // 获取用户在该集群的权限配置
    const accessible = permissionStore.getAccessibleNamespaces(clusterId.value)
    
    // 空数组表示不限制（可访问所有）
    if (accessible.length === 0) {
      return allNamespaces.value
    }
    
    // 特殊标记表示无权限
    if (accessible.includes('__none__')) {
      return []
    }
    
    // 过滤只保留有权限的命名空间
    return allNamespaces.value.filter(ns => accessible.includes(ns))
  })

  /**
   * 是否有全部访问权限
   */
  const hasFullAccess = computed(() => {
    if (permissionStore.state.isSuperAdmin) return true
    const accessible = permissionStore.getAccessibleNamespaces(clusterId.value)
    return accessible.length === 0
  })

  /**
   * 命名空间选择器的选项（带"全部"选项）
   */
  const namespaceOptions = computed(() => {
    const options = []
    
    // 如果有全部访问权限且配置允许，添加"全部"选项
    if (showAllOption && hasFullAccess.value) {
      options.push({ value: '', label: '所有命名空间' })
    }
    
    // 添加可访问的命名空间
    filteredNamespaces.value.forEach(ns => {
      options.push({ value: ns, label: ns })
    })
    
    return options
  })

  /**
   * 权限提示信息
   */
  const permissionHint = computed(() => {
    if (permissionStore.state.isSuperAdmin) {
      return ''
    }
    
    const accessible = permissionStore.getAccessibleNamespaces(clusterId.value)
    if (accessible.length === 0) {
      return ''
    }
    
    if (accessible.includes('__none__')) {
      return '您暂无该集群的命名空间访问权限'
    }
    
    return `您在该集群可访问 ${accessible.length} 个命名空间`
  })

  /**
   * 加载命名空间列表
   */
  const loadNamespaces = async () => {
    if (!clusterId.value) {
      console.warn('[useFilteredNamespaces] 未选择集群，跳过加载命名空间')
      return
    }

    loading.value = true
    error.value = ''
    
    try {
      const res = await namespaceApi.list({ page: 1, limit: 500 })
      const list = res?.data?.list || res?.data || []
      
      allNamespaces.value = (Array.isArray(list) ? list : []).map(ns => {
        // 后端可能返回字符串或对象
        return typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
      }).filter(Boolean)
      
      console.log('[useFilteredNamespaces] 加载命名空间完成', {
        total: allNamespaces.value.length,
        filtered: filteredNamespaces.value.length,
        hasFullAccess: hasFullAccess.value
      })
      
      // 如果当前选中的命名空间不在可访问列表中，自动切换
      if (selectedNamespace.value && !hasFullAccess.value) {
        if (!filteredNamespaces.value.includes(selectedNamespace.value)) {
          selectedNamespace.value = filteredNamespaces.value[0] || ''
          console.log('[useFilteredNamespaces] 自动切换命名空间为', selectedNamespace.value)
        }
      }
    } catch (e) {
      console.error('[useFilteredNamespaces] 加载命名空间失败:', e)
      error.value = e?.msg || e?.message || '加载命名空间失败'
      // 失败时用默认值兜底
      allNamespaces.value = ['default', 'kube-system']
    } finally {
      loading.value = false
    }
  }

  /**
   * 检查是否可以访问指定命名空间
   * @param {string} namespace 命名空间名称
   * @returns {boolean}
   */
  const canAccessNamespace = (namespace) => {
    return permissionStore.canAccessNamespace(clusterId.value, namespace)
  }

  /**
   * 获取用于 API 请求的命名空间参数
   * @returns {string|undefined} 命名空间，如果选择"全部"则返回 undefined
   */
  const getNamespaceParam = () => {
    if (!selectedNamespace.value && hasFullAccess.value) {
      return undefined // 全部命名空间，API 不传该参数
    }
    return selectedNamespace.value || filteredNamespaces.value[0] || 'default'
  }

  // 监听集群变化，重新加载命名空间
  watch(clusterId, (newId, oldId) => {
    if (newId && newId !== oldId) {
      console.log('[useFilteredNamespaces] 集群切换，重新加载命名空间')
      loadNamespaces()
    }
  })

  // 自动加载
  if (autoLoad) {
    onMounted(async () => {
      // 确保权限已加载
      await permissionStore.loadPermissions()
      // 加载命名空间
      await loadNamespaces()
    })
  }

  return {
    // 状态
    loading,
    error,
    allNamespaces,
    filteredNamespaces,
    selectedNamespace,
    namespaceOptions,
    hasFullAccess,
    permissionHint,
    clusterId,
    
    // 方法
    loadNamespaces,
    canAccessNamespace,
    getNamespaceParam
  }
}

export default useFilteredNamespaces

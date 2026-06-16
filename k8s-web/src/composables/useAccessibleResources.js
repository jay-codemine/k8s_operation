/**
 * 可访问资源组合式函数
 * 基于 RBAC + Scope 权限模型，获取用户有权限访问的资源
 * 
 * 核心功能：
 * 1. 获取用户可访问的集群列表（基于 sys_user_cluster 表）
 * 2. 获取用户在指定集群可访问的命名空间列表
 * 3. 自动过滤无权限的资源
 */
import { ref, computed, watch } from 'vue'
import { getAccessibleClusters, getAccessibleNamespaces } from '@/api/rbac'
import { getClusterList } from '@/api/cluster'
import permissionStore from '@/stores/permission'

/**
 * 获取用户可访问的集群列表
 * 
 * @param {Object} options 配置选项
 * @param {boolean} options.immediate 是否立即加载，默认 true
 * @param {boolean} options.useAllForAdmin 超级管理员是否使用完整列表，默认 true
 * @returns {Object} { clusters, loading, error, refresh }
 */
export function useAccessibleClusters(options = {}) {
  const { immediate = true, useAllForAdmin = true } = options
  
  const clusters = ref([])
  const loading = ref(false)
  const error = ref(null)
  
  // 在线集群数量
  const onlineCount = computed(() => {
    return clusters.value.filter(c => c.status === 0 || c.status === 'online').length
  })
  
  // 加载可访问的集群
  async function loadClusters() {
    loading.value = true
    error.value = null
    
    try {
      // 超级管理员使用完整列表
      if (useAllForAdmin && permissionStore.state.isSuperAdmin) {
        const res = await getClusterList({ page: 1, limit: 1000 })
        const data = res.data || res
        clusters.value = data.list || data.items || data || []
      } else {
        // 普通用户使用权限过滤的列表
        const res = await getAccessibleClusters()
        const data = res.data || res
        clusters.value = data.list || data.items || data || []
      }
    } catch (e) {
      console.error('[useAccessibleClusters] 加载集群失败:', e)
      error.value = e.message || '加载集群失败'
      clusters.value = []
    } finally {
      loading.value = false
    }
  }
  
  // 刷新
  function refresh() {
    return loadClusters()
  }
  
  // 立即加载
  if (immediate) {
    loadClusters()
  }
  
  return {
    clusters,
    loading,
    error,
    onlineCount,
    refresh,
    loadClusters
  }
}

/**
 * 获取用户在指定集群可访问的命名空间列表
 * 
 * @param {Ref|number} clusterId 集群ID（可以是 ref 或普通值）
 * @param {Object} options 配置选项
 * @param {boolean} options.immediate 是否立即加载，默认 true
 * @param {Function} options.fetchAll 获取集群所有命名空间的函数
 * @returns {Object} { namespaces, loading, error, refresh, isRestricted }
 */
export function useAccessibleNamespaces(clusterId, options = {}) {
  const { immediate = true, fetchAll = null } = options
  
  const namespaces = ref([])
  const loading = ref(false)
  const error = ref(null)
  
  // 是否有命名空间限制（非超管且有具体限制）
  const isRestricted = computed(() => {
    if (permissionStore.state.isSuperAdmin) return false
    const accessible = permissionStore.getAccessibleNamespaces(getClusterIdValue())
    return accessible.length > 0 && !accessible.includes('*')
  })
  
  // 获取 clusterId 的实际值（支持 ref 和普通值）
  function getClusterIdValue() {
    return typeof clusterId === 'object' && clusterId.value !== undefined
      ? clusterId.value
      : clusterId
  }
  
  // 加载可访问的命名空间
  async function loadNamespaces() {
    const cid = getClusterIdValue()
    if (!cid) {
      namespaces.value = []
      return
    }
    
    loading.value = true
    error.value = null
    
    try {
      // 超级管理员获取所有命名空间
      if (permissionStore.state.isSuperAdmin) {
        if (fetchAll) {
          // 如果提供了获取全部命名空间的函数
          const allNs = await fetchAll(cid)
          namespaces.value = allNs
        } else {
          // 默认返回空数组表示所有
          namespaces.value = []
        }
        return
      }
      
      // 普通用户获取权限内的命名空间
      const res = await getAccessibleNamespaces(cid)
      const data = res.data || res
      const accessibleNs = data.namespaces || []
      
      // 如果返回空数组，表示可访问所有
      if (accessibleNs.length === 0) {
        if (fetchAll) {
          namespaces.value = await fetchAll(cid)
        } else {
          namespaces.value = []
        }
      } else {
        namespaces.value = accessibleNs
      }
    } catch (e) {
      console.error('[useAccessibleNamespaces] 加载命名空间失败:', e)
      error.value = e.message || '加载命名空间失败'
      namespaces.value = []
    } finally {
      loading.value = false
    }
  }
  
  // 刷新
  function refresh() {
    return loadNamespaces()
  }
  
  // 监听 clusterId 变化
  if (typeof clusterId === 'object' && clusterId.value !== undefined) {
    watch(clusterId, (newVal) => {
      if (newVal) {
        loadNamespaces()
      } else {
        namespaces.value = []
      }
    })
  }
  
  // 立即加载
  if (immediate && getClusterIdValue()) {
    loadNamespaces()
  }
  
  return {
    namespaces,
    loading,
    error,
    isRestricted,
    refresh,
    loadNamespaces
  }
}

/**
 * 过滤命名空间列表
 * 根据用户权限过滤命名空间
 * 
 * @param {number|string} clusterId 集群ID
 * @param {Array} namespaces 原始命名空间列表
 * @returns {Array} 过滤后的命名空间列表
 */
export function filterNamespacesByPermission(clusterId, namespaces) {
  // 超级管理员可以看到所有
  if (permissionStore.state.isSuperAdmin) return namespaces
  
  // 获取可访问的命名空间
  const accessible = permissionStore.getAccessibleNamespaces(clusterId)
  
  // 如果没有限制，返回所有
  if (accessible.length === 0 || accessible.includes('*')) return namespaces
  
  // 过滤
  return namespaces.filter(ns => {
    const name = typeof ns === 'string' ? ns : (ns.name || ns.metadata?.name)
    return accessible.includes(name)
  })
}

/**
 * 检查是否有访问集群的权限
 * 
 * @param {number|string} clusterId 集群ID
 * @param {string} action 操作类型 (view/create/update/delete/exec)
 * @returns {boolean}
 */
export function canAccessCluster(clusterId, action = 'view') {
  return permissionStore.canAccessCluster(clusterId, action)
}

/**
 * 检查是否有访问命名空间的权限
 * 
 * @param {number|string} clusterId 集群ID
 * @param {string} namespace 命名空间名称
 * @returns {boolean}
 */
export function canAccessNamespace(clusterId, namespace) {
  return permissionStore.canAccessNamespace(clusterId, namespace)
}

/**
 * 组合使用：同时获取可访问的集群和命名空间
 * 
 * @param {Object} options 配置选项
 * @returns {Object}
 */
export function useAccessibleResources(options = {}) {
  const selectedClusterId = ref(options.defaultClusterId || null)
  
  const { 
    clusters, 
    loading: clustersLoading, 
    error: clustersError,
    onlineCount,
    refresh: refreshClusters 
  } = useAccessibleClusters({ 
    immediate: options.immediate !== false 
  })
  
  const {
    namespaces,
    loading: namespacesLoading,
    error: namespacesError,
    isRestricted: namespacesRestricted,
    refresh: refreshNamespaces
  } = useAccessibleNamespaces(selectedClusterId, {
    immediate: false,
    fetchAll: options.fetchAllNamespaces
  })
  
  // 选择集群
  function selectCluster(clusterId) {
    selectedClusterId.value = clusterId
  }
  
  // 刷新所有
  async function refreshAll() {
    await refreshClusters()
    if (selectedClusterId.value) {
      await refreshNamespaces()
    }
  }
  
  return {
    // 集群相关
    clusters,
    selectedClusterId,
    clustersLoading,
    clustersError,
    onlineCount,
    selectCluster,
    refreshClusters,
    
    // 命名空间相关
    namespaces,
    namespacesLoading,
    namespacesError,
    namespacesRestricted,
    refreshNamespaces,
    
    // 工具方法
    refreshAll,
    canAccessCluster,
    canAccessNamespace,
    filterNamespacesByPermission
  }
}

export default useAccessibleResources

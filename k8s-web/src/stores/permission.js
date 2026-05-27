/**
 * 权限状态管理 - 三域六角色模型 (v2)
 * 核心逻辑：角色 → 三域 scope (platform/cluster/cicd) → 级别 (none/read/write/admin)
 * 补充细控：集群级 access_level + 命名空间限制
 */
import { reactive, computed, readonly } from 'vue'
import { getUserPermissions } from '@/api/rbac'

// ==================== 常量定义 ====================

/** 角色类型 */
export const ROLE_TYPES = {
  SUPER_ADMIN: 'super_admin',
  PLATFORM_ADMIN: 'platform_admin',
  DEVOPS: 'devops',
  DEVELOPER: 'developer',
  TESTER: 'tester',
  VIEWER: 'viewer',
  CUSTOM: 'custom',
  // deprecated
  CLUSTER_ADMIN: 'cluster_admin'
}

/** 三大功能域 */
export const SCOPES = {
  PLATFORM: 'platform', // 平台域
  CLUSTER: 'cluster',   // 集群域
  CICD: 'cicd'          // 发布域
}

/** 权限级别（单调递增） */
export const ACCESS_LEVELS = {
  NONE: 'none',
  READ: 'read',
  WRITE: 'write',
  ADMIN: 'admin'
}

/** 级别比较顺序 */
const LEVEL_ORDER = { none: 0, read: 1, write: 2, admin: 3 }
function levelGte(a, b) {
  return (LEVEL_ORDER[a] || 0) >= (LEVEL_ORDER[b] || 0)
}

/** 操作类型 */
export const ACTIONS = {
  VIEW: 'view',
  CREATE: 'create',
  UPDATE: 'update',
  DELETE: 'delete',
  EXEC: 'exec',
  MANAGE: 'manage'
}

/** 资源类型 */
export const RESOURCES = {
  CLUSTER: 'cluster',
  WORKLOAD: 'workload',
  NETWORK: 'network',
  CONFIG: 'config',
  STORAGE: 'storage',
  NODE: 'node',
  MONITOR: 'monitor',
  USER: 'user',
  ROLE: 'role',
  SETTINGS: 'settings',
  PIPELINE: 'pipeline',
  ARTIFACT: 'artifact'
}

/**
 * 路由权限配置（scope + minLevel）
 * 未列入的路由默认允许访问
 */
export const ROUTE_SCOPES = {
  // 🏛 平台域
  '/platform/health': { scope: 'platform', minLevel: 'read' },
  '/platform/settings': { scope: 'platform', minLevel: 'admin' },
  '/platform/appstore': { scope: 'platform', minLevel: 'write' },
  '/platform/appstore/records': { scope: 'platform', minLevel: 'write' },
  '/security/users': { scope: 'platform', minLevel: 'admin' },
  '/security/roles': { scope: 'platform', minLevel: 'admin' },
  '/security/authorization': { scope: 'platform', minLevel: 'write' },
  '/security/audit': { scope: 'platform', minLevel: 'read' },
  '/security/ai-approvals': { scope: 'platform', minLevel: 'read' },
  '/security/diagnosis': { scope: 'cluster', minLevel: 'read' },
  '/security/rbac/serviceaccounts': { scope: 'cluster', minLevel: 'read' },
  '/security/rbac/roles': { scope: 'cluster', minLevel: 'read' },
  '/security/rbac/rolebindings': { scope: 'cluster', minLevel: 'write' },
  '/security/rbac/permission-check': { scope: 'cluster', minLevel: 'read' },
  // 兼容旧路径
  '/users': { scope: 'platform', minLevel: 'admin' },
  '/rbac': { scope: 'platform', minLevel: 'admin' },
  '/user-permissions': { scope: 'platform', minLevel: 'admin' },

  // ☸ 集群域
  '/clusters': { scope: 'cluster', minLevel: 'read' },
  '/environments': { scope: 'cluster', minLevel: 'read' },

  // 🚀 发布域
  '/cicd/pipelines': { scope: 'cicd', minLevel: 'read' },
  '/cicd/pipelines/create': { scope: 'cicd', minLevel: 'write' },
  '/cicd/releases': { scope: 'cicd', minLevel: 'read' },
  '/cicd/templates': { scope: 'cicd', minLevel: 'admin' },
  '/cicd/approvals': { scope: 'cicd', minLevel: 'admin' },
  '/cicd/artifacts': { scope: 'cicd', minLevel: 'read' },
  '/cicd/agents': { scope: 'cicd', minLevel: 'admin' },
  '/images/repositories': { scope: 'cicd', minLevel: 'admin' },
  '/images/browse': { scope: 'cicd', minLevel: 'read' },
  '/images/cleanup': { scope: 'cicd', minLevel: 'admin' },

  // 监控（属于集群域 — 进入监控模块需 cluster:read，数据源管理需 platform:admin）
  '/monitoring': { scope: 'cluster', minLevel: 'read' },
  '/monitoring/datasources': { scope: 'platform', minLevel: 'admin' }
}

// ==================== 权限状态 ====================

const state = reactive({
  loaded: false,
  loading: false,

  // 用户基本信息
  userId: 0,
  username: '',
  isSuperAdmin: false,

  // 平台角色列表
  roles: [],

  // ⭐ 三域有效级别（取所有角色中的最高值）
  scopes: {
    platform: 'none',
    cluster: 'none',
    cicd: 'none'
  },

  // 集群级权限 { clusterId: { access_level, role_type, namespaces, can_view... } }
  clusterPermissions: {},
  accessibleClusterIds: [],

  // 权限定义列表
  permissions: []
})

// ==================== 计算属性 ====================

/** 用户角色类型列表 */
const roleTypes = computed(() => {
  const platformRoles = state.roles.map(r => r.role_type || r.name)
  const clusterRoles = Object.values(state.clusterPermissions)
    .map(p => p.role_type)
    .filter(Boolean)
  return [...new Set([...platformRoles, ...clusterRoles])]
})

/** 是否为管理员（平台域 admin） */
const isAdmin = computed(() => {
  return state.isSuperAdmin || levelGte(state.scopes.platform, 'admin')
})

/** 是否为集群管理员（集群域 admin） */
const isClusterAdmin = computed(() => {
  return state.isSuperAdmin || levelGte(state.scopes.cluster, 'admin')
})

/** 是否有开发权限（集群域 write） */
const isDeveloper = computed(() => {
  return isClusterAdmin.value || levelGte(state.scopes.cluster, 'write')
})

/** 是否有 CICD 管理权限 */
const isCICDAdmin = computed(() => {
  return state.isSuperAdmin || levelGte(state.scopes.cicd, 'admin')
})

// ==================== 权限检查方法 ====================

/**
 * ⭐ 核心：检查用户在指定域是否达到最低级别
 * @param {string} scope  - platform/cluster/cicd
 * @param {string} minLevel - none/read/write/admin
 */
function hasScopeAccess(scope, minLevel = 'read') {
  if (state.isSuperAdmin) return true
  const userLevel = state.scopes[scope] || 'none'
  return levelGte(userLevel, minLevel)
}

/**
 * 检查是否有访问某个菜单/路由的权限
 */
function canAccessMenu(path) {
  if (state.isSuperAdmin) return true
  const routeScope = ROUTE_SCOPES[path]
  if (!routeScope) return true // 未配置则默认允许
  return hasScopeAccess(routeScope.scope, routeScope.minLevel)
}

/**
 * 检查是否有访问某个集群的权限
 */
function canAccessCluster(clusterId, action = ACTIONS.VIEW) {
  if (state.isSuperAdmin) return true

  // 先检查角色级别 scope_cluster
  const clusterScope = state.scopes.cluster || 'none'
  switch (action) {
    case ACTIONS.VIEW:
      if (levelGte(clusterScope, 'read')) return true
      break
    case ACTIONS.CREATE:
    case ACTIONS.UPDATE:
    case ACTIONS.EXEC:
      if (levelGte(clusterScope, 'write')) return true
      break
    case ACTIONS.DELETE:
    case ACTIONS.MANAGE:
      if (levelGte(clusterScope, 'admin')) return true
      break
  }

  // 回退到集群级细粒度权限
  const perm = state.clusterPermissions[clusterId]
  if (!perm) return false

  // 优先使用 access_level
  const al = perm.access_level || 'none'
  switch (action) {
    case ACTIONS.VIEW: return levelGte(al, 'read')
    case ACTIONS.CREATE:
    case ACTIONS.UPDATE:
    case ACTIONS.EXEC: return levelGte(al, 'write')
    case ACTIONS.DELETE:
    case ACTIONS.MANAGE: return levelGte(al, 'admin')
    default: return levelGte(al, 'read')
  }
}

/**
 * 检查命名空间访问权限
 */
function canAccessNamespace(clusterId, namespace) {
  if (state.isSuperAdmin) return true
  if (levelGte(state.scopes.cluster, 'admin')) return true
  
  const perm = state.clusterPermissions[clusterId]
  if (!perm) return false
  if (!perm.namespaces || perm.namespaces.length === 0) return true
  if (perm.namespaces.includes('*')) return true
  return perm.namespaces.includes(namespace)
}

/**
 * 获取用户在某集群可访问的命名空间
 */
function getAccessibleNamespaces(clusterId) {
  if (state.isSuperAdmin) return []
  if (levelGte(state.scopes.cluster, 'admin')) return []
  
  const perm = state.clusterPermissions[clusterId]
  if (!perm) return ['__none__']
  if (!perm.namespaces || perm.namespaces.length === 0) return []
  if (perm.namespaces.includes('*')) return []
  return perm.namespaces
}

/**
 * 检查资源操作权限
 */
function hasPermission(resource, action, clusterId = null) {
  if (state.isSuperAdmin) return true
  if (clusterId) {
    if (!canAccessCluster(clusterId, action)) return false
  }
  // viewer 只能查看
  if (state.scopes.cluster === 'read' && state.scopes.cicd === 'read' && state.scopes.platform === 'none') {
    return action === ACTIONS.VIEW
  }
  return true
}

/**
 * 过滤命名空间列表
 */
function filterNamespaces(clusterId, namespaces) {
  if (state.isSuperAdmin) return namespaces
  const accessible = getAccessibleNamespaces(clusterId)
  if (accessible.length === 0) return namespaces
  return namespaces.filter(ns => {
    const name = typeof ns === 'string' ? ns : (ns.name || ns.metadata?.name)
    return accessible.includes(name)
  })
}

// ==================== 状态管理方法 ====================

/**
 * 加载用户权限信息
 */
async function loadPermissions(force = false) {
  if (state.loaded && !force) return
  if (state.loading) return

  state.loading = true

  try {
    const res = await getUserPermissions()
    const data = res.data || res

    state.userId = data.user_id || 0
    state.username = data.username || ''
    state.isSuperAdmin = data.is_super_admin || false
    state.roles = data.roles || []
    state.permissions = data.permissions || []

    // ⭐ 解析三域 scope（后端 v2 接口返回 scopes 对象）
    if (data.scopes) {
      state.scopes.platform = data.scopes.platform || 'none'
      state.scopes.cluster = data.scopes.cluster || 'none'
      state.scopes.cicd = data.scopes.cicd || 'none'
    } else {
      // 兼容旧接口：从角色列表推导
      state.scopes.platform = 'none'
      state.scopes.cluster = 'none'
      state.scopes.cicd = 'none'
      if (state.isSuperAdmin) {
        state.scopes.platform = 'admin'
        state.scopes.cluster = 'admin'
        state.scopes.cicd = 'admin'
      } else {
        for (const role of state.roles) {
          if (role.scope_platform && levelGte(role.scope_platform, state.scopes.platform)) {
            state.scopes.platform = role.scope_platform
          }
          if (role.scope_cluster && levelGte(role.scope_cluster, state.scopes.cluster)) {
            state.scopes.cluster = role.scope_cluster
          }
          if (role.scope_cicd && levelGte(role.scope_cicd, state.scopes.cicd)) {
            state.scopes.cicd = role.scope_cicd
          }
        }
      }
    }

    // 解析集群权限
    const clusterPerms = data.cluster_permissions || []
    state.clusterPermissions = {}
    state.accessibleClusterIds = []

    clusterPerms.forEach(cp => {
      let namespaces = []
      if (cp.namespaces) {
        try {
          namespaces = typeof cp.namespaces === 'string'
            ? JSON.parse(cp.namespaces)
            : cp.namespaces
        } catch (e) {
          namespaces = []
        }
      }

      state.clusterPermissions[cp.cluster_id] = {
        access_level: cp.access_level || 'read',
        can_view: cp.can_view,
        can_create: cp.can_create,
        can_update: cp.can_update,
        can_delete: cp.can_delete,
        can_exec: cp.can_exec,
        role_type: cp.role_type,
        namespaces: namespaces,
        expire_at: cp.expire_at
      }

      const al = cp.access_level || (cp.can_view ? 'read' : 'none')
      if (levelGte(al, 'read')) {
        state.accessibleClusterIds.push(cp.cluster_id)
      }
    })

    state.loaded = true
    console.log('[Permission] v2 权限加载成功', {
      userId: state.userId,
      isSuperAdmin: state.isSuperAdmin,
      scopes: { ...state.scopes },
      roles: state.roles.map(r => r.name),
      clusters: state.accessibleClusterIds
    })
  } catch (e) {
    console.error('[Permission] 加载权限失败', e)
    state.loaded = true
  } finally {
    state.loading = false
  }
}

/**
 * 清除权限信息（退出登录时调用）
 */
function clearPermissions() {
  state.loaded = false
  state.loading = false
  state.userId = 0
  state.username = ''
  state.isSuperAdmin = false
  state.roles = []
  state.scopes.platform = 'none'
  state.scopes.cluster = 'none'
  state.scopes.cicd = 'none'
  state.clusterPermissions = {}
  state.accessibleClusterIds = []
  state.permissions = []
}

/**
 * 初始化权限（登录成功后调用）
 */
function initPermissions(userInfo) {
  if (!userInfo) return

  state.userId = userInfo.user_id || userInfo.id || 0
  state.username = userInfo.username || ''
  state.isSuperAdmin = userInfo.is_super_admin || false
  state.roles = userInfo.roles || []

  // 解析 scopes
  if (userInfo.scopes) {
    state.scopes.platform = userInfo.scopes.platform || 'none'
    state.scopes.cluster = userInfo.scopes.cluster || 'none'
    state.scopes.cicd = userInfo.scopes.cicd || 'none'
  } else if (state.isSuperAdmin) {
    state.scopes.platform = 'admin'
    state.scopes.cluster = 'admin'
    state.scopes.cicd = 'admin'
  }

  // 解析集群权限
  if (userInfo.cluster_permissions) {
    const clusterPerms = userInfo.cluster_permissions
    state.clusterPermissions = {}
    state.accessibleClusterIds = []

    clusterPerms.forEach(cp => {
      let namespaces = []
      if (cp.namespaces) {
        try {
          namespaces = typeof cp.namespaces === 'string'
            ? JSON.parse(cp.namespaces)
            : cp.namespaces
        } catch (e) {
          namespaces = []
        }
      }

      state.clusterPermissions[cp.cluster_id] = {
        access_level: cp.access_level || 'read',
        can_view: cp.can_view,
        can_create: cp.can_create,
        can_update: cp.can_update,
        can_delete: cp.can_delete,
        can_exec: cp.can_exec,
        role_type: cp.role_type,
        namespaces: namespaces
      }

      const al = cp.access_level || (cp.can_view ? 'read' : 'none')
      if (levelGte(al, 'read')) {
        state.accessibleClusterIds.push(cp.cluster_id)
      }
    })

    state.loaded = true
  }
}

// ==================== 导出 ====================

export const permissionStore = {
  // 状态（只读）
  state: readonly(state),

  // 计算属性
  roleTypes,
  isAdmin,
  isClusterAdmin,
  isDeveloper,
  isCICDAdmin,

  // ⭐ 核心权限检查
  hasScopeAccess,
  canAccessMenu,
  canAccessCluster,
  canAccessNamespace,
  getAccessibleNamespaces,
  hasPermission,
  filterNamespaces,

  // 状态管理
  loadPermissions,
  clearPermissions,
  initPermissions
}

export default permissionStore

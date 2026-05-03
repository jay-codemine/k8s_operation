// src/router/index.js
import {createRouter, createWebHistory} from 'vue-router'
import permissionStore from '@/stores/permission'

import Login from '@/views/auth/Login.vue'
import Layout from '@/components/Layout.vue'
import Dashboard from '@/views/dashboard/Dashboard.vue'

/**
 * 路由权限配置
 * 角色分类:
 *   - super_admin: 超级管理员，全部权限
 *   - platform_admin: 平台管理员，平台级管理
 *   - cluster_admin: 集群管理员，集群级管理
 *   - cicd_admin: CI/CD 管理员
 *   - developer: 开发人员
 *   - viewer: 只读用户
 */
const routePermissions = {
  // ==================== 平台管理 ====================
  '/platform/health': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/platform/settings': ['super_admin', 'platform_admin'],
  '/platform/appstore': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/platform/appstore/records': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/platform/appstore/install': ['super_admin', 'platform_admin', 'cluster_admin'],
  
  // ==================== 用户与权限管理（精简后） ====================
  '/security/users': ['super_admin', 'platform_admin'],
  '/security/roles': ['super_admin', 'platform_admin'],
  '/security/authorization': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/diagnosis': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  
  // 兼容旧路径
  '/users': ['super_admin', 'platform_admin'],
  '/rbac': ['super_admin', 'platform_admin'],
  '/user-permissions': ['super_admin', 'platform_admin'],
  
  // ==================== 安全审计 ====================
  '/security/audit': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/ai-approvals': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  '/security/rbac/serviceaccounts': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/roles': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/rolebindings': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/permission-check': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  
  // ==================== CI/CD 流水线 ====================
  '/cicd/templates': ['super_admin', 'platform_admin'],
  '/cicd/pipelines': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/pipelines/create': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/releases': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/approvals': ['super_admin', 'platform_admin', 'cicd_admin'],
  '/cicd/artifacts': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'viewer'],
  '/cicd/agents': ['super_admin', 'platform_admin', 'cicd_admin'],
  
  // ==================== 镜像管理 ====================
  '/images/repositories': ['super_admin', 'platform_admin'],
  '/images/browse': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'viewer'],
  '/images/cleanup': ['super_admin', 'platform_admin'],
  
  // ==================== 环境管理 ====================
  '/environments': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  
  // ==================== 集群管理 ====================
  '/clusters': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin', 'developer', 'viewer']
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {path: '/', redirect: '/login'},
    {path: '/login', component: Login},

    {
      path: '/',
      component: Layout,
      meta: {requiresAuth: true},
      children: [
        // ✅ 首页入口：访问 / 自动跳转 dashboard
        {
          path: '',
          redirect: '/dashboard',
        },

        // ✅ 默认首页（快速入门页）
        {
          path: 'dashboard',
          component: () => import('@/views/dashboard/Dashboard.vue'),
          children: [
            {
              path: '',
              component: () => import('@/views/dashboard/Home.vue'),
            },
          ],
        },

        // 平台级：集群列表（选集群入口）
        {path: 'clusters', component: () => import('@/views/cluster/Clusters.vue')},
        { path: 'platform/health', component: () => import('@/views/platform/health/PlatformHealth.vue') },
        { path: 'platform/appstore', component: () => import('@/views/platform/appstore/AppStore.vue') },
        { path: 'platform/appstore/records', component: () => import('@/views/platform/appstore/AppInstallRecords.vue') },
        { path: 'platform/appstore/install/:id', component: () => import('@/views/platform/appstore/AppInstallDetail.vue') },
        { path: 'platform/settings', component: () => import('@/views/platform/settings/PlatformSettings.vue') },
        
        // 安全和 RBAC（精简后的5个模块）
        { path: 'security/users', component: () => import('@/views/security/UserManagement.vue') },
        { path: 'security/roles', component: () => import('@/views/security/RoleManagement.vue') },
        { path: 'security/authorization', component: () => import('@/views/security/AuthorizationManagement.vue') },
        { path: 'security/audit', component: () => import('@/views/security/audit/AuditLog.vue') },
        { path: 'security/ai-approvals', component: () => import('@/views/security/AIApprovals.vue') },
        { path: 'security/diagnosis', component: () => import('@/views/security/rbac/PermissionCheck.vue') },
        
        // 兼容旧路径
        { path: 'security/rbac/serviceaccounts', component: () => import('@/views/security/rbac/ServiceAccounts.vue') },
        { path: 'security/rbac/roles', component: () => import('@/views/security/rbac/Roles.vue') },
        { path: 'security/rbac/rolebindings', component: () => import('@/views/security/rbac/RoleBindings.vue') },
        { path: 'security/rbac/permission-check', component: () => import('@/views/security/rbac/PermissionCheck.vue') },

        // ✅ 集群级：所有 k8s 功能都放这里
        {
          path: 'c/:clusterId',
          component: () => import('@/layouts/ClusterLayout.vue'),
          children: [
            { path: 'nodes', component: () => import('@/views/cluster/Nodes.vue') },
            { path: 'namespaces', component: () => import('@/views/cluster/Namespaces.vue') },

            { path: 'workloads/pods', component: () => import('@/views/workloads/Pods.vue') },
            { path: 'workloads/deployments', component: () => import('@/views/workloads/Deployments.vue') },
            { path: 'workloads/statefulsets', component: () => import('@/views/workloads/Statefulsets.vue') },
            { path: 'workloads/daemonsets', component: () => import('@/views/workloads/Daemonsets.vue') },
            { path: 'workloads/jobs', component: () => import('@/views/workloads/Jobs.vue') },
            { path: 'workloads/cronjobs', component: () => import('@/views/workloads/Cronjobs.vue') },
            { path: 'networking/services', component: () => import('@/views/networking/Services.vue') },
            { path: 'networking/ingresses', component: () => import('@/views/networking/Ingress.vue') },

            { path: 'config/configmaps', component: () => import('@/views/config/ConfigMaps.vue') },
            { path: 'config/secrets', component: () => import('@/views/config/Secrets.vue') },

            { path: 'storage/storageclasses', component: () => import('@/views/storage/StorageClasses.vue') },
            { path: 'storage/persistentvolumes', component: () => import('@/views/storage/Persistentvolumes.vue') },
            { path: 'storage/persistentvolumeclaims', component: () => import('@/views/storage/Persistentvolumeclaims.vue') },
          ],
        },


        // 平台功能（不需要 clusterId）
        {path: 'users', component: () => import('@/views/platform/Users.vue')},
        {path: 'rbac', component: () => import('@/views/platform/RBACPermissions.vue')},
        {path: 'user-permissions', component: () => import('@/views/platform/UserPermissions.vue')},

        // CICD 流水线
        {path: 'cicd/pipelines', component: () => import('@/views/cicd/Pipelines.vue')},
        {path: 'cicd/pipelines/create', component: () => import('@/views/cicd/PipelineCreate.vue')},
        {path: 'cicd/pipelines/:id', component: () => import('@/views/cicd/PipelineDetail.vue')},
        {path: 'cicd/pipelines/:id/edit', component: () => import('@/views/cicd/PipelineCreate.vue')},
        {path: 'cicd/templates', component: () => import('@/views/cicd/PipelineTemplates.vue')},
        // CICD 发布管理
        {path: 'cicd/releases', component: () => import('@/views/cicd/Releases.vue')},
        // CICD 审批管理
        {path: 'cicd/approvals', component: () => import('@/views/cicd/Approvals.vue')},
        // CICD 制品库管理
        {path: 'cicd/artifacts', component: () => import('@/views/cicd/Artifacts.vue')},
        // CICD 构建探针管理
        {path: 'cicd/agents', component: () => import('@/views/cicd/BuildAgents.vue')},

        {
          path: 'images/repositories',
          component: () => import('@/views/images/ImageRepositories.vue')
        },
        {path: 'images/browse', component: () => import('@/views/images/Images.vue')},
        {path: 'images/browse/:repoId', component: () => import('@/views/images/Images.vue')},
        {path: 'images/cleanup', component: () => import('@/views/images/CleanupPolicies.vue')},
        {path: 'images/:repoId', component: () => import('@/views/images/Images.vue')},

        {path: 'environments', component: () => import('@/views/environments/K8sEnvironments.vue')},

        // // ✅ 旧路径：统一引导去 clusters（让用户先选集群）
        // {path: 'workloads/pods', redirect: '/clusters'},
        // {path: 'clusters/nodes', redirect: '/clusters'},
        // {path: 'clusters/namespaces', redirect: '/clusters'},
      ],
    },

    {
      path: '/:pathMatch(.*)*',
      component: () => import('@/views/error/NotFound.vue'),
    },
    // 403 权限拒绝页面
    {
      path: '/forbidden',
      component: () => import('@/views/error/Forbidden.vue'),
    }
  ],
})

router.beforeEach(async (to, from, next) => {
  const requiresAuth = to.matched.some((r) => r.meta.requiresAuth)
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  
  // 未登录时跳转到登录页
  if (requiresAuth && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }
  
  // 登录页直接放行
  if (to.path === '/login') {
    next()
    return
  }
  
  // 已登录时，确保加载了权限
  if (token && !permissionStore.state.loaded) {
    try {
      await permissionStore.loadPermissions()
    } catch (e) {
      console.error('加载权限失败', e)
    }
  }
  
  // 检查路由权限
  const routeRoles = routePermissions[to.path]
  if (routeRoles) {
    // 超级管理员可以访问所有页面
    if (permissionStore.state.isSuperAdmin) {
      next()
      return
    }
    
    // 检查用户角色
    const userRoles = permissionStore.roleTypes.value
    const hasPermission = routeRoles.some(role => userRoles.includes(role))
    
    if (!hasPermission) {
      // 无权限时跳转到 403 页面
      next({ 
        path: '/forbidden', 
        query: { 
          type: 'role',
          path: to.path,
          role: routeRoles.join(', ')
        } 
      })
      return
    }
  }
  
  // 集群级路由权限检查
  if (to.path.startsWith('/c/') && to.params.clusterId) {
    const clusterId = parseInt(to.params.clusterId)
    if (clusterId && !permissionStore.state.isSuperAdmin) {
      const canAccess = permissionStore.canAccessCluster(clusterId, 'view')
      if (!canAccess) {
        next({ 
          path: '/forbidden', 
          query: { 
            type: 'cluster',
            path: to.path,
            clusterId: clusterId
          } 
        })
        return
      }
    }
  }
  
  next()
})

export default router

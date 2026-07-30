// src/router/index.js
import {createRouter, createWebHistory} from 'vue-router'
import permissionStore from '@/stores/permission'
import { ROUTE_SCOPES } from '@/stores/permission'

import Login from '@/views/auth/Login.vue'
import Layout from '@/components/Layout.vue'
import Dashboard from '@/views/dashboard/Dashboard.vue'

/**
 * 路由权限配置 (v2: 基于三域 scope + 权限级别)
 * 见 stores/permission.js 中的 ROUTE_SCOPES 配置
 * 角色分类:
 *   - super_admin: 超级管理员，全部权限
 *   - platform_admin: 平台管理员
 *   - devops: 运维工程师
 *   - developer: 开发工程师
 *   - tester: 测试工程师
 *   - viewer: 观察者
 */

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {path: '/', redirect: '/login'},
    {path: '/login', component: Login},
    // License 激活页（公开：平台未授权时由 http 拦截器强制跳转至此）
    {path: '/license', component: () => import('@/views/platform/license/LicenseActivate.vue')},

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
        { path: 'platform/aiops', component: () => import('@/views/platform/AIOps.vue') },
        { path: 'platform/appstore', component: () => import('@/views/platform/appstore/AppStore.vue') },
        { path: 'platform/appstore/records', component: () => import('@/views/platform/appstore/AppInstallRecords.vue') },
        { path: 'platform/appstore/install/:id', component: () => import('@/views/platform/appstore/AppInstallDetail.vue') },
        { path: 'platform/settings', component: () => import('@/views/platform/settings/PlatformSettings.vue') },
        
        // 安全和 RBAC（兼容旧路径，已迁移至 /admin/*）
        { path: 'security/users', redirect: '/admin/users' },
        { path: 'security/roles', redirect: '/admin/roles' },
        { path: 'security/authorization', redirect: '/admin/roles' },
        { path: 'security/audit', redirect: '/admin/audit' },
        { path: 'security/ai-approvals', redirect: '/admin/approvals' },
        { path: 'security/diagnosis', redirect: '/admin/roles' },
        { path: 'security/ldap', redirect: '/admin/identity' },
        
        // 兼容旧路径
        { path: 'security/rbac/serviceaccounts', redirect: '/admin/service-accounts' },
        { path: 'security/rbac/roles', component: () => import('@/views/security/rbac/Roles.vue') },
        { path: 'security/rbac/rolebindings', component: () => import('@/views/security/rbac/RoleBindings.vue') },
        { path: 'security/rbac/permission-check', redirect: '/admin/roles' },

        // ⭐ 平台管理（IAM 统一收口）
        { path: 'admin/users', component: () => import('@/views/security/UserManagement.vue') },
        { path: 'admin/roles', component: () => import('@/views/admin/PermissionRoles.vue') },
        { path: 'admin/tenants', component: () => import('@/views/platform/TenantManagement.vue') },
        { path: 'admin/identity', component: () => import('@/views/admin/IdentitySource.vue') },
        { path: 'admin/approvals', component: () => import('@/views/admin/ApprovalCenter.vue') },
        { path: 'admin/audit', component: () => import('@/views/admin/AuditCenter.vue') },
        { path: 'admin/service-accounts', component: () => import('@/views/admin/ServiceAccountCenter.vue') },
        { path: 'admin/settings', component: () => import('@/views/platform/settings/PlatformSettings.vue') },

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
            { path: 'workloads/hpa', component: () => import('@/views/workloads/HPA.vue') },
            { path: 'workloads/vpa', component: () => import('@/views/workloads/VPA.vue') },
            { path: 'networking/services', component: () => import('@/views/networking/Services.vue') },
            { path: 'networking/ingresses', component: () => import('@/views/networking/Ingress.vue') },

            { path: 'config/configmaps', component: () => import('@/views/config/ConfigMaps.vue') },
            { path: 'config/secrets', component: () => import('@/views/config/Secrets.vue') },

            { path: 'storage/storageclasses', component: () => import('@/views/storage/StorageClasses.vue') },
            { path: 'storage/persistentvolumes', component: () => import('@/views/storage/Persistentvolumes.vue') },
            { path: 'storage/persistentvolumeclaims', component: () => import('@/views/storage/Persistentvolumeclaims.vue') },

            // 扩展资源（CRD/CR 管理）
            { path: 'extensions/crd', component: () => import('@/views/extensions/Customresourcedefinitions.vue') },
            { path: 'extensions/cr-instances', component: () => import('@/views/extensions/CrInstances.vue') },
            { path: 'extensions/yaml-workbench', component: () => import('@/views/extensions/YamlWorkbench.vue') },
          ],
        },


        // 平台功能（不需要 clusterId）
        {path: 'users', component: () => import('@/views/platform/Users.vue')},
        {path: 'rbac', component: () => import('@/views/platform/RBACPermissions.vue')},
        {path: 'user-permissions', component: () => import('@/views/platform/UserPermissions.vue')},

        // CICD 应用中心（统一应用视图）
        {path: 'cicd/apps', component: () => import('@/views/cicd/AppCenter.vue')},
        // CICD 流水线
        {path: 'cicd/pipelines', component: () => import('@/views/cicd/Pipelines.vue')},
        {path: 'cicd/pipelines/create', component: () => import('@/views/cicd/PipelineCreate.vue')},
        {path: 'cicd/pipelines/:id', component: () => import('@/views/cicd/PipelineDetail.vue')},
        {path: 'cicd/pipelines/:id/edit', component: () => import('@/views/cicd/PipelineCreate.vue')},
        // GitOps 流水线（ArgoCD）
        {path: 'cicd/gitops/create', component: () => import('@/views/cicd/GitOpsCreate.vue')},
        {path: 'cicd/gitops/:id/edit', component: () => import('@/views/cicd/GitOpsCreate.vue')},
        {path: 'cicd/gitops/releases', component: () => import('@/views/cicd/GitOpsReleases.vue')},
        {path: 'cicd/build-records', component: () => import('@/views/cicd/BuildRecords.vue')},
        {path: 'cicd/templates', component: () => import('@/views/cicd/PipelineTemplates.vue')},
        // CICD 发布管理
        {path: 'cicd/releases', component: () => import('@/views/cicd/Releases.vue')},
        {path: 'cicd/release-history', component: () => import('@/views/cicd/ReleaseHistory.vue')},
        // CICD 审批管理
        {path: 'cicd/approvals', component: () => import('@/views/cicd/Approvals.vue')},
        // CICD 审批策略设置
        {path: 'cicd/approval-policy', component: () => import('@/views/cicd/ApprovalPolicy.vue')},
        // CICD 制品库管理
        {path: 'cicd/artifacts', component: () => import('@/views/cicd/Artifacts.vue')},
        // CICD 构建探针管理
        {path: 'cicd/agents', component: () => import('@/views/cicd/BuildAgents.vue')},
        // CICD 快速接入
        {path: 'cicd/quick-onboard', component: () => import('@/views/cicd/QuickOnboard.vue')},
        // CICD 镜像晋级（独立顶层页：选流水线看晋级链，build once promote everywhere）
        {path: 'cicd/promotion', component: () => import('@/views/cicd/Promotion.vue')},
        // CICD 环境管理（全局环境 cicd_environment 增删改，统一维护晋级目标环境）
        {path: 'cicd/environments', component: () => import('@/views/cicd/CicdEnvironments.vue')},

        {
          path: 'images/repositories',
          component: () => import('@/views/images/ImageRepositories.vue')
        },
        {path: 'images/browse', component: () => import('@/views/images/Images.vue')},
        {path: 'images/browse/:repoId', component: () => import('@/views/images/Images.vue')},
        {path: 'images/cleanup', component: () => import('@/views/images/CleanupPolicies.vue')},
        {path: 'images/:repoId', component: () => import('@/views/images/Images.vue')},

        {path: 'environments', component: () => import('@/views/environments/K8sEnvironments.vue')},
        
        // 监控中心（子路由架构）
        {
          path: 'monitoring',
          component: () => import('@/views/monitoring/MonitorLayout.vue'),
          children: [
            { path: '', component: () => import('@/views/monitoring/Monitoring.vue') },
            { path: 'datasources', component: () => import('@/views/monitoring/Datasources.vue') },
            { path: 'alert-rules', component: () => import('@/views/monitoring/AlertRules.vue') },
            { path: 'alert-events', component: () => import('@/views/monitoring/AlertEvents.vue') },
            { path: 'notify-channels', component: () => import('@/views/monitoring/NotifyChannels.vue') },
            { path: 'silence-rules', component: () => import('@/views/monitoring/SilenceRules.vue') },
          ],
        },

        // CRD 扩展资源管理（需要通过集群上下文 /c/:clusterId/extensions/ 访问）
        // 保留顶层路由作为跳转入口 → 使用 localStorage 中缓存的集群 ID
        { 
          path: 'extensions/crd', 
          redirect: () => {
            const raw = localStorage.getItem('currentCluster')
            const cid = raw ? JSON.parse(raw)?.id : null
            return cid ? `/c/${cid}/extensions/crd` : '/clusters'
          }
        },
        { 
          path: 'extensions/cr-instances', 
          redirect: () => {
            const raw = localStorage.getItem('currentCluster')
            const cid = raw ? JSON.parse(raw)?.id : null
            return cid ? `/c/${cid}/extensions/cr-instances` : '/clusters'
          }
        },
        { 
          path: 'extensions/yaml-workbench', 
          redirect: () => {
            const raw = localStorage.getItem('currentCluster')
            const cid = raw ? JSON.parse(raw)?.id : null
            return cid ? `/c/${cid}/extensions/yaml-workbench` : '/clusters'
          }
        },

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

  // 登录页/License 激活页直接放行
  if (to.path === '/login' || to.path === '/license') {
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

  // 超级管理员跳过所有权限检查
  if (permissionStore.state.isSuperAdmin) {
    next()
    return
  }

  // ⭐ v2: 基于三域 scope 检查路由权限
  const routeScope = ROUTE_SCOPES[to.path]
  if (routeScope) {
    const hasAccess = permissionStore.hasScopeAccess(routeScope.scope, routeScope.minLevel)
    if (!hasAccess) {
      next({
        path: '/forbidden',
        query: {
          type: 'scope',
          path: to.path,
          scope: routeScope.scope,
          required: routeScope.minLevel
        }
      })
      return
    }
  }

  // 集群级路由权限检查
  if (to.path.startsWith('/c/') && to.params.clusterId) {
    const clusterId = parseInt(to.params.clusterId)
    if (clusterId) {
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

// ================================================================
// 全局 chunk 加载失败容错：新版部署后旧哈希 chunk 404 时强制全页刷新
// ================================================================
// Vite 按路由拆分代码，每次构建文件名带有内容哈希.
// 部署新版本后，旧浏览器 tab 里的路由懒加载会请求已被删除的旧 chunk,
// 导致 "Failed to fetch dynamically imported module" 白屏.
// 拦截此错误并强制刷新，用户无感升级.
// ================================================================
router.onError((error, to) => {
  if (error.message?.includes('Failed to fetch dynamically imported module')) {
    console.warn('[Router] chunk 加载失败（可能为新版部署）, 强制刷新', to?.fullPath)
    window.location.href = to?.fullPath || '/dashboard'
  }
})

// 兜底：捕获未被路由拦截的异步组件加载错误
window.addEventListener('unhandledrejection', (event) => {
  const msg = event.reason?.message || ''
  if (msg.includes('Failed to fetch dynamically imported module')) {
    console.warn('[Global] 异步模块加载失败, 刷新页面')
    window.location.reload()
  }
})

export default router

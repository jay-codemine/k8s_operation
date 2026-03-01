// src/router/index.js
import {createRouter, createWebHistory} from 'vue-router'

import Login from '@/views/auth/Login.vue'
import Layout from '@/components/Layout.vue'
import Dashboard from '@/views/dashboard/Dashboard.vue'

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
        { path: 'platform/settings', component: () => import('@/views/platform/settings/PlatformSettings.vue') },
        
        // 安全和 RBAC
        { path: 'security/audit', component: () => import('@/views/security/audit/AuditLog.vue') },
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
            { path: 'workloads/statefulsets', component: () => import('@/views/workloads/StatefulSets.vue') },
            { path: 'workloads/daemonsets', component: () => import('@/views/workloads/DaemonSets.vue') },
            { path: 'workloads/jobs', component: () => import('@/views/workloads/Jobs.vue') },
            { path: 'workloads/cronjobs', component: () => import('@/views/workloads/CronJobs.vue') },
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

        {
          path: 'images/repositories',
          component: () => import('@/views/images/ImageRepositories.vue')
        },
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
    }
  ],
})

router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some((r) => r.meta.requiresAuth)
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  
  if (requiresAuth && !token) {
    // 未登录时，带上原目标路径跳转到登录页
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router

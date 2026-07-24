<!-- src/layouts/ClusterLayout.vue -->
<template>
  <div class="cluster-layout">
    <div class="cluster-bar">
      <div class="left">
        <button class="btn btn-mini" @click="backToClusters">← 返回集群管理</button>
        <div class="cluster-info">
          <div class="name">{{ clusterName }}</div>
          <div class="meta">ClusterID: {{ clusterId }}</div>
        </div>
      </div>

      <div class="right">
        <span class="hint">当前所有请求会自动携带 X-Cluster-ID</span>
      </div>
    </div>

    <div class="cluster-main">
      <!-- 集群级菜单 -->
      <aside class="cluster-menu">
        <div class="menu-title">基础</div>
        <a
          v-if="canView('nodes')"
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/nodes`) }"
          @click="go('nodes')"
        >🖥️ 节点</a>
        <a
          v-if="canView('namespaces')"
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/namespaces`) }"
          @click="go('namespaces')"
        >📦 命名空间</a>

        <template v-if="canViewGroup('workloads')">
          <div class="menu-title">工作负载</div>
          <a
            v-if="canView('pods')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/pods`) }"
            @click="go('workloads/pods')"
          >🧬 Pods</a>
          <a
            v-if="canView('deployments')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/deployments`) }"
            @click="go('workloads/deployments')"
          >📌 Deployments</a>
          <a 
            v-if="canView('statefulsets')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/statefulsets`) }"
            @click="go('workloads/statefulsets')"
          >📦 StatefulSets</a>
          <a 
            v-if="canView('daemonsets')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/daemonsets`) }"
            @click="go('workloads/daemonsets')"
          >🛰️ DaemonSets</a>
          <a 
            v-if="canView('jobs')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/jobs`) }"
            @click="go('workloads/jobs')"
          >🧰 Jobs</a>
          <a 
            v-if="canView('cronjobs')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/cronjobs`) }"
            @click="go('workloads/cronjobs')"
          >⏰ CronJobs</a>
          <a
            v-if="canView('hpa')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/hpa`) }"
            @click="go('workloads/hpa')"
          >📈 HPA 弹性扩缩容</a>
          <a
            v-if="canView('vpa')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/workloads/vpa`) }"
            @click="go('workloads/vpa')"
          >📊 VPA 垂直扩缩容</a>
        </template>

        <template v-if="canViewGroup('config')">
          <div class="menu-title">配置</div>
          <a
            v-if="canView('configmaps')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/config/configmaps`) }"
            @click="go('config/configmaps')"
          >🧾 ConfigMaps</a>
          <a
            v-if="canView('secrets')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/config/secrets`) }"
            @click="go('config/secrets')"
          >🔐 Secrets</a>
        </template>

        <template v-if="canViewGroup('storage')">
          <div class="menu-title">存储</div>
          <a
            v-if="canView('storageclasses')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/storage/storageclasses`) }"
            @click="go('storage/storageclasses')"
          >💾 StorageClasses</a>
          <a
            v-if="canView('pv')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/storage/persistentvolumes`) }"
            @click="go('storage/persistentvolumes')"
          >💾 PV</a>
          <a
            v-if="canView('pvc')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/storage/persistentvolumeclaims`) }"
            @click="go('storage/persistentvolumeclaims')"
          >💾 PVC</a>
        </template>

        <template v-if="canViewGroup('networking')">
          <div class="menu-title">网络</div>
          <a
            v-if="canView('services')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/networking/services`) }"
            @click="go('networking/services')"
          >🌐 Services</a>
          <a
            v-if="canView('ingress')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/networking/ingresses`) }"
            @click="go('networking/ingresses')"
          >🚪 Ingress</a>
        </template>

        <template v-if="canViewGroup('extensions')">
          <div class="menu-title">扩展资源</div>
          <a
            v-if="canView('crd')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/extensions/crd`) }"
            @click="go('extensions/crd')"
          >🧩 CRD 资源管理</a>
          <a
            v-if="canView('crd')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/extensions/cr-instances`) }"
            @click="go('extensions/cr-instances')"
          >📋 CR 实例管理</a>
          <a
            v-if="canView('crd')"
            class="menu-item"
            :class="{ active: isActive(`/c/${clusterId}/extensions/yaml-workbench`) }"
            @click="go('extensions/yaml-workbench')"
          >📝 YAML 工作台</a>
        </template>
      </aside>

      <!-- 资源页内容 -->
      <section class="cluster-content">
        <router-view/>
      </section>
    </div>
  </div>
</template>

<script setup>
import {computed, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useClusterStore} from '@/stores/cluster'
import permissionStore from '@/stores/permission'

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()

/**
 * 基于实际授权的集群资源权限配置
 * 参考 Rancher/Kuboard/KubeSphere 权限模型
 * 
 * 角色权限层级：
 * - cluster_admin: 集群完整权限，可查看所有资源
 * - developer: 开发者权限，可查看常用工作负载、服务、配置
 * - viewer: 只读权限，仅查看基础资源
 * - cicd_admin: CI/CD管理员，可查看部署相关资源
 */
const ROLE_RESOURCE_MAP = {
  // 集群管理员 - 所有资源
  cluster_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'hpa', 'vpa',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'crd'
  ],
  // 开发者 - 常用资源（劅除节点、存储类等集群级资源）
  developer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'hpa', 'vpa',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets', 'namespaces', 'crd'
  ],
  // 只读 - 可查看所有资源（与 cluster_admin 相同，但无操作权限）
  viewer: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'hpa', 'vpa',
    'services', 'ingress', 'pv', 'pvc', 'storageclasses',
    'configmaps', 'secrets', 'namespaces', 'nodes', 'crd'
  ],
  // CI/CD管理员 - 部署相关
  cicd_admin: [
    'pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs',
    'hpa', 'vpa',
    'services', 'ingress', 'pvc', 'configmaps', 'secrets', 'namespaces', 'crd'
  ]
}

// 资源分组
const resourceGroups = {
  workloads: ['pods', 'deployments', 'statefulsets', 'daemonsets', 'jobs', 'cronjobs', 'hpa', 'vpa'],
  config: ['configmaps', 'secrets'],
  storage: ['storageclasses', 'pv', 'pvc'],
  networking: ['services', 'ingress'],
  extensions: ['crd']
}

/**
 * 获取用户的最高角色类型
 */
const userHighestRole = computed(() => {
  if (permissionStore.state.isSuperAdmin) return 'super_admin'
  
  const roleTypes = permissionStore.roleTypes.value
  
  // 角色优先级
  if (roleTypes.includes('super_admin')) return 'super_admin'
  if (roleTypes.includes('platform_admin')) return 'platform_admin'
  if (roleTypes.includes('cluster_admin')) return 'cluster_admin'
  if (roleTypes.includes('cicd_admin')) return 'cicd_admin'
  if (roleTypes.includes('developer')) return 'developer'
  if (roleTypes.includes('viewer')) return 'viewer'
  
  return null
})

/**
 * 获取用户可查看的资源列表
 */
const userAccessibleResources = computed(() => {
  const role = userHighestRole.value
  if (!role) return []
  
  // super_admin/platform_admin 等同于 cluster_admin
  if (role === 'super_admin' || role === 'platform_admin') {
    return ROLE_RESOURCE_MAP.cluster_admin
  }
  
  return ROLE_RESOURCE_MAP[role] || []
})

/**
 * 检查是否有权限查看资源（基于实际授权）
 */
const canView = (resource) => {
  if (permissionStore.state.isSuperAdmin) return true
  return userAccessibleResources.value.includes(resource)
}

/**
 * 检查是否有权限查看资源分组（分组中至少有一个资源可见）
 */
const canViewGroup = (group) => {
  if (permissionStore.state.isSuperAdmin) return true
  const resources = resourceGroups[group]
  if (!resources) return true
  return resources.some(r => canView(r))
}

// 1) 路由参数 clusterId（数字）
const clusterId = computed(() => Number(route.params.clusterId))

// 2) 集群名称：优先取 store.current
const clusterName = computed(() => {
  const c = clusterStore.current
  if (c?.id === clusterId.value) return c.cluster_name || `cluster-${c.id}`
  return `cluster-${clusterId.value}`
})

// 3) 进入 /c/:clusterId 时，保证 store.current 至少有 id（刷新不丢）
watchEffect(() => {
  const cid = clusterId.value
  if (!cid) return

  clusterStore.hydrate?.()

  if (!clusterStore.current || Number(clusterStore.current.id) !== cid) {
    clusterStore.setCurrent({id: cid, cluster_name: `cluster-${cid}`})
  }
})

/**
 * ✅ active 判断：用 startsWith 更稳（比如带 query/hash 时也能高亮）
 */
const isActive = (fullPrefix) => {
  return String(route.path || '').startsWith(fullPrefix)
}

// go：基于 /c/:clusterId 拼接
const go = (subPath) => {
  router.push(`/c/${clusterId.value}/${subPath}`)
}

const backToClusters = () => router.push('/clusters')
</script>

<style scoped>
.cluster-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 100vh;
  background: transparent;
  padding: 0.75rem; /* 12px → 0.75rem */
  box-sizing: border-box;
}

.cluster-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 0.875rem; /* 14px → 0.875rem */
  padding: 0.75rem 0.875rem;
}

.left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.cluster-info .name {
  font-weight: 800;
  font-size: 1rem;
}

.cluster-info .meta {
  color: #64748b;
  font-size: 0.75rem;
}

.hint {
  color: #64748b;
  font-size: 0.75rem;
}

.cluster-main {
  flex: 1;
  display: flex;
  margin-top: 0.75rem;
  gap: 0.75rem;
  min-height: 0;
  overflow: hidden;
}

.cluster-menu {
  flex-shrink: 0;
  width: 14rem; /* 240px → 14rem，会随字体缩放 */
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 0.875rem;
  padding: 0.75rem;
  overflow-y: auto;
}

.menu-title {
  margin: 0.625rem 0.375rem;
  color: #94a3b8;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.menu-item {
  display: block;
  padding: 0.625rem;
  border-radius: 0.625rem;
  cursor: pointer;
  color: #334155;
  text-decoration: none;
  margin: 0.375rem 0;
  font-size: 0.875rem;
  transition: background 0.2s, color 0.2s;
}

.menu-item:hover {
  background: #eef4fb;
}

.menu-item.active {
  background: #e6f0fb;
  color: #2f6fd0;
  font-weight: 700;
}

.cluster-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e6eaf2;
  border-radius: 0.875rem;
  padding: 1rem;
  min-width: 0;
  min-height: 0;
  overflow-x: auto;
  overflow-y: auto;
}

.btn {
  padding: 0.5rem 0.75rem;
  border: 0;
  border-radius: 0.625rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
}

.btn-mini {
  background: #e2e8f0;
}

.btn-mini:hover {
  background: #cbd5e1;
}

/* ===== 响应式断点 ===== */
/* 大屏幕：侧边栏更宽 */
@media (min-width: 1920px) {
  .cluster-menu {
    width: 16rem;
  }
}

/* 中等屏幕：侧边栏收窄 */
@media (max-width: 1440px) {
  .cluster-menu {
    width: 13rem;
  }
}

/* 小屏幕：侧边栏更窄 */
@media (max-width: 1200px) {
  .cluster-menu {
    width: 11rem;
  }
  
  .menu-item {
    padding: 0.5rem;
    font-size: 0.8125rem;
  }
}

/* 平板：垂直布局 */
@media (max-width: 768px) {
  .cluster-main {
    flex-direction: column;
  }

  .cluster-menu {
    width: 100%;
    max-height: 12rem;
  }
  
  .cluster-bar {
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  
  .hint {
    display: none;
  }
}
</style>

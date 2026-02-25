<!-- src/layouts/ClusterLayout.vue -->
<template>
  <div class="cluster-layout">
    <div class="cluster-bar">
      <div class="left">
        <button class="btn btn-mini" @click="backToClusters">← 返回集群列表</button>
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
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/nodes`) }"
          @click="go('nodes')"
        >🖥️ 节点</a>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/namespaces`) }"
          @click="go('namespaces')"
        >📦 命名空间</a>

        <div class="menu-title">工作负载</div>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/workloads/pods`) }"
          @click="go('workloads/pods')"
        >🧬 Pods</a>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/workloads/deployments`) }"
          @click="go('workloads/deployments')"
        >📌 Deployments</a>

        <!-- ✅ 新增 -->
        <a class="menu-item"
           :class="{ active: isActive(`/c/${clusterId}/workloads/statefulsets`) }"
           @click="go('workloads/statefulsets')"
        >📦 StatefulSets</a>

        <a class="menu-item"
           :class="{ active: isActive(`/c/${clusterId}/workloads/daemonsets`) }"
           @click="go('workloads/daemonsets')"
        >🛰️ DaemonSets</a>

        <a class="menu-item"
           :class="{ active: isActive(`/c/${clusterId}/workloads/jobs`) }"
           @click="go('workloads/jobs')"
        >🧰 Jobs</a>

        <a class="menu-item"
           :class="{ active: isActive(`/c/${clusterId}/workloads/cronjobs`) }"
           @click="go('workloads/cronjobs')"
        >⏰ CronJobs</a>

        <div class="menu-title">配置</div>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/config/configmaps`) }"
          @click="go('config/configmaps')"
        >🧾 ConfigMaps</a>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/config/secrets`) }"
          @click="go('config/secrets')"
        >🔐 Secrets</a>
        <div class="menu-title">存储</div>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/storage/storageclasses`) }"
          @click="go('storage/storageclasses')"
        >💾 StorageClasses</a>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/storage/persistentvolumes`) }"
          @click="go('storage/persistentvolumes')"
        >💾 PV</a>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/storage/persistentvolumeclaims`) }"
          @click="go('storage/persistentvolumeclaims')"
        >💾 PVC</a>

        <div class="menu-title">网络</div>
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/networking/services`) }"
          @click="go('networking/services')"
        >🌐 Services</a>

        <!-- ✅ 关键修复：路由里是 networking/ingresses（复数） -->
        <a
          class="menu-item"
          :class="{ active: isActive(`/c/${clusterId}/networking/ingresses`) }"
          @click="go('networking/ingresses')"
        >🚪 Ingress</a>
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

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()

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
  background: #f6f7fb;
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
  background: #f1f5ff;
}

.menu-item.active {
  background: #e8efff;
  color: #326ce5;
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

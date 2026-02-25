<template>
  <div class="app-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 100 100">
            <g fill="#326ce5">
              <path
                d="M39.971 5.05c-3.607-.418-7.152.532-9.957 2.694l-1.315 1.035 1.703 2.17 1.227-1.006c1.946-1.606 4.387-2.537 6.984-2.537 3.97 0 7.513 2.053 9.616 5.298l1.127 1.629 2.012-1.388-1.217-1.772c-2.563-3.722-6.534-6.13-10.984-6.23z"
              />
              <path
                d="M42.026 94.796c3.608.418 7.153-.53 9.958-2.693l1.315-1.035-1.702-2.17-1.228 1.006c-1.947 1.606-4.388 2.537-6.985 2.537-3.969 0-7.512-2.053-9.615-5.298l-1.127-1.63-2.012 1.388 1.217 1.772c2.564 3.723 6.535 6.13 10.984 6.23z"
              />
              <path
                d="M5.136 42.085c.416 3.606-.532 7.152-2.694 9.957l-1.035 1.315 2.17 1.703 1.006-1.227c1.606-1.946 2.537-4.387 2.537-6.984 0-3.97-2.053-7.513-5.298-9.616l-1.629-1.127 1.388-2.012 1.772 1.217c3.722 2.563 6.13 6.534 6.23 10.984z"
              />
              <path
                d="M94.961 39.912c-.418-3.608.53-7.153 2.693-9.958l1.035-1.315-2.17-1.702-1.006 1.228c-1.606 1.947-2.537 4.388-2.537 6.985 0 3.969 2.053 7.512 5.298 9.615l1.63 1.127-1.388 2.012-1.772-1.217c-3.723-2.564-6.13-6.535-6.23-10.984z"
              />
            </g>
          </svg>
          <span>Kubernetes Admin</span>
        </div>
      </div>

      <!-- 导航菜单 -->
      <nav class="sidebar-nav">
        <div v-for="(group, index) in menuGroups" :key="index" class="menu-group">
          <div class="group-header" @click="group.path ? router.push(group.path) : toggleGroupCollapse(index)">
            <span class="collapse-icon">
              <span class="arrow" :class="{ expanded: !group.collapsed }"></span>
            </span>
            <span class="group-icon">{{ group.icon }}</span>
            <span class="group-name">{{ group.name }}</span>
          </div>

          <div v-if="group.items && group.items.length > 0"
               :class="['group-content', { collapsed: group.collapsed }]">
            <router-link
              v-for="(item, itemIndex) in group.items"
              :key="itemIndex"
              :to="item.path"
              class="nav-item"
              active-class="nav-item-active"
            >
              <span class="nav-icon">📄</span>
              <span class="nav-text">{{ item.label }}</span>
            </router-link>
          </div>
        </div>
      </nav>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content">
      <header class="top-nav">
        <div class="nav-left">
          <button class="menu-toggle" @click="toggleSidebar">
            <span class="toggle-icon">☰</span>
          </button>
        </div>
        <div class="nav-right">
          <div class="user-info">
            <span class="username">{{ username }}</span>
            <button @click="handleLogout" class="logout-btn">退出登录</button>
          </div>
        </div>
      </header>

      <div class="page-content">
        <router-view/>
      </div>
    </main>
  </div>
</template>

<script setup>
import {computed, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {logout} from '@/api/auth'

const router = useRouter()
const route = useRoute()

const sidebarCollapsed = ref(false)

const username = computed(() => {
  const userStr = localStorage.getItem('user') || sessionStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user.username || 'Admin'
    } catch (e) {
      return 'Admin'
    }
  }
  return 'Admin'
})

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 退出登录：后端 logout + 前端清理
const clearLocalAuth = () => {
  localStorage.removeItem('token')
  sessionStorage.removeItem('token')
  localStorage.removeItem('user')
  sessionStorage.removeItem('user')
}

const handleLogout = async () => {
  try {
    await logout()
  } catch (e) {
    console.error('logout failed', e)
  } finally {
    clearLocalAuth()
    router.replace('/login')
  }
}

/**
 * ✅ 方案 A：Layout（全局侧边栏）只放“平台级菜单”
 * - dashboard / clusters / users / cicd / images / environments
 * - 集群内部（nodes/pods/networking/config...）交给 ClusterLayout.vue
 */
const menuGroups = ref([
  // ✅ 顶层：首页（单项也行）
  {
    name: '首页',
    icon: '🏠',
    count: 0,
    collapsed: false,
    match: ['/dashboard'],
    path: '/dashboard', // ✅ 关键：点组头就跳转
  },
  {
    name: '平台',
    icon: '🏷️',
    count: 5,
    collapsed: true,
    match: ['/dashboard', '/clusters', '/platform'],
    items: [
      { path: '/clusters', label: '集群列表' },
      { path: '/platform/health', label: '平台健康' },
      { path: '/platform/appstore', label: '应用商城' },
      { path: '/platform/settings', label: '系统设置' },
    ],
  },
  {
    name: '安全',
    icon: '🛡️',
    count: 2,
    collapsed: true,
    match: ['/users', '/security'],
    items: [
      { path: '/users', label: '用户列表' },
      { path: '/security/audit', label: '审计日志' },
      { path: '/security/rbac/serviceaccounts', label: 'ServiceAccount 管理' },
      { path: '/security/rbac/roles', label: 'Role 管理' },
      { path: '/security/rbac/rolebindings', label: 'RoleBinding 管理' },
      { path: '/security/rbac/permission-check', label: '权限校验工具' },
    ],
  },
  {
    name: 'CI/CD',
    icon: '⚡',
    count: 4,
    collapsed: true,
    match: ['/cicd'],
    items: [
      { path: '/cicd/pipelines', label: '流水线管理' },
      { path: '/cicd/templates', label: '流水线模板管理' },
      // 你如果还有 /cicd/pipelines/create /:id 这些一般不放菜单
    ],
  },
  {
    name: '镜像与环境',
    icon: '📦',
    count: 2,
    collapsed: true,
    match: ['/images', '/environments'],
    items: [
      { path: '/images/repositories', label: '镜像仓库管理' },
      { path: '/environments', label: 'K8s环境管理' },
    ],
  },
])


const toggleGroupCollapse = (groupIndex) => {
  menuGroups.value[groupIndex].collapsed = !menuGroups.value[groupIndex].collapsed
}

// ✅ 自动展开：根据当前路由展开对应分组，其它分组折叠
const syncMenuWithRoute = () => {
  const currentPath = route.path
  menuGroups.value.forEach((group) => {
    if (!group.match || group.match.length === 0) return
    const hit = group.match.some((prefix) => currentPath.startsWith(prefix))
    group.collapsed = !hit
  })
}

syncMenuWithRoute()

watch(
  () => route.path,
  () => syncMenuWithRoute()
)
</script>

<style scoped>
/* ===== 主布局 ===== */
.app-layout {
  display: flex;
  height: 100vh;
  background-color: #f0f2f5;
}

/* ===== 侧边栏 ===== */
.sidebar {
  width: 15rem; /* 240px → 15rem */
  background-color: #2d3748;
  color: #ffffff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.sidebar.collapsed {
  width: 4rem; /* 64px → 4rem */
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid #4a5568;
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo svg {
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
}

.logo span {
  font-size: 1.125rem;
  font-weight: 600;
  color: #326ce5;
  white-space: nowrap;
}

.sidebar-nav {
  flex: 1;
  padding: 0.75rem 0;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.menu-group {
  margin-bottom: 0.375rem;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: #e2e8f0;
  background-color: #2d3748;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
  font-size: 0.875rem;
  border-left: 3px solid transparent;
}

.group-header:hover {
  background-color: #4a5568;
  border-left-color: #326ce5;
  padding-left: 1.25rem;
}

.collapse-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.125rem;
  height: 1.125rem;
  flex-shrink: 0;
}

.arrow {
  position: relative;
  width: 1.125rem;
  height: 1.125rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.arrow::before {
  content: '';
  position: absolute;
  width: 0.375rem;
  height: 0.375rem;
  border-right: 2px solid #a0aec0;
  border-bottom: 2px solid #a0aec0;
  transform: rotate(-135deg);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.arrow.expanded {
  transform: rotate(180deg);
}

.arrow.expanded::before {
  border-color: #326ce5;
  transform: rotate(45deg);
}

.group-content {
  padding-left: 0.75rem;
  overflow: hidden;
  transition: max-height 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  max-height: 500px;
  opacity: 1;
}

.group-content.collapsed {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem 0.625rem 2.5rem;
  color: #e2e8f0;
  text-decoration: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 0.8125rem;
  border-left: 2px solid transparent;
  position: relative;
  overflow: hidden;
}

.nav-item:hover {
  background-color: #4a5568;
  color: #ffffff;
  padding-left: 2.75rem;
}

.nav-item-active {
  background-color: #326ce5;
  color: #ffffff;
  border-left-color: #ffffff;
  box-shadow: 2px 0 8px rgba(50, 108, 229, 0.3);
}

/* ===== 主内容区 ===== */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.top-nav {
  height: 3.5rem; /* 56px */
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  flex-shrink: 0;
}

.menu-toggle {
  background: none;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
  color: #4a5568;
  padding: 0.625rem;
  border-radius: 0.375rem;
  transition: background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.menu-toggle:hover {
  background-color: #f7fafc;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.username {
  font-size: 0.875rem;
  color: #4a5568;
}

.logout-btn {
  padding: 0.5rem 1rem;
  background-color: #326ce5;
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logout-btn:hover {
  background-color: #2558c9;
}

.page-content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  padding: clamp(1rem, 3vw, 1.5rem);
}

/* ===== 响应式断点 ===== */
/* 大屏幕 */
@media (min-width: 1920px) {
  .sidebar {
    width: 16rem;
  }
}

/* 中等屏幕 */
@media (max-width: 1440px) {
  .sidebar {
    width: 14rem;
  }
}

/* 小屏幕 */
@media (max-width: 1200px) {
  .sidebar {
    width: 12rem;
  }
  
  .nav-item {
    padding-left: 2rem;
  }
  
  .nav-item:hover {
    padding-left: 2.25rem;
  }
}

/* 平板 */
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 1000;
    width: 15rem;
    transform: translateX(0);
    transition: transform 0.3s ease;
  }
  
  .sidebar.collapsed {
    transform: translateX(-100%);
    width: 15rem;
  }
  
  .logo span {
    font-size: 1rem;
  }
  
  .top-nav {
    height: 3rem;
    padding: 0 1rem;
  }
}
</style>

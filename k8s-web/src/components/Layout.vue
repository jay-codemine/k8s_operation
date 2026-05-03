<template>
  <div class="app-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo">
          <div class="logo-icon">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
              <defs>
                <linearGradient id="k8s-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" style="stop-color:#326CE5"/>
                  <stop offset="100%" style="stop-color:#54A3FF"/>
                </linearGradient>
              </defs>
              <circle cx="16" cy="16" r="15" fill="url(#k8s-gradient)"/>
              <g fill="#fff" transform="translate(6,6) scale(0.625)">
                <polygon points="16,0 20,12 32,12 22,20 26,32 16,24 6,32 10,20 0,12 12,12"/>
              </g>
            </svg>
          </div>
          <div class="logo-text">
            <span class="logo-title">K8s Platform</span>
            <span class="logo-version">v2.0</span>
          </div>
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

      <!-- ✅ 底部固定区域（大厂风格） -->
      <div class="sidebar-footer">
        <div class="footer-divider"></div>
        
        <!-- 应用商城 -->
        <router-link to="/platform/appstore" class="footer-item" active-class="footer-item-active">
          <div class="footer-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="1"/>
              <rect x="14" y="3" width="7" height="7" rx="1"/>
              <rect x="3" y="14" width="7" height="7" rx="1"/>
              <rect x="14" y="14" width="7" height="7" rx="1"/>
            </svg>
          </div>
          <span class="footer-text">应用商城</span>
        </router-link>

        <!-- 系统设置 -->
        <router-link to="/platform/settings" class="footer-item" active-class="footer-item-active">
          <div class="footer-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
          </div>
          <span class="footer-text">系统设置</span>
          <span class="footer-badge">NEW</span>
        </router-link>

        <!-- 帮助中心 -->
        <div class="footer-item" @click="showHelp = true">
          <div class="footer-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
              <line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
          </div>
          <span class="footer-text">帮助中心</span>
        </div>

        <!-- 用户信息卡片 -->
        <div class="user-card">
          <div class="user-avatar">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
            </svg>
          </div>
          <div class="user-details">
            <span class="user-name">{{ username }}</span>
            <span class="user-role">{{ userRoleDisplay }}</span>
          </div>
          <button class="logout-icon" @click="handleLogout" title="退出登录">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16 17 21 12 16 7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
          </button>
        </div>
      </div>
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
          <!-- 简化的操作区域 -->
          <div class="nav-actions">
            <button class="nav-action-btn" title="通知">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
                <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
              </svg>
              <span class="notification-badge">3</span>
            </button>
            <button class="nav-action-btn" title="搜索">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="8"/>
                <line x1="21" y1="21" x2="16.65" y2="16.65"/>
              </svg>
            </button>
          </div>
        </div>
      </header>

      <div class="page-content">
        <router-view/>
      </div>
    </main>

    <!-- AI 助手悬浮组件 -->
    <AiAssistant />
  </div>
</template>

<script setup>
import {computed, ref, reactive, watch, onMounted} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {logout} from '@/api/auth'
import permissionStore from '@/stores/permission'
import AiAssistant from '@/components/AiAssistant.vue'

const router = useRouter()
const route = useRoute()

const sidebarCollapsed = ref(false)
const showHelp = ref(false)

// 用户名
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

// 用户角色显示
const userRoleDisplay = computed(() => {
  if (permissionStore.state.isSuperAdmin) return '超级管理员'
  if (permissionStore.isAdmin.value) return '平台管理员'
  if (permissionStore.isClusterAdmin.value) return '集群管理员'
  if (permissionStore.isDeveloper.value) return '开发者'
  return '普通用户'
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
  permissionStore.clearPermissions() // 清除权限缓存
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
 * 菜单权限配置
 * 角色分类:
 *   - super_admin: 超级管理员，全部权限
 *   - platform_admin: 平台管理员
 *   - cluster_admin: 集群管理员
 *   - cicd_admin: CI/CD 管理员
 *   - developer: 开发人员
 *   - viewer: 只读用户
 */
const menuPermissions = {
  // ==================== 平台管理 ====================
  '/clusters': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin', 'developer', 'viewer'],
  '/platform/health': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/platform/settings': ['super_admin', 'platform_admin'],
  '/platform/appstore': ['super_admin', 'platform_admin', 'cluster_admin'],
  
  // ==================== 用户与权限管理（精简后） ====================
  '/security/users': ['super_admin', 'platform_admin'],
  '/security/roles': ['super_admin', 'platform_admin'],
  '/security/authorization': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/diagnosis': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  
  // 兼容旧路径
  '/users': ['super_admin', 'platform_admin'],
  '/user-permissions': ['super_admin', 'platform_admin'],
  '/rbac': ['super_admin', 'platform_admin'],
  
  // ==================== 安全审计 ====================
  '/security/audit': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/ai-approvals': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  '/security/rbac/serviceaccounts': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/roles': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/rolebindings': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/permission-check': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  
  // ==================== CI/CD 流水线 ====================
  '/cicd/pipelines': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/releases': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/templates': ['super_admin', 'platform_admin'],
  '/cicd/approvals': ['super_admin', 'platform_admin', 'cicd_admin'],
  
  // ==================== 镜像管理 ====================
  '/images/repositories': ['super_admin', 'platform_admin'],
  '/images/browse': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'viewer'],
  '/images/cleanup': ['super_admin', 'platform_admin'],
  
}

/**
 * 检查菜单项是否可见
 * 大厂风格：无权限即隐藏
 */
const isMenuVisible = (path) => {
  // 超级管理员可以看到所有菜单
  if (permissionStore.state.isSuperAdmin) return true
  
  // 获取菜单权限配置
  const roles = menuPermissions[path]
  if (!roles) return true // 未配置则默认可见
  
  // 检查用户角色（平台角色 + 集群权限角色）
  const userRoles = permissionStore.roleTypes.value || []
  
  // 没有任何角色时，只允许访问基础功能
  if (userRoles.length === 0) {
    // 基础功能：首页、权限诊断
    const basicPaths = ['/dashboard', '/security/diagnosis']
    return basicPaths.includes(path)
  }
  
  return roles.some(role => userRoles.includes(role))
}

/**
 * 过滤菜单项，只显示有权限的
 */
const filterMenuItems = (items) => {
  if (!items) return []
  return items.filter(item => isMenuVisible(item.path))
}

/**
 * 侧边栏菜单配置（响应式）
 */
const menuGroupsConfig = reactive([
  // 首页
  {
    name: '首页',
    icon: '🏠',
    count: 0,
    collapsed: false,
    match: ['/dashboard'],
    path: '/dashboard',
  },
  // 平台
  {
    name: '平台',
    icon: '🏷️',
    count: 2,
    collapsed: true,
    match: ['/dashboard', '/clusters', '/platform'],
    items: [
      { path: '/clusters', label: '集群管理' },
      { path: '/platform/health', label: '平台健康' },
    ],
  },
  // 安全（精简后的5个菜单，大厂风格）
  {
    name: '安全',
    icon: '🛡️',
    count: 5,
    collapsed: true,
    match: ['/users', '/security', '/rbac', '/user-permissions'],
    items: [
      { path: '/security/users', label: '用户管理' },
      { path: '/security/roles', label: '角色管理' },
      { path: '/security/authorization', label: '授权管理' },
      { path: '/security/audit', label: '审计日志' },
      { path: '/security/ai-approvals', label: 'AI 审批管理' },
      { path: '/security/diagnosis', label: '权限诊断' },
    ],
  },
  // CI/CD
  {
    name: 'CI/CD',
    icon: '⚡',
    count: 4,
    collapsed: true,
    match: ['/cicd'],
    items: [
      { path: '/cicd/pipelines', label: '流水线管理' },
      { path: '/cicd/releases', label: '发布管理' },
      { path: '/cicd/approvals', label: '审批管理' },
      { path: '/cicd/artifacts', label: '制品管理' },
      { path: '/cicd/templates', label: '流水线模板' },
      { path: '/cicd/agents', label: '构建探针' },
    ],
  },
  // 镜像管理
  {
    name: '镜像管理',
    icon: '📦',
    count: 3,
    collapsed: true,
    match: ['/images'],
    items: [
      { path: '/images/repositories', label: '镜像仓库管理' },
      { path: '/images/browse', label: '镜像浏览' },
      { path: '/images/cleanup', label: '清理策略' },
    ],
  },
])

// 动态计算可见菜单
const menuGroups = computed(() => {
  return menuGroupsConfig.map(group => {
    const visibleItems = filterMenuItems(group.items)
    return {
      ...group,
      items: visibleItems,
      count: visibleItems.length,
      // 如果所有子菜单都不可见，则隐藏整个分组
      visible: !group.items || visibleItems.length > 0 || group.path
    }
  }).filter(group => group.visible)
})

const toggleGroupCollapse = (groupIndex) => {
  // 需要找到原始配置的索引
  const visibleGroups = menuGroups.value
  if (groupIndex < visibleGroups.length) {
    const groupName = visibleGroups[groupIndex].name
    const configIndex = menuGroupsConfig.findIndex(g => g.name === groupName)
    if (configIndex >= 0) {
      menuGroupsConfig[configIndex].collapsed = !menuGroupsConfig[configIndex].collapsed
    }
  }
}

// 自动展开：根据当前路由展开对应分组
const syncMenuWithRoute = () => {
  const currentPath = route.path
  menuGroupsConfig.forEach((group) => {
    if (!group.match || group.match.length === 0) return
    const hit = group.match.some((prefix) => currentPath.startsWith(prefix))
    group.collapsed = !hit
  })
}

// 加载用户权限
onMounted(async () => {
  try {
    await permissionStore.loadPermissions()
  } catch (e) {
    console.error('加载权限失败', e)
  }
})

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
  background-color: var(--bg-secondary, #f9fafb);
}

/* ===== 侧边栏 - 现代紫蓝渐变 ===== */
.sidebar {
  width: 15rem;
  background: linear-gradient(180deg, #1e1b4b 0%, #312e81 50%, #3730a3 100%);
  color: #ffffff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  box-shadow: 4px 0 20px rgba(30, 27, 75, 0.3);
}

.sidebar.collapsed {
  width: 4rem;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
}

.logo-icon svg {
  width: 100%;
  height: 100%;
}

.logo-text {
  display: flex;
  flex-direction: column;
}

.logo-title {
  font-size: 1rem;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
}

.logo-version {
  font-size: 0.65rem;
  color: #a5b4fc;
  font-weight: 500;
  margin-top: 2px;
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
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
  font-size: 0.875rem;
  border-left: 3px solid transparent;
}

.group-header:hover {
  background-color: rgba(255, 255, 255, 0.1);
  border-left-color: #818cf8;
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
  border-color: #818cf8;
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
  background-color: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  padding-left: 2.75rem;
}

.nav-item-active {
  background: linear-gradient(90deg, #6366f1 0%, #8b5cf6 100%);
  color: #ffffff;
  border-left-color: #ffffff;
  box-shadow: 2px 0 12px rgba(99, 102, 241, 0.4);
}

/* ===== 底部固定区域（大厂风格） ===== */
.sidebar-footer {
  margin-top: auto;
  padding: 0.5rem 0.75rem 1rem;
}

.footer-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  margin-bottom: 0.75rem;
}

.footer-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  color: #a0aec0;
  text-decoration: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 0.25rem;
}

.footer-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.footer-item-active {
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.3), rgba(139, 92, 246, 0.1));
  color: #a5b4fc;
}

.footer-icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

.footer-icon svg {
  width: 100%;
  height: 100%;
}

.footer-text {
  font-size: 0.8125rem;
  font-weight: 500;
  flex: 1;
}

.footer-badge {
  padding: 0.125rem 0.375rem;
  background: linear-gradient(135deg, #6366f1 0%, #ec4899 100%);
  color: #fff;
  font-size: 0.625rem;
  font-weight: 600;
  border-radius: 0.25rem;
  text-transform: uppercase;
}

/* 用户信息卡片 */
.user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  margin-top: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.user-avatar {
  width: 2.25rem;
  height: 2.25rem;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-avatar svg {
  width: 1.25rem;
  height: 1.25rem;
  color: #ffffff;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #ffffff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  display: block;
  font-size: 0.6875rem;
  color: #a0aec0;
  margin-top: 1px;
}

.logout-icon {
  width: 2rem;
  height: 2rem;
  padding: 0.375rem;
  background: transparent;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  color: #a0aec0;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.logout-icon:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.logout-icon svg {
  width: 100%;
  height: 100%;
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

.nav-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-action-btn {
  position: relative;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0.5rem;
  background: transparent;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s ease;
}

.nav-action-btn:hover {
  background: #f1f5f9;
  color: #334155;
}

.nav-action-btn svg {
  width: 100%;
  height: 100%;
}

.notification-badge {
  position: absolute;
  top: 0.125rem;
  right: 0.125rem;
  min-width: 1rem;
  height: 1rem;
  padding: 0 0.25rem;
  background: linear-gradient(135deg, #ef4444 0%, #ec4899 100%);
  color: #fff;
  font-size: 0.625rem;
  font-weight: 600;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
  padding: 16px 20px;
  background: #f8fafc;
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

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
            <span class="logo-title">K8sOperation</span>
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
            <span class="group-icon" v-html="getIconSvg(group.icon, 18)"></span>
            <span class="group-name">{{ group.name }}</span>
          </div>

          <div v-if="group.items && group.items.length > 0"
               :class="['group-content', { collapsed: group.collapsed }]">
            <template v-for="(item, itemIndex) in group.items" :key="itemIndex">
              <!-- 分组标题（不可点击） -->
              <div v-if="item.section" class="nav-section-label">
                <span class="nav-section-text">{{ item.section }}</span>
              </div>
              <!-- 常规导航项 -->
              <router-link
                v-else
                :to="item.path"
                class="nav-item"
                active-class="nav-item-active"
              >
                <span class="nav-icon" v-html="getIconSvg(item.icon || 'file', 16)"></span>
                <span class="nav-text">{{ item.label }}</span>
              </router-link>
            </template>
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
        <router-view v-slot="{ Component, route: viewRoute }">
          <transition name="page-fade" mode="default">
            <component :is="Component" :key="viewRoute.fullPath" />
          </transition>
        </router-view>
      </div>
    </main>

    <!-- AI 助手悬浮组件 -->
    <AiAssistant />
  </div>
</template>

<script setup>
import {computed, ref, reactive, watch, onMounted, onUnmounted} from 'vue'

// SVG 图标映射（大厂线条风格 24x24）
const iconPaths = {
  'home': '<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>',
  'cluster': '<circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="6" y2="12"/><line x1="18" y1="12" x2="22" y2="12"/>',
  'heart': '<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>',
  'robot': '<rect x="3" y="11" width="18" height="10" rx="2"/><circle cx="12" cy="5" r="2"/><path d="M12 7v4"/><line x1="8" y1="16" x2="8.01" y2="16"/><line x1="16" y1="16" x2="16.01" y2="16"/>',
  'rocket': '<path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09zM12 15l-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"/><path d="M9 12H4l3 3"/><path d="M15 12h5l-3 3"/>',
  'apps': '<rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/>',
  'send': '<line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>',
  'branch': '<line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/>',
  'sync': '<polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>',
  'history': '<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>',
  'archive': '<polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/>',
  'upload': '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>',
  'calendar': '<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>',
  'check-circle': '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>',
  'file': '<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/>',
  'desktop': '<rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>',
  'search': '<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>',
  'delete': '<polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/>',
  'dashboard': '<rect x="3" y="3" width="7" height="9"/><rect x="14" y="3" width="7" height="5"/><rect x="14" y="12" width="7" height="9"/><rect x="3" y="16" width="7" height="5"/>',
  'storage': '<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>',
  'bell': '<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>',
  'exclamation-circle': '<circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>',
  'command': '<path d="M18 3a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3 3 3 0 0 0 3-3 3 3 0 0 0-3-3H6a3 3 0 0 0-3 3 3 3 0 0 0 3 3 3 3 0 0 0 3-3V6a3 3 0 0 0-3-3 3 3 0 0 0-3 3 3 3 0 0 0 3 3h12a3 3 0 0 0 3-3 3 3 0 0 0-3-3z"/>',
  'code-block': '<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>',
  'code': '<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>',
  'settings': '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
  'user-group': '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  'safe': '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>',
  'lock': '<rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>',
  'user': '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
}

function getIconSvg(name, size = 16) {
  const path = iconPaths[name] || iconPaths['file']
  return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${path}</svg>`
}
import {useRoute, useRouter} from 'vue-router'
import {Message} from '@arco-design/web-vue'
import {logout} from '@/api/auth'
import permissionStore from '@/stores/permission'
import AiAssistant from '@/components/AiAssistant.vue'

const router = useRouter()
const route = useRoute()

const isMobileViewport = () => window.matchMedia('(max-width: 768px)').matches
const sidebarCollapsed = ref(isMobileViewport())
const showHelp = ref(false)

// 用户名
const username = computed(() => {
  const userStr = localStorage.getItem('user') || sessionStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user.username || 'Admin'
    } catch {
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
    Message.success({ content: '退出成功', duration: 1500 })
  } catch (e) {
    // 即使后端失败，前端也正常退出（JWT 无法服务端失效）
    console.warn('logout api failed, clearing local auth anyway', e)
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
  '/platform/aiops': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin'],
  
  // ==================== 平台管理（IAM 统一收口） ====================
  '/admin/users': ['super_admin', 'platform_admin'],
  '/admin/roles': ['super_admin', 'platform_admin'],
  '/admin/identity': ['super_admin', 'platform_admin'],
  '/admin/approvals': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  '/admin/audit': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/admin/service-accounts': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/admin/settings': ['super_admin', 'platform_admin'],
  
  // 兼容旧路径
  '/security/users': ['super_admin', 'platform_admin'],
  '/security/roles': ['super_admin', 'platform_admin'],
  '/security/authorization': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/ldap': ['super_admin', 'platform_admin'],
  '/security/diagnosis': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  '/users': ['super_admin', 'platform_admin'],
  '/user-permissions': ['super_admin', 'platform_admin'],
  '/rbac': ['super_admin', 'platform_admin'],
  '/security/audit': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/security/ai-approvals': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  '/security/rbac/serviceaccounts': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/roles': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/rolebindings': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/security/rbac/permission-check': ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'viewer'],
  
  // ==================== CI/CD 流水线 ====================
  '/cicd/apps': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'tester', 'viewer'],
  '/cicd/quick-onboard': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/pipelines': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/gitops': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/build-records': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'tester', 'viewer'],
  '/cicd/artifacts': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'tester', 'viewer'],
  '/cicd/releases': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/release-history': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'tester', 'viewer'],
  '/cicd/approvals': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer'],
  '/cicd/approval-policy': ['super_admin', 'platform_admin'],
  '/cicd/templates': ['super_admin', 'platform_admin'],
  '/cicd/agents': ['super_admin', 'platform_admin', 'cicd_admin'],
  
  // ==================== 监控中心 ====================
  '/monitoring': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin', 'developer', 'viewer'],
  '/monitoring/datasources': ['super_admin', 'platform_admin', 'cluster_admin'],
  '/monitoring/alert-rules': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin'],
  '/monitoring/alert-events': ['super_admin', 'platform_admin', 'cluster_admin', 'cicd_admin', 'developer', 'viewer'],
  
  // ==================== 镜像管理 ====================
  '/images/repositories': ['super_admin', 'platform_admin'],
  '/images/browse': ['super_admin', 'platform_admin', 'cicd_admin', 'cluster_admin', 'developer', 'viewer'],
  '/images/cleanup': ['super_admin', 'platform_admin'],
  
  // ==================== CRD 扩展 ====================
  '/extensions/crd': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/extensions/cr-instances': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  '/extensions/yaml-workbench': ['super_admin', 'platform_admin', 'cluster_admin', 'developer'],
  
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
    icon: 'home',
    collapsed: false,
    match: ['/dashboard'],
    path: '/dashboard',
  },
  // 资源中心
  {
    name: '资源中心',
    icon: 'cluster',
    count: 3,
    collapsed: true,
    match: ['/dashboard', '/clusters', '/platform/health', '/platform/aiops'],
    items: [
      { path: '/clusters', label: '集群管理', icon: 'cluster' },
      { path: '/platform/health', label: '平台健康', icon: 'heart' },
      { path: '/platform/aiops', label: '智能运维', icon: 'robot' },
    ],
  },
  // CI/CD
  {
    name: 'CI/CD',
    icon: 'rocket',
    count: 12,
    collapsed: true,
    match: ['/cicd'],
    items: [
      { section: '应用中心' },
      { path: '/cicd/apps', label: '应用总览', icon: 'apps' },
      { path: '/cicd/quick-onboard', label: '快速接入', icon: 'send' },
      { section: '流水线 · CI' },
      { path: '/cicd/pipelines', label: '流水线管理', icon: 'branch' },
      { path: '/cicd/gitops/releases', label: 'GitOps 发布', icon: 'sync' },
      { path: '/cicd/build-records', label: '构建记录', icon: 'history' },
      { path: '/cicd/artifacts', label: '制品库', icon: 'archive' },
      { section: '发布中心 · CD' },
      { path: '/cicd/releases', label: '发布管理', icon: 'upload' },
      { path: '/cicd/release-history', label: '发布历史', icon: 'calendar' },
      { path: '/cicd/approvals', label: '审批工单', icon: 'check-circle' },
      { section: '规则中心' },
      { path: '/cicd/approval-policy', label: '审批策略', icon: 'file' },
      { path: '/cicd/templates', label: '流水线模板', icon: 'file' },
      { path: '/cicd/agents', label: '构建节点', icon: 'desktop' },
    ],
  },
  // 镜像管理
  {
    name: '镜像管理',
    icon: 'archive',
    count: 3,
    collapsed: true,
    match: ['/images'],
    items: [
      { path: '/images/repositories', label: '镜像仓库管理', icon: 'archive' },
      { path: '/images/browse', label: '镜像浏览', icon: 'search' },
      { path: '/images/cleanup', label: '清理策略', icon: 'delete' },
    ],
  },
  // 监控中心
  {
    name: '监控中心',
    icon: 'dashboard',
    count: 4,
    collapsed: true,
    match: ['/monitoring'],
    items: [
      { path: '/monitoring', label: '监控总览', icon: 'dashboard' },
      { path: '/monitoring/datasources', label: '数据源管理', icon: 'storage' },
      { path: '/monitoring/alert-rules', label: '告警规则', icon: 'bell' },
      { path: '/monitoring/alert-events', label: '告警事件', icon: 'exclamation-circle' },
    ],
  },
  // CRD 扩展
  {
    name: 'CRD 扩展',
    icon: 'command',
    count: 3,
    collapsed: true,
    match: ['/extensions'],
    items: [
      { path: '/extensions/crd', label: 'CRD 资源管理', icon: 'command' },
      { path: '/extensions/cr-instances', label: 'CR 实例管理', icon: 'code-block' },
      { path: '/extensions/yaml-workbench', label: 'YAML 工作台', icon: 'code' },
    ],
  },
  // 平台管理
  {
    name: '平台管理',
    icon: 'settings',
    count: 7,
    collapsed: true,
    match: ['/admin'],
    items: [
      { path: '/admin/users', label: '用户与组织', icon: 'user-group' },
      { path: '/admin/roles', label: '权限与角色', icon: 'safe' },
      { path: '/admin/identity', label: '身份认证', icon: 'lock' },
      { path: '/admin/approvals', label: '审批中心', icon: 'check-circle' },
      { path: '/admin/audit', label: '审计中心', icon: 'search' },
      { path: '/admin/service-accounts', label: '服务账号', icon: 'user' },
      { path: '/admin/settings', label: '系统设置', icon: 'settings' },
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
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

const handleResize = () => {
  if (isMobileViewport()) {
    sidebarCollapsed.value = true
  }
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
  background-color: #f2f3f5;
}

/* ===== 侧边栏 - 字节/Arco Pro 深色中性风格 ===== */
.sidebar {
  width: 15rem;
  background: #1d2129;
  color: #ffffff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.08);
}

.sidebar.collapsed {
  width: 4rem;
}

.sidebar.collapsed .logo-text,
.sidebar.collapsed .group-name,
.sidebar.collapsed .nav-text,
.sidebar.collapsed .footer-text,
.sidebar.collapsed .footer-badge,
.sidebar.collapsed .user-details,
.sidebar.collapsed .logout-icon {
  display: none;
}

.sidebar.collapsed .group-header,
.sidebar.collapsed .nav-item,
.sidebar.collapsed .footer-item {
  justify-content: center;
  padding-left: 0;
  padding-right: 0;
}

.sidebar.collapsed .group-header:hover,
.sidebar.collapsed .nav-item:hover {
  padding-left: 0;
}

.sidebar.collapsed .group-content {
  padding-left: 0;
}

.sidebar.collapsed .nav-item {
  padding: 0.625rem 0;
}

.sidebar.collapsed .user-card {
  justify-content: center;
  padding: 0.5rem;
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
  color: #86909c;
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
  background-color: rgba(255, 255, 255, 0.08);
  border-left-color: #4080FF;
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
  border-color: #4080FF;
  transform: rotate(45deg);
}

.group-content {
  padding-left: 0.75rem;
  overflow: hidden;
  transition: max-height 0.3s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  max-height: 800px;
  opacity: 1;
}

.group-content.collapsed {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

/* 分组标签（企业级导航分组头 - 大厂风格） */
.nav-section-label {
  padding: 0.75rem 1rem 0.375rem 1.5rem;
  margin-top: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-section-label::before {
  content: '';
  width: 3px;
  height: 12px;
  background: linear-gradient(135deg, #165DFF, #14C9C9);
  border-radius: 2px;
  flex-shrink: 0;
}

.nav-section-text {
  font-size: 0.6875rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.06em;
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
  background: rgba(22, 93, 255, 0.15);
  color: #ffffff;
  border-left-color: #165DFF;
  box-shadow: none;
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
  background: rgba(22, 93, 255, 0.12);
  color: #4080FF;
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
  background: linear-gradient(135deg, #165DFF 0%, #14C9C9 100%);
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
  background: linear-gradient(135deg, #165DFF 0%, #4080FF 100%);
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
  background: #f2f3f5;
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

/* 页面切换过渡动画（大厂风格柔和淡入） */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>

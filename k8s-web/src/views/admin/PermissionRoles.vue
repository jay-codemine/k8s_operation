<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">权限与角色</h1>
        <p class="page-desc">统一管理平台角色、权限策略、RBAC规则和权限诊断</p>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="tab-container">
      <div class="tab-nav">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['tab-btn', { active: activeTab === tab.key }]"
          @click="activeTab = tab.key"
        >
          <span class="tab-icon">{{ tab.icon }}</span>
          <span class="tab-label">{{ tab.label }}</span>
        </button>
      </div>

      <div class="tab-content">
        <!-- 角色管理 -->
        <RoleManagement v-if="activeTab === 'roles'" />
        <!-- 权限矩阵 -->
        <AuthorizationManagement v-if="activeTab === 'authorization'" />
        <!-- RBAC 管理 -->
        <div v-if="activeTab === 'rbac'" class="rbac-panel">
          <div class="rbac-nav">
            <button
              v-for="sub in rbacSubTabs"
              :key="sub.key"
              :class="['sub-tab-btn', { active: rbacActiveTab === sub.key }]"
              @click="rbacActiveTab = sub.key"
            >
              {{ sub.label }}
            </button>
          </div>
          <div class="rbac-content">
            <Roles v-if="rbacActiveTab === 'k8s-roles'" />
            <RoleBindings v-if="rbacActiveTab === 'bindings'" />
          </div>
        </div>
        <!-- 权限诊断 -->
        <PermissionCheck v-if="activeTab === 'diagnosis'" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import RoleManagement from '@/views/security/RoleManagement.vue'
import AuthorizationManagement from '@/views/security/AuthorizationManagement.vue'
import PermissionCheck from '@/views/security/rbac/PermissionCheck.vue'
import Roles from '@/views/security/rbac/Roles.vue'
import RoleBindings from '@/views/security/rbac/RoleBindings.vue'

const activeTab = ref('roles')
const rbacActiveTab = ref('k8s-roles')

const tabs = [
  { key: 'roles', label: '角色管理', icon: '👥' },
  { key: 'authorization', label: '权限矩阵', icon: '📊' },
  { key: 'rbac', label: 'RBAC 管理', icon: '🔐' },
  { key: 'diagnosis', label: '权限诊断', icon: '🔍' },
]

const rbacSubTabs = [
  { key: 'k8s-roles', label: 'K8s Roles' },
  { key: 'bindings', label: 'RoleBindings' },
]
</script>

<style scoped>
.admin-page {
  padding: 0;
}

.page-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 24px 32px;
  border-radius: 12px;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  margin: 0 0 6px;
}

.page-desc {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
}

.tab-container {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.tab-nav {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #e2e8f0;
  padding: 0 24px;
  background: #fafbfc;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 14px 20px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #334155;
  background: rgba(99, 102, 241, 0.04);
}

.tab-btn.active {
  color: #6366f1;
  border-bottom-color: #6366f1;
  background: rgba(99, 102, 241, 0.06);
}

.tab-icon {
  font-size: 16px;
}

.tab-content {
  padding: 0;
  min-height: 500px;
}

.rbac-panel {
  padding: 16px 24px;
}

.rbac-nav {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.sub-tab-btn {
  padding: 8px 16px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.sub-tab-btn:hover {
  background: #e2e8f0;
}

.sub-tab-btn.active {
  background: #6366f1;
  color: #fff;
  border-color: #6366f1;
}

.rbac-content {
  min-height: 400px;
}
</style>

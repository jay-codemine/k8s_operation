<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">审计中心</h1>
        <p class="page-desc">全面追踪平台操作记录，登录审计、权限变更、集群操作和发布记录</p>
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
        <!-- 全部日志（现有审计日志组件） -->
        <AuditLog v-if="activeTab === 'all'" />

        <!-- 登录审计 -->
        <div v-if="activeTab === 'login'" class="audit-category">
          <div class="category-header">
            <h3>登录审计</h3>
            <p>追踪用户登录/退出行为，异常登录检测</p>
          </div>
          <AuditLog :filter-module="'auth'" />
        </div>

        <!-- 权限变更 -->
        <div v-if="activeTab === 'permission'" class="audit-category">
          <div class="category-header">
            <h3>权限变更</h3>
            <p>追踪角色分配、权限修改等操作</p>
          </div>
          <AuditLog :filter-module="'rbac'" />
        </div>

        <!-- 集群操作 -->
        <div v-if="activeTab === 'cluster'" class="audit-category">
          <div class="category-header">
            <h3>集群操作</h3>
            <p>追踪 K8s 资源创建/修改/删除操作</p>
          </div>
          <AuditLog :filter-module="'cluster'" />
        </div>

        <!-- 发布记录 -->
        <div v-if="activeTab === 'release'" class="audit-category">
          <div class="category-header">
            <h3>发布记录</h3>
            <p>追踪 CI/CD 构建与发布操作记录</p>
          </div>
          <AuditLog :filter-module="'cicd'" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import AuditLog from '@/views/security/audit/AuditLog.vue'

const activeTab = ref('all')

const tabs = [
  { key: 'all', label: '全部日志', icon: '📋' },
  { key: 'login', label: '登录审计', icon: '🔐' },
  { key: 'permission', label: '权限变更', icon: '🛡️' },
  { key: 'cluster', label: '集群操作', icon: '☸️' },
  { key: 'release', label: '发布记录', icon: '🚀' },
]
</script>

<style scoped>
.admin-page {
  padding: 0;
}

.page-header {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
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

.audit-category {
  padding: 0;
}

.category-header {
  padding: 16px 24px;
  border-bottom: 1px solid #f1f5f9;
  background: #fafbfc;
}

.category-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 4px;
}

.category-header p {
  font-size: 12px;
  color: #64748b;
  margin: 0;
}
</style>

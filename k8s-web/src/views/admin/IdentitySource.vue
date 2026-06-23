<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">身份认证</h1>
        <p class="page-desc">管理企业身份源，支持 LDAP、AD 域、飞书、企业微信等多种认证方式</p>
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
          <span v-if="tab.status" :class="['status-badge', tab.status]">{{ tab.statusText }}</span>
        </button>
      </div>

      <div class="tab-content">
        <!-- LDAP/AD -->
        <LDAPSettings v-if="activeTab === 'ldap'" />

        <!-- 飞书 -->
        <div v-if="activeTab === 'feishu'" class="coming-soon">
          <div class="coming-icon">🐦</div>
          <h3>飞书登录</h3>
          <p>通过飞书开放平台实现企业 SSO 登录</p>
          <div class="feature-list">
            <div class="feature-item">✅ 飞书扫码登录</div>
            <div class="feature-item">✅ 自动同步组织架构</div>
            <div class="feature-item">✅ 审批消息推送</div>
          </div>
          <span class="badge-coming">即将支持</span>
        </div>

        <!-- 企业微信 -->
        <div v-if="activeTab === 'wecom'" class="coming-soon">
          <div class="coming-icon">💬</div>
          <h3>企业微信</h3>
          <p>通过企业微信实现统一身份认证</p>
          <div class="feature-list">
            <div class="feature-item">✅ 企微扫码登录</div>
            <div class="feature-item">✅ 通讯录同步</div>
            <div class="feature-item">✅ 消息通知</div>
          </div>
          <span class="badge-coming">即将支持</span>
        </div>

        <!-- OIDC/OAuth2 -->
        <div v-if="activeTab === 'oidc'" class="coming-soon">
          <div class="coming-icon">🔑</div>
          <h3>OIDC / OAuth2</h3>
          <p>标准 OpenID Connect 协议集成，支持接入 GitLab、GitHub、Keycloak 等</p>
          <div class="feature-list">
            <div class="feature-item">✅ GitLab OAuth</div>
            <div class="feature-item">✅ GitHub OAuth</div>
            <div class="feature-item">✅ Keycloak OIDC</div>
            <div class="feature-item">✅ 自定义 OAuth2 Provider</div>
          </div>
          <span class="badge-coming">即将支持</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import LDAPSettings from '@/views/security/LDAPSettings.vue'

const activeTab = ref('ldap')

const tabs = [
  { key: 'ldap', label: 'LDAP / AD 域', icon: '🏢', status: 'active', statusText: '已启用' },
  { key: 'feishu', label: '飞书', icon: '🐦', status: 'planned', statusText: '规划中' },
  { key: 'wecom', label: '企业微信', icon: '💬', status: 'planned', statusText: '规划中' },
  { key: 'oidc', label: 'OIDC / OAuth2', icon: '🔑', status: 'planned', statusText: '规划中' },
]
</script>

<style scoped>
.admin-page {
  padding: 0;
}

.page-header {
  background: linear-gradient(135deg, #0ea5e9 0%, #6366f1 100%);
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

.status-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.status-badge.active {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.planned {
  background: #fef3c7;
  color: #d97706;
}

.tab-content {
  padding: 0;
  min-height: 500px;
}

.coming-soon {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.coming-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.coming-soon h3 {
  font-size: 20px;
  color: #1e293b;
  margin: 0 0 8px;
}

.coming-soon p {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 20px;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 20px;
}

.feature-item {
  font-size: 13px;
  color: #475569;
}

.badge-coming {
  padding: 6px 16px;
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  color: #fff;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}
</style>

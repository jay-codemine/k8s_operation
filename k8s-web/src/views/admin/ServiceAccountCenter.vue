<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">服务账号</h1>
        <p class="page-desc">管理平台集成的服务账号，包括 Jenkins、GitLab、Harbor 和 API Token</p>
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
        <!-- K8s ServiceAccount -->
        <ServiceAccounts v-if="activeTab === 'k8s-sa'" />

        <!-- Jenkins -->
        <div v-if="activeTab === 'jenkins'" class="service-panel">
          <div class="service-header">
            <div class="service-info">
              <div class="service-icon jenkins">J</div>
              <div>
                <h3>Jenkins 服务账号</h3>
                <p>用于 CI/CD 流水线执行的 Jenkins 凭证管理</p>
              </div>
            </div>
            <span class="status-connected">已连接</span>
          </div>
          <div class="service-details">
            <div class="detail-item">
              <span class="detail-label">服务地址</span>
              <span class="detail-value">配置于系统设置 → Jenkins 集成</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">账号类型</span>
              <span class="detail-value">API Token</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">权限范围</span>
              <span class="detail-value">Pipeline 执行、Job 创建、构建触发</span>
            </div>
          </div>
        </div>

        <!-- Harbor -->
        <div v-if="activeTab === 'harbor'" class="service-panel">
          <div class="service-header">
            <div class="service-info">
              <div class="service-icon harbor">H</div>
              <div>
                <h3>Harbor Robot Account</h3>
                <p>用于镜像推送/拉取的 Harbor 机器人账号</p>
              </div>
            </div>
            <span class="status-connected">已连接</span>
          </div>
          <div class="service-details">
            <div class="detail-item">
              <span class="detail-label">服务地址</span>
              <span class="detail-value">配置于系统设置 → 镜像仓库</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">账号类型</span>
              <span class="detail-value">Robot Account</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">权限范围</span>
              <span class="detail-value">镜像 Push/Pull、Tag 管理</span>
            </div>
          </div>
        </div>

        <!-- API Token -->
        <div v-if="activeTab === 'api-token'" class="service-panel">
          <div class="service-header">
            <div class="service-info">
              <div class="service-icon api">🔑</div>
              <div>
                <h3>API Token 管理</h3>
                <p>管理平台 API 访问令牌，用于第三方系统集成</p>
              </div>
            </div>
          </div>
          <div class="token-list">
            <div class="empty-state">
              <p>暂无 API Token，点击下方按钮创建</p>
              <button class="btn-create">+ 创建 Token</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ServiceAccounts from '@/views/security/rbac/ServiceAccounts.vue'

const activeTab = ref('k8s-sa')

const tabs = [
  { key: 'k8s-sa', label: 'K8s ServiceAccount', icon: '☸️', status: 'active', statusText: '可用' },
  { key: 'jenkins', label: 'Jenkins', icon: '🔧', status: 'active', statusText: '已连接' },
  { key: 'harbor', label: 'Harbor', icon: '🐳', status: 'active', statusText: '已连接' },
  { key: 'api-token', label: 'API Token', icon: '🔑', status: null, statusText: '' },
]
</script>

<style scoped>
.admin-page {
  padding: 0;
}

.page-header {
  background: linear-gradient(135deg, #8b5cf6 0%, #6366f1 100%);
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

.tab-content {
  padding: 0;
  min-height: 500px;
}

.service-panel {
  padding: 24px;
}

.service-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  background: #f8fafc;
  border-radius: 10px;
  margin-bottom: 20px;
}

.service-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.service-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
}

.service-icon.jenkins {
  background: linear-gradient(135deg, #d33833, #ef5350);
}

.service-icon.harbor {
  background: linear-gradient(135deg, #0277bd, #42a5f5);
}

.service-icon.api {
  background: linear-gradient(135deg, #7c3aed, #a78bfa);
  font-size: 24px;
}

.service-info h3 {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 4px;
}

.service-info p {
  font-size: 13px;
  color: #64748b;
  margin: 0;
}

.status-connected {
  padding: 4px 12px;
  background: #dcfce7;
  color: #16a34a;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 500;
}

.service-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #fafbfc;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
}

.detail-label {
  font-size: 13px;
  color: #64748b;
  min-width: 100px;
  font-weight: 500;
}

.detail-value {
  font-size: 13px;
  color: #1e293b;
}

.token-list {
  padding: 40px;
}

.empty-state {
  text-align: center;
}

.empty-state p {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 16px;
}

.btn-create {
  padding: 8px 20px;
  background: #6366f1;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-create:hover {
  background: #4f46e5;
}
</style>

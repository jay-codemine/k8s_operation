<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">审批中心</h1>
        <p class="page-desc">统一管理发布审批、权限审批、AI风险审核和审批策略配置</p>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ pendingCount }}</span>
          <span class="stat-label">待审批</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ todayCount }}</span>
          <span class="stat-label">今日已处理</span>
        </div>
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
          <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
        </button>
      </div>

      <div class="tab-content">
        <!-- 发布审批 -->
        <Approvals v-if="activeTab === 'release'" />
        <!-- AI 风险审核 -->
        <AIApprovals v-if="activeTab === 'ai-risk'" />
        <!-- 审批策略 -->
        <ApprovalPolicy v-if="activeTab === 'policy'" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Approvals from '@/views/cicd/Approvals.vue'
import AIApprovals from '@/views/security/AIApprovals.vue'
import ApprovalPolicy from '@/views/cicd/ApprovalPolicy.vue'

const activeTab = ref('release')
const pendingCount = ref(0)
const todayCount = ref(0)

const tabs = [
  { key: 'release', label: '发布审批', icon: '🚀', badge: null },
  { key: 'ai-risk', label: 'AI 风险审核', icon: '🤖', badge: null },
  { key: 'policy', label: '审批策略', icon: '📄', badge: null },
]
</script>

<style scoped>
.admin-page {
  padding: 0;
}

.page-header {
  background: linear-gradient(135deg, #f59e0b 0%, #ef4444 100%);
  padding: 24px 32px;
  border-radius: 12px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
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

.header-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  backdrop-filter: blur(4px);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
}

.stat-label {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.8);
  margin-top: 2px;
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
  position: relative;
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

.tab-badge {
  padding: 1px 6px;
  background: #ef4444;
  color: #fff;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
  min-width: 18px;
  text-align: center;
}

.tab-content {
  padding: 0;
  min-height: 500px;
}
</style>

<template>
  <div class="forbidden-container">
    <div class="forbidden-content">
      <!-- 大图标 -->
      <div class="forbidden-icon">
        <svg viewBox="0 0 200 200" width="180" height="180">
          <circle cx="100" cy="100" r="90" fill="none" stroke="#e5e7eb" stroke-width="8"/>
          <circle cx="100" cy="100" r="70" fill="#fef2f2"/>
          <path d="M70 70 L130 130 M130 70 L70 130" stroke="#ef4444" stroke-width="12" stroke-linecap="round"/>
        </svg>
      </div>
      
      <!-- 错误码 -->
      <div class="error-code">403</div>
      
      <!-- 标题 -->
      <h1 class="forbidden-title">访问受限</h1>
      
      <!-- 描述 -->
      <p class="forbidden-desc">
        {{ errorMessage }}
      </p>
      
      <!-- 详情卡片 -->
      <div class="detail-card" v-if="showDetail">
        <div class="detail-item">
          <span class="detail-label">请求路径</span>
          <span class="detail-value">{{ requestPath }}</span>
        </div>
        <div class="detail-item" v-if="requiredRole">
          <span class="detail-label">所需权限</span>
          <span class="detail-value">{{ requiredRole }}</span>
        </div>
        <div class="detail-item" v-if="clusterId">
          <span class="detail-label">目标集群</span>
          <span class="detail-value">集群 #{{ clusterId }}</span>
        </div>
      </div>
      
      <!-- 操作按钮 -->
      <div class="action-buttons">
        <button class="btn-primary" @click="goBack">
          <span class="btn-icon">←</span>
          返回上一页
        </button>
        <button class="btn-secondary" @click="goHome">
          <span class="btn-icon">🏠</span>
          返回首页
        </button>
        <button class="btn-outline" @click="contactAdmin">
          <span class="btn-icon">📧</span>
          联系管理员
        </button>
      </div>
      
      <!-- 提示信息 -->
      <div class="help-tip">
        <span class="tip-icon">💡</span>
        如需访问此资源，请联系系统管理员申请相应权限
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'

const route = useRoute()
const router = useRouter()

// 错误类型
const errorType = computed(() => route.query.type || 'page')

// 错误信息
const errorMessage = computed(() => {
  switch (errorType.value) {
    case 'cluster':
      return '您没有权限访问该集群资源，请联系管理员授权'
    case 'role':
      return '您的角色权限不足，无法访问此功能模块'
    case 'api':
      return '接口访问被拒绝，请确认您拥有相应的操作权限'
    default:
      return '抱歉，您没有权限访问此页面'
  }
})

// 请求路径
const requestPath = computed(() => route.query.path || route.fullPath)

// 所需角色
const requiredRole = computed(() => route.query.role || '')

// 集群ID
const clusterId = computed(() => route.query.clusterId || '')

// 是否显示详情
const showDetail = computed(() => requestPath.value || requiredRole.value || clusterId.value)

// 返回上一页
const goBack = () => {
  if (window.history.length > 2) {
    router.back()
  } else {
    router.push('/dashboard')
  }
}

// 返回首页
const goHome = () => {
  router.push('/dashboard')
}

// 联系管理员
const contactAdmin = () => {
  Message.info('请联系系统管理员处理')
}
</script>

<style scoped>
.forbidden-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  padding: 20px;
}

.forbidden-content {
  text-align: center;
  max-width: 500px;
  width: 100%;
}

.forbidden-icon {
  margin-bottom: 24px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.error-code {
  font-size: 72px;
  font-weight: 800;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 8px;
  line-height: 1;
}

.forbidden-title {
  font-size: 28px;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 12px 0;
}

.forbidden-desc {
  font-size: 16px;
  color: #64748b;
  margin: 0 0 24px 0;
  line-height: 1.6;
}

.detail-card {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  text-align: left;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f1f5f9;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 13px;
  color: #94a3b8;
}

.detail-value {
  font-size: 13px;
  color: #334155;
  font-family: 'Monaco', 'Menlo', monospace;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-buttons {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 24px;
}

.action-buttons button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn-icon {
  font-size: 14px;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.btn-outline {
  background: transparent;
  color: #64748b;
  border: 1px solid #e2e8f0;
}

.btn-outline:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.help-tip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #fefce8;
  border-radius: 8px;
  font-size: 13px;
  color: #854d0e;
}

.tip-icon {
  font-size: 16px;
}
</style>

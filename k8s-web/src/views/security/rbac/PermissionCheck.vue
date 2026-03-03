<template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>🔍 权限校验工具</h1>
      <p>测试用户或 ServiceAccount 对特定资源的访问权限</p>
    </div>

    <div class="permission-check-container">
      <!-- 主体选择 -->
      <div class="check-section">
        <h2>1. 选择主体</h2>
        <div class="form-row">
          <div class="form-group">
            <label>主体类型</label>
            <select v-model="checkForm.subjectType">
              <option value="User">User（用户）</option>
              <option value="ServiceAccount">ServiceAccount（服务账户）</option>
            </select>
          </div>

          <div class="form-group" v-if="checkForm.subjectType === 'User'">
            <label>用户名</label>
            <input v-model="checkForm.username" type="text" placeholder="例如：admin" />
          </div>

          <div class="form-group" v-if="checkForm.subjectType === 'ServiceAccount'">
            <label>ServiceAccount 命名空间</label>
            <select v-model="checkForm.saNamespace">
              <option value="">请选择...</option>
              <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
            </select>
          </div>

          <div class="form-group" v-if="checkForm.subjectType === 'ServiceAccount'">
            <label>ServiceAccount 名称</label>
            <input v-model="checkForm.saName" type="text" placeholder="例如：default" />
          </div>
        </div>
      </div>

      <!-- 资源和操作 -->
      <div class="check-section">
        <h2>2. 指定资源和操作</h2>
        <div class="form-row">
          <div class="form-group">
            <label>命名空间（可选）</label>
            <select v-model="checkForm.namespace">
              <option value="">全集群</option>
              <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
            </select>
          </div>

          <div class="form-group">
            <label>API Group</label>
            <input v-model="checkForm.apiGroup" type="text" placeholder="留空表示核心 API" />
            <p class="help-text">例如：apps, batch, networking.k8s.io</p>
          </div>

          <div class="form-group">
            <label>资源类型 *</label>
            <input v-model="checkForm.resource" type="text" placeholder="例如：pods, deployments" required />
          </div>

          <div class="form-group">
            <label>资源名称（可选）</label>
            <input v-model="checkForm.resourceName" type="text" placeholder="留空表示所有" />
          </div>

          <div class="form-group">
            <label>操作（Verb） *</label>
            <select v-model="checkForm.verb" required>
              <option value="get">get（查看单个）</option>
              <option value="list">list（列表）</option>
              <option value="watch">watch（监听）</option>
              <option value="create">create（创建）</option>
              <option value="update">update（更新）</option>
              <option value="patch">patch（补丁）</option>
              <option value="delete">delete（删除）</option>
            </select>
          </div>
        </div>

        <button class="btn btn-primary btn-large" @click="checkPermission" :disabled="loading">
          {{ loading ? '检查中...' : '🔍 检查权限' }}
        </button>
      </div>

      <!-- 检查结果 -->
      <div v-if="checkResult" class="check-section result-section">
        <h2>3. 检查结果</h2>
        <div class="result-card" :class="checkResult.allowed ? 'allowed' : 'denied'">
          <div class="result-icon">
            {{ checkResult.allowed ? '✅' : '❌' }}
          </div>
          <div class="result-content">
            <h3>{{ checkResult.allowed ? '允许访问' : '拒绝访问' }}</h3>
            <p class="result-summary">
              主体 <strong>{{ getSubjectIdentifier() }}</strong>
              {{ checkResult.allowed ? '可以' : '不能' }}
              对资源 <strong>{{ getResourceIdentifier() }}</strong>
              执行 <strong>{{ checkForm.verb }}</strong> 操作
            </p>

            <div v-if="checkResult.reason" class="result-reason">
              <strong>原因：</strong>
              <span>{{ checkResult.reason }}</span>
            </div>

            <div v-if="checkResult.matchedRoles && checkResult.matchedRoles.length > 0" class="matched-roles">
              <strong>匹配的角色：</strong>
              <div class="role-chips">
                <span v-for="role in checkResult.matchedRoles" :key="role" class="role-chip">
                  {{ role }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 批量检查 -->
      <div class="check-section">
        <h2>4. 批量权限检查（可选）</h2>
        <p class="help-text">快速检查对多种资源的常见操作权限</p>
        
        <div class="batch-check-controls">
          <button class="btn btn-secondary" @click="batchCheckCommonResources" :disabled="loading">
            检查常见资源权限
          </button>
          <button class="btn btn-secondary" @click="batchCheckCurrentNamespace" :disabled="loading || !checkForm.namespace">
            检查当前命名空间权限
          </button>
        </div>

        <div v-if="batchResults.length > 0" class="batch-results">
          <h3>批量检查结果</h3>
          <table class="batch-table">
            <thead>
              <tr>
                <th>资源</th>
                <th>操作</th>
                <th>结果</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(result, index) in batchResults" :key="index">
                <td>{{ result.resource }}</td>
                <td>{{ result.verb }}</td>
                <td>
                  <span class="result-badge" :class="result.allowed ? 'allowed' : 'denied'">
                    {{ result.allowed ? '✅ 允许' : '❌ 拒绝' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Message } from '@arco-design/web-vue'

// 数据状态
const loading = ref(false)
const namespaces = ref(['default', 'kube-system', 'kube-public'])

// 检查表单
const checkForm = ref({
  subjectType: 'User',
  username: '',
  saNamespace: 'default',
  saName: '',
  namespace: '',
  apiGroup: '',
  resource: '',
  resourceName: '',
  verb: 'get'
})

// 检查结果
const checkResult = ref(null)
const batchResults = ref([])

// 检查权限
const checkPermission = async () => {
  // 验证
  if (checkForm.value.subjectType === 'User' && !checkForm.value.username) {
    Message.warning({ content: '请输入用户名' })
    return
  }
  if (checkForm.value.subjectType === 'ServiceAccount' && (!checkForm.value.saNamespace || !checkForm.value.saName)) {
    Message.warning({ content: '请选择 ServiceAccount' })
    return
  }
  if (!checkForm.value.resource) {
    Message.warning({ content: '请输入资源类型' })
    return
  }

  loading.value = true
  try {
    // TODO: 调用后端 API
    // const res = await rbacApi.checkPermission(checkForm.value)
    
    // 模拟结果
    const isAllowed = Math.random() > 0.3
    checkResult.value = {
      allowed: isAllowed,
      reason: isAllowed 
        ? '主体通过 RoleBinding "developer-binding" 拥有此权限' 
        : '主体没有匹配的 RoleBinding 授予此权限',
      matchedRoles: isAllowed ? ['developer', 'view'] : []
    }

    Message.success({ content: '检查完成' })
  } catch (error) {
    console.error('检查失败:', error)
    Message.error({ content: '检查失败' })
  } finally {
    loading.value = false
  }
}

// 批量检查常见资源
const batchCheckCommonResources = async () => {
  loading.value = true
  batchResults.value = []

  const checks = [
    { resource: 'pods', verb: 'get' },
    { resource: 'pods', verb: 'list' },
    { resource: 'pods', verb: 'create' },
    { resource: 'pods', verb: 'delete' },
    { resource: 'deployments', verb: 'get' },
    { resource: 'deployments', verb: 'list' },
    { resource: 'services', verb: 'get' },
    { resource: 'services', verb: 'create' }
  ]

  try {
    for (const check of checks) {
      // TODO: 调用真实 API
      const allowed = Math.random() > 0.3
      batchResults.value.push({
        resource: check.resource,
        verb: check.verb,
        allowed
      })
      // 模拟延迟
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    Message.success({ content: '批量检查完成' })
  } catch (error) {
    console.error('批量检查失败:', error)
    Message.error({ content: '批量检查失败' })
  } finally {
    loading.value = false
  }
}

// 批量检查当前命名空间
const batchCheckCurrentNamespace = async () => {
  if (!checkForm.value.namespace) {
    Message.warning({ content: '请先选择命名空间' })
    return
  }

  Message.info({ content: '开始检查当前命名空间权限...' })
  await batchCheckCommonResources()
}

// 辅助函数
const getSubjectIdentifier = () => {
  if (checkForm.value.subjectType === 'User') {
    return checkForm.value.username
  }
  return `${checkForm.value.saNamespace}/${checkForm.value.saName}`
}

const getResourceIdentifier = () => {
  const group = checkForm.value.apiGroup ? `${checkForm.value.apiGroup}/` : ''
  const ns = checkForm.value.namespace ? `${checkForm.value.namespace}/` : ''
  const name = checkForm.value.resourceName ? `/${checkForm.value.resourceName}` : ''
  return `${group}${ns}${checkForm.value.resource}${name}`
}
</script>

<style scoped>
@import '@/styles/resource-common.css';

.permission-check-container {
  max-width: 1200px;
  margin: 0 auto;
}

.check-section {
  background-color: #fff;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.check-section h2 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #333;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.btn-large {
  padding: 12px 32px;
  font-size: 16px;
  font-weight: 600;
}

.result-section {
  background-color: #f9f9f9;
}

.result-card {
  display: flex;
  gap: 24px;
  padding: 24px;
  border-radius: 8px;
  border: 2px solid;
}

.result-card.allowed {
  background-color: #e8f5e9;
  border-color: #4caf50;
}

.result-card.denied {
  background-color: #ffebee;
  border-color: #f44336;
}

.result-icon {
  font-size: 48px;
  flex-shrink: 0;
}

.result-content {
  flex: 1;
}

.result-content h3 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.result-summary {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
  margin-bottom: 12px;
}

.result-summary strong {
  color: #333;
  font-weight: 600;
}

.result-reason {
  padding: 12px;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 13px;
}

.matched-roles {
  margin-top: 12px;
}

.role-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.role-chip {
  background-color: #e3f2fd;
  color: #1976d2;
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
}

.batch-check-controls {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.batch-results {
  margin-top: 24px;
}

.batch-results h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

.batch-table {
  width: 100%;
  border-collapse: collapse;
  background-color: #fff;
}

.batch-table th,
.batch-table td {
  padding: 12px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.batch-table th {
  background-color: #f5f5f5;
  font-weight: 600;
}

.result-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.result-badge.allowed {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.result-badge.denied {
  background-color: #ffebee;
  color: #c62828;
}
</style>

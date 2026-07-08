<template>
  <div class="rbac-view">
    <!-- 页面头部 -->
    <div class="rbac-header">
      <div class="rbac-header-left">
        <div class="rbac-icon">🔍</div>
        <div class="rbac-title-group">
          <h1>权限校验工具</h1>
          <p>测试用户或 ServiceAccount 对特定资源的访问权限</p>
        </div>
      </div>
      <div class="rbac-header-right">
        <ClusterSelector
          v-model="selectedClusterId"
          :clusters="clusters"
          :show-all-option="false"
          label="集群"
          @change="onClusterChange"
        />
      </div>
    </div>

    <div class="permission-check-container">
      <!-- 步骤 1: 选择主体 -->
      <div class="check-card">
        <div class="card-header">
          <span class="step-number">1</span>
          <h2>选择主体</h2>
        </div>
        <div class="card-body">
          <div class="rbac-form-row">
            <div class="rbac-form-group">
              <label>主体类型</label>
              <select v-model="checkForm.subjectType">
                <option value="User">User（用户）</option>
                <option value="ServiceAccount">ServiceAccount（服务账户）</option>
              </select>
            </div>
            <div class="rbac-form-group" v-if="checkForm.subjectType === 'User'">
              <label>用户名</label>
              <input v-model="checkForm.username" placeholder="例如：admin" />
            </div>
            <div class="rbac-form-group" v-if="checkForm.subjectType === 'ServiceAccount'">
              <label>SA 命名空间</label>
              <select v-model="checkForm.saNamespace">
                <option value="">请选择...</option>
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>
            <div class="rbac-form-group" v-if="checkForm.subjectType === 'ServiceAccount'">
              <label>SA 名称</label>
              <input v-model="checkForm.saName" placeholder="例如：default" />
            </div>
          </div>
        </div>
      </div>

      <!-- 步骤 2: 指定资源和操作 -->
      <div class="check-card">
        <div class="card-header">
          <span class="step-number">2</span>
          <h2>指定资源和操作</h2>
        </div>
        <div class="card-body">
          <div class="rbac-form-row">
            <div class="rbac-form-group">
              <label>命名空间（可选）</label>
              <select v-model="checkForm.namespace">
                <option value="">全集群</option>
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>
            <div class="rbac-form-group">
              <label>API Group</label>
              <input v-model="checkForm.apiGroup" placeholder="留空表示核心 API" />
              <span class="help-text">例如：apps, batch, networking.k8s.io</span>
            </div>
            <div class="rbac-form-group">
              <label>资源类型 *</label>
              <input v-model="checkForm.resource" placeholder="例如：pods, deployments" required />
            </div>
            <div class="rbac-form-group">
              <label>资源名称（可选）</label>
              <input v-model="checkForm.resourceName" placeholder="留空表示所有" />
            </div>
            <div class="rbac-form-group">
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
          <button class="rbac-btn rbac-btn-primary check-btn" @click="checkPermission" :disabled="loading || !selectedClusterId">
            {{ loading ? '检查中...' : '🔍 检查权限' }}
          </button>
        </div>
      </div>

      <!-- 步骤 3: 检查结果 -->
      <div v-if="checkResult" class="check-card result-card" :class="checkResult.allowed ? 'allowed' : 'denied'">
        <div class="card-header">
          <span class="step-number">3</span>
          <h2>检查结果</h2>
        </div>
        <div class="card-body">
          <div class="result-display">
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
                <strong>原因：</strong>{{ checkResult.reason }}
              </div>
              <div v-if="checkResult.matched_roles?.length > 0" class="matched-roles">
                <strong>匹配的角色：</strong>
                <div class="role-chips">
                  <span v-for="role in checkResult.matched_roles" :key="role" class="rbac-tag rbac-tag-primary">
                    {{ role }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 步骤 4: 批量检查 -->
      <div class="check-card">
        <div class="card-header">
          <span class="step-number">4</span>
          <h2>批量权限检查（可选）</h2>
        </div>
        <div class="card-body">
          <p class="help-text">快速检查对多种资源的常见操作权限</p>
          <div class="batch-actions">
            <button class="rbac-btn rbac-btn-secondary" @click="batchCheckCommonResources" :disabled="loading || !selectedClusterId || !hasValidSubject">
              检查常见资源权限
            </button>
            <button class="rbac-btn rbac-btn-secondary" @click="batchCheckCurrentNamespace" :disabled="loading || !checkForm.namespace || !selectedClusterId || !hasValidSubject">
              检查当前命名空间权限
            </button>
          </div>

          <div v-if="batchResults.length > 0" class="batch-results">
            <h3>批量检查结果</h3>
            <div class="rbac-table-container">
              <table class="rbac-table">
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
                      <span class="rbac-tag" :class="result.allowed ? 'rbac-tag-success' : 'rbac-tag-danger'">
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
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import ClusterSelector from '@/components/cluster/ClusterSelector.vue'
import { checkSubjectAccess, batchCheckSubjectAccess } from '@/api/k8sRbac'
import { getClusterList } from '@/api/cluster'
import { getNamespaces } from '@/api/namespace'
import permissionStore from '@/stores/permission'

// 集群选择
const clusters = ref([])
const selectedClusterId = ref('')
const namespaces = ref([])
const loading = ref(false)

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

// 验证主体是否有效
const hasValidSubject = computed(() => {
  if (checkForm.value.subjectType === 'User') {
    return !!checkForm.value.username
  }
  return !!checkForm.value.saNamespace && !!checkForm.value.saName
})

// 监听集群变化
const onClusterChange = (clusterId) => {
  if (clusterId) {
    loadNamespaces()
    checkResult.value = null
    batchResults.value = []
  }
}

watch(selectedClusterId, (val) => {
  if (val) {
    loadNamespaces()
    checkResult.value = null
    batchResults.value = []
  }
})

// 加载集群列表
const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0 && res.data?.list) {
      // 权限过滤：只显示用户有权限访问的集群
      const allClusters = res.data.list
      clusters.value = allClusters
        .filter(c => permissionStore.state.isSuperAdmin || permissionStore.state.accessibleClusterIds.includes(c.id))
        .map(c => ({ ...c, name: c.cluster_name || c.name }))
      if (clusters.value.length > 0 && !selectedClusterId.value) {
        selectedClusterId.value = clusters.value[0].id
      }
    }
  } catch (error) {
    console.error('加载集群列表失败:', error)
  }
}

// 加载命名空间列表
const loadNamespaces = async () => {
  if (!selectedClusterId.value) return
  try {
    const res = await getNamespaces(selectedClusterId.value)
    if (res.code === 0 && res.data?.list) {
      namespaces.value = res.data.list.map(ns => ns.name || ns)
    }
  } catch (error) {
    namespaces.value = ['default', 'kube-system', 'kube-public']
  }
}

// 获取主体标识
const getSubjectIdentifier = () => {
  if (checkForm.value.subjectType === 'User') {
    return `User:${checkForm.value.username}`
  }
  return `SA:${checkForm.value.saNamespace}/${checkForm.value.saName}`
}

// 获取资源标识
const getResourceIdentifier = () => {
  const parts = []
  if (checkForm.value.namespace) parts.push(checkForm.value.namespace)
  if (checkForm.value.apiGroup) parts.push(checkForm.value.apiGroup)
  parts.push(checkForm.value.resource)
  if (checkForm.value.resourceName) parts.push(checkForm.value.resourceName)
  return parts.join('/')
}

// 检查权限
const checkPermission = async () => {
  if (!selectedClusterId.value || !checkForm.value.resource) {
    Message.warning({ content: '请选择集群并填写资源类型' })
    return
  }
  if (!hasValidSubject.value) {
    Message.warning({ content: '请填写完整的主体信息' })
    return
  }
  loading.value = true
  try {
    const res = await checkSubjectAccess(selectedClusterId.value, {
      subject_type: checkForm.value.subjectType,
      username: checkForm.value.username,
      sa_namespace: checkForm.value.saNamespace,
      sa_name: checkForm.value.saName,
      namespace: checkForm.value.namespace,
      api_group: checkForm.value.apiGroup,
      resource: checkForm.value.resource,
      resource_name: checkForm.value.resourceName,
      verb: checkForm.value.verb
    })
    checkResult.value = res.code === 0 ? res.data : { allowed: false, reason: res.msg || '检查失败' }
  } catch (error) {
    checkResult.value = { allowed: false, reason: error.msg || error.message || '检查失败' }
  } finally {
    loading.value = false
  }
}

// 批量检查常见资源
const batchCheckCommonResources = async () => {
  const resources = [
    { resource: 'pods', verb: 'get' },
    { resource: 'pods', verb: 'list' },
    { resource: 'pods', verb: 'create' },
    { resource: 'pods', verb: 'delete' },
    { resource: 'deployments', verb: 'get' },
    { resource: 'deployments', verb: 'list' },
    { resource: 'services', verb: 'get' },
    { resource: 'configmaps', verb: 'get' },
    { resource: 'secrets', verb: 'get' }
  ]
  await doBatchCheck(resources)
}

// 批量检查当前命名空间权限
const batchCheckCurrentNamespace = async () => {
  const verbs = ['get', 'list', 'create', 'update', 'delete']
  const resources = []
  for (const verb of verbs) {
    resources.push({ resource: 'pods', verb, namespace: checkForm.value.namespace })
    resources.push({ resource: 'deployments', verb, namespace: checkForm.value.namespace, api_group: 'apps' })
  }
  await doBatchCheck(resources)
}

// 执行批量检查
const doBatchCheck = async (resources) => {
  loading.value = true
  batchResults.value = []
  try {
    const res = await batchCheckSubjectAccess(selectedClusterId.value, {
      subject_type: checkForm.value.subjectType,
      username: checkForm.value.username,
      sa_namespace: checkForm.value.saNamespace,
      sa_name: checkForm.value.saName,
      checks: resources
    })
    batchResults.value = res.code === 0 && res.data?.results ? res.data.results : []
  } catch (error) {
    Message.error({ content: '批量检查失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

onMounted(() => { loadClusters() })
</script>

<style scoped>
@import '@/assets/styles/rbac-common.css';

.permission-check-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 检查卡片 - 浅色主题 */
.check-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.step-number {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 50%;
  font-size: 14px;
  font-weight: 600;
  color: white;
}

.card-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.card-body {
  padding: 20px;
}

/* 检查按钮 */
.check-btn {
  margin-top: 16px;
  padding: 12px 24px;
  font-size: 14px;
}

/* 结果卡片 */
.result-card.allowed {
  border-color: #a7f3d0;
  background: linear-gradient(135deg, #ecfdf5, #ffffff);
}

.result-card.denied {
  border-color: #fecaca;
  background: linear-gradient(135deg, #fef2f2, #ffffff);
}

.result-display {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.result-icon {
  font-size: 48px;
  flex-shrink: 0;
}

.result-content {
  flex: 1;
}

.result-content h3 {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.result-summary {
  font-size: 14px;
  color: #4b5563;
  margin: 0 0 12px;
}

.result-summary strong {
  color: #6366f1;
}

.result-reason {
  padding: 10px 12px;
  background: #f3f4f6;
  border-radius: 6px;
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 12px;
}

.matched-roles {
  margin-top: 12px;
}

.matched-roles strong {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 8px;
}

.role-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

/* 批量检查 */
.batch-actions {
  display: flex;
  gap: 12px;
  margin-top: 12px;
}

.batch-results {
  margin-top: 20px;
}

.batch-results h3 {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
  margin: 0 0 12px;
}

.rbac-tag-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.rbac-tag-success {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

.rbac-tag-primary {
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
}

.help-text {
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}
</style>

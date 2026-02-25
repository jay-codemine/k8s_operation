<template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>🎭 Role 管理</h1>
      <p>管理 Kubernetes 角色权限，支持 Role（命名空间级）和 ClusterRole（集群级）</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 Role 名称..."
        />
      </div>

      <div class="filter-dropdown">
        <select v-model="roleTypeFilter">
          <option value="">所有类型</option>
          <option value="Role">Role（命名空间级）</option>
          <option value="ClusterRole">ClusterRole（集群级）</option>
        </select>
      </div>

      <div class="filter-dropdown" v-if="roleTypeFilter !== 'ClusterRole'">
        <select v-model="namespaceFilter" @change="loadRoles">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <button class="btn btn-primary" @click="openCreateModal">+ 创建 Role</button>
        <button class="btn btn-success" @click="openTemplateModal">📋 使用模板</button>
        <button class="btn btn-secondary" @click="loadRoles" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 表格视图 -->
    <div class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th style="width: 250px;">名称</th>
            <th style="width: 120px;">类型</th>
            <th style="width: 150px;">命名空间</th>
            <th style="width: 100px;">规则数</th>
            <th style="width: 150px;">创建时间</th>
            <th style="width: 260px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="role in filteredRoles" :key="`${role.type}-${role.namespace || 'cluster'}-${role.name}`">
            <td>
              <div class="resource-name">
                <span class="icon">{{ role.type === 'ClusterRole' ? '🌍' : '📁' }}</span>
                <span>{{ role.name }}</span>
              </div>
            </td>
            <td>
              <span class="type-badge" :class="role.type">
                {{ role.type === 'ClusterRole' ? '集群级' : '命名空间' }}
              </span>
            </td>
            <td>
              <span v-if="role.namespace" class="namespace-tag">{{ role.namespace }}</span>
              <span v-else class="cluster-tag">全集群</span>
            </td>
            <td>{{ role.rules?.length || 0 }}</td>
            <td>{{ formatDate(role.created_at) }}</td>
            <td class="actions">
              <button class="btn btn-sm btn-info" @click="viewRules(role)" title="查看规则">
                👁️ 规则
              </button>
              <button class="btn btn-sm btn-success" @click="viewBindings(role)" title="查看绑定">
                🔗 绑定
              </button>
              <button class="btn btn-sm btn-warning" @click="editRole(role)" title="编辑">
                ✏️ 编辑
              </button>
              <button class="btn btn-sm btn-danger" @click="deleteRole(role)" title="删除">
                🗑️ 删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="filteredRoles.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">暂无 Role，请先创建</div>
      </div>
    </div>

    <!-- 创建 Role 模态框 -->
    <div v-if="showCreateModal" class="modal" @click.self="closeCreateModal">
      <div class="modal-content modal-extra-large">
        <div class="modal-header">
          <h2>{{ editMode ? '编辑' : '创建' }} Role</h2>
          <button class="close-btn" @click="closeCreateModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitRole">
            <!-- 基本信息 -->
            <div class="form-section">
              <h3>基本信息</h3>
              <div class="form-row">
                <div class="form-group">
                  <label>名称 *</label>
                  <input v-model="roleForm.name" type="text" placeholder="例如：developer" required />
                </div>

                <div class="form-group">
                  <label>类型 *</label>
                  <select v-model="roleForm.type" required>
                    <option value="Role">Role（命名空间级）</option>
                    <option value="ClusterRole">ClusterRole（集群级）</option>
                  </select>
                </div>

                <div class="form-group" v-if="roleForm.type === 'Role'">
                  <label>命名空间 *</label>
                  <select v-model="roleForm.namespace" required>
                    <option value="">请选择命名空间</option>
                    <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 权限规则 -->
            <div class="form-section">
              <h3>
                权限规则
                <span class="help-text">（定义此角色可以对哪些资源执行哪些操作）</span>
              </h3>

              <div v-for="(rule, index) in roleForm.rules" :key="index" class="rule-card">
                <div class="rule-header">
                  <h4>规则 {{ index + 1 }}</h4>
                  <button type="button" class="btn btn-sm btn-danger" @click="removeRule(index)">删除规则</button>
                </div>

                <div class="rule-body">
                  <!-- API Groups -->
                  <div class="form-group">
                    <label>API Groups *</label>
                    <div class="chip-input">
                      <span v-for="(group, gIndex) in rule.apiGroups" :key="gIndex" class="chip">
                        {{ group || '""（核心 API）' }}
                        <button type="button" @click="removeApiGroup(index, gIndex)">&times;</button>
                      </span>
                      <input
                        v-model="apiGroupInput[index]"
                        type="text"
                        placeholder="输入 API Group，回车添加（留空表示核心 API）"
                        @keyup.enter="addApiGroup(index)"
                      />
                    </div>
                    <p class="help-text">
                      常见：空（核心）、apps、batch、networking.k8s.io、rbac.authorization.k8s.io
                    </p>
                  </div>

                  <!-- Resources -->
                  <div class="form-group">
                    <label>Resources *</label>
                    <div class="chip-input">
                      <span v-for="(resource, rIndex) in rule.resources" :key="rIndex" class="chip">
                        {{ resource }}
                        <button type="button" @click="removeResource(index, rIndex)">&times;</button>
                      </span>
                      <input
                        v-model="resourceInput[index]"
                        type="text"
                        placeholder="输入资源类型，回车添加"
                        @keyup.enter="addResource(index)"
                      />
                    </div>
                    <p class="help-text">
                      常见：pods, deployments, services, configmaps, secrets, nodes
                    </p>
                  </div>

                  <!-- Verbs -->
                  <div class="form-group">
                    <label>Verbs（操作） *</label>
                    <div class="verbs-checkbox-group">
                      <label v-for="verb in availableVerbs" :key="verb">
                        <input type="checkbox" :value="verb" v-model="rule.verbs" />
                        <span>{{ verb }}</span>
                      </label>
                    </div>
                    <button type="button" class="btn btn-sm btn-secondary" @click="selectAllVerbs(index)">
                      全选
                    </button>
                    <button type="button" class="btn btn-sm btn-secondary" @click="selectReadOnlyVerbs(index)">
                      只读权限
                    </button>
                  </div>

                  <!-- Resource Names (可选) -->
                  <div class="form-group">
                    <label>Resource Names（可选，限制特定资源）</label>
                    <div class="chip-input">
                      <span v-for="(name, nIndex) in rule.resourceNames" :key="nIndex" class="chip">
                        {{ name }}
                        <button type="button" @click="removeResourceName(index, nIndex)">&times;</button>
                      </span>
                      <input
                        v-model="resourceNameInput[index]"
                        type="text"
                        placeholder="输入资源名称，回车添加"
                        @keyup.enter="addResourceName(index)"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <button type="button" class="btn btn-secondary" @click="addRule">+ 添加规则</button>
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? '提交中...' : (editMode ? '更新' : '创建') }}
              </button>
              <button type="button" class="btn btn-secondary" @click="closeCreateModal">取消</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 模板选择模态框 -->
    <div v-if="showTemplateModal" class="modal" @click.self="closeTemplateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>选择 Role 模板</h2>
          <button class="close-btn" @click="closeTemplateModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="template-grid">
            <div
              v-for="template in roleTemplates"
              :key="template.id"
              class="template-card"
              @click="useTemplate(template)"
            >
              <div class="template-icon">{{ template.icon }}</div>
              <h3>{{ template.name }}</h3>
              <p>{{ template.description }}</p>
              <div class="template-meta">
                <span>{{ template.rules.length }} 条规则</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 规则详情模态框 -->
    <div v-if="showRulesModal" class="modal" @click.self="closeRulesModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>{{ currentRole?.name }} - 权限规则</h2>
          <button class="close-btn" @click="closeRulesModal">&times;</button>
        </div>
        <div class="modal-body">
          <div v-for="(rule, index) in currentRole?.rules" :key="index" class="rule-display">
            <h4>规则 {{ index + 1 }}</h4>
            <table class="rule-table">
              <tr>
                <td class="label">API Groups:</td>
                <td>
                  <span v-for="group in rule.apiGroups" :key="group" class="rule-tag">
                    {{ group || '""（核心）' }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="label">Resources:</td>
                <td>
                  <span v-for="resource in rule.resources" :key="resource" class="rule-tag">
                    {{ resource }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="label">Verbs:</td>
                <td>
                  <span v-for="verb in rule.verbs" :key="verb" class="rule-tag verb-tag">
                    {{ verb }}
                  </span>
                </td>
              </tr>
              <tr v-if="rule.resourceNames && rule.resourceNames.length > 0">
                <td class="label">Resource Names:</td>
                <td>
                  <span v-for="name in rule.resourceNames" :key="name" class="rule-tag">
                    {{ name }}
                  </span>
                </td>
              </tr>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'

// 数据状态
const loading = ref(false)
const searchQuery = ref('')
const roleTypeFilter = ref('')
const namespaceFilter = ref('')
const roles = ref([])
const namespaces = ref(['default', 'kube-system', 'kube-public'])

// 模态框状态
const showCreateModal = ref(false)
const showTemplateModal = ref(false)
const showRulesModal = ref(false)
const editMode = ref(false)
const currentRole = ref(null)

// 表单数据
const roleForm = ref({
  name: '',
  type: 'Role',
  namespace: 'default',
  rules: []
})

// 可用的 Verbs
const availableVerbs = ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete', 'deletecollection']

// 输入辅助
const apiGroupInput = ref({})
const resourceInput = ref({})
const resourceNameInput = ref({})

// Role 模板
const roleTemplates = ref([
  {
    id: 'view',
    name: '只读权限',
    icon: '👁️',
    description: '允许查看所有资源，但不能修改',
    rules: [
      {
        apiGroups: ['', 'apps', 'batch'],
        resources: ['pods', 'deployments', 'services', 'jobs', 'cronjobs'],
        verbs: ['get', 'list', 'watch']
      }
    ]
  },
  {
    id: 'edit',
    name: '编辑权限',
    icon: '✏️',
    description: '允许查看和编辑大部分资源',
    rules: [
      {
        apiGroups: ['', 'apps'],
        resources: ['pods', 'deployments', 'services', 'configmaps'],
        verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']
      }
    ]
  },
  {
    id: 'admin',
    name: '管理员权限',
    icon: '👑',
    description: '完整的管理权限（除 RBAC）',
    rules: [
      {
        apiGroups: ['*'],
        resources: ['*'],
        verbs: ['*']
      }
    ]
  },
  {
    id: 'pod-reader',
    name: 'Pod 只读',
    icon: '📦',
    description: '只能查看 Pod',
    rules: [
      {
        apiGroups: [''],
        resources: ['pods', 'pods/log', 'pods/status'],
        verbs: ['get', 'list', 'watch']
      }
    ]
  }
])

// 过滤后的 Roles
const filteredRoles = computed(() => {
  let result = roles.value

  if (roleTypeFilter.value) {
    result = result.filter(r => r.type === roleTypeFilter.value)
  }

  if (namespaceFilter.value && roleTypeFilter.value !== 'ClusterRole') {
    result = result.filter(r => r.namespace === namespaceFilter.value)
  }

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(r => r.name.toLowerCase().includes(query))
  }

  return result
})

// 加载 Roles
const loadRoles = async () => {
  loading.value = true
  try {
    // 模拟数据
    roles.value = [
      {
        name: 'view',
        type: 'ClusterRole',
        namespace: null,
        rules: [
          {
            apiGroups: [''],
            resources: ['pods', 'services'],
            verbs: ['get', 'list', 'watch']
          }
        ],
        created_at: '2024-01-15T10:30:00Z'
      },
      {
        name: 'developer',
        type: 'Role',
        namespace: 'default',
        rules: [
          {
            apiGroups: ['', 'apps'],
            resources: ['pods', 'deployments', 'services'],
            verbs: ['get', 'list', 'create', 'update']
          }
        ],
        created_at: '2024-02-01T14:20:00Z'
      }
    ]
  } catch (error) {
    console.error('加载失败:', error)
    Message.error({ content: '加载失败' })
  } finally {
    loading.value = false
  }
}

// 创建/更新 Role
const submitRole = async () => {
  // 验证
  if (roleForm.value.rules.length === 0) {
    Message.warning({ content: '请至少添加一条规则' })
    return
  }

  loading.value = true
  try {
    // TODO: 调用 API
    Message.success({ content: editMode.value ? '更新成功' : '创建成功' })
    closeCreateModal()
    await loadRoles()
  } catch (error) {
    console.error('提交失败:', error)
    Message.error({ content: '提交失败' })
  } finally {
    loading.value = false
  }
}

// 删除 Role
const deleteRole = async (role) => {
  if (!confirm(`确认删除 ${role.type} "${role.name}"？`)) return

  loading.value = true
  try {
    // TODO: 调用 API
    Message.success({ content: '删除成功' })
    await loadRoles()
  } catch (error) {
    console.error('删除失败:', error)
    Message.error({ content: '删除失败' })
  } finally {
    loading.value = false
  }
}

// 规则管理
const addRule = () => {
  roleForm.value.rules.push({
    apiGroups: [],
    resources: [],
    verbs: [],
    resourceNames: []
  })
}

const removeRule = (index) => {
  roleForm.value.rules.splice(index, 1)
}

// API Group 管理
const addApiGroup = (ruleIndex) => {
  const value = apiGroupInput.value[ruleIndex]?.trim()
  if (value !== undefined && !roleForm.value.rules[ruleIndex].apiGroups.includes(value)) {
    roleForm.value.rules[ruleIndex].apiGroups.push(value)
    apiGroupInput.value[ruleIndex] = ''
  }
}

const removeApiGroup = (ruleIndex, groupIndex) => {
  roleForm.value.rules[ruleIndex].apiGroups.splice(groupIndex, 1)
}

// Resource 管理
const addResource = (ruleIndex) => {
  const value = resourceInput.value[ruleIndex]?.trim()
  if (value && !roleForm.value.rules[ruleIndex].resources.includes(value)) {
    roleForm.value.rules[ruleIndex].resources.push(value)
    resourceInput.value[ruleIndex] = ''
  }
}

const removeResource = (ruleIndex, resourceIndex) => {
  roleForm.value.rules[ruleIndex].resources.splice(resourceIndex, 1)
}

// Resource Name 管理
const addResourceName = (ruleIndex) => {
  const value = resourceNameInput.value[ruleIndex]?.trim()
  if (value && !roleForm.value.rules[ruleIndex].resourceNames.includes(value)) {
    roleForm.value.rules[ruleIndex].resourceNames.push(value)
    resourceNameInput.value[ruleIndex] = ''
  }
}

const removeResourceName = (ruleIndex, nameIndex) => {
  roleForm.value.rules[ruleIndex].resourceNames.splice(nameIndex, 1)
}

// Verbs 快捷选择
const selectAllVerbs = (ruleIndex) => {
  roleForm.value.rules[ruleIndex].verbs = [...availableVerbs]
}

const selectReadOnlyVerbs = (ruleIndex) => {
  roleForm.value.rules[ruleIndex].verbs = ['get', 'list', 'watch']
}

// 查看规则
const viewRules = (role) => {
  currentRole.value = role
  showRulesModal.value = true
}

// 查看绑定
const viewBindings = (role) => {
  Message.info({ content: `查看 ${role.name} 的绑定（功能开发中）` })
}

// 编辑
const editRole = (role) => {
  editMode.value = true
  roleForm.value = JSON.parse(JSON.stringify(role))
  showCreateModal.value = true
}

// 使用模板
const useTemplate = (template) => {
  roleForm.value.rules = JSON.parse(JSON.stringify(template.rules))
  closeTemplateModal()
  showCreateModal.value = true
  Message.success({ content: `已应用模板：${template.name}` })
}

// 模态框管理
const openCreateModal = () => {
  editMode.value = false
  roleForm.value = {
    name: '',
    type: 'Role',
    namespace: 'default',
    rules: []
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const openTemplateModal = () => {
  showTemplateModal.value = true
}

const closeTemplateModal = () => {
  showTemplateModal.value = false
}

const closeRulesModal = () => {
  showRulesModal.value = false
  currentRole.value = null
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadRoles()
})
</script>

<style scoped>
@import '@/styles/resource-common.css';

.form-section {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e0e0e0;
}

.form-section h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #333;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.rule-card {
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #f9f9f9;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.rule-header h4 {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.rule-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.chip-input {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 8px;
  background-color: #fff;
}

.chip {
  background-color: #e3f2fd;
  color: #1976d2;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.chip button {
  background: none;
  border: none;
  color: #1976d2;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 0;
}

.chip-input input {
  flex: 1;
  min-width: 200px;
  border: none;
  outline: none;
  font-size: 14px;
}

.verbs-checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
  padding: 12px;
  background-color: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.verbs-checkbox-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.type-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.type-badge.Role {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.type-badge.ClusterRole {
  background-color: #e3f2fd;
  color: #1565c0;
}

.cluster-tag {
  background-color: #f3e5f5;
  color: #6a1b9a;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
}

.template-card {
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
}

.template-card:hover {
  border-color: #1976d2;
  box-shadow: 0 4px 12px rgba(25, 118, 210, 0.2);
  transform: translateY(-2px);
}

.template-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.template-card h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
}

.template-card p {
  font-size: 13px;
  color: #666;
  margin-bottom: 12px;
}

.template-meta {
  font-size: 12px;
  color: #999;
}

.rule-display {
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.rule-display h4 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

.rule-table {
  width: 100%;
  border-collapse: collapse;
}

.rule-table td {
  padding: 8px;
  vertical-align: top;
}

.rule-table .label {
  font-weight: 600;
  color: #666;
  width: 150px;
}

.rule-tag {
  display: inline-block;
  background-color: #e8eaf6;
  color: #3f51b5;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  margin: 2px;
}

.verb-tag {
  background-color: #fff3e0;
  color: #e65100;
}

.modal-extra-large {
  max-width: 1200px;
  max-height: 90vh;
  overflow-y: auto;
}
</style>

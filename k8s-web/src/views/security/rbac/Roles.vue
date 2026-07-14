<template>
  <div class="rbac-view">
    <!-- 页面头部 -->
    <div class="rbac-header">
      <div class="rbac-header-left">
        <div class="rbac-icon">🎭</div>
        <div class="rbac-title-group">
          <h1>Role 管理</h1>
          <p>管理 Kubernetes 角色权限，支持 Role（命名空间级）和 ClusterRole（集群级）</p>
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

    <!-- 统计卡片 -->
    <div class="rbac-stats">
      <div class="stat-card primary">
        <div class="stat-header">
          <div class="stat-icon">📊</div>
        </div>
        <div class="stat-value">{{ roles.length }}</div>
        <div class="stat-label">总角色数</div>
      </div>
      <div class="stat-card success">
        <div class="stat-header">
          <div class="stat-icon">📁</div>
        </div>
        <div class="stat-value">{{ roleCount }}</div>
        <div class="stat-label">命名空间级 Role</div>
      </div>
      <div class="stat-card warning">
        <div class="stat-header">
          <div class="stat-icon">🌍</div>
        </div>
        <div class="stat-value">{{ clusterRoleCount }}</div>
        <div class="stat-label">集群级 ClusterRole</div>
      </div>
      <div class="stat-card info">
        <div class="stat-header">
          <div class="stat-icon">📜</div>
        </div>
        <div class="stat-value">{{ totalRulesCount }}</div>
        <div class="stat-label">权限规则总数</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="rbac-toolbar">
      <div class="toolbar-left">
        <div class="rbac-search">
          <span class="search-icon">🔍</span>
          <input
            type="text"
            v-model="searchQuery"
            placeholder="搜索 Role 名称..."
          />
        </div>
        <div class="rbac-filter">
          <select v-model="roleTypeFilter">
            <option value="">所有类型</option>
            <option value="Role">Role（命名空间级）</option>
            <option value="ClusterRole">ClusterRole（集群级）</option>
          </select>
        </div>
        <div class="rbac-filter" v-if="roleTypeFilter !== 'ClusterRole'">
          <select v-model="namespaceFilter" @change="loadRoles">
            <option value="">所有命名空间</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>
      </div>
      <div class="toolbar-right">
        <button class="rbac-btn rbac-btn-secondary" @click="loadRoles" :disabled="loading">
          {{ loading ? '⏳' : '🔄' }} 刷新
        </button>
        <button class="rbac-btn rbac-btn-secondary" @click="openTemplateModal">
          📋 使用模板
        </button>
        <button class="rbac-btn rbac-btn-primary" @click="openCreateModal">
          + 创建 Role
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && roles.length === 0" class="rbac-loading">
      <div class="rbac-spinner"></div>
      <div class="rbac-loading-text">加载中...</div>
    </div>

    <!-- 表格 -->
    <div v-else class="rbac-table-container">
      <table class="rbac-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>命名空间</th>
            <th>规则数</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="role in filteredRoles" :key="`${role.type}-${role.namespace || 'cluster'}-${role.name}`">
            <td>
              <div class="resource-name">
                <div class="icon" :class="role.type === 'ClusterRole' ? 'cluster' : 'namespace'">
                  {{ role.type === 'ClusterRole' ? '🌍' : '📁' }}
                </div>
                <div>
                  <div class="name">{{ role.name }}</div>
                  <div class="meta" v-if="role.rules">{{ role.rules.length }} 条规则</div>
                </div>
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
            <td>
              <span class="rbac-tag rbac-tag-info">{{ role.rules?.length || 0 }}</span>
            </td>
            <td>{{ formatDate(role.created_at) }}</td>
            <td>
              <div class="rbac-actions">
                <button class="rbac-action-btn view" @click="viewRules(role)" title="查看规则">👁️</button>
                <button class="rbac-action-btn" @click="viewBindings(role)" title="查看绑定">🔗</button>
                <button class="rbac-action-btn edit" @click="editRole(role)" title="编辑">✏️</button>
                <button class="rbac-action-btn delete" @click="deleteRole(role)" title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 空状态 -->
      <div v-if="filteredRoles.length === 0" class="rbac-empty">
        <div class="rbac-empty-icon">📦</div>
        <div class="rbac-empty-title">暂无 Role</div>
        <div class="rbac-empty-desc">{{ searchQuery ? '没有匹配的结果' : '点击上方按钮创建第一个 Role' }}</div>
        <button v-if="!searchQuery" class="rbac-btn rbac-btn-primary" @click="openCreateModal">
          + 创建 Role
        </button>
      </div>
    </div>

    <!-- 创建 Role 模态框 -->
    <div v-if="showCreateModal" class="rbac-modal" @click.self="closeCreateModal">
      <div class="rbac-modal-content xlarge">
        <div class="rbac-modal-header">
          <h2>{{ editMode ? '编辑' : '创建' }} Role</h2>
          <button class="rbac-modal-close" @click="closeCreateModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <form @submit.prevent="submitRole">
            <!-- 基本信息 -->
            <div class="rbac-form-section">
              <h3>基本信息</h3>
              <div class="rbac-form-row">
                <div class="rbac-form-group">
                  <label>名称 *</label>
                  <input v-model="roleForm.name" type="text" placeholder="例如：developer" required />
                </div>
                <div class="rbac-form-group">
                  <label>类型 *</label>
                  <select v-model="roleForm.type" required>
                    <option value="Role">Role（命名空间级）</option>
                    <option value="ClusterRole">ClusterRole（集群级）</option>
                  </select>
                </div>
                <div class="rbac-form-group" v-if="roleForm.type === 'Role'">
                  <label>命名空间 *</label>
                  <select v-model="roleForm.namespace" required>
                    <option value="">请选择命名空间</option>
                    <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 权限规则 -->
            <div class="rbac-form-section">
              <h3>权限规则 <span class="help-text">（定义此角色可以对哪些资源执行哪些操作）</span></h3>
              
              <div v-for="(rule, index) in roleForm.rules" :key="index" class="rule-card">
                <div class="rule-header">
                  <h4>规则 {{ index + 1 }}</h4>
                  <button type="button" class="rbac-btn rbac-btn-sm rbac-btn-danger" @click="removeRule(index)">
                    删除规则
                  </button>
                </div>
                <div class="rule-body">
                  <!-- API Groups -->
                  <div class="rbac-form-group">
                    <label>API Groups *</label>
                    <div class="chip-input">
                      <span v-for="(group, gIndex) in rule.apiGroups" :key="gIndex" class="chip">
                        {{ group || '""（核心 API）' }}
                        <button type="button" @click="removeApiGroup(index, gIndex)">&times;</button>
                      </span>
                      <input
                        v-model="apiGroupInput[index]"
                        type="text"
                        placeholder="输入 API Group，回车添加"
                        @keyup.enter="addApiGroup(index)"
                      />
                    </div>
                    <span class="help-text">常见：空（核心）、apps、batch、networking.k8s.io</span>
                  </div>

                  <!-- Resources -->
                  <div class="rbac-form-group">
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
                    <span class="help-text">常见：pods, deployments, services, configmaps, secrets</span>
                  </div>

                  <!-- Verbs -->
                  <div class="rbac-form-group">
                    <label>Verbs（操作） *</label>
                    <div class="verbs-grid">
                      <label v-for="verb in availableVerbs" :key="verb" class="verb-checkbox">
                        <input type="checkbox" :value="verb" v-model="rule.verbs" />
                        <span>{{ verb }}</span>
                      </label>
                    </div>
                    <div class="quick-actions">
                      <button type="button" class="rbac-btn rbac-btn-sm rbac-btn-secondary" @click="selectAllVerbs(index)">
                        全选
                      </button>
                      <button type="button" class="rbac-btn rbac-btn-sm rbac-btn-secondary" @click="selectReadOnlyVerbs(index)">
                        只读权限
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <button type="button" class="rbac-btn rbac-btn-secondary" @click="addRule">
                + 添加规则
              </button>
            </div>
          </form>
        </div>
        <div class="rbac-modal-footer">
          <button class="rbac-btn rbac-btn-secondary" @click="closeCreateModal">取消</button>
          <button class="rbac-btn rbac-btn-primary" @click="submitRole" :disabled="loading">
            {{ loading ? '提交中...' : (editMode ? '更新' : '创建') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 模板选择模态框 -->
    <div v-if="showTemplateModal" class="rbac-modal" @click.self="closeTemplateModal">
      <div class="rbac-modal-content large">
        <div class="rbac-modal-header">
          <h2>选择 Role 模板</h2>
          <button class="rbac-modal-close" @click="closeTemplateModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
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
                <span class="rbac-tag rbac-tag-info">{{ template.rules.length }} 条规则</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 规则详情模态框 -->
    <div v-if="showRulesModal" class="rbac-modal" @click.self="closeRulesModal">
      <div class="rbac-modal-content large">
        <div class="rbac-modal-header">
          <h2>{{ currentRole?.name }} - 权限规则</h2>
          <button class="rbac-modal-close" @click="closeRulesModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <div v-for="(rule, index) in currentRole?.rules" :key="index" class="rule-display">
            <h4>规则 {{ index + 1 }}</h4>
            <div class="rule-detail-grid">
              <div class="rule-detail-item">
                <label>API Groups</label>
                <div class="rule-tags">
                  <span v-for="group in rule.apiGroups" :key="group" class="rbac-tag rbac-tag-primary">
                    {{ group || '""（核心）' }}
                  </span>
                </div>
              </div>
              <div class="rule-detail-item">
                <label>Resources</label>
                <div class="rule-tags">
                  <span v-for="resource in rule.resources" :key="resource" class="rbac-tag rbac-tag-info">
                    {{ resource }}
                  </span>
                </div>
              </div>
              <div class="rule-detail-item">
                <label>Verbs</label>
                <div class="rule-tags">
                  <span v-for="verb in rule.verbs" :key="verb" class="rbac-tag rbac-tag-warning">
                    {{ verb }}
                  </span>
                </div>
              </div>
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
import { listRoles, createRole as createRoleApi, deleteRole as deleteRoleApi } from '@/api/k8sRbac'
import { getClusterList } from '@/api/cluster'
import { getNamespaces } from '@/api/namespace'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// 数据状态
const loading = ref(true)
const searchQuery = ref('')
const roleTypeFilter = ref('')
const namespaceFilter = ref('')
const roles = ref([])
const namespaces = ref([])

// 集群选择
const clusters = ref([])
const selectedClusterId = ref('')

// 统计数据
const roleCount = computed(() => roles.value.filter(r => r.type === 'Role').length)
const clusterRoleCount = computed(() => roles.value.filter(r => r.type === 'ClusterRole').length)
const totalRulesCount = computed(() => roles.value.reduce((sum, r) => sum + (r.rules?.length || 0), 0))

// 监听集群变化
const onClusterChange = (clusterId) => {
  if (clusterId) {
    loadNamespaces()
    loadRoles()
  }
}

watch(selectedClusterId, (val) => {
  if (val) {
    loadNamespaces()
    loadRoles()
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
    Message.error({ content: '加载集群列表失败' })
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
    console.error('加载命名空间失败:', error)
    namespaces.value = ['default', 'kube-system', 'kube-public']
  }
}

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

// Role 模板
const roleTemplates = ref([
  {
    id: 'view',
    name: '只读权限',
    icon: '👁️',
    description: '允许查看所有资源，但不能修改',
    rules: [{ apiGroups: ['', 'apps', 'batch'], resources: ['pods', 'deployments', 'services', 'jobs'], verbs: ['get', 'list', 'watch'] }]
  },
  {
    id: 'edit',
    name: '编辑权限',
    icon: '✏️',
    description: '允许查看和编辑大部分资源',
    rules: [{ apiGroups: ['', 'apps'], resources: ['pods', 'deployments', 'services', 'configmaps'], verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'] }]
  },
  {
    id: 'admin',
    name: '管理员权限',
    icon: '👑',
    description: '完整的管理权限（除 RBAC）',
    rules: [{ apiGroups: ['*'], resources: ['*'], verbs: ['*'] }]
  },
  {
    id: 'pod-reader',
    name: 'Pod 只读',
    icon: '📦',
    description: '只能查看 Pod',
    rules: [{ apiGroups: [''], resources: ['pods', 'pods/log', 'pods/status'], verbs: ['get', 'list', 'watch'] }]
  }
])

// 过滤后的 Roles
const filteredRoles = computed(() => {
  let result = roles.value
  if (roleTypeFilter.value) result = result.filter(r => r.type === roleTypeFilter.value)
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
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await listRoles(selectedClusterId.value, namespaceFilter.value)
    roles.value = res.code === 0 && res.data?.list ? res.data.list : []
  } catch (error) {
    console.error('加载 Role 失败:', error)
    Message.error({ content: '加载失败: ' + (error.msg || error.message || '网络错误') })
    roles.value = []
  } finally {
    loading.value = false
  }
}

// 提交 Role
const submitRole = async () => {
  if (!selectedClusterId.value || roleForm.value.rules.length === 0) {
    Message.warning({ content: '请确保已选择集群并添加至少一条规则' })
    return
  }
  loading.value = true
  try {
    await createRoleApi(selectedClusterId.value, {
      name: roleForm.value.name,
      type: roleForm.value.type,
      namespace: roleForm.value.type === 'Role' ? roleForm.value.namespace : '',
      rules: roleForm.value.rules
    })
    Message.success({ content: editMode.value ? '更新成功' : '创建成功' })
    closeCreateModal()
    await loadRoles()
  } catch (error) {
    Message.error({ content: '提交失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 删除 Role
const deleteRole = async (role) => {
  const ok = await showConfirm({
    title: '确认删除 Role',
    type: 'danger',
    details: [
      { label: '类型', value: role.type },
      { label: '名称', value: role.name, mono: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  loading.value = true
  try {
    await deleteRoleApi(selectedClusterId.value, role.type, role.namespace || '', role.name)
    Message.success({ content: '删除成功' })
    await loadRoles()
  } catch (error) {
    Message.error({ content: '删除失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 规则管理
const addRule = () => {
  roleForm.value.rules.push({ apiGroups: [], resources: [], verbs: [], resourceNames: [] })
}
const removeRule = (index) => roleForm.value.rules.splice(index, 1)

const addApiGroup = (ruleIndex) => {
  const value = apiGroupInput.value[ruleIndex]?.trim()
  if (value !== undefined && !roleForm.value.rules[ruleIndex].apiGroups.includes(value)) {
    roleForm.value.rules[ruleIndex].apiGroups.push(value)
    apiGroupInput.value[ruleIndex] = ''
  }
}
const removeApiGroup = (ruleIndex, groupIndex) => roleForm.value.rules[ruleIndex].apiGroups.splice(groupIndex, 1)

const addResource = (ruleIndex) => {
  const value = resourceInput.value[ruleIndex]?.trim()
  if (value && !roleForm.value.rules[ruleIndex].resources.includes(value)) {
    roleForm.value.rules[ruleIndex].resources.push(value)
    resourceInput.value[ruleIndex] = ''
  }
}
const removeResource = (ruleIndex, resourceIndex) => roleForm.value.rules[ruleIndex].resources.splice(resourceIndex, 1)

const selectAllVerbs = (ruleIndex) => { roleForm.value.rules[ruleIndex].verbs = [...availableVerbs] }
const selectReadOnlyVerbs = (ruleIndex) => { roleForm.value.rules[ruleIndex].verbs = ['get', 'list', 'watch'] }

const viewRules = (role) => { currentRole.value = role; showRulesModal.value = true }
const viewBindings = (role) => { Message.info({ content: `查看 ${role.name} 的绑定（功能开发中）` }) }
const editRole = (role) => { editMode.value = true; roleForm.value = JSON.parse(JSON.stringify(role)); showCreateModal.value = true }
const useTemplate = (template) => {
  roleForm.value.rules = JSON.parse(JSON.stringify(template.rules))
  closeTemplateModal()
  showCreateModal.value = true
  Message.success({ content: `已应用模板：${template.name}` })
}

const openCreateModal = () => {
  editMode.value = false
  roleForm.value = { name: '', type: 'Role', namespace: 'default', rules: [] }
  showCreateModal.value = true
}
const closeCreateModal = () => { showCreateModal.value = false }
const openTemplateModal = () => { showTemplateModal.value = true }
const closeTemplateModal = () => { showTemplateModal.value = false }
const closeRulesModal = () => { showRulesModal.value = false; currentRole.value = null }

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => { loadClusters() })
</script>

<style scoped>
@import '@/assets/styles/rbac-common.css';

/* 资源名称图标 */
.resource-name .icon.cluster {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
}

.resource-name .icon.namespace {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

/* 规则卡片 */
.rule-card {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.rule-header h4 {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
  margin: 0;
}

.rule-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Chip 输入 */
.chip-input {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 10px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  min-height: 44px;
}

.chip-input input {
  flex: 1;
  min-width: 180px;
  background: transparent;
  border: none;
  color: #f3f4f6;
  font-size: 13px;
  outline: none;
}

.chip-input input::placeholder {
  color: #6b7280;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  border-radius: 20px;
  font-size: 12px;
}

.chip button {
  background: none;
  border: none;
  color: #a5b4fc;
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  line-height: 1;
}

.chip button:hover {
  color: #ef4444;
}

/* Verbs 网格 */
.verbs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 10px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
}

.verb-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #d1d5db;
  font-size: 13px;
}

.verb-checkbox input {
  accent-color: #6366f1;
}

.quick-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

/* 模板网格 */
.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}

.template-card {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.template-card:hover {
  border-color: rgba(99, 102, 241, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
}

.template-icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.template-card h3 {
  font-size: 15px;
  font-weight: 600;
  color: #f3f4f6;
  margin: 0 0 8px;
}

.template-card p {
  font-size: 12px;
  color: #9ca3af;
  margin: 0 0 12px;
}

/* 规则详情 */
.rule-display {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}

.rule-display h4 {
  font-size: 13px;
  font-weight: 600;
  color: #e5e7eb;
  margin: 0 0 12px;
}

.rule-detail-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-detail-item label {
  display: block;
  font-size: 11px;
  font-weight: 500;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.rule-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.help-text {
  font-size: 11px;
  color: #6b7280;
  font-weight: normal;
}
</style>

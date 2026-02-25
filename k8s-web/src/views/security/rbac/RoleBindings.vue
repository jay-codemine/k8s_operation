<template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>🔗 RoleBinding 管理</h1>
      <p>绑定用户/ServiceAccount 到 Role，实现权限授予</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 RoleBinding 名称..."
        />
      </div>

      <div class="filter-dropdown">
        <select v-model="bindingTypeFilter">
          <option value="">所有类型</option>
          <option value="RoleBinding">RoleBinding（命名空间级）</option>
          <option value="ClusterRoleBinding">ClusterRoleBinding（集群级）</option>
        </select>
      </div>

      <div class="filter-dropdown" v-if="bindingTypeFilter !== 'ClusterRoleBinding'">
        <select v-model="namespaceFilter" @change="loadBindings">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <button class="btn btn-primary" @click="openCreateModal">+ 创建绑定</button>
        <button class="btn btn-secondary" @click="loadBindings" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 表格视图 -->
    <div class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th style="width: 250px;">绑定名称</th>
            <th style="width: 120px;">类型</th>
            <th style="width: 150px;">命名空间</th>
            <th style="width: 200px;">绑定的 Role</th>
            <th style="width: 100px;">主体数</th>
            <th style="width: 150px;">创建时间</th>
            <th style="width: 200px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="binding in filteredBindings" :key="`${binding.type}-${binding.namespace || 'cluster'}-${binding.name}`">
            <td>
              <div class="resource-name">
                <span class="icon">🔗</span>
                <span>{{ binding.name }}</span>
              </div>
            </td>
            <td>
              <span class="type-badge" :class="binding.type">
                {{ binding.type === 'ClusterRoleBinding' ? '集群级' : '命名空间' }}
              </span>
            </td>
            <td>
              <span v-if="binding.namespace" class="namespace-tag">{{ binding.namespace }}</span>
              <span v-else class="cluster-tag">全集群</span>
            </td>
            <td>
              <span class="role-reference">
                {{ binding.roleRef.kind }}/{{ binding.roleRef.name }}
              </span>
            </td>
            <td>{{ binding.subjects?.length || 0 }}</td>
            <td>{{ formatDate(binding.created_at) }}</td>
            <td class="actions">
              <button class="btn btn-sm btn-info" @click="viewSubjects(binding)" title="查看主体">
                👥 主体
              </button>
              <button class="btn btn-sm btn-warning" @click="editBinding(binding)" title="编辑">
                ✏️ 编辑
              </button>
              <button class="btn btn-sm btn-danger" @click="deleteBinding(binding)" title="删除">
                🗑️ 删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="filteredBindings.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">暂无 RoleBinding，请先创建</div>
      </div>
    </div>

    <!-- 创建 RoleBinding 模态框 -->
    <div v-if="showCreateModal" class="modal" @click.self="closeCreateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>{{ editMode ? '编辑' : '创建' }} RoleBinding</h2>
          <button class="close-btn" @click="closeCreateModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitBinding">
            <!-- 基本信息 -->
            <div class="form-section">
              <h3>基本信息</h3>
              <div class="form-row">
                <div class="form-group">
                  <label>绑定名称 *</label>
                  <input v-model="bindingForm.name" type="text" placeholder="例如：developer-binding" required />
                </div>

                <div class="form-group">
                  <label>类型 *</label>
                  <select v-model="bindingForm.type" required>
                    <option value="RoleBinding">RoleBinding（命名空间级）</option>
                    <option value="ClusterRoleBinding">ClusterRoleBinding（集群级）</option>
                  </select>
                </div>

                <div class="form-group" v-if="bindingForm.type === 'RoleBinding'">
                  <label>命名空间 *</label>
                  <select v-model="bindingForm.namespace" required>
                    <option value="">请选择命名空间</option>
                    <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Role 引用 -->
            <div class="form-section">
              <h3>绑定的 Role</h3>
              <div class="form-row">
                <div class="form-group">
                  <label>Role 类型 *</label>
                  <select v-model="bindingForm.roleRef.kind" required>
                    <option value="Role">Role（命名空间级）</option>
                    <option value="ClusterRole">ClusterRole（集群级）</option>
                  </select>
                </div>

                <div class="form-group">
                  <label>Role 名称 *</label>
                  <select v-model="bindingForm.roleRef.name" required>
                    <option value="">请选择 Role</option>
                    <option v-for="role in availableRoles" :key="role.name" :value="role.name">
                      {{ role.name }}
                    </option>
                  </select>
                  <p class="help-text">可以绑定到任何 Role 或 ClusterRole</p>
                </div>
              </div>
            </div>

            <!-- 主体（Subjects） -->
            <div class="form-section">
              <h3>
                授予权限的主体
                <span class="help-text">（可以是 User、Group 或 ServiceAccount）</span>
              </h3>

              <div v-for="(subject, index) in bindingForm.subjects" :key="index" class="subject-card">
                <div class="subject-header">
                  <h4>主体 {{ index + 1 }}</h4>
                  <button type="button" class="btn btn-sm btn-danger" @click="removeSubject(index)">删除</button>
                </div>

                <div class="subject-body">
                  <div class="form-row">
                    <div class="form-group">
                      <label>类型 *</label>
                      <select v-model="subject.kind" required>
                        <option value="User">User（用户）</option>
                        <option value="Group">Group（用户组）</option>
                        <option value="ServiceAccount">ServiceAccount（服务账户）</option>
                      </select>
                    </div>

                    <div class="form-group">
                      <label>名称 *</label>
                      <input v-model="subject.name" type="text" placeholder="例如：admin 或 default" required />
                    </div>

                    <div class="form-group" v-if="subject.kind === 'ServiceAccount'">
                      <label>命名空间 *</label>
                      <select v-model="subject.namespace" required>
                        <option value="">请选择命名空间</option>
                        <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                      </select>
                    </div>
                  </div>

                  <!-- 快捷选择 ServiceAccount -->
                  <div v-if="subject.kind === 'ServiceAccount'" class="quick-select">
                    <label>或从列表中选择：</label>
                    <select @change="selectServiceAccount(index, $event)">
                      <option value="">请选择...</option>
                      <option v-for="sa in serviceAccounts" :key="sa.name" :value="`${sa.namespace}/${sa.name}`">
                        {{ sa.namespace }}/{{ sa.name }}
                      </option>
                    </select>
                  </div>
                </div>
              </div>

              <button type="button" class="btn btn-secondary" @click="addSubject">+ 添加主体</button>
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

    <!-- 主体详情模态框 -->
    <div v-if="showSubjectsModal" class="modal" @click.self="closeSubjectsModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ currentBinding?.name }} - 授权主体</h2>
          <button class="close-btn" @click="closeSubjectsModal">&times;</button>
        </div>
        <div class="modal-body">
          <table class="detail-table">
            <thead>
              <tr>
                <th>类型</th>
                <th>名称</th>
                <th>命名空间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(subject, index) in currentBinding?.subjects" :key="index">
                <td>
                  <span class="subject-kind-badge" :class="subject.kind">
                    {{ getSubjectKindText(subject.kind) }}
                  </span>
                </td>
                <td>{{ subject.name }}</td>
                <td>{{ subject.namespace || '-' }}</td>
              </tr>
            </tbody>
          </table>
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
const bindingTypeFilter = ref('')
const namespaceFilter = ref('')
const bindings = ref([])
const namespaces = ref(['default', 'kube-system', 'kube-public'])
const availableRoles = ref([])
const serviceAccounts = ref([])

// 模态框状态
const showCreateModal = ref(false)
const showSubjectsModal = ref(false)
const editMode = ref(false)
const currentBinding = ref(null)

// 表单数据
const bindingForm = ref({
  name: '',
  type: 'RoleBinding',
  namespace: 'default',
  roleRef: {
    kind: 'Role',
    name: ''
  },
  subjects: []
})

// 过滤后的绑定
const filteredBindings = computed(() => {
  let result = bindings.value

  if (bindingTypeFilter.value) {
    result = result.filter(b => b.type === bindingTypeFilter.value)
  }

  if (namespaceFilter.value && bindingTypeFilter.value !== 'ClusterRoleBinding') {
    result = result.filter(b => b.namespace === namespaceFilter.value)
  }

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(b => b.name.toLowerCase().includes(query))
  }

  return result
})

// 加载绑定
const loadBindings = async () => {
  loading.value = true
  try {
    // 模拟数据
    bindings.value = [
      {
        name: 'developer-binding',
        type: 'RoleBinding',
        namespace: 'default',
        roleRef: {
          kind: 'Role',
          name: 'developer'
        },
        subjects: [
          {
            kind: 'ServiceAccount',
            name: 'default',
            namespace: 'default'
          },
          {
            kind: 'User',
            name: 'alice'
          }
        ],
        created_at: '2024-02-01T14:20:00Z'
      },
      {
        name: 'admin-binding',
        type: 'ClusterRoleBinding',
        namespace: null,
        roleRef: {
          kind: 'ClusterRole',
          name: 'cluster-admin'
        },
        subjects: [
          {
            kind: 'User',
            name: 'admin'
          }
        ],
        created_at: '2024-01-15T10:30:00Z'
      }
    ]

    // 加载可用的 Roles
    availableRoles.value = [
      { name: 'view', kind: 'ClusterRole' },
      { name: 'edit', kind: 'ClusterRole' },
      { name: 'developer', kind: 'Role' }
    ]

    // 加载 ServiceAccounts
    serviceAccounts.value = [
      { name: 'default', namespace: 'default' },
      { name: 'admin-sa', namespace: 'kube-system' }
    ]
  } catch (error) {
    console.error('加载失败:', error)
    Message.error({ content: '加载失败' })
  } finally {
    loading.value = false
  }
}

// 创建/更新绑定
const submitBinding = async () => {
  // 验证
  if (bindingForm.value.subjects.length === 0) {
    Message.warning({ content: '请至少添加一个主体' })
    return
  }

  loading.value = true
  try {
    // TODO: 调用 API
    Message.success({ content: editMode.value ? '更新成功' : '创建成功' })
    closeCreateModal()
    await loadBindings()
  } catch (error) {
    console.error('提交失败:', error)
    Message.error({ content: '提交失败' })
  } finally {
    loading.value = false
  }
}

// 删除绑定
const deleteBinding = async (binding) => {
  if (!confirm(`确认删除 ${binding.type} "${binding.name}"？\n这将移除相关主体的权限！`)) return

  loading.value = true
  try {
    // TODO: 调用 API
    Message.success({ content: '删除成功' })
    await loadBindings()
  } catch (error) {
    console.error('删除失败:', error)
    Message.error({ content: '删除失败' })
  } finally {
    loading.value = false
  }
}

// 主体管理
const addSubject = () => {
  bindingForm.value.subjects.push({
    kind: 'ServiceAccount',
    name: '',
    namespace: 'default'
  })
}

const removeSubject = (index) => {
  bindingForm.value.subjects.splice(index, 1)
}

const selectServiceAccount = (subjectIndex, event) => {
  const value = event.target.value
  if (!value) return

  const [namespace, name] = value.split('/')
  bindingForm.value.subjects[subjectIndex].namespace = namespace
  bindingForm.value.subjects[subjectIndex].name = name
}

// 查看主体
const viewSubjects = (binding) => {
  currentBinding.value = binding
  showSubjectsModal.value = true
}

// 编辑
const editBinding = (binding) => {
  editMode.value = true
  bindingForm.value = JSON.parse(JSON.stringify(binding))
  showCreateModal.value = true
}

// 模态框管理
const openCreateModal = () => {
  editMode.value = false
  bindingForm.value = {
    name: '',
    type: 'RoleBinding',
    namespace: 'default',
    roleRef: {
      kind: 'Role',
      name: ''
    },
    subjects: []
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const closeSubjectsModal = () => {
  showSubjectsModal.value = false
  currentBinding.value = null
}

// 辅助函数
const getSubjectKindText = (kind) => {
  const map = {
    User: '用户',
    Group: '用户组',
    ServiceAccount: '服务账户'
  }
  return map[kind] || kind
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadBindings()
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

.subject-card {
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #f9f9f9;
}

.subject-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.subject-header h4 {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.subject-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.quick-select {
  padding-top: 12px;
  border-top: 1px dashed #ccc;
}

.quick-select label {
  font-size: 13px;
  color: #666;
  display: block;
  margin-bottom: 8px;
}

.quick-select select {
  width: 100%;
}

.role-reference {
  background-color: #e8f5e9;
  color: #2e7d32;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-family: 'Courier New', monospace;
}

.subject-kind-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.subject-kind-badge.User {
  background-color: #e3f2fd;
  color: #1565c0;
}

.subject-kind-badge.Group {
  background-color: #f3e5f5;
  color: #6a1b9a;
}

.subject-kind-badge.ServiceAccount {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
}

.detail-table th,
.detail-table td {
  padding: 12px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.detail-table th {
  background-color: #f5f5f5;
  font-weight: 600;
}
</style>

<template>
  <div class="rbac-view">
    <!-- 页面头部 -->
    <div class="rbac-header">
      <div class="rbac-header-left">
        <div class="rbac-icon">🔗</div>
        <div class="rbac-title-group">
          <h1>RoleBinding 管理</h1>
          <p>绑定用户/ServiceAccount 到 Role，实现权限授予</p>
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
        <div class="stat-value">{{ bindings.length }}</div>
        <div class="stat-label">总绑定数</div>
      </div>
      <div class="stat-card success">
        <div class="stat-header">
          <div class="stat-icon">📁</div>
        </div>
        <div class="stat-value">{{ roleBindingCount }}</div>
        <div class="stat-label">命名空间级 RoleBinding</div>
      </div>
      <div class="stat-card warning">
        <div class="stat-header">
          <div class="stat-icon">🌍</div>
        </div>
        <div class="stat-value">{{ clusterRoleBindingCount }}</div>
        <div class="stat-label">集群级 ClusterRoleBinding</div>
      </div>
      <div class="stat-card info">
        <div class="stat-header">
          <div class="stat-icon">👥</div>
        </div>
        <div class="stat-value">{{ totalSubjectsCount }}</div>
        <div class="stat-label">授权主体总数</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="rbac-toolbar">
      <div class="toolbar-left">
        <div class="rbac-search">
          <span class="search-icon">🔍</span>
          <input v-model="searchQuery" placeholder="搜索 RoleBinding 名称..." />
        </div>
        <div class="rbac-filter">
          <select v-model="bindingTypeFilter">
            <option value="">所有类型</option>
            <option value="RoleBinding">RoleBinding（命名空间级）</option>
            <option value="ClusterRoleBinding">ClusterRoleBinding（集群级）</option>
          </select>
        </div>
        <div class="rbac-filter" v-if="bindingTypeFilter !== 'ClusterRoleBinding'">
          <select v-model="namespaceFilter" @change="loadBindings">
            <option value="">所有命名空间</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>
      </div>
      <div class="toolbar-right">
        <button class="rbac-btn rbac-btn-secondary" @click="loadBindings" :disabled="loading">
          {{ loading ? '⏳' : '🔄' }} 刷新
        </button>
        <button class="rbac-btn rbac-btn-primary" @click="openCreateModal">
          + 创建绑定
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && bindings.length === 0" class="rbac-loading">
      <div class="rbac-spinner"></div>
      <div class="rbac-loading-text">加载中...</div>
    </div>

    <!-- 表格 -->
    <div v-else class="rbac-table-container">
      <table class="rbac-table">
        <thead>
          <tr>
            <th>绑定名称</th>
            <th>类型</th>
            <th>命名空间</th>
            <th>绑定的 Role</th>
            <th>主体数</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="binding in filteredBindings" :key="`${binding.type}-${binding.namespace || 'cluster'}-${binding.name}`">
            <td>
              <div class="resource-name">
                <div class="icon">🔗</div>
                <div>
                  <div class="name">{{ binding.name }}</div>
                  <div class="meta">{{ binding.subjects?.length || 0 }} 个主体</div>
                </div>
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
                {{ binding.role_ref?.kind || '-' }}/{{ binding.role_ref?.name || '-' }}
              </span>
            </td>
            <td>
              <span class="rbac-tag rbac-tag-info">{{ binding.subjects?.length || 0 }}</span>
            </td>
            <td>{{ formatDate(binding.created_at) }}</td>
            <td>
              <div class="rbac-actions">
                <button class="rbac-action-btn view" @click="viewSubjects(binding)" title="查看主体">👥</button>
                <button class="rbac-action-btn edit" @click="editBinding(binding)" title="编辑">✏️</button>
                <button class="rbac-action-btn delete" @click="deleteBinding(binding)" title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 空状态 -->
      <div v-if="filteredBindings.length === 0" class="rbac-empty">
        <div class="rbac-empty-icon">📦</div>
        <div class="rbac-empty-title">暂无 RoleBinding</div>
        <div class="rbac-empty-desc">{{ searchQuery ? '没有匹配的结果' : '点击上方按钮创建第一个绑定' }}</div>
        <button v-if="!searchQuery" class="rbac-btn rbac-btn-primary" @click="openCreateModal">
          + 创建绑定
        </button>
      </div>
    </div>

    <!-- 创建 RoleBinding 模态框 -->
    <div v-if="showCreateModal" class="rbac-modal" @click.self="closeCreateModal">
      <div class="rbac-modal-content large">
        <div class="rbac-modal-header">
          <h2>{{ editMode ? '编辑' : '创建' }} RoleBinding</h2>
          <button class="rbac-modal-close" @click="closeCreateModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <form @submit.prevent="submitBinding">
            <!-- 基本信息 -->
            <div class="rbac-form-section">
              <h3>基本信息</h3>
              <div class="rbac-form-row">
                <div class="rbac-form-group">
                  <label>绑定名称 *</label>
                  <input v-model="bindingForm.name" placeholder="例如：developer-binding" required />
                </div>
                <div class="rbac-form-group">
                  <label>类型 *</label>
                  <select v-model="bindingForm.type" required>
                    <option value="RoleBinding">RoleBinding（命名空间级）</option>
                    <option value="ClusterRoleBinding">ClusterRoleBinding（集群级）</option>
                  </select>
                </div>
                <div class="rbac-form-group" v-if="bindingForm.type === 'RoleBinding'">
                  <label>命名空间 *</label>
                  <select v-model="bindingForm.namespace" required>
                    <option value="">请选择命名空间</option>
                    <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Role 引用 -->
            <div class="rbac-form-section">
              <h3>绑定的 Role</h3>
              <div class="rbac-form-row">
                <div class="rbac-form-group">
                  <label>Role 类型 *</label>
                  <select v-model="bindingForm.roleRef.kind" required>
                    <option value="Role">Role（命名空间级）</option>
                    <option value="ClusterRole">ClusterRole（集群级）</option>
                  </select>
                </div>
                <div class="rbac-form-group">
                  <label>Role 名称 *</label>
                  <select v-model="bindingForm.roleRef.name" required>
                    <option value="">请选择 Role</option>
                    <option v-for="role in availableRoles" :key="role.name" :value="role.name">
                      {{ role.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 主体 -->
            <div class="rbac-form-section">
              <h3>授予权限的主体 <span class="help-text">（可以是 User、Group 或 ServiceAccount）</span></h3>
              
              <div v-for="(subject, index) in bindingForm.subjects" :key="index" class="subject-card">
                <div class="subject-header">
                  <h4>主体 {{ index + 1 }}</h4>
                  <button type="button" class="rbac-btn rbac-btn-sm rbac-btn-danger" @click="removeSubject(index)">
                    删除
                  </button>
                </div>
                <div class="subject-body">
                  <div class="rbac-form-row">
                    <div class="rbac-form-group">
                      <label>类型 *</label>
                      <select v-model="subject.kind" required>
                        <option value="User">User（用户）</option>
                        <option value="Group">Group（用户组）</option>
                        <option value="ServiceAccount">ServiceAccount（服务账户）</option>
                      </select>
                    </div>
                    <div class="rbac-form-group">
                      <label>名称 *</label>
                      <input v-model="subject.name" placeholder="例如：admin 或 default" required />
                    </div>
                    <div class="rbac-form-group" v-if="subject.kind === 'ServiceAccount'">
                      <label>命名空间 *</label>
                      <select v-model="subject.namespace" required>
                        <option value="">请选择命名空间</option>
                        <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>

              <button type="button" class="rbac-btn rbac-btn-secondary" @click="addSubject">
                + 添加主体
              </button>
            </div>
          </form>
        </div>
        <div class="rbac-modal-footer">
          <button class="rbac-btn rbac-btn-secondary" @click="closeCreateModal">取消</button>
          <button class="rbac-btn rbac-btn-primary" @click="submitBinding" :disabled="loading">
            {{ loading ? '提交中...' : (editMode ? '更新' : '创建') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 主体详情模态框 -->
    <div v-if="showSubjectsModal" class="rbac-modal" @click.self="closeSubjectsModal">
      <div class="rbac-modal-content">
        <div class="rbac-modal-header">
          <h2>{{ currentBinding?.name }} - 授权主体</h2>
          <button class="rbac-modal-close" @click="closeSubjectsModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <div class="subjects-list">
            <div v-for="(subject, index) in currentBinding?.subjects" :key="index" class="subject-item">
              <div class="subject-icon" :class="subject.kind">
                {{ getSubjectIcon(subject.kind) }}
              </div>
              <div class="subject-info">
                <div class="subject-name">{{ subject.name }}</div>
                <div class="subject-meta">
                  <span class="subject-kind-badge" :class="subject.kind">
                    {{ getSubjectKindText(subject.kind) }}
                  </span>
                  <span v-if="subject.namespace" class="namespace-tag">{{ subject.namespace }}</span>
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
import { listRoleBindings, createRoleBinding as createBindingApi, deleteRoleBinding as deleteBindingApi, listRoles } from '@/api/k8sRbac'
import { getClusterList } from '@/api/cluster'
import { getNamespaces } from '@/api/namespace'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// 数据状态
const loading = ref(true)
const searchQuery = ref('')
const bindingTypeFilter = ref('')
const namespaceFilter = ref('')
const bindings = ref([])
const namespaces = ref([])
const availableRoles = ref([])

// 集群选择
const clusters = ref([])
const selectedClusterId = ref('')

// 统计数据
const roleBindingCount = computed(() => bindings.value.filter(b => b.type === 'RoleBinding').length)
const clusterRoleBindingCount = computed(() => bindings.value.filter(b => b.type === 'ClusterRoleBinding').length)
const totalSubjectsCount = computed(() => bindings.value.reduce((sum, b) => sum + (b.subjects?.length || 0), 0))

// 监听集群变化
const onClusterChange = (clusterId) => {
  if (clusterId) {
    loadNamespaces()
    loadBindings()
  }
}

watch(selectedClusterId, (val) => {
  if (val) {
    loadNamespaces()
    loadBindings()
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
  roleRef: { kind: 'Role', name: '' },
  subjects: []
})

// 过滤后的绑定
const filteredBindings = computed(() => {
  let result = bindings.value
  if (bindingTypeFilter.value) result = result.filter(b => b.type === bindingTypeFilter.value)
  if (namespaceFilter.value && bindingTypeFilter.value !== 'ClusterRoleBinding') {
    result = result.filter(b => b.namespace === namespaceFilter.value)
  }
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(b => b.name.toLowerCase().includes(query))
  }
  return result
})

// 加载绑定（并行加载优化）
const loadBindings = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    // 并行加载 RoleBindings 和 Roles
    const [bindingsRes, rolesRes] = await Promise.all([
      listRoleBindings(selectedClusterId.value, namespaceFilter.value),
      listRoles(selectedClusterId.value, '')
    ])
    
    bindings.value = bindingsRes.code === 0 && bindingsRes.data?.list ? bindingsRes.data.list : []
    
    if (rolesRes.code === 0 && rolesRes.data?.list) {
      availableRoles.value = rolesRes.data.list.map(r => ({ name: r.name, kind: r.type }))
    }
  } catch (error) {
    console.error('加载 RoleBinding 失败:', error)
    Message.error({ content: '加载失败: ' + (error.msg || error.message || '网络错误') })
    bindings.value = []
  } finally {
    loading.value = false
  }
}

// 提交绑定
const submitBinding = async () => {
  if (!selectedClusterId.value || bindingForm.value.subjects.length === 0) {
    Message.warning({ content: '请确保已选择集群并添加至少一个主体' })
    return
  }
  loading.value = true
  try {
    await createBindingApi(selectedClusterId.value, {
      name: bindingForm.value.name,
      type: bindingForm.value.type,
      namespace: bindingForm.value.type === 'RoleBinding' ? bindingForm.value.namespace : '',
      role_ref: bindingForm.value.roleRef,
      subjects: bindingForm.value.subjects
    })
    Message.success({ content: editMode.value ? '更新成功' : '创建成功' })
    closeCreateModal()
    await loadBindings()
  } catch (error) {
    Message.error({ content: '提交失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 删除绑定
const deleteBinding = async (binding) => {
  const ok = await showConfirm({
    title: '确认删除角色绑定',
    content: '删除后相关主体将失去对应权限。',
    type: 'danger',
    details: [
      { label: '类型', value: binding.type },
      { label: '名称', value: binding.name, mono: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  loading.value = true
  try {
    await deleteBindingApi(selectedClusterId.value, binding.type, binding.namespace || '', binding.name)
    Message.success({ content: '删除成功' })
    await loadBindings()
  } catch (error) {
    Message.error({ content: '删除失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 主体管理
const addSubject = () => { bindingForm.value.subjects.push({ kind: 'ServiceAccount', name: '', namespace: 'default' }) }
const removeSubject = (index) => { bindingForm.value.subjects.splice(index, 1) }
const viewSubjects = (binding) => { currentBinding.value = binding; showSubjectsModal.value = true }
const editBinding = (binding) => { editMode.value = true; bindingForm.value = JSON.parse(JSON.stringify(binding)); showCreateModal.value = true }

const openCreateModal = () => {
  editMode.value = false
  bindingForm.value = { name: '', type: 'RoleBinding', namespace: 'default', roleRef: { kind: 'Role', name: '' }, subjects: [] }
  showCreateModal.value = true
}
const closeCreateModal = () => { showCreateModal.value = false }
const closeSubjectsModal = () => { showSubjectsModal.value = false; currentBinding.value = null }

const getSubjectKindText = (kind) => ({ User: '用户', Group: '用户组', ServiceAccount: '服务账户' }[kind] || kind)
const getSubjectIcon = (kind) => ({ User: '👤', Group: '👥', ServiceAccount: '🤖' }[kind] || '❓')
const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

onMounted(() => { loadClusters() })
</script>

<style scoped>
@import '@/assets/styles/rbac-common.css';

/* Role 引用标签 */
.role-reference {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background: rgba(16, 185, 129, 0.12);
  color: #6ee7b7;
  font-size: 12px;
  font-family: 'Fira Code', monospace;
  border-radius: 4px;
  border: 1px solid rgba(16, 185, 129, 0.2);
}

/* 主体卡片 */
.subject-card {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}

.subject-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.subject-header h4 {
  font-size: 14px;
  font-weight: 600;
  color: #e5e7eb;
  margin: 0;
}

/* 主体列表 */
.subjects-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.subject-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
}

.subject-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 18px;
}

.subject-icon.User { background: rgba(59, 130, 246, 0.15); }
.subject-icon.Group { background: rgba(139, 92, 246, 0.15); }
.subject-icon.ServiceAccount { background: rgba(16, 185, 129, 0.15); }

.subject-info { flex: 1; }
.subject-name { font-weight: 500; color: #f3f4f6; margin-bottom: 4px; }
.subject-meta { display: flex; gap: 8px; align-items: center; }

.subject-kind-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.subject-kind-badge.User { background: rgba(59, 130, 246, 0.15); color: #93c5fd; }
.subject-kind-badge.Group { background: rgba(139, 92, 246, 0.15); color: #c4b5fd; }
.subject-kind-badge.ServiceAccount { background: rgba(16, 185, 129, 0.15); color: #6ee7b7; }

.help-text {
  font-size: 11px;
  color: #6b7280;
  font-weight: normal;
}
</style>

<template>
  <div class="security-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <div class="header-content">
        <div class="header-icon">🔐</div>
        <div class="header-text">
          <h1>授权管理</h1>
          <p>管理用户、组、ServiceAccount 到角色的绑定关系</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-card">
          <div class="stat-value">{{ authorizationCount }}</div>
          <div class="stat-label">授权总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ clusters.length }}</div>
          <div class="stat-label">管理集群</div>
        </div>
      </div>
    </div>

    <!-- 视角切换 -->
    <div class="view-switch">
      <button :class="['view-btn', { active: viewMode === 'user' }]" @click="viewMode = 'user'">
        <span class="view-icon">👤</span> 按用户
      </button>
      <button :class="['view-btn', { active: viewMode === 'role' }]" @click="viewMode = 'role'">
        <span class="view-icon">🎭</span> 按角色
      </button>
      <button :class="['view-btn', { active: viewMode === 'cluster' }]" @click="viewMode = 'cluster'">
        <span class="view-icon">☸️</span> 按集群
      </button>
    </div>

    <!-- 按用户视角 -->
    <div v-show="viewMode === 'user'" class="view-content">
      <div class="content-header">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input v-model="userSearch" placeholder="搜索用户..." />
        </div>
        <button class="btn btn-primary" @click="openAuthModal('user')">
          ➕ 授权用户
        </button>
      </div>

      <div class="auth-list">
        <div v-for="user in filteredUserAuths" :key="user.id" class="auth-card">
          <div class="auth-main">
            <div class="auth-avatar">{{ user.username?.charAt(0)?.toUpperCase() }}</div>
            <div class="auth-info">
              <span class="auth-name">{{ user.username }}</span>
              <span class="auth-email">{{ user.email || '-' }}</span>
            </div>
          </div>
          <div class="auth-roles">
            <span v-for="role in user.roles" :key="role.id" class="role-chip" :style="{ backgroundColor: role.color + '20', color: role.color }">
              {{ role.display_name || role.name }}
            </span>
          </div>
          <div class="auth-clusters">
            <span class="cluster-count">{{ user.cluster_permissions?.length || 0 }} 个集群</span>
          </div>
          <div class="auth-actions">
            <button class="btn-text" @click="editUserAuth(user)">编辑</button>
            <button class="btn-text danger" @click="revokeUserAuth(user)">撤销</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 按角色视角 -->
    <div v-show="viewMode === 'role'" class="view-content">
      <div class="content-header">
        <select v-model="selectedRole" class="filter-select">
          <option value="">选择角色</option>
          <option v-for="role in allRoles" :key="role.id" :value="role.id">
            {{ role.display_name || role.name }}
          </option>
        </select>
        <button class="btn btn-secondary" @click="loadRoleBindings">🔄 刷新</button>
      </div>

      <div v-if="!selectedRole" class="empty-state">
        <div class="empty-icon">🎭</div>
        <p>请选择角色查看绑定对象</p>
      </div>

      <div v-else-if="roleBindings.length === 0" class="empty-state">
        <div class="empty-icon">👥</div>
        <p>该角色暂无绑定用户</p>
      </div>

      <div v-else class="binding-list">
        <div v-for="binding in roleBindings" :key="binding.id" class="binding-card">
          <div class="binding-avatar">{{ binding.subject_name?.charAt(0)?.toUpperCase() }}</div>
          <div class="binding-type">
            <span :class="['type-badge', binding.subject_type]">
              {{ getSubjectTypeLabel(binding.subject_type) }}
            </span>
          </div>
          <div class="binding-info">
            <span class="binding-name">{{ binding.subject_name }}</span>
            <span class="binding-email" v-if="binding.email">{{ binding.email }}</span>
          </div>
          <span :class="['status-badge', binding.status === 1 ? 'active' : 'inactive']">
            {{ binding.status === 1 ? '激活' : '禁用' }}
          </span>
          <div class="binding-actions">
            <button class="btn-text danger" @click="removeBinding(binding)">移除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 按集群视角 -->
    <div v-show="viewMode === 'cluster'" class="view-content">
      <div class="content-header">
        <select v-model="selectedCluster" class="filter-select" @change="loadClusterAuths">
          <option value="">选择集群</option>
          <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name }}</option>
        </select>
        <button class="btn btn-secondary" @click="loadClusterAuths">🔄 刷新</button>
      </div>

      <div v-if="!selectedCluster" class="empty-state">
        <div class="empty-icon">☸️</div>
        <p>请选择集群查看授权用户</p>
      </div>

      <div v-else class="cluster-auth-grid">
        <div class="auth-section">
          <div class="section-title">
            <span class="section-icon">👥</span> 用户授权
          </div>
          <div class="auth-items">
            <div v-for="auth in clusterUserAuths" :key="auth.id" class="auth-item">
              <div class="item-main">
                <span class="item-name">{{ auth.username }}</span>
                <span class="item-role">{{ auth.role_type }}</span>
              </div>
              <div class="item-perms">
                <span v-if="auth.can_view" class="perm view">查看</span>
                <span v-if="auth.can_create" class="perm create">创建</span>
                <span v-if="auth.can_update" class="perm update">更新</span>
                <span v-if="auth.can_delete" class="perm delete">删除</span>
                <span v-if="auth.can_exec" class="perm exec">执行</span>
              </div>
            </div>
          </div>
        </div>

        <div class="auth-section">
          <div class="section-title">
            <span class="section-icon">🤖</span> ServiceAccount 授权
          </div>
          <div class="auth-items">
            <div v-for="sa in clusterSaAuths" :key="sa.name" class="auth-item">
              <div class="item-main">
                <span class="item-name">{{ sa.name }}</span>
                <span class="item-ns">{{ sa.namespace }}</span>
              </div>
              <button class="btn-text" @click="viewSaBindings(sa)">查看绑定</button>
            </div>
            <div v-if="!clusterSaAuths.length" class="no-data">暂无 ServiceAccount 授权</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 授权弹窗 -->
    <div v-if="showAuthModal" class="modal-mask" @click="showAuthModal = false">
      <div class="modal modal-lg" @click.stop>
        <div class="modal-header">
          <h3>{{ authModalMode === 'user' ? '用户授权' : '角色绑定' }}</h3>
          <button class="btn-close" @click="showAuthModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>选择用户</label>
            <select v-model="authForm.user_id">
              <option value="">请选择</option>
              <option v-for="user in allUsers" :key="user.id" :value="user.id">
                {{ user.username }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>分配角色</label>
            <div class="role-checkboxes">
              <label v-for="role in allRoles" :key="role.id" class="checkbox-item">
                <input type="checkbox" :value="role.id" v-model="authForm.role_ids" />
                <span class="role-dot" :style="{ backgroundColor: role.color }"></span>
                {{ role.display_name || role.name }}
              </label>
            </div>
          </div>
          <div class="form-group">
            <label>集群权限</label>
            <div class="cluster-auth-form">
              <div v-for="cluster in clusters" :key="cluster.id" class="cluster-perm-row">
                <input type="checkbox" :value="cluster.id" v-model="authForm.cluster_ids" />
                <span class="cluster-name">{{ cluster.cluster_name }}</span>
                <div class="perm-switches" v-if="authForm.cluster_ids.includes(cluster.id)">
                  <label><input type="checkbox" v-model="authForm.perms[cluster.id].can_view" /> 查看</label>
                  <label><input type="checkbox" v-model="authForm.perms[cluster.id].can_create" /> 创建</label>
                  <label><input type="checkbox" v-model="authForm.perms[cluster.id].can_update" /> 更新</label>
                  <label><input type="checkbox" v-model="authForm.perms[cluster.id].can_delete" /> 删除</label>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showAuthModal = false">取消</button>
          <button class="btn btn-primary" @click="saveAuth" :disabled="submitting">
            {{ submitting ? '保存中...' : '确认授权' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getUserList } from '@/api/user'
import { getAllRoles, getClusterPermissionList, assignUserRole, createClusterPermission, deleteClusterPermission, getRoleUsers, getUserRBACInfo } from '@/api/rbac'
import { getClusterList } from '@/api/cluster'
import { listServiceAccounts, listRoleBindings as listK8sRoleBindings } from '@/api/k8sRbac'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

const viewMode = ref('user')
const clusters = ref([])
const allUsers = ref([])
const allRoles = ref([])
const loading = ref(false)
const submitting = ref(false)

// 用户视角
const userSearch = ref('')
const userAuthorizations = ref([])

// 角色视角
const selectedRole = ref('')
const roleBindings = ref([])
const roleLoading = ref(false)

// 集群视角
const selectedCluster = ref('')
const clusterUserAuths = ref([])
const clusterSaAuths = ref([])
const clusterLoading = ref(false)

// 弹窗
const showAuthModal = ref(false)
const authModalMode = ref('user')
const authForm = ref({
  user_id: '',
  role_ids: [],
  cluster_ids: [],
  perms: {}
})

const authorizationCount = computed(() => userAuthorizations.value.length)

const filteredUserAuths = computed(() => {
  if (!userSearch.value) return userAuthorizations.value
  const q = userSearch.value.toLowerCase()
  return userAuthorizations.value.filter(u =>
    u.username?.toLowerCase().includes(q) ||
    u.email?.toLowerCase().includes(q)
  )
})

const loadUsers = async () => {
  try {
    const res = await getUserList({ page: 1, limit: 1000 })
    if (res.code === 0) {
      const users = res.data?.list || res.data || []
      allUsers.value = users
      // 加载每个用户的角色和集群权限
      const usersWithAuth = await Promise.all(users.map(async (user) => {
        try {
          const rbacRes = await getUserRBACInfo(user.id)
          if (rbacRes.code === 0 && rbacRes.data) {
            return { ...user, ...rbacRes.data }
          }
        } catch (e) {
          // 忽略单个用户的加载错误
        }
        return user
      }))
      userAuthorizations.value = usersWithAuth
    }
  } catch (e) {
    console.error('加载用户失败', e)
  }
}

const loadRoles = async () => {
  try {
    const res = await getAllRoles()
    if (res.code === 0) {
      // 后端返回 {list: [...], total: x} 或直接数组
      allRoles.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载角色失败', e)
  }
}

const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0) {
      clusters.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载集群失败', e)
  }
}

const loadRoleBindings = async () => {
  if (!selectedRole.value) return
  try {
    const res = await getRoleUsers(selectedRole.value)
    if (res.code === 0) {
      const users = res.data?.list || res.data || []
      // 转换为binding格式
      roleBindings.value = users.map(u => ({
        id: u.id,
        subject_type: 'user',
        subject_name: u.username,
        email: u.email,
        status: u.status
      }))
    }
  } catch (e) {
    console.error('加载角色绑定失败', e)
    roleBindings.value = []
  }
}

const loadClusterAuths = async () => {
  if (!selectedCluster.value) return
  try {
    const res = await getClusterPermissionList({ cluster_id: selectedCluster.value, page: 1, limit: 100 })
    if (res.code === 0) {
      clusterUserAuths.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载集群授权失败', e)
    clusterUserAuths.value = []
  }

  // 加载 ServiceAccount
  try {
    const res = await listServiceAccounts(selectedCluster.value, '')
    if (res.code === 0) {
      clusterSaAuths.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载SA失败', e)
    clusterSaAuths.value = []
  }
}

const openAuthModal = (mode) => {
  authModalMode.value = mode
  authForm.value = {
    user_id: '',
    role_ids: [],
    cluster_ids: [],
    perms: {}
  }
  // 初始化集群权限表单
  clusters.value.forEach(c => {
    authForm.value.perms[c.id] = {
      can_view: true,
      can_create: false,
      can_update: false,
      can_delete: false,
      can_exec: false
    }
  })
  showAuthModal.value = true
}

const saveAuth = async () => {
  if (!authForm.value.user_id) {
    alert('请选择用户')
    return
  }
  submitting.value = true
  try {
    // 分配角色
    await assignUserRole({
      user_id: Number(authForm.value.user_id),
      role_ids: authForm.value.role_ids.map(Number)
    })
    // 分配集群权限
    for (const clusterId of authForm.value.cluster_ids) {
      await createClusterPermission({
        user_id: Number(authForm.value.user_id),
        cluster_id: Number(clusterId),
        role_type: 'custom',
        ...authForm.value.perms[clusterId]
      })
    }
    showAuthModal.value = false
    loadUsers()
  } catch (e) {
    console.error('授权失败', e)
    alert('授权失败: ' + (e.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const editUserAuth = (user) => {
  authModalMode.value = 'user'
  authForm.value = {
    user_id: user.id,
    role_ids: user.roles?.map(r => r.id) || [],
    cluster_ids: user.cluster_permissions?.map(p => p.cluster_id) || [],
    perms: {}
  }
  clusters.value.forEach(c => {
    const perm = user.cluster_permissions?.find(p => p.cluster_id === c.id)
    authForm.value.perms[c.id] = perm ? { 
      can_view: perm.can_view,
      can_create: perm.can_create,
      can_update: perm.can_update,
      can_delete: perm.can_delete,
      can_exec: perm.can_exec
    } : {
      can_view: true,
      can_create: false,
      can_update: false,
      can_delete: false,
      can_exec: false
    }
  })
  showAuthModal.value = true
}

const revokeUserAuth = async (user) => {
  const ok = await showConfirm({
    title: '确认撤销授权',
    content: '将撤销该用户的所有角色和集群权限。',
    type: 'danger',
    details: [
      { label: '用户名', value: user.username },
    ],
    confirmText: '确认撤销',
    cancelText: '取消',
  })
  if (!ok) return
  try {
    // 清空角色
    await assignUserRole({
      user_id: Number(user.id),
      role_ids: []
    })
    // 删除所有集群权限
    if (user.cluster_permissions?.length) {
      for (const perm of user.cluster_permissions) {
        await deleteClusterPermission(perm.id)
      }
    }
    loadUsers()
  } catch (e) {
    console.error('撤销失败', e)
    alert('撤销失败')
  }
}

const removeBinding = async (binding) => {
  const ok = await showConfirm({
    title: '确认移除角色绑定',
    content: '移除后该用户将失去对应角色权限。',
    type: 'danger',
    details: [
      { label: '用户名', value: binding.subject_name },
    ],
    confirmText: '确认移除',
    cancelText: '取消',
  })
  if (!ok) return
  try {
    // 获取用户当前角色，移除当前角色
    const user = userAuthorizations.value.find(u => u.id === binding.id)
    const newRoleIds = (user?.roles || []).filter(r => r.id !== Number(selectedRole.value)).map(r => r.id)
    await assignUserRole({
      user_id: Number(binding.id),
      role_ids: newRoleIds
    })
    loadRoleBindings()
    loadUsers()
  } catch (e) {
    console.error('移除失败', e)
    alert('移除失败')
  }
}

const viewSaBindings = (sa) => {
  alert(`ServiceAccount: ${sa.metadata?.name || sa.name}\n命名空间: ${sa.metadata?.namespace || sa.namespace}`)
}

const getSubjectTypeLabel = (type) => {
  const labels = {
    user: '用户',
    group: '用户组',
    serviceaccount: 'ServiceAccount'
  }
  return labels[type] || type
}

// 监听角色选择变化
watch(() => selectedRole.value, (newVal) => {
  if (newVal) {
    loadRoleBindings()
  } else {
    roleBindings.value = []
  }
})

// 监听集群权限表单变化
watch(() => authForm.value.cluster_ids, (ids) => {
  ids.forEach(id => {
    if (!authForm.value.perms[id]) {
      authForm.value.perms[id] = {
        can_view: true,
        can_create: false,
        can_update: false,
        can_delete: false
      }
    }
  })
}, { deep: true })

onMounted(() => {
  loadUsers()
  loadRoles()
  loadClusters()
})
</script>

<style scoped>
.security-view {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #1e3a5f 0%, #2c5282 100%);
  border-radius: 12px;
  color: white;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  font-size: 36px;
}

.header-text h1 {
  margin: 0;
  font-size: 22px;
}

.header-text p {
  margin: 4px 0 0;
  opacity: 0.8;
  font-size: 14px;
}

.header-stats {
  display: flex;
  gap: 16px;
}

.stat-card {
  background: rgba(255, 255, 255, 0.15);
  padding: 12px 20px;
  border-radius: 8px;
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
}

.stat-label {
  font-size: 12px;
  opacity: 0.8;
}

.view-switch {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  padding: 4px;
  background: white;
  border-radius: 8px;
  width: fit-content;
}

.view-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #64748b;
  transition: all 0.2s;
}

.view-btn.active {
  background: #3b82f6;
  color: white;
}

.view-content {
  background: white;
  border-radius: 12px;
  padding: 20px;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: #f8fafc;
}

.search-box input {
  border: none;
  background: transparent;
  outline: none;
  width: 200px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  min-width: 180px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.auth-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.auth-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: all 0.2s;
}

.auth-card:hover {
  border-color: #94a3b8;
}

.auth-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 180px;
}

.auth-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.auth-info {
  display: flex;
  flex-direction: column;
}

.auth-name {
  font-weight: 600;
  font-size: 14px;
}

.auth-email {
  font-size: 12px;
  color: #64748b;
}

.auth-roles {
  flex: 1;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.role-chip {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
}

.auth-clusters {
  min-width: 80px;
}

.cluster-count {
  font-size: 13px;
  color: #64748b;
}

.auth-actions {
  display: flex;
  gap: 8px;
}

.btn-text {
  background: none;
  border: none;
  color: #3b82f6;
  cursor: pointer;
  font-size: 13px;
}

.btn-text.danger {
  color: #ef4444;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.binding-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.binding-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.type-badge {
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.type-badge.user {
  background: #dbeafe;
  color: #1e40af;
}

.type-badge.group {
  background: #dcfce7;
  color: #166534;
}

.type-badge.serviceaccount {
  background: #fef3c7;
  color: #92400e;
}

.binding-info {
  flex: 1;
}

.binding-name {
  font-weight: 500;
}

.binding-scope {
  font-size: 12px;
  color: #64748b;
  margin-left: 12px;
}

.cluster-auth-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.auth-section {
  background: #f8fafc;
  border-radius: 10px;
  padding: 16px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e2e8f0;
}

.auth-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.auth-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: white;
  border-radius: 6px;
}

.item-main {
  display: flex;
  align-items: center;
  gap: 10px;
}

.item-name {
  font-weight: 500;
  font-size: 14px;
}

.item-role, .item-ns {
  font-size: 12px;
  color: #64748b;
  padding: 2px 6px;
  background: #e2e8f0;
  border-radius: 3px;
}

.item-perms {
  display: flex;
  gap: 4px;
}

.perm {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.perm.view { background: #dbeafe; color: #1e40af; }
.perm.create { background: #dcfce7; color: #166534; }
.perm.update { background: #fef3c7; color: #92400e; }
.perm.delete { background: #fee2e2; color: #991b1b; }
.perm.exec { background: #f3e8ff; color: #7c3aed; }

.no-data {
  color: #94a3b8;
  font-size: 13px;
  text-align: center;
  padding: 20px;
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal {
  width: 480px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.2);
}

.modal.modal-lg {
  width: 640px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #94a3b8;
}

.modal-body {
  padding: 24px;
  max-height: 60vh;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  font-size: 14px;
}

.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.role-checkboxes {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 6px;
  cursor: pointer;
}

.checkbox-item:hover {
  background: #f1f5f9;
}

.role-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.cluster-auth-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cluster-perm-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 6px;
}

.cluster-name {
  min-width: 120px;
  font-weight: 500;
}

.perm-switches {
  display: flex;
  gap: 12px;
}

.perm-switches label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  cursor: pointer;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

.binding-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
}

.binding-email {
  font-size: 12px;
  color: #64748b;
  margin-left: 8px;
}

.status-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
}

.status-badge.active {
  background: #dcfce7;
  color: #166534;
}

.status-badge.inactive {
  background: #fee2e2;
  color: #991b1b;
}
</style>

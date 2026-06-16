<template>
  <div class="security-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <div class="header-content">
        <div class="header-icon">👥</div>
        <div class="header-text">
          <h1>用户管理</h1>
          <p>管理平台用户、分配角色、查看授权范围</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-card">
          <div class="stat-value">{{ users.length }}</div>
          <div class="stat-label">用户总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ activeUsers }}</div>
          <div class="stat-label">活跃用户</div>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input v-model="searchQuery" placeholder="搜索用户名、邮箱..." />
        </div>
        <select v-model="statusFilter" class="filter-select">
          <option value="">全部状态</option>
          <option value="active">正常</option>
          <option value="disabled">已禁用</option>
        </select>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-secondary" @click="loadUsers">
          🔄 刷新
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          ➕ 创建用户
        </button>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="data-table">
      <table>
        <thead>
          <tr>
            <th>用户名</th>
            <th>邮箱/工号</th>
            <th>状态</th>
            <th>所属角色</th>
            <th>最近登录</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.id">
            <td>
              <div class="user-cell">
                <div class="user-avatar">{{ user.username?.charAt(0)?.toUpperCase() }}</div>
                <span>{{ user.username }}</span>
              </div>
            </td>
            <td>{{ user.email || user.employee_id || '-' }}</td>
            <td>
              <span :class="['status-tag', user.status === 1 ? 'active' : 'disabled']">
                {{ user.status === 1 ? '正常' : '已禁用' }}
              </span>
            </td>
            <td>
              <div class="role-tags">
                <span v-for="role in user.roles" :key="role.id" class="role-tag" :style="{ backgroundColor: role.color + '20', color: role.color }">
                  {{ role.display_name || role.name }}
                </span>
                <span v-if="!user.roles?.length" class="no-role">未分配</span>
              </div>
            </td>
            <td>{{ formatTime(user.last_login_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn-text" @click="viewUserDetail(user)">查看</button>
                <button class="btn-text" @click="openRoleModal(user)">分配角色</button>
                <button class="btn-text" @click="toggleUserStatus(user)">
                  {{ user.status === 1 ? '禁用' : '启用' }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 用户详情抽屉 -->
    <div v-if="showDetailDrawer" class="drawer-mask" @click="showDetailDrawer = false">
      <div class="drawer" @click.stop>
        <div class="drawer-header">
          <h3>用户详情</h3>
          <button class="btn-close" @click="showDetailDrawer = false">✕</button>
        </div>
        <div class="drawer-body" v-if="selectedUser">
          <div class="detail-section">
            <div class="detail-title">基本信息</div>
            <div class="detail-item">
              <label>用户名</label>
              <span>{{ selectedUser.username }}</span>
            </div>
            <div class="detail-item">
              <label>邮箱</label>
              <span>{{ selectedUser.email || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>状态</label>
              <span :class="['status-tag', selectedUser.status === 1 ? 'active' : 'disabled']">
                {{ selectedUser.status === 1 ? '正常' : '已禁用' }}
              </span>
            </div>
          </div>
          <div class="detail-section">
            <div class="detail-title">角色与权限</div>
            <div class="role-list">
              <div v-for="role in selectedUser.roles" :key="role.id" class="role-card">
                <div class="role-dot" :style="{ backgroundColor: role.color }"></div>
                <div class="role-info">
                  <span class="role-name">{{ role.display_name || role.name }}</span>
                  <span class="role-type">{{ getRoleTypeLabel(role.role_type) }}</span>
                </div>
              </div>
              <div v-if="!selectedUser.roles?.length" class="no-data">暂未分配角色</div>
            </div>
          </div>
          <div class="detail-section">
            <div class="detail-title">集群授权</div>
            <div class="auth-list">
              <div v-for="perm in selectedUser.cluster_permissions" :key="perm.cluster_id" class="auth-card">
                <span class="cluster-name">{{ perm.cluster_name }}</span>
                <span class="perm-tags">
                  <span v-if="perm.can_view" class="perm-tag view">查看</span>
                  <span v-if="perm.can_create" class="perm-tag create">创建</span>
                  <span v-if="perm.can_update" class="perm-tag update">更新</span>
                  <span v-if="perm.can_delete" class="perm-tag delete">删除</span>
                </span>
              </div>
              <div v-if="!selectedUser.cluster_permissions?.length" class="no-data">暂无集群授权</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建用户弹窗 -->
    <div v-if="showCreateModal" class="modal-mask" @click="showCreateModal = false">
      <div class="modal modal-create" @click.stop>
        <div class="modal-header">
          <h3>➕ 创建用户</h3>
          <button class="btn-close" @click="showCreateModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>用户名 <span class="required">*</span></label>
            <input v-model="createForm.username" placeholder="请输入用户名" />
          </div>
          <div class="form-group">
            <label>密码 <span class="required">*</span></label>
            <input type="password" v-model="createForm.password" placeholder="请输入密码" />
          </div>
          <div class="form-group">
            <label>确认密码 <span class="required">*</span></label>
            <input type="password" v-model="createForm.confirm_password" placeholder="再次输入密码" />
          </div>
          <div class="form-group">
            <label>邮箱</label>
            <input v-model="createForm.email" placeholder="请输入邮箱" />
          </div>
          <div class="form-group">
            <label>手机号</label>
            <input v-model="createForm.phone" placeholder="请输入手机号" />
          </div>
          <div class="form-group">
            <label>分配角色</label>
            <div class="role-checkboxes">
              <label v-for="role in allRoles" :key="role.id" class="checkbox-item">
                <input type="checkbox" :value="role.id" v-model="createForm.role_ids" />
                <span class="role-dot" :style="{ backgroundColor: role.color }"></span>
                {{ role.display_name || role.name }}
              </label>
            </div>
          </div>
          <div v-if="createError" class="form-error">{{ createError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn btn-primary" @click="submitCreateUser" :disabled="submitting">
            {{ submitting ? '创建中...' : '确认创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 分配角色弹窗（大厂风格） -->
    <div v-if="showRoleModal" class="modal-mask" @click="showRoleModal = false">
      <div class="modal modal-role" @click.stop>
        <div class="modal-header">
          <h3>🎭 分配角色</h3>
          <button class="btn-close" @click="showRoleModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="target-user">
            <div class="user-avatar-lg">{{ roleForm.username?.charAt(0)?.toUpperCase() }}</div>
            <div class="user-info-lg">
              <span class="username">{{ roleForm.username }}</span>
              <span class="hint">已选 {{ roleForm.role_ids.length }} 个角色</span>
            </div>
          </div>
          
          <div class="role-section">
            <div class="section-header">
              <span>可分配角色</span>
              <span class="role-count">{{ allRoles.length }} 个</span>
            </div>
            <div v-if="allRoles.length === 0" class="empty-roles">
              <span>暂无可用角色，请先创建角色</span>
            </div>
            <div v-else class="role-grid">
              <div
                v-for="role in allRoles"
                :key="role.id"
                :class="['role-card', { selected: roleForm.role_ids.includes(role.id) }]"
                @click="toggleRole(role.id)"
              >
                <div class="role-card-header">
                  <div class="role-dot" :style="{ backgroundColor: role.color || '#326ce5' }"></div>
                  <span class="role-name">{{ role.display_name || role.name }}</span>
                  <span class="check-mark" v-if="roleForm.role_ids.includes(role.id)">✓</span>
                </div>
                <div class="role-type-tag">{{ getRoleTypeLabel(role.role_type) }}</div>
                <div class="role-desc">{{ role.description || '暂无描述' }}</div>
              </div>
            </div>
          </div>
          
          <!-- 已选角色摘要 -->
          <div v-if="roleForm.role_ids.length > 0" class="selected-summary">
            <span class="summary-label">已选角色：</span>
            <div class="selected-tags">
              <span 
                v-for="roleId in roleForm.role_ids" 
                :key="roleId" 
                class="selected-tag"
                @click="toggleRole(roleId)"
              >
                {{ getRoleName(roleId) }}
                <span class="remove-icon">×</span>
              </span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRoleModal = false">取消</button>
          <button class="btn btn-primary" @click="saveUserRoles" :disabled="submitting">
            {{ submitting ? '保存中...' : '确认分配' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getUserList, createUser, updateUserStatus } from '@/api/user'
import { getAllRoles, getUserRBACInfo, assignUserRole } from '@/api/rbac'

const users = ref([])
const allRoles = ref([])
const searchQuery = ref('')
const statusFilter = ref('')
const loading = ref(false)
const submitting = ref(false)

const showDetailDrawer = ref(false)
const selectedUser = ref(null)

const showRoleModal = ref(false)
const roleForm = ref({
  user_id: 0,
  username: '',
  role_ids: []
})

// 创建用户
const showCreateModal = ref(false)
const createError = ref('')
const createForm = ref({
  username: '',
  password: '',
  confirm_password: '',
  email: '',
  phone: '',
  role_ids: []
})

const activeUsers = computed(() => users.value.filter(u => u.status === 1).length)

const filteredUsers = computed(() => {
  return users.value.filter(u => {
    const matchSearch = !searchQuery.value ||
      u.username?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      u.email?.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchStatus = !statusFilter.value ||
      (statusFilter.value === 'active' ? u.status === 1 : u.status !== 1)
    return matchSearch && matchStatus
  })
})

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getUserList({ page: 1, limit: 1000 })
    if (res.code === 0) {
      users.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载用户失败', e)
  } finally {
    loading.value = false
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

const viewUserDetail = async (user) => {
  selectedUser.value = { ...user }
  showDetailDrawer.value = true
  // 加载用户详细权限信息
  try {
    const res = await getUserRBACInfo(user.id)
    if (res.code === 0 && res.data) {
      selectedUser.value = { ...selectedUser.value, ...res.data }
    }
  } catch (e) {
    console.error('加载用户权限失败', e)
  }
}

const openRoleModal = (user) => {
  roleForm.value = {
    user_id: user.id,
    username: user.username,
    role_ids: user.roles?.map(r => r.id) || []
  }
  showRoleModal.value = true
}

const toggleRole = (roleId) => {
  const idx = roleForm.value.role_ids.indexOf(roleId)
  if (idx >= 0) {
    roleForm.value.role_ids.splice(idx, 1)
  } else {
    roleForm.value.role_ids.push(roleId)
  }
}

const saveUserRoles = async () => {
  submitting.value = true
  try {
    const res = await assignUserRole({
      user_id: roleForm.value.user_id,
      role_ids: roleForm.value.role_ids
    })
    if (res.code === 0) {
      showRoleModal.value = false
      loadUsers()
    }
  } catch (e) {
    console.error('分配角色失败', e)
  } finally {
    submitting.value = false
  }
}

const toggleUserStatus = async (user) => {
  const newStatus = user.status === 1 ? 0 : 1
  try {
    const res = await updateUserStatus(user, newStatus)
    if (res.code === 0) {
      loadUsers()
    }
  } catch (e) {
    console.error('更新状态失败', e)
  }
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString()
}

const getRoleTypeLabel = (type) => {
  const labels = {
    super_admin: '超级管理员',
    platform_admin: '平台管理员',
    cluster_admin: '集群管理员',
    developer: '开发者',
    viewer: '只读用户'
  }
  return labels[type] || type
}

const getRoleName = (roleId) => {
  const role = allRoles.value.find(r => r.id === roleId)
  return role ? (role.display_name || role.name) : ''
}

const openCreateModal = () => {
  createForm.value = {
    username: '',
    password: '',
    confirm_password: '',
    email: '',
    phone: '',
    role_ids: []
  }
  createError.value = ''
  showCreateModal.value = true
}

const submitCreateUser = async () => {
  createError.value = ''
  // 验证
  if (!createForm.value.username) {
    createError.value = '请输入用户名'
    return
  }
  if (!createForm.value.password) {
    createError.value = '请输入密码'
    return
  }
  if (createForm.value.password.length < 6) {
    createError.value = '密码长度至少6位'
    return
  }
  if (createForm.value.password !== createForm.value.confirm_password) {
    createError.value = '两次密码不一致'
    return
  }
  
  submitting.value = true
  try {
    const res = await createUser({
      username: createForm.value.username,
      password: createForm.value.password,
      password_confirm: createForm.value.confirm_password
    })
    if (res.code === 0) {
      // 如果选了角色，分配角色
      if (createForm.value.role_ids.length > 0 && res.data?.id) {
        await assignUserRole({
          user_id: res.data.id,
          role_ids: createForm.value.role_ids
        })
      }
      showCreateModal.value = false
      loadUsers()
    } else {
      createError.value = res.message || '创建失败'
    }
  } catch (e) {
    createError.value = e.message || '创建失败'
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadUsers()
  loadRoles()
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

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 16px;
  background: white;
  border-radius: 8px;
}

.toolbar-left, .toolbar-right {
  display: flex;
  gap: 12px;
  align-items: center;
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
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn:hover {
  opacity: 0.9;
}

.data-table {
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 14px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

th {
  background: #f8fafc;
  font-weight: 600;
  color: #64748b;
  font-size: 13px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
}

.status-tag {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
}

.status-tag.active {
  background: #dcfce7;
  color: #15803d;
}

.status-tag.disabled {
  background: #fee2e2;
  color: #b91c1c;
}

.role-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.role-tag {
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.no-role {
  color: #94a3b8;
  font-size: 12px;
}

.action-btns {
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

.btn-text:hover {
  text-decoration: underline;
}

/* 抽屉 */
.drawer-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  width: 420px;
  background: white;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 20px rgba(0, 0, 0, 0.1);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.drawer-header h3 {
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #94a3b8;
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-title {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e2e8f0;
}

.detail-item {
  display: flex;
  margin-bottom: 10px;
}

.detail-item label {
  width: 80px;
  color: #64748b;
  font-size: 13px;
}

.detail-item span {
  flex: 1;
  font-size: 14px;
}

.role-list, .auth-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.role-card, .auth-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
}

.role-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.role-info {
  display: flex;
  flex-direction: column;
}

.role-name {
  font-weight: 500;
  font-size: 14px;
}

.role-type, .role-desc {
  font-size: 12px;
  color: #64748b;
}

.perm-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.perm-tag {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.perm-tag.view { background: #dbeafe; color: #1e40af; }
.perm-tag.create { background: #dcfce7; color: #166534; }
.perm-tag.update { background: #fef3c7; color: #92400e; }
.perm-tag.delete { background: #fee2e2; color: #991b1b; }

.no-data {
  color: #94a3b8;
  font-size: 13px;
  padding: 12px;
  text-align: center;
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
  width: 500px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.2);
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

.modal-body {
  padding: 24px;
  max-height: 400px;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

.target-user {
  margin-bottom: 16px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 6px;
}

.target-user .label {
  color: #64748b;
}

.target-user .value {
  font-weight: 600;
}

.role-selection {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.role-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.role-item:hover {
  border-color: #94a3b8;
}

.role-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.check-icon {
  margin-left: auto;
  color: #3b82f6;
  font-weight: 600;
}

/* 分配角色弹窗（大厂风格） */
.modal-role {
  width: 640px;
  max-height: 80vh;
}

.modal-role .modal-body {
  max-height: 60vh;
  overflow-y: auto;
  padding: 20px 24px;
}

.modal-role .target-user {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  margin-bottom: 20px;
}

.user-avatar-lg {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 20px;
}

.user-info-lg {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-info-lg .username {
  font-weight: 600;
  font-size: 16px;
  color: #1e293b;
}

.user-info-lg .hint {
  font-size: 13px;
  color: #64748b;
}

.role-section {
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  color: #475569;
}

.role-count {
  font-size: 12px;
  color: #94a3b8;
  font-weight: normal;
}

.empty-roles {
  text-align: center;
  padding: 40px;
  color: #94a3b8;
  background: #f8fafc;
  border-radius: 8px;
}

.role-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.role-card {
  padding: 14px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

.role-card:hover {
  border-color: #94a3b8;
  background: #f8fafc;
}

.role-card.selected {
  border-color: #3b82f6;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.role-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.role-card .role-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.role-card .role-name {
  font-weight: 600;
  font-size: 14px;
  color: #1e293b;
  flex: 1;
}

.check-mark {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #3b82f6;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.role-type-tag {
  display: inline-block;
  padding: 2px 8px;
  background: #e2e8f0;
  border-radius: 4px;
  font-size: 11px;
  color: #64748b;
  margin-bottom: 6px;
}

.role-card .role-desc {
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-summary {
  padding: 12px 16px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  margin-top: 16px;
}

.summary-label {
  font-size: 13px;
  color: #166534;
  margin-right: 8px;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.selected-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: #dcfce7;
  border-radius: 16px;
  font-size: 12px;
  color: #166534;
  cursor: pointer;
  transition: all 0.2s;
}

.selected-tag:hover {
  background: #bbf7d0;
}

.selected-tag .remove-icon {
  font-size: 14px;
  font-weight: 600;
  opacity: 0.7;
}

.selected-tag:hover .remove-icon {
  opacity: 1;
}

/* 创建用户弹窗 */
.modal-create {
  width: 480px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.form-group .required {
  color: #ef4444;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-error {
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #dc2626;
  font-size: 13px;
  margin-top: 12px;
}

.role-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.checkbox-item:hover {
  background: #f9fafb;
}

.checkbox-item input {
  width: auto;
}

.checkbox-item .role-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
</style>

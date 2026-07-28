<template>
  <div class="rbac-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <div class="header-content">
        <div class="header-icon">🛡️</div>
        <div class="header-text">
          <h1>权限管理</h1>
          <p>基于RBAC的细粒度权限控制，管理用户角色和集群访问权限</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-card">
          <div class="stat-value">{{ roles.length }}</div>
          <div class="stat-label">角色总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ clusterPermissions.length }}</div>
          <div class="stat-label">权限配置</div>
        </div>
        <div class="stat-card highlight">
          <div class="stat-value">{{ clusters.length }}</div>
          <div class="stat-label">管理集群</div>
        </div>
      </div>
    </div>

    <!-- 标签页导航 -->
    <div class="tab-nav">
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'roles' }"
        @click="activeTab = 'roles'"
      >
        <span class="tab-icon">👥</span>
        角色管理
      </button>
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'cluster-permissions' }"
        @click="activeTab = 'cluster-permissions'"
      >
        <span class="tab-icon">☸️</span>
        集群权限
      </button>
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'user-permissions' }"
        @click="activeTab = 'user-permissions'"
      >
        <span class="tab-icon">🔐</span>
        用户授权
      </button>
    </div>

    <!-- 角色管理 -->
    <div v-show="activeTab === 'roles'" class="tab-content">
      <div class="content-header">
        <h2>系统角色</h2>
        <button class="btn btn-primary" @click="openRoleModal('create')">
          ➕ 创建角色
        </button>
      </div>

      <div class="roles-grid">
        <div 
          v-for="role in roles" 
          :key="role.id" 
          class="role-card"
          :class="{ 'system-role': role.is_system }"
        >
          <div class="role-header" :style="{ borderLeftColor: role.color || '#326ce5' }">
            <div class="role-icon" :style="{ background: role.color || '#326ce5' }">
              {{ getRoleIcon(role.role_type) }}
            </div>
            <div class="role-info">
              <h3>{{ role.display_name }}</h3>
              <span class="role-name">{{ role.name }}</span>
            </div>
            <div class="role-badge" :class="role.role_type">
              {{ getRoleTypeText(role.role_type) }}
            </div>
          </div>
          <div class="role-body">
            <p class="role-desc">{{ role.description || '暂无描述' }}</p>
            <div class="role-meta">
              <span v-if="role.is_system" class="system-tag">🔒 系统内置</span>
              <span class="user-count">👤 {{ role.user_count || 0 }} 用户</span>
            </div>
          </div>
          <div class="role-footer">
            <button 
              class="role-action" 
              @click="openRoleModal('edit', role)"
              :disabled="role.is_system"
            >
              ✏️ 编辑
            </button>
            <button 
              class="role-action danger" 
              @click="handleDeleteRole(role)"
              :disabled="role.is_system"
            >
              🗑️ 删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 集群权限管理 -->
    <div v-show="activeTab === 'cluster-permissions'" class="tab-content">
      <div class="content-header">
        <h2>集群权限分配</h2>
        <div class="header-actions">
          <select v-model="filterClusterId" class="filter-select">
            <option :value="0">全部集群</option>
            <option v-for="c in clusters" :key="c.id" :value="c.id">
              {{ c.cluster_name }}
            </option>
          </select>
          <button class="btn btn-primary" @click="openClusterPermModal('create')">
            ➕ 添加权限
          </button>
        </div>
      </div>

      <div class="table-container">
        <table class="permission-table">
          <thead>
            <tr>
              <th style="width: 60px;">ID</th>
              <th>用户</th>
              <th>集群</th>
              <th>角色类型</th>
              <th>权限</th>
              <th>命名空间</th>
              <th style="width: 140px;">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="perm in filteredClusterPermissions" :key="perm.id">
              <td>{{ perm.id }}</td>
              <td>
                <div class="user-cell">
                  <span class="user-avatar">👤</span>
                  {{ perm.username || `用户#${perm.user_id}` }}
                </div>
              </td>
              <td>
                <div class="cluster-cell">
                  <span class="cluster-icon">☸️</span>
                  {{ perm.cluster_name || `集群#${perm.cluster_id}` }}
                </div>
              </td>
              <td>
                <span class="role-badge" :class="perm.role_type">
                  {{ getRoleTypeText(perm.role_type) }}
                </span>
              </td>
              <td>
                <div class="permission-tags">
                  <span v-if="perm.can_view" class="perm-tag view">查看</span>
                  <span v-if="perm.can_create" class="perm-tag create">创建</span>
                  <span v-if="perm.can_update" class="perm-tag update">更新</span>
                  <span v-if="perm.can_delete" class="perm-tag delete">删除</span>
                  <span v-if="perm.can_exec" class="perm-tag exec">执行</span>
                </div>
              </td>
              <td>
                <span v-if="perm.ns_list && perm.ns_list.length > 0" class="ns-list">
                  {{ perm.ns_list.join(', ') }}
                </span>
                <span v-else class="all-ns">全部命名空间</span>
              </td>
              <td>
                <div class="action-btns">
                  <button class="btn-mini" @click="openClusterPermModal('edit', perm)">
                    ✏️
                  </button>
                  <button class="btn-mini danger" @click="handleDeleteClusterPerm(perm)">
                    🗑️
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="filteredClusterPermissions.length === 0" class="empty-state">
          <div class="empty-icon">📭</div>
          <div class="empty-text">暂无权限配置</div>
          <button class="btn btn-primary" @click="openClusterPermModal('create')">
            添加第一个权限
          </button>
        </div>
      </div>
    </div>

    <!-- 用户授权面板 -->
    <div v-show="activeTab === 'user-permissions'" class="tab-content">
      <div class="content-header">
        <h2>用户权限配置</h2>
        <div class="user-search">
          <input 
            type="text" 
            v-model="userSearchQuery" 
            placeholder="输入用户ID查询权限..."
            @keyup.enter="searchUserPermissions"
          />
          <button class="btn btn-primary" @click="searchUserPermissions">
            🔍 查询
          </button>
        </div>
      </div>

      <div v-if="selectedUserInfo" class="user-permission-panel">
        <div class="user-header">
          <div class="user-avatar-large">👤</div>
          <div class="user-info">
            <h3>{{ selectedUserInfo.username || `用户 #${selectedUserInfo.user_id}` }}</h3>
            <span v-if="selectedUserInfo.is_super_admin" class="super-admin-badge">
              👑 超级管理员
            </span>
          </div>
        </div>

        <div class="permission-sections">
          <!-- 角色分配 -->
          <div class="permission-section">
            <div class="section-header">
              <h4>🎭 角色分配</h4>
              <button class="btn btn-sm" @click="openUserRoleModal">
                配置角色
              </button>
            </div>
            <div class="role-list">
              <div v-for="role in selectedUserInfo.roles" :key="role.id" class="role-item">
                <span class="role-dot" :style="{ background: role.color }"></span>
                <span>{{ role.display_name }}</span>
              </div>
              <div v-if="!selectedUserInfo.roles?.length" class="no-roles">
                暂未分配角色
              </div>
            </div>
          </div>

          <!-- 集群权限 -->
          <div class="permission-section">
            <div class="section-header">
              <h4>☸️ 集群权限</h4>
              <button class="btn btn-sm" @click="openBatchClusterModal">
                批量配置
              </button>
            </div>
            <div class="cluster-perm-list">
              <div 
                v-for="cp in selectedUserInfo.cluster_permissions" 
                :key="cp.id" 
                class="cluster-perm-item"
              >
                <div class="cluster-info">
                  <span class="cluster-name">{{ cp.cluster_name }}</span>
                  <span class="role-badge small" :class="cp.role_type">
                    {{ getRoleTypeText(cp.role_type) }}
                  </span>
                </div>
                <div class="perm-icons">
                  <span v-if="cp.can_view" class="perm-icon" title="查看">👁️</span>
                  <span v-if="cp.can_create" class="perm-icon" title="创建">➕</span>
                  <span v-if="cp.can_update" class="perm-icon" title="更新">✏️</span>
                  <span v-if="cp.can_delete" class="perm-icon" title="删除">🗑️</span>
                  <span v-if="cp.can_exec" class="perm-icon" title="执行">⚡</span>
                </div>
              </div>
              <div v-if="!selectedUserInfo.cluster_permissions?.length" class="no-perms">
                暂无集群权限配置
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="user-search-hint">
        <div class="hint-icon">🔍</div>
        <p>输入用户ID查询其权限配置</p>
      </div>
    </div>

    <!-- 角色编辑弹窗 -->
    <div v-if="showRoleModal" class="modal">
      <div class="modal-backdrop" @click="showRoleModal = false"></div>
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ roleModalMode === 'create' ? '创建角色' : '编辑角色' }}</h2>
          <button class="close-btn" @click="showRoleModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitRole">
            <div class="form-group" v-if="roleModalMode === 'create'">
              <label>角色标识 <span class="required">*</span></label>
              <input type="text" v-model="roleForm.name" placeholder="如: custom_role" required />
              <span class="form-hint">唯一标识，创建后不可修改</span>
            </div>
            <div class="form-group">
              <label>显示名称 <span class="required">*</span></label>
              <input type="text" v-model="roleForm.display_name" placeholder="如: 自定义角色" required />
            </div>
            <div class="form-group">
              <label>角色类型 <span class="required">*</span></label>
              <select v-model="roleForm.role_type" required>
                <option value="super_admin">超级管理员</option>
                <option value="cluster_admin">集群管理员</option>
                <option value="developer">开发者</option>
                <option value="viewer">只读用户</option>
                <option value="custom">自定义</option>
              </select>
            </div>
            <div class="form-group">
              <label>描述</label>
              <textarea v-model="roleForm.description" placeholder="角色描述..." rows="3"></textarea>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>颜色</label>
                <input type="color" v-model="roleForm.color" />
              </div>
              <div class="form-group">
                <label>图标</label>
                <select v-model="roleForm.icon">
                  <option value="user">👤 用户</option>
                  <option value="admin">👑 管理员</option>
                  <option value="dev">💻 开发者</option>
                  <option value="viewer">👁️ 观察者</option>
                </select>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" class="btn btn-secondary" @click="showRoleModal = false">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '提交中...' : '确认' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 集群权限编辑弹窗 -->
    <div v-if="showClusterPermModal" class="modal">
      <div class="modal-backdrop" @click="showClusterPermModal = false"></div>
      <div class="modal-content modal-lg">
        <div class="modal-header">
          <h2>{{ clusterPermModalMode === 'create' ? '添加集群权限' : '编辑集群权限' }}</h2>
          <button class="close-btn" @click="showClusterPermModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitClusterPerm">
            <div class="form-row">
              <div class="form-group">
                <label>用户ID <span class="required">*</span></label>
                <input type="number" v-model.number="clusterPermForm.user_id" required :disabled="clusterPermModalMode === 'edit'" />
              </div>
              <div class="form-group">
                <label>集群 <span class="required">*</span></label>
                <select v-model.number="clusterPermForm.cluster_id" required :disabled="clusterPermModalMode === 'edit'">
                  <option :value="0" disabled>请选择集群</option>
                  <option v-for="c in clusters" :key="c.id" :value="c.id">
                    {{ c.cluster_name }}
                  </option>
                </select>
              </div>
            </div>
            <div class="form-group">
              <label>角色类型 <span class="required">*</span></label>
              <div class="role-type-cards">
                <label 
                  v-for="rt in roleTypes" 
                  :key="rt.value" 
                  class="role-type-card"
                  :class="{ selected: clusterPermForm.role_type === rt.value }"
                >
                  <input type="radio" v-model="clusterPermForm.role_type" :value="rt.value" />
                  <span class="rt-icon">{{ rt.icon }}</span>
                  <span class="rt-name">{{ rt.label }}</span>
                  <span class="rt-desc">{{ rt.desc }}</span>
                </label>
              </div>
            </div>
            <div class="form-group">
              <label>权限配置</label>
              <div class="permission-checkboxes">
                <label class="perm-checkbox">
                  <input type="checkbox" v-model="clusterPermForm.can_view" />
                  <span class="perm-label view">👁️ 查看</span>
                </label>
                <label class="perm-checkbox">
                  <input type="checkbox" v-model="clusterPermForm.can_create" />
                  <span class="perm-label create">➕ 创建</span>
                </label>
                <label class="perm-checkbox">
                  <input type="checkbox" v-model="clusterPermForm.can_update" />
                  <span class="perm-label update">✏️ 更新</span>
                </label>
                <label class="perm-checkbox">
                  <input type="checkbox" v-model="clusterPermForm.can_delete" />
                  <span class="perm-label delete">🗑️ 删除</span>
                </label>
                <label class="perm-checkbox">
                  <input type="checkbox" v-model="clusterPermForm.can_exec" />
                  <span class="perm-label exec">⚡ 执行</span>
                </label>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" class="btn btn-secondary" @click="showClusterPermModal = false">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '提交中...' : '确认' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 用户角色分配弹窗 -->
    <div v-if="showUserRoleModal" class="modal">
      <div class="modal-backdrop" @click="showUserRoleModal = false"></div>
      <div class="modal-content">
        <div class="modal-header">
          <h2>配置用户角色</h2>
          <button class="close-btn" @click="showUserRoleModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="role-selection">
            <label 
              v-for="role in roles" 
              :key="role.id" 
              class="role-select-item"
              :class="{ selected: userRoleForm.role_ids.includes(role.id) }"
            >
              <input 
                type="checkbox" 
                :value="role.id" 
                v-model="userRoleForm.role_ids"
              />
              <span class="role-color" :style="{ background: role.color }"></span>
              <span class="role-name">{{ role.display_name }}</span>
              <span class="role-type-tag">{{ getRoleTypeText(role.role_type) }}</span>
            </label>
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-secondary" @click="showUserRoleModal = false">取消</button>
            <button class="btn btn-primary" @click="submitUserRole" :disabled="submitting">
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()
import {
  getRoleList,
  getAllRoles,
  createRole,
  updateRole,
  deleteRole,
  getClusterPermissionList,
  createClusterPermission,
  updateClusterPermission,
  deleteClusterPermission,
  getUserRBACInfo,
  assignUserRole
} from '@/api/rbac'
import { getClusterList } from '@/api/cluster'

// ==================== 状态 ====================
const activeTab = ref('roles')
const loading = ref(true)
const submitting = ref(false)

// 角色数据
const roles = ref([])

// 集群权限数据
const clusterPermissions = ref([])
const clusters = ref([])
const filterClusterId = ref(0)

// 用户权限查询
const userSearchQuery = ref('')
const selectedUserInfo = ref(null)

// 弹窗状态
const showRoleModal = ref(false)
const roleModalMode = ref('create')
const roleForm = ref({
  id: 0,
  name: '',
  display_name: '',
  description: '',
  role_type: 'developer',
  color: '#326ce5',
  icon: 'user'
})

const showClusterPermModal = ref(false)
const clusterPermModalMode = ref('create')
const clusterPermForm = ref({
  id: 0,
  user_id: 0,
  cluster_id: 0,
  role_type: 'viewer',
  can_view: true,
  can_create: false,
  can_update: false,
  can_delete: false,
  can_exec: false
})

const showUserRoleModal = ref(false)
const userRoleForm = ref({
  user_id: 0,
  role_ids: []
})

// ==================== 计算属性 ====================
const filteredClusterPermissions = computed(() => {
  if (filterClusterId.value === 0) return clusterPermissions.value
  return clusterPermissions.value.filter(p => p.cluster_id === filterClusterId.value)
})

const roleTypes = [
  { value: 'cluster_admin', label: '集群管理员', icon: '👑', desc: '完整集群管理权限' },
  { value: 'developer', label: '开发者', icon: '💻', desc: '创建和管理工作负载' },
  { value: 'viewer', label: '只读用户', icon: '👁️', desc: '仅查看权限' }
]

// ==================== 方法 ====================
const getRoleIcon = (type) => {
  const icons = {
    super_admin: '👑',
    cluster_admin: '🛡️',
    developer: '💻',
    viewer: '👁️',
    custom: '⚙️'
  }
  return icons[type] || '👤'
}

const getRoleTypeText = (type) => {
  const texts = {
    super_admin: '超级管理员',
    cluster_admin: '集群管理员',
    developer: '开发者',
    viewer: '只读',
    custom: '自定义'
  }
  return texts[type] || type
}

// 加载数据
const fetchRoles = async () => {
  try {
    const res = await getRoleList({ page: 1, limit: 100 })
    roles.value = res.data?.list || res.list || []
  } catch (e) {
    console.error('获取角色列表失败:', e)
    Message.error('获取角色列表失败')
  }
}

const fetchClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 100 })
    clusters.value = res.data?.list || res.list || []
  } catch (e) {
    console.error('获取集群列表失败:', e)
  }
}

const fetchClusterPermissions = async () => {
  try {
    const res = await getClusterPermissionList({ page: 1, limit: 100 })
    clusterPermissions.value = res.data?.list || res.list || []
  } catch (e) {
    console.error('获取集群权限失败:', e)
    Message.error('获取集群权限失败')
  }
}

// 角色操作
const openRoleModal = (mode, role = null) => {
  roleModalMode.value = mode
  if (mode === 'edit' && role) {
    roleForm.value = { ...role }
  } else {
    roleForm.value = {
      id: 0,
      name: '',
      display_name: '',
      description: '',
      role_type: 'developer',
      color: '#326ce5',
      icon: 'user'
    }
  }
  showRoleModal.value = true
}

const submitRole = async () => {
  submitting.value = true
  try {
    if (roleModalMode.value === 'create') {
      await createRole(roleForm.value)
      Message.success('角色创建成功')
    } else {
      await updateRole(roleForm.value)
      Message.success('角色更新成功')
    }
    showRoleModal.value = false
    await fetchRoles()
  } catch (e) {
    Message.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDeleteRole = async (role) => {
  if (role.is_system) {
    Message.warning('系统内置角色不可删除')
    return
  }
  const ok = await showConfirm({
    title: '确认删除角色',
    content: '删除后关联用户将失去该角色权限。',
    type: 'danger',
    details: [
      { label: '角色名称', value: role.display_name },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  
  try {
    await deleteRole(role.id)
    Message.success('删除成功')
    await fetchRoles()
  } catch (e) {
    Message.error('删除失败: ' + (e.message || '未知错误'))
  }
}

// 集群权限操作
const openClusterPermModal = (mode, perm = null) => {
  clusterPermModalMode.value = mode
  if (mode === 'edit' && perm) {
    clusterPermForm.value = { ...perm }
  } else {
    clusterPermForm.value = {
      id: 0,
      user_id: 0,
      cluster_id: 0,
      role_type: 'viewer',
      can_view: true,
      can_create: false,
      can_update: false,
      can_delete: false,
      can_exec: false
    }
  }
  showClusterPermModal.value = true
}

const submitClusterPerm = async () => {
  submitting.value = true
  try {
    if (clusterPermModalMode.value === 'create') {
      await createClusterPermission(clusterPermForm.value)
      Message.success('权限添加成功')
    } else {
      await updateClusterPermission(clusterPermForm.value)
      Message.success('权限更新成功')
    }
    showClusterPermModal.value = false
    await fetchClusterPermissions()
  } catch (e) {
    Message.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDeleteClusterPerm = async (perm) => {
  const ok = await showConfirm({
    title: '确认删除权限配置',
    content: '删除后用户将失去对应集群的访问权限。',
    type: 'danger',
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  
  try {
    await deleteClusterPermission(perm.id)
    Message.success('删除成功')
    await fetchClusterPermissions()
  } catch (e) {
    Message.error('删除失败: ' + (e.message || '未知错误'))
  }
}

// 用户权限查询
const searchUserPermissions = async () => {
  const userId = parseInt(userSearchQuery.value)
  if (!userId || userId <= 0) {
    Message.warning('请输入有效的用户ID')
    return
  }
  
  try {
    const res = await getUserRBACInfo(userId)
    selectedUserInfo.value = res.data || res
    Message.success('查询成功')
  } catch (e) {
    Message.error('查询失败: ' + (e.message || '用户不存在'))
    selectedUserInfo.value = null
  }
}

// 用户角色分配
const openUserRoleModal = () => {
  if (!selectedUserInfo.value) return
  userRoleForm.value = {
    user_id: selectedUserInfo.value.user_id,
    role_ids: selectedUserInfo.value.roles?.map(r => r.id) || []
  }
  showUserRoleModal.value = true
}

const submitUserRole = async () => {
  submitting.value = true
  try {
    await assignUserRole(userRoleForm.value)
    Message.success('角色分配成功')
    showUserRoleModal.value = false
    // 重新查询用户信息
    await searchUserPermissions()
  } catch (e) {
    Message.error('分配失败: ' + (e.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

const openBatchClusterModal = () => {
  if (!selectedUserInfo.value) return
  clusterPermForm.value = {
    id: 0,
    user_id: selectedUserInfo.value.user_id,
    cluster_id: 0,
    role_type: 'viewer',
    can_view: true,
    can_create: false,
    can_update: false,
    can_delete: false,
    can_exec: false
  }
  clusterPermModalMode.value = 'create'
  showClusterPermModal.value = true
}

// ==================== 生命周期 ====================
onMounted(async () => {
  loading.value = true
  await Promise.all([
    fetchRoles(),
    fetchClusters(),
    fetchClusterPermissions()
  ])
  loading.value = false
})
</script>

<style scoped>
/* ==================== 主容器 ==================== */
.rbac-view {
  padding: 24px;
  max-width: 1600px;
  margin: 0 auto;
}

/* ==================== 页面头部 ==================== */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  color: white;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  font-size: 48px;
}

.header-text h1 {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px 0;
}

.header-text p {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.header-stats {
  display: flex;
  gap: 16px;
}

.stat-card {
  background: rgba(255, 255, 255, 0.15);
  padding: 16px 24px;
  border-radius: 12px;
  text-align: center;
  backdrop-filter: blur(10px);
}

.stat-card.highlight {
  background: rgba(255, 255, 255, 0.25);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
}

.stat-label {
  font-size: 12px;
  opacity: 0.9;
  margin-top: 4px;
}

/* ==================== 标签页导航 ==================== */
.tab-nav {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  background: #f1f5f9;
  padding: 6px;
  border-radius: 12px;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 24px;
  border: none;
  background: transparent;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #334155;
  background: rgba(255, 255, 255, 0.5);
}

.tab-btn.active {
  background: white;
  color: #1e293b;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.tab-icon {
  font-size: 18px;
}

/* ==================== 内容区域 ==================== */
.tab-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content-header h2 {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.filter-select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: white;
}

/* ==================== 按钮 ==================== */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  transform: translateY(-1px);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ==================== 角色卡片 ==================== */
.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.role-card {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.2s;
}

.role-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.role-card.system-role {
  border: 1px solid #e2e8f0;
}

.role-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-left: 4px solid;
  background: #f8fafc;
}

.role-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: white;
}

.role-info {
  flex: 1;
}

.role-info h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #1e293b;
}

.role-name {
  font-size: 12px;
  color: #64748b;
  font-family: monospace;
}

.role-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.role-badge.super_admin {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #92400e;
}

.role-badge.cluster_admin {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  color: #1e40af;
}

.role-badge.developer {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  color: #065f46;
}

.role-badge.viewer {
  background: linear-gradient(135deg, #f1f5f9 0%, #e2e8f0 100%);
  color: #475569;
}

.role-badge.custom {
  background: linear-gradient(135deg, #fce7f3 0%, #fbcfe8 100%);
  color: #9d174d;
}

.role-badge.small {
  padding: 2px 8px;
  font-size: 10px;
}

.role-body {
  padding: 16px;
}

.role-desc {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 12px 0;
  line-height: 1.5;
}

.role-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #94a3b8;
}

.system-tag {
  color: #f59e0b;
}

.role-footer {
  display: flex;
  border-top: 1px solid #f1f5f9;
}

.role-action {
  flex: 1;
  padding: 12px;
  border: none;
  background: transparent;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.role-action:hover:not(:disabled) {
  background: #f8fafc;
}

.role-action.danger:hover:not(:disabled) {
  background: #fef2f2;
  color: #dc2626;
}

.role-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* ==================== 权限表格 ==================== */
.table-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.permission-table {
  width: 100%;
  border-collapse: collapse;
}

.permission-table th {
  background: #f8fafc;
  padding: 14px 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
}

.permission-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  font-size: 14px;
}

.permission-table tbody tr:hover {
  background: #f8fafc;
}

.user-cell, .cluster-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar, .cluster-icon {
  font-size: 18px;
}

.permission-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.perm-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.perm-tag.view { background: #dbeafe; color: #1e40af; }
.perm-tag.create { background: #d1fae5; color: #065f46; }
.perm-tag.update { background: #fef3c7; color: #92400e; }
.perm-tag.delete { background: #fee2e2; color: #991b1b; }
.perm-tag.exec { background: #fce7f3; color: #9d174d; }

.ns-list {
  font-size: 12px;
  color: #64748b;
  font-family: monospace;
}

.all-ns {
  font-size: 12px;
  color: #94a3b8;
  font-style: italic;
}

.action-btns {
  display: flex;
  gap: 6px;
}

.btn-mini {
  padding: 6px 10px;
  border: none;
  background: #f1f5f9;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-mini:hover {
  background: #e2e8f0;
}

.btn-mini.danger:hover {
  background: #fee2e2;
}

/* ==================== 空状态 ==================== */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
  margin-bottom: 20px;
}

/* ==================== 用户权限面板 ==================== */
.user-search {
  display: flex;
  gap: 12px;
}

.user-search input {
  width: 200px;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
}

.user-search input:focus {
  outline: none;
  border-color: #3b82f6;
}

.user-search-hint {
  text-align: center;
  padding: 80px 20px;
  background: #f8fafc;
  border-radius: 12px;
}

.hint-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.user-permission-panel {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f1f5f9;
  margin-bottom: 24px;
}

.user-avatar-large {
  font-size: 48px;
  width: 64px;
  height: 64px;
  background: #f1f5f9;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-info h3 {
  font-size: 20px;
  margin: 0 0 8px 0;
}

.super-admin-badge {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #92400e;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.permission-sections {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.permission-section {
  background: #f8fafc;
  border-radius: 12px;
  padding: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h4 {
  margin: 0;
  font-size: 16px;
  color: #1e293b;
}

.role-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.role-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: white;
  border-radius: 8px;
}

.role-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.no-roles, .no-perms {
  color: #94a3b8;
  font-size: 14px;
  text-align: center;
  padding: 20px;
}

.cluster-perm-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cluster-perm-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: white;
  border-radius: 8px;
}

.cluster-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cluster-name {
  font-weight: 500;
}

.perm-icons {
  display: flex;
  gap: 4px;
}

.perm-icon {
  font-size: 14px;
}

/* ==================== 模态框 ==================== */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-backdrop {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
}

.modal-content {
  position: relative;
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  animation: slideUp 0.3s ease;
}

.modal-content.modal-lg {
  max-width: 700px;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  font-size: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #e2e8f0;
}

.modal-body {
  padding: 24px;
  max-height: calc(90vh - 140px);
  overflow-y: auto;
}

/* ==================== 表单 ==================== */
.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #334155;
  margin-bottom: 8px;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group input[type="color"] {
  width: 60px;
  height: 40px;
  padding: 4px;
}

.form-hint {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
}

.required {
  color: #ef4444;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-row .form-group {
  flex: 1;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #f1f5f9;
}

/* 角色类型卡片 */
.role-type-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.role-type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.role-type-card:hover {
  border-color: #94a3b8;
}

.role-type-card.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.role-type-card input {
  display: none;
}

.rt-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.rt-name {
  font-weight: 600;
  font-size: 14px;
  color: #1e293b;
}

.rt-desc {
  font-size: 11px;
  color: #64748b;
  margin-top: 4px;
}

/* 权限复选框 */
.permission-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.perm-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.perm-checkbox:hover {
  border-color: #94a3b8;
}

.perm-checkbox input:checked + .perm-label {
  font-weight: 600;
}

.perm-label {
  font-size: 14px;
}

.perm-label.view { color: #1e40af; }
.perm-label.create { color: #065f46; }
.perm-label.update { color: #92400e; }
.perm-label.delete { color: #991b1b; }
.perm-label.exec { color: #9d174d; }

/* 角色选择列表 */
.role-selection {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.role-select-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.role-select-item:hover {
  border-color: #94a3b8;
}

.role-select-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.role-color {
  width: 12px;
  height: 12px;
  border-radius: 4px;
}

.role-type-tag {
  margin-left: auto;
  font-size: 11px;
  color: #64748b;
}
</style>

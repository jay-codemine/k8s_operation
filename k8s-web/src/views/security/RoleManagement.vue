<template>
  <div class="security-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <div class="header-content">
        <div class="header-icon">🎭</div>
        <div class="header-text">
          <h1>角色管理</h1>
          <p>管理平台角色、配置权限、查看绑定用户</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-card">
          <div class="stat-value">{{ roles.length }}</div>
          <div class="stat-label">角色总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ systemRoles }}</div>
          <div class="stat-label">系统角色</div>
        </div>
      </div>
    </div>

    <!-- 标签页 -->
    <div class="tab-nav">
      <button :class="['tab-btn', { active: activeTab === 'platform' }]" @click="activeTab = 'platform'">
        <span class="tab-icon">🏷️</span> 平台角色
      </button>
      <button :class="['tab-btn', { active: activeTab === 'k8s' }]" @click="activeTab = 'k8s'">
        <span class="tab-icon">☸️</span> Kubernetes 角色
      </button>
    </div>

    <!-- 平台角色 -->
    <div v-show="activeTab === 'platform'" class="tab-content">
      <div class="content-header">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input v-model="searchQuery" placeholder="搜索角色..." />
        </div>
        <button class="btn btn-primary" @click="openCreateModal">
          ➕ 创建角色
        </button>
      </div>

      <div class="role-grid">
        <div v-for="role in filteredRoles" :key="role.id" class="role-card">
          <div class="role-header">
            <div class="role-dot" :style="{ backgroundColor: role.color }"></div>
            <div class="role-info">
              <span class="role-name">{{ role.display_name || role.name }}</span>
              <span class="role-type">{{ getRoleTypeLabel(role.role_type) }}</span>
            </div>
            <div class="role-actions" v-if="!role.is_system">
              <button class="btn-icon" @click="editRole(role)">✏️</button>
              <button class="btn-icon" @click="deleteRole(role)">🗑️</button>
            </div>
            <span v-else class="system-badge">系统</span>
          </div>
          <div class="role-desc">{{ role.description || '暂无描述' }}</div>
          <div class="role-meta">
            <span class="meta-item">
              <span class="meta-icon">👥</span>
              {{ role.user_count || 0 }} 用户
            </span>
            <span class="meta-item">
              <span class="meta-icon">📅</span>
              {{ formatTime(role.created_at) }}
            </span>
          </div>
          <div class="role-footer">
            <button class="btn-text" @click="viewRoleUsers(role)">查看绑定用户</button>
            <button class="btn-text primary" @click="openPermissionModal(role)">配置权限</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Kubernetes 角色 -->
    <div v-show="activeTab === 'k8s'" class="tab-content">
      <div class="content-header">
        <div class="toolbar-left">
          <select v-model="selectedClusterId" class="filter-select" @change="loadK8sRoles">
            <option value="">选择集群</option>
            <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name }}</option>
          </select>
          <select v-model="k8sRoleType" class="filter-select" @change="loadK8sRoles">
            <option value="">所有类型</option>
            <option value="Role">Role（命名空间级）</option>
            <option value="ClusterRole">ClusterRole（集群级）</option>
          </select>
        </div>
        <button class="btn btn-secondary" @click="loadK8sRoles">
          🔄 刷新
        </button>
      </div>

      <div v-if="!selectedClusterId" class="empty-state">
        <div class="empty-icon">☸️</div>
        <p>请先选择集群查看 Kubernetes 角色</p>
      </div>

      <div v-else class="k8s-role-list">
        <div v-for="role in k8sRoles" :key="role.name" class="k8s-role-card">
          <div class="k8s-role-header">
            <span :class="['type-badge', role.type === 'ClusterRole' ? 'cluster' : 'namespace']">
              {{ role.type }}
            </span>
            <span class="k8s-role-name">{{ role.name }}</span>
            <span v-if="role.namespace" class="namespace-tag">{{ role.namespace }}</span>
          </div>
          <div class="k8s-role-rules">
            <span class="rules-count">{{ role.rules?.length || 0 }} 条规则</span>
            <button class="btn-text" @click="viewK8sRoleDetail(role)">查看详情</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑角色弹窗 -->
    <div v-if="showRoleModal" class="modal-mask" @click="showRoleModal = false">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>{{ roleModalMode === 'create' ? '创建角色' : '编辑角色' }}</h3>
          <button class="btn-close" @click="showRoleModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>角色标识 <span class="required">*</span></label>
            <input v-model="roleForm.name" placeholder="例如：project_admin" :disabled="roleModalMode === 'edit'" />
          </div>
          <div class="form-group">
            <label>显示名称 <span class="required">*</span></label>
            <input v-model="roleForm.display_name" placeholder="例如：项目管理员" />
          </div>
          <div class="form-group">
            <label>角色类型</label>
            <select v-model="roleForm.role_type">
              <option value="platform_admin">平台管理员</option>
              <option value="cluster_admin">集群管理员</option>
              <option value="developer">开发者</option>
              <option value="viewer">只读用户</option>
              <option value="custom">自定义</option>
            </select>
          </div>
          <div class="form-group">
            <label>描述</label>
            <textarea v-model="roleForm.description" rows="3" placeholder="角色描述..."></textarea>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>颜色</label>
              <input type="color" v-model="roleForm.color" />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRoleModal = false">取消</button>
          <button class="btn btn-primary" @click="saveRole" :disabled="submitting">
            {{ submitting ? '保存中...' : '确认' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 权限配置弹窗（大厂风格） -->
    <div v-if="showPermModal" class="modal-mask" @click="showPermModal = false">
      <div class="modal modal-perm" @click.stop>
        <div class="modal-header">
          <h3>🔐 权限配置 - {{ currentRole?.display_name }}</h3>
          <button class="btn-close" @click="showPermModal = false">✕</button>
        </div>
        <div class="modal-body perm-body">
          <!-- 权限模块分组 -->
          <div v-for="module in permissionModules" :key="module.key" class="perm-module">
            <div class="module-header" @click="module.collapsed = !module.collapsed">
              <span class="module-icon">{{ module.icon }}</span>
              <span class="module-name">{{ module.name }}</span>
              <span class="module-count">{{ getModuleSelectedCount(module) }}/{{ module.permissions.length }}</span>
              <span class="collapse-icon">{{ module.collapsed ? '▶' : '▼' }}</span>
            </div>
            <div v-show="!module.collapsed" class="module-content">
              <div class="module-desc">{{ module.description }}</div>
              <div class="perm-grid">
                <div
                  v-for="perm in module.permissions"
                  :key="perm.id"
                  :class="['perm-item', { selected: selectedPermIds.includes(perm.id) }]"
                  @click="togglePermission(perm.id)"
                >
                  <span class="perm-check">{{ selectedPermIds.includes(perm.id) ? '✓' : '' }}</span>
                  <div class="perm-info">
                    <span class="perm-name">{{ perm.display_name }}</span>
                    <span class="perm-code">{{ perm.name }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 快捷操作 -->
          <div class="quick-actions">
            <span class="quick-label">快捷操作：</span>
            <button class="btn-quick" @click="selectAllPerms">全选</button>
            <button class="btn-quick" @click="clearAllPerms">清空</button>
            <button class="btn-quick" @click="selectViewPerms">仅查看</button>
          </div>
        </div>
        <div class="modal-footer">
          <div class="perm-summary">
            已选 <strong>{{ selectedPermIds.length }}</strong> 个权限
          </div>
          <button class="btn btn-secondary" @click="showPermModal = false">取消</button>
          <button class="btn btn-primary" @click="savePermissions" :disabled="permSubmitting">
            {{ permSubmitting ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 绑定用户弹窗 -->
    <div v-if="showUsersModal" class="modal-mask" @click="showUsersModal = false">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>👥 绑定用户 - {{ currentRole?.display_name }}</h3>
          <button class="btn-close" @click="showUsersModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="roleUsers.length === 0" class="empty-users">
            暂无绑定用户
          </div>
          <div v-else class="user-list">
            <div v-for="user in roleUsers" :key="user.id" class="user-item">
              <div class="user-avatar">{{ user.username?.charAt(0)?.toUpperCase() }}</div>
              <div class="user-info">
                <span class="user-name">{{ user.username }}</span>
                <span class="user-email">{{ user.email || '-' }}</span>
              </div>
              <span :class="['status-tag', user.status === 1 ? 'active' : 'inactive']">
                {{ user.status === 1 ? '激活' : '禁用' }}
              </span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showUsersModal = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive } from 'vue'
import { getRoleList, getAllRoles, createRole, updateRole, deleteRole as deleteRoleApi, getPermissionList, getRolePermissions, updateRolePermissions, getRoleUsers } from '@/api/rbac'
import { getClusterList } from '@/api/cluster'
import { listRoles } from '@/api/k8sRbac'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

const activeTab = ref('platform')
const roles = ref([])
const clusters = ref([])
const searchQuery = ref('')
const loading = ref(false)
const submitting = ref(false)

// K8s 角色
const selectedClusterId = ref('')
const k8sRoleType = ref('')
const k8sRoles = ref([])

// 弹窗
const showRoleModal = ref(false)
const roleModalMode = ref('create')
const roleForm = ref({
  id: 0,
  name: '',
  display_name: '',
  role_type: 'developer',
  description: '',
  color: '#326ce5'
})

// 权限配置
const showPermModal = ref(false)
const currentRole = ref(null)
const allPermissions = ref([])
const selectedPermIds = ref([])
const permSubmitting = ref(false)

// 用户列表
const showUsersModal = ref(false)
const roleUsers = ref([])

// 权限模块分组（5大模块）
const permissionModules = reactive([
  {
    key: 'platform',
    name: '平台权限',
    icon: '🏠',
    description: '控制菜单可见性、系统功能、平台管理能力',
    collapsed: false,
    permissions: []
  },
  {
    key: 'k8s',
    name: 'K8s资源权限',
    icon: '☸️',
    description: '控制集群资源操作：Pod/Deployment/Service/Namespace等',
    collapsed: false,
    permissions: []
  },
  {
    key: 'cicd',
    name: 'CI/CD权限',
    icon: '⚡',
    description: '控制流水线、发布、审批、回滚等能力',
    collapsed: false,
    permissions: []
  },
  {
    key: 'image',
    name: '镜像与环境',
    icon: '📦',
    description: '控制镜像仓库、环境配置的访问权限',
    collapsed: true,
    permissions: []
  },
  {
    key: 'security',
    name: '安全权限',
    icon: '🛡️',
    description: '控制用户管理、角色管理、审计日志等',
    collapsed: true,
    permissions: []
  }
])

const systemRoles = computed(() => (roles.value || []).filter(r => r.is_system).length)

const filteredRoles = computed(() => {
  const list = roles.value || []
  if (!searchQuery.value) return list
  const q = searchQuery.value.toLowerCase()
  return list.filter(r =>
    r.name?.toLowerCase().includes(q) ||
    r.display_name?.toLowerCase().includes(q)
  )
})

const loadRoles = async () => {
  loading.value = true
  try {
    const res = await getAllRoles()
    if (res.code === 0) {
      roles.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载角色失败', e)
    roles.value = []
  } finally {
    loading.value = false
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

const loadK8sRoles = async () => {
  if (!selectedClusterId.value) return
  try {
    const res = await listRoles(selectedClusterId.value, '', k8sRoleType.value)
    if (res.code === 0) {
      k8sRoles.value = res.data || []
    }
  } catch (e) {
    console.error('加载K8s角色失败', e)
  }
}

const openCreateModal = () => {
  roleModalMode.value = 'create'
  roleForm.value = {
    id: 0,
    name: '',
    display_name: '',
    role_type: 'developer',
    description: '',
    color: '#326ce5'
  }
  showRoleModal.value = true
}

const editRole = (role) => {
  roleModalMode.value = 'edit'
  roleForm.value = { ...role }
  showRoleModal.value = true
}

const saveRole = async () => {
  if (!roleForm.value.name || !roleForm.value.display_name) {
    alert('请填写必填项')
    return
  }
  submitting.value = true
  try {
    const api = roleModalMode.value === 'create' ? createRole : updateRole
    const res = await api(roleForm.value)
    if (res.code === 0) {
      showRoleModal.value = false
      loadRoles()
    }
  } catch (e) {
    console.error('保存失败', e)
  } finally {
    submitting.value = false
  }
}

const deleteRole = async (role) => {
  const ok = await showConfirm({
    title: '确认删除角色',
    content: '删除后关联用户将失去该角色权限。',
    type: 'danger',
    details: [
      { label: '角色名称', value: role.display_name || role.name },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  try {
    const res = await deleteRoleApi(role.id)
    if (res.code === 0) {
      loadRoles()
    }
  } catch (e) {
    console.error('删除失败', e)
  }
}

const getRoleTypeLabel = (type) => {
  const labels = {
    super_admin: '超级管理员',
    platform_admin: '平台管理员',
    cluster_admin: '集群管理员',
    cicd_admin: 'CI/CD管理员',
    developer: '开发者',
    viewer: '只读用户',
    custom: '自定义'
  }
  return labels[type] || type
}

const formatTime = (ts) => {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleDateString()
}

const viewRoleUsers = async (role) => {
  currentRole.value = role
  roleUsers.value = []
  showUsersModal.value = true
  try {
    const res = await getRoleUsers(role.id)
    if (res.code === 0) {
      roleUsers.value = res.data?.list || res.data || []
    }
  } catch (e) {
    console.error('加载用户列表失败', e)
  }
}

// 打开权限配置弹窗
const openPermissionModal = async (role) => {
  currentRole.value = role
  selectedPermIds.value = []
  showPermModal.value = true
  
  // 加载所有权限
  if (allPermissions.value.length === 0) {
    try {
      const res = await getPermissionList()
      if (res.code === 0) {
        allPermissions.value = res.data?.list || res.data || []
        categorizePermissions()
      }
    } catch (e) {
      console.error('加载权限列表失败', e)
    }
  }
  
  // 加载角色已有权限
  try {
    const res = await getRolePermissions(role.id)
    if (res.code === 0) {
      const perms = res.data?.list || res.data || []
      selectedPermIds.value = perms.map(p => p.id || p.permission_id)
    }
  } catch (e) {
    console.error('加载角色权限失败', e)
  }
}

// 分类权限到各模块
const categorizePermissions = () => {
  const moduleMap = {
    platform: ['platform', 'cluster', 'health'],
    k8s: ['namespace', 'deployment', 'pod', 'service', 'configmap', 'secret', 'ingress', 'pvc', 'statefulset', 'daemonset', 'job', 'cronjob'],
    cicd: ['pipeline', 'release', 'approval', 'build'],
    image: ['image', 'registry', 'environment', 'env'],
    security: ['user', 'role', 'permission', 'audit', 'rbac']
  }
  
  permissionModules.forEach(module => {
    module.permissions = allPermissions.value.filter(p => {
      const resourceType = p.resource_type || p.name?.split(':')[0] || ''
      return moduleMap[module.key]?.some(k => resourceType.includes(k))
    })
  })
}

// 切换权限
const togglePermission = (permId) => {
  const idx = selectedPermIds.value.indexOf(permId)
  if (idx >= 0) {
    selectedPermIds.value.splice(idx, 1)
  } else {
    selectedPermIds.value.push(permId)
  }
}

// 获取模块已选数量
const getModuleSelectedCount = (module) => {
  return module.permissions.filter(p => selectedPermIds.value.includes(p.id)).length
}

// 全选
const selectAllPerms = () => {
  selectedPermIds.value = allPermissions.value.map(p => p.id)
}

// 清空
const clearAllPerms = () => {
  selectedPermIds.value = []
}

// 仅查看权限
const selectViewPerms = () => {
  selectedPermIds.value = allPermissions.value
    .filter(p => p.action === 'view' || p.name?.includes(':view'))
    .map(p => p.id)
}

// 保存权限配置
const savePermissions = async () => {
  if (!currentRole.value) return
  permSubmitting.value = true
  try {
    const res = await updateRolePermissions({
      role_id: currentRole.value.id,
      permission_ids: selectedPermIds.value
    })
    if (res.code === 0) {
      showPermModal.value = false
      loadRoles()
    }
  } catch (e) {
    console.error('保存权限失败', e)
  } finally {
    permSubmitting.value = false
  }
}

const viewK8sRoleDetail = (role) => {
  alert(`查看 K8s 角色 "${role.name}" 详情（功能开发中）`)
}

onMounted(() => {
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

.tab-nav {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  padding: 4px;
  background: white;
  border-radius: 8px;
  width: fit-content;
}

.tab-btn {
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

.tab-btn.active {
  background: #3b82f6;
  color: white;
}

.tab-content {
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

.toolbar-left {
  display: flex;
  gap: 12px;
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
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.role-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.role-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 16px;
  transition: all 0.2s;
}

.role-card:hover {
  border-color: #94a3b8;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.role-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.role-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.role-info {
  flex: 1;
}

.role-name {
  font-weight: 600;
  font-size: 15px;
  display: block;
}

.role-type {
  font-size: 12px;
  color: #64748b;
}

.role-actions {
  display: flex;
  gap: 4px;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  padding: 4px;
  border-radius: 4px;
}

.btn-icon:hover {
  background: #e2e8f0;
}

.system-badge {
  padding: 2px 8px;
  background: #e2e8f0;
  border-radius: 4px;
  font-size: 11px;
  color: #64748b;
}

.role-desc {
  color: #64748b;
  font-size: 13px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.role-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.role-footer {
  display: flex;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.btn-text {
  background: none;
  border: none;
  color: #3b82f6;
  cursor: pointer;
  font-size: 13px;
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

.k8s-role-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.k8s-role-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.k8s-role-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.type-badge {
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.type-badge.cluster {
  background: #fef3c7;
  color: #92400e;
}

.type-badge.namespace {
  background: #dbeafe;
  color: #1e40af;
}

.k8s-role-name {
  font-weight: 500;
}

.namespace-tag {
  padding: 2px 6px;
  background: #e2e8f0;
  border-radius: 3px;
  font-size: 11px;
  color: #64748b;
}

.k8s-role-rules {
  display: flex;
  align-items: center;
  gap: 16px;
}

.rules-count {
  font-size: 13px;
  color: #64748b;
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
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  font-size: 14px;
}

.required {
  color: #ef4444;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-row .form-group {
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

/* 权限配置弹窗 */
.modal-perm {
  width: 720px;
  max-height: 85vh;
}

.perm-body {
  max-height: 60vh;
  overflow-y: auto;
  padding: 20px 24px;
}

.perm-module {
  margin-bottom: 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
}

.module-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #f8fafc;
  cursor: pointer;
  user-select: none;
}

.module-header:hover {
  background: #f1f5f9;
}

.module-icon {
  font-size: 18px;
}

.module-name {
  font-weight: 600;
  flex: 1;
}

.module-count {
  font-size: 12px;
  color: #64748b;
  background: #e2e8f0;
  padding: 2px 8px;
  border-radius: 10px;
}

.collapse-icon {
  font-size: 10px;
  color: #94a3b8;
}

.module-content {
  padding: 16px;
}

.module-desc {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 12px;
}

.perm-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.perm-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.perm-item:hover {
  border-color: #94a3b8;
  background: #f8fafc;
}

.perm-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.perm-check {
  width: 18px;
  height: 18px;
  border: 2px solid #d1d5db;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.perm-item.selected .perm-check {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.perm-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.perm-name {
  font-size: 13px;
  font-weight: 500;
}

.perm-code {
  font-size: 11px;
  color: #94a3b8;
  font-family: monospace;
}

.quick-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 8px;
  margin-top: 16px;
}

.quick-label {
  font-size: 13px;
  color: #64748b;
}

.btn-quick {
  padding: 4px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 12px;
}

.btn-quick:hover {
  background: #f1f5f9;
}

.perm-summary {
  flex: 1;
  font-size: 13px;
  color: #64748b;
}

.perm-summary strong {
  color: #3b82f6;
  font-size: 16px;
}

/* 用户列表弹窗 */
.empty-users {
  text-align: center;
  padding: 40px;
  color: #94a3b8;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 400px;
  overflow-y: auto;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: #f8fafc;
  border-radius: 8px;
}

.user-avatar {
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

.user-item .user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-name {
  font-weight: 500;
}

.user-email {
  font-size: 12px;
  color: #94a3b8;
}

.status-tag {
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
}

.status-tag.active {
  background: #dcfce7;
  color: #166534;
}

.status-tag.inactive {
  background: #fee2e2;
  color: #991b1b;
}

.btn-text.primary {
  font-weight: 600;
}
</style>

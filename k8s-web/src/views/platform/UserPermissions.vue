<template>
  <div class="permission-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>用户授权</h1>
        <p class="subtitle">为用户分配集群访问权限</p>
      </div>
      <button class="btn-add" @click="openAddModal">
        <span class="icon">+</span> 添加授权
      </button>
    </div>

    <!-- 快捷统计 -->
    <div class="quick-stats">
      <div class="stat-item">
        <span class="stat-num">{{ stats.totalUsers }}</span>
        <span class="stat-label">已授权用户</span>
      </div>
      <div class="stat-item">
        <span class="stat-num">{{ stats.totalPermissions }}</span>
        <span class="stat-label">授权记录</span>
      </div>
      <div class="stat-item">
        <span class="stat-num">{{ clusters.length }}</span>
        <span class="stat-label">可用集群</span>
      </div>
    </div>

    <!-- 搜索筛选 -->
    <div class="filter-row">
      <input 
        v-model="searchQuery" 
        type="text" 
        class="search-input" 
        placeholder="搜索用户..." 
      />
      <select v-model="clusterFilter" class="filter-select">
        <option value="">全部集群</option>
        <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name }}</option>
      </select>
    </div>

    <!-- 授权列表 -->
    <div class="permission-list">
      <div v-if="loading" class="loading-box">加载中...</div>
      
      <div v-else-if="filteredPermissions.length === 0" class="empty-box">
        <div class="empty-icon">📭</div>
        <p>暂无授权记录</p>
        <button class="btn-link" @click="openAddModal">添加第一条授权</button>
      </div>

      <div v-else class="perm-grid">
        <div v-for="perm in filteredPermissions" :key="perm.id" class="perm-card">
          <div class="card-header">
            <div class="user-avatar">{{ perm.username?.charAt(0)?.toUpperCase() || 'U' }}</div>
            <div class="user-info">
              <div class="user-name">{{ perm.username }}</div>
              <div class="cluster-name">{{ perm.cluster_name }}</div>
            </div>
            <span class="role-tag" :class="perm.role_type">{{ getRoleLabel(perm.role_type) }}</span>
          </div>
          <div class="card-footer">
            <button class="btn-text" @click="editPermission(perm)">编辑</button>
            <button class="btn-text danger" @click="deletePermission(perm)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加/编辑弹窗 -->
    <div v-if="showModal" class="modal-mask" @click.self="closeModal">
      <div class="modal-box">
        <div class="modal-head">
          <h3>{{ editMode ? '编辑授权' : '添加授权' }}</h3>
          <button class="close-x" @click="closeModal">×</button>
        </div>
        
        <div class="modal-content">
          <!-- 用户选择 -->
          <div class="form-item">
            <label>用户</label>
            <select v-model="form.user_id" :disabled="editMode" class="form-select">
              <option value="">选择用户</option>
              <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}</option>
            </select>
          </div>

          <!-- 集群选择 -->
          <div class="form-item">
            <label>集群</label>
            <select v-model="form.cluster_id" :disabled="editMode" class="form-select">
              <option value="">选择集群</option>
              <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name }}</option>
            </select>
          </div>

          <!-- 角色卡片选择 -->
          <div class="form-item">
            <label>角色</label>
            <div class="role-cards">
              <div 
                class="role-card" 
                :class="{ active: form.role_type === 'cluster_admin' }"
                @click="selectRole('cluster_admin')"
              >
                <div class="role-icon admin">👑</div>
                <div class="role-name">管理员</div>
                <div class="role-desc">完整集群权限</div>
              </div>
              <div 
                class="role-card" 
                :class="{ active: form.role_type === 'developer' }"
                @click="selectRole('developer')"
              >
                <div class="role-icon dev">💻</div>
                <div class="role-name">开发者</div>
                <div class="role-desc">创建和管理工作负载</div>
              </div>
              <div 
                class="role-card" 
                :class="{ active: form.role_type === 'viewer' }"
                @click="selectRole('viewer')"
              >
                <div class="role-icon viewer">👁️</div>
                <div class="role-name">只读</div>
                <div class="role-desc">仅查看资源</div>
              </div>
            </div>
          </div>

          <!-- 权限预览 -->
          <div v-if="form.role_type" class="perm-preview">
            <div class="preview-title">权限预览</div>
            <div class="preview-items">
              <span class="preview-item" :class="{ enabled: form.can_view }">查看</span>
              <span class="preview-item" :class="{ enabled: form.can_create }">创建</span>
              <span class="preview-item" :class="{ enabled: form.can_update }">更新</span>
              <span class="preview-item" :class="{ enabled: form.can_delete }">删除</span>
              <span class="preview-item" :class="{ enabled: form.can_exec }">终端</span>
            </div>
          </div>

          <div v-if="formError" class="form-error">{{ formError }}</div>
        </div>

        <div class="modal-foot">
          <button class="btn-cancel" @click="closeModal">取消</button>
          <button 
            class="btn-confirm" 
            @click="submitForm"
            :disabled="saving || !form.user_id || !form.cluster_id || !form.role_type"
          >
            {{ saving ? '保存中...' : '确认' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div v-if="showDeleteModal" class="modal-mask" @click.self="showDeleteModal = false">
      <div class="modal-box small">
        <div class="modal-head danger">
          <h3>删除授权</h3>
          <button class="close-x" @click="showDeleteModal = false">×</button>
        </div>
        <div class="modal-content">
          <p class="confirm-text">
            确定删除 <strong>{{ deleteTarget?.username }}</strong> 在 
            <strong>{{ deleteTarget?.cluster_name }}</strong> 的授权？
          </p>
        </div>
        <div class="modal-foot">
          <button class="btn-cancel" @click="showDeleteModal = false">取消</button>
          <button class="btn-danger" @click="confirmDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getUserList } from '@/api/user'
import { getClusterList } from '@/api/cluster'
import {
  getClusterPermissionList,
  createClusterPermission,
  updateClusterPermission,
  deleteClusterPermission
} from '@/api/rbac'

// 状态
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)

// 数据
const users = ref([])
const clusters = ref([])
const permissions = ref([])

// 筛选
const searchQuery = ref('')
const clusterFilter = ref('')

// 弹窗
const showModal = ref(false)
const showDeleteModal = ref(false)
const editMode = ref(false)
const deleteTarget = ref(null)
const formError = ref('')

// 表单
const form = ref({
  id: 0,
  user_id: '',
  cluster_id: '',
  role_type: '',
  can_view: true,
  can_create: false,
  can_update: false,
  can_delete: false,
  can_exec: false
})

// 统计
const stats = computed(() => {
  const userSet = new Set(permissions.value.map(p => p.user_id))
  return {
    totalUsers: userSet.size,
    totalPermissions: permissions.value.length
  }
})

// 过滤
const filteredPermissions = computed(() => {
  return permissions.value.filter(p => {
    const matchSearch = !searchQuery.value || 
      p.username?.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchCluster = !clusterFilter.value || p.cluster_id === clusterFilter.value
    return matchSearch && matchCluster
  })
})

// 角色标签
const getRoleLabel = (role) => {
  const labels = {
    cluster_admin: '管理员',
    developer: '开发者',
    viewer: '只读',
    custom: '自定义',
    cicd_admin: 'CI/CD管理员'
  }
  return labels[role] || role
}

// 选择角色
const selectRole = (role) => {
  form.value.role_type = role
  if (role === 'cluster_admin') {
    form.value.can_view = true
    form.value.can_create = true
    form.value.can_update = true
    form.value.can_delete = true
    form.value.can_exec = true
  } else if (role === 'developer') {
    form.value.can_view = true
    form.value.can_create = true
    form.value.can_update = true
    form.value.can_delete = false
    form.value.can_exec = true
  } else if (role === 'viewer') {
    form.value.can_view = true
    form.value.can_create = false
    form.value.can_update = false
    form.value.can_delete = false
    form.value.can_exec = false
  }
}

// 加载数据
const loadUsers = async () => {
  try {
    const res = await getUserList({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      users.value = res.data.list || res.data || []
    }
  } catch (e) {
    console.error('加载用户失败:', e)
  }
}

const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0 && res.data) {
      clusters.value = res.data.list || res.data || []
    }
  } catch (e) {
    console.error('加载集群失败:', e)
  }
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const res = await getClusterPermissionList({ page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      const list = Array.isArray(res.data.list) ? res.data.list : (Array.isArray(res.data) ? res.data : [])
      permissions.value = list.map(p => {
        const user = users.value.find(u => u.id === p.user_id)
        const cluster = clusters.value.find(c => c.id === p.cluster_id)
        return {
          ...p,
          username: user?.username || `用户${p.user_id}`,
          cluster_name: cluster?.cluster_name || `集群${p.cluster_id}`
        }
      })
    }
  } catch (e) {
    console.error('加载权限失败:', e)
  } finally {
    loading.value = false
  }
}

// 弹窗操作
const openAddModal = () => {
  editMode.value = false
  form.value = {
    id: 0,
    user_id: '',
    cluster_id: '',
    role_type: '',
    can_view: true,
    can_create: false,
    can_update: false,
    can_delete: false,
    can_exec: false
  }
  formError.value = ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  formError.value = ''
}

const editPermission = (perm) => {
  editMode.value = true
  form.value = {
    id: perm.id,
    user_id: perm.user_id,
    cluster_id: perm.cluster_id,
    role_type: perm.role_type || 'viewer',
    can_view: perm.can_view,
    can_create: perm.can_create,
    can_update: perm.can_update,
    can_delete: perm.can_delete,
    can_exec: perm.can_exec
  }
  formError.value = ''
  showModal.value = true
}

const deletePermission = (perm) => {
  deleteTarget.value = perm
  showDeleteModal.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const res = await deleteClusterPermission(deleteTarget.value.id)
    if (res.code === 0) {
      Message.success('删除成功')
      showDeleteModal.value = false
      deleteTarget.value = null
      await loadPermissions()
    } else {
      Message.error(res.msg || '删除失败')
    }
  } catch (e) {
    Message.error('删除失败')
  } finally {
    deleting.value = false
  }
}

// 提交表单
const submitForm = async () => {
  formError.value = ''
  if (!form.value.user_id || !form.value.cluster_id || !form.value.role_type) {
    formError.value = '请完整填写表单'
    return
  }

  saving.value = true
  try {
    const data = {
      user_id: Number(form.value.user_id),
      cluster_id: Number(form.value.cluster_id),
      role_type: form.value.role_type,
      can_view: form.value.can_view,
      can_create: form.value.can_create,
      can_update: form.value.can_update,
      can_delete: form.value.can_delete,
      can_exec: form.value.can_exec,
      namespaces: []
    }

    let res
    if (editMode.value) {
      data.id = form.value.id
      res = await updateClusterPermission(data)
    } else {
      res = await createClusterPermission(data)
    }

    if (res.code === 0) {
      Message.success(editMode.value ? '修改成功' : '授权成功')
      closeModal()
      await loadPermissions()
    } else {
      formError.value = res.msg || '操作失败'
    }
  } catch (e) {
    formError.value = '操作失败'
  } finally {
    saving.value = false
  }
}

// 初始化
onMounted(async () => {
  await Promise.all([loadUsers(), loadClusters()])
  await loadPermissions()
})
</script>

<style scoped>
.permission-page {
  padding: 0;
  min-height: 100%;
}

/* 头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.subtitle {
  color: #64748b;
  font-size: 14px;
  margin: 4px 0 0;
}

.btn-add {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-add:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-add .icon {
  font-size: 18px;
  font-weight: 300;
}

/* 统计 */
.quick-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
}

.stat-num {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
}

/* 筛选 */
.filter-row {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.search-input {
  flex: 1;
  max-width: 280px;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

/* 列表 */
.permission-list {
  background: white;
  border-radius: 12px;
  padding: 20px;
  min-height: 300px;
}

.loading-box, .empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #64748b;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.btn-link {
  color: #3b82f6;
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  margin-top: 8px;
}

/* 卡片网格 */
.perm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.perm-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 16px;
  transition: all 0.2s;
}

.perm-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
}

.user-info {
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: #1e293b;
  font-size: 15px;
}

.cluster-name {
  color: #64748b;
  font-size: 13px;
}

.role-tag {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.role-tag.cluster_admin {
  background: #fef3c7;
  color: #d97706;
}

.role-tag.developer {
  background: #dbeafe;
  color: #2563eb;
}

.role-tag.viewer {
  background: #f1f5f9;
  color: #64748b;
}

.role-tag.cicd_admin {
  background: #ede9fe;
  color: #7c3aed;
}

.card-footer {
  display: flex;
  gap: 16px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.btn-text {
  background: none;
  border: none;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  padding: 0;
}

.btn-text:hover {
  color: #3b82f6;
}

.btn-text.danger:hover {
  color: #ef4444;
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-box {
  background: white;
  border-radius: 16px;
  width: 480px;
  max-width: 90vw;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-box.small {
  width: 400px;
}

.modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-head.danger {
  background: #fef2f2;
}

.modal-head h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.close-x {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: #f1f5f9;
  color: #64748b;
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-x:hover {
  background: #e2e8f0;
}

.modal-content {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.form-item {
  margin-bottom: 20px;
}

.form-item label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 8px;
}

.form-select {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-select:disabled {
  background: #f8fafc;
  cursor: not-allowed;
}

/* 角色卡片 */
.role-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.role-card {
  padding: 16px 12px;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.role-card:hover {
  border-color: #cbd5e1;
}

.role-card.active {
  border-color: #3b82f6;
  background: #eff6ff;
}

.role-icon {
  font-size: 28px;
  margin-bottom: 8px;
}

.role-name {
  font-weight: 600;
  color: #1e293b;
  font-size: 14px;
  margin-bottom: 4px;
}

.role-desc {
  font-size: 11px;
  color: #64748b;
  line-height: 1.3;
}

/* 权限预览 */
.perm-preview {
  background: #f8fafc;
  border-radius: 8px;
  padding: 12px 16px;
  margin-top: 16px;
}

.preview-title {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 8px;
}

.preview-items {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.preview-item {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  background: #e2e8f0;
  color: #94a3b8;
}

.preview-item.enabled {
  background: #dcfce7;
  color: #16a34a;
}

.form-error {
  background: #fef2f2;
  color: #dc2626;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 14px;
  margin-top: 16px;
}

.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

.btn-cancel {
  padding: 10px 20px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  color: #64748b;
  font-size: 14px;
  cursor: pointer;
}

.btn-cancel:hover {
  background: #f8fafc;
}

.btn-confirm {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.btn-confirm:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-danger {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #ef4444;
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.confirm-text {
  color: #374151;
  line-height: 1.6;
}

.confirm-text strong {
  color: #1e293b;
}
</style>

<template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>用户管理</h1>
      <p>系统用户列表与权限管理</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input 
          type="text" 
          v-model="searchQuery" 
          placeholder="搜索用户名..."
          @input="onSearchInput"
        />
      </div>

      <div class="filter-dropdown">
        <select v-model="roleFilter">
          <option value="">所有角色</option>
          <option value="admin">管理员</option>
          <option value="user">普通用户</option>
        </select>
      </div>

      <div class="filter-dropdown">
        <select v-model="statusFilter">
          <option value="">所有状态</option>
          <option value="active">激活</option>
          <option value="inactive">禁用</option>
        </select>
      </div>

      <div class="action-buttons">
        <button 
          v-if="!batchMode" 
          class="btn btn-batch" 
          @click="enterBatchMode"
        >
          ☑️ 批量操作
        </button>
        <button 
          v-if="batchMode" 
          class="btn btn-secondary" 
          @click="exitBatchMode"
        >
          ✖️ 退出批量
        </button>

        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>

        <button class="btn btn-primary" @click="showCreateModal = true">创建用户</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedUsers.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedUsers.length }} 个用户</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="batchDelete" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="loading && users.length === 0" class="loading-state">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input 
                type="checkbox" 
                :checked="isAllSelected" 
                @change="toggleSelectAll"
                title="全选/取消全选"
              />
            </th>
            <th style="width: 100px;">ID</th>
            <th style="min-width: 150px;">用户名</th>
            <th style="width: 120px;">角色</th>
            <th style="width: 100px;">状态</th>
            <th style="width: 170px;">创建时间</th>
            <th style="width: 120px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="user in paginatedUsers" 
            :key="user.id"
            :class="{ 'row-selected': isUserSelected(user) }"
          >
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isUserSelected(user)" 
                @change="toggleUserSelection(user)"
              />
            </td>
            <td>{{ user.id }}</td>
            <td>
              <div class="user-name">
                <span class="icon">👤</span>
                <span>{{ user.username }}</span>
              </div>
            </td>
            <td>
              <span class="role-badge" :class="user.role">
                {{ user.role === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </td>
            <td>
              <span class="status-indicator" :class="user.status === 1 ? 'active' : 'inactive'">
                {{ user.status === 1 ? '激活' : '禁用' }}
              </span>
            </td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td class="actions">
              <button 
                v-if="user.status === 1" 
                class="btn btn-sm btn-secondary" 
                @click="handleToggleStatus(user)" 
                title="禁用账号"
              >
                🚫
              </button>
              <button 
                v-else 
                class="btn btn-sm btn-success" 
                @click="handleToggleStatus(user)" 
                title="激活账号"
              >
                ✅
              </button>
              <button class="btn btn-sm btn-warning" @click="handleEdit(user)" title="编辑">
                ✏️
              </button>
              <button class="btn btn-sm btn-danger" @click="handleDelete(user)" title="删除">
                🗑️
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页控件（增强版） -->
    <div class="pagination">
      <div class="pagination-info">
        共 {{ totalItems }} 条记录，当前第 {{ currentPage }}/{{ totalPages }} 页
        <select v-model.number="itemsPerPage" @change="onPageSizeChange" class="page-size-select">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
      </div>
      <div class="pagination-controls">
        <button @click="goToPage(1)" :disabled="currentPage === 1">首页</button>
        <button @click="goToPage(currentPage - 1)" :disabled="currentPage === 1">上一页</button>
        <input 
          v-model.number="jumpPage" 
          type="number" 
          min="1" 
          :max="totalPages" 
          placeholder="页码" 
          @keyup.enter="jumpToPage" 
        />
        <button @click="goToPage(currentPage + 1)" :disabled="currentPage === totalPages">下一页</button>
        <button @click="goToPage(totalPages)" :disabled="currentPage === totalPages">尾页</button>
      </div>
    </div>

    <!-- 创建用户模态框 -->
    <div v-if="showCreateModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>创建用户</h2>
          <button class="close-btn" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="handleCreateUser">
            <div class="form-group">
              <label for="create-username">用户名</label>
              <input
                id="create-username"
                v-model="userForm.username"
                type="text"
                required
              />
            </div>
            <div class="form-group">
              <label for="create-password">密码</label>
              <input
                id="create-password"
                v-model="userForm.password"
                type="password"
                required
              />
            </div>
            <div class="form-group">
              <label for="create-role">角色</label>
              <select id="create-role" v-model="userForm.role" required>
                <option value="admin">管理员</option>
                <option value="user">普通用户</option>
              </select>
            </div>
            <div class="form-group">
              <label for="create-status">状态</label>
              <select id="create-status" v-model="userForm.status" required>
                <option :value="1">激活</option>
                <option :value="0">禁用</option>
              </select>
            </div>
            <div class="form-actions">
              <button type="button" class="cancel-btn" @click="showCreateModal = false">取消</button>
              <button type="submit" class="submit-btn">创建</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 编辑用户模态框 -->
    <div v-if="showEditModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>编辑用户</h2>
          <button class="close-btn" @click="showEditModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="handleUpdateUser">
            <div class="form-group">
              <label for="edit-username">用户名</label>
              <input
                id="edit-username"
                v-model="userForm.username"
                type="text"
                required
              />
            </div>
            <div class="form-group">
              <label for="edit-password">密码 (留空则不修改)</label>
              <input
                id="edit-password"
                v-model="userForm.password"
                type="password"
              />
            </div>
            <div class="form-group">
              <label for="edit-role">角色</label>
              <select id="edit-role" v-model="userForm.role" required>
                <option value="admin">管理员</option>
                <option value="user">普通用户</option>
              </select>
            </div>
            <div class="form-group">
              <label for="edit-status">状态</label>
              <select id="edit-status" v-model="userForm.status" required>
                <option :value="1">激活</option>
                <option :value="0">禁用</option>
              </select>
            </div>
            <div class="form-actions">
              <button type="button" class="cancel-btn" @click="showEditModal = false">取消</button>
              <button type="submit" class="submit-btn">更新</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import { getUserList, createUser as createUserApi, updateUser as updateUserApi, deleteUser as deleteUserApi } from '@/api/user'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// 加载状态
const loading = ref(true)
const errorMsg = ref('')
const autoRefresh = ref(false)
const batchMode = ref(false)
const selectedUsers = ref([])
const roleFilter = ref('')
const statusFilter = ref('')
const totalItems = ref(0)
const totalPages = ref(1)
const jumpPage = ref(1)
let refreshTimer = null
let searchTimeout = null

// 用户数据
const users = ref([])

// 搜索和分页
const searchQuery = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)

// 模态框状态
const showCreateModal = ref(false)
const showEditModal = ref(false)

// 表单数据
const userForm = ref({
  id: null,
  username: '',
  password: '',
  role: 'user',
  status: 1
})

// 监听分页和筛选变化
watch([currentPage, itemsPerPage, roleFilter, statusFilter], () => {
  refreshList()
})

// 过滤后的用户（现在由后端处理，这里直接返回）
const filteredUsers = computed(() => {
  return users.value
})

// 分页后的用户（现在由后端分页，直接返回）
const paginatedUsers = computed(() => {
  return users.value
})

// 处理编辑
const handleEdit = (user) => {
  userForm.value = { 
    id: user.id,
    username: user.username,
    password: '', // 不回显密码
    role: user.role || 'user',
    status: user.status
  }
  showEditModal.value = true
}

const handleDelete = async (user) => {
  const ok = await showConfirm({
    title: '确认删除用户',
    content: '删除后该用户将无法登录系统。',
    type: 'danger',
    details: [
      { label: '用户名', value: user.username },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  
  try {
    loading.value = true
    await deleteUserApi(user.id)
    Message.success('删除成功')
    await refreshList()
  } catch (err) {
    Message.error('删除失败: ' + (err.msg || err.message))
  } finally {
    loading.value = false
  }
}

const handleCreateUser = async () => {
  try {
    loading.value = true
    await createUserApi({
      username: userForm.value.username,
      password: userForm.value.password,
      role: userForm.value.role,
      status: userForm.value.status
    })
    showCreateModal.value = false
    resetForm()
    Message.success('用户创建成功')
    await refreshList()
  } catch (err) {
    Message.error('创建失败: ' + (err.msg || err.message))
  } finally {
    loading.value = false
  }
}

const handleUpdateUser = async () => {
  try {
    loading.value = true
    const updateData = {
      id: userForm.value.id,
      username: userForm.value.username,
      role: userForm.value.role,
      status: userForm.value.status
    }
    // 只有填写了密码才更新
    if (userForm.value.password) {
      updateData.password = userForm.value.password
    }
    await updateUserApi(updateData)
    showEditModal.value = false
    resetForm()
    Message.success('用户更新成功')
    await refreshList()
  } catch (err) {
    Message.error('更新失败: ' + (err.msg || err.message))
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  userForm.value = {
    id: null,
    username: '',
    password: '',
    role: 'user',
    status: 1
  }
}

// 刷新列表 - 调用真实后端接口
const refreshList = async () => {
  loading.value = true
  errorMsg.value = ''

  try {
    // 处理状态筛选值
    let statusValue = ''
    if (statusFilter.value === 'active') {
      statusValue = '1'
    } else if (statusFilter.value === 'inactive') {
      statusValue = '0'
    }

    const res = await getUserList({
      page: currentPage.value,
      limit: itemsPerPage.value,
      username: searchQuery.value,
      role: roleFilter.value,
      status: statusValue
    })
    
    // http拦截器返回格式: { code: 0, msg: "OK", data: { list: [...], total: n } }
    if (res.code === 0 && res.data) {
      if (res.data.list) {
        users.value = res.data.list
        totalItems.value = res.data.total || res.data.list.length
      } else if (Array.isArray(res.data)) {
        users.value = res.data
        totalItems.value = res.data.length
      } else {
        users.value = []
        totalItems.value = 0
      }
      totalPages.value = Math.ceil(totalItems.value / itemsPerPage.value) || 1
    } else {
      throw new Error(res.msg || '获取数据失败')
    }
  } catch (err) {
    errorMsg.value = '获取用户列表失败: ' + (err.msg || err.message || '未知错误')
    Message.error(errorMsg.value)
    users.value = []
  } finally {
    loading.value = false
  }
}

// 日期格式化
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  // 处理各种日期格式
  if (typeof dateStr === 'string') {
    return dateStr.replace('T', ' ').split('.')[0]
  }
  if (dateStr instanceof Date) {
    return dateStr.toLocaleString('zh-CN')
  }
  // 其他类型尝试转换
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return '-'
  }
}

// ==================== 批量操作 ====================
const enterBatchMode = () => {
  batchMode.value = true
  selectedUsers.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedUsers.value = []
}

const clearSelection = () => {
  selectedUsers.value = []
}

const isUserSelected = (user) => {
  return selectedUsers.value.some(u => u.id === user.id)
}

const toggleUserSelection = (user) => {
  const index = selectedUsers.value.findIndex(u => u.id === user.id)
  if (index === -1) {
    selectedUsers.value.push(user)
  } else {
    selectedUsers.value.splice(index, 1)
  }
}

const isAllSelected = computed(() => {
  return paginatedUsers.value.length > 0 && 
         paginatedUsers.value.every(u => isUserSelected(u))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedUsers.value = []
  } else {
    selectedUsers.value = [...paginatedUsers.value]
  }
}

const batchDelete = async () => {
  const ok = await showConfirm({
    title: '确认批量删除',
    content: `即将删除 ${selectedUsers.value.length} 个用户，此操作不可恢复。`,
    type: 'danger',
    details: [
      { label: '删除数量', value: `${selectedUsers.value.length} 个用户`, danger: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  
  try {
    loading.value = true
    // 调用真实API批量删除
    for (const user of selectedUsers.value) {
      await deleteUserApi(user.id)
    }
    Message.success(`成功删除 ${selectedUsers.value.length} 个用户`)
    selectedUsers.value = []
    batchMode.value = false
    await refreshList()
  } catch (err) {
    Message.error('批量删除失败: ' + (err.msg || err.message))
  } finally {
    loading.value = false
  }
}

// ==================== 分页操作 ====================
const goToPage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

const onPageSizeChange = () => {
  currentPage.value = 1
}

const jumpToPage = () => {
  if (jumpPage.value >= 1 && jumpPage.value <= totalPages.value) {
    currentPage.value = jumpPage.value
  }
}

const onSearchInput = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    refreshList()
  }, 300)
}

// ==================== 生命周期 ====================
onMounted(() => {
  refreshList()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
})
</script>

<style scoped>
/* ==================== Rancher/Kuboard 风格样式 ====================  */

/* 主容器 */
.resource-view {
  padding: 0;
}

/* 页面头部 */
.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #64748b;
  font-size: 14px;
  margin: 0;
}

/* 操作栏 */
.action-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 250px;
  max-width: 400px;
}

.search-box input {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-box input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-dropdown select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-dropdown select:hover {
  border-color: #cbd5e1;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-left: auto;
}

/* 按钮 */
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
  white-space: nowrap;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.btn-secondary:hover {
  background: #e2e8f0;
  border-color: #cbd5e1;
}

.btn-batch {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
}

.btn-batch:hover {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.4);
  transform: translateY(-1px);
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.btn-warning {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
}

.btn-warning:hover {
  background: linear-gradient(135deg, #d97706 0%, #b45309 100%);
  transform: translateY(-1px);
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.btn-danger:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

/* 自动刷新开关 */
.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
}

.auto-refresh-toggle:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.auto-refresh-toggle input[type="checkbox"] {
  cursor: pointer;
}

.refresh-indicator {
  color: #22c55e;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* 错误提示 */
.error-box {
  padding: 12px 16px;
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border-left: 4px solid #ef4444;
  border-radius: 8px;
  color: #991b1b;
  margin-bottom: 16px;
  font-size: 14px;
}

/* 批量操作浮动栏 */
.batch-action-bar {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #3b82f6;
  border-radius: 10px;
  margin-bottom: 16px;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
  color: #1e40af;
  font-size: 15px;
}

.batch-clear {
  padding: 6px 12px;
  background: white;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  color: #3b82f6;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-clear:hover {
  background: #3b82f6;
  color: white;
}

.batch-actions {
  display: flex;
  gap: 10px;
}

.batch-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
}

.batch-btn.danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 2px 6px rgba(239, 68, 68, 0.3);
}

.batch-btn.danger:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
}

/* 加载状态 */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #64748b;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 表格 */
.table-container {
  background: white;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
}

.resource-table thead {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 2px solid #e2e8f0;
}

.resource-table th {
  padding: 14px 16px;
  text-align: left;
  font-weight: 600;
  font-size: 13px;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-table tbody tr {
  border-bottom: 1px solid #f1f5f9;
  transition: background 0.15s;
}

.resource-table tbody tr:hover {
  background: #f8fafc;
}

.resource-table tbody tr.row-selected {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.resource-table td {
  padding: 14px 16px;
  font-size: 14px;
  color: #334155;
}

/* 用户名 */
.user-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.user-name .icon {
  font-size: 18px;
}

/* 角色徽章 */
.role-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.role-badge.admin {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #92400e;
  border: 1px solid #fbbf24;
}

.role-badge.user {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  color: #1e40af;
  border: 1px solid #60a5fa;
}

/* 状态指示器 */
.status-indicator {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator.active {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  color: #065f46;
  border: 1px solid #10b981;
}

.status-indicator.inactive {
  background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
  color: #991b1b;
  border: 1px solid #ef4444;
}

/* 操作按钮 */
.actions {
  display: flex;
  gap: 8px;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  flex-wrap: wrap;
  gap: 16px;
}

.pagination-info {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #64748b;
  font-size: 14px;
}

.page-size-select {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  font-size: 13px;
  cursor: pointer;
}

.pagination-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.pagination-controls button {
  padding: 8px 14px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.pagination-controls button:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.pagination-controls button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-controls input {
  width: 60px;
  padding: 8px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  text-align: center;
  font-size: 13px;
}

/* 模态框 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background-color: white;
  padding: 0;
  border-radius: 12px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #64748b;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #e2e8f0;
  color: #1e293b;
}

.modal-body {
  padding: 24px;
  max-height: calc(90vh - 180px);
  overflow-y: auto;
}

.modal-body form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 600;
  color: #334155;
  font-size: 14px;
}

.form-group input,
.form-group select {
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding: 20px 24px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

.cancel-btn,
.submit-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.cancel-btn:hover {
  background-color: #e2e8f0;
}

.submit-btn {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.submit-btn:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px);
}
</style>

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
              <span class="status-indicator" :class="user.status">
                {{ user.status === 'active' ? '激活' : '禁用' }}
              </span>
            </td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td class="actions">
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
          <form @submit.prevent="createUser">
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
                <option value="active">激活</option>
                <option value="inactive">禁用</option>
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
          <form @submit.prevent="updateUser">
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
                <option value="active">激活</option>
                <option value="inactive">禁用</option>
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
import { ref, computed } from 'vue'
import Pagination from '@/components/Pagination.vue'

// 模拟用户数据
const users = ref([
  {
    id: 1,
    username: 'admin',
    password: '123456',
    role: 'admin',
    status: 'active',
    createdAt: '2023-01-01'
  },
  {
    id: 2,
    username: 'user1',
    password: 'password1',
    role: 'user',
    status: 'active',
    createdAt: '2023-01-02'
  },
  {
    id: 3,
    username: 'user2',
    password: 'password2',
    role: 'user',
    status: 'inactive',
    createdAt: '2023-01-03'
  }
])

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
  status: 'active'
})

// 过滤后的用户
const filteredUsers = computed(() => {
  return users.value.filter(user => {
    return user.username.toLowerCase().includes(searchQuery.value.toLowerCase())
  })
})

// 分页后的用户
const paginatedUsers = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredUsers.value.slice(startIndex, endIndex)
})

// 处理编辑
const handleEdit = (user) => {
  userForm.value = { ...user }
  showEditModal.value = true
}

const handleDelete = async (user) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 ${user.username} 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // TODO: 连接后端 API
    const index = users.value.findIndex(u => u.id === user.id)
    if (index !== -1) {
      users.value.splice(index, 1)
    }

    ElMessage.success('删除成功')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const createUser = () => {
  // TODO: 连接后端 API
  const newUser = {
    ...userForm.value,
    id: Date.now(),
    created_at: new Date().toISOString()
  }
  users.value.push(newUser)
  showCreateModal.value = false
  resetForm()
  ElMessage.success('用户创建成功')
}

const updateUser = () => {
  // TODO: 连接后端 API
  const index = users.value.findIndex(u => u.id === userForm.value.id)
  if (index !== -1) {
    const updatedUser = { ...users.value[index], ...userForm.value }
    if (!userForm.value.password) {
      updatedUser.password = users.value[index].password
    }
    users.value[index] = updatedUser
    showEditModal.value = false
    resetForm()
    ElMessage.success('用户更新成功')
  }
}

const resetForm = () => {
  userForm.value = {
    id: null,
    username: '',
    password: '',
    role: 'user',
    status: 'active'
  }
}

// 刷新列表
const refreshList = async () => {
  loading.value = true
  errorMsg.value = ''

  try {
    // TODO: 连接后端 API
    // const response = await userApi.list({ page: currentPage.value, limit: itemsPerPage.value })
    // users.value = response.data.list

    // 模拟数据
    await new Promise(resolve => setTimeout(resolve, 500))
    if (users.value.length === 0) {
      users.value = [
        {
          id: 1,
          username: 'admin',
          password: '123456',
          role: 'admin',
          status: 'active',
          created_at: '2023-01-01 10:00:00'
        },
        {
          id: 2,
          username: 'user1',
          password: 'password1',
          role: 'user',
          status: 'active',
          created_at: '2023-01-02 11:30:00'
        },
        {
          id: 3,
          username: 'user2',
          password: 'password2',
          role: 'user',
          status: 'inactive',
          created_at: '2023-01-03 14:20:00'
        }
      ]
    }
  } catch (err) {
    errorMsg.value = '获取用户列表失败: ' + (err.message || '未知错误')
    ElMessage.error(errorMsg.value)
  } finally {
    loading.value = false
  }
}

// 日期格式化
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return dateStr.replace('T', ' ').split('.')[0]
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

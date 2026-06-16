<template>
  <div class="image-registry-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <span class="title-icon">📦</span>
          镜像仓库管理
        </h1>
        <p class="page-desc">统一管理 Docker、Harbor、ACR 等镜像仓库，支持连接检测和默认仓库设置</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-refresh" @click="loadData" :disabled="loading">
          <span class="btn-icon">🔄</span>
          刷新
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          <span class="btn-icon">+</span>
          添加仓库
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card total">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">仓库总数</div>
        </div>
      </div>
      <div class="stat-card connected">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.connected }}</div>
          <div class="stat-label">已连接</div>
        </div>
      </div>
      <div class="stat-card disconnected">
        <div class="stat-icon">❌</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.disconnected }}</div>
          <div class="stat-label">连接失败</div>
        </div>
      </div>
      <div class="stat-card types">
        <div class="stat-icon">🏷️</div>
        <div class="stat-content">
          <div class="stat-value">{{ Object.keys(stats.type_counts || {}).length }}</div>
          <div class="stat-label">仓库类型</div>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar-card">
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input 
          v-model="searchKeyword" 
          type="text" 
          placeholder="搜索仓库名称、URL..." 
          class="search-input"
          @input="handleSearch"
        />
        <button v-if="searchKeyword" class="clear-btn" @click="clearSearch">×</button>
      </div>
      <div class="filter-box">
        <select v-model="filterType" class="filter-select" @change="handleFilter">
          <option value="">全部类型</option>
          <option value="docker">Docker Hub</option>
          <option value="harbor">Harbor</option>
          <option value="gcr">GCR</option>
          <option value="ecr">AWS ECR</option>
          <option value="acr">阿里云 ACR</option>
          <option value="quay">Quay.io</option>
        </select>
      </div>
      <button class="btn btn-check-all" @click="checkAllConnections" :disabled="checkingAll">
        {{ checkingAll ? '检测中...' : '🔌 检测全部连接' }}
      </button>
    </div>

    <!-- 数据表格 -->
    <div class="table-card">
      <div class="table-loading" v-if="loading">
        <div class="loading-spinner"></div>
        <span>加载中...</span>
      </div>
      <table class="data-table" v-else>
        <thead>
          <tr>
            <th class="col-name">仓库名称</th>
            <th class="col-type">类型</th>
            <th class="col-url">仓库地址</th>
            <th class="col-status">状态</th>
            <th class="col-default">默认</th>
            <th class="col-time">最后检测</th>
            <th class="col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="registries.length === 0">
            <td colspan="7" class="empty-cell">
              <div class="empty-state">
                <span class="empty-icon">📭</span>
                <span class="empty-text">暂无镜像仓库，点击上方"添加仓库"按钮创建</span>
              </div>
            </td>
          </tr>
          <tr v-for="registry in registries" :key="registry.id" :class="{ 'is-default': registry.is_default }">
            <td class="col-name">
              <div class="name-cell">
                <span class="registry-icon">{{ getTypeIcon(registry.type) }}</span>
                <div class="name-info">
                  <span class="registry-name">{{ registry.name }}</span>
                  <span class="registry-desc" v-if="registry.description">{{ registry.description }}</span>
                </div>
                <span class="default-badge" v-if="registry.is_default">默认</span>
              </div>
            </td>
            <td class="col-type">
              <span class="type-tag" :class="`type-${registry.type}`">
                {{ getTypeName(registry.type) }}
              </span>
            </td>
            <td class="col-url">
              <div class="url-cell">
                <span class="url-text">{{ registry.url }}</span>
                <button class="copy-btn" @click="copyUrl(registry.url)" title="复制地址">📋</button>
              </div>
            </td>
            <td class="col-status">
              <div class="status-cell" :class="`status-${registry.status}`">
                <span class="status-dot"></span>
                <span class="status-text">{{ getStatusText(registry.status) }}</span>
              </div>
            </td>
            <td class="col-default">
              <button 
                class="default-btn" 
                :class="{ active: registry.is_default }"
                @click="handleSetDefault(registry)"
                :disabled="registry.is_default"
                :title="registry.is_default ? '当前默认仓库' : '设为默认'"
              >
                {{ registry.is_default ? '⭐' : '☆' }}
              </button>
            </td>
            <td class="col-time">
              <span class="time-text" v-if="registry.last_check_at">
                {{ formatTime(registry.last_check_at) }}
              </span>
              <span class="time-text empty" v-else>-</span>
            </td>
            <td class="col-actions">
              <div class="action-group">
                <button class="action-btn check" @click="checkConnection(registry)" :disabled="registry.checking" title="检测连接">
                  {{ registry.checking ? '⏳' : '🔌' }}
                </button>
                <button class="action-btn edit" @click="openEditModal(registry)" title="编辑">✏️</button>
                <button class="action-btn delete" @click="handleDelete(registry)" title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrapper" v-if="total > pageSize">
      <div class="pagination-info">
        共 {{ total }} 条记录，第 {{ currentPage }}/{{ Math.ceil(total / pageSize) }} 页
      </div>
      <div class="pagination-controls">
        <button class="page-btn" :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">上一页</button>
        <span class="page-number">{{ currentPage }}</span>
        <button class="page-btn" :disabled="currentPage >= Math.ceil(total / pageSize)" @click="changePage(currentPage + 1)">下一页</button>
      </div>
    </div>

    <!-- 添加/编辑模态框 -->
    <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
      <div class="modal-container">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑镜像仓库' : '添加镜像仓库' }}</h3>
          <button class="modal-close" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitForm">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label required">仓库名称</label>
                <input 
                  type="text" 
                  v-model="formData.name" 
                  class="form-input" 
                  placeholder="如: My Harbor"
                  required
                />
              </div>
              <div class="form-group">
                <label class="form-label required">仓库类型</label>
                <select v-model="formData.type" class="form-select" required>
                  <option value="docker">Docker Hub</option>
                  <option value="harbor">Harbor</option>
                  <option value="gcr">Google GCR</option>
                  <option value="ecr">AWS ECR</option>
                  <option value="acr">阿里云 ACR</option>
                  <option value="quay">Quay.io</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label required">仓库地址</label>
              <input 
                type="url" 
                v-model="formData.url" 
                class="form-input" 
                placeholder="如: https://registry.example.com"
                required
              />
              <span class="form-hint">完整的仓库 URL 地址</span>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">用户名</label>
                <input 
                  type="text" 
                  v-model="formData.username" 
                  class="form-input" 
                  placeholder="认证用户名（可选）"
                />
              </div>
              <div class="form-group">
                <label class="form-label">密码</label>
                <input 
                  type="password" 
                  v-model="formData.password" 
                  class="form-input" 
                  :placeholder="isEdit ? '留空保持不变' : '认证密码（可选）'"
                />
              </div>
            </div>

            <!-- 阿里云 ACR 配置 -->
            <div class="acr-config-section" v-if="formData.type === 'acr'">
              <div class="section-title">
                <span class="section-icon">🔑</span>
                阿里云 AccessKey 配置
              </div>
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label required">AccessKey ID</label>
                  <input 
                    type="text" 
                    v-model="formData.access_key_id" 
                    class="form-input" 
                    placeholder="阿里云 AccessKey ID"
                    required
                  />
                </div>
                <div class="form-group">
                  <label class="form-label required">AccessKey Secret</label>
                  <input 
                    type="password" 
                    v-model="formData.access_key_secret" 
                    class="form-input" 
                    :placeholder="isEdit ? '留空保持不变' : '阿里云 AccessKey Secret'"
                  />
                </div>
              </div>
              <div class="form-group">
                <label class="form-label">区域</label>
                <select v-model="formData.region" class="form-select">
                  <option value="">自动识别（从 URL 解析）</option>
                  <option value="cn-hangzhou">华东 1 (杭州)</option>
                  <option value="cn-shanghai">华东 2 (上海)</option>
                  <option value="cn-beijing">华北 2 (北京)</option>
                  <option value="cn-shenzhen">华南 1 (深圳)</option>
                  <option value="cn-chengdu">西南 1 (成都)</option>
                  <option value="cn-hongkong">中国香港</option>
                  <option value="ap-southeast-1">新加坡</option>
                  <option value="us-west-1">美国西部 1</option>
                </select>
                <span class="form-hint">如果 URL 包含区域信息（如 registry.cn-hangzhou.aliyuncs.com），可留空自动识别</span>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">描述</label>
              <textarea 
                v-model="formData.description" 
                class="form-textarea" 
                placeholder="仓库用途描述..."
                rows="2"
              ></textarea>
            </div>

            <div class="form-row checkbox-row">
              <label class="checkbox-label">
                <input type="checkbox" v-model="formData.insecure" />
                <span class="checkbox-text">跳过 TLS 证书验证</span>
                <span class="checkbox-hint">（不安全，仅用于测试环境）</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="formData.is_default" />
                <span class="checkbox-text">设为默认仓库</span>
              </label>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-cancel" @click="closeModal">取消</button>
              <button type="submit" class="btn btn-submit" :disabled="submitting">
                {{ submitting ? '提交中...' : (isEdit ? '保存' : '创建') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div class="modal-overlay" v-if="showDeleteConfirm" @click.self="showDeleteConfirm = false">
      <div class="modal-container confirm-modal">
        <div class="modal-header danger">
          <h3>⚠️ 确认删除</h3>
        </div>
        <div class="modal-body">
          <p class="confirm-text">确定要删除仓库 <strong>{{ deleteTarget?.name }}</strong> 吗？</p>
          <p class="confirm-warning">此操作不可恢复，请谨慎操作。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-cancel" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="confirmDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import {
  getRegistryList,
  getRegistryStats,
  createRegistry,
  updateRegistry,
  deleteRegistry,
  checkRegistryConnection,
  setDefaultRegistry
} from '@/api/image.js'

export default {
  name: 'ImageRepositories',
  setup() {
    // 数据状态
    const registries = ref([])
    const stats = ref({ total: 0, connected: 0, disconnected: 0, type_counts: {} })
    const loading = ref(false)
    const checkingAll = ref(false)
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    
    // 搜索过滤
    const searchKeyword = ref('')
    const filterType = ref('')
    
    // 模态框状态
    const showModal = ref(false)
    const isEdit = ref(false)
    const submitting = ref(false)
    const formData = reactive({
      id: null,
      name: '',
      type: 'docker',
      url: '',
      username: '',
      password: '',
      access_key_id: '',
      access_key_secret: '',
      region: '',
      insecure: false,
      description: '',
      is_default: false
    })
    
    // 删除确认
    const showDeleteConfirm = ref(false)
    const deleteTarget = ref(null)
    const deleting = ref(false)
    
    // 类型配置
    const typeConfig = {
      docker: { name: 'Docker Hub', icon: '🐳' },
      harbor: { name: 'Harbor', icon: '⚓' },
      gcr: { name: 'GCR', icon: '☁️' },
      ecr: { name: 'AWS ECR', icon: '🟠' },
      acr: { name: '阿里云 ACR', icon: '🔶' },
      quay: { name: 'Quay.io', icon: '🔴' }
    }
    
    // 加载数据
    const loadData = async () => {
      loading.value = true
      try {
        const [listRes, statsRes] = await Promise.all([
          getRegistryList({
            keyword: searchKeyword.value,
            type: filterType.value,
            page: currentPage.value,
            page_size: pageSize.value
          }),
          getRegistryStats()
        ])
        
        if (listRes.code === 0) {
          registries.value = listRes.data?.list || []
          total.value = listRes.data?.total || 0
        }
        if (statsRes.code === 0) {
          stats.value = statsRes.data || { total: 0, connected: 0, disconnected: 0, type_counts: {} }
        }
      } catch (error) {
        console.error('加载数据失败:', error)
      } finally {
        loading.value = false
      }
    }
    
    // 搜索处理
    let searchTimer = null
    const handleSearch = () => {
      clearTimeout(searchTimer)
      searchTimer = setTimeout(() => {
        currentPage.value = 1
        loadData()
      }, 300)
    }
    
    const clearSearch = () => {
      searchKeyword.value = ''
      currentPage.value = 1
      loadData()
    }
    
    const handleFilter = () => {
      currentPage.value = 1
      loadData()
    }
    
    // 分页
    const changePage = (page) => {
      currentPage.value = page
      loadData()
    }
    
    // 模态框操作
    const openCreateModal = () => {
      isEdit.value = false
      Object.assign(formData, {
        id: null,
        name: '',
        type: 'docker',
        url: '',
        username: '',
        password: '',
        access_key_id: '',
        access_key_secret: '',
        region: '',
        insecure: false,
        description: '',
        is_default: false
      })
      showModal.value = true
    }
    
    const openEditModal = (registry) => {
      isEdit.value = true
      Object.assign(formData, {
        id: registry.id,
        name: registry.name,
        type: registry.type,
        url: registry.url,
        username: registry.username || '',
        password: '',
        access_key_id: registry.access_key_id || '',
        access_key_secret: '',
        region: registry.region || '',
        insecure: registry.insecure || false,
        description: registry.description || '',
        is_default: registry.is_default || false
      })
      showModal.value = true
    }
    
    const closeModal = () => {
      showModal.value = false
    }
    
    // 提交表单
    const submitForm = async () => {
      submitting.value = true
      try {
        const data = { ...formData }
        const res = isEdit.value 
          ? await updateRegistry(data)
          : await createRegistry(data)
        
        if (res.code === 0) {
          alert(isEdit.value ? '更新成功' : '创建成功')
          closeModal()
          loadData()
        } else {
          alert(res.msg || '操作失败')
        }
      } catch (error) {
        alert('操作失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }
    
    // 删除操作
    const handleDelete = (registry) => {
      deleteTarget.value = registry
      showDeleteConfirm.value = true
    }
    
    const confirmDelete = async () => {
      if (!deleteTarget.value) return
      deleting.value = true
      try {
        const res = await deleteRegistry(deleteTarget.value.id)
        if (res.code === 0) {
          alert('删除成功')
          showDeleteConfirm.value = false
          loadData()
        } else {
          alert(res.msg || '删除失败')
        }
      } catch (error) {
        alert('删除失败: ' + error.message)
      } finally {
        deleting.value = false
      }
    }
    
    // 检测连接
    const checkConnection = async (registry) => {
      registry.checking = true
      try {
        const res = await checkRegistryConnection(registry.id)
        if (res.code === 0 && res.data) {
          registry.status = res.data.status
          registry.last_check_at = res.data.last_check_at
          registry.last_error = res.data.last_error
        }
      } catch (error) {
        console.error('检测失败:', error)
      } finally {
        registry.checking = false
      }
    }
    
    const checkAllConnections = async () => {
      checkingAll.value = true
      const promises = registries.value.map(r => checkConnection(r))
      await Promise.all(promises)
      checkingAll.value = false
      loadData()
    }
    
    // 设置默认
    const handleSetDefault = async (registry) => {
      if (registry.is_default) return
      try {
        const res = await setDefaultRegistry(registry.id)
        if (res.code === 0) {
          loadData()
        } else {
          alert(res.msg || '设置失败')
        }
      } catch (error) {
        alert('设置失败: ' + error.message)
      }
    }
    
    // 工具函数
    const getTypeIcon = (type) => typeConfig[type]?.icon || '📦'
    const getTypeName = (type) => typeConfig[type]?.name || type
    
    const getStatusText = (status) => {
      const map = { connected: '已连接', disconnected: '连接失败', unknown: '未检测' }
      return map[status] || status
    }
    
    const formatTime = (timestamp) => {
      if (!timestamp) return '-'
      const date = new Date(timestamp * 1000)
      return date.toLocaleString('zh-CN')
    }
    
    const copyUrl = (url) => {
      navigator.clipboard.writeText(url).then(() => {
        alert('已复制到剪贴板')
      }).catch(() => {
        alert('复制失败')
      })
    }
    
    onMounted(() => {
      loadData()
    })
    
    return {
      registries, stats, loading, checkingAll, total, currentPage, pageSize,
      searchKeyword, filterType,
      showModal, isEdit, submitting, formData,
      showDeleteConfirm, deleteTarget, deleting,
      loadData, handleSearch, clearSearch, handleFilter, changePage,
      openCreateModal, openEditModal, closeModal, submitForm,
      handleDelete, confirmDelete,
      checkConnection, checkAllConnections, handleSetDefault,
      getTypeIcon, getTypeName, getStatusText, formatTime, copyUrl
    }
  }
}
</script>

<style scoped>
/* 页面容器 */
.image-registry-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #1a202c;
}

.title-icon {
  font-size: 28px;
}

.page-desc {
  margin: 0;
  color: #718096;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.stat-icon {
  font-size: 32px;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.stat-card.total .stat-icon { background: #ebf5ff; }
.stat-card.connected .stat-icon { background: #e6fffa; }
.stat-card.disconnected .stat-icon { background: #fff5f5; }
.stat-card.types .stat-icon { background: #faf5ff; }

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1a202c;
}

.stat-label {
  font-size: 13px;
  color: #718096;
  margin-top: 4px;
}

/* 工具栏 */
.toolbar-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  margin-bottom: 24px;
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 400px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 14px;
}

.search-input {
  width: 100%;
  padding: 10px 36px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66,153,225,0.15);
}

.clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border: none;
  background: #e2e8f0;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  color: #718096;
}

.filter-select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background: white;
  cursor: pointer;
  min-width: 140px;
}

/* 按钮样式 */
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

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #3182ce, #2b6cb0);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66,153,225,0.35);
}

.btn-refresh {
  background: #edf2f7;
  color: #4a5568;
}

.btn-refresh:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-check-all {
  background: #805ad5;
  color: white;
}

.btn-check-all:hover:not(:disabled) {
  background: #6b46c1;
}

/* 数据表格 */
.table-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.table-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px;
  color: #718096;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid #e2e8f0;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  background: #f7fafc;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 16px;
  border-bottom: 1px solid #edf2f7;
  vertical-align: middle;
}

.data-table tr:hover {
  background: #f7fafc;
}

.data-table tr.is-default {
  background: #fffbeb;
}

/* 表格单元格 */
.name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.registry-icon {
  font-size: 24px;
}

.name-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.registry-name {
  font-weight: 600;
  color: #1a202c;
}

.registry-desc {
  font-size: 12px;
  color: #718096;
}

.default-badge {
  background: #fbbf24;
  color: white;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.type-tag {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.type-docker { background: #e6fffa; color: #047857; }
.type-harbor { background: #fef3c7; color: #92400e; }
.type-gcr { background: #dbeafe; color: #1e40af; }
.type-ecr { background: #fed7aa; color: #9a3412; }
.type-acr { background: #fde68a; color: #92400e; }
.type-quay { background: #fecaca; color: #991b1b; }

.url-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.url-text {
  font-size: 13px;
  color: #4a5568;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.copy-btn:hover {
  opacity: 1;
}

.status-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-connected .status-dot { background: #10b981; }
.status-disconnected .status-dot { background: #ef4444; }
.status-unknown .status-dot { background: #9ca3af; }

.status-text {
  font-size: 13px;
}

.status-connected .status-text { color: #047857; }
.status-disconnected .status-text { color: #dc2626; }
.status-unknown .status-text { color: #6b7280; }

.default-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: none;
  font-size: 18px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.default-btn:hover:not(:disabled) {
  background: #fef3c7;
}

.default-btn.active {
  color: #f59e0b;
  cursor: default;
}

.time-text {
  font-size: 13px;
  color: #718096;
}

.time-text.empty {
  color: #cbd5e0;
}

.action-group {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.action-btn:hover:not(:disabled) {
  background: #edf2f7;
}

.action-btn.delete:hover:not(:disabled) {
  background: #fed7d7;
}

.empty-cell {
  text-align: center;
  padding: 60px !important;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-icon {
  font-size: 48px;
}

.empty-text {
  color: #718096;
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  margin-top: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.pagination-info {
  color: #718096;
  font-size: 14px;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-number {
  padding: 8px 14px;
  background: #4299e1;
  color: white;
  border-radius: 6px;
  font-weight: 500;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-container {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
  animation: modalIn 0.2s ease;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.confirm-modal {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.modal-header.danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255,255,255,0.2);
  color: white;
  border-radius: 8px;
  font-size: 20px;
  cursor: pointer;
  transition: background 0.2s;
}

.modal-close:hover {
  background: rgba(255,255,255,0.3);
}

.modal-body {
  padding: 24px;
  max-height: 60vh;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
  margin-top: 20px;
}

/* 表单样式 */
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
}

.form-label.required::after {
  content: ' *';
  color: #e53e3e;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66,153,225,0.15);
}

.form-hint {
  font-size: 12px;
  color: #718096;
  margin-top: 4px;
}

.checkbox-row {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-text {
  font-size: 14px;
  color: #4a5568;
}

.checkbox-hint {
  font-size: 12px;
  color: #a0aec0;
}

/* 阿里云 ACR 配置区域 */
.acr-config-section {
  margin: 20px 0;
  padding: 20px;
  background: linear-gradient(135deg, #fef9e7 0%, #fff8e1 100%);
  border: 1px solid #fcd34d;
  border-radius: 12px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 16px;
}

.section-icon {
  font-size: 18px;
}

.acr-config-section .form-group {
  margin-bottom: 12px;
}

.acr-config-section .form-row {
  margin-bottom: 12px;
}

.btn-cancel {
  background: #edf2f7;
  color: #4a5568;
}

.btn-cancel:hover {
  background: #e2e8f0;
}

.btn-submit {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.btn-submit:hover:not(:disabled) {
  background: linear-gradient(135deg, #3182ce, #2b6cb0);
}

.btn-danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: linear-gradient(135deg, #e53e3e, #c53030);
}

.confirm-text {
  font-size: 15px;
  color: #4a5568;
  margin-bottom: 8px;
}

.confirm-warning {
  font-size: 13px;
  color: #e53e3e;
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .toolbar-card {
    flex-wrap: wrap;
  }
  
  .search-box {
    width: 100%;
    max-width: none;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
}
</style>

<template>
  <div class="image-browse-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <span class="title-icon">🖼️</span>
          镜像浏览
        </h1>
        <p class="page-desc">浏览仓库中的镜像项目、版本标签，支持镜像删除</p>
      </div>
    </div>

    <!-- 仓库选择器 -->
    <div class="selector-card">
      <div class="selector-group">
        <label class="selector-label">选择仓库：</label>
        <select v-model="selectedRegistryId" class="selector-input" @change="loadRepositories">
          <option value="">请选择镜像仓库</option>
          <option v-for="reg in registries" :key="reg.id" :value="reg.id">
            {{ getTypeIcon(reg.type) }} {{ reg.name }} ({{ reg.url }})
          </option>
        </select>
        <button class="btn btn-refresh" @click="loadRegistries" title="刷新仓库列表">🔄</button>
      </div>
      <div class="registry-status" v-if="selectedRegistry">
        <span class="status-dot" :class="selectedRegistry.status"></span>
        <span class="status-text">{{ selectedRegistry.status === 'connected' ? '已连接' : '未连接' }}</span>
      </div>
    </div>

    <div class="content-area" v-if="selectedRegistryId">
      <!-- 左侧：镜像项目列表 -->
      <div class="sidebar">
        <div class="sidebar-header">
          <h3>镜像项目</h3>
          <span class="count-badge">{{ repositories.length }}</span>
        </div>
        <div class="search-box">
          <input 
            v-model="repoSearch" 
            type="text" 
            placeholder="搜索镜像..."
            class="search-input"
          />
        </div>
        <div class="repo-list" v-if="!loadingRepos">
          <div 
            v-for="repo in filteredRepositories" 
            :key="repo.full_name"
            class="repo-item"
            :class="{ active: selectedRepo?.full_name === repo.full_name }"
            @click="selectRepository(repo)"
          >
            <span class="repo-icon">📦</span>
            <div class="repo-info">
              <span class="repo-name">{{ repo.name }}</span>
              <span class="repo-tags" v-if="repo.tag_count">{{ repo.tag_count }} 个标签</span>
            </div>
          </div>
          <div v-if="filteredRepositories.length === 0" class="empty-hint">
            暂无镜像项目
          </div>
        </div>
        <div class="loading-hint" v-else>
          <div class="spinner"></div>
          加载中...
        </div>
      </div>

      <!-- 右侧：标签列表 -->
      <div class="main-content">
        <div v-if="selectedRepo">
          <!-- 镜像信息 -->
          <div class="repo-detail-header">
            <div class="repo-detail-info">
              <h2>{{ selectedRepo.full_name }}</h2>
              <p class="repo-desc" v-if="selectedRepo.description">{{ selectedRepo.description }}</p>
              <div class="repo-meta">
                <span v-if="selectedRepo.pull_count">
                  <strong>{{ selectedRepo.pull_count }}</strong> 次拉取
                </span>
              </div>
            </div>
            <button class="btn btn-refresh" @click="loadTags" :disabled="loadingTags">
              {{ loadingTags ? '加载中...' : '🔄 刷新标签' }}
            </button>
          </div>

          <!-- 标签表格 -->
          <div class="tags-table-card">
            <table class="data-table">
              <thead>
                <tr>
                  <th>标签名</th>
                  <th>Digest</th>
                  <th>大小</th>
                  <th>推送时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="loadingTags">
                  <td colspan="5" class="loading-cell">
                    <div class="spinner"></div>
                    加载中...
                  </td>
                </tr>
                <tr v-else-if="tags.length === 0">
                  <td colspan="5" class="empty-cell">暂无标签</td>
                </tr>
                <tr v-for="tag in tags" :key="tag.name" :class="{ 'highlight': tag.name === 'latest' }">
                  <td>
                    <div class="tag-cell">
                      <span class="tag-name">{{ tag.name }}</span>
                      <span class="latest-badge" v-if="tag.name === 'latest'">最新</span>
                    </div>
                  </td>
                  <td>
                    <span class="digest-text" :title="tag.digest">
                      {{ tag.digest ? tag.digest.substring(0, 16) + '...' : '-' }}
                    </span>
                  </td>
                  <td>{{ formatSize(tag.size) }}</td>
                  <td>{{ formatTime(tag.pushed_at) }}</td>
                  <td>
                    <div class="action-group">
                      <button class="action-btn copy" @click="copyImageUrl(tag)" title="复制镜像地址">📋</button>
                      <button class="action-btn detail" @click="viewTagDetail(tag)" title="查看详情">🔍</button>
                      <button class="action-btn delete" @click="confirmDeleteTag(tag)" title="删除">🗑️</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 未选择镜像时的提示 -->
        <div v-else class="no-selection">
          <span class="empty-icon">👈</span>
          <p>请从左侧选择一个镜像项目</p>
        </div>
      </div>
    </div>

    <!-- 未选择仓库时的提示 -->
    <div v-else class="no-registry">
      <span class="empty-icon">☁️</span>
      <p>请先选择一个镜像仓库</p>
      <router-link to="/image-repositories" class="btn btn-primary">
        前往仓库管理
      </router-link>
    </div>

    <!-- 镜像详情模态框 -->
    <div class="modal-overlay" v-if="showDetailModal" @click.self="showDetailModal = false">
      <div class="modal-container detail-modal">
        <div class="modal-header">
          <h3>镜像详情</h3>
          <button class="modal-close" @click="showDetailModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="detail-section" v-if="tagDetail">
            <div class="detail-row">
              <span class="detail-label">镜像地址</span>
              <span class="detail-value code">{{ getFullImageUrl(selectedTag) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Digest</span>
              <span class="detail-value code">{{ tagDetail.digest }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">大小</span>
              <span class="detail-value">{{ formatSize(tagDetail.size) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">MediaType</span>
              <span class="detail-value">{{ tagDetail.media_type || '-' }}</span>
            </div>
            <div class="detail-row" v-if="tagDetail.layers && tagDetail.layers.length">
              <span class="detail-label">层数</span>
              <span class="detail-value">{{ tagDetail.layers.length }} 层</span>
            </div>
          </div>
          <div class="loading-hint" v-else>
            <div class="spinner"></div>
            加载中...
          </div>
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
          <p class="confirm-text">确定要删除镜像标签吗？</p>
          <p class="confirm-image">
            <code>{{ selectedRepo?.full_name }}:{{ deleteTarget?.name }}</code>
          </p>
          <p class="confirm-warning">此操作不可恢复，请谨慎操作！</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-cancel" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="doDeleteTag" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  getAllRegistries,
  listRepositories,
  listTags,
  getImageDetail,
  deleteImageTag
} from '@/api/image.js'

export default {
  name: 'Images',
  setup() {
    const route = useRoute()
    
    // 状态
    const registries = ref([])
    const selectedRegistryId = ref('')
    const repositories = ref([])
    const tags = ref([])
    const repoSearch = ref('')
    const selectedRepo = ref(null)
    
    // 加载状态
    const loadingRepos = ref(false)
    const loadingTags = ref(false)
    
    // 详情相关
    const showDetailModal = ref(false)
    const selectedTag = ref(null)
    const tagDetail = ref(null)
    
    // 删除相关
    const showDeleteConfirm = ref(false)
    const deleteTarget = ref(null)
    const deleting = ref(false)
    
    // 类型配置
    const typeConfig = {
      docker: '🐳',
      harbor: '⚓',
      gcr: '☁️',
      ecr: '🟠',
      acr: '🔶',
      quay: '🔴'
    }
    
    // 计算属性
    const selectedRegistry = computed(() => {
      return registries.value.find(r => r.id === selectedRegistryId.value)
    })
    
    const filteredRepositories = computed(() => {
      if (!repoSearch.value) return repositories.value
      const search = repoSearch.value.toLowerCase()
      return repositories.value.filter(r => 
        r.name.toLowerCase().includes(search) || 
        r.full_name.toLowerCase().includes(search)
      )
    })
    
    // 加载仓库列表
    const loadRegistries = async () => {
      try {
        const res = await getAllRegistries()
        if (res.code === 0) {
          registries.value = res.data?.list || []
        }
      } catch (error) {
        console.error('加载仓库失败:', error)
      }
    }
    
    // 加载镜像项目
    const loadRepositories = async () => {
      if (!selectedRegistryId.value) {
        repositories.value = []
        return
      }
      
      loadingRepos.value = true
      selectedRepo.value = null
      tags.value = []
      
      try {
        const res = await listRepositories(selectedRegistryId.value)
        if (res.code === 0) {
          repositories.value = res.data?.list || []
        } else {
          alert(res.msg || '加载镜像项目失败')
        }
      } catch (error) {
        alert('加载镜像项目失败: ' + error.message)
      } finally {
        loadingRepos.value = false
      }
    }
    
    // 选择镜像项目
    const selectRepository = (repo) => {
      selectedRepo.value = repo
      loadTags()
    }
    
    // 加载标签
    const loadTags = async () => {
      if (!selectedRepo.value) return
      
      loadingTags.value = true
      try {
        const res = await listTags(selectedRegistryId.value, selectedRepo.value.full_name)
        if (res.code === 0) {
          tags.value = res.data?.list || []
        } else {
          alert(res.msg || '加载标签失败')
        }
      } catch (error) {
        alert('加载标签失败: ' + error.message)
      } finally {
        loadingTags.value = false
      }
    }
    
    // 查看标签详情
    const viewTagDetail = async (tag) => {
      selectedTag.value = tag
      tagDetail.value = null
      showDetailModal.value = true
      
      try {
        const res = await getImageDetail(
          selectedRegistryId.value, 
          selectedRepo.value.full_name,
          tag.name
        )
        if (res.code === 0) {
          tagDetail.value = res.data
        }
      } catch (error) {
        console.error('获取详情失败:', error)
      }
    }
    
    // 复制镜像地址
    const copyImageUrl = (tag) => {
      const url = getFullImageUrl(tag)
      navigator.clipboard.writeText(url).then(() => {
        alert('已复制到剪贴板')
      }).catch(() => {
        alert('复制失败')
      })
    }
    
    // 获取完整镜像地址
    const getFullImageUrl = (tag) => {
      if (!selectedRegistry.value || !selectedRepo.value) return ''
      const url = selectedRegistry.value.url.replace(/^https?:\/\//, '')
      return `${url}/${selectedRepo.value.full_name}:${tag.name}`
    }
    
    // 确认删除
    const confirmDeleteTag = (tag) => {
      deleteTarget.value = tag
      showDeleteConfirm.value = true
    }
    
    // 执行删除
    const doDeleteTag = async () => {
      if (!deleteTarget.value) return
      
      deleting.value = true
      try {
        const res = await deleteImageTag(
          selectedRegistryId.value,
          selectedRepo.value.full_name,
          deleteTarget.value.name
        )
        if (res.code === 0) {
          alert('删除成功')
          showDeleteConfirm.value = false
          loadTags()
        } else {
          alert(res.msg || '删除失败')
        }
      } catch (error) {
        alert('删除失败: ' + error.message)
      } finally {
        deleting.value = false
      }
    }
    
    // 工具函数
    const getTypeIcon = (type) => typeConfig[type] || '📦'
    
    const formatSize = (bytes) => {
      if (!bytes) return '-'
      const units = ['B', 'KB', 'MB', 'GB']
      let i = 0
      while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024
        i++
      }
      return bytes.toFixed(1) + ' ' + units[i]
    }
    
    const formatTime = (timestamp) => {
      if (!timestamp) return '-'
      return new Date(timestamp * 1000).toLocaleString('zh-CN')
    }
    
    // 监听路由参数
    watch(() => route.params.repoId, (newId) => {
      if (newId && registries.value.length > 0) {
        selectedRegistryId.value = parseInt(newId)
        loadRepositories()
      }
    })
    
    onMounted(() => {
      loadRegistries().then(() => {
        if (route.params.repoId) {
          selectedRegistryId.value = parseInt(route.params.repoId)
          loadRepositories()
        }
      })
    })
    
    return {
      registries, selectedRegistryId, selectedRegistry,
      repositories, tags, repoSearch, selectedRepo,
      filteredRepositories,
      loadingRepos, loadingTags,
      showDetailModal, selectedTag, tagDetail,
      showDeleteConfirm, deleteTarget, deleting,
      loadRegistries, loadRepositories, selectRepository, loadTags,
      viewTagDetail, copyImageUrl, getFullImageUrl,
      confirmDeleteTag, doDeleteTag,
      getTypeIcon, formatSize, formatTime
    }
  }
}
</script>

<style scoped>
.image-browse-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.page-header {
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

.title-icon { font-size: 28px; }

.page-desc {
  margin: 0;
  color: #718096;
  font-size: 14px;
}

/* 选择器卡片 */
.selector-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  margin-bottom: 24px;
}

.selector-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.selector-label {
  font-weight: 500;
  color: #4a5568;
}

.selector-input {
  min-width: 400px;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
}

.registry-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.status-dot.connected { background: #10b981; }
.status-dot.disconnected { background: #ef4444; }
.status-dot.unknown { background: #9ca3af; }

/* 内容区域 */
.content-area {
  display: flex;
  gap: 24px;
}

/* 左侧边栏 */
.sidebar {
  width: 300px;
  flex-shrink: 0;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.count-badge {
  background: #4299e1;
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
}

.sidebar .search-box {
  padding: 12px;
  border-bottom: 1px solid #e2e8f0;
}

.sidebar .search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 13px;
}

.repo-list {
  max-height: calc(100vh - 400px);
  overflow-y: auto;
}

.repo-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f7fafc;
  transition: all 0.2s;
}

.repo-item:hover {
  background: #f7fafc;
}

.repo-item.active {
  background: #ebf5ff;
  border-left: 3px solid #4299e1;
}

.repo-icon {
  font-size: 20px;
}

.repo-info {
  flex: 1;
  min-width: 0;
}

.repo-name {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #2d3748;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.repo-tags {
  display: block;
  font-size: 12px;
  color: #718096;
  margin-top: 2px;
}

/* 主内容区 */
.main-content {
  flex: 1;
  min-width: 0;
}

.repo-detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  margin-bottom: 20px;
}

.repo-detail-header h2 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #1a202c;
}

.repo-desc {
  color: #718096;
  font-size: 14px;
  margin: 0 0 8px 0;
}

.repo-meta {
  font-size: 13px;
  color: #718096;
}

/* 标签表格 */
.tags-table-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 14px 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  background: #f7fafc;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #edf2f7;
}

.data-table tr:hover {
  background: #f7fafc;
}

.data-table tr.highlight {
  background: #fffbeb;
}

.tag-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-name {
  font-weight: 500;
  color: #2d3748;
}

.latest-badge {
  background: #fbbf24;
  color: white;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 8px;
  font-weight: 500;
}

.digest-text {
  font-family: monospace;
  font-size: 12px;
  color: #718096;
}

.action-group {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.action-btn:hover { background: #edf2f7; }
.action-btn.delete:hover { background: #fed7d7; }

/* 空状态 */
.no-selection,
.no-registry {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.no-selection p,
.no-registry p {
  color: #718096;
  font-size: 16px;
  margin: 0 0 16px 0;
}

.empty-hint, .loading-hint {
  padding: 40px;
  text-align: center;
  color: #718096;
}

.loading-cell, .empty-cell {
  text-align: center;
  padding: 40px !important;
  color: #718096;
}

.loading-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

/* 加载动画 */
.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #e2e8f0;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to { transform: rotate(360deg); }
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
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66,153,225,0.35);
}

.btn-refresh {
  background: #edf2f7;
  color: #4a5568;
  padding: 8px 12px;
}

.btn-refresh:hover { background: #e2e8f0; }

.btn-cancel {
  background: #edf2f7;
  color: #4a5568;
}

.btn-danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
  color: white;
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
}

.modal-container {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.detail-modal {
  max-width: 600px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
  border-radius: 12px 12px 0 0;
}

.modal-header.danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
}

.modal-close {
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(255,255,255,0.2);
  color: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 18px;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e2e8f0;
}

/* 详情区域 */
.detail-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: #718096;
  font-weight: 500;
}

.detail-value {
  font-size: 14px;
  color: #2d3748;
}

.detail-value.code {
  font-family: monospace;
  background: #f7fafc;
  padding: 8px;
  border-radius: 6px;
  word-break: break-all;
}

/* 确认模态框 */
.confirm-text {
  font-size: 15px;
  color: #4a5568;
  margin: 0 0 12px 0;
}

.confirm-image {
  margin: 0 0 12px 0;
}

.confirm-image code {
  background: #f7fafc;
  padding: 8px 12px;
  border-radius: 6px;
  display: block;
  font-family: monospace;
}

.confirm-warning {
  font-size: 13px;
  color: #e53e3e;
  margin: 0;
}

/* 响应式 */
@media (max-width: 1024px) {
  .content-area {
    flex-direction: column;
  }
  
  .sidebar {
    width: 100%;
  }
  
  .repo-list {
    max-height: 300px;
  }
}
</style>

<template>
  <div class="agents-page">
    <!-- ====== 页面头部 - 渐变标题区 ====== -->
    <div class="page-hero">
      <div class="hero-content">
        <div class="hero-left">
          <div class="hero-breadcrumb">
            <span class="breadcrumb-item">CI/CD</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">构建探针</span>
          </div>
          <h1 class="hero-title">
            <div class="hero-icon-wrapper">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                <path d="M2 17l10 5 10-5"/>
                <path d="M2 12l10 5 10-5"/>
              </svg>
            </div>
            构建探针管理
          </h1>
          <p class="hero-desc">统一管理可观测性探针、诊断工具与安全扫描 Agent，自动注入 CI/CD 构建流程</p>
        </div>
        <div class="hero-right">
          <button class="hero-btn primary" @click="showUploadModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            上传探针
          </button>
          <button class="hero-btn secondary" @click="loadAgents" :disabled="loading">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            {{ loading ? '刷新中...' : '刷新' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ====== 数据概览卡片 ====== -->
    <div class="overview-section">
      <div class="overview-cards">
        <div class="overview-card">
          <div class="oc-icon total"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg></div>
          <div class="oc-body"><div class="oc-value">{{ agents.length }}</div><div class="oc-label">探针总数</div></div>
        </div>
        <div class="overview-card" v-for="s in categoryStats" :key="s.key">
          <div class="oc-icon" :class="s.color"><span class="cat-emoji">{{ s.icon }}</span></div>
          <div class="oc-body"><div class="oc-value">{{ s.count }}</div><div class="oc-label">{{ s.label }}</div></div>
        </div>
        <div class="overview-card">
          <div class="oc-icon active"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></div>
          <div class="oc-body"><div class="oc-value">{{ activeCount }}</div><div class="oc-label">已启用</div></div>
        </div>
      </div>
    </div>

    <!-- ====== 筛选工具栏 ====== -->
    <div class="main-section">
      <div class="toolbar">
        <div class="toolbar-left">
          <div class="search-box">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="keyword" type="text" placeholder="搜索探针名称、版本号..." @input="debounceSearch" />
            <button v-if="keyword" class="search-clear" @click="keyword = ''; loadAgents()">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
        </div>
        <div class="toolbar-right">
          <div class="filter-chips">
            <select v-model="filters.category" class="chip-select" @change="loadAgents">
              <option value="">全部分类</option>
              <option value="observability">可观测性</option>
              <option value="diagnostics">诊断工具</option>
              <option value="security">安全扫描</option>
              <option value="custom">自定义</option>
            </select>
            <select v-model="filters.scope" class="chip-select" @change="loadAgents">
              <option value="">全部语言</option>
              <option value="java">Java</option>
              <option value="go">Go</option>
              <option value="python">Python</option>
              <option value="all">通用</option>
            </select>
            <select v-model="filters.status" class="chip-select" @change="loadAgents">
              <option value="">全部状态</option>
              <option value="active">已启用</option>
              <option value="inactive">已停用</option>
            </select>
          </div>
        </div>
      </div>

      <!-- 加载中 -->
      <div v-if="loading && agents.length === 0" class="state-container">
        <div class="loader"><div class="loader-ring"></div><div class="loader-ring"></div><div class="loader-ring"></div></div>
        <p class="state-text">正在加载探针数据...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="agents.length === 0" class="state-container empty">
        <div class="empty-illustration">
          <svg viewBox="0 0 200 160" fill="none">
            <rect x="40" y="30" width="120" height="100" rx="8" fill="#f3f4f6" stroke="#e5e7eb" stroke-width="2"/>
            <rect x="55" y="50" width="90" height="8" rx="4" fill="#e5e7eb"/>
            <rect x="55" y="66" width="60" height="8" rx="4" fill="#e5e7eb"/>
            <circle cx="100" cy="115" r="12" fill="#dbeafe" stroke="#93c5fd" stroke-width="2"/>
            <path d="M95 115l3 3 7-7" stroke="#3b82f6" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h3 class="empty-title">暂无构建探针</h3>
        <p class="empty-desc">上传 Agent JAR/二进制文件，流水线构建时将自动注入到 Docker 镜像中</p>
        <button class="empty-action" @click="showUploadModal = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          上传第一个探针
        </button>
      </div>

      <!-- ====== 卡片列表 ====== -->
      <div v-else class="agent-grid">
        <div class="agent-card" v-for="a in agents" :key="a.id" :class="{ inactive: a.status === 'inactive' }">
          <div class="card-header">
            <div class="card-category" :class="a.category">
              {{ getCategoryLabel(a.category) }}
            </div>
            <div class="card-status" :class="a.status">
              <span class="status-dot"></span>
              {{ a.status === 'active' ? '已启用' : '已停用' }}
            </div>
          </div>

          <div class="card-body">
            <div class="card-icon-row">
              <div class="card-agent-icon" :class="a.category">
                {{ getCategoryEmoji(a.category) }}
              </div>
              <div class="card-info">
                <h4 class="card-name" :title="a.display_name || a.name">{{ a.display_name || a.name }}</h4>
                <div class="card-meta-line">
                  <span class="meta-tag scope" :class="a.scope">{{ getScopeLabel(a.scope) }}</span>
                  <span class="meta-tag version" v-if="a.version">v{{ a.version }}</span>
                </div>
              </div>
            </div>
            <p class="card-desc" v-if="a.description">{{ a.description }}</p>

            <div class="card-stats">
              <div class="stat-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <span>{{ a.download_count || 0 }} 次下载</span>
              </div>
              <div class="stat-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
                <span>{{ a.used_count || 0 }} 次引用</span>
              </div>
              <div class="stat-item" v-if="a.file_size">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                <span>{{ formatSize(a.file_size) }}</span>
              </div>
            </div>
          </div>

          <div class="card-footer">
            <button class="card-action-btn" :class="{ warn: a.status === 'active' }" @click.stop="handleToggle(a)" :title="a.status === 'active' ? '停用' : '启用'">
              <svg v-if="a.status === 'active'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            </button>
            <button class="card-action-btn" @click.stop="handleDownload(a)" title="下载">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            </button>
            <button class="card-action-btn" @click.stop="openEditModal(a)" title="编辑">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
            </button>
            <button class="card-action-btn danger" @click.stop="handleDelete(a)" title="删除">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ====== 上传弹窗 ====== -->
    <Transition name="modal">
      <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
        <div class="modal-container upload-modal">
          <div class="modal-header">
            <h2>上传构建探针</h2>
            <button class="modal-close" @click="showUploadModal = false">&times;</button>
          </div>
          <div class="modal-body">
            <!-- 拖拽上传区 -->
            <div class="upload-zone" :class="{ dragover: isDragover, hasFile: uploadForm.file }" @dragover.prevent="isDragover = true" @dragleave="isDragover = false" @drop.prevent="handleDrop">
              <input ref="fileInput" type="file" class="upload-input" @change="handleFileChange" accept=".jar,.so,.bin,.exe,.tar.gz,.zip,.whl,.py,.sh" />
              <div v-if="!uploadForm.file" class="upload-placeholder">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                <p>拖拽文件到此处，或 <span class="upload-link" @click="$refs.fileInput.click()">点击浏览</span></p>
                <p class="upload-hint">支持 .jar, .so, .bin, .tar.gz, .zip, .whl 等格式</p>
              </div>
              <div v-else class="upload-file-info">
                <div class="file-icon">📦</div>
                <div class="file-detail">
                  <span class="file-name">{{ uploadForm.file.name }}</span>
                  <span class="file-size">{{ formatSize(uploadForm.file.size) }}</span>
                </div>
                <button class="file-remove" @click="uploadForm.file = null">&times;</button>
              </div>
            </div>

            <div class="form-grid">
              <div class="form-group">
                <label>探针名称 <span class="required">*</span></label>
                <input v-model="uploadForm.name" type="text" placeholder="如 opentelemetry-javaagent" />
              </div>
              <div class="form-group">
                <label>显示名称 <span class="required">*</span></label>
                <input v-model="uploadForm.display_name" type="text" placeholder="如 OpenTelemetry Java Agent" />
              </div>
              <div class="form-group">
                <label>分类 <span class="required">*</span></label>
                <select v-model="uploadForm.category">
                  <option value="observability">可观测性</option>
                  <option value="diagnostics">诊断工具</option>
                  <option value="security">安全扫描</option>
                  <option value="custom">自定义</option>
                </select>
              </div>
              <div class="form-group">
                <label>适用语言 <span class="required">*</span></label>
                <select v-model="uploadForm.scope">
                  <option value="java">Java</option>
                  <option value="go">Go</option>
                  <option value="python">Python</option>
                  <option value="all">通用</option>
                </select>
              </div>
              <div class="form-group">
                <label>版本号</label>
                <input v-model="uploadForm.version" type="text" placeholder="如 1.32.0" />
              </div>
              <div class="form-group">
                <label>Docker 目标路径</label>
                <input v-model="uploadForm.docker_copy_dest" type="text" placeholder="如 /app/opentelemetry-javaagent.jar" />
              </div>
              <div class="form-group full-width">
                <label>环境变量 Key</label>
                <input v-model="uploadForm.env_key" type="text" placeholder="如 JAVA_TOOL_OPTIONS" />
              </div>
              <div class="form-group full-width">
                <label>环境变量 Value</label>
                <input v-model="uploadForm.env_value" type="text" placeholder="如 -javaagent:/app/opentelemetry-javaagent.jar" />
              </div>
              <div class="form-group full-width">
                <label>描述</label>
                <textarea v-model="uploadForm.description" rows="2" placeholder="探针用途说明..."></textarea>
              </div>
            </div>
          </div>

          <!-- 上传进度 -->
          <div v-if="uploading" class="upload-progress-bar-wrap">
            <div class="upload-progress-fill" :style="{ width: uploadProgress + '%' }"></div>
            <span class="upload-progress-text">{{ uploadProgress }}%</span>
          </div>

          <div class="modal-footer">
            <button class="modal-btn cancel" @click="showUploadModal = false">取消</button>
            <button class="modal-btn confirm" @click="handleUpload" :disabled="uploading || !uploadForm.file || !uploadForm.name">
              {{ uploading ? '上传中...' : '确认上传' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ====== 编辑弹窗 ====== -->
    <Transition name="modal">
      <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
        <div class="modal-container edit-modal">
          <div class="modal-header">
            <h2>编辑探针信息</h2>
            <button class="modal-close" @click="showEditModal = false">&times;</button>
          </div>
          <div class="modal-body">
            <div class="form-grid">
              <div class="form-group">
                <label>显示名称</label>
                <input v-model="editForm.display_name" type="text" />
              </div>
              <div class="form-group">
                <label>版本号</label>
                <input v-model="editForm.version" type="text" />
              </div>
              <div class="form-group">
                <label>分类</label>
                <select v-model="editForm.category">
                  <option value="observability">可观测性</option>
                  <option value="diagnostics">诊断工具</option>
                  <option value="security">安全扫描</option>
                  <option value="custom">自定义</option>
                </select>
              </div>
              <div class="form-group">
                <label>适用语言</label>
                <select v-model="editForm.scope">
                  <option value="java">Java</option>
                  <option value="go">Go</option>
                  <option value="python">Python</option>
                  <option value="all">通用</option>
                </select>
              </div>
              <div class="form-group">
                <label>Docker 目标路径</label>
                <input v-model="editForm.docker_copy_dest" type="text" />
              </div>
              <div class="form-group">
                <label>环境变量 Key</label>
                <input v-model="editForm.env_key" type="text" />
              </div>
              <div class="form-group full-width">
                <label>环境变量 Value</label>
                <input v-model="editForm.env_value" type="text" />
              </div>
              <div class="form-group full-width">
                <label>描述</label>
                <textarea v-model="editForm.description" rows="2"></textarea>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="modal-btn cancel" @click="showEditModal = false">取消</button>
            <button class="modal-btn confirm" @click="handleUpdate" :disabled="updating">
              {{ updating ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ====== 删除确认弹窗 ====== -->
    <Transition name="modal">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
        <div class="modal-container delete-modal">
          <div class="modal-header danger">
            <h2>确认删除</h2>
            <button class="modal-close" @click="showDeleteConfirm = false">&times;</button>
          </div>
          <div class="modal-body">
            <p class="delete-warning">确定要删除探针 <strong>{{ deleteTarget?.display_name || deleteTarget?.name }}</strong> 吗？</p>
            <p class="delete-hint">此操作将同时删除已上传的探针文件，且不可恢复。</p>
          </div>
          <div class="modal-footer">
            <button class="modal-btn cancel" @click="showDeleteConfirm = false">取消</button>
            <button class="modal-btn danger" @click="confirmDelete" :disabled="deleting">
              {{ deleting ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  getBuildAgents, uploadBuildAgent, updateBuildAgent,
  toggleBuildAgent, deleteBuildAgent, downloadBuildAgent
} from '@/api/cicd'

const agents = ref([])
const loading = ref(false)
const keyword = ref('')
const filters = ref({ category: '', scope: '', status: '' })

// 上传相关
const showUploadModal = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const isDragover = ref(false)
const fileInput = ref(null)
const uploadForm = ref({
  file: null, name: '', display_name: '', category: 'observability',
  scope: 'java', version: '', description: '', docker_copy_dest: '',
  env_key: '', env_value: ''
})

// 编辑相关
const showEditModal = ref(false)
const updating = ref(false)
const editForm = ref({})

// 删除相关
const showDeleteConfirm = ref(false)
const deleting = ref(false)
const deleteTarget = ref(null)

// 统计
const categoryStats = computed(() => {
  const map = {
    observability: { key: 'observability', label: '可观测性', icon: '📡', color: 'obs', count: 0 },
    diagnostics: { key: 'diagnostics', label: '诊断工具', icon: '🔬', color: 'diag', count: 0 },
    security: { key: 'security', label: '安全扫描', icon: '🛡️', color: 'sec', count: 0 },
    custom: { key: 'custom', label: '自定义', icon: '🧩', color: 'cust', count: 0 }
  }
  agents.value.forEach(a => { if (map[a.category]) map[a.category].count++ })
  return Object.values(map).filter(s => s.count > 0)
})
const activeCount = computed(() => agents.value.filter(a => a.status === 'active').length)

// 标签辅助
const getCategoryLabel = (c) => ({ observability: '可观测性', diagnostics: '诊断工具', security: '安全扫描', custom: '自定义' }[c] || c)
const getCategoryEmoji = (c) => ({ observability: '📡', diagnostics: '🔬', security: '🛡️', custom: '🧩' }[c] || '📦')
const getScopeLabel = (s) => ({ java: 'Java', go: 'Go', python: 'Python', all: '通用' }[s] || s)

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++ }
  return size.toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

let searchTimer = null
const debounceSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadAgents(), 300)
}

const loadAgents = async () => {
  loading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if (keyword.value) params.keyword = keyword.value
    if (filters.value.category) params.category = filters.value.category
    if (filters.value.scope) params.scope = filters.value.scope
    if (filters.value.status) params.status = filters.value.status
    const res = await getBuildAgents(params)
    agents.value = res?.data?.list || res?.list || []
  } catch (e) {
    console.error('加载探针列表失败', e)
  } finally {
    loading.value = false
  }
}

const handleFileChange = (e) => {
  const file = e.target.files[0]
  if (file) {
    uploadForm.value.file = file
    if (!uploadForm.value.name) {
      uploadForm.value.name = file.name.replace(/\.[^/.]+$/, '')
    }
    if (!uploadForm.value.display_name) {
      uploadForm.value.display_name = file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' ')
    }
  }
}
const handleDrop = (e) => {
  isDragover.value = false
  const file = e.dataTransfer.files[0]
  if (file) {
    uploadForm.value.file = file
    if (!uploadForm.value.name) uploadForm.value.name = file.name.replace(/\.[^/.]+$/, '')
    if (!uploadForm.value.display_name) uploadForm.value.display_name = file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' ')
  }
}

const handleUpload = async () => {
  if (!uploadForm.value.file || !uploadForm.value.name) return
  uploading.value = true
  uploadProgress.value = 0
  try {
    const fd = new FormData()
    fd.append('file', uploadForm.value.file)
    for (const key of ['name', 'display_name', 'category', 'scope', 'version', 'description', 'docker_copy_dest', 'env_key', 'env_value']) {
      if (uploadForm.value[key]) fd.append(key, uploadForm.value[key])
    }
    await uploadBuildAgent(fd, (e) => {
      if (e.total > 0) uploadProgress.value = Math.round((e.loaded / e.total) * 100)
    })
    showUploadModal.value = false
    resetUploadForm()
    loadAgents()
  } catch (e) {
    console.error('上传失败', e)
    alert('上传失败: ' + (e?.response?.data?.msg || e.message))
  } finally {
    uploading.value = false
  }
}

const resetUploadForm = () => {
  uploadForm.value = { file: null, name: '', display_name: '', category: 'observability', scope: 'java', version: '', description: '', docker_copy_dest: '', env_key: '', env_value: '' }
}

const openEditModal = (agent) => {
  editForm.value = { id: agent.id, display_name: agent.display_name, version: agent.version, category: agent.category, scope: agent.scope, docker_copy_dest: agent.docker_copy_dest, env_key: agent.env_key, env_value: agent.env_value, description: agent.description }
  showEditModal.value = true
}
const handleUpdate = async () => {
  updating.value = true
  try {
    await updateBuildAgent(editForm.value)
    showEditModal.value = false
    loadAgents()
  } catch (e) {
    alert('更新失败: ' + (e?.response?.data?.msg || e.message))
  } finally {
    updating.value = false
  }
}

const handleToggle = async (agent) => {
  try {
    await toggleBuildAgent(agent.id)
    loadAgents()
  } catch (e) {
    alert('切换状态失败')
  }
}

const handleDownload = (agent) => {
  window.open(downloadBuildAgent(agent.id), '_blank')
}

const handleDelete = (agent) => {
  deleteTarget.value = agent
  showDeleteConfirm.value = true
}
const confirmDelete = async () => {
  deleting.value = true
  try {
    await deleteBuildAgent(deleteTarget.value.id)
    showDeleteConfirm.value = false
    loadAgents()
  } catch (e) {
    alert('删除失败')
  } finally {
    deleting.value = false
  }
}

onMounted(() => loadAgents())
</script>

<style scoped>
.agents-page { min-height: 100vh; background: #f8fafc; }

/* ===== Hero ===== */
.page-hero { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 2rem 2.5rem 1.75rem; position: relative; overflow: hidden; }
.page-hero::before { content: ''; position: absolute; top: -50%; right: -20%; width: 60%; height: 200%; background: radial-gradient(ellipse, rgba(255,255,255,0.08) 0%, transparent 70%); pointer-events: none; }
.hero-content { display: flex; justify-content: space-between; align-items: flex-end; position: relative; z-index: 1; }
.hero-breadcrumb { display: flex; align-items: center; gap: 6px; font-size: 0.8rem; color: rgba(255,255,255,0.65); margin-bottom: 0.5rem; }
.breadcrumb-current { color: rgba(255,255,255,0.9); }
.breadcrumb-sep { color: rgba(255,255,255,0.35); }
.hero-title { display: flex; align-items: center; gap: 12px; font-size: 1.6rem; font-weight: 700; color: #fff; margin: 0; }
.hero-icon-wrapper { width: 40px; height: 40px; background: rgba(255,255,255,0.15); border-radius: 12px; display: flex; align-items: center; justify-content: center; backdrop-filter: blur(10px); }
.hero-icon-wrapper svg { width: 22px; height: 22px; color: #fff; }
.hero-desc { color: rgba(255,255,255,0.7); font-size: 0.875rem; margin-top: 0.5rem; }
.hero-right { display: flex; gap: 10px; }
.hero-btn { display: flex; align-items: center; gap: 6px; padding: 0.55rem 1.1rem; border-radius: 10px; font-size: 0.85rem; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s; }
.hero-btn svg { width: 16px; height: 16px; }
.hero-btn.primary { background: #fff; color: #667eea; box-shadow: 0 2px 12px rgba(0,0,0,0.15); }
.hero-btn.primary:hover { transform: translateY(-1px); box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
.hero-btn.secondary { background: rgba(255,255,255,0.15); color: #fff; backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.2); }
.hero-btn.secondary:hover { background: rgba(255,255,255,0.25); }

/* ===== Overview Cards ===== */
.overview-section { padding: 1.25rem 2.5rem 0; margin-top: -1rem; position: relative; z-index: 2; }
.overview-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 1rem; }
.overview-card { background: #fff; border-radius: 14px; padding: 1.1rem 1.25rem; display: flex; align-items: center; gap: 14px; box-shadow: 0 1px 4px rgba(0,0,0,0.06), 0 4px 16px rgba(0,0,0,0.03); transition: all 0.25s; }
.overview-card:hover { transform: translateY(-2px); box-shadow: 0 8px 25px rgba(0,0,0,0.08); }
.oc-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.oc-icon svg { width: 22px; height: 22px; }
.oc-icon.total { background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; }
.oc-icon.obs { background: linear-gradient(135deg, #06b6d4, #0891b2); color: #fff; }
.oc-icon.diag { background: linear-gradient(135deg, #f59e0b, #d97706); color: #fff; }
.oc-icon.sec { background: linear-gradient(135deg, #ef4444, #dc2626); color: #fff; }
.oc-icon.cust { background: linear-gradient(135deg, #8b5cf6, #7c3aed); color: #fff; }
.oc-icon.active { background: linear-gradient(135deg, #10b981, #059669); color: #fff; }
.cat-emoji { font-size: 1.25rem; }
.oc-body { flex: 1; }
.oc-value { font-size: 1.4rem; font-weight: 700; color: #1e293b; line-height: 1.2; }
.oc-label { font-size: 0.78rem; color: #94a3b8; margin-top: 2px; }

/* ===== Toolbar ===== */
.main-section { padding: 1.25rem 2.5rem 2rem; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.25rem; gap: 1rem; flex-wrap: wrap; }
.toolbar-left { flex: 1; min-width: 260px; max-width: 420px; }
.search-box { position: relative; }
.search-box input { width: 100%; padding: 0.6rem 2.5rem 0.6rem 2.5rem; border: 1.5px solid #e2e8f0; border-radius: 10px; font-size: 0.85rem; background: #fff; transition: all 0.2s; outline: none; }
.search-box input:focus { border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,0.1); }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: #94a3b8; }
.search-clear { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; padding: 2px; color: #94a3b8; }
.search-clear svg { width: 14px; height: 14px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.filter-chips { display: flex; gap: 8px; }
.chip-select { padding: 0.5rem 0.75rem; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 0.8rem; background: #fff; color: #475569; cursor: pointer; outline: none; transition: border-color 0.2s; }
.chip-select:focus { border-color: #667eea; }

/* ===== States ===== */
.state-container { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 4rem 1rem; }
.loader { display: flex; gap: 6px; }
.loader-ring { width: 10px; height: 10px; border-radius: 50%; background: #667eea; animation: pulse 1.2s ease-in-out infinite; }
.loader-ring:nth-child(2) { animation-delay: 0.15s; }
.loader-ring:nth-child(3) { animation-delay: 0.3s; }
@keyframes pulse { 0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; } 40% { transform: scale(1); opacity: 1; } }
.state-text { color: #94a3b8; font-size: 0.9rem; margin-top: 1rem; }
.empty-illustration svg { width: 160px; height: 130px; }
.empty-title { color: #334155; font-size: 1.1rem; margin: 1rem 0 0.5rem; }
.empty-desc { color: #94a3b8; font-size: 0.85rem; max-width: 400px; text-align: center; }
.empty-action { display: inline-flex; align-items: center; gap: 6px; margin-top: 1.5rem; padding: 0.6rem 1.5rem; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; border: none; border-radius: 10px; font-size: 0.85rem; font-weight: 600; cursor: pointer; transition: all 0.2s; text-decoration: none; }
.empty-action svg { width: 16px; height: 16px; }
.empty-action:hover { transform: translateY(-1px); box-shadow: 0 4px 15px rgba(102,126,234,0.35); }

/* ===== Agent Grid ===== */
.agent-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 1.25rem; }
.agent-card { background: #fff; border-radius: 16px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,0.05), 0 4px 16px rgba(0,0,0,0.03); transition: all 0.3s cubic-bezier(0.4,0,0.2,1); border: 1.5px solid transparent; }
.agent-card:hover { transform: translateY(-3px); box-shadow: 0 12px 35px rgba(0,0,0,0.1); border-color: rgba(102,126,234,0.15); }
.agent-card.inactive { opacity: 0.65; }
.agent-card.inactive:hover { opacity: 0.85; }

.card-header { display: flex; justify-content: space-between; align-items: center; padding: 0.9rem 1.25rem 0; }
.card-category { font-size: 0.7rem; font-weight: 600; padding: 3px 10px; border-radius: 20px; text-transform: uppercase; letter-spacing: 0.5px; }
.card-category.observability { background: #ecfeff; color: #0891b2; }
.card-category.diagnostics { background: #fffbeb; color: #d97706; }
.card-category.security { background: #fef2f2; color: #dc2626; }
.card-category.custom { background: #f5f3ff; color: #7c3aed; }
.card-status { display: flex; align-items: center; gap: 5px; font-size: 0.75rem; font-weight: 500; }
.card-status.active { color: #059669; }
.card-status.inactive { color: #94a3b8; }
.status-dot { width: 7px; height: 7px; border-radius: 50%; }
.card-status.active .status-dot { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.4); }
.card-status.inactive .status-dot { background: #cbd5e1; }

.card-body { padding: 1rem 1.25rem; }
.card-icon-row { display: flex; align-items: center; gap: 12px; margin-bottom: 0.75rem; }
.card-agent-icon { width: 48px; height: 48px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; flex-shrink: 0; }
.card-agent-icon.observability { background: linear-gradient(135deg, #e0f2fe, #cffafe); }
.card-agent-icon.diagnostics { background: linear-gradient(135deg, #fef3c7, #fde68a); }
.card-agent-icon.security { background: linear-gradient(135deg, #fee2e2, #fecaca); }
.card-agent-icon.custom { background: linear-gradient(135deg, #ede9fe, #ddd6fe); }
.card-info { flex: 1; min-width: 0; }
.card-name { font-size: 0.95rem; font-weight: 650; color: #1e293b; margin: 0 0 4px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-meta-line { display: flex; gap: 6px; flex-wrap: wrap; }
.meta-tag { font-size: 0.68rem; font-weight: 600; padding: 2px 8px; border-radius: 6px; }
.meta-tag.scope { background: #f1f5f9; color: #475569; }
.meta-tag.scope.java { background: #fef3c7; color: #92400e; }
.meta-tag.scope.go { background: #d1fae5; color: #065f46; }
.meta-tag.scope.python { background: #dbeafe; color: #1e40af; }
.meta-tag.version { background: #f0fdf4; color: #166534; }
.card-desc { font-size: 0.8rem; color: #64748b; line-height: 1.5; margin: 0.5rem 0; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.card-stats { display: flex; gap: 1rem; margin-top: 0.75rem; padding-top: 0.75rem; border-top: 1px solid #f1f5f9; }
.stat-item { display: flex; align-items: center; gap: 4px; font-size: 0.75rem; color: #94a3b8; }
.stat-item svg { width: 13px; height: 13px; }

.card-footer { display: flex; justify-content: flex-end; gap: 4px; padding: 0.5rem 1rem 0.75rem; border-top: 1px solid #f8fafc; }
.card-action-btn { width: 32px; height: 32px; border-radius: 8px; border: none; background: transparent; color: #64748b; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.card-action-btn svg { width: 15px; height: 15px; }
.card-action-btn:hover { background: #f1f5f9; color: #334155; }
.card-action-btn.warn:hover { background: #fef3c7; color: #d97706; }
.card-action-btn.danger:hover { background: #fee2e2; color: #dc2626; }

/* ===== Modal ===== */
.modal-overlay { position: fixed; inset: 0; background: rgba(15,23,42,0.5); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-container { background: #fff; border-radius: 20px; width: 90%; box-shadow: 0 25px 60px rgba(0,0,0,0.15); overflow: hidden; }
.upload-modal, .edit-modal { max-width: 680px; }
.delete-modal { max-width: 440px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 1.25rem 1.75rem; border-bottom: 1px solid #f1f5f9; }
.modal-header h2 { font-size: 1.1rem; font-weight: 650; color: #1e293b; margin: 0; }
.modal-header.danger h2 { color: #dc2626; }
.modal-close { width: 32px; height: 32px; border: none; background: #f1f5f9; border-radius: 8px; font-size: 1.2rem; color: #64748b; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.modal-close:hover { background: #e2e8f0; }
.modal-body { padding: 1.5rem 1.75rem; max-height: 65vh; overflow-y: auto; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 1rem 1.75rem; border-top: 1px solid #f1f5f9; }
.modal-btn { padding: 0.55rem 1.5rem; border-radius: 10px; font-size: 0.85rem; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s; }
.modal-btn.cancel { background: #f1f5f9; color: #475569; }
.modal-btn.cancel:hover { background: #e2e8f0; }
.modal-btn.confirm { background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; }
.modal-btn.confirm:hover { transform: translateY(-1px); box-shadow: 0 4px 15px rgba(102,126,234,0.35); }
.modal-btn.confirm:disabled { opacity: 0.5; cursor: not-allowed; transform: none; box-shadow: none; }
.modal-btn.danger { background: linear-gradient(135deg, #ef4444, #dc2626); color: #fff; }
.modal-btn.danger:hover { box-shadow: 0 4px 15px rgba(239,68,68,0.35); }

/* Upload Zone */
.upload-zone { position: relative; border: 2px dashed #d1d5db; border-radius: 14px; padding: 2rem; text-align: center; transition: all 0.25s; margin-bottom: 1.25rem; cursor: pointer; background: #fafbfc; }
.upload-zone.dragover { border-color: #667eea; background: rgba(102,126,234,0.04); }
.upload-zone.hasFile { border-style: solid; border-color: #10b981; background: #f0fdf4; }
.upload-input { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
.upload-placeholder svg { width: 40px; height: 40px; color: #94a3b8; margin-bottom: 0.75rem; }
.upload-placeholder p { color: #64748b; font-size: 0.85rem; margin: 0.25rem 0; }
.upload-link { color: #667eea; font-weight: 600; cursor: pointer; }
.upload-hint { font-size: 0.75rem; color: #94a3b8; }
.upload-file-info { display: flex; align-items: center; gap: 12px; }
.file-icon { font-size: 2rem; }
.file-detail { display: flex; flex-direction: column; flex: 1; text-align: left; }
.file-name { font-weight: 600; color: #1e293b; font-size: 0.9rem; }
.file-size { color: #94a3b8; font-size: 0.78rem; }
.file-remove { width: 28px; height: 28px; border-radius: 50%; border: none; background: #fee2e2; color: #dc2626; font-size: 1.1rem; cursor: pointer; display: flex; align-items: center; justify-content: center; position: relative; z-index: 2; }

/* Form Grid */
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 5px; }
.form-group.full-width { grid-column: 1 / -1; }
.form-group label { font-size: 0.8rem; font-weight: 600; color: #475569; }
.required { color: #ef4444; }
.form-group input, .form-group select, .form-group textarea { padding: 0.55rem 0.85rem; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 0.85rem; outline: none; transition: border-color 0.2s; background: #fff; color: #1e293b; font-family: inherit; }
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { border-color: #667eea; box-shadow: 0 0 0 3px rgba(102,126,234,0.08); }
.form-group textarea { resize: vertical; }

/* Upload Progress */
.upload-progress-bar-wrap { position: relative; height: 6px; background: #e2e8f0; margin: 0 1.75rem; border-radius: 3px; overflow: hidden; }
.upload-progress-fill { height: 100%; background: linear-gradient(90deg, #667eea, #764ba2); border-radius: 3px; transition: width 0.3s; }
.upload-progress-text { position: absolute; right: 0; top: -20px; font-size: 0.72rem; color: #667eea; font-weight: 600; }

/* Delete modal */
.delete-warning { font-size: 0.9rem; color: #334155; margin: 0 0 0.5rem; }
.delete-hint { font-size: 0.8rem; color: #94a3b8; }

/* Transitions */
.modal-enter-active, .modal-leave-active { transition: all 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal-container, .modal-leave-to .modal-container { transform: scale(0.95) translateY(10px); }

/* Responsive */
@media (max-width: 768px) {
  .page-hero { padding: 1.5rem; }
  .hero-content { flex-direction: column; align-items: flex-start; gap: 1rem; }
  .overview-section, .main-section { padding-left: 1rem; padding-right: 1rem; }
  .overview-cards { grid-template-columns: repeat(2, 1fr); }
  .agent-grid { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; }
  .toolbar-left { max-width: 100%; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>

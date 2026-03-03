<template>
  <div class="releases-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">发布管理</h1>
        <p class="page-desc">管理应用发布记录，支持回滚和重新部署</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-primary" @click="showCreateDialog = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          创建发布
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ stats.total }}</span>
          <span class="stat-label">总发布数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon deploying">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ stats.deploying }}</span>
          <span class="stat-label">部署中</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ stats.success }}</span>
          <span class="stat-label">发布成功</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon failed">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ stats.failed }}</span>
          <span class="stat-label">发布失败</span>
        </div>
      </div>
    </div>

    <!-- 过滤和搜索 -->
    <div class="filter-bar">
      <div :class="['search-wrapper', { focused: searchFocused }]">
        <div class="search-box">
          <div class="search-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/>
              <line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
          </div>
          <input 
            v-model="searchKeyword" 
            type="text" 
            placeholder="搜索应用名、工作负载、镜像..."
            @input="handleSearch"
            @focus="searchFocused = true"
            @blur="searchFocused = false"
          />
          <button 
            v-if="searchKeyword" 
            class="clear-btn" 
            @click="clearSearch"
            title="清除搜索"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="search-hints">
          <span class="hint-label">支持搜索:</span>
          <span class="hint-tag">应用名</span>
          <span class="hint-tag">工作负载</span>
          <span class="hint-tag">镜像仓库</span>
          <span class="hint-tag">镜像标签</span>
        </div>
      </div>
      <div class="filter-tags">
        <button 
          :class="['filter-tag', { active: statusFilter === '' }]"
          @click="statusFilter = ''"
        >全部</button>
        <button 
          :class="['filter-tag', { active: statusFilter === 'deploying' }]"
          @click="statusFilter = 'deploying'"
        >
          <span class="status-dot deploying"></span>
          部署中
        </button>
        <button 
          :class="['filter-tag', { active: statusFilter === 'success' }]"
          @click="statusFilter = 'success'"
        >
          <span class="status-dot success"></span>
          成功
        </button>
        <button 
          :class="['filter-tag', { active: statusFilter === 'failed' }]"
          @click="statusFilter = 'failed'"
        >
          <span class="status-dot failed"></span>
          失败
        </button>
        <button 
          :class="['filter-tag', { active: statusFilter === 'rollback' }]"
          @click="statusFilter = 'rollback'"
        >
          <span class="status-dot rollback"></span>
          已回滚
        </button>
      </div>
    </div>

    <!-- 发布列表 -->
    <div class="releases-list">
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <span>正在加载发布记录...</span>
      </div>

      <div v-else-if="filteredReleases.length === 0" class="empty-state">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
          <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
          <line x1="12" y1="22.08" x2="12" y2="12"/>
        </svg>
        <h3>暂无发布记录</h3>
        <p>点击右上角创建第一个发布</p>
      </div>

      <template v-else>
        <div 
          v-for="release in filteredReleases" 
          :key="release.id"
          :class="['release-card', `status-${normalizeStatus(release.status)}`]"
        >
          <div class="release-main">
            <div class="release-icon">
              <svg v-if="normalizeStatus(release.status) === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                <polyline points="22 4 12 14.01 9 11.01"/>
              </svg>
              <svg v-else-if="normalizeStatus(release.status) === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="15" y1="9" x2="9" y2="15"/>
                <line x1="9" y1="9" x2="15" y2="15"/>
              </svg>
              <svg v-else-if="normalizeStatus(release.status) === 'deploying'" class="spinning" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
              <svg v-else-if="normalizeStatus(release.status) === 'rollback'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="1 4 1 10 7 10"/>
                <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
            </div>
            <div class="release-info">
              <div class="release-header">
                <span class="release-name">{{ release.app_name || release.name || '-' }}</span>
                <span :class="['release-status', `status-${normalizeStatus(release.status)}`]">
                  {{ statusText(release.status) }}
                </span>
                <span v-if="release.strategy" class="release-strategy">
                  {{ strategyText(release.strategy) }}
                </span>
              </div>
              <!-- 第二行：工作负载 + 镜像 -->
              <div class="release-workload">
                <span class="workload-badge">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="3" width="18" height="18" rx="2"/>
                    <path d="M9 3v18"/>
                    <path d="M15 3v18"/>
                    <path d="M3 9h18"/>
                    <path d="M3 15h18"/>
                  </svg>
                  {{ release.workload_kind || 'Deployment' }}/{{ release.workload_name || '-' }}
                </span>
                <span v-if="release.container_name" class="container-badge">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                  </svg>
                  {{ release.container_name }}
                </span>
              </div>
              <!-- 第三行：镜像信息 -->
              <div v-if="release.image_repo" class="release-image">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                  <line x1="7" y1="2" x2="7" y2="22"/>
                  <line x1="17" y1="2" x2="17" y2="22"/>
                  <line x1="2" y1="12" x2="22" y2="12"/>
                  <line x1="2" y1="7" x2="7" y2="7"/>
                  <line x1="2" y1="17" x2="7" y2="17"/>
                  <line x1="17" y1="17" x2="22" y2="17"/>
                  <line x1="17" y1="7" x2="22" y2="7"/>
                </svg>
                <span class="image-text" :title="getFullImage(release)">
                  {{ formatImage(release) }}
                </span>
              </div>
              <!-- 第四行：元信息 -->
              <div class="release-meta">
                <span class="meta-item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                  </svg>
                  {{ release.namespace || 'default' }}
                </span>
                <span class="meta-item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                  {{ formatDate(release.created_at) }}
                </span>
                <span v-if="release.message" class="meta-item message" :title="release.message">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                  </svg>
                  {{ truncateMessage(release.message) }}
                </span>
              </div>
              <!-- 进度条（部署中状态） -->
              <div v-if="normalizeStatus(release.status) === 'deploying'" class="deploy-progress">
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: (release.progress || 30) + '%' }"></div>
                </div>
                <span class="progress-text">部署进度: {{ release.progress || 30 }}%</span>
              </div>
            </div>
          </div>
          <div class="release-actions">
            <button 
              class="action-btn" 
              @click="viewRelease(release)"
              title="查看详情"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                <circle cx="12" cy="12" r="3"/>
              </svg>
            </button>
            <button 
              v-if="normalizeStatus(release.status) === 'deploying'"
              class="action-btn cancel" 
              @click="cancelRelease(release)"
              title="取消发布"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="6" y="6" width="12" height="12" rx="2"/>
              </svg>
            </button>
            <button 
              v-if="normalizeStatus(release.status) === 'success'"
              class="action-btn rollback" 
              @click="rollbackRelease(release)"
              title="回滚"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="1 4 1 10 7 10"/>
                <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
              </svg>
            </button>
            <button 
              v-if="normalizeStatus(release.status) === 'failed'"
              class="action-btn retry" 
              @click="retryRelease(release)"
              title="重试"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
            </button>
            <button 
              class="action-btn logs" 
              @click="viewLogs(release)"
              title="查看日志"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
            </button>
          </div>
        </div>
      </template>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="pagination">
      <button 
        class="page-btn" 
        :disabled="currentPage === 1"
        @click="currentPage--"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
      <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
      <button 
        class="page-btn" 
        :disabled="currentPage === totalPages"
        @click="currentPage++"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </button>
    </div>

    <!-- 创建发布弹窗 -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>创建发布</h3>
          <button class="close-btn" @click="showCreateDialog = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label>选择流水线</label>
            <select v-model="createForm.pipeline_id">
              <option value="">请选择流水线</option>
              <option v-for="p in pipelines" :key="p.id" :value="p.id">
                {{ p.name }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>发布名称</label>
            <input v-model="createForm.name" type="text" placeholder="例如: v1.0.0-release" />
          </div>
          <div class="form-group">
            <label>版本号</label>
            <input v-model="createForm.version" type="text" placeholder="例如: v1.0.0" />
          </div>
          <div class="form-group">
            <label>目标命名空间</label>
            <input v-model="createForm.namespace" type="text" placeholder="例如: production" />
          </div>
          <div class="form-group">
            <label>镜像地址</label>
            <input v-model="createForm.image" type="text" placeholder="例如: registry.cn-hangzhou.aliyuncs.com/xxx/app:v1.0.0" />
          </div>
          <div class="form-group">
            <label>备注（可选）</label>
            <textarea v-model="createForm.remark" placeholder="发布说明..."></textarea>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-outline" @click="showCreateDialog = false">取消</button>
          <button class="btn btn-primary" @click="handleCreate" :disabled="creating">
            {{ creating ? '创建中...' : '创建发布' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 日志弹窗 -->
    <div v-if="showLogsDialog" class="dialog-overlay" @click.self="showLogsDialog = false">
      <div class="dialog large">
        <div class="dialog-header">
          <h3>发布日志 - {{ currentRelease?.name }}</h3>
          <button class="close-btn" @click="showLogsDialog = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body logs-body">
          <pre class="logs-content">{{ releaseLogs || '暂无日志...' }}</pre>
        </div>
      </div>
    </div>

    <!-- 确认弹窗 -->
    <div v-if="showConfirmDialog" class="dialog-overlay" @click.self="showConfirmDialog = false">
      <div class="dialog small">
        <div class="dialog-header">
          <h3>{{ confirmTitle }}</h3>
        </div>
        <div class="dialog-body">
          <p>{{ confirmMessage }}</p>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-outline" @click="showConfirmDialog = false">取消</button>
          <button :class="['btn', confirmBtnClass]" @click="confirmAction" :disabled="confirming">
            {{ confirming ? '处理中...' : confirmBtnText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { getPipelines } from '@/api/platform/pipeline'
import {
  getReleases,
  createRelease,
  cancelRelease as cancelReleaseApi,
  rollbackRelease as rollbackReleaseApi,
  retryRelease as retryReleaseApi
} from '@/api/cicd'

export default {
  name: 'Releases',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const releases = ref([])
    const searchKeyword = ref('')
    const searchFocused = ref(false)
    const statusFilter = ref('')
    const currentPage = ref(1)
    const pageSize = 10
    const total = ref(0)
    
    // 标准化状态（后端 Succeeded -> success，Running -> deploying）
    const normalizeStatus = (status) => {
      const map = {
        Pending: 'pending',
        Queued: 'deploying',
        Running: 'deploying',
        Succeeded: 'success',
        Failed: 'failed',
        Canceled: 'failed',
        Rollback: 'rollback'
      }
      return map[status] || status
    }
    
    // 统计（使用后端返回的数据统计，需要归一化状态）
    const stats = computed(() => {
      const list = releases.value
      return {
        total: total.value || list.length,
        deploying: list.filter(r => normalizeStatus(r.status) === 'deploying').length,
        success: list.filter(r => normalizeStatus(r.status) === 'success').length,
        failed: list.filter(r => normalizeStatus(r.status) === 'failed').length
      }
    })

    // 直接使用后端返回的列表（不做前端过滤）
    const filteredReleases = computed(() => releases.value)

    const totalPages = computed(() => Math.ceil(total.value / pageSize))

    // 加载数据
    const loadReleases = async () => {
      loading.value = true
      try {
        // 状态映射：前端小写 -> 后端首字母大写
        const statusMap = {
          deploying: 'Running',
          success: 'Succeeded',
          failed: 'Failed',
          rollback: 'Rollback',
          pending: 'Pending'
        }
        const backendStatus = statusFilter.value ? statusMap[statusFilter.value] : undefined
        
        const response = await getReleases({
          page: currentPage.value,
          page_size: pageSize,
          keyword: searchKeyword.value || undefined,
          status: backendStatus
        })
        if (response.code === 0) {
          releases.value = response.data?.list || []
          total.value = response.data?.total || 0
        } else {
          throw new Error(response.msg || '获取发布单列表失败')
        }
      } catch (error) {
        console.error('加载发布单失败:', error)
        Message.error({ content: error.message || '加载发布单失败' })
      } finally {
        loading.value = false
      }
    }

    // 流水线列表
    const pipelines = ref([])
    const loadPipelines = async () => {
      try {
        const response = await getPipelines()
        if (response.code === 0) {
          pipelines.value = response.data.list || response.data || []
        }
      } catch (error) {
        console.error('加载流水线失败:', error)
      }
    }

    // 创建发布
    const showCreateDialog = ref(false)
    const creating = ref(false)
    const createForm = ref({
      pipeline_id: '',
      name: '',
      version: '',
      namespace: 'production',
      image: '',
      remark: ''
    })

    const handleCreate = async () => {
      if (!createForm.value.name || !createForm.value.version) {
        Message.warning({ content: '请填写发布名称和版本号' })
        return
      }
      creating.value = true
      try {
        const response = await createRelease({
          pipeline_id: createForm.value.pipeline_id ? Number(createForm.value.pipeline_id) : undefined,
          name: createForm.value.name,
          version: createForm.value.version,
          namespace: createForm.value.namespace,
          image: createForm.value.image,
          description: createForm.value.remark
        })
        if (response.code === 0) {
          Message.success({ content: '发布创建成功' })
          showCreateDialog.value = false
          createForm.value = { pipeline_id: '', name: '', version: '', namespace: 'production', image: '', remark: '' }
          loadReleases()
        } else {
          throw new Error(response.msg || '创建失败')
        }
      } catch (error) {
        Message.error({ content: error.message || '创建发布单失败' })
      } finally {
        creating.value = false
      }
    }

    // 确认弹窗
    const showConfirmDialog = ref(false)
    const confirmTitle = ref('')
    const confirmMessage = ref('')
    const confirmBtnText = ref('确认')
    const confirmBtnClass = ref('btn-primary')
    const confirming = ref(false)
    const pendingAction = ref(null)

    const openConfirm = (title, message, btnText, btnClass, action) => {
      confirmTitle.value = title
      confirmMessage.value = message
      confirmBtnText.value = btnText
      confirmBtnClass.value = btnClass
      pendingAction.value = action
      showConfirmDialog.value = true
    }

    const confirmAction = async () => {
      if (pendingAction.value) {
        confirming.value = true
        try {
          await pendingAction.value()
        } finally {
          confirming.value = false
          showConfirmDialog.value = false
        }
      }
    }

    // 操作
    const viewRelease = (release) => {
      // 跳转到详情或打开弹窗
      Message.info({ content: `查看发布: ${release.name}` })
    }

    const cancelRelease = (release) => {
      openConfirm(
        '取消发布',
        `确定要取消发布 "${release.name}" 吗？这将中止当前部署过程。`,
        '取消发布',
        'btn-warning',
        async () => {
          const response = await cancelReleaseApi(release.id)
          if (response.code === 0) {
            Message.success({ content: '发布已取消' })
            loadReleases()
          } else {
            throw new Error(response.msg || '取消失败')
          }
        }
      )
    }

    const rollbackRelease = (release) => {
      openConfirm(
        '回滚发布',
        `确定要回滚发布 "${release.name}" 吗？这将恢复到上一个稳定版本。`,
        '确认回滚',
        'btn-warning',
        async () => {
          const response = await rollbackReleaseApi(release.id)
          if (response.code === 0) {
            Message.success({ content: '回滚成功' })
            loadReleases()
          } else {
            throw new Error(response.msg || '回滚失败')
          }
        }
      )
    }

    const retryRelease = (release) => {
      openConfirm(
        '重试发布',
        `确定要重新发布 "${release.name}" 吗？`,
        '重新发布',
        'btn-primary',
        async () => {
          const response = await retryReleaseApi(release.id)
          if (response.code === 0) {
            Message.success({ content: '已重新开始发布' })
            loadReleases()
          } else {
            throw new Error(response.msg || '重试失败')
          }
        }
      )
    }

    // 日志
    const showLogsDialog = ref(false)
    const currentRelease = ref(null)
    const releaseLogs = ref('')

    const viewLogs = async (release) => {
      currentRelease.value = release
      showLogsDialog.value = true
      releaseLogs.value = `[${new Date().toISOString()}] 开始发布 ${release.name}
[${new Date().toISOString()}] 拉取镜像中...
[${new Date().toISOString()}] 镜像拉取成功
[${new Date().toISOString()}] 更新 Deployment...
[${new Date().toISOString()}] 等待 Pod 就绪...
[${new Date().toISOString()}] ${release.status === 'success' ? '发布完成' : release.status === 'failed' ? '发布失败: 镜像拉取超时' : '发布中...'}`
    }

    // 搜索处理（防抖）
    let searchTimer = null
    const handleSearch = () => {
      if (searchTimer) clearTimeout(searchTimer)
      searchTimer = setTimeout(() => {
        currentPage.value = 1
        loadReleases()
      }, 300)
    }

    // 清除搜索
    const clearSearch = () => {
      searchKeyword.value = ''
      currentPage.value = 1
      loadReleases()
    }

    // 状态过滤变化时触发查询
    watch(statusFilter, () => {
      currentPage.value = 1
      loadReleases()
    })

    // 分页变化时触发查询
    watch(currentPage, () => {
      loadReleases()
    })

    const statusText = (status) => {
      const map = {
        deploying: '部署中',
        success: '发布成功',
        failed: '发布失败',
        rollback: '已回滚',
        pending: '等待中',
        // 后端返回的状态值映射
        Pending: '等待中',
        Queued: '排队中',
        Running: '部署中',
        Succeeded: '发布成功',
        Failed: '发布失败',
        Canceled: '已取消',
        Rollback: '已回滚'
      }
      return map[status] || status
    }

    // 发布策略文本
    const strategyText = (strategy) => {
      const map = {
        rolling: '滚动更新',
        recreate: '重建',
        canary: '金丝雀',
        bluegreen: '蓝绿部署'
      }
      return map[strategy] || strategy
    }

    // 格式化镜像（智能显示）
    const formatImage = (release) => {
      const repo = release.image_repo || ''
      const tag = release.image_tag || ''
      
      if (!repo && !tag) return '-'
      
      // 如果 repo 已经包含 tag（如 nginx:1.26），直接返回
      if (repo.includes(':')) return repo
      
      // 如果只有 repo 没有 tag
      if (repo && !tag) return repo
      
      // 组合 repo 和 tag
      // 如果 repo 很长，取最后两部分
      const parts = repo.split('/')
      const shortRepo = parts.length > 2 ? '.../' + parts.slice(-2).join('/') : repo
      return `${shortRepo}:${tag}`
    }

    // 获取完整镜像地址
    const getFullImage = (release) => {
      const repo = release.image_repo || ''
      const tag = release.image_tag || ''
      if (!repo) return '-'
      if (repo.includes(':')) return repo
      return tag ? `${repo}:${tag}` : repo
    }

    // 截断消息
    const truncateMessage = (msg) => {
      if (!msg) return ''
      return msg.length > 30 ? msg.substring(0, 30) + '...' : msg
    }

    const formatDate = (timestamp) => {
      if (!timestamp) return '-'
      // 后端返回的是秒级时间戳，需要转换为毫秒
      const ts = timestamp > 1e11 ? timestamp : timestamp * 1000
      const date = new Date(ts)
      const now = new Date()
      const diff = now - date
      if (diff < 0) return date.toLocaleDateString('zh-CN')
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
      if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
      return date.toLocaleDateString('zh-CN')
    }

    onMounted(() => {
      loadReleases()
      loadPipelines()
    })

    return {
      loading,
      releases,
      searchKeyword,
      searchFocused,
      statusFilter,
      currentPage,
      totalPages,
      stats,
      filteredReleases,
      pipelines,
      showCreateDialog,
      creating,
      createForm,
      handleCreate,
      showConfirmDialog,
      confirmTitle,
      confirmMessage,
      confirmBtnText,
      confirmBtnClass,
      confirming,
      confirmAction,
      viewRelease,
      cancelRelease,
      rollbackRelease,
      retryRelease,
      showLogsDialog,
      currentRelease,
      releaseLogs,
      viewLogs,
      handleSearch,
      clearSearch,
      statusText,
      normalizeStatus,
      strategyText,
      formatImage,
      getFullImage,
      truncateMessage,
      formatDate
    }
  }
}
</script>

<style scoped>
.releases-view {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  min-height: 100vh;
  background: #f5f7fa;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.page-desc {
  color: #64748b;
  margin: 0;
}

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon svg {
  width: 28px;
  height: 28px;
}

.stat-icon.total { background: linear-gradient(135deg, #667eea, #764ba2); color: white; }
.stat-icon.deploying { background: linear-gradient(135deg, #f6d365, #fda085); color: white; }
.stat-icon.success { background: linear-gradient(135deg, #11998e, #38ef7d); color: white; }
.stat-icon.failed { background: linear-gradient(135deg, #eb3349, #f45c43); color: white; }

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1a202c;
}

.stat-label {
  font-size: 14px;
  color: #64748b;
}

/* 过滤栏 */
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}

/* 搜索框包装器 */
.search-wrapper {
  flex: 1;
  max-width: 520px;
}

.search-box {
  display: flex;
  align-items: center;
  padding: 0;
  background: white;
  border-radius: 12px;
  border: 2px solid #e2e8f0;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.search-wrapper.focused .search-box {
  border-color: #4299e1;
  box-shadow: 0 0 0 4px rgba(66, 153, 225, 0.15), 0 4px 12px rgba(66, 153, 225, 0.2);
}

.search-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  flex-shrink: 0;
}

.search-icon svg {
  width: 20px;
  height: 20px;
}

.search-box input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 14px;
  padding: 14px 16px;
  background: transparent;
  color: #1a202c;
}

.search-box input::placeholder {
  color: #a0aec0;
}

.clear-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  margin-right: 8px;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: #fee2e2;
  color: #dc2626;
}

.clear-btn svg {
  width: 16px;
  height: 16px;
}

/* 搜索提示标签 */
.search-hints {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  padding-left: 4px;
  flex-wrap: wrap;
}

.hint-label {
  font-size: 12px;
  color: #94a3b8;
}

.hint-tag {
  font-size: 11px;
  padding: 3px 8px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  color: #0284c7;
  border-radius: 10px;
  border: 1px solid #bae6fd;
}

.filter-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  background: white;
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tag:hover {
  border-color: #cbd5e0;
}

.filter-tag.active {
  background: #4299e1;
  border-color: #4299e1;
  color: white;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.deploying { background: #d97706; }
.status-dot.success { background: #059669; }
.status-dot.failed { background: #dc2626; }
.status-dot.rollback { background: #7c3aed; }

.filter-tag.active .status-dot {
  background: white;
}

/* 发布列表 */
.releases-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px;
  background: white;
  border-radius: 16px;
  color: #64748b;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e2e8f0;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state svg {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  color: #94a3b8;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  color: #4a5568;
}

.empty-state p {
  margin: 0;
}

/* 发布卡片 */
.release-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: white;
  border-radius: 16px;
  border-left: 4px solid #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: all 0.2s;
}

.release-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.release-card.status-deploying { border-left-color: #d97706; }
.release-card.status-success { border-left-color: #059669; }
.release-card.status-failed { border-left-color: #dc2626; }
.release-card.status-rollback { border-left-color: #7c3aed; }

.release-main {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  flex: 1;
}

.release-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.release-icon svg {
  width: 24px;
  height: 24px;
}

.release-card.status-deploying .release-icon { background: #fef3c7; color: #d97706; }
.release-card.status-success .release-icon { background: #d1fae5; color: #059669; }
.release-card.status-failed .release-icon { background: #fee2e2; color: #dc2626; }
.release-card.status-rollback .release-icon { background: #ede9fe; color: #7c3aed; }

.spinning {
  animation: spin 1s linear infinite;
}

.release-info {
  flex: 1;
}

.release-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.release-name {
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
}

.release-status {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 12px;
}

.release-status.status-deploying { background: #fef3c7; color: #d97706; }
.release-status.status-success { background: #d1fae5; color: #059669; }
.release-status.status-failed { background: #fee2e2; color: #dc2626; }
.release-status.status-rollback { background: #ede9fe; color: #7c3aed; }
.release-status.status-pending { background: #f1f5f9; color: #64748b; }

/* 发布策略标签 */
.release-strategy {
  font-size: 11px;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 10px;
  background: #f1f5f9;
  color: #64748b;
}

/* 工作负载行 */
.release-workload {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.workload-badge, .container-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #475569;
  padding: 4px 10px;
  background: #f1f5f9;
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.workload-badge svg, .container-badge svg {
  width: 14px;
  height: 14px;
  color: #64748b;
}

/* 镜像信息行 */
.release-image {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding: 6px 10px;
  background: #fefce8;
  border-radius: 6px;
  border-left: 3px solid #eab308;
}

.release-image svg {
  width: 16px;
  height: 16px;
  color: #ca8a04;
  flex-shrink: 0;
}

.image-text {
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: #854d0e;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 400px;
}

.release-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #64748b;
}

.meta-item.message {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #94a3b8;
  font-style: italic;
}

.meta-item svg {
  width: 14px;
  height: 14px;
  color: #94a3b8;
}

.deploy-progress {
  margin-top: 12px;
}

.progress-bar {
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #f6d365, #fda085);
  border-radius: 3px;
  transition: width 0.3s;
}

.progress-text {
  font-size: 12px;
  color: #d97706;
  margin-top: 4px;
  display: block;
}

.release-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 40px;
  height: 40px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  color: #64748b;
}

.action-btn:hover {
  border-color: #4299e1;
  color: #4299e1;
  background: #ebf8ff;
}

.action-btn.cancel:hover {
  border-color: #d97706;
  color: #d97706;
  background: #fef3c7;
}

.action-btn.rollback:hover {
  border-color: #7c3aed;
  color: #7c3aed;
  background: #ede9fe;
}

.action-btn.retry:hover {
  border-color: #059669;
  color: #059669;
  background: #d1fae5;
}

.action-btn svg {
  width: 18px;
  height: 18px;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.page-btn {
  width: 40px;
  height: 40px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: #4299e1;
  color: #4299e1;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-btn svg {
  width: 20px;
  height: 20px;
}

.page-info {
  font-size: 14px;
  color: #64748b;
}

/* 弹窗 */
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dialog.large {
  max-width: 800px;
}

.dialog.small {
  max-width: 400px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.dialog-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f1f5f9;
  color: #1a202c;
}

.close-btn svg {
  width: 20px;
  height: 20px;
}

.dialog-body {
  padding: 24px;
  overflow-y: auto;
}

.dialog-body.logs-body {
  background: #1e293b;
  padding: 0;
}

.logs-content {
  padding: 20px;
  margin: 0;
  font-family: 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #e2e8f0;
  white-space: pre-wrap;
  min-height: 300px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
}

/* 表单 */
.form-group {
  margin-bottom: 20px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
  margin-bottom: 8px;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.15);
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.btn svg {
  width: 18px;
  height: 18px;
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
}

.btn-warning {
  background: linear-gradient(135deg, #ed8936, #dd6b20);
  color: white;
}

.btn-warning:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(237, 137, 54, 0.4);
}

.btn-outline {
  background: white;
  color: #4a5568;
  border-color: #e2e8f0;
}

.btn-outline:hover {
  border-color: #cbd5e0;
  background: #f7fafc;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .stats-row {
    grid-template-columns: 1fr;
  }
  
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    max-width: none;
  }
  
  .filter-tags {
    overflow-x: auto;
    padding-bottom: 8px;
  }
  
  .release-card {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .release-actions {
    justify-content: flex-end;
  }
}
</style>

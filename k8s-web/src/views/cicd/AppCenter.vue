<template>
  <div class="app-center">
    <!-- 页面标题 - 深色 Banner 风格（与其他CI/CD页面统一） -->
    <div class="page-banner">
      <div class="banner-inner">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M12 2L2 7l10 5 10-5-10-5z"/>
              <path d="M2 17l10 5 10-5"/>
              <path d="M2 12l10 5 10-5"/>
            </svg>
          </div>
          <div>
            <h1 class="banner-title">应用中心</h1>
            <p class="banner-desc">统一管理应用生命周期：代码 → 构建 → 部署 → 运维</p>
          </div>
        </div>
        <div class="banner-actions">
          <button class="btn-action" @click="loadApps" :disabled="loading">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            <span>{{ loading ? '加载中...' : '刷新' }}</span>
          </button>
          <button class="btn-action primary" @click="$router.push('/cicd/pipelines/create')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            <span>注册应用</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 统计概览 -->
    <div class="stats-row">
      <div class="stat-item">
        <span class="stat-number">{{ apps.length }}</span>
        <span class="stat-label">应用总数</span>
      </div>
      <div class="stat-item running">
        <span class="stat-number">{{ runningCount }}</span>
        <span class="stat-label">构建中</span>
      </div>
      <div class="stat-item success">
        <span class="stat-number">{{ healthyCount }}</span>
        <span class="stat-label">运行正常</span>
      </div>
      <div class="stat-item warning">
        <span class="stat-number">{{ failedCount }}</span>
        <span class="stat-label">异常</span>
      </div>
    </div>

    <!-- 搜索与视图切换 -->
    <div class="toolbar">
      <div class="search-wrapper">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input v-model="searchQuery" type="text" class="search-input" placeholder="搜索应用名称、Git 仓库..." />
      </div>
      <div class="toolbar-right">
        <select v-model="envFilter" class="filter-select">
          <option value="">全部环境</option>
          <option value="dev">开发</option>
          <option value="test">测试</option>
          <option value="staging">预发布</option>
          <option value="production">生产</option>
        </select>
        <div class="view-switch">
          <button :class="['view-btn', { active: viewMode === 'card' }]" @click="viewMode = 'card'" title="卡片视图">
            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 3h8v8H3V3zm0 10h8v8H3v-8zm10-10h8v8h-8V3zm0 10h8v8h-8v-8z"/></svg>
          </button>
          <button :class="['view-btn', { active: viewMode === 'list' }]" @click="viewMode = 'list'" title="列表视图">
            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 4h18v2H3V4zm0 7h18v2H3v-2zm0 7h18v2H3v-2z"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>正在加载应用列表...</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="filteredApps.length === 0 && !loading" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="8" y="12" width="48" height="36" rx="4"/>
          <path d="M8 22h48"/>
          <circle cx="16" cy="17" r="2" fill="currentColor"/>
          <circle cx="22" cy="17" r="2" fill="currentColor"/>
          <circle cx="28" cy="17" r="2" fill="currentColor"/>
          <path d="M24 38l6-6 6 6"/>
          <path d="M30 32v12"/>
        </svg>
      </div>
      <h3>{{ searchQuery ? '没有匹配的应用' : '还没有注册应用' }}</h3>
      <p>{{ searchQuery ? '请尝试其他关键词' : '注册应用后，平台将自动为您配置 CI/CD 流水线' }}</p>
      <button v-if="!searchQuery" class="btn-action primary" style="margin-top: 16px;" @click="$router.push('/cicd/pipelines/create')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>注册第一个应用</span>
        </button>
    </div>

    <!-- 卡片视图 -->
    <div v-else-if="viewMode === 'card'" class="app-grid">
      <div v-for="app in filteredApps" :key="app.id" class="app-card" :class="getAppStatusClass(app)">
        <!-- 卡片头部 -->
        <div class="card-header">
          <div class="app-avatar" :class="getLanguageClass(app.language)">
            {{ (app.name || '?').charAt(0).toUpperCase() }}
          </div>
          <div class="app-meta">
            <h3 class="app-name" @click="viewApp(app)">{{ app.name }}</h3>
            <span class="app-lang">{{ getLanguageLabel(app.language) }}</span>
          </div>
          <div class="app-status-badge" :class="getAppStatus(app)">
            <span class="status-dot"></span>
            {{ getAppStatusText(app) }}
          </div>
        </div>

        <!-- 核心信息 - 3层结构 -->
        <div class="card-body">
          <!-- Tier 1: 基础信息 -->
          <div class="info-tier">
            <div class="info-item">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/></svg>
              <span class="info-text" :title="app.git_repo">{{ formatGitRepo(app.git_repo) }}</span>
            </div>
            <div class="info-item">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 3v12"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>
              <span class="info-text">{{ app.branch || 'main' }}</span>
            </div>
          </div>

          <!-- Tier 2: 部署信息 -->
          <div class="info-tier deploy-tier">
            <div class="info-item">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              <span class="info-text ns-tag">{{ app.target_namespace || 'default' }}</span>
            </div>
            <div class="info-item" v-if="app.target_workload_name">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
              <span class="info-text">{{ app.target_workload_kind || 'Deployment' }}/{{ app.target_workload_name }}</span>
            </div>
          </div>

          <!-- Tier 3: 最近构建/发布状态 -->
          <div class="info-tier build-tier" v-if="app.lastRunStatus">
            <div class="last-build">
              <span class="build-status" :class="app.lastRunStatus">
                <span class="status-dot-sm"></span>
                {{ getBuildStatusText(app.lastRunStatus) }}
              </span>
              <span class="build-time">{{ formatTime(app.lastRunTime) }}</span>
            </div>
          </div>
        </div>

        <!-- 卡片操作栏 -->
        <div class="card-actions">
          <button class="action-btn primary" @click="runApp(app)" :disabled="app.lastRunStatus === 'running'" title="发布">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            <span>发布</span>
          </button>
          <button class="action-btn" @click="viewApp(app)" title="详情">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
            <span>详情</span>
          </button>
          <button class="action-btn" @click="editApp(app)" title="配置">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68 1.65 1.65 0 0 0 10 3.17V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            <span>配置</span>
          </button>
          <button class="action-btn" @click="viewReleases(app)" title="发布历史">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <span>历史</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 列表视图 -->
    <div v-else class="app-list">
      <div class="list-header">
        <span class="col-app">应用</span>
        <span class="col-env">环境</span>
        <span class="col-status">状态</span>
        <span class="col-branch">分支</span>
        <span class="col-last">最近构建</span>
        <span class="col-actions">操作</span>
      </div>
      <div v-for="app in filteredApps" :key="app.id" class="list-row" :class="getAppStatusClass(app)">
        <div class="col-app">
          <div class="app-avatar-sm" :class="getLanguageClass(app.language)">
            {{ (app.name || '?').charAt(0).toUpperCase() }}
          </div>
          <div class="app-info-col">
            <span class="app-name-text" @click="viewApp(app)">{{ app.name }}</span>
            <span class="app-repo-text">{{ formatGitRepo(app.git_repo) }}</span>
          </div>
        </div>
        <div class="col-env">
          <span class="env-badge">{{ app.target_namespace || 'default' }}</span>
        </div>
        <div class="col-status">
          <span class="status-pill" :class="getAppStatus(app)">
            <span class="status-dot"></span>
            {{ getAppStatusText(app) }}
          </span>
        </div>
        <div class="col-branch">
          <code>{{ app.branch || 'main' }}</code>
        </div>
        <div class="col-last">
          <span v-if="app.lastRunTime" class="time-text">{{ formatTime(app.lastRunTime) }}</span>
          <span v-else class="text-muted">-</span>
        </div>
        <div class="col-actions">
          <button class="row-action-btn primary" @click="runApp(app)" :disabled="app.lastRunStatus === 'running'" title="发布">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
          </button>
          <button class="row-action-btn" @click="viewApp(app)" title="详情">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
          </button>
          <button class="row-action-btn" @click="editApp(app)" title="配置">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { getPipelines, triggerPipeline } from '@/api/platform/pipeline'

export default {
  name: 'AppCenter',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const apps = ref([])
    const searchQuery = ref('')
    const envFilter = ref('')
    const viewMode = ref('card')

    // 加载应用列表（基于流水线聚合）
    const loadApps = async () => {
      loading.value = true
      try {
        const res = await getPipelines({ page: 1, page_size: 100 })
        if (res.code === 0) {
          apps.value = (res.data?.list || res.data || []).map(p => ({
            ...p,
            language: p.language || detectLanguage(p),
            branch: p.branch || p.git_branch || 'main',
            git_repo: p.git_repo || p.git_url || '',
            lastRunStatus: p.last_run_status || p.lastRunStatus || '',
            lastRunTime: p.last_run_time || p.lastRunTime || '',
            target_namespace: p.target_namespace || 'default',
            target_workload_name: p.target_workload_name || '',
            target_workload_kind: p.target_workload_kind || 'Deployment',
            target_cluster_id: p.target_cluster_id || 0,
          }))
        }
      } catch (e) {
        Message.error({ content: '加载应用列表失败' })
        console.error(e)
      } finally {
        loading.value = false
      }
    }

    // 语言探测
    const detectLanguage = (p) => {
      const name = (p.name || '').toLowerCase()
      const env = JSON.stringify(p.env_vars || []).toLowerCase()
      if (env.includes('java') || name.includes('spring')) return 'java'
      if (env.includes('golang') || env.includes('go build')) return 'go'
      if (env.includes('python') || env.includes('pip')) return 'python'
      if (env.includes('npm') || env.includes('node') || name.includes('frontend')) return 'nodejs'
      return p.language || 'unknown'
    }

    // 统计
    const runningCount = computed(() => apps.value.filter(a => a.lastRunStatus === 'running' || a.lastRunStatus === 'building').length)
    const healthyCount = computed(() => apps.value.filter(a => a.lastRunStatus === 'success' || a.lastRunStatus === 'SUCCESS').length)
    const failedCount = computed(() => apps.value.filter(a => a.lastRunStatus === 'failed' || a.lastRunStatus === 'FAILURE').length)

    // 过滤
    const filteredApps = computed(() => {
      let result = apps.value
      if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(a =>
          (a.name || '').toLowerCase().includes(q) ||
          (a.git_repo || '').toLowerCase().includes(q) ||
          (a.description || '').toLowerCase().includes(q)
        )
      }
      if (envFilter.value) {
        result = result.filter(a => {
          const ns = (a.target_namespace || '').toLowerCase()
          if (envFilter.value === 'production') return ns.includes('prod') || ns === 'production'
          if (envFilter.value === 'staging') return ns.includes('stag') || ns === 'staging' || ns === 'pre'
          if (envFilter.value === 'test') return ns.includes('test') || ns === 'qa'
          if (envFilter.value === 'dev') return ns.includes('dev') || ns === 'development'
          return true
        })
      }
      return result
    })

    // 工具函数
    const getLanguageLabel = (lang) => {
      const map = { java: 'Java', go: 'Go', python: 'Python', nodejs: 'Node.js', unknown: '-' }
      return map[lang] || lang || '-'
    }

    const getLanguageClass = (lang) => `lang-${lang || 'unknown'}`

    const getAppStatus = (app) => {
      const s = (app.lastRunStatus || '').toLowerCase()
      if (s === 'running' || s === 'building') return 'running'
      if (s === 'success') return 'healthy'
      if (s === 'failed' || s === 'failure') return 'error'
      return 'idle'
    }

    const getAppStatusText = (app) => {
      const s = (app.lastRunStatus || '').toLowerCase()
      if (s === 'running' || s === 'building') return '构建中'
      if (s === 'success') return '运行正常'
      if (s === 'failed' || s === 'failure') return '构建失败'
      return '待部署'
    }

    const getAppStatusClass = (app) => `status-${getAppStatus(app)}`

    const getBuildStatusText = (status) => {
      const map = { running: '构建中', building: '构建中', success: '成功', SUCCESS: '成功', failed: '失败', FAILURE: '失败' }
      return map[status] || status || '-'
    }

    const formatGitRepo = (url) => {
      if (!url) return '-'
      return url.replace(/^https?:\/\//, '').replace(/\.git$/, '')
    }

    const formatTime = (t) => {
      if (!t) return '-'
      const date = new Date(typeof t === 'number' && t < 1e12 ? t * 1000 : t)
      const now = new Date()
      const diff = now - date
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
      if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
      return date.toLocaleDateString('zh-CN')
    }

    // 操作
    const viewApp = (app) => router.push(`/cicd/pipelines/${app.id}`)
    const editApp = (app) => router.push(`/cicd/pipelines/${app.id}/edit`)
    const viewReleases = (app) => router.push(`/cicd/releases?app=${app.name}`)

    const runApp = async (app) => {
      try {
        Message.info({ content: `正在启动 "${app.name}" 发布...` })
        const res = await triggerPipeline(app.id)
        if (res.code === 0) {
          Message.success({ content: '发布启动成功，正在跳转构建界面...' })
          router.push(`/cicd/pipelines/${app.id}?tab=stages`)
        } else {
          throw new Error(res.msg || '启动失败')
        }
      } catch (e) {
        Message.error({ content: e.message || '发布启动失败' })
      }
    }

    onMounted(loadApps)

    return {
      loading, apps, searchQuery, envFilter, viewMode,
      runningCount, healthyCount, failedCount, filteredApps,
      loadApps, getLanguageLabel, getLanguageClass, getAppStatus, getAppStatusText,
      getAppStatusClass, getBuildStatusText, formatGitRepo, formatTime,
      viewApp, editApp, viewReleases, runApp
    }
  }
}
</script>

<style scoped>
/* ===== 页面布局 ===== */
.app-center {
  padding: 24px 32px;
  max-width: 1440px;
  margin: 0 auto;
}

/* ===== Banner - 深色渐变风格（大厂统一视觉） ===== */
.page-banner {
  background: linear-gradient(135deg, #1d2129 0%, #2d3748 100%);
  border-radius: 12px;
  padding: 24px 32px;
  margin-bottom: 24px;
  color: #fff;
}

.banner-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.banner-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.banner-icon svg {
  width: 40px;
  height: 40px;
}

.banner-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
}

.banner-desc {
  font-size: 14px;
  color: #a0aec0;
  margin: 4px 0 0;
}

.banner-actions {
  display: flex;
  gap: 12px;
}

.btn-action {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-action:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

.btn-action:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }
.btn-action svg { width: 16px; height: 16px; }

.btn-action.primary {
  background: linear-gradient(135deg, #165DFF, #4e7cf6);
  border-color: transparent;
  box-shadow: 0 2px 8px rgba(22, 93, 255, 0.4);
}

.btn-action.primary:hover {
  box-shadow: 0 4px 12px rgba(22, 93, 255, 0.5);
}

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 统计行 ===== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-item {
  background: #fff;
  border-radius: 10px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  border: 1px solid #f0f0f3;
  transition: all 0.2s;
}

.stat-item:hover { box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06); }

.stat-number {
  font-size: 28px;
  font-weight: 700;
  color: #1d2129;
  line-height: 1.2;
}

.stat-item.running .stat-number { color: #3491fa; }
.stat-item.success .stat-number { color: #00b42a; }
.stat-item.warning .stat-number { color: #f53f3f; }

.stat-label {
  font-size: 12px;
  color: #86909c;
  margin-top: 4px;
}

/* ===== 工具栏 ===== */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 380px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #c9cdd4;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #4e7cf6;
  box-shadow: 0 0 0 2px rgba(78, 124, 246, 0.1);
}

.toolbar-right { display: flex; gap: 8px; align-items: center; }

.filter-select {
  padding: 7px 12px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  font-size: 13px;
  background: #fff;
  cursor: pointer;
}

.view-switch { display: flex; border: 1px solid #e5e6eb; border-radius: 6px; overflow: hidden; }

.view-btn {
  padding: 6px 10px;
  border: none;
  background: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}

.view-btn svg { width: 16px; height: 16px; color: #86909c; }
.view-btn.active { background: #f2f3f5; }
.view-btn.active svg { color: #4e7cf6; }
.view-btn + .view-btn { border-left: 1px solid #e5e6eb; }

/* ===== 卡片网格 ===== */
.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.app-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #f0f0f3;
  overflow: hidden;
  transition: all 0.25s;
}

.app-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.app-card.status-running { border-left: 3px solid #3491fa; }
.app-card.status-healthy { border-left: 3px solid #00b42a; }
.app-card.status-error { border-left: 3px solid #f53f3f; }
.app-card.status-idle { border-left: 3px solid #c9cdd4; }

/* ===== 卡片头部 ===== */
.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 16px 12px;
}

.app-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.app-avatar.lang-java { background: linear-gradient(135deg, #f5a623, #e68f17); }
.app-avatar.lang-go { background: linear-gradient(135deg, #00add8, #0097c7); }
.app-avatar.lang-python { background: linear-gradient(135deg, #3776ab, #ffd43b); }
.app-avatar.lang-nodejs { background: linear-gradient(135deg, #68a063, #4caf50); }
.app-avatar.lang-unknown { background: linear-gradient(135deg, #86909c, #6b7785); }

.app-meta { flex: 1; min-width: 0; }

.app-name {
  font-size: 15px;
  font-weight: 600;
  color: #1d2129;
  margin: 0;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.app-name:hover { color: #4e7cf6; }

.app-lang {
  font-size: 11px;
  color: #86909c;
}

.app-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  flex-shrink: 0;
}

.app-status-badge.running { background: #e8f4ff; color: #3491fa; }
.app-status-badge.healthy { background: #e8ffea; color: #00b42a; }
.app-status-badge.error { background: #ffece8; color: #f53f3f; }
.app-status-badge.idle { background: #f7f8fa; color: #86909c; }

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.app-status-badge.running .status-dot { animation: pulse 1.5s infinite; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* ===== 卡片主体 ===== */
.card-body { padding: 0 16px 12px; }

.info-tier {
  padding: 8px 0;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.info-tier + .info-tier { border-top: 1px solid #f7f8fa; }

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-icon { width: 14px; height: 14px; color: #c9cdd4; flex-shrink: 0; }

.info-text {
  font-size: 12px;
  color: #4e5969;
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ns-tag {
  background: #f2f3f5;
  padding: 1px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 11px;
}

.last-build {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.build-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
}

.build-status.success, .build-status.SUCCESS { color: #00b42a; }
.build-status.failed, .build-status.FAILURE { color: #f53f3f; }
.build-status.running { color: #3491fa; }

.status-dot-sm {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}

.build-time { font-size: 11px; color: #c9cdd4; margin-left: auto; }

/* ===== 卡片操作栏 ===== */
.card-actions {
  display: flex;
  border-top: 1px solid #f7f8fa;
  padding: 0;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 10px 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 12px;
  color: #86909c;
  transition: all 0.2s;
}

.action-btn:hover { background: #f7f8fa; color: #4e5969; }
.action-btn.primary { color: #4e7cf6; font-weight: 500; }
.action-btn.primary:hover { background: #f0f5ff; }
.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.action-btn svg { width: 14px; height: 14px; }
.action-btn + .action-btn { border-left: 1px solid #f7f8fa; }

/* ===== 列表视图 ===== */
.app-list {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f3;
  overflow: hidden;
}

.list-header {
  display: grid;
  grid-template-columns: 2.5fr 1fr 1fr 0.8fr 1fr 100px;
  padding: 12px 16px;
  background: #f7f8fa;
  font-size: 12px;
  font-weight: 600;
  color: #86909c;
  border-bottom: 1px solid #f0f0f3;
}

.list-row {
  display: grid;
  grid-template-columns: 2.5fr 1fr 1fr 0.8fr 1fr 100px;
  padding: 12px 16px;
  align-items: center;
  border-bottom: 1px solid #f7f8fa;
  transition: background 0.15s;
}

.list-row:hover { background: #fafbfc; }
.list-row:last-child { border-bottom: none; }

.col-app { display: flex; align-items: center; gap: 10px; }

.app-avatar-sm {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.app-avatar-sm.lang-java { background: linear-gradient(135deg, #f5a623, #e68f17); }
.app-avatar-sm.lang-go { background: linear-gradient(135deg, #00add8, #0097c7); }
.app-avatar-sm.lang-python { background: linear-gradient(135deg, #3776ab, #ffd43b); }
.app-avatar-sm.lang-nodejs { background: linear-gradient(135deg, #68a063, #4caf50); }
.app-avatar-sm.lang-unknown { background: linear-gradient(135deg, #86909c, #6b7785); }

.app-info-col { min-width: 0; }

.app-name-text {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #1d2129;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.app-name-text:hover { color: #4e7cf6; }

.app-repo-text {
  display: block;
  font-size: 11px;
  color: #c9cdd4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.env-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f2f3f5;
  border-radius: 4px;
  font-size: 11px;
  font-family: 'SF Mono', Monaco, monospace;
  color: #4e5969;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.status-pill.running { background: #e8f4ff; color: #3491fa; }
.status-pill.healthy { background: #e8ffea; color: #00b42a; }
.status-pill.error { background: #ffece8; color: #f53f3f; }
.status-pill.idle { background: #f7f8fa; color: #86909c; }

.col-branch code {
  font-size: 11px;
  background: #f7f8fa;
  padding: 2px 6px;
  border-radius: 3px;
  color: #4e5969;
}

.time-text { font-size: 12px; color: #86909c; }
.text-muted { font-size: 12px; color: #c9cdd4; }

.col-actions { display: flex; gap: 4px; }

.row-action-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.row-action-btn svg { width: 14px; height: 14px; color: #86909c; }
.row-action-btn:hover { background: #f2f3f5; }
.row-action-btn:hover svg { color: #4e5969; }
.row-action-btn.primary svg { color: #4e7cf6; }
.row-action-btn.primary:hover { background: #f0f5ff; }
.row-action-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ===== 加载/空状态 ===== */
.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.loading-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #f2f3f5;
  border-top-color: #4e7cf6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.loading-state p, .empty-state p { color: #86909c; font-size: 13px; margin-top: 12px; }
.empty-state h3 { font-size: 16px; color: #4e5969; margin: 16px 0 4px; }

.empty-icon svg { width: 64px; height: 64px; color: #c9cdd4; }

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .app-center { padding: 16px; }
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .app-grid { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrapper { max-width: 100%; }
}
</style>

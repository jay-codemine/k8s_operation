<template>
  <div class="gitops-releases-page">
    <!-- Hero Banner - 大厂深色渐变 -->
    <div class="hero-banner">
      <div class="hero-bg-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
      </div>
      <div class="hero-content">
        <div class="hero-left">
          <div class="hero-icon-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <polyline points="16 3 21 3 21 8"/>
              <line x1="4" y1="20" x2="21" y2="3"/>
              <polyline points="21 16 21 21 16 21"/>
              <line x1="15" y1="15" x2="21" y2="21"/>
              <line x1="4" y1="4" x2="9" y2="9"/>
            </svg>
          </div>
          <div class="hero-text">
            <h1>GitOps 发布中心</h1>
            <p>ArgoCD + Argo Workflows · Pull-based 声明式部署 · 发布历史追溯</p>
          </div>
        </div>
        <div class="hero-right">
          <button class="btn-hero-publish" @click="openPublishDialog">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            发布应用
          </button>
          <button class="btn-hero-secondary" @click="$router.push('/cicd/gitops/create')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            新建 GitOps
          </button>
        </div>
      </div>
    </div>

    <!-- 指标卡片 - 企业级 KPI -->
    <div class="kpi-row">
      <div class="kpi-card" :class="{ active: statusFilter === '' }" @click="statusFilter = ''">
        <div class="kpi-icon-wrap total">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
        </div>
        <div class="kpi-body">
          <span class="kpi-value">{{ stats.total }}</span>
          <span class="kpi-label">全部发布</span>
        </div>
      </div>
      <div class="kpi-card" :class="{ active: statusFilter === 'running' }" @click="statusFilter = 'running'">
        <div class="kpi-icon-wrap running">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div class="kpi-body">
          <span class="kpi-value">{{ stats.running }}</span>
          <span class="kpi-label">进行中</span>
        </div>
      </div>
      <div class="kpi-card" :class="{ active: statusFilter === 'Synced' }" @click="statusFilter = 'Synced'">
        <div class="kpi-icon-wrap synced">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        </div>
        <div class="kpi-body">
          <span class="kpi-value">{{ stats.synced }}</span>
          <span class="kpi-label">已同步</span>
        </div>
      </div>
      <div class="kpi-card" :class="{ active: statusFilter === 'Failed' }" @click="statusFilter = 'Failed'">
        <div class="kpi-icon-wrap failed">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
        </div>
        <div class="kpi-body">
          <span class="kpi-value">{{ stats.failed }}</span>
          <span class="kpi-label">失败</span>
        </div>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input v-model="keyword" @input="debounceSearch" placeholder="搜索应用名、ArgoCD App、Git 仓库..." class="search-input" />
      </div>
      <div class="action-right">
        <button class="btn-action has-label" :class="{ active: showLeaderView }" @click="showLeaderView = !showLeaderView" title="领导视图">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><path d="M20 8v6"/><path d="M23 11h-6"/></svg>
          <span>领导</span>
        </button>
        <button class="btn-action has-label" :class="{ active: showAdvSearch }" @click="showAdvSearch = !showAdvSearch" title="高级搜索">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><path d="M11 8v6M8 11h6"/></svg>
          <span>搜索</span>
        </button>
        <button class="btn-action" :class="{ active: viewMode === 'timeline' }" @click="viewMode = 'timeline'" title="时间线">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </button>
        <button class="btn-action" :class="{ active: viewMode === 'card' }" @click="viewMode = 'card'" title="卡片">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 3h8v8H3V3zm0 10h8v8H3v-8zm10-10h8v8h-8V3zm0 10h8v8h-8v-8z"/></svg>
        </button>
        <button class="btn-action" @click="loadData" :disabled="loading" title="刷新">
          <svg :class="{ spin: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        </button>
      </div>
    </div>

    <!-- 领导视图面板 -->
    <Transition name="slide">
      <div v-if="showLeaderView" class="leader-panel">
        <div class="leader-grid">
          <div class="leader-card primary">
            <div class="leader-card-icon synced">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            </div>
            <div class="leader-card-body">
              <span class="leader-value">{{ leaderStats.success_rate }}%</span>
              <span class="leader-label">成功率</span>
            </div>
            <div class="leader-bar"><div class="leader-bar-fill" :style="{ width: leaderStats.success_rate + '%' }"></div></div>
          </div>
          <div class="leader-card">
            <div class="leader-card-icon today">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
            </div>
            <div class="leader-card-body">
              <span class="leader-value">{{ leaderStats.today_count }}</span>
              <span class="leader-label">今日发布</span>
            </div>
          </div>
          <div class="leader-card">
            <div class="leader-card-icon apps">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>
            </div>
            <div class="leader-card-body">
              <span class="leader-value">{{ leaderStats.active_apps }}</span>
              <span class="leader-label">活跃应用</span>
            </div>
          </div>
          <div class="leader-card warn">
            <div class="leader-card-icon pending">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            </div>
            <div class="leader-card-body">
              <span class="leader-value">{{ leaderStats.pending_sync }}</span>
              <span class="leader-label">待同步</span>
            </div>
          </div>
        </div>
        <div class="leader-meta-row">
          <div class="leader-meta-item">
            <span class="meta-dot synced"></span>
            已同步 <strong>{{ leaderStats.synced }}</strong> 次
          </div>
          <div class="leader-meta-item">
            <span class="meta-dot failed"></span>
            失败 <strong>{{ leaderStats.failed }}</strong> 次
          </div>
          <div class="leader-meta-item">
            <span class="meta-dot running"></span>
            进行中 <strong>{{ leaderStats.running }}</strong> 次
          </div>
          <div class="leader-meta-item">
            <span class="meta-label">平均同步</span>
            <strong>{{ leaderStats.avg_sync_sec }}s</strong>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 高级搜索面板 -->
    <Transition name="slide">
      <div v-if="showAdvSearch" class="adv-search-panel">
        <div class="adv-search-row">
          <div class="adv-field">
            <label>应用名</label>
            <input v-model="advSearch.app_name" placeholder="输入应用名..." @keyup.enter="applyAdvSearch" />
          </div>
          <div class="adv-field">
            <label>同步状态</label>
            <select v-model="advSearch.sync_status" @change="applyAdvSearch">
              <option value="">全部</option>
              <option value="Synced">已同步</option>
              <option value="OutOfSync">漂移</option>
              <option value="Unknown">未知</option>
            </select>
          </div>
          <div class="adv-field">
            <label>发布状态</label>
            <select v-model="advSearch.status" @change="applyAdvSearch">
              <option value="">全部</option>
              <option value="Succeeded">成功</option>
              <option value="Failed">失败</option>
              <option value="Running">运行中</option>
            </select>
          </div>
          <div class="adv-field">
            <label>环境</label>
            <input v-model="advSearch.env" placeholder="命名空间..." @keyup.enter="applyAdvSearch" />
          </div>
        </div>
        <div class="adv-search-row">
          <div class="adv-field">
            <label>开始日期</label>
            <input type="date" v-model="advSearch.date_from" @change="applyAdvSearch" />
          </div>
          <div class="adv-field">
            <label>结束日期</label>
            <input type="date" v-model="advSearch.date_to" @change="applyAdvSearch" />
          </div>
          <div class="adv-field actions">
            <button class="btn-adv-search" @click="applyAdvSearch">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              搜索
            </button>
            <button class="btn-adv-reset" @click="resetAdvSearch">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              重置
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 加载中 -->
    <div v-if="loading && releases.length === 0" class="skeleton-grid">
      <div v-for="i in 3" :key="i" class="skeleton-card">
        <div class="skeleton-line w60"></div>
        <div class="skeleton-line w80"></div>
        <div class="skeleton-line w40"></div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="filteredReleases.length === 0 && !loading" class="empty-hero">
      <div class="empty-visual">
        <svg viewBox="0 0 200 120" fill="none" class="empty-svg">
          <rect x="40" y="20" width="120" height="80" rx="12" stroke="#cbd5e1" stroke-width="2" fill="#f8fafc"/>
          <circle cx="80" cy="60" r="18" stroke="#94a3b8" stroke-width="1.5" fill="#fff"/>
          <circle cx="120" cy="60" r="18" stroke="#94a3b8" stroke-width="1.5" fill="#fff"/>
          <line x1="62" y1="55" x2="98" y2="55" stroke="#cbd5e1" stroke-width="1.5" stroke-dasharray="4"/>
          <line x1="102" y1="55" x2="138" y2="55" stroke="#cbd5e1" stroke-width="1.5" stroke-dasharray="4"/>
          <circle cx="100" cy="32" r="8" stroke="#94a3b8" stroke-width="1.5" fill="#e2e8f0"/>
          <path d="M96 32l3 3 5-5" stroke="#94a3b8" stroke-width="1.5"/>
        </svg>
      </div>
      <h3>暂无 GitOps 发布记录</h3>
      <p>创建 GitOps 流水线并点击「发布应用」开始</p>
      <div class="empty-actions">
        <button class="btn-hero-publish" @click="$router.push('/cicd/gitops/create')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新建 GitOps 流水线
        </button>
      </div>
    </div>

    <!-- 时间线视图 -->
    <div v-else-if="viewMode === 'timeline'" class="timeline-view">
      <div class="timeline-track"></div>
      <div v-for="(rel, idx) in filteredReleases" :key="rel.id" class="timeline-item" :class="`sync-${normalizeSync(rel.sync_status)}`">
        <div class="timeline-dot">
          <div class="dot-inner"></div>
        </div>
        <div class="timeline-card">
          <div class="tl-header">
            <div class="tl-app">
              <span class="tl-app-avatar">{{ (rel.app_name || '?')[0].toUpperCase() }}</span>
              <div>
                <span class="tl-app-name">{{ rel.app_name || '-' }}</span>
                <span class="tl-app-id">#{{ rel.id }}</span>
              </div>
            </div>
            <div class="tl-status-row">
              <span :class="['sync-badge', `sync-${normalizeSync(rel.sync_status)}`]">
                {{ syncLabel(rel.sync_status) }}
              </span>
              <span :class="['status-dot-sm', rel.status === 'Succeeded' ? 'ok' : rel.status === 'Failed' ? 'err' : 'warn']"></span>
              <span class="tl-time">{{ formatTime(rel.created_at) }}</span>
            </div>
          </div>
          <div class="tl-body">
            <div class="tl-meta">
              <div class="tl-meta-item">
                <span class="meta-label">ArgoCD</span>
                <code class="meta-value">{{ rel.argo_app || '-' }}</code>
              </div>
              <div class="tl-meta-item">
                <span class="meta-label">Workflow</span>
                <code class="meta-value">{{ rel.workflow || '-' }}</code>
              </div>
              <div class="tl-meta-item">
                <span class="meta-label">镜像</span>
                <code class="meta-value img">{{ formatImage(rel) }}</code>
              </div>
            </div>
          </div>
          <div class="tl-footer">
            <button class="tl-btn" @click="viewDetail(rel)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              详情
            </button>
            <button v-if="rel.sync_status !== 'Synced' && rel.sync_status !== 'synced'" class="tl-btn primary" @click="triggerSync(rel)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              同步
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="card-grid">
      <div v-for="rel in filteredReleases" :key="rel.id" class="release-card" :class="`sync-${normalizeSync(rel.sync_status)}`">
        <div class="card-top">
          <div class="card-app">
            <div class="card-avatar" :class="`sync-${normalizeSync(rel.sync_status)}`">
              {{ (rel.app_name || '?')[0].toUpperCase() }}
            </div>
            <div class="card-app-info">
              <span class="card-app-name">{{ rel.app_name || '-' }}</span>
              <span class="card-pipeline-name" v-if="rel.pipeline_name">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/></svg>
                {{ rel.pipeline_name }}
              </span>
            </div>
          </div>
          <span :class="['sync-pill', `sync-${normalizeSync(rel.sync_status)}`]">
            <span class="sync-dot"></span>
            {{ syncLabel(rel.sync_status) }}
          </span>
        </div>

        <div class="card-body">
          <div class="card-info-row">
            <div class="card-info-item">
              <span class="info-label">ArgoCD App</span>
              <code class="info-code">{{ rel.argo_app || '-' }}</code>
            </div>
            <div class="card-info-item">
              <span class="info-label">Workflow</span>
              <code class="info-code">{{ rel.workflow || '-' }}</code>
            </div>
          </div>
          <div class="card-info-row">
            <div class="card-info-item">
              <span class="info-label">镜像</span>
              <code class="info-code img">{{ formatImage(rel) }}</code>
            </div>
            <div class="card-info-item">
              <span class="info-label">Revision</span>
              <code class="info-code">{{ (rel.sync_revision || '-').substring(0, 7) }}</code>
            </div>
          </div>
        </div>

        <div class="card-footer">
          <span class="card-time">{{ formatTime(rel.created_at) }}</span>
          <div class="card-actions">
            <button class="card-btn" @click="viewDetail(rel)">详情</button>
            <button v-if="rel.sync_status !== 'Synced'" class="card-btn sync" @click="triggerSync(rel)">同步</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="pagination-bar">
      <span class="page-info">共 {{ total }} 条</span>
      <div class="page-btns">
        <button :disabled="page === 1" @click="page--">‹</button>
        <button v-for="p in pages" :key="p" :class="{ active: p === page }" @click="page = p">{{ p }}</button>
        <button :disabled="page >= totalPages" @click="page++">›</button>
      </div>
    </div>

    <!-- 快速发布弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showPublish" class="modal-overlay" @click.self="showPublish = false">
          <div class="modal-dialog publish-dialog">
            <div class="modal-head">
              <div class="modal-head-icon gitops">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/></svg>
              </div>
              <div>
                <h3>发布 GitOps 应用</h3>
                <p>选择一个 GitOps 流水线，填写版本号即刻触发构建与同步</p>
              </div>
              <button class="modal-close" @click="showPublish = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-body">
              <div class="field">
                <label>选择 GitOps 流水线 <span class="required">*</span></label>
                <select v-model="publishForm.pipeline_id">
                  <option value="">请选择要发布的应用</option>
                  <option v-for="p in gitopsPipelines" :key="p.id" :value="p.id">
                    {{ p.name }} — {{ p.argo_app || '未配置 ArgoCD' }}
                  </option>
                </select>
              </div>
              <div v-if="selectedPipeline" class="pipeline-preview">
                <div class="preview-item">
                  <span class="preview-label">ArgoCD App</span>
                  <code>{{ selectedPipeline.argo_app || '未配置' }}</code>
                </div>
                <div class="preview-item">
                  <span class="preview-label">Git Repo</span>
                  <code>{{ selectedPipeline.git_repo || '-' }}</code>
                </div>
                <div class="preview-item">
                  <span class="preview-label">分支</span>
                  <code>{{ selectedPipeline.git_branch || 'main' }}</code>
                </div>
              </div>
              <div class="field">
                <label>版本号 / 镜像标签 <span class="required">*</span></label>
                <input v-model="publishForm.version" placeholder="例如: v1.2.3 或 latest" />
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="showPublish = false">取消</button>
              <button class="btn-confirm gitops" @click="doPublish" :disabled="publishing || !publishForm.pipeline_id">
                <span v-if="publishing" class="spinner-sm"></span>
                <svg v-else viewBox="0 0 24 24" fill="currentColor" width="16" height="16"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                {{ publishing ? '发布中...' : '立即发布' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { getReleases, getGitOpsReleaseStats, getGitOpsReleaseSearch, triggerGitOpsSync } from '@/api/cicd.js'
import { getPipelines, runPipeline } from '@/api/platform/pipeline.js'

const router = useRouter()

// 状态
const loading = ref(false)
const releases = ref([])
const gitopsPipelines = ref([])
const keyword = ref('')
const statusFilter = ref('')
const viewMode = ref('timeline')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showPublish = ref(false)
const publishing = ref(false)
const showLeaderView = ref(false)
const showAdvSearch = ref(false)

// 统计数据
const stats = reactive({ total: 0, running: 0, synced: 0, failed: 0 })
// 领导视图统计
const leaderStats = reactive({ total: 0, synced: 0, failed: 0, running: 0, pending_sync: 0, today_count: 0, avg_sync_sec: 0, active_apps: 0, success_rate: 0 })
// 高级搜索
const advSearch = reactive({ app_name: '', sync_status: '', status: '', env: '', date_from: '', date_to: '' })

// 发布表单
const publishForm = reactive({ pipeline_id: '', version: '' })

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
const pages = computed(() => {
  const p = []
  for (let i = 1; i <= Math.min(totalPages.value, 7); i++) p.push(i)
  return p
})

const filteredReleases = computed(() => {
  let result = releases.value
  if (statusFilter.value) {
    if (statusFilter.value === 'running') result = result.filter(r => ['Running', 'Queued'].includes(r.status))
    else if (statusFilter.value === 'Synced') result = result.filter(r => r.sync_status === 'Synced')
    else if (statusFilter.value === 'Failed') result = result.filter(r => r.status === 'Failed')
    else result = result.filter(r => r.status === statusFilter.value)
  }
  return result
})

const selectedPipeline = computed(() => {
  if (!publishForm.pipeline_id) return null
  return gitopsPipelines.value.find(p => p.id === Number(publishForm.pipeline_id))
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    // 加载 GitOps 流水线列表
    const pipeRes = await getPipelines({ page: 1, page_size: 200 })
    if (pipeRes.code === 0) {
      gitopsPipelines.value = (pipeRes.data?.list || []).filter(p => p.deploy_mode === 'gitops')
    }

    // 加载发布记录 - 使用增强搜索 API
    const searchParams = {
      page: page.value, page_size: pageSize.value,
      keyword: keyword.value || undefined,
      app_name: advSearch.app_name || undefined,
      sync_status: advSearch.sync_status || undefined,
      status: advSearch.status || undefined,
      env: advSearch.env || undefined,
      date_from: advSearch.date_from || undefined,
      date_to: advSearch.date_to || undefined,
    }
    // 移除空值
    Object.keys(searchParams).forEach(k => { if (!searchParams[k]) delete searchParams[k] })

    const relRes = await getGitOpsReleaseSearch(searchParams)
    if (relRes.code === 0) {
      releases.value = relRes.data?.list || []
      total.value = relRes.data?.total || 0

      // 计算统计
      stats.total = total.value
      stats.synced = releases.value.filter(r => r.sync_status === 'Synced').length
      stats.failed = releases.value.filter(r => r.status === 'Failed').length
      stats.running = releases.value.filter(r => ['Running', 'Queued'].includes(r.status)).length
    }

    // 加载领导视图统计（独立 API）
    try {
      const statsRes = await getGitOpsReleaseStats()
      if (statsRes.code === 0 && statsRes.data?.stats) {
        Object.assign(leaderStats, statsRes.data.stats)
      }
    } catch (_) { /* 非关键 */ }
  } catch (e) {
    console.error('加载失败', e)
  } finally {
    loading.value = false
  }
}

// 发布
const openPublishDialog = () => { showPublish.value = true; publishForm.pipeline_id = ''; publishForm.version = '' }
const doPublish = async () => {
  if (!publishForm.pipeline_id) { Message.warning('请选择流水线'); return }
  publishing.value = true
  try {
    const res = await runPipeline(Number(publishForm.pipeline_id))
    if (res.code === 0) {
      Message.success('发布已启动，正在构建中...')
      showPublish.value = false
      const pipeline = selectedPipeline.value
      if (pipeline) router.push(`/cicd/pipelines/${pipeline.id}?tab=stages`)
    } else {
      throw new Error(res.msg || '发布失败')
    }
  } catch (e) {
    Message.error(e.message || '发布失败')
  } finally {
    publishing.value = false
  }
}

// 触发同步
const triggerSync = async (rel) => {
  if (!rel.pipeline_id) return
  try {
    const res = await triggerGitOpsSync(rel.pipeline_id)
    if (res.code === 0) {
      Message.success('ArgoCD 同步已触发')
      loadData()
    } else {
      Message.error(res.msg || '同步失败')
    }
  } catch (e) {
    Message.error('同步失败')
  }
}

const viewDetail = (rel) => {
  if (rel.pipeline_id) router.push(`/cicd/pipelines/${rel.pipeline_id}?tab=stages`)
}

const debounceSearch = () => {
  clearTimeout(window._gitopsSearchTimer)
  window._gitopsSearchTimer = setTimeout(() => loadData(), 300)
}
const applyAdvSearch = () => { page.value = 1; loadData() }
const resetAdvSearch = () => {
  Object.assign(advSearch, { app_name: '', sync_status: '', status: '', env: '', date_from: '', date_to: '' })
  page.value = 1; loadData()
}

// 格式化
const normalizeSync = (s) => (s || 'Unknown').toLowerCase()
const syncLabel = (s) => {
  const map = { Synced: '已同步', synced: '已同步', OutOfSync: '漂移', Unknown: '未知', '' : '未同步' }
  return map[s] || s || '未同步'
}
const formatImage = (rel) => {
  if (!rel.image_repo) return '-'
  const repo = rel.image_repo.replace(/^https?:\/\//, '').split('/').slice(-2).join('/')
  return `${repo}:${rel.image_tag || 'latest'}`
}
const formatTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const now = new Date()
  const diff = now - d
  if (diff < 3600000) return `${Math.floor(diff/60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff/3600000)}小时前`
  return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

watch([keyword, statusFilter], () => { page.value = 1; loadData() })
watch(showLeaderView, async (val) => { if (val) { try { const r = await getGitOpsReleaseStats(); if (r.code === 0 && r.data?.stats) Object.assign(leaderStats, r.data.stats) } catch (_) {} } })
watch(page, () => loadData())

onMounted(loadData)
</script>

<style scoped>
.gitops-releases-page { min-height: 100vh; padding: 0 24px 24px; }

/* ===== Hero Banner ===== */
.hero-banner {
  position: relative;
  background: linear-gradient(135deg, #0d9488 0%, #0891b2 50%, #0f766e 100%);
  border-radius: 20px;
  padding: 40px 44px;
  margin-bottom: 24px;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(13, 148, 136, 0.25);
}
.hero-bg-shapes { position: absolute; inset: 0; pointer-events: none; }
.shape {
  position: absolute; border-radius: 50%; background: rgba(255,255,255,0.06);
}
.shape-1 { width: 300px; height: 300px; top: -120px; right: -60px; }
.shape-2 { width: 200px; height: 200px; bottom: -80px; left: 10%; }
.shape-3 { width: 120px; height: 120px; top: 30%; right: 30%; }
.hero-content { position: relative; display: flex; justify-content: space-between; align-items: center; z-index: 1; }
.hero-left { display: flex; align-items: center; gap: 18px; }
.hero-icon-wrap {
  width: 56px; height: 56px; border-radius: 16px;
  background: rgba(255,255,255,0.18); backdrop-filter: blur(10px);
  display: flex; align-items: center; justify-content: center;
  color: #fff;
}
.hero-icon-wrap svg { width: 28px; height: 28px; }
.hero-text h1 { font-size: 26px; font-weight: 700; color: #fff; margin: 0; letter-spacing: -0.5px; }
.hero-text p { font-size: 14px; color: rgba(255,255,255,0.8); margin: 6px 0 0; }
.hero-right { display: flex; gap: 10px; }
.btn-hero-publish {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 24px; background: #fff; color: #0d9488;
  border: none; border-radius: 12px; font-size: 15px; font-weight: 600; cursor: pointer;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12); transition: all 0.2s;
}
.btn-hero-publish:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.18); }
.btn-hero-secondary {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 24px; background: rgba(255,255,255,0.15);
  color: #fff; border: 1px solid rgba(255,255,255,0.25);
  border-radius: 12px; font-size: 15px; font-weight: 600; cursor: pointer;
  transition: all 0.2s;
}
.btn-hero-secondary:hover { background: rgba(255,255,255,0.25); }

/* ===== KPI 卡片 ===== */
.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px; }
.kpi-card {
  background: #fff; border-radius: 14px; padding: 20px 24px;
  display: flex; align-items: center; gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); cursor: pointer;
  border: 2px solid transparent; transition: all 0.2s;
}
.kpi-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.kpi-card.active { border-color: #0d9488; box-shadow: 0 0 0 3px rgba(13,148,136,0.1); }
.kpi-icon-wrap { width: 46px; height: 46px; border-radius: 12px; display: flex; align-items: center; justify-content: center; }
.kpi-icon-wrap svg { width: 22px; height: 22px; }
.kpi-icon-wrap.total { background: #f0fdfa; color: #0d9488; }
.kpi-icon-wrap.running { background: #fef3c7; color: #d97706; }
.kpi-icon-wrap.synced { background: #d1fae5; color: #059669; }
.kpi-icon-wrap.failed { background: #fee2e2; color: #dc2626; }
.kpi-body { display: flex; flex-direction: column; }
.kpi-value { font-size: 26px; font-weight: 700; color: #1f2937; }
.kpi-label { font-size: 13px; color: #6b7280; margin-top: 2px; }

/* ===== 操作栏 ===== */
.action-bar { display: flex; justify-content: space-between; align-items: center; gap: 16px; margin-bottom: 20px; }
.search-box {
  flex: 1; max-width: 420px; position: relative;
}
.search-box svg { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); width: 18px; height: 18px; color: #94a3b8; }
.search-input {
  width: 100%; padding: 12px 16px 12px 44px; border: 1px solid #e2e8f0; border-radius: 10px;
  font-size: 14px; background: #fff; outline: none;
}
.search-input:focus { border-color: #0d9488; box-shadow: 0 0 0 3px rgba(13,148,136,0.1); }
.action-right { display: flex; gap: 8px; }
.btn-action {
  width: 38px; height: 38px; display: flex; align-items: center; justify-content: center;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer;
  color: #64748b; transition: all 0.15s;
}
.btn-action:hover, .btn-action.active { border-color: #0d9488; color: #0d9488; }
.btn-action svg { width: 18px; height: 18px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ===== 时间线 ===== */
.timeline-view { position: relative; padding-left: 40px; }
.timeline-track {
  position: absolute; left: 19px; top: 0; bottom: 0; width: 2px;
  background: linear-gradient(180deg, #0d9488 0%, #e2e8f0 100%);
}
.timeline-item { position: relative; margin-bottom: 24px; }
.timeline-dot {
  position: absolute; left: -40px; top: 20px;
  width: 40px; height: 40px; display: flex; align-items: center; justify-content: center;
}
.dot-inner {
  width: 14px; height: 14px; border-radius: 50%; background: #0d9488;
  border: 3px solid #ccfbf1; box-shadow: 0 0 0 4px rgba(13,148,136,0.15);
}
.timeline-item.sync-failed .dot-inner { background: #ef4444; border-color: #fee2e2; box-shadow: 0 0 0 4px rgba(239,68,68,0.15); }
.timeline-item.sync-unknown .dot-inner { background: #94a3b8; border-color: #f1f5f9; }
.timeline-card {
  background: #fff; border-radius: 14px; padding: 20px 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); border: 1px solid #f1f5f9;
  transition: all 0.2s;
}
.timeline-card:hover { box-shadow: 0 6px 20px rgba(0,0,0,0.08); }
.tl-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.tl-app { display: flex; align-items: center; gap: 12px; }
.tl-app-avatar {
  width: 38px; height: 38px; border-radius: 10px; background: linear-gradient(135deg, #0d9488, #0891b2);
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 16px; font-weight: 700;
}
.tl-app-name { font-size: 15px; font-weight: 600; color: #1f2937; }
.tl-app-id { font-size: 12px; color: #94a3b8; margin-left: 6px; }
.tl-status-row { display: flex; align-items: center; gap: 10px; }
.tl-time { font-size: 12px; color: #94a3b8; }
.tl-body { margin-bottom: 14px; }
.tl-meta { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
.tl-meta-item { display: flex; flex-direction: column; gap: 4px; }
.meta-label { font-size: 11px; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; }
.meta-value { font-size: 13px; color: #374151; background: #f8fafc; padding: 4px 8px; border-radius: 4px; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.meta-value.img { font-family: 'SF Mono', monospace; font-size: 12px; }
.tl-footer { display: flex; gap: 8px; border-top: 1px solid #f1f5f9; padding-top: 12px; }
.tl-btn {
  display: flex; align-items: center; gap: 4px; padding: 6px 14px;
  border: 1px solid #e2e8f0; border-radius: 8px; background: #fff;
  font-size: 13px; color: #64748b; cursor: pointer; transition: all 0.15s;
}
.tl-btn:hover { border-color: #0d9488; color: #0d9488; }
.tl-btn.primary { background: #f0fdfa; border-color: #99f6e4; color: #0d9488; }
.tl-btn.primary:hover { background: #0d9488; color: #fff; }
.tl-btn svg { width: 14px; height: 14px; }

/* ===== 卡片视图 ===== */
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 16px; }
.release-card {
  background: #fff; border-radius: 14px; overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); border: 1px solid #f1f5f9;
  transition: all 0.25s;
}
.release-card:hover { transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.1); }
.release-card.sync-synced { border-left: 3px solid #059669; }
.release-card.sync-failed { border-left: 3px solid #ef4444; }
.card-top { display: flex; justify-content: space-between; align-items: center; padding: 18px 20px 12px; }
.card-app { display: flex; align-items: center; gap: 12px; }
.card-avatar {
  width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center;
  font-size: 16px; font-weight: 700; color: #fff;
}
.card-avatar.sync-synced { background: linear-gradient(135deg, #059669, #10b981); }
.card-avatar.sync-failed { background: linear-gradient(135deg, #dc2626, #ef4444); }
.card-avatar.sync-unknown { background: linear-gradient(135deg, #0d9488, #0891b2); }
.card-app-name { font-size: 15px; font-weight: 600; color: #1f2937; }
.card-pipeline-name { display: flex; align-items: center; gap: 4px; font-size: 11px; color: #94a3b8; margin-top: 2px; }
.sync-pill {
  display: flex; align-items: center; gap: 6px; padding: 5px 12px; border-radius: 20px;
  font-size: 12px; font-weight: 600;
}
.sync-pill.sync-synced { background: #d1fae5; color: #059669; }
.sync-pill.sync-failed { background: #fee2e2; color: #dc2626; }
.sync-pill.sync-unknown { background: #f1f5f9; color: #64748b; }
.sync-dot { width: 7px; height: 7px; border-radius: 50%; }
.sync-synced .sync-dot { background: #059669; }
.sync-failed .sync-dot { background: #dc2626; }
.sync-unknown .sync-dot { background: #94a3b8; }

.card-body { padding: 0 20px 12px; }
.card-info-row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 8px; }
.card-info-item { display: flex; flex-direction: column; gap: 3px; }
.info-label { font-size: 11px; color: #94a3b8; text-transform: uppercase; }
.info-code { font-size: 12px; color: #374151; background: #f8fafc; padding: 3px 8px; border-radius: 4px; font-family: 'SF Mono', monospace; max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.info-code.img { font-size: 11px; }
.card-footer { display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; border-top: 1px solid #f1f5f9; }
.card-time { font-size: 12px; color: #94a3b8; }
.card-actions { display: flex; gap: 6px; }
.card-btn {
  padding: 5px 12px; border: 1px solid #e2e8f0; border-radius: 6px;
  background: #fff; font-size: 12px; color: #64748b; cursor: pointer;
}
.card-btn:hover { border-color: #0d9488; color: #0d9488; }
.card-btn.sync { background: #f0fdfa; border-color: #99f6e4; color: #0d9488; }
.card-btn.sync:hover { background: #0d9488; color: #fff; }

/* ===== 按钮增强 ===== */
.btn-action.has-label {
  width: auto; padding: 0 12px; gap: 6px; font-size: 12px; font-weight: 500;
}

/* ===== 领导视图面板 ===== */
.leader-panel {
  background: linear-gradient(135deg, #f8fafc 0%, #f0fdfa 100%);
  border: 1px solid #ccfbf1; border-radius: 16px;
  padding: 20px 24px; margin-bottom: 20px;
}
.leader-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.leader-card {
  background: #fff; border-radius: 12px; padding: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #f1f5f9;
  display: flex; align-items: center; gap: 14px; transition: all 0.2s;
}
.leader-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.leader-card.primary { border-color: #059669; background: linear-gradient(135deg, #f0fdfa, #ecfdf5); }
.leader-card.warn { border-color: #fbbf24; background: linear-gradient(135deg, #fffbeb, #fef3c7); }
.leader-card-icon {
  width: 42px; height: 42px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.leader-card-icon svg { width: 20px; height: 20px; }
.leader-card-icon.synced { background: #d1fae5; color: #059669; }
.leader-card-icon.today { background: #dbeafe; color: #2563eb; }
.leader-card-icon.apps { background: #ede9fe; color: #7c3aed; }
.leader-card-icon.pending { background: #fef3c7; color: #d97706; }
.leader-card-body { display: flex; flex-direction: column; min-width: 0; }
.leader-value { font-size: 22px; font-weight: 700; color: #1f2937; line-height: 1; }
.leader-label { font-size: 12px; color: #6b7280; margin-top: 3px; }
.leader-bar { flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; min-width: 40px; }
.leader-bar-fill { height: 100%; background: linear-gradient(90deg, #059669, #10b981); border-radius: 3px; transition: width 0.6s ease; }
.leader-meta-row { display: flex; gap: 24px; padding-top: 4px; }
.leader-meta-item { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #6b7280; }
.leader-meta-item strong { color: #1f2937; }
.meta-dot { width: 8px; height: 8px; border-radius: 50%; }
.meta-dot.synced { background: #059669; }
.meta-dot.failed { background: #dc2626; }
.meta-dot.running { background: #d97706; }
.meta-label { font-size: 12px; color: #94a3b8; margin-right: 4px; }

/* ===== 高级搜索面板 ===== */
.adv-search-panel {
  background: #fff; border: 1px solid #e2e8f0; border-radius: 14px;
  padding: 18px 20px; margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.adv-search-row { display: flex; gap: 12px; margin-bottom: 10px; }
.adv-search-row:last-child { margin-bottom: 0; }
.adv-field { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.adv-field label { font-size: 11px; font-weight: 600; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px; }
.adv-field input, .adv-field select {
  padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 8px;
  font-size: 13px; outline: none; background: #fff;
}
.adv-field input:focus, .adv-field select:focus { border-color: #0d9488; box-shadow: 0 0 0 2px rgba(13,148,136,0.1); }
.adv-field.actions { flex-direction: row; align-items: flex-end; gap: 8px; }
.btn-adv-search {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 18px; background: linear-gradient(135deg, #0d9488, #0f766e);
  color: #fff; border: none; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer;
}
.btn-adv-search:hover { box-shadow: 0 2px 8px rgba(13,148,136,0.3); }
.btn-adv-reset {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 18px; border: 1px solid #e2e8f0; border-radius: 8px;
  background: #fff; font-size: 13px; color: #64748b; cursor: pointer;
}
.btn-adv-reset:hover { border-color: #94a3b8; color: #374151; }

/* ===== 动画 ===== */
.slide-enter-active, .slide-leave-active { transition: all 0.25s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateY(-8px); }

/* ===== 空状态 ===== */
.empty-hero { text-align: center; padding: 80px 20px; }
.empty-visual { margin-bottom: 20px; }
.empty-svg { width: 200px; height: 120px; }
.empty-hero h3 { font-size: 18px; color: #374151; margin: 0 0 8px; }
.empty-hero p { font-size: 14px; color: #94a3b8; margin: 0 0 24px; }
.empty-actions { display: flex; justify-content: center; gap: 10px; }

/* ===== 骨架屏 ===== */
.skeleton-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.skeleton-card { background: #fff; border-radius: 14px; padding: 24px; display: flex; flex-direction: column; gap: 12px; }
.skeleton-line { height: 14px; background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%); background-size: 200% 100%; animation: shimmer 1.5s infinite; border-radius: 4px; }
.skeleton-line.w60 { width: 60%; }
.skeleton-line.w80 { width: 80%; }
.skeleton-line.w40 { width: 40%; }
@keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

/* ===== 分页 ===== */
.pagination-bar { display: flex; justify-content: flex-end; align-items: center; gap: 16px; margin-top: 20px; }
.page-info { font-size: 13px; color: #94a3b8; }
.page-btns { display: flex; gap: 4px; }
.page-btns button {
  width: 34px; height: 34px; border: 1px solid #e2e8f0; border-radius: 8px;
  background: #fff; font-size: 13px; color: #64748b; cursor: pointer;
}
.page-btns button:hover:not(:disabled) { border-color: #0d9488; color: #0d9488; }
.page-btns button.active { background: #0d9488; color: #fff; border-color: #0d9488; }
.page-btns button:disabled { opacity: 0.4; cursor: not-allowed; }

/* ===== 同步标识 ===== */
.sync-badge {
  display: inline-flex; align-items: center; gap: 4px; padding: 3px 10px;
  border-radius: 12px; font-size: 12px; font-weight: 600;
}
.sync-badge.sync-synced { background: #d1fae5; color: #059669; }
.sync-badge.sync-failed { background: #fee2e2; color: #dc2626; }
.sync-badge.sync-unknown { background: #f1f5f9; color: #64748b; }

.status-dot-sm { width: 7px; height: 7px; border-radius: 50%; }
.status-dot-sm.ok { background: #059669; }
.status-dot-sm.err { background: #dc2626; }
.status-dot-sm.warn { background: #d97706; }

/* ===== 弹窗 ===== */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 18px; width: 520px; box-shadow: 0 20px 60px rgba(0,0,0,0.2); }
.publish-dialog { width: 540px; }
.modal-head { display: flex; align-items: flex-start; gap: 14px; padding: 24px 28px 0; }
.modal-head-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; }
.modal-head-icon.gitops { background: linear-gradient(135deg, #ccfbf1, #99f6e4); color: #0d9488; }
.modal-head-icon svg { width: 22px; height: 22px; }
.modal-head h3 { font-size: 18px; font-weight: 700; color: #1f2937; margin: 0; }
.modal-head p { font-size: 13px; color: #6b7280; margin: 4px 0 0; }
.modal-close { margin-left: auto; background: none; border: none; cursor: pointer; color: #94a3b8; padding: 4px; }
.modal-close:hover { color: #374151; }
.modal-body { padding: 20px 28px; }
.field { margin-bottom: 16px; }
.field label { display: block; font-size: 13px; font-weight: 600; color: #374151; margin-bottom: 6px; }
.required { color: #ef4444; }
.field select, .field input {
  width: 100%; padding: 10px 14px; border: 1px solid #d1d5db; border-radius: 8px;
  font-size: 14px; outline: none; transition: border-color 0.15s;
}
.field select:focus, .field input:focus { border-color: #0d9488; box-shadow: 0 0 0 3px rgba(13,148,136,0.1); }
.pipeline-preview { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 14px; margin-bottom: 16px; display: flex; flex-direction: column; gap: 8px; }
.preview-item { display: flex; justify-content: space-between; align-items: center; font-size: 13px; }
.preview-label { color: #6b7280; }
.preview-item code { background: #f1f5f9; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.modal-foot { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 28px 24px; border-top: 1px solid #f1f5f9; }
.btn-cancel {
  padding: 10px 20px; border: 1px solid #e2e8f0; border-radius: 10px;
  background: #fff; font-size: 14px; color: #64748b; cursor: pointer;
}
.btn-confirm {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 24px; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 600; color: #fff; cursor: pointer;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}
.btn-confirm.gitops { background: linear-gradient(135deg, #0d9488, #0f766e); }
.btn-confirm:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(13,148,136,0.3); }
.btn-confirm:disabled { opacity: 0.5; cursor: not-allowed; }
.spinner-sm {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff;
  border-radius: 50%; animation: spin 0.6s linear infinite; display: inline-block;
}
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s, transform 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal-dialog, .modal-leave-to .modal-dialog { transform: scale(0.95); }
</style>

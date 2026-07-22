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
          <button class="btn-action" @click="refreshAll" :disabled="loading">
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

    <!-- 统计概览（实时发布看板） -->
    <div class="stats-row">
      <div class="stat-item">
        <span class="stat-number">{{ buildStats.today_builds || 0 }}</span>
        <span class="stat-label">今日构建</span>
      </div>
      <div class="stat-item success">
        <span class="stat-number">{{ successRateText }}</span>
        <span class="stat-label">构建成功率</span>
      </div>
      <div class="stat-item">
        <span class="stat-number">{{ avgDurationText }}</span>
        <span class="stat-label">平均耗时</span>
      </div>
      <div class="stat-item running">
        <span class="stat-number">{{ runningCount }}</span>
        <span class="stat-label">构建中</span>
      </div>
    </div>

    <!-- 环境筛选（彩色 tab，发布第一入口） -->
    <div class="env-bar">
      <span class="env-bar-label">环境</span>
      <div class="env-tabs">
        <button
          :class="['env-tab', { active: envFilter === '' }]"
          @click="envFilter = ''"
        >全部环境</button>
        <button
          v-for="env in envOptions"
          :key="env.key"
          :class="['env-tab', { active: envFilter === env.key }]"
          :style="envFilter === env.key ? { color: env.color, borderColor: env.color, background: env.color + '14' } : {}"
          @click="envFilter = env.key"
        >
          <span class="env-tab-dot" :style="{ background: env.color }"></span>
          {{ env.label }}
        </button>
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
        <button :class="['filter-toggle', { active: showFilter || activeFilterCount > 0 }]" @click="showFilter = !showFilter">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/></svg>
          <span>筛选</span>
          <span v-if="activeFilterCount > 0" class="filter-count">{{ activeFilterCount }}</span>
        </button>
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

    <!-- 高级筛选面板（应用类型 / 状态 / 发布人 / 时间） -->
    <transition name="filter-slide">
      <div v-if="showFilter" class="filter-panel">
        <div class="filter-group">
          <span class="filter-label">应用类型</span>
          <div class="filter-chips">
            <button :class="['chip', { active: langFilter === '' }]" @click="langFilter = ''">全部</button>
            <button
              v-for="opt in langOptions"
              :key="opt.key"
              :class="['chip', { active: langFilter === opt.key }]"
              @click="langFilter = opt.key"
            >{{ opt.label }}</button>
          </div>
        </div>
        <div class="filter-group">
          <span class="filter-label">构建状态</span>
          <div class="filter-chips">
            <button :class="['chip', { active: statusFilter === '' }]" @click="statusFilter = ''">全部</button>
            <button
              v-for="opt in statusOptions"
              :key="opt.key"
              :class="['chip', { active: statusFilter === opt.key }]"
              @click="statusFilter = opt.key"
            >{{ opt.label }}</button>
          </div>
        </div>
        <div class="filter-group">
          <span class="filter-label">发布人</span>
          <div class="filter-chips">
            <button :class="['chip', { active: userFilter === '' }]" @click="userFilter = ''">全部</button>
            <button
              v-for="u in userOptions"
              :key="u"
              :class="['chip', { active: userFilter === u }]"
              @click="userFilter = u"
            >{{ u }}</button>
          </div>
        </div>
        <div class="filter-group">
          <span class="filter-label">最近构建</span>
          <div class="filter-chips">
            <button
              v-for="opt in timeOptions"
              :key="opt.key"
              :class="['chip', { active: timeFilter === opt.key }]"
              @click="timeFilter = opt.key"
            >{{ opt.label }}</button>
          </div>
        </div>
        <div class="filter-actions">
          <button class="filter-reset" :disabled="activeFilterCount === 0" @click="resetFilters">重置</button>
        </div>
      </div>
    </transition>

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

          <!-- Tier 2.5: 版本信息（镜像标签 + 提交） -->
          <div class="info-tier version-tier" v-if="app.lastRunTag || app.lastCommit">
            <div class="info-item ver-item" v-if="app.lastRunTag" :title="app.lastRunImage" @click.stop="copyImage(app)">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12.89 1.45l8 4A2 2 0 0 1 22 7.24v9.53a2 2 0 0 1-1.11 1.79l-8 4a2 2 0 0 1-1.79 0l-8-4a2 2 0 0 1-1.1-1.8V7.24a2 2 0 0 1 1.11-1.79l8-4a2 2 0 0 1 1.78 0z"/><polyline points="2.32 6.16 12 11 21.68 6.16"/><line x1="12" y1="22.76" x2="12" y2="11"/></svg>
              <span class="ver-tag">{{ app.lastRunTag }}</span>
              <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </div>
            <div class="info-item" v-if="app.lastCommit" :title="app.lastCommitMsg">
              <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="4"/><line x1="1.05" y1="12" x2="7" y2="12"/><line x1="17.01" y1="12" x2="22.96" y2="12"/></svg>
              <code class="commit-hash">{{ shortCommit(app.lastCommit) }}</code>
            </div>
          </div>

          <!-- Tier 3: 最近构建/发布状态 -->
          <div class="info-tier build-tier" v-if="app.lastRunStatus">
            <div class="last-build">
              <span class="build-status" :class="app.lastRunStatus">
                <span class="status-dot-sm"></span>
                {{ getBuildStatusText(app.lastRunStatus) }}
              </span>
              <span class="build-user" v-if="app.lastTriggerUser" title="发布人">{{ app.lastTriggerUser }}</span>
              <span class="build-dur" v-if="app.lastDuration > 0">· {{ formatDuration(app.lastDuration) }}</span>
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
        <span class="col-version">版本</span>
        <span class="col-user">发布人</span>
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
          <span class="env-badge" :style="{ color: getEnvColor(resolveEnvKey(app)), background: getEnvColor(resolveEnvKey(app)) + '14' }">{{ getEnvLabel(resolveEnvKey(app)) }}</span>
        </div>
        <div class="col-status">
          <span class="status-pill" :class="getAppStatus(app)">
            <span class="status-dot"></span>
            {{ getAppStatusText(app) }}
          </span>
        </div>
        <div class="col-version">
          <span v-if="app.lastRunTag" class="ver-tag ver-tag-sm" :title="app.lastRunImage" @click.stop="copyImage(app)">{{ app.lastRunTag }}</span>
          <code v-if="app.lastCommit" class="commit-hash" :title="app.lastCommitMsg">{{ shortCommit(app.lastCommit) }}</code>
          <span v-if="!app.lastRunTag && !app.lastCommit" class="text-muted">-</span>
        </div>
        <div class="col-user">
          <span v-if="app.lastTriggerUser" class="user-text">{{ app.lastTriggerUser }}</span>
          <span v-else class="text-muted">-</span>
        </div>
        <div class="col-last">
          <span v-if="app.lastRunTime" class="time-text">{{ formatTime(app.lastRunTime) }}</span>
          <span v-else class="text-muted">-</span>
          <span v-if="app.lastDuration > 0" class="dur-text">{{ formatDuration(app.lastDuration) }}</span>
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

    <!-- 发布 Modal：目标环境 / 版本 / 策略 / 副本数 -->
    <transition name="modal-fade">
      <div v-if="releaseVisible" class="release-mask" @click.self="closeRelease">
        <div class="release-dialog">
          <div class="release-head">
            <div class="release-title">
              <span class="release-app-dot" :class="getLanguageClass(releaseApp.language)"></span>
              发布 · {{ releaseApp.name }}
            </div>
            <button class="release-close" @click="closeRelease">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="release-body">
            <div class="release-field">
              <label class="release-label">目标环境</label>
              <div class="release-chips">
                <button
                  v-for="env in envOptions"
                  :key="env.key"
                  :class="['release-chip', { active: releaseForm.deploy_env === env.key }]"
                  :style="releaseForm.deploy_env === env.key ? { color: env.color, borderColor: env.color, background: env.color + '14' } : {}"
                  @click="releaseForm.deploy_env = env.key"
                >
                  <span class="release-chip-dot" :style="{ background: env.color }"></span>
                  {{ env.label }}
                </button>
              </div>
            </div>
            <div class="release-field">
              <label class="release-label">版本 / 分支</label>
              <input v-model="releaseForm.branch" type="text" class="release-input" placeholder="如 main、release/v1.2.0" />
            </div>
            <div class="release-field">
              <label class="release-label">发布策略</label>
              <div class="release-strategies">
                <button
                  v-for="opt in strategyOptions"
                  :key="opt.key"
                  :class="['release-strategy', { active: releaseForm.strategy === opt.key }]"
                  @click="releaseForm.strategy = opt.key"
                >
                  <span class="release-strategy-name">{{ opt.label }}</span>
                  <span class="release-strategy-desc">{{ opt.desc }}</span>
                </button>
              </div>
            </div>
            <div class="release-field">
              <label class="release-label">副本数</label>
              <div class="release-stepper">
                <button class="stepper-btn" @click="releaseForm.replicas = Math.max(1, releaseForm.replicas - 1)">−</button>
                <input v-model.number="releaseForm.replicas" type="number" min="1" class="stepper-input" />
                <button class="stepper-btn" @click="releaseForm.replicas = releaseForm.replicas + 1">+</button>
              </div>
            </div>
          </div>
          <div class="release-foot">
            <button class="release-btn cancel" @click="closeRelease">取消</button>
            <button class="release-btn confirm" :disabled="releaseSubmitting" @click="submitRelease">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              {{ releaseSubmitting ? '发布中...' : '确认发布' }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { getPipelines, runPipeline, getBuildStats } from '@/api/platform/pipeline'
import { getEnvironmentList } from '@/api/cicd/environment'
import { getEnvColor, getEnvLabel, normalizeEnv, formatDuration, shortCommit, DEFAULT_ENVS } from '@/utils/cicdEnv'

export default {
  name: 'AppCenter',
  setup() {
    const router = useRouter()
    const loading = ref(true)
    const apps = ref([])
    const searchQuery = ref('')
    const envFilter = ref('')
    const viewMode = ref('card')
    const buildStats = ref({})       // 构建统计（今日/成功率/平均耗时/构建中）
    const environments = ref([])     // 全局环境列表（带颜色）
    // 高级筛选状态
    const showFilter = ref(false)
    const langFilter = ref('')
    const statusFilter = ref('')
    const userFilter = ref('')
    const timeFilter = ref('')
    // 发布 Modal 状态
    const releaseVisible = ref(false)
    const releaseApp = ref({})
    const releaseSubmitting = ref(false)
    const releaseForm = ref({ branch: 'main', deploy_env: '', strategy: 'rolling', replicas: 1 })
    const strategyOptions = [
      { key: 'rolling', label: '滚动更新', desc: '逐步替换旧实例，零停机' },
      { key: 'blue-green', label: '蓝绿发布', desc: '新旧环境并存，一键切换' },
      { key: 'canary', label: '灰度发布', desc: '按比例放量，稳步验证' },
    ]

    // 加载应用列表（基于流水线聚合）
    const loadApps = async () => {
      loading.value = true
      try {
        const res = await getPipelines({ page: 1, page_size: 100 })
        if (res.code === 0) {
          apps.value = (res.data?.list || res.data || []).map(p => ({
            ...p,
            language: p.language_type || p.language || detectLanguage(p),
            branch: p.git_branch || p.branch || 'main',
            git_repo: p.git_repo || p.git_url || '',
            lastRunStatus: p.last_run_status || p.lastRunStatus || '',
            lastRunTime: p.last_run_time || p.lastRunTime || '',
            target_namespace: p.target_namespace || 'default',
            target_workload_name: p.target_workload_name || '',
            target_workload_kind: p.target_workload_kind || 'Deployment',
            target_cluster_id: p.target_cluster_id || 0,
            // 关联环境ID（强隔离：优先于 deploy_env/命名空间猜测）
            environment_id: p.environment_id || 0,
            // 归一化环境（优先 deploy_env，回退命名空间猜测）
            envKey: normalizeEnv(p.deploy_env || p.target_namespace || ''),
            // 最近一次运行摘要（后端 join 填充）
            lastRunImage: p.last_run_image || '',
            lastRunTag: p.last_run_tag || '',
            lastCommit: p.last_commit || '',
            lastCommitMsg: p.last_commit_msg || '',
            lastDuration: p.last_duration || 0,
            lastTriggerUser: p.last_trigger_user || '',
          }))
        }
      } catch (e) {
        Message.error({ content: '加载应用列表失败' })
        console.error(e)
      } finally {
        loading.value = false
      }
    }

    // 加载构建统计（顶部实时看板）
    const loadBuildStats = async () => {
      try {
        const res = await getBuildStats(7)
        if (res.code === 0) {
          buildStats.value = res.data?.stats || res.data || {}
        }
      } catch (e) {
        console.error('加载构建统计失败', e)
      }
    }

    // 加载全局环境列表（用于彩色 tab）
    const loadEnvironments = async () => {
      try {
        const res = await getEnvironmentList({ page: 1, page_size: 100 })
        if (res.code === 0) {
          environments.value = res.data?.list || res.data || []
        }
      } catch (e) {
        console.error('加载环境列表失败', e)
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

    // 统计（优先后端 build-stats，回退客户端聚合）
    const runningCount = computed(() => {
      if (buildStats.value.running_builds != null) return buildStats.value.running_builds
      return apps.value.filter(a => a.lastRunStatus === 'running' || a.lastRunStatus === 'building').length
    })
    const successRateText = computed(() => {
      const r = buildStats.value.success_rate
      if (r == null) return '-'
      return `${Math.round(r * 10) / 10}%`
    })
    const avgDurationText = computed(() => {
      const d = buildStats.value.avg_duration
      if (!d) return '-'
      return formatDuration(Math.round(d))
    })

    // 环境彩色 tab 选项（优先后端环境列表，回退默认规范）
    const envOptions = computed(() => {
      if (environments.value.length > 0) {
        return environments.value.map(e => {
          const key = normalizeEnv(e.code || e.name || e.env || '')
          return { key, label: e.display_name || getEnvLabel(key), color: e.color || getEnvColor(key) }
        })
      }
      return DEFAULT_ENVS
    })

    // 命名空间 → 环境key 映射（来自后端 cicd_environment 定义）
    // 用于把 target_namespace（如 demo/k8soperation）解析成标准环境（dev/test/staging/prod）
    const nsEnvMap = computed(() => {
      const m = {}
      environments.value.forEach(e => {
        const key = normalizeEnv(e.name || e.code || e.env || '')
        if (e.namespace && key) m[e.namespace] = key
      })
      return m
    })

    // 解析应用环境：environment_id 强绑定优先 → deploy_env → 命名空间匹配环境定义 → 命名空间自身归一化
    const resolveEnvKey = (app) => {
      if (app.environment_id) {
        const env = environments.value.find(e => e.id === app.environment_id)
        if (env) return normalizeEnv(env.name || env.code || env.env || '')
      }
      const de = normalizeEnv(app.deploy_env || '')
      if (de) return de
      const ns = app.target_namespace || ''
      if (nsEnvMap.value[ns]) return nsEnvMap.value[ns]
      return normalizeEnv(ns)
    }

    // 高级筛选选项
    const statusOptions = [
      { key: 'healthy', label: '运行正常' },
      { key: 'running', label: '构建中' },
      { key: 'error', label: '构建失败' },
      { key: 'idle', label: '待部署' },
    ]
    const timeOptions = [
      { key: '', label: '全部' },
      { key: '1d', label: '今天' },
      { key: '7d', label: '近7天' },
      { key: '30d', label: '近30天' },
    ]
    // 应用类型选项：仅列出当前数据中出现的语言
    const langOptions = computed(() => {
      const seen = new Set()
      apps.value.forEach(a => { if (a.language) seen.add(a.language) })
      return Array.from(seen).map(k => ({ key: k, label: getLanguageLabel(k) }))
    })
    // 发布人选项：当前数据中出现过的发布人
    const userOptions = computed(() => {
      const seen = new Set()
      apps.value.forEach(a => { if (a.lastTriggerUser) seen.add(a.lastTriggerUser) })
      return Array.from(seen)
    })
    // 活跃筛选数量（用于按钮角标，环境 tab 单独展示不计入）
    const activeFilterCount = computed(() => {
      return [langFilter.value, statusFilter.value, userFilter.value, timeFilter.value].filter(Boolean).length
    })

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
        result = result.filter(a => resolveEnvKey(a) === envFilter.value)
      }
      if (langFilter.value) {
        result = result.filter(a => a.language === langFilter.value)
      }
      if (statusFilter.value) {
        result = result.filter(a => getAppStatus(a) === statusFilter.value)
      }
      if (userFilter.value) {
        result = result.filter(a => a.lastTriggerUser === userFilter.value)
      }
      if (timeFilter.value) {
        const days = { '1d': 1, '7d': 7, '30d': 30 }[timeFilter.value] || 0
        if (days > 0) {
          const threshold = Date.now() - days * 86400000
          result = result.filter(a => {
            const t = a.lastRunTime
            if (!t) return false
            const ms = typeof t === 'number' && t < 1e12 ? t * 1000 : new Date(t).getTime()
            return ms >= threshold
          })
        }
      }
      return result
    })

    // 重置高级筛选
    const resetFilters = () => {
      langFilter.value = ''
      statusFilter.value = ''
      userFilter.value = ''
      timeFilter.value = ''
    }

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

    // 打开发布对话框（采集 环境/版本/策略/副本）
    const runApp = (app) => {
      if (app.lastRunStatus === 'running') return
      releaseApp.value = app
      const cfg = app.deploy_config || {}
      releaseForm.value = {
        branch: app.branch || 'main',
        deploy_env: resolveEnvKey(app) || (envOptions.value[0] && envOptions.value[0].key) || '',
        strategy: cfg.strategy || 'rolling',
        replicas: Number(cfg.replicas) || 1,
      }
      releaseVisible.value = true
    }

    const closeRelease = () => { releaseVisible.value = false }

    // 提交发布：透传 环境/版本/策略/副本 到 runPipeline
    const submitRelease = async () => {
      const app = releaseApp.value
      releaseSubmitting.value = true
      try {
        Message.info({ content: `正在启动 "${app.name}" 发布...` })
        const res = await runPipeline(app.id, {
          branch: releaseForm.value.branch,
          deploy_env: releaseForm.value.deploy_env,
          strategy: releaseForm.value.strategy,
          replicas: releaseForm.value.replicas,
        })
        if (res.code === 0) {
          Message.success({ content: '发布启动成功，正在跳转构建界面...' })
          releaseVisible.value = false
          router.push(`/cicd/pipelines/${app.id}?tab=stages`)
        } else {
          throw new Error(res.msg || '启动失败')
        }
      } catch (e) {
        Message.error({ content: e.message || '发布启动失败' })
      } finally {
        releaseSubmitting.value = false
      }
    }

    onMounted(() => {
      loadApps()
      loadBuildStats()
      loadEnvironments()
    })

    // 刷新：应用列表 + 统计一同刷新
    const refreshAll = () => {
      loadApps()
      loadBuildStats()
    }

    // 复制镜像地址
    const copyImage = async (app) => {
      const img = app.lastRunImage || app.lastRunTag
      if (!img) return
      try {
        await navigator.clipboard.writeText(img)
        Message.success({ content: '镜像地址已复制' })
      } catch {
        Message.error({ content: '复制失败' })
      }
    }

    return {
      loading, apps, searchQuery, envFilter, viewMode,
      buildStats, runningCount, successRateText, avgDurationText, envOptions, filteredApps,
      showFilter, langFilter, statusFilter, userFilter, timeFilter,
      langOptions, statusOptions, userOptions, timeOptions, activeFilterCount, resetFilters,
      loadApps, refreshAll, getLanguageLabel, getLanguageClass, getAppStatus, getAppStatusText,
      getAppStatusClass, getBuildStatusText, formatGitRepo, formatTime,
      getEnvColor, getEnvLabel, shortCommit, formatDuration, copyImage, resolveEnvKey,
      viewApp, editApp, viewReleases, runApp,
      releaseVisible, releaseApp, releaseSubmitting, releaseForm, strategyOptions,
      closeRelease, submitRelease
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

/* ===== 环境彩色 tab 栏 ===== */
.env-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #fff;
  border: 1px solid #f0f0f3;
  border-radius: 10px;
}

.env-bar-label {
  font-size: 13px;
  font-weight: 600;
  color: #4e5969;
  flex-shrink: 0;
}

.env-tabs { display: flex; flex-wrap: wrap; gap: 8px; }

.env-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid #e5e6eb;
  border-radius: 16px;
  background: #fff;
  font-size: 13px;
  color: #4e5969;
  cursor: pointer;
  transition: all 0.2s;
}

.env-tab:hover { border-color: #c9cdd4; background: #f7f8fa; }
.env-tab.active { font-weight: 600; }

.env-tab-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* ===== 高级筛选 ===== */
.filter-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 12px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #fff;
  font-size: 13px;
  color: #4e5969;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-toggle:hover { border-color: #c9cdd4; background: #f7f8fa; }
.filter-toggle.active { color: #165DFF; border-color: #165DFF; background: #eef3ff; }
.filter-toggle svg { width: 15px; height: 15px; }

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #165DFF;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
}

.filter-panel {
  background: #fff;
  border: 1px solid #f0f0f3;
  border-radius: 10px;
  padding: 16px 20px;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.filter-group {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.filter-label {
  flex-shrink: 0;
  width: 64px;
  font-size: 13px;
  color: #86909c;
  padding-top: 5px;
}

.filter-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  padding: 4px 12px;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  background: #fff;
  font-size: 12px;
  color: #4e5969;
  cursor: pointer;
  transition: all 0.2s;
}

.chip:hover { border-color: #c9cdd4; background: #f7f8fa; }
.chip.active {
  color: #165DFF;
  border-color: #165DFF;
  background: #eef3ff;
  font-weight: 600;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f7f8fa;
  padding-top: 12px;
}

.filter-reset {
  padding: 5px 16px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #fff;
  font-size: 12px;
  color: #4e5969;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-reset:hover:not(:disabled) { border-color: #f53f3f; color: #f53f3f; }
.filter-reset:disabled { opacity: 0.5; cursor: not-allowed; }

.filter-slide-enter-active, .filter-slide-leave-active { transition: all 0.2s ease; }
.filter-slide-enter-from, .filter-slide-leave-to { opacity: 0; transform: translateY(-6px); }

/* ===== 版本/提交 ===== */
.ver-item { cursor: pointer; }

.ver-tag {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 11px;
  color: #165DFF;
  background: #eef3ff;
  padding: 1px 6px;
  border-radius: 4px;
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ver-tag-sm { cursor: pointer; display: inline-block; }
.ver-item:hover .ver-tag { background: #dce6ff; }

.copy-icon { width: 12px; height: 12px; color: #c9cdd4; }
.ver-item:hover .copy-icon { color: #165DFF; }

.commit-hash {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 11px;
  color: #86909c;
  background: #f2f3f5;
  padding: 1px 6px;
  border-radius: 4px;
}

.build-user { font-size: 11px; color: #4e5969; font-weight: 500; }
.build-dur { font-size: 11px; color: #86909c; }
.user-text { font-size: 12px; color: #4e5969; }
.dur-text {
  display: block;
  font-size: 11px;
  color: #c9cdd4;
  margin-top: 2px;
}

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
  grid-template-columns: 2.2fr 0.85fr 0.9fr 1.3fr 0.95fr 1fr 96px;
  padding: 12px 16px;
  background: #f7f8fa;
  font-size: 12px;
  font-weight: 600;
  color: #86909c;
  border-bottom: 1px solid #f0f0f3;
}

.list-row {
  display: grid;
  grid-template-columns: 2.2fr 0.85fr 0.9fr 1.3fr 0.95fr 1fr 96px;
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

/* ===== 发布 Modal ===== */
.release-mask {
  position: fixed;
  inset: 0;
  background: rgba(29, 33, 41, 0.45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.release-dialog {
  width: 480px;
  max-width: calc(100vw - 32px);
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

.release-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px;
  border-bottom: 1px solid #f2f3f5;
}

.release-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.release-app-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  background: #4e7cf6;
}

.release-close {
  display: flex;
  border: none;
  background: transparent;
  color: #86909c;
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
}
.release-close:hover { background: #f2f3f5; color: #4e5969; }
.release-close svg { width: 18px; height: 18px; }

.release-body {
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.release-field { display: flex; flex-direction: column; gap: 8px; }

.release-label {
  font-size: 13px;
  font-weight: 500;
  color: #4e5969;
}

.release-chips { display: flex; flex-wrap: wrap; gap: 8px; }

.release-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}
.release-chip:hover { border-color: #c9cdd4; }
.release-chip.active { font-weight: 600; }
.release-chip-dot { width: 8px; height: 8px; border-radius: 50%; }

.release-input {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  font-size: 14px;
  color: #1d2129;
  box-sizing: border-box;
  transition: border-color 0.15s;
}
.release-input:focus { outline: none; border-color: #4e7cf6; }

.release-strategies {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.release-strategy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  border: 1px solid #e5e6eb;
  border-radius: 10px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
}
.release-strategy:hover { border-color: #c9cdd4; }
.release-strategy.active {
  border-color: #4e7cf6;
  background: #4e7cf60d;
}
.release-strategy-name { font-size: 13px; font-weight: 600; color: #1d2129; }
.release-strategy-desc { font-size: 11px; color: #86909c; line-height: 1.4; }

.release-stepper {
  display: inline-flex;
  align-items: center;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  overflow: hidden;
  width: fit-content;
}

.stepper-btn {
  width: 36px;
  height: 38px;
  border: none;
  background: #f7f8fa;
  color: #4e5969;
  font-size: 18px;
  cursor: pointer;
}
.stepper-btn:hover { background: #e5e6eb; }

.stepper-input {
  width: 64px;
  height: 38px;
  border: none;
  border-left: 1px solid #e5e6eb;
  border-right: 1px solid #e5e6eb;
  text-align: center;
  font-size: 14px;
  color: #1d2129;
}
.stepper-input:focus { outline: none; }

.release-foot {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 22px;
  border-top: 1px solid #f2f3f5;
}

.release-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.release-btn svg { width: 15px; height: 15px; }
.release-btn.cancel {
  border: 1px solid #e5e6eb;
  background: #fff;
  color: #4e5969;
}
.release-btn.cancel:hover { background: #f7f8fa; }
.release-btn.confirm {
  border: 1px solid #4e7cf6;
  background: #4e7cf6;
  color: #fff;
}
.release-btn.confirm:hover { background: #3a6ae8; }
.release-btn.confirm:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; }
</style>

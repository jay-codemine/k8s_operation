<template>
  <div class="releases-page">
    <!-- 顶部 Banner - Rancher 深色风格 -->
    <div class="page-banner">
      <div class="banner-inner">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
          </div>
          <div>
            <h1 class="banner-title">发布单中心</h1>
            <p class="banner-desc">直接发布：选定镜像并指定集群（开发 / 测试 / 生产）一键部署，覆盖审批 → 部署 → 回滚全流程</p>
            <p class="banner-hint">需要一次构建、跨环境逐级晋级同一镜像？请使用<router-link class="banner-inline-link" to="/cicd/promotion">镜像晋级</router-link></p>
          </div>
        </div>
        <div class="banner-actions">
          <button class="btn-banner-new-app" @click="$router.push('/cicd/pipelines/create')" title="新建应用，配置构建流水线">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
            <span>新建应用</span>
          </button>
          <button class="btn-banner-sync" @click="syncFromPipeline" :disabled="syncing">
            <svg :class="{ spinning: syncing }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 2v6h-6"/><path d="M3 12a9 9 0 0 1 15-6.7L21 8M3 22v-6h6"/><path d="M21 12a9 9 0 0 1-15 6.7L3 16"/></svg>
            <span>{{ syncing ? '同步中...' : '同步流水线' }}</span>
          </button>
          <button class="btn-banner-refresh" @click="loadAll" :disabled="loading">
            <svg :class="{ spinning: loading }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
            <span>刷新</span>
          </button>
          <button class="btn-banner-create" @click="showCreateDialog = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            <span>创建发布</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 指标卡片 - Kuboard 风格 -->
    <div class="metrics-row">
      <div class="metric-card" :class="{ active: statusFilter === '' }" @click="setFilter('')">
        <div class="metric-icon-wrap total">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-num">{{ statsData.total }}</span>
          <span class="metric-label">总发布数</span>
        </div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'deploying' }" @click="setFilter('deploying')">
        <div class="metric-icon-wrap deploying">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-num">{{ statsData.deploying }}</span>
          <span class="metric-label">部署中</span>
        </div>
        <span v-if="statsData.deploying > 0" class="metric-badge deploying">LIVE</span>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'success' }" @click="setFilter('success')">
        <div class="metric-icon-wrap success">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-num">{{ statsData.success }}</span>
          <span class="metric-label">发布成功</span>
        </div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'failed' }" @click="setFilter('failed')">
        <div class="metric-icon-wrap failed">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-num">{{ statsData.failed }}</span>
          <span class="metric-label">发布失败</span>
        </div>
      </div>
      <div class="metric-card" :class="{ active: statusFilter === 'rollback' }" @click="setFilter('rollback')">
        <div class="metric-icon-wrap rollback">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
        </div>
        <div class="metric-body">
          <span class="metric-num">{{ statsData.rollback }}</span>
          <span class="metric-label">已回滚</span>
        </div>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="content-area">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <h3 class="section-title">发布记录</h3>
          <span class="record-badge">{{ total }} 条</span>
        </div>
        <div class="toolbar-right">
          <button class="filter-toggle-btn" :class="{ active: showFilterPanel }" @click="showFilterPanel = !showFilterPanel">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/></svg>
            <span>筛选</span>
            <span v-if="activeFilterCount > 0" class="filter-count">{{ activeFilterCount }}</span>
          </button>
          <div class="view-toggle">
            <button :class="['toggle-btn', { active: releaseViewMode === 'card' }]" @click="releaseViewMode = 'card'" title="卡片视图">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 3h8v8H3V3zm0 10h8v8H3v-8zm10-10h8v8h-8V3zm0 10h8v8h-8v-8z"/></svg>
            </button>
            <button :class="['toggle-btn', { active: releaseViewMode === 'table' }]" @click="releaseViewMode = 'table'" title="表格视图">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 4h18v2H3V4zm0 7h18v2H3v-2zm0 7h18v2H3v-2z"/></svg>
            </button>
          </div>
          <div class="search-box" :class="{ focused: searchFocused }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="searchKeyword" placeholder="搜索应用名、镜像..." @input="handleSearch" @focus="searchFocused = true" @blur="searchFocused = false" />
            <button v-if="searchKeyword" class="clear-btn" @click="clearSearch">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- 高级筛选面板 -->
      <transition name="filter-slide">
        <div v-if="showFilterPanel" class="filter-panel">
          <div class="filter-row">
            <div class="filter-item">
              <label class="filter-label">应用</label>
              <select v-model="filterApp" class="filter-select" @change="applyFilters">
                <option value="">全部应用</option>
                <option v-for="p in pipelines" :key="p.id" :value="p.name">{{ p.name }}</option>
              </select>
            </div>
            <div class="filter-item">
              <label class="filter-label">状态</label>
              <select v-model="filterStatus" class="filter-select" @change="applyFilters">
                <option value="">全部状态</option>
                <option value="Pending">等待中</option>
                <option value="AwaitingApproval">待审批</option>
                <option value="Running">部署中</option>
                <option value="Succeeded">发布成功</option>
                <option value="Failed">发布失败</option>
                <option value="Rollback">已回滚</option>
                <option value="Canceled">已取消</option>
              </select>
            </div>
            <div class="filter-item">
              <label class="filter-label">命名空间</label>
              <select v-model="filterNamespace" class="filter-select" @change="applyFilters">
                <option value="">全部命名空间</option>
                <option v-for="ns in namespaceOptions" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>
            <div class="filter-item">
              <label class="filter-label">时间范围</label>
              <select v-model="filterTimeRange" class="filter-select" @change="applyFilters">
                <option value="">全部时间</option>
                <option value="1h">最近 1 小时</option>
                <option value="24h">最近 24 小时</option>
                <option value="7d">最近 7 天</option>
                <option value="30d">最近 30 天</option>
                <option value="90d">最近 90 天</option>
              </select>
            </div>
            <div class="filter-item">
              <label class="filter-label">发布人</label>
              <select v-model="filterCreator" class="filter-select" @change="applyFilters">
                <option value="">全部</option>
                <option v-for="u in creatorOptions" :key="u" :value="u">{{ u }}</option>
              </select>
            </div>
            <div class="filter-actions">
              <button class="filter-reset-btn" @click="resetFilters" :disabled="activeFilterCount === 0">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><path d="M19 6l-1.5 14.5a2 2 0 0 1-2 1.5H8.5a2 2 0 0 1-2-1.5L5 6"/></svg>
                重置
              </button>
            </div>
          </div>
        </div>
      </transition>

      <!-- 批量操作工具栏 -->
      <transition name="batch-slide">
        <div v-if="selectedIds.length > 0" class="batch-toolbar-v2">
          <div class="batch-top-row">
            <span class="batch-count">已选择 <strong>{{ selectedIds.length }}</strong> 个应用</span>
            <div class="batch-actions">
              <button class="batch-btn publish" @click="handleBatchRetry" :disabled="batchLoading">批量发布</button>
              <button class="batch-btn rollback" @click="handleBatchRollback" :disabled="batchLoading">批量回滚</button>
              <button class="batch-btn stop" @click="handleBatchCancel" :disabled="batchLoading">批量取消</button>
              <button class="batch-btn clear" @click="selectedIds = []">清除全部</button>
            </div>
          </div>
          <div class="batch-chips">
            <span v-for="id in selectedIds" :key="id" class="batch-chip">
              <span class="chip-name">{{ releases.find(r => r.id === id)?.app_name || releases.find(r => r.id === id)?.workload_name || '#' + id }}</span>
              <span class="chip-status" :class="releases.find(r => r.id === id)?.status?.toLowerCase()">{{ releases.find(r => r.id === id)?.status || '—' }}</span>
              <button class="chip-remove" @click="selectedIds = selectedIds.filter(sid => sid !== id)">&times;</button>
            </span>
          </div>
        </div>
      </transition>

      <!-- 加载 -->
      <div v-if="loading" class="loading-state">
        <div class="loader"><div class="dot"></div><div class="dot"></div><div class="dot"></div></div>
        <span>正在加载发布记录...</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="releases.length === 0" class="empty-state">
        <!-- 无应用引导：先建应用再发布 -->
        <template v-if="pipelines.length === 0">
          <div class="onboard-cards">
            <div class="onboard-hero">
              <div class="onboard-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                  <path d="M2 17l10 5 10-5"/>
                  <path d="M2 12l10 5 10-5"/>
                </svg>
              </div>
              <h3>还没有配置任何应用</h3>
              <p>先创建一个应用，即可自动完成构建、部署、发布全流程</p>
            </div>
            <div class="onboard-steps">
              <div class="onboard-step">
                <div class="step-num">1</div>
                <div class="step-content">
                  <div class="step-title">新建应用</div>
                  <div class="step-desc">填写应用名 + Git 仓库 + 语言类型，3 个字段完成配置</div>
                </div>
              </div>
              <div class="onboard-step-arrow">→</div>
              <div class="onboard-step">
                <div class="step-num">2</div>
                <div class="step-content">
                  <div class="step-title">触发构建</div>
                  <div class="step-desc">git push 或手动触发，Jenkins 自动构建镜像</div>
                </div>
              </div>
              <div class="onboard-step-arrow">→</div>
              <div class="onboard-step">
                <div class="step-num">3</div>
                <div class="step-content">
                  <div class="step-title">一键发布</div>
                  <div class="step-desc">选应用 + 填版本号，直接部署到 K8s</div>
                </div>
              </div>
            </div>
            <div class="onboard-actions">
              <button class="btn-onboard-primary" @click="$router.push('/cicd/pipelines/create')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                快速新建应用
              </button>
              <button class="btn-onboard-secondary" @click="showCreateDialog = true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
                手动创建发布单
              </button>
            </div>
          </div>
        </template>
        <!-- 有应用但无发布记录 -->
        <template v-else>
          <div class="empty-svg">
            <svg viewBox="0 0 200 160" fill="none">
              <rect x="35" y="15" width="130" height="95" rx="10" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
              <rect x="52" y="38" width="96" height="8" rx="4" fill="#d6e4ff"/>
              <rect x="52" y="54" width="68" height="8" rx="4" fill="#d6e4ff"/>
              <rect x="52" y="70" width="80" height="8" rx="4" fill="#d6e4ff"/>
              <circle cx="100" cy="135" r="18" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
              <path d="M93 135l4 4 10-10" stroke="#4e7cf6" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <h3>暂无发布记录</h3>
          <p style="margin-bottom:16px;">已有 {{ pipelines.length }} 个应用，立即发布</p>
          <button class="btn-onboard-primary" style="margin:0 auto;" @click="showCreateDialog = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            创建第一个发布
          </button>
        </template>
      </div>

      <!-- 卡片视图 - 大厂级发布单卡片（3层信息结构） -->
      <div v-else-if="releaseViewMode === 'card'" class="release-cards-grid">
        <div v-for="rel in releases" :key="rel.id" class="release-card" :class="[`card-${normalizeStatus(rel.status)}`, { selected: selectedIds.includes(rel.id) }]">
          <!-- 卡片头部：应用 + 状态 -->
          <div class="rc-header">
            <div class="rc-check">
              <input type="checkbox" class="row-checkbox" :checked="selectedIds.includes(rel.id)" @change="toggleSelect(rel.id)" />
            </div>
            <div class="rc-app-info">
              <div class="rc-avatar" :class="normalizeStatus(rel.status)">
                {{ (rel.app_name || rel.name || '?').charAt(0).toUpperCase() }}
              </div>
              <div class="rc-app-meta">
                <span class="rc-app-name">{{ rel.app_name || rel.name || '-' }}</span>
                <span class="rc-app-id">#{{ rel.id }}</span>
              </div>
            </div>
            <span class="rc-status-badge" :class="normalizeStatus(rel.status)">
              <span class="rc-status-dot"></span>
              {{ statusText(rel.status) }}
            </span>
          </div>

          <!-- Tier 1: 基础信息（谁发的 / 环境 / 策略） -->
          <div class="rc-tier">
            <div class="rc-field">
              <span class="rc-label">模式</span>
              <span :class="['deploy-mode-tag', rel.deploy_mode === 'gitops' ? 'mode-gitops' : 'mode-jenkins']">
                {{ rel.deploy_mode === 'gitops' ? 'GitOps' : 'Jenkins' }}
              </span>
            </div>
            <div class="rc-field">
              <span class="rc-label">环境</span>
              <span class="rc-value ns">{{ rel.namespace || 'default' }}</span>
            </div>
            <div class="rc-field">
              <span class="rc-label">策略</span>
              <span class="rc-value">{{ strategyText(rel.strategy) || '滚动更新' }}</span>
            </div>
            <div class="rc-field">
              <span class="rc-label">时间</span>
              <span class="rc-value">{{ formatDate(rel.created_at) }}</span>
            </div>
          </div>

          <!-- Tier 2: 构建信息（镜像 / tag） -->
          <div class="rc-tier">
            <div class="rc-field full">
              <span class="rc-label">镜像</span>
              <code class="rc-image" :title="getFullImage(rel)">{{ formatImage(rel) }}</code>
            </div>
          </div>

          <!-- Tier 3: 部署信息（K8s目标） -->
          <div class="rc-tier">
            <div class="rc-field">
              <span class="rc-label">工作负载</span>
              <code class="rc-workload">{{ rel.workload_kind || 'Deployment' }}/{{ rel.workload_name || '-' }}</code>
            </div>
            <div class="rc-field" v-if="rel.container_name">
              <span class="rc-label">容器</span>
              <span class="rc-value">{{ rel.container_name }}</span>
            </div>
          </div>

          <!-- 失败原因 -->
          <div v-if="rel.message && (rel.status === 'Failed' || rel.status === 'Canceled')" class="rc-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <span>{{ rel.message.length > 80 ? rel.message.slice(0, 80) + '...' : rel.message }}</span>
          </div>

          <!-- 卡片操作栏 -->
          <div class="rc-actions">
            <button class="rc-act-btn" @click="viewRelease(rel)" title="查看详情">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              详情
            </button>
            <button v-if="normalizeStatus(rel.status) === 'success' || normalizeStatus(rel.status) === 'failed'" class="rc-act-btn primary" @click="retryRelease(rel)" title="重新发布">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              重发
            </button>
            <button v-if="normalizeStatus(rel.status) === 'success' || rel.status === 'Failed'" class="rc-act-btn warning" @click="rollbackRelease(rel)" :title="rel.status === 'Failed' ? '回滚到部署前版本' : '回滚'">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
              回滚
            </button>
            <button v-if="normalizeStatus(rel.status) === 'deploying'" class="rc-act-btn danger" @click="cancelRelease(rel)" title="取消">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
              取消
            </button>
            <button v-if="canEdit(rel.status)" class="rc-act-btn" @click="editRelease(rel)" title="编辑">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              编辑
            </button>
            <button v-if="canDelete(rel.status)" class="rc-act-btn danger-text" @click="deleteRelease(rel)" title="删除">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              删除
            </button>
          </div>
        </div>
      </div>

      <!-- 表格视图 - Rancher 风格 -->
      <div v-else class="table-wrapper">
        <table class="data-table">
          <thead>
            <tr>
              <th style="width: 40px;">
                <input type="checkbox" class="row-checkbox" :checked="isAllSelected" :indeterminate.prop="isIndeterminate" @change="toggleAll" />
              </th>
              <th>应用</th>
              <th>模式</th>
              <th>版本号</th>
              <th>状态</th>
              <th>命名空间</th>
              <th>集群</th>
              <th>工作负载</th>
              <th>镜像</th>
              <th>发布人</th>
              <th>策略</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rel in releases" :key="rel.id" :class="[`row-${normalizeStatus(rel.status)}`, { 'row-selected': selectedIds.includes(rel.id) }]">
              <td>
                <input type="checkbox" class="row-checkbox" :checked="selectedIds.includes(rel.id)" @change="toggleSelect(rel.id)" />
              </td>
              <td>
                <div class="app-cell">
                  <div class="app-avatar" :class="normalizeStatus(rel.status)">
                    {{ (rel.app_name || rel.name || '?').charAt(0).toUpperCase() }}
                  </div>
                  <div class="app-info">
                    <span class="app-name">{{ rel.app_name || rel.name || '-' }}</span>
                    <span class="app-id">#{{ rel.id }}</span>
                  </div>
                </div>
              </td>
              <td>
                <code class="version-tag" v-if="rel.image_tag">{{ rel.image_tag }}</code>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span :class="['deploy-mode-tag', rel.deploy_mode === 'gitops' ? 'mode-gitops' : 'mode-jenkins']">
                  {{ rel.deploy_mode === 'gitops' ? 'GitOps' : 'Jenkins' }}
                </span>
              </td>
              <td>
                <div class="status-cell">
                  <span class="status-pill" :class="normalizeStatus(rel.status)">
                    <span class="status-dot"></span>
                    {{ statusText(rel.status) }}
                  </span>
                  <span v-if="rel.message && (rel.status === 'Failed' || rel.status === 'Canceled')" class="fail-reason" :title="rel.message">
                    {{ rel.message.length > 40 ? rel.message.slice(0, 40) + '...' : rel.message }}
                  </span>
                </div>
              </td>
              <td><span class="ns-badge">{{ rel.namespace || 'default' }}</span></td>
              <td><span class="cluster-badge" v-if="rel.cluster_name || rel.cluster_id">{{ rel.cluster_name || `cluster-${rel.cluster_id}` }}</span><span v-else class="text-muted">-</span></td>
              <td>
                <div class="workload-cell">
                  <code class="workload-tag">{{ rel.workload_kind || 'Deployment' }}/{{ rel.workload_name || '-' }}</code>
                  <span v-if="rel.container_name" class="container-tag">{{ rel.container_name }}</span>
                </div>
              </td>
              <td>
                <code class="image-code" :title="getFullImage(rel)">{{ formatImage(rel) }}</code>
              </td>
              <td>
                <div class="creator-cell" v-if="rel.creator || rel.created_by">
                  <span class="creator-avatar">{{ (rel.creator || rel.created_by || '?').charAt(0).toUpperCase() }}</span>
                  <span class="creator-name">{{ rel.creator || rel.created_by || '-' }}</span>
                </div>
                <span v-else class="text-muted">系统</span>
              </td>
              <td><span class="strategy-tag" v-if="rel.strategy">{{ strategyText(rel.strategy) }}</span><span v-else class="text-muted">-</span></td>
              <td><span class="time-text">{{ formatDate(rel.created_at) }}</span></td>
              <td>
                <div class="actions-cell">
                  <button class="act-btn view" @click="viewRelease(rel)" title="查看详情">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                  </button>
                  <button v-if="canEdit(rel.status)" class="act-btn edit" @click="editRelease(rel)" title="编辑">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button v-if="normalizeStatus(rel.status) === 'deploying'" class="act-btn cancel" @click="cancelRelease(rel)" title="取消">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
                  </button>
                  <button v-if="normalizeStatus(rel.status) === 'success' || rel.status === 'Failed'" class="act-btn rollback" @click="rollbackRelease(rel)" :title="rel.status === 'Failed' ? '回滚到部署前版本' : '回滚'">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                  </button>
                  <button v-if="normalizeStatus(rel.status) === 'success' || normalizeStatus(rel.status) === 'failed' || normalizeStatus(rel.status) === 'rollback'" class="act-btn retry" @click="retryRelease(rel)" :title="normalizeStatus(rel.status) === 'failed' ? '重试' : '重新部署'">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                  </button>
                  <button v-if="canDelete(rel.status)" class="act-btn delete" @click="deleteRelease(rel)" title="删除">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页（现代化三段式布局） -->
      <div v-if="total > 0" class="pagination-wrapper">
        <div class="pagination-left">
          <span class="pagination-summary">共 <strong>{{ total }}</strong> 条</span>
        </div>
        <div class="pagination-center">
          <button class="pagination-btn" @click="goToPage(1)" :disabled="currentPage === 1" title="首页">«</button>
          <button class="pagination-btn" @click="goToPage(currentPage - 1)" :disabled="currentPage === 1" title="上一页">‹</button>
          <template v-for="page in visiblePages" :key="page">
            <button v-if="typeof page === 'number'" class="pagination-btn page-number" :class="{ active: currentPage === page }" @click="goToPage(page)">{{ page }}</button>
            <span v-else class="pagination-ellipsis">...</span>
          </template>
          <button class="pagination-btn" @click="goToPage(currentPage + 1)" :disabled="currentPage === totalPages" title="下一页">›</button>
          <button class="pagination-btn" @click="goToPage(totalPages)" :disabled="currentPage === totalPages" title="尾页">»</button>
        </div>
        <div class="pagination-right">
          <select v-model.number="pageSizeRef" @change="onPageSizeChange" class="page-size-select">
            <option :value="10">10 条/页</option>
            <option :value="20">20 条/页</option>
            <option :value="50">50 条/页</option>
            <option :value="100">100 条/页</option>
          </select>
          <span class="pagination-goto">前往</span>
          <input v-model.number="jumpPage" type="number" min="1" :max="totalPages" class="page-jump-input" @keyup.enter="jumpToPage" />
        </div>
      </div>
    </div>

    <!-- 创建发布弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
          <div class="modal-dialog">
            <div class="modal-head create">
              <div class="modal-head-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              </div>
              <h3>创建发布</h3>
              <button class="modal-close" @click="showCreateDialog = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-body">
              <!-- 无应用时：引导新建 -->
              <div v-if="pipelines.length === 0" class="no-app-guide">
                <div class="guide-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5"/>
                    <path d="M2 12l10 5 10-5"/>
                  </svg>
                </div>
                <div class="guide-text">
                  <div class="guide-title">还没有配置应用</div>
                  <div class="guide-desc">建议先创建应用配置，以后发布只需选应用+填版本号</div>
                </div>
                <button class="btn-guide-create" @click="showCreateDialog = false; $router.push('/cicd/pipelines/create')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                  新建应用
                </button>
              </div>
              <div v-if="pipelines.length === 0" class="guide-divider">
                <span>或者手动配置发布目标</span>
              </div>
              <div class="field">
                <label>选择应用 <span v-if="pipelines.length > 0" class="required">*</span><span v-else class="optional">（选填）</span></label>
                <select v-model="createForm.pipeline_id" @change="onPipelineSelect">
                  <option value="">请选择要发布的应用</option>
                  <option v-for="p in pipelines" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
              </div>
              <!-- 应用配置预览 -->
              <div v-if="selectedPipelineInfo" class="pipeline-inherit-hint">
                <div class="hint-title">📦 部署目标：</div>
                <div class="hint-items" v-if="selectedPipelineInfo.namespace || selectedPipelineInfo.workload || selectedPipelineInfo.image_repo">
                  <span v-if="selectedPipelineInfo.image_repo">镜像仓库: <b>{{ selectedPipelineInfo.image_repo }}</b></span>
                  <span v-if="selectedPipelineInfo.namespace">命名空间: <b>{{ selectedPipelineInfo.namespace }}</b></span>
                  <span v-if="selectedPipelineInfo.workload">工作负载: <b>{{ selectedPipelineInfo.workload }}</b></span>
                  <span v-if="selectedPipelineInfo.container">容器: <b>{{ selectedPipelineInfo.container }}</b></span>
                </div>
                <div class="hint-items hint-warning" v-else>
                  <span>⚠️ 该应用尚未配置部署目标（命名空间/工作负载/镜像仓库），发布后将仅触发构建</span>
                  <a class="hint-link" @click="showCreateDialog = false; $router.push(`/cicd/pipelines/${createForm.pipeline_id}?tab=settings`)">去配置 →</a>
                </div>
                <div class="hint-note" v-if="selectedPipelineInfo.namespace || selectedPipelineInfo.workload || selectedPipelineInfo.image_repo">只需填写版本号即可发布</div>
              </div>
              <div class="field">
                <label>版本号 / 镜像标签 <span class="required">*</span></label>
                <input v-model="createForm.version" placeholder="例如: v1.2.3 或 abc1234" />
              </div>
              <!-- 未选择应用时：手动填写部署目标（新应用首次发布） -->
              <template v-if="!createForm.pipeline_id">
                <div class="field">
                  <label>应用名称 <span class="required">*</span></label>
                  <input v-model="createForm.name" placeholder="例如: user-service" />
                </div>
                <div class="field-row">
                  <div class="field">
                    <label>命名空间</label>
                    <input v-model="createForm.namespace" placeholder="production" />
                  </div>
                  <div class="field">
                    <label>工作负载类型</label>
                    <select v-model="createForm.workload_kind">
                      <option value="Deployment">Deployment</option>
                      <option value="StatefulSet">StatefulSet</option>
                      <option value="DaemonSet">DaemonSet</option>
                    </select>
                  </div>
                </div>
                <div class="field-row">
                  <div class="field">
                    <label>工作负载名称</label>
                    <input v-model="createForm.workload_name" placeholder="与应用名一致则可留空" />
                  </div>
                  <div class="field">
                    <label>容器名称</label>
                    <input v-model="createForm.container_name" placeholder="留空则默认第一个容器" />
                  </div>
                </div>
                <div class="field">
                  <label>镜像地址 <span class="required">*</span></label>
                  <input v-model="createForm.image" placeholder="harbor.example.com/proj/app:v1.0.0" />
                </div>
                <div class="field">
                  <label>目标集群 <span class="required">*</span></label>
                  <select v-model="createForm.cluster_id">
                    <option value="">请选择目标集群</option>
                    <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name || c.name }}</option>
                  </select>
                </div>
              </template>
              <div class="field">
                <label>备注 <span class="optional">(选填)</span></label>
                <textarea v-model="createForm.remark" placeholder="发布说明..." rows="2"></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="showCreateDialog = false">取消</button>
              <button class="btn-confirm create" @click="handleCreate" :disabled="creating">
                {{ creating ? '创建中...' : '创建发布' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 确认弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showConfirmDialog" class="modal-overlay" @click.self="showConfirmDialog = false">
          <div class="modal-dialog small">
            <div class="modal-head" :class="confirmType">
              <div class="modal-head-icon">
                <svg v-if="confirmType === 'warning'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                <svg v-else-if="confirmType === 'danger'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
              </div>
              <h3>{{ confirmTitle }}</h3>
              <button class="modal-close" @click="showConfirmDialog = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-body"><p class="confirm-msg">{{ confirmMessage }}</p></div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="showConfirmDialog = false">取消</button>
              <button class="btn-confirm" :class="confirmType" @click="confirmAction" :disabled="confirming">
                {{ confirming ? '处理中...' : confirmBtnText }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 编辑发布弹窗 -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showEditDialog" class="modal-overlay" @click.self="showEditDialog = false">
          <div class="modal-dialog">
            <div class="modal-head create">
              <div class="modal-head-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </div>
              <h3>编辑发布单</h3>
              <button class="modal-close" @click="showEditDialog = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="modal-body">
              <div class="field">
                <label>应用名称</label>
                <input v-model="editForm.app_name" placeholder="应用名称" />
              </div>
              <div class="field-row">
                <div class="field">
                  <label>命名空间</label>
                  <input v-model="editForm.namespace" placeholder="default" />
                </div>
                <div class="field">
                  <label>工作负载类型</label>
                  <select v-model="editForm.workload_kind">
                    <option value="Deployment">Deployment</option>
                    <option value="StatefulSet">StatefulSet</option>
                    <option value="DaemonSet">DaemonSet</option>
                  </select>
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>工作负载名称</label>
                  <input v-model="editForm.workload_name" placeholder="工作负载名称" />
                </div>
                <div class="field">
                  <label>容器名称</label>
                  <input v-model="editForm.container_name" placeholder="容器名称" />
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>镜像仓库</label>
                  <input v-model="editForm.image_repo" placeholder="镜像仓库地址" />
                </div>
                <div class="field">
                  <label>镜像标签</label>
                  <input v-model="editForm.image_tag" placeholder="latest" />
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>发布策略</label>
                  <select v-model="editForm.strategy">
                    <option value="rolling">滚动更新</option>
                    <option value="recreate">重建</option>
                    <option value="canary">金丝雀</option>
                    <option value="bluegreen">蓝绿部署</option>
                  </select>
                </div>
                <div class="field">
                  <label>超时时间(秒)</label>
                  <input v-model.number="editForm.timeout_sec" type="number" placeholder="300" />
                </div>
              </div>
              <div class="field">
                <label>备注 <span class="optional">(选填)</span></label>
                <textarea v-model="editForm.message" placeholder="发布说明..." rows="2"></textarea>
              </div>
            </div>
            <div class="modal-foot">
              <button class="btn-cancel" @click="showEditDialog = false">取消</button>
              <button class="btn-confirm create" @click="handleEdit" :disabled="editing">
                {{ editing ? '保存中...' : '保存修改' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>

  <!-- 批量发布确认弹窗 -->
  <RollbackDialog
    :visible="showBatchPublishDialog"
    :confirming="batchLoading"
    batch-mode
    batch-action="发布"
    :selected-items="batchPublishItems"
    @close="showBatchPublishDialog = false"
    @confirm="confirmBatchPublish"
  />
  <!-- 批量回滚确认弹窗 -->
  <RollbackDialog
    :visible="showBatchRollbackDialog"
    :confirming="batchLoading"
    batch-mode
    :selected-items="batchRollbackItems"
    @close="showBatchRollbackDialog = false"
    @confirm="confirmBatchRollback"
  />
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { getPipelines } from '@/api/platform/pipeline'
import { getK8sClusterList } from '@/api/platform/cluster'
import RollbackDialog from '@/components/RollbackDialog.vue'
import {
  getReleases,
  getReleaseStats,
  createRelease,
  updateRelease as updateReleaseApi,
  deleteRelease as deleteReleaseApi,
  cancelRelease as cancelReleaseApi,
  rollbackRelease as rollbackReleaseApi,
  retryRelease as retryReleaseApi,
  batchRetryRelease as batchRetryReleaseApi,
  batchRollbackRelease as batchRollbackReleaseApi,
  batchCancelRelease as batchCancelReleaseApi,
  syncReleasesFromPipeline as syncReleasesApi
} from '@/api/cicd'

export default {
  name: 'Releases',
  setup() {
    const router = useRouter()
    const loading = ref(true)
    const releases = ref([])
    const searchKeyword = ref('')
    const searchFocused = ref(false)
    const statusFilter = ref('')
    const releaseViewMode = ref('card') // 默认卡片视图
    const currentPage = ref(1)
    const pageSizeRef = ref(10)
    const total = ref(0)
    const jumpPage = ref(1)

    // 高级筛选
    const showFilterPanel = ref(false)
    const filterApp = ref('')
    const filterStatus = ref('')
    const filterNamespace = ref('')
    const filterTimeRange = ref('')
    const filterCreator = ref('')

    // 筛选选项（从已加载数据中动态提取）
    const namespaceOptions = computed(() => {
      const nsSet = new Set(releases.value.map(r => r.namespace).filter(Boolean))
      return [...nsSet].sort()
    })
    const creatorOptions = computed(() => {
      const cSet = new Set(releases.value.map(r => r.creator || r.created_by).filter(Boolean))
      return [...cSet].sort()
    })
    const activeFilterCount = computed(() => {
      let count = 0
      if (filterApp.value) count++
      if (filterStatus.value) count++
      if (filterNamespace.value) count++
      if (filterTimeRange.value) count++
      if (filterCreator.value) count++
      return count
    })
    const applyFilters = () => {
      currentPage.value = 1
      loadReleases()
    }
    const resetFilters = () => {
      filterApp.value = ''
      filterStatus.value = ''
      filterNamespace.value = ''
      filterTimeRange.value = ''
      filterCreator.value = ''
      currentPage.value = 1
      loadReleases()
    }

    // 批量选择
    const selectedIds = ref([])
    const batchLoading = ref(false)
    const syncing = ref(false)
    const isAllSelected = computed(() => {
      if (releases.value.length === 0) return false
      return releases.value.every(r => selectedIds.value.includes(r.id))
    })
    const isIndeterminate = computed(() => {
      if (releases.value.length === 0) return false
      const selected = releases.value.filter(r => selectedIds.value.includes(r.id))
      return selected.length > 0 && selected.length < releases.value.length
    })
    const toggleAll = (e) => {
      if (e.target.checked) {
        const ids = releases.value.map(r => r.id)
        selectedIds.value = [...new Set([...selectedIds.value, ...ids])]
      } else {
        const ids = releases.value.map(r => r.id)
        selectedIds.value = selectedIds.value.filter(id => !ids.includes(id))
      }
    }
    const toggleSelect = (id) => {
      const idx = selectedIds.value.indexOf(id)
      if (idx === -1) selectedIds.value.push(id)
      else selectedIds.value.splice(idx, 1)
    }

    // 后端动态统计数据
    const statsData = ref({ total: 0, deploying: 0, success: 0, failed: 0, rollback: 0 })

    const normalizeStatus = (status) => {
      const map = { Pending: 'pending', AwaitingApproval: 'awaiting', Queued: 'deploying', Running: 'deploying', Succeeded: 'success', Failed: 'failed', Canceled: 'failed', Rollback: 'rollback' }
      return map[status] || status
    }

    const setFilter = (s) => { statusFilter.value = statusFilter.value === s ? '' : s }

    const totalPages = computed(() => Math.ceil(total.value / pageSizeRef.value) || 1)

    // 智能页码显示
    const visiblePages = computed(() => {
      const tp = totalPages.value
      const current = currentPage.value
      const pages = []
      if (tp <= 7) {
        for (let i = 1; i <= tp; i++) pages.push(i)
      } else {
        if (current <= 4) {
          for (let i = 1; i <= 5; i++) pages.push(i)
          pages.push('...')
          pages.push(tp)
        } else if (current >= tp - 3) {
          pages.push(1)
          pages.push('...')
          for (let i = tp - 4; i <= tp; i++) pages.push(i)
        } else {
          pages.push(1)
          pages.push('...')
          for (let i = current - 1; i <= current + 1; i++) pages.push(i)
          pages.push('...')
          pages.push(tp)
        }
      }
      return pages
    })

    const goToPage = (page) => {
      if (page < 1 || page > totalPages.value) return
      currentPage.value = page
      jumpPage.value = page
    }

    const jumpToPage = () => {
      const page = parseInt(jumpPage.value)
      if (page >= 1 && page <= totalPages.value) {
        currentPage.value = page
      }
    }

    const onPageSizeChange = () => {
      currentPage.value = 1
      loadReleases()
    }

    // 加载统计
    const loadStats = async () => {
      try {
        const res = await getReleaseStats()
        if (res.code === 0 && res.data?.stats) {
          const s = res.data.stats
          statsData.value = {
            total: s.total || 0,
            deploying: (s.Running || 0) + (s.Queued || 0) + (s.Pending || 0) + (s.AwaitingApproval || 0),
            success: s.Succeeded || 0,
            failed: (s.Failed || 0) + (s.Canceled || 0),
            rollback: s.Rollback || 0
          }
        }
      } catch (e) {
        console.error('加载统计失败:', e)
      }
    }

    // 加载列表
    const loadReleases = async () => {
      loading.value = true
      try {
        const statusMap = { deploying: 'Running', success: 'Succeeded', failed: 'Failed', rollback: 'Rollback', pending: 'Pending' }
        // 优先使用高级筛选面板的状态，其次使用顶部卡片的状态筛选
        const backendStatus = filterStatus.value || (statusFilter.value ? statusMap[statusFilter.value] : undefined)
        const params = {
          page: currentPage.value,
          page_size: pageSizeRef.value,
          keyword: searchKeyword.value || undefined,
          status: backendStatus || undefined,
          app_name: filterApp.value || undefined,
          namespace: filterNamespace.value || undefined,
          creator: filterCreator.value || undefined,
          time_range: filterTimeRange.value || undefined
        }
        const response = await getReleases(params)
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

    const loadAll = () => { Promise.all([loadReleases(), loadStats()]) }

    // 流水线
    const pipelines = ref([])
    const loadPipelines = async () => {
      try {
        const r = await getPipelines()
        if (r.code === 0) pipelines.value = r.data.list || r.data || []
      } catch (e) { console.error('加载流水线失败:', e) }
    }

    // 集群列表（用于手动发布时选择目标集群）
    const clusters = ref([])
    const loadClusters = async () => {
      try {
        const r = await getK8sClusterList()
        if (r.code === 0) clusters.value = r.data?.list || r.data || []
      } catch (e) { console.error('加载集群列表失败:', e) }
    }

    // 创建
    const showCreateDialog = ref(false)
    const creating = ref(false)
    const createForm = ref({ pipeline_id: '', name: '', version: '', namespace: 'production', image: '', remark: '', workload_kind: 'Deployment', workload_name: '', container_name: '', cluster_id: '' })
    const selectedPipelineInfo = ref(null)

    // 选择应用后自动显示部署目标信息
    const onPipelineSelect = () => {
      const pid = createForm.value.pipeline_id
      if (!pid) {
        selectedPipelineInfo.value = null
        return
      }
      const p = pipelines.value.find(item => item.id == pid)
      if (p) {
        selectedPipelineInfo.value = {
          namespace: p.target_namespace || '',
          workload: p.target_workload_name ? `${p.target_workload_kind || 'Deployment'}/${p.target_workload_name}` : '',
          container: p.target_container || '',
          cluster: p.target_cluster_id || '',
          image_repo: (p.env_vars || []).find(e => e.name === 'IMAGE_REPO')?.value || ''
        }
      } else {
        selectedPipelineInfo.value = null
      }
    }

    const handleCreate = async () => {
      if (!createForm.value.version) {
        Message.warning({ content: '请填写版本号/镜像标签' }); return
      }
      if (!createForm.value.pipeline_id && !createForm.value.image) {
        Message.warning({ content: '请选择应用或填写镜像地址' }); return
      }
      if (!createForm.value.pipeline_id && !createForm.value.name) {
        Message.warning({ content: '请填写应用名称' }); return
      }
      if (!createForm.value.pipeline_id && !createForm.value.cluster_id) {
        Message.warning({ content: '请选择目标集群' }); return
      }
      creating.value = true
      try {
        // 自动生成应用名称
        const selectedPipeline = pipelines.value.find(p => p.id == createForm.value.pipeline_id)
        const appName = selectedPipeline ? selectedPipeline.name : createForm.value.name

        // 解析完整镜像地址为 image_repo + image_tag
        // 例：harbor.example.com/proj/app:v1.0.0 → repo=harbor.example.com/proj/app, tag=v1.0.0
        let imageRepo = ''
        let imageTag = createForm.value.version
        if (createForm.value.image) {
          const lastColon = createForm.value.image.lastIndexOf(':')
          if (lastColon > 0 && !createForm.value.image.substring(lastColon + 1).includes('/')) {
            imageRepo = createForm.value.image.substring(0, lastColon)
            imageTag = createForm.value.image.substring(lastColon + 1) || createForm.value.version
          } else {
            imageRepo = createForm.value.image
          }
        }

        // 构建符合后端 CicdReleaseCreateRequest 的 payload
        const payload = {
          pipeline_id: createForm.value.pipeline_id ? Number(createForm.value.pipeline_id) : undefined,
          app_name: appName,
          image_tag: imageTag,
          image_repo: imageRepo || (selectedPipelineInfo.value ? selectedPipelineInfo.value.image_repo : '') || undefined,
          namespace: createForm.value.namespace || 'production',
          message: createForm.value.remark || ''
        }
        // 新应用手动发布时传入部署目标
        if (!createForm.value.pipeline_id) {
          payload.image_repo = imageRepo
          payload.workload_kind = createForm.value.workload_kind || 'Deployment'
          payload.workload_name = createForm.value.workload_name || appName
          payload.container_name = createForm.value.container_name || ''
          payload.cluster_ids = [Number(createForm.value.cluster_id)]
        }
        const r = await createRelease(payload)
        if (r.code === 0) {
          Message.success({ content: '发布单已创建，正在跳转构建界面...' }); showCreateDialog.value = false
          createForm.value = { pipeline_id: '', name: '', version: '', namespace: 'production', image: '', remark: '', workload_kind: 'Deployment', workload_name: '', container_name: '', cluster_id: '' }
          selectedPipelineInfo.value = null
          // 创建发布后跳转到关联流水线的执行阶段页面，实时查看构建进度
          if (payload.pipeline_id) {
            router.push(`/cicd/pipelines/${payload.pipeline_id}?tab=stages`)
          } else {
            router.push('/cicd/approvals')
          }
        } else { throw new Error(r.msg || '创建失败') }
      } catch (e) { Message.error({ content: e.message || '创建发布单失败' }) }
      finally { creating.value = false }
    }

    // 确认弹窗
    const showConfirmDialog = ref(false)
    const confirmTitle = ref('')
    const confirmMessage = ref('')
    const confirmBtnText = ref('确认')
    const confirmType = ref('warning')
    const confirming = ref(false)
    const pendingAction = ref(null)
    const openConfirm = (title, message, btnText, type, action) => {
      confirmTitle.value = title; confirmMessage.value = message; confirmBtnText.value = btnText; confirmType.value = type; pendingAction.value = action; showConfirmDialog.value = true
    }
    const confirmAction = async () => {
      if (pendingAction.value) {
        confirming.value = true
        try { await pendingAction.value() } finally { confirming.value = false; showConfirmDialog.value = false }
      }
    }

    const viewRelease = (rel) => { Message.info({ content: `查看发布: ${rel.app_name || rel.name}` }) }

    // 编辑功能
    const showEditDialog = ref(false)
    const editing = ref(false)
    const editForm = ref({ id: 0, app_name: '', namespace: '', workload_kind: 'Deployment', workload_name: '', container_name: '', image_repo: '', image_tag: '', strategy: 'rolling', timeout_sec: 300, message: '' })
    const canEdit = (status) => ['Pending', 'Failed', 'Canceled'].includes(status)
    const canDelete = (status) => !['Running', 'Queued'].includes(status)

    const editRelease = (rel) => {
      editForm.value = {
        id: rel.id,
        app_name: rel.app_name || '',
        namespace: rel.namespace || 'default',
        workload_kind: rel.workload_kind || 'Deployment',
        workload_name: rel.workload_name || '',
        container_name: rel.container_name || '',
        image_repo: rel.image_repo || '',
        image_tag: rel.image_tag || '',
        strategy: rel.strategy || 'rolling',
        timeout_sec: rel.timeout_sec || 300,
        message: rel.message || ''
      }
      showEditDialog.value = true
    }
    const handleEdit = async () => {
      if (!editForm.value.app_name) {
        Message.warning({ content: '请填写应用名称' }); return
      }
      editing.value = true
      try {
        const r = await updateReleaseApi(editForm.value)
        if (r.code === 0) {
          Message.success({ content: '发布单更新成功' }); showEditDialog.value = false; loadAll()
        } else { throw new Error(r.msg || '更新失败') }
      } catch (e) { Message.error({ content: e.message || '更新发布单失败' }) }
      finally { editing.value = false }
    }

    // 删除功能
    const deleteRelease = (rel) => {
      openConfirm('删除发布单', `确定要删除发布单 "${rel.app_name || rel.name}" 吗？此操作不可恢复。`, '确认删除', 'danger', async () => {
        const r = await deleteReleaseApi(rel.id)
        if (r.code === 0) { Message.success({ content: '发布单已删除' }); loadAll() } else { throw new Error(r.msg || '删除失败') }
      })
    }

    const cancelRelease = (rel) => {
      openConfirm('取消发布', `确定要取消发布 "${rel.app_name || rel.name}" 吗？`, '取消发布', 'warning', async () => {
        const r = await cancelReleaseApi(rel.id)
        if (r.code === 0) { Message.success({ content: '发布已取消' }); loadAll() } else { throw new Error(r.msg || '取消失败') }
      })
    }
    const rollbackRelease = (rel) => {
      const isFailed = rel.status === 'Failed'
      const msg = isFailed
        ? `发布单 "${rel.app_name || rel.name}" 部署失败，确定回滚吗？将工作负载恢复到本次部署前的镜像。`
        : `确定要回滚 "${rel.app_name || rel.name}" 吗？将恢复到上一个稳定版本。`
      openConfirm('回滚发布', msg, '确认回滚', 'warning', async () => {
        const r = await rollbackReleaseApi(rel.id)
        if (r.code === 0) { Message.success({ content: '回滚已提交' }); loadAll() } else { throw new Error(r.msg || '回滚失败') }
      })
    }
    const retryRelease = (rel) => {
      openConfirm('重试发布', `确定要重新发布 "${rel.app_name || rel.name}" 吗？`, '重新发布', 'create', async () => {
        const r = await retryReleaseApi(rel.id)
        if (r.code === 0) { Message.success({ content: '已重新开始发布' }); loadAll() } else { throw new Error(r.msg || '重试失败') }
      })
    }

    // 搜索
    let searchTimer = null
    const handleSearch = () => {
      if (searchTimer) clearTimeout(searchTimer)
      searchTimer = setTimeout(() => { currentPage.value = 1; loadReleases() }, 300)
    }
    const clearSearch = () => { searchKeyword.value = ''; currentPage.value = 1; loadReleases() }

    // 批量发布
    const showBatchPublishDialog = ref(false)
    const batchPublishItems = ref([])

    const handleBatchRetry = () => {
      const selected = releases.value.filter(r => selectedIds.value.includes(r.id) && ['Failed', 'Canceled', 'Succeeded'].includes(r.status))
      if (selected.length === 0) {
        Message.warning({ content: '请选择可发布的记录（失败/已取消/已成功状态）' }); return
      }
      batchPublishItems.value = selected.map(r => ({
        name: r.app_name || r.workload_name || `#${r.id}`,
        currentImage: r.image_tag || r.image || '—',
        targetImage: '重新构建',
        id: r.id
      }))
      showBatchPublishDialog.value = true
    }

    const confirmBatchPublish = async () => {
      const ids = batchPublishItems.value.map(i => i.id)
      batchLoading.value = true
      showBatchPublishDialog.value = false
      try {
        const r = await batchRetryReleaseApi(ids)
        if (r.code === 0) {
          const data = r.data || {}
          Message.success({ content: data.message || `批量发布完成`, duration: 3000 })
          selectedIds.value = []
          loadAll()
        } else { throw new Error(r.msg || '批量发布失败') }
      } catch (e) { Message.error({ content: e.message || '批量发布失败' }) }
      finally { batchLoading.value = false }
    }

    // 批量回滚
    const showBatchRollbackDialog = ref(false)
    const batchRollbackItems = ref([])

    const handleBatchRollback = () => {
      const selected = releases.value.filter(r => selectedIds.value.includes(r.id) && ['Succeeded', 'Running'].includes(r.status))
      if (selected.length === 0) {
        const statuses = [...new Set(releases.value.filter(r => selectedIds.value.includes(r.id)).map(r => r.status))]
        Message.warning({ content: `选中的发布单状态为 [${statuses.join(", ")}]，仅 Succeeded/Running 状态可回滚` }); return
      }
      batchRollbackItems.value = selected.map(r => ({
        name: r.app_name || r.workload_name || `#${r.id}`,
        currentImage: r.image_tag || r.image || r.current_image || '—',
        targetImage: r.prev_image || '上一版本',
        id: r.id
      }))
      showBatchRollbackDialog.value = true
    }

    const confirmBatchRollback = async () => {
      const ids = batchRollbackItems.value.map(i => i.id)
      batchLoading.value = true
      showBatchRollbackDialog.value = false
      try {
        const r = await batchRollbackReleaseApi(ids)
        if (r.code === 0) {
          const data = r.data || {}
          Message.success({ content: data.message || `批量回滚完成`, duration: 3000 })
          selectedIds.value = []
          loadAll()
        } else { throw new Error(r.msg || '批量回滚失败') }
      } catch (e) { Message.error({ content: e.message || '批量回滚失败' }) }
      finally { batchLoading.value = false }
    }

    // 批量取消
    const handleBatchCancel = () => {
      const ids = selectedIds.value.filter(id => {
        const rel = releases.value.find(r => r.id === id)
        return rel && ['Pending', 'Queued', 'Running', 'Succeeded'].includes(rel.status)
      })
      if (ids.length === 0) {
        Message.warning({ content: '请选择可取消的记录（等待中/排队中/运行中/已成功状态）' }); return
      }
      openConfirm('批量取消', `确定要取消已选的 ${ids.length} 个发布单吗？\n已部署成功/运行中的会触发回滚，未部署的直接取消。`, '确认取消', 'warning', async () => {
        batchLoading.value = true
        try {
          const r = await batchCancelReleaseApi(ids)
          if (r.code === 0) {
            const data = r.data || {}
            Message.success({ content: data.message || `批量取消完成`, duration: 3000 })
            selectedIds.value = []
            loadAll()
          } else { throw new Error(r.msg || '批量取消失败') }
        } catch (e) { Message.error({ content: e.message || '批量取消失败' }) }
        finally { batchLoading.value = false }
      })
    }

    // 同步流水线记录
    const syncFromPipeline = async () => {
      syncing.value = true
      try {
        const r = await syncReleasesApi()
        if (r.code === 0) {
          const data = r.data || {}
          if (data.synced > 0) {
            Message.success({ content: data.message || `同步完成：新增 ${data.synced} 条记录`, duration: 3000 })
            loadAll()
          } else {
            Message.info({ content: '所有流水线记录已同步，无新数据' })
          }
        } else { throw new Error(r.msg || '同步失败') }
      } catch (e) { Message.error({ content: e.message || '同步流水线记录失败' }) }
      finally { syncing.value = false }
    }

    watch(statusFilter, () => { currentPage.value = 1; loadReleases() })
    watch(currentPage, () => { loadReleases() })

    const statusText = (status) => {
      const map = { deploying: '部署中', success: '发布成功', failed: '发布失败', rollback: '已回滚', pending: '等待中', Pending: '等待中', AwaitingApproval: '待审批', Queued: '排队中', Running: '部署中', Succeeded: '发布成功', Failed: '发布失败', Canceled: '已取消', Rollback: '已回滚', awaiting: '待审批' }
      return map[status] || status
    }
    const strategyText = (s) => ({ rolling: '滚动更新', recreate: '重建', canary: '金丝雀', bluegreen: '蓝绿部署' })[s] || s
    const formatImage = (rel) => {
      const repo = rel.image_repo || '', tag = rel.image_tag || ''
      if (!repo && !tag) return '-'
      if (repo.includes(':')) return repo
      if (repo && !tag) return repo
      const parts = repo.split('/')
      const short = parts.length > 2 ? '.../' + parts.slice(-2).join('/') : repo
      return `${short}:${tag}`
    }
    const getFullImage = (rel) => {
      const repo = rel.image_repo || '', tag = rel.image_tag || ''
      if (!repo) return '-'
      if (repo.includes(':')) return repo
      return tag ? `${repo}:${tag}` : repo
    }
    const formatDate = (ts) => {
      if (!ts) return '-'
      const t = ts > 1e11 ? ts : ts * 1000
      const d = new Date(t), now = new Date(), diff = now - d
      if (diff < 0) return d.toLocaleDateString('zh-CN')
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    }

    onMounted(() => {
      loadAll(); loadPipelines(); loadClusters()
      // 支持从流水线列表直接跳转创建发布单（预选流水线）
      const pipelineId = router.currentRoute.value.query.pipeline_id
      const action = router.currentRoute.value.query.action
      if (pipelineId && action === 'create') {
        // 等待 pipelines 加载后预选
        const unwatch = watch(pipelines, (val) => {
          if (val.length > 0) {
            createForm.value.pipeline_id = Number(pipelineId)
            showCreateDialog.value = true
            unwatch()
          }
        })
      }
    })

    return {
      loading, releases, searchKeyword, searchFocused, statusFilter, releaseViewMode, currentPage, totalPages, total,
      statsData, setFilter, pipelines, clusters, showCreateDialog, creating, createForm, handleCreate,
      selectedPipelineInfo, onPipelineSelect,
      showEditDialog, editing, editForm, editRelease, handleEdit, canEdit, canDelete, deleteRelease,
      showConfirmDialog, confirmTitle, confirmMessage, confirmBtnText, confirmType, confirming, confirmAction,
      viewRelease, cancelRelease, rollbackRelease, retryRelease, handleSearch, clearSearch,
      statusText, normalizeStatus, strategyText, formatImage, getFullImage, formatDate, loadAll,
      visiblePages, goToPage, jumpPage, jumpToPage, pageSizeRef, onPageSizeChange,
      selectedIds, batchLoading, isAllSelected, isIndeterminate, toggleAll, toggleSelect,
      handleBatchRetry, handleBatchRollback, handleBatchCancel,
      syncing, syncFromPipeline,
      showFilterPanel, filterApp, filterStatus, filterNamespace, filterTimeRange, filterCreator,
      namespaceOptions, creatorOptions, activeFilterCount, applyFilters, resetFilters
    }
  }
}
</script>

<style scoped>
.releases-page { min-height: 100vh; background: #f4f6f9; }

/* ---- Banner ---- */
.page-banner {
  background: linear-gradient(135deg, #1a2332 0%, #2d3e50 50%, #34495e 100%);
  padding: 28px 32px;
  position: relative; overflow: hidden;
}
.page-banner::before {
  content: ''; position: absolute; top: -50%; right: -8%; width: 380px; height: 380px; border-radius: 50%;
  background: radial-gradient(circle, rgba(78,124,246,0.12) 0%, transparent 70%); pointer-events: none;
}
.banner-inner { display: flex; align-items: center; justify-content: space-between; max-width: 100%; margin: 0 auto; position: relative; z-index: 1; }
.banner-left { display: flex; align-items: center; gap: 16px; }
.banner-icon {
  width: 48px; height: 48px; background: rgba(255,255,255,0.1); border-radius: 12px;
  display: flex; align-items: center; justify-content: center; border: 1px solid rgba(255,255,255,0.08);
}
.banner-icon svg { width: 26px; height: 26px; color: #67d5b5; }
.banner-title { margin: 0; font-size: 22px; font-weight: 600; color: #fff; letter-spacing: 0.5px; }
.banner-desc { margin: 4px 0 0; font-size: 13px; color: rgba(255,255,255,0.55); }
.banner-hint { margin: 4px 0 0; font-size: 12px; color: rgba(255,255,255,0.4); }
.banner-inline-link { color: #7dd3fc; text-decoration: none; margin-left: 2px; }
.banner-inline-link:hover { color: #bae6fd; text-decoration: underline; }
.banner-actions { display: flex; gap: 10px; }
.btn-banner-refresh, .btn-banner-create, .btn-banner-sync, .btn-banner-new-app {
  display: flex; align-items: center; gap: 6px; padding: 9px 18px; border-radius: 8px;
  font-size: 13px; cursor: pointer; transition: all 0.25s; border: 1px solid rgba(255,255,255,0.15);
}
.btn-banner-refresh { background: rgba(255,255,255,0.1); color: #fff; }
.btn-banner-refresh:hover { background: rgba(255,255,255,0.18); }
.btn-banner-refresh:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-banner-refresh svg, .btn-banner-create svg, .btn-banner-sync svg, .btn-banner-new-app svg { width: 16px; height: 16px; }
.btn-banner-new-app { background: rgba(255,255,255,0.12); color: #fff; border-color: rgba(255,255,255,0.25); }
.btn-banner-new-app:hover { background: rgba(255,255,255,0.22); transform: translateY(-1px); }
.btn-banner-create { background: linear-gradient(135deg, #4e7cf6, #3b5fe0); color: #fff; border-color: transparent; font-weight: 600; }
.btn-banner-create:hover { box-shadow: 0 4px 14px rgba(78,124,246,0.4); transform: translateY(-1px); }
.btn-banner-sync { background: linear-gradient(135deg, #10b981, #059669); color: #fff; border-color: transparent; font-weight: 600; }
.btn-banner-sync:hover { box-shadow: 0 4px 14px rgba(16,185,129,0.4); transform: translateY(-1px); }
.btn-banner-sync:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

/* ---- Onboarding ---- */
.onboard-cards {
  display: flex; flex-direction: column; align-items: center; width: 100%; max-width: 720px;
  padding: 40px 32px; background: linear-gradient(135deg, #f8faff 0%, #eef2ff 100%);
  border-radius: 16px; border: 1px solid #e0e7ff;
}
.onboard-hero { text-align: center; margin-bottom: 32px; }
.onboard-icon {
  width: 64px; height: 64px; margin: 0 auto 16px; border-radius: 16px;
  background: linear-gradient(135deg, #4e7cf6, #3b5fe0);
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8px 24px rgba(78,124,246,0.3);
}
.onboard-icon svg { width: 32px; height: 32px; color: #fff; }
.onboard-hero h3 { margin: 0 0 8px; font-size: 20px; font-weight: 700; color: #1e293b; }
.onboard-hero p { margin: 0; font-size: 14px; color: #64748b; }
.onboard-steps {
  display: flex; align-items: center; gap: 16px; margin-bottom: 32px;
  background: #fff; padding: 20px 24px; border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06); width: 100%;
}
.onboard-step { display: flex; align-items: flex-start; gap: 12px; flex: 1; }
.step-num {
  width: 28px; height: 28px; border-radius: 50%; background: linear-gradient(135deg, #4e7cf6, #3b5fe0);
  color: #fff; font-size: 13px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.step-content .step-title { font-size: 14px; font-weight: 600; color: #1e293b; }
.step-content .step-desc { font-size: 12px; color: #64748b; margin-top: 3px; }
.onboard-step-arrow { font-size: 20px; color: #cbd5e1; flex-shrink: 0; align-self: center; }
.onboard-actions { display: flex; gap: 12px; }
.btn-onboard-primary {
  display: flex; align-items: center; gap: 8px; padding: 12px 28px; border-radius: 10px;
  background: linear-gradient(135deg, #4e7cf6, #3b5fe0); color: #fff; border: none;
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.25s;
  box-shadow: 0 4px 14px rgba(78,124,246,0.35);
}
.btn-onboard-primary:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(78,124,246,0.45); }
.btn-onboard-primary svg { width: 16px; height: 16px; }
.btn-onboard-secondary {
  display: flex; align-items: center; gap: 8px; padding: 12px 24px; border-radius: 10px;
  background: #fff; color: #64748b; border: 1.5px solid #e2e8f0;
  font-size: 14px; cursor: pointer; transition: all 0.25s;
}
.btn-onboard-secondary:hover { border-color: #4e7cf6; color: #4e7cf6; }
.btn-onboard-secondary svg { width: 16px; height: 16px; }

/* ---- Modal: No-app guide ---- */
.no-app-guide {
  display: flex; align-items: center; gap: 12px; padding: 14px 16px;
  background: linear-gradient(135deg, #f0f7ff, #eef2ff); border-radius: 10px;
  border: 1px solid #c7d7fd; margin-bottom: 4px;
}
.guide-icon { color: #4e7cf6; flex-shrink: 0; }
.guide-text { flex: 1; }
.guide-title { font-size: 14px; font-weight: 600; color: #1e293b; margin-bottom: 3px; }
.guide-desc { font-size: 12px; color: #64748b; }
.btn-guide-create {
  display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 8px;
  background: #4e7cf6; color: #fff; border: none; font-size: 13px; font-weight: 600;
  cursor: pointer; white-space: nowrap; transition: all 0.2s; flex-shrink: 0;
}
.btn-guide-create:hover { background: #3b5fe0; transform: translateY(-1px); }
.btn-guide-create svg { width: 14px; height: 14px; }
.guide-divider {
  display: flex; align-items: center; gap: 12px; margin: 10px 0;
  font-size: 12px; color: #94a3b8;
}
.guide-divider::before, .guide-divider::after {
  content: ''; flex: 1; height: 1px; background: #e2e8f0;
}
.guide-divider span { white-space: nowrap; }

/* ---- Metrics ---- */
.metrics-row {
  display: grid; grid-template-columns: repeat(5, 1fr); gap: 14px;
  padding: 20px 32px 0; max-width: 100%; margin: -18px auto 0; position: relative; z-index: 2;
}
.metric-card {
  display: flex; align-items: center; gap: 12px; padding: 16px 18px;
  background: #fff; border-radius: 10px; cursor: pointer; transition: all 0.25s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04);
  border: 2px solid transparent; position: relative;
}
.metric-card:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,0.1); }
.metric-card.active { border-color: #4e7cf6; box-shadow: 0 2px 12px rgba(78,124,246,0.15); }
.metric-icon-wrap { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.metric-icon-wrap svg { width: 20px; height: 20px; }
.metric-icon-wrap.total    { background: #eef2ff; color: #4e7cf6; }
.metric-icon-wrap.deploying { background: #fff8e1; color: #f59e0b; }
.metric-icon-wrap.success  { background: #ecfdf5; color: #10b981; }
.metric-icon-wrap.failed   { background: #fef2f2; color: #ef4444; }
.metric-icon-wrap.rollback { background: #f5f3ff; color: #8b5cf6; }
.metric-body { display: flex; flex-direction: column; flex: 1; }
.metric-num { font-size: 24px; font-weight: 700; color: #1e293b; line-height: 1.2; font-variant-numeric: tabular-nums; }
.metric-label { font-size: 11px; color: #94a3b8; margin-top: 2px; font-weight: 500; text-transform: uppercase; letter-spacing: 0.5px; }
.metric-badge {
  position: absolute; top: 8px; right: 8px; font-size: 9px; font-weight: 700; padding: 2px 6px;
  border-radius: 4px; letter-spacing: 1px; animation: pulse 2s ease-in-out infinite;
}
.metric-badge.deploying { background: #fff8e1; color: #f59e0b; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.5; } }

/* ---- Content ---- */
.content-area { padding: 20px 32px 32px; max-width: 100%; margin: 0 auto; }
.toolbar { display: flex; align-items: center; justify-content: space-between; padding: 14px 0; }
.toolbar-left { display: flex; align-items: center; gap: 12px; }
.section-title { margin: 0; font-size: 16px; font-weight: 600; color: #1e293b; }
.record-badge { font-size: 12px; color: #94a3b8; background: #f1f5f9; padding: 3px 10px; border-radius: 10px; font-weight: 500; }
.search-box {
  display: flex; align-items: center; gap: 8px; padding: 7px 14px;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 8px; transition: all 0.2s;
}
.search-box.focused { border-color: #4e7cf6; box-shadow: 0 0 0 3px rgba(78,124,246,0.1); }
.search-box svg { width: 16px; height: 16px; color: #94a3b8; flex-shrink: 0; }
.search-box input { border: none; outline: none; font-size: 13px; color: #334155; width: 220px; background: transparent; }
.search-box input::placeholder { color: #cbd5e1; }
.clear-btn { background: none; border: none; cursor: pointer; padding: 2px; color: #94a3b8; display: flex; }
.clear-btn:hover { color: #ef4444; }
.clear-btn svg { width: 14px; height: 14px; }

/* ---- Filter Toggle Button ---- */
.filter-toggle-btn {
  display: flex; align-items: center; gap: 6px; padding: 7px 14px;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 13px; color: #475569; cursor: pointer; transition: all 0.2s; font-weight: 500;
}
.filter-toggle-btn:hover { border-color: #4e7cf6; color: #4e7cf6; background: #f8faff; }
.filter-toggle-btn.active { border-color: #4e7cf6; color: #4e7cf6; background: #eef2ff; }
.filter-toggle-btn svg { width: 14px; height: 14px; }
.filter-count {
  background: #4e7cf6; color: #fff; font-size: 10px; font-weight: 700;
  width: 18px; height: 18px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
}

/* ---- Filter Panel (Advanced) ---- */
.filter-panel {
  background: #fff; border: 1px solid #e5e9f2; border-radius: 10px;
  padding: 16px 20px; margin-bottom: 14px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.02);
}
.filter-row {
  display: flex; align-items: flex-end; gap: 14px; flex-wrap: wrap;
}
.filter-item { display: flex; flex-direction: column; gap: 4px; min-width: 150px; flex: 1; }
.filter-label {
  font-size: 11px; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.5px;
}
.filter-select {
  padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 7px;
  font-size: 13px; color: #334155; background: #f8fafc; cursor: pointer;
  transition: all 0.2s; outline: none; appearance: auto;
}
.filter-select:hover { border-color: #94a3b8; }
.filter-select:focus { border-color: #4e7cf6; box-shadow: 0 0 0 3px rgba(78,124,246,0.1); background: #fff; }
.filter-actions { display: flex; align-items: flex-end; }
.filter-reset-btn {
  display: flex; align-items: center; gap: 5px; padding: 8px 14px;
  border: 1px solid #e2e8f0; border-radius: 7px; background: #fff;
  font-size: 12px; color: #64748b; cursor: pointer; transition: all 0.2s;
}
.filter-reset-btn:hover:not(:disabled) { border-color: #ef4444; color: #ef4444; background: #fef2f2; }
.filter-reset-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.filter-reset-btn svg { width: 13px; height: 13px; }

/* Filter slide transition */
.filter-slide-enter-active, .filter-slide-leave-active { transition: all 0.25s ease; }
.filter-slide-enter-from, .filter-slide-leave-to { opacity: 0; transform: translateY(-8px); max-height: 0; margin-bottom: 0; padding: 0; overflow: hidden; }
.filter-slide-enter-to, .filter-slide-leave-from { max-height: 200px; }

/* ---- Batch Toolbar (Full-width prominent) ---- */
/* === 新版批量操作工具栏 === */
.batch-toolbar-v2 {
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
  border: 1px solid #475569;
  border-radius: 12px;
  padding: 16px 20px;
  margin: 12px 0;
  box-shadow: 0 8px 32px rgba(0,0,0,0.15);
  animation: batchSlideIn 0.2s ease;
}
@keyframes batchSlideIn { from { opacity: 0; transform: translateY(-6px); } to { opacity: 1; transform: translateY(0); } }
.batch-slide-enter-active, .batch-slide-leave-active { transition: all 0.2s ease; }
.batch-slide-enter-from, .batch-slide-leave-to { opacity: 0; transform: translateY(-6px); }

.batch-top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.batch-count { font-size: 15px; color: #e2e8f0; font-weight: 500; }
.batch-count strong { color: #60a5fa; font-size: 18px; }
.batch-actions { display: flex; gap: 8px; }
.batch-btn {
  padding: 7px 16px; border: none; border-radius: 7px;
  font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.15s;
}
.batch-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.batch-btn.publish { background: #10b981; color: #fff; }
.batch-btn.publish:hover:not(:disabled) { background: #059669; }
.batch-btn.rollback { background: #f59e0b; color: #fff; }
.batch-btn.rollback:hover:not(:disabled) { background: #d97706; }
.batch-btn.stop { background: #ef4444; color: #fff; }
.batch-btn.stop:hover:not(:disabled) { background: #dc2626; }
.batch-btn.clear { background: rgba(255,255,255,0.08); color: #94a3b8; border: 1px solid rgba(255,255,255,0.12); }
.batch-btn.clear:hover { background: rgba(255,255,255,0.14); color: #e2e8f0; }

.batch-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.batch-chip {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 8px 4px 12px; border-radius: 18px;
  background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1);
  font-size: 12px; transition: all 0.15s;
}
.batch-chip:hover { background: rgba(255,255,255,0.1); }
.chip-name { color: #e2e8f0; font-weight: 500; max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chip-status { padding: 1px 6px; border-radius: 8px; font-size: 10px; font-weight: 600; }
.chip-status.succeeded, .chip-status.success { background: rgba(16,185,129,0.2); color: #34d399; }
.chip-status.running { background: rgba(59,130,246,0.2); color: #60a5fa; }
.chip-status.failed { background: rgba(239,68,68,0.2); color: #f87171; }
.chip-status.awaiting_approval { background: rgba(245,158,11,0.15); color: #fbbf24; }
.chip-remove {
  width: 16px; height: 16px; border-radius: 50%; border: none;
  background: rgba(255,255,255,0.1); color: #94a3b8;
  cursor: pointer; font-size: 10px; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.chip-remove:hover { background: #ef4444; color: #fff; }

/* ---- Checkbox ---- */
.row-checkbox { width: 16px; height: 16px; cursor: pointer; accent-color: #4e7cf6; }
.data-table tbody tr.row-selected { background: #f0f7ff !important; }
.data-table tbody tr.row-selected:hover { background: #e0efff !important; }

/* ---- Table ---- */
.table-wrapper {
  background: #fff; border-radius: 10px; overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04);
}
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table thead { background: #f8fafc; border-bottom: 1px solid #e2e8f0; }
.data-table th {
  padding: 11px 16px; text-align: left; font-weight: 600; color: #64748b;
  font-size: 11px; text-transform: uppercase; letter-spacing: 0.8px; white-space: nowrap;
}
.data-table td { padding: 14px 16px; border-bottom: 1px solid #f1f5f9; color: #334155; vertical-align: middle; }
.data-table tbody tr { transition: background 0.15s; }
.data-table tbody tr:hover { background: #f8fafc; }
.data-table tbody tr:last-child td { border-bottom: none; }
.data-table tbody tr.row-deploying { border-left: 3px solid #f59e0b; }
.data-table tbody tr.row-success   { border-left: 3px solid #10b981; }
.data-table tbody tr.row-failed    { border-left: 3px solid #ef4444; }
.data-table tbody tr.row-rollback  { border-left: 3px solid #8b5cf6; }
.data-table tbody tr.row-pending   { border-left: 3px solid #cbd5e1; }

.app-cell { display: flex; align-items: center; gap: 10px; }
.app-avatar {
  width: 34px; height: 34px; border-radius: 8px; font-size: 14px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0;
}
.app-avatar.deploying { background: linear-gradient(135deg, #f59e0b, #d97706); }
.app-avatar.success   { background: linear-gradient(135deg, #10b981, #059669); }
.app-avatar.failed    { background: linear-gradient(135deg, #ef4444, #dc2626); }
.app-avatar.rollback  { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.app-avatar.pending   { background: linear-gradient(135deg, #94a3b8, #64748b); }
.app-info { display: flex; flex-direction: column; }
.app-name { font-weight: 600; color: #1e293b; }
.app-id { font-size: 11px; color: #94a3b8; }

.status-pill {
  display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px;
  border-radius: 6px; font-size: 12px; font-weight: 600; white-space: nowrap;
}
.status-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.status-pill.deploying  { background: #fffbeb; color: #d97706; }
.status-pill.deploying .status-dot { background: #f59e0b; box-shadow: 0 0 6px rgba(245,158,11,0.4); animation: pulse 2s infinite; }
.status-pill.success { background: #ecfdf5; color: #059669; }
.status-pill.success .status-dot { background: #10b981; }
.status-pill.failed  { background: #fef2f2; color: #dc2626; }
.status-pill.failed .status-dot  { background: #ef4444; }
.status-pill.rollback { background: #f5f3ff; color: #7c3aed; }
.status-pill.rollback .status-dot { background: #8b5cf6; }
.status-pill.pending { background: #f8fafc; color: #64748b; }
.status-pill.pending .status-dot { background: #cbd5e1; }
.status-pill.awaiting { background: #fff7ed; color: #c2410c; }
.status-pill.awaiting .status-dot { background: #f97316; box-shadow: 0 0 6px rgba(249,115,22,0.4); animation: pulse 2s infinite; }

.status-cell { display: flex; flex-direction: column; gap: 4px; }
.fail-reason {
  font-size: 11px; color: #dc2626; line-height: 1.4;
  max-width: 220px; overflow: hidden; text-overflow: ellipsis;
  white-space: nowrap; cursor: help;
  background: #fef2f2; padding: 2px 8px; border-radius: 4px;
  border-left: 2px solid #ef4444;
}

.workload-cell { display: flex; flex-direction: column; gap: 4px; }
.workload-tag { font-size: 12px; background: #f1f5f9; padding: 2px 8px; border-radius: 4px; color: #475569; font-family: 'SF Mono','Fira Code',monospace; }
.container-tag { font-size: 11px; color: #94a3b8; }
.image-code { font-size: 11px; background: #f1f5f9; padding: 3px 8px; border-radius: 4px; color: #475569; font-family: 'SF Mono',monospace; word-break: break-all; }
.ns-badge { font-size: 11px; background: #eff6ff; color: #2563eb; padding: 3px 8px; border-radius: 5px; font-weight: 600; }
.strategy-tag { font-size: 11px; background: #f1f5f9; color: #64748b; padding: 3px 8px; border-radius: 5px; }
.time-text { font-size: 12px; color: #64748b; white-space: nowrap; }
.text-muted { color: #cbd5e1; }

/* ---- New columns: version / cluster / creator ---- */
.version-tag {
  font-size: 12px; background: linear-gradient(135deg, #ecfdf5, #d1fae5); color: #065f46;
  padding: 3px 10px; border-radius: 5px; font-family: 'SF Mono','Fira Code',monospace;
  font-weight: 600; border: 1px solid #a7f3d0;
}
.cluster-badge {
  font-size: 11px; background: #fdf4ff; color: #86198f; padding: 3px 8px;
  border-radius: 5px; font-weight: 600; border: 1px solid #f0abfc;
}
.creator-cell { display: flex; align-items: center; gap: 6px; }
.creator-avatar {
  width: 24px; height: 24px; border-radius: 50%; font-size: 11px; font-weight: 700;
  display: inline-flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #e0e7ff, #c7d2fe); color: #4338ca;
}
.creator-name { font-size: 12px; color: #475569; font-weight: 500; }

.actions-cell { display: flex; gap: 4px; }
.act-btn {
  width: 30px; height: 30px; border: none; border-radius: 7px; display: inline-flex;
  align-items: center; justify-content: center; cursor: pointer; transition: all 0.2s;
}
.act-btn svg { width: 15px; height: 15px; }
.act-btn.view { background: #f1f5f9; color: #64748b; }
.act-btn.view:hover { background: #4e7cf6; color: #fff; }
.act-btn.cancel { background: #fff8e1; color: #f59e0b; }
.act-btn.cancel:hover { background: #f59e0b; color: #fff; }
.act-btn.rollback { background: #f5f3ff; color: #8b5cf6; }
.act-btn.rollback:hover { background: #8b5cf6; color: #fff; }
.act-btn.retry { background: #ecfdf5; color: #10b981; }
.act-btn.retry:hover { background: #10b981; color: #fff; }
.act-btn.edit { background: #eef2ff; color: #4e7cf6; }
.act-btn.edit:hover { background: #4e7cf6; color: #fff; }
.act-btn.delete { background: #fef2f2; color: #ef4444; }
.act-btn.delete:hover { background: #ef4444; color: #fff; }

/* ---- Loading / Empty ---- */
.loading-state, .empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 80px 20px; background: #fff; border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06); color: #94a3b8;
}
.loader { display: flex; gap: 8px; margin-bottom: 20px; }
.dot { width: 12px; height: 12px; border-radius: 50%; background: #4e7cf6; animation: bounce 1.4s ease-in-out infinite both; }
.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }
@keyframes bounce { 0%,80%,100% { transform: scale(0); opacity: 0.5; } 40% { transform: scale(1); opacity: 1; } }
.empty-svg svg { width: 160px; height: 130px; }
.empty-state h3 { margin: 16px 0 6px; font-size: 16px; font-weight: 600; color: #475569; }
.empty-state p { margin: 0; font-size: 13px; color: #94a3b8; }

/* ---- Pagination (Modern) ---- */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 14px 20px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  flex-wrap: wrap;
  gap: 14px;
}
.pagination-left { display: flex; align-items: center; }
.pagination-summary { font-size: 13px; color: #64748b; }
.pagination-summary strong { color: #1e293b; font-weight: 600; }
.pagination-center { display: flex; align-items: center; gap: 4px; }
.pagination-btn {
  min-width: 34px; height: 34px; border: 1px solid #e2e8f0; border-radius: 6px;
  background: #fff; color: #475569; font-size: 14px; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; transition: all 0.2s;
}
.pagination-btn:hover:not(:disabled) { border-color: #4e7cf6; color: #4e7cf6; background: #f0f5ff; }
.pagination-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pagination-btn.page-number.active { background: #4e7cf6; color: #fff; border-color: #4e7cf6; font-weight: 600; }
.pagination-ellipsis { color: #94a3b8; font-size: 14px; padding: 0 4px; }
.pagination-right { display: flex; align-items: center; gap: 8px; }
.page-size-select {
  padding: 6px 10px; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 12px; color: #475569; background: #fff; cursor: pointer;
}
.page-size-select:focus { outline: none; border-color: #4e7cf6; }
.pagination-goto { font-size: 12px; color: #64748b; }
.page-jump-input {
  width: 50px; padding: 5px 8px; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 12px; text-align: center; color: #475569;
}
.page-jump-input:focus { outline: none; border-color: #4e7cf6; }

/* ---- Modal ---- */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(15,23,42,0.55); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; z-index: 9999;
}
.modal-dialog {
  background: #fff; border-radius: 14px; width: 560px; max-width: 92%;
  box-shadow: 0 25px 60px rgba(0,0,0,0.2); overflow: hidden;
}
.modal-dialog.small { width: 440px; }
.modal-head {
  display: flex; align-items: center; gap: 12px; padding: 20px 24px; position: relative;
}
.modal-head.create  { background: linear-gradient(135deg, #eef2ff, #dbeafe); }
.modal-head.warning { background: linear-gradient(135deg, #fffbeb, #fef3c7); }
.modal-head-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
}
.modal-head.create .modal-head-icon  { background: #4e7cf6; color: #fff; }
.modal-head.warning .modal-head-icon { background: #f59e0b; color: #fff; }
.modal-head.danger  .modal-head-icon { background: #ef4444; color: #fff; }
.modal-head.danger { background: linear-gradient(135deg, #fef2f2, #fee2e2); }
.modal-head-icon svg { width: 22px; height: 22px; }
.modal-head h3 { margin: 0; font-size: 17px; font-weight: 600; color: #1e293b; flex: 1; }
.modal-close {
  background: none; border: none; cursor: pointer; padding: 4px; border-radius: 6px;
  color: #94a3b8; transition: all 0.2s;
}
.modal-close:hover { background: rgba(0,0,0,0.06); color: #475569; }
.modal-close svg { width: 20px; height: 20px; }
.modal-body { padding: 20px 24px; }
.confirm-msg { margin: 0; font-size: 14px; color: #475569; line-height: 1.6; }
.field { margin-bottom: 16px; }
.field label { display: block; font-size: 13px; font-weight: 600; color: #334155; margin-bottom: 6px; }
.optional { color: #94a3b8; font-weight: 400; }
.field input, .field select, .field textarea {
  width: 100%; padding: 9px 14px; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 13px; color: #334155; transition: all 0.2s; box-sizing: border-box; font-family: inherit;
}
.field input:focus, .field select:focus, .field textarea:focus {
  outline: none; border-color: #4e7cf6; box-shadow: 0 0 0 3px rgba(78,124,246,0.1);
}
.field textarea { resize: vertical; }
.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.modal-foot {
  display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px;
  background: #f8fafc; border-top: 1px solid #f1f5f9;
}
.btn-cancel {
  padding: 9px 20px; background: #fff; border: 1px solid #e2e8f0; border-radius: 8px;
  color: #64748b; font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s;
}
.btn-cancel:hover { background: #f1f5f9; color: #334155; }
.btn-confirm {
  padding: 9px 24px; border: none; border-radius: 8px; color: #fff;
  font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-confirm.create  { background: #4e7cf6; }
.btn-confirm.create:hover  { background: #3b5fe0; box-shadow: 0 2px 10px rgba(78,124,246,0.3); }
.btn-confirm.warning { background: #f59e0b; }
.btn-confirm.warning:hover { background: #d97706; box-shadow: 0 2px 10px rgba(245,158,11,0.3); }
.btn-confirm.danger { background: #ef4444; }
.btn-confirm.danger:hover { background: #dc2626; box-shadow: 0 2px 10px rgba(239,68,68,0.3); }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-enter-active { transition: all 0.3s ease; }
.modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal-dialog { transform: scale(0.95) translateY(10px); }
.modal-leave-to .modal-dialog { transform: scale(0.97); }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Pipeline Inherit Hint */
.pipeline-inherit-hint {
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  border-radius: 8px;
  padding: 12px 14px;
  margin-bottom: 16px;
}
.pipeline-inherit-hint .hint-title {
  font-size: 13px;
  font-weight: 600;
  color: #065f46;
  margin-bottom: 6px;
}
.pipeline-inherit-hint .hint-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
}
.pipeline-inherit-hint .hint-items span {
  font-size: 12px;
  color: #047857;
  background: rgba(16, 185, 129, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
}
.pipeline-inherit-hint .hint-items b {
  color: #064e3b;
}
.pipeline-inherit-hint .hint-note {
  font-size: 11px;
  color: #6b7280;
  margin-top: 8px;
  font-style: italic;
}
.pipeline-inherit-hint .hint-items.hint-warning {
  flex-direction: column;
  gap: 6px;
}
.pipeline-inherit-hint .hint-items.hint-warning span {
  background: rgba(245, 158, 11, 0.1);
  color: #92400e;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
}
.pipeline-inherit-hint .hint-link {
  font-size: 12px;
  color: #4e7cf6;
  cursor: pointer;
  text-decoration: underline;
  font-weight: 500;
}
.pipeline-inherit-hint .hint-link:hover {
  color: #3b63d4;
}

@media (max-width: 1200px) {
  .metrics-row { grid-template-columns: repeat(3, 1fr); }
  .release-cards-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .metrics-row { grid-template-columns: repeat(2, 1fr); }
  .release-cards-grid { grid-template-columns: 1fr; }
  .page-banner { padding: 20px; }
  .content-area { padding: 16px 20px; }
  .toolbar { flex-direction: column; gap: 12px; align-items: stretch; }
  .table-wrapper { overflow-x: auto; }
}

/* ---- View Toggle ---- */
.view-toggle { display: flex; border: 1px solid #e2e8f0; border-radius: 6px; overflow: hidden; margin-right: 8px; }
.toggle-btn {
  padding: 6px 10px; border: none; background: #fff; cursor: pointer;
  display: flex; align-items: center; transition: all 0.15s;
}
.toggle-btn svg { width: 15px; height: 15px; color: #94a3b8; }
.toggle-btn.active { background: #f1f5f9; }
.toggle-btn.active svg { color: #4e7cf6; }
.toggle-btn + .toggle-btn { border-left: 1px solid #e2e8f0; }

/* ---- Release Cards Grid ---- */
.release-cards-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.release-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
  overflow: hidden;
  transition: all 0.25s;
}
.release-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
.release-card.selected { border-color: #4e7cf6; box-shadow: 0 0 0 2px rgba(78,124,246,0.15); }
.release-card.card-deploying { border-left: 3px solid #f59e0b; }
.release-card.card-success   { border-left: 3px solid #10b981; }
.release-card.card-failed    { border-left: 3px solid #ef4444; }
.release-card.card-rollback  { border-left: 3px solid #8b5cf6; }
.release-card.card-pending   { border-left: 3px solid #cbd5e1; }
.release-card.card-awaiting  { border-left: 3px solid #f97316; }

/* Card Header */
.rc-header {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 14px 10px;
}
.rc-check { flex-shrink: 0; }
.rc-app-info { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.rc-avatar {
  width: 32px; height: 32px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 700; color: #fff; flex-shrink: 0;
}
.rc-avatar.deploying { background: linear-gradient(135deg, #f59e0b, #d97706); }
.rc-avatar.success   { background: linear-gradient(135deg, #10b981, #059669); }
.rc-avatar.failed    { background: linear-gradient(135deg, #ef4444, #dc2626); }
.rc-avatar.rollback  { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.rc-avatar.pending   { background: linear-gradient(135deg, #94a3b8, #64748b); }
.rc-avatar.awaiting  { background: linear-gradient(135deg, #f97316, #ea580c); }

.rc-app-meta { min-width: 0; }
.rc-app-name {
  display: block; font-size: 14px; font-weight: 600; color: #1e293b;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.rc-app-id { font-size: 11px; color: #94a3b8; }

.rc-status-badge {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 3px 9px; border-radius: 10px;
  font-size: 11px; font-weight: 600; flex-shrink: 0;
}
.rc-status-badge.deploying { background: #fffbeb; color: #d97706; }
.rc-status-badge.success { background: #ecfdf5; color: #059669; }
.rc-status-badge.failed { background: #fef2f2; color: #dc2626; }
.rc-status-badge.rollback { background: #f5f3ff; color: #7c3aed; }
.rc-status-badge.pending { background: #f8fafc; color: #64748b; }
.rc-status-badge.awaiting { background: #fff7ed; color: #c2410c; }
.rc-status-dot {
  width: 6px; height: 6px; border-radius: 50%; background: currentColor;
}
.rc-status-badge.deploying .rc-status-dot,
.rc-status-badge.awaiting .rc-status-dot { animation: pulse 1.5s infinite; }

/* Card Tiers */
.rc-tier {
  display: flex; flex-wrap: wrap; gap: 8px 16px;
  padding: 8px 14px;
  border-top: 1px solid #f8fafc;
}
.rc-field { display: flex; flex-direction: column; gap: 2px; }
.rc-field.full { flex: 1; min-width: 0; }
.rc-label { font-size: 10px; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; font-weight: 500; }
.rc-value { font-size: 12px; color: #475569; }
.rc-value.ns {
  background: #eff6ff; color: #2563eb; padding: 1px 6px;
  border-radius: 4px; font-size: 11px; font-family: 'SF Mono', monospace;
}
.rc-image {
  font-size: 11px; background: #f8fafc; padding: 3px 8px; border-radius: 4px;
  color: #475569; font-family: 'SF Mono', monospace;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  display: block;
}
.rc-workload {
  font-size: 11px; background: #f1f5f9; padding: 2px 6px;
  border-radius: 4px; color: #475569; font-family: 'SF Mono', monospace;
}

/* Error Banner */
.rc-error {
  display: flex; align-items: flex-start; gap: 6px;
  margin: 0 14px; padding: 8px 10px;
  background: #fef2f2; border-radius: 6px;
  border-left: 2px solid #ef4444;
}
.rc-error svg { width: 14px; height: 14px; color: #ef4444; flex-shrink: 0; margin-top: 1px; }
.rc-error span { font-size: 11px; color: #dc2626; line-height: 1.4; word-break: break-all; }

/* Card Actions */
.rc-actions {
  display: flex; border-top: 1px solid #f8fafc;
  margin-top: 8px;
}
.rc-act-btn {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 4px;
  padding: 9px 4px; border: none; background: transparent;
  cursor: pointer; font-size: 11px; color: #64748b; transition: all 0.15s;
}
.rc-act-btn:hover { background: #f8fafc; color: #334155; }
.rc-act-btn.primary { color: #4e7cf6; font-weight: 600; }
.rc-act-btn.primary:hover { background: #f0f5ff; }
.rc-act-btn.warning { color: #f59e0b; }
.rc-act-btn.warning:hover { background: #fffbeb; }
.rc-act-btn.danger { color: #ef4444; }
.rc-act-btn.danger:hover { background: #fef2f2; }
.rc-act-btn.danger-text { color: #94a3b8; }
.rc-act-btn.danger-text:hover { background: #fef2f2; color: #ef4444; }
.rc-act-btn svg { width: 13px; height: 13px; }
.rc-act-btn + .rc-act-btn { border-left: 1px solid #f8fafc; }
</style>

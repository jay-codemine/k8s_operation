<template>
  <div class="artifacts-page">
    <!-- ====== 页面头部 - 大厂风格渐变标题区 ====== -->
    <div class="page-hero">
      <div class="hero-content">
        <div class="hero-left">
          <div class="hero-breadcrumb">
            <span class="breadcrumb-item">CI/CD</span>
            <span class="breadcrumb-sep">/</span>
            <span class="breadcrumb-current">制品管理</span>
          </div>
          <h1 class="hero-title">
            <div class="hero-icon-wrapper">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
            </div>
            制品库
          </h1>
          <p class="hero-desc">统一管理构建产物，支持多语言制品存储、版本溯源与安全校验</p>
        </div>
        <div class="hero-right">
          <button class="hero-btn primary" @click="openCreateModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            新增制品
          </button>
          <button class="hero-btn secondary" @click="loadArtifacts" :disabled="loading">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            {{ loading ? '刷新中...' : '刷新数据' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ====== 上传进度条（全局浮层） ====== -->
    <Transition name="fade">
      <div v-if="uploading || creating" class="upload-progress-overlay">
        <div class="upload-progress-card">
          <div class="upload-progress-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </div>
          <div class="upload-progress-body">
            <div class="upload-progress-title">正在上传制品文件...</div>
            <div class="upload-progress-bar">
              <div class="upload-progress-fill" :style="{ width: uploadProgress + '%' }"></div>
            </div>
            <div class="upload-progress-text">{{ uploadProgress }}%</div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ====== 数据概览卡片 ====== -->
    <div class="overview-section">
      <div class="overview-cards">
        <div class="overview-card">
          <div class="oc-icon total">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            </svg>
          </div>
          <div class="oc-body">
            <div class="oc-value">{{ total }}</div>
            <div class="oc-label">制品总数</div>
          </div>
          <div class="oc-trend up" v-if="total > 0">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/></svg>
          </div>
        </div>

        <div class="overview-card" v-for="s in artifactTypeStats" :key="s.type">
          <div class="oc-icon" :class="s.color">
            <component :is="getTypeIcon(s.type)" />
          </div>
          <div class="oc-body">
            <div class="oc-value">{{ s.count }}</div>
            <div class="oc-label">{{ s.label }}</div>
          </div>
          <div class="oc-sparkline">
            <div class="sparkline-bar" :style="{ height: getSparkHeight(s.count) + '%' }"></div>
            <div class="sparkline-bar" :style="{ height: getSparkHeight(s.count * 0.7) + '%' }"></div>
            <div class="sparkline-bar" :style="{ height: getSparkHeight(s.count * 1.2) + '%' }"></div>
            <div class="sparkline-bar active" :style="{ height: getSparkHeight(s.count) + '%' }"></div>
          </div>
        </div>

        <div class="overview-card" v-if="artifactTypeStats.length === 0 && !loading" style="grid-column: span 2;">
          <div class="oc-icon muted">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          </div>
          <div class="oc-body">
            <div class="oc-value" style="font-size:18px;color:#9ca3af;">暂无统计</div>
            <div class="oc-label">运行流水线后将自动生成制品</div>
          </div>
        </div>
      </div>
    </div>

    <!-- ====== 主体内容区 ====== -->
    <div class="main-section">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <div class="search-box">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="keyword" type="text" placeholder="搜索制品名称、版本号或 Git 分支..." @input="debounceSearch" />
            <button v-if="keyword" class="search-clear" @click="keyword = ''; loadArtifacts()">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
        </div>
        <div class="toolbar-right">
            <button v-if="selectedIds.length > 0" class="batch-del-btn" @click="handleBatchDelete">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              批量删除 ({{ selectedIds.length }})
            </button>
          <div class="filter-chips">
            <select v-model="filters.artifact_type" class="chip-select" @change="loadArtifacts">
              <option value="">全部类型</option>
              <option value="jar">JAR 包</option>
              <option value="binary">二进制</option>
              <option value="dist">前端产物</option>
              <option value="wheel">Python Wheel</option>
              <option value="image">Docker 镜像</option>
            </select>
            <select v-model="filters.language_type" class="chip-select" @change="loadArtifacts">
              <option value="">全部语言</option>
              <option value="java">Java</option>
              <option value="go">Go</option>
              <option value="frontend">Frontend</option>
              <option value="python">Python</option>
            </select>
            <select v-model="filters.status" class="chip-select" @change="loadArtifacts">
              <option value="">全部状态</option>
              <option value="ready">就绪</option>
              <option value="expired">已过期</option>
            </select>
          </div>
          <div class="view-toggle">
            <button :class="['toggle-btn', { active: viewMode === 'card' }]" @click="viewMode = 'card'" title="卡片视图">
              <svg viewBox="0 0 24 24" fill="currentColor"><rect x="3" y="3" width="8" height="8" rx="1"/><rect x="13" y="3" width="8" height="8" rx="1"/><rect x="3" y="13" width="8" height="8" rx="1"/><rect x="13" y="13" width="8" height="8" rx="1"/></svg>
            </button>
            <button :class="['toggle-btn', { active: viewMode === 'table' }]" @click="viewMode = 'table'" title="列表视图">
              <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 4h18v2H3V4zm0 7h18v2H3v-2zm0 7h18v2H3v-2z"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- 加载中 -->
      <div v-if="loading && artifacts.length === 0" class="state-container">
        <div class="loader">
          <div class="loader-ring"></div>
          <div class="loader-ring"></div>
          <div class="loader-ring"></div>
        </div>
        <p class="state-text">正在加载制品数据...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="artifacts.length === 0" class="state-container empty">
        <div class="empty-illustration">
          <svg viewBox="0 0 200 160" fill="none">
            <rect x="40" y="30" width="120" height="100" rx="8" fill="#f3f4f6" stroke="#e5e7eb" stroke-width="2"/>
            <rect x="55" y="50" width="90" height="8" rx="4" fill="#e5e7eb"/>
            <rect x="55" y="66" width="60" height="8" rx="4" fill="#e5e7eb"/>
            <rect x="55" y="82" width="75" height="8" rx="4" fill="#e5e7eb"/>
            <circle cx="100" cy="115" r="12" fill="#dbeafe" stroke="#93c5fd" stroke-width="2"/>
            <path d="M95 115l3 3 7-7" stroke="#3b82f6" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h3 class="empty-title">暂无制品记录</h3>
        <p class="empty-desc">运行 CI/CD 流水线后，构建产物将自动上传到制品库进行统一管理</p>
        <router-link to="/cicd/pipelines" class="empty-action">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
          前往流水线
        </router-link>
      </div>

      <!-- ====== 卡片视图 ====== -->
      <div v-else-if="viewMode === 'card'" class="card-grid">
        <div class="artifact-card" v-for="a in artifacts" :key="a.id" @click="showDetail(a)">
          <div class="card-header">
            <div class="card-type-badge" :class="getArtifactTypeClass(a.artifact_type)">
              {{ getArtifactTypeLabel(a.artifact_type) }}
            </div>
            <div class="card-status" :class="a.status">
              <span class="status-dot"></span>
              {{ getStatusLabel(a.status) }}
            </div>
          </div>
          <div class="card-body">
            <h4 class="card-name" :title="a.name">{{ a.name }}</h4>
            <div class="card-meta">
              <span class="meta-item version" v-if="a.version">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
                {{ a.version }}
              </span>
              <span class="meta-item branch" v-if="a.git_branch">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>
                {{ a.git_branch }}
              </span>
            </div>
          </div>
          <div class="card-footer">
            <div class="card-info">
              <span class="info-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/></svg>
                {{ formatFileSize(a.file_size) }}
              </span>
              <span class="info-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                {{ a.download_count || 0 }}
              </span>
              <span class="info-item time">{{ formatTimeAgo(a.created_at) }}</span>
            </div>
            <div class="card-actions" @click.stop>
              <button class="card-action-btn primary" @click="handleSmartDownload(a)" :title="getDownloadTitle(a)">
                <svg v-if="a.file_path" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="8" y="2" width="8" height="4" rx="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
              </button>
              <button class="card-action-btn edit" @click="openEditModal(a)" title="编辑">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button class="card-action-btn danger" @click="handleDelete(a)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ====== 表格视图 ====== -->
      <div v-else class="table-wrapper">
        <table class="pro-table">
          <thead>
            <tr>
              <th class="col-check"><input type="checkbox" :checked="selectedIds.length === artifacts.length && artifacts.length > 0" @change="toggleSelectAll" /></th>
              <th class="col-name">制品名称</th>
              <th class="col-type">类型</th>
              <th class="col-lang">语言</th>
              <th class="col-version">版本</th>
              <th class="col-size">大小</th>
              <th class="col-branch">Git 分支</th>
              <th class="col-status">状态</th>
              <th class="col-download">下载</th>
              <th class="col-time">创建时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in artifacts" :key="a.id" class="table-row" @click="showDetail(a)">
              <td class="col-check" @click.stop><input type="checkbox" :checked="selectedIds.includes(a.id)" @change="toggleSelect(a.id)" /></td>
              <td class="col-name">
                <div class="name-cell">
                  <div class="name-avatar" :class="getArtifactTypeClass(a.artifact_type)">
                    {{ getArtifactTypeLabel(a.artifact_type).charAt(0) }}
                  </div>
                  <div class="name-info">
                    <span class="name-text" :title="a.name">{{ a.name }}</span>
                    <span class="name-sha" v-if="a.git_commit">{{ a.git_commit.slice(0, 8) }}</span>
                  </div>
                </div>
              </td>
              <td><span class="badge type" :class="getArtifactTypeClass(a.artifact_type)">{{ getArtifactTypeLabel(a.artifact_type) }}</span></td>
              <td><span class="badge lang" :class="getLanguageClass(a.language_type)">{{ getLanguageLabel(a.language_type) }}</span></td>
              <td><code class="mono-text">{{ a.version || '-' }}</code></td>
              <td class="size-cell">{{ formatFileSize(a.file_size) }}</td>
              <td>
                <span class="branch-chip" v-if="a.git_branch">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>
                  {{ a.git_branch }}
                </span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span class="status-badge" :class="a.status">
                  <span class="status-indicator"></span>
                  {{ getStatusLabel(a.status) }}
                </span>
              </td>
              <td class="download-cell">
                <span class="download-count">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  {{ a.download_count || 0 }}
                </span>
              </td>
              <td class="time-cell">{{ formatTime(a.created_at) }}</td>
              <td @click.stop>
                <div class="action-group">
                  <button class="act-btn download" @click="handleSmartDownload(a)" :title="getDownloadTitle(a)">
                    <svg v-if="a.file_path" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                    <svg v-else-if="a.image_repo" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="8" y="2" width="8" height="4" rx="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
                    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </button>
                  <button class="act-btn info" @click="showDetail(a)" title="查看详情">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                  </button>
                  <button class="act-btn edit" @click="openEditModal(a)" title="编辑">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button v-if="!a.file_path" class="act-btn upload" @click="handleAttachFile(a)" title="上传文件" :disabled="uploading">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  </button>
                  <button class="act-btn delete" @click="handleDelete(a)" title="删除">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="artifacts.length > 0" class="pagination-bar">
        <div class="page-summary">
          共 <strong>{{ total }}</strong> 个制品，第 {{ pagination.page }} / {{ totalPages }} 页
        </div>
        <div class="page-controls">
          <button class="page-btn" :disabled="pagination.page <= 1" @click="changePage(1)" title="首页">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="11 17 6 12 11 7"/><polyline points="18 17 13 12 18 7"/></svg>
          </button>
          <button class="page-btn" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <template v-for="p in visiblePages" :key="p">
            <button v-if="p === '...'" class="page-btn dots" disabled>...</button>
            <button v-else :class="['page-btn num', { active: p === pagination.page }]" @click="changePage(p)">{{ p }}</button>
          </template>
          <button class="page-btn" :disabled="pagination.page >= totalPages" @click="changePage(pagination.page + 1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
          </button>
          <button class="page-btn" :disabled="pagination.page >= totalPages" @click="changePage(totalPages)" title="末页">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="13 17 18 12 13 7"/><polyline points="6 17 11 12 6 7"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- ====== 编辑制品弹窗 ====== -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="showEditModal" class="drawer-mask" @click.self="showEditModal = false">
          <div class="edit-modal">
            <div class="drawer-header">
              <h3 class="drawer-title">编辑制品</h3>
              <button class="drawer-close" @click="showEditModal = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="edit-body">
              <div class="edit-field">
                <label>制品名称</label>
                <input v-model="editForm.name" type="text" placeholder="制品名称" />
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>版本号</label>
                  <input v-model="editForm.version" type="text" placeholder="如 v1.0.0" />
                </div>
                <div class="edit-field">
                  <label>制品类型</label>
                  <select v-model="editForm.artifact_type">
                    <option value="">不修改</option>
                    <option value="jar">JAR 包</option>
                    <option value="binary">二进制</option>
                    <option value="dist">前端产物</option>
                    <option value="wheel">Python Wheel</option>
                    <option value="image">Docker 镜像</option>
                    <option value="archive">通用压缩包</option>
                  </select>
                </div>
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>状态</label>
                  <select v-model="editForm.status">
                    <option value="">不修改</option>
                    <option value="ready">就绪</option>
                    <option value="expired">已过期</option>
                  </select>
                </div>
              </div>
              <div class="edit-field">
                <label>镜像仓库</label>
                <input v-model="editForm.image_repo" type="text" placeholder="如 harbor.example.com/proj/app" />
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>镜像标签</label>
                  <input v-model="editForm.image_tag" type="text" placeholder="如 v1.0.0" />
                </div>
                <div class="edit-field">
                  <label>镜像摘要</label>
                  <input v-model="editForm.image_digest" type="text" placeholder="sha256:..." />
                </div>
              </div>
            </div>
            <div class="drawer-footer">
              <button class="drawer-btn primary" @click="handleUpdate">保存修改</button>
              <button class="drawer-btn ghost" @click="showEditModal = false">取消</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ====== 新增制品弹窗 ====== -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="showCreateModal" class="drawer-mask" @click.self="showCreateModal = false">
          <div class="edit-modal">
            <div class="drawer-header">
              <h3 class="drawer-title">新增制品</h3>
              <button class="drawer-close" @click="showCreateModal = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="edit-body">
              <div class="edit-field">
                <label>制品名称 <span style="color:#e53e3e">*</span></label>
                <input v-model="createForm.name" type="text" placeholder="如 order-service-v1.0.0.jar" />
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>版本号</label>
                  <input v-model="createForm.version" type="text" placeholder="如 v1.0.0" />
                </div>
                <div class="edit-field">
                  <label>制品类型</label>
                  <select v-model="createForm.artifact_type">
                    <option value="jar">JAR 包</option>
                    <option value="binary">二进制</option>
                    <option value="dist">前端产物</option>
                    <option value="wheel">Python Wheel</option>
                    <option value="image">Docker 镜像</option>
                    <option value="archive">通用压缩包</option>
                  </select>
                </div>
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>语言类型</label>
                  <select v-model="createForm.language_type">
                    <option value="java">Java</option>
                    <option value="go">Go</option>
                    <option value="frontend">Frontend</option>
                    <option value="python">Python</option>
                  </select>
                </div>
                <div class="edit-field">
                  <label>构建号</label>
                  <input v-model.number="createForm.build_number" type="number" placeholder="如 42" />
                </div>
              </div>
              <div class="edit-field">
                <label>Git 仓库地址</label>
                <input v-model="createForm.git_repo" type="text" placeholder="https://github.com/org/repo.git" />
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>Git 分支</label>
                  <input v-model="createForm.git_branch" type="text" placeholder="main" />
                </div>
                <div class="edit-field">
                  <label>Git Commit</label>
                  <input v-model="createForm.git_commit" type="text" placeholder="commit SHA" />
                </div>
              </div>
              <div class="create-section-title">镜像信息（可选）</div>
              <div class="edit-field">
                <label>镜像仓库</label>
                <input v-model="createForm.image_repo" type="text" placeholder="harbor.example.com/proj/app" />
              </div>
              <div class="edit-row">
                <div class="edit-field">
                  <label>镜像标签</label>
                  <input v-model="createForm.image_tag" type="text" placeholder="v1.0.0" />
                </div>
                <div class="edit-field">
                  <label>镜像摘要</label>
                  <input v-model="createForm.image_digest" type="text" placeholder="sha256:..." />
                </div>
              </div>
              <div class="create-section-title">上传制品文件（可选）</div>
              <div class="upload-zone" @click="$refs.createFileInput && $refs.createFileInput.click()" @dragover.prevent @drop.prevent="(e) => { const f = e.dataTransfer.files[0]; if(f) { createFile = f; if(!createForm.name) createForm.name = f.name } }">
                <input ref="createFileInput" type="file" style="display:none" accept=".jar,.war,.zip,.tar,.gz,.tgz,.whl,.bin,.exe,.dll,.so" @change="onCreateFileChange" />
                <div v-if="!createFile" class="upload-placeholder">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  <span>点击或拖拽文件到此处上传</span>
                  <span class="upload-hint">支持 JAR、WAR、ZIP、TAR 等格式</span>
                </div>
                <div v-else class="upload-selected">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                  <div class="upload-file-info">
                    <span class="upload-file-name">{{ createFile.name }}</span>
                    <span class="upload-file-size">{{ formatFileSize(createFile.size) }}</span>
                  </div>
                  <button class="upload-remove" @click.stop="createFile = null" title="移除文件">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
              </div>
            </div>
            <div class="drawer-footer">
              <button class="drawer-btn primary" @click="handleCreate" :disabled="creating">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                {{ creating ? '创建中...' : '确认创建' }}
              </button>
              <button class="drawer-btn ghost" @click="showCreateModal = false">取消</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ====== 制品详情抽屉 ====== -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="showDetailDrawer" class="drawer-mask" @click.self="showDetailDrawer = false">
          <div class="drawer-panel">
            <div class="drawer-header">
              <div class="drawer-title-group">
                <h3 class="drawer-title">制品详情</h3>
                <span class="drawer-subtitle" v-if="currentArtifact">{{ currentArtifact.name }}</span>
              </div>
              <button class="drawer-close" @click="showDetailDrawer = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>

            <div class="drawer-body" v-if="currentArtifact">
              <!-- 顶部状态条 -->
              <div class="detail-status-bar" :class="currentArtifact.status">
                <span class="dsb-dot"></span>
                <span class="dsb-label">{{ getStatusLabel(currentArtifact.status) }}</span>
                <span class="dsb-type">{{ getArtifactTypeLabel(currentArtifact.artifact_type) }} · {{ getLanguageLabel(currentArtifact.language_type) }}</span>
              </div>

              <!-- Tab切换 -->
              <div class="detail-tabs">
                <button :class="['tab-btn', { active: detailTab === 'info' }]" @click="detailTab = 'info'">基本信息</button>
                <button :class="['tab-btn', { active: detailTab === 'git' }]" @click="detailTab = 'git'">Git 溯源</button>
                <button :class="['tab-btn', { active: detailTab === 'image' }]" @click="detailTab = 'image'">镜像信息</button>
                <button :class="['tab-btn', { active: detailTab === 'build' }]" @click="detailTab = 'build'">构建数据</button>
              </div>

              <!-- 基本信息 -->
              <div v-if="detailTab === 'info'" class="detail-section">
                <div class="detail-grid">
                  <div class="dg-item"><div class="dg-label">制品名称</div><div class="dg-value">{{ currentArtifact.name }}</div></div>
                  <div class="dg-item"><div class="dg-label">版本号</div><div class="dg-value mono">{{ currentArtifact.version || '-' }}</div></div>
                  <div class="dg-item"><div class="dg-label">文件大小</div><div class="dg-value">{{ formatFileSize(currentArtifact.file_size) }}</div></div>
                  <div class="dg-item"><div class="dg-label">下载次数</div><div class="dg-value">{{ currentArtifact.download_count || 0 }} 次</div></div>
                  <div class="dg-item"><div class="dg-label">存储类型</div><div class="dg-value">{{ currentArtifact.storage_type || 'local' }}</div></div>
                  <div class="dg-item"><div class="dg-label">创建时间</div><div class="dg-value">{{ formatTime(currentArtifact.created_at) }}</div></div>
                  <div class="dg-item full"><div class="dg-label">SHA256 校验和</div><div class="dg-value mono break-all">{{ currentArtifact.sha256 || '-' }}</div></div>
                  <div class="dg-item full"><div class="dg-label">存储路径</div><div class="dg-value mono break-all">{{ currentArtifact.file_path || '-' }}</div></div>
                </div>
              </div>

              <!-- Git 溯源 -->
              <div v-if="detailTab === 'git'" class="detail-section">
                <div class="detail-grid">
                  <div class="dg-item full"><div class="dg-label">Git 仓库</div><div class="dg-value mono break-all">{{ currentArtifact.git_repo || '-' }}</div></div>
                  <div class="dg-item"><div class="dg-label">Git 分支</div><div class="dg-value"><span class="branch-chip sm" v-if="currentArtifact.git_branch"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="6" y1="3" x2="6" y2="15"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>{{ currentArtifact.git_branch }}</span><span v-else>-</span></div></div>
                  <div class="dg-item"><div class="dg-label">Git Commit</div><div class="dg-value mono">{{ currentArtifact.git_commit || '-' }}</div></div>
                  <div class="dg-item"><div class="dg-label">构建号</div><div class="dg-value">#{{ currentArtifact.build_number || '-' }}</div></div>
                  <div class="dg-item"><div class="dg-label">关联流水线</div><div class="dg-value">{{ currentArtifact.pipeline_id ? 'Pipeline #' + currentArtifact.pipeline_id : '-' }}</div></div>
                </div>
              </div>

              <!-- 镜像信息 -->
              <div v-if="detailTab === 'image'" class="detail-section">
                <div class="detail-grid">
                  <div class="dg-item full"><div class="dg-label">镜像仓库</div><div class="dg-value mono break-all">{{ currentArtifact.image_repo || '未关联镜像' }}</div></div>
                  <div class="dg-item"><div class="dg-label">镜像标签</div><div class="dg-value mono">{{ currentArtifact.image_tag || '-' }}</div></div>
                  <div class="dg-item full"><div class="dg-label">镜像摘要</div><div class="dg-value mono break-all">{{ currentArtifact.image_digest || '-' }}</div></div>
                </div>
              </div>

              <!-- 构建数据 -->
              <div v-if="detailTab === 'build'" class="detail-section">
                <div class="detail-grid">
                  <div class="dg-item"><div class="dg-label">构建耗时</div><div class="dg-value">{{ currentArtifact.build_duration ? currentArtifact.build_duration + ' 秒' : '-' }}</div></div>
                  <div class="dg-item"><div class="dg-label">运行记录</div><div class="dg-value">{{ currentArtifact.run_id ? 'Run #' + currentArtifact.run_id : '-' }}</div></div>
                </div>
                <div v-if="currentArtifact.build_log" class="build-log-box">
                  <div class="log-header">构建摘要日志</div>
                  <pre class="log-content">{{ currentArtifact.build_log }}</pre>
                </div>
              </div>
            </div>

            <div class="drawer-footer" v-if="currentArtifact">
              <button v-if="currentArtifact.file_path" class="drawer-btn primary" @click="handleDownload(currentArtifact)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                下载制品
              </button>
              <button v-if="!currentArtifact.file_path" class="drawer-btn primary" @click="handleAttachFile(currentArtifact)" :disabled="uploading">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                {{ uploading ? '上传中...' : '上传文件' }}
              </button>
              <button class="drawer-btn danger-outline" @click="handleDelete(currentArtifact)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
              <button class="drawer-btn secondary" @click="openEditModal(currentArtifact); showDetailDrawer = false">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    编辑
                  </button>
                  <button class="drawer-btn ghost" @click="showDetailDrawer = false">关闭</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, h } from 'vue'
import { getArtifacts, getArtifactStats, getArtifactDetail, deleteArtifact, downloadArtifact, updateArtifact, batchDeleteArtifacts, createArtifactRecord, getArtifactDownloadUrl, attachArtifactFile, uploadArtifact } from '@/api/cicd'

const getToken = () => localStorage.getItem('token') || sessionStorage.getItem('token')
// ====== 轻量 Toast 提示 ======
const showToast = (msg, type = 'info') => {
  const colors = { success: '#38a169', error: '#e53e3e', info: '#3182ce', warning: '#dd6b20' }
  const el = document.createElement('div')
  el.textContent = msg
  Object.assign(el.style, {
    position: 'fixed', top: '20px', left: '50%', transform: 'translateX(-50%)',
    padding: '10px 24px', borderRadius: '8px', color: '#fff', fontSize: '14px',
    background: colors[type] || colors.info, zIndex: '99999', boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    transition: 'opacity 0.3s', opacity: '1'
  })
  document.body.appendChild(el)
  setTimeout(() => { el.style.opacity = '0'; setTimeout(() => el.remove(), 300) }, 2500)
}

// ====== 状态 ======
const loading = ref(false)
const artifacts = ref([])
const total = ref(0)
const keyword = ref('')
const showDetailDrawer = ref(false)
const currentArtifact = ref(null)
const artifactTypeStats = ref([])
const viewMode = ref('table')
const detailTab = ref('info')
const selectedIds = ref([])
const editingArtifact = ref(null)
const showEditModal = ref(false)
const editForm = reactive({ name: '', version: '', artifact_type: '', status: '', image_repo: '', image_tag: '', image_digest: '' })
const showCreateModal = ref(false)
const creating = ref(false)
const createFile = ref(null)
const createForm = reactive({
  name: '', version: '', artifact_type: 'jar', language_type: 'java',
  build_number: 0, git_repo: '', git_branch: '', git_commit: '',
  image_repo: '', image_tag: '', image_digest: ''
})
const uploading = ref(false)
const uploadProgress = ref(0) // 上传进度 0-100

const pagination = reactive({ page: 1, page_size: 20 })
const filters = reactive({ artifact_type: '', language_type: '', status: '' })

const totalPages = computed(() => Math.ceil(total.value / pagination.page_size) || 1)

const visiblePages = computed(() => {
  const pages = []
  const tp = totalPages.value
  const cp = pagination.page
  if (tp <= 7) { for (let i = 1; i <= tp; i++) pages.push(i) }
  else {
    pages.push(1)
    if (cp > 3) pages.push('...')
    for (let i = Math.max(2, cp - 1); i <= Math.min(tp - 1, cp + 1); i++) pages.push(i)
    if (cp < tp - 2) pages.push('...')
    pages.push(tp)
  }
  return pages
})

// ====== 方法 ======
let searchTimer = null
const debounceSearch = () => { clearTimeout(searchTimer); searchTimer = setTimeout(() => loadArtifacts(), 300) }

const changePage = (page) => { pagination.page = page; loadArtifacts() }

const loadArtifacts = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.page_size, ...filters }
    if (keyword.value) params.keyword = keyword.value
    const res = await getArtifacts(params)
    artifacts.value = res.data?.data?.list || res.data?.list || []
    total.value = res.data?.data?.total || res.data?.total || 0
  } catch (e) {
    showToast('加载制品列表失败: ' + (e.message || ''), 'error')
  } finally { loading.value = false }
}

const loadStats = async () => {
  try {
    const res = await getArtifactStats()
    const stats = res.data?.data?.stats || res.data?.stats || []
    const typeMap = {
      jar: { label: 'JAR 包', color: 'java' }, binary: { label: '二进制文件', color: 'go' },
      dist: { label: '前端产物', color: 'frontend' }, wheel: { label: 'Python Wheel', color: 'python' },
      image: { label: 'Docker 镜像', color: 'docker' },
    }
    artifactTypeStats.value = stats.map(s => ({
      type: s.artifact_type, count: s.count,
      label: typeMap[s.artifact_type]?.label || s.artifact_type,
      color: typeMap[s.artifact_type]?.color || 'total',
    }))
  } catch (e) { console.warn('加载统计失败', e) }
}

const showDetail = async (a) => {
  detailTab.value = 'info'
  try {
    const res = await getArtifactDetail(a.id)
    currentArtifact.value = res.data?.data?.artifact || res.data?.artifact || a
  } catch { currentArtifact.value = a }
  showDetailDrawer.value = true
}

const handleDownload = (a) => {
  if (!a.file_path) {
    showToast('该制品暂无文件可下载', 'warning')
    return
  }
  const token = getToken()
  if (!token) {
    showToast('认证已过期，请重新登录', 'error')
    return
  }
  const url = getArtifactDownloadUrl(a.id, token)
  window.open(url, '_blank')
  showToast('正在下载 ' + (a.name || '制品文件'), 'success')
  // 延迟刷新下载计数
  setTimeout(() => loadArtifacts(), 1500)
}

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    showToast('已复制到剪贴板', 'success')
  }).catch(() => {
    // fallback
    const ta = document.createElement('textarea')
    ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0'
    document.body.appendChild(ta); ta.select(); document.execCommand('copy')
    document.body.removeChild(ta)
    showToast('已复制到剪贴板', 'success')
  })
}

const handleSmartDownload = (a) => {
  if (a.file_path) {
    handleDownload(a)
  } else if (a.image_repo) {
    const fullImage = a.image_tag ? `${a.image_repo}:${a.image_tag}` : a.image_repo
    copyToClipboard(`docker pull ${fullImage}`)
  } else {
    showToast('该制品暂无可下载文件，可通过「上传文件」补充', 'info')
  }
}

// ====== 补传文件 ======
const handleAttachFile = (a) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.jar,.war,.zip,.tar,.gz,.tgz,.whl,.bin,.exe,.dll,.so'
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    uploading.value = true
    uploadProgress.value = 0
    try {
      await attachArtifactFile(a.id, file, (pct) => { uploadProgress.value = pct })
      showToast(`文件 ${file.name} 上传成功`, 'success')
      loadArtifacts(); loadStats()
      if (showDetailDrawer.value && currentArtifact.value?.id === a.id) {
        showDetail(a)
      }
    } catch (e) {
      showToast('文件上传失败: ' + (e.message || ''), 'error')
    } finally { uploading.value = false; uploadProgress.value = 0 }
  }
  input.click()
}

const getDownloadTitle = (a) => {
  if (a.file_path) return '下载制品文件'
  if (a.image_repo) return '复制镜像拉取命令'
  return '暂无可下载内容'
}

const handleDelete = async (a) => {
  if (!window.confirm(`确定要删除制品「${a.name}」吗？此操作不可恢复。`)) return
  try {
    await deleteArtifact(a.id)
    showToast('删除成功', 'success')
    showDetailDrawer.value = false
    loadArtifacts(); loadStats()
  } catch (e) { showToast('删除失败: ' + (e.message || ''), 'error') }
}

const toggleSelect = (id) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}
const toggleSelectAll = () => {
  if (selectedIds.value.length === artifacts.value.length) selectedIds.value = []
  else selectedIds.value = artifacts.value.map(a => a.id)
}
const handleBatchDelete = async () => {
  if (selectedIds.value.length === 0) { showToast('请先选择要删除的制品', 'warning'); return }
  if (!window.confirm(`确定要批量删除 ${selectedIds.value.length} 个制品吗？此操作不可恢复。`)) return
  try {
    const res = await batchDeleteArtifacts(selectedIds.value)
    const affected = res.data?.data?.affected || res.data?.affected || selectedIds.value.length
    showToast(`批量删除成功，共删除 ${affected} 个制品`, 'success')
    selectedIds.value = []
    loadArtifacts(); loadStats()
  } catch (e) { showToast('批量删除失败: ' + (e.message || ''), 'error') }
}
const openCreateModal = () => {
  createForm.name = ''
  createForm.version = ''
  createForm.artifact_type = 'jar'
  createForm.language_type = 'java'
  createForm.build_number = 0
  createForm.git_repo = ''
  createForm.git_branch = ''
  createForm.git_commit = ''
  createForm.image_repo = ''
  createForm.image_tag = ''
  createForm.image_digest = ''
  createFile.value = null
  showCreateModal.value = true
}
const onCreateFileChange = (e) => {
  const file = e.target.files[0]
  if (!file) return
  createFile.value = file
  // 自动填充名称
  if (!createForm.name) createForm.name = file.name
}
const handleCreate = async () => {
  if (!createForm.name.trim() && !createFile.value) { showToast('请输入制品名称或上传文件', 'warning'); return }
  creating.value = true
  uploadProgress.value = 0
  try {
    if (createFile.value) {
      // 带文件上传 → 走 upload 接口（multipart）
      const meta = { ...createForm }
      if (!meta.name) meta.name = createFile.value.name
      await uploadArtifact(createFile.value, meta, (pct) => { uploadProgress.value = pct })
      showToast('制品上传成功', 'success')
    } else {
      // 仅创建记录（镜像类型等）
      await createArtifactRecord({ ...createForm })
      showToast('创建成功', 'success')
    }
    showCreateModal.value = false
    loadArtifacts(); loadStats()
  } catch (e) { showToast('创建失败: ' + (e.message || ''), 'error') }
  finally { creating.value = false; uploadProgress.value = 0 }
}
const openEditModal = (a) => {
  editingArtifact.value = a
  editForm.name = a.name || ''
  editForm.version = a.version || ''
  editForm.artifact_type = a.artifact_type || ''
  editForm.status = a.status || ''
  editForm.image_repo = a.image_repo || ''
  editForm.image_tag = a.image_tag || ''
  editForm.image_digest = a.image_digest || ''
  showEditModal.value = true
}
const handleUpdate = async () => {
  if (!editingArtifact.value) return
  const data = { id: editingArtifact.value.id }
  if (editForm.name && editForm.name !== editingArtifact.value.name) data.name = editForm.name
  if (editForm.version && editForm.version !== editingArtifact.value.version) data.version = editForm.version
  if (editForm.artifact_type && editForm.artifact_type !== editingArtifact.value.artifact_type) data.artifact_type = editForm.artifact_type
  if (editForm.status && editForm.status !== editingArtifact.value.status) data.status = editForm.status
  if (editForm.image_repo && editForm.image_repo !== editingArtifact.value.image_repo) data.image_repo = editForm.image_repo
  if (editForm.image_tag && editForm.image_tag !== editingArtifact.value.image_tag) data.image_tag = editForm.image_tag
  if (editForm.image_digest && editForm.image_digest !== editingArtifact.value.image_digest) data.image_digest = editForm.image_digest
  if (Object.keys(data).length <= 1) { showToast('没有修改任何字段', 'warning'); return }
  try {
    await updateArtifact(data)
    showToast('更新成功', 'success')
    showEditModal.value = false
    loadArtifacts()
  } catch (e) { showToast('更新失败: ' + (e.message || ''), 'error') }
}

// ====== 工具函数 ======
const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB'
  return (bytes / 1073741824).toFixed(2) + ' GB'
}

const formatTime = (ts) => { if (!ts) return '-'; return new Date(ts * 1000).toLocaleString('zh-CN', { hour12: false }) }

const formatTimeAgo = (ts) => {
  if (!ts) return '-'
  const diff = Math.floor(Date.now() / 1000 - ts)
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  if (diff < 2592000) return Math.floor(diff / 86400) + ' 天前'
  return formatTime(ts)
}

const getArtifactTypeClass = (t) => ({ jar: 'java', binary: 'go', dist: 'frontend', wheel: 'python', image: 'docker' }[t] || 'default')
const getArtifactTypeLabel = (t) => ({ jar: 'JAR', binary: 'Binary', dist: 'Dist', wheel: 'Wheel', image: 'Image', archive: 'Archive' }[t] || t || '-')
const getLanguageClass = (l) => ({ java: 'java', go: 'go', frontend: 'frontend', python: 'python' }[l] || '')
const getLanguageLabel = (l) => ({ java: 'Java', go: 'Go', frontend: 'Frontend', python: 'Python' }[l] || l || '-')
const getStatusLabel = (s) => ({ ready: '就绪', uploading: '上传中', expired: '已过期', deleted: '已删除' }[s] || s || '-')
const getSparkHeight = (v) => Math.min(100, Math.max(15, (v / (total.value || 1)) * 100 + 20))

const getTypeIcon = (type) => {
  return { render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2' }, [
    h('path', { d: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z' }),
    h('polyline', { points: '14 2 14 8 20 8' })
  ])}
}

onMounted(() => { loadArtifacts(); loadStats() })
</script>

<style scoped>
/* ===== 页面布局 ===== */
.artifacts-page { min-height: 100vh; background: #f0f2f5; }

/* ===== Hero 区域 ===== */
.page-hero { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 28px 32px 24px; }
.hero-content { display: flex; justify-content: space-between; align-items: flex-start; max-width: 1400px; margin: 0 auto; }
.hero-breadcrumb { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; font-size: 13px; }
.breadcrumb-item { color: rgba(255,255,255,0.65); }
.breadcrumb-sep { color: rgba(255,255,255,0.35); }
.breadcrumb-current { color: #fff; font-weight: 500; }
.hero-title { font-size: 26px; font-weight: 700; color: #fff; margin: 0 0 8px; display: flex; align-items: center; gap: 12px; }
.hero-icon-wrapper { width: 40px; height: 40px; background: rgba(255,255,255,0.18); border-radius: 12px; display: flex; align-items: center; justify-content: center; backdrop-filter: blur(10px); }
.hero-icon-wrapper svg { width: 22px; height: 22px; color: #fff; }
.hero-desc { margin: 0; color: rgba(255,255,255,0.8); font-size: 14px; line-height: 1.6; }
.hero-right { display: flex; gap: 10px; flex-shrink: 0; padding-top: 20px; }
.hero-btn { display: inline-flex; align-items: center; gap: 7px; padding: 9px 18px; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s; }
.hero-btn svg { width: 15px; height: 15px; }
.hero-btn.primary { background: rgba(255,255,255,0.95); color: #667eea; border: 1px solid rgba(255,255,255,0.9); }
.hero-btn.primary:hover { background: #fff; box-shadow: 0 4px 14px rgba(0,0,0,0.15); }
.hero-btn.secondary { background: rgba(255,255,255,0.18); color: #fff; backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.25); }
.hero-btn.secondary:hover { background: rgba(255,255,255,0.28); }
.hero-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 概览卡片 ===== */
.overview-section { max-width: 1400px; margin: -20px auto 0; padding: 0 32px; position: relative; z-index: 1; }
.overview-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; }
.overview-card { background: #fff; border-radius: 14px; padding: 22px; display: flex; align-items: center; gap: 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); transition: all 0.3s cubic-bezier(.4,0,.2,1); border: 1px solid transparent; }
.overview-card:hover { transform: translateY(-2px); box-shadow: 0 8px 25px rgba(0,0,0,0.1); border-color: rgba(102,126,234,0.15); }
.oc-icon { width: 52px; height: 52px; border-radius: 14px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.oc-icon svg { width: 26px; height: 26px; color: #fff; }
.oc-icon.total { background: linear-gradient(135deg, #667eea, #764ba2); }
.oc-icon.java { background: linear-gradient(135deg, #f56565, #e53e3e); }
.oc-icon.go { background: linear-gradient(135deg, #38b2ac, #319795); }
.oc-icon.frontend { background: linear-gradient(135deg, #48bb78, #38a169); }
.oc-icon.python { background: linear-gradient(135deg, #ed8936, #dd6b20); }
.oc-icon.docker { background: linear-gradient(135deg, #4299e1, #3182ce); }
.oc-icon.muted { background: #e2e8f0; }
.oc-icon.muted svg { color: #a0aec0; }
.oc-body { flex: 1; min-width: 0; }
.oc-value { font-size: 28px; font-weight: 800; color: #1a202c; line-height: 1.2; }
.oc-label { font-size: 13px; color: #718096; margin-top: 2px; }
.oc-trend { width: 32px; height: 32px; border-radius: 8px; background: #f0fff4; display: flex; align-items: center; justify-content: center; }
.oc-trend.up svg { width: 16px; height: 16px; color: #38a169; }
.oc-sparkline { display: flex; align-items: flex-end; gap: 3px; height: 36px; }
.sparkline-bar { width: 6px; border-radius: 3px; background: #e2e8f0; transition: height 0.3s; min-height: 6px; }
.sparkline-bar.active { background: linear-gradient(180deg, #667eea, #764ba2); }

/* ===== 主体内容 ===== */
.main-section { max-width: 1400px; margin: 20px auto 0; padding: 0 32px 32px; }

/* ===== 工具栏 ===== */
.toolbar { display: flex; justify-content: space-between; align-items: center; gap: 16px; margin-bottom: 16px; background: #fff; border-radius: 12px; padding: 14px 20px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.toolbar-left { flex: 1; }
.search-box { position: relative; max-width: 380px; }
.search-box input { width: 100%; padding: 9px 36px 9px 38px; border: 1.5px solid #e2e8f0; border-radius: 10px; font-size: 13.5px; color: #2d3748; background: #f7fafc; transition: all 0.2s; }
.search-box input:focus { border-color: #667eea; background: #fff; outline: none; box-shadow: 0 0 0 3px rgba(102,126,234,0.12); }
.search-box input::placeholder { color: #a0aec0; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: #a0aec0; pointer-events: none; }
.search-clear { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); width: 24px; height: 24px; border: none; background: #edf2f7; border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; }
.search-clear svg { width: 12px; height: 12px; color: #718096; }
.toolbar-right { display: flex; align-items: center; gap: 12px; }
.filter-chips { display: flex; gap: 8px; }
.chip-select { padding: 7px 12px; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 13px; color: #4a5568; background: #fff; cursor: pointer; transition: all 0.2s; }
.chip-select:focus { border-color: #667eea; outline: none; }
.view-toggle { display: flex; border: 1.5px solid #e2e8f0; border-radius: 8px; overflow: hidden; }
.toggle-btn { width: 36px; height: 34px; border: none; background: #fff; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.toggle-btn svg { width: 16px; height: 16px; color: #a0aec0; }
.toggle-btn.active { background: #667eea; }
.toggle-btn.active svg { color: #fff; }
.toggle-btn + .toggle-btn { border-left: 1px solid #e2e8f0; }

/* ===== 加载/空状态 ===== */
.state-container { text-align: center; padding: 80px 20px; background: #fff; border-radius: 16px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.loader { display: flex; justify-content: center; gap: 8px; margin-bottom: 20px; }
.loader-ring { width: 12px; height: 12px; border-radius: 50%; background: #667eea; animation: bounce 1.4s ease-in-out infinite both; }
.loader-ring:nth-child(1) { animation-delay: -0.32s; }
.loader-ring:nth-child(2) { animation-delay: -0.16s; }
@keyframes bounce { 0%,80%,100% { transform: scale(0.4); opacity: 0.4; } 40% { transform: scale(1); opacity: 1; } }
.state-text { font-size: 14px; color: #718096; }
.empty-illustration svg { width: 180px; height: 140px; margin-bottom: 20px; }
.empty-title { font-size: 18px; font-weight: 600; color: #2d3748; margin: 0 0 8px; }
.empty-desc { font-size: 14px; color: #718096; margin: 0 0 24px; max-width: 400px; display: inline-block; }
.empty-action { display: inline-flex; align-items: center; gap: 8px; padding: 10px 22px; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; border-radius: 10px; font-size: 14px; font-weight: 600; text-decoration: none; transition: all 0.2s; }
.empty-action:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(102,126,234,0.4); }
.empty-action svg { width: 16px; height: 16px; }

/* ===== 卡片视图 ===== */
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 16px; }
.artifact-card { background: #fff; border-radius: 14px; border: 1.5px solid #edf2f7; overflow: hidden; cursor: pointer; transition: all 0.25s cubic-bezier(.4,0,.2,1); }
.artifact-card:hover { border-color: #c3dafe; box-shadow: 0 8px 30px rgba(102,126,234,0.12); transform: translateY(-2px); }
.card-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px 0; }
.card-type-badge { font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
.card-type-badge.java { background: #fff5f5; color: #e53e3e; }
.card-type-badge.go { background: #e6fffa; color: #319795; }
.card-type-badge.frontend { background: #f0fff4; color: #38a169; }
.card-type-badge.python { background: #fffaf0; color: #dd6b20; }
.card-type-badge.docker { background: #ebf8ff; color: #3182ce; }
.card-type-badge.default { background: #f7fafc; color: #718096; }
.card-status { display: flex; align-items: center; gap: 5px; font-size: 12px; font-weight: 500; }
.card-status .status-dot { width: 7px; height: 7px; border-radius: 50%; }
.card-status.ready { color: #38a169; }
.card-status.ready .status-dot { background: #48bb78; box-shadow: 0 0 6px rgba(72,187,120,0.4); }
.card-status.expired { color: #e53e3e; }
.card-status.expired .status-dot { background: #fc8181; }
.card-status.uploading { color: #d69e2e; }
.card-status.uploading .status-dot { background: #ecc94b; animation: pulse-dot 2s ease-in-out infinite; }
@keyframes pulse-dot { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
.card-body { padding: 14px 20px; }
.card-name { font-size: 15px; font-weight: 600; color: #1a202c; margin: 0 0 10px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-meta { display: flex; gap: 14px; flex-wrap: wrap; }
.meta-item { display: flex; align-items: center; gap: 5px; font-size: 12.5px; color: #718096; }
.meta-item svg { width: 13px; height: 13px; flex-shrink: 0; }
.meta-item.version { color: #667eea; font-weight: 500; }
.meta-item.branch { color: #38a169; }
.card-footer { display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; background: #f7fafc; border-top: 1px solid #edf2f7; }
.card-info { display: flex; gap: 14px; }
.info-item { display: flex; align-items: center; gap: 4px; font-size: 12px; color: #a0aec0; }
.info-item svg { width: 12px; height: 12px; }
.info-item.time { font-size: 11.5px; }
.card-actions { display: flex; gap: 6px; }
.card-action-btn { width: 30px; height: 30px; border: none; border-radius: 7px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; background: transparent; }
.card-action-btn svg { width: 14px; height: 14px; }
.card-action-btn.primary { color: #667eea; }
.card-action-btn.primary:hover { background: #ebf4ff; }
.card-action-btn.edit { color: #ed8936; }
.card-action-btn.edit:hover { background: #fffaf0; }
.card-action-btn.danger { color: #e53e3e; }
.card-action-btn.danger:hover { background: #fff5f5; }

/* ===== 表格视图 ===== */
.table-wrapper { background: #fff; border-radius: 14px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,0.04); border: 1px solid #edf2f7; }
.pro-table { width: 100%; border-collapse: collapse; }
.pro-table th { padding: 13px 16px; text-align: left; font-size: 12px; font-weight: 700; color: #718096; text-transform: uppercase; letter-spacing: 0.5px; background: #f7fafc; border-bottom: 2px solid #edf2f7; white-space: nowrap; }
.pro-table td { padding: 14px 16px; border-bottom: 1px solid #f7fafc; font-size: 13.5px; color: #2d3748; }
.table-row { cursor: pointer; transition: background 0.15s; }
.table-row:hover { background: #f7fafc; }
.name-cell { display: flex; align-items: center; gap: 12px; }
.name-avatar { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 800; color: #fff; flex-shrink: 0; }
.name-avatar.java { background: linear-gradient(135deg, #f56565, #e53e3e); }
.name-avatar.go { background: linear-gradient(135deg, #38b2ac, #319795); }
.name-avatar.frontend { background: linear-gradient(135deg, #48bb78, #38a169); }
.name-avatar.python { background: linear-gradient(135deg, #ed8936, #dd6b20); }
.name-avatar.docker { background: linear-gradient(135deg, #4299e1, #3182ce); }
.name-avatar.default { background: linear-gradient(135deg, #a0aec0, #718096); }
.name-info { min-width: 0; }
.name-text { display: block; font-weight: 600; color: #1a202c; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 260px; }
.name-sha { display: block; font-size: 11px; color: #a0aec0; font-family: 'SF Mono', 'Fira Code', monospace; margin-top: 2px; }
.badge { display: inline-block; padding: 3px 10px; border-radius: 6px; font-size: 11.5px; font-weight: 600; }
.badge.type.java { background: #fff5f5; color: #e53e3e; }
.badge.type.go { background: #e6fffa; color: #319795; }
.badge.type.frontend { background: #f0fff4; color: #38a169; }
.badge.type.python { background: #fffaf0; color: #dd6b20; }
.badge.type.docker { background: #ebf8ff; color: #3182ce; }
.badge.type.default { background: #f7fafc; color: #718096; }
.badge.lang.java { background: #fed7d7; color: #c53030; }
.badge.lang.go { background: #b2f5ea; color: #285e61; }
.badge.lang.frontend { background: #c6f6d5; color: #276749; }
.badge.lang.python { background: #fefcbf; color: #975a16; }
.mono-text { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12.5px; color: #667eea; background: #f7fafc; padding: 2px 6px; border-radius: 4px; }
.size-cell { color: #718096; font-variant-numeric: tabular-nums; }
.branch-chip { display: inline-flex; align-items: center; gap: 4px; padding: 3px 10px; background: #f0fff4; color: #38a169; border-radius: 6px; font-size: 12px; font-weight: 500; }
.branch-chip svg { width: 12px; height: 12px; }
.branch-chip.sm { font-size: 13px; padding: 4px 12px; }
.text-muted { color: #cbd5e0; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 4px 10px; border-radius: 20px; font-size: 12px; font-weight: 600; }
.status-badge .status-indicator { width: 6px; height: 6px; border-radius: 50%; }
.status-badge.ready { background: #f0fff4; color: #38a169; }
.status-badge.ready .status-indicator { background: #48bb78; }
.status-badge.expired { background: #fff5f5; color: #e53e3e; }
.status-badge.expired .status-indicator { background: #fc8181; }
.status-badge.uploading { background: #fffff0; color: #d69e2e; }
.status-badge.uploading .status-indicator { background: #ecc94b; }
.download-cell { text-align: center; }
.download-count { display: inline-flex; align-items: center; gap: 4px; color: #718096; font-size: 13px; }
.download-count svg { width: 13px; height: 13px; }
.time-cell { color: #a0aec0; font-size: 12.5px; white-space: nowrap; }
.action-group { display: flex; gap: 4px; }
.act-btn { width: 32px; height: 32px; border: none; border-radius: 8px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; background: transparent; }
.act-btn svg { width: 15px; height: 15px; }
.act-btn.download { color: #667eea; }
.act-btn.download:hover { background: #ebf4ff; }
.act-btn.info { color: #4299e1; }
.act-btn.info:hover { background: #ebf8ff; }
.act-btn.edit { color: #ed8936; }
.act-btn.edit:hover { background: #fffaf0; }
.act-btn.delete { color: #e53e3e; }
.act-btn.delete:hover { background: #fff5f5; }
.act-btn.upload { color: #38a169; }
.act-btn.upload:hover { background: #f0fff4; }
.act-btn.upload:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 分页 ===== */
.pagination-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 0; }
.page-summary { font-size: 13px; color: #718096; }
.page-summary strong { color: #2d3748; }
.page-controls { display: flex; align-items: center; gap: 4px; }
.page-btn { width: 34px; height: 34px; border: 1.5px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.page-btn svg { width: 14px; height: 14px; color: #718096; }
.page-btn:hover:not(:disabled) { border-color: #667eea; background: #f7fafc; }
.page-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.page-btn.num { font-size: 13px; color: #4a5568; font-weight: 500; }
.page-btn.num.active { background: #667eea; border-color: #667eea; color: #fff; }
.page-btn.dots { border: none; background: transparent; cursor: default; font-size: 14px; color: #a0aec0; }

/* ===== 抽屉详情 ===== */
.drawer-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.45); z-index: 2000; display: flex; justify-content: flex-end; backdrop-filter: blur(2px); }
.drawer-panel { width: 520px; max-width: 90vw; background: #fff; height: 100vh; display: flex; flex-direction: column; box-shadow: -8px 0 30px rgba(0,0,0,0.12); }
.drawer-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 24px 28px 20px; border-bottom: 1px solid #edf2f7; }
.drawer-title { font-size: 20px; font-weight: 700; color: #1a202c; margin: 0; }
.drawer-subtitle { font-size: 13px; color: #718096; margin-top: 4px; display: block; max-width: 380px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.drawer-close { width: 36px; height: 36px; border: none; background: #f7fafc; border-radius: 10px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; }
.drawer-close svg { width: 18px; height: 18px; color: #718096; }
.drawer-close:hover { background: #edf2f7; }
.drawer-body { flex: 1; overflow-y: auto; padding: 0; }
.detail-status-bar { display: flex; align-items: center; gap: 8px; padding: 14px 28px; font-size: 13px; font-weight: 600; }
.detail-status-bar.ready { background: #f0fff4; color: #38a169; }
.detail-status-bar.expired { background: #fff5f5; color: #e53e3e; }
.detail-status-bar.uploading { background: #fffff0; color: #d69e2e; }
.dsb-dot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; }
.dsb-label { margin-right: auto; }
.dsb-type { font-weight: 400; opacity: 0.7; }
.detail-tabs { display: flex; border-bottom: 2px solid #edf2f7; padding: 0 28px; }
.tab-btn { padding: 12px 18px; font-size: 13.5px; font-weight: 600; color: #718096; border: none; background: transparent; cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px; transition: all 0.2s; }
.tab-btn:hover { color: #4a5568; }
.tab-btn.active { color: #667eea; border-bottom-color: #667eea; }
.detail-section { padding: 24px 28px; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.dg-item { display: flex; flex-direction: column; gap: 4px; }
.dg-item.full { grid-column: 1 / -1; }
.dg-label { font-size: 11.5px; color: #a0aec0; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
.dg-value { font-size: 14px; color: #1a202c; }
.dg-value.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 13px; }
.dg-value.break-all { word-break: break-all; }
.build-log-box { margin-top: 20px; border: 1px solid #edf2f7; border-radius: 10px; overflow: hidden; }
.log-header { padding: 10px 16px; background: #f7fafc; font-size: 12px; font-weight: 600; color: #718096; border-bottom: 1px solid #edf2f7; }
.log-content { padding: 16px; margin: 0; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; color: #2d3748; line-height: 1.7; background: #fafafa; max-height: 200px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; }
.drawer-footer { display: flex; gap: 10px; padding: 20px 28px; border-top: 1px solid #edf2f7; background: #f7fafc; }
.drawer-btn { display: inline-flex; align-items: center; gap: 7px; padding: 10px 20px; border-radius: 10px; font-size: 13.5px; font-weight: 600; cursor: pointer; border: none; transition: all 0.2s; }
.drawer-btn svg { width: 15px; height: 15px; }
.drawer-btn.primary { background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; }
.drawer-btn.primary:hover { box-shadow: 0 4px 14px rgba(102,126,234,0.4); }
.drawer-btn.danger-outline { background: transparent; color: #e53e3e; border: 1.5px solid #fed7d7; }
.drawer-btn.danger-outline:hover { background: #fff5f5; }
.drawer-btn.secondary { background: #edf2f7; color: #4a5568; border: 1.5px solid #e2e8f0; }
.drawer-btn.secondary:hover { background: #e2e8f0; }
.drawer-btn.ghost { background: transparent; color: #718096; border: 1.5px solid #e2e8f0; margin-left: auto; }
.drawer-btn.ghost:hover { background: #edf2f7; }

/* ===== 批量删除按钮 ===== */
.batch-del-btn { display: inline-flex; align-items: center; gap: 6px; padding: 7px 16px; border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; border: 1.5px solid #fed7d7; background: #fff5f5; color: #e53e3e; transition: all 0.2s; }
.batch-del-btn:hover { background: #fed7d7; }
.batch-del-btn svg { width: 14px; height: 14px; }
.col-check { width: 40px; text-align: center; }
.col-check input[type="checkbox"] { width: 16px; height: 16px; cursor: pointer; accent-color: #667eea; }

/* ===== 编辑弹窗 ===== */
.edit-modal { width: 520px; max-width: 90vw; background: #fff; border-radius: 16px; margin: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.2); display: flex; flex-direction: column; max-height: 90vh; }
.edit-body { padding: 24px 28px; overflow-y: auto; }
.edit-field { margin-bottom: 16px; }
.edit-field label { display: block; font-size: 12px; font-weight: 600; color: #718096; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
.edit-field input, .edit-field select { width: 100%; padding: 9px 14px; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 13.5px; color: #2d3748; background: #f7fafc; transition: all 0.2s; box-sizing: border-box; }
.edit-field input:focus, .edit-field select:focus { border-color: #667eea; background: #fff; outline: none; box-shadow: 0 0 0 3px rgba(102,126,234,0.12); }
.edit-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.create-section-title { font-size: 13px; font-weight: 700; color: #667eea; margin: 8px 0 12px; padding-top: 12px; border-top: 1px dashed #e2e8f0; text-transform: uppercase; letter-spacing: 0.5px; }

/* ===== 上传区域 ===== */
.upload-zone { border: 2px dashed #d2d6dc; border-radius: 12px; padding: 24px; text-align: center; cursor: pointer; transition: all 0.2s; background: #fafbfc; margin-bottom: 16px; }
.upload-zone:hover { border-color: #667eea; background: #f5f7ff; }
.upload-placeholder { display: flex; flex-direction: column; align-items: center; gap: 8px; color: #a0aec0; }
.upload-placeholder svg { width: 36px; height: 36px; color: #cbd5e0; }
.upload-placeholder span { font-size: 13px; }
.upload-hint { font-size: 11.5px; color: #cbd5e0; }
.upload-selected { display: flex; align-items: center; gap: 12px; }
.upload-selected > svg { width: 32px; height: 32px; color: #667eea; flex-shrink: 0; }
.upload-file-info { flex: 1; text-align: left; min-width: 0; }
.upload-file-name { display: block; font-size: 13.5px; font-weight: 600; color: #2d3748; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.upload-file-size { display: block; font-size: 12px; color: #a0aec0; margin-top: 2px; }
.upload-remove { width: 28px; height: 28px; border: none; background: #fee2e2; border-radius: 7px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; flex-shrink: 0; }
.upload-remove svg { width: 14px; height: 14px; color: #e53e3e; }
.upload-remove:hover { background: #fecaca; }

/* ===== 抽屉动画 ===== */
.drawer-enter-active, .drawer-leave-active { transition: all 0.3s ease; }
.drawer-enter-active .drawer-panel, .drawer-leave-active .drawer-panel { transition: transform 0.3s cubic-bezier(.4,0,.2,1); }
.drawer-enter-from { opacity: 0; }
.drawer-enter-from .drawer-panel { transform: translateX(100%); }
.drawer-leave-to { opacity: 0; }
.drawer-leave-to .drawer-panel { transform: translateX(100%); }

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .page-hero { padding: 20px 16px 18px; }
  .hero-title { font-size: 22px; }
  .overview-section, .main-section { padding: 0 16px 16px; }
  .overview-cards { grid-template-columns: repeat(2, 1fr); }
  .toolbar { flex-wrap: wrap; }
  .toolbar-right { width: 100%; justify-content: space-between; }
  .filter-chips { flex-wrap: wrap; }
  .card-grid { grid-template-columns: 1fr; }
  .drawer-panel { width: 100%; }
}

/* ===== 上传进度浮层 ===== */
.upload-progress-overlay { position: fixed; top: 0; left: 0; right: 0; z-index: 99998; display: flex; justify-content: center; padding-top: 20px; pointer-events: none; }
.upload-progress-card { display: flex; align-items: center; gap: 14px; background: #fff; border-radius: 14px; padding: 16px 24px; box-shadow: 0 8px 32px rgba(0,0,0,0.15); pointer-events: auto; min-width: 340px; border: 1.5px solid rgba(102,126,234,0.2); animation: slideDown 0.3s ease-out; }
@keyframes slideDown { from { transform: translateY(-20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
.upload-progress-icon { width: 42px; height: 42px; border-radius: 12px; background: linear-gradient(135deg, #667eea, #764ba2); display: flex; align-items: center; justify-content: center; flex-shrink: 0; animation: pulse-icon 1.5s ease-in-out infinite; }
.upload-progress-icon svg { width: 22px; height: 22px; color: #fff; }
@keyframes pulse-icon { 0%,100% { transform: scale(1); } 50% { transform: scale(1.08); } }
.upload-progress-body { flex: 1; min-width: 0; }
.upload-progress-title { font-size: 13.5px; font-weight: 600; color: #2d3748; margin-bottom: 8px; }
.upload-progress-bar { height: 8px; background: #edf2f7; border-radius: 4px; overflow: hidden; }
.upload-progress-fill { height: 100%; background: linear-gradient(90deg, #667eea, #764ba2); border-radius: 4px; transition: width 0.3s ease; min-width: 2%; }
.upload-progress-text { font-size: 12px; color: #718096; margin-top: 4px; text-align: right; font-variant-numeric: tabular-nums; }
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>

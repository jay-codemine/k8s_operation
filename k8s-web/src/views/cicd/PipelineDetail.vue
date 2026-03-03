“<template>
  ：
  <div class="pipeline-detail-view">
    <!-- 顶部导航 -->
    <div class="breadcrumb">
      <router-link to="/cicd/pipelines" class="breadcrumb-link">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
        流水线列表
      </router-link>
      <span class="separator">/</span>
      <span class="current">{{ pipeline.name || '加载中...' }}</span>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>正在加载流水线详情...</p>
    </div>

    <template v-else-if="pipeline.id">
      <!-- 流水线头部 -->
      <div class="pipeline-header">
        <div class="header-left">
          <div class="title-row">
            <span :class="['status-indicator', `status-${pipeline.status}`]"></span>
            <h1 class="pipeline-title">{{ pipeline.name }}</h1>
            <span :class="['status-badge', `status-${pipeline.status}`]">
              {{ statusText(pipeline.status) }}
            </span>
          </div>
          <p class="pipeline-desc">{{ pipeline.description || '暂无描述' }}</p>
          <div class="pipeline-meta">
            <div class="meta-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77"/>
              </svg>
              <span>{{ pipeline.git_repo }}</span>
            </div>
            <div class="meta-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="6" y1="3" x2="6" y2="15"/>
                <circle cx="18" cy="6" r="3"/>
                <circle cx="6" cy="18" r="3"/>
                <path d="M18 9a9 9 0 0 1-9 9"/>
              </svg>
              <span>{{ pipeline.git_branch }}</span>
            </div>
          </div>
          <!-- 上次运行信息和快捷操作 -->
          <div class="last-run-row">
            <div class="last-run-info">
              <span class="run-label">上次运行</span>
              <span :class="['run-status-badge', `status-${pipeline.last_run_status}`]">
                {{ runStatusText(pipeline.last_run_status) }}
              </span>
              <span class="run-time">{{ formatDate(pipeline.last_run_time) }}</span>
            </div>
            <div class="last-run-actions">
              <!-- 停止按钮：运行中或pending状态显示 -->
              <button
                v-if="pipeline.status === 'running' || pipeline.last_run_status === 'pending' || pipeline.last_run_status === 'running'"
                class="mini-btn btn-stop"
                @click="handleStop"
                title="停止构建"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="6" y="6" width="12" height="12" rx="2"/>
                </svg>
                停止
              </button>
              <!-- 重新发布按钮：非运行状态显示 -->
              <button
                v-if="pipeline.status !== 'running' && pipeline.last_run_status !== 'running'"
                class="mini-btn btn-rerun"
                @click="handleRun"
                title="重新发布"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="23 4 23 10 17 10"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
                重新发布
              </button>
            </div>
          </div>
        </div>
        <div class="header-actions">
          <button
            class="btn btn-success"
            @click="showRunDialog = true"
          >
            <svg viewBox="0 0 24 24" fill="currentColor">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            {{ pipeline.status === 'running' ? '重新运行' : '运行' }}
          </button>
          <button
            v-if="pipeline.status === 'running' || pipeline.last_run_status === 'pending'"
            class="btn btn-warning"
            @click="handleStop"
          >
            <svg v-if="pipeline.last_run_status === 'pending'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="15" y1="9" x2="9" y2="15"/>
              <line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="6" y="6" width="12" height="12" rx="2"/>
            </svg>
            {{ pipeline.last_run_status === 'pending' ? '取消构建' : '停止构建' }}
          </button>
          <button class="btn btn-outline" @click="handleEdit">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            编辑
          </button>
        </div>
      </div>

      <!-- Tab 导航 -->
      <div class="tab-nav">
        <button
          :class="['tab-btn', { active: activeTab === 'overview' }]"
          @click="activeTab = 'overview'"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <path d="M3 9h18"/>
            <path d="M9 21V9"/>
          </svg>
          概览
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'stages' }]"
          @click="activeTab = 'stages'"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
          </svg>
          执行阶段
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'logs' }]"
          @click="activeTab = 'logs'; loadLogs()"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
          构建日志
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'history' }]"
          @click="activeTab = 'history'; loadHistory()"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          运行历史
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'config' }]"
          @click="activeTab = 'config'"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
          配置
        </button>
      </div>

      <!-- Tab 内容 -->
      <div class="tab-content">
        <!-- 概览 -->
        <div v-if="activeTab === 'overview'" class="overview-tab">
          <!-- 最近运行状态 -->
          <div class="section">
            <h3 class="section-title">最近运行状态</h3>
            <div class="status-cards">
              <div class="status-card">
                <div class="card-icon" :class="`status-${pipeline.last_run_status}`">
                  <svg v-if="pipeline.last_run_status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                    <polyline points="22 4 12 14.01 9 11.01"/>
                  </svg>
                  <svg v-else-if="pipeline.last_run_status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                  </svg>
                  <svg v-else-if="pipeline.last_run_status === 'running'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                </div>
                <div class="card-content">
                  <span class="card-label">运行状态</span>
                  <span class="card-value">{{ runStatusText(pipeline.last_run_status) }}</span>
                </div>
              </div>
              <div class="status-card">
                <div class="card-icon neutral">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                    <line x1="16" y1="2" x2="16" y2="6"/>
                    <line x1="8" y1="2" x2="8" y2="6"/>
                    <line x1="3" y1="10" x2="21" y2="10"/>
                  </svg>
                </div>
                <div class="card-content">
                  <span class="card-label">运行时间</span>
                  <span class="card-value">{{ formatDate(pipeline.last_run_time) }}</span>
                </div>
              </div>
              <div class="status-card">
                <div class="card-icon neutral">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
                  </svg>
                </div>
                <div class="card-content">
                  <span class="card-label">构建号</span>
                  <span class="card-value">#{{ pipeline.last_build_number || '-' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 最近部署版本信息 -->
          <div v-if="pipeline.auto_deploy || pipeline.last_deploy_image" class="section">
            <h3 class="section-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="title-icon">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
              最近部署版本
            </h3>
            <div class="version-info-card">
              <div class="version-grid">
                <div class="version-item">
                  <span class="version-label">部署状态</span>
                  <span :class="['version-value', 'deploy-status', `status-${pipeline.last_deploy_status || 'pending'}`]">
                    {{ deployStatusText(pipeline.last_deploy_status) }}
                  </span>
                </div>
                <div class="version-item">
                  <span class="version-label">部署时间</span>
                  <span class="version-value">{{ pipeline.last_deploy_time ? formatFullDate(pipeline.last_deploy_time) : '-' }}</span>
                </div>
                <div class="version-item full">
                  <span class="version-label">镜像地址</span>
                  <span class="version-value code-text">{{ pipeline.last_deploy_image || '-' }}</span>
                </div>
                <div v-if="pipeline.last_deploy_digest" class="version-item full">
                  <span class="version-label">镜像 Digest</span>
                  <span class="version-value code-text digest">{{ pipeline.last_deploy_digest }}</span>
                </div>
                <div v-if="pipeline.last_deploy_version" class="version-item">
                  <span class="version-label">版本号</span>
                  <span class="version-value tag">{{ pipeline.last_deploy_version }}</span>
                </div>
              </div>
              <div v-if="pipeline.auto_deploy" class="deploy-target-info">
                <div class="target-label">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="22" y1="12" x2="18" y2="12"/>
                    <line x1="6" y1="12" x2="2" y2="12"/>
                    <line x1="12" y1="6" x2="12" y2="2"/>
                    <line x1="12" y1="22" x2="12" y2="18"/>
                  </svg>
                  部署目标
                </div>
                <div class="target-value">
                  {{ pipeline.target_namespace || '-' }} /
                  {{ pipeline.target_workload_kind || 'Deployment' }} /
                  {{ pipeline.target_workload_name || '-' }}
                  <span v-if="pipeline.target_container" class="container-name">
                    (容器: {{ pipeline.target_container }})
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 错误信息展示（只在失败时显示） -->
          <div v-if="latestRun && latestRun.error_message && (latestRun.status === 'failed' || pipeline.last_run_status === 'failed')" class="section error-section">
            <h3 class="section-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="error-icon">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              错误信息
            </h3>
            <div class="error-box">
              <div class="error-content">
                <p class="error-message">{{ latestRun.error_message }}</p>
                <p class="error-time">失败时间: {{ formatFullDate(latestRun.finished_at) }}</p>
              </div>
            </div>
          </div>

          <!-- 快速操作 -->
          <div class="section">
            <h3 class="section-title">快速操作</h3>
            <div class="quick-actions">
              <button class="quick-action-btn" @click="handleRun" :disabled="pipeline.status === 'running'">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polygon points="5 3 19 12 5 21 5 3"/>
                </svg>
                <span>运行流水线</span>
              </button>
              <button class="quick-action-btn" @click="activeTab = 'logs'; loadLogs()">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                </svg>
                <span>查看日志</span>
              </button>
              <button class="quick-action-btn" @click="activeTab = 'history'; loadHistory()">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <polyline points="12 6 12 12 16 14"/>
                </svg>
                <span>运行历史</span>
              </button>
              <button class="quick-action-btn" @click="handleEdit">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
                <span>编辑配置</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 执行阶段 -->
        <div v-if="activeTab === 'stages'" class="stages-tab">
          <!-- 阶段筛选和操作栏 -->
          <div class="stages-toolbar">
            <div class="filter-tabs">
              <button
                :class="['filter-tab', { active: stageFilter === '' }]"
                @click="stageFilter = ''"
              >
                全部
                <span class="filter-count">{{ pipelineStages.length }}</span>
              </button>
              <button
                :class="['filter-tab', { active: stageFilter === 'success' }]"
                @click="stageFilter = 'success'"
              >
                <span class="status-dot success"></span>
                成功
                <span class="filter-count">{{ getStageStatusCount('success') }}</span>
              </button>
              <button
                :class="['filter-tab', { active: stageFilter === 'failed' }]"
                @click="stageFilter = 'failed'"
              >
                <span class="status-dot failed"></span>
                失败
                <span class="filter-count">{{ getStageStatusCount('failed') }}</span>
              </button>
              <button
                :class="['filter-tab', { active: stageFilter === 'running' }]"
                @click="stageFilter = 'running'"
              >
                <span class="status-dot running"></span>
                运行中
                <span class="filter-count">{{ getStageStatusCount('running') }}</span>
              </button>
              <button
                :class="['filter-tab', { active: stageFilter === 'pending' }]"
                @click="stageFilter = 'pending'"
              >
                <span class="status-dot pending"></span>
                待执行
                <span class="filter-count">{{ getStageStatusCount('pending') }}</span>
              </button>
            </div>
            <button class="toolbar-btn" @click="loadStages" :disabled="stagesLoading">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              {{ stagesLoading ? '加载中...' : '刷新' }}
            </button>
          </div>

          <!-- 加载状态 -->
          <div v-if="stagesLoading && pipelineStages.length === 0" class="stages-loading">
            <div class="loading-spinner"></div>
            <p>正在加载阶段数据...</p>
          </div>

          <!-- 阶段流水线视图（Jenkins Blue Ocean 风格） -->
          <div v-else-if="filteredStages.length > 0" class="stages-pipeline-container">
            <div class="stages-pipeline-track">
              <div
                v-for="(stage, index) in filteredStages"
                :key="stage.name"
                :class="['pipeline-stage', `status-${stage.status}`, { selected: selectedStage && selectedStage.name === stage.name, 'is-first': index === 0, 'is-last': index === filteredStages.length - 1 }]"
              >
                <!-- 连接线（左侧） -->
                <div v-if="index > 0" :class="['stage-connector-line', 'left', getConnectorStatus(index - 1)]"></div>
                
                <!-- 阶段卡片 -->
                <div class="stage-card" @click="selectStage(stage)">
                  <!-- 状态图标 -->
                  <div :class="['stage-status-icon', `status-${stage.status}`, { 'status-running': stage.status === 'deploying' }]">
                    <svg v-if="stage.status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                      <polyline points="20 6 9 17 4 12"/>
                    </svg>
                    <svg v-else-if="stage.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                      <line x1="18" y1="6" x2="6" y2="18"/>
                      <line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                    <div v-else-if="stage.status === 'running' || stage.status === 'deploying'" class="running-spinner"></div>
                    <svg v-else-if="stage.status === 'waiting'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <polyline points="12 6 12 12 16 14"/>
                    </svg>
                    <div v-else class="pending-dot"></div>
                  </div>
                  
                  <!-- 阶段名称 -->
                  <div class="stage-label">{{ stage.name }}</div>
                  
                  <!-- 耗时标签 -->
                  <div :class="['stage-duration-badge', `status-${stage.status}`, { 'status-running': stage.status === 'deploying' }]">
                    <span v-if="stage.status === 'running' || stage.status === 'deploying'" class="duration-running">
                      <span class="running-dot"></span>
                      {{ calculateStageDuration(stage) || (stage.status === 'deploying' ? '部署中' : '运行中') }}
                    </span>
                    <span v-else-if="stage.status === 'success' || stage.status === 'failed'">{{ stage.duration || calculateStageDuration(stage) || '-' }}</span>
                    <span v-else>-</span>
                  </div>
                </div>
                
                <!-- 连接线（右侧） -->
                <div v-if="index < filteredStages.length - 1" :class="['stage-connector-line', 'right', getConnectorStatus(index)]"></div>
              </div>
            </div>
            
            <!-- 执行进度指示器 -->
            <div v-if="hasRunningStage" class="pipeline-progress-indicator">
              <div class="progress-text">
                <span class="progress-icon"></span>
                正在执行: {{ currentRunningStage?.name || '-' }}
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-else class="stages-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            <p>{{ stageFilter ? '没有匹配的阶段' : '暂无阶段数据，请运行流水线' }}</p>
          </div>

          <!-- 失败阶段错误信息摘要（参考 Jenkins Blue Ocean 风格） -->
          <div v-if="failedStages.length > 0" class="failed-stages-summary">
            <div class="failed-summary-header">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="15" y1="9" x2="9" y2="15"/>
                <line x1="9" y1="9" x2="15" y2="15"/>
              </svg>
              <span>失败阶段 ({{ failedStages.length }})</span>
            </div>
            <div class="failed-stages-list">
              <div v-for="stage in failedStages" :key="stage.id" class="failed-stage-item" @click="selectStage(stage)">
                <div class="failed-stage-info">
                  <span class="failed-stage-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="15" y1="9" x2="9" y2="15"/>
                      <line x1="9" y1="9" x2="15" y2="15"/>
                    </svg>
                  </span>
                  <div class="failed-stage-content">
                    <span class="failed-stage-name">{{ stage.name }}</span>
                    <span class="failed-stage-status">失败</span>
                  </div>
                  <span class="failed-stage-duration">{{ stage.duration || '-' }}</span>
                </div>
                <div v-if="stage.error_message || stage.error_msg" class="failed-stage-error">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                    <line x1="12" y1="9" x2="12" y2="13"/>
                    <line x1="12" y1="17" x2="12.01" y2="17"/>
                  </svg>
                  <span class="error-text">{{ stage.error_message || stage.error_msg }}</span>
                </div>
                <div v-else class="failed-stage-error no-message">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                  <span class="error-text">点击查看构建日志获取详细错误信息</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 选中阶段详情（点击阶段卡片后展示） -->
          <div v-if="selectedStage" :class="['selected-stage-detail', { expanded: stageDetailExpanded }]">
            <!-- 可点击的头部（展开/收起） -->
            <div class="detail-header clickable" @click="toggleStageDetail">
              <span :class="['status-dot', `status-${selectedStage.status}`]"></span>
              <span class="stage-title">{{ selectedStage.name }}</span>
              <!-- 阶段类型标签 -->
              <span v-if="selectedStage.type === 'approval'" :class="['stage-type-badge', 'approval', `approval-${selectedStage.status}`]">{{ approvalBadgeText(selectedStage.status) }}</span>
              <span v-if="selectedStage.type === 'deploy'" class="stage-type-badge deploy">部署</span>
              <span class="stage-status">{{ stageStatusText(selectedStage.status) }}</span>
              <span class="stage-duration-tag">{{ selectedStage.duration || calculateStageDuration(selectedStage) || '-' }}</span>
              
              <!-- 展开/收起指示器 -->
              <div class="expand-indicator">
                <span class="expand-hint">{{ stageDetailExpanded ? '收起详情' : '展开详情' }}</span>
                <svg :class="['expand-icon', { rotated: stageDetailExpanded }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              
              <!-- 部署成功时在头部直接显示回滚按钮 -->
              <button 
                v-if="selectedStage.type === 'deploy' && selectedStage.status === 'success'" 
                class="btn btn-rollback-mini" 
                @click.stop="handleRollback(selectedStage)" 
                :disabled="rollingBack"
                title="回滚到上一版本"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="1 4 1 10 7 10"/>
                  <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
                </svg>
                {{ rollingBack ? '回滚中...' : '回滚' }}
              </button>
              
              <!-- 部署进行中时显示取消按钮 -->
              <button 
                v-if="selectedStage.type === 'deploy' && (selectedStage.status === 'running' || selectedStage.status === 'deploying')" 
                class="btn btn-cancel-mini" 
                @click.stop="handleCancelDeploy(selectedStage)" 
                :disabled="cancelling"
                title="取消部署"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="15" y1="9" x2="9" y2="15"/>
                  <line x1="9" y1="9" x2="15" y2="15"/>
                </svg>
                {{ cancelling ? '取消中...' : '取消' }}
              </button>
            </div>
            
            <!-- 可展开/收起的详情内容 -->
            <transition name="slide-fade">
              <div v-show="stageDetailExpanded" class="detail-body">
              <!-- 审批阶段操作 -->
              <div v-if="selectedStage.type === 'approval' && (selectedStage.status === 'waiting' || selectedStage.status === 'pending')" class="stage-action-panel approval-panel-enhanced">
                <div class="approval-header">
                  <div class="approval-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                    </svg>
                  </div>
                  <div class="approval-title">
                    <h4>人工审批</h4>
                    <p>该阶段需要人工审批确认才能继续部署</p>
                  </div>
                </div>
                <div class="approval-options">
                  <label :class="['approval-option', { selected: approvalDecision === 'approve' }]" @click="approvalDecision = 'approve'">
                    <div class="option-radio"><span class="radio-inner"></span></div>
                    <div class="option-content">
                      <span class="option-icon approve"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg></span>
                      <span class="option-label">通过</span>
                      <span class="option-desc">确认并继续执行部署</span>
                    </div>
                  </label>
                  <label :class="['approval-option', { selected: approvalDecision === 'reject' }]" @click="approvalDecision = 'reject'">
                    <div class="option-radio"><span class="radio-inner"></span></div>
                    <div class="option-content">
                      <span class="option-icon reject"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></span>
                      <span class="option-label">拒绝</span>
                      <span class="option-desc">取消本次部署</span>
                    </div>
                  </label>
                </div>
                <div class="approval-comment">
                  <label class="comment-label">审批备注 <span class="optional">(可选)</span></label>
                  <textarea v-model="approvalComment" class="comment-input" placeholder="请输入审批备注..." rows="3"></textarea>
                </div>
                <div class="approval-actions">
                  <button :class="['btn', 'btn-approval', approvalDecision === 'approve' ? 'approve' : 'reject']" @click.stop="submitApproval(selectedStage.id)" :disabled="approving">
                    <svg v-if="approving" class="loading-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/></svg>
                    <svg v-else-if="approvalDecision === 'approve'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
                    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    {{ approving ? '处理中...' : (approvalDecision === 'approve' ? '确认通过' : '确认拒绝') }}
                  </button>
                </div>
              </div>

              <!-- 审批已通过 -->
              <div v-if="selectedStage.type === 'approval' && (selectedStage.status === 'success' || selectedStage.status === 'approved')" class="stage-action-panel approval-result-panel approved">
                <div class="approval-result-header">
                  <div class="result-icon approved"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></div>
                  <div class="result-content"><h4>审批已通过</h4><p>该阶段已经审批通过，可以继续执行部署</p></div>
                </div>
                <div v-if="selectedStage.approval_info" class="approval-meta">
                  <span v-if="selectedStage.approval_info.approver_name">审批人: {{ selectedStage.approval_info.approver_name }}</span>
                  <span v-if="selectedStage.approval_info.approved_at">审批时间: {{ formatFullDate(selectedStage.approval_info.approved_at) }}</span>
                </div>
                <div v-if="selectedStage.approval_info && selectedStage.approval_info.comment" class="approval-comment-display">
                  <span class="comment-label">审批备注:</span>
                  <span class="comment-text">{{ selectedStage.approval_info.comment }}</span>
                </div>
              </div>

              <!-- 审批已拒绝 -->
              <div v-if="selectedStage.type === 'approval' && (selectedStage.status === 'failed' || selectedStage.status === 'rejected')" class="stage-action-panel approval-result-panel rejected">
                <div class="approval-result-header">
                  <div class="result-icon rejected"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg></div>
                  <div class="result-content"><h4>审批已拒绝</h4><p>该阶段审批被拒绝，部署已取消</p></div>
                </div>
                <div v-if="selectedStage.approval_info" class="approval-meta">
                  <span v-if="selectedStage.approval_info.approver_name">拒绝人: {{ selectedStage.approval_info.approver_name }}</span>
                  <span v-if="selectedStage.approval_info.approved_at">拒绝时间: {{ formatFullDate(selectedStage.approval_info.approved_at) }}</span>
                </div>
                <div v-if="selectedStage.approval_info && selectedStage.approval_info.comment" class="approval-comment-display">
                  <span class="comment-label">拒绝原因:</span>
                  <span class="comment-text">{{ selectedStage.approval_info.comment }}</span>
                </div>
              </div>

              <!-- 部署阶段操作 -->
              <div v-if="selectedStage.type === 'deploy' && selectedStage.can_operate" class="stage-action-panel deploy-panel">
                <div v-if="selectedStage.config_warning" class="config-warning">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                  <div class="warning-content">
                    <span class="warning-text">{{ selectedStage.config_warning }}</span>
                    <router-link :to="`/cicd/pipelines/${pipeline.id}/edit`" class="config-link">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                      去配置
                    </router-link>
                  </div>
                </div>
                <div v-else class="action-info">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
                  <span>点击执行部署到 K8s 集群</span>
                </div>
                <div v-if="selectedStage.deploy_info && !selectedStage.config_warning" class="deploy-info">
                  <span>集群: {{ selectedStage.deploy_info.cluster_name || selectedStage.deploy_info.cluster_id || '-' }}</span>
                  <span>命名空间: {{ selectedStage.deploy_info.namespace || '-' }}</span>
                  <span>工作负载: {{ selectedStage.deploy_info.workload_name || '-' }}</span>
                  <span>镜像: {{ selectedStage.deploy_info.image || '-' }}</span>
                </div>
                <div class="action-buttons">
                  <button class="btn btn-primary" @click.stop="handleDeployStage(selectedStage.id)" :disabled="deploying || !!selectedStage.config_warning">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                    {{ deploying ? '部署中...' : '执行部署' }}
                  </button>
                </div>
              </div>

              <!-- 部署成功 -->
              <div v-if="selectedStage.type === 'deploy' && selectedStage.status === 'success' && selectedStage.deploy_info" class="deploy-success-info">
                <div class="success-badge"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>部署成功</div>
                
                <!-- 版本变更信息（参考 Rancher/Kuboard 风格） -->
                <div v-if="selectedStage.deploy_info.old_image && selectedStage.deploy_info.old_image !== selectedStage.deploy_info.image" class="version-change-card">
                  <div class="version-change-header">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                    <span>版本变更</span>
                  </div>
                  <div class="version-change-content">
                    <div class="version-item old">
                      <span class="version-label">旧版本</span>
                      <span class="version-value">{{ selectedStage.deploy_info.old_image }}</span>
                    </div>
                    <div class="version-arrow">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
                    </div>
                    <div class="version-item new">
                      <span class="version-label">新版本</span>
                      <span class="version-value">{{ selectedStage.deploy_info.image }}</span>
                    </div>
                  </div>
                </div>
                
                <div class="deploy-details">
                  <span>集群: {{ selectedStage.deploy_info.cluster_name || selectedStage.deploy_info.cluster_id }}</span>
                  <span>命名空间: {{ selectedStage.deploy_info.namespace }}</span>
                  <span>工作负载: {{ selectedStage.deploy_info.workload_name }}</span>
                  <span>镜像: {{ selectedStage.deploy_info.image }}</span>
                </div>
                <div class="deploy-actions rollback-actions">
                  <button class="btn btn-rollback" @click.stop="quickRollbackToPrevious(selectedStage)" :disabled="rollingBack">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                    {{ rollingBack ? '回滚中...' : '回滚到上一版本' }}
                  </button>
                  <button class="btn btn-rollback-select" @click.stop="handleRollback(selectedStage)" :disabled="rollingBack">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/><line x1="9" y1="17" x2="13" y2="17"/></svg>
                    指定版本回滚
                  </button>
                </div>
              </div>

              <!-- 部署进行中 -->
              <div v-if="selectedStage.type === 'deploy' && (selectedStage.status === 'running' || selectedStage.status === 'deploying')" class="deploy-progress-panel">
                <div class="progress-header"><div class="progress-spinner"></div><span>部署进行中...</span></div>
                
                <!-- 版本变更信息 -->
                <div v-if="selectedStage.deploy_info && selectedStage.deploy_info.old_image" class="version-change-mini">
                  <span class="old-version">{{ selectedStage.deploy_info.old_image }}</span>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="arrow-icon"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
                  <span class="new-version">{{ selectedStage.deploy_info.image }}</span>
                </div>
                
                <div v-if="selectedStage.deploy_info" class="deploy-info-mini">
                  <span>工作负载: {{ selectedStage.deploy_info.workload_name }}</span>
                  <span>镜像: {{ selectedStage.deploy_info.image }}</span>
                </div>
                
                <!-- 实时部署日志预览 -->
                <div v-if="selectedStage.logs" class="deploy-logs-preview">
                  <div class="logs-preview-header">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                    <span>部署日志</span>
                  </div>
                  <pre class="logs-preview-content">{{ selectedStage.logs }}</pre>
                </div>
                
                <div class="deploy-actions">
                  <button class="btn btn-cancel" @click.stop="handleCancelDeploy(selectedStage)" :disabled="cancelling">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
                    {{ cancelling ? '取消中...' : '取消部署' }}
                  </button>
                </div>
              </div>

              <!-- 部署失败 -->
              <div v-if="selectedStage.type === 'deploy' && selectedStage.status === 'failed'" class="deploy-failed-info">
                <div class="failed-badge"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>部署失败</div>
                <div v-if="selectedStage.error_message" class="failed-reason">
                  <span class="reason-label">失败原因:</span>
                  <span class="reason-text">{{ selectedStage.error_message }}</span>
                </div>
                <div v-if="selectedStage.deploy_info" class="deploy-details">
                  <span>集群: {{ selectedStage.deploy_info.cluster_name || selectedStage.deploy_info.cluster_id }}</span>
                  <span>命名空间: {{ selectedStage.deploy_info.namespace }}</span>
                  <span>工作负载: {{ selectedStage.deploy_info.workload_name }}</span>
                  <span>镜像: {{ selectedStage.deploy_info.image }}</span>
                </div>
                <div class="retry-deploy-actions">
                  <button class="btn btn-retry" @click.stop="handleRetryDeploy(selectedStage.id)" :disabled="deploying">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                    {{ deploying ? '部署中...' : '重新部署' }}
                  </button>
                </div>
              </div>

              <!-- 部署日志 -->
              <div v-if="selectedStage.type === 'deploy' && selectedStage.logs && (selectedStage.status === 'success' || selectedStage.status === 'failed' || selectedStage.status === 'running' || selectedStage.status === 'deploying')" class="deploy-logs-panel">
                <div class="logs-toggle" @click="selectedStage.showLogs = !selectedStage.showLogs">
                  <svg :class="['toggle-icon', { expanded: selectedStage.showLogs }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
                  <span>查看部署日志</span>
                </div>
                <pre v-show="selectedStage.showLogs" class="deploy-logs-content">{{ selectedStage.logs }}</pre>
              </div>

              <!-- 非审批/部署阶段的详细信息（参考 Rancher/Kuboard/KubeSphere 风格） -->
              <div v-if="selectedStage.type !== 'approval' && selectedStage.type !== 'deploy'" class="stage-detail-card">
                <!-- 阶段概览卡片 -->
                <div :class="['stage-overview-card', `status-${selectedStage.status}`]">
                  <div class="overview-header">
                    <div class="overview-icon">
                      <svg v-if="selectedStage.status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
                      <svg v-else-if="selectedStage.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                      <svg v-else-if="selectedStage.status === 'running'" class="spinning" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
                      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/></svg>
                    </div>
                    <div class="overview-title">
                      <h4>{{ selectedStage.name }}</h4>
                      <span :class="['status-label', `status-${selectedStage.status}`]">{{ stageStatusText(selectedStage.status) }}</span>
                    </div>
                    <div class="overview-duration">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                      <span>{{ selectedStage.duration || calculateStageDuration(selectedStage) || '-' }}</span>
                    </div>
                  </div>
                </div>

                <!-- 阶段信息网格（Kuboard 风格） -->
                <div class="stage-info-grid">
                  <div class="info-card">
                    <div class="info-card-icon type">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
                    </div>
                    <div class="info-card-content">
                      <span class="info-card-label">阶段类型</span>
                      <span class="info-card-value">{{ getStageTypeName(selectedStage.type) }}</span>
                    </div>
                  </div>
                  <div class="info-card">
                    <div class="info-card-icon time">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                    </div>
                    <div class="info-card-content">
                      <span class="info-card-label">开始时间</span>
                      <span class="info-card-value">{{ getStageStartTime(selectedStage) }}</span>
                    </div>
                  </div>
                  <div class="info-card">
                    <div class="info-card-icon time">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                    </div>
                    <div class="info-card-content">
                      <span class="info-card-label">结束时间</span>
                      <span class="info-card-value">{{ getStageEndTime(selectedStage) }}</span>
                    </div>
                  </div>
                  <div class="info-card">
                    <div :class="['info-card-icon', 'duration', selectedStage.status === 'running' ? 'running' : '']">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
                    </div>
                    <div class="info-card-content">
                      <span class="info-card-label">执行耗时</span>
                      <span :class="['info-card-value', 'duration', selectedStage.status === 'running' ? 'running' : '']">{{ selectedStage.duration || calculateStageDuration(selectedStage) || '-' }}</span>
                    </div>
                  </div>
                </div>

                <!-- 运行中进度（KubeSphere 风格） -->
                <div v-if="selectedStage.status === 'running'" class="stage-progress-panel">
                  <div class="progress-header">
                    <div class="progress-indicator"></div>
                    <span>正在执行 {{ selectedStage.name }}...</span>
                  </div>
                  <div class="progress-bar-container">
                    <div class="progress-bar-track">
                      <div class="progress-bar-fill"></div>
                    </div>
                  </div>
                  <div class="progress-tips">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                    <span>可以切换到「构建日志」标签查看实时日志输出</span>
                  </div>
                </div>

                <!-- 失败错误信息（Rancher 风格） -->
                <div v-if="selectedStage.status === 'failed'" class="stage-error-panel">
                  <div class="error-panel-header">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                    <span>阶段执行失败</span>
                  </div>
                  <div class="error-panel-content">
                    <div v-if="selectedStage.error_message || selectedStage.error_msg" class="error-message-box">
                      <div class="error-label">错误信息</div>
                      <pre class="error-text">{{ selectedStage.error_message || selectedStage.error_msg }}</pre>
                    </div>
                    <div v-else class="error-message-box no-detail">
                      <div class="error-label">错误详情</div>
                      <p class="error-hint">请查看构建日志获取详细错误信息</p>
                    </div>
                  </div>
                  <div class="error-panel-actions">
                    <button class="btn btn-view-logs" @click="activeTab = 'logs'; loadLogs()">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
                      查看完整日志
                    </button>
                    <button class="btn btn-retry-stage" @click="handleRun" :disabled="pipeline.status === 'running'">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                      重新运行
                    </button>
                  </div>
                </div>

                <!-- 成功信息（KubeSphere 风格） -->
                <div v-if="selectedStage.status === 'success'" class="stage-success-panel">
                  <div class="success-panel-header">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                    <span>阶段执行成功</span>
                  </div>
                  <div v-if="selectedStage.type === 'push' && latestRun && latestRun.image_url" class="success-artifact">
                    <div class="artifact-label">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                      构建产物
                    </div>
                    <div class="artifact-content">
                      <span class="artifact-image">{{ latestRun.image_url }}</span>
                      <button class="copy-btn" @click="copyText(latestRun.image_url)" title="复制镜像地址">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- 日志预览区域（Rancher 风格） -->
                <div class="stage-logs-preview">
                  <div class="logs-preview-header" @click="toggleStageLogs">
                    <svg :class="['toggle-arrow', { expanded: showStageLogs }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                    <span>执行日志</span>
                    <button class="view-full-logs" @click.stop="activeTab = 'logs'; loadLogs()">查看完整日志</button>
                  </div>
                  <div v-show="showStageLogs" class="logs-preview-content">
                    <pre v-if="logs">{{ getStageLogsPreview() }}</pre>
                    <div v-else class="logs-preview-empty">
                      <span>暂无日志，请运行流水线后查看</span>
                    </div>
                  </div>
                </div>
              </div>
              </div>
            </transition>
          </div>

          <!-- 未选中阶段提示 -->
          <div v-else-if="filteredStages.length > 0" class="no-stage-selected">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"/>
            </svg>
            <p>点击上方阶段卡片查看详情</p>
          </div>
        </div>

        <!-- 构建日志 -->
        <div v-if="activeTab === 'logs'" class="logs-tab">
          <div class="logs-toolbar">
            <div class="toolbar-left">
              <span class="log-label">构建号: #{{ pipeline.last_build_number || '-' }}</span>
            </div>
            <div class="toolbar-right">
              <button class="toolbar-btn" @click="refreshLogs" :disabled="logsLoading">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="23 4 23 10 17 10"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
                刷新
              </button>
              <button class="toolbar-btn" @click="copyLogs">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                </svg>
                复制
              </button>
              <button class="toolbar-btn" @click="downloadLogs">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                  <polyline points="7 10 12 15 17 10"/>
                  <line x1="12" y1="15" x2="12" y2="3"/>
                </svg>
                下载
              </button>
              <label class="auto-scroll">
                <input type="checkbox" v-model="autoScroll" />
                自动滚动
              </label>
            </div>
          </div>
          <div class="logs-container" ref="logsContainer">
            <div v-if="logsLoading" class="logs-loading">
              <div class="loading-spinner small"></div>
              正在加载日志...
            </div>
            <pre v-else-if="logs" class="logs-content">{{ logs }}</pre>
            <!-- 错误状态：显示友好提示和重新构建按钮 -->
            <div v-else-if="logsError" class="logs-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              <p class="error-message">{{ logsError }}</p>
              <button class="btn btn-primary" @click="handleRun" :disabled="pipeline.status === 'running'">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="23 4 23 10 17 10"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
                重新运行流水线
              </button>
            </div>
            <div v-else class="logs-empty">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              <p>暂无日志，请先运行流水线</p>
            </div>
          </div>
        </div>

        <!-- 运行历史 -->
        <div v-if="activeTab === 'history'" class="history-tab">
          <div class="history-toolbar">
            <!-- 状态筛选按钮 -->
            <div class="filter-tabs">
              <button
                :class="['filter-tab', { active: historyFilter === '' }]"
                @click="historyFilter = ''"
              >
                全部
                <span class="filter-count">{{ history.length }}</span>
              </button>
              <button
                :class="['filter-tab', { active: historyFilter === 'success' }]"
                @click="historyFilter = 'success'"
              >
                <span class="status-dot success"></span>
                成功
                <span class="filter-count">{{ getHistoryStatusCount('success') }}</span>
              </button>
              <button
                :class="['filter-tab', { active: historyFilter === 'failed' }]"
                @click="historyFilter = 'failed'"
              >
                <span class="status-dot failed"></span>
                失败
                <span class="filter-count">{{ getHistoryStatusCount('failed') }}</span>
              </button>
              <button
                :class="['filter-tab', { active: historyFilter === 'running' }]"
                @click="historyFilter = 'running'"
              >
                <span class="status-dot running"></span>
                运行中
                <span class="filter-count">{{ getHistoryStatusCount('running') }}</span>
              </button>
            </div>
            <button class="toolbar-btn" @click="loadHistory" :disabled="historyLoading">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              刷新
            </button>
          </div>
          <div v-if="historyLoading" class="history-loading">
            <div class="loading-spinner small"></div>
            正在加载运行历史...
          </div>
          <div v-else-if="filteredHistory.length === 0" class="history-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            <p>{{ historyFilter ? '没有匹配的运行记录' : '暂无运行历史' }}</p>
          </div>
          <div v-else class="history-list">
            <div
              v-for="run in filteredHistory"
              :key="run.id"
              :class="['history-item', `status-${run.status}`, { expanded: expandedRunId === run.id }]"
            >
              <!-- 主体内容区域（可点击展开） -->
              <div class="history-item-main" @click="toggleRunDetail(run)">
                <div class="history-icon">
                  <svg v-if="run.status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                    <polyline points="22 4 12 14.01 9 11.01"/>
                  </svg>
                  <svg v-else-if="run.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                  </svg>
                  <svg v-else-if="run.status === 'running'" class="spinning" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                  </svg>
                </div>
                <div class="history-info">
                  <div class="history-title">
                    <span class="build-number">#{{ run.build_number || run.id }}</span>
                    <span :class="['history-status', `status-${run.status}`]">{{ runStatusText(run.status) }}</span>
                  </div>
                  <div class="history-meta">
                    <span>{{ formatDate(run.started_at || run.created_at) }}</span>
                    <span v-if="run.duration_sec">· 耗时 {{ formatDuration(run.duration_sec) }}</span>
                    <span v-if="run.git_branch">· {{ run.git_branch }}</span>
                  </div>
                </div>
                <div class="history-expand-icon">
                  <svg :class="{ 'rotated': expandedRunId === run.id }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="6 9 12 15 18 9"/>
                  </svg>
                </div>
              </div>
              
              <!-- 详情展开区域 -->
              <div v-if="expandedRunId === run.id" class="history-detail">
                <div class="detail-grid">
                  <div class="detail-item">
                    <span class="detail-label">构建号</span>
                    <span class="detail-value">#{{ run.build_number || run.id }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">状态</span>
                    <span :class="['detail-value', 'status-badge', `status-${run.status}`]">{{ runStatusText(run.status) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">开始时间</span>
                    <span class="detail-value">{{ formatFullDate(run.started_at || run.created_at) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">结束时间</span>
                    <span class="detail-value">{{ run.finished_at ? formatFullDate(run.finished_at) : '-' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">耗时</span>
                    <span class="detail-value">{{ run.duration_sec ? formatDuration(run.duration_sec) : '-' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">Git 分支</span>
                    <span class="detail-value mono">{{ run.git_branch || pipeline.git_branch || '-' }}</span>
                  </div>
                  <div v-if="run.trigger_type" class="detail-item">
                    <span class="detail-label">触发方式</span>
                    <span class="detail-value">{{ run.trigger_type === 'manual' ? '手动触发' : (run.trigger_type === 'webhook' ? 'Webhook' : run.trigger_type) }}</span>
                  </div>
                  <div v-if="run.triggered_by" class="detail-item">
                    <span class="detail-label">触发人</span>
                    <span class="detail-value">{{ run.triggered_by }}</span>
                  </div>
                </div>
                
                <!-- 错误信息 -->
                <div v-if="run.status === 'failed' && (run.error_message || run.error_msg)" class="detail-error">
                  <div class="error-title">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="12" y1="8" x2="12" y2="12"/>
                      <line x1="12" y1="16" x2="12.01" y2="16"/>
                    </svg>
                    错误信息
                  </div>
                  <div class="error-content">{{ run.error_message || run.error_msg }}</div>
                </div>
                
                <!-- 操作按钮 -->
                <div class="detail-actions">
                  <button class="detail-btn primary" @click.stop="viewRunLogs(run)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                    </svg>
                    查看日志
                  </button>
                  <button
                    v-if="run.status === 'failed'"
                    class="detail-btn"
                    @click.stop="retryRun(run)"
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="23 4 23 10 17 10"/>
                      <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                    </svg>
                    重试
                  </button>
                </div>
              </div>
              
              <!-- 右侧快捷操作按钮 -->
              <div class="history-actions" @click.stop>
                <button class="action-btn" @click="viewRunLogs(run)" title="查看日志">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                  </svg>
                </button>
                <button
                  v-if="run.status === 'failed'"
                  class="action-btn retry"
                  @click="retryRun(run)"
                  title="重试"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="23 4 23 10 17 10"/>
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 配置 -->
        <div v-if="activeTab === 'config'" class="config-tab">
          <div class="config-section">
            <h3 class="config-title">基本信息</h3>
            <div class="config-grid">
              <div class="config-item">
                <span class="config-label">流水线名称</span>
                <span class="config-value">{{ pipeline.name }}</span>
              </div>
              <div class="config-item">
                <span class="config-label">描述</span>
                <span class="config-value">{{ pipeline.description || '-' }}</span>
              </div>
              <div class="config-item">
                <span class="config-label">创建时间</span>
                <span class="config-value">{{ formatFullDate(pipeline.created_at) }}</span>
              </div>
            </div>
          </div>

          <div class="config-section">
            <h3 class="config-title">Git 配置</h3>
            <div class="config-grid">
              <div class="config-item">
                <span class="config-label">仓库地址</span>
                <span class="config-value code">{{ pipeline.git_repo }}</span>
              </div>
              <div class="config-item">
                <span class="config-label">分支</span>
                <span class="config-value">{{ pipeline.git_branch }}</span>
              </div>
            </div>
          </div>

          <div class="config-section">
            <h3 class="config-title">Jenkins 配置</h3>
            <div class="config-grid">
              <div class="config-item">
                <span class="config-label">Jenkins URL</span>
                <span class="config-value code">{{ pipeline.jenkins_url || '使用全局配置' }}</span>
              </div>
              <div class="config-item">
                <span class="config-label">Job 名称</span>
                <span class="config-value">{{ pipeline.jenkins_job }}</span>
              </div>
            </div>
          </div>

          <div class="config-section" v-if="pipeline.env_vars && pipeline.env_vars.length">
            <h3 class="config-title">环境变量</h3>
            <div class="env-vars-list">
              <div v-for="env in pipeline.env_vars" :key="env.name" class="env-var-item">
                <span class="env-name">{{ env.name }}</span>
                <span class="env-value">{{ env.value }}</span>
              </div>
            </div>
          </div>

          <div class="config-section" v-if="pipeline.deploy_config">
            <h3 class="config-title">部署策略配置</h3>
            <pre class="config-json">{{ JSON.stringify(pipeline.deploy_config, null, 2) }}</pre>
          </div>

          <!-- 自动部署配置 -->
          <div class="config-section" v-if="pipeline.auto_deploy">
            <h3 class="config-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="title-icon">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v6l4 2"/>
              </svg>
              自动部署配置
            </h3>
            <div class="auto-deploy-config-display">
              <div class="config-grid">
                <div class="config-item">
                  <span class="config-label">自动部署</span>
                  <span class="config-value">
                    <span class="status-badge enabled">已启用</span>
                  </span>
                </div>
                <div class="config-item">
                  <span class="config-label">部署环境</span>
                  <span :class="['config-value', 'env-badge', `env-${pipeline.deploy_env}`]">
                    {{ envLabel(pipeline.deploy_env) }}
                  </span>
                </div>
                <div class="config-item">
                  <span class="config-label">目标集群 ID</span>
                  <span class="config-value">{{ pipeline.target_cluster_id || '-' }}</span>
                </div>
                <div class="config-item">
                  <span class="config-label">目标命名空间</span>
                  <span class="config-value">{{ pipeline.target_namespace || '-' }}</span>
                </div>
                <div class="config-item">
                  <span class="config-label">工作负载类型</span>
                  <span class="config-value">{{ pipeline.target_workload_kind || 'Deployment' }}</span>
                </div>
                <div class="config-item">
                  <span class="config-label">工作负载名称</span>
                  <span class="config-value">{{ pipeline.target_workload_name || '-' }}</span>
                </div>
                <div class="config-item">
                  <span class="config-label">容器名称</span>
                  <span class="config-value">{{ pipeline.target_container || '默认第一个' }}</span>
                </div>
                <div class="config-item">
                  <span class="config-label">需要审批</span>
                  <span class="config-value">
                    <span v-if="pipeline.require_approval" class="status-badge warning">是</span>
                    <span v-else class="status-badge">否</span>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- 错误状态 -->
    <div v-else class="error-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      <h3>加载失败</h3>
      <p>{{ errorMsg || '无法加载流水线详情' }}</p>
      <button class="btn btn-primary" @click="loadPipeline">重试</button>
    </div>

    <!-- 运行配置弹窗 -->
    <div v-if="showRunDialog" class="modal-overlay" @click.self="showRunDialog = false">
      <div class="modal-content run-dialog">
        <div class="modal-header">
          <h3>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            运行流水线
          </h3>
          <button class="close-btn" @click="showRunDialog = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <!-- 基本信息 -->
          <div class="run-info">
            <div class="info-item">
              <span class="info-label">流水线</span>
              <span class="info-value">{{ pipeline.name }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Git 分支</span>
              <span class="info-value">{{ pipeline.git_branch }}</span>
            </div>
          </div>

          <!-- 自动部署配置展示 -->
          <div v-if="pipeline.auto_deploy" class="deploy-config-section">
            <h4 class="section-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v6l4 2"/>
              </svg>
              自动部署配置
            </h4>
            <div class="deploy-config-info">
              <div class="config-row">
                <span class="config-key">部署环境</span>
                <span :class="['config-val', `env-${pipeline.deploy_env}`]">{{ envLabel(pipeline.deploy_env) }}</span>
              </div>
              <div class="config-row">
                <span class="config-key">目标集群</span>
                <span class="config-val">{{ pipeline.target_cluster_id || '默认集群' }}</span>
              </div>
              <div class="config-row">
                <span class="config-key">命名空间</span>
                <span class="config-val">{{ pipeline.target_namespace || '-' }}</span>
              </div>
              <div class="config-row">
                <span class="config-key">工作负载</span>
                <span class="config-val">{{ pipeline.target_workload_kind || 'Deployment' }} / {{ pipeline.target_workload_name || '-' }}</span>
              </div>
              <div v-if="pipeline.require_approval" class="approval-notice">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="8" x2="12" y2="12"/>
                  <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                生产环境部署需要审批确认
              </div>
            </div>
          </div>

          <div v-else class="no-deploy-notice">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="16" x2="12" y2="12"/>
              <line x1="12" y1="8" x2="12.01" y2="8"/>
            </svg>
            <span>未配置自动部署，构建完成后需手动部署</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRunDialog = false">取消</button>
          <button class="btn btn-success" @click="confirmRun" :disabled="runSubmitting">
            <svg v-if="!runSubmitting" viewBox="0 0 24 24" fill="currentColor">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            <span v-else class="loading-spinner small"></span>
            {{ runSubmitting ? '启动中...' : '确认运行' }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- 版本选择弹窗（参考 Rancher/KubeSphere 设计） -->
  <div v-if="showVersionDialog" class="modal-overlay" @click.self="showVersionDialog = false">
    <div class="modal-container version-dialog">
      <div class="modal-header">
        <h3>选择回滚版本</h3>
        <button class="close-btn" @click="showVersionDialog = false">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="modal-body">
        <div v-if="versionLoading" class="version-loading">
          <div class="loading-spinner"></div>
          <span>正在加载历史版本...</span>
        </div>
        <div v-else-if="versionHistory.length === 0" class="version-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <span>没有可回滚的历史版本</span>
        </div>
        <div v-else class="version-list">
          <div
            v-for="version in versionHistory"
            :key="version.rs_name"
            class="version-item"
            :class="{ selected: selectedVersion?.rs_name === version.rs_name, current: version.is_current }"
            @click="selectedVersion = version"
          >
            <div class="version-info">
              <span class="version-revision">Revision {{ version.revision }}</span>
              <span v-if="version.is_current" class="current-badge">当前版本</span>
            </div>
            <div class="version-details">
              <span class="version-rs">{{ version.rs_name }}</span>
              <span class="version-image" :title="version.image">{{ version.image }}</span>
            </div>
            <div class="version-time">
              {{ formatFullDate(version.created_at) }}
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="showVersionDialog = false">取消</button>
        <button
          class="btn btn-warning"
          @click="confirmRollbackToVersion"
          :disabled="!selectedVersion || selectedVersion.is_current || rollingBack"
        >
          <svg v-if="!rollingBack" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="1 4 1 10 7 10"/>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
          </svg>
          <span v-else class="loading-spinner small"></span>
          {{ rollingBack ? '回滚中...' : '确认回滚' }}
        </button>
      </div>
    </div>
  </div>

  <!-- 回滚结果详情弹窗 -->
  <div v-if="showRollbackResultDialog" class="modal-overlay" @click.self="!rollbackInProgress && (showRollbackResultDialog = false)">
    <div class="modal-container rollback-result-dialog">
      <div class="modal-header" :class="rollbackInProgress ? 'in-progress' : (rollbackResultData?.success ? 'success' : 'error')">
        <h3>
          <div v-if="rollbackInProgress" class="progress-spinner small"></div>
          <svg v-else-if="rollbackResultData?.success" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          {{ rollbackInProgress ? '回滚进行中...' : (rollbackResultData?.success ? '回滚成功' : '回滚失败') }}
        </h3>
        <button v-if="!rollbackInProgress" class="close-btn" @click="showRollbackResultDialog = false">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="modal-body">
        <div class="rollback-details">
          <div class="detail-item">
            <span class="label">目标版本:</span>
            <span class="value">{{ rollbackResultData?.target_rs || '-' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">命名空间:</span>
            <span class="value">{{ rollbackResultData?.namespace || '-' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">工作负载:</span>
            <span class="value">{{ rollbackResultData?.workload_name || '-' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">回滚前镜像:</span>
            <span class="value image">{{ rollbackResultData?.old_image || '-' }}</span>
          </div>
          <div v-if="rollbackResultData?.new_image" class="detail-item">
            <span class="label">回滚后镜像:</span>
            <span class="value image">{{ rollbackResultData?.new_image || '-' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">回滚时间:</span>
            <span class="value">{{ rollbackResultData?.rollback_at || '-' }}</span>
          </div>
        </div>
        <!-- 回滚日志（类似部署日志面板风格） -->
        <div class="rollback-log-panel">
          <div class="log-header">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
            </svg>
            <span>回滚日志</span>
            <div v-if="rollbackInProgress" class="log-loading">
              <div class="mini-spinner"></div>
              <span>执行中...</span>
            </div>
          </div>
          <pre class="log-content">{{ rollbackResultData?.message || '等待执行...' }}</pre>
        </div>
      </div>
      <div class="modal-footer">
        <button v-if="!rollbackInProgress" class="btn btn-primary" @click="showRollbackResultDialog = false">确定</button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import {
  getPipelineDetail,
  runPipeline,
  stopPipeline,
  getPipelineLogs,
  getPipelineHistory,
  getPipelineStages,
  getPipelineStatus,
  getRunStages,
  getStageLogs,
  approveStage,
  executeDeployStage,
  cancelDeployStage,
  rollbackDeployStage,
  getDeployHistory
} from '@/api/platform/pipeline'
import deploymentsApi from '@/api/cluster/workloads/deployments'

export default {
  name: 'PipelineDetail',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const pipelineId = computed(() => route.params.id)

    const pipeline = ref({})
    const loading = ref(true)
    const errorMsg = ref('')
    const activeTab = ref('overview')
    const latestRun = ref(null) // 最新运行记录（包含错误信息）

    // 日志相关
    const logs = ref('')
    const logsLoading = ref(false)
    const logsError = ref('')  // 日志加载错误信息
    const autoScroll = ref(true)
    const logsContainer = ref(null)
    const logLineCount = ref(0)  // 已加载的行数（用于增量获取）
    const isFirstLoad = ref(true)  // 是否首次加载

    // 历史相关
    const history = ref([])
    const historyLoading = ref(false)
    const historyFilter = ref('')
    const expandedRunId = ref(null) // 展开的运行记录ID

    // 切换运行记录详情展开/收起
    const toggleRunDetail = (run) => {
      if (expandedRunId.value === run.id) {
        expandedRunId.value = null
      } else {
        expandedRunId.value = run.id
      }
    }

    // 筛选后的历史记录
    const filteredHistory = computed(() => {
      if (!historyFilter.value) return history.value
      return history.value.filter(run => run.status === historyFilter.value)
    })

    // 获取历史状态计数
    const getHistoryStatusCount = (status) => {
      return history.value.filter(run => run.status === status).length
    }

    // 阶段数据（从 Jenkins API 获取）
    // 注意：初始状态为模拟数据，实际数据从 API 获取
    const pipelineStages = ref([
      { name: '代码检出', status: 'success', duration: '3s', type: 'checkout', steps: [{ name: 'Git Clone', status: 'success' }] },
      { name: '构建', status: 'success', duration: '5s', type: 'build', steps: [{ name: 'Compile', status: 'success' }, { name: 'Package', status: 'success' }] },
      { name: '测试', status: 'success', duration: '2s', type: 'test', steps: [{ name: 'Unit Test', status: 'success' }] },
      { name: '推送镜像', status: 'success', duration: '8s', type: 'push', steps: [{ name: 'Push Image', status: 'success' }] },
      { name: '人工审批', status: 'pending', duration: '-', type: 'approval', steps: [] },
      { name: '部署', status: 'pending', duration: '-', type: 'deploy', steps: [] }
    ])
    const stagesLoading = ref(false)
    const expandedStages = ref([]) // 展开的阶段列表
    const selectedStage = ref(null) // 当前选中的阶段
    const stageFilter = ref('')

    // 运行弹窗相关
    const showRunDialog = ref(false)
    const runSubmitting = ref(false)

    // 审批和部署操作相关
    const approving = ref(false)
    const deploying = ref(false)
    const rollingBack = ref(false)  // 回滚中
    const cancelling = ref(false)   // 取消中
    const approvalDecision = ref('approve')  // 默认通过
    const approvalComment = ref('')  // 审批备注

    // 版本选择弹窗相关
    const showVersionDialog = ref(false)
    const versionHistory = ref([])  // 历史版本列表
    const versionLoading = ref(false)
    const selectedVersion = ref(null)  // 选中的版本
    const currentStageForRollback = ref(null)  // 当前要回滚的阶段

    // 回滚结果弹窗相关
    const showRollbackResultDialog = ref(false)
    const rollbackResultData = ref(null)  // 回滚结果详情
    const rollbackInProgress = ref(false)  // 回滚进行中标志

    // 阶段日志预览相关
    const showStageLogs = ref(true)  // 是否展开日志预览

    // 追加回滚日志
    const appendRollbackLog = (msg) => {
      const timestamp = new Date().toLocaleString('zh-CN')
      const logLine = `[${timestamp}] ${msg}\n`
      if (rollbackResultData.value) {
        rollbackResultData.value.message = (rollbackResultData.value.message || '') + logLine
      }
    }

    // 筛选后的阶段
    const filteredStages = computed(() => {
      if (!stageFilter.value) return pipelineStages.value
      return pipelineStages.value.filter(stage => stage.status === stageFilter.value)
    })

    // 失败的阶段（用于错误摘要展示，参考 Jenkins Blue Ocean）
    const failedStages = computed(() => {
      return pipelineStages.value.filter(stage => stage.status === 'failed')
    })

    // 是否有运行中的阶段（包括 running 和 deploying）
    const hasRunningStage = computed(() => {
      return pipelineStages.value.some(stage => stage.status === 'running' || stage.status === 'deploying')
    })

    // 当前运行中的阶段（包括 running 和 deploying）
    const currentRunningStage = computed(() => {
      return pipelineStages.value.find(stage => stage.status === 'running' || stage.status === 'deploying')
    })

    // 获取连接线状态（根据前一个阶段的状态决定连接线颜色）
    const getConnectorStatus = (stageIndex) => {
      const stage = pipelineStages.value[stageIndex]
      if (!stage) return 'pending'
      return stage.status
    }

    // 获取阶段状态计数
    const getStageStatusCount = (status) => {
      return pipelineStages.value.filter(stage => stage.status === status).length
    }

    // 轮询定时器
    let statusPollingTimer = null
    let logsPollingTimer = null

    // 加载流水线阶段数据
    const loadStages = async () => {
      stagesLoading.value = true
      try {
        let stages = null

        // 优先从数据库获取阶段数据（包含审批/部署阶段）
        if (latestRun.value && latestRun.value.id) {
          const response = await getRunStages(latestRun.value.id)
          if (response.code === 0 && response.data && response.data.stages) {
            stages = response.data.stages
            console.log('[loadStages] 从数据库获取阶段数据:', stages.map(s => ({ name: s.name, type: s.type, status: s.status })))
          }
        }

        // 回退到从 Jenkins 获取阶段数据
        if (!stages) {
          const response = await getPipelineStages(pipelineId.value)
          if (response.code === 0 && response.data && response.data.stages) {
            stages = response.data.stages
            console.log('[loadStages] 从 Jenkins 获取阶段数据:', stages.map(s => ({ name: s.name, type: s.type, status: s.status })))
          }
        }

        // 如果获取到阶段数据，根据流水线状态智能推断阶段状态
        if (stages && stages.length > 0) {
          pipelineStages.value = inferStageStatus(stages)
          console.log('[loadStages] 处理后的阶段数据:', pipelineStages.value.map(s => ({ name: s.name, type: s.type, status: s.status })))
        }
      } catch (error) {
        console.error('加载阶段数据失败:', error)
      } finally {
        stagesLoading.value = false
      }
    }

    // 智能推断阶段状态（参考 Rancher/KubeSphere 设计）
    // 当 API 返回的状态都是 pending 时，根据流水线整体状态推断
    const inferStageStatus = (stages) => {
      // 获取流水线状态（优先级：latest_run > pipeline）
      const runStatus = latestRun.value?.status || pipeline.value.last_run_status || pipeline.value.status
      const buildStageTypes = ['checkout', 'build', 'test', 'push']

      console.log('[inferStageStatus] runStatus:', runStatus, 'stages:', stages.length)

      return stages.map((stage, index) => {
        const stageType = stage.type || stage.stage_type || ''
        const isBuildStage = buildStageTypes.includes(stageType)
        const currentStatus = stage.status || 'pending'

        // 审批阶段：保持后端返回的实际状态，不覆盖
        // 如果后端已经返回 success/approved/failed/rejected，直接使用
        if (stageType === 'approval') {
          if (['success', 'approved', 'failed', 'rejected'].includes(currentStatus)) {
            return stage  // 保持后端返回的实际状态
          }
          // 只有当状态是 pending 且构建成功时，才推断为 waiting
          if (currentStatus === 'pending' && (runStatus === 'success' || runStatus === 'SUCCESS')) {
            return { ...stage, status: 'waiting' }
          }
          return stage
        }

        // 如果阶段已经有明确状态（非 pending），不覆盖
        if (currentStatus && currentStatus !== 'pending') {
          return stage
        }

        // 根据流水线状态推断构建阶段状态
        if (runStatus === 'success' || runStatus === 'SUCCESS') {
          // 构建成功：所有构建阶段都成功
          if (isBuildStage) {
            return { ...stage, status: 'success', duration: stage.duration || getDemoStageDuration(stageType) }
          }
          // 部署阶段
          if (stageType === 'deploy') {
            return { ...stage, status: 'pending' }
          }
        } else if (runStatus === 'failed' || runStatus === 'FAILURE') {
          // 构建失败：最后一个构建阶段失败
          if (isBuildStage) {
            const buildStages = stages.filter(s => buildStageTypes.includes(s.type || s.stage_type || ''))
            const currentBuildIndex = buildStages.findIndex(s => s.name === stage.name)
            if (currentBuildIndex < buildStages.length - 1) {
              return { ...stage, status: 'success', duration: stage.duration || getDemoStageDuration(stageType) }
            } else {
              return { ...stage, status: 'failed', duration: stage.duration || '-' }
            }
          }
        } else if (runStatus === 'running' || runStatus === 'IN_PROGRESS') {
          // 运行中：最后一个构建阶段运行中
          if (isBuildStage) {
            const buildStages = stages.filter(s => buildStageTypes.includes(s.type || s.stage_type || ''))
            const currentBuildIndex = buildStages.findIndex(s => s.name === stage.name)
            if (currentBuildIndex < buildStages.length - 1) {
              return { ...stage, status: 'success', duration: stage.duration || getDemoStageDuration(stageType) }
            } else {
              return { ...stage, status: 'running', duration: '-' }
            }
          }
        }

        return stage
      })
    }

    // 获取演示用的阶段耗时
    const getDemoStageDuration = (stageType) => {
      const durations = {
        'checkout': '3s',
        'build': '15s',
        'test': '8s',
        'push': '12s'
      }
      return durations[stageType] || '-'
    }

    // 加载流水线详情
    const loadPipeline = async () => {
      loading.value = true
      errorMsg.value = ''
      try {
        // 使用 status API 获取完整信息（包含 latest_run）
        const response = await getPipelineStatus(pipelineId.value)
        if (response.code === 0) {
          pipeline.value = response.data.pipeline || response.data
          // 获取最新运行记录（包含阶段信息）
          if (response.data.latest_run) {
            latestRun.value = response.data.latest_run
          }
          // 加载阶段数据 - 使用 await 确保数据加载完成后再渲染
          await loadStages()
          // 如果正在运行，开始轮询
          if (pipeline.value.status === 'running') {
            startPolling()
          }
        } else {
          throw new Error(response.msg || '获取详情失败')
        }
      } catch (error) {
        errorMsg.value = error.message
        pipeline.value = {}
      } finally {
        loading.value = false
      }
    }

    // 开始轮询状态和日志（实时更新，大厂风格）
    const startPolling = () => {
      // 清理旧的定时器
      stopPolling()

      // 每 1.5 秒轮询状态和阶段（更快响应，实时体验）
      statusPollingTimer = setInterval(async () => {
        try {
          const response = await getPipelineStatus(pipelineId.value)
          if (response.code === 0) {
            const newPipeline = response.data.pipeline || response.data
            pipeline.value = { ...pipeline.value, ...newPipeline }

            // 更新最新运行记录（包含错误信息）
            if (response.data.latest_run) {
              latestRun.value = response.data.latest_run
            }

            // 加载阶段数据 - 使用 await 确保状态即时更新
            if (newPipeline.last_build_number) {
              const previousRunningStage = pipelineStages.value.find(s => s.status === 'running' || s.status === 'deploying')
              await loadStages()
              
              // 实时跟踪当前执行的阶段（大厂风格）
              autoSelectRunningStage(previousRunningStage)
            }

            // 如果不再运行，停止轮询
            if (newPipeline.status !== 'running') {
              stopPolling()
              // 最后再获取一次完整日志
              if (activeTab.value === 'logs') {
                await loadLogs()
              }
              
              // 根据最终状态显示消息（优先判断状态，而不是错误信息）
              const status = newPipeline.last_run_status
              const statusText = runStatusText(status)
              
              if (status === 'success' || status === 'SUCCESS') {
                // 构建成功，显示绿色 ✓
                Message.success({ content: `构建成功！${statusText}` })
              } else if (status === 'failed' || status === 'FAILURE') {
                // 构建失败，显示红色 ✗ + 错误信息
                const errorMsg = response.data.latest_run?.error_message || '构建失败'
                Message.error({ content: errorMsg, duration: 5000 })
              } else {
                Message.info({ content: `流水线执行完成，状态: ${statusText}` })
              }
            }
          }
        } catch (error) {
          console.error('轮询状态失败:', error)
        }
      }, 1500)  // 1.5秒间隔，更快响应

      // 每 2 秒轮询日志（如果在日志 Tab 或有选中阶段）
      logsPollingTimer = setInterval(async () => {
        if (pipeline.value.status === 'running') {
          // 在日志 Tab 或选中了阶段时加载日志
          if (activeTab.value === 'logs' || (selectedStage.value && stageDetailExpanded.value)) {
            await loadLogs()
          }
        }
      }, 2000)  // 2秒间隔
    }

    // 自动选中当前执行的阶段（大厂风格实时跟踪）
    const autoSelectRunningStage = (previousRunningStage) => {
      // 找到当前运行中的阶段（包括 running 和 deploying）
      const currentRunning = pipelineStages.value.find(s => s.status === 'running' || s.status === 'deploying')
      
      // 如果有新的运行中阶段，自动选中并展开
      if (currentRunning) {
        // 检查是否切换到了新阶段
        const isNewStage = !previousRunningStage || previousRunningStage.name !== currentRunning.name
        
        if (isNewStage || !selectedStage.value) {
          // 自动选中当前运行的阶段
          selectedStage.value = currentRunning
          stageDetailExpanded.value = true
        } else if (selectedStage.value && selectedStage.value.name === currentRunning.name) {
          // 更新当前选中阶段的数据
          selectedStage.value = currentRunning
        }
      } else {
        // 没有运行中的阶段，检查是否有失败的阶段
        const failedStage = pipelineStages.value.find(s => s.status === 'failed')
        const wasRunning = selectedStage.value && (selectedStage.value.status === 'running' || selectedStage.value.status === 'deploying')
        if (failedStage && (!selectedStage.value || wasRunning)) {
          // 自动选中失败的阶段
          selectedStage.value = failedStage
          stageDetailExpanded.value = true
        }
      }
      
      // 同步更新选中阶段的数据
      if (selectedStage.value) {
        const updatedStage = pipelineStages.value.find(s => s.id === selectedStage.value.id || s.name === selectedStage.value.name)
        if (updatedStage) {
          selectedStage.value = updatedStage
        }
      }
    }

    // 停止轮询
    const stopPolling = () => {
      if (statusPollingTimer) {
        clearInterval(statusPollingTimer)
        statusPollingTimer = null
      }
      if (logsPollingTimer) {
        clearInterval(logsPollingTimer)
        logsPollingTimer = null
      }
    }

    // 加载日志（支持增量加载，避免闪烁）
    const loadLogs = async (forceRefresh = false) => {
      // 首次加载或强制刷新时显示 loading
      if (isFirstLoad.value || forceRefresh) {
        logsLoading.value = true
        logsError.value = ''
        logs.value = ''
        logLineCount.value = 0
      }

      try {
        // 增量获取：从上次加载的行数开始
        const startLine = forceRefresh ? 0 : logLineCount.value
        const response = await getPipelineLogs(pipelineId.value, null, startLine)

        if (response.code === 0) {
          const newLogs = response.data.logs || ''
          const totalLines = response.data.total_lines || 0

          if (newLogs) {
            if (isFirstLoad.value || forceRefresh) {
              // 首次加载：直接设置
              logs.value = newLogs
              isFirstLoad.value = false
            } else {
              // 增量加载：追加新内容
              logs.value += newLogs
            }

            // 更新已加载行数
            if (totalLines > 0) {
              logLineCount.value = totalLines
            } else {
              // 如果后端没返回总行数，自己计算
              logLineCount.value += newLogs.split('\n').filter(line => line).length
            }

            // 平滑滚动到底部
            if (autoScroll.value) {
              nextTick(() => {
                if (logsContainer.value) {
                  logsContainer.value.scrollTo({
                    top: logsContainer.value.scrollHeight,
                    behavior: 'smooth'
                  })
                }
              })
            }
          }
        } else {
          // 处理后端返回的错误
          logsError.value = response.msg || '加载日志失败'
          logs.value = ''
        }
      } catch (error) {
        console.error('加载日志失败:', error)
        logsError.value = error.message || '加载日志失败'
        logs.value = ''
      } finally {
        logsLoading.value = false
      }
    }

    // 刷新日志（强制重新加载）
    const refreshLogs = () => {
      isFirstLoad.value = true
      loadLogs(true)
    }

    // 加载历史
    const loadHistory = async () => {
      historyLoading.value = true
      try {
        const response = await getPipelineHistory(pipelineId.value)
        if (response.code === 0) {
          history.value = response.data.list || response.data || []
        }
      } catch (error) {
        console.error('加载历史失败:', error)
      } finally {
        historyLoading.value = false
      }
    }

    // 操作
    const handleRun = async () => {
      try {
        Message.info({ content: '正在获取最新配置...' })
        
        // 立即清理旧数据，让用户看到清理状态
        clearOldRunData()
        
        // 先刷新获取最新流水线配置
        await loadPipeline()
        
        Message.info({ content: '正在启动流水线...' })
        // 传入 force: true 自动清理旧的失败/运行中构建
        const response = await runPipeline(pipelineId.value, { force: true })
        if (response.code === 0) {
          Message.success({ content: '流水线启动成功，正在执行中...' })
          // 刷新状态并开始轮询
          await loadPipeline()
          await loadStages()
          // 自动切换到日志 Tab
          activeTab.value = 'logs'
          loadLogs(true)
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || '启动失败' })
      }
    }

    // 清理旧的运行数据（重新运行前调用）
    const clearOldRunData = () => {
      // 清理日志
      isFirstLoad.value = true
      logs.value = ''
      logLineCount.value = 0
      logsError.value = ''
      
      // 清理选中的阶段详情（因为旧数据将无效）
      selectedStage.value = null
      stageDetailExpanded.value = false
      
      // 清理最新运行记录的错误信息
      if (latestRun.value) {
        latestRun.value = {
          ...latestRun.value,
          error_message: '',
          error_msg: '',
          status: 'running'
        }
      }
      
      // 重置阶段状态为 pending
      pipelineStages.value = pipelineStages.value.map(stage => ({
        ...stage,
        status: 'pending',
        duration: '-',
        error_message: '',
        error_msg: '',
        started_at: null,
        finished_at: null
      }))
    }

    // 确认运行（从弹窗）
    const confirmRun = async () => {
      runSubmitting.value = true
      try {
        await handleRun()
        showRunDialog.value = false
      } finally {
        runSubmitting.value = false
      }
    }

    const handleStop = async () => {
      const isPending = pipeline.value.last_run_status === 'pending'
      const actionText = isPending ? '取消' : '停止'
      try {
        Message.info({ content: `正在${actionText}构建...` })
        const response = await stopPipeline(pipelineId.value)
        if (response.code === 0) {
          Message.success({ content: `构建已${actionText}` })
          stopPolling()
          loadPipeline()
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || `${actionText}失败` })
      }
    }

    const handleEdit = () => {
      router.push(`/cicd/pipelines/${pipelineId.value}/edit`)
    }

    const viewRunLogs = (run) => {
      // 查看历史日志时重新加载
      isFirstLoad.value = true
      activeTab.value = 'logs'
      loadLogs(true)
    }

    // 切换阶段展开/收起
    const toggleStageExpand = (stageName) => {
      const index = expandedStages.value.indexOf(stageName)
      if (index === -1) {
        expandedStages.value.push(stageName)
      } else {
        expandedStages.value.splice(index, 1)
      }
    }

    // 阶段详情展开状态
    const stageDetailExpanded = ref(false)

    // 选中阶段（点击阶段卡片时触发）
    const selectStage = (stage) => {
      // 如果点击的是已选中的阶段，切换展开/收起
      if (selectedStage.value && selectedStage.value.name === stage.name) {
        stageDetailExpanded.value = !stageDetailExpanded.value
      } else {
        selectedStage.value = stage
        stageDetailExpanded.value = true  // 选中新阶段时自动展开
      }
      // 确保选中阶段展开
      if (!expandedStages.value.includes(stage.name)) {
        expandedStages.value.push(stage.name)
      }
    }

    // 切换阶段详情展开/收起
    const toggleStageDetail = () => {
      stageDetailExpanded.value = !stageDetailExpanded.value
    }

    // 查看阶段日志
    const viewStageLog = (stage) => {
      activeTab.value = 'logs'
      loadLogs()
      Message.info({ content: `已跳转到构建日志，当前阶段: ${stage.name}` })
    }

    const retryRun = async (run) => {
      await handleRun()
    }

    // 审批阶段操作（旧版方法保留）
    const handleApproveStage = async (stageId, action) => {
      approving.value = true
      try {
        const actionText = action === 'approve' ? '通过' : '拒绝'
        Message.info({ content: `正在处理审批${actionText}...` })
        const response = await approveStage(stageId, action, '')
        if (response.code === 0) {
          Message.success({ content: `审批${actionText}成功` })
          // 刷新阶段数据
          await loadStages()
          await loadPipeline()
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || '审批操作失败' })
      } finally {
        approving.value = false
      }
    }

    // 提交审批（新版，支持备注）
    // 优化：审批通过后自动触发部署，无需用户手动点击
    const submitApproval = async (stageId) => {
      approving.value = true
      try {
        const action = approvalDecision.value
        const comment = approvalComment.value
        const actionText = action === 'approve' ? '通过' : '拒绝'

        Message.info({ content: `正在处理审批${actionText}...` })
        const response = await approveStage(stageId, action, comment)

        if (response.code === 0) {
          Message.success({ content: `审批${actionText}成功` })
          
          // 立即更新选中阶段的状态（让用户立刻看到变化）
          if (selectedStage.value && selectedStage.value.id === stageId) {
            selectedStage.value = {
              ...selectedStage.value,
              status: action === 'approve' ? 'success' : 'failed',
              approval_info: {
                approver_name: '当前用户',
                approved_at: Date.now(),
                comment: comment
              }
            }
          }
          
          // 重置表单
          approvalDecision.value = 'approve'
          approvalComment.value = ''
          
          // 刷新阶段数据（获取服务器最新状态）
          await loadStages()
          await loadPipeline()
          
          // 同步更新选中阶段为最新数据
          if (selectedStage.value) {
            const updatedStage = pipelineStages.value.find(s => s.id === stageId)
            if (updatedStage) {
              selectedStage.value = updatedStage
            }
          }

          // 审批通过后自动触发部署
          if (action === 'approve') {
            // 找到下一个待执行的部署阶段
            const deployStage = pipelineStages.value.find(
              s => s.type === 'deploy' && (s.status === 'pending' || s.can_operate)
            )
            if (deployStage) {
              Message.info({ content: '审批通过，正在自动启动部署...' })
              // 自动触发部署
              await handleDeployStage(deployStage.id)
            }
          }
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || '审批操作失败' })
      } finally {
        approving.value = false
      }
    }

    // 部署阶段操作
    // 优化：启动部署前先刷新获取最新配置
    const handleDeployStage = async (stageId) => {
      deploying.value = true
      try {
        Message.info({ content: '正在获取最新配置...' })
        // 先刷新流水线和阶段数据，获取最新配置
        await loadPipeline()
        await loadStages()
        
        // 检查选中阶段是否有配置警告
        const currentStage = pipelineStages.value.find(s => s.id === stageId)
        if (currentStage && currentStage.config_warning) {
          Message.warning({ content: currentStage.config_warning, duration: 5000 })
          deploying.value = false
          return
        }
        
        Message.info({ content: '正在启动部署...' })
        const response = await executeDeployStage(stageId)
        if (response.code === 0) {
          Message.success({ content: '部署已启动，正在监控部署状态...' })
          // 刷新阶段数据
          await loadStages()
          // 启动部署状态轮询
          startDeployPolling(stageId)
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || '启动部署失败' })
      } finally {
        deploying.value = false
      }
    }

    // 重新部署（失败后重试）
    // 优化：重试前先刷新获取最新配置
    const handleRetryDeploy = async (stageId) => {
      deploying.value = true
      try {
        Message.info({ content: '正在获取最新配置...' })
        // 先刷新流水线和阶段数据，获取最新配置
        await loadPipeline()
        await loadStages()
        
        // 检查选中阶段是否有配置警告
        const currentStage = pipelineStages.value.find(s => s.id === stageId)
        if (currentStage && currentStage.config_warning) {
          Message.warning({ content: currentStage.config_warning, duration: 5000 })
          deploying.value = false
          return
        }
        
        Message.info({ content: '正在重新部署...' })
        const response = await executeDeployStage(stageId, { retry: true })
        if (response.code === 0) {
          Message.success({ content: '重新部署已启动，正在监控部署状态...' })
          // 刷新阶段数据
          await loadStages()
          // 启动部署状态轮询
          startDeployPolling(stageId)
        } else {
          throw new Error(response.msg)
        }
      } catch (error) {
        Message.error({ content: error.message || '重新部署失败' })
      } finally {
        deploying.value = false
      }
    }

    // 部署状态轮询（3秒间隔，更快响应）
    let deployPollingTimer = null
    const startDeployPolling = (stageId) => {
      // 先停止之前的轮询
      stopDeployPolling()

      // 每 3 秒轮询部署状态
      deployPollingTimer = setInterval(async () => {
        try {
          await loadStages()

          // 检查部署阶段状态
          const deployStage = pipelineStages.value.find(s => s.id === stageId)
          if (deployStage) {
            console.log('[deployPolling] 部署状态:', deployStage.status)

            // 同步更新 selectedStage（解决上下状态不同步问题）
            if (selectedStage.value && selectedStage.value.id === stageId) {
              selectedStage.value = deployStage
            }

            if (deployStage.status === 'success') {
              stopDeployPolling()
              Message.success({ content: '部署成功！', duration: 5000 })
            } else if (deployStage.status === 'failed') {
              stopDeployPolling()
              const errorMsg = deployStage.error_message || '部署失败'
              Message.error({ content: errorMsg, duration: 5000 })
            }
            // running 状态继续轮询
          }
        } catch (error) {
          console.error('[deployPolling] 轮询出错:', error)
        }
      }, 3000)  // 3秒间隔，比构建轮询更频繁
    }

    const stopDeployPolling = () => {
      if (deployPollingTimer) {
        clearInterval(deployPollingTimer)
        deployPollingTimer = null
      }
    }

    // 回滚部署（参考 Rancher/KubeSphere 设计）
    // 点击后打开版本选择弹窗
    // 优化：操作前先刷新获取最新配置
    const handleRollback = async (stage) => {
      if (!stage.id) {
        Message.warning({ content: '缺少阶段信息' })
        return
      }

      // 先刷新获取最新配置
      Message.info({ content: '正在获取最新配置...' })
      await loadPipeline()
      await loadStages()

      currentStageForRollback.value = stage
      versionLoading.value = true
      showVersionDialog.value = true

      try {
        // 使用新的阶段历史 API
        const res = await getDeployHistory(stage.id)
        if (res.code === 0 && res.data && res.data.revisions) {
          versionHistory.value = res.data.revisions
        } else {
          throw new Error('获取历史版本失败')
        }
      } catch (error) {
        console.error('[handleRollback] 错误:', error)
        Message.error({ content: error.message || '获取历史版本失败' })
        showVersionDialog.value = false
      } finally {
        versionLoading.value = false
      }
    }

    // 确认回滚到指定版本
    const confirmRollbackToVersion = async () => {
      if (!selectedVersion.value || !currentStageForRollback.value) {
        Message.warning({ content: '请选择要回滚的版本' })
        return
      }

      const stage = currentStageForRollback.value
      const targetVersion = selectedVersion.value

      // 关闭版本选择弹窗
      showVersionDialog.value = false
      selectedVersion.value = null

      // 立即显示回滚日志弹窗（进行中状态）
      rollbackInProgress.value = true
      rollbackResultData.value = {
        success: null,  // null 表示进行中
        target_rs: targetVersion.rs_name,
        old_image: stage.deploy_info?.image || '',
        new_image: targetVersion.image || '',
        namespace: stage.deploy_info?.namespace || '',
        workload_name: stage.deploy_info?.workload_name || '',
        rollback_at: new Date().toLocaleString('zh-CN'),
        user_id: '',
        message: ''
      }
      showRollbackResultDialog.value = true

      // 实时追加日志
      appendRollbackLog('开始回滚操作...')
      appendRollbackLog(`目标命名空间: ${stage.deploy_info?.namespace || '-'}`)
      appendRollbackLog(`目标工作负载: ${stage.deploy_info?.workload_name || '-'}`)
      appendRollbackLog(`目标版本: ${targetVersion.rs_name}`)
      appendRollbackLog(`目标镜像: ${targetVersion.image || '-'}`)
      appendRollbackLog('[INFO] 正在校验目标版本...')

      rollingBack.value = true
      try {
        appendRollbackLog('[INFO] 正在更新 Deployment 配置...')
        
        const res = await rollbackDeployStage(
          stage.id,
          targetVersion.rs_name
        )

        // 更新回滚结果
        if (res.data) {
          rollbackResultData.value = {
            ...res.data,
            message: rollbackResultData.value.message + '\n' + (res.data.message || '')
          }
        }
        rollbackInProgress.value = false

        if (res.code === 0) {
          appendRollbackLog('[INFO] Deployment 更新已提交')
          appendRollbackLog(`[INFO] 回滚成功！`)
          appendRollbackLog(`回滚前镜像: ${res.data?.old_image || '-'}`)
          appendRollbackLog(`回滚后镜像: ${res.data?.new_image || '-'}`)
          Message.success({ content: '回滚成功！' })
          await loadStages()
        } else {
          appendRollbackLog(`[ERROR] 回滚失败: ${res.msg || '未知错误'}`)
          Message.error({ content: res.msg || '回滚失败' })
        }
      } catch (error) {
        console.error('[confirmRollbackToVersion] 错误:', error)
        appendRollbackLog(`[ERROR] 回滚异常: ${error.message || '未知错误'}`)
        rollbackInProgress.value = false
        rollbackResultData.value.success = false
        Message.error({ content: error.message || '回滚失败' })
      } finally {
        rollingBack.value = false
      }
    }

    // 快速回滚到上一版本（后端自动找到上一个版本）
    // 优化：操作前先刷新获取最新配置
    const quickRollbackToPrevious = async (stage) => {
      if (!stage.id) {
        Message.warning({ content: '缺少阶段信息' })
        return
      }

      // 确认回滚
      const confirmed = window.confirm(
        `确定要回滚到上一个版本吗？`
      )
      if (!confirmed) return

      // 先刷新获取最新配置
      Message.info({ content: '正在获取最新配置...' })
      await loadPipeline()
      await loadStages()

      // 立即显示回滚日志弹窗（进行中状态）
      rollbackInProgress.value = true
      rollbackResultData.value = {
        success: null,  // null 表示进行中
        target_rs: '',
        old_image: '',
        new_image: '',
        namespace: stage.deploy_info?.namespace || '',
        workload_name: stage.deploy_info?.workload_name || '',
        rollback_at: new Date().toLocaleString('zh-CN'),
        user_id: '',
        message: ''
      }
      showRollbackResultDialog.value = true

      // 实时追加日志
      appendRollbackLog('开始回滚操作...')
      appendRollbackLog(`目标命名空间: ${stage.deploy_info?.namespace || '-'}`)
      appendRollbackLog(`目标工作负载: ${stage.deploy_info?.workload_name || '-'}`)
      appendRollbackLog('正在查找上一个版本...')

      rollingBack.value = true
      try {
        // 传空字符串或 __previous__，后端自动找到上一个版本
        const res = await rollbackDeployStage(stage.id, '__previous__')

        // 更新回滚结果
        if (res.data) {
          rollbackResultData.value = {
            ...res.data,
            message: rollbackResultData.value.message + '\n' + (res.data.message || '')
          }
        }
        rollbackInProgress.value = false

        if (res.code === 0) {
          appendRollbackLog(`回滚成功！目标版本: ${res.data?.target_rs || '-'}`)
          appendRollbackLog(`回滚前镜像: ${res.data?.old_image || '-'}`)
          appendRollbackLog(`回滚后镜像: ${res.data?.new_image || '-'}`)
          Message.success({ content: '回滚成功！' })
          await loadStages()
        } else {
          appendRollbackLog(`回滚失败: ${res.msg || '未知错误'}`)
          Message.error({ content: res.msg || '回滚失败' })
        }
      } catch (error) {
        console.error('[quickRollbackToPrevious] 错误:', error)
        appendRollbackLog(`回滚异常: ${error.message || '未知错误'}`)
        rollbackInProgress.value = false
        rollbackResultData.value.success = false
        Message.error({ content: error.message || '回滚失败' })
      } finally {
        rollingBack.value = false
      }
    }

    // 取消部署（智能判断：未执行的取消，已执行的回滚）
    // 优化：操作前先刷新获取最新配置
    const handleCancelDeploy = async (stage) => {
      if (!stage.id) {
        Message.warning({ content: '缺少阶段信息' })
        return
      }

      // 确认取消
      const confirmed = window.confirm(
        `确定要取消部署吗？\n\n- 如果部署未执行，将直接取消\n- 如果部署已执行，将回滚到上一个版本`
      )
      if (!confirmed) return

      // 先刷新获取最新配置
      Message.info({ content: '正在获取最新配置...' })
      await loadPipeline()
      await loadStages()

      cancelling.value = true
      try {
        Message.info({ content: '正在取消部署...' })

        // 使用新的取消 API
        const res = await cancelDeployStage(stage.id)

        if (res.code === 0) {
          if (res.data?.action === 'rollback') {
            Message.success({ content: `已回滚到 ${res.data.target_rs}` })
          } else {
            Message.success({ content: '部署已取消' })
          }
          // 停止轮询
          stopDeployPolling()
          // 刷新阶段数据
          await loadStages()
        } else {
          throw new Error(res.msg || '取消失败')
        }
      } catch (error) {
        console.error('[handleCancelDeploy] 错误:', error)
        Message.error({ content: error.message || '取消部署失败' })
      } finally {
        cancelling.value = false
      }
    }

    const copyLogs = () => {
      if (logs.value) {
        navigator.clipboard.writeText(logs.value)
        Message.success({ content: '日志已复制' })
      }
    }

    const downloadLogs = () => {
      if (logs.value) {
        const blob = new Blob([logs.value], { type: 'text/plain' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `pipeline-${pipelineId.value}-logs.txt`
        a.click()
        URL.revokeObjectURL(url)
      }
    }

    // 格式化
    const statusText = (status) => {
      const map = { idle: '空闲', running: '运行中', disabled: '已禁用', error: '错误' }
      return map[status] || status
    }

    const runStatusText = (status) => {
      const map = { success: '成功', failed: '失败', running: '运行中', pending: '等待中', aborted: '已中止', '': '未运行' }
      return map[status] || status
    }

    const deployStatusText = (status) => {
      const map = {
        success: '部署成功',
        failed: '部署失败',
        pending: '等待部署',
        deploying: '部署中',
        approval_pending: '待审批',
        '': '未部署'
      }
      return map[status] || status
    }

    const envLabel = (env) => {
      const map = { dev: '开发环境', staging: '预发环境', prod: '生产环境' }
      return map[env] || env
    }

    const stageStatusText = (status) => {
      const map = {
        success: '完成',
        failed: '失败',
        running: '执行中',
        deploying: '部署中',  // 部署阶段特有状态
        pending: '等待',
        waiting: '待通过',
        skipped: '已跳过',
        aborted: '已中止',
        approved: '已通过',
        rejected: '已拒绝'
      }
      return map[status] || status
    }

    // 判断阶段是否正在运行（包含 running 和 deploying）
    const isStageRunning = (status) => {
      return status === 'running' || status === 'deploying'
    }

    // 审批阶段标签文本
    const approvalBadgeText = (status) => {
      const map = {
        waiting: '待通过',
        pending: '待通过',
        success: '已通过',
        approved: '已通过',
        failed: '已拒绝',
        rejected: '已拒绝'
      }
      return map[status] || '审批'
    }

    // 阶段类型名称
    const getStageTypeName = (type) => {
      const map = {
        checkout: '代码检出',
        build: '构建',
        test: '测试',
        push: '推送镜像',
        approval: '人工审批',
        deploy: '部署'
      }
      return map[type] || type
    }

    const formatDate = (timestamp) => {
      if (!timestamp) return '-'
      const date = new Date(typeof timestamp === 'number' && timestamp < 10000000000 ? timestamp * 1000 : timestamp)
      const now = new Date()
      const diff = now - date
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
      return date.toLocaleDateString('zh-CN')
    }

    const formatFullDate = (timestamp) => {
      if (!timestamp) return '-'
      const date = new Date(typeof timestamp === 'number' && timestamp < 10000000000 ? timestamp * 1000 : timestamp)
      return date.toLocaleString('zh-CN')
    }

    const formatDuration = (seconds) => {
      if (!seconds) return '-'
      if (seconds < 60) return `${seconds}秒`
      if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
      return `${Math.floor(seconds / 3600)}时${Math.floor((seconds % 3600) / 60)}分`
    }

    // 计算阶段耗时（实时计算）
    const calculateStageDuration = (stage) => {
      if (!stage) return '-'
      if (stage.duration) return stage.duration
      // 支持多种字段名
      const startTime = stage.started_at || stage.start_time || stage.startTime
      const endTime = stage.finished_at || stage.end_time || stage.endTime || stage.finish_time
      if (startTime) {
        const start = typeof startTime === 'number' && startTime < 10000000000 ? startTime * 1000 : startTime
        const end = endTime 
          ? (typeof endTime === 'number' && endTime < 10000000000 ? endTime * 1000 : endTime)
          : Date.now()
        const durationSec = Math.floor((end - start) / 1000)
        if (durationSec < 60) return `${durationSec}秒`
        if (durationSec < 3600) return `${Math.floor(durationSec / 60)}分${durationSec % 60}秒`
        return `${Math.floor(durationSec / 3600)}时${Math.floor((durationSec % 3600) / 60)}分`
      }
      return '-'
    }

    // 获取阶段开始时间（支持多种字段名）
    const getStageStartTime = (stage) => {
      if (!stage) return '-'
      const startTime = stage.started_at || stage.start_time || stage.startTime || stage.created_at
      if (startTime) {
        return formatFullDate(startTime)
      }
      // 如果正在运行，显示当前时间作为开始时间
      if (stage.status === 'running' || stage.status === 'deploying') {
        return '刚刚开始'
      }
      return '-'
    }

    // 获取阶段结束时间（支持多种字段名）
    const getStageEndTime = (stage) => {
      if (!stage) return '-'
      const endTime = stage.finished_at || stage.end_time || stage.endTime || stage.finish_time || stage.updated_at
      if (endTime) {
        return formatFullDate(endTime)
      }
      // 如果正在运行，显示运行中
      if (stage.status === 'running' || stage.status === 'deploying') {
        return '运行中...'
      }
      return '-'
    }

    // 切换阶段日志预览展开/折叠
    const toggleStageLogs = () => {
      showStageLogs.value = !showStageLogs.value
    }

    // 获取阶段日志预览（最后20行）
    const getStageLogsPreview = () => {
      if (!logs.value) return ''
      const lines = logs.value.split('\n')
      const previewLines = lines.slice(-20)  // 显示最后20行
      return previewLines.join('\n')
    }

    // 复制文本到剪贴板
    const copyText = (text) => {
      if (text) {
        navigator.clipboard.writeText(text)
        Message.success({ content: '已复制到剪贴板' })
      }
    }

    // URL 参数处理
    watch(() => route.query.tab, (tab) => {
      if (tab) activeTab.value = tab
    }, { immediate: true })

    onMounted(() => {
      loadPipeline()
    })

    // 清理定时器
    onBeforeUnmount(() => {
      stopPolling()
      stopDeployPolling()
    })

    return {
      pipeline,
      loading,
      errorMsg,
      activeTab,
      latestRun, // 最新运行记录（包含错误信息）
      logs,
      logsLoading,
      logsError,
      autoScroll,
      logsContainer,
      history,
      historyLoading,
      historyFilter,
      filteredHistory,
      expandedRunId,
      toggleRunDetail,
      getHistoryStatusCount,
      pipelineStages,
      stagesLoading,
      expandedStages,
      selectedStage,
      stageDetailExpanded,
      stageFilter,
      filteredStages,
      failedStages,
      hasRunningStage,
      currentRunningStage,
      getConnectorStatus,
      getStageStatusCount,
      loadPipeline,
      loadLogs,
      refreshLogs,
      loadHistory,
      loadStages,
      handleRun,
      confirmRun,
      showRunDialog,
      runSubmitting,
      handleStop,
      handleEdit,
      viewRunLogs,
      toggleStageExpand,
      selectStage,
      toggleStageDetail,
      viewStageLog,
      retryRun,
      handleApproveStage,
      handleDeployStage,
      handleRetryDeploy,
      handleRollback,
      handleCancelDeploy,
      confirmRollbackToVersion,
      quickRollbackToPrevious,
      showVersionDialog,
      versionHistory,
      versionLoading,
      selectedVersion,
      showRollbackResultDialog,
      rollbackResultData,
      rollbackInProgress,
      submitApproval,
      approving,
      deploying,
      rollingBack,
      cancelling,
      approvalDecision,
      approvalComment,
      copyLogs,
      downloadLogs,
      statusText,
      runStatusText,
      deployStatusText,
      envLabel,
      stageStatusText,
      approvalBadgeText,
      getStageTypeName,
      formatDate,
      formatFullDate,
      formatDuration,
      calculateStageDuration,
      getStageStartTime,
      getStageEndTime,
      toggleStageLogs,
      getStageLogsPreview,
      showStageLogs,
      copyText
    }
  }
}
</script>

<style scoped>
.pipeline-detail-view {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  min-height: 100vh;
  background: #f5f7fa;
}

/* 面包屑 */
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  font-size: 14px;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #4299e1;
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-link svg {
  width: 16px;
  height: 16px;
}

.separator {
  color: #cbd5e0;
}

.current {
  color: #4a5568;
  font-weight: 500;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
  color: #718096;
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

.loading-spinner.small {
  width: 24px;
  height: 24px;
  border-width: 2px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 头部 */
.pipeline-header {
  background: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.status-indicator.status-idle { background: #3182ce; }
.status-indicator.status-running { background: #d97706; animation: pulse 1.5s infinite; }
.status-indicator.status-disabled { background: #a0aec0; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.pipeline-title {
  font-size: 24px;
  font-weight: 700;
  color: #1a202c;
  margin: 0;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.status-idle { background: #ebf8ff; color: #3182ce; }
.status-badge.status-running { background: #fef3c7; color: #d97706; }
.status-badge.status-disabled { background: #f1f5f9; color: #64748b; }

.pipeline-desc {
  color: #718096;
  margin: 0 0 16px 0;
  font-size: 14px;
}

.pipeline-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #64748b;
}

.meta-item svg {
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

/* 上次运行信息行 */
.last-run-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.last-run-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.run-label {
  font-size: 13px;
  color: #64748b;
}

.run-status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.run-status-badge.status-success {
  background: #d1fae5;
  color: #059669;
}

.run-status-badge.status-failed {
  background: #fee2e2;
  color: #dc2626;
}

.run-status-badge.status-running {
  background: #dbeafe;
  color: #2563eb;
}

.run-status-badge.status-pending {
  background: #fef3c7;
  color: #d97706;
}

.run-status-badge.status-cancelled {
  background: #f1f5f9;
  color: #64748b;
}

.run-status-badge.status-aborted {
  background: #f1f5f9;
  color: #64748b;
}

.run-time {
  font-size: 13px;
  color: #94a3b8;
}

.last-run-actions {
  display: flex;
  gap: 8px;
}

.mini-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.mini-btn svg {
  width: 14px;
  height: 14px;
}

.mini-btn.btn-stop {
  background: #fee2e2;
  color: #dc2626;
}

.mini-btn.btn-stop:hover {
  background: #fecaca;
}

.mini-btn.btn-rerun {
  background: #dbeafe;
  color: #2563eb;
}

.mini-btn.btn-rerun:hover {
  background: #bfdbfe;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* Tab 导航 */
.tab-nav {
  display: flex;
  gap: 4px;
  background: white;
  padding: 8px;
  border-radius: 12px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #64748b;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  background: #f1f5f9;
}

.tab-btn.active {
  background: #4299e1;
  color: white;
}

.tab-btn svg {
  width: 18px;
  height: 18px;
}

/* Tab 内容 */
.tab-content {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* 概览 */
.section {
  margin-bottom: 32px;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  margin: 0 0 16px 0;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.status-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 12px;
}

.card-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-icon svg {
  width: 24px;
  height: 24px;
}

.card-icon.status-success { background: #d1fae5; color: #059669; }
.card-icon.status-failed { background: #fee2e2; color: #dc2626; }
.card-icon.status-running { background: #fef3c7; color: #d97706; }
.card-icon.neutral { background: #e2e8f0; color: #64748b; }

.card-content {
  display: flex;
  flex-direction: column;
}

.card-label {
  font-size: 13px;
  color: #94a3b8;
}

.card-value {
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-action-btn:hover:not(:disabled) {
  border-color: #4299e1;
  background: #ebf8ff;
}

.quick-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.quick-action-btn svg {
  width: 32px;
  height: 32px;
  color: #4299e1;
}

.quick-action-btn span {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

/* 阶段 */
/* 阶段筛选工具栏 */
.stages-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

/* 阶段加载和空状态 */
.stages-loading, .stages-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: #64748b;
  background: #f8fafc;
  border-radius: 12px;
  margin-bottom: 20px;
}

.stages-loading p, .stages-empty p {
  margin: 16px 0 0 0;
  font-size: 14px;
}

.stages-empty svg {
  width: 48px;
  height: 48px;
  color: #94a3b8;
}

/* ==================== Jenkins Blue Ocean 风格阶段视图 ==================== */
.stages-pipeline {
  display: flex;
  align-items: stretch;
  justify-content: flex-start;
  padding: 24px;
  margin-bottom: 32px;
  overflow-x: auto;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  gap: 0;
}

/* 新版流水线容器 */
.stages-pipeline-container {
  background: white;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  padding: 32px 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.stages-pipeline-track {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  overflow-x: auto;
  padding: 8px 0;
  gap: 0;
}

/* 阶段节点 */
.pipeline-stage {
  display: flex;
  align-items: center;
  position: relative;
  flex-shrink: 0;
}

.pipeline-stage.is-first {
  padding-left: 0;
}

/* 连接线 */
.stage-connector-line {
  width: 40px;
  height: 4px;
  flex-shrink: 0;
  position: relative;
  background: #e2e8f0;
  transition: all 0.3s ease;
}

.stage-connector-line.success {
  background: linear-gradient(90deg, #22c55e 0%, #22c55e 100%);
}

.stage-connector-line.failed {
  background: linear-gradient(90deg, #22c55e 0%, #ef4444 100%);
}

.stage-connector-line.running {
  background: linear-gradient(90deg, #22c55e 0%, #3b82f6 50%);
  position: relative;
  overflow: hidden;
}

.stage-connector-line.running::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.6) 50%, transparent 100%);
  animation: connectorShine 1.5s ease-in-out infinite;
}

@keyframes connectorShine {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* 阶段卡片 */
.stage-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 24px;
  min-width: 130px;
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}

.stage-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.pipeline-stage.selected .stage-card {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

/* 成功状态 */
.pipeline-stage.status-success .stage-card {
  background: linear-gradient(145deg, #f0fdf4 0%, #dcfce7 100%);
  border-color: #22c55e;
}

/* 失败状态 */
.pipeline-stage.status-failed .stage-card {
  background: linear-gradient(145deg, #fef2f2 0%, #fecaca 100%);
  border-color: #ef4444;
}

/* 运行中状态 */
.pipeline-stage.status-running .stage-card,
.pipeline-stage.status-deploying .stage-card {
  background: linear-gradient(145deg, #eff6ff 0%, #dbeafe 100%);
  border-color: #3b82f6;
  animation: stagePulse 2s ease-in-out infinite;
}

@keyframes stagePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.4); }
  50% { box-shadow: 0 0 0 8px rgba(59, 130, 246, 0); }
}

/* 等待状态 */
.pipeline-stage.status-waiting .stage-card {
  background: linear-gradient(145deg, #fefce8 0%, #fef3c7 100%);
  border-color: #f59e0b;
}

/* 待执行状态 */
.pipeline-stage.status-pending .stage-card {
  background: #f8fafc;
  border-color: #e2e8f0;
  opacity: 0.7;
}

/* 状态图标 */
.stage-status-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  transition: all 0.3s ease;
}

.stage-status-icon svg {
  width: 24px;
  height: 24px;
}

.stage-status-icon.status-success {
  background: #22c55e;
}

.stage-status-icon.status-success svg {
  stroke: white;
}

.stage-status-icon.status-failed {
  background: #ef4444;
}

.stage-status-icon.status-failed svg {
  stroke: white;
}

.stage-status-icon.status-running,
.stage-status-icon.status-deploying {
  background: #3b82f6;
}

.stage-status-icon.status-waiting {
  background: #f59e0b;
}

.stage-status-icon.status-waiting svg {
  stroke: white;
}

.stage-status-icon.status-pending {
  background: #e2e8f0;
}

/* 运行中旋转动画 */
.running-spinner {
  width: 26px;
  height: 26px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

/* 待执行圆点 */
.pending-dot {
  width: 12px;
  height: 12px;
  background: #94a3b8;
  border-radius: 50%;
}

/* 阶段名称 */
.stage-label {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  text-align: center;
  white-space: nowrap;
}

.pipeline-stage.status-pending .stage-label {
  color: #94a3b8;
}

/* 耗时标签 */
.stage-duration-badge {
  margin-top: 8px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(0, 0, 0, 0.05);
  color: #64748b;
}

.stage-duration-badge.status-success {
  background: rgba(34, 197, 94, 0.15);
  color: #166534;
}

.stage-duration-badge.status-failed {
  background: rgba(239, 68, 68, 0.15);
  color: #dc2626;
}

.stage-duration-badge.status-running,
.stage-duration-badge.status-deploying {
  background: rgba(59, 130, 246, 0.15);
  color: #1d4ed8;
}

.duration-running {
  display: flex;
  align-items: center;
  gap: 6px;
}

.running-dot {
  width: 6px;
  height: 6px;
  background: #3b82f6;
  border-radius: 50%;
  animation: blink 1s ease-in-out infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* 执行进度指示器 */
.pipeline-progress-indicator {
  margin-top: 20px;
  padding: 14px 20px;
  background: linear-gradient(90deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #93c5fd;
  border-radius: 10px;
}

.progress-text {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #1d4ed8;
}

.progress-icon {
  width: 10px;
  height: 10px;
  background: #3b82f6;
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

/* 阶段节点容器 */
.stage-node {
  display: flex;
  align-items: stretch;
  position: relative;
  flex: 1;
  min-width: 140px;
  max-width: 200px;
}

/* 连接线 - Jenkins 风格 */
.stage-connector {
  width: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.stage-connector::before {
  content: '';
  width: 100%;
  height: 3px;
  background: #e2e8f0;
  position: absolute;
}

/* 连接线状态颜色 */
.stage-node.status-success .stage-connector::before {
  background: #10b981;
}

.stage-node.status-running .stage-connector::before {
  background: linear-gradient(90deg, #10b981 0%, #3b82f6 50%, #e2e8f0 100%);
  animation: connectorFlow 1.5s ease-in-out infinite;
}

.stage-node.status-failed .stage-connector::before {
  background: linear-gradient(90deg, #10b981 0%, #ef4444 100%);
}

@keyframes connectorFlow {
  0% { background-position: 0% 50%; }
  100% { background-position: 100% 50%; }
}

/* 阶段内容 - Jenkins 方块卡片风格 */
.stage-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px 16px;
  min-width: 120px;
  background: white;
  border-radius: 8px;
  border: 2px solid #e2e8f0;
  transition: background 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease;
  position: relative;
}

/* 成功状态 - 绿色背景（参考 Jenkins） */
.stage-node.status-success .stage-content {
  background: linear-gradient(180deg, #dcfce7 0%, #bbf7d0 100%);
  border-color: #22c55e;
  box-shadow: 0 2px 8px rgba(34, 197, 94, 0.2);
}

/* 失败状态 - 红色背景 */
.stage-node.status-failed .stage-content {
  background: linear-gradient(180deg, #fee2e2 0%, #fecaca 100%);
  border-color: #ef4444;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.2);
}

/* 运行中状态 - 蓝色脉冲 */
.stage-node.status-running .stage-content {
  background: linear-gradient(180deg, #dbeafe 0%, #bfdbfe 100%);
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2);
  animation: runningPulse 2s ease-in-out infinite;
}

/* 等待状态 */
.stage-node.status-pending .stage-content {
  background: #f8fafc;
  border-color: #e2e8f0;
}

/* 等待审批状态 - 橙色闪烁 */
.stage-node.status-waiting .stage-content {
  background: linear-gradient(180deg, #fef3c7 0%, #fde68a 100%);
  border-color: #f59e0b;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.2);
  animation: waitingPulse 2s ease-in-out infinite;
}

@keyframes waitingPulse {
  0%, 100% { box-shadow: 0 2px 8px rgba(245, 158, 11, 0.2); }
  50% { box-shadow: 0 4px 16px rgba(245, 158, 11, 0.4); }
}

@keyframes runningPulse {
  0%, 100% { box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2); }
  50% { box-shadow: 0 4px 16px rgba(59, 130, 246, 0.4); }
}

/* 阶段图标 - 更紧凑 */
.stage-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 2px solid #e2e8f0;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.stage-icon svg {
  width: 20px;
  height: 20px;
  color: #94a3b8;
}

/* 成功图标 */
.stage-node.status-success .stage-icon {
  background: #22c55e;
  border-color: #16a34a;
}

.stage-node.status-success .stage-icon svg {
  color: white;
}

/* 失败图标 */
.stage-node.status-failed .stage-icon {
  background: #ef4444;
  border-color: #dc2626;
}

.stage-node.status-failed .stage-icon svg {
  color: white;
}

/* 运行中图标 */
.stage-node.status-running .stage-icon {
  background: #3b82f6;
  border-color: #2563eb;
}

.stage-node.status-running .stage-icon svg {
  color: white;
}

/* 等待图标 */
.stage-node.status-pending .stage-icon {
  background: #f1f5f9;
  border-color: #e2e8f0;
}

/* 等待审批图标 - 橙色 */
.stage-node.status-waiting .stage-icon {
  background: #f59e0b;
  border-color: #d97706;
}

.stage-node.status-waiting .stage-icon svg {
  color: white;
}

/* 等待审批文字颜色 */
.stage-node.status-waiting .stage-name { color: #92400e; }
.stage-node.status-waiting .stage-duration { color: #b45309; }

/* 旋转动画 */
.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 阶段信息 */
.stage-info {
  text-align: center;
}

.stage-name {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 2px;
}

.stage-duration {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

/* 状态对应的文字颜色 */
.stage-node.status-success .stage-name { color: #166534; }
.stage-node.status-success .stage-duration { color: #15803d; }
.stage-node.status-failed .stage-name { color: #991b1b; }
.stage-node.status-failed .stage-duration { color: #b91c1c; }
.stage-node.status-running .stage-name { color: #1e40af; }
.stage-node.status-running .stage-duration { color: #2563eb; }

/* 阶段卡片选中状态 */
.stage-node {
  cursor: pointer;
}

.stage-node.selected .stage-content {
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.4), 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.stage-node:hover .stage-content {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stage-node.selected:hover .stage-content {
  transform: translateY(-2px);
}

/* 选中阶段详情区域 */
.selected-stage-detail {
  margin-top: 24px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.selected-stage-detail .detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 1px solid #e2e8f0;
}

.selected-stage-detail .stage-title {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
}

.selected-stage-detail .stage-status {
  margin-left: auto;
  font-size: 13px;
  color: #64748b;
}

.selected-stage-detail .stage-duration-tag {
  padding: 4px 10px;
  background: #f1f5f9;
  border-radius: 4px;
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.selected-stage-detail .detail-body {
  padding: 20px;
  overflow: hidden;
}

/* 展开/收起动画 */
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.2s ease-in;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
  max-height: 0;
}

.slide-fade-enter-to,
.slide-fade-leave-from {
  opacity: 1;
  transform: translateY(0);
}

/* 可点击的头部 */
.selected-stage-detail .detail-header.clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.selected-stage-detail .detail-header.clickable:hover {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
}

/* 展开/收起指示器 */
.expand-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  padding: 4px 12px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 16px;
  transition: all 0.2s ease;
}

.expand-indicator:hover {
  background: rgba(59, 130, 246, 0.15);
}

.expand-hint {
  font-size: 12px;
  color: #3b82f6;
  font-weight: 500;
}

.expand-icon {
  width: 16px;
  height: 16px;
  color: #3b82f6;
  transition: transform 0.3s ease;
}

.expand-icon.rotated {
  transform: rotate(180deg);
}

/* 选中阶段详情卡片状态 */
.selected-stage-detail.expanded {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

/* 未选中阶段提示 */
.no-stage-selected {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  margin-top: 24px;
  background: #f8fafc;
  border: 2px dashed #e2e8f0;
  border-radius: 12px;
  color: #94a3b8;
}

.no-stage-selected svg {
  width: 48px;
  height: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.no-stage-selected p {
  font-size: 14px;
}

/* 阶段基本信息展示 */
.stage-basic-info {
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
}

.stage-basic-info .info-row {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #e2e8f0;
}

.stage-basic-info .info-row:last-child {
  border-bottom: none;
}

.stage-basic-info .info-label {
  width: 100px;
  font-size: 13px;
  color: #64748b;
  flex-shrink: 0;
}

.stage-basic-info .info-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
}

.stage-basic-info .info-value.status-text.status-success {
  color: #16a34a;
}

.stage-basic-info .info-value.status-text.status-failed {
  color: #dc2626;
}

.stage-basic-info .info-value.status-text.status-running {
  color: #2563eb;
}

.stage-basic-info .info-row.error {
  background: #fef2f2;
  margin: 8px -16px -16px;
  padding: 12px 16px;
  border-radius: 0 0 8px 8px;
}

.stage-basic-info .error-text {
  color: #dc2626;
  word-break: break-all;
}

.stage-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.stage-detail-card {
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  cursor: pointer;
  transition: background 0.2s;
}

.detail-header:hover {
  background: #f1f5f9;
}

.expand-icon {
  width: 16px;
  height: 16px;
  color: #94a3b8;
  transition: transform 0.2s;
  margin-left: auto;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.status-success { background: #059669; }
.status-dot.status-failed { background: #dc2626; }
.status-dot.status-running { background: #3b82f6; animation: pulse 1.5s infinite; }
.status-dot.status-waiting { background: #f59e0b; animation: pulse 1.5s infinite; }
.status-dot.status-pending { background: #94a3b8; }

.stage-title {
  flex: 1;
  font-weight: 600;
  color: #1a202c;
}

.stage-status {
  font-size: 12px;
  color: #64748b;
}

.detail-body {
  padding: 12px 16px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
}

.step-item:not(:last-child) {
  border-bottom: 1px dashed #e2e8f0;
}

.step-icon {
  width: 16px;
  height: 16px;
}

.step-icon.success { color: #059669; }
.step-icon.failed { color: #dc2626; }
.step-icon.pending { color: #94a3b8; }

.step-name {
  flex: 1;
  font-size: 13px;
  color: #4a5568;
}

.step-duration {
  font-size: 12px;
  color: #94a3b8;
}

.stage-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #e2e8f0;
}

.view-log-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  color: #2563eb;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.view-log-btn:hover {
  background: #dbeafe;
  border-color: #93c5fd;
}

.view-log-btn svg {
  width: 16px;
  height: 16px;
}

/* 日志 */
.logs-toolbar, .history-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar-left, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.log-label {
  font-size: 14px;
  color: #4a5568;
  font-weight: 500;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  color: #4a5568;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.toolbar-btn:hover:not(:disabled) {
  border-color: #cbd5e0;
  background: #f7fafc;
}

.toolbar-btn:disabled {
  opacity: 0.5;
}

.toolbar-btn svg {
  width: 16px;
  height: 16px;
}

.auto-scroll {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
}

.logs-container {
  background: #1e293b;
  border-radius: 12px;
  min-height: 400px;
  max-height: 600px;
  overflow: auto;
}

.logs-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 100px;
  color: #94a3b8;
}

.logs-content {
  padding: 20px;
  margin: 0;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #e2e8f0;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px;
  color: #64748b;
}

.logs-empty svg {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
  color: #475569;
}

/* 日志错误状态 */
.logs-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 40px;
  text-align: center;
}

.logs-error svg {
  width: 56px;
  height: 56px;
  margin-bottom: 16px;
  color: #f59e0b;
}

.logs-error .error-message {
  font-size: 15px;
  color: #64748b;
  margin: 0 0 24px 0;
  max-width: 400px;
  line-height: 1.6;
}

.logs-error .btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  font-size: 14px;
  font-weight: 600;
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.logs-error .btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
}

.logs-error .btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.logs-error .btn svg {
  width: 18px;
  height: 18px;
  margin: 0;
  color: white;
}

/* 历史筛选按钮 */
.history-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 16px;
}

.filter-tabs {
  display: flex;
  gap: 8px;
  flex: 1;
}

.filter-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover {
  border-color: #cbd5e0;
  background: #f7fafc;
}

.filter-tab.active {
  border-color: #4299e1;
  background: #ebf8ff;
  color: #2b6cb0;
}

.filter-tab .status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.filter-tab .status-dot.success { background: #10b981; }
.filter-tab .status-dot.failed { background: #ef4444; }
.filter-tab .status-dot.running { background: #f59e0b; animation: pulse 1.5s infinite; }

.filter-count {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  background: #f1f5f9;
  border-radius: 10px;
  color: #64748b;
}

.filter-tab.active .filter-count {
  background: #bee3f8;
  color: #2b6cb0;
}

/* 历史 */
.history-loading, .history-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: #64748b;
}

.history-empty svg {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
  color: #94a3b8;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  display: flex;
  flex-direction: column;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  transition: all 0.2s;
  overflow: hidden;
  position: relative;
}

.history-item:hover {
  border-color: #cbd5e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.history-item.expanded {
  border-color: #4299e1;
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.15);
}

.history-item-main {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 100px 16px 20px;
  cursor: pointer;
  transition: background 0.15s;
}

.history-item-main:hover {
  background: #f8fafc;
}

.history-expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  color: #94a3b8;
  transition: transform 0.2s;
}

.history-expand-icon svg {
  width: 18px;
  height: 18px;
  transition: transform 0.2s;
}

.history-expand-icon svg.rotated {
  transform: rotate(180deg);
}

.history-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.history-icon svg {
  width: 20px;
  height: 20px;
}

.history-item.status-success .history-icon { background: #d1fae5; color: #059669; }
.history-item.status-failed .history-icon { background: #fee2e2; color: #dc2626; }
.history-item.status-running .history-icon { background: #fef3c7; color: #d97706; }
.history-item.status-pending .history-icon { background: #f1f5f9; color: #64748b; }

.history-info {
  flex: 1;
}

.history-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.build-number {
  font-size: 15px;
  font-weight: 600;
  color: #1a202c;
}

.history-status {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.history-status.status-success { background: #d1fae5; color: #059669; }
.history-status.status-failed { background: #fee2e2; color: #dc2626; }
.history-status.status-running { background: #fef3c7; color: #d97706; }
.history-status.status-pending { background: #f1f5f9; color: #64748b; }

.history-meta {
  font-size: 13px;
  color: #64748b;
}

.history-actions {
  display: flex;
  gap: 8px;
  position: absolute;
  right: 20px;
  top: 16px;
}

.history-actions .action-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.history-actions .action-btn:hover {
  border-color: #4299e1;
  color: #4299e1;
}

.history-actions .action-btn.retry:hover {
  border-color: #d97706;
  color: #d97706;
}

.history-actions .action-btn svg {
  width: 16px;
  height: 16px;
}

/* 历史记录详情展开区域 */
.history-detail {
  padding: 0 20px 20px 76px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 16px 0;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
}

.detail-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
}

.detail-value.mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.detail-value.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  width: fit-content;
}

.detail-value.status-badge.status-success { background: #d1fae5; color: #059669; }
.detail-value.status-badge.status-failed { background: #fee2e2; color: #dc2626; }
.detail-value.status-badge.status-running { background: #fef3c7; color: #d97706; }
.detail-value.status-badge.status-pending { background: #f1f5f9; color: #64748b; }
.detail-value.status-badge.status-aborted { background: #f1f5f9; color: #64748b; }

.detail-error {
  margin-top: 12px;
  padding: 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.detail-error .error-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #dc2626;
  margin-bottom: 8px;
}

.detail-error .error-title svg {
  width: 16px;
  height: 16px;
}

.detail-error .error-content {
  font-size: 12px;
  color: #7f1d1d;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.detail-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.detail-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  border: 1px solid #e2e8f0;
  background: white;
  color: #475569;
}

.detail-btn:hover {
  border-color: #94a3b8;
  background: #f8fafc;
}

.detail-btn.primary {
  background: #4299e1;
  color: white;
  border-color: #4299e1;
}

.detail-btn.primary:hover {
  background: #3182ce;
  border-color: #3182ce;
}

.detail-btn svg {
  width: 16px;
  height: 16px;
}

/* 配置 */
.config-section {
  margin-bottom: 32px;
}

.config-section:last-child {
  margin-bottom: 0;
}

.config-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  margin: 0 0 16px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid #e2e8f0;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.config-label {
  font-size: 13px;
  color: #94a3b8;
}

.config-value {
  font-size: 14px;
  color: #1a202c;
}

.config-value.code {
  font-family: 'Consolas', monospace;
  background: #f1f5f9;
  padding: 8px 12px;
  border-radius: 6px;
  word-break: break-all;
}

.env-vars-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.env-var-item {
  display: flex;
  gap: 16px;
  padding: 10px 16px;
  background: #f8fafc;
  border-radius: 8px;
}

.env-name {
  font-weight: 600;
  color: #4a5568;
  min-width: 150px;
}

.env-value {
  color: #64748b;
  font-family: monospace;
}

.config-json {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  font-family: 'Consolas', monospace;
  font-size: 13px;
  overflow-x: auto;
}

/* 错误状态 */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
  text-align: center;
}

.error-state svg {
  width: 64px;
  height: 64px;
  color: #dc2626;
  margin-bottom: 20px;
}

.error-state h3 {
  font-size: 20px;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.error-state p {
  color: #64748b;
  margin: 0 0 24px 0;
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

.btn-primary:hover {
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
}

.btn-success {
  background: linear-gradient(135deg, #48bb78, #38a169);
  color: white;
}

.btn-success:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(72, 187, 120, 0.4);
}

.btn-warning {
  background: linear-gradient(135deg, #ed8936, #dd6b20);
  color: white;
}

.btn-warning:hover {
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
  .status-cards {
    grid-template-columns: 1fr;
  }

  .quick-actions {
    grid-template-columns: repeat(2, 1fr);
  }

  .stage-details {
    grid-template-columns: 1fr;
  }

  .config-grid {
    grid-template-columns: 1fr;
  }
}

/* 错误信息样式 */
.error-section .section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e53e3e;
}

.error-section .error-icon {
  width: 20px;
  height: 20px;
  stroke: #e53e3e;
}

.error-box {
  background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
  border: 1px solid #fc8181;
  border-radius: 12px;
  padding: 20px;
  margin-top: 12px;
}

.error-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.error-message {
  color: #c53030;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.6;
  margin: 0;
  word-break: break-word;
}

.error-time {
  color: #9b2c2c;
  font-size: 12px;
  margin: 0;
  opacity: 0.8;
}

@media (max-width: 768px) {
  .pipeline-header {
    flex-direction: column;
    gap: 20px;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .btn {
    flex: 1;
  }

  .tab-nav {
    overflow-x: auto;
  }

  .tab-btn {
    white-space: nowrap;
  }
}

/* ==================== 版本信息展示 ==================== */
.section-title .title-icon {
  width: 20px;
  height: 20px;
  vertical-align: middle;
  margin-right: 8px;
}

.version-info-card {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: 12px;
  padding: 20px;
  margin-top: 12px;
}

.version-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.version-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-item.full {
  grid-column: span 2;
}

.version-label {
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
}

.version-value {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
}

.version-value.code-text {
  font-family: 'JetBrains Mono', Monaco, Consolas, monospace;
  font-size: 13px;
  background: #f1f5f9;
  padding: 8px 12px;
  border-radius: 6px;
  word-break: break-all;
}

.version-value.digest {
  font-size: 11px;
  color: #64748b;
}

.version-value.tag {
  display: inline-block;
  background: #dbeafe;
  color: #1d4ed8;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 13px;
}

.deploy-status {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.deploy-status.status-success {
  background: #d1fae5;
  color: #065f46;
}

.deploy-status.status-failed {
  background: #fee2e2;
  color: #991b1b;
}

.deploy-status.status-pending {
  background: #f3f4f6;
  color: #6b7280;
}

.deploy-status.status-deploying {
  background: #fef3c7;
  color: #92400e;
}

.deploy-status.status-approval_pending {
  background: #fef9c3;
  color: #854d0e;
}

.deploy-target-info {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed #bae6fd;
}

.target-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
  font-weight: 500;
  margin-bottom: 8px;
}

.target-label svg {
  width: 16px;
  height: 16px;
}

.target-value {
  font-size: 14px;
  color: #1e293b;
  font-weight: 500;
}

.container-name {
  color: #64748b;
  font-weight: 400;
}

/* 自动部署配置展示 */
.auto-deploy-config-display {
  background: #f8fafc;
  border-radius: 12px;
  padding: 20px;
  margin-top: 12px;
}

.config-title .title-icon {
  width: 18px;
  height: 18px;
  vertical-align: middle;
  margin-right: 6px;
}

.status-badge.enabled {
  background: #d1fae5;
  color: #065f46;
}

.status-badge.warning {
  background: #fef3c7;
  color: #92400e;
}

.env-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.env-badge.env-dev {
  background: #d1fae5;
  color: #065f46;
}

.env-badge.env-staging {
  background: #fef3c7;
  color: #92400e;
}

.env-badge.env-prod {
  background: #fee2e2;
  color: #991b1b;
}

@media (max-width: 640px) {
  .version-grid {
    grid-template-columns: 1fr;
  }

  .version-item.full {
    grid-column: span 1;
  }
}

/* ==================== 运行弹窗样式 ==================== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: white;
  border-radius: 16px;
  width: 500px;
  max-width: 90vw;
  max-height: 80vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.modal-header h3 svg {
  width: 22px;
  height: 22px;
  color: #48bb78;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: #f7fafc;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #e2e8f0;
}

.close-btn svg {
  width: 18px;
  height: 18px;
  color: #718096;
}

.modal-body {
  padding: 24px;
  max-height: 60vh;
  overflow-y: auto;
}

.run-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 12px;
  color: #718096;
  font-weight: 500;
}

.info-value {
  font-size: 14px;
  color: #1a202c;
  font-weight: 600;
}

.deploy-config-section {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 12px;
  padding: 16px;
}

.deploy-config-section .section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: #0369a1;
}

.deploy-config-section .section-title svg {
  width: 18px;
  height: 18px;
}

.deploy-config-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.config-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-key {
  font-size: 13px;
  color: #64748b;
}

.config-val {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
}

.config-val.env-dev {
  color: #16a34a;
}

.config-val.env-staging {
  color: #d97706;
}

.config-val.env-prod {
  color: #dc2626;
}

.approval-notice {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px 12px;
  background: #fef3c7;
  border-radius: 8px;
  font-size: 13px;
  color: #92400e;
}

.approval-notice svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.no-deploy-notice {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  background: #f1f5f9;
  border-radius: 12px;
  font-size: 14px;
  color: #64748b;
}

.no-deploy-notice svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  background: #f7fafc;
  border-top: 1px solid #e2e8f0;
}

.modal-footer .btn {
  min-width: 100px;
}

/* 阶段类型标签 */
.stage-type-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.stage-type-badge.approval {
  background: #fef3c7;
  color: #92400e;
}

/* 审批标签状态变体 */
.stage-type-badge.approval.approval-waiting,
.stage-type-badge.approval.approval-pending {
  background: #fef3c7;
  color: #92400e;
}

.stage-type-badge.approval.approval-success,
.stage-type-badge.approval.approval-approved {
  background: #d1fae5;
  color: #059669;
}

.stage-type-badge.approval.approval-failed,
.stage-type-badge.approval.approval-rejected {
  background: #fee2e2;
  color: #dc2626;
}

.stage-type-badge.deploy {
  background: #dbeafe;
  color: #1e40af;
}

/* 阶段操作面板 */
.stage-action-panel {
  padding: 16px;
  margin-bottom: 12px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.stage-action-panel .action-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  color: #475569;
  font-size: 14px;
}

.stage-action-panel .action-info svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.stage-action-panel.approval-panel {
  background: #fffbeb;
  border-color: #fde68a;
}

.stage-action-panel.approval-panel .action-info svg {
  color: #d97706;
}

/* 优化后的审批面板 - 参考 KubeSphere/Rancher 设计 */
.approval-panel-enhanced {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border: 2px solid #fcd34d;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
}

.approval-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.approval-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.approval-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.approval-title h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #92400e;
}

.approval-title p {
  margin: 0;
  font-size: 13px;
  color: #a16207;
}

/* 审批选项卡片 */
.approval-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.approval-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.approval-option:hover {
  border-color: #d1d5db;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.approval-option.selected {
  border-color: #10b981;
  background: #ecfdf5;
}

.approval-option.selected .option-radio .radio-inner {
  transform: scale(1);
}

.option-radio {
  width: 20px;
  height: 20px;
  border: 2px solid #d1d5db;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}

.approval-option.selected .option-radio {
  border-color: #10b981;
  background: #10b981;
}

.radio-inner {
  width: 8px;
  height: 8px;
  background: white;
  border-radius: 50%;
  transform: scale(0);
  transition: transform 0.2s;
}

.option-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.option-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.option-icon.approve {
  background: #d1fae5;
  color: #059669;
}

.option-icon.reject {
  background: #fee2e2;
  color: #dc2626;
}

.option-icon svg {
  width: 18px;
  height: 18px;
}

.option-label {
  font-size: 15px;
  font-weight: 600;
  color: #374151;
}

.option-desc {
  font-size: 12px;
  color: #6b7280;
  margin-left: auto;
}

/* 审批备注 */
.approval-comment {
  margin-bottom: 16px;
}

.comment-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 8px;
}

.comment-label .optional {
  font-weight: 400;
  color: #9ca3af;
}

.comment-input {
  width: 100%;
  padding: 12px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 13px;
  resize: vertical;
  transition: all 0.2s;
  box-sizing: border-box;
}

.comment-input:focus {
  outline: none;
  border-color: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.1);
}

/* 审批按钮 */
.approval-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-approval {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  font-size: 14px;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
}

.btn-approval.approve {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.btn-approval.approve:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
}

.btn-approval.reject {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.btn-approval.reject:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.4);
}

.btn-approval:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

.btn-approval svg {
  width: 18px;
  height: 18px;
}

.btn-approval .loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 审批结果面板样式 */
.approval-result-panel {
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 16px;
}

.approval-result-panel.approved {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  border: 2px solid #34d399;
}

.approval-result-panel.rejected {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border: 2px solid #f87171;
}

.approval-result-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.result-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.result-icon.approved {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.result-icon.rejected {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.result-icon svg {
  width: 24px;
  height: 24px;
}

.result-content h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
}

.approval-result-panel.approved .result-content h4 {
  color: #059669;
}

.approval-result-panel.rejected .result-content h4 {
  color: #dc2626;
}

.result-content p {
  margin: 0;
  font-size: 13px;
  color: #6b7280;
}

.approval-meta {
  display: flex;
  gap: 20px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  font-size: 13px;
  color: #6b7280;
}

.approval-comment-display {
  margin-top: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
  font-size: 13px;
}

.approval-comment-display .comment-label {
  display: inline;
  font-weight: 500;
  color: #374151;
  margin-right: 8px;
}

.approval-comment-display .comment-text {
  color: #4b5563;
}

.stage-action-panel.deploy-panel {
  background: #eff6ff;
  border-color: #bfdbfe;
}

.stage-action-panel.deploy-panel .action-info svg {
  color: #2563eb;
}

/* 配置警告样式 */
.config-warning {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  margin-bottom: 12px;
  background: #fffbeb;
  border: 1px solid #fcd34d;
  border-radius: 8px;
}

.config-warning > svg {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  color: #d97706;
}

.config-warning .warning-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.config-warning .warning-text {
  font-size: 13px;
  color: #92400e;
  line-height: 1.5;
}

.config-warning .config-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #2563eb;
  text-decoration: none;
  transition: color 0.15s;
}

.config-warning .config-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.config-warning .config-link svg {
  width: 14px;
  height: 14px;
}

.stage-action-panel .action-buttons {
  display: flex;
  gap: 10px;
}

.stage-action-panel .btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.stage-action-panel .btn svg {
  width: 16px;
  height: 16px;
}

.stage-action-panel .btn.btn-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.stage-action-panel .btn.btn-success:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.stage-action-panel .btn.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.stage-action-panel .btn.btn-danger:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.stage-action-panel .btn.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

.stage-action-panel .btn.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.stage-action-panel .btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

/* 部署信息 */
.deploy-info {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #64748b;
}

.deploy-info span {
  display: flex;
  align-items: center;
}

/* 部署成功信息 */
.deploy-success-info {
  padding: 16px;
  margin-bottom: 12px;
  background: #f0fdf4;
  border: 1px solid #86efac;
  border-radius: 8px;
}

.deploy-success-info .success-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #16a34a;
}

.deploy-success-info .success-badge svg {
  width: 20px;
  height: 20px;
}

.deploy-success-info .deploy-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 13px;
  color: #15803d;
}

/* 部署进行中状态 */
.deploy-progress-panel {
  padding: 16px;
  margin-bottom: 12px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
}

.deploy-progress-panel .progress-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #d97706;
}

.deploy-progress-panel .progress-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid #fde68a;
  border-top-color: #d97706;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.deploy-progress-panel .deploy-info-mini {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 10px;
  font-size: 12px;
  color: #92400e;
}

/* 部署失败信息 */
.deploy-failed-info {
  padding: 16px;
  margin-bottom: 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.deploy-failed-info .failed-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #dc2626;
}

.deploy-failed-info .failed-badge svg {
  width: 20px;
  height: 20px;
}

.deploy-failed-info .failed-reason {
  padding: 10px;
  margin-bottom: 10px;
  background: #fee2e2;
  border-radius: 6px;
  font-size: 13px;
}

.deploy-failed-info .failed-reason .reason-label {
  font-weight: 500;
  color: #991b1b;
  margin-right: 8px;
}

.deploy-failed-info .failed-reason .reason-text {
  color: #dc2626;
  word-break: break-all;
}

.deploy-failed-info .deploy-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 13px;
  color: #7f1d1d;
}

/* 重新部署按钮 */
.retry-deploy-actions {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #fecaca;
}

.btn-retry {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #dc2626;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-retry:hover:not(:disabled) {
  background: #b91c1c;
}

.btn-retry:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-retry svg {
  width: 16px;
  height: 16px;
}

/* 部署操作按钮区域（回滚/取消） */
.deploy-actions {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  display: flex;
  gap: 10px;
}

.deploy-success-info .deploy-actions {
  border-top-color: #bbf7d0;
}

.deploy-progress-panel .deploy-actions {
  border-top-color: #fde68a;
}

/* 回滚按钮（参考 Rancher/KubeSphere 设计） */
.btn-rollback {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fff7ed;
  color: #c2410c;
  border: 1px solid #fed7aa;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-rollback:hover:not(:disabled) {
  background: #ffedd5;
  border-color: #fb923c;
}

.btn-rollback:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-rollback svg {
  width: 16px;
  height: 16px;
}

/* 回滚按钮组布局 */
.rollback-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

/* 指定版本回滚按钮 */
.btn-rollback-select {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #f8fafc;
  color: #475569;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-rollback-select:hover:not(:disabled) {
  background: #f1f5f9;
  border-color: #94a3b8;
  color: #334155;
}

.btn-rollback-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-rollback-select svg {
  width: 16px;
  height: 16px;
}

/* 回滚按钮 Mini 版本（头部显示） */
.btn-rollback-mini {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  margin-left: auto;
  margin-right: 8px;
  background: #fff7ed;
  color: #c2410c;
  border: 1px solid #fed7aa;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-rollback-mini:hover:not(:disabled) {
  background: #ffedd5;
  border-color: #fb923c;
}

.btn-rollback-mini:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-rollback-mini svg {
  width: 14px;
  height: 14px;
}

/* 取消按钮 Mini 版本（头部显示） */
.btn-cancel-mini {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  margin-left: auto;
  margin-right: 8px;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel-mini:hover:not(:disabled) {
  background: #fee2e2;
  border-color: #f87171;
}

.btn-cancel-mini:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel-mini svg {
  width: 14px;
  height: 14px;
}

/* 取消部署按钮 */
.btn-cancel {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel:hover:not(:disabled) {
  background: #fee2e2;
  border-color: #f87171;
}

.btn-cancel:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel svg {
  width: 16px;
  height: 16px;
}

/* 部署日志展示 */
.deploy-logs-panel {
  margin-bottom: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.deploy-logs-panel .logs-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #f8fafc;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  transition: background 0.15s;
}

.deploy-logs-panel .logs-toggle:hover {
  background: #f1f5f9;
}

.deploy-logs-panel .toggle-icon {
  width: 16px;
  height: 16px;
  transition: transform 0.2s;
}

.deploy-logs-panel .toggle-icon.expanded {
  transform: rotate(180deg);
}

.deploy-logs-panel .deploy-logs-content {
  margin: 0;
  padding: 14px;
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
}

/* 版本变更卡片（参考 Rancher/Kuboard 风格） */
.version-change-card {
  margin: 12px 0;
  padding: 14px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: 10px;
}

.version-change-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
  color: #0369a1;
}

.version-change-header svg {
  width: 18px;
  height: 18px;
}

.version-change-content {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.version-change-content .version-item {
  flex: 1;
  min-width: 200px;
  padding: 10px 12px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.version-change-content .version-item.old {
  border-left: 3px solid #f59e0b;
}

.version-change-content .version-item.new {
  border-left: 3px solid #10b981;
}

.version-change-content .version-label {
  display: block;
  font-size: 11px;
  font-weight: 500;
  color: #64748b;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.version-change-content .version-value {
  display: block;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
  color: #1e293b;
  word-break: break-all;
}

.version-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: #0ea5e9;
  border-radius: 50%;
  flex-shrink: 0;
}

.version-arrow svg {
  width: 16px;
  height: 16px;
  color: white;
}

/* 部署进行中的版本变更迷你版 */
.version-change-mini {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin: 10px 0;
  background: #f8fafc;
  border-radius: 8px;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
}

.version-change-mini .old-version {
  color: #f59e0b;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-change-mini .new-version {
  color: #10b981;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-change-mini .arrow-icon {
  width: 16px;
  height: 16px;
  color: #64748b;
  flex-shrink: 0;
}

/* 部署日志预览（部署进行中时显示） */
.deploy-logs-preview {
  margin-top: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.logs-preview-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f1f5f9;
  font-size: 12px;
  font-weight: 500;
  color: #475569;
}

.logs-preview-header svg {
  width: 14px;
  height: 14px;
}

.logs-preview-content {
  margin: 0;
  padding: 12px;
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 11px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
}

/* 阶段错误信息 */
.stage-error {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  margin-bottom: 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  font-size: 13px;
  color: #dc2626;
}

.stage-error svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  margin-top: 2px;
}

/* 阶段步骤 */
.stage-steps {
  margin-bottom: 12px;
}

/* 等待审批状态 */
.status-dot.status-waiting {
  background: #f59e0b;
  animation: pulse 1.5s infinite;
}

/* 已跳过状态 */
.status-dot.status-skipped {
  background: #94a3b8;
}

/* 版本选择弹窗 */
.version-dialog {
  width: 600px;
  max-width: 90vw;
}

.version-loading,
.version-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #718096;
}

.version-loading svg,
.version-empty svg {
  width: 48px;
  height: 48px;
  margin-bottom: 12px;
}

.version-list {
  max-height: 400px;
  overflow-y: auto;
}

.version-item {
  padding: 14px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.15s;
}

.version-item:hover {
  background: #f8fafc;
  border-color: #cbd5e0;
}

.version-item.selected {
  background: #ebf8ff;
  border-color: #4299e1;
}

.version-item.current {
  background: #f0fdf4;
  border-color: #86efac;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.version-revision {
  font-weight: 600;
  color: #1a202c;
}

.current-badge {
  padding: 2px 8px;
  background: #dcfce7;
  color: #16a34a;
  font-size: 11px;
  font-weight: 500;
  border-radius: 10px;
}

.version-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 6px;
}

.version-rs {
  font-size: 13px;
  color: #4a5568;
  font-family: 'Monaco', 'Menlo', monospace;
}

.version-image {
  font-size: 12px;
  color: #718096;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-time {
  font-size: 12px;
  color: #a0aec0;
}

.btn-warning {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: #f59e0b;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
}

.btn-warning:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-warning svg {
  width: 18px;
  height: 18px;
}

/* 回滚结果弹窗 */
.rollback-result-dialog {
  width: 600px;
  max-width: 90vw;
}

.rollback-result-dialog .modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-radius: 12px 12px 0 0;
}

.rollback-result-dialog .modal-header.success {
  background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
}

.rollback-result-dialog .modal-header.error {
  background: linear-gradient(135deg, #fef2f2 0%, #fecaca 100%);
}

.rollback-result-dialog .modal-header.in-progress {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
}

.rollback-result-dialog .modal-header h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.rollback-result-dialog .modal-header.success h3 {
  color: #166534;
}

.rollback-result-dialog .modal-header.error h3 {
  color: #dc2626;
}

.rollback-result-dialog .modal-header.in-progress h3 {
  color: #1d4ed8;
}

.rollback-result-dialog .modal-header h3 svg {
  width: 24px;
  height: 24px;
}

.rollback-result-dialog .modal-header h3 .progress-spinner {
  width: 22px;
  height: 22px;
  border: 3px solid #93c5fd;
  border-top-color: #1d4ed8;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.rollback-details {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  margin-bottom: 20px;
  background: #f8fafc;
  padding: 16px;
  border-radius: 8px;
}

.rollback-details .detail-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.rollback-details .label {
  flex-shrink: 0;
  width: 100px;
  color: #64748b;
  font-size: 13px;
}

.rollback-details .value {
  color: #1a202c;
  font-size: 13px;
  word-break: break-all;
}

.rollback-details .value.image {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #4a5568;
}

/* 回滚日志面板（类似部署日志风格） */
.rollback-log-panel {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.rollback-log-panel .log-header {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #1e293b;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 500;
  color: #94a3b8;
  border-bottom: 1px solid #334155;
}

.rollback-log-panel .log-header svg {
  width: 16px;
  height: 16px;
  stroke: #94a3b8;
}

.rollback-log-panel .log-header .log-loading {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  color: #60a5fa;
  font-size: 12px;
}

.rollback-log-panel .log-header .mini-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid #334155;
  border-top-color: #60a5fa;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.rollback-log-panel .log-content {
  background: #0f172a;
  color: #e2e8f0;
  padding: 14px;
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.7;
  max-height: 280px;
  min-height: 150px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 旧版本回滚日志样式（兼容） */
.rollback-log {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.rollback-log .log-header {
  background: #f1f5f9;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
}

.rollback-log .log-content {
  background: #1e293b;
  color: #e2e8f0;
  padding: 14px;
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 失败阶段摘要（参考 Jenkins Blue Ocean 风格） */
.failed-stages-summary {
  margin-top: 24px;
  background: linear-gradient(135deg, #fef2f2 0%, #fff5f5 100%);
  border: 1px solid #fecaca;
  border-radius: 12px;
  overflow: hidden;
}

.failed-summary-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: linear-gradient(90deg, #dc2626 0%, #ef4444 100%);
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.failed-summary-header svg {
  width: 20px;
  height: 20px;
  stroke: white;
}

.failed-stages-list {
  padding: 12px;
}

.failed-stage-item {
  background: white;
  border: 1px solid #fecaca;
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.failed-stage-item:last-child {
  margin-bottom: 0;
}

.failed-stage-item:hover {
  background: #fff5f5;
  border-color: #f87171;
  box-shadow: 0 2px 8px rgba(220, 38, 38, 0.1);
}

.failed-stage-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.failed-stage-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #fef2f2;
  border-radius: 50%;
  flex-shrink: 0;
}

.failed-stage-icon svg {
  width: 18px;
  height: 18px;
  stroke: #dc2626;
}

.failed-stage-content {
  flex: 1;
  min-width: 0;
}

.failed-stage-name {
  display: block;
  font-weight: 600;
  font-size: 14px;
  color: #1a202c;
  margin-bottom: 2px;
}

.failed-stage-status {
  font-size: 12px;
  color: #dc2626;
  font-weight: 500;
}

.failed-stage-duration {
  font-size: 13px;
  color: #64748b;
  flex-shrink: 0;
}

.failed-stage-error {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  background: #fef2f2;
  border-radius: 6px;
  border-left: 3px solid #dc2626;
}

.failed-stage-error svg {
  width: 18px;
  height: 18px;
  stroke: #dc2626;
  flex-shrink: 0;
  margin-top: 1px;
}

.failed-stage-error .error-text {
  font-size: 13px;
  color: #991b1b;
  line-height: 1.5;
  word-break: break-word;
}

.failed-stage-error.no-message {
  background: #f8fafc;
  border-left-color: #94a3b8;
}

.failed-stage-error.no-message svg {
  stroke: #64748b;
}

.failed-stage-error.no-message .error-text {
  color: #64748b;
  font-style: italic;
}

/* 阶段详情卡片（参考 Rancher/Kuboard/KubeSphere 风格） */
.stage-detail-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 阶段概览卡片 */
.stage-overview-card {
  background: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.stage-overview-card.status-success {
  border-color: #86efac;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
}

.stage-overview-card.status-failed {
  border-color: #fca5a5;
  background: linear-gradient(135deg, #fef2f2 0%, #fecaca 100%);
}

.stage-overview-card.status-running {
  border-color: #93c5fd;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.overview-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.overview-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.overview-icon svg {
  width: 28px;
  height: 28px;
}

.status-success .overview-icon svg { stroke: #16a34a; }
.status-failed .overview-icon svg { stroke: #dc2626; }
.status-running .overview-icon svg { stroke: #2563eb; }
.status-pending .overview-icon svg { stroke: #64748b; }

.overview-title {
  flex: 1;
}

.overview-title h4 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.overview-title .status-label {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-label.status-success { background: #dcfce7; color: #166534; }
.status-label.status-failed { background: #fef2f2; color: #dc2626; }
.status-label.status-running { background: #dbeafe; color: #1d4ed8; }
.status-label.status-pending { background: #f1f5f9; color: #64748b; }

.overview-duration {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #475569;
}

.overview-duration svg {
  width: 20px;
  height: 20px;
  stroke: #64748b;
}

/* 阶段信息网格 */
.stage-info-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

@media (max-width: 900px) {
  .stage-info-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.info-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: all 0.2s;
}

.info-card:hover {
  border-color: #cbd5e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.info-card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  flex-shrink: 0;
}

.info-card-icon svg {
  width: 20px;
  height: 20px;
}

.info-card-icon.type {
  background: #f0f9ff;
}

.info-card-icon.type svg { stroke: #0284c7; }

.info-card-icon.time {
  background: #f5f3ff;
}

.info-card-icon.time svg { stroke: #7c3aed; }

.info-card-icon.duration {
  background: #fef3c7;
}

.info-card-icon.duration svg { stroke: #d97706; }

.info-card-icon.duration.running {
  background: #dbeafe;
  animation: pulse 1.5s infinite;
}

.info-card-icon.duration.running svg { stroke: #2563eb; }

.info-card-content {
  flex: 1;
  min-width: 0;
}

.info-card-label {
  display: block;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 4px;
}

.info-card-value {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  word-break: break-all;
}

.info-card-value.duration.running {
  color: #2563eb;
}

/* 运行中进度面板 */
.stage-progress-panel {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 1px solid #93c5fd;
  border-radius: 12px;
  padding: 20px;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  font-size: 15px;
  font-weight: 600;
  color: #1d4ed8;
}

.progress-indicator {
  width: 12px;
  height: 12px;
  background: #2563eb;
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}

.progress-bar-container {
  margin-bottom: 16px;
}

.progress-bar-track {
  height: 8px;
  background: #bfdbfe;
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  width: 60%;
  background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%);
  border-radius: 4px;
  animation: progress-indeterminate 1.5s ease-in-out infinite;
}

@keyframes progress-indeterminate {
  0% { transform: translateX(-100%); width: 30%; }
  50% { width: 60%; }
  100% { transform: translateX(200%); width: 30%; }
}

.progress-tips {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 8px;
  font-size: 13px;
  color: #475569;
}

.progress-tips svg {
  width: 16px;
  height: 16px;
  stroke: #64748b;
  flex-shrink: 0;
}

/* 失败错误面板 */
.stage-error-panel {
  background: linear-gradient(135deg, #fef2f2 0%, #fecaca 100%);
  border: 1px solid #fca5a5;
  border-radius: 12px;
  overflow: hidden;
}

.error-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: linear-gradient(90deg, #dc2626 0%, #ef4444 100%);
  color: white;
  font-weight: 600;
  font-size: 15px;
}

.error-panel-header svg {
  width: 22px;
  height: 22px;
  stroke: white;
}

.error-panel-content {
  padding: 20px;
}

.error-message-box {
  background: white;
  border: 1px solid #fecaca;
  border-radius: 8px;
  padding: 16px;
}

.error-message-box .error-label {
  font-size: 12px;
  font-weight: 600;
  color: #dc2626;
  margin-bottom: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.error-message-box .error-text {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  color: #991b1b;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
  background: #fef2f2;
  padding: 12px;
  border-radius: 6px;
  max-height: 200px;
  overflow-y: auto;
}

.error-message-box.no-detail .error-hint {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

.error-panel-actions {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.5);
  border-top: 1px solid #fecaca;
}

.btn-view-logs {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #475569;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-view-logs:hover {
  background: #f8fafc;
  border-color: #cbd5e0;
}

.btn-view-logs svg {
  width: 18px;
  height: 18px;
}

.btn-retry-stage {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: #dc2626;
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-retry-stage:hover:not(:disabled) {
  background: #b91c1c;
}

.btn-retry-stage:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-retry-stage svg {
  width: 18px;
  height: 18px;
}

/* 成功信息面板 */
.stage-success-panel {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  border: 1px solid #86efac;
  border-radius: 12px;
  overflow: hidden;
}

.success-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: linear-gradient(90deg, #16a34a 0%, #22c55e 100%);
  color: white;
  font-weight: 600;
  font-size: 15px;
}

.success-panel-header svg {
  width: 22px;
  height: 22px;
  stroke: white;
}

.success-artifact {
  padding: 16px 20px;
}

.artifact-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #166534;
  margin-bottom: 10px;
}

.artifact-label svg {
  width: 18px;
  height: 18px;
  stroke: #16a34a;
}

.artifact-content {
  display: flex;
  align-items: center;
  gap: 10px;
  background: white;
  border: 1px solid #86efac;
  border-radius: 8px;
  padding: 12px 16px;
}

.artifact-image {
  flex: 1;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  color: #374151;
  word-break: break-all;
}

.copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: #f0fdf4;
  border: 1px solid #86efac;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #dcfce7;
}

.copy-btn svg {
  width: 16px;
  height: 16px;
  stroke: #16a34a;
}

/* 日志预览区域 */
.stage-logs-preview {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
}

.logs-preview-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}

.logs-preview-header:hover {
  background: #f1f5f9;
}

.logs-preview-header svg {
  width: 18px;
  height: 18px;
  stroke: #64748b;
}

.logs-preview-header .toggle-arrow {
  transition: transform 0.2s;
}

.logs-preview-header .toggle-arrow.expanded {
  transform: rotate(180deg);
}

.logs-preview-header span {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #475569;
}

.view-full-logs {
  padding: 6px 14px;
  background: #4299e1;
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.view-full-logs:hover {
  background: #3182ce;
}

.logs-preview-content {
  padding: 0;
}

.logs-preview-content pre {
  margin: 0;
  padding: 16px;
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.7;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-preview-empty {
  padding: 40px 20px;
  text-align: center;
  color: #94a3b8;
  font-size: 14px;
}
</style>

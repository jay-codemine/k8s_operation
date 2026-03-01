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
            <div class="meta-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="3" width="20" height="14" rx="2"/>
                <line x1="8" y1="21" x2="16" y2="21"/>
                <line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
              <span>{{ pipeline.jenkins_job || '未配置' }}</span>
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

          <!-- 阶段流水线视图 -->
          <div v-else-if="filteredStages.length > 0" class="stages-pipeline">
            <div
              v-for="(stage, index) in filteredStages"
              :key="stage.name"
              :class="['stage-node', `status-${stage.status}`]"
            >
              <div class="stage-connector" v-if="index > 0"></div>
              <div class="stage-content">
                <div class="stage-icon">
                  <svg v-if="stage.status === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                  <svg v-else-if="stage.status === 'failed'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                  <svg v-else-if="stage.status === 'running'" class="spinning" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                  </svg>
                  <svg v-else-if="stage.status === 'waiting'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                  </svg>
                </div>
                <div class="stage-info">
                  <span class="stage-name">{{ stage.name }}</span>
                  <span class="stage-duration">{{ stage.duration || '-' }}</span>
                </div>
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

          <!-- 阶段详情 -->
          <div v-if="filteredStages.length > 0" class="stage-details">
            <div v-for="stage in filteredStages" :key="stage.name" class="stage-detail-card">
              <div class="detail-header" @click="toggleStageExpand(stage.name)">
                <span :class="['status-dot', `status-${stage.status}`]"></span>
                <span class="stage-title">{{ stage.name }}</span>
                <!-- 阶段类型标签 -->
                <span v-if="stage.type === 'approval'" :class="['stage-type-badge', 'approval', `approval-${stage.status}`]">{{ approvalBadgeText(stage.status) }}</span>
                <span v-if="stage.type === 'deploy'" class="stage-type-badge deploy">部署</span>
                <span class="stage-status">{{ stageStatusText(stage.status) }}</span>
                <svg :class="['expand-icon', { expanded: expandedStages.includes(stage.name) }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </div>
              <div v-show="expandedStages.includes(stage.name)" class="detail-body">
                <!-- 审批阶段操作 - 参考 KubeSphere/Rancher 设计 -->
                <div v-if="stage.type === 'approval' && (stage.status === 'waiting' || stage.status === 'pending')" class="stage-action-panel approval-panel-enhanced">
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

                  <!-- 审批选项 -->
                  <div class="approval-options">
                    <label :class="['approval-option', { selected: approvalDecision === 'approve' }]" @click="approvalDecision = 'approve'">
                      <div class="option-radio">
                        <span class="radio-inner"></span>
                      </div>
                      <div class="option-content">
                        <span class="option-icon approve">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                            <polyline points="20 6 9 17 4 12"/>
                          </svg>
                        </span>
                        <span class="option-label">通过</span>
                        <span class="option-desc">确认并继续执行部署</span>
                      </div>
                    </label>
                    <label :class="['approval-option', { selected: approvalDecision === 'reject' }]" @click="approvalDecision = 'reject'">
                      <div class="option-radio">
                        <span class="radio-inner"></span>
                      </div>
                      <div class="option-content">
                        <span class="option-icon reject">
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                            <line x1="18" y1="6" x2="6" y2="18"/>
                            <line x1="6" y1="6" x2="18" y2="18"/>
                          </svg>
                        </span>
                        <span class="option-label">拒绝</span>
                        <span class="option-desc">取消本次部署</span>
                      </div>
                    </label>
                  </div>

                  <!-- 审批备注 -->
                  <div class="approval-comment">
                    <label class="comment-label">审批备注 <span class="optional">(可选)</span></label>
                    <textarea
                      v-model="approvalComment"
                      class="comment-input"
                      placeholder="请输入审批备注..."
                      rows="3"
                    ></textarea>
                  </div>

                  <!-- 提交按钮 -->
                  <div class="approval-actions">
                    <button
                      :class="['btn', 'btn-approval', approvalDecision === 'approve' ? 'approve' : 'reject']"
                      @click.stop="submitApproval(stage.id)"
                      :disabled="approving"
                    >
                      <svg v-if="approving" class="loading-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"/>
                      </svg>
                      <svg v-else-if="approvalDecision === 'approve'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"/>
                      </svg>
                      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="18" y1="6" x2="6" y2="18"/>
                        <line x1="6" y1="6" x2="18" y2="18"/>
                      </svg>
                      {{ approving ? '处理中...' : (approvalDecision === 'approve' ? '确认通过' : '确认拒绝') }}
                    </button>
                  </div>
                </div>

                <!-- 审批已通过状态展示 -->
                <div v-if="stage.type === 'approval' && (stage.status === 'success' || stage.status === 'approved')" class="stage-action-panel approval-result-panel approved">
                  <div class="approval-result-header">
                    <div class="result-icon approved">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                        <polyline points="22 4 12 14.01 9 11.01"/>
                      </svg>
                    </div>
                    <div class="result-content">
                      <h4>审批已通过</h4>
                      <p>该阶段已经审批通过，可以继续执行部署</p>
                    </div>
                  </div>
                  <div v-if="stage.approval_info" class="approval-meta">
                    <span v-if="stage.approval_info.approver_name">审批人: {{ stage.approval_info.approver_name }}</span>
                    <span v-if="stage.approval_info.approved_at">审批时间: {{ formatFullDate(stage.approval_info.approved_at) }}</span>
                  </div>
                  <div v-if="stage.approval_info && stage.approval_info.comment" class="approval-comment-display">
                    <span class="comment-label">审批备注:</span>
                    <span class="comment-text">{{ stage.approval_info.comment }}</span>
                  </div>
                </div>

                <!-- 审批已拒绝状态展示 -->
                <div v-if="stage.type === 'approval' && (stage.status === 'failed' || stage.status === 'rejected')" class="stage-action-panel approval-result-panel rejected">
                  <div class="approval-result-header">
                    <div class="result-icon rejected">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"/>
                        <line x1="15" y1="9" x2="9" y2="15"/>
                        <line x1="9" y1="9" x2="15" y2="15"/>
                      </svg>
                    </div>
                    <div class="result-content">
                      <h4>审批已拒绝</h4>
                      <p>该阶段审批被拒绝，部署已取消</p>
                    </div>
                  </div>
                  <div v-if="stage.approval_info" class="approval-meta">
                    <span v-if="stage.approval_info.approver_name">拒绝人: {{ stage.approval_info.approver_name }}</span>
                    <span v-if="stage.approval_info.approved_at">拒绝时间: {{ formatFullDate(stage.approval_info.approved_at) }}</span>
                  </div>
                  <div v-if="stage.approval_info && stage.approval_info.comment" class="approval-comment-display">
                    <span class="comment-label">拒绝原因:</span>
                    <span class="comment-text">{{ stage.approval_info.comment }}</span>
                  </div>
                </div>

                <!-- 部署阶段操作 -->
                <div v-if="stage.type === 'deploy' && stage.can_operate" class="stage-action-panel deploy-panel">
                  <div class="action-info">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                      <line x1="8" y1="21" x2="16" y2="21"/>
                      <line x1="12" y1="17" x2="12" y2="21"/>
                    </svg>
                    <span>点击执行部署到 K8s 集群</span>
                  </div>
                  <div v-if="stage.deploy_info" class="deploy-info">
                    <span>集群: {{ stage.deploy_info.cluster_name || stage.deploy_info.cluster_id }}</span>
                    <span>命名空间: {{ stage.deploy_info.namespace }}</span>
                    <span>工作负载: {{ stage.deploy_info.workload_name }}</span>
                    <span>镜像: {{ stage.deploy_info.image }}</span>
                  </div>
                  <div class="action-buttons">
                    <button class="btn btn-primary" @click.stop="handleDeployStage(stage.id)" :disabled="deploying">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polygon points="5 3 19 12 5 21 5 3"/>
                      </svg>
                      {{ deploying ? '部署中...' : '执行部署' }}
                    </button>
                  </div>
                </div>

                <!-- 部署成功信息 -->
                <div v-if="stage.type === 'deploy' && stage.status === 'success' && stage.deploy_info" class="deploy-success-info">
                  <div class="success-badge">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                      <polyline points="22 4 12 14.01 9 11.01"/>
                    </svg>
                    部署成功
                  </div>
                  <div class="deploy-details">
                    <span>集群: {{ stage.deploy_info.cluster_name || stage.deploy_info.cluster_id }}</span>
                    <span>命名空间: {{ stage.deploy_info.namespace }}</span>
                    <span>工作负载: {{ stage.deploy_info.workload_name }}</span>
                    <span>镜像: {{ stage.deploy_info.image }}</span>
                  </div>
                </div>

                <!-- 部署进行中状态 -->
                <div v-if="stage.type === 'deploy' && stage.status === 'running'" class="deploy-progress-panel">
                  <div class="progress-header">
                    <div class="progress-spinner"></div>
                    <span>部署进行中...</span>
                  </div>
                  <div v-if="stage.deploy_info" class="deploy-info-mini">
                    <span>工作负载: {{ stage.deploy_info.workload_name }}</span>
                    <span>镜像: {{ stage.deploy_info.image }}</span>
                  </div>
                </div>

                <!-- 部署失败信息 -->
                <div v-if="stage.type === 'deploy' && stage.status === 'failed'" class="deploy-failed-info">
                  <div class="failed-badge">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="15" y1="9" x2="9" y2="15"/>
                      <line x1="9" y1="9" x2="15" y2="15"/>
                    </svg>
                    部署失败
                  </div>
                  <div v-if="stage.error_message" class="failed-reason">
                    <span class="reason-label">失败原因:</span>
                    <span class="reason-text">{{ stage.error_message }}</span>
                  </div>
                  <div v-if="stage.deploy_info" class="deploy-details">
                    <span>集群: {{ stage.deploy_info.cluster_name || stage.deploy_info.cluster_id }}</span>
                    <span>命名空间: {{ stage.deploy_info.namespace }}</span>
                    <span>工作负载: {{ stage.deploy_info.workload_name }}</span>
                    <span>镜像: {{ stage.deploy_info.image }}</span>
                  </div>
                  <!-- 重新部署按钮 -->
                  <div class="retry-deploy-actions">
                    <button class="btn btn-retry" @click.stop="handleRetryDeploy(stage.id)" :disabled="deploying">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="23 4 23 10 17 10"/>
                        <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                      </svg>
                      {{ deploying ? '部署中...' : '重新部署' }}
                    </button>
                  </div>
                </div>

                <!-- 部署日志展示（Rollout 步骤） -->
                <div v-if="stage.type === 'deploy' && stage.logs && (stage.status === 'success' || stage.status === 'failed' || stage.status === 'running')" class="deploy-logs-panel">
                  <div class="logs-toggle" @click="stage.showLogs = !stage.showLogs">
                    <svg :class="['toggle-icon', { expanded: stage.showLogs }]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="6 9 12 15 18 9"/>
                    </svg>
                    <span>查看部署日志</span>
                  </div>
                  <pre v-show="stage.showLogs" class="deploy-logs-content">{{ stage.logs }}</pre>
                </div>

                <!-- Jenkins 构建阶段步骤 -->
                <div v-if="stage.steps && stage.steps.length > 0" class="stage-steps">
                  <div v-for="step in stage.steps" :key="step.name" class="step-item">
                    <svg v-if="step.status === 'success'" class="step-icon success" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="20 6 9 17 4 12"/>
                    </svg>
                    <svg v-else-if="step.status === 'failed'" class="step-icon failed" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"/>
                      <line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                    <svg v-else class="step-icon pending" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <circle cx="12" cy="12" r="10"/>
                    </svg>
                    <span class="step-name">{{ step.name }}</span>
                    <span class="step-duration">{{ step.duration || '-' }}</span>
                  </div>
                </div>

                <!-- 错误信息 -->
                <div v-if="stage.error_msg" class="stage-error">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                  <span>{{ stage.error_msg }}</span>
                </div>

                <!-- 查看日志按钮 -->
                <div class="stage-actions">
                  <button v-if="stage.has_logs" class="view-log-btn" @click.stop="viewStageLog(stage)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                      <line x1="16" y1="13" x2="8" y2="13"/>
                      <line x1="16" y1="17" x2="8" y2="17"/>
                    </svg>
                    查看阶段日志
                  </button>
                  <button class="view-log-btn" @click.stop="activeTab = 'logs'; loadLogs()">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                      <polyline points="14 2 14 8 20 8"/>
                      <line x1="16" y1="13" x2="8" y2="13"/>
                      <line x1="16" y1="17" x2="8" y2="17"/>
                    </svg>
                    查看构建日志
                  </button>
                </div>
              </div>
            </div>
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
              :class="['history-item', `status-${run.status}`]"
            >
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
              <div class="history-actions">
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
  executeDeployStage
} from '@/api/platform/pipeline'

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
    const stageFilter = ref('')

    // 运行弹窗相关
    const showRunDialog = ref(false)
    const runSubmitting = ref(false)

    // 审批和部署操作相关
    const approving = ref(false)
    const deploying = ref(false)
    const approvalDecision = ref('approve')  // 默认通过
    const approvalComment = ref('')  // 审批备注

    // 筛选后的阶段
    const filteredStages = computed(() => {
      if (!stageFilter.value) return pipelineStages.value
      return pipelineStages.value.filter(stage => stage.status === stageFilter.value)
    })

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

    // 开始轮询状态和日志
    const startPolling = () => {
      // 清理旧的定时器
      stopPolling()

      // 每 5 秒轮询状态和阶段（参考 Jenkins 默认设计）
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
              await loadStages()
            }

            // 如果不再运行，停止轮询
            if (newPipeline.status !== 'running') {
              stopPolling()
              // 最后再获取一次完整日志
              if (activeTab.value === 'logs') {
                await loadLogs()
              }
              // 如果有错误信息，显示错误
              if (response.data.latest_run && response.data.latest_run.error_message) {
                Message.error({ content: response.data.latest_run.error_message, duration: 5000 })
              } else {
                Message.info({ content: `流水线执行完成，状态: ${runStatusText(newPipeline.last_run_status)}` })
              }
            }
          }
        } catch (error) {
          console.error('轮询状态失败:', error)
        }
      }, 5000)

      // 每 5 秒轮询日志（如果在日志 Tab，参考 Jenkins 默认设计）
      logsPollingTimer = setInterval(async () => {
        if (activeTab.value === 'logs' && pipeline.value.status === 'running') {
          await loadLogs()
        }
      }, 5000)
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
        Message.info({ content: '正在启动流水线...' })
        // 传入 force: true 自动清理旧的失败/运行中构建
        const response = await runPipeline(pipelineId.value, { force: true })
        if (response.code === 0) {
          Message.success({ content: '流水线启动成功，正在执行中...' })
          // 重置日志状态（新构建）
          isFirstLoad.value = true
          logs.value = ''
          logLineCount.value = 0
          logsError.value = ''
          // 刷新状态并开始轮询
          await loadPipeline()
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
          // 重置表单
          approvalDecision.value = 'approve'
          approvalComment.value = ''
          // 刷新阶段数据
          await loadStages()
          await loadPipeline()

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
    // 优化：启动部署后立即轮询状态，直到完成
    const handleDeployStage = async (stageId) => {
      deploying.value = true
      try {
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
    // 优化：重试后立即轮询状态
    const handleRetryDeploy = async (stageId) => {
      deploying.value = true
      try {
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
        pending: '等待',
        waiting: '待通过',
        skipped: '已跳过',
        aborted: '已中止',
        approved: '已通过',
        rejected: '已拒绝'
      }
      return map[status] || status
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
      getHistoryStatusCount,
      pipelineStages,
      stagesLoading,
      expandedStages,
      stageFilter,
      filteredStages,
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
      viewStageLog,
      retryRun,
      handleApproveStage,
      handleDeployStage,
      handleRetryDeploy,
      submitApproval,
      approving,
      deploying,
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
      formatDate,
      formatFullDate,
      formatDuration
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
  transition: all 0.3s ease;
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
  transition: all 0.3s ease;
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
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  transition: all 0.2s;
}

.history-item:hover {
  border-color: #cbd5e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
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
</style>

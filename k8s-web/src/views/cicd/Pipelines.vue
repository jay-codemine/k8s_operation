<template>
  <div class="pipeline-view">
    <!-- 顶部标题区 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <svg class="title-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
          </svg>
          流水线管理
        </h1>
        <p class="page-desc">管理和监控 CI/CD 流水线，实现自动化构建、测试和部署</p>
      </div>
      <div class="header-right">
        <button class="btn btn-outline" @click="loadPipelines" :disabled="loading">
          <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          {{ loading ? '加载中...' : '刷新' }}
        </button>
        <button v-if="canOperate" class="btn btn-primary" @click="createPipeline">
          <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          创建流水线
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon total">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <path d="M3 9h18"/>
            <path d="M9 21V9"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ total }}</span>
          <span class="stat-label">流水线总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon running">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ runningCount }}</span>
          <span class="stat-label">运行中</span>
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
          <span class="stat-value">{{ successCount }}</span>
          <span class="stat-label">上次成功</span>
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
          <span class="stat-value">{{ failedCount }}</span>
          <span class="stat-label">上次失败</span>
        </div>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-bar">
      <div class="search-wrapper">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          placeholder="搜索流水线名称、描述或 Git 仓库..."
        />
      </div>
      <div class="filter-right">
        <div class="filter-tabs">
          <button 
            :class="['filter-tab', { active: statusFilter === '' }]"
            @click="statusFilter = ''"
          >
            全部
          </button>
          <button 
            :class="['filter-tab', { active: statusFilter === 'running' }]"
            @click="statusFilter = 'running'"
          >
            <span class="status-dot running"></span>
            运行中
          </button>
          <button 
            :class="['filter-tab', { active: statusFilter === 'idle' }]"
            @click="statusFilter = 'idle'"
          >
            <span class="status-dot idle"></span>
            空闲
          </button>
        </div>
        <div class="view-switch">
          <button 
            :class="['view-btn', { active: viewMode === 'card' }]"
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M3 3h8v8H3V3zm0 10h8v8H3v-8zm10-10h8v8h-8V3zm0 10h8v8h-8v-8z"/>
            </svg>
          </button>
          <button 
            :class="['view-btn', { active: viewMode === 'table' }]"
            @click="viewMode = 'table'"
            title="列表视图"
          >
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M3 4h18v2H3V4zm0 7h18v2H3v-2zm0 7h18v2H3v-2z"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMsg" class="error-alert">
      <svg class="alert-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      <span>{{ errorMsg }}</span>
      <button class="alert-close" @click="errorMsg = ''">×</button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && pipelines.length === 0" class="loading-container">
      <div class="loading-spinner"></div>
      <p>正在加载流水线...</p>
    </div>
    
    <!-- 空状态 -->
    <div v-else-if="filteredPipelines.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
        </svg>
      </div>
      <h3>暂无流水线</h3>
      <p>点击“创建流水线”按钮开始您的第一条流水线</p>
      <button v-if="canOperate" class="btn btn-primary" @click="createPipeline">
        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        创建流水线
      </button>
    </div>
    
    <!-- 表格视图 -->
    <div v-else-if="viewMode === 'table'" class="table-container">
      <!-- 批量操作工具栏 -->
      <transition name="slide-down">
        <div v-if="selectedIds.length > 0" class="batch-toolbar">
          <div class="batch-info">
            <span class="batch-count">已选择 <strong>{{ selectedIds.length }}</strong> 项</span>
            <button class="batch-clear" @click="clearSelection">取消选择</button>
          </div>
          <div class="batch-actions">
            <button v-if="canOperate" class="batch-btn primary" @click="batchRunPipelines" :disabled="batchRunDisabled">
              <svg viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              批量发布
            </button>
            <button v-if="canOperate" class="batch-btn warning" @click="batchStopPipelines" :disabled="batchStopDisabled">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
              取消发布
            </button>
            <button v-if="canOperate" class="batch-btn danger" @click="batchDeletePipelines">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              批量删除
            </button>
          </div>
        </div>
      </transition>
      <table class="pipeline-table">
        <thead>
          <tr>
            <th style="width: 48px;" class="checkbox-col">
              <label class="checkbox-wrapper">
                <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" :indeterminate="isIndeterminate" />
                <span class="checkmark"></span>
              </label>
            </th>
            <th style="width: 200px;">流水线名称</th>
            <th>描述</th>
            <th style="width: 100px;">状态</th>
            <th style="width: 100px;">上次运行</th>
            <th style="width: 140px;">运行时间</th>
            <th style="width: 200px;">Git 仓库</th>
            <th style="width: 100px;">分支</th>
            <th style="width: 200px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pipeline in paginatedPipelines" :key="pipeline.id" :class="{ 'row-selected': selectedIds.includes(pipeline.id) }">
            <td class="checkbox-col">
              <label class="checkbox-wrapper">
                <input type="checkbox" :checked="selectedIds.includes(pipeline.id)" @change="toggleSelect(pipeline.id)" />
                <span class="checkmark"></span>
              </label>
            </td>
            <td>
              <div class="table-name-cell">
                <span :class="['status-indicator', `status-${pipeline.status}`]"></span>
                <span class="pipeline-link" @click="viewPipeline(pipeline.id)">{{ pipeline.name }}</span>
              </div>
            </td>
            <td>
              <span class="table-desc">{{ pipeline.description || '-' }}</span>
            </td>
            <td>
              <span :class="['status-tag', `status-${pipeline.status}`]">
                {{ pipeline.status === 'running' ? '运行中' : '空闲' }}
              </span>
            </td>
            <td>
              <span :class="['run-status-tag', `status-${pipeline.lastRunStatus}`]">
                {{ runStatusText(pipeline.lastRunStatus) }}
              </span>
            </td>
            <td>
              <span class="table-time">{{ formatDate(pipeline.lastRunTime) }}</span>
            </td>
            <td>
              <span class="table-repo" :title="pipeline.gitRepo">{{ formatRepoUrl(pipeline.gitRepo) }}</span>
            </td>
            <td>
              <span class="table-branch">{{ pipeline.branch || 'main' }}</span>
            </td>
            <td>
              <div class="table-actions">
                <button 
                  v-if="canOperate && pipeline.status !== 'running'"
                  class="action-btn-sm run"
                  @click="handleRunPipeline(pipeline)"
                  title="运行"
                >
                  ▶ 运行
                </button>
                <button 
                  v-if="canOperate && (pipeline.status === 'running' || pipeline.lastRunStatus === 'pending')"
                  class="action-btn-sm stop"
                  @click="handleStopPipeline(pipeline)"
                  title="停止"
                >
                  ■ 停止
                </button>
                <button class="action-btn-sm" @click="viewPipeline(pipeline.id)" title="详情">查看</button>
                <button v-if="canOperate" class="action-btn-sm" @click="editPipeline(pipeline.id)" title="编辑">编辑</button>
                <button v-if="canOperate" class="action-btn-sm danger" @click="handleDeletePipeline(pipeline)" title="删除">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 卡片视图 -->
    <div v-else class="pipeline-grid-wrapper">
      <!-- 批量操作工具栏 -->
      <transition name="slide-down">
        <div v-if="selectedIds.length > 0" class="batch-toolbar card-batch-toolbar">
          <div class="batch-info">
            <span class="batch-count">已选择 <strong>{{ selectedIds.length }}</strong> 项</span>
            <button class="batch-clear" @click="clearSelection">取消选择</button>
          </div>
          <div class="batch-actions">
            <button v-if="canOperate" class="batch-btn primary" @click="batchRunPipelines" :disabled="batchRunDisabled">
              <svg viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              批量发布
            </button>
            <button v-if="canOperate" class="batch-btn warning" @click="batchStopPipelines" :disabled="batchStopDisabled">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
              取消发布
            </button>
            <button v-if="canOperate" class="batch-btn danger" @click="batchDeletePipelines">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              批量删除
            </button>
          </div>
        </div>
      </transition>
      <div class="pipeline-grid">
        <div 
          v-for="pipeline in paginatedPipelines" 
          :key="pipeline.id"
          :class="['pipeline-card', { 'is-running': pipeline.status === 'running', 'is-selected': selectedIds.includes(pipeline.id) }]"
        >
          <!-- 卡片复选框 -->
          <label class="card-checkbox" @click.stop>
            <input type="checkbox" :checked="selectedIds.includes(pipeline.id)" @change="toggleSelect(pipeline.id)" />
            <span class="checkmark"></span>
          </label>
          <!-- 卡片头部 -->
        <div class="card-header">
          <div class="pipeline-info">
            <div class="pipeline-name-row">
              <span :class="['status-indicator', `status-${pipeline.status}`]"></span>
              <h3 class="pipeline-name" @click="viewPipeline(pipeline.id)">
                {{ pipeline.name }}
              </h3>
            </div>
            <p class="pipeline-desc">{{ pipeline.description || '暂无描述' }}</p>
          </div>
          <div class="pipeline-actions">
            <button 
              v-if="canOperate"
              class="action-btn run" 
              @click="handleRunPipeline(pipeline)"
              :disabled="pipeline.status === 'running'"
              title="运行"
            >
              <svg viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
            </button>
            <button class="action-btn more" @click="toggleMenu(pipeline.id)" title="更多">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="1"/>
                <circle cx="12" cy="5" r="1"/>
                <circle cx="12" cy="19" r="1"/>
              </svg>
            </button>
            <!-- 下拉菜单 -->
            <div v-if="activeMenu === pipeline.id" class="dropdown-menu">
              <button @click="viewPipeline(pipeline.id)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
                查看详情
              </button>
              <button v-if="canOperate" @click="editPipeline(pipeline.id)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
                编辑配置
              </button>
              <button v-if="canOperate && (pipeline.status === 'running' || pipeline.lastRunStatus === 'pending')" @click="handleStopPipeline(pipeline)" class="danger">
                <svg v-if="pipeline.lastRunStatus === 'pending'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="15" y1="9" x2="9" y2="15"/>
                  <line x1="9" y1="9" x2="15" y2="15"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="6" y="6" width="12" height="12" rx="2"/>
                </svg>
                {{ pipeline.lastRunStatus === 'pending' ? '取消构建' : '停止构建' }}
              </button>
              <button v-if="canOperate" @click="handleDeletePipeline(pipeline)" class="danger">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                删除流水线
              </button>
            </div>
          </div>
        </div>

        <!-- 运行状态 -->
        <div class="run-status">
          <div class="status-info">
            <div class="status-row">
              <span class="label">上次运行</span>
              <span :class="['run-status-tag', `status-${pipeline.lastRunStatus}`]">
                {{ runStatusText(pipeline.lastRunStatus) }}
              </span>
            </div>
            <div class="status-row">
              <span class="label">运行时间</span>
              <span class="value">{{ formatDate(pipeline.lastRunTime) }}</span>
            </div>
          </div>
          <div class="status-actions">
            <!-- 停止按钮：运行中或pending状态显示 -->
            <button
              v-if="canOperate && (pipeline.status === 'running' || pipeline.lastRunStatus === 'pending' || pipeline.lastRunStatus === 'running')"
              class="action-mini-btn btn-stop"
              @click.stop="handleStopPipeline(pipeline)"
              title="停止构建"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="6" y="6" width="12" height="12" rx="2"/>
              </svg>
              停止
            </button>
            <!-- 重新发布按钮：非运行状态显示 -->
            <button
              v-if="canOperate && (pipeline.status !== 'running' && pipeline.lastRunStatus !== 'running' && pipeline.lastRunStatus !== 'pending')"
              class="action-mini-btn btn-rerun"
              @click.stop="handleRunPipeline(pipeline)"
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

        <!-- 阶段进度条（运行中显示） -->
        <div v-if="pipeline.status === 'running'" class="stages-progress">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: '60%' }"></div>
          </div>
          <span class="progress-text">构建中...</span>
        </div>

        <!-- Git 信息 -->
        <div class="git-info">
          <div class="git-repo">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/>
            </svg>
            <span class="repo-url" :title="pipeline.gitRepo">{{ formatRepoUrl(pipeline.gitRepo) }}</span>
          </div>
          <div class="git-branch">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="6" y1="3" x2="6" y2="15"/>
              <circle cx="18" cy="6" r="3"/>
              <circle cx="6" cy="18" r="3"/>
              <path d="M18 9a9 9 0 0 1-9 9"/>
            </svg>
            <span>{{ pipeline.branch || 'main' }}</span>
          </div>
        </div>

        <!-- 卡片底部 -->
        <div class="card-footer">
          <button class="footer-btn" @click="viewPipeline(pipeline.id)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            详情
          </button>
          <button class="footer-btn" @click="viewHistory(pipeline.id)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            历史
          </button>
          <button class="footer-btn" @click="viewLogs(pipeline.id)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
            </svg>
            日志
          </button>
        </div>
      </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="filteredPipelines.length > 0" class="pagination">
      <span class="page-total">共 {{ filteredPipelines.length }} 条</span>
      <div class="page-controls">
        <button class="page-arrow" :disabled="currentPage === 1" @click="currentPage--">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
        </button>
        <span class="page-current">{{ currentPage }}</span>
        <button class="page-arrow" :disabled="currentPage >= localTotalPages" @click="currentPage++">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
      <div class="page-size-select">
        <select v-model.number="pageSize" @change="currentPage = 1">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
      </div>
      <div class="page-jump">
        <span>前往</span>
        <input type="number" v-model.number="jumpPage" min="1" :max="localTotalPages" @keyup.enter="goToPage" />
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import {
  getPipelines as fetchPipelines,
  runPipeline as triggerPipeline,
  stopPipeline as cancelPipeline,
  deletePipeline as removePipeline
} from '@/api/platform/pipeline'
import permissionStore from '@/stores/permission'

export default {
  name: 'Pipelines',
  setup() {
    const router = useRouter()
    const pipelines = ref([])
    const searchQuery = ref('')
    const statusFilter = ref('')
    const viewMode = ref('card')
    const currentPage = ref(1)
    const pageSize = ref(12)
    const total = ref(0)
    const loading = ref(false)
    const errorMsg = ref('')
    const activeMenu = ref(null)
    const jumpPage = ref(1)
    const selectedIds = ref([])

    // ===== 多选功能 =====
    const isAllSelected = computed(() => {
      if (paginatedPipelines.value.length === 0) return false
      return paginatedPipelines.value.every(p => selectedIds.value.includes(p.id))
    })
    const isIndeterminate = computed(() => {
      const selectedInPage = paginatedPipelines.value.filter(p => selectedIds.value.includes(p.id)).length
      return selectedInPage > 0 && selectedInPage < paginatedPipelines.value.length
    })
    const batchRunDisabled = computed(() => {
      return pipelines.value.filter(p => selectedIds.value.includes(p.id)).some(p => p.status === 'running')
    })
    const batchStopDisabled = computed(() => {
      return !pipelines.value.filter(p => selectedIds.value.includes(p.id)).some(p => p.status === 'running' || p.lastRunStatus === 'pending')
    })
    const toggleSelect = (id) => {
      const idx = selectedIds.value.indexOf(id)
      if (idx > -1) {
        selectedIds.value.splice(idx, 1)
      } else {
        selectedIds.value.push(id)
      }
    }
    const toggleSelectAll = () => {
      if (isAllSelected.value) {
        paginatedPipelines.value.forEach(p => {
          const idx = selectedIds.value.indexOf(p.id)
          if (idx > -1) selectedIds.value.splice(idx, 1)
        })
      } else {
        paginatedPipelines.value.forEach(p => {
          if (!selectedIds.value.includes(p.id)) selectedIds.value.push(p.id)
        })
      }
    }
    const clearSelection = () => { selectedIds.value = [] }

    // ===== 操作权限控制 =====
    // viewer 角色只能查看，不能执行任何修改操作
    const canOperate = computed(() => {
      if (permissionStore.state.isSuperAdmin) return true
      const roleTypes = permissionStore.roleTypes.value
      console.log('[Pipelines] 当前用户角色类型:', roleTypes)
      // viewer 角色无操作权限
      if (roleTypes.length === 1 && roleTypes.includes('viewer')) return false
      // 需要 developer 或更高权限才能操作流水线
      return roleTypes.some(r => ['super_admin', 'platform_admin', 'cluster_admin', 'developer', 'cicd_admin'].includes(r))
    })
    
    // TODO: 调试用 - 临时允许所有操作（上线前需删除）
    // const canOperate = ref(true)

    // 统计数据
    const runningCount = computed(() => 
      pipelines.value.filter(p => p.status === 'running').length
    )
    const successCount = computed(() => 
      pipelines.value.filter(p => p.lastRunStatus === 'success').length
    )
    const failedCount = computed(() => 
      pipelines.value.filter(p => p.lastRunStatus === 'failed').length
    )
    const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
    const localTotalPages = computed(() => Math.ceil(filteredPipelines.value.length / pageSize.value))
    
    // 生成分页按钮
    const displayPages = computed(() => {
      const pages = []
      const total = localTotalPages.value
      const current = currentPage.value
      
      if (total <= 7) {
        for (let i = 1; i <= total; i++) pages.push(i)
      } else {
        if (current <= 3) {
          pages.push(1, 2, 3, 4, '...', total)
        } else if (current >= total - 2) {
          pages.push(1, '...', total - 3, total - 2, total - 1, total)
        } else {
          pages.push(1, '...', current - 1, current, current + 1, '...', total)
        }
      }
      return pages
    })

    // 过滤后的流水线列表
    const filteredPipelines = computed(() => {
      let result = pipelines.value
      if (statusFilter.value) {
        result = result.filter(p => p.status === statusFilter.value)
      }
      return result
    })

    // 分页后的流水线列表
    const paginatedPipelines = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value
      return filteredPipelines.value.slice(start, start + pageSize.value)
    })

    // 加载流水线列表
    const loadPipelines = async () => {
      loading.value = true
      errorMsg.value = ''
      try {
        const response = await fetchPipelines({
          page: currentPage.value,
          page_size: pageSize.value,
          keyword: searchQuery.value || undefined,
          status: statusFilter.value || undefined
        })
        
        if (response.code === 0) {
          pipelines.value = (response.data?.list || []).map(item => ({
            id: item.id,
            name: item.name,
            description: item.description,
            status: item.status || 'idle',
            lastRunStatus: item.last_run_status || '',
            lastRunTime: item.last_run_time ? new Date(item.last_run_time * 1000).toISOString() : null,
            gitRepo: item.git_repo,
            branch: item.git_branch,
            jenkinsJob: item.jenkins_job
          }))
          total.value = response.data?.total || 0
        } else {
          throw new Error(response.msg || '获取流水线列表失败')
        }
      } catch (error) {
        console.error('加载流水线失败:', error)
        errorMsg.value = error.message || '获取流水线列表失败'
        pipelines.value = []
        total.value = 0
      } finally {
        loading.value = false
      }
    }

    // 运行流水线
    const handleRunPipeline = async (pipeline) => {
      if (!pipeline || !pipeline.id) {
        Message.error({ content: '流水线 ID 无效' })
        return
      }
      try {
        Message.info({ content: `正在启动流水线 "${pipeline.name}"...` })
        const response = await triggerPipeline(pipeline.id)
        if (response.code === 0) {
          Message.success({ content: '流水线启动成功' })
          loadPipelines()
        } else {
          throw new Error(response.msg || '启动失败')
        }
      } catch (error) {
        Message.error({ content: error.message || '启动流水线失败' })
      }
      activeMenu.value = null
    }

    // 取消/停止构建
    const handleStopPipeline = async (pipeline) => {
      if (!pipeline || !pipeline.id) {
        Message.error({ content: '流水线 ID 无效' })
        return
      }
      const isPending = pipeline.lastRunStatus === 'pending'
      const actionText = isPending ? '取消' : '停止'
      try {
        Message.info({ content: `正在${actionText}构建 "${pipeline.name}"...` })
        const response = await cancelPipeline(pipeline.id)
        if (response.code === 0) {
          Message.success({ content: `构建已${actionText}` })
          loadPipelines()
        } else {
          throw new Error(response.msg || `${actionText}失败`)
        }
      } catch (error) {
        Message.error({ content: error.message || `${actionText}构建失败` })
      }
      activeMenu.value = null
    }

    // 删除流水线
    const handleDeletePipeline = async (pipeline) => {
      if (!confirm(`确定要删除流水线 "${pipeline.name}" 吗？此操作不可恢复！`)) {
        return
      }
      try {
        const response = await removePipeline(pipeline.id)
        if (response.code === 0) {
          Message.success({ content: '删除成功' })
          loadPipelines()
        } else {
          throw new Error(response.msg || '删除失败')
        }
      } catch (error) {
        Message.error({ content: error.message || '删除流水线失败' })
      }
      activeMenu.value = null
    }

    // 批量运行流水线
    const batchRunPipelines = async () => {
      if (selectedIds.value.length === 0) return
      const toRun = pipelines.value.filter(p => selectedIds.value.includes(p.id) && p.status !== 'running')
      if (toRun.length === 0) {
        Message.warning({ content: '所选流水线均在运行中' })
        return
      }
      if (!confirm(`确定要批量发布 ${toRun.length} 条流水线吗？`)) return
      Message.info({ content: `正在启动 ${toRun.length} 条流水线...` })
      
      try {
        const response = await batchRunPipelines(selectedIds.value)
        if (response.code === 0) {
          const successCount = response.data?.success_count || 0
          const failCount = response.data?.fail_count || 0
          Message.success({ content: `成功发布 ${successCount} 条，失败 ${failCount} 条` })
        } else {
          throw new Error(response.msg || '批量发布失败')
        }
      } catch (error) {
        console.error('批量发布失败:', error)
        Message.error({ content: error.message || '批量发布失败' })
      }
      
      selectedIds.value = []
      loadPipelines()
    }

    // 批量停止流水线
    const batchStopPipelines = async () => {
      if (selectedIds.value.length === 0) return
      const toStop = pipelines.value.filter(p => selectedIds.value.includes(p.id) && (p.status === 'running' || p.lastRunStatus === 'pending'))
      if (toStop.length === 0) {
        Message.warning({ content: '所选流水线均未在运行' })
        return
      }
      if (!confirm(`确定要取消发布 ${toStop.length} 条流水线吗？`)) return
      Message.info({ content: `正在停止 ${toStop.length} 条流水线...` })
      
      try {
        const response = await batchStopPipelines(selectedIds.value)
        if (response.code === 0) {
          const successCount = response.data?.success_count || 0
          const failCount = response.data?.fail_count || 0
          Message.success({ content: `成功停止 ${successCount} 条，失败 ${failCount} 条` })
        } else {
          throw new Error(response.msg || '批量停止失败')
        }
      } catch (error) {
        console.error('批量停止失败:', error)
        Message.error({ content: error.message || '批量停止失败' })
      }
      
      selectedIds.value = []
      loadPipelines()
    }

    // 批量删除流水线
    const batchDeletePipelines = async () => {
      if (selectedIds.value.length === 0) return
      if (!confirm(`确定要删除 ${selectedIds.value.length} 条流水线吗？此操作不可恢复！`)) return
      let deleteSuccessCount = 0
      for (const id of selectedIds.value) {
        try {
          const res = await removePipeline(id)
          if (res.code === 0) deleteSuccessCount++
        } catch (e) { /* ignore */ }
      }
      Message.success({ content: `成功删除 ${deleteSuccessCount} 条流水线` })
      selectedIds.value = []
      loadPipelines()
    }

    // 菜单操作
    const toggleMenu = (id) => {
      activeMenu.value = activeMenu.value === id ? null : id
    }

    const closeMenu = (e) => {
      if (!e.target.closest('.dropdown-menu') && !e.target.closest('.action-btn.more')) {
        activeMenu.value = null
      }
    }

    // 导航
    const createPipeline = () => router.push('/cicd/pipelines/create')
    const viewPipeline = (id) => router.push(`/cicd/pipelines/${id}`)
    const editPipeline = (id) => router.push(`/cicd/pipelines/${id}/edit`)
    const viewHistory = (id) => router.push(`/cicd/pipelines/${id}?tab=history`)
    const viewLogs = (id) => router.push(`/cicd/pipelines/${id}?tab=logs`)

    // 页码跳转
    const goToPage = () => {
      const page = Math.max(1, Math.min(jumpPage.value, localTotalPages.value))
      currentPage.value = page
      jumpPage.value = page
    }

    // 格式化
    const formatDate = (dateString) => {
      if (!dateString) return '-'
      const date = new Date(dateString)
      const now = new Date()
      const diff = now - date
      
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
      if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
      
      return date.toLocaleDateString('zh-CN')
    }

    const formatRepoUrl = (url) => {
      if (!url) return '-'
      return url.replace(/^https?:\/\//, '').replace(/\.git$/, '')
    }

    const runStatusText = (status) => {
      const map = {
        success: '成功',
        failed: '失败',
        running: '运行中',
        pending: '等待中',
        aborted: '已中止',
        '': '未运行'
      }
      return map[status] || status
    }

    // 监听
    watch([searchQuery, statusFilter], () => {
      currentPage.value = 1
      loadPipelines()
    })

    watch(currentPage, loadPipelines)

    onMounted(() => {
      loadPipelines()
      document.addEventListener('click', closeMenu)
    })

    onBeforeUnmount(() => {
      document.removeEventListener('click', closeMenu)
    })

    return {
      pipelines,
      searchQuery,
      statusFilter,
      viewMode,
      currentPage,
      pageSize,
      total,
      totalPages,
      localTotalPages,
      displayPages,
      loading,
      errorMsg,
      activeMenu,
      jumpPage,
      selectedIds,
      isAllSelected,
      isIndeterminate,
      batchRunDisabled,
      batchStopDisabled,
      runningCount,
      successCount,
      failedCount,
      filteredPipelines,
      paginatedPipelines,
      canOperate,
      loadPipelines,
      handleRunPipeline,
      handleStopPipeline,
      handleDeletePipeline,
      batchRunPipelines,
      batchStopPipelines,
      batchDeletePipelines,
      toggleSelect,
      toggleSelectAll,
      clearSelection,
      toggleMenu,
      createPipeline,
      viewPipeline,
      editPipeline,
      viewHistory,
      viewLogs,
      goToPage,
      formatDate,
      formatRepoUrl,
      runStatusText
    }
  }
}
</script>

<style scoped>
.pipeline-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.title-icon {
  width: 28px;
  height: 28px;
  color: #4299e1;
}

.page-desc {
  color: #718096;
  font-size: 14px;
  margin: 0;
}

.header-right {
  display: flex;
  gap: 12px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon svg {
  width: 24px;
  height: 24px;
}

.stat-icon.total { background: #ebf8ff; color: #3182ce; }
.stat-icon.running { background: #fef3c7; color: #d97706; }
.stat-icon.success { background: #d1fae5; color: #059669; }
.stat-icon.failed { background: #fee2e2; color: #dc2626; }

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
  font-size: 13px;
  color: #718096;
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.view-switch {
  display: flex;
  gap: 2px;
  padding: 4px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.view-btn {
  padding: 8px 10px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.view-btn svg {
  width: 18px;
  height: 18px;
}

.view-btn:hover {
  color: #64748b;
}

.view-btn.active {
  background: #4299e1;
  color: white;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: #a0aec0;
}

.search-input {
  width: 100%;
  padding: 12px 14px 12px 44px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  background: white;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.15);
}

.filter-tabs {
  display: flex;
  gap: 8px;
  background: white;
  padding: 4px;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.filter-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #718096;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover {
  background: #f7fafc;
}

.filter-tab.active {
  background: #4299e1;
  color: white;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.running {
  background: #d97706;
  animation: pulse 1.5s infinite;
}

.status-dot.idle {
  background: #3182ce;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 错误提示 */
.error-alert {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #fee2e2;
  border: 1px solid #fca5a5;
  border-radius: 10px;
  color: #dc2626;
  margin-bottom: 20px;
}

.alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.alert-close {
  margin-left: auto;
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #dc2626;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
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

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  background: white;
  border-radius: 16px;
  text-align: center;
}

.empty-icon {
  width: 80px;
  height: 80px;
  margin-bottom: 20px;
  color: #cbd5e0;
}

.empty-icon svg {
  width: 100%;
  height: 100%;
}

.empty-state h3 {
  font-size: 18px;
  color: #4a5568;
  margin: 0 0 8px 0;
}

.empty-state p {
  color: #a0aec0;
  margin: 0 0 24px 0;
}

/* 表格视图 */
.table-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* 批量操作工具栏 */
.batch-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px;
  background: linear-gradient(135deg, #1e40af 0%, #2563eb 100%);
  border-radius: 10px;
  margin-bottom: 16px;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-size: 14px;
  color: #ffffff;
}

.batch-count strong {
  font-weight: 700;
  color: #fbbf24;
  font-size: 16px;
}

.batch-clear {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #ffffff;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 4px;
  transition: all 0.15s;
}

.batch-clear:hover {
  background: rgba(255, 255, 255, 0.25);
  border-color: rgba(255, 255, 255, 0.5);
}

.batch-actions {
  display: flex;
  gap: 10px;
}

.batch-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 18px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.batch-btn svg {
  width: 14px;
  height: 14px;
}

.batch-btn.primary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 3px 8px rgba(16, 185, 129, 0.35);
}

.batch-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.45);
}

.batch-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.batch-btn.warning {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  box-shadow: 0 3px 8px rgba(245, 158, 11, 0.35);
}

.batch-btn.warning:hover:not(:disabled) {
  background: linear-gradient(135deg, #d97706 0%, #b45309 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.45);
}

.batch-btn.warning:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.batch-btn.danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  box-shadow: 0 3px 8px rgba(239, 68, 68, 0.35);
}

.batch-btn.danger:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.45);
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.2s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 复选框 */
.checkbox-col {
  width: 48px;
  text-align: center;
}

.checkbox-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  cursor: pointer;
}

.checkbox-wrapper input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.checkmark {
  width: 18px;
  height: 18px;
  border: 2px solid #cbd5e1;
  border-radius: 4px;
  background: white;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.checkbox-wrapper:hover .checkmark {
  border-color: #4299e1;
}

.checkbox-wrapper input:checked + .checkmark {
  background: #4299e1;
  border-color: #4299e1;
}

.checkbox-wrapper input:checked + .checkmark::after {
  content: '';
  width: 5px;
  height: 9px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  margin-bottom: 2px;
}

.checkbox-wrapper input:indeterminate + .checkmark {
  background: #4299e1;
  border-color: #4299e1;
}

.checkbox-wrapper input:indeterminate + .checkmark::after {
  content: '';
  width: 10px;
  height: 2px;
  background: white;
}

.row-selected {
  background: #eff6ff !important;
}

.pipeline-table {
  width: 100%;
  border-collapse: collapse;
}

.pipeline-table th {
  background: #f8fafc;
  padding: 14px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #e2e8f0;
}

.pipeline-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  font-size: 14px;
  color: #334155;
}

.pipeline-table tbody tr:hover {
  background: #f8fafc;
}

.table-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pipeline-link {
  color: #3182ce;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.15s;
}

.pipeline-link:hover {
  color: #2563eb;
  text-decoration: underline;
}

.table-desc {
  color: #64748b;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.table-time {
  color: #64748b;
  font-size: 13px;
}

.table-repo {
  color: #64748b;
  font-size: 13px;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.table-branch {
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #475569;
  font-family: monospace;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.status-tag.status-running {
  background: #fef3c7;
  color: #d97706;
}

.status-tag.status-idle {
  background: #dbeafe;
  color: #2563eb;
}

.table-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.action-btn-sm {
  padding: 4px 10px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  background: white;
  font-size: 12px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn-sm:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.action-btn-sm.run {
  background: #d1fae5;
  border-color: #a7f3d0;
  color: #059669;
}

.action-btn-sm.run:hover {
  background: #a7f3d0;
}

.action-btn-sm.stop {
  background: #fee2e2;
  border-color: #fecaca;
  color: #dc2626;
}

.action-btn-sm.stop:hover {
  background: #fecaca;
}

.action-btn-sm.danger {
  color: #dc2626;
}

.action-btn-sm.danger:hover {
  background: #fee2e2;
  border-color: #fecaca;
}

/* 流水线网格包装器 */
.pipeline-grid-wrapper {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-batch-toolbar {
  border-radius: 12px;
  margin-bottom: 0;
}

/* 流水线网格 */
.pipeline-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}

/* 流水线卡片 */
.pipeline-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
  border: 1px solid transparent;
  position: relative;
}

.pipeline-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
  border-color: #e2e8f0;
}

.pipeline-card.is-running {
  border-color: #fbbf24;
  box-shadow: 0 0 0 2px rgba(251, 191, 36, 0.2);
}

.pipeline-card.is-selected {
  border-color: #4299e1;
  box-shadow: 0 0 0 2px rgba(66, 153, 225, 0.3);
  background: #f0f9ff;
}

/* 卡片复选框 */
.card-checkbox {
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 10;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.card-checkbox input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.card-checkbox .checkmark {
  width: 20px;
  height: 20px;
  border: 2px solid #cbd5e1;
  border-radius: 6px;
  background: white;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.pipeline-card:hover .card-checkbox .checkmark {
  border-color: #4299e1;
}

.card-checkbox input:checked + .checkmark {
  background: #4299e1;
  border-color: #4299e1;
}

.card-checkbox input:checked + .checkmark::after {
  content: '';
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  margin-bottom: 2px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 20px 20px 20px 48px;
  border-bottom: 1px solid #f1f5f9;
}

.pipeline-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-indicator.status-idle { background: #3182ce; }
.status-indicator.status-running { 
  background: #d97706; 
  animation: pulse 1.5s infinite;
}
.status-indicator.status-disabled { background: #a0aec0; }
.status-indicator.status-error { background: #dc2626; }

.pipeline-name {
  font-size: 16px;
  font-weight: 600;
  color: #1a202c;
  margin: 0;
  cursor: pointer;
  transition: color 0.2s;
}

.pipeline-name:hover {
  color: #4299e1;
}

.pipeline-desc {
  font-size: 13px;
  color: #718096;
  margin: 6px 0 0 20px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pipeline-actions {
  display: flex;
  gap: 8px;
  position: relative;
}

.action-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.action-btn svg {
  width: 16px;
  height: 16px;
}

.action-btn.run {
  background: #d1fae5;
  color: #059669;
}

.action-btn.run:hover:not(:disabled) {
  background: #059669;
  color: white;
}

.action-btn.run:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.more {
  background: #f1f5f9;
  color: #64748b;
}

.action-btn.more:hover {
  background: #e2e8f0;
}

/* 下拉菜单 */
.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  z-index: 100;
  overflow: hidden;
}

.dropdown-menu button {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  background: none;
  font-size: 14px;
  color: #4a5568;
  cursor: pointer;
  transition: background 0.2s;
}

.dropdown-menu button:hover {
  background: #f7fafc;
}

.dropdown-menu button.danger {
  color: #dc2626;
}

.dropdown-menu button.danger:hover {
  background: #fee2e2;
}

.dropdown-menu svg {
  width: 16px;
  height: 16px;
}

/* 运行状态 */
.run-status {
  padding: 12px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  background: #f8fafc;
}

.status-info {
  display: flex;
  gap: 24px;
}

.status-actions {
  display: flex;
  gap: 8px;
}

.action-mini-btn {
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

.action-mini-btn svg {
  width: 14px;
  height: 14px;
}

.action-mini-btn.btn-stop {
  background: #fee2e2;
  color: #dc2626;
}

.action-mini-btn.btn-stop:hover {
  background: #fecaca;
}

.action-mini-btn.btn-rerun {
  background: #dbeafe;
  color: #2563eb;
}

.action-mini-btn.btn-rerun:hover {
  background: #bfdbfe;
}

.status-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-row .label {
  font-size: 12px;
  color: #94a3b8;
}

.status-row .value {
  font-size: 13px;
  color: #475569;
  font-weight: 500;
}

.run-status-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}

.run-status-tag.status-success { background: #d1fae5; color: #059669; }
.run-status-tag.status-failed { background: #fee2e2; color: #dc2626; }
.run-status-tag.status-running { background: #fef3c7; color: #d97706; }
.run-status-tag.status-pending { background: #f1f5f9; color: #64748b; }
.run-status-tag.status-aborted { background: #f1f5f9; color: #64748b; }
.run-status-tag.status- { background: #f1f5f9; color: #94a3b8; }

/* 阶段进度 */
.stages-progress {
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(90deg, #fef3c7, #fde68a);
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #d97706;
  border-radius: 3px;
  animation: progress-pulse 1.5s ease-in-out infinite;
}

@keyframes progress-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.progress-text {
  font-size: 12px;
  font-weight: 600;
  color: #92400e;
}

/* Git 信息 */
.git-info {
  padding: 16px 20px;
  display: flex;
  gap: 20px;
  border-top: 1px solid #f1f5f9;
}

.git-repo, .git-branch {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #64748b;
}

.git-repo svg, .git-branch svg {
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

.repo-url {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 卡片底部 */
.card-footer {
  display: flex;
  border-top: 1px solid #f1f5f9;
}

.footer-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 14px;
  border: none;
  background: none;
  font-size: 13px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.footer-btn:hover {
  background: #f8fafc;
  color: #4299e1;
}

.footer-btn:not(:last-child) {
  border-right: 1px solid #f1f5f9;
}

.footer-btn svg {
  width: 16px;
  height: 16px;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  padding: 12px 0;
}

.page-total {
  font-size: 14px;
  color: #64748b;
}

.page-controls {
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-arrow {
  width: 32px;
  height: 32px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  transition: all 0.15s;
}

.page-arrow:hover:not(:disabled) {
  border-color: #4299e1;
  color: #4299e1;
}

.page-arrow:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-arrow svg {
  width: 14px;
  height: 14px;
}

.page-current {
  min-width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #4299e1;
  color: white;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
}

.page-size-select select {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  font-size: 14px;
  color: #334155;
  cursor: pointer;
  outline: none;
}

.page-size-select select:focus {
  border-color: #4299e1;
}

.page-jump {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #64748b;
}

.page-jump input {
  width: 50px;
  padding: 6px 8px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  text-align: center;
  outline: none;
}

.page-jump input:focus {
  border-color: #4299e1;
}

.page-jump input::-webkit-inner-spin-button,
.page-jump input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
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

.btn-icon {
  width: 18px;
  height: 18px;
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  border-color: #3182ce;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.4);
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
  transform: none !important;
}

/* 响应式 */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-right {
    width: 100%;
  }
  
  .header-right .btn {
    flex: 1;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .search-wrapper {
    max-width: none;
    width: 100%;
  }
  
  .pipeline-grid {
    grid-template-columns: 1fr;
  }
}
</style>

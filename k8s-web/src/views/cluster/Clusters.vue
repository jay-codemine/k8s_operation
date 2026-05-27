<template>
  <div class="cluster-view">
    <!-- 页面头部：大厂 Dashboard 风格 -->
    <div class="page-header">
      <div class="page-title-section">
        <div class="title-icon">
          <svg width="28" height="28" viewBox="0 0 32 32" fill="none">
            <circle cx="16" cy="16" r="14" stroke="currentColor" stroke-width="2"/>
            <path d="M16 6L16 16M16 16L24 20M16 16L8 20M16 16L16 26" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            <circle cx="16" cy="6" r="2" fill="currentColor"/>
            <circle cx="24" cy="20" r="2" fill="currentColor"/>
            <circle cx="8" cy="20" r="2" fill="currentColor"/>
            <circle cx="16" cy="26" r="2" fill="currentColor"/>
          </svg>
        </div>
        <div>
          <h1>集群管理</h1>
          <p class="page-subtitle">统一管理和监控 Kubernetes 集群的健康状态与生命周期</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn btn-ghost" :disabled="loading" @click="fetchList">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M1 4v6h6M23 20v-6h-6"/>
            <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>
          </svg>
          刷新
        </button>
        <button v-if="canOperate" class="btn btn-primary" @click="openCreate">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          接入集群
        </button>
      </div>
    </div>

    <!-- 统计概览卡片 -->
    <div class="stats-overview">
      <div class="stat-card">
        <div class="stat-icon total">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="2" width="20" height="20" rx="3"/>
            <path d="M12 6v12M6 12h12"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ clusters.length }}</div>
          <div class="stat-label">集群总数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon healthy">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ clusterStats.healthy }}</div>
          <div class="stat-label">运行正常</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon unhealthy">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ clusterStats.unhealthy }}</div>
          <div class="stat-label">连接异常</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon pending">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ clusterStats.pending }}</div>
          <div class="stat-label">待检测</div>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="search-box">
          <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            type="text"
            v-model="searchQuery"
            placeholder="搜索集群名称..."
            @keyup.enter="fetchList"
          />
        </div>
        <div class="filter-group">
          <button 
            v-for="f in filters" 
            :key="f.key"
            class="filter-btn" 
            :class="{ active: statusFilter === f.key }"
            @click="setFilter(f.key)"
          >
            <span class="filter-dot" :class="f.key"></span>
            {{ f.label }}
          </button>
        </div>
      </div>
      <div class="toolbar-right">
        <button 
          v-if="canOperate && selectedIds.length > 0" 
          class="btn btn-danger-solid" 
          :disabled="loading"
          @click="onBatchDelete"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
          批量删除 ({{ selectedIds.length }})
        </button>
        <div class="view-switch">
          <button 
            class="view-btn" 
            :class="{ active: viewMode === 'table' }" 
            @click="viewMode = 'table'"
            title="列表视图"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
          <button 
            class="view-btn" 
            :class="{ active: viewMode === 'card' }" 
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <div class="table-scroll">
        <table class="resource-table">
          <thead>
          <tr>
            <th v-if="canOperate" style="width: 48px;">
              <input 
                type="checkbox" 
                :checked="isAllSelected" 
                :indeterminate="isIndeterminate"
                @change="toggleAll"
                class="row-checkbox"
              />
            </th>
            <th style="width: 70px;">ID</th>
            <th>集群名称</th>
            <th style="width: 150px;">K8s 版本</th>
            <th style="width: 120px;">状态</th>
            <th style="width: 180px;">最近检测</th>
            <th style="width: 240px;">操作</th>
          </tr>
          </thead>

          <tbody>
          <tr v-for="c in paginatedClusters" :key="c.id" :class="{ 'row-selected': selectedIds.includes(c.id) }" @click="enterCluster(c)" style="cursor:pointer;">
            <td v-if="canOperate" @click.stop>
              <input 
                type="checkbox" 
                :checked="selectedIds.includes(c.id)" 
                @change="toggleSelect(c.id)"
                class="row-checkbox"
              />
            </td>
            <td>
              <span class="id-badge">{{ c.id }}</span>
            </td>

            <td>
              <div class="cluster-name-cell">
                <div class="cluster-avatar" :class="statusClass(c.status)">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <path d="M12 6v6l4 2"/>
                  </svg>
                </div>
                <div class="cluster-name-info">
                  <a href="javascript:void(0)" class="cluster-link" @click.prevent="enterCluster(c)">
                    {{ c.cluster_name }}
                  </a>
                  <div v-if="c.last_error" class="cluster-error-hint" :title="c.last_error">
                    {{ c.last_error }}
                  </div>
                </div>
              </div>
            </td>

            <td>
              <span class="version-tag" v-if="c.cluster_version">{{ c.cluster_version }}</span>
              <span v-else class="text-muted">-</span>
            </td>

            <td>
              <span class="status-badge" :class="statusClass(c.status)">
                <span class="status-dot"></span>
                {{ statusText(c.status) }}
              </span>
            </td>

            <td>
              <span class="time-text">{{ formatCheckAt(c.last_check_at) }}</span>
            </td>

            <td @click.stop>
              <div class="action-group">
                <button v-if="canOperate" class="action-btn" :disabled="testingId === c.id || loading"
                        @click="openEdit(c)" title="编辑集群">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                  编辑
                </button>

                <button
                  class="action-btn action-btn-primary"
                  :disabled="testingId === c.id || loading"
                  @click="testCluster(c)"
                  title="执行健康检查"
                >
                  <svg v-if="testingId !== c.id" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
                  </svg>
                  <span v-else class="spin-icon">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                    </svg>
                  </span>
                  {{ testingId === c.id ? '检测中' : '检查' }}
                </button>

                <button
                  v-if="canOperate"
                  class="action-btn action-btn-danger"
                  :disabled="testingId === c.id || loading"
                  @click="onDelete(c)"
                  title="删除集群"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
          </tbody>
        </table>
      </div>

      <div v-if="filteredClusters.length === 0" class="empty-state">
        <div class="empty-illustration">
          <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
            <circle cx="60" cy="60" r="50" stroke="#e2e8f0" stroke-width="2" stroke-dasharray="6 4"/>
            <circle cx="60" cy="60" r="20" stroke="#cbd5e1" stroke-width="2"/>
            <path d="M60 40v20l12 6" stroke="#94a3b8" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="empty-title">
          {{ searchQuery ? '没有找到匹配的集群' : '还没有接入集群' }}
        </div>
        <div class="empty-desc">
          {{ searchQuery ? '尝试调整搜索关键词或筛选条件' : '点击「接入集群」开始管理您的 Kubernetes 集群' }}
        </div>
        <button v-if="!searchQuery && canOperate" class="btn btn-primary" @click="openCreate" style="margin-top: 16px;">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          接入集群
        </button>
      </div>

      <Pagination
        v-if="filteredClusters.length > 0"
        v-model:currentPage="currentPage"
        :totalItems="filteredClusters.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="filteredClusters.length === 0" class="empty-state">
        <div class="empty-illustration">
          <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
            <circle cx="60" cy="60" r="50" stroke="#e2e8f0" stroke-width="2" stroke-dasharray="6 4"/>
            <circle cx="60" cy="60" r="20" stroke="#cbd5e1" stroke-width="2"/>
            <path d="M60 40v20l12 6" stroke="#94a3b8" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="empty-title">{{ searchQuery ? '没有找到匹配的集群' : '还没有接入集群' }}</div>
        <div class="empty-desc">{{ searchQuery ? '尝试调整搜索关键词' : '点击「接入集群」开始管理' }}</div>
      </div>
      
      <div class="cards-grid">
        <div v-for="c in paginatedClusters" :key="c.id" class="cluster-card" :class="statusClass(c.status)">
          <!-- 卡片状态条 -->
          <div class="card-status-bar" :class="statusClass(c.status)"></div>
          
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <div class="card-avatar" :class="statusClass(c.status)">
                <svg width="22" height="22" viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="16" cy="16" r="12"/>
                  <path d="M16 8v8l5 3"/>
                </svg>
              </div>
              <div class="card-title-info">
                <h3 class="card-title">
                  <a href="javascript:void(0)" class="cluster-link" @click.prevent="enterCluster(c)">
                    {{ c.cluster_name }}
                  </a>
                </h3>
                <span class="card-id">ID: {{ c.id }}</span>
              </div>
              <span class="status-badge" :class="statusClass(c.status)">
                <span class="status-dot"></span>
                {{ statusText(c.status) }}
              </span>
            </div>
          </div>

          <!-- 卡片主体 -->
          <div class="card-body">
            <div class="card-meta-grid">
              <div class="card-meta-item">
                <div class="meta-icon">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                    <line x1="7" y1="7" x2="7.01" y2="7"/>
                  </svg>
                </div>
                <div>
                  <div class="meta-label">K8s 版本</div>
                  <div class="meta-value">{{ c.cluster_version || '-' }}</div>
                </div>
              </div>
              <div class="card-meta-item">
                <div class="meta-icon">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                </div>
                <div>
                  <div class="meta-label">最近检查</div>
                  <div class="meta-value">{{ formatCheckAt(c.last_check_at) }}</div>
                </div>
              </div>
            </div>

            <!-- 错误信息 -->
            <div v-if="c.last_error" class="card-error">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              <span>{{ c.last_error }}</span>
            </div>
          </div>

          <!-- 卡片底部 -->
          <div class="card-footer">
            <button 
              class="card-btn card-btn-enter" 
              :disabled="testingId === c.id || loading"
              @click="enterCluster(c)"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/>
                <polyline points="10 17 15 12 10 7"/>
                <line x1="15" y1="12" x2="3" y2="12"/>
              </svg>
              进入管理
            </button>
            <button 
              class="card-btn"
              :disabled="testingId === c.id || loading"
              @click="testCluster(c)"
            >
              {{ testingId === c.id ? '检测中...' : '健康检查' }}
            </button>
            <div class="card-btn-more" v-if="canOperate">
              <button class="card-btn card-btn-icon" @click="openEdit(c)" title="编辑">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
              </button>
              <button class="card-btn card-btn-icon card-btn-danger" @click="onDelete(c)" title="删除">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <Pagination
        v-if="filteredClusters.length > 0"
        v-model:currentPage="currentPage"
        :totalItems="filteredClusters.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 创建/编辑 弹窗 -->
    <div class="modal" v-if="showFormModal">
      <div class="modal-backdrop" @click="closeForm"></div>

      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ formMode === 'create' ? '创建集群' : '编辑集群' }}</h2>
          <button class="close-btn" @click="closeForm">&times;</button>
        </div>

        <div class="modal-body">
          <form class="form" @submit.prevent="submitForm">
            <div class="topbar" v-if="formMode === 'edit'">
              <div class="chip">ClusterID: {{ form.id }}</div>
              <div class="muted">编辑模式：不填写 kubeconfig 不会覆盖</div>
            </div>

            <div class="card">
              <div class="card-title">基本信息</div>

              <div class="grid">
                <div class="field" v-if="formMode === 'edit'">
                  <label>ID</label>
                  <input type="number" v-model="form.id" disabled/>
                </div>

                <div class="field">
                  <label>集群名称 <span class="required">*</span></label>
                  <input
                    type="text"
                    v-model="form.cluster_name"
                    placeholder="例如：测试环境 k8s 集群"
                    required
                  />
                </div>

                <div class="field">
                  <label>K8s 版本 <span class="required">*</span></label>
                  <input
                    type="text"
                    v-model="form.cluster_version"
                    placeholder="如 v1.28.3"
                    required
                  />
                </div>
              </div>
            </div>

            <div class="card">
              <div class="card-title">
                kubeconfig（高级）
                <span class="hint">创建模式必填；编辑模式可选，不填则不更新</span>
              </div>

              <div class="upload-row">
                <input
                  class="file"
                  type="file"
                  accept=".yaml,.yml,.conf,.json,.txt"
                  @change="onKubeconfigFileChange"
                />
                <button type="button" class="btn small ghost" @click="clearKubeconfigText">
                  清空文本
                </button>
              </div>

              <div class="alert">
                <span class="alert-icon">⚠️</span>
                <div class="alert-text">
                  上传 kubeconfig 文件后会自动填入下方文本框；编辑时不填 kubeconfig 则不会更新。
                </div>
              </div>

              <textarea
                v-model="form.kube_config"
                class="codebox"
                rows="10"
                placeholder="粘贴 kubeconfig 原文（YAML/JSON），或使用上方上传文件"
                :required="formMode === 'create'"
              ></textarea>
            </div>

            <div class="footer">
              <button type="button" class="btn ghost" @click="closeForm">取消</button>
              <button type="submit" class="btn primary" :disabled="loading">
                {{ loading ? '提交中...' : '提交保存' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue'
import {Message} from '@arco-design/web-vue'
import {useRouter} from 'vue-router'
import Pagination from '@/components/Pagination.vue'
import {useClusterStore} from '@/stores/cluster'
import permissionStore from '@/stores/permission'
import http from '@/api/http'
import {useConfirmDialog} from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

import {
  createCluster,
  deleteCluster,
  batchDeleteCluster,
  getClusterList,
  initCluster,
  updateCluster,
} from '@/api/cluster'

const router = useRouter()
const clusterStore = useClusterStore()

// ===== 操作权限控制 =====
// viewer 角色只能查看，不能执行任何修改操作
const canOperate = computed(() => {
  if (permissionStore.state.isSuperAdmin) return true
  const roleTypes = permissionStore.roleTypes.value
  // viewer 角色无操作权限
  if (roleTypes.length === 1 && roleTypes.includes('viewer')) return false
  // 需要 cluster_admin 或更高权限才能操作集群
  return roleTypes.some(r => ['super_admin', 'platform_admin', 'cluster_admin'].includes(r))
})

// ===== 列表数据 =====
const clusters = ref([])

// 视图模式：table（表格） 或 card（卡片）
const viewMode = ref('table')

// 集群统计
const clusterStats = computed(() => {
  const all = clusters.value
  return {
    healthy: all.filter(c => Number(c.status) === 0).length,
    unhealthy: all.filter(c => Number(c.status) === 1).length,
    pending: all.filter(c => Number(c.status) === 2).length,
  }
})

// ===== 批量选择 =====
const selectedIds = ref([])
const isAllSelected = computed(() => {
  if (paginatedClusters.value.length === 0) return false
  return paginatedClusters.value.every(c => selectedIds.value.includes(c.id))
})
const isIndeterminate = computed(() => {
  if (paginatedClusters.value.length === 0) return false
  const selected = paginatedClusters.value.filter(c => selectedIds.value.includes(c.id))
  return selected.length > 0 && selected.length < paginatedClusters.value.length
})

const toggleAll = (e) => {
  if (e.target.checked) {
    const currentPageIds = paginatedClusters.value.map(c => c.id)
    selectedIds.value = [...new Set([...selectedIds.value, ...currentPageIds])]
  } else {
    const currentPageIds = paginatedClusters.value.map(c => c.id)
    selectedIds.value = selectedIds.value.filter(id => !currentPageIds.includes(id))
  }
}

const toggleSelect = (id) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx === -1) {
    selectedIds.value.push(id)
  } else {
    selectedIds.value.splice(idx, 1)
  }
}

// ===== UI =====
const searchQuery = ref('')
const statusFilter = ref('all') // all | ok | bad | pending
const currentPage = ref(1)
const itemsPerPage = ref(10)
const loading = ref(false)
const testingId = ref(null)

// 筛选按钮配置
const filters = [
  { key: 'all', label: '全部' },
  { key: 'ok', label: '正常' },
  { key: 'bad', label: '异常' },
  { key: 'pending', label: '待检测' }
]

// ===== 弹窗表单 =====
const showFormModal = ref(false)
const formMode = ref('create') // create | edit
const form = ref({
  id: 0,
  cluster_name: '',
  cluster_version: '',
  kube_config: '',
})

onMounted(() => {
  clusterStore.hydrate?.()
  fetchList()
})

const setFilter = (v) => {
  statusFilter.value = v
  currentPage.value = 1
}

const enterCluster = async (c) => {
  clusterStore.setCurrent(c)
  router.push(`/c/${Number(c.id)}/nodes`)
}

/**
 * 你后端三态：0=OK 1=Bad 2=Pending
 * 如果你后端枚举不一样，把这里改一下就行
 */
const statusText = (s) => {
  const n = Number(s)
  if (n === 0) return '正常'
  if (n === 1) return '异常'
  if (n === 2) return '待检测'
  return '未知'
}
const statusClass = (s) => {
  const n = Number(s)
  if (n === 0) return 'connected'
  if (n === 1) return 'disconnected'
  if (n === 2) return 'pending'
  return 'unknown'
}

const pickMsg = (body, fallback = '') => {
  if (!body) return fallback
  // 优先取 details 中的详细错误信息
  if (Array.isArray(body?.details) && body.details.length > 0) return body.details.join('；')
  if (body?.data?.message) return body.data.message
  if (body?.msg) return body.msg
  if (body?.message) return body.message
  return fallback
}
const unwrapErrorBody = (e) => {
  if (e?.response?.data) return e.response.data
  if (e?.code || e?.msg || e?.message) return e
  return null
}
const isOk = (body) => Number(body?.code) === 0

// ✅ 拉取列表：只信后端（status/last_check_at/last_error 都由后端写库）
const fetchList = async () => {
  loading.value = true
  try {
    const body = await getClusterList({
      cluster_name: searchQuery.value || '',
      page: 1,
      limit: 1000,
    })

    const list = body?.data?.list || body?.data?.items || body?.list || body?.items || []
    clusters.value = Array.isArray(list) ? list.map((x) => ({
      ...x,
      id: Number(x?.id),
      status: Number(x?.status),
      last_check_at: Number(x?.last_check_at || 0),
      last_error: String(x?.last_error || ''),
    })) : []

    // 如果当前页超范围，拉回最后一页/第一页
    const totalPages = Math.max(1, Math.ceil(filteredClusters.value.length / itemsPerPage.value))
    if (currentPage.value > totalPages) currentPage.value = totalPages
    if (currentPage.value < 1) currentPage.value = 1
  } catch (e) {
    console.error(e)
    clusters.value = []
    Message.error({content: '拉取集群列表失败', duration: 2200})
  } finally {
    loading.value = false
  }
}

const filteredClusters = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()

  return clusters.value.filter((c) => {
    // 权限过滤：只显示用户有权限访问的集群
    const hasPermission = permissionStore.state.isSuperAdmin ||
      permissionStore.state.accessibleClusterIds.includes(c.id)
    if (!hasPermission) return false

    const hitName = !q || String(c.cluster_name || '').toLowerCase().includes(q)

    const s = Number(c.status)
    const hitStatus =
      statusFilter.value === 'all' ||
      (statusFilter.value === 'ok' && s === 0) ||
      (statusFilter.value === 'bad' && s === 1) ||
      (statusFilter.value === 'pending' && s === 2)

    return hitName && hitStatus
  })
})

const paginatedClusters = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredClusters.value.slice(start, end)
})

const openCreate = () => {
  formMode.value = 'create'
  form.value = {id: 0, cluster_name: '', cluster_version: '', kube_config: ''}
  showFormModal.value = true
}

const openEdit = (c) => {
  formMode.value = 'edit'
  form.value = {
    id: Number(c.id),
    cluster_name: c.cluster_name,
    cluster_version: c.cluster_version,
    kube_config: '',
  }
  showFormModal.value = true
}

const closeForm = () => {
  showFormModal.value = false
}

// kubeconfig 文件上传
const onKubeconfigFileChange = (evt) => {
  const file = evt?.target?.files?.[0]
  if (!file) return

  if (file.size > 1024 * 1024) {
    Message.warning({content: '文件过大（>1MB），请确认是否为 kubeconfig 文件', duration: 2200})
    evt.target.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    form.value.kube_config = String(reader.result || '')
    Message.success({content: 'kubeconfig 已读取到文本框', duration: 1600})
  }
  reader.onerror = () => {
    Message.error({content: '读取文件失败，请重试', duration: 2200})
  }
  reader.readAsText(file)
  evt.target.value = ''
}

const clearKubeconfigText = () => {
  form.value.kube_config = ''
}

// ✅ 创建/更新（更新 kubeconfig 后，后端应该把 status 置为 Pending，等 init 再写库）
const submitForm = async () => {
  loading.value = true
  try {
    if (formMode.value === 'create') {
      const body = await createCluster({
        cluster_name: form.value.cluster_name,
        cluster_version: form.value.cluster_version,
        kube_config: form.value.kube_config,
      })
      Message.success({content: body?.msg || '创建成功', duration: 1800})
      closeForm()
      await fetchList()
      return
    }

    // 编辑模式 - 检查是否更新 kubeconfig
    const hasKubeconfigUpdate = form.value.kube_config && form.value.kube_config.trim()
    
    // 二次确认 - 更新 kubeconfig 是高危操作
    if (hasKubeconfigUpdate) {
      const ok = await showConfirm({
        title: '确认更新集群配置',
        content: '此操作将替换现有的集群连接凭证，可能导致连接失败。',
        type: 'warning',
        details: [
          { label: '集群名称', value: form.value.cluster_name },
          { label: '集群 ID', value: String(form.value.id), mono: true },
          { label: 'K8s 版本', value: form.value.cluster_version },
        ],
        tip: '更新 kubeconfig 后建议立即执行健康检查，确保集群连接正常。',
        confirmText: '确认更新',
        cancelText: '取消',
      })
      if (!ok) {
        loading.value = false
        return
      }
    }

    const payload = {
      id: form.value.id,
      cluster_name: form.value.cluster_name,
      cluster_version: form.value.cluster_version,
    }
    if (hasKubeconfigUpdate) {
      payload.kube_config = form.value.kube_config
    }

    const body = await updateCluster(payload, { _silent: true })
    Message.success({
      content: hasKubeconfigUpdate 
        ? (body?.msg || '更新成功，建议立即执行健康检查') 
        : (body?.msg || '更新成功'),
      duration: hasKubeconfigUpdate ? 2600 : 1800
    })
    closeForm()
    await fetchList()
  } catch (e) {
    console.error(e)
    const body = unwrapErrorBody(e)
    Message.error({
      content: `提交失败：${pickMsg(body, e?.message || '请检查参数')}`,
      duration: 2600,
    })
  } finally {
    loading.value = false
  }
}

const onDelete = async (c) => {
  // 二次确认 - 删除集群是高危操作
  const okDel = await showConfirm({
    title: '确认删除集群',
    content: '删除集群将移除所有关联配置和 kubeconfig，此操作不可逆！',
    type: 'danger',
    details: [
      { label: '集群名称', value: c.cluster_name },
      { label: '集群 ID', value: String(c.id), mono: true },
      { label: 'K8s 版本', value: c.cluster_version || '-' },
      { label: '当前状态', value: statusText(c.status), highlight: true },
    ],
    tip: '此操作不可逆，请谨慎确认。',
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!okDel) return;
  
  loading.value = true
  try {
    const body = await deleteCluster({id: c.id})
    Message.success({content: body?.msg || '删除成功', duration: 1800})
    selectedIds.value = selectedIds.value.filter(id => id !== c.id)
    await fetchList()
  } catch (e) {
    console.error(e)
    const body = unwrapErrorBody(e)
    Message.error({
      content: `删除失败：${pickMsg(body, e?.message || '请查看后端日志')}`,
      duration: 2600,
    })
  } finally {
    loading.value = false
  }
}

// 批量删除
const onBatchDelete = async () => {
  if (selectedIds.value.length === 0) {
    Message.warning({content: '请先选择要删除的集群', duration: 1800})
    return
  }
  const okBatch = await showConfirm({
    title: `确认批量删除 ${selectedIds.value.length} 个集群`,
    content: '删除集群将移除所有关联配置和 kubeconfig，此操作不可逆！',
    type: 'danger',
    details: [
      { label: '删除数量', value: `${selectedIds.value.length} 个集群`, danger: true },
      { label: '集群 ID', value: selectedIds.value.join(', '), mono: true },
    ],
    tip: '批量删除不可恢复，请仔细核对已选集群。',
    confirmText: '确认批量删除',
    cancelText: '取消',
  })
  if (!okBatch) return;

  loading.value = true
  try {
    const body = await batchDeleteCluster({ids: selectedIds.value})
    Message.success({content: body?.data?.msg || body?.msg || `批量删除成功`, duration: 1800})
    selectedIds.value = []
    await fetchList()
  } catch (e) {
    console.error(e)
    const body = unwrapErrorBody(e)
    Message.error({
      content: `批量删除失败：${pickMsg(body, e?.message || '请查看后端日志')}`,
      duration: 2600,
    })
  } finally {
    loading.value = false
  }
}

// ✅ 健康检查：调用连通性检测接口，前端5秒超时
const testCluster = async (c) => {
  testingId.value = c.id
  try {
    // 前端5秒超时控制
    const timeout = new Promise((_, reject) => 
      setTimeout(() => reject(new Error('timeout')), 5000)
    )
    const request = http.get(`/api/v1/platform/health/cluster/${c.id}/connectivity`)
    
    const res = await Promise.race([request, timeout])
    if (res.code === 0 && res.data) {
      Message[res.data.connected ? 'success' : 'error']({ 
        content: res.data.connected ? 'ok' : '异常', 
        duration: 1500 
      })
    } else {
      Message.error({ content: '异常', duration: 1500 })
    }
  } catch (e) {
    Message.error({ content: '异常', duration: 1500 })
  } finally {
    testingId.value = null
    await fetchList()
  }
}

const testInModal = async () => {
  if (formMode.value !== 'edit' || !form.value.id) {
    Message.warning({content: '请先创建集群后再检查', duration: 1600})
    return
  }
  testingId.value = form.value.id
  try {
    const body = await initCluster({id: form.value.id})
    const ok = isOk(body)
    Message[ok ? 'success' : 'error']({
      content: ok ? (body?.msg || '初始化成功') : (body?.msg || '初始化失败'),
      duration: ok ? 1800 : 2600,
    })
  } catch (e) {
    const body = unwrapErrorBody(e)
    Message.error({
      content: `初始化失败：${pickMsg(body, e?.message || 'K8s 集群初始化失败')}`,
      duration: 2600,
    })
  } finally {
    testingId.value = null
    await fetchList()
  }
}

const formatCheckAt = (ts) => {
  const n = Number(ts || 0)
  if (!n) return '-'
  const d = new Date(n * 1000)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}
</script>

<style scoped>
/* ===== 全局容器 ===== */
.cluster-view {
  width: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  box-sizing: border-box;
  gap: 20px;
}

/* ===== 页面头部 ===== */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #326ce5 0%, #1d4ed8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.3);
}

.page-title-section h1 {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
  color: #0f172a;
  letter-spacing: -0.3px;
}

.page-subtitle {
  margin: 2px 0 0;
  font-size: 13px;
  color: #64748b;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* ===== 统计卡片 ===== */
.stats-overview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e8ecf2;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
  transform: translateY(-1px);
}

.stat-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.total {
  background: linear-gradient(135deg, #eef2ff 0%, #e0e7ff 100%);
  color: #4f46e5;
}

.stat-icon.healthy {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.stat-icon.unhealthy {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
}

.stat-icon.pending {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  color: #d97706;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
  font-weight: 500;
}

/* ===== 工具栏 ===== */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 14px 18px;
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e8ecf2;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-box {
  position: relative;
  width: 260px;
}

.search-box .search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
  pointer-events: none;
}

.search-box input {
  width: 100%;
  padding: 9px 12px 9px 38px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  outline: none;
  font-size: 13px;
  transition: all 0.2s;
  background: #f8fafc;
  color: #1e293b;
}

.search-box input:focus {
  border-color: #326ce5;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.08);
}

.filter-group {
  display: flex;
  gap: 2px;
  padding: 3px;
  background: #f1f5f9;
  border-radius: 10px;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.filter-btn:hover { color: #1e293b; }

.filter-btn.active {
  background: #fff;
  color: #326ce5;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  font-weight: 600;
}

.filter-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #cbd5e1;
}

.filter-dot.ok { background: #22c55e; }
.filter-dot.bad { background: #ef4444; }
.filter-dot.pending { background: #f59e0b; }
.filter-dot.all { background: #6366f1; }

.view-switch {
  display: flex;
  gap: 2px;
  padding: 3px;
  background: #f1f5f9;
  border-radius: 10px;
}

.view-btn {
  padding: 7px 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.view-btn:hover { color: #475569; }

.view-btn.active {
  background: #fff;
  color: #326ce5;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}

/* ===== 按钮系统 ===== */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 9px 18px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn:hover:not(:disabled) { transform: translateY(-1px); }
.btn:active:not(:disabled) { transform: translateY(0); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-primary {
  background: linear-gradient(135deg, #326ce5 0%, #2557c5 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.3);
}

.btn-primary:hover:not(:disabled) {
  box-shadow: 0 4px 14px rgba(50, 108, 229, 0.4);
}

.btn-ghost {
  background: #fff;
  color: #475569;
  border: 1.5px solid #e2e8f0;
}

.btn-ghost:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.btn-danger-solid {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}

/* ===== 表格 ===== */
.table-container {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e8ecf2;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  overflow: hidden;
}

.table-scroll {
  max-height: calc(100vh - 380px);
  min-height: 300px;
  overflow: auto;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
}

.resource-table th {
  background: #f8fafc;
  text-align: left;
  padding: 13px 16px;
  border-bottom: 1px solid #e8ecf2;
  font-weight: 600;
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.resource-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  font-size: 13px;
  color: #334155;
  vertical-align: middle;
}

.resource-table tbody tr {
  transition: background 0.15s;
}

.resource-table tbody tr:hover {
  background: #f8fafc;
}

.resource-table tbody tr.row-selected {
  background: #eff6ff;
}

.row-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #326ce5;
  border-radius: 4px;
}

/* 集群名称单元格 */
.cluster-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cluster-avatar {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cluster-avatar.connected {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.cluster-avatar.disconnected {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
}

.cluster-avatar.pending {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  color: #d97706;
}

.cluster-avatar.unknown {
  background: #f1f5f9;
  color: #64748b;
}

.cluster-name-info {
  min-width: 0;
}

.cluster-link {
  color: #1e40af;
  cursor: pointer;
  font-weight: 600;
  font-size: 13.5px;
  text-decoration: none;
  transition: color 0.15s;
}

.cluster-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.cluster-error-hint {
  margin-top: 3px;
  font-size: 11px;
  color: #94a3b8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 320px;
}

.id-badge {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  background: #f1f5f9;
  padding: 3px 8px;
  border-radius: 6px;
  font-family: 'SF Mono', monospace;
}

.version-tag {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  background: #f1f5f9;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}

.text-muted { color: #94a3b8; }

.time-text {
  font-size: 12.5px;
  color: #64748b;
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.connected {
  background: rgba(34, 197, 94, 0.1);
  color: #15803d;
}

.status-badge.disconnected {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.status-badge.pending {
  background: rgba(245, 158, 11, 0.1);
  color: #b45309;
}

.status-badge.unknown {
  background: rgba(148, 163, 184, 0.1);
  color: #64748b;
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  position: relative;
}

.status-badge.connected .status-dot {
  background: #22c55e;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.2);
  animation: pulse-green 2s infinite;
}

.status-badge.disconnected .status-dot {
  background: #ef4444;
}

.status-badge.pending .status-dot {
  background: #f59e0b;
  animation: pulse-yellow 2s infinite;
}

.status-badge.unknown .status-dot {
  background: #94a3b8;
}

@keyframes pulse-green {
  0%, 100% { box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.2); }
  50% { box-shadow: 0 0 0 5px rgba(34, 197, 94, 0); }
}

@keyframes pulse-yellow {
  0%, 100% { box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2); }
  50% { box-shadow: 0 0 0 5px rgba(245, 158, 11, 0); }
}

/* 操作按钮组 */
.action-group {
  display: flex;
  gap: 6px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border-radius: 8px;
  border: 1.5px solid #e2e8f0;
  background: #fff;
  font-size: 12px;
  font-weight: 500;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.action-btn-primary {
  color: #2563eb;
  border-color: #bfdbfe;
  background: #eff6ff;
}

.action-btn-primary:hover:not(:disabled) {
  background: #dbeafe;
  border-color: #93c5fd;
}

.action-btn-danger {
  color: #dc2626;
  border-color: #fecaca;
  background: #fef2f2;
  padding: 6px 8px;
}

.action-btn-danger:hover:not(:disabled) {
  background: #fee2e2;
  border-color: #fca5a5;
}

.spin-icon {
  display: inline-flex;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ===== 空状态 ===== */
.empty-state {
  padding: 80px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.empty-illustration {
  margin-bottom: 20px;
  opacity: 0.6;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: #334155;
  margin-bottom: 6px;
}

.empty-desc {
  font-size: 13px;
  color: #94a3b8;
}

/* ===== Modal ===== */
.modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 20px;
  overflow-y: auto;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
}

.modal-content {
  position: relative;
  width: 100%;
  max-width: 720px;
  max-height: calc(100vh - 80px);
  background: #fff;
  border-radius: 18px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 24px;
  background: #f8fafc;
  border-bottom: 1px solid #e8ecf2;
  flex-shrink: 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: #0f172a;
}

.close-btn {
  border: none;
  background: transparent;
  font-size: 22px;
  cursor: pointer;
  line-height: 1;
  color: #94a3b8;
  padding: 6px;
  border-radius: 8px;
  transition: all 0.15s;
}

.close-btn:hover {
  color: #1e293b;
  background: #e2e8f0;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

/* ===== Modal Form ===== */
.form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  background: #f8fafc;
  border: 1px solid #e8ecf2;
  border-radius: 12px;
  padding: 12px 16px;
}

.chip {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
  background: rgba(50, 108, 229, 0.1);
  border-radius: 20px;
  padding: 6px 12px;
}

.card {
  background: #fff;
  border: 1px solid #e8ecf2;
  border-radius: 14px;
  padding: 18px;
}

.card > .card-title {
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 14px;
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 14px;
}

.card > .card-title .hint {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 400;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.field label {
  display: block;
  font-size: 13px;
  color: #475569;
  margin-bottom: 6px;
  font-weight: 500;
}

.field input {
  width: 100%;
  height: 40px;
  border-radius: 10px;
  border: 1.5px solid #e2e8f0;
  padding: 0 14px;
  outline: none;
  font-size: 13px;
  transition: all 0.15s;
  box-sizing: border-box;
}

.field input:focus {
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.08);
}

.field input:disabled {
  background: #f8fafc;
  color: #64748b;
}

.required { color: #ef4444; }

.upload-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.file {
  flex: 1;
  min-width: 200px;
}

.alert {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  background: rgba(245, 158, 11, 0.06);
  border: 1px solid rgba(245, 158, 11, 0.15);
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 12px;
}

.alert-icon { font-size: 16px; line-height: 1.4; }

.alert-text {
  font-size: 12.5px;
  color: #92400e;
  line-height: 1.5;
}

.codebox {
  width: 100%;
  min-height: 200px;
  resize: vertical;
  border-radius: 12px;
  border: 1.5px solid #1e293b;
  padding: 14px;
  background: #0f172a;
  color: #e2e8f0;
  font-family: 'SF Mono', 'Menlo', 'Monaco', 'Consolas', monospace;
  font-size: 12.5px;
  line-height: 1.6;
  outline: none;
  box-sizing: border-box;
}

.codebox:focus {
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
}

.btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2557c5 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(50, 108, 229, 0.3);
}

.btn.ghost {
  background: #f1f5f9;
  color: #475569;
}

.btn.ghost:hover {
  background: #e2e8f0;
}

.btn.small {
  padding: 8px 14px;
  border-radius: 8px;
}

.muted { color: #64748b; }

/* ===== 卡片视图 ===== */
.cards-container {
  padding: 0;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 18px;
  margin-bottom: 20px;
}

.cluster-card {
  background: white;
  border: 1px solid #e8ecf2;
  border-radius: 16px;
  overflow: hidden;
  transition: all 0.25s ease;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  position: relative;
}

.cluster-card:hover {
  box-shadow: 0 8px 24px rgba(0,0,0,0.08);
  transform: translateY(-3px);
  border-color: #cbd5e1;
}

.card-status-bar {
  height: 3px;
  width: 100%;
}

.card-status-bar.connected { background: linear-gradient(90deg, #22c55e, #4ade80); }
.card-status-bar.disconnected { background: linear-gradient(90deg, #ef4444, #f87171); }
.card-status-bar.pending { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.card-status-bar.unknown { background: #e2e8f0; }

.card-header {
  padding: 18px 20px 0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-avatar.connected {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.card-avatar.disconnected {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
}

.card-avatar.pending {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  color: #d97706;
}

.card-avatar.unknown {
  background: #f1f5f9;
  color: #64748b;
}

.card-title-info {
  flex: 1;
  min-width: 0;
}

.card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
}

.card-title .cluster-link {
  color: #1e293b;
}

.card-title .cluster-link:hover {
  color: #1d4ed8;
}

.card-id {
  font-size: 11px;
  color: #94a3b8;
  font-family: 'SF Mono', monospace;
  margin-top: 2px;
}

.card-body {
  padding: 16px 20px;
}

.card-meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.card-meta-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.meta-icon {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  flex-shrink: 0;
}

.meta-label {
  font-size: 11px;
  color: #94a3b8;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.meta-value {
  font-size: 13px;
  color: #1e293b;
  font-weight: 600;
  margin-top: 2px;
}

.card-error {
  margin-top: 12px;
  padding: 10px 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #dc2626;
  font-size: 12px;
}

.card-error span {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  border-top: 1px solid #f1f5f9;
  background: #fafbfc;
}

.card-btn {
  padding: 7px 14px;
  border-radius: 8px;
  border: 1.5px solid #e2e8f0;
  background: #fff;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.card-btn:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.card-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.card-btn-enter {
  background: linear-gradient(135deg, #326ce5 0%, #2557c5 100%);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 2px 6px rgba(50, 108, 229, 0.2);
}

.card-btn-enter:hover:not(:disabled) {
  background: linear-gradient(135deg, #2557c5 0%, #1e40af 100%);
  border-color: transparent;
}

.card-btn-more {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

.card-btn-icon {
  padding: 7px;
  min-width: 32px;
  justify-content: center;
}

.card-btn-danger {
  color: #dc2626;
  border-color: #fecaca;
}

.card-btn-danger:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #fca5a5;
}

/* ===== 响应式 ===== */
@media (max-width: 1024px) {
  .stats-overview {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-overview {
    grid-template-columns: 1fr 1fr;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .toolbar-left {
    flex-direction: column;
  }
  
  .search-box {
    width: 100%;
  }
  
  .table-scroll {
    max-height: calc(100vh - 440px);
  }
  
  .cards-grid {
    grid-template-columns: 1fr;
  }
  
  .modal {
    padding: 16px;
  }
  
  .modal-content {
    max-height: calc(100vh - 32px);
  }
}
</style>

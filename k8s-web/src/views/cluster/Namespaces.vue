<template>
  <div class="resource-view">
    <!-- 页面头部 - Rancher 风格 -->
    <div class="view-header">
      <div class="header-left">
        <div class="header-icon">🗂️</div>
        <div class="header-text">
          <h1>命名空间</h1>
          <p>管理 Kubernetes 集群中的命名空间，用于隔离资源和访问控制</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <div class="stat-value">{{ namespaces.length }}</div>
          <div class="stat-label">总数</div>
        </div>
        <div class="stat-item active">
          <div class="stat-value">{{ activeCount }}</div>
          <div class="stat-label">Active</div>
        </div>
        <div class="stat-item warning">
          <div class="stat-value">{{ terminatingCount }}</div>
          <div class="stat-label">Terminating</div>
        </div>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="action-bar-left">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="搜索命名空间..." 
            @input="onSearchInput"
            @keyup.enter="refreshList"
          />
        </div>
        
        <!-- 状态筛选 -->
        <div class="filter-buttons">
          <button 
            class="btn btn-filter" 
            :class="{ active: statusFilter === 'all' }"
            @click="setStatusFilter('all')"
          >
            全部
          </button>
          <button 
            class="btn btn-filter" 
            :class="{ active: statusFilter === 'Active' }"
            @click="setStatusFilter('Active')"
          >
            Active
          </button>
          <button 
            class="btn btn-filter" 
            :class="{ active: statusFilter === 'Terminating' }"
            @click="setStatusFilter('Terminating')"
          >
            Terminating
          </button>
        </div>
      </div>

      <div class="action-bar-right">
        <!-- 自动刷新开关 -->
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          自动刷新
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        
        <!-- 批量操作按钮 -->
        <button 
          v-if="!batchMode" 
          class="btn btn-batch" 
          @click="enterBatchMode"
          title="进入批量操作模式"
        >
          ☑️ 批量操作
        </button>
        <button 
          v-if="batchMode" 
          class="btn btn-secondary" 
          @click="exitBatchMode"
        >
          ✖️ 退出批量
        </button>

        <!-- 视图切换 -->
        <div class="view-toggle">
          <button 
            class="btn btn-view" 
            :class="{ active: viewMode === 'table' }" 
            @click="viewMode = 'table'"
            title="表格视图"
          >
            📋
          </button>
          <button 
            class="btn btn-view" 
            :class="{ active: viewMode === 'card' }" 
            @click="viewMode = 'card'"
            title="卡片视图"
          >
            🗂️
          </button>
        </div>
        
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '刷新中...' : '🔄 刷新' }}
        </button>
        <button class="btn btn-primary" @click="showCreateModal = true">
          ➕ 创建命名空间
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && namespaces.length === 0" class="loading-state">
      <div class="loading-spinner"></div>
      <div class="loading-text">加载中...</div>
    </div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedNamespaces.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedNamespaces.length }} 个命名空间</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除（高危）">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- ========== 表格视图 ========== -->
    <div v-if="viewMode === 'table'" class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input 
                type="checkbox" 
                :checked="isAllSelected" 
                @change="toggleSelectAll"
                title="全选/取消全选"
                ref="selectAllCheckbox"
              />
            </th>
            <th style="width: 100px;">状态</th>
            <th style="width: 200px;">名称</th>
            <th style="width: 100px; text-align: center;">Pods</th>
            <th style="width: 100px; text-align: center;">Services</th>
            <th style="width: 110px; text-align: center;">Deployments</th>
            <th style="width: 150px;">资源使用</th>
            <th style="width: 140px;">创建时间</th>
            <th style="width: 180px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="namespace in paginatedNamespaces" :key="namespace.name" class="table-row" :class="{ 'row-selected': isNamespaceSelected(namespace) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isNamespaceSelected(namespace)" 
                @change="toggleNamespaceSelection(namespace)"
                :disabled="isSystemNamespace(namespace.name)"
              />
            </td>
            <td>
              <span class="status-badge" :class="namespace.status.toLowerCase()">
                <span class="status-dot"></span>
                {{ namespace.status }}
              </span>
            </td>
            <td>
              <div class="namespace-info">
                <span class="namespace-icon">📁</span>
                <span class="namespace-name" :title="namespace.name">{{ namespace.name }}</span>
                <span v-if="isSystemNamespace(namespace.name)" class="system-tag">系统</span>
              </div>
            </td>
            <td style="text-align: center;">
              <span class="count-badge pods">{{ namespace.podCount }}</span>
            </td>
            <td style="text-align: center;">
              <span class="count-badge services">{{ namespace.serviceCount }}</span>
            </td>
            <td style="text-align: center;">
              <span class="count-badge deployments">{{ namespace.deploymentCount }}</span>
            </td>
            <td>
              <div class="resource-usage">
                <div class="usage-item">
                  <span class="usage-icon">⚡</span>
                  <div class="usage-bar-wrap">
                    <div class="usage-bar">
                      <div class="usage-fill" :style="{ width: namespace.cpuUsage }" :class="getUsageClass(namespace.cpuUsage)"></div>
                    </div>
                  </div>
                  <span class="usage-text">{{ namespace.cpuUsage }}</span>
                </div>
                <div class="usage-item">
                  <span class="usage-icon">💾</span>
                  <div class="usage-bar-wrap">
                    <div class="usage-bar">
                      <div class="usage-fill" :style="{ width: namespace.memoryUsage }" :class="getUsageClass(namespace.memoryUsage)"></div>
                    </div>
                  </div>
                  <span class="usage-text">{{ namespace.memoryUsage }}</span>
                </div>
              </div>
            </td>
            <td>
              <span class="date-text">{{ namespace.createdAt }}</span>
            </td>
            <td>
              <div class="action-buttons-inline">
                <button class="action-btn" @click="viewNamespace(namespace)" title="查看详情">
                  👁️ 详情
                </button>
                <button class="action-btn" @click="openYamlPreview(namespace)" title="查看/编辑 YAML">
                  📝 YAML
                </button>
                <button 
                  class="action-btn danger" 
                  @click="confirmDelete(namespace)" 
                  title="删除"
                  :disabled="isSystemNamespace(namespace.name)"
                >
                  🗑️ 删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      
      <div v-if="!loading && filteredNamespaces.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-text">没有找到匹配的命名空间</div>
        <button class="btn btn-primary" @click="showCreateModal = true">创建第一个命名空间</button>
      </div>
      
      <Pagination 
        v-if="total > 0" 
        v-model:currentPage="currentPage" 
        :totalItems="total" 
        :itemsPerPage="itemsPerPage" 
      />
    </div>

    <!-- ========== 卡片视图 ========== -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="!loading && filteredNamespaces.length === 0" class="empty-state">
        <div class="empty-icon">🗂️</div>
        <div class="empty-text">没有找到匹配的命名空间</div>
        <button class="btn btn-primary" @click="showCreateModal = true">创建第一个命名空间</button>
      </div>
      
      <div class="cards-grid">
        <div 
          v-for="namespace in paginatedNamespaces" 
          :key="namespace.name" 
          class="namespace-card"
          :class="{ 
            'system-namespace': isSystemNamespace(namespace.name),
            'card-selected': isNamespaceSelected(namespace)
          }"
        >
          <!-- 批量选择复选框 -->
          <div v-if="batchMode && !isSystemNamespace(namespace.name)" class="card-checkbox">
            <input 
              type="checkbox" 
              :checked="isNamespaceSelected(namespace)" 
              @change="toggleNamespaceSelection(namespace)"
            />
          </div>
          
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">📁</span>
              <h3 class="card-title" :title="namespace.name">{{ namespace.name }}</h3>
            </div>
            <div class="card-badges">
              <span class="status-badge" :class="namespace.status.toLowerCase()">
                <span class="status-dot"></span>
                {{ namespace.status }}
              </span>
              <span v-if="isSystemNamespace(namespace.name)" class="system-tag">系统</span>
            </div>
          </div>

          <!-- 资源统计 -->
          <div class="card-stats">
            <div class="stat-box clickable" @click="navigateToResource(namespace, 'pods')" title="点击查看 Pods 列表">
              <div class="stat-icon pods">📦</div>
              <div class="stat-content">
                <div class="stat-number">{{ namespace.podCount }}</div>
                <div class="stat-name">Pods</div>
              </div>
            </div>
            <div class="stat-box clickable" @click="navigateToResource(namespace, 'services')" title="点击查看 Services 列表">
              <div class="stat-icon services">🔌</div>
              <div class="stat-content">
                <div class="stat-number">{{ namespace.serviceCount }}</div>
                <div class="stat-name">Services</div>
              </div>
            </div>
            <div class="stat-box clickable" @click="navigateToResource(namespace, 'deployments')" title="点击查看 Deployments 列表">
              <div class="stat-icon deployments">🚀</div>
              <div class="stat-content">
                <div class="stat-number">{{ namespace.deploymentCount }}</div>
                <div class="stat-name">Deployments</div>
              </div>
            </div>
          </div>

          <!-- 资源使用 -->
          <div class="card-section">
            <div class="section-label">资源使用</div>
            <div class="resource-bars">
              <div class="resource-bar-item">
                <div class="bar-label">
                  <span>⚡ CPU</span>
                  <span>{{ namespace.cpuUsage }}</span>
                </div>
                <div class="bar-track">
                  <div class="bar-fill" :style="{ width: namespace.cpuUsage }" :class="getUsageClass(namespace.cpuUsage)"></div>
                </div>
              </div>
              <div class="resource-bar-item">
                <div class="bar-label">
                  <span>💾 内存</span>
                  <span>{{ namespace.memoryUsage }}</span>
                </div>
                <div class="bar-track">
                  <div class="bar-fill" :style="{ width: namespace.memoryUsage }" :class="getUsageClass(namespace.memoryUsage)"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 创建时间 -->
          <div class="card-section">
            <div class="section-label">创建时间</div>
            <div class="section-value">{{ namespace.createdAt }}</div>
          </div>

          <!-- 卡片操作 -->
          <div class="card-footer">
            <button class="card-action-btn primary" @click="viewNamespace(namespace)" title="查看详情">
              👁️ 详情
            </button>
            <button class="card-action-btn" @click="openYamlPreview(namespace)" title="查看/编辑 YAML">
              📝 YAML
            </button>
            <button 
              class="card-action-btn danger" 
              @click="confirmDelete(namespace)" 
              title="删除"
              :disabled="isSystemNamespace(namespace.name)"
            >
              🗑️ 删除
            </button>
          </div>
        </div>
      </div>
      
      <Pagination 
        v-if="total > 0" 
        v-model:currentPage="currentPage" 
        :totalItems="total" 
        :itemsPerPage="itemsPerPage" 
      />
    </div>

    <!-- ========== 创建命名空间模态框 ========== -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal-content modal-create">
        <div class="modal-header">
          <h3>🗂️ 创建新命名空间</h3>
          <button class="modal-close" @click="closeModals">×</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="createNamespace">
            <div class="form-section">
              <div class="form-group">
                <label for="name">
                  命名空间名称 <span class="required">*</span>
                </label>
                <input 
                  type="text" 
                  id="name" 
                  v-model="namespaceForm.name" 
                  required 
                  placeholder="my-namespace" 
                  pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?"
                />
                <div class="form-hint">
                  小写字母、数字和连字符，必须以字母或数字开头和结尾
                </div>
              </div>
              
              <div class="form-group">
                <label for="description">描述（可选）</label>
                <input 
                  type="text" 
                  id="description" 
                  v-model="namespaceForm.description" 
                  placeholder="命名空间用途说明..."
                />
              </div>
            </div>
            
            <!-- 资源配额提示 -->
            <div class="quota-hint">
              <div class="hint-icon">💡</div>
              <div class="hint-content">
                <div class="hint-title">默认资源配额</div>
                <div class="hint-text">
                  系统将自动创建默认配额：CPU 4核、内存 8Gi、Pod 上限 110 个
                </div>
              </div>
            </div>
          </form>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModals">取消</button>
          <button class="btn btn-primary" @click="createNamespace" :disabled="loading || !namespaceForm.name">
            {{ loading ? '创建中...' : '✅ 创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========== 查看详情模态框 ========== -->
    <div v-if="showViewModal && selectedNamespace" class="modal-overlay" @click.self="closeModals">
      <div class="modal-content modal-detail">
        <div class="modal-header">
          <h3>📋 命名空间详情</h3>
          <button class="modal-close" @click="closeModals">×</button>
        </div>
        <div class="modal-body">
          <!-- 基本信息卡片 -->
          <div class="detail-card">
            <div class="detail-card-header">
              <span class="detail-card-icon">📁</span>
              <span>基本信息</span>
            </div>
            <div class="detail-card-body">
              <div class="detail-row">
                <span class="detail-label">名称</span>
                <span class="detail-value">{{ selectedNamespace.name }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">状态</span>
                <span class="status-badge" :class="selectedNamespace.status.toLowerCase()">
                  <span class="status-dot"></span>
                  {{ selectedNamespace.status }}
                </span>
              </div>
              <div class="detail-row">
                <span class="detail-label">创建时间</span>
                <span class="detail-value">{{ selectedNamespace.createdAt }}</span>
              </div>
            </div>
          </div>

          <!-- 资源统计卡片 -->
          <div class="detail-card">
            <div class="detail-card-header">
              <span class="detail-card-icon">📊</span>
              <span>资源统计</span>
              <span class="header-hint">点击可跳转到对应资源列表</span>
            </div>
            <div class="detail-card-body">
              <div class="stats-grid">
                <div class="stats-item clickable" @click="navigateToResource(selectedNamespace, 'pods')" title="点击查看 Pods 列表">
                  <div class="stats-icon pods">📦</div>
                  <div class="stats-info">
                    <div class="stats-number">{{ selectedNamespace.podCount }}</div>
                    <div class="stats-label">Pods</div>
                  </div>
                  <div class="stats-arrow">→</div>
                </div>
                <div class="stats-item clickable" @click="navigateToResource(selectedNamespace, 'services')" title="点击查看 Services 列表">
                  <div class="stats-icon services">🔌</div>
                  <div class="stats-info">
                    <div class="stats-number">{{ selectedNamespace.serviceCount }}</div>
                    <div class="stats-label">Services</div>
                  </div>
                  <div class="stats-arrow">→</div>
                </div>
                <div class="stats-item clickable" @click="navigateToResource(selectedNamespace, 'deployments')" title="点击查看 Deployments 列表">
                  <div class="stats-icon deployments">🚀</div>
                  <div class="stats-info">
                    <div class="stats-number">{{ selectedNamespace.deploymentCount }}</div>
                    <div class="stats-label">Deployments</div>
                  </div>
                  <div class="stats-arrow">→</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Labels 和 Annotations -->
          <div class="detail-card" v-if="selectedNamespace.labels || selectedNamespace.annotations">
            <div class="detail-card-header">
              <span class="detail-card-icon">🏷️</span>
              <span>标签与注解</span>
            </div>
            <div class="detail-card-body">
              <div v-if="selectedNamespace.labels && Object.keys(selectedNamespace.labels).length > 0" class="labels-section">
                <div class="section-subtitle">Labels</div>
                <div class="tags-container">
                  <span v-for="(value, key) in selectedNamespace.labels" :key="key" class="label-tag">
                    {{ key }}={{ value }}
                  </span>
                </div>
              </div>
              <div v-if="selectedNamespace.annotations && Object.keys(selectedNamespace.annotations).length > 0" class="annotations-section">
                <div class="section-subtitle">Annotations</div>
                <div class="tags-container">
                  <span v-for="(value, key) in selectedNamespace.annotations" :key="key" class="annotation-tag" :title="value">
                    {{ key }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModals">关闭</button>
        </div>
      </div>
    </div>

    <!-- ========== 删除确认模态框 ========== -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal-content modal-delete">
        <div class="modal-header danger">
          <h3>⚠️ 删除命名空间</h3>
          <button class="modal-close" @click="closeModals">×</button>
        </div>
        <div class="modal-body">
          <div class="delete-warning">
            <div class="warning-icon">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">危险操作！</div>
              <div class="warning-text">
                删除命名空间将会<strong>级联删除</strong>该空间下的所有资源，包括：
              </div>
              <ul class="warning-list">
                <li>所有 Pods、Deployments、StatefulSets</li>
                <li>所有 Services、ConfigMaps、Secrets</li>
                <li>所有 PVC、ServiceAccounts 等</li>
              </ul>
            </div>
          </div>
          
          <div class="delete-target">
            <div class="target-label">即将删除：</div>
            <div class="target-name">{{ namespaceToDelete?.name }}</div>
            <div class="target-stats" v-if="namespaceToDelete">
              <span>📦 {{ namespaceToDelete.podCount }} Pods</span>
              <span>🔌 {{ namespaceToDelete.serviceCount }} Services</span>
              <span>🚀 {{ namespaceToDelete.deploymentCount }} Deployments</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModals">取消</button>
          <button class="btn btn-danger" @click="deleteNamespace" :disabled="loading">
            {{ loading ? '删除中...' : '🗑️ 确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除预览弹窗（高危操作） -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="closeBatchDeleteModal">
      <div class="modal-content modal-batch-preview modal-danger">
        <div class="modal-header danger-header">
          <h3>🗑️ 批量删除命名空间预览（高危操作）</h3>
          <button class="close-btn" @click="closeBatchDeleteModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 警告提示 -->
          <div class="danger-warning">
            <div class="warning-icon-large">⚠️</div>
            <div class="warning-content">
              <div class="warning-title">将删除以下命名空间及其所有资源</div>
              <ul class="warning-list">
                <li>命名空间内的所有 Pod、Service、Deployment 等资源将被删除</li>
                <li>此操作不可撤销！</li>
              </ul>
            </div>
          </div>
          
          <!-- 受影响命名空间列表 -->
          <div class="preview-section">
            <div class="section-title">受影响命名空间 ({{ selectedNamespaces.length }})</div>
            <div class="affected-namespaces-detail">
              <div v-for="ns in selectedNamespaces" :key="ns.name" class="affected-ns-card">
                <div class="ns-info">
                  <span class="ns-name">📁 {{ ns.name }}</span>
                  <span class="status-badge" :class="ns.status.toLowerCase()">{{ ns.status }}</span>
                </div>
                <div class="ns-stats">
                  <span class="stat-tag">📦 {{ ns.podCount }} Pods</span>
                  <span class="stat-tag">🔌 {{ ns.serviceCount }} Services</span>
                  <span class="stat-tag">🚀 {{ ns.deploymentCount }} Deployments</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 二次确认 -->
          <div class="confirm-section">
            <div class="section-title">请输入 "DELETE" 确认操作</div>
            <input 
              v-model="deleteConfirmText" 
              placeholder="请输入 DELETE" 
              class="confirm-input"
              :class="{ valid: deleteConfirmText === 'DELETE' }"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchDeleteModal">取消</button>
          <button 
            class="btn btn-danger" 
            @click="executeBatchDelete" 
            :disabled="deleteConfirmText !== 'DELETE' || batchExecuting"
          >
            {{ batchExecuting ? '执行中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========== YAML 查看/编辑模态框 ========== -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content yaml-modal">
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlNamespace?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true">
              ✈️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="modal-close" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="loadingYaml" class="loading-state">
            <div class="loading-spinner"></div>
            <div class="loading-text">Loading YAML...</div>
          </div>
          <div v-else-if="yamlError" class="error-box">{{ yamlError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea v-if="yamlEditMode" v-model="yamlContent" class="yaml-editor" spellcheck="false"></textarea>
            <pre v-else class="yaml-content">{{ yamlContent }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">取消</button>
          <button v-if="yamlEditMode" class="btn btn-primary" @click="applyYamlChanges" :disabled="savingYaml">
            {{ savingYaml ? '保存中...' : '应用更改' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, watchEffect } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import namespacesApi from '@/api/cluster/namespaces'

const router = useRouter()
const route = useRoute()

// ========== 状态变量 ==========
const searchQuery = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10) // 与 Pods.vue 保持一致
const loading = ref(false)
const viewMode = ref('card') // 默认卡片视图
const statusFilter = ref('all')
let searchDebounceTimer = null

// ========== 自动刷新 ==========
const autoRefresh = ref(false)
let autoRefreshTimer = null

const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    fetchNamespaces()
  }, 90000) // 90秒刷新
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

watch(autoRefresh, (val) => val ? startAutoRefresh() : stopAutoRefresh())

// 模态框状态
const showCreateModal = ref(false)
const showViewModal = ref(false)
const showDeleteModal = ref(false)

// ========== YAML 查看/编辑相关 ==========
const showYamlModal = ref(false)
const selectedYamlNamespace = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)
const yamlError = ref('')

// 数据
const namespaces = ref([])
const namespaceForm = ref({ name: '', description: '' })
const namespaceToDelete = ref(null)
const selectedNamespace = ref(null)

// 系统命名空间列表
const systemNamespaces = ['kube-system', 'kube-public', 'kube-node-lease', 'default']

// ========== 批量操作相关 ==========
const batchMode = ref(false)
const selectedNamespaces = ref([])
const showBatchDeleteModal = ref(false)
const deleteConfirmText = ref('')
const batchExecuting = ref(false)

const enterBatchMode = () => {
  batchMode.value = true
  // 默认全选当前页（排除系统命名空间），用户可取消不需要的项
  selectedNamespaces.value = paginatedNamespaces.value.filter(ns => !isSystemNamespace(ns.name))
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedNamespaces.value = []
}

const clearSelection = () => {
  selectedNamespaces.value = []
}

const isNamespaceSelected = (ns) => {
  return selectedNamespaces.value.some(n => n.name === ns.name)
}

const toggleNamespaceSelection = (ns) => {
  if (isSystemNamespace(ns.name)) return
  
  const index = selectedNamespaces.value.findIndex(n => n.name === ns.name)
  if (index >= 0) {
    selectedNamespaces.value.splice(index, 1)
  } else {
    selectedNamespaces.value.push(ns)
  }
}

const isAllSelected = computed(() => {
  const selectableNs = paginatedNamespaces.value.filter(ns => !isSystemNamespace(ns.name))
  return selectableNs.length > 0 && selectableNs.every(ns => isNamespaceSelected(ns))
})

// 部分选中状态（排除系统命名空间）
const isPartialSelected = computed(() => {
  const selectableNs = paginatedNamespaces.value.filter(ns => !isSystemNamespace(ns.name))
  if (selectableNs.length === 0) return false
  const selectedCount = selectableNs.filter(ns => isNamespaceSelected(ns)).length
  return selectedCount > 0 && selectedCount < selectableNs.length
})

// 全选复选框 ref
const selectAllCheckbox = ref(null)

// 设置 indeterminate 状态
watchEffect(() => {
  if (selectAllCheckbox.value) {
    selectAllCheckbox.value.indeterminate = isPartialSelected.value
  }
})

const toggleSelectAll = () => {
  const selectableNs = paginatedNamespaces.value.filter(ns => !isSystemNamespace(ns.name))
  if (isAllSelected.value) {
    selectableNs.forEach(ns => {
      const index = selectedNamespaces.value.findIndex(n => n.name === ns.name)
      if (index >= 0) selectedNamespaces.value.splice(index, 1)
    })
  } else {
    selectableNs.forEach(ns => {
      if (!isNamespaceSelected(ns)) {
        selectedNamespaces.value.push(ns)
      }
    })
  }
}

// 批量删除
const openBatchDeletePreview = () => {
  deleteConfirmText.value = ''
  showBatchDeleteModal.value = true
}

const closeBatchDeleteModal = () => {
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
}

const executeBatchDelete = async () => {
  if (deleteConfirmText.value !== 'DELETE') return
  
  batchExecuting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const ns of selectedNamespaces.value) {
    try {
      await namespacesApi.delete({ name: ns.name })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${ns.name}:`, e)
      failCount++
    }
  }
  
  batchExecuting.value = false
  showBatchDeleteModal.value = false
  deleteConfirmText.value = ''
  
  if (failCount === 0) {
    Message.success({ content: `成功删除 ${successCount} 个命名空间`, duration: 2200 })
  } else {
    Message.warning({ content: `成功 ${successCount} 个，失败 ${failCount} 个`, duration: 2200 })
  }
  
  exitBatchMode()
  fetchNamespaces()
}

// ========== 计算属性 ==========
const activeCount = computed(() => namespaces.value.filter(ns => ns.status === 'Active').length)
const terminatingCount = computed(() => namespaces.value.filter(ns => ns.status === 'Terminating').length)

const filteredNamespaces = computed(() => {
  let result = namespaces.value

  // 状态筛选
  if (statusFilter.value !== 'all') {
    result = result.filter(ns => ns.status === statusFilter.value)
  }

  // 搜索筛选
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(ns => ns.name.toLowerCase().includes(query))
  }

  return result
})

const paginatedNamespaces = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredNamespaces.value.slice(start, start + itemsPerPage.value)
})

// 总数 = 过滤后的数量（与 Pods.vue 保持一致）
const total = computed(() => filteredNamespaces.value.length)

// ========== 生命周期 ==========
onMounted(() => {
  fetchNamespaces()
})

onUnmounted(() => {
  stopAutoRefresh()
})

// ========== 方法 ==========
const fetchNamespaces = async () => {
  loading.value = true
  try {
    const response = await namespacesApi.list({
      name: searchQuery.value || '',
      page: 1,
      limit: 1000,
    })
    
    const list = response?.data?.list || response?.data?.items || response?.list || response?.items || []
    
    namespaces.value = Array.isArray(list) ? list.map(ns => ({
      name: ns.name || ns.metadata?.name || '',
      status: ns.status?.phase || 'Active',
      podCount: ns.pod_num || 0,
      serviceCount: ns.service_num || 0,
      deploymentCount: ns.deployment_num || 0,
      memoryUsage: '0%',
      cpuUsage: '0%',
      createdAt: formatDate(ns.creation_timestamp || ns.metadata?.creationTimestamp),
      labels: ns.labels || ns.metadata?.labels || {},
      annotations: ns.annotations || ns.metadata?.annotations || {},
    })) : []
    
  } catch (error) {
    console.error('Failed to fetch namespaces:', error)
    Message.error({ content: '获取命名空间列表失败', duration: 2200 })
    namespaces.value = []
  } finally {
    loading.value = false
  }
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toISOString().split('T')[0]
}

const isSystemNamespace = (name) => systemNamespaces.includes(name)

const getUsageClass = (usage) => {
  const percentage = parseInt(usage) || 0
  if (percentage > 80) return 'high'
  if (percentage > 50) return 'medium'
  return 'low'
}

// 状态筛选（前端过滤，不需要重新请求）
const setStatusFilter = (status) => {
  statusFilter.value = status
  currentPage.value = 1
}

// 搜索输入防抖（前端过滤）
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1
  }, 300)
}

const createNamespace = async () => {
  if (!namespaceForm.value.name) {
    Message.warning({ content: '请输入命名空间名称', duration: 2200 })
    return
  }

  loading.value = true
  try {
    await namespacesApi.create({
      name: namespaceForm.value.name,
      description: namespaceForm.value.description || '',
    })
    
    Message.success({ content: '命名空间创建成功', duration: 2200 })
    showCreateModal.value = false
    namespaceForm.value = { name: '', description: '' }
    await fetchNamespaces()
    // 创建后启动自动刷新 15 秒，观察状态变化
    autoRefresh.value = true
    setTimeout(() => { autoRefresh.value = false }, 15000)
  } catch (error) {
    console.error('Failed to create namespace:', error)
    const errMsg = error?.response?.data?.message || error?.message || '创建失败'
    Message.error({ content: `创建命名空间失败: ${errMsg}`, duration: 2200 })
  } finally {
    loading.value = false
  }
}

const viewNamespace = async (namespace) => {
  loading.value = true
  try {
    const response = await namespacesApi.detail({ name: namespace.name })
    const data = response?.data || response || {}
    
    // 合并列表数据和详情数据，优先使用详情接口返回的数据
    selectedNamespace.value = {
      ...namespace,
      name: data.name || namespace.name,
      status: data.status || namespace.status,
      labels: data.labels || namespace.labels || {},
      annotations: data.annotations || namespace.annotations || {},
      createdAt: formatDate(data.created_at) || namespace.createdAt,
      // 资源统计 - 从详情接口获取
      podCount: data.pod_num ?? namespace.podCount ?? 0,
      serviceCount: data.service_num ?? namespace.serviceCount ?? 0,
      deploymentCount: data.deployment_num ?? namespace.deploymentCount ?? 0,
    }
    showViewModal.value = true
  } catch (error) {
    console.error('Failed to fetch namespace detail:', error)
    // 即使详情获取失败，也显示基本信息
    selectedNamespace.value = { ...namespace }
    showViewModal.value = true
  } finally {
    loading.value = false
  }
}

const viewResources = (namespace) => {
  // TODO: 跳转到资源列表或显示资源弹窗
  Message.info({ content: `查看 ${namespace.name} 的资源`, duration: 2200 })
}

// 跳转到对应资源页面
const navigateToResource = (namespace, resourceType) => {
  // 获取当前集群 ID
  const clusterId = route.params.clusterId
  const nsName = typeof namespace === 'string' ? namespace : namespace.name
  
  // 关闭模态框
  closeModals()
  
  // 根据资源类型跳转到不同页面
  const routes = {
    pods: `/c/${clusterId}/workloads/pods`,
    services: `/c/${clusterId}/network/services`,
    deployments: `/c/${clusterId}/workloads/deployments`,
  }
  
  const targetPath = routes[resourceType]
  if (targetPath) {
    // 跳转并携带命名空间参数
    router.push({
      path: targetPath,
      query: { namespace: nsName }
    })
  }
}

const confirmDelete = (namespace) => {
  if (isSystemNamespace(namespace.name)) {
    Message.warning({ content: '系统命名空间不允许删除', duration: 2200 })
    return
  }
  namespaceToDelete.value = namespace
  showDeleteModal.value = true
}

const deleteNamespace = async () => {
  if (!namespaceToDelete.value) return
  
  loading.value = true
  try {
    await namespacesApi.delete({ name: namespaceToDelete.value.name })
    
    Message.success({ content: '命名空间删除成功', duration: 2200 })
    showDeleteModal.value = false
    namespaceToDelete.value = null
    await fetchNamespaces()
  } catch (error) {
    console.error('Failed to delete namespace:', error)
    const errMsg = error?.response?.data?.message || error?.message || '删除失败'
    Message.error({ content: `删除命名空间失败: ${errMsg}`, duration: 2200 })
  } finally {
    loading.value = false
  }
}

const refreshList = () => {
  fetchNamespaces()
}

const closeModals = () => {
  showCreateModal.value = false
  showViewModal.value = false
  showDeleteModal.value = false
}

// ========== YAML 查看/编辑功能 ==========
const openYamlPreview = async (namespace) => {
  selectedYamlNamespace.value = namespace
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
  loadingYaml.value = true
  showYamlModal.value = true

  try {
    const res = await namespacesApi.yaml({ name: namespace.name })
    if (res.code === 0) {
      yamlContent.value = res.data?.yaml || res.data || '# YAML 内容为空'
    } else {
      yamlError.value = res.msg || '获取 YAML 失败'
    }
  } catch (e) {
    yamlError.value = e?.msg || e?.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlNamespace.value = null
  yamlContent.value = ''
  yamlError.value = ''
  yamlEditMode.value = false
}

const applyYamlChanges = async () => {
  if (!yamlContent.value?.trim()) {
    Message.error({ content: '请输入 YAML 内容' })
    return
  }
  savingYaml.value = true
  try {
    const res = await namespacesApi.applyYaml({
      name: selectedYamlNamespace.value.name,
      yaml: yamlContent.value
    })
    if (res.code === 0) {
      Message.success({ content: 'YAML 应用成功' })
      closeYamlModal()
      fetchNamespaces()
      autoRefresh.value = true
      setTimeout(() => { autoRefresh.value = false }, 15000)
    } else {
      Message.error({ content: res.msg || '应用 YAML 失败' })
    }
  } catch (e) {
    Message.error({ content: e?.msg || e?.message || '应用 YAML 失败' })
  } finally {
    savingYaml.value = false
  }
}
</script>

<style scoped>
/* ========== 基础布局 ========== */
.resource-view {
  max-width: 1600px;
  margin: 0 auto;
  padding: 0 1rem;
}

/* ========== 页面头部 - Rancher 风格 ========== */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 1rem;
  color: white;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.header-icon {
  font-size: 2.5rem;
  background: rgba(255, 255, 255, 0.2);
  padding: 0.75rem;
  border-radius: 0.75rem;
}

.header-text h1 {
  margin: 0 0 0.25rem 0;
  font-size: 1.75rem;
  font-weight: 700;
}

.header-text p {
  margin: 0;
  font-size: 0.875rem;
  opacity: 0.9;
}

.header-stats {
  display: flex;
  gap: 1.5rem;
}

.stat-item {
  text-align: center;
  padding: 0.75rem 1.25rem;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 0.75rem;
  backdrop-filter: blur(10px);
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
}

.stat-label {
  font-size: 0.75rem;
  opacity: 0.9;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* ========== 操作栏 ========== */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.action-bar-left,
.action-bar-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

/* 自动刷新开关 */
.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  cursor: pointer;
  color: #4a5568;
}

.auto-refresh-toggle:hover {
  color: #326ce5;
}

.auto-refresh-toggle input[type="checkbox"] {
  cursor: pointer;
}

.refresh-indicator {
  color: #34d399;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  font-size: 1rem;
}

.search-box input {
  padding: 0.625rem 0.75rem 0.625rem 2.25rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  width: 16rem;
  transition: all 0.2s;
}

.search-box input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.filter-buttons {
  display: flex;
  gap: 0.25rem;
  background: #f7fafc;
  padding: 0.25rem;
  border-radius: 0.5rem;
}

.btn-filter {
  padding: 0.5rem 1rem;
  border: none;
  background: transparent;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-filter:hover {
  background: #e2e8f0;
}

.btn-filter.active {
  background: white;
  color: #667eea;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.view-toggle {
  display: flex;
  gap: 0.25rem;
  background: #f7fafc;
  padding: 0.25rem;
  border-radius: 0.5rem;
}

.btn-view {
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  border-radius: 0.375rem;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-view:hover {
  background: #e2e8f0;
}

.btn-view.active {
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.btn {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ========== 表格视图 ========== */
.table-container {
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.resource-table th {
  background: #f8fafc;
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.resource-table td {
  padding: 1rem;
  font-size: 0.875rem;
  color: #1e293b;
  border-bottom: 1px solid #f1f5f9;
}

.table-row:hover {
  background: #f8fafc;
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.status-badge.active .status-dot {
  background: #22c55e;
  box-shadow: 0 0 8px #22c55e;
}

.status-badge.terminating {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.status-badge.terminating .status-dot {
  background: #ef4444;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 命名空间信息 */
.namespace-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.namespace-icon {
  font-size: 1.125rem;
}

.namespace-name {
  font-weight: 500;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.system-tag {
  font-size: 0.625rem;
  padding: 0.125rem 0.5rem;
  background: #fef3c7;
  color: #92400e;
  border-radius: 0.25rem;
  font-weight: 600;
}

/* 数量徽章 */
.count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  padding: 0.25rem 0.625rem;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 600;
}

.count-badge.pods {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

.count-badge.services {
  background: rgba(168, 85, 247, 0.1);
  color: #9333ea;
}

.count-badge.deployments {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

/* 资源使用条 */
.resource-usage {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.usage-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.usage-icon {
  font-size: 0.75rem;
}

.usage-bar-wrap {
  flex: 1;
  max-width: 60px;
}

.usage-bar {
  height: 0.375rem;
  background: #e2e8f0;
  border-radius: 0.25rem;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  border-radius: 0.25rem;
  transition: width 0.3s;
}

.usage-fill.low { background: #22c55e; }
.usage-fill.medium { background: #f59e0b; }
.usage-fill.high { background: #ef4444; }

.usage-text {
  font-size: 0.75rem;
  color: #64748b;
  min-width: 2rem;
}

.date-text {
  font-size: 0.8125rem;
  color: #64748b;
  font-family: monospace;
}

/* 操作按钮 */
.action-buttons-inline {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background: #f1f5f9;
  color: #475569;
  white-space: nowrap;
}

.action-btn:hover:not(:disabled) {
  background: #e2e8f0;
}

.action-btn.danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.action-btn.danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ========== 卡片视图 ========== */
.cards-container {
  min-height: 400px;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.namespace-card {
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s;
  border: 1px solid transparent;
}

.namespace-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  border-color: #667eea;
}

.namespace-card.system-namespace {
  border-left: 3px solid #f59e0b;
}

.card-header {
  padding: 1rem;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 1px solid #e2e8f0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.card-icon {
  font-size: 1.25rem;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-badges {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* 卡片统计 */
.card-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
  padding: 1rem;
  background: #f8fafc;
}

.stat-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem;
  background: white;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
  transition: all 0.2s;
}

.stat-box.clickable {
  cursor: pointer;
}

.stat-box.clickable:hover {
  border-color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.stat-box.clickable:hover .stat-number {
  color: #667eea;
}

.stat-icon {
  font-size: 1.25rem;
  padding: 0.375rem;
  border-radius: 0.375rem;
}

.stat-icon.pods { background: rgba(59, 130, 246, 0.1); }
.stat-icon.services { background: rgba(168, 85, 247, 0.1); }
.stat-icon.deployments { background: rgba(34, 197, 94, 0.1); }

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-number {
  font-size: 1.125rem;
  font-weight: 700;
  color: #1e293b;
}

.stat-name {
  font-size: 0.625rem;
  color: #64748b;
  text-transform: uppercase;
}

/* 卡片内容 */
.card-section {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid #f1f5f9;
}

.section-label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.section-value {
  font-size: 0.875rem;
  color: #1e293b;
}

.resource-bars {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.resource-bar-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.bar-label {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: #64748b;
}

.bar-track {
  height: 0.375rem;
  background: #e2e8f0;
  border-radius: 0.25rem;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 0.25rem;
  transition: width 0.3s;
}

.bar-fill.low { background: linear-gradient(90deg, #22c55e, #4ade80); }
.bar-fill.medium { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.bar-fill.high { background: linear-gradient(90deg, #ef4444, #f87171); }

/* 卡片底部按钮 */
.card-footer {
  display: flex;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  background: #f8fafc;
}

.card-action-btn {
  flex: 1;
  padding: 0.5rem;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
  color: #475569;
  border: 1px solid #e2e8f0;
  white-space: nowrap;
}

.card-action-btn:hover:not(:disabled) {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.card-action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
}

.card-action-btn.primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.4);
}

.card-action-btn.danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  border-color: rgba(239, 68, 68, 0.2);
}

.card-action-btn.danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.card-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ========== 模态框 ========== */
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
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
  width: 90%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-detail {
  max-width: 600px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header.danger {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #64748b;
  padding: 0.25rem;
  line-height: 1;
  border-radius: 0.25rem;
}

.modal-close:hover {
  background: #f1f5f9;
}

.modal-body {
  padding: 1.5rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

/* 表单样式 */
.form-section {
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
}

.form-group input {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-hint {
  margin-top: 0.375rem;
  font-size: 0.75rem;
  color: #6b7280;
}

.quota-hint {
  display: flex;
  gap: 0.75rem;
  padding: 1rem;
  background: #f0f9ff;
  border-radius: 0.5rem;
  border: 1px solid #bae6fd;
}

.hint-icon {
  font-size: 1.25rem;
}

.hint-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #0369a1;
  margin-bottom: 0.25rem;
}

.hint-text {
  font-size: 0.8125rem;
  color: #0c4a6e;
}

/* 详情卡片 */
.detail-card {
  background: #f8fafc;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  overflow: hidden;
}

.detail-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  font-weight: 600;
  font-size: 0.875rem;
  color: #1e293b;
}

.header-hint {
  margin-left: auto;
  font-size: 0.6875rem;
  font-weight: 400;
  color: #94a3b8;
}

.detail-card-icon {
  font-size: 1rem;
}

.detail-card-body {
  padding: 1rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  border-bottom: 1px solid #e2e8f0;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 0.8125rem;
  color: #64748b;
}

.detail-value {
  font-size: 0.875rem;
  color: #1e293b;
  font-weight: 500;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}

.stats-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.75rem;
  background: white;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
  transition: all 0.2s;
}

.stats-item.clickable {
  cursor: pointer;
  position: relative;
}

.stats-item.clickable:hover {
  border-color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.stats-item.clickable:hover .stats-number {
  color: #667eea;
}

.stats-arrow {
  margin-left: auto;
  font-size: 1rem;
  color: #cbd5e1;
  transition: all 0.2s;
}

.stats-item.clickable:hover .stats-arrow {
  color: #667eea;
  transform: translateX(3px);
}

.stats-icon {
  font-size: 1.5rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
}

.stats-icon.pods { background: rgba(59, 130, 246, 0.1); }
.stats-icon.services { background: rgba(168, 85, 247, 0.1); }
.stats-icon.deployments { background: rgba(34, 197, 94, 0.1); }

.stats-number {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1e293b;
}

.stats-label {
  font-size: 0.6875rem;
  color: #64748b;
  text-transform: uppercase;
}

.section-subtitle {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 0.5rem;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.label-tag,
.annotation-tag {
  padding: 0.25rem 0.5rem;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  font-size: 0.6875rem;
  font-family: monospace;
  color: #475569;
}

.labels-section,
.annotations-section {
  margin-bottom: 0.75rem;
}

.annotations-section {
  margin-bottom: 0;
}

/* 删除警告 */
.delete-warning {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  background: #fef2f2;
  border-radius: 0.5rem;
  border: 1px solid #fecaca;
  margin-bottom: 1rem;
}

.warning-icon {
  font-size: 2rem;
}

.warning-title {
  font-size: 1rem;
  font-weight: 600;
  color: #991b1b;
  margin-bottom: 0.5rem;
}

.warning-text {
  font-size: 0.875rem;
  color: #7f1d1d;
  margin-bottom: 0.5rem;
}

.warning-list {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.8125rem;
  color: #991b1b;
}

.warning-list li {
  margin-bottom: 0.25rem;
}

.delete-target {
  padding: 1rem;
  background: #f8fafc;
  border-radius: 0.5rem;
  text-align: center;
}

.target-label {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.375rem;
}

.target-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #dc2626;
  margin-bottom: 0.5rem;
}

.target-stats {
  display: flex;
  justify-content: center;
  gap: 1rem;
  font-size: 0.75rem;
  color: #64748b;
}

/* ========== 空状态 & 加载状态 ========== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  color: #64748b;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-text {
  font-size: 1rem;
  margin-bottom: 1rem;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
}

.loading-spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid #e2e8f0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: #64748b;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ========== 响应式 ========== */
@media (max-width: 1024px) {
  .view-header {
    flex-direction: column;
    gap: 1rem;
  }
  
  .header-stats {
    width: 100%;
    justify-content: flex-start;
  }
  
  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .action-bar-left,
  .action-bar-right {
    justify-content: space-between;
  }
  
  .search-box input {
    width: 100%;
  }
  
  .cards-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  }
}

@media (max-width: 768px) {
  .filter-buttons {
    display: none;
  }
  
  .cards-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .card-stats {
    grid-template-columns: 1fr;
  }
}

/* ========== 批量操作样式 ========== */
.btn-batch {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
}

.btn-batch:hover {
  opacity: 0.9;
}

.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
  font-size: 14px;
}

.batch-clear {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.batch-clear:hover {
  background: rgba(255,255,255,0.3);
}

.batch-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.batch-btn {
  padding: 8px 16px;
  background: white;
  color: #4a5568;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: #f7fafc;
}

.batch-btn.danger {
  background: #fed7d7;
  color: #c53030;
}

.batch-btn.danger:hover {
  background: #feb2b2;
}

.row-selected {
  background: rgba(102, 126, 234, 0.1) !important;
}

/* 批量预览弹窗 */
.modal-batch-preview {
  max-width: 700px;
}

.preview-section {
  margin-bottom: 20px;
}

.preview-section .section-title {
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 12px;
  font-size: 14px;
}

.affected-namespaces-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.affected-ns-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7fafc;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.ns-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ns-info .ns-name {
  font-weight: 600;
  color: #2d3748;
}

.ns-stats {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.stat-tag {
  font-size: 11px;
  color: #718096;
  background: #edf2f7;
  padding: 2px 8px;
  border-radius: 4px;
}

.modal-danger .danger-header {
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
}

.danger-warning {
  display: flex;
  gap: 16px;
  background: #fff5f5;
  border: 1px solid #feb2b2;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.warning-icon-large {
  font-size: 32px;
}

.warning-title {
  font-weight: 600;
  color: #c53030;
  margin-bottom: 8px;
}

.warning-list {
  margin: 0;
  padding-left: 20px;
  color: #742a2a;
  font-size: 13px;
}

.warning-list li {
  margin-bottom: 4px;
}

.confirm-section {
  margin-top: 20px;
}

.confirm-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 16px;
  text-align: center;
  font-weight: 600;
  letter-spacing: 2px;
}

.confirm-input:focus {
  outline: none;
  border-color: #fc8181;
}

.confirm-input.valid {
  border-color: #48bb78;
  background: #f0fff4;
}

/* 卡片批量选择样式 */
.card-checkbox {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 10;
}

.card-checkbox input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
  accent-color: #667eea;
}

.namespace-card {
  position: relative;
}

.card-selected {
  border: 2px solid #667eea !important;
  background: rgba(102, 126, 234, 0.05) !important;
}

/* ========== YAML 模态框样式 ========== */
.yaml-modal {
  width: 90%;
  max-width: 1000px;
  max-height: 90vh;
}

.yaml-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.yaml-modal-body {
  padding: 0;
  max-height: 60vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.yaml-editor-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 400px;
  max-height: 60vh;
}

.yaml-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
  padding: 16px;
  border: none;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
}

.yaml-editor:focus {
  outline: none;
}

.yaml-content {
  margin: 0;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  background: #f7f8fa;
  color: #2d3748;
  overflow: auto;
  min-height: 400px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.error-box {
  padding: 16px;
  background: #fff5f5;
  color: #c53030;
  border-radius: 8px;
  margin: 16px;
}
</style>

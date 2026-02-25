<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>密钥管理</h1>
      <p>Kubernetes集群中的Secret列表</p>
    </div>
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索密钥名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: typeFilter === 'all' }" @click="setTypeFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'Opaque' }" @click="setTypeFilter('Opaque')">
          Opaque
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'kubernetes.io/tls' }" @click="setTypeFilter('kubernetes.io/tls')">
          TLS
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'kubernetes.io/dockerconfigjson' }" @click="setTypeFilter('kubernetes.io/dockerconfigjson')">
          Docker
        </button>
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
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

        <!-- 视图切换按钮 -->
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
        
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        <button class="btn btn-primary" @click="showCreateModal = true">创建Secret</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedSecrets.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedSecrets.length }} 个 Secret</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 数据统计信息 -->
    <div v-if="!loading && secrets.length > 0" class="stats-bar">
      <div class="stat-item">
        <span class="stat-label">🔐 Opaque:</span>
        <span class="stat-value">{{ getTypeCount('Opaque') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">🔒 TLS:</span>
        <span class="stat-value">{{ getTypeCount('kubernetes.io/tls') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">🐳 Docker:</span>
        <span class="stat-value">{{ getTypeCount('kubernetes.io/dockerconfigjson') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">当前页:</span>
        <span class="stat-value">{{ paginatedSecrets.length }}</span>
      </div>
    </div>

    <!-- 表格视图 -->
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
              />
            </th>
            <th style="width: 100px;">类型</th>
            <th style="min-width: 200px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="width: 100px;">数据项</th>
            <th style="min-width: 180px;">标签</th>
            <th style="width: 170px; white-space: nowrap;">创建时间</th>
            <th style="width: 180px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="secret in paginatedSecrets" :key="secret.name + secret.namespace" :class="{ 'row-selected': isSecretSelected(secret) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isSecretSelected(secret)" 
                @change="toggleSecretSelection(secret)"
              />
            </td>
            <td>
              <span class="type-badge" :class="getTypeBadgeClass(secret.type)">
                {{ getTypeShortName(secret.type) }}
              </span>
            </td>
            <td>
              <div class="secret-name">
                <span class="icon">🔑</span>
                <span>{{ secret.name }}</span>
                <span class="age-badge" :title="secret.created_at">{{ getAge(secret.created_at) }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ secret.namespace }}</span>
            </td>
            <td>
              <span class="data-count">{{ secret.data_count || 0 }} 项</span>
            </td>
            <td>
              <div class="label-tags">
                <span v-for="(value, key) in getLimitedLabels(secret.labels)" :key="key" class="label-tag" :title="`${key}=${value}`">
                  {{ key }}={{ value }}
                </span>
                <span v-if="Object.keys(secret.labels || {}).length === 0" class="label-empty">-</span>
                <span v-if="Object.keys(secret.labels || {}).length > 3" class="label-more">+{{ Object.keys(secret.labels).length - 3 }}</span>
              </div>
            </td>
            <td style="white-space: nowrap;">{{ secret.created_at }}</td>
            <td>
              <div class="action-icons">
                <button class="action-btn icon-only" @click="viewSecret(secret)" title="查看详情">
                  📋
                </button>
                <button class="action-btn icon-only" @click="decodeSecret(secret)" title="解码数据">
                  🔓
                </button>
                <button class="action-btn icon-only" @click="viewYaml(secret)" title="查看YAML">
                  📝
                </button>
                <button class="action-btn icon-only danger" @click="confirmDelete(secret)" title="删除">
                  🗑️
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-if="!loading && filteredSecrets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p class="empty-title">暂无Secret数据</p>
        <p class="empty-desc">当前筛选条件下没有找到Secret，试试调整筛选条件或创建新的Secret</p>
        <button class="btn btn-primary" @click="showCreateModal = true" style="margin-top: 16px;">
          创建第一个Secret
        </button>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="card-grid">
      <div v-for="secret in paginatedSecrets" :key="secret.name + secret.namespace" class="resource-card" :class="{ 'card-selected': isSecretSelected(secret) }">
        <!-- 批量选择复选框 -->
        <div v-if="batchMode" class="card-checkbox">
          <input 
            type="checkbox" 
            :checked="isSecretSelected(secret)" 
            @change="toggleSecretSelection(secret)"
          />
        </div>

        <div class="card-header" :class="`type-${getTypeBadgeClass(secret.type)}`">
          <div class="card-title">
            <span class="card-icon">🔑</span>
            <h3>{{ secret.name }}</h3>
          </div>
          <span class="type-badge" :class="getTypeBadgeClass(secret.type)">
            {{ getTypeShortName(secret.type) }}
          </span>
        </div>

        <div class="card-body">
          <div class="card-meta">
            <div class="card-meta-item">
              <span class="label">命名空间:</span>
              <span class="namespace-badge">{{ secret.namespace }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">数据项:</span>
              <span>{{ secret.data_count || 0 }} 项</span>
            </div>
            <div class="card-meta-item">
              <span class="label">类型:</span>
              <span>{{ secret.type }}</span>
            </div>
            <div class="card-meta-item">
              <span class="label">创建时间:</span>
              <span>{{ secret.created_at }}</span>
            </div>
          </div>
        </div>

        <div class="card-footer">
          <button class="card-btn primary" @click="viewSecret(secret)">📋 详情</button>
          <button class="card-btn" @click="decodeSecret(secret)">🔓 解码</button>
          <button class="card-btn danger" @click="confirmDelete(secret)">🗑️ 删除</button>
        </div>
      </div>

      <div v-if="loading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-if="!loading && filteredSecrets.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p class="empty-title">暂无Secret数据</p>
        <p class="empty-desc">当前筛选条件下没有找到Secret，试试调整筛选条件或创建新的Secret</p>
        <button class="btn btn-primary" @click="showCreateModal = true" style="margin-top: 16px;">
          创建第一个Secret
        </button>
      </div>
    </div>

    <!-- 分页（现代化三段式布局） -->
    <div v-if="filteredSecrets.length > 0" class="pagination-wrapper">
      <div class="pagination-left">
        <span class="pagination-summary">共 <strong>{{ filteredSecrets.length }}</strong> 条</span>
      </div>
      <div class="pagination-center">
        <button 
          class="pagination-btn" 
          @click="goToPage(1)" 
          :disabled="currentPage === 1"
          title="首页"
        >
          «
        </button>
        <button 
          class="pagination-btn" 
          @click="goToPage(currentPage - 1)" 
          :disabled="currentPage === 1"
          title="上一页"
        >
          ‹
        </button>
        
        <!-- 页码按钮组 -->
        <template v-for="page in visiblePages" :key="page">
          <button 
            v-if="typeof page === 'number'"
            class="pagination-btn page-number" 
            :class="{ active: currentPage === page }"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>
          <span v-else class="pagination-ellipsis">...</span>
        </template>
        
        <button 
          class="pagination-btn" 
          @click="goToPage(currentPage + 1)" 
          :disabled="currentPage === totalPages"
          title="下一页"
        >
          ›
        </button>
        <button 
          class="pagination-btn" 
          @click="goToPage(totalPages)" 
          :disabled="currentPage === totalPages"
          title="尾页"
        >
          »
        </button>
      </div>
      <div class="pagination-right">
        <select v-model.number="itemsPerPage" @change="onPageSizeChange" class="page-size-select">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
          <option :value="100">100 条/页</option>
        </select>
        <span class="pagination-goto">前往</span>
        <input 
          v-model.number="jumpPage" 
          type="number" 
          min="1" 
          :max="totalPages" 
          class="page-jump-input" 
          @keyup.enter="jumpToPage" 
        />
      </div>
    </div>

    <!-- 创建Secret模态框（Rancher/Kuboard风格） -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content modal-create-secret" @click.stop>
        <div class="modal-header">
          <h2>🔑 创建 Secret</h2>
          <!-- 视图切换按钮 -->
          <div class="view-toggle-buttons">
            <button 
              class="view-toggle-btn" 
              :class="{ active: createMode === 'form' }"
              @click="createMode = 'form'"
              title="表单模式"
            >
              📝 表单
            </button>
            <button 
              class="view-toggle-btn" 
              :class="{ active: createMode === 'yaml' }"
              @click="createMode = 'yaml'"
              title="YAML模式"
            >
              📄 YAML
            </button>
          </div>
          <button @click="closeCreateModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
            <form @submit.prevent="createSecret">
              <div class="form-section">
                <div class="section-header">
                  <span class="section-icon">📋</span>
                  <h3>基本信息</h3>
                </div>
                <div class="form-row">
                  <div class="form-group">
                    <label class="required">命名空间</label>
                    <select v-model="secretForm.namespace" required>
                      <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                    </select>
                  </div>
                  <div class="form-group">
                    <label class="required">Secret 名称</label>
                    <input type="text" v-model="secretForm.name" required placeholder="例如: my-secret" />
                  </div>
                </div>
                <div class="form-group">
                  <label class="required">类型</label>
                  <select v-model="secretForm.type" required>
                    <option value="Opaque">Opaque（通用密钥）</option>
                    <option value="kubernetes.io/tls">kubernetes.io/tls（TLS证书）</option>
                    <option value="kubernetes.io/dockerconfigjson">kubernetes.io/dockerconfigjson（Docker配置）</option>
                    <option value="kubernetes.io/basic-auth">kubernetes.io/basic-auth（基本认证）</option>
                    <option value="kubernetes.io/ssh-auth">kubernetes.io/ssh-auth（SSH认证）</option>
                  </select>
                </div>
              </div>
              
              <div class="form-section">
                <div class="section-header">
                  <span class="section-icon">🔐</span>
                  <h3>数据项</h3>
                  <span class="section-tip">输入明文值，系统自动进行 Base64 编码</span>
                </div>
                <div class="data-items-container">
                  <div v-for="(item, index) in secretForm.dataItems" :key="index" class="data-item-row">
                    <div class="data-item-fields">
                      <input type="text" v-model="item.key" placeholder="键名（如: username）" class="data-key" />
                      <textarea v-model="item.value" placeholder="值（明文）" class="data-value" rows="2"></textarea>
                    </div>
                    <button type="button" class="remove-btn" @click="removeDataItem(index)" title="删除此项">
                      🗑️
                    </button>
                  </div>
                  <button type="button" class="add-btn" @click="addDataItem">
                    ➕ 添加数据项
                  </button>
                </div>
              </div>
            </form>
          </div>
          
          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-create-section">
            <div class="yaml-toolbar">
              <button type="button" class="btn btn-sm btn-secondary" @click="loadSecretYamlTemplate">
                📋 加载模板
              </button>
              <button type="button" class="btn btn-sm btn-secondary" @click="clearYamlContent">
                🗑️ 清空
              </button>
              <button type="button" class="btn btn-sm btn-secondary" @click="formatYaml">
                ✨ 格式化
              </button>
            </div>
            
            <div class="yaml-editor-container">
              <div class="yaml-editor-wrapper">
                <div class="editor-label">YAML 内容</div>
                <textarea 
                  v-model="createYamlContent" 
                  class="yaml-editor create-yaml-editor"
                  spellcheck="false"
                  placeholder="apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
stringData:
  username: admin
  password: secret123"
                ></textarea>
              </div>
              
              <div class="yaml-preview-wrapper">
                <div class="editor-label">预览</div>
                <div class="yaml-preview">
                  <pre v-if="yamlPreviewContent">{{ yamlPreviewContent }}</pre>
                  <div v-else class="preview-placeholder">在左侧输入 YAML 内容后显示预览</div>
                </div>
              </div>
            </div>
            
            <div v-if="createYamlError" class="yaml-error">
              <span class="error-icon">⚠️</span>
              {{ createYamlError }}
            </div>
            
            <div class="yaml-editor-footer">
              <div class="yaml-tips">
                <strong>💡 提示：</strong>
                <ul>
                  <li>✅ 使用 <code>stringData</code> 字段输入明文数据，K8s 会自动进行 Base64 编码</li>
                  <li>🔗 支持多资源 YAML（用 <code>---</code> 分隔）</li>
                  <li>📦 可同时创建：ConfigMap + Secret 等组合资源</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button type="button" class="btn btn-secondary" @click="closeCreateModal">取消</button>
            <button type="button" class="btn btn-primary" @click="createSecret" :disabled="creating">
              <span class="btn-icon">🚀</span>
              {{ creating ? '创建中...' : '创建 Secret' }}
            </button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" class="btn btn-secondary" @click="closeCreateModal">取消</button>
            <button 
              type="button" 
              class="btn btn-primary" 
              @click="createSecretFromYaml" 
              :disabled="creating || !createYamlContent.trim()"
            >
              <span class="btn-icon">🚀</span>
              {{ creating ? '创建中...' : '从 YAML 创建' }}
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content modal-danger" @click.stop>
        <div class="modal-header">
          <h2>🗑️ 确认删除</h2>
          <button @click="showDeleteModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="danger-warning">
            <p>确定要删除 Secret <strong>{{ selectedSecret?.name }}</strong> 吗？</p>
            <p class="warning-text">此操作不可撤销！</p>
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
            <button type="button" class="btn btn-danger" @click="deleteSecret" :disabled="deleting">
              {{ deleting ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量删除确认模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click="showBatchDeleteModal = false">
      <div class="modal-content modal-danger" @click.stop>
        <div class="modal-header">
          <h2>🗑️ 批量删除确认</h2>
          <button @click="showBatchDeleteModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="danger-warning">
            <p>将删除以下 {{ selectedSecrets.length }} 个 Secret：</p>
            <ul class="delete-list">
              <li v-for="s in selectedSecrets" :key="s.name + s.namespace">
                {{ s.namespace }}/{{ s.name }}
              </li>
            </ul>
            <p class="warning-text">此操作不可撤销！请输入 DELETE 确认：</p>
            <input type="text" v-model="deleteConfirmText" placeholder="输入 DELETE 确认" class="confirm-input" />
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-secondary" @click="showBatchDeleteModal = false">取消</button>
            <button 
              type="button" 
              class="btn btn-danger" 
              @click="executeBatchDelete" 
              :disabled="deleteConfirmText !== 'DELETE' || batchDeleting"
            >
              {{ batchDeleting ? '删除中...' : '确认批量删除' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="showDetailModal = false">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h2>📋 Secret 详情 - {{ detailData?.name }}</h2>
          <button @click="showDetailModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDetail" class="loading-state">加载中...</div>
          <div v-else-if="detailData" class="detail-content">
            <div class="detail-section">
              <h4>基本信息</h4>
              <table class="detail-table">
                <tbody>
                  <tr>
                    <td class="label">名称:</td>
                    <td>{{ detailData.name }}</td>
                  </tr>
                  <tr>
                    <td class="label">命名空间:</td>
                    <td>{{ detailData.namespace }}</td>
                  </tr>
                  <tr>
                    <td class="label">类型:</td>
                    <td>{{ detailData.type }}</td>
                  </tr>
                  <tr>
                    <td class="label">数据项数:</td>
                    <td>{{ detailData.data_count || Object.keys(detailData.data || {}).length }}</td>
                  </tr>
                  <tr>
                    <td class="label">创建时间:</td>
                    <td>{{ detailData.created_at }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="detail-section" v-if="detailData.labels && Object.keys(detailData.labels).length > 0">
              <h4>标签</h4>
              <div class="label-list">
                <span v-for="(value, key) in detailData.labels" :key="key" class="label-tag">
                  {{ key }}={{ value }}
                </span>
              </div>
            </div>
            <div class="detail-section" v-if="detailData.data && Object.keys(detailData.data).length > 0">
              <h4>数据项（Base64编码）</h4>
              <table class="detail-table data-table">
                <tbody>
                  <tr v-for="(value, key) in detailData.data" :key="key">
                    <td class="label">{{ key }}:</td>
                    <td class="data-value-cell">
                      <code>{{ truncateValue(value) }}</code>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDetailModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 解码数据模态框 -->
    <div v-if="showDecodeModal" class="modal-overlay" @click="showDecodeModal = false">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h2>🔓 Secret 解码数据 - {{ decodedSecret?.name }}</h2>
          <button @click="showDecodeModal = false" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingDecode" class="loading-state">解码中...</div>
          <div v-else-if="decodedData" class="decode-content">
            <div class="decode-warning">
              <span class="warning-icon">⚠️</span>
              <span>以下为解码后的明文数据，请妥善保管</span>
            </div>
            <table class="detail-table data-table">
              <tbody>
                <tr v-for="(value, key) in decodedData" :key="key">
                  <td class="label">{{ key }}:</td>
                  <td class="data-value-cell">
                    <pre class="decoded-value">{{ value }}</pre>
                    <button class="copy-btn" @click="copyToClipboard(value)" title="复制">📋</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state">
            <p>暂无数据</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDecodeModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框（增强版） -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content yaml-modal" @click.stop>
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlSecret?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml" :disabled="loadingYaml">
              📄 下载
            </button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true" :disabled="loadingYaml">
              ✈️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="close-btn" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="selectedYamlSecret" class="info-box" style="margin-bottom: 16px;">
            <div><strong>Secret:</strong> {{ selectedYamlSecret.name }}</div>
            <div><strong>命名空间:</strong> {{ selectedYamlSecret.namespace }}</div>
          </div>
          
          <div v-if="loadingYaml" class="loading-state">
            <div class="loading-spinner"></div>
            <div class="loading-text">Loading YAML...</div>
          </div>
          <div v-else-if="yamlViewError" class="error-box">{{ yamlViewError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea 
              v-if="yamlEditMode" 
              v-model="yamlContent" 
              class="yaml-editor" 
              spellcheck="false"
            ></textarea>
            <pre v-else class="yaml-content-pre">{{ yamlContent }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">取消</button>
          <button 
            v-if="yamlEditMode" 
            class="btn btn-primary" 
            @click="applyYamlChanges" 
            :disabled="savingYaml"
          >
            {{ savingYaml ? '保存中...' : '应用更改' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import secretApi from '@/api/cluster/config/secret'
import namespaceApi from '@/api/cluster/config/namespace'

// ==================== 状态管理 ====================
const loading = ref(false)
const errorMsg = ref('')
const secrets = ref([])
const total = ref(0)

// 视图模式
const viewMode = ref('table') // 'table' | 'card'

// 批量操作
const batchMode = ref(false)
const selectedSecrets = ref([])
const showBatchDeleteModal = ref(false)
const deleteConfirmText = ref('')
const batchDeleting = ref(false)

// 搜索和过滤
const searchQuery = ref('')
const namespaceFilter = ref('')
const typeFilter = ref('all')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)
const jumpPage = ref(1)

// 自动刷新
const autoRefresh = ref(false)
let refreshTimer = null

// 模态框状态
const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const showDetailModal = ref(false)
const showDecodeModal = ref(false)
const showYamlModal = ref(false)

// 操作状态
const creating = ref(false)
const deleting = ref(false)
const loadingDetail = ref(false)
const loadingDecode = ref(false)
const loadingYaml = ref(false)

// 选中的 Secret
const selectedSecret = ref(null)
const detailData = ref(null)
const decodedSecret = ref(null)
const decodedData = ref(null)
const selectedYamlSecret = ref(null)
const yamlContent = ref('')
const yamlEditMode = ref(false)
const yamlViewError = ref('')
const savingYaml = ref(false)

// YAML 创建相关
const createMode = ref('form') // 'form' | 'yaml'
const createYamlContent = ref('')
const createYamlError = ref('')
const yamlPreviewContent = ref('')

// 命名空间列表
const namespaces = ref(['default', 'kube-system', 'kube-public'])

// 创建表单
const secretForm = ref({
  namespace: 'default',
  name: '',
  type: 'Opaque',
  dataItems: [{ key: '', value: '' }]
})

// ==================== 计算属性 ====================
const filteredSecrets = computed(() => {
  let result = secrets.value

  // 命名空间过滤
  if (namespaceFilter.value) {
    result = result.filter(s => s.namespace === namespaceFilter.value)
  }

  // 类型过滤
  if (typeFilter.value !== 'all') {
    result = result.filter(s => s.type === typeFilter.value)
  }

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(query) ||
      s.namespace.toLowerCase().includes(query)
    )
  }

  return result
})

const paginatedSecrets = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredSecrets.value.slice(startIndex, endIndex)
})

const isAllSelected = computed(() => {
  return paginatedSecrets.value.length > 0 && 
         paginatedSecrets.value.every(s => isSecretSelected(s))
})

// 分页相关计算属性
const totalPages = computed(() => Math.ceil(filteredSecrets.value.length / itemsPerPage.value) || 1)

// 智能页码显示
const visiblePages = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  const pages = []

  if (total <= 7) {
    // 总页数 <= 7，全部显示
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // 总页数 > 7，智能显示
    if (current <= 4) {
      // 当前页在前面：1 2 3 4 5 ... total
      for (let i = 1; i <= 5; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 3) {
      // 当前页在后面：1 ... total-4 total-3 total-2 total-1 total
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      // 当前页在中间：1 ... current-1 current current+1 ... total
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    }
  }

  return pages
})

// ==================== 方法 ====================
// 获取 Secret 列表
const fetchSecrets = async () => {
  try {
    loading.value = true
    errorMsg.value = ''

    const res = await secretApi.list({
      namespace: '',  // 查询所有命名空间
      page: 1,
      limit: 1000  // 前端分页
    })

    if (res.code === 0 && res.data) {
      const list = res.data.list || []
      secrets.value = list.map(item => ({
        name: item.name,
        namespace: item.namespace,
        type: item.type,
        data_count: item.data_count || 0,
        labels: item.labels || {},
        created_at: item.created_at
      }))
      total.value = res.data.total || list.length
    }
  } catch (error) {
    console.error('获取Secret列表失败:', error)
    errorMsg.value = error.kube_message_error || error.message || '获取Secret列表失败'
  } finally {
    loading.value = false
  }
}

// 获取命名空间列表
const fetchNamespaces = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
    
    if (namespaces.value.length === 0) {
      namespaces.value = ['default', 'kube-system', 'kube-public']
    }
  } catch (e) {
    console.error('获取命名空间列表失败:', e)
    namespaces.value = ['default', 'kube-system', 'kube-public']
  }
}

// 刷新列表
const refreshList = () => {
  fetchSecrets()
}

// 搜索输入处理
const onSearchInput = () => {
  currentPage.value = 1
}

// 分页大小改变
const onPageSizeChange = () => {
  currentPage.value = 1
}

// 分页跳转方法
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

// 设置类型过滤
const setTypeFilter = (type) => {
  typeFilter.value = type
  currentPage.value = 1
}

// 获取类型统计
const getTypeCount = (type) => {
  return secrets.value.filter(s => s.type === type).length
}

// 格式化时间（相对时间）
const getAge = (timeStr) => {
  if (!timeStr) return '-'
  try {
    const time = new Date(timeStr)
    const now = new Date()
    const diff = now - time
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days}d`
    if (hours > 0) return `${hours}h`
    if (minutes > 0) return `${minutes}m`
    return '刚刚'
  } catch {
    return timeStr
  }
}

// 获取类型简称
const getTypeShortName = (type) => {
  const typeMap = {
    'Opaque': 'Opaque',
    'kubernetes.io/tls': 'TLS',
    'kubernetes.io/dockerconfigjson': 'Docker',
    'kubernetes.io/service-account-token': 'SA Token',
    'kubernetes.io/basic-auth': 'Basic Auth',
    'kubernetes.io/ssh-auth': 'SSH Auth'
  }
  return typeMap[type] || type
}

// 获取类型样式类
const getTypeBadgeClass = (type) => {
  const classMap = {
    'Opaque': 'opaque',
    'kubernetes.io/tls': 'tls',
    'kubernetes.io/dockerconfigjson': 'docker',
    'kubernetes.io/service-account-token': 'sa-token'
  }
  return classMap[type] || 'default'
}

// 获取有限的标签
const getLimitedLabels = (labels) => {
  if (!labels) return {}
  const entries = Object.entries(labels).slice(0, 3)
  return Object.fromEntries(entries)
}

// 截断值
const truncateValue = (value) => {
  if (!value) return ''
  return value.length > 50 ? value.substring(0, 50) + '...' : value
}

// ==================== 批量操作 ====================
const enterBatchMode = () => {
  batchMode.value = true
  selectedSecrets.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedSecrets.value = []
}

const clearSelection = () => {
  selectedSecrets.value = []
}

const isSecretSelected = (secret) => {
  return selectedSecrets.value.some(s => s.name === secret.name && s.namespace === secret.namespace)
}

const toggleSecretSelection = (secret) => {
  const index = selectedSecrets.value.findIndex(s => s.name === secret.name && s.namespace === secret.namespace)
  if (index >= 0) {
    selectedSecrets.value.splice(index, 1)
  } else {
    selectedSecrets.value.push(secret)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedSecrets.value.forEach(secret => {
      const index = selectedSecrets.value.findIndex(s => s.name === secret.name && s.namespace === secret.namespace)
      if (index >= 0) selectedSecrets.value.splice(index, 1)
    })
  } else {
    paginatedSecrets.value.forEach(secret => {
      if (!isSecretSelected(secret)) {
        selectedSecrets.value.push(secret)
      }
    })
  }
}

const openBatchDeletePreview = () => {
  deleteConfirmText.value = ''
  showBatchDeleteModal.value = true
}

const executeBatchDelete = async () => {
  if (deleteConfirmText.value !== 'DELETE') return
  
  batchDeleting.value = true
  let successCount = 0
  let failCount = 0
  
  for (const secret of selectedSecrets.value) {
    try {
      await secretApi.delete({
        name: secret.name,
        namespace: secret.namespace
      })
      successCount++
    } catch (e) {
      console.error(`Failed to delete ${secret.name}:`, e)
      failCount++
    }
  }
  
  batchDeleting.value = false
  showBatchDeleteModal.value = false
  
  if (successCount > 0) {
    alert(`成功删除 ${successCount} 个 Secret${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
  }
  
  exitBatchMode()
  fetchSecrets()
}

// ==================== CRUD 操作 ====================
// 创建 Secret
const createSecret = async () => {
  if (!secretForm.value.name) {
    alert('请填写 Secret 名称')
    return
  }

  creating.value = true
  try {
    // 构建 string_data
    const stringData = {}
    secretForm.value.dataItems.forEach(item => {
      if (item.key && item.value) {
        stringData[item.key] = item.value
      }
    })

    await secretApi.create({
      namespace: secretForm.value.namespace,
      name: secretForm.value.name,
      type: secretForm.value.type,
      string_data: stringData
    })

    showCreateModal.value = false
    resetSecretForm()
    fetchSecrets()
    alert('Secret 创建成功')
  } catch (error) {
    console.error('创建Secret失败:', error)
    alert(error.kube_message_error || error.message || '创建Secret失败')
  } finally {
    creating.value = false
  }
}

// 重置表单
const resetSecretForm = () => {
  secretForm.value = {
    namespace: 'default',
    name: '',
    type: 'Opaque',
    dataItems: [{ key: '', value: '' }]
  }
}

// 关闭创建模态框
const closeCreateModal = () => {
  showCreateModal.value = false
  resetSecretForm()
  createMode.value = 'form'
  createYamlContent.value = ''
  createYamlError.value = ''
  yamlPreviewContent.value = ''
}

// ==================== YAML 创建相关 ====================
// 加载 Secret YAML 模板
const loadSecretYamlTemplate = () => {
  createYamlContent.value = `# Secret YAML 模板
# 示例 1: 通用密钥 (Opaque)
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
  labels:
    app: my-app
type: Opaque
stringData:
  # 使用 stringData 输入明文，K8s 会自动 Base64 编码
  username: admin
  password: secret123
  config.yaml: |
    database:
      host: localhost
      port: 3306`
  createYamlError.value = ''
  updateYamlPreview()
}

// 清空 YAML 内容
const clearYamlContent = () => {
  createYamlContent.value = ''
  yamlPreviewContent.value = ''
  createYamlError.value = ''
}

// 格式化 YAML
const formatYaml = () => {
  // 简单的格式化：去除尾部空格，统一缩进
  const lines = createYamlContent.value.split('\n')
  const formatted = lines.map(line => line.trimEnd()).join('\n')
  createYamlContent.value = formatted
  updateYamlPreview()
}

// 更新 YAML 预览
const updateYamlPreview = () => {
  if (!createYamlContent.value.trim()) {
    yamlPreviewContent.value = ''
    createYamlError.value = ''
    return
  }
  
  try {
    // 简单验证
    const content = createYamlContent.value
    if (!content.includes('apiVersion:') || !content.includes('kind:')) {
      createYamlError.value = 'YAML 必须包含 apiVersion 和 kind 字段'
      yamlPreviewContent.value = ''
      return
    }
    
    createYamlError.value = ''
    yamlPreviewContent.value = content
  } catch (e) {
    createYamlError.value = `YAML 格式错误: ${e.message}`
    yamlPreviewContent.value = ''
  }
}

// 监听 createYamlContent 变化
watch(createYamlContent, () => {
  updateYamlPreview()
})

// 监听 createMode 变化
watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !createYamlContent.value.trim()) {
    loadSecretYamlTemplate()
  }
})

// 从 YAML 创建 Secret
const createSecretFromYaml = async () => {
  if (!createYamlContent.value.trim()) {
    createYamlError.value = '请输入 YAML 内容'
    return
  }
  
  // 简单验证
  if (!createYamlContent.value.includes('kind:')) {
    createYamlError.value = 'YAML 中必须包含 "kind:" 字段'
    return
  }
  
  creating.value = true
  createYamlError.value = ''
  
  try {
    const res = await secretApi.createFromYaml({ yaml: createYamlContent.value })
    if (res.code === 0) {
      alert('Secret 创建成功')
      closeCreateModal()
      await fetchSecrets()
    } else {
      createYamlError.value = res.msg || '创建失败'
    }
  } catch (e) {
    console.error('创建失败:', e)
    createYamlError.value = e?.kube_message_error || e?.msg || e?.message || '创建失败'
  } finally {
    creating.value = false
  }
}

// 添加数据项
const addDataItem = () => {
  secretForm.value.dataItems.push({ key: '', value: '' })
}

// 删除数据项
const removeDataItem = (index) => {
  secretForm.value.dataItems.splice(index, 1)
}

// 确认删除
const confirmDelete = (secret) => {
  selectedSecret.value = secret
  showDeleteModal.value = true
}

// 删除 Secret
const deleteSecret = async () => {
  deleting.value = true
  try {
    await secretApi.delete({
      namespace: selectedSecret.value.namespace,
      name: selectedSecret.value.name
    })

    showDeleteModal.value = false
    selectedSecret.value = null
    fetchSecrets()
  } catch (error) {
    console.error('删除Secret失败:', error)
    alert(error.kube_message_error || error.message || '删除Secret失败')
  } finally {
    deleting.value = false
  }
}

// 查看详情
const viewSecret = async (secret) => {
  selectedSecret.value = secret
  showDetailModal.value = true
  loadingDetail.value = true
  detailData.value = null

  try {
    const res = await secretApi.detail({
      namespace: secret.namespace,
      name: secret.name
    })

    if (res.code === 0 && res.data) {
      detailData.value = res.data
    }
  } catch (error) {
    console.error('获取详情失败:', error)
    alert(error.kube_message_error || error.message || '获取详情失败')
    showDetailModal.value = false
  } finally {
    loadingDetail.value = false
  }
}

// 解码 Secret
const decodeSecret = async (secret) => {
  decodedSecret.value = secret
  showDecodeModal.value = true
  loadingDecode.value = true
  decodedData.value = null

  try {
    const res = await secretApi.decode({
      namespace: secret.namespace,
      name: secret.name
    })

    if (res.code === 0 && res.data) {
      decodedData.value = res.data.decoded_data || res.data
    }
  } catch (error) {
    console.error('解码失败:', error)
    alert(error.kube_message_error || error.message || '解码失败')
    showDecodeModal.value = false
  } finally {
    loadingDecode.value = false
  }
}

// 查看 YAML
const viewYaml = async (secret) => {
  selectedYamlSecret.value = secret
  showYamlModal.value = true
  loadingYaml.value = true
  yamlContent.value = ''
  yamlViewError.value = ''
  yamlEditMode.value = false

  try {
    const res = await secretApi.yaml({
      namespace: secret.namespace,
      name: secret.name
    })

    if (res.code === 0 && res.data) {
      yamlContent.value = res.data.yaml || res.data
    } else {
      yamlViewError.value = res.msg || '获取 YAML 失败'
    }
  } catch (error) {
    console.error('获取YAML失败:', error)
    yamlViewError.value = error.kube_message_error || error.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

// 关闭 YAML 模态框
const closeYamlModal = () => {
  showYamlModal.value = false
  selectedYamlSecret.value = null
  yamlContent.value = ''
  yamlEditMode.value = false
  yamlViewError.value = ''
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value || !selectedYamlSecret.value) return
  
  const blob = new Blob([yamlContent.value], { type: 'text/yaml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${selectedYamlSecret.value.name}.yaml`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

// 应用 YAML 更改
const applyYamlChanges = async () => {
  if (!yamlContent.value.trim() || !selectedYamlSecret.value) return
  
  savingYaml.value = true
  yamlViewError.value = ''
  
  try {
    const res = await secretApi.applyYaml({
      namespace: selectedYamlSecret.value.namespace,
      name: selectedYamlSecret.value.name,
      yaml: yamlContent.value
    })
    
    if (res.code === 0) {
      alert('YAML 更新成功')
      closeYamlModal()
      await fetchSecrets()
    } else {
      yamlViewError.value = res.msg || '更新失败'
    }
  } catch (e) {
    console.error('应用 YAML 失败:', e)
    yamlViewError.value = e?.kube_message_error || e?.msg || e?.message || '更新失败'
  } finally {
    savingYaml.value = false
  }
}

// 复制到剪贴板
const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    alert('已复制到剪贴板')
  } catch (e) {
    console.error('复制失败:', e)
  }
}

// ==================== 自动刷新 ====================
watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshTimer = setInterval(fetchSecrets, 90000) // 90秒刷新
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
})

// ==================== 生命周期 ====================
onMounted(() => {
  fetchNamespaces()
  fetchSecrets()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
/* 基础布局 */
.resource-view {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #7f8c8d;
  margin: 0;
  font-size: 14px;
}

/* 操作栏 */
.action-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.search-box input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 250px;
  font-size: 14px;
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn-filter {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn-filter:hover {
  background: #f8f9fa;
}

.btn-filter.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.filter-dropdown select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 14px;
  user-select: none;
}

.refresh-indicator {
  color: #4caf50;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn-primary {
  background: #326ce5;
  color: white;
}

.btn-primary:hover {
  background: #2554c7;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #5a6268;
}

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-danger:hover {
  background: #c82333;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 视图切换 */
.view-toggle {
  display: flex;
  gap: 4px;
}

.btn-view {
  padding: 8px 12px;
  background: white;
  border: 1px solid #ddd;
}

.btn-view.active {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-batch {
  background: #17a2b8;
  color: white;
}

.btn-batch:hover {
  background: #138496;
}

/* 错误提示 */
.error-box {
  background: #fee;
  color: #c33;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
  border-left: 4px solid #c33;
}

/* 统计栏 */
.stats-bar {
  display: flex;
  gap: 24px;
  background: white;
  padding: 16px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  color: #666;
  font-size: 14px;
}

.stat-value {
  font-weight: 600;
  font-size: 16px;
  color: #333;
}

/* 批量操作浮动栏 */
.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.batch-count {
  font-weight: 600;
}

.batch-clear {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
}

.batch-actions {
  display: flex;
  gap: 8px;
}

.batch-btn {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: rgba(255,255,255,0.3);
}

.batch-btn.danger {
  background: rgba(220,53,69,0.8);
}

.batch-btn.danger:hover {
  background: rgba(220,53,69,1);
}

/* 表格容器 */
.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.resource-table th {
  background: #f8f9fa;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #495057;
  border-bottom: 2px solid #dee2e6;
}

.resource-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f1f3f4;
  vertical-align: middle;
}

.resource-table tbody tr:hover {
  background: #f8f9fa;
}

.row-selected {
  background: #e3f2fd !important;
}

/* 类型徽章 */
.type-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.type-badge.opaque {
  background: #e8f5e9;
  color: #2e7d32;
}

.type-badge.tls {
  background: #e3f2fd;
  color: #1565c0;
}

.type-badge.docker {
  background: #fff3e0;
  color: #ef6c00;
}

.type-badge.sa-token {
  background: #f3e5f5;
  color: #7b1fa2;
}

.type-badge.default {
  background: #f5f5f5;
  color: #616161;
}

/* Secret 名称 */
.secret-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.secret-name .icon {
  font-size: 16px;
}

.age-badge {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  color: #666;
}

/* 命名空间徽章 */
.namespace-badge {
  background: #e6fffa;
  color: #0d9488;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

/* 数据计数 */
.data-count {
  font-weight: 500;
  color: #666;
}

/* 标签 */
.label-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.label-tag {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  color: #555;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.label-empty {
  color: #999;
}

.label-more {
  background: #326ce5;
  color: white;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

/* 操作按钮 */
.action-icons {
  display: flex;
  gap: 4px;
}

.action-btn {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  background: #f8f9fa;
}

.action-btn:hover {
  background: #e9ecef;
}

.action-btn.icon-only {
  padding: 6px 8px;
}

.action-btn.danger:hover {
  background: #fee;
  color: #c33;
}

/* 加载和空状态 */
.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #326ce5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  color: #999;
}

/* 卡片视图 */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.resource-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
  transition: all 0.2s;
  position: relative;
}

.resource-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.card-selected {
  border: 2px solid #326ce5;
}

.card-checkbox {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 1;
}

.card-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-icon {
  font-size: 20px;
}

.card-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.card-body {
  padding: 16px;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.card-meta-item .label {
  color: #666;
}

.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  background: #f8f9fa;
  border-top: 1px solid #f0f0f0;
}

.card-btn {
  flex: 1;
  padding: 8px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  background: white;
  border: 1px solid #ddd;
}

.card-btn:hover {
  background: #f0f0f0;
}

.card-btn.primary {
  background: #326ce5;
  color: white;
  border-color: #326ce5;
}

.card-btn.primary:hover {
  background: #2554c7;
}

.card-btn.danger {
  color: #dc3545;
  border-color: #dc3545;
}

.card-btn.danger:hover {
  background: #dc3545;
  color: white;
}

/* 现代化分页样式 */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 16px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  flex-wrap: wrap;
  gap: 16px;
}

.pagination-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pagination-summary {
  font-size: 14px;
  color: #6b7280;
}

.pagination-summary strong {
  color: #1f2937;
  font-weight: 600;
}

.pagination-center {
  display: flex;
  gap: 6px;
  align-items: center;
}

.pagination-btn {
  min-width: 36px;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: white;
  color: #374151;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.pagination-btn:hover:not(:disabled) {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination-btn.page-number.active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-color: #3b82f6;
  color: white;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.pagination-ellipsis {
  padding: 0 8px;
  color: #9ca3af;
  font-size: 14px;
}

.pagination-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-size-select {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.page-size-select:hover {
  border-color: #9ca3af;
}

.page-size-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.pagination-goto {
  font-size: 14px;
  color: #6b7280;
}

.page-jump-input {
  width: 60px;
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  text-align: center;
  transition: all 0.2s;
}

.page-jump-input:hover {
  border-color: #9ca3af;
}

.page-jump-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal-content.large {
  max-width: 700px;
}

.modal-content.yaml-modal {
  max-width: 900px;
}

.modal-content.modal-danger .modal-header {
  background: #fee;
  border-bottom: 1px solid #fcc;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #eee;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  padding: 0;
  line-height: 1;
}

.close-btn:hover {
  color: #333;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid #eee;
}

/* 表单 */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #333;
}

.form-group label.required::after {
  content: ' *';
  color: #dc3545;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 2px rgba(50, 108, 229, 0.1);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
}

/* 数据项输入 */
.data-item-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.data-key {
  flex: 1;
}

.data-value {
  flex: 2;
}

.remove-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 8px;
}

.add-btn {
  background: #f0f0f0;
  border: 1px dashed #ccc;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  width: 100%;
  margin-top: 8px;
}

.add-btn:hover {
  background: #e8e8e8;
}

/* 危险警告 */
.danger-warning {
  background: #fff3cd;
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.warning-text {
  color: #856404;
  font-weight: 500;
}

.delete-list {
  max-height: 200px;
  overflow-y: auto;
  margin: 12px 0;
  padding-left: 20px;
}

.confirm-input {
  width: 100%;
  padding: 8px 12px;
  border: 2px solid #dc3545;
  border-radius: 4px;
  margin-top: 8px;
}

/* 详情 */
.detail-content {
  padding: 0;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
  font-size: 14px;
  color: #666;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
}

.detail-table td {
  padding: 8px 0;
  font-size: 14px;
}

.detail-table td.label {
  width: 120px;
  color: #666;
  font-weight: 500;
}

.data-table td {
  border-bottom: 1px solid #f0f0f0;
}

.data-value-cell {
  position: relative;
}

.data-value-cell code {
  background: #f5f5f5;
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 12px;
  word-break: break-all;
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

/* 解码内容 */
.decode-content {
  padding: 0;
}

.decode-warning {
  background: #fff3cd;
  padding: 12px 16px;
  border-radius: 4px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.warning-icon {
  font-size: 18px;
}

.decoded-value {
  background: #f5f5f5;
  padding: 8px 12px;
  border-radius: 4px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.copy-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 4px 8px;
  cursor: pointer;
  font-size: 12px;
}

.copy-btn:hover {
  background: #f0f0f0;
}

/* YAML 编辑器 */
.yaml-content {
  height: 500px;
}

.yaml-editor {
  width: 100%;
  height: 100%;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
  resize: none;
  background: #f8f9fa;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
}

/* 加载状态 */
.loading-state {
  text-align: center;
  padding: 40px;
  color: #666;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #326ce5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 12px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  color: #666;
  font-size: 14px;
}

/* 创建模态框 - Rancher/Kuboard 风格 */
.modal-create-secret {
  width: 95%;
  max-width: 1200px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-create-secret .modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e8e8e8;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.modal-create-secret .modal-header h2 {
  margin: 0;
  font-size: 18px;
}

.view-toggle-buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
  margin-right: 16px;
}

.view-toggle-btn {
  padding: 6px 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: transparent;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.view-toggle-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.view-toggle-btn.active {
  background: rgba(255, 255, 255, 0.2);
  border-color: white;
}

.modal-create-secret .modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  background: #f5f7fa;
}

/* 表单分区 */
.form-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eee;
}

.section-header h3 {
  margin: 0;
  font-size: 15px;
  color: #333;
}

.section-icon {
  font-size: 18px;
}

.section-tip {
  margin-left: auto;
  font-size: 12px;
  color: #999;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #333;
  font-size: 13px;
}

.form-group label.required::after {
  content: ' *';
  color: #dc3545;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #326ce5;
}

/* 数据项容器 */
.data-items-container {
  background: #fafafa;
  border-radius: 4px;
  padding: 12px;
}

.data-item-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
}

.data-item-fields {
  flex: 1;
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 12px;
}

.data-key {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
}

.data-value {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  font-family: 'Courier New', monospace;
  resize: vertical;
  min-height: 60px;
}

.remove-btn {
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #dc3545;
  border-radius: 4px;
  color: #dc3545;
  cursor: pointer;
  transition: all 0.2s;
}

.remove-btn:hover {
  background: #dc3545;
  color: white;
}

.add-btn {
  width: 100%;
  padding: 10px;
  background: white;
  border: 2px dashed #ddd;
  border-radius: 4px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.add-btn:hover {
  border-color: #326ce5;
  color: #326ce5;
  background: #f0f7ff;
}

/* YAML 创建分区 */
.yaml-create-section {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.yaml-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.yaml-editor-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  flex: 1;
  min-height: 400px;
}

.yaml-editor-wrapper,
.yaml-preview-wrapper {
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.editor-label {
  padding: 10px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e8e8e8;
  font-weight: 500;
  font-size: 13px;
  color: #333;
}

.create-yaml-editor {
  flex: 1;
  width: 100%;
  min-height: 350px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  border: none;
  padding: 16px;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
}

.create-yaml-editor:focus {
  outline: none;
}

.yaml-preview {
  flex: 1;
  overflow-y: auto;
  background: #f5f5f5;
}

.yaml-preview pre {
  margin: 0;
  padding: 16px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
}

.preview-placeholder {
  padding: 40px;
  text-align: center;
  color: #999;
}

.yaml-error {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fee;
  color: #c33;
  padding: 12px 16px;
  border-radius: 4px;
  margin-top: 12px;
}

.error-icon {
  font-size: 18px;
}

.yaml-editor-footer {
  margin-top: 16px;
}

.yaml-tips {
  background: #e8f4fd;
  padding: 12px 16px;
  border-radius: 4px;
  border-left: 4px solid #326ce5;
}

.yaml-tips strong {
  display: block;
  margin-bottom: 8px;
  color: #326ce5;
}

.yaml-tips ul {
  margin: 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin-bottom: 4px;
  font-size: 13px;
  color: #555;
}

.yaml-tips code {
  background: #fff;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
}

/* YAML 查看模态框 */
.yaml-modal {
  width: 90%;
  max-width: 900px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.yaml-modal .modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.yaml-header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.yaml-modal-body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.yaml-modal-body .yaml-editor-wrapper {
  flex: 1;
  min-height: 400px;
}

.yaml-modal-body .yaml-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
}

.yaml-content-pre {
  width: 100%;
  height: 100%;
  min-height: 400px;
  max-height: 500px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
  background: #f8f9fa;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.info-box {
  background: #f0f7ff;
  padding: 12px 16px;
  border-radius: 4px;
  border-left: 4px solid #326ce5;
}

.info-box div {
  margin-bottom: 4px;
  font-size: 13px;
}

.info-box div:last-child {
  margin-bottom: 0;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e8e8e8;
  background: white;
}

.btn-icon {
  margin-right: 4px;
}
</style>
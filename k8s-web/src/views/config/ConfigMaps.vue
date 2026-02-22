<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>ConfigMap 管理</h1>
      <p>Kubernetes 集群中的配置管理</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 ConfigMap 名称..."
          @input="onSearchInput"
        />
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
        <button class="btn btn-primary" @click="showCreateModal = true">创建 ConfigMap</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedConfigMaps.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedConfigMaps.length }} 个 ConfigMap</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeleteModal" title="批量删除">
          🗑️ 批量删除
        </button>
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
          <th style="min-width: 200px;">名称</th>
          <th style="width: 130px;">命名空间</th>
          <th style="width: 100px;">数据项</th>
          <th style="width: 170px;">创建时间</th>
          <th style="width: 150px;">操作</th>
        </tr>
        </thead>
        <tbody>
        <tr v-if="loading">
          <td :colspan="batchMode ? 6 : 5" class="loading-row">
            <div class="loading-spinner"></div>
            <span>加载中...</span>
          </td>
        </tr>
        <tr v-else-if="paginatedConfigMaps.length === 0">
          <td :colspan="batchMode ? 6 : 5" class="empty-row">
            暂无数据
          </td>
        </tr>
        <tr
          v-else
          v-for="(cm, index) in paginatedConfigMaps"
          :key="`cm-${index}-${cm.name || 'unnamed'}-${cm.namespace || 'default'}`"
          :class="{ 'row-selected': isConfigMapSelected(cm) }"
        >
          <td v-if="batchMode">
            <input
              type="checkbox"
              :checked="isConfigMapSelected(cm)"
              @change="toggleConfigMapSelection(cm)"
            />
          </td>
          <td>
            <div class="configmap-name">
              <span class="icon">📋</span>
              <span>{{ cm.name }}</span>
            </div>
          </td>
          <td>
            <span class="namespace-badge">{{ cm.namespace }}</span>
          </td>
          <td>
            <span class="data-count">{{ cm.dataCount || 0 }} 项</span>
          </td>
          <td>
            <span class="timestamp">{{ formatTime(cm.createdAt) }}</span>
          </td>
          <td>
            <div class="action-btns">
              <button class="btn-icon" @click="viewDetail(cm)" title="查看详情">ℹ️</button>
              <button class="btn-icon" @click="viewYaml(cm)" title="查看 YAML">📝</button>
              <button class="btn-icon" @click="editConfigMap(cm)" title="编辑">✏️</button>
              <button class="btn-icon danger" @click="confirmDelete(cm)" title="删除">🗑️</button>
            </div>
          </td>
        </tr>
        </tbody>
      </table>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="card-grid">
      <div v-if="loading" class="loading-card">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      <div v-else-if="paginatedConfigMaps.length === 0" class="empty-state">
        <p>暂无 ConfigMap</p>
      </div>
      <div
        v-else
        v-for="(cm, index) in paginatedConfigMaps"
        :key="`card-${index}-${cm.name || 'unnamed'}-${cm.namespace || 'default'}`"
        class="resource-card"
        :class="{ 'card-selected': isConfigMapSelected(cm) }"
      >
        <div v-if="batchMode" class="card-checkbox">
          <input
            type="checkbox"
            :checked="isConfigMapSelected(cm)"
            @change="toggleConfigMapSelection(cm)"
          />
        </div>
        <div class="card-header">
          <div class="card-title">
            <span class="icon">📋</span>
            <h3>{{ cm.name }}</h3>
          </div>
          <span class="namespace-badge">{{ cm.namespace }}</span>
        </div>
        <div class="card-body">
          <div class="card-info">
            <span class="info-label">数据项:</span>
            <span class="info-value">{{ cm.dataCount || 0 }}</span>
          </div>
          <div class="card-info">
            <span class="info-label">创建时间:</span>
            <span class="info-value">{{ formatTime(cm.createdAt) }}</span>
          </div>
        </div>
        <div class="card-footer">
          <button class="btn-card" @click="viewDetail(cm)">查看详情</button>
          <button class="btn-card" @click="editConfigMap(cm)">编辑</button>
          <button class="btn-card danger" @click="confirmDelete(cm)">删除</button>
        </div>
      </div>
    </div>

    <!-- 分页（现代化三段式布局） -->
    <div v-if="filteredConfigMaps.length > 0" class="pagination-wrapper">
      <div class="pagination-left">
        <span class="pagination-summary">共 <strong>{{ filteredConfigMaps.length }}</strong> 条</span>
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

    <!-- 创建/编辑模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="closeCreateModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>{{ editMode ? '编辑' : '创建' }} ConfigMap</h3>
          <div v-if="!editMode" class="mode-toggle">
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
          <button @click="closeCreateModal" class="close-btn">×</button>
        </div>

        <div class="modal-body scrollable">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
            <div class="form-group">
              <label>名称 <span class="required">*</span></label>
              <input
                v-model="configMapForm.name"
                type="text"
                class="form-input"
                placeholder="configmap-name"
                :disabled="editMode"
              />
            </div>

            <!-- ✅【这里是你要的新增：命名空间选择 + 新建命名空间按钮】 -->
            <div class="form-group">
              <label>命名空间 <span class="required">*</span></label>

              <div class="ns-row">
                <select v-model="configMapForm.namespace" class="form-select" :disabled="editMode">
                  <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                </select>

                <button
                  v-if="!editMode"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openCreateNamespaceModal"
                  title="创建新命名空间"
                >
                  ➕ 新建命名空间
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>数据 (Data)</label>
              <div class="data-editor">
                <div v-for="(item, index) in configMapForm.data" :key="index" class="data-row">
                  <input
                    v-model="item.key"
                    class="data-input data-key"
                    placeholder="键名"
                  />
                  <textarea
                    v-model="item.value"
                    class="data-textarea"
                    placeholder="值"
                    rows="3"
                  ></textarea>
                  <button
                    type="button"
                    class="btn-icon btn-remove"
                    @click="removeDataItem(index)"
                  >
                    🗑️
                  </button>
                </div>
                <button type="button" class="btn btn-secondary btn-sm" @click="addDataItem">
                  ➕ 添加数据项
                </button>
              </div>
            </div>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <p class="yaml-hint">
                ✨ 支持多资源 YAML 创建（用 <code>---</code> 分隔），可同时创建 ConfigMap、Service、Deployment 等依赖资源
              </p>
              <div class="yaml-header-buttons">
                <button class="load-template-btn" @click="loadMultiResourceYamlTemplate">
                  📑 加载多资源模板
                </button>
                <button class="load-template-btn" @click="loadConfigMapYamlTemplate">
                  📄 ConfigMap 模板
                </button>
                <button class="copy-yaml-btn" @click="copyYamlContent">
                  📋 复制
                </button>
                <button class="reset-yaml-btn" @click="resetYamlContent">
                  🔄 重置
                </button>
              </div>

              <!-- 多资源预览面板 -->
              <div v-if="multiResourcePreview.resources.length > 0" class="multi-resource-preview">
                <div class="preview-header">
                  <h4>🔍 资源预览 ({{ multiResourcePreview.resources.length }} 个资源)</h4>
                  <button
                    v-if="multiResourcePreview.errors.length === 0"
                    class="btn btn-sm btn-primary"
                    @click="applyMultiResourceYaml"
                    :disabled="creating"
                  >
                    {{ creating ? '创建中...' : '🚀 一键创建所有资源' }}
                  </button>
                </div>

                <!-- 依赖关系图 -->
                <div v-if="multiResourcePreview.dependencies.length > 0" class="dependencies-info">
                  <h5>🔗 依赖关系</h5>
                  <div class="dependency-graph">
                    <div
                      v-for="dep in multiResourcePreview.dependencies"
                      :key="dep.id"
                      class="dependency-item"
                    >
                      <span class="dep-source">{{ dep.source.kind }}/{{ dep.source.name }}</span>
                      <span class="dep-arrow">→</span>
                      <span class="dep-target">{{ dep.target.kind }}/{{ dep.target.name }}</span>
                    </div>
                  </div>
                </div>

                <!-- 资源列表 -->
                <div class="resource-list">
                  <div
                    v-for="resource in multiResourcePreview.resources"
                    :key="`${resource.kind}-${resource.name}`"
                    class="resource-item"
                    :class="{
                      'has-warnings': resource.warnings?.length > 0,
                      'has-errors': multiResourcePreview.errors.some(e => e.includes(resource.name))
                    }"
                  >
                    <div class="resource-header">
                      <span class="resource-kind" :class="getResourceKindClass(resource.kind)">
                        {{ getResourceIcon(resource.kind) }} {{ resource.kind }}
                      </span>
                      <span class="resource-name">{{ resource.name }}</span>
                      <span class="resource-namespace">({{ resource.namespace }})</span>
                      <span class="resource-order">#{{ resource.order }}</span>
                    </div>

                    <div v-if="resource.warnings?.length > 0" class="resource-warnings">
                      <div v-for="warning in resource.warnings" :key="warning" class="warning-item">
                        ⚠️ {{ warning }}
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 错误信息 -->
                <div v-if="multiResourcePreview.errors.length > 0" class="preview-errors">
                  <h5>❌ 依赖错误</h5>
                  <div class="error-list">
                    <div
                      v-for="error in multiResourcePreview.errors"
                      :key="error"
                      class="error-item"
                    >
                      {{ error }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <textarea
              v-model="yamlCreateContent"
              class="yaml-editor"
              placeholder="输入或粘贴 YAML 内容..."
              spellcheck="false"
            ></textarea>

            <div v-if="yamlCreateError" class="yaml-error">
              <span class="error-icon">⚠️</span>
              {{ yamlCreateError }}
            </div>

            <div class="yaml-editor-footer">
              <div class="yaml-tips">
                <strong>💡 提示：</strong>
                <ul>
                  <li>✅ 支持单资源或多资源 YAML（用 <code>---</code> 分隔）</li>
                  <li>🔗 可同时创建：ConfigMap + Service + Deployment / PVC + ConfigMap + Deployment</li>
                  <li>📦 资源创建顺序：ConfigMap/Secret → Service → Deployment</li>
                  <li>🚀 点击"加载多资源模板"获取完整的多资源示例</li>
                  <li>🔍 输入 YAML 后会自动解析并显示资源预览</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeCreateModal" class="btn btn-secondary">取消</button>
          <button
            v-if="createMode === 'form'"
            @click="submitConfigMap"
            class="btn btn-primary"
            :disabled="creating"
          >
            {{ creating ? '处理中...' : (editMode ? '更新' : '创建') }}
          </button>
          <button
            v-if="createMode === 'yaml'"
            @click="createConfigMapFromYaml"
            class="btn btn-primary"
            :disabled="creating || !yamlCreateContent"
          >
            {{ creating ? '应用中...' : '✅ 应用 YAML' }}
          </button>
          <button
            v-if="createMode === 'yaml' && multiResourcePreview.resources.length > 0 && multiResourcePreview.errors.length === 0"
            @click="applyMultiResourceYaml"
            class="btn btn-success"
            :disabled="creating"
          >
            {{ creating ? '创建中...' : `🚀 创建全部 (${multiResourcePreview.resources.length} 个资源)` }}
          </button>
        </div>
      </div>
    </div>

    <!-- ✅【新增：创建命名空间弹窗】 -->
    <div
      v-if="showCreateNamespaceModal"
      class="modal-overlay"
      @click="closeCreateNamespaceModal"
    >
      <div class="modal-content" style="max-width:520px" @click.stop>
        <div class="modal-header">
          <h3>创建命名空间</h3>
          <button class="close-btn" @click="closeCreateNamespaceModal">×</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>命名空间名称 <span class="required">*</span></label>
            <input
              v-model="newNamespaceName"
              class="form-input"
              placeholder="例如: dev / test / prod"
              @keyup.enter="createNamespace"
            />
            <div style="margin-top:8px;color:#6b7280;font-size:12px;">
              仅支持：小写字母/数字/-，长度 1-63，且不能以 - 开头或结尾（如：my-ns-1）
            </div>
          </div>

          <div v-if="createNamespaceError" class="error-box">
            {{ createNamespaceError }}
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeCreateNamespaceModal">取消</button>
          <button class="btn btn-primary" @click="createNamespace" :disabled="creatingNamespace">
            {{ creatingNamespace ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click="showDetailModal = false">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>ConfigMap 详情</h3>
          <button @click="showDetailModal = false" class="close-btn">×</button>
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
                  <td class="label">数据项数量:</td>
                  <td>{{ detailData.dataCount || 0 }}</td>
                </tr>
                <tr>
                  <td class="label">创建时间:</td>
                  <td>{{ detailData.createdAt }}</td>
                </tr>
                </tbody>
              </table>
            </div>
            <div class="detail-section" v-if="detailData.data && Object.keys(detailData.data).length > 0">
              <h4>数据内容</h4>
              <div class="data-display">
                <div v-for="(value, key) in detailData.data" :key="key" class="data-item">
                  <div class="data-item-key">{{ key }}</div>
                  <pre class="data-item-value">{{ value }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showDetailModal = false" class="btn btn-secondary">关闭</button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click="closeYamlModal">
      <div class="modal-content yaml-modal" @click.stop>
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlConfigMap?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml" :disabled="loadingYaml">
              📄 下载
            </button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true">
              ✏️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="close-btn" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="loadingYaml" class="loading-state">加载 YAML...</div>
          <div v-else-if="yamlViewError" class="error-box">{{ yamlViewError }}</div>
          <div v-else class="yaml-editor-wrapper">
            <textarea v-if="yamlEditMode" v-model="yamlViewContent" class="yaml-editor" spellcheck="false"></textarea>
            <pre v-else class="yaml-content">{{ yamlViewContent }}</pre>
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

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content delete-confirm" @click.stop>
        <div class="modal-header">
          <h3>⚠️ 确认删除</h3>
          <button @click="showDeleteModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除 ConfigMap <strong>{{ selectedConfigMap?.name }}</strong> 吗？</p>
          <p class="warning-text">删除后将无法恢复，请谨慎操作。</p>
        </div>
        <div class="modal-footer">
          <button @click="showDeleteModal = false" class="btn btn-secondary">取消</button>
          <button @click="deleteConfigMap" class="btn btn-danger" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click="closeBatchDeleteModal">
      <div class="modal-content batch-delete-modal" @click.stop>
        <div class="modal-header">
          <h3>⚠️ 确认批量删除</h3>
          <button @click="closeBatchDeleteModal" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <p class="warning-text" style="margin-bottom: 16px;">
            您即将删除以下 {{ selectedConfigMaps.length }} 个 ConfigMap，此操作<strong>不可恢复</strong>！
          </p>
          <div class="batch-delete-list">
            <div v-for="cm in selectedConfigMaps" :key="cm.name + cm.namespace" class="batch-delete-item">
              🗑️ {{ cm.namespace }}/{{ cm.name }}
            </div>
          </div>
          <div class="confirm-input-group">
            <label>请输入 <code>DELETE</code> 确认删除：</label>
            <input
              type="text"
              v-model="deleteConfirmText"
              class="form-input"
              placeholder="DELETE"
              @keyup.enter="executeBatchDelete"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeBatchDeleteModal" class="btn btn-secondary">取消</button>
          <button
            @click="executeBatchDelete"
            class="btn btn-danger"
            :disabled="deleteConfirmText !== 'DELETE' || batchExecuting"
          >
            {{ batchExecuting ? '删除中...' : `确认删除 ${selectedConfigMaps.length} 个` }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import configmapApi from '@/api/cluster/config/configmap'
import namespaceApi from '@/api/cluster/config/namespace'

// ==================== 状态管理 ====================
const loading = ref(false)
const errorMsg = ref('')
const configmaps = ref([])
const total = ref(0)

// 视图模式
const viewMode = ref('table') // 'table' | 'card'

// 批量操作
const batchMode = ref(false)
const selectedConfigMaps = ref([])
const showBatchDeleteModal = ref(false)
const deleteConfirmText = ref('')
const batchExecuting = ref(false)

// 搜索和过滤
const searchQuery = ref('')
const namespaceFilter = ref('')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)
const jumpPage = ref(1)

// 自动刷新
const autoRefresh = ref(false)
let refreshTimer = null

// 模态框状态
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const showDeleteModal = ref(false)
const showYamlModal = ref(false)

// ✅ 新增：创建命名空间弹窗
const showCreateNamespaceModal = ref(false)
const newNamespaceName = ref('')
const creatingNamespace = ref(false)
const createNamespaceError = ref('')

// 操作状态
const creating = ref(false)
const deleting = ref(false)
const loadingDetail = ref(false)
const loadingYaml = ref(false)
const savingYaml = ref(false)

// 选中的 ConfigMap
const selectedConfigMap = ref(null)
const selectedYamlConfigMap = ref(null)
const detailData = ref(null)

// 编辑模式
const editMode = ref(false)

// 创建模式
const createMode = ref('form') // 'form' | 'yaml'

// YAML 相关
const yamlViewContent = ref('')
const yamlEditMode = ref(false)
const yamlViewError = ref('')
const yamlCreateContent = ref('')  // YAML 创建内容
const yamlCreateError = ref('')    // YAML 创建错误信息

// 多资源预览数据
const multiResourcePreview = ref({
  resources: [],
  dependencies: [],
  errors: [],
  total: 0
})

// 命名空间
const namespaces = ref(['default', 'kube-system', 'kube-public'])

// 创建表单
const configMapForm = ref({
  name: '',
  namespace: 'default',
  data: [{ key: '', value: '' }]
})

// ==================== 计算属性 ====================
const filteredConfigMaps = computed(() => {
  let result = configmaps.value

  // 命名空间过滤
  if (namespaceFilter.value) {
    result = result.filter(cm => cm.namespace === namespaceFilter.value)
  }

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(cm =>
      (cm.name || '').toLowerCase().includes(query) ||
      (cm.namespace || '').toLowerCase().includes(query)
    )
  }

  return result
})

const paginatedConfigMaps = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredConfigMaps.value.slice(startIndex, endIndex)
})

const isAllSelected = computed(() => {
  return paginatedConfigMaps.value.length > 0 &&
    paginatedConfigMaps.value.every(cm => isConfigMapSelected(cm))
})

// 分页相关计算属性
const totalPages = computed(() => Math.ceil(filteredConfigMaps.value.length / itemsPerPage.value) || 1)

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

// ✅（推荐）拉取命名空间列表：用于下拉框更真实，不仅仅靠 ConfigMap 反推
const fetchNamespaces = async () => {
  try {
    // 你后端如果没有 namespaceApi.list，就删掉这段，保留默认即可
    const res = await namespaceApi.list?.()
    if (res?.code === 0 && res?.data) {
      const list = res.data.list || res.data || []
      const nsNames = list
        .map(item => item?.name || item?.metadata?.name)
        .filter(Boolean)

      const nsSet = new Set(['default', 'kube-system', 'kube-public'])
      nsNames.forEach(n => nsSet.add(n))
      namespaces.value = Array.from(nsSet).sort()
    }
  } catch (e) {
    // 不影响主功能，静默即可
    console.warn('fetchNamespaces skipped or failed:', e)
  }
}

// 获取 ConfigMap 列表
const fetchConfigMaps = async () => {
  try {
    loading.value = true
    errorMsg.value = ''

    const res = await configmapApi.list({
      namespace: '',  // 查询所有命名空间
      page: 1,
      limit: 1000  // 前端分页
    })

    if (res.code === 0 && res.data) {
      const list = res.data.list || res.data || []

      configmaps.value = list.map((item) => {
        return {
          name: item.name || item.metadata?.name || '',
          namespace: item.namespace || item.metadata?.namespace || 'default',
          dataCount: item.data_count || (item.data ? Object.keys(item.data).length : 0),
          createdAt: item.created_at || item.createdAt || item.metadata?.creationTimestamp || '',
          data: item.data || {}
        }
      })

      total.value = res.data.total || list.length

      // 动态更新命名空间列表（从 CM 反推兜底）
      const nsSet = new Set(['default', 'kube-system', 'kube-public'])
      configmaps.value.forEach(cm => {
        if (cm.namespace) nsSet.add(cm.namespace)
      })
      namespaces.value = Array.from(nsSet).sort()
    }
  } catch (error) {
    console.error('获取 ConfigMap 列表失败:', error)
    errorMsg.value = error.message || '获取 ConfigMap 列表失败'
  } finally {
    loading.value = false
  }
}

// 刷新列表
const refreshList = () => {
  fetchConfigMaps()
}

// 搜索输入处理
const onSearchInput = () => {
  currentPage.value = 1
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  try {
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time
  }
}

// 分页处理
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

// ==================== 批量操作 ====================
const enterBatchMode = () => {
  batchMode.value = true
  selectedConfigMaps.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedConfigMaps.value = []
}

const isConfigMapSelected = (cm) => {
  return selectedConfigMaps.value.some(
    selected => selected.name === cm.name && selected.namespace === cm.namespace
  )
}

const toggleConfigMapSelection = (cm) => {
  const index = selectedConfigMaps.value.findIndex(
    selected => selected.name === cm.name && selected.namespace === cm.namespace
  )
  if (index > -1) {
    selectedConfigMaps.value.splice(index, 1)
  } else {
    selectedConfigMaps.value.push(cm)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    // 取消全选
    paginatedConfigMaps.value.forEach(cm => {
      const index = selectedConfigMaps.value.findIndex(
        selected => selected.name === cm.name && selected.namespace === cm.namespace
      )
      if (index > -1) {
        selectedConfigMaps.value.splice(index, 1)
      }
    })
  } else {
    // 全选
    paginatedConfigMaps.value.forEach(cm => {
      if (!isConfigMapSelected(cm)) {
        selectedConfigMaps.value.push(cm)
      }
    })
  }
}

const clearSelection = () => {
  selectedConfigMaps.value = []
}

const openBatchDeleteModal = () => {
  if (selectedConfigMaps.value.length === 0) {
    alert('请先选择要删除的 ConfigMap')
    return
  }
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

  for (const cm of selectedConfigMaps.value) {
    try {
      await configmapApi.deleteConfigMap({
        namespace: cm.namespace,
        name: cm.name
      })
      successCount++
    } catch (error) {
      console.error(`删除 ${cm.namespace}/${cm.name} 失败:`, error)
      failCount++
    }
  }

  batchExecuting.value = false
  closeBatchDeleteModal()

  alert(`批量删除完成\n成功: ${successCount} 个\n失败: ${failCount} 个`)

  if (successCount > 0) {
    selectedConfigMaps.value = []
    fetchConfigMaps()
  }
}

// ==================== ✅ 新增：命名空间创建 ====================

const openCreateNamespaceModal = () => {
  newNamespaceName.value = ''
  createNamespaceError.value = ''
  showCreateNamespaceModal.value = true
}

const closeCreateNamespaceModal = () => {
  showCreateNamespaceModal.value = false
  newNamespaceName.value = ''
  createNamespaceError.value = ''
}

const createNamespace = async () => {
  const name = newNamespaceName.value.trim()

  if (!name) {
    createNamespaceError.value = '命名空间名称不能为空'
    return
  }

  // k8s namespace DNS label：小写字母/数字/-，不能以-开头结尾，<=63
  const nsReg = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
  if (!nsReg.test(name) || name.length > 63) {
    createNamespaceError.value =
      '命名空间格式不合法：只能包含小写字母、数字和 -，不能以 - 开头或结尾，长度 ≤ 63'
    return
  }

  creatingNamespace.value = true
  createNamespaceError.value = ''

  try {
    // 你后端约定：POST /namespace/create 之类
    // 如果你的参数不是 {name}，改这里即可
    const res = await namespaceApi.create({ name })

    if (res?.code !== undefined && res.code !== 0) {
      createNamespaceError.value = res.msg || '创建命名空间失败'
      return
    }

    // 成功：加入下拉 & 自动选中
    namespaces.value = Array.from(new Set([...namespaces.value, name])).sort()
    configMapForm.value.namespace = name

    closeCreateNamespaceModal()
    alert(`命名空间 ${name} 创建成功`)
  } catch (e) {
    console.error('创建命名空间失败:', e)
    createNamespaceError.value = e?.msg || e?.message || '创建命名空间失败'
  } finally {
    creatingNamespace.value = false
  }
}

// ==================== CRUD 操作 ====================
// 创建/编辑 ConfigMap
const submitConfigMap = async () => {
  if (!configMapForm.value.name) {
    alert('请输入 ConfigMap 名称')
    return
  }

  const data = {}
  configMapForm.value.data.forEach(item => {
    if (item.key) data[item.key] = item.value
  })

  creating.value = true
  try {
    if (editMode.value) {
      await configmapApi.updateData({
        namespace: configMapForm.value.namespace,
        name: configMapForm.value.name,
        data
      })
      alert('ConfigMap 更新成功')
    } else {
      await configmapApi.create({
        namespace: configMapForm.value.namespace,
        name: configMapForm.value.name,
        data
      })
      alert('ConfigMap 创建成功')
    }
    closeCreateModal()
    fetchConfigMaps()
  } catch (error) {
    console.error('操作失败:', error)
    alert(error.message || '操作失败')
  } finally {
    creating.value = false
  }
}

// 编辑 ConfigMap
const editConfigMap = async (cm) => {
  editMode.value = true
  loadingDetail.value = true

  try {
    const res = await configmapApi.detail({
      namespace: cm.namespace,
      name: cm.name
    })

    if (res.code === 0 && res.data) {
      const data = res.data.data || {}
      configMapForm.value = {
        name: cm.name,
        namespace: cm.namespace,
        data: Object.keys(data).length > 0
          ? Object.entries(data).map(([key, value]) => ({ key, value }))
          : [{ key: '', value: '' }]
      }
      showCreateModal.value = true
    }
  } catch (error) {
    console.error('获取详情失败:', error)
    alert('获取详情失败')
  } finally {
    loadingDetail.value = false
  }
}

// 查看详情
const viewDetail = async (cm) => {
  selectedConfigMap.value = cm
  showDetailModal.value = true
  loadingDetail.value = true
  detailData.value = null

  try {
    const res = await configmapApi.detail({
      namespace: cm.namespace,
      name: cm.name
    })

    if (res.code === 0 && res.data) {
      detailData.value = {
        name: res.data.name,
        namespace: res.data.namespace,
        dataCount: res.data.data ? Object.keys(res.data.data).length : 0,
        createdAt: res.data.created_at || res.data.createdAt,
        data: res.data.data || {}
      }
    }
  } catch (error) {
    console.error('获取详情失败:', error)
    alert('获取详情失败')
  } finally {
    loadingDetail.value = false
  }
}

// 确认删除
const confirmDelete = (cm) => {
  selectedConfigMap.value = cm
  showDeleteModal.value = true
}

// 删除 ConfigMap
const deleteConfigMap = async () => {
  deleting.value = true
  try {
    await configmapApi.deleteConfigMap({
      namespace: selectedConfigMap.value.namespace,
      name: selectedConfigMap.value.name
    })
    alert('ConfigMap 删除成功')
    showDeleteModal.value = false
    fetchConfigMaps()
  } catch (error) {
    console.error('删除失败:', error)
    alert(error.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

// ==================== YAML 操作 ====================
const viewYaml = async (cm) => {
  selectedYamlConfigMap.value = cm
  showYamlModal.value = true
  loadingYaml.value = true
  yamlViewContent.value = ''
  yamlViewError.value = ''
  yamlEditMode.value = false

  try {
    const res = await configmapApi.getYaml({
      namespace: cm.namespace,
      name: cm.name
    })

    if (res.code === 0 && res.data) {
      yamlViewContent.value = res.data.yaml || res.data
    }
  } catch (error) {
    console.error('获取 YAML 失败:', error)
    yamlViewError.value = error.message || '获取 YAML 失败'
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  yamlViewContent.value = ''
  yamlEditMode.value = false
  yamlViewError.value = ''
}

const downloadYaml = () => {
  const blob = new Blob([yamlViewContent.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedYamlConfigMap.value.name}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

const applyYamlChanges = async () => {
  savingYaml.value = true
  try {
    await configmapApi.applyYaml({
      yaml: yamlViewContent.value
    })
    alert('YAML 应用成功')
    closeYamlModal()
    fetchConfigMaps()
  } catch (error) {
    console.error('应用 YAML 失败:', error)
    alert(error.message || '应用 YAML 失败')
  } finally {
    savingYaml.value = false
  }
}

// ==================== 表单操作 ====================
const addDataItem = () => {
  configMapForm.value.data.push({ key: '', value: '' })
}

const removeDataItem = (index) => {
  if (configMapForm.value.data.length > 1) {
    configMapForm.value.data.splice(index, 1)
  }
}

// ==================== YAML 创建功能 ====================
const loadMultiResourceYamlTemplate = () => {
  yamlCreateContent.value = `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
  labels:
    app: multi-example
data:
  application.properties: |
    server.port=8080
    spring.profiles.active=prod
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: default
  labels:
    app: multi-example
spec:
  selector:
    app: multi-example
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: default
  labels:
    app: multi-example
spec:
  replicas: 2
  selector:
    matchLabels:
      app: multi-example
  template:
    metadata:
      labels:
        app: multi-example
    spec:
      containers:
      - name: app
        image: nginx:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
`
  yamlCreateError.value = ''
  multiResourcePreview.value = { resources: [], dependencies: [], errors: [], total: 0 }
  alert('已加载多资源 YAML 模板，请修改后创建')
}

const loadConfigMapYamlTemplate = () => {
  yamlCreateContent.value = `apiVersion: v1
kind: ConfigMap
metadata:
  name: example-configmap
  namespace: default
  labels:
    app: example
data:
  database.host: "mysql.default.svc.cluster.local"
  database.port: "3306"
  database.name: "myapp"

  application.yaml: |
    server:
      port: 8080
      host: 0.0.0.0

    logging:
      level: INFO
      format: json

  nginx.conf: |
    server {
      listen 80;
      server_name example.com;

      location / {
        proxy_pass http://backend:8080;
      }
    }`
  yamlCreateError.value = ''
  multiResourcePreview.value = { resources: [], dependencies: [], errors: [], total: 0 }
  alert('已加载 ConfigMap YAML 模板，请修改后创建')
}

const copyYamlContent = () => {
  if (!yamlCreateContent.value) {
    alert('没有可复制的内容')
    return
  }
  navigator.clipboard.writeText(yamlCreateContent.value)
    .then(() => alert('已复制到剪贴板'))
    .catch(() => alert('复制失败'))
}

const resetYamlContent = () => {
  if (!yamlCreateContent.value || confirm('确定要清空当前 YAML 内容吗？')) {
    yamlCreateContent.value = ''
    yamlCreateError.value = ''
  }
}

const createConfigMapFromYaml = async () => {
  if (!yamlCreateContent.value.trim()) {
    alert('请输入 YAML 内容')
    return
  }

  try {
    if (!yamlCreateContent.value.includes('kind: ConfigMap')) {
      yamlCreateError.value = 'YAML 中必须包含 "kind: ConfigMap"'
      return
    }
    if (!yamlCreateContent.value.includes('apiVersion: v1')) {
      yamlCreateError.value = 'YAML 中必须包含 "apiVersion: v1"'
      return
    }
    yamlCreateError.value = ''
  } catch (e) {
    yamlCreateError.value = `YAML 格式错误: ${e.message}`
    return
  }

  creating.value = true
  try {
    const res = await configmapApi.applyYaml({ yaml: yamlCreateContent.value })
    if (res.code === 0) {
      alert('ConfigMap 应用成功！')
      showCreateModal.value = false
      resetConfigMapForm()
      await fetchConfigMaps()
    } else {
      alert(res.msg || '应用失败')
      yamlCreateError.value = res.msg || '应用失败'
    }
  } catch (e) {
    const msg = e?.msg || e?.message || '应用失败'
    alert(msg)
    yamlCreateError.value = msg
  } finally {
    creating.value = false
  }
}

const parseMultiResourceYaml = async (yamlContent) => {
  try {
    const res = await configmapApi.parseMultiYaml({ yaml: yamlContent })
    if (res.code === 0) {
      multiResourcePreview.value = res.data

      const dependencies = []
      res.data.resources.forEach((resource, index) => {
        if (resource.dependsOn && resource.dependsOn.length > 0) {
          resource.dependsOn.forEach(dep => {
            dependencies.push({
              id: `${index}-${dep.kind}-${dep.name}`,
              source: { kind: resource.kind, name: resource.name },
              target: { kind: dep.kind, name: dep.name }
            })
          })
        }
      })
      multiResourcePreview.value.dependencies = dependencies
    } else {
      multiResourcePreview.value = { resources: [], dependencies: [], errors: [res.msg || '解析失败'], total: 0 }
    }
  } catch (error) {
    multiResourcePreview.value = { resources: [], dependencies: [], errors: [error?.msg || error?.message || '解析失败'], total: 0 }
  }
}

const applyMultiResourceYaml = async () => {
  if (!yamlCreateContent.value.trim()) {
    alert('请输入 YAML 内容')
    return
  }

  if (multiResourcePreview.value.errors.length > 0) {
    alert('存在依赖错误，请修正后再创建:\n' + multiResourcePreview.value.errors.join('\n'))
    return
  }

  creating.value = true
  try {
    const res = await configmapApi.applyMultiYaml({ yaml: yamlCreateContent.value })
    if (res.code === 0) {
      const createdCount = res.data.created?.length || 0
      const failedCount = res.data.failed?.length || 0

      let message = `多资源创建完成！\n`
      message += `✅ 成功创建: ${createdCount} 个资源\n`
      if (failedCount > 0) {
        message += `❌ 创建失败: ${failedCount} 个资源\n\n`
        res.data.failed.forEach(fail => {
          message += `- ${fail.kind}/${fail.name}: ${fail.error}\n`
        })
      }

      alert(message)

      if (createdCount > 0) {
        showCreateModal.value = false
        resetConfigMapForm()
        await fetchConfigMaps()
      }
    } else {
      alert(res.msg || '多资源创建失败')
    }
  } catch (error) {
    alert(error?.msg || error?.message || '多资源创建失败')
  } finally {
    creating.value = false
  }
}

const getResourceIcon = (kind) => {
  const icons = {
    'ConfigMap': '📋',
    'Secret': '🔒',
    'Service': '🌐',
    'Deployment': '🚀',
    'StatefulSet': '🏛️',
    'DaemonSet': '🔁',
    'Job': '⚙️',
    'CronJob': '⏰',
    'PersistentVolumeClaim': '💾',
    'Namespace': '📂'
  }
  return icons[kind] || '📦'
}

const getResourceKindClass = (kind) => {
  const classes = {
    'ConfigMap': 'kind-configmap',
    'Secret': 'kind-secret',
    'Service': 'kind-service',
    'Deployment': 'kind-deployment',
    'StatefulSet': 'kind-statefulset',
    'DaemonSet': 'kind-daemonset',
    'Job': 'kind-job',
    'CronJob': 'kind-cronjob',
    'PersistentVolumeClaim': 'kind-pvc',
    'Namespace': 'kind-namespace'
  }
  return classes[kind] || 'kind-default'
}

// 重置表单
const resetConfigMapForm = () => {
  configMapForm.value = {
    name: '',
    namespace: 'default',
    data: [{ key: '', value: '' }]
  }
  yamlCreateContent.value = ''
  yamlCreateError.value = ''
  createMode.value = 'form'
  editMode.value = false

  multiResourcePreview.value = { resources: [], dependencies: [], errors: [], total: 0 }
}

const closeCreateModal = () => {
  showCreateModal.value = false
  resetConfigMapForm()
}

// ==================== 生命周期 ====================
onMounted(async () => {
  // 推荐：先拉 ns（如果你后端没有 list 会自动跳过）
  await fetchNamespaces()
  fetchConfigMaps()
})

// 监听 YAML 内容变化，实现实时预览
watch(yamlCreateContent, async (newVal) => {
  if (newVal && newVal.trim().length > 10) {
    clearTimeout(window.multiYamlPreviewTimer)
    window.multiYamlPreviewTimer = setTimeout(async () => {
      await parseMultiResourceYaml(newVal)
    }, 1000)
  } else {
    multiResourcePreview.value = { resources: [], dependencies: [], errors: [], total: 0 }
  }
})

// 监听自动刷新
watch(autoRefresh, (newVal) => {
  if (newVal) {
    refreshTimer = setInterval(() => {
      fetchConfigMaps()
    }, 90000) // 90秒刷新
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>


<style scoped>
/* 多资源预览样式 */
.multi-resource-preview {
  margin-top: 20px;
  padding: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e2e8f0;
}

.preview-header h4 {
  margin: 0;
  color: #1e293b;
  font-size: 16px;
  font-weight: 600;
}

.dependencies-info {
  margin-bottom: 16px;
}

.dependencies-info h5 {
  margin: 0 0 12px 0;
  color: #334155;
  font-size: 14px;
  font-weight: 500;
}

.dependency-graph {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dependency-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}

.dep-source {
  font-weight: 500;
  color: #0f172a;
}

.dep-arrow {
  margin: 0 8px;
  color: #94a3b8;
}

.dep-target {
  color: #475569;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.resource-item {
  padding: 12px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  transition: all 0.2s;
}

.resource-item:hover {
  border-color: #cbd5e1;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.resource-item.has-warnings {
  border-left: 4px solid #f59e0b;
  background: #fef3c7;
}

.resource-item.has-errors {
  border-left: 4px solid #ef4444;
  background: #fee2e2;
}

.resource-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.resource-kind {
  font-weight: 600;
  font-size: 13px;
  padding: 2px 8px;
  border-radius: 4px;
}

.resource-kind.kind-configmap { color: #2563eb; background: #dbeafe; }
.resource-kind.kind-secret { color: #7e22ce; background: #f3e8ff; }
.resource-kind.kind-service { color: #0891b2; background: #cffafe; }
.resource-kind.kind-deployment { color: #059669; background: #dcfce7; }
.resource-kind.kind-statefulset { color: #0d9488; background: #ccfbf1; }
.resource-kind.kind-daemonset { color: #c2410c; background: #fed7aa; }
.resource-kind.kind-job { color: #9333ea; background: #f5d0fe; }
.resource-kind.kind-cronjob { color: #a855f7; background: #ede9fe; }
.resource-kind.kind-pvc { color: #0f766e; background: #ccfbf1; }
.resource-kind.kind-namespace { color: #4338ca; background: #e0e7ff; }
.resource-kind.kind-default { color: #374151; background: #f3f4f6; }

.resource-name {
  font-weight: 500;
  color: #1f2937;
}

.resource-namespace {
  color: #6b7280;
  font-size: 12px;
}

.resource-order {
  background: #e5e7eb;
  color: #374151;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.resource-warnings {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #fbbf24;
}

.warning-item {
  font-size: 12px;
  color: #b45309;
  margin-bottom: 4px;
}

.preview-errors {
  padding: 12px;
  background: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 6px;
}

.preview-errors h5 {
  margin: 0 0 8px 0;
  color: #b91c1c;
  font-size: 14px;
  font-weight: 500;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.error-item {
  font-size: 13px;
  color: #b91c1c;
  padding: 6px 10px;
  background: white;
  border-radius: 4px;
}

/* 按钮样式增强 */
.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover {
  background: #059669;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 12px;
}

/* YAML 提示样式更新 */
.yaml-hint code {
  background: #f1f5f9;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}

.yaml-tips ul {
  margin: 8px 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin: 4px 0;
  font-size: 13px;
  color: #475569;
}

.yaml-tips code {
  background: #f1f5f9;
  padding: 1px 3px;
  border-radius: 3px;
  font-family: monospace;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .preview-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .resource-header {
    flex-wrap: wrap;
  }

  .dependency-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .dep-arrow {
    transform: rotate(90deg);
  }
}
</style>

<style scoped>
/* 复用 Deployment 的样式 */
.resource-view {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 8px;
}

.view-header p {
  color: #6b7280;
  font-size: 14px;
}

.action-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 200px;
}

.search-box input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.filter-dropdown select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background: white;
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
}

.btn-batch {
  background: #8b5cf6;
  color: white;
}

.btn-batch:hover {
  background: #7c3aed;
}

.view-toggle {
  display: flex;
  gap: 4px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 6px;
}

.btn-view {
  padding: 6px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 4px;
  font-size: 16px;
}

.btn-view.active {
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  cursor: pointer;
}

.refresh-indicator {
  color: #22c55e;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.error-box {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  margin-bottom: 16px;
  font-size: 14px;
}

/* 批量操作浮动栏 */
.batch-action-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px 20px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  display: flex;
  gap: 20px;
  align-items: center;
  z-index: 100;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-count {
  font-weight: 600;
  color: #1f2937;
}

.batch-clear {
  padding: 4px 8px;
  background: #f3f4f6;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}

.batch-actions {
  display: flex;
  gap: 8px;
}

.batch-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  background: #f3f4f6;
  color: #374151;
}

.batch-btn:hover {
  background: #e5e7eb;
}

.batch-btn.danger {
  background: #fee2e2;
  color: #dc2626;
}

.batch-btn.danger:hover {
  background: #fecaca;
}

/* 表格样式 */
.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
}

.resource-table thead {
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.resource-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.resource-table tbody tr {
  border-bottom: 1px solid #f3f4f6;
  transition: background 0.2s;
}

.resource-table tbody tr:hover {
  background: #f9fafb;
}

.resource-table tbody tr.row-selected {
  background: #eff6ff;
}

.resource-table td {
  padding: 12px 16px;
  font-size: 14px;
  color: #1f2937;
}

.loading-row,
.empty-row {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.loading-spinner {
  border: 3px solid #f3f4f6;
  border-top: 3px solid #3b82f6;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  animation: spin 1s linear infinite;
  display: inline-block;
  margin-right: 8px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.configmap-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon {
  font-size: 18px;
}

.namespace-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #eff6ff;
  color: #1d4ed8;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.data-count {
  color: #6b7280;
  font-size: 13px;
}

.timestamp {
  color: #6b7280;
  font-size: 13px;
}

.action-btns {
  display: flex;
  gap: 8px;
}

.btn-icon {
  padding: 4px 8px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 16px;
  border-radius: 4px;
  transition: background 0.2s;
}

.btn-icon:hover {
  background: #f3f4f6;
}

.btn-icon.danger:hover {
  background: #fee2e2;
}

/* 卡片视图 */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.resource-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s;
  position: relative;
}

.resource-card:hover {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.resource-card.card-selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.card-checkbox {
  position: absolute;
  top: 12px;
  right: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title h3 {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.card-body {
  margin-bottom: 12px;
}

.card-info {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}

.info-label {
  color: #6b7280;
}

.info-value {
  color: #1f2937;
  font-weight: 500;
}

.card-footer {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #f3f4f6;
}

.btn-card {
  flex: 1;
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  background: white;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-card:hover {
  background: #f9fafb;
}

.btn-card.danger {
  color: #dc2626;
  border-color: #fecaca;
}

.btn-card.danger:hover {
  background: #fef2f2;
}

.loading-card,
.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
}

/* 模态框样式 */
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
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.1);
}

.modal-content.large {
  max-width: 900px;
}

.modal-content.yaml-modal {
  max-width: 1200px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 24px;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e5e7eb;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
}

.form-input,
.form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 数据编辑器 */
.data-editor {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px;
}

.data-row {
  display: grid;
  grid-template-columns: 200px 1fr 40px;
  gap: 8px;
  margin-bottom: 8px;
  align-items: start;
}

.data-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.data-textarea {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  font-family: 'Courier New', monospace;
  resize: vertical;
  min-height: 60px;
}

.btn-remove {
  padding: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 4px;
  font-size: 16px;
}

.btn-remove:hover {
  background: #fee2e2;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

/* 详情显示 */
.loading-state {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 12px;
}

.detail-table {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.detail-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #f3f4f6;
  font-size: 14px;
}

.detail-table td.label {
  font-weight: 600;
  color: #6b7280;
  width: 150px;
  background: #f9fafb;
}

.data-display {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.data-item {
  border-bottom: 1px solid #f3f4f6;
}

.data-item:last-child {
  border-bottom: none;
}

.data-item-key {
  padding: 8px 12px;
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
  font-size: 13px;
  border-bottom: 1px solid #e5e7eb;
}

.data-item-value {
  padding: 12px;
  margin: 0;
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

/* YAML 编辑器 */
.yaml-header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.yaml-modal-body {
  padding: 0;
  min-height: 400px;
  max-height: 600px;
}

.yaml-editor-wrapper {
  height: 100%;
}

.yaml-editor,
.yaml-content {
  width: 100%;
  height: 600px;
  padding: 16px;
  border: none;
  background: #1e293b;
  color: #e2e8f0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow: auto;
  margin: 0;
}

.yaml-editor:focus {
  outline: none;
}

/* 删除确认 */
.delete-confirm {
  max-width: 500px;
}

.warning-text {
  color: #dc2626;
  font-size: 14px;
}

/* 批量删除 */
.batch-delete-modal {
  max-width: 600px;
}

.batch-delete-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  background: #f9fafb;
}

.batch-delete-item {
  padding: 8px;
  margin-bottom: 4px;
  background: white;
  border-radius: 4px;
  font-size: 14px;
}

.confirm-input-group {
  margin-top: 16px;
}

.confirm-input-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.confirm-input-group code {
  padding: 2px 6px;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  font-size: 13px;
}

/* ==================== YAML 编辑器样式 ====================  */
/* 模式切换按钮 */
.mode-toggle {
  display: flex;
  gap: 4px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 8px;
}

.view-toggle-btn {
  padding: 6px 16px;
  border: none;
  background: transparent;
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.view-toggle-btn:hover {
  background: rgba(255, 255, 255, 0.5);
  color: #374151;
}

.view-toggle-btn.active {
  background: white;
  color: #3b82f6;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  font-weight: 600;
}

/* 可滚动模态框主体 */
.modal-body.scrollable {
  max-height: calc(90vh - 200px);
  overflow-y: auto;
}

/* YAML 编辑器容器 */
.yaml-editor-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.yaml-editor-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.yaml-editor-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.yaml-hint {
  margin: 0;
  padding: 10px 14px;
  background: #eff6ff;
  border-left: 3px solid #3b82f6;
  border-radius: 4px;
  font-size: 13px;
  color: #1e40af;
  line-height: 1.5;
}

.yaml-hint code {
  padding: 2px 6px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-weight: 600;
}

.yaml-header-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.load-template-btn,
.copy-yaml-btn,
.reset-yaml-btn {
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  font-size: 13px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.load-template-btn:hover {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.copy-yaml-btn:hover {
  background: #10b981;
  border-color: #10b981;
  color: white;
}

.reset-yaml-btn:hover {
  background: #f59e0b;
  border-color: #f59e0b;
  color: white;
}

/* YAML 编辑器 */
.yaml-editor {
  width: 100%;
  min-height: 400px;
  max-height: 600px;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background: #1e1e1e;
  color: #d4d4d4;
  resize: vertical;
  box-sizing: border-box;
  tab-size: 2;
}

.yaml-editor:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.yaml-editor::placeholder {
  color: #6b7280;
}

/* YAML 错误提示 */
.yaml-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 13px;
  line-height: 1.5;
}

.error-icon {
  font-size: 18px;
  flex-shrink: 0;
}

/* YAML 编辑器底部 */
.yaml-editor-footer {
  padding: 14px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.yaml-tips {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.6;
}

.yaml-tips strong {
  color: #374151;
  font-weight: 600;
}

.yaml-tips ul {
  margin: 8px 0 0 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin-bottom: 4px;
}

/* YAML 查看模态框 */
.yaml-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.yaml-modal-body {
  max-height: calc(80vh - 140px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.yaml-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 8px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
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
</style>

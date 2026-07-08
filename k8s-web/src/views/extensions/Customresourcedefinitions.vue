<template>
  <div class="crd-management">
    <!-- ========== 顶部面包屑 & 标题 ========== -->
    <div class="page-header">
      <div class="header-left">
        <div class="breadcrumb" v-if="activeView === 'cr-list'">
          <span class="bc-link" @click="backToCRDList">CRD 管理</span>
          <span class="bc-sep">/</span>
          <span class="bc-current">{{ activeCRD?.kind }} 实例</span>
        </div>
        <h1 v-else class="page-title">
          <span class="title-icon">🧩</span>
          Custom Resource Definitions
        </h1>
        <p class="page-desc" v-if="activeView === 'crd-list'">管理集群中的自定义资源定义，查看和操作 CR 实例</p>
      </div>
      <div class="header-actions">
        <button class="btn-icon" @click="fetchCRDs" :disabled="loading" title="刷新">
          <span :class="{ spinning: loading }">⟳</span>
        </button>
      </div>
    </div>

    <!-- ========== CRD 列表视图 ========== -->
    <div v-if="activeView === 'crd-list'" class="crd-list-view">
      <!-- 统计卡片 -->
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-value">{{ crds.length }}</div>
          <div class="stat-label">CRD 总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ uniqueGroups.length }}</div>
          <div class="stat-label">API Groups</div>
        </div>
        <div class="stat-card accent">
          <div class="stat-value">{{ namespacedCount }}</div>
          <div class="stat-label">Namespaced</div>
        </div>
        <div class="stat-card purple">
          <div class="stat-value">{{ clusterCount }}</div>
          <div class="stat-label">Cluster Scope</div>
        </div>
      </div>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <div class="search-wrapper">
            <span class="search-icon">🔍</span>
            <input
              type="text"
              v-model="searchQuery"
              placeholder="按名称、Kind 或 Group 搜索..."
              class="search-input"
            />
            <span v-if="searchQuery" class="search-clear" @click="searchQuery = ''">✕</span>
          </div>
          <div class="filter-chips">
            <button
              class="chip"
              :class="{ active: scopeFilter === '' }"
              @click="scopeFilter = ''"
            >全部</button>
            <button
              class="chip"
              :class="{ active: scopeFilter === 'Namespaced' }"
              @click="scopeFilter = 'Namespaced'"
            >Namespaced</button>
            <button
              class="chip"
              :class="{ active: scopeFilter === 'Cluster' }"
              @click="scopeFilter = 'Cluster'"
            >Cluster</button>
          </div>
          <select v-model="groupFilter" class="group-select">
            <option value="">所有 Group</option>
            <option v-for="g in uniqueGroups" :key="g" :value="g">{{ g }}</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn-primary" @click="openCreateCRD">
            <span class="btn-icon-inner">+</span> 创建 CRD
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading && crds.length === 0" class="loading-state">
        <div class="loader"></div>
        <p>正在加载 CRD 列表...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="!loading && filteredCRDs.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <h3>{{ searchQuery || scopeFilter || groupFilter ? '没有匹配的 CRD' : '暂无 CRD' }}</h3>
        <p>{{ searchQuery || scopeFilter || groupFilter ? '尝试调整搜索条件' : '当前集群未安装任何自定义资源定义' }}</p>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="errorMsg" class="error-state">
        <div class="error-icon">⚠️</div>
        <h3>获取失败</h3>
        <p>{{ errorMsg }}</p>
        <button class="btn-outline" @click="fetchCRDs">重试</button>
      </div>

      <!-- CRD 表格 -->
      <div v-else class="table-wrapper">
        <table class="crd-table">
          <thead>
            <tr>
              <th class="col-name">名称</th>
              <th class="col-group">API Group</th>
              <th class="col-kind">Kind</th>
              <th class="col-version">版本</th>
              <th class="col-scope">作用域</th>
              <th class="col-status">状态</th>
              <th class="col-time">创建时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="crd in paginatedCRDs"
              :key="crd.name"
              class="crd-row"
              @click="viewCRInstances(crd)"
            >
              <td class="col-name">
                <div class="name-cell">
                  <span class="crd-icon">📋</span>
                  <div>
                    <div class="name-primary">{{ crd.name }}</div>
                  </div>
                </div>
              </td>
              <td class="col-group">
                <span class="group-badge">{{ crd.group }}</span>
              </td>
              <td class="col-kind">
                <code class="kind-code">{{ crd.kind }}</code>
              </td>
              <td class="col-version">
                <span class="version-tag" v-for="v in (crd.versions || [crd.version])" :key="v">{{ v }}</span>
              </td>
              <td class="col-scope">
                <span class="scope-badge" :class="crd.scope?.toLowerCase()">{{ crd.scope }}</span>
              </td>
              <td class="col-status">
                <span class="status-dot" :class="crd.status?.toLowerCase()"></span>
                <span class="status-text">{{ crd.status }}</span>
              </td>
              <td class="col-time">{{ formatTime(crd.createdAt) }}</td>
              <td class="col-actions" @click.stop>
                <div class="action-group">
                  <button class="action-btn-text primary" @click="viewCRInstances(crd)" title="查看 CR 实例">
                    CR 实例
                  </button>
                  <button class="action-btn-text default" @click="viewCRDYaml(crd)" title="查看 YAML">
                    YAML
                  </button>
                  <button class="action-btn-text danger" @click="confirmDeleteCRD(crd)" title="删除 CRD">
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="filteredCRDs.length > pageSize" class="pagination-bar">
        <span class="page-info">共 {{ filteredCRDs.length }} 条</span>
        <div class="page-controls">
          <button class="page-btn" :disabled="currentPage <= 1" @click="currentPage--">‹</button>
          <template v-for="p in totalPages" :key="p">
            <button class="page-btn" :class="{ active: p === currentPage }" @click="currentPage = p">{{ p }}</button>
          </template>
          <button class="page-btn" :disabled="currentPage >= totalPages" @click="currentPage++">›</button>
        </div>
        <select v-model="pageSize" class="page-size-select">
          <option :value="10">10 条/页</option>
          <option :value="20">20 条/页</option>
          <option :value="50">50 条/页</option>
        </select>
      </div>
    </div>

    <!-- ========== CR 实例列表视图 ========== -->
    <div v-if="activeView === 'cr-list'" class="cr-list-view">
      <!-- CR 头部信息 -->
      <div class="cr-header-card">
        <div class="cr-header-info">
          <div class="cr-meta">
            <span class="cr-meta-item"><strong>Kind:</strong> {{ activeCRD?.kind }}</span>
            <span class="cr-meta-item"><strong>Group:</strong> {{ activeCRD?.group }}</span>
            <span class="cr-meta-item"><strong>Version:</strong> {{ activeCRD?.version }}</span>
            <span class="cr-meta-item"><strong>Scope:</strong>
              <span class="scope-badge" :class="activeCRD?.scope?.toLowerCase()">{{ activeCRD?.scope }}</span>
            </span>
          </div>
        </div>
        <div class="cr-header-actions">
          <button class="btn-primary" @click="openCreateCR">
            <span class="btn-icon-inner">+</span> 创建实例
          </button>
          <button class="btn-icon" @click="fetchCRInstances" :disabled="crLoading" title="刷新">
            <span :class="{ spinning: crLoading }">⟳</span>
          </button>
        </div>
      </div>

      <!-- CR 命名空间筛选 -->
      <div class="cr-toolbar" v-if="activeCRD?.scope === 'Namespaced'">
        <select v-model="crNamespaceFilter" class="ns-select" @change="fetchCRInstances">
          <option value="">所有命名空间</option>
          <option v-for="ns in crNamespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
        <input v-model="crSearchQuery" placeholder="搜索 CR 名称..." class="cr-search-input" />
      </div>

      <!-- CR 加载 -->
      <div v-if="crLoading" class="loading-state">
        <div class="loader"></div>
        <p>正在加载 CR 实例...</p>
      </div>

      <!-- CR 空状态 -->
      <div v-else-if="crInstances.length === 0" class="empty-state">
        <div class="empty-icon">📭</div>
        <h3>暂无 {{ activeCRD?.kind }} 实例</h3>
        <p>点击「创建实例」按钮来创建第一个自定义资源</p>
        <button class="btn-primary" @click="openCreateCR">创建实例</button>
      </div>

      <!-- CR 表格 -->
      <div v-else class="table-wrapper">
        <table class="crd-table cr-table">
          <thead>
            <tr>
              <th>名称</th>
              <th v-if="activeCRD?.scope === 'Namespaced'">命名空间</th>
              <th>创建时间</th>
              <th>标签</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cr in filteredCRInstances" :key="cr.name + cr.namespace">
              <td>
                <div class="name-cell">
                  <span class="crd-icon">📄</span>
                  <span class="name-primary">{{ cr.name }}</span>
                </div>
              </td>
              <td v-if="activeCRD?.scope === 'Namespaced'">
                <span class="ns-badge">{{ cr.namespace }}</span>
              </td>
              <td>{{ formatTime(cr.createdAt) }}</td>
              <td>
                <div class="label-tags">
                  <span v-for="(val, key) in (cr.labels || {})" :key="key" class="label-tag">
                    {{ key }}={{ val }}
                  </span>
                  <span v-if="!cr.labels || Object.keys(cr.labels).length === 0" class="no-labels">-</span>
                </div>
              </td>
              <td @click.stop>
                <div class="action-group">
                  <button class="action-btn-text primary" @click="viewCRYaml(cr)" title="查看 YAML">
                    YAML
                  </button>
                  <button class="action-btn-text default" @click="editCR(cr)" title="编辑">
                    编辑
                  </button>
                  <button class="action-btn-text danger" @click="confirmDeleteCR(cr)" title="删除">
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ========== YAML 查看器抽屉 ========== -->
    <Teleport to="body">
      <transition name="drawer">
        <div v-if="showYamlDrawer" class="drawer-overlay" @click="showYamlDrawer = false">
          <div class="drawer-panel" @click.stop>
            <div class="drawer-header">
              <h3>{{ yamlDrawerTitle }}</h3>
              <button class="drawer-close" @click="showYamlDrawer = false">✕</button>
            </div>
            <div class="drawer-body">
              <YamlHighlight :content="yamlContent" :title="yamlDrawerTitle" :showLineNumbers="true" maxHeight="calc(100vh - 140px)" />
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- ========== YAML 编辑器模态框（创建/编辑）========== -->
    <Teleport to="body">
      <transition name="modal">
        <div v-if="showEditorModal" class="modal-overlay" @click="showEditorModal = false">
          <div class="editor-modal-pro" @click.stop>
            <!-- 顶部状态栏 -->
            <div class="editor-topbar">
              <div class="editor-topbar-left">
                <span class="editor-badge">
                  <span class="editor-badge-icon">{{ editorMode === 'create' ? '＋' : '✎' }}</span>
                  {{ editorMode === 'create' ? 'CREATE' : 'EDIT' }}
                </span>
                <h3 class="editor-title-pro">{{ editorTitle }}</h3>
              </div>
              <div class="editor-topbar-right">
                <button class="editor-topbar-btn" @click="showEditorModal = false" title="关闭">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
            </div>

            <!-- 资源信息条 -->
            <div class="editor-info-bar">
              <div class="editor-info-tags">
                <span class="info-tag" v-if="activeCRD && activeView === 'cr-list'">
                  <span class="info-tag-label">Kind</span>
                  <span class="info-tag-value">{{ activeCRD.kind }}</span>
                </span>
                <span class="info-tag" v-if="activeCRD && activeView === 'cr-list'">
                  <span class="info-tag-label">Group</span>
                  <span class="info-tag-value">{{ activeCRD.group }}</span>
                </span>
                <span class="info-tag" v-if="activeCRD && activeView === 'cr-list'">
                  <span class="info-tag-label">Version</span>
                  <span class="info-tag-value">{{ activeCRD.version }}</span>
                </span>
                <span class="info-tag" v-if="activeView === 'crd-list'">
                  <span class="info-tag-label">Resource</span>
                  <span class="info-tag-value">CustomResourceDefinition</span>
                </span>
              </div>
            </div>

            <!-- 工具栏 -->
            <div class="editor-toolbar">
              <div class="editor-toolbar-left">
                <select v-model="selectedTemplate" @change="applyTemplate" class="template-select">
                  <option value="">-- 选择模板 --</option>
                  <option v-for="tpl in availableTemplates" :key="tpl.name" :value="tpl.name">{{ tpl.label }}</option>
                </select>
                <button class="toolbar-action-btn" @click="formatYaml" title="格式化 YAML">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="21" y1="10" x2="3" y2="10"/><line x1="21" y1="6" x2="3" y2="6"/><line x1="21" y1="14" x2="3" y2="14"/><line x1="21" y1="18" x2="3" y2="18"/></svg>
                  格式化
                </button>
                <button class="toolbar-action-btn" @click="copyYaml" title="复制">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                  复制
                </button>
                <button class="toolbar-action-btn danger" @click="clearEditor" title="清空内容">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
                  清空
                </button>
              </div>
              <div class="editor-toolbar-right">
                <button
                  class="toolbar-dryrun-btn"
                  @click="doDryRun"
                  :disabled="dryRunning"
                  :class="{ running: dryRunning }"
                >
                  <svg v-if="!dryRunning" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2v-4M9 21H5a2 2 0 01-2-2v-4"/></svg>
                  <span v-if="dryRunning" class="dryrun-spinner"></span>
                  {{ dryRunning ? '校验中...' : 'DryRun 预校验' }}
                </button>
              </div>
            </div>

            <!-- DryRun 结果通知 -->
            <transition name="slide-down">
              <div v-if="dryRunResult" class="dryrun-banner" :class="dryRunResult.success ? 'success' : 'error'">
                <div class="dryrun-banner-icon">
                  <svg v-if="dryRunResult.success" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                  <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
                </div>
                <span class="dryrun-banner-text">{{ dryRunResult.message }}</span>
                <button class="dryrun-banner-close" @click="dryRunResult = null">✕</button>
              </div>
            </transition>

            <!-- 编辑区域（带行号） -->
            <div class="editor-workspace">
              <div class="editor-gutter">
                <div class="gutter-line" v-for="n in editorLineCount" :key="n">{{ n }}</div>
              </div>
              <textarea
                ref="editorTextarea"
                v-model="editorYaml"
                class="yaml-editor-pro"
                spellcheck="false"
                placeholder="# 在此输入 YAML 内容..."
                @scroll="syncScroll"
                @keydown="handleEditorKeydown"
              ></textarea>
            </div>

            <!-- 底部操作栏 -->
            <div class="editor-footer-pro">
              <div class="editor-footer-left">
                <span class="footer-stat">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                  {{ editorLineCount }} 行
                </span>
                <span class="footer-stat">{{ editorYaml.length }} 字符</span>
                <span class="footer-stat mode-tag">YAML</span>
              </div>
              <div class="editor-footer-right">
                <button class="btn-editor-cancel" @click="showEditorModal = false">
                  取消
                </button>
                <button class="btn-editor-submit" @click="submitEditor" :disabled="submitting">
                  <span v-if="submitting" class="submit-spinner"></span>
                  {{ submitting ? '提交中...' : editorMode === 'create' ? '确认创建' : '确认更新' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- ========== 删除确认 ========== -->
    <Teleport to="body">
      <transition name="modal">
        <div v-if="showDeleteConfirm" class="modal-overlay" @click="showDeleteConfirm = false">
          <div class="delete-dialog" @click.stop>
            <div class="delete-icon-area">
              <div class="delete-icon-circle">⚠️</div>
            </div>
            <h3 class="delete-title">确认删除</h3>
            <p class="delete-desc">
              确定要删除 <code>{{ deleteTarget?.name }}</code> 吗？
            </p>
            <p class="delete-warning" v-if="deleteType === 'crd'">
              此操作将同时删除该 CRD 下的所有 CR 实例，且不可恢复！
            </p>
            <p class="delete-warning" v-else>
              此操作不可恢复，请谨慎操作。
            </p>
            <div class="delete-actions">
              <button class="btn-cancel" @click="showDeleteConfirm = false">取消</button>
              <button class="btn-danger" @click="executeDelete" :disabled="deleting">
                {{ deleting ? '删除中...' : '确认删除' }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import crdApi from '@/api/cluster/extensions/crd'
import YamlHighlight from '@/components/YamlHighlight.vue'

const route = useRoute()

// ========== 视图状态 ==========
const activeView = ref('crd-list') // 'crd-list' | 'cr-list'
const activeCRD = ref(null)

// ========== CRD 列表 ==========
const crds = ref([])
const loading = ref(false)
const errorMsg = ref('')
const searchQuery = ref('')
const scopeFilter = ref('')
const groupFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

// ========== CR 实例 ==========
const crInstances = ref([])
const crLoading = ref(false)
const crNamespaceFilter = ref('')
const crSearchQuery = ref('')
const crNamespaces = ref([])

// ========== YAML 抽屉 ==========
const showYamlDrawer = ref(false)
const yamlContent = ref('')
const yamlDrawerTitle = ref('')

// ========== YAML 编辑器 ==========
const showEditorModal = ref(false)
const editorTitle = ref('')
const editorMode = ref('create') // 'create' | 'edit'
const editorYaml = ref('')
const editingCR = ref(null)
const submitting = ref(false)
const dryRunning = ref(false)
const dryRunResult = ref(null)
const selectedTemplate = ref('')
const editorTextarea = ref(null)

// 行号计算
const editorLineCount = computed(() => {
  return Math.max((editorYaml.value || '').split('\n').length, 1)
})

// Gutter 滚动同步
const syncScroll = (e) => {
  const gutter = e.target?.parentElement?.querySelector('.editor-gutter')
  if (gutter) gutter.scrollTop = e.target.scrollTop
}

// 编辑器快捷键
const handleEditorKeydown = (e) => {
  // Tab 插入 2 空格
  if (e.key === 'Tab') {
    e.preventDefault()
    const ta = e.target
    const start = ta.selectionStart
    const end = ta.selectionEnd
    editorYaml.value = editorYaml.value.substring(0, start) + '  ' + editorYaml.value.substring(end)
    nextTick(() => { ta.selectionStart = ta.selectionEnd = start + 2 })
  }
}

// YAML 模板列表
const crdTemplates = [
  { name: 'basic-crd', label: 'CRD 基础模板', yaml: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.mygroup.example.com
spec:
  group: mygroup.example.com
  names:
    kind: Example
    listKind: ExampleList
    plural: examples
    singular: example
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              x-kubernetes-preserve-unknown-fields: true
` },
]

const crTemplates = [
  { name: 'prometheus-rule', label: 'PrometheusRule 告警规则', yaml: `apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-alert-rule
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  groups:
    - name: custom.rules
      rules:
        - alert: HighMemoryUsage
          expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes < 0.1
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "节点内存使用率过高"
            description: "节点 {{ $labels.instance }} 内存可用率低于 10%"
` },
  { name: 'service-monitor', label: 'ServiceMonitor 服务监控', yaml: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  selector:
    matchLabels:
      app: my-service
  namespaceSelector:
    matchNames:
      - default
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
` },
  { name: 'generic-cr', label: '通用 CR 模板', yaml: '' },
]

const availableTemplates = computed(() => {
  return activeView.value === 'crd-list' ? crdTemplates : crTemplates
})

const applyTemplate = () => {
  if (!selectedTemplate.value) return
  const templates = activeView.value === 'crd-list' ? crdTemplates : crTemplates
  const tpl = templates.find(t => t.name === selectedTemplate.value)
  if (tpl && tpl.yaml) {
    editorYaml.value = tpl.yaml
  } else if (tpl && !tpl.yaml && activeCRD.value) {
    // 通用模板：根据当前 CRD 生成
    const plural = activeCRD.value.resource
    editorYaml.value = `apiVersion: ${activeCRD.value.group}/${activeCRD.value.version}
kind: ${activeCRD.value.kind}
metadata:
  name: my-${plural.slice(0, -1) || 'resource'}-01
  namespace: default
spec:
  # 在此定义资源规格
`
  }
  selectedTemplate.value = ''
}

// 格式化 YAML（简单缩进规整）
const formatYaml = () => {
  // 基本处理：移除尾部空白
  editorYaml.value = editorYaml.value.split('\n').map(l => l.trimEnd()).join('\n').trim() + '\n'
}

// 复制 YAML
const copyYaml = async () => {
  try {
    await navigator.clipboard.writeText(editorYaml.value)
    dryRunResult.value = { success: true, message: '已复制到剪贴板' }
    setTimeout(() => { if (dryRunResult.value?.message === '已复制到剪贴板') dryRunResult.value = null }, 2000)
  } catch {
    dryRunResult.value = { success: false, message: '复制失败，请手动选择复制' }
  }
}

// 清空编辑器
const clearEditor = () => {
  editorYaml.value = ''
  dryRunResult.value = null
}

// ========== 删除 ==========
const showDeleteConfirm = ref(false)
const deleteTarget = ref(null)
const deleteType = ref('crd') // 'crd' | 'cr'
const deleting = ref(false)

// ========== 计算属性 ==========
const uniqueGroups = computed(() => {
  const groups = new Set(crds.value.map(c => c.group))
  return [...groups].sort()
})

const namespacedCount = computed(() => crds.value.filter(c => c.scope === 'Namespaced').length)
const clusterCount = computed(() => crds.value.filter(c => c.scope === 'Cluster').length)

const filteredCRDs = computed(() => {
  return crds.value.filter(crd => {
    const matchSearch = !searchQuery.value ||
      crd.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      crd.kind?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      crd.group?.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchScope = !scopeFilter.value || crd.scope === scopeFilter.value
    const matchGroup = !groupFilter.value || crd.group === groupFilter.value
    return matchSearch && matchScope && matchGroup
  })
})

const totalPages = computed(() => Math.ceil(filteredCRDs.value.length / pageSize.value))

const paginatedCRDs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredCRDs.value.slice(start, start + pageSize.value)
})

const filteredCRInstances = computed(() => {
  if (!crSearchQuery.value) return crInstances.value
  return crInstances.value.filter(cr =>
    cr.name?.toLowerCase().includes(crSearchQuery.value.toLowerCase())
  )
})

// ========== CRD 操作 ==========
const fetchCRDs = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const params = {}
    if (searchQuery.value) params.keyword = searchQuery.value
    const res = await crdApi.listCRDs(params)
    if (res && res.code === 0) {
      const items = res.data?.list || res.data?.items || []
      crds.value = items.map(item => ({
        name: item.name,
        group: item.group,
        version: item.version,
        kind: item.kind,
        scope: item.scope,
        status: item.status || 'Unknown',
        description: item.description || '',
        createdAt: item.created_at || '-',
        versions: item.versions || [item.version],
        resource: item.resource || item.kind?.toLowerCase() + 's'
      }))
    } else {
      errorMsg.value = res?.msg || '获取 CRD 列表失败'
    }
  } catch (err) {
    errorMsg.value = err?.response?.data?.msg || err?.msg || err?.message || '网络错误'
  } finally {
    loading.value = false
  }
}

const viewCRDYaml = async (crd) => {
  yamlDrawerTitle.value = crd.name
  yamlContent.value = '# 加载中...'
  showYamlDrawer.value = true
  try {
    // 使用 CR YAML 接口获取 CRD 的完整 YAML（CRD 本身也是 apiextensions.k8s.io 的资源）
    const res = await crdApi.getCRYaml({
      group: 'apiextensions.k8s.io',
      version: 'v1',
      resource: 'customresourcedefinitions',
      name: crd.name
    })
    if (res && res.code === 0) {
      yamlContent.value = res.data?.yaml || '# 无内容'
    } else {
      yamlContent.value = `# 获取失败: ${res?.msg || '未知错误'}`
    }
  } catch (err) {
    yamlContent.value = `# 获取失败: ${err?.message || '网络错误'}`
  }
}

const openCreateCRD = () => {
  editorTitle.value = '创建 CustomResourceDefinition'
  editorMode.value = 'create'
  dryRunResult.value = null
  editorYaml.value = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.mygroup.example.com
spec:
  group: mygroup.example.com
  names:
    kind: Example
    listKind: ExampleList
    plural: examples
    singular: example
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              x-kubernetes-preserve-unknown-fields: true
            status:
              type: object
              x-kubernetes-preserve-unknown-fields: true
`
  showEditorModal.value = true
}

const confirmDeleteCRD = (crd) => {
  deleteTarget.value = crd
  deleteType.value = 'crd'
  showDeleteConfirm.value = true
}

// ========== CR 实例操作 ==========
const viewCRInstances = (crd) => {
  activeCRD.value = crd
  activeView.value = 'cr-list'
  crInstances.value = []
  crNamespaceFilter.value = ''
  crSearchQuery.value = ''
  fetchCRInstances()
}

const backToCRDList = () => {
  activeView.value = 'crd-list'
  activeCRD.value = null
}

const fetchCRInstances = async () => {
  if (!activeCRD.value) return
  crLoading.value = true
  try {
    const params = {
      group: activeCRD.value.group,
      version: activeCRD.value.version,
      resource: activeCRD.value.resource
    }
    if (crNamespaceFilter.value) params.namespace = crNamespaceFilter.value
    const res = await crdApi.listCRs(params)
    if (res && res.code === 0) {
      const items = res.data?.list || res.data?.items || []
      crInstances.value = items.map(item => {
        // 兼容两种格式：后端转换后的扁平结构 / 原始 K8s Unstructured 结构
        const meta = item.metadata || {}
        return {
          name: item.name || meta.name || '',
          namespace: item.namespace || meta.namespace || '',
          createdAt: item.created_at || meta.creationTimestamp || '-',
          labels: item.labels || meta.labels || {},
          uid: item.uid || meta.uid || '',
          resourceVersion: item.resource_version || meta.resourceVersion || ''
        }
      })
      // 收集命名空间
      const nsSet = new Set(crInstances.value.map(i => i.namespace).filter(Boolean))
      crNamespaces.value = [...nsSet].sort()
    }
  } catch (err) {
    console.error('fetchCRInstances error:', err)
  } finally {
    crLoading.value = false
  }
}

const openCreateCR = () => {
  if (!activeCRD.value) return
  editorTitle.value = `创建 ${activeCRD.value.kind} 实例`
  editorMode.value = 'create'
  editingCR.value = null
  dryRunResult.value = null
  const plural = activeCRD.value.resource
  editorYaml.value = `apiVersion: ${activeCRD.value.group}/${activeCRD.value.version}
kind: ${activeCRD.value.kind}
metadata:
  name: my-${plural.slice(0, -1) || 'resource'}-01
  namespace: default
spec:
  # 在此定义资源规格
  replicas: 1
`
  showEditorModal.value = true
}

const editCR = async (cr) => {
  if (!activeCRD.value) return
  editorTitle.value = `编辑 ${cr.name}`
  editorMode.value = 'edit'
  editingCR.value = cr
  dryRunResult.value = null
  editorYaml.value = '# 加载中...'
  showEditorModal.value = true
  try {
    const res = await crdApi.getCRYaml({
      group: activeCRD.value.group,
      version: activeCRD.value.version,
      resource: activeCRD.value.resource,
      namespace: cr.namespace,
      name: cr.name
    })
    if (res && res.code === 0) {
      editorYaml.value = res.data?.yaml || ''
    } else {
      editorYaml.value = `# 获取 YAML 失败: ${res?.msg}`
    }
  } catch (err) {
    editorYaml.value = `# 获取失败: ${err?.message}`
  }
}

const viewCRYaml = async (cr) => {
  yamlDrawerTitle.value = cr.name
  yamlContent.value = '# 加载中...'
  showYamlDrawer.value = true
  try {
    const res = await crdApi.getCRYaml({
      group: activeCRD.value.group,
      version: activeCRD.value.version,
      resource: activeCRD.value.resource,
      namespace: cr.namespace,
      name: cr.name
    })
    if (res && res.code === 0) {
      yamlContent.value = res.data?.yaml || ''
    } else {
      yamlContent.value = `# 获取失败: ${res?.msg}`
    }
  } catch (err) {
    yamlContent.value = `# 获取失败: ${err?.message}`
  }
}

const confirmDeleteCR = (cr) => {
  deleteTarget.value = cr
  deleteType.value = 'cr'
  showDeleteConfirm.value = true
}

// ========== 编辑器提交 ==========
const submitEditor = async () => {
  if (!editorYaml.value.trim()) return
  submitting.value = true
  try {
    if (activeView.value === 'crd-list' && editorMode.value === 'create') {
      // 创建 CRD（通过 CR API 作为 apiextensions.k8s.io 资源）
      const res = await crdApi.createCR({
        group: 'apiextensions.k8s.io',
        version: 'v1',
        resource: 'customresourcedefinitions',
        yaml: editorYaml.value
      })
      if (res && res.code === 0) {
        showEditorModal.value = false
        await fetchCRDs()
      } else {
        dryRunResult.value = { success: false, message: res?.msg || '创建失败' }
      }
    } else if (editorMode.value === 'create') {
      // 创建 CR
      const res = await crdApi.createCR({
        group: activeCRD.value.group,
        version: activeCRD.value.version,
        resource: activeCRD.value.resource,
        namespace: crNamespaceFilter.value || undefined,
        yaml: editorYaml.value
      })
      if (res && res.code === 0) {
        showEditorModal.value = false
        await fetchCRInstances()
      } else {
        dryRunResult.value = { success: false, message: res?.msg || '创建失败' }
      }
    } else {
      // 更新 CR
      const res = await crdApi.updateCR({
        group: activeCRD.value.group,
        version: activeCRD.value.version,
        resource: activeCRD.value.resource,
        namespace: editingCR.value?.namespace,
        name: editingCR.value?.name,
        yaml: editorYaml.value
      })
      if (res && res.code === 0) {
        showEditorModal.value = false
        await fetchCRInstances()
      } else {
        dryRunResult.value = { success: false, message: res?.msg || '更新失败' }
      }
    }
  } catch (err) {
    const details = Array.isArray(err?.details) ? err.details[0] : null
    dryRunResult.value = { success: false, message: details || err?.msg || err?.response?.data?.msg || err?.message || '操作失败' }
  } finally {
    submitting.value = false
  }
}

// ========== DryRun ==========
const doDryRun = async () => {
  if (!editorYaml.value.trim()) return
  dryRunning.value = true
  dryRunResult.value = null
  try {
    let res
    if (activeView.value === 'crd-list') {
      res = await crdApi.dryRun({
        group: 'apiextensions.k8s.io',
        version: 'v1',
        resource: 'customresourcedefinitions',
        yaml: editorYaml.value,
        is_update: false
      })
    } else {
      res = await crdApi.dryRun({
        group: activeCRD.value.group,
        version: activeCRD.value.version,
        resource: activeCRD.value.resource,
        namespace: editingCR.value?.namespace || crNamespaceFilter.value || undefined,
        name: editingCR.value?.name,
        yaml: editorYaml.value,
        is_update: editorMode.value === 'edit'
      })
    }
    if (res && res.code === 0) {
      // 后端返回 code:0 但 data.valid 可能为 false（DryRun 校验不通过）
      const dr = res.data
      if (dr?.valid) {
        dryRunResult.value = { success: true, message: dr.message || '✓ 校验通过，YAML 格式和内容均有效' }
      } else {
        const errMsg = (dr?.errors && dr.errors.length > 0) ? dr.errors.join('; ') : (dr?.message || '校验失败')
        dryRunResult.value = { success: false, message: errMsg }
      }
    } else {
      dryRunResult.value = { success: false, message: res?.msg || '校验失败' }
    }
  } catch (err) {
    const details = Array.isArray(err?.details) ? err.details[0] : null
    dryRunResult.value = { success: false, message: details || err?.msg || err?.response?.data?.msg || err?.message || '校验请求失败' }
  } finally {
    dryRunning.value = false
  }
}

// ========== 删除执行 ==========
const executeDelete = async () => {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    let res
    if (deleteType.value === 'crd') {
      res = await crdApi.deleteCRD({ name: deleteTarget.value.name })
    } else {
      res = await crdApi.deleteCR({
        group: activeCRD.value.group,
        version: activeCRD.value.version,
        resource: activeCRD.value.resource,
        namespace: deleteTarget.value.namespace,
        name: deleteTarget.value.name
      })
    }
    if (res && res.code === 0) {
      showDeleteConfirm.value = false
      if (deleteType.value === 'crd') {
        await fetchCRDs()
      } else {
        await fetchCRInstances()
      }
    } else {
      alert(res?.msg || '删除失败')
    }
  } catch (err) {
    alert(err?.response?.data?.msg || err?.msg || err?.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

// ========== 工具函数 ==========
const formatTime = (time) => {
  if (!time || time === '-') return '-'
  try {
    const d = new Date(time)
    return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  } catch {
    return time
  }
}

// ========== 监听搜索 ==========
let searchTimer = null
watch(searchQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
  }, 300)
})

watch([scopeFilter, groupFilter], () => {
  currentPage.value = 1
})

// ========== 初始化 ==========
onMounted(() => {
  fetchCRDs()
})
</script>

<style scoped>
/* ========== 全局变量 ========== */
.crd-management {
  --primary: #6366f1;
  --primary-hover: #4f46e5;
  --primary-bg: rgba(99,102,241,0.06);
  --primary-border: rgba(99,102,241,0.2);
  --danger: #ef4444;
  --success: #10b981;
  --warning: #f59e0b;
  --bg-page: #f8fafc;
  --bg-card: #ffffff;
  --bg-hover: #f1f5f9;
  --border: #e2e8f0;
  --border-light: #f1f5f9;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.02);
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,0.05), 0 2px 4px -2px rgba(0,0,0,0.03);
  --shadow-lg: 0 10px 15px -3px rgba(0,0,0,0.06), 0 4px 6px -4px rgba(0,0,0,0.04);
  --radius: 12px;
  --radius-sm: 8px;

  padding: 24px 28px;
  min-height: 100%;
  background: var(--bg-page);
}

/* ========== 页头 ========== */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon { font-size: 24px; }

.page-desc {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 4px;
}

.bc-link {
  color: var(--primary);
  cursor: pointer;
  font-weight: 500;
}
.bc-link:hover { text-decoration: underline; }
.bc-sep { color: var(--text-muted); }
.bc-current { color: var(--text-primary); font-weight: 600; }

/* ========== 统计卡片 ========== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 18px 20px;
  box-shadow: var(--shadow-sm);
  transition: all 0.2s;
}
.stat-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}
.stat-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1.2;
}
.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.stat-card.accent .stat-value { color: var(--primary); }
.stat-card.purple .stat-value { color: #8b5cf6; }

/* ========== 工具栏 ========== */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.search-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  font-size: 14px;
  opacity: 0.5;
}

.search-input {
  padding: 9px 32px 9px 36px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  width: 280px;
  background: var(--bg-card);
  transition: all 0.2s;
}
.search-input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(78,143,247,0.1);
}

.search-clear {
  position: absolute;
  right: 10px;
  cursor: pointer;
  color: var(--text-muted);
  font-size: 12px;
}

.filter-chips {
  display: flex;
  gap: 6px;
}

.chip {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid var(--border);
  background: var(--bg-card);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.chip:hover { border-color: var(--primary); color: var(--primary); }
.chip.active {
  background: linear-gradient(135deg, var(--primary), #8b5cf6);
  color: white;
  border-color: var(--primary);
}

.group-select {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
}

/* ========== 按钮 ========== */
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 18px;
  background: linear-gradient(135deg, var(--primary), #8b5cf6);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.btn-primary:hover { background: linear-gradient(135deg, var(--primary-hover), #7c3aed); transform: translateY(-1px); box-shadow: 0 4px 12px rgba(99,102,241,0.35); }
.btn-primary:disabled { opacity: 0.6; pointer-events: none; }

.btn-outline {
  padding: 7px 14px;
  background: transparent;
  color: var(--primary);
  border: 1px solid var(--primary);
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-outline:hover { background: rgba(78,143,247,0.06); }
.btn-outline.small { padding: 5px 10px; font-size: 12px; }

.btn-cancel {
  padding: 9px 18px;
  background: #f0f2f5;
  color: var(--text-secondary);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-cancel:hover { background: #e4e7ec; }

.btn-danger {
  padding: 9px 18px;
  background: var(--danger);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-danger:hover { background: #e04245; }
.btn-danger:disabled { opacity: 0.6; }

.btn-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}
.btn-icon:hover { border-color: var(--primary); background: rgba(78,143,247,0.04); }
.btn-icon:disabled { opacity: 0.5; pointer-events: none; }
.btn-icon-inner { font-size: 16px; font-weight: 700; }

.spinning { display: inline-block; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ========== 表格 ========== */
.table-wrapper {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.crd-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.crd-table thead tr {
  background: #f8f9fb;
  border-bottom: 1px solid var(--border);
}

.crd-table th {
  padding: 12px 16px;
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  text-align: left;
}

.crd-table td {
  padding: 14px 16px;
  font-size: 13px;
  color: var(--text-primary);
  border-bottom: 1px solid #f0f2f5;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 列宽分配 */
.col-name { width: 26%; }
.col-group { width: 16%; }
.col-kind { width: 12%; }
.col-version { width: 8%; }
.col-scope { width: 9%; }
.col-status { width: 9%; }
.col-time { width: 10%; font-size: 12px; color: var(--text-muted); white-space: nowrap; }
.col-actions { width: 10%; }

.crd-row {
  cursor: pointer;
  transition: background 0.15s;
}
.crd-row:hover { background: var(--bg-hover); }
.crd-row:last-child td { border-bottom: none; }

.name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.crd-icon { font-size: 16px; }
.name-primary {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
}

.group-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f0f5ff;
  color: #2f54eb;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kind-code {
  background: #f4f5f7;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: #d63384;
}

.version-tag {
  display: inline-block;
  padding: 2px 8px;
  background: #f6ffed;
  color: #389e0d;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  margin-right: 4px;
}

.scope-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.scope-badge.namespaced { background: #e6f7ff; color: #0958d9; }
.scope-badge.cluster { background: #f9f0ff; color: #722ed1; }

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: 6px;
}
.status-dot.established { background: var(--success); }
.status-dot.unknown { background: var(--text-muted); }
.status-text { font-size: 12px; color: var(--text-secondary); }

/* ========== 操作按钮 ========== */
.action-group {
  display: flex;
  gap: 6px;
}

.action-btn-text {
  padding: 4px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s;
  background: var(--bg-card);
  white-space: nowrap;
}
.action-btn-text.primary {
  color: var(--primary);
  border-color: rgba(78,143,247,0.3);
  background: rgba(78,143,247,0.04);
}
.action-btn-text.primary:hover {
  background: rgba(78,143,247,0.1);
  border-color: var(--primary);
}
.action-btn-text.default {
  color: var(--text-secondary);
}
.action-btn-text.default:hover {
  color: var(--primary);
  border-color: var(--primary);
  background: rgba(78,143,247,0.04);
}
.action-btn-text.danger {
  color: var(--danger);
  border-color: rgba(255,77,79,0.3);
  background: rgba(255,77,79,0.02);
}
.action-btn-text.danger:hover {
  background: rgba(255,77,79,0.08);
  border-color: var(--danger);
}

/* ========== 分页 ========== */
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
}
.page-info { font-size: 12px; color: var(--text-muted); }
.page-controls { display: flex; gap: 4px; }
.page-btn {
  min-width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-secondary);
}
.page-btn:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
.page-btn.active { background: var(--primary); color: white; border-color: var(--primary); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-size-select {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 12px;
  background: var(--bg-card);
}

/* ========== 状态提示 ========== */
.loading-state,
.empty-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.loader {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

.empty-icon,
.error-icon { font-size: 48px; margin-bottom: 12px; }
.empty-state h3,
.error-state h3 { margin: 0 0 8px; font-size: 16px; color: var(--text-primary); }
.empty-state p,
.error-state p { margin: 0 0 16px; font-size: 13px; color: var(--text-muted); }

/* ========== CR 区域 ========== */
.cr-header-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 16px 20px;
  margin-bottom: 16px;
  box-shadow: var(--shadow-sm);
}

.cr-meta {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}
.cr-meta-item {
  font-size: 13px;
  color: var(--text-secondary);
}
.cr-meta-item strong { color: var(--text-primary); margin-right: 4px; }

.cr-header-actions { display: flex; gap: 8px; align-items: center; }

.cr-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.ns-select {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--bg-card);
}
.cr-search-input {
  flex: 1;
  max-width: 300px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
}
.cr-search-input:focus { outline: none; border-color: var(--primary); }

.ns-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f0f5ff;
  color: #1d39c4;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.label-tags { display: flex; flex-wrap: wrap; gap: 4px; max-width: 300px; }
.label-tag {
  display: inline-block;
  padding: 1px 6px;
  background: #f4f5f7;
  border-radius: 3px;
  font-size: 10px;
  color: var(--text-secondary);
  font-family: monospace;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.no-labels { font-size: 12px; color: var(--text-muted); }

/* CR 表格覆盖 fixed 布局 */
.cr-table { table-layout: auto; }

/* ========== 抽屉 ========== */
.drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 2000;
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: 55%;
  max-width: 720px;
  min-width: 400px;
  background: #1e2030;
  display: flex;
  flex-direction: column;
  box-shadow: -8px 0 24px rgba(0,0,0,0.2);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.drawer-header h3 {
  margin: 0;
  font-size: 15px;
  color: #e0e0e0;
  font-weight: 600;
}
.drawer-close {
  background: none;
  border: none;
  color: #8b95b0;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}
.drawer-close:hover { background: rgba(255,255,255,0.08); color: #fff; }

.drawer-body {
  flex: 1;
  overflow: auto;
  padding: 0;
}
.drawer-body :deep(.yaml-highlight-wrapper) {
  border-radius: 0;
  border: none;
  box-shadow: none;
}

/* ========== 编辑器模态框 Pro ========== */
.editor-modal-pro {
  width: 80%;
  max-width: 1000px;
  min-width: 580px;
  height: 88vh;
  max-height: 88vh;
  background: #1b1d2e;
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.05);
  overflow: hidden;
}

/* 顶部栏 */
.editor-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 20px;
  background: #252840;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.editor-topbar-left { display: flex; align-items: center; gap: 12px; }
.editor-topbar-right { display: flex; align-items: center; gap: 8px; }
.editor-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
}
.editor-badge-icon { font-size: 12px; }
.editor-title-pro {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #e8eaf6;
}
.editor-topbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255,255,255,0.06);
  border-radius: 6px;
  color: #8b95b0;
  cursor: pointer;
  transition: all 0.2s;
}
.editor-topbar-btn:hover { background: rgba(255,255,255,0.12); color: #fff; }

/* 资源信息条 */
.editor-info-bar {
  padding: 10px 20px;
  background: #1e2035;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}
.editor-info-tags { display: flex; gap: 12px; flex-wrap: wrap; }
.info-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}
.info-tag-label { color: #6b7394; font-weight: 500; }
.info-tag-value {
  color: #a5b4fc;
  background: rgba(165, 180, 252, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
}

/* 工具栏 */
.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #22253a;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}
.editor-toolbar-left { display: flex; align-items: center; gap: 8px; }
.editor-toolbar-right { display: flex; align-items: center; gap: 8px; }

.template-select {
  padding: 6px 12px;
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 6px;
  background: #2a2d44;
  color: #c8cce0;
  font-size: 12px;
  cursor: pointer;
  outline: none;
}
.template-select:focus { border-color: #4e8ff7; }
.template-select option { background: #2a2d44; color: #c8cce0; }

.toolbar-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 6px;
  background: transparent;
  color: #8b95b0;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.toolbar-action-btn:hover { background: rgba(255,255,255,0.06); color: #e0e6f0; border-color: rgba(255,255,255,0.2); }
.toolbar-action-btn.danger { color: #fca5a5; border-color: rgba(239,68,68,0.3); }
.toolbar-action-btn.danger:hover { background: rgba(239,68,68,0.12); color: #fca5a5; border-color: rgba(239,68,68,0.5); }

.toolbar-dryrun-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: none;
  border-radius: 6px;
  background: linear-gradient(135deg, #10b981, #059669);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.toolbar-dryrun-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); }
.toolbar-dryrun-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.toolbar-dryrun-btn.running { background: #374151; }

.dryrun-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

/* DryRun 结果 Banner */
.dryrun-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 20px;
  font-size: 13px;
  animation: slideIn 0.3s ease;
}
.dryrun-banner.success { background: rgba(16, 185, 129, 0.12); color: #6ee7b7; border-bottom: 1px solid rgba(16, 185, 129, 0.2); }
.dryrun-banner.error { background: rgba(239, 68, 68, 0.12); color: #fca5a5; border-bottom: 1px solid rgba(239, 68, 68, 0.2); }
.dryrun-banner-icon { display: flex; align-items: center; }
.dryrun-banner-text { flex: 1; }
.dryrun-banner-close {
  background: none;
  border: none;
  color: inherit;
  opacity: 0.6;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 6px;
}
.dryrun-banner-close:hover { opacity: 1; }

/* 编辑区域 */
.editor-workspace {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

.editor-gutter {
  width: 48px;
  min-width: 48px;
  padding: 16px 0;
  background: #181a2a;
  border-right: 1px solid rgba(255,255,255,0.04);
  overflow: hidden;
  text-align: right;
  user-select: none;
}
.gutter-line {
  height: 22.1px;
  line-height: 22.1px;
  padding-right: 12px;
  font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
  font-size: 12px;
  color: #4a5073;
}

.yaml-editor-pro {
  flex: 1;
  width: 100%;
  padding: 16px 20px;
  border: none;
  resize: none;
  font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
  font-size: 13px;
  line-height: 22.1px;
  background: #1e2030;
  color: #cdd6f4;
  outline: none;
  tab-size: 2;
  white-space: pre;
  overflow: auto;
}
.yaml-editor-pro::placeholder { color: #4a5073; }
.yaml-editor-pro::-webkit-scrollbar { width: 8px; height: 8px; }
.yaml-editor-pro::-webkit-scrollbar-track { background: transparent; }
.yaml-editor-pro::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 4px; }
.yaml-editor-pro::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.2); }

/* 底部操作栏 */
.editor-footer-pro {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: #252840;
  border-top: 1px solid rgba(255,255,255,0.06);
}
.editor-footer-left { display: flex; align-items: center; gap: 16px; }
.editor-footer-right { display: flex; align-items: center; gap: 10px; }
.footer-stat {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: #6b7394;
}
.footer-stat svg { opacity: 0.7; }
.mode-tag {
  padding: 2px 8px;
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
  border-radius: 3px;
  font-weight: 600;
  font-size: 10px;
  letter-spacing: 0.5px;
}

.btn-editor-cancel {
  padding: 8px 18px;
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 8px;
  background: transparent;
  color: #8b95b0;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-editor-cancel:hover { background: rgba(255,255,255,0.06); color: #c8cce0; }

.btn-editor-submit {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 22px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-editor-submit:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99, 102, 241, 0.35); }
.btn-editor-submit:disabled { opacity: 0.5; cursor: not-allowed; transform: none; box-shadow: none; }

.submit-spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

/* Slide-down animation */
.slide-down-enter-active { transition: all 0.3s ease; }
.slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from { opacity: 0; transform: translateY(-8px); }
.slide-down-leave-to { opacity: 0; transform: translateY(-8px); }

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes slideIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

/* ========== 删除确认 ========== */
.delete-dialog {
  width: 400px;
  max-width: 90vw;
  background: var(--bg-card);
  border-radius: 16px;
  padding: 32px 28px 24px;
  text-align: center;
  box-shadow: var(--shadow-lg);
}

.delete-icon-area { margin-bottom: 12px; }
.delete-icon-circle { font-size: 40px; }
.delete-title { font-size: 17px; font-weight: 700; color: var(--text-primary); margin: 0 0 8px; }
.delete-desc { font-size: 13px; color: var(--text-secondary); margin: 0 0 8px; }
.delete-desc code { background: #f4f5f7; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
.delete-warning { font-size: 12px; color: var(--danger); margin: 0 0 20px; font-weight: 500; }
.delete-actions { display: flex; gap: 10px; justify-content: center; }

/* ========== 模态框通用 ========== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ========== 动画 ========== */
.drawer-enter-active { transition: all 0.3s ease; }
.drawer-leave-active { transition: all 0.25s ease; }
.drawer-enter-from .drawer-panel { transform: translateX(100%); }
.drawer-leave-to .drawer-panel { transform: translateX(100%); }
.drawer-enter-from { opacity: 0; }
.drawer-leave-to { opacity: 0; }

.modal-enter-active { transition: all 0.25s ease; }
.modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from { opacity: 0; }
.modal-enter-from .editor-modal-pro,
.modal-enter-from .delete-dialog { transform: scale(0.95); opacity: 0; }
.modal-leave-to { opacity: 0; }
.modal-leave-to .editor-modal-pro,
.modal-leave-to .delete-dialog { transform: scale(0.95); opacity: 0; }

/* ========== 响应式 ========== */
@media (max-width: 1200px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .drawer-panel { width: 70%; }
  .editor-modal-pro { width: 92%; }
}

@media (max-width: 768px) {
  .crd-management { padding: 16px; }
  .stats-row { grid-template-columns: 1fr 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left { flex-direction: column; }
  .search-input { width: 100%; }
  .editor-modal-pro { width: 98%; height: 95vh; max-height: 95vh; }
  .editor-toolbar { flex-direction: column; gap: 8px; }
  .editor-toolbar-left { flex-wrap: wrap; }
}
</style>

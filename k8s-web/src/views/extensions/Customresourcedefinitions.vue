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
          <div class="editor-modal" @click.stop>
            <div class="editor-header">
              <h3>{{ editorTitle }}</h3>
              <div class="editor-header-actions">
                <button class="btn-outline small" @click="doDryRun" :disabled="dryRunning">
                  {{ dryRunning ? '校验中...' : '🧪 DryRun 校验' }}
                </button>
                <button class="editor-close" @click="showEditorModal = false">✕</button>
              </div>
            </div>
            <!-- DryRun 结果 -->
            <div v-if="dryRunResult" class="dryrun-result" :class="dryRunResult.success ? 'success' : 'error'">
              <span class="dryrun-icon">{{ dryRunResult.success ? '✅' : '❌' }}</span>
              <span>{{ dryRunResult.message }}</span>
            </div>
            <div class="editor-body">
              <textarea
                v-model="editorYaml"
                class="yaml-editor"
                spellcheck="false"
                placeholder="# 在此输入 YAML 内容..."
              ></textarea>
            </div>
            <div class="editor-footer">
              <button class="btn-cancel" @click="showEditorModal = false">取消</button>
              <button class="btn-primary" @click="submitEditor" :disabled="submitting">
                {{ submitting ? '提交中...' : editorMode === 'create' ? '创建' : '更新' }}
              </button>
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
import { ref, computed, onMounted, watch } from 'vue'
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
    dryRunResult.value = { success: false, message: err?.response?.data?.msg || err?.msg || err?.message || '操作失败' }
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
      dryRunResult.value = { success: true, message: '✓ 校验通过，YAML 格式和内容均有效' }
    } else {
      dryRunResult.value = { success: false, message: res?.msg || '校验失败' }
    }
  } catch (err) {
    dryRunResult.value = { success: false, message: err?.response?.data?.msg || err?.msg || err?.message || '校验请求失败' }
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
  --primary: #4e8ff7;
  --primary-hover: #3b7be0;
  --danger: #ff4d4f;
  --success: #52c41a;
  --warning: #faad14;
  --bg-page: #f7f8fc;
  --bg-card: #ffffff;
  --bg-hover: #f5f7fa;
  --border: #e8ecf0;
  --text-primary: #1a2138;
  --text-secondary: #5e6c84;
  --text-muted: #97a0af;
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.04);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.06);
  --shadow-lg: 0 8px 24px rgba(0,0,0,0.08);
  --radius: 10px;
  --radius-sm: 6px;

  padding: 28px 32px;
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
.stat-card.purple .stat-value { color: #7c3aed; }

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
  background: var(--primary);
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
  background: var(--primary);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-primary:hover { background: var(--primary-hover); transform: translateY(-1px); box-shadow: 0 4px 12px rgba(78,143,247,0.3); }
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

/* ========== 编辑器模态框 ========== */
.editor-modal {
  width: 70%;
  max-width: 900px;
  min-width: 500px;
  max-height: 90vh;
  background: var(--bg-card);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
}
.editor-header h3 { margin: 0; font-size: 16px; font-weight: 700; color: var(--text-primary); }
.editor-header-actions { display: flex; gap: 10px; align-items: center; }
.editor-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px 8px;
  border-radius: 4px;
}
.editor-close:hover { background: #f0f2f5; color: var(--text-primary); }

.dryrun-result {
  padding: 10px 24px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.dryrun-result.success { background: #f6ffed; color: #389e0d; border-bottom: 1px solid #b7eb8f; }
.dryrun-result.error { background: #fff2f0; color: #cf1322; border-bottom: 1px solid #ffccc7; }
.dryrun-icon { font-size: 15px; }

.editor-body {
  flex: 1;
  overflow: auto;
  padding: 0;
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  height: 100%;
  padding: 16px 20px;
  border: none;
  resize: none;
  font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
  font-size: 13px;
  line-height: 1.7;
  background: #1e1e2e;
  color: #cdd6f4;
  outline: none;
  tab-size: 2;
}
.yaml-editor::placeholder { color: #6c7086; }

.editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 24px;
  border-top: 1px solid var(--border);
  background: #f8f9fb;
}

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
.modal-enter-from .editor-modal,
.modal-enter-from .delete-dialog { transform: scale(0.95); opacity: 0; }
.modal-leave-to { opacity: 0; }
.modal-leave-to .editor-modal,
.modal-leave-to .delete-dialog { transform: scale(0.95); opacity: 0; }

/* ========== 响应式 ========== */
@media (max-width: 1200px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .drawer-panel { width: 70%; }
  .editor-modal { width: 85%; }
}

@media (max-width: 768px) {
  .crd-management { padding: 16px; }
  .stats-row { grid-template-columns: 1fr 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left { flex-direction: column; }
  .search-input { width: 100%; }
}
</style>

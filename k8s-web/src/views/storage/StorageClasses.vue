<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>StorageClass 管理</h1>
      <p>Kubernetes 集群中的存储类列表（集群级资源）</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索 StorageClass..." @input="onSearchInput" />
      </div>

      <div class="action-buttons">
        <button v-if="!batchMode" class="btn btn-batch" @click="enterBatchMode" title="进入批量操作模式">
          ☑️ 批量操作
        </button>
        <button v-if="batchMode" class="btn btn-secondary" @click="exitBatchMode">
          ✖️ 退出批量
        </button>
        
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        <button class="btn btn-primary" @click="openCreateModal">创建 StorageClass</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedItems.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedItems.length }} 个 StorageClass</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="batchDelete" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 表格视图 -->
    <div class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" title="全选/取消全选" />
            </th>
            <th style="width: 80px;">默认</th>
            <th style="min-width: 180px;">名称</th>
            <th style="min-width: 200px;">Provisioner</th>
            <th style="width: 120px;">回收策略</th>
            <th style="width: 180px;">绑定模式</th>
            <th style="width: 100px;">允许扩容</th>
            <th style="width: 170px;">创建时间</th>
            <th style="width: 100px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sc in paginatedList" :key="sc.name" :class="{ 'row-selected': isItemSelected(sc) }">
            <td v-if="batchMode">
              <input type="checkbox" :checked="isItemSelected(sc)" @change="toggleItemSelection(sc)" />
            </td>
            <td>
              <span v-if="sc.is_default" class="default-badge">⭐ 默认</span>
              <span v-else>-</span>
            </td>
            <td>
              <div class="sc-name">
                <span class="icon">💾</span>
                <span>{{ sc.name }}</span>
              </div>
            </td>
            <td>
              <span class="provisioner-text">{{ sc.provisioner }}</span>
            </td>
            <td>
              <span class="policy-badge" :class="sc.reclaim_policy.toLowerCase()">
                {{ sc.reclaim_policy }}
              </span>
            </td>
            <td>
              <span class="binding-badge">{{ sc.volume_binding_mode }}</span>
            </td>
            <td>
              <span :class="sc.allow_expansion ? 'badge-yes' : 'badge-no'">
                {{ sc.allow_expansion ? '✓ 是' : '✗ 否' }}
              </span>
            </td>
            <td>{{ sc.created_at }}</td>
            <td>
              <div class="action-icons">
                <button class="action-btn" @click="viewDetail(sc)" title="查看详情">
                  📋 详情
                </button>
                <div class="more-btn">
                  <button class="icon-btn" @click="toggleMoreOptions(sc, $event)">
                    ⋮ 更多
                  </button>
                  <div v-if="showMoreOptions && selectedItem === sc" class="more-menu" :style="menuStyle">
                    <button class="menu-item" @click="openYamlPreview(sc)">
                      <span class="menu-icon">📝</span>
                      <span>查看/编辑 YAML</span>
                    </button>
                    <button class="menu-item" @click="downloadYaml(sc)">
                      <span class="menu-icon">⬇️</span>
                      <span>下载 YAML</span>
                    </button>
                    <div class="menu-divider"></div>
                    <button class="menu-item danger" @click="deleteItem(sc)">
                      <span class="menu-icon">🗑️</span>
                      <span>删除</span>
                    </button>
                  </div>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页（增强版 - 参考 Deployment.vue） -->
    <div class="pagination-wrapper">
      <div class="pagination-left">
        <span class="pagination-summary">共 <strong>{{ totalItems }}</strong> 条</span>
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
        <select v-model.number="pageSize" @change="onPageSizeChange" class="page-size-select">
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

    <!-- 创建模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>创建 StorageClass</h3>
          <div class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: createMode === 'form' }" @click="createMode = 'form'">📋 表单</button>
            <button class="view-toggle-btn" :class="{ active: createMode === 'yaml' }" @click="createMode = 'yaml'">📄 YAML</button>
          </div>
          <button class="close-btn" @click="closeCreateModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'" class="form-mode">
            <div class="form-section">
              <h4>📝 基本信息</h4>
              <div class="form-group">
                <label>名称 <span class="required">*</span></label>
                <input v-model="formData.name" type="text" placeholder="例如: fast-storage" />
              </div>
              <div class="form-group">
                <label>Provisioner <span class="required">*</span></label>
                <input v-model="formData.provisioner" type="text" placeholder="例如: kubernetes.io/no-provisioner" />
              </div>
            </div>

            <div class="form-section">
              <h4>⚙️ 配置选项</h4>
              <div class="form-group">
                <label>回收策略</label>
                <select v-model="formData.reclaimPolicy">
                  <option value="Delete">Delete（删除）</option>
                  <option value="Retain">Retain（保留）</option>
                </select>
              </div>
              <div class="form-group">
                <label>绑定模式</label>
                <select v-model="formData.volumeBindingMode">
                  <option value="Immediate">Immediate（立即绑定）</option>
                  <option value="WaitForFirstConsumer">WaitForFirstConsumer（延迟绑定）</option>
                </select>
              </div>
              <div class="form-group">
                <label class="checkbox-label">
                  <input v-model="formData.allowVolumeExpansion" type="checkbox" />
                  允许卷扩容
                </label>
              </div>
            </div>

            <div class="form-section">
              <h4>🔧 参数 (Parameters)</h4>
              <div v-for="(param, index) in formData.parameters" :key="index" class="param-row">
                <input v-model="param.key" type="text" placeholder="参数名" class="param-key-input" />
                <input v-model="param.value" type="text" placeholder="参数值" class="param-value-input" />
                <button class="btn-icon-danger" @click="removeParameter(index)" title="删除">🗑️</button>
              </div>
              <button class="btn btn-link" @click="addParameter">+ 添加参数</button>
            </div>

            <div class="form-section">
              <h4>📦 挂载选项 (Mount Options)</h4>
              <div v-for="(opt, index) in formData.mountOptions" :key="index" class="mount-option-row">
                <input v-model="formData.mountOptions[index]" type="text" placeholder="挂载选项" />
                <button class="btn-icon-danger" @click="removeMountOption(index)" title="删除">🗑️</button>
              </div>
              <button class="btn btn-link" @click="addMountOption">+ 添加挂载选项</button>
            </div>

            <div v-if="createYamlError" class="error-message">{{ createYamlError }}</div>
          </div>

          <!-- YAML 模式 -->
          <div v-else class="yaml-mode">
            <textarea v-model="createYamlContent" placeholder="请输入 StorageClass YAML 配置..." class="yaml-editor"></textarea>
            <div v-if="createYamlError" class="error-message">{{ createYamlError }}</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeCreateModal">取消</button>
          <button v-if="createMode === 'yaml'" class="btn btn-link" @click="loadYamlTemplate">加载 YAML 模板</button>
          <button class="btn btn-primary" @click="createMode === 'form' ? createFromForm() : createFromYaml()" :disabled="creating">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>StorageClass YAML - {{ currentItem.name }}</h3>
          <div class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: yamlMode === 'view' }" @click="yamlMode = 'view'">
              👁️ 查看
            </button>
            <button class="view-toggle-btn" :class="{ active: yamlMode === 'edit' }" @click="yamlMode = 'edit'">
              ✏️ 编辑
            </button>
          </div>
          <button class="close-btn" @click="closeYamlModal">×</button>
        </div>
        <div class="modal-body">
          <textarea v-if="yamlMode === 'view'" :value="currentYaml" class="yaml-editor" readonly></textarea>
          <textarea v-else v-model="editedYaml" class="yaml-editor" placeholder="编辑 YAML..."></textarea>
          <div v-if="yamlError" class="error-message">{{ yamlError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">关闭</button>
          <button v-if="yamlMode === 'view'" class="btn btn-link" @click="downloadCurrentYaml">下载 YAML</button>
          <button v-if="yamlMode === 'edit'" class="btn btn-primary" @click="applyYamlChanges" :disabled="yamlSaving">
            {{ yamlSaving ? '保存中...' : '应用更改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="closeDetailModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>StorageClass 详情 - {{ currentItem.name }}</h3>
          <button class="close-btn" @click="closeDetailModal">×</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h4>基本信息</h4>
            <div class="detail-grid">
              <div class="detail-item">
                <span class="label">名称:</span>
                <span class="value">{{ currentItem.name }}</span>
              </div>
              <div class="detail-item">
                <span class="label">Provisioner:</span>
                <span class="value">{{ currentItem.provisioner }}</span>
              </div>
              <div class="detail-item">
                <span class="label">回收策略:</span>
                <span class="value">{{ currentItem.reclaim_policy }}</span>
              </div>
              <div class="detail-item">
                <span class="label">绑定模式:</span>
                <span class="value">{{ currentItem.volume_binding_mode }}</span>
              </div>
              <div class="detail-item">
                <span class="label">允许扩容:</span>
                <span class="value">{{ currentItem.allow_expansion ? '是' : '否' }}</span>
              </div>
              <div class="detail-item">
                <span class="label">是否默认:</span>
                <span class="value">{{ currentItem.is_default ? '是' : '否' }}</span>
              </div>
              <div class="detail-item">
                <span class="label">创建时间:</span>
                <span class="value">{{ currentItem.created_at }}</span>
              </div>
            </div>
          </div>
          <div v-if="currentItem.parameters && Object.keys(currentItem.parameters).length > 0" class="detail-section">
            <h4>参数 (Parameters)</h4>
            <div class="param-list">
              <div v-for="(value, key) in currentItem.parameters" :key="key" class="param-item">
                <span class="param-key">{{ key }}:</span>
                <span class="param-value">{{ value }}</span>
              </div>
            </div>
          </div>
          <div v-if="currentItem.mount_options && currentItem.mount_options.length > 0" class="detail-section">
            <h4>挂载选项 (Mount Options)</h4>
            <div class="options-list">
              <span v-for="opt in currentItem.mount_options" :key="opt" class="option-badge">{{ opt }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeDetailModal">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import storageclassesApi from '@/api/cluster/storage/storageclasses'

// =============== 状态管理 ===============
const storageClasses = ref([])
const loading = ref(false)
const errorMsg = ref('')

// 搜索和过滤
const searchQuery = ref('')
const debouncedSearchQuery = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const jumpPage = ref(1)

// 批量操作
const batchMode = ref(false)
const selectedItems = ref([])

// 自动刷新
const autoRefresh = ref(false)
const refreshTimer = ref(null)

// 模态框
const showCreateModal = ref(false)
const showYamlModal = ref(false)
const showDetailModal = ref(false)
const creating = ref(false)
const yamlSaving = ref(false)

// 创建相关
const createMode = ref('form')
const createYamlContent = ref('')
const createYamlError = ref('')

// 表单数据
const formData = ref({
  name: '',
  provisioner: '',
  reclaimPolicy: 'Delete',
  volumeBindingMode: 'Immediate',
  allowVolumeExpansion: false,
  parameters: [],
  mountOptions: []
})

// YAML 查看/编辑
const currentItem = ref({})
const currentYaml = ref('')
const editedYaml = ref('')
const yamlMode = ref('view')
const yamlError = ref('')

// 更多菜单
const showMoreOptions = ref(false)
const selectedItem = ref(null)
const menuStyle = ref({})

// =============== 计算属性 ===============
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value) || 1)

// 可见页码按钮（智能显示）
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

const filteredList = computed(() => {
  let result = storageClasses.value

  // 搜索过滤
  if (debouncedSearchQuery.value) {
    const query = debouncedSearchQuery.value.toLowerCase()
    result = result.filter(sc => 
      sc.name.toLowerCase().includes(query) ||
      sc.provisioner.toLowerCase().includes(query)
    )
  }

  return result
})

const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredList.value.slice(start, end)
})

const isAllSelected = computed(() => {
  return paginatedList.value.length > 0 && 
         paginatedList.value.every(sc => selectedItems.value.some(s => s.name === sc.name))
})

// =============== 搜索防抖 ===============
let searchTimeout = null
const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    debouncedSearchQuery.value = searchQuery.value
    currentPage.value = 1
  }, 300)
}

// =============== API 调用 ===============
const fetchList = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await storageclassesApi.list({
      page: 1, // 前端分页
      limit: 1000, // 获取所有数据
      name: ''
    })
    if (res.code === 0) {
      storageClasses.value = res.data.list || []
      totalItems.value = storageClasses.value.length
    } else {
      errorMsg.value = res.message || '获取 StorageClass 列表失败'
    }
  } catch (error) {
    console.error('获取 StorageClass 列表失败:', error)
    errorMsg.value = error.message || '网络错误'
  } finally {
    loading.value = false
  }
}

const refreshList = () => {
  fetchList()
}

// =============== 批量操作 ===============
const enterBatchMode = () => {
  batchMode.value = true
  selectedItems.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedItems.value = []
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedItems.value = selectedItems.value.filter(
      s => !paginatedList.value.some(sc => sc.name === s.name)
    )
  } else {
    paginatedList.value.forEach(sc => {
      if (!selectedItems.value.some(s => s.name === sc.name)) {
        selectedItems.value.push(sc)
      }
    })
  }
}

const isItemSelected = (sc) => {
  return selectedItems.value.some(s => s.name === sc.name)
}

const toggleItemSelection = (sc) => {
  const index = selectedItems.value.findIndex(s => s.name === sc.name)
  if (index > -1) {
    selectedItems.value.splice(index, 1)
  } else {
    selectedItems.value.push(sc)
  }
}

const clearSelection = () => {
  selectedItems.value = []
}

const batchDelete = () => {
  if (selectedItems.value.length === 0) return
  
  const names = selectedItems.value.map(sc => sc.name).join(', ')
  if (!confirm(`确定要删除这 ${selectedItems.value.length} 个 StorageClass 吗？\n${names}`)) {
    return
  }

  const promises = selectedItems.value.map(sc => 
    storageclassesApi.delete({ name: sc.name })
  )

  Promise.all(promises)
    .then(results => {
      const successCount = results.filter(r => r.code === 0).length
      alert(`批量删除完成！成功: ${successCount}, 失败: ${results.length - successCount}`)
      clearSelection()
      exitBatchMode()
      refreshList()
    })
    .catch(error => {
      console.error('批量删除失败:', error)
      alert('批量删除失败: ' + error.message)
    })
}

// =============== 单个操作 ===============
const deleteItem = async (sc) => {
  if (!confirm(`确定要删除 StorageClass "${sc.name}" 吗？`)) return

  try {
    const res = await storageclassesApi.delete({ name: sc.name })
    if (res.code === 0) {
      alert('删除成功')
      refreshList()
    } else {
      alert('删除失败: ' + (res.message || '未知错误'))
    }
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败: ' + error.message)
  }
}

const viewDetail = (sc) => {
  currentItem.value = sc
  showDetailModal.value = true
}

const closeDetailModal = () => {
  showDetailModal.value = false
  currentItem.value = {}
}

// =============== 创建操作 ===============
const openCreateModal = () => {
  // 重置表单
  formData.value = {
    name: '',
    provisioner: '',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: false,
    parameters: [],
    mountOptions: []
  }
  createMode.value = 'form'
  createYamlContent.value = ''
  createYamlError.value = ''
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
  createYamlContent.value = ''
  createYamlError.value = ''
}

// 表单创建
const createFromForm = async () => {
  // 验证
  if (!formData.value.name.trim()) {
    createYamlError.value = '请输入名称'
    return
  }
  if (!formData.value.provisioner.trim()) {
    createYamlError.value = '请输入 Provisioner'
    return
  }

  creating.value = true
  createYamlError.value = ''

  try {
    // 构建请求数据
    const params = {}
    formData.value.parameters.forEach(p => {
      if (p.key && p.value) {
        params[p.key] = p.value
      }
    })

    const requestData = {
      name: formData.value.name,
      provisioner: formData.value.provisioner,
      reclaim_policy: formData.value.reclaimPolicy,
      volume_binding_mode: formData.value.volumeBindingMode,
      allow_volume_expansion: formData.value.allowVolumeExpansion,
      parameters: params,
      mount_options: formData.value.mountOptions.filter(opt => opt.trim())
    }

    const res = await storageclassesApi.create(requestData)
    if (res.code === 0) {
      alert('StorageClass 创建成功！')
      closeCreateModal()
      refreshList()
    } else {
      createYamlError.value = res.kube_message_error || res.message || '创建失败'
    }
  } catch (error) {
    console.error('创建 StorageClass 失败:', error)
    createYamlError.value = error.message || '创建失败'
  } finally {
    creating.value = false
  }
}

// 添加/删除参数
const addParameter = () => {
  formData.value.parameters.push({ key: '', value: '' })
}

const removeParameter = (index) => {
  formData.value.parameters.splice(index, 1)
}

// 添加/删除挂载选项
const addMountOption = () => {
  formData.value.mountOptions.push('')
}

const removeMountOption = (index) => {
  formData.value.mountOptions.splice(index, 1)
}

const loadYamlTemplate = () => {
  createYamlContent.value = `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: my-storageclass
provisioner: kubernetes.io/no-provisioner
reclaimPolicy: Retain
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
parameters:
  type: pd-standard`
}

const createFromYaml = async () => {
  if (!createYamlContent.value.trim()) {
    createYamlError.value = '请输入 YAML 内容'
    return
  }

  creating.value = true
  createYamlError.value = ''

  try {
    const res = await storageclassesApi.createFromYaml({ yaml: createYamlContent.value })
    if (res.code === 0) {
      alert('StorageClass 创建成功！')
      closeCreateModal()
      refreshList()
    } else {
      createYamlError.value = res.kube_message_error || res.message || '创建失败'
    }
  } catch (error) {
    console.error('创建 StorageClass 失败:', error)
    createYamlError.value = error.message || '创建失败'
  } finally {
    creating.value = false
  }
}

// =============== YAML 查看/编辑 ===============
const openYamlPreview = async (sc) => {
  currentItem.value = sc
  yamlMode.value = 'view'
  yamlError.value = ''
  showYamlModal.value = true

  try {
    const res = await storageclassesApi.yaml({ name: sc.name })
    if (res.code === 0) {
      currentYaml.value = res.data.yaml
      editedYaml.value = res.data.yaml
    } else {
      yamlError.value = '获取 YAML 失败: ' + (res.message || '未知错误')
    }
  } catch (error) {
    console.error('获取 YAML 失败:', error)
    yamlError.value = '获取 YAML 失败: ' + error.message
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  currentItem.value = {}
  currentYaml.value = ''
  editedYaml.value = ''
  yamlError.value = ''
}

const applyYamlChanges = async () => {
  if (!editedYaml.value.trim()) {
    yamlError.value = 'YAML 内容不能为空'
    return
  }

  yamlSaving.value = true
  yamlError.value = ''

  try {
    const res = await storageclassesApi.applyYaml({ yaml: editedYaml.value })
    if (res.code === 0) {
      alert('YAML 应用成功！')
      closeYamlModal()
      refreshList()
    } else {
      yamlError.value = res.kube_message_error || res.message || '应用失败'
    }
  } catch (error) {
    console.error('应用 YAML 失败:', error)
    yamlError.value = error.message || '应用失败'
  } finally {
    yamlSaving.value = false
  }
}

const downloadYaml = async (sc) => {
  try {
    await storageclassesApi.downloadYaml(sc.name)
  } catch (error) {
    console.error('下载 YAML 失败:', error)
    alert('下载失败: ' + error.message)
  }
}

const downloadCurrentYaml = () => {
  if (!currentYaml.value) return
  const blob = new Blob([currentYaml.value], { type: 'text/yaml' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `storageclass-${currentItem.value.name}.yaml`
  link.click()
  URL.revokeObjectURL(link.href)
}

// =============== 更多菜单 ===============
const toggleMoreOptions = (sc, event) => {
  if (showMoreOptions.value && selectedItem.value === sc) {
    showMoreOptions.value = false
    selectedItem.value = null
  } else {
    selectedItem.value = sc
    showMoreOptions.value = true
    
    const button = event.target.closest('.more-btn')
    const rect = button.getBoundingClientRect()
    menuStyle.value = {
      top: `${rect.bottom + 5}px`,
      right: `${window.innerWidth - rect.right}px`
    }
  }
}

// =============== 分页操作 ===============
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

const onPageSizeChange = () => {
  currentPage.value = 1
}

// =============== 自动刷新 ===============
const startAutoRefresh = () => {
  if (refreshTimer.value) return
  refreshTimer.value = setInterval(() => {
    if (autoRefresh.value) {
      fetchList()
    }
  }, 90000) // 90秒
}

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
}

watch(autoRefresh, (val) => {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

// =============== 生命周期 ===============
onMounted(() => {
  fetchList()
  startAutoRefresh()
  
  // 全局点击关闭更多菜单
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.more-btn')) {
      showMoreOptions.value = false
      selectedItem.value = null
    }
  })
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
/* 基础布局 */
.resource-view {
  padding: 20px;
  max-width: 1800px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 30px;
}

.view-header h1 {
  font-size: 28px;
  color: #1f2937;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #6b7280;
  margin: 0;
}

/* 操作栏 */
.action-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 250px;
}

.search-box input {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 10px;
  align-items: center;
}

/* 按钮样式 */
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
}

.btn-batch {
  background: #8b5cf6;
  color: white;
}

.btn-batch:hover {
  background: #7c3aed;
}

.btn-link {
  background: transparent;
  color: #3b82f6;
  border: 1px solid #3b82f6;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 自动刷新 */
.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  cursor: pointer;
}

.auto-refresh-toggle input {
  cursor: pointer;
}

.refresh-indicator {
  color: #10b981;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 错误提示 */
.error-box {
  background: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

/* 批量操作浮动栏 */
.batch-action-bar {
  position: fixed;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  padding: 16px 24px;
  display: flex;
  gap: 24px;
  align-items: center;
  z-index: 1000;
}

.batch-info {
  display: flex;
  gap: 12px;
  align-items: center;
}

.batch-count {
  font-weight: 600;
  color: #3b82f6;
}

.batch-clear {
  background: #f3f4f6;
  border: none;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.batch-actions {
  display: flex;
  gap: 8px;
}

.batch-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  background: #3b82f6;
  color: white;
}

.batch-btn.danger {
  background: #dc2626;
}

.batch-btn:hover {
  opacity: 0.9;
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
  min-width: 1100px;
}

.resource-table thead {
  background: #f9fafb;
  border-bottom: 2px solid #e5e7eb;
}

.resource-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  font-size: 13px;
}

.resource-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
  font-size: 14px;
}

.resource-table tbody tr:hover {
  background: #f9fafb;
}

.row-selected {
  background: #eff6ff !important;
}

/* 表格内容样式 */
.sc-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.default-badge {
  color: #f59e0b;
  font-weight: 500;
}

.provisioner-text {
  font-family: monospace;
  font-size: 13px;
  color: #6b7280;
}

.policy-badge,
.binding-badge {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  background: #f3f4f6;
  color: #374151;
}

.policy-badge.delete {
  background: #fee2e2;
  color: #dc2626;
}

.policy-badge.retain {
  background: #dcfce7;
  color: #16a34a;
}

.badge-yes {
  color: #16a34a;
  font-weight: 500;
}

.badge-no {
  color: #9ca3af;
}

/* 操作按钮 */
.action-icons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.more-btn {
  position: relative;
}

.icon-btn {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 16px;
}

.icon-btn:hover {
  background: #f3f4f6;
}

/* 更多菜单 */
.more-menu {
  position: fixed;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  z-index: 1000;
  overflow: hidden;
}

.menu-item {
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: white;
  text-align: left;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: background 0.2s;
}

.menu-item:hover {
  background: #f3f4f6;
}

.menu-item.danger {
  color: #dc2626;
}

.menu-icon {
  font-size: 16px;
}

.menu-divider {
  height: 1px;
  background: #e5e7eb;
  margin: 4px 0;
}

/* 分页样式（现代化设计） */
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
  gap: 4px;
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
  background: #f9fafb;
  border-color: #3b82f6;
  color: #3b82f6;
}

.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  color: #9ca3af;
}

.pagination-btn.page-number {
  font-weight: 500;
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
  user-select: none;
}

.pagination-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-size-select {
  padding: 6px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: white;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.page-size-select:hover {
  border-color: #3b82f6;
}

.page-size-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.pagination-goto {
  font-size: 13px;
  color: #6b7280;
}

.page-jump-input {
  width: 50px;
  padding: 6px 8px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  text-align: center;
  font-size: 13px;
  transition: all 0.2s;
}

.page-jump-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 移除或隐藏旧的分页样式 */
.pagination-info,
.pagination-controls {
  display: none;
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
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal-large {
  max-width: 900px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1f2937;
}

.view-toggle-buttons {
  display: flex;
  gap: 8px;
}

.view-toggle-btn {
  padding: 6px 14px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #9ca3af;
  line-height: 1;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

/* 表单样式 */
.form-mode {
  max-height: 600px;
  overflow-y: auto;
}

.form-section {
  margin-bottom: 24px;
  padding: 20px;
  background: #f9fafb;
  border-radius: 8px;
}

.form-section h4 {
  font-size: 16px;
  margin: 0 0 16px 0;
  color: #1f2937;
  font-weight: 600;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  color: #374151;
  font-weight: 500;
}

.form-group input[type="text"],
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-group input[type="text"]:focus,
.form-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.required {
  color: #ef4444;
  margin-left: 2px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal !important;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

/* 参数和挂载选项样式 */
.param-row,
.mount-option-row {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
  align-items: center;
}

.param-key-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.param-value-input {
  flex: 2;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.mount-option-row input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.btn-icon-danger {
  background: #fee2e2;
  color: #dc2626;
  border: none;
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}

.btn-icon-danger:hover {
  background: #fecaca;
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  padding: 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  resize: vertical;
}

.error-message {
  margin-top: 12px;
  padding: 12px;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 6px;
  font-size: 13px;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 详情模态框 */
.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  font-size: 16px;
  margin: 0 0 12px 0;
  color: #1f2937;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 12px;
}

.detail-item {
  display: flex;
  gap: 8px;
}

.detail-item .label {
  font-weight: 500;
  color: #6b7280;
  min-width: 120px;
}

.detail-item .value {
  color: #1f2937;
}

.param-list,
.options-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.param-item {
  padding: 8px 12px;
  background: #f3f4f6;
  border-radius: 6px;
  font-size: 13px;
}

.param-key {
  font-weight: 500;
  color: #6b7280;
  margin-right: 4px;
}

.param-value {
  color: #1f2937;
  font-family: monospace;
}

.option-badge {
  padding: 6px 12px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  font-size: 13px;
}
</style>

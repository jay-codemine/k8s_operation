<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>PVC 管理</h1>
      <p>Kubernetes 持久卷声明列表</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索 PVC 名称..." @input="onSearchInput" />
      </div>

      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }" @click="setStatusFilter('all')">
          全部
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Bound' }" @click="setStatusFilter('Bound')">
          Bound
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Pending' }" @click="setStatusFilter('Pending')">
          Pending
        </button>
        <button class="btn btn-filter" :class="{ active: statusFilter === 'Lost' }" @click="setStatusFilter('Lost')">
          Lost
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

        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
        <button class="btn btn-primary" @click="openCreateModal">创建 PVC</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedPVCs.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedPVCs.length }} 个 PVC</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 表格容器 -->
    <div class="table-container">
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
            <th style="width: 100px;">状态</th>
            <th style="min-width: 180px;">名称</th>
            <th style="width: 130px;">命名空间</th>
            <th style="width: 120px;">容量</th>
            <th style="min-width: 150px;">访问模式</th>
            <th style="min-width: 150px;">存储类</th>
            <th style="min-width: 150px;">绑定的 PV</th>
            <th style="width: 170px;">创建时间</th>
            <th style="width: 100px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pvc in paginatedPVCs" :key="pvc.name" :class="{ 'row-selected': isPVCSelected(pvc) }">
            <td v-if="batchMode">
              <input 
                type="checkbox" 
                :checked="isPVCSelected(pvc)" 
                @change="togglePVCSelection(pvc)"
              />
            </td>
            <td>
              <span class="status-indicator" :class="pvc.status.toLowerCase()">
                {{ pvc.status }}
              </span>
            </td>
            <td>
              <div class="pvc-name">
                <span class="icon">💾</span>
                <span>{{ pvc.name }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-badge">{{ pvc.namespace }}</span>
            </td>
            <td>{{ pvc.capacity }}</td>
            <td>
              <div class="access-modes">
                <span v-for="mode in pvc.accessModes" :key="mode" class="mode-badge">
                  {{ mode }}
                </span>
              </div>
            </td>
            <td>{{ pvc.storageClassName || '-' }}</td>
            <td>{{ pvc.volumeName || '-' }}</td>
            <td>{{ pvc.createdAt }}</td>
            <td>
              <div class="action-icons">
                <button class="icon-btn primary" @click="editYaml(pvc)" title="编辑">
                  ✏️
                </button>
                <button class="icon-btn expand" @click="openExpandModal(pvc)" title="扩容">
                  🔼
                </button>
                <button class="icon-btn danger" @click="deleteSinglePVC(pvc)" title="删除">
                  🗑️
                </button>
                
                <!-- 更多菜单 -->
                <div class="more-actions-wrapper">
                  <button 
                    class="icon-btn more-btn" 
                    @click="toggleMoreMenu(pvc)"
                    title="更多操作"
                  >
                    ⋮
                  </button>
                  <div 
                    v-if="activeMoreMenu === pvc.name" 
                    class="more-menu"
                    @click.stop
                  >
                    <button class="menu-item" @click="viewYaml(pvc)">
                      <span class="menu-icon">📄</span>
                      <span>查看 YAML</span>
                    </button>
                    <button class="menu-item" @click="downloadPVCYaml(pvc)">
                      <span class="menu-icon">💾</span>
                      <span>下载 YAML</span>
                    </button>
                  </div>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页组件 -->
    <Pagination 
      v-if="totalFromServer > 0"
      v-model:currentPage="currentPage"
      v-model:itemsPerPage="itemsPerPage"
      :totalItems="totalFromServer"
    />

    <div v-if="!loading && pvcs.length === 0" class="empty-state">
      <div class="empty-icon">📦</div>
      <div class="empty-text">暂无 PVC 数据</div>
    </div>

    <!-- 创建 PVC 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div 
        ref="createModalRef"
        class="modal-content modal-create-pvc resizable-modal"
        :style="createModalStyle"
      >
        <!-- 8个拖拽手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="createStartResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="createStartResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="createStartResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="createStartResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="createStartResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="createStartResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="createStartResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="createStartResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h2>💾 创建 PVC</h2>
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
          <button class="close-btn" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'">
            <form @submit.prevent="createPVC">
              <div class="form-section">
                <div class="section-header"><span class="section-icon">📋</span><h3>基本信息</h3></div>
                <div class="section-body">
                  <div class="form-group">
                    <label for="pvcName">PVC 名称 <span class="required">*</span></label>
                    <input type="text" id="pvcName" v-model="pvcForm.name" class="form-input" required placeholder="输入 PVC 名称" />
                  </div>
                  <div class="form-group">
                    <label for="pvcNamespace">命名空间 <span class="required">*</span></label>
                    <select id="pvcNamespace" v-model="pvcForm.namespace" class="form-select" required>
                      <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                    </select>
                  </div>
                </div>
              </div>
              
              <div class="form-section">
                <div class="section-header"><span class="section-icon">💾</span><h3>存储配置</h3></div>
                <div class="section-body">
                  <div class="form-group">
                    <label for="pvcStorage">容量 <span class="required">*</span></label>
                    <input type="text" id="pvcStorage" v-model="pvcForm.storage" class="form-input" required placeholder="例如: 10Gi" />
                    <div class="form-hint">支持单位：Gi、Mi、Ti</div>
                  </div>
                  <div class="form-group">
                    <label for="pvcStorageClass">存储类</label>
                    <input type="text" id="pvcStorageClass" v-model="pvcForm.storageClassName" class="form-input" placeholder="留空使用默认" />
                  </div>
                  <div class="form-group">
                    <label for="pvcAccessMode">访问模式 <span class="required">*</span></label>
                    <select id="pvcAccessMode" v-model="pvcForm.accessMode" class="form-select" required>
                      <option value="ReadWriteOnce">ReadWriteOnce (单节点读写)</option>
                      <option value="ReadOnlyMany">ReadOnlyMany (多节点只读)</option>
                      <option value="ReadWriteMany">ReadWriteMany (多节点读写)</option>
                    </select>
                  </div>
                </div>
              </div>
            </form>
          </div>
          
          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-editor-container">
            <div class="yaml-editor-header">
              <h3>📄 YAML 配置</h3>
              <div class="yaml-header-buttons">
                <button class="load-template-btn" @click="loadPVCYamlTemplate">
                  📁 加载模板
                </button>
                <button class="copy-yaml-btn" @click="copyYamlContent">
                  📋 复制
                </button>
                <button class="reset-yaml-btn" @click="resetYamlContent">
                  🔄 重置
                </button>
              </div>
            </div>
            
            <textarea 
              v-model="yamlContent" 
              class="yaml-editor"
              placeholder="输入或粘贴 YAML 内容..."
              spellcheck="false"
            ></textarea>
            
            <div v-if="yamlError" class="yaml-error">
              <span class="error-icon">⚠️</span>
              {{ yamlError }}
            </div>
            
            <div class="yaml-editor-footer">
              <div class="yaml-tips">
                <strong>💡 提示：</strong>
                <ul>
                  <li>支持完整的 Kubernetes PVC 配置</li>
                  <li>可以通过“加载模板”获取示例 YAML</li>
                  <li>创建前会验证 YAML 格式的正确性</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <!-- 表单模式按钮 -->
          <template v-if="createMode === 'form'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="button" class="btn btn-primary" @click="createPVC" :disabled="creating">{{ creating ? '创建中...' : '创建' }}</button>
          </template>
          
          <!-- YAML 模式按钮 -->
          <template v-if="createMode === 'yaml'">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button 
              type="button"
              class="btn btn-primary" 
              @click="createPVCFromYaml"
              :disabled="!yamlContent"
            >
              <span class="btn-icon">💾</span>从 YAML 创建
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- YAML 查看/编辑模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div 
        ref="yamlModalRef"
        class="modal-content yaml-modal resizable-modal"
        :style="yamlModalStyle"
      >
        <!-- 8个拖拽手柄 -->
        <div class="resize-handle resize-handle-top" @mousedown="yamlStartResize($event, 'top')"></div>
        <div class="resize-handle resize-handle-bottom" @mousedown="yamlStartResize($event, 'bottom')"></div>
        <div class="resize-handle resize-handle-left" @mousedown="yamlStartResize($event, 'left')"></div>
        <div class="resize-handle resize-handle-right" @mousedown="yamlStartResize($event, 'right')"></div>
        <div class="resize-handle resize-handle-top-left" @mousedown="yamlStartResize($event, 'top-left')"></div>
        <div class="resize-handle resize-handle-top-right" @mousedown="yamlStartResize($event, 'top-right')"></div>
        <div class="resize-handle resize-handle-bottom-left" @mousedown="yamlStartResize($event, 'bottom-left')"></div>
        <div class="resize-handle resize-handle-bottom-right" @mousedown="yamlStartResize($event, 'bottom-right')"></div>
        
        <div class="modal-header">
          <h3>📝 {{ yamlEditMode ? '编辑 YAML' : '查看 YAML' }} - {{ selectedYamlPVC?.name }}</h3>
          <div class="yaml-header-actions">
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="downloadYaml" :disabled="loadingYaml">
              📄 下载
            </button>
            <button v-if="!yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = true">
              ✈️ 编辑模式
            </button>
            <button v-if="yamlEditMode" class="btn btn-sm btn-secondary" @click="yamlEditMode = false">
              👁️ 预览模式
            </button>
            <button class="close-btn" @click="closeYamlModal">×</button>
          </div>
        </div>
        <div class="modal-body yaml-modal-body">
          <div v-if="loadingYaml" class="loading-state">Loading YAML...</div>
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

    <!-- PVC 扩容模态框 -->
    <div v-if="showExpandModal" class="modal-overlay" @click="showExpandModal = false">
      <div class="modal-content" @click.stop style="width: 600px; max-width: 90vw;">
        <div class="modal-header">
          <h2>🔼 PVC 扩容</h2>
          <button class="close-btn" @click="showExpandModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <!-- PVC 基本信息 -->
          <div class="info-box" style="margin-bottom: 20px;">
            <div><strong>PVC 名称:</strong> {{ expandForm.name }}</div>
            <div><strong>命名空间:</strong> {{ expandForm.namespace }}</div>
            <div><strong>状态:</strong> <span class="status-indicator" :class="expandForm.status.toLowerCase()">{{ expandForm.status }}</span></div>
            <div v-if="expandForm.storageClassName"><strong>存储类:</strong> {{ expandForm.storageClassName }}</div>
            <div v-if="expandForm.volumeName"><strong>绑定 PV:</strong> {{ expandForm.volumeName }}</div>
          </div>

          <!-- 容量预览对比 -->
          <div class="capacity-preview-box" style="margin-bottom: 20px;">
            <div class="capacity-header">📊 容量变更预览</div>
            <div class="capacity-comparison">
              <div class="capacity-item current">
                <div class="capacity-label">原容量</div>
                <div class="capacity-value">{{ expandForm.currentCapacity }}</div>
              </div>
              <div class="capacity-arrow">→</div>
              <div class="capacity-item new">
                <div class="capacity-label">目标容量</div>
                <div class="capacity-value" :class="{ placeholder: !expandForm.newCapacity }">
                  {{ expandForm.newCapacity || '待输入' }}
                </div>
              </div>
            </div>
            <div v-if="expandForm.newCapacity" class="capacity-diff">
              <span class="diff-label">增量:</span>
              <span class="diff-value">{{ calculateCapacityDiff() }}</span>
            </div>
          </div>

          <!-- 扩容限制说明 -->
          <div class="warning-box" style="margin-bottom: 20px;">
            <div class="warning-icon">⚠️</div>
            <div>
              <p><strong>扩容限制：</strong></p>
              <ul style="margin: 8px 0 0 20px; line-height: 1.8;">
                <li>⚡ <strong>只能扩大不能缩小</strong>：Kubernetes 不支持 PVC 容量缩减</li>
                <li>📦 <strong>StorageClass 支持</strong>：需要 StorageClass 设置 <code>allowVolumeExpansion: true</code></li>
                <li>💾 <strong>底层存储支持</strong>：动态供应的 PVC 需要底层驱动支持（AWS EBS、GCE PD、Ceph 等）</li>
                <li>🔄 <strong>Pod 重启</strong>：文件系统扩容可能需要重启使用该 PVC 的 Pod</li>
              </ul>
            </div>
          </div>

          <!-- 输入新容量 -->
          <div class="form-group">
            <label class="required">目标容量</label>
            <input 
              type="text" 
              v-model="expandForm.newCapacity" 
              class="form-input" 
              placeholder="例如: 20Gi" 
              required
            >
            <span class="form-hint">示例: 20Gi, 50Gi, 100Gi（必须大于原容量 {{ expandForm.currentCapacity }}）</span>
          </div>

          <div v-if="expandError" class="error-message">
            {{ expandError }}
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showExpandModal = false">取消</button>
          <button class="btn btn-primary" @click="expandPVC" :disabled="expanding || !expandForm.newCapacity">
            {{ expanding ? '扩容中...' : '确认扩容' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除预览（待补充）-->
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import pvcApi from '@/api/cluster/storage/pvc'
import namespaceApi from '@/api/cluster/config/namespace'
import { useClusterStore } from '@/stores/cluster'
import { useResizableModal } from '@/composables/useResizableModal'

// 获取认证头
const getAuthHeaders = () => {
  const headers = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const clusterStore = useClusterStore()
  const getClusterIdFromPath = () => {
    try {
      const m = window.location.pathname.match(/\/c\/([^/]+)/)
      return m ? decodeURIComponent(m[1]) : ''
    } catch {
      return ''
    }
  }
  const cid = clusterStore.current?.id ?? getClusterIdFromPath()
  if (cid !== undefined && cid !== null && cid !== '') {
    headers['X-Cluster-ID'] = String(cid)
  }
  return headers
}

// 状态变量
const loading = ref(false)
const errorMsg = ref('')
const searchQuery = ref('')
const statusFilter = ref('all')
const namespaceFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const namespaces = ref([])
const pvcs = ref([])
const totalFromServer = ref(0)

// 批量操作相关
const batchMode = ref(false)
const selectedPVCs = ref([])

// 模态框状态
const showCreateModal = ref(false)
const showBatchDeleteModal = ref(false)
const showYamlModal = ref(false)
const showExpandModal = ref(false)
const yamlEditMode = ref(false)

// 创建 PVC 表单数据
const pvcForm = ref({
  name: '',
  namespace: 'default',
  storage: '',
  storageClassName: '',
  accessMode: 'ReadWriteOnce'
})
const creating = ref(false)
const createMode = ref('form') // 'form' 或 'yaml'

// YAML 相关状态
const yamlContent = ref('')
const yamlError = ref('')
const loadingYaml = ref(false)
const savingYaml = ref(false)
const selectedYamlPVC = ref(null)

// 扩容相关状态
const expanding = ref(false)
const expandError = ref('')
const expandForm = ref({
  name: '',
  namespace: '',
  currentCapacity: '',
  newCapacity: '',
  status: '',
  storageClassName: '',
  volumeName: ''
})

// 自动刷新
const autoRefresh = ref(false)
let autoRefreshTimer = null
const AUTO_REFRESH_INTERVAL = 90000 // 90秒

// 防抖搜索
let searchDebounceTimer = null
const onSearchInput = () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1
    fetchPVCs()
  }, 500)
}

// 状态过滤
const setStatusFilter = (filter) => {
  statusFilter.value = filter
  currentPage.value = 1
  fetchPVCs()
}

// 监听过滤条件
watch(namespaceFilter, () => {
  currentPage.value = 1
  fetchPVCs()
})

// 批量操作
const enterBatchMode = () => {
  batchMode.value = true
  selectedPVCs.value = [...paginatedPVCs.value]
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedPVCs.value = []
}

const clearSelection = () => {
  selectedPVCs.value = []
}

const isPVCSelected = (pvc) => {
  return selectedPVCs.value.some(p => p.name === pvc.name && p.namespace === pvc.namespace)
}

const togglePVCSelection = (pvc) => {
  const index = selectedPVCs.value.findIndex(p => p.name === pvc.name && p.namespace === pvc.namespace)
  if (index >= 0) {
    selectedPVCs.value.splice(index, 1)
  } else {
    selectedPVCs.value.push(pvc)
  }
}

const isAllSelected = computed(() => {
  return paginatedPVCs.value.length > 0 && 
         paginatedPVCs.value.every(pvc => isPVCSelected(pvc))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedPVCs.value.forEach(pvc => {
      const index = selectedPVCs.value.findIndex(p => p.name === pvc.name && p.namespace === pvc.namespace)
      if (index >= 0) selectedPVCs.value.splice(index, 1)
    })
  } else {
    paginatedPVCs.value.forEach(pvc => {
      if (!isPVCSelected(pvc)) {
        selectedPVCs.value.push(pvc)
      }
    })
  }
}

// 分页计算
const paginatedPVCs = computed(() => pvcs.value)

// 获取命名空间列表
const fetchNamespaces = async () => {
  try {
    const { data } = await namespaceApi.list()
    if (data && data.list) {
      namespaces.value = data.list.map(ns => ns.name)
    }
  } catch (error) {
    console.error('获取命名空间列表失败:', error)
  }
}

// 获取 PVC 列表
const fetchPVCs = async () => {
  loading.value = true
  errorMsg.value = ''

  try {
    const params = {
      namespace: namespaceFilter.value || 'default',
      name: searchQuery.value,
      page: currentPage.value,
      limit: itemsPerPage.value
    }

    const { data } = await pvcApi.list(params)
    
    if (data && data.list) {
      pvcs.value = data.list
      totalFromServer.value = data.total || 0
    } else {
      pvcs.value = []
      totalFromServer.value = 0
    }
  } catch (error) {
    console.error('获取 PVC 列表失败:', error)
    errorMsg.value = '获取 PVC 列表失败：' + (error.response?.data?.message || error.message)
    pvcs.value = []
    totalFromServer.value = 0
  } finally {
    loading.value = false
  }
}

// 刷新列表
const refreshList = () => {
  fetchPVCs()
}

// 删除单个 PVC
const deleteSinglePVC = async (pvc) => {
  if (!confirm(`确定要删除 PVC "${pvc.name}" 吗？此操作无法撤销。`)) return
  
  try {
    await pvcApi.delete({
      namespace: pvc.namespace,
      name: pvc.name
    })
    Message.success({ content: `PVC "${pvc.name}" 删除成功`, duration: 2200 })
    fetchPVCs()
  } catch (error) {
    Message.error({ content: `删除失败：${error.response?.data?.message || error.message}`, duration: 3000 })
  }
}

// 批量删除预览
const openBatchDeletePreview = () => {
  if (selectedPVCs.value.length === 0) {
    Message.warning({ content: '请先选择要删除的 PVC', duration: 2200 })
    return
  }
  showBatchDeleteModal.value = true
}

// ==========================
// 更多菜单功能
// ==========================
const activeMoreMenu = ref(null)

const toggleMoreMenu = (pvc) => {
  if (activeMoreMenu.value === pvc.name) {
    activeMoreMenu.value = null
  } else {
    activeMoreMenu.value = pvc.name
  }
}

// 点击页面其他地方关闭菜单
const closeMoreMenu = () => {
  activeMoreMenu.value = null
}

// 下载 PVC YAML
const downloadPVCYaml = async (pvc) => {
  activeMoreMenu.value = null
  
  try {
    const { data } = await pvcApi.detail({
      namespace: pvc.namespace,
      name: pvc.name
    })
    
    const yamlContent = data.yaml || JSON.stringify(data, null, 2)
    const blob = new Blob([yamlContent], { type: 'text/yaml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${pvc.name}.yaml`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    Message.success({ content: `${pvc.name}.yaml 文件已下载`, duration: 2000 })
  } catch (error) {
    Message.error({ content: `下载失败：${error.response?.data?.message || error.message}`, duration: 3000 })
  }
}

// ==========================
// 创建 PVC 功能
// ==========================
const openCreateModal = () => {
  pvcForm.value = {
    name: '',
    namespace: namespaceFilter.value || 'default',
    storage: '',
    storageClassName: '',
    accessMode: 'ReadWriteOnce'
  }
  createMode.value = 'form'
  yamlContent.value = ''
  yamlError.value = ''
  showCreateModal.value = true
}

// 表单模式创建 PVC
const createPVC = async () => {
  creating.value = true
  try {
    const createData = {
      name: pvcForm.value.name,
      namespace: pvcForm.value.namespace,
      storage: pvcForm.value.storage,
      storageClassName: pvcForm.value.storageClassName || undefined,
      accessMode: pvcForm.value.accessMode
    }
    
    await pvcApi.create(createData)
    Message.success({ content: `PVC "${pvcForm.value.name}" 创建成功`, duration: 2200 })
    showCreateModal.value = false
    fetchPVCs()
  } catch (error) {
    Message.error({ content: `创建失败：${error.response?.data?.message || error.message}`, duration: 3000 })
  } finally {
    creating.value = false
  }
}

// 加载 PVC YAML 模板
const loadPVCYamlTemplate = () => {
  yamlContent.value = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: example-pvc
  namespace: ${pvcForm.value.namespace || 'default'}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard  # 可选，留空使用默认存储类`
  yamlError.value = ''
  Message.success({ content: '已加载 PVC YAML 模板', duration: 2000 })
}

// 复制 YAML 内容
const copyYamlContent = async () => {
  if (!yamlContent.value) {
    Message.warning({ content: 'YAML 内容为空', duration: 2000 })
    return
  }
  try {
    await navigator.clipboard.writeText(yamlContent.value)
    Message.success({ content: 'YAML 已复制到剪贴板', duration: 2000 })
  } catch (err) {
    Message.error({ content: '复制失败', duration: 2000 })
  }
}

// 重置 YAML 内容
const resetYamlContent = () => {
  if (confirm('确定要重置 YAML 内容吗？')) {
    yamlContent.value = ''
    yamlError.value = ''
    Message.info({ content: 'YAML 内容已重置', duration: 2000 })
  }
}

// 从 YAML 创建 PVC
const createPVCFromYaml = async () => {
  if (!yamlContent.value.trim()) {
    Message.warning({ content: '请输入 YAML 内容', duration: 2200 })
    return
  }
  
  creating.value = true
  yamlError.value = ''
  
  try {
    await pvcApi.createFromYaml({ yaml: yamlContent.value })
    Message.success({ content: 'PVC 创建成功', duration: 2200 })
    showCreateModal.value = false
    yamlContent.value = ''
    fetchPVCs()
  } catch (error) {
    const errMsg = error.response?.data?.message || error.message
    yamlError.value = errMsg
    Message.error({ content: `创建失败：${errMsg}`, duration: 3000 })
  } finally {
    creating.value = false
  }
}

// ==========================
// YAML 查看/编辑功能
// ==========================
// 查看 YAML
const viewYaml = async (pvc) => {
  selectedYamlPVC.value = pvc
  yamlEditMode.value = false
  showYamlModal.value = true
  loadingYaml.value = true
  yamlError.value = ''
  yamlContent.value = ''
  
  try {
    const { data } = await pvcApi.detail({
      namespace: pvc.namespace,
      name: pvc.name
    })
    yamlContent.value = data.yaml || JSON.stringify(data, null, 2)
  } catch (error) {
    yamlError.value = '加载 YAML 失败：' + (error.response?.data?.message || error.message)
  } finally {
    loadingYaml.value = false
  }
}

// 编辑 YAML
const editYaml = async (pvc) => {
  selectedYamlPVC.value = pvc
  yamlEditMode.value = true
  showYamlModal.value = true
  loadingYaml.value = true
  yamlError.value = ''
  yamlContent.value = ''
  
  try {
    const { data } = await pvcApi.detail({
      namespace: pvc.namespace,
      name: pvc.name
    })
    yamlContent.value = data.yaml || JSON.stringify(data, null, 2)
  } catch (error) {
    yamlError.value = '加载 YAML 失败：' + (error.response?.data?.message || error.message)
  } finally {
    loadingYaml.value = false
  }
}

// 关闭 YAML 模态框
const closeYamlModal = () => {
  showYamlModal.value = false
  yamlEditMode.value = false
  yamlContent.value = ''
  yamlError.value = ''
  selectedYamlPVC.value = null
}

// 应用 YAML 更改
const applyYamlChanges = async () => {
  if (!yamlContent.value.trim()) {
    Message.warning({ content: 'YAML 内容不能为空', duration: 2200 })
    return
  }
  
  if (!selectedYamlPVC.value) {
    Message.error({ content: '未选中 PVC', duration: 2200 })
    return
  }
  
  savingYaml.value = true
  yamlError.value = ''
  
  try {
    await pvcApi.applyYaml({
      namespace: selectedYamlPVC.value.namespace,
      name: selectedYamlPVC.value.name,
      yaml: yamlContent.value
    })
    Message.success({ content: 'YAML 更改已应用', duration: 2200 })
    closeYamlModal()
    fetchPVCs()
  } catch (error) {
    const errMsg = error.response?.data?.message || error.message
    yamlError.value = errMsg
    Message.error({ content: `应用失败：${errMsg}`, duration: 3000 })
  } finally {
    savingYaml.value = false
  }
}

// 下载 YAML
const downloadYaml = () => {
  if (!yamlContent.value) {
    Message.warning({ content: 'YAML 内容为空', duration: 2000 })
    return
  }
  
  const blob = new Blob([yamlContent.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedYamlPVC.value?.name || 'pvc'}.yaml`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  Message.success({ content: 'YAML 文件已下载', duration: 2000 })
}

// ==========================
// PVC 扩容功能
// ==========================
const openExpandModal = (pvc) => {
  expandForm.value = {
    name: pvc.name,
    namespace: pvc.namespace,
    currentCapacity: pvc.capacity,
    newCapacity: '',
    status: pvc.status,
    storageClassName: pvc.storageClassName || '未设置',
    volumeName: pvc.volumeName || ''
  }
  expandError.value = ''
  showExpandModal.value = true
}

const expandPVC = async () => {
  expanding.value = true
  expandError.value = ''
  try {
    // 验证新容量
    if (!expandForm.value.newCapacity) {
      expandError.value = '请输入目标容量'
      return
    }

    // 简单验证格式
    const capacityRegex = /^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$/
    if (!capacityRegex.test(expandForm.value.newCapacity)) {
      expandError.value = '无效的容量格式，示例: 20Gi, 50Gi'
      return
    }

    // 调用 API
    await pvcApi.resize({
      namespace: expandForm.value.namespace,
      name: expandForm.value.name,
      storage: expandForm.value.newCapacity
    })

    Message.success({ 
      content: `PVC ${expandForm.value.name} 扩容成功！新容量: ${expandForm.value.newCapacity}`,
      duration: 3000
    })
    showExpandModal.value = false
    fetchPVCs()
  } catch (error) {
    console.error('PVC 扩容失败:', error)
    expandError.value = error.response?.data?.message || error.message || '扩容失败'
    Message.error({ content: `扩容失败：${expandError.value}`, duration: 3000 })
  } finally {
    expanding.value = false
  }
}

// 计算容量增量差值
const calculateCapacityDiff = () => {
  if (!expandForm.value.newCapacity || !expandForm.value.currentCapacity) {
    return '-'
  }
  
  try {
    const parseCapacity = (cap) => {
      const match = cap.match(/^([\d.]+)(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$/)
      if (!match) return 0
      
      const value = parseFloat(match[1])
      const unit = match[2] || ''
      
      // 转换为字节
      const multipliers = {
        'Ki': 1024, 'Mi': 1024**2, 'Gi': 1024**3, 'Ti': 1024**4, 'Pi': 1024**5, 'Ei': 1024**6,
        'K': 1000, 'M': 1000**2, 'G': 1000**3, 'T': 1000**4, 'P': 1000**5, 'E': 1000**6
      }
      
      return value * (multipliers[unit] || 1)
    }
    
    const currentBytes = parseCapacity(expandForm.value.currentCapacity)
    const newBytes = parseCapacity(expandForm.value.newCapacity)
    const diffBytes = newBytes - currentBytes
    
    if (diffBytes <= 0) {
      return '无变化或减少（不允许）'
    }
    
    // 格式化输出
    const formatBytes = (bytes) => {
      if (bytes >= 1024**4) return `${(bytes / 1024**4).toFixed(2)}Ti`
      if (bytes >= 1024**3) return `${(bytes / 1024**3).toFixed(2)}Gi`
      if (bytes >= 1024**2) return `${(bytes / 1024**2).toFixed(2)}Mi`
      if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)}Ki`
      return `${bytes}B`
    }
    
    const percentage = ((diffBytes / currentBytes) * 100).toFixed(1)
    return `+${formatBytes(diffBytes)} (+${percentage}%)`
  } catch (e) {
    return '计算失败'
  }
}

// ==========================
// 可拖拽调整大小功能
// ==========================
const createModalRef = ref(null)
const yamlModalRef = ref(null)

// 创建模态框拖拽
const {
  modalRef: _createModalRef,
  modalStyle: createModalStyle,
  startResize: createStartResize
} = useResizableModal({ initialWidth: '800px', initialHeight: '70vh' })

// YAML 模态框拖拽
const {
  modalRef: _yamlModalRef,
  modalStyle: yamlModalStyle,
  startResize: yamlStartResize
} = useResizableModal({ initialWidth: '1000px', initialHeight: '80vh' })

// 同步 ref
createModalRef.value = _createModalRef.value
yamlModalRef.value = _yamlModalRef.value

// 自动刷新
const startAutoRefresh = () => {
  if (autoRefreshTimer) return
  autoRefreshTimer = setInterval(() => {
    if (!loading.value) fetchPVCs()
  }, AUTO_REFRESH_INTERVAL)
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

watch(autoRefresh, (newVal) => {
  if (newVal) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

// 监听分页变化
watch([currentPage, itemsPerPage], () => {
  fetchPVCs()
})

// 生命周期
onMounted(() => {
  fetchNamespaces().then(fetchPVCs)
  // 监听页面点击事件，关闭更多菜单
  document.addEventListener('click', closeMoreMenu)
})

onUnmounted(() => {
  stopAutoRefresh()
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  document.removeEventListener('click', closeMoreMenu)
})
</script>

<style scoped>
@import '@/assets/styles/resizable-modal.css';

.resource-view {
  max-width: 1400px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 32px;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 8px;
}

.view-header p {
  font-size: 16px;
  color: #718096;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.search-box input {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  width: 280px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.filter-buttons {
  display: flex;
  gap: 8px;
}

.btn-filter {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  color: #4a5568;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-filter:hover {
  border-color: #326ce5;
  color: #326ce5;
}

.btn-filter.active {
  background: #326ce5;
  border-color: #326ce5;
  color: white;
}

.filter-dropdown select {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  background-color: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #4a5568;
  cursor: pointer;
}

.auto-refresh-toggle input {
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

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
}

.btn-primary:hover {
  background-color: #2554c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.3);
}

.btn-secondary {
  background-color: #e2e8f0;
  color: #4a5568;
}

.btn-secondary:hover {
  background-color: #cbd5e0;
}

.btn-batch {
  background-color: #8b5cf6;
  color: white;
}

.error-box {
  background: #fff5f5;
  border: 1px solid #feb2b2;
  color: #c53030;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 16px;
  color: white;
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
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.batch-btn:hover {
  background: rgba(255,255,255,0.3);
}

.batch-btn.danger {
  background: rgba(239, 68, 68, 0.8);
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow-x: auto;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1400px;
}

.resource-table th {
  background-color: #f7fafc;
  text-align: left;
  padding: 16px 20px;
  font-size: 14px;
  font-weight: 600;
  color: #4a5568;
  border-bottom: 1px solid #e2e8f0;
}

.resource-table td {
  padding: 16px 20px;
  font-size: 14px;
  color: #2d3748;
  border-bottom: 1px solid #f7fafc;
}

.resource-table tbody tr:hover {
  background-color: #f7fafc;
}

.row-selected {
  background-color: #ebf5ff !important;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator.bound {
  background-color: rgba(52, 211, 153, 0.1);
  color: #34d399;
}

.status-indicator.pending {
  background-color: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.status-indicator.lost {
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.pvc-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.namespace-badge {
  background-color: #e2e8f0;
  color: #4a5568;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.access-modes {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.mode-badge {
  background-color: rgba(50, 108, 229, 0.1);
  color: #326ce5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.action-icons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.icon-btn {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  padding: 6px 10px;
  border-radius: 6px;
  transition: all 0.2s;
}

.icon-btn:hover {
  background-color: #e2e8f0;
  transform: scale(1.1);
}

.icon-btn.primary:hover {
  background-color: rgba(50, 108, 229, 0.1);
  color: #326ce5;
}

.icon-btn.danger:hover {
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.icon-btn.expand {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.icon-btn.expand:hover {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: scale(1.1);
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.3);
}

/* 更多菜单样式 */
.more-actions-wrapper {
  position: relative;
}

.more-btn {
  font-size: 20px;
  font-weight: bold;
  padding: 2px 8px;
}

.more-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  z-index: 1000;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: white;
  color: #2d3748;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.menu-item:hover {
  background-color: #f7fafc;
  color: #326ce5;
}

.menu-icon {
  font-size: 16px;
}

/* 扩容相关样式 */
.info-box {
  background: #f8f9fa;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}

.info-box > div {
  padding: 6px 0;
}

.capacity-preview-box {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px solid #38bdf8;
  border-radius: 12px;
  padding: 20px;
}

.capacity-header {
  font-size: 16px;
  font-weight: 600;
  color: #0c4a6e;
  margin-bottom: 16px;
  text-align: center;
}

.capacity-comparison {
  display: flex;
  align-items: center;
  justify-content: space-around;
  gap: 20px;
}

.capacity-item {
  flex: 1;
  text-align: center;
  padding: 16px;
  border-radius: 8px;
}

.capacity-item.current {
  background: rgba(156, 163, 175, 0.2);
  border: 2px solid #9ca3af;
}

.capacity-item.new {
  background: rgba(16, 185, 129, 0.1);
  border: 2px solid #10b981;
}

.capacity-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.capacity-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
}

.capacity-value.placeholder {
  color: #9ca3af;
  font-size: 18px;
}

.capacity-arrow {
  font-size: 28px;
  color: #10b981;
  font-weight: bold;
}

.capacity-diff {
  margin-top: 16px;
  text-align: center;
  padding: 12px;
  background: rgba(16, 185, 129, 0.1);
  border-radius: 6px;
}

.diff-label {
  font-size: 13px;
  color: #6b7280;
  margin-right: 8px;
}

.diff-value {
  font-size: 15px;
  font-weight: 600;
  color: #10b981;
}

.warning-box {
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  gap: 12px;
}

.warning-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.warning-box ul {
  margin: 0;
  padding-left: 0;
  list-style-position: inside;
}

.error-message {
  background: #fee2e2;
  border: 1px solid #fca5a5;
  color: #dc2626;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 14px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  color: #718096;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
}

/* ==========================
   模态框样式
   ========================== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h2,
.modal-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.close-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 28px;
  cursor: pointer;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: rotate(90deg);
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
  max-height: calc(85vh - 160px);
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background-color: #f7fafc;
}

/* ==========================
   表单样式
   ========================== */
.form-section {
  margin-bottom: 24px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.section-header {
  background: linear-gradient(135deg, #f7fafc 0%, #e2e8f0 100%);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
}

.section-icon {
  font-size: 18px;
}

.section-body {
  padding: 16px;
  background: white;
}

.form-group {
  margin-bottom: 16px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
}

.required {
  color: #ef4444;
}

.form-input,
.form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #cbd5e0;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #718096;
}

/* ==========================
   视图切换按钮
   ========================== */
.view-toggle-buttons {
  display: flex;
  gap: 8px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.view-toggle-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.view-toggle-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.view-toggle-btn.active {
  background: white;
  color: #667eea;
  border-color: white;
  font-weight: 600;
}

/* ==========================
   YAML 编辑器样式
   ========================== */
.yaml-editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.yaml-editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.yaml-editor-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
}

.yaml-header-buttons {
  display: flex;
  gap: 8px;
}

.load-template-btn,
.copy-yaml-btn,
.reset-yaml-btn {
  padding: 6px 12px;
  font-size: 12px;
  background: #e2e8f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.load-template-btn:hover,
.copy-yaml-btn:hover,
.reset-yaml-btn:hover {
  background: #cbd5e0;
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  padding: 16px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  border: 1px solid #cbd5e0;
  border-radius: 8px;
  background-color: #f7fafc;
  color: #2d3748;
  resize: vertical;
}

.yaml-editor:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.yaml-error {
  margin-top: 12px;
  padding: 12px;
  background: #fff5f5;
  border: 1px solid #feb2b2;
  border-radius: 6px;
  color: #c53030;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-icon {
  font-size: 18px;
}

.yaml-editor-footer {
  margin-top: 12px;
}

.yaml-tips {
  padding: 12px;
  background: #ebf8ff;
  border-left: 4px solid #4299e1;
  border-radius: 6px;
  font-size: 13px;
  color: #2c5282;
}

.yaml-tips strong {
  display: block;
  margin-bottom: 8px;
}

.yaml-tips ul {
  margin: 0;
  padding-left: 20px;
}

.yaml-tips li {
  margin-bottom: 4px;
}

/* ==========================
   YAML 模态框样式
   ========================== */
.yaml-modal {
  width: 900px;
  max-width: 95vw;
  height: 700px;
  max-height: 90vh;
}

.yaml-header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.btn-sm {
  padding: 8px 16px;
  font-size: 13px;
}

.yaml-modal-body {
  padding: 0;
}

.yaml-editor-wrapper {
  height: 100%;
  padding: 16px;
}

.yaml-content {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  background-color: #f7fafc;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  margin: 0;
  color: #2d3748;
}

.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 400px;
  color: #718096;
  font-size: 16px;
}

.error-box {
  background: #fff5f5;
  border: 1px solid #feb2b2;
  color: #c53030;
  padding: 12px 16px;
  border-radius: 8px;
  margin: 16px;
}

/* ==========================
   按钮样式
   ========================== */
.btn-icon {
  margin-right: 4px;
}


</style>

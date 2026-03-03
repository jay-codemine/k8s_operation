<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>服务管理</h1>
      <p>Kubernetes集群中的服务列表</p>
    </div>
    
    <!-- 类型过滤按钮 -->
    <div class="action-bar">
      <div class="filter-buttons">
        <button class="btn btn-filter" :class="{ active: typeFilter === '' }" @click="setTypeFilter('')">
          全部 <span class="filter-count">{{ services.length }}</span>
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'ClusterIP' }" @click="setTypeFilter('ClusterIP')">
          🔵 ClusterIP <span class="filter-count">{{ getTypeCount('ClusterIP') }}</span>
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'NodePort' }" @click="setTypeFilter('NodePort')">
          🟢 NodePort <span class="filter-count">{{ getTypeCount('NodePort') }}</span>
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'LoadBalancer' }" @click="setTypeFilter('LoadBalancer')">
          🟡 LoadBalancer <span class="filter-count">{{ getTypeCount('LoadBalancer') }}</span>
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'ExternalName' }" @click="setTypeFilter('ExternalName')">
          🔗 ExternalName <span class="filter-count">{{ getTypeCount('ExternalName') }}</span>
        </button>
        <button class="btn btn-filter" :class="{ active: typeFilter === 'Headless' }" @click="setTypeFilter('Headless')">
          ⚫ Headless <span class="filter-count">{{ getTypeCount('Headless') }}</span>
        </button>
      </div>

      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索服务名称..." @input="onSearchInput" />
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <button v-if="!batchMode" class="btn btn-batch" @click="enterBatchMode" title="进入批量操作模式">
          ☑️ 批量操作
        </button>
        <button v-if="batchMode" class="btn btn-secondary" @click="exitBatchMode">
          ✖️ 退出批量
        </button>
        <button class="btn btn-primary" @click="openCreateModal">+ 创建服务</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 分页大小选择器 -->
    <div class="pagination-controls">
      <select v-model.number="itemsPerPage" class="page-size-select" @change="onPageSizeChange">
        <option :value="10">10 条/页</option>
        <option :value="20">20 条/页</option>
        <option :value="50">50 条/页</option>
        <option :value="100">100 条/页</option>
      </select>
      <span class="total-info">共 {{ filteredServices.length }} 条</span>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedServices.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedServices.length }} 个 Service</span>
        <button class="batch-clear" @click="clearSelection">清空选择</button>
      </div>
      <div class="batch-actions">
        <button class="batch-btn danger" @click="openBatchDeletePreview" title="批量删除">
          🗑️ 批量删除
        </button>
      </div>
    </div>

    <!-- 表格 -->
    <div class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th v-if="batchMode" style="width: 40px;">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" title="全选/取消全选" />
            </th>
            <th>名称</th>
            <th>命名空间</th>
            <th>类型</th>
            <th>集群IP</th>
            <th>端口</th>
            <th>外部IP</th>
            <th>选择器</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="batchMode ? 11 : 10" class="loading-row">加载中...</td>
          </tr>
          <tr v-else-if="paginatedServices.length === 0">
            <td :colspan="batchMode ? 11 : 10" class="empty-row">暂无数据</td>
          </tr>
          <tr v-for="service in paginatedServices" :key="`${service.namespace}-${service.name}`" :class="{ 'row-selected': isServiceSelected(service) }">
            <td v-if="batchMode">
              <input type="checkbox" :checked="isServiceSelected(service)" @change="toggleServiceSelection(service)" />
            </td>
            <td>
              <div class="service-name">
                <span class="icon">🔌</span>
                <span>{{ service.name }}</span>
              </div>
            </td>
            <td><span class="namespace-badge">{{ service.namespace }}</span></td>
            <td>
              <span class="type-badge" :class="service.type.toLowerCase()">{{ service.type }}</span>
            </td>
            <td>{{ service.cluster_ip || '-' }}</td>
            <td>{{ service.ports || '-' }}</td>
            <td>{{ service.external_ip || '-' }}</td>
            <td>
              <div class="selector-tags" v-if="service.selector && Object.keys(service.selector).length > 0">
                <span v-for="(v, k) in getLimitedSelector(service.selector)" :key="k" class="selector-tag">{{ k }}={{ v }}</span>
                <span v-if="Object.keys(service.selector).length > 2" class="selector-tag more">+{{ Object.keys(service.selector).length - 2 }}</span>
              </div>
              <span v-else>-</span>
            </td>
            <td>{{ getAge(service.created_at) }}</td>
            <td>
              <div class="action-icons">
                <button class="icon-btn" title="查看详情" @click="viewService(service)">👁️</button>
                <button class="icon-btn" title="查看 YAML" @click="viewYaml(service)">📄</button>
                <button class="icon-btn" title="查看 Endpoints" @click="viewEndpoints(service)">🔗</button>
                <button class="icon-btn" title="关联 Deployment" @click="viewRelatedDeployments(service)">🚀</button>
                <button class="icon-btn danger" title="删除" @click="confirmDelete(service)">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <Pagination v-if="filteredServices.length > 0" v-model:currentPage="currentPage" :totalItems="filteredServices.length" :itemsPerPage="itemsPerPage" />
    </div>

    <!-- 创建服务模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>创建服务</h3>
          <div class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: createMode === 'form' }" @click="createMode = 'form'">📝 表单</button>
            <button class="view-toggle-btn" :class="{ active: createMode === 'yaml' }" @click="createMode = 'yaml'">📄 YAML</button>
          </div>
          <button class="close-btn" @click="closeCreateModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'" class="form-mode">
            <div class="form-row">
              <div class="form-group">
                <label>名称 <span class="required">*</span></label>
                <input v-model="serviceForm.name" type="text" placeholder="服务名称" />
              </div>
              <div class="form-group">
                <label>命名空间 <span class="required">*</span></label>
                <select v-model="serviceForm.namespace">
                  <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>类型 <span class="required">*</span></label>
                <select v-model="serviceForm.type">
                  <option value="ClusterIP">ClusterIP</option>
                  <option value="NodePort">NodePort</option>
                  <option value="LoadBalancer">LoadBalancer</option>
                  <option value="ExternalName">ExternalName</option>
                  <option value="Headless">Headless (ClusterIP=None)</option>
                </select>
              </div>
              <div class="form-group" v-if="serviceForm.type === 'ExternalName'">
                <label>外部域名 <span class="required">*</span></label>
                <input v-model="serviceForm.externalName" type="text" placeholder="例如: my-service.example.com" />
              </div>
            </div>
            <div class="form-group">
              <label>端口配置</label>
              <div v-for="(port, index) in serviceForm.ports" :key="index" class="port-row">
                <input v-model.number="port.port" type="number" placeholder="端口" class="port-input" />
                <input v-model.number="port.targetPort" type="number" placeholder="目标端口" class="port-input" />
                <input v-if="serviceForm.type === 'NodePort'" v-model.number="port.nodePort" type="number" placeholder="节点端口" class="port-input" />
                <select v-model="port.protocol" class="protocol-select">
                  <option value="TCP">TCP</option>
                  <option value="UDP">UDP</option>
                </select>
                <button class="btn-icon remove" @click="removePort(index)" v-if="serviceForm.ports.length > 1">✕</button>
              </div>
              <button class="btn btn-link" @click="addPort">+ 添加端口</button>
            </div>
            <div class="form-group">
              <label>选择器（格式: key=value，多个用逗号分隔）</label>
              <input v-model="selectorInput" type="text" placeholder="例如: app=nginx,version=v1" />
            </div>
          </div>

          <!-- YAML 模式 -->
          <div v-if="createMode === 'yaml'" class="yaml-mode">
            <div class="yaml-toolbar">
              <button class="btn btn-sm" @click="loadServiceYamlTemplate">📋 加载模板</button>
              <button class="btn btn-sm" @click="clearYamlContent">🗑️ 清空</button>
              <button class="btn btn-sm" @click="formatYaml">✨ 格式化</button>
            </div>
            <div class="yaml-editor-container">
              <div class="yaml-editor">
                <label>YAML 内容</label>
                <textarea v-model="createYamlContent" class="yaml-textarea" placeholder="粘贴或输入 Service YAML..."></textarea>
              </div>
              <div class="yaml-preview">
                <label>预览</label>
                <pre class="yaml-preview-content">{{ yamlPreviewContent || '输入 YAML 后显示预览...' }}</pre>
              </div>
            </div>
            <div v-if="createYamlError" class="yaml-error">{{ createYamlError }}</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeCreateModal">取消</button>
          <button v-if="createMode === 'form'" class="btn btn-primary" @click="createService" :disabled="creating">
            {{ creating ? '创建中...' : '创建' }}
          </button>
          <button v-if="createMode === 'yaml'" class="btn btn-primary" @click="createServiceFromYaml" :disabled="creating">
            {{ creating ? '创建中...' : '从 YAML 创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="showDetailModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>服务详情</h3>
          <button class="close-btn" @click="showDetailModal = false">×</button>
        </div>
        <div class="modal-body" v-if="detailData">
          <div class="detail-item"><span class="detail-label">名称:</span><span class="detail-value">{{ detailData.name }}</span></div>
          <div class="detail-item"><span class="detail-label">命名空间:</span><span class="detail-value">{{ detailData.namespace }}</span></div>
          <div class="detail-item"><span class="detail-label">类型:</span><span class="type-badge" :class="detailData.type?.toLowerCase()">{{ detailData.type }}</span></div>
          <div class="detail-item"><span class="detail-label">集群IP:</span><span class="detail-value">{{ detailData.cluster_ip || '-' }}</span></div>
          <div class="detail-item"><span class="detail-label">端口:</span><span class="detail-value">{{ detailData.ports || '-' }}</span></div>
          <div class="detail-item"><span class="detail-label">外部IP:</span><span class="detail-value">{{ detailData.external_ip || '-' }}</span></div>
          <div class="detail-item">
            <span class="detail-label">选择器:</span>
            <div class="selector-tags" v-if="detailData.selector">
              <span v-for="(v, k) in detailData.selector" :key="k" class="selector-tag">{{ k }}={{ v }}</span>
            </div>
            <span v-else>-</span>
          </div>
          <div class="detail-item"><span class="detail-label">创建时间:</span><span class="detail-value">{{ detailData.created_at }}</span></div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showDetailModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- YAML 查看模态框 -->
    <div v-if="showYamlModal" class="modal-overlay" @click.self="closeYamlModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Service YAML - {{ selectedService?.name }}</h3>
          <div class="yaml-mode-toggle">
            <button class="view-toggle-btn" :class="{ active: yamlViewMode === 'view' }" @click="yamlViewMode = 'view'">👁️ 查看</button>
            <button class="view-toggle-btn" :class="{ active: yamlViewMode === 'edit' }" @click="yamlViewMode = 'edit'">✏️ 编辑</button>
          </div>
          <button class="close-btn" @click="closeYamlModal">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingYaml" class="loading-state">加载中...</div>
          <div v-else>
            <pre v-if="yamlViewMode === 'view'" class="yaml-content">{{ yamlContent }}</pre>
            <textarea v-else v-model="yamlEditContent" class="yaml-textarea full-height"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="downloadYaml">⬇️ 下载</button>
          <button v-if="yamlViewMode === 'edit'" class="btn btn-primary" @click="applyYamlChanges" :disabled="applyingYaml">
            {{ applyingYaml ? '应用中...' : '应用更改' }}
          </button>
          <button class="btn btn-secondary" @click="closeYamlModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- Endpoints 模态框 -->
    <div v-if="showEndpointsModal" class="modal-overlay" @click.self="showEndpointsModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Endpoints - {{ selectedService?.name }}</h3>
          <button class="close-btn" @click="showEndpointsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingEndpoints" class="loading-state">加载中...</div>
          <div v-else-if="endpoints.length === 0" class="empty-state">
            <div class="empty-icon">🔗</div>
            <div class="empty-text">暂无 Endpoints</div>
          </div>
          <div v-else>
            <div v-for="(ep, idx) in endpoints" :key="idx" class="endpoint-item">
              <span class="endpoint-ip">{{ ep.ip }}</span>
              <span class="endpoint-port" v-if="ep.port">:{{ ep.port }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showEndpointsModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- Deployment 关联模态框 -->
    <div v-if="showDeploymentsModal" class="modal-overlay" @click.self="showDeploymentsModal = false">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>🚀 关联 Deployments</h3>
          <button class="close-btn" @click="showDeploymentsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="info-box">
            <div><strong>服务:</strong> {{ deploymentsService?.name }}</div>
            <div><strong>命名空间:</strong> {{ deploymentsService?.namespace }}</div>
            <div><strong>Deployment 数量:</strong> {{ deploymentsList.length }}</div>
            <div><strong>Selector:</strong> 
              <span class="selector-tags">
                <span v-for="(v, k) in deploymentsService?.selector" :key="k" class="selector-tag">
                  {{ k }}={{ v }}
                </span>
              </span>
            </div>
          </div>
          <div v-if="loadingDeployments" class="loading-state">加载 Deployments...</div>
          <div v-else-if="deploymentsList.length > 0">
            <table class="simple-table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>状态</th>
                  <th>副本数</th>
                  <th>镜像</th>
                  <th>创建时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="dep in deploymentsList" :key="dep.name">
                  <td>
                    <div class="deployment-name">
                      <span class="icon">🚀</span>
                      <span>{{ dep.name }}</span>
                    </div>
                  </td>
                  <td>
                    <span class="status-indicator" :class="(dep.status || 'unknown').toLowerCase()">
                      {{ dep.status || 'Unknown' }}
                    </span>
                  </td>
                  <td>{{ dep.readyReplicas }}/{{ dep.desiredReplicas }}</td>
                  <td>{{ dep.image || '-' }}</td>
                  <td>{{ dep.createdAt || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="empty-state-small">
            <div class="empty-icon">🚀</div>
            <div class="empty-text">暂无关联 Deployments</div>
            <div class="empty-hint">Deployment 的 labels 需要匹配 Service 的 selector</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="showDeploymentsModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>删除服务</h3>
          <button class="close-btn" @click="showDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除服务 <strong class="highlight">{{ selectedService?.name }}</strong> 吗？此操作不可撤销。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="deleteService" :disabled="deleting">
            {{ deleting ? '删除中...' : '删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="showBatchDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>批量删除服务</h3>
          <button class="close-btn" @click="showBatchDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除以下 <strong>{{ selectedServices.length }}</strong> 个服务吗？</p>
          <ul class="delete-list">
            <li v-for="s in selectedServices" :key="`${s.namespace}-${s.name}`">
              {{ s.namespace }}/{{ s.name }}
            </li>
          </ul>
          <div class="confirm-input">
            <label>请输入 <strong>DELETE</strong> 确认删除：</label>
            <input v-model="deleteConfirmText" type="text" placeholder="DELETE" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showBatchDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="executeBatchDelete" :disabled="deleteConfirmText !== 'DELETE' || batchDeleting">
            {{ batchDeleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import serviceApi from '@/api/cluster/networking/service'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import namespaceApi from '@/api/cluster/config/namespace'

// ==================== 状态变量 ====================
const services = ref([])
const namespaces = ref(['default'])
const loading = ref(false)
const errorMsg = ref('')

// 过滤与搜索
const searchQuery = ref('')
const namespaceFilter = ref('')
const typeFilter = ref('')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)
const total = ref(0)

// 批量操作
const batchMode = ref(false)
const selectedServices = ref([])
const deleteConfirmText = ref('')
const batchDeleting = ref(false)

// 模态框
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const showYamlModal = ref(false)
const showEndpointsModal = ref(false)
const showDeleteModal = ref(false)
const showBatchDeleteModal = ref(false)
const showDeploymentsModal = ref(false)  // Deployment 关联弹窗

// 选中的服务
const selectedService = ref(null)
const detailData = ref(null)

// 创建表单
const createMode = ref('form')
const creating = ref(false)
const serviceForm = ref({
  name: '',
  namespace: 'default',
  type: 'ClusterIP',
  ports: [{ port: 80, targetPort: 80, nodePort: null, protocol: 'TCP' }],
  externalName: ''
})
const selectorInput = ref('')
const createYamlContent = ref('')
const createYamlError = ref('')
const yamlPreviewContent = ref('')

// YAML 查看
const yamlContent = ref('')
const yamlEditContent = ref('')
const yamlViewMode = ref('view')
const loadingYaml = ref(false)
const applyingYaml = ref(false)

// Endpoints
const endpoints = ref([])
const loadingEndpoints = ref(false)

// 删除
const deleting = ref(false)

// Deployment 关联
const deploymentsList = ref([])          // 关联 Deployment 列表
const deploymentsService = ref(null)     // 查看 Deployment 的 Service
const loadingDeployments = ref(false)   // 加载 Deployment 状态

// ==================== 计算属性 ====================
const filteredServices = computed(() => {
  let result = services.value

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(query) ||
      s.namespace.toLowerCase().includes(query) ||
      s.cluster_ip?.toLowerCase().includes(query)
    )
  }

  if (namespaceFilter.value) {
    result = result.filter(s => s.namespace === namespaceFilter.value)
  }

  if (typeFilter.value) {
    if (typeFilter.value === 'Headless') {
      result = result.filter(s => s.cluster_ip === 'None')
    } else {
      result = result.filter(s => s.type === typeFilter.value)
    }
  }

  return result
})

const paginatedServices = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredServices.value.slice(startIndex, endIndex)
})

const isAllSelected = computed(() => {
  return paginatedServices.value.length > 0 && 
         paginatedServices.value.every(s => isServiceSelected(s))
})

// ==================== 方法 ====================
const fetchServices = async () => {
  try {
    loading.value = true
    errorMsg.value = ''
    const res = await serviceApi.list({ namespace: '', page: 1, limit: 1000 })
    if (res.code === 0 && res.data) {
      services.value = res.data.list || []
      total.value = res.data.total || services.value.length
    }
  } catch (error) {
    console.error('获取 Service 列表失败:', error)
    errorMsg.value = error.kube_message_error || error.message || '获取 Service 列表失败'
  } finally {
    loading.value = false
  }
}

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

const refreshList = () => {
  fetchServices()
}

const onSearchInput = () => {
  currentPage.value = 1
}

const onPageSizeChange = () => {
  currentPage.value = 1
}

const setTypeFilter = (type) => {
  typeFilter.value = type
  currentPage.value = 1
}

const getTypeCount = (type) => {
  if (type === 'Headless') {
    return services.value.filter(s => s.cluster_ip === 'None').length
  }
  return services.value.filter(s => s.type === type).length
}

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

const getLimitedSelector = (selector) => {
  if (!selector) return {}
  const entries = Object.entries(selector).slice(0, 2)
  return Object.fromEntries(entries)
}

// ==================== 批量操作 ====================
const enterBatchMode = () => {
  batchMode.value = true
  selectedServices.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedServices.value = []
}

const clearSelection = () => {
  selectedServices.value = []
}

const isServiceSelected = (service) => {
  return selectedServices.value.some(s => s.name === service.name && s.namespace === service.namespace)
}

const toggleServiceSelection = (service) => {
  const index = selectedServices.value.findIndex(s => s.name === service.name && s.namespace === service.namespace)
  if (index >= 0) {
    selectedServices.value.splice(index, 1)
  } else {
    selectedServices.value.push(service)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedServices.value.forEach(service => {
      const index = selectedServices.value.findIndex(s => s.name === service.name && s.namespace === service.namespace)
      if (index >= 0) selectedServices.value.splice(index, 1)
    })
  } else {
    paginatedServices.value.forEach(service => {
      if (!isServiceSelected(service)) {
        selectedServices.value.push(service)
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
  for (const service of selectedServices.value) {
    try {
      await serviceApi.delete({ namespace: service.namespace, name: service.name })
      successCount++
    } catch (e) {
      console.error(`删除 ${service.name} 失败:`, e)
      failCount++
    }
  }
  batchDeleting.value = false
  showBatchDeleteModal.value = false
  selectedServices.value = []
  alert(`批量删除完成：成功 ${successCount} 个，失败 ${failCount} 个`)
  fetchServices()
}

// ==================== 创建服务 ====================
const openCreateModal = () => {
  serviceForm.value = {
    name: '',
    namespace: 'default',
    type: 'ClusterIP',
    ports: [{ port: 80, targetPort: 80, nodePort: null, protocol: 'TCP' }],
    externalName: ''
  }
  selectorInput.value = ''
  createMode.value = 'form'
  createYamlContent.value = ''
  createYamlError.value = ''
  yamlPreviewContent.value = ''
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
}

const addPort = () => {
  serviceForm.value.ports.push({ port: 80, targetPort: 80, nodePort: null, protocol: 'TCP' })
}

const removePort = (index) => {
  serviceForm.value.ports.splice(index, 1)
}

const createService = async () => {
  if (!serviceForm.value.name || !serviceForm.value.namespace) {
    alert('请填写名称和命名空间')
    return
  }
  if (serviceForm.value.type === 'ExternalName' && !serviceForm.value.externalName) {
    alert('ExternalName 类型需要填写外部域名')
    return
  }
  creating.value = true
  try {
    // 解析选择器
    const selector = {}
    if (selectorInput.value.trim()) {
      selectorInput.value.split(',').forEach(pair => {
        const [k, v] = pair.split('=').map(s => s.trim())
        if (k && v) selector[k] = v
      })
    }
    
    const data = {
      namespace: serviceForm.value.namespace,
      name: serviceForm.value.name,
      type: serviceForm.value.type === 'Headless' ? 'ClusterIP' : serviceForm.value.type,
      selector_labels: Object.entries(selector).map(([k, v]) => ({ key: k, value: v })),
      ports: serviceForm.value.type === 'ExternalName' ? [] : serviceForm.value.ports.map(p => ({
        port: p.port,
        target_port: p.targetPort,
        node_port: serviceForm.value.type === 'NodePort' ? p.nodePort : null,
        protocol: p.protocol
      }))
    }
    
    // ExternalName 特殊字段
    if (serviceForm.value.type === 'ExternalName') {
      data.external_name = serviceForm.value.externalName
    }
    
    // Headless Service: ClusterIP = "None"
    if (serviceForm.value.type === 'Headless') {
      data.cluster_ip = 'None'
    }
    const res = await serviceApi.create(data)
    if (res.code === 0) {
      alert('Service 创建成功')
      closeCreateModal()
      fetchServices()
    } else {
      alert(res.msg || '创建失败')
    }
  } catch (e) {
    console.error('创建失败:', e)
    alert(e?.kube_message_error || e?.msg || e?.message || '创建失败')
  } finally {
    creating.value = false
  }
}

// ==================== YAML 创建 ====================
const loadServiceYamlTemplate = () => {
  createYamlContent.value = `# Service YAML 模板
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
  labels:
    app: my-app
spec:
  type: ClusterIP
  selector:
    app: my-app
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
---
# ExternalName Service 示例
apiVersion: v1
kind: Service
metadata:
  name: external-service
  namespace: default
spec:
  type: ExternalName
  externalName: my-service.example.com
---
# Headless Service 示例
apiVersion: v1
kind: Service
metadata:
  name: headless-service
  namespace: default
spec:
  clusterIP: None
  selector:
    app: my-app
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP`
  createYamlError.value = ''
  updateYamlPreview()
}

const clearYamlContent = () => {
  createYamlContent.value = ''
  yamlPreviewContent.value = ''
  createYamlError.value = ''
}

const formatYaml = () => {
  const lines = createYamlContent.value.split('\n')
  const formatted = lines.map(line => line.trimEnd()).join('\n')
  createYamlContent.value = formatted
  updateYamlPreview()
}

const updateYamlPreview = () => {
  if (!createYamlContent.value.trim()) {
    yamlPreviewContent.value = ''
    createYamlError.value = ''
    return
  }
  try {
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

watch(createYamlContent, () => {
  updateYamlPreview()
})

watch(createMode, (newMode) => {
  if (newMode === 'yaml' && !createYamlContent.value.trim()) {
    loadServiceYamlTemplate()
  }
})

const createServiceFromYaml = async () => {
  if (!createYamlContent.value.trim()) {
    createYamlError.value = '请输入 YAML 内容'
    return
  }
  if (!createYamlContent.value.includes('kind:')) {
    createYamlError.value = 'YAML 中必须包含 "kind:" 字段'
    return
  }
  creating.value = true
  createYamlError.value = ''
  try {
    const res = await serviceApi.createFromYaml({ yaml: createYamlContent.value })
    if (res.code === 0) {
      alert('Service 创建成功')
      closeCreateModal()
      fetchServices()
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

// ==================== 查看详情 ====================
const viewService = (service) => {
  detailData.value = service
  showDetailModal.value = true
}

// ==================== YAML 查看 ====================
const viewYaml = async (service) => {
  selectedService.value = service
  showYamlModal.value = true
  loadingYaml.value = true
  yamlViewMode.value = 'view'
  yamlContent.value = ''
  yamlEditContent.value = ''
  try {
    const res = await serviceApi.yaml({ namespace: service.namespace, name: service.name })
    if (res.code === 0 && res.data) {
      yamlContent.value = res.data.yaml || ''
      yamlEditContent.value = yamlContent.value
    }
  } catch (e) {
    console.error('获取 YAML 失败:', e)
    yamlContent.value = `# 获取 YAML 失败: ${e.message}`
  } finally {
    loadingYaml.value = false
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  selectedService.value = null
}

const downloadYaml = () => {
  const content = yamlViewMode.value === 'edit' ? yamlEditContent.value : yamlContent.value
  const blob = new Blob([content], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedService.value?.name || 'service'}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

const applyYamlChanges = async () => {
  if (!yamlEditContent.value.trim()) {
    alert('YAML 内容不能为空')
    return
  }
  applyingYaml.value = true
  try {
    const res = await serviceApi.applyYaml({ yaml: yamlEditContent.value })
    if (res.code === 0) {
      alert('YAML 应用成功')
      yamlContent.value = yamlEditContent.value
      yamlViewMode.value = 'view'
      fetchServices()
    } else {
      alert(res.msg || '应用失败')
    }
  } catch (e) {
    console.error('应用 YAML 失败:', e)
    alert(e?.kube_message_error || e?.msg || e?.message || '应用失败')
  } finally {
    applyingYaml.value = false
  }
}

// ==================== Endpoints ====================
const viewEndpoints = async (service) => {
  selectedService.value = service
  showEndpointsModal.value = true
  loadingEndpoints.value = true
  endpoints.value = []
  try {
    const res = await serviceApi.endpoints({ namespace: service.namespace, name: service.name })
    if (res.code === 0 && res.data) {
      endpoints.value = res.data.endpoints || []
    }
  } catch (e) {
    console.error('获取 Endpoints 失败:', e)
  } finally {
    loadingEndpoints.value = false
  }
}

// ==================== Deployment 关联 ====================
const viewRelatedDeployments = async (service) => {
  deploymentsService.value = service
  loadingDeployments.value = true
  deploymentsList.value = []
  showDeploymentsModal.value = true
  
  try {
    // 获取所有 Deployment 列表，然后在前端过滤
    const res = await deploymentsApi.list({ 
      namespace: service.namespace,
      page: 1,
      limit: 1000  // 获取足够多的数据进行过滤
    })
    
    if (res.code !== 0) {
      Message.error({ content: '获取 Deployment 列表失败' })
      return
    }
    
    const allDeployments = res.data?.list || res.data || []
    const serviceSelector = service.selector || {}
    
    // 如果 Service 没有 selector，不匹配任何 Deployment
    if (Object.keys(serviceSelector).length === 0) {
      Message.info({ 
        content: '该 Service 没有 selector，无法关联 Deployment',
        duration: 3000
      })
      return
    }
    
    // 过滤出匹配 Service selector 的 Deployment
    deploymentsList.value = allDeployments.filter(dep => {
      const depLabels = dep.selector || {}
      
      // 检查 Service 的 selector 是否全部匹配 Deployment 的 labels
      // Service selector 的每一个键值对都必须在 Deployment labels 中存在
      return Object.entries(serviceSelector).every(([key, value]) => {
        return depLabels[key] === value
      })
    }).map(dep => {
      // 转换为前端需要的格式
      // 后端字段: replicas(期望), ready_replicas(就绪)
      return {
        name: dep.name,
        namespace: dep.namespace,
        status: dep.status,
        readyReplicas: dep.readyReplicas || dep.ready_replicas || 0,
        desiredReplicas: dep.desiredReplicas || dep.desired_replicas || dep.replicas || 0,
        image: dep.image,
        selector: dep.selector,
        createdAt: dep.createdAt || dep.created_at,
        raw: dep
      }
    })
    
    if (deploymentsList.value.length === 0) {
      Message.info({ 
        content: `没有找到匹配的 Deployment，请确认 Deployment 的 labels 与 Service 的 selector 一致`,
        duration: 3000
      })
    }
  } catch (e) {
    console.error('获取关联 Deployment 失败:', e)
    Message.error({ content: '获取关联 Deployment 失败' })
  } finally {
    loadingDeployments.value = false
  }
}

// ==================== 删除 ====================
const confirmDelete = (service) => {
  selectedService.value = service
  showDeleteModal.value = true
}

const deleteService = async () => {
  deleting.value = true
  try {
    await serviceApi.delete({ namespace: selectedService.value.namespace, name: selectedService.value.name })
    showDeleteModal.value = false
    selectedService.value = null
    fetchServices()
  } catch (error) {
    console.error('删除失败:', error)
    alert(error.kube_message_error || error.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

// ==================== 生命周期 ====================
onMounted(() => {
  fetchServices()
  fetchNamespaces()
})
</script>

<style scoped>
.resource-view { max-width: 1600px; margin: 0 auto; padding: 20px; }
.view-header { margin-bottom: 24px; }
.view-header h1 { font-size: 28px; font-weight: 700; color: #2d3748; margin-bottom: 8px; }
.view-header p { font-size: 14px; color: #718096; }

.action-bar { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; margin-bottom: 16px; }
.filter-buttons { display: flex; gap: 8px; flex-wrap: wrap; }
.btn-filter { padding: 8px 16px; border: 1px solid #e2e8f0; border-radius: 20px; background: white; cursor: pointer; font-size: 13px; transition: all 0.2s; }
.btn-filter:hover { border-color: #326ce5; color: #326ce5; }
.btn-filter.active { background: #326ce5; color: white; border-color: #326ce5; }
.filter-count { font-size: 12px; opacity: 0.8; margin-left: 4px; }

.search-box input { padding: 10px 16px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; width: 250px; }
.filter-dropdown select { padding: 10px 16px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: white; }
.action-buttons { display: flex; gap: 8px; margin-left: auto; }

.btn { padding: 10px 20px; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.btn-primary { background: #326ce5; color: white; }
.btn-primary:hover { background: #2554c7; }
.btn-secondary { background: #e2e8f0; color: #4a5568; }
.btn-secondary:hover { background: #cbd5e0; }
.btn-batch { background: #805ad5; color: white; }
.btn-batch:hover { background: #6b46c1; }
.btn-danger { background: #e53e3e; color: white; }
.btn-danger:hover { background: #c53030; }
.btn-link { background: none; border: none; color: #326ce5; cursor: pointer; padding: 4px 8px; }

.pagination-controls { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.page-size-select { padding: 6px 12px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; }
.total-info { color: #718096; font-size: 13px; }

.error-box { background: #fed7d7; color: #c53030; padding: 12px 16px; border-radius: 8px; margin-bottom: 16px; }

.batch-action-bar { display: flex; justify-content: space-between; align-items: center; background: #805ad5; color: white; padding: 12px 20px; border-radius: 8px; margin-bottom: 16px; }
.batch-info { display: flex; align-items: center; gap: 12px; }
.batch-count { font-weight: 600; }
.batch-clear { background: rgba(255,255,255,0.2); border: none; color: white; padding: 4px 12px; border-radius: 4px; cursor: pointer; }
.batch-actions { display: flex; gap: 8px; }
.batch-btn { background: rgba(255,255,255,0.2); border: none; color: white; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 13px; }
.batch-btn.danger { background: #e53e3e; }

.table-container { background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); overflow: hidden; }
.resource-table { width: 100%; border-collapse: collapse; table-layout: auto; min-width: 1400px; }
.resource-table th { background: #f7fafc; text-align: left; padding: 14px 16px; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e2e8f0; white-space: nowrap; }
.resource-table td { padding: 14px 16px; font-size: 13px; color: #2d3748; border-bottom: 1px solid #f7fafc; }
.resource-table tbody tr:hover { background: #f7fafc; }
.row-selected { background: #ebf4ff !important; }

.loading-row, .empty-row { text-align: center; color: #718096; padding: 40px !important; }

.service-name { display: flex; align-items: center; gap: 8px; }
.service-name .icon { font-size: 16px; }
.namespace-badge { display: inline-block; padding: 4px 8px; background: rgba(50,108,229,0.1); color: #326ce5; border-radius: 4px; font-size: 12px; }

.type-badge { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.type-badge.clusterip { background: rgba(50,108,229,0.1); color: #326ce5; }
.type-badge.nodeport { background: rgba(16,185,129,0.1); color: #10b981; }
.type-badge.loadbalancer { background: rgba(245,158,11,0.1); color: #f59e0b; }
.type-badge.externalname { background: rgba(139,92,246,0.1); color: #8b5cf6; }
.type-badge.headless { background: rgba(107,114,128,0.1); color: #6b7280; }

.selector-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.selector-tag { display: inline-block; padding: 2px 6px; background: #edf2f7; color: #4a5568; border-radius: 3px; font-size: 11px; }
.selector-tag.more { background: #e2e8f0; color: #718096; }

.action-icons { display: flex; gap: 6px; }
.icon-btn { background: none; border: none; font-size: 14px; cursor: pointer; padding: 4px 6px; border-radius: 4px; transition: background 0.2s; }
.icon-btn:hover { background: #edf2f7; }
.icon-btn.danger:hover { background: #fed7d7; }

/* 模态框 */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; z-index: 1000; }
.modal-content { background: white; border-radius: 12px; width: 90%; max-width: 500px; max-height: 85vh; overflow-y: auto; }
.modal-large { max-width: 900px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px; border-bottom: 1px solid #e2e8f0; }
.modal-header h3 { margin: 0; font-size: 18px; font-weight: 600; }
.close-btn { background: none; border: none; font-size: 24px; color: #718096; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 20px; border-top: 1px solid #e2e8f0; }

.view-toggle-buttons { display: flex; gap: 4px; }
.view-toggle-btn { padding: 6px 12px; border: 1px solid #e2e8f0; background: white; cursor: pointer; font-size: 12px; border-radius: 4px; }
.view-toggle-btn.active { background: #326ce5; color: white; border-color: #326ce5; }

/* 表单 */
.form-row { display: flex; gap: 16px; }
.form-row .form-group { flex: 1; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; font-size: 14px; font-weight: 500; color: #4a5568; }
.form-group input, .form-group select, .form-group textarea { width: 100%; padding: 10px 12px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; }
.required { color: #e53e3e; }

.port-row { display: flex; gap: 8px; margin-bottom: 8px; align-items: center; }
.port-input { width: 100px !important; }
.protocol-select { width: 80px !important; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 14px; padding: 4px; }
.btn-icon.remove { color: #e53e3e; }

/* YAML 模式 */
.yaml-mode { height: 400px; display: flex; flex-direction: column; }
.yaml-toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.btn-sm { padding: 6px 12px; font-size: 12px; }
.yaml-editor-container { display: flex; gap: 16px; flex: 1; min-height: 0; }
.yaml-editor, .yaml-preview { flex: 1; display: flex; flex-direction: column; }
.yaml-textarea { flex: 1; font-family: monospace; font-size: 13px; padding: 12px; border: 1px solid #e2e8f0; border-radius: 8px; resize: none; }
.yaml-preview-content { flex: 1; background: #f7fafc; padding: 12px; border-radius: 8px; font-family: monospace; font-size: 13px; overflow: auto; white-space: pre-wrap; margin: 0; }
.yaml-error { color: #e53e3e; font-size: 13px; margin-top: 8px; padding: 8px; background: #fed7d7; border-radius: 4px; }
.yaml-content { background: #f7fafc; padding: 16px; border-radius: 8px; font-family: monospace; font-size: 13px; overflow: auto; white-space: pre-wrap; max-height: 400px; margin: 0; }
.full-height { height: 400px; }

/* 详情 */
.detail-item { display: flex; align-items: flex-start; margin-bottom: 12px; font-size: 14px; }
.detail-label { width: 100px; font-weight: 600; color: #4a5568; flex-shrink: 0; }
.detail-value { color: #2d3748; }
.highlight { color: #326ce5; font-weight: 600; }

/* Endpoints */
.endpoint-item { padding: 8px 12px; background: #f7fafc; border-radius: 4px; margin-bottom: 8px; font-family: monospace; font-size: 13px; }
.endpoint-ip { color: #2d3748; }
.endpoint-port { color: #718096; }

/* 批量删除 */
.delete-list { max-height: 200px; overflow-y: auto; background: #f7fafc; padding: 12px; border-radius: 8px; margin: 12px 0; }
.delete-list li { padding: 4px 0; font-size: 13px; color: #4a5568; }
.confirm-input { margin-top: 16px; }
.confirm-input label { display: block; margin-bottom: 8px; font-size: 14px; }
.confirm-input input { width: 100%; padding: 10px 12px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; }

.loading-state, .empty-state { text-align: center; padding: 40px; color: #718096; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-text { font-size: 14px; }

/* Deployment 关联 */
.info-box { background: #f7fafc; padding: 16px; border-radius: 8px; margin-bottom: 16px; }
.info-box > div { margin-bottom: 8px; font-size: 14px; }
.info-box > div:last-child { margin-bottom: 0; }
.info-box strong { color: #4a5568; font-weight: 600; }

.simple-table { width: 100%; border-collapse: collapse; margin-top: 16px; }
.simple-table th { background: #f7fafc; padding: 12px; text-align: left; font-size: 13px; font-weight: 600; color: #4a5568; border-bottom: 2px solid #e2e8f0; }
.simple-table td { padding: 12px; font-size: 13px; color: #2d3748; border-bottom: 1px solid #f7fafc; }
.simple-table tbody tr:hover { background: #f7fafc; }

.deployment-name { display: flex; align-items: center; gap: 8px; }
.deployment-name .icon { font-size: 16px; }

.status-indicator { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.status-indicator.running { background: rgba(16,185,129,0.1); color: #10b981; }
.status-indicator.updating { background: rgba(245,158,11,0.1); color: #f59e0b; }
.status-indicator.stopped { background: rgba(107,114,128,0.1); color: #6b7280; }
.status-indicator.failed { background: rgba(229,62,62,0.1); color: #e53e3e; }
.status-indicator.unknown { background: rgba(203,213,224,0.1); color: #718096; }

.empty-state-small { text-align: center; padding: 40px 20px; color: #718096; }
.empty-state-small .empty-icon { font-size: 40px; margin-bottom: 12px; opacity: 0.5; }
.empty-state-small .empty-text { font-size: 14px; font-weight: 600; margin-bottom: 4px; }
.empty-state-small .empty-hint { font-size: 12px; color: #a0aec0; }
</style>

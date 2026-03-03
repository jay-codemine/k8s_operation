<template>
  <div class="resource-view">
    <div class="view-header">
      <h1>Ingress 管理</h1>
      <p>Kubernetes 集群中的 Ingress 列表</p>
    </div>
    
    <!-- 过滤和操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input type="text" v-model="searchQuery" placeholder="搜索 Ingress 名称、主机..." @input="onSearchInput" />
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
        <button class="btn btn-primary" @click="openCreateModal">+ 创建 Ingress</button>
        <button class="btn btn-secondary" @click="refreshList" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 批量操作浮动栏 -->
    <div v-if="batchMode && selectedIngresses.length > 0" class="batch-action-bar">
      <div class="batch-info">
        <span class="batch-count">已选择 {{ selectedIngresses.length }} 个 Ingress</span>
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
            <th>主机</th>
            <th>路径</th>
            <th>服务</th>
            <th>Class</th>
            <th>TLS</th>
            <th>Address</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="batchMode ? 11 : 10" class="loading-row">加载中...</td>
          </tr>
          <tr v-else-if="paginatedIngresses.length === 0">
            <td :colspan="batchMode ? 11 : 10" class="empty-row">暂无数据</td>
          </tr>
          <tr v-for="ing in paginatedIngresses" :key="`${ing.namespace}-${ing.name}`" :class="{ 'row-selected': isIngressSelected(ing) }">
            <td v-if="batchMode">
              <input type="checkbox" :checked="isIngressSelected(ing)" @change="toggleIngressSelection(ing)" />
            </td>
            <td>
              <div class="ingress-name">
                <span class="icon">🌐</span>
                <span>{{ ing.name }}</span>
              </div>
            </td>
            <td><span class="namespace-badge">{{ ing.namespace }}</span></td>
            <td>{{ ing.hosts || '-' }}</td>
            <td>{{ ing.paths || '-' }}</td>
            <td>{{ ing.services || '-' }}</td>
            <td>{{ ing.ingress_class || '-' }}</td>
            <td>{{ ing.tls ? '✅' : '-' }}</td>
            <td>{{ ing.address || '-' }}</td>
            <td>{{ getAge(ing.created_at) }}</td>
            <td>
              <div class="action-icons">
                <button class="icon-btn" title="查看 YAML" @click="viewYaml(ing)">📄</button>
                <button class="icon-btn danger" title="删除" @click="confirmDelete(ing)">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页（现代化三段式布局） -->
      <div v-if="filteredIngresses.length > 0" class="pagination-wrapper">
        <div class="pagination-left">
          <span class="pagination-summary">共 <strong>{{ filteredIngresses.length }}</strong> 条</span>
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
    </div>

    <!-- 创建 Ingress 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>创建 Ingress</h3>
          <div class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: createMode === 'form' }" @click="createMode = 'form'">📝 表单</button>
            <button class="view-toggle-btn" :class="{ active: createMode === 'yaml' }" @click="createMode = 'yaml'">📄 YAML</button>
          </div>
          <button class="close-btn" @click="closeCreateModal">×</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="createMode === 'form'" class="form-mode">
            <!-- 基本信息 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">📋</span>
                <h3>基本信息</h3>
              </div>
              <div class="section-body">
                <div class="form-row">
                  <div class="form-group">
                    <label>名称 <span class="required">*</span></label>
                    <input v-model="ingressForm.name" type="text" placeholder="例如: my-ingress" class="form-input" />
                    <div class="form-hint">用于标识此 Ingress，建议使用小写字母和连字符</div>
                  </div>
                  <div class="form-group">
                    <label>命名空间 <span class="required">*</span></label>
                    <div class="namespace-controls">
                      <div v-if="!showNamespaceInput" class="namespace-select-group">
                        <select v-model="ingressForm.namespace" class="form-select">
                          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
                        </select>
                        <button type="button" class="btn btn-secondary btn-sm" @click="showNamespaceInput = true">
                          ➕ 创建新命名空间
                        </button>
                      </div>
                      <div v-else class="namespace-input-group">
                        <input v-model="newNamespace" type="text" placeholder="输入新命名空间" class="form-input" />
                        <button type="button" class="btn btn-primary btn-sm" @click="createNamespace">确定</button>
                        <button type="button" class="btn btn-secondary btn-sm" @click="cancelNamespaceInput">取消</button>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="form-row">
                  <div class="form-group">
                    <label>Ingress Class</label>
                    <input v-model="ingressForm.ingressClass" type="text" placeholder="例如: nginx" class="form-input" />
                    <div class="form-hint">指定 Ingress 控制器类型，留空使用默认</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 规则配置 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🌐</span>
                <h3>规则配置</h3>
              </div>
              <div class="section-body">
                <div v-for="(rule, ruleIndex) in ingressForm.rules" :key="ruleIndex" class="rule-card">
                  <div class="rule-header">
                    <span class="rule-label">规则 {{ ruleIndex + 1 }}</span>
                    <button v-if="ingressForm.rules.length > 1" class="btn-remove" @click="removeRule(ruleIndex)" type="button">
                      🗑️ 删除规则
                    </button>
                  </div>
                  
                  <div class="form-group">
                    <label>主机名 <span class="required">*</span></label>
                    <input v-model="rule.host" type="text" placeholder="例如: example.com" class="form-input" />
                    <div class="form-hint">访问的域名，支持通配符如 *.example.com</div>
                  </div>

                  <div class="paths-section">
                    <label>路径规则</label>
                    <div v-for="(path, pathIndex) in rule.paths" :key="pathIndex" class="path-row">
                      <div class="path-fields">
                        <div class="path-field">
                          <input v-model="path.path" type="text" placeholder="路径 例如: /" class="path-input" />
                        </div>
                        <div class="path-field">
                          <select v-model="path.pathType" class="pathtype-select">
                            <option value="Prefix">Prefix（前缀）</option>
                            <option value="Exact">Exact（精确）</option>
                            <option value="ImplementationSpecific">ImplementationSpecific</option>
                          </select>
                        </div>
                        <div class="path-field">
                          <input v-model="path.serviceName" type="text" placeholder="服务名" class="service-input" />
                        </div>
                        <div class="path-field">
                          <input v-model.number="path.servicePort" type="number" placeholder="端口" class="port-input" />
                        </div>
                        <button class="btn-icon danger" @click="removePath(ruleIndex, pathIndex)" type="button" title="删除路径">
                          ✖
                        </button>
                      </div>
                    </div>
                    <button class="btn-add-path" @click="addPath(ruleIndex)" type="button">
                      + 添加路径
                    </button>
                  </div>
                </div>
                <button class="btn-add-rule" @click="addRule" type="button">
                  + 添加规则
                </button>
              </div>
            </div>

            <!-- TLS 配置 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🔒</span>
                <h3>TLS 配置（可选）</h3>
              </div>
              <div class="section-body">
                <div class="form-group">
                  <label class="toggle-label">
                    <input type="checkbox" v-model="ingressForm.enableTls" />
                    <span>启用 TLS/SSL</span>
                  </label>
                </div>
                <div v-if="ingressForm.enableTls">
                  <div v-for="(tls, tlsIndex) in ingressForm.tls" :key="tlsIndex" class="tls-card">
                    <div class="tls-header">
                      <span class="tls-label">TLS {{ tlsIndex + 1 }}</span>
                      <button v-if="ingressForm.tls.length > 1" class="btn-remove" @click="removeTls(tlsIndex)" type="button">
                        🗑️ 删除
                      </button>
                    </div>
                    <div class="form-group">
                      <label>Secret 名称 <span class="required">*</span></label>
                      <input v-model="tls.secretName" type="text" placeholder="tls-secret" class="form-input" />
                      <div class="form-hint">包含 TLS 证书的 Secret 名称</div>
                    </div>
                    <div class="form-group">
                      <label>主机列表</label>
                      <input v-model="tls.hostsInput" type="text" placeholder="多个主机用逗号分隔，例如: example.com,www.example.com" class="form-input" />
                    </div>
                  </div>
                  <button class="btn-add-tls" @click="addTls" type="button">
                    + 添加 TLS 配置
                  </button>
                </div>
              </div>
            </div>

            <!-- 注解配置 -->
            <div class="form-section">
              <div class="section-header">
                <span class="section-icon">🏷️</span>
                <h3>注解（Annotations）</h3>
              </div>
              <div class="section-body">
                <div class="annotations-list">
                  <div v-for="(anno, index) in ingressForm.annotations" :key="index" class="annotation-row">
                    <input v-model="anno.key" type="text" placeholder="键" class="anno-key" />
                    <input v-model="anno.value" type="text" placeholder="值" class="anno-value" />
                    <button class="btn-icon danger" @click="removeAnnotation(index)" type="button">✖</button>
                  </div>
                </div>
                <button class="btn-add-annotation" @click="addAnnotation" type="button">
                  + 添加注解
                </button>
                <div class="form-hint">
                  常用注解示例：nginx.ingress.kubernetes.io/rewrite-target: /
                </div>
              </div>
            </div>
          </div>

          <!-- YAML 模式 -->
          <div v-else class="yaml-mode">
            <div class="yaml-toolbar">
              <button class="btn btn-small" @click="loadIngressYamlTemplate">📝 加载模板</button>
              <button class="btn btn-small" @click="updateYamlPreview">🔄 更新预览</button>
              <span v-if="createYamlError" class="yaml-error">❌ {{ createYamlError }}</span>
            </div>
            <textarea v-model="createYamlContent" class="yaml-editor" placeholder="粘贴或编辑 YAML 配置..."></textarea>
            <div v-if="yamlPreview" class="yaml-preview">
              <h4>YAML 预览</h4>
              <pre>{{ yamlPreview }}</pre>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeCreateModal">取消</button>
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
          <h3>{{ currentIngress.name }} - YAML</h3>
          <div class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: yamlMode === 'view' }" @click="yamlMode = 'view'">👁️ 查看</button>
            <button class="view-toggle-btn" :class="{ active: yamlMode === 'edit' }" @click="yamlMode = 'edit'">✏️ 编辑</button>
          </div>
          <button class="close-btn" @click="closeYamlModal">×</button>
        </div>
        <div class="modal-body">
          <div class="yaml-actions">
            <button v-if="yamlMode === 'view'" class="btn btn-small" @click="downloadYaml">📥 下载 YAML</button>
            <button v-if="yamlMode === 'edit'" class="btn btn-small" @click="validateYaml">✅ 验证 YAML</button>
          </div>
          <div v-if="yamlMode === 'view'">
            <pre class="yaml-viewer">{{ currentYaml }}</pre>
          </div>
          <div v-else>
            <textarea v-model="editYamlContent" class="yaml-editor"></textarea>
            <div v-if="yamlValidationError" class="error-message">{{ yamlValidationError }}</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeYamlModal">关闭</button>
          <button v-if="yamlMode === 'edit'" class="btn btn-primary" @click="applyYaml" :disabled="applying">
            {{ applying ? '应用中...' : '应用更改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content modal-small">
        <div class="modal-header">
          <h3>确认删除</h3>
          <button class="close-btn" @click="showDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除 Ingress <strong>{{ deleteTarget.name }}</strong> 吗？</p>
          <p class="text-danger">此操作不可恢复！</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="deleteIngress" :disabled="deleting">
            {{ deleting ? '删除中...' : '删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除预览模态框 -->
    <div v-if="showBatchDeleteModal" class="modal-overlay" @click.self="showBatchDeleteModal = false">
      <div class="modal-content modal-medium">
        <div class="modal-header">
          <h3>批量删除预览</h3>
          <button class="close-btn" @click="showBatchDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>即将删除 <strong>{{ selectedIngresses.length }}</strong> 个 Ingress：</p>
          <ul class="delete-preview-list">
            <li v-for="ing in selectedIngresses" :key="`${ing.namespace}-${ing.name}`">
              {{ ing.namespace }} / {{ ing.name }}
            </li>
          </ul>
          <p class="text-danger">⚠️ 此操作不可恢复！</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showBatchDeleteModal = false">取消</button>
          <button class="btn btn-danger" @click="executeBatchDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ingressApi from '@/api/cluster/networking/ingress'
import namespaceApi from '@/api/cluster/config/namespace'

const router = useRouter()

// 状态
const ingresses = ref([])
const loading = ref(false)
const errorMsg = ref('')
const searchQuery = ref('')
const namespaceFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const jumpPage = ref(1)

// 批量操作
const batchMode = ref(false)
const selectedIngresses = ref([])

// 模态框
const showCreateModal = ref(false)
const showYamlModal = ref(false)
const showDeleteModal = ref(false)
const showBatchDeleteModal = ref(false)

// 创建相关
const createMode = ref('form')
const createYamlContent = ref('')
const yamlPreview = ref('')
const createYamlError = ref('')
const creating = ref(false)

// 表单数据
const ingressForm = ref({
  name: '',
  namespace: 'default',
  ingressClass: '',
  rules: [
    {
      host: '',
      paths: [
        {
          path: '/',
          pathType: 'Prefix',
          serviceName: '',
          servicePort: 80
        }
      ]
    }
  ],
  enableTls: false,
  tls: [
    {
      secretName: '',
      hostsInput: ''
    }
  ],
  annotations: []
})

// 命名空间创建
const showNamespaceInput = ref(false)
const newNamespace = ref('')

// YAML 查看/编辑
const currentIngress = ref({})
const currentYaml = ref('')
const yamlMode = ref('view')
const editYamlContent = ref('')
const yamlValidationError = ref('')
const applying = ref(false)

// 删除相关
const deleteTarget = ref({})
const deleting = ref(false)

// 命名空间列表（动态获取）
const namespaces = ref([])

// 过滤后的 Ingress 列表
const filteredIngresses = computed(() => {
  let result = ingresses.value

  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(ing =>
      ing.name.toLowerCase().includes(query) ||
      ing.namespace.toLowerCase().includes(query) ||
      (ing.hosts && ing.hosts.toLowerCase().includes(query))
    )
  }

  // 命名空间过滤
  if (namespaceFilter.value) {
    result = result.filter(ing => ing.namespace === namespaceFilter.value)
  }

  return result
})

// 分页后的列表
const paginatedIngresses = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredIngresses.value.slice(startIndex, endIndex)
})

// 全选状态
const isAllSelected = computed(() => {
  return paginatedIngresses.value.length > 0 && paginatedIngresses.value.every(ing => isIngressSelected(ing))
})

// 分页相关计算属性
const totalPages = computed(() => Math.ceil(filteredIngresses.value.length / itemsPerPage.value) || 1)

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

// 获取命名空间列表
const fetchNamespaces = async () => {
  try {
    const res = await namespaceApi.list({ page: 1, limit: 1000 })
    const list = res?.data?.list || res?.data?.items || []
    namespaces.value = (Array.isArray(list) ? list : []).map(ns =>
      typeof ns === 'string' ? ns : (ns?.metadata?.name || ns?.name || ns)
    ).filter(Boolean)
    
    // 如果没有命名空间，提供默认值
    if (namespaces.value.length === 0) {
      namespaces.value = ['default']
    }
  } catch (e) {
    console.error('获取命名空间失败:', e)
    namespaces.value = ['default', 'kube-system']
  }
}

// 加载列表
const loadIngressList = async () => {
  loading.value = true
  errorMsg.value = ''
  
  try {
    const params = {
      namespace: namespaceFilter.value || '',
      name: '',
      page: 1,
      limit: 1000
    }
    
    const res = await ingressApi.list(params)
    
    if (res.code === 0) {
      ingresses.value = res.data.list || []
    } else {
      errorMsg.value = res.message || '获取 Ingress 列表失败'
    }
  } catch (error) {
    console.error('加载 Ingress 列表失败:', error)
    errorMsg.value = error.message || '加载失败'
  } finally {
    loading.value = false
  }
}

// 刷新列表
const refreshList = () => {
  loadIngressList()
}

// 搜索防抖
let searchTimeout = null
const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
  }, 300)
}

// 分页大小变化
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

// 时间格式化
const getAge = (createdAt) => {
  if (!createdAt) return '-'
  const created = new Date(createdAt)
  const now = new Date()
  const diffMs = now - created
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  const diffHours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  
  if (diffDays > 0) return `${diffDays}天前`
  if (diffHours > 0) return `${diffHours}小时前`
  return '刚刚'
}

// 批量操作
const enterBatchMode = () => {
  batchMode.value = true
  selectedIngresses.value = []
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedIngresses.value = []
}

const isIngressSelected = (ing) => {
  return selectedIngresses.value.some(s => s.namespace === ing.namespace && s.name === ing.name)
}

const toggleIngressSelection = (ing) => {
  if (isIngressSelected(ing)) {
    selectedIngresses.value = selectedIngresses.value.filter(s => !(s.namespace === ing.namespace && s.name === ing.name))
  } else {
    selectedIngresses.value.push(ing)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    paginatedIngresses.value.forEach(ing => {
      if (isIngressSelected(ing)) {
        selectedIngresses.value = selectedIngresses.value.filter(s => !(s.namespace === ing.namespace && s.name === ing.name))
      }
    })
  } else {
    paginatedIngresses.value.forEach(ing => {
      if (!isIngressSelected(ing)) {
        selectedIngresses.value.push(ing)
      }
    })
  }
}

const clearSelection = () => {
  selectedIngresses.value = []
}

// 创建 Ingress
const openCreateModal = () => {
  createMode.value = 'form'
  createYamlContent.value = ''
  yamlPreview.value = ''
  createYamlError.value = ''
  showNamespaceInput.value = false
  newNamespace.value = ''
  // 重置表单
  ingressForm.value = {
    name: '',
    namespace: namespaces.value[0] || 'default',
    ingressClass: '',
    rules: [
      {
        host: '',
        paths: [
          {
            path: '/',
            pathType: 'Prefix',
            serviceName: '',
            servicePort: 80
          }
        ]
      }
    ],
    enableTls: false,
    tls: [
      {
        secretName: '',
        hostsInput: ''
      }
    ],
    annotations: []
  }
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
  createYamlContent.value = ''
  yamlPreview.value = ''
  createYamlError.value = ''
  showNamespaceInput.value = false
  newNamespace.value = ''
}

// 命名空间操作
const createNamespace = async () => {
  if (!newNamespace.value || !newNamespace.value.trim()) {
    alert('请输入命名空间名称')
    return
  }
  
  const ns = newNamespace.value.trim()
  if (namespaces.value.includes(ns)) {
    alert('命名空间已存在')
    return
  }
  
  try {
    // 调用 API 创建命名空间
    const res = await namespaceApi.create({ name: ns })
    if (res.code === 0) {
      alert('命名空间创建成功！')
      // 刷新命名空间列表
      await fetchNamespaces()
      // 选中新创建的命名空间
      ingressForm.value.namespace = ns
      showNamespaceInput.value = false
      newNamespace.value = ''
    } else {
      alert('创建失败：' + (res.message || '未知错误'))
    }
  } catch (e) {
    console.error('创建命名空间失败:', e)
    alert('创建失败：' + (e.message || '网络错误'))
  }
}

const cancelNamespaceInput = () => {
  showNamespaceInput.value = false
  newNamespace.value = ''
}

// 表单操作方法
const addRule = () => {
  ingressForm.value.rules.push({
    host: '',
    paths: [
      {
        path: '/',
        pathType: 'Prefix',
        serviceName: '',
        servicePort: 80
      }
    ]
  })
}

const removeRule = (index) => {
  if (ingressForm.value.rules.length > 1) {
    ingressForm.value.rules.splice(index, 1)
  }
}

const addPath = (ruleIndex) => {
  ingressForm.value.rules[ruleIndex].paths.push({
    path: '/',
    pathType: 'Prefix',
    serviceName: '',
    servicePort: 80
  })
}

const removePath = (ruleIndex, pathIndex) => {
  const rule = ingressForm.value.rules[ruleIndex]
  if (rule.paths.length > 1) {
    rule.paths.splice(pathIndex, 1)
  }
}

const addTls = () => {
  ingressForm.value.tls.push({
    secretName: '',
    hostsInput: ''
  })
}

const removeTls = (index) => {
  if (ingressForm.value.tls.length > 1) {
    ingressForm.value.tls.splice(index, 1)
  }
}

const addAnnotation = () => {
  ingressForm.value.annotations.push({
    key: '',
    value: ''
  })
}

const removeAnnotation = (index) => {
  ingressForm.value.annotations.splice(index, 1)
}

// 从表单创建
const createFromForm = async () => {
  // 表单验证
  if (!ingressForm.value.name) {
    alert('请填写 Ingress 名称')
    return
  }
  if (!ingressForm.value.namespace) {
    alert('请选择命名空间')
    return
  }
  
  // 验证规则
  for (const rule of ingressForm.value.rules) {
    if (!rule.host) {
      alert('请填写所有规则的主机名')
      return
    }
    for (const path of rule.paths) {
      if (!path.serviceName || !path.servicePort) {
        alert('请填写所有路径的服务名和端口')
        return
      }
    }
  }

  creating.value = true
  createYamlError.value = ''

  try {
    // 构建 YAML 字符串
    let yamlStr = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${ingressForm.value.name}
  namespace: ${ingressForm.value.namespace}`

    // 注解
    if (ingressForm.value.annotations.length > 0) {
      yamlStr += '\n  annotations:'
      ingressForm.value.annotations.forEach(anno => {
        if (anno.key && anno.value) {
          yamlStr += `\n    ${anno.key}: "${anno.value}"`
        }
      })
    }

    yamlStr += '\nspec:'

    // IngressClass
    if (ingressForm.value.ingressClass) {
      yamlStr += `\n  ingressClassName: ${ingressForm.value.ingressClass}`
    }

    // 规则
    yamlStr += '\n  rules:'
    ingressForm.value.rules.forEach(rule => {
      yamlStr += `\n    - host: ${rule.host}`
      yamlStr += '\n      http:'
      yamlStr += '\n        paths:'
      rule.paths.forEach(path => {
        yamlStr += `\n          - path: ${path.path}`
        yamlStr += `\n            pathType: ${path.pathType}`
        yamlStr += '\n            backend:'
        yamlStr += '\n              service:'
        yamlStr += `\n                name: ${path.serviceName}`
        yamlStr += '\n                port:'
        yamlStr += `\n                  number: ${path.servicePort}`
      })
    })

    // TLS
    if (ingressForm.value.enableTls) {
      const validTls = ingressForm.value.tls.filter(tls => tls.secretName)
      if (validTls.length > 0) {
        yamlStr += '\n  tls:'
        validTls.forEach(tls => {
          yamlStr += `\n    - secretName: ${tls.secretName}`
          if (tls.hostsInput) {
            const hosts = tls.hostsInput.split(',').map(h => h.trim()).filter(h => h)
            if (hosts.length > 0) {
              yamlStr += '\n      hosts:'
              hosts.forEach(host => {
                yamlStr += `\n        - ${host}`
              })
            }
          }
        })
      }
    }
    
    const res = await ingressApi.createFromYaml({ yaml: yamlStr })
    
    if (res.code === 0) {
      alert('Ingress 创建成功！')
      closeCreateModal()
      refreshList()
    } else {
      createYamlError.value = res.kube_message_error || res.message || '创建失败'
    }
  } catch (error) {
    console.error('创建 Ingress 失败:', error)
    createYamlError.value = error.message || '创建失败'
  } finally {
    creating.value = false
  }
}

const loadIngressYamlTemplate = () => {
  createYamlContent.value = `# Ingress YAML 模板
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
    - host: example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 80
  # TLS 配置（可选）
  # tls:
  #   - hosts:
  #       - example.com
  #     secretName: tls-secret`
  createYamlError.value = ''
  updateYamlPreview()
}

const updateYamlPreview = () => {
  yamlPreview.value = createYamlContent.value
  createYamlError.value = ''
}

const createFromYaml = async () => {
  if (!createYamlContent.value.trim()) {
    createYamlError.value = '请输入 YAML 内容'
    return
  }

  creating.value = true
  createYamlError.value = ''

  try {
    const res = await ingressApi.createFromYaml({ yaml: createYamlContent.value })
    
    if (res.code === 0) {
      alert('Ingress 创建成功！')
      closeCreateModal()
      refreshList()
    } else {
      createYamlError.value = res.kube_message_error || res.message || '创建失败'
    }
  } catch (error) {
    console.error('创建 Ingress 失败:', error)
    createYamlError.value = error.message || '创建失败'
  } finally {
    creating.value = false
  }
}

// 查看 YAML
const viewYaml = async (ing) => {
  currentIngress.value = ing
  yamlMode.value = 'view'
  yamlValidationError.value = ''
  
  try {
    const res = await ingressApi.yaml({ namespace: ing.namespace, name: ing.name })
    if (res.code === 0) {
      currentYaml.value = res.data.yaml
      editYamlContent.value = res.data.yaml
      showYamlModal.value = true
    } else {
      alert(res.message || '获取 YAML 失败')
    }
  } catch (error) {
    console.error('获取 YAML 失败:', error)
    alert(error.message || '获取 YAML 失败')
  }
}

const closeYamlModal = () => {
  showYamlModal.value = false
  currentIngress.value = {}
  currentYaml.value = ''
  editYamlContent.value = ''
  yamlValidationError.value = ''
}

const validateYaml = () => {
  try {
    // 简单验证 YAML 格式
    if (!editYamlContent.value.includes('kind:') || !editYamlContent.value.includes('Ingress')) {
      yamlValidationError.value = 'YAML 必须包含 kind: Ingress'
      return false
    }
    yamlValidationError.value = ''
    alert('YAML 格式验证通过')
    return true
  } catch (error) {
    yamlValidationError.value = '无效的 YAML 格式'
    return false
  }
}

const applyYaml = async () => {
  if (!validateYaml()) return

  applying.value = true
  yamlValidationError.value = ''

  try {
    const res = await ingressApi.applyYaml({ yaml: editYamlContent.value })
    
    if (res.code === 0) {
      alert('YAML 应用成功！')
      closeYamlModal()
      refreshList()
    } else {
      yamlValidationError.value = res.kube_message_error || res.message || '应用失败'
    }
  } catch (error) {
    console.error('应用 YAML 失败:', error)
    yamlValidationError.value = error.message || '应用失败'
  } finally {
    applying.value = false
  }
}

const downloadYaml = () => {
  const blob = new Blob([currentYaml.value], { type: 'text/yaml' })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${currentIngress.value.name}-ingress.yaml`
  a.click()
  window.URL.revokeObjectURL(url)
}

// 删除 Ingress
const confirmDelete = (ing) => {
  deleteTarget.value = ing
  showDeleteModal.value = true
}

const deleteIngress = async () => {
  deleting.value = true

  try {
    const res = await ingressApi.delete({
      namespace: deleteTarget.value.namespace,
      name: deleteTarget.value.name
    })

    if (res.code === 0) {
      alert('Ingress 删除成功！')
      showDeleteModal.value = false
      refreshList()
    } else {
      alert(res.message || '删除失败')
    }
  } catch (error) {
    console.error('删除失败:', error)
    alert(error.message || '删除失败')
  } finally {
    deleting.value = false
  }
}

// 批量删除
const openBatchDeletePreview = () => {
  if (selectedIngresses.value.length === 0) {
    alert('请先选择要删除的 Ingress')
    return
  }
  showBatchDeleteModal.value = true
}

const executeBatchDelete = async () => {
  deleting.value = true

  try {
    const deletePromises = selectedIngresses.value.map(ing =>
      ingressApi.delete({ namespace: ing.namespace, name: ing.name })
    )

    await Promise.all(deletePromises)
    
    alert(`成功删除 ${selectedIngresses.value.length} 个 Ingress`)
    showBatchDeleteModal.value = false
    selectedIngresses.value = []
    refreshList()
  } catch (error) {
    console.error('批量删除失败:', error)
    alert(error.message || '批量删除失败')
  } finally {
    deleting.value = false
  }
}

// 初始化
onMounted(() => {
  fetchNamespaces().then(() => {
    loadIngressList()
  })
})
</script>

<style scoped>
/* 使用与 Services.vue 相同的样式 */
.resource-view {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 8px 0;
}

.view-header p {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

.action-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 250px;
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
  cursor: pointer;
}

.action-buttons {
  display: flex;
  gap: 8px;
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
  background: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background: #4b5563;
}

.btn-batch {
  background: #8b5cf6;
  color: white;
}

.btn-batch:hover {
  background: #7c3aed;
}

.pagination-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
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
}

.page-jump-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.total-info {
  font-size: 14px;
  color: #6b7280;
}

.error-box {
  background: #fee2e2;
  color: #991b1b;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.batch-action-bar {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 16px 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  gap: 24px;
  align-items: center;
  z-index: 100;
}

.batch-info {
  display: flex;
  gap: 12px;
  align-items: center;
}

.batch-count {
  font-weight: 600;
  color: #1f2937;
}

.batch-clear {
  padding: 4px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background: white;
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
  transition: all 0.2s;
}

.batch-btn.danger {
  background: #ef4444;
  color: white;
}

.batch-btn.danger:hover {
  background: #dc2626;
}

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1200px;
}

.resource-table thead {
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.resource-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  font-size: 13px;
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
  color: #6b7280;
  padding: 40px 16px !important;
}

.ingress-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.icon {
  font-size: 18px;
}

.namespace-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.action-icons {
  display: flex;
  gap: 8px;
}

.icon-btn {
  padding: 6px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.icon-btn.danger {
  color: #dc2626;
  border-color: #fecaca;
}

.icon-btn.danger:hover {
  background: #fee2e2;
  border-color: #ef4444;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.2);
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-small {
  width: 90%;
  max-width: 450px;
}

.modal-medium {
  width: 90%;
  max-width: 600px;
}

.modal-large {
  width: 90%;
  max-width: 900px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.view-toggle-buttons {
  display: flex;
  gap: 4px;
}

.view-toggle-btn {
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  background: white;
  border-radius: 6px;
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
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 24px;
  cursor: pointer;
  color: #6b7280;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
}

.yaml-mode {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.yaml-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn-small {
  padding: 6px 12px;
  font-size: 13px;
}

.yaml-error {
  color: #dc2626;
  font-size: 13px;
}

.yaml-editor {
  width: 100%;
  min-height: 400px;
  padding: 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
}

.yaml-preview {
  border: 1px solid #d1d5db;
  border-radius: 6px;
  padding: 12px;
  background: #f9fafb;
}

.yaml-preview h4 {
  font-size: 14px;
  margin: 0 0 8px 0;
  color: #374151;
}

.yaml-preview pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #1f2937;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.yaml-actions {
  margin-bottom: 12px;
}

.yaml-viewer {
  background: #f9fafb;
  padding: 16px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-message {
  margin-top: 8px;
  padding: 8px 12px;
  background: #fee2e2;
  color: #991b1b;
  border-radius: 4px;
  font-size: 13px;
}

.text-danger {
  color: #dc2626;
  font-weight: 500;
  margin-top: 8px;
}

.delete-preview-list {
  max-height: 300px;
  overflow-y: auto;
  background: #f9fafb;
  border-radius: 6px;
  padding: 12px;
  margin: 12px 0;
}

.delete-preview-list li {
  padding: 6px 0;
  border-bottom: 1px solid #e5e7eb;
  font-size: 14px;
}

.delete-preview-list li:last-child {
  border-bottom: none;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 表单模式样式 */
.form-mode {
  max-height: 70vh;
  overflow-y: auto;
  padding-right: 8px;
}

.form-section {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 20px;
  overflow: hidden;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.section-icon {
  font-size: 20px;
}

.section-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.section-body {
  padding: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
  margin-left: 2px;
}

.form-input,
.form-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
  background: white;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-hint {
  font-size: 12px;
  color: #6b7280;
}

/* 命名空间控件 */
.namespace-controls {
  width: 100%;
}

.namespace-select-group,
.namespace-input-group {
  display: flex;
  gap: 8px;
  align-items: center;
}

.namespace-select-group select {
  flex: 1;
}

.namespace-input-group input {
  flex: 1;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
  white-space: nowrap;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}

.toggle-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

/* 规则卡片 */
.rule-card,
.tls-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 12px;
}

.rule-header,
.tls-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rule-label,
.tls-label {
  font-weight: 600;
  color: #1f2937;
}

.btn-remove {
  padding: 4px 12px;
  border: 1px solid #fecaca;
  background: white;
  color: #dc2626;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-remove:hover {
  background: #fee2e2;
  border-color: #ef4444;
}

/* 路径配置 */
.paths-section {
  margin-top: 12px;
}

.path-row {
  margin-bottom: 8px;
}

.path-fields {
  display: grid;
  grid-template-columns: 2fr 1.5fr 2fr 1fr auto;
  gap: 8px;
  align-items: center;
}

.path-field input,
.path-field select {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
}

.btn-icon {
  width: 32px;
  height: 32px;
  border: 1px solid #d1d5db;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-icon.danger {
  color: #dc2626;
  border-color: #fecaca;
}

.btn-icon.danger:hover {
  background: #fee2e2;
  border-color: #ef4444;
}

.btn-add-path,
.btn-add-rule,
.btn-add-tls,
.btn-add-annotation {
  padding: 8px 16px;
  border: 1px dashed #9ca3af;
  background: white;
  color: #4b5563;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 8px;
}

.btn-add-path:hover,
.btn-add-rule:hover,
.btn-add-tls:hover,
.btn-add-annotation:hover {
  background: #f3f4f6;
  border-color: #6b7280;
  color: #1f2937;
}

/* 注解配置 */
.annotations-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.annotation-row {
  display: grid;
  grid-template-columns: 1fr 2fr auto;
  gap: 8px;
  align-items: center;
}

.anno-key,
.anno-value {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
}
</style>

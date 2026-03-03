 <template>
  <div class="resource-view">
    <!-- 页面头部 -->
    <div class="view-header">
      <h1>🔐 ServiceAccount 管理</h1>
      <p>管理 Kubernetes 服务账户，用于 Pod 身份认证和权限控制</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 ServiceAccount 名称..."
          @input="onSearchInput"
        />
      </div>

      <div class="filter-dropdown">
        <select v-model="namespaceFilter" @change="loadServiceAccounts">
          <option value="">所有命名空间</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>

      <div class="action-buttons">
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>

        <button class="btn btn-primary" @click="showCreateModal = true">+ 创建 ServiceAccount</button>
        <button class="btn btn-secondary" @click="loadServiceAccounts" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && serviceAccounts.length === 0" class="loading-state">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 表格视图 -->
    <div v-else class="table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th style="width: 200px;">名称</th>
            <th style="width: 150px;">命名空间</th>
            <th style="width: 100px;">Secrets</th>
            <th style="width: 150px;">创建时间</th>
            <th style="width: 260px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sa in filteredServiceAccounts" :key="`${sa.namespace}-${sa.name}`">
            <td>
              <div class="resource-name">
                <span class="icon">👤</span>
                <span>{{ sa.name }}</span>
              </div>
            </td>
            <td>
              <span class="namespace-tag">{{ sa.namespace }}</span>
            </td>
            <td>{{ sa.secrets?.length || 0 }}</td>
            <td>{{ formatDate(sa.created_at) }}</td>
            <td class="actions">
              <button class="btn btn-sm btn-info" @click="viewDetails(sa)" title="查看详情">
                👁️ 详情
              </button>
              <button class="btn btn-sm btn-success" @click="viewBindings(sa)" title="查看绑定">
                🔗 绑定
              </button>
              <button class="btn btn-sm btn-warning" @click="editServiceAccount(sa)" title="编辑">
                ✏️ 编辑
              </button>
              <button class="btn btn-sm btn-danger" @click="deleteServiceAccount(sa)" title="删除">
                🗑️ 删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="filteredServiceAccounts.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">
          {{ searchQuery ? '没有匹配的 ServiceAccount' : '暂无 ServiceAccount，请先创建' }}
        </div>
      </div>
    </div>

    <!-- 创建 ServiceAccount 模态框 -->
    <div v-if="showCreateModal" class="modal" @click.self="closeCreateModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>创建 ServiceAccount</h2>
          <button class="close-btn" @click="closeCreateModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="createServiceAccount">
            <div class="form-group">
              <label>名称 *</label>
              <input v-model="saForm.name" type="text" placeholder="例如：my-service-account" required />
            </div>

            <div class="form-group">
              <label>命名空间 *</label>
              <select v-model="saForm.namespace" required>
                <option value="">请选择命名空间</option>
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>

            <div class="form-group">
              <label>标签（可选）</label>
              <div class="label-input-group">
                <div v-for="(label, index) in saForm.labels" :key="index" class="label-row">
                  <input v-model="label.key" type="text" placeholder="键" />
                  <input v-model="label.value" type="text" placeholder="值" />
                  <button type="button" class="btn btn-sm btn-danger" @click="removeLabel(index)">-</button>
                </div>
                <button type="button" class="btn btn-sm btn-secondary" @click="addLabel">+ 添加标签</button>
              </div>
            </div>

            <div class="form-group">
              <label>
                <input type="checkbox" v-model="saForm.autoMountToken" />
                自动挂载 ServiceAccount Token
              </label>
              <p class="help-text">选中后，使用此 ServiceAccount 的 Pod 会自动挂载 Token</p>
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? '创建中...' : '创建' }}
              </button>
              <button type="button" class="btn btn-secondary" @click="closeCreateModal">取消</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailsModal" class="modal" @click.self="closeDetailsModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h2>ServiceAccount 详情</h2>
          <button class="close-btn" @click="closeDetailsModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h3>基本信息</h3>
            <div class="detail-grid">
              <div class="detail-item">
                <label>名称</label>
                <span>{{ currentSA?.name }}</span>
              </div>
              <div class="detail-item">
                <label>命名空间</label>
                <span>{{ currentSA?.namespace }}</span>
              </div>
              <div class="detail-item">
                <label>创建时间</label>
                <span>{{ formatDate(currentSA?.created_at) }}</span>
              </div>
              <div class="detail-item">
                <label>自动挂载 Token</label>
                <span>{{ currentSA?.automount_token ? '是' : '否' }}</span>
              </div>
            </div>
          </div>

          <div class="detail-section">
            <h3>关联的 Secrets ({{ currentSA?.secrets?.length || 0 }})</h3>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Secret 名称</th>
                  <th>类型</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="secret in currentSA?.secrets" :key="secret.name">
                  <td>{{ secret.name }}</td>
                  <td>{{ secret.type }}</td>
                  <td>
                    <button class="btn btn-sm btn-info" @click="viewSecret(secret)">查看</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="detail-section">
            <h3>标签</h3>
            <div class="labels-display">
              <span v-for="(value, key) in currentSA?.labels" :key="key" class="label-tag">
                {{ key }}={{ value }}
              </span>
              <span v-if="!currentSA?.labels || Object.keys(currentSA.labels).length === 0" class="empty-hint">
                无标签
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Message } from '@arco-design/web-vue'

// API 导入（待创建）
// import serviceAccountApi from '@/api/cluster/rbac/serviceaccount'
// import namespaceApi from '@/api/cluster/config/namespace'

// 数据状态
const loading = ref(false)
const searchQuery = ref('')
const namespaceFilter = ref('')
const autoRefresh = ref(false)
const serviceAccounts = ref([])
const namespaces = ref(['default', 'kube-system', 'kube-public', 'kube-node-lease'])

// 模态框状态
const showCreateModal = ref(false)
const showDetailsModal = ref(false)
const currentSA = ref(null)

// 表单数据
const saForm = ref({
  name: '',
  namespace: 'default',
  labels: [],
  autoMountToken: true
})

// 过滤后的数据
const filteredServiceAccounts = computed(() => {
  let result = serviceAccounts.value

  // 命名空间过滤
  if (namespaceFilter.value) {
    result = result.filter(sa => sa.namespace === namespaceFilter.value)
  }

  // 搜索过滤
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(sa => sa.name.toLowerCase().includes(query))
  }

  return result
})

// 加载 ServiceAccounts
const loadServiceAccounts = async () => {
  loading.value = true
  try {
    // TODO: 调用真实 API
    // const res = await serviceAccountApi.list({ namespace: namespaceFilter.value })
    // serviceAccounts.value = res.data?.list || []

    // 模拟数据
    serviceAccounts.value = [
      {
        name: 'default',
        namespace: 'default',
        secrets: [{ name: 'default-token-xxxxx', type: 'kubernetes.io/service-account-token' }],
        automount_token: true,
        created_at: '2024-01-15T10:30:00Z',
        labels: { 'app': 'system' }
      },
      {
        name: 'admin-sa',
        namespace: 'kube-system',
        secrets: [{ name: 'admin-token-yyyyy', type: 'kubernetes.io/service-account-token' }],
        automount_token: true,
        created_at: '2024-02-01T14:20:00Z',
        labels: { 'role': 'admin' }
      }
    ]
  } catch (error) {
    console.error('加载 ServiceAccount 失败:', error)
    Message.error({ content: '加载失败' })
  } finally {
    loading.value = false
  }
}

// 创建 ServiceAccount
const createServiceAccount = async () => {
  loading.value = true
  try {
    // TODO: 调用真实 API
    // await serviceAccountApi.create(saForm.value)
    Message.success({ content: '创建成功' })
    closeCreateModal()
    await loadServiceAccounts()
  } catch (error) {
    console.error('创建失败:', error)
    Message.error({ content: '创建失败' })
  } finally {
    loading.value = false
  }
}

// 删除 ServiceAccount
const deleteServiceAccount = async (sa) => {
  if (!confirm(`确认删除 ServiceAccount "${sa.name}"？`)) return

  loading.value = true
  try {
    // TODO: 调用真实 API
    // await serviceAccountApi.delete(sa.namespace, sa.name)
    Message.success({ content: '删除成功' })
    await loadServiceAccounts()
  } catch (error) {
    console.error('删除失败:', error)
    Message.error({ content: '删除失败' })
  } finally {
    loading.value = false
  }
}

// 查看详情
const viewDetails = (sa) => {
  currentSA.value = sa
  showDetailsModal.value = true
}

// 查看绑定
const viewBindings = (sa) => {
  Message.info({ content: `查看 ${sa.name} 的角色绑定（功能开发中）` })
}

// 编辑
const editServiceAccount = (sa) => {
  Message.info({ content: `编辑 ${sa.name}（功能开发中）` })
}

// 标签管理
const addLabel = () => {
  saForm.value.labels.push({ key: '', value: '' })
}

const removeLabel = (index) => {
  saForm.value.labels.splice(index, 1)
}

// 关闭模态框
const closeCreateModal = () => {
  showCreateModal.value = false
  saForm.value = {
    name: '',
    namespace: 'default',
    labels: [],
    autoMountToken: true
  }
}

const closeDetailsModal = () => {
  showDetailsModal.value = false
  currentSA.value = null
}

// 搜索输入防抖
let searchTimeout = null
const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    // 触发搜索
  }, 300)
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 自动刷新
let refreshInterval = null
onMounted(() => {
  loadServiceAccounts()
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>

<style scoped>
@import '@/styles/resource-common.css';

.detail-section {
  margin-bottom: 24px;
}

.detail-section h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: #333;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item label {
  font-size: 12px;
  color: #666;
  font-weight: 500;
}

.detail-item span {
  font-size: 14px;
  color: #333;
}

.detail-table {
  width: 100%;
  border-collapse: collapse;
}

.detail-table th,
.detail-table td {
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.detail-table th {
  background-color: #f5f5f5;
  font-weight: 600;
}

.labels-display {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  background-color: #e3f2fd;
  color: #1976d2;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
}

.label-input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.label-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.label-row input {
  flex: 1;
}

.namespace-tag {
  background-color: #f0f0f0;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.help-text {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.empty-hint {
  color: #999;
  font-style: italic;
}
</style>

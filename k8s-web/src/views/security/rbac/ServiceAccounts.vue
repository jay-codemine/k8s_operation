<template>
  <div class="rbac-view">
    <!-- 页面头部 -->
    <div class="rbac-header">
      <div class="rbac-header-left">
        <div class="rbac-icon">🔐</div>
        <div class="rbac-title-group">
          <h1>ServiceAccount 管理</h1>
          <p>管理 Kubernetes 服务账户，用于 Pod 身份认证和权限控制</p>
        </div>
      </div>
      <div class="rbac-header-right">
        <ClusterSelector
          v-model="selectedClusterId"
          :clusters="clusters"
          :show-all-option="false"
          label="集群"
          @change="onClusterChange"
        />
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="rbac-stats">
      <div class="stat-card primary">
        <div class="stat-header">
          <div class="stat-icon">👤</div>
        </div>
        <div class="stat-value">{{ serviceAccounts.length }}</div>
        <div class="stat-label">ServiceAccount 总数</div>
      </div>
      <div class="stat-card success">
        <div class="stat-header">
          <div class="stat-icon">🔑</div>
        </div>
        <div class="stat-value">{{ totalSecretsCount }}</div>
        <div class="stat-label">关联 Secrets</div>
      </div>
      <div class="stat-card info">
        <div class="stat-header">
          <div class="stat-icon">📁</div>
        </div>
        <div class="stat-value">{{ uniqueNamespaces }}</div>
        <div class="stat-label">涉及命名空间</div>
      </div>
      <div class="stat-card warning">
        <div class="stat-header">
          <div class="stat-icon">🔄</div>
        </div>
        <div class="stat-value">{{ autoMountCount }}</div>
        <div class="stat-label">自动挂载 Token</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="rbac-toolbar">
      <div class="toolbar-left">
        <div class="rbac-search">
          <span class="search-icon">🔍</span>
          <input v-model="searchQuery" placeholder="搜索 ServiceAccount 名称..." @input="onSearchInput" />
        </div>
        <div class="rbac-filter">
          <select v-model="namespaceFilter" @change="loadServiceAccounts">
            <option value="">所有命名空间</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" />
          <span>自动刷新</span>
          <span v-if="autoRefresh" class="refresh-indicator">●</span>
        </label>
      </div>
      <div class="toolbar-right">
        <button class="rbac-btn rbac-btn-secondary" @click="loadServiceAccounts" :disabled="loading">
          {{ loading ? '⏳' : '🔄' }} 刷新
        </button>
        <button class="rbac-btn rbac-btn-primary" @click="showCreateModal = true">
          + 创建 ServiceAccount
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && serviceAccounts.length === 0" class="rbac-loading">
      <div class="rbac-spinner"></div>
      <div class="rbac-loading-text">加载中...</div>
    </div>

    <!-- 表格 -->
    <div v-else class="rbac-table-container">
      <table class="rbac-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>命名空间</th>
            <th>Secrets</th>
            <th>自动挂载</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sa in filteredServiceAccounts" :key="`${sa.namespace}-${sa.name}`">
            <td>
              <div class="resource-name">
                <div class="icon">👤</div>
                <div>
                  <div class="name">{{ sa.name }}</div>
                  <div class="meta">{{ sa.secrets?.length || 0 }} Secrets</div>
                </div>
              </div>
            </td>
            <td>
              <span class="namespace-tag">{{ sa.namespace }}</span>
            </td>
            <td>
              <span class="rbac-tag rbac-tag-info">{{ sa.secrets?.length || 0 }}</span>
            </td>
            <td>
              <span class="rbac-tag" :class="sa.automount_token ? 'rbac-tag-success' : 'rbac-tag-default'">
                {{ sa.automount_token ? '是' : '否' }}
              </span>
            </td>
            <td>{{ formatDate(sa.created_at) }}</td>
            <td>
              <div class="rbac-actions">
                <button class="rbac-action-btn view" @click="viewDetails(sa)" title="查看详情">👁️</button>
                <button class="rbac-action-btn" @click="viewBindings(sa)" title="查看绑定">🔗</button>
                <button class="rbac-action-btn edit" @click="editServiceAccount(sa)" title="编辑">✏️</button>
                <button class="rbac-action-btn delete" @click="deleteServiceAccount(sa)" title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 空状态 -->
      <div v-if="filteredServiceAccounts.length === 0" class="rbac-empty">
        <div class="rbac-empty-icon">👤</div>
        <div class="rbac-empty-title">暂无 ServiceAccount</div>
        <div class="rbac-empty-desc">{{ searchQuery ? '没有匹配的结果' : '点击上方按钮创建' }}</div>
        <button v-if="!searchQuery" class="rbac-btn rbac-btn-primary" @click="showCreateModal = true">
          + 创建 ServiceAccount
        </button>
      </div>
    </div>

    <!-- 创建 ServiceAccount 模态框 -->
    <div v-if="showCreateModal" class="rbac-modal" @click.self="closeCreateModal">
      <div class="rbac-modal-content">
        <div class="rbac-modal-header">
          <h2>创建 ServiceAccount</h2>
          <button class="rbac-modal-close" @click="closeCreateModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <form @submit.prevent="createServiceAccount">
            <div class="rbac-form-group">
              <label>名称 *</label>
              <input v-model="saForm.name" placeholder="例如：my-service-account" required />
            </div>
            <div class="rbac-form-group">
              <label>命名空间 *</label>
              <select v-model="saForm.namespace" required>
                <option value="">请选择命名空间</option>
                <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
              </select>
            </div>
            <div class="rbac-form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="saForm.autoMountToken" />
                <span>自动挂载 ServiceAccount Token</span>
              </label>
              <span class="help-text">选中后，使用此 ServiceAccount 的 Pod 会自动挂载 Token</span>
            </div>
          </form>
        </div>
        <div class="rbac-modal-footer">
          <button class="rbac-btn rbac-btn-secondary" @click="closeCreateModal">取消</button>
          <button class="rbac-btn rbac-btn-primary" @click="createServiceAccount" :disabled="loading">
            {{ loading ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 详情模态框 -->
    <div v-if="showDetailsModal" class="rbac-modal" @click.self="closeDetailsModal">
      <div class="rbac-modal-content large">
        <div class="rbac-modal-header">
          <h2>ServiceAccount 详情</h2>
          <button class="rbac-modal-close" @click="closeDetailsModal">&times;</button>
        </div>
        <div class="rbac-modal-body">
          <div class="detail-section">
            <h3>基本信息</h3>
            <div class="detail-grid">
              <div class="detail-item">
                <label>名称</label>
                <span>{{ currentSA?.name }}</span>
              </div>
              <div class="detail-item">
                <label>命名空间</label>
                <span class="namespace-tag">{{ currentSA?.namespace }}</span>
              </div>
              <div class="detail-item">
                <label>创建时间</label>
                <span>{{ formatDate(currentSA?.created_at) }}</span>
              </div>
              <div class="detail-item">
                <label>自动挂载 Token</label>
                <span class="rbac-tag" :class="currentSA?.automount_token ? 'rbac-tag-success' : 'rbac-tag-default'">
                  {{ currentSA?.automount_token ? '是' : '否' }}
                </span>
              </div>
            </div>
          </div>

          <div class="detail-section">
            <h3>关联的 Secrets ({{ currentSA?.secrets?.length || 0 }})</h3>
            <div v-if="currentSA?.secrets?.length" class="secrets-list">
              <div v-for="secret in currentSA.secrets" :key="secret.name" class="secret-item">
                <span class="secret-icon">🔑</span>
                <span class="secret-name">{{ secret.name }}</span>
              </div>
            </div>
            <div v-else class="empty-secrets">暂无关联的 Secrets</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import ClusterSelector from '@/components/cluster/ClusterSelector.vue'
import { listServiceAccounts, createServiceAccount as createSAApi, deleteServiceAccount as deleteSAApi } from '@/api/k8sRbac'
import { getClusterList } from '@/api/cluster'
import { getNamespaces } from '@/api/namespace'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

const { confirm: showConfirm } = useConfirmDialog()

// 数据状态
const loading = ref(false)
const searchQuery = ref('')
const namespaceFilter = ref('')
const serviceAccounts = ref([])
const namespaces = ref([])
const autoRefresh = ref(false)
let refreshTimer = null

// 集群选择
const clusters = ref([])
const selectedClusterId = ref('')

// 统计数据
const totalSecretsCount = computed(() => serviceAccounts.value.reduce((sum, sa) => sum + (sa.secrets?.length || 0), 0))
const uniqueNamespaces = computed(() => new Set(serviceAccounts.value.map(sa => sa.namespace)).size)
const autoMountCount = computed(() => serviceAccounts.value.filter(sa => sa.automount_token).length)

// 监听集群变化
const onClusterChange = (clusterId) => {
  if (clusterId) {
    loadNamespaces()
    loadServiceAccounts()
  }
}

watch(selectedClusterId, (val) => {
  if (val) {
    loadNamespaces()
    loadServiceAccounts()
  }
})

// 自动刷新
watch(autoRefresh, (val) => {
  if (val) {
    refreshTimer = setInterval(loadServiceAccounts, 30000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

// 加载集群列表
const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 100 })
    if (res.code === 0 && res.data?.list) {
      // 权限过滤：只显示用户有权限访问的集群
      const allClusters = res.data.list
      clusters.value = allClusters
        .filter(c => permissionStore.state.isSuperAdmin || permissionStore.state.accessibleClusterIds.includes(c.id))
        .map(c => ({ ...c, name: c.cluster_name || c.name }))
      if (clusters.value.length > 0 && !selectedClusterId.value) {
        selectedClusterId.value = clusters.value[0].id
      }
    }
  } catch (error) {
    console.error('加载集群列表失败:', error)
  }
}

// 加载命名空间列表
const loadNamespaces = async () => {
  if (!selectedClusterId.value) return
  try {
    const res = await getNamespaces(selectedClusterId.value)
    if (res.code === 0 && res.data?.list) {
      namespaces.value = res.data.list.map(ns => ns.name || ns)
    }
  } catch (error) {
    namespaces.value = ['default', 'kube-system', 'kube-public']
  }
}

// 模态框状态
const showCreateModal = ref(false)
const showDetailsModal = ref(false)
const currentSA = ref(null)

// 表单数据
const saForm = ref({ name: '', namespace: '', autoMountToken: true })

// 过滤后的 ServiceAccount
const filteredServiceAccounts = computed(() => {
  let result = serviceAccounts.value
  if (namespaceFilter.value) result = result.filter(sa => sa.namespace === namespaceFilter.value)
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(sa => sa.name.toLowerCase().includes(query))
  }
  return result
})

// 加载 ServiceAccount
const loadServiceAccounts = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await listServiceAccounts(selectedClusterId.value, namespaceFilter.value)
    serviceAccounts.value = res.code === 0 && res.data?.list ? res.data.list : []
  } catch (error) {
    console.error('加载 ServiceAccount 失败:', error)
    Message.error({ content: '加载失败: ' + (error.msg || error.message || '网络错误') })
    serviceAccounts.value = []
  } finally {
    loading.value = false
  }
}

// 创建 ServiceAccount
const createServiceAccount = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    await createSAApi(selectedClusterId.value, {
      name: saForm.value.name,
      namespace: saForm.value.namespace,
      automount_token: saForm.value.autoMountToken
    })
    Message.success({ content: '创建成功' })
    closeCreateModal()
    await loadServiceAccounts()
  } catch (error) {
    Message.error({ content: '创建失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 删除 ServiceAccount
const deleteServiceAccount = async (sa) => {
  const ok = await showConfirm({
    title: '确认删除 ServiceAccount',
    type: 'danger',
    details: [
      { label: '名称', value: sa.name, mono: true },
    ],
    confirmText: '确认删除',
    cancelText: '取消',
  })
  if (!ok) return
  loading.value = true
  try {
    await deleteSAApi(selectedClusterId.value, sa.namespace, sa.name)
    Message.success({ content: '删除成功' })
    await loadServiceAccounts()
  } catch (error) {
    Message.error({ content: '删除失败: ' + (error.msg || error.message) })
  } finally {
    loading.value = false
  }
}

// 查看详情
const viewDetails = (sa) => { currentSA.value = sa; showDetailsModal.value = true }
const viewBindings = (sa) => { Message.info({ content: `查看 ${sa.name} 的绑定（功能开发中）` }) }
const editServiceAccount = (sa) => { Message.info({ content: `编辑 ${sa.name}（功能开发中）` }) }
const onSearchInput = () => {}

const closeCreateModal = () => {
  showCreateModal.value = false
  saForm.value = { name: '', namespace: '', autoMountToken: true }
}
const closeDetailsModal = () => { showDetailsModal.value = false; currentSA.value = null }
const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

onMounted(() => { loadClusters() })
onUnmounted(() => { if (refreshTimer) clearInterval(refreshTimer) })
</script>

<style scoped>
@import '@/assets/styles/rbac-common.css';

/* 自动刷新切换 */
.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  color: #4b5563;
  font-size: 13px;
  cursor: pointer;
}

.auto-refresh-toggle input {
  accent-color: #6366f1;
}

.refresh-indicator {
  color: #10b981;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* 详情区块 */
.detail-section {
  margin-bottom: 24px;
}

.detail-section h3 {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin: 0 0 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e5e7eb;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-item label {
  font-size: 11px;
  font-weight: 500;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-item span {
  font-size: 14px;
  color: #1f2937;
}

/* Secrets 列表 */
.secrets-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.secret-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}

.secret-icon {
  font-size: 16px;
}

.secret-name {
  font-family: 'Fira Code', monospace;
  font-size: 13px;
  color: #6366f1;
}

.empty-secrets {
  padding: 20px;
  text-align: center;
  color: #6b7280;
  font-size: 13px;
}

/* 复选框标签 */
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #374151;
  font-size: 14px;
}

.checkbox-label input {
  accent-color: #6366f1;
}

.help-text {
  font-size: 11px;
  color: #6b7280;
  margin-top: 6px;
}
</style>

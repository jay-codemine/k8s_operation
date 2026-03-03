<template>
  <div class="cluster-view">
    <div class="view-header">
      <h1>集群管理</h1>
      <p>基础 CRUD + kubeconfig 上传/更新 + 连接测试</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="按集群名称搜索..."
          @keyup.enter="fetchList"
        />
      </div>

      <div class="action-buttons">
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
        
        <div class="filter-buttons">
          <button class="btn btn-filter" :class="{ active: statusFilter === 'all' }"
                  @click="setFilter('all')">
            全部
          </button>
          <button class="btn btn-filter" :class="{ active: statusFilter === 'ok' }"
                  @click="setFilter('ok')">
            正常
          </button>
          <button class="btn btn-filter" :class="{ active: statusFilter === 'bad' }"
                  @click="setFilter('bad')">
            异常
          </button>
          <button class="btn btn-filter" :class="{ active: statusFilter === 'pending' }"
                  @click="setFilter('pending')">
            待检测
          </button>
        </div>

        <div class="filter-dropdown">
          <select v-model.number="itemsPerPage">
            <option :value="10">10 条/页</option>
            <option :value="20">20 条/页</option>
            <option :value="50">50 条/页</option>
            <option :value="100">100 条/页</option>
          </select>
        </div>

        <div class="filter-dropdown">
          <input
            type="number"
            min="1"
            v-model.number="currentPage"
            @keyup.enter="fetchList"
            class="page-input"
            placeholder="页码"
          />
        </div>

        <button class="btn btn-primary" @click="openCreate">➕ 创建集群</button>
        <button class="btn btn-secondary" :disabled="loading" @click="fetchList">🔄 刷新</button>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-container">
      <div class="table-scroll">
        <table class="resource-table">
          <thead>
          <tr>
            <th style="width: 80px;">ID</th>
            <th>集群名称</th>
            <th style="width: 140px;">版本</th>
            <th style="width: 140px;">状态</th>
            <th style="width: 160px;">最近检测</th>
            <th style="width: 260px;">操作</th>
          </tr>
          </thead>

          <tbody>
          <tr v-for="c in paginatedClusters" :key="c.id">
            <td>{{ c.id }}</td>

            <td>
              <a href="javascript:void(0)" class="cluster-link" @click.prevent="enterCluster(c)">
                {{ c.cluster_name }}
              </a>
              <div v-if="c.last_error" class="row-sub muted" :title="c.last_error">
                {{ c.last_error }}
              </div>
            </td>

            <td>{{ c.cluster_version || '-' }}</td>

            <td>
                <span class="status-indicator" :class="statusClass(c.status)">
                  {{ statusText(c.status) }}
                </span>
            </td>

            <td>
              <span class="muted">{{ formatCheckAt(c.last_check_at) }}</span>
            </td>

            <td>
              <div class="op">
                <button class="btn btn-mini" :disabled="testingId === c.id || loading"
                        @click="openEdit(c)">
                  编辑
                </button>

                <button
                  class="btn btn-mini btn-info"
                  :disabled="testingId === c.id || loading"
                  @click="testCluster(c)"
                >
                  {{ testingId === c.id ? '检测中...' : '健康检查' }}
                </button>

                <button
                  class="btn btn-mini btn-danger"
                  :disabled="testingId === c.id || loading"
                  @click="onDelete(c)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
          </tbody>
        </table>
      </div>

      <div v-if="filteredClusters.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">
          {{ searchQuery ? '没有匹配结果' : '暂无集群，请先创建' }}
        </div>
      </div>

      <Pagination
        v-if="filteredClusters.length > 0"
        v-model:currentPage="currentPage"
        :totalItems="filteredClusters.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'card'" class="cards-container">
      <div v-if="filteredClusters.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <div class="empty-text">
          {{ searchQuery ? '没有匹配结果' : '暂无集群，请先创建' }}
        </div>
      </div>
      
      <div class="cards-grid">
        <div v-for="c in paginatedClusters" :key="c.id" class="cluster-card">
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="card-title-row">
              <span class="card-icon">☸️</span>
              <h3 class="card-title">
                <a href="javascript:void(0)" class="cluster-link" @click.prevent="enterCluster(c)">
                  {{ c.cluster_name }}
                </a>
              </h3>
              <span class="status-indicator" :class="statusClass(c.status)">
                {{ statusText(c.status) }}
              </span>
            </div>
            <div class="card-id">集群 ID: {{ c.id }}</div>
          </div>

          <!-- 卡片主体 -->
          <div class="card-body">
            <!-- K8s 版本 -->
            <div class="card-section">
              <div class="section-label">Kubernetes 版本</div>
              <div class="meta-value">{{ c.cluster_version || '-' }}</div>
            </div>

            <!-- 最近检测 -->
            <div class="card-section">
              <div class="section-label">最近健康检查</div>
              <div class="meta-value">{{ formatCheckAt(c.last_check_at) }}</div>
            </div>

            <!-- 错误信息（如果有） -->
            <div v-if="c.last_error" class="card-section">
              <div class="section-label">错误信息</div>
              <div class="error-text">{{ c.last_error }}</div>
            </div>
          </div>

          <!-- 卡片底部按钮 -->
          <div class="card-footer">
            <button 
              class="card-action-btn" 
              :disabled="testingId === c.id || loading"
              @click="enterCluster(c)"
              title="进入集群"
            >
              🔗 进入
            </button>
            <button 
              class="card-action-btn" 
              :disabled="testingId === c.id || loading"
              @click="openEdit(c)"
              title="编辑"
            >
              ✏️ 编辑
            </button>
            <button 
              class="card-action-btn" 
              :disabled="testingId === c.id || loading"
              @click="testCluster(c)"
              title="健康检查"
            >
              {{ testingId === c.id ? '⏳ 检测中' : '🔍 检查' }}
            </button>
            <button 
              class="card-action-btn danger" 
              :disabled="testingId === c.id || loading"
              @click="onDelete(c)"
              title="删除"
            >
              🗑️ 删除
            </button>
          </div>
        </div>
      </div>
      
      <Pagination
        v-if="filteredClusters.length > 0"
        v-model:currentPage="currentPage"
        :totalItems="filteredClusters.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 创建/编辑 弹窗 -->
    <div class="modal" v-if="showFormModal">
      <div class="modal-backdrop" @click="closeForm"></div>

      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ formMode === 'create' ? '创建集群' : '编辑集群' }}</h2>
          <button class="close-btn" @click="closeForm">&times;</button>
        </div>

        <div class="modal-body">
          <form class="form" @submit.prevent="submitForm">
            <div class="topbar" v-if="formMode === 'edit'">
              <div class="chip">ClusterID: {{ form.id }}</div>
              <div class="muted">编辑模式：不填写 kubeconfig 不会覆盖</div>
            </div>

            <div class="card">
              <div class="card-title">基本信息</div>

              <div class="grid">
                <div class="field" v-if="formMode === 'edit'">
                  <label>ID</label>
                  <input type="number" v-model="form.id" disabled/>
                </div>

                <div class="field">
                  <label>集群名称 <span class="required">*</span></label>
                  <input
                    type="text"
                    v-model="form.cluster_name"
                    placeholder="例如：测试环境 k8s 集群"
                    required
                  />
                </div>

                <div class="field">
                  <label>K8s 版本 <span class="required">*</span></label>
                  <input
                    type="text"
                    v-model="form.cluster_version"
                    placeholder="如 v1.28.3"
                    required
                  />
                </div>
              </div>
            </div>

            <div class="card">
              <div class="card-title">
                kubeconfig（高级）
                <span class="hint">创建模式必填；编辑模式可选，不填则不更新</span>
              </div>

              <div class="upload-row">
                <input
                  class="file"
                  type="file"
                  accept=".yaml,.yml,.conf,.json,.txt"
                  @change="onKubeconfigFileChange"
                />
                <button type="button" class="btn small ghost" @click="clearKubeconfigText">
                  清空文本
                </button>
              </div>

              <div class="alert">
                <span class="alert-icon">⚠️</span>
                <div class="alert-text">
                  上传 kubeconfig 文件后会自动填入下方文本框；编辑时不填 kubeconfig 则不会更新。
                </div>
              </div>

              <textarea
                v-model="form.kube_config"
                class="codebox"
                rows="10"
                placeholder="粘贴 kubeconfig 原文（YAML/JSON），或使用上方上传文件"
                :required="formMode === 'create'"
              ></textarea>
            </div>

            <div class="footer">
              <button type="button" class="btn ghost" @click="closeForm">取消</button>
              <button type="submit" class="btn primary" :disabled="loading">
                {{ loading ? '提交中...' : '提交保存' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue'
import {Message} from '@arco-design/web-vue'
import {useRouter} from 'vue-router'
import Pagination from '@/components/Pagination.vue'
import {useClusterStore} from '@/stores/cluster'

import {
  createCluster,
  deleteCluster,
  getClusterList,
  initCluster,
  updateCluster,
} from '@/api/cluster'

const router = useRouter()
const clusterStore = useClusterStore()

// ===== 列表数据 =====
const clusters = ref([])

// 视图模式：table（表格） 或 card（卡片）
const viewMode = ref('table')

// ===== UI =====
const searchQuery = ref('')
const statusFilter = ref('all') // all | ok | bad | pending
const currentPage = ref(1)
const itemsPerPage = ref(10)
const loading = ref(false)
const testingId = ref(null)

// ===== 弹窗表单 =====
const showFormModal = ref(false)
const formMode = ref('create') // create | edit
const form = ref({
  id: 0,
  cluster_name: '',
  cluster_version: '',
  kube_config: '',
})

onMounted(() => {
  clusterStore.hydrate?.()
  fetchList()
})

const setFilter = (v) => {
  statusFilter.value = v
  currentPage.value = 1
}

const enterCluster = async (c) => {
  clusterStore.setCurrent(c)
  router.push(`/c/${Number(c.id)}/nodes`)
}

/**
 * 你后端三态：0=OK 1=Bad 2=Pending
 * 如果你后端枚举不一样，把这里改一下就行
 */
const statusText = (s) => {
  const n = Number(s)
  if (n === 0) return '正常'
  if (n === 1) return '异常'
  if (n === 2) return '待检测'
  return '未知'
}
const statusClass = (s) => {
  const n = Number(s)
  if (n === 0) return 'connected'
  if (n === 1) return 'disconnected'
  if (n === 2) return 'pending'
  return 'unknown'
}

const pickMsg = (body, fallback = '') => {
  if (!body) return fallback
  if (body?.data?.message) return body.data.message
  if (body?.msg) return body.msg
  if (body?.message) return body.message
  if (Array.isArray(body?.details) && body.details.length > 0) return body.details.join('；')
  return fallback
}
const unwrapErrorBody = (e) => {
  if (e?.response?.data) return e.response.data
  if (e?.code || e?.msg || e?.message) return e
  return null
}
const isOk = (body) => Number(body?.code) === 0

// ✅ 拉取列表：只信后端（status/last_check_at/last_error 都由后端写库）
const fetchList = async () => {
  loading.value = true
  try {
    const body = await getClusterList({
      cluster_name: searchQuery.value || '',
      page: 1,
      limit: 1000,
    })

    const list = body?.data?.list || body?.data?.items || body?.list || body?.items || []
    clusters.value = Array.isArray(list) ? list.map((x) => ({
      ...x,
      id: Number(x?.id),
      status: Number(x?.status),
      last_check_at: Number(x?.last_check_at || 0),
      last_error: String(x?.last_error || ''),
    })) : []

    // 如果当前页超范围，拉回最后一页/第一页
    const totalPages = Math.max(1, Math.ceil(filteredClusters.value.length / itemsPerPage.value))
    if (currentPage.value > totalPages) currentPage.value = totalPages
    if (currentPage.value < 1) currentPage.value = 1
  } catch (e) {
    console.error(e)
    clusters.value = []
    Message.error({content: '拉取集群列表失败', duration: 2200})
  } finally {
    loading.value = false
  }
}

const filteredClusters = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()

  return clusters.value.filter((c) => {
    const hitName = !q || String(c.cluster_name || '').toLowerCase().includes(q)

    const s = Number(c.status)
    const hitStatus =
      statusFilter.value === 'all' ||
      (statusFilter.value === 'ok' && s === 0) ||
      (statusFilter.value === 'bad' && s === 1) ||
      (statusFilter.value === 'pending' && s === 2)

    return hitName && hitStatus
  })
})

const paginatedClusters = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredClusters.value.slice(start, end)
})

const openCreate = () => {
  formMode.value = 'create'
  form.value = {id: 0, cluster_name: '', cluster_version: '', kube_config: ''}
  showFormModal.value = true
}

const openEdit = (c) => {
  formMode.value = 'edit'
  form.value = {
    id: Number(c.id),
    cluster_name: c.cluster_name,
    cluster_version: c.cluster_version,
    kube_config: '',
  }
  showFormModal.value = true
}

const closeForm = () => {
  showFormModal.value = false
}

// kubeconfig 文件上传
const onKubeconfigFileChange = (evt) => {
  const file = evt?.target?.files?.[0]
  if (!file) return

  if (file.size > 1024 * 1024) {
    Message.warning({content: '文件过大（>1MB），请确认是否为 kubeconfig 文件', duration: 2200})
    evt.target.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    form.value.kube_config = String(reader.result || '')
    Message.success({content: 'kubeconfig 已读取到文本框', duration: 1600})
  }
  reader.onerror = () => {
    Message.error({content: '读取文件失败，请重试', duration: 2200})
  }
  reader.readAsText(file)
  evt.target.value = ''
}

const clearKubeconfigText = () => {
  form.value.kube_config = ''
}

// ✅ 创建/更新（更新 kubeconfig 后，后端应该把 status 置为 Pending，等 init 再写库）
const submitForm = async () => {
  loading.value = true
  try {
    if (formMode.value === 'create') {
      const body = await createCluster({
        cluster_name: form.value.cluster_name,
        cluster_version: form.value.cluster_version,
        kube_config: form.value.kube_config,
      })
      Message.success({content: body?.msg || '创建成功', duration: 1800})
      closeForm()
      await fetchList()
      return
    }

    // 编辑模式 - 检查是否更新 kubeconfig
    const hasKubeconfigUpdate = form.value.kube_config && form.value.kube_config.trim()
    
    // 二次确认 - 更新 kubeconfig 是高危操作
    if (hasKubeconfigUpdate) {
      if (!confirm(`⚠️ 确认更新集群配置？

集群名称: ${form.value.cluster_name}
集群 ID: ${form.value.id}
版本: ${form.value.cluster_version}

您正在更新 kubeconfig 配置！
此操作将替换现有的集群连接凭证，可能导致连接失败。
建议更新后立即执行健康检查。

确认继续？`)) {
        loading.value = false
        return
      }
    }

    const payload = {
      id: form.value.id,
      cluster_name: form.value.cluster_name,
      cluster_version: form.value.cluster_version,
    }
    if (hasKubeconfigUpdate) {
      payload.kube_config = form.value.kube_config
    }

    const body = await updateCluster(payload)
    Message.success({
      content: hasKubeconfigUpdate 
        ? (body?.msg || '更新成功，建议立即执行健康检查') 
        : (body?.msg || '更新成功'),
      duration: hasKubeconfigUpdate ? 2600 : 1800
    })
    closeForm()
    await fetchList()
  } catch (e) {
    console.error(e)
    const body = unwrapErrorBody(e)
    Message.error({
      content: `提交失败：${pickMsg(body, e?.message || '请检查参数')}`,
      duration: 2600,
    })
  } finally {
    loading.value = false
  }
}

const onDelete = async (c) => {
  // 二次确认 - 删除集群是高危操作
  if (!confirm(`⚠️ 确认删除集群？

集群名称: ${c.cluster_name}
集群 ID: ${c.id}
版本: ${c.cluster_version || '-'}
状态: ${statusText(c.status)}

警告：删除集群将移除所有关联配置和 kubeconfig！
此操作不可逆，请确认！`)) return;
  
  loading.value = true
  try {
    const body = await deleteCluster({id: c.id})
    Message.success({content: body?.msg || '删除成功', duration: 1800})
    await fetchList()
  } catch (e) {
    console.error(e)
    const body = unwrapErrorBody(e)
    Message.error({
      content: `删除失败：${pickMsg(body, e?.message || '请查看后端日志')}`,
      duration: 2600,
    })
  } finally {
    loading.value = false
  }
}

// ✅ 健康检查：只负责触发后端 init；状态以“后端写库 + list 返回”为准
const testCluster = async (c) => {
  testingId.value = c.id
  try {
    const body = await initCluster({id: c.id})
    const ok = isOk(body)
    Message[ok ? 'success' : 'error']({
      content: ok ? (body?.msg || '初始化成功') : (body?.msg || '初始化失败'),
      duration: ok ? 1800 : 2600,
    })
  } catch (e) {
    const body = unwrapErrorBody(e)
    Message.error({
      content: `初始化失败：${pickMsg(body, e?.message || 'K8s 集群初始化失败')}`,
      duration: 2600,
    })
  } finally {
    testingId.value = null
    // 关键：让后端写库后的 status 回显
    await fetchList()
  }
}

const testInModal = async () => {
  if (formMode.value !== 'edit' || !form.value.id) {
    Message.warning({content: '请先创建集群后再检查', duration: 1600})
    return
  }
  testingId.value = form.value.id
  try {
    const body = await initCluster({id: form.value.id})
    const ok = isOk(body)
    Message[ok ? 'success' : 'error']({
      content: ok ? (body?.msg || '初始化成功') : (body?.msg || '初始化失败'),
      duration: ok ? 1800 : 2600,
    })
  } catch (e) {
    const body = unwrapErrorBody(e)
    Message.error({
      content: `初始化失败：${pickMsg(body, e?.message || 'K8s 集群初始化失败')}`,
      duration: 2600,
    })
  } finally {
    testingId.value = null
    await fetchList()
  }
}

const formatCheckAt = (ts) => {
  const n = Number(ts || 0)
  if (!n) return '-'
  const d = new Date(n * 1000)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}
</script>

<style scoped>
/* 容器：支持全屏自适应 */
.cluster-view {
  width: 100%;
  max-width: 1800px;
  margin: 0 auto;
  padding: 20px 24px;
  min-height: calc(100vh - 60px);
  box-sizing: border-box;
}

.view-header {
  margin-bottom: 20px;
}

.view-header h1 {
  font-size: clamp(22px, 3vw, 28px);
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1e293b;
}

.view-header p {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

/* 操作栏：弹性布局自适应 */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.search-box {
  flex: 1;
  min-width: 200px;
  max-width: 400px;
}

.search-box input {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  outline: none;
  font-size: 14px;
  transition: all 0.2s ease;
}

.search-box input:focus {
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-buttons {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.filter-dropdown select,
.filter-dropdown input {
  padding: 9px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-dropdown select:focus,
.filter-dropdown input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.12);
}

.page-input {
  width: 72px;
}

/* 按钮通用样式 */
.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #326ce5 0%, #2557c5 100%);
  color: #fff;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-filter {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid transparent;
  padding: 8px 14px;
  border-radius: 20px;
  font-size: 13px;
}

.btn-filter:hover {
  background: #e2e8f0;
}

.btn-filter.active {
  background: #326ce5;
  color: #fff;
}

.btn-mini {
  padding: 6px 12px;
  border-radius: 8px;
  background: #f1f5f9;
  font-size: 13px;
}

.btn-mini:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.btn-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.btn-info {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

.btn-info:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.2);
}

/* 表格容器：弹性高度 */
.table-container {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.table-scroll {
  flex: 1;
  max-height: calc(100vh - 320px);
  min-height: 300px;
  overflow: auto;
  -webkit-overflow-scrolling: touch;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 800px;
}

.resource-table th {
  background: #f8fafc;
  text-align: left;
  padding: 14px 16px;
  border-bottom: 1px solid #e2e8f0;
  font-weight: 600;
  font-size: 13px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.resource-table td {
  padding: 16px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
  font-size: 14px;
}

.resource-table tbody tr {
  transition: background-color 0.15s ease;
}

.resource-table tbody tr:hover {
  background-color: #f8fafc;
}

.op {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 状态标签 */
.status-indicator {
  display: inline-flex;
  align-items: center;
  padding: 5px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-indicator.connected {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.status-indicator.disconnected {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.status-indicator.pending {
  background: rgba(245, 158, 11, 0.1);
  color: #d97706;
}

.status-indicator.unknown {
  background: rgba(148, 163, 184, 0.15);
  color: #64748b;
}

.cluster-link {
  color: #326ce5;
  cursor: pointer;
  font-weight: 600;
  text-decoration: none;
  transition: color 0.15s ease;
}

.cluster-link:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.row-sub {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.4;
  max-width: 400px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.muted {
  color: #64748b;
}

/* 空状态 */
.empty-state {
  padding: 80px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #94a3b8;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 15px;
}

/* ===== Modal：响应式弹窗 ===== */
.modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 20px;
  overflow-y: auto;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
}

.modal-content {
  position: relative;
  width: 100%;
  max-width: 720px;
  max-height: calc(100vh - 80px);
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
}

.close-btn {
  border: none;
  background: transparent;
  font-size: 24px;
  cursor: pointer;
  line-height: 1;
  color: #94a3b8;
  padding: 4px;
  border-radius: 8px;
  transition: all 0.15s ease;
}

.close-btn:hover {
  color: #1e293b;
  background: #e2e8f0;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

/* ===== Modal Form ===== */
.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px 16px;
}

.chip {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
  background: rgba(50, 108, 229, 0.1);
  border-radius: 20px;
  padding: 6px 12px;
}

.card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 16px;
}

.card-title {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 12px;
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 15px;
}

.card-title .hint {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 400;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.field label {
  display: block;
  font-size: 13px;
  color: #475569;
  margin-bottom: 6px;
  font-weight: 500;
}

.field input {
  width: 100%;
  height: 40px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  padding: 0 14px;
  outline: none;
  font-size: 14px;
  transition: all 0.15s ease;
  box-sizing: border-box;
}

.field input:focus {
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.field input:disabled {
  background: #f8fafc;
  color: #64748b;
}

.required {
  color: #ef4444;
}

.upload-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.file {
  flex: 1;
  min-width: 200px;
}

.alert {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 12px;
}

.alert-icon {
  font-size: 16px;
  line-height: 1.4;
}

.alert-text {
  font-size: 13px;
  color: #92400e;
  line-height: 1.5;
}

.codebox {
  width: 100%;
  min-height: 200px;
  resize: vertical;
  border-radius: 10px;
  border: 1px solid #1e293b;
  padding: 14px;
  background: #0f172a;
  color: #e2e8f0;
  font-family: 'SF Mono', 'Menlo', 'Monaco', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
  outline: none;
  box-sizing: border-box;
}

.codebox:focus {
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.15);
}

.footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
}

.btn.primary {
  background: linear-gradient(135deg, #326ce5 0%, #2557c5 100%);
  color: #fff;
}

.btn.ghost {
  background: #f1f5f9;
  color: #475569;
}

.btn.ghost:hover {
  background: #e2e8f0;
}

.btn.info {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

.btn.small {
  padding: 8px 14px;
  border-radius: 8px;
}

.test-tip {
  font-size: 12px;
  color: #64748b;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .cluster-view {
    padding: 16px;
  }
  
  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    max-width: none;
  }
  
  .action-buttons {
    justify-content: flex-start;
  }
  
  .table-scroll {
    max-height: calc(100vh - 380px);
  }
  
  .modal {
    padding: 20px 12px;
  }
  
  .modal-content {
    max-height: calc(100vh - 40px);
  }
}

/* ==================== */
/* 视图切换按钮样式 */
/* ==================== */
.view-toggle {
  display: flex;
  gap: 0;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #d1d5db;
}

.btn-view {
  padding: 8px 16px;
  background: white;
  border: none;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
  border-right: 1px solid #d1d5db;
}

.btn-view:last-child {
  border-right: none;
}

.btn-view:hover {
  background: #f3f4f6;
}

.btn-view.active {
  background: #3b82f6;
  color: white;
}

/* ==================== */
/* 卡片视图样式 */
/* ==================== */
.cards-container {
  padding: 0;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.cluster-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.cluster-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

/* 卡片头部 */
.card-header {
  padding: 16px;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  border-bottom: 1px solid #e2e8f0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.card-icon {
  font-size: 28px;
}

.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  flex: 1;
}

.card-title .cluster-link {
  color: #3b82f6;
  text-decoration: none;
  transition: color 0.2s;
}

.card-title .cluster-link:hover {
  color: #2563eb;
  text-decoration: underline;
}

.card-id {
  font-size: 12px;
  color: #64748b;
  font-family: monospace;
}

/* 卡片主体 */
.card-body {
  padding: 16px;
}

.card-section {
  margin-bottom: 14px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-label {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.meta-value {
  font-size: 14px;
  color: #1e293b;
  word-break: break-all;
}

.error-text {
  font-size: 13px;
  color: #dc2626;
  background: #fef2f2;
  padding: 8px;
  border-radius: 6px;
  word-break: break-all;
}

/* 卡片底部按钮 */
.card-footer {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  flex-wrap: wrap;
}

.card-action-btn {
  flex: 1;
  min-width: 80px;
  padding: 8px 12px;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.card-action-btn:hover:not(:disabled) {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.card-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.card-action-btn.danger {
  color: #dc2626;
}

.card-action-btn.danger:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #fca5a5;
}

/* 响应式 - 小屏幕单列 */
@media (max-width: 768px) {
  .cards-grid {
    grid-template-columns: 1fr;
  }
  
  .card-footer {
    flex-direction: column;
  }
  
  .card-action-btn {
    width: 100%;
  }
}

</style>

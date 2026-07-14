<template>
  <div class="vpa-mgmt">
    <div class="view-header">
      <h1>VPA 弹性扩缩容（垂直扩展）</h1>
      <p>基于历史 CPU / 内存使用情况自动调整容器的资源 Requests / Limits。需先在集群安装 vertical-pod-autoscaler。</p>
    </div>

    <!-- VPA 不可用提示 -->
    <div v-if="!checking && available === false" class="warn-box">
      ⚠️ 当前集群未检测到 VPA Operator。请先在集群中部署
      <a href="https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler" target="_blank">
        vertical-pod-autoscaler
      </a>
      后再使用此功能。
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <input class="search-input" v-model="filter.namespace" placeholder="命名空间过滤" @keyup.enter="loadList" />
      <input class="search-input" v-model="filter.name" placeholder="VPA 名称（模糊）" @keyup.enter="loadList" />
      <button class="btn primary" @click="loadList" :disabled="!available">🔍 查询</button>
      <button class="btn" @click="onRefresh" :disabled="loading">🔄 刷新</button>
      <button class="btn primary" @click="openCreate" :disabled="!available">➕ 创建 VPA</button>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 数据统计 -->
    <div class="stats-bar" v-if="list.length">
      <span>总计: <b>{{ total }}</b></span>
      <span>Active: <b class="green">{{ getStatusCount('Active') }}</b></span>
      <span>Pending: <b class="yellow">{{ getStatusCount('Pending') }}</b></span>
    </div>

    <!-- 列表 -->
    <table class="vpa-table">
      <thead>
        <tr>
          <th>状态</th>
          <th>名称 / 命名空间</th>
          <th>目标</th>
          <th>更新模式</th>
          <th>容器</th>
          <th>资源限制范围</th>
          <th>推荐资源</th>
          <th>创建时间</th>
          <th style="width: 220px">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="9" style="text-align:center;padding:50px 0"><div class="loading-spinner"></div><p style="color:#64748b;margin-top:16px">加载中...</p></td></tr>
        <tr v-else-if="!list.length"><td colspan="9" class="muted">暂无数据</td></tr>
        <tr v-for="row in paginatedList" :key="row.namespace + '/' + row.name">
          <td>
            <span class="status" :class="row.status?.toLowerCase()">{{ row.status }}</span>
          </td>
          <td>
            <div><b>{{ row.name }}</b></div>
            <div class="muted">{{ row.namespace }}</div>
          </td>
          <td>
            <div>{{ row.target_kind }}</div>
            <div class="muted">{{ row.target_name }}</div>
          </td>
          <td>
            <span class="badge" :class="modeColor(row.update_mode)">{{ row.update_mode }}</span>
          </td>
          <td>
            <span>{{ row.container_name || '*' }}</span>
            <div class="muted">{{ (row.controlled_resources || []).join(',') }}</div>
          </td>
          <td>
            <div class="muted">CPU: {{ row.min_allowed_cpu || '-' }} ~ {{ row.max_allowed_cpu || '-' }}</div>
            <div class="muted">内存: {{ row.min_allowed_mem || '-' }} ~ {{ row.max_allowed_mem || '-' }}</div>
          </td>
          <td>
            <div v-if="row.recommendation">
              <button class="btn-mini" @click="showRecommendation(row)">📊 查看推荐</button>
            </div>
            <span v-else class="muted">收集中...</span>
          </td>
          <td class="muted">{{ row.creation_timestamp }}</td>
          <td>
            <button class="btn-mini" @click="openEdit(row)">✏️ 编辑</button>
            <button class="btn-mini danger" @click="onDelete(row)">🗑 删除</button>
          </td>
        </tr>
      </tbody>
    </table>
n    <!-- 分页 -->
    <Pagination v-if="total > 0" v-model:currentPage="currentPage" v-model:itemsPerPage="itemsPerPage" :totalItems="total" />

    <!-- 创建/编辑弹窗 -->
    <div v-if="formVisible" class="modal-mask" @click.self="formVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ form.isEdit ? '编辑 VPA' : '创建 VPA' }}</h2>
          <div v-if="!form.isEdit" class="view-toggle-buttons">
            <button class="view-toggle-btn" :class="{ active: createMode === 'form' }" @click="createMode = 'form'">📝 表单</button>
            <button class="view-toggle-btn" :class="{ active: createMode === 'yaml' }" @click="switchToYamlMode">📄 YAML</button>
          </div>
          <button class="close" @click="formVisible = false">✖</button>
        </div>
        <div class="modal-body">
          <!-- 表单模式 -->
          <div v-if="form.isEdit || createMode === 'form'">
          <div class="row">
            <label>命名空间 *</label>
            <input v-model="form.namespace" :disabled="form.isEdit" />
          </div>
          <div class="row">
            <label>VPA 名称 *</label>
            <input v-model="form.name" :disabled="form.isEdit" />
          </div>
          <div class="row">
            <label>目标资源 *</label>
            <div style="display: flex; gap: 8px; flex: 1">
              <select v-model="form.target_kind" style="width: 140px">
                <option>Deployment</option>
                <option>StatefulSet</option>
                <option>DaemonSet</option>
              </select>
              <input v-model="form.target_name" placeholder="目标名称" style="flex: 1" />
            </div>
          </div>
          <div class="row">
            <label>更新模式 *</label>
            <select v-model="form.update_mode">
              <option value="Off">Off（仅推荐，不自动更新）</option>
              <option value="Initial">Initial（仅创建时应用）</option>
              <option value="Recreate">Recreate（重建 Pod 应用）</option>
              <option value="Auto">Auto（推荐）</option>
            </select>
          </div>
          <div class="row">
            <label>容器名称</label>
            <input v-model="form.container_name" placeholder="留空表示所有容器（*）" />
          </div>
          <div class="row">
            <label>受控资源</label>
            <label class="checkbox">
              <input type="checkbox" v-model="form.has_cpu" /> CPU
            </label>
            <label class="checkbox">
              <input type="checkbox" v-model="form.has_mem" /> 内存
            </label>
          </div>
          <div class="row">
            <label>CPU 最小</label>
            <input v-model="form.min_allowed_cpu" placeholder="例: 100m" />
          </div>
          <div class="row">
            <label>CPU 最大</label>
            <input v-model="form.max_allowed_cpu" placeholder="例: 4" />
          </div>
          <div class="row">
            <label>内存 最小</label>
            <input v-model="form.min_allowed_mem" placeholder="例: 128Mi" />
          </div>
          <div class="row">
            <label>内存 最大</label>
            <input v-model="form.max_allowed_mem" placeholder="例: 4Gi" />
          </div>
          </div>
          <!-- YAML 模式 -->
          <div v-else class="yaml-editor-wrapper">
            <div class="yaml-toolbar">
              <button class="btn-mini" @click="loadVPAYamlTemplate">📄 加载模板</button>
              <button class="btn-mini" @click="copyYamlContent">📋 复制</button>
              <button class="btn-mini" @click="resetYamlContent">🔄 重置</button>
              <span class="muted" style="margin-left:auto">支持 autoscaling.k8s.io/v1 VerticalPodAutoscaler 单资源</span>
            </div>
            <textarea v-model="yamlContent" class="yaml-textarea" spellcheck="false"
              placeholder="请输入或加载 VPA YAML 模板..."></textarea>
            <div v-if="yamlError" class="error-box" style="margin-top:8px">{{ yamlError }}</div>
            <div class="hint">YAML 必须包含 kind: VerticalPodAutoscaler 与 apiVersion: autoscaling.k8s.io/v1。</div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="formVisible = false">取消</button>
          <button class="btn primary" @click="onSubmitForm" :disabled="submitting">
            {{ submitting ? '提交中...' : '确定' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 推荐结果弹窗 -->
    <div v-if="recVisible" class="modal-mask" @click.self="recVisible = false">
      <div class="modal" style="width: 700px">
        <div class="modal-header">
          <h2>VPA 推荐资源：{{ recRow.namespace }} / {{ recRow.name }}</h2>
          <button class="close" @click="recVisible = false">✖</button>
        </div>
        <div class="modal-body">
          <pre class="json-view">{{ JSON.stringify(recRow.recommendation, null, 2) }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="recVisible = false">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Pagination from '@/components/Pagination.vue'
import { useRoute } from 'vue-router'
import { vpaApi } from '@/api/cluster/workloads/autoscaler'

const route = useRoute()
const clusterId = computed(() => route.params.clusterId)

const available = ref(true)
const checking = ref(true)
const list = ref([])
const total = ref(0)
const loading = ref(true)
const currentPage = ref(1)
const itemsPerPage = ref(10)
const paginatedList = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return list.value.slice(start, start + itemsPerPage.value)
})
const submitting = ref(false)
const errorMsg = ref('')
const filter = ref({ namespace: '', name: '', page: 1, limit: 1000 })

const formVisible = ref(false)
const createMode = ref('form') // 'form' | 'yaml'
const yamlContent = ref('')
const yamlError = ref('')
const defaultForm = () => ({
  isEdit: false,
  namespace: '',
  name: '',
  target_kind: 'Deployment',
  target_name: '',
  target_api_version: 'apps/v1',
  update_mode: 'Auto',
  container_name: '',
  has_cpu: true,
  has_mem: true,
  min_allowed_cpu: '',
  min_allowed_mem: '',
  max_allowed_cpu: '',
  max_allowed_mem: '',
})
const form = ref(defaultForm())

const recVisible = ref(false)
const recRow = ref({ namespace: '', name: '', recommendation: {} })

async function checkAvailable() {
  void clusterId.value
  checking.value = true
  try {
    const res = await vpaApi.available()
    available.value = !!res?.data?.available
  } catch (e) {
    available.value = false
  } finally {
    checking.value = false
  }
}

async function loadList() {
  if (!available.value) return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await vpaApi.list(filter.value)
    list.value = res?.data?.list || []
    total.value = res?.data?.total || 0
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '加载失败'
  } finally {
    loading.value = false
  }
}

function onRefresh() { loadList() }

function getStatusCount(s) {
  return list.value.filter((r) => r.status === s).length
}

function modeColor(m) {
  return {
    Off: 'gray',
    Initial: 'blue',
    Recreate: 'orange',
    Auto: 'green',
  }[m] || ''
}

function openCreate() {
  form.value = defaultForm()
  createMode.value = 'form'
  yamlContent.value = ''
  yamlError.value = ''
  formVisible.value = true
}

function switchToYamlMode() {
  createMode.value = 'yaml'
  if (!yamlContent.value.trim()) {
    loadVPAYamlTemplate()
  }
}

function loadVPAYamlTemplate() {
  const ns = form.value.namespace || filter.value.namespace || 'default'
  const name = form.value.name || 'example-vpa'
  const targetKind = form.value.target_kind || 'Deployment'
  const targetName = form.value.target_name || 'example-deployment'
  const updateMode = form.value.update_mode || 'Auto'
  const containerName = form.value.container_name || '*'
  const minCPU = form.value.min_allowed_cpu || '100m'
  const minMem = form.value.min_allowed_mem || '128Mi'
  const maxCPU = form.value.max_allowed_cpu || '2'
  const maxMem = form.value.max_allowed_mem || '4Gi'
  yamlContent.value = `apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: ${name}
  namespace: ${ns}
spec:
  targetRef:
    apiVersion: apps/v1
    kind: ${targetKind}
    name: ${targetName}
  updatePolicy:
    updateMode: ${updateMode}
  resourcePolicy:
    containerPolicies:
      - containerName: "${containerName}"
        controlledResources:
          - cpu
          - memory
        minAllowed:
          cpu: ${minCPU}
          memory: ${minMem}
        maxAllowed:
          cpu: ${maxCPU}
          memory: ${maxMem}
`
  yamlError.value = ''
}

async function copyYamlContent() {
  try {
    await navigator.clipboard.writeText(yamlContent.value || '')
    yamlError.value = ''
  } catch {
    yamlError.value = '复制失败，请手动选择复制'
  }
}

function resetYamlContent() {
  if (!yamlContent.value || confirm('确定要重置 YAML 内容吗？')) {
    yamlContent.value = ''
    yamlError.value = ''
  }
}

async function createVPAFromYaml() {
  yamlError.value = ''
  const content = (yamlContent.value || '').trim()
  if (!content) {
    yamlError.value = '请输入 YAML 内容'
    return false
  }
  if (!/kind:\s*VerticalPodAutoscaler/.test(content)) {
    yamlError.value = 'YAML 必须包含 kind: VerticalPodAutoscaler'
    return false
  }
  if (!/apiVersion:\s*autoscaling\.k8s\.io\//.test(content)) {
    yamlError.value = 'YAML 必须包含 apiVersion: autoscaling.k8s.io/v1'
    return false
  }
  try {
    await vpaApi.createFromYaml({ yaml: content })
    return true
  } catch (e) {
    yamlError.value = e?.response?.data?.msg || e?.message || 'YAML 创建失败'
    return false
  }
}

function openEdit(row) {
  const cr = row.controlled_resources || []
  form.value = {
    isEdit: true,
    namespace: row.namespace,
    name: row.name,
    target_kind: row.target_kind || 'Deployment',
    target_name: row.target_name || '',
    target_api_version: 'apps/v1',
    update_mode: row.update_mode || 'Auto',
    container_name: row.container_name || '',
    has_cpu: cr.length === 0 || cr.includes('cpu'),
    has_mem: cr.length === 0 || cr.includes('memory'),
    min_allowed_cpu: row.min_allowed_cpu || '',
    min_allowed_mem: row.min_allowed_mem || '',
    max_allowed_cpu: row.max_allowed_cpu || '',
    max_allowed_mem: row.max_allowed_mem || '',
  }
  formVisible.value = true
}

async function onSubmitForm() {
  errorMsg.value = ''
  // YAML 模式（仅创建时可用）
  if (!form.value.isEdit && createMode.value === 'yaml') {
    submitting.value = true
    try {
      const ok = await createVPAFromYaml()
      if (ok) {
        formVisible.value = false
        await loadList()
      }
    } finally {
      submitting.value = false
    }
    return
  }

  if (!form.value.namespace || !form.value.name || !form.value.target_name) {
    errorMsg.value = '请填写必填项'
    return
  }
  if (!form.value.has_cpu && !form.value.has_mem) {
    errorMsg.value = '请至少勾选 CPU 或 内存 中的一个受控资源'
    return
  }
  submitting.value = true
  try {
    const ctrl = []
    if (form.value.has_cpu) ctrl.push('cpu')
    if (form.value.has_mem) ctrl.push('memory')
    const data = {
      namespace: form.value.namespace,
      name: form.value.name,
      target_kind: form.value.target_kind,
      target_name: form.value.target_name,
      target_api_version: form.value.target_api_version,
      update_mode: form.value.update_mode,
      container_name: form.value.container_name,
      controlled_resources: ctrl,
      min_allowed_cpu: form.value.min_allowed_cpu,
      min_allowed_mem: form.value.min_allowed_mem,
      max_allowed_cpu: form.value.max_allowed_cpu,
      max_allowed_mem: form.value.max_allowed_mem,
    }
    if (form.value.isEdit) {
      await vpaApi.update(data)
    } else {
      await vpaApi.create(data)
    }
    formVisible.value = false
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '提交失败'
  } finally {
    submitting.value = false
  }
}

async function onDelete(row) {
  if (!confirm(`确定删除 VPA: ${row.namespace}/${row.name} ?`)) return
  try {
    await vpaApi.delete({ namespace: row.namespace, name: row.name })
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '删除失败'
  }
}

function showRecommendation(row) {
  recRow.value = row
  recVisible.value = true
}

onMounted(async () => {
  await checkAvailable()
  if (available.value) {
    await loadList()
  }
})
</script>

<style scoped>
.vpa-mgmt { padding: 16px 24px; }
.view-header h1 { margin: 0 0 4px 0; }
.view-header p { color: #666; margin: 0 0 16px 0; }
.warn-box {
  padding: 12px 16px; background: #fffbe6; border: 1px solid #ffe58f;
  color: #ad6800; border-radius: 4px; margin-bottom: 12px;
}
.action-bar {
  display: flex; gap: 8px; align-items: center; flex-wrap: wrap;
  margin-bottom: 12px;
}
.search-input {
  height: 34px; padding: 0 8px; border: 1px solid #d8d8d8; border-radius: 4px;
  width: 180px;
}
.btn {
  height: 34px; padding: 0 12px; background: #fff; border: 1px solid #d8d8d8;
  border-radius: 4px; cursor: pointer;
}
.btn:disabled { background: #f5f5f5; color: #aaa; cursor: not-allowed; }
.btn:hover:not(:disabled) { border-color: #1677ff; color: #1677ff; }
.btn.primary { background: #1677ff; color: #fff; border-color: #1677ff; }
.btn.primary:hover:not(:disabled) { background: #4096ff; }
.btn-mini {
  height: 26px; padding: 0 8px; font-size: 12px;
  background: #fff; border: 1px solid #d8d8d8; border-radius: 4px;
  cursor: pointer; margin-right: 4px;
}
.btn-mini.danger { color: #ff4d4f; border-color: #ff4d4f; }
.btn-mini:hover { color: #1677ff; border-color: #1677ff; }
.error-box {
  padding: 8px 12px; background: #fff2f0; border: 1px solid #ffccc7;
  color: #ff4d4f; border-radius: 4px; margin-bottom: 12px;
}
.stats-bar {
  display: flex; gap: 24px; padding: 8px 12px; background: #fafafa;
  border-radius: 4px; margin-bottom: 12px; font-size: 13px;
}
.green { color: #52c41a; }
.yellow { color: #faad14; }
.red { color: #ff4d4f; }
.vpa-table {
  width: 100%; border-collapse: collapse; background: #fff;
  border: 1px solid #f0f0f0; font-size: 13px;
}
.vpa-table th, .vpa-table td {
  padding: 10px 12px; border-bottom: 1px solid #f0f0f0; text-align: left;
}
.vpa-table th { background: #fafafa; font-weight: 600; }
.muted { color: #999; }
.status {
  display: inline-block; padding: 2px 8px; border-radius: 3px;
  font-size: 12px; font-weight: 500;
}
.status.active { background: #f6ffed; color: #52c41a; border: 1px solid #b7eb8f; }
.status.pending { background: #fffbe6; color: #faad14; border: 1px solid #ffe58f; }
.badge {
  display: inline-block; padding: 2px 6px; border-radius: 3px;
  font-size: 12px; font-weight: 500;
  background: #e6f7ff; color: #1677ff;
}
.badge.gray { background: #f0f0f0; color: #666; }
.badge.green { background: #f6ffed; color: #52c41a; }
.badge.orange { background: #fff7e6; color: #fa8c16; }
.badge.blue { background: #e6f7ff; color: #1677ff; }
.modal-mask {
  position: fixed; left: 0; top: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5); z-index: 1000;
  display: flex; align-items: center; justify-content: center;
}
.modal {
  background: #fff; border-radius: 6px; width: 600px; max-height: 90vh;
  display: flex; flex-direction: column;
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 20px; border-bottom: 1px solid #f0f0f0;
}
.modal-header h2 { margin: 0; font-size: 16px; }
.close { background: none; border: none; cursor: pointer; font-size: 18px; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 20px; border-top: 1px solid #f0f0f0;
}
.row {
  display: flex; align-items: center; margin-bottom: 12px; gap: 12px;
}
.row label { width: 160px; color: #555; font-size: 13px; }
.row label.checkbox { width: auto; display: flex; align-items: center; gap: 4px; }
.row input, .row select {
  flex: 1; height: 32px; padding: 0 8px;
  border: 1px solid #d8d8d8; border-radius: 4px;
}
.json-view {
  background: #fafafa; padding: 12px; border-radius: 4px;
  font-family: monospace; font-size: 12px;
  max-height: 60vh; overflow: auto;
}

/* 表单 / YAML 双模式切换按钮 */
.view-toggle-buttons { display: flex; gap: 4px; margin: 0 16px; }
.view-toggle-btn {
  height: 28px; padding: 0 12px; font-size: 12px;
  background: #fafafa; border: 1px solid #d8d8d8; border-radius: 4px;
  cursor: pointer; color: #555;
}
.view-toggle-btn:hover { color: #1677ff; border-color: #1677ff; }
.view-toggle-btn.active { background: #1677ff; color: #fff; border-color: #1677ff; }

/* YAML 编辑器 */
.yaml-editor-wrapper { display: flex; flex-direction: column; gap: 8px; }
.yaml-toolbar { display: flex; align-items: center; gap: 8px; }
.yaml-textarea {
  width: 100%; min-height: 360px; padding: 12px;
  border: 1px solid #d8d8d8; border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Menlo', monospace;
  font-size: 13px; line-height: 1.6; resize: vertical;
  background: #fafafa;
}
.yaml-textarea:focus { outline: none; border-color: #1677ff; background: #fff; }
.hint { font-size: 12px; color: #faad14; margin-top: 4px; }
.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #326ce5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto;
}
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>

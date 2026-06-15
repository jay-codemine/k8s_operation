<template>
  <div class="hpa-mgmt">
    <div class="view-header">
      <h1>HPA 弹性扩缩容（水平扩展）</h1>
      <p>基于 CPU / 内存利用率自动调整 Deployment / StatefulSet 的副本数。618 促销场景支持批量预扩容。</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <input class="search-input" v-model="filter.namespace" placeholder="命名空间过滤" @keyup.enter="loadList" />
      <input class="search-input" v-model="filter.name" placeholder="HPA 名称（模糊）" @keyup.enter="loadList" />
      <button class="btn primary" @click="loadList">🔍 查询</button>
      <button class="btn" @click="onRefresh" :disabled="loading">🔄 刷新</button>
      <button class="btn primary" @click="openCreate">➕ 创建 HPA</button>
      <button class="btn warning" @click="openBatchScale" :disabled="!selected.length">
        🎉 批量扩缩容（{{ selected.length }}）
      </button>
      <button class="btn" @click="onBatchStatus" :disabled="!selected.length">📊 批量状态查询</button>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>

    <!-- 数据统计 -->
    <div class="stats-bar" v-if="list.length">
      <span>总计: <b>{{ total }}</b></span>
      <span>Active: <b class="green">{{ getStatusCount('Active') }}</b></span>
      <span>Pending: <b class="yellow">{{ getStatusCount('Pending') }}</b></span>
      <span>Failed: <b class="red">{{ getStatusCount('Failed') }}</b></span>
      <span>当前选中: <b>{{ selected.length }}</b></span>
    </div>

    <!-- 列表 -->
    <table class="hpa-table">
      <thead>
        <tr>
          <th style="width: 40px"><input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" /></th>
          <th>状态</th>
          <th>名称 / 命名空间</th>
          <th>目标</th>
          <th>副本</th>
          <th>CPU (目标/当前)</th>
          <th>内存 (目标/当前)</th>
          <th>创建时间</th>
          <th style="width: 280px">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading"><td colspan="9" class="muted">加载中...</td></tr>
        <tr v-else-if="!list.length"><td colspan="9" class="muted">暂无数据</td></tr>
        <tr v-for="row in list" :key="row.namespace + '/' + row.name">
          <td><input type="checkbox" :checked="isSelected(row)" @change="toggleSelect(row)" /></td>
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
            <span class="badge">{{ row.current_replicas }}/{{ row.desired_replicas }}</span>
            <div class="muted">范围: {{ row.min_replicas }} ~ {{ row.max_replicas }}</div>
          </td>
          <td>
            <span v-if="row.cpu_target_util != null">
              {{ row.cpu_target_util }}% / {{ row.current_cpu_util ?? '-' }}%
            </span>
            <span v-else class="muted">-</span>
          </td>
          <td>
            <span v-if="row.mem_target_util != null">
              {{ row.mem_target_util }}% / {{ row.current_memory_util ?? '-' }}%
            </span>
            <span v-else class="muted">-</span>
          </td>
          <td class="muted">{{ row.creation_timestamp }}</td>
          <td>
            <button class="btn-mini" @click="openScale(row)">🔧 调副本</button>
            <button class="btn-mini" @click="openEdit(row)">✏️ 编辑</button>
            <button class="btn-mini danger" @click="onDelete(row)">🗑 删除</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- 创建/编辑弹窗 -->
    <div v-if="formVisible" class="modal-mask" @click.self="formVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ form.isEdit ? '编辑 HPA' : '创建 HPA' }}</h2>
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
            <label>HPA 名称 *</label>
            <input v-model="form.name" :disabled="form.isEdit" />
          </div>
          <div class="row">
            <label>目标资源 *</label>
            <div style="display: flex; gap: 8px; flex: 1">
              <select v-model="form.target_kind" style="width: 140px">
                <option>Deployment</option>
                <option>StatefulSet</option>
                <option>ReplicaSet</option>
              </select>
              <input v-model="form.target_name" placeholder="目标名称" style="flex: 1" />
            </div>
          </div>
          <div class="row">
            <label>副本范围 *</label>
            <div style="display: flex; gap: 8px; flex: 1">
              <input type="number" v-model.number="form.min_replicas" placeholder="min" min="0" />
              <input type="number" v-model.number="form.max_replicas" placeholder="max" min="1" />
            </div>
          </div>
          <div class="row">
            <label>CPU 目标利用率 (%)</label>
            <input type="number" v-model.number="form.cpu_target_util" placeholder="例: 70（不需要时留空）" min="1" max="100" />
          </div>
          <div class="row">
            <label>内存目标利用率 (%)</label>
            <input type="number" v-model.number="form.mem_target_util" placeholder="例: 75（不需要时留空）" min="1" max="100" />
          </div>
          <div class="row">
            <label>扩容稳定窗口（秒）</label>
            <input type="number" v-model.number="form.scale_up_stab_seconds" placeholder="可选" min="0" />
          </div>
          <div class="row">
            <label>缩容稳定窗口（秒）</label>
            <input type="number" v-model.number="form.scale_down_stab_seconds" placeholder="可选" min="0" />
          </div>
          <div class="hint">至少填写 CPU 或 内存 中的一个目标利用率。</div>
          </div>
          <!-- YAML 模式 -->
          <div v-else class="yaml-editor-wrapper">
            <div class="yaml-toolbar">
              <button class="btn-mini" @click="loadHPAYamlTemplate">📄 加载模板</button>
              <button class="btn-mini" @click="copyYamlContent">📋 复制</button>
              <button class="btn-mini" @click="resetYamlContent">🔄 重置</button>
              <span class="muted" style="margin-left:auto">支持 autoscaling/v2 HorizontalPodAutoscaler 单资源</span>
            </div>
            <textarea v-model="yamlContent" class="yaml-textarea" spellcheck="false"
              placeholder="请输入或加载 HPA YAML 模板..."></textarea>
            <div v-if="yamlError" class="error-box" style="margin-top:8px">{{ yamlError }}</div>
            <div class="hint">YAML 必须包含 kind: HorizontalPodAutoscaler 与 apiVersion: autoscaling/v2。</div>
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

    <!-- 单独修改副本数弹窗 -->
    <div v-if="scaleVisible" class="modal-mask" @click.self="scaleVisible = false">
      <div class="modal" style="width: 480px">
        <div class="modal-header">
          <h2>修改副本数：{{ scaleForm.namespace }} / {{ scaleForm.name }}</h2>
          <button class="close" @click="scaleVisible = false">✖</button>
        </div>
        <div class="modal-body">
          <div class="row">
            <label>最小副本</label>
            <input type="number" v-model.number="scaleForm.min_replicas" min="0" />
          </div>
          <div class="row">
            <label>最大副本</label>
            <input type="number" v-model.number="scaleForm.max_replicas" min="1" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="scaleVisible = false">取消</button>
          <button class="btn primary" @click="onScaleSubmit">确定</button>
        </div>
      </div>
    </div>

    <!-- 批量扩缩容（618 模式） -->
    <div v-if="batchVisible" class="modal-mask" @click.self="batchVisible = false">
      <div class="modal" style="width: 720px">
        <div class="modal-header">
          <h2>🎉 批量扩缩容（618 促销）</h2>
          <button class="close" @click="batchVisible = false">✖</button>
        </div>
        <div class="modal-body">
          <div class="quick-actions">
            <span>快捷模板：</span>
            <button class="btn-mini" @click="applyTemplate('promo')">促销扩容（min×3, max×3）</button>
            <button class="btn-mini" @click="applyTemplate('peak')">峰值扩容（min×2, max×2）</button>
            <button class="btn-mini" @click="applyTemplate('back')">恢复（按原最小副本回缩）</button>
          </div>
          <table class="hpa-table small">
            <thead>
              <tr>
                <th>命名空间</th>
                <th>HPA 名称</th>
                <th>新 Min</th>
                <th>新 Max</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(it, idx) in batchItems" :key="it.namespace + it.name">
                <td>{{ it.namespace }}</td>
                <td>{{ it.name }}</td>
                <td><input type="number" v-model.number="batchItems[idx].min_replicas" min="0" /></td>
                <td><input type="number" v-model.number="batchItems[idx].max_replicas" min="1" /></td>
              </tr>
            </tbody>
          </table>
          <div v-if="batchResults.length" class="batch-result">
            <h3>执行结果（成功 {{ batchSuccessCnt }} / 失败 {{ batchFailCnt }}）</h3>
            <table class="hpa-table small">
              <thead>
                <tr>
                  <th>命名空间/名称</th>
                  <th>结果</th>
                  <th>当前/期望</th>
                  <th>min~max</th>
                  <th>消息</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="r in batchResults" :key="r.namespace + r.name">
                  <td>{{ r.namespace }} / {{ r.name }}</td>
                  <td>
                    <span class="status" :class="r.success ? 'active' : 'failed'">
                      {{ r.success ? '成功' : '失败' }}
                    </span>
                  </td>
                  <td>{{ r.current_replicas }}/{{ r.desired_replicas }}</td>
                  <td>{{ r.min_replicas }} ~ {{ r.max_replicas }}</td>
                  <td class="muted">{{ r.message }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="batchVisible = false">关闭</button>
          <button class="btn warning" @click="onBatchSubmit" :disabled="submitting">
            {{ submitting ? '执行中...' : '🚀 立即执行批量扩缩容' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { hpaApi } from '@/api/cluster/workloads/autoscaler'

const route = useRoute()
const clusterId = computed(() => route.params.clusterId)

const list = ref([])
const total = ref(0)
const loading = ref(false)
const submitting = ref(false)
const errorMsg = ref('')
const filter = ref({ namespace: '', name: '', page: 1, limit: 50 })
const selected = ref([])

// 创建/编辑表单
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
  min_replicas: 1,
  max_replicas: 10,
  cpu_target_util: 70,
  mem_target_util: null,
  scale_up_stab_seconds: null,
  scale_down_stab_seconds: null,
})
const form = ref(defaultForm())

// scale 表单
const scaleVisible = ref(false)
const scaleForm = ref({ namespace: '', name: '', min_replicas: 1, max_replicas: 10 })

// 批量
const batchVisible = ref(false)
const batchItems = ref([])
const batchResults = ref([])
const batchSuccessCnt = ref(0)
const batchFailCnt = ref(0)
const originalMinMap = ref({}) // 记录原始 min 值用于恢复

function setClusterHeader() {
  // X-Cluster-ID 由 http 拦截器自动从路由参数或 store 中提取，这里仅作为提示使用
  void clusterId.value
}

async function loadList() {
  setClusterHeader()
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await hpaApi.list(filter.value)
    list.value = res?.data?.list || []
    total.value = res?.data?.total || 0
    // 缓存原始 min/max 供恢复模板使用
    list.value.forEach((item) => {
      const k = `${item.namespace}/${item.name}`
      if (originalMinMap.value[k] == null) {
        originalMinMap.value[k] = { min: item.min_replicas, max: item.max_replicas }
      }
    })
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

const isAllSelected = computed(
  () => list.value.length > 0 && selected.value.length === list.value.length,
)

function isSelected(row) {
  return selected.value.find((x) => x.namespace === row.namespace && x.name === row.name)
}

function toggleSelect(row) {
  if (isSelected(row)) {
    selected.value = selected.value.filter((x) => !(x.namespace === row.namespace && x.name === row.name))
  } else {
    selected.value.push({ namespace: row.namespace, name: row.name })
  }
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selected.value = []
  } else {
    selected.value = list.value.map((r) => ({ namespace: r.namespace, name: r.name }))
  }
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
    loadHPAYamlTemplate()
  }
}

function loadHPAYamlTemplate() {
  const ns = form.value.namespace || filter.value.namespace || 'default'
  const name = form.value.name || 'example-hpa'
  const targetKind = form.value.target_kind || 'Deployment'
  const targetName = form.value.target_name || 'example-deployment'
  const minReplicas = form.value.min_replicas ?? 1
  const maxReplicas = form.value.max_replicas ?? 10
  const cpu = form.value.cpu_target_util ?? 70
  yamlContent.value = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${name}
  namespace: ${ns}
  labels:
    app: ${targetName}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: ${targetKind}
    name: ${targetName}
  minReplicas: ${minReplicas}
  maxReplicas: ${maxReplicas}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: ${cpu}
    # 可选：内存利用率指标
    # - type: Resource
    #   resource:
    #     name: memory
    #     target:
    #       type: Utilization
    #       averageUtilization: 75
  # 可选：扩/缩容稳定窗口
  # behavior:
  #   scaleUp:
  #     stabilizationWindowSeconds: 60
  #   scaleDown:
  #     stabilizationWindowSeconds: 300
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

async function createHPAFromYaml() {
  yamlError.value = ''
  const content = (yamlContent.value || '').trim()
  if (!content) {
    yamlError.value = '请输入 YAML 内容'
    return false
  }
  if (!/kind:\s*HorizontalPodAutoscaler/.test(content)) {
    yamlError.value = 'YAML 必须包含 kind: HorizontalPodAutoscaler'
    return false
  }
  if (!/apiVersion:\s*autoscaling\//.test(content)) {
    yamlError.value = 'YAML 必须包含 apiVersion: autoscaling/v2'
    return false
  }
  try {
    await hpaApi.createFromYaml({ yaml: content })
    return true
  } catch (e) {
    yamlError.value = e?.response?.data?.msg || e?.message || 'YAML 创建失败'
    return false
  }
}

function openEdit(row) {
  form.value = {
    isEdit: true,
    namespace: row.namespace,
    name: row.name,
    target_kind: row.target_kind || 'Deployment',
    target_name: row.target_name || '',
    target_api_version: 'apps/v1',
    min_replicas: row.min_replicas,
    max_replicas: row.max_replicas,
    cpu_target_util: row.cpu_target_util ?? null,
    mem_target_util: row.mem_target_util ?? null,
    scale_up_stab_seconds: null,
    scale_down_stab_seconds: null,
  }
  formVisible.value = true
}

async function onSubmitForm() {
  errorMsg.value = ''
  // YAML 模式（仅创建时可用）
  if (!form.value.isEdit && createMode.value === 'yaml') {
    setClusterHeader()
    submitting.value = true
    try {
      const ok = await createHPAFromYaml()
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
  if (!form.value.cpu_target_util && !form.value.mem_target_util) {
    errorMsg.value = '请至少填写 CPU 或 内存 目标利用率'
    return
  }
  setClusterHeader()
  submitting.value = true
  try {
    const data = { ...form.value }
    delete data.isEdit
    if (form.value.isEdit) {
      await hpaApi.update(data)
    } else {
      await hpaApi.create(data)
    }
    formVisible.value = false
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '提交失败'
  } finally {
    submitting.value = false
  }
}

function openScale(row) {
  scaleForm.value = {
    namespace: row.namespace,
    name: row.name,
    min_replicas: row.min_replicas,
    max_replicas: row.max_replicas,
  }
  scaleVisible.value = true
}

async function onScaleSubmit() {
  setClusterHeader()
  try {
    await hpaApi.scale(scaleForm.value)
    scaleVisible.value = false
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '修改失败'
  }
}

async function onDelete(row) {
  if (!confirm(`确定删除 HPA: ${row.namespace}/${row.name} ?`)) return
  setClusterHeader()
  try {
    await hpaApi.delete({ namespace: row.namespace, name: row.name })
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '删除失败'
  }
}

function openBatchScale() {
  batchResults.value = []
  batchSuccessCnt.value = 0
  batchFailCnt.value = 0
  batchItems.value = selected.value.map((s) => {
    const cur = list.value.find((r) => r.namespace === s.namespace && r.name === s.name)
    return {
      namespace: s.namespace,
      name: s.name,
      min_replicas: cur?.min_replicas ?? 1,
      max_replicas: cur?.max_replicas ?? 10,
    }
  })
  batchVisible.value = true
}

function applyTemplate(t) {
  batchItems.value.forEach((it) => {
    const k = `${it.namespace}/${it.name}`
    const orig = originalMinMap.value[k]
    if (t === 'promo') {
      it.min_replicas = (orig?.min ?? it.min_replicas) * 3
      it.max_replicas = (orig?.max ?? it.max_replicas) * 3
    } else if (t === 'peak') {
      it.min_replicas = (orig?.min ?? it.min_replicas) * 2
      it.max_replicas = (orig?.max ?? it.max_replicas) * 2
    } else if (t === 'back') {
      it.min_replicas = orig?.min ?? it.min_replicas
      it.max_replicas = orig?.max ?? it.max_replicas
    }
  })
}

async function onBatchSubmit() {
  setClusterHeader()
  submitting.value = true
  try {
    const res = await hpaApi.batchScale({ items: batchItems.value })
    batchResults.value = res?.data?.results || []
    batchSuccessCnt.value = res?.data?.success || 0
    batchFailCnt.value = res?.data?.fail || 0
    await loadList()
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '批量扩缩容失败'
  } finally {
    submitting.value = false
  }
}

async function onBatchStatus() {
  setClusterHeader()
  try {
    const res = await hpaApi.batchStatus({ items: selected.value })
    batchResults.value = res?.data?.results || []
    batchSuccessCnt.value = res?.data?.success || 0
    batchFailCnt.value = res?.data?.fail || 0
    batchItems.value = batchResults.value.map((r) => ({
      namespace: r.namespace,
      name: r.name,
      min_replicas: r.min_replicas,
      max_replicas: r.max_replicas,
    }))
    batchVisible.value = true
  } catch (e) {
    errorMsg.value = e?.response?.data?.msg || e?.message || '查询失败'
  }
}

onMounted(() => {
  setClusterHeader()
  loadList()
})
</script>

<style scoped>
.hpa-mgmt { padding: 16px 24px; }
.view-header h1 { margin: 0 0 4px 0; }
.view-header p { color: #666; margin: 0 0 16px 0; }
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
.btn:hover { border-color: #1677ff; color: #1677ff; }
.btn.primary { background: #1677ff; color: #fff; border-color: #1677ff; }
.btn.primary:hover { background: #4096ff; }
.btn.warning { background: #faad14; color: #fff; border-color: #faad14; }
.btn.warning:hover { background: #ffc53d; }
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
.hpa-table {
  width: 100%; border-collapse: collapse; background: #fff;
  border: 1px solid #f0f0f0; font-size: 13px;
}
.hpa-table th, .hpa-table td {
  padding: 10px 12px; border-bottom: 1px solid #f0f0f0; text-align: left;
}
.hpa-table th { background: #fafafa; font-weight: 600; }
.hpa-table.small td, .hpa-table.small th { padding: 6px 8px; }
.muted { color: #999; }
.status {
  display: inline-block; padding: 2px 8px; border-radius: 3px;
  font-size: 12px; font-weight: 500;
}
.status.active { background: #f6ffed; color: #52c41a; border: 1px solid #b7eb8f; }
.status.pending { background: #fffbe6; color: #faad14; border: 1px solid #ffe58f; }
.status.failed { background: #fff2f0; color: #ff4d4f; border: 1px solid #ffccc7; }
.badge {
  display: inline-block; padding: 2px 6px; border-radius: 3px;
  background: #e6f7ff; color: #1677ff; font-size: 12px; font-weight: 500;
}

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
.row input, .row select {
  flex: 1; height: 32px; padding: 0 8px;
  border: 1px solid #d8d8d8; border-radius: 4px;
}
.hint { font-size: 12px; color: #faad14; margin-top: 8px; }
.quick-actions {
  display: flex; gap: 8px; align-items: center; margin-bottom: 12px;
  padding: 8px; background: #fafafa; border-radius: 4px;
}
.batch-result { margin-top: 16px; }
.batch-result h3 { font-size: 14px; margin: 0 0 8px 0; }

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
</style>

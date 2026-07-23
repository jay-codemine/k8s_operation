<template>
  <div class="datasources-page">
    <!-- 顶部区域 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">
          <span class="title-icon">🔌</span>
          数据源管理
        </h2>
        <span class="page-subtitle">统一管理 Prometheus、Loki、Alertmanager 等监控数据源</span>
      </div>
      <div class="header-actions">
        <button class="btn-ghost" @click="refreshAll" :disabled="refreshing">
          <span class="icon" :class="{ spinning: refreshing }">↻</span> 批量检测
        </button>
        <button class="btn-primary" @click="showQuickAdd = true">
          <span class="icon">+</span> 新增数据源
        </button>
      </div>
    </div>

    <!-- 数据源类型快捷入口 -->
    <div class="type-cards">
      <div
        class="type-card"
        v-for="t in typePresets"
        :key="t.type"
        :class="{ active: filters.type === t.type }"
        @click="toggleFilter(t.type)"
      >
        <div class="type-card-icon" :style="{ background: t.gradient }">
          <span>{{ t.icon }}</span>
        </div>
        <div class="type-card-info">
          <span class="type-card-name">{{ t.label }}</span>
          <span class="type-card-count">{{ getTypeCount(t.type) }} 个</span>
        </div>
      </div>
    </div>

    <!-- 搜索与统计 -->
    <div class="toolbar">
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input v-model="filters.keyword" placeholder="搜索数据源名称、URL、描述..." @input="debouncedLoad" />
        <button class="clear-btn" v-if="filters.keyword" @click="filters.keyword = ''; loadList()">×</button>
      </div>
      <div class="stats-bar">
        <span class="stat-item">
          <span class="stat-dot connected"></span> 已连接 {{ connectedCount }}
        </span>
        <span class="stat-item">
          <span class="stat-dot disconnected"></span> 未连接 {{ disconnectedCount }}
        </span>
        <span class="stat-item total">共 {{ list.length }} 个数据源</span>
      </div>
    </div>

    <!-- 数据源列表 -->
    <div class="ds-list" v-if="list.length">
      <div
        class="ds-card"
        v-for="ds in filteredList"
        :key="ds.id"
        :class="{ disabled: !ds.enabled }"
      >
        <!-- 顶部色条 -->
        <div class="ds-card-stripe" :style="{ background: getTypeGradient(ds.type) }"></div>
        
        <div class="ds-card-content">
          <!-- 头部 -->
          <div class="ds-card-head">
            <div class="ds-icon-wrap" :style="{ background: getTypeBg(ds.type) }">
              <span class="ds-icon">{{ getTypeIcon(ds.type) }}</span>
            </div>
            <div class="ds-main-info">
              <div class="ds-name-row">
                <span class="ds-name">{{ ds.name }}</span>
                <span class="ds-default-tag" v-if="ds.is_default">默认</span>
                <span class="ds-type-tag">{{ getTypeLabel(ds.type) }}</span>
              <span class="ds-cluster-tag global" v-if="!ds.cluster_id">🌐 全局</span>
              <span class="ds-cluster-tag bound" v-else>☸️ {{ getClusterName(ds.cluster_id) }}</span>
              </div>
              <span class="ds-url">{{ ds.url }}</span>
            </div>
            <div class="ds-status-wrap">
              <div class="status-indicator" :class="ds.status || 'unknown'">
                <span class="status-pulse" v-if="ds.status === 'connected'"></span>
                <span class="status-icon">{{ getStatusIcon(ds.status) }}</span>
              </div>
              <span class="status-label">{{ statusMap[ds.status] || '未检测' }}</span>
            </div>
          </div>

          <!-- 元信息 -->
          <div class="ds-meta-row">
            <div class="meta-chip">
              <span class="meta-label">认证</span>
              <span class="meta-value">{{ authTypeMap[ds.auth_type] || '无' }}</span>
            </div>
            <div class="meta-chip">
              <span class="meta-label">超时</span>
              <span class="meta-value">{{ ds.timeout || 30 }}s</span>
            </div>
            <div class="meta-chip">
              <span class="meta-label">采集间隔</span>
              <span class="meta-value">{{ ds.scrape_interval || 15 }}s</span>
            </div>
            <div class="meta-chip" v-if="ds.last_check_at">
              <span class="meta-label">最后检测</span>
              <span class="meta-value">{{ formatTime(ds.last_check_at) }}</span>
            </div>
          </div>

          <!-- 描述 -->
          <p class="ds-desc" v-if="ds.description">{{ ds.description }}</p>

          <!-- 操作栏 -->
          <div class="ds-actions">
            <button class="action-btn test" @click="testConnection(ds)" :disabled="testing === ds.id">
              {{ testing === ds.id ? '检测中...' : '🔗 连通测试' }}
            </button>
            <button class="action-btn edit" @click="openEdit(ds)">✏️ 编辑</button>
            <button class="action-btn toggle" @click="toggleEnabled(ds)">
              {{ ds.enabled ? '⏸ 禁用' : '▶️ 启用' }}
            </button>
            <button class="action-btn default" v-if="!ds.is_default" @click="setDefault(ds)">⭐ 设默认</button>
            <button class="action-btn danger" @click="confirmDelete(ds)">🗑️ 删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-else-if="!loading">
      <div class="empty-visual">
        <div class="empty-circles">
          <div class="circle c1"></div>
          <div class="circle c2"></div>
          <div class="circle c3"></div>
        </div>
        <span class="empty-icon-large">🔌</span>
      </div>
      <h3>暂未配置数据源</h3>
      <p>添加 Prometheus / Loki / Alertmanager 等数据源，开启全栈可观测性</p>
      <button class="btn-primary lg" @click="showQuickAdd = true">+ 添加第一个数据源</button>
    </div>

    <!-- ============ 快捷新增面板 ============ -->
    <div class="modal-overlay" v-if="showQuickAdd" @click.self="showQuickAdd = false">
      <div class="quick-add-panel">
        <div class="panel-header">
          <h3>选择数据源类型</h3>
          <button class="modal-close" @click="showQuickAdd = false">×</button>
        </div>
        <div class="type-grid">
          <div
            class="type-option"
            v-for="t in typePresets"
            :key="t.type"
            @click="quickCreate(t)"
          >
            <div class="type-option-icon" :style="{ background: t.gradient }">
              <span>{{ t.icon }}</span>
            </div>
            <span class="type-option-name">{{ t.label }}</span>
            <span class="type-option-desc">{{ t.desc }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ============ 编辑弹窗 ============ -->
    <div class="modal-overlay" v-if="dialogVisible" @click.self="dialogVisible = false">
      <div class="modal-dialog wide">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-type-icon" :style="{ background: getTypeGradient(form.type) }">
              {{ getTypeIcon(form.type) }}
            </span>
            <h3>{{ editingId ? '编辑数据源' : '新增数据源' }}</h3>
          </div>
          <button class="modal-close" @click="dialogVisible = false">×</button>
        </div>
        
        <div class="modal-body">
          <div class="form-section">
            <h4 class="section-title">基本信息</h4>
            <div class="form-grid">
              <div class="form-item">
                <label>名称 <span class="required">*</span></label>
                <input v-model="form.name" placeholder="如: prod-prometheus-01" />
              </div>
              <div class="form-item">
                <label>类型 <span class="required">*</span></label>
                <select v-model="form.type" @change="onTypeChange">
                  <option v-for="t in typePresets" :key="t.type" :value="t.type">{{ t.label }}</option>
                </select>
              </div>
            </div>
            <div class="form-grid" style="margin-top:14px;">
              <div class="form-item">
                <label>关联监控集群 <span class="hint-mini">（决定该数据源在哪个 K8s 集群下展示）</span></label>
                <select v-model.number="form.cluster_id">
                  <option :value="0">🌐 全局共享（不绑定集群）</option>
                  <option v-for="c in clusterList" :key="c.id" :value="Number(c.id)">☸️ {{ c.cluster_name }}</option>
                </select>
              </div>
              <div class="form-item">
                <label>当前归属</label>
                <div class="cluster-badge-preview">
                  <span v-if="!form.cluster_id" class="cluster-chip global">🌐 全局共享</span>
                  <span v-else class="cluster-chip bound">☸️ {{ getClusterName(form.cluster_id) || '未知集群' }}</span>
                </div>
              </div>
            </div>
            <div class="form-item full">
              <label>连接地址 <span class="required">*</span></label>
              <div class="url-input-wrap">
                <input v-model="form.url" :placeholder="getUrlPlaceholder(form.type)" />
                <span class="url-hint">{{ getUrlHint(form.type) }}</span>
              </div>
            </div>
            <div class="form-item full">
              <label>描述</label>
              <input v-model="form.description" placeholder="数据源用途描述（可选）" />
            </div>
          </div>

          <div class="form-section">
            <h4 class="section-title">连接配置</h4>
            <div class="form-grid">
              <div class="form-item">
                <label>认证方式</label>
                <select v-model="form.auth_type">
                  <option value="none">无认证</option>
                  <option value="basic">Basic Auth</option>
                  <option value="bearer">Bearer Token</option>
                  <option value="tls">TLS 双向认证</option>
                </select>
              </div>
              <div class="form-item">
                <label>超时时间(秒)</label>
                <input v-model.number="form.timeout" type="number" min="5" max="120" />
              </div>
              <div class="form-item">
                <label>采集间隔(秒)</label>
                <input v-model.number="form.scrape_interval" type="number" min="5" max="300" />
              </div>
              <div class="form-item">
                <label>访问模式</label>
                <select v-model="form.access_mode">
                  <option value="proxy">Proxy (后端代理)</option>
                  <option value="direct">Direct (直连)</option>
                </select>
              </div>
            </div>
          </div>

          <!-- 认证信息 -->
          <div class="form-section" v-if="form.auth_type !== 'none'">
            <h4 class="section-title">认证信息</h4>
            <div class="form-grid">
              <div class="form-item" v-if="form.auth_type === 'basic'">
                <label>用户名</label>
                <input v-model="form.auth_user" placeholder="Username" />
              </div>
              <div class="form-item">
                <label>{{ form.auth_type === 'bearer' ? 'Token' : '密码' }}</label>
                <input v-model="form.auth_pass" type="password" placeholder="认证凭据" />
              </div>
            </div>
          </div>

          <!-- 开关 -->
          <div class="form-section">
            <h4 class="section-title">高级选项</h4>
            <div class="switch-row">
              <label class="switch-label">
                <input type="checkbox" v-model="form.is_default" />
                <span class="switch-slider"></span>
                <span>设为默认数据源</span>
              </label>
              <label class="switch-label">
                <input type="checkbox" v-model="form.enabled" />
                <span class="switch-slider"></span>
                <span>启用</span>
              </label>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-ghost" @click="testFormConnection" :disabled="testingForm">
            {{ testingForm ? '测试中...' : '🔗 连通测试' }}
          </button>
          <span class="test-result" v-if="testResult" :class="testResult.ok ? 'success' : 'fail'">
            {{ testResult.ok ? 'check-circle' : '❌' }} {{ testResult.msg }}
          </span>
          <div class="footer-right">
            <button class="btn-outline" @click="dialogVisible = false">取消</button>
            <button class="btn-primary" @click="submitForm" :disabled="submitting">
              {{ submitting ? '保存中...' : '💾 保存' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-dialog sm">
        <div class="modal-header danger">
          <h3>⚠️ 确认删除</h3>
          <button class="modal-close" @click="deleteTarget = null">×</button>
        </div>
        <div class="modal-body">
          <div class="delete-info">
            <p>确定要删除数据源 <b>{{ deleteTarget.name }}</b> 吗？</p>
            <div class="delete-warning">
              <span>⚠️</span>
              <span>删除后，关联此数据源的告警规则将无法正常评估和触发通知。</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="deleteTarget = null">取消</button>
          <button class="btn-danger" @click="doDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import {
  listDatasources,
  createDatasource,
  updateDatasource,
  deleteDatasource,
  testDatasourceConnection,
  testDatasourceById,
} from '@/api/monitoring'
import { getK8sClusterList } from '@/api/platform/cluster.js'

// ===== 类型预设 =====
const typePresets = [
  { type: 'prometheus', label: 'Prometheus', icon: '🔥', desc: '时序指标监控', gradient: 'linear-gradient(135deg, #e25e3e, #ff8f6b)', defaultUrl: 'http://prometheus:9090' },
  { type: 'loki', label: 'Loki', icon: '📜', desc: '日志聚合引擎', gradient: 'linear-gradient(135deg, #2563eb, #60a5fa)', defaultUrl: 'http://loki:3100' },
  { type: 'alertmanager', label: 'Alertmanager', icon: '🚨', desc: '告警路由分发', gradient: 'linear-gradient(135deg, #dc2626, #f87171)', defaultUrl: 'http://alertmanager:9093' },
  { type: 'victoriametrics', label: 'VictoriaMetrics', icon: 'dashboard', desc: '高性能时序存储', gradient: 'linear-gradient(135deg, #7c3aed, #a78bfa)', defaultUrl: 'http://victoria-metrics:8428' },
  { type: 'thanos', label: 'Thanos', icon: '♾️', desc: '多集群指标聚合', gradient: 'linear-gradient(135deg, #0891b2, #67e8f9)', defaultUrl: 'http://thanos-query:9090' },
  { type: 'n9e', label: '夜莺 Nightingale', icon: '🦉', desc: '开源统一告警平台', gradient: 'linear-gradient(135deg, #d97706, #fbbf24)', defaultUrl: 'http://n9e:17000' },
  { type: 'grafana', label: 'Grafana', icon: 'dashboard', desc: '可视化看板', gradient: 'linear-gradient(135deg, #ea580c, #fb923c)', defaultUrl: 'http://grafana:3000' },
]

// ===== 状态 =====
const list = ref([])
const loading = ref(true)
const refreshing = ref(false)
const testing = ref(null)
const testingForm = ref(false)
const testResult = ref(null)
const dialogVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const deleteTarget = ref(null)
const showQuickAdd = ref(false)
const filters = reactive({ keyword: '', type: '' })

const statusMap = { connected: '已连接', disconnected: '未连接', unknown: '未检测' }
const authTypeMap = { none: '无认证', basic: 'Basic Auth', bearer: 'Bearer Token', tls: 'TLS' }

const defaultForm = () => ({
  name: '', type: 'prometheus', url: '', description: '',
  cluster_id: 0,
  access_mode: 'proxy', auth_type: 'none', auth_user: '', auth_pass: '',
  is_default: false, enabled: true, timeout: 30, scrape_interval: 15,
})

// 集群列表（用于关联下拉）
const clusterList = ref([])
async function loadClusters() {
  try {
    const res = await getK8sClusterList({ page: 1, limit: 200 })
    clusterList.value = res?.data?.list || []
  } catch { clusterList.value = [] }
}
function getClusterName(id) {
  if (!id || Number(id) === 0) return ''
  const c = clusterList.value.find(x => Number(x.id) === Number(id))
  return c?.cluster_name || ('#' + id)
}
const form = reactive(defaultForm())

// ===== 计算属性 =====
const filteredList = computed(() => {
  let items = list.value
  if (filters.type) items = items.filter(d => d.type === filters.type)
  return items
})
const connectedCount = computed(() => list.value.filter(d => d.status === 'connected').length)
const disconnectedCount = computed(() => list.value.filter(d => d.status === 'disconnected').length)

// ===== 工具函数 =====
function getTypeCount(type) {
  return list.value.filter(d => d.type === type).length
}
function getTypeIcon(type) {
  return typePresets.find(t => t.type === type)?.icon || 'wifi'
}
function getTypeLabel(type) {
  return typePresets.find(t => t.type === type)?.label || type
}
function getTypeGradient(type) {
  return typePresets.find(t => t.type === type)?.gradient || 'linear-gradient(135deg, #6b7280, #9ca3af)'
}
function getTypeBg(type) {
  const map = {
    prometheus: '#fef2f2', loki: '#eff6ff', alertmanager: '#fef2f2',
    victoriametrics: '#f5f3ff', thanos: '#ecfeff', n9e: '#fffbeb', grafana: '#fff7ed'
  }
  return map[type] || '#f9fafb'
}
function getStatusIcon(status) {
  return { connected: '●', disconnected: '●', unknown: '○' }[status] || '○'
}
function getUrlPlaceholder(type) {
  return typePresets.find(t => t.type === type)?.defaultUrl || 'http://host:port'
}
function getUrlHint(type) {
  const hints = {
    prometheus: '健康检查端点: /-/healthy',
    loki: '健康检查端点: /ready',
    alertmanager: '健康检查端点: /-/healthy',
    victoriametrics: '健康检查端点: /health',
    thanos: '健康检查端点: /-/healthy',
    n9e: '健康检查端点: /api/n9e/heartbeat',
    grafana: '健康检查端点: /api/health',
  }
  return hints[type] || ''
}
function formatTime(ts) {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

// ===== 筛选 =====
function toggleFilter(type) {
  filters.type = filters.type === type ? '' : type
}

// ===== 数据加载 =====
let debounceTimer = null
const debouncedLoad = () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(loadList, 300)
}

async function loadList() {
  loading.value = true
  try {
    const res = await listDatasources({ page: 1, size: 100, keyword: filters.keyword })
    if (res?.code === 0) list.value = res.data?.items || []
  } catch {} finally { loading.value = false }
}

// ===== 快捷新增 =====
function quickCreate(preset) {
  showQuickAdd.value = false
  editingId.value = null
  Object.assign(form, defaultForm())
  form.type = preset.type
  form.url = preset.defaultUrl
  form.name = preset.type + '-01'
  testResult.value = null
  dialogVisible.value = true
}

// ===== 编辑 =====
function openEdit(ds) {
  editingId.value = ds.id
  Object.assign(form, { ...ds })
  testResult.value = null
  dialogVisible.value = true
}

function onTypeChange() {
  const preset = typePresets.find(t => t.type === form.type)
  if (preset && !form.url) form.url = preset.defaultUrl
}

// ===== 提交 =====
async function submitForm() {
  if (!form.name || !form.url || !form.type) return alert('请填写必填字段（名称、类型、URL）')
  submitting.value = true
  try {
    if (editingId.value) {
      await updateDatasource(editingId.value, form)
    } else {
      await createDatasource(form)
    }
    dialogVisible.value = false
    loadList()
  } catch (e) {
    alert('保存失败: ' + (e?.msg || e?.message || '未知错误'))
  } finally { submitting.value = false }
}

// ===== 删除 =====
function confirmDelete(ds) { deleteTarget.value = ds }
async function doDelete() {
  try {
    await deleteDatasource(deleteTarget.value.id)
    deleteTarget.value = null
    loadList()
  } catch {}
}

// ===== 连通测试 =====
async function testConnection(ds) {
  testing.value = ds.id
  try {
    const res = await testDatasourceById(ds.id)
    if (res?.code === 0) {
      const { connected, message } = res.data
      alert(connected ? '✅ ' + message : '❌ ' + message)
      loadList()
    }
  } catch { alert('测试请求失败') } finally { testing.value = null }
}

async function testFormConnection() {
  testingForm.value = true
  testResult.value = null
  try {
    const res = await testDatasourceConnection(form)
    if (res?.code === 0) {
      testResult.value = { ok: res.data.connected, msg: res.data.message }
    }
  } catch { testResult.value = { ok: false, msg: '请求失败' } }
  finally { testingForm.value = false }
}

// ===== 启禁用 =====
async function toggleEnabled(ds) {
  try {
    await updateDatasource(ds.id, { ...ds, enabled: !ds.enabled })
    loadList()
  } catch {}
}

// ===== 设默认 =====
async function setDefault(ds) {
  try {
    await updateDatasource(ds.id, { ...ds, is_default: true })
    loadList()
  } catch {}
}

// ===== 批量检测 =====
async function refreshAll() {
  refreshing.value = true
  try {
    for (const ds of list.value) {
      if (ds.enabled) await testDatasourceById(ds.id)
    }
    await loadList()
  } catch {} finally { refreshing.value = false }
}

onMounted(() => { loadClusters(); loadList() })
</script>

<style scoped>
.datasources-page { padding: 24px 28px; min-height: 100vh; background: #f8fafc; }

/* Header */
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.page-title { margin: 0; font-size: 22px; color: #0f172a; display: flex; align-items: center; gap: 8px; font-weight: 700; }
.title-icon { font-size: 24px; }
.page-subtitle { font-size: 13px; color: #64748b; margin-top: 4px; display: block; }
.header-actions { display: flex; gap: 10px; }
.btn-primary { background: linear-gradient(135deg, #4f46e5, #7c3aed); color: #fff; border: none; padding: 10px 20px; border-radius: 10px; font-size: 14px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.2s; box-shadow: 0 2px 8px rgba(79,70,229,0.3); }
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(79,70,229,0.4); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.btn-primary.lg { padding: 12px 28px; font-size: 15px; }
.btn-ghost { background: #fff; color: #475569; border: 1px solid #e2e8f0; padding: 10px 18px; border-radius: 10px; font-size: 14px; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.2s; }
.btn-ghost:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-ghost:disabled { opacity: 0.6; }
.btn-outline { background: #fff; color: #475569; border: 1px solid #e2e8f0; padding: 10px 18px; border-radius: 10px; font-size: 14px; cursor: pointer; }
.btn-outline:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-danger { background: #dc2626; color: #fff; border: none; padding: 10px 18px; border-radius: 10px; font-size: 14px; cursor: pointer; }
.btn-danger:hover { background: #b91c1c; }

/* Type Cards */
.type-cards { display: flex; gap: 12px; margin-bottom: 20px; overflow-x: auto; padding-bottom: 4px; }
.type-card { display: flex; align-items: center; gap: 10px; padding: 12px 16px; background: #fff; border: 1px solid #e8ecf0; border-radius: 12px; cursor: pointer; transition: all 0.2s; min-width: 150px; user-select: none; }
.type-card:hover { border-color: #c7d2fe; box-shadow: 0 2px 8px rgba(0,0,0,0.04); }
.type-card.active { border-color: #4f46e5; background: #f5f3ff; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.type-card-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.type-card-info { display: flex; flex-direction: column; }
.type-card-name { font-size: 13px; font-weight: 600; color: #1e293b; }
.type-card-count { font-size: 11px; color: #94a3b8; }

/* Toolbar */
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.search-box { position: relative; flex: 1; max-width: 400px; }
.search-box input { width: 100%; padding: 10px 14px 10px 36px; border: 1px solid #e2e8f0; border-radius: 10px; font-size: 14px; outline: none; background: #fff; box-sizing: border-box; }
.search-box input:focus { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); font-size: 14px; }
.clear-btn { position: absolute; right: 10px; top: 50%; transform: translateY(-50%); background: none; border: none; font-size: 18px; color: #94a3b8; cursor: pointer; }
.stats-bar { display: flex; gap: 16px; align-items: center; }
.stat-item { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #64748b; }
.stat-item.total { font-weight: 600; color: #1e293b; }
.stat-dot { width: 8px; height: 8px; border-radius: 50%; }
.stat-dot.connected { background: #10b981; }
.stat-dot.disconnected { background: #ef4444; }

/* DS List */
.ds-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(480px, 1fr)); gap: 16px; }
.ds-card { background: #fff; border-radius: 14px; border: 1px solid #e8ecf0; overflow: hidden; transition: all 0.25s; position: relative; }
.ds-card:hover { box-shadow: 0 8px 24px rgba(0,0,0,0.06); transform: translateY(-2px); }
.ds-card.disabled { opacity: 0.55; }
.ds-card-stripe { height: 4px; }
.ds-card-content { padding: 18px 22px; }
.ds-card-head { display: flex; align-items: center; gap: 14px; margin-bottom: 14px; }
.ds-icon-wrap { width: 46px; height: 46px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0; }
.ds-main-info { flex: 1; min-width: 0; }
.ds-name-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.ds-name { font-size: 15px; font-weight: 600; color: #0f172a; }
.ds-default-tag { background: linear-gradient(135deg, #4f46e5, #7c3aed); color: #fff; padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 500; }
.ds-type-tag { background: #f1f5f9; color: #64748b; padding: 2px 8px; border-radius: 4px; font-size: 11px; }
.ds-url { font-size: 12px; color: #94a3b8; font-family: 'JetBrains Mono', monospace; display: block; margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ds-status-wrap { display: flex; flex-direction: column; align-items: center; gap: 4px; }
.status-indicator { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; position: relative; }
.status-indicator.connected { background: #d1fae5; color: #059669; }
.status-indicator.disconnected { background: #fee2e2; color: #dc2626; }
.status-indicator.unknown { background: #f1f5f9; color: #94a3b8; }
.status-icon { font-size: 12px; z-index: 1; }
.status-pulse { position: absolute; inset: 0; border-radius: 50%; background: #10b981; opacity: 0.3; animation: pulse 2s infinite; }
@keyframes pulse { 0%, 100% { transform: scale(1); opacity: 0.3; } 50% { transform: scale(1.3); opacity: 0; } }
.status-label { font-size: 11px; color: #64748b; }

/* Meta */
.ds-meta-row { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 10px; }
.meta-chip { background: #f8fafc; border: 1px solid #f1f5f9; padding: 4px 10px; border-radius: 6px; display: flex; align-items: center; gap: 4px; }
.meta-label { font-size: 11px; color: #94a3b8; }
.meta-value { font-size: 12px; color: #334155; font-weight: 500; }
.ds-desc { font-size: 13px; color: #64748b; margin: 8px 0; line-height: 1.5; }

/* Actions */
.ds-actions { display: flex; gap: 8px; padding-top: 12px; border-top: 1px solid #f1f5f9; flex-wrap: wrap; }
.action-btn { padding: 6px 12px; border-radius: 7px; font-size: 12px; cursor: pointer; border: 1px solid #e2e8f0; background: #fff; color: #475569; transition: all 0.15s; }
.action-btn:hover { background: #f8fafc; }
.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.action-btn.test:hover { border-color: #4f46e5; color: #4f46e5; }
.action-btn.edit:hover { border-color: #2563eb; color: #2563eb; }
.action-btn.toggle:hover { border-color: #d97706; color: #d97706; }
.action-btn.default:hover { border-color: #7c3aed; color: #7c3aed; }
.action-btn.danger { color: #dc2626; border-color: #fecaca; }
.action-btn.danger:hover { background: #fef2f2; }

/* Empty */
.empty-state { text-align: center; padding: 80px 20px; }
.empty-visual { position: relative; display: inline-block; margin-bottom: 24px; }
.empty-circles { position: absolute; inset: -20px; }
.circle { position: absolute; border-radius: 50%; border: 2px dashed #e2e8f0; animation: rotate 20s linear infinite; }
.c1 { inset: 0; } .c2 { inset: 10px; animation-direction: reverse; } .c3 { inset: 20px; }
@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.empty-icon-large { font-size: 64px; position: relative; z-index: 1; }
.empty-state h3 { color: #0f172a; font-size: 18px; margin: 0 0 8px; }
.empty-state p { color: #64748b; margin: 0 0 24px; font-size: 14px; }

/* Quick Add Panel */
.quick-add-panel { background: #fff; border-radius: 16px; width: 680px; max-height: 85vh; overflow-y: auto; box-shadow: 0 25px 60px rgba(0,0,0,0.15); }
.panel-header { display: flex; justify-content: space-between; align-items: center; padding: 22px 28px; border-bottom: 1px solid #f1f5f9; }
.panel-header h3 { margin: 0; font-size: 18px; color: #0f172a; }
.type-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 14px; padding: 24px 28px; }
.type-option { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 20px 16px; border: 1px solid #e8ecf0; border-radius: 14px; cursor: pointer; transition: all 0.2s; text-align: center; }
.type-option:hover { border-color: #4f46e5; background: #faf5ff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(79,70,229,0.1); }
.type-option-icon { width: 48px; height: 48px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.type-option-name { font-size: 14px; font-weight: 600; color: #1e293b; }
.type-option-desc { font-size: 11px; color: #94a3b8; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(15,23,42,0.6); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 16px; width: 600px; max-height: 85vh; overflow-y: auto; box-shadow: 0 25px 60px rgba(0,0,0,0.2); }
.modal-dialog.wide { width: 680px; }
.modal-dialog.sm { width: 440px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 28px; border-bottom: 1px solid #f1f5f9; }
.modal-header.danger { background: #fef2f2; }
.modal-header h3 { margin: 0; font-size: 18px; color: #0f172a; }
.modal-title-row { display: flex; align-items: center; gap: 12px; }
.modal-type-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #94a3b8; transition: color 0.2s; }
.modal-close:hover { color: #1e293b; }
.modal-body { padding: 24px 28px; }
.modal-footer { display: flex; align-items: center; gap: 12px; padding: 16px 28px; border-top: 1px solid #f1f5f9; }
.footer-right { margin-left: auto; display: flex; gap: 10px; }
.test-result { font-size: 13px; }
.test-result.success { color: #059669; }
.test-result.fail { color: #dc2626; }

/* Form */
.form-section { margin-bottom: 24px; }
.section-title { font-size: 14px; font-weight: 600; color: #334155; margin: 0 0 14px; padding-bottom: 8px; border-bottom: 1px solid #f1f5f9; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-item { display: flex; flex-direction: column; gap: 6px; }
.form-item.full { grid-column: 1 / -1; margin-top: 14px; }
.form-item label { font-size: 13px; font-weight: 500; color: #475569; }
.form-item input, .form-item select { padding: 10px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; background: #fff; transition: all 0.2s; }
.form-item input:focus, .form-item select:focus { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.required { color: #ef4444; }
.url-input-wrap { position: relative; }
.url-input-wrap input { width: 100%; box-sizing: border-box; }
.url-hint { display: block; font-size: 11px; color: #94a3b8; margin-top: 4px; }

/* Switch */
.switch-row { display: flex; gap: 28px; }
.switch-label { display: flex; align-items: center; gap: 10px; cursor: pointer; font-size: 14px; color: #475569; }
.switch-label input { display: none; }
.switch-slider { width: 36px; height: 20px; background: #e2e8f0; border-radius: 10px; position: relative; transition: all 0.2s; }
.switch-slider::after { content: ''; position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; background: #fff; border-radius: 50%; transition: all 0.2s; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.switch-label input:checked + .switch-slider { background: #4f46e5; }
.switch-label input:checked + .switch-slider::after { left: 18px; }

/* Cluster binding */
.hint-mini { font-size: 11px; color: #94a3b8; font-weight: 400; margin-left: 4px; }
.cluster-badge-preview { display: flex; align-items: center; min-height: 40px; }
.cluster-chip { display: inline-flex; align-items: center; gap: 4px; padding: 6px 12px; border-radius: 8px; font-size: 13px; font-weight: 500; }
.cluster-chip.global { background: linear-gradient(135deg, #f0fdfa, #ccfbf1); color: #0f766e; border: 1px solid #99f6e4; }
.cluster-chip.bound { background: linear-gradient(135deg, #eff6ff, #dbeafe); color: #1d4ed8; border: 1px solid #bfdbfe; }
.ds-cluster-tag { padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 500; }
.ds-cluster-tag.global { background: #ccfbf1; color: #0f766e; }
.ds-cluster-tag.bound { background: #dbeafe; color: #1d4ed8; }

/* Delete */
.delete-info p { margin: 0 0 12px; font-size: 15px; color: #334155; }
.delete-warning { display: flex; gap: 8px; padding: 12px; background: #fefce8; border: 1px solid #fef08a; border-radius: 8px; font-size: 13px; color: #854d0e; }

/* Animation */
.icon { display: inline-block; transition: transform 0.3s; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>

<template>
  <div class="audit-log-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <span class="title-icon">🛡️</span>
          操作审计日志
        </h1>
        <p class="page-desc">记录平台所有用户操作，支持多维度筛选与合规审计</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-ghost" @click="showRetentionModal = true">
          <span class="btn-icon">⏱️</span> 保留策略
        </button>
        <button class="btn btn-ghost" @click="handleExport" :disabled="exporting">
          <span class="btn-icon">📤</span> {{ exporting ? '导出中...' : '导出' }}
        </button>
        <button class="btn btn-primary" @click="loadData" :disabled="loading">
          <span class="btn-icon">🔄</span> 刷新
        </button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ stats.total_today || 0 }}</div>
        <div class="stat-label">今日操作</div>
        <div class="stat-trend up">📈</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.total_week || 0 }}</div>
        <div class="stat-label">本周操作</div>
        <div class="stat-trend">📊</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ formatNumber(stats.total_all || 0) }}</div>
        <div class="stat-label">累计操作</div>
        <div class="stat-trend">📋</div>
      </div>
      <div class="stat-card">
        <div class="stat-value success">{{ (stats.success_rate || 0).toFixed(1) }}%</div>
        <div class="stat-label">操作成功率</div>
        <div class="stat-trend">✅</div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-group">
        <div class="filter-item">
          <input v-model="filters.keyword" type="text" placeholder="🔍 搜索操作内容..." 
                 class="filter-input search-input" @keyup.enter="handleSearch" />
        </div>
        <div class="filter-item">
          <select v-model="filters.module" class="filter-select" @change="handleSearch">
            <option value="">所有模块</option>
            <option value="auth">🔐 认证</option>
            <option value="cluster">☸️ 集群</option>
            <option value="workload">📦 工作负载</option>
            <option value="network">🌐 网络</option>
            <option value="config">⚙️ 配置</option>
            <option value="storage">💾 存储</option>
            <option value="cicd">🚀 CI/CD</option>
            <option value="rbac">🛡️ 权限</option>
            <option value="platform">🏛️ 平台</option>
            <option value="ai">🤖 AI</option>
            <option value="monitoring">📡 监控</option>
            <option value="image">🐳 镜像</option>
          </select>
        </div>
        <div class="filter-item">
          <select v-model="filters.action" class="filter-select" @change="handleSearch">
            <option value="">所有操作</option>
            <option value="create">➕ 创建</option>
            <option value="update">✏️ 更新</option>
            <option value="delete">🗑️ 删除</option>
            <option value="login">🔑 登录</option>
            <option value="logout">🚪 登出</option>
            <option value="deploy">🚀 部署</option>
            <option value="approve">✅ 审批</option>
            <option value="exec">⚡ 执行</option>
          </select>
        </div>
        <div class="filter-item">
          <select v-model="filters.status" class="filter-select" @change="handleSearch">
            <option value="">所有状态</option>
            <option value="success">✅ 成功</option>
            <option value="failed">❌ 失败</option>
          </select>
        </div>
        <div class="filter-item">
          <select v-model="timeRange" class="filter-select" @change="handleTimeRangeChange">
            <option value="3600">最近 1 小时</option>
            <option value="21600">最近 6 小时</option>
            <option value="86400">最近 24 小时</option>
            <option value="259200">最近 3 天</option>
            <option value="604800">最近 7 天</option>
            <option value="2592000">最近 30 天</option>
            <option value="0">全部时间</option>
          </select>
        </div>
      </div>
      <div class="filter-actions">
        <button class="btn btn-text" @click="resetFilters">重置筛选</button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrapper">
      <div v-if="loading && logs.length === 0" class="loading-state">
        <div class="loading-spinner"></div>
        <p>加载审计数据中...</p>
      </div>

      <table v-else class="audit-table">
        <thead>
          <tr>
            <th class="col-time">操作时间</th>
            <th class="col-user">操作人</th>
            <th class="col-module">模块</th>
            <th class="col-action">操作</th>
            <th class="col-target">操作目标</th>
            <th class="col-status">状态</th>
            <th class="col-ip">IP</th>
            <th class="col-duration">耗时</th>
            <th class="col-detail">详情</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" :class="{ 'row-failed': log.status === 'failed' }">
            <td class="col-time">
              <span class="time-text">{{ formatTime(log.created_at) }}</span>
            </td>
            <td class="col-user">
              <div class="user-cell">
                <span class="user-avatar">{{ (log.username || '?')[0].toUpperCase() }}</span>
                <span class="user-name">{{ log.username || '-' }}</span>
              </div>
            </td>
            <td class="col-module">
              <span :class="['module-badge', `module-${log.module}`]">
                {{ getModuleIcon(log.module) }} {{ getModuleLabel(log.module) }}
              </span>
            </td>
            <td class="col-action">
              <span :class="['action-tag', `action-${log.action}`]">
                {{ log.action_display || log.action }}
              </span>
            </td>
            <td class="col-target">
              <span class="target-text" :title="log.target_name || log.request_uri">
                {{ log.target_name || truncatePath(log.request_uri) || '-' }}
              </span>
            </td>
            <td class="col-status">
              <span :class="['status-dot', log.status]"></span>
              <span :class="['status-text', log.status]">{{ log.status === 'success' ? '成功' : '失败' }}</span>
            </td>
            <td class="col-ip">
              <span class="ip-text">{{ log.user_ip || '-' }}</span>
            </td>
            <td class="col-duration">
              <span class="duration-text">{{ log.duration_ms ? log.duration_ms + 'ms' : '-' }}</span>
            </td>
            <td class="col-detail">
              <button class="btn-detail" @click="showDetail(log)">查看</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="!loading && logs.length === 0" class="empty-state">
        <div class="empty-icon">📭</div>
        <div class="empty-title">暂无审计记录</div>
        <div class="empty-desc">当用户执行操作后，审计日志将自动记录在此</div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination-bar">
      <div class="pagination-info">
        共 <strong>{{ total }}</strong> 条记录，当前第 {{ page }} / {{ Math.ceil(total / pageSize) }} 页
      </div>
      <div class="pagination-controls">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">‹ 上一页</button>
        <span class="page-current">{{ page }}</span>
        <button class="page-btn" :disabled="page >= Math.ceil(total / pageSize)" @click="changePage(page + 1)">下一页 ›</button>
        <select v-model="pageSize" class="page-size-select" @change="handleSearch">
          <option :value="20">20条/页</option>
          <option :value="50">50条/页</option>
          <option :value="100">100条/页</option>
        </select>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <div v-if="detailVisible" class="modal-overlay" @click.self="detailVisible = false">
      <div class="modal-content detail-modal">
        <div class="modal-header">
          <h3>📋 审计日志详情</h3>
          <button class="modal-close" @click="detailVisible = false">✕</button>
        </div>
        <div class="modal-body" v-if="detailLog">
          <div class="detail-grid">
            <div class="detail-item">
              <label>操作时间</label>
              <span>{{ formatTimeFull(detailLog.created_at) }}</span>
            </div>
            <div class="detail-item">
              <label>操作人</label>
              <span>{{ detailLog.username }} (ID: {{ detailLog.user_id }})</span>
            </div>
            <div class="detail-item">
              <label>来源IP</label>
              <span>{{ detailLog.user_ip || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>操作模块</label>
              <span>{{ getModuleLabel(detailLog.module) }}</span>
            </div>
            <div class="detail-item">
              <label>操作类型</label>
              <span>{{ detailLog.action_display || detailLog.action }}</span>
            </div>
            <div class="detail-item">
              <label>操作状态</label>
              <span :class="detailLog.status">{{ detailLog.status === 'success' ? '✅ 成功' : '❌ 失败' }}</span>
            </div>
            <div class="detail-item full">
              <label>请求路径</label>
              <span class="mono">{{ detailLog.request_method }} {{ detailLog.request_uri }}</span>
            </div>
            <div class="detail-item">
              <label>目标类型</label>
              <span>{{ detailLog.target_type || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>目标名称</label>
              <span>{{ detailLog.target_name || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>响应码</label>
              <span>{{ detailLog.response_code || '-' }}</span>
            </div>
            <div class="detail-item">
              <label>耗时</label>
              <span>{{ detailLog.duration_ms }}ms</span>
            </div>
            <div v-if="detailLog.cluster_name" class="detail-item">
              <label>集群</label>
              <span>{{ detailLog.cluster_name }}</span>
            </div>
            <div v-if="detailLog.namespace" class="detail-item">
              <label>命名空间</label>
              <span>{{ detailLog.namespace }}</span>
            </div>
            <div v-if="detailLog.error_message" class="detail-item full">
              <label>错误信息</label>
              <span class="error-text">{{ detailLog.error_message }}</span>
            </div>
            <div v-if="detailLog.request_body" class="detail-item full">
              <label>请求体</label>
              <pre class="code-block">{{ formatJSON(detailLog.request_body) }}</pre>
            </div>
            <div class="detail-item full">
              <label>User-Agent</label>
              <span class="mono small">{{ detailLog.user_agent || '-' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 保留策略弹窗 -->
    <div v-if="showRetentionModal" class="modal-overlay" @click.self="showRetentionModal = false">
      <div class="modal-content retention-modal">
        <div class="modal-header">
          <h3>⏱️ 审计日志保留策略</h3>
          <button class="modal-close" @click="showRetentionModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="retention-info">
            <p class="retention-desc">
              设置审计日志的保留时间。超过保留期限的日志将被自动清理。
              选择「永久保留」则所有日志永不删除，适用于合规审计要求。
            </p>
          </div>
          <div class="retention-form">
            <div class="form-group">
              <label class="form-label">保留模式</label>
              <div class="radio-group">
                <label class="radio-item" :class="{ active: !retentionForm.is_permanent }">
                  <input type="radio" :value="false" v-model="retentionForm.is_permanent" />
                  <span class="radio-icon">📅</span>
                  <span>按天数保留</span>
                </label>
                <label class="radio-item" :class="{ active: retentionForm.is_permanent }">
                  <input type="radio" :value="true" v-model="retentionForm.is_permanent" />
                  <span class="radio-icon">♾️</span>
                  <span>永久保留</span>
                </label>
              </div>
            </div>
            <div v-if="!retentionForm.is_permanent" class="form-group">
              <label class="form-label">保留天数</label>
              <div class="retention-presets">
                <button v-for="preset in retentionPresets" :key="preset.value"
                  :class="['preset-btn', { active: retentionForm.retention_days === preset.value }]"
                  @click="retentionForm.retention_days = preset.value">
                  {{ preset.label }}
                </button>
              </div>
              <div class="custom-days">
                <input type="number" v-model.number="retentionForm.retention_days" 
                       min="1" max="3650" class="days-input" />
                <span class="days-suffix">天</span>
              </div>
            </div>
            <div v-if="retentionForm.is_permanent" class="permanent-note">
              <span class="note-icon">ℹ️</span>
              永久保留模式下，审计日志将永不自动删除。请确保有足够的数据库存储空间。
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="showRetentionModal = false">取消</button>
          <button class="btn btn-danger" @click="handleCleanup" :disabled="cleaningUp">
            {{ cleaningUp ? '清理中...' : '手动清理' }}
          </button>
          <button class="btn btn-primary" @click="handleSaveRetention" :disabled="savingRetention">
            {{ savingRetention ? '保存中...' : '保存策略' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { Message } from '@arco-design/web-vue'
import {
  getAuditLogs,
  getAuditStatistics,
  getAuditRetention,
  updateAuditRetention,
  cleanupAuditLogs,
  exportAuditLogs
} from '@/api/platform/audit'

// Props支持（用于嵌入审计中心时按模块筛选）
const props = defineProps({
  filterModule: {
    type: String,
    default: ''
  }
})

// ========== 状态 ==========
const loading = ref(true)
const exporting = ref(false)
const logs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const timeRange = ref('86400')

// 统计
const stats = ref({})

// 筛选
const filters = reactive({
  keyword: '',
  module: props.filterModule || '',
  action: '',
  status: '',
  username: ''
})

// 监听 prop 变化自动更新筛选
watch(() => props.filterModule, (val) => {
  if (val) {
    filters.module = val
    handleSearch()
  }
})

// 详情弹窗
const detailVisible = ref(false)
const detailLog = ref(null)

// 保留策略
const showRetentionModal = ref(false)
const savingRetention = ref(false)
const cleaningUp = ref(false)
const retentionForm = reactive({
  retention_days: 30,
  is_permanent: false
})
const retentionPresets = [
  { label: '7天', value: 7 },
  { label: '30天', value: 30 },
  { label: '90天', value: 90 },
  { label: '180天', value: 180 },
  { label: '365天', value: 365 },
  { label: '3年', value: 1095 },
]

// ========== 方法 ==========
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      ...filters,
    }
    // 时间范围
    if (timeRange.value && timeRange.value !== '0') {
      params.start_time = Math.floor(Date.now() / 1000) - parseInt(timeRange.value)
    }

    const res = await getAuditLogs(params)
    if (res.code === 0 && res.data) {
      logs.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (e) {
    console.error('加载审计日志失败:', e)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const res = await getAuditStatistics()
    if (res.code === 0 && res.data) {
      stats.value = res.data
    }
  } catch (e) {
    console.error('加载统计失败:', e)
  }
}

const loadRetention = async () => {
  try {
    const res = await getAuditRetention()
    if (res.code === 0 && res.data) {
      retentionForm.retention_days = res.data.retention_days || 30
      retentionForm.is_permanent = res.data.is_permanent || false
    }
  } catch (e) {
    console.error('加载保留策略失败:', e)
  }
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const handleTimeRangeChange = () => {
  page.value = 1
  loadData()
}

const changePage = (p) => {
  page.value = p
  loadData()
}

const resetFilters = () => {
  filters.keyword = ''
  filters.module = ''
  filters.action = ''
  filters.status = ''
  filters.username = ''
  timeRange.value = '86400'
  page.value = 1
  loadData()
}

const showDetail = (log) => {
  detailLog.value = log
  detailVisible.value = true
}

const handleSaveRetention = async () => {
  savingRetention.value = true
  try {
    const res = await updateAuditRetention({
      retention_days: retentionForm.is_permanent ? 0 : retentionForm.retention_days,
      is_permanent: retentionForm.is_permanent
    })
    if (res.code === 0) {
      Message.success({ content: '保留策略已保存' })
      showRetentionModal.value = false
    } else {
      Message.error({ content: res.msg || '保存失败' })
    }
  } catch (e) {
    Message.error({ content: '保存失败: ' + (e.msg || e.message) })
  } finally {
    savingRetention.value = false
  }
}

const handleCleanup = async () => {
  if (retentionForm.is_permanent) {
    Message.warning({ content: '永久保留模式下无需清理' })
    return
  }
  cleaningUp.value = true
  try {
    const res = await cleanupAuditLogs()
    if (res.code === 0) {
      Message.success({ content: `清理完成，已删除 ${res.data?.affected || 0} 条过期日志` })
      loadData()
      loadStats()
    }
  } catch (e) {
    Message.error({ content: '清理失败: ' + (e.msg || e.message) })
  } finally {
    cleaningUp.value = false
  }
}

const handleExport = async () => {
  exporting.value = true
  try {
    const params = { ...filters }
    if (timeRange.value && timeRange.value !== '0') {
      params.start_time = Math.floor(Date.now() / 1000) - parseInt(timeRange.value)
    }
    const res = await exportAuditLogs(params)
    if (res.code === 0 && res.data?.list) {
      // 生成 CSV 并下载
      const csvContent = generateCSV(res.data.list)
      downloadCSV(csvContent, `audit_log_${Date.now()}.csv`)
      Message.success({ content: `已导出 ${res.data.total} 条记录` })
    }
  } catch (e) {
    Message.error({ content: '导出失败' })
  } finally {
    exporting.value = false
  }
}

// ========== 工具函数 ==========
const formatTime = (ts) => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${mi}:${ss}`
}

const formatTimeFull = (ts) => {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

const formatNumber = (n) => {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return n
}

const formatJSON = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

const truncatePath = (path) => {
  if (!path) return ''
  if (path.length > 40) return path.slice(0, 40) + '...'
  return path
}

const getModuleIcon = (module) => {
  const map = {
    auth: '🔐', cluster: '☸️', workload: '📦', network: '🌐',
    config: '⚙️', storage: '💾', cicd: '🚀', rbac: '🛡️',
    platform: '🏛️', ai: '🤖', monitoring: '📡', image: '🐳'
  }
  return map[module] || '📋'
}

const getModuleLabel = (module) => {
  const map = {
    auth: '认证', cluster: '集群', workload: '工作负载', network: '网络',
    config: '配置', storage: '存储', cicd: 'CI/CD', rbac: '权限',
    platform: '平台', ai: 'AI', monitoring: '监控', image: '镜像'
  }
  return map[module] || module
}

const generateCSV = (list) => {
  const headers = ['时间', '用户', '模块', '操作', '目标', '状态', 'IP', '耗时(ms)', '请求路径']
  const rows = list.map(l => [
    formatTimeFull(l.created_at),
    l.username,
    getModuleLabel(l.module),
    l.action_display || l.action,
    l.target_name || '',
    l.status,
    l.user_ip,
    l.duration_ms,
    l.request_uri
  ])
  return [headers, ...rows].map(r => r.join(',')).join('\n')
}

const downloadCSV = (content, filename) => {
  const blob = new Blob(['\ufeff' + content], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
}

// ========== 生命周期 ==========
onMounted(() => {
  loadData()
  loadStats()
  loadRetention()
})
</script>

<style scoped>
.audit-log-page { padding: 0; }

/* 页面头部 */
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.page-title { font-size: 26px; font-weight: 700; color: #1e293b; margin: 0 0 6px; display: flex; align-items: center; gap: 10px; }
.title-icon { font-size: 28px; }
.page-desc { color: #64748b; font-size: 14px; margin: 0; }
.header-actions { display: flex; gap: 10px; }

/* 按钮 */
.btn { padding: 9px 18px; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; display: inline-flex; align-items: center; gap: 6px; transition: all 0.2s; }
.btn-primary { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); color: #fff; box-shadow: 0 2px 8px rgba(59,130,246,0.3); }
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
.btn-ghost { background: #f8fafc; color: #475569; border: 1px solid #e2e8f0; }
.btn-ghost:hover { background: #f1f5f9; border-color: #cbd5e1; }
.btn-danger { background: #ef4444; color: #fff; }
.btn-danger:hover { background: #dc2626; }
.btn-text { background: none; border: none; color: #3b82f6; cursor: pointer; font-size: 13px; }
.btn-text:hover { text-decoration: underline; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; transform: none !important; }
.btn-icon { font-size: 15px; }

/* 统计卡片 */
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.stat-card { background: #fff; border-radius: 12px; padding: 20px 24px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); border: 1px solid #f1f5f9; position: relative; overflow: hidden; }
.stat-card::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px; background: linear-gradient(90deg, #3b82f6, #8b5cf6); }
.stat-value { font-size: 28px; font-weight: 700; color: #1e293b; margin-bottom: 4px; }
.stat-value.success { color: #16a34a; }
.stat-label { font-size: 13px; color: #64748b; }
.stat-trend { position: absolute; top: 16px; right: 16px; font-size: 20px; }

/* 筛选栏 */
.filter-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; padding: 16px 20px; background: #fff; border-radius: 12px; border: 1px solid #f1f5f9; }
.filter-group { display: flex; gap: 10px; flex-wrap: wrap; flex: 1; }
.filter-item { }
.filter-input, .filter-select { padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 13px; background: #f8fafc; transition: all 0.2s; min-width: 130px; }
.filter-input:focus, .filter-select:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59,130,246,0.1); background: #fff; }
.search-input { min-width: 220px; }

/* 表格 */
.table-wrapper { background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,0.04); border: 1px solid #f1f5f9; }
.audit-table { width: 100%; border-collapse: collapse; }
.audit-table th { padding: 13px 14px; background: linear-gradient(135deg, #f8fafc, #f1f5f9); font-weight: 600; font-size: 13px; color: #475569; text-align: left; border-bottom: 1px solid #e2e8f0; white-space: nowrap; }
.audit-table td { padding: 12px 14px; border-bottom: 1px solid #f1f5f9; font-size: 13px; color: #334155; }
.audit-table tr:hover { background: #f8fafc; }
.audit-table tr.row-failed { background: #fef2f2; }
.audit-table tr.row-failed:hover { background: #fee2e2; }

.col-time { width: 120px; }
.col-user { width: 120px; }
.col-module { width: 100px; }
.col-action { width: 120px; }
.col-status { width: 80px; }
.col-ip { width: 110px; }
.col-duration { width: 70px; }
.col-detail { width: 60px; }

.time-text { color: #64748b; font-size: 12px; font-family: 'JetBrains Mono', monospace; }
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 26px; height: 26px; border-radius: 50%; background: linear-gradient(135deg, #3b82f6, #8b5cf6); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 600; }
.user-name { font-weight: 500; }

.module-badge { display: inline-block; padding: 3px 10px; border-radius: 6px; font-size: 12px; font-weight: 500; background: #f1f5f9; color: #475569; white-space: nowrap; }
.module-auth { background: #fef3c7; color: #92400e; }
.module-cluster { background: #dbeafe; color: #1e40af; }
.module-workload { background: #e0e7ff; color: #3730a3; }
.module-cicd { background: #dcfce7; color: #166534; }
.module-rbac { background: #fce7f3; color: #9d174d; }
.module-platform { background: #f3e8ff; color: #6b21a8; }
.module-ai { background: #cffafe; color: #155e75; }

.action-tag { display: inline-block; padding: 3px 8px; border-radius: 4px; font-size: 12px; background: #f1f5f9; }
.action-create { background: #dcfce7; color: #166534; }
.action-update { background: #dbeafe; color: #1e40af; }
.action-delete { background: #fee2e2; color: #991b1b; }
.action-login { background: #fef3c7; color: #92400e; }
.action-deploy { background: #e0e7ff; color: #3730a3; }

.status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
.status-dot.success { background: #16a34a; }
.status-dot.failed { background: #dc2626; }
.status-text.success { color: #16a34a; }
.status-text.failed { color: #dc2626; }

.ip-text { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #64748b; }
.duration-text { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: #64748b; }
.target-text { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: inline-block; }

.btn-detail { background: none; border: 1px solid #e2e8f0; border-radius: 6px; padding: 4px 10px; font-size: 12px; color: #3b82f6; cursor: pointer; }
.btn-detail:hover { background: #eff6ff; border-color: #3b82f6; }

/* 加载/空状态 */
.loading-state { display: flex; flex-direction: column; align-items: center; padding: 60px; color: #64748b; }
.loading-spinner { width: 36px; height: 36px; border: 3px solid #e2e8f0; border-top-color: #3b82f6; border-radius: 50%; animation: spin 0.7s linear infinite; margin-bottom: 12px; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-state { padding: 60px; text-align: center; }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-title { font-size: 16px; font-weight: 600; color: #334155; margin-bottom: 6px; }
.empty-desc { font-size: 13px; color: #94a3b8; }

/* 分页 */
.pagination-bar { display: flex; justify-content: space-between; align-items: center; padding: 14px 20px; background: #f8fafc; border-top: 1px solid #f1f5f9; }
.pagination-info { font-size: 13px; color: #64748b; }
.pagination-info strong { color: #1e293b; }
.pagination-controls { display: flex; align-items: center; gap: 8px; }
.page-btn { padding: 6px 14px; border: 1px solid #e2e8f0; border-radius: 6px; background: #fff; font-size: 13px; cursor: pointer; }
.page-btn:hover:not(:disabled) { background: #f1f5f9; }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-current { padding: 6px 12px; background: #3b82f6; color: #fff; border-radius: 6px; font-size: 13px; font-weight: 600; }
.page-size-select { padding: 6px 10px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; }

/* 弹窗 */
.modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-content { background: #fff; border-radius: 16px; width: 90%; max-width: 700px; max-height: 85vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.15); }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #f1f5f9; }
.modal-header h3 { margin: 0; font-size: 18px; font-weight: 600; }
.modal-close { background: none; border: none; font-size: 20px; cursor: pointer; color: #94a3b8; width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.modal-close:hover { background: #f1f5f9; color: #475569; }
.modal-body { padding: 24px; }
.modal-footer { padding: 16px 24px; border-top: 1px solid #f1f5f9; display: flex; justify-content: flex-end; gap: 10px; }

/* 详情弹窗 */
.detail-modal { max-width: 800px; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.detail-item.full { grid-column: 1 / -1; }
.detail-item label { font-size: 12px; color: #94a3b8; font-weight: 500; text-transform: uppercase; }
.detail-item span { font-size: 14px; color: #1e293b; }
.detail-item .mono { font-family: 'JetBrains Mono', monospace; font-size: 13px; }
.detail-item .small { font-size: 11px; color: #64748b; word-break: break-all; }
.detail-item .error-text { color: #dc2626; }
.code-block { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 12px; font-size: 12px; font-family: 'JetBrains Mono', monospace; max-height: 200px; overflow: auto; white-space: pre-wrap; word-break: break-all; }

/* 保留策略弹窗 */
.retention-modal { max-width: 560px; }
.retention-desc { color: #64748b; font-size: 14px; line-height: 1.6; margin: 0 0 20px; }
.retention-form { }
.form-group { margin-bottom: 20px; }
.form-label { display: block; font-size: 14px; font-weight: 600; color: #334155; margin-bottom: 10px; }
.radio-group { display: flex; gap: 12px; }
.radio-item { display: flex; align-items: center; gap: 8px; padding: 12px 20px; border: 2px solid #e2e8f0; border-radius: 10px; cursor: pointer; transition: all 0.2s; flex: 1; }
.radio-item:hover { border-color: #94a3b8; }
.radio-item.active { border-color: #3b82f6; background: #eff6ff; }
.radio-item input { display: none; }
.radio-icon { font-size: 20px; }
.retention-presets { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
.preset-btn { padding: 8px 16px; border: 1px solid #e2e8f0; border-radius: 8px; background: #f8fafc; font-size: 13px; cursor: pointer; transition: all 0.2s; }
.preset-btn:hover { border-color: #3b82f6; }
.preset-btn.active { background: #3b82f6; color: #fff; border-color: #3b82f6; }
.custom-days { display: flex; align-items: center; gap: 8px; }
.days-input { width: 100px; padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; }
.days-suffix { color: #64748b; font-size: 14px; }
.permanent-note { display: flex; align-items: flex-start; gap: 8px; padding: 14px 16px; background: #f0f9ff; border-radius: 8px; border: 1px solid #bae6fd; color: #0369a1; font-size: 13px; line-height: 1.5; }
.note-icon { font-size: 16px; flex-shrink: 0; }

/* 响应式 */
@media (max-width: 1200px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr; }
  .page-header { flex-direction: column; gap: 16px; }
  .filter-bar { flex-direction: column; gap: 12px; }
  .detail-grid { grid-template-columns: 1fr; }
}
</style>

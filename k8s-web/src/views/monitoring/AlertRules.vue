<template>
  <div class="alert-rules-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h3>告警规则</h3>
        <span class="header-desc">配置基于 PromQL 的告警规则，支持多级别通知</span>
      </div>
      <button class="btn-primary" @click="openDialog()">
        <span>+</span> 新增规则
      </button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <input v-model="filters.keyword" placeholder="搜索规则名称..." class="search-input" @input="debouncedLoad" />
      <select v-model="filters.severity" @change="loadList" class="filter-select">
        <option value="">全部级别</option>
        <option value="critical">Critical</option>
        <option value="warning">Warning</option>
        <option value="info">Info</option>
      </select>
      <select v-model="filters.group" @change="loadList" class="filter-select">
        <option value="">全部分组</option>
        <option v-for="g in groups" :key="g" :value="g">{{ g }}</option>
      </select>
    </div>

    <!-- 规则列表 -->
    <div class="rules-table-wrapper" v-if="list.length">
      <table class="data-table">
        <thead>
          <tr>
            <th>规则名称</th>
            <th>分组</th>
            <th>级别</th>
            <th>PromQL 表达式</th>
            <th>持续时间</th>
            <th>评估状态</th>
            <th>启用</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in list" :key="rule.id" :class="{ disabled: !rule.enabled }">
            <td class="rule-name">{{ rule.name }}</td>
            <td><span class="group-tag">{{ rule.group }}</span></td>
            <td><span class="severity-badge" :class="rule.severity">{{ rule.severity }}</span></td>
            <td class="expr-cell"><code>{{ truncateExpr(rule.expr) }}</code></td>
            <td>{{ rule.duration }}</td>
            <td>
              <span class="eval-status" :class="rule.last_eval_result || 'unknown'">
                {{ evalStatusMap[rule.last_eval_result] || '未评估' }}
              </span>
            </td>
            <td>
              <label class="toggle-switch">
                <input type="checkbox" :checked="rule.enabled" @change="handleToggle(rule)" />
                <span class="toggle-slider"></span>
              </label>
            </td>
            <td class="action-cell">
              <button class="btn-icon" @click="openDialog(rule)" title="编辑">✏️</button>
              <button class="btn-icon danger" @click="confirmDelete(rule)" title="删除">🗑️</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 空状态 -->
    <div class="empty-state" v-else>
      <div class="empty-icon">📋</div>
      <h3>暂无告警规则</h3>
      <p>创建告警规则后，系统将自动检测异常指标并触发通知</p>
      <div class="template-section">
        <h4>快速创建（常用模板）</h4>
        <div class="template-grid">
          <button class="template-card" v-for="tpl in templates" :key="tpl.name" @click="openDialogFromTemplate(tpl)">
            <span class="tpl-icon">{{ tpl.icon }}</span>
            <span class="tpl-name">{{ tpl.name }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 规则模板快速创建区（有数据时也显示） -->
    <div class="template-section" v-if="list.length">
      <h4>📌 快速添加常用规则</h4>
      <div class="template-grid">
        <button class="template-card" v-for="tpl in templates" :key="tpl.name" @click="openDialogFromTemplate(tpl)">
          <span class="tpl-icon">{{ tpl.icon }}</span>
          <span class="tpl-name">{{ tpl.name }}</span>
        </button>
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <div class="modal-overlay" v-if="dialogVisible" @click.self="dialogVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>{{ editingId ? '编辑告警规则' : '新增告警规则' }}</h3>
          <button class="modal-close" @click="dialogVisible = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row-group">
            <div class="form-row half">
              <label>规则名称 <span class="required">*</span></label>
              <input v-model="form.name" placeholder="如: 节点CPU使用率过高" />
            </div>
            <div class="form-row half">
              <label>分组</label>
              <input v-model="form.group" placeholder="default" />
            </div>
          </div>
          <div class="form-row-group">
            <div class="form-row half">
              <label>严重级别 <span class="required">*</span></label>
              <select v-model="form.severity">
                <option value="critical">🔴 Critical</option>
                <option value="warning">🟡 Warning</option>
                <option value="info">🔵 Info</option>
              </select>
            </div>
            <div class="form-row half">
              <label>持续时间 (for)</label>
              <input v-model="form.duration" placeholder="5m" />
            </div>
          </div>
          <div class="form-row">
            <label>PromQL 表达式 <span class="required">*</span></label>
            <textarea v-model="form.expr" rows="3" placeholder='如: 100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80'></textarea>
          </div>
          <div class="form-row">
            <label>告警摘要</label>
            <input v-model="form.summary" placeholder="告警摘要模板，支持 {{ $labels.instance }}" />
          </div>
          <div class="form-row">
            <label>告警描述</label>
            <textarea v-model="form.description" rows="2" placeholder="详细描述"></textarea>
          </div>
          <div class="form-row-group">
            <div class="form-row half">
              <label>通知渠道</label>
              <select v-model="form.notify_channels">
                <option value="">无</option>
                <option v-for="ch in notifyChannelList" :key="ch.id" :value="ch.type + ':' + ch.id">
                  {{ ch.name }} ({{ ch.type }})
                </option>
              </select>
              <span class="form-hint">前往「通知渠道」页配置更多渠道</span>
            </div>
            <div class="form-row half">
              <label>评估间隔(秒)</label>
              <input v-model.number="form.eval_interval" type="number" min="10" max="600" />
            </div>
          </div>
          <div class="form-row" v-if="form.notify_channels === 'webhook'">
            <label>Webhook URL</label>
            <input v-model="form.notify_url" placeholder="https://..." />
          </div>
          <div class="form-row-inline">
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.enabled" /> 立即启用
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="dialogVisible = false">取消</button>
          <button class="btn-primary" @click="submitForm" :disabled="submitting">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-dialog modal-sm">
        <div class="modal-header"><h3>确认删除</h3><button class="modal-close" @click="deleteTarget = null">×</button></div>
        <div class="modal-body">
          <p>确定删除告警规则 <b>{{ deleteTarget.name }}</b>？</p>
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
import { ref, reactive, onMounted } from 'vue'
import {
  listAlertRules, createAlertRule, updateAlertRule, deleteAlertRule,
  toggleAlertRule, getAlertRuleGroups, listNotifyChannels,
} from '@/api/monitoring'

const list = ref([])
const groups = ref([])
const notifyChannelList = ref([])
const dialogVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const deleteTarget = ref(null)
const filters = reactive({ keyword: '', severity: '', group: '' })

const evalStatusMap = { normal: '正常', firing: '告警中', pending: '待触发', error: '异常', unknown: '未评估' }

const templates = [
  { name: 'CPU > 80%', icon: '💻', expr: '100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80', severity: 'warning', duration: '5m', summary: '集群CPU使用率超过80%' },
  { name: '内存 > 85%', icon: '🧠', expr: 'avg(100 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100)) > 85', severity: 'warning', duration: '5m', summary: '集群内存使用率超过85%' },
  { name: '磁盘 > 90%', icon: '💾', expr: 'avg(100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} * 100)) > 90', severity: 'critical', duration: '10m', summary: '磁盘使用率超过90%' },
  { name: 'Pod 重启', icon: '🔄', expr: 'increase(kube_pod_container_status_restarts_total[1h]) > 3', severity: 'warning', duration: '1m', summary: 'Pod 1小时内重启超过3次' },
  { name: '节点宕机', icon: '🖥️', expr: 'up{job="node-exporter"} == 0', severity: 'critical', duration: '2m', summary: '节点 {{ $labels.instance }} 不可达' },
  { name: 'API 5xx', icon: '🌐', expr: 'sum(rate(http_requests_total{status=~"5.."}[5m])) > 1', severity: 'critical', duration: '3m', summary: 'HTTP 5xx 错误率升高' },
]

const defaultForm = () => ({
  name: '', group: 'default', severity: 'warning', expr: '', duration: '5m',
  summary: '', description: '', labels: '', annotations: '', enabled: true,
  notify_channels: '', notify_url: '', eval_interval: 60, datasource_id: 0,
})
const form = reactive(defaultForm())

let debounceTimer = null
const debouncedLoad = () => { clearTimeout(debounceTimer); debounceTimer = setTimeout(loadList, 300) }

const truncateExpr = (expr) => expr?.length > 60 ? expr.slice(0, 60) + '...' : expr

async function loadList() {
  try {
    const [res, gRes, chRes] = await Promise.all([
      listAlertRules({ page: 1, size: 50, ...filters }),
      getAlertRuleGroups(),
      listNotifyChannels({}),
    ])
    if (res?.code === 0) list.value = res.data?.items || []
    if (gRes?.code === 0) groups.value = gRes.data || []
    if (chRes?.code === 0) notifyChannelList.value = (chRes.data?.items || []).filter(c => c.enabled)
  } catch {}
}

function openDialog(rule = null) {
  if (rule) {
    editingId.value = rule.id
    Object.assign(form, { ...rule })
  } else {
    editingId.value = null
    Object.assign(form, defaultForm())
  }
  dialogVisible.value = true
}

function openDialogFromTemplate(tpl) {
  editingId.value = null
  Object.assign(form, { ...defaultForm(), name: tpl.name, expr: tpl.expr, severity: tpl.severity, duration: tpl.duration, summary: tpl.summary })
  dialogVisible.value = true
}

async function submitForm() {
  if (!form.name || !form.expr || !form.severity) return alert('请填写必填字段')
  submitting.value = true
  try {
    if (editingId.value) {
      await updateAlertRule(editingId.value, form)
    } else {
      await createAlertRule(form)
    }
    dialogVisible.value = false
    loadList()
  } catch (e) {
    alert('保存失败: ' + (e?.msg || e?.message || ''))
  } finally { submitting.value = false }
}

async function handleToggle(rule) {
  try {
    await toggleAlertRule(rule.id, !rule.enabled)
    loadList()
  } catch {}
}

function confirmDelete(rule) { deleteTarget.value = rule }

async function doDelete() {
  try {
    await deleteAlertRule(deleteTarget.value.id)
    deleteTarget.value = null
    loadList()
  } catch {}
}

onMounted(loadList)
</script>

<style scoped>
.alert-rules-page { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-left h3 { margin: 0; font-size: 20px; color: #1a202c; }
.header-desc { font-size: 13px; color: #718096; }
.btn-primary { background: #4f46e5; color: #fff; border: none; padding: 8px 18px; border-radius: 8px; font-size: 14px; cursor: pointer; font-weight: 500; display: flex; align-items: center; gap: 4px; }
.btn-primary:hover { background: #4338ca; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-input { flex: 1; max-width: 280px; padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #4f46e5; }
.filter-select { padding: 8px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; background: #fff; }

.rules-table-wrapper { background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: #f7f8fa; padding: 12px 16px; text-align: left; font-weight: 600; color: #4a5568; border-bottom: 1px solid #e8ecf0; }
.data-table td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; }
.data-table tr:hover { background: #f7fafc; }
.data-table tr.disabled { opacity: 0.5; }
.rule-name { font-weight: 500; color: #1a202c; }
.group-tag { background: #edf2f7; color: #4a5568; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.severity-badge { padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; }
.severity-badge.critical { background: #fef2f2; color: #dc2626; }
.severity-badge.warning { background: #fffbeb; color: #d97706; }
.severity-badge.info { background: #eff6ff; color: #2563eb; }
.expr-cell code { font-size: 12px; background: #f7f8fa; padding: 2px 6px; border-radius: 4px; color: #6b7280; }
.eval-status { padding: 3px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; }
.eval-status.normal { background: #ecfdf5; color: #059669; }
.eval-status.firing { background: #fef2f2; color: #dc2626; }
.eval-status.pending { background: #fffbeb; color: #d97706; }
.eval-status.error { background: #fef2f2; color: #dc2626; }
.eval-status.unknown { background: #f3f4f6; color: #6b7280; }

.toggle-switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; cursor: pointer; inset: 0; background: #cbd5e1; border-radius: 20px; transition: 0.3s; }
.toggle-slider::before { content: ''; position: absolute; width: 16px; height: 16px; left: 2px; bottom: 2px; background: #fff; border-radius: 50%; transition: 0.3s; }
.toggle-switch input:checked + .toggle-slider { background: #4f46e5; }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(16px); }

.action-cell { white-space: nowrap; }
.btn-icon { background: none; border: none; font-size: 16px; cursor: pointer; padding: 4px 6px; border-radius: 4px; }
.btn-icon:hover { background: #f1f5f9; }
.btn-icon.danger:hover { background: #fef2f2; }

.template-section { margin-top: 24px; padding: 20px; background: #fff; border-radius: 12px; border: 1px solid #e8ecf0; }
.template-section h4 { margin: 0 0 12px; font-size: 14px; color: #4a5568; }
.template-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 10px; }
.template-card { display: flex; align-items: center; gap: 8px; padding: 10px 14px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 13px; transition: all 0.15s; }
.template-card:hover { border-color: #4f46e5; background: #f5f3ff; }
.tpl-icon { font-size: 18px; }
.tpl-name { color: #4a5568; font-weight: 500; }

.empty-state { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 56px; margin-bottom: 12px; }
.empty-state h3 { color: #1a202c; margin: 0 0 8px; }
.empty-state p { color: #718096; margin: 0 0 24px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 14px; width: 560px; max-height: 85vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.15); }
.modal-lg { width: 680px; }
.modal-sm { width: 420px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #e8ecf0; }
.modal-header h3 { margin: 0; font-size: 18px; }
.modal-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #a0aec0; }
.modal-body { padding: 24px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid #e8ecf0; }
.form-row { margin-bottom: 16px; }
.form-row label { display: block; font-size: 13px; font-weight: 500; color: #4a5568; margin-bottom: 6px; }
.form-row input, .form-row select, .form-row textarea { width: 100%; padding: 9px 14px; border: 1px solid #e2e8f0; border-radius: 8px; font-size: 14px; outline: none; box-sizing: border-box; font-family: inherit; }
.form-row textarea { resize: vertical; font-family: 'Fira Code', monospace, inherit; }
.form-row input:focus, .form-row select:focus, .form-row textarea:focus { border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.form-row-group { display: flex; gap: 16px; }
.form-row.half { flex: 1; }
.form-row-inline { display: flex; gap: 24px; margin-top: 8px; }
.checkbox-label { display: flex; align-items: center; gap: 6px; font-size: 14px; color: #4a5568; cursor: pointer; }
.required { color: #ef4444; }
.btn-outline { padding: 8px 16px; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-size: 14px; }
.btn-outline:hover { border-color: #4f46e5; color: #4f46e5; }
.btn-danger { background: #dc2626; color: #fff; border: none; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-size: 14px; }
.btn-danger:hover { background: #b91c1c; }
</style>

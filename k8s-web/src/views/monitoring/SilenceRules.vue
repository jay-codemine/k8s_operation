<template>
  <div class="silence-rules-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h3>告警降噪</h3>
        <span class="header-desc">通过静默、抑制、聚合规则减少告警风暴，精准触达关键告警</span>
      </div>
      <button class="btn-primary" @click="handleAddRule()">
        <span>+</span> 新增规则
      </button>
    </div>

    <!-- 规则类型 Tab -->
    <div class="rule-type-tabs">
      <button v-for="tab in ruleTabs" :key="tab.value"
        :class="['tab-btn', { active: activeTab === tab.value }]"
        @click="activeTab = tab.value; loadList()">
        <span class="tab-icon">{{ tab.icon }}</span>
        <span class="tab-label">{{ tab.label }}</span>
        <span class="tab-desc">{{ tab.desc }}</span>
      </button>
    </div>

    <!-- 静默规则列表 -->
    <div class="rules-list" v-if="activeTab === 'silence'">
      <div class="rule-card" v-for="rule in list" :key="rule.id" :class="{ expired: isExpired(rule), disabled: !rule.enabled }">
        <div class="rule-card-header">
          <div class="rule-status-indicator" :class="getRuleStatus(rule)"></div>
          <div class="rule-title-area">
            <h4>{{ rule.name }}</h4>
            <span class="rule-status-text">{{ getRuleStatusText(rule) }}</span>
          </div>
          <div class="rule-actions">
            <button class="btn-icon" @click="openDialog(rule)" title="编辑">✏️</button>
            <button class="btn-icon danger" @click="confirmDelete(rule)" title="删除">🗑️</button>
          </div>
        </div>
        <div class="rule-card-body">
          <!-- 匹配条件展示 -->
          <div class="matcher-tags">
            <span class="matcher-tag" v-for="(m, i) in parseMatchers(rule.matchers)" :key="i">
              {{ m.label }} {{ m.op }} {{ m.value }}
            </span>
            <span class="matcher-tag empty" v-if="!parseMatchers(rule.matchers).length">匹配所有告警</span>
          </div>
          <!-- 时间信息 -->
          <div class="rule-time-info">
            <span v-if="rule.starts_at">
              <b>开始:</b> {{ formatTime(rule.starts_at) }}
            </span>
            <span v-if="rule.ends_at">
              <b>结束:</b> {{ formatTime(rule.ends_at) }}
            </span>
            <span v-if="rule.duration && !rule.ends_at">
              <b>持续:</b> {{ rule.duration }}
            </span>
            <span v-if="rule.repeat_type !== 'once'">
              <b>重复:</b> {{ repeatTypeMap[rule.repeat_type] || rule.repeat_type }}
            </span>
          </div>
          <div class="rule-comment" v-if="rule.comment">💬 {{ rule.comment }}</div>
        </div>
      </div>
      <div class="empty-state" v-if="!list.length">
        <div class="empty-icon">🔇</div>
        <h3>暂无静默规则</h3>
        <p>创建静默规则后，匹配的告警将被静默，不再发送通知</p>
      </div>
    </div>

    <!-- 抑制规则列表 -->
    <div class="rules-list" v-if="activeTab === 'inhibit'">
      <div class="rule-card" v-for="rule in inhibitList" :key="rule.id" :class="{ disabled: !rule.enabled }">
        <div class="rule-card-header">
          <div class="rule-status-indicator" :class="rule.enabled ? 'active' : 'disabled'"></div>
          <div class="rule-title-area">
            <h4>{{ rule.name }}</h4>
            <span class="rule-status-text">{{ rule.enabled ? '生效中' : '已停用' }}</span>
          </div>
          <div class="rule-actions">
            <button class="btn-icon" @click="openInhibitDialog(rule)" title="编辑">✏️</button>
            <button class="btn-icon danger" @click="confirmDeleteInhibit(rule)" title="删除">🗑️</button>
          </div>
        </div>
        <div class="rule-card-body">
          <div class="inhibit-flow">
            <div class="inhibit-box source">
              <span class="box-label">源告警 (高优)</span>
              <div class="matcher-tags">
                <span class="matcher-tag" v-for="(m, i) in parseMatchers(rule.source_matchers)" :key="'s'+i">
                  {{ m.label }} {{ m.op }} {{ m.value }}
                </span>
              </div>
            </div>
            <div class="inhibit-arrow">→ 抑制 →</div>
            <div class="inhibit-box target">
              <span class="box-label">目标告警 (低优)</span>
              <div class="matcher-tags">
                <span class="matcher-tag" v-for="(m, i) in parseMatchers(rule.target_matchers)" :key="'t'+i">
                  {{ m.label }} {{ m.op }} {{ m.value }}
                </span>
              </div>
            </div>
          </div>
          <div class="rule-meta" v-if="rule.equal_labels">
            <span class="meta-label">关联标签:</span>
            <span class="label-chip" v-for="l in rule.equal_labels.split(',')" :key="l">{{ l.trim() }}</span>
          </div>
          <div class="rule-comment" v-if="rule.description">💬 {{ rule.description }}</div>
        </div>
      </div>
      <div class="empty-state" v-if="!inhibitList.length">
        <div class="empty-icon">🛡️</div>
        <h3>暂无抑制规则</h3>
        <p>配置后，当高优先级告警触发时，会自动抑制同类低优先级告警通知</p>
      </div>
    </div>

    <!-- 聚合规则列表 -->
    <div class="rules-list" v-if="activeTab === 'aggregate'">
      <div class="rule-card" v-for="rule in aggregateList" :key="rule.id" :class="{ disabled: !rule.enabled }">
        <div class="rule-card-header">
          <div class="rule-status-indicator" :class="rule.enabled ? 'active' : 'disabled'"></div>
          <div class="rule-title-area">
            <h4>{{ rule.name }}</h4>
            <span class="rule-status-text">{{ rule.enabled ? '生效中' : '已停用' }}</span>
          </div>
          <div class="rule-actions">
            <button class="btn-icon" @click="openAggregateDialog(rule)" title="编辑">✏️</button>
            <button class="btn-icon danger" @click="confirmDeleteAggregate(rule)" title="删除">🗑️</button>
          </div>
        </div>
        <div class="rule-card-body">
          <div class="aggregate-config">
            <div class="agg-item"><b>聚合维度:</b> <span class="label-chip" v-for="g in rule.group_by.split(',')" :key="g">{{ g.trim() }}</span></div>
            <div class="agg-item"><b>首次延迟:</b> {{ rule.group_wait }}</div>
            <div class="agg-item"><b>批次间隔:</b> {{ rule.group_interval }}</div>
            <div class="agg-item"><b>重复间隔:</b> {{ rule.repeat_interval }}</div>
          </div>
        </div>
      </div>
      <div class="empty-state" v-if="!aggregateList.length">
        <div class="empty-icon">📦</div>
        <h3>暂无聚合规则</h3>
        <p>配置后，同类告警将被合并为一条通知发送，有效减少告警风暴</p>
      </div>
    </div>

    <!-- 新增/编辑静默规则弹窗 -->
    <div class="modal-overlay" v-if="dialogVisible" @click.self="dialogVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>{{ editingId ? '编辑静默规则' : '新增静默规则' }}</h3>
          <button class="modal-close" @click="dialogVisible = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <label>规则名称 <span class="required">*</span></label>
            <input v-model="form.name" placeholder="如: 维护窗口-节点升级" />
          </div>
          <div class="form-section">
            <h4 class="section-title">匹配条件</h4>
            <p class="section-desc">符合所有条件的告警将被静默</p>
            <div v-for="(m, i) in form.matchersList" :key="i" class="matcher-row">
              <input v-model="m.label" placeholder="标签名 (如 severity)" class="matcher-input" />
              <select v-model="m.op" class="matcher-op">
                <option value="=">=</option>
                <option value="!=">!=</option>
                <option value="=~">=~ (包含)</option>
                <option value="!~">!~ (不包含)</option>
              </select>
              <input v-model="m.value" placeholder="值 (如 info)" class="matcher-input" />
              <button class="btn-icon danger" @click="form.matchersList.splice(i, 1)">✕</button>
            </div>
            <button class="btn-add-matcher" @click="form.matchersList.push({ label: '', op: '=', value: '' })">+ 添加条件</button>
          </div>
          <div class="form-section">
            <h4 class="section-title">生效时间</h4>
            <div class="form-row-group">
              <div class="form-row half">
                <label>开始时间</label>
                <input type="datetime-local" v-model="form.starts_at_str" />
              </div>
              <div class="form-row half">
                <label>结束时间 (或填持续时间)</label>
                <input type="datetime-local" v-model="form.ends_at_str" />
              </div>
            </div>
            <div class="form-row-group">
              <div class="form-row half">
                <label>持续时间 (与结束时间二选一)</label>
                <input v-model="form.duration" placeholder="如 2h / 1d / 30m" />
              </div>
              <div class="form-row half">
                <label>重复类型</label>
                <select v-model="form.repeat_type">
                  <option value="once">仅一次</option>
                  <option value="daily">每天</option>
                  <option value="weekly">每周</option>
                  <option value="cron">Cron表达式</option>
                </select>
              </div>
            </div>
            <div class="form-row" v-if="form.repeat_type === 'cron'">
              <label>Cron 表达式</label>
              <input v-model="form.repeat_cron" placeholder="0 2 * * * (每天凌晨2点)" />
            </div>
          </div>
          <div class="form-row">
            <label>备注</label>
            <textarea v-model="form.comment" rows="2" placeholder="静默原因说明"></textarea>
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

    <!-- 新增/编辑抑制规则弹窗 -->
    <div class="modal-overlay" v-if="inhibitDialogVisible" @click.self="inhibitDialogVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>{{ inhibitEditingId ? '编辑抑制规则' : '新增抑制规则' }}</h3>
          <button class="modal-close" @click="inhibitDialogVisible = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <label>规则名称 <span class="required">*</span></label>
            <input v-model="inhibitForm.name" placeholder="如: Critical抑制Warning" />
          </div>
          <div class="form-section">
            <h4 class="section-title">源告警匹配 (高优先级)</h4>
            <p class="section-desc">当匹配此条件的告警触发时，会抑制下方目标告警的通知</p>
            <div v-for="(m, i) in inhibitForm.sourceMatchersList" :key="'s'+i" class="matcher-row">
              <input v-model="m.label" placeholder="标签名 (如 severity)" class="matcher-input" />
              <select v-model="m.op" class="matcher-op">
                <option value="=">=</option>
                <option value="!=">!=</option>
                <option value="=~">=~</option>
                <option value="!~">!~</option>
              </select>
              <input v-model="m.value" placeholder="值 (如 critical)" class="matcher-input" />
              <button class="btn-icon danger" @click="inhibitForm.sourceMatchersList.splice(i, 1)">✕</button>
            </div>
            <button class="btn-add-matcher" @click="inhibitForm.sourceMatchersList.push({ label: '', op: '=', value: '' })">+ 添加源条件</button>
          </div>
          <div class="form-section">
            <h4 class="section-title">目标告警匹配 (被抑制)</h4>
            <p class="section-desc">当源告警存在时，匹配此条件的告警通知将被抑制</p>
            <div v-for="(m, i) in inhibitForm.targetMatchersList" :key="'t'+i" class="matcher-row">
              <input v-model="m.label" placeholder="标签名 (如 severity)" class="matcher-input" />
              <select v-model="m.op" class="matcher-op">
                <option value="=">=</option>
                <option value="!=">!=</option>
                <option value="=~">=~</option>
                <option value="!~">!~</option>
              </select>
              <input v-model="m.value" placeholder="值 (如 warning)" class="matcher-input" />
              <button class="btn-icon danger" @click="inhibitForm.targetMatchersList.splice(i, 1)">✕</button>
            </div>
            <button class="btn-add-matcher" @click="inhibitForm.targetMatchersList.push({ label: '', op: '=', value: '' })">+ 添加目标条件</button>
          </div>
          <div class="form-row">
            <label>关联标签</label>
            <input v-model="inhibitForm.equal_labels" placeholder="instance,alertname（逗号分隔，源和目标共有的标签）" />
            <p class="field-hint">只有当源告警与目标告警在这些标签上值相同时才会抑制</p>
          </div>
          <div class="form-row">
            <label>描述</label>
            <textarea v-model="inhibitForm.description" rows="2" placeholder="规则说明..."></textarea>
          </div>
          <div class="form-row-inline">
            <label class="checkbox-label">
              <input type="checkbox" v-model="inhibitForm.enabled" /> 立即启用
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="inhibitDialogVisible = false">取消</button>
          <button class="btn-primary" @click="submitInhibitForm" :disabled="submitting">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 新增/编辑聚合规则弹窗 -->
    <div class="modal-overlay" v-if="aggregateDialogVisible" @click.self="aggregateDialogVisible = false">
      <div class="modal-dialog modal-lg">
        <div class="modal-header">
          <h3>{{ aggregateEditingId ? '编辑聚合规则' : '新增聚合规则' }}</h3>
          <button class="modal-close" @click="aggregateDialogVisible = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <label>规则名称 <span class="required">*</span></label>
            <input v-model="aggregateForm.name" placeholder="如: 默认聚合策略" />
          </div>
          <div class="form-section">
            <h4 class="section-title">聚合配置</h4>
            <p class="section-desc">相同分组标签的告警将被合并为一条通知</p>
            <div class="form-row">
              <label>分组标签 (Group By) <span class="required">*</span></label>
              <input v-model="aggregateForm.group_by" placeholder="alertname,severity,instance（逗号分隔）" />
              <p class="field-hint">具有相同分组标签值的告警将被聚合为一组发送</p>
            </div>
            <div class="form-row-group">
              <div class="form-row half">
                <label>首次等待 (Group Wait)</label>
                <input v-model="aggregateForm.group_wait" placeholder="30s" />
                <p class="field-hint">收到第一条告警后等待此时间，再将同组告警合并发送</p>
              </div>
              <div class="form-row half">
                <label>分组间隔 (Group Interval)</label>
                <input v-model="aggregateForm.group_interval" placeholder="5m" />
                <p class="field-hint">如果有新告警加入已有分组，最快多久发送一次更新</p>
              </div>
            </div>
            <div class="form-row">
              <label>重复间隔 (Repeat Interval)</label>
              <input v-model="aggregateForm.repeat_interval" placeholder="4h" />
              <p class="field-hint">告警持续未恢复时，重新发送通知的间隔</p>
            </div>
          </div>
          <div class="form-row">
            <label>关联通知渠道 ID</label>
            <input v-model="aggregateForm.channel_ids" placeholder="1,2（逗号分隔渠道ID）" />
          </div>
          <div class="form-row-inline">
            <label class="checkbox-label">
              <input type="checkbox" v-model="aggregateForm.enabled" /> 立即启用
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-outline" @click="aggregateDialogVisible = false">取消</button>
          <button class="btn-primary" @click="submitAggregateForm" :disabled="submitting">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget = null">
      <div class="modal-dialog modal-sm">
        <div class="modal-header"><h3>确认删除</h3><button class="modal-close" @click="deleteTarget = null">×</button></div>
        <div class="modal-body"><p>确定删除规则 <b>{{ deleteTarget.name }}</b>？</p></div>
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
import { Message } from '@arco-design/web-vue'
import {
  listSilenceRules, createSilenceRule, updateSilenceRule, deleteSilenceRule,
  listInhibitRules, createInhibitRule, updateInhibitRule, deleteInhibitRule,
  listAggregateRules, createAggregateRule, updateAggregateRule, deleteAggregateRule,
} from '@/api/monitoring'

const activeTab = ref('silence')
const list = ref([])
const inhibitList = ref([])
const aggregateList = ref([])
const dialogVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const deleteTarget = ref(null)

const ruleTabs = [
  { value: 'silence', icon: '🔇', label: '静默规则', desc: '按时间窗口和标签匹配静默告警' },
  { value: 'inhibit', icon: '🛡️', label: '抑制规则', desc: '高优告警触发时自动抑制低优同类' },
  { value: 'aggregate', icon: '📦', label: '聚合规则', desc: '合并同类告警减少通知频率' },
]
const repeatTypeMap = { once: '仅一次', daily: '每天', weekly: '每周', cron: 'Cron' }

const defaultForm = () => ({
  name: '', type: 'silence', matchers: '', matchersList: [{ label: '', op: '=', value: '' }],
  starts_at: 0, ends_at: 0, starts_at_str: '', ends_at_str: '',
  duration: '', repeat_type: 'once', repeat_cron: '', comment: '', enabled: true,
})
const form = reactive(defaultForm())

function parseMatchers(json) {
  try { return JSON.parse(json) || [] } catch { return [] }
}
function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts * 1000).toLocaleString('zh-CN', { hour12: false })
}
function isExpired(rule) {
  if (!rule.ends_at) return false
  return rule.ends_at < Date.now() / 1000
}
function getRuleStatus(rule) {
  if (!rule.enabled) return 'disabled'
  if (isExpired(rule)) return 'expired'
  return 'active'
}
function getRuleStatusText(rule) {
  if (!rule.enabled) return '已停用'
  if (isExpired(rule)) return '已过期'
  return '生效中'
}

function openDialog(rule = null) {
  Object.assign(form, defaultForm())
  editingId.value = null
  if (rule) {
    editingId.value = rule.id
    Object.assign(form, rule)
    form.matchersList = parseMatchers(rule.matchers)
    if (!form.matchersList.length) form.matchersList = [{ label: '', op: '=', value: '' }]
    if (rule.starts_at) form.starts_at_str = new Date(rule.starts_at * 1000).toISOString().slice(0, 16)
    if (rule.ends_at) form.ends_at_str = new Date(rule.ends_at * 1000).toISOString().slice(0, 16)
  }
  dialogVisible.value = true
}

// 抑制规则弹窗
const inhibitDialogVisible = ref(false)
const inhibitEditingId = ref(null)
const defaultInhibitForm = () => ({
  name: '', source_matchers: '', target_matchers: '',
  sourceMatchersList: [{ label: 'severity', op: '=', value: 'critical' }],
  targetMatchersList: [{ label: 'severity', op: '=', value: 'warning' }],
  equal_labels: 'instance,alertname', description: '', enabled: true,
})
const inhibitForm = reactive(defaultInhibitForm())

function openInhibitDialog(rule = null) {
  Object.assign(inhibitForm, defaultInhibitForm())
  inhibitEditingId.value = null
  if (rule) {
    inhibitEditingId.value = rule.id
    Object.assign(inhibitForm, rule)
    inhibitForm.sourceMatchersList = parseMatchers(rule.source_matchers)
    inhibitForm.targetMatchersList = parseMatchers(rule.target_matchers)
    if (!inhibitForm.sourceMatchersList.length) inhibitForm.sourceMatchersList = [{ label: '', op: '=', value: '' }]
    if (!inhibitForm.targetMatchersList.length) inhibitForm.targetMatchersList = [{ label: '', op: '=', value: '' }]
  }
  inhibitDialogVisible.value = true
}

async function submitInhibitForm() {
  if (!inhibitForm.name) { Message.warning('请填写规则名称'); return }
  submitting.value = true
  try {
    const payload = {
      name: inhibitForm.name,
      source_matchers: JSON.stringify(inhibitForm.sourceMatchersList.filter(m => m.label && m.value)),
      target_matchers: JSON.stringify(inhibitForm.targetMatchersList.filter(m => m.label && m.value)),
      equal_labels: inhibitForm.equal_labels,
      description: inhibitForm.description,
      enabled: inhibitForm.enabled,
    }
    if (inhibitEditingId.value) {
      await updateInhibitRule(inhibitEditingId.value, payload)
      Message.success('更新成功')
    } else {
      await createInhibitRule(payload)
      Message.success('创建成功')
    }
    inhibitDialogVisible.value = false
    loadList()
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally { submitting.value = false }
}

// 聚合规则弹窗
const aggregateDialogVisible = ref(false)
const aggregateEditingId = ref(null)
const defaultAggregateForm = () => ({
  name: '', group_by: 'alertname,severity', group_wait: '30s',
  group_interval: '5m', repeat_interval: '4h', channel_ids: '', enabled: true,
})
const aggregateForm = reactive(defaultAggregateForm())

function openAggregateDialog(rule = null) {
  Object.assign(aggregateForm, defaultAggregateForm())
  aggregateEditingId.value = null
  if (rule) {
    aggregateEditingId.value = rule.id
    Object.assign(aggregateForm, rule)
  }
  aggregateDialogVisible.value = true
}

async function submitAggregateForm() {
  if (!aggregateForm.name || !aggregateForm.group_by) { Message.warning('请填写必填字段'); return }
  submitting.value = true
  try {
    const payload = {
      name: aggregateForm.name,
      group_by: aggregateForm.group_by,
      group_wait: aggregateForm.group_wait,
      group_interval: aggregateForm.group_interval,
      repeat_interval: aggregateForm.repeat_interval,
      channel_ids: aggregateForm.channel_ids,
      enabled: aggregateForm.enabled,
    }
    if (aggregateEditingId.value) {
      await updateAggregateRule(aggregateEditingId.value, payload)
      Message.success('更新成功')
    } else {
      await createAggregateRule(payload)
      Message.success('创建成功')
    }
    aggregateDialogVisible.value = false
    loadList()
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally { submitting.value = false }
}

// 新增规则按钮 - 根据当前 Tab 打开对应弹窗
function handleAddRule() {
  if (activeTab.value === 'silence') openDialog()
  else if (activeTab.value === 'inhibit') openInhibitDialog()
  else openAggregateDialog()
}

function confirmDelete(rule) { deleteTarget.value = rule }
function confirmDeleteInhibit(rule) { deleteTarget.value = rule }
function confirmDeleteAggregate(rule) { deleteTarget.value = rule }

async function loadList() {
  try {
    if (activeTab.value === 'silence') {
      const res = await listSilenceRules({ type: 'silence' })
      if (res?.code === 0) list.value = res.data?.items || []
    } else if (activeTab.value === 'inhibit') {
      const res = await listInhibitRules({})
      if (res?.code === 0) inhibitList.value = res.data?.items || []
    } else if (activeTab.value === 'aggregate') {
      const res = await listAggregateRules({})
      if (res?.code === 0) aggregateList.value = res.data?.items || []
    }
  } catch {}
}

async function submitForm() {
  if (!form.name) { Message.warning('请填写规则名称'); return }
  submitting.value = true
  try {
    // 序列化 matchers
    const validMatchers = form.matchersList.filter(m => m.label && m.value)
    const payload = {
      ...form,
      type: 'silence',
      matchers: JSON.stringify(validMatchers),
      starts_at: form.starts_at_str ? Math.floor(new Date(form.starts_at_str).getTime() / 1000) : 0,
      ends_at: form.ends_at_str ? Math.floor(new Date(form.ends_at_str).getTime() / 1000) : 0,
    }
    delete payload.matchersList
    delete payload.starts_at_str
    delete payload.ends_at_str

    if (editingId.value) {
      await updateSilenceRule(editingId.value, payload)
      Message.success('更新成功')
    } else {
      await createSilenceRule(payload)
      Message.success('创建成功')
    }
    dialogVisible.value = false
    loadList()
  } catch (e) {
    Message.error(e?.msg || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function doDelete() {
  try {
    if (activeTab.value === 'silence') await deleteSilenceRule(deleteTarget.value.id)
    else if (activeTab.value === 'inhibit') await deleteInhibitRule(deleteTarget.value.id)
    else await deleteAggregateRule(deleteTarget.value.id)
    Message.success('删除成功')
    deleteTarget.value = null
    loadList()
  } catch {}
}

onMounted(loadList)
</script>

<style scoped>
.silence-rules-page { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { font-size: 18px; font-weight: 700; color: #1f2937; margin: 0; }
.header-desc { font-size: 13px; color: #9ca3af; margin-left: 12px; }
.btn-primary { padding: 8px 18px; background: #4f46e5; color: #fff; border: none; border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; }
.btn-primary:hover { background: #4338ca; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

/* 规则类型 Tab */
.rule-type-tabs { display: flex; gap: 12px; margin-bottom: 24px; }
.tab-btn { flex: 1; display: flex; flex-direction: column; align-items: flex-start; padding: 16px 20px; border: 2px solid #e5e7eb; border-radius: 12px; background: #fff; cursor: pointer; transition: all 0.2s; text-align: left; }
.tab-btn:hover { border-color: #a5b4fc; }
.tab-btn.active { border-color: #4f46e5; background: #f5f3ff; }
.tab-icon { font-size: 24px; margin-bottom: 6px; }
.tab-label { font-size: 14px; font-weight: 600; color: #1f2937; }
.tab-desc { font-size: 12px; color: #9ca3af; margin-top: 4px; }
.tab-btn.active .tab-label { color: #4f46e5; }

/* 规则卡片 */
.rules-list { display: flex; flex-direction: column; gap: 12px; }
.rule-card { background: #fff; border: 1px solid #e8ecf0; border-radius: 12px; padding: 18px 20px; transition: all 0.2s; }
.rule-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); }
.rule-card.expired { opacity: 0.6; border-style: dashed; }
.rule-card.disabled { opacity: 0.5; }
.rule-card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.rule-status-indicator { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.rule-status-indicator.active { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.4); }
.rule-status-indicator.expired { background: #9ca3af; }
.rule-status-indicator.disabled { background: #d1d5db; }
.rule-title-area { flex: 1; }
.rule-title-area h4 { font-size: 15px; font-weight: 600; color: #1f2937; margin: 0; }
.rule-status-text { font-size: 12px; color: #6b7280; }
.rule-actions { display: flex; gap: 6px; }
.btn-icon { width: 30px; height: 30px; border: none; background: #f3f4f6; border-radius: 6px; font-size: 14px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: background 0.2s; }
.btn-icon:hover { background: #e5e7eb; }
.btn-icon.danger:hover { background: #fee2e2; }

/* 匹配标签 */
.matcher-tags { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 10px; }
.matcher-tag { display: inline-flex; padding: 3px 10px; background: #eff6ff; color: #2563eb; border-radius: 4px; font-size: 12px; font-family: 'SF Mono', monospace; }
.matcher-tag.empty { background: #f3f4f6; color: #9ca3af; font-family: inherit; }
.rule-time-info { display: flex; gap: 16px; font-size: 12px; color: #6b7280; margin-bottom: 6px; }
.rule-time-info b { color: #374151; }
.rule-comment { font-size: 12px; color: #6b7280; font-style: italic; }
.rule-meta { display: flex; align-items: center; gap: 8px; margin-top: 8px; font-size: 12px; }
.meta-label { color: #6b7280; font-weight: 500; }
.label-chip { display: inline-block; padding: 2px 8px; background: #f0fdf4; color: #16a34a; border-radius: 4px; font-size: 11px; font-weight: 500; }

/* 抑制规则流程图 */
.inhibit-flow { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; }
.inhibit-box { flex: 1; padding: 10px 14px; border-radius: 8px; border: 1px dashed #d1d5db; }
.inhibit-box.source { border-color: #fca5a5; background: #fef2f2; }
.inhibit-box.target { border-color: #93c5fd; background: #eff6ff; }
.box-label { font-size: 11px; font-weight: 600; color: #6b7280; text-transform: uppercase; margin-bottom: 6px; display: block; }
.inhibit-arrow { font-size: 13px; color: #9ca3af; font-weight: 600; white-space: nowrap; }

/* 聚合配置 */
.aggregate-config { display: flex; flex-wrap: wrap; gap: 16px; }
.agg-item { font-size: 13px; color: #4b5563; }
.agg-item b { color: #374151; }

/* 空状态 */
.empty-state { text-align: center; padding: 48px 20px; }
.empty-icon { font-size: 42px; margin-bottom: 12px; }
.empty-state h3 { font-size: 16px; color: #374151; margin-bottom: 6px; }
.empty-state p { font-size: 13px; color: #9ca3af; }

/* 弹窗 */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-dialog { background: #fff; border-radius: 16px; width: 90%; max-height: 85vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.2); }
.modal-lg { max-width: 600px; }
.modal-sm { max-width: 420px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 20px 24px; border-bottom: 1px solid #f3f4f6; }
.modal-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.modal-close { width: 32px; height: 32px; border: none; background: #f3f4f6; border-radius: 8px; font-size: 18px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.modal-close:hover { background: #e5e7eb; }
.modal-body { padding: 24px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 24px; border-top: 1px solid #f3f4f6; }
.btn-outline { padding: 8px 18px; border: 1px solid #d1d5db; border-radius: 8px; background: #fff; font-size: 13px; color: #4b5563; cursor: pointer; }
.btn-danger { padding: 8px 18px; border: none; border-radius: 8px; background: #ef4444; color: #fff; font-size: 13px; cursor: pointer; }
.btn-danger:hover { background: #dc2626; }

/* 表单 */
.form-section { margin: 16px 0; }
.section-title { font-size: 14px; font-weight: 600; color: #374151; margin: 0 0 4px; }
.section-desc { font-size: 12px; color: #9ca3af; margin: 0 0 12px; }
.form-row { margin-bottom: 14px; }
.form-row label { display: block; font-size: 13px; font-weight: 500; color: #374151; margin-bottom: 5px; }
.form-row input, .form-row select, .form-row textarea { width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 13px; box-sizing: border-box; }
.form-row input:focus, .form-row select:focus, .form-row textarea:focus { outline: none; border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79,70,229,0.1); }
.form-row-group { display: flex; gap: 16px; }
.form-row.half { flex: 1; }
.form-row-inline { margin-top: 8px; }
.required { color: #ef4444; }
.checkbox-label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #4b5563; cursor: pointer; }
.checkbox-label input { width: auto; }

/* 匹配条件行 */
.matcher-row { display: flex; gap: 8px; margin-bottom: 8px; align-items: center; }
.matcher-input { flex: 1; padding: 7px 10px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 13px; }
.matcher-input:focus { outline: none; border-color: #4f46e5; }
.matcher-op { width: 90px; padding: 7px 8px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 13px; }
.btn-add-matcher { padding: 6px 14px; border: 1px dashed #d1d5db; border-radius: 6px; background: #fff; font-size: 12px; color: #4f46e5; cursor: pointer; transition: all 0.2s; }
.btn-add-matcher:hover { border-color: #4f46e5; background: #f5f3ff; }
.field-hint { font-size: 11px; color: #9ca3af; margin: 4px 0 0; }
</style>

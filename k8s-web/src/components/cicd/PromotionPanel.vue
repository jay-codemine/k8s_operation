<template>
  <div class="promotion-panel">
    <!-- 顶部工具栏 -->
    <div class="promo-toolbar">
      <div class="promo-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
          <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
          <line x1="12" y1="22.08" x2="12" y2="12"/>
        </svg>
        <span>镜像晋级链</span>
        <span class="promo-subtitle">一次构建，跨环境晋级同一不可变镜像</span>
      </div>
      <div class="promo-actions">
        <button class="promo-btn ghost" @click="loadAll" :disabled="loading">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          {{ loading ? '加载中...' : '刷新' }}
        </button>
        <button class="promo-btn primary" @click="openConfig">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
          配置环境目标
        </button>
      </div>
    </div>

    <!-- 加载中 -->
    <div v-if="loading && chain.length === 0" class="promo-loading">
      <div class="promo-spinner"></div>
      <p>正在加载晋级链...</p>
    </div>

    <!-- 空状态 -->
    <div v-else-if="chain.length === 0" class="promo-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
        <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
      </svg>
      <p>尚未配置任何环境部署目标</p>
      <span>为该流水线绑定 dev / test / staging / prod 等环境后，即可将同一镜像逐级晋级发布</span>
      <button class="promo-btn primary" @click="openConfig">立即配置环境目标</button>
    </div>

    <!-- 晋级链 -->
    <div v-else class="promo-chain">
      <div v-for="(node, idx) in chain" :key="node.env" class="chain-item">
        <div :class="['env-card', `env-${node.env}`, { deployed: !!node.current_release_id }]">
          <div class="env-accent" :style="{ background: envColor(node.env) }"></div>
          <div class="env-card-head">
            <span :class="['env-badge', `env-${node.env}`]" :style="{ background: envColor(node.env) }">{{ envLabel(node.env) }}</span>
            <span class="env-code">{{ node.env }}</span>
            <span v-if="node.require_approval" class="env-tag approval" title="晋级到该环境需要审批">🔒 需审批</span>
            <span v-if="node.auto_deploy" class="env-tag auto" title="CI 成功后自动部署">⚡ 自动</span>
          </div>

          <!-- 当前运行的不可变镜像（大厂风格：突出版本 chip） -->
          <div class="env-version">
            <template v-if="node.current_release_id">
              <div class="ver-row">
                <span :class="['ver-dot', `st-${(node.current_release_status || '').toLowerCase()}`]"></span>
                <span class="ver-tag mono" :title="fullImage(node)">{{ node.current_image_tag || node.current_image_digest || '—' }}</span>
                <span class="ver-status">{{ releaseStatusText(node.current_release_status) }}</span>
              </div>
              <div class="ver-time">{{ formatTs(node.current_deploy_time) }}</div>
            </template>
            <div v-else class="ver-none">尚未部署</div>
          </div>

          <div class="env-target">
            <div class="target-line">
              <span class="tl-label">集群</span>
              <span class="tl-value">{{ node.cluster_name || ('#' + node.cluster_id) }}</span>
            </div>
            <div class="target-line">
              <span class="tl-label">负载</span>
              <span class="tl-value mono" :title="`${node.namespace}/${node.workload_kind}/${node.workload_name}`">{{ node.namespace }}/{{ node.workload_name }}</span>
            </div>
          </div>

          <div class="env-card-foot">
            <button class="promo-btn block promote" @click="openPromote(node)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="5" y1="12" x2="19" y2="12"/>
                <polyline points="12 5 19 12 12 19"/>
              </svg>
              晋级到{{ envLabel(node.env) }}
            </button>
          </div>
        </div>
        <!-- 晋级连接线：同一镜像流转 + 审批门禁 -->
        <div v-if="idx < chain.length - 1" class="chain-arrow">
          <div class="ca-line"></div>
          <div :class="['ca-node', { gated: chain[idx + 1].require_approval }]" :title="chain[idx + 1].require_approval ? '晋级到下一环境需审批' : '同一镜像逐级晋级'">
            <svg v-if="chain[idx + 1].require_approval" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2"/>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="5" y1="12" x2="19" y2="12"/>
              <polyline points="12 5 19 12 12 19"/>
            </svg>
          </div>
          <div class="ca-line"></div>
        </div>
      </div>
    </div>

    <!-- ============ 配置环境目标弹窗 ============ -->
    <Teleport to="body">
      <div v-if="showConfigModal" class="promo-modal-mask" @click.self="showConfigModal = false">
        <div class="promo-modal wide">
          <div class="pm-head">
            <h3>配置环境部署目标</h3>
            <button class="pm-close" @click="showConfigModal = false">✕</button>
          </div>
          <div class="pm-body">
            <p class="pm-tip">为每个环境绑定部署目标；晋级时将复用已构建镜像发布到这里。生产/预发建议开启「需审批」。</p>
            <div class="target-rows">
              <div class="tr-header">
                <span>环境</span><span>集群</span><span>命名空间</span><span>类型</span>
                <span>工作负载</span><span>容器</span><span>来源</span><span>审批</span><span></span>
              </div>
              <div v-for="(row, i) in editTargets" :key="i" class="tr-row">
                <select v-model="row.env" class="tr-input">
                  <option v-for="e in envOptions" :key="e.value" :value="e.value">{{ e.label }}({{ e.value }})</option>
                </select>
                <select v-model.number="row.cluster_id" class="tr-input">
                  <option :value="0">选择集群</option>
                  <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name || c.name }}</option>
                </select>
                <input v-model="row.namespace" class="tr-input" placeholder="default" />
                <select v-model="row.workload_kind" class="tr-input">
                  <option value="Deployment">Deployment</option>
                  <option value="StatefulSet">StatefulSet</option>
                  <option value="DaemonSet">DaemonSet</option>
                </select>
                <input v-model="row.workload_name" class="tr-input" placeholder="工作负载名" />
                <input v-model="row.container" class="tr-input" placeholder="容器(默认同名)" />
                <select v-model="row.promote_from" class="tr-input">
                  <option value="">首环境</option>
                  <option v-for="e in envOptions" :key="e.value" :value="e.value">{{ e.value }}</option>
                </select>
                <label class="tr-check"><input type="checkbox" v-model="row.require_approval" /></label>
                <button class="tr-del" @click="removeRow(i)" title="删除">✕</button>
              </div>
            </div>
            <button class="promo-btn ghost add-row" @click="addRow">+ 添加环境</button>
          </div>
          <div class="pm-foot">
            <button class="promo-btn ghost" @click="showConfigModal = false">取消</button>
            <button class="promo-btn primary" @click="saveConfig" :disabled="saving">{{ saving ? '保存中...' : '保存配置' }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ============ 晋级弹窗 ============ -->
    <Teleport to="body">
      <div v-if="showPromoteModal" class="promo-modal-mask" @click.self="showPromoteModal = false">
        <div class="promo-modal">
          <div class="pm-head">
            <h3>镜像晋级 → {{ envLabel(promoteForm.target_env) }}({{ promoteForm.target_env }})</h3>
            <button class="pm-close" @click="showPromoteModal = false">✕</button>
          </div>
          <div class="pm-body">
            <!-- 晋级流转横幅：明确「哪个镜像、从哪来、到哪去」 -->
            <div class="promote-flow">
              <span class="pf-from">{{ promoteForm.source === 'upstream' && promoteForm.promote_from ? envLabel(promoteForm.promote_from) : (promoteForm.source === 'manual' ? '手动镜像' : '最新构建') }}</span>
              <svg class="pf-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
              <span :class="['pf-to', `env-${promoteForm.target_env}`]">{{ envLabel(promoteForm.target_env) }}</span>
            </div>
            <div class="pm-field">
              <label>镜像来源</label>
              <div class="src-radios">
                <label><input type="radio" value="run" v-model="promoteForm.source" /> 最新构建产物</label>
                <label v-if="upstreamNode"><input type="radio" value="upstream" v-model="promoteForm.source" /> 从上游({{ promoteForm.promote_from }})晋级</label>
                <label><input type="radio" value="manual" v-model="promoteForm.source" /> 手动指定</label>
              </div>
            </div>

            <div v-if="promoteForm.source === 'run'" class="pm-field">
              <label>运行记录 ID（构建产出镜像的 Run）</label>
              <input v-model.number="promoteForm.source_run_id" class="pm-input" type="number" placeholder="留空 = 自动使用最新一次成功构建" />
              <p class="pm-hint">留空时后端自动选取该流水线最近一次已构建的镜像进行晋级。</p>
            </div>

            <div v-else-if="promoteForm.source === 'upstream'" class="pm-field">
              <label>上游环境当前镜像</label>
              <div class="pm-readonly mono">{{ upstreamNode ? fullImage(upstreamNode) : '上游环境暂无部署' }}</div>
            </div>

            <template v-else>
              <div class="pm-field">
                <label>镜像仓库</label>
                <input v-model="promoteForm.image_repo" class="pm-input" placeholder="registry/app" />
              </div>
              <div class="pm-field-inline">
                <div class="pm-field">
                  <label>Tag</label>
                  <input v-model="promoteForm.image_tag" class="pm-input" placeholder="v1.0.0" />
                </div>
                <div class="pm-field">
                  <label>Digest（可选）</label>
                  <input v-model="promoteForm.image_digest" class="pm-input" placeholder="sha256:..." />
                </div>
              </div>
            </template>

            <div class="pm-field">
              <label>晋级说明</label>
              <textarea v-model="promoteForm.reason" class="pm-input" rows="2" placeholder="例如：验收通过，晋级至生产"></textarea>
            </div>
            <p v-if="targetRequiresApproval" class="pm-warn">该环境已开启审批，提交后将进入审批流程，审批通过才会执行部署。</p>
          </div>
          <div class="pm-foot">
            <button class="promo-btn ghost" @click="showPromoteModal = false">取消</button>
            <button class="promo-btn primary" @click="submitPromote" :disabled="promoting">{{ promoting ? '提交中...' : '确认晋级' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { getClusterList } from '@/api/cluster'
import {
  getPipelineTargets,
  savePipelineTargets,
  promotePipeline,
  getPromotionChain,
  getK8sEnvironments,
  getReleaseDetail
} from '@/api/cicd'

const props = defineProps({
  pipelineId: { type: [String, Number], required: true },
  pipeline: { type: Object, default: () => ({}) }
})

// 内置回退环境（当全局环境表为空或加载失败时保底，兼容旧数据）
const FALLBACK_ENVS = [
  { value: 'dev', label: '开发', order: 1, color: '#10b981', cluster_id: 0, namespace: 'default' },
  { value: 'test', label: '测试', order: 2, color: '#3b82f6', cluster_id: 0, namespace: 'default' },
  { value: 'staging', label: '预发', order: 3, color: '#f59e0b', cluster_id: 0, namespace: 'default' },
  { value: 'prod', label: '生产', order: 4, color: '#ef4444', cluster_id: 0, namespace: 'default' }
]
const ENV_FALLBACK_COLORS = { dev: '#10b981', test: '#3b82f6', staging: '#f59e0b', prod: '#ef4444' }

// 全局环境选项（来自 cicd_environment 表，可在「环境管理」中增删改）
const envOptions = ref([])

const loading = ref(false)
const saving = ref(false)
const promoting = ref(false)
const chain = ref([])
const clusters = ref([])

const showConfigModal = ref(false)
const editTargets = ref([])

const showPromoteModal = ref(false)
const promoteForm = reactive({
  target_env: '',
  promote_from: '',
  source: 'run',
  source_run_id: null,
  source_release_id: 0,
  image_repo: '',
  image_tag: '',
  image_digest: '',
  reason: ''
})

// ====== 轻量 Toast ======
const showToast = (msg, type = 'info', duration = 2500) => {
  const colors = { success: '#38a169', error: '#e53e3e', info: '#3182ce', warning: '#dd6b20' }
  const el = document.createElement('div')
  el.textContent = msg
  Object.assign(el.style, {
    position: 'fixed', top: '20px', left: '50%', transform: 'translateX(-50%)',
    padding: '10px 24px', borderRadius: '8px', color: '#fff', fontSize: '14px',
    background: colors[type] || colors.info, zIndex: '99999', boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
    transition: 'opacity 0.3s', opacity: '1'
  })
  document.body.appendChild(el)
  setTimeout(() => { el.style.opacity = '0'; setTimeout(() => el.remove(), 300) }, duration)
}

// ====== 桌面通知（浏览器 Notification API，权限惰性申请）======
const ensureNotifyPermission = () => {
  try {
    if (typeof Notification === 'undefined') return
    if (Notification.permission === 'default') Notification.requestPermission().catch(() => {})
  } catch { /* 忽略：部分浏览器/非 HTTPS 环境不支持 */ }
}
const notifyDesktop = (title, body) => {
  try {
    if (typeof Notification !== 'undefined' && Notification.permission === 'granted') {
      const n = new Notification(title, { body, tag: 'cicd-promote', icon: '/favicon.ico' })
      setTimeout(() => { try { n.close() } catch { /* noop */ } }, 6000)
    }
  } catch { /* 忽略通知异常，不影响主流程 */ }
}

// ====== 晋级完成轮询：拿到 release_id 后轮询发布单状态到终态再通知 ======
// 终态：Succeeded / Failed / Canceled / Rollback；AwaitingApproval 为待审批（仅提示一次，继续轮询）
const pollPromotion = (releaseId, envText) => {
  if (!releaseId) return
  const maxAttempts = 60 // 60 次 * 3s ≈ 3 分钟（客户端尽力而为，页面关闭即停止）
  let attempts = 0
  let approvalNotified = false
  const finish = async (msg, type, title, body, duration) => {
    showToast(msg, type, duration)
    notifyDesktop(title, body)
    await loadChain()
  }
  const tick = async () => {
    attempts++
    try {
      const res = await getReleaseDetail(releaseId)
      const rel = res?.data?.release || res?.data?.data?.release
      const status = rel?.status
      if (status === 'Succeeded') {
        await finish(`✅ ${envText} 晋级完成`, 'success', '镜像晋级完成', `${envText} 已成功部署`, 5000)
        return
      }
      if (status === 'Failed' || status === 'Canceled') {
        const tip = rel?.message ? `：${rel.message}` : ''
        await finish(`❌ ${envText} 晋级失败${tip}`, 'error', '镜像晋级失败', `${envText} 部署失败${tip}`, 6000)
        return
      }
      if (status === 'Rollback') {
        await finish(`⚠️ ${envText} 已回滚`, 'warning', '镜像晋级已回滚', `${envText} 已回滚`, 5000)
        return
      }
      if (status === 'AwaitingApproval' && !approvalNotified) {
        approvalNotified = true
        showToast(`⏳ ${envText} 等待审批中，审批通过后将自动部署`, 'info', 5000)
        notifyDesktop('镜像晋级待审批', `${envText} 正在等待审批`)
      }
    } catch { /* 忽略单次轮询失败，继续重试 */ }
    if (attempts < maxAttempts) setTimeout(tick, 3000)
  }
  setTimeout(tick, 3000)
}

const pid = computed(() => Number(props.pipelineId))

const envLabel = (env) => envOptions.value.find(e => e.value === env)?.label || env
const envColor = (env) => envOptions.value.find(e => e.value === env)?.color || ENV_FALLBACK_COLORS[env] || '#6366f1'
const releaseStatusText = (s) => {
  const map = {
    Pending: '待处理', AwaitingApproval: '待审批', Queued: '排队中', Running: '部署中',
    Succeeded: '已部署', Failed: '失败', Canceled: '已取消', Rollback: '已回滚'
  }
  return map[s] || s || '-'
}
const fullImage = (node) => {
  if (!node) return '-'
  const repo = node.current_image_repo || ''
  if (!repo) return '-'
  if (node.current_image_digest) return `${repo}@${node.current_image_digest}`
  return node.current_image_tag ? `${repo}:${node.current_image_tag}` : repo
}
const formatTs = (ts) => {
  if (!ts) return ''
  const d = new Date(Number(ts) * 1000)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleString('zh-CN', { hour12: false })
}

const targetRequiresApproval = computed(() => {
  const n = chain.value.find(x => x.env === promoteForm.target_env)
  return !!(n && n.require_approval)
})
const upstreamNode = computed(() => {
  if (!promoteForm.promote_from) return null
  return chain.value.find(x => x.env === promoteForm.promote_from && x.current_release_id) || null
})

// ====== 数据加载 ======
const parseList = (res, key) => res?.data?.[key] || res?.data?.data?.[key] || []

const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 1000 })
    const data = res.data || res
    clusters.value = data.list || data.items || data || []
  } catch (e) {
    console.warn('加载集群失败', e)
  }
}

// 加载全局环境（替换写死的四环境，支持在「环境管理」中自由增删改）
const loadEnvOptions = async () => {
  try {
    const res = await getK8sEnvironments({ page: 1, page_size: 1000 })
    const list = res?.data?.list || res?.data?.data?.list || (Array.isArray(res?.data) ? res.data : [])
    const mapped = (Array.isArray(list) ? list : []).map(e => ({
      value: e.name,
      label: e.display_name || e.name,
      order: e.sort_order || 0,
      color: e.color || '',
      cluster_id: e.cluster_id || 0,
      namespace: e.namespace || 'default',
      require_approval: !!e.require_approval
    })).sort((a, b) => a.order - b.order)
    envOptions.value = mapped.length > 0 ? mapped : FALLBACK_ENVS
  } catch (e) {
    envOptions.value = FALLBACK_ENVS
  }
}

const loadChain = async () => {
  if (!pid.value) return
  loading.value = true
  try {
    const res = await getPromotionChain(pid.value)
    chain.value = parseList(res, 'chain')
  } catch (e) {
    showToast('加载晋级链失败: ' + (e.message || ''), 'error')
  } finally {
    loading.value = false
  }
}

const loadAll = async () => {
  await Promise.all([loadClusters(), loadEnvOptions(), loadChain()])
}

// ====== 配置 ======
const seedDefaults = () => {
  const p = props.pipeline || {}
  return envOptions.value.map((e, i) => ({
    env: e.value,
    cluster_id: e.cluster_id || p.target_cluster_id || 0,
    namespace: e.namespace || p.target_namespace || 'default',
    workload_kind: p.target_workload_kind || 'Deployment',
    workload_name: p.target_workload_name || '',
    container: p.target_container || '',
    auto_deploy: e.value === 'dev',
    require_approval: e.require_approval || e.value === 'staging' || e.value === 'prod',
    promote_from: i === 0 ? '' : envOptions.value[i - 1].value,
    sort_order: e.order || (i + 1)
  }))
}

const openConfig = async () => {
  if (clusters.value.length === 0) await loadClusters()
  try {
    const res = await getPipelineTargets(pid.value)
    const list = parseList(res, 'list')
    if (list.length > 0) {
      editTargets.value = list.map(t => ({
        env: t.env,
        cluster_id: t.cluster_id,
        namespace: t.namespace,
        workload_kind: t.workload_kind || 'Deployment',
        workload_name: t.workload_name,
        container: t.container,
        auto_deploy: !!t.auto_deploy,
        require_approval: !!t.require_approval,
        promote_from: t.promote_from || '',
        sort_order: t.sort_order || 0
      }))
    } else {
      editTargets.value = seedDefaults()
    }
  } catch {
    editTargets.value = seedDefaults()
  }
  showConfigModal.value = true
}

const addRow = () => {
  const used = editTargets.value.map(r => r.env)
  const next = envOptions.value.find(e => !used.includes(e.value)) || envOptions.value[0] || FALLBACK_ENVS[0]
  editTargets.value.push({
    env: next.value, cluster_id: next.cluster_id || props.pipeline?.target_cluster_id || 0,
    namespace: next.namespace || 'default', workload_kind: 'Deployment', workload_name: '',
    container: '', auto_deploy: false, require_approval: !!next.require_approval,
    promote_from: '', sort_order: editTargets.value.length + 1
  })
}
const removeRow = (i) => editTargets.value.splice(i, 1)

const saveConfig = async () => {
  // 前端校验
  const seen = new Set()
  for (const row of editTargets.value) {
    if (!row.env) { showToast('存在未选择环境的行', 'warning'); return }
    if (seen.has(row.env)) { showToast(`环境 ${row.env} 重复`, 'warning'); return }
    seen.add(row.env)
    if (!row.cluster_id) { showToast(`环境 ${row.env} 未选择集群`, 'warning'); return }
    if (!row.workload_name) { showToast(`环境 ${row.env} 未填写工作负载`, 'warning'); return }
  }
  saving.value = true
  try {
    const res = await savePipelineTargets({
      pipeline_id: pid.value,
      targets: editTargets.value.map((r, i) => ({ ...r, sort_order: r.sort_order || (i + 1) }))
    })
    if (res.code === 0) {
      showToast('环境目标配置已保存', 'success')
      showConfigModal.value = false
      await loadChain()
    } else {
      showToast(res.msg || '保存失败', 'error')
    }
  } catch (e) {
    showToast('保存失败: ' + (e.message || ''), 'error')
  } finally {
    saving.value = false
  }
}

// ====== 晋级 ======
const openPromote = (node) => {
  promoteForm.target_env = node.env
  promoteForm.promote_from = node.promote_from || ''
  promoteForm.source = node.promote_from ? 'upstream' : 'run'
  promoteForm.source_run_id = props.pipeline?.last_run_id || null
  promoteForm.source_release_id = 0
  promoteForm.image_repo = ''
  promoteForm.image_tag = ''
  promoteForm.image_digest = ''
  promoteForm.reason = ''
  // 无上游部署则回退到最新构建
  if (promoteForm.source === 'upstream' && !upstreamNode.value) {
    promoteForm.source = 'run'
  }
  showPromoteModal.value = true
}

const submitPromote = async () => {
  const payload = {
    pipeline_id: pid.value,
    target_env: promoteForm.target_env,
    reason: promoteForm.reason
  }
  if (promoteForm.source === 'run') {
    if (promoteForm.source_run_id) payload.source_run_id = Number(promoteForm.source_run_id)
  } else if (promoteForm.source === 'upstream') {
    if (!upstreamNode.value) { showToast('上游环境暂无可晋级的镜像', 'warning'); return }
    payload.source_release_id = upstreamNode.value.current_release_id
  } else {
    if (!promoteForm.image_repo) { showToast('请填写镜像仓库', 'warning'); return }
    payload.image_repo = promoteForm.image_repo
    payload.image_tag = promoteForm.image_tag
    payload.image_digest = promoteForm.image_digest
  }
  promoting.value = true
  try {
    ensureNotifyPermission()
    const envText = envLabel(promoteForm.target_env)
    const res = await promotePipeline(payload)
    if (res.code === 0) {
      const rid = res.data?.release_id || res.data?.data?.release_id
      showToast(res.data?.message || '晋级已提交', 'success')
      showPromoteModal.value = false
      await loadChain()
      // 异步轮询该发布单直到部署完成/失败，届时弹出完成通知
      if (rid) pollPromotion(rid, envText)
    } else {
      showToast(res.msg || '晋级失败', 'error')
    }
  } catch (e) {
    showToast('晋级失败: ' + (e.message || ''), 'error')
  } finally {
    promoting.value = false
  }
}

loadAll()
defineExpose({ refresh: loadAll })
</script>

<style scoped>
.promotion-panel { padding: 4px 2px; }
.promo-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 18px; flex-wrap: wrap; gap: 10px; }
.promo-title { display: flex; align-items: center; gap: 8px; font-size: 16px; font-weight: 600; color: #1f2937; }
.promo-title svg { width: 20px; height: 20px; color: #6366f1; }
.promo-subtitle { font-size: 12px; font-weight: 400; color: #9ca3af; margin-left: 4px; }
.promo-actions { display: flex; gap: 8px; }
.promo-btn { display: inline-flex; align-items: center; gap: 6px; padding: 7px 14px; border-radius: 8px; font-size: 13px; cursor: pointer; border: 1px solid transparent; transition: all .15s; }
.promo-btn svg { width: 15px; height: 15px; }
.promo-btn.primary { background: #6366f1; color: #fff; }
.promo-btn.primary:hover { background: #4f46e5; }
.promo-btn.ghost { background: #fff; color: #4b5563; border-color: #e5e7eb; }
.promo-btn.ghost:hover { background: #f9fafb; }
.promo-btn.small { padding: 5px 10px; font-size: 12px; }
.promo-btn:disabled { opacity: .6; cursor: not-allowed; }

.promo-loading, .promo-empty { text-align: center; padding: 48px 16px; color: #9ca3af; }
.promo-empty svg { width: 46px; height: 46px; margin-bottom: 12px; color: #d1d5db; }
.promo-empty p { font-size: 15px; color: #6b7280; margin: 0 0 6px; }
.promo-empty span { display: block; font-size: 13px; margin-bottom: 16px; }
.promo-spinner { width: 32px; height: 32px; border: 3px solid #e5e7eb; border-top-color: #6366f1; border-radius: 50%; margin: 0 auto 12px; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.promo-chain { display: flex; align-items: stretch; flex-wrap: wrap; gap: 0; }
.chain-item { display: flex; align-items: stretch; }
.env-card { position: relative; width: 236px; border: 1px solid #e5e7eb; border-radius: 14px; padding: 0 14px 14px; background: #fff; display: flex; flex-direction: column; gap: 10px; box-shadow: 0 1px 3px rgba(0,0,0,0.05); overflow: hidden; transition: box-shadow .18s, transform .18s; }
.env-card:hover { box-shadow: 0 8px 24px rgba(79,70,229,0.14); transform: translateY(-2px); }
.env-card.deployed { border-color: #c7d2fe; }
.env-accent { position: absolute; top: 0; left: 0; right: 0; height: 4px; }
.env-card.env-dev .env-accent { background: #10b981; }
.env-card.env-test .env-accent { background: #3b82f6; }
.env-card.env-staging .env-accent { background: #f59e0b; }
.env-card.env-prod .env-accent { background: #ef4444; }
.env-card-head { display: flex; align-items: center; gap: 6px; padding-top: 16px; flex-wrap: wrap; }
.env-badge { font-size: 12px; font-weight: 700; padding: 2px 10px; border-radius: 6px; color: #fff; }
.env-badge.env-dev { background: #10b981; }
.env-badge.env-test { background: #3b82f6; }
.env-badge.env-staging { background: #f59e0b; }
.env-badge.env-prod { background: #ef4444; }
.env-code { font-size: 11px; color: #9ca3af; font-family: 'SFMono-Regular', Consolas, monospace; }
.env-tag { font-size: 10.5px; padding: 1px 7px; border-radius: 5px; }
.env-tag.approval { background: #fef3c7; color: #b45309; margin-left: auto; }
.env-tag.auto { background: #dbeafe; color: #1d4ed8; }

.env-version { background: #f8fafc; border: 1px solid #eef2f7; border-radius: 10px; padding: 9px 10px; min-height: 46px; }
.ver-row { display: flex; align-items: center; gap: 7px; }
.ver-dot { width: 8px; height: 8px; border-radius: 50%; background: #d1d5db; flex-shrink: 0; }
.ver-dot.st-succeeded { background: #10b981; box-shadow: 0 0 0 3px rgba(16,185,129,.15); }
.ver-dot.st-running, .ver-dot.st-queued { background: #3b82f6; box-shadow: 0 0 0 3px rgba(59,130,246,.15); }
.ver-dot.st-failed { background: #ef4444; box-shadow: 0 0 0 3px rgba(239,68,68,.15); }
.ver-dot.st-awaitingapproval, .ver-dot.st-pending { background: #f59e0b; box-shadow: 0 0 0 3px rgba(245,158,11,.15); }
.ver-tag { font-weight: 600; color: #3730a3; background: #eef2ff; padding: 1px 8px; border-radius: 6px; max-width: 116px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ver-status { font-size: 11px; color: #6b7280; margin-left: auto; white-space: nowrap; }
.ver-time { font-size: 10.5px; color: #9ca3af; margin-top: 5px; }
.ver-none { font-size: 12px; color: #b0b6be; font-style: italic; }

.env-target { display: flex; flex-direction: column; gap: 5px; }
.target-line { display: flex; gap: 6px; font-size: 12px; }
.tl-label { color: #9ca3af; min-width: 34px; }
.tl-value { color: #374151; word-break: break-all; }
.mono { font-family: 'SFMono-Regular', Consolas, monospace; font-size: 11.5px; }

.env-card-foot { margin-top: auto; }
.promo-btn.block { width: 100%; justify-content: center; }
.promo-btn.promote { background: linear-gradient(135deg, #6366f1, #4f46e5); color: #fff; font-weight: 600; box-shadow: 0 2px 8px rgba(79,70,229,.28); }
.promo-btn.promote:hover { background: linear-gradient(135deg, #4f46e5, #4338ca); box-shadow: 0 4px 14px rgba(79,70,229,.4); }

/* 晋级连接线 */
.chain-arrow { display: flex; flex-direction: row; align-items: center; align-self: center; padding: 0 3px; }
.ca-line { width: 14px; height: 2px; background: repeating-linear-gradient(90deg, #cbd5e1 0 5px, transparent 5px 9px); }
.ca-node { display: flex; align-items: center; justify-content: center; width: 30px; height: 30px; border-radius: 50%; background: #eef2ff; color: #6366f1; flex-shrink: 0; }
.ca-node.gated { background: #fef3c7; color: #d97706; }
.ca-node svg { width: 15px; height: 15px; }

/* Modal */
.promo-modal-mask { position: fixed; inset: 0; background: rgba(15,23,42,0.45); display: flex; align-items: center; justify-content: center; z-index: 9000; }
.promo-modal { background: #fff; border-radius: 14px; width: 480px; max-width: 94vw; max-height: 88vh; display: flex; flex-direction: column; box-shadow: 0 20px 50px rgba(0,0,0,0.25); }
.promo-modal.wide { width: 860px; }
.pm-head { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid #f0f1f3; }
.pm-head h3 { margin: 0; font-size: 15px; color: #1f2937; }
.pm-close { border: none; background: none; font-size: 16px; color: #9ca3af; cursor: pointer; }
.pm-body { padding: 18px 20px; overflow-y: auto; }
.pm-tip { font-size: 12.5px; color: #6b7280; margin: 0 0 14px; line-height: 1.5; }
.pm-field { margin-bottom: 14px; }
.pm-field-inline { display: flex; gap: 12px; }
.pm-field-inline .pm-field { flex: 1; }
.pm-field label { display: block; font-size: 12.5px; color: #4b5563; margin-bottom: 5px; }
.pm-input { width: 100%; padding: 8px 10px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 13px; box-sizing: border-box; }
.pm-input:focus { outline: none; border-color: #a5b4fc; }
.pm-readonly { padding: 8px 10px; background: #f9fafb; border-radius: 8px; font-size: 12px; color: #4b5563; word-break: break-all; }
.src-radios { display: flex; flex-wrap: wrap; gap: 14px; font-size: 13px; color: #374151; }
.src-radios label { display: inline-flex; align-items: center; gap: 5px; cursor: pointer; }
.pm-warn { font-size: 12px; color: #b45309; background: #fef3c7; padding: 8px 10px; border-radius: 8px; margin: 4px 0 0; }
.pm-hint { font-size: 11.5px; color: #9ca3af; margin: 6px 0 0; }
.promote-flow { display: flex; align-items: center; justify-content: center; gap: 12px; padding: 12px; margin-bottom: 16px; background: linear-gradient(135deg,#f5f7ff,#eef2ff); border: 1px solid #e0e7ff; border-radius: 10px; }
.pf-from { font-size: 13px; font-weight: 600; color: #64748b; padding: 3px 12px; background: #fff; border-radius: 7px; border: 1px solid #e2e8f0; }
.pf-arrow { width: 22px; height: 22px; color: #6366f1; }
.pf-to { font-size: 13px; font-weight: 700; color: #fff; padding: 3px 14px; border-radius: 7px; }
.pf-to.env-dev { background: #10b981; }
.pf-to.env-test { background: #3b82f6; }
.pf-to.env-staging { background: #f59e0b; }
.pf-to.env-prod { background: #ef4444; }
.pm-foot { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid #f0f1f3; }

/* target rows */
.target-rows { display: flex; flex-direction: column; gap: 6px; }
.tr-header, .tr-row { display: grid; grid-template-columns: 84px 120px 96px 108px 1fr 96px 84px 44px 28px; gap: 6px; align-items: center; }
.tr-header { font-size: 11px; color: #9ca3af; padding: 0 2px; }
.tr-input { width: 100%; padding: 6px 7px; border: 1px solid #e5e7eb; border-radius: 6px; font-size: 12px; box-sizing: border-box; }
.tr-check { display: flex; justify-content: center; }
.tr-del { border: none; background: #fef2f2; color: #ef4444; border-radius: 6px; cursor: pointer; padding: 5px; font-size: 12px; }
.add-row { margin-top: 12px; }
</style>

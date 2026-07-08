<template>
  <div class="yaml-workbench">
    <!-- 页头 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <span class="title-icon">🛠️</span>
          YAML 工作台
        </h1>
        <p class="page-desc">批量创建资源、YAML 校验、格式化与 DryRun 预检</p>
      </div>
    </div>

    <!-- 主体两栏布局 -->
    <div class="workbench-body">
      <!-- 左侧编辑器 -->
      <div class="editor-section">
        <!-- 工具栏 -->
        <div class="wb-toolbar">
          <div class="wb-toolbar-left">
            <select v-model="targetGroup" class="wb-select">
              <option value="">API Group (可选)</option>
              <option v-for="g in groups" :key="g" :value="g">{{ g }}</option>
            </select>
            <input v-model="targetVersion" placeholder="version (如 v1)" class="wb-input small" />
            <input v-model="targetResource" placeholder="resource (如 prometheusrules)" class="wb-input" />
          </div>
          <div class="wb-toolbar-right">
            <button class="wb-btn outline" @click="formatContent">格式化</button>
            <button class="wb-btn outline" @click="clearContent">清空</button>
          </div>
        </div>

        <!-- 编辑区 -->
        <div class="wb-editor-area">
          <div class="wb-gutter">
            <div class="wb-gutter-line" v-for="n in lineCount" :key="n">{{ n }}</div>
          </div>
          <textarea
            v-model="yamlContent"
            class="wb-textarea"
            spellcheck="false"
            placeholder="# 粘贴或编写 YAML 内容...&#10;# 支持多资源（用 --- 分隔）&#10;---&#10;apiVersion: v1&#10;kind: ConfigMap&#10;metadata:&#10;  name: example"
            @scroll="syncGutter"
            @keydown.tab.prevent="insertTab"
          ></textarea>
        </div>

        <!-- 底部状态栏 -->
        <div class="wb-statusbar">
          <span class="wb-stat">{{ lineCount }} 行</span>
          <span class="wb-stat">{{ yamlContent.length }} 字符</span>
          <span class="wb-stat">{{ documentCount }} 个文档</span>
          <span class="wb-stat-tag">YAML</span>
        </div>
      </div>

      <!-- 右侧操作面板 -->
      <div class="action-panel">
        <!-- 操作卡片 -->
        <div class="action-card">
          <h3 class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2v-4M9 21H5a2 2 0 01-2-2v-4"/></svg>
            DryRun 校验
          </h3>
          <p class="card-desc">在不实际创建的情况下，验证 YAML 是否合法</p>
          <button class="action-primary-btn" @click="runDryRun" :disabled="dryRunning || !yamlContent.trim()">
            {{ dryRunning ? '校验中...' : '执行 DryRun' }}
          </button>
        </div>

        <div class="action-card">
          <h3 class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
            提交创建
          </h3>
          <p class="card-desc">将 YAML 中的资源提交到集群进行创建</p>
          <button class="action-submit-btn" @click="submitCreate" :disabled="creating || !yamlContent.trim()">
            {{ creating ? '创建中...' : '提交到集群' }}
          </button>
        </div>

        <!-- 结果面板 -->
        <div class="result-panel" v-if="results.length > 0">
          <h3 class="card-title result-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            执行结果
          </h3>
          <div class="result-list">
            <div v-for="(r, idx) in results" :key="idx" class="result-item" :class="r.success ? 'success' : 'error'">
              <span class="result-icon">{{ r.success ? '✓' : '✕' }}</span>
              <span class="result-text">{{ r.message }}</span>
            </div>
          </div>
          <button class="clear-results-btn" @click="results = []">清除结果</button>
        </div>

        <!-- 常用模板 -->
        <div class="template-card">
          <h3 class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="3" x2="9" y2="21"/></svg>
            快速模板
          </h3>
          <div class="template-list">
            <button class="template-btn" @click="applyTemplate('prometheus-rule')">PrometheusRule</button>
            <button class="template-btn" @click="applyTemplate('service-monitor')">ServiceMonitor</button>
            <button class="template-btn" @click="applyTemplate('alertmanager-config')">AlertmanagerConfig</button>
            <button class="template-btn" @click="applyTemplate('configmap')">ConfigMap</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import crdApi from '@/api/cluster/extensions/crd'

const yamlContent = ref('')
const targetGroup = ref('')
const targetVersion = ref('')
const targetResource = ref('')
const groups = ref([])
const dryRunning = ref(false)
const creating = ref(false)
const results = ref([])
const crdMap = ref({}) // kind → { group, version, resource }

const lineCount = computed(() => Math.max((yamlContent.value || '').split('\n').length, 1))
const documentCount = computed(() => {
  if (!yamlContent.value.trim()) return 0
  return yamlContent.value.split(/^---$/m).filter(d => d.trim()).length
})

// 加载 API Groups + CRD Kind 映射
onMounted(async () => {
  try {
    const res = await crdApi.listCRDs({})
    if (res?.code === 0) {
      const items = res.data?.list || res.data?.items || []
      const gs = new Set(items.map(i => i.group))
      groups.value = [...gs].sort()
      // 构建 kind → resource 映射（用于自动识别）
      const map = {}
      items.forEach(i => {
        const key = `${i.kind}|${i.group}`
        map[key] = { group: i.group, version: i.version, resource: i.resource || (i.kind?.toLowerCase() + 's') }
        // 也存纯 kind 映射（无歧义时使用）
        if (!map[i.kind]) map[i.kind] = { group: i.group, version: i.version, resource: i.resource || (i.kind?.toLowerCase() + 's') }
      })
      crdMap.value = map
    }
  } catch (e) { console.error(e) }
})

// ===== 自动从 YAML 解析 apiVersion/kind 并填充字段 =====
const parseYamlMeta = (yaml) => {
  if (!yaml) return null
  const apiVersionMatch = yaml.match(/^apiVersion:\s*(.+)$/m)
  const kindMatch = yaml.match(/^kind:\s*(.+)$/m)
  if (!apiVersionMatch || !kindMatch) return null
  const apiVersion = apiVersionMatch[1].trim()
  const kind = kindMatch[1].trim()
  // 解析 group/version：'monitoring.coreos.com/v1' → group='monitoring.coreos.com', version='v1'
  // 'v1' → group='', version='v1'
  let group = '', version = apiVersion
  if (apiVersion.includes('/')) {
    const parts = apiVersion.split('/')
    group = parts.slice(0, -1).join('/')
    version = parts[parts.length - 1]
  }
  return { group, version, kind }
}

// Kind → resource 名转换（先查 CRD 映射，再用通用规则）
const kindToResource = (kind, group) => {
  // 精确匹配 kind+group
  const exact = crdMap.value[`${kind}|${group}`]
  if (exact) return exact.resource
  // 模糊匹配纯 kind
  const fuzzy = crdMap.value[kind]
  if (fuzzy) return fuzzy.resource
  // 通用英语复数化规则
  const lower = kind.toLowerCase()
  if (lower.endsWith('s') || lower.endsWith('x') || lower.endsWith('z')) return lower + 'es'
  if (lower.endsWith('y') && !['a','e','i','o','u'].includes(lower[lower.length-2])) return lower.slice(0, -1) + 'ies'
  return lower + 's'
}

// 监听 YAML 内容变化，自动填充 group/version/resource
let autoFillTimer = null
watch(yamlContent, (newVal) => {
  clearTimeout(autoFillTimer)
  autoFillTimer = setTimeout(() => {
    const meta = parseYamlMeta(newVal)
    if (meta) {
      targetGroup.value = meta.group
      targetVersion.value = meta.version
      targetResource.value = kindToResource(meta.kind, meta.group)
    }
  }, 500) // 防抖 500ms
})

// DryRun
const runDryRun = async () => {
  if (!yamlContent.value.trim()) return
  if (!targetVersion.value || !targetResource.value) {
    results.value = [{ success: false, message: '请填写 version 和 resource 字段' }]
    return
  }
  dryRunning.value = true
  results.value = []
  try {
    // 从 YAML 解析 namespace（让后端知道在哪个命名空间做校验）
    const nsMatch = yamlContent.value.match(/^  namespace:\s*(.+)$/m)
    const ns = nsMatch ? nsMatch[1].trim() : undefined
    const res = await crdApi.dryRun({
      group: targetGroup.value || undefined,
      version: targetVersion.value,
      resource: targetResource.value,
      namespace: ns,
      yaml: yamlContent.value,
      is_update: false
    })
    if (res?.code === 0) {
      // 后端返回 code:0 但 data.valid 可能为 false（DryRun 校验不通过）
      const dr = res.data
      if (dr?.valid) {
        results.value = [{ success: true, message: dr.message || 'DryRun 校验通过，YAML 合法且可创建' }]
      } else {
        const errMsg = (dr?.errors && dr.errors.length > 0) ? dr.errors.join('; ') : (dr?.message || '校验失败')
        results.value = [{ success: false, message: errMsg }]
      }
    } else {
      results.value = [{ success: false, message: res?.msg || '校验失败' }]
    }
  } catch (e) {
    results.value = [{ success: false, message: e?.msg || e?.response?.data?.msg || e?.message || '请求失败' }]
  } finally { dryRunning.value = false }
}

// 提交创建
const submitCreate = async () => {
  if (!yamlContent.value.trim()) return
  if (!targetVersion.value || !targetResource.value) {
    results.value = [{ success: false, message: '请填写 version 和 resource 字段' }]
    return
  }
  creating.value = true
  results.value = []
  try {
    const nsMatch = yamlContent.value.match(/^  namespace:\s*(.+)$/m)
    const ns = nsMatch ? nsMatch[1].trim() : undefined
    const res = await crdApi.createCR({
      group: targetGroup.value || undefined,
      version: targetVersion.value,
      resource: targetResource.value,
      namespace: ns,
      yaml: yamlContent.value
    })
    if (res?.code === 0) {
      results.value = [{ success: true, message: `创建成功: ${res.data?.name || 'OK'}` }]
    } else {
      results.value = [{ success: false, message: res?.msg || '创建失败' }]
    }
  } catch (e) {
    const details = Array.isArray(e?.details) ? e.details[0] : null
    results.value = [{ success: false, message: details || e?.msg || e?.response?.data?.msg || e?.message || '创建失败' }]
  } finally { creating.value = false }
}

// 模板
const templates = {
  'prometheus-rule': { group: 'monitoring.coreos.com', version: 'v1', resource: 'prometheusrules', yaml: `apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-alert-rule
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  groups:
    - name: custom.rules
      rules:
        - alert: HighMemoryUsage
          expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes < 0.1
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "节点内存使用率过高"
` },
  'service-monitor': { group: 'monitoring.coreos.com', version: 'v1', resource: 'servicemonitors', yaml: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  selector:
    matchLabels:
      app: my-service
  endpoints:
    - port: metrics
      interval: 30s
` },
  'alertmanager-config': { group: 'monitoring.coreos.com', version: 'v1alpha1', resource: 'alertmanagerconfigs', yaml: `apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: my-alert-config
  namespace: monitoring
spec:
  route:
    receiver: default
    groupBy: ['alertname']
  receivers:
    - name: default
` },
  'configmap': { group: '', version: 'v1', resource: 'configmaps', yaml: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  config.yaml: |
    key: value
` },
}

const applyTemplate = (name) => {
  const tpl = templates[name]
  if (tpl) {
    yamlContent.value = tpl.yaml
    targetGroup.value = tpl.group
    targetVersion.value = tpl.version
    targetResource.value = tpl.resource
  }
}

// 工具
const formatContent = () => { yamlContent.value = yamlContent.value.split('\n').map(l => l.trimEnd()).join('\n').trim() + '\n' }
const clearContent = () => { yamlContent.value = ''; results.value = [] }
const syncGutter = (e) => { const g = e.target?.parentElement?.querySelector('.wb-gutter'); if (g) g.scrollTop = e.target.scrollTop }
const insertTab = (e) => { const ta = e.target; const s = ta.selectionStart; yamlContent.value = yamlContent.value.substring(0, s) + '  ' + yamlContent.value.substring(ta.selectionEnd); nextTick(() => { ta.selectionStart = ta.selectionEnd = s + 2 }) }
</script>

<style scoped>
.yaml-workbench {
  --primary: #6366f1;
  --danger: #ef4444;
  --success: #10b981;
  --bg-page: #f8fafc;
  --bg-card: #ffffff;
  --border: #e2e8f0;
  --border-light: #f1f5f9;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --radius: 12px;
  --radius-sm: 8px;
  padding: 24px 28px;
  min-height: 100%;
  background: var(--bg-page);
}
.page-header { margin-bottom: 20px; display: flex; align-items: center; gap: 14px; }
.header-left { display: flex; align-items: center; gap: 14px; }
.page-title { font-size: 20px; font-weight: 700; color: var(--text-primary); margin: 0; display: flex; align-items: center; gap: 10px; }
.title-icon { font-size: 22px; }
.page-desc { margin: 4px 0 0; font-size: 13px; color: var(--text-muted); }

/* 两栏布局 */
.workbench-body { display: grid; grid-template-columns: 1fr 300px; gap: 20px; height: calc(100vh - 180px); }

/* 编辑器区 */
.editor-section { display: flex; flex-direction: column; background: #1b1d2e; border-radius: var(--radius); overflow: hidden; box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
.wb-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 10px 16px; background: #252840; border-bottom: 1px solid rgba(255,255,255,0.06); gap: 8px; flex-wrap: wrap; }
.wb-toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.wb-toolbar-right { display: flex; gap: 8px; }
.wb-select { padding: 6px 12px; border: 1px solid rgba(255,255,255,0.12); border-radius: 6px; background: #2a2d44; color: #c8cce0; font-size: 12px; }
.wb-select:focus { outline: none; border-color: var(--primary); }
.wb-input { padding: 6px 12px; border: 1px solid rgba(255,255,255,0.12); border-radius: 6px; background: #2a2d44; color: #c8cce0; font-size: 12px; min-width: 140px; }
.wb-input.small { min-width: 80px; max-width: 100px; }
.wb-input:focus { outline: none; border-color: var(--primary); }
.wb-btn { padding: 6px 14px; border: 1px solid rgba(255,255,255,0.12); border-radius: 6px; background: transparent; color: #8b95b0; font-size: 12px; cursor: pointer; }
.wb-btn:hover { background: rgba(255,255,255,0.06); color: #e0e6f0; }

/* 编辑器 */
.wb-editor-area { flex: 1; display: flex; overflow: hidden; }
.wb-gutter { width: 44px; min-width: 44px; padding: 14px 0; background: #181a2a; border-right: 1px solid rgba(255,255,255,0.04); overflow: hidden; text-align: right; user-select: none; }
.wb-gutter-line { height: 21px; line-height: 21px; padding-right: 10px; font-family: 'JetBrains Mono', monospace; font-size: 11px; color: #4a5073; }
.wb-textarea { flex: 1; padding: 14px 16px; border: none; resize: none; font-family: 'JetBrains Mono', monospace; font-size: 13px; line-height: 21px; background: #1e2030; color: #cdd6f4; outline: none; tab-size: 2; white-space: pre; overflow: auto; }
.wb-textarea::placeholder { color: #4a5073; }
.wb-textarea::-webkit-scrollbar { width: 8px; }
.wb-textarea::-webkit-scrollbar-track { background: transparent; }
.wb-textarea::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 4px; }

/* 状态栏 */
.wb-statusbar { display: flex; align-items: center; gap: 16px; padding: 8px 16px; background: #252840; border-top: 1px solid rgba(255,255,255,0.06); }
.wb-stat { font-size: 11px; color: #6b7394; }
.wb-stat-tag { padding: 2px 8px; background: rgba(99,102,241,0.15); color: #a5b4fc; border-radius: 3px; font-weight: 600; font-size: 10px; }

/* 右侧操作面板 */
.action-panel { display: flex; flex-direction: column; gap: 16px; overflow-y: auto; }
.action-card { background: var(--bg-card); border-radius: var(--radius); padding: 18px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); }
.card-title { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0 0 8px; display: flex; align-items: center; gap: 8px; }
.card-desc { font-size: 12px; color: var(--text-secondary); margin: 0 0 14px; line-height: 1.5; }
.action-primary-btn { width: 100%; padding: 10px; border: none; border-radius: var(--radius-sm); background: linear-gradient(135deg, var(--success), #059669); color: #fff; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.action-primary-btn:hover:not(:disabled) { box-shadow: 0 4px 12px rgba(16,185,129,0.3); transform: translateY(-1px); }
.action-primary-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.action-submit-btn { width: 100%; padding: 10px; border: none; border-radius: var(--radius-sm); background: linear-gradient(135deg, var(--primary), #8b5cf6); color: #fff; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.action-submit-btn:hover:not(:disabled) { box-shadow: 0 4px 12px rgba(99,102,241,0.3); transform: translateY(-1px); }
.action-submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* 结果面板 */
.result-panel { background: var(--bg-card); border-radius: var(--radius); padding: 18px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); }
.result-title { margin-bottom: 12px; }
.result-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }
.result-item { display: flex; align-items: flex-start; gap: 8px; padding: 8px 12px; border-radius: 6px; font-size: 12px; line-height: 1.5; word-break: break-all; }
.result-item.success { background: #ecfdf5; color: #065f46; }
.result-item.error { background: #fef2f2; color: #991b1b; }
.result-icon { font-weight: 700; font-size: 13px; flex-shrink: 0; }
.result-text { flex: 1; }
.clear-results-btn { width: 100%; padding: 6px; border: 1px solid var(--border); border-radius: 6px; background: transparent; color: var(--text-muted); font-size: 11px; cursor: pointer; }
.clear-results-btn:hover { background: #f5f7fa; color: var(--text-secondary); }

/* 模板卡片 */
.template-card { background: var(--bg-card); border-radius: var(--radius); padding: 18px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.04); }
.template-list { display: flex; flex-direction: column; gap: 6px; margin-top: 12px; }
.template-btn { width: 100%; padding: 9px 14px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: #fff; color: var(--text-primary); font-size: 12px; text-align: left; cursor: pointer; transition: all 0.15s; }
.template-btn:hover { border-color: var(--primary); color: var(--primary); background: rgba(99,102,241,0.04); }

@media (max-width: 1024px) {
  .workbench-body { grid-template-columns: 1fr; height: auto; }
  .editor-section { min-height: 500px; }
}
</style>

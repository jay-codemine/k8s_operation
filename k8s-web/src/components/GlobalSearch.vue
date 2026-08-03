<template>
  <teleport to="body">
    <transition name="gs-fade">
      <div v-if="visible" class="gs-overlay" @click.self="close">
        <div class="gs-modal">
          <div class="gs-input-row">
            <svg class="gs-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/>
              <line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
            <input
              ref="inputRef"
              v-model="keyword"
              class="gs-input"
              placeholder="搜索页面、流水线... (↑↓ 选择，Enter 跳转，Esc 关闭)"
              @keydown.down.prevent="move(1)"
              @keydown.up.prevent="move(-1)"
              @keydown.enter.prevent="confirm"
              @keydown.esc="close"
            />
          </div>

          <div class="gs-results">
            <div v-if="!keyword.trim()" class="gs-hint">
              <div class="gs-hint-title">快捷导航</div>
              <div
                v-for="(item, i) in hotPages"
                :key="'hot-' + i"
                class="gs-item"
                :class="{ 'gs-active': flatIndex(item) === activeIndex }"
                @click="go(item)"
                @mouseenter="activeIndex = flatIndex(item)"
              >
                <span class="gs-item-icon">📄</span>
                <span class="gs-item-label">{{ item.label }}</span>
                <span class="gs-item-group">{{ item.group }}</span>
              </div>
            </div>

            <template v-else>
              <div v-if="pageResults.length" class="gs-group-label">页面</div>
              <div
                v-for="(item, i) in pageResults"
                :key="'page-' + i"
                class="gs-item"
                :class="{ 'gs-active': flatIndex(item) === activeIndex }"
                @click="go(item)"
                @mouseenter="activeIndex = flatIndex(item)"
              >
                <span class="gs-item-icon">📄</span>
                <span class="gs-item-label" v-html="highlight(item.label)"></span>
                <span class="gs-item-group">{{ item.group }}</span>
              </div>

              <div v-if="pipelineResults.length" class="gs-group-label">流水线</div>
              <div
                v-for="item in pipelineResults"
                :key="'pl-' + item.id"
                class="gs-item"
                :class="{ 'gs-active': flatIndex(item) === activeIndex }"
                @click="go(item)"
                @mouseenter="activeIndex = flatIndex(item)"
              >
                <span class="gs-item-icon">🚀</span>
                <span class="gs-item-label" v-html="highlight(item.name)"></span>
                <span class="gs-item-group">{{ item.description || item.git_repo || '流水线' }}</span>
              </div>

              <div v-if="!pageResults.length && !pipelineResults.length && !searching" class="gs-empty">
                未找到「{{ keyword }}」相关内容
              </div>
              <div v-if="searching" class="gs-empty">搜索中...</div>
            </template>
          </div>

          <div class="gs-footer">
            <span class="gs-kbd">↑↓</span> 选择
            <span class="gs-kbd">Enter</span> 跳转
            <span class="gs-kbd">Esc</span> 关闭
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { getPipelines } from '@/api/platform/pipeline'

const props = defineProps({
  visible: Boolean,
})
const emit = defineEmits(['update:visible'])

const router = useRouter()
const keyword = ref('')
const searching = ref(false)
const pipelineResults = ref([])
const activeIndex = ref(0)
const inputRef = ref(null)

// ==================== 静态页面索引 ====================
// 与 Layout 侧边栏菜单保持一致；keywords 支持英文/拼音模糊匹配
const PAGE_INDEX = [
  { group: '首页', label: '工作台', path: '/dashboard', keywords: 'home dashboard shouye' },
  { group: '资源中心', label: '集群管理', path: '/clusters', keywords: 'cluster ji ziyuan' },
  { group: '资源中心', label: '平台健康', path: '/platform/health', keywords: 'health jiankang' },
  { group: '资源中心', label: '智能运维', path: '/platform/aiops', keywords: 'aiops ai zhineng' },
  { group: 'CI/CD', label: '应用总览', path: '/cicd/apps', keywords: 'app application yingyong' },
  { group: 'CI/CD', label: '快速接入', path: '/cicd/quick-onboard', keywords: 'onboard kuaisu jieru' },
  { group: 'CI/CD', label: '流水线管理', path: '/cicd/pipelines', keywords: 'pipeline liushuixian' },
  { group: 'CI/CD', label: '创建流水线', path: '/cicd/pipelines/create', keywords: 'pipeline create chuangjian' },
  { group: 'CI/CD', label: 'GitOps 发布', path: '/cicd/gitops/releases', keywords: 'gitops argo fabu' },
  { group: 'CI/CD', label: '构建记录', path: '/cicd/build-records', keywords: 'build record goujian jilu' },
  { group: 'CI/CD', label: '制品库', path: '/cicd/artifacts', keywords: 'artifact zhipin' },
  { group: 'CI/CD', label: '发布管理', path: '/cicd/releases', keywords: 'release fabu guanli' },
  { group: 'CI/CD', label: '镜像晋级', path: '/cicd/promotion', keywords: 'promotion jinjing' },
  { group: 'CI/CD', label: '发布历史', path: '/cicd/release-history', keywords: 'release history lishi' },
  { group: 'CI/CD', label: '审批工单', path: '/cicd/approvals', keywords: 'approval shenpi gongdan' },
  { group: 'CI/CD', label: '环境管理', path: '/cicd/environments', keywords: 'environment huanjing' },
  { group: 'CI/CD', label: '审批策略', path: '/cicd/approval-policy', keywords: 'approval policy shenpi celue' },
  { group: 'CI/CD', label: '流水线模板', path: '/cicd/templates', keywords: 'template muban' },
  { group: 'CI/CD', label: '构建节点', path: '/cicd/agents', keywords: 'agent jenkins node goujian jiedian' },
  { group: '镜像管理', label: '镜像仓库管理', path: '/images/repositories', keywords: 'image registry repo jingxiang cangku' },
  { group: '镜像管理', label: '镜像浏览', path: '/images/browse', keywords: 'image browse liulan' },
  { group: '镜像管理', label: '清理策略', path: '/images/cleanup', keywords: 'cleanup qingli celue' },
  { group: '监控中心', label: '监控总览', path: '/monitoring', keywords: 'monitor overview jiankong zonglan' },
  { group: '监控中心', label: '数据源管理', path: '/monitoring/datasources', keywords: 'datasource prometheus shujuyuan' },
  { group: '监控中心', label: '告警规则', path: '/monitoring/alert-rules', keywords: 'alert rule gaojing guize' },
  { group: '监控中心', label: '告警事件', path: '/monitoring/alert-events', keywords: 'alert event gaojing shijian' },
  { group: '监控中心', label: '通知渠道', path: '/monitoring/notify-channels', keywords: 'notify channel tongzhi qudao dingtalk' },
  { group: '监控中心', label: '静默规则', path: '/monitoring/silence-rules', keywords: 'silence jingmo' },
  { group: 'CRD 扩展', label: 'CRD 资源管理', path: '/extensions/crd', keywords: 'crd custom resource' },
  { group: 'CRD 扩展', label: 'CR 实例管理', path: '/extensions/cr-instances', keywords: 'cr instance shili' },
  { group: 'CRD 扩展', label: 'YAML 工作台', path: '/extensions/yaml-workbench', keywords: 'yaml workbench gongzuotai' },
  { group: '平台管理', label: '用户与组织', path: '/admin/users', keywords: 'user yonghu zuzhi' },
  { group: '平台管理', label: '权限与角色', path: '/admin/roles', keywords: 'role rbac quanxian juese' },
  { group: '平台管理', label: '身份认证', path: '/admin/identity', keywords: 'identity ldap shenfen' },
  { group: '平台管理', label: '审批中心', path: '/admin/approvals', keywords: 'approval shenpi zhongxin' },
  { group: '平台管理', label: '审计中心', path: '/admin/audit', keywords: 'audit shenji' },
  { group: '平台管理', label: '服务账号', path: '/admin/service-accounts', keywords: 'service account fuwu zhanghao' },
  { group: '平台管理', label: '系统设置', path: '/admin/settings', keywords: 'settings shezhi' },
  { group: '其他', label: '应用商城', path: '/platform/appstore', keywords: 'appstore store shangcheng' },
]

const hotPages = computed(() => [
  PAGE_INDEX.find(p => p.path === '/dashboard'),
  PAGE_INDEX.find(p => p.path === '/clusters'),
  PAGE_INDEX.find(p => p.path === '/cicd/pipelines'),
  PAGE_INDEX.find(p => p.path === '/monitoring/alert-events'),
].filter(Boolean))

// ==================== 过滤逻辑 ====================
const pageResults = computed(() => {
  const q = keyword.value.toLowerCase().trim()
  if (!q) return []
  return PAGE_INDEX.filter(p =>
    p.label.toLowerCase().includes(q) ||
    p.group.toLowerCase().includes(q) ||
    p.path.toLowerCase().includes(q) ||
    (p.keywords || '').includes(q)
  ).slice(0, 8)
})

// 当前可视的扁平结果列表（键盘导航用）
const flatList = computed(() => {
  const q = keyword.value.trim()
  if (!q) return hotPages.value
  return [...pageResults.value, ...pipelineResults.value]
})

const flatIndex = (item) => flatList.value.indexOf(item)

// ==================== 流水线远程搜索 ====================
let debounceTimer = null
watch(keyword, (val) => {
  activeIndex.value = 0
  clearTimeout(debounceTimer)
  const q = val.trim()
  if (!q) {
    pipelineResults.value = []
    searching.value = false
    return
  }
  searching.value = true
  debounceTimer = setTimeout(async () => {
    try {
      const res = await getPipelines({ keyword: q, page: 1, page_size: 5 })
      if (res?.code === 0) {
        pipelineResults.value = (res.data?.list || res.data?.items || []).slice(0, 5)
      }
    } catch {
      pipelineResults.value = []
    } finally {
      searching.value = false
    }
  }, 300)
})

// ==================== 交互 ====================
const highlight = (text) => {
  const q = keyword.value.trim()
  if (!q || !text) return text
  const idx = String(text).toLowerCase().indexOf(q.toLowerCase())
  if (idx < 0) return text
  const s = String(text)
  return s.slice(0, idx) + '<mark>' + s.slice(idx, idx + q.length) + '</mark>' + s.slice(idx + q.length)
}

const move = (delta) => {
  const len = flatList.value.length
  if (!len) return
  activeIndex.value = (activeIndex.value + delta + len) % len
}

const confirm = () => {
  const item = flatList.value[activeIndex.value]
  if (item) go(item)
}

const go = (item) => {
  close()
  if (item.path) {
    router.push(item.path)
  } else if (item.id) {
    router.push(`/cicd/pipelines/${item.id}`)
  }
}

const close = () => {
  emit('update:visible', false)
}

// 打开时聚焦输入框并重置状态
watch(() => props.visible, async (v) => {
  if (v) {
    keyword.value = ''
    pipelineResults.value = []
    activeIndex.value = 0
    await nextTick()
    inputRef.value?.focus()
  }
})
</script>

<style scoped>
.gs-overlay {
  position: fixed; inset: 0; z-index: 1200;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
  display: flex; justify-content: center;
  padding-top: 14vh;
}

.gs-modal {
  width: 560px; max-width: 92vw;
  max-height: 60vh;
  display: flex; flex-direction: column;
  background: #fff; border-radius: 14px;
  box-shadow: 0 24px 80px rgba(0,0,0,.25);
  overflow: hidden;
  height: fit-content;
}

.gs-input-row {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 18px;
  border-bottom: 1px solid #f1f5f9;
}
.gs-icon { width: 18px; height: 18px; color: #94a3b8; flex-shrink: 0; }
.gs-input {
  flex: 1; border: none; outline: none;
  font-size: 15px; color: #1e293b;
  background: transparent;
}
.gs-input::placeholder { color: #a0aec0; font-size: 13px; }

.gs-results { overflow-y: auto; padding: 6px 0; }

.gs-hint-title, .gs-group-label {
  padding: 8px 18px 4px;
  font-size: 11px; font-weight: 600; color: #a0aec0;
}

.gs-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 18px; cursor: pointer;
  transition: background .12s;
}
.gs-item:hover, .gs-active { background: #f1f5ff; }
.gs-item-icon { font-size: 14px; flex-shrink: 0; }
.gs-item-label { font-size: 14px; color: #1e293b; }
.gs-item-label :deep(mark) { background: #fef08a; color: inherit; border-radius: 2px; padding: 0 1px; }
.gs-item-group {
  margin-left: auto; flex-shrink: 0;
  font-size: 11px; color: #a0aec0;
  max-width: 45%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.gs-empty { padding: 36px 20px; text-align: center; color: #a0aec0; font-size: 13px; }

.gs-footer {
  display: flex; align-items: center; gap: 6px;
  padding: 10px 18px;
  border-top: 1px solid #f1f5f9;
  font-size: 11px; color: #94a3b8;
}
.gs-kbd {
  padding: 2px 6px; border: 1px solid #e2e8f0; border-radius: 4px;
  background: #f8fafc; font-size: 10px; color: #64748b;
  font-family: monospace;
}

.gs-fade-enter-active { transition: opacity .15s ease; }
.gs-fade-leave-active { transition: opacity .1s ease; }
.gs-fade-enter-from, .gs-fade-leave-to { opacity: 0; }
</style>

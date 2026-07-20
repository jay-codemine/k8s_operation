<template>
  <div class="promotion-page">
    <!-- 页头 -->
    <div class="page-head">
      <div class="ph-title">
        <h2>镜像晋级</h2>
        <span class="ph-sub">一次构建，跨环境晋级同一不可变镜像（build once, promote everywhere）</span>
      </div>
      <router-link class="ph-link" to="/cicd/environments">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
        环境管理
      </router-link>
    </div>

    <!-- 流水线选择器 -->
    <div class="selector-bar">
      <label class="sel-label">选择流水线</label>
      <div class="sel-search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input v-model="keyword" class="sel-input" placeholder="搜索流水线名称..." @input="onSearch" />
      </div>
      <select v-model.number="selectedId" class="sel-select" @change="onSelect">
        <option :value="0" disabled>请选择流水线</option>
        <option v-for="p in filteredPipelines" :key="p.id" :value="p.id">
          {{ p.name }}{{ p.language_type ? '（' + p.language_type + '）' : '' }}
        </option>
      </select>
      <span class="sel-count">共 {{ pipelines.length }} 条流水线</span>
    </div>

    <!-- 空状态：未选择 -->
    <div v-if="!selectedId" class="page-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
        <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
        <line x1="12" y1="22.08" x2="12" y2="12"/>
      </svg>
      <p>请先选择一条流水线</p>
      <span>选择流水线后即可查看其 dev → test → staging → prod 晋级链，并将同一镜像逐级晋级发布。</span>
    </div>

    <!-- 晋级面板 -->
    <div v-else class="panel-wrap">
      <PromotionPanel :key="selectedId" :pipeline-id="selectedId" :pipeline="selectedPipeline" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PromotionPanel from '@/components/cicd/PromotionPanel.vue'
import { getPipelines, getPipelineDetail } from '@/api/cicd'

const route = useRoute()
const router = useRouter()

const pipelines = ref([])
const keyword = ref('')
const selectedId = ref(0)
const selectedPipeline = ref({})

const filteredPipelines = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return pipelines.value
  return pipelines.value.filter(p => (p.name || '').toLowerCase().includes(kw))
})

const onSearch = () => {}

const parseList = (res) => {
  const d = res?.data || res || {}
  return d.list || d.items || d.data?.list || (Array.isArray(d) ? d : [])
}

const loadPipelines = async () => {
  try {
    const res = await getPipelines({ page: 1, page_size: 1000 })
    pipelines.value = parseList(res)
  } catch (e) {
    console.warn('加载流水线失败', e)
    pipelines.value = []
  }
}

const onSelect = async () => {
  if (!selectedId.value) return
  // 先用列表项兜底，再拉详情补全部署目标默认值
  selectedPipeline.value = pipelines.value.find(p => p.id === selectedId.value) || {}
  // 同步到 URL query，便于分享/刷新保持
  router.replace({ query: { ...route.query, pipeline_id: selectedId.value } })
  try {
    const res = await getPipelineDetail(selectedId.value)
    const detail = res?.data || res
    if (detail && (detail.id || detail.ID)) selectedPipeline.value = detail
  } catch {
    // 详情失败不阻塞，PromotionPanel 会走后端全局环境默认
  }
}

const init = async () => {
  await loadPipelines()
  const qid = Number(route.query.pipeline_id || 0)
  if (qid > 0 && pipelines.value.some(p => p.id === qid)) {
    selectedId.value = qid
    await onSelect()
  }
}

init()
</script>

<style scoped>
.promotion-page {
  padding: 20px 24px;
  min-height: 100%;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;
}
.ph-title h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1a202c;
}
.ph-sub {
  display: block;
  margin-top: 4px;
  font-size: 13px;
  color: #718096;
}
.ph-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #fff;
  color: #4a5568;
  font-size: 13px;
  text-decoration: none;
  transition: all 0.15s;
}
.ph-link:hover {
  border-color: #6366f1;
  color: #6366f1;
}
.ph-link svg {
  width: 15px;
  height: 15px;
}

.selector-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid #edf2f7;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  margin-bottom: 18px;
  flex-wrap: wrap;
}
.sel-label {
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}
.sel-search {
  position: relative;
  display: flex;
  align-items: center;
}
.sel-search svg {
  position: absolute;
  left: 10px;
  width: 15px;
  height: 15px;
  color: #a0aec0;
}
.sel-input {
  padding: 8px 12px 8px 32px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  width: 220px;
  outline: none;
}
.sel-input:focus {
  border-color: #6366f1;
}
.sel-select {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 13px;
  min-width: 260px;
  outline: none;
  background: #fff;
}
.sel-select:focus {
  border-color: #6366f1;
}
.sel-count {
  font-size: 12px;
  color: #a0aec0;
  margin-left: auto;
}

.page-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 72px 24px;
  background: #fff;
  border: 1px dashed #e2e8f0;
  border-radius: 12px;
  text-align: center;
}
.page-empty svg {
  width: 56px;
  height: 56px;
  color: #cbd5e0;
  margin-bottom: 16px;
}
.page-empty p {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: #4a5568;
}
.page-empty span {
  font-size: 13px;
  color: #a0aec0;
  max-width: 460px;
}

.panel-wrap {
  background: #fff;
  border: 1px solid #edf2f7;
  border-radius: 12px;
  padding: 18px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
</style>

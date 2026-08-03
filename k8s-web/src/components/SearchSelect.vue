<template>
  <div class="ss-wrap" ref="wrapRef">
    <div class="ss-trigger" :class="{ 'ss-open': open, 'ss-disabled': disabled }" @click="toggle">
      <span v-if="modelValue" class="ss-val">{{ selectedLabel }}</span>
      <span v-else class="ss-placeholder">{{ placeholder || '请选择' }}</span>
      <span class="ss-arrow">▼</span>
    </div>
    <transition name="ss-drop">
      <div v-if="open" class="ss-drop">
        <div class="ss-search" v-if="searchable">
          <input ref="searchRef" v-model="keyword" class="ss-input" :placeholder="searchPlaceholder || '搜索...'" @input="onSearch" @click.stop />
        </div>
        <div class="ss-list" ref="listRef">
          <template v-if="recentItems.length && !keyword">
            <div class="ss-group-label">⭐ 最近使用</div>
            <div v-for="item in recentItems" :key="'recent-' + item.value" class="ss-option" :class="{ 'ss-sel': modelValue === item.value }" @click.stop="select(item)">
              <span class="ss-opt-label">{{ item.label }}</span>
              <span v-if="item.sub" class="ss-opt-sub">{{ item.sub }}</span>
            </div>
            <div v-if="filteredOptions.length" class="ss-group-label">全部</div>
          </template>
          <div v-for="item in filteredOptions" :key="item.value" class="ss-option" :class="{ 'ss-sel': modelValue === item.value }" @click.stop="select(item)">
            <span class="ss-opt-label">{{ item.label }}</span>
            <span v-if="item.sub" class="ss-opt-sub">{{ item.sub }}</span>
          </div>
          <div v-if="!filteredOptions.length" class="ss-empty">{{ keyword ? '无匹配结果' : emptyText }}</div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  modelValue: [String, Number],
  options: { type: Array, default: () => [] },
  placeholder: String,
  searchable: { type: Boolean, default: true },
  searchPlaceholder: String,
  disabled: Boolean,
  emptyText: { type: String, default: '暂无可选项' },
  recentKey: String   // localStorage key for recent items
})

const emit = defineEmits(['update:modelValue', 'search'])

const open = ref(false)
const keyword = ref('')
const searchRef = ref(null)
const wrapRef = ref(null)
const listRef = ref(null)
const recentItems = ref([])

const selectedLabel = computed(() => {
  const opt = props.options.find(o => o.value === props.modelValue)
  return opt ? opt.label : props.modelValue
})

const filteredOptions = computed(() => {
  if (!props.searchable || !keyword.value) return props.options
  const q = keyword.value.toLowerCase().trim()
  if (!q) return props.options
  return props.options.filter(o =>
    String(o.label).toLowerCase().includes(q) ||
    String(o.sub || '').toLowerCase().includes(q) ||
    String(o.value).toLowerCase().includes(q)
  )
})

function onSearch() {
  emit('search', keyword.value)
}

function select(item) {
  emit('update:modelValue', item.value)
  open.value = false
  keyword.value = ''
  // save to recent
  if (props.recentKey) {
    const key = props.recentKey
    const stored = JSON.parse(localStorage.getItem(key) || '[]')
    const idx = stored.findIndex(r => r.value === item.value)
    if (idx >= 0) stored.splice(idx, 1)
    stored.unshift({ value: item.value, label: item.label, sub: item.sub })
    localStorage.setItem(key, JSON.stringify(stored.slice(0, 5)))
    recentItems.value = stored.slice(0, 5)
  }
}

function toggle() {
  if (props.disabled) return
  open.value = !open.value
  if (open.value) { keyword.value = ''; nextTick(() => searchRef.value?.focus()) }
}

function onClickOutside(e) {
  if (wrapRef.value && !wrapRef.value.contains(e.target)) open.value = false
}

onMounted(() => {
  document.addEventListener('click', onClickOutside)
  if (props.recentKey) {
    recentItems.value = JSON.parse(localStorage.getItem(props.recentKey) || '[]').slice(0, 5)
  }
})
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
/* 与平台表单控件（.form-input/.form-select）一致的浅色风格；
   暗色主题由 App.vue 的 [data-theme="dark"] .ss-* 全局规则覆盖 */
.ss-wrap { position:relative; width:100%; flex:1; min-width:0; }
.ss-trigger {
  display:flex; align-items:center; justify-content:space-between;
  width:100%; padding:12px 16px; border:2px solid #e2e8f0;
  border-radius:10px; background:#f7fafc; color:#2d3748;
  font-size:14px; cursor:pointer; transition:all .3s;
}
.ss-trigger:hover:not(.ss-disabled) { border-color:#cbd5e1; }
.ss-open { border-color:#4299e1 !important; background:#fff; box-shadow:0 0 0 4px rgba(66,153,225,.1); }
.ss-disabled { opacity:.6; cursor:not-allowed; }
.ss-placeholder { color:#a0aec0; }
.ss-val { color:#2d3748; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.ss-arrow { font-size:10px; color:#a0aec0; margin-left:8px; flex-shrink:0; }

.ss-drop {
  position:absolute; top:calc(100% + 6px); left:0; right:0; z-index:500;
  background:#fff; border:1px solid #e2e8f0; border-radius:10px;
  box-shadow:0 12px 40px rgba(0,0,0,.12); overflow:hidden;
}
.ss-search { padding:10px 12px; border-bottom:1px solid #e2e8f0; }
.ss-input {
  width:100%; padding:8px 12px; border-radius:8px;
  border:1px solid #e2e8f0; background:#f7fafc;
  color:#2d3748; font-size:13px; outline:none; transition:border-color .2s;
}
.ss-input:focus { border-color:#4299e1; background:#fff; }
.ss-input::placeholder { color:#a0aec0; }
.ss-list { max-height:240px; overflow-y:auto; padding:4px 0; }
.ss-group-label { padding:6px 16px; font-size:11px; color:#a0aec0; font-weight:600; }
.ss-option {
  display:flex; align-items:center; justify-content:space-between;
  padding:10px 16px; font-size:13px; cursor:pointer; transition:background .15s;
}
.ss-option:hover { background:#edf2f7; }
.ss-sel { background:linear-gradient(135deg, #ebf8ff 0%, #e6fffa 100%); font-weight:600; }
.ss-opt-label { color:#2d3748; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.ss-opt-sub { font-size:11px; color:#a0aec0; margin-left:12px; flex-shrink:0; }
.ss-empty { padding:24px; text-align:center; color:#a0aec0; font-size:13px; }

.ss-drop-enter-active { transition:all .15s ease; }
.ss-drop-leave-active { transition:all .1s ease; }
.ss-drop-enter-from, .ss-drop-leave-to { opacity:0; transform:translateY(-6px); }
</style>

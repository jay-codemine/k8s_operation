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
          <div v-if="recentItems.length && !keyword" class="ss-group-label">⭐ 最近使用</div>
          <div v-for="item in (keyword ? filteredOptions : (recentItems.length && !keyword ? recentItems : filteredOptions))" :key="item.value" class="ss-option" :class="{ 'ss-sel': modelValue === item.value }" @click.stop="select(item)">
            <span class="ss-opt-label">{{ item.label }}</span>
            <span v-if="item.sub" class="ss-opt-sub">{{ item.sub }}</span>
          </div>
          <div v-if="!keyword && recentItems.length && filteredOptions.length" class="ss-group-label">全部</div>
          <div v-if="keyword || !recentItems.length" v-for="item in filteredOptions" :key="item.value" class="ss-option" :class="{ 'ss-sel': modelValue === item.value }" @click.stop="select(item)">
            <span class="ss-opt-label">{{ item.label }}</span>
            <span v-if="item.sub" class="ss-opt-sub">{{ item.sub }}</span>
          </div>
          <div v-if="!filteredOptions.length && !recentItems.length" class="ss-empty">无匹配结果</div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  modelValue: [String, Number],
  options: { type: Array, default: () => [] },
  placeholder: String,
  searchable: { type: Boolean, default: true },
  searchPlaceholder: String,
  disabled: Boolean,
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
  const q = keyword.value.toLowerCase()
  return props.options.filter(o =>
    o.label.toLowerCase().includes(q) ||
    (o.sub || '').toLowerCase().includes(q)
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
.ss-wrap { position:relative; width:100%; }
.ss-trigger {
  display:flex; align-items:center; justify-content:space-between;
  height:40px; padding:0 12px; border:1px solid rgba(255,255,255,.12);
  border-radius:8px; background:rgba(255,255,255,.05); color:#e2e8f0;
  font-size:13px; cursor:pointer; transition:border-color .15s;
}
.ss-trigger:hover:not(.ss-disabled) { border-color:rgba(58,132,255,.4); }
.ss-open { border-color:#3A84FF !important; box-shadow:0 0 0 3px rgba(58,132,255,.12); }
.ss-disabled { opacity:.5; cursor:not-allowed; }
.ss-placeholder { color:#5a6273; }
.ss-val { color:#e2e8f0; }
.ss-arrow { font-size:10px; color:#5a6273; margin-left:8px; }

.ss-drop {
  position:absolute; top:44px; left:0; right:0; z-index:500;
  background:#1e293b; border:1px solid #334155; border-radius:10px;
  box-shadow:0 12px 40px rgba(0,0,0,.4); overflow:hidden;
}
.ss-search { padding:8px 10px; border-bottom:1px solid #334155; }
.ss-input {
  width:100%; height:32px; padding:0 10px; border-radius:6px;
  border:1px solid #334155; background:rgba(255,255,255,.05);
  color:#e2e8f0; font-size:12px; outline:none;
}
.ss-input:focus { border-color:#3A84FF; }
.ss-input::placeholder { color:#5a6273; }
.ss-list { max-height:240px; overflow-y:auto; padding:4px 0; }
.ss-group-label { padding:6px 14px; font-size:11px; color:#64748b; font-weight:600; }
.ss-option {
  display:flex; align-items:center; justify-content:space-between;
  padding:8px 14px; font-size:13px; cursor:pointer; transition:background .1s;
}
.ss-option:hover { background:rgba(58,132,255,.08); }
.ss-sel { background:rgba(58,132,255,.12); color:#60a5fa; font-weight:600; }
.ss-opt-label { color:#e2e8f0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.ss-opt-sub { font-size:11px; color:#64748b; margin-left:12px; flex-shrink:0; }
.ss-empty { padding:24px; text-align:center; color:#64748b; font-size:13px; }

.ss-drop-enter-active { transition:all .15s ease; }
.ss-drop-leave-active { transition:all .1s ease; }
.ss-drop-enter-from, .ss-drop-leave-to { opacity:0; transform:translateY(-6px); }
</style>

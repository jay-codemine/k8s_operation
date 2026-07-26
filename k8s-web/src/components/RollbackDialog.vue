<template>
  <div v-if="visible" class="rb-overlay" @click.self="$emit('close')">
    <div class="rb-dialog">
      <!-- Header -->
      <div class="rb-hdr">
        <div class="rb-hdr-left">
          <span class="rb-hdr-icon"><AppIcon name="sync" size="18"/></span>
          <span>{{ batchMode ? `批量${batchAction}` : '版本回滚' }}</span>
          <span v-if="resourceName && !batchMode" class="rb-hdr-name">{{ resourceName }}</span>
        </div>
        <button class="rb-close" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <!-- Version Comparison -->
      <div class="rb-compare" v-if="!batchMode && selectedRevision">
        <div class="rb-card rb-card-from">
          <div class="rb-card-tag">当前运行</div>
          <div class="rb-card-meta">
            <div class="rb-meta-row"><span class="rb-meta-lbl">镜像</span><span class="rb-meta-val">{{ currentInfo?.image || '—' }}</span></div>
            <div class="rb-meta-row"><span class="rb-meta-lbl">副本</span><span class="rb-meta-val">{{ currentInfo?.replicas || '—' }}</span></div>
            <div class="rb-meta-row"><span class="rb-meta-lbl">部署</span><span class="rb-meta-val">{{ currentInfo?.deployedAt || '—' }}</span></div>
          </div>
        </div>
        <div class="rb-arrow">
          <div class="rb-arrow-line"></div>
          <div class="rb-arrow-head">→</div>
        </div>
        <div class="rb-card rb-card-to">
          <div class="rb-card-tag to">回滚目标</div>
          <div class="rb-card-meta">
            <div class="rb-meta-row"><span class="rb-meta-lbl">版本</span><span class="rb-meta-val">{{ selectedRevision?.name || '—' }}</span></div>
            <div class="rb-meta-row"><span class="rb-meta-lbl">镜像</span><span class="rb-meta-val">{{ selectedRevision?.image || '—' }}</span></div>
            <div class="rb-meta-row"><span class="rb-meta-lbl">创建</span><span class="rb-meta-val">{{ selectedRevision?.createdAt || '—' }}</span></div>
          </div>
        </div>
      </div>

      <!-- Batch List -->
      <div class="rb-batch" v-if="batchMode && selectedItems.length">
        <div class="rb-batch-hint">{{ selectedItems.length }} 个应用将{{ batchAction }}到上一版本</div>
        <div v-for="item in selectedItems" :key="item.name" class="rb-batch-item">
          <span class="rb-batch-name">{{ item.name }}</span>
          <span class="rb-batch-info" v-if="item.currentImage && item.targetImage">{{ item.currentImage }} → {{ item.targetImage }}</span>
          <span class="rb-batch-info" v-else>{{ item.currentImage || item.targetImage || '' }}</span>
        </div>
      </div>

      <!-- Revision Table -->
      <div class="rb-table-area" v-if="!batchMode">
        <div class="rb-table-hint">选择要回滚到的历史版本</div>
        <div class="rb-table-scroll">
          <table class="rb-table">
            <thead>
              <tr><th></th><th>版本号</th><th>Revision</th><th>镜像</th><th>副本</th><th>时间</th></tr>
            </thead>
            <tbody>
              <tr v-for="(rev, idx) in revisions" :key="rev.name"
                  :class="{ 'rb-row-sel': selectedRevision?.name === rev.name, 'rb-row-cur': idx === 0 }"
                  @click="selectRevision(rev)">
                <td><div class="rb-radio" :class="{ on: selectedRevision?.name === rev.name }"><span v-if="selectedRevision?.name === rev.name"></span></div></td>
                <td><span class="rb-badge" :class="idx===0?'rb-badge-latest':''">{{ idx === 0 ? 'Latest' : 'v'+(revisions.length-idx) }}</span></td>
                <td class="rb-td-mono">{{ rev.name?.substring(0,40) || '—' }}</td>
                <td class="rb-td-img">{{ rev.image || '—' }}</td>
                <td>{{ rev.replicas || '—' }}</td>
                <td class="rb-td-time">{{ rev.createdAt || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Loading -->
      <div class="rb-loading" v-if="loading">加载版本历史…</div>

      <!-- Footer -->
      <div class="rb-ftr">
        <div class="rb-ftr-note" v-if="!batchMode">回滚将触发滚动更新，不影响业务可用性</div>
        <div class="rb-ftr-btns">
          <button class="rb-btn-cancel" @click="$emit('close')">取消</button>
          <button class="rb-btn-go" :disabled="!canConfirm || confirming" @click="onConfirm">
            <span class="rb-btn-spin" v-if="confirming"></span>
            {{ confirming ? `${batchAction}中…` : batchMode ? `确认${batchAction} ${selectedItems.length} 个` : `回滚到此版本` }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  visible: Boolean, loading: Boolean, confirming: Boolean,
  resourceName: String, currentInfo: Object,
  revisions: { type: Array, default: () => [] },
  selectedRevision: Object, batchMode: Boolean,
  batchAction: { type: String, default: '回滚' },
  selectedItems: { type: Array, default: () => [] }
})

const emit = defineEmits(['close', 'select', 'confirm'])
const canConfirm = computed(() => props.batchMode ? props.selectedItems.length > 0 : !!props.selectedRevision)
function selectRevision(rev) { emit('select', rev) }
function onConfirm() { emit('confirm') }
</script>

<style scoped>
.rb-overlay { position:fixed; inset:0; z-index:1000; background:rgba(15,23,42,.55); backdrop-filter:blur(6px); display:flex; align-items:center; justify-content:center; padding:20px; }
.rb-dialog { width:100%; max-width:800px; max-height:88vh; background:#fff; border-radius:18px; box-shadow:0 30px 100px rgba(0,0,0,.25); display:flex; flex-direction:column; overflow:hidden; }

/* header */
.rb-hdr { display:flex; align-items:center; justify-content:space-between; padding:20px 24px; border-bottom:1px solid #f1f5f9; flex-shrink:0; }
.rb-hdr-left { display:flex; align-items:center; gap:8px; font-size:17px; font-weight:700; color:#0f172a; }
.rb-hdr-icon { color:#f59e0b; }
.rb-hdr-name { font-weight:400; color:#64748b; font-size:14px; }
.rb-hdr-name::before { content:'·'; margin:0 6px; }
.rb-close { width:32px; height:32px; border-radius:8px; border:none; background:#f1f5f9; color:#64748b; cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .15s; }
.rb-close:hover { background:#e2e8f0; color:#0f172a; }

/* compare cards */
.rb-compare { display:flex; align-items:stretch; gap:20px; padding:20px 24px; border-bottom:1px solid #f1f5f9; }
.rb-card { flex:1; border-radius:14px; padding:16px 18px; position:relative; }
.rb-card-from { border:1.5px solid #e2e8f0; background:#f8fafc; }
.rb-card-to { border:1.5px solid #fbbf24; background:#fffbeb; }
.rb-card-tag { display:inline-block; padding:2px 10px; border-radius:8px; font-size:11px; font-weight:700; color:#fff; margin-bottom:10px; }
.rb-card-from .rb-card-tag { background:#64748b; }
.rb-card-tag.to { background:#f59e0b; }
.rb-card-meta { display:flex; flex-direction:column; gap:6px; }
.rb-meta-row { display:flex; justify-content:space-between; align-items:center; font-size:13px; }
.rb-meta-lbl { color:#94a3b8; }
.rb-meta-val { color:#334155; font-weight:500; text-align:right; max-width:60%; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.rb-arrow { display:flex; flex-direction:column; align-items:center; justify-content:center; gap:4px; flex-shrink:0; padding:0 4px; }
.rb-arrow-line { width:28px; height:2px; background:#e2e8f0; }
.rb-arrow-head { font-size:18px; color:#94a3b8; }

/* batch list */
.rb-batch { padding:20px 24px; border-bottom:1px solid #f1f5f9; max-height:260px; overflow-y:auto; }
.rb-batch-hint { font-size:13px; color:#64748b; margin-bottom:12px; }
.rb-batch-item { display:flex; justify-content:space-between; align-items:center; padding:10px 14px; background:#f8fafc; border-radius:10px; margin-bottom:6px; font-size:13px; }
.rb-batch-name { font-weight:600; color:#0f172a; }
.rb-batch-info { color:#64748b; font-size:12px; }

/* table */
.rb-table-area { padding:0 24px; }
.rb-table-hint { font-size:13px; color:#64748b; padding:16px 0 8px; }
.rb-table-scroll { max-height:280px; overflow-y:auto; }
.rb-table { width:100%; border-collapse:collapse; font-size:13px; }
.rb-table th { text-align:left; padding:10px 12px; color:#94a3b8; font-weight:500; font-size:11px; text-transform:uppercase; letter-spacing:.5px; border-bottom:1px solid #e2e8f0; position:sticky; top:0; background:#fff; }
.rb-table td { padding:12px; border-bottom:1px solid #f1f5f9; cursor:pointer; transition:background .1s; }
.rb-table tbody tr:hover td { background:#f8fafc; }
.rb-row-sel td { background:#fffbeb !important; }
.rb-row-cur td { font-weight:500; }
.rb-radio { width:18px; height:18px; border-radius:50%; border:2px solid #cbd5e1; display:flex; align-items:center; justify-content:center; transition:all .15s; }
.rb-radio.on { border-color:#f59e0b; background:#fef3c7; }
.rb-radio.on span { width:6px; height:6px; border-radius:50%; background:#f59e0b; }
.rb-badge { padding:2px 10px; border-radius:10px; font-size:11px; font-weight:600; background:#f1f5f9; color:#64748b; }
.rb-badge-latest { background:#dbeafe; color:#2563eb; }
.rb-td-mono { font-family:monospace; font-size:12px; max-width:180px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.rb-td-img { max-width:240px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.rb-td-time { color:#94a3b8; font-size:12px; white-space:nowrap; }

.rb-loading { text-align:center; padding:40px; color:#94a3b8; }

/* footer */
.rb-ftr { display:flex; align-items:center; justify-content:space-between; padding:16px 24px; border-top:1px solid #f1f5f9; flex-shrink:0; }
.rb-ftr-note { font-size:12px; color:#94a3b8; }
.rb-ftr-btns { display:flex; gap:10px; }
.rb-btn-cancel { padding:9px 22px; border-radius:8px; border:1px solid #e2e8f0; background:#fff; color:#64748b; font-size:14px; cursor:pointer; }
.rb-btn-cancel:hover { background:#f8fafc; }
.rb-btn-go { padding:9px 22px; border-radius:8px; border:none; background:#f59e0b; color:#fff; font-size:14px; font-weight:600; cursor:pointer; transition:all .15s; }
.rb-btn-go:hover:not(:disabled) { background:#d97706; box-shadow:0 4px 12px rgba(245,158,11,.3); }
.rb-btn-go:disabled { opacity:.4; cursor:not-allowed; }
.rb-btn-spin { display:inline-block; width:14px; height:14px; border:2px solid #fff; border-top-color:transparent; border-radius:50%; animation:rbSpin .6s linear infinite; margin-right:4px; vertical-align:middle; }
@keyframes rbSpin { to { transform:rotate(360deg); } }
</style>

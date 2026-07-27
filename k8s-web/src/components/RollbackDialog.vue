<template>
  <div v-if="visible" class="rbo" @click.self="$emit('close')">
    <div class="rbd">
      <div class="rbd-bar">
        <div class="rbd-bar-l">
          <span class="rbd-bar-i"><AppIcon name="sync" size="18"/></span>
          <span class="rbd-bar-t">{{ batchMode ? `批量${batchAction}` : '版本回滚' }}</span>
          <span v-if="resourceName && !batchMode" class="rbd-bar-n">{{ resourceName }}</span>
        </div>
        <button class="rbd-x" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="rbd-body">
        <!-- Version Comparison -->
        <div class="rbd-compare" v-if="!batchMode && selectedRevision">
          <div class="rbd-col rbd-col-cur">
            <span class="rbd-badge rbd-badge-cur">当前运行</span>
            <div class="rbd-kv"><span>镜像</span><span class="rbd-kv-v" :title="currentInfo?.image">{{ currentInfo?.image || '—' }}</span></div>
            <div class="rbd-kv"><span>副本</span><span class="rbd-kv-v">{{ currentInfo?.replicas && currentInfo.replicas !== '0' ? currentInfo.replicas : '—' }}</span></div>
            <div class="rbd-kv"><span>部署</span><span class="rbd-kv-v" :title="currentInfo?.deployedAt">{{ currentInfo?.deployedAt || '—' }}</span></div>
          </div>
          <div class="rbd-arr">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#94a3b8" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
          </div>
          <div class="rbd-col rbd-col-old">
            <span class="rbd-badge rbd-badge-old">回滚目标</span>
            <div class="rbd-kv"><span>版本</span><span class="rbd-kv-v" :title="selectedRevision?.name">{{ selectedRevision?.name?.length > 28 ? '...' + selectedRevision?.name?.slice(-28) : (selectedRevision?.name || '—') }}</span></div>
            <div class="rbd-kv"><span>镜像</span><span class="rbd-kv-v" :title="selectedRevision?.image">{{ selectedRevision?.image || '—' }}</span></div>
            <div class="rbd-kv"><span>创建</span><span class="rbd-kv-v" :title="selectedRevision?.createdAt">{{ selectedRevision?.createdAt || '—' }}</span></div>
          </div>
        </div>

        <!-- Batch selection -->
        <div class="rbd-batch" v-if="batchMode && selectedItems.length">
          <div class="rbd-batch-t">{{ selectedItems.length }} 个应用将{{ batchAction }}到上一版本</div>
          <div v-for="it in selectedItems" :key="it.name" class="rbd-batch-i">
            <span>{{ it.name }}</span>
            <span class="rbd-batch-a" v-if="it.currentImage && it.targetImage">{{ it.currentImage }} → {{ it.targetImage }}</span>
          </div>
        </div>

        <!-- Revision Table -->
        <div class="rbd-tbl-wrap" v-if="!batchMode">
          <div class="rbd-tbl-h" v-if="revisions.length">选择历史版本 — 点击行选中</div>
          <div class="rbd-tbl-scroll">
            <table class="rbd-tbl" v-if="revisions.length">
              <thead><tr><th></th><th>版本</th><th>Revision</th><th>镜像</th><th>副本</th><th>时间</th></tr></thead>
              <tbody>
                <tr v-for="(rev,idx) in revisions" :key="rev.name"
                    :class="{sel:selectedRevision?.name===rev.name, cur:idx===0}"
                    @click="selectRevision(rev)">
                  <td><div class="rbd-dot" :class="{on:selectedRevision?.name===rev.name}"><span v-if="selectedRevision?.name===rev.name"></span></div></td>
                  <td><span class="rbd-tag" :class="idx===0?'rbd-tag-l':''">{{idx===0?'Latest':'v'+(revisions.length-idx)}}</span></td>
                  <td class="mono" :title="rev.name">{{ rev.name?.length>30?'...'+rev.name.slice(-30):(rev.name||'—') }}</td>
                  <td class="img" :title="rev.image">{{ rev.image||'—' }}</td>
                  <td>{{ rev.replicas ? rev.replicas : '—' }}</td>
                  <td class="time" :title="rev.createdAt">{{ rev.createdAt||'—' }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="rbd-empty">暂无历史版本记录</div>
          </div>
        </div>

        <div class="rbd-load" v-if="loading">加载版本历史…</div>
      </div>

      <div class="rbd-ft">
        <div class="rbd-ft-note" v-if="!batchMode">回滚触发滚动更新，不影响服务可用性</div>
        <div class="rbd-ft-btns">
          <button class="rbd-bt-c" @click="$emit('close')">取消</button>
          <button class="rbd-bt-g" :disabled="!canConfirm||confirming" @click="onConfirm">
            <span v-if="confirming" class="sp"></span>
            {{ confirming?`${batchAction}中…`:batchMode?`确认${batchAction} ${selectedItems.length} 个`:'回滚到此版本' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({
  visible:Boolean,loading:Boolean,confirming:Boolean,resourceName:String,currentInfo:Object,
  revisions:{type:Array,default:()=>[]},selectedRevision:Object,batchMode:Boolean,
  batchAction:{type:String,default:'回滚'},selectedItems:{type:Array,default:()=>[]}
})
const emit = defineEmits(['close','select','confirm'])
const canConfirm = computed(()=>props.batchMode?props.selectedItems.length>0:!!props.selectedRevision)
function selectRevision(r){emit('select',r)}
function onConfirm(){emit('confirm')}
</script>

<style scoped>
.rbo{position:fixed;inset:0;z-index:1000;background:rgba(15,23,42,.5);backdrop-filter:blur(6px);display:flex;align-items:center;justify-content:center;padding:20px}
.rbd{width:100%;max-width:900px;max-height:88vh;background:#fff;border-radius:18px;box-shadow:0 30px 100px rgba(0,0,0,.25);display:flex;flex-direction:column;overflow:hidden}
.rbd-bar{display:flex;align-items:center;justify-content:space-between;padding:18px 24px;border-bottom:1px solid #f1f5f9;flex-shrink:0}
.rbd-bar-l{display:flex;align-items:center;gap:8px;font-size:16px;font-weight:700;color:#0f172a}
.rbd-bar-i{color:#f59e0b}.rbd-bar-n{font-weight:400;color:#94a3b8;font-size:13px}.rbd-bar-n::before{content:'·';margin:0 6px}
.rbd-x{width:32px;height:32px;border-radius:8px;border:none;background:#f1f5f9;color:#94a3b8;cursor:pointer;display:flex;align-items:center;justify-content:center}
.rbd-x:hover{background:#e2e8f0;color:#0f172a}

.rbd-body{padding:20px 24px;overflow-y:auto;flex:1}
.rbd-compare{display:flex;align-items:stretch;gap:16px;margin-bottom:20px}
.rbd-col{flex:1;border-radius:14px;padding:16px 18px;min-width:0}
.rbd-col-cur{border:1.5px solid #e2e8f0;background:#f8fafc}
.rbd-col-old{border:1.5px solid #fcd34d;background:#fffbeb}
.rbd-badge{display:inline-block;padding:2px 10px;border-radius:8px;font-size:11px;font-weight:700;color:#fff;margin-bottom:10px}
.rbd-badge-cur{background:#64748b}.rbd-badge-old{background:#d97706}
.rbd-kv{display:flex;justify-content:space-between;align-items:center;padding:3px 0;font-size:12px;gap:8px}
.rbd-kv>span:first-child{color:#94a3b8;flex-shrink:0}.rbd-kv-v{color:#334155;font-weight:500;text-align:right;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.rbd-arr{display:flex;align-items:center;justify-content:center;flex-shrink:0;width:36px}

.rbd-batch{padding:0 0 16px}
.rbd-batch-t{font-size:13px;color:#64748b;margin-bottom:10px}
.rbd-batch-i{display:flex;justify-content:space-between;align-items:center;padding:8px 14px;background:#f8fafc;border-radius:10px;margin-bottom:6px;font-size:13px;gap:8px}
.rbd-batch-a{color:#94a3b8;font-size:12px;flex-shrink:0}

.rbd-tbl-h{font-size:13px;color:#64748b;padding-bottom:8px}
.rbd-tbl-scroll{max-height:280px;overflow-y:auto}
.rbd-tbl{width:100%;border-collapse:collapse;font-size:13px;table-layout:fixed}
.rbd-tbl th{text-align:left;padding:9px 10px;color:#94a3b8;font-weight:500;font-size:11px;text-transform:uppercase;letter-spacing:.5px;border-bottom:1px solid #e2e8f0;position:sticky;top:0;background:#fff}
.rbd-tbl th:nth-child(1){width:36px}.rbd-tbl th:nth-child(2){width:70px}.rbd-tbl th:nth-child(3){width:160px}.rbd-tbl th:nth-child(4){width:auto}.rbd-tbl th:nth-child(5){width:60px}.rbd-tbl th:nth-child(6){width:150px}
.rbd-tbl td{padding:10px;border-bottom:1px solid #f1f5f9;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.rbd-tbl td.mono{font-family:monospace;font-size:12px}.rbd-tbl td.img{max-width:0}.rbd-tbl td.time{color:#94a3b8;font-size:12px}
.rbd-tbl tbody tr:hover td{background:#f8fafc}.rbd-tbl .sel td{background:#fffbeb!important}.rbd-tbl .cur td{font-weight:500}
.rbd-dot{width:18px;height:18px;border-radius:50%;border:2px solid #cbd5e1;display:flex;align-items:center;justify-content:center;transition:all .15s;flex-shrink:0}
.rbd-dot.on{border-color:#d97706;background:#fef3c7}.rbd-dot.on span{width:6px;height:6px;border-radius:50%;background:#d97706}
.rbd-tag{padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;background:#f1f5f9;color:#64748b}
.rbd-tag-l{background:#dbeafe;color:#2563eb}
.rbd-empty{text-align:center;padding:40px;color:#94a3b8;font-size:13px}
.rbd-load{text-align:center;padding:40px;color:#94a3b8}

.rbd-ft{display:flex;align-items:center;justify-content:space-between;padding:14px 24px;border-top:1px solid #f1f5f9;flex-shrink:0}
.rbd-ft-note{font-size:12px;color:#94a3b8}
.rbd-ft-btns{display:flex;gap:10px}
.rbd-bt-c{padding:9px 22px;border-radius:8px;border:1px solid #e2e8f0;background:#fff;color:#64748b;font-size:14px;cursor:pointer}.rbd-bt-c:hover{background:#f8fafc}
.rbd-bt-g{padding:9px 22px;border-radius:8px;border:none;background:#d97706;color:#fff;font-size:14px;font-weight:600;cursor:pointer;display:flex;align-items:center;gap:6px}
.rbd-bt-g:hover:not(:disabled){background:#b45309;box-shadow:0 4px 12px rgba(217,119,6,.3)}
.rbd-bt-g:disabled{opacity:.4;cursor:not-allowed}
.sp{width:14px;height:14px;border:2px solid #fff;border-top-color:transparent;border-radius:50%;animation:rbSpin .6s linear infinite}
@keyframes rbSpin{to{transform:rotate(360deg)}}
</style>

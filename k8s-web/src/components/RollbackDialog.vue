<template>
  <div v-if="visible" class="rollback-overlay" @click.self="$emit('close')">
    <div class="rollback-dialog">
      <!-- Header -->
      <div class="rollback-header">
        <div class="rollback-header-left">
          <AppIcon name="sync" size="20"/>
          <span>{{ batchMode ? `批量${batchAction}` : '版本回滚' }}</span>
        </div>
        <button class="rollback-close" @click="$emit('close')">&times;</button>
      </div>

      <!-- Sub Header -->
      <div class="rollback-subtitle">
        {{ batchMode ? `已选择 ${selectedItems.length} 个资源` : resourceName }}
      </div>

      <!-- Version Comparison Cards -->
      <div class="rollback-cards" v-if="!batchMode && selectedRevision">
        <div class="version-card current">
          <div class="card-badge current-badge">当前版本</div>
          <div class="card-body">
            <div class="card-row">
              <span class="card-label">镜像</span>
              <span class="card-value">{{ currentInfo?.image || '—' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">副本数</span>
              <span class="card-value">{{ currentInfo?.replicas || '—' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">部署时间</span>
              <span class="card-value">{{ currentInfo?.deployedAt || '—' }}</span>
            </div>
          </div>
        </div>
        <div class="version-arrow">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="#94a3b8" stroke-width="2">
            <line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/>
          </svg>
        </div>
        <div class="version-card target">
          <div class="card-badge target-badge">目标版本</div>
          <div class="card-body">
            <div class="card-row">
              <span class="card-label">ReplicaSet</span>
              <span class="card-value">{{ selectedRevision?.name || '—' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">镜像</span>
              <span class="card-value">{{ selectedRevision?.image || '—' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">创建时间</span>
              <span class="card-value">{{ selectedRevision?.createdAt || '—' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Batch Mode: Selected Items List -->
      <div class="batch-list" v-if="batchMode && selectedItems.length">
        <div v-for="item in selectedItems" :key="item.name" class="batch-item">
          <span class="batch-name">{{ item.name }}</span>
          <span class="batch-to">{{ item.currentImage }} → {{ item.targetImage || '上一版本' }}</span>
        </div>
      </div>

      <!-- Revision Selection Table -->
      <div class="rollback-table-wrap" v-if="!batchMode">
        <table class="rollback-table">
          <thead>
            <tr>
              <th></th>
              <th>版本</th>
              <th>镜像</th>
              <th>副本</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(rev, idx) in revisions" :key="rev.name"
                :class="{ 'row-selected': selectedRevision?.name === rev.name, 'row-current': idx === 0 }"
                @click="selectRevision(rev)">
              <td>
                <div class="radio-dot" :class="{ active: selectedRevision?.name === rev.name }">
                  <span v-if="selectedRevision?.name === rev.name" class="dot-inner"></span>
                </div>
              </td>
              <td>
                <span class="rev-tag" :class="idx === 0 ? 'rev-latest' : ''">{{ idx === 0 ? '最新' : '#' + (revisions.length - idx) }}</span>
              </td>
              <td class="rev-image">{{ rev.image || '—' }}</td>
              <td>{{ rev.replicas || '—' }}</td>
              <td class="rev-time">{{ rev.createdAt || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Loading -->
      <div class="rollback-loading" v-if="loading">加载版本历史...</div>

      <!-- Footer -->
      <div class="rollback-footer">
        <button class="btn-cancel" @click="$emit('close')">取消</button>
        <button class="btn-confirm" :disabled="!canConfirm || confirming" @click="onConfirm">
          <AppIcon name="sync" size="14" v-if="!confirming"/>
          {{ confirming ? `${batchAction}中...` : batchMode ? `${batchAction} ${selectedItems.length} 个资源` : `确认${batchAction}` }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  visible: Boolean,
  loading: Boolean,
  confirming: Boolean,
  resourceName: String,
  currentInfo: Object,
  revisions: { type: Array, default: () => [] },
  selectedRevision: Object,
  batchMode: Boolean,
  batchAction: { type: String, default: '回滚' },
  selectedItems: { type: Array, default: () => [] }
})

const emit = defineEmits(['close', 'select', 'confirm'])

const canConfirm = computed(() => {
  if (props.batchMode) return props.selectedItems.length > 0
  return !!props.selectedRevision
})

function selectRevision(rev) { emit('select', rev) }
function onConfirm() { emit('confirm') }
</script>

<style scoped>
.rollback-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(15,23,42,0.6); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; padding: 20px;
}
.rollback-dialog {
  width: 100%; max-width: 720px; max-height: 85vh; overflow-y: auto;
  background: #fff; border-radius: 16px;
  box-shadow: 0 25px 80px rgba(0,0,0,0.25);
}
.rollback-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px 0;
  font-size: 18px; font-weight: 700; color: #1e293b;
}
.rollback-header-left { display: flex; align-items: center; gap: 8px; }
.rollback-close {
  background: none; border: none; font-size: 24px; cursor: pointer; color: #94a3b8;
  padding: 0; line-height: 1;
}
.rollback-subtitle { padding: 4px 24px 16px; font-size: 13px; color: #64748b; }
.rollback-cards { display: flex; align-items: center; gap: 16px; padding: 0 24px 20px; }
.version-card {
  flex: 1; border-radius: 12px; padding: 16px; position: relative;
  border: 2px solid #e2e8f0;
}
.version-card.current { border-color: #3b82f6; background: #f8fafc; }
.version-card.target { border-color: #f59e0b; background: #fffbeb; }
.card-badge {
  position: absolute; top: -10px; left: 16px;
  padding: 2px 10px; border-radius: 10px; font-size: 11px; font-weight: 600; color: #fff;
}
.current-badge { background: #3b82f6; }
.target-badge { background: #f59e0b; }
.card-body { margin-top: 6px; }
.card-row { display: flex; justify-content: space-between; padding: 4px 0; font-size: 12px; }
.card-label { color: #94a3b8; }
.card-value { color: #334155; font-weight: 500; max-width: 60%; text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.version-arrow { flex-shrink: 0; }
.rollback-table-wrap { padding: 0 24px; max-height: 240px; overflow-y: auto; }
.rollback-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.rollback-table th { text-align: left; padding: 8px 12px; color: #94a3b8; font-weight: 500; border-bottom: 1px solid #e2e8f0; position: sticky; top: 0; background: #fff; }
.rollback-table td { padding: 10px 12px; border-bottom: 1px solid #f1f5f9; cursor: pointer; }
.rollback-table tr:hover td { background: #f8fafc; }
.row-selected td { background: #eff6ff; }
.row-current td { font-weight: 500; }
.radio-dot {
  width: 18px; height: 18px; border-radius: 50%; border: 2px solid #cbd5e1;
  display: flex; align-items: center; justify-content: center; transition: all 0.2s;
}
.radio-dot.active { border-color: #3b82f6; }
.dot-inner { width: 8px; height: 8px; border-radius: 50%; background: #3b82f6; }
.rev-tag { padding: 1px 8px; border-radius: 10px; font-size: 11px; background: #f1f5f9; color: #64748b; }
.rev-tag.rev-latest { background: #dbeafe; color: #2563eb; }
.rev-image { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rev-time { color: #94a3b8; font-size: 12px; }
.rollback-loading { text-align: center; padding: 40px; color: #94a3b8; }
.rollback-footer {
  display: flex; justify-content: flex-end; gap: 12px; padding: 20px 24px;
  border-top: 1px solid #f1f5f9;
}
.btn-cancel {
  padding: 10px 24px; border-radius: 8px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 14px; cursor: pointer;
}
.btn-cancel:hover { background: #f8fafc; }
.btn-confirm {
  padding: 10px 24px; border-radius: 8px; border: none;
  background: linear-gradient(135deg, #f59e0b, #d97706); color: #fff;
  font-size: 14px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; gap: 6px;
  transition: all 0.2s;
}
.btn-confirm:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(245,158,11,0.3); }
.btn-confirm:disabled { opacity: 0.5; cursor: not-allowed; }

/* Batch list */
.batch-list { padding: 0 24px 16px; max-height: 200px; overflow-y: auto; }
.batch-item {
  display: flex; justify-content: space-between; padding: 8px 12px;
  background: #f8fafc; border-radius: 8px; margin-bottom: 6px; font-size: 13px;
}
.batch-name { font-weight: 600; color: #1e293b; }
.batch-to { color: #64748b; }
</style>

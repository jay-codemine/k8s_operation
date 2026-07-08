<template>
  <div class="approval-policy-page">
    <!-- 页面头部 -->
    <div class="page-banner">
      <div class="banner-content">
        <div class="banner-left">
          <div class="banner-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
          </div>
          <div>
            <h1 class="banner-title">审批策略设置</h1>
            <p class="banner-desc">为每个部署环境配置多级审批流程，保障发布安全合规</p>
          </div>
        </div>
        <button class="btn-save" @click="saveAll" :disabled="!hasChanges || saving">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
            <polyline points="17 21 17 13 7 13 7 21"/>
            <polyline points="7 3 7 8 15 8"/>
          </svg>
          <span>{{ saving ? '保存中...' : '保存全部' }}</span>
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <div class="loader"><div class="loader-ring"></div></div>
      <span>加载环境配置中...</span>
    </div>

    <!-- 环境审批卡片列表 -->
    <div v-else class="env-cards">
      <div v-for="env in environments" :key="env.id" class="env-card" :class="{ enabled: env.require_approval }">
        <!-- 环境头部 -->
        <div class="env-header">
          <div class="env-info">
            <span class="env-color-dot" :style="{ background: env.color }"></span>
            <h3 class="env-name">{{ env.display_name || env.name }}</h3>
            <span class="env-badge" :class="env.name">{{ env.name }}</span>
          </div>
          <label class="toggle-switch">
            <input type="checkbox" v-model="env.require_approval" @change="markChanged(env)" />
            <span class="toggle-slider"></span>
            <span class="toggle-label">{{ env.require_approval ? '需要审批' : '无需审批' }}</span>
          </label>
        </div>

        <!-- 审批级别配置（展开） -->
        <div v-if="env.require_approval" class="approval-levels-section">
          <div class="levels-header">
            <span class="levels-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-sm">
                <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                <circle cx="8.5" cy="7" r="4"/>
                <line x1="20" y1="8" x2="20" y2="14"/>
                <line x1="23" y1="11" x2="17" y2="11"/>
              </svg>
              审批链 ({{ env.approval_levels?.length || 0 }} 级)
            </span>
            <button class="btn-add-level" @click="addLevel(env)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
              </svg>
              添加审批级别
            </button>
          </div>

          <!-- 无审批级别时 -->
          <div v-if="!env.approval_levels || env.approval_levels.length === 0" class="empty-levels">
            <p>尚未配置审批级别，请点击「添加审批级别」按钮</p>
            <p class="empty-hint">未配置审批级别时，默认为单人审批</p>
          </div>

          <!-- 审批级别列表 -->
          <div v-else class="levels-list">
            <div v-for="(level, idx) in env.approval_levels" :key="idx" class="level-item">
              <div class="level-badge">{{ idx + 1 }}</div>
              <div class="level-fields">
                <div class="field-row">
                  <div class="field-group">
                    <label>级别名称</label>
                    <input type="text" v-model="level.label" placeholder="如：测试负责人" 
                           @input="markChanged(env)" class="field-input" />
                  </div>
                  <div class="field-group">
                    <label>角色标识</label>
                    <input type="text" v-model="level.role" placeholder="如：test_lead" 
                           @input="markChanged(env)" class="field-input" />
                  </div>
                </div>
                <div class="field-row">
                  <div class="field-group full">
                    <label>审批人 (用户ID，逗号分隔)</label>
                    <input type="text" :value="(level.user_ids || []).join(',')"
                           @input="updateUserIds(env, idx, $event.target.value)"
                           placeholder="如: 1,2,3" class="field-input" />
                  </div>
                </div>
              </div>
              <button class="btn-remove-level" @click="removeLevel(env, idx)" title="删除此级别">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 审批流程预览 -->
          <div v-if="env.approval_levels && env.approval_levels.length > 0" class="flow-preview">
            <span class="flow-label">审批流程：</span>
            <div class="flow-steps">
              <span class="flow-step" v-for="(level, idx) in env.approval_levels" :key="idx">
                <span class="step-num">{{ idx + 1 }}</span>
                {{ level.label || '审批人' }}
                <svg v-if="idx < env.approval_levels.length - 1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="flow-arrow">
                  <polyline points="9 18 15 12 9 6"/>
                </svg>
              </span>
              <span class="flow-step final">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                部署
              </span>
            </div>
          </div>
        </div>

        <!-- 关闭审批时的提示 -->
        <div v-else class="disabled-hint">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>
          </svg>
          <span>该环境部署无需审批，构建成功后将直接部署</span>
        </div>
      </div>
    </div>

    <!-- 保存成功提示 -->
    <div v-if="showSuccess" class="success-toast">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
      </svg>
      审批策略保存成功
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getEnvironmentList, updateApprovalPolicy } from '@/api/cicd/environment.js'

const loading = ref(true)
const saving = ref(false)
const hasChanges = ref(false)
const showSuccess = ref(false)
const environments = ref([])
const changedEnvIds = ref(new Set())

onMounted(async () => {
  await loadEnvironments()
})

async function loadEnvironments() {
  loading.value = true
  try {
    const res = await getEnvironmentList({ page: 1, page_size: 50 })
    if (res.code === 0 && res.data?.list) {
      environments.value = res.data.list.map(env => ({
        ...env,
        approval_levels: env.approval_levels || []
      }))
    }
  } catch (e) {
    console.error('加载环境列表失败:', e)
  } finally {
    loading.value = false
  }
}

function markChanged(env) {
  changedEnvIds.value.add(env.id)
  hasChanges.value = true
}

function addLevel(env) {
  if (!env.approval_levels) env.approval_levels = []
  env.approval_levels.push({
    level: env.approval_levels.length + 1,
    role: '',
    label: '',
    user_ids: []
  })
  markChanged(env)
}

function removeLevel(env, idx) {
  env.approval_levels.splice(idx, 1)
  // 重新编号
  env.approval_levels.forEach((l, i) => { l.level = i + 1 })
  markChanged(env)
}

function updateUserIds(env, idx, value) {
  const ids = value.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n) && n > 0)
  env.approval_levels[idx].user_ids = ids
  markChanged(env)
}

async function saveAll() {
  saving.value = true
  try {
    for (const env of environments.value) {
      if (changedEnvIds.value.has(env.id)) {
        await updateApprovalPolicy(env.id, env.require_approval, env.approval_levels || [])
      }
    }
    changedEnvIds.value.clear()
    hasChanges.value = false
    showSuccess.value = true
    setTimeout(() => { showSuccess.value = false }, 3000)
  } catch (e) {
    console.error('保存审批策略失败:', e)
    alert('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.approval-policy-page {
  padding: 0;
  min-height: 100vh;
  background: #f8fafc;
}

.page-banner {
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
  padding: 28px 32px;
  margin-bottom: 24px;
}
.banner-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
}
.banner-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.banner-icon {
  width: 44px; height: 44px;
  background: rgba(99, 102, 241, 0.2);
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
}
.banner-icon svg { width: 24px; height: 24px; color: #a5b4fc; }
.banner-title { color: #f8fafc; font-size: 20px; font-weight: 700; margin: 0; }
.banner-desc { color: #94a3b8; font-size: 13px; margin: 4px 0 0; }

.btn-save {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 20px;
  background: #6366f1; color: white;
  border: none; border-radius: 8px;
  font-size: 14px; font-weight: 600;
  cursor: pointer; transition: all 0.2s;
}
.btn-save:hover:not(:disabled) { background: #4f46e5; transform: translateY(-1px); }
.btn-save:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-save svg { width: 16px; height: 16px; }

.loading-state {
  display: flex; flex-direction: column; align-items: center;
  padding: 80px 0; gap: 16px; color: #64748b;
}
.loader-ring {
  width: 32px; height: 32px;
  border: 3px solid #e2e8f0; border-top-color: #6366f1;
  border-radius: 50%; animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.env-cards {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 32px 32px;
  display: flex; flex-direction: column; gap: 20px;
}

.env-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s;
}
.env-card.enabled { border-color: #6366f1; box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.1); }

.env-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
}
.env-info { display: flex; align-items: center; gap: 12px; }
.env-color-dot { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }
.env-name { font-size: 16px; font-weight: 600; color: #1e293b; margin: 0; }
.env-badge {
  padding: 2px 10px; border-radius: 4px;
  font-size: 11px; font-weight: 700; text-transform: uppercase;
  background: #f1f5f9; color: #64748b;
}
.env-badge.dev { background: #dcfce7; color: #166534; }
.env-badge.test, .env-badge.staging { background: #fef3c7; color: #92400e; }
.env-badge.prod { background: #fee2e2; color: #991b1b; }

/* Toggle Switch */
.toggle-switch {
  display: flex; align-items: center; gap: 10px; cursor: pointer;
}
.toggle-switch input { display: none; }
.toggle-slider {
  width: 44px; height: 24px;
  background: #cbd5e1; border-radius: 12px;
  position: relative; transition: background 0.3s;
}
.toggle-slider::after {
  content: '';
  position: absolute; top: 3px; left: 3px;
  width: 18px; height: 18px;
  background: white; border-radius: 50%;
  transition: transform 0.3s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-switch input:checked + .toggle-slider { background: #6366f1; }
.toggle-switch input:checked + .toggle-slider::after { transform: translateX(20px); }
.toggle-label { font-size: 13px; color: #64748b; font-weight: 500; }

/* Approval Levels Section */
.approval-levels-section { padding: 20px 24px; }
.levels-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 16px;
}
.levels-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 600; color: #475569;
}
.icon-sm { width: 16px; height: 16px; }

.btn-add-level {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px;
  background: #f0fdf4; color: #16a34a;
  border: 1px solid #bbf7d0; border-radius: 6px;
  font-size: 12px; font-weight: 600;
  cursor: pointer; transition: all 0.2s;
}
.btn-add-level:hover { background: #dcfce7; }
.btn-add-level svg { width: 14px; height: 14px; }

.empty-levels {
  text-align: center; padding: 24px;
  background: #f8fafc; border-radius: 8px;
  color: #64748b; font-size: 13px;
}
.empty-hint { margin-top: 4px; font-size: 12px; color: #94a3b8; }

.levels-list { display: flex; flex-direction: column; gap: 12px; }
.level-item {
  display: flex; align-items: flex-start; gap: 12px;
  padding: 16px;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
}
.level-badge {
  width: 28px; height: 28px; flex-shrink: 0;
  background: #6366f1; color: white;
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 700;
}
.level-fields { flex: 1; display: flex; flex-direction: column; gap: 8px; }
.field-row { display: flex; gap: 12px; }
.field-group { flex: 1; }
.field-group.full { flex: 2; }
.field-group label { display: block; font-size: 11px; color: #64748b; margin-bottom: 4px; font-weight: 500; }
.field-input {
  width: 100%; padding: 8px 12px;
  border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 13px; color: #1e293b;
  transition: border-color 0.2s;
}
.field-input:focus { outline: none; border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }

.btn-remove-level {
  width: 32px; height: 32px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  background: none; border: 1px solid transparent;
  border-radius: 6px; cursor: pointer; color: #94a3b8;
  transition: all 0.2s;
}
.btn-remove-level:hover { background: #fee2e2; color: #ef4444; border-color: #fecaca; }
.btn-remove-level svg { width: 16px; height: 16px; }

/* Flow Preview */
.flow-preview {
  margin-top: 16px; padding: 12px 16px;
  background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px;
  display: flex; align-items: center; gap: 12px;
}
.flow-label { font-size: 12px; color: #1d4ed8; font-weight: 600; white-space: nowrap; }
.flow-steps { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; }
.flow-step {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; background: white;
  border: 1px solid #bfdbfe; border-radius: 4px;
  font-size: 12px; color: #1e40af; font-weight: 500;
}
.step-num {
  width: 16px; height: 16px;
  background: #3b82f6; color: white; border-radius: 50%;
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 10px; font-weight: 700;
}
.flow-arrow { width: 14px; height: 14px; color: #93c5fd; margin: 0 2px; }
.flow-step.final { background: #dcfce7; border-color: #86efac; color: #166534; }
.flow-step.final svg { width: 14px; height: 14px; }

/* Disabled Hint */
.disabled-hint {
  display: flex; align-items: center; gap: 8px;
  padding: 16px 24px;
  color: #94a3b8; font-size: 13px;
}
.disabled-hint svg { width: 16px; height: 16px; flex-shrink: 0; }

/* Success Toast */
.success-toast {
  position: fixed; bottom: 32px; right: 32px;
  display: flex; align-items: center; gap: 8px;
  padding: 14px 20px;
  background: #059669; color: white;
  border-radius: 8px; font-size: 14px; font-weight: 500;
  box-shadow: 0 10px 25px rgba(5, 150, 105, 0.3);
  animation: slideIn 0.3s ease;
  z-index: 9999;
}
.success-toast svg { width: 18px; height: 18px; }
@keyframes slideIn { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
</style>

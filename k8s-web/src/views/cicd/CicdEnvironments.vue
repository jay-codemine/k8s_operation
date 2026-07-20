<template>
  <div class="env-manage">
    <!-- 页头 -->
    <div class="page-head">
      <div class="ph-title">
        <h2>环境管理</h2>
        <span class="ph-sub">统一维护 dev / test / staging / prod 等全局环境；流水线晋级时自动继承环境默认集群与命名空间，无需逐条重复配置</span>
      </div>
      <div class="ph-actions">
        <button class="btn ghost" @click="loadList" :disabled="loading">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          {{ loading ? '加载中...' : '刷新' }}
        </button>
        <button class="btn primary" @click="openCreate">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新建环境
        </button>
      </div>
    </div>

    <!-- 列表 -->
    <div class="table-card">
      <div v-if="loading && list.length === 0" class="tc-loading">
        <div class="spinner"></div><p>正在加载环境...</p>
      </div>
      <div v-else-if="list.length === 0" class="tc-empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        <p>暂无环境</p>
        <span>点击「新建环境」创建 dev / test / staging / prod 等部署环境</span>
      </div>
      <table v-else class="env-table">
        <thead>
          <tr>
            <th style="width:60px">排序</th>
            <th>环境</th>
            <th>标识</th>
            <th>默认集群</th>
            <th>命名空间</th>
            <th style="width:90px">审批</th>
            <th>更新时间</th>
            <th style="width:140px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="env in list" :key="env.id">
            <td><span class="order-badge">{{ env.sort_order }}</span></td>
            <td>
              <span class="env-badge" :style="{ background: env.color || '#6366f1' }">{{ env.display_name || env.name }}</span>
            </td>
            <td><code class="mono">{{ env.name }}</code></td>
            <td>
              <span v-if="env.cluster_id > 0">{{ env.cluster_name || ('#' + env.cluster_id) }}</span>
              <span v-else class="muted">未绑定</span>
            </td>
            <td><code class="mono">{{ env.namespace || 'default' }}</code></td>
            <td>
              <span v-if="env.require_approval" class="tag approval">🔒 需审批</span>
              <span v-else class="tag free">自动</span>
              <span v-if="env.auto_rollback_on_fail" class="tag rollback" title="部署失败时自动回滚到上一版本">↩ 失败自动回滚</span>
            </td>
            <td class="muted">{{ formatTs(env.modified_at) }}</td>
            <td>
              <button class="act edit" @click="openEdit(env)">编辑</button>
              <button class="act del" @click="confirmDelete(env)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 创建/编辑弹窗 -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
        <div class="modal">
          <div class="m-head">
            <h3>{{ editing ? '编辑环境' : '新建环境' }}</h3>
            <button class="m-close" @click="showModal = false">✕</button>
          </div>
          <div class="m-body">
            <div class="field">
              <label>环境标识 <i>*</i></label>
              <input v-model.trim="form.name" class="input" placeholder="dev / test / staging / prod" :disabled="editing" />
              <p class="hint">英文小写唯一标识，创建后不可修改；晋级/发布单按此匹配环境。</p>
            </div>
            <div class="field">
              <label>显示名称 <i>*</i></label>
              <input v-model.trim="form.display_name" class="input" placeholder="开发 / 测试 / 预发 / 生产" />
            </div>
            <div class="field-row">
              <div class="field">
                <label>默认集群 <i>*</i></label>
                <select v-model.number="form.cluster_id" class="input">
                  <option :value="0" disabled>请选择集群</option>
                  <option v-for="c in clusters" :key="c.id" :value="c.id">{{ c.cluster_name || c.name }}</option>
                </select>
              </div>
              <div class="field">
                <label>默认命名空间</label>
                <input v-model.trim="form.namespace" class="input" placeholder="default" />
              </div>
            </div>
            <div class="field-row">
              <div class="field">
                <label>排序</label>
                <input v-model.number="form.sort_order" class="input" type="number" min="0" placeholder="数字越小越靠前" />
              </div>
              <div class="field">
                <label>环境颜色</label>
                <div class="color-row">
                  <input v-model="form.color" type="color" class="color-picker" />
                  <input v-model.trim="form.color" class="input" placeholder="#6366f1" />
                </div>
              </div>
            </div>
            <div class="field">
              <label>描述</label>
              <input v-model.trim="form.description" class="input" placeholder="环境用途说明（可选）" />
            </div>
            <div class="field">
              <label class="check">
                <input type="checkbox" v-model="form.require_approval" />
                <span>晋级到该环境需要审批（建议生产/预发开启）</span>
              </label>
            </div>
            <div class="field">
              <label class="check">
                <input type="checkbox" v-model="form.auto_rollback_on_fail" />
                <span>部署失败时自动回滚到上一版本（建议生产环境开启）</span>
              </label>
              <p class="hint">开启后，该环境发布/晋级失败将自动下发一个回滚发布单（跳过审批），将工作负载恢复至部署前镜像。</p>
            </div>
          </div>
          <div class="m-foot">
            <button class="btn ghost" @click="showModal = false">取消</button>
            <button class="btn primary" @click="save" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 删除确认 -->
    <Teleport to="body">
      <div v-if="showDelete" class="modal-mask" @click.self="showDelete = false">
        <div class="modal sm">
          <div class="m-head"><h3>删除环境</h3><button class="m-close" @click="showDelete = false">✕</button></div>
          <div class="m-body">
            <p class="del-tip">确认删除环境 <strong>{{ deleteTarget?.display_name || deleteTarget?.name }}</strong>（{{ deleteTarget?.name }}）？</p>
            <p class="hint">删除后该环境将不再出现在晋级链中；已发布的历史记录不受影响。</p>
          </div>
          <div class="m-foot">
            <button class="btn ghost" @click="showDelete = false">取消</button>
            <button class="btn danger" @click="doDelete" :disabled="deleting">{{ deleting ? '删除中...' : '确认删除' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import {
  getK8sEnvironments,
  createK8sEnvironment,
  updateK8sEnvironment,
  deleteK8sEnvironment
} from '@/api/cicd'
import { getClusterList } from '@/api/cluster'

const list = ref([])
const clusters = ref([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)

const showModal = ref(false)
const editing = ref(false)
const editingId = ref(0)
const form = reactive({
  name: '',
  display_name: '',
  description: '',
  cluster_id: 0,
  namespace: 'default',
  color: '#6366f1',
  sort_order: 0,
  require_approval: false,
  auto_rollback_on_fail: false
})

const showDelete = ref(false)
const deleteTarget = ref(null)

// ====== 轻量 Toast ======
const showToast = (msg, type = 'info') => {
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
  setTimeout(() => { el.style.opacity = '0'; setTimeout(() => el.remove(), 300) }, 2500)
}

const formatTs = (ts) => {
  if (!ts) return '-'
  const d = new Date(Number(ts) * 1000)
  if (isNaN(d.getTime())) return '-'
  return d.toLocaleString('zh-CN', { hour12: false })
}

const parseList = (res) => {
  const d = res?.data || res || {}
  return d.list || d.items || d.data?.list || (Array.isArray(d) ? d : [])
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await getK8sEnvironments({ page: 1, page_size: 1000 })
    const arr = parseList(res)
    // 按 sort_order 升序展示
    list.value = (Array.isArray(arr) ? arr : []).sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
  } catch (e) {
    showToast('加载环境失败: ' + (e.message || ''), 'error')
  } finally {
    loading.value = false
  }
}

const loadClusters = async () => {
  try {
    const res = await getClusterList({ page: 1, limit: 1000 })
    const data = res.data || res
    clusters.value = data.list || data.items || data || []
  } catch (e) {
    console.warn('加载集群失败', e)
  }
}

const resetForm = () => {
  form.name = ''
  form.display_name = ''
  form.description = ''
  form.cluster_id = 0
  form.namespace = 'default'
  form.color = '#6366f1'
  form.sort_order = list.value.length > 0 ? Math.max(...list.value.map(e => e.sort_order || 0)) + 1 : 1
  form.require_approval = false
  form.auto_rollback_on_fail = false
}

const openCreate = async () => {
  editing.value = false
  editingId.value = 0
  resetForm()
  if (clusters.value.length === 0) await loadClusters()
  showModal.value = true
}

const openEdit = async (env) => {
  editing.value = true
  editingId.value = env.id
  form.name = env.name
  form.display_name = env.display_name || ''
  form.description = env.description || ''
  form.cluster_id = env.cluster_id || 0
  form.namespace = env.namespace || 'default'
  form.color = env.color || '#6366f1'
  form.sort_order = env.sort_order || 0
  form.require_approval = !!env.require_approval
  form.auto_rollback_on_fail = !!env.auto_rollback_on_fail
  if (clusters.value.length === 0) await loadClusters()
  showModal.value = true
}

const save = async () => {
  if (!form.name) return showToast('请填写环境标识', 'warning')
  if (!form.display_name) return showToast('请填写显示名称', 'warning')
  if (!form.cluster_id || form.cluster_id <= 0) return showToast('请选择默认集群', 'warning')

  saving.value = true
  try {
    const payload = {
      name: form.name,
      display_name: form.display_name,
      description: form.description,
      cluster_id: form.cluster_id,
      namespace: form.namespace || 'default',
      color: form.color,
      sort_order: form.sort_order || 0,
      require_approval: form.require_approval,
      auto_rollback_on_fail: form.auto_rollback_on_fail
    }
    let res
    if (editing.value) {
      res = await updateK8sEnvironment(editingId.value, payload)
    } else {
      res = await createK8sEnvironment(payload)
    }
    if (res && res.code !== undefined && res.code !== 0) {
      throw new Error(res.msg || res.message || '保存失败')
    }
    showToast(editing.value ? '环境已更新' : '环境已创建', 'success')
    showModal.value = false
    await loadList()
  } catch (e) {
    showToast('保存失败: ' + (e.message || ''), 'error')
  } finally {
    saving.value = false
  }
}

const confirmDelete = (env) => {
  deleteTarget.value = env
  showDelete.value = true
}

const doDelete = async () => {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const res = await deleteK8sEnvironment(deleteTarget.value.id)
    if (res && res.code !== undefined && res.code !== 0) {
      throw new Error(res.msg || res.message || '删除失败')
    }
    showToast('环境已删除', 'success')
    showDelete.value = false
    await loadList()
  } catch (e) {
    showToast('删除失败: ' + (e.message || ''), 'error')
  } finally {
    deleting.value = false
  }
}

loadList()
loadClusters()
</script>

<style scoped>
.env-manage { padding: 20px 24px; }

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;
  gap: 16px;
}
.ph-title h2 { margin: 0; font-size: 20px; font-weight: 700; color: #1a202c; }
.ph-sub { display: block; margin-top: 4px; font-size: 13px; color: #718096; max-width: 640px; }
.ph-actions { display: flex; gap: 10px; flex-shrink: 0; }

.btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px; font-size: 13px; font-weight: 500;
  cursor: pointer; border: 1px solid transparent; transition: all 0.15s;
}
.btn svg { width: 15px; height: 15px; }
.btn.primary { background: #6366f1; color: #fff; }
.btn.primary:hover:not(:disabled) { background: #4f46e5; }
.btn.ghost { background: #fff; color: #4a5568; border-color: #e2e8f0; }
.btn.ghost:hover:not(:disabled) { border-color: #6366f1; color: #6366f1; }
.btn.danger { background: #e53e3e; color: #fff; }
.btn.danger:hover:not(:disabled) { background: #c53030; }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }

.table-card {
  background: #fff; border: 1px solid #edf2f7; border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04); overflow: hidden;
}
.tc-loading, .tc-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 64px 24px; text-align: center;
}
.tc-empty svg { width: 52px; height: 52px; color: #cbd5e0; margin-bottom: 14px; }
.tc-empty p { margin: 0 0 6px; font-size: 15px; font-weight: 600; color: #4a5568; }
.tc-empty span { font-size: 13px; color: #a0aec0; }
.spinner {
  width: 32px; height: 32px; border: 3px solid #e2e8f0; border-top-color: #6366f1;
  border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 12px;
}
@keyframes spin { to { transform: rotate(360deg); } }

.env-table { width: 100%; border-collapse: collapse; }
.env-table th {
  text-align: left; padding: 12px 16px; font-size: 12px; font-weight: 600;
  color: #718096; background: #f7fafc; border-bottom: 1px solid #edf2f7;
}
.env-table td {
  padding: 12px 16px; font-size: 13px; color: #2d3748;
  border-bottom: 1px solid #f7fafc; vertical-align: middle;
}
.env-table tbody tr:hover { background: #fafbff; }
.order-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 6px; background: #edf2f7;
  font-size: 12px; font-weight: 600; color: #4a5568;
}
.env-badge {
  display: inline-block; padding: 3px 12px; border-radius: 12px;
  color: #fff; font-size: 12px; font-weight: 600;
}
.mono { font-family: 'SFMono-Regular', Consolas, monospace; font-size: 12px; }
code.mono { background: #f7fafc; padding: 2px 8px; border-radius: 5px; color: #4a5568; }
.muted { color: #a0aec0; }
.tag { display: inline-block; padding: 2px 10px; border-radius: 10px; font-size: 12px; }
.tag.approval { background: #fef3c7; color: #b45309; }
.tag.free { background: #d1fae5; color: #047857; }
.tag.rollback { display: inline-block; margin-top: 4px; background: #f5f3ff; color: #7c3aed; }
.act {
  padding: 4px 12px; border-radius: 6px; font-size: 12px; cursor: pointer;
  border: 1px solid #e2e8f0; background: #fff; margin-right: 6px; transition: all 0.15s;
}
.act.edit:hover { border-color: #6366f1; color: #6366f1; }
.act.del:hover { border-color: #e53e3e; color: #e53e3e; }

.modal-mask {
  position: fixed; inset: 0; background: rgba(0, 0, 0, 0.45);
  display: flex; align-items: center; justify-content: center; z-index: 9999;
}
.modal {
  width: 520px; max-width: calc(100vw - 40px); background: #fff;
  border-radius: 14px; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2); overflow: hidden;
}
.modal.sm { width: 420px; }
.m-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18px 22px; border-bottom: 1px solid #edf2f7;
}
.m-head h3 { margin: 0; font-size: 16px; font-weight: 700; color: #1a202c; }
.m-close {
  border: none; background: none; font-size: 18px; color: #a0aec0;
  cursor: pointer; line-height: 1;
}
.m-close:hover { color: #4a5568; }
.m-body { padding: 22px; max-height: 65vh; overflow-y: auto; }
.field { margin-bottom: 16px; }
.field-row { display: flex; gap: 16px; }
.field-row .field { flex: 1; }
.field label {
  display: block; margin-bottom: 6px; font-size: 13px; font-weight: 600; color: #2d3748;
}
.field label i { color: #e53e3e; font-style: normal; }
.input {
  width: 100%; padding: 9px 12px; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: 13px; outline: none; box-sizing: border-box; background: #fff;
}
.input:focus { border-color: #6366f1; }
.input:disabled { background: #f7fafc; color: #a0aec0; }
.hint { margin: 6px 0 0; font-size: 12px; color: #a0aec0; }
.color-row { display: flex; gap: 8px; align-items: center; }
.color-picker {
  width: 42px; height: 38px; padding: 2px; border: 1px solid #e2e8f0;
  border-radius: 8px; cursor: pointer; flex-shrink: 0;
}
.check { display: flex; align-items: center; gap: 8px; cursor: pointer; font-weight: 500; }
.check input { width: 16px; height: 16px; }
.m-foot {
  display: flex; justify-content: flex-end; gap: 10px;
  padding: 16px 22px; border-top: 1px solid #edf2f7; background: #fafbfc;
}
.del-tip { margin: 0 0 8px; font-size: 14px; color: #2d3748; }
</style>

<template>
  <div class="cleanup-policy-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <span class="title-icon">🧹</span>
          镜像清理策略
        </h1>
        <p class="page-desc">配置镜像自动清理规则，按天数或版本数自动删除过期镜像</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-refresh" @click="loadData" :disabled="loading">🔄 刷新</button>
        <button v-if="canOperate" class="btn btn-primary" @click="openCreateModal">+ 创建策略</button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">📋</div>
        <div class="stat-content">
          <div class="stat-value">{{ policies.length }}</div>
          <div class="stat-label">策略总数</div>
        </div>
      </div>
      <div class="stat-card enabled">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <div class="stat-value">{{ enabledCount }}</div>
          <div class="stat-label">已启用</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🗑️</div>
        <div class="stat-content">
          <div class="stat-value">{{ totalDeleted }}</div>
          <div class="stat-label">累计清理</div>
        </div>
      </div>
    </div>

    <!-- 策略列表 -->
    <div class="table-card">
      <table class="data-table" v-if="!loading">
        <thead>
          <tr>
            <th>策略名称</th>
            <th>关联仓库</th>
            <th>匹配规则</th>
            <th>保留策略</th>
            <th>执行计划</th>
            <th>状态</th>
            <th>上次执行</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="policies.length === 0">
            <td colspan="8" class="empty-cell">
              <div class="empty-state">
                <span class="empty-icon">📭</span>
                <span>暂无清理策略，点击"创建策略"添加</span>
              </div>
            </td>
          </tr>
          <tr v-for="policy in policies" :key="policy.id">
            <td>
              <div class="policy-name">
                <span class="name-text">{{ policy.name }}</span>
                <span class="policy-desc" v-if="policy.description">{{ policy.description }}</span>
              </div>
            </td>
            <td>
              <span class="registry-name">{{ policy.registry_name || '-' }}</span>
            </td>
            <td>
              <div class="pattern-cell">
                <code class="pattern">{{ policy.repository_pattern }}</code>
                <code class="pattern">{{ policy.tag_pattern }}</code>
              </div>
            </td>
            <td>
              <div class="keep-cell">
                <span>保留 {{ policy.keep_last_count }} 个版本</span>
                <span>保留 {{ policy.keep_days }} 天</span>
              </div>
            </td>
            <td>
              <code class="cron">{{ policy.cron_expression }}</code>
            </td>
            <td>
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  :checked="policy.enabled" 
                  @change="togglePolicy(policy)"
                  :disabled="!canOperate"
                />
                <span class="toggle-slider"></span>
              </label>
            </td>
            <td>
              <div class="last-run" v-if="policy.last_run_at">
                <span class="run-time">{{ formatTime(policy.last_run_at) }}</span>
                <span class="run-result">{{ policy.last_run_result }}</span>
              </div>
              <span v-else class="no-run">未执行</span>
            </td>
            <td>
              <div class="action-group">
                <button v-if="canOperate" class="action-btn run" @click="runPolicy(policy)" title="立即执行">▶️</button>
                <button v-if="canOperate" class="action-btn edit" @click="openEditModal(policy)" title="编辑">✏️</button>
                <button class="action-btn logs" @click="viewLogs(policy)" title="日志">📜</button>
                <button v-if="canOperate" class="action-btn delete" @click="confirmDelete(policy)" title="删除">🗑️</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="loading-state" v-else>
        <div class="spinner"></div>
        加载中...
      </div>
    </div>

    <!-- 创建/编辑模态框 -->
    <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
      <div class="modal-container">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑清理策略' : '创建清理策略' }}</h3>
          <button class="modal-close" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitForm">
            <div class="form-group">
              <label class="form-label required">策略名称</label>
              <input type="text" v-model="formData.name" class="form-input" placeholder="如：开发镜像清理" required />
            </div>
            
            <div class="form-group">
              <label class="form-label required">关联仓库</label>
              <select v-model="formData.registry_id" class="form-select" required>
                <option value="">请选择</option>
                <option v-for="reg in registries" :key="reg.id" :value="reg.id">
                  {{ reg.name }}
                </option>
              </select>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">镜像匹配模式</label>
                <input type="text" v-model="formData.repository_pattern" class="form-input" placeholder="如：dev/* 或 *" />
                <span class="form-hint">支持通配符 *，如 dev/* 匹配 dev 下所有镜像</span>
              </div>
              <div class="form-group">
                <label class="form-label">标签匹配模式</label>
                <input type="text" v-model="formData.tag_pattern" class="form-input" placeholder="如：* 或 v*" />
                <span class="form-hint">支持通配符 *</span>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">保留最近版本数</label>
                <input type="number" v-model.number="formData.keep_last_count" class="form-input" min="1" max="100" />
                <span class="form-hint">每个镜像保留最近 N 个标签</span>
              </div>
              <div class="form-group">
                <label class="form-label">保留天数</label>
                <input type="number" v-model.number="formData.keep_days" class="form-input" min="1" max="365" />
                <span class="form-hint">保留最近 N 天内的镜像</span>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">执行计划 (Cron)</label>
              <input type="text" v-model="formData.cron_expression" class="form-input" placeholder="0 2 * * *" />
              <span class="form-hint">Cron 表达式，如 "0 2 * * *" 表示每天凌晨2点执行</span>
            </div>

            <div class="form-group">
              <label class="form-label">描述</label>
              <textarea v-model="formData.description" class="form-textarea" rows="2" placeholder="策略用途说明..."></textarea>
            </div>

            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="formData.enabled" />
                <span>立即启用此策略</span>
              </label>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-cancel" @click="closeModal">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '提交中...' : (isEdit ? '保存' : '创建') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 日志模态框 -->
    <div class="modal-overlay" v-if="showLogsModal" @click.self="showLogsModal = false">
      <div class="modal-container logs-modal">
        <div class="modal-header">
          <h3>执行日志 - {{ selectedPolicy?.name }}</h3>
          <button class="modal-close" @click="showLogsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="logs-list" v-if="logs.length > 0">
            <div class="log-item" v-for="log in logs" :key="log.id" :class="log.status">
              <div class="log-header">
                <span class="log-time">{{ formatTime(log.start_time) }}</span>
                <span class="log-status" :class="log.status">{{ log.status }}</span>
              </div>
              <div class="log-stats">
                <span>扫描: {{ log.scanned_count }}</span>
                <span>删除: {{ log.deleted_count }}</span>
                <span>释放: {{ formatSize(log.freed_size) }}</span>
                <span>耗时: {{ formatDuration(log.start_time, log.end_time) }}</span>
              </div>
              <div class="log-error" v-if="log.error_message">{{ log.error_message }}</div>
            </div>
          </div>
          <div class="empty-logs" v-else>暂无执行日志</div>
        </div>
      </div>
    </div>

    <!-- 删除确认 -->
    <div class="modal-overlay" v-if="showDeleteConfirm" @click.self="showDeleteConfirm = false">
      <div class="modal-container confirm-modal">
        <div class="modal-header danger">
          <h3>⚠️ 确认删除</h3>
        </div>
        <div class="modal-body">
          <p>确定要删除策略 <strong>{{ deleteTarget?.name }}</strong> 吗？</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-cancel" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="doDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, reactive, onMounted } from 'vue'
import {
  getAllRegistries,
  getCleanupPolicies,
  createCleanupPolicy,
  updateCleanupPolicy,
  deleteCleanupPolicy,
  toggleCleanupPolicy,
  runCleanupPolicy,
  getCleanupLogs
} from '@/api/image.js'
import permissionStore from '@/stores/permission'
import { useConfirmDialog } from '@/composables/useConfirmDialog'

export default {
  name: 'CleanupPolicies',
  setup() {
    const { confirm: showConfirm } = useConfirmDialog()
    const policies = ref([])
    const registries = ref([])
    const loading = ref(true)
    
    // ===== 操作权限控制 =====
    const canOperate = computed(() => {
      if (permissionStore.state.isSuperAdmin) return true
      const roleTypes = permissionStore.roleTypes.value
      if (roleTypes.length === 1 && roleTypes.includes('viewer')) return false
      return roleTypes.some(r => ['super_admin', 'platform_admin'].includes(r))
    })
    
    // 模态框
    const showModal = ref(false)
    const isEdit = ref(false)
    const submitting = ref(false)
    const formData = reactive({
      id: null,
      registry_id: '',
      name: '',
      enabled: true,
      repository_pattern: '*',
      tag_pattern: '*',
      keep_last_count: 5,
      keep_days: 30,
      cron_expression: '0 2 * * *',
      description: ''
    })
    
    // 日志
    const showLogsModal = ref(false)
    const selectedPolicy = ref(null)
    const logs = ref([])
    
    // 删除
    const showDeleteConfirm = ref(false)
    const deleteTarget = ref(null)
    const deleting = ref(false)
    
    // 统计
    const enabledCount = computed(() => policies.value.filter(p => p.enabled).length)
    const totalDeleted = computed(() => policies.value.reduce((sum, p) => sum + (p.deleted_count || 0), 0))
    
    // 加载数据
    const loadData = async () => {
      loading.value = true
      try {
        const [policyRes, regRes] = await Promise.all([
          getCleanupPolicies({ page: 1, page_size: 100 }),
          getAllRegistries()
        ])
        if (policyRes.code === 0) {
          policies.value = policyRes.data?.list || []
        }
        if (regRes.code === 0) {
          registries.value = regRes.data?.list || []
        }
      } catch (error) {
        console.error('加载失败:', error)
      } finally {
        loading.value = false
      }
    }
    
    // 创建/编辑
    const openCreateModal = () => {
      isEdit.value = false
      Object.assign(formData, {
        id: null,
        registry_id: '',
        name: '',
        enabled: true,
        repository_pattern: '*',
        tag_pattern: '*',
        keep_last_count: 5,
        keep_days: 30,
        cron_expression: '0 2 * * *',
        description: ''
      })
      showModal.value = true
    }
    
    const openEditModal = (policy) => {
      isEdit.value = true
      Object.assign(formData, {
        id: policy.id,
        registry_id: policy.registry_id,
        name: policy.name,
        enabled: policy.enabled,
        repository_pattern: policy.repository_pattern,
        tag_pattern: policy.tag_pattern,
        keep_last_count: policy.keep_last_count,
        keep_days: policy.keep_days,
        cron_expression: policy.cron_expression,
        description: policy.description
      })
      showModal.value = true
    }
    
    const closeModal = () => {
      showModal.value = false
    }
    
    const submitForm = async () => {
      submitting.value = true
      try {
        const data = { ...formData }
        const res = isEdit.value 
          ? await updateCleanupPolicy(data)
          : await createCleanupPolicy(data)
        
        if (res.code === 0) {
          alert(isEdit.value ? '更新成功' : '创建成功')
          closeModal()
          loadData()
        } else {
          alert(res.msg || '操作失败')
        }
      } catch (error) {
        alert('操作失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }
    
    // 启用/禁用
    const togglePolicy = async (policy) => {
      try {
        const res = await toggleCleanupPolicy(policy.id, !policy.enabled)
        if (res.code === 0) {
          policy.enabled = !policy.enabled
        } else {
          alert(res.msg || '操作失败')
        }
      } catch (error) {
        alert('操作失败: ' + error.message)
      }
    }
    
    // 立即执行
    const runPolicy = async (policy) => {
      const ok = await showConfirm({
        title: '确认立即执行',
        content: '将立即执行此清理策略。',
        type: 'warning',
        details: [
          { label: '策略名称', value: policy.name },
        ],
        confirmText: '确认执行',
        cancelText: '取消',
      })
      if (!ok) return
      
      try {
        const res = await runCleanupPolicy(policy.id)
        if (res.code === 0) {
          alert('已开始执行，请稍后查看日志')
        } else {
          alert(res.msg || '执行失败')
        }
      } catch (error) {
        alert('执行失败: ' + error.message)
      }
    }
    
    // 查看日志
    const viewLogs = async (policy) => {
      selectedPolicy.value = policy
      showLogsModal.value = true
      logs.value = []
      
      try {
        const res = await getCleanupLogs({ policy_id: policy.id, limit: 20 })
        if (res.code === 0) {
          logs.value = res.data?.list || []
        }
      } catch (error) {
        console.error('获取日志失败:', error)
      }
    }
    
    // 删除
    const confirmDelete = (policy) => {
      deleteTarget.value = policy
      showDeleteConfirm.value = true
    }
    
    const doDelete = async () => {
      if (!deleteTarget.value) return
      
      deleting.value = true
      try {
        const res = await deleteCleanupPolicy(deleteTarget.value.id)
        if (res.code === 0) {
          alert('删除成功')
          showDeleteConfirm.value = false
          loadData()
        } else {
          alert(res.msg || '删除失败')
        }
      } catch (error) {
        alert('删除失败: ' + error.message)
      } finally {
        deleting.value = false
      }
    }
    
    // 工具函数
    const formatTime = (ts) => {
      if (!ts) return '-'
      return new Date(ts * 1000).toLocaleString('zh-CN')
    }
    
    const formatSize = (bytes) => {
      if (!bytes) return '0 B'
      const units = ['B', 'KB', 'MB', 'GB']
      let i = 0
      while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024
        i++
      }
      return bytes.toFixed(1) + ' ' + units[i]
    }
    
    const formatDuration = (start, end) => {
      if (!start || !end) return '-'
      const seconds = end - start
      if (seconds < 60) return seconds + '秒'
      return Math.floor(seconds / 60) + '分' + (seconds % 60) + '秒'
    }
    
    onMounted(() => {
      loadData()
    })
    
    return {
      policies, registries, loading,
      showModal, isEdit, submitting, formData,
      showLogsModal, selectedPolicy, logs,
      showDeleteConfirm, deleteTarget, deleting,
      enabledCount, totalDeleted,
      canOperate,
      loadData, openCreateModal, openEditModal, closeModal, submitForm,
      togglePolicy, runPolicy, viewLogs,
      confirmDelete, doDelete,
      formatTime, formatSize, formatDuration
    }
  }
}
</script>

<style scoped>
.cleanup-policy-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.title-icon { font-size: 28px; }

.page-desc {
  margin: 0;
  color: #718096;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.stat-icon {
  font-size: 32px;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f7fafc;
  border-radius: 12px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
}

.stat-label {
  font-size: 13px;
  color: #718096;
}

/* 表格 */
.table-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 14px 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: #4a5568;
  background: #f7fafc;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #edf2f7;
  vertical-align: top;
}

.policy-name .name-text {
  font-weight: 600;
  color: #2d3748;
  display: block;
}

.policy-name .policy-desc {
  font-size: 12px;
  color: #718096;
  margin-top: 2px;
}

.pattern-cell, .keep-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pattern {
  font-family: monospace;
  font-size: 12px;
  background: #f7fafc;
  padding: 2px 6px;
  border-radius: 4px;
}

.cron {
  font-family: monospace;
  font-size: 12px;
  background: #f0fff4;
  color: #276749;
  padding: 4px 8px;
  border-radius: 4px;
}

.keep-cell span {
  font-size: 13px;
  color: #4a5568;
}

/* 开关 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #cbd5e0;
  transition: .3s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: #48bb78;
}

input:checked + .toggle-slider:before {
  transform: translateX(20px);
}

/* 上次执行 */
.last-run {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.run-time {
  font-size: 12px;
  color: #718096;
}

.run-result {
  font-size: 12px;
  color: #4a5568;
}

.no-run {
  font-size: 13px;
  color: #a0aec0;
}

/* 操作按钮 */
.action-group {
  display: flex;
  gap: 6px;
}

.action-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: #f7fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.action-btn:hover { background: #edf2f7; }
.action-btn.delete:hover { background: #fed7d7; }
.action-btn.run:hover { background: #c6f6d5; }

/* 空状态 */
.empty-cell {
  text-align: center;
  padding: 60px !important;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #718096;
}

.empty-icon { font-size: 48px; }

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px;
  color: #718096;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #e2e8f0;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66,153,225,0.35);
}

.btn-refresh {
  background: #edf2f7;
  color: #4a5568;
}

.btn-cancel {
  background: #edf2f7;
  color: #4a5568;
}

.btn-danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
  color: white;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-container {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.logs-modal {
  max-width: 700px;
}

.confirm-modal {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #4299e1, #3182ce);
  color: white;
}

.modal-header.danger {
  background: linear-gradient(135deg, #f56565, #e53e3e);
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
}

.modal-close {
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(255,255,255,0.2);
  color: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 18px;
}

.modal-body {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e2e8f0;
}

/* 表单 */
.form-group {
  margin-bottom: 16px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
}

.form-label.required::after {
  content: ' *';
  color: #e53e3e;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66,153,225,0.15);
}

.form-hint {
  font-size: 12px;
  color: #718096;
  margin-top: 4px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

/* 日志列表 */
.logs-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-item {
  padding: 12px;
  background: #f7fafc;
  border-radius: 8px;
  border-left: 3px solid #cbd5e0;
}

.log-item.success {
  border-left-color: #48bb78;
}

.log-item.failed {
  border-left-color: #f56565;
}

.log-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.log-time {
  font-size: 13px;
  color: #718096;
}

.log-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
}

.log-status.success {
  background: #c6f6d5;
  color: #276749;
}

.log-status.failed {
  background: #fed7d7;
  color: #c53030;
}

.log-status.running {
  background: #fefcbf;
  color: #975a16;
}

.log-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #4a5568;
}

.log-error {
  margin-top: 8px;
  font-size: 12px;
  color: #c53030;
  background: #fff5f5;
  padding: 8px;
  border-radius: 4px;
}

.empty-logs {
  text-align: center;
  padding: 40px;
  color: #718096;
}
</style>


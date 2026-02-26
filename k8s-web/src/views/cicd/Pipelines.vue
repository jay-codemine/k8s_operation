<template>
  <div class="pipeline-management">
    <div class="view-header">
      <h1>🚀 流水线管理</h1>
      <p>管理和监控 CI/CD 流水线</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input
          v-model="searchQuery"
          placeholder="搜索流水线名称、描述或 Git 仓库..."
          class="search-input"
        />
      </div>
      <div class="action-buttons">
        <button class="btn btn-primary" @click="createPipeline">
          + 创建流水线
        </button>
        <button class="btn btn-secondary" @click="loadPipelines" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>
    <div v-if="loading && pipelines.length === 0" class="loading-state">加载中...</div>

    <div v-else class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 80px;">ID</th>
            <th>流水线名称</th>
            <th>描述</th>
            <th style="width: 100px;">状态</th>
            <th style="width: 120px;">上次运行</th>
            <th style="width: 160px;">运行时间</th>
            <th>Git仓库</th>
            <th style="width: 100px;">分支</th>
            <th style="width: 260px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="paginatedPipelines.length === 0">
            <td colspan="9" class="empty-row">
              <div class="empty-state">
                <div class="empty-icon">📋</div>
                <div class="empty-text">暂无流水线数据</div>
                <div class="empty-hint">点击上方“创建流水线”按钮开始</div>
              </div>
            </td>
          </tr>
          <tr v-for="pipeline in paginatedPipelines" :key="pipeline.id">
            <td>{{ pipeline.id }}</td>
            <td>
              <div class="pipeline-name">
                <span class="icon">🚀</span>
                <span>{{ pipeline.name }}</span>
              </div>
            </td>
            <td>{{ pipeline.description }}</td>
            <td>
              <span :class="['status-tag', `status-${pipeline.status}`]">
                {{ statusText(pipeline.status) }}
              </span>
            </td>
            <td>
              <span :class="['status-tag', `status-${pipeline.lastRunStatus}`]">
                {{ runStatusText(pipeline.lastRunStatus) }}
              </span>
            </td>
            <td>{{ formatDate(pipeline.lastRunTime) }}</td>
            <td class="git-repo">{{ pipeline.gitRepo }}</td>
            <td>
              <span class="branch-tag">🌿 {{ pipeline.branch }}</span>
            </td>
            <td>
              <div class="action-buttons">
                <button class="btn btn-sm btn-view" @click="viewPipeline(pipeline.id)" title="查看详情">
                  👁️ 查看
                </button>
                <button class="btn btn-sm btn-primary" @click="runPipeline(pipeline.id)" title="运行流水线">
                  ▶️ 运行
                </button>
                <button class="btn btn-sm btn-danger" @click="deletePipeline(pipeline.id)" title="删除">
                  🗑️ 删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination
        v-if="total > 0"
        v-model:currentPage="currentPage"
        :totalItems="total"
        :itemsPerPage="pageSize"
      />
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'
import {
  getPipelines as fetchPipelines,
  runPipeline as triggerPipeline,
  deletePipeline as removePipeline
} from '@/api/platform/pipeline'

export default {
  name: 'Pipelines',
  components: {
    Pagination
  },
  setup() {
    const router = useRouter()
    const pipelines = ref([])
    const searchQuery = ref('')
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const loading = ref(false)
    const errorMsg = ref('')

    // 调用后端 API 加载流水线列表
    const loadPipelines = async () => {
      loading.value = true
      errorMsg.value = ''
      try {
        const response = await fetchPipelines({
          page: currentPage.value,
          page_size: pageSize.value,
          keyword: searchQuery.value || undefined
        })
        
        if (response.code === 0) {
          // 后端返回的数据结构转换为前端所需格式
          pipelines.value = (response.data?.list || []).map(item => ({
            id: item.id,
            name: item.name,
            description: item.description,
            status: item.status,
            lastRunStatus: item.last_run_status || 'pending',
            lastRunTime: item.last_run_time ? new Date(item.last_run_time * 1000).toISOString() : null,
            gitRepo: item.git_repo,
            branch: item.git_branch
          }))
          total.value = response.data?.total || 0
        } else {
          throw new Error(response.msg || '获取流水线列表失败')
        }
      } catch (error) {
        console.error('加载流水线失败:', error)
        errorMsg.value = error.message || '获取流水线列表失败'
        pipelines.value = []
        total.value = 0
      } finally {
        loading.value = false
      }
    }

    // 搜索时重置页码并重新加载
    watch(searchQuery, () => {
      currentPage.value = 1
      loadPipelines()
    })

    // 页码变化时重新加载
    watch(currentPage, () => {
      loadPipelines()
    })

    const filteredPipelines = computed(() => pipelines.value)

    const paginatedPipelines = computed(() => pipelines.value)

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      const date = new Date(dateString)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      })
    }

    const statusText = (status) => {
      const statusMap = {
        idle: '空闲',
        running: '运行中',
        disabled: '已禁用',
        error: '错误'
      }
      return statusMap[status] || status
    }

    const runStatusText = (status) => {
      const statusMap = {
        success: '成功',
        failed: '失败',
        running: '运行中',
        pending: '等待中',
        aborted: '已中止',
        cancelled: '已取消'
      }
      return statusMap[status] || status
    }

    const createPipeline = () => {
      router.push('/cicd/pipelines/create')
    }

    const viewPipeline = (id) => {
      router.push(`/cicd/pipelines/${id}`)
    }

    // 调用后端 API 运行流水线（触发 Jenkins 构建）
    const runPipeline = async (id) => {
      try {
        Message.info({ content: `正在启动流水线 #${id}...` })
        
        const response = await triggerPipeline(id)
        
        if (response.code === 0) {
          Message.success({ content: '流水线启动成功，正在触发 Jenkins 构建' })
          // 重新加载列表以更新状态
          setTimeout(() => loadPipelines(), 1000)
        } else {
          throw new Error(response.msg || '启动流水线失败')
        }
      } catch (error) {
        console.error('启动流水线失败:', error)
        Message.error({ content: error.message || '启动流水线失败' })
      }
    }

    // 调用后端 API 删除流水线
    const deletePipeline = async (id) => {
      if (!confirm('确定要删除这条流水线吗？此操作不可恢复！')) {
        return
      }
      
      try {
        Message.info({ content: `正在删除流水线 #${id}...` })
        
        const response = await removePipeline(id)
        
        if (response.code === 0) {
          Message.success({ content: '删除流水线成功' })
          loadPipelines()
        } else {
          throw new Error(response.msg || '删除流水线失败')
        }
      } catch (error) {
        console.error('删除流水线失败:', error)
        Message.error({ content: error.message || '删除流水线失败' })
      }
    }

    onMounted(() => {
      loadPipelines()
    })

    return {
      pipelines,
      searchQuery,
      currentPage,
      pageSize,
      total,
      loading,
      errorMsg,
      filteredPipelines,
      paginatedPipelines,
      formatDate,
      statusText,
      runStatusText,
      createPipeline,
      viewPipeline,
      runPipeline,
      deletePipeline,
      loadPipelines
    }
  }
}
</script>

<style scoped>
.pipeline-management {
  padding: 24px;
  max-width: 1600px;
  margin: 0 auto;
}

/* 视图头部 */
.view-header {
  margin-bottom: 24px;
}

.view-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #1a202c;
  margin: 0 0 8px 0;
}

.view-header p {
  color: #718096;
  font-size: 14px;
  margin: 0;
}

/* 操作栏 */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.search-input {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.action-buttons {
  display: flex;
  gap: 12px;
}

/* 错误和加载状态 */
.error-box {
  background: #fff5f5;
  border: 1px solid #fc8181;
  color: #c53030;
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;
}

.loading-state {
  text-align: center;
  padding: 60px 20px;
  color: #718096;
  font-size: 16px;
}

/* 表格容器 */
.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 14px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.data-table th {
  background-color: #f7fafc;
  font-weight: 600;
  color: #4a5568;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.data-table tbody tr:hover {
  background-color: #f7fafc;
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

/* 空状态 */
.empty-row {
  text-align: center;
}

.empty-state {
  padding: 60px 20px;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-text {
  font-size: 16px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 14px;
  color: #a0aec0;
}

/* 流水线名称 */
.pipeline-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #2d3748;
}

.pipeline-name .icon {
  font-size: 16px;
}

/* Git 仓库 */
.git-repo {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: #4a5568;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 分支标签 */
.branch-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: #edf2f7;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  color: #2d3748;
}

/* 状态标签 */
.status-tag {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status-idle {
  background-color: #e6f7ff;
  color: #0958d9;
}

.status-running {
  background-color: #fff7e6;
  color: #d46b08;
}

.status-disabled {
  background-color: #f5f5f5;
  color: #8c8c8c;
}

.status-error {
  background-color: #fff1f0;
  color: #cf1322;
}

.status-success {
  background-color: #f6ffed;
  color: #389e0d;
}

.status-failed {
  background-color: #fff1f0;
  color: #cf1322;
}

.status-pending {
  background-color: #fafafa;
  color: #595959;
}

.status-cancelled {
  background-color: #f0f0f0;
  color: #8c8c8c;
}

/* 按钮 */
.btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  background: white;
  color: #4a5568;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2554c7;
}

.btn-secondary {
  background-color: #718096;
  color: white;
  border-color: #718096;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #4a5568;
}

.btn-danger {
  background-color: #e53e3e;
  color: white;
  border-color: #e53e3e;
}

.btn-danger:hover:not(:disabled) {
  background-color: #c53030;
}

.btn-view {
  background-color: #4a5568;
  color: white;
  border-color: #4a5568;
}

.btn-view:hover:not(:disabled) {
  background-color: #2d3748;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>

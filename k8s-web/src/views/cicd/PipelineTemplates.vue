<template>
  <div class="pipeline-templates-container">
    <div class="view-header">
      <h1>📋 流水线模板管理</h1>
      <p>管理和配置 CI/CD 流水线模板</p>
    </div>

    <div class="action-bar">
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索模板名称、类型或描述..."
          class="search-input"
        />
        <select v-model="typeFilter" class="type-filter">
          <option value="">所有类型</option>
          <option value="frontend">前端应用</option>
          <option value="backend">后端服务</option>
          <option value="database">数据库</option>
          <option value="microservice">微服务</option>
        </select>
      </div>
      <div class="action-buttons">
        <button class="btn btn-primary" @click="showCreateModal = true">
          + 创建模板
        </button>
        <button class="btn btn-secondary" @click="loadTemplates" :disabled="loading">
          {{ loading ? '加载中...' : '🔄 刷新' }}
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-box">{{ errorMsg }}</div>
    <div v-if="loading && pipelineTemplates.length === 0" class="loading-state">加载中...</div>

    <div v-else class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 80px;">ID</th>
            <th>模板名称</th>
            <th style="width: 120px;">类型</th>
            <th>描述</th>
            <th style="width: 100px;">阶段数</th>
            <th style="width: 160px;">创建时间</th>
            <th style="width: 160px;">更新时间</th>
            <th style="width: 240px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="paginatedTemplates.length === 0">
            <td colspan="8" class="empty-row">
              <div class="empty-state">
                <div class="empty-icon">📋</div>
                <div class="empty-text">暂无模板数据</div>
                <div class="empty-hint">点击上方"创建模板"按钮开始</div>
              </div>
            </td>
          </tr>
          <tr v-for="template in paginatedTemplates" :key="template.id">
            <td>{{ template.id }}</td>
            <td>
              <div class="template-name">
                <span class="icon">📋</span>
                <span>{{ template.name }}</span>
              </div>
            </td>
            <td>
              <span :class="['type-badge', template.type]">
                {{ typeText(template.type) }}
              </span>
            </td>
            <td class="description">{{ template.description }}</td>
            <td class="text-center">
              <span class="stage-count">{{ template.stages.length }} 个</span>
            </td>
            <td>{{ formatDate(template.createdAt) }}</td>
            <td>{{ formatDate(template.updatedAt) }}</td>
            <td>
              <div class="action-buttons">
                <button class="btn btn-sm btn-view" @click="handleDetail(template)" title="查看详情">
                  👁️ 详情
                </button>
                <button class="btn btn-sm btn-edit" @click="handleEdit(template)" title="编辑">
                  ✏️ 编辑
                </button>
                <button class="btn btn-sm btn-danger" @click="handleDelete(template)" title="删除">
                  🗑️ 删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination
        v-if="filteredTemplates.length > 0"
        v-model:currentPage="currentPage"
        :totalItems="filteredTemplates.length"
        :itemsPerPage="itemsPerPage"
      />
    </div>

    <!-- 创建模板模态框 -->
    <div v-if="showCreateModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>创建流水线模板</h2>
          <button class="close-btn" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="createTemplate">
            <div class="form-group">
              <label for="create-name">模板名称</label>
              <input
                id="create-name"
                v-model="templateForm.name"
                type="text"
                required
              />
            </div>
            <div class="form-group">
              <label for="create-description">描述</label>
              <textarea
                id="create-description"
                v-model="templateForm.description"
                rows="3"
              ></textarea>
            </div>
            <div class="form-group">
              <label for="create-type">模板类型</label>
              <select id="create-type" v-model="templateForm.type" required>
                <option value="frontend">前端应用</option>
                <option value="backend">后端服务</option>
                <option value="database">数据库</option>
              </select>
            </div>
            <div class="form-group">
              <label>默认环境变量</label>
              <div class="env-vars-list" v-for="(envVar, index) in templateForm.defaultEnvVars" :key="index">
                <div class="env-var-row">
                  <input
                    v-model="envVar.name"
                    type="text"
                    placeholder="变量名"
                    class="env-var-name"
                  />
                  <input
                    v-model="envVar.value"
                    type="text"
                    placeholder="变量值"
                    class="env-var-value"
                  />
                  <button
                    type="button"
                    class="remove-env-var-btn"
                    @click="templateForm.defaultEnvVars.splice(index, 1)"
                  >
                    删除
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="add-env-var-btn"
                @click="templateForm.defaultEnvVars.push({ name: '', value: '' })"
              >
                添加环境变量
              </button>
            </div>
            <div class="form-group">
              <label>默认部署配置</label>
              <div class="deployment-config">
                <div class="config-row">
                  <label>副本数:</label>
                  <input
                    v-model.number="templateForm.defaultDeploymentConfig.replicas"
                    type="number"
                    min="1"
                  />
                </div>
                <div class="config-row">
                  <label>部署策略:</label>
                  <select v-model="templateForm.defaultDeploymentConfig.strategy">
                    <option value="rollingUpdate">滚动更新</option>
                    <option value="recreate">重新创建</option>
                  </select>
                </div>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" class="cancel-btn" @click="showCreateModal = false">取消</button>
              <button type="submit" class="submit-btn">创建</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 编辑模板模态框 -->
    <div v-if="showEditModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>编辑流水线模板</h2>
          <button class="close-btn" @click="showEditModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="updateTemplate">
            <div class="form-group">
              <label for="edit-name">模板名称</label>
              <input
                id="edit-name"
                v-model="templateForm.name"
                type="text"
                required
              />
            </div>
            <div class="form-group">
              <label for="edit-description">描述</label>
              <textarea
                id="edit-description"
                v-model="templateForm.description"
                rows="3"
              ></textarea>
            </div>
            <div class="form-group">
              <label for="edit-type">模板类型</label>
              <select id="edit-type" v-model="templateForm.type" required>
                <option value="frontend">前端应用</option>
                <option value="backend">后端服务</option>
                <option value="database">数据库</option>
              </select>
            </div>
            <div class="form-group">
              <label>默认环境变量</label>
              <div class="env-vars-list" v-for="(envVar, index) in templateForm.defaultEnvVars" :key="index">
                <div class="env-var-row">
                  <input
                    v-model="envVar.name"
                    type="text"
                    placeholder="变量名"
                    class="env-var-name"
                  />
                  <input
                    v-model="envVar.value"
                    type="text"
                    placeholder="变量值"
                    class="env-var-value"
                  />
                  <button
                    type="button"
                    class="remove-env-var-btn"
                    @click="templateForm.defaultEnvVars.splice(index, 1)"
                  >
                    删除
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="add-env-var-btn"
                @click="templateForm.defaultEnvVars.push({ name: '', value: '' })"
              >
                添加环境变量
              </button>
            </div>
            <div class="form-group">
              <label>默认部署配置</label>
              <div class="deployment-config">
                <div class="config-row">
                  <label>副本数:</label>
                  <input
                    v-model.number="templateForm.defaultDeploymentConfig.replicas"
                    type="number"
                    min="1"
                  />
                </div>
                <div class="config-row">
                  <label>部署策略:</label>
                  <select v-model="templateForm.defaultDeploymentConfig.strategy">
                    <option value="rollingUpdate">滚动更新</option>
                    <option value="recreate">重新创建</option>
                  </select>
                </div>
              </div>
            </div>
            <div class="form-actions">
              <button type="button" class="cancel-btn" @click="showEditModal = false">取消</button>
              <button type="submit" class="submit-btn">更新</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 模板详情模态框 -->
    <div v-if="showDetailModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ selectedTemplate.name }} - 详情</h2>
          <button class="close-btn" @click="showDetailModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h3>基本信息</h3>
            <div class="detail-item">
              <label>模板名称:</label>
              <span>{{ selectedTemplate.name }}</span>
            </div>
            <div class="detail-item">
              <label>类型:</label>
              <span>{{ selectedTemplate.type }}</span>
            </div>
            <div class="detail-item">
              <label>描述:</label>
              <span>{{ selectedTemplate.description }}</span>
            </div>
            <div class="detail-item">
              <label>创建时间:</label>
              <span>{{ formatDate(selectedTemplate.createdAt) }}</span>
            </div>
            <div class="detail-item">
              <label>更新时间:</label>
              <span>{{ formatDate(selectedTemplate.updatedAt) }}</span>
            </div>
          </div>

          <div class="detail-section">
            <h3>流水线阶段</h3>
            <ul class="stages-list">
              <li v-for="(stage, index) in selectedTemplate.stages" :key="index" class="stage-item">
                <div class="stage-name">{{ index + 1 }}. {{ stage.name }}</div>
                <div class="stage-description">{{ stage.description }}</div>
              </li>
            </ul>
          </div>

          <div class="detail-section">
            <h3>默认环境变量</h3>
            <table class="env-vars-table">
              <thead>
                <tr>
                  <th>变量名</th>
                  <th>变量值</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(envVar, index) in selectedTemplate.defaultEnvVars" :key="index">
                  <td>{{ envVar.name }}</td>
                  <td>{{ envVar.value }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="detail-section">
            <h3>默认部署配置</h3>
            <div class="deployment-config-detail">
              <div class="config-item">
                <label>副本数:</label>
                <span>{{ selectedTemplate.defaultDeploymentConfig.replicas }}</span>
              </div>
              <div class="config-item">
                <label>部署策略:</label>
                <span>{{ selectedTemplate.defaultDeploymentConfig.strategy }}</span>
              </div>
              <div v-if="selectedTemplate.defaultDeploymentConfig.resources" class="config-item">
                <label>资源限制:</label>
                <div class="resources">
                  <div>CPU: {{ selectedTemplate.defaultDeploymentConfig.resources.requests.cpu }} / {{ selectedTemplate.defaultDeploymentConfig.resources.limits.cpu }}</div>
                  <div>内存: {{ selectedTemplate.defaultDeploymentConfig.resources.requests.memory }} / {{ selectedTemplate.defaultDeploymentConfig.resources.limits.memory }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import Pagination from '@/components/Pagination.vue'

// 静态模拟数据
const MOCK_TEMPLATES = [
  {
    id: 1,
    name: 'Vue3 前端项目模板',
    description: '适用于 Vue3 + Vite 前端项目的标准流水线',
    type: 'frontend',
    stages: [
      { name: 'checkout', description: '拉取代码' },
      { name: 'install', description: '安装依赖 (npm install)' },
      { name: 'build', description: '构建应用 (npm run build)' },
      { name: 'test', description: '运行测试 (npm run test)' },
      { name: 'build-image', description: '构建 Docker 镜像' },
      { name: 'deploy', description: '部署到 K8s' }
    ],
    defaultEnvVars: [
      { name: 'NODE_ENV', value: 'production' },
      { name: 'VITE_API_BASE', value: 'https://api.example.com' }
    ],
    defaultDeploymentConfig: {
      replicas: 3,
      strategy: 'rollingUpdate',
      resources: {
        limits: { cpu: '500m', memory: '512Mi' },
        requests: { cpu: '200m', memory: '256Mi' }
      }
    },
    createdAt: '2024-01-15 10:30:00',
    updatedAt: '2024-02-01 14:20:00'
  },
  {
    id: 2,
    name: 'Go 微服务模板',
    description: '适用于 Go 语言微服务的标准流水线',
    type: 'backend',
    stages: [
      { name: 'checkout', description: '拉取代码' },
      { name: 'test', description: '运行单元测试' },
      { name: 'build', description: '编译 Go 二进制' },
      { name: 'build-image', description: '构建 Docker 镜像' },
      { name: 'push-image', description: '推送镜像到仓库' },
      { name: 'deploy', description: '部署到 K8s' }
    ],
    defaultEnvVars: [
      { name: 'GO_ENV', value: 'production' },
      { name: 'CGO_ENABLED', value: '0' },
      { name: 'GOOS', value: 'linux' }
    ],
    defaultDeploymentConfig: {
      replicas: 5,
      strategy: 'rollingUpdate',
      resources: {
        limits: { cpu: '1000m', memory: '1Gi' },
        requests: { cpu: '500m', memory: '512Mi' }
      }
    },
    createdAt: '2024-01-20 09:15:00',
    updatedAt: '2024-02-05 16:45:00'
  },
  {
    id: 3,
    name: 'MySQL 数据库模板',
    description: '适用于 MySQL 数据库迁移和初始化',
    type: 'database',
    stages: [
      { name: 'checkout', description: '拉取代码' },
      { name: 'validate', description: '验证 SQL 脚本' },
      { name: 'backup', description: '备份现有数据库' },
      { name: 'migrate', description: '执行数据库迁移' },
      { name: 'verify', description: '验证迁移结果' }
    ],
    defaultEnvVars: [
      { name: 'DB_HOST', value: 'mysql-service' },
      { name: 'DB_PORT', value: '3306' },
      { name: 'DB_NAME', value: 'mydb' }
    ],
    defaultDeploymentConfig: {
      replicas: 1,
      strategy: 'recreate',
      resources: {
        limits: { cpu: '200m', memory: '256Mi' },
        requests: { cpu: '100m', memory: '128Mi' }
      }
    },
    createdAt: '2024-02-01 11:00:00',
    updatedAt: '2024-02-08 10:30:00'
  },
  {
    id: 4,
    name: 'Spring Boot 微服务模板',
    description: '适用于 Spring Boot 微服务项目',
    type: 'microservice',
    stages: [
      { name: 'checkout', description: '拉取代码' },
      { name: 'test', description: '运行单元测试' },
      { name: 'build', description: 'Maven 构建' },
      { name: 'sonar', description: '代码质量分析' },
      { name: 'build-image', description: '构建 Docker 镜像' },
      { name: 'deploy', description: '部署到 K8s' }
    ],
    defaultEnvVars: [
      { name: 'JAVA_OPTS', value: '-Xmx512m -Xms256m' },
      { name: 'SPRING_PROFILES_ACTIVE', value: 'prod' }
    ],
    defaultDeploymentConfig: {
      replicas: 4,
      strategy: 'rollingUpdate',
      resources: {
        limits: { cpu: '2000m', memory: '2Gi' },
        requests: { cpu: '1000m', memory: '1Gi' }
      }
    },
    createdAt: '2024-02-03 14:00:00',
    updatedAt: '2024-02-09 09:00:00'
  }
]

const router = useRouter()

// 模板列表数据
const pipelineTemplates = ref([])
const loading = ref(false)
const errorMsg = ref('')

// 搜索和过滤
const searchQuery = ref('')
const typeFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(10)

// 模态框状态
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDetailModal = ref(false)

// 选中的模板
const selectedTemplate = ref(null)

// 表单数据
const templateForm = ref({
  id: null,
  name: '',
  description: '',
  type: 'frontend',
  stages: [
    { name: 'checkout', description: '拉取代码' },
    { name: 'install', description: '安装依赖' },
    { name: 'build', description: '构建应用' },
    { name: 'test', description: '运行测试' },
    { name: 'build-image', description: '构建镜像' },
    { name: 'deploy', description: '部署到K8s' }
  ],
  defaultEnvVars: [
    { name: 'NODE_ENV', value: 'production' }
  ],
  defaultDeploymentConfig: {
    replicas: 3,
    strategy: 'rollingUpdate',
    resources: {
      limits: { cpu: '500m', memory: '512Mi' },
      requests: { cpu: '200m', memory: '256Mi' }
    }
  }
})

// 获取模板列表
const loadTemplates = async () => {
  loading.value = true
  errorMsg.value = ''
  try {
    // 模拟 API 调用延迟
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // 使用静态数据
    pipelineTemplates.value = MOCK_TEMPLATES
    
    // 如果需要调用真实 API，取消下面注释
    // const response = await getPipelineTemplates()
    // if (response.code === 0) {
    //   pipelineTemplates.value = response.data
    // } else {
    //   throw new Error(response.msg || '获取模板列表失败')
    // }
  } catch (error) {
    console.error('加载模板失败:', error)
    errorMsg.value = error.message || '获取模板列表失败，显示模拟数据'
    // 出错时也显示模拟数据
    pipelineTemplates.value = MOCK_TEMPLATES
  } finally {
    loading.value = false
  }
}

// 过滤后的模板
const filteredTemplates = computed(() => {
  return pipelineTemplates.value.filter(template => {
    const query = searchQuery.value.toLowerCase()
    const matchesSearch = !query || 
      template.name.toLowerCase().includes(query) ||
      template.description.toLowerCase().includes(query)
    const matchesType = !typeFilter.value || template.type === typeFilter.value
    return matchesSearch && matchesType
  })
})

// 分页后的模板
const paginatedTemplates = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredTemplates.value.slice(startIndex, endIndex)
})

// 格式化日期
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

// 类型文本映射
const typeText = (type) => {
  const typeMap = {
    frontend: '前端应用',
    backend: '后端服务',
    database: '数据库',
    microservice: '微服务'
  }
  return typeMap[type] || type
}

// 处理详情查看
const handleDetail = (template) => {
  selectedTemplate.value = template
  showDetailModal.value = true
}

// 处理编辑
const handleEdit = (template) => {
  templateForm.value = JSON.parse(JSON.stringify(template))
  showEditModal.value = true
}

// 处理删除
const handleDelete = async (template) => {
  if (!confirm(`确定要删除模板「${template.name}」吗？此操作不可恢复！`)) {
    return
  }
  
  try {
    Message.info({ content: `正在删除模板 #${template.id}...` })
    
    // 模拟删除
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // 从本地移除
    const index = pipelineTemplates.value.findIndex(t => t.id === template.id)
    if (index !== -1) {
      pipelineTemplates.value.splice(index, 1)
    }
    
    Message.success({ content: '删除模板成功' })
    
    // 如果需要调用真实 API，取消下面注释
    // const response = await deletePipelineTemplate(template.id)
    // if (response.code === 0) {
    //   Message.success({ content: '删除模板成功' })
    //   loadTemplates()
    // } else {
    //   throw new Error(response.msg)
    // }
  } catch (error) {
    console.error('删除模板失败:', error)
    Message.error({ content: error.message || '删除模板失败' })
  }
}

// 创建模板
const createTemplate = async () => {
  try {
    Message.info({ content: '正在创建模板...' })
    
    // 模拟创建
    await new Promise(resolve => setTimeout(resolve, 800))
    
    // 添加到本地
    const newTemplate = {
      ...templateForm.value,
      id: pipelineTemplates.value.length + 1,
      createdAt: new Date().toLocaleString('zh-CN'),
      updatedAt: new Date().toLocaleString('zh-CN')
    }
    pipelineTemplates.value.push(newTemplate)
    
    Message.success({ content: '创建模板成功' })
    showCreateModal.value = false
    resetForm()
    
    // 如果需要调用真实 API，取消下面注释
    // const response = await createPipelineTemplate(templateForm.value)
    // if (response.code === 0) {
    //   Message.success({ content: '创建模板成功' })
    //   loadTemplates()
    //   showCreateModal.value = false
    //   resetForm()
    // } else {
    //   throw new Error(response.msg)
    // }
  } catch (error) {
    console.error('创建模板失败:', error)
    Message.error({ content: error.message || '创建模板失败' })
  }
}

// 更新模板
const updateTemplate = async () => {
  try {
    Message.info({ content: '正在更新模板...' })
    
    // 模拟更新
    await new Promise(resolve => setTimeout(resolve, 800))
    
    // 更新本地数据
    const index = pipelineTemplates.value.findIndex(t => t.id === templateForm.value.id)
    if (index !== -1) {
      pipelineTemplates.value[index] = {
        ...templateForm.value,
        updatedAt: new Date().toLocaleString('zh-CN')
      }
    }
    
    Message.success({ content: '更新模板成功' })
    showEditModal.value = false
    resetForm()
    
    // 如果需要调用真实 API，取消下面注释
    // const response = await updatePipelineTemplate(templateForm.value.id, templateForm.value)
    // if (response.code === 0) {
    //   Message.success({ content: '更新模板成功' })
    //   loadTemplates()
    //   showEditModal.value = false
    //   resetForm()
    // } else {
    //   throw new Error(response.msg)
    // }
  } catch (error) {
    console.error('更新模板失败:', error)
    Message.error({ content: error.message || '更新模板失败' })
  }
}

// 重置表单
const resetForm = () => {
  templateForm.value = {
    id: null,
    name: '',
    description: '',
    type: 'frontend',
    stages: [
      { name: 'checkout', description: '拉取代码' },
      { name: 'install', description: '安装依赖' },
      { name: 'build', description: '构建应用' },
      { name: 'test', description: '运行测试' },
      { name: 'build-image', description: '构建镜像' },
      { name: 'deploy', description: '部署到K8s' }
    ],
    defaultEnvVars: [
      { name: 'NODE_ENV', value: 'production' }
    ],
    defaultDeploymentConfig: {
      replicas: 3,
      strategy: 'rollingUpdate',
      resources: {
        limits: { cpu: '500m', memory: '512Mi' },
        requests: { cpu: '200m', memory: '256Mi' }
      }
    }
  }
}

// 初始化获取模板列表
onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.pipeline-templates-container {
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
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  flex: 1;
  max-width: 400px;
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

.type-filter {
  padding: 10px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  min-width: 150px;
  cursor: pointer;
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

/* 模板名称 */
.template-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #2d3748;
}

.template-name .icon {
  font-size: 16px;
}

/* 描述 */
.description {
  color: #718096;
  font-size: 14px;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-center {
  text-align: center;
}

.stage-count {
  display: inline-block;
  padding: 4px 10px;
  background: #edf2f7;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
}

/* 类型标签 */
.type-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.type-badge.frontend {
  background-color: #e6f7ff;
  color: #0958d9;
}

.type-badge.backend {
  background-color: #f6ffed;
  color: #389e0d;
}

.type-badge.database {
  background-color: #fff7e6;
  color: #d46b08;
}

.type-badge.microservice {
  background-color: #f9f0ff;
  color: #531dab;
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

.btn-edit {
  background-color: #ecc94b;
  color: #2d3748;
  border-color: #ecc94b;
}

.btn-edit:hover:not(:disabled) {
  background-color: #d69e2e;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 模态框样式 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  overflow-y: auto;
}

.modal-content {
  background-color: white;
  padding: 24px;
  border-radius: 8px;
  width: 100%;
  max-width: 700px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  margin: 20px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eee;
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
}

.modal-body form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-weight: 500;
  color: #333;
}

.form-group input,
.form-group select,
.form-group textarea {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
}

.env-vars-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 12px;
}

.env-var-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.env-var-name,
.env-var-value {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.env-var-name {
  width: 150px;
}

.env-var-value {
  flex: 1;
}

.remove-env-var-btn,
.add-env-var-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.remove-env-var-btn {
  background-color: #dc3545;
  color: white;
}

.add-env-var-btn {
  background-color: #28a745;
  color: white;
  align-self: flex-start;
}

.deployment-config {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.config-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.config-row label {
  width: 100px;
  font-weight: 500;
}

.config-row input,
.config-row select {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}

.cancel-btn,
.submit-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.cancel-btn {
  background-color: #f8f9fa;
  color: #333;
  border: 1px solid #ddd;
}

.cancel-btn:hover {
  background-color: #e9ecef;
}

.submit-btn {
  background-color: #326ce5;
  color: white;
}

.submit-btn:hover {
  background-color: #2858b8;
}

/* 详情模态框样式 */
.detail-section {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.detail-section h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
}

.detail-item label {
  width: 120px;
  font-weight: 500;
  color: #666;
}

.detail-item span {
  flex: 1;
}

.stages-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.stage-item {
  margin-bottom: 12px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.stage-name {
  font-weight: 600;
  margin-bottom: 4px;
}

.stage-description {
  color: #666;
  font-size: 14px;
}

.env-vars-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 8px;
}

.env-vars-table th,
.env-vars-table td {
  padding: 8px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.env-vars-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.deployment-config-detail {
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.config-item {
  display: flex;
  margin-bottom: 8px;
}

.config-item label {
  width: 120px;
  font-weight: 500;
  color: #666;
}

.config-item span {
  flex: 1;
}

.resources {
  margin-left: 120px;
  margin-top: 4px;
  color: #666;
}
</style>

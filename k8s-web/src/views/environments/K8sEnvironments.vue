<template>
  <div class="k8s-environments">
    <h2>K8s环境管理</h2>

    <div class="toolbar">
      <button class="btn btn-primary" @click="showAddModal = true">创建环境</button>
      <input
        v-model="searchQuery"
        placeholder="搜索环境"
        class="search-input"
      />
    </div>

    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>环境名称</th>
          <th>描述</th>
          <th>环境类型</th>
          <th>集群名称</th>
          <th>API URL</th>
          <th>命名空间</th>
          <th>状态</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="env in paginatedEnvironments" :key="env.id">
          <td>{{ env.id }}</td>
          <td>{{ env.name }}</td>
          <td>{{ env.description }}</td>
          <td>
            <span class="type-tag" :class="`type-${env.type || env.envType}`">
              {{ getEnvTypeName(env.type || env.envType) }}
            </span>
          </td>
          <td>{{ env.clusterName }}</td>
          <td>{{ env.apiUrl }}</td>
          <td>{{ env.namespace }}</td>
          <td>
            <span class="status-tag" :class="`status-${env.status}`">
              <span class="status-icon" v-if="env.status === 'connected'">✓</span>
              <span class="status-icon" v-else>✗</span>
              {{ env.status }}
            </span>
          </td>
          <td>
            <div class="action-buttons">
              <button class="btn btn-view" @click="handleView(env)">查看</button>
              <button class="btn btn-edit" @click="handleEdit(env)">编辑</button>
              <button class="btn btn-danger" @click="handleDelete(env)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <Pagination
      :current-page="currentPage"
      :total-items="filteredEnvironments.length"
      :items-per-page="pageSize"
      @update:currentPage="currentPage = $event"
    />

    <!-- 创建K8s环境模态框 -->
    <div v-if="showAddModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>创建K8s环境</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitEnvironment">
            <div class="form-group">
              <label for="name">环境名称</label>
              <input
                type="text"
                id="name"
                v-model="envForm.name"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="description">描述</label>
              <textarea
                id="description"
                v-model="envForm.description"
                rows="3"
                class="form-textarea"
              ></textarea>
            </div>
            <div class="form-group">
              <label for="clusterName">集群名称</label>
              <input
                type="text"
                id="clusterName"
                v-model="envForm.clusterName"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="apiUrl">API URL</label>
              <input
                type="url"
                id="apiUrl"
                v-model="envForm.apiUrl"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="type">环境类型</label>
              <select
                id="type"
                v-model="envForm.type"
                required
                class="form-input"
              >
                <option value="development">开发环境</option>
                <option value="testing">测试环境</option>
                <option value="staging">预发布环境</option>
                <option value="production">生产环境</option>
              </select>
            </div>
            <div class="form-group">
              <label for="namespace">命名空间</label>
              <input
                type="text"
                id="namespace"
                v-model="envForm.namespace"
                required
                class="form-input"
              />
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '提交中...' : '提交' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 编辑K8s环境模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>编辑K8s环境</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitEnvironment">
            <div class="form-group">
              <label for="edit-name">环境名称</label>
              <input
                type="text"
                id="edit-name"
                v-model="envForm.name"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="edit-description">描述</label>
              <textarea
                id="edit-description"
                v-model="envForm.description"
                rows="3"
                class="form-textarea"
              ></textarea>
            </div>
            <div class="form-group">
              <label for="edit-clusterName">集群名称</label>
              <input
                type="text"
                id="edit-clusterName"
                v-model="envForm.clusterName"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="edit-apiUrl">API URL</label>
              <input
                type="url"
                id="edit-apiUrl"
                v-model="envForm.apiUrl"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="edit-type">环境类型</label>
              <select
                id="edit-type"
                v-model="envForm.type"
                required
                class="form-input"
              >
                <option value="development">开发环境</option>
                <option value="testing">测试环境</option>
                <option value="staging">预发布环境</option>
                <option value="production">生产环境</option>
              </select>
            </div>
            <div class="form-group">
              <label for="edit-namespace">命名空间</label>
              <input
                type="text"
                id="edit-namespace"
                v-model="envForm.namespace"
                required
                class="form-input"
              />
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '提交中...' : '保存' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 查看K8s环境详情模态框 -->
    <div v-if="showViewModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>K8s环境详情</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-container">
            <div class="detail-section">
              <h4>基本信息</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">环境名称：</span>
                  <span class="detail-value">{{ viewEnv.name }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">描述：</span>
                  <span class="detail-value">{{ viewEnv.description }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">环境类型：</span>
                  <span class="detail-value">
                    <span class="type-tag" :class="`type-${viewEnv.type || viewEnv.envType}`">
                      {{ getEnvTypeName(viewEnv.type || viewEnv.envType) }}
                    </span>
                  </span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">状态：</span>
                  <span class="detail-value">
                    <span class="status-tag" :class="`status-${viewEnv.status}`">
                      <span class="status-icon" v-if="viewEnv.status === 'connected'">✓</span>
                      <span class="status-icon" v-else>✗</span>
                      {{ viewEnv.status }}
                    </span>
                  </span>
                </div>
              </div>
            </div>

            <div class="detail-section">
              <h4>集群信息</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">集群名称：</span>
                  <span class="detail-value">{{ viewEnv.clusterName }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">API URL：</span>
                  <span class="detail-value">{{ viewEnv.apiUrl }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">命名空间：</span>
                  <span class="detail-value">{{ viewEnv.namespace }}</span>
                </div>
              </div>
            </div>

            <div class="detail-section" v-if="viewEnv.certificateAuthority || viewEnv.clientCertificate">
              <h4>认证信息</h4>
              <div class="detail-grid">
                <div class="detail-item full-width">
                  <span class="detail-label">证书颁发机构：</span>
                  <div class="detail-value code-block">{{ viewEnv.certificateAuthority || '无' }}</div>
                </div>
                <div class="detail-item full-width">
                  <span class="detail-label">客户端证书：</span>
                  <div class="detail-value code-block">{{ viewEnv.clientCertificate || '无' }}</div>
                </div>
                <div class="detail-item full-width">
                  <span class="detail-label">客户端密钥：</span>
                  <div class="detail-value code-block">{{ viewEnv.clientKey ? '*** 已配置 ***' : '无' }}</div>
                </div>
              </div>
            </div>

            <div class="detail-section" v-if="viewEnv.createdAt || viewEnv.updatedAt">
              <h4>时间信息</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">创建时间：</span>
                  <span class="detail-value">{{ formatDate(viewEnv.createdAt) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">更新时间：</span>
                  <span class="detail-value">{{ formatDate(viewEnv.updatedAt) }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" @click="closeModal">关闭</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  getK8sEnvironments,
  deleteK8sEnvironment,
  createK8sEnvironment,
  updateK8sEnvironment,
  getK8sEnvironmentDetail
} from '@/api/cicd.js'
import Pagination from '@/components/Pagination.vue'

export default {
  name: 'K8sEnvironments',
  components: {
    Pagination
  },
  setup() {
    const router = useRouter()
    const environments = ref([])
    const searchQuery = ref('')
    const currentPage = ref(1)
    const pageSize = ref(10)

    // 模态框状态
    const showAddModal = ref(false)
    const showEditModal = ref(false)
    const showViewModal = ref(false)
    const submitting = ref(false)

    // 表单数据
    const envForm = ref({
      id: null,
      name: '',
      description: '',
      clusterName: '',
      apiUrl: '',
      namespace: 'default',
      type: 'development',
      status: 'disconnected'
    })

    // 查看详情数据
    const viewEnv = ref({
      id: null,
      name: '',
      description: '',
      type: 'development',
      status: 'disconnected',
      clusterName: '',
      apiUrl: '',
      namespace: 'default'
    })

    const loadEnvironments = async () => {
      try {
        const response = await getK8sEnvironments()
        if (response.code === 0) {
          environments.value = response.data
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取K8s环境列表失败')
      }
    }

    const filteredEnvironments = computed(() => {
      if (!searchQuery.value) {
        return environments.value
      }
      return environments.value.filter(env =>
        env.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        env.description.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        env.clusterName.toLowerCase().includes(searchQuery.value.toLowerCase())
      )
    })

    const paginatedEnvironments = computed(() => {
      const startIndex = (currentPage.value - 1) * pageSize.value
      const endIndex = startIndex + pageSize.value
      return filteredEnvironments.value.slice(startIndex, endIndex)
    })

    // 关闭模态框
    const getEnvTypeName = (type) => {
      const typeMap = {
        production: '生产环境',
        staging: '预发布环境',
        'pre-production': '预发布环境',
        preprod: '预发布环境',
        testing: '测试环境',
        development: '开发环境'
      }
      return typeMap[type] || type
    }

    const closeModal = () => {
      showAddModal.value = false
      showEditModal.value = false
      showViewModal.value = false
      envForm.value = {
        id: null,
        name: '',
        description: '',
        clusterName: '',
        apiUrl: '',
        namespace: 'default',
        type: 'development', // 默认环境类型
        status: 'disconnected'
      }
      viewEnv.value = {
        id: null,
        name: '',
        description: '',
        type: 'development',
        status: 'disconnected',
        clusterName: '',
        apiUrl: '',
        namespace: 'default'
      }
    }

    // 格式化日期
    const formatDate = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleString()
    }

    // 处理查看
    const handleView = async (env) => {
      try {
        const response = await getK8sEnvironmentDetail(env.id)
        if (response.code === 0) {
          viewEnv.value = response.data
          showViewModal.value = true
          showAddModal.value = false
          showEditModal.value = false
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取环境详情失败')
      }
    }

    // 处理编辑
    const handleEdit = (env) => {
      envForm.value = {
        ...env,
        type: env.type || env.envType || 'development' // 兼容两种环境类型字段名
      }
      showEditModal.value = true
      showAddModal.value = false
    }

    // 处理删除
    const handleDelete = (env) => {
      if (confirm(`确定要删除K8s环境 ${env.name} 吗？`)) {
        deleteK8sEnvironment(env.id)
          .then(response => {
            if (response.code === 0) {
              alert('删除K8s环境成功')
              loadEnvironments()
            } else {
              alert(response.msg)
            }
          })
          .catch(error => {
            alert('删除K8s环境失败')
          })
      }
    }

    // 提交环境表单
    const submitEnvironment = async () => {
      try {
        submitting.value = true
        let response

        if (showAddModal.value) {
          response = await createK8sEnvironment(envForm.value)
        } else if (showEditModal.value && envForm.value.id) {
          response = await updateK8sEnvironment(envForm.value.id, envForm.value)
        }

        if (response && response.code === 0) {
          alert(showAddModal.value ? '创建K8s环境成功' : '更新K8s环境成功')
          loadEnvironments()
          closeModal()
        } else if (response) {
          alert(response.msg)
        }
      } catch (error) {
        alert(showAddModal.value ? '创建K8s环境失败' : '更新K8s环境失败')
      } finally {
        submitting.value = false
      }
    }

    onMounted(() => {
      loadEnvironments()
    })

    return {
      environments,
      searchQuery,
      currentPage,
      pageSize,
      filteredEnvironments,
      paginatedEnvironments,
      showAddModal,
      showEditModal,
      showViewModal,
      envForm,
      viewEnv,
      submitting,
      handleView,
      handleEdit,
      handleDelete,
      submitEnvironment,
      closeModal,
      getEnvTypeName
    }
  }
}
</script>

<style scoped>
.k8s-environments {
  padding: 20px;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  gap: 20px;
}

.search-input {
  padding: 8px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  width: 300px;
  font-size: 14px;
}

.search-input:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  background-color: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.data-table th {
  background-color: #f7fafc;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.data-table tr:hover {
  background-color: #f7fafc;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-connected {
  background-color: #d4edda;
  color: #155724;
}

.status-disconnected {
  background-color: #f8d7da;
  color: #721c24;
}

.status-icon {
  font-weight: bold;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 6px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary {
  background-color: #326ce5;
  color: white;
  border-color: #326ce5;
}

.btn-primary:hover {
  background-color: #2554c7;
}

.btn-danger {
  background-color: #e53e3e;
  color: white;
  border-color: #e53e3e;
}

.btn-danger:hover {
  background-color: #c53030;
}

.btn-view {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.btn-view:hover {
  background-color: #5a6268;
}

.btn-edit {
  background-color: #ffc107;
  color: #212529;
  border-color: #ffc107;
}

.btn-edit:hover {
  background-color: #e0a800;
}
/* 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal {
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 500px;
  overflow: hidden;
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: white;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: rgba(255, 255, 255, 0.15);
  transform: rotate(90deg);
}

.modal-body {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.form-input, .form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s ease;
  background-color: white;
}

.form-input:focus, .form-textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
  font-family: inherit;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  background-color: #f8fafc;
  border-top: 1px solid #e2e8f0;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 100px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #e2e8f0;
  color: #4a5568;
}

.btn-secondary:hover {
  background-color: #cbd5e0;
  transform: translateY(-1px);
}

/* 环境类型标签样式 */
.type-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
  text-transform: capitalize;
}

.type-production {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.type-staging {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.type-pre-production {
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}

.type-testing {
  background-color: #d1ecf1;
  color: #0c5460;
  border: 1px solid #bee5eb;
}

.type-development {
  background-color: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

/* 状态标签样式增强 */
.status-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.status-connected {
  background-color: #d4edda;
  color: #155724;
}

.status-disconnected {
  background-color: #f8d7da;
  color: #721c24;
}

/* 表格样式增强 */
.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  overflow: hidden;
}

.data-table th {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  padding: 16px;
  text-align: left;
}

.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.data-table tr:hover {
  background-color: #f8fafc;
}

/* 工具栏样式增强 */
.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  gap: 16px;
  padding: 16px 20px;
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.search-input {
  flex: 1;
  max-width: 350px;
  padding: 10px 16px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

/* 按钮组样式 */
.action-buttons {
  display: flex;
  gap: 8px;
}

.btn-view, .btn-edit, .btn-danger {
  padding: 6px 12px;
  font-size: 13px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.btn-view {
  background-color: #17a2b8;
  color: white;
}

.btn-view:hover {
  background-color: #138496;
  transform: translateY(-1px);
}

.btn-edit {
  background-color: #ffc107;
  color: #212529;
}

.btn-edit:hover {
  background-color: #e0a800;
  transform: translateY(-1px);
}

.btn-danger {
  background-color: #dc3545;
  color: white;
}

.btn-danger:hover {
  background-color: #c82333;
  transform: translateY(-1px);
}

/* 标题样式 */
h2 {
  margin: 0 0 24px 0;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 主容器样式 */
.k8s-environments {
  padding: 24px;
  background-color: var(--bg-color);
  min-height: 100vh;
}

/* 分页样式增强 */
.pagination-section {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  padding: 16px;
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}
</style>

<template>
  <div class="image-repositories">
    <h2>镜像仓库管理</h2>

    <div class="toolbar">
      <button class="btn btn-primary" @click="showAddModal = true">添加镜像仓库</button>
      <input
        v-model="searchQuery"
        placeholder="搜索镜像仓库"
        class="search-input"
      />
    </div>

    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>仓库名称</th>
          <th>类型</th>
          <th>URL</th>
          <th>状态</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="repo in paginatedRepositories" :key="repo.id">
          <td>{{ repo.id }}</td>
          <td>{{ repo.name }}</td>
          <td>
            <span :class="['type-tag', `type-${repo.type}`]">{{ repo.type }}</span>
          </td>
          <td>{{ repo.url }}</td>
          <td>
            <span :class="['status-tag', `status-${repo.status}`]">
              <span class="status-icon" v-if="repo.status === 'connected'">✓</span>
              <span class="status-icon" v-else>✗</span>
              {{ repo.status }}
            </span>
          </td>
          <td>
            <div class="action-buttons">
              <button class="btn btn-view" @click="viewRepository(repo)">查看</button>
              <button class="btn btn-edit" @click="editRepository(repo)">编辑</button>
              <button class="btn btn-danger" @click="deleteRepository(repo.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <Pagination
      :current-page="currentPage"
      :total-items="filteredRepositories.length"
      :items-per-page="pageSize"
      @update:currentPage="currentPage = $event"
    />

    <!-- 添加镜像仓库模态框 -->
    <div v-if="showAddModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>添加镜像仓库</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitForm">
            <div class="form-group">
              <label for="name">仓库名称</label>
              <input
                type="text"
                id="name"
                v-model="formData.name"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="type">类型</label>
              <select
                id="type"
                v-model="formData.type"
                required
                class="form-select"
              >
                <option value="docker">Docker</option>
                <option value="harbor">Harbor</option>
                <option value="gcr">GCR</option>
                <option value="ecr">ECR</option>
                <option value="acr">ACR</option>
              </select>
            </div>
            <div class="form-group">
              <label for="url">URL</label>
              <input
                type="url"
                id="url"
                v-model="formData.url"
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

    <!-- 编辑镜像仓库模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>编辑镜像仓库</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitForm">
            <div class="form-group">
              <label for="edit-name">仓库名称</label>
              <input
                type="text"
                id="edit-name"
                v-model="formData.name"
                required
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label for="edit-type">类型</label>
              <select
                id="edit-type"
                v-model="formData.type"
                required
                class="form-select"
              >
                <option value="docker">Docker</option>
                <option value="harbor">Harbor</option>
                <option value="gcr">GCR</option>
                <option value="ecr">ECR</option>
                <option value="acr">ACR</option>
              </select>
            </div>
            <div class="form-group">
              <label for="edit-url">URL</label>
              <input
                type="url"
                id="edit-url"
                v-model="formData.url"
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

    <!-- 查看镜像仓库详情模态框 -->
    <div v-if="showViewModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>镜像仓库详情</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-container">
            <div class="detail-section">
              <h4>基本信息</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">仓库名称：</span>
                  <span class="detail-value">{{ viewRepo.name }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">类型：</span>
                  <span class="detail-value">
                    <span class="type-tag" :class="`type-${viewRepo.type}`">
                      {{ viewRepo.type }}
                    </span>
                  </span>
                </div>
                <div class="detail-item full-width">
                  <span class="detail-label">URL：</span>
                  <span class="detail-value">{{ viewRepo.url }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">状态：</span>
                  <span class="detail-value">
                    <span class="status-tag" :class="`status-${viewRepo.status}`">
                      <span class="status-icon" v-if="viewRepo.status === 'connected'">✓</span>
                      <span class="status-icon" v-else>✗</span>
                      {{ viewRepo.status }}
                    </span>
                  </span>
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
  getImageRepositories,
  createImageRepository,
  updateImageRepository,
  deleteImageRepository
} from '@/api/cicd.js'
import Pagination from '@/components/Pagination.vue'

export default {
  name: 'ImageRepositories',
  components: {
    Pagination
  },
  setup() {
    const router = useRouter()
    const repositories = ref([])
    const searchQuery = ref('')
    const currentPage = ref(1)
    const pageSize = ref(10)

    // 模态框状态
    const showAddModal = ref(false)
    const showEditModal = ref(false)
    const submitting = ref(false)

    // 表单数据
    const formData = ref({
      id: null,
      name: '',
      type: 'docker',
      url: ''
    })

    const loadRepositories = async () => {
      try {
        const response = await getImageRepositories()
        if (response.code === 0) {
          repositories.value = response.data
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('获取镜像仓库列表失败')
      }
    }

    const filteredRepositories = computed(() => {
      if (!searchQuery.value) {
        return repositories.value
      }
      return repositories.value.filter(repo =>
        repo.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        repo.url.toLowerCase().includes(searchQuery.value.toLowerCase())
      )
    })

    const paginatedRepositories = computed(() => {
      const startIndex = (currentPage.value - 1) * pageSize.value
      const endIndex = startIndex + pageSize.value
      return filteredRepositories.value.slice(startIndex, endIndex)
    })

    // 打开添加模态框
    const addRepository = () => {
      formData.value = {
        id: null,
        name: '',
        type: 'docker',
        url: ''
      }
      showAddModal.value = true
      showEditModal.value = false
    }

    // 打开编辑模态框
    const editRepository = (repo) => {
      formData.value = { ...repo }
      showEditModal.value = true
      showAddModal.value = false
    }

    // 关闭模态框
    const closeModal = () => {
      showAddModal.value = false
      showEditModal.value = false
      showViewModal.value = false
      formData.value = {
        id: null,
        name: '',
        type: 'docker',
        url: ''
      }
      viewRepo.value = {}
    }

    // 提交表单
    const submitForm = async () => {
      try {
        submitting.value = true
        let response

        if (formData.value.id) {
          // 更新操作
          response = await updateImageRepository(formData.value.id, formData.value)
        } else {
          // 添加操作
          response = await createImageRepository(formData.value)
        }

        if (response.code === 0) {
          alert(response.msg)
          loadRepositories()
          closeModal()
        } else {
          alert(response.msg)
        }
      } catch (error) {
        alert('操作失败')
      } finally {
        submitting.value = false
      }
    }

    // 查看镜像
    const viewImages = (id) => {
      router.push(`/images/${id}`)
    }

    // 查看镜像仓库详情
    const showViewModal = ref(false)
    const viewRepo = ref({})

    const viewRepository = (repo) => {
      // 模拟获取详情，实际项目中应该调用API
      viewRepo.value = {...repo}
      showViewModal.value = true
      showAddModal.value = false
      showEditModal.value = false
    }

    // 删除镜像仓库
    const deleteRepository = async (id) => {
      if (confirm('确定要删除这个镜像仓库吗？')) {
        try {
          const response = await deleteImageRepository(id)
          if (response.code === 0) {
            alert('删除成功')
            loadRepositories()
          } else {
            alert(response.msg)
          }
        } catch (error) {
          alert('删除失败')
        }
      }
    }

    onMounted(() => {
      loadRepositories()
    })

    return {
      repositories,
      searchQuery,
      currentPage,
      pageSize,
      filteredRepositories,
      paginatedRepositories,
      showAddModal,
      showEditModal,
      showViewModal,
      formData,
      submitting,
      viewRepo,
      addRepository,
      editRepository,
      closeModal,
      submitForm,
      viewImages,
      viewRepository,
      deleteRepository
    }
  }
}
</script>

<style scoped>
.image-repositories {
  padding: 20px;
  background-color: var(--bg-color);
  min-height: 100vh;
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

.type-tag {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.type-docker {
  background-color: #d4edda;
  color: #155724;
}

.type-harbor {
  background-color: #fff3cd;
  color: #856404;
}

.type-gcr {
  background-color: #d1ecf1;
  color: #0c5460;
}

.type-ecr {
  background-color: #cce5ff;
  color: #004085;
}

.type-acr {
  background-color: #f8d7da;
  color: #721c24;
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

.btn-danger {
  background-color: #e53e3e;
  color: white;
  border-color: #e53e3e;
}

.btn-danger:hover {
  background-color: #c53030;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  border-color: #6c757d;
}

.btn-secondary:hover {
  background-color: #5a6268;
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
}

.modal {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 500px;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #f7fafc;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #4a5568;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #718096;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background-color: #edf2f7;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #4a5568;
  font-size: 14px;
}

.form-input,
.form-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #326ce5;
  box-shadow: 0 0 0 3px rgba(50, 108, 229, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e2e8f0;
  background-color: #f7fafc;
}

.modal-footer .btn {
  padding: 8px 16px;
  font-size: 14px;
}

/* 详情模态框样式 */
.detail-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-section {
  background-color: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  color: #4a5568;
  font-size: 16px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-weight: 600;
  color: #718096;
  font-size: 13px;
}

.detail-value {
  color: #2d3748;
  font-size: 14px;
}

/* 优化表格样式 */
.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  overflow: hidden;
}

.data-table th {
  background-color: #326ce5;
  color: white;
  font-weight: 600;
  font-size: 14px;
  padding: 16px;
  text-align: left;
  border-bottom: none;
}

.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #2d3748;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.data-table tr:hover {
  background-color: #f8fafc;
}

/* 优化工具栏样式 */
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

/* 优化按钮样式 */
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 500;
  min-width: 80px;
}

.btn-view {
  background-color: #17a2b8;
  color: white;
}

.btn-view:hover {
  background-color: #138496;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(23, 162, 184, 0.3);
}

.btn-edit {
  background-color: #ffc107;
  color: #212529;
}

.btn-edit:hover {
  background-color: #e0a800;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 193, 7, 0.3);
}

.btn-danger {
  background-color: #dc3545;
  color: white;
}

.btn-danger:hover {
  background-color: #c82333;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(220, 53, 69, 0.3);
}

.btn-primary {
  background-color: #326ce5;
  color: white;
}

.btn-primary:hover {
  background-color: #2554c7;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(50, 108, 229, 0.3);
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background-color: #5a6268;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(108, 117, 125, 0.3);
}
</style>

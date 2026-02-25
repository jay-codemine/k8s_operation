<template>
  <div class="resource-list">
    <h2>CustomResourceDefinition 管理</h2>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-filter">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 CustomResourceDefinition..."
          class="search-input"
        />
        <select v-model="scopeFilter" class="filter-select">
          <option value="">所有作用域</option>
          <option value="Namespaced">Namespaced</option>
          <option value="Cluster">Cluster</option>
        </select>
      </div>
      <button @click="showCreateModal = true" class="create-btn">创建 CRD</button>
    </div>

    <!-- CustomResourceDefinition 列表 -->
    <div class="resource-table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>组</th>
            <th>版本</th>
            <th>Kind</th>
            <th>作用域</th>
            <th>状态</th>
            <th>描述</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="crd in paginatedCRDs" :key="crd.name">
            <td class="resource-name">{{ crd.name }}</td>
            <td>{{ crd.group }}</td>
            <td>{{ crd.version }}</td>
            <td>{{ crd.kind }}</td>
            <td>
              <span :class="['scope-indicator', crd.scope]">
                {{ crd.scope }}
              </span>
            </td>
            <td>
              <span :class="['status-indicator', crd.status]">
                {{ crd.status }}
              </span>
            </td>
            <td class="description">{{ crd.description || '-' }}</td>
            <td>{{ crd.createdAt }}</td>
            <td class="action-buttons">
              <button @click="editCRD(crd)" class="edit-btn">编辑</button>
              <button @click="confirmDelete(crd)" class="delete-btn">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页组件 -->
    <Pagination v-if="filteredCRDs.length > 0" v-model:currentPage="currentPage" :totalItems="filteredCRDs.length" :itemsPerPage="itemsPerPage" />

    <!-- 创建 CRD 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>创建 CustomResourceDefinition</h3>
          <button @click="showCreateModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>名称</label>
            <input type="text" v-model="crdForm.name" class="form-input" placeholder="example-resources.example.com" required>
          </div>
          <div class="form-group">
            <label>组</label>
            <input type="text" v-model="crdForm.group" class="form-input" placeholder="example.com" required>
          </div>
          <div class="form-group">
            <label>版本</label>
            <input type="text" v-model="crdForm.version" class="form-input" placeholder="v1alpha1" required>
          </div>
          <div class="form-group">
            <label>Kind</label>
            <input type="text" v-model="crdForm.kind" class="form-input" placeholder="ExampleResource" required>
          </div>
          <div class="form-group">
            <label>作用域</label>
            <select v-model="crdForm.scope" class="form-select">
              <option value="Namespaced">Namespaced</option>
              <option value="Cluster">Cluster</option>
            </select>
          </div>
          <div class="form-group">
            <label>状态</label>
            <select v-model="crdForm.status" class="form-select">
              <option value="Established">Established</option>
              <option value="NamesAccepted">NamesAccepted</option>
              <option value="NonStructuralSchema">NonStructuralSchema</option>
            </select>
          </div>
          <div class="form-group">
            <label>描述</label>
            <textarea v-model="crdForm.description" class="form-textarea" placeholder="CRD 描述..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCreateModal = false" class="cancel-btn">取消</button>
          <button @click="createCRD" class="submit-btn">创建</button>
        </div>
      </div>
    </div>

    <!-- 编辑 CRD 模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click="showEditModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>编辑 CustomResourceDefinition</h3>
          <button @click="showEditModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>名称</label>
            <input type="text" v-model="editForm.name" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>组</label>
            <input type="text" v-model="editForm.group" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>版本</label>
            <input type="text" v-model="editForm.version" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>Kind</label>
            <input type="text" v-model="editForm.kind" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>作用域</label>
            <input type="text" v-model="editForm.scope" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>状态</label>
            <select v-model="editForm.status" class="form-select">
              <option value="Established">Established</option>
              <option value="NamesAccepted">NamesAccepted</option>
              <option value="NonStructuralSchema">NonStructuralSchema</option>
            </select>
          </div>
          <div class="form-group">
            <label>描述</label>
            <textarea v-model="editForm.description" class="form-textarea" placeholder="CRD 描述..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showEditModal = false" class="cancel-btn">取消</button>
          <button @click="updateCRD" class="submit-btn">更新</button>
        </div>
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal-content delete-confirm" @click.stop>
        <div class="modal-header">
          <h3>确认删除</h3>
          <button @click="showDeleteModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <p>确定要删除 CustomResourceDefinition <strong>{{ selectedCRD.name }}</strong> 吗？</p>
          <p class="warning-text">此操作不可恢复，并且会删除所有相关的自定义资源实例。</p>
        </div>
        <div class="modal-footer">
          <button @click="showDeleteModal = false" class="cancel-btn">取消</button>
          <button @click="deleteCRD" class="delete-btn">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import Pagination from '@/components/Pagination.vue'

// 搜索和过滤
const searchQuery = ref('')
const scopeFilter = ref('')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)

// 模态框状态
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)

// 选中的 CRD
const selectedCRD = ref(null)

// 创建表单
const crdForm = ref({
  name: '',
  group: '',
  version: 'v1alpha1',
  kind: '',
  scope: 'Namespaced',
  status: 'Established',
  description: ''
})

// 编辑表单
const editForm = ref({
  name: '',
  group: '',
  version: '',
  kind: '',
  scope: '',
  status: '',
  description: ''
})

// 模拟 CRD 数据
const crds = ref([
  {
    name: 'cronjobs.batch.training',
    group: 'batch.training',
    version: 'v1alpha1',
    kind: 'CronJob',
    scope: 'Namespaced',
    status: 'Established',
    description: 'A custom CronJob implementation',
    createdAt: '2025-12-01'
  },
  {
    name: 'appservices.apps.training',
    group: 'apps.training',
    version: 'v1',
    kind: 'AppService',
    scope: 'Namespaced',
    status: 'Established',
    description: 'Application service resource',
    createdAt: '2025-12-02'
  },
  {
    name: 'clusters.infra.training',
    group: 'infra.training',
    version: 'v1beta1',
    kind: 'Cluster',
    scope: 'Cluster',
    status: 'Established',
    description: 'Cluster management resource',
    createdAt: '2025-12-03'
  },
  {
    name: 'databases.db.training',
    group: 'db.training',
    version: 'v1alpha1',
    kind: 'Database',
    scope: 'Namespaced',
    status: 'NamesAccepted',
    description: 'Database resource',
    createdAt: '2025-12-04'
  },
  {
    name: 'networks.net.training',
    group: 'net.training',
    version: 'v1',
    kind: 'Network',
    scope: 'Cluster',
    status: 'Established',
    description: 'Network configuration resource',
    createdAt: '2025-12-05'
  },
  {
    name: 'storages.storage.training',
    group: 'storage.training',
    version: 'v1beta1',
    kind: 'Storage',
    scope: 'Namespaced',
    status: 'NonStructuralSchema',
    description: 'Storage resource',
    createdAt: '2025-12-06'
  },
  {
    name: 'monitors.monitoring.training',
    group: 'monitoring.training',
    version: 'v1alpha1',
    kind: 'Monitor',
    scope: 'Namespaced',
    status: 'Established',
    description: 'Monitoring configuration resource',
    createdAt: '2025-12-07'
  },
  {
    name: 'secrets.sec.training',
    group: 'sec.training',
    version: 'v1',
    kind: 'SecretStore',
    scope: 'Cluster',
    status: 'NamesAccepted',
    description: 'Secret management resource',
    createdAt: '2025-12-08'
  },
  {
    name: 'policies.policy.training',
    group: 'policy.training',
    version: 'v1beta1',
    kind: 'Policy',
    scope: 'Cluster',
    status: 'Established',
    description: 'Policy enforcement resource',
    createdAt: '2025-12-09'
  }
])

// 过滤后的 CRD 列表
const filteredCRDs = computed(() => {
  return crds.value.filter(crd => {
    const matchesSearch = crd.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                         crd.kind.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesScope = !scopeFilter.value || crd.scope === scopeFilter.value
    return matchesSearch && matchesScope
  })
})

// 分页后的 CRD 列表
const paginatedCRDs = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredCRDs.value.slice(startIndex, endIndex)
})

// 创建 CRD
const createCRD = () => {
  const newCRD = {
    ...crdForm.value,
    createdAt: new Date().toISOString().split('T')[0]
  }

  crds.value.push(newCRD)
  showCreateModal.value = false

  // 重置表单
  crdForm.value = {
    name: '',
    group: '',
    version: 'v1alpha1',
    kind: '',
    scope: 'Namespaced',
    status: 'Established',
    description: ''
  }
}

// 编辑 CRD
const editCRD = (crd) => {
  selectedCRD.value = crd
  editForm.value = {
    ...crd
  }
  showEditModal.value = true
}

// 更新 CRD
const updateCRD = () => {
  const index = crds.value.findIndex(crd => crd.name === editForm.value.name)
  if (index !== -1) {
    crds.value[index] = {
      ...crds.value[index],
      status: editForm.value.status,
      description: editForm.value.description
    }
  }
  showEditModal.value = false
}

// 确认删除
const confirmDelete = (crd) => {
  selectedCRD.value = crd
  showDeleteModal.value = true
}

// 删除 CRD
const deleteCRD = () => {
  const index = crds.value.findIndex(crd => crd.name === selectedCRD.value.name)
  if (index !== -1) {
    crds.value.splice(index, 1)
  }
  showDeleteModal.value = false
  selectedCRD.value = null
}
</script>

<style scoped>
.resource-list {
  padding: 20px;
}

h2 {
  margin-bottom: 20px;
  color: #333;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.search-filter {
  display: flex;
  gap: 10px;
}

.search-input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 300px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: white;
}

.create-btn {
  padding: 8px 16px;
  background-color: #326ce5;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.create-btn:hover {
  background-color: #2554c7;
}

.resource-table-container {
  overflow-x: auto;
}

.resource-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.resource-table th,
.resource-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.resource-table th {
  background-color: #f5f5f5;
  font-weight: 600;
  color: #555;
}

.status-indicator,
.scope-indicator {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-indicator.Established {
  background-color: #d4edda;
  color: #155724;
}

.status-indicator.NamesAccepted {
  background-color: #d1ecf1;
  color: #0c5460;
}

.status-indicator.NonStructuralSchema {
  background-color: #fff3cd;
  color: #856404;
}

.scope-indicator.Namespaced {
  background-color: #e7f3ff;
  color: #0066cc;
}

.scope-indicator.Cluster {
  background-color: #f0e6ff;
  color: #6600cc;
}

.resource-name {
  font-weight: 600;
}

.description {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.edit-btn,
.delete-btn {
  padding: 4px 8px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: background-color 0.3s;
}

.edit-btn {
  background-color: #ffc107;
  color: #333;
}

.edit-btn:hover {
  background-color: #e0a800;
}

.delete-btn {
  background-color: #dc3545;
  color: white;
}

.delete-btn:hover {
  background-color: #c82333;
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
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  margin: 0;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #999;
}

.close-btn:hover {
  color: #333;
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
  font-weight: 500;
  color: #555;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-textarea {
  height: 80px;
  resize: vertical;
}

.form-input:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid #eee;
}

.submit-btn,
.cancel-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.3s;
}

.submit-btn {
  background-color: #326ce5;
  color: white;
}

.submit-btn:hover {
  background-color: #2554c7;
}

.cancel-btn {
  background-color: #6c757d;
  color: white;
}

.cancel-btn:hover {
  background-color: #5a6268;
}

.delete-confirm .modal-body {
  text-align: center;
}

.warning-text {
  color: #dc3545;
  font-weight: 500;
}
</style>

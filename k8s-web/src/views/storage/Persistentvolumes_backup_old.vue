<template>
  <div class="resource-list">
    <h2>PersistentVolume 管理</h2>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-filter">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="搜索 PersistentVolume..."
          class="search-input"
        />
        <select v-model="statusFilter" class="filter-select">
          <option value="">所有状态</option>
          <option value="Available">Available</option>
          <option value="Bound">Bound</option>
          <option value="Released">Released</option>
          <option value="Failed">Failed</option>
        </select>
      </div>
      <button @click="showCreateModal = true" class="create-btn">创建 PV</button>
    </div>

    <!-- PersistentVolume 列表 -->
    <div class="resource-table-container">
      <table class="resource-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>状态</th>
            <th>容量</th>
            <th>访问模式</th>
            <th>回收策略</th>
            <th>存储类型</th>
            <th>挂载路径</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pv in paginatedPVs" :key="pv.name">
            <td class="resource-name">{{ pv.name }}</td>
            <td>
              <span :class="['status-indicator', pv.status]">
                {{ pv.status }}
              </span>
            </td>
            <td>{{ pv.capacity }}</td>
            <td>{{ pv.accessModes }}</td>
            <td>{{ pv.reclaimPolicy }}</td>
            <td>{{ pv.storageClassName || '-' }}</td>
            <td>{{ pv.mountPath }}</td>
            <td>{{ pv.createdAt }}</td>
            <td class="action-buttons">
              <button @click="editPV(pv)" class="edit-btn">编辑</button>
              <button @click="confirmDelete(pv)" class="delete-btn">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页组件 -->
    <Pagination v-if="filteredPVs.length > 0" v-model:currentPage="currentPage" :totalItems="filteredPVs.length" :itemsPerPage="itemsPerPage" />

    <!-- 创建 PV 模态框 -->
    <div v-if="showCreateModal" class="modal-overlay" @click="showCreateModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>创建 PersistentVolume</h3>
          <button @click="showCreateModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>名称</label>
            <input type="text" v-model="pvForm.name" class="form-input" required>
          </div>
          <div class="form-group">
            <label>容量</label>
            <input type="text" v-model="pvForm.capacity" class="form-input" placeholder="1Gi" required>
          </div>
          <div class="form-group">
            <label>访问模式</label>
            <select v-model="pvForm.accessModes" class="form-select">
              <option value="ReadWriteOnce">ReadWriteOnce</option>
              <option value="ReadOnlyMany">ReadOnlyMany</option>
              <option value="ReadWriteMany">ReadWriteMany</option>
            </select>
          </div>
          <div class="form-group">
            <label>回收策略</label>
            <select v-model="pvForm.reclaimPolicy" class="form-select">
              <option value="Delete">Delete</option>
              <option value="Retain">Retain</option>
              <option value="Recycle">Recycle</option>
            </select>
          </div>
          <div class="form-group">
            <label>存储类型</label>
            <input type="text" v-model="pvForm.storageClassName" class="form-input" placeholder="standard">
          </div>
          <div class="form-group">
            <label>挂载路径</label>
            <input type="text" v-model="pvForm.mountPath" class="form-input" placeholder="/mnt/data" required>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCreateModal = false" class="cancel-btn">取消</button>
          <button @click="createPV" class="submit-btn">创建</button>
        </div>
      </div>
    </div>

    <!-- 编辑 PV 模态框 -->
    <div v-if="showEditModal" class="modal-overlay" @click="showEditModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>编辑 PersistentVolume</h3>
          <button @click="showEditModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>名称</label>
            <input type="text" v-model="editForm.name" class="form-input" disabled>
          </div>
          <div class="form-group">
            <label>状态</label>
            <select v-model="editForm.status" class="form-select">
              <option value="Available">Available</option>
              <option value="Bound">Bound</option>
              <option value="Released">Released</option>
              <option value="Failed">Failed</option>
            </select>
          </div>
          <div class="form-group">
            <label>容量</label>
            <input type="text" v-model="editForm.capacity" class="form-input" required>
          </div>
          <div class="form-group">
            <label>访问模式</label>
            <select v-model="editForm.accessModes" class="form-select">
              <option value="ReadWriteOnce">ReadWriteOnce</option>
              <option value="ReadOnlyMany">ReadOnlyMany</option>
              <option value="ReadWriteMany">ReadWriteMany</option>
            </select>
          </div>
          <div class="form-group">
            <label>回收策略</label>
            <select v-model="editForm.reclaimPolicy" class="form-select">
              <option value="Delete">Delete</option>
              <option value="Retain">Retain</option>
              <option value="Recycle">Recycle</option>
            </select>
          </div>
          <div class="form-group">
            <label>存储类型</label>
            <input type="text" v-model="editForm.storageClassName" class="form-input" placeholder="standard">
          </div>
          <div class="form-group">
            <label>挂载路径</label>
            <input type="text" v-model="editForm.mountPath" class="form-input" required>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showEditModal = false" class="cancel-btn">取消</button>
          <button @click="updatePV" class="submit-btn">更新</button>
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
          <p>确定要删除 PersistentVolume <strong>{{ selectedPV.name }}</strong> 吗？</p>
          <p class="warning-text">此操作不可恢复。</p>
        </div>
        <div class="modal-footer">
          <button @click="showDeleteModal = false" class="cancel-btn">取消</button>
          <button @click="deletePV" class="delete-btn">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import Pagination from '@/components/Pagination.vue'
import pvApi from '@/api/cluster/storage/pv'

// =============== 状态管理 ===============
const pvs = ref([])
const loading = ref(false)
const errorMsg = ref('')

// 搜索和过滤
const searchQuery = ref('')
const debouncedSearchQuery = ref('')
const statusFilter = ref('')
const storageClassFilter = ref('')

// 分页
const currentPage = ref(1)
const itemsPerPage = ref(10)
const jumpPage = ref(1)

// 批量操作
const batchMode = ref(false)
const selectedPVs = ref([])

// 自动刷新
const autoRefresh = ref(false)
const refreshTimer = ref(null)

// 模态框状态
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const showYamlModal = ref(false)
const showDetailModal = ref(false)
const creating = ref(false)
const yamlSaving = ref(false)

// 创建相关
const createMode = ref('form')
const createYamlContent = ref('')
const createYamlError = ref('')

// 选中的 PV
const selectedPV = ref(null)

// 创建表单
const pvForm = ref({
  name: '',
  capacity: '1Gi',
  accessModes: ['ReadWriteOnce'],
  reclaimPolicy: 'Retain',
  storageClassName: '',
  volumeSource: {
    type: 'hostPath',
    hostPath: '/mnt/data',
    nfsServer: '',
    nfsPath: '/'
  }
})

// 编辑表单
const editForm = ref({
  name: '',
  capacity: '',
  accessModes: [],
  reclaimPolicy: '',
  storageClassName: ''
})

// YAML 查看/编辑
const currentYaml = ref('')
const editedYaml = ref('')
const yamlMode = ref('view')
const yamlError = ref('')

// 模拟 PV 数据
const pvs = ref([
  {
    name: 'pv-001',
    status: 'Available',
    capacity: '5Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Delete',
    storageClassName: 'standard',
    mountPath: '/mnt/data/pv001',
    createdAt: '2025-12-01'
  },
  {
    name: 'pv-002',
    status: 'Bound',
    capacity: '10Gi',
    accessModes: 'ReadWriteMany',
    reclaimPolicy: 'Retain',
    storageClassName: 'gp3',
    mountPath: '/mnt/data/pv002',
    createdAt: '2025-12-02'
  },
  {
    name: 'pv-003',
    status: 'Available',
    capacity: '20Gi',
    accessModes: 'ReadOnlyMany',
    reclaimPolicy: 'Delete',
    storageClassName: 'io1',
    mountPath: '/mnt/data/pv003',
    createdAt: '2025-12-03'
  },
  {
    name: 'pv-004',
    status: 'Released',
    capacity: '1Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Retain',
    storageClassName: '',
    mountPath: '/mnt/data/pv004',
    createdAt: '2025-12-04'
  },
  {
    name: 'pv-005',
    status: 'Bound',
    capacity: '5Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Delete',
    storageClassName: 'local-storage',
    mountPath: '/mnt/data/pv005',
    createdAt: '2025-12-05'
  },
  {
    name: 'pv-006',
    status: 'Failed',
    capacity: '15Gi',
    accessModes: 'ReadWriteMany',
    reclaimPolicy: 'Delete',
    storageClassName: 'managed-nfs-storage',
    mountPath: '/mnt/data/pv006',
    createdAt: '2025-12-06'
  },
  {
    name: 'pv-007',
    status: 'Available',
    capacity: '30Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Retain',
    storageClassName: 'azure-disk',
    mountPath: '/mnt/data/pv007',
    createdAt: '2025-12-07'
  },
  {
    name: 'pv-008',
    status: 'Bound',
    capacity: '8Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Delete',
    storageClassName: 'gce-pd',
    mountPath: '/mnt/data/pv008',
    createdAt: '2025-12-08'
  },
  {
    name: 'pv-009',
    status: 'Available',
    capacity: '2Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Delete',
    storageClassName: 'portworx-volume',
    mountPath: '/mnt/data/pv009',
    createdAt: '2025-12-09'
  }
])

// 过滤后的 PV 列表
const filteredPVs = computed(() => {
  return pvs.value.filter(pv => {
    const matchesSearch = pv.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus = !statusFilter.value || pv.status === statusFilter.value
    return matchesSearch && matchesStatus
  })
})

// 分页后的 PV 列表
const paginatedPVs = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return filteredPVs.value.slice(startIndex, endIndex)
})

// 创建 PV
const createPV = () => {
  const newPV = {
    ...pvForm.value,
    createdAt: new Date().toISOString().split('T')[0]
  }

  pvs.value.push(newPV)
  showCreateModal.value = false

  // 重置表单
  pvForm.value = {
    name: '',
    status: 'Available',
    capacity: '1Gi',
    accessModes: 'ReadWriteOnce',
    reclaimPolicy: 'Delete',
    storageClassName: '',
    mountPath: ''
  }
}

// 编辑 PV
const editPV = (pv) => {
  selectedPV.value = pv
  editForm.value = {
    ...pv
  }
  showEditModal.value = true
}

// 更新 PV
const updatePV = () => {
  const index = pvs.value.findIndex(pv => pv.name === editForm.value.name)
  if (index !== -1) {
    pvs.value[index] = {
      ...pvs.value[index],
      ...editForm.value
    }
  }
  showEditModal.value = false
}

// 确认删除
const confirmDelete = (pv) => {
  selectedPV.value = pv
  showDeleteModal.value = true
}

// 删除 PV
const deletePV = () => {
  const index = pvs.value.findIndex(pv => pv.name === selectedPV.value.name)
  if (index !== -1) {
    pvs.value.splice(index, 1)
  }
  showDeleteModal.value = false
  selectedPV.value = null
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

.status-indicator {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-indicator.Available {
  background-color: #d4edda;
  color: #155724;
}

.status-indicator.Bound {
  background-color: #d1ecf1;
  color: #0c5460;
}

.status-indicator.Released {
  background-color: #fff3cd;
  color: #856404;
}

.status-indicator.Failed {
  background-color: #f8d7da;
  color: #721c24;
}

.resource-name {
  font-weight: 600;
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
.form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
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

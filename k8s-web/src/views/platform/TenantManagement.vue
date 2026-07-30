<template>
  <div class="tenant-page">
    <div class="page-header">
      <h2>租户管理</h2>
      <span class="subtitle">管理平台多租户组织，创建、编辑、启停租户</span>
    </div>

    <div class="toolbar">
      <a-button type="primary" @click="openCreate" v-permission="['super_admin']">
        <template #icon><icon-plus /></template>
        新建租户
      </a-button>
      <span class="hint" v-if="!isSuperAdmin">仅超级管理员可管理租户</span>
    </div>

    <a-table :data="tenants" :loading="loading" :pagination="false" column-resizable>
      <template #columns>
        <a-table-column title="ID" data-index="id" :width="80" />
        <a-table-column title="租户名称" data-index="name" />
        <a-table-column title="编码" data-index="code" />
        <a-table-column title="状态" data-index="status" :width="100">
          <template #cell="{ record }">
            <a-tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '启用' : '禁用' }}
            </a-tag>
          </template>
        </a-table-column>
        <a-table-column title="创建时间" :width="180">
          <template #cell="{ record }">
            {{ formatTime(record.created_at) }}
          </template>
        </a-table-column>
        <a-table-column title="操作" :width="200" v-if="isSuperAdmin">
          <template #cell="{ record }">
            <a-space>
              <a-button size="small" @click="openEdit(record)">编辑</a-button>
              <a-popconfirm content="确定删除此租户？" @ok="handleDelete(record)">
                <a-button size="small" status="danger" :disabled="record.id === 1">删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </a-table-column>
      </template>
    </a-table>

    <!-- 创建/编辑弹窗 -->
    <a-modal v-model:visible="modalVisible" :title="editingId ? '编辑租户' : '新建租户'" @ok="handleSave" @cancel="closeModal">
      <a-form :model="form" layout="vertical">
        <a-form-item label="租户名称" required>
          <a-input v-model="form.name" placeholder="如：XX公司" />
        </a-form-item>
        <a-form-item label="租户编码" required>
          <a-input v-model="form.code" placeholder="如：company-a（字母+数字）" :disabled="!!editingId" />
        </a-form-item>
        <a-form-item label="状态" v-if="editingId">
          <a-switch v-model="form.status" :checked-value="1" :unchecked-value="0" checked-text="启用" unchecked-text="禁用" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Message } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import http from '@/api/http'
import permissionStore from '@/stores/permission'

const tenants = ref([])
const loading = ref(false)
const modalVisible = ref(false)
const editingId = ref(null)
const form = ref({ name: '', code: '', status: 1 })
const isSuperAdmin = computed(() => permissionStore.state.isSuperAdmin)

const formatTime = (ts) => {
  if (!ts || ts === 0) return '—'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

const loadTenants = async () => {
  loading.value = true
  try {
    const res = await http.get('/api/v1/tenants')
    tenants.value = res.data?.items || []
  } catch (e) {
    console.warn('加载租户列表失败', e)
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  form.value = { name: '', code: '', status: 1 }
  modalVisible.value = true
}

const openEdit = (row) => {
  editingId.value = row.id
  form.value = { name: row.name, code: row.code, status: row.status }
  modalVisible.value = true
}

const closeModal = () => {
  modalVisible.value = false
  editingId.value = null
}

const handleSave = async () => {
  if (!form.value.name || !form.value.code) {
    Message.warning('名称和编码不能为空')
    return
  }
  try {
    if (editingId.value) {
      await http.put(`/api/v1/tenants/${editingId.value}`, form.value)
      Message.success('更新成功')
    } else {
      await http.post('/api/v1/tenants', form.value)
      Message.success('创建成功')
    }
    closeModal()
    loadTenants()
  } catch (e) {
    Message.error(e?.response?.data?.msg || '操作失败')
  }
}

const handleDelete = async (row) => {
  try {
    await http.delete(`/api/v1/tenants/${row.id}`)
    Message.success('删除成功')
    loadTenants()
  } catch (e) {
    Message.error(e?.response?.data?.msg || '删除失败')
  }
}

onMounted(loadTenants)
</script>

<style scoped>
.tenant-page { padding: 24px; }
.page-header { margin-bottom: 20px; }
.page-header h2 { margin: 0 0 4px; font-size: 20px; color: #0f172a; }
.subtitle { font-size: 13px; color: #94a3b8; }
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.hint { font-size: 12px; color: #94a3b8; }
</style>

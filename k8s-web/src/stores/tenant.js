import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import http from '@/api/http'

export const useTenantStore = defineStore('tenant', () => {
  const current = ref(null)
  const list = ref([])
  const loading = ref(false)

  const isMultiTenant = computed(() => list.value.length > 1)
  const currentId = computed(() => current.value?.id || 1)
  const currentName = computed(() => current.value?.name || '默认租户')

  const fetchTenants = async () => {
    loading.value = true
    try {
      const res = await http.get('/api/v1/platform/tenants')
      if (res.code === 0 && res.data) {
        list.value = res.data.items || res.data || []
        // 从 JWT 或缓存取当前租户
        const storedId = localStorage.getItem('tenant_id')
        if (storedId) {
          const found = list.value.find(t => t.id === Number(storedId))
          if (found) current.value = found
        }
        if (!current.value && list.value.length > 0) {
          current.value = list.value[0]
        }
      }
    } catch {
      // 降级：使用默认租户
      current.value = { id: 1, name: '默认租户', code: 'default' }
      list.value = [current.value]
    } finally {
      loading.value = false
    }
  }

  const switchTenant = async (tenant) => {
    current.value = tenant
    localStorage.setItem('tenant_id', String(tenant.id))
    // 重新加载页面以刷新 JWT（新 token 包含 tenant_id）
    window.location.reload()
  }

  return { current, list, loading, isMultiTenant, currentId, currentName, fetchTenants, switchTenant }
})

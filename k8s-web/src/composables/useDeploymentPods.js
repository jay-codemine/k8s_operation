// src/composables/useDeploymentPods.js
// 封装：从 Deployment/StatefulSet/DaemonSet 获取关联 Pod 列表，按创建时间排序（最新在前）
import { ref } from 'vue'
import deploymentsApi from '@/api/cluster/workloads/deployments'
import statefulsetsApi from '@/api/cluster/workloads/statefulsets'
import daemonsetsApi from '@/api/cluster/workloads/daemonsets'

/** 根据负载类型选择正确的 API */
function getPodsApi(workloadKind) {
  const kind = (workloadKind || '').toLowerCase()
  if (kind === 'statefulset') return statefulsetsApi.pods
  if (kind === 'daemonset') return daemonsetsApi.pods
  // 默认 Deployment
  return deploymentsApi.pods
}

export function useDeploymentPods() {
  const pods = ref([])
  const loading = ref(false)
  const error = ref('')
  const latestPod = ref(null)

  /**
   * 获取工作负载关联的所有 Pod，按创建时间降序排列
   * @param {string} namespace   - 命名空间
   * @param {string} name        - 工作负载名称
   * @param {string} workloadKind - 类型: Deployment | StatefulSet | DaemonSet
   * @returns {Promise<Array>} 排序后的 Pod 列表
   */
  async function fetchPods(namespace, name, workloadKind = 'Deployment') {
    if (!namespace || !name) {
      error.value = '缺少命名空间或工作负载名称'
      return []
    }
    loading.value = true
    error.value = ''
    try {
      const apiFn = getPodsApi(workloadKind)
      const res = await apiFn({ namespace, name })
      const list = res.code === 0
        ? (res.data?.pods || res.data?.items || res.data?.list || res.data || [])
        : []

      // 解析 pod 信息（兼容后端原生 K8s 对象和已格式化对象）
      const parsed = list.map(p => ({
        name: p.metadata?.name || p.name,
        namespace: p.metadata?.namespace || namespace,
        status: p.status?.phase || p.status || 'Unknown',
        nodeName: p.spec?.nodeName || p.node_name || p.nodeName || '-',
        hostIP: p.status?.hostIP || p.host_ip || '-',
        podIP: p.status?.podIP || p.pod_ip || '-',
        containers: (p.spec?.containers || p.containers || []).map(c => ({
          name: c.name,
          image: c.image,
          ready: false,
          restartCount: 0,
        })),
        containerStatuses: (p.status?.containerStatuses || p.container_statuses || []).map(cs => ({
          name: cs.name,
          ready: cs.ready,
          restartCount: cs.restartCount || cs.restart_count || 0,
          image: cs.image,
        })),
        createdAt: p.metadata?.creationTimestamp || p.created_at || null,
        labels: p.metadata?.labels || p.labels || {},
        conditions: p.status?.conditions || p.conditions || [],
        restartCount: p.status?.containerStatuses
          ? p.status.containerStatuses.reduce((sum, cs) => sum + (cs.restartCount || 0), 0)
          : p.restart_count || 0,
      }))

      // 合并容器状态信息
      parsed.forEach(p => {
        p.containers.forEach(c => {
          const cs = p.containerStatuses.find(s => s.name === c.name)
          if (cs) {
            c.ready = cs.ready
            c.restartCount = cs.restartCount
          }
        })
      })

      // 按创建时间降序：最新的在前
      parsed.sort((a, b) => {
        const ta = a.createdAt ? new Date(a.createdAt).getTime() : 0
        const tb = b.createdAt ? new Date(b.createdAt).getTime() : 0
        return tb - ta
      })

      pods.value = parsed
      latestPod.value = parsed.length > 0 ? parsed[0] : null
      return parsed
    } catch (e) {
      error.value = e.message || '获取 Pod 列表失败'
      pods.value = []
      latestPod.value = null
      return []
    } finally {
      loading.value = false
    }
  }

  /** 获取运行中的 Pod 列表 */
  function runningPods() {
    return pods.value.filter(p => p.status === 'Running')
  }

  /** 格式化创建时间为相对时间 */
  function formatAge(createdAt) {
    if (!createdAt) return '-'
    const now = Date.now()
    const created = new Date(createdAt).getTime()
    const diff = now - created
    if (diff < 0) return '刚刚'
    const secs = Math.floor(diff / 1000)
    if (secs < 60) return `${secs}秒前`
    const mins = Math.floor(secs / 60)
    if (mins < 60) return `${mins}分钟前`
    const hours = Math.floor(mins / 60)
    if (hours < 24) return `${hours}小时前`
    const days = Math.floor(hours / 24)
    return `${days}天前`
  }

  return {
    pods,
    loading,
    error,
    latestPod,
    fetchPods,
    runningPods,
    formatAge,
  }
}

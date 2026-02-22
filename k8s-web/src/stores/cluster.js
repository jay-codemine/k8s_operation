import { defineStore } from 'pinia'

export const useClusterStore = defineStore('cluster', {
  state: () => ({
    current: null, // { id, cluster_name, ... }
  }),
  actions: {
    setCurrent(cluster) {
      this.current = cluster ? { ...cluster, id: Number(cluster.id) } : null

      // ✅ 可选：持久化，刷新/重新打开页面不丢
      if (this.current) {
        localStorage.setItem('currentCluster', JSON.stringify(this.current))
      } else {
        localStorage.removeItem('currentCluster')
      }
    },
    hydrate() {
      const raw = localStorage.getItem('currentCluster')
      if (!raw) return
      try {
        const c = JSON.parse(raw)
        this.current = c ? { ...c, id: Number(c.id) } : null
      } catch {
        this.current = null
      }
    },
  },
})

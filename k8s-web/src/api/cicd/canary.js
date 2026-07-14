import http from '@/api/http'

const canaryApi = {
  // 手动晋升金丝雀
  promote(data) {
    return http.post('/api/v1/k8s/cicd/canary/promote', data)
  },
  // 回滚金丝雀
  rollback(data) {
    return http.post('/api/v1/k8s/cicd/canary/rollback', data)
  },
  // 获取金丝雀状态
  status(params) {
    return http.get('/api/v1/k8s/cicd/canary/status', { params })
  },
  // 调整流量比例
  setTrafficSplit(data) {
    return http.post('/api/v1/k8s/cicd/canary/traffic-split', data)
  },
}

export default canaryApi

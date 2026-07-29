// src/api/http.js
import axios from 'axios'
import {Message} from '@arco-design/web-vue'
import {useClusterStore} from '@/stores/cluster'
import {pinia} from '@/stores'

// ===== 错误消息去重，防止短时间内重复 toast =====
const recentErrors = new Map()
const ERROR_DEDUP_INTERVAL = 3000 // 3秒内相同消息不重复弹出

function showDedupError(msg) {
  const now = Date.now()
  const lastShown = recentErrors.get(msg)
  if (lastShown && now - lastShown < ERROR_DEDUP_INTERVAL) return
  recentErrors.set(msg, now)
  // 清理过期记录
  if (recentErrors.size > 20) {
    for (const [key, time] of recentErrors) {
      if (now - time > ERROR_DEDUP_INTERVAL) recentErrors.delete(key)
    }
  }
  Message.error({ content: msg, duration: 2000 })
}

const http = axios.create({
  // nginx 反向代理 /api/ → 后端，同源相对路径，无需关心端口差异
  baseURL: '',
  timeout: 45000,
  withCredentials: false, // JWT Header 模式：不走 Cookie
})

// ===== token 工具 =====
const getToken = () => localStorage.getItem('token') || sessionStorage.getItem('token')

const setToken = (token) => {
  if (localStorage.getItem('token')) {
    localStorage.setItem('token', token)
  } else if (sessionStorage.getItem('token')) {
    sessionStorage.setItem('token', token)
  } else {
    // 都没有时默认写 localStorage
    localStorage.setItem('token', token)
  }
}

const clearAuth = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  sessionStorage.removeItem('token')
  sessionStorage.removeItem('user')
}

// ===== 并发刷新锁 =====
let isRefreshing = false
let refreshQueue = []

// ===== 这些接口不要触发 refresh（避免死循环）=====
const isAuthPublicApi = (url = '') =>
  url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh') || url.includes('/auth/logout')

// ===== 判断请求 URL 是否需要携带 X-Cluster-ID =====
// 只有真正访问目标集群资源的接口才需要（B 类路由）
// 登录/RBAC/集群管理(CRUD)/CICD 等接口不需要
const needsClusterID = (url = '') => {
  if (!url) return false
  // 不需要 cluster_id 的接口前缀（A类 + 平台级）
  const skipPrefixes = [
    '/auth/',           // 登录/注册/刷新
    '/k8s/cluster/',    // 集群 CRUD 管理
    '/k8s/cicd/',       // CICD 流水线
    '/rbac/',           // 权限管理
    '/user/',           // 用户管理
    '/platform/',       // 平台功能
    '/image/',          // 镜像管理
    '/ai/',             // AI 助手
    '/monitoring/',      // 监控
    '/helloworld',      // 健康检查
  ]
  return !skipPrefixes.some(prefix => url.includes(prefix))
}

// ===== 从 URL 兑底拿 clusterId：/c/:clusterId/... =====
const getClusterIdFromPath = () => {
  try {
    const m = window.location.pathname.match(/\/c\/([^/]+)/)
    return m ? decodeURIComponent(m[1]) : ''
  } catch {
    return ''
  }
}

// ==================
// 请求拦截器：自动带 JWT + X-Cluster-ID
// ==================
http.interceptors.request.use(
  (config) => {
    config.headers = config.headers || {}

    // 1) JWT
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 2) X-Cluster-ID：只在访问集群资源接口时才携带
    //    登录/RBAC/集群CRUD/CICD 等接口不需要 cluster_id
    if (needsClusterID(config.url)) {
      const clusterStore = useClusterStore(pinia)
      const cid = clusterStore.current?.id ?? getClusterIdFromPath()

      // 注意：cid 可能是 number / string；只要不是 null/undefined/空字符串，就写入 header
      if (cid !== undefined && cid !== null && cid !== '') {
        config.headers['X-Cluster-ID'] = String(cid)
      }
    }

    // 3) 禁用 GET 请求缓存：添加时间戳参数
    if (config.method?.toLowerCase() === 'get') {
      config.params = config.params || {}
      config.params._t = Date.now()
    }

    return config
  },
  (error) => Promise.reject(error)
)

// ==================
// 响应拦截器：401 → refresh → 重试
// ==================
http.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const original = error.config
    const status = error.response?.status
    const data = error.response?.data

    // 非 401：按你原逻辑弹错（支持 _silent 静默模式）
    if (status !== 401 && data?.code !== 401) {
      // License 未授权(110001)/已过期(110002)：强制跳转激活页（避免激活页自身死循环）
      if ((data?.code === 110001 || data?.code === 110002) && window.location.pathname !== '/license') {
        window.location.assign('/license')
        return Promise.reject(data || error)
      }
      if (!original._silent) {
        const msg =
          (Array.isArray(data?.details) && data.details[0]) ||
          data?.msg ||
          data?.message ||
          error?.message ||
          '请求失败'
        showDedupError(msg)
      }
      return Promise.reject(data || error)
    }

    // 401：login/register/refresh 自己不要 refresh（否则无限循环）
    if (!original || isAuthPublicApi(original.url) || original._retry) {
      clearAuth()
      window.location.assign('/login')
      return Promise.reject(data || error)
    }

    original._retry = true

    // 已有 refresh 在跑：排队等待新 token
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        refreshQueue.push((newToken) => {
          if (!newToken) {
            reject(error)
            return
          }
          original.headers = original.headers || {}
          original.headers.Authorization = `Bearer ${newToken}`
          resolve(http(original))
        })
      })
    }

    // 开始刷新
    isRefreshing = true

    try {
      // 用一个“裸请求”避免触发本身拦截器递归
      const refreshClient = axios.create({
        baseURL: http.defaults.baseURL,
        timeout: 5000,
        withCredentials: false,
      })

      // refresh 也需要带旧 token（你的后端 refresh 是 Bearer oldToken）
      const oldToken = getToken()
      const r = await refreshClient.post(
        '/api/v1/auth/refresh',
        {},
        {
          headers: oldToken ? {Authorization: `Bearer ${oldToken}`} : {},
        }
      )

      // 兼容不同返回结构
      const body = r.data
      const newToken = body?.data?.token || body?.token || body?.data?.data?.token

      if (!newToken) throw new Error('no token in refresh response')

      setToken(newToken)

      // 放行队列
      refreshQueue.forEach((cb) => cb(newToken))
      refreshQueue = []

      // 重试原请求
      original.headers = original.headers || {}
      original.headers.Authorization = `Bearer ${newToken}`
      return http(original)
    } catch (e) {
      refreshQueue.forEach((cb) => cb(null))
      refreshQueue = []
      clearAuth()
      window.location.assign('/login')
      return Promise.reject(e)
    } finally {
      isRefreshing = false
    }
  }
)

export default http

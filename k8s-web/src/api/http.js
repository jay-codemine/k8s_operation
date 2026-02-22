// src/api/http.js
import axios from 'axios'
import {Message} from '@arco-design/web-vue'
import {useClusterStore} from '@/stores/cluster'
import {pinia} from '@/stores'

const http = axios.create({
  baseURL: '',
  timeout: 15000,
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
  url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh')

// ===== 从 URL 兜底拿 clusterId：/c/:clusterId/... =====
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

    // 2) X-Cluster-ID（数字也 OK，header 强转 string）
    //    组件外用 store 一定要传 pinia；并且用 URL 做兜底，避免刷新时 store 还没恢复
    const clusterStore = useClusterStore(pinia)
    const cid = clusterStore.current?.id ?? getClusterIdFromPath()

    // 注意：cid 可能是 number / string；只要不是 null/undefined/空字符串，就写入 header
    if (cid !== undefined && cid !== null && cid !== '') {
      config.headers['X-Cluster-ID'] = String(cid)
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

    // 非 401：按你原逻辑弹错
    if (status !== 401 && data?.code !== 401) {
      const msg =
        (Array.isArray(data?.details) && data.details[0]) ||
        data?.msg ||
        data?.message ||
        error?.message ||
        '请求失败'
      Message.error({content: msg, duration: 2000})
      return Promise.reject(data || error)
    }


    if (status !== 401 && data?.code !== 401) {
      return Promise.reject(error)
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
        '/auth/refresh',
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

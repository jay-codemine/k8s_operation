// fetchWithAuth: 为原生 fetch 添加 JWT 认证头
// 用于替代 view 中的裸 fetch() 调用，确保通过 JWT 鉴权

const getToken = () => localStorage.getItem('token') || sessionStorage.getItem('token')

export function fetchWithAuth(url, options = {}) {
  const token = getToken()
  const headers = { ...options.headers }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  return fetch(url, { ...options, headers })
}

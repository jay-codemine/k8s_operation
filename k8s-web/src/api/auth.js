// src/api/auth.js
import http from './http'
import {API_BASE} from './paths'

// 注册
export function register(data) {
  return http.post(`${API_BASE}/auth/register`, data)
}

// 登录
export function login(data) {
  return http.post(`${API_BASE}/auth/login`, data)
}

// 刷新 token
export function refreshToken() {
  return http.post(`${API_BASE}/auth/refresh`)
}

// 退出登录
export function logout() {
  return http.post(`${API_BASE}/auth/logout`)
}

// 忘记密码
export function forgotPassword(data) {
  return http.post(`${API_BASE}/auth/forgot_password`, data)
}

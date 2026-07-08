// src/api/ldap.js
import http from './http'
import { API_BASE } from './paths'

// 获取 LDAP 配置状态
export function getLDAPConfig() {
  return http.get(`${API_BASE}/ldap/config`)
}

// 测试 LDAP 连接
export function testLDAPConnection() {
  return http.post(`${API_BASE}/ldap/test-connection`)
}

// 同步 LDAP 用户
export function syncLDAPUsers() {
  return http.post(`${API_BASE}/ldap/sync-users`)
}

// 获取 LDAP 状态（是否启用/连接状态）
export function getLDAPStatus() {
  return http.get(`${API_BASE}/ldap/status`)
}

import http from './http'

/**
 * 用户管理 API
 */

// 获取用户列表
export function getUserList(params) {
  return http({
    url: '/api/v1/user/list',
    method: 'get',
    params: {
      page: params.page || 1,
      limit: params.limit || 10,
      username: params.username || '',
      role: params.role || '',
      status: params.status !== undefined ? params.status : ''
    }
  })
}

// 创建用户
export function createUser(data) {
  return http({
    url: '/api/v1/user/create',
    method: 'post',
    data
  })
}

// 更新用户
export function updateUser(data) {
  return http({
    url: '/api/v1/user/update',
    method: 'post',
    data
  })
}

// 删除用户
export function deleteUser(id) {
  return http({
    url: '/api/v1/user/delete',
    method: 'post',
    data: { id }
  })
}

// 批量删除用户
export function batchDeleteUsers(ids) {
  return Promise.all(ids.map(id => deleteUser(id)))
}

// 更新用户状态
export function updateUserStatus(user, status) {
  return http({
    url: '/api/v1/user/update',
    method: 'post',
    data: { 
      id: user.id, 
      username: user.username,
      status 
    }
  })
}

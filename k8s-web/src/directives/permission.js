/**
 * v-permission 权限指令
 * 用于按钮级权限控制
 * 
 * 使用方式:
 * 1. 角色权限: v-permission="['super_admin', 'platform_admin']"
 * 2. 集群权限: v-permission="{ cluster: clusterId, action: 'delete' }"
 * 3. 资源权限: v-permission="{ resource: 'deployment', action: 'create', cluster: clusterId }"
 */
import permissionStore from '@/stores/permission'

/**
 * 检查权限
 * @param {Array|Object} value 权限配置
 * @returns {boolean}
 */
function checkPermission(value) {
  // 超级管理员拥有所有权限
  if (permissionStore.state.isSuperAdmin) {
    return true
  }
  
  // 数组格式：角色列表
  if (Array.isArray(value)) {
    const userRoles = permissionStore.roleTypes.value
    return value.some(role => userRoles.includes(role))
  }
  
  // 对象格式：详细权限配置
  if (typeof value === 'object' && value !== null) {
    // 集群权限检查
    if (value.cluster) {
      const action = value.action || 'view'
      return permissionStore.canAccessCluster(value.cluster, action)
    }
    
    // 资源权限检查
    if (value.resource) {
      const action = value.action || 'view'
      return permissionStore.hasPermission(value.resource, action, value.cluster)
    }
    
    // 角色检查
    if (value.roles) {
      const userRoles = permissionStore.roleTypes.value
      return value.roles.some(role => userRoles.includes(role))
    }
  }
  
  return true
}

/**
 * 权限指令定义
 */
export const permissionDirective = {
  mounted(el, binding) {
    const hasPermission = checkPermission(binding.value)
    
    if (!hasPermission) {
      // 默认隐藏元素
      if (binding.modifiers.disable) {
        // 使用 .disable 修饰符时，禁用而不是隐藏
        el.disabled = true
        el.classList.add('permission-disabled')
        el.title = '您没有此操作的权限'
      } else {
        // 默认移除元素
        el.parentNode?.removeChild(el)
      }
    }
  },
  
  updated(el, binding) {
    const hasPermission = checkPermission(binding.value)
    
    if (!hasPermission) {
      if (binding.modifiers.disable) {
        el.disabled = true
        el.classList.add('permission-disabled')
      } else {
        el.style.display = 'none'
      }
    } else {
      if (binding.modifiers.disable) {
        el.disabled = false
        el.classList.remove('permission-disabled')
      } else {
        el.style.display = ''
      }
    }
  }
}

/**
 * 权限检查组件 - 用于包裹需要权限控制的内容
 */
export const PermissionWrapper = {
  name: 'PermissionWrapper',
  props: {
    // 角色列表
    roles: {
      type: Array,
      default: () => []
    },
    // 集群ID
    cluster: {
      type: [Number, String],
      default: null
    },
    // 操作类型
    action: {
      type: String,
      default: 'view'
    },
    // 资源类型
    resource: {
      type: String,
      default: null
    },
    // 无权限时的显示方式: hide/disable/placeholder
    fallback: {
      type: String,
      default: 'hide'
    }
  },
  setup(props, { slots }) {
    const hasPermission = () => {
      // 超级管理员
      if (permissionStore.state.isSuperAdmin) return true
      
      // 角色检查
      if (props.roles.length > 0) {
        const userRoles = permissionStore.roleTypes.value
        if (!props.roles.some(role => userRoles.includes(role))) {
          return false
        }
      }
      
      // 集群权限
      if (props.cluster) {
        if (!permissionStore.canAccessCluster(props.cluster, props.action)) {
          return false
        }
      }
      
      // 资源权限
      if (props.resource) {
        if (!permissionStore.hasPermission(props.resource, props.action, props.cluster)) {
          return false
        }
      }
      
      return true
    }
    
    return () => {
      if (hasPermission()) {
        return slots.default?.()
      }
      
      // 无权限时的处理
      switch (props.fallback) {
        case 'placeholder':
          return slots.placeholder?.() || null
        case 'disable':
          // 对于 disable 模式，仍然渲染但添加 disabled 属性
          return slots.default?.()
        default:
          return null
      }
    }
  }
}

/**
 * 安装权限插件
 */
export function setupPermission(app) {
  // 注册指令
  app.directive('permission', permissionDirective)
  
  // 注册组件
  app.component('PermissionWrapper', PermissionWrapper)
  
  // 添加全局方法
  app.config.globalProperties.$hasPermission = (permission) => {
    return checkPermission(permission)
  }
  
  // 添加全局权限状态
  app.config.globalProperties.$permission = permissionStore
}

export default {
  install: setupPermission
}

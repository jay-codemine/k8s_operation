/**
 * useConfirmDialog.js - 全局二次确认弹窗 composable
 * 
 * 替代浏览器原生 confirm()，提供大厂级 UI 体验
 * 支持多种场景：危险操作、信息确认、警告提示
 * 
 * 使用示例：
 *   const { confirm } = useConfirmDialog()
 *   const ok = await confirm({
 *     title: '确认更新镜像',
 *     content: '此操作将触发滚动更新',
 *     type: 'warning',  // info | warning | danger | success
 *     details: [
 *       { label: 'Deployment', value: 'default/my-app' },
 *       { label: '新镜像', value: 'nginx:1.25' },
 *     ],
 *     confirmText: '确认更新',
 *     cancelText: '取消',
 *   })
 *   if (!ok) return
 */
import { ref, reactive } from 'vue'

// 全局单例状态
const visible = ref(false)
const dialogState = reactive({
  title: '',
  content: '',
  type: 'warning', // info | warning | danger | success
  details: [],     // [{ label, value, highlight? }]
  confirmText: '确认',
  cancelText: '取消',
  icon: '',        // 可选自定义 icon
  tip: '',         // 底部提示文字
})

let resolvePromise = null

const openDialog = (options = {}) => {
  return new Promise((resolve) => {
    resolvePromise = resolve
    Object.assign(dialogState, {
      title: options.title || '操作确认',
      content: options.content || '',
      type: options.type || 'warning',
      details: options.details || [],
      confirmText: options.confirmText || '确认',
      cancelText: options.cancelText || '取消',
      icon: options.icon || '',
      tip: options.tip || '',
    })
    visible.value = true
  })
}

const handleConfirm = () => {
  visible.value = false
  resolvePromise?.(true)
  resolvePromise = null
}

const handleCancel = () => {
  visible.value = false
  resolvePromise?.(false)
  resolvePromise = null
}

export function useConfirmDialog() {
  return {
    // 状态（给 ConfirmDialog 组件消费）
    visible,
    dialogState,
    handleConfirm,
    handleCancel,
    // 调用方法（给业务组件使用）
    confirm: openDialog,
  }
}

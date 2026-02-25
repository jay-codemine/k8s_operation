/**
 * 可拖拽调整大小的模态框 Composable
 * 
 * 使用方法：
 * import { useResizableModal } from '@/composables/useResizableModal'
 * 
 * const { modalRef, modalStyle, resizeHandles, startResize } = useResizableModal({
 *   initialWidth: '800px',
 *   initialHeight: 'auto'
 * })
 */

import { ref, onUnmounted } from 'vue'

export function useResizableModal(options = {}) {
  const {
    initialWidth = '800px',
    initialHeight = 'auto',
    minWidth = 400,
    minHeight = 300,
    maxWidthPercent = 0.95,
    maxHeightPercent = 0.95
  } = options

  // 模态框引用
  const modalRef = ref(null)

  // 模态框样式
  const modalStyle = ref({
    width: initialWidth,
    maxWidth: '90vw',
    height: initialHeight,
    maxHeight: '90vh'
  })

  // 拖拽状态
  const resizing = ref({
    isResizing: false,
    direction: '',
    startX: 0,
    startY: 0,
    startWidth: 0,
    startHeight: 0
  })

  // 开始拖拽
  const startResize = (event, direction) => {
    event.preventDefault()
    event.stopPropagation()
    
    const modal = modalRef.value
    if (!modal) return
    
    const rect = modal.getBoundingClientRect()
    
    resizing.value = {
      isResizing: true,
      direction,
      startX: event.clientX,
      startY: event.clientY,
      startWidth: rect.width,
      startHeight: rect.height,
      startLeft: rect.left,
      startTop: rect.top
    }
    
    document.addEventListener('mousemove', handleResize)
    document.addEventListener('mouseup', stopResize)
    document.body.style.userSelect = 'none'
    document.body.style.cursor = getCursor(direction)
  }

  // 处理拖拽
  const handleResize = (event) => {
    if (!resizing.value.isResizing) return
    
    const deltaX = event.clientX - resizing.value.startX
    const deltaY = event.clientY - resizing.value.startY
    const direction = resizing.value.direction
    
    let newWidth = resizing.value.startWidth
    let newHeight = resizing.value.startHeight
    
    const maxWidth = window.innerWidth * maxWidthPercent
    const maxHeight = window.innerHeight * maxHeightPercent
    
    // 根据拖拽方向调整大小
    if (direction.includes('right')) {
      newWidth = Math.min(Math.max(resizing.value.startWidth + deltaX, minWidth), maxWidth)
    }
    if (direction.includes('left')) {
      newWidth = Math.min(Math.max(resizing.value.startWidth - deltaX, minWidth), maxWidth)
    }
    if (direction.includes('bottom')) {
      newHeight = Math.min(Math.max(resizing.value.startHeight + deltaY, minHeight), maxHeight)
    }
    if (direction.includes('top')) {
      newHeight = Math.min(Math.max(resizing.value.startHeight - deltaY, minHeight), maxHeight)
    }
    
    modalStyle.value = {
      width: `${newWidth}px`,
      height: `${newHeight}px`,
      maxWidth: `${maxWidthPercent * 100}vw`,
      maxHeight: `${maxHeightPercent * 100}vh`
    }
  }

  // 停止拖拽
  const stopResize = () => {
    resizing.value.isResizing = false
    document.removeEventListener('mousemove', handleResize)
    document.removeEventListener('mouseup', stopResize)
    document.body.style.userSelect = ''
    document.body.style.cursor = ''
  }

  // 获取鼠标指针样式
  const getCursor = (direction) => {
    const cursors = {
      'top': 'n-resize',
      'bottom': 's-resize',
      'left': 'w-resize',
      'right': 'e-resize',
      'top-left': 'nw-resize',
      'top-right': 'ne-resize',
      'bottom-left': 'sw-resize',
      'bottom-right': 'se-resize'
    }
    return cursors[direction] || 'default'
  }

  // 拖拽手柄HTML片段（用于模板）
  const resizeHandles = `
    <div class="resize-handle resize-handle-top" @mousedown="startResize($event, 'top')"></div>
    <div class="resize-handle resize-handle-bottom" @mousedown="startResize($event, 'bottom')"></div>
    <div class="resize-handle resize-handle-left" @mousedown="startResize($event, 'left')"></div>
    <div class="resize-handle resize-handle-right" @mousedown="startResize($event, 'right')"></div>
    <div class="resize-handle resize-handle-top-left" @mousedown="startResize($event, 'top-left')"></div>
    <div class="resize-handle resize-handle-top-right" @mousedown="startResize($event, 'top-right')"></div>
    <div class="resize-handle resize-handle-bottom-left" @mousedown="startResize($event, 'bottom-left')"></div>
    <div class="resize-handle resize-handle-bottom-right" @mousedown="startResize($event, 'bottom-right')"></div>
  `

  // 组件卸载时清理
  onUnmounted(() => {
    stopResize()
  })

  return {
    modalRef,
    modalStyle,
    resizeHandles,
    startResize
  }
}

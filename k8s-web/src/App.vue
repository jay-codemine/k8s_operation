<!-- src/App.vue -->
<template>
  <RouterView />
  <ConfirmDialog />
</template>

<script setup>
import { RouterView } from 'vue-router'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
</script>

<style>
/* ===== 全局重置 ===== */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* ===== 响应式根字体大小 ===== */
html {
  /* 
   * 响应式字体：基于视口宽度自动缩放
   * - 最小 13px（小屏幕/高分屏 100% 缩放）
   * - 最大 16px（大屏幕）
   * - 中间值：0.8vw（视口宽度的 0.8%）
   * 
   * 1920px 屏幕：0.8% = 15.36px
   * 1440px 屏幕：0.8% = 11.52px → 取 min 13px
   * 2560px 屏幕：0.8% = 20.48px → 取 max 16px
   */
  font-size: clamp(13px, 0.8vw, 16px);
  height: 100%;
  width: 100%;
  overflow: hidden;
}

body {
  font-size: 1rem; /* 继承 html 的响应式字体 */
  height: 100%;
  width: 100%;
  overflow: hidden;
  line-height: 1.5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  height: 100%;
  width: 100%;
  overflow: auto;
}

/* ===== 响应式工具类 ===== */
/* 可在任何组件中使用这些类 */
.text-xs { font-size: 0.75rem; }   /* 12px @ 16px base */
.text-sm { font-size: 0.875rem; }  /* 14px @ 16px base */
.text-base { font-size: 1rem; }    /* 16px @ 16px base */
.text-lg { font-size: 1.125rem; }  /* 18px @ 16px base */
.text-xl { font-size: 1.25rem; }   /* 20px @ 16px base */

/* ===== 针对不同屏幕尺寸的细调 ===== */
/* 超大屏幕（4K / 2560px+） */
@media (min-width: 2560px) {
  html {
    font-size: 16px;
  }
}

/* 大屏幕（1920px - 2559px）*/
@media (min-width: 1920px) and (max-width: 2559px) {
  html {
    font-size: 15px;
  }
}

/* 中等屏幕（1440px - 1919px）*/
@media (min-width: 1440px) and (max-width: 1919px) {
  html {
    font-size: 14px;
  }
}

/* 小屏幕（1280px - 1439px）*/
@media (min-width: 1280px) and (max-width: 1439px) {
  html {
    font-size: 13px;
  }
}

/* 更小屏幕（1024px - 1279px）*/
@media (min-width: 1024px) and (max-width: 1279px) {
  html {
    font-size: 12px;
  }
}

/* 平板及以下（< 1024px）*/
@media (max-width: 1023px) {
  html {
    font-size: 14px; /* 移动端保持可读性 */
  }
}

/* ========================================
   全局布局修复 - 解决所有视图内容截断问题
   ======================================== */

/* 1. 主容器：移除 max-width 限制，让内容填满可用空间 */
.resource-view {
  max-width: 100% !important;
  width: 100% !important;
  margin: 0 !important;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* 2. 表格容器：优化滚动和圆角 */
.table-container {
  width: 100% !important;
  overflow-x: auto !important;
  overflow-y: visible !important;
  border-radius: 12px !important;
  background: #fff !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04) !important;
}

/* 3. 表格：移除过大的 min-width，自适应内容 */
.resource-table {
  width: 100% !important;
  min-width: 0 !important;
  table-layout: auto !important;
  border-collapse: collapse !important;
}

/* 4. 表头/单元格：紧凑化，减少 padding */
.resource-table th {
  padding: 12px 14px !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  color: #475569 !important;
  background: #f8fafc !important;
  border-bottom: 2px solid #e2e8f0 !important;
  white-space: nowrap !important;
  letter-spacing: 0.02em !important;
}

.resource-table td {
  padding: 12px 14px !important;
  font-size: 13px !important;
  color: #334155 !important;
  border-bottom: 1px solid #f1f5f9 !important;
  vertical-align: middle !important;
}

.resource-table tbody tr {
  transition: background-color 0.15s ease !important;
}

.resource-table tbody tr:hover {
  background-color: #f8fafc !important;
}

/* 5. 页面标题区 - 更紧凑 */
.view-header {
  margin-bottom: 18px !important;
}

.view-header h1 {
  font-size: 24px !important;
  font-weight: 700 !important;
  color: #1e293b !important;
  margin-bottom: 4px !important;
  letter-spacing: -0.02em !important;
}

.view-header p {
  font-size: 14px !important;
  color: #64748b !important;
  margin: 0 !important;
}

/* 6. 操作栏 - 紧凑对齐 */
.action-bar {
  display: flex !important;
  align-items: center !important;
  flex-wrap: wrap !important;
  gap: 10px !important;
  margin-bottom: 16px !important;
}

/* 7. 搜索框 - 自适应宽度 */
.search-box input {
  width: clamp(180px, 18vw, 300px) !important;
  padding: 8px 14px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 8px !important;
  font-size: 13px !important;
  transition: all 0.2s ease !important;
}

.search-box input:focus {
  outline: none !important;
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1) !important;
}

/* 8. 筛选按钮 - 更现代 */
.btn-filter {
  padding: 6px 14px !important;
  font-size: 13px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 6px !important;
  background: #fff !important;
  color: #64748b !important;
  cursor: pointer !important;
  transition: all 0.2s !important;
  white-space: nowrap !important;
}

.btn-filter:hover {
  background: #f1f5f9 !important;
  border-color: #cbd5e1 !important;
}

.btn-filter.active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%) !important;
  border-color: #3b82f6 !important;
  color: #fff !important;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3) !important;
}

/* 9. 按钮通用优化 */
.btn {
  padding: 8px 16px !important;
  border-radius: 8px !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  cursor: pointer !important;
  transition: all 0.2s ease !important;
  white-space: nowrap !important;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%) !important;
  color: #fff !important;
  border: none !important;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.25) !important;
}

.btn-primary:hover {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35) !important;
  transform: translateY(-1px) !important;
}

.btn-secondary {
  background: #f1f5f9 !important;
  color: #475569 !important;
  border: 1px solid #e2e8f0 !important;
}

.btn-secondary:hover {
  background: #e2e8f0 !important;
}

/* 10. 状态指示器优化 */
.status-indicator {
  display: inline-flex !important;
  align-items: center !important;
  gap: 5px !important;
  padding: 4px 10px !important;
  border-radius: 6px !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  white-space: nowrap !important;
}

.status-indicator::before {
  content: '' !important;
  display: inline-block !important;
  width: 6px !important;
  height: 6px !important;
  border-radius: 50% !important;
  flex-shrink: 0 !important;
}

.status-indicator.running::before { background: #22c55e !important; }
.status-indicator.failed::before { background: #ef4444 !important; }
.status-indicator.updating::before { background: #f59e0b !important; }
.status-indicator.stopped::before { background: #94a3b8 !important; }
.status-indicator.pending::before { background: #3b82f6 !important; }
.status-indicator.succeeded::before { background: #22c55e !important; }

.status-indicator.running {
  background: rgba(34, 197, 94, 0.08) !important;
  color: #16a34a !important;
}
.status-indicator.failed {
  background: rgba(239, 68, 68, 0.08) !important;
  color: #dc2626 !important;
}
.status-indicator.updating {
  background: rgba(245, 158, 11, 0.08) !important;
  color: #d97706 !important;
}
.status-indicator.stopped {
  background: rgba(148, 163, 184, 0.1) !important;
  color: #64748b !important;
}
.status-indicator.pending {
  background: rgba(59, 130, 246, 0.08) !important;
  color: #2563eb !important;
}
.status-indicator.succeeded {
  background: rgba(34, 197, 94, 0.08) !important;
  color: #16a34a !important;
}

/* 11. 命名空间标签 */
.namespace-badge {
  display: inline-flex !important;
  padding: 3px 8px !important;
  background: #eff6ff !important;
  color: #2563eb !important;
  border-radius: 5px !important;
  font-size: 12px !important;
  font-weight: 500 !important;
  white-space: nowrap !important;
}

/* 12. 错误框 */
.error-box {
  background: #fef2f2 !important;
  border: 1px solid #fecaca !important;
  color: #dc2626 !important;
  padding: 10px 14px !important;
  border-radius: 8px !important;
  margin-bottom: 14px !important;
  font-size: 13px !important;
}

/* 13. 分页优化 */
.pagination {
  display: flex !important;
  justify-content: center !important;
  align-items: center !important;
  gap: 6px !important;
  padding: 14px 0 !important;
  flex-shrink: 0 !important;
}

/* 14. 弹窗优化 */
.modal-overlay {
  position: fixed !important;
  inset: 0 !important;
  background: rgba(15, 23, 42, 0.5) !important;
  backdrop-filter: blur(4px) !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  z-index: 1000 !important;
}

.modal-content {
  background: #fff !important;
  border-radius: 16px !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25) !important;
  max-height: 85vh !important;
  overflow-y: auto !important;
}

/* 15. 全局滚动条美化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* 16. 表格横向滚动条特殊处理 */
.table-container::-webkit-scrollbar {
  height: 6px;
}

.table-container::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 0 0 12px 12px;
}

.table-container::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

/* 17. 文本溢出处理 */
.deployment-name,
.pod-name,
.resource-name {
  white-space: nowrap !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  max-width: 200px !important;
}

/* 18. 选择器标签溢出处理 */
.selector-tags {
  display: flex !important;
  flex-wrap: wrap !important;
  gap: 4px !important;
  max-width: 220px !important;
}

.selector-tag {
  display: inline-block !important;
  padding: 2px 6px !important;
  background: #f1f5f9 !important;
  color: #475569 !important;
  border-radius: 4px !important;
  font-size: 11px !important;
  white-space: nowrap !important;
  max-width: 180px !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
}

/* 19. 镜像名称溢出 */
.image-text,
.resource-table td .image-name {
  max-width: 280px !important;
  white-space: nowrap !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  display: inline-block !important;
}
</style>

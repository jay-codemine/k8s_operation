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
/* ===== 主题：亮色 / 暗色 ===== */
:root,
[data-theme="light"] {
  --bg-primary: #ffffff;
  --bg-secondary: #f8fafc;
  --bg-tertiary: #f1f5f9;
  --bg-card: #ffffff;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --border-color: #e2e8f0;
  --shadow-sm: 0 1px 3px rgba(15, 23, 42, 0.04);
  --shadow-md: 0 4px 20px rgba(0, 0, 0, 0.06);
  --accent: #6366f1;
  --accent-hover: #4f46e5;
  --badge-bg: #eef2ff;
  --input-bg: #ffffff;
  --hover-bg: #f1f5f9;
}

[data-theme="dark"] {
  --bg-primary: #0b1120;
  --bg-secondary: #111827;
  --bg-tertiary: #1e293b;
  --bg-card: #1a2332;
  --text-primary: #e2e8f0;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;
  --border-color: #1e293b;
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 20px rgba(0, 0, 0, 0.4);
  --accent: #818cf8;
  --accent-hover: #6366f1;
  --badge-bg: #1e1b4b;
  --input-bg: #1e293b;
  --hover-bg: #1e293b;
}

/* ===== 全局重置 ===== */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 14px;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

body {
  font-size: 1rem;
  height: 100%;
  width: 100%;
  overflow: hidden;
  line-height: 1.5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: background 0.3s, color 0.3s;
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

/* 表格横向滚动条特殊处理 */
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

/* ===== 暗色模式：全面覆盖（穿透 Vue scoped） ===== */
[data-theme="dark"] {
  color-scheme: dark;
}

/* 全局背景 */
[data-theme="dark"] body,
[data-theme="dark"] #app,
[data-theme="dark"] .home-page,
[data-theme="dark"] .health-dashboard {
  background: #0b1120 !important;
}

/* 主内容区 */
[data-theme="dark"] .page-content {
  background: #0b1120 !important;
}

/* 侧边栏 */
[data-theme="dark"] .sidebar,
[data-theme="dark"] div[class*="sidebar"] {
  background: #0f172a !important;
  border-right-color: #1e293b !important;
}
[data-theme="dark"] .sidebar .group-header,
[data-theme="dark"] .sidebar .nav-item {
  color: #94a3b8 !important;
}
[data-theme="dark"] .sidebar .group-header:hover,
[data-theme="dark"] .sidebar .nav-item:hover {
  background: #1e293b !important;
  color: #e2e8f0 !important;
}
[data-theme="dark"] .sidebar .nav-item.active {
  background: rgba(99,102,241,0.15) !important;
  color: #818cf8 !important;
}

/* 顶部导航 */
[data-theme="dark"] .top-nav,
[data-theme="dark"] header[class*="top-nav"],
[data-theme="dark"] header[class*="nav"] {
  background: #0f172a !important;
  border-bottom: 1px solid #1e293b !important;
}

/* 卡片和面板 */
[data-theme="dark"] .stat-card,
[data-theme="dark"] .panel,
[data-theme="dark"] div[class*="panel"],
[data-theme="dark"] div[class*="card"]:not(.metric-card),
[data-theme="dark"] div[class*="Card"] {
  background: #1a2332 !important;
  border-color: #1e293b !important;
  box-shadow: 0 1px 3px rgba(0,0,0,0.3) !important;
}

/* 指标卡片 */
[data-theme="dark"] .metric-card,
[data-theme="dark"] div[class*="metric-card"] {
  background: #1a2332 !important;
  border-color: #1e293b !important;
}

/* 表格 */
[data-theme="dark"] .arco-table,
[data-theme="dark"] table[class*="table"],
[data-theme="dark"] .mini-table {
  background: #1a2332 !important;
}
[data-theme="dark"] .arco-table th,
[data-theme="dark"] thead th {
  background: #111827 !important;
  color: #94a3b8 !important;
  border-color: #1e293b !important;
}
[data-theme="dark"] .arco-table td,
[data-theme="dark"] tbody td {
  background: #1a2332 !important;
  color: #e2e8f0 !important;
  border-color: #1e293b !important;
}

/* 文字 */
[data-theme="dark"] h1,
[data-theme="dark"] h2,
[data-theme="dark"] h3,
[data-theme="dark"] .stat-value,
[data-theme="dark"] .metric-val,
[data-theme="dark"] .card-title,
[data-theme="dark"] .panel-header h3 {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .stat-label,
[data-theme="dark"] .metric-lbl,
[data-theme="dark"] .card-subtitle,
[data-theme="dark"] .section-sub {
  color: #64748b !important;
}

/* 输入框 */
[data-theme="dark"] input,
[data-theme="dark"] select,
[data-theme="dark"] textarea {
  background: #1e293b !important;
  color: #e2e8f0 !important;
  border-color: #334155 !important;
}

/* 弹窗 */
[data-theme="dark"] .arco-modal-content,
[data-theme="dark"] .arco-modal {
  background: #1a2332 !important;
}

/* 标签 */
[data-theme="dark"] .badge-green {
  background: rgba(16,185,129,0.15) !important;
  color: #34d399 !important;
}
[data-theme="dark"] .badge-red {
  background: rgba(239,68,68,0.15) !important;
  color: #f87171 !important;
}
[data-theme="dark"] .badge-amber {
  background: rgba(245,158,11,0.15) !important;
  color: #fbbf24 !important;
}

/* 按钮 */
[data-theme="dark"] .btn-refresh,
[data-theme="dark"] button[class*="primary"] {
  background: #6366f1 !important;
}

/* 滚动条 */
[data-theme="dark"]::-webkit-scrollbar-thumb {
  background: #334155 !important;
}
[data-theme="dark"]::-webkit-scrollbar-track {
  background: #0f172a !important;
}

/* 系统状态栏 */
[data-theme="dark"] .system-bar {
  background: #1a2332 !important;
  border-color: #1e293b !important;
}

/* 分页 */
[data-theme="dark"] .arco-pagination li {
  background: #1e293b !important;
  color: #e2e8f0 !important;
}

/* ===== 暗色模式补充：业务页面浅色硬编码覆盖 ===== */
/* 页面顶层容器（浅色渐变/白色背景） */
[data-theme="dark"] .pipeline-wizard,
[data-theme="dark"] .page-wrapper,
[data-theme="dark"] .page-container,
[data-theme="dark"] .list-container,
[data-theme="dark"] .detail-container {
  background: #0b1120 !important;
}

/* 向导/表单内的浅色区块 */
[data-theme="dark"] .wizard-card,
[data-theme="dark"] .form-card,
[data-theme="dark"] .step-content,
[data-theme="dark"] .config-summary,
[data-theme="dark"] .branch-selector,
[data-theme="dark"] .branch-search,
[data-theme="dark"] .env-chip,
[data-theme="dark"] .kind-card,
[data-theme="dark"] .deploy-preview,
[data-theme="dark"] .preview-item {
  background: #1a2332 !important;
  border-color: #334155 !important;
}

/* 表单文字 */
[data-theme="dark"] label,
[data-theme="dark"] .form-label {
  color: #cbd5e1 !important;
}
[data-theme="dark"] p,
[data-theme="dark"] .input-hint,
[data-theme="dark"] .summary-label,
[data-theme="dark"] .confirm-label,
[data-theme="dark"] .preview-label,
[data-theme="dark"] .kind-desc {
  color: #94a3b8 !important;
}
[data-theme="dark"] .summary-value,
[data-theme="dark"] .branch-item,
[data-theme="dark"] .kind-name,
[data-theme="dark"] .env-chip,
[data-theme="dark"] .confirm-value,
[data-theme="dark"] .preview-value,
[data-theme="dark"] .form-group h4,
[data-theme="dark"] .section-title {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .branch-item:hover {
  background: #1e293b !important;
}
[data-theme="dark"] .branch-item.selected {
  background: rgba(99,102,241,0.15) !important;
}
[data-theme="dark"] .env-chip.selected,
[data-theme="dark"] .kind-card.selected {
  background: rgba(99,102,241,0.15) !important;
  border-color: #6366f1 !important;
  color: #818cf8 !important;
}

/* 弹窗 */
[data-theme="dark"] .confirm-modal,
[data-theme="dark"] .k8s-import-modal,
[data-theme="dark"] .topology-card {
  background: #1a2332 !important;
  border-color: #334155 !important;
  color: #e2e8f0 !important;
}

/* SearchSelect 可搜索下拉 */
[data-theme="dark"] .ss-trigger {
  background: #1e293b !important;
  border-color: #334155 !important;
}
[data-theme="dark"] .ss-open {
  background: #1e293b !important;
  border-color: #6366f1 !important;
  box-shadow: 0 0 0 4px rgba(99,102,241,0.15) !important;
}
[data-theme="dark"] .ss-val,
[data-theme="dark"] .ss-opt-label {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .ss-placeholder,
[data-theme="dark"] .ss-arrow {
  color: #64748b !important;
}
[data-theme="dark"] .ss-drop {
  background: #1a2332 !important;
  border-color: #334155 !important;
  box-shadow: 0 12px 40px rgba(0,0,0,.5) !important;
}
[data-theme="dark"] .ss-search {
  border-color: #334155 !important;
}
[data-theme="dark"] .ss-input {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #e2e8f0 !important;
}
[data-theme="dark"] .ss-option:hover {
  background: #1e293b !important;
}
[data-theme="dark"] .ss-sel {
  background: rgba(99,102,241,0.18) !important;
}
[data-theme="dark"] .ss-group-label,
[data-theme="dark"] .ss-opt-sub,
[data-theme="dark"] .ss-empty {
  color: #64748b !important;
}

/* 顶栏操作按钮（通知/搜索/主题） */
[data-theme="dark"] .nav-action-btn {
  color: #94a3b8 !important;
}
[data-theme="dark"] .nav-action-btn:hover {
  background: #1e293b !important;
  color: #e2e8f0 !important;
}

/* 通知铃铛面板 */
[data-theme="dark"] .nb-panel {
  background: #1a2332 !important;
  border-color: #334155 !important;
  box-shadow: 0 12px 40px rgba(0,0,0,.5) !important;
}
[data-theme="dark"] .nb-header {
  border-color: #334155 !important;
}
[data-theme="dark"] .nb-title {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .nb-read-all {
  color: #818cf8 !important;
}
[data-theme="dark"] .nb-item {
  border-color: #243044 !important;
}
[data-theme="dark"] .nb-item:hover {
  background: #1e293b !important;
}
[data-theme="dark"] .nb-unread {
  background: rgba(99,102,241,.10) !important;
}
[data-theme="dark"] .nb-unread:hover {
  background: rgba(99,102,241,.16) !important;
}
[data-theme="dark"] .nb-rule {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .nb-summary,
[data-theme="dark"] .nb-time,
[data-theme="dark"] .nb-empty {
  color: #64748b !important;
}
[data-theme="dark"] .nb-footer {
  border-color: #334155 !important;
  color: #818cf8 !important;
}
[data-theme="dark"] .nb-footer:hover {
  background: #1e293b !important;
}
[data-theme="dark"] .nb-status.firing {
  background: rgba(220,38,38,.15) !important;
  color: #f87171 !important;
}
[data-theme="dark"] .nb-status.resolved {
  background: rgba(5,150,105,.15) !important;
  color: #34d399 !important;
}
[data-theme="dark"] .nb-status.silenced {
  background: #334155 !important;
  color: #94a3b8 !important;
}

/* 全局搜索弹窗 */
[data-theme="dark"] .gs-modal {
  background: #1a2332 !important;
  box-shadow: 0 24px 80px rgba(0,0,0,.6) !important;
}
[data-theme="dark"] .gs-input-row {
  border-color: #334155 !important;
}
[data-theme="dark"] .gs-input {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .gs-item:hover,
[data-theme="dark"] .gs-active {
  background: #24304a !important;
}
[data-theme="dark"] .gs-item-label {
  color: #e2e8f0 !important;
}
[data-theme="dark"] .gs-item-label mark {
  background: rgba(250,204,21,.25) !important;
  color: #fde047 !important;
}
[data-theme="dark"] .gs-item-group,
[data-theme="dark"] .gs-group-label,
[data-theme="dark"] .gs-hint-title,
[data-theme="dark"] .gs-empty {
  color: #64748b !important;
}
[data-theme="dark"] .gs-footer {
  border-color: #334155 !important;
  color: #64748b !important;
}
[data-theme="dark"] .gs-kbd {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}
</style>

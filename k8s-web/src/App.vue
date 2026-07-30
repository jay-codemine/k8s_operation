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

/* ===== 暗色模式全局覆盖 ===== */
[data-theme="dark"] .sidebar,
[data-theme="dark"] .top-nav,
[data-theme="dark"] .stat-card,
[data-theme="dark"] .panel,
[data-theme="dark"] .metric-card,
[data-theme="dark"] .overview-card,
[data-theme="dark"] .component-card,
[data-theme="dark"] .arco-table,
[data-theme="dark"] .arco-modal,
[data-theme="dark"] .arco-card,
[data-theme="dark"] .arco-form-item {
  background: var(--bg-card);
  border-color: var(--border-color);
  color: var(--text-primary);
}

[data-theme="dark"] .top-nav {
  background: var(--bg-secondary);
  border-bottom-color: var(--border-color);
}

[data-theme="dark"] .sidebar {
  background: var(--bg-secondary);
  border-right-color: var(--border-color);
}

[data-theme="dark"] input,
[data-theme="dark"] select,
[data-theme="dark"] textarea,
[data-theme="dark"] .arco-input,
[data-theme="dark"] .arco-select-view {
  background: var(--input-bg);
  color: var(--text-primary);
  border-color: var(--border-color);
}

[data-theme="dark"] .arco-table th {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

[data-theme="dark"] .arco-table td {
  border-bottom-color: var(--border-color);
}

[data-theme="dark"] .stat-value,
[data-theme="dark"] .metric-val,
[data-theme="dark"] h1, [data-theme="dark"] h2, [data-theme="dark"] h3 {
  color: var(--text-primary);
}

[data-theme="dark"] .stat-label,
[data-theme="dark"] .metric-lbl,
[data-theme="dark"] .sys-name {
  color: var(--text-muted);
}

[data-theme="dark"] .panel-header,
[data-theme="dark"] .card-header {
  border-bottom-color: var(--border-color);
}

[data-theme="dark"] .btn-refresh {
  background: var(--accent);
}
[data-theme="dark"] .btn-refresh:hover {
  background: var(--accent-hover);
}

[data-theme="dark"]::-webkit-scrollbar-thumb {
  background: #334155;
}
</style>

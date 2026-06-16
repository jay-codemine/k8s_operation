<template>
  <div class="yaml-highlight-wrapper" :class="{ 'with-line-numbers': showLineNumbers }">
    <div class="yaml-toolbar" v-if="showToolbar">
      <div class="yaml-title">
        <span class="yaml-dot yaml-dot-red"></span>
        <span class="yaml-dot yaml-dot-yellow"></span>
        <span class="yaml-dot yaml-dot-green"></span>
        <span class="yaml-filename">{{ title || 'YAML' }}</span>
      </div>
      <div class="yaml-actions">
        <button class="yaml-action-btn" @click="copyToClipboard" :class="{ copied: copied }">
          <span v-if="copied">✓ 已复制</span>
          <span v-else>📋 复制</span>
        </button>
        <button class="yaml-action-btn" @click="downloadYaml" v-if="allowDownload">⬇️ 下载</button>
      </div>
    </div>
    <div class="yaml-body" :style="{ maxHeight: maxHeight }">
      <div v-if="showLineNumbers" class="line-numbers">
        <div v-for="n in lineCount" :key="n" class="line-number">{{ n }}</div>
      </div>
      <pre class="yaml-code" v-html="highlightedHtml"></pre>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  content: { type: String, default: '' },
  title: { type: String, default: '' },
  showLineNumbers: { type: Boolean, default: true },
  showToolbar: { type: Boolean, default: true },
  allowDownload: { type: Boolean, default: false },
  maxHeight: { type: String, default: '500px' },
  filename: { type: String, default: 'resource.yaml' }
})

const copied = ref(false)

const lineCount = computed(() => {
  return props.content ? props.content.split('\n').length : 1
})

function escapeHtml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function highlightYaml(yaml) {
  if (!yaml) return '<span class="yaml-placeholder">暂无内容...</span>'
  
  const lines = yaml.split('\n')
  return lines.map((line) => {
    const escaped = escapeHtml(line)
    
    // 空行
    if (!line.trim()) return `<div class="yaml-line">${escaped}</div>`
    
    // 注释行
    if (/^\s*#/.test(line)) {
      return `<div class="yaml-line"><span class="yaml-comment">${escaped}</span></div>`
    }
    
    // 文档分隔符
    if (/^---/.test(line.trim())) {
      return `<div class="yaml-line"><span class="yaml-separator">${escaped}</span></div>`
    }
    
    // 行内高亮处理
    let highlighted = escaped
      // 键（冒号前的内容）
      .replace(/^(\s*)([\w\-./]+)(:\s)/, '$1<span class="yaml-key">$2</span><span class="yaml-colon">$3</span>')
      // 布尔值
      .replace(/(:\s*)(true|false|yes|no|on|off)(\s*$|\s*#)/gi, '$1<span class="yaml-boolean">$2</span>$3')
      // 数字
      .replace(/(:\s*)(\d+)(\s*$|\s*#)/g, '$1<span class="yaml-number">$2</span>$3')
      // 字符串值（引号包裹）
      .replace(/(:\s*)("(?:[^"\\]|\\.)*")/g, '$1<span class="yaml-string">$2</span>')
      .replace(/(:\s*)('(?:[^'\\]|\\.)*')/g, '$1<span class="yaml-string">$2</span>')
      // 未引号字符串值（简单匹配）
      .replace(/(:\s+)([^\s#][^#]*?)(\s*#|$)/, (match, p1, p2, p3) => {
        if (/^(true|false|yes|no|on|off|\d+)$/i.test(p2.trim())) return match
        if (p2.startsWith('"') || p2.startsWith("'")) return match
        return `${p1}<span class="yaml-value">${p2}</span>${p3}`
      })
      // 列表标记 - 或 *
      .replace(/^(\s*)-(\s)/, '$1<span class="yaml-bullet">-</span>$2')
      // 行内注释
      .replace(/(\s+)(#.*)$/, '$1<span class="yaml-comment">$2</span>')
    
    return `<div class="yaml-line">${highlighted}</div>`
  }).join('')
}

const highlightedHtml = computed(() => highlightYaml(props.content))

function copyToClipboard() {
  navigator.clipboard.writeText(props.content).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  })
}

function downloadYaml() {
  const blob = new Blob([props.content], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = props.filename
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
.yaml-highlight-wrapper {
  background: #1e1e2e;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.06);
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', Consolas, Monaco, monospace;
  font-size: 13px;
  line-height: 1.7;
}

.yaml-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: rgba(0, 0, 0, 0.25);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.yaml-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.yaml-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.yaml-dot-red { background: #ff5f56; }
.yaml-dot-yellow { background: #ffbd2e; }
.yaml-dot-green { background: #27c93f; }

.yaml-filename {
  color: #a0a3bd;
  font-size: 12px;
  margin-left: 8px;
  font-weight: 500;
}

.yaml-actions {
  display: flex;
  gap: 6px;
}

.yaml-action-btn {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cdd6f4;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.yaml-action-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.2);
}

.yaml-action-btn.copied {
  background: rgba(16, 185, 129, 0.2);
  border-color: rgba(16, 185, 129, 0.3);
  color: #10b981;
}

.yaml-body {
  display: flex;
  overflow: auto;
  max-height: 500px;
}

.yaml-body::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.yaml-body::-webkit-scrollbar-track {
  background: transparent;
}

.yaml-body::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 4px;
}

.yaml-body::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}

.line-numbers {
  background: rgba(0, 0, 0, 0.2);
  padding: 12px 8px 12px 12px;
  text-align: right;
  color: #6c7086;
  font-size: 12px;
  user-select: none;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.line-number {
  min-width: 20px;
  line-height: 1.7;
}

.yaml-code {
  margin: 0;
  padding: 12px 16px;
  color: #cdd6f4;
  white-space: pre;
  overflow: visible;
  flex: 1;
  tab-size: 2;
}

.yaml-placeholder {
  color: #6c7086;
  font-style: italic;
}

:deep(.yaml-line) {
  min-height: 1.7em;
}

:deep(.yaml-key) {
  color: #89b4fa;
  font-weight: 600;
}

:deep(.yaml-colon) {
  color: #6c7086;
}

:deep(.yaml-value) {
  color: #a6e3a1;
}

:deep(.yaml-string) {
  color: #f9e2af;
}

:deep(.yaml-number) {
  color: #fab387;
}

:deep(.yaml-boolean) {
  color: #cba6f7;
}

:deep(.yaml-comment) {
  color: #6c7086;
  font-style: italic;
}

:deep(.yaml-bullet) {
  color: #f38ba8;
  font-weight: 700;
}

:deep(.yaml-separator) {
  color: #f38ba8;
  font-weight: 700;
}
</style>

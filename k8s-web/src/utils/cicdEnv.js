// src/utils/cicdEnv.js
// CI/CD 环境统一规范：颜色 / 标签 / 通用格式化
// 全站（AppCenter / Pipelines / PromotionPanel）共用，避免颜色散落各处

// 环境颜色规范（大厂 DevOps 视觉规范）
//   Dev 绿 / Test 蓝 / Staging 橙 / Prod 红
export const ENV_COLORS = {
  dev: '#10b981',
  development: '#10b981',
  test: '#3b82f6',
  qa: '#3b82f6',
  staging: '#f59e0b',
  pre: '#f59e0b',
  prod: '#ef4444',
  production: '#ef4444',
}

// 环境中文标签
export const ENV_LABELS = {
  dev: '开发',
  development: '开发',
  test: '测试',
  qa: '测试',
  staging: '预发',
  pre: '预发',
  prod: '生产',
  production: '生产',
}

// 默认环境顺序（后端未返回环境列表时的兜底）
export const DEFAULT_ENVS = [
  { key: 'dev', label: '开发', color: ENV_COLORS.dev },
  { key: 'test', label: '测试', color: ENV_COLORS.test },
  { key: 'staging', label: '预发', color: ENV_COLORS.staging },
  { key: 'prod', label: '生产', color: ENV_COLORS.prod },
]

// 根据环境标识归一化到 dev/test/staging/prod
export function normalizeEnv(env) {
  const e = (env || '').toString().toLowerCase()
  if (e.includes('prod')) return 'prod'
  if (e.includes('stag') || e === 'pre') return 'staging'
  if (e.includes('test') || e === 'qa') return 'test'
  if (e.includes('dev')) return 'dev'
  return e
}

// 获取环境颜色，未命中返回中性灰
export function getEnvColor(env) {
  const key = (env || '').toString().toLowerCase()
  return ENV_COLORS[key] || ENV_COLORS[normalizeEnv(env)] || '#86909c'
}

// 获取环境中文标签
export function getEnvLabel(env) {
  const key = (env || '').toString().toLowerCase()
  return ENV_LABELS[key] || ENV_LABELS[normalizeEnv(env)] || env || '-'
}

// 根据命名空间猜测环境（兼容旧数据：流水线只有单个 target_namespace）
export function guessEnvFromNamespace(ns) {
  return normalizeEnv(ns)
}

// 格式化耗时：320 -> 5m20s，3720 -> 1h2m0s
export function formatDuration(seconds) {
  const s = Number(seconds) || 0
  if (s <= 0) return '-'
  if (s >= 3600) {
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = s % 60
    return `${h}h${m}m${sec}s`
  }
  if (s >= 60) {
    const m = Math.floor(s / 60)
    const sec = s % 60
    return `${m}m${sec}s`
  }
  return `${s}s`
}

// 从镜像地址提取标签：registry/app:tag -> tag
export function extractImageTag(image) {
  if (!image) return ''
  let img = image
  const at = img.indexOf('@')
  if (at > 0) img = img.slice(0, at)
  const colon = img.lastIndexOf(':')
  const slash = img.lastIndexOf('/')
  if (colon > slash) return img.slice(colon + 1)
  return 'latest'
}

// 截取提交短哈希（7 位）
export function shortCommit(commit) {
  if (!commit) return ''
  return commit.slice(0, 7)
}

export default {
  ENV_COLORS,
  ENV_LABELS,
  DEFAULT_ENVS,
  normalizeEnv,
  getEnvColor,
  getEnvLabel,
  guessEnvFromNamespace,
  formatDuration,
  extractImageTag,
  shortCommit,
}

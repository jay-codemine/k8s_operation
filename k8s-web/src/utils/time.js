// src/utils/time.js — 统一时区工具，强制 Asia/Shanghai

const TZ = 'Asia/Shanghai'

const pad = n => String(n).padStart(2, '0')

/**
 * 格式化 Unix 时间戳（秒或毫秒）为 "YYYY-MM-DD HH:mm"
 */
export function formatDateTime(ts) {
  if (!ts) return '-'
  const t = ts > 1e11 ? ts : ts * 1000
  const d = new Date(t)
  const parts = new Intl.DateTimeFormat('zh-CN', {
    timeZone: TZ, year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', hour12: false,
  }).formatToParts(d)
  const m = {}
  parts.forEach(p => { m[p.type] = p.value })
  return `${m.year}-${m.month}-${m.day} ${m.hour}:${m.minute}`
}

/**
 * 相对时间：刚刚 / X分钟前 / X小时前 / 完整日期
 */
export function formatRelative(ts) {
  if (!ts) return '-'
  const t = ts > 1e11 ? ts : ts * 1000
  const d = new Date(t)
  const now = Date.now()
  const diff = now - t
  if (diff < 0) return new Intl.DateTimeFormat('zh-CN', { timeZone: TZ, year: 'numeric', month: '2-digit', day: '2-digit' }).format(d)
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return formatDateTime(ts)
}

/**
 * 完整日期时间（用于详情页）
 */
export function formatFull(ts) {
  if (!ts) return '-'
  const t = ts > 1e11 ? ts : ts * 1000
  return new Date(t).toLocaleString('zh-CN', { timeZone: TZ, hour12: false })
}

// 通用格式化

// 时间: YYYY-MM-DD HH:mm:ss; 空值、无效值和零值时间显示为 -
export const formatDateTime = (value?: string | null) => {
  if (!value) return '-'
  const d = new Date(value)
  if (isNaN(d.getTime()) || d.getFullYear() < 1971) return '-'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

// 标签 / 选择器: k1=v1, k2=v2
export const formatLabels = (m?: Record<string, string> | null) =>
  Object.entries(m || {})
    .map(([k, v]) => `${k}=${v}`)
    .join(', ')

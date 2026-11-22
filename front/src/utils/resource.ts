// CPU / 内存的格式化; 后端返回的 CPU 单位为毫核, 内存单位为字节

export interface PodResources {
  cpuRequest: number
  // 0 表示不限制
  cpuLimit: number
  memoryRequest: number
  memoryLimit: number
}

export interface PodUsage {
  // null 表示没有数据(未安装 metrics-server 或容器组未运行)
  cpuUsage?: number | null
  memoryUsage?: number | null
}

const trim = (n: number, digits: number) => String(Number(n.toFixed(digits)))

// 250 -> 250m, 1500 -> 1.5 核
export const formatCpu = (milli?: number | null) => {
  if (milli === undefined || milli === null) return '-'
  if (milli < 1000) return `${milli}m`
  return `${trim(milli / 1000, 2)} 核`
}

// 字节 -> Ki / Mi / Gi
export const formatMemory = (bytes?: number | null) => {
  if (bytes === undefined || bytes === null) return '-'
  const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti']
  let value = bytes
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${trim(value, value >= 100 ? 0 : 1)}${units[i]}`
}

export type UsageKind = 'cpu' | 'memory'

export const formatUsage = (kind: UsageKind, value?: number | null) => (kind === 'cpu' ? formatCpu(value) : formatMemory(value))

// 用量占上限(没有上限时占预留)的百分比; 都没有时返回 null
export const usagePercent = (used: number | null | undefined, request: number, limit: number) => {
  if (used === undefined || used === null) return null
  const base = limit || request
  if (!base) return null
  return Math.round((used / base) * 100)
}

// Kubernetes 数量(例如 500m、2、1.5Gi、100M)转换为数字; CPU 结果单位为核, 内存为字节; 无法解析时返回 null
const SUFFIX: Record<string, number> = {
  n: 1e-9,
  u: 1e-6,
  m: 1e-3,
  '': 1,
  k: 1e3,
  M: 1e6,
  G: 1e9,
  T: 1e12,
  P: 1e15,
  E: 1e18,
  Ki: 2 ** 10,
  Mi: 2 ** 20,
  Gi: 2 ** 30,
  Ti: 2 ** 40,
  Pi: 2 ** 50,
  Ei: 2 ** 60
}

export const QUANTITY_PATTERN = /^([+-]?\d+(?:\.\d+)?)(?:([eE][+-]?\d+)|(Ki|Mi|Gi|Ti|Pi|Ei|n|u|m|k|M|G|T|P|E)?)$/

export function parseQuantity(value?: string | number | null): number | null {
  if (value === undefined || value === null || value === '') return null
  if (typeof value === 'number') return value
  const m = String(value).trim().match(QUANTITY_PATTERN)
  if (!m) return null
  const base = Number(m[1])
  if (m[2]) return base * 10 ** Number(m[2].slice(1))
  return base * (SUFFIX[m[3] || ''] ?? 1)
}

export const isQuantity = (value: string) => QUANTITY_PATTERN.test(value.trim())


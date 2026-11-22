// 资源配额(ResourceQuota)与默认资源限制(LimitRange)
import { parseQuantity } from '@/utils/resource'

export interface QuotaItem {
  key: string
  label: string
  placeholder: string
}

export const QUOTA_GROUPS: { title: string; items: QuotaItem[] }[] = [
  {
    title: '计算资源',
    items: [
      { key: 'requests.cpu', label: 'CPU 预留总量', placeholder: '如 4 或 4000m' },
      { key: 'limits.cpu', label: 'CPU 上限总量', placeholder: '如 8' },
      { key: 'requests.memory', label: '内存预留总量', placeholder: '如 8Gi' },
      { key: 'limits.memory', label: '内存上限总量', placeholder: '如 16Gi' }
    ]
  },
  {
    title: '存储',
    items: [
      { key: 'requests.storage', label: '存储容量总量', placeholder: '如 100Gi' },
      { key: 'persistentvolumeclaims', label: '持久卷声明数量', placeholder: '如 10' }
    ]
  },
  {
    title: '对象数量',
    items: [
      { key: 'pods', label: '容器组', placeholder: '如 50' },
      { key: 'services', label: '服务', placeholder: '如 20' },
      { key: 'services.nodeports', label: 'NodePort 服务', placeholder: '如 5' },
      { key: 'services.loadbalancers', label: 'LoadBalancer 服务', placeholder: '如 2' },
      { key: 'configmaps', label: 'ConfigMap', placeholder: '如 50' },
      { key: 'secrets', label: 'Secret', placeholder: '如 50' },
      { key: 'count/deployments.apps', label: 'Deployment', placeholder: '如 20' },
      { key: 'count/statefulsets.apps', label: 'StatefulSet', placeholder: '如 10' },
      { key: 'count/jobs.batch', label: 'Job', placeholder: '如 20' },
      { key: 'count/cronjobs.batch', label: 'CronJob', placeholder: '如 10' }
    ]
  }
]

const QUOTA_TEXT: Record<string, string> = Object.fromEntries(QUOTA_GROUPS.flatMap(g => g.items.map(i => [i.key, i.label])))
Object.assign(QUOTA_TEXT, { cpu: 'CPU 预留总量', memory: '内存预留总量', 'requests.ephemeral-storage': '临时存储预留总量', 'limits.ephemeral-storage': '临时存储上限总量' })

export const quotaText = (key: string) => {
  if (QUOTA_TEXT[key]) return QUOTA_TEXT[key]
  const m = key.match(/^(.+)\.storageclass\.storage\.k8s\.io\/(requests\.storage|persistentvolumeclaims)$/)
  if (m) return `存储类 ${m[1]} 的${m[2] === 'requests.storage' ? '存储容量' : '持久卷声明数量'}`
  return key
}

export interface QuotaUsageRow {
  key: string
  label: string
  used: string
  hard: string
  percent: number | null
}

// 配额用量: 按 hard 中的资源逐项列出
export function quotaUsageRows(hard: Record<string, string> = {}, used: Record<string, string> = {}): QuotaUsageRow[] {
  return Object.keys(hard)
    .sort((a, b) => quotaOrder(a) - quotaOrder(b) || a.localeCompare(b))
    .map(key => {
      const h = parseQuantity(hard[key])
      const u = parseQuantity(used[key] ?? '0')
      return {
        key,
        label: quotaText(key),
        used: used[key] ?? '0',
        hard: hard[key],
        percent: h && u !== null ? Math.round((u / h) * 100) : h === 0 ? 100 : null
      }
    })
}

const ORDER = QUOTA_GROUPS.flatMap(g => g.items.map(i => i.key))
const quotaOrder = (key: string) => {
  const i = ORDER.indexOf(key)
  return i < 0 ? ORDER.length : i
}

export const LIMIT_TYPE_TEXT: Record<string, string> = {
  Container: '容器',
  Pod: '容器组',
  PersistentVolumeClaim: '持久卷声明'
}

export const LIMIT_FIELDS: { key: 'min' | 'max' | 'defaultRequest' | 'default' | 'maxLimitRequestRatio'; label: string }[] = [
  { key: 'defaultRequest', label: '默认预留' },
  { key: 'default', label: '默认上限' },
  { key: 'min', label: '最小值' },
  { key: 'max', label: '最大值' },
  { key: 'maxLimitRequestRatio', label: '上限 / 预留最大比例' }
]

export interface LimitRow {
  type: string
  resource: string
  values: Record<string, string>
}

// LimitRange 的每一项按资源展开为行
export function limitRows(limits: Record<string, any>[] = []): LimitRow[] {
  const rows: LimitRow[] = []
  limits.forEach(item => {
    const resources = new Set<string>()
    LIMIT_FIELDS.forEach(f => Object.keys(item[f.key] || {}).forEach(r => resources.add(r)))
    resources.forEach(resource => {
      const values: Record<string, string> = {}
      LIMIT_FIELDS.forEach(f => {
        if (item[f.key]?.[resource] !== undefined) values[f.key] = item[f.key][resource]
      })
      rows.push({ type: item.type, resource, values })
    })
  })
  return rows
}

export const RESOURCE_NAME_TEXT: Record<string, string> = { cpu: 'CPU', memory: '内存', storage: '存储', 'ephemeral-storage': '临时存储' }

// 平台创建的对象名称
export const DEFAULT_QUOTA_NAME = 'default-quota'
export const DEFAULT_LIMIT_RANGE_NAME = 'default-limits'

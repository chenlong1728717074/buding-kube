// 节点相关的常量与展示
type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

export const NODE_STATUS: Record<string, { text: string; tag: TagType }> = {
  Ready: { text: '就绪', tag: 'success' },
  NotReady: { text: '未就绪', tag: 'danger' },
  Unknown: { text: '未知', tag: 'warning' }
}

export const nodeStatusText = (s?: string) => NODE_STATUS[s || '']?.text || s || '-'
export const nodeStatusTag = (s?: string): TagType => NODE_STATUS[s || '']?.tag || 'info'

export const roleText = (role: string) =>
  (({ master: '控制节点', 'control-plane': '控制节点', worker: '工作节点', edge: '边缘节点' }) as Record<string, string>)[role] || role

// 与后端 IsKubeletLabel 保持一致: 由 kubelet / 云厂商组件维护, 编辑时保留
const KUBELET_LABELS = new Set([
  'kubernetes.io/hostname',
  'kubernetes.io/os',
  'kubernetes.io/arch',
  'beta.kubernetes.io/os',
  'beta.kubernetes.io/arch',
  'node.kubernetes.io/instance-type',
  'beta.kubernetes.io/instance-type',
  'topology.kubernetes.io/region',
  'topology.kubernetes.io/zone',
  'failure-domain.beta.kubernetes.io/region',
  'failure-domain.beta.kubernetes.io/zone'
])

export const isKubeletLabel = (key: string) => KUBELET_LABELS.has(key) || key.startsWith('feature.node.kubernetes.io/')

export const TAINT_EFFECTS = [
  { value: 'NoSchedule', label: '不调度（NoSchedule）', desc: '不容忍该污点的容器组不会调度到节点，已运行的不受影响' },
  { value: 'PreferNoSchedule', label: '尽量不调度（PreferNoSchedule）', desc: '调度器尽量避开该节点，资源不足时仍可能调度' },
  { value: 'NoExecute', label: '不调度并驱逐（NoExecute）', desc: '不调度，且会立即驱逐节点上不容忍该污点的容器组' }
]

export const taintEffectText = (effect?: string) => TAINT_EFFECTS.find(e => e.value === effect)?.label || effect || '-'

export const taintText = (t: { key: string; value?: string; effect: string }) => `${t.key}${t.value ? `=${t.value}` : ''}:${t.effect}`

// 节点状态条件: Ready 为 True 时正常, 其他压力类条件为 False 时正常
export const NODE_CONDITION_TEXT: Record<string, string> = {
  Ready: '节点就绪（Ready）',
  MemoryPressure: '内存压力（MemoryPressure）',
  DiskPressure: '磁盘压力（DiskPressure）',
  PIDPressure: '进程数压力（PIDPressure）',
  NetworkUnavailable: '网络不可用（NetworkUnavailable）'
}

export const nodeConditionHealthy = (type: string, status: string) => (type === 'Ready' ? status === 'True' : status !== 'True')

export const percentOf = (value?: number | null, total?: number) => {
  if (value === undefined || value === null || !total) return null
  return Math.round((value / total) * 100)
}

export const percentLevel = (p: number | null) => (p === null ? '' : p >= 90 ? 'is-danger' : p >= 70 ? 'is-warning' : 'is-ok')

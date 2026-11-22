// 通用资源列表的列名翻译与单元格格式化
import type { ResourceTableColumn } from '@/api/resource'

// 行操作菜单中的额外操作
export interface ResourceAction {
  command: string
  label: string
}

const COLUMN_TEXT: Record<string, string> = {
  Ready: '就绪',
  'Up-to-date': '已更新',
  Available: '可用',
  Desired: '期望',
  Current: '当前',
  Containers: '容器',
  Images: '镜像',
  Selector: '选择器',
  Reference: '目标',
  Targets: '指标（当前/目标）',
  MinPods: '最小副本',
  MaxPods: '最大副本',
  Replicas: '当前副本',
  'Pod-Selector': '容器组选择器',
  Request: '预留配额（已用/上限）',
  Limit: '上限配额（已用/上限）',
  'Min Available': '最少可用',
  'Max Unavailable': '最多不可用',
  'Allowed Disruptions': '允许中断数',
  Value: '优先级值',
  'Global-Default': '全局默认',
  'Preemption Policy': '抢占策略',
  Controller: '控制器',
  Parameters: '参数',
  Status: '状态',
  Type: '类型',
  Data: '数据',
  Version: '版本',
  Kind: '类型'
}

// 由页面单独展示的列
const HIDDEN_COLUMNS = new Set(['Name', 'Namespace', 'Age', 'Created At'])

export const columnText = (c: ResourceTableColumn) => COLUMN_TEXT[c.name] || c.name

export const visibleColumns = (columns: ResourceTableColumn[], wide: boolean) =>
  columns.map((c, index) => ({ ...c, index })).filter(c => !HIDDEN_COLUMNS.has(c.name) && (wide || c.priority === 0))

export const formatCell = (value: unknown): string => {
  if (value === undefined || value === null || value === '' || value === '<none>' || value === '<unset>') return '-'
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

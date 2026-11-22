// 各类资源状态的文字和标签颜色
type TagType = 'success' | 'warning' | 'danger' | 'info' | 'primary'

const CLUSTER_STATUS: Record<string, string> = {
  Active: '运行中',
  Running: '运行中',
  Error: '异常',
  Failed: '失败',
  Pending: '等待中',
  Unknown: '未知'
}
export const clusterStatusText = (status?: string) => (status && CLUSTER_STATUS[status]) || status || '未知'

const USER_STATUS: Record<string, [string, TagType]> = {
  active: ['正常', 'success'],
  inactive: ['禁用', 'danger'],
  suspended: ['暂停', 'warning'],
  expired: ['过期', 'info']
}
export const userStatusText = (status?: string) => (status && USER_STATUS[status]?.[0]) || status || '未知'
export const userStatusType = (status?: string): TagType => (status && USER_STATUS[status]?.[1]) || 'info'

const NAMESPACE_STATUS: Record<string, [string, TagType]> = {
  Active: ['活跃', 'success'],
  Terminating: ['删除中', 'warning']
}
export const namespaceStatusText = (status?: string) => (status && NAMESPACE_STATUS[status]?.[0]) || status || '-'
export const namespaceStatusType = (status?: string): TagType => (status && NAMESPACE_STATUS[status]?.[1]) || 'danger'

// PV / PVC 的 phase
const VOLUME_PHASE: Record<string, TagType> = {
  Available: 'primary',
  Bound: 'success',
  Released: 'warning',
  Pending: 'warning',
  Failed: 'danger',
  Lost: 'danger'
}
export const volumePhaseType = (phase?: string): TagType => (phase && VOLUME_PHASE[phase]) || 'info'

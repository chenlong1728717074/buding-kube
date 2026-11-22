import type { RouteLocationRaw } from 'vue-router'

// 资源在本平台的详情页; 没有详情页时返回 null
export function objectDetailRoute(clusterId: string, kind?: string, namespace?: string, name?: string): RouteLocationRaw | null {
  if (!kind || !name) return null
  const base = `/cluster/${clusterId}`
  const ns = namespace || ''
  const namespaced: Record<string, string> = {
    Pod: 'pod/detail',
    Deployment: 'workload/deployment/detail',
    StatefulSet: 'workload/statefulset/detail',
    DaemonSet: 'workload/daemonset/detail',
    Job: 'job/job/detail',
    CronJob: 'job/cronjob/detail',
    Service: 'service/service/detail',
    Ingress: 'service/ingress/detail',
    ConfigMap: 'config/configmap/detail',
    Secret: 'config/secret/detail',
    ServiceAccount: 'config/serviceaccount/detail'
  }
  if (namespaced[kind] && ns) return { path: `${base}/${namespaced[kind]}`, query: { namespace: ns, name } }
  if (kind === 'Node') return { path: `${base}/node/detail`, query: { hostname: name } }
  if (kind === 'Namespace') return { path: `${base}/namespace/detail`, query: { namespace: name } }
  if (kind === 'StorageClass') return { path: `${base}/storage/storageclass/detail`, query: { name } }
  return null
}

// 由时间生成「3 分钟前」这样的文本
export function timeAgo(value?: string | null) {
  if (!value) return '-'
  const t = new Date(value).getTime()
  if (isNaN(t)) return '-'
  const seconds = Math.max(0, Math.round((Date.now() - t) / 1000))
  if (seconds < 60) return `${seconds} 秒前`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  return `${Math.floor(seconds / 86400)} 天前`
}

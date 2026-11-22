// Job / CronJob 共用的常量、格式化和表单工具
import type { JobContainerDTO, JobStatus, JobTemplateDTO } from '@/api/job'
import type { ConcurrencyPolicy } from '@/api/cronjob'

type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

export interface KVRow {
  key: string
  value: string
}

export const JOB_STATUS: Record<JobStatus, { text: string; tag: TagType }> = {
  Pending: { text: '等待中', tag: 'warning' },
  Running: { text: '运行中', tag: 'primary' },
  Complete: { text: '已完成', tag: 'success' },
  Failed: { text: '失败', tag: 'danger' },
  Suspended: { text: '已暂停', tag: 'info' }
}

export const JOB_STATUS_OPTIONS = (Object.keys(JOB_STATUS) as JobStatus[]).map(key => ({ value: key, label: JOB_STATUS[key].text }))

export const jobStatusText = (status?: string) => JOB_STATUS[status as JobStatus]?.text || status || '-'
export const jobStatusTag = (status?: string): TagType => JOB_STATUS[status as JobStatus]?.tag || 'info'

// 容器组状态
export const podStatusTag = (status?: string): TagType => {
  if (!status) return 'info'
  if (status === 'Running') return 'primary'
  if (status === 'Succeeded' || status === 'Completed') return 'success'
  if (status === 'Pending' || status === 'ContainerCreating' || status === 'PodInitializing' || status === 'Terminating') return 'warning'
  if (/Failed|Error|BackOff|Err|OOM|Invalid|DeadlineExceeded|Evicted/i.test(status)) return 'danger'
  return 'info'
}

export const podStatusText = (status?: string) =>
  (({ Running: '运行中', Succeeded: '已完成', Completed: '已完成', Pending: '等待中', Failed: '失败', Terminating: '删除中' }) as Record<string, string>)[
    status || ''
  ] || status || '-'

export const CONCURRENCY_POLICIES: { value: ConcurrencyPolicy; label: string; desc: string }[] = [
  { value: 'Allow', label: '允许并发', desc: '上一次还没结束时，照常启动新的任务' },
  { value: 'Forbid', label: '禁止并发', desc: '上一次还没结束时，跳过这次调度' },
  { value: 'Replace', label: '替换', desc: '上一次还没结束时，停止它并启动新的任务' }
]

export const concurrencyText = (policy?: string) => CONCURRENCY_POLICIES.find(p => p.value === policy)?.label || policy || '-'

export const SCHEDULE_PRESETS = [
  { label: '每 5 分钟', value: '*/5 * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天 0 点', value: '0 0 * * *' },
  { label: '每天 2 点', value: '0 2 * * *' },
  { label: '每周一 0 点', value: '0 0 * * 1' },
  { label: '每月 1 日 0 点', value: '0 0 1 * *' }
]

export const TIMEZONE_OPTIONS = ['Asia/Shanghai', 'UTC', 'Asia/Hong_Kong', 'Asia/Tokyo', 'Asia/Singapore', 'Europe/London', 'America/New_York', 'America/Chicago', 'America/Los_Angeles']

const WEEKDAYS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六', '周日']
const MACROS: Record<string, string> = {
  '@yearly': '每年 1 月 1 日 00:00',
  '@annually': '每年 1 月 1 日 00:00',
  '@monthly': '每月 1 日 00:00',
  '@weekly': '每周日 00:00',
  '@daily': '每天 00:00',
  '@midnight': '每天 00:00',
  '@hourly': '每小时整点'
}

const isNum = (s: string) => /^\d+$/.test(s)
const pad = (s: string) => s.padStart(2, '0')

// 常见 Cron 表达式的中文说明; 识别不了时返回空字符串
export function describeCron(expr?: string): string {
  const value = (expr || '').trim()
  if (!value) return ''
  if (MACROS[value]) return MACROS[value]
  if (value.startsWith('@every ')) return `每隔 ${value.slice(7)}`
  const f = value.split(/\s+/)
  if (f.length !== 5) return ''
  const [min, hour, dom, mon, dow] = f
  const anyDate = dom === '*' && mon === '*'

  if (/^\*\/\d+$/.test(min) && hour === '*' && anyDate && dow === '*') return `每 ${min.slice(2)} 分钟`
  if (min === '*' && hour === '*' && anyDate && dow === '*') return '每分钟'
  if (isNum(min) && hour === '*' && anyDate && dow === '*') return `每小时第 ${min} 分钟`
  if (isNum(min) && /^\*\/\d+$/.test(hour) && anyDate && dow === '*') return `每 ${hour.slice(2)} 小时（第 ${min} 分钟）`
  if (!isNum(min) || !/^\d+(,\d+)*$/.test(hour)) return ''

  const time = hour.split(',').map(h => `${pad(h)}:${pad(min)}`).join('、')
  if (anyDate && dow === '*') return `每天 ${time}`
  if (anyDate && /^[0-7](,[0-7])*$/.test(dow)) return `每${dow.split(',').map(d => WEEKDAYS[Number(d)]).join('、')} ${time}`
  if (anyDate && dow === '1-5') return `每个工作日 ${time}`
  if (mon === '*' && dow === '*' && /^\d+(,\d+)*$/.test(dom)) return `每月 ${dom.split(',').join('、')} 日 ${time}`
  if (isNum(mon) && isNum(dom) && dow === '*') return `每年 ${mon} 月 ${dom} 日 ${time}`
  return ''
}

// 2026-11-17 00:04:41
export const formatDateTime = (value?: string | null) => {
  if (!value) return '-'
  const d = new Date(value)
  if (isNaN(d.getTime()) || d.getFullYear() < 1971) return '-'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

// 耗时: 未结束时计算到现在
export const formatDuration = (start?: string | null, end?: string | null) => {
  if (!start) return '-'
  const begin = new Date(start).getTime()
  const finish = end ? new Date(end).getTime() : Date.now()
  if (isNaN(begin) || isNaN(finish) || finish < begin) return '-'
  let seconds = Math.round((finish - begin) / 1000)
  const days = Math.floor(seconds / 86400)
  seconds -= days * 86400
  const hours = Math.floor(seconds / 3600)
  seconds -= hours * 3600
  const minutes = Math.floor(seconds / 60)
  seconds -= minutes * 60
  if (days) return `${days}天${hours}小时`
  if (hours) return `${hours}小时${minutes}分`
  if (minutes) return `${minutes}分${seconds}秒`
  return `${seconds}秒`
}

// 秒数显示为易读文本
export const formatSeconds = (value?: number | null) => {
  if (value === undefined || value === null) return '-'
  if (value < 60) return `${value} 秒`
  if (value % 86400 === 0) return `${value} 秒（${value / 86400} 天）`
  if (value % 3600 === 0) return `${value} 秒（${value / 3600} 小时）`
  if (value % 60 === 0) return `${value} 秒（${value / 60} 分钟）`
  return `${value} 秒`
}

// ---------- 创建表单 ----------

export type CommandMode = 'default' | 'shell' | 'exec'

export interface ContainerForm {
  name: string
  image: string
  imagePullPolicy: string
  commandMode: CommandMode
  // Shell 模式下的脚本
  script: string
  // 命令与参数模式下, 每行一项
  commandText: string
  argsText: string
  workingDir: string
  env: KVRow[]
  cpuRequest: string
  cpuLimit: string
  memoryRequest: string
  memoryLimit: string
}

export interface JobSettingsForm {
  completions: number
  parallelism: number
  backoffLimit: number
  // 0 表示不限制
  activeDeadlineSeconds: number
  autoClean: boolean
  ttlSecondsAfterFinished: number
  restartPolicy: 'Never' | 'OnFailure'
}

export const createContainerForm = (index = 0): ContainerForm => ({
  name: index === 0 ? 'main' : `container-${index + 1}`,
  image: '',
  imagePullPolicy: 'IfNotPresent',
  commandMode: 'shell',
  script: '',
  commandText: '',
  argsText: '',
  workingDir: '',
  env: [],
  cpuRequest: '',
  cpuLimit: '',
  memoryRequest: '',
  memoryLimit: ''
})

export const createJobSettingsForm = (): JobSettingsForm => ({
  completions: 1,
  parallelism: 1,
  backoffLimit: 6,
  activeDeadlineSeconds: 0,
  autoClean: false,
  ttlSecondsAfterFinished: 3600,
  restartPolicy: 'Never'
})

const lines = (text: string) =>
  text
    .split('\n')
    .map(s => s.trim())
    .filter(Boolean)

export function containerFormToDTO(c: ContainerForm): JobContainerDTO {
  const dto: JobContainerDTO = {
    name: c.name.trim(),
    image: c.image.trim(),
    imagePullPolicy: c.imagePullPolicy || undefined,
    workingDir: c.workingDir.trim() || undefined,
    env: c.env.filter(e => e.key.trim()).map(e => ({ name: e.key.trim(), value: e.value })),
    cpuRequest: c.cpuRequest.trim() || undefined,
    cpuLimit: c.cpuLimit.trim() || undefined,
    memoryRequest: c.memoryRequest.trim() || undefined,
    memoryLimit: c.memoryLimit.trim() || undefined
  }
  if (c.commandMode === 'shell' && c.script.trim()) {
    dto.command = ['/bin/sh', '-c']
    dto.args = [c.script.replace(/\r\n/g, '\n').trim()]
  } else if (c.commandMode === 'exec') {
    dto.command = lines(c.commandText)
    dto.args = lines(c.argsText)
  }
  return dto
}

export function buildJobTemplateDTO(settings: JobSettingsForm, containers: ContainerForm[]): JobTemplateDTO {
  return {
    completions: settings.completions,
    parallelism: settings.parallelism,
    backoffLimit: settings.backoffLimit,
    activeDeadlineSeconds: settings.activeDeadlineSeconds > 0 ? settings.activeDeadlineSeconds : undefined,
    ttlSecondsAfterFinished: settings.autoClean ? settings.ttlSecondsAfterFinished : undefined,
    restartPolicy: settings.restartPolicy,
    containers: containers.map(containerFormToDTO)
  }
}

const compact = <T extends Record<string, unknown>>(obj: T) => {
  const result: Record<string, unknown> = {}
  Object.entries(obj).forEach(([k, v]) => {
    if (v === undefined || v === '' || (Array.isArray(v) && !v.length)) return
    if (v && typeof v === 'object' && !Array.isArray(v) && !Object.keys(v).length) return
    result[k] = v
  })
  return result
}

// 生成与后端一致的任务规格, 用于 YAML 预览
export function jobSpecPreview(t: JobTemplateDTO, podLabels?: Record<string, string>) {
  const containers = t.containers.map(c => {
    const requests = compact({ cpu: c.cpuRequest, memory: c.memoryRequest })
    const limits = compact({ cpu: c.cpuLimit, memory: c.memoryLimit })
    return compact({
      name: c.name,
      image: c.image,
      imagePullPolicy: c.imagePullPolicy,
      command: c.command,
      args: c.args,
      workingDir: c.workingDir,
      env: c.env,
      resources: compact({ requests, limits })
    })
  })
  return compact({
    completions: t.completions,
    parallelism: t.parallelism,
    backoffLimit: t.backoffLimit,
    activeDeadlineSeconds: t.activeDeadlineSeconds,
    ttlSecondsAfterFinished: t.ttlSecondsAfterFinished,
    template: compact({
      metadata: podLabels && Object.keys(podLabels).length ? { labels: podLabels } : undefined,
      spec: { restartPolicy: t.restartPolicy || 'Never', containers }
    })
  })
}

// 名称校验: 任务名会作为容器组标签值, 必须是 DNS 标签
export const DNS_LABEL = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/

export const jobYamlExample = (namespace = 'default') =>
  [
    'apiVersion: batch/v1',
    'kind: Job',
    'metadata:',
    '  name: example-job',
    `  namespace: ${namespace}`,
    'spec:',
    '  completions: 1',
    '  parallelism: 1',
    '  backoffLimit: 6',
    '  template:',
    '    spec:',
    '      restartPolicy: Never',
    '      containers:',
    '        - name: main',
    '          image: busybox:1.36',
    '          command: ["/bin/sh", "-c"]',
    '          args: ["echo hello from job && sleep 5"]'
  ].join('\n')

export const cronJobYamlExample = (namespace = 'default') =>
  [
    'apiVersion: batch/v1',
    'kind: CronJob',
    'metadata:',
    '  name: example-cronjob',
    `  namespace: ${namespace}`,
    'spec:',
    '  schedule: "*/5 * * * *"',
    '  timeZone: Asia/Shanghai',
    '  concurrencyPolicy: Forbid',
    '  successfulJobsHistoryLimit: 3',
    '  failedJobsHistoryLimit: 1',
    '  jobTemplate:',
    '    spec:',
    '      backoffLimit: 2',
    '      template:',
    '        spec:',
    '          restartPolicy: OnFailure',
    '          containers:',
    '            - name: main',
    '              image: busybox:1.36',
    '              command: ["/bin/sh", "-c"]',
    '              args: ["date; echo cron running"]'
  ].join('\n')

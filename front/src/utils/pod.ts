// 容器组的状态展示
import type { PodContainerStatusVO } from '@/api/pod'

type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

export const POD_PHASE_OPTIONS = [
  { value: 'Running', label: '运行中' },
  { value: 'Pending', label: '等待中' },
  { value: 'Succeeded', label: '已完成' },
  { value: 'Failed', label: '失败' },
  { value: 'Unknown', label: '未知' }
]

export const QOS_TEXT: Record<string, string> = {
  Guaranteed: '有保证（Guaranteed）',
  Burstable: '可突发（Burstable）',
  BestEffort: '尽力而为（BestEffort）'
}

export const RESTART_POLICY_TEXT: Record<string, string> = {
  Always: '总是重启（Always）',
  OnFailure: '失败时重启（OnFailure）',
  Never: '不重启（Never）'
}

// 容器状态: Running / Waiting(原因) / Terminated(原因, 退出码)
export const containerStateText = (s?: PodContainerStatusVO) => {
  if (!s) return '未启动'
  if (s.state === 'Running') return s.ready ? '运行中' : '运行中（未就绪）'
  if (s.state === 'Waiting') return s.reason || '等待中'
  if (s.state === 'Terminated') return `${s.reason || '已退出'}（退出码 ${s.exitCode}）`
  return '未知'
}

export const containerStateTag = (s?: PodContainerStatusVO): TagType => {
  if (!s) return 'info'
  if (s.state === 'Running') return s.ready ? 'success' : 'warning'
  if (s.state === 'Terminated') return s.exitCode === 0 ? 'info' : 'danger'
  if (s.state === 'Waiting') return /BackOff|Err|Invalid|OOM/i.test(s.reason || '') ? 'danger' : 'warning'
  return 'info'
}

export const podYamlExample = (namespace = 'default') =>
  [
    'apiVersion: v1',
    'kind: Pod',
    'metadata:',
    '  name: example-pod',
    `  namespace: ${namespace}`,
    '  labels:',
    '    app: example-pod',
    'spec:',
    '  containers:',
    '    - name: nginx',
    '      image: nginx:1.27',
    '      ports:',
    '        - containerPort: 80'
  ].join('\n')

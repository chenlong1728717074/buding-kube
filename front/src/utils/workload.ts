// 工作负载(Deployment / StatefulSet / DaemonSet)的常量、表单和 YAML 预览
import type { ProbeDTO, VolumeClaimTemplateDTO, WorkloadContainerDTO, WorkloadKind, WorkloadVolumeDTO } from '@/api/workload'
import { DNS_LABEL, containerFormToDTO, createContainerForm, type ContainerForm, type KVRow } from '@/utils/job'
import type { CapacityUnit } from '@/utils/storageclass'

type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

export interface WorkloadKindMeta {
  kind: WorkloadKind
  label: string
  // 中文说明, 与 KubeSphere 一致
  cn: string
  icon: string
  // 列表页路由(集群下的相对路径)
  listPath: string
  desc: string
}

export const WORKLOAD_KINDS: Record<WorkloadKind, WorkloadKindMeta> = {
  deployment: { kind: 'deployment', label: 'Deployment', cn: '无状态工作负载', icon: 'deploy', listPath: 'workload/deployment', desc: '管理无状态应用的部署与扩缩容' },
  statefulset: { kind: 'statefulset', label: 'StatefulSet', cn: '有状态工作负载', icon: 'sts', listPath: 'workload/statefulset', desc: '管理有状态应用的有序部署' },
  daemonset: { kind: 'daemonset', label: 'DaemonSet', cn: '守护进程集', icon: 'ds', listPath: 'workload/daemonset', desc: '管理在每个节点上运行的守护进程' }
}

export const workloadDetailPath = (clusterId: string, kind: WorkloadKind) => `/cluster/${clusterId}/${WORKLOAD_KINDS[kind].listPath}/detail`

export const workloadStatusTag = (status?: string): TagType =>
  (({ 运行中: 'success', 更新中: 'warning', 已停止: 'info', 未知: 'info' }) as Record<string, TagType>)[status || ''] || 'info'

export const WORKLOAD_STATUS_OPTIONS = ['运行中', '更新中', '已停止']

export const strategyText = (kind: WorkloadKind, strategy?: string) => {
  if (strategy === 'Recreate') return '重新创建（Recreate）'
  if (strategy === 'OnDelete') return '删除容器组时更新（OnDelete）'
  return kind === 'deployment' ? '滚动更新（RollingUpdate）' : '滚动更新（RollingUpdate）'
}

export const VOLUME_TYPES: { value: WorkloadVolumeDTO['type']; label: string; desc: string; color: string }[] = [
  { value: 'pvc', label: '持久卷声明', desc: '挂载已有的 PVC，数据持久保存', color: 'linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%)' },
  { value: 'configMap', label: 'ConfigMap', desc: '把配置文件挂载到容器中', color: 'linear-gradient(135deg, #34d399 0%, #059669 100%)' },
  { value: 'secret', label: 'Secret', desc: '把密钥、证书挂载到容器中', color: 'linear-gradient(135deg, #f472b6 0%, #db2777 100%)' },
  { value: 'emptyDir', label: '临时存储', desc: '容器组删除后数据随之删除', color: 'linear-gradient(135deg, #fbbf24 0%, #d97706 100%)' },
  { value: 'hostPath', label: '主机路径', desc: '挂载节点上的目录或文件', color: 'linear-gradient(135deg, #94a3b8 0%, #475569 100%)' }
]

export const volumeTypeText = (type?: string) =>
  VOLUME_TYPES.find(t => t.value === type)?.label ||
  (({ projected: '投射卷', downwardAPI: 'DownwardAPI', nfs: 'NFS', other: '其他' }) as Record<string, string>)[type || ''] ||
  type ||
  '-'

// ---------- 表单 ----------

export interface PortForm {
  name: string
  containerPort?: number
  protocol: 'TCP' | 'UDP' | 'SCTP'
}

export interface ProbeForm {
  enabled: boolean
  type: 'http' | 'tcp' | 'exec'
  path: string
  port?: number
  scheme: 'HTTP' | 'HTTPS'
  commandText: string
  initialDelaySeconds: number
  periodSeconds: number
  timeoutSeconds: number
  failureThreshold: number
  successThreshold: number
}

export interface MountForm {
  volume: string
  mountPath: string
  subPath: string
  readOnly: boolean
}

export interface WorkloadContainerForm extends ContainerForm {
  init: boolean
  ports: PortForm[]
  liveness: ProbeForm
  readiness: ProbeForm
  startup: ProbeForm
  mounts: MountForm[]
}

export interface VolumeForm {
  name: string
  type: WorkloadVolumeDTO['type']
  source: string
  readOnly: boolean
  medium: '' | 'Memory'
  sizeLimit: string
  hostPathType: string
}

export interface ClaimTemplateForm {
  name: string
  storageClassName: string
  accessMode: string
  capacityValue: number
  capacityUnit: CapacityUnit
}

export const createProbeForm = (): ProbeForm => ({
  enabled: false,
  type: 'http',
  path: '/',
  port: undefined,
  scheme: 'HTTP',
  commandText: '',
  initialDelaySeconds: 0,
  periodSeconds: 10,
  timeoutSeconds: 1,
  failureThreshold: 3,
  successThreshold: 1
})

export const createWorkloadContainerForm = (index = 0): WorkloadContainerForm => ({
  ...createContainerForm(index),
  // 工作负载默认使用镜像自带的启动命令
  commandMode: 'default',
  init: false,
  ports: [],
  liveness: createProbeForm(),
  readiness: createProbeForm(),
  startup: createProbeForm(),
  mounts: []
})

export const createVolumeForm = (index = 0, type: VolumeForm['type'] = 'pvc'): VolumeForm => ({
  name: `volume-${index + 1}`,
  type,
  source: '',
  readOnly: false,
  medium: '',
  sizeLimit: '',
  hostPathType: ''
})

export const createClaimTemplateForm = (index = 0): ClaimTemplateForm => ({
  name: index === 0 ? 'data' : `data-${index + 1}`,
  storageClassName: '',
  accessMode: 'ReadWriteOnce',
  capacityValue: 10,
  capacityUnit: 'Gi'
})

const probeToDTO = (p: ProbeForm, allowSuccess: boolean): ProbeDTO | undefined => {
  if (!p.enabled) return undefined
  const dto: ProbeDTO = {
    type: p.type,
    initialDelaySeconds: p.initialDelaySeconds,
    periodSeconds: p.periodSeconds,
    timeoutSeconds: p.timeoutSeconds,
    failureThreshold: p.failureThreshold
  }
  if (allowSuccess) dto.successThreshold = p.successThreshold
  if (p.type === 'http') Object.assign(dto, { path: p.path.trim() || '/', port: p.port, scheme: p.scheme })
  else if (p.type === 'tcp') dto.port = p.port
  else
    dto.command = p.commandText
      .split('\n')
      .map(s => s.trim())
      .filter(Boolean)
  return dto
}

export function workloadContainerToDTO(c: WorkloadContainerForm): WorkloadContainerDTO {
  return {
    ...containerFormToDTO(c),
    init: c.init,
    ports: c.ports
      .filter(p => p.containerPort)
      .map(p => ({ name: p.name.trim() || undefined, containerPort: p.containerPort as number, protocol: p.protocol })),
    livenessProbe: c.init ? undefined : probeToDTO(c.liveness, false),
    readinessProbe: c.init ? undefined : probeToDTO(c.readiness, true),
    startupProbe: c.init ? undefined : probeToDTO(c.startup, false),
    volumeMounts: c.mounts
      .filter(m => m.volume && m.mountPath.trim())
      .map(m => ({ volume: m.volume, mountPath: m.mountPath.trim(), subPath: m.subPath.trim() || undefined, readOnly: m.readOnly }))
  }
}

export const volumeToDTO = (v: VolumeForm): WorkloadVolumeDTO => ({
  name: v.name.trim(),
  type: v.type,
  source: v.type === 'emptyDir' ? undefined : v.source.trim(),
  readOnly: v.type === 'pvc' ? v.readOnly : undefined,
  medium: v.type === 'emptyDir' ? v.medium : undefined,
  sizeLimit: v.type === 'emptyDir' ? v.sizeLimit.trim() || undefined : undefined,
  hostPathType: v.type === 'hostPath' ? v.hostPathType || undefined : undefined
})

export const claimToDTO = (c: ClaimTemplateForm): VolumeClaimTemplateDTO => ({
  name: c.name.trim(),
  storageClassName: c.storageClassName || undefined,
  accessModes: [c.accessMode],
  capacity: `${c.capacityValue}${c.capacityUnit}`
})

export const rowsToRecord = (rows: KVRow[]) => {
  const result: Record<string, string> = {}
  rows.forEach(r => {
    if (r.key.trim()) result[r.key.trim()] = r.value
  })
  return result
}

export const recordToRows = (m?: Record<string, string> | null, hidden: string[] = []): KVRow[] =>
  Object.entries(m || {})
    .filter(([k]) => !hidden.includes(k))
    .map(([key, value]) => ({ key, value }))

// ---------- YAML 预览(与后端生成的对象保持一致) ----------

const clean = (obj: Record<string, unknown>) => {
  const result: Record<string, unknown> = {}
  Object.entries(obj).forEach(([k, v]) => {
    if (v === undefined || v === '' || v === null) return
    if (Array.isArray(v) && !v.length) return
    if (typeof v === 'object' && !Array.isArray(v) && !Object.keys(v as object).length) return
    result[k] = v
  })
  return result
}

const probePreview = (p?: ProbeDTO) => {
  if (!p) return undefined
  const base = clean({
    initialDelaySeconds: p.initialDelaySeconds || undefined,
    periodSeconds: p.periodSeconds,
    timeoutSeconds: p.timeoutSeconds,
    failureThreshold: p.failureThreshold,
    successThreshold: p.successThreshold
  })
  if (p.type === 'http') return { httpGet: { path: p.path, port: p.port, scheme: p.scheme }, ...base }
  if (p.type === 'tcp') return { tcpSocket: { port: p.port }, ...base }
  return { exec: { command: p.command }, ...base }
}

export const containerPreview = (c: WorkloadContainerDTO) =>
  clean({
    name: c.name,
    image: c.image,
    imagePullPolicy: c.imagePullPolicy,
    command: c.command,
    args: c.args,
    workingDir: c.workingDir,
    ports: c.ports?.map(p => clean({ name: p.name, containerPort: p.containerPort, protocol: p.protocol })),
    env: c.env,
    resources: clean({
      requests: clean({ cpu: c.cpuRequest, memory: c.memoryRequest }),
      limits: clean({ cpu: c.cpuLimit, memory: c.memoryLimit })
    }),
    livenessProbe: probePreview(c.livenessProbe),
    readinessProbe: probePreview(c.readinessProbe),
    startupProbe: probePreview(c.startupProbe),
    volumeMounts: c.volumeMounts?.map(m => clean({ name: m.volume, mountPath: m.mountPath, subPath: m.subPath, readOnly: m.readOnly || undefined }))
  })

export const volumePreview = (v: WorkloadVolumeDTO) => {
  switch (v.type) {
    case 'pvc':
      return { name: v.name, persistentVolumeClaim: clean({ claimName: v.source, readOnly: v.readOnly || undefined }) }
    case 'configMap':
      return { name: v.name, configMap: { name: v.source } }
    case 'secret':
      return { name: v.name, secret: { secretName: v.source } }
    case 'emptyDir':
      return { name: v.name, emptyDir: clean({ medium: v.medium, sizeLimit: v.sizeLimit }) }
    default:
      return { name: v.name, hostPath: clean({ path: v.source, type: v.hostPathType }) }
  }
}

export const workloadYamlExample = (kind: WorkloadKind, namespace = 'default') => {
  const meta = ['metadata:', `  name: example-${kind}`, `  namespace: ${namespace}`, '  labels:', `    app: example-${kind}`]
  const template = [
    '  template:',
    '    metadata:',
    '      labels:',
    `        app: example-${kind}`,
    '    spec:',
    '      containers:',
    '        - name: nginx',
    '          image: nginx:1.27',
    '          ports:',
    '            - containerPort: 80'
  ]
  const selector = ['  selector:', '    matchLabels:', `      app: example-${kind}`]
  if (kind === 'deployment') return ['apiVersion: apps/v1', 'kind: Deployment', ...meta, 'spec:', '  replicas: 2', ...selector, ...template].join('\n')
  if (kind === 'statefulset')
    return ['apiVersion: apps/v1', 'kind: StatefulSet', ...meta, 'spec:', `  serviceName: example-${kind}`, '  replicas: 2', ...selector, ...template].join('\n')
  return ['apiVersion: apps/v1', 'kind: DaemonSet', ...meta, 'spec:', ...selector, ...template].join('\n')
}

// ---------- 校验(工作负载和容器组共用), 返回错误信息, 没有错误时返回空字符串 ----------

const QUANTITY = /^\d+(\.\d+)?(m|k|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$/
const PORT_NAME = /^(?=.*[a-z])[a-z0-9]([-a-z0-9]*[a-z0-9])?$/

export function validateWorkloadContainers(forms: WorkloadContainerForm[]): string {
  const names = new Set<string>()
  if (!forms.some(c => !c.init)) return '除初始化容器外，至少需要一个工作容器'
  for (const c of forms) {
    const label = `容器 ${c.name || '(未命名)'}`
    if (!DNS_LABEL.test(c.name.trim())) return `${label}的名称不合法：只能包含小写字母、数字和 "-"`
    if (names.has(c.name.trim())) return `容器名称 ${c.name} 重复`
    names.add(c.name.trim())
    if (!c.image.trim() || /\s/.test(c.image.trim())) return `请填写${label}的镜像`
    const resources: [string, string][] = [
      [c.cpuRequest, 'CPU 预留'],
      [c.cpuLimit, 'CPU 上限'],
      [c.memoryRequest, '内存预留'],
      [c.memoryLimit, '内存上限']
    ]
    for (const [value, text] of resources) {
      if (value.trim() && !QUANTITY.test(value.trim())) return `${label}的${text}格式不正确，例如 100m、0.5、128Mi`
    }
    const portKeys = new Set<string>()
    for (const p of c.ports) {
      if (!p.containerPort) continue
      const key = `${p.containerPort}/${p.protocol}`
      if (portKeys.has(key)) return `${label}的端口 ${key} 重复`
      portKeys.add(key)
      if (p.name.trim() && (!PORT_NAME.test(p.name.trim()) || p.name.trim().length > 15)) return `${label}的端口名称 ${p.name} 不合法：最多 15 个字符，小写字母、数字和 "-"`
    }
    if (!c.init) {
      const probes: [ProbeForm, string][] = [
        [c.readiness, '就绪检查'],
        [c.liveness, '存活检查'],
        [c.startup, '启动检查']
      ]
      for (const [probe, text] of probes) {
        if (!probe.enabled) continue
        if (probe.type !== 'exec' && !probe.port) return `请填写${label}${text}的端口`
        if (probe.type === 'exec' && !probe.commandText.trim()) return `请填写${label}${text}执行的命令`
      }
    }
    const badEnv = c.env.find(e => e.key.trim() && !/^[-._a-zA-Z][-._a-zA-Z0-9]*$/.test(e.key.trim()))
    if (badEnv) return `${label}的环境变量名 ${badEnv.key} 不合法`
  }
  return ''
}

export function validateWorkloadStorage(volumes: VolumeForm[], claims: ClaimTemplateForm[], forms: WorkloadContainerForm[]): string {
  const names = new Set<string>()
  for (const t of claims) {
    if (!DNS_LABEL.test(t.name.trim())) return `存储卷模板名称 ${t.name || '(空)'} 不合法`
    if (names.has(t.name.trim())) return `存储卷名称 ${t.name} 重复`
    names.add(t.name.trim())
  }
  for (const v of volumes) {
    if (!DNS_LABEL.test(v.name.trim())) return `存储卷名称 ${v.name || '(空)'} 不合法`
    if (names.has(v.name.trim())) return `存储卷名称 ${v.name} 重复`
    names.add(v.name.trim())
    if (v.type === 'hostPath' && !v.source.trim().startsWith('/')) return `存储卷 ${v.name} 的主机路径必须是绝对路径`
    if (['pvc', 'configMap', 'secret'].includes(v.type) && !v.source) return `请选择存储卷 ${v.name} 使用的${volumeTypeText(v.type)}`
    if (v.sizeLimit.trim() && !QUANTITY.test(v.sizeLimit.trim())) return `存储卷 ${v.name} 的容量上限格式不正确`
  }
  for (const c of forms) {
    const paths = new Set<string>()
    for (const m of c.mounts) {
      if (!m.volume && !m.mountPath.trim()) continue
      if (!names.has(m.volume)) return `容器 ${c.name} 挂载的存储卷 ${m.volume || '(未选择)'} 不存在`
      if (!m.mountPath.trim().startsWith('/')) return `容器 ${c.name} 的挂载路径必须是绝对路径`
      if (paths.has(m.mountPath.trim())) return `容器 ${c.name} 的挂载路径 ${m.mountPath} 重复`
      paths.add(m.mountPath.trim())
    }
  }
  return ''
}

// 同步主机时区时追加的存储卷和挂载(预览用, 与后端一致)
export function withHostTime(podSpec: Record<string, unknown>) {
  const mount = (list?: Record<string, unknown>[]) =>
    list?.map(c => ({ ...c, volumeMounts: [...((c.volumeMounts as unknown[]) || []), { name: 'host-time', mountPath: '/etc/localtime', readOnly: true }] }))
  if (podSpec.initContainers) podSpec.initContainers = mount(podSpec.initContainers as Record<string, unknown>[])
  podSpec.containers = mount(podSpec.containers as Record<string, unknown>[])
  podSpec.volumes = [...((podSpec.volumes as unknown[]) || []), { name: 'host-time', hostPath: { path: '/etc/localtime' } }]
}

// 工作负载所属控制器在本平台的详情页(集群下的相对路径); 没有详情页时返回空字符串
export const ownerDetailPath = (kind?: string) =>
  (({
    Deployment: 'workload/deployment/detail',
    StatefulSet: 'workload/statefulset/detail',
    DaemonSet: 'workload/daemonset/detail',
    Job: 'job/job/detail',
    CronJob: 'job/cronjob/detail'
  }) as Record<string, string>)[kind || ''] || ''


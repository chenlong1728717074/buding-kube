import type { StorageTypeKey } from '@/api/storageclass'

// 与后端 pkg/consts/kube.go 中的注解保持一致
export const SC_ANNOTATIONS = {
  storageType: 'kube.buding.goaigc.fun/storage-type',
  isDefault: 'storageclass.kubernetes.io/is-default-class',
  allowClone: 'storageclass.kubesphere.io/allow-clone',
  accessModes: 'storageclass.kubesphere.io/supported-access-modes'
} as const

export const STORAGE_TYPE_OPTIONS: Array<{ value: Exclude<StorageTypeKey, ''>; label: string; desc: string }> = [
  { value: 'block', label: '块存储', desc: '如 Ceph RBD、Longhorn、云硬盘，通常单节点读写' },
  { value: 'file', label: '文件存储', desc: '如 NFS、CephFS、NAS，支持多个 Pod 共享读写' },
  { value: 'object', label: '对象存储', desc: '如 S3、OSS、MinIO，通过 CSI 挂载' }
]

export const storageTypeLabel = (key?: string) =>
  STORAGE_TYPE_OPTIONS.find(o => o.value === key)?.label || '未分类'

export const ACCESS_MODE_OPTIONS = [
  { value: 'ReadWriteOnce', label: 'RWO', desc: '单节点读写' },
  { value: 'ReadOnlyMany', label: 'ROX', desc: '多节点只读' },
  { value: 'ReadWriteMany', label: 'RWX', desc: '多节点读写' },
  { value: 'ReadWriteOncePod', label: 'RWOP', desc: '单 Pod 读写' }
]

export interface ParamRow {
  key: string
  value: string
}

// 存储系统模板(参考 KubeSphere 创建存储类时的「存储系统」选择)
export interface ProvisionerTemplate {
  key: string
  name: string
  short: string
  color: string
  desc: string
  provisioner: string
  storageType: StorageTypeKey
  reclaimPolicy: 'Delete' | 'Retain'
  volumeBindingMode: 'Immediate' | 'WaitForFirstConsumer'
  allowVolumeExpansion: boolean
  allowVolumeClone: boolean
  accessModes: string[]
  mountOptions: string[]
  params: ParamRow[]
}

const cephSecretParams = (prefix: 'rbd' | 'cephfs'): ParamRow[] => [
  { key: 'csi.storage.k8s.io/provisioner-secret-name', value: `csi-${prefix}-secret` },
  { key: 'csi.storage.k8s.io/provisioner-secret-namespace', value: 'ceph-csi' },
  { key: 'csi.storage.k8s.io/controller-expand-secret-name', value: `csi-${prefix}-secret` },
  { key: 'csi.storage.k8s.io/controller-expand-secret-namespace', value: 'ceph-csi' },
  { key: 'csi.storage.k8s.io/node-stage-secret-name', value: `csi-${prefix}-secret` },
  { key: 'csi.storage.k8s.io/node-stage-secret-namespace', value: 'ceph-csi' }
]

export const PROVISIONER_TEMPLATES: ProvisionerTemplate[] = [
  {
    key: 'nfs',
    name: 'NFS',
    short: 'NFS',
    color: 'linear-gradient(135deg, #34d399 0%, #059669 100%)',
    desc: 'NFS CSI 驱动，多个 Pod 可以共享读写同一个卷',
    provisioner: 'nfs.csi.k8s.io',
    storageType: 'file',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: true,
    allowVolumeClone: false,
    accessModes: ['ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany'],
    mountOptions: ['nfsvers=4.1'],
    params: [
      { key: 'server', value: '' },
      { key: 'share', value: '' }
    ]
  },
  {
    key: 'ceph-rbd',
    name: 'Ceph RBD',
    short: 'RBD',
    color: 'linear-gradient(135deg, #f87171 0%, #dc2626 100%)',
    desc: 'Ceph 块存储，性能较好，适合数据库等单节点读写场景',
    provisioner: 'rbd.csi.ceph.com',
    storageType: 'block',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: true,
    allowVolumeClone: true,
    accessModes: ['ReadWriteOnce'],
    mountOptions: ['discard'],
    params: [
      { key: 'clusterID', value: '' },
      { key: 'pool', value: '' },
      { key: 'imageFeatures', value: 'layering' },
      { key: 'csi.storage.k8s.io/fstype', value: 'ext4' },
      ...cephSecretParams('rbd')
    ]
  },
  {
    key: 'cephfs',
    name: 'CephFS',
    short: 'CFS',
    color: 'linear-gradient(135deg, #fb923c 0%, #ea580c 100%)',
    desc: 'Ceph 文件系统，支持多节点共享读写',
    provisioner: 'cephfs.csi.ceph.com',
    storageType: 'file',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: true,
    allowVolumeClone: true,
    accessModes: ['ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany'],
    mountOptions: [],
    params: [
      { key: 'clusterID', value: '' },
      { key: 'fsName', value: '' },
      ...cephSecretParams('cephfs')
    ]
  },
  {
    key: 'longhorn',
    name: 'Longhorn',
    short: 'LH',
    color: 'linear-gradient(135deg, #818cf8 0%, #4f46e5 100%)',
    desc: '云原生分布式块存储，自带副本和快照',
    provisioner: 'driver.longhorn.io',
    storageType: 'block',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: true,
    allowVolumeClone: true,
    accessModes: ['ReadWriteOnce', 'ReadWriteMany'],
    mountOptions: [],
    params: [
      { key: 'numberOfReplicas', value: '3' },
      { key: 'staleReplicaTimeout', value: '30' },
      { key: 'fsType', value: 'ext4' }
    ]
  },
  {
    key: 'local',
    name: '本地存储',
    short: 'LOC',
    color: 'linear-gradient(135deg, #94a3b8 0%, #475569 100%)',
    desc: '使用节点本地磁盘，需要手动创建 PV，不支持动态供给',
    provisioner: 'kubernetes.io/no-provisioner',
    storageType: 'block',
    reclaimPolicy: 'Retain',
    volumeBindingMode: 'WaitForFirstConsumer',
    allowVolumeExpansion: false,
    allowVolumeClone: false,
    accessModes: ['ReadWriteOnce'],
    mountOptions: [],
    params: []
  },
  {
    key: 'custom',
    name: '自定义',
    short: '···',
    color: 'linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%)',
    desc: '填写任意供应者，例如云厂商的 CSI 驱动',
    provisioner: '',
    storageType: '',
    reclaimPolicy: 'Delete',
    volumeBindingMode: 'Immediate',
    allowVolumeExpansion: false,
    allowVolumeClone: false,
    accessModes: [],
    mountOptions: [],
    params: []
  }
]

export const DEFAULT_STORAGECLASS_YAML = `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: my-storageclass
  annotations:
    alias: 示例存储类
    describe: 通过 YAML 创建
    ${SC_ANNOTATIONS.storageType}: file
provisioner: nfs.csi.k8s.io
parameters:
  server: 192.168.1.100
  share: /data/nfs
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
mountOptions:
  - nfsvers=4.1
`

// 键值行转对象, 忽略空键
export const rowsToMap = (rows: ParamRow[]): Record<string, string> => {
  const result: Record<string, string> = {}
  for (const row of rows) {
    const key = row.key.trim()
    if (key) result[key] = row.value
  }
  return result
}

// 对象转键值行
export const mapToRows = (map?: Record<string, string> | null): ParamRow[] =>
  Object.entries(map || {}).map(([key, value]) => ({ key, value: String(value ?? '') }))

export const ACCESS_MODE_SHORT: Record<string, string> = {
  ReadWriteOnce: 'RWO',
  ReadOnlyMany: 'ROX',
  ReadWriteMany: 'RWX',
  ReadWriteOncePod: 'RWOP'
}

export const formatAccessModes = (modes?: string[]) =>
  modes && modes.length ? modes.map(m => ACCESS_MODE_SHORT[m] || m).join(' / ') : '-'

export const CAPACITY_UNITS = ['Mi', 'Gi', 'Ti'] as const
export type CapacityUnit = (typeof CAPACITY_UNITS)[number]

// "10Gi" -> { value: 10, unit: 'Gi' }; 无法识别时返回 null
export const parseCapacity = (text?: string): { value: number; unit: CapacityUnit } | null => {
  const m = /^(\d+(?:\.\d+)?)(Mi|Gi|Ti)$/.exec((text || '').trim())
  return m ? { value: Number(m[1]), unit: m[2] as CapacityUnit } : null
}

// 用于比较容量大小(单位 Mi)
export const capacityToMi = (text?: string) => {
  const parsed = parseCapacity(text)
  if (!parsed) return 0
  return parsed.value * (parsed.unit === 'Ti' ? 1024 * 1024 : parsed.unit === 'Gi' ? 1024 : 1)
}

export const PV_SOURCE_LABELS: Record<string, string> = {
  nfs: 'NFS',
  hostPath: 'HostPath',
  local: '本地存储',
  csi: 'CSI',
  rbd: 'Ceph RBD',
  cephfs: 'CephFS',
  iscsi: 'iSCSI',
  fc: 'FC'
}

export const DEFAULT_PV_YAML = `apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv-01
  annotations:
    alias: 示例持久卷
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: ""
  volumeMode: Filesystem
  mountOptions:
    - nfsvers=4.1
  nfs:
    server: 192.168.1.100
    path: /data/nfs/pv-01
`

export const defaultPvcYaml = (namespace = 'default') => `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: ${namespace}
  annotations:
    alias: 示例持久卷声明
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  # 不填写 storageClassName 时使用集群默认存储类
  # storageClassName: nfs-client
  volumeMode: Filesystem
`

import request, { type ApiResponse, type PageResponse } from '@/utils/request'
import type { ContainerTemplateVO, EventVO, JobContainerDTO, JobPodVO } from '@/api/job'

// Deployment / StatefulSet / DaemonSet 通用接口, 对应后端 /workload
export type WorkloadKind = 'deployment' | 'statefulset' | 'daemonset'

export interface WorkloadTarget {
  clusterId: string
  kind: WorkloadKind
  namespace: string
  name: string
}

export interface ContainerPortVO {
  name: string
  containerPort: number
  protocol: string
}

export interface WorkloadVolumeVO {
  name: string
  // pvc / configMap / secret / emptyDir / hostPath / projected / downwardAPI / nfs / other
  type: string
  source: string
}

export interface VolumeClaimTemplateVO {
  name: string
  storageClassName: string
  accessModes: string[]
  capacity: string
}

export interface WorkloadRevisionVO {
  revision: number
  name: string
  images: string[]
  replicas: number
  readyReplicas: number
  current: boolean
  changeCause?: string
  createTime: string
}

export interface WorkloadServiceRefVO {
  name: string
  type: string
  clusterIP: string
  ports: string
}

export interface WorkloadConditionVO {
  type: string
  status: string
  reason?: string
  message?: string
  lastTime?: string | null
}

export interface WorkloadDetailVO {
  kind: WorkloadKind
  name: string
  namespace: string
  alias?: string
  describe?: string
  status: string
  replicas: number
  readyReplicas: number
  updatedReplicas: number
  availableReplicas: number
  strategy: string
  maxSurge?: string
  maxUnavailable?: string
  partition?: number | null
  minReadySeconds: number
  revisionHistoryLimit?: number | null
  podManagementPolicy?: string
  serviceName?: string
  paused: boolean
  currentRevision: number
  selector?: Record<string, string> | null
  podLabels?: Record<string, string> | null
  nodeSelector?: Record<string, string> | null
  serviceAccountName?: string
  imagePullSecrets: string[]
  hostNetwork: boolean
  creator?: string
  uid: string
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  createTime: string
  updateTime?: string | null
  conditions: WorkloadConditionVO[]
  containers: ContainerTemplateVO[]
  volumes: WorkloadVolumeVO[]
  volumeClaimTemplates: VolumeClaimTemplateVO[]
  pods: JobPodVO[]
  revisions: WorkloadRevisionVO[]
  services: WorkloadServiceRefVO[]
  events: EventVO[]
}

export interface WorkloadSelectorVO {
  kind: WorkloadKind
  name: string
  labels?: Record<string, string> | null
  ports: ContainerPortVO[]
}

export interface ProbeDTO {
  type: 'http' | 'tcp' | 'exec'
  path?: string
  port?: number
  scheme?: 'HTTP' | 'HTTPS'
  command?: string[]
  initialDelaySeconds?: number
  periodSeconds?: number
  timeoutSeconds?: number
  failureThreshold?: number
  successThreshold?: number
}

export interface WorkloadContainerDTO extends JobContainerDTO {
  init?: boolean
  ports?: { name?: string; containerPort: number; protocol?: string }[]
  livenessProbe?: ProbeDTO
  readinessProbe?: ProbeDTO
  startupProbe?: ProbeDTO
  volumeMounts?: { volume: string; mountPath: string; subPath?: string; readOnly?: boolean }[]
}

export interface WorkloadVolumeDTO {
  name: string
  type: 'pvc' | 'configMap' | 'secret' | 'emptyDir' | 'hostPath'
  source?: string
  readOnly?: boolean
  medium?: string
  sizeLimit?: string
  hostPathType?: string
}

export interface VolumeClaimTemplateDTO {
  name: string
  storageClassName?: string
  accessModes: string[]
  capacity: string
}

export interface WorkloadCreateDTO {
  clusterId: string
  kind: WorkloadKind
  namespace: string
  name: string
  alias?: string
  describe?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  replicas?: number
  containers: WorkloadContainerDTO[]
  volumes?: WorkloadVolumeDTO[]
  volumeClaimTemplates?: VolumeClaimTemplateDTO[]
  strategy?: string
  maxSurge?: string
  maxUnavailable?: string
  partition?: number
  minReadySeconds?: number
  revisionHistoryLimit?: number
  podManagementPolicy?: string
  serviceName?: string
  nodeSelector?: Record<string, string>
  serviceAccountName?: string
  imagePullSecrets?: string[]
  syncHostTimezone?: boolean
  terminationGracePeriodSeconds?: number
  service?: { type: 'ClusterIP' | 'NodePort'; name?: string }
}

// 列表项(三种工作负载字段相同)
export interface WorkloadVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  status: string
  containers: number
  replicas: number
  readyReplicas: number
  updateTime?: string
  createTime: string
}

export interface WorkloadListQuery {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

const LIST_PATH: Record<WorkloadKind, string> = {
  deployment: '/deployment/list',
  statefulset: '/statefulSet/list',
  daemonset: '/daemonSet/list'
}

export const workloadApi = {
  // 列表, 后端按创建时间倒序
  list: (kind: WorkloadKind, params: WorkloadListQuery): Promise<ApiResponse<PageResponse<WorkloadVO>>> => request.get(LIST_PATH[kind], { params }),

  getDetail: (params: WorkloadTarget): Promise<ApiResponse<WorkloadDetailVO>> => request.get('/workload', { params }),

  // 命名空间中的工作负载及其容器组标签, 创建服务时选择用
  selectors: (params: { clusterId: string; namespace: string }): Promise<ApiResponse<WorkloadSelectorVO[]>> =>
    request.get('/workload/selectors', { params }),

  create: (data: WorkloadCreateDTO): Promise<ApiResponse<void>> => request.post('/workload', data),

  updateInfo: (data: WorkloadTarget & { alias: string; describe: string }): Promise<ApiResponse<void>> => request.put('/workload/info', data),

  scale: (data: WorkloadTarget & { replicas: number }): Promise<ApiResponse<void>> => request.put('/workload/scale', data),

  updateImages: (data: WorkloadTarget & { images: Record<string, string> }): Promise<ApiResponse<void>> => request.put('/workload/image', data),

  // 重新创建容器组(kubectl rollout restart)
  restart: (data: WorkloadTarget): Promise<ApiResponse<void>> => request.put('/workload/restart', data),

  rollback: (data: WorkloadTarget & { revision: number }): Promise<ApiResponse<void>> => request.put('/workload/rollback', data),

  delete: (params: WorkloadTarget): Promise<ApiResponse<void>> => request.delete('/workload', { params })
}

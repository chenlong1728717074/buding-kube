import request, { type ApiResponse, type PageResponse } from '@/utils/request'
import type { EventVO, JobPodVO } from '@/api/job'
import type { WorkloadConditionVO } from '@/api/workload'

// 节点资源: CPU 单位为毫核, 内存单位为字节; 用量为 null 表示没有 metrics-server 数据
export interface NodeResourceVO {
  cpuCapacity: number
  cpuAllocatable: number
  cpuRequests: number
  cpuLimits: number
  cpuUsage?: number | null
  memoryCapacity: number
  memoryAllocatable: number
  memoryRequests: number
  memoryLimits: number
  memoryUsage?: number | null
  podCount: number
  podAllocatable: number
}

export interface NodeVO extends NodeResourceVO {
  hostname: string
  role: string
  roles: string[]
  status: 'Ready' | 'NotReady' | 'Unknown' | string
  ip: string
  cpu: string
  memory: string
  unSchedule: boolean
  kubeletVersion: string
  taintCount: number
  createTime: string
}

export interface NodeTaintVO {
  key: string
  value: string
  effect: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute' | string
  // 由节点控制器维护, 不能手动修改
  system: boolean
  timeAdded?: string | null
}

export interface NodeDetailVO {
  server: {
    hostname: string
    role: string
    status: string
    ip: string
    unSchedule: boolean
    osImage: string
    osType: string
    arch: string
    kernelVersion: string
    containerRuntime: string
    kubeletVersion: string
    kubeProxyVersion: string
    createTime: string
  }
  runtime: {
    cpu: string
    memory: string
    ephemeralStorage: string
    nvidiaGpu?: string
    amdGpu?: string
    intelGpu?: string
  }
  resources: NodeResourceVO
  roles: string[]
  addresses: { type: string; address: string }[]
  taints: NodeTaintVO[]
  conditions: WorkloadConditionVO[]
  events: EventVO[]
  pods: JobPodVO[]
  annotations?: Record<string, string> | null
  labels?: Record<string, string> | null
  podCIDR?: string
  providerID?: string
  uid: string
}

export interface NodeQueryDTO {
  clusterId: string
  keyword?: string
  status?: string
  // control-plane / worker
  role?: string
  page: number
  pageSize: number
}

export interface NodeTarget {
  clusterId: string
  hostname: string
}

export interface NodeDrainOptions {
  deleteEmptyDirData: boolean
  force: boolean
  // 小于 0 表示使用容器组自己的设置
  gracePeriodSeconds: number
  timeoutSeconds: number
}

export interface DrainPodVO {
  namespace: string
  name: string
  ownerKind?: string
  ownerName?: string
  status: string
  // evict(驱逐) / delete(已结束, 直接删除) / skip(跳过) / block(阻止排空)
  action: 'evict' | 'delete' | 'skip' | 'block'
  reason?: string
}

export interface DrainPreviewVO {
  pods: DrainPodVO[]
  evict: number
  skip: number
  blocked: number
}

export interface DrainResultVO {
  cordoned: boolean
  evicted: string[]
  failed: DrainPodVO[]
  skipped: number
}

export const nodeApi = {
  getList: (params: NodeQueryDTO): Promise<ApiResponse<PageResponse<NodeVO>>> => request.get('/node/list', { params }),

  getDetail: (params: NodeTarget): Promise<ApiResponse<NodeDetailVO>> => request.get('/node', { params }),

  // 停止 / 启用调度
  setSchedule: (data: NodeTarget & { unSchedule: boolean }): Promise<ApiResponse<void>> => request.put('/node/unSchedule', data),

  // 整体替换标签, 系统维护的标签由后端保留
  updateLabels: (data: NodeTarget & { labels: Record<string, string> }): Promise<ApiResponse<void>> => request.put('/node/labels', data),

  // 整体替换污点, 节点控制器维护的污点由后端保留
  updateTaints: (data: NodeTarget & { taints: { key: string; value: string; effect: string }[] }): Promise<ApiResponse<void>> =>
    request.put('/node/taints', data),

  drainPreview: (params: NodeTarget & NodeDrainOptions): Promise<ApiResponse<DrainPreviewVO>> => request.get('/node/drain/preview', { params }),

  // 驱逐可能等待 PodDisruptionBudget, 耗时较长, 不设置超时
  drain: (data: NodeTarget & NodeDrainOptions): Promise<ApiResponse<DrainResultVO>> => request.post('/node/drain', data, { timeout: 0 }),

  delete: (params: { clusterId: string; name: string }): Promise<ApiResponse<void>> => request.delete('/node', { params })
}

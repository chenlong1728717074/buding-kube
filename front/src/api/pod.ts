import request from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'

// Pod基础信息
export interface PodVO {
  creationTimestamp: string
  namespace: string
  name: string
  status: string
  nodeName: string
  hostIP: string
  podIP: string
}

// Pod详情信息
export interface PodInfoVO {
  creationTimestamp: string
  namespace: string
  name: string
  status: string
  nodeName: string
  hostIP: string
  podIP: string
  yaml: string
  containers: PodContainerVO[]
  annotations: Record<string, string>
  labels: Record<string, string>
  metaData: any
  spec: any
  events: EventVO[]
}

// Pod容器信息
export interface PodContainerVO {
  name: string
  ready: boolean
  restartCount: number
  image: string
  started: boolean
  volumeMounts: any[]
  startedAt: string
  args: string[]
  envs: PodEnvVO[]
  ports: PodPortVO[]
}

// Pod环境变量
export interface PodEnvVO {
  name: string
  value: string
}

// Pod端口信息
export interface PodPortVO {
  name: string
  hostPort: number
  containerPort: number
  hostIP: string
  protocol: string
}

// 事件信息
export interface EventVO {
  type: string
  reason: string
  message: string
  firstTimestamp: string
  lastTimestamp: string
  count: number
}

// Pod查询参数
export interface PodQueryDTO {
  clusterId: string
  namespace: string
  status?: string
  page: number
  pageSize: number
}

// Pod基础参数
export interface PodDTO {
  clusterId: string
  namespace: string
  name: string
}

// Pod API
export const podApi = {
  // 获取Pod列表
  getList(params: PodQueryDTO): Promise<ApiResponse<PageResponse<PodVO>>> {
    return request.get('/pod/list', { params })
  },

  // 获取Pod详情
  getInfo(params: PodDTO): Promise<ApiResponse<PodInfoVO>> {
    return request.get('/pod', { params })
  },

  // 删除Pod
  delete(params: PodDTO): Promise<ApiResponse<void>> {
    return request.delete('/pod', { params })
  }
}
import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'
import type { EventVO, JobPodVO } from '@/api/job'
import type { WorkloadKind } from '@/api/workload'

// Service查询参数
export interface ServiceQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  // ClusterIP / NodePort / LoadBalancer / ExternalName / Headless
  type?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export interface ServicePortVO {
  name: string
  protocol: string
  port: number
  targetPort: string
  nodePort: number
}

// Service视图对象
export interface ServiceVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  yaml?: string
  status: string
  internalIP: string
  externalIP: string
  internalAccess: string
  externalAccess: string
  accessMode: string
  serviceType: string
  headless: boolean
  ports: ServicePortVO[]
  selector?: Record<string, string> | null
  createTime: string
}

export interface ServiceEndpointVO {
  ip: string
  ready: boolean
  podName?: string
  nodeName?: string
  ports: string
}

export interface ServiceWorkloadVO {
  kind: WorkloadKind
  name: string
  replicas: number
  readyReplicas: number
}

export interface ServiceDetailVO extends ServiceVO {
  clusterIP: string
  clusterIPs?: string[] | null
  externalIPs?: string[] | null
  sessionAffinity: string
  sessionAffinityTimeout?: number | null
  externalName?: string
  externalTrafficPolicy?: string
  loadBalancerIngress: string[]
  ipFamilies: string[]
  creator?: string
  uid: string
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  endpoints: ServiceEndpointVO[]
  pods: JobPodVO[]
  workloads: ServiceWorkloadVO[]
  ingresses: string[]
  events: EventVO[]
}

export interface ServicePortDTO {
  name?: string
  protocol: string
  port: number
  targetPort?: string
  nodePort?: number
}

// 创建 / 编辑服务
export interface ServiceFormDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName'
  headless?: boolean
  selector?: Record<string, string>
  ports: ServicePortDTO[]
  sessionAffinity?: 'None' | 'ClientIP'
  sessionAffinityTimeout?: number
  externalName?: string
  externalTrafficPolicy?: 'Cluster' | 'Local' | ''
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

// Service基础参数
export interface ServiceBaseDTO {
  clusterId: string
  namespace: string
  name: string
}

// Service API
export const serviceApi = {
  // 获取Service列表
  getServices: (params: ServiceQueryDTO): Promise<ApiResponse<PageResponse<ServiceVO>>> => {
    return request.get('/service/list', { params })
  },

  getDetail: (params: ServiceBaseDTO): Promise<ApiResponse<ServiceDetailVO>> => {
    return request.get('/service', { params })
  },

  create: (data: ServiceFormDTO): Promise<ApiResponse<void>> => {
    return request.post('/service', data)
  },

  update: (data: ServiceFormDTO): Promise<ApiResponse<void>> => {
    return request.put('/service', data)
  },

  // 删除Service
  deleteService: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/service', { params })
  }
}

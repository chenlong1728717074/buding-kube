import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'
import type { EventVO } from '@/api/job'

// Ingress查询参数
export interface IngressQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// Ingress视图对象
export interface IngressVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  ingressClassName?: string | null
  hosts: string[]
  addresses: string[]
  tls: boolean
  ruleCount: number
  createTime: string
  yaml?: string
}

export interface IngressRuleVO {
  host: string
  path: string
  pathType: string
  serviceName: string
  servicePort: string
  serviceExists: boolean
}

export interface IngressTLSVO {
  hosts: string[]
  secretName: string
}

export interface IngressDetailVO extends IngressVO {
  rules: IngressRuleVO[]
  tlsList: IngressTLSVO[]
  defaultBackend?: string
  creator?: string
  uid: string
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  events: EventVO[]
}

export interface IngressClassVO {
  name: string
  controller: string
  isDefault: boolean
}

export interface IngressPathDTO {
  path: string
  pathType: 'Prefix' | 'Exact' | 'ImplementationSpecific'
  serviceName: string
  servicePort: string
}

export interface IngressFormDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  ingressClassName?: string
  rules: { host: string; paths: IngressPathDTO[] }[]
  tls?: { hosts: string[]; secretName: string }[]
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

// Ingress API
export const ingressApi = {
  // 获取Ingress列表
  getIngressList: (params: IngressQueryDTO): Promise<ApiResponse<PageResponse<IngressVO>>> => {
    return request.get('/ingress/list', { params })
  },

  getDetail: (params: { clusterId: string; namespace: string; name: string }): Promise<ApiResponse<IngressDetailVO>> => {
    return request.get('/ingress', { params })
  },

  create: (data: IngressFormDTO): Promise<ApiResponse<void>> => {
    return request.post('/ingress', data)
  },

  update: (data: IngressFormDTO): Promise<ApiResponse<void>> => {
    return request.put('/ingress', data)
  },

  // 集群中的 Ingress 类
  getClasses: (params: { clusterId: string }): Promise<ApiResponse<IngressClassVO[]>> => {
    return request.get('/ingress/classes', { params })
  },

  // 命名空间中 kubernetes.io/tls 类型的 Secret
  getTlsSecrets: (params: { clusterId: string; namespace: string }): Promise<ApiResponse<string[]>> => {
    return request.get('/ingress/tlsSecrets', { params })
  },

  // 删除Ingress
  deleteIngress: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/ingress', { params })
  }
}

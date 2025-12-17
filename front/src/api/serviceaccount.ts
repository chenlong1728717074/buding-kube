import request, { type ApiResponse, type PageResponse } from '@/utils/request'

export interface ServiceAccountPageQueryDTO {
  clusterId: string
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export interface ServiceAccountBaseDTO {
  clusterId: string
  namespace: string
  name: string
  force?: boolean
}

export interface ServiceAccountCreateDTO {
  clusterId: string
  namespace: string
  name: string
  automountServiceAccountToken?: boolean
  imagePullSecrets?: string[]
  secrets?: string[]
  annotations?: Record<string, string>
  labels?: Record<string, string>
}

export interface ServiceAccountApplyDTO {
  clusterId: string
  namespace?: string
  yaml: string
}

export interface ServiceAccountVO {
  name: string
  namespace: string
  automountServiceAccountToken?: boolean
  imagePullSecrets?: { name: string }[]
  secrets?: { name: string }[]
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
  yaml?: string
}

export const serviceAccountApi = {
  getList(params: ServiceAccountPageQueryDTO): Promise<ApiResponse<PageResponse<ServiceAccountVO>>> {
    return request.get('/serviceAccount/list', { params })
  },
  add(data: ServiceAccountCreateDTO): Promise<ApiResponse<void>> {
    return request.post('/serviceAccount/add', data)
  },
  update(data: ServiceAccountCreateDTO): Promise<ApiResponse<void>> {
    return request.put('/serviceAccount', data)
  },
  applyYaml(data: ServiceAccountApplyDTO): Promise<ApiResponse<void>> {
    return request.post('/serviceAccount/apply', data)
  },
  delete(params: ServiceAccountBaseDTO): Promise<ApiResponse<void>> {
    return request.delete('/serviceAccount', { params })
  }
}

export default serviceAccountApi


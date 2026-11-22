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

export interface ServiceAccountVO {
  name: string
  namespace: string
  automountServiceAccountToken?: boolean
  imagePullSecrets?: { name: string }[]
  secrets?: { name: string }[]
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
}

export const serviceAccountApi = {
  getList(params: ServiceAccountPageQueryDTO): Promise<ApiResponse<PageResponse<ServiceAccountVO>>> {
    return request.get('/serviceAccount/list', { params })
  },
  update(data: ServiceAccountCreateDTO): Promise<ApiResponse<void>> {
    return request.put('/serviceAccount', data)
  },
  delete(params: ServiceAccountBaseDTO): Promise<ApiResponse<void>> {
    return request.delete('/serviceAccount', { params })
  }
}

export default serviceAccountApi


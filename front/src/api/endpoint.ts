import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from './types'

// Endpoint查询参数
export interface EndpointQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// Endpoint视图对象
export interface EndpointVO {
  name: string
  namespace: string
  createTime: string
  yaml?: string
}

// Endpoint API
export const endpointApi = {
  // 获取Endpoint列表
  getEndpointList: (params: EndpointQueryDTO): Promise<ApiResponse<PageResponse<EndpointVO>>> => {
    return request.get('/endpoint/list', { params })
  }
}
import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/types/api'

// Service查询参数
export interface ServiceQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
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
  createTime: string
}

// Service API
export const serviceApi = {
  // 获取Service列表
  getServices: (params: ServiceQueryDTO): Promise<ApiResponse<PageResponse<ServiceVO>>> => {
    return request.get('/service/list', { params })
  }
}
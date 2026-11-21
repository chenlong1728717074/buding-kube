import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'

// EndpointSlice查询参数
export interface EndpointSliceQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// EndpointSlice视图对象
export interface EndpointSliceVO {
  name: string
  namespace: string
  addressType: string
  createTime: string
  yaml?: string
}

// EndpointSlice API
export const endpointSliceApi = {
  // 获取EndpointSlice列表
  getEndpointSliceList: (params: EndpointSliceQueryDTO): Promise<ApiResponse<PageResponse<EndpointSliceVO>>> => {
    return request.get('/endpointslice/list', { params })
  },

  // 删除EndpointSlice(注意: 由控制器管理, 删除后可能被自动重建)
  deleteEndpointSlice: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/endpointslice', { params })
  }
}
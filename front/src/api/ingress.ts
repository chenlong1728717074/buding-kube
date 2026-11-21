import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'

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
  ingressClassName?: string
  createTime: string
  yaml?: string
}

// Ingress API
export const ingressApi = {
  // 获取Ingress列表
  getIngressList: (params: IngressQueryDTO): Promise<ApiResponse<PageResponse<IngressVO>>> => {
    return request.get('/ingress/list', { params })
  },

  // 删除Ingress
  deleteIngress: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/ingress', { params })
  }
}
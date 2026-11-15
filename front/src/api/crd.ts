import request, { type PageResponse } from '@/utils/request'

// CRD 列表项
export interface CrdVO {
  name: string
  kind: string
  group: string
  version: string
  scope: string
  createTime: string
}

// CRD 查询参数
export interface CrdQueryDTO {
  clusterId: string
  page?: number
  pageSize?: number
  keyword?: string
}

export const crdApi = {
  // 获取 CRD 列表
  getCrds: (params: CrdQueryDTO) => {
    return request.get<PageResponse<CrdVO>>('/crd/list', { params })
  },

  // 获取 CRD 原始对象(用于渲染 YAML)
  getCrd: (clusterId: string, name: string) => {
    return request.get('/crd', { params: { clusterId, name } })
  },

  // 删除 CRD(连带其下所有自定义资源)
  deleteCrd: (clusterId: string, name: string) => {
    return request.delete('/crd', { params: { clusterId, name } })
  }
}

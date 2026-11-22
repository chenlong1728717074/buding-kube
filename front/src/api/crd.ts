import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// CRD 概要(不含 schema); 完整定义通过 /resource/yaml 按需查看
export interface CrdVO {
  name: string
  kind: string
  listKind: string
  plural: string
  singular: string
  shortNames: string[]
  group: string
  // 存储版本
  version: string
  // 可用的版本, 存储版本排在第一个
  versions: string[]
  scope: string
  createTime: string
}

export interface CrdQueryDTO {
  clusterId: string
  page?: number
  pageSize?: number
  keyword?: string
  // 忽略后端缓存重新查询
  refresh?: boolean
}

export const crdApi = {
  // CRD 列表: 后端缓存 1 分钟, 翻页和搜索不会重新查询集群; 首次加载 CRD 很多时可能较慢, 放宽超时
  getCrds: (params: CrdQueryDTO): Promise<ApiResponse<PageResponse<CrdVO>>> => request.get('/crd/list', { params, timeout: 60000 }),

  // 单个 CRD 的概要
  getCrd: (clusterId: string, name: string): Promise<ApiResponse<CrdVO>> => request.get('/crd', { params: { clusterId, name } }),

  // 删除 CRD(连带其下所有自定义资源)
  deleteCrd: (clusterId: string, name: string) => request.delete('/crd', { params: { clusterId, name } })
}

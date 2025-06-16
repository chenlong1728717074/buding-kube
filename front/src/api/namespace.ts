import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 命名空间基础DTO
export interface NamespaceBaseDTO {
  clusterId: string
  namespace?: string
  force?: boolean
}

// 命名空间分页查询DTO
export interface NamespacePageQueryDTO extends NamespaceBaseDTO {
  page?: number
  pageSize?: number
  keyword?: string
}

// 命名空间创建DTO
export interface NamespaceCreateDTO {
  clusterId: string
  namespace: string
  alias?: string
  describe?: string
  annotations?: Record<string, string>
  yaml?: string
}

// 命名空间应用DTO
export interface NamespaceApplyDTO {
  clusterId: string
  yaml: string
}

// 命名空间VO
export interface NamespaceVO {
  name: string
  alias?: string
  describe?: string
  generation?: number
  version?: string
  active: string
  uid: string
  resourceVersion: string
  annotations: Record<string, string>
  labels: Record<string, string>
  creationTimestamp: string
  metadata: any
  spec: any
  yaml?: string
  resources?: any
}

// 命名空间API
export const namespaceApi = {
  // 获取命名空间列表
  getList(params: NamespacePageQueryDTO): Promise<ApiResponse<PageResponse<NamespaceVO>>> {
    return request.get('/namespace/list', { params })
  },

  // 获取命名空间详情
  getInfo(params: NamespaceBaseDTO): Promise<ApiResponse<NamespaceVO>> {
    return request.get('/namespace', { params })
  },

  // 创建命名空间
  create(data: NamespaceCreateDTO): Promise<ApiResponse<void>> {
    return request.post('/namespace', data)
  },

  // 更新命名空间
  update(data: NamespaceCreateDTO): Promise<ApiResponse<void>> {
    return request.put('/namespace', data)
  },

  // 删除命名空间
  delete(params: NamespaceBaseDTO): Promise<ApiResponse<void>> {
    return request.delete('/namespace', { params })
  },

  // 应用YAML
  apply(data: NamespaceApplyDTO): Promise<ApiResponse<void>> {
    return request.post('/namespace/apply', data)
  }
}

export default namespaceApi
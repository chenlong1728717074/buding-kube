import request, { type ApiResponse, type PageResponse } from '@/utils/request'

export interface ConfigMapPageQueryDTO {
  clusterId: string
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export interface ConfigMapBaseDTO {
  clusterId: string
  namespace: string
  name: string
  force?: boolean
}

export interface ConfigMapInfoUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
}

export interface ConfigMapDataUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  data: Record<string, string>
}

export interface ConfigMapApplyDTO {
  clusterId: string
  yaml: string
}

export interface ConfigMapVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  data?: Record<string, string>
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
  yaml?: string
}

export const configMapApi = {
  getList(params: ConfigMapPageQueryDTO): Promise<ApiResponse<PageResponse<ConfigMapVO>>> {
    return request.get('/configMap/list', { params })
  },
  updateInfo(data: ConfigMapInfoUpdateDTO): Promise<ApiResponse<void>> {
    return request.put('/configMap', data)
  },
  updateData(data: ConfigMapDataUpdateDTO): Promise<ApiResponse<void>> {
    return request.put('/configMap/data', data)
  },
  applyYaml(data: ConfigMapApplyDTO): Promise<ApiResponse<void>> {
    return request.post('/configMap/apply', data)
  },
  delete(params: ConfigMapBaseDTO): Promise<ApiResponse<void>> {
    return request.delete('/configMap', { params })
  }
}

export default configMapApi

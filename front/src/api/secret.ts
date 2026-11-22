import request, { type ApiResponse, type PageResponse } from '@/utils/request'

export interface SecretPageQueryDTO {
  clusterId: string
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export interface SecretBaseDTO {
  clusterId: string
  namespace: string
  name: string
  force?: boolean
}

export interface SecretInfoUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
}

export interface SecretDataUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  stringData: Record<string, string>
}

export interface SecretCreateDTO {
  clusterId: string
  namespace: string
  name: string
  type?: string
  alias?: string
  describe?: string
  stringData?: Record<string, string>
  annotations?: Record<string, string>
  labels?: Record<string, string>
}

export interface SecretVO {
  name: string
  namespace: string
  type?: string
  alias?: string
  describe?: string
  data?: Record<string, string>
  stringData?: Record<string, string>
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
}

export const secretApi = {
  getList(params: SecretPageQueryDTO): Promise<ApiResponse<PageResponse<SecretVO>>> {
    // 全部命名空间的 Secret 较多时查询稍慢, 单独放宽超时
    return request.get('/secret/list', { params, timeout: 30000 })
  },
  getInfo(params: SecretBaseDTO): Promise<ApiResponse<SecretVO>> {
    return request.get('/secret', { params })
  },
  add(data: SecretCreateDTO): Promise<ApiResponse<void>> {
    return request.post('/secret/add', data)
  },
  updateInfo(data: SecretInfoUpdateDTO): Promise<ApiResponse<void>> {
    return request.put('/secret', data)
  },
  updateData(data: SecretDataUpdateDTO): Promise<ApiResponse<void>> {
    return request.put('/secret/data', data)
  },
  delete(params: SecretBaseDTO): Promise<ApiResponse<void>> {
    return request.delete('/secret', { params })
  }
}

export default secretApi


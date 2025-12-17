import request, { type ApiResponse, type PageResponse } from '@/utils/request'

export interface StorageClassPageQueryDTO {
  clusterId: string
  page?: number
  pageSize?: number
  keyword?: string
}

export interface StorageClassVO {
  name: string
  provisioner?: string
  parameters?: Record<string, string>
  reclaimPolicy?: string
  volumeBindingMode?: string
  allowVolumeExpansion?: boolean
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
  createTime?: string
  yaml?: string
}

export const storageClassApi = {
  getList(params: StorageClassPageQueryDTO): Promise<ApiResponse<PageResponse<StorageClassVO>>> {
    return request.get('/storageClass/list', { params })
  }
}

export default storageClassApi


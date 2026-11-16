import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// PersistentVolume 查询参数
export interface PVQueryDTO {
  clusterId: string
  page?: number
  pageSize?: number
  keyword?: string
}

// PersistentVolume 列表项
export interface PVVO {
  name: string
  namespace?: string
  capacity?: Record<string, string>
  accessModes?: string[]
  persistentVolumeReclaimPolicy?: string
  storageClassName?: string
  volumeMode?: string
  phase?: string
  message?: string
  reason?: string
  creationTimestamp?: string
  createTime?: string
  yaml?: string
}

export const pvApi = {
  // 获取PersistentVolume列表
  getPVList: (params: PVQueryDTO): Promise<ApiResponse<PageResponse<PVVO>>> => {
    return request.get('/pv/list', { params })
  },

  // 删除PersistentVolume
  deletePV: (params: { clusterId: string; name: string }) => {
    return request.delete('/pv', { params })
  }
}

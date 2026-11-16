import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// PersistentVolumeClaim 查询参数
export interface PVCQueryDTO {
  clusterId: string
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// PersistentVolumeClaim 列表项
export interface PVCVO {
  name: string
  namespace: string
  accessModes?: string[]
  storageClassName?: string
  volumeMode?: string
  volumeName?: string
  phase?: string
  capacity?: Record<string, string>
  creationTimestamp?: string
  createTime?: string
  yaml?: string
}

export const pvcApi = {
  // 获取PersistentVolumeClaim列表
  getPVCList: (params: PVCQueryDTO): Promise<ApiResponse<PageResponse<PVCVO>>> => {
    return request.get('/pvc/list', { params })
  },

  // 删除PersistentVolumeClaim
  deletePVC: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/pvc', { params })
  }
}

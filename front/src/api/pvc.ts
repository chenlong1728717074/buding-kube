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
  // 申请的容量(未绑定时 capacity 为空)
  resources?: { requests?: Record<string, string> }
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp?: string
  createTime?: string
  yaml?: string
}

export interface PVCCreateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  // 申请容量, 如 10Gi
  capacity: string
  accessModes: string[]
  // 不传时使用集群默认存储类; 传空字符串表示不使用存储类
  storageClassName?: string
  // 绑定已有持久卷时填写
  volumeName?: string
  volumeMode?: 'Filesystem' | 'Block'
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

export const pvcApi = {
  // 创建PersistentVolumeClaim
  create: (data: PVCCreateDTO): Promise<ApiResponse<void>> => {
    return request.post('/pvc', data)
  },

  // 获取PersistentVolumeClaim列表
  getPVCList: (params: PVCQueryDTO): Promise<ApiResponse<PageResponse<PVCVO>>> => {
    return request.get('/pvc/list', { params })
  },

  // 删除PersistentVolumeClaim
  deletePVC: (params: { clusterId: string; namespace: string; name: string }) => {
    return request.delete('/pvc', { params })
  }
}

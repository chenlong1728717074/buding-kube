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
  alias?: string
  describe?: string
  // 存储来源: nfs / hostPath / local / csi 等
  sourceType?: string
  // 绑定的持久卷声明
  claimNamespace?: string
  claimName?: string
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

export type PVSourceType = 'nfs' | 'hostPath' | 'local' | 'csi'

export interface PVCreateDTO {
  clusterId: string
  name: string
  alias?: string
  describe?: string
  // 容量, 如 10Gi
  capacity: string
  accessModes: string[]
  reclaimPolicy?: 'Retain' | 'Delete' | 'Recycle'
  storageClassName?: string
  volumeMode?: 'Filesystem' | 'Block'
  mountOptions?: string[]
  labels?: Record<string, string>
  sourceType: PVSourceType
  nfs?: { server: string; path: string; readOnly?: boolean }
  hostPath?: { path: string; type?: string }
  local?: { path: string; node: string; fsType?: string }
  csi?: { driver: string; volumeHandle: string; fsType?: string; readOnly?: boolean; volumeAttributes?: Record<string, string> }
}

export const pvApi = {
  // 创建PersistentVolume
  create: (data: PVCreateDTO): Promise<ApiResponse<void>> => {
    return request.post('/pv', data)
  },

  // 获取PersistentVolume列表
  getPVList: (params: PVQueryDTO): Promise<ApiResponse<PageResponse<PVVO>>> => {
    return request.get('/pv/list', { params })
  },

  // 删除PersistentVolume
  deletePV: (params: { clusterId: string; name: string }) => {
    return request.delete('/pv', { params })
  }
}

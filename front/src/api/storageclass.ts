import type { PVCVO } from '@/api/pvc'
import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 存储类型标识, 空字符串表示按供应者自动识别
export type StorageTypeKey = '' | 'block' | 'file' | 'object'

export interface StorageClassPageQueryDTO {
  clusterId: string
  page?: number
  pageSize?: number
  keyword?: string
  // 支持 block / file / object 或 未分类
  storageType?: string
}

export interface StorageClassVO {
  name: string
  alias?: string
  describe?: string
  provisioner?: string
  // 中文名: 块存储 / 文件存储 / 对象存储 / 未分类
  storageType?: string
  storageTypeKey?: StorageTypeKey
  // 为 true 表示存储类型是按供应者自动识别的
  storageTypeAuto?: boolean
  isDefault?: boolean
  pvcCount?: number
  allowVolumeClone?: boolean
  accessModes?: string[]
  mountOptions?: string[]
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

// 详情: 附带使用该存储类的持久卷声明
export interface StorageClassDetailVO extends StorageClassVO {
  persistentVolumeClaims: PVCVO[]
  // 使用该存储类的持久卷数量
  pvCount: number
}

export interface StorageClassCreateDTO {
  clusterId: string
  name: string
  alias?: string
  describe?: string
  storageType?: StorageTypeKey
  provisioner: string
  parameters?: Record<string, string>
  reclaimPolicy?: 'Delete' | 'Retain'
  volumeBindingMode?: 'Immediate' | 'WaitForFirstConsumer'
  allowVolumeExpansion?: boolean
  allowVolumeClone?: boolean
  accessModes?: string[]
  mountOptions?: string[]
  isDefault?: boolean
}

// 供应者、参数、回收策略、绑定模式创建后不可修改(Kubernetes 限制)
export interface StorageClassUpdateDTO {
  clusterId: string
  name: string
  alias?: string
  describe?: string
  storageType?: StorageTypeKey
  allowVolumeExpansion?: boolean
  allowVolumeClone?: boolean
  isDefault?: boolean
}

export const storageClassApi = {
  getList(params: StorageClassPageQueryDTO): Promise<ApiResponse<PageResponse<StorageClassVO>>> {
    return request.get('/storageClass/list', { params })
  },

  getDetail(params: { clusterId: string; name: string }): Promise<ApiResponse<StorageClassDetailVO>> {
    return request.get('/storageClass', { params })
  },

  create(data: StorageClassCreateDTO): Promise<ApiResponse<void>> {
    return request.post('/storageClass', data)
  },

  update(data: StorageClassUpdateDTO): Promise<ApiResponse<void>> {
    return request.put('/storageClass', data)
  },

  // 通过 YAML 创建或更新
  apply(data: { clusterId: string; yaml: string }): Promise<ApiResponse<void>> {
    return request.post('/storageClass/apply', data)
  },

  // 删除StorageClass
  deleteStorageClass: (params: { clusterId: string; name: string }) => {
    return request.delete('/storageClass', { params })
  }
}

export default storageClassApi

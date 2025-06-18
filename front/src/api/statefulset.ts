import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// StatefulSet查询参数
export interface StatefulSetQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  status?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// StatefulSet视图对象
export interface StatefulSetVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  yaml?: string
  status: string
  containers: number
  replicas: number
  readyReplicas: number
  updateTime?: string
  createTime: string
}

// StatefulSet API
export const statefulSetApi = {
  // 获取StatefulSet列表
  getStatefulSets: (params: StatefulSetQueryDTO): Promise<ApiResponse<PageResponse<StatefulSetVO>>> => {
    return request.get('/statefulSet/list', { params })
  }
}
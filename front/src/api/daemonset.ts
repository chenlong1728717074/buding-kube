import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// DaemonSet查询参数
export interface DaemonSetQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  status?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// DaemonSet视图对象
export interface DaemonSetVO {
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

// DaemonSet API
export const daemonSetApi = {
  // 获取DaemonSet列表
  getDaemonSets: (params: DaemonSetQueryDTO): Promise<ApiResponse<PageResponse<DaemonSetVO>>> => {
    return request.get('/daemonSet/list', { params })
  }
}
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
  },
  
  // 删除StatefulSet
  deleteStatefulSet: (data: { clusterId: string; namespace: string; name: string }): Promise<ApiResponse<any>> => {
    return request.delete('/statefulSet', { data })
  },
  
  // 更新StatefulSet（修改别名和描述）
  updateStatefulSet: (data: { clusterId: string; namespace: string; name: string; alias: string; describe: string }): Promise<ApiResponse<any>> => {
    return request.put('/statefulSet', data)
  },
  
  // 重建StatefulSet
  rolloutStatefulSet: (data: { clusterId: string; namespace: string; name: string }): Promise<ApiResponse<any>> => {
    return request.put('/statefulSet/rollout', data)
  },
  
  // 应用YAML（创建或修改）
  applyStatefulSet: (data: { clusterId: string; namespace?: string; name?: string; yaml: string }): Promise<ApiResponse<any>> => {
    return request.post('/statefulSet/apply', data)
  }
}
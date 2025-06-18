import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// Deployment查询参数
export interface DeploymentQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  status?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// Deployment视图对象
export interface DeploymentVO {
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

// Deployment API
export const deploymentApi = {
  // 获取Deployment列表
  getDeployments: (params: DeploymentQueryDTO): Promise<ApiResponse<PageResponse<DeploymentVO>>> => {
    return request.get('/deployment/list', { params })
  }
}
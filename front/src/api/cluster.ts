import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 集群信息
export interface ClusterVO {
  id?: string
  name: string
  alias?: string
  describe?: string
  apiServer?: string
  version?: string
  status?: string
  nodeCount?: number
  namespaceCount?: number
  podCount?: number
  deploymentCount?: number
  createTime?: string
  updateTime?: string
}

// 集群查询参数
export interface ClusterQueryDTO {
  page?: number
  pageSize?: number
  keyword?: string
  status?: string
  version?: string
}

// 集群创建/更新参数(config 与 uri+token 二选一,更新时留空表示沿用原配置)
export interface ClusterSaveDTO {
  name: string
  alias?: string
  describe?: string
  config?: string
  uri?: string
  token?: string
}

// 集群管理API
export const clusterApi = {
  // 获取集群列表
  getClusters: (params: ClusterQueryDTO): Promise<ApiResponse<PageResponse<ClusterVO>>> => {
    return request.get('/cluster/list', { params })
  },

  // 获取单个集群
  getCluster: (name: string): Promise<ApiResponse<ClusterVO>> => {
    return request.get(`/cluster/${name}`)
  },

  // 创建集群
  createCluster: (data: ClusterSaveDTO): Promise<ApiResponse<ClusterVO>> => {
    return request.post('/cluster', data)
  },

  // 更新集群
  updateCluster: (name: string, data: ClusterSaveDTO): Promise<ApiResponse<ClusterVO>> => {
    return request.put(`/cluster/${name}`, data)
  },

  // 删除集群
  deleteCluster: (name: string): Promise<ApiResponse<void>> => {
    return request.delete(`/cluster/${name}`)
  },

  // 获取节点列表
  getNodes: (clusterId: string, params?: any): Promise<ApiResponse<PageResponse<any>>> => {
    return request.get(`/node/list`, { params: { ...params, clusterId } })
  },

  // 获取节点详情
  getNodeDetail: (clusterId: string, hostname: string): Promise<ApiResponse<any>> => {
    return request.get('/node', {
      params: {
        clusterId,
        hostname
      }
    })
  },

  // 切换节点调度状态
  toggleNodeSchedule: (clusterId: string, hostname: string, unSchedule: boolean) => {
    return request.put('/node/unSchedule', {
      clusterId,
      hostname,
      unSchedule
    })
  },

  // 删除Node对象(节点重新注册后会再次出现)
  deleteNode: (clusterId: string, name: string) => {
    return request.delete('/node', { params: { clusterId, name } })
  }
}

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
  // uri+token 方式: API Server 的 CA 证书; insecure 为 true 时跳过证书校验
  caData?: string
  insecure?: boolean
}

// 集群资源使用单项指标
export interface ResourceMetricVO {
  used: number
  total: number
  percent: number
  usedText: string
  totalText: string
}

// 集群节点资源使用率汇总
export interface ClusterResourceUsageVO {
  cpu: ResourceMetricVO
  memory: ResourceMetricVO
  storage: ResourceMetricVO
  nodeTotal: number
  nodeReady: number
  podTotal: number
  podRunning: number
  source: 'metrics-server' | 'requests' | string
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

  // 获取集群节点资源使用情况
  getClusterUsage: (clusterId: string): Promise<ApiResponse<ClusterResourceUsageVO>> => {
    return request.get('/node/usage', { params: { clusterId } })
  }
}

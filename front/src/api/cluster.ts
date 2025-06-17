import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 集群信息
export interface ClusterVO {
  id?: string
  name: string
  alias?: string
  endpoint: string
  version: string
  status: string
  nodeCount: number
  namespaceCount: number
  podCount: number
  createTime: string
  updateTime: string
  description?: string
  config?: string
}

// 集群查询参数
export interface ClusterQueryDTO {
  page?: number
  pageSize?: number
  keyword?: string
  status?: string
  version?: string
}

// 创建集群参数
export interface CreateClusterDTO {
  name: string
  alias?: string
  describe?: string
  config: string
}

// 更新集群参数
export interface UpdateClusterDTO {
  id?: string
  name: string
  alias?: string
  describe?: string
  config: string
}

// 集群详情返回值
export interface ClusterDetailDTO {
  id?: string
  name: string
  alias?: string
  describe?: string
  config: string
}

// 集群管理API
export const clusterApi = {
  // 获取集群列表
  getClusters: (params: ClusterQueryDTO) => {
    return request.get<PageResponse<ClusterVO>>('/cluster/list', { params })
  },

  // 获取单个集群
  getCluster: (id: string) => {
    return request.get<ClusterVO>(`/cluster/${id}`)
  },

  // 创建集群
  createCluster: (data: CreateClusterDTO) => {
    return request.post<ClusterVO>('/cluster', data)
  },

  // 更新集群
  updateCluster: (name: string, data: UpdateClusterDTO) => {
    return request.put<ClusterVO>(`/cluster/${name}`, data)
  },

  // 删除集群
  deleteCluster: (name: string) => {
    return request.delete(`/cluster/${name}`)
  },

  // 测试集群连接
  testConnection: (name: string) => {
    return request.post<{ connected: boolean; message: string }>(`/cluster/${name}/test`, {})
  },

  // 获取集群统计信息
  getStats: (name: string) => {
    return request.get(`/cluster/${name}/stats`)
  },

  // 获取集群事件
  getEvents: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/events`, { params })
  },

  // 获取节点列表
  getNodes: (clusterId: string, params?: any) => {
    return request.get(`/node/list`, { params: { ...params, clusterId } })
  },

  // 获取节点详情
  getNodeDetail: (clusterId: string, hostname: string) => {
    return request.get('/node', {
      params: {
        clusterId,
        hostname
      }
    })
  },

  // 禁止/允许节点调度
  cordonNode: (clusterName: string, nodeName: string, cordon: boolean) => {
    return request.post(`/cluster/${clusterName}/nodes/${nodeName}/cordon`, { cordon })
  },

  // 驱逐节点Pod
  drainNode: (clusterName: string, nodeName: string) => {
    return request.post(`/cluster/${clusterName}/nodes/${nodeName}/drain`)
  },

  // 切换节点调度状态
  toggleNodeSchedule: (clusterId: string, hostname: string, unSchedule: boolean) => {
    return request.put('/node/unSchedule', {
      clusterId,
      hostname,
      unSchedule
    })
  },

  // 获取Pod列表
  getPods: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/pods`, { params })
  },

  // 获取Pod详情
  getPodDetail: (clusterName: string, namespace: string, podName: string) => {
    return request.get(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}`)
  },

  // 删除Pod
  deletePod: (clusterName: string, namespace: string, podName: string) => {
    return request.delete(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}`)
  },

  // 获取Pod日志
  getPodLogs: (clusterName: string, namespace: string, podName: string, container: string, params?: any) => {
    return request.get(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}/logs`, {
      params: { container, ...params }
    })
  },

  // 获取Pod YAML
  getPodYaml: (clusterName: string, namespace: string, podName: string) => {
    return request.get(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}/yaml`)
  },

  // 获取Pod事件
  getPodEvents: (clusterName: string, namespace: string, podName: string) => {
    return request.get(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}/events`)
  },

  // 进入Pod容器
  execPod: (clusterName: string, namespace: string, podName: string, container: string) => {
    return request.post(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}/exec`, { container })
  },

  // 获取命名空间列表
  getNamespaces: (name: string) => {
    return request.get(`/cluster/${name}/namespaces`)
  },

  // 获取服务列表
  getServices: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/services`, { params })
  },

  // 获取部署列表
  getDeployments: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/deployments`, { params })
  },

  // 获取配置映射列表
  getConfigMaps: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/configmaps`, { params })
  },

  // 获取密钥列表
  getSecrets: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/secrets`, { params })
  },

  // 获取持久卷列表
  getPersistentVolumes: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/persistentvolumes`, { params })
  },

  // 获取存储类列表
  getStorageClasses: (name: string) => {
    return request.get(`/cluster/${name}/storageclasses`)
  },

  // 获取入口列表
  getIngresses: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/ingresses`, { params })
  },

  // 获取监控指标
  getMetrics: (name: string, params?: any) => {
    return request.get(`/cluster/${name}/metrics`, { params })
  },

  // 获取节点监控指标
  getNodeMetrics: (clusterName: string, nodeName: string, params?: any) => {
    return request.get(`/cluster/${clusterName}/nodes/${nodeName}/metrics`, { params })
  },

  // 获取Pod监控指标
  getPodMetrics: (clusterName: string, namespace: string, podName: string, params?: any) => {
    return request.get(`/cluster/${clusterName}/namespaces/${namespace}/pods/${podName}/metrics`, { params })
  }
}

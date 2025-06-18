import { clusterApi } from '@/api/cluster'
import { namespaceApi } from '@/api/namespace'
import type { ClusterVO } from '@/api/cluster'
import type { NamespaceVO } from '@/api/namespace'

// 集群数据获取函数
export const useClusterFetcher = () => {
  return async (params: { page: number; pageSize: number; keyword?: string }) => {
    const queryParams = {
      page: params.page,
      pageSize: params.pageSize,
      name: params.keyword || undefined
    }
    
    const response = await clusterApi.getClusters(queryParams)
    return {
      data: response.data.items || [],
      total: response.data.total || 0
    }
  }
}

// 命名空间数据获取函数
export const useNamespaceFetcher = (clusterId?: string) => {
  return async (params: { page: number; pageSize: number; keyword?: string }) => {
    if (!clusterId) {
      return { data: [], total: 0 }
    }
    
    const queryParams = {
      page: params.page,
      pageSize: params.pageSize,
      clusterId,
      name: params.keyword || undefined
    }
    
    const response = await namespaceApi.getList(queryParams)
    return {
      data: response.data.items || [],
      total: response.data.total || 0
    }
  }
}

// 集群选项配置
export const clusterSelectConfig = {
  optionKey: 'id',
  optionLabel: 'name',
  optionValue: 'id',
  placeholder: '请选择集群',
  pageSize: 20
}

// 命名空间选项配置
export const namespaceSelectConfig = {
  optionKey: 'name',
  optionLabel: 'name', 
  optionValue: 'name',
  placeholder: '请选择命名空间',
  pageSize: 20
}
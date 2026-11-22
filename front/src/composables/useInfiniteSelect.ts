import { namespaceApi } from '@/api/namespace'

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

// 命名空间选项配置
export const namespaceSelectConfig = {
  optionKey: 'name',
  optionLabel: 'name', 
  optionValue: 'name',
  placeholder: '请选择命名空间',
  pageSize: 20
}
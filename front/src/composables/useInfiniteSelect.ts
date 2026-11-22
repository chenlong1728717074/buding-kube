import { namespaceApi } from '@/api/namespace'
import type { OptionItem } from '@/components/InfiniteSelect.vue'

// 命名空间分页获取函数, 配合 InfiniteSelect 使用; 关键字在后端按名称过滤
export const useNamespaceFetcher = (clusterId?: string) => {
  return async (params: { page: number; pageSize: number; keyword?: string }): Promise<{ data: OptionItem[]; total: number }> => {
    if (!clusterId) {
      return { data: [], total: 0 }
    }
    const response = await namespaceApi.getList({
      clusterId,
      page: params.page,
      pageSize: params.pageSize,
      keyword: params.keyword || undefined
    })
    return {
      data: (response.data.items || []) as unknown as OptionItem[],
      total: response.data.total || 0
    }
  }
}

// 命名空间选项配置: 可输入搜索, 每页 20 个
export const namespaceSelectConfig = {
  optionKey: 'name',
  optionLabel: 'name',
  optionValue: 'name',
  placeholder: '请选择命名空间',
  filterable: true,
  remote: true,
  pageSize: 20
}

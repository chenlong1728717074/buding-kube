import request, { type ApiResponse, type PageResponse } from '@/utils/request'
import type { EventVO } from '@/api/job'

export interface EventQuery {
  clusterId: string
  namespace?: string
  // Normal / Warning
  type?: string
  // 涉及对象的类型、名称、UID
  kind?: string
  name?: string
  uid?: string
  keyword?: string
  page: number
  pageSize: number
}

export const eventApi = {
  // 按最后发生时间倒序
  list: (params: EventQuery): Promise<ApiResponse<PageResponse<EventVO>>> => request.get('/event/list', { params })
}

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
  // 返回全部事件; 默认只保留最近 1000 条
  all?: boolean
  // 忽略后端的翻页缓存
  refresh?: boolean
  page: number
  pageSize: number
}

export interface EventPage extends PageResponse<EventVO> {
  // 符合条件的事件总数
  matched: number
  // 为 true 时只返回了最近 limit 条
  limited: boolean
  limit: number
}

export const eventApi = {
  // 按最后发生时间倒序
  list: (params: EventQuery): Promise<ApiResponse<EventPage>> => request.get('/event/list', { params })
}

import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/types/api'

// CronJob查询参数
export interface CronJobQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// CronJob视图对象
export interface CronJobVO {
  name: string
  namespace: string
  schedule: string
  suspend?: boolean
  concurrencyPolicy: string
  startingDeadlineSeconds?: number
  successfulJobsHistoryLimit?: number
  failedJobsHistoryLimit?: number
  lastScheduleTime?: string
  generation: number
  version: string
  uid: string
  resourceVersion: string
  annotations?: Record<string, string>
  labels?: Record<string, string>
  creationTimestamp: string
  createTime: string
  yaml?: string
}

// CronJob API
export const cronJobApi = {
  // 获取CronJob列表
  getCronJobs: (params: CronJobQueryDTO): Promise<ApiResponse<PageResponse<CronJobVO>>> => {
    return request.get('/cronJob/list', { params })
  }
}
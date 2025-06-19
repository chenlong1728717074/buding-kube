import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/types/api'

// Job查询参数
export interface JobQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// Job视图对象
export interface JobVO {
  name: string
  namespace: string
  parallelism?: number
  completions?: number
  activeDeadlineSeconds?: number
  backoffLimit?: number
  ttlSecondsAfterFinished?: number
  active: number
  succeeded: number
  failed: number
  startTime?: string
  completionTime?: string
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

// Job API
export const jobApi = {
  // 获取Job列表
  getJobs: (params: JobQueryDTO): Promise<ApiResponse<PageResponse<JobVO>>> => {
    return request.get('/job/list', { params })
  }
}
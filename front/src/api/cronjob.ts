import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'
import type { ContainerTemplateVO, EventVO, JobPodVO, JobTemplateDTO, JobTarget, JobVO } from '@/api/job'

export type ConcurrencyPolicy = 'Allow' | 'Forbid' | 'Replace'

// CronJob查询参数
export interface CronJobQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  // Active(运行中) / Suspended(已暂停)
  status?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// CronJob列表项
export interface CronJobVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  schedule: string
  timeZone?: string
  concurrencyPolicy: ConcurrencyPolicy
  suspend: boolean
  // 正在运行的任务数
  active: number
  createTime: string
  lastRunTime?: string | null
  lastSuccessfulTime?: string | null
}

// CronJob详情
export interface CronJobDetailVO extends CronJobVO {
  startingDeadlineSeconds?: number | null
  successfulJobsHistoryLimit: number
  failedJobsHistoryLimit: number
  completions: number
  parallelism: number
  backoffLimit: number
  activeDeadlineSeconds?: number | null
  ttlSecondsAfterFinished?: number | null
  restartPolicy?: string
  creator?: string
  uid: string
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  containers: ContainerTemplateVO[]
  // 定时任务创建的任务
  jobs: JobVO[]
  pods: JobPodVO[]
  events: EventVO[]
}

export interface CronJobCreateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  labels?: Record<string, string>
  schedule: string
  timeZone?: string
  concurrencyPolicy?: ConcurrencyPolicy
  startingDeadlineSeconds?: number
  successfulJobsHistoryLimit?: number
  failedJobsHistoryLimit?: number
  suspend?: boolean
  template: JobTemplateDTO
}

export interface CronJobUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  schedule?: string
  // 空字符串表示清除时区
  timeZone?: string
  concurrencyPolicy?: ConcurrencyPolicy
  suspend?: boolean
  // 0 表示不限制
  startingDeadlineSeconds?: number
  successfulJobsHistoryLimit?: number
  failedJobsHistoryLimit?: number
}

// CronJob API
export const cronJobApi = {
  // 获取CronJob列表
  getCronJobs: (params: CronJobQueryDTO): Promise<ApiResponse<PageResponse<CronJobVO>>> => {
    return request.get('/cronJob/list', { params })
  },

  getDetail: (params: JobTarget): Promise<ApiResponse<CronJobDetailVO>> => {
    return request.get('/cronJob', { params })
  },

  create: (data: CronJobCreateDTO): Promise<ApiResponse<void>> => {
    return request.post('/cronJob', data)
  },

  update: (data: CronJobUpdateDTO): Promise<ApiResponse<void>> => {
    return request.put('/cronJob', data)
  },

  // 暂停 / 启用
  suspend: (data: JobTarget & { suspend: boolean }): Promise<ApiResponse<void>> => {
    return request.put('/cronJob/suspend', data)
  },

  // 立即执行一次, 返回创建的任务名
  trigger: (data: JobTarget): Promise<ApiResponse<{ jobName: string }>> => {
    return request.post('/cronJob/trigger', data)
  },

  delete: (params: JobTarget): Promise<ApiResponse<void>> => {
    return request.delete('/cronJob', { params })
  }
}

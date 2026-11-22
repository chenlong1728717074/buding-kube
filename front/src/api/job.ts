import { request } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'
import type { PodResources, PodUsage } from '@/utils/resource'

export type JobStatus = 'Pending' | 'Running' | 'Complete' | 'Failed' | 'Suspended'

// Job查询参数
export interface JobQueryDTO {
  clusterId: string
  namespace?: string
  name?: string
  status?: string
  page?: number
  pageSize?: number
  keyword?: string
}

// Job列表项
export interface JobVO {
  name: string
  namespace: string
  alias?: string
  describe?: string
  status: JobStatus
  message?: string
  completions: number
  parallelism: number
  backoffLimit: number
  active: number
  succeeded: number
  failed: number
  // 所属定时任务
  ownerCronJob?: string
  // 是否由定时任务手动触发
  manual?: boolean
  startTime?: string | null
  completionTime?: string | null
  createTime: string
  lastRunTime?: string | null
}

// 容器模板(详情展示)
export interface ContainerEnvVO {
  name: string
  value: string
  // 值来自 ConfigMap / Secret 等引用时的说明
  source?: string
}

export interface ContainerTemplateVO {
  name: string
  init: boolean
  image: string
  imagePullPolicy?: string
  command?: string[] | null
  args?: string[] | null
  workingDir?: string
  envs: ContainerEnvVO[]
  requests?: Record<string, string> | null
  limits?: Record<string, string> | null
  // 以下字段工作负载才有
  ports?: { name: string; containerPort: number; protocol: string }[]
  livenessProbe?: string
  readinessProbe?: string
  startupProbe?: string
  volumeMounts?: { name: string; mountPath: string; subPath?: string; readOnly: boolean }[]
}

// 运行记录
export interface JobRunRecordVO {
  run: number
  current: boolean
  status: JobStatus
  message?: string
  completions: number
  succeeded: number
  failed: number
  startTime?: string | null
  completionTime?: string | null
}

// 任务 / 工作负载 / 服务的容器组(含资源设置与实时用量)
export interface JobPodVO extends PodResources, PodUsage {
  name: string
  namespace: string
  jobName?: string
  status: string
  message?: string
  nodeName?: string
  podIP?: string
  ready: string
  restarts: number
  startTime?: string | null
  createTime: string
}

export interface EventVO {
  type: string
  reason: string
  // Kind/Name
  object: string
  message: string
  source: string
  count: number
  firstTime: string
  lastTime: string
  namespace?: string
  objectKind?: string
  objectName?: string
  objectNamespace?: string
}

// Job详情
export interface JobDetailVO extends JobVO {
  activeDeadlineSeconds?: number | null
  ttlSecondsAfterFinished?: number | null
  restartPolicy?: string
  suspend: boolean
  creator?: string
  uid: string
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  containers: ContainerTemplateVO[]
  records: JobRunRecordVO[]
  pods: JobPodVO[]
  events: EventVO[]
}

// 创建表单里的容器
export interface JobContainerDTO {
  name: string
  image: string
  imagePullPolicy?: string
  command?: string[]
  args?: string[]
  workingDir?: string
  env?: { name: string; value: string }[]
  cpuRequest?: string
  cpuLimit?: string
  memoryRequest?: string
  memoryLimit?: string
}

// 任务设置, Job 和 CronJob 共用
export interface JobTemplateDTO {
  completions?: number
  parallelism?: number
  backoffLimit?: number
  // 不大于 0 表示不限制
  activeDeadlineSeconds?: number
  // 不传表示不自动清理
  ttlSecondsAfterFinished?: number
  restartPolicy?: 'Never' | 'OnFailure'
  containers: JobContainerDTO[]
}

export interface JobCreateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  labels?: Record<string, string>
  template: JobTemplateDTO
}

export interface JobUpdateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  parallelism?: number
  // 0 表示不限制
  activeDeadlineSeconds?: number
  // -1 表示不自动清理
  ttlSecondsAfterFinished?: number
}

export interface JobTarget {
  clusterId: string
  namespace: string
  name: string
}

// Job API
export const jobApi = {
  // 获取Job列表
  getJobs: (params: JobQueryDTO): Promise<ApiResponse<PageResponse<JobVO>>> => {
    return request.get('/job/list', { params })
  },

  // 获取Job详情
  getDetail: (params: JobTarget): Promise<ApiResponse<JobDetailVO>> => {
    return request.get('/job', { params })
  },

  create: (data: JobCreateDTO): Promise<ApiResponse<void>> => {
    return request.post('/job', data)
  },

  update: (data: JobUpdateDTO): Promise<ApiResponse<void>> => {
    return request.put('/job', data)
  },

  // 重新执行: 按原规格重新创建任务
  rerun: (data: JobTarget): Promise<ApiResponse<void>> => {
    return request.post('/job/rerun', data)
  },

  delete: (params: JobTarget): Promise<ApiResponse<void>> => {
    return request.delete('/job', { params })
  }
}

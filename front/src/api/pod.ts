import request from '@/utils/request'
import { redirectToLogin } from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'
import type { ContainerTemplateVO, EventVO } from '@/api/job'
import type { PodResources, PodUsage } from '@/utils/resource'
import type { WorkloadConditionVO, WorkloadContainerDTO, WorkloadVolumeDTO, WorkloadVolumeVO } from '@/api/workload'

// Pod列表项(资源设置与实时用量见 PodResources / PodUsage)
export interface PodVO extends PodResources, PodUsage {
  creationTimestamp: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  // 阶段: Pending / Running / Succeeded / Failed / Unknown
  status: string
  // 与 kubectl get pod 的 STATUS 列一致, 例如 CrashLoopBackOff
  displayStatus: string
  message?: string
  ready: string
  restarts: number
  nodeName: string
  hostIP: string
  podIP: string
  // 所属工作负载, ReplicaSet 创建的显示为 Deployment
  ownerKind?: string
  ownerName?: string
}

// 容器运行状态
export interface PodContainerStatusVO {
  name: string
  init: boolean
  ready: boolean
  started: boolean
  restartCount: number
  // Running / Waiting / Terminated
  state: string
  reason?: string
  message?: string
  exitCode: number
  startedAt?: string | null
  finishedAt?: string | null
  lastReason?: string
  lastExitCode: number
  lastFinishedAt?: string | null
  imageID?: string
}

// 容器的资源设置与用量
export interface PodContainerUsageVO extends PodResources, PodUsage {
  name: string
  init: boolean
}

// Pod详情
export interface PodDetailVO extends PodResources, PodUsage {
  name: string
  namespace: string
  alias?: string
  describe?: string
  status: string
  displayStatus: string
  message?: string
  reason?: string
  ready: string
  restarts: number
  nodeName: string
  hostIP: string
  podIP: string
  podIPs: string[]
  qosClass: string
  restartPolicy: string
  serviceAccountName?: string
  priorityClassName?: string
  dnsPolicy?: string
  hostNetwork: boolean
  ownerKind?: string
  ownerName?: string
  creator?: string
  uid: string
  startTime?: string | null
  createTime: string
  deletionTime?: string | null
  labels?: Record<string, string> | null
  annotations?: Record<string, string> | null
  nodeSelector?: Record<string, string> | null
  tolerations: string[]
  conditions: WorkloadConditionVO[]
  containers: ContainerTemplateVO[]
  containerStatuses: PodContainerStatusVO[]
  volumes: WorkloadVolumeVO[]
  events: EventVO[]
  containerUsages: PodContainerUsageVO[]
}

// 分步创建
export interface PodCreateDTO {
  clusterId: string
  namespace: string
  name: string
  alias?: string
  describe?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  containers: WorkloadContainerDTO[]
  volumes?: WorkloadVolumeDTO[]
  restartPolicy?: 'Always' | 'OnFailure' | 'Never'
  nodeName?: string
  nodeSelector?: Record<string, string>
  serviceAccountName?: string
  imagePullSecrets?: string[]
  syncHostTimezone?: boolean
  terminationGracePeriodSeconds?: number
  hostNetwork?: boolean
}

// Pod查询参数
export interface PodQueryDTO {
  clusterId: string
  namespace?: string
  status?: string
  keyword?: string
  // createTime(默认) / cpu / memory / restarts / name
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  page: number
  pageSize: number
}

// Pod基础参数
export interface PodDTO {
  clusterId: string
  namespace: string
  name: string
  // 强制删除(不等待优雅终止), 用于一直处于删除中的容器组
  force?: boolean
}

// Pod日志查询参数
export interface PodLogDTO {
  clusterId: string
  namespace: string
  name: string
  containerName?: string
  follow?: boolean
  tailLines?: number
  sinceTime?: string
}

// Pod文件操作DTO
export interface PodFileDTO {
  clusterId: string
  namespace: string
  name: string
  containerName: string
  filePath: string
}

// Pod API
export const podApi = {
  // 获取Pod列表
  getList(params: PodQueryDTO): Promise<ApiResponse<PageResponse<PodVO>>> {
    return request.get('/pod/list', { params })
  },

  // 获取Pod详情
  getDetail(params: PodDTO): Promise<ApiResponse<PodDetailVO>> {
    return request.get('/pod', { params })
  },

  // 分步创建
  create(data: PodCreateDTO): Promise<ApiResponse<void>> {
    return request.post('/pod', data)
  },

  // 编辑别名和描述
  updateInfo(data: PodDTO & { alias: string; describe: string }): Promise<ApiResponse<void>> {
    return request.put('/pod', data)
  },

  // 更新镜像, 键为容器名
  updateImages(data: PodDTO & { images: Record<string, string> }): Promise<ApiResponse<void>> {
    return request.put('/pod/image', data)
  },

  // 删除Pod
  delete(params: PodDTO): Promise<ApiResponse<void>> {
    return request.delete('/pod', { params })
  },

  // 驱逐Pod
  expel(params: PodDTO): Promise<ApiResponse<void>> {
    return request.get('/pod/expel', { params })
  },

  // 获取Pod日志 - 流式处理
  async getLogs(params: PodLogDTO, onData: (data: string) => void, onError?: (error: Error) => void, signal?: AbortSignal): Promise<void> {
    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
    const token = sessionStorage.getItem('token') || localStorage.getItem('token')
    
    try {
      const response = await fetch(`${baseURL}/pod/logs`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'text/plain',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify(params),
        signal: signal // 添加AbortSignal支持
      })

      const contentType = response.headers.get('content-type') || ''

      if (response.ok && contentType.includes('application/json')) {
        const payload = await response.json().catch(() => null)
        const numericCode = Number(payload?.code)
        const msg = payload?.msg

        if (numericCode === 401) {
          await redirectToLogin(msg)
          throw new Error(msg || '未授权')
        }

        if (Number.isFinite(numericCode) && numericCode !== 200) {
          throw new Error(msg || '请求失败')
        }
      }
      
      if (response.status === 401) {
        await redirectToLogin()
        throw new Error('未授权')
      }
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('无法获取响应流')
      }
      
      const decoder = new TextDecoder()
      let buffer = ''
      
      try {
        while (true) {
          const { done, value } = await reader.read()
          
          if (done) {
            // 处理剩余的缓冲区数据
            if (buffer.trim()) {
              onData(buffer)
            }
            break
          }
          
          // 解码数据并添加到缓冲区
          buffer += decoder.decode(value, { stream: true })
          
          // 按行处理数据
          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // 保留最后一个不完整的行
          
          for (const line of lines) {
            if (line.trim()) {
              onData(line + '\n')
            }
          }
        }
      } finally {
        reader.releaseLock()
      }
    } catch (error) {
      console.error('获取日志流失败:', error)
      if (onError) {
        onError(error as Error)
      }
      throw error
    }
  },

  // 下载文件(大文件耗时可能较长, 不设置超时)
  async downloadFile(params: PodFileDTO): Promise<Blob> {
    const blob = await request.post<Blob, Blob>('/pod/download', params, {
      responseType: 'blob',
      timeout: 0
    })
    // 后端出错时返回的是 JSON({code, msg}), blob 模式下响应拦截器识别不了, 需要在这里解析,
    // 否则会把错误信息当成 tar 文件保存下来
    if (blob.type.includes('application/json')) {
      let msg = '下载文件失败'
      try {
        const payload = JSON.parse(await blob.text())
        msg = payload?.msg || msg
      } catch {
        // 忽略解析失败, 使用默认提示
      }
      throw new Error(msg)
    }
    return blob
  },

  // 上传文件
  uploadFile(params: PodFileDTO, file: File): Promise<ApiResponse<void>> {
    const formData = new FormData()
    formData.append('clusterId', params.clusterId)
    formData.append('namespace', params.namespace)
    formData.append('name', params.name)
    formData.append('containerName', params.containerName)
    formData.append('filePath', params.filePath)
    formData.append('file', file)
    
    // 大文件上传耗时可能较长, 不设置超时
    return request.post('/pod/upload', formData, {
      timeout: 0,
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

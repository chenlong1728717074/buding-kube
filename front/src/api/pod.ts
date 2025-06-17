import request from '@/utils/request'
import type { ApiResponse, PageResponse } from '@/utils/request'

// Pod基础信息
export interface PodVO {
  creationTimestamp: string
  namespace: string
  name: string
  status: string
  nodeName: string
  hostIP: string
  podIP: string
}

// Pod详情信息
export interface PodInfoVO {
  creationTimestamp: string
  namespace: string
  name: string
  status: string
  nodeName: string
  hostIP: string
  podIP: string
  yaml: string
  containers: PodContainerVO[]
  annotations: Record<string, string>
  labels: Record<string, string>
  metaData: any
  spec: any
  events: EventVO[]
}

// Pod容器信息
export interface PodContainerVO {
  name: string
  ready: boolean
  restartCount: number
  image: string
  started: boolean
  volumeMounts: any[]
  startedAt: string
  args: string[]
  envs: PodEnvVO[]
  ports: PodPortVO[]
}

// Pod环境变量
export interface PodEnvVO {
  name: string
  value: string
}

// Pod端口信息
export interface PodPortVO {
  name: string
  hostPort: number
  containerPort: number
  hostIP: string
  protocol: string
}

// 事件信息
export interface EventVO {
  type: string
  reason: string
  message: string
  firstTimestamp: string
  lastTimestamp: string
  count: number
}

// Pod查询参数
export interface PodQueryDTO {
  clusterId: string
  namespace: string
  status?: string
  page: number
  pageSize: number
}

// Pod基础参数
export interface PodDTO {
  clusterId: string
  namespace: string
  name: string
}

// Pod日志查询参数
export interface PodLogDTO {
  clusterId: string
  namespace: string
  name: string
  container?: string
  follow?: boolean
  tailLines?: number
  sinceTime?: string
}

// Pod API
export const podApi = {
  // 获取Pod列表
  getList(params: PodQueryDTO): Promise<ApiResponse<PageResponse<PodVO>>> {
    return request.get('/pod/list', { params })
  },

  // 获取Pod详情
  getInfo(params: PodDTO): Promise<ApiResponse<PodInfoVO>> {
    return request.get('/pod', { params })
  },

  // 删除Pod
  delete(params: PodDTO): Promise<ApiResponse<void>> {
    return request.delete('/pod', { params })
  },

  // 获取Pod日志 - 流式处理
  async getLogs(params: PodLogDTO, onData: (data: string) => void, onError?: (error: Error) => void, signal?: AbortSignal): Promise<void> {
    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
    const token = localStorage.getItem('token')
    
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
  }
}

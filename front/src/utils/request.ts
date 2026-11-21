import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 通用API响应格式
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页响应格式
export interface PageResponse<T = any> {
  code: number
  msg: string
  data: {
    list: T[]
    total: number
    page: number
    size: number
  }
}

// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json;charset=UTF-8'
  }
})

let redirectingToLogin = false

// UI 版式检查用 mock 开关: 仅当构建时设置 VITE_UI_MOCK=1 才生效, 正常构建不含此逻辑路径
const uiMockEnabled = import.meta.env.VITE_UI_MOCK === '1'
const uiMockHandler = (config: AxiosRequestConfig) => {
  const url = config.url || ''
  if (url.includes('/user/list')) {
    return Promise.resolve({
      data: {
        code: 200,
        msg: 'ok',
        data: {
          items: [
            { username: 'admin', realName: '管理员', email: '', role: 'super', status: 'active', enabled: true, department: '', phone: '', creationTimestamp: '2026-11-21T17:06:21Z' },
            { username: 'zhangsan', realName: '张三', email: 'zhangsan@example.com', role: 'admin', status: 'active', enabled: true, department: '研发部', phone: '13800138000', creationTimestamp: '2026-11-20T09:00:00Z' },
            { username: 'lisi', realName: '李四', email: '', role: 'normal', status: 'inactive', enabled: false, department: '运维部', phone: '', creationTimestamp: '2026-11-19T08:00:00Z' }
          ],
          total: 3,
          page: 1,
          size: 10
        }
      }
    })
  }
  if (url.includes('/user')) {
    return Promise.resolve({ data: { code: 200, msg: 'ok' } })
  }
  return Promise.reject(new Error('mock: ' + url))
}

export const clearAuthStorage = () => {
  sessionStorage.removeItem('token')
  sessionStorage.removeItem('userInfo')
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')
}

export const redirectToLogin = async (message?: string) => {
  if (redirectingToLogin) return
  redirectingToLogin = true
  clearAuthStorage()
  try {
    await router.push('/login')
  } finally {
    redirectingToLogin = false
  }
  ElMessage.error(message || '登录已过期，请重新登录')
}

// 请求拦截器
request.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    if (uiMockEnabled) {
      return uiMockHandler(config) as any
    }
    // 添加token到请求头
    const token = sessionStorage.getItem('token') || localStorage.getItem('token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    
    // 如果是标准的后端响应格式 {code, msg, data}
    if (data && typeof data === 'object' && 'code' in data) {
      const { code, msg } = data as any
      const numericCode = typeof code === 'number' ? code : Number(code)
      
      if (numericCode === 200) {
        // 成功响应，返回完整的响应数据
        return data
      } else if (numericCode === 401) {
        redirectToLogin(msg)
        return Promise.reject(new Error(msg || '未授权'))
      } else {
        // 其他错误
        ElMessage.error(msg || '请求失败')
        return Promise.reject(new Error(msg || '请求失败'))
      }
    }
    
    // 如果不是标准格式，直接返回数据
    return data
  },
  (error) => {
    console.error('响应错误:', error)
    
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 401:
          redirectToLogin(data?.msg)
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          break
        case 404:
          ElMessage.error('网络连接异常，请检查网络')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          if (Number(data?.code) === 401) {
            redirectToLogin(data?.msg)
            break
          }
          ElMessage.error(data?.msg || `请求失败 (${status})`)
      }
    } else if (error.request) {
      ElMessage.error('网络连接失败，请检查网络')
    } else {
      ElMessage.error('请求配置错误')
    }
    
    return Promise.reject(error)
  }
)

export default request

// 导出request实例
export { request }

// 导出常用的请求方法
export const get = <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  return request.get(url, config)
}

export const post = <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return request.post(url, data, config)
}

export const put = <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return request.put(url, data, config)
}

export const del = <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  return request.delete(url, config)
}

export const patch = <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
  return request.patch(url, data, config)
}

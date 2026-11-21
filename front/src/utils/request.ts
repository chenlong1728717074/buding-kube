import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 通用API响应格式
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页响应格式(与后端 vo.PageResponse 字段一一对应)
export interface PageResponse<T = any> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPage: number
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
  // 同步清空内存中的用户状态, 否则守卫读到过期的 isLoggedIn 会把 /login 弹回 /home
  try {
    const { useUserStore } = await import('@/stores/user')
    useUserStore().clearUserData()
  } catch {}
  try {
    await router.push('/login')
  } finally {
    redirectingToLogin = false
  }
  ElMessage.error(message || '登录已过期，请重新登录')
}

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
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

// 导出request实例(job/cronjob/service 等 api 文件使用具名导入)
export { request }


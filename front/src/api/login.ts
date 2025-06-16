import request, { type ApiResponse } from '@/utils/request'

// 登录请求DTO
export interface LoginDTO {
  username: string
  password: string
  captcha?: string
  captchaId?: string
}

// 登录响应VO
export interface LoginVO {
  token: string
  refreshToken?: string
  user: {
    id: string
    username: string
    email?: string
    role: number
    status: number
    department?: string
    phone?: string
  }
}

// 刷新Token DTO
export interface RefreshTokenDTO {
  refreshToken: string
}

// API响应格式
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 认证相关API
export const loginApi = {
  // 用户登录
  login: (data: LoginDTO): Promise<ApiResponse<LoginVO>> => {
    return request.post('/auth/login', data)
  },

  // 用户登出
  logout: (): Promise<ApiResponse<null>> => {
    return request.post('/auth/logout')
  },

  // 刷新token
  refreshToken: (data?: RefreshTokenDTO): Promise<ApiResponse<{ token: string; refreshToken?: string }>> => {
    return request.post('/auth/refresh', data)
  },

  // 获取验证码
  getCaptcha: (): Promise<ApiResponse<{ captchaId: string; captchaImage: string }>> => {
    return request.get('/auth/captcha')
  },

  // 验证token有效性
  verifyToken: (): Promise<ApiResponse<{ valid: boolean }>> => {
    return request.get('/auth/verify')
  },

  // 忘记密码 - 发送重置邮件
  forgotPassword: (email: string): Promise<ApiResponse<null>> => {
    return request.post('/auth/forgot-password', { email })
  },

  // 重置密码（通过邮件链接）
  resetPasswordByEmail: (data: { token: string; newPassword: string }): Promise<ApiResponse<null>> => {
    return request.post('/auth/reset-password', data)
  }
}

// 兼容旧版本的导出
export const authApi = loginApi
export default loginApi
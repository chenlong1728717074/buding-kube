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

// 认证相关API
export const loginApi = {
  // 用户登录
  login: (data: LoginDTO): Promise<ApiResponse<LoginVO>> => {
    return request.post('/auth/login', data)
  },

  // 用户登出
  logout: (): Promise<ApiResponse<null>> => {
    return request.post('/auth/logout')
  }
}

export default loginApi
import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 用户角色枚举
export enum UserRole {
  SUPER_ADMIN = 'super',
  ADMIN = 'admin',
  USER = 'normal'
}

// 用户状态枚举
export enum UserStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  SUSPENDED = 'suspended',
  EXPIRED = 'expired'
}

// 用户信息VO
export interface UserVO {
  username: string
  email?: string
  role: UserRole | string
  status: UserStatus | string
  enabled?: boolean
  department?: string
  phone?: string
  createdAt?: string
  updatedAt?: string
  lastLogin?: string
}

// 创建用户DTO
export interface CreateUserDTO {
  username: string
  realName?: string
  password: string
  email?: string
  role: UserRole | string
  department?: string
  phone?: string
  status?: UserStatus | string
}

// 更新用户DTO
export interface UpdateUserDTO {
  username: string
  realName?: string
  email?: string
  role?: UserRole | string
  department?: string
  phone?: string
  status?: UserStatus | string
}

// 用户查询DTO
export interface UserQueryDTO {
  username?: string
  email?: string
  role?: UserRole | string
  status?: UserStatus | string
  page?: number
  pageSize?: number
}

// 启用/禁用DTO
export interface UpdateUserEnableDTO {
  username: string
  enable: boolean
}

// 重置密码DTO
export interface ResetPasswordDTO {
  username: string
}

// 修改密码DTO
export interface ChangePasswordDTO {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

// API响应格式
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 分页响应格式
export interface PageResponse<T = any> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// 用户管理相关API
export const userApi = {
  // 获取用户列表
  getUsers: (params: UserQueryDTO): Promise<ApiResponse<PageResponse<UserVO>>> => {
    return request.get('/user/list', { params })
  },

  // 根据用户名获取用户
  getUserById: (username: string): Promise<ApiResponse<UserVO>> => {
    return request.get(`/user/${username}`)
  },

  // 获取当前用户信息
  getCurrentUser: (): Promise<ApiResponse<UserVO>> => {
    return request.get('/user/profile')
  },

  // 创建用户
  createUser: (data: CreateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.post('/user', data)
  },

  // 更新用户
  updateUser: (data: UpdateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.put('/user', data)
  },

  // 删除用户
  deleteUser: (username: string): Promise<ApiResponse<null>> => {
    return request.delete(`/user/${username}`)
  },

  // 批量删除用户
  batchDeleteUsers: (usernames: string[]): Promise<ApiResponse<null>> => {
    return request.post('/user/batchDelete', { usernames })
  },

  // 启用/禁用用户
  toggleUserEnable: (data: UpdateUserEnableDTO): Promise<ApiResponse<null>> => {
    return request.put('/user/enable', data)
  },

  // 重置用户密码
  resetPassword: (data: ResetPasswordDTO): Promise<ApiResponse<null>> => {
    return request.post(`/user/resetPassword/${data.username}`)
  },

  // 更新用户资料
  updateProfile: (data: UpdateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.put('/user/profile', data)
  },

  // 修改密码
  changePassword: (data: ChangePasswordDTO): Promise<ApiResponse<null>> => {
    return request.put('/auth/changePassword', data)
  },

  // 发送邮箱验证
  sendEmailVerification: (): Promise<ApiResponse<null>> => {
    return request.post('/user/verify-email')
  },

  // 检查用户名是否可用
  checkUsername: (username: string): Promise<ApiResponse<{ available: boolean }>> => {
    return request.get(`/user/check-username`, { params: { username } })
  },

  // 检查邮箱是否可用
  checkEmail: (email: string): Promise<ApiResponse<{ available: boolean }>> => {
    return request.get(`/user/check-email`, { params: { email } })
  }
}

export default userApi
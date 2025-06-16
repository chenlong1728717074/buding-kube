import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 用户角色枚举
export enum UserRole {
  SUPER_ADMIN = 1,
  ADMIN = 2,
  USER = 3
}

// 用户状态枚举
export enum UserStatus {
  DISABLED = 0,
  ENABLED = 1
}

// 用户信息VO
export interface UserVO {
  id: string
  username: string
  email?: string
  role: number
  status: number
  department?: string
  phone?: string
  createdAt?: string
  updatedAt?: string
}

// 创建用户DTO
export interface CreateUserDTO {
  username: string
  password: string
  email?: string
  role: number
  department?: string
  phone?: string
  status?: number
}

// 更新用户DTO
export interface UpdateUserDTO {
  id: string
  username?: string
  email?: string
  role?: number
  department?: string
  phone?: string
  status?: number
}

// 用户查询DTO
export interface UserQueryDTO {
  username?: string
  email?: string
  role?: number
  status?: number
  page?: number
  pageSize?: number
}

// 重置密码DTO
export interface ResetPasswordDTO {
  userId: string
  newPassword: string
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

  // 根据ID获取用户
  getUserById: (id: string): Promise<ApiResponse<UserVO>> => {
    return request.get(`/user/${id}`)
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
  deleteUser: (id: string): Promise<ApiResponse<null>> => {
    return request.delete(`/user/${id}`)
  },

  // 批量删除用户
  batchDeleteUsers: (ids: string[]): Promise<ApiResponse<null>> => {
    return request.post('/user/batch-delete', { ids })
  },

  // 切换用户状态
  toggleUserStatus: (id: string, status: number): Promise<ApiResponse<null>> => {
    return request.put(`/user/${id}/status`, { status })
  },

  // 重置用户密码
  resetPassword: (data: ResetPasswordDTO): Promise<ApiResponse<null>> => {
    const { userId, newPassword } = data
    return request.post(`/user/${userId}/reset-password`, { password: newPassword })
  },

  // 更新用户资料
  updateProfile: (data: UpdateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.put('/user/profile', data)
  },

  // 修改密码
  changePassword: (data: ChangePasswordDTO): Promise<ApiResponse<null>> => {
    return request.put('/user/password', data)
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
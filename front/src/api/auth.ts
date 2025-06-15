import request from '@/utils/request'

// 用户角色枚举
export enum UserRole {
  SUPER_ADMIN = 'super_admin',
  ADMIN = 'admin',
  USER = 'user'
}

// 用户状态枚举
export enum UserStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  BANNED = 'banned'
}

// 登录请求DTO
export interface LoginDTO {
  username: string
  password: string
}

// 用户信息VO
export interface UserVO {
  id: number
  username: string
  email?: string
  role: UserRole
  status: UserStatus
  token?: string
  createdAt: string
  updatedAt: string
}

// 创建用户DTO
export interface CreateUserDTO {
  username: string
  password: string
  email?: string
  role: UserRole
}

// 更新用户DTO
export interface UpdateUserDTO {
  id: number
  username?: string
  email?: string
  role?: UserRole
  status?: UserStatus
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

// 用户列表查询参数
export interface UserListParams {
  page?: number
  pageSize?: number
  username?: string
  role?: UserRole
  status?: UserStatus
}

// 认证相关API
export const authApi = {
  // 用户登录
  login: (data: LoginDTO): Promise<ApiResponse<UserVO>> => {
    return request.post('/auth/login', data)
  },

  // 获取当前用户信息
  getCurrentUser: (): Promise<ApiResponse<UserVO>> => {
    return request.get('/auth/me')
  },

  // 刷新token
  refreshToken: (): Promise<ApiResponse<{ token: string }>> => {
    return request.post('/auth/refresh')
  },

  // 用户登出
  logout: (): Promise<ApiResponse<null>> => {
    return request.post('/auth/logout')
  }
}

// 用户管理相关API
export const userApi = {
  // 获取用户列表
  getUsers: (params: UserListParams): Promise<ApiResponse<PageResponse<UserVO>>> => {
    return request.get('/users', { params })
  },

  // 根据ID获取用户
  getUserById: (id: number): Promise<ApiResponse<UserVO>> => {
    return request.get(`/users/${id}`)
  },

  // 创建用户
  createUser: (data: CreateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.post('/users', data)
  },

  // 更新用户
  updateUser: (data: UpdateUserDTO): Promise<ApiResponse<UserVO>> => {
    return request.put(`/users/${data.id}`, data)
  },

  // 删除用户
  deleteUser: (id: number): Promise<ApiResponse<null>> => {
    return request.delete(`/users/${id}`)
  },

  // 批量删除用户
  batchDeleteUsers: (ids: number[]): Promise<ApiResponse<null>> => {
    return request.post('/users/batch-delete', { ids })
  },

  // 重置用户密码
  resetPassword: (id: number, newPassword: string): Promise<ApiResponse<null>> => {
    return request.post(`/users/${id}/reset-password`, { password: newPassword })
  }
}

// 兼容旧版本的导出
export const loginApi = authApi.login
export default authApi
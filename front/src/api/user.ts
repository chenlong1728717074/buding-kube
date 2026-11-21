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
  realName?: string
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
  password?: string
}

// 修改密码DTO
export interface ChangePasswordDTO {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

// 用户管理相关API
export const userApi = {
  // 获取用户列表
  getUsers: (params: UserQueryDTO): Promise<ApiResponse<PageResponse<UserVO>>> => {
    return request.get('/user/list', { params })
  },

  // 获取当前登录用户的个人信息
  getProfile: (): Promise<ApiResponse<UserVO>> => {
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

  // 重置用户密码(password 留空则重置为 123456)
  resetPassword: (data: ResetPasswordDTO): Promise<ApiResponse<null>> => {
    return request.post(`/user/resetPassword/${data.username}`, { password: data.password || '' })
  },

  // 更新个人资料(仅本人可改基本资料字段)
  updateProfile: (data: { realName?: string; email?: string; phone?: string }): Promise<ApiResponse<UserVO>> => {
    return request.put('/user/profile', data)
  },

  // 修改密码
  changePassword: (data: ChangePasswordDTO): Promise<ApiResponse<null>> => {
    return request.put('/auth/changePassword', data)
  }
}

export default userApi
import request from '@/utils/request'

// 用户信息
export interface UserVO {
  id?: string
  username: string
  email: string
  realName?: string
  phone?: string
  avatar?: string
  department?: string
  position?: string
  role: string
  status: string
  bio?: string
  lastLoginTime?: string
  createTime: string
  updateTime: string
}

// 用户查询参数
export interface UserQueryDTO {
  page?: number
  pageSize?: number
  username?: string
  email?: string
  role?: string
  status?: string
  department?: string
}

// 创建用户参数
export interface CreateUserDTO {
  username: string
  email: string
  password: string
  realName?: string
  phone?: string
  department?: string
  position?: string
  role: string
  status?: string
  bio?: string
}

// 更新用户参数
export interface UpdateUserDTO {
  id?: string
  username?: string
  email?: string
  realName?: string
  phone?: string
  department?: string
  position?: string
  role?: string
  status?: string
  bio?: string
}

// 修改密码参数
export interface ChangePasswordDTO {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}

// 重置密码参数
export interface ResetPasswordDTO {
  userId: string
  newPassword: string
}

// 登录历史
export interface LoginHistoryVO {
  id: string
  userId: string
  username: string
  ip: string
  userAgent: string
  location?: string
  loginTime: string
  status: string
  message?: string
}

// 用户权限
export interface UserPermissionVO {
  id: string
  name: string
  code: string
  type: string
  description?: string
  children?: UserPermissionVO[]
}

// 用户管理API
export const userApi = {
  // 获取用户列表
  getUsers: (params: UserQueryDTO) => {
    return request.get<PageResponse<UserVO>>('/user/list', { params })
  },

  // 获取单个用户
  getUser: (id: string) => {
    return request.get<UserVO>(`/user/${id}`)
  },

  // 获取当前用户信息
  getCurrentUser: () => {
    return request.get<UserVO>('/user/profile')
  },

  // 创建用户
  createUser: (data: CreateUserDTO) => {
    return request.post<UserVO>('/user', data)
  },

  // 更新用户
  updateUser: (id: string, data: UpdateUserDTO) => {
    return request.put<UserVO>(`/user/${id}`, data)
  },

  // 更新当前用户信息
  updateProfile: (data: UpdateUserDTO) => {
    return request.put<UserVO>('/user/profile', data)
  },

  // 删除用户
  deleteUser: (id: string) => {
    return request.delete(`/user/${id}`)
  },

  // 批量删除用户
  batchDeleteUsers: (ids: string[]) => {
    return request.delete('/user/batch', { data: { ids } })
  },

  // 启用/禁用用户
  toggleUserStatus: (id: string, status: string) => {
    return request.put(`/user/${id}/status`, { status })
  },

  // 修改密码
  changePassword: (data: ChangePasswordDTO) => {
    return request.put('/user/password', data)
  },

  // 重置密码
  resetPassword: (data: ResetPasswordDTO) => {
    return request.put('/user/reset-password', data)
  },

  // 上传头像
  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return request.post('/user/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 获取用户登录历史
  getLoginHistory: (params?: any) => {
    return request.get<PageResponse<LoginHistoryVO>>('/user/login-history', { params })
  },

  // 获取用户登录历史（指定用户）
  getUserLoginHistory: (userId: string, params?: any) => {
    return request.get<PageResponse<LoginHistoryVO>>(`/user/${userId}/login-history`, { params })
  },

  // 获取用户权限
  getUserPermissions: (userId?: string) => {
    const url = userId ? `/user/${userId}/permissions` : '/user/permissions'
    return request.get<UserPermissionVO[]>(url)
  },

  // 更新用户权限
  updateUserPermissions: (userId: string, permissionIds: string[]) => {
    return request.put(`/user/${userId}/permissions`, { permissionIds })
  },

  // 获取所有权限列表
  getAllPermissions: () => {
    return request.get<UserPermissionVO[]>('/user/permissions/all')
  },

  // 获取角色列表
  getRoles: () => {
    return request.get('/user/roles')
  },

  // 获取部门列表
  getDepartments: () => {
    return request.get('/user/departments')
  },

  // 验证邮箱
  verifyEmail: (email: string) => {
    return request.post('/user/verify-email', { email })
  },

  // 确认邮箱验证
  confirmEmailVerification: (token: string) => {
    return request.post('/user/confirm-email', { token })
  },

  // 绑定手机号
  bindPhone: (phone: string, code: string) => {
    return request.post('/user/bind-phone', { phone, code })
  },

  // 发送手机验证码
  sendPhoneCode: (phone: string) => {
    return request.post('/user/send-phone-code', { phone })
  },

  // 解绑手机号
  unbindPhone: (code: string) => {
    return request.post('/user/unbind-phone', { code })
  },

  // 启用两步验证
  enableTwoFactor: () => {
    return request.post('/user/two-factor/enable')
  },

  // 禁用两步验证
  disableTwoFactor: (code: string) => {
    return request.post('/user/two-factor/disable', { code })
  },

  // 获取两步验证二维码
  getTwoFactorQR: () => {
    return request.get('/user/two-factor/qr')
  },

  // 验证两步验证码
  verifyTwoFactor: (code: string) => {
    return request.post('/user/two-factor/verify', { code })
  },

  // 获取用户统计信息
  getUserStats: () => {
    return request.get('/user/stats')
  },

  // 导出用户列表
  exportUsers: (params?: UserQueryDTO) => {
    return request.get('/user/export', { 
      params,
      responseType: 'blob'
    })
  },

  // 导入用户
  importUsers: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/user/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 下载用户导入模板
  downloadImportTemplate: () => {
    return request.get('/user/import-template', {
      responseType: 'blob'
    })
  },

  // 检查用户名是否可用
  checkUsername: (username: string) => {
    return request.get(`/user/check-username/${username}`)
  },

  // 检查邮箱是否可用
  checkEmail: (email: string) => {
    return request.get(`/user/check-email/${email}`)
  },

  // 获取用户活动日志
  getUserActivityLogs: (userId?: string, params?: any) => {
    const url = userId ? `/user/${userId}/activity-logs` : '/user/activity-logs'
    return request.get(url, { params })
  },

  // 锁定用户
  lockUser: (id: string, reason?: string) => {
    return request.put(`/user/${id}/lock`, { reason })
  },

  // 解锁用户
  unlockUser: (id: string) => {
    return request.put(`/user/${id}/unlock`)
  },

  // 强制用户下线
  forceLogout: (id: string) => {
    return request.post(`/user/${id}/force-logout`)
  },

  // 获取在线用户列表
  getOnlineUsers: (params?: any) => {
    return request.get('/user/online', { params })
  },

  // 发送系统通知
  sendNotification: (userIds: string[], title: string, content: string) => {
    return request.post('/user/notification', {
      userIds,
      title,
      content
    })
  },

  // 获取用户通知
  getUserNotifications: (params?: any) => {
    return request.get('/user/notifications', { params })
  },

  // 标记通知为已读
  markNotificationRead: (notificationId: string) => {
    return request.put(`/user/notifications/${notificationId}/read`)
  },

  // 删除通知
  deleteNotification: (notificationId: string) => {
    return request.delete(`/user/notifications/${notificationId}`)
  },

  // 清空所有通知
  clearAllNotifications: () => {
    return request.delete('/user/notifications/all')
  }
}

export default userApi
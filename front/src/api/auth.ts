// 重新导出登录和用户相关的API
export { loginApi, authApi } from './login'
export { userApi } from './user'

// 重新导出类型定义
export type { LoginDTO, LoginVO, ApiResponse } from './login'
export type { 
  UserVO, 
  CreateUserDTO, 
  UpdateUserDTO, 
  UserQueryDTO, 
  ResetPasswordDTO, 
  ChangePasswordDTO,
  UserRole,
  UserStatus,
  PageResponse
} from './user'

// 兼容旧版本的导出
export { loginApi as default } from './login'
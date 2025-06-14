package dto

import "buding-kube/internal/model"

// LoginDTO 登录参数
type LoginDTO struct {
	Username string `json:"username" example:"zhangsan"` // 用户名，用户登录账号
	Password string `json:"password" example:"123456"`   // 密码，用户登录密码
}

// CreateUserDTO 创建用户参数
type CreateUserDTO struct {
	Username string         `json:"username" binding:"required" example:"zhangsan"` // 用户名，用户登录账号
	Password string         `json:"password" binding:"required" example:"123456"`   // 密码，用户登录密码
	Email    string         `json:"email" example:"zhangsan@example.com"`           // 邮箱
	Role     model.UserRole `json:"role" binding:"required" example:"2"`            // 角色: 1=超级管理员 2=管理员 3=普通用户
}

// UpdateUserDTO 更新用户参数
type UpdateUserDTO struct {
	Email    string         `json:"email" example:"zhangsan@example.com"` // 邮箱
	Password string         `json:"password" example:"123456"`            // 密码，用户登录密码
	Role     model.UserRole `json:"role" example:"2"`                     // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status   int            `json:"status" example:"1"`                   // 状态: 1=正常 0=禁用
}

// UserQueryDTO 用户查询参数
type UserQueryDTO struct {
	Username string `form:"username"` // 用户名
	Role     int    `form:"role"`     // 角色
	Status   int    `form:"status"`   // 状态
	PageQueryDTO
}

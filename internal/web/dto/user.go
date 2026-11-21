package dto

import (
	"buding-kube/internal/kube"
)

// CreateUserDTO 创建用户参数
type CreateUserDTO struct {
	Username   string    `json:"username" binding:"required" example:"zhangsan"` // 用户名，用户登录账号
	RealName   string    `json:"realName" binding:"required" example:"张三"`       // 真实姓名
	Password   string    `json:"password"  example:"123456"`                     // 密码，用户登录密码
	Email      string    `json:"email" example:"zhangsan@example.com"`           // 邮箱
	Phone      string    `json:"phone" example:"13800138000"`                    // 手机号
	Role       kube.Role `json:"role" binding:"required" example:"admin"`        // 角色: super/admin/normal
	Department string    `json:"department" example:"研发部"`                       // 部门
	Status     string    `json:"status" example:"active"`                        // 状态: active/inactive/suspended/expired
}

// UpdateUserDTO 更新用户参数(字段均可选, username 用于定位用户)
type UpdateUserDTO struct {
	Username   string    `json:"username" binding:"required" example:"zhangsan"` // 用户名，用户登录账号
	RealName   string    `json:"realName" example:"张三"`                         // 真实姓名
	Email      string    `json:"email" example:"zhangsan@example.com"`           // 邮箱
	Phone      string    `json:"phone" example:"13800138000"`                    // 手机号
	Role       kube.Role `json:"role" example:"admin"`                           // 角色: super/admin/normal
	Department string    `json:"department" example:"研发部"`                       // 部门
	Status     string    `json:"status" example:"active"`                        // 状态: active/inactive/suspended/expired
}

// ProfileUpdateDTO 个人资料更新参数(仅本人可改基本资料, 不含角色/状态等敏感字段)
type ProfileUpdateDTO struct {
	RealName string `json:"realName" example:"张三"`                         // 真实姓名
	Email    string `json:"email" example:"zhangsan@example.com"`           // 邮箱
	Phone    string `json:"phone" example:"13800138000"`                    // 手机号
}

// UserQueryDTO 用户查询参数
type UserQueryDTO struct {
	Username     string    `form:"username" example:"zhangsan"` // 用户名，精确匹配
	Role         kube.Role `form:"role" example:"admin"`        // 角色: super/admin/normal
	Status       string    `form:"status" example:"active"`     // 状态: active/inactive/suspended/expired
	PageQueryDTO           // 嵌入分页查询基础参数
}

// UpdateUserStatusDTO 更新用户状态参数
type UpdateUserStatusDTO struct {
	Username string `json:"username" binding:"required" example:"zhangsan"`
	Status   string `json:"status" binding:"required" example:"active"`
}

// UpdateUserEnableDTO 启用/禁用用户参数
type UpdateUserEnableDTO struct {
	Username string `json:"username" binding:"required" example:"zhangsan"`
	Enable   bool   `json:"enable"`
}

// BatchDeleteUsersDTO 批量删除用户参数
type BatchDeleteUsersDTO struct {
	Usernames []string `json:"usernames" binding:"required,min=1"`
}

// ResetPasswordDTO 重置用户密码参数(密码留空则使用默认密码 123456)
type ResetPasswordDTO struct {
	Password string `json:"password" example:"newPass123"`
}

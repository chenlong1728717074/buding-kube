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
	Role       kube.Role `json:"role" binding:"required" example:"admin"`        // 角色: super/admin/normal
	Department string    `json:"department" example:"研发部"`                       // 部门
	Status     string    `json:"status" example:"active"`                        // 状态: active/inactive/suspended/expired
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

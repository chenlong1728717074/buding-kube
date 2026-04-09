package dto

import (
	"buding-kube/internal/kube"
)

// CreateUserDTO 创建用户参数
type CreateUserDTO struct {
	Username   string    `json:"username" binding:"required" example:"zhangsan"` // 用户名，用户登录账号
	Password   string    `json:"password" binding:"required" example:"123456"`   // 密码，用户登录密码
	Email      string    `json:"email" example:"zhangsan@example.com"`           // 邮箱
	Role       kube.Role `json:"role" binding:"required" example:"admin"`        // 角色: 1=超级管理员 2=管理员 3=普通用户
	Department string    `json:"department" example:"研发部"`                       //部门
	Status     int       `json:"status" example:"0"`                             //状态
}

// UserQueryDTO 用户查询参数
type UserQueryDTO struct {
	Username     string    `form:"username" example:"zhangsan"` // 用户名，精确匹配
	Role         kube.Role `form:"role" example:"admin"`        // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status       string    `form:"status" example:"active"`     // 状态: 1=正常 0=禁用
	PageQueryDTO           // 嵌入分页查询基础参数
}

package dto

import "buding-kube/internal/model"

// LoginDTO 登录参数
type LoginDTO struct {
	Username string `json:"username" example:"zhangsan" binding:"required"` // 用户名，用户登录账号
	Password string `json:"password" example:"123456" binding:"required"`   // 密码，用户登录密码
}

// CreateUserDTO 创建用户参数
type CreateUserDTO struct {
	Username   string         `json:"username" binding:"required" example:"zhangsan"`         // 用户名，用户登录账号
	Password   string         `json:"password" binding:"required" example:"123456"`           // 密码，用户登录密码
	Email      string         `json:"email" example:"zhangsan@example.com"`                   // 邮箱
	Role       model.UserRole `json:"role" binding:"required" example:"2"`                    // 角色: 1=超级管理员 2=管理员 3=普通用户
	Department string         `json:"department" example:"研发部"`                               //部门
	Phone      string         `json:"phone" example:"131420203214"`                           //电话号码
	Cluster    string         `json:"cluster" example:"0599432d-98fa-4da4-8c84-96a01ffaf113"` //授权的集群
	Status     int            `json:"status" example:"0"`                                     //状态
}

// UpdateUserDTO 更新用户参数
type UpdateUserDTO struct {
	Id         string         `json:"id" binding:"required" example:"0599432d-98fa-4da4-8c84-96a01ffaf113"` // id
	Username   string         `json:"username" binding:"required" example:"zhangsan"`                       // 用户名，用户登录账号
	Password   string         `json:"password" example:"123456"`                                            // 密码，用户登录密码
	Email      string         `json:"email" example:"zhangsan@example.com"`                                 // 邮箱
	Role       model.UserRole `json:"role" binding:"required" example:"2"`                                  // 角色: 1=超级管理员 2=管理员 3=普通用户
	Department string         `json:"department" example:"研发部"`                                             //部门
	Phone      string         `json:"phone" example:"131420203214"`                                         //电话号码
	Cluster    string         `json:"cluster" example:"0599432d-98fa-4da4-8c84-96a01ffaf113"`               //授权的集群
	Status     int            `json:"status" example:"0"`
}

// UserQueryDTO 用户查询参数
type UserQueryDTO struct {
	Username     string `form:"username" example:"zhangsan"` // 用户名，精确匹配
	Role         int    `form:"role" example:"2"`            // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status       int    `form:"status" example:"1"`          // 状态: 1=正常 0=禁用
	PageQueryDTO        // 嵌入分页查询基础参数
}

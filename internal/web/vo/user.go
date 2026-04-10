package vo

import (
	"buding-kube/internal/kube"
	"time"
)

// UserVO 用户视图对象
type UserVO struct {
	Username   string            `json:"username" example:"zhangsan"`          // 用户名
	Email      string            `json:"email" example:"zhangsan@example.com"` // 邮箱
	Role       kube.Role         `json:"role" example:"1"`                     // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status     string            `json:"status" example:"active"`              // 状态: active/inactive/suspended/expired
	Enabled    bool              `json:"enabled"`                              // 是否启用
	Department string            `json:"department" example:"研发部"`             //部门
	Attr       map[string]string `json:"attr" example:"附加信息"`                  //附加信息
	LastLogin  time.Time         `json:"lastLogin" example:"2025-10-21"`       //附加信息
}

// User2VO 将用户模型转换为视图对象
func User2VO(user *kube.User) *UserVO {
	result := UserVO{
		Username:   user.Name,
		Email:      user.Spec.Email,
		Role:       user.Spec.Role,
		Department: user.Spec.Department,
		Attr:       user.Spec.Attributes,
		Status:     user.Status.State,
		Enabled:    user.Spec.Enabled,
	}
	if user.Status.LastLogin != nil {
		result.LastLogin = user.Status.LastLogin.Time
	}
	return &result
}

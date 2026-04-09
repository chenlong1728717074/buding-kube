package vo

import (
	"buding-kube/internal/kube"
)

type LoginUserVO struct {
	Username   string            `json:"username" example:"zhangsan"`                             // 用户名
	Email      string            `json:"email" example:"zhangsan@example.com"`                    // 邮箱
	Role       int               `json:"role" example:"1"`                                        // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status     string            `json:"status" example:"1"`                                      // 状态: 1=正常 0=禁用
	Token      string            `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT令牌
	Department string            `json:"department" example:"研发部"`                                //部门
	Attr       map[string]string `json:"attr" example:"附加信息"`                                     //附加信息
}

func ToLoginUserVO(user *kube.User, token string) *LoginUserVO {
	return &LoginUserVO{
		Username:   user.Name,
		Email:      user.Spec.Email,
		Role:       kube.RoleLevel(user.Spec.Role),
		Status:     user.Status.State,
		Token:      token,
		Department: user.Spec.Department,
		Attr:       user.Spec.Attributes,
	}
}

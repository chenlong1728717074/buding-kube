package vo

import "buding-kube/internal/model"

// UserVO 用户视图对象
type UserVO struct {
	Username  string         `json:"username" example:"zhangsan"`                             // 用户名
	Email     string         `json:"email" example:"zhangsan@example.com"`                    // 邮箱
	Role      model.UserRole `json:"role" example:"1"`                                        // 角色: 1=超级管理员 2=管理员 3=普通用户
	Namespace string         `json:"namespace" example:"default"`                             // 命名空间
	Status    int            `json:"status" example:"1"`                                      // 状态: 1=正常 0=禁用
	Token     string         `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT令牌
	Cluster   string         `json:"cluster" example:"cluster-1"`                             // 集群名称
}

// User2VO 将用户模型转换为视图对象
func User2VO(user model.User, token string) *UserVO {
	return &UserVO{
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		Namespace: user.Namespace,
		Cluster:   user.Cluster,
		Token:     token,
	}
}

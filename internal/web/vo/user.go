package vo

import "buding-kube/internal/model"

type UserVO struct {
	Username  string         `json:"username" example:"zhangsan"`
	Email     string         `json:"email" example:"zhangsan@example.com"`
	Role      model.UserRole `json:"role" example:"1"`
	Namespace string         `json:"namespace" example:"default"`
	Status    int            `json:"status" example:"1"`
	Token     string         `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Cluster   string         `json:"cluster" example:"cluster-1"`
}

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

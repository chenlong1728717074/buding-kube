package vo

import "buding-kube/internal/model"

type UserVO struct {
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Role      model.UserRole `json:"role"`
	Namespace string         `json:"namespace"`
	Status    int            `json:"status"`
	Token     string         `json:"token"`
	Cluster   string         `json:"cluster"`
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

package model

import (
	"buding-kube/pkg/utils/password"
)

type UserRole int

const (
	SuperAdmin                 UserRole = 1 // 超级管理员
	Admin                      UserRole = 2 // 管理员
	Normal                     UserRole = 3 // 普通用户
	UserConfigSecretLabelKey   string   = "buding-kube.com/username"
	UserConfigSecretLabelValue string   = "admin"
	UserPass                   string   = "123456"
)

type User struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Salt      string   `json:"salt"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`
	Namespace string   `json:"namespace"`
	Cluster   string   `json:"cluster"`
	Status    int      `json:"status"`
}

func AdminUser() *User {
	password, salt := password.GeneratePassword(UserPass)
	return &User{
		Username: UserConfigSecretLabelValue,
		Role:     SuperAdmin,
		Status:   1,
		Password: password,
		Salt:     salt,
	}
}

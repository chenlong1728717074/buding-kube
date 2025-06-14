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

	UserStatusActive   int = 1 // 用户状态: 正常
	UserStatusDisabled int = 0 // 用户状态: 禁用
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
		Status:   UserStatusActive,
		Password: password,
		Salt:     salt,
	}
}

// NewUser 创建一个新用户
func NewUser(username, plainPassword, email string, role UserRole) *User {
	encodedPassword, salt := password.GeneratePassword(plainPassword)
	return &User{
		Username: username,
		Password: encodedPassword,
		Salt:     salt,
		Email:    email,
		Role:     role,
		Status:   UserStatusActive,
	}
}

// IsSuperAdmin 判断是否为超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Role == SuperAdmin
}

// IsAdmin 判断是否为管理员(包括超级管理员)
func (u *User) IsAdmin() bool {
	return u.Role == Admin || u.Role == SuperAdmin
}

// CanManageUser 判断当前用户是否可以管理目标用户
func (u *User) CanManageUser(target *User) bool {
	// 超级管理员可以管理所有用户
	if u.IsSuperAdmin() {
		return true
	}

	// 管理员可以管理普通用户和自己
	if u.IsAdmin() {
		return target.Role == Normal || u.Username == target.Username
	}

	// 普通用户只能管理自己
	return u.Username == target.Username
}

// CanDeleteUser 判断当前用户是否可以删除目标用户
func (u *User) CanDeleteUser(target *User) bool {
	// 任何人都不能删除超级管理员
	if target.IsSuperAdmin() {
		return false
	}

	// 用户不能删除自己
	if u.Username == target.Username {
		return false
	}

	// 超级管理员可以删除任何人(除了自己和其他超级管理员)
	if u.IsSuperAdmin() {
		return true
	}

	// 管理员可以删除普通用户
	if u.IsAdmin() {
		return target.Role == Normal
	}

	// 普通用户不能删除任何人
	return false
}

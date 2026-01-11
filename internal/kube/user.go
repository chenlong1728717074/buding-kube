package kube

import (
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"context"
	"fmt"
	"github.com/alexedwards/argon2id"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	// RoleSuperAdmin 角色定义
	RoleSuperAdmin = "super"
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleReadOnly   = "readonly"

	// UserStateActive 用户状态
	UserStateActive    = "active"
	UserStateLocked    = "locked"
	UserStateSuspended = "suspended"
	UserStateDeleted   = "deleted"

	// PermissionClusterCreate 权限定义
	PermissionClusterCreate = "cluster:create"
	PermissionClusterRead   = "cluster:read"
	PermissionClusterUpdate = "cluster:update"
	PermissionClusterDelete = "cluster:delete"
	PermissionUserCreate    = "user:create"
	PermissionUserRead      = "user:read"
	PermissionUserUpdate    = "user:update"
	PermissionUserDelete    = "user:delete"

	// LabelRole 标签定义
	LabelRole       = "role"
	LabelDepartment = "department"
	LabelUserConfig = "user-config"

	SuperAdmin     string = "admin"
	SuperAdminPass string = "123456"
)

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserSpec   `json:"spec,omitempty"`
	Status            UserStatus `json:"status,omitempty"`
}

// UserSpec 用户规格定义
type UserSpec struct {
	Email      string            `json:"email"`
	RealName   string            `json:"realName,omitempty"`
	Password   string            `json:"password"`
	Role       string            `json:"role"` // super/admin/normal
	Enabled    bool              `json:"enabled"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type UserQuota struct {
	// 最大集群数
	MaxClusters int `json:"maxClusters,omitempty"`

	// 最大 CPU 核心数
	MaxCPU string `json:"maxCPU,omitempty"`

	// 最大内存
	MaxMemory string `json:"maxMemory,omitempty"`

	// 最大存储
	MaxStorage string `json:"maxStorage,omitempty"`
}

// UserStatus 用户状态
type UserStatus struct {
	State     string       `json:"state,omitempty"`     // active/inactive/suspended/expired
	LastLogin *metav1.Time `json:"lastLogin,omitempty"` // 最后登录时间
}

// UsedResources 已使用的资源
type UsedResources struct {
	// 已使用的集群数
	Clusters int `json:"clusters,omitempty"`

	// 已使用的 CPU
	CPU string `json:"cpu,omitempty"`

	// 已使用的内存
	Memory string `json:"memory,omitempty"`

	// 已使用的存储
	Storage string `json:"storage,omitempty"`
}

// UserList 用户列表
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// IsEnabled 检查用户是否启用
func (u *User) IsEnabled() bool {
	return u.Spec.Enabled
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status.State == "active"
}

// IsSuperAdmin 检查是否是超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.Spec.Role == "super"
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Spec.Role == "admin" || u.Spec.Role == "super"
}

func (u *User) CheckPassword(password string) bool {
	match, _ := argon2id.ComparePasswordAndHash(password, u.Spec.Password)
	return match
}

func CreateUser(user *User, client dynamic.Interface) error {
	ctx := context.Background()

	unstructuredObj, err := utils.ToUnstructured(user)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	// 使用动态客户端创建 (User 是 NamespaceScoped)
	_, err = client.Resource(UserGVR).Create(
		ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	logs.Info("✅ 创建用户: %s", user.Name)
	return nil
}

func buildSuperAdmin() *User {
	generatePassword, _ := argon2id.CreateHash(SuperAdminPass, argon2id.DefaultParams)

	adminUser := &User{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", GroupCrd, ClusterCrdVersion),
			Kind:       UserKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SuperAdmin,
		},
		Spec: UserSpec{
			Role:     RoleSuperAdmin,
			Password: generatePassword,
			Enabled:  true,
		},
		Status: UserStatus{
			State: UserStateActive,
		},
	}
	return adminUser
}

package kube

import (
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"context"
	"fmt"

	"github.com/alexedwards/argon2id"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	// RoleSuperAdmin 角色定义
	RoleSuperAdmin = "super"
	RoleAdmin      = "admin"
	RoleNormal     = "normal"

	// UserStateActive 用户状态
	UserStateActive    = "active"
	UserStateInactive  = "inactive"
	UserStateSuspended = "suspended"
	UserStateExpired   = "expired"

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

type Role string

var roleLevel = map[Role]int{
	RoleNormal:     3,
	RoleAdmin:      2,
	RoleSuperAdmin: 1,
}

func RoleLevel(role Role) int {
	return roleLevel[role]
}

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserSpec   `json:"spec,omitempty"`
	Status            UserStatus `json:"status,omitempty"`
}

type LoginUser struct {
	Username string
	Role     Role
}

func (u *LoginUser) CanManageUser(user *User) bool {
	if u == nil || user == nil {
		return false
	}

	if u.Username == user.Name {
		return true
	}

	myLevel, ok1 := roleLevel[u.Role]
	targetLevel, ok2 := roleLevel[user.Spec.Role]

	if !ok1 || !ok2 {
		return false
	}

	// 权限高（数值小）才能管理权限低（数值大）
	return myLevel < targetLevel
}

func (u *LoginUser) CanDeleteUser(target *User) bool {
	if u == nil || target == nil {
		return false
	}

	// ❌ 禁止删除自己
	if u.Username == target.Name {
		return false
	}

	// 其他规则复用管理逻辑
	return u.CanManageUser(target)
}

// UserSpec 用户规格定义
type UserSpec struct {
	Email      string            `json:"email"`
	RealName   string            `json:"realName,omitempty"`
	Password   string            `json:"password"`
	Role       Role              `json:"role"` // super/admin/normal
	Enabled    bool              `json:"enabled"`
	Department string            `json:"department"`
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

func (u *User) getRoleLevel() int {
	return roleLevel[u.Spec.Role]
}

func (u *User) HasPermission(requiredRole Role) bool {
	return roleLevel[u.Spec.Role] <= roleLevel[requiredRole]
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
			APIVersion: fmt.Sprintf("%s/%s", consts.GroupCrd, consts.ClusterCrdVersion),
			Kind:       consts.UserKind,
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

// GetUser 获取单个用户
func GetUser(name string) (*User, error) {
	ctx := context.Background()

	unstructuredObj, err := GlobalClient.DynamicClient.Resource(UserGVR).Get(
		ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("用户不存在: %s", name)
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	user := &User{}
	if err := utils.FromUnstructured(unstructuredObj, user); err != nil {
		return nil, fmt.Errorf("转换用户数据失败: %w", err)
	}

	return user, nil
}

// ListUsers  带过滤条件的用户列表查询
func ListUsers(opts metav1.ListOptions) ([]User, error) {
	ctx := context.Background()

	list, err := GlobalClient.DynamicClient.Resource(UserGVR).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("列出用户失败: %w", err)
	}

	users := make([]User, 0, len(list.Items))
	for _, item := range list.Items {
		user := User{}
		if err := utils.FromUnstructured(&item, &user); err != nil {
			return nil, fmt.Errorf("转换用户数据失败: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser 更新用户
func UpdateUser(user *User) error {
	ctx := context.Background()

	// 先获取现有资源获得 ResourceVersion
	existing, err := GlobalClient.DynamicClient.Resource(UserGVR).Get(
		ctx, user.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 设置 ResourceVersion（必须）
	user.ResourceVersion = existing.GetResourceVersion()

	unstructuredObj, err := utils.ToUnstructured(user)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	_, err = GlobalClient.DynamicClient.Resource(UserGVR).Update(
		ctx, unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	fmt.Printf("✅ 更新用户: %s\n", user.Name)
	return nil
}

// DeleteUser 删除用户
func DeleteUser(name string) error {
	ctx := context.Background()

	err := GlobalClient.DynamicClient.
		Resource(UserGVR).
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

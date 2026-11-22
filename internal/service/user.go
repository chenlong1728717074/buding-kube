package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/alexedwards/argon2id"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	userSrv  *UserService
	userOnce sync.Once
)

type UserService struct{}

func GetSingletonUserService() *UserService {
	userOnce.Do(func() {
		userSrv = NewUserService()
	})
	return userSrv
}

func NewUserService() *UserService {
	return &UserService{}
}

// emailRegexp 邮箱格式校验(允许为空, 填写时必须合法)
var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
	if email == "" {
		return nil
	}
	if !emailRegexp.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

// ensureEmailUnique 校验邮箱未被其他用户占用(允许为空)
func (s *UserService) ensureEmailUnique(email, excludeUsername string) error {
	if email == "" {
		return nil
	}
	users, err := kube.ListUsers(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range users {
		if users[i].Spec.Email == email && users[i].Name != excludeUsername {
			return errors.New("邮箱已被其他用户使用")
		}
	}
	return nil
}

// GetUser  根据用户名获取用户
func (s *UserService) GetUser(name string) (*kube.User, error) {
	var err error
	var usr *kube.User
	usr, err = kube.GetUser(name)
	if err != nil {
		logs.Error("获取用户失败: %s", err.Error())
		return nil, err
	}
	return usr, nil
}

// ListUsers 获取用户列表
// ErrPermissionDenied 没有权限
var ErrPermissionDenied = errors.New("权限不足")

// ListUsers 用户列表: 超级管理员查看全部, 管理员只查看普通用户, 普通用户没有权限
func (s *UserService) ListUsers(query dto.UserQueryDTO, currentUser *kube.LoginUser) ([]*vo.UserVO, error) {
	if currentUser == nil {
		return nil, ErrPermissionDenied
	}
	var visible func(role kube.Role) bool
	switch currentUser.Role {
	case kube.RoleSuperAdmin:
		visible = func(kube.Role) bool { return true }
	case kube.RoleAdmin:
		visible = func(role kube.Role) bool { return role == kube.RoleNormal }
	default:
		return nil, ErrPermissionDenied
	}
	users, err := kube.ListUsers(metav1.ListOptions{})
	if err != nil {
		logs.Error("获取用户列表失败: %s", err.Error())
		return nil, err
	}

	result := make([]*vo.UserVO, 0, len(users))

	for i := range users {
		u := &users[i]
		if !visible(u.Spec.Role) {
			continue
		}

		voUser := vo.User2VO(u)
		if voUser == nil {
			continue
		}

		if query.Username != "" &&
			!strings.Contains(voUser.Username, query.Username) {
			continue
		}

		if query.Status != "" &&
			voUser.Status != query.Status {
			continue
		}

		if query.Role != "" &&
			string(voUser.Role) != string(query.Role) {
			continue
		}

		result = append(result, voUser)
	}

	return result, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req dto.CreateUserDTO, currentUser *kube.LoginUser) error {
	var user *kube.User
	var err error
	_, err = kube.GetUser(req.Username)
	if err == nil {
		return errors.New("用户已存在")
	}
	if req.Password == "" {
		return errors.New("请输入密码")
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if err := s.ensureEmailUnique(req.Email, req.Username); err != nil {
		return err
	}
	if req.Role == kube.RoleSuperAdmin {
		return errors.New("不允许创建超级管理员")
	}
	user, err = BuildUserFromDTO(req)
	if err != nil {
		return fmt.Errorf("构建用户失败: %w", err)
	}
	if !currentUser.CanManageUser(user) {
		return errors.New("权限不足")
	}
	return kube.CreateUser(user, kube.GlobalClient.DynamicClient)
}

// UpdateUser 更新用户
// 注意: 用户名(username)是不可修改的，它作为唯一标识符用于查询用户
func (s *UserService) UpdateUser(req dto.UpdateUserDTO, currentUser *kube.LoginUser) error {
	var user *kube.User
	var err error
	user, err = kube.GetUser(req.Username)
	if err != nil {
		return err
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if err := s.ensureEmailUnique(req.Email, req.Username); err != nil {
		return err
	}
	if !currentUser.CanManageUser(user) {
		return errors.New("权限不足")
	}
	// 角色与状态是管理字段: 修改角色要求操作者是管理员且不能改自己的角色, 防止普通用户自行提权
	if req.Role != "" && req.Role != user.Spec.Role {
		if kube.RoleLevel(currentUser.Role) > kube.RoleLevel(kube.RoleAdmin) {
			return errors.New("权限不足")
		}
		if req.Username == currentUser.Username {
			return errors.New("不能修改自己的角色")
		}
	}
	if req.Username == kube.SuperAdmin && req.Role != kube.RoleSuperAdmin {
		return errors.New("不能修改admin用户的角色")
	}
	if req.Username != kube.SuperAdmin && req.Role == kube.RoleSuperAdmin {
		return errors.New("只有admin是超级管理员")
	}
	if req.Username == kube.SuperAdmin && req.Status != "" && req.Status != user.Status.State {
		return errors.New("admin 用户的启用状态和状态不允许修改")
	}
	UpdateUserFromDTO(req, user)
	return kube.UpdateUser(user)
}

// UpdateProfile 更新个人信息(仅允许修改本人的基本资料, 不涉及角色/状态)
func (s *UserService) UpdateProfile(req dto.ProfileUpdateDTO, currentUser *kube.LoginUser) error {
	user, err := kube.GetUser(currentUser.Username)
	if err != nil {
		return err
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if err := s.ensureEmailUnique(req.Email, user.Name); err != nil {
		return err
	}
	user.Spec.RealName = req.RealName
	user.Spec.Email = req.Email
	user.Spec.Phone = req.Phone
	return kube.UpdateUser(user)
}

func BuildUserFromDTO(req dto.CreateUserDTO) (*kube.User, error) {
	password, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	state := kube.UserStateActive
	if req.Status != "" {
		state = req.Status
	}
	return &kube.User{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", consts.GroupCrd, consts.ClusterCrdVersion),
			Kind:       consts.UserKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Username,
		},
		Spec: kube.UserSpec{
			Email:      req.Email,
			RealName:   req.RealName,
			Phone:      req.Phone,
			Password:   password,
			Role:       req.Role,
			Enabled:    state == kube.UserStateActive,
			Department: req.Department,
			Attributes: map[string]string{},
		},
		Status: kube.UserStatus{
			State: state,
		},
	}, nil
}

func UpdateUserFromDTO(req dto.UpdateUserDTO, user *kube.User) {
	user.Spec.RealName = req.RealName
	user.Spec.Phone = req.Phone
	user.Spec.Email = req.Email
	if req.Role != "" {
		user.Spec.Role = req.Role
	}
	user.Spec.Department = req.Department
	if req.Status != "" {
		user.Status.State = req.Status
		user.Spec.Enabled = req.Status == kube.UserStateActive
	}
}

// UpdateUserStatus 更新用户状态（active/inactive/suspended/expired）
func (s *UserService) UpdateUserStatus(req dto.UpdateUserStatusDTO, currentUser *kube.LoginUser) error {
	user, err := kube.GetUser(req.Username)
	if err != nil {
		return err
	}

	if !currentUser.CanManageUser(user) {
		return errors.New("权限不足")
	}

	switch req.Status {
	case kube.UserStateActive, kube.UserStateInactive, kube.UserStateSuspended, kube.UserStateExpired:
	default:
		return errors.New("非法状态，仅支持 active/inactive/suspended/expired")
	}

	if user.Name == kube.SuperAdmin && req.Status != user.Status.State {
		return errors.New("admin 用户的启用状态和状态不允许修改")
	}

	user.Status.State = req.Status
	user.Spec.Enabled = req.Status == "active"

	return kube.UpdateUser(user)
}

// UpdateUserEnable 启用/禁用用户（仅用于前端启用/禁用开关）
func (s *UserService) UpdateUserEnable(req dto.UpdateUserEnableDTO, currentUser *kube.LoginUser) error {
	user, err := kube.GetUser(req.Username)
	if err != nil {
		return err
	}

	if !currentUser.CanManageUser(user) {
		return errors.New("权限不足")
	}

	if user.Name == kube.SuperAdmin && req.Enable != user.Spec.Enabled {
		return errors.New("admin 用户的启用状态和状态不允许修改")
	}

	user.Spec.Enabled = req.Enable
	if req.Enable {
		user.Status.State = kube.UserStateActive
	} else {
		user.Status.State = kube.UserStateInactive
	}

	return kube.UpdateUser(user)
}

// BatchDeleteUsers 批量删除用户
func (s *UserService) BatchDeleteUsers(usernames []string, currentUser *kube.LoginUser) error {
	for _, username := range usernames {
		if err := s.DeleteUser(username, currentUser); err != nil {
			return fmt.Errorf("删除用户 %s 失败: %w", username, err)
		}
	}
	return nil
}

// ResetPassword 重置用户密码: newPassword 为空时使用默认密码 123456
func (s *UserService) ResetPassword(name, newPassword string, currentUser *kube.LoginUser) error {
	if name == currentUser.Username {
		// 自己改密码必须走 ChangePassword(校验旧密码), 否则任意用户可绕过旧密码校验重置自己的密码
		return errors.New("不能重置自己的密码, 请使用修改密码功能")
	}
	user, err := kube.GetUser(name)
	if err != nil {
		return err
	}

	if !currentUser.CanManageUser(user) {
		return errors.New("权限不足")
	}

	if strings.TrimSpace(newPassword) == "" {
		newPassword = "123456"
	}

	hashedPassword, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	user.Spec.Password = hashedPassword
	return kube.UpdateUser(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(name string, currentUser *kube.LoginUser) error {
	// 获取目标用户
	var user *kube.User
	var err error
	user, err = kube.GetUser(name)
	if err != nil {
		return err
	}
	// 不能删除自己
	if currentUser.Username == user.Name {
		return errors.New("不能删除自己")
	}

	// 不能删除 admin 超级管理员账号
	if user.Name == kube.SuperAdmin {
		return errors.New("admin 用户不允许删除")
	}

	// 检查权限
	if !currentUser.CanDeleteUser(user) {
		return errors.New("权限不足")
	}

	// 执行删除
	err = kube.DeleteUser(name)
	if err != nil {
		logs.Error("删除用户失败: %s", err.Error())
		return err
	}

	return nil
}

package service

import (
	"buding-kube/internal/model"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/kube"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/password"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
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

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, username)

	items, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		logs.Error("获取用户失败: %s", err.Error())
		return nil, err
	}

	if len(items.Items) == 0 {
		logs.Error("用户不存在")
		return nil, errors.New("用户不存在")
	}

	secret := items.Items[0]
	var user model.User
	if err := json.Unmarshal(secret.Data["config"], &user); err != nil {
		logs.Error("解析用户数据失败: %s", err.Error())
		return nil, err
	}

	return &user, nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(query dto.UserQueryDTO) ([]vo.UserVO, error) {
	labelSelector := fmt.Sprintf("%s", model.UserConfigSecretLabelKey)

	items, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		logs.Error("获取用户列表失败: %s", err.Error())
		return nil, err
	}

	userVOs := make([]vo.UserVO, 0)
	for _, secret := range items.Items {
		var user model.User
		if err := json.Unmarshal(secret.Data["config"], &user); err != nil {
			logs.Error("解析用户数据失败: %s", err.Error())
			continue
		}
		// 根据查询条件过滤
		if query.Username != "" && user.Username != query.Username {
			continue
		}
		if query.Role > 0 && int(user.Role) != query.Role {
			continue
		}
		userVO := vo.UserVO{}
		copier.Copy(&userVO, &user)
		userVOs = append(userVOs, userVO)
	}
	return userVOs, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req dto.CreateUserDTO, currentUser *model.User) error {
	// 检查权限
	if currentUser.Role > model.Admin {
		return errors.New("权限不足")
	}

	// 超管才能创建其他超管和管理员
	if req.Role < model.Normal && !currentUser.IsSuperAdmin() {
		return errors.New("普通管理员只能创建普通用户")
	}

	// 检查用户是否已存在
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, req.Username)
	existingUsers, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		logs.Error("检查用户失败: %s", err.Error())
		return err
	}

	if len(existingUsers.Items) > 0 {
		return errors.New("用户已存在")
	}

	// 创建用户
	user := model.NewUser(req.Username, req.Password, req.Email, req.Role)

	// 序列化用户信息
	userData, err := json.Marshal(user)
	if err != nil {
		logs.Error("序列化用户数据失败: %s", err.Error())
		return err
	}

	// 创建 Secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("user-%s", user.Username),
			Labels: map[string]string{
				model.UserConfigSecretLabelKey: user.Username,
			},
		},
		Data: map[string][]byte{
			"config": userData,
		},
		Type: v1.SecretTypeOpaque,
	}

	_, err = kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建用户失败: %s", err.Error())
		return err
	}

	return nil
}

// UpdateUser 更新用户
// 注意: 用户名(username)是不可修改的，它作为唯一标识符用于查询用户
func (s *UserService) UpdateUser(username string, req dto.UpdateUserDTO, currentUser *model.User) error {
	// 获取目标用户
	targetUser, err := s.GetUserByUsername(username)
	if err != nil {
		return err
	}

	// 检查权限
	if !currentUser.CanManageUser(targetUser) {
		return errors.New("权限不足")
	}

	// 非超级管理员不能修改用户角色为超级管理员
	if req.Role == model.SuperAdmin && !currentUser.IsSuperAdmin() {
		return errors.New("只有超级管理员才能设置超级管理员角色")
	}

	// 普通管理员不能修改其他管理员
	if targetUser.Role == model.Admin && currentUser.Role == model.Admin && currentUser.Username != targetUser.Username {
		return errors.New("管理员不能修改其他管理员")
	}

	// 更新字段
	if req.Email != "" {
		targetUser.Email = req.Email
	}
	if req.Role > 0 {
		targetUser.Role = req.Role
	}

	// 更新密码
	if req.Password != "" {
		newPassword, salt := password.GeneratePassword(req.Password)
		targetUser.Password = newPassword
		targetUser.Salt = salt
	}

	// 更新状态
	if req.Status != targetUser.Status {
		// 超级管理员不能被禁用
		if targetUser.IsSuperAdmin() && req.Status == model.UserStatusDisabled {
			return errors.New("不能禁用超级管理员")
		}
		targetUser.Status = req.Status
	}

	// 序列化并更新
	userData, err := json.Marshal(targetUser)
	if err != nil {
		logs.Error("序列化用户数据失败: %s", err.Error())
		return err
	}

	// 获取原 Secret
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, username)
	items, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil || len(items.Items) == 0 {
		logs.Error("获取用户 Secret 失败: %s", err.Error())
		return errors.New("用户不存在")
	}

	secret := items.Items[0]
	secret.Data["config"] = userData

	_, err = kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新用户失败: %s", err.Error())
		return err
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(username string, currentUser *model.User) error {
	// 获取目标用户
	targetUser, err := s.GetUserByUsername(username)
	if err != nil {
		return err
	}

	// 不能删除自己
	if currentUser.Username == username {
		return errors.New("不能删除自己")
	}

	// 不能删除超级管理员
	if targetUser.IsSuperAdmin() {
		return errors.New("不能删除超级管理员")
	}

	// 检查权限
	if !currentUser.CanDeleteUser(targetUser) {
		return errors.New("权限不足")
	}

	// 执行删除
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, username)
	items, err := kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil || len(items.Items) == 0 {
		logs.Error("获取用户 Secret 失败: %s", err.Error())
		return errors.New("用户不存在")
	}

	secretName := items.Items[0].Name

	err = kube.InClusterClientSet.CoreV1().Secrets(kube.ServerNamespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除用户失败: %s", err.Error())
		return err
	}

	return nil
}

package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"errors"
	"strings"
	"sync"

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
func (s *UserService) ListUsers(query dto.UserQueryDTO) ([]*vo.UserVO, error) {
	users, err := kube.ListUsers(metav1.ListOptions{})
	if err != nil {
		logs.Error("获取用户列表失败: %s", err.Error())
		return nil, err
	}

	result := make([]*vo.UserVO, 0, len(users))

	for i := range users {
		u := &users[i]

		voUser := vo.User2VO(u)
		if voUser == nil {
			continue
		}

		if query.Username != "" &&
			!strings.Contains(voUser.Username, query.Username) {
			continue
		}

		//if query.Role != "" &&
		//	voUser.Role != query.Role {
		//	continue
		//}

		if query.Status != "" &&
			voUser.Status != query.Status {
			continue
		}

		result = append(result, voUser)
	}

	return result, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req dto.CreateUserDTO) error {

	return nil
}

// UpdateUser 更新用户
// 注意: 用户名(username)是不可修改的，它作为唯一标识符用于查询用户
func (s *UserService) UpdateUser(req dto.CreateUserDTO) error {

	return nil
}

func (s *UserService) updateBasic(req dto.CreateUserDTO, targetUser *kube.User) {

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

	// 不能删除超级管理员
	if user.IsSuperAdmin() {
		return errors.New("不能删除超级管理员")
	}

	// 检查权限
	//if !currentUser.CanDeleteUser(targetUser) {
	//	return errors.New("权限不足")
	//}

	// 执行删除
	err = kube.DeleteUser(name)
	if err != nil {
		logs.Error("删除用户失败: %s", err.Error())
		return err
	}

	return nil
}

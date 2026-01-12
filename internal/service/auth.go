package service

import (
	kube2 "buding-kube/internal/kube"
	"buding-kube/internal/model"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"errors"
	"github.com/alexedwards/argon2id"
	"sync"
)

var (
	authSrv  *AuthService
	authOnce sync.Once
)

type AuthService struct {
}

func GetSingletonAuthService() *AuthService {
	authOnce.Do(func() {
		authSrv = NewAuthService()
	})
	return authSrv
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(login dto.LoginDTO) (*vo.UserVO, error) {
	var err error
	var user *kube2.User
	user, err = kube2.GetUser(login.Username)
	if err != nil {
		logs.Error("获取用户失败: %s", err.Error())
		return nil, err
	}
	if !user.Spec.Enabled {
		return nil, errors.New("用户已停用")
	}
	var match bool
	match, err = argon2id.ComparePasswordAndHash(login.Password, user.Spec.Password)
	if err != nil {
		return nil, errors.New("密码校验失败,原因是:" + err.Error())
	}
	if !match {
		return nil, errors.New("用户名或密码错误")
	}
	u := model.User{
		Username: login.Username,
		Role:     1,
		Status:   1,
		Email:    user.Spec.Email,
	}
	token, err := jwt.GenerateToken(&u)
	if err != nil {
		logs.Error("token 生成失败 %v", err)
		return nil, errors.New("token 生成失败")
	}
	return vo.User2VO(u, token), nil
}

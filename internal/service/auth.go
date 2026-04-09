package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"errors"
	"sync"

	"github.com/alexedwards/argon2id"
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

func (s *AuthService) Login(login dto.LoginDTO) (*vo.LoginUserVO, error) {
	var err error
	var user *kube.User
	user, err = kube.GetUser(login.Username)
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
	token, err := jwt.Token.GenerateToken(user)
	if err != nil {
		logs.Error("token 生成失败 %v", err)
		return nil, errors.New("token 生成失败")
	}
	return vo.ToLoginUserVO(user, token), nil
}

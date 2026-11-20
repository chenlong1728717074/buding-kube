package service

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"errors"
	"sync"
	"time"

	"github.com/alexedwards/argon2id"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// 更新最后登录时间(失败不影响登录)
	user.Status.LastLogin = &metav1.Time{Time: time.Now()}
	if err := kube.UpdateUser(user); err != nil {
		logs.Error("更新最后登录时间失败: %v", err)
	}

	token, err := jwt.Token.GenerateToken(user)
	if err != nil {
		logs.Error("token 生成失败 %v", err)
		return nil, errors.New("token 生成失败")
	}
	return vo.ToLoginUserVO(user, token), nil
}

// ChangePassword 当前登录用户修改密码
func (s *AuthService) ChangePassword(currentUser *kube.LoginUser, req dto.ChangePasswordDTO) error {
	user, err := kube.GetUser(currentUser.Username)
	if err != nil {
		return err
	}

	match, err := argon2id.ComparePasswordAndHash(req.CurrentPassword, user.Spec.Password)
	if err != nil {
		return errors.New("当前密码校验失败")
	}
	if !match {
		return errors.New("当前密码错误")
	}

	hashedPassword, err := argon2id.CreateHash(req.NewPassword, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	user.Spec.Password = hashedPassword
	return kube.UpdateUser(user)
}

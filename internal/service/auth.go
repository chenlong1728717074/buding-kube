package service

import (
	"buding-kube/internal/model"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/kube"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"buding-kube/pkg/utils/password"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, login.Username)
	item, err := kube.InClusterClientSet.CoreV1().Secrets("buding").
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		logs.Error("获取用户失败: %s", err.Error())
		return nil, err
	}
	if len(item.Items) == 0 {
		logs.Error("用户不存在")
		return nil, errors.New("用户不存在")
	}
	secret := item.Items[0]
	var user model.User
	json.Unmarshal(secret.Data["config"], &user)
	if user.Status == 0 {
		return nil, errors.New("用户已停用")
	}
	if !password.CheckPassword(login.Password, user.Salt, user.Password) {
		logs.Error("密码校验失败%s %s", login.Username, login.Password)
		return nil, errors.New("密码校验失败")
	}
	token, err := jwt.GenerateToken(&user)
	if err != nil {
		logs.Error("token 生成失败 %v", err)
		return nil, errors.New("token 生成失败")
	}
	return vo.User2VO(user, token), nil
}

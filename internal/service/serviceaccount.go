package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strings"
	"sync"
)

var (
	serviceAccountSrv  *ServiceAccountService
	serviceAccountOnce sync.Once
)

type ServiceAccountService struct {
}

func NewServiceAccountService() *ServiceAccountService {
	return &ServiceAccountService{}
}

func GetSingletonServiceAccountService() *ServiceAccountService {
	serviceAccountOnce.Do(func() {
		serviceAccountSrv = NewServiceAccountService()
	})
	return serviceAccountSrv
}

// List - 查询ServiceAccount列表
func (s *ServiceAccountService) List(ctx context.Context, query dto.ServiceAccountPageQueryBaseDTO) ([]vo.ServiceAccountVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	items, err := clientSet.CoreV1().ServiceAccounts(query.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取ServiceAccount失败: %v", err)
		return nil, fmt.Errorf("获取ServiceAccount失败: %v", err)
	}

	result := make([]vo.ServiceAccountVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.ServiceAccount2VO(item)
			result = append(result, vi)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Update - 更新ServiceAccount
func (s *ServiceAccountService) Update(ctx context.Context, update dto.ServiceAccountCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(update.ClusterId)
	if err != nil {
		return err
	}
	serviceAccount, err := s.getServiceAccount(ctx, clientSet, update.Namespace, update.Name)
	if err != nil {
		return err
	}

	// 更新 AutomountServiceAccountToken
	if update.AutomountServiceAccountToken != nil {
		serviceAccount.AutomountServiceAccountToken = update.AutomountServiceAccountToken
	}

	// 更新 ImagePullSecrets
	serviceAccount.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
	for _, secretName := range update.ImagePullSecrets {
		serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secretName,
		})
	}

	// 更新 Secrets
	serviceAccount.Secrets = make([]corev1.ObjectReference, 0)
	for _, secretName := range update.Secrets {
		serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{
			Name: secretName,
		})
	}

	// 更新 Annotations
	if update.Annotations != nil {
		serviceAccount.Annotations = update.Annotations
	}

	// 更新 Labels
	if update.Labels != nil {
		serviceAccount.Labels = update.Labels
	}

	_, err = clientSet.CoreV1().ServiceAccounts(update.Namespace).Update(ctx, serviceAccount, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ServiceAccount失败: %v", err)
		return err
	}
	logs.Info("ServiceAccount %s/%s 更新成功", update.Namespace, update.Name)
	return nil
}

// Delete - 删除ServiceAccount
func (s *ServiceAccountService) Delete(ctx context.Context, base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	if base.Name == "" || base.Namespace == "" {
		logs.Error("ServiceAccountName 或 Namespace 为空")
		return errors.New("ServiceAccountName 或 Namespace 为空")
	}

	err = clientSet.CoreV1().ServiceAccounts(base.Namespace).Delete(ctx, base.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除ServiceAccount失败: %v", err)
		return fmt.Errorf("删除ServiceAccount失败: %v", err)
	}
	logs.Info("ServiceAccount 删除成功: %s", base.Name)
	return nil
}

// getServiceAccount - 获取ServiceAccount
func (s *ServiceAccountService) getServiceAccount(ctx context.Context, clientSet *kubernetes.Clientset, namespace, name string) (*corev1.ServiceAccount, error) {
	serviceAccount, err := clientSet.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Error("ServiceAccount不存在: %s/%s", namespace, name)
			return nil, errors.New("ServiceAccount不存在")
		}
		logs.Error("获取ServiceAccount失败: %v", err)
		return nil, err
	}
	return serviceAccount, nil
}

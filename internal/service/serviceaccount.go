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
	"sigs.k8s.io/yaml"
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
func (s *ServiceAccountService) List(query dto.ServiceAccountPageQueryBaseDTO) ([]vo.ServiceAccountVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	items, err := clientSet.CoreV1().ServiceAccounts(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取ServiceAccount失败: %v", err)
		return nil, fmt.Errorf("获取ServiceAccount失败: %v", err)
	}

	result := make([]vo.ServiceAccountVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.ServiceAccount2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化ServiceAccount失败: %v", err)
			}
			vi.Yaml = string(yamlData)
			result = append(result, vi)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Save - 创建ServiceAccount
func (s *ServiceAccountService) Save(create dto.ServiceAccountCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(create.ClusterId)
	if err != nil {
		return err
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        create.Name,
			Namespace:   create.Namespace,
			Annotations: create.Annotations,
			Labels:      create.Labels,
		},
	}

	// 设置 AutomountServiceAccountToken
	if create.AutomountServiceAccountToken != nil {
		serviceAccount.AutomountServiceAccountToken = create.AutomountServiceAccountToken
	}

	// 设置 ImagePullSecrets
	if len(create.ImagePullSecrets) > 0 {
		serviceAccount.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
		for _, secretName := range create.ImagePullSecrets {
			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{
				Name: secretName,
			})
		}
	}

	// 设置 Secrets
	if len(create.Secrets) > 0 {
		serviceAccount.Secrets = make([]corev1.ObjectReference, 0)
		for _, secretName := range create.Secrets {
			serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{
				Name: secretName,
			})
		}
	}

	_, err = clientSet.CoreV1().ServiceAccounts(create.Namespace).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建ServiceAccount失败: %v", err)
		return err
	}
	logs.Info("ServiceAccount %s/%s 创建成功", create.Namespace, create.Name)
	return nil
}

// Update - 更新ServiceAccount
func (s *ServiceAccountService) Update(update dto.ServiceAccountCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(update.ClusterId)
	if err != nil {
		return err
	}
	serviceAccount, err := s.getServiceAccount(clientSet, update.Namespace, update.Name)
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

	_, err = clientSet.CoreV1().ServiceAccounts(update.Namespace).Update(context.TODO(), serviceAccount, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ServiceAccount失败: %v", err)
		return err
	}
	logs.Info("ServiceAccount %s/%s 更新成功", update.Namespace, update.Name)
	return nil
}

// Delete - 删除ServiceAccount
func (s *ServiceAccountService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	if base.Name == "" || base.Namespace == "" {
		logs.Error("ServiceAccountName 或 Namespace 为空")
		return errors.New("ServiceAccountName 或 Namespace 为空")
	}

	err = clientSet.CoreV1().ServiceAccounts(base.Namespace).Delete(context.TODO(), base.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除ServiceAccount失败: %v", err)
		return fmt.Errorf("删除ServiceAccount失败: %v", err)
	}
	logs.Info("ServiceAccount 删除成功: %s", base.Name)
	return nil
}

// Apply - 通过YAML应用ServiceAccount
func (s *ServiceAccountService) Apply(apply dto.BaseYamlApplyDTO) error {
	clientSet, err := ClusterMap.Get(apply.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", apply.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	var serviceAccount corev1.ServiceAccount
	err = yaml.Unmarshal([]byte(apply.Yaml), &serviceAccount)
	if err != nil {
		logs.Error("反序列化ServiceAccount失败: %v", err)
		return fmt.Errorf("反序列化ServiceAccount失败: %v", err)
	}

	if serviceAccount.Name != "" {
		existingServiceAccount, err := clientSet.CoreV1().ServiceAccounts(apply.Namespace).Get(context.TODO(), serviceAccount.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// 不存在则创建
				_, err = clientSet.CoreV1().ServiceAccounts(apply.Namespace).Create(context.TODO(), &serviceAccount, metav1.CreateOptions{})
				if err != nil {
					logs.Error("创建ServiceAccount失败: %v", err)
					return fmt.Errorf("创建ServiceAccount失败: %v", err)
				}
				logs.Info("ServiceAccount 创建成功: %s", serviceAccount.Name)
				return nil
			}
			logs.Error("获取现有ServiceAccount失败: %v", err)
			return fmt.Errorf("获取现有ServiceAccount失败: %v", err)
		}

		// 存在则更新
		existingServiceAccount.AutomountServiceAccountToken = serviceAccount.AutomountServiceAccountToken
		existingServiceAccount.ImagePullSecrets = serviceAccount.ImagePullSecrets
		existingServiceAccount.Secrets = serviceAccount.Secrets
		existingServiceAccount.Annotations = serviceAccount.Annotations
		existingServiceAccount.Labels = serviceAccount.Labels

		_, err = clientSet.CoreV1().ServiceAccounts(apply.Namespace).Update(context.TODO(), existingServiceAccount, metav1.UpdateOptions{})
		if err != nil {
			logs.Error("更新ServiceAccount失败: %v", err)
			return fmt.Errorf("更新ServiceAccount失败: %v", err)
		}
		logs.Info("ServiceAccount 更新成功: %s", serviceAccount.Name)
		return nil
	}

	return errors.New("ServiceAccount名称不能为空")
}

// getServiceAccount - 获取ServiceAccount
func (s *ServiceAccountService) getServiceAccount(clientSet *kubernetes.Clientset, namespace, name string) (*corev1.ServiceAccount, error) {
	serviceAccount, err := clientSet.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

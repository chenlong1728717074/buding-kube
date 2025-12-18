package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

var (
	secretSrv  *SecretService
	secretOnce sync.Once
)

type SecretService struct {
}

func NewSecretService() *SecretService {
	return &SecretService{}
}

func GetSingletonSecretService() *SecretService {
	secretOnce.Do(func() {
		secretSrv = NewSecretService()
	})
	return secretSrv
}

// List - 查询Secret列表
func (s *SecretService) List(query dto.ResourcePageQueryBaseDTO) ([]vo.SecretVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	namespace := query.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	items, err := clientSet.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取Secret失败: %v", err)
		return nil, fmt.Errorf("获取Secret失败: %v", err)
	}

	result := make([]vo.SecretVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.Secret2VO(item)
			vi.Data = nil
			vi.StringData = nil
			vi.Metadata = metav1.ObjectMeta{}
			vi.Annotations = nil
			vi.Labels = nil
			vi.ResourceVersion = ""
			vi.Uid = ""
			vi.Version = ""
			vi.Yaml = ""
			result = append(result, vi)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func (s *SecretService) Get(base dto.BaseDTO) (*vo.SecretVO, error) {
	clientSet, err := s.getClientSet(base.ClusterId)
	if err != nil {
		return nil, err
	}
	if base.Name == "" || base.Namespace == "" {
		return nil, errors.New("Secret名称或命名空间为空")
	}
	secret, err := s.getSecret(clientSet, base.Namespace, base.Name)
	if err != nil {
		return nil, err
	}

	vi := vo.Secret2VO(*secret)
	copy := secret.DeepCopy()
	copy.ObjectMeta.ManagedFields = nil
	copy.ObjectMeta.ResourceVersion = ""
	copy.ObjectMeta.CreationTimestamp = metav1.Time{}
	yamlData, err := yaml.Marshal(copy)
	if err != nil {
		logs.Error("序列化Secret失败: %v", err)
	} else {
		vi.Yaml = string(yamlData)
	}
	return &vi, nil
}

// Save - 创建Secret
func (s *SecretService) Save(create dto.SecretCreateDTO) error {
	clientSet, err := s.getClientSet(create.ClusterId)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        create.Name,
			Namespace:   create.Namespace,
			Annotations: make(map[string]string),
			Labels:      create.Labels,
		},
		Type: corev1.SecretType(create.Type),
	}

	// 设置别名和描述
	s.setAnnotations(secret, create.Alias, create.Describe, create.Annotations)

	// 优先使用 StringData，Kubernetes 会自动转换为 Base64 编码的 Data
	if create.StringData != nil && len(create.StringData) > 0 {
		secret.StringData = create.StringData
	} else if create.Data != nil {
		secret.Data = create.Data
	}

	// 如果没有指定类型，默认为 Opaque
	if secret.Type == "" {
		secret.Type = corev1.SecretTypeOpaque
	}

	_, err = clientSet.CoreV1().Secrets(create.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建Secret失败: %v", err)
		return err
	}
	logs.Info("Secret %s/%s 创建成功", create.Namespace, create.Name)
	return nil
}

// UpdateInfo - 只允许修改别名和描述
func (s *SecretService) UpdateInfo(update dto.SecretCreateDTO) error {
	clientSet, err := s.getClientSet(update.ClusterId)
	if err != nil {
		return err
	}

	secret, err := s.getSecret(clientSet, update.Namespace, update.Name)
	if err != nil {
		return err
	}

	s.setAnnotations(secret, update.Alias, update.Describe, nil)

	_, err = clientSet.CoreV1().Secrets(update.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新Secret信息失败: %v", err)
		return err
	}
	logs.Info("Secret %s/%s 信息更新成功", update.Namespace, update.Name)
	return nil
}

// UpdateData - 修改Secret数据
func (s *SecretService) UpdateData(updateData dto.SecretCreateDTO) error {
	clientSet, err := s.getClientSet(updateData.ClusterId)
	if err != nil {
		return err
	}

	secret, err := s.getSecret(clientSet, updateData.Namespace, updateData.Name)
	if err != nil {
		return err
	}

	// 使用 StringData 更新，Kubernetes 会自动处理编码
	if updateData.StringData != nil {
		secret.StringData = updateData.StringData
		// 清空旧的 Data，避免冲突
		secret.Data = nil
	} else {
		secret.Data = make(map[string][]byte)
	}

	_, err = clientSet.CoreV1().Secrets(updateData.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新Secret数据失败: %v", err)
		return err
	}
	logs.Info("Secret %s/%s 数据更新成功", updateData.Namespace, updateData.Name)
	return nil
}

func (s *SecretService) UpdateSetting(update dto.SecretCreateDTO) error {
	clientSet, err := s.getClientSet(update.ClusterId)
	if err != nil {
		return err
	}

	secret, err := s.getSecret(clientSet, update.Namespace, update.Name)
	if err != nil {
		return err
	}

	if update.Type != "" {
		secret.Type = corev1.SecretType(update.Type)
	}

	secret.Labels = update.Labels

	alias := ""
	describe := ""
	if secret.Annotations != nil {
		alias = secret.Annotations["alias"]
		describe = secret.Annotations["describe"]
	}
	newAnn := make(map[string]string)
	if alias != "" {
		newAnn["alias"] = alias
	}
	if describe != "" {
		newAnn["describe"] = describe
	}
	for k, v := range update.Annotations {
		if k == "alias" || k == "describe" {
			continue
		}
		newAnn[k] = v
	}
	secret.Annotations = newAnn

	_, err = clientSet.CoreV1().Secrets(update.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新Secret设置失败: %v", err)
		return err
	}
	logs.Info("Secret %s/%s 设置更新成功", update.Namespace, update.Name)
	return nil
}

// Apply BaseYamlApplyDTO - 通过YAML应用Secret
func (s *SecretService) Apply(apply dto.BaseYamlApplyDTO) error {
	clientSet, err := ClusterMap.Get(apply.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", apply.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	var secret corev1.Secret
	err = yaml.Unmarshal([]byte(apply.Yaml), &secret)
	if err != nil {
		logs.Error("反序列化Secret失败: %v", err)
		return fmt.Errorf("反序列化Secret失败: %v", err)
	}

	if secret.Name != "" {
		existingSecret, err := clientSet.CoreV1().Secrets(apply.Namespace).Get(context.TODO(), secret.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// 不存在则创建
				_, err = clientSet.CoreV1().Secrets(apply.Namespace).Create(context.TODO(), &secret, metav1.CreateOptions{})
				if err != nil {
					logs.Error("创建Secret失败: %v", err)
					return fmt.Errorf("创建Secret失败: %v", err)
				}
				logs.Info("Secret 创建成功: %s", secret.Name)
				return nil
			}
			logs.Error("获取现有Secret失败: %v", err)
			return fmt.Errorf("获取现有Secret失败: %v", err)
		}

		// 存在则更新
		existingSecret.Data = secret.Data
		existingSecret.StringData = secret.StringData
		existingSecret.Type = secret.Type
		existingSecret.Annotations = secret.Annotations
		existingSecret.Labels = secret.Labels

		_, err = clientSet.CoreV1().Secrets(apply.Namespace).Update(context.TODO(), existingSecret, metav1.UpdateOptions{})
		if err != nil {
			logs.Error("更新Secret失败: %v", err)
			return fmt.Errorf("更新Secret失败: %v", err)
		}
		logs.Info("Secret 更新成功: %s", secret.Name)
		return nil
	}

	return errors.New("Secret名称不能为空")
}

// Delete - 删除Secret
func (s *SecretService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	if base.Name == "" || base.Namespace == "" {
		logs.Error("SecretName 或 Namespace 为空")
		return errors.New("SecretName 或 Namespace 为空")
	}

	err = clientSet.CoreV1().Secrets(base.Namespace).Delete(context.TODO(), base.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除Secret失败: %v", err)
		return fmt.Errorf("删除Secret失败: %v", err)
	}
	logs.Info("Secret 删除成功: %s", base.Name)
	return nil
}

// getClientSet - 获取集群客户端
func (s *SecretService) getClientSet(clusterId string) (*kubernetes.Clientset, error) {
	clientSet, err := ClusterMap.Get(clusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", clusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	return clientSet, nil
}

// getSecret - 获取Secret
func (s *SecretService) getSecret(clientSet *kubernetes.Clientset, namespace, name string) (*corev1.Secret, error) {
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Error("Secret不存在: %s/%s", namespace, name)
			return nil, errors.New("Secret不存在")
		}
		logs.Error("获取Secret失败: %v", err)
		return nil, err
	}
	return secret, nil
}

// setAnnotations - 设置注解
func (s *SecretService) setAnnotations(secret *corev1.Secret, alias, describe string, annotations map[string]string) {
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	if alias != "" {
		secret.Annotations["alias"] = alias
	}
	if describe != "" {
		secret.Annotations["describe"] = describe
	}
	if annotations != nil {
		for key, value := range annotations {
			if key != "alias" && key != "describe" {
				secret.Annotations[key] = value
			}
		}
	}
}

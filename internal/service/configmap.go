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
	configMapSrv  *ConfigMapService
	configMapOnce sync.Once
)

type ConfigMapService struct {
}

func NewConfigMapService() *ConfigMapService {
	return &ConfigMapService{}
}

func GetSingletonConfigMapService() *ConfigMapService {
	configMapOnce.Do(func() {
		configMapSrv = NewConfigMapService()
	})
	return configMapSrv
}

func (s *ConfigMapService) List(query dto.ConfigMapPageQueryBaseDTO) ([]vo.ConfigMapVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().ConfigMaps(query.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取ConfigMap失败: %v", err)
		return nil, fmt.Errorf("获取ConfigMap失败: %v", err)
	}
	result := make([]vo.ConfigMapVO, 0)
	for _, item := range items.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			vi := vo.ConfigMap2VO(item)
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化ConfigMap失败: %v", err)
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

func (s *ConfigMapService) BaseYamlApplyDTO(apply dto.BaseYamlApplyDTO) error {
	clientSet, err := ClusterMap.Get(apply.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", apply.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	var configMap corev1.ConfigMap
	err = yaml.Unmarshal([]byte(apply.Yaml), &configMap)
	if err != nil {
		logs.Error("反序列化ConfigMap失败: %v", err)
		return fmt.Errorf("反序列化ConfigMap失败: %v", err)
	}

	if configMap.Name != "" {
		existingConfigMap, err := clientSet.CoreV1().ConfigMaps(apply.Namespace).Get(context.TODO(), configMap.Name, metav1.GetOptions{})
		if err != nil {
			logs.Error("获取现有ConfigMap失败: %v", err)
			return fmt.Errorf("获取现有ConfigMap失败: %v", err)
		}
		existingConfigMap.Data = configMap.Data
		existingConfigMap.Annotations = configMap.Annotations

		_, err = clientSet.CoreV1().ConfigMaps(apply.Namespace).Update(context.TODO(), existingConfigMap, metav1.UpdateOptions{})
		if err != nil {
			logs.Error("更新ConfigMap失败: %v", err)
			return fmt.Errorf("更新ConfigMap失败: %v", err)
		}
		logs.Info("ConfigMap 更新成功: %s", configMap.Name)
		return nil
	}

	_, err = clientSet.CoreV1().ConfigMaps(apply.Namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建ConfigMap失败: %v", err)
		return fmt.Errorf("创建ConfigMap失败: %v", err)
	}
	logs.Info("ConfigMap 创建成功: %s", configMap.Name)

	return nil
}

func (s *ConfigMapService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	// 获取
	if base.Name == "" || base.Namespace == "" {
		logs.Error("ConfigMapName 或 Namespace 为空")
		return errors.New("ConfigMapName 或 Namespace 为空")
	}
	// 删除
	err = clientSet.CoreV1().ConfigMaps(base.Namespace).Delete(context.TODO(), base.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除ConfigMap失败: %v", err)
		return fmt.Errorf("删除ConfigMap失败: %v", err)
	}
	logs.Info("ConfigMap 删除成功: %s", base.Name)
	return nil
}

func (s *ConfigMapService) Save(create dto.ConfigMapCreateDTO) error {
	clientSet, err := s.getClientSet(create.ClusterId)
	if err != nil {
		return err
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        create.Name,
			Namespace:   create.Namespace,
			Annotations: make(map[string]string),
		},
		Data: make(map[string]string),
	}
	s.setAnnotations(configMap, create.Alias, create.Describe, create.Annotations)
	if create.Data != nil {
		configMap.Data = create.Data
	}
	_, err = clientSet.CoreV1().ConfigMaps(create.Namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建ConfigMap失败: %v", err)
		return err
	}
	logs.Info("ConfigMap %s/%s 创建成功", create.Namespace, create.Name)
	return nil
}

func (s *ConfigMapService) UpdateInfo(update dto.BaseInfoDTO) error {
	clientSet, err := s.getClientSet(update.ClusterId)
	if err != nil {
		return err
	}
	configMap, err := s.getConfigMap(clientSet, update.Namespace, update.Name)
	if err != nil {
		return err
	}
	s.setAnnotations(configMap, update.Alias, update.Describe, nil)
	_, err = clientSet.CoreV1().ConfigMaps(update.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ConfigMap信息失败: %v", err)
		return err
	}
	logs.Info("ConfigMap %s/%s 信息更新成功", update.Namespace, update.Name)
	return nil
}

func (s *ConfigMapService) UpdateData(updateData dto.ConfigMapDataDTO) error {
	clientSet, err := s.getClientSet(updateData.ClusterId)
	if err != nil {
		return err
	}
	configMap, err := s.getConfigMap(clientSet, updateData.Namespace, updateData.Name)
	if err != nil {
		return err
	}
	if updateData.Data != nil {
		configMap.Data = updateData.Data
	} else {
		configMap.Data = make(map[string]string)
	}
	_, err = clientSet.CoreV1().ConfigMaps(updateData.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ConfigMap数据失败: %v", err)
		return err
	}
	logs.Info("ConfigMap %s/%s 数据更新成功", updateData.Namespace, updateData.Name)
	return nil
}

func (s *ConfigMapService) getClientSet(clusterId string) (*kubernetes.Clientset, error) {
	clientSet, err := ClusterMap.Get(clusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", clusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	return clientSet, nil
}

func (s *ConfigMapService) getConfigMap(clientSet *kubernetes.Clientset, namespace, name string) (*corev1.ConfigMap, error) {
	configMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Error("ConfigMap不存在: %s/%s", namespace, name)
			return nil, errors.New("ConfigMap不存在")
		}
		logs.Error("获取ConfigMap失败: %v", err)
		return nil, err
	}
	return configMap, nil
}

func (s *ConfigMapService) setAnnotations(configMap *corev1.ConfigMap, alias, describe string, annotations map[string]string) {
	if configMap.Annotations == nil {
		configMap.Annotations = make(map[string]string)
	}
	if alias != "" {
		configMap.Annotations["alias"] = alias
	}
	if describe != "" {
		configMap.Annotations["describe"] = describe
	}
	if annotations != nil {
		for key, value := range annotations {
			if key != "alias" && key != "describe" {
				configMap.Annotations[key] = value
			}
		}
	}
}

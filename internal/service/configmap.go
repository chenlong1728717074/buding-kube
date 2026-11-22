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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
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

// List 查询 ConfigMap 列表.
// 列表页只展示名称 / 别名 / 创建时间, 而 ConfigMap 的 data 可能很大(例如 Grafana 仪表盘、
// Prometheus 规则, 单个就有几百 KB). 之前的实现拉取完整对象, 还给每一条都生成一遍 YAML,
// 数据量大时请求很慢. 现在改为 Table 格式只取元数据(见 listAsTable), data 和 YAML 由详情接口按需获取
func (s *ConfigMapService) List(ctx context.Context, query dto.ConfigMapPageQueryBaseDTO) ([]vo.ConfigMapVO, error) {
	clientSet, err := s.getClientSet(query.ClusterId)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	items, err := s.listConfigMapSummaries(ctx, clientSet, query.Namespace)
	if err != nil {
		// 极老版本的集群不支持 Table 格式时, 回退为普通列表
		logs.Warn("以 Table 格式查询 ConfigMap 失败, 回退为普通列表: %v", err)
		items, err = s.listConfigMapsFallback(ctx, clientSet, query.Namespace)
		if err != nil {
			logs.Error("获取ConfigMap失败: %v", err)
			return nil, fmt.Errorf("获取ConfigMap失败: %v", err)
		}
	}

	result := make([]vo.ConfigMapVO, 0, len(items))
	for _, item := range items {
		if query.Keyword != "" && !strings.Contains(item.Name, query.Keyword) {
			continue
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func (s *ConfigMapService) listConfigMapSummaries(ctx context.Context, clientSet *kubernetes.Clientset, namespace string) ([]vo.ConfigMapVO, error) {
	rows, err := listAsTable(ctx, clientSet.CoreV1().RESTClient(), "configmaps", namespace)
	if err != nil {
		return nil, err
	}
	result := make([]vo.ConfigMapVO, 0, len(rows))
	for _, row := range rows {
		item := vo.ConfigMapSummary2VO(row.Meta)
		item.DataCount = row.cellInt("Data")
		result = append(result, item)
	}
	return result, nil
}

func (s *ConfigMapService) listConfigMapsFallback(ctx context.Context, clientSet *kubernetes.Clientset, namespace string) ([]vo.ConfigMapVO, error) {
	list, err := clientSet.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	result := make([]vo.ConfigMapVO, 0, len(list.Items))
	for i := range list.Items {
		item := vo.ConfigMapSummary2VO(list.Items[i].ObjectMeta)
		item.DataCount = len(list.Items[i].Data) + len(list.Items[i].BinaryData)
		result = append(result, item)
	}
	return result, nil
}

// Get 查询单个 ConfigMap 的完整信息(含 data 和 YAML)
func (s *ConfigMapService) Get(ctx context.Context, base dto.BaseDTO) (*vo.ConfigMapVO, error) {
	if base.Name == "" || base.Namespace == "" {
		return nil, errors.New("ConfigMap名称或命名空间为空")
	}
	clientSet, err := s.getClientSet(base.ClusterId)
	if err != nil {
		return nil, err
	}
	cm, err := clientSet.CoreV1().ConfigMaps(base.Namespace).Get(ctx, base.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.New("ConfigMap不存在")
		}
		logs.Error("获取ConfigMap失败: %v", err)
		return nil, err
	}

	result := vo.ConfigMap2VO(*cm)
	cmCopy := cm.DeepCopy()
	cmCopy.ManagedFields = nil
	cmCopy.ResourceVersion = ""
	cmCopy.CreationTimestamp = metav1.Time{}
	yamlData, err := yaml.Marshal(cmCopy)
	if err != nil {
		logs.Error("序列化ConfigMap失败: %v", err)
	} else {
		result.Yaml = string(yamlData)
	}
	result.DataCount = len(cm.Data) + len(cm.BinaryData)
	return &result, nil
}

func (s *ConfigMapService) Apply(apply dto.BaseYamlApplyDTO) error {
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

	if configMap.Name == "" {
		return errors.New("ConfigMap名称不能为空")
	}

	namespace := apply.Namespace
	if namespace == "" {
		namespace = configMap.Namespace
	}
	if namespace == "" {
		return errors.New("Namespace不能为空")
	}

	existingConfigMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMap.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			configMap.Namespace = namespace
			_, err = clientSet.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
			if err != nil {
				logs.Error("创建ConfigMap失败: %v", err)
				return fmt.Errorf("创建ConfigMap失败: %v", err)
			}
			logs.Info("ConfigMap 创建成功: %s", configMap.Name)
			return nil
		}
		logs.Error("获取现有ConfigMap失败: %v", err)
		return fmt.Errorf("获取现有ConfigMap失败: %v", err)
	}

	existingConfigMap.Data = configMap.Data
	existingConfigMap.Annotations = configMap.Annotations
	existingConfigMap.Labels = configMap.Labels

	_, err = clientSet.CoreV1().ConfigMaps(namespace).Update(context.TODO(), existingConfigMap, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ConfigMap失败: %v", err)
		return fmt.Errorf("更新ConfigMap失败: %v", err)
	}
	logs.Info("ConfigMap 更新成功: %s", configMap.Name)
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
	if create.Labels != nil {
		configMap.Labels = create.Labels
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

func (s *ConfigMapService) UpdateSetting(update dto.ConfigMapSettingDTO) error {
	clientSet, err := s.getClientSet(update.ClusterId)
	if err != nil {
		return err
	}
	configMap, err := s.getConfigMap(clientSet, update.Namespace, update.Name)
	if err != nil {
		return err
	}

	alias := ""
	describe := ""
	if configMap.Annotations != nil {
		alias = configMap.Annotations["alias"]
		describe = configMap.Annotations["describe"]
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
	configMap.Annotations = newAnn
	configMap.Labels = update.Labels

	_, err = clientSet.CoreV1().ConfigMaps(update.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新ConfigMap设置失败: %v", err)
		return err
	}
	logs.Info("ConfigMap %s/%s 设置更新成功", update.Namespace, update.Name)
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

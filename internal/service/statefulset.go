package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

var (
	statefulSetSrv  *StatefulSetService
	statefulSetOnce sync.Once
)

type StatefulSetService struct {
}

func NewStatefulSetService() *StatefulSetService {
	return &StatefulSetService{}
}

func GetSingletonStatefulSetService() *StatefulSetService {
	statefulSetOnce.Do(func() {
		statefulSetSrv = NewStatefulSetService()
	})
	return statefulSetSrv
}

func (s *StatefulSetService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	// 支持空命名空间查询所有命名空间
	namespace := query.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	sts, err := clientSet.AppsV1().StatefulSets(namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取StatefulSets失败: %v", err)
		return nil, err
	}
	result := make([]vo.WorkloadVO, 0)
	for _, item := range sts.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			yamlData, err := yaml.Marshal(item)
			if err != nil {
				logs.Error("序列化StatefulSet失败: %v", err)
			}
			sv := vo.StatefulSet2WorkloadVO(item)
			sv.Yaml = string(yamlData)
			result = append(result, sv)
		}
	}
	return result, nil
}

func (s *StatefulSetService) Update(query dto.WorkloadUpdateDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	sts, err := clientSet.AppsV1().StatefulSets(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取StatefulSet失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("获取StatefulSet失败 %v", err)
	}
	if query.Alias != "" {
		sts.Annotations["alias"] = query.Alias
	}
	if query.Describe != "" {
		sts.Annotations["describe"] = query.Describe
	}
	_, err = clientSet.AppsV1().StatefulSets(query.Namespace).Update(context.TODO(), sts, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("修改StatefulSet失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("修改StatefulSet失败 %v", err)
	}
	return nil
}

func (s *StatefulSetService) Delete(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	err = clientSet.AppsV1().StatefulSets(query.Namespace).Delete(context.TODO(), query.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("删除失败 %v", err)
	}
	return nil
}

func (s *StatefulSetService) Rollout(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	// 获取当前的StatefulSet
	statefulSet, err := clientSet.AppsV1().StatefulSets(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取StatefulSet失败: %v", err)
		return fmt.Errorf("获取StatefulSet失败: %v", err)
	}

	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = make(map[string]string)
	}
	statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新StatefulSet
	_, err = clientSet.AppsV1().StatefulSets(query.Namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("重建StatefulSet失败: %v", err)
		return fmt.Errorf("重建StatefulSet失败: %v", err)
	}
	logs.Info("成功触发StatefulSet [%s/%s] 重建", query.Namespace, query.Name)
	return nil
}

func (s *StatefulSetService) Apply(query dto.WorkloadApplyDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	var sts v1.StatefulSet

	err = yaml.Unmarshal([]byte(query.Yaml), &sts)
	if err != nil {
		logs.Error("解析yaml失败: %v", err)
		return err
	}
	existingStatefulSet, err := clientSet.AppsV1().StatefulSets(query.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})

	if err != nil {
		//不存在则创建
		if apierrors.IsNotFound(err) {
			logs.Info("StatefulSet不存在，开始创建: %s/%s", query.Namespace, sts.Name)
			_, err := clientSet.AppsV1().StatefulSets(query.Namespace).Create(context.TODO(), &sts, metav1.CreateOptions{})
			if err != nil {
				logs.Error("创建StatefulSet失败: %s %v", query.ClusterId, err)
				return fmt.Errorf("创建StatefulSet失败 %v", err)
			}
			logs.Info("创建StatefulSet成功: %s/%s", query.Namespace, sts.Name)
			return nil
		}
		return err
	}
	sts.ResourceVersion = existingStatefulSet.ResourceVersion
	_, err = clientSet.AppsV1().StatefulSets(query.Namespace).Update(context.TODO(), &sts, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("修改StatefulSet失败: %s %v", query.ClusterId, err)
		return fmt.Errorf("修改StatefulSet失败 %v", err)
	}
	logs.Info("更新StatefulSet成功: %s/%s", query.Namespace, sts.Name)
	return nil
}

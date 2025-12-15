package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	daemonSetSrv  *DaemonSetService
	daemonSetOnce sync.Once
)

type DaemonSetService struct {
}

func NewDaemonSetService() *DaemonSetService {
	return &DaemonSetService{}
}

func GetSingletonDaemonSetService() *DaemonSetService {
	daemonSetOnce.Do(func() {
		daemonSetSrv = NewDaemonSetService()
	})
	return daemonSetSrv
}

func (s *DaemonSetService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
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
	ds, err := clientSet.AppsV1().DaemonSets(namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取DaemonSets失败: %v", err)
		return nil, err
	}
	result := make([]vo.WorkloadVO, 0)
	for _, item := range ds.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			copy.Status = v1.DaemonSetStatus{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化DaemonSet失败: %v", err)
			}
			dv := vo.DaemonSet2WorkloadVO(item)
			dv.Yaml = string(yamlData)
			result = append(result, dv)
		}
	}
	return result, nil
}

func (s *DaemonSetService) Update(query dto.WorkloadUpdateDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	ds, err := clientSet.AppsV1().DaemonSets(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取DaemonSet失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("获取DaemonSet失败 %v", err)
	}
	if query.Alias != "" {
		ds.Annotations["alias"] = query.Alias
	}
	if query.Describe != "" {
		ds.Annotations["describe"] = query.Describe
	}
	_, err = clientSet.AppsV1().DaemonSets(query.Namespace).Update(context.TODO(), ds, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("修改DaemonSet失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("修改DaemonSet失败 %v", err)
	}
	return nil
}

func (s *DaemonSetService) Delete(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	err = clientSet.AppsV1().DaemonSets(query.Namespace).Delete(context.TODO(), query.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("删除失败 %v", err)
	}
	return nil
}

func (s *DaemonSetService) Rollout(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	// 获取当前的DaemonSet
	daemonSet, err := clientSet.AppsV1().DaemonSets(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取DaemonSet失败: %v", err)
		return fmt.Errorf("获取DaemonSet失败: %v", err)
	}

	if daemonSet.Spec.Template.Annotations == nil {
		daemonSet.Spec.Template.Annotations = make(map[string]string)
	}
	daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新DaemonSet
	_, err = clientSet.AppsV1().DaemonSets(query.Namespace).Update(context.TODO(), daemonSet, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("重建DaemonSet失败: %v", err)
		return fmt.Errorf("重建DaemonSet失败: %v", err)
	}
	logs.Info("成功触发DaemonSet [%s/%s] 重建", query.Namespace, query.Name)
	return nil
}

func (s *DaemonSetService) Apply(query dto.WorkloadApplyDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	var ds v1.DaemonSet

	err = yaml.Unmarshal([]byte(query.Yaml), &ds)
	if err != nil {
		logs.Error("解析yaml失败: %v", err)
		return err
	}
	existingDaemonSet, err := clientSet.AppsV1().DaemonSets(query.Namespace).Get(context.TODO(), ds.Name, metav1.GetOptions{})

	if err != nil {
		//不存在则创建
		if apierrors.IsNotFound(err) {
			logs.Info("DaemonSet不存在，开始创建: %s/%s", query.Namespace, ds.Name)
			_, err := clientSet.AppsV1().DaemonSets(query.Namespace).Create(context.TODO(), &ds, metav1.CreateOptions{})
			if err != nil {
				logs.Error("创建DaemonSet失败: %s %v", query.ClusterId, err)
				return fmt.Errorf("创建DaemonSet失败 %v", err)
			}
			logs.Info("创建DaemonSet成功: %s/%s", query.Namespace, ds.Name)
			return nil
		}
		return err
	}
	ds.ResourceVersion = existingDaemonSet.ResourceVersion
	_, err = clientSet.AppsV1().DaemonSets(query.Namespace).Update(context.TODO(), &ds, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("修改DaemonSet失败: %s %v", query.ClusterId, err)
		return fmt.Errorf("修改DaemonSet失败 %v", err)
	}
	logs.Info("更新DaemonSet成功: %s/%s", query.Namespace, ds.Name)
	return nil
}

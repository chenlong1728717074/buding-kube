package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"sync"
)

var (
	namespaceSrv  *NamespaceService
	namespaceOnce sync.Once
)

type NamespaceService struct {
}

func GetSingletonNamespaceService() *NamespaceService {
	namespaceOnce.Do(func() {
		namespaceSrv = NewNamespaceService()
	})
	return namespaceSrv
}

func NewNamespaceService() *NamespaceService {
	return &NamespaceService{}
}

func (s NamespaceService) Save(create dto.NamespaceCreateDTO) error {
	//获取连接
	clientSet, err := ClusterMap.Get(create.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", create.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	// 检查命名空间是否存在
	_, err = clientSet.CoreV1().Namespaces().Get(context.TODO(), create.Namespace, metav1.GetOptions{})
	if err == nil {
		// 命名空间已存在
		logs.Error("命名空间已存在: %s", create.Namespace)
		return errors.New("命名空间已存在")
	}

	if !apierrors.IsNotFound(err) {
		logs.Error("获取命名空间失败: %v", err)
		return err
	}

	// 命名空间不存在，创建新的命名空间
	logs.Info("创建新命名空间: %s", create.Namespace)

	// 定义命名空间对象
	nsSpec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: create.Namespace,
			Annotations: map[string]string{
				"alias":    create.Alias,
				"describe": create.Describe,
			},
		},
	}
	// 创建命名空间
	_, err = clientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metav1.CreateOptions{})
	if err != nil {
		logs.Error("创建命名空间失败: %v", err)
		return err
	}
	logs.Info("命名空间 %s 创建成功", create.Namespace)
	return nil
}

func (s NamespaceService) Update(create dto.NamespaceCreateDTO) error {
	//获取连接
	clientSet, err := ClusterMap.Get(create.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", create.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	// 获取命名空间
	ns, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), create.Namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Error("命名空间不存在: %s", create.Namespace)
			return errors.New("命名空间不存在")
		}
		logs.Error("获取命名空间失败: %v", err)
		return err
	}

	// 更新命名空间注释
	logs.Info("更新命名空间信息: %s", create.Namespace)
	if ns.Annotations == nil {
		ns.Annotations = make(map[string]string)
	}
	ns.Annotations["alias"] = create.Alias
	ns.Annotations["describe"] = create.Describe

	_, err = clientSet.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新命名空间失败: %v", err)
		return err
	}
	logs.Info("命名空间 %s 更新成功", create.Namespace)
	return nil
}

func (s NamespaceService) Delete(base dto.NamespaceBaseDTO) error {
	//获取连接
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if base.Namespace == apiv1.NamespaceSystem {
		return errors.New("禁止删除 kube-system 命名空间")
	}
	options := metav1.DeleteOptions{}
	if base.Force {
		//设置为 0 表示跳过优雅关闭阶段，立即删除资源
		//级联删除所有依赖资源（阻塞直到子资源都删除）
		flag := int64(0)
		policy := metav1.DeletePropagationForeground
		options.GracePeriodSeconds = &flag
		options.PropagationPolicy = &policy
	}
	if err := clientSet.CoreV1().Namespaces().Delete(context.TODO(), base.Namespace, options); err != nil {
		logs.Error("删除命名空间失败: %v", err)
		return err
	}
	clientSet.CoreV1()
	return nil
}

func (s NamespaceService) List(query dto.NamespacePageQueryBaseDTO) ([]vo.NamespaceVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取命名空间失败: %v", err)
		return nil, err
	}
	result := make([]vo.NamespaceVO, 0)
	for _, item := range namespaces.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			result = append(result, vo.Namespace2VO(&item))
		}
	}
	//按照时间倒叙排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTimestamp.After(result[j].CreationTimestamp)
	})
	return result, nil
}

func (s NamespaceService) GetById(base dto.NamespaceBaseDTO) (*vo.NamespaceVO, error) {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	ns, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), base.Namespace, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取命名空间失败: %v", err)
		return nil, err
	}
	result := vo.Namespace2VO(ns)
	return &result, nil
}

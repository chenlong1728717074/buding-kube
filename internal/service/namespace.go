package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clientSet := ClusterMap[create.ClusterId]
	if clientSet == nil {
		logs.Error("集群不存在: %s", create.ClusterId)
		return errors.New("集群不存在")
	}

	// 检查命名空间是否存在
	_, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), create.Namespace, metav1.GetOptions{})
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
	clientSet := ClusterMap[create.ClusterId]
	if clientSet == nil {
		logs.Error("集群不存在: %s", create.ClusterId)
		return errors.New("集群不存在")
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

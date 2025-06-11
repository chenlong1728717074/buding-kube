package kube

import (
	"buding-kube/pkg/logs"
	"context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var InClusterClientSet *kubernetes.Clientset

func init() {
	config, err := clientcmd.BuildConfigFromFlags("", "configs/k3s.yaml")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logs.Fatal("获取k8s连接失败")
	}
	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), "buding", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Info("命名空间不存在")
			// 定义命名空间对象
			nsSpec := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "buding",
				},
			}
			// 创建命名空间
			ns, err = clientset.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
			if err != nil {
				logs.Fatal("创建命名空间失败: %v", err)
			}
			logs.Info("命名空间 %s 创建成功。", ns.Name)
		} else {
			logs.Fatal("获取命名空间失败: %v", err)
		}
	}
	InClusterClientSet = clientset
}

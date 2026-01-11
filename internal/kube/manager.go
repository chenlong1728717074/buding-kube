package kube

import (
	"fmt"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClientManager struct {
	Clientset *kubernetes.Clientset
	//kubernetes.Client 无法操作CRD，需要用apiextensionsclientset的
	CRDClient     *apiextensionsclientset.Clientset
	DynamicClient dynamic.Interface
	Config        *rest.Config
	Namespace     string
}

// NewClientManager k8s管理工具的管理器
func NewClientManager(kubeConfig *rest.Config) (*ClientManager, error) {
	// 2. 创建 K8s 原生资源客户端
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Clientset 失败: %w", err)
	}

	// 3. 创建 CRD 客户端 (用于管理 CRD 定义)
	crdClient, err := apiextensionsclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 CRD Client 失败: %w", err)
	}

	// 4. 创建动态客户端 (用于操作 CRD 实例)
	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Dynamic Client 失败: %w", err)
	}

	return &ClientManager{
		Clientset:     clientset,
		CRDClient:     crdClient,
		DynamicClient: dynamicClient,
		Config:        kubeConfig,
	}, nil
}

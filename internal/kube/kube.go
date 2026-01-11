package kube

import (
	"buding-kube/internal/model"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var (
	InClusterClientSet      *kubernetes.Clientset
	ServerNamespace         string
	DefaultServerNamespace  = "kube-buding"
	DefaultServerConfigPath = "configs/conf.yaml"
	GlobalClient            *ClientManager
)

func init() {
	var err error
	var kubeConfig *rest.Config
	kubeConfig, err = getConfig()
	//serverConfig := config.GetConfig()

	if err != nil {
		logs.Fatal("初始化k8s配置失败 %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logs.Fatal("获取k8s连接失败")
	}
	if err := initNs(clientSet); err != nil {
		logs.Fatal("初始化命名空间失败")
	}

	if err := initSuperUser(clientSet); err != nil {
		logs.Fatal("初始化超级用户失败")
	}

	InClusterClientSet = clientSet
	init2()
}

func init2() {
	var err error
	//加载集群配置
	kubeConfig, namespace, err := getConfig2()
	//加载集群全局管理器
	GlobalClient, err = NewClientManager(kubeConfig)
	if err != nil {
		logs.Fatal("加载集群配置,初始化集群管理器失败 %v", err)
	}
	GlobalClient.Namespace = namespace
	if err = InitNamespace(GlobalClient.Clientset, namespace); err != nil {
		logs.Fatal("初始化 集群命名空间 失败 %v", err)
	}
	// 初始化 CRD ，用于管理user和cluster
	crds := []*apiextensionsv1.CustomResourceDefinition{
		buildClusterCRD(),
		buildUserCRD(),
	}
	if err = EnsureCRDs(crds, GlobalClient.CRDClient); err != nil {
		logs.Fatal("初始化 CRD 失败 %v", err)
	}
	//初始化超级用户
	if err = InitSuperAdmin(GlobalClient.DynamicClient); err != nil {
		logs.Fatal("初始化失败,问题是: %v", err)
	}
}

func getConfig() (*rest.Config, error) {
	//当在k8s环境中部署的时候我们就用k8s的配置
	env := os.Getenv("KUBE_RUNTIME_ENV")
	if env == "KUBE_POD" {
		data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, err
		}
		ServerNamespace = string(data)
		return rest.InClusterConfig()
	}
	//如果不是k8s的环境就拿环境变量
	ns := os.Getenv("KUBE_RUNTIME_NAMESPACE")
	ServerNamespace = DefaultServerNamespace
	if ns != "" {
		ServerNamespace = ns
	}
	//加载非k8s环境的配置文件
	configPath := os.Getenv("KUBE_RUNTIME_NAMESPACE")
	if configPath == "" {
		configPath = DefaultServerConfigPath
	}
	return clientcmd.BuildConfigFromFlags("", configPath)
}

func getConfig2() (*rest.Config, string, error) {
	env := os.Getenv("KUBE_RUNTIME_ENV")
	var namespace string

	// 在 K8s Pod 内运行
	if env == "KUBE_POD" {
		// 读取当前 Pod 所在的命名空间
		data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, "", fmt.Errorf("读取命名空间失败: %w", err)
		}
		namespace = string(data)

		// 使用 InClusterConfig ，在k8s集群中运行的pod可以执行这个
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, "", fmt.Errorf("获取集群内配置失败: %w", err)
		}
		return config, namespace, nil
	}

	// 如果不再k8s容器中运行的话就需要 从环境变量读取命名空间
	namespace = os.Getenv("KUBE_RUNTIME_NAMESPACE")
	if namespace == "" {
		namespace = DefaultServerNamespace
	}

	// 从环境变量读取配置文件路径
	configPath := os.Getenv("KUBE_CONFIG_PATH")
	if configPath == "" {
		configPath = DefaultServerConfigPath
	}

	// 加载 kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, "", fmt.Errorf("加载配置文件失败: %w", err)
	}

	return config, namespace, nil
}

func initSuperUser(clientSet *kubernetes.Clientset) error {
	labelSelector := fmt.Sprintf("%s=%s", model.UserConfigSecretLabelKey, model.UserConfigSecretLabelValue)
	item, err := clientSet.CoreV1().Secrets(ServerNamespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		logs.Fatal("获取管理员失败: %v", err)
	}
	if len(item.Items) != 0 {
		return nil
	}
	logs.Info("管理员不存在,开始初始化管理员")
	user := model.AdminUser()
	marshal, _ := json.Marshal(user)
	adminSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: ServerNamespace,
			Labels: map[string]string{
				model.UserConfigSecretLabelKey: model.UserConfigSecretLabelValue,
			},
		},
		StringData: map[string]string{
			"config": string(marshal),
		},
		Type: corev1.SecretTypeOpaque,
	}

	_, createErr := clientSet.CoreV1().Secrets(ServerNamespace).Create(context.TODO(), adminSecret, metav1.CreateOptions{})
	if createErr != nil {
		return fmt.Errorf("创建管理员 Secret 失败: %w", createErr)
	}
	logs.Info("默认管理员初始化成功")
	return nil
}

func initNs(clientSet *kubernetes.Clientset) error {
	ns, err := clientSet.CoreV1().Namespaces().Get(context.Background(), ServerNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logs.Info("命名空间不存在")
			// 定义命名空间对象
			nsSpec := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ServerNamespace,
				},
			}
			// 创建命名空间
			ns, err = clientSet.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
			if err != nil {
				logs.Fatal("创建命名空间失败: %v", err)
			}
			logs.Info("命名空间 %s 创建成功。", ns.Name)
		} else {
			logs.Fatal("获取命名空间失败: %v", err)
		}
	}
	return nil
}

func InitSuperAdmin(dynamicClient dynamic.Interface) error {
	ctx := context.Background()

	var err error
	// 1. 检查是否已存在超级管理员
	_, err = dynamicClient.Resource(UserGVR).Get(ctx, SuperAdmin, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return errors.New("查询管理员用户失败")
	}
	logs.Info("📝 管理员不存在，开始初始化...")
	// 2. 创建超级管理员用户 CRD
	adminUser := buildSuperAdmin()

	// 3. 使用动态客户端创建
	if err = CreateUser(adminUser, dynamicClient); err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}

	logs.Info("✅ 默认超级管理员初始化成功")
	return nil
}

func InitNamespace(clientset *kubernetes.Clientset, namespace string) error {
	ctx := context.Background()

	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// 命名空间不存在，创建
			nsSpec := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}

			_, err = clientset.CoreV1().Namespaces().Create(ctx, nsSpec, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("创建命名空间失败: %w", err)
			}
			logs.Info("✅ 创建命名空间: %s", namespace)
			return nil
		}
		return fmt.Errorf("获取命名空间失败: %w", err)
	}

	logs.Info("✅ 命名空间已存在: %s", namespace)
	return nil
}

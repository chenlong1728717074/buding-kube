package kube

import (
	"buding-kube/internal/model"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var (
	InClusterClientSet     *kubernetes.Clientset
	ServerNamespace        string
	DefaultServerNamespace = "kube-buding"
)

func init() {
	config, err := getConfig()
	if err != nil {
		logs.Fatal("初始化k8s配置失败 %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logs.Fatal("获取k8s连接失败")
	}
	if err := initNS(clientSet); err != nil {
		logs.Fatal("初始化命名空间失败")
	}

	if err := initSuperUser(clientSet); err != nil {
		logs.Fatal("初始化超级用户失败")
	}

	InClusterClientSet = clientSet
}

func getConfig() (*rest.Config, error) {
	env := os.Getenv("KUBE_RUNTIME_ENV")
	if env == "KUBE_POD" {
		data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, err
		}
		ServerNamespace = string(data)
		return rest.InClusterConfig()
	}
	ns := os.Getenv("KUBE_RUNTIME_NAMESPACE")
	ServerNamespace = DefaultServerNamespace
	if ns != "" {
		ServerNamespace = ns
	}
	return clientcmd.BuildConfigFromFlags("", "configs/conf.yaml")
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

func initNS(clientSet *kubernetes.Clientset) error {
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

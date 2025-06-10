package test

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "../configs/k3s.yaml")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	list, err := clientset.AppsV1().Deployments("work").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}
	for _, item := range list.Items {
		fmt.Println(item.Name)
	}
}
func TestNode(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "../configs/k3s.yaml")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//节点
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, node := range nodes.Items {
		fmt.Println(node.Name, node.Status, node.Spec)
	}
}
func TestNs(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "../configs/k3s.yaml")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//ns
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, item := range namespaces.Items {
		fmt.Println(item.Name, item.Spec)
		//pods
		list, err := clientset.CoreV1().Pods(item.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, pod := range list.Items {
			fmt.Println(pod.Name, pod.Spec.Containers[0].Image, pod.Status)
		}
		fmt.Println("-----------------------------------------------")
	}
}

func TestDep(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "../configs/k3s.yaml")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//ns

	dep, _ := clientset.AppsV1().Deployments("work").
		Get(context.Background(), "buding-ai", metav1.GetOptions{})
	dep.Spec.Template.Spec.Containers[0].Image = "crpi-5fhtvlj50ip555mt.cn-guangzhou.personal.cr.aliyuncs.com/docker-hubes/buding-ai:v1.1.0.6"
	var num int32 = 1
	dep.Spec.Replicas = &num
	updated, err := clientset.AppsV1().Deployments("work").Update(context.Background(), dep, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("update deployment failed: %v", err)
	}

	fmt.Printf("Deployment %s \n", updated.Name)
}

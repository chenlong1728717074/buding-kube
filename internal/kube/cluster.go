package kube

import (
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils"
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec,omitempty"`
	Status            ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	ApiServer   string            `json:"apiServer"`
	Region      string            `json:"region,omitempty"`
	Alias       string            `json:"alias,omitempty"`
	Description string            `json:"description,omitempty"`
	KubeConfig  string            `json:"kubeConfig,omitempty"`
	Auth        ClusterAuth       `json:"auth,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type ClusterAuth struct {
	Type string `json:"type,omitempty"`
}

type ClusterStatus struct {
	Health      string       `json:"health,omitempty"`
	LastChecked *metav1.Time `json:"lastChecked,omitempty"`
	NodeCount   int64        `json:"nodeCount,omitempty"`
	Message     string       `json:"message,omitempty"`
	State       string       `json:"state,omitempty"`
	Version     string       `json:"version,omitempty"`
}

func CreateCluster(cluster *Cluster, client dynamic.Interface) error {
	ctx := context.Background()
	unstructuredObj, err := utils.ToUnstructured(cluster)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}
	_, err = client.Resource(ClusterGVR).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建集群失败: %w", err)
	}
	logs.Info("✅ 创建集群: %s", cluster.Name)
	return nil
}

func GetCluster(name string) (*Cluster, error) {
	ctx := context.Background()
	unstructuredObj, err := GlobalClient.DynamicClient.Resource(ClusterGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("获取集群失败: %w", err)
	}
	cluster := &Cluster{}
	if err := utils.FromUnstructured(unstructuredObj, cluster); err != nil {
		return nil, fmt.Errorf("转换集群数据失败: %w", err)
	}
	return cluster, nil
}

func ListClusters(opts metav1.ListOptions) ([]Cluster, error) {
	ctx := context.Background()
	list, err := GlobalClient.DynamicClient.Resource(ClusterGVR).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("列出集群失败: %w", err)
	}
	clusters := make([]Cluster, 0, len(list.Items))
	for _, item := range list.Items {
		cluster := Cluster{}
		if err := utils.FromUnstructured(&item, &cluster); err != nil {
			return nil, fmt.Errorf("转换集群数据失败: %w", err)
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func UpdateCluster(cluster *Cluster) error {
	ctx := context.Background()
	existing, err := GlobalClient.DynamicClient.Resource(ClusterGVR).Get(ctx, cluster.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取集群失败: %w", err)
	}
	cluster.ResourceVersion = existing.GetResourceVersion()
	unstructuredObj, err := utils.ToUnstructured(cluster)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}
	_, err = GlobalClient.DynamicClient.Resource(ClusterGVR).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新集群失败: %w", err)
	}
	logs.Info("✅ 更新集群: %s", cluster.Name)
	return nil
}

func DeleteCluster(name string) error {
	ctx := context.Background()
	if err := GlobalClient.DynamicClient.Resource(ClusterGVR).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func BuildCluster(name, alias, describe, kubeConfig, authType, apiServer, version, state string) *Cluster {
	now := metav1.Now()
	if state == "" {
		state = "Active"
	}
	if version == "" {
		version = "unknown"
	}
	if authType == "" {
		authType = "kubeconfig"
	}
	return &Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/%s", consts.GroupCrd, consts.ClusterCrdVersion),
			Kind:       consts.ClusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ClusterSpec{
			ApiServer:   apiServer,
			Alias:       alias,
			Description: describe,
			KubeConfig:  kubeConfig,
			Auth: ClusterAuth{
				Type: authType,
			},
		},
		Status: ClusterStatus{
			Health:      "healthy",
			LastChecked: &now,
			Message:     "连接正常",
			State:       state,
			Version:     version,
		},
	}
}

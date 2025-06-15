package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

var (
	nodeSrv  *NodeService
	nodeOnce sync.Once
)

type NodeService struct {
}

func GetSingletonNodeService() *NodeService {
	nodeOnce.Do(func() {
		nodeSrv = NewNodeService()
	})
	return nodeSrv
}

func NewNodeService() *NodeService {
	return &NodeService{}
}

func (s NodeService) List(query dto.NodeQueryDTO) ([]vo.NodeVO, error) {
	//获取连接
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, err
	}
	list, _ := clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取节点失败 集群:%s 异常:%v", query.ClusterId, err)
		return nil, err
	}
	result := make([]vo.NodeVO, 0)
	for _, item := range list.Items {
		result = append(result, vo.Node2VO(&item))
	}
	return result, nil
}

func (s *NodeService) UnSchedule(query dto.NodeUnScheduleDTO) error {
	node, clientSet, err := s.GetNode(query.NodeInfoQueryDTO)
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = *query.UnSchedule
	_, err = clientSet.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新节点失败: %s %s %s", query.ClusterId, query.Hostname, err.Error())
		return err
	}
	return nil
}

func (s *NodeService) GetNode(query dto.NodeInfoQueryDTO) (*corev1.Node, *kubernetes.Clientset, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, nil, err
	}
	node, err := clientSet.CoreV1().Nodes().Get(context.TODO(), query.Hostname, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取节点失败: %s %s %s", query.ClusterId, query.Hostname, err.Error())
		return nil, nil, err
	}
	return node, clientSet, nil
}

func (s *NodeService) GetNodeInfo(query dto.NodeInfoQueryDTO) (*vo.NodeInfoVO, error) {
	//获取node
	node, clientSet, err := s.GetNode(query)
	if err != nil {
		return nil, err
	}
	//获取事件
	events, err := clientSet.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "involvedObject.kind=Node,involvedObject.name=" + node.Name,
	})
	if err != nil {
		logs.Error("获取节点事件失败: %s %s %v", query.ClusterId, query.Hostname, err)
		return nil, err
	}
	//获取pod
	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})
	if err != nil {
		logs.Error("获取节点事件失败: %s %s %v", query.ClusterId, query.Hostname, err)
		return nil, err
	}
	//metrics-server相关暂时不获取
	return vo.Node2InfoVO(node, events, pods), nil
}

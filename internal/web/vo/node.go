package vo

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
	"time"
)

type NodeVO struct {
	Hostname   string `json:"hostname"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	IP         string `json:"ip"`
	CPU        string `json:"cpu"`
	Memory     string `json:"memory"`
	UnSchedule bool   `json:"unSchedule"`
}

type NodeInfoVO struct {
	Serve       NodeServerInfoVO  `json:"server"`
	Runtime     NodeRuntimeVO     `json:"runtime"`
	Events      []EventVO         `json:"events"`
	Pods        []NodePodVO       `json:"pods"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Metadata    metav1.ObjectMeta `json:"metadata"`
	Spec        corev1.NodeSpec   `json:"spec"`
}

type NodeServerInfoVO struct {
	Hostname         string    `json:"hostname"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	IP               string    `json:"ip"`
	UnSchedule       bool      `json:"unSchedule"`
	OsImage          string    `json:"osImage"`
	OsType           string    `json:"osType"`
	Arch             string    `json:"arch"`
	KernelVersion    string    `json:"kernelVersion"`
	ContainerRuntime string    `json:"containerRuntime"`
	KubeletVersion   string    `json:"kubeletVersion"`
	KubeProxyVersion string    `json:"kubeProxyVersion"`
	CreateTime       time.Time `json:"createTime"`
}

type NodeRuntimeVO struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	EphemeralStorage string `json:"ephemeralStorage"`
	NvidiaGPU        string `json:"nvidiaGpu,omitempty"`
	AMDGPU           string `json:"amdGpu,omitempty"`
	IntelGPU         string `json:"intelGpu,omitempty"`
}

type NodePodVO struct {
	Name        string    `json:"name"`        // Pod名称
	Namespace   string    `json:"namespace"`   // 命名空间
	Image       string    `json:"image"`       // 镜像信息
	Status      string    `json:"status"`      // 运行状态
	Ready       string    `json:"ready"`       // 就绪状态 "1/1"
	Restarts    int32     `json:"restarts"`    // 重启次数
	Age         string    `json:"age"`         // 运行时长
	IP          string    `json:"ip"`          // Pod IP
	Node        string    `json:"node"`        // 节点名称
	Ports       []string  `json:"ports"`       // 端口信息
	CreatedTime time.Time `json:"createdTime"` // 创建时间
}

func buildNodePods(pods *corev1.PodList) []NodePodVO {
	var result []NodePodVO
	if pods == nil {
		return result
	}
	for _, pod := range pods.Items {
		// 跳过已完成或失败的Pod（除非需要显示）
		//if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		//	continue
		//}
		podVO := NodePodVO{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Image:       getMainContainerImage(&pod),
			Status:      string(pod.Status.Phase),
			Ready:       getPodReadyStatus(&pod),
			Restarts:    getPodRestartCount(&pod),
			Age:         time.Since(pod.CreationTimestamp.Time).String(),
			IP:          pod.Status.PodIP,
			Node:        pod.Spec.NodeName,
			Ports:       getPodPorts(&pod),
			CreatedTime: pod.CreationTimestamp.Time,
		}
		result = append(result, podVO)
	}
	// 按创建时间倒序排列
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedTime.After(result[j].CreatedTime)
	})

	return result
}

// 获取主容器镜像
func getMainContainerImage(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) > 0 {
		image := pod.Spec.Containers[0].Image
		// 简化镜像名称显示
		if parts := strings.Split(image, "/"); len(parts) > 1 {
			return parts[len(parts)-1]
		}
		return image
	}
	return ""
}

// 获取Pod就绪状态
func getPodReadyStatus(pod *corev1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Spec.Containers)

	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			readyContainers++
		}
	}

	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// 获取Pod重启次数
func getPodRestartCount(pod *corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

// 获取Pod端口信息
func getPodPorts(pod *corev1.Pod) []string {
	var ports []string

	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			portStr := fmt.Sprintf("%d", port.ContainerPort)
			if port.Protocol != "" && port.Protocol != corev1.ProtocolTCP {
				portStr += "/" + string(port.Protocol)
			}
			if port.Name != "" {
				portStr += " (" + port.Name + ")"
			}
			ports = append(ports, portStr)
		}
	}

	return ports
}

func Node2InfoVO(node *corev1.Node, events *corev1.EventList, pods *corev1.PodList) *NodeInfoVO {
	return &NodeInfoVO{
		Runtime:     buildNodeRuntime(node),
		Serve:       buildNodeServer(node),
		Pods:        buildNodePods(pods),
		Events:      buildNodeEvents(events),
		Metadata:    node.ObjectMeta,
		Annotations: node.Annotations,
		Labels:      node.Labels,
		Spec:        node.Spec,
	}
}

func buildNodeServer(node *corev1.Node) NodeServerInfoVO {
	return NodeServerInfoVO{
		Hostname:         node.Name,
		Role:             GetNodeRole(node),
		Status:           GetNodeStatus(node),
		IP:               GetNodeIp(node),
		UnSchedule:       node.Spec.Unschedulable,
		OsImage:          node.Status.NodeInfo.OSImage,
		OsType:           node.Status.NodeInfo.OperatingSystem,
		Arch:             node.Status.NodeInfo.Architecture,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion: node.Status.NodeInfo.KubeProxyVersion,
		CreateTime:       node.CreationTimestamp.Time,
	}
}

func buildNodeRuntime(node *corev1.Node) NodeRuntimeVO {
	result := NodeRuntimeVO{
		CPU:              node.Status.Allocatable.Cpu().String(),
		Memory:           node.Status.Allocatable.Memory().String(),
		EphemeralStorage: node.Status.Allocatable.StorageEphemeral().String(),
	}
	if gpu, exists := node.Status.Allocatable["nvidia.com/gpu"]; exists {
		result.NvidiaGPU = gpu.String()
	}
	if gpu, exists := node.Status.Allocatable["amd.com/gpu"]; exists {
		result.AMDGPU = gpu.String()
	}
	if gpu, exists := node.Status.Allocatable["gpu.intel.com/i915"]; exists {
		result.IntelGPU = gpu.String()
	}
	return result
}

func Node2VO(node *corev1.Node) NodeVO {
	return NodeVO{
		Hostname:   node.Name,
		Role:       GetNodeRole(node),
		Status:     GetNodeStatus(node),
		IP:         GetNodeIp(node),
		CPU:        node.Status.Allocatable.Cpu().String(),
		Memory:     node.Status.Allocatable.Memory().String(),
		UnSchedule: node.Spec.Unschedulable,
	}
}

func GetNodeIp(node *corev1.Node) string {
	ip := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip = addr.Address
			break
		}
	}
	return ip
}

func GetNodeStatus(node *corev1.Node) string {
	status := "Unknown"
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}
	return status
}

func GetNodeRole(node *corev1.Node) string {
	role := "worker"
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		role = "master"
	} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		role = "control-plane"
	}
	return role
}

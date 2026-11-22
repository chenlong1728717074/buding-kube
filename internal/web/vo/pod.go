package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// PodVO Pod 列表项. 不包含 YAML, YAML 由详情接口按需加载
type PodVO struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Namespace         string    `json:"namespace"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	NodeName          string    `json:"nodeName"`
	HostIP            string    `json:"hostIP"`
	PodIP             string    `json:"podIP"`
}

func Pod2VO(item *corev1.Pod) PodVO {
	return PodVO{
		CreationTimestamp: item.CreationTimestamp.Time,
		Name:              item.Name,
		Namespace:         item.Namespace,
		Status:            string(item.Status.Phase),
		HostIP:            item.Status.HostIP,
		PodIP:             item.Status.PodIP,
		NodeName:          item.Spec.NodeName,
	}
}

type PodInfoVO struct {
	BaseVO
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Status            string            `json:"status"`
	NodeName          string            `json:"nodeName"`
	HostIP            string            `json:"hostIP"`
	PodIP             string            `json:"podIP"`
	Yaml              string            `json:"yaml"`
	Containers        []PodContainerVO  `json:"containers"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	MetaData          metav1.ObjectMeta `json:"metaData"`
	Spec              corev1.PodSpec    `json:"spec"`
	Events            []EventVO         `json:"events"`
}

type PodContainerVO struct {
	Name         string               `json:"name"`
	Ready        bool                 `json:"ready"`
	RestartCount int32                `json:"restartCount"`
	Image        string               `json:"image"`
	Started      bool                 `json:"started"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts"`
	StartedAt    time.Time            `json:"startedAt"`
	Args         []string             `json:"args"`
	Envs         []PodEnvVO           `json:"envs"`
	Ports        []PodPortVO          `json:"ports"`
}

type PodEnvVO struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PodPortVO struct {
	Name          string `json:"name"`
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	HostIP        string `json:"hostIP"`
	Protocol      string `json:"protocol"`
}

func Pod2InfoVO(item *corev1.Pod, events *corev1.EventList) PodInfoVO {
	return PodInfoVO{
		CreationTimestamp: item.CreationTimestamp.Time,
		BaseVO: BaseVO{
			Name:      item.Name,
			Namespace: item.Namespace,
		},
		Status:      string(item.Status.Phase),
		HostIP:      item.Status.HostIP,
		PodIP:       item.Status.PodIP,
		NodeName:    item.Spec.NodeName,
		Containers:  buildPodContainer(item),
		Annotations: item.Annotations,
		Labels:      item.Labels,
		MetaData:    item.ObjectMeta,
		Spec:        item.Spec,
		Events:      buildNodeEvents(events),
	}
}

// buildPodContainer 构建容器信息.
// 按容器名匹配状态(ContainerStatuses 与 Spec.Containers 的顺序不保证一致),
// 并兼容 Pending 阶段还没有状态、Started 为 nil 的情况, 避免空指针 panic
func buildPodContainer(item *corev1.Pod) []PodContainerVO {
	statusMap := make(map[string]corev1.ContainerStatus, len(item.Status.ContainerStatuses))
	for _, status := range item.Status.ContainerStatuses {
		statusMap[status.Name] = status
	}
	result := make([]PodContainerVO, 0, len(item.Spec.Containers))
	for _, container := range item.Spec.Containers {
		c := PodContainerVO{
			Name:         container.Name,
			Image:        container.Image,
			VolumeMounts: container.VolumeMounts,
			Args:         container.Args,
			Envs:         buildPodEnvVO(container.Env),
			Ports:        buildPodPorts(container.Ports),
		}
		if status, ok := statusMap[container.Name]; ok {
			c.Ready = status.Ready
			c.RestartCount = status.RestartCount
			c.Started = status.Started != nil && *status.Started
			if status.State.Running != nil {
				c.StartedAt = status.State.Running.StartedAt.Time
			}
		}
		result = append(result, c)
	}
	return result
}

func buildPodPorts(ports []corev1.ContainerPort) []PodPortVO {
	result := make([]PodPortVO, 0)
	for _, item := range ports {
		result = append(result, PodPortVO{
			Name:          item.Name,
			HostPort:      item.HostPort,
			ContainerPort: item.ContainerPort,
			HostIP:        item.HostIP,
			Protocol:      string(item.Protocol),
		})
	}
	return result
}

func buildPodEnvVO(env []corev1.EnvVar) []PodEnvVO {
	result := make([]PodEnvVO, 0)
	for _, envVar := range env {
		result = append(result, PodEnvVO{
			Name:  envVar.Name,
			Value: envVar.Value,
		})
	}
	return result
}

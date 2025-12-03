package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type PodVO struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Namespace         string    `json:"namespace"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	NodeName          string    `json:"nodeName"`
	HostIP            string    `json:"hostIP"`
	PodIP             string    `json:"podIP"`
	Yaml              string    `json:"yaml"`
}

func Pod2VO(item corev1.Pod, yaml string) PodVO {
	return PodVO{
		CreationTimestamp: item.CreationTimestamp.Time,
		Name:              item.Name,
		Namespace:         item.Namespace,
		Status:            string(item.Status.Phase),
		HostIP:            item.Status.HostIP,
		PodIP:             item.Status.PodIP,
		NodeName:          item.Spec.NodeName,
		Yaml:              yaml,
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

func buildPodContainer(item *corev1.Pod) []PodContainerVO {
	result := make([]PodContainerVO, 0)
	containers := item.Spec.Containers
	for index, status := range item.Status.ContainerStatuses {
		container := containers[index]
		vo := PodContainerVO{
			Name:         container.Name,
			Ready:        status.Ready,
			RestartCount: status.RestartCount,
			Image:        container.Image,
			Started:      *status.Started,
			VolumeMounts: container.VolumeMounts,
			Args:         container.Args,
			Envs:         buildPodEnvVO(container.Env),
			Ports:        buildPodPorts(container.Ports),
		}
		if status.State.Running != nil {
			vo.StartedAt = status.State.Running.StartedAt.Time
		}
		result = append(result, vo)
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

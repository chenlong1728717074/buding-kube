package vo

import (
	corev1 "k8s.io/api/core/v1"
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
}

func Pod2VO(item corev1.Pod) PodVO {
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
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Namespace         string    `json:"namespace"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	NodeName          string    `json:"nodeName"`
	HostIP            string    `json:"hostIP"`
	PodIP             string    `json:"podIP"`
}

package vo

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type NamespaceVO struct {
	Name              string               `json:"name"`
	Alias             string               `json:"alias"`
	Describe          string               `json:"describe"`
	Generation        int64                `json:"generation"`
	Version           string               `json:"version"`
	Status            string               `json:"active"`
	Uid               string               `json:"uid"`
	ResourceVersion   string               `json:"resourceVersion"`
	Annotations       map[string]string    `json:"annotations"`
	Labels            map[string]string    `json:"labels"`
	CreationTimestamp time.Time            `json:"creationTimestamp"`
	Metadata          metav1.ObjectMeta    `json:"metadata"`
	Spec              corev1.NamespaceSpec `json:"spec"`
}

func Namespace2VO(ns *corev1.Namespace) NamespaceVO {
	return NamespaceVO{
		Name:              ns.Name,
		Alias:             ns.Annotations["alias"],
		Describe:          ns.Annotations["describe"],
		Status:            string(ns.Status.Phase),
		Annotations:       ns.Annotations,
		CreationTimestamp: ns.CreationTimestamp.Time,
		ResourceVersion:   ns.ResourceVersion,
		Version:           ns.APIVersion,
		Uid:               string(ns.UID),
		Labels:            ns.Labels,
		Metadata:          ns.ObjectMeta,
		Spec:              ns.Spec,
	}
}

package vo

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

type ServiceVO struct {
	Name           string    `json:"name"`
	Namespace      string    `json:"namespace"`
	Alias          string    `json:"alias"`
	Describe       string    `json:"describe"`
	Yaml           string    `json:"yaml"`
	Status         string    `json:"status"`
	CreateTime     time.Time `json:"createTime"`
	InternalIP     string    `json:"internalIP"`     // 内部IP (ClusterIP)
	ExternalIP     string    `json:"externalIP"`     // 外部IP
	InternalAccess string    `json:"internalAccess"` // 内部访问域名
	ExternalAccess string    `json:"externalAccess"` // 外部访问地址
	AccessMode     string    `json:"accessMode"`     // 访问模式
	ServiceType    string    `json:"serviceType"`    // Service类型
}

// 获取内部IP (ClusterIP)
func getInternalIP(svc corev1.Service) string {
	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return "-"
	}
	return svc.Spec.ClusterIP
}

// 获取外部IP
func getExternalIP(svc corev1.Service) string {
	// 优先检查 ExternalIPs
	if len(svc.Spec.ExternalIPs) > 0 {
		return strings.Join(svc.Spec.ExternalIPs, ",")
	}

	// 检查 LoadBalancer Ingress
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		if svc.Status.LoadBalancer.Ingress[0].IP != "" {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			return svc.Status.LoadBalancer.Ingress[0].Hostname
		}
	}

	return "-"
}

// 获取内部访问地址 (域名)
func getInternalAccess(svc corev1.Service) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
}

// 获取外部访问地址
func getExternalAccess(svc corev1.Service) string {
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				return svc.Status.LoadBalancer.Ingress[0].IP
			}
			if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				return svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		}
		return "Pending"

	case corev1.ServiceTypeNodePort:
		if len(svc.Spec.Ports) > 0 {
			return fmt.Sprintf("<NodeIP>:%d", svc.Spec.Ports[0].NodePort)
		}
		return "-"

	case corev1.ServiceTypeExternalName:
		return svc.Spec.ExternalName

	default:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "-"
	}
}

// 获取访问模式
func getAccessMode(svc corev1.Service) string {
	if svc.Spec.ClusterIP == "None" {
		return "域名" // Headless Service
	}
	return "VirtualIP"
}

// 获取服务类型
func getServiceType(svc corev1.Service) string {
	return string(svc.Spec.Type)
}

// 获取服务状态
func getServiceStatus(svc corev1.Service) string {
	// 根据实际需求定义状态逻辑
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return "Ready"
		}
		return "Pending"
	}
	return "Ready"
}

func Service2VO(item corev1.Service) ServiceVO {
	return ServiceVO{
		Name:           item.Name,
		Namespace:      item.Namespace,
		Alias:          item.Annotations["alias"],
		Describe:       item.Annotations["describe"],
		Status:         getServiceStatus(item),
		CreateTime:     item.CreationTimestamp.Time,
		InternalIP:     getInternalIP(item),
		ExternalIP:     getExternalIP(item),
		InternalAccess: getInternalAccess(item),
		ExternalAccess: getExternalAccess(item),
		AccessMode:     getAccessMode(item),
		ServiceType:    getServiceType(item),
	}
}

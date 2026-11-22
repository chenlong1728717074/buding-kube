package vo

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

type ServiceVO struct {
	BaseVO
	Status         string            `json:"status"`
	CreateTime     time.Time         `json:"createTime"`
	InternalIP     string            `json:"internalIP"`     // 内部IP (ClusterIP)
	ExternalIP     string            `json:"externalIP"`     // 外部IP
	InternalAccess string            `json:"internalAccess"` // 内部访问域名
	ExternalAccess string            `json:"externalAccess"` // 外部访问地址
	AccessMode     string            `json:"accessMode"`     // 访问模式
	ServiceType    string            `json:"serviceType"`    // Service类型
	Headless       bool              `json:"headless"`
	Ports          []ServicePortVO   `json:"ports"`
	Selector       map[string]string `json:"selector"`
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
		BaseVO: BaseVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Alias:     item.Annotations["alias"],
			Describe:  item.Annotations["describe"],
		},
		Status:         getServiceStatus(item),
		CreateTime:     item.CreationTimestamp.Time,
		InternalIP:     getInternalIP(item),
		ExternalIP:     getExternalIP(item),
		InternalAccess: getInternalAccess(item),
		ExternalAccess: getExternalAccess(item),
		AccessMode:     getAccessMode(item),
		ServiceType:    getServiceType(item),
		Headless:       item.Spec.ClusterIP == corev1.ClusterIPNone,
		Ports:          servicePorts2VO(item.Spec.Ports),
		Selector:       item.Spec.Selector,
	}
}

// ServicePortVO 服务端口
type ServicePortVO struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort"`
}

// ServiceEndpointVO 服务的一个后端地址
type ServiceEndpointVO struct {
	IP       string `json:"ip"`
	Ready    bool   `json:"ready"`
	PodName  string `json:"podName"`
	NodeName string `json:"nodeName"`
	Ports    string `json:"ports"`
}

// ServiceWorkloadVO 被服务选中的工作负载
type ServiceWorkloadVO struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Replicas      int32  `json:"replicas"`
	ReadyReplicas int32  `json:"readyReplicas"`
}

// ServiceDetailVO 服务详情
type ServiceDetailVO struct {
	ServiceVO
	ClusterIP              string              `json:"clusterIP"`
	ClusterIPs             []string            `json:"clusterIPs"`
	Headless               bool                `json:"headless"`
	ExternalIPs            []string            `json:"externalIPs"`
	SessionAffinity        string              `json:"sessionAffinity"`
	SessionAffinityTimeout *int32              `json:"sessionAffinityTimeout"`
	ExternalName           string              `json:"externalName"`
	ExternalTrafficPolicy  string              `json:"externalTrafficPolicy"`
	LoadBalancerIngress    []string            `json:"loadBalancerIngress"`
	IPFamilies             []string            `json:"ipFamilies"`
	Creator                string              `json:"creator"`
	Uid                    string              `json:"uid"`
	Labels                 map[string]string   `json:"labels"`
	Annotations            map[string]string   `json:"annotations"`
	Endpoints              []ServiceEndpointVO `json:"endpoints"`
	Pods                   []JobPodVO          `json:"pods"`
	Workloads              []ServiceWorkloadVO `json:"workloads"`
	Ingresses              []string            `json:"ingresses"`
	Events                 []EventVO           `json:"events"`
}

func servicePorts2VO(ports []corev1.ServicePort) []ServicePortVO {
	result := make([]ServicePortVO, 0, len(ports))
	for _, p := range ports {
		protocol := string(p.Protocol)
		if protocol == "" {
			protocol = "TCP"
		}
		target := p.TargetPort.String()
		if target == "" || target == "0" {
			target = fmt.Sprint(p.Port)
		}
		result = append(result, ServicePortVO{Name: p.Name, Protocol: protocol, Port: p.Port, TargetPort: target, NodePort: p.NodePort})
	}
	return result
}

func Service2DetailVO(svc *corev1.Service) ServiceDetailVO {
	result := ServiceDetailVO{
		ServiceVO:             Service2VO(*svc),
		ClusterIP:             svc.Spec.ClusterIP,
		ClusterIPs:            svc.Spec.ClusterIPs,
		Headless:              svc.Spec.ClusterIP == corev1.ClusterIPNone,
		ExternalIPs:           svc.Spec.ExternalIPs,
		SessionAffinity:       string(svc.Spec.SessionAffinity),
		ExternalName:          svc.Spec.ExternalName,
		ExternalTrafficPolicy: string(svc.Spec.ExternalTrafficPolicy),
		LoadBalancerIngress:   []string{},
		IPFamilies:            []string{},
		Creator:               ResourceCreator(svc.Annotations),
		Uid:                   string(svc.UID),
		Labels:                svc.Labels,
		Annotations:           svc.Annotations,
		Endpoints:             []ServiceEndpointVO{},
		Pods:                  []JobPodVO{},
		Workloads:             []ServiceWorkloadVO{},
		Ingresses:             []string{},
		Events:                []EventVO{},
	}
	if cfg := svc.Spec.SessionAffinityConfig; cfg != nil && cfg.ClientIP != nil {
		result.SessionAffinityTimeout = cfg.ClientIP.TimeoutSeconds
	}
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			result.LoadBalancerIngress = append(result.LoadBalancerIngress, ing.IP)
		} else if ing.Hostname != "" {
			result.LoadBalancerIngress = append(result.LoadBalancerIngress, ing.Hostname)
		}
	}
	for _, f := range svc.Spec.IPFamilies {
		result.IPFamilies = append(result.IPFamilies, string(f))
	}
	return result
}

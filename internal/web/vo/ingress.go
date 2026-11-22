package vo

import (
	"fmt"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
)

type IngressVO struct {
	BaseVO
	IngressClassName *string   `json:"ingressClassName"` // Ingress类名称，指定使用的Ingress控制器
	Hosts            []string  `json:"hosts"`            // 规则中的域名
	Addresses        []string  `json:"addresses"`        // 控制器分配的访问地址
	TLS              bool      `json:"tls"`              // 是否启用 HTTPS
	RuleCount        int       `json:"ruleCount"`        // 转发路径数量
	CreateTime       time.Time `json:"createTime"`       // 资源创建时间
}

// IngressRuleVO 展开后的一条转发规则
type IngressRuleVO struct {
	Host          string `json:"host"`
	Path          string `json:"path"`
	PathType      string `json:"pathType"`
	ServiceName   string `json:"serviceName"`
	ServicePort   string `json:"servicePort"`
	ServiceExists bool   `json:"serviceExists"`
}

type IngressTLSVO struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressDetailVO 应用路由详情
type IngressDetailVO struct {
	IngressVO
	Rules          []IngressRuleVO   `json:"rules"`
	TLSList        []IngressTLSVO    `json:"tlsList"`
	DefaultBackend string            `json:"defaultBackend"`
	Creator        string            `json:"creator"`
	Uid            string            `json:"uid"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	Events         []EventVO         `json:"events"`
}

// IngressClassVO 集群中的 Ingress 类
type IngressClassVO struct {
	Name       string `json:"name"`
	Controller string `json:"controller"`
	IsDefault  bool   `json:"isDefault"`
}

func Ingress2VO(ing networkingv1.Ingress) IngressVO {
	result := IngressVO{
		BaseVO: BaseVO{
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Alias:     ing.Annotations["alias"],
			Describe:  ing.Annotations["describe"],
		},
		IngressClassName: ing.Spec.IngressClassName,
		Hosts:            []string{},
		Addresses:        []string{},
		TLS:              len(ing.Spec.TLS) > 0,
		CreateTime:       ing.CreationTimestamp.Time,
	}
	if result.IngressClassName == nil {
		// 旧版本使用注解指定类
		if c := ing.Annotations["kubernetes.io/ingress.class"]; c != "" {
			result.IngressClassName = &c
		}
	}
	seen := map[string]bool{}
	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		if host == "" {
			host = "*"
		}
		if !seen[host] {
			seen[host] = true
			result.Hosts = append(result.Hosts, host)
		}
		if rule.HTTP != nil {
			result.RuleCount += len(rule.HTTP.Paths)
		}
	}
	for _, lb := range ing.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			result.Addresses = append(result.Addresses, lb.IP)
		} else if lb.Hostname != "" {
			result.Addresses = append(result.Addresses, lb.Hostname)
		}
	}
	return result
}

// IngressBackendText 后端服务和端口
func IngressBackendText(b *networkingv1.IngressBackend) (string, string) {
	if b == nil {
		return "", ""
	}
	if b.Service != nil {
		port := b.Service.Port.Name
		if port == "" {
			port = fmt.Sprint(b.Service.Port.Number)
		}
		return b.Service.Name, port
	}
	if b.Resource != nil {
		return b.Resource.Kind + "/" + b.Resource.Name, ""
	}
	return "", ""
}

// Ingress2DetailVO serviceNames 为命名空间中已有的服务, 用于标记后端服务是否存在
func Ingress2DetailVO(ing *networkingv1.Ingress, serviceNames map[string]bool) IngressDetailVO {
	result := IngressDetailVO{
		IngressVO:   Ingress2VO(*ing),
		Rules:       []IngressRuleVO{},
		TLSList:     []IngressTLSVO{},
		Creator:     ResourceCreator(ing.Annotations),
		Uid:         string(ing.UID),
		Labels:      ing.Labels,
		Annotations: ing.Annotations,
		Events:      []EventVO{},
	}
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for i := range rule.HTTP.Paths {
			p := &rule.HTTP.Paths[i]
			svc, port := IngressBackendText(&p.Backend)
			item := IngressRuleVO{Host: rule.Host, Path: p.Path, ServiceName: svc, ServicePort: port, ServiceExists: serviceNames[svc]}
			if p.PathType != nil {
				item.PathType = string(*p.PathType)
			}
			result.Rules = append(result.Rules, item)
		}
	}
	for _, t := range ing.Spec.TLS {
		hosts := t.Hosts
		if hosts == nil {
			hosts = []string{}
		}
		result.TLSList = append(result.TLSList, IngressTLSVO{Hosts: hosts, SecretName: t.SecretName})
	}
	if svc, port := IngressBackendText(ing.Spec.DefaultBackend); svc != "" {
		result.DefaultBackend = svc
		if port != "" {
			result.DefaultBackend += ":" + port
		}
	}
	return result
}

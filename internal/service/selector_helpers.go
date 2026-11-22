package service

import (
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// listPodsBySelector 按标签选择器查询容器组; 选择器为空时返回空列表(空选择器会匹配所有容器组)
func listPodsBySelector(ctx context.Context, cs *kubernetes.Clientset, namespace string, selector *metav1.LabelSelector) ([]vo.JobPodVO, error) {
	if selector == nil {
		return []vo.JobPodVO{}, nil
	}
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}
	if sel.Empty() {
		return []vo.JobPodVO{}, nil
	}
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: sel.String(), ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	result := make([]vo.JobPodVO, 0, len(pods.Items))
	for i := range pods.Items {
		result = append(result, vo.Pod2JobPodVO(&pods.Items[i]))
	}
	sortJobPods(result)
	attachJobPodMetrics(ctx, cs, namespace, result)
	return result, nil
}

// selectorMatches 服务选择器是否选中这组容器组标签
func selectorMatches(selector, podLabels map[string]string) bool {
	if len(selector) == 0 || len(podLabels) == 0 {
		return false
	}
	return labels.SelectorFromSet(labels.Set(selector)).Matches(labels.Set(podLabels))
}

// servicesSelecting 命名空间中选中这组容器组标签的服务
func servicesSelecting(ctx context.Context, cs *kubernetes.Clientset, namespace string, podLabels map[string]string) []vo.WorkloadServiceRefVO {
	result := make([]vo.WorkloadServiceRefVO, 0)
	if len(podLabels) == 0 {
		return result
	}
	svcs, err := cs.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Warn("获取服务列表失败: %v", err)
		return result
	}
	for i := range svcs.Items {
		if selectorMatches(svcs.Items[i].Spec.Selector, podLabels) {
			result = append(result, vo.Service2RefVO(&svcs.Items[i]))
		}
	}
	return result
}

// workloadsSelectedBy 命名空间中被服务选择器选中的工作负载
func workloadsSelectedBy(ctx context.Context, cs *kubernetes.Clientset, namespace string, selector map[string]string) []vo.ServiceWorkloadVO {
	result := make([]vo.ServiceWorkloadVO, 0)
	if len(selector) == 0 {
		return result
	}
	opts := metav1.ListOptions{ResourceVersion: "0"}
	if items, err := cs.AppsV1().Deployments(namespace).List(ctx, opts); err == nil {
		for i := range items.Items {
			d := &items.Items[i]
			if selectorMatches(selector, d.Spec.Template.Labels) {
				result = append(result, vo.ServiceWorkloadVO{Kind: WorkloadDeployment, Name: d.Name, Replicas: d.Status.Replicas, ReadyReplicas: d.Status.ReadyReplicas})
			}
		}
	}
	if items, err := cs.AppsV1().StatefulSets(namespace).List(ctx, opts); err == nil {
		for i := range items.Items {
			s := &items.Items[i]
			if selectorMatches(selector, s.Spec.Template.Labels) {
				result = append(result, vo.ServiceWorkloadVO{Kind: WorkloadStatefulSet, Name: s.Name, Replicas: s.Status.Replicas, ReadyReplicas: s.Status.ReadyReplicas})
			}
		}
	}
	if items, err := cs.AppsV1().DaemonSets(namespace).List(ctx, opts); err == nil {
		for i := range items.Items {
			d := &items.Items[i]
			if selectorMatches(selector, d.Spec.Template.Labels) {
				result = append(result, vo.ServiceWorkloadVO{Kind: WorkloadDaemonSet, Name: d.Name, Replicas: d.Status.DesiredNumberScheduled, ReadyReplicas: d.Status.NumberReady})
			}
		}
	}
	return result
}

// serviceEndpoints 服务的后端地址(来自 EndpointSlice)
func serviceEndpoints(ctx context.Context, cs *kubernetes.Clientset, svc *corev1.Service) []vo.ServiceEndpointVO {
	result := make([]vo.ServiceEndpointVO, 0)
	slices, err := cs.DiscoveryV1().EndpointSlices(svc.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: discoveryv1.LabelServiceName + "=" + svc.Name,
	})
	if err != nil {
		logs.Warn("获取服务 %s/%s 的 EndpointSlice 失败: %v", svc.Namespace, svc.Name, err)
		return result
	}
	for i := range slices.Items {
		slice := &slices.Items[i]
		ports := make([]string, 0, len(slice.Ports))
		for _, p := range slice.Ports {
			text := ""
			if p.Port != nil {
				text = fmt.Sprint(*p.Port)
			}
			if p.Protocol != nil {
				text += "/" + string(*p.Protocol)
			}
			if p.Name != nil && *p.Name != "" {
				text = *p.Name + " " + text
			}
			ports = append(ports, text)
		}
		portText := strings.Join(ports, "，")
		for _, ep := range slice.Endpoints {
			ready := ep.Conditions.Ready == nil || *ep.Conditions.Ready
			podName, nodeName := "", ""
			if ep.TargetRef != nil && ep.TargetRef.Kind == "Pod" {
				podName = ep.TargetRef.Name
			}
			if ep.NodeName != nil {
				nodeName = *ep.NodeName
			}
			for _, addr := range ep.Addresses {
				result = append(result, vo.ServiceEndpointVO{IP: addr, Ready: ready, PodName: podName, NodeName: nodeName, Ports: portText})
			}
		}
	}
	return result
}

// ingressesForService 使用该服务作为后端的应用路由
func ingressesForService(ctx context.Context, cs *kubernetes.Clientset, namespace, name string) []string {
	result := make([]string, 0)
	items, err := cs.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Warn("获取应用路由失败: %v", err)
		return result
	}
	for i := range items.Items {
		ing := &items.Items[i]
		used := ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil && ing.Spec.DefaultBackend.Service.Name == name
		for _, rule := range ing.Spec.Rules {
			if used || rule.HTTP == nil {
				continue
			}
			for _, p := range rule.HTTP.Paths {
				if p.Backend.Service != nil && p.Backend.Service.Name == name {
					used = true
					break
				}
			}
		}
		if used {
			result = append(result, ing.Name)
		}
	}
	return result
}

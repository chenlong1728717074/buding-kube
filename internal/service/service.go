package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	kubeSrv     *KubeSrvService
	kubeSrvOnce sync.Once
)

type KubeSrvService struct {
}

func NewKubeSrvService() *KubeSrvService {
	return &KubeSrvService{}
}

func GetSingletonKubeSrvService() *KubeSrvService {
	kubeSrvOnce.Do(func() {
		kubeSrv = NewKubeSrvService()
	})
	return kubeSrv
}

func (s *KubeSrvService) List(query dto.ServiceQueryDTO) ([]vo.ServiceVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.CoreV1().Services(query.Namespace).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取service失败: %v", err)
		return nil, fmt.Errorf("获取service失败: %v", err)
	}
	keyword := strings.TrimSpace(query.Keyword)
	svcType := strings.TrimSpace(query.Type)
	result := make([]vo.ServiceVO, 0, len(items.Items))
	for _, item := range items.Items {
		if keyword != "" && !strings.Contains(item.Name, keyword) {
			continue
		}
		vi := vo.Service2VO(item)
		if svcType == "Headless" && !vi.Headless {
			continue
		}
		if svcType != "" && svcType != "Headless" && vi.ServiceType != svcType {
			continue
		}
		result = append(result, vi)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// Get 服务详情: 端口、后端地址、容器组、工作负载、应用路由和事件
func (s *KubeSrvService) Get(ctx context.Context, query dto.ServiceBaseDTO) (*vo.ServiceDetailVO, error) {
	cs, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return nil, err
	}
	svc, err := cs.CoreV1().Services(query.Namespace).Get(ctx, query.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("服务 %s 不存在", query.Name)
		}
		return nil, fmt.Errorf("获取服务失败: %v", err)
	}
	detail := vo.Service2DetailVO(svc)
	namespace := svc.Namespace
	detail.Endpoints = serviceEndpoints(ctx, cs, svc)
	if len(svc.Spec.Selector) > 0 {
		if pods, podErr := listPodsBySelector(ctx, cs, namespace, &metav1.LabelSelector{MatchLabels: svc.Spec.Selector}); podErr == nil {
			detail.Pods = pods
		} else {
			logs.Warn("获取服务 %s/%s 的容器组失败: %v", namespace, svc.Name, podErr)
		}
		detail.Workloads = workloadsSelectedBy(ctx, cs, namespace, svc.Spec.Selector)
	}
	detail.Ingresses = ingressesForService(ctx, cs, namespace, svc.Name)
	detail.Events = listObjectEvents(ctx, cs, namespace, string(svc.UID))
	return &detail, nil
}

// buildServicePorts 端口; 有多个端口时自动补全名称(Kubernetes 要求多端口必须命名)
func buildServicePorts(items []dto.ServicePortDTO, svcType corev1.ServiceType) ([]corev1.ServicePort, error) {
	result := make([]corev1.ServicePort, 0, len(items))
	seen := map[string]bool{}
	names := map[string]bool{}
	for i, p := range items {
		label := fmt.Sprintf("第 %d 个端口", i+1)
		name := strings.TrimSpace(p.Name)
		target := strings.TrimSpace(p.TargetPort)
		if p.Port == 0 && name == "" && target == "" {
			continue
		}
		if p.Port < 1 || p.Port > 65535 {
			return nil, fmt.Errorf("%s的服务端口应在 1 到 65535 之间", label)
		}
		protocol, err := parseProtocol(p.Protocol)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", label, err)
		}
		key := fmt.Sprintf("%d/%s", p.Port, protocol)
		if seen[key] {
			return nil, fmt.Errorf("端口 %s 重复", key)
		}
		seen[key] = true
		sp := corev1.ServicePort{Port: p.Port, Protocol: protocol}
		if name != "" {
			if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
				return nil, fmt.Errorf("%s的名称 %s 不合法: 只能包含小写字母、数字和 \"-\"", label, name)
			}
			if names[name] {
				return nil, fmt.Errorf("端口名称 %s 重复", name)
			}
			names[name] = true
			sp.Name = name
		}
		switch n, convErr := strconv.Atoi(target); {
		case target == "":
			sp.TargetPort = intstr.FromInt32(p.Port)
		case convErr == nil:
			if n < 1 || n > 65535 {
				return nil, fmt.Errorf("%s的容器端口应在 1 到 65535 之间", label)
			}
			sp.TargetPort = intstr.FromInt32(int32(n))
		default:
			if errs := validation.IsValidPortName(target); len(errs) > 0 {
				return nil, fmt.Errorf("%s的容器端口 %s 不是有效的端口号或端口名称", label, target)
			}
			sp.TargetPort = intstr.FromString(target)
		}
		if p.NodePort != 0 {
			if svcType != corev1.ServiceTypeNodePort && svcType != corev1.ServiceTypeLoadBalancer {
				return nil, errors.New("只有 NodePort / LoadBalancer 类型可以指定节点端口")
			}
			if p.NodePort < 1 || p.NodePort > 65535 {
				return nil, fmt.Errorf("%s的节点端口不正确", label)
			}
			sp.NodePort = p.NodePort
		}
		result = append(result, sp)
	}
	if len(result) > 1 {
		for i := range result {
			if result[i].Name != "" {
				continue
			}
			name := fmt.Sprintf("%s-%d", strings.ToLower(string(result[i].Protocol)), result[i].Port)
			if names[name] {
				return nil, errors.New("有多个端口时，请为每个端口填写不同的名称")
			}
			names[name] = true
			result[i].Name = name
		}
	}
	return result, nil
}

func buildServiceSpec(req dto.ServiceFormDTO) (corev1.ServiceSpec, error) {
	spec := corev1.ServiceSpec{}
	svcType := corev1.ServiceType(strings.TrimSpace(req.Type))
	if svcType == "" {
		svcType = corev1.ServiceTypeClusterIP
	}
	switch svcType {
	case corev1.ServiceTypeClusterIP, corev1.ServiceTypeNodePort, corev1.ServiceTypeLoadBalancer, corev1.ServiceTypeExternalName:
	default:
		return spec, fmt.Errorf("不支持的服务类型: %s", req.Type)
	}
	spec.Type = svcType
	if svcType == corev1.ServiceTypeExternalName {
		name := strings.TrimSpace(req.ExternalName)
		if name == "" || len(validation.IsDNS1123Subdomain(name)) > 0 {
			return spec, errors.New("请填写正确的外部服务域名，例如 db.example.com")
		}
		spec.ExternalName = name
	}
	if req.Headless {
		if svcType != corev1.ServiceTypeClusterIP {
			return spec, errors.New("内部域名（Headless）方式只能使用 ClusterIP 类型")
		}
		spec.ClusterIP = corev1.ClusterIPNone
	}
	selector := cleanStringMap(req.Selector)
	if err := validateLabelMap(selector, "选择器"); err != nil {
		return spec, err
	}
	if svcType != corev1.ServiceTypeExternalName {
		spec.Selector = selector
	}
	ports, err := buildServicePorts(req.Ports, svcType)
	if err != nil {
		return spec, err
	}
	if len(ports) == 0 && !req.Headless && svcType != corev1.ServiceTypeExternalName {
		return spec, errors.New("请至少添加一个端口")
	}
	spec.Ports = ports
	switch strings.TrimSpace(req.SessionAffinity) {
	case "", string(corev1.ServiceAffinityNone):
		spec.SessionAffinity = corev1.ServiceAffinityNone
	case string(corev1.ServiceAffinityClientIP):
		if svcType == corev1.ServiceTypeExternalName {
			return spec, errors.New("外部服务不支持会话保持")
		}
		timeout := int32(10800)
		if req.SessionAffinityTimeout > 0 {
			if req.SessionAffinityTimeout > 86400 {
				return spec, errors.New("会话保持时间不能超过 86400 秒")
			}
			timeout = req.SessionAffinityTimeout
		}
		spec.SessionAffinity = corev1.ServiceAffinityClientIP
		spec.SessionAffinityConfig = &corev1.SessionAffinityConfig{ClientIP: &corev1.ClientIPConfig{TimeoutSeconds: &timeout}}
	default:
		return spec, fmt.Errorf("不支持的会话保持方式: %s", req.SessionAffinity)
	}
	if svcType == corev1.ServiceTypeNodePort || svcType == corev1.ServiceTypeLoadBalancer {
		switch policy := corev1.ServiceExternalTrafficPolicy(strings.TrimSpace(req.ExternalTrafficPolicy)); policy {
		case "":
			spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyCluster
		case corev1.ServiceExternalTrafficPolicyCluster, corev1.ServiceExternalTrafficPolicyLocal:
			spec.ExternalTrafficPolicy = policy
		default:
			return spec, fmt.Errorf("不支持的外部流量策略: %s", req.ExternalTrafficPolicy)
		}
	}
	return spec, nil
}

func validateServiceName(name string) error {
	if errs := validation.IsDNS1035Label(name); len(errs) > 0 {
		return errors.New("服务名称不合法: 需要以小写字母开头，只能包含小写字母、数字和 \"-\"，最多 63 个字符")
	}
	return nil
}

// Create 创建服务
func (s *KubeSrvService) Create(ctx context.Context, req dto.ServiceFormDTO, creator string) error {
	name := strings.TrimSpace(req.Name)
	if err := validateServiceName(name); err != nil {
		return err
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	spec, err := buildServiceSpec(req)
	if err != nil {
		return err
	}
	labels := cleanStringMap(req.Labels)
	if err = validateLabelMap(labels, "标签"); err != nil {
		return err
	}
	annotations := cleanStringMap(req.Annotations)
	if err = validateAnnotationKeys(annotations); err != nil {
		return err
	}
	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: withCreator(withAliasDescribe(annotations, req.Alias, req.Describe), creator),
		},
		Spec: spec,
	}
	if _, err = cs.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("服务 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建Service失败: %v", err)
		return fmt.Errorf("创建服务失败: %v", err)
	}
	logs.Info("Service %s/%s 创建成功", namespace, name)
	return nil
}

// Update 编辑服务的类型、端口、选择器、会话保持等设置; 内部访问模式(虚拟 IP / Headless)创建后不能修改
func (s *KubeSrvService) Update(ctx context.Context, req dto.ServiceFormDTO) error {
	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	svc, err := cs.CoreV1().Services(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("服务 %s 不存在", req.Name)
		}
		return fmt.Errorf("获取服务失败: %v", err)
	}
	spec, err := buildServiceSpec(req)
	if err != nil {
		return err
	}
	wasHeadless := svc.Spec.ClusterIP == corev1.ClusterIPNone
	if svc.Spec.Type != corev1.ServiceTypeExternalName && spec.Type != corev1.ServiceTypeExternalName && wasHeadless != req.Headless {
		return errors.New("内部访问模式（虚拟 IP / 内部域名）创建后不能修改，如需调整请重新创建服务")
	}
	// 未指定节点端口时沿用原来分配的, 避免端口变化
	if spec.Type == corev1.ServiceTypeNodePort || spec.Type == corev1.ServiceTypeLoadBalancer {
		for i := range spec.Ports {
			if spec.Ports[i].NodePort != 0 {
				continue
			}
			for _, old := range svc.Spec.Ports {
				if old.Port == spec.Ports[i].Port && old.Protocol == spec.Ports[i].Protocol {
					spec.Ports[i].NodePort = old.NodePort
				}
			}
		}
	}

	svc.Spec.Type = spec.Type
	svc.Spec.Selector = spec.Selector
	svc.Spec.Ports = spec.Ports
	svc.Spec.SessionAffinity = spec.SessionAffinity
	svc.Spec.SessionAffinityConfig = spec.SessionAffinityConfig
	svc.Spec.ExternalName = spec.ExternalName
	svc.Spec.ExternalTrafficPolicy = spec.ExternalTrafficPolicy
	if spec.ExternalTrafficPolicy != corev1.ServiceExternalTrafficPolicyLocal || spec.Type != corev1.ServiceTypeLoadBalancer {
		svc.Spec.HealthCheckNodePort = 0
	}
	if spec.Type != corev1.ServiceTypeLoadBalancer {
		svc.Spec.AllocateLoadBalancerNodePorts = nil
	}
	switch {
	case spec.Type == corev1.ServiceTypeExternalName:
		// 外部服务不分配集群 IP
		svc.Spec.ClusterIP = ""
		svc.Spec.ClusterIPs = nil
		svc.Spec.IPFamilies = nil
		svc.Spec.IPFamilyPolicy = nil
		svc.Spec.InternalTrafficPolicy = nil
	case svc.Spec.ClusterIP == "" && spec.ClusterIP == corev1.ClusterIPNone:
		// 由外部服务改为 Headless
		svc.Spec.ClusterIP = corev1.ClusterIPNone
	}
	svc.Annotations = withAliasDescribe(svc.Annotations, req.Alias, req.Describe)

	if _, err = cs.CoreV1().Services(svc.Namespace).Update(ctx, svc, metav1.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			return errors.New("服务已被修改，请刷新后重试")
		}
		logs.Error("更新Service失败: %v", err)
		return fmt.Errorf("更新服务失败: %v", err)
	}
	return nil
}

// Delete 删除Service
func (s *KubeSrvService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.CoreV1().Services(base.Namespace).Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		logs.Error("删除service失败: %v", err)
		return fmt.Errorf("删除service失败: %v", err)
	}
	return nil
}

// keepCreator 编辑时保留创建者注解
func keepCreator(newAnnotations, old map[string]string) map[string]string {
	for _, key := range []string{consts.CreatorAnnotation, consts.KubeSphereCreatorAnnotation} {
		if v, ok := old[key]; ok {
			if newAnnotations == nil {
				newAnnotations = map[string]string{}
			}
			newAnnotations[key] = v
		}
	}
	return newAnnotations
}

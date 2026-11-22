package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	ingressSrv  *IngressService
	ingressOnce sync.Once
)

type IngressService struct {
}

func NewIngressService() *IngressService {
	return &IngressService{}
}

func GetSingletonIngressService() *IngressService {
	ingressOnce.Do(func() {
		ingressSrv = NewIngressService()
	})
	return ingressSrv
}

func (s *IngressService) List(query dto.IngressPageQueryBaseDTO) ([]vo.IngressVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.NetworkingV1().Ingresses(query.Namespace).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取Ingress失败: %v", err)
		return nil, fmt.Errorf("获取Ingress失败: %v", err)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Name)
	}
	result := make([]vo.IngressVO, 0, len(items.Items))
	for _, item := range items.Items {
		if keyword == "" || strings.Contains(item.Name, keyword) || ingressHasHost(&item, keyword) {
			result = append(result, vo.Ingress2VO(item))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func ingressHasHost(ing *networkingv1.Ingress, keyword string) bool {
	for _, rule := range ing.Spec.Rules {
		if strings.Contains(rule.Host, keyword) {
			return true
		}
	}
	return false
}

// Get 应用路由详情
func (s *IngressService) Get(ctx context.Context, base dto.IngressBaseDTO) (*vo.IngressDetailVO, error) {
	if strings.TrimSpace(base.Namespace) == "" || strings.TrimSpace(base.Name) == "" {
		return nil, errors.New("命名空间和名称不能为空")
	}
	cs, err := ClusterMap.GetClientSet(base.ClusterId)
	if err != nil {
		return nil, err
	}
	ing, err := cs.NetworkingV1().Ingresses(base.Namespace).Get(ctx, base.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("应用路由 %s 不存在", base.Name)
		}
		return nil, fmt.Errorf("获取应用路由失败: %v", err)
	}
	serviceNames := map[string]bool{}
	if svcs, listErr := cs.CoreV1().Services(base.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"}); listErr == nil {
		for _, svc := range svcs.Items {
			serviceNames[svc.Name] = true
		}
	}
	detail := vo.Ingress2DetailVO(ing, serviceNames)
	detail.Events = listObjectEvents(ctx, cs, ing.Namespace, string(ing.UID))
	return &detail, nil
}

func buildIngressSpec(req dto.IngressFormDTO) (networkingv1.IngressSpec, error) {
	spec := networkingv1.IngressSpec{}
	if c := strings.TrimSpace(req.IngressClassName); c != "" {
		spec.IngressClassName = &c
	}
	if len(req.Rules) == 0 {
		return spec, errors.New("请至少添加一条路由规则")
	}
	for i, rule := range req.Rules {
		host := strings.ToLower(strings.TrimSpace(rule.Host))
		if host != "" {
			check := strings.TrimPrefix(host, "*.")
			if len(validation.IsDNS1123Subdomain(check)) > 0 {
				return spec, fmt.Errorf("第 %d 条规则的域名 %s 不合法", i+1, rule.Host)
			}
		}
		if len(rule.Paths) == 0 {
			return spec, fmt.Errorf("第 %d 条规则至少需要一个路径", i+1)
		}
		paths := make([]networkingv1.HTTPIngressPath, 0, len(rule.Paths))
		for j, p := range rule.Paths {
			label := fmt.Sprintf("第 %d 条规则的第 %d 个路径", i+1, j+1)
			path := strings.TrimSpace(p.Path)
			if path == "" {
				path = "/"
			}
			if !strings.HasPrefix(path, "/") {
				return spec, fmt.Errorf("%s必须以 / 开头", label)
			}
			pathType := networkingv1.PathType(strings.TrimSpace(p.PathType))
			switch pathType {
			case "":
				pathType = networkingv1.PathTypePrefix
			case networkingv1.PathTypePrefix, networkingv1.PathTypeExact, networkingv1.PathTypeImplementationSpecific:
			default:
				return spec, fmt.Errorf("%s的匹配方式不支持: %s", label, p.PathType)
			}
			svcName := strings.TrimSpace(p.ServiceName)
			if svcName == "" {
				return spec, fmt.Errorf("请选择%s转发到的服务", label)
			}
			port := strings.TrimSpace(p.ServicePort)
			if port == "" {
				return spec, fmt.Errorf("请选择%s的服务端口", label)
			}
			backend := networkingv1.IngressServiceBackend{Name: svcName}
			if n, convErr := strconv.Atoi(port); convErr == nil {
				if n < 1 || n > 65535 {
					return spec, fmt.Errorf("%s的服务端口不正确", label)
				}
				backend.Port.Number = int32(n)
			} else {
				backend.Port.Name = port
			}
			pt := pathType
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     path,
				PathType: &pt,
				Backend:  networkingv1.IngressBackend{Service: &backend},
			})
		}
		spec.Rules = append(spec.Rules, networkingv1.IngressRule{
			Host:             host,
			IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{Paths: paths}},
		})
	}
	for i, t := range req.TLS {
		secret := strings.TrimSpace(t.SecretName)
		if secret == "" {
			return spec, fmt.Errorf("请选择第 %d 个 HTTPS 配置使用的证书", i+1)
		}
		hosts := cleanStrings(t.Hosts)
		for k := range hosts {
			hosts[k] = strings.ToLower(hosts[k])
		}
		spec.TLS = append(spec.TLS, networkingv1.IngressTLS{Hosts: hosts, SecretName: secret})
	}
	return spec, nil
}

// Create 创建应用路由
func (s *IngressService) Create(ctx context.Context, req dto.IngressFormDTO, creator string) error {
	name := strings.TrimSpace(req.Name)
	if len(validation.IsDNS1123Subdomain(name)) > 0 {
		return errors.New("名称不合法: 只能包含小写字母、数字、\"-\" 和 \".\"，以字母或数字开头和结尾")
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	spec, err := buildIngressSpec(req)
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
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: withCreator(withAliasDescribe(annotations, req.Alias, req.Describe), creator),
		},
		Spec: spec,
	}
	if _, err = cs.NetworkingV1().Ingresses(namespace).Create(ctx, ing, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("应用路由 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建Ingress失败: %v", err)
		return fmt.Errorf("创建应用路由失败: %v", err)
	}
	logs.Info("Ingress %s/%s 创建成功", namespace, name)
	return nil
}

// Update 编辑路由规则、HTTPS、Ingress 类、标签和注解
func (s *IngressService) Update(ctx context.Context, req dto.IngressFormDTO) error {
	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	ing, err := cs.NetworkingV1().Ingresses(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("应用路由 %s 不存在", req.Name)
		}
		return fmt.Errorf("获取应用路由失败: %v", err)
	}
	spec, err := buildIngressSpec(req)
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
	ing.Spec.IngressClassName = spec.IngressClassName
	ing.Spec.Rules = spec.Rules
	ing.Spec.TLS = spec.TLS
	ing.Labels = labels
	ing.Annotations = withAliasDescribe(keepCreator(annotations, ing.Annotations), req.Alias, req.Describe)
	if _, err = cs.NetworkingV1().Ingresses(ing.Namespace).Update(ctx, ing, metav1.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			return errors.New("应用路由已被修改，请刷新后重试")
		}
		logs.Error("更新Ingress失败: %v", err)
		return fmt.Errorf("更新应用路由失败: %v", err)
	}
	return nil
}

// Classes 集群中的 Ingress 类
func (s *IngressService) Classes(ctx context.Context, clusterId string) ([]vo.IngressClassVO, error) {
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, err
	}
	items, err := cs.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 类失败: %v", err)
	}
	result := make([]vo.IngressClassVO, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, vo.IngressClassVO{
			Name:       item.Name,
			Controller: item.Spec.Controller,
			IsDefault:  item.Annotations[networkingv1.AnnotationIsDefaultIngressClass] == "true",
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// TLSSecrets 命名空间中的证书(kubernetes.io/tls 类型的 Secret)
func (s *IngressService) TLSSecrets(ctx context.Context, clusterId, namespace string) ([]string, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, errors.New("请选择命名空间")
	}
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, err
	}
	items, err := cs.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("type", string(corev1.SecretTypeTLS)).String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取证书失败: %v", err)
	}
	result := make([]string, 0, len(items.Items))
	for _, item := range items.Items {
		result = append(result, item.Name)
	}
	sort.Strings(result)
	return result, nil
}

// Delete 删除Ingress
func (s *IngressService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.NetworkingV1().Ingresses(base.Namespace).Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		logs.Error("删除ingress失败: %v", err)
		return fmt.Errorf("删除ingress失败: %v", err)
	}
	return nil
}

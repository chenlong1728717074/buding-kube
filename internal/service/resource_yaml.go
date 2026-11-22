package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// resourceKind 描述一种可以查看 / 编辑 YAML 的资源
type resourceKind struct {
	kind       string
	gvr        schema.GroupVersionResource
	namespaced bool
	// dropStatus 为 true 时查看 YAML 去掉 status, 便于直接编辑后提交(status 由控制器维护, 提交了也会被忽略)
	dropStatus bool
}

func gvr(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
}

// resourceKinds 按小写 Kind 索引, 与前端 api/resource.ts 中的 YamlResourceKind 保持一致
var resourceKinds = map[string]resourceKind{
	"deployment":               {"Deployment", gvr("apps", "v1", "deployments"), true, true},
	"statefulset":              {"StatefulSet", gvr("apps", "v1", "statefulsets"), true, true},
	"daemonset":                {"DaemonSet", gvr("apps", "v1", "daemonsets"), true, true},
	"replicaset":               {"ReplicaSet", gvr("apps", "v1", "replicasets"), true, true},
	"cronjob":                  {"CronJob", gvr("batch", "v1", "cronjobs"), true, true},
	"job":                      {"Job", gvr("batch", "v1", "jobs"), true, true},
	"pod":                      {"Pod", gvr("", "v1", "pods"), true, true},
	"service":                  {"Service", gvr("", "v1", "services"), true, true},
	"ingress":                  {"Ingress", gvr("networking.k8s.io", "v1", "ingresses"), true, true},
	"endpoints":                {"Endpoints", gvr("", "v1", "endpoints"), true, false},
	"endpointslice":            {"EndpointSlice", gvr("discovery.k8s.io", "v1", "endpointslices"), true, false},
	"configmap":                {"ConfigMap", gvr("", "v1", "configmaps"), true, false},
	"secret":                   {"Secret", gvr("", "v1", "secrets"), true, false},
	"serviceaccount":           {"ServiceAccount", gvr("", "v1", "serviceaccounts"), true, false},
	"namespace":                {"Namespace", gvr("", "v1", "namespaces"), false, true},
	"node":                     {"Node", gvr("", "v1", "nodes"), false, true},
	"persistentvolume":         {"PersistentVolume", gvr("", "v1", "persistentvolumes"), false, true},
	"persistentvolumeclaim":    {"PersistentVolumeClaim", gvr("", "v1", "persistentvolumeclaims"), true, true},
	"storageclass":             {"StorageClass", gvr("storage.k8s.io", "v1", "storageclasses"), false, false},
	"customresourcedefinition": {"CustomResourceDefinition", gvr("apiextensions.k8s.io", "v1", "customresourcedefinitions"), false, true},
	"horizontalpodautoscaler":  {"HorizontalPodAutoscaler", gvr("autoscaling", "v2", "horizontalpodautoscalers"), true, true},
	"networkpolicy":            {"NetworkPolicy", gvr("networking.k8s.io", "v1", "networkpolicies"), true, true},
	"resourcequota":            {"ResourceQuota", gvr("", "v1", "resourcequotas"), true, true},
	"limitrange":               {"LimitRange", gvr("", "v1", "limitranges"), true, false},
	"poddisruptionbudget":      {"PodDisruptionBudget", gvr("policy", "v1", "poddisruptionbudgets"), true, true},
	"priorityclass":            {"PriorityClass", gvr("scheduling.k8s.io", "v1", "priorityclasses"), false, false},
	"ingressclass":             {"IngressClass", gvr("networking.k8s.io", "v1", "ingressclasses"), false, false},
}

// customResourceKind 自定义资源(CRD 的实例), 由调用方指定 group / version / resource / kind
const customResourceKind = "custom"

// ResourceRef 资源类型: Kind 为内置类型(见 resourceKinds); Kind 为 custom 时用 Group/Version/Resource/KindName 指定自定义资源
type ResourceRef struct {
	Kind       string
	Group      string
	Version    string
	Resource   string
	KindName   string
	Namespaced bool
}

func (r ResourceRef) resolve() (resourceKind, error) {
	if !strings.EqualFold(strings.TrimSpace(r.Kind), customResourceKind) {
		return lookupResourceKind(r.Kind)
	}
	group, version := strings.TrimSpace(r.Group), strings.TrimSpace(r.Version)
	resource, kind := strings.TrimSpace(r.Resource), strings.TrimSpace(r.KindName)
	if version == "" || resource == "" || kind == "" {
		return resourceKind{}, errors.New("自定义资源需要指定 version、resource 和 kindName")
	}
	if group != "" && len(validation.IsDNS1123Subdomain(group)) > 0 {
		return resourceKind{}, fmt.Errorf("API 组不合法: %s", group)
	}
	if len(validation.IsDNS1123Label(version)) > 0 || len(validation.IsDNS1123Label(resource)) > 0 {
		return resourceKind{}, fmt.Errorf("资源 %s/%s 不合法", version, resource)
	}
	return resourceKind{kind: kind, gvr: gvr(group, version, resource), namespaced: r.Namespaced, dropStatus: true}, nil
}

func lookupResourceKind(kind string) (resourceKind, error) {
	rk, ok := resourceKinds[strings.ToLower(strings.TrimSpace(kind))]
	if !ok {
		return resourceKind{}, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	return rk, nil
}

// dynamicClients 按集群的 rest.Config 缓存动态客户端; 集群配置更新后会生成新的 Config, 自然换成新的客户端
var dynamicClients sync.Map

func getDynamicClient(clusterId string) (dynamic.Interface, error) {
	cfg, err := ClusterMap.GetConfig(clusterId)
	if err != nil {
		return nil, fmt.Errorf("获取集群失败: %w", err)
	}
	if c, ok := dynamicClients.Load(cfg); ok {
		return c.(dynamic.Interface), nil
	}
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建集群客户端失败: %w", err)
	}
	actual, _ := dynamicClients.LoadOrStore(cfg, c)
	return actual.(dynamic.Interface), nil
}

func resourceClient(dyn dynamic.Interface, rk resourceKind, namespace string) dynamic.ResourceInterface {
	if rk.namespaced {
		return dyn.Resource(rk.gvr).Namespace(namespace)
	}
	return dyn.Resource(rk.gvr)
}

// ResourceYaml 查询单个资源并生成精简后的 YAML(去掉 managedFields、resourceVersion 等运行时字段).
// 列表接口不再逐条生成 YAML, 查看 YAML 时由前端按需调用这里
func ResourceYaml(ctx context.Context, clusterId string, ref ResourceRef, namespace, name string) (string, error) {
	rk, err := ref.resolve()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("资源名称不能为空")
	}
	if rk.namespaced && namespace == "" {
		return "", errors.New("命名空间不能为空")
	}
	dyn, err := getDynamicClient(clusterId)
	if err != nil {
		return "", err
	}
	obj, err := resourceClient(dyn, rk, namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", fmt.Errorf("%s %s 不存在", rk.kind, name)
		}
		return "", fmt.Errorf("获取%s失败: %w", rk.kind, err)
	}
	return marshalCleanYaml(obj, rk)
}

func marshalCleanYaml(obj *unstructured.Unstructured, rk resourceKind) (string, error) {
	content := obj.DeepCopy().Object
	if meta, ok := content["metadata"].(map[string]interface{}); ok {
		delete(meta, "managedFields")
		delete(meta, "resourceVersion")
		delete(meta, "creationTimestamp")
	}
	if rk.dropStatus {
		delete(content, "status")
	}
	data, err := yaml.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("序列化%s失败: %w", rk.kind, err)
	}
	return string(data), nil
}

// ApplyYamlParams 保存 YAML 的参数
type ApplyYamlParams struct {
	ClusterId string
	Ref       ResourceRef
	// Namespace / Name 为编辑前的资源标识; 有值时不允许在 YAML 里修改名称和命名空间
	Namespace string
	Name      string
	Yaml      string
	// AllowCreate 资源不存在时是否创建(YAML 添加); 编辑时为 false, 避免资源已被删除时又被重新创建
	AllowCreate bool
	// CreateOnly 只创建, 同名资源已存在时报错(表单创建时使用, 避免覆盖已有资源)
	CreateOnly bool
}

// ApplyResourceYaml 用 YAML 更新(或创建)资源.
// 以服务端当前的 resourceVersion 提交(查看 YAML 时去掉了这个字段), 即以最后一次保存为准
func ApplyResourceYaml(ctx context.Context, p ApplyYamlParams) error {
	rk, err := p.Ref.resolve()
	if err != nil {
		return err
	}
	text := strings.TrimSpace(p.Yaml)
	if text == "" {
		return errors.New("YAML 内容不能为空")
	}
	jsonData, err := yaml.YAMLToJSON([]byte(text))
	if err != nil {
		return fmt.Errorf("YAML 格式错误: %v", err)
	}
	obj := &unstructured.Unstructured{}
	if err = obj.UnmarshalJSON(jsonData); err != nil {
		return fmt.Errorf("YAML 内容无效: %v", err)
	}
	if obj.GetKind() != rk.kind {
		return fmt.Errorf("kind 应为 %s, 实际为 %s", rk.kind, obj.GetKind())
	}
	if gv, gvErr := schema.ParseGroupVersion(obj.GetAPIVersion()); gvErr != nil || gv.Group != rk.gvr.Group {
		return fmt.Errorf("apiVersion 不正确: %s", obj.GetAPIVersion())
	}

	name := obj.GetName()
	if name == "" {
		return errors.New("metadata.name 不能为空")
	}
	if p.Name != "" && name != p.Name {
		return fmt.Errorf("不能修改资源名称(原名称为 %s), 如需改名请新建资源", p.Name)
	}
	namespace := ""
	if rk.namespaced {
		namespace = obj.GetNamespace()
		if namespace == "" {
			namespace = p.Namespace
		}
		if namespace == "" {
			return errors.New("metadata.namespace 不能为空")
		}
		if p.Namespace != "" && namespace != p.Namespace {
			return fmt.Errorf("不能修改命名空间(原命名空间为 %s)", p.Namespace)
		}
		obj.SetNamespace(namespace)
	} else {
		obj.SetNamespace("")
	}
	obj.SetManagedFields(nil)

	dyn, err := getDynamicClient(p.ClusterId)
	if err != nil {
		return err
	}
	client := resourceClient(dyn, rk, namespace)
	existing, err := client.Get(ctx, name, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		if !p.AllowCreate {
			return fmt.Errorf("%s %s 不存在, 可能已被删除", rk.kind, name)
		}
		obj.SetResourceVersion("")
		if _, err = client.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("创建%s失败: %v", rk.kind, err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("获取%s失败: %v", rk.kind, err)
	case p.CreateOnly:
		return fmt.Errorf("%s %s 已存在", rk.kind, name)
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	if _, err = client.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("更新%s失败: %v", rk.kind, err)
	}
	return nil
}

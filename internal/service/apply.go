package service

import (
	"sync"
)

var (
	applySrv  *ApplyService
	applyOnce sync.Once
)

type ApplyService struct {
}

func NewApplyService() *ApplyService {
	return &ApplyService{}
}

func GetSingletonApplyService() *ApplyService {
	applyOnce.Do(func() {
		applySrv = NewApplyService()
	})
	return applySrv
}

/*
// SimpleKubectlApply 简单的 kubectl apply 实现
type SimpleKubectlApply struct {
	dynamicClient dynamic.Interface
	restMapper    *restmapper.DeferredDiscoveryRESTMapper
}

// NewSimpleKubectlApply 创建 apply 客户端
func NewSimpleKubectlApply(kubeconfigPath string) (*SimpleKubectlApply, error) {
	// 1. 加载配置 - 支持远程 K8s
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	// 2. 创建 dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// 3. 创建 discovery client 用于资源映射
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %v", err)
	}
	cachedDiscoveryClient := memory.NewMemCacheClient(discoveryClient)

	// 4. 创建 REST mapper
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)

	return &SimpleKubectlApply{
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
	}, nil
}

// Apply 执行 kubectl apply 逻辑
func (s *SimpleKubectlApply) Apply(yamlContent string) error {
	// 分割多个 YAML 文档
	documents := strings.Split(yamlContent, "---")

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		if err := s.applyDocument(doc); err != nil {
			return fmt.Errorf("failed to apply document: %v", err)
		}
	}

	return nil
}

// applyDocument 处理单个 YAML 文档
func (s *SimpleKubectlApply) applyDocument(yamlDoc string) error {
	// 1. 解析 YAML 为 Unstructured 对象
	obj := &unstructured.Unstructured{}
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decoder.Decode([]byte(yamlDoc), nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %v", err)
	}

	// 2. 获取 GVR (GroupVersionResource)
	gvr, err := s.getGVR(*gvk)
	if err != nil {
		return fmt.Errorf("failed to get GVR for %s: %v", gvk, err)
	}

	// 3. 执行 Server-Side Apply - 这是最接近 kubectl apply 的方式
	ctx := context.TODO()
	namespace := obj.GetNamespace()
	name := obj.GetName()

	// 如果是 namespace 级别的资源
	if namespace != "" {
		_, err = s.dynamicClient.Resource(gvr).Namespace(namespace).Apply(
			ctx, name, obj, metav1.ApplyOptions{
				FieldManager: "simple-kubectl-apply",
			})
	} else {
		// 集群级别的资源
		_, err = s.dynamicClient.Resource(gvr).Apply(
			ctx, name, obj, metav1.ApplyOptions{
				FieldManager: "simple-kubectl-apply",
			})
	}

	if err != nil {
		return fmt.Errorf("failed to apply %s/%s: %v", gvk.Kind, name, err)
	}

	fmt.Printf("Applied %s/%s\n", gvk.Kind, name)
	return nil
}

// getGVR 将 GVK 转换为 GVR
func (s *SimpleKubectlApply) getGVR(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	mapping, err := s.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil
}
*/

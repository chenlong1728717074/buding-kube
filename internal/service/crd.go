package service

import (
	"buding-kube/internal/web/vo"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CRD 对象包含完整的 OpenAPI schema, 安装了 Istio / Crossplane / Prometheus Operator 等组件的集群,
// 全部 CRD 加起来可能有几十 MB. 列表页只需要名称、组、版本等概要, 因此:
//   - 一次查询全部 CRD 后只保留概要并缓存, 翻页、搜索都在缓存上进行, 不再重复查询
//   - 使用 resourceVersion=0 由 apiserver 的缓存返回, 并使用 protobuf, 减少传输和解析时间
const (
	crdCacheTTL     = time.Minute
	crdListTimeout  = 60 * time.Second
	protobufContent = "application/vnd.kubernetes.protobuf"
)

type crdCacheEntry struct {
	// 集群配置更新后会生成新的 Config, 旧缓存随之失效
	cfg   *rest.Config
	at    time.Time
	items []vo.CrdVO
}

var (
	crdCache   sync.Map // clusterId -> crdCacheEntry
	crdLoading sync.Map // clusterId -> *sync.Mutex, 同一集群同时只查询一次
	crdClients sync.Map // *rest.Config -> *apiextensionsclientset.Clientset
)

// crdClient apiextensions 客户端(kubernetes.Clientset 不包含 CRD), 按集群配置缓存
func crdClient(clusterId string) (*apiextensionsclientset.Clientset, *rest.Config, error) {
	cfg, err := ClusterMap.GetConfig(clusterId)
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群失败: %w", err)
	}
	if c, ok := crdClients.Load(cfg); ok {
		return c.(*apiextensionsclientset.Clientset), cfg, nil
	}
	pbCfg := rest.CopyConfig(cfg)
	pbCfg.ContentType = protobufContent
	pbCfg.AcceptContentTypes = protobufContent + ",application/json"
	c, err := apiextensionsclientset.NewForConfig(pbCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("创建集群客户端失败: %w", err)
	}
	actual, _ := crdClients.LoadOrStore(cfg, c)
	return actual.(*apiextensionsclientset.Clientset), cfg, nil
}

func crd2VO(item *apiextensionsv1.CustomResourceDefinition) vo.CrdVO {
	result := vo.CrdVO{
		Name:       item.Name,
		Kind:       item.Spec.Names.Kind,
		ListKind:   item.Spec.Names.ListKind,
		Plural:     item.Spec.Names.Plural,
		Singular:   item.Spec.Names.Singular,
		ShortNames: item.Spec.Names.ShortNames,
		Group:      item.Spec.Group,
		Scope:      string(item.Spec.Scope),
		Versions:   make([]string, 0, len(item.Spec.Versions)),
		CreateTime: item.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}
	if result.ShortNames == nil {
		result.ShortNames = []string{}
	}
	for _, ver := range item.Spec.Versions {
		if ver.Storage {
			result.Version = ver.Name
		}
	}
	if result.Version != "" {
		result.Versions = append(result.Versions, result.Version)
	}
	for _, ver := range item.Spec.Versions {
		if ver.Served && ver.Name != result.Version {
			result.Versions = append(result.Versions, ver.Name)
		}
	}
	return result
}

// ListCrds 集群中所有 CRD 的概要, 按名称排序; refresh 为 true 时忽略缓存
func ListCrds(ctx context.Context, clusterId string, refresh bool) ([]vo.CrdVO, error) {
	client, cfg, err := crdClient(clusterId)
	if err != nil {
		return nil, err
	}
	fresh := func() ([]vo.CrdVO, bool) {
		if v, ok := crdCache.Load(clusterId); ok {
			e := v.(crdCacheEntry)
			if e.cfg == cfg && time.Since(e.at) < crdCacheTTL {
				return e.items, true
			}
		}
		return nil, false
	}
	if !refresh {
		if items, ok := fresh(); ok {
			return items, nil
		}
	}
	lockV, _ := crdLoading.LoadOrStore(clusterId, &sync.Mutex{})
	lock := lockV.(*sync.Mutex)
	start := time.Now()
	lock.Lock()
	defer lock.Unlock()
	// 等锁期间其他请求可能已经查询完成
	if v, ok := crdCache.Load(clusterId); ok {
		if e := v.(crdCacheEntry); e.cfg == cfg && e.at.After(start) {
			return e.items, nil
		}
	}
	if !refresh {
		if items, ok := fresh(); ok {
			return items, nil
		}
	}

	listCtx, cancel := context.WithTimeout(ctx, crdListTimeout)
	defer cancel()
	list, err := client.ApiextensionsV1().CustomResourceDefinitions().List(listCtx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("没有权限查询 CRD，请检查集群凭证的 RBAC 权限")
		}
		return nil, fmt.Errorf("查询 CRD 失败: %v", err)
	}
	items := make([]vo.CrdVO, 0, len(list.Items))
	for i := range list.Items {
		items = append(items, crd2VO(&list.Items[i]))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	crdCache.Store(clusterId, crdCacheEntry{cfg: cfg, at: time.Now(), items: items})
	return items, nil
}

// FilterCrds 按名称 / Kind / 组过滤, 不区分大小写
func FilterCrds(items []vo.CrdVO, keyword string) []vo.CrdVO {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return items
	}
	result := make([]vo.CrdVO, 0)
	for _, c := range items {
		if strings.Contains(strings.ToLower(c.Name), keyword) ||
			strings.Contains(strings.ToLower(c.Kind), keyword) ||
			strings.Contains(strings.ToLower(c.Group), keyword) {
			result = append(result, c)
		}
	}
	return result
}

// GetCrd 单个 CRD 的概要; 只查这一个对象, 不加载整个列表
func GetCrd(ctx context.Context, clusterId, name string) (*vo.CrdVO, error) {
	client, _, err := crdClient(clusterId)
	if err != nil {
		return nil, err
	}
	item, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("CRD %s 不存在", name)
		}
		return nil, fmt.Errorf("获取 CRD 失败: %v", err)
	}
	result := crd2VO(item)
	return &result, nil
}

// DeleteCrd 删除 CRD(连带其下所有自定义资源)
func DeleteCrd(ctx context.Context, clusterId, name string) error {
	client, _, err := crdClient(clusterId)
	if err != nil {
		return err
	}
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err = client.ApiextensionsV1().CustomResourceDefinitions().Delete(deleteCtx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("删除 CRD 失败: %v", err)
	}
	crdCache.Delete(clusterId)
	return nil
}

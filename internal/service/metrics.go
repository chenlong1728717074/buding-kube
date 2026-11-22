package service

import (
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
)

// 容器组 / 节点的实时用量来自 metrics-server(metrics.k8s.io).
// 没有安装或服务异常时, 一段时间内不再请求, 避免每次列表都等到超时
const (
	metricsTimeout = 3 * time.Second
	// metricsRetryInterval 判定不可用后, 多久之后再尝试
	metricsRetryInterval = time.Minute
	metricsBasePath      = "/apis/metrics.k8s.io/v1beta1"
)

// metricsUnavailable 按集群客户端记录 metrics-server 不可用的截止时间; 集群配置更新后客户端会重建, 状态自然失效
var metricsUnavailable sync.Map

func metricsDisabled(cs *kubernetes.Clientset) bool {
	v, ok := metricsUnavailable.Load(cs)
	if !ok {
		return false
	}
	if time.Now().Before(v.(time.Time)) {
		return true
	}
	metricsUnavailable.Delete(cs)
	return false
}

// metricsGet 请求 metrics.k8s.io; 失败时返回 nil.
// single 为 true 时是查询单个对象: 404 只表示该对象还没有采集到, 不算服务不可用
func metricsGet(ctx context.Context, cs *kubernetes.Clientset, path string, single bool) []byte {
	if metricsDisabled(cs) {
		return nil
	}
	reqCtx, cancel := context.WithTimeout(ctx, metricsTimeout)
	defer cancel()
	raw, err := cs.RESTClient().Get().AbsPath(metricsBasePath + path).DoRaw(reqCtx)
	switch {
	case err == nil:
		metricsUnavailable.Delete(cs)
		return raw
	case ctx.Err() != nil:
		// 调用方已取消(例如浏览器关闭了页面), 不代表服务不可用
	case single && apierrors.IsNotFound(err):
	default:
		if _, loaded := metricsUnavailable.Swap(cs, time.Now().Add(metricsRetryInterval)); !loaded {
			logs.Warn("metrics-server 不可用, %s 内不再请求用量: %v", metricsRetryInterval, err)
		}
	}
	return nil
}

func parseQuantity(value string, milli bool) int64 {
	q, err := resource.ParseQuantity(value)
	if err != nil {
		return 0
	}
	if milli {
		return q.MilliValue()
	}
	return q.Value()
}

// ---------- 容器组 ----------

type podMetricsItem struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Containers []struct {
		Name  string `json:"name"`
		Usage struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"usage"`
	} `json:"containers"`
}

// podMetrics 一个容器组的用量: CPU 毫核, 内存字节
type podMetrics struct {
	cpu        int64
	memory     int64
	containers map[string][2]int64
}

func toPodMetrics(item *podMetricsItem) *podMetrics {
	m := &podMetrics{containers: make(map[string][2]int64, len(item.Containers))}
	for _, c := range item.Containers {
		cpu, mem := parseQuantity(c.Usage.CPU, true), parseQuantity(c.Usage.Memory, false)
		m.cpu += cpu
		m.memory += mem
		m.containers[c.Name] = [2]int64{cpu, mem}
	}
	return m
}

// fetchPodMetrics 命名空间(为空表示全部)中所有容器组的用量, 键为 命名空间/名称; 不可用时返回 nil
func fetchPodMetrics(ctx context.Context, cs *kubernetes.Clientset, namespace string) map[string]*podMetrics {
	path := "/pods"
	if namespace != "" {
		path = fmt.Sprintf("/namespaces/%s/pods", url.PathEscape(namespace))
	}
	raw := metricsGet(ctx, cs, path, false)
	if raw == nil {
		return nil
	}
	var resp struct {
		Items []podMetricsItem `json:"items"`
	}
	if json.Unmarshal(raw, &resp) != nil {
		return nil
	}
	result := make(map[string]*podMetrics, len(resp.Items))
	for i := range resp.Items {
		item := &resp.Items[i]
		result[item.Metadata.Namespace+"/"+item.Metadata.Name] = toPodMetrics(item)
	}
	return result
}

// fetchSinglePodMetrics 单个容器组的用量; 刚启动(还没有采集到)或不可用时返回 nil
func fetchSinglePodMetrics(ctx context.Context, cs *kubernetes.Clientset, namespace, name string) *podMetrics {
	raw := metricsGet(ctx, cs, fmt.Sprintf("/namespaces/%s/pods/%s", url.PathEscape(namespace), url.PathEscape(name)), true)
	if raw == nil {
		return nil
	}
	var item podMetricsItem
	if json.Unmarshal(raw, &item) != nil {
		return nil
	}
	return toPodMetrics(&item)
}

func (m *podMetrics) usage() vo.PodUsage {
	if m == nil {
		return vo.PodUsage{}
	}
	cpu, mem := m.cpu, m.memory
	return vo.PodUsage{CPUUsage: &cpu, MemoryUsage: &mem}
}

func (m *podMetrics) containerUsage(name string) vo.PodUsage {
	if m == nil {
		return vo.PodUsage{}
	}
	v, ok := m.containers[name]
	if !ok {
		return vo.PodUsage{}
	}
	cpu, mem := v[0], v[1]
	return vo.PodUsage{CPUUsage: &cpu, MemoryUsage: &mem}
}

// attachJobPodMetrics 给工作负载 / 任务 / 服务详情中的容器组补充用量
func attachJobPodMetrics(ctx context.Context, cs *kubernetes.Clientset, namespace string, pods []vo.JobPodVO) {
	if len(pods) == 0 {
		return
	}
	metrics := fetchPodMetrics(ctx, cs, namespace)
	if metrics == nil {
		return
	}
	for i := range pods {
		pods[i].PodUsage = metrics[pods[i].Namespace+"/"+pods[i].Name].usage()
	}
}

// ---------- 节点 ----------

type nodeMetricsItem struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Usage struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
	} `json:"usage"`
}

func (item *nodeMetricsItem) usage() [2]int64 {
	return [2]int64{parseQuantity(item.Usage.CPU, true), parseQuantity(item.Usage.Memory, false)}
}

// fetchNodeMetrics 所有节点的用量 [CPU 毫核, 内存字节]; 不可用时返回 nil
func fetchNodeMetrics(ctx context.Context, cs *kubernetes.Clientset) map[string][2]int64 {
	raw := metricsGet(ctx, cs, "/nodes", false)
	if raw == nil {
		return nil
	}
	var resp struct {
		Items []nodeMetricsItem `json:"items"`
	}
	if json.Unmarshal(raw, &resp) != nil {
		return nil
	}
	result := make(map[string][2]int64, len(resp.Items))
	for i := range resp.Items {
		result[resp.Items[i].Metadata.Name] = resp.Items[i].usage()
	}
	return result
}

// fetchSingleNodeMetrics 单个节点的用量; 不可用时返回 nil
func fetchSingleNodeMetrics(ctx context.Context, cs *kubernetes.Clientset, name string) *[2]int64 {
	raw := metricsGet(ctx, cs, "/nodes/"+url.PathEscape(name), true)
	if raw == nil {
		return nil
	}
	var item nodeMetricsItem
	if json.Unmarshal(raw, &item) != nil {
		return nil
	}
	u := item.usage()
	return &u
}

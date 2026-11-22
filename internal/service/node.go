package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	nodeSrv  *NodeService
	nodeOnce sync.Once
)

type NodeService struct {
}

func GetSingletonNodeService() *NodeService {
	nodeOnce.Do(func() {
		nodeSrv = NewNodeService()
	})
	return nodeSrv
}

func NewNodeService() *NodeService {
	return &NodeService{}
}

func (s NodeService) List(query dto.NodeQueryDTO) ([]vo.NodeVO, error) {
	//获取连接
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, err
	}
	list, _ := clientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logs.Error("获取节点失败 集群:%s 异常:%v", query.ClusterId, err)
		return nil, err
	}
	result := make([]vo.NodeVO, 0)
	for _, item := range list.Items {
		result = append(result, vo.Node2VO(&item))
	}
	return result, nil
}

func (s *NodeService) UnSchedule(query dto.NodeUnScheduleDTO) error {
	node, clientSet, err := s.GetNode(query.NodeInfoQueryDTO)
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = *query.UnSchedule
	_, err = clientSet.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("更新节点失败: %s %s %s", query.ClusterId, query.Hostname, err.Error())
		return err
	}
	return nil
}

func (s *NodeService) GetNode(query dto.NodeInfoQueryDTO) (*corev1.Node, *kubernetes.Clientset, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, nil, err
	}
	node, err := clientSet.CoreV1().Nodes().Get(context.TODO(), query.Hostname, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取节点失败: %s %s %s", query.ClusterId, query.Hostname, err.Error())
		return nil, nil, err
	}
	return node, clientSet, nil
}

func (s *NodeService) GetNodeInfo(query dto.NodeInfoQueryDTO) (*vo.NodeInfoVO, error) {
	//获取node
	node, clientSet, err := s.GetNode(query)
	if err != nil {
		return nil, err
	}
	//获取事件
	events, err := clientSet.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "involvedObject.kind=Node,involvedObject.name=" + node.Name,
	})
	if err != nil {
		logs.Error("获取节点事件失败: %s %s %v", query.ClusterId, query.Hostname, err)
		return nil, err
	}
	//获取pod
	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})
	if err != nil {
		logs.Error("获取节点事件失败: %s %s %v", query.ClusterId, query.Hostname, err)
		return nil, err
	}
	//metrics-server相关暂时不获取
	return vo.Node2InfoVO(node, events, pods), nil
}

// Delete 删除Node对象(节点重新注册后会再次出现)
func (s *NodeService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return err
	}
	if err := clientSet.CoreV1().Nodes().Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除node失败: %v", err)
		return err
	}
	return nil
}

type nodeMetricsListResponse struct {
	Items []struct {
		Usage struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"usage"`
	} `json:"items"`
}

type kubeletStatsSummaryResponse struct {
	Node struct {
		Fs *struct {
			CapacityBytes *uint64 `json:"capacityBytes"`
			UsedBytes     *uint64 `json:"usedBytes"`
		} `json:"fs"`
	} `json:"node"`
}

// GetClusterUsage 聚合集群节点 CPU / 内存 / 存储实时使用率及节点/Pod健康状态
// CPU/内存优先取 metrics-server 实时指标，未安装时回退到活跃 Pod Requests
// 存储优先取 kubelet /stats/summary 文件系统用量，不可达时回退到 PVC 与 Pod 临时存储请求
func (s *NodeService) GetClusterUsage(query dto.NodeUsageQueryDTO) (*vo.ClusterResourceUsageVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	nodeList, err := clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Error("获取节点列表失败 集群:%s 异常:%v", query.ClusterId, err)
		return nil, err
	}

	var (
		nodeTotal         int64
		nodeReady         int64
		cpuTotalMilli     int64
		memTotalBytes     int64
		storageTotalBytes int64
		nodeNames         = make([]string, 0, len(nodeList.Items))
	)

	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		nodeTotal++
		if vo.GetNodeStatus(node) == "Ready" {
			nodeReady++
		}
		nodeNames = append(nodeNames, node.Name)

		cpuQ := node.Status.Allocatable.Cpu()
		if cpuQ.IsZero() {
			cpuQ = node.Status.Capacity.Cpu()
		}
		cpuTotalMilli += cpuQ.MilliValue()

		memQ := node.Status.Allocatable.Memory()
		if memQ.IsZero() {
			memQ = node.Status.Capacity.Memory()
		}
		memTotalBytes += memQ.Value()

		storageQ := node.Status.Allocatable.StorageEphemeral()
		if storageQ.IsZero() {
			storageQ = node.Status.Capacity.StorageEphemeral()
		}
		storageTotalBytes += storageQ.Value()
	}

	var (
		wg                 sync.WaitGroup
		metricsOK          bool
		metricsCPUMilli    int64
		metricsMemBytes    int64
		podTotal           int64
		podRunning         int64
		podReqCPUMilli     int64
		podReqMemBytes     int64
		podReqStorageBytes int64
		fsOK               bool
		fsUsedBytes        uint64
		fsCapBytes         uint64
		pvcUsedBytes       int64
	)

	// 1. 尝试从 metrics-server 获取实时节点 CPU/内存用量
	wg.Add(1)
	go func() {
		defer wg.Done()
		raw, mErr := clientSet.RESTClient().Get().AbsPath("/apis/metrics.k8s.io/v1beta1/nodes").DoRaw(ctx)
		if mErr != nil {
			return
		}
		var resp nodeMetricsListResponse
		if json.Unmarshal(raw, &resp) != nil || len(resp.Items) == 0 {
			return
		}
		var cpuSum, memSum int64
		for _, item := range resp.Items {
			if q, qErr := resource.ParseQuantity(item.Usage.CPU); qErr == nil {
				cpuSum += q.MilliValue()
			}
			if q, qErr := resource.ParseQuantity(item.Usage.Memory); qErr == nil {
				memSum += q.Value()
			}
		}
		metricsCPUMilli = cpuSum
		metricsMemBytes = memSum
		metricsOK = true
	}()

	// 2. 统计 Pod 运行数及 Requests 回退值
	wg.Add(1)
	go func() {
		defer wg.Done()
		podList, pErr := clientSet.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if pErr != nil {
			return
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			podTotal++
			if pod.Status.Phase == corev1.PodRunning {
				podRunning++
			}
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				continue
			}
			for _, c := range pod.Spec.Containers {
				if c.Resources.Requests != nil {
					podReqCPUMilli += c.Resources.Requests.Cpu().MilliValue()
					podReqMemBytes += c.Resources.Requests.Memory().Value()
					podReqStorageBytes += c.Resources.Requests.StorageEphemeral().Value()
				}
			}
		}
	}()

	// 3. 并发查询节点 kubelet /stats/summary 获取真实文件系统使用率
	wg.Add(1)
	go func() {
		defer wg.Done()
		var (
			fsWg sync.WaitGroup
			mu   sync.Mutex
			sem  = make(chan struct{}, 6)
		)
		for _, name := range nodeNames {
			fsWg.Add(1)
			go func(nodeName string) {
				defer fsWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				sCtx, sCancel := context.WithTimeout(ctx, 3*time.Second)
				defer sCancel()

				raw, sErr := clientSet.CoreV1().RESTClient().Get().
					Resource("nodes").
					Name(nodeName).
					SubResource("proxy", "stats", "summary").
					DoRaw(sCtx)
				if sErr != nil {
					return
				}
				var summary kubeletStatsSummaryResponse
				if json.Unmarshal(raw, &summary) != nil || summary.Node.Fs == nil {
					return
				}
				if summary.Node.Fs.UsedBytes != nil && summary.Node.Fs.CapacityBytes != nil && *summary.Node.Fs.CapacityBytes > 0 {
					mu.Lock()
					fsUsedBytes += *summary.Node.Fs.UsedBytes
					fsCapBytes += *summary.Node.Fs.CapacityBytes
					fsOK = true
					mu.Unlock()
				}
			}(name)
		}
		fsWg.Wait()
	}()

	// 4. PVC 已绑定存储量作为存储回退来源
	wg.Add(1)
	go func() {
		defer wg.Done()
		pvcList, pvcErr := clientSet.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if pvcErr != nil {
			return
		}
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			if pvc.Status.Phase == corev1.ClaimBound {
				if q, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
					pvcUsedBytes += q.Value()
					continue
				}
			}
			if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				pvcUsedBytes += q.Value()
			}
		}
	}()

	wg.Wait()

	// 组装 CPU 与内存指标
	source := "requests"
	cpuUsedMilli := podReqCPUMilli
	memUsedBytes := podReqMemBytes
	if metricsOK {
		source = "metrics-server"
		cpuUsedMilli = metricsCPUMilli
		memUsedBytes = metricsMemBytes
	}

	const bytesPerGiB = 1024.0 * 1024.0 * 1024.0

	cpuUsed := float64(cpuUsedMilli) / 1000.0
	cpuTotal := float64(cpuTotalMilli) / 1000.0

	memUsedGi := float64(memUsedBytes) / bytesPerGiB
	memTotalGi := float64(memTotalBytes) / bytesPerGiB

	var storageUsedGi, storageTotalGi float64
	if fsOK && fsCapBytes > 0 {
		storageUsedGi = float64(fsUsedBytes) / bytesPerGiB
		storageTotalGi = float64(fsCapBytes) / bytesPerGiB
	} else {
		storageUsedGi = float64(pvcUsedBytes+podReqStorageBytes) / bytesPerGiB
		storageTotalGi = float64(storageTotalBytes) / bytesPerGiB
	}

	return &vo.ClusterResourceUsageVO{
		CPU:        buildResourceMetric(cpuUsed, cpuTotal, "核"),
		Memory:     buildResourceMetric(memUsedGi, memTotalGi, "Gi"),
		Storage:    buildResourceMetric(storageUsedGi, storageTotalGi, "Gi"),
		NodeTotal:  nodeTotal,
		NodeReady:  nodeReady,
		PodTotal:   podTotal,
		PodRunning: podRunning,
		Source:     source,
	}, nil
}

func buildResourceMetric(used, total float64, unit string) vo.ResourceMetricVO {
	if used < 0 {
		used = 0
	}
	if total < 0 {
		total = 0
	}
	percent := 0.0
	if total > 0 {
		percent = math.Min(100, math.Max(0, (used/total)*100))
	}
	return vo.ResourceMetricVO{
		Used:      roundTo(used, 2),
		Total:     roundTo(total, 2),
		Percent:   roundTo(percent, 1),
		UsedText:  formatMetricNumber(used),
		TotalText: fmt.Sprintf("%s %s", formatMetricNumber(total), unit),
	}
}

func roundTo(val float64, digits int) float64 {
	factor := math.Pow(10, float64(digits))
	return math.Round(val*factor) / factor
}

func formatMetricNumber(val float64) string {
	if val == 0 {
		return "0"
	}
	var s string
	switch {
	case val < 1:
		s = fmt.Sprintf("%.2f", val)
	case val < 100:
		s = fmt.Sprintf("%.1f", val)
	default:
		s = fmt.Sprintf("%.0f", val)
	}
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	topologySrv  *TopologyService
	topologyOnce sync.Once
)

type TopologyService struct{}

func NewTopologyService() *TopologyService {
	return &TopologyService{}
}

func GetSingletonTopologyService() *TopologyService {
	topologyOnce.Do(func() {
		topologySrv = NewTopologyService()
	})
	return topologySrv
}

func topoID(kind, name string) string {
	return kind + "/" + name
}

func labelsMatch(selector, labels map[string]string) bool {
	if len(selector) == 0 || len(labels) == 0 {
		return false
	}
	for k, v := range selector {
		if lv, ok := labels[k]; !ok || lv != v {
			return false
		}
	}
	return true
}

func workloadStatusType(status string) string {
	switch status {
	case "运行中":
		return "success"
	case "更新中":
		return "warning"
	case "已停止", "未知":
		return "info"
	default:
		return "danger"
	}
}

func podTopoStatus(pod *corev1.Pod) (string, string, string) {
	ready := 0
	total := len(pod.Spec.Containers)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			r := cs.State.Waiting.Reason
			if r != "ContainerCreating" && r != "PodInitializing" {
				return r, "danger", fmt.Sprintf("%d/%d 就绪", ready, total)
			}
		}
		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 && pod.Status.Phase != corev1.PodSucceeded {
			reason := cs.State.Terminated.Reason
			if reason == "" {
				reason = "Error"
			}
			return reason, "danger", fmt.Sprintf("%d/%d 就绪", ready, total)
		}
	}
	summary := fmt.Sprintf("%d/%d 就绪", ready, total)
	switch pod.Status.Phase {
	case corev1.PodRunning:
		if total > 0 && ready == total {
			return "Running", "success", summary
		}
		return "Running", "warning", summary
	case corev1.PodSucceeded:
		return "Completed", "info", summary
	case corev1.PodPending:
		return "Pending", "warning", summary
	case corev1.PodFailed:
		return "Failed", "danger", summary
	default:
		return string(pod.Status.Phase), "info", summary
	}
}

func formatServicePorts(ports []corev1.ServicePort) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.NodePort > 0 {
			parts = append(parts, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}
	}
	return strings.Join(parts, ", ")
}

func formatLabelsShort(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func formatServicePortBindings(ports []corev1.ServicePort) (string, []string) {
	if len(ports) == 0 {
		return "", nil
	}
	shortParts := make([]string, 0, len(ports))
	details := make([]string, 0, len(ports))
	for _, p := range ports {
		proto := string(p.Protocol)
		if proto == "" {
			proto = "TCP"
		}
		target := p.TargetPort.String()
		if target == "" || target == "0" {
			target = fmt.Sprintf("%d", p.Port)
		}
		if target == fmt.Sprintf("%d", p.Port) {
			shortParts = append(shortParts, fmt.Sprintf(":%d", p.Port))
			if p.NodePort > 0 {
				details = append(details, fmt.Sprintf("端口映射: %d (NodePort %d) → %s/%s", p.Port, p.NodePort, target, proto))
			} else {
				details = append(details, fmt.Sprintf("端口映射: %d/%s", p.Port, proto))
			}
		} else {
			shortParts = append(shortParts, fmt.Sprintf("%d→%s", p.Port, target))
			if p.NodePort > 0 {
				details = append(details, fmt.Sprintf("端口映射: %d (NodePort %d) → %s/%s", p.Port, p.NodePort, target, proto))
			} else {
				details = append(details, fmt.Sprintf("端口映射: %d → %s/%s", p.Port, target, proto))
			}
		}
	}
	short := shortParts[0]
	if len(shortParts) > 1 {
		short = fmt.Sprintf("%s 等%d端口", shortParts[0], len(shortParts))
	}
	return short, details
}

func formatIngressBackendPort(port networkingv1.ServiceBackendPort) string {
	if port.Number > 0 {
		return fmt.Sprintf(":%d", port.Number)
	}
	if port.Name != "" {
		return ":" + port.Name
	}
	return ""
}

func formatHPABinding(hpa autoscalingv2.HorizontalPodAutoscaler) (string, []string) {
	minR := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minR = *hpa.Spec.MinReplicas
	}
	details := []string{
		fmt.Sprintf("目标对象: %s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name),
		fmt.Sprintf("副本范围: %d ~ %d（当前 %d / 期望 %d）", minR, hpa.Spec.MaxReplicas, hpa.Status.CurrentReplicas, hpa.Status.DesiredReplicas),
	}
	metricShort := ""
	for _, m := range hpa.Spec.Metrics {
		if m.Type == autoscalingv2.ResourceMetricSourceType && m.Resource != nil {
			resName := strings.ToUpper(string(m.Resource.Name))
			if m.Resource.Target.AverageUtilization != nil {
				text := fmt.Sprintf("%s %d%%", resName, *m.Resource.Target.AverageUtilization)
				details = append(details, "扩缩指标: "+text)
				if metricShort == "" {
					metricShort = text
				}
			} else if m.Resource.Target.AverageValue != nil {
				text := fmt.Sprintf("%s %s", resName, m.Resource.Target.AverageValue.String())
				details = append(details, "扩缩指标: "+text)
				if metricShort == "" {
					metricShort = text
				}
			}
		}
	}
	label := fmt.Sprintf("副本 %d~%d", minR, hpa.Spec.MaxReplicas)
	if metricShort != "" {
		label = fmt.Sprintf("%s · %s", label, metricShort)
	}
	return label, details
}

type volumeMountBinding struct {
	short   string
	details []string
}

func collectVolumeMountBindings(spec *corev1.PodSpec) map[string]volumeMountBinding {
	res := make(map[string]volumeMountBinding)
	allContainers := make([]corev1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
	allContainers = append(allContainers, spec.InitContainers...)
	allContainers = append(allContainers, spec.Containers...)
	for _, c := range allContainers {
		for _, vm := range c.VolumeMounts {
			if vm.Name == "" || vm.MountPath == "" {
				continue
			}
			cur := res[vm.Name]
			roSuffix := ""
			modeText := "读写"
			if vm.ReadOnly {
				roSuffix = " (RO)"
				modeText = "只读"
			}
			if cur.short == "" {
				cur.short = vm.MountPath + roSuffix
			}
			detail := fmt.Sprintf("容器 %s 挂载: %s (%s)", c.Name, vm.MountPath, modeText)
			if vm.SubPath != "" {
				detail += fmt.Sprintf(" [subPath: %s]", vm.SubPath)
			}
			cur.details = append(cur.details, detail)
			res[vm.Name] = cur
		}
	}
	return res
}

// GetNamespaceTopology 并发拉取指定命名空间下的核心 K8s 资源并构建关联拓扑图
func (s *TopologyService) GetNamespaceTopology(ctx context.Context, query dto.TopologyQueryDTO) (*vo.TopologyVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	ns := strings.TrimSpace(query.Namespace)
	if ns == "" {
		return nil, errors.New("命名空间不能为空")
	}

	tCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	var (
		wg              sync.WaitGroup
		ingresses       []networkingv1.Ingress
		hpas            []autoscalingv2.HorizontalPodAutoscaler
		services        []corev1.Service
		deployments     []appsv1.Deployment
		statefulSets    []appsv1.StatefulSet
		daemonSets      []appsv1.DaemonSet
		replicaSets     []appsv1.ReplicaSet
		cronJobs        []batchv1.CronJob
		jobs            []batchv1.Job
		pods            []corev1.Pod
		configMaps      []corev1.ConfigMap
		secrets         []corev1.Secret
		pvcs            []corev1.PersistentVolumeClaim
		serviceAccounts []corev1.ServiceAccount
	)

	listOpts := metav1.ListOptions{}
	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	run(func() {
		if l, e := clientSet.NetworkingV1().Ingresses(ns).List(tCtx, listOpts); e == nil {
			ingresses = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.AutoscalingV2().HorizontalPodAutoscalers(ns).List(tCtx, listOpts); e == nil {
			hpas = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().Services(ns).List(tCtx, listOpts); e == nil {
			services = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.AppsV1().Deployments(ns).List(tCtx, listOpts); e == nil {
			deployments = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.AppsV1().StatefulSets(ns).List(tCtx, listOpts); e == nil {
			statefulSets = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.AppsV1().DaemonSets(ns).List(tCtx, listOpts); e == nil {
			daemonSets = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.AppsV1().ReplicaSets(ns).List(tCtx, listOpts); e == nil {
			replicaSets = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.BatchV1().CronJobs(ns).List(tCtx, listOpts); e == nil {
			cronJobs = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.BatchV1().Jobs(ns).List(tCtx, listOpts); e == nil {
			jobs = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().Pods(ns).List(tCtx, listOpts); e == nil {
			pods = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().ConfigMaps(ns).List(tCtx, listOpts); e == nil {
			configMaps = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().Secrets(ns).List(tCtx, listOpts); e == nil {
			secrets = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().PersistentVolumeClaims(ns).List(tCtx, listOpts); e == nil {
			pvcs = l.Items
		}
	})
	run(func() {
		if l, e := clientSet.CoreV1().ServiceAccounts(ns).List(tCtx, listOpts); e == nil {
			serviceAccounts = l.Items
		}
	})

	wg.Wait()

	nodeMap := make(map[string]vo.TopologyNodeVO)
	edgeMap := make(map[string]vo.TopologyEdgeVO)
	edgeShortLabels := make(map[string][]string)

	addNode := func(n vo.TopologyNodeVO) {
		nodeMap[n.ID] = n
	}
	appendUnique := func(list []string, item string) []string {
		item = strings.TrimSpace(item)
		if item == "" {
			return list
		}
		for _, existing := range list {
			if existing == item {
				return list
			}
		}
		return append(list, item)
	}
	addEdge := func(source, target, relation, shortLabel string, details ...string) {
		if source == "" || target == "" || source == target {
			return
		}
		key := source + "->" + target + ":" + relation
		edge, exists := edgeMap[key]
		if !exists {
			edge = vo.TopologyEdgeVO{
				ID:       key,
				Source:   source,
				Target:   target,
				Relation: relation,
			}
		}
		for _, d := range details {
			edge.Details = appendUnique(edge.Details, d)
		}
		if strings.TrimSpace(shortLabel) != "" {
			labels := appendUnique(edgeShortLabels[key], shortLabel)
			edgeShortLabels[key] = labels
			if len(labels) == 1 {
				edge.Label = labels[0]
			} else if len(labels) > 1 {
				edge.Label = fmt.Sprintf("%s 等%d项", labels[0], len(labels))
			}
		}
		edgeMap[key] = edge
	}

	// 记录每个 Pod 的 owner 以便判断活跃 ReplicaSet 及 Service 关联归属
	rsOwnedPods := make(map[string]bool)
	workloadHasPod := make(map[string]bool)
	rsToDeploy := make(map[string]string)

	for _, rs := range replicaSets {
		for _, ow := range rs.OwnerReferences {
			if ow.Kind == "Deployment" {
				rsToDeploy[rs.Name] = ow.Name
				break
			}
		}
	}
	for _, pod := range pods {
		for _, ow := range pod.OwnerReferences {
			if ow.Kind == "ReplicaSet" {
				rsOwnedPods[ow.Name] = true
				if depName, ok := rsToDeploy[ow.Name]; ok {
					workloadHasPod[topoID("Deployment", depName)] = true
				}
			} else if ow.Kind == "StatefulSet" || ow.Kind == "DaemonSet" || ow.Kind == "Job" {
				workloadHasPod[topoID(ow.Kind, ow.Name)] = true
			}
		}
	}

	// ---------------- Layer 0: Ingress & HPA ----------------
	for _, ing := range ingresses {
		hosts := make([]string, 0, len(ing.Spec.Rules))
		for _, r := range ing.Spec.Rules {
			if r.Host != "" {
				hosts = append(hosts, r.Host)
			}
		}
		summary := "默认规则"
		if len(hosts) > 0 {
			summary = strings.Join(hosts, ", ")
		}
		className := "-"
		if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName != "" {
			className = *ing.Spec.IngressClassName
		}
		id := topoID("Ingress", ing.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Ingress",
			Name:       ing.Name,
			Namespace:  ns,
			Status:     "正常",
			StatusType: "success",
			Layer:      0,
			Summary:    summary,
			Extra: map[string]string{
				"IngressClass": className,
				"域名":           summary,
			},
		})
	}

	for _, hpa := range hpas {
		minR := int32(1)
		if hpa.Spec.MinReplicas != nil {
			minR = *hpa.Spec.MinReplicas
		}
		id := topoID("HPA", hpa.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "HPA",
			Name:       hpa.Name,
			Namespace:  ns,
			Status:     "正常",
			StatusType: "success",
			Layer:      0,
			Summary:    fmt.Sprintf("副本 %d (%d-%d)", hpa.Status.CurrentReplicas, minR, hpa.Spec.MaxReplicas),
			Extra: map[string]string{
				"目标资源": fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name),
				"副本范围": fmt.Sprintf("%d ~ %d", minR, hpa.Spec.MaxReplicas),
				"当前副本": fmt.Sprintf("%d", hpa.Status.CurrentReplicas),
			},
		})
	}

	// ---------------- Layer 1: Service ----------------
	for _, svc := range services {
		id := topoID("Service", svc.Name)
		portsText := formatServicePorts(svc.Spec.Ports)
		ip := svc.Spec.ClusterIP
		if ip == "" || ip == "None" {
			ip = "Headless"
		}
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Service",
			Name:       svc.Name,
			Namespace:  ns,
			Status:     string(svc.Spec.Type),
			StatusType: "success",
			Layer:      1,
			Summary:    fmt.Sprintf("%s · %s", svc.Spec.Type, portsText),
			Extra: map[string]string{
				"类型":   string(svc.Spec.Type),
				"集群IP": ip,
				"端口":   portsText,
			},
		})
	}

	// ---------------- Layer 2: Workloads ----------------
	for _, dep := range deployments {
		wvo := vo.Deployment2WorkloadVO(dep)
		id := topoID("Deployment", dep.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Deployment",
			Name:       dep.Name,
			Namespace:  ns,
			Status:     wvo.Status,
			StatusType: workloadStatusType(wvo.Status),
			Layer:      2,
			Summary:    fmt.Sprintf("%d/%d 就绪", wvo.ReadyReplicas, wvo.Replicas),
			Extra: map[string]string{
				"副本状态": fmt.Sprintf("%d / %d", wvo.ReadyReplicas, wvo.Replicas),
				"容器数":  fmt.Sprintf("%d", wvo.Containers),
			},
		})
	}

	for _, sts := range statefulSets {
		wvo := vo.StatefulSet2WorkloadVO(sts)
		id := topoID("StatefulSet", sts.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "StatefulSet",
			Name:       sts.Name,
			Namespace:  ns,
			Status:     wvo.Status,
			StatusType: workloadStatusType(wvo.Status),
			Layer:      2,
			Summary:    fmt.Sprintf("%d/%d 就绪", wvo.ReadyReplicas, wvo.Replicas),
			Extra: map[string]string{
				"副本状态":       fmt.Sprintf("%d / %d", wvo.ReadyReplicas, wvo.Replicas),
				"Headless服务": sts.Spec.ServiceName,
			},
		})
	}

	for _, ds := range daemonSets {
		wvo := vo.DaemonSet2WorkloadVO(ds)
		id := topoID("DaemonSet", ds.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "DaemonSet",
			Name:       ds.Name,
			Namespace:  ns,
			Status:     wvo.Status,
			StatusType: workloadStatusType(wvo.Status),
			Layer:      2,
			Summary:    fmt.Sprintf("%d/%d 节点就绪", wvo.ReadyReplicas, wvo.Replicas),
			Extra: map[string]string{
				"节点就绪": fmt.Sprintf("%d / %d", wvo.ReadyReplicas, wvo.Replicas),
				"容器数":  fmt.Sprintf("%d", wvo.Containers),
			},
		})
	}

	for _, rs := range replicaSets {
		desired := int32(1)
		if rs.Spec.Replicas != nil {
			desired = *rs.Spec.Replicas
		}
		// 过滤已缩容到 0 且无活跃 Pod 的历史 ReplicaSet，避免拓扑图杂乱
		if desired == 0 && rs.Status.Replicas == 0 && !rsOwnedPods[rs.Name] {
			continue
		}
		status := "运行中"
		statusType := "success"
		if desired == 0 {
			status = "已停止"
			statusType = "info"
		} else if rs.Status.ReadyReplicas < desired {
			status = "更新中"
			statusType = "warning"
		}
		id := topoID("ReplicaSet", rs.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "ReplicaSet",
			Name:       rs.Name,
			Namespace:  ns,
			Status:     status,
			StatusType: statusType,
			Layer:      2,
			Summary:    fmt.Sprintf("%d/%d 就绪", rs.Status.ReadyReplicas, desired),
			Extra: map[string]string{
				"副本状态": fmt.Sprintf("%d / %d", rs.Status.ReadyReplicas, desired),
			},
		})
	}

	for _, cj := range cronJobs {
		status := "运行中"
		statusType := "success"
		if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
			status = "已暂停"
			statusType = "warning"
		}
		id := topoID("CronJob", cj.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "CronJob",
			Name:       cj.Name,
			Namespace:  ns,
			Status:     status,
			StatusType: statusType,
			Layer:      2,
			Summary:    cj.Spec.Schedule,
			Extra: map[string]string{
				"调度表达式": cj.Spec.Schedule,
				"活跃任务":  fmt.Sprintf("%d", len(cj.Status.Active)),
			},
		})
	}

	for _, job := range jobs {
		status := "运行中"
		statusType := "warning"
		if job.Status.Failed > 0 && job.Status.Active == 0 {
			status = "失败"
			statusType = "danger"
		} else if job.Status.Succeeded > 0 && job.Status.Active == 0 {
			status = "已完成"
			statusType = "success"
		}
		completions := int32(1)
		if job.Spec.Completions != nil {
			completions = *job.Spec.Completions
		}
		id := topoID("Job", job.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Job",
			Name:       job.Name,
			Namespace:  ns,
			Status:     status,
			StatusType: statusType,
			Layer:      2,
			Summary:    fmt.Sprintf("完成 %d/%d", job.Status.Succeeded, completions),
			Extra: map[string]string{
				"完成数": fmt.Sprintf("%d / %d", job.Status.Succeeded, completions),
				"失败数": fmt.Sprintf("%d", job.Status.Failed),
			},
		})
	}

	// ---------------- Layer 3: Pod ----------------
	for i := range pods {
		pod := &pods[i]
		status, statusType, summary := podTopoStatus(pod)
		id := topoID("Pod", pod.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Pod",
			Name:       pod.Name,
			Namespace:  ns,
			Status:     status,
			StatusType: statusType,
			Layer:      3,
			Summary:    summary,
			Extra: map[string]string{
				"节点":    pod.Spec.NodeName,
				"PodIP": pod.Status.PodIP,
				"就绪容器":  summary,
			},
		})
	}

	// ---------------- Layer 4: ConfigMap / Secret / PVC / ServiceAccount ----------------
	for _, cm := range configMaps {
		id := topoID("ConfigMap", cm.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "ConfigMap",
			Name:       cm.Name,
			Namespace:  ns,
			Status:     "配置项",
			StatusType: "info",
			Layer:      4,
			Summary:    fmt.Sprintf("%d 项配置", len(cm.Data)+len(cm.BinaryData)),
			Extra: map[string]string{
				"键值数量": fmt.Sprintf("%d", len(cm.Data)+len(cm.BinaryData)),
			},
		})
	}

	for _, sec := range secrets {
		// 过滤 Helm 内部发布记录 Secret，避免拓扑图被大量发布快照淹没
		if string(sec.Type) == "helm.sh/release.v1" {
			continue
		}
		id := topoID("Secret", sec.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "Secret",
			Name:       sec.Name,
			Namespace:  ns,
			Status:     "密钥",
			StatusType: "info",
			Layer:      4,
			Summary:    fmt.Sprintf("%d 项 · %s", len(sec.Data), sec.Type),
			Extra: map[string]string{
				"类型":  string(sec.Type),
				"数据项": fmt.Sprintf("%d", len(sec.Data)),
			},
		})
	}

	for _, pvc := range pvcs {
		capText := "-"
		if q, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			capText = q.String()
		} else if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			capText = q.String()
		}
		statusType := "warning"
		if pvc.Status.Phase == corev1.ClaimBound {
			statusType = "success"
		} else if pvc.Status.Phase == corev1.ClaimLost {
			statusType = "danger"
		}
		sc := "-"
		if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName != "" {
			sc = *pvc.Spec.StorageClassName
		}
		id := topoID("PVC", pvc.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "PVC",
			Name:       pvc.Name,
			Namespace:  ns,
			Status:     string(pvc.Status.Phase),
			StatusType: statusType,
			Layer:      4,
			Summary:    fmt.Sprintf("%s · %s", capText, pvc.Status.Phase),
			Extra: map[string]string{
				"容量":   capText,
				"存储类":  sc,
				"绑定PV": pvc.Spec.VolumeName,
			},
		})
	}

	for _, sa := range serviceAccounts {
		id := topoID("ServiceAccount", sa.Name)
		addNode(vo.TopologyNodeVO{
			ID:         id,
			Kind:       "ServiceAccount",
			Name:       sa.Name,
			Namespace:  ns,
			Status:     "服务账户",
			StatusType: "info",
			Layer:      4,
			Summary:    "身份账户",
			Extra: map[string]string{
				"Secrets": fmt.Sprintf("%d", len(sa.Secrets)),
			},
		})
	}

	// 缺失依赖占位节点辅助函数（例如 Pod 引用了不存在的 ConfigMap/Secret/PVC）
	ensureTargetNode := func(kind, name string, layer int) string {
		id := topoID(kind, name)
		if _, ok := nodeMap[id]; !ok {
			addNode(vo.TopologyNodeVO{
				ID:         id,
				Kind:       kind,
				Name:       name,
				Namespace:  ns,
				Status:     "未找到",
				StatusType: "danger",
				Layer:      layer,
				Summary:    "引用的资源不存在",
			})
		}
		return id
	}

	// ---------------- 构建关联关系边（含具体绑定信息） ----------------

	// 1. Ingress -> Service / TLS Secret
	for _, ing := range ingresses {
		src := topoID("Ingress", ing.Name)
		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil && ing.Spec.DefaultBackend.Service.Name != "" {
			svcRef := ing.Spec.DefaultBackend.Service
			portText := formatIngressBackendPort(svcRef.Port)
			dst := ensureTargetNode("Service", svcRef.Name, 1)
			label := "默认后端" + portText
			addEdge(src, dst, "路由", label, fmt.Sprintf("默认路由 → %s%s", svcRef.Name, portText))
		}
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, p := range rule.HTTP.Paths {
				if p.Backend.Service != nil && p.Backend.Service.Name != "" {
					svcRef := p.Backend.Service
					portText := formatIngressBackendPort(svcRef.Port)
					pathStr := p.Path
					if pathStr == "" {
						pathStr = "/"
					}
					routePrefix := pathStr
					if rule.Host != "" {
						routePrefix = rule.Host + pathStr
					}
					pathType := ""
					if p.PathType != nil {
						pathType = string(*p.PathType)
					}
					shortLabel := routePrefix
					if portText != "" {
						shortLabel = fmt.Sprintf("%s → %s", routePrefix, portText)
					}
					detail := fmt.Sprintf("路由规则: %s → %s%s", routePrefix, svcRef.Name, portText)
					if pathType != "" {
						detail += fmt.Sprintf(" (%s)", pathType)
					}
					dst := ensureTargetNode("Service", svcRef.Name, 1)
					addEdge(src, dst, "路由", shortLabel, detail)
				}
			}
		}
		for _, tls := range ing.Spec.TLS {
			if tls.SecretName != "" {
				dst := ensureTargetNode("Secret", tls.SecretName, 4)
				hostsText := strings.Join(tls.Hosts, ", ")
				shortLabel := "TLS 证书"
				detail := fmt.Sprintf("TLS 证书 Secret: %s", tls.SecretName)
				if hostsText != "" {
					shortLabel = hostsText
					detail = fmt.Sprintf("TLS 域名: %s (Secret: %s)", hostsText, tls.SecretName)
				}
				addEdge(src, dst, "TLS", shortLabel, detail)
			}
		}
	}

	// 2. HPA -> Target Workload
	for _, hpa := range hpas {
		src := topoID("HPA", hpa.Name)
		targetKind := hpa.Spec.ScaleTargetRef.Kind
		targetName := hpa.Spec.ScaleTargetRef.Name
		if targetKind != "" && targetName != "" {
			dst := topoID(targetKind, targetName)
			if _, exists := nodeMap[dst]; exists {
				label, details := formatHPABinding(hpa)
				addEdge(src, dst, "伸缩", label, details...)
			}
		}
	}

	// 3. OwnerReferences: Deployment -> ReplicaSet, CronJob -> Job, Workload/RS/Job -> Pod
	cronJobMap := make(map[string]batchv1.CronJob, len(cronJobs))
	for _, cj := range cronJobs {
		cronJobMap[cj.Name] = cj
	}

	for _, rs := range replicaSets {
		rsID := topoID("ReplicaSet", rs.Name)
		if _, ok := nodeMap[rsID]; !ok {
			continue
		}
		desired := int32(1)
		if rs.Spec.Replicas != nil {
			desired = *rs.Spec.Replicas
		}
		rev := rs.Annotations["deployment.kubernetes.io/revision"]
		shortLabel := fmt.Sprintf("副本 %d/%d", rs.Status.ReadyReplicas, desired)
		if rev != "" {
			shortLabel = fmt.Sprintf("rev#%s · %d/%d", rev, rs.Status.ReadyReplicas, desired)
		}
		details := []string{
			fmt.Sprintf("副本状态: %d/%d 就绪", rs.Status.ReadyReplicas, desired),
		}
		if rev != "" {
			details = append(details, fmt.Sprintf("发布版本: Revision #%s", rev))
		}
		if rs.Spec.Selector != nil && len(rs.Spec.Selector.MatchLabels) > 0 {
			details = append(details, "标签选择器: "+formatLabelsShort(rs.Spec.Selector.MatchLabels))
		}
		for _, ow := range rs.OwnerReferences {
			if ow.Kind == "Deployment" {
				depID := topoID("Deployment", ow.Name)
				if _, exists := nodeMap[depID]; exists {
					addEdge(depID, rsID, "管理", shortLabel, append([]string{fmt.Sprintf("OwnerReference: Deployment/%s", ow.Name)}, details...)...)
				}
			}
		}
	}

	for _, job := range jobs {
		jobID := topoID("Job", job.Name)
		for _, ow := range job.OwnerReferences {
			if ow.Kind == "CronJob" {
				cjID := topoID("CronJob", ow.Name)
				if _, exists := nodeMap[cjID]; exists {
					shortLabel := "定时调度"
					details := []string{fmt.Sprintf("OwnerReference: CronJob/%s", ow.Name)}
					if cj, ok := cronJobMap[ow.Name]; ok && cj.Spec.Schedule != "" {
						shortLabel = cj.Spec.Schedule
						details = append(details, "调度表达式: "+cj.Spec.Schedule)
					}
					addEdge(cjID, jobID, "调度", shortLabel, details...)
				}
			}
		}
	}

	// 记录 Pod 归属的顶层工作负载，供 Service 智能连线避免重复蜘蛛网
	podTopOwner := make(map[string]string)
	for i := range pods {
		pod := &pods[i]
		podID := topoID("Pod", pod.Name)
		shortLabel := "Pod 实例"
		if pod.Spec.NodeName != "" {
			shortLabel = "节点 " + pod.Spec.NodeName
		} else if pod.Status.PodIP != "" {
			shortLabel = pod.Status.PodIP
		}
		baseDetails := make([]string, 0, 4)
		if pod.Spec.NodeName != "" {
			baseDetails = append(baseDetails, "调度节点: "+pod.Spec.NodeName)
		}
		if pod.Status.PodIP != "" {
			baseDetails = append(baseDetails, "Pod IP: "+pod.Status.PodIP)
		}
		if len(pod.Spec.Containers) > 0 {
			cNames := make([]string, 0, len(pod.Spec.Containers))
			for _, c := range pod.Spec.Containers {
				cNames = append(cNames, c.Name)
			}
			baseDetails = append(baseDetails, "容器列表: "+strings.Join(cNames, ", "))
		}
		for _, ow := range pod.OwnerReferences {
			ownerID := topoID(ow.Kind, ow.Name)
			if _, exists := nodeMap[ownerID]; exists {
				details := append([]string{fmt.Sprintf("OwnerReference: %s/%s", ow.Kind, ow.Name)}, baseDetails...)
				addEdge(ownerID, podID, "管理", shortLabel, details...)
				if ow.Kind == "ReplicaSet" {
					if depName, ok := rsToDeploy[ow.Name]; ok {
						podTopOwner[podID] = topoID("Deployment", depName)
					} else {
						podTopOwner[podID] = ownerID
					}
				} else {
					podTopOwner[podID] = ownerID
				}
			}
		}
	}

	// 4. Service -> Workload（优先匹配工作负载模板或 Pod 所属工作负载，无控制器管理的独立 Pod 直接连 Pod）
	for _, svc := range services {
		svcID := topoID("Service", svc.Name)
		if len(svc.Spec.Selector) > 0 {
			portShort, portDetails := formatServicePortBindings(svc.Spec.Ports)
			selText := formatLabelsShort(svc.Spec.Selector)
			shortLabel := portShort
			if shortLabel == "" {
				shortLabel = selText
			}
			details := make([]string, 0, len(portDetails)+2)
			details = append(details, portDetails...)
			if selText != "" {
				details = append(details, "标签匹配 (Selector): "+selText)
			}
			if svc.Spec.ClusterIP != "" {
				details = append(details, fmt.Sprintf("服务类型: %s (%s)", svc.Spec.Type, svc.Spec.ClusterIP))
			}

			for _, dep := range deployments {
				if labelsMatch(svc.Spec.Selector, dep.Spec.Template.Labels) {
					addEdge(svcID, topoID("Deployment", dep.Name), "转发", shortLabel, details...)
				}
			}
			for _, sts := range statefulSets {
				if labelsMatch(svc.Spec.Selector, sts.Spec.Template.Labels) {
					addEdge(svcID, topoID("StatefulSet", sts.Name), "转发", shortLabel, details...)
				}
			}
			for _, ds := range daemonSets {
				if labelsMatch(svc.Spec.Selector, ds.Spec.Template.Labels) {
					addEdge(svcID, topoID("DaemonSet", ds.Name), "转发", shortLabel, details...)
				}
			}
			for i := range pods {
				pod := &pods[i]
				if labelsMatch(svc.Spec.Selector, pod.Labels) {
					podID := topoID("Pod", pod.Name)
					if topOwner, ok := podTopOwner[podID]; ok {
						addEdge(svcID, topOwner, "转发", shortLabel, details...)
					} else {
						addEdge(svcID, podID, "转发", shortLabel, details...)
					}
				}
			}
		}
	}

	// StatefulSet -> Headless Service
	svcMap := make(map[string]corev1.Service, len(services))
	for _, svc := range services {
		svcMap[svc.Name] = svc
	}
	for _, sts := range statefulSets {
		if sts.Spec.ServiceName != "" {
			svcID := topoID("Service", sts.Spec.ServiceName)
			if _, exists := nodeMap[svcID]; exists {
				shortLabel := "serviceName: " + sts.Spec.ServiceName
				details := []string{"StatefulSet serviceName: " + sts.Spec.ServiceName}
				if svc, ok := svcMap[sts.Spec.ServiceName]; ok {
					portShort, portDetails := formatServicePortBindings(svc.Spec.Ports)
					if portShort != "" {
						shortLabel = portShort
					}
					details = append(details, portDetails...)
				}
				addEdge(svcID, topoID("StatefulSet", sts.Name), "Headless", shortLabel, details...)
			}
		}
	}

	// 5. 提取 PodSpec 依赖（ConfigMap / Secret / PVC / ServiceAccount）
	linkPodSpecDeps := func(srcID string, spec *corev1.PodSpec) {
		mountBindings := collectVolumeMountBindings(spec)
		for _, vol := range spec.Volumes {
			mb := mountBindings[vol.Name]
			volShort := mb.short
			if volShort == "" {
				volShort = "卷 " + vol.Name
			}
			volDetails := append([]string{"卷名 (Volume): " + vol.Name}, mb.details...)

			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName != "" {
				dst := ensureTargetNode("PVC", vol.PersistentVolumeClaim.ClaimName, 4)
				details := append([]string{"声明 PVC: " + vol.PersistentVolumeClaim.ClaimName}, volDetails...)
				addEdge(srcID, dst, "挂载", volShort, details...)
			}
			if vol.ConfigMap != nil && vol.ConfigMap.Name != "" {
				dst := ensureTargetNode("ConfigMap", vol.ConfigMap.Name, 4)
				addEdge(srcID, dst, "挂载", volShort, volDetails...)
			}
			if vol.Secret != nil && vol.Secret.SecretName != "" {
				dst := ensureTargetNode("Secret", vol.Secret.SecretName, 4)
				addEdge(srcID, dst, "挂载", volShort, volDetails...)
			}
			if vol.Projected != nil {
				projDetails := append([]string{"Projected 卷: " + vol.Name}, mb.details...)
				for _, ps := range vol.Projected.Sources {
					if ps.ConfigMap != nil && ps.ConfigMap.Name != "" && ps.ConfigMap.Name != "kube-root-ca.crt" {
						dst := ensureTargetNode("ConfigMap", ps.ConfigMap.Name, 4)
						addEdge(srcID, dst, "挂载", volShort, projDetails...)
					}
					if ps.Secret != nil && ps.Secret.Name != "" {
						dst := ensureTargetNode("Secret", ps.Secret.Name, 4)
						addEdge(srcID, dst, "挂载", volShort, projDetails...)
					}
				}
			}
		}

		allContainers := make([]corev1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
		allContainers = append(allContainers, spec.InitContainers...)
		allContainers = append(allContainers, spec.Containers...)
		for _, c := range allContainers {
			for _, ef := range c.EnvFrom {
				prefixNote := ""
				if ef.Prefix != "" {
					prefixNote = fmt.Sprintf(" (前缀 %s)", ef.Prefix)
				}
				if ef.ConfigMapRef != nil && ef.ConfigMapRef.Name != "" {
					dst := ensureTargetNode("ConfigMap", ef.ConfigMapRef.Name, 4)
					addEdge(srcID, dst, "引用", "envFrom 注入", fmt.Sprintf("容器 %s: envFrom 全量注入%s", c.Name, prefixNote))
				}
				if ef.SecretRef != nil && ef.SecretRef.Name != "" {
					dst := ensureTargetNode("Secret", ef.SecretRef.Name, 4)
					addEdge(srcID, dst, "引用", "envFrom 注入", fmt.Sprintf("容器 %s: envFrom 全量注入%s", c.Name, prefixNote))
				}
			}
			for _, env := range c.Env {
				if env.ValueFrom == nil {
					continue
				}
				if env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" {
					ref := env.ValueFrom.ConfigMapKeyRef
					dst := ensureTargetNode("ConfigMap", ref.Name, 4)
					addEdge(srcID, dst, "引用", fmt.Sprintf("env:%s", env.Name), fmt.Sprintf("容器 %s: ENV %s ← key %s", c.Name, env.Name, ref.Key))
				}
				if env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" {
					ref := env.ValueFrom.SecretKeyRef
					dst := ensureTargetNode("Secret", ref.Name, 4)
					addEdge(srcID, dst, "引用", fmt.Sprintf("env:%s", env.Name), fmt.Sprintf("容器 %s: ENV %s ← key %s", c.Name, env.Name, ref.Key))
				}
			}
		}

		for _, ips := range spec.ImagePullSecrets {
			if ips.Name != "" {
				dst := ensureTargetNode("Secret", ips.Name, 4)
				addEdge(srcID, dst, "拉取凭证", "imagePullSecrets", fmt.Sprintf("镜像拉取凭证 (imagePullSecrets): %s", ips.Name))
			}
		}

		// 仅关联非 default 的自定义 ServiceAccount，避免所有 Pod 都指向 default 造成连线拥挤
		if spec.ServiceAccountName != "" && spec.ServiceAccountName != "default" {
			dst := ensureTargetNode("ServiceAccount", spec.ServiceAccountName, 4)
			addEdge(srcID, dst, "身份", "SA: "+spec.ServiceAccountName, fmt.Sprintf("serviceAccountName: %s", spec.ServiceAccountName))
		}
	}

	for i := range pods {
		linkPodSpecDeps(topoID("Pod", pods[i].Name), &pods[i].Spec)
	}
	// 当工作负载当前无活跃 Pod（如副本为 0）时，从模板提取配置/存储关联，依然能看到完整依赖
	for i := range deployments {
		id := topoID("Deployment", deployments[i].Name)
		if !workloadHasPod[id] {
			linkPodSpecDeps(id, &deployments[i].Spec.Template.Spec)
		}
	}
	for i := range statefulSets {
		id := topoID("StatefulSet", statefulSets[i].Name)
		if !workloadHasPod[id] {
			linkPodSpecDeps(id, &statefulSets[i].Spec.Template.Spec)
		}
	}
	for i := range daemonSets {
		id := topoID("DaemonSet", daemonSets[i].Name)
		if !workloadHasPod[id] {
			linkPodSpecDeps(id, &daemonSets[i].Spec.Template.Spec)
		}
	}

	nodes := make([]vo.TopologyNodeVO, 0, len(nodeMap))
	summary := make(map[string]int)
	for _, n := range nodeMap {
		nodes = append(nodes, n)
		summary[n.Kind]++
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Layer != nodes[j].Layer {
			return nodes[i].Layer < nodes[j].Layer
		}
		if nodes[i].Kind != nodes[j].Kind {
			return nodes[i].Kind < nodes[j].Kind
		}
		return nodes[i].Name < nodes[j].Name
	})

	edges := make([]vo.TopologyEdgeVO, 0, len(edgeMap))
	for _, e := range edgeMap {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].ID < edges[j].ID
	})

	return &vo.TopologyVO{
		Namespace: ns,
		Nodes:     nodes,
		Edges:     edges,
		Summary:   summary,
	}, nil
}

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

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	mirrorPodAnnotation = "kubernetes.io/config.mirror"
	defaultDrainTimeout = 120
	maxDrainTimeout     = 600
	drainConcurrency    = 10
	drainRetryInterval  = 5 * time.Second
)

// kubeletLabels 由 kubelet / 云厂商组件维护的节点标签, 手动修改会被覆盖, 编辑时保留原值
var kubeletLabels = map[string]bool{
	"kubernetes.io/hostname":                   true,
	"kubernetes.io/os":                         true,
	"kubernetes.io/arch":                       true,
	"beta.kubernetes.io/os":                    true,
	"beta.kubernetes.io/arch":                  true,
	"node.kubernetes.io/instance-type":         true,
	"beta.kubernetes.io/instance-type":         true,
	"topology.kubernetes.io/region":            true,
	"topology.kubernetes.io/zone":              true,
	"failure-domain.beta.kubernetes.io/region": true,
	"failure-domain.beta.kubernetes.io/zone":   true,
}

// IsKubeletLabel 由系统维护的节点标签
func IsKubeletLabel(key string) bool {
	return kubeletLabels[key] || strings.HasPrefix(key, "feature.node.kubernetes.io/")
}

// ---------- 标签 / 污点 ----------

// updateNode 读取最新的节点后修改并保存, 遇到版本冲突时自动重试(节点状态由 kubelet 频繁更新)
func (s *NodeService) updateNode(ctx context.Context, query dto.NodeInfoQueryDTO, mutate func(node *corev1.Node) error) error {
	cs, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node, getErr := cs.CoreV1().Nodes().Get(ctx, query.Hostname, metav1.GetOptions{})
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return fmt.Errorf("节点 %s 不存在", query.Hostname)
			}
			return getErr
		}
		if mErr := mutate(node); mErr != nil {
			return mErr
		}
		_, updErr := cs.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return updErr
	})
}

// UpdateLabels 设置节点标签; 系统维护的标签保留原值
func (s *NodeService) UpdateLabels(ctx context.Context, req dto.NodeLabelsDTO) error {
	labels := cleanStringMap(req.Labels)
	if err := validateLabelMap(labels, "标签"); err != nil {
		return err
	}
	err := s.updateNode(ctx, req.NodeInfoQueryDTO, func(node *corev1.Node) error {
		next := map[string]string{}
		for k, v := range node.Labels {
			if IsKubeletLabel(k) {
				next[k] = v
			}
		}
		for k, v := range labels {
			if !IsKubeletLabel(k) {
				next[k] = v
			}
		}
		node.Labels = next
		return nil
	})
	if err != nil {
		logs.Error("更新节点标签失败: %s %s %v", req.ClusterId, req.Hostname, err)
		return fmt.Errorf("更新节点标签失败: %v", err)
	}
	return nil
}

var validTaintEffects = map[corev1.TaintEffect]bool{
	corev1.TaintEffectNoSchedule:       true,
	corev1.TaintEffectPreferNoSchedule: true,
	corev1.TaintEffectNoExecute:        true,
}

// UpdateTaints 设置节点污点; 节点控制器维护的污点保留原值
func (s *NodeService) UpdateTaints(ctx context.Context, req dto.NodeTaintsDTO) error {
	taints := make([]corev1.Taint, 0, len(req.Taints))
	seen := map[string]bool{}
	for i, t := range req.Taints {
		key := strings.TrimSpace(t.Key)
		if key == "" {
			continue
		}
		label := fmt.Sprintf("第 %d 个污点", i+1)
		if errs := validation.IsQualifiedName(key); len(errs) > 0 {
			return fmt.Errorf("%s的键 %s 不合法: %s", label, key, strings.Join(errs, "; "))
		}
		value := strings.TrimSpace(t.Value)
		if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
			return fmt.Errorf("%s的值 %s 不合法: 最多 63 个字符，只能包含字母、数字、\"-\"、\"_\" 和 \".\"", label, value)
		}
		effect := corev1.TaintEffect(strings.TrimSpace(t.Effect))
		if !validTaintEffects[effect] {
			return fmt.Errorf("%s的效果不正确: %s", label, t.Effect)
		}
		if vo.IsSystemTaint(key) {
			continue
		}
		id := key + ":" + string(effect)
		if seen[id] {
			return fmt.Errorf("污点 %s 重复(同一个键和效果只能有一个)", id)
		}
		seen[id] = true
		taints = append(taints, corev1.Taint{Key: key, Value: value, Effect: effect})
	}
	err := s.updateNode(ctx, req.NodeInfoQueryDTO, func(node *corev1.Node) error {
		next := make([]corev1.Taint, 0, len(node.Spec.Taints)+len(taints))
		old := map[string]*metav1.Time{}
		for _, t := range node.Spec.Taints {
			if vo.IsSystemTaint(t.Key) {
				next = append(next, t)
			} else {
				old[t.Key+":"+string(t.Effect)] = t.TimeAdded
			}
		}
		now := metav1.Now()
		for _, t := range taints {
			// NoExecute 污点按添加时间计算容忍时长, 已有的保留原时间
			if t.Effect == corev1.TaintEffectNoExecute {
				if added := old[t.Key+":"+string(t.Effect)]; added != nil {
					t.TimeAdded = added
				} else {
					t.TimeAdded = &now
				}
			}
			next = append(next, t)
		}
		node.Spec.Taints = next
		return nil
	})
	if err != nil {
		logs.Error("更新节点污点失败: %s %s %v", req.ClusterId, req.Hostname, err)
		return fmt.Errorf("更新节点污点失败: %v", err)
	}
	return nil
}

// ---------- 排空 ----------

// classifyDrainPod 与 kubectl drain 的规则一致: DaemonSet 和静态容器组跳过, 使用 emptyDir 或没有控制器的需要确认
func classifyDrainPod(pod *corev1.Pod, req dto.NodeDrainDTO) vo.DrainPodVO {
	ownerKind, ownerName := vo.PodOwner(pod)
	status, _ := podStatusForDrain(pod)
	item := vo.DrainPodVO{Namespace: pod.Namespace, Name: pod.Name, OwnerKind: ownerKind, OwnerName: ownerName, Status: status}
	if _, ok := pod.Annotations[mirrorPodAnnotation]; ok {
		item.Action, item.Reason = "skip", "静态容器组，由 kubelet 管理"
		return item
	}
	if pod.DeletionTimestamp != nil {
		item.Action, item.Reason = "skip", "正在删除"
		return item
	}
	if ownerKind == "DaemonSet" {
		item.Action, item.Reason = "skip", "DaemonSet 容器组，会在节点上自动重建，跳过"
		return item
	}
	if !vo.IsPodActive(pod) {
		item.Action, item.Reason = "delete", "已结束，直接删除"
		return item
	}
	reasons := make([]string, 0, 2)
	blocked := false
	if ownerKind == "" {
		if req.Force {
			reasons = append(reasons, "没有控制器管理，驱逐后不会重建")
		} else {
			blocked = true
			reasons = append(reasons, "没有控制器管理，驱逐后不会重建，需要勾选「驱逐单独创建的容器组」")
		}
	}
	for _, v := range pod.Spec.Volumes {
		if v.EmptyDir != nil {
			if req.DeleteEmptyDirData {
				reasons = append(reasons, "临时存储（emptyDir）中的数据会丢失")
			} else {
				blocked = true
				reasons = append(reasons, "使用临时存储（emptyDir），需要勾选「删除临时存储数据」")
			}
			break
		}
	}
	item.Action = "evict"
	if blocked {
		item.Action = "block"
	}
	item.Reason = strings.Join(reasons, "；")
	return item
}

func podStatusForDrain(pod *corev1.Pod) (string, string) {
	row := vo.Pod2VO(pod)
	return row.DisplayStatus, row.Message
}

func (s *NodeService) drainPods(ctx context.Context, cs *kubernetes.Clientset, req dto.NodeDrainDTO) ([]corev1.Pod, []vo.DrainPodVO, error) {
	pods, err := cs.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + req.Hostname})
	if err != nil {
		return nil, nil, fmt.Errorf("获取节点上的容器组失败: %v", err)
	}
	items := make([]vo.DrainPodVO, 0, len(pods.Items))
	for i := range pods.Items {
		items = append(items, classifyDrainPod(&pods.Items[i], req))
	}
	return pods.Items, items, nil
}

func sortDrainPods(items []vo.DrainPodVO) {
	order := map[string]int{"block": 0, "evict": 1, "delete": 2, "skip": 3}
	sort.SliceStable(items, func(i, j int) bool {
		if order[items[i].Action] != order[items[j].Action] {
			return order[items[i].Action] < order[items[j].Action]
		}
		return items[i].Namespace+"/"+items[i].Name < items[j].Namespace+"/"+items[j].Name
	})
}

// DrainPreview 排空前预览每个容器组的处理方式
func (s *NodeService) DrainPreview(ctx context.Context, req dto.NodeDrainDTO) (*vo.DrainPreviewVO, error) {
	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return nil, err
	}
	_, items, err := s.drainPods(ctx, cs, req)
	if err != nil {
		return nil, err
	}
	sortDrainPods(items)
	result := &vo.DrainPreviewVO{Pods: items}
	for _, item := range items {
		switch item.Action {
		case "block":
			result.Blocked++
		case "skip":
			result.Skip++
		default:
			result.Evict++
		}
	}
	return result, nil
}

// Drain 排空节点: 停止调度后通过 Eviction API 驱逐容器组(遵守 PodDisruptionBudget), 与 kubectl drain 相同.
// 有被阻止的容器组时不做任何修改; 受 PodDisruptionBudget 限制时在超时时间内重试
func (s *NodeService) Drain(ctx context.Context, req dto.NodeDrainDTO, operator string) (*vo.DrainResultVO, error) {
	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return nil, err
	}
	pods, items, err := s.drainPods(ctx, cs, req)
	if err != nil {
		return nil, err
	}
	blocked := make([]string, 0)
	for _, item := range items {
		if item.Action == "block" {
			blocked = append(blocked, item.Namespace+"/"+item.Name)
		}
	}
	if len(blocked) > 0 {
		if len(blocked) > 5 {
			blocked = append(blocked[:5], fmt.Sprintf("等 %d 个", len(blocked)))
		}
		return nil, fmt.Errorf("以下容器组不能直接驱逐: %s，请勾选对应的选项后重试", strings.Join(blocked, "、"))
	}

	result := &vo.DrainResultVO{Evicted: []string{}, Failed: []vo.DrainPodVO{}}
	err = s.updateNode(ctx, req.NodeInfoQueryDTO, func(node *corev1.Node) error {
		result.Cordoned = !node.Spec.Unschedulable
		node.Spec.Unschedulable = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("停止调度失败: %v", err)
	}
	logs.Info("用户 %s 开始排空节点 %s/%s", operator, req.ClusterId, req.Hostname)

	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultDrainTimeout
	}
	if timeout > maxDrainTimeout {
		timeout = maxDrainTimeout
	}
	drainCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, drainConcurrency)
	)
	for i := range pods {
		item := items[i]
		if item.Action == "skip" {
			result.Skipped++
			continue
		}
		pod := &pods[i]
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			evictErr := evictPod(drainCtx, cs, pod, item.Action == "delete", req.GracePeriodSeconds)
			mu.Lock()
			defer mu.Unlock()
			if evictErr != nil {
				item.Reason = evictErr.Error()
				result.Failed = append(result.Failed, item)
				return
			}
			result.Evicted = append(result.Evicted, pod.Namespace+"/"+pod.Name)
		}()
	}
	wg.Wait()
	sort.Strings(result.Evicted)
	logs.Info("节点 %s/%s 排空完成: 驱逐 %d 个, 失败 %d 个, 跳过 %d 个", req.ClusterId, req.Hostname, len(result.Evicted), len(result.Failed), result.Skipped)
	return result, nil
}

// evictPod 驱逐一个容器组; 被 PodDisruptionBudget 拒绝(429)时等待后重试, 直到超时
func evictPod(ctx context.Context, cs *kubernetes.Clientset, pod *corev1.Pod, deleteDirectly bool, gracePeriod int64) error {
	opts := metav1.DeleteOptions{}
	if gracePeriod >= 0 {
		g := gracePeriod
		opts.GracePeriodSeconds = &g
	}
	for {
		var err error
		if deleteDirectly {
			err = cs.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, opts)
		} else {
			o := opts
			err = cs.PolicyV1().Evictions(pod.Namespace).Evict(ctx, &policyv1.Eviction{
				ObjectMeta:    metav1.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
				DeleteOptions: &o,
			})
		}
		switch {
		case err == nil, apierrors.IsNotFound(err):
			return nil
		case apierrors.IsTooManyRequests(err):
			select {
			case <-ctx.Done():
				return errors.New("受 PodDisruptionBudget 限制，超时前未能驱逐")
			case <-time.After(drainRetryInterval):
			}
		case ctx.Err() != nil:
			return errors.New("超时")
		default:
			return err
		}
	}
}

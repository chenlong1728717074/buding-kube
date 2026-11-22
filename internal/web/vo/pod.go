package vo

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// PodVO Pod 列表项. 不包含 YAML, YAML 按需加载
type PodVO struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Namespace         string    `json:"namespace"`
	Name              string    `json:"name"`
	Alias             string    `json:"alias"`
	Describe          string    `json:"describe"`
	// Status 阶段(Pending / Running / Succeeded / Failed / Unknown)
	Status string `json:"status"`
	// DisplayStatus 与 kubectl get pod 的 STATUS 列一致, 例如 CrashLoopBackOff
	DisplayStatus string `json:"displayStatus"`
	Message       string `json:"message"`
	Ready         string `json:"ready"`
	Restarts      int32  `json:"restarts"`
	NodeName      string `json:"nodeName"`
	HostIP        string `json:"hostIP"`
	PodIP         string `json:"podIP"`
	// OwnerKind / OwnerName 所属工作负载; ReplicaSet 创建的显示为所属的 Deployment
	OwnerKind string `json:"ownerKind"`
	OwnerName string `json:"ownerName"`
	PodResources
	PodUsage
}

// PodContainerStatusVO 容器运行状态
type PodContainerStatusVO struct {
	Name         string `json:"name"`
	Init         bool   `json:"init"`
	Ready        bool   `json:"ready"`
	Started      bool   `json:"started"`
	RestartCount int32  `json:"restartCount"`
	// State Running / Waiting / Terminated
	State      string     `json:"state"`
	Reason     string     `json:"reason"`
	Message    string     `json:"message"`
	ExitCode   int32      `json:"exitCode"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	// 上一次退出的信息, 排查容器反复重启时使用
	LastReason     string     `json:"lastReason"`
	LastExitCode   int32      `json:"lastExitCode"`
	LastFinishedAt *time.Time `json:"lastFinishedAt"`
	ImageID        string     `json:"imageID"`
}

// PodDetailVO 容器组详情
type PodDetailVO struct {
	BaseVO
	Status             string                 `json:"status"`
	DisplayStatus      string                 `json:"displayStatus"`
	Message            string                 `json:"message"`
	Reason             string                 `json:"reason"`
	Ready              string                 `json:"ready"`
	Restarts           int32                  `json:"restarts"`
	NodeName           string                 `json:"nodeName"`
	HostIP             string                 `json:"hostIP"`
	PodIP              string                 `json:"podIP"`
	PodIPs             []string               `json:"podIPs"`
	QosClass           string                 `json:"qosClass"`
	RestartPolicy      string                 `json:"restartPolicy"`
	ServiceAccountName string                 `json:"serviceAccountName"`
	PriorityClassName  string                 `json:"priorityClassName"`
	DNSPolicy          string                 `json:"dnsPolicy"`
	HostNetwork        bool                   `json:"hostNetwork"`
	OwnerKind          string                 `json:"ownerKind"`
	OwnerName          string                 `json:"ownerName"`
	Creator            string                 `json:"creator"`
	Uid                string                 `json:"uid"`
	StartTime          *time.Time             `json:"startTime"`
	CreateTime         time.Time              `json:"createTime"`
	DeletionTime       *time.Time             `json:"deletionTime"`
	Labels             map[string]string      `json:"labels"`
	Annotations        map[string]string      `json:"annotations"`
	NodeSelector       map[string]string      `json:"nodeSelector"`
	Tolerations        []string               `json:"tolerations"`
	Conditions         []WorkloadConditionVO  `json:"conditions"`
	Containers         []ContainerTemplateVO  `json:"containers"`
	ContainerStatuses  []PodContainerStatusVO `json:"containerStatuses"`
	Volumes            []WorkloadVolumeVO     `json:"volumes"`
	Events             []EventVO              `json:"events"`
	PodResources
	PodUsage
	// ContainerUsages 每个容器的资源设置与用量
	ContainerUsages []PodContainerUsageVO `json:"containerUsages"`
}

// PodOwner 所属的控制器; ReplicaSet 按 pod-template-hash 推断出 Deployment 名称
func PodOwner(pod *corev1.Pod) (string, string) {
	for _, ref := range pod.OwnerReferences {
		if ref.Controller == nil || !*ref.Controller {
			continue
		}
		if ref.Kind == "ReplicaSet" {
			if hash := pod.Labels["pod-template-hash"]; hash != "" && strings.HasSuffix(ref.Name, "-"+hash) {
				return "Deployment", strings.TrimSuffix(ref.Name, "-"+hash)
			}
		}
		return ref.Kind, ref.Name
	}
	return "", ""
}

func podReadyAndRestarts(pod *corev1.Pod) (string, int32) {
	ready := 0
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
		restarts += cs.RestartCount
	}
	return fmt.Sprintf("%d/%d", ready, len(pod.Spec.Containers)), restarts
}

func Pod2VO(item *corev1.Pod) PodVO {
	status, message := podDisplayStatus(item)
	ready, restarts := podReadyAndRestarts(item)
	ownerKind, ownerName := PodOwner(item)
	return PodVO{
		CreationTimestamp: item.CreationTimestamp.Time,
		Name:              item.Name,
		Namespace:         item.Namespace,
		Alias:             item.Annotations["alias"],
		Describe:          item.Annotations["describe"],
		Status:            string(item.Status.Phase),
		DisplayStatus:     status,
		Message:           message,
		Ready:             ready,
		Restarts:          restarts,
		HostIP:            item.Status.HostIP,
		PodIP:             item.Status.PodIP,
		NodeName:          item.Spec.NodeName,
		OwnerKind:         ownerKind,
		OwnerName:         ownerName,
		PodResources:      PodResourceTotals(&item.Spec),
	}
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func containerStatus2VO(cs corev1.ContainerStatus, init bool) PodContainerStatusVO {
	item := PodContainerStatusVO{
		Name:         cs.Name,
		Init:         init,
		Ready:        cs.Ready,
		Started:      cs.Started != nil && *cs.Started,
		RestartCount: cs.RestartCount,
		ImageID:      cs.ImageID,
	}
	switch {
	case cs.State.Running != nil:
		item.State = "Running"
		item.StartedAt = timePtr(cs.State.Running.StartedAt.Time)
	case cs.State.Waiting != nil:
		item.State = "Waiting"
		item.Reason = cs.State.Waiting.Reason
		item.Message = cs.State.Waiting.Message
	case cs.State.Terminated != nil:
		t := cs.State.Terminated
		item.State = "Terminated"
		item.Reason = t.Reason
		item.Message = t.Message
		item.ExitCode = t.ExitCode
		item.StartedAt = timePtr(t.StartedAt.Time)
		item.FinishedAt = timePtr(t.FinishedAt.Time)
	}
	if t := cs.LastTerminationState.Terminated; t != nil {
		item.LastReason = t.Reason
		item.LastExitCode = t.ExitCode
		item.LastFinishedAt = timePtr(t.FinishedAt.Time)
	}
	return item
}

func tolerationText(t corev1.Toleration) string {
	key := t.Key
	if key == "" {
		key = "*"
	}
	text := key
	if t.Operator == corev1.TolerationOpEqual || (t.Operator == "" && t.Value != "") {
		text += "=" + t.Value
	}
	if t.Effect != "" {
		text += ":" + string(t.Effect)
	}
	if t.TolerationSeconds != nil {
		text += fmt.Sprintf("（%d 秒）", *t.TolerationSeconds)
	}
	return text
}

// Pod2DetailVO 详情; 所属工作负载和事件由 service 层补充
func Pod2DetailVO(pod *corev1.Pod) PodDetailVO {
	status, message := podDisplayStatus(pod)
	ready, restarts := podReadyAndRestarts(pod)
	ownerKind, ownerName := PodOwner(pod)
	result := PodDetailVO{
		BaseVO: BaseVO{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Alias:     pod.Annotations["alias"],
			Describe:  pod.Annotations["describe"],
		},
		Status:             string(pod.Status.Phase),
		DisplayStatus:      status,
		Message:            firstNonEmpty(message, pod.Status.Message),
		Reason:             pod.Status.Reason,
		Ready:              ready,
		Restarts:           restarts,
		NodeName:           pod.Spec.NodeName,
		HostIP:             pod.Status.HostIP,
		PodIP:              pod.Status.PodIP,
		PodIPs:             []string{},
		QosClass:           string(pod.Status.QOSClass),
		RestartPolicy:      string(pod.Spec.RestartPolicy),
		ServiceAccountName: pod.Spec.ServiceAccountName,
		PriorityClassName:  pod.Spec.PriorityClassName,
		DNSPolicy:          string(pod.Spec.DNSPolicy),
		HostNetwork:        pod.Spec.HostNetwork,
		OwnerKind:          ownerKind,
		OwnerName:          ownerName,
		Creator:            ResourceCreator(pod.Annotations),
		Uid:                string(pod.UID),
		CreateTime:         pod.CreationTimestamp.Time,
		Labels:             pod.Labels,
		Annotations:        pod.Annotations,
		NodeSelector:       pod.Spec.NodeSelector,
		Tolerations:        []string{},
		Conditions:         []WorkloadConditionVO{},
		Containers:         PodSpec2Containers(&pod.Spec),
		ContainerStatuses:  []PodContainerStatusVO{},
		Volumes:            PodVolumes2VO(pod.Spec.Volumes),
		Events:             []EventVO{},
		PodResources:       PodResourceTotals(&pod.Spec),
		ContainerUsages:    PodContainerUsages(&pod.Spec),
	}
	if pod.Status.StartTime != nil {
		result.StartTime = timePtr(pod.Status.StartTime.Time)
	}
	if pod.DeletionTimestamp != nil {
		result.DeletionTime = timePtr(pod.DeletionTimestamp.Time)
	}
	for _, ip := range pod.Status.PodIPs {
		result.PodIPs = append(result.PodIPs, ip.IP)
	}
	for _, t := range pod.Spec.Tolerations {
		result.Tolerations = append(result.Tolerations, tolerationText(t))
	}
	for _, c := range pod.Status.Conditions {
		result.Conditions = append(result.Conditions, newCondition(string(c.Type), string(c.Status), c.Reason, c.Message, c.LastTransitionTime))
	}
	for _, cs := range pod.Status.InitContainerStatuses {
		result.ContainerStatuses = append(result.ContainerStatuses, containerStatus2VO(cs, true))
	}
	for _, cs := range pod.Status.ContainerStatuses {
		result.ContainerStatuses = append(result.ContainerStatuses, containerStatus2VO(cs, false))
	}
	return result
}

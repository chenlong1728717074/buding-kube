package vo

import (
	v1 "k8s.io/api/apps/v1"
	"time"
)

func DaemonSet2WorkloadVO(item v1.DaemonSet) WorkloadVO {
	return WorkloadVO{
		BaseVO: BaseVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Alias:     item.Annotations["alias"],
			Describe:  item.Annotations["describe"],
		},
		Status:        GetDaemonSetStatusText(item),
		Replicas:      int(item.Status.DesiredNumberScheduled),
		ReadyReplicas: int(item.Status.NumberReady),
		Containers:    len(item.Spec.Template.Spec.Containers),
		UpdateTime:    GetDaemonSetUpdateTime(item),
		CreateTime:    item.ObjectMeta.CreationTimestamp.Time,
	}
}

func GetDaemonSetUpdateTime(ds v1.DaemonSet) *time.Time {
	for _, cond := range ds.Status.Conditions {
		if cond.Type == "" {
			t := cond.LastTransitionTime.Time
			return &t
		}
	}
	return nil
}

func GetDaemonSetStatusText(ds v1.DaemonSet) string {
	// 获取副本状态
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady
	updated := ds.Status.UpdatedNumberScheduled
	available := ds.Status.NumberAvailable

	// 1. 已停止：期望调度节点数为 0
	if desired == 0 {
		if ds.Status.ObservedGeneration < ds.Generation {
			return "更新中"
		}
		return "已停止"
	}
	// 2. 更新中：控制器未观测到最新 Generation，或更新/就绪/可用节点数未达到预期
	if ds.Status.ObservedGeneration < ds.Generation ||
		updated < desired ||
		ready < desired ||
		available < desired {
		return "更新中"
	}
	// 3. 运行中：一切就绪
	return "运行中"
}

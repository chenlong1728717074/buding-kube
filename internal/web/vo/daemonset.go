package vo

import (
	v1 "k8s.io/api/apps/v1"
	"time"
)

func DaemonSet2WorkloadVO(item v1.DaemonSet) WorkloadVO {
	return WorkloadVO{
		Name:          item.Name,
		Namespace:     item.Namespace,
		Alias:         item.Annotations["alias"],
		Describe:      item.Annotations["describe"],
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

	// 1. 已停止：副本数为 0
	if desired == 0 {
		return "已停止"
	}
	// 2. 更新中：更新的副本数还没达到预期
	if updated < desired || ready < desired {
		return "更新中"
	}
	// 3. 运行中：一切就绪
	if updated == desired && ready == desired {
		return "运行中"
	}
	// 4. 默认兜底
	return "未知"
}

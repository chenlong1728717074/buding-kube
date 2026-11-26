package vo

import (
	v1 "k8s.io/api/apps/v1"
	"time"
)

func StatefulSet2WorkloadVO(item v1.StatefulSet) WorkloadVO {
	return WorkloadVO{
		BaseVO: BaseVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Alias:     item.Annotations["alias"],
			Describe:  item.Annotations["describe"],
		},
		Status:        GetStatefulSetStatusText(item),
		Replicas:      int(int32Value(item.Spec.Replicas, 1)),
		ReadyReplicas: int(item.Status.ReadyReplicas),
		Containers:    len(item.Spec.Template.Spec.Containers),
		UpdateTime:    GetStatefulSetUpdateTime(item),
		CreateTime:    item.ObjectMeta.CreationTimestamp.Time,
	}
}

func GetStatefulSetUpdateTime(sts v1.StatefulSet) *time.Time {
	for _, cond := range sts.Status.Conditions {
		if cond.Type == "" {
			t := cond.LastTransitionTime.Time
			return &t
		}
	}
	return nil
}

func GetStatefulSetStatusText(sts v1.StatefulSet) string {
	// 以 Spec.Replicas 作为期望副本数（未设置时 K8s 默认 1）
	desired := int32Value(sts.Spec.Replicas, 1)
	ready := sts.Status.ReadyReplicas
	updated := sts.Status.UpdatedReplicas

	// 1. 已停止 / 缩容至 0 中
	if desired == 0 {
		if sts.Status.Replicas > 0 || sts.Status.ObservedGeneration < sts.Generation {
			return "更新中"
		}
		return "已停止"
	}
	// 2. 更新中：控制器尚未观测到最新 Generation，或副本/更新/版本未完全同步
	if sts.Status.ObservedGeneration < sts.Generation ||
		sts.Status.Replicas != desired ||
		updated < desired ||
		ready < desired ||
		(sts.Status.UpdateRevision != "" && sts.Status.CurrentRevision != "" && sts.Status.CurrentRevision != sts.Status.UpdateRevision) {
		return "更新中"
	}
	// 3. 运行中：一切就绪
	return "运行中"
}

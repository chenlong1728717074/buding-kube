package vo

import (
	v1 "k8s.io/api/apps/v1"
	"time"
)

func StatefulSet2WorkloadVO(item v1.StatefulSet) WorkloadVO {
	return WorkloadVO{
		Name:          item.Name,
		Namespace:     item.Namespace,
		Alias:         item.Annotations["alias"],
		Describe:      item.Annotations["describe"],
		Status:        GetStatefulSetStatusText(item),
		Replicas:      int(item.Status.Replicas),
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
	// 获取副本状态
	desired := sts.Status.Replicas
	ready := sts.Status.ReadyReplicas
	updated := sts.Status.UpdatedReplicas

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

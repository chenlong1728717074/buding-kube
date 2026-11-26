package vo

import (
	v1 "k8s.io/api/apps/v1"
	"time"
)

func Deployment2WorkloadVO(item v1.Deployment) WorkloadVO {
	return WorkloadVO{
		BaseVO: BaseVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Alias:     item.Annotations["alias"],
			Describe:  item.Annotations["describe"],
		},
		Status:        GetDeploymentStatusText(item),
		Replicas:      int(int32Value(item.Spec.Replicas, 1)),
		ReadyReplicas: int(item.Status.ReadyReplicas),
		Containers:    len(item.Spec.Template.Spec.Containers),
		UpdateTime:    GetDeploymentUpdateTime(item),
		CreateTime:    item.ObjectMeta.CreationTimestamp.Time,
	}
}

func GetDeploymentUpdateTime(deploy v1.Deployment) *time.Time {
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == v1.DeploymentProgressing {
			t := cond.LastUpdateTime.Time
			return &t
		}
	}
	return nil
}

func GetDeploymentStatusText(deploy v1.Deployment) string {
	// 以 Spec.Replicas 作为期望副本数（未设置时 K8s 默认 1），避免刚更新时 Status.Replicas 仍是旧值
	desired := int32Value(deploy.Spec.Replicas, 1)
	ready := deploy.Status.ReadyReplicas
	updated := deploy.Status.UpdatedReplicas
	available := deploy.Status.AvailableReplicas

	// 1. 已停止 / 缩容至 0 中
	if desired == 0 {
		if deploy.Status.Replicas > 0 || deploy.Status.ObservedGeneration < deploy.Generation {
			return "更新中"
		}
		return "已停止"
	}
	// 2. 更新中：控制器尚未观测到最新 Generation，或副本/就绪/可用数未达到预期
	if deploy.Status.ObservedGeneration < deploy.Generation ||
		deploy.Status.Replicas != desired ||
		updated < desired ||
		ready < desired ||
		available < desired {
		return "更新中"
	}
	// 3. 运行中：一切就绪
	return "运行中"
}

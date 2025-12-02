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
		},
		Alias:         item.Annotations["alias"],
		Describe:      item.Annotations["describe"],
		Status:        GetDeploymentStatusText(item),
		Replicas:      int(item.Status.Replicas),
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
	// 获取副本状态
	desired := deploy.Status.Replicas
	ready := deploy.Status.ReadyReplicas
	updated := deploy.Status.UpdatedReplicas
	available := deploy.Status.AvailableReplicas

	// 1. 已停止：副本数为 0
	if desired == 0 {
		return "已停止"
	}
	// 2. 更新中：更新的副本数还没达到预期
	if updated < desired || ready < desired || available < desired {
		return "更新中"
	}
	// 3. 运行中：一切就绪
	if updated == desired && ready == desired && available == desired {
		return "运行中"
	}
	// 4. 默认兜底
	return "未知"
}

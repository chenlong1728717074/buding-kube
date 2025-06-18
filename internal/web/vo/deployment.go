package vo

import (
	"k8s.io/api/apps/v1"
	"time"
)

type DeploymentVO struct {
	Name          string     `json:"name"`
	Namespace     string     `json:"namespace"`
	Alias         string     `json:"alias"`
	Describe      string     `json:"describe"`
	Yaml          string     `json:"yaml"`
	Status        string     `json:"status"`
	Containers    int        `json:"containers"`
	Replicas      int        `json:"replicas"`      // 期望的 Pod 副本数
	ReadyReplicas int        `json:"readyReplicas"` // 当前 Ready 状态的 Pod 数
	UpdateTime    *time.Time `json:"updateTime"`
	CreateTime    time.Time  `json:"createTime"`
}

func Deployment2VO(item v1.Deployment) DeploymentVO {
	return DeploymentVO{
		Name:          item.Name,
		Namespace:     item.Namespace,
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

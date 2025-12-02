package vo

import (
	appsv1 "k8s.io/api/apps/v1"
	"time"
)

type WorkloadVO struct {
	BaseVO
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

// ReplicaSet2VO 将ReplicaSet对象转换为WorkloadVO
func ReplicaSet2VO(rs *appsv1.ReplicaSet) WorkloadVO {
	status := "Running"
	if rs.Status.Replicas != rs.Status.ReadyReplicas {
		status = "Pending"
	}
	if rs.Status.Replicas == 0 {
		status = "Stopped"
	}

	var updateTime *time.Time
	if len(rs.Status.Conditions) > 0 {
		t := rs.Status.Conditions[len(rs.Status.Conditions)-1].LastTransitionTime.Time
		updateTime = &t
	}

	replicas := int32(0)
	if rs.Spec.Replicas != nil {
		replicas = *rs.Spec.Replicas
	}

	alias := ""
	describe := ""
	if rs.Annotations != nil {
		alias = rs.Annotations["alias"]
		describe = rs.Annotations["describe"]
	}

	return WorkloadVO{
		BaseVO: BaseVO{
			Name:      rs.Name,
			Namespace: rs.Namespace,
		},
		Alias:         alias,
		Describe:      describe,
		Status:        status,
		Containers:    len(rs.Spec.Template.Spec.Containers),
		Replicas:      int(replicas),
		ReadyReplicas: int(rs.Status.ReadyReplicas),
		UpdateTime:    updateTime,
		CreateTime:    rs.CreationTimestamp.Time,
	}
}

package vo

import (
	"time"
)

type WorkloadVO struct {
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

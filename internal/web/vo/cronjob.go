package vo

import (
	batchv1 "k8s.io/api/batch/v1"
	"time"
)

type CronJobVO struct {
	BaseVO
	Alias              string    `json:"alias"`
	Describe           string    `json:"describe"`
	CreateTime         time.Time `json:"createTime"`
	LastRunTime        time.Time `json:"lastRunTime"`
	LastSuccessfulTime time.Time `json:"lastSuccessfulTime"`
	Schedule           string    `json:"schedule"` // Cron表达式
	Yaml               string    `json:"yaml"`
}

func CronJob2VO(cj batchv1.CronJob) CronJobVO {
	result := CronJobVO{
		BaseVO: BaseVO{
			Name:      cj.Name,
			Namespace: cj.Namespace,
		},
		Alias:      cj.Annotations["alias"],
		Describe:   cj.Annotations["describe"],
		CreateTime: cj.CreationTimestamp.Time,
		Schedule:   cj.Spec.Schedule,
	}
	if cj.Status.LastScheduleTime != nil {
		result.LastRunTime = cj.Status.LastScheduleTime.Time
	}
	if cj.Status.LastSuccessfulTime != nil {
		result.LastSuccessfulTime = cj.Status.LastSuccessfulTime.Time
	}
	return result
}

package vo

import (
	batchv1 "k8s.io/api/batch/v1"
	"time"
)

type JobVO struct {
	BaseVO
	Alias       string    `json:"alias"`
	Describe    string    `json:"describe"`
	CreateTime  time.Time `json:"createTime"`
	LastRunTime time.Time `json:"lastRunTime"`
	Yaml        string    `json:"yaml"`
}

func Job2VO(job batchv1.Job) JobVO {
	var lastRunTime time.Time
	if job.Status.StartTime != nil {
		lastRunTime = job.Status.StartTime.Time
	}

	return JobVO{
		BaseVO: BaseVO{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
		Alias:       job.Annotations["alias"],
		Describe:    job.Annotations["describe"],
		CreateTime:  job.CreationTimestamp.Time,
		LastRunTime: lastRunTime,
	}
}

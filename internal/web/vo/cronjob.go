package vo

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type CronJobVO struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Schedule                   string            `json:"schedule"`
	Suspend                    *bool             `json:"suspend"`
	ConcurrencyPolicy          string            `json:"concurrencyPolicy"`
	StartingDeadlineSeconds    *int64            `json:"startingDeadlineSeconds"`
	SuccessfulJobsHistoryLimit *int32            `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     *int32            `json:"failedJobsHistoryLimit"`
	LastScheduleTime           *time.Time        `json:"lastScheduleTime"`
	Generation                 int64             `json:"generation"`
	Version                    string            `json:"version"`
	Uid                        string            `json:"uid"`
	ResourceVersion            string            `json:"resourceVersion"`
	Annotations                map[string]string `json:"annotations"`
	Labels                     map[string]string `json:"labels"`
	CreationTimestamp          time.Time         `json:"creationTimestamp"`
	CreateTime                 time.Time         `json:"createTime"`
	Metadata                   metav1.ObjectMeta `json:"metadata"`
	Yaml                       string            `json:"yaml"`
}

func CronJob2VO(cj batchv1.CronJob) CronJobVO {
	return CronJobVO{
		Name:                       cj.Name,
		Namespace:                  cj.Namespace,
		Schedule:                   cj.Spec.Schedule,
		Suspend:                    cj.Spec.Suspend,
		ConcurrencyPolicy:          string(cj.Spec.ConcurrencyPolicy),
		StartingDeadlineSeconds:    cj.Spec.StartingDeadlineSeconds,
		SuccessfulJobsHistoryLimit: cj.Spec.SuccessfulJobsHistoryLimit,
		FailedJobsHistoryLimit:     cj.Spec.FailedJobsHistoryLimit,
		LastScheduleTime:           &cj.Status.LastScheduleTime.Time,
		Generation:                 cj.Generation,
		Version:                    cj.APIVersion,
		Uid:                        string(cj.UID),
		ResourceVersion:            cj.ResourceVersion,
		Annotations:                cj.Annotations,
		Labels:                     cj.Labels,
		CreationTimestamp:          cj.CreationTimestamp.Time,
		CreateTime:                 cj.CreationTimestamp.Time,
		Metadata:                   cj.ObjectMeta,
	}
}

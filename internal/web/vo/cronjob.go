package vo

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
)

// CronJobVO 定时任务列表项
type CronJobVO struct {
	BaseVO
	// Schedule Cron 表达式
	Schedule           string     `json:"schedule"`
	TimeZone           string     `json:"timeZone"`
	ConcurrencyPolicy  string     `json:"concurrencyPolicy"`
	Suspend            bool       `json:"suspend"`
	Active             int        `json:"active"`
	CreateTime         time.Time  `json:"createTime"`
	LastRunTime        *time.Time `json:"lastRunTime"`
	LastSuccessfulTime *time.Time `json:"lastSuccessfulTime"`
}

// CronJobDetailVO 定时任务详情
type CronJobDetailVO struct {
	CronJobVO
	StartingDeadlineSeconds    *int64 `json:"startingDeadlineSeconds"`
	SuccessfulJobsHistoryLimit int32  `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     int32  `json:"failedJobsHistoryLimit"`
	// 以下为任务模板的设置
	Completions             int32                 `json:"completions"`
	Parallelism             int32                 `json:"parallelism"`
	BackoffLimit            int32                 `json:"backoffLimit"`
	ActiveDeadlineSeconds   *int64                `json:"activeDeadlineSeconds"`
	TTLSecondsAfterFinished *int32                `json:"ttlSecondsAfterFinished"`
	RestartPolicy           string                `json:"restartPolicy"`
	Creator                 string                `json:"creator"`
	Uid                     string                `json:"uid"`
	Labels                  map[string]string     `json:"labels"`
	Annotations             map[string]string     `json:"annotations"`
	Containers              []ContainerTemplateVO `json:"containers"`
	// Jobs 定时任务创建的任务(运行记录)
	Jobs   []JobVO    `json:"jobs"`
	Pods   []JobPodVO `json:"pods"`
	Events []EventVO  `json:"events"`
}

func CronJob2VO(cj batchv1.CronJob) CronJobVO {
	result := CronJobVO{
		BaseVO: BaseVO{
			Name:      cj.Name,
			Namespace: cj.Namespace,
			Alias:     cj.Annotations["alias"],
			Describe:  cj.Annotations["describe"],
		},
		Schedule:          cj.Spec.Schedule,
		ConcurrencyPolicy: string(cj.Spec.ConcurrencyPolicy),
		Suspend:           cj.Spec.Suspend != nil && *cj.Spec.Suspend,
		Active:            len(cj.Status.Active),
		CreateTime:        cj.CreationTimestamp.Time,
	}
	if result.ConcurrencyPolicy == "" {
		result.ConcurrencyPolicy = string(batchv1.AllowConcurrent)
	}
	if cj.Spec.TimeZone != nil {
		result.TimeZone = *cj.Spec.TimeZone
	}
	if cj.Status.LastScheduleTime != nil {
		t := cj.Status.LastScheduleTime.Time
		result.LastRunTime = &t
	}
	if cj.Status.LastSuccessfulTime != nil {
		t := cj.Status.LastSuccessfulTime.Time
		result.LastSuccessfulTime = &t
	}
	return result
}

func CronJob2DetailVO(cj *batchv1.CronJob) CronJobDetailVO {
	spec := cj.Spec.JobTemplate.Spec
	return CronJobDetailVO{
		CronJobVO:                  CronJob2VO(*cj),
		StartingDeadlineSeconds:    cj.Spec.StartingDeadlineSeconds,
		SuccessfulJobsHistoryLimit: int32Value(cj.Spec.SuccessfulJobsHistoryLimit, 3),
		FailedJobsHistoryLimit:     int32Value(cj.Spec.FailedJobsHistoryLimit, 1),
		Completions:                int32Value(spec.Completions, 1),
		Parallelism:                int32Value(spec.Parallelism, 1),
		BackoffLimit:               int32Value(spec.BackoffLimit, 6),
		ActiveDeadlineSeconds:      spec.ActiveDeadlineSeconds,
		TTLSecondsAfterFinished:    spec.TTLSecondsAfterFinished,
		RestartPolicy:              string(spec.Template.Spec.RestartPolicy),
		Creator:                    ResourceCreator(cj.Annotations),
		Uid:                        string(cj.UID),
		Labels:                     cj.Labels,
		Annotations:                cj.Annotations,
		Containers:                 PodSpec2Containers(&spec.Template.Spec),
		Jobs:                       []JobVO{},
		Pods:                       []JobPodVO{},
		Events:                     []EventVO{},
	}
}

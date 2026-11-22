package dto

// CronJobCreateDTO 创建定时任务
type CronJobCreateDTO struct {
	ClusterId string            `json:"clusterId" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Alias     string            `json:"alias"`
	Describe  string            `json:"describe"`
	Labels    map[string]string `json:"labels"`
	// Schedule Cron 表达式, 例如 0 2 * * *
	Schedule string `json:"schedule" binding:"required"`
	// TimeZone 时区, 例如 Asia/Shanghai; 为空时使用 kube-controller-manager 所在时区
	TimeZone string `json:"timeZone"`
	// ConcurrencyPolicy 并发策略: Allow(默认) / Forbid / Replace
	ConcurrencyPolicy string `json:"concurrencyPolicy"`
	// StartingDeadlineSeconds 错过调度时间后, 多少秒内仍可启动
	StartingDeadlineSeconds    *int64         `json:"startingDeadlineSeconds"`
	SuccessfulJobsHistoryLimit *int32         `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     *int32         `json:"failedJobsHistoryLimit"`
	Suspend                    bool           `json:"suspend"`
	Template                   JobTemplateDTO `json:"template"`
}

// CronJobUpdateDTO 编辑定时任务; 字符串为空或指针为空表示不修改
type CronJobUpdateDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	Schedule  string `json:"schedule"`
	// TimeZone 传空字符串表示清除时区
	TimeZone                   *string `json:"timeZone"`
	ConcurrencyPolicy          string  `json:"concurrencyPolicy"`
	Suspend                    *bool   `json:"suspend"`
	StartingDeadlineSeconds    *int64  `json:"startingDeadlineSeconds"`
	SuccessfulJobsHistoryLimit *int32  `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     *int32  `json:"failedJobsHistoryLimit"`
}

// CronJobSuspendDTO 暂停 / 启用定时任务
type CronJobSuspendDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Suspend   bool   `json:"suspend"`
}

type CronJobBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type CronJobPageQueryBaseDTO struct {
	PageQueryDTO
	CronJobBaseDTO
	// Status 状态过滤: Active(运行中) / Suspended(已暂停)
	Status string `json:"status" form:"status"`
}

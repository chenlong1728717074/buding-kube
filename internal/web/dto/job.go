package dto

// JobEnvDTO 环境变量
type JobEnvDTO struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// JobContainerDTO 任务的容器
type JobContainerDTO struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	// ImagePullPolicy 镜像拉取策略: IfNotPresent / Always / Never, 为空时由 Kubernetes 决定
	ImagePullPolicy string `json:"imagePullPolicy"`
	// Command / Args 每项一个参数
	Command       []string    `json:"command"`
	Args          []string    `json:"args"`
	WorkingDir    string      `json:"workingDir"`
	Env           []JobEnvDTO `json:"env"`
	CPURequest    string      `json:"cpuRequest"`
	CPULimit      string      `json:"cpuLimit"`
	MemoryRequest string      `json:"memoryRequest"`
	MemoryLimit   string      `json:"memoryLimit"`
}

// JobTemplateDTO 任务设置, 任务和定时任务共用
type JobTemplateDTO struct {
	// Completions 需要成功完成的容器组数量, 默认 1
	Completions *int32 `json:"completions"`
	// Parallelism 并行运行的容器组数量, 默认 1
	Parallelism *int32 `json:"parallelism"`
	// BackoffLimit 失败后的最大重试次数, 默认 6
	BackoffLimit *int32 `json:"backoffLimit"`
	// ActiveDeadlineSeconds 最大运行时间(秒), 不大于 0 表示不限制
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds"`
	// TTLSecondsAfterFinished 结束后多少秒自动删除, 为空表示不自动删除
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished"`
	// RestartPolicy 重启策略: Never(默认) / OnFailure
	RestartPolicy string            `json:"restartPolicy"`
	Containers    []JobContainerDTO `json:"containers"`
}

// JobCreateDTO 创建任务
type JobCreateDTO struct {
	ClusterId string            `json:"clusterId" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Alias     string            `json:"alias"`
	Describe  string            `json:"describe"`
	Labels    map[string]string `json:"labels"`
	Template  JobTemplateDTO    `json:"template"`
}

// JobUpdateDTO 编辑任务. 任务创建后容器、完成数量等设置不能修改, 这里只包含可修改的字段; 指针为空表示不修改
type JobUpdateDTO struct {
	ClusterId               string `json:"clusterId" binding:"required"`
	Namespace               string `json:"namespace" binding:"required"`
	Name                    string `json:"name" binding:"required"`
	Alias                   string `json:"alias"`
	Describe                string `json:"describe"`
	Parallelism             *int32 `json:"parallelism"`
	ActiveDeadlineSeconds   *int64 `json:"activeDeadlineSeconds"`
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished"`
}

type JobBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type JobPageQueryBaseDTO struct {
	PageQueryDTO
	JobBaseDTO
	// Status 状态过滤: Pending / Running / Complete / Failed / Suspended
	Status string `json:"status" form:"status"`
}

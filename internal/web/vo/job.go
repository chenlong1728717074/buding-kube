package vo

import (
	"buding-kube/pkg/consts"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// Job 状态
const (
	JobStatusPending   = "Pending"
	JobStatusRunning   = "Running"
	JobStatusComplete  = "Complete"
	JobStatusFailed    = "Failed"
	JobStatusSuspended = "Suspended"
)

// JobVO 任务列表项
type JobVO struct {
	BaseVO
	Status      string `json:"status"`
	Message     string `json:"message"`
	Completions int32  `json:"completions"`
	Parallelism int32  `json:"parallelism"`
	// BackoffLimit 最大重试次数
	BackoffLimit int32 `json:"backoffLimit"`
	Active       int32 `json:"active"`
	Succeeded    int32 `json:"succeeded"`
	Failed       int32 `json:"failed"`
	// OwnerCronJob 所属的定时任务
	OwnerCronJob string `json:"ownerCronJob"`
	// Manual 是否由定时任务手动触发
	Manual         bool       `json:"manual"`
	StartTime      *time.Time `json:"startTime"`
	CompletionTime *time.Time `json:"completionTime"`
	CreateTime     time.Time  `json:"createTime"`
	LastRunTime    *time.Time `json:"lastRunTime"`
}

// JobDetailVO 任务详情
type JobDetailVO struct {
	JobVO
	ActiveDeadlineSeconds   *int64                `json:"activeDeadlineSeconds"`
	TTLSecondsAfterFinished *int32                `json:"ttlSecondsAfterFinished"`
	RestartPolicy           string                `json:"restartPolicy"`
	Suspend                 bool                  `json:"suspend"`
	Creator                 string                `json:"creator"`
	Uid                     string                `json:"uid"`
	Labels                  map[string]string     `json:"labels"`
	Annotations             map[string]string     `json:"annotations"`
	Containers              []ContainerTemplateVO `json:"containers"`
	Records                 []JobRunRecordVO      `json:"records"`
	Pods                    []JobPodVO            `json:"pods"`
	Events                  []EventVO             `json:"events"`
}

// JobRunRecordVO 一次运行的记录; 重新执行前的记录保存在注解里
type JobRunRecordVO struct {
	Run            int        `json:"run"`
	Current        bool       `json:"current"`
	Status         string     `json:"status"`
	Message        string     `json:"message,omitempty"`
	Completions    int32      `json:"completions"`
	Succeeded      int32      `json:"succeeded"`
	Failed         int32      `json:"failed"`
	StartTime      *time.Time `json:"startTime"`
	CompletionTime *time.Time `json:"completionTime"`
}

// ContainerTemplateVO 容器模板
type ContainerTemplateVO struct {
	Name            string            `json:"name"`
	Init            bool              `json:"init"`
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy"`
	Command         []string          `json:"command"`
	Args            []string          `json:"args"`
	WorkingDir      string            `json:"workingDir"`
	Envs            []ContainerEnvVO  `json:"envs"`
	Requests        map[string]string `json:"requests"`
	Limits          map[string]string `json:"limits"`
	Ports           []ContainerPortVO `json:"ports"`
	LivenessProbe   string            `json:"livenessProbe"`
	ReadinessProbe  string            `json:"readinessProbe"`
	StartupProbe    string            `json:"startupProbe"`
	VolumeMounts    []VolumeMountVO   `json:"volumeMounts"`
}

// ContainerEnvVO 环境变量; Source 不为空时表示值来自 ConfigMap / Secret 等引用
type ContainerEnvVO struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

// JobPodVO 任务创建的容器组
type JobPodVO struct {
	Name       string     `json:"name"`
	Namespace  string     `json:"namespace"`
	JobName    string     `json:"jobName"`
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	NodeName   string     `json:"nodeName"`
	PodIP      string     `json:"podIP"`
	Ready      string     `json:"ready"`
	Restarts   int32      `json:"restarts"`
	StartTime  *time.Time `json:"startTime"`
	CreateTime time.Time  `json:"createTime"`
	PodResources
	PodUsage
}

func Job2VO(job batchv1.Job) JobVO {
	status, message := JobStatus(&job)
	result := JobVO{
		BaseVO: BaseVO{
			Name:      job.Name,
			Namespace: job.Namespace,
			Alias:     job.Annotations["alias"],
			Describe:  job.Annotations["describe"],
		},
		Status:         status,
		Message:        message,
		Completions:    int32Value(job.Spec.Completions, 1),
		Parallelism:    int32Value(job.Spec.Parallelism, 1),
		BackoffLimit:   int32Value(job.Spec.BackoffLimit, 6),
		Active:         job.Status.Active,
		Succeeded:      job.Status.Succeeded,
		Failed:         job.Status.Failed,
		Manual:         job.Annotations[consts.CronJobInstantiateAnnotation] == "manual",
		CompletionTime: JobFinishTime(&job),
		CreateTime:     job.CreationTimestamp.Time,
	}
	for _, ref := range job.OwnerReferences {
		if ref.Kind == "CronJob" {
			result.OwnerCronJob = ref.Name
			break
		}
	}
	if job.Status.StartTime != nil {
		t := job.Status.StartTime.Time
		result.StartTime = &t
		result.LastRunTime = &t
	}
	return result
}

func Job2DetailVO(job *batchv1.Job) JobDetailVO {
	// 运行记录单独展示, 不放在注解里
	annotations := make(map[string]string, len(job.Annotations))
	for k, v := range job.Annotations {
		if k != consts.JobRunRecordsAnnotation {
			annotations[k] = v
		}
	}
	return JobDetailVO{
		JobVO:                   Job2VO(*job),
		ActiveDeadlineSeconds:   job.Spec.ActiveDeadlineSeconds,
		TTLSecondsAfterFinished: job.Spec.TTLSecondsAfterFinished,
		RestartPolicy:           string(job.Spec.Template.Spec.RestartPolicy),
		Suspend:                 job.Spec.Suspend != nil && *job.Spec.Suspend,
		Creator:                 ResourceCreator(job.Annotations),
		Uid:                     string(job.UID),
		Labels:                  job.Labels,
		Annotations:             annotations,
		Containers:              PodSpec2Containers(&job.Spec.Template.Spec),
		Records:                 []JobRunRecordVO{},
		Pods:                    []JobPodVO{},
		Events:                  []EventVO{},
	}
}

// JobStatus 根据 conditions 判断任务状态, 第二个返回值是失败原因
func JobStatus(job *batchv1.Job) (string, string) {
	for _, c := range job.Status.Conditions {
		if c.Status != corev1.ConditionTrue {
			continue
		}
		switch c.Type {
		case batchv1.JobComplete:
			return JobStatusComplete, c.Message
		case batchv1.JobFailed:
			return JobStatusFailed, firstNonEmpty(c.Message, c.Reason)
		}
	}
	if job.Spec.Suspend != nil && *job.Spec.Suspend {
		return JobStatusSuspended, ""
	}
	if job.Status.Active > 0 {
		return JobStatusRunning, ""
	}
	return JobStatusPending, ""
}

// JobFinishTime 结束时间: 成功取 completionTime, 失败取失败条件的时间
func JobFinishTime(job *batchv1.Job) *time.Time {
	if job.Status.CompletionTime != nil {
		t := job.Status.CompletionTime.Time
		return &t
	}
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			t := c.LastTransitionTime.Time
			return &t
		}
	}
	return nil
}

// CurrentJobRunRecord 当前这次运行的记录
func CurrentJobRunRecord(job *batchv1.Job, run int) JobRunRecordVO {
	status, message := JobStatus(job)
	record := JobRunRecordVO{
		Run:            run,
		Current:        true,
		Status:         status,
		Message:        message,
		Completions:    int32Value(job.Spec.Completions, 1),
		Succeeded:      job.Status.Succeeded,
		Failed:         job.Status.Failed,
		CompletionTime: JobFinishTime(job),
	}
	if job.Status.StartTime != nil {
		t := job.Status.StartTime.Time
		record.StartTime = &t
	}
	return record
}

// PodSpec2Containers 提取容器模板(初始化容器在前)
func PodSpec2Containers(spec *corev1.PodSpec) []ContainerTemplateVO {
	result := make([]ContainerTemplateVO, 0, len(spec.InitContainers)+len(spec.Containers))
	for i := range spec.InitContainers {
		result = append(result, container2Template(&spec.InitContainers[i], true))
	}
	for i := range spec.Containers {
		result = append(result, container2Template(&spec.Containers[i], false))
	}
	return result
}

func container2Template(c *corev1.Container, init bool) ContainerTemplateVO {
	envs := make([]ContainerEnvVO, 0, len(c.Env)+len(c.EnvFrom))
	for _, env := range c.Env {
		value, source := envValueSource(env)
		envs = append(envs, ContainerEnvVO{Name: env.Name, Value: value, Source: source})
	}
	for _, from := range c.EnvFrom {
		switch {
		case from.ConfigMapRef != nil:
			envs = append(envs, ContainerEnvVO{Name: from.Prefix + "*", Source: fmt.Sprintf("ConfigMap %s 的全部键", from.ConfigMapRef.Name)})
		case from.SecretRef != nil:
			envs = append(envs, ContainerEnvVO{Name: from.Prefix + "*", Source: fmt.Sprintf("Secret %s 的全部键", from.SecretRef.Name)})
		}
	}
	return ContainerTemplateVO{
		Name:            c.Name,
		Init:            init,
		Image:           c.Image,
		ImagePullPolicy: string(c.ImagePullPolicy),
		Command:         c.Command,
		Args:            c.Args,
		WorkingDir:      c.WorkingDir,
		Envs:            envs,
		Requests:        resourceListToMap(c.Resources.Requests),
		Limits:          resourceListToMap(c.Resources.Limits),
		Ports:           containerPorts(c),
		LivenessProbe:   probeText(c.LivenessProbe),
		ReadinessProbe:  probeText(c.ReadinessProbe),
		StartupProbe:    probeText(c.StartupProbe),
		VolumeMounts:    volumeMounts(c),
	}
}

func envValueSource(env corev1.EnvVar) (string, string) {
	from := env.ValueFrom
	if from == nil {
		return env.Value, ""
	}
	switch {
	case from.ConfigMapKeyRef != nil:
		return "", fmt.Sprintf("ConfigMap %s / %s", from.ConfigMapKeyRef.Name, from.ConfigMapKeyRef.Key)
	case from.SecretKeyRef != nil:
		return "", fmt.Sprintf("Secret %s / %s", from.SecretKeyRef.Name, from.SecretKeyRef.Key)
	case from.FieldRef != nil:
		return "", "字段 " + from.FieldRef.FieldPath
	case from.ResourceFieldRef != nil:
		return "", "资源 " + from.ResourceFieldRef.Resource
	default:
		return "", "引用"
	}
}

func resourceListToMap(list corev1.ResourceList) map[string]string {
	if len(list) == 0 {
		return nil
	}
	result := make(map[string]string, len(list))
	for name, qty := range list {
		result[string(name)] = qty.String()
	}
	return result
}

// Pod2JobPodVO 容器组的展示信息, 状态与 kubectl get pod 的 STATUS 列接近
func Pod2JobPodVO(pod *corev1.Pod) JobPodVO {
	status, message := podDisplayStatus(pod)
	ready := 0
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
		restarts += cs.RestartCount
	}
	result := JobPodVO{
		Name:         pod.Name,
		Namespace:    pod.Namespace,
		JobName:      firstNonEmpty(pod.Labels["batch.kubernetes.io/job-name"], pod.Labels["job-name"]),
		Status:       status,
		Message:      message,
		NodeName:     pod.Spec.NodeName,
		PodIP:        pod.Status.PodIP,
		Ready:        fmt.Sprintf("%d/%d", ready, len(pod.Spec.Containers)),
		Restarts:     restarts,
		CreateTime:   pod.CreationTimestamp.Time,
		PodResources: PodResourceTotals(&pod.Spec),
	}
	if pod.Status.StartTime != nil {
		t := pod.Status.StartTime.Time
		result.StartTime = &t
	}
	return result
}

func podDisplayStatus(pod *corev1.Pod) (string, string) {
	if pod.DeletionTimestamp != nil {
		return "Terminating", ""
	}
	status, message := string(pod.Status.Phase), pod.Status.Message
	if pod.Status.Reason != "" {
		status = pod.Status.Reason
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if w := cs.State.Waiting; w != nil && w.Reason != "" {
			return w.Reason, w.Message
		}
		if t := cs.State.Terminated; t != nil && pod.Status.Phase == corev1.PodFailed && t.Reason != "" {
			status, message = t.Reason, firstNonEmpty(t.Message, message)
		}
	}
	return status, message
}

// ResourceCreator 创建者: 优先使用本平台写入的注解, 其次 KubeSphere 的注解
func ResourceCreator(annotations map[string]string) string {
	return firstNonEmpty(annotations[consts.CreatorAnnotation], annotations[consts.KubeSphereCreatorAnnotation])
}

func int32Value(p *int32, def int32) int32 {
	if p == nil {
		return def
	}
	return *p
}

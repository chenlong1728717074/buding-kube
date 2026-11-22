package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultJobBackoffLimit int32 = 6
	// maxJobRunRecords 最多保留的历史运行记录
	maxJobRunRecords = 20
	// maxCronJobNameLength 定时任务名称最长 52: 控制器创建任务时会追加 11 位后缀, 而任务名称不能超过 63
	maxCronJobNameLength = 52
)

// jobControllerLabels 任务控制器自动添加的标签, 重新执行时需要去掉, 否则新任务的选择器校验不通过
var jobControllerLabels = []string{
	"controller-uid",
	"job-name",
	"batch.kubernetes.io/controller-uid",
	"batch.kubernetes.io/job-name",
}

var validPullPolicies = map[corev1.PullPolicy]bool{
	corev1.PullAlways:       true,
	corev1.PullIfNotPresent: true,
	corev1.PullNever:        true,
}

// validateJobName 任务名称会作为容器组的标签值, 必须是不超过 63 个字符的 DNS 标签
func validateJobName(name string, maxLen int) error {
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("名称不合法: 只能包含小写字母、数字和 \"-\", 且以字母或数字开头和结尾")
	}
	if len(name) > maxLen {
		return fmt.Errorf("名称不能超过 %d 个字符", maxLen)
	}
	return nil
}

// buildJobSpec 根据表单生成任务规格, podLabels 会加到容器组模板上
func buildJobSpec(t dto.JobTemplateDTO, podLabels map[string]string) (batchv1.JobSpec, error) {
	completions := int32(1)
	if t.Completions != nil {
		if *t.Completions < 1 {
			return batchv1.JobSpec{}, errors.New("完成数量至少为 1")
		}
		completions = *t.Completions
	}
	parallelism := int32(1)
	if t.Parallelism != nil {
		if *t.Parallelism < 0 {
			return batchv1.JobSpec{}, errors.New("并行数量不能小于 0")
		}
		parallelism = *t.Parallelism
	}
	backoff := defaultJobBackoffLimit
	if t.BackoffLimit != nil {
		if *t.BackoffLimit < 0 {
			return batchv1.JobSpec{}, errors.New("最大重试次数不能小于 0")
		}
		backoff = *t.BackoffLimit
	}

	restart := corev1.RestartPolicy(strings.TrimSpace(t.RestartPolicy))
	switch restart {
	case "":
		restart = corev1.RestartPolicyNever
	case corev1.RestartPolicyNever, corev1.RestartPolicyOnFailure:
	default:
		return batchv1.JobSpec{}, errors.New("重启策略只能是 Never 或 OnFailure")
	}

	containers, err := buildJobContainers(t.Containers)
	if err != nil {
		return batchv1.JobSpec{}, err
	}

	spec := batchv1.JobSpec{
		Completions:  &completions,
		Parallelism:  &parallelism,
		BackoffLimit: &backoff,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
			Spec: corev1.PodSpec{
				RestartPolicy: restart,
				Containers:    containers,
			},
		},
	}
	if t.ActiveDeadlineSeconds != nil && *t.ActiveDeadlineSeconds > 0 {
		v := *t.ActiveDeadlineSeconds
		spec.ActiveDeadlineSeconds = &v
	}
	if t.TTLSecondsAfterFinished != nil && *t.TTLSecondsAfterFinished >= 0 {
		v := *t.TTLSecondsAfterFinished
		spec.TTLSecondsAfterFinished = &v
	}
	return spec, nil
}

func buildJobContainers(items []dto.JobContainerDTO) ([]corev1.Container, error) {
	if len(items) == 0 {
		return nil, errors.New("请至少添加一个容器")
	}
	seen := map[string]bool{}
	result := make([]corev1.Container, 0, len(items))
	for i, item := range items {
		label := fmt.Sprintf("第 %d 个容器", i+1)
		name := strings.TrimSpace(item.Name)
		if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
			return nil, fmt.Errorf("%s的名称不合法: 只能包含小写字母、数字和 \"-\"", label)
		}
		if seen[name] {
			return nil, fmt.Errorf("容器名称 %s 重复", name)
		}
		seen[name] = true

		image := strings.TrimSpace(item.Image)
		if image == "" {
			return nil, fmt.Errorf("请填写%s的镜像", label)
		}
		c := corev1.Container{
			Name:       name,
			Image:      image,
			Command:    cleanStrings(item.Command),
			Args:       cleanStrings(item.Args),
			WorkingDir: strings.TrimSpace(item.WorkingDir),
		}
		if p := corev1.PullPolicy(strings.TrimSpace(item.ImagePullPolicy)); p != "" {
			if !validPullPolicies[p] {
				return nil, fmt.Errorf("%s的镜像拉取策略不正确: %s", label, p)
			}
			c.ImagePullPolicy = p
		}
		for _, env := range item.Env {
			envName := strings.TrimSpace(env.Name)
			if envName == "" {
				continue
			}
			if errs := validation.IsEnvVarName(envName); len(errs) > 0 {
				return nil, fmt.Errorf("%s的环境变量名 %s 不合法", label, envName)
			}
			c.Env = append(c.Env, corev1.EnvVar{Name: envName, Value: env.Value})
		}
		resources, err := buildContainerResources(item, label)
		if err != nil {
			return nil, err
		}
		c.Resources = resources
		result = append(result, c)
	}
	return result, nil
}

func buildContainerResources(item dto.JobContainerDTO, label string) (corev1.ResourceRequirements, error) {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}
	fields := []struct {
		value string
		name  corev1.ResourceName
		list  corev1.ResourceList
		text  string
	}{
		{item.CPURequest, corev1.ResourceCPU, requests, "CPU 预留"},
		{item.CPULimit, corev1.ResourceCPU, limits, "CPU 上限"},
		{item.MemoryRequest, corev1.ResourceMemory, requests, "内存预留"},
		{item.MemoryLimit, corev1.ResourceMemory, limits, "内存上限"},
	}
	for _, f := range fields {
		value := strings.TrimSpace(f.value)
		if value == "" {
			continue
		}
		qty, err := resource.ParseQuantity(value)
		if err != nil || qty.Sign() <= 0 {
			return corev1.ResourceRequirements{}, fmt.Errorf("%s的%s格式不正确: %s", label, f.text, value)
		}
		f.list[f.name] = qty
	}
	for name, req := range requests {
		if limit, ok := limits[name]; ok && req.Cmp(limit) > 0 {
			text := "内存"
			if name == corev1.ResourceCPU {
				text = "CPU"
			}
			return corev1.ResourceRequirements{}, fmt.Errorf("%s的%s预留不能大于上限", label, text)
		}
	}
	result := corev1.ResourceRequirements{}
	if len(requests) > 0 {
		result.Requests = requests
	}
	if len(limits) > 0 {
		result.Limits = limits
	}
	return result, nil
}

// withCreator 写入创建者注解
func withCreator(annotations map[string]string, creator string) map[string]string {
	creator = strings.TrimSpace(creator)
	if creator == "" {
		return annotations
	}
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[consts.CreatorAnnotation] = creator
	return annotations
}

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

func stripJobControllerLabels(labels map[string]string) map[string]string {
	result := copyStringMap(labels)
	for _, key := range jobControllerLabels {
		delete(result, key)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// listJobPods 按任务的选择器查询它创建的容器组
func listJobPods(ctx context.Context, clientSet *kubernetes.Clientset, job *batchv1.Job) ([]vo.JobPodVO, error) {
	if job.Spec.Selector == nil {
		return []vo.JobPodVO{}, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, err
	}
	pods, err := clientSet.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector:   selector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	result := make([]vo.JobPodVO, 0, len(pods.Items))
	for i := range pods.Items {
		result = append(result, vo.Pod2JobPodVO(&pods.Items[i]))
	}
	sortJobPods(result)
	attachJobPodMetrics(ctx, clientSet, job.Namespace, result)
	return result, nil
}

func sortJobPods(pods []vo.JobPodVO) {
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreateTime.After(pods[j].CreateTime)
	})
}

// listObjectEvents 按对象 UID 查询事件, 避免把同名旧对象的事件混进来; 失败时返回空列表
func listObjectEvents(ctx context.Context, clientSet *kubernetes.Clientset, namespace, uid string) []vo.EventVO {
	events, err := clientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.uid=" + uid,
	})
	if err != nil {
		logs.Warn("获取事件失败: %v", err)
		return []vo.EventVO{}
	}
	return vo.Events2VO(events)
}

// readJobRunRecords 读取重新执行前保存的运行记录
func readJobRunRecords(job *batchv1.Job) []vo.JobRunRecordVO {
	raw := job.Annotations[consts.JobRunRecordsAnnotation]
	if raw == "" {
		return nil
	}
	var records []vo.JobRunRecordVO
	if err := json.Unmarshal([]byte(raw), &records); err != nil {
		logs.Warn("解析任务 %s/%s 的运行记录失败: %v", job.Namespace, job.Name, err)
		return nil
	}
	return records
}

// jobRunRecords 运行记录: 当前这次在最前, 之后是历史记录(新的在前)
func jobRunRecords(job *batchv1.Job) []vo.JobRunRecordVO {
	history := readJobRunRecords(job)
	run := 1
	if n := len(history); n > 0 {
		run = history[n-1].Run + 1
	}
	result := make([]vo.JobRunRecordVO, 0, len(history)+1)
	result = append(result, vo.CurrentJobRunRecord(job, run))
	for i := len(history) - 1; i >= 0; i-- {
		record := history[i]
		record.Current = false
		result = append(result, record)
	}
	return result
}

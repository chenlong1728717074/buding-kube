package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
)

var (
	cronJobSrv  *CronJobService
	cronJobOnce sync.Once
)

type CronJobService struct {
}

func NewCronJobService() *CronJobService {
	return &CronJobService{}
}

func GetSingletonCronJobService() *CronJobService {
	cronJobOnce.Do(func() {
		cronJobSrv = NewCronJobService()
	})
	return cronJobSrv
}

var cronMacros = map[string]bool{
	"@yearly": true, "@annually": true, "@monthly": true, "@weekly": true,
	"@daily": true, "@midnight": true, "@hourly": true,
}

var cronFieldRegexp = regexp.MustCompile(`^[0-9A-Za-z*?/,\-]+$`)

// normalizeSchedule 初步校验 Cron 表达式(完整校验由 apiserver 完成)
func normalizeSchedule(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("请填写调度规则")
	}
	if strings.Contains(value, "TZ=") {
		return "", errors.New("请在「时区」中设置时区，不要写在调度规则里")
	}
	if strings.HasPrefix(value, "@") {
		if cronMacros[value] || strings.HasPrefix(value, "@every ") {
			return value, nil
		}
		return "", fmt.Errorf("不支持的调度规则: %s", value)
	}
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return "", errors.New("调度规则需要 5 个字段：分 时 日 月 周，例如 0 2 * * *")
	}
	for _, f := range fields {
		if !cronFieldRegexp.MatchString(f) {
			return "", fmt.Errorf("调度规则中的 %s 不合法", f)
		}
	}
	return strings.Join(fields, " "), nil
}

func normalizeConcurrencyPolicy(value string) (batchv1.ConcurrencyPolicy, error) {
	switch p := batchv1.ConcurrencyPolicy(strings.TrimSpace(value)); p {
	case "":
		return batchv1.AllowConcurrent, nil
	case batchv1.AllowConcurrent, batchv1.ForbidConcurrent, batchv1.ReplaceConcurrent:
		return p, nil
	default:
		return "", fmt.Errorf("不支持的并发策略: %s, 可选 Allow / Forbid / Replace", value)
	}
}

func (s *CronJobService) List(query dto.CronJobPageQueryBaseDTO) ([]vo.CronJobVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.BatchV1().CronJobs(query.Namespace).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取CronJob失败: %v", err)
		return nil, fmt.Errorf("获取CronJob失败: %v", err)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Name)
	}
	status := strings.TrimSpace(query.Status)
	result := make([]vo.CronJobVO, 0, len(items.Items))
	for i := range items.Items {
		item := items.Items[i]
		if keyword != "" && !strings.Contains(item.Name, keyword) {
			continue
		}
		vi := vo.CronJob2VO(item)
		if (status == "Suspended" && !vi.Suspend) || (status == "Active" && vi.Suspend) {
			continue
		}
		result = append(result, vi)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func (s *CronJobService) getCronJob(ctx context.Context, clusterId, namespace, name string) (*kubernetes.Clientset, *batchv1.CronJob, error) {
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, nil, errors.New("命名空间和名称不能为空")
	}
	clientSet, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, nil, err
	}
	cj, err := clientSet.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, fmt.Errorf("定时任务 %s 不存在", name)
		}
		return nil, nil, fmt.Errorf("获取定时任务失败: %v", err)
	}
	return clientSet, cj, nil
}

// Get 定时任务详情: 属性、它创建的任务(运行记录)、容器组、容器模板和事件
func (s *CronJobService) Get(ctx context.Context, base dto.CronJobBaseDTO) (*vo.CronJobDetailVO, error) {
	clientSet, cj, err := s.getCronJob(ctx, base.ClusterId, base.Namespace, base.Name)
	if err != nil {
		return nil, err
	}
	detail := vo.CronJob2DetailVO(cj)

	owned := map[string]bool{}
	jobs, err := clientSet.BatchV1().Jobs(cj.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Warn("获取定时任务 %s/%s 的任务失败: %v", cj.Namespace, cj.Name, err)
	} else {
		for i := range jobs.Items {
			job := &jobs.Items[i]
			for _, ref := range job.OwnerReferences {
				if ref.UID == cj.UID {
					owned[string(job.UID)] = true
					detail.Jobs = append(detail.Jobs, vo.Job2VO(*job))
					break
				}
			}
		}
		sort.Slice(detail.Jobs, func(i, j int) bool {
			return detail.Jobs[i].CreateTime.After(detail.Jobs[j].CreateTime)
		})
	}

	if len(owned) > 0 {
		pods, err := clientSet.CoreV1().Pods(cj.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
		if err != nil {
			logs.Warn("获取定时任务 %s/%s 的容器组失败: %v", cj.Namespace, cj.Name, err)
		} else {
			for i := range pods.Items {
				pod := &pods.Items[i]
				for _, ref := range pod.OwnerReferences {
					if ref.Kind == "Job" && owned[string(ref.UID)] {
						detail.Pods = append(detail.Pods, vo.Pod2JobPodVO(pod))
						break
					}
				}
			}
			sortJobPods(detail.Pods)
			attachJobPodMetrics(ctx, clientSet, cj.Namespace, detail.Pods)
		}
	}

	detail.Events = listObjectEvents(ctx, clientSet, cj.Namespace, string(cj.UID))
	return &detail, nil
}

// Create 创建定时任务
func (s *CronJobService) Create(ctx context.Context, req dto.CronJobCreateDTO, creator string) error {
	name := strings.TrimSpace(req.Name)
	if err := validateJobName(name, maxCronJobNameLength); err != nil {
		return err
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	schedule, err := normalizeSchedule(req.Schedule)
	if err != nil {
		return err
	}
	policy, err := normalizeConcurrencyPolicy(req.ConcurrencyPolicy)
	if err != nil {
		return err
	}
	labels := cleanStringMap(req.Labels)
	jobSpec, err := buildJobSpec(req.Template, copyStringMap(labels))
	if err != nil {
		return err
	}
	successLimit, failLimit := int32(3), int32(1)
	if req.SuccessfulJobsHistoryLimit != nil {
		if *req.SuccessfulJobsHistoryLimit < 0 {
			return errors.New("保留的成功任务数不能小于 0")
		}
		successLimit = *req.SuccessfulJobsHistoryLimit
	}
	if req.FailedJobsHistoryLimit != nil {
		if *req.FailedJobsHistoryLimit < 0 {
			return errors.New("保留的失败任务数不能小于 0")
		}
		failLimit = *req.FailedJobsHistoryLimit
	}
	suspend := req.Suspend

	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: withCreator(withAliasDescribe(nil, req.Alias, req.Describe), creator),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   schedule,
			ConcurrencyPolicy:          policy,
			Suspend:                    &suspend,
			SuccessfulJobsHistoryLimit: &successLimit,
			FailedJobsHistoryLimit:     &failLimit,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: copyStringMap(labels)},
				Spec:       jobSpec,
			},
		},
	}
	if tz := strings.TrimSpace(req.TimeZone); tz != "" {
		cj.Spec.TimeZone = &tz
	}
	if req.StartingDeadlineSeconds != nil && *req.StartingDeadlineSeconds > 0 {
		v := *req.StartingDeadlineSeconds
		cj.Spec.StartingDeadlineSeconds = &v
	}

	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	if _, err = clientSet.BatchV1().CronJobs(namespace).Create(ctx, cj, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("定时任务 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建CronJob失败: %v", err)
		return fmt.Errorf("创建定时任务失败: %v", err)
	}
	logs.Info("CronJob %s/%s 创建成功", namespace, name)
	return nil
}

// Update 编辑别名、描述和定时设置
func (s *CronJobService) Update(ctx context.Context, req dto.CronJobUpdateDTO) error {
	clientSet, cj, err := s.getCronJob(ctx, req.ClusterId, req.Namespace, req.Name)
	if err != nil {
		return err
	}
	cj.Annotations = withAliasDescribe(cj.Annotations, req.Alias, req.Describe)
	if strings.TrimSpace(req.Schedule) != "" {
		schedule, err := normalizeSchedule(req.Schedule)
		if err != nil {
			return err
		}
		cj.Spec.Schedule = schedule
	}
	if req.TimeZone != nil {
		if tz := strings.TrimSpace(*req.TimeZone); tz == "" {
			cj.Spec.TimeZone = nil
		} else {
			cj.Spec.TimeZone = &tz
		}
	}
	if strings.TrimSpace(req.ConcurrencyPolicy) != "" {
		policy, err := normalizeConcurrencyPolicy(req.ConcurrencyPolicy)
		if err != nil {
			return err
		}
		cj.Spec.ConcurrencyPolicy = policy
	}
	if req.Suspend != nil {
		v := *req.Suspend
		cj.Spec.Suspend = &v
	}
	if req.StartingDeadlineSeconds != nil {
		if *req.StartingDeadlineSeconds <= 0 {
			cj.Spec.StartingDeadlineSeconds = nil
		} else {
			v := *req.StartingDeadlineSeconds
			cj.Spec.StartingDeadlineSeconds = &v
		}
	}
	if req.SuccessfulJobsHistoryLimit != nil {
		if *req.SuccessfulJobsHistoryLimit < 0 {
			return errors.New("保留的成功任务数不能小于 0")
		}
		v := *req.SuccessfulJobsHistoryLimit
		cj.Spec.SuccessfulJobsHistoryLimit = &v
	}
	if req.FailedJobsHistoryLimit != nil {
		if *req.FailedJobsHistoryLimit < 0 {
			return errors.New("保留的失败任务数不能小于 0")
		}
		v := *req.FailedJobsHistoryLimit
		cj.Spec.FailedJobsHistoryLimit = &v
	}
	if _, err = clientSet.BatchV1().CronJobs(cj.Namespace).Update(ctx, cj, metav1.UpdateOptions{}); err != nil {
		logs.Error("更新CronJob失败: %v", err)
		return fmt.Errorf("更新定时任务失败: %v", err)
	}
	return nil
}

// Suspend 暂停或启用定时任务; 暂停只影响之后的调度, 已经在运行的任务不受影响
func (s *CronJobService) Suspend(ctx context.Context, req dto.CronJobSuspendDTO) error {
	clientSet, cj, err := s.getCronJob(ctx, req.ClusterId, req.Namespace, req.Name)
	if err != nil {
		return err
	}
	suspend := req.Suspend
	cj.Spec.Suspend = &suspend
	if _, err = clientSet.BatchV1().CronJobs(cj.Namespace).Update(ctx, cj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("更新定时任务失败: %v", err)
	}
	return nil
}

// Trigger 立即执行一次: 按任务模板创建任务, 与 kubectl create job --from=cronjob/<name> 相同
func (s *CronJobService) Trigger(ctx context.Context, base dto.CronJobBaseDTO, creator string) (string, error) {
	clientSet, cj, err := s.getCronJob(ctx, base.ClusterId, base.Namespace, base.Name)
	if err != nil {
		return "", err
	}
	annotations := copyStringMap(cj.Spec.JobTemplate.Annotations)
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[consts.CronJobInstantiateAnnotation] = "manual"
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            manualJobName(cj.Name),
			Namespace:       cj.Namespace,
			Labels:          copyStringMap(cj.Spec.JobTemplate.Labels),
			Annotations:     withCreator(annotations, creator),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cj, batchv1.SchemeGroupVersion.WithKind("CronJob"))},
		},
		Spec: *cj.Spec.JobTemplate.Spec.DeepCopy(),
	}
	created, err := clientSet.BatchV1().Jobs(cj.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		logs.Error("手动触发CronJob %s/%s 失败: %v", cj.Namespace, cj.Name, err)
		return "", fmt.Errorf("创建任务失败: %v", err)
	}
	return created.Name, nil
}

// manualJobName 手动触发的任务名: <定时任务名>-manual-<随机 5 位>, 总长度不超过 63
func manualJobName(cronJobName string) string {
	suffix := "-manual-" + utilrand.String(5)
	if maxLen := 63 - len(suffix); len(cronJobName) > maxLen {
		cronJobName = strings.TrimRight(cronJobName[:maxLen], "-.")
	}
	return cronJobName + suffix
}

// Delete 删除定时任务, 同时删除它创建的任务和容器组
func (s *CronJobService) Delete(ctx context.Context, base dto.CronJobBaseDTO) error {
	if strings.TrimSpace(base.Namespace) == "" || strings.TrimSpace(base.Name) == "" {
		return errors.New("命名空间和名称不能为空")
	}
	clientSet, err := ClusterMap.GetClientSet(base.ClusterId)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	err = clientSet.BatchV1().CronJobs(base.Namespace).Delete(ctx, base.Name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil && !apierrors.IsNotFound(err) {
		logs.Error("删除CronJob失败: %v", err)
		return fmt.Errorf("删除定时任务失败: %v", err)
	}
	return nil
}

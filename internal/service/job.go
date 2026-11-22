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
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	jobSrv  *JobService
	jobOnce sync.Once
)

type JobService struct {
}

func NewJobService() *JobService {
	return &JobService{}
}

func GetSingletonJobService() *JobService {
	jobOnce.Do(func() {
		jobSrv = NewJobService()
	})
	return jobSrv
}

func (s *JobService) List(ctx context.Context, query dto.JobPageQueryBaseDTO) ([]vo.JobVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.BatchV1().Jobs(query.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取Job失败: %v", err)
		return nil, fmt.Errorf("获取Job失败: %v", err)
	}
	// 前端搜索框使用 keyword, 旧版本使用 name, 两个都支持
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(query.Name)
	}
	status := strings.TrimSpace(query.Status)
	result := make([]vo.JobVO, 0, len(items.Items))
	for i := range items.Items {
		item := items.Items[i]
		if keyword != "" && !strings.Contains(item.Name, keyword) {
			continue
		}
		vi := vo.Job2VO(item)
		if status != "" && vi.Status != status {
			continue
		}
		result = append(result, vi)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

func (s *JobService) getJob(ctx context.Context, clusterId, namespace, name string) (*kubernetes.Clientset, *batchv1.Job, error) {
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, nil, errors.New("命名空间和名称不能为空")
	}
	clientSet, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, nil, err
	}
	job, err := clientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, fmt.Errorf("任务 %s 不存在", name)
		}
		return nil, nil, fmt.Errorf("获取任务失败: %v", err)
	}
	return clientSet, job, nil
}

// Get 任务详情: 属性、运行记录、容器组、容器模板和事件
func (s *JobService) Get(ctx context.Context, base dto.JobBaseDTO) (*vo.JobDetailVO, error) {
	clientSet, job, err := s.getJob(ctx, base.ClusterId, base.Namespace, base.Name)
	if err != nil {
		return nil, err
	}
	detail := vo.Job2DetailVO(job)
	detail.Records = jobRunRecords(job)
	if pods, err := listJobPods(ctx, clientSet, job); err == nil {
		detail.Pods = pods
	} else {
		logs.Warn("获取任务 %s/%s 的容器组失败: %v", job.Namespace, job.Name, err)
	}
	detail.Events = listObjectEvents(ctx, clientSet, job.Namespace, string(job.UID))
	return &detail, nil
}

// Create 创建任务
func (s *JobService) Create(ctx context.Context, req dto.JobCreateDTO, creator string) error {
	name := strings.TrimSpace(req.Name)
	if err := validateJobName(name, 63); err != nil {
		return err
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	labels := cleanStringMap(req.Labels)
	spec, err := buildJobSpec(req.Template, copyStringMap(labels))
	if err != nil {
		return err
	}
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: withCreator(withAliasDescribe(nil, req.Alias, req.Describe), creator),
		},
		Spec: spec,
	}
	if _, err = clientSet.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("任务 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建Job失败: %v", err)
		return fmt.Errorf("创建任务失败: %v", err)
	}
	logs.Info("Job %s/%s 创建成功", namespace, name)
	return nil
}

// Update 编辑别名、描述、并行数量、最大运行时间和自动清理时间
func (s *JobService) Update(ctx context.Context, req dto.JobUpdateDTO) error {
	clientSet, job, err := s.getJob(ctx, req.ClusterId, req.Namespace, req.Name)
	if err != nil {
		return err
	}
	job.Annotations = withAliasDescribe(job.Annotations, req.Alias, req.Describe)
	if req.Parallelism != nil {
		if *req.Parallelism < 0 {
			return errors.New("并行数量不能小于 0")
		}
		v := *req.Parallelism
		job.Spec.Parallelism = &v
	}
	if req.ActiveDeadlineSeconds != nil {
		if *req.ActiveDeadlineSeconds <= 0 {
			job.Spec.ActiveDeadlineSeconds = nil
		} else {
			v := *req.ActiveDeadlineSeconds
			job.Spec.ActiveDeadlineSeconds = &v
		}
	}
	if req.TTLSecondsAfterFinished != nil {
		if *req.TTLSecondsAfterFinished < 0 {
			job.Spec.TTLSecondsAfterFinished = nil
		} else {
			v := *req.TTLSecondsAfterFinished
			job.Spec.TTLSecondsAfterFinished = &v
		}
	}
	if _, err = clientSet.BatchV1().Jobs(job.Namespace).Update(ctx, job, metav1.UpdateOptions{}); err != nil {
		logs.Error("更新Job失败: %v", err)
		return fmt.Errorf("更新任务失败: %v", err)
	}
	return nil
}

// Delete 删除任务, 同时删除它创建的容器组
func (s *JobService) Delete(ctx context.Context, base dto.JobBaseDTO) error {
	if strings.TrimSpace(base.Namespace) == "" || strings.TrimSpace(base.Name) == "" {
		return errors.New("命名空间和名称不能为空")
	}
	clientSet, err := ClusterMap.GetClientSet(base.ClusterId)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	err = clientSet.BatchV1().Jobs(base.Namespace).Delete(ctx, base.Name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil && !apierrors.IsNotFound(err) {
		logs.Error("删除Job失败: %v", err)
		return fmt.Errorf("删除任务失败: %v", err)
	}
	return nil
}

// Rerun 重新执行任务.
// 任务创建后规格不能修改, 和 KubeSphere 一样通过「删除后按原规格重新创建」实现;
// 本次运行的结果写入新任务的注解, 详情页的运行记录可以看到每一次的执行情况
func (s *JobService) Rerun(ctx context.Context, base dto.JobBaseDTO) error {
	clientSet, job, err := s.getJob(ctx, base.ClusterId, base.Namespace, base.Name)
	if err != nil {
		return err
	}
	if job.Status.Active > 0 {
		return errors.New("任务正在运行，结束后才能重新执行")
	}

	history := readJobRunRecords(job)
	run := 1
	if n := len(history); n > 0 {
		run = history[n-1].Run + 1
	}
	finished := vo.CurrentJobRunRecord(job, run)
	finished.Current = false
	history = append(history, finished)
	if len(history) > maxJobRunRecords {
		history = history[len(history)-maxJobRunRecords:]
	}
	raw, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("保存运行记录失败: %w", err)
	}

	annotations := copyStringMap(job.Annotations)
	if annotations == nil {
		annotations = map[string]string{}
	}
	delete(annotations, "batch.kubernetes.io/job-tracking")
	annotations[consts.JobRunRecordsAnnotation] = string(raw)

	// 选择器和控制器标签由 apiserver 重新生成
	spec := job.Spec.DeepCopy()
	spec.Selector = nil
	spec.ManualSelector = nil
	spec.Suspend = nil
	spec.Template.Labels = stripJobControllerLabels(spec.Template.Labels)
	newJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            job.Name,
			Namespace:       job.Namespace,
			Labels:          stripJobControllerLabels(job.Labels),
			Annotations:     annotations,
			OwnerReferences: job.OwnerReferences,
		},
		Spec: *spec,
	}

	jobs := clientSet.BatchV1().Jobs(job.Namespace)
	propagation := metav1.DeletePropagationBackground
	err = jobs.Delete(ctx, job.Name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
		Preconditions:     &metav1.Preconditions{UID: &job.UID},
	})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("删除旧任务失败: %v", err)
	}
	err = wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		_, getErr := jobs.Get(ctx, job.Name, metav1.GetOptions{})
		return apierrors.IsNotFound(getErr), nil
	})
	if err != nil {
		return errors.New("等待旧任务删除超时，请稍后刷新后再试")
	}
	if _, err = jobs.Create(ctx, newJob, metav1.CreateOptions{}); err != nil {
		logs.Error("重新创建Job %s/%s 失败: %v", job.Namespace, job.Name, err)
		return fmt.Errorf("旧任务已删除，但重新创建失败: %v", err)
	}
	logs.Info("Job %s/%s 已重新执行(第 %d 次)", job.Namespace, job.Name, run+1)
	return nil
}

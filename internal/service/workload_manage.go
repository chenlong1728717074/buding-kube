package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	WorkloadDeployment  = "deployment"
	WorkloadStatefulSet = "statefulset"
	WorkloadDaemonSet   = "daemonset"

	restartedAtAnnotation        = "kubectl.kubernetes.io/restartedAt"
	deploymentRevisionAnnotation = "deployment.kubernetes.io/revision"
	changeCauseAnnotation        = "kubernetes.io/change-cause"
)

var workloadKindNames = map[string]string{
	WorkloadDeployment:  "Deployment",
	WorkloadStatefulSet: "StatefulSet",
	WorkloadDaemonSet:   "DaemonSet",
}

func normalizeWorkloadKind(kind string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	if _, ok := workloadKindNames[k]; !ok {
		return "", fmt.Errorf("不支持的工作负载类型: %s", kind)
	}
	return k, nil
}

// workloadRef 三种工作负载的公共部分; meta / template / selector 指向原对象, 修改后调用 save 保存
type workloadRef struct {
	kind     string
	meta     *metav1.ObjectMeta
	template *corev1.PodTemplateSpec
	selector *metav1.LabelSelector
	deploy   *appsv1.Deployment
	sts      *appsv1.StatefulSet
	ds       *appsv1.DaemonSet
}

func (r *workloadRef) label() string {
	return workloadKindNames[r.kind]
}

func getWorkload(ctx context.Context, cs *kubernetes.Clientset, kind, namespace, name string) (*workloadRef, error) {
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, errors.New("命名空间和名称不能为空")
	}
	ref := &workloadRef{kind: kind}
	var err error
	switch kind {
	case WorkloadDeployment:
		ref.deploy, err = cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			ref.meta, ref.template, ref.selector = &ref.deploy.ObjectMeta, &ref.deploy.Spec.Template, ref.deploy.Spec.Selector
		}
	case WorkloadStatefulSet:
		ref.sts, err = cs.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			ref.meta, ref.template, ref.selector = &ref.sts.ObjectMeta, &ref.sts.Spec.Template, ref.sts.Spec.Selector
		}
	default:
		ref.ds, err = cs.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			ref.meta, ref.template, ref.selector = &ref.ds.ObjectMeta, &ref.ds.Spec.Template, ref.ds.Spec.Selector
		}
	}
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("%s %s 不存在", ref.label(), name)
		}
		return nil, fmt.Errorf("获取%s失败: %v", ref.label(), err)
	}
	return ref, nil
}

func (r *workloadRef) save(ctx context.Context, cs *kubernetes.Clientset) error {
	var err error
	namespace := r.meta.Namespace
	switch r.kind {
	case WorkloadDeployment:
		_, err = cs.AppsV1().Deployments(namespace).Update(ctx, r.deploy, metav1.UpdateOptions{})
	case WorkloadStatefulSet:
		_, err = cs.AppsV1().StatefulSets(namespace).Update(ctx, r.sts, metav1.UpdateOptions{})
	default:
		_, err = cs.AppsV1().DaemonSets(namespace).Update(ctx, r.ds, metav1.UpdateOptions{})
	}
	if err != nil {
		if apierrors.IsConflict(err) {
			return fmt.Errorf("%s 已被修改，请刷新后重试", r.label())
		}
		return fmt.Errorf("更新%s失败: %v", r.label(), err)
	}
	return nil
}

var (
	workloadSrv  *WorkloadService
	workloadOnce sync.Once
)

// WorkloadService Deployment / StatefulSet / DaemonSet 的详情、创建和常用操作
type WorkloadService struct {
}

func GetSingletonWorkloadService() *WorkloadService {
	workloadOnce.Do(func() {
		workloadSrv = &WorkloadService{}
	})
	return workloadSrv
}

func (s *WorkloadService) open(ctx context.Context, t dto.WorkloadTargetDTO) (*kubernetes.Clientset, *workloadRef, error) {
	kind, err := normalizeWorkloadKind(t.Kind)
	if err != nil {
		return nil, nil, err
	}
	cs, err := ClusterMap.GetClientSet(t.ClusterId)
	if err != nil {
		return nil, nil, err
	}
	ref, err := getWorkload(ctx, cs, kind, t.Namespace, t.Name)
	if err != nil {
		return nil, nil, err
	}
	return cs, ref, nil
}

// Get 详情: 属性、容器组、版本记录、服务和事件
func (s *WorkloadService) Get(ctx context.Context, t dto.WorkloadTargetDTO) (*vo.WorkloadDetailVO, error) {
	cs, ref, err := s.open(ctx, t)
	if err != nil {
		return nil, err
	}
	var detail vo.WorkloadDetailVO
	switch ref.kind {
	case WorkloadDeployment:
		detail = vo.Deployment2DetailVO(ref.deploy)
	case WorkloadStatefulSet:
		detail = vo.StatefulSet2DetailVO(ref.sts)
	default:
		detail = vo.DaemonSet2DetailVO(ref.ds)
	}
	namespace := ref.meta.Namespace
	if pods, podErr := listPodsBySelector(ctx, cs, namespace, ref.selector); podErr == nil {
		detail.Pods = pods
	} else {
		logs.Warn("获取 %s/%s 的容器组失败: %v", namespace, ref.meta.Name, podErr)
	}
	if revisions, current, revErr := workloadRevisions(ctx, cs, ref); revErr == nil {
		detail.Revisions = revisions
		if current > 0 {
			detail.CurrentRevision = current
		}
	} else {
		logs.Warn("获取 %s/%s 的版本记录失败: %v", namespace, ref.meta.Name, revErr)
	}
	detail.Services = servicesSelecting(ctx, cs, namespace, ref.template.Labels)
	detail.Events = listObjectEvents(ctx, cs, namespace, string(ref.meta.UID))
	return &detail, nil
}

func ownedReplicaSets(ctx context.Context, cs *kubernetes.Clientset, d *appsv1.Deployment) ([]*appsv1.ReplicaSet, error) {
	list, err := cs.AppsV1().ReplicaSets(d.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	result := make([]*appsv1.ReplicaSet, 0)
	for i := range list.Items {
		if metav1.IsControlledBy(&list.Items[i], d) {
			result = append(result, &list.Items[i])
		}
	}
	return result, nil
}

func ownedControllerRevisions(ctx context.Context, cs *kubernetes.Clientset, ref *workloadRef) ([]*appsv1.ControllerRevision, error) {
	list, err := cs.AppsV1().ControllerRevisions(ref.meta.Namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	result := make([]*appsv1.ControllerRevision, 0)
	for i := range list.Items {
		if metav1.IsControlledBy(&list.Items[i], ref.meta) {
			result = append(result, &list.Items[i])
		}
	}
	return result, nil
}

// revisionTemplate ControllerRevision 中保存的容器组模板
func revisionTemplate(cr *appsv1.ControllerRevision) *corev1.PodTemplateSpec {
	var data struct {
		Spec struct {
			Template corev1.PodTemplateSpec `json:"template"`
		} `json:"spec"`
	}
	if len(cr.Data.Raw) == 0 || json.Unmarshal(cr.Data.Raw, &data) != nil {
		return nil
	}
	return &data.Spec.Template
}

// workloadRevisions 版本记录(新的在前)和当前版本号
func workloadRevisions(ctx context.Context, cs *kubernetes.Clientset, ref *workloadRef) ([]vo.WorkloadRevisionVO, int64, error) {
	result := make([]vo.WorkloadRevisionVO, 0)
	var current int64
	if ref.kind == WorkloadDeployment {
		rsList, err := ownedReplicaSets(ctx, cs, ref.deploy)
		if err != nil {
			return nil, 0, err
		}
		current, _ = strconv.ParseInt(ref.deploy.Annotations[deploymentRevisionAnnotation], 10, 64)
		for _, rs := range rsList {
			rev, _ := strconv.ParseInt(rs.Annotations[deploymentRevisionAnnotation], 10, 64)
			result = append(result, vo.WorkloadRevisionVO{
				Revision:      rev,
				Name:          rs.Name,
				Images:        vo.TemplateImages(&rs.Spec.Template),
				Replicas:      rs.Status.Replicas,
				ReadyReplicas: rs.Status.ReadyReplicas,
				Current:       rev == current,
				ChangeCause:   rs.Annotations[changeCauseAnnotation],
				CreateTime:    rs.CreationTimestamp.Time,
			})
		}
	} else {
		crs, err := ownedControllerRevisions(ctx, cs, ref)
		if err != nil {
			return nil, 0, err
		}
		currentName := ""
		if ref.kind == WorkloadStatefulSet {
			currentName = ref.sts.Status.UpdateRevision
		}
		var maxRev int64
		for _, cr := range crs {
			if cr.Revision > maxRev {
				maxRev = cr.Revision
			}
		}
		for _, cr := range crs {
			isCurrent := cr.Name == currentName || (currentName == "" && cr.Revision == maxRev)
			if isCurrent {
				current = cr.Revision
			}
			images := []string{}
			if tpl := revisionTemplate(cr); tpl != nil {
				images = vo.TemplateImages(tpl)
			}
			result = append(result, vo.WorkloadRevisionVO{
				Revision:    cr.Revision,
				Name:        cr.Name,
				Images:      images,
				Current:     isCurrent,
				ChangeCause: cr.Annotations[changeCauseAnnotation],
				CreateTime:  cr.CreationTimestamp.Time,
			})
		}
	}
	vo.SortRevisions(result)
	return result, current, nil
}

// UpdateInfo 编辑别名和描述
func (s *WorkloadService) UpdateInfo(ctx context.Context, req dto.WorkloadInfoDTO) error {
	cs, ref, err := s.open(ctx, req.WorkloadTargetDTO)
	if err != nil {
		return err
	}
	ref.meta.Annotations = withAliasDescribe(ref.meta.Annotations, req.Alias, req.Describe)
	return ref.save(ctx, cs)
}

// Scale 调整副本数
func (s *WorkloadService) Scale(ctx context.Context, req dto.WorkloadScaleDTO) error {
	if req.Replicas < 0 || req.Replicas > 1000 {
		return errors.New("副本数应在 0 到 1000 之间")
	}
	cs, ref, err := s.open(ctx, req.WorkloadTargetDTO)
	if err != nil {
		return err
	}
	replicas := req.Replicas
	switch ref.kind {
	case WorkloadDeployment:
		ref.deploy.Spec.Replicas = &replicas
	case WorkloadStatefulSet:
		ref.sts.Spec.Replicas = &replicas
	default:
		return errors.New("DaemonSet 在每个符合条件的节点上运行一个容器组，不能调整副本数")
	}
	return ref.save(ctx, cs)
}

// UpdateImages 更新容器镜像
func (s *WorkloadService) UpdateImages(ctx context.Context, req dto.WorkloadImageDTO) error {
	cs, ref, err := s.open(ctx, req.WorkloadTargetDTO)
	if err != nil {
		return err
	}
	changes := make([]string, 0)
	apply := func(containers []corev1.Container) error {
		for i := range containers {
			c := &containers[i]
			image := strings.TrimSpace(req.Images[c.Name])
			if image == "" || image == c.Image {
				continue
			}
			if strings.ContainsAny(image, " \t") {
				return fmt.Errorf("容器 %s 的镜像不能包含空格", c.Name)
			}
			c.Image = image
			changes = append(changes, c.Name+"="+image)
		}
		return nil
	}
	if err = apply(ref.template.Spec.InitContainers); err != nil {
		return err
	}
	if err = apply(ref.template.Spec.Containers); err != nil {
		return err
	}
	if len(changes) == 0 {
		return errors.New("镜像没有变化")
	}
	if ref.meta.Annotations == nil {
		ref.meta.Annotations = map[string]string{}
	}
	// Deployment 会把这个注解复制到 ReplicaSet, 在版本记录中显示
	ref.meta.Annotations[changeCauseAnnotation] = "更新镜像 " + strings.Join(changes, ", ")
	return ref.save(ctx, cs)
}

// Restart 重新创建容器组, 与 kubectl rollout restart 相同
func (s *WorkloadService) Restart(ctx context.Context, t dto.WorkloadTargetDTO) error {
	cs, ref, err := s.open(ctx, t)
	if err != nil {
		return err
	}
	if ref.template.Annotations == nil {
		ref.template.Annotations = map[string]string{}
	}
	ref.template.Annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
	return ref.save(ctx, cs)
}

// Rollback 回退到指定版本
func (s *WorkloadService) Rollback(ctx context.Context, req dto.WorkloadRollbackDTO) error {
	cs, ref, err := s.open(ctx, req.WorkloadTargetDTO)
	if err != nil {
		return err
	}
	if ref.kind == WorkloadDeployment {
		rsList, rsErr := ownedReplicaSets(ctx, cs, ref.deploy)
		if rsErr != nil {
			return fmt.Errorf("获取版本记录失败: %v", rsErr)
		}
		current, _ := strconv.ParseInt(ref.deploy.Annotations[deploymentRevisionAnnotation], 10, 64)
		if req.Revision == current {
			return errors.New("已经是当前版本")
		}
		for _, rs := range rsList {
			rev, _ := strconv.ParseInt(rs.Annotations[deploymentRevisionAnnotation], 10, 64)
			if rev != req.Revision {
				continue
			}
			tpl := rs.Spec.Template.DeepCopy()
			delete(tpl.Labels, appsv1.DefaultDeploymentUniqueLabelKey)
			ref.deploy.Spec.Template = *tpl
			if ref.deploy.Annotations == nil {
				ref.deploy.Annotations = map[string]string{}
			}
			ref.deploy.Annotations[changeCauseAnnotation] = fmt.Sprintf("回退到版本 #%d", req.Revision)
			return ref.save(ctx, cs)
		}
		return fmt.Errorf("版本 #%d 不存在，可能已被清理", req.Revision)
	}

	crs, err := ownedControllerRevisions(ctx, cs, ref)
	if err != nil {
		return fmt.Errorf("获取版本记录失败: %v", err)
	}
	_, current, _ := workloadRevisions(ctx, cs, ref)
	if req.Revision == current {
		return errors.New("已经是当前版本")
	}
	for _, cr := range crs {
		if cr.Revision != req.Revision {
			continue
		}
		if len(cr.Data.Raw) == 0 {
			return errors.New("版本数据为空，无法回退")
		}
		// ControllerRevision 中保存的是容器组模板的补丁, 与 kubectl rollout undo 相同
		if ref.kind == WorkloadStatefulSet {
			_, err = cs.AppsV1().StatefulSets(ref.meta.Namespace).Patch(ctx, ref.meta.Name, types.StrategicMergePatchType, cr.Data.Raw, metav1.PatchOptions{})
		} else {
			_, err = cs.AppsV1().DaemonSets(ref.meta.Namespace).Patch(ctx, ref.meta.Name, types.StrategicMergePatchType, cr.Data.Raw, metav1.PatchOptions{})
		}
		if err != nil {
			return fmt.Errorf("回退失败: %v", err)
		}
		return nil
	}
	return fmt.Errorf("版本 #%d 不存在，可能已被清理", req.Revision)
}

// Delete 删除工作负载, 同时删除它的容器组; StatefulSet 的持久卷声明会保留
func (s *WorkloadService) Delete(ctx context.Context, t dto.WorkloadTargetDTO) error {
	kind, err := normalizeWorkloadKind(t.Kind)
	if err != nil {
		return err
	}
	cs, err := ClusterMap.GetClientSet(t.ClusterId)
	if err != nil {
		return err
	}
	propagation := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{PropagationPolicy: &propagation}
	switch kind {
	case WorkloadDeployment:
		err = cs.AppsV1().Deployments(t.Namespace).Delete(ctx, t.Name, opts)
	case WorkloadStatefulSet:
		err = cs.AppsV1().StatefulSets(t.Namespace).Delete(ctx, t.Name, opts)
	default:
		err = cs.AppsV1().DaemonSets(t.Namespace).Delete(ctx, t.Name, opts)
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("删除%s失败: %v", workloadKindNames[kind], err)
	}
	return nil
}

// Selectors 命名空间中的工作负载及其容器组标签, 创建服务时选择用
func (s *WorkloadService) Selectors(ctx context.Context, clusterId, namespace string) ([]vo.WorkloadSelectorVO, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil, errors.New("请选择命名空间")
	}
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{ResourceVersion: "0"}
	result := make([]vo.WorkloadSelectorVO, 0)
	deploys, err := cs.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment 失败: %v", err)
	}
	for i := range deploys.Items {
		d := &deploys.Items[i]
		result = append(result, vo.NewWorkloadSelectorVO(WorkloadDeployment, d.Name, &d.Spec.Template))
	}
	if items, listErr := cs.AppsV1().StatefulSets(namespace).List(ctx, opts); listErr == nil {
		for i := range items.Items {
			st := &items.Items[i]
			result = append(result, vo.NewWorkloadSelectorVO(WorkloadStatefulSet, st.Name, &st.Spec.Template))
		}
	}
	if items, listErr := cs.AppsV1().DaemonSets(namespace).List(ctx, opts); listErr == nil {
		for i := range items.Items {
			d := &items.Items[i]
			result = append(result, vo.NewWorkloadSelectorVO(WorkloadDaemonSet, d.Name, &d.Spec.Template))
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

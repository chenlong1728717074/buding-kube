package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Create 分步创建容器组. 单独创建的容器组被删除后不会自动重建, 长期运行的应用建议使用 Deployment
func (s *PodService) Create(ctx context.Context, req dto.PodCreateDTO, creator string) error {
	name := strings.TrimSpace(req.Name)
	if len(validation.IsDNS1123Subdomain(name)) > 0 {
		return errors.New("名称不合法: 只能包含小写字母、数字、\"-\" 和 \".\"，以字母或数字开头和结尾")
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}
	labels := cleanStringMap(req.Labels)
	if err := validateLabelMap(labels, "标签"); err != nil {
		return err
	}
	nodeSelector := cleanStringMap(req.NodeSelector)
	if err := validateLabelMap(nodeSelector, "节点选择器"); err != nil {
		return err
	}
	annotations := cleanStringMap(req.Annotations)
	if err := validateAnnotationKeys(annotations); err != nil {
		return err
	}

	volumes, mountable, err := buildWorkloadVolumes(req.Volumes)
	if err != nil {
		return err
	}
	containers, initContainers, err := buildWorkloadContainers(req.Containers, mountable)
	if err != nil {
		return err
	}
	if req.SyncHostTimezone {
		if mountable[hostTimeVolume] {
			return fmt.Errorf("存储卷名称 %s 已被占用，无法同步主机时区", hostTimeVolume)
		}
		volumes = append(volumes, corev1.Volume{
			Name:         hostTimeVolume,
			VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc/localtime"}},
		})
		mountHostTime(containers)
		mountHostTime(initContainers)
	}

	restart := corev1.RestartPolicy(strings.TrimSpace(req.RestartPolicy))
	switch restart {
	case "":
		restart = corev1.RestartPolicyAlways
	case corev1.RestartPolicyAlways, corev1.RestartPolicyOnFailure, corev1.RestartPolicyNever:
	default:
		return fmt.Errorf("不支持的重启策略: %s", req.RestartPolicy)
	}

	spec := corev1.PodSpec{
		InitContainers:     initContainers,
		Containers:         containers,
		Volumes:            volumes,
		RestartPolicy:      restart,
		NodeName:           strings.TrimSpace(req.NodeName),
		NodeSelector:       nodeSelector,
		ServiceAccountName: strings.TrimSpace(req.ServiceAccountName),
		HostNetwork:        req.HostNetwork,
	}
	if req.HostNetwork {
		// 使用节点网络时仍然解析集群内的服务域名
		spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	for _, secret := range cleanStrings(req.ImagePullSecrets) {
		spec.ImagePullSecrets = append(spec.ImagePullSecrets,
			corev1.LocalObjectReference{Name: secret},
		)
	}
	if req.TerminationGracePeriodSeconds != nil && *req.TerminationGracePeriodSeconds >= 0 {
		v := *req.TerminationGracePeriodSeconds
		spec.TerminationGracePeriodSeconds = &v
	}

	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: withCreator(withAliasDescribe(annotations, req.Alias, req.Describe), creator),
		},
		Spec: spec,
	}
	if _, err = cs.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("容器组 %s/%s 已存在", namespace, name)
		}
		logs.Error("创建Pod失败: %v", err)
		return fmt.Errorf("创建容器组失败: %v", err)
	}
	logs.Info("Pod %s/%s 创建成功", namespace, name)
	return nil
}

func (s *PodService) getPod(ctx context.Context, clusterId, namespace, name string) (*corev1.Pod, error) {
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, err
	}
	pod, err := cs.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("容器组 %s 不存在", name)
		}
		return nil, fmt.Errorf("获取容器组失败: %v", err)
	}
	return pod, nil
}

func (s *PodService) savePod(ctx context.Context, clusterId string, pod *corev1.Pod) error {
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return err
	}
	if _, err = cs.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			return errors.New("容器组已被修改，请刷新后重试")
		}
		return fmt.Errorf("更新容器组失败: %v", err)
	}
	return nil
}

// UpdateInfo 编辑别名和描述(元数据可以随时修改)
func (s *PodService) UpdateInfo(ctx context.Context, req dto.PodInfoDTO) error {
	pod, err := s.getPod(ctx, req.ClusterId, req.Namespace, req.Name)
	if err != nil {
		return err
	}
	pod.Annotations = withAliasDescribe(pod.Annotations, req.Alias, req.Describe)
	return s.savePod(ctx, req.ClusterId, pod)
}

// UpdateImages 更新镜像. 容器组创建后只有镜像等少数字段可以修改, 修改后 kubelet 会重启对应的容器
func (s *PodService) UpdateImages(ctx context.Context, req dto.PodImageDTO) error {
	pod, err := s.getPod(ctx, req.ClusterId, req.Namespace, req.Name)
	if err != nil {
		return err
	}
	if owner := metav1.GetControllerOf(pod); owner != nil {
		logs.Info("容器组 %s/%s 由 %s/%s 管理, 直接修改镜像会在重建后失效", pod.Namespace, pod.Name, owner.Kind, owner.Name)
	}
	changed := 0
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
			changed++
		}
		return nil
	}
	if err = apply(pod.Spec.InitContainers); err != nil {
		return err
	}
	if err = apply(pod.Spec.Containers); err != nil {
		return err
	}
	if changed == 0 {
		return errors.New("镜像没有变化")
	}
	return s.savePod(ctx, req.ClusterId, pod)
}

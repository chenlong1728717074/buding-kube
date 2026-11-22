package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const hostTimeVolume = "host-time"

// Create 分步创建工作负载
func (s *WorkloadService) Create(ctx context.Context, req dto.WorkloadCreateDTO, creator string) error {
	kind, err := normalizeWorkloadKind(req.Kind)
	if err != nil {
		return err
	}
	kindName := workloadKindNames[kind]
	name := strings.TrimSpace(req.Name)
	// StatefulSet 控制器会创建 <名称>-<哈希> 的 ControllerRevision 标签值, 名称需要留出空间
	maxLen := 63
	if kind == WorkloadStatefulSet {
		maxLen = 52
	}
	if err = validateJobName(name, maxLen); err != nil {
		return err
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		return errors.New("请选择命名空间")
	}

	podLabels := cleanStringMap(req.Labels)
	if len(podLabels) == 0 {
		podLabels = map[string]string{"app": name}
	}
	if err = validateLabelMap(podLabels, "标签"); err != nil {
		return err
	}
	nodeSelector := cleanStringMap(req.NodeSelector)
	if err = validateLabelMap(nodeSelector, "节点选择器"); err != nil {
		return err
	}

	volumes, mountable, err := buildWorkloadVolumes(req.Volumes)
	if err != nil {
		return err
	}
	var claims []corev1.PersistentVolumeClaim
	if kind == WorkloadStatefulSet {
		if claims, err = buildClaimTemplates(req.VolumeClaimTemplates, mountable); err != nil {
			return err
		}
		for _, c := range claims {
			mountable[c.Name] = true
		}
	} else if len(req.VolumeClaimTemplates) > 0 {
		return errors.New("只有 StatefulSet 支持存储卷模板")
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

	podSpec := corev1.PodSpec{
		InitContainers:     initContainers,
		Containers:         containers,
		Volumes:            volumes,
		NodeSelector:       nodeSelector,
		ServiceAccountName: strings.TrimSpace(req.ServiceAccountName),
	}
	for _, secret := range cleanStrings(req.ImagePullSecrets) {
		podSpec.ImagePullSecrets = append(podSpec.ImagePullSecrets, corev1.LocalObjectReference{Name: secret})
	}
	if req.TerminationGracePeriodSeconds != nil && *req.TerminationGracePeriodSeconds >= 0 {
		v := *req.TerminationGracePeriodSeconds
		podSpec.TerminationGracePeriodSeconds = &v
	}
	if req.RevisionHistoryLimit != nil && *req.RevisionHistoryLimit < 0 {
		return errors.New("保留的历史版本数不能小于 0")
	}
	if req.MinReadySeconds < 0 {
		return errors.New("最短就绪时间不能小于 0")
	}

	annotations := cleanStringMap(req.Annotations)
	if err = validateAnnotationKeys(annotations); err != nil {
		return err
	}
	meta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      copyStringMap(podLabels),
		Annotations: withCreator(withAliasDescribe(annotations, req.Alias, req.Describe), creator),
	}
	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Labels: copyStringMap(podLabels)},
		Spec:       podSpec,
	}
	selector := &metav1.LabelSelector{MatchLabels: copyStringMap(podLabels)}
	replicas := int32(1)
	if req.Replicas != nil {
		if *req.Replicas < 0 || *req.Replicas > 1000 {
			return errors.New("副本数应在 0 到 1000 之间")
		}
		replicas = *req.Replicas
	}

	cs, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}

	var createErr error
	autoHeadless := false
	serviceName := ""
	switch kind {
	case WorkloadDeployment:
		strategy, sErr := buildDeploymentStrategy(req)
		if sErr != nil {
			return sErr
		}
		obj := &appsv1.Deployment{
			ObjectMeta: meta,
			Spec: appsv1.DeploymentSpec{
				Replicas:             &replicas,
				Selector:             selector,
				Template:             template,
				Strategy:             strategy,
				MinReadySeconds:      req.MinReadySeconds,
				RevisionHistoryLimit: req.RevisionHistoryLimit,
			},
		}
		_, createErr = cs.AppsV1().Deployments(namespace).Create(ctx, obj, metav1.CreateOptions{})
	case WorkloadStatefulSet:
		strategy, sErr := buildStatefulSetStrategy(req)
		if sErr != nil {
			return sErr
		}
		policy := appsv1.PodManagementPolicyType(strings.TrimSpace(req.PodManagementPolicy))
		switch policy {
		case "":
			policy = appsv1.OrderedReadyPodManagement
		case appsv1.OrderedReadyPodManagement, appsv1.ParallelPodManagement:
		default:
			return fmt.Errorf("不支持的容器组管理策略: %s", req.PodManagementPolicy)
		}
		serviceName = strings.TrimSpace(req.ServiceName)
		if serviceName == "" {
			serviceName = name
			autoHeadless = true
		}
		if errs := validation.IsDNS1035Label(serviceName); len(errs) > 0 {
			return fmt.Errorf("服务名称 %s 不合法: 需要以小写字母开头，只能包含小写字母、数字和 \"-\"", serviceName)
		}
		obj := &appsv1.StatefulSet{
			ObjectMeta: meta,
			Spec: appsv1.StatefulSetSpec{
				Replicas:             &replicas,
				Selector:             selector,
				Template:             template,
				ServiceName:          serviceName,
				UpdateStrategy:       strategy,
				PodManagementPolicy:  policy,
				VolumeClaimTemplates: claims,
				MinReadySeconds:      req.MinReadySeconds,
				RevisionHistoryLimit: req.RevisionHistoryLimit,
			},
		}
		_, createErr = cs.AppsV1().StatefulSets(namespace).Create(ctx, obj, metav1.CreateOptions{})
	default:
		strategy, sErr := buildDaemonSetStrategy(req)
		if sErr != nil {
			return sErr
		}
		obj := &appsv1.DaemonSet{
			ObjectMeta: meta,
			Spec: appsv1.DaemonSetSpec{
				Selector:             selector,
				Template:             template,
				UpdateStrategy:       strategy,
				MinReadySeconds:      req.MinReadySeconds,
				RevisionHistoryLimit: req.RevisionHistoryLimit,
			},
		}
		_, createErr = cs.AppsV1().DaemonSets(namespace).Create(ctx, obj, metav1.CreateOptions{})
	}
	if createErr != nil {
		if apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("%s %s/%s 已存在", kindName, namespace, name)
		}
		logs.Error("创建%s失败: %v", kindName, createErr)
		return fmt.Errorf("创建%s失败: %v", kindName, createErr)
	}
	logs.Info("%s %s/%s 创建成功", kindName, namespace, name)

	// StatefulSet 需要 Headless 服务为每个容器组提供固定的域名, 与 KubeSphere 一样自动创建
	if autoHeadless {
		if err = ensureHeadlessService(ctx, cs, namespace, serviceName, podLabels, containers, creator); err != nil {
			return fmt.Errorf("%s 已创建，但创建 Headless 服务失败: %v", kindName, err)
		}
	}
	if req.Service != nil && kind != WorkloadStatefulSet && strings.TrimSpace(req.Service.Type) != "" {
		if err = createWorkloadService(ctx, cs, namespace, name, req.Service, podLabels, containers, creator); err != nil {
			return fmt.Errorf("%s 已创建，但创建服务失败: %v", kindName, err)
		}
	}
	return nil
}

func validateLabelMap(m map[string]string, what string) error {
	for k, v := range m {
		if errs := validation.IsQualifiedName(k); len(errs) > 0 {
			return fmt.Errorf("%s的键 %s 不合法: %s", what, k, strings.Join(errs, "; "))
		}
		if errs := validation.IsValidLabelValue(v); len(errs) > 0 {
			return fmt.Errorf("%s %s 的值 %s 不合法: 最多 63 个字符，只能包含字母、数字、\"-\"、\"_\" 和 \".\"", what, k, v)
		}
	}
	return nil
}

func validateAnnotationKeys(m map[string]string) error {
	for k := range m {
		if errs := validation.IsQualifiedName(strings.ToLower(k)); len(errs) > 0 {
			return fmt.Errorf("注解的键 %s 不合法: %s", k, strings.Join(errs, "; "))
		}
	}
	return nil
}

func mountHostTime(containers []corev1.Container) {
	for i := range containers {
		exists := false
		for _, m := range containers[i].VolumeMounts {
			if m.MountPath == "/etc/localtime" {
				exists = true
				break
			}
		}
		if !exists {
			containers[i].VolumeMounts = append(containers[i].VolumeMounts, corev1.VolumeMount{Name: hostTimeVolume, MountPath: "/etc/localtime", ReadOnly: true})
		}
	}
}

func buildWorkloadVolumes(items []dto.WorkloadVolumeDTO) ([]corev1.Volume, map[string]bool, error) {
	result := make([]corev1.Volume, 0, len(items))
	names := map[string]bool{}
	for i, item := range items {
		label := fmt.Sprintf("第 %d 个存储卷", i+1)
		name := strings.TrimSpace(item.Name)
		if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
			return nil, nil, fmt.Errorf("%s的名称不合法: 只能包含小写字母、数字和 \"-\"", label)
		}
		if names[name] {
			return nil, nil, fmt.Errorf("存储卷名称 %s 重复", name)
		}
		names[name] = true
		source := strings.TrimSpace(item.Source)
		v := corev1.Volume{Name: name}
		switch strings.TrimSpace(item.Type) {
		case "pvc":
			if source == "" {
				return nil, nil, fmt.Errorf("请选择存储卷 %s 使用的持久卷声明", name)
			}
			v.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{ClaimName: source, ReadOnly: item.ReadOnly}
		case "configMap":
			if source == "" {
				return nil, nil, fmt.Errorf("请选择存储卷 %s 使用的 ConfigMap", name)
			}
			v.ConfigMap = &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: source}}
		case "secret":
			if source == "" {
				return nil, nil, fmt.Errorf("请选择存储卷 %s 使用的 Secret", name)
			}
			v.Secret = &corev1.SecretVolumeSource{SecretName: source}
		case "emptyDir":
			ed := &corev1.EmptyDirVolumeSource{}
			if item.Medium == string(corev1.StorageMediumMemory) {
				ed.Medium = corev1.StorageMediumMemory
			}
			if limit := strings.TrimSpace(item.SizeLimit); limit != "" {
				qty, qErr := resource.ParseQuantity(limit)
				if qErr != nil || qty.Sign() <= 0 {
					return nil, nil, fmt.Errorf("存储卷 %s 的容量上限格式不正确: %s", name, limit)
				}
				ed.SizeLimit = &qty
			}
			v.EmptyDir = ed
		case "hostPath":
			if !strings.HasPrefix(source, "/") {
				return nil, nil, fmt.Errorf("存储卷 %s 的主机路径必须是绝对路径", name)
			}
			hp := &corev1.HostPathVolumeSource{Path: source}
			if t := strings.TrimSpace(item.HostPathType); t != "" {
				ht := corev1.HostPathType(t)
				hp.Type = &ht
			}
			v.HostPath = hp
		default:
			return nil, nil, fmt.Errorf("存储卷 %s 的类型不支持: %s", name, item.Type)
		}
		result = append(result, v)
	}
	return result, names, nil
}

func buildClaimTemplates(items []dto.VolumeClaimTemplateDTO, volumeNames map[string]bool) ([]corev1.PersistentVolumeClaim, error) {
	result := make([]corev1.PersistentVolumeClaim, 0, len(items))
	seen := map[string]bool{}
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
			return nil, fmt.Errorf("第 %d 个存储卷模板的名称不合法: 只能包含小写字母、数字和 \"-\"", i+1)
		}
		if seen[name] || volumeNames[name] {
			return nil, fmt.Errorf("存储卷名称 %s 重复", name)
		}
		seen[name] = true
		qty, err := parseStorageQuantity(item.Capacity)
		if err != nil {
			return nil, fmt.Errorf("存储卷模板 %s: %v", name, err)
		}
		modes, err := parseAccessModes(item.AccessModes)
		if err != nil {
			return nil, fmt.Errorf("存储卷模板 %s: %v", name, err)
		}
		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: modes,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: qty},
				},
			},
		}
		if sc := strings.TrimSpace(item.StorageClassName); sc != "" {
			pvc.Spec.StorageClassName = &sc
		}
		result = append(result, pvc)
	}
	return result, nil
}

func buildWorkloadContainers(items []dto.WorkloadContainerDTO, mountable map[string]bool) ([]corev1.Container, []corev1.Container, error) {
	if len(items) == 0 {
		return nil, nil, errors.New("请至少添加一个容器")
	}
	base := make([]dto.JobContainerDTO, len(items))
	for i := range items {
		base[i] = items[i].JobContainerDTO
	}
	built, err := buildJobContainers(base)
	if err != nil {
		return nil, nil, err
	}
	containers := make([]corev1.Container, 0, len(built))
	var initContainers []corev1.Container
	for i := range built {
		item := items[i]
		c := built[i]
		label := "容器 " + c.Name
		if c.Ports, err = buildContainerPorts(item.Ports, label); err != nil {
			return nil, nil, err
		}
		if item.Init && (hasProbe(item.LivenessProbe) || hasProbe(item.ReadinessProbe) || hasProbe(item.StartupProbe)) {
			return nil, nil, fmt.Errorf("%s是初始化容器，不支持健康检查", label)
		}
		if c.LivenessProbe, err = buildProbe(item.LivenessProbe, label+"的存活检查", false); err != nil {
			return nil, nil, err
		}
		if c.ReadinessProbe, err = buildProbe(item.ReadinessProbe, label+"的就绪检查", true); err != nil {
			return nil, nil, err
		}
		if c.StartupProbe, err = buildProbe(item.StartupProbe, label+"的启动检查", false); err != nil {
			return nil, nil, err
		}
		paths := map[string]bool{}
		for _, m := range item.VolumeMounts {
			volume, path := strings.TrimSpace(m.Volume), strings.TrimSpace(m.MountPath)
			if volume == "" && path == "" {
				continue
			}
			if !mountable[volume] {
				return nil, nil, fmt.Errorf("%s挂载的存储卷 %s 不存在", label, volume)
			}
			if !strings.HasPrefix(path, "/") {
				return nil, nil, fmt.Errorf("%s的挂载路径必须是绝对路径: %s", label, path)
			}
			if paths[path] {
				return nil, nil, fmt.Errorf("%s的挂载路径 %s 重复", label, path)
			}
			paths[path] = true
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{Name: volume, MountPath: path, SubPath: strings.TrimSpace(m.SubPath), ReadOnly: m.ReadOnly})
		}
		if item.Init {
			initContainers = append(initContainers, c)
		} else {
			containers = append(containers, c)
		}
	}
	if len(containers) == 0 {
		return nil, nil, errors.New("除初始化容器外，至少需要一个工作容器")
	}
	return containers, initContainers, nil
}

func parseProtocol(value string) (corev1.Protocol, error) {
	switch p := corev1.Protocol(strings.ToUpper(strings.TrimSpace(value))); p {
	case "":
		return corev1.ProtocolTCP, nil
	case corev1.ProtocolTCP, corev1.ProtocolUDP, corev1.ProtocolSCTP:
		return p, nil
	default:
		return "", fmt.Errorf("不支持的协议: %s", value)
	}
}

func buildContainerPorts(items []dto.ContainerPortDTO, label string) ([]corev1.ContainerPort, error) {
	var result []corev1.ContainerPort
	seen := map[string]bool{}
	names := map[string]bool{}
	for _, p := range items {
		name := strings.TrimSpace(p.Name)
		if p.ContainerPort == 0 && name == "" {
			continue
		}
		if p.ContainerPort < 1 || p.ContainerPort > 65535 {
			return nil, fmt.Errorf("%s的端口 %d 不正确，应在 1 到 65535 之间", label, p.ContainerPort)
		}
		protocol, err := parseProtocol(p.Protocol)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", label, err)
		}
		key := fmt.Sprintf("%d/%s", p.ContainerPort, protocol)
		if seen[key] {
			return nil, fmt.Errorf("%s的端口 %s 重复", label, key)
		}
		seen[key] = true
		cp := corev1.ContainerPort{ContainerPort: p.ContainerPort, Protocol: protocol}
		if name != "" {
			if errs := validation.IsValidPortName(name); len(errs) > 0 {
				return nil, fmt.Errorf("%s的端口名称 %s 不合法: 最多 15 个字符，只能包含小写字母、数字和 \"-\"，且至少包含一个字母", label, name)
			}
			if names[name] {
				return nil, fmt.Errorf("%s的端口名称 %s 重复", label, name)
			}
			names[name] = true
			cp.Name = name
		}
		result = append(result, cp)
	}
	return result, nil
}

func hasProbe(p *dto.ProbeDTO) bool {
	return p != nil && strings.TrimSpace(p.Type) != ""
}

// buildProbe 健康检查; 只有就绪检查可以设置成功阈值
func buildProbe(p *dto.ProbeDTO, label string, allowSuccessThreshold bool) (*corev1.Probe, error) {
	if !hasProbe(p) {
		return nil, nil
	}
	probe := &corev1.Probe{}
	checkPort := func() error {
		if p.Port < 1 || p.Port > 65535 {
			return fmt.Errorf("%s的端口不正确", label)
		}
		return nil
	}
	switch strings.TrimSpace(p.Type) {
	case "http":
		if err := checkPort(); err != nil {
			return nil, err
		}
		path := strings.TrimSpace(p.Path)
		if path == "" {
			path = "/"
		}
		if !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("%s的路径必须以 / 开头", label)
		}
		scheme := corev1.URISchemeHTTP
		if strings.EqualFold(strings.TrimSpace(p.Scheme), "HTTPS") {
			scheme = corev1.URISchemeHTTPS
		}
		probe.HTTPGet = &corev1.HTTPGetAction{Path: path, Port: intstr.FromInt32(p.Port), Scheme: scheme}
	case "tcp":
		if err := checkPort(); err != nil {
			return nil, err
		}
		probe.TCPSocket = &corev1.TCPSocketAction{Port: intstr.FromInt32(p.Port)}
	case "exec":
		command := cleanStrings(p.Command)
		if len(command) == 0 {
			return nil, fmt.Errorf("请填写%s执行的命令", label)
		}
		probe.Exec = &corev1.ExecAction{Command: command}
	default:
		return nil, fmt.Errorf("%s的方式不支持: %s", label, p.Type)
	}
	if p.InitialDelaySeconds < 0 || p.PeriodSeconds < 0 || p.TimeoutSeconds < 0 || p.FailureThreshold < 0 || p.SuccessThreshold < 0 {
		return nil, fmt.Errorf("%s的时间和次数不能小于 0", label)
	}
	probe.InitialDelaySeconds = p.InitialDelaySeconds
	probe.PeriodSeconds = p.PeriodSeconds
	probe.TimeoutSeconds = p.TimeoutSeconds
	probe.FailureThreshold = p.FailureThreshold
	if allowSuccessThreshold {
		probe.SuccessThreshold = p.SuccessThreshold
	}
	return probe, nil
}

// parseIntOrPercent 解析 1 或 25% 这样的值, 空字符串返回 nil(使用默认值)
func parseIntOrPercent(value, what string) (*intstr.IntOrString, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if strings.HasSuffix(value, "%") {
		n, err := strconv.Atoi(strings.TrimSuffix(value, "%"))
		if err != nil || n < 0 || n > 100 {
			return nil, fmt.Errorf("%s格式不正确，例如 25%% 或 1", what)
		}
		v := intstr.FromString(value)
		return &v, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("%s格式不正确，例如 25%% 或 1", what)
	}
	v := intstr.FromInt32(int32(n))
	return &v, nil
}

func buildDeploymentStrategy(req dto.WorkloadCreateDTO) (appsv1.DeploymentStrategy, error) {
	switch appsv1.DeploymentStrategyType(strings.TrimSpace(req.Strategy)) {
	case "", appsv1.RollingUpdateDeploymentStrategyType:
		surge, err := parseIntOrPercent(req.MaxSurge, "最大超出副本数")
		if err != nil {
			return appsv1.DeploymentStrategy{}, err
		}
		unavailable, err := parseIntOrPercent(req.MaxUnavailable, "最大不可用副本数")
		if err != nil {
			return appsv1.DeploymentStrategy{}, err
		}
		if surge != nil && unavailable != nil && surge.IntValue() == 0 && unavailable.IntValue() == 0 {
			return appsv1.DeploymentStrategy{}, errors.New("最大超出副本数和最大不可用副本数不能同时为 0")
		}
		return appsv1.DeploymentStrategy{
			Type:          appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{MaxSurge: surge, MaxUnavailable: unavailable},
		}, nil
	case appsv1.RecreateDeploymentStrategyType:
		return appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}, nil
	default:
		return appsv1.DeploymentStrategy{}, fmt.Errorf("不支持的更新策略: %s", req.Strategy)
	}
}

func buildStatefulSetStrategy(req dto.WorkloadCreateDTO) (appsv1.StatefulSetUpdateStrategy, error) {
	switch appsv1.StatefulSetUpdateStrategyType(strings.TrimSpace(req.Strategy)) {
	case "", appsv1.RollingUpdateStatefulSetStrategyType:
		strategy := appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType}
		if req.Partition != nil {
			if *req.Partition < 0 {
				return strategy, errors.New("分批更新的序号不能小于 0")
			}
			if *req.Partition > 0 {
				p := *req.Partition
				strategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{Partition: &p}
			}
		}
		return strategy, nil
	case appsv1.OnDeleteStatefulSetStrategyType:
		return appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}, nil
	default:
		return appsv1.StatefulSetUpdateStrategy{}, fmt.Errorf("不支持的更新策略: %s", req.Strategy)
	}
}

func buildDaemonSetStrategy(req dto.WorkloadCreateDTO) (appsv1.DaemonSetUpdateStrategy, error) {
	switch appsv1.DaemonSetUpdateStrategyType(strings.TrimSpace(req.Strategy)) {
	case "", appsv1.RollingUpdateDaemonSetStrategyType:
		unavailable, err := parseIntOrPercent(req.MaxUnavailable, "最大不可用容器组数")
		if err != nil {
			return appsv1.DaemonSetUpdateStrategy{}, err
		}
		strategy := appsv1.DaemonSetUpdateStrategy{Type: appsv1.RollingUpdateDaemonSetStrategyType}
		if unavailable != nil {
			if unavailable.IntValue() == 0 {
				return strategy, errors.New("最大不可用容器组数不能为 0")
			}
			strategy.RollingUpdate = &appsv1.RollingUpdateDaemonSet{MaxUnavailable: unavailable}
		}
		return strategy, nil
	case appsv1.OnDeleteDaemonSetStrategyType:
		return appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType}, nil
	default:
		return appsv1.DaemonSetUpdateStrategy{}, fmt.Errorf("不支持的更新策略: %s", req.Strategy)
	}
}

// servicePortsFromContainers 由容器端口生成服务端口
func servicePortsFromContainers(containers []corev1.Container) []corev1.ServicePort {
	result := make([]corev1.ServicePort, 0)
	seen := map[string]bool{}
	names := map[string]bool{}
	for _, c := range containers {
		for _, p := range c.Ports {
			protocol := p.Protocol
			if protocol == "" {
				protocol = corev1.ProtocolTCP
			}
			key := fmt.Sprintf("%d/%s", p.ContainerPort, protocol)
			if seen[key] {
				continue
			}
			seen[key] = true
			name := p.Name
			if name == "" || names[name] {
				name = fmt.Sprintf("%s-%d", strings.ToLower(string(protocol)), p.ContainerPort)
			}
			names[name] = true
			result = append(result, corev1.ServicePort{Name: name, Protocol: protocol, Port: p.ContainerPort, TargetPort: intstr.FromInt32(p.ContainerPort)})
		}
	}
	return result
}

func createWorkloadService(ctx context.Context, cs *kubernetes.Clientset, namespace, workloadName string, opt *dto.WorkloadServiceDTO, selector map[string]string, containers []corev1.Container, creator string) error {
	name := strings.TrimSpace(opt.Name)
	if name == "" {
		name = workloadName
	}
	if errs := validation.IsDNS1035Label(name); len(errs) > 0 {
		return fmt.Errorf("服务名称 %s 不合法: 需要以小写字母开头，只能包含小写字母、数字和 \"-\"", name)
	}
	ports := servicePortsFromContainers(containers)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      copyStringMap(selector),
			Annotations: withCreator(nil, creator),
		},
		Spec: corev1.ServiceSpec{Selector: copyStringMap(selector), Ports: ports},
	}
	switch strings.TrimSpace(opt.Type) {
	case "Headless":
		svc.Spec.ClusterIP = corev1.ClusterIPNone
	case string(corev1.ServiceTypeClusterIP), string(corev1.ServiceTypeNodePort):
		if len(ports) == 0 {
			return errors.New("容器没有设置端口，无法创建服务")
		}
		svc.Spec.Type = corev1.ServiceType(strings.TrimSpace(opt.Type))
	default:
		return fmt.Errorf("不支持的服务类型: %s", opt.Type)
	}
	if _, err := cs.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("服务 %s 已存在", name)
		}
		return err
	}
	return nil
}

// ensureHeadlessService StatefulSet 使用的 Headless 服务, 已存在时不处理
func ensureHeadlessService(ctx context.Context, cs *kubernetes.Clientset, namespace, name string, selector map[string]string, containers []corev1.Container, creator string) error {
	_, err := cs.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	return createWorkloadService(ctx, cs, namespace, name, &dto.WorkloadServiceDTO{Type: "Headless", Name: name}, selector, containers, creator)
}

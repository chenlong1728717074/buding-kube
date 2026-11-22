package vo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// WorkloadDetailVO 工作负载(Deployment / StatefulSet / DaemonSet)详情
type WorkloadDetailVO struct {
	BaseVO
	Kind   string `json:"kind"`
	Status string `json:"status"`
	// Replicas 期望副本数; DaemonSet 为需要运行的节点数
	Replicas          int32 `json:"replicas"`
	ReadyReplicas     int32 `json:"readyReplicas"`
	UpdatedReplicas   int32 `json:"updatedReplicas"`
	AvailableReplicas int32 `json:"availableReplicas"`
	// Strategy 更新策略
	Strategy             string                  `json:"strategy"`
	MaxSurge             string                  `json:"maxSurge"`
	MaxUnavailable       string                  `json:"maxUnavailable"`
	Partition            *int32                  `json:"partition"`
	MinReadySeconds      int32                   `json:"minReadySeconds"`
	RevisionHistoryLimit *int32                  `json:"revisionHistoryLimit"`
	PodManagementPolicy  string                  `json:"podManagementPolicy"`
	ServiceName          string                  `json:"serviceName"`
	Paused               bool                    `json:"paused"`
	CurrentRevision      int64                   `json:"currentRevision"`
	Selector             map[string]string       `json:"selector"`
	PodLabels            map[string]string       `json:"podLabels"`
	NodeSelector         map[string]string       `json:"nodeSelector"`
	ServiceAccountName   string                  `json:"serviceAccountName"`
	ImagePullSecrets     []string                `json:"imagePullSecrets"`
	HostNetwork          bool                    `json:"hostNetwork"`
	Creator              string                  `json:"creator"`
	Uid                  string                  `json:"uid"`
	Labels               map[string]string       `json:"labels"`
	Annotations          map[string]string       `json:"annotations"`
	CreateTime           time.Time               `json:"createTime"`
	UpdateTime           *time.Time              `json:"updateTime"`
	Conditions           []WorkloadConditionVO   `json:"conditions"`
	Containers           []ContainerTemplateVO   `json:"containers"`
	Volumes              []WorkloadVolumeVO      `json:"volumes"`
	VolumeClaimTemplates []VolumeClaimTemplateVO `json:"volumeClaimTemplates"`
	Pods                 []JobPodVO              `json:"pods"`
	Revisions            []WorkloadRevisionVO    `json:"revisions"`
	Services             []WorkloadServiceRefVO  `json:"services"`
	Events               []EventVO               `json:"events"`
}

type ContainerPortVO struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type VolumeMountVO struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath"`
	ReadOnly  bool   `json:"readOnly"`
}

// WorkloadVolumeVO 存储卷; Type: pvc / configMap / secret / emptyDir / hostPath / 其他
type WorkloadVolumeVO struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

type VolumeClaimTemplateVO struct {
	Name             string   `json:"name"`
	StorageClassName string   `json:"storageClassName"`
	AccessModes      []string `json:"accessModes"`
	Capacity         string   `json:"capacity"`
}

// WorkloadRevisionVO 版本记录
type WorkloadRevisionVO struct {
	Revision      int64     `json:"revision"`
	Name          string    `json:"name"`
	Images        []string  `json:"images"`
	Replicas      int32     `json:"replicas"`
	ReadyReplicas int32     `json:"readyReplicas"`
	Current       bool      `json:"current"`
	ChangeCause   string    `json:"changeCause"`
	CreateTime    time.Time `json:"createTime"`
}

// WorkloadServiceRefVO 选中该工作负载容器组的服务
type WorkloadServiceRefVO struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP"`
	Ports     string `json:"ports"`
}

type WorkloadConditionVO struct {
	Type     string     `json:"type"`
	Status   string     `json:"status"`
	Reason   string     `json:"reason"`
	Message  string     `json:"message"`
	LastTime *time.Time `json:"lastTime"`
}

// WorkloadSelectorVO 创建服务时选择工作负载用: 容器组标签和容器端口
type WorkloadSelectorVO struct {
	Kind   string            `json:"kind"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Ports  []ContainerPortVO `json:"ports"`
}

func newWorkloadDetail(kind string, meta *metav1.ObjectMeta, tpl *corev1.PodTemplateSpec, selector *metav1.LabelSelector) WorkloadDetailVO {
	result := WorkloadDetailVO{
		BaseVO: BaseVO{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Alias:     meta.Annotations["alias"],
			Describe:  meta.Annotations["describe"],
		},
		Kind:                 kind,
		PodLabels:            tpl.Labels,
		NodeSelector:         tpl.Spec.NodeSelector,
		ServiceAccountName:   tpl.Spec.ServiceAccountName,
		ImagePullSecrets:     []string{},
		HostNetwork:          tpl.Spec.HostNetwork,
		Creator:              ResourceCreator(meta.Annotations),
		Uid:                  string(meta.UID),
		Labels:               meta.Labels,
		Annotations:          meta.Annotations,
		CreateTime:           meta.CreationTimestamp.Time,
		Conditions:           []WorkloadConditionVO{},
		Containers:           PodSpec2Containers(&tpl.Spec),
		Volumes:              PodVolumes2VO(tpl.Spec.Volumes),
		VolumeClaimTemplates: []VolumeClaimTemplateVO{},
		Pods:                 []JobPodVO{},
		Revisions:            []WorkloadRevisionVO{},
		Services:             []WorkloadServiceRefVO{},
		Events:               []EventVO{},
	}
	if selector != nil {
		result.Selector = selector.MatchLabels
	}
	for _, s := range tpl.Spec.ImagePullSecrets {
		result.ImagePullSecrets = append(result.ImagePullSecrets, s.Name)
	}
	return result
}

func Deployment2DetailVO(d *appsv1.Deployment) WorkloadDetailVO {
	r := newWorkloadDetail("deployment", &d.ObjectMeta, &d.Spec.Template, d.Spec.Selector)
	r.Status = GetDeploymentStatusText(*d)
	r.Replicas = int32Value(d.Spec.Replicas, 1)
	r.ReadyReplicas = d.Status.ReadyReplicas
	r.UpdatedReplicas = d.Status.UpdatedReplicas
	r.AvailableReplicas = d.Status.AvailableReplicas
	r.Strategy = string(d.Spec.Strategy.Type)
	if r.Strategy == "" {
		r.Strategy = string(appsv1.RollingUpdateDeploymentStrategyType)
	}
	if ru := d.Spec.Strategy.RollingUpdate; ru != nil {
		r.MaxSurge = intOrStringText(ru.MaxSurge)
		r.MaxUnavailable = intOrStringText(ru.MaxUnavailable)
	}
	r.MinReadySeconds = d.Spec.MinReadySeconds
	r.RevisionHistoryLimit = d.Spec.RevisionHistoryLimit
	r.Paused = d.Spec.Paused
	r.UpdateTime = GetDeploymentUpdateTime(*d)
	for _, c := range d.Status.Conditions {
		r.Conditions = append(r.Conditions, newCondition(string(c.Type), string(c.Status), c.Reason, c.Message, c.LastUpdateTime))
	}
	return r
}

func StatefulSet2DetailVO(s *appsv1.StatefulSet) WorkloadDetailVO {
	r := newWorkloadDetail("statefulset", &s.ObjectMeta, &s.Spec.Template, s.Spec.Selector)
	r.Status = GetStatefulSetStatusText(*s)
	r.Replicas = int32Value(s.Spec.Replicas, 1)
	r.ReadyReplicas = s.Status.ReadyReplicas
	r.UpdatedReplicas = s.Status.UpdatedReplicas
	r.AvailableReplicas = s.Status.AvailableReplicas
	r.Strategy = string(s.Spec.UpdateStrategy.Type)
	if r.Strategy == "" {
		r.Strategy = string(appsv1.RollingUpdateStatefulSetStrategyType)
	}
	if ru := s.Spec.UpdateStrategy.RollingUpdate; ru != nil {
		r.Partition = ru.Partition
		r.MaxUnavailable = intOrStringText(ru.MaxUnavailable)
	}
	r.MinReadySeconds = s.Spec.MinReadySeconds
	r.RevisionHistoryLimit = s.Spec.RevisionHistoryLimit
	r.PodManagementPolicy = string(s.Spec.PodManagementPolicy)
	if r.PodManagementPolicy == "" {
		r.PodManagementPolicy = string(appsv1.OrderedReadyPodManagement)
	}
	r.ServiceName = s.Spec.ServiceName
	r.VolumeClaimTemplates = ClaimTemplates2VO(s.Spec.VolumeClaimTemplates)
	for _, c := range s.Status.Conditions {
		r.Conditions = append(r.Conditions, newCondition(string(c.Type), string(c.Status), c.Reason, c.Message, c.LastTransitionTime))
	}
	return r
}

func DaemonSet2DetailVO(d *appsv1.DaemonSet) WorkloadDetailVO {
	r := newWorkloadDetail("daemonset", &d.ObjectMeta, &d.Spec.Template, d.Spec.Selector)
	r.Status = GetDaemonSetStatusText(*d)
	r.Replicas = d.Status.DesiredNumberScheduled
	r.ReadyReplicas = d.Status.NumberReady
	r.UpdatedReplicas = d.Status.UpdatedNumberScheduled
	r.AvailableReplicas = d.Status.NumberAvailable
	r.Strategy = string(d.Spec.UpdateStrategy.Type)
	if r.Strategy == "" {
		r.Strategy = string(appsv1.RollingUpdateDaemonSetStrategyType)
	}
	if ru := d.Spec.UpdateStrategy.RollingUpdate; ru != nil {
		r.MaxSurge = intOrStringText(ru.MaxSurge)
		r.MaxUnavailable = intOrStringText(ru.MaxUnavailable)
	}
	r.MinReadySeconds = d.Spec.MinReadySeconds
	r.RevisionHistoryLimit = d.Spec.RevisionHistoryLimit
	for _, c := range d.Status.Conditions {
		r.Conditions = append(r.Conditions, newCondition(string(c.Type), string(c.Status), c.Reason, c.Message, c.LastTransitionTime))
	}
	return r
}

func newCondition(t, status, reason, message string, at metav1.Time) WorkloadConditionVO {
	c := WorkloadConditionVO{Type: t, Status: status, Reason: reason, Message: message}
	if !at.IsZero() {
		v := at.Time
		c.LastTime = &v
	}
	return c
}

// NewWorkloadSelectorVO 工作负载的容器组标签和端口
func NewWorkloadSelectorVO(kind, name string, tpl *corev1.PodTemplateSpec) WorkloadSelectorVO {
	return WorkloadSelectorVO{
		Kind:   kind,
		Name:   name,
		Labels: tpl.Labels,
		Ports:  TemplatePorts(tpl),
	}
}

// TemplatePorts 容器组模板中所有工作容器的端口
func TemplatePorts(tpl *corev1.PodTemplateSpec) []ContainerPortVO {
	result := make([]ContainerPortVO, 0)
	for i := range tpl.Spec.Containers {
		result = append(result, containerPorts(&tpl.Spec.Containers[i])...)
	}
	return result
}

// TemplateImages 容器组模板中的镜像
func TemplateImages(tpl *corev1.PodTemplateSpec) []string {
	result := make([]string, 0, len(tpl.Spec.Containers))
	for _, c := range tpl.Spec.Containers {
		result = append(result, c.Image)
	}
	return result
}

func containerPorts(c *corev1.Container) []ContainerPortVO {
	result := make([]ContainerPortVO, 0, len(c.Ports))
	for _, p := range c.Ports {
		protocol := string(p.Protocol)
		if protocol == "" {
			protocol = string(corev1.ProtocolTCP)
		}
		result = append(result, ContainerPortVO{Name: p.Name, ContainerPort: p.ContainerPort, Protocol: protocol})
	}
	return result
}

func volumeMounts(c *corev1.Container) []VolumeMountVO {
	result := make([]VolumeMountVO, 0, len(c.VolumeMounts))
	for _, m := range c.VolumeMounts {
		result = append(result, VolumeMountVO{Name: m.Name, MountPath: m.MountPath, SubPath: m.SubPath, ReadOnly: m.ReadOnly})
	}
	return result
}

func int32Or(v, def int32) int32 {
	if v == 0 {
		return def
	}
	return v
}

// probeText 健康检查的说明文字
func probeText(p *corev1.Probe) string {
	if p == nil {
		return ""
	}
	var target string
	switch {
	case p.HTTPGet != nil:
		scheme := string(p.HTTPGet.Scheme)
		if scheme == "" {
			scheme = "HTTP"
		}
		target = fmt.Sprintf("%s GET :%s%s", scheme, p.HTTPGet.Port.String(), p.HTTPGet.Path)
	case p.TCPSocket != nil:
		target = "TCP :" + p.TCPSocket.Port.String()
	case p.Exec != nil:
		target = "执行命令 " + strings.Join(p.Exec.Command, " ")
	case p.GRPC != nil:
		target = fmt.Sprintf("gRPC :%d", p.GRPC.Port)
	default:
		target = "未知方式"
	}
	return fmt.Sprintf("%s（延迟 %d 秒，间隔 %d 秒，超时 %d 秒，失败 %d 次判定异常）",
		target, p.InitialDelaySeconds, int32Or(p.PeriodSeconds, 10), int32Or(p.TimeoutSeconds, 1), int32Or(p.FailureThreshold, 3))
}

// PodVolumes2VO 存储卷列表
func PodVolumes2VO(volumes []corev1.Volume) []WorkloadVolumeVO {
	result := make([]WorkloadVolumeVO, 0, len(volumes))
	for _, v := range volumes {
		item := WorkloadVolumeVO{Name: v.Name}
		switch {
		case v.PersistentVolumeClaim != nil:
			item.Type, item.Source = "pvc", v.PersistentVolumeClaim.ClaimName
		case v.ConfigMap != nil:
			item.Type, item.Source = "configMap", v.ConfigMap.Name
		case v.Secret != nil:
			item.Type, item.Source = "secret", v.Secret.SecretName
		case v.EmptyDir != nil:
			item.Type = "emptyDir"
			if v.EmptyDir.Medium == corev1.StorageMediumMemory {
				item.Source = "内存"
			}
			if v.EmptyDir.SizeLimit != nil {
				item.Source = strings.TrimSpace(item.Source + " 上限 " + v.EmptyDir.SizeLimit.String())
			}
		case v.HostPath != nil:
			item.Type, item.Source = "hostPath", v.HostPath.Path
		case v.Projected != nil:
			item.Type = "projected"
			names := make([]string, 0)
			for _, s := range v.Projected.Sources {
				switch {
				case s.ConfigMap != nil:
					names = append(names, "ConfigMap "+s.ConfigMap.Name)
				case s.Secret != nil:
					names = append(names, "Secret "+s.Secret.Name)
				case s.ServiceAccountToken != nil:
					names = append(names, "ServiceAccount Token")
				case s.DownwardAPI != nil:
					names = append(names, "DownwardAPI")
				}
			}
			item.Source = strings.Join(names, "，")
		case v.DownwardAPI != nil:
			item.Type = "downwardAPI"
		case v.NFS != nil:
			item.Type, item.Source = "nfs", v.NFS.Server+":"+v.NFS.Path
		default:
			item.Type = "other"
		}
		result = append(result, item)
	}
	return result
}

// ClaimTemplates2VO StatefulSet 的存储卷模板
func ClaimTemplates2VO(pvcs []corev1.PersistentVolumeClaim) []VolumeClaimTemplateVO {
	result := make([]VolumeClaimTemplateVO, 0, len(pvcs))
	for _, pvc := range pvcs {
		item := VolumeClaimTemplateVO{Name: pvc.Name, AccessModes: []string{}}
		if pvc.Spec.StorageClassName != nil {
			item.StorageClassName = *pvc.Spec.StorageClassName
		}
		for _, m := range pvc.Spec.AccessModes {
			item.AccessModes = append(item.AccessModes, string(m))
		}
		if qty, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			item.Capacity = qty.String()
		}
		result = append(result, item)
	}
	return result
}

// Service2RefVO 服务的简要信息
func Service2RefVO(svc *corev1.Service) WorkloadServiceRefVO {
	t := string(svc.Spec.Type)
	if svc.Spec.ClusterIP == corev1.ClusterIPNone {
		t = "Headless"
	}
	return WorkloadServiceRefVO{
		Name:      svc.Name,
		Type:      t,
		ClusterIP: svc.Spec.ClusterIP,
		Ports:     ServicePortsText(svc.Spec.Ports),
	}
}

// ServicePortsText 例如 80/TCP→8080，443/TCP→8443（节点端口 30443）
func ServicePortsText(ports []corev1.ServicePort) string {
	items := make([]string, 0, len(ports))
	for _, p := range ports {
		protocol := string(p.Protocol)
		if protocol == "" {
			protocol = "TCP"
		}
		text := fmt.Sprintf("%d/%s", p.Port, protocol)
		if target := p.TargetPort.String(); target != "" && target != "0" && target != fmt.Sprint(p.Port) {
			text += "→" + target
		}
		if p.NodePort > 0 {
			text += fmt.Sprintf("（节点端口 %d）", p.NodePort)
		}
		items = append(items, text)
	}
	return strings.Join(items, "，")
}

func intOrStringText(v *intstr.IntOrString) string {
	if v == nil {
		return ""
	}
	return v.String()
}

// SortRevisions 版本号倒序
func SortRevisions(items []WorkloadRevisionVO) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Revision > items[j].Revision
	})
}

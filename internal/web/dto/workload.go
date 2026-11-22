package dto

type WorkloadBaseDTO struct {
	NamespaceBaseDTO
	Name string `json:"name" form:"name" `
}

type WorkloadQueryDTO struct {
	PageQueryDTO
	WorkloadBaseDTO
	Status string `json:"status" form:"status"`
}

// WorkloadTargetDTO 指定一个工作负载; Kind: deployment / statefulset / daemonset
type WorkloadTargetDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Kind      string `json:"kind" form:"kind" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
}

// WorkloadInfoDTO 编辑别名和描述
type WorkloadInfoDTO struct {
	WorkloadTargetDTO
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
}

// WorkloadScaleDTO 调整副本数
type WorkloadScaleDTO struct {
	WorkloadTargetDTO
	Replicas int32 `json:"replicas"`
}

// WorkloadRollbackDTO 回退到指定版本
type WorkloadRollbackDTO struct {
	WorkloadTargetDTO
	Revision int64 `json:"revision" binding:"required"`
}

// WorkloadImageDTO 更新镜像, 键为容器名
type WorkloadImageDTO struct {
	WorkloadTargetDTO
	Images map[string]string `json:"images" binding:"required"`
}

// ContainerPortDTO 容器端口
type ContainerPortDTO struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"containerPort"`
	// Protocol TCP(默认) / UDP / SCTP
	Protocol string `json:"protocol"`
}

// ProbeDTO 健康检查; Type 为空表示不检查
type ProbeDTO struct {
	// Type http / tcp / exec
	Type string `json:"type"`
	Path string `json:"path"`
	Port int32  `json:"port"`
	// Scheme HTTP(默认) / HTTPS
	Scheme              string   `json:"scheme"`
	Command             []string `json:"command"`
	InitialDelaySeconds int32    `json:"initialDelaySeconds"`
	PeriodSeconds       int32    `json:"periodSeconds"`
	TimeoutSeconds      int32    `json:"timeoutSeconds"`
	FailureThreshold    int32    `json:"failureThreshold"`
	SuccessThreshold    int32    `json:"successThreshold"`
}

// VolumeMountDTO 容器挂载存储卷
type VolumeMountDTO struct {
	Volume    string `json:"volume"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath"`
	ReadOnly  bool   `json:"readOnly"`
}

// WorkloadContainerDTO 工作负载的容器, 在任务容器的基础上增加端口、健康检查和存储挂载
type WorkloadContainerDTO struct {
	JobContainerDTO
	// Init 是否为初始化容器
	Init           bool               `json:"init"`
	Ports          []ContainerPortDTO `json:"ports"`
	LivenessProbe  *ProbeDTO          `json:"livenessProbe"`
	ReadinessProbe *ProbeDTO          `json:"readinessProbe"`
	StartupProbe   *ProbeDTO          `json:"startupProbe"`
	VolumeMounts   []VolumeMountDTO   `json:"volumeMounts"`
}

// WorkloadVolumeDTO 存储卷
type WorkloadVolumeDTO struct {
	Name string `json:"name"`
	// Type pvc / configMap / secret / emptyDir / hostPath
	Type string `json:"type"`
	// Source 持久卷声明名 / ConfigMap 名 / Secret 名 / 主机路径
	Source   string `json:"source"`
	ReadOnly bool   `json:"readOnly"`
	// Medium emptyDir 的介质, Memory 表示使用内存
	Medium       string `json:"medium"`
	SizeLimit    string `json:"sizeLimit"`
	HostPathType string `json:"hostPathType"`
}

// VolumeClaimTemplateDTO StatefulSet 的存储卷模板, 每个容器组自动创建一个持久卷声明
type VolumeClaimTemplateDTO struct {
	Name             string   `json:"name"`
	StorageClassName string   `json:"storageClassName"`
	AccessModes      []string `json:"accessModes"`
	Capacity         string   `json:"capacity"`
}

// WorkloadServiceDTO 创建工作负载时同时创建的服务
type WorkloadServiceDTO struct {
	// Type ClusterIP / NodePort / Headless
	Type string `json:"type"`
	// Name 为空时与工作负载同名
	Name string `json:"name"`
}

// WorkloadCreateDTO 创建工作负载
type WorkloadCreateDTO struct {
	ClusterId   string            `json:"clusterId" binding:"required"`
	Kind        string            `json:"kind" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Alias       string            `json:"alias"`
	Describe    string            `json:"describe"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	// Replicas 副本数, DaemonSet 忽略
	Replicas             *int32                   `json:"replicas"`
	Containers           []WorkloadContainerDTO   `json:"containers"`
	Volumes              []WorkloadVolumeDTO      `json:"volumes"`
	VolumeClaimTemplates []VolumeClaimTemplateDTO `json:"volumeClaimTemplates"`
	// Strategy 更新策略: Deployment 为 RollingUpdate / Recreate; StatefulSet、DaemonSet 为 RollingUpdate / OnDelete
	Strategy       string `json:"strategy"`
	MaxSurge       string `json:"maxSurge"`
	MaxUnavailable string `json:"maxUnavailable"`
	// Partition StatefulSet 分批更新: 只更新序号不小于该值的容器组
	Partition            *int32 `json:"partition"`
	MinReadySeconds      int32  `json:"minReadySeconds"`
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit"`
	// PodManagementPolicy StatefulSet 容器组管理策略: OrderedReady(默认) / Parallel
	PodManagementPolicy string `json:"podManagementPolicy"`
	// ServiceName StatefulSet 使用的 Headless 服务, 为空时自动创建同名服务
	ServiceName        string            `json:"serviceName"`
	NodeSelector       map[string]string `json:"nodeSelector"`
	ServiceAccountName string            `json:"serviceAccountName"`
	ImagePullSecrets   []string          `json:"imagePullSecrets"`
	// SyncHostTimezone 挂载主机的 /etc/localtime, 使容器与节点时区一致
	SyncHostTimezone              bool   `json:"syncHostTimezone"`
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds"`
	// Service 同时创建服务(Deployment / DaemonSet)
	Service *WorkloadServiceDTO `json:"service"`
}

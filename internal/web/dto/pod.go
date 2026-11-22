package dto

type PodDTO struct {
	NamespaceBaseDTO
	Name string `json:"name" form:"name" binding:"required"`
}

type PodQueryDTO struct {
	NamespaceBaseDTO
	PageQueryDTO
	Status string `json:"status" form:"status"`
	// SortBy 排序字段: createTime(默认) / cpu / memory / restarts / name
	SortBy string `json:"sortBy" form:"sortBy"`
	// SortOrder desc(默认) / asc
	SortOrder string `json:"sortOrder" form:"sortOrder"`
}

type PodLogDTO struct {
	PodDTO
	SinceTime     *CustomTime `json:"sinceTime"`
	ContainerName string      `json:"containerName"`
	Follow        bool        `json:"follow"`
	TailLines     int64       `json:"tailLines"`
}

type PodDownloadDTO struct {
	PodDTO
	FilePath      string `json:"filePath" form:"filePath"`
	ContainerName string `json:"containerName" form:"containerName"`
}

// PodCreateDTO 分步创建容器组, 容器和存储卷的格式与工作负载相同
type PodCreateDTO struct {
	ClusterId   string                 `json:"clusterId" binding:"required"`
	Namespace   string                 `json:"namespace" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Alias       string                 `json:"alias"`
	Describe    string                 `json:"describe"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Containers  []WorkloadContainerDTO `json:"containers"`
	Volumes     []WorkloadVolumeDTO    `json:"volumes"`
	// RestartPolicy Always(默认) / OnFailure / Never
	RestartPolicy string `json:"restartPolicy"`
	// NodeName 直接指定运行的节点, 不经过调度器
	NodeName           string            `json:"nodeName"`
	NodeSelector       map[string]string `json:"nodeSelector"`
	ServiceAccountName string            `json:"serviceAccountName"`
	ImagePullSecrets   []string          `json:"imagePullSecrets"`
	// SyncHostTimezone 挂载主机的 /etc/localtime
	SyncHostTimezone              bool   `json:"syncHostTimezone"`
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds"`
	// HostNetwork 使用节点网络
	HostNetwork bool `json:"hostNetwork"`
}

// PodInfoDTO 编辑别名和描述
type PodInfoDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
}

// PodImageDTO 更新镜像, 键为容器名
type PodImageDTO struct {
	ClusterId string            `json:"clusterId" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Images    map[string]string `json:"images" binding:"required"`
}

// KubectlTerminalDTO Web kubectl(WebSocket 握手的查询参数)
type KubectlTerminalDTO struct {
	ClusterId string `form:"clusterId" binding:"required"`
}

// PodTerminalDTO 进入容器终端(WebSocket 握手的查询参数)
type PodTerminalDTO struct {
	ClusterId string `form:"clusterId" binding:"required"`
	Namespace string `form:"namespace" binding:"required"`
	Name      string `form:"name" binding:"required"`
	// Container 为空时使用默认容器
	Container string `form:"container"`
	// Shell auto(默认, 优先 bash) / bash / sh
	Shell string `form:"shell"`
}

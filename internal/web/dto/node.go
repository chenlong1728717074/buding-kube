package dto

type NodeQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	PageQueryDTO
	// Status Ready / NotReady / Unknown
	Status string `json:"status" form:"status"`
	// Role 角色过滤: control-plane(包含 master) / worker
	Role string `json:"role" form:"role"`
}

type NodeUnScheduleDTO struct {
	NodeInfoQueryDTO
	UnSchedule *bool `json:"unSchedule" form:"unSchedule" binding:"required"`
}

type NodeInfoQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Hostname  string `json:"hostname" form:"hostname" binding:"required"`
}

type NodeUsageQueryDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
}

// NodeLabelsDTO 设置节点标签(整体替换); 由 kubelet 维护的标签会自动保留
type NodeLabelsDTO struct {
	NodeInfoQueryDTO
	Labels map[string]string `json:"labels"`
}

// NodeTaintDTO 污点
type NodeTaintDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	// Effect NoSchedule / PreferNoSchedule / NoExecute
	Effect string `json:"effect"`
}

// NodeTaintsDTO 设置节点污点(整体替换); 由节点控制器维护的污点会自动保留
type NodeTaintsDTO struct {
	NodeInfoQueryDTO
	Taints []NodeTaintDTO `json:"taints"`
}

// NodeDrainDTO 排空节点, 与 kubectl drain 的选项对应; 预览时作为查询参数
type NodeDrainDTO struct {
	NodeInfoQueryDTO
	// DeleteEmptyDirData 驱逐使用 emptyDir 的容器组(临时数据会丢失)
	DeleteEmptyDirData bool `json:"deleteEmptyDirData" form:"deleteEmptyDirData"`
	// Force 驱逐没有控制器管理的容器组(驱逐后不会重建)
	Force bool `json:"force" form:"force"`
	// GracePeriodSeconds 优雅终止时间, 小于 0 表示使用容器组自己的设置
	GracePeriodSeconds int64 `json:"gracePeriodSeconds" form:"gracePeriodSeconds"`
	// TimeoutSeconds 等待驱逐完成的最长时间(受 PodDisruptionBudget 限制时会重试), 默认 120
	TimeoutSeconds int `json:"timeoutSeconds" form:"timeoutSeconds"`
}

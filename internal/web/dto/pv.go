package dto

// PVNFSSource NFS 存储来源
type PVNFSSource struct {
	Server   string `json:"server"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly"`
}

// PVHostPathSource 节点目录(仅适合单节点测试)
type PVHostPathSource struct {
	Path string `json:"path"`
	// Type 为空时不检查; 可选 DirectoryOrCreate / Directory / FileOrCreate / File 等
	Type string `json:"type"`
}

// PVLocalSource 节点本地盘, 必须指定节点
type PVLocalSource struct {
	Path   string `json:"path"`
	Node   string `json:"node"`
	FSType string `json:"fsType"`
}

// PVCSISource CSI 驱动提供的已有卷
type PVCSISource struct {
	Driver           string            `json:"driver"`
	VolumeHandle     string            `json:"volumeHandle"`
	FSType           string            `json:"fsType"`
	ReadOnly         bool              `json:"readOnly"`
	VolumeAttributes map[string]string `json:"volumeAttributes"`
}

// PVCreateDTO 创建持久卷
type PVCreateDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	// Capacity 容量, 如 10Gi
	Capacity    string   `json:"capacity" binding:"required"`
	AccessModes []string `json:"accessModes" binding:"required"`
	// ReclaimPolicy 回收策略: Retain(默认) / Delete / Recycle(仅 NFS、HostPath)
	ReclaimPolicy    string            `json:"reclaimPolicy"`
	StorageClassName string            `json:"storageClassName"`
	VolumeMode       string            `json:"volumeMode"`
	MountOptions     []string          `json:"mountOptions"`
	Labels           map[string]string `json:"labels"`
	// SourceType 存储来源: nfs / hostPath / local / csi, 对应下面其中一个字段
	SourceType string            `json:"sourceType" binding:"required"`
	NFS        *PVNFSSource      `json:"nfs"`
	HostPath   *PVHostPathSource `json:"hostPath"`
	Local      *PVLocalSource    `json:"local"`
	CSI        *PVCSISource      `json:"csi"`
}

type PVBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type PVPageQueryBaseDTO struct {
	PageQueryDTO
	PVBaseDTO
}

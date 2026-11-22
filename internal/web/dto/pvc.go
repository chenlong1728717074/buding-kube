package dto

// PVCCreateDTO 创建持久卷声明
type PVCCreateDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	// Capacity 申请容量, 如 10Gi
	Capacity    string   `json:"capacity" binding:"required"`
	AccessModes []string `json:"accessModes" binding:"required"`
	// StorageClassName 为 nil 时使用集群默认存储类; 指向空字符串表示不使用存储类(绑定无存储类的持久卷)
	StorageClassName *string `json:"storageClassName"`
	// VolumeName 绑定已有持久卷时填写
	VolumeName  string            `json:"volumeName"`
	VolumeMode  string            `json:"volumeMode"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type PVCBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type PVCPageQueryBaseDTO struct {
	PageQueryDTO
	PVCBaseDTO
}

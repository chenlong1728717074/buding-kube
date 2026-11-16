package dto

type StorageClassCreateDTO struct {
	ClusterId            string            `json:"clusterId" form:"clusterId" binding:"required"`
	Name                 string            `json:"name" form:"name" binding:"required"`
	Provisioner          string            `json:"provisioner" form:"provisioner" binding:"required"`
	Parameters           map[string]string `json:"parameters"`
	ReclaimPolicy        *string           `json:"reclaimPolicy"`
	VolumeBindingMode    *string           `json:"volumeBindingMode"`
	AllowVolumeExpansion *bool             `json:"allowVolumeExpansion"`
	AllowedTopologies    []string          `json:"allowedTopologies"`
	Annotations          map[string]string `json:"annotations"`
	Labels               map[string]string `json:"labels"`
	Yaml                 string            `json:"yaml"`
}

type StorageClassBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type StorageClassPageQueryBaseDTO struct {
	PageQueryDTO
	StorageClassBaseDTO
	// StorageType 存储类型过滤: 块存储/对象存储/文件存储/未分类
	StorageType string `json:"storageType" form:"storageType"`
}

type StorageClassApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}

package dto

// StorageClassCreateDTO 创建存储类
type StorageClassCreateDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	// StorageType 存储类型: block / file / object; 为空时按供应者自动识别
	StorageType string            `json:"storageType"`
	Provisioner string            `json:"provisioner" binding:"required"`
	Parameters  map[string]string `json:"parameters"`
	// ReclaimPolicy 回收策略: Delete / Retain, 默认 Delete
	ReclaimPolicy string `json:"reclaimPolicy"`
	// VolumeBindingMode 卷绑定模式: Immediate / WaitForFirstConsumer, 默认 Immediate
	VolumeBindingMode    string   `json:"volumeBindingMode"`
	AllowVolumeExpansion bool     `json:"allowVolumeExpansion"`
	AllowVolumeClone     bool     `json:"allowVolumeClone"`
	AccessModes          []string `json:"accessModes"`
	MountOptions         []string `json:"mountOptions"`
	IsDefault            bool     `json:"isDefault"`
}

// StorageClassUpdateDTO 编辑存储类.
// 供应者、参数、回收策略、绑定模式创建后不可修改(Kubernetes 限制), 这里只包含可修改的字段
type StorageClassUpdateDTO struct {
	ClusterId            string `json:"clusterId" binding:"required"`
	Name                 string `json:"name" binding:"required"`
	Alias                string `json:"alias"`
	Describe             string `json:"describe"`
	StorageType          string `json:"storageType"`
	AllowVolumeExpansion *bool  `json:"allowVolumeExpansion"`
	AllowVolumeClone     *bool  `json:"allowVolumeClone"`
	IsDefault            *bool  `json:"isDefault"`
}

type StorageClassBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type StorageClassPageQueryBaseDTO struct {
	PageQueryDTO
	StorageClassBaseDTO
	// StorageType 存储类型过滤, 支持 block / file / object 或 块存储 / 文件存储 / 对象存储 / 未分类
	StorageType string `json:"storageType" form:"storageType"`
}

type StorageClassApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}

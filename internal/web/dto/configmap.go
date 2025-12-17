package dto

type ConfigMapCreateDTO struct {
	BaseInfoDTO
	Data        map[string]string `json:"data"`
	BinaryData  map[string][]byte `json:"binaryData"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
}

type ConfigMapBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type ConfigMapPageQueryBaseDTO struct {
	PageQueryDTO
	ConfigMapBaseDTO
}

type ConfigMapDataDTO struct {
	BaseInfoDTO
	Data map[string]string `json:"data"`
}

type ConfigMapSettingDTO struct {
	BaseInfoDTO
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

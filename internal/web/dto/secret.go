package dto

type SecretCreateDTO struct {
	ClusterId   string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace   string            `json:"namespace" form:"namespace" binding:"required"`
	Name        string            `json:"name" form:"name" binding:"required"`
	Type        string            `json:"type" form:"type"`
	Data        map[string][]byte `json:"data"`
	StringData  map[string]string `json:"stringData"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Yaml        string            `json:"yaml"`
}

type SecretBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type SecretPageQueryBaseDTO struct {
	PageQueryDTO
	SecretBaseDTO
}

type SecretApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}

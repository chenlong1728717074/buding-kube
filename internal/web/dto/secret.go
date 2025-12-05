package dto

type SecretCreateDTO struct {
	BaseInfoDTO
	Type        string            `json:"type" form:"type"`
	Data        map[string][]byte `json:"data"`
	StringData  map[string]string `json:"stringData"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

type SecretApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}

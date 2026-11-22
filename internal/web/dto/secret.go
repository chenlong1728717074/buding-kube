package dto

type SecretCreateDTO struct {
	BaseInfoDTO
	Type        string            `json:"type" form:"type"`
	Data        map[string][]byte `json:"data"`
	StringData  map[string]string `json:"stringData"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

package dto

type ServiceAccountCreateDTO struct {
	ClusterId                    string            `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace                    string            `json:"namespace" form:"namespace" binding:"required"`
	Name                         string            `json:"name" form:"name" binding:"required"`
	AutomountServiceAccountToken *bool             `json:"automountServiceAccountToken"`
	ImagePullSecrets             []string          `json:"imagePullSecrets"`
	Secrets                      []string          `json:"secrets"`
	Annotations                  map[string]string `json:"annotations"`
	Labels                       map[string]string `json:"labels"`
	Yaml                         string            `json:"yaml"`
}

type ServiceAccountBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace"`
	Name      string `json:"name" form:"name"`
	Force     bool   `json:"force" form:"force"`
}

type ServiceAccountPageQueryBaseDTO struct {
	PageQueryDTO
	ServiceAccountBaseDTO
}

type ServiceAccountApplyDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Yaml      string `json:"yaml" binding:"required"`
}

package dto

type NamespaceCreateDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
}

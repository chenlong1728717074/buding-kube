package dto

type NodeCreateDTO struct {
	Id       string `json:"id"`
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

package dto

// 集群创建参数
type ClusterCreateDTO struct {
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

// 集群修改参数
type ClusterUpdateDTO struct {
	Id       string `json:"id"`
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

// 集群详情返回值
type ClusterDetailDTO struct {
	Id       string `json:"id"`
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

// 节点创建参数（保持原有结构）
type NodeCreateDTO struct {
	Id       string `json:"id"`
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

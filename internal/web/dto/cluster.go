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
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

// 集群详情返回值
type ClusterDetailDTO struct {
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config" binding:"required"`
}

// 集群创建参数
type NodeCreateDTO struct {
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config"`
	Uri      string `json:"uri"`
	Token    string `json:"token"`
}

// 集群更新参数（配置可选，留空表示不更新 kubeconfig）
type NodeUpdateDTO struct {
	Name     string `json:"name"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config"`
	Uri      string `json:"uri"`
	Token    string `json:"token"`
}

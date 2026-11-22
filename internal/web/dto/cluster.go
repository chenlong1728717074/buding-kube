package dto

// 集群创建/更新参数（配置可选，留空表示沿用已有 kubeconfig；支持 config 或 uri+token 两种方式）
type ClusterSaveDTO struct {
	Name     string `json:"name" binding:"required"`
	Alias    string `json:"alias"`
	Describe string `json:"describe"`
	Config   string `json:"config"`
	Uri      string `json:"uri"`
	Token    string `json:"token"`
	// CaData uri+token 方式的 API Server CA 证书(PEM 或其 base64), 用于校验服务端证书
	CaData string `json:"caData"`
	// Insecure uri+token 方式跳过服务端证书校验(不安全, 仅用于测试环境)
	Insecure bool `json:"insecure"`
}

// 集群分页查询参数
type ClusterPageQueryDTO struct {
	PageQueryDTO
	Status string `form:"status" json:"status"` // Active/Error/Unknown，为空时不过滤
}

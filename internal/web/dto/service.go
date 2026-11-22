package dto

type ServiceQueryDTO struct {
	PageQueryDTO
	NamespaceBaseDTO
	// Type 类型过滤: ClusterIP / NodePort / LoadBalancer / ExternalName / Headless
	Type string `json:"type" form:"type"`
}

type ServiceBaseDTO struct {
	ClusterId string `json:"clusterId" form:"clusterId" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
}

// ServicePortDTO 服务端口
type ServicePortDTO struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     int32  `json:"port"`
	// TargetPort 容器端口号或容器端口名称, 为空时与 Port 相同
	TargetPort string `json:"targetPort"`
	// NodePort 节点端口, 0 表示自动分配
	NodePort int32 `json:"nodePort"`
}

// ServiceFormDTO 创建 / 编辑服务
type ServiceFormDTO struct {
	ClusterId string `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	Describe  string `json:"describe"`
	// Type ClusterIP / NodePort / LoadBalancer / ExternalName
	Type string `json:"type"`
	// Headless 内部访问模式为「内部域名」, 不分配虚拟 IP
	Headless bool              `json:"headless"`
	Selector map[string]string `json:"selector"`
	Ports    []ServicePortDTO  `json:"ports"`
	// SessionAffinity None / ClientIP
	SessionAffinity        string `json:"sessionAffinity"`
	SessionAffinityTimeout int32  `json:"sessionAffinityTimeout"`
	ExternalName           string `json:"externalName"`
	// ExternalTrafficPolicy Cluster / Local
	ExternalTrafficPolicy string            `json:"externalTrafficPolicy"`
	Labels                map[string]string `json:"labels"`
	Annotations           map[string]string `json:"annotations"`
}

package vo

type ClusterQueryVO struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	Describe       string `json:"describe"`
	Status         string `json:"status"`
	Version        string `json:"version"`
	ApiServer      string `json:"apiServer"`
	CreateTime     string `json:"createTime"`
	NodeCount      int64  `json:"nodeCount"`
	PodCount       int64  `json:"podCount"`
	NamespaceCount int64  `json:"namespaceCount"`
}

// ClusterVO 集群详情返回值(不含 kubeconfig,避免凭证外泄)
type ClusterVO struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Alias           string `json:"alias"`
	Describe        string `json:"describe"`
	Status          string `json:"status"`
	Version         string `json:"version"`
	ApiServer       string `json:"apiServer"`
	CreateTime      string `json:"createTime"`
	NodeCount       int64  `json:"nodeCount"`
	PodCount        int64  `json:"podCount"`
	NamespaceCount  int64  `json:"namespaceCount"`
	DeploymentCount int64  `json:"deploymentCount"`
}

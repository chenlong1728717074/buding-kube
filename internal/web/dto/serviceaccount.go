package dto

type ServiceAccountCreateDTO struct {
	BaseDTO
	//  是否自动将 ServiceAccount 的 token 挂载到 Pod
	// 默认为 true,设置为 false 可以禁止 Pod 自动访问 Kubernetes API,提升安全性
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken"`
	// ImagePullSecrets 镜像拉取 Secret 列表
	// 用于从私有镜像仓库拉取镜像,列表中的 Secret 必须在同一 namespace 中存在
	ImagePullSecrets []string `json:"imagePullSecrets"`
	// Secrets 关联的 Secret 列表  这些 Secret 会被自动挂载到使用此 ServiceAccount 的 Pod 中
	Secrets     []string          `json:"secrets"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

type ServiceAccountBaseDTO struct {
	BaseDTO
	Force bool `json:"force" form:"force"`
}

type ServiceAccountPageQueryBaseDTO struct {
	PageQueryDTO
	ServiceAccountBaseDTO
}

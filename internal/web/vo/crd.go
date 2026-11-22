package vo

// CrdVO CRD 概要(不含 schema)
type CrdVO struct {
	Name       string   `json:"name"`       // CRD 全名, 如 clusters.kube.buding.goaigc.fun
	Kind       string   `json:"kind"`       // 资源 Kind
	ListKind   string   `json:"listKind"`   // 列表 Kind
	Plural     string   `json:"plural"`     // 复数名, 即 REST 路径中的资源名
	Singular   string   `json:"singular"`   // 单数名
	ShortNames []string `json:"shortNames"` // 短名
	Group      string   `json:"group"`      // API 组
	Version    string   `json:"version"`    // 存储版本
	Versions   []string `json:"versions"`   // 可用(served)的版本, 存储版本排在第一个
	Scope      string   `json:"scope"`      // Namespaced / Cluster
	CreateTime string   `json:"createTime"` // 创建时间
}

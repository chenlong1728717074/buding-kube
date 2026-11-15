package vo

// CrdVO CRD 列表返回值
type CrdVO struct {
	Name       string `json:"name"`       // CRD 全名, 如 clusters.kube.buding.goaigc.fun
	Kind       string `json:"kind"`       // 资源 Kind
	Group      string `json:"group"`      // API 组
	Version    string `json:"version"`    // 存储版本
	Scope      string `json:"scope"`      // Namespaced / Cluster
	CreateTime string `json:"createTime"` // 创建时间
}

package vo

// TopologyNodeVO 命名空间资源拓扑节点
type TopologyNodeVO struct {
	ID         string            `json:"id"`
	Kind       string            `json:"kind"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	StatusType string            `json:"statusType"` // success / warning / danger / info
	Layer      int               `json:"layer"`      // 0:入口/伸缩 1:服务 2:工作负载 3:Pod 4:配置/存储/账户
	Summary    string            `json:"summary"`
	Extra      map[string]string `json:"extra,omitempty"`
}

// TopologyEdgeVO 命名空间资源拓扑关系边
type TopologyEdgeVO struct {
	ID       string   `json:"id"`
	Source   string   `json:"source"`
	Target   string   `json:"target"`
	Relation string   `json:"relation"`
	Label    string   `json:"label,omitempty"`
	Details  []string `json:"details,omitempty"`
}

// TopologyVO 命名空间资源关联拓扑图
type TopologyVO struct {
	Namespace string           `json:"namespace"`
	Nodes     []TopologyNodeVO `json:"nodes"`
	Edges     []TopologyEdgeVO `json:"edges"`
	Summary   map[string]int   `json:"summary"`
}

package vo

const (
	SupportYes     = "supported"
	SupportNo      = "unsupported"
	SupportUnknown = "unknown"
)

// NetworkPolicySupportVO 集群网络插件是否支持网络策略
type NetworkPolicySupportVO struct {
	// Plugin 识别到的网络插件, 多个时用 " + " 连接; 未识别时为空
	Plugin string `json:"plugin"`
	// Support supported / unsupported / unknown
	Support string `json:"support"`
	Message string `json:"message"`
}

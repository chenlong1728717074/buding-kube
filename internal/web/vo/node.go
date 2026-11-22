package vo

import (
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const nodeRolePrefix = "node-role.kubernetes.io/"

// NodeResourceVO 节点的资源: 可分配量、容器组的预留(已分配)和上限合计、实时用量.
// CPU 单位为毫核, 内存单位为字节; 用量为 nil 表示没有 metrics-server 数据
type NodeResourceVO struct {
	CPUCapacity       int64  `json:"cpuCapacity"`
	CPUAllocatable    int64  `json:"cpuAllocatable"`
	CPURequests       int64  `json:"cpuRequests"`
	CPULimits         int64  `json:"cpuLimits"`
	CPUUsage          *int64 `json:"cpuUsage"`
	MemoryCapacity    int64  `json:"memoryCapacity"`
	MemoryAllocatable int64  `json:"memoryAllocatable"`
	MemoryRequests    int64  `json:"memoryRequests"`
	MemoryLimits      int64  `json:"memoryLimits"`
	MemoryUsage       *int64 `json:"memoryUsage"`
	PodCount          int64  `json:"podCount"`
	PodAllocatable    int64  `json:"podAllocatable"`
}

type NodeVO struct {
	Hostname string   `json:"hostname"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles"`
	Status   string   `json:"status"`
	IP       string   `json:"ip"`
	// CPU / Memory 可分配量的文本, 兼容旧页面
	CPU            string    `json:"cpu"`
	Memory         string    `json:"memory"`
	UnSchedule     bool      `json:"unSchedule"`
	KubeletVersion string    `json:"kubeletVersion"`
	TaintCount     int       `json:"taintCount"`
	CreateTime     time.Time `json:"createTime"`
	NodeResourceVO
}

type NodeInfoVO struct {
	Serve       NodeServerInfoVO      `json:"server"`
	Runtime     NodeRuntimeVO         `json:"runtime"`
	Resources   NodeResourceVO        `json:"resources"`
	Roles       []string              `json:"roles"`
	Addresses   []NodeAddressVO       `json:"addresses"`
	Taints      []NodeTaintVO         `json:"taints"`
	Conditions  []WorkloadConditionVO `json:"conditions"`
	Events      []EventVO             `json:"events"`
	Pods        []JobPodVO            `json:"pods"`
	Annotations map[string]string     `json:"annotations"`
	Labels      map[string]string     `json:"labels"`
	PodCIDR     string                `json:"podCIDR"`
	ProviderID  string                `json:"providerID"`
	Uid         string                `json:"uid"`
}

type NodeServerInfoVO struct {
	Hostname         string    `json:"hostname"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	IP               string    `json:"ip"`
	UnSchedule       bool      `json:"unSchedule"`
	OsImage          string    `json:"osImage"`
	OsType           string    `json:"osType"`
	Arch             string    `json:"arch"`
	KernelVersion    string    `json:"kernelVersion"`
	ContainerRuntime string    `json:"containerRuntime"`
	KubeletVersion   string    `json:"kubeletVersion"`
	KubeProxyVersion string    `json:"kubeProxyVersion"`
	CreateTime       time.Time `json:"createTime"`
}

type NodeRuntimeVO struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	EphemeralStorage string `json:"ephemeralStorage"`
	NvidiaGPU        string `json:"nvidiaGpu,omitempty"`
	AMDGPU           string `json:"amdGpu,omitempty"`
	IntelGPU         string `json:"intelGpu,omitempty"`
}

type NodeAddressVO struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type NodeTaintVO struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
	// System 由节点控制器维护(例如 node.kubernetes.io/not-ready), 不能手动修改
	System    bool       `json:"system"`
	TimeAdded *time.Time `json:"timeAdded"`
}

type ResourceMetricVO struct {
	Used      float64 `json:"used"`
	Total     float64 `json:"total"`
	Percent   float64 `json:"percent"`
	UsedText  string  `json:"usedText"`
	TotalText string  `json:"totalText"`
}

type ClusterResourceUsageVO struct {
	CPU        ResourceMetricVO `json:"cpu"`
	Memory     ResourceMetricVO `json:"memory"`
	Storage    ResourceMetricVO `json:"storage"`
	NodeTotal  int64            `json:"nodeTotal"`
	NodeReady  int64            `json:"nodeReady"`
	PodTotal   int64            `json:"podTotal"`
	PodRunning int64            `json:"podRunning"`
	Source     string           `json:"source"`
}

// DrainPodVO 排空节点时每个容器组的处理方式
type DrainPodVO struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	OwnerKind string `json:"ownerKind"`
	OwnerName string `json:"ownerName"`
	Status    string `json:"status"`
	// Action evict(驱逐) / delete(已结束, 直接删除) / skip(跳过) / block(阻止排空)
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// DrainPreviewVO 排空预览
type DrainPreviewVO struct {
	Pods    []DrainPodVO `json:"pods"`
	Evict   int          `json:"evict"`
	Skip    int          `json:"skip"`
	Blocked int          `json:"blocked"`
}

// DrainResultVO 排空结果
type DrainResultVO struct {
	Cordoned bool         `json:"cordoned"`
	Evicted  []string     `json:"evicted"`
	Failed   []DrainPodVO `json:"failed"`
	Skipped  int          `json:"skipped"`
}

// IsSystemTaint 由节点控制器 / 云厂商组件维护的污点
func IsSystemTaint(key string) bool {
	return strings.HasPrefix(key, "node.kubernetes.io/") || strings.HasPrefix(key, "node.cloudprovider.kubernetes.io/")
}

// NodeRoles 节点角色(node-role.kubernetes.io/<角色> 标签), 没有时为 worker
func NodeRoles(node *corev1.Node) []string {
	roles := make([]string, 0)
	for k := range node.Labels {
		if strings.HasPrefix(k, nodeRolePrefix) {
			if role := strings.TrimPrefix(k, nodeRolePrefix); role != "" {
				roles = append(roles, role)
			}
		}
	}
	if role := node.Labels["kubernetes.io/role"]; role != "" && len(roles) == 0 {
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	sort.Strings(roles)
	return roles
}

// NodeBaseResources 节点的容量和可分配量; 容器组的统计和用量由 service 层补充
func NodeBaseResources(node *corev1.Node) NodeResourceVO {
	alloc, capacity := node.Status.Allocatable, node.Status.Capacity
	r := NodeResourceVO{
		CPUCapacity:       capacity.Cpu().MilliValue(),
		CPUAllocatable:    alloc.Cpu().MilliValue(),
		MemoryCapacity:    capacity.Memory().Value(),
		MemoryAllocatable: alloc.Memory().Value(),
		PodAllocatable:    alloc.Pods().Value(),
	}
	if r.CPUAllocatable == 0 {
		r.CPUAllocatable = r.CPUCapacity
	}
	if r.MemoryAllocatable == 0 {
		r.MemoryAllocatable = r.MemoryCapacity
	}
	if r.PodAllocatable == 0 {
		r.PodAllocatable = capacity.Pods().Value()
	}
	return r
}

// AddPod 把一个未结束的容器组计入节点的已分配资源
func (r *NodeResourceVO) AddPod(pod *corev1.Pod) {
	total := PodResourceTotals(&pod.Spec)
	r.PodCount++
	r.CPURequests += total.CPURequest
	r.CPULimits += total.CPULimit
	r.MemoryRequests += total.MemoryRequest
	r.MemoryLimits += total.MemoryLimit
}

// IsPodActive 未结束的容器组才占用节点资源
func IsPodActive(pod *corev1.Pod) bool {
	return pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed
}

func Node2InfoVO(node *corev1.Node) *NodeInfoVO {
	result := &NodeInfoVO{
		Runtime:     buildNodeRuntime(node),
		Serve:       buildNodeServer(node),
		Resources:   NodeBaseResources(node),
		Roles:       NodeRoles(node),
		Addresses:   []NodeAddressVO{},
		Taints:      []NodeTaintVO{},
		Conditions:  []WorkloadConditionVO{},
		Events:      []EventVO{},
		Pods:        []JobPodVO{},
		Annotations: node.Annotations,
		Labels:      node.Labels,
		PodCIDR:     node.Spec.PodCIDR,
		ProviderID:  node.Spec.ProviderID,
		Uid:         string(node.UID),
	}
	for _, a := range node.Status.Addresses {
		result.Addresses = append(result.Addresses, NodeAddressVO{Type: string(a.Type), Address: a.Address})
	}
	for _, t := range node.Spec.Taints {
		item := NodeTaintVO{Key: t.Key, Value: t.Value, Effect: string(t.Effect), System: IsSystemTaint(t.Key)}
		if t.TimeAdded != nil {
			item.TimeAdded = timePtr(t.TimeAdded.Time)
		}
		result.Taints = append(result.Taints, item)
	}
	for _, c := range node.Status.Conditions {
		result.Conditions = append(result.Conditions, newCondition(string(c.Type), string(c.Status), c.Reason, c.Message, c.LastHeartbeatTime))
	}
	return result
}

func buildNodeServer(node *corev1.Node) NodeServerInfoVO {
	return NodeServerInfoVO{
		Hostname:         node.Name,
		Role:             GetNodeRole(node),
		Status:           GetNodeStatus(node),
		IP:               GetNodeIp(node),
		UnSchedule:       node.Spec.Unschedulable,
		OsImage:          node.Status.NodeInfo.OSImage,
		OsType:           node.Status.NodeInfo.OperatingSystem,
		Arch:             node.Status.NodeInfo.Architecture,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion: node.Status.NodeInfo.KubeProxyVersion,
		CreateTime:       node.CreationTimestamp.Time,
	}
}

func buildNodeRuntime(node *corev1.Node) NodeRuntimeVO {
	result := NodeRuntimeVO{
		CPU:              node.Status.Allocatable.Cpu().String(),
		Memory:           node.Status.Allocatable.Memory().String(),
		EphemeralStorage: node.Status.Allocatable.StorageEphemeral().String(),
	}
	if gpu, exists := node.Status.Allocatable["nvidia.com/gpu"]; exists {
		result.NvidiaGPU = gpu.String()
	}
	if gpu, exists := node.Status.Allocatable["amd.com/gpu"]; exists {
		result.AMDGPU = gpu.String()
	}
	if gpu, exists := node.Status.Allocatable["gpu.intel.com/i915"]; exists {
		result.IntelGPU = gpu.String()
	}
	return result
}

func Node2VO(node *corev1.Node) NodeVO {
	return NodeVO{
		Hostname:       node.Name,
		Role:           GetNodeRole(node),
		Roles:          NodeRoles(node),
		Status:         GetNodeStatus(node),
		IP:             GetNodeIp(node),
		CPU:            node.Status.Allocatable.Cpu().String(),
		Memory:         node.Status.Allocatable.Memory().String(),
		UnSchedule:     node.Spec.Unschedulable,
		KubeletVersion: node.Status.NodeInfo.KubeletVersion,
		TaintCount:     len(node.Spec.Taints),
		CreateTime:     node.CreationTimestamp.Time,
		NodeResourceVO: NodeBaseResources(node),
	}
}

func GetNodeIp(node *corev1.Node) string {
	ip := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip = addr.Address
			break
		}
	}
	return ip
}

func GetNodeStatus(node *corev1.Node) string {
	status := "Unknown"
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}
	return status
}

func GetNodeRole(node *corev1.Node) string {
	role := "worker"
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		role = "master"
	} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		role = "control-plane"
	}
	return role
}

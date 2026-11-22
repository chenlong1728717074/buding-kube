package vo

import (
	corev1 "k8s.io/api/core/v1"
)

// PodResources 容器组的资源设置(CPU 单位为毫核, 内存单位为字节); 上限为 0 表示不限制
type PodResources struct {
	CPURequest    int64 `json:"cpuRequest"`
	CPULimit      int64 `json:"cpuLimit"`
	MemoryRequest int64 `json:"memoryRequest"`
	MemoryLimit   int64 `json:"memoryLimit"`
}

// PodUsage 实时用量(来自 metrics-server); 为 nil 表示没有数据(未安装 metrics-server 或容器组未运行)
type PodUsage struct {
	CPUUsage    *int64 `json:"cpuUsage"`
	MemoryUsage *int64 `json:"memoryUsage"`
}

// PodContainerUsageVO 单个容器的资源设置与用量
type PodContainerUsageVO struct {
	Name string `json:"name"`
	Init bool   `json:"init"`
	PodResources
	PodUsage
}

func containerResources(c *corev1.Container) PodResources {
	r := PodResources{
		CPURequest:    c.Resources.Requests.Cpu().MilliValue(),
		MemoryRequest: c.Resources.Requests.Memory().Value(),
	}
	if q, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
		r.CPULimit = q.MilliValue()
	}
	if q, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
		r.MemoryLimit = q.Value()
	}
	return r
}

func isSidecar(c *corev1.Container) bool {
	return c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

// PodResourceTotals 容器组的资源合计, 与调度器的计算方式一致:
// 工作容器和边车容器相加, 再与普通初始化容器中的最大值比较取大; 任一容器没有上限时整体视为不限制
func PodResourceTotals(spec *corev1.PodSpec) PodResources {
	var total PodResources
	cpuUnlimited, memUnlimited := false, false
	add := func(r PodResources) {
		total.CPURequest += r.CPURequest
		total.MemoryRequest += r.MemoryRequest
		total.CPULimit += r.CPULimit
		total.MemoryLimit += r.MemoryLimit
		cpuUnlimited = cpuUnlimited || r.CPULimit == 0
		memUnlimited = memUnlimited || r.MemoryLimit == 0
	}
	for i := range spec.Containers {
		add(containerResources(&spec.Containers[i]))
	}
	for i := range spec.InitContainers {
		if isSidecar(&spec.InitContainers[i]) {
			add(containerResources(&spec.InitContainers[i]))
		}
	}
	for i := range spec.InitContainers {
		c := &spec.InitContainers[i]
		if isSidecar(c) {
			continue
		}
		r := containerResources(c)
		total.CPURequest = max(total.CPURequest, r.CPURequest)
		total.MemoryRequest = max(total.MemoryRequest, r.MemoryRequest)
		if !cpuUnlimited {
			total.CPULimit = max(total.CPULimit, r.CPULimit)
		}
		if !memUnlimited {
			total.MemoryLimit = max(total.MemoryLimit, r.MemoryLimit)
		}
	}
	if cpuUnlimited {
		total.CPULimit = 0
	}
	if memUnlimited {
		total.MemoryLimit = 0
	}
	return total
}

// PodContainerUsages 每个容器的资源设置, 用量由 service 层补充
func PodContainerUsages(spec *corev1.PodSpec) []PodContainerUsageVO {
	result := make([]PodContainerUsageVO, 0, len(spec.InitContainers)+len(spec.Containers))
	for i := range spec.InitContainers {
		c := &spec.InitContainers[i]
		result = append(result, PodContainerUsageVO{Name: c.Name, Init: true, PodResources: containerResources(c)})
	}
	for i := range spec.Containers {
		c := &spec.Containers[i]
		result = append(result, PodContainerUsageVO{Name: c.Name, PodResources: containerResources(c)})
	}
	return result
}

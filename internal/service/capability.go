package service

import (
	"buding-kube/internal/web/vo"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 网络策略由网络插件实现, Kubernetes 本身不会拒绝创建, 插件不支持时策略不生效.
// 这里按集群中运行的网络插件组件(DaemonSet)判断, 结果缓存一段时间
const networkPolicyCacheTTL = 5 * time.Minute

type cniRule struct {
	plugin  string
	support bool
	// match 根据 DaemonSet 名称和容器镜像判断
	match func(name string, images []string) bool
}

func nameOrImage(names []string, images ...string) func(string, []string) bool {
	return func(name string, imgs []string) bool {
		for _, n := range names {
			if name == n || strings.HasPrefix(name, n+"-") {
				return true
			}
		}
		for _, img := range imgs {
			for _, want := range images {
				if strings.Contains(img, want) {
					return true
				}
			}
		}
		return false
	}
}

// 支持网络策略的插件排在前面; Canal = Flannel + Calico 策略
var cniRules = []cniRule{
	{"Calico", true, nameOrImage([]string{"calico-node"}, "calico/node")},
	{"Canal", true, nameOrImage([]string{"canal", "rke2-canal"})},
	{"Cilium", true, nameOrImage([]string{"cilium"}, "cilium/cilium")},
	{"Antrea", true, nameOrImage([]string{"antrea-agent"}, "antrea/antrea")},
	{"Weave Net", true, nameOrImage([]string{"weave-net"}, "weaveworks/weave-kube")},
	{"kube-router", true, nameOrImage([]string{"kube-router"}, "kube-router")},
	{"Kube-OVN", true, nameOrImage([]string{"kube-ovn-cni"}, "kube-ovn")},
	{"Azure NPM", true, nameOrImage([]string{"azure-npm"})},
	{"GKE Dataplane V2", true, nameOrImage([]string{"anetd"})},
	{"Flannel", false, nameOrImage([]string{"kube-flannel-ds", "kube-flannel", "flannel"}, "flannel/flannel", "flannel-io/flannel")},
}

type networkPolicyCache struct {
	at     time.Time
	result vo.NetworkPolicySupportVO
}

var networkPolicySupport sync.Map

func daemonSetImages(ds *appsv1.DaemonSet) []string {
	spec := ds.Spec.Template.Spec
	images := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))
	for _, c := range spec.InitContainers {
		images = append(images, c.Image)
	}
	for _, c := range spec.Containers {
		images = append(images, c.Image)
	}
	return images
}

// awsNetworkPolicyEnabled AWS VPC CNI 从 1.14 开始支持网络策略, 需要在 aws-node 的 aws-eks-nodeagent 容器中开启
func awsNetworkPolicyEnabled(ds *appsv1.DaemonSet) bool {
	for _, c := range ds.Spec.Template.Spec.Containers {
		if c.Name != "aws-eks-nodeagent" {
			continue
		}
		for _, arg := range c.Args {
			if arg == "--enable-network-policy=true" {
				return true
			}
		}
	}
	return false
}

// NetworkPolicySupport 判断集群网络插件是否支持网络策略
func NetworkPolicySupport(ctx context.Context, clusterId string) (*vo.NetworkPolicySupportVO, error) {
	if v, ok := networkPolicySupport.Load(clusterId); ok {
		if c := v.(networkPolicyCache); time.Since(c.at) < networkPolicyCacheTTL {
			result := c.result
			return &result, nil
		}
	}
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return nil, err
	}
	list, err := cs.AppsV1().DaemonSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, fmt.Errorf("查询网络插件失败: %v", err)
	}

	supported, unsupported := make([]string, 0), make([]string, 0)
	seen := map[string]bool{}
	add := func(plugin string, ok bool) {
		if seen[plugin] {
			return
		}
		seen[plugin] = true
		if ok {
			supported = append(supported, plugin)
		} else {
			unsupported = append(unsupported, plugin)
		}
	}
	for i := range list.Items {
		ds := &list.Items[i]
		if ds.Name == "aws-node" {
			if awsNetworkPolicyEnabled(ds) {
				add("AWS VPC CNI（已开启网络策略）", true)
			} else {
				add("AWS VPC CNI（未开启网络策略）", false)
			}
			continue
		}
		images := daemonSetImages(ds)
		for _, rule := range cniRules {
			if rule.match(ds.Name, images) {
				add(rule.plugin, rule.support)
				break
			}
		}
	}

	result := vo.NetworkPolicySupportVO{}
	switch {
	case len(supported) > 0:
		result.Plugin = strings.Join(append(supported, unsupported...), " + ")
		result.Support = vo.SupportYes
		result.Message = fmt.Sprintf("网络插件 %s 支持网络策略", strings.Join(supported, "、"))
	case len(unsupported) > 0:
		result.Plugin = strings.Join(unsupported, " + ")
		result.Support = vo.SupportNo
		result.Message = fmt.Sprintf("网络插件 %s 不支持网络策略，创建的策略不会生效，流量不受限制", result.Plugin)
	default:
		// K3s 内置 Flannel(不以 DaemonSet 运行)和基于 kube-router 的网络策略控制器(默认开启)
		if nodes, nodeErr := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{ResourceVersion: "0", Limit: 1}); nodeErr == nil &&
			len(nodes.Items) > 0 && strings.Contains(nodes.Items[0].Status.NodeInfo.KubeletVersion, "+k3s") {
			result.Plugin = "K3s"
			result.Support = vo.SupportYes
			result.Message = "K3s 内置的网络策略控制器支持网络策略（启动时使用了 --disable-network-policy 的除外）"
		} else {
			result.Support = vo.SupportUnknown
			result.Message = "未能识别集群的网络插件，请确认插件支持网络策略（Calico、Cilium 等支持，Flannel 不支持）"
		}
	}
	networkPolicySupport.Store(clusterId, networkPolicyCache{at: time.Now(), result: result})
	return &result, nil
}

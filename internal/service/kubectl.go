package service

import (
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// Web kubectl: 与 KubeSphere 相同, 在目标集群中为每个用户运行一个带 kubectl / helm 的容器组,
// 使用绑定了 cluster-admin 的服务账户, 终端通过 exec 进入该容器组.
const (
	kubectlNamespace      = "buding-kubectl"
	kubectlServiceAccount = "buding-kubectl"
	kubectlBinding        = "buding-kubectl-cluster-admin"
	KubectlContainer      = "kubectl"
	// kubectlDefaultImage 包含 kubectl、helm、bash, 可通过环境变量 KUBECTL_IMAGE 替换(例如内网镜像仓库)
	kubectlDefaultImage = "alpine/k8s:1.35.10"
	// kubectlLifetime 容器组运行 24 小时后自动退出, 下次进入时重新创建
	kubectlLifetime     = "86400"
	kubectlReadyTimeout = 3 * time.Minute
	managedByLabel      = "app.kubernetes.io/managed-by"
	managedByValue      = "buding-kube"
)

var kubectlNameInvalid = regexp.MustCompile(`[^a-z0-9-]+`)

func kubectlImage() string {
	if img := strings.TrimSpace(os.Getenv("KUBECTL_IMAGE")); img != "" {
		return img
	}
	return kubectlDefaultImage
}

// kubectlPodName 每个用户一个容器组; 用户名可能包含不能用于名称的字符, 追加哈希避免冲突
func kubectlPodName(username string) string {
	sum := sha256.Sum256([]byte(username))
	name := strings.Trim(kubectlNameInvalid.ReplaceAllString(strings.ToLower(username), "-"), "-")
	if len(name) > 40 {
		name = strings.Trim(name[:40], "-")
	}
	if name == "" {
		name = "user"
	}
	return "kubectl-" + name + "-" + hex.EncodeToString(sum[:4])
}

func permissionHint(err error, what string) error {
	if apierrors.IsForbidden(err) {
		return fmt.Errorf("平台使用的集群凭证没有权限%s，Web kubectl 需要能创建命名空间、服务账户和 ClusterRoleBinding", what)
	}
	return fmt.Errorf("%s失败: %v", what, err)
}

// ensureKubectlRBAC 创建命名空间、服务账户和 cluster-admin 绑定(已存在时跳过)
func ensureKubectlRBAC(ctx context.Context, cs *kubernetes.Clientset) error {
	labels := map[string]string{managedByLabel: managedByValue}
	if _, err := cs.CoreV1().Namespaces().Get(ctx, kubectlNamespace, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: kubectlNamespace, Labels: labels}}
		if _, err = cs.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return permissionHint(err, "创建命名空间 "+kubectlNamespace)
		}
	} else if err != nil {
		return permissionHint(err, "查询命名空间 "+kubectlNamespace)
	}
	if _, err := cs.CoreV1().ServiceAccounts(kubectlNamespace).Get(ctx, kubectlServiceAccount, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: kubectlServiceAccount, Namespace: kubectlNamespace, Labels: labels}}
		if _, err = cs.CoreV1().ServiceAccounts(kubectlNamespace).Create(ctx, sa, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return permissionHint(err, "创建服务账户")
		}
	} else if err != nil {
		return permissionHint(err, "查询服务账户")
	}
	if _, err := cs.RbacV1().ClusterRoleBindings().Get(ctx, kubectlBinding, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: kubectlBinding, Labels: labels},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "cluster-admin"},
			Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: kubectlServiceAccount, Namespace: kubectlNamespace}},
		}
		if _, err = cs.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return permissionHint(err, "创建 ClusterRoleBinding")
		}
	} else if err != nil {
		return permissionHint(err, "查询 ClusterRoleBinding")
	}
	return nil
}

func newKubectlPod(name, username string) *corev1.Pod {
	grace := int64(0)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   kubectlNamespace,
			Labels:      map[string]string{"app.kubernetes.io/name": "buding-kubectl", managedByLabel: managedByValue},
			Annotations: map[string]string{consts.CreatorAnnotation: username},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName:            kubectlServiceAccount,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: &grace,
			NodeSelector:                  map[string]string{"kubernetes.io/os": "linux"},
			Containers: []corev1.Container{{
				Name:            KubectlContainer,
				Image:           kubectlImage(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sleep", kubectlLifetime},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
					Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("512Mi")},
				},
			}},
		},
	}
}

// podUsable 运行中且镜像与当前配置一致
func podUsable(pod *corev1.Pod) bool {
	return pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning &&
		len(pod.Spec.Containers) > 0 && pod.Spec.Containers[0].Image == kubectlImage()
}

// EnsureKubectlPod 准备用户的 kubectl 容器组并等待运行, progress 用于向终端输出进度
func EnsureKubectlPod(ctx context.Context, clusterId, username string, progress func(string)) (string, string, error) {
	cs, err := ClusterMap.GetClientSet(clusterId)
	if err != nil {
		return "", "", err
	}
	if err = ensureKubectlRBAC(ctx, cs); err != nil {
		return "", "", err
	}
	name := kubectlPodName(username)
	pods := cs.CoreV1().Pods(kubectlNamespace)

	pod, err := pods.Get(ctx, name, metav1.GetOptions{})
	switch {
	case err == nil && podUsable(pod):
		return kubectlNamespace, name, nil
	case err == nil && pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodPending && pod.Spec.Containers[0].Image == kubectlImage():
		// 正在启动, 直接等待
	case err == nil:
		// 已退出(超过 24 小时)或镜像配置已变更, 删除后重新创建
		progress("正在回收旧的 kubectl 环境…")
		if delErr := pods.Delete(ctx, name, metav1.DeleteOptions{}); delErr != nil && !apierrors.IsNotFound(delErr) {
			return "", "", permissionHint(delErr, "删除旧的容器组")
		}
		if waitErr := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
			_, getErr := pods.Get(ctx, name, metav1.GetOptions{})
			return apierrors.IsNotFound(getErr), nil
		}); waitErr != nil {
			return "", "", errors.New("等待旧的 kubectl 容器组删除超时，请稍后重试")
		}
		pod = nil
	case apierrors.IsNotFound(err):
		pod = nil
	default:
		return "", "", permissionHint(err, "查询 kubectl 容器组")
	}
	if pod == nil {
		progress(fmt.Sprintf("正在创建 kubectl 环境（镜像 %s，首次使用需要拉取镜像，可能需要几分钟）…", kubectlImage()))
		if _, err = pods.Create(ctx, newKubectlPod(name, username), metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return "", "", permissionHint(err, "创建 kubectl 容器组")
		}
		logs.Info("为用户 %s 在集群 %s 创建 kubectl 容器组 %s/%s", username, clusterId, kubectlNamespace, name)
	}

	lastHint := ""
	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, kubectlReadyTimeout, true, func(ctx context.Context) (bool, error) {
		p, getErr := pods.Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return false, nil
		}
		if podUsable(p) {
			return true, nil
		}
		if p.Status.Phase == corev1.PodFailed || p.Status.Phase == corev1.PodSucceeded {
			return false, fmt.Errorf("kubectl 容器组已退出: %s", firstNonEmptyString(p.Status.Message, p.Status.Reason, string(p.Status.Phase)))
		}
		for _, c := range p.Status.Conditions {
			if c.Type == corev1.PodScheduled && c.Status == corev1.ConditionFalse && c.Reason == corev1.PodReasonUnschedulable {
				return false, fmt.Errorf("kubectl 容器组无法调度: %s", c.Message)
			}
		}
		for _, st := range p.Status.ContainerStatuses {
			if w := st.State.Waiting; w != nil {
				switch w.Reason {
				case "ErrImagePull", "ImagePullBackOff", "InvalidImageName":
					return false, fmt.Errorf("拉取镜像 %s 失败: %s。内网环境请设置后端环境变量 KUBECTL_IMAGE 指向可访问的镜像", kubectlImage(), w.Message)
				case "ContainerCreating":
					if lastHint != w.Reason {
						lastHint = w.Reason
						progress("正在拉取镜像并启动容器…")
					}
				}
			}
		}
		return false, nil
	})
	if err != nil {
		// 拉取镜像失败、无法调度或超时时删除容器组, 不让它一直占用资源预留; 下次连接时重新创建
		cleanupKubectlPod(cs, name)
		if errors.Is(err, context.DeadlineExceeded) || wait.Interrupted(err) {
			return "", "", errors.New("等待 kubectl 容器组启动超时，已清理，请稍后重试或检查镜像是否可以拉取")
		}
		return "", "", fmt.Errorf("%v（已清理未启动的容器组）", err)
	}
	return kubectlNamespace, name, nil
}

// cleanupKubectlPod 删除没有启动成功的容器组; 调用方的上下文可能已超时, 使用新的上下文
func cleanupKubectlPod(cs *kubernetes.Clientset, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cs.CoreV1().Pods(kubectlNamespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		logs.Warn("清理 kubectl 容器组失败: %s/%s %v", kubectlNamespace, name, err)
	}
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

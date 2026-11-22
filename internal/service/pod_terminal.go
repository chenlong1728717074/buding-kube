package service

import (
	"buding-kube/internal/web/dto"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	utilexec "k8s.io/client-go/util/exec"
)

// defaultContainerAnnotation 与 kubectl exec 一致, 未指定容器时优先使用该注解中的容器
const defaultContainerAnnotation = "kubectl.kubernetes.io/default-container"

// terminalShells 通过 /bin/sh 启动, 统一设置 TERM, 避免 clear / top / vi 等命令报 "TERM environment variable not set"
var terminalShells = map[string]string{
	"auto": "if command -v bash >/dev/null 2>&1; then exec bash; elif command -v ash >/dev/null 2>&1; then exec ash; else exec sh; fi",
	"bash": "exec bash",
	"sh":   "exec sh",
}

// PodTerminal 已校验过的终端会话, 调用 Stream 开始交互
type PodTerminal struct {
	Container string
	Shell     string
	executor  remotecommand.Executor
}

// Stream 连接容器的标准输入输出; 使用 TTY 时 stderr 合并到 stdout
func (t *PodTerminal) Stream(ctx context.Context, stdin io.Reader, stdout io.Writer, sizeQueue remotecommand.TerminalSizeQueue) error {
	return t.executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})
}

// OpenTerminal 校验容器组和容器, 创建执行器(此时还没有连接到容器)
func (s *PodService) OpenTerminal(ctx context.Context, req dto.PodTerminalDTO) (*PodTerminal, error) {
	shell := strings.TrimSpace(req.Shell)
	if shell == "" {
		shell = "auto"
	}
	script, ok := terminalShells[shell]
	if !ok {
		return nil, fmt.Errorf("不支持的 Shell: %s", req.Shell)
	}
	cache, err := ClusterMap.GetCache(req.ClusterId)
	if err != nil {
		return nil, errors.New("获取集群失败")
	}
	pod, err := cache.clientSet.CoreV1().Pods(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("容器组 %s 不存在", req.Name)
		}
		return nil, fmt.Errorf("获取容器组失败: %v", err)
	}
	if pod.DeletionTimestamp != nil {
		return nil, errors.New("容器组正在删除，无法打开终端")
	}
	container, err := pickTerminalContainer(pod, strings.TrimSpace(req.Container))
	if err != nil {
		return nil, err
	}

	execReq := cache.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/sh", "-c", "export TERM=xterm-256color; " + script},
			Stdin:     true,
			Stdout:    true,
			// TTY 模式下 stderr 已合并到 stdout, kubelet 不允许两者同时开启
			Stderr: false,
			TTY:    true,
		}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(cache.config, "POST", execReq.URL())
	if err != nil {
		return nil, fmt.Errorf("创建执行器失败: %v", err)
	}
	return &PodTerminal{Container: container, Shell: shell, executor: executor}, nil
}

// pickTerminalContainer 未指定时依次使用默认容器注解、第一个运行中的工作容器
func pickTerminalContainer(pod *v1.Pod, name string) (string, error) {
	running := map[string]bool{}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Running != nil {
			running[cs.Name] = true
		}
	}
	// 初始化容器中的边车容器(restartPolicy: Always)会一直运行, 也允许进入
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Running != nil {
			running[cs.Name] = true
		}
	}
	if name == "" {
		name = pod.Annotations[defaultContainerAnnotation]
		if !running[name] {
			for _, c := range pod.Spec.Containers {
				if running[c.Name] {
					name = c.Name
					break
				}
			}
		}
	}
	if name == "" {
		return "", errors.New("容器组中没有运行中的容器")
	}
	exists := false
	for _, c := range pod.Spec.Containers {
		exists = exists || c.Name == name
	}
	for _, c := range pod.Spec.InitContainers {
		exists = exists || c.Name == name
	}
	if !exists {
		return "", fmt.Errorf("容器 %s 不存在", name)
	}
	if !running[name] {
		return "", fmt.Errorf("容器 %s 没有在运行，无法打开终端", name)
	}
	return name, nil
}

// TerminalExitStatus Shell 退出时返回退出码; 其他错误(连接失败等)返回 false
func TerminalExitStatus(err error) (int, bool) {
	if err == nil {
		return 0, true
	}
	var exitErr utilexec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitStatus(), true
	}
	return 0, false
}

// TerminalErrorMessage 把连接错误转换为提示
func TerminalErrorMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "no such file or directory"), strings.Contains(msg, "executable file not found"):
		return "容器中没有 /bin/sh（例如 distroless 等精简镜像），无法打开终端"
	case strings.Contains(msg, "Forbidden"), strings.Contains(msg, "forbidden"):
		return "没有权限执行 pods/exec，请检查集群凭证的 RBAC 权限"
	default:
		return "终端连接失败: " + msg
	}
}

package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"

	"buding-kube/pkg/logs"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// maxResumeRetries 下载流中断后的最大续传次数, 避免文件变化等情况下无限重试
const maxResumeRetries = 5

// 用户输入的路径只通过 sh 的位置参数 "$1" 传入, 不会被拼接进脚本, 也不会被 shell 再次解析,
// 从根本上杜绝命令注入(例如 filePath = "/tmp; rm -rf /data")
const (
	// sizeScript 计算 tar 包大小; 文件不存在时直接报错, 避免得到一个错误的大小
	sizeScript = `if [ ! -e "$1" ]; then echo "file not found: $1" >&2; exit 2; fi; tar cf - "$1" | wc -c`
	// readScript 从第 $2 个字节开始输出 tar 包, 用于断点续传
	readScript = `tar cf - "$1" | tail -c+"$2"`
)

// ErrInvalidPath 非法的容器内路径
var ErrInvalidPath = errors.New("非法的文件路径")

// SanitizePath 校验并规范化容器内的文件路径(统一为 Linux 风格)
func SanitizePath(p string) (string, error) {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if p == "" {
		return "", fmt.Errorf("%w: 路径不能为空", ErrInvalidPath)
	}
	if strings.ContainsAny(p, "\x00\r\n") {
		return "", fmt.Errorf("%w: 包含非法字符", ErrInvalidPath)
	}
	p = path.Clean(p)
	// 以 '-' 开头会被 tar 当成命令行选项
	if strings.HasPrefix(p, "-") {
		return "", fmt.Errorf("%w: 不能以 '-' 开头", ErrInvalidPath)
	}
	return p, nil
}

// TarPipe 通过 exec 在容器内执行 tar, 以流的方式读取文件, 断流后自动续传
type TarPipe struct {
	ctx       context.Context
	config    *rest.Config
	client    rest.Interface
	namespace string
	name      string
	container string
	filePath  string

	reader    *io.PipeReader
	size      uint64
	bytesRead uint64
	retries   int
}

// NewTarPipe 创建 tar 管道. ctx 应传入请求的 Context, 客户端断开时容器内的 tar 进程会随之终止
func NewTarPipe(ctx context.Context, config *rest.Config, client rest.Interface, namespace, name, container, filePath string) (*TarPipe, error) {
	cleanPath, err := SanitizePath(filePath)
	if err != nil {
		return nil, err
	}
	t := &TarPipe{
		ctx:       ctx,
		config:    config,
		client:    client,
		namespace: namespace,
		name:      name,
		container: container,
		filePath:  cleanPath,
	}
	if err = t.fetchSize(); err != nil {
		return nil, err
	}
	if err = t.startStream(0); err != nil {
		return nil, err
	}
	return t, nil
}

// Size tar 包的总字节数
func (t *TarPipe) Size() uint64 {
	return t.size
}

// newExecutor 构造执行器: sh -c <script> sh <args...>, args 依次对应脚本中的 $1 $2 ...
func (t *TarPipe) newExecutor(script string, args ...string) (remotecommand.Executor, error) {
	command := append([]string{"sh", "-c", script, "sh"}, args...)
	req := t.client.Post().
		Resource("pods").
		Name(t.name).
		Namespace(t.namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: t.container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	return remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
}

// fetchSize 获取 tar 包大小
func (t *TarPipe) fetchSize() error {
	executor, err := t.newExecutor(sizeScript, t.filePath)
	if err != nil {
		return fmt.Errorf("创建执行器失败: %w", err)
	}
	var stdout, stderr bytes.Buffer
	if err = executor.StreamWithContext(t.ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return fmt.Errorf("获取文件大小失败: %v %s", err, strings.TrimSpace(stderr.String()))
	}
	size, err := strconv.ParseUint(strings.TrimSpace(stdout.String()), 10, 64)
	if err != nil {
		return fmt.Errorf("解析文件大小失败: %w", err)
	}
	if size == 0 {
		return fmt.Errorf("文件无法读取: %s", strings.TrimSpace(stderr.String()))
	}
	t.size = size
	return nil
}

// startStream 从 offset 字节处(从 0 开始计)开始读取 tar 流
func (t *TarPipe) startStream(offset uint64) error {
	executor, err := t.newExecutor(readScript, t.filePath, strconv.FormatUint(offset+1, 10))
	if err != nil {
		return fmt.Errorf("创建执行器失败: %w", err)
	}
	reader, writer := io.Pipe()
	t.reader = reader
	go func() {
		// 只使用局部变量, 不再写外层的 err, 消除数据竞争
		var stderr bytes.Buffer
		streamErr := executor.StreamWithContext(t.ctx, remotecommand.StreamOptions{
			Stdout: writer,
			Stderr: &stderr,
		})
		if streamErr != nil {
			logs.Warn("读取文件流中断: %v %s", streamErr, strings.TrimSpace(stderr.String()))
		}
		// streamErr 为 nil 时等价于 Close, 读端收到 io.EOF
		_ = writer.CloseWithError(streamErr)
	}()
	return nil
}

// Read 实现 io.Reader. 流异常中断时从已读位置续传, 超过最大重试次数后返回错误
func (t *TarPipe) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)
	// 先累计本次读到的字节, 即使同时返回了错误
	t.bytesRead += uint64(n)
	if err == nil {
		return n, nil
	}
	// 已读完全部内容
	if t.bytesRead >= t.size {
		return n, io.EOF
	}
	// 调用方已关闭或请求已取消, 不再续传
	if errors.Is(err, io.ErrClosedPipe) {
		return n, err
	}
	if ctxErr := t.ctx.Err(); ctxErr != nil {
		return n, ctxErr
	}
	if t.retries >= maxResumeRetries {
		return n, fmt.Errorf("文件传输中断, 已重试 %d 次仍失败(已传输 %d/%d 字节): %w", t.retries, t.bytesRead, t.size, err)
	}
	t.retries++
	logs.Warn("文件流中断, 从第 %d 字节开始第 %d 次续传: %v", t.bytesRead, t.retries, err)
	_ = t.reader.Close()
	if startErr := t.startStream(t.bytesRead); startErr != nil {
		return n, startErr
	}
	return n, nil
}

// Close 关闭管道, 容器内的 tar 进程会因写入失败而退出
func (t *TarPipe) Close() error {
	if t.reader != nil {
		return t.reader.Close()
	}
	return nil
}

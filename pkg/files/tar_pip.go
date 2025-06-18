package files

import (
	"buding-kube/pkg/logs"
	"context"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"strconv"
	"strings"
)

// tarPipe 结构体用于处理tar格式的文件传输
type tarPipe struct {
	config    *rest.Config
	client    rest.Interface
	reader    *io.PipeReader
	outStream *io.PipeWriter
	bytesRead uint64
	size      uint64
	ctx       context.Context
	namespace string
	name      string
	container string
	filePath  string
}

// newTarPipe 创建新的tar管道
func NewTarPipe(ctx context.Context, config *rest.Config, client rest.Interface, namespace, name, container, filePath string) (*tarPipe, error) {
	t := &tarPipe{
		config:    config,
		client:    client,
		namespace: namespace,
		name:      name,
		container: container,
		filePath:  filePath,
		ctx:       ctx,
	}

	if err := t.getFileSize(); err != nil {
		return nil, err
	}
	if err := t.initReadFrom(0); err != nil {
		return nil, err
	}
	return t, nil
}

// getFileSize 获取tar文件的大小
func (t *tarPipe) getFileSize() error {
	req := t.client.Post().
		Resource("pods").
		Name(t.name).
		Namespace(t.namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: t.container,
			Command:   []string{"sh", "-c", fmt.Sprintf("tar cf - %s | wc -c", t.filePath)},
			Stdin:     false,
			Stdout:    true,
			Stderr:    false,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		if err = exec.StreamWithContext(t.ctx, remotecommand.StreamOptions{
			Stdin:             nil,
			Stdout:            writer,
			Stderr:            nil,
			TerminalSizeQueue: nil,
		}); err != nil {
			logs.Error("获取文件大小失败: %v", err)
		}
	}()

	result, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	num, err := strconv.ParseUint(strings.TrimSpace(string(result)), 10, 64)
	if err != nil {
		return err
	}
	t.size = num
	return nil
}

// initReadFrom 初始化从指定位置开始读取
func (t *tarPipe) initReadFrom(n uint64) error {
	t.reader, t.outStream = io.Pipe()

	req := t.client.Post().
		Resource("pods").
		Name(t.name).
		Namespace(t.namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: t.container,
			Command:   []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", t.filePath, n+1)},
			Stdin:     false,
			Stdout:    true,
			Stderr:    false,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, "POST", req.URL())
	if err != nil {
		return err
	}

	go func() {
		defer t.outStream.Close()
		if err = exec.StreamWithContext(t.ctx, remotecommand.StreamOptions{
			Stdin:             nil,
			Stdout:            t.outStream,
			Stderr:            nil,
			TerminalSizeQueue: nil,
		}); err != nil {
			logs.Error("读取文件流失败: %v", err)
		}
	}()
	return nil
}

// Read 实现io.Reader接口
func (t *tarPipe) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)
	if err != nil {
		if t.bytesRead == t.size {
			return n, io.EOF
		}
		return n, t.initReadFrom(t.bytesRead)
	}
	t.bytesRead += uint64(n)
	return n, nil
}

// Close 关闭tar管道
func (t *tarPipe) Close() error {
	if t.reader != nil {
		return t.reader.Close()
	}
	return nil
}

package service

import (
	"archive/tar"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/files"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/yaml"
)

// defaultLogTailLines 未指定起始时间和行数时默认返回的日志行数
const defaultLogTailLines int64 = 1000

var (
	podSrv  *PodService
	podOnce sync.Once
)

type PodService struct {
}

func GetSingletonPodService() *PodService {
	podOnce.Do(func() {
		podSrv = NewPodService()
	})
	return podSrv
}

func NewPodService() *PodService {
	return &PodService{}
}

func (s *PodService) Delete(pod dto.PodDTO) error {
	clientSet, err := ClusterMap.GetClientSet(pod.ClusterId)
	if err != nil {
		return err
	}
	options := metav1.DeleteOptions{}
	// 强制删除
	if pod.Force {
		gracePeriod := int64(0)
		options.GracePeriodSeconds = &gracePeriod
	}
	if err = clientSet.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, options); err != nil {
		logs.Error("删除pod失败: %s %s/%s %v", pod.ClusterId, pod.Namespace, pod.Name, err)
		return err
	}
	return nil
}

// List 查询 Pod 列表.
// 列表只返回概要信息, YAML 由详情接口按需加载, 避免 Pod 较多时逐个序列化、响应体过大
func (s *PodService) List(query dto.PodQueryDTO) ([]vo.PodVO, error) {
	clientSet, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{}
	if query.Status != "" {
		listOptions.FieldSelector = fmt.Sprintf("status.phase=%s", query.Status)
	}
	pods, err := clientSet.CoreV1().Pods(query.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取pod失败: %v", err)
		return nil, err
	}
	result := make([]vo.PodVO, 0, len(pods.Items))
	for i := range pods.Items {
		item := &pods.Items[i]
		if query.Keyword != "" && !strings.Contains(item.Name, query.Keyword) {
			continue
		}
		result = append(result, vo.Pod2VO(item))
	}
	// 按照创建时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTimestamp.After(result[j].CreationTimestamp)
	})
	return result, nil
}

func (s *PodService) GetById(query dto.PodDTO) (*vo.PodInfoVO, error) {
	clientSet, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	pod, err := clientSet.CoreV1().Pods(query.Namespace).Get(ctx, query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取pod失败: %v", err)
		return nil, err
	}
	yamlData, err := podToYaml(pod)
	if err != nil {
		logs.Error("序列化pod失败: %v", err)
		return nil, err
	}
	events, err := clientSet.CoreV1().Events(query.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", query.Name),
	})
	if err != nil {
		logs.Error("获取pod事件失败: %v", err)
		return nil, err
	}
	result := vo.Pod2InfoVO(pod, events)
	result.Yaml = yamlData
	return &result, nil
}

// podToYaml 生成精简后的 Pod YAML(去掉 managedFields、status 等运行时字段)
func podToYaml(pod *v1.Pod) (string, error) {
	podCopy := pod.DeepCopy()
	podCopy.ManagedFields = nil
	podCopy.ResourceVersion = ""
	podCopy.CreationTimestamp = metav1.Time{}
	podCopy.Status = v1.PodStatus{}
	data, err := yaml.Marshal(podCopy)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *PodService) Expel(query dto.PodDTO) error {
	clientSet, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return err
	}
	// 构建 Eviction 对象
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      query.Name,
			Namespace: query.Namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	return clientSet.PolicyV1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
}

// PodLog 获取 Pod 日志流. ctx 应传入请求的 Context, 客户端断开时日志流随之关闭
func (s *PodService) PodLog(ctx context.Context, query dto.PodLogDTO) (io.ReadCloser, error) {
	clientSet, err := ClusterMap.GetClientSet(query.ClusterId)
	if err != nil {
		return nil, err
	}
	options, err := s.buildLogOption(ctx, clientSet, query)
	if err != nil {
		logs.Error("构建日志选项失败: %s %v", query.ClusterId, err)
		return nil, err
	}
	// Stream 内部会先检查请求构建阶段的错误, 无需再单独判断 req.Error()
	stream, err := clientSet.CoreV1().Pods(query.Namespace).GetLogs(query.Name, options).Stream(ctx)
	if err != nil {
		logs.Error("获取日志流失败: %s %s/%s %v", query.ClusterId, query.Namespace, query.Name, err)
		return nil, fmt.Errorf("获取日志流失败: %w", err)
	}
	return stream, nil
}

// Download 以 tar 流的形式下载容器内的文件或目录
func (s *PodService) Download(ctx context.Context, query dto.PodDownloadDTO) (io.ReadCloser, error) {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	reader, err := files.NewTarPipe(ctx, cache.config, cache.clientSet.CoreV1().RESTClient(),
		query.Namespace, query.Name, query.ContainerName, query.FilePath)
	if err != nil {
		return nil, fmt.Errorf("创建tar管道失败: %w", err)
	}
	logs.Info("文件下载开始: %s/%s %s (%d 字节)", query.Namespace, query.Name, query.FilePath, reader.Size())
	return reader, nil
}

// resolveUploadTarget 解析上传的目标目录和文件名.
// filePath 以上传文件名结尾时视为完整路径(如 /app/config.json), 否则视为目标目录(如 /app)
func resolveUploadTarget(filePath, uploadName string) (string, string, error) {
	fileName := path.Base(strings.ReplaceAll(strings.TrimSpace(uploadName), "\\", "/"))
	if fileName == "." || fileName == "/" || fileName == ".." || strings.ContainsAny(fileName, "\x00\r\n") {
		return "", "", errors.New("非法的上传文件名")
	}
	target, err := files.SanitizePath(filePath)
	if err != nil {
		return "", "", err
	}
	if path.Base(target) == fileName {
		return path.Dir(target), fileName, nil
	}
	return target, fileName, nil
}

// writeTarEntry 把单个文件以流的方式写入 tar, 不会把文件整体读入内存
func writeTarEntry(tw *tar.Writer, name string, size int64, src io.Reader) error {
	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Mode:     0644,
		Size:     size,
		ModTime:  time.Now(),
	}); err != nil {
		return fmt.Errorf("写入tar header失败: %w", err)
	}
	written, err := io.Copy(tw, src)
	if err != nil {
		return fmt.Errorf("写入文件内容失败: %w", err)
	}
	if written != size {
		return fmt.Errorf("文件大小不一致: 期望 %d 字节, 实际 %d 字节", size, written)
	}
	return nil
}

// UploadWithTar 以 tar 流的方式上传文件到容器.
// 文件内容边读边传, 只占用少量缓冲区, 大文件上传不会撑爆内存
func (s *PodService) UploadWithTar(ctx context.Context, query dto.PodDownloadDTO, file multipart.File, header *multipart.FileHeader) error {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	targetDir, fileName, err := resolveUploadTarget(query.FilePath, header.Filename)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	// exec 提前失败时关闭读端, 写入 goroutine 会因 ErrClosedPipe 退出, 不会泄漏
	defer reader.Close()
	go func() {
		tarWriter := tar.NewWriter(writer)
		err := writeTarEntry(tarWriter, fileName, header.Size, file)
		if err == nil {
			err = tarWriter.Close()
		}
		_ = writer.CloseWithError(err)
	}()

	// tar 参数以独立的 argv 传递, 不经过 shell
	req := cache.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(query.Name).
		Namespace(query.Namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: query.ContainerName,
			Command:   []string{"tar", "-xf", "-", "-C", targetDir},
			Stdin:     true,
			Stdout:    false,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cache.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建执行器失败: %w", err)
	}

	var stderrBuf bytes.Buffer
	if err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  reader,
		Stderr: &stderrBuf,
	}); err != nil {
		logs.Error("文件上传失败: %v, stderr: %s", err, stderrBuf.String())
		return fmt.Errorf("文件上传失败: %v %s", err, strings.TrimSpace(stderrBuf.String()))
	}
	if stderrBuf.Len() > 0 {
		logs.Warn("上传过程中的警告信息: %s", stderrBuf.String())
	}
	logs.Info("文件上传成功: %s/%s %s", query.Namespace, query.Name, path.Join(targetDir, fileName))
	return nil
}

func (s *PodService) buildLogOption(ctx context.Context, clientSet *kubernetes.Clientset, query dto.PodLogDTO) (*v1.PodLogOptions, error) {
	containerName := query.ContainerName
	if containerName == "" {
		pod, err := clientSet.CoreV1().Pods(query.Namespace).Get(ctx, query.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取Pod详情失败, 无法确认容器名称: %w", err)
		}
		if len(pod.Spec.Containers) == 0 {
			return nil, errors.New("pod 中没有定义容器")
		}
		containerName = pod.Spec.Containers[0].Name
	}
	options := &v1.PodLogOptions{
		Follow:     query.Follow,
		Timestamps: true,
		Previous:   false,
		Container:  containerName,
	}
	// 时间和条数二选一
	if query.SinceTime != nil {
		t := metav1.NewTime(query.SinceTime.ToTime())
		options.SinceTime = &t
	} else {
		tailLines := query.TailLines
		if tailLines <= 0 {
			tailLines = defaultLogTailLines
		}
		options.TailLines = &tailLines
	}
	return options, nil
}

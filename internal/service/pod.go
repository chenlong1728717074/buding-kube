package service

import (
	"archive/tar"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/files"
	"buding-kube/pkg/logs"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"mime/multipart"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"sync"
)

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
	clientSet, err := ClusterMap.Get(pod.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", pod.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	options := metav1.DeleteOptions{}
	//强制删除
	if pod.Force {
		flag := int64(0)
		options.GracePeriodSeconds = &flag
	}
	err = clientSet.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除pod失败: %s %s", pod.ClusterId, err.Error())
		return err
	}
	return nil
}

func (s *PodService) List(query dto.PodQueryDTO) ([]vo.PodVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
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
	result := make([]vo.PodVO, 0)
	for _, item := range pods.Items {
		if query.Keyword == "" || strings.Contains(item.Name, query.Keyword) {
			yamlData, err := yaml.Marshal(&item)
			if err != nil {
				logs.Error("序列化pod失败: %v", err)
			}
			result = append(result, vo.Pod2VO(item, string(yamlData)))
		}
	}
	//按照时间倒叙排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTimestamp.After(result[j].CreationTimestamp)
	})
	return result, nil
}

func (s *PodService) GetById(query dto.PodDTO) (*vo.PodInfoVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	pod, err := clientSet.CoreV1().Pods(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取命pod失败: %v", err)
		return nil, err
	}
	yamlData, err := yaml.Marshal(pod)
	if err != nil {
		logs.Error("序列化pod失败: %v", err)
		return nil, err
	}
	events, err := clientSet.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", query.Name),
	})
	if err != nil {
		logs.Error("获取pod事件: %v", err)
		return nil, err
	}
	result := vo.Pod2InfoVO(pod, events)
	result.Yaml = string(yamlData)
	return &result, nil
}

func (s *PodService) Expel(query dto.PodDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	//构建 Eviction 对象
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      query.Name,
			Namespace: query.Namespace,
		},
		// 设置退出时间
		DeleteOptions: &metav1.DeleteOptions{
			// GracePeriodSeconds: pointer.Int64(30),
		},
	}
	return clientSet.PolicyV1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
}

func (ps *PodService) PodLog(query dto.PodLogDTO) (io.ReadCloser, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	options, err := ps.buildLogOption(clientSet, query)
	req := clientSet.CoreV1().Pods(query.Namespace).GetLogs(query.Name, options)
	if req.Error() != nil {
		logs.Error("连接到pod日志失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("连接到pod日志失败")
	}
	stream, err := req.Stream(context.Background())
	if err != nil {
		logs.Error("获取日志流失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取日志流失败")
	}
	return stream, nil
}

func (s *PodService) Download(query dto.PodDownloadDTO) (io.ReadCloser, error) {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}

	// 规范化文件路径为Linux风格
	filePath := strings.ReplaceAll(query.FilePath, "\\", "/")

	// 使用tar管道进行文件下载
	reader, err := files.NewTarPipe(context.TODO(), cache.config, cache.clientSet.CoreV1().RESTClient(),
		query.Namespace, query.Name, query.ContainerName, filePath)
	if err != nil {
		return nil, fmt.Errorf("创建tar管道失败: %v", err)
	}

	logs.Info("文件下载开始: %s", filePath)
	return reader, nil
}

func (s *PodService) Upload(query dto.PodDownloadDTO, file multipart.File, header *multipart.FileHeader) error {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	// 优雅处理文件路径，兼容Windows和Linux
	targetPath := formatTargetPath(query.FilePath, header.Filename)

	// 构建 exec 请求
	req := cache.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(query.Name).
		Namespace(query.Namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: query.ContainerName,
			Command:   []string{"/bin/sh", "-c", fmt.Sprintf("cat > %s", targetPath)},
			Stdin:     true,
			Stdout:    false,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	// 创建 executor
	executor, err := remotecommand.NewSPDYExecutor(cache.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建执行器失败: %v", err)
	}

	// 准备输入输出流
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return errors.New("读取文件内容失败: " + err.Error())
	}

	stdin := bytes.NewReader(fileContent)
	var stderrBuf bytes.Buffer

	// 执行命令
	err = executor.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            nil,
		Stderr:            &stderrBuf,
		TerminalSizeQueue: nil,
	})

	if err != nil {
		logs.Error("文件上传失败: %v, stderr: %s", err, stderrBuf.String())
		return fmt.Errorf("文件上传失败: %v", err)
	}

	// 检查 stderr 是否有错误信息
	if stderrBuf.Len() > 0 {
		logs.Warn("上传过程中的警告信息: %s", stderrBuf.String())
	}

	logs.Info("文件上传成功: %s", targetPath)
	return nil
}

// formatTargetPath 处理目标路径和文件名，确保路径格式正确
// 如果传入的路径已经包含文件名，则直接使用；否则将文件名添加到路径后面
func formatTargetPath(basePath string, filename string) string {
	// 先规范化路径分隔符为Linux风格 (因为Pod中是Linux环境)
	basePath = strings.ReplaceAll(basePath, "\\", "/")

	// 去除路径末尾的斜杠，便于后续处理
	basePath = strings.TrimRight(basePath, "/")

	// 检查路径中是否已包含文件名
	if strings.HasSuffix(basePath, "/"+filename) || basePath == filename {
		return basePath
	}

	// 如果路径为空，直接返回文件名
	if basePath == "" {
		return filename
	}

	// 否则合并路径和文件名
	return basePath + "/" + filename
}

// UploadWithTar 使用 tar 方式上传（推荐，更稳定）
func (s *PodService) UploadWithTar(query dto.PodDownloadDTO, file multipart.File, header *multipart.FileHeader) error {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}

	// 优雅处理文件路径，兼容Windows和Linux
	targetPath := formatTargetPath(query.FilePath, header.Filename)

	// 从目标路径中提取目录和文件名
	targetPath = strings.ReplaceAll(targetPath, "\\", "/")
	lastSlash := strings.LastIndex(targetPath, "/")
	var targetDir, fileName string
	if lastSlash == -1 {
		// 没有路径分隔符，文件在根目录
		targetDir = "/"
		fileName = targetPath
	} else {
		targetDir = targetPath[:lastSlash]
		if targetDir == "" {
			targetDir = "/"
		}
		fileName = targetPath[lastSlash+1:]
	}

	// 创建 tar 流
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		tarWriter := tar.NewWriter(writer)
		defer tarWriter.Close()

		// 读取文件内容
		fileContent, err := io.ReadAll(file)
		if err != nil {
			logs.Error("读取文件失败: %v", err)
			writer.CloseWithError(err)
			return
		}

		// 写入 tar header
		if err := tarWriter.WriteHeader(&tar.Header{
			Name: fileName,
			Mode: 0644,
			Size: int64(len(fileContent)),
		}); err != nil {
			logs.Error("写入tar header失败: %v", err)
			writer.CloseWithError(err)
			return
		}

		// 写入文件内容
		if _, err := tarWriter.Write(fileContent); err != nil {
			logs.Error("写入文件内容失败: %v", err)
			writer.CloseWithError(err)
			return
		}
	}()

	// 构建 exec 请求（使用 tar 命令）
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

	// 创建 executor
	executor, err := remotecommand.NewSPDYExecutor(cache.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建执行器失败: %v", err)
	}

	var stderrBuf bytes.Buffer

	// 执行命令
	err = executor.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:             reader,
		Stdout:            nil,
		Stderr:            &stderrBuf,
		TerminalSizeQueue: nil,
	})

	if err != nil {
		logs.Error("文件上传失败: %v, stderr: %s", err, stderrBuf.String())
		return fmt.Errorf("文件上传失败: %v", err)
	}

	// 检查 stderr 是否有错误信息
	if stderrBuf.Len() > 0 {
		logs.Warn("上传过程中的警告信息: %s", stderrBuf.String())
	}

	logs.Info("文件上传成功: %s", targetPath)
	return nil
}

func (s *PodService) buildLogOption(clientSet *kubernetes.Clientset, query dto.PodLogDTO) (*v1.PodLogOptions, error) {
	containerName := query.ContainerName
	if containerName == "" {
		pod, err := clientSet.CoreV1().Pods(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
		if err != nil {
			logs.Error("获取Pod详情失败，无法确认容器名称: %s", err.Error())
			return nil, errors.New("获取Pod详情失败")
		}
		containers := pod.Spec.Containers
		if len(containers) == 0 {
			return nil, errors.New("pod 中没有定义容器")
		}
		containerName = containers[0].Name
	}
	options := v1.PodLogOptions{
		Follow:     query.Follow,
		Timestamps: true,
		Previous:   false,
		Container:  containerName,
	}
	if query.SinceTime != nil {
		time := metav1.NewTime(query.SinceTime.ToTime())
		options.SinceTime = &time
	}
	if options.SinceTime == nil && query.TailLines == 0 {
		var defaultLines int64 = 1000
		options.TailLines = &defaultLines
	}
	return &options, nil
}

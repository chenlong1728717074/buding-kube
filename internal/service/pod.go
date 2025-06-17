package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (*PodService) PodLog(query dto.PodLogDTO) (io.ReadCloser, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	options := v1.PodLogOptions{
		Follow:     query.Follow,
		Timestamps: true,
		Previous:   false,
	}
	if query.SinceTime != nil {
		time := metav1.NewTime(query.SinceTime.ToTime())
		options.SinceTime = &time
	}

	req := clientSet.CoreV1().Pods(query.Namespace).GetLogs(query.Name, &options)
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

func (s *PodService) PodDownloadDTO(query dto.PodDownloadDTO) (interface{}, error) {
	return nil, nil
}

func (s *PodService) Upload(query dto.PodDownloadDTO, file multipart.File, header *multipart.FileHeader) error {
	cache, err := ClusterMap.GetCache(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	// 构建 exec 请求
	//req := cache.clientSet.CoreV1().RESTClient().Post().
	//	Resource("pods").
	//	Name(query.Name).
	//	Namespace(query.Namespace).
	//	SubResource("exec").
	//	VersionedParams(&v1.PodExecOptions{
	//		Container: query.ContainerName,
	//		Command:   []string{"/bin/sh", "-c", fmt.Sprintf("cat > %s", query.FilePath)},
	//		Stdin:     true,
	//		Stdout:    true,
	//		Stderr:    true,
	//		TTY:       false,
	//	}, scheme.ParameterCodec)

	req := cache.clientSet.CoreV1().RESTClient().Post().
		AbsPath(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/exec", query.Namespace, query.Name)).
		VersionedParams(&v1.PodExecOptions{
			Container: query.ContainerName,
			Command:   []string{"/bin/sh", "-c", fmt.Sprintf("cat > %s", query.FilePath)},
			Stdin:     true,
			Stdout:    true,
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
		return errors.New("读取文件内容失败" + err.Error())
	}
	stdin := bytes.NewReader(fileContent)
	var stdout, stderr io.Writer

	// 执行命令
	err = executor.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})

	if err != nil {
		logs.Error("文件上传失败: %v", err)
		return fmt.Errorf("文件上传失败: %v", err)
	}

	logs.Info("文件上传成功: %s", query.FilePath)
	return nil
}

package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

var (
	deploymentSrv  *DeploymentService
	deploymentOnce sync.Once
)

type DeploymentService struct {
}

func NewDeploymentService() *DeploymentService {
	return &DeploymentService{}
}

func GetSingletonDeploymentService() *DeploymentService {
	deploymentOnce.Do(func() {
		deploymentSrv = NewDeploymentService()
	})
	return deploymentSrv
}

func (s *DeploymentService) List(query dto.WorkloadQueryDTO) ([]vo.WorkloadVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	listOptions := metav1.ListOptions{}
	// 支持空命名空间查询所有命名空间
	namespace := query.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	dyts, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), listOptions)
	if err != nil {
		logs.Error("获取Deployments失败: %v", err)
		return nil, err
	}
	result := make([]vo.WorkloadVO, 0)
	for _, item := range dyts.Items {
		if query.Name == "" || strings.Contains(item.Name, query.Name) {
			copy := item.DeepCopy()
			copy.ObjectMeta.ManagedFields = nil
			copy.ObjectMeta.ResourceVersion = ""
			copy.ObjectMeta.CreationTimestamp = metav1.Time{}
			copy.Status = v1.DeploymentStatus{}
			yamlData, err := yaml.Marshal(copy)
			if err != nil {
				logs.Error("序列化Deployment失败: %v", err)
			}
			dv := vo.Deployment2WorkloadVO(item)
			dv.Yaml = string(yamlData)
			result = append(result, dv)
		}
	}
	return result, nil
}

func (s *DeploymentService) Update(query dto.WorkloadUpdateDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	dy, err := clientSet.AppsV1().Deployments(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取Deployments失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("获取Deployments失败 %v", err)
	}
	if query.Alias != "" {
		dy.Annotations["alias"] = query.Alias
	}
	if query.Describe != "" {
		dy.Annotations["describe"] = query.Describe
	}
	_, err = clientSet.AppsV1().Deployments(query.Namespace).Update(context.TODO(), dy, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("修改Deployments失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("修改Deployments失败 %v", err)
	}
	return nil
}

func (s *DeploymentService) Delete(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	err = clientSet.AppsV1().Deployments(query.Namespace).Delete(context.TODO(), query.Name, metav1.DeleteOptions{})
	if err != nil {
		logs.Error("删除失败: %s %s", query.ClusterId, err.Error())
		return fmt.Errorf("删除失败 %v", err)
	}
	return nil
}

func (s *DeploymentService) Rollout(query dto.WorkloadBaseDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	// 获取当前的Deployment
	deployment, err := clientSet.AppsV1().Deployments(query.Namespace).Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logs.Error("获取Deployment失败: %v", err)
		return fmt.Errorf("获取Deployment失败: %v", err)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// 更新Deployment
	_, err = clientSet.AppsV1().Deployments(query.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		logs.Error("重建Deployment失败: %v", err)
		return fmt.Errorf("重建Deployment失败: %v", err)
	}
	logs.Info("成功触发Deployment [%s/%s] 重建", query.Namespace, query.Name)
	return nil

	/*
		// 使用Patch(用于大型Deployment)
		patchData := map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]string{
							"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
						},
					},
				},
			},
		}

		patchBytes, err := json.Marshal(patchData)
		if err != nil {
			logs.Error("序列化Patch数据失败: %v", err)
			return fmt.Errorf("序列化Patch数据失败: %v", err)
		}

		_, err = clientSet.AppsV1().Deployments(query.Namespace).Patch(
			context.TODO(),
			query.Name,
			types.StrategicMergePatchType,
			patchBytes,
			metav1.PatchOptions{},
		)
		if err != nil {
			logs.Error("重建Deployment失败: %v", err)
			return fmt.Errorf("重建Deployment失败: %v", err)
		}

		logs.Info("成功触发Deployment [%s/%s] 重建", query.Namespace, query.Name)
		return nil
	*/
}

func (s *DeploymentService) Apply(query dto.WorkloadApplyDTO) error {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	var dy v1.Deployment

	err = yaml.Unmarshal([]byte(query.Yaml), &dy)
	if err != nil {
		logs.Error("解析yaml失败: %v", err)
		return err
	}
	//修改必须基于namespbace
	//if query.Namespace!="" {
	//	dy.Namespace=query.Namespace
	//}
	existingDeploy, err := clientSet.AppsV1().Deployments(query.Namespace).Get(context.TODO(), dy.Name, metav1.GetOptions{})

	if err != nil {
		//不存在则创建
		if apierrors.IsNotFound(err) {
			logs.Info("Deployment不存在，开始创建: %s/%s", query.Namespace, dy.Name)
			_, err := clientSet.AppsV1().Deployments(query.Namespace).Create(context.TODO(), &dy, metav1.CreateOptions{})
			if err != nil {
				logs.Error("创建Deployments失败: %s %v", query.ClusterId, err)
				return fmt.Errorf("创建Deployments失败 %v", err)
			}
			logs.Info("创建Deployment成功: %s/%s", query.Namespace, dy.Name)
			return nil
		}
		return err
	}
	_, err = clientSet.AppsV1().Deployments(query.Namespace).Update(context.TODO(), &dy, metav1.UpdateOptions{})
	dy.ResourceVersion = existingDeploy.ResourceVersion
	if err != nil {
		logs.Error("修改Deployments失败: %s %v", query.ClusterId, err)
		return fmt.Errorf("修改Deployments失败 %v", err)
	}
	logs.Info("更新Deployment成功: %s/%s", query.Namespace, dy.Name)
	return nil
}

package kube

import (
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"fmt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

var (
	ClusterGVR = schema.GroupVersionResource{
		Group:    "kube.buding.goaigc.fun",
		Version:  "v1",
		Resource: "clusters",
	}

	UserGVR = schema.GroupVersionResource{
		Group:    "kube.buding.goaigc.fun",
		Version:  "v1",
		Resource: "users",
	}
)

// buildClusterCRD 构建集群 CRD 定义
// 用于存储和管理 Kubernetes 集群的元数据信息
func buildClusterCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		// ObjectMeta 定义 CRD 的元数据
		ObjectMeta: metav1.ObjectMeta{
			Name: consts.ClusterCrd,
		},
		// Spec 定义
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: consts.GroupCrd,
			// CRD 的名称定义(单数、复数、kind 等)
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:     "clusters",
				Singular:   "cluster",
				Kind:       consts.ClusterKind,
				ShortNames: []string{"cls"},
			},
			Scope: apiextensionsv1.ClusterScoped,
			// 定义支持的 API 版本列表
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					// 版本号
					Name:    consts.ClusterCrdVersion,
					Served:  true,
					Storage: true,
					// Schema
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							// 根对象类型
							Type: "object",
							// 定义对象的属性
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								// spec 字段定义集群的期望状态
								"spec": {
									Type: "object",
									// spec 中的必需字段
									Required: []string{"apiServer"},
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										// 集群的 API Server 地址
										"apiServer": {
											Type:        "string",
											Description: "Kubernetes API Server URL",
											Pattern:     `^https?://.*`,
										},
										// 区域
										"region": {
											Type:        "string",
											Description: "Cluster region/zone",
										},
										// 别名
										"alias": {
											Type:        "string",
											Description: "Cluster alias",
										},
										// 集群描述信息
										"description": {
											Type:        "string",
											Description: "Cluster description",
										},
										// 认证信息
										"auth": {
											Type: "object",
											Properties: map[string]apiextensionsv1.JSONSchemaProps{
												// 认证类型: token, certificate, kubeconfig
												"type": {
													Type: "string",
													Enum: []apiextensionsv1.JSON{
														{Raw: []byte(`"token"`)},
														{Raw: []byte(`"certificate"`)},
														{Raw: []byte(`"kubeconfig"`)},
													},
												},
												// Secret 引用,存储敏感认证信息
												"secretRef": {
													Type: "object",
													Properties: map[string]apiextensionsv1.JSONSchemaProps{
														"name": {
															Type:        "string",
															Description: "Secret name containing auth credentials",
														},
														"namespace": {
															Type:        "string",
															Description: "Secret namespace",
														},
													},
												},
											},
										},
										// 集群标签,用于分类和筛选
										"labels": {
											Type: "object",
											// 允许任意键值对
											AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "string",
												},
											},
										},
									},
								},
								// status 字段定义集群的当前状态
								"status": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										// 集群健康状态
										"health": {
											Type: "string",
											Enum: []apiextensionsv1.JSON{
												{Raw: []byte(`"healthy"`)},
												{Raw: []byte(`"unhealthy"`)},
												{Raw: []byte(`"unknown"`)},
											},
										},
										// 最后检查时间
										"lastChecked": {
											Type:   "string",
											Format: "date-time",
										},
										// 集群节点数量
										"nodeCount": {
											Type: "integer",
										},
										// 状态消息
										"message": {
											Type: "string",
										},
									},
								},
							},
						},
					},
					// 定义额外的打印列,用于 kubectl get 命令的输出
					AdditionalPrinterColumns: []apiextensionsv1.CustomResourceColumnDefinition{
						{
							Name:        "API-Server",
							Type:        "string",
							Description: "Kubernetes API Server",
							// JSONPath 表达式,指定从资源中提取的字段
							JSONPath: ".spec.apiServer",
						},
						{
							Name:     "Region",
							Type:     "string",
							JSONPath: ".spec.region",
						},
						{
							Name:     "Health",
							Type:     "string",
							JSONPath: ".status.health",
						},
						{
							Name:     "Age",
							Type:     "date",
							JSONPath: ".metadata.creationTimestamp",
						},
					},
				},
			},
		},
	}
}

// buildUserCRD 构建用户 CRD 定义
// 用于存储和管理访问 Kubernetes 集群的用户信息
func buildUserCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		// ObjectMeta 定义 CRD 的元数据
		ObjectMeta: metav1.ObjectMeta{
			// CRD 的名称必须符合 <plural>.<group> 的格式
			Name: consts.UserCrd,
		},
		// Spec 定义 CRD 的规范
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			// API 组名
			Group: consts.GroupCrd,
			// CRD 的名称定义
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				// 复数名称
				Plural: "users",
				// 单数名称
				Singular: "user",
				// Kind 名称
				Kind: consts.UserKind,
				// 短名称
				ShortNames: []string{"usr"},
			},
			Scope: apiextensionsv1.ClusterScoped,
			// 定义支持的 API 版本
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					// 版本号
					Name: consts.ClusterCrdVersion,
					// 通过 API 服务
					Served: true,
					// 存储版本
					Storage: true,
					// Schema 定义资源的验证规则
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							// 根对象类型
							Type: "object",
							// 定义对象的属性
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								// spec 字段定义用户的期望状态
								"spec": {
									Type: "object",
									// spec 中的必需字段
									Required: []string{"role", "password", "enabled"},
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										// 用户邮箱(唯一标识)
										"email": {
											Type:        "string",
											Description: "User email address",
											// 邮箱格式验证
											//Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
										},
										"realName": {
											Type:        "string",
											Description: "User real name",
										},
										"password": {
											Type:        "string",
											Description: "User password",
										},
										// 用户角色列表
										"role": {
											Type:        "string",
											Description: "User role for RBAC",
											Enum: []apiextensionsv1.JSON{
												{Raw: []byte(`"super"`)},
												{Raw: []byte(`"admin"`)},
												{Raw: []byte(`"normal"`)},
											},
										},
										// 用户是否启用
										"enabled": {
											Type:        "boolean",
											Description: "Whether the user is enabled",
											Default:     &apiextensionsv1.JSON{Raw: []byte(`true`)},
										},
										// 用户元数据(自定义键值对)
										"attributes": {
											Type: "object",
											// 允许任意键值对
											AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "string",
												},
											},
										},
									},
								},
								// status 字段定义用户的当前状态
								"status": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										// 用户状态
										"state": {
											Type: "string",
											Enum: []apiextensionsv1.JSON{
												{Raw: []byte(`"active"`)},
												{Raw: []byte(`"inactive"`)},
												{Raw: []byte(`"suspended"`)},
												{Raw: []byte(`"expired"`)},
											},
										},
										// 最后登录时间
										"lastLogin": {
											Type:   "string",
											Format: "date-time",
										},
									},
								},
							},
						},
					},
					// 定义额外的打印列
					AdditionalPrinterColumns: []apiextensionsv1.CustomResourceColumnDefinition{
						{
							Name:        "Email",
							Type:        "string",
							Description: "User email",
							JSONPath:    ".spec.email",
						},
						{
							Name:     "Enabled",
							Type:     "boolean",
							JSONPath: ".spec.enabled",
						},
						{
							Name:     "State",
							Type:     "string",
							JSONPath: ".status.state",
						},
						{
							Name:     "Age",
							Type:     "date",
							JSONPath: ".metadata.creationTimestamp",
						},
					},
				},
			},
		},
	}
}

// EnsureCRDs 确保 CRD 定义存在
func EnsureCRDs(crds []*apiextensionsv1.CustomResourceDefinition, client *apiextensionsclientset.Clientset) error {
	for _, crd := range crds {
		if err := createOrUpdateCRD(crd, client); err != nil {
			return fmt.Errorf("创建/更新 CRD %s 失败: %w", crd.Name, err)
		}

		// 等待 CRD 就绪
		if err := waitForCRDReady(crd.Name, client); err != nil {
			return fmt.Errorf("等待 CRD %s 就绪失败: %w", crd.Name, err)
		}

		logs.Info("✅ CRD %s 已就绪", crd.Name)
	}

	return nil
}

func createOrUpdateCRD(crd *apiextensionsv1.CustomResourceDefinition, client *apiextensionsclientset.Clientset) error {
	ctx := context.Background()
	existing, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// 创建新的 CRD
		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	} // 更新现有 CRD (保留 ResourceVersion)
	crd.ResourceVersion = existing.ResourceVersion
	_, err = client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
	return err
}

func waitForCRDReady(crdName string, client *apiextensionsclientset.Clientset) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return wait.PollUntilContextTimeout(
		ctx,
		1*time.Second,  // 轮询间隔
		30*time.Second, // 超时时间
		true,           // 立即执行第一次检查
		func(ctx context.Context) (done bool, err error) {
			crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(
				ctx, crdName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			for _, cond := range crd.Status.Conditions {
				if cond.Type == apiextensionsv1.Established &&
					cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			}
			return false, nil
		},
	)
}

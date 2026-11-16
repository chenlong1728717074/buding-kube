package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/logs"
	"context"
	"sort"
	"strings"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
)

type CrdApi struct {
	router *gin.RouterGroup
	BaseApi
}

func NewCrdApi(router *gin.RouterGroup) *CrdApi {
	api := CrdApi{
		router: router,
	}
	api.Router()
	return &api
}

func (api *CrdApi) Router() {
	api.router.GET("/list", api.List)
	api.router.GET("", middleware.Blocker(), api.Detail)
	api.router.DELETE("", middleware.Blocker(), api.Delete)
}

// getCrdClient 通过缓存的集群配置构造 apiextensions 客户端
// kubernetes.Clientset 不包含 CRD 操作接口, 需使用独立的 apiextensionsclientset
func (api *CrdApi) getCrdClient(clusterId string) (*apiextensionsclientset.Clientset, error) {
	restConfig, err := service.ClusterMap.GetConfig(clusterId)
	if err != nil {
		return nil, err
	}
	return apiextensionsclientset.NewForConfig(restConfig)
}

// List 查询当前集群的 CRD 列表
func (api *CrdApi) List(ctx *gin.Context) {
	var query dto.CrdPageQueryDTO
	if err := api.BindQuery(ctx, &query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	crdClient, err := api.getCrdClient(query.ClusterId)
	if err != nil {
		api.InternalError(ctx, "获取集群客户端失败:", err)
		return
	}
	crdList, err := crdClient.ApiextensionsV1().CustomResourceDefinitions().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logs.Error("查询CRD列表失败:%v", err)
		api.InternalError(ctx, "查询失败:", err)
		return
	}

	keyword := strings.TrimSpace(query.Keyword)
	result := make([]vo.CrdVO, 0, len(crdList.Items))
	for _, item := range crdList.Items {
		storageVersion := ""
		for _, ver := range item.Spec.Versions {
			if ver.Storage {
				storageVersion = ver.Name
				break
			}
		}
		cvo := vo.CrdVO{
			Name:       item.Name,
			Kind:       item.Spec.Names.Kind,
			Group:      item.Spec.Group,
			Version:    storageVersion,
			Scope:      string(item.Spec.Scope),
			CreateTime: item.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}
		if keyword != "" &&
			!strings.Contains(cvo.Name, keyword) &&
			!strings.Contains(cvo.Kind, keyword) &&
			!strings.Contains(cvo.Group, keyword) {
			continue
		}
		result = append(result, cvo)
	}
	// 按名称排序保证顺序稳定
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })

	response := BuildPageResponse(result, query.Page, query.PageSize)
	api.SuccessWithData(ctx, response)
}

// Detail 查询单个 CRD 原始对象(前端渲染 YAML)
func (api *CrdApi) Detail(ctx *gin.Context) {
	var base dto.CrdBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	crdClient, err := api.getCrdClient(base.ClusterId)
	if err != nil {
		api.InternalError(ctx, "获取集群客户端失败:", err)
		return
	}
	item, err := crdClient.ApiextensionsV1().CustomResourceDefinitions().
		Get(context.Background(), base.Name, metav1.GetOptions{})
	if err != nil {
		api.InternalError(ctx, "获取失败:%v", err)
		return
	}
	api.SuccessWithData(ctx, item)
}

// Delete 删除 CRD(连带其下所有自定义资源, 危险操作)
func (api *CrdApi) Delete(ctx *gin.Context) {
	var base dto.CrdBaseDTO
	if err := api.BindQuery(ctx, &base); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	crdClient, err := api.getCrdClient(base.ClusterId)
	if err != nil {
		api.InternalError(ctx, "获取集群客户端失败:", err)
		return
	}
	deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = crdClient.ApiextensionsV1().CustomResourceDefinitions().
		Delete(deleteCtx, base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除CRD失败:%s %v", base.Name, err)
		api.InternalError(ctx, "删除失败:%v", err)
		return
	}
	api.SuccessMsg(ctx, "删除成功")
}

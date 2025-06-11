package api

import (
	"buding-kube/internal/service"
	"github.com/gin-gonic/gin"
)

type NamespacesApi struct {
	router *gin.RouterGroup
	srv    *service.NamespaceService
	BaseApi
}

func NewNamespacesApi(router *gin.RouterGroup) *NamespacesApi {
	api := NamespacesApi{
		router: router,
		srv:    service.GetSingletonNamespaceService(),
	}
	api.Router()
	return &api
}

func (api *NamespacesApi) Router() {
	api.router.GET("/:id", api.Info)
	api.router.DELETE("/:id", api.Delete)
	api.router.GET("/list", api.List)
	api.router.POST("", api.Add)
}

func (*NamespacesApi) List(ctx *gin.Context) {

}

func (api *NamespacesApi) Info(context *gin.Context) {

}

func (api *NamespacesApi) Delete(context *gin.Context) {

}

func (api *NamespacesApi) Add(context *gin.Context) {

}

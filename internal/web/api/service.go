package api

import (
	"buding-kube/internal/service"
	"github.com/gin-gonic/gin"
)

type KubeSrvApi struct {
	router *gin.RouterGroup
	srv    *service.KubeSrvService
	BaseApi
}

func NewKubeSrvApi(router *gin.RouterGroup) *KubeSrvApi {
	api := KubeSrvApi{
		router: router,
		srv:    service.GetSingletonKubeSrvService(),
	}
	api.Router()
	return &api
}

func (api *KubeSrvApi) Router() {
	api.router.GET("/list", api.List)
}

func (api *KubeSrvApi) List(context *gin.Context) {

}

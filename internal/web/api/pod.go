package api

import (
	"buding-kube/internal/service"
	"github.com/gin-gonic/gin"
)

type PodApi struct {
	router *gin.RouterGroup
	srv    *service.PodService
	BaseApi
}

func NewPodApi(router *gin.RouterGroup) *PodApi {
	api := PodApi{
		router: router,
		srv:    service.GetSingletonPodService(),
	}
	api.Router()
	return &api
}

func (api *PodApi) Router() {

}

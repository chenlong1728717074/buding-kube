package api

import (
	"buding-kube/internal/service"
	"github.com/gin-gonic/gin"
)

type ApplyApi struct {
	router *gin.RouterGroup
	srv    *service.ApplyService
	BaseApi
}

func NewApplyApi(router *gin.RouterGroup) *ApplyApi {
	api := ApplyApi{
		router: router,
		srv:    service.GetSingletonApplyService(),
	}
	api.Router()
	return &api
}

func (a *ApplyApi) Router() {

}

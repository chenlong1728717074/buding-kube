package api

import (
	"buding-kube/internal/service"
	"github.com/gin-gonic/gin"
)

type ToolApi struct {
	router *gin.RouterGroup
	srv    *service.ToolService
	BaseApi
}

func NewToolApi(router *gin.RouterGroup) *ToolApi {
	api := ToolApi{
		router: router,
		srv:    service.GetSingletonToolService(),
	}
	api.Router()
	return &api
}

func (a *ToolApi) Router() {

}

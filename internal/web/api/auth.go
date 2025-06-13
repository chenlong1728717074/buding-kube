package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type AuthApi struct {
	router *gin.RouterGroup
	srv    *service.AuthService
	BaseApi
}

func NewAuthApi(router *gin.RouterGroup) *AuthApi {
	api := AuthApi{
		router: router,
		srv:    service.GetSingletonAuthService(),
	}
	api.Router()
	return &api
}

func (api *AuthApi) Router() {
	api.router.POST("/login", api.login)
}

func (api *AuthApi) login(ctx *gin.Context) {
	var login dto.LoginDTO
	if err := ctx.ShouldBindJSON(&login); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.Login(login)
	if err != nil {
		api.InternalError(ctx, "登录失败:", err)
		return
	}
	api.SuccessWithData(ctx, result)
}

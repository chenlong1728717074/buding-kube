package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

// AuthApi 认证API
type AuthApi struct {
	router *gin.RouterGroup
	srv    *service.AuthService
	BaseApi
}

// NewAuthApi 创建认证API控制器
func NewAuthApi(router *gin.RouterGroup) *AuthApi {
	api := AuthApi{
		router: router,
		srv:    service.GetSingletonAuthService(),
	}
	api.Router()
	return &api
}

// Router 配置路由
func (api *AuthApi) Router() {
	api.router.POST("/login", api.Login)
	api.router.POST("/logout", api.Logout)
}

// @Summary 用户登录
// @Description 用户登录，返回用户信息和Token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param login body dto.LoginDTO true "登录参数，包含用户名（username）和密码（password）"
// @Success 200 {object} vo.Response{data=vo.UserVO} "登录成功"
// @Failure 400 {object} vo.Response "参数绑定错误"
// @Failure 500 {object} vo.Response "用户名或密码错误"
// @Failure 500 {object} vo.Response "用户已禁用"
// @Failure 500 {object} vo.Response "登录失败"
// @Router /api/auth/login [post]
func (api *AuthApi) Login(ctx *gin.Context) {
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

func (api *AuthApi) Logout(ctx *gin.Context) {
	api.SuccessMsg(ctx, "登出成功")
}

package api

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"
	"buding-kube/internal/web/vo"

	"github.com/gin-gonic/gin"
)

// UserApi 用户管理API
type UserApi struct {
	router *gin.RouterGroup
	srv    *service.UserService
	BaseApi
}

// NewUserApi 创建用户API控制器
func NewUserApi(router *gin.RouterGroup) *UserApi {
	api := UserApi{
		router: router,
		srv:    service.GetSingletonUserService(),
	}
	api.Router()
	return &api
}

// Router 配置路由
func (api *UserApi) Router() {
	api.router.POST("", BindJSON[dto.CreateUserDTO](api.CreateUser))
	api.router.PUT("", middleware.Blocker(), BindJSON[dto.CreateUserDTO](api.UpdateUser))
	api.router.GET("/list", BindQuery[dto.UserQueryDTO](api.ListUsers))
	api.router.GET("/:name", api.GetUser)
	api.router.DELETE("/:name", middleware.Blocker(), api.DeleteUser)
}

func (api *UserApi) ListUsers(ctx *gin.Context, query dto.UserQueryDTO) {
	result, err := api.srv.ListUsers(query)
	if err != nil {
		api.InternalError(ctx, "获取用户列表失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(result, query.Page, query.PageSize))
}

func (api *UserApi) GetUser(ctx *gin.Context) {
	name := api.GetParam(ctx, "name")
	// 获取当前用户
	var currentUser *kube.LoginUser
	var err error
	if currentUser, err = api.CurrentUser(ctx); err != nil {
		api.Fail(ctx, vo.CodeInternalError, err.Error())
	}
	user, err := api.srv.GetUser(name)
	if err != nil {
		api.NotFound(ctx, "用户不存在")
		return
	}
	// 检查权限
	if !currentUser.CanManageUser(user) {
		api.Forbidden(ctx, "权限不足")
		return
	}

	api.SuccessWithData(ctx, user)
}

func (api *UserApi) CreateUser(ctx *gin.Context, req dto.CreateUserDTO) {
	if err := api.srv.CreateUser(req); err != nil {
		api.InternalError(ctx, "创建用户失败", err)
		return
	}
	api.SuccessMsg(ctx, "创建用户成功")
}

func (api *UserApi) UpdateUser(ctx *gin.Context, req dto.CreateUserDTO) {

	api.SuccessMsg(ctx, "更新用户成功")
}

// DeleteUser 删除用户
func (api *UserApi) DeleteUser(ctx *gin.Context) {
	name := api.GetParam(ctx, "name")

	// 获取当前用户
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return
	}

	err = api.srv.DeleteUser(name, currentUser)
	if err != nil {
		api.InternalError(ctx, "删除用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "删除用户成功")
}

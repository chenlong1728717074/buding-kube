package api

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/middleware"

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
	api.router.PUT("/status", middleware.Blocker(), BindJSON[dto.UpdateUserStatusDTO](api.UpdateUserStatus))
	api.router.PUT("/enable", middleware.Blocker(), BindJSON[dto.UpdateUserEnableDTO](api.UpdateUserEnable))
	api.router.POST("/batchDelete", middleware.Blocker(), BindJSON[dto.BatchDeleteUsersDTO](api.BatchDeleteUsers))
	api.router.POST("/resetPassword/:name", middleware.Blocker(), BindStringParam("name", api.ResetPassword))
	api.router.GET("/list", BindQuery[dto.UserQueryDTO](api.ListUsers))
	api.router.GET("/:name", BindStringParam("name", api.GetUser))
	api.router.DELETE("/:name", middleware.Blocker(), BindStringParam("name", api.DeleteUser))
}

func (api *UserApi) ListUsers(ctx *gin.Context, query dto.UserQueryDTO) {
	result, err := api.srv.ListUsers(query)
	if err != nil {
		api.InternalError(ctx, "获取用户列表失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(result, query.Page, query.PageSize))
}

func (api *UserApi) GetUser(ctx *gin.Context, name string) {
	// 获取当前用户
	var currentUser *kube.LoginUser
	var err error
	currentUser = api.CurrentUser(ctx)
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
	if err := api.srv.CreateUser(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "创建用户失败", err)
		return
	}
	api.SuccessMsg(ctx, "创建用户成功")
}

func (api *UserApi) UpdateUser(ctx *gin.Context, req dto.CreateUserDTO) {
	if err := api.srv.UpdateUser(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "修改用户失败", err)
		return
	}
	api.SuccessMsg(ctx, "更新用户成功")
}

func (api *UserApi) UpdateUserStatus(ctx *gin.Context, req dto.UpdateUserStatusDTO) {
	if err := api.srv.UpdateUserStatus(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "更新用户状态失败", err)
		return
	}
	api.SuccessMsg(ctx, "更新用户状态成功")
}

func (api *UserApi) UpdateUserEnable(ctx *gin.Context, req dto.UpdateUserEnableDTO) {
	if err := api.srv.UpdateUserEnable(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "更新用户启用状态失败", err)
		return
	}
	api.SuccessMsg(ctx, "更新用户启用状态成功")
}

func (api *UserApi) BatchDeleteUsers(ctx *gin.Context, req dto.BatchDeleteUsersDTO) {
	if err := api.srv.BatchDeleteUsers(req.Usernames, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "批量删除用户失败", err)
		return
	}
	api.SuccessMsg(ctx, "批量删除用户成功")
}

func (api *UserApi) ResetPassword(ctx *gin.Context, name string) {
	if err := api.srv.ResetPassword(name, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "重置用户密码失败", err)
		return
	}
	api.SuccessMsg(ctx, "重置用户密码成功")
}

// DeleteUser 删除用户
func (api *UserApi) DeleteUser(ctx *gin.Context, name string) {
	// 获取当前用户
	var currentUser *kube.LoginUser
	var err error

	currentUser = api.CurrentUser(ctx)
	err = api.srv.DeleteUser(name, currentUser)
	if err != nil {
		api.InternalError(ctx, "删除用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "删除用户成功")
}

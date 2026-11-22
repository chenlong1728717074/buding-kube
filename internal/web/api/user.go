package api

import (
	"errors"

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
	api.router.POST("", middleware.Blocker(), BindJSON[dto.CreateUserDTO](api.CreateUser))
	api.router.PUT("", middleware.Blocker(), BindJSON[dto.UpdateUserDTO](api.UpdateUser))
	api.router.GET("/profile", api.GetProfile)
	api.router.PUT("/profile", middleware.Blocker(), BindJSON[dto.ProfileUpdateDTO](api.UpdateProfile))
	api.router.PUT("/status", middleware.Blocker(), BindJSON[dto.UpdateUserStatusDTO](api.UpdateUserStatus))
	api.router.PUT("/enable", middleware.Blocker(), BindJSON[dto.UpdateUserEnableDTO](api.UpdateUserEnable))
	api.router.POST("/batchDelete", middleware.Blocker(), BindJSON[dto.BatchDeleteUsersDTO](api.BatchDeleteUsers))
	api.router.POST("/resetPassword/:name", middleware.Blocker(), BindStringParam("name", api.ResetPassword))
	api.router.GET("/list", BindQuery[dto.UserQueryDTO](api.ListUsers))
	api.router.GET("/:name", BindStringParam("name", api.GetUser))
	api.router.DELETE("/:name", middleware.Blocker(), BindStringParam("name", api.DeleteUser))
}

// ListUsers 超级管理员查看全部用户, 管理员查看普通用户, 普通用户没有权限
func (api *UserApi) ListUsers(ctx *gin.Context, query dto.UserQueryDTO) {
	result, err := api.srv.ListUsers(query, api.CurrentUser(ctx))
	if errors.Is(err, service.ErrPermissionDenied) {
		api.Forbidden(ctx, "权限不足")
		return
	}
	if err != nil {
		api.InternalError(ctx, "获取用户列表失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(result, query.Page, query.PageSize))
}

// GetUser 用户详情; 不返回密码哈希等内部字段
func (api *UserApi) GetUser(ctx *gin.Context, name string) {
	currentUser := api.CurrentUser(ctx)
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

	api.SuccessWithData(ctx, vo.User2VO(user))
}

func (api *UserApi) CreateUser(ctx *gin.Context, req dto.CreateUserDTO) {
	if err := api.srv.CreateUser(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "创建用户失败", err)
		return
	}
	api.SuccessMsg(ctx, "创建用户成功")
}

// GetProfile 获取当前登录用户的个人信息
func (api *UserApi) GetProfile(ctx *gin.Context) {
	currentUser := api.CurrentUser(ctx)
	user, err := api.srv.GetUser(currentUser.Username)
	if err != nil {
		api.NotFound(ctx, "用户不存在")
		return
	}
	api.SuccessWithData(ctx, vo.User2VO(user))
}

// UpdateProfile 更新当前登录用户的个人信息(仅基本资料字段)
func (api *UserApi) UpdateProfile(ctx *gin.Context, req dto.ProfileUpdateDTO) {
	if err := api.srv.UpdateProfile(req, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "更新个人信息失败", err)
		return
	}
	api.SuccessMsg(ctx, "更新个人信息成功")
}

func (api *UserApi) UpdateUser(ctx *gin.Context, req dto.UpdateUserDTO) {
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
	// 密码通过可选请求体传入, 留空则使用默认密码 123456
	var body dto.ResetPasswordDTO
	_ = ctx.ShouldBindJSON(&body)
	if err := api.srv.ResetPassword(name, body.Password, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "重置用户密码失败", err)
		return
	}
	api.SuccessMsg(ctx, "重置用户密码成功")
}

// DeleteUser 删除用户
func (api *UserApi) DeleteUser(ctx *gin.Context, name string) {
	if err := api.srv.DeleteUser(name, api.CurrentUser(ctx)); err != nil {
		api.InternalError(ctx, "删除用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "删除用户成功")
}

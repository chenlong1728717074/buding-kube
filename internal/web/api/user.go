package api

import (
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"github.com/gin-gonic/gin"
)

type UserApi struct {
	router *gin.RouterGroup
	srv    *service.UserService
	BaseApi
}

func NewUserApi(router *gin.RouterGroup) *UserApi {
	api := UserApi{
		router: router,
		srv:    service.GetSingletonUserService(),
	}
	api.Router()
	return &api
}

func (api *UserApi) Router() {
	api.router.GET("", api.listUsers)
	api.router.GET("/:username", api.getUser)
	api.router.POST("", api.createUser)
	api.router.PUT("/:username", api.updateUser)
	api.router.DELETE("/:username", api.deleteUser)
}

// @Summary 获取用户列表
// @Description 根据查询条件获取用户列表，支持分页
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1"
// @Param pageSize query int false "每页数量，默认10"
// @Param username query string false "用户名，精确匹配"
// @Param role query int false "角色，1=超级管理员 2=管理员 3=普通用户"
// @Param status query int false "状态，1=正常 0=禁用"
// @Success 200 {object} vo.Response{data=vo.PageResponse{items=[]vo.UserVO}} "获取成功"
// @Failure 400 {object} vo.Response "参数绑定错误"
// @Failure 401 {object} vo.Response "未授权"
// @Failure 500 {object} vo.Response "获取失败"
// @Router /api/user [get]
func (api *UserApi) listUsers(ctx *gin.Context) {
	var query dto.UserQueryDTO
	if err := ctx.ShouldBindQuery(&query); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	result, err := api.srv.ListUsers(query)
	if err != nil {
		api.InternalError(ctx, "获取用户列表失败", err)
		return
	}
	api.SuccessWithData(ctx, BuildPageResponse(result, query.Page, query.PageSize))
}

// @Summary 获取用户信息
// @Description 根据用户名获取用户详情
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "用户名"
// @Success 200 {object} vo.Response{data=vo.UserVO} "获取成功"
// @Failure 400 {object} vo.Response "参数绑定错误"
// @Failure 401 {object} vo.Response "未授权"
// @Failure 404 {object} vo.Response "用户不存在"
// @Failure 500 {object} vo.Response "获取失败"
// @Router /api/user/{username} [get]
func (api *UserApi) getUser(ctx *gin.Context) {
	username := api.GetParam(ctx, "username")
	// 获取当前用户
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return
	}
	user, err := api.srv.GetUserByUsername(username)
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

// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.CreateUserDTO true "用户信息"
// @Success 200 {object} vo.Response "创建成功"
// @Failure 400 {object} vo.Response "参数绑定错误"
// @Failure 401 {object} vo.Response "未授权"
// @Failure 403 {object} vo.Response "权限不足"
// @Failure 500 {object} vo.Response "创建失败"
// @Router /api/user [post]
func (api *UserApi) createUser(ctx *gin.Context) {
	var req dto.CreateUserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.ParamBindError(ctx, err)
		return
	}
	// 获取当前用户
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return
	}
	err = api.srv.CreateUser(req, currentUser)
	if err != nil {
		api.InternalError(ctx, "创建用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "创建用户成功")
}

// @Summary 更新用户
// @Description 更新用户信息（不能修改用户名）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "用户名（不可修改）"
// @Param user body dto.UpdateUserDTO true "用户信息"
// @Success 200 {object} vo.Response "更新成功"
// @Failure 400 {object} vo.Response "参数绑定错误"
// @Failure 401 {object} vo.Response "未授权"
// @Failure 403 {object} vo.Response "权限不足"
// @Failure 404 {object} vo.Response "用户不存在"
// @Failure 500 {object} vo.Response "更新失败"
// @Router /api/user/{username} [put]
func (api *UserApi) updateUser(ctx *gin.Context) {
	username := api.GetParam(ctx, "username")

	var req dto.UpdateUserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		api.ParamBindError(ctx, err)
		return
	}

	// 获取当前用户
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return
	}

	// 获取目标用户，确认其存在
	targetUser, err := api.srv.GetUserByUsername(username)
	if err != nil {
		api.NotFound(ctx, "用户不存在")
		return
	}

	// 检查是否有权限修改此用户
	if !currentUser.CanManageUser(targetUser) {
		api.Forbidden(ctx, "权限不足")
		return
	}

	err = api.srv.UpdateUser(username, req, currentUser)
	if err != nil {
		api.InternalError(ctx, "更新用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "更新用户成功")
}

// @Summary 删除用户
// @Description 删除用户（不能删除自己或超级管理员）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "用户名"
// @Success 200 {object} vo.Response "删除成功"
// @Failure 401 {object} vo.Response "未授权"
// @Failure 403 {object} vo.Response "权限不足，不能删除自己或超级管理员"
// @Failure 404 {object} vo.Response "用户不存在"
// @Failure 500 {object} vo.Response "删除失败"
// @Router /api/user/{username} [delete]
func (api *UserApi) deleteUser(ctx *gin.Context) {
	username := api.GetParam(ctx, "username")

	// 获取当前用户
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return
	}

	err = api.srv.DeleteUser(username, currentUser)
	if err != nil {
		api.InternalError(ctx, "删除用户失败", err)
		return
	}

	api.SuccessMsg(ctx, "删除用户成功")
}

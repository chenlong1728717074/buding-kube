package dto

// LoginDTO 登录参数
type LoginDTO struct {
	Username string `json:"username" example:"zhangsan" binding:"required"` // 用户名，用户登录账号
	Password string `json:"password" example:"123456" binding:"required"`   // 密码，用户登录密码
}

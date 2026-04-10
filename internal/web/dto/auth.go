package dto

// LoginDTO 登录参数
type LoginDTO struct {
	Username string `json:"username" example:"zhangsan" binding:"required"` // 用户名，用户登录账号
	Password string `json:"password" example:"123456" binding:"required"`   // 密码，用户登录密码
}

// ChangePasswordDTO 修改当前登录用户密码
// 注意：这是用户主动修改密码，不是管理员重置密码
// 管理员重置密码走 user/resetPassword 接口，默认重置为 123456
type ChangePasswordDTO struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

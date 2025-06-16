package vo

import "buding-kube/internal/model"

// UserVO 用户视图对象
type UserVO struct {
	Id         string         `json:"id" example:"usada-31231a"`                               //id
	Username   string         `json:"username" example:"zhangsan"`                             // 用户名
	Email      string         `json:"email" example:"zhangsan@example.com"`                    // 邮箱
	Role       model.UserRole `json:"role" example:"1"`                                        // 角色: 1=超级管理员 2=管理员 3=普通用户
	Status     int            `json:"status" example:"1"`                                      // 状态: 1=正常 0=禁用
	Token      string         `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT令牌
	Department string         `json:"department" example:"研发部"`                                //部门
	Phone      string         `json:"phone" example:"131420203214"`                            //电话号码
	Cluster    string         `json:"cluster" example:"cluster-1"`                             // 集群名称
}

// User2VO 将用户模型转换为视图对象
func User2VO(user model.User, token string) *UserVO {
	return &UserVO{
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		Status:     user.Status,
		Cluster:    user.Cluster,
		Token:      token,
		Phone:      user.Phone,
		Department: user.Department,
	}
}

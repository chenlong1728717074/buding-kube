package password

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

// Options 密码加密选项
var Options = bcrypt.DefaultCost

// GeneratePassword 生成加密密码和盐值
// 参数:
// - code: 原始密码
// 返回:
// - 加密后的密码
// - 盐值
func GeneratePassword(code string) (string, string) {
	// 生成随机盐值
	salt := generateSalt(16)

	// 加密密码
	encoded := Encode(code, salt)

	return encoded, salt
}

// CheckPassword 验证密码是否正确
// 参数:
// - code: 待验证的原始密码
// - salt: 盐值
// - encoded: 加密后的密码
// 返回:
// - 是否验证通过
func CheckPassword(code, salt, encoded string) bool {
	return Verify(code, salt, encoded, Options)
}

// Encode 使用bcrypt加密密码
// 参数:
// - code: 原始密码
// - salt: 盐值
// 返回:
// - 加密后的密码
func Encode(code string, salt string) string {
	// 将密码与盐值组合
	combined := code + salt

	// 使用bcrypt加密
	hashed, err := bcrypt.GenerateFromPassword([]byte(combined), Options)
	if err != nil {
		// 出错时返回空字符串
		return ""
	}

	return string(hashed)
}

// Verify 验证密码
// 参数:
// - code: 待验证的原始密码
// - salt: 盐值
// - encoded: 加密后的密码
// - cost: 加密成本
// 返回:
// - 是否验证通过
func Verify(code, salt, encoded string, cost int) bool {
	// 将密码与盐值组合
	combined := code + salt

	// 使用bcrypt验证
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(combined))

	return err == nil
}

// generateSalt 生成随机盐值
// 参数:
// - length: 盐值长度
// 返回:
// - 随机盐值
func generateSalt(length int) string {
	b := make([]byte, length)

	// 使用安全随机数生成器
	_, err := rand.Read(b)
	if err != nil {
		// 如果生成随机数失败，使用默认字符串
		return "defaultsalt12345"
	}

	// 使用base64编码为字符串
	return base64.StdEncoding.EncodeToString(b)[:length]
}

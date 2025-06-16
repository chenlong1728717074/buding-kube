package jwt

import (
	"buding-kube/internal/model"
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

var jwtSecret = []byte("your-secret-key") // 在实际应用中应该从配置文件中读取

type Claims struct {
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	Cluster  string         `json:"cluster"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT token
func GenerateToken(user *model.User) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)

	claims := Claims{
		Username: user.Username,
		Role:     user.Role,
		Cluster:  user.Cluster,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  nowTime.Unix(),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString(jwtSecret)
}

// ParseToken 解析JWT token
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

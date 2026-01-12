package jwt

import (
	"buding-kube/internal/model"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var Token TokenManager

func init() {
	Token = TokenManager{
		JwtSecret: getJWTSecret(),
	}
}

type TokenManager struct {
	JwtSecret []byte
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		panic("JWT_SECRET_KEY environment variable is not set")
	}
	return []byte(secret)
}

type Claims struct {
	Username string `json:"username"`
	Role     int    `json:"role"`
	Cluster  string `json:"cluster"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func (m *TokenManager) GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := Claims{
		Username: user.Username,
		Role:     int(user.Role),
		Cluster:  user.Cluster,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kube.buding.goaigc.fun", // 添加签发者
			Subject:   user.Username,            // 添加主题
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ParseToken 解析JWT token
func (m *TokenManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法，防止算法替换攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.JwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token is malformed")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("token signature is invalid")
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != "kube.buding.goaigc.fun" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}

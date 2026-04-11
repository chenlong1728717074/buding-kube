package utils

import (
	"buding-kube/pkg/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const secretPrefix = "enc:v1:"

func buildSecretKey() ([]byte, error) {
	key := strings.TrimSpace(os.Getenv("KUBE_SECRET_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("JWT_SECRET_KEY"))
	}
	if key == "" {
		key = strings.TrimSpace(config.GetConfig().Server.JwtSecret)
	}
	if key == "" {
		return nil, errors.New("未配置加密密钥，请设置 KUBE_SECRET_KEY 或 JWT_SECRET_KEY")
	}
	sum := sha256.Sum256([]byte(key))
	return sum[:], nil
}

func EncryptSensitive(plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", nil
	}
	if strings.HasPrefix(plain, secretPrefix) {
		return plain, nil
	}

	secretKey, err := buildSecretKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("初始化加密器失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化GCM失败: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成随机nonce失败: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, ciphertext...)
	return secretPrefix + base64.StdEncoding.EncodeToString(payload), nil
}

func DecryptSensitive(cipherValue string) (string, error) {
	cipherValue = strings.TrimSpace(cipherValue)
	if cipherValue == "" {
		return "", nil
	}
	if !strings.HasPrefix(cipherValue, secretPrefix) {
		return cipherValue, nil
	}

	raw := strings.TrimPrefix(cipherValue, secretPrefix)
	payload, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("解码密文失败: %w", err)
	}

	secretKey, err := buildSecretKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("初始化解密器失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化GCM失败: %w", err)
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("密文格式无效")
	}

	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}
	return string(plain), nil
}

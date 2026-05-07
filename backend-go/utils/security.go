package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
	"wholesale/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tempPasswordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// HashPassword 生成 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateTemporaryPassword 生成随机临时密码
func GenerateTemporaryPassword(length int) string {
	if length < 8 {
		length = 10
	}
	result := make([]byte, length)
	alphabetLen := big.NewInt(int64(len(tempPasswordAlphabet)))
	for i := range result {
		idx, _ := rand.Int(rand.Reader, alphabetLen)
		result[i] = tempPasswordAlphabet[idx.Int64()]
	}
	return string(result)
}

// Claims JWT 声明
type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// CreateAccessToken 生成 JWT token
func CreateAccessToken(userID int64) (string, error) {
	expireAt := time.Now().UTC().Add(
		time.Duration(config.C.AccessTokenExpireMinutes) * time.Minute,
	)
	claims := Claims{
		Sub: fmt.Sprintf("%d", userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.SecretKey))
}

// ParseToken 解析并验证 JWT token，返回 user_id
func ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(config.C.SecretKey), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	var uid int64
	_, err = fmt.Sscanf(claims.Sub, "%d", &uid)
	if err != nil {
		return 0, fmt.Errorf("invalid sub: %s", claims.Sub)
	}
	return uid, nil
}

package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   int    `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret        []byte
	expires       time.Duration
	refreshSecret []byte
	refreshExpiry time.Duration
}

func NewJWTManager(secret string, expiresHours int) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		expires:       time.Duration(expiresHours) * time.Hour,
		refreshSecret: []byte(secret + ":refresh"), // 使用不同的密钥派生
		refreshExpiry: time.Duration(expiresHours*7) * time.Hour, // refresh token 有效期为 access token 的 7 倍
	}
}

func (j *JWTManager) Generate(userID string) (string, error) {
	return j.generateToken(userID, 0, j.secret, j.expires, "wuxie-api")
}

func (j *JWTManager) GenerateWithRole(userID string, role int) (string, error) {
	return j.generateToken(userID, role, j.secret, j.expires, "wuxie-api")
}

func (j *JWTManager) generateToken(userID string, role int, secret []byte, expires time.Duration, issuer string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (j *JWTManager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 验证 Issuer
		if claims.Issuer != "" && claims.Issuer != "wuxie-api" {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// GenerateRefreshToken 生成 refresh token
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wuxie-api-refresh",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.refreshSecret)
}

// ParseRefreshToken 验证 refresh token
func (j *JWTManager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Issuer != "" && claims.Issuer != "wuxie-api-refresh" {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// RefreshExpirySeconds 返回 refresh token 的有效期（秒），供客户端使用
func (j *JWTManager) RefreshExpirySeconds() int {
	return int(j.refreshExpiry.Seconds())
}

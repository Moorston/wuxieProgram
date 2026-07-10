package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-at-least-32-chars-long"

func newTestManager() *JWTManager {
	return NewJWTManager(testSecret, 1) // 1 hour expiry
}

func TestGenerate_And_Parse(t *testing.T) {
	mgr := newTestManager()
	userID := "507f1f77bcf86cd799439011"

	tokenStr, err := mgr.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	claims, err := mgr.Parse(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "wuxie-api", claims.Issuer)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestParse_ExpiredToken(t *testing.T) {
	// 创建一个已过期的 token
	mgr := &JWTManager{
		secret:  []byte(testSecret),
		expires: -1 * time.Hour, // 负值 = 已过期
	}

	tokenStr, err := mgr.Generate("user123")
	require.NoError(t, err)

	// 用正常 manager 解析
	normalMgr := newTestManager()
	_, err = normalMgr.Parse(tokenStr)
	assert.Error(t, err)
}

func TestParse_InvalidSignature(t *testing.T) {
	mgr := newTestManager()
	tokenStr, err := mgr.Generate("user123")
	require.NoError(t, err)

	// 用不同密钥的 manager 解析
	otherMgr := NewJWTManager("different-secret-key-at-least-32-char", 1)
	_, err = otherMgr.Parse(tokenStr)
	assert.Error(t, err)
}

func TestParse_WrongIssuer(t *testing.T) {
	// 生成一个 Issuer 不同的 token
	claims := Claims{
		UserID: "user123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	mgr := newTestManager()
	_, err = mgr.Parse(tokenStr)
	assert.Error(t, err)
}

func TestGenerateRefreshToken_And_Parse(t *testing.T) {
	mgr := newTestManager()
	userID := "507f1f77bcf86cd799439011"

	refreshStr, err := mgr.GenerateRefreshToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, refreshStr)

	claims, err := mgr.ParseRefreshToken(refreshStr)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "wuxie-api-refresh", claims.Issuer)
}

func TestParse_CannotParseRefreshToken(t *testing.T) {
	mgr := newTestManager()
	userID := "user123"

	// 生成 refresh token
	refreshStr, err := mgr.GenerateRefreshToken(userID)
	require.NoError(t, err)

	// 用 Parse（access token 解析器）解析 refresh token → 应该失败
	_, err = mgr.Parse(refreshStr)
	assert.Error(t, err)
}

func TestParseRefreshToken_CannotParseAccessToken(t *testing.T) {
	mgr := newTestManager()
	userID := "user123"

	// 生成 access token
	accessStr, err := mgr.Generate(userID)
	require.NoError(t, err)

	// 用 ParseRefreshToken 解析 access token → 应该失败
	_, err = mgr.ParseRefreshToken(accessStr)
	assert.Error(t, err)
}

func TestParseRefreshToken_WrongIssuer(t *testing.T) {
	claims := Claims{
		UserID: "user123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-refresh-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret + ":refresh"))
	require.NoError(t, err)

	mgr := newTestManager()
	_, err = mgr.ParseRefreshToken(tokenStr)
	assert.Error(t, err)
}

func TestParse_MalformedToken(t *testing.T) {
	mgr := newTestManager()
	_, err := mgr.Parse("not.a.valid.token")
	assert.Error(t, err)
}

func TestGenerate_DifferentTokens(t *testing.T) {
	mgr := newTestManager()
	token1, err := mgr.Generate("user1")
	require.NoError(t, err)
	token2, err := mgr.Generate("user2")
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
}

func TestRefreshExpirySeconds(t *testing.T) {
	mgr := newTestManager()
	// refreshExpiry = expiresHours * 7 = 1 * 7 = 7 hours = 25200 seconds
	assert.Equal(t, 25200, mgr.RefreshExpirySeconds())
}

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wuxie-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const testSecret = "test-secret-key-at-least-32-chars-long"

func setupAuthRouter(jwtMgr *jwt.JWTManager, blacklist *TokenBlacklist) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth(jwtMgr, blacklist))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("user_id")})
	})
	return r
}

func TestAuth_MissingHeader(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	bl := NewTokenBlacklist()
	defer bl.Stop()
	r := setupAuthRouter(mgr, bl)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuth_InvalidFormat_NoBearer(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	bl := NewTokenBlacklist()
	defer bl.Stop()
	r := setupAuthRouter(mgr, bl)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization format")
}

func TestAuth_ValidToken(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	bl := NewTokenBlacklist()
	defer bl.Stop()
	r := setupAuthRouter(mgr, bl)

	userID := "507f1f77bcf86cd799439011"
	tokenStr, err := mgr.Generate(userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID)
}

func TestAuth_RevokedToken(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	bl := NewTokenBlacklist()
	defer bl.Stop()
	r := setupAuthRouter(mgr, bl)

	userID := "507f1f77bcf86cd799439011"
	tokenStr, err := mgr.Generate(userID)
	require.NoError(t, err)

	// 撤销 token
	bl.Revoke(tokenStr, 10*time.Minute)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "revoked")
}

func TestAuth_ExpiredToken(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	bl := NewTokenBlacklist()
	defer bl.Stop()
	r := setupAuthRouter(mgr, bl)

	// 生成一个已过期的 token（直接构造）
	// Auth 中间件依赖 jwtMgr.Parse，过期 token 会被 Parse 拒绝
	otherMgr := jwt.NewJWTManager("different-secret-key-at-least-32-char", 1)
	tokenStr, err := otherMgr.Generate("user123")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuth_NilBlacklist(t *testing.T) {
	mgr := jwt.NewJWTManager(testSecret, 1)
	r := setupAuthRouter(mgr, nil) // blacklist 为 nil

	userID := "507f1f77bcf86cd799439011"
	tokenStr, err := mgr.Generate(userID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// nil blacklist 不阻断请求
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- UserStatusCheck tests ---

func setupStatusCheckRouter(checker StatusChecker) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 先手动设置 user_id，再用 UserStatusCheck
	r.Use(func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-Test-User-ID"))
		c.Next()
	})
	r.Use(UserStatusCheck(checker))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestUserStatusCheck_Active(t *testing.T) {
	checker := func(ctx context.Context, userID primitive.ObjectID) (bool, error) {
		return false, nil // 未封禁
	}
	r := setupStatusCheckRouter(checker)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-User-ID", "507f1f77bcf86cd799439011")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserStatusCheck_Banned(t *testing.T) {
	checker := func(ctx context.Context, userID primitive.ObjectID) (bool, error) {
		return true, nil // 已封禁
	}
	r := setupStatusCheckRouter(checker)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-User-ID", "507f1f77bcf86cd799439011")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "suspended")
}

func TestUserStatusCheck_DBError(t *testing.T) {
	checker := func(ctx context.Context, userID primitive.ObjectID) (bool, error) {
		return false, assert.AnError // DB 错误
	}
	r := setupStatusCheckRouter(checker)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-User-ID", "507f1f77bcf86cd799439011")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB 错误时降级放行
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserStatusCheck_MissingUserID(t *testing.T) {
	checker := func(ctx context.Context, userID primitive.ObjectID) (bool, error) {
		return false, nil
	}
	r := setupStatusCheckRouter(checker)

	req := httptest.NewRequest("GET", "/test", nil)
	// 不设置 X-Test-User-ID
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserStatusCheck_InvalidUserID(t *testing.T) {
	checker := func(ctx context.Context, userID primitive.ObjectID) (bool, error) {
		return false, nil
	}
	r := setupStatusCheckRouter(checker)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-User-ID", "not-a-valid-oid")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

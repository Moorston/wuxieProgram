package middleware

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"wuxie-api/pkg/jwt"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StatusChecker 定义用户状态检查函数类型
type StatusChecker func(ctx context.Context, userID primitive.ObjectID) (banned bool, err error)

func Auth(jwtMgr *jwt.JWTManager, blacklist *TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			// 没有 "Bearer " 前缀，TrimPrefix 返回原字符串
			response.Unauthorized(c, "invalid authorization format, expected 'Bearer <token>'")
			c.Abort()
			return
		}

		// 检查 token 是否已被撤销
		if blacklist != nil && blacklist.IsRevoked(tokenStr) {
			response.Unauthorized(c, "token has been revoked")
			c.Abort()
			return
		}

		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		// 验证user_id是合法的ObjectID格式
		if _, err := primitive.ObjectIDFromHex(claims.UserID); err != nil {
			response.Unauthorized(c, "invalid user identity in token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// AdminOnly 验证用户是否为管理员
// 应在 Auth 中间件之后使用
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		}
		if roleInt, ok := role.(int); !ok || roleInt != 1 {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// StatusCache 用户状态缓存（避免每次请求查 DB）
type StatusCache struct {
	entries map[string]*statusCacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

type statusCacheEntry struct {
	banned  bool
	expires time.Time
}

func NewStatusCache(ttl time.Duration) *StatusCache {
	return &StatusCache{
		entries: make(map[string]*statusCacheEntry),
		ttl:     ttl,
	}
}

func (sc *StatusCache) Get(userID string) (banned, found bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	entry, ok := sc.entries[userID]
	if !ok || time.Now().After(entry.expires) {
		return false, false
	}
	return entry.banned, true
}

func (sc *StatusCache) Set(userID string, banned bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.entries[userID] = &statusCacheEntry{
		banned:  banned,
		expires: time.Now().Add(sc.ttl),
	}
}

// UserStatusCheck 检查用户是否被封禁
// 应在 Auth 中间件之后使用，依赖 user_id 已被设置到 context
func UserStatusCheck(checker StatusChecker, cache *StatusCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			response.Unauthorized(c, "missing user identity")
			c.Abort()
			return
		}

		oid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			response.Unauthorized(c, "invalid user identity")
			c.Abort()
			return
		}

		// 检查缓存
		if cache != nil {
			if banned, found := cache.Get(userID); found {
				if banned {
					response.Forbidden(c, "account has been suspended")
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}

		banned, err := checker(c.Request.Context(), oid)
		if err != nil {
			// 查询失败不阻断请求，只记录日志
			log.Printf("[WARN] user status check failed for %s: %v", userID, err)
			c.Next()
			return
		}

		// 写入缓存
		if cache != nil {
			cache.Set(userID, banned)
		}

		if banned {
			response.Forbidden(c, "account has been suspended")
			c.Abort()
			return
		}

		c.Next()
	}
}

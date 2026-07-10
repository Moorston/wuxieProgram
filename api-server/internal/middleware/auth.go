package middleware

import (
	"context"
	"log"
	"strings"

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
		c.Next()
	}
}

// UserStatusCheck 检查用户是否被封禁
// 应在 Auth 中间件之后使用，依赖 user_id 已被设置到 context
// TODO: 生产环境应增加 Redis 缓存（TTL 30s），避免每次请求查 DB
func UserStatusCheck(checker StatusChecker) gin.HandlerFunc {
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

		banned, err := checker(c.Request.Context(), oid)
		if err != nil {
			// 查询失败不阻断请求，只记录日志
			log.Printf("[WARN] user status check failed for %s: %v", userID, err)
			c.Next()
			return
		}

		if banned {
			response.Forbidden(c, "account has been suspended")
			c.Abort()
			return
		}

		c.Next()
	}
}

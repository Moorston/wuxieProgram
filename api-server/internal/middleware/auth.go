package middleware

import (
	"strings"

	"wuxie-api/pkg/jwt"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Auth(jwtMgr *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
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

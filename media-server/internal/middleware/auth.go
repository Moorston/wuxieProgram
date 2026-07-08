package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APISecretAuth 校验请求中的 API 密钥，保护内部接口
func APISecretAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 支持两种方式: Authorization: Bearer <secret> 或 X-API-Key: <secret>
		token := ""

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = c.GetHeader("X-API-Key")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			return
		}

		if token != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid API key",
			})
			return
		}

		c.Next()
	}
}

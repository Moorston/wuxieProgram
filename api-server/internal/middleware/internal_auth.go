package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalAuth 验证内部API调用（如media-server回调）
func InternalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取内部密钥
		providedSecret := c.GetHeader("X-Internal-Secret")

		// 验证密钥
		if providedSecret == "" || providedSecret != secret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Invalid internal secret",
			})
			return
		}

		c.Next()
	}
}

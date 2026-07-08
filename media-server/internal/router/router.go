package router

import (
	"wuxie-media/internal/handler"
	"wuxie-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(uploadH *handler.UploadHandler, mediaH *handler.MediaHandler, apiSecret string) *gin.Engine {
	r := gin.Default()

	// 公开路由 - 无需认证
	r.GET("/health", uploadH.Health)

	// 受保护路由 - 需要 API 密钥认证
	protected := r.Group("/media")
	protected.Use(middleware.APISecretAuth(apiSecret))
	{
		protected.GET("/upload/presign", uploadH.Presign)
		protected.POST("/upload/callback", uploadH.UploadCallback)
		protected.GET("/url", mediaH.GetURL)
	}

	return r
}

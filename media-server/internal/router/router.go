package router

import (
	"wuxie-media/internal/handler"

	"github.com/gin-gonic/gin"
)

func Setup(uploadH *handler.UploadHandler, mediaH *handler.MediaHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", uploadH.Health)

	media := r.Group("/media")
	{
		media.GET("/upload/presign", uploadH.Presign)
		media.POST("/upload/callback", uploadH.UploadCallback)
		media.GET("/url", mediaH.GetURL)
	}

	return r
}

package router

import (
	"wuxie-api/internal/handler"
	"wuxie-api/internal/middleware"
	"wuxie-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup(
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	checkinH *handler.CheckinHandler,
	socialH *handler.SocialHandler,
	rankH *handler.RankHandler,
	groupH *handler.GroupHandler,
	jwtMgr *jwt.JWTManager,
	logger *zap.Logger,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.RequestID())
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		// 公开接口
		api.POST("/auth/login", authH.Login)

		// 需要鉴权的接口
		auth := api.Group("", middleware.Auth(jwtMgr))
		{
			// 用户
			auth.GET("/user/profile", userH.GetProfile)
			auth.PUT("/user/profile", userH.UpdateProfile)

			// 打卡
			auth.POST("/checkin/prepare", checkinH.Prepare)
			auth.GET("/checkin/list", checkinH.GetList)
			auth.GET("/checkin/mine", checkinH.GetMine)
			auth.DELETE("/checkin/:id", checkinH.Delete)

			// 社交
			auth.POST("/checkin/:id/like", socialH.ToggleLike)
			auth.POST("/checkin/:id/comment", socialH.AddComment)
			auth.GET("/checkin/:id/comments", socialH.GetComments)

			// 排行榜
			auth.GET("/rank", rankH.GetRankList)
			auth.GET("/rank/me", rankH.GetMyRank)

			// 考核组
			auth.GET("/group/list", groupH.List)
			auth.GET("/group/:id", groupH.Detail)
		}

		// 内部接口（media-server回调）
		internal := api.Group("/internal")
		{
			internal.POST("/transcode/done", checkinH.TranscodeCallback)
		}
	}

	return r
}

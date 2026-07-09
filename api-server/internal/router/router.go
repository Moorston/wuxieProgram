package router

import (
	"net/http"
	"time"

	"wuxie-api/internal/config"
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
	trainingH *handler.TrainingHandler,
	notifH *handler.NotificationHandler,
	insightH *handler.InsightHandler,
	resourceH *handler.ResourceHandler,
	jwtMgr *jwt.JWTManager,
	logger *zap.Logger,
	cfg *config.Config,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.RequestID())
	r.Use(gin.Recovery())

	// 健康检查端点（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

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
			auth.GET("/checkin/search", checkinH.Search)
			auth.GET("/checkin/:id", checkinH.GetByID)
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

			// 训练计划
			auth.POST("/training/plan", trainingH.CreatePlan)
			auth.GET("/training/plans", trainingH.ListPlans)
			auth.GET("/training/plan/:id", trainingH.GetPlan)
			auth.PUT("/training/plan/:id", trainingH.UpdatePlan)
			auth.DELETE("/training/plan/:id", trainingH.DeletePlan)
			auth.GET("/training/today", trainingH.TodayTasks)
			auth.PUT("/training/task/:plan_id/:day/:task_idx", trainingH.UpdateTask)
			auth.GET("/training/plan/:id/report", trainingH.GetReport)

			// 训练模板
			auth.GET("/training/template/list", trainingH.ListTemplates)
			auth.GET("/training/template/:id", trainingH.GetTemplate)
			auth.POST("/training/template/:id/apply", trainingH.ApplyTemplate)

			// 通知
			auth.GET("/notification/list", notifH.List)
			auth.GET("/notification/unread", notifH.UnreadCount)
			auth.PUT("/notification/read/:id", notifH.MarkRead)
			auth.PUT("/notification/read-all", notifH.MarkAllRead)
			auth.DELETE("/notification/:id", notifH.Delete)
			auth.GET("/notification/settings", notifH.GetSettings)
			auth.PUT("/notification/settings", notifH.UpdateSettings)

			// 感悟笔记
			auth.POST("/insight", insightH.Create)
			auth.GET("/insight/list", insightH.List)
			auth.GET("/insight/public", insightH.ListPublic)
			auth.GET("/insight/tags", insightH.GetTags)
			auth.GET("/insight/mood-stats", insightH.MoodStats)
			auth.GET("/insight/on-this-day", insightH.OnThisDay)
			auth.GET("/insight/:id", insightH.GetByID)
			auth.PUT("/insight/:id", insightH.Update)
			auth.DELETE("/insight/:id", insightH.Delete)
			auth.POST("/insight/:id/like", insightH.Like)

			// 个人资料库
			auth.GET("/resource/upload/presign", resourceH.Presign)
			auth.POST("/resource/upload/callback", resourceH.UploadCallback)
			auth.POST("/resource", resourceH.Create)
			auth.GET("/resource/list", resourceH.List)
			auth.GET("/resource/tags", resourceH.GetTags)
			auth.GET("/resource/stats", resourceH.GetStats)
			auth.GET("/resource/favorites", resourceH.ListFavorites)
			auth.GET("/resource/:id", resourceH.GetByID)
			auth.PUT("/resource/:id", resourceH.Update)
			auth.DELETE("/resource/:id", resourceH.Delete)
			auth.POST("/resource/:id/favorite", resourceH.ToggleFavorite)
		}

		// 内部接口（media-server回调），需要内部API密钥认证
		internal := api.Group("/internal")
		internal.Use(middleware.InternalAuth(cfg.MediaSecret))
		{
			internal.POST("/transcode/done", checkinH.TranscodeCallback)
		}
	}

	return r
}
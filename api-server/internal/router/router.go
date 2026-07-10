package router

import (
	"net/http"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/handler"
	"wuxie-api/internal/middleware"
	"wuxie-api/internal/repository"
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
	analyticsH *handler.AnalyticsHandler,
	followH *handler.FollowHandler,
	compH *handler.CompetitionHandler,
	badgeH *handler.BadgeHandler,
	annH *handler.GroupAnnouncementHandler,
	challengeH *handler.ChallengeHandler,
	rankHistoryH *handler.RankHistoryHandler,
	adminH *handler.AdminHandler,
	jwtMgr *jwt.JWTManager,
	blacklist *middleware.TokenBlacklist,
	userRepo *repository.UserRepo,
	logger *zap.Logger,
	cfg *config.Config,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
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
		api.POST("/auth/login", middleware.LoginRateLimit(), authH.Login)
		api.POST("/auth/refresh", middleware.LoginRateLimit(), authH.Refresh)

		// 需要鉴权的接口
		auth := api.Group("", middleware.Auth(jwtMgr, blacklist), middleware.UserStatusCheck(userRepo.IsBanned))
		{
			// 用户
			auth.GET("/user/profile", userH.GetProfile)
			auth.PUT("/user/profile", userH.UpdateProfile)
			auth.GET("/user/privacy", userH.GetPrivacySettings)
			auth.PUT("/user/privacy", userH.UpdatePrivacySettings)
			auth.GET("/user/level", userH.GetUserLevel)
			auth.POST("/auth/logout", authH.Logout)

			// 打卡
			auth.POST("/checkin/prepare", checkinH.Prepare)
			auth.GET("/checkin/list", checkinH.GetList)
			auth.GET("/checkin/mine", checkinH.GetMine)
			auth.GET("/checkin/export", checkinH.ExportMine)
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
			auth.GET("/rank/trend", rankHistoryH.GetRankTrend)

			// 考核组
			auth.GET("/group/list", groupH.List)
			auth.GET("/group/:id", groupH.Detail)
			auth.POST("/group/:id/invite", groupH.GenerateInviteCode)
			auth.POST("/group/join", groupH.JoinByInviteCode)
			auth.POST("/group/:id/remove-member", groupH.RemoveMember)
			auth.POST("/group/:id/leave", groupH.LeaveGroup)
			auth.POST("/group/:id/set-leader", groupH.SetLeader)

			// 团组公告
			auth.POST("/group/announcements", annH.Create)
			auth.GET("/group/:group_id/announcements", annH.List)
			auth.DELETE("/group/announcements/:id", annH.Delete)

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

			// 数据分析
			auth.GET("/analytics/checkin-heatmap", analyticsH.GetCheckinHeatmap)
			auth.GET("/analytics/checkin-trend", analyticsH.GetCheckinTrend)
			auth.GET("/analytics/overview", analyticsH.GetOverview)

			// 关注系统
			auth.POST("/follow/:id", followH.Follow)
			auth.DELETE("/follow/:id", followH.Unfollow)
			auth.GET("/follow/following", followH.GetFollowing)
			auth.GET("/follow/followers", followH.GetFollowers)
			auth.GET("/feed", followH.GetFeed)
			auth.GET("/user/:id/profile", followH.GetUserProfile)

			// 赛事活动
			auth.GET("/competitions", compH.List)
			auth.GET("/competitions/:id", compH.Detail)
			auth.POST("/competitions/:id/submit", compH.Submit)
			auth.GET("/competitions/:id/entries", compH.Entries)
			auth.GET("/competitions/:id/ranking", compH.Ranking)
			auth.POST("/competitions/:id/entries/:entryId/score", compH.Score)

			// 徽章
			auth.GET("/badges", badgeH.GetAllBadges)
			auth.GET("/badges/my", badgeH.GetUserBadges)

			// 打卡挑战
			auth.POST("/challenges", challengeH.Create)
			auth.GET("/challenges", challengeH.List)
			auth.GET("/challenges/:id", challengeH.Detail)
			auth.POST("/challenges/:id/join", challengeH.Join)
			auth.GET("/challenges/:id/ranking", challengeH.Ranking)

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

		// 管理后台接口（需要管理员认证）
		admin := api.Group("/admin")
		{
			admin.POST("/login", adminH.Login)
			adminAuth := admin.Group("", middleware.Auth(jwtMgr, blacklist), middleware.AdminOnly())
			{
				adminAuth.GET("/dashboard", adminH.GetDashboard)
				adminAuth.GET("/config", adminH.GetSystemConfig)

				// 用户管理
				adminAuth.GET("/users", adminH.GetUsers)
				adminAuth.GET("/users/:id", adminH.GetUserDetail)
				adminAuth.PUT("/users/:id/ban", adminH.BanUser)
				adminAuth.PUT("/users/:id/unban", adminH.UnbanUser)
				adminAuth.POST("/users/batch-ban", adminH.BatchBanUsers)

				// 内容管理
				adminAuth.GET("/checkins", adminH.GetCheckins)
				adminAuth.DELETE("/checkins/:id", adminH.DeleteCheckin)
				adminAuth.POST("/checkins/batch-delete", adminH.BatchDeleteCheckins)
				adminAuth.GET("/insights", adminH.GetInsights)
				adminAuth.DELETE("/insights/:id", adminH.DeleteInsight)
				adminAuth.POST("/insights/batch-delete", adminH.BatchDeleteInsights)

				// 导出
				adminAuth.GET("/export/users", adminH.ExportUsers)
				adminAuth.GET("/export/checkins", adminH.ExportCheckins)
				adminAuth.GET("/export/insights", adminH.ExportInsights)

				// 操作日志
				adminAuth.GET("/audit-logs", adminH.GetAuditLogs)

				// 赛事管理
				adminAuth.POST("/competitions", compH.Create)
				adminAuth.GET("/competitions", compH.AdminList)
				adminAuth.PUT("/competitions/:id", compH.AdminUpdate)
				adminAuth.GET("/competitions/:id/export", compH.ExportRanking)
			}
		}
	}

	return r
}
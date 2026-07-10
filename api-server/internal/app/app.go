package app

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/handler"
	"wuxie-api/internal/middleware"
	"wuxie-api/internal/repository"
	"wuxie-api/internal/router"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/jwt"
	wxpkg "wuxie-api/pkg/wx"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// App 是 Composition Root，负责组装所有依赖
type App struct {
	server     *http.Server
	mongo      *mongo.Client
	blacklist  *middleware.TokenBlacklist
	cronCancel context.CancelFunc
	logger     *zap.Logger
}

// New 创建 App 实例，完成所有依赖注入
func New(cfg *config.Config) (*App, error) {
	// 验证JWT密钥强度
	if len(cfg.JWT.Secret) < 32 {
		log.Fatal("FATAL: JWT secret must be at least 32 characters long")
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// 连接MongoDB（带连接池配置）
	mongoURI := cfg.Mongo.URI
	if !strings.Contains(mongoURI, "maxPoolSize") {
		separator := "?"
		if strings.Contains(mongoURI, "?") {
			separator = "&"
		}
		mongoURI += separator + "maxPoolSize=100&minPoolSize=10&socketTimeoutMS=30000&serverSelectionTimeoutMS=5000"
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(cfg.Mongo.Database)

	// 基础设施
	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expires)
	tokenBlacklist := middleware.NewTokenBlacklist()

	// Repository
	userRepo := repository.NewUserRepo(db)
	checkinRepo := repository.NewCheckinRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	groupRepo := repository.NewGroupRepo(db)
	rankRepo := repository.NewRankRepo(db)
	trainingRepo := repository.NewTrainingRepo(db)
	templateRepo := repository.NewTemplateRepo(db)
	notifRepo := repository.NewNotificationRepo(db)
	notifSettingsRepo := repository.NewNotificationSettingsRepo(db)
	insightRepo := repository.NewInsightRepo(db)
	insightTagRepo := repository.NewInsightTagRepo(db)
	insightLikeRepo := repository.NewInsightLikeRepo(db)
	resourceRepo := repository.NewResourceRepo(db)
	resourceTagRepo := repository.NewResourceTagRepo(db)

	// Audit Log
	auditLogRepo := repository.NewAuditLogRepo(db)

	// Follow
	followRepo := repository.NewFollowRepo(db)

	// Competition
	compRepo := repository.NewCompetitionRepo(db)
	entryRepo := repository.NewCompetitionEntryRepo(db)

	// Badge
	badgeRepo := repository.NewBadgeRepo(db)
	userBadgeRepo := repository.NewUserBadgeRepo(db)

	// Group Announcement
	annRepo := repository.NewGroupAnnouncementRepo(db)

	// Challenge
	challengeRepo := repository.NewChallengeRepo(db)
	participantRepo := repository.NewChallengeParticipantRepo(db)

	// 创建索引
	ctx := context.Background()
	for name, ensureFn := range map[string]func(context.Context) error{
		"user":           userRepo.EnsureIndexes,
		"checkin":        checkinRepo.EnsureIndexes,
		"like":           likeRepo.EnsureIndexes,
		"training":       trainingRepo.EnsureIndexes,
		"template":       templateRepo.EnsureIndexes,
		"notification":   notifRepo.EnsureIndexes,
		"notif_settings": notifSettingsRepo.EnsureIndexes,
		"insight":        insightRepo.EnsureIndexes,
		"insight_tag":    insightTagRepo.EnsureIndexes,
		"insight_like":   insightLikeRepo.EnsureIndexes,
		"resource":       resourceRepo.EnsureIndexes,
		"resource_tag":   resourceTagRepo.EnsureIndexes,
		"audit_log":      auditLogRepo.EnsureIndexes,
		"follow":         followRepo.EnsureIndexes,
		"competition":    compRepo.EnsureIndexes,
		"competition_entry": entryRepo.EnsureIndexes,
		"badge":           badgeRepo.EnsureIndexes,
		"user_badge":      userBadgeRepo.EnsureIndexes,
		"group_announcement": annRepo.EnsureIndexes,
		"challenge":        challengeRepo.EnsureIndexes,
		"challenge_participant": participantRepo.EnsureIndexes,
	} {
		if err := ensureFn(ctx); err != nil {
			log.Printf("WARNING: ensure index %s failed: %v", name, err)
		}
	}

	// Service
	authService := service.NewAuthService(userRepo, jwtMgr, cfg, logger)
	userService := service.NewUserService(userRepo)
	checkinService := service.NewCheckinService(checkinRepo, userRepo, cfg.MediaURL)
	notifService := service.NewNotificationService(notifRepo, notifSettingsRepo, userRepo)
	socialService := service.NewSocialService(commentRepo, likeRepo, checkinRepo, userRepo, notifService)
	rankService := service.NewRankService(rankRepo)
	groupService := service.NewGroupService(groupRepo, userRepo, logger)
	trainingService := service.NewTrainingService(trainingRepo, templateRepo, notifService)
	insightService := service.NewInsightService(insightRepo, insightTagRepo, insightLikeRepo, userRepo)
	resourceService := service.NewResourceService(resourceRepo, resourceTagRepo, userRepo)

	// WX Client
	var wxClient *wxpkg.Client
	if cfg.WX.AppID != "" && cfg.WX.Secret != "" {
		wxClient = wxpkg.NewClient(cfg.WX.AppID, cfg.WX.Secret)
	}

	// Cron Service
	cronService := service.NewCronService(userRepo, checkinRepo, rankRepo, trainingRepo, notifRepo, wxClient, cfg, logger)

	// Admin Service
	adminService := service.NewAdminService(userRepo, checkinRepo, insightRepo, auditLogRepo, jwtMgr, cfg, logger)

	// Analytics Service
	analyticsService := service.NewAnalyticsService(checkinRepo, userRepo, logger)

	// Follow Service
	followService := service.NewFollowService(followRepo, checkinRepo, insightRepo, userRepo, logger)

	// Competition Service
	compService := service.NewCompetitionService(compRepo, entryRepo, checkinRepo, userRepo, logger)

	// Badge Service
	badgeService := service.NewBadgeService(badgeRepo, userBadgeRepo, userRepo, logger)
	badgeService.SeedDefaults(ctx) // 初始化默认徽章

	// Group Announcement Service
	annService := service.NewGroupAnnouncementService(annRepo, groupRepo, userRepo, logger)

	// Challenge Service
	challengeService := service.NewChallengeService(challengeRepo, participantRepo, userRepo, logger)

	// Handler
	authH := handler.NewAuthHandler(authService, jwtMgr, tokenBlacklist, logger)
	userH := handler.NewUserHandler(userService, logger)
	checkinH := handler.NewCheckinHandler(checkinService, socialService, logger)
	socialH := handler.NewSocialHandler(socialService, logger)
	rankH := handler.NewRankHandler(rankService, logger)
	groupH := handler.NewGroupHandler(groupService, logger)
	trainingH := handler.NewTrainingHandler(trainingService, logger)
	notifH := handler.NewNotificationHandler(notifService, logger)
	insightH := handler.NewInsightHandler(insightService, logger)
	resourceH := handler.NewResourceHandler(resourceService, logger)
	adminH := handler.NewAdminHandler(adminService, logger)
	analyticsH := handler.NewAnalyticsHandler(analyticsService)
	followH := handler.NewFollowHandler(followService, logger)
	compH := handler.NewCompetitionHandler(compService, logger)
	badgeH := handler.NewBadgeHandler(badgeService, logger)
	annH := handler.NewGroupAnnouncementHandler(annService, logger)
	challengeH := handler.NewChallengeHandler(challengeService, logger)

	// 路由
	r := router.Setup(authH, userH, checkinH, socialH, rankH, groupH, trainingH, notifH, insightH, resourceH, analyticsH, followH, compH, badgeH, annH, challengeH, adminH, jwtMgr, tokenBlacklist, userRepo, logger, cfg)

	// HTTP Server
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 定时任务
	cronCtx, cronCancel := context.WithCancel(context.Background())
	go startCronJobs(cronCtx, cronService, cfg)

	return &App{
		server:     srv,
		mongo:      mongoClient,
		blacklist:  tokenBlacklist,
		cronCancel: cronCancel,
		logger:     logger,
	}, nil
}

// Run 启动 HTTP 服务器（非阻塞）
func (a *App) Run() error {
	a.logger.Info("server starting", zap.String("addr", a.server.Addr))
	return a.server.ListenAndServe()
}

// Shutdown 优雅关闭所有资源
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("shutting down...")

	// 停止定时任务
	a.cronCancel()

	// 关闭 HTTP 服务器
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("server shutdown error", zap.Error(err))
	}

	// 停止 Token 黑名单清理
	a.blacklist.Stop()

	// 关闭 MongoDB
	disconnectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := a.mongo.Disconnect(disconnectCtx); err != nil {
		a.logger.Error("mongo disconnect error", zap.Error(err))
	}

	a.logger.Sync()
}

// startCronJobs 启动所有定时任务
func startCronJobs(ctx context.Context, cronService *service.CronService, cfg *config.Config) {
	// 排行榜刷新（每 10 分钟）
	go func() {
		cronService.RefreshAllRanks(ctx)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cronService.RefreshAllRanks(ctx)
			}
		}
	}()

	// 训练提醒（每天定时）
	go func() {
		for {
			now := time.Now()
			remindHour := cfg.WX.RemindHour
			if remindHour == 0 {
				remindHour = 20
			}
			next := time.Date(now.Year(), now.Month(), now.Day(), remindHour, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.AddDate(0, 0, 1)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(next.Sub(now)):
				cronService.SendTrainingReminders(ctx)
			}
		}
	}()
}

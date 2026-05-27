package main

import (
	"context"
	"log"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/handler"
	"wuxie-api/internal/middleware"
	"wuxie-api/internal/repository"
	"wuxie-api/internal/router"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/jwt"
	wxpkg "wuxie-api/pkg/wx"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 连接MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		log.Fatalf("failed to connect mongo: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(cfg.Mongo.Database)

	// 初始化各层
	jwtMgr := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expires)

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
	resourceRepo := repository.NewResourceRepo(db)
	resourceTagRepo := repository.NewResourceTagRepo(db)

	// 创建索引
	ctx := context.Background()
	userRepo.EnsureIndexes(ctx)
	checkinRepo.EnsureIndexes(ctx)
	trainingRepo.EnsureIndexes(ctx)
	templateRepo.EnsureIndexes(ctx)
	notifRepo.EnsureIndexes(ctx)
	notifSettingsRepo.EnsureIndexes(ctx)
	insightRepo.EnsureIndexes(ctx)
	insightTagRepo.EnsureIndexes(ctx)
	resourceRepo.EnsureIndexes(ctx)
	resourceTagRepo.EnsureIndexes(ctx)

	// Service
	authService := service.NewAuthService(userRepo, jwtMgr, cfg)
	userService := service.NewUserService(userRepo)
	checkinService := service.NewCheckinService(checkinRepo, userRepo, cfg.MediaURL)
	notifService := service.NewNotificationService(notifRepo, notifSettingsRepo, userRepo)
	socialService := service.NewSocialService(commentRepo, likeRepo, checkinRepo, userRepo, notifService)
	rankService := service.NewRankService(rankRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	trainingService := service.NewTrainingService(trainingRepo, templateRepo, notifService)
	insightService := service.NewInsightService(insightRepo, insightTagRepo, userRepo)
	resourceService := service.NewResourceService(resourceRepo, resourceTagRepo, userRepo)

	// WX Client
	var wxClient *wxpkg.Client
	if cfg.WX.AppID != "" && cfg.WX.Secret != "" {
		wxClient = wxpkg.NewClient(cfg.WX.AppID, cfg.WX.Secret)
	}

	// Cron Service
	cronService := service.NewCronService(userRepo, checkinRepo, rankRepo, trainingRepo, notifRepo, wxClient, cfg)

	// 启动定时任务
	go func() {
		cronService.RefreshAllRanks(ctx)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cronService.RefreshAllRanks(ctx)
		}
	}()

	// 训练提醒定时任务（每天检查）
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
			time.Sleep(next.Sub(now))
			cronService.SendTrainingReminders(ctx)
		}
	}()

	// Handler
	authH := handler.NewAuthHandler(authService)
	userH := handler.NewUserHandler(userService)
	checkinH := handler.NewCheckinHandler(checkinService, socialService)
	socialH := handler.NewSocialHandler(socialService)
	rankH := handler.NewRankHandler(rankService)
	groupH := handler.NewGroupHandler(groupService)
	trainingH := handler.NewTrainingHandler(trainingService)
	notifH := handler.NewNotificationHandler(notifService)
	insightH := handler.NewInsightHandler(insightService)
	resourceH := handler.NewResourceHandler(resourceService)

	// 路由
	r := router.Setup(authH, userH, checkinH, socialH, rankH, groupH, trainingH, notifH, insightH, resourceH, jwtMgr, logger)

	// 设置模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Printf("api-server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

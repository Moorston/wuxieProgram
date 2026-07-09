package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// 验证JWT密钥强度
	if len(cfg.JWT.Secret) < 32 {
		log.Fatal("FATAL: JWT secret must be at least 32 characters long")
	}
	if cfg.JWT.Secret == "wuxie-jwt-secret-change-in-production" {
		log.Println("WARNING: Using default JWT secret. Please change it in production!")
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	// 连接MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		log.Fatalf("failed to connect mongo: %v", err)
	}

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
	insightLikeRepo := repository.NewInsightLikeRepo(db)
	resourceRepo := repository.NewResourceRepo(db)
	resourceTagRepo := repository.NewResourceTagRepo(db)

	// 创建索引（忽略已存在的索引错误）
	ctx := context.Background()
	for name, ensureFn := range map[string]func(context.Context) error{
		"user":             userRepo.EnsureIndexes,
		"checkin":          checkinRepo.EnsureIndexes,
		"like":             likeRepo.EnsureIndexes,
		"training":         trainingRepo.EnsureIndexes,
		"template":         templateRepo.EnsureIndexes,
		"notification":     notifRepo.EnsureIndexes,
		"notif_settings":   notifSettingsRepo.EnsureIndexes,
		"insight":          insightRepo.EnsureIndexes,
		"insight_tag":      insightTagRepo.EnsureIndexes,
		"insight_like":     insightLikeRepo.EnsureIndexes,
		"resource":         resourceRepo.EnsureIndexes,
		"resource_tag":     resourceTagRepo.EnsureIndexes,
	} {
		if err := ensureFn(ctx); err != nil {
			log.Printf("WARNING: ensure index %s failed: %v", name, err)
		}
	}

	// Service
	authService := service.NewAuthService(userRepo, jwtMgr, cfg)
	userService := service.NewUserService(userRepo)
	checkinService := service.NewCheckinService(checkinRepo, userRepo, cfg.MediaURL)
	notifService := service.NewNotificationService(notifRepo, notifSettingsRepo, userRepo)
	socialService := service.NewSocialService(commentRepo, likeRepo, checkinRepo, userRepo, notifService)
	rankService := service.NewRankService(rankRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)
	trainingService := service.NewTrainingService(trainingRepo, templateRepo, notifService)
	insightService := service.NewInsightService(insightRepo, insightTagRepo, insightLikeRepo, userRepo)
	resourceService := service.NewResourceService(resourceRepo, resourceTagRepo, userRepo)

	// WX Client
	var wxClient *wxpkg.Client
	if cfg.WX.AppID != "" && cfg.WX.Secret != "" {
		wxClient = wxpkg.NewClient(cfg.WX.AppID, cfg.WX.Secret)
	}

	// Cron Service
	cronService := service.NewCronService(userRepo, checkinRepo, rankRepo, trainingRepo, notifRepo, wxClient, cfg)

	// 创建可取消的上下文用于定时任务
	cronCtx, cronCancel := context.WithCancel(context.Background())
	defer cronCancel()

	// 启动定时任务
	go func() {
		cronService.RefreshAllRanks(cronCtx)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-cronCtx.Done():
				log.Println("[cron] rank refresh stopped")
				return
			case <-ticker.C:
				cronService.RefreshAllRanks(cronCtx)
			}
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
			select {
			case <-cronCtx.Done():
				log.Println("[cron] training reminder stopped")
				return
			case <-time.After(next.Sub(now)):
				cronService.SendTrainingReminders(cronCtx)
			}
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

	// 设置模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 路由
	r := router.Setup(authH, userH, checkinH, socialH, rankH, groupH, trainingH, notifH, insightH, resourceH, jwtMgr, logger, cfg)

	// 配置HTTP服务器（含超时控制）
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动HTTP服务（goroutine，不阻塞）
	go func() {
		log.Printf("api-server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// 等待中断信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// 先关闭HTTP服务（最多等待30秒）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	// 关闭MongoDB连接
	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Printf("mongo disconnect error: %v", err)
	}

	log.Println("server exited")
}

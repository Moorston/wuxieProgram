package main

import (
	"context"
	"log"

	"wuxie-api/internal/config"
	"wuxie-api/internal/handler"
	"wuxie-api/internal/middleware"
	"wuxie-api/internal/repository"
	"wuxie-api/internal/router"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/jwt"

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

	// 创建索引
	ctx := context.Background()
	userRepo.EnsureIndexes(ctx)
	checkinRepo.EnsureIndexes(ctx)

	// Service
	authService := service.NewAuthService(userRepo, jwtMgr, cfg)
	userService := service.NewUserService(userRepo)
	checkinService := service.NewCheckinService(checkinRepo, userRepo, cfg.MediaURL)
	socialService := service.NewSocialService(commentRepo, likeRepo, checkinRepo, userRepo)
	rankService := service.NewRankService(rankRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)

	// Handler
	authH := handler.NewAuthHandler(authService)
	userH := handler.NewUserHandler(userService)
	checkinH := handler.NewCheckinHandler(checkinService, socialService)
	socialH := handler.NewSocialHandler(socialService)
	rankH := handler.NewRankHandler(rankService)
	groupH := handler.NewGroupHandler(groupService)

	// 路由
	r := router.Setup(authH, userH, checkinH, socialH, rankH, groupH, jwtMgr, logger)

	// 设置模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Printf("api-server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

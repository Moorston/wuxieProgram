package main

import (
	"context"
	"log"

	"wuxie-media/internal/config"
	"wuxie-media/internal/handler"
	"wuxie-media/internal/router"
	"wuxie-media/internal/worker"
	miniopkg "wuxie-media/pkg/minio"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// 初始化MinIO
	minioClient, err := miniopkg.NewClient(miniopkg.Config{
		Endpoint:  cfg.MinIO.Endpoint,
		AccessKey: cfg.MinIO.AccessKey,
		SecretKey: cfg.MinIO.SecretKey,
		UseSSL:    cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalf("failed to init minio: %v", err)
	}

	// 确保存储桶存在
	ctx := context.Background()
	if err := minioClient.EnsureBuckets(ctx, cfg.MinIO.RawBucket, cfg.MinIO.VideoBucket, cfg.MinIO.CoverBucket, "resource"); err != nil {
		log.Fatalf("failed to ensure buckets: %v", err)
	}

	// 初始化Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	// 启动转码Worker
	w := worker.NewWorker(rdb, minioClient, cfg)
	go w.Start(ctx)

	// 启动HTTP服务
	uploadH := handler.NewUploadHandler(minioClient, cfg, rdb)
	mediaH := handler.NewMediaHandler(minioClient, cfg)
	r := router.Setup(uploadH, mediaH)

	log.Printf("media-server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

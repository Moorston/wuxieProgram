package handler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"wuxie-media/internal/config"
	"wuxie-media/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	minioclient "wuxie-media/pkg/minio"
)

type UploadHandler struct {
	minioClient *minioclient.Client
	cfg         *config.Config
	rdb         *redis.Client
}

func NewUploadHandler(minioClient *minioclient.Client, cfg *config.Config, rdb *redis.Client) *UploadHandler {
	return &UploadHandler{minioClient: minioClient, cfg: cfg, rdb: rdb}
}

// Presign 获取预签名上传URL
func (h *UploadHandler) Presign(c *gin.Context) {
	checkinID := c.Query("checkin_id")
	if checkinID == "" {
		response.BadRequest(c, "checkin_id is required")
		return
	}

	ext := c.DefaultQuery("ext", "mp4")
	// 验证文件扩展名，防止路径遍历
	allowedExt := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !allowedExt.MatchString(ext) || len(ext) > 10 {
		response.BadRequest(c, "invalid file extension")
		return
	}
	objectName := fmt.Sprintf("%s/%s.%s", time.Now().Format("20060102"), uuid.New().String(), ext)

	presignedURL, err := h.minioClient.PresignPutURL(c.Request.Context(), h.cfg.MinIO.RawBucket, objectName, 15*time.Minute)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"upload_url":  presignedURL,
		"object_name": objectName,
		"bucket":      h.cfg.MinIO.RawBucket,
	})
}

type UploadCallbackReq struct {
	CheckinID  string `json:"checkin_id" binding:"required"`
	ObjectName string `json:"object_name" binding:"required"`
	Bucket     string `json:"bucket" binding:"required"`
	FileSize   int64  `json:"file_size"`
}

// UploadCallback 上传完成后回调，加入转码队列
func (h *UploadHandler) UploadCallback(c *gin.Context) {
	var req UploadCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	// 构建转码任务
	task := map[string]interface{}{
		"checkin_id":  req.CheckinID,
		"object_name": req.ObjectName,
		"bucket":      req.Bucket,
		"file_size":   req.FileSize,
		"created_at":  time.Now().Unix(),
	}

	taskJSON, _ := json.Marshal(task)

	// 推入Redis队列
	if h.rdb != nil {
		h.rdb.LPush(c.Request.Context(), "transcode:queue", taskJSON)
	}

	response.Success(c, gin.H{"status": "queued"})
}

type MediaHandler struct {
	minioClient *minioclient.Client
	cfg         *config.Config
}

func NewMediaHandler(minioClient *minioclient.Client, cfg *config.Config) *MediaHandler {
	return &MediaHandler{minioClient: minioClient, cfg: cfg}
}

// GetURL 获取视频播放URL
func (h *MediaHandler) GetURL(c *gin.Context) {
	objectName := c.Query("object")
	bucket := c.DefaultQuery("bucket", h.cfg.MinIO.VideoBucket)

	if objectName == "" {
		response.BadRequest(c, "object is required")
		return
	}

	url, err := h.minioClient.PresignGetURL(c.Request.Context(), bucket, objectName, 2*time.Hour)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"url": url})
}

// Health 健康检查
func (h *UploadHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}

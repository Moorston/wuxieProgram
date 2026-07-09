package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"wuxie-media/internal/config"
	"wuxie-media/pkg/ffmpeg"
	miniopkg "wuxie-media/pkg/minio"

	"github.com/redis/go-redis/v9"
)

type Task struct {
	CheckinID  string `json:"checkin_id"`
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	FileSize   int64  `json:"file_size"`
	CreatedAt  int64  `json:"created_at"`
}

type Worker struct {
	rdb         *redis.Client
	minioClient *miniopkg.Client
	cfg         *config.Config
	httpClient  *http.Client
}

func NewWorker(rdb *redis.Client, minioClient *miniopkg.Client, cfg *config.Config) *Worker {
	return &Worker{
		rdb:         rdb,
		minioClient: minioClient,
		cfg:         cfg,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Printf("transcode worker started, concurrency: %d", w.cfg.FFmpeg.Workers)

	for i := 0; i < w.cfg.FFmpeg.Workers; i++ {
		go w.processLoop(ctx, i)
	}
}

func (w *Worker) processLoop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 从队列取任务，阻塞等待
		result, err := w.rdb.BRPop(ctx, 5*time.Second, "transcode:queue").Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			time.Sleep(time.Second)
			continue
		}

		if len(result) < 2 {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("[worker %d] invalid task: %v", workerID, err)
			continue
		}

		log.Printf("[worker %d] processing checkin: %s", workerID, task.CheckinID)
		w.process(ctx, &task)
	}
}

func (w *Worker) process(ctx context.Context, task *Task) {
	tmpDir := filepath.Join(os.TempDir(), "wuxie-transcode", task.CheckinID)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input"+filepath.Ext(task.ObjectName))
	outputPath := filepath.Join(tmpDir, "output.mp4")
	coverPath := filepath.Join(tmpDir, "cover.jpg")

	// 1. 从MinIO下载原始文件
	if err := w.minioClient.FGetObject(ctx, task.Bucket, task.ObjectName, inputPath); err != nil {
		log.Printf("download failed: %v", err)
		w.notifyFailed(task.CheckinID, err.Error())
		return
	}

	// 2. FFmpeg转码
	duration, err := ffmpeg.Transcode(ctx, w.cfg.FFmpeg.Binary, ffmpeg.TranscodeOptions{
		InputPath:  inputPath,
		OutputPath: outputPath,
		CoverPath:  coverPath,
		MaxWidth:   1280,
	})
	if err != nil {
		log.Printf("transcode failed: %v", err)
		w.notifyFailed(task.CheckinID, err.Error())
		return
	}

	// 3. 上传转码后视频到video桶
	videoObjectName := task.ObjectName
	if err := w.minioClient.FPutObject(ctx, w.cfg.MinIO.VideoBucket, videoObjectName, outputPath, "video/mp4"); err != nil {
		log.Printf("upload video failed: %v", err)
		w.notifyFailed(task.CheckinID, err.Error())
		return
	}

	// 4. 上传封面图到cover桶
	coverObjectName := task.ObjectName[:len(task.ObjectName)-len(filepath.Ext(task.ObjectName))] + ".jpg"
	if err := w.minioClient.FPutObject(ctx, w.cfg.MinIO.CoverBucket, coverObjectName, coverPath, "image/jpeg"); err != nil {
		log.Printf("upload cover failed: %v", err)
		// 封面上传失败不影响主流程
	}

	// 5. 回调api-server
	videoURL := fmt.Sprintf("/%s/%s", w.cfg.MinIO.VideoBucket, videoObjectName)
	coverURL := fmt.Sprintf("/%s/%s", w.cfg.MinIO.CoverBucket, coverObjectName)
	w.notifySuccess(task.CheckinID, videoURL, coverURL, duration)

	log.Printf("transcode done: %s, duration: %.2fs", task.CheckinID, duration)
}

func (w *Worker) notifySuccess(checkinID, videoURL, coverURL string, duration float64) {
	payload := map[string]interface{}{
		"checkin_id": checkinID,
		"video_url":  videoURL,
		"cover_url":  coverURL,
		"duration":   duration,
	}
	w.callback(payload)
}

func (w *Worker) notifyFailed(checkinID, errMsg string) {
	payload := map[string]interface{}{
		"checkin_id": checkinID,
		"error":      errMsg,
	}
	w.callback(payload)
}

func (w *Worker) callback(payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	url := w.cfg.APIServer + "/api/internal/transcode/done"

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("callback create request failed: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", w.cfg.APISecret)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		log.Printf("callback failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("callback returned %d", resp.StatusCode)
	}
}

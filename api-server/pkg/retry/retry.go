package retry

import (
	"net/http"
	"time"
)

// Config 重试配置
type Config struct {
	MaxAttempts int           // 最大尝试次数（含首次）
	BaseDelay   time.Duration // 基础退避延迟
	IsRetryable func(error) bool // 判断是否值得重试
}

// DefaultConfig 返回默认配置（3 次尝试，500ms 基础退避）
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		IsRetryable: defaultIsRetryable,
	}
}

// Do 执行带重试的操作
// fn 返回 (result, retryable_error) 或 (result, non_retryable_error)
func Do[T any](cfg Config, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := cfg.BaseDelay * time.Duration(1<<(attempt-1))
			time.Sleep(delay)
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if cfg.IsRetryable != nil && !cfg.IsRetryable(err) {
			return zero, err
		}
	}

	return zero, lastErr
}

// DoVoid 执行带重试的操作（无返回值版本）
func DoVoid(cfg Config, fn func() error) error {
	_, err := Do(cfg, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

// HTTPClient 返回带默认超时的 HTTP 客户端
func HTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// defaultIsRetryable 默认的可重试判断：网络错误和 EOF
func defaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// 检查是否为 net.Error（超时、连接拒绝等）
	type netError interface {
		Error() string
		Timeout() bool
		Temporary() bool
	}
	if ne, ok := err.(netError); ok {
		return ne.Timeout() || ne.Temporary()
	}
	// EOF 和 unexpected EOF
	errStr := err.Error()
	if errStr == "EOF" || errStr == "unexpected EOF" {
		return true
	}
	return false
}

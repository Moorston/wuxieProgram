package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器（基于IP的简易实现）
type RateLimiter struct {
	visitors map[string]*visitor
	mu        sync.Mutex
	maxRequests int
	window      time.Duration
}

type visitor struct {
	count    int
	resetAt time.Time
}

// NewRateLimiter 创建速率限制器
// maxRequests: 时间窗口内最大请求数
// window: 时间窗口
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	if maxRequests <= 0 {
		maxRequests = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	limiter := &RateLimiter{
		visitors: make(map[string]*visitor),
		maxRequests: maxRequests,
		window:      window,
	}

	// 启动清理协程
	go limiter.cleanupLoop()

	return limiter
}

func (r *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(r.window)
		r.mu.Lock()
		now := time.Now()
		for ip, v := range r.visitors {
			if now.After(v.resetAt) {
				delete(r.visitors, ip)
			}
		}
		r.mu.Unlock()
	}
}

// Limit 返回速率限制中间件
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		r.mu.Lock()
		v, exists := r.visitors[ip]
		now := time.Now()

		if !exists || now.After(v.resetAt) {
			// 新访客或窗口过期
			r.visitors[ip] = &visitor{
				count:    1,
				resetAt: now.Add(r.window),
			}
			r.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		if v.count > r.maxRequests {
			r.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": v.resetAt.Format(time.RFC3339),
			})
			return
		}

		r.mu.Unlock()
		c.Next()
	}
}

// LoginRateLimit 登录接口专用速率限制
// 5分钟内最多5次尝试
func LoginRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(5, 5*time.Minute)
	
	return func(c *gin.Context) {
		// 只对登录接口应用严格限制
		if c.Request.URL.Path == "/api/auth/login" {
			limiter.Limit()(c)
			return
		}
		c.Next()
	}
}

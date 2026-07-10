package middleware

import (
	"sync"
	"time"
)

// TokenBlacklist 管理已撤销的 token
// 生产环境应替换为 Redis 实现以支持多实例
type TokenBlacklist struct {
	mu       sync.RWMutex
	revoked  map[string]time.Time // token -> 过期时间
	stopCh   chan struct{}
}

// NewTokenBlacklist 创建 Token 黑名单管理器
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		revoked: make(map[string]time.Time),
		stopCh:  make(chan struct{}),
	}
	go bl.cleanupLoop()
	return bl
}

// Revoke 撤销指定 token，ttl 为 token 剩余有效期
func (bl *TokenBlacklist) Revoke(tokenStr string, ttl time.Duration) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.revoked[tokenStr] = time.Now().Add(ttl)
}

// IsRevoked 检查 token 是否已被撤销
func (bl *TokenBlacklist) IsRevoked(tokenStr string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	expiry, exists := bl.revoked[tokenStr]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		// 持有写锁时直接删除，避免竞态
		delete(bl.revoked, tokenStr)
		return false
	}
	return true
}

// cleanupLoop 定期清理过期的撤销记录
func (bl *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			bl.mu.Lock()
			now := time.Now()
			for token, expiry := range bl.revoked {
				if now.After(expiry) {
					delete(bl.revoked, token)
				}
			}
			bl.mu.Unlock()
		case <-bl.stopCh:
			return
		}
	}
}

// Stop 停止清理协程
func (bl *TokenBlacklist) Stop() {
	close(bl.stopCh)
}

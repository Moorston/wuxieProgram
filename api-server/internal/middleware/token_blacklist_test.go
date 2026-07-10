package middleware

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsRevoked_NotRevoked(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	assert.False(t, bl.IsRevoked("some-token"))
}

func TestIsRevoked_Revoked(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	tokenStr := "test-token-123"
	bl.Revoke(tokenStr, 10*time.Minute)

	assert.True(t, bl.IsRevoked(tokenStr))
}

func TestIsRevoked_DifferentTokens(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	bl.Revoke("token-a", 10*time.Minute)

	assert.True(t, bl.IsRevoked("token-a"))
	assert.False(t, bl.IsRevoked("token-b"))
}

func TestIsRevoked_Expired(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	tokenStr := "expired-token"
	bl.Revoke(tokenStr, 1*time.Millisecond)

	// 等待过期
	time.Sleep(5 * time.Millisecond)

	assert.False(t, bl.IsRevoked(tokenStr))
}

func TestIsRevoked_ZeroTTL(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	tokenStr := "zero-ttl-token"
	bl.Revoke(tokenStr, 0)

	// TTL 为 0，应该已过期
	assert.False(t, bl.IsRevoked(tokenStr))
}

func TestIsRevoked_Concurrent(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	var wg sync.WaitGroup
	numGoroutines := 100

	// 并发 Revoke
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			token := fmt.Sprintf("token-%d", id)
			bl.Revoke(token, 10*time.Minute)
		}(i)
	}
	wg.Wait()

	// 并发 IsRevoked
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			token := fmt.Sprintf("token-%d", id)
			assert.True(t, bl.IsRevoked(token))
		}(i)
	}
	wg.Wait()
}

func TestIsRevoked_ConcurrentReadWrite(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	var wg sync.WaitGroup
	done := make(chan struct{})

	// 写 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				bl.Revoke(fmt.Sprintf("rw-token-%d", i), 10*time.Minute)
			}
		}
	}()

	// 读 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				bl.IsRevoked(fmt.Sprintf("rw-token-%d", i))
			}
		}
	}()

	// 运行一小段时间
	time.Sleep(50 * time.Millisecond)
	close(done)
	wg.Wait()
}

func TestRevoke_Overwrite(t *testing.T) {
	bl := NewTokenBlacklist()
	defer bl.Stop()

	tokenStr := "overwrite-token"

	// 第一次撤销，短 TTL
	bl.Revoke(tokenStr, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	assert.False(t, bl.IsRevoked(tokenStr)) // 已过期

	// 第二次撤销，长 TTL
	bl.Revoke(tokenStr, 10*time.Minute)
	assert.True(t, bl.IsRevoked(tokenStr)) // 仍然有效
}

func TestStop(t *testing.T) {
	bl := NewTokenBlacklist()
	bl.Revoke("token", 10*time.Minute)

	// Stop 不应该 panic
	bl.Stop()

	// 再次 Stop 应该 panic（channel 已关闭）
	// 这是预期行为，不测试
}

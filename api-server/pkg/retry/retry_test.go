package retry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig 返回测试用配置（极短延迟）
func testConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   1, // 1 nanosecond — 几乎无延迟
		IsRetryable: defaultIsRetryable,
	}
}

func TestDo_SuccessFirstAttempt(t *testing.T) {
	cfg := testConfig()
	attempts := 0

	result, err := Do(cfg, func() (string, error) {
		attempts++
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 1, attempts) // 只尝试一次
}

func TestDo_RetryThenSuccess(t *testing.T) {
	cfg := testConfig()
	attempts := 0

	result, err := Do(cfg, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", &tempNetError{msg: "connection refused"}
		}
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 3, attempts)
}

func TestDo_NonRetryableError(t *testing.T) {
	cfg := testConfig()
	attempts := 0

	_, err := Do(cfg, func() (string, error) {
		attempts++
		return "", fmt.Errorf("business logic error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "business logic error")
	assert.Equal(t, 1, attempts) // 不重试
}

func TestDo_AllRetriesFail(t *testing.T) {
	cfg := testConfig()
	attempts := 0

	_, err := Do(cfg, func() (string, error) {
		attempts++
		return "", &tempNetError{msg: "timeout"}
	})

	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
}

func TestDo_NilIsRetryable(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		BaseDelay:   1,
		IsRetryable: nil, // nil = 不检查，总是重试
	}
	attempts := 0

	_, err := Do(cfg, func() (string, error) {
		attempts++
		return "", fmt.Errorf("any error")
	})

	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // 全部重试
}

func TestDoVoid_Success(t *testing.T) {
	cfg := testConfig()
	err := DoVoid(cfg, func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestDoVoid_Error(t *testing.T) {
	cfg := testConfig()
	err := DoVoid(cfg, func() error {
		return fmt.Errorf("failed")
	})
	assert.Error(t, err)
}

func TestDo_CustomIsRetryable(t *testing.T) {
	cfg := Config{
		MaxAttempts: 3,
		BaseDelay:   1,
		IsRetryable: func(err error) bool {
			return err.Error() == "retry-me" // 只重试特定错误
		},
	}
	attempts := 0

	// 可重试错误 → 重试
	_, err := Do(cfg, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", fmt.Errorf("retry-me")
		}
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, attempts)

	// 不可重试错误 → 不重试
	attempts = 0
	_, err = Do(cfg, func() (string, error) {
		attempts++
		return "", fmt.Errorf("do-not-retry")
	})
	assert.Error(t, err)
	assert.Equal(t, 1, attempts)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.NotNil(t, cfg.IsRetryable)
}

// --- net.Error mock ---

type tempNetError struct{ msg string }

func (e *tempNetError) Error() string   { return e.msg }
func (e *tempNetError) Timeout() bool   { return true }
func (e *tempNetError) Temporary() bool { return true }

package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	"wuxie-api/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// --- Mock HTTP Transport ---

type responseEntry struct {
	resp *http.Response
	err  error
}

type mockTransport struct {
	mu        sync.Mutex
	responses []responseEntry
	callCount int
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.callCount >= len(t.responses) {
		return nil, fmt.Errorf("mockTransport: no more responses (call %d)", t.callCount)
	}
	entry := t.responses[t.callCount]
	t.callCount++
	return entry.resp, entry.err
}

func (t *mockTransport) CallCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.callCount
}

func jsonResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// net.Error mock for testing network retries
type tempNetError struct {
	msg string
}

func (e *tempNetError) Error() string   { return e.msg }
func (e *tempNetError) Timeout() bool   { return true }
func (e *tempNetError) Temporary() bool { return true }

// --- Mock UserRepo ---

type mockUserRepo struct {
	upsertFn       func(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error)
	findByID       func(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	findByIDs      func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error)
	findByGroupID  func(ctx context.Context, groupID primitive.ObjectID) ([]*model.User, error)
	update         func(ctx context.Context, id primitive.ObjectID, update bson.M) error
	incrScore      func(ctx context.Context, id primitive.ObjectID, score int) error
	findTopByScore func(ctx context.Context, limit int) ([]*model.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error { return nil }
func (m *mockUserRepo) FindByOpenID(ctx context.Context, openid string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	if m.findByID != nil {
		return m.findByID(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockUserRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
	if m.findByIDs != nil {
		return m.findByIDs(ctx, ids)
	}
	return nil, nil
}
func (m *mockUserRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]*model.User, error) {
	if m.findByGroupID != nil {
		return m.findByGroupID(ctx, groupID)
	}
	return nil, nil
}
func (m *mockUserRepo) FindTopByScore(ctx context.Context, limit int) ([]*model.User, error) {
	if m.findTopByScore != nil {
		return m.findTopByScore(ctx, limit)
	}
	return nil, nil
}
func (m *mockUserRepo) UpsertByOpenID(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, openid, nickname, avatar, gender)
	}
	return nil, false, fmt.Errorf("not implemented")
}
func (m *mockUserRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if m.update != nil {
		return m.update(ctx, id, update)
	}
	return nil
}
func (m *mockUserRepo) IncrScore(ctx context.Context, id primitive.ObjectID, score int) error {
	if m.incrScore != nil {
		return m.incrScore(ctx, id, score)
	}
	return nil
}
func (m *mockUserRepo) IsBanned(ctx context.Context, id primitive.ObjectID) (bool, error) {
	return false, nil
}

// --- Helper ---

func testConfig() *config.Config {
	return &config.Config{
		WX: config.WXConfig{
			AppID:  "test-appid",
			Secret: "test-secret",
		},
	}
}

func testLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// --- WXLogin Tests ---

func TestWXLogin_NewUser(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &mockUserRepo{
		upsertFn: func(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error) {
			return &model.User{
				ID:       userID,
				OpenID:   openid,
				Nickname: nickname,
				Avatar:   avatar,
				Gender:   gender,
			}, true, nil
		},
	}

	transport := &mockTransport{
		responses: []responseEntry{
			{resp: jsonResponse(200, `{"openid":"wx_openid_123"}`)},
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	result, err := svc.WXLogin(context.Background(), "wx_code", "TestUser", "http://avatar.url", 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, userID, result.User.ID)
	assert.Equal(t, "TestUser", result.User.Nickname)
}

func TestWXLogin_ExistingUser(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &mockUserRepo{
		upsertFn: func(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error) {
			return &model.User{
				ID:       userID,
				OpenID:   openid,
				Nickname: nickname,
				Avatar:   avatar,
				Gender:   gender,
			}, false, nil
		},
	}

	transport := &mockTransport{
		responses: []responseEntry{
			{resp: jsonResponse(200, `{"openid":"wx_openid_456"}`)},
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	result, err := svc.WXLogin(context.Background(), "wx_code", "NewName", "http://new.avatar", 2)
	require.NoError(t, err)
	assert.Equal(t, "NewName", result.User.Nickname)
}

func TestWXLogin_WXAPIError(t *testing.T) {
	repo := &mockUserRepo{}
	transport := &mockTransport{
		responses: []responseEntry{
			{resp: jsonResponse(200, `{"errcode":40029,"errmsg":"invalid code"}`)},
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	_, err := svc.WXLogin(context.Background(), "bad_code", "User", "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "40029")
	// 微信业务错误不应重试
	assert.Equal(t, 1, transport.CallCount())
}

func TestWXLogin_NetworkRetry(t *testing.T) {
	// 注意：此测试因 getOpenID 的退避 sleep 需要 ~1.5 秒
	// TODO: 后续可注入 sleep 函数以加速测试
	repo := &mockUserRepo{
		upsertFn: func(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error) {
			return &model.User{ID: primitive.NewObjectID(), OpenID: openid}, true, nil
		},
	}

	transport := &mockTransport{
		responses: []responseEntry{
			{err: &tempNetError{msg: "connection refused"}},  // attempt 1: retry
			{err: &tempNetError{msg: "connection timeout"}},  // attempt 2: retry
			{resp: jsonResponse(200, `{"openid":"wx_retry_ok"}`)}, // attempt 3: success
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	result, err := svc.WXLogin(context.Background(), "code", "User", "", 0)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, 3, transport.CallCount())
}

func TestWXLogin_AllRetriesFail(t *testing.T) {
	repo := &mockUserRepo{}
	transport := &mockTransport{
		responses: []responseEntry{
			{err: &tempNetError{msg: "error 1"}},
			{err: &tempNetError{msg: "error 2"}},
			{err: &tempNetError{msg: "error 3"}},
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	_, err := svc.WXLogin(context.Background(), "code", "User", "", 0)
	assert.Error(t, err)
	assert.Equal(t, 3, transport.CallCount())
}

func TestWXLogin_NonNetworkError_NoRetry(t *testing.T) {
	repo := &mockUserRepo{}
	// 非 net.Error（如 context.Canceled）不应重试
	transport := &mockTransport{
		responses: []responseEntry{
			{err: context.Canceled},
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger(), client)

	_, err := svc.WXLogin(context.Background(), "code", "User", "", 0)
	assert.Error(t, err)
	assert.Equal(t, 1, transport.CallCount()) // 不重试
}

// --- RefreshToken Tests ---

func TestRefreshToken_Valid(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return &model.User{ID: id, Status: 0}, nil
		},
	}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger())

	refreshStr, err := jwtMgr.GenerateRefreshToken(userID.Hex())
	require.NoError(t, err)

	newToken, newRefresh, err := svc.RefreshToken(context.Background(), refreshStr)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, refreshStr, newRefresh)
}

func TestRefreshToken_InvalidSignature(t *testing.T) {
	repo := &mockUserRepo{}
	mgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, mgr, testConfig(), testLogger())

	// 用不同密钥生成的 token 无法解析
	otherMgr := jwt.NewJWTManager("different-secret-key-at-least-32-char", 1)
	refreshStr, err := otherMgr.GenerateRefreshToken("user123")
	require.NoError(t, err)

	_, _, err = svc.RefreshToken(context.Background(), refreshStr)
	assert.Error(t, err)
}

func TestRefreshToken_UserBanned(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return &model.User{ID: id, Status: 1}, nil // 封禁
		},
	}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger())

	refreshStr, err := jwtMgr.GenerateRefreshToken(userID.Hex())
	require.NoError(t, err)

	_, _, err = svc.RefreshToken(context.Background(), refreshStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suspended")
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}

	jwtMgr := jwt.NewJWTManager("test-secret-key-at-least-32-chars-long", 1)
	svc := NewAuthService(repo, jwtMgr, testConfig(), testLogger())

	refreshStr, err := jwtMgr.GenerateRefreshToken(userID.Hex())
	require.NoError(t, err)

	_, _, err = svc.RefreshToken(context.Background(), refreshStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// --- isNetworkError Tests ---

func TestIsNetworkError(t *testing.T) {
	assert.True(t, isNetworkError(&tempNetError{msg: "timeout"}))
	assert.True(t, isNetworkError(io.EOF))
	assert.True(t, isNetworkError(io.ErrUnexpectedEOF))
	assert.False(t, isNetworkError(context.Canceled))
	assert.False(t, isNetworkError(fmt.Errorf("regular error")))
}

// 确保 mockUserRepo 实现了 UserRepoInterface
var _ repository.UserRepoInterface = (*mockUserRepo)(nil)

// unused suppressor for net import
var _ net.Error = (*tempNetError)(nil)

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	apperrors "wuxie-api/pkg/errors"
	"wuxie-api/pkg/jwt"
	"wuxie-api/pkg/retry"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AuthService struct {
	userRepo   repository.UserRepoInterface
	jwtMgr     *jwt.JWTManager
	cfg        *config.Config
	logger     *zap.Logger
	httpClient *http.Client
}

func NewAuthService(userRepo repository.UserRepoInterface, jwtMgr *jwt.JWTManager, cfg *config.Config, logger *zap.Logger, httpClient ...*http.Client) *AuthService {
	client := http.DefaultClient
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}
	return &AuthService{userRepo: userRepo, jwtMgr: jwtMgr, cfg: cfg, logger: logger, httpClient: client}
}

// LoginResult 登录结果
type LoginResult struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) WXLogin(ctx context.Context, code string, nickname, avatar string, gender int) (*LoginResult, error) {
	openid, err := s.getOpenID(code)
	if err != nil {
		return nil, fmt.Errorf("get openid failed: %w", err)
	}

	// 原子化：查找或创建用户，同时更新资料
	user, isCreated, err := s.userRepo.UpsertByOpenID(ctx, openid, nickname, avatar, gender)
	if err != nil {
		return nil, fmt.Errorf("upsert user failed: %w", err)
	}

	if isCreated {
		s.logger.Info("new user registered",
			zap.String("user_id", user.ID.Hex()),
			zap.String("nickname", nickname),
		)
	}

	token, err := s.jwtMgr.Generate(user.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	refreshToken, err := s.jwtMgr.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("generate refresh token failed: %w", err)
	}

	return &LoginResult{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshToken 使用 refresh token 获取新的 access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	claims, err := s.jwtMgr.ParseRefreshToken(refreshTokenStr)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apperrors.ErrInvalidRefresh, err)
	}

	// 验证用户仍然存在且未被封禁
	oid, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return "", "", apperrors.ErrInvalidUserID
	}
	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil {
		return "", "", apperrors.ErrUserNotFound
	}
	if user.IsBanned() {
		return "", "", apperrors.ErrAccountSuspended
	}

	// 生成新的双 token
	newToken, err := s.jwtMgr.Generate(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("generate token failed: %w", err)
	}
	newRefreshToken, err := s.jwtMgr.GenerateRefreshToken(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token failed: %w", err)
	}

	return newToken, newRefreshToken, nil
}

// wxAPIError 微信 API 业务错误
type wxAPIError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *wxAPIError) Error() string {
	return fmt.Sprintf("wx api error: %d %s", e.ErrCode, e.ErrMsg)
}

func (s *AuthService) getOpenID(code string) (string, error) {
	wxURL := "https://api.weixin.qq.com/sns/jscode2session"
	form := fmt.Sprintf("appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.cfg.WX.AppID, s.cfg.WX.Secret, code)

	cfg := retry.Config{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		IsRetryable: isNetworkError,
	}

	openid, err := retry.Do(cfg, func() (string, error) {
		resp, err := s.httpClient.Post(wxURL, "application/x-www-form-urlencoded", strings.NewReader(form))
		if err != nil {
			return "", fmt.Errorf("wx api request failed: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read wx api response failed: %w", err)
		}

		var result struct {
			wxAPIError
			OpenID string `json:"openid"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", fmt.Errorf("parse wx api response failed: %w", err)
		}

		if result.ErrCode != 0 {
			return "", &result.wxAPIError
		}

		return result.OpenID, nil
	})

	return openid, err
}

// isNetworkError 判断是否为网络层错误（值得重试）
func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	return false
}
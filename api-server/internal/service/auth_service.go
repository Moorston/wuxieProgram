package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	"wuxie-api/pkg/jwt"
)

type AuthService struct {
	userRepo *repository.UserRepo
	jwtMgr   *jwt.JWTManager
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepo, jwtMgr *jwt.JWTManager, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, jwtMgr: jwtMgr, cfg: cfg}
}

func (s *AuthService) WXLogin(ctx context.Context, code string, nickname, avatar string, gender int) (string, *model.User, error) {
	openid, err := s.getOpenID(code)
	if err != nil {
		return "", nil, fmt.Errorf("get openid failed: %w", err)
	}

	user, err := s.userRepo.FindByOpenID(ctx, openid)
	if err != nil {
		user = &model.User{
			OpenID:   openid,
			Nickname: nickname,
			Avatar:   avatar,
			Gender:   gender,
			Score:    0,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return "", nil, fmt.Errorf("create user failed: %w", err)
		}
	}

	token, err := s.jwtMgr.Generate(user.ID.Hex())
	if err != nil {
		return "", nil, fmt.Errorf("generate token failed: %w", err)
	}

	return token, user, nil
}

func (s *AuthService) getOpenID(code string) (string, error) {
	wxURL := "https://api.weixin.qq.com/sns/jscode2session"

	client := &http.Client{Timeout: 10 * time.Second}
	form := fmt.Sprintf("appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.cfg.WX.AppID, s.cfg.WX.Secret, code)

	resp, err := client.Post(wxURL, "application/x-www-form-urlencoded", strings.NewReader(form))
	if err != nil {
		return "", fmt.Errorf("wx api request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read wx api response failed: %w", err)
	}
	var result struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse wx api response failed: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wx api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return result.OpenID, nil
}
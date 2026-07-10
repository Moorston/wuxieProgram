package handler

import (
	"strings"
	"time"

	"wuxie-api/internal/middleware"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/jwt"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtMgr      *jwt.JWTManager
	blacklist   *middleware.TokenBlacklist
	logger      *zap.Logger
}

func NewAuthHandler(authService *service.AuthService, jwtMgr *jwt.JWTManager, blacklist *middleware.TokenBlacklist, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, jwtMgr: jwtMgr, blacklist: blacklist, logger: logger}
}

type LoginReq struct {
	Code     string `json:"code" binding:"required,max=256"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int    `json:"gender"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	result, err := h.authService.WXLogin(c.Request.Context(), req.Code, req.Nickname, req.Avatar, req.Gender)
	if err != nil {
		h.logger.Error("login failed",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	// 登录成功审计日志
	h.logger.Info("login success",
		zap.String("user_id", result.User.ID.Hex()),
		zap.String("nickname", result.User.Nickname),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	)

	response.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		response.Success(c, nil)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// 解析 token 获取剩余有效期
	claims, err := h.jwtMgr.Parse(tokenStr)
	if err != nil {
		// token 已无效，无需撤销
		response.Success(c, nil)
		return
	}

	// 计算 token 剩余有效期
	var ttl time.Duration
	if claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
	}
	if ttl > 0 {
		h.blacklist.Revoke(tokenStr, ttl)
	}

	h.logger.Info("user logged out",
		zap.String("user_id", claims.UserID),
		zap.String("client_ip", c.ClientIP()),
	)

	response.Success(c, nil)
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	newToken, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Warn("refresh token failed",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		respondWithError(c, err)
		return
	}

	response.Success(c, gin.H{
		"token":         newToken,
		"refresh_token": newRefreshToken,
	})
}
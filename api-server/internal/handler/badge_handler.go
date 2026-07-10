package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BadgeHandler struct {
	badgeService *service.BadgeService
	logger       *zap.Logger
}

func NewBadgeHandler(badgeService *service.BadgeService, logger *zap.Logger) *BadgeHandler {
	return &BadgeHandler{badgeService: badgeService, logger: logger}
}

// GetAllBadges 获取所有徽章定义
func (h *BadgeHandler) GetAllBadges(c *gin.Context) {
	badges, err := h.badgeService.GetAllBadges(c.Request.Context())
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"badges": badges})
}

// GetUserBadges 获取当前用户的徽章
func (h *BadgeHandler) GetUserBadges(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	badges, err := h.badgeService.GetUserBadges(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"badges": badges})
}

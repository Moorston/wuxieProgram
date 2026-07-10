package handler

import (
	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), oid)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}

type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), oid, req.Nickname, req.Avatar); err != nil {
		h.logger.Error("update profile failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// GetPrivacySettings 获取隐私设置
func (h *UserHandler) GetPrivacySettings(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	visibility, err := h.userService.GetPrivacySettings(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"default_visibility": visibility})
}

type UpdateVisibilityReq struct {
	Visibility model.Visibility `json:"visibility"`
}

// UpdatePrivacySettings 更新隐私设置
func (h *UserHandler) UpdatePrivacySettings(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req UpdateVisibilityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.userService.UpdateDefaultVisibility(c.Request.Context(), oid, req.Visibility); err != nil {
		if err == service.ErrInvalidVisibility {
			response.BadRequest(c, "invalid visibility value")
			return
		}
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}
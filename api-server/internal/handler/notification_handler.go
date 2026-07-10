package handler

import (
	"strconv"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	notifService *service.NotificationService
	logger       *zap.Logger
}

func NewNotificationHandler(notifService *service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{notifService: notifService, logger: logger}
}

func (h *NotificationHandler) List(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	notifications, total, err := h.notifService.List(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		h.logger.Error("list notifications failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  notifications,
		"total": total,
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	count, err := h.notifService.UnreadCount(c.Request.Context(), oid)
	if err != nil {
		h.logger.Error("get unread count failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.notifService.MarkRead(c.Request.Context(), oid, id); err != nil {
		h.logger.Error("mark notification read failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	if err := h.notifService.MarkAllRead(c.Request.Context(), oid); err != nil {
		h.logger.Error("mark all notifications read failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.notifService.Delete(c.Request.Context(), oid, id); err != nil {
		h.logger.Error("delete notification failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) GetSettings(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	settings, err := h.notifService.GetSettings(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, settings)
}

func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.notifService.UpdateSettings(c.Request.Context(), oid, req); err != nil {
		h.logger.Error("update notification settings failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}
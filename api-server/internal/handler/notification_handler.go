package handler

import (
	"strconv"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, _ := primitive.ObjectIDFromHex(userID)

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
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  notifications,
		"total": total,
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, _ := primitive.ObjectIDFromHex(userID)

	count, err := h.notifService.UnreadCount(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.notifService.MarkRead(c.Request.Context(), id, uid); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	if err := h.notifService.MarkAllRead(c.Request.Context(), uid); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.notifService.Delete(c.Request.Context(), id, uid); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) GetSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	settings, err := h.notifService.GetSettings(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, settings)
}

type UpdateSettingsReq struct {
	LikeNotify     *bool   `json:"like_notify"`
	CommentNotify  *bool   `json:"comment_notify"`
	PlanRemind     *bool   `json:"plan_remind"`
	PlanRemindTime *string `json:"plan_remind_time"`
	GroupNotify    *bool   `json:"group_notify"`
}

func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	var req UpdateSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.notifService.UpdateSettings(c.Request.Context(), uid,
		req.LikeNotify, req.CommentNotify, req.PlanRemind, req.GroupNotify, req.PlanRemindTime,
	); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

package handler

import (
	"log"
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
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
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
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.notifService.MarkRead(c.Request.Context(), id, uid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	if err := h.notifService.MarkAllRead(c.Request.Context(), uid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.notifService.Delete(c.Request.Context(), id, uid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *NotificationHandler) GetSettings(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	settings, err := h.notifService.GetSettings(c.Request.Context(), uid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
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
	uid, ok := getUserID(c)
	if !ok {
		return
	}

	var req UpdateSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.notifService.UpdateSettings(c.Request.Context(), uid,
		req.LikeNotify, req.CommentNotify, req.PlanRemind, req.GroupNotify, req.PlanRemindTime,
	); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

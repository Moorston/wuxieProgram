package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type SocialHandler struct {
	socialService *service.SocialService
	logger        *zap.Logger
}

func NewSocialHandler(socialService *service.SocialService, logger *zap.Logger) *SocialHandler {
	return &SocialHandler{socialService: socialService, logger: logger}
}

func (h *SocialHandler) ToggleLike(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	liked, err := h.socialService.ToggleLike(c.Request.Context(), checkinID, oid)
	if err != nil {
		h.logger.Error("toggle like failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"liked": liked})
}

type CommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

func (h *SocialHandler) AddComment(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	var req CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "content is required")
		return
	}

	comment, err := h.socialService.AddComment(c.Request.Context(), checkinID, oid, req.Content)
	if err != nil {
		h.logger.Error("add comment failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, comment)
}

func (h *SocialHandler) GetComments(c *gin.Context) {
	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	page, pageSize := parsePagination(c, 20)

	comments, total, err := h.socialService.GetComments(c.Request.Context(), checkinID, page, pageSize)
	if err != nil {
		h.logger.Error("get comments failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  comments,
		"total": total,
	})
}
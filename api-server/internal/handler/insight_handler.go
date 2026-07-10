package handler

import (
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type InsightHandler struct {
	insightService *service.InsightService
	logger         *zap.Logger
}

func NewInsightHandler(insightService *service.InsightService, logger *zap.Logger) *InsightHandler {
	return &InsightHandler{insightService: insightService, logger: logger}
}

type CreateInsightReq struct {
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images"`
	Mood       string   `json:"mood"`
	Tags       []string `json:"tags"`
	IsPublic   bool     `json:"is_public"`
}

type UpdateInsightReq struct {
	Content  string   `json:"content"`
	Images   []string `json:"images"`
	Mood     string   `json:"mood"`
	Tags     []string `json:"tags"`
	IsPublic *bool    `json:"is_public"`
}

func (h *InsightHandler) Create(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateInsightReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	insight, err := h.insightService.Create(c.Request.Context(), oid, &model.Insight{
		Content:  req.Content,
		Images:   req.Images,
		Mood:     req.Mood,
		Tags:     req.Tags,
		IsPublic: req.IsPublic,
	})

	if err != nil {
		h.logger.Error("create insight failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"id": insight.ID.Hex()})
}

func (h *InsightHandler) GetByID(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	insight, err := h.insightService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "insight not found")
		return
	}

	response.Success(c, insight)
}

func (h *InsightHandler) List(c *gin.Context) {
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

	tag := c.Query("tag")
	mood := c.Query("mood")

	insights, total, err := h.insightService.List(c.Request.Context(), oid, tag, mood, page, pageSize)
	if err != nil {
		h.logger.Error("list insights failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": insights, "total": total})
}

func (h *InsightHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	insights, total, err := h.insightService.ListPublic(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("list public insights failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": insights, "total": total})
}

func (h *InsightHandler) Update(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req UpdateInsightReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.insightService.Update(c.Request.Context(), oid, id, req.Tags, map[string]interface{}{
		"content":   req.Content,
		"images":    req.Images,
		"mood":      req.Mood,
		"tags":      req.Tags,
		"is_public": req.IsPublic,
	}); err != nil {
		h.logger.Error("update insight failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *InsightHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.insightService.Delete(c.Request.Context(), oid, id); err != nil {
		h.logger.Error("delete insight failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *InsightHandler) GetTags(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	tags, err := h.insightService.GetTags(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, tags)
}

func (h *InsightHandler) MoodStats(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	stats, err := h.insightService.MoodStats(c.Request.Context(), oid, days)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, stats)
}

func (h *InsightHandler) OnThisDay(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	insights, err := h.insightService.OnThisDay(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, insights)
}

func (h *InsightHandler) Like(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	liked, err := h.insightService.Like(c.Request.Context(), oid, id)
	if err != nil {
		h.logger.Error("toggle insight like failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"liked": liked})
}
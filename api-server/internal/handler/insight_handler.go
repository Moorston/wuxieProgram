package handler

import (
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InsightHandler struct {
	insightService *service.InsightService
}

func NewInsightHandler(insightService *service.InsightService) *InsightHandler {
	return &InsightHandler{insightService: insightService}
}

type CreateInsightReq struct {
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images"`
	Mood       string   `json:"mood"`
	Tags       []string `json:"tags"`
	CheckinID  string   `json:"checkin_id"`
	PlanID     string   `json:"plan_id"`
	PlanDay    int      `json:"plan_day"`
	Visibility string   `json:"visibility"`
}

func (h *InsightHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req CreateInsightReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	insight := &model.Insight{
		Content:    req.Content,
		Images:     req.Images,
		Mood:       model.MoodType(req.Mood),
		Tags:       req.Tags,
		Visibility: model.Visibility(req.Visibility),
		PlanDay:    req.PlanDay,
	}

	if req.CheckinID != "" {
		if id, err := primitive.ObjectIDFromHex(req.CheckinID); err == nil {
			insight.CheckinID = id
		}
	}
	if req.PlanID != "" {
		if id, err := primitive.ObjectIDFromHex(req.PlanID); err == nil {
			insight.PlanID = id
		}
	}

	if err := h.insightService.Create(c.Request.Context(), oid, insight); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, insight)
}

func (h *InsightHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid insight id")
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
	userID := c.GetString("user_id")
	oid, _ := primitive.ObjectIDFromHex(userID)

	tag := c.Query("tag")
	mood := c.Query("mood")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	insights, total, err := h.insightService.List(c.Request.Context(), oid, tag, mood, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  insights,
		"total": total,
	})
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
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  insights,
		"total": total,
	})
}

type UpdateInsightReq struct {
	Content    string   `json:"content"`
	Images     []string `json:"images"`
	Mood       string   `json:"mood"`
	Tags       []string `json:"tags"`
	Visibility string   `json:"visibility"`
}

func (h *InsightHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid insight id")
		return
	}

	var req UpdateInsightReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	update := map[string]interface{}{}
	if req.Content != "" {
		update["content"] = req.Content
	}
	if req.Images != nil {
		update["images"] = req.Images
	}
	if req.Mood != "" {
		update["mood"] = req.Mood
	}
	if req.Tags != nil {
		update["tags"] = req.Tags
	}
	if req.Visibility != "" {
		update["visibility"] = req.Visibility
	}

	if err := h.insightService.Update(c.Request.Context(), id, uid, update); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *InsightHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid insight id")
		return
	}

	if err := h.insightService.Delete(c.Request.Context(), id, uid); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *InsightHandler) GetTags(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	tags, err := h.insightService.GetTags(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tags)
}

func (h *InsightHandler) MoodStats(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 {
		days = 30
	}

	stats, err := h.insightService.MoodStats(c.Request.Context(), uid, days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

func (h *InsightHandler) OnThisDay(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	insights, err := h.insightService.OnThisDay(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, insights)
}

func (h *InsightHandler) Like(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid insight id")
		return
	}

	if err := h.insightService.Like(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

package handler

import (
	"strconv"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type TrainingHandler struct {
	trainingService *service.TrainingService
	logger          *zap.Logger
}

func NewTrainingHandler(trainingService *service.TrainingService, logger *zap.Logger) *TrainingHandler {
	return &TrainingHandler{trainingService: trainingService, logger: logger}
}

type CreatePlanReq struct {
	GroupID     string           `json:"group_id"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Days        []model.TrainingDay `json:"days" binding:"required"`
	StartDate   string           `json:"start_date"`
}

func (h *TrainingHandler) CreatePlan(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreatePlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	plan := &model.TrainingPlan{
		UserID:      oid,
		Title:       req.Title,
		Description: req.Description,
		Days:        req.Days,
	}

	if req.GroupID != "" {
		if gid, err := primitive.ObjectIDFromHex(req.GroupID); err == nil {
			plan.GroupID = gid
		}
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			plan.StartDate = t
		}
	}

	if err := h.trainingService.CreatePlan(c.Request.Context(), oid, plan); err != nil {
		h.logger.Error("create training plan failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"id": plan.ID.Hex()})
}

func (h *TrainingHandler) ListPlans(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var status *model.PlanStatus
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			st := model.PlanStatus(v)
			status = &st
		}
	}

	plans, total, err := h.trainingService.ListPlans(c.Request.Context(), oid, page, pageSize, status)
	if err != nil {
		h.logger.Error("list training plans failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": plans, "total": total})
}

func (h *TrainingHandler) GetPlan(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	plan, err := h.trainingService.GetPlan(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "plan not found")
		return
	}

	response.Success(c, plan)
}

func (h *TrainingHandler) UpdatePlan(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.trainingService.UpdatePlan(c.Request.Context(), oid, id, req); err != nil {
		h.logger.Error("update training plan failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *TrainingHandler) DeletePlan(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.trainingService.DeletePlan(c.Request.Context(), oid, id); err != nil {
		h.logger.Error("delete training plan failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *TrainingHandler) TodayTasks(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	plans, err := h.trainingService.TodayTasks(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, plans)
}

func (h *TrainingHandler) UpdateTask(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	planID, err := primitive.ObjectIDFromHex(c.Param("plan_id"))
	if err != nil {
		response.BadRequest(c, "invalid plan_id")
		return
	}

	day, err := strconv.Atoi(c.Param("day"))
	if err != nil {
		response.BadRequest(c, "invalid day")
		return
	}

	taskIdx, err := strconv.Atoi(c.Param("task_idx"))
	if err != nil {
		response.BadRequest(c, "invalid task_idx")
		return
	}

	var req struct {
		Status    int    `json:"status"`
		CheckinID string `json:"checkin_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.trainingService.UpdateTask(c.Request.Context(), oid, planID, day, taskIdx, req.Status, req.CheckinID); err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *TrainingHandler) GetReport(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	report, err := h.trainingService.GetReport(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "plan not found")
		return
	}

	response.Success(c, report)
}

func (h *TrainingHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	templates, total, err := h.trainingService.ListTemplates(c.Request.Context(), c.Query("category"), c.Query("style"), page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": templates, "total": total})
}

func (h *TrainingHandler) GetTemplate(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	template, err := h.trainingService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "template not found")
		return
	}

	response.Success(c, template)
}

func (h *TrainingHandler) ApplyTemplate(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req struct {
		StartDate string `json:"start_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)

	plan, err := h.trainingService.ApplyTemplate(c.Request.Context(), oid, id, startDate)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"id": plan.ID.Hex()})
}
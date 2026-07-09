package handler

import (
	"log"
	"strconv"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingHandler struct {
	trainingService *service.TrainingService
}

func NewTrainingHandler(trainingService *service.TrainingService) *TrainingHandler {
	return &TrainingHandler{trainingService: trainingService}
}

type CreatePlanReq struct {
	GroupID     string           `json:"group_id"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	StartDate   string           `json:"start_date" binding:"required"`
	EndDate     string           `json:"end_date" binding:"required"`
	Days        []model.TrainingDay `json:"days"`
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

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(c, "invalid start_date format, use 2006-01-02")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.BadRequest(c, "invalid end_date format, use 2006-01-02")
		return
	}

	plan := &model.TrainingPlan{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Days:        req.Days,
	}

	if req.GroupID != "" {
		if id, err := primitive.ObjectIDFromHex(req.GroupID); err == nil {
			plan.GroupID = id
		}
	}

	if err := h.trainingService.CreatePlan(c.Request.Context(), oid, plan); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, plan)
}

func (h *TrainingHandler) GetPlan(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	plan, err := h.trainingService.GetPlan(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "plan not found")
		return
	}

	if plan.UserID != oid {
		response.Forbidden(c, "no access")
		return
	}

	response.Success(c, plan)
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
		v := model.PlanStatus(0)
		switch s {
		case "0":
			v = model.PlanStatusDraft
		case "1":
			v = model.PlanStatusActive
		case "2":
			v = model.PlanStatusCompleted
		case "3":
			v = model.PlanStatusTerminated
		}
		status = &v
	}

	plans, total, err := h.trainingService.ListPlans(c.Request.Context(), oid, status, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  plans,
		"total": total,
	})
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

	plan, err := h.trainingService.GetPlan(c.Request.Context(), id)
	if err != nil || plan.UserID != oid {
		response.Forbidden(c, "no access")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	allowed := map[string]bool{"title": true, "description": true, "status": true}
	update := map[string]interface{}{}
	for k, v := range req {
		if allowed[k] {
			update[k] = v
		}
	}

	if err := h.trainingService.UpdatePlan(c.Request.Context(), id, oid, update); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
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

	if err := h.trainingService.DeletePlan(c.Request.Context(), id, oid); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
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

	tasks, err := h.trainingService.GetTodayTasks(c.Request.Context(), oid)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, tasks)
}

type UpdateTaskReq struct {
	Status    int    `json:"status"`
	CheckinID string `json:"checkin_id"`
}

func (h *TrainingHandler) UpdateTask(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	planID, ok := getObjectID(c, "plan_id")
	if !ok {
		return
	}

	plan, err := h.trainingService.GetPlan(c.Request.Context(), planID)
	if err != nil || plan.UserID != oid {
		response.Forbidden(c, "no access")
		return
	}

	dayIndex, _ := strconv.Atoi(c.Param("day"))
	taskIndex, _ := strconv.Atoi(c.Param("task_idx"))

	// 校验day/task索引范围
	if dayIndex < 0 || dayIndex >= len(plan.Days) {
		response.BadRequest(c, "invalid day index")
		return
	}
	if taskIndex < 0 || taskIndex >= len(plan.Days[dayIndex].Tasks) {
		response.BadRequest(c, "invalid task index")
		return
	}

	var req UpdateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	var checkinID *primitive.ObjectID
	if req.CheckinID != "" {
		cid, err := primitive.ObjectIDFromHex(req.CheckinID)
		if err == nil {
			checkinID = &cid
		}
	}

	status := model.TaskStatus(req.Status)
	if err := h.trainingService.UpdateTaskStatus(c.Request.Context(), planID, dayIndex, taskIndex, status, checkinID); err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *TrainingHandler) GetReport(c *gin.Context) {
	planID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	report, err := h.trainingService.GetReport(c.Request.Context(), planID)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, report)
}

func (h *TrainingHandler) ListTemplates(c *gin.Context) {
	category := c.Query("category")
	style := c.Query("style")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	templates, total, err := h.trainingService.ListTemplates(c.Request.Context(), category, style, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  templates,
		"total": total,
	})
}

func (h *TrainingHandler) GetTemplate(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	t, err := h.trainingService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "template not found")
		return
	}

	response.Success(c, t)
}

type ApplyTemplateReq struct {
	StartDate string `json:"start_date" binding:"required"`
}

func (h *TrainingHandler) ApplyTemplate(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	templateID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req ApplyTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(c, "invalid start_date format, use 2006-01-02")
		return
	}

	plan, err := h.trainingService.ApplyTemplate(c.Request.Context(), oid, templateID, startDate)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, plan)
}

package handler

import (
	"strconv"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) GetCheckinHeatmap(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	if months <= 0 || months > 12 {
		months = 6
	}

	heatmap, err := h.analyticsService.GetCheckinHeatmap(c.Request.Context(), oid, months)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, heatmap)
}

func (h *AnalyticsHandler) GetCheckinTrend(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}

	trend, err := h.analyticsService.GetCheckinTrend(c.Request.Context(), oid, days)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, trend)
}

func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	overview, err := h.analyticsService.GetOverview(c.Request.Context(), oid)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, overview)
}

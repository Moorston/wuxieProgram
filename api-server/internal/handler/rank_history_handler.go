package handler

import (
	"strconv"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RankHistoryHandler struct {
	historyService *service.RankHistoryService
	logger         *zap.Logger
}

func NewRankHistoryHandler(historyService *service.RankHistoryService, logger *zap.Logger) *RankHistoryHandler {
	return &RankHistoryHandler{historyService: historyService, logger: logger}
}

// GetRankTrend 获取排名趋势
func (h *RankHistoryHandler) GetRankTrend(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	period := model.RankPeriod(c.DefaultQuery("period", "all"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	trend, err := h.historyService.GetUserRankTrend(c.Request.Context(), oid, period, days)
	if err != nil {
		h.logger.Error("get rank trend failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"trend": trend})
}

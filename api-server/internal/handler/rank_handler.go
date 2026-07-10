package handler

import (
	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RankHandler struct {
	rankService *service.RankService
	logger      *zap.Logger
}

func NewRankHandler(rankService *service.RankService, logger *zap.Logger) *RankHandler {
	return &RankHandler{rankService: rankService, logger: logger}
}

func (h *RankHandler) GetRankList(c *gin.Context) {
	period := model.RankPeriod(c.DefaultQuery("period", "all"))
	page, pageSize := parsePagination(c, 20)

	entries, err := h.rankService.GetRankList(c.Request.Context(), period, page, pageSize)
	if err != nil {
		h.logger.Error("get rank list failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": entries})
}

func (h *RankHandler) GetMyRank(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	period := model.RankPeriod(c.DefaultQuery("period", "all"))

	entry, err := h.rankService.GetUserRank(c.Request.Context(), oid, period)
	if err != nil {
		response.NotFound(c, "rank not found")
		return
	}

	response.Success(c, entry)
}
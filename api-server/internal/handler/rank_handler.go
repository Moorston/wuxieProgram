package handler

import (
	"log"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RankHandler struct {
	rankService *service.RankService
}

func NewRankHandler(rankService *service.RankService) *RankHandler {
	return &RankHandler{rankService: rankService}
}

func (h *RankHandler) GetRankList(c *gin.Context) {
	period := model.RankPeriod(c.DefaultQuery("period", "all"))
	page, pageSize := parsePagination(c, 20)

	entries, err := h.rankService.GetRankList(c.Request.Context(), period, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, entries)
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
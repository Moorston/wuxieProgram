package handler

import (
	"fmt"
	"strconv"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CompetitionHandler struct {
	compService *service.CompetitionService
	logger      *zap.Logger
}

func NewCompetitionHandler(compService *service.CompetitionService, logger *zap.Logger) *CompetitionHandler {
	return &CompetitionHandler{compService: compService, logger: logger}
}

// --- 管理员 API ---

type CreateCompetitionReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Rules       string `json:"rules"`
	GroupID     string `json:"group_id"`
}

func (h *CompetitionHandler) Create(c *gin.Context) {
	var req CreateCompetitionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	comp := &model.Competition{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.CompetitionStatusActive,
		Rules:       req.Rules,
	}

	if req.GroupID != "" {
		if gid, err := primitive.ObjectIDFromHex(req.GroupID); err == nil {
			comp.GroupID = gid
		}
	}

	if err := h.compService.CreateCompetition(c.Request.Context(), comp); err != nil {
		if err == service.ErrInvalidCompetitionDate {
			response.BadRequest(c, "invalid competition dates")
			return
		}
		h.logger.Error("create competition failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"id": comp.ID.Hex()})
}

func (h *CompetitionHandler) AdminList(c *gin.Context) {
	page, pageSize := parsePagination(c, 20)

	var status *model.CompetitionStatus
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			st := model.CompetitionStatus(v)
			status = &st
		}
	}

	comps, total, err := h.compService.ListCompetitions(c.Request.Context(), page, pageSize, status)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": comps, "total": total})
}

func (h *CompetitionHandler) AdminUpdate(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.compService.UpdateCompetition(c.Request.Context(), id, req); err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// --- 用户 API ---

func (h *CompetitionHandler) List(c *gin.Context) {
	comps, err := h.compService.ListActiveCompetitions(c.Request.Context())
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}
	response.Success(c, gin.H{"list": comps})
}

func (h *CompetitionHandler) Detail(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	comp, err := h.compService.GetCompetition(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "competition not found")
		return
	}

	response.Success(c, comp)
}

type SubmitEntryReq struct {
	CheckinID string `json:"checkin_id" binding:"required"`
}

func (h *CompetitionHandler) Submit(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	compID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req SubmitEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "checkin_id is required")
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(req.CheckinID)
	if err != nil {
		response.BadRequest(c, "invalid checkin_id")
		return
	}

	entry, err := h.compService.SubmitEntry(c.Request.Context(), compID, oid, checkinID)
	if err != nil {
		if err == service.ErrCompetitionNotFound {
			response.NotFound(c, "competition not found")
			return
		}
		if err == service.ErrCompetitionNotActive {
			response.BadRequest(c, "competition is not active")
			return
		}
		if err == service.ErrAlreadySubmitted {
			response.BadRequest(c, "already submitted")
			return
		}
		h.logger.Error("submit entry failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, entry)
}

func (h *CompetitionHandler) Entries(c *gin.Context) {
	compID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 20)
	entries, total, err := h.compService.ListEntries(c.Request.Context(), compID, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": entries, "total": total})
}

func (h *CompetitionHandler) Ranking(c *gin.Context) {
	compID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	ranking, err := h.compService.GetRanking(c.Request.Context(), compID)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"ranking": ranking})
}

type ScoreEntryReq struct {
	Score float64 `json:"score" binding:"required"`
}

func (h *CompetitionHandler) Score(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	entryID, ok := getObjectID(c, "entryId")
	if !ok {
		return
	}

	var req ScoreEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	if err := h.compService.ScoreEntry(c.Request.Context(), entryID, oid, req.Score); err != nil {
		if err == service.ErrInvalidScore {
			response.BadRequest(c, "score must be between 0 and 100")
			return
		}
		if err == service.ErrEntryNotFound {
			response.NotFound(c, "entry not found")
			return
		}
		if err == service.ErrCompetitionNotActive {
			response.BadRequest(c, "competition is not active")
			return
		}
		h.logger.Error("score entry failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// ExportRanking 导出赛事排行榜为 CSV
func (h *CompetitionHandler) ExportRanking(c *gin.Context) {
	compID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	ranking, err := h.compService.GetRanking(c.Request.Context(), compID)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=competition_%s_ranking.csv", compID.Hex()))
	c.Writer.Write([]byte("\xEF\xBB\xBF"))
	c.Writer.Write([]byte("排名,用户ID,昵称,分数\n"))
	for _, r := range ranking {
		nickname := "-"
		if r.User != nil {
			nickname = r.User.Nickname
		}
		c.Writer.Write([]byte(fmt.Sprintf("%d,%s,%s,%.1f\n",
			r.Rank, r.Entry.UserID.Hex(), nickname, r.Score)))
	}
}

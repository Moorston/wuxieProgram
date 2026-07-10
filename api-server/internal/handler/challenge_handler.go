package handler

import (
	"strconv"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChallengeHandler struct {
	challengeService *service.ChallengeService
	logger           *zap.Logger
}

func NewChallengeHandler(challengeService *service.ChallengeService, logger *zap.Logger) *ChallengeHandler {
	return &ChallengeHandler{challengeService: challengeService, logger: logger}
}

type CreateChallengeReq struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	Duration    int    `json:"duration" binding:"required"`
}

// Create 创建挑战
func (h *ChallengeHandler) Create(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	challenge, err := h.challengeService.CreateChallenge(c.Request.Context(), oid, req.Title, req.Description, req.Duration)
	if err != nil {
		if err == service.ErrInvalidDuration {
			response.BadRequest(c, "invalid duration")
			return
		}
		h.logger.Error("create challenge failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, challenge)
}

// List 获取挑战列表
func (h *ChallengeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 50 { pageSize = 20 }

	challenges, total, err := h.challengeService.ListChallenges(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": challenges, "total": total})
}

// Detail 获取挑战详情
func (h *ChallengeHandler) Detail(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	challenge, err := h.challengeService.GetChallenge(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "challenge not found")
		return
	}

	response.Success(c, challenge)
}

// Join 参加挑战
func (h *ChallengeHandler) Join(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	participant, err := h.challengeService.JoinChallenge(c.Request.Context(), id, oid)
	if err != nil {
		if err == service.ErrChallengeNotFound {
			response.NotFound(c, "challenge not found")
			return
		}
		if err == service.ErrChallengeNotActive {
			response.BadRequest(c, "challenge is not active")
			return
		}
		if err == service.ErrAlreadyJoined {
			response.BadRequest(c, "already joined this challenge")
			return
		}
		h.logger.Error("join challenge failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, participant)
}

// Ranking 获取挑战排行榜
func (h *ChallengeHandler) Ranking(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	ranking, err := h.challengeService.GetChallengeRanking(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"ranking": ranking})
}

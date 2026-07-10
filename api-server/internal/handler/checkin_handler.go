package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CheckinHandler struct {
	checkinService *service.CheckinService
	socialService  *service.SocialService
	logger         *zap.Logger
}

func NewCheckinHandler(checkinService *service.CheckinService, socialService *service.SocialService, logger *zap.Logger) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService, socialService: socialService, logger: logger}
}

type PrepareReq struct {
	Description string `json:"description" binding:"max=200"`
}

func (h *CheckinHandler) Prepare(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req PrepareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	checkin, err := h.checkinService.Prepare(c.Request.Context(), oid, req.Description)
	if err != nil {
		h.logger.Error("prepare checkin failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, checkin)
}

func (h *CheckinHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkin, err := h.checkinService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "checkin not found")
		return
	}

	likedMap, err := h.socialService.BatchIsLiked(c.Request.Context(), []primitive.ObjectID{id}, oid)
	if err != nil {
		h.logger.Warn("batch is_liked failed", zap.Error(err))
	}
	checkin.IsLiked = likedMap[id]

	response.Success(c, checkin)
}

func (h *CheckinHandler) GetList(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 10)

	var groupID *primitive.ObjectID
	if gid := c.Query("group_id"); gid != "" {
		if id, err := primitive.ObjectIDFromHex(gid); err == nil {
			groupID = &id
		}
	}

	checkins, total, err := h.checkinService.GetList(c.Request.Context(), oid, groupID, page, pageSize)
	if err != nil {
		h.logger.Error("get checkin list failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	checkinIDs := make([]primitive.ObjectID, len(checkins))
	for i, ci := range checkins {
		checkinIDs[i] = ci.ID
	}
	likedMap, err := h.socialService.BatchIsLiked(c.Request.Context(), checkinIDs, oid)
	if err != nil {
		h.logger.Warn("batch is_liked failed", zap.Error(err))
	}
	for i := range checkins {
		checkins[i].IsLiked = likedMap[checkins[i].ID]
	}

	response.Success(c, gin.H{
		"list":      checkins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CheckinHandler) GetMine(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 10)

	checkins, total, err := h.checkinService.GetMine(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		h.logger.Error("get mine checkins failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{
		"list":  checkins,
		"total": total,
	})
}

func (h *CheckinHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid checkin id")
		return
	}

	if err := h.checkinService.Delete(c.Request.Context(), checkinID, oid); err != nil {
		h.logger.Error("delete checkin failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

func (h *CheckinHandler) Search(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	keyword := c.Query("q")
	if keyword == "" {
		response.BadRequest(c, "keyword is required")
		return
	}

	page, pageSize := parsePagination(c, 10)

	checkins, total, err := h.checkinService.Search(c.Request.Context(), oid, keyword, page, pageSize)
	if err != nil {
		h.logger.Error("search checkin failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	checkinIDs := make([]primitive.ObjectID, len(checkins))
	for i, ci := range checkins {
		checkinIDs[i] = ci.ID
	}
	likedMap, err := h.socialService.BatchIsLiked(c.Request.Context(), checkinIDs, oid)
	if err != nil {
		h.logger.Warn("batch is_liked failed", zap.Error(err))
	}
	for i := range checkins {
		checkins[i].IsLiked = likedMap[checkins[i].ID]
	}

	response.Success(c, gin.H{
		"list":      checkins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 回调接口：media-server 转码完成后调用
type TranscodeCallbackReq struct {
	CheckinID string  `json:"checkin_id" binding:"required"`
	VideoURL  string  `json:"video_url" binding:"required"`
	CoverURL  string  `json:"cover_url" binding:"required"`
	Duration  float64 `json:"duration"`
}

func (h *CheckinHandler) TranscodeCallback(c *gin.Context) {
	var req TranscodeCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	checkinID, err := primitive.ObjectIDFromHex(req.CheckinID)
	if err != nil {
		response.BadRequest(c, "invalid checkin_id")
		return
	}

	if err := h.checkinService.Callback(c.Request.Context(), checkinID, req.VideoURL, req.CoverURL, req.Duration); err != nil {
		h.logger.Error("transcode callback failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}
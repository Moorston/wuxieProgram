package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GroupAnnouncementHandler struct {
	annService *service.GroupAnnouncementService
	logger     *zap.Logger
}

func NewGroupAnnouncementHandler(annService *service.GroupAnnouncementService, logger *zap.Logger) *GroupAnnouncementHandler {
	return &GroupAnnouncementHandler{annService: annService, logger: logger}
}

type CreateAnnouncementReq struct {
	GroupID  string `json:"group_id" binding:"required"`
	Title    string `json:"title" binding:"required,max=100"`
	Content  string `json:"content" binding:"required,max=2000"`
	IsPinned bool   `json:"is_pinned"`
}

// Create 创建公告
func (h *GroupAnnouncementHandler) Create(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid params")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		response.BadRequest(c, "invalid group_id")
		return
	}

	ann, err := h.annService.Create(c.Request.Context(), groupID, oid, req.Title, req.Content, req.IsPinned)
	if err != nil {
		if err == service.ErrGroupNotFound {
			response.NotFound(c, "group not found")
			return
		}
		if err == service.ErrNotGroupLeader {
			response.Forbidden(c, "only group leader can create announcements")
			return
		}
		h.logger.Error("create announcement failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, ann)
}

// List 获取团组公告列表
func (h *GroupAnnouncementHandler) List(c *gin.Context) {
	groupID, ok := getObjectID(c, "group_id")
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 20)

	announcements, total, err := h.annService.List(c.Request.Context(), groupID, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": announcements, "total": total})
}

// Delete 删除公告
func (h *GroupAnnouncementHandler) Delete(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	annID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.annService.Delete(c.Request.Context(), annID, oid); err != nil {
		if err == service.ErrAnnouncementNotFound {
			response.NotFound(c, "announcement not found")
			return
		}
		if err == service.ErrAccessDenied {
			response.Forbidden(c, "access denied")
			return
		}
		h.logger.Error("delete announcement failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

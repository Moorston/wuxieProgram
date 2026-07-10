package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GroupHandler struct {
	groupService *service.GroupService
	logger       *zap.Logger
}

func NewGroupHandler(groupService *service.GroupService, logger *zap.Logger) *GroupHandler {
	return &GroupHandler{groupService: groupService, logger: logger}
}

func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.groupService.List(c.Request.Context())
	if err != nil {
		h.logger.Error("list groups failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, groups)
}

func (h *GroupHandler) Detail(c *gin.Context) {
	id, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	group, err := h.groupService.GetDetail(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "group not found")
		return
	}

	response.Success(c, group)
}

// GenerateInviteCode 生成邀请码（组长操作）
func (h *GroupHandler) GenerateInviteCode(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	groupID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	code, err := h.groupService.GenerateInviteCode(c.Request.Context(), groupID, oid)
	if err != nil {
		if err == service.ErrGroupNotFound {
			response.NotFound(c, "group not found")
			return
		}
		if err == service.ErrNotGroupLeader {
			response.Forbidden(c, "only group leader can generate invite code")
			return
		}
		h.logger.Error("generate invite code failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"invite_code": code})
}

// JoinByInviteCode 通过邀请码加入团组
func (h *GroupHandler) JoinByInviteCode(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invite code is required")
		return
	}

	group, err := h.groupService.JoinByInviteCode(c.Request.Context(), oid, req.Code)
	if err != nil {
		if err == service.ErrInvalidInviteCode {
			response.BadRequest(c, "invalid invite code")
			return
		}
		if err == service.ErrAlreadyMember {
			response.BadRequest(c, "already a member of this group")
			return
		}
		h.logger.Error("join by invite code failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"group": group})
}

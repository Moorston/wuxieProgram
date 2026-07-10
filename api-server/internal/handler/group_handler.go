package handler

import (
	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// RemoveMember 移除成员（组长操作）
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	groupID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "user_id is required")
		return
	}

	targetID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	if err := h.groupService.RemoveMember(c.Request.Context(), groupID, oid, targetID); err != nil {
		if err == service.ErrGroupNotFound {
			response.NotFound(c, "group not found")
			return
		}
		if err == service.ErrNotGroupLeader {
			response.Forbidden(c, "only group leader can remove members")
			return
		}
		if err == service.ErrNotGroupMember {
			response.BadRequest(c, "user is not a member of this group")
			return
		}
		if err == service.ErrCannotRemoveLeader {
			response.BadRequest(c, "cannot remove the group leader")
			return
		}
		h.logger.Error("remove member failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// LeaveGroup 成员主动退出团组
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	groupID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	if err := h.groupService.LeaveGroup(c.Request.Context(), groupID, oid); err != nil {
		if err == service.ErrGroupNotFound {
			response.NotFound(c, "group not found")
			return
		}
		if err == service.ErrNotGroupMember {
			response.BadRequest(c, "not a member of this group")
			return
		}
		if err == service.ErrLeaderCannotLeave {
			response.BadRequest(c, "leader cannot leave; transfer leadership first")
			return
		}
		h.logger.Error("leave group failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// SetLeader 转让组长
func (h *GroupHandler) SetLeader(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	groupID, ok := getObjectID(c, "id")
	if !ok {
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "user_id is required")
		return
	}

	newLeaderID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.BadRequest(c, "invalid user_id")
		return
	}

	if err := h.groupService.SetLeader(c.Request.Context(), groupID, oid, newLeaderID); err != nil {
		if err == service.ErrGroupNotFound {
			response.NotFound(c, "group not found")
			return
		}
		if err == service.ErrNotGroupLeader {
			response.Forbidden(c, "only group leader can transfer leadership")
			return
		}
		if err == service.ErrNotGroupMember {
			response.BadRequest(c, "target user is not a member of this group")
			return
		}
		h.logger.Error("set leader failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

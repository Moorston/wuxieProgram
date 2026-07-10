package handler

import (
	"strconv"

	"wuxie-api/internal/service"
	"wuxie-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type FollowHandler struct {
	followService *service.FollowService
	logger        *zap.Logger
}

func NewFollowHandler(followService *service.FollowService, logger *zap.Logger) *FollowHandler {
	return &FollowHandler{followService: followService, logger: logger}
}

// Follow 关注用户
func (h *FollowHandler) Follow(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.followService.Follow(c.Request.Context(), oid, targetID); err != nil {
		if err == service.ErrCannotFollowSelf {
			response.BadRequest(c, "cannot follow yourself")
			return
		}
		h.logger.Error("follow failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// Unfollow 取消关注
func (h *FollowHandler) Unfollow(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.followService.Unfollow(c.Request.Context(), oid, targetID); err != nil {
		h.logger.Error("unfollow failed", zap.Error(err))
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, nil)
}

// GetFollowing 获取关注列表
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 20)
	users, total, err := h.followService.GetFollowing(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": users, "total": total})
}

// GetFollowers 获取粉丝列表
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, pageSize := parsePagination(c, 20)
	users, total, err := h.followService.GetFollowers(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": users, "total": total})
}

// GetFeed 获取动态流
func (h *FollowHandler) GetFeed(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	feed, err := h.followService.GetFeed(c.Request.Context(), oid, page, pageSize)
	if err != nil {
		response.InternalError(c, "internal server error")
		return
	}

	response.Success(c, gin.H{"list": feed})
}

// GetUserProfile 获取用户主页
func (h *FollowHandler) GetUserProfile(c *gin.Context) {
	oid, ok := getUserID(c)
	if !ok {
		return
	}

	targetID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	profile, err := h.followService.GetUserProfile(c.Request.Context(), targetID, oid)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, profile)
}

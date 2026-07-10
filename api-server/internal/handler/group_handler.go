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
		h.logger.Error("list groups failed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
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

	group, err := h.groupService.Detail(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "group not found")
		return
	}

	response.Success(c, group)
}